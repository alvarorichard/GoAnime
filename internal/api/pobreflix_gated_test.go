package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The user's report (2026-10-27): "Todo Poderoso" (2003), which played Dublado
// from Pobreflix days earlier, no longer plays. Checked live, hop by hop:
//
//   - every Pobreflix page (movie and episode) now lists one "Player Principal"
//     per audio, a https://streambetter.lol/filme/<tmdb> page — not a file;
//   - that page answers HTML titled "Verificando..." with a Cloudflare
//     Turnstile, and no media until a browser passes it;
//   - GoAnime counted any listed URL as a playable file, so it returned the
//     challenge page as the stream and never reached the panel or the vidsrc
//     fallback it uses when a source has no file.
//
// The dub itself is out of reach without a browser, which GoAnime does not
// ship, so the fix is that the challenge page is never handed to the player,
// the fallbacks run, and the user is asked — with the real reason — before the
// original audio stands in for the dub. This cascade drives every hop above.

// gatedMoviePage is the live movie page's embedded JSON, hosts renamed .test.
const gatedMoviePage = `<!doctype html><html><body>
<script id="initialMovieData" type="application/json">{"type":"movie","item":{"id":310,"tmdb_id":310,"name":"Todo Poderoso","original_name":"Todo Poderoso","slug":"todo-poderoso","year":2003},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"https://streambetter.test/filme/310"}],"leg":[{"name":"Player Principal Legendado","quality":"HD","url":"https://streambetter.test/filme/310"}]}}</script>
</body></html>`

// gatedEpisodePage is the live episode page, whose episode id is a string.
const gatedEpisodePage = `<!doctype html><html><body>
<script id="initialEpisodeData" type="application/json">{"series":{"id":66732,"tmdb_id":66732,"name":"Stranger Things","slug":"stranger-things"},"episode":{"id":"66732-1-1","tmdb_id":66732,"season":1,"episode":1,"title":"Episódio 1"},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"https://streambetter.test/serie/66732/1/1"}],"leg":[]}}</script>
</body></html>`

const turnstilePage = `<!doctype html><html><head><title>Verificando...</title>
<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script></head>
<body><div class="cf-turnstile" data-sitekey="x"></div></body></html>`

// gatedSite serves the hops of the report from one server and counts what the
// app asked of the challenge page.
type gatedSite struct {
	srv        *httptest.Server
	gateHits   atomic.Int32
	masterHits atomic.Int32
}

func useGatedPobreflixSite(t *testing.T) *gatedSite {
	t.Helper()
	site := &gatedSite{}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, p := r.Header.Get("X-Original-Host"), r.URL.Path
		switch {
		case host == "" && p == "/filme/todo-poderoso.html":
			_, _ = fmt.Fprint(w, gatedMoviePage)
		case host == "" && p == "/assistir/stranger-things-t1-ep1.html":
			_, _ = fmt.Fprint(w, gatedEpisodePage)
		case host == "streambetter.test":
			site.gateHits.Add(1)
			_, _ = fmt.Fprint(w, turnstilePage)
		case host == "painel.test" && p == "/filme/"+tpIMDb:
			_, _ = fmt.Fprint(w, todoPoderosoPanel)
		case host != "":
			http.NotFound(w, r) // a dead player host
		// The vidsrc chain, reached directly through GOANIME_VIDSRC_URL.
		case strings.HasPrefix(p, "/vs_src.php"):
			_, _ = fmt.Fprint(w, `{"src":"`+srv.URL+`/embed/movie/`+tpIMDb+`?vs=abc"}`)
		case strings.HasPrefix(p, "/embed/player/"):
			_, _ = fmt.Fprint(w, `<script>window.CONFIG = {"mediaType":"movie","api":"`+srv.URL+
				`/data/api.php?type=movie&imdb=`+tpIMDb+`","apiToken":"tok1","apiTokenRefresh":false};</script>`)
		case strings.HasPrefix(p, "/embed/"):
			_, _ = fmt.Fprint(w, `<script>window.CFG = {"playerUrl":"/embed/player/movie/`+tpIMDb+`?vs=def"}</script>`)
		case strings.HasPrefix(p, "/data/api.php"):
			_, _ = fmt.Fprint(w, `{"status_code":"200","data":{"stream_urls":["`+srv.URL+`/pl/master.m3u8"]}}`)
		case strings.HasSuffix(p, "/generate.php"):
			_, _ = fmt.Fprint(w, "testtoken")
		case strings.HasPrefix(p, "/pl/"):
			site.masterHits.Add(1)
			_, _ = fmt.Fprint(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=5172190,CODECS=\"mp4a.40.2,avc1.640028\",RESOLUTION=1920x1080\n/pl/1080/index.m3u8\n")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	site.srv = srv

	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://painel.test")
	t.Setenv("GOANIME_VIDSRC_URL", srv.URL)
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	sfxEnrichByIDFn = func(m *models.Media, isMovie bool) error {
		if !isMovie || m.TMDBID != tpTMDB {
			return fmt.Errorf("unexpected record lookup: movie=%v tmdb=%d", isMovie, m.TMDBID)
		}
		m.IMDBID = tpIMDb
		m.TMDBDetails = &models.TMDBDetails{Title: "Bruce Almighty"}
		m.Year = "2003"
		return nil
	}

	prev := pbClientFn
	pbClientFn = func() *pobreflix.Client {
		return pobreflix.NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	}
	t.Cleanup(func() { pbClientFn = prev })
	rememberPobreflixGate(srv.URL+"/filme/todo-poderoso.html", nil)
	return site
}

func gatedTodoPoderoso(site *gatedSite) *models.Anime {
	return &models.Anime{
		Name:      "Todo Poderoso",
		URL:       site.srv.URL + "/filme/todo-poderoso.html",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}
}

// TestPobreflixGatedMovieCascade is the report end to end, with the user
// saying "Back" to the original audio: nothing plays, and the message names
// the player that stopped the dub.
func TestPobreflixGatedMovieCascade(t *testing.T) {
	site := useGatedPobreflixSite(t)
	asked := recordSwaps(false, nil)
	anime := gatedTodoPoderoso(site)

	var eps []models.Episode
	stage(t, "1 the page lists a player page, not a file", func(t *testing.T) {
		title, err := pbClientFn().Title(t.Context(), anime.URL)
		require.NoError(t, err)
		assert.Equal(t, 310, title.TMDBID)
		require.NotEmpty(t, title.Servers)
		assert.Empty(t, pobreflix.DirectServers(title.Servers), "nothing the player can open")
		assert.False(t, title.HasDirect(), "the title must not take the direct path")
	})
	stage(t, "2 listing reaches the panel by IMDb id", func(t *testing.T) {
		var err error
		eps, err = GetPobreflixEpisodes(anime)
		require.NoError(t, err)
		require.Len(t, eps, 1)
		assert.Equal(t, "https://painel.test/filme/"+tpIMDb, eps[0].URL)
	})
	stage(t, "3 the dub does not play and the challenge page is never handed out", func(t *testing.T) {
		got, err := GetPobreflixStreamURL(anime, &eps[0], "best")
		require.Error(t, err)
		assert.Empty(t, got, "no stream reaches mpv")
		assert.NotContains(t, got, "streambetter")
		assert.ErrorIs(t, err, ErrAudioUnavailable)
		assert.Zero(t, site.gateHits.Load(), "GoAnime does not fetch the Turnstile page")
		assert.Equal(t, int32(1), site.masterHits.Load(), "the fallback did resolve the original audio")

		require.Len(t, *asked, 1, "the user is asked before the original audio replaces the dub")
		swap := (*asked)[0]
		assert.Equal(t, pobreflix.SourceName, swap.Source)
		assert.Equal(t, startflix.AudioDubbed, swap.Want)
		assert.Equal(t, startflix.AudioSubtitled, swap.Have)
		why := strings.Join(swap.Why, "\n")
		assert.Contains(t, why, "streambetter.test", "the player behind the browser check is named")
		assert.Contains(t, why, "browser check")
		assert.Contains(t, err.Error(), "streambetter.test", "the refusal says why too")
	})
}

// TestPobreflixGatedMoviePlaysOriginalWhenAsked: the user takes the original
// audio. The fallback's master plays, never the challenge page.
func TestPobreflixGatedMoviePlaysOriginalWhenAsked(t *testing.T) {
	site := useGatedPobreflixSite(t)
	asked := recordSwaps(true, nil)
	anime := gatedTodoPoderoso(site)
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)

	got, err := GetPobreflixStreamURL(anime, &eps[0], "best")

	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, "/pl/master.m3u8?token=testtoken"), "the fallback's master: %s", got)
	assert.Len(t, *asked, 1)
	assert.Zero(t, site.gateHits.Load())
}

// TestPobreflixGatedEpisodeIsNotAFile: a series episode whose only server is a
// player page does not resolve to that page. Before the fix the episode page
// did not even decode (its id became a string), so the direct path ended in
// ErrNoEpisode for a reason that had nothing to do with the servers.
func TestPobreflixGatedEpisodeIsNotAFile(t *testing.T) {
	site := useGatedPobreflixSite(t)
	anime := &models.Anime{
		Name:      "Stranger Things",
		URL:       site.srv.URL + "/serie/stranger-things.html",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeTV,
	}
	title := pobreflix.Title{Kind: pobreflix.KindSeries, Slug: "stranger-things", TMDBID: 66732}
	episode := &models.Episode{
		Number:   "1",
		Num:      1,
		SeasonID: "1",
		URL:      site.srv.URL + "/assistir/stranger-things-t1-ep1.html",
	}

	got, err := pobreflixDirectEpisodeURL(t.Context(), anime, title, episode, "best")

	require.ErrorIs(t, err, errPobreflixGated, "the page decodes, and its only server is a player page")
	assert.Empty(t, got)
	assert.Contains(t, pobreflixGateLine(anime.URL), "streambetter.test")
	assert.Zero(t, site.gateHits.Load())
}

// TestPobreflixGateLine: the line is recorded for a gated title, and cleared
// once the title lists a file again.
func TestPobreflixGateLine(t *testing.T) {
	const page = "https://pobreflix.test/filme/x.html"
	t.Cleanup(func() { rememberPobreflixGate(page, nil) })

	_, err := pobreflixFiles(page, []pobreflix.Server{{URL: "https://streambetter.lol/filme/1"}})
	require.ErrorIs(t, err, errPobreflixGated)
	assert.Equal(t, "Pobreflix's own player (streambetter.lol) asks for a browser check GoAnime can't pass.", pobreflixGateLine(page))

	files, err := pobreflixFiles(page, []pobreflix.Server{{URL: "https://streambetter.lol/filme/1"}, {URL: "https://cdn.test/a.mp4"}})
	require.NoError(t, err)
	assert.Len(t, files, 1, "a file next to a player page plays")
	assert.Empty(t, pobreflixGateLine(page), "a title that has a file again is no longer gated")

	files, err = pobreflixFiles(page, nil)
	require.NoError(t, err, "a page that lists nothing is not a gate")
	assert.Empty(t, files)
}
