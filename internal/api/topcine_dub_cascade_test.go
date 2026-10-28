package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The user's report (2026-10-21): "Todo Poderoso" (2003) from TopCine plays
// in English, though TopCine offers it Dublado. Every hop was checked live:
//
//   - topcine3.site/filme/todo-poderoso iframes azullog.top/filme/310, whose
//     language check answers Dublado and Legendado — both on 1take.top, which
//     serves a Cloudflare Turnstile ("Verificando acesso") before any video.
//   - GoAnime plays TopCine titles through the StartFlix panel by id. The
//     panel's /filme/tt0315327 lists five "HD #N Dublado" players on hosts
//     whose DNS is gone (actelecup, letsupload, fembed, streamango, 4toshare).
//   - The vidsrc fallback then resolves the movie: one master, one muxed
//     audio track, English.
//   - GoAnime told mpv to prefer Portuguese and said nothing; mpv played the
//     only track there was.
//
// The dub is out of reach without a browser, so the fix is that GoAnime never
// passes the original audio off as the dub: it asks, says why, and tells mpv
// the truth. This cascade drives every hop above against one stub server —
// the real TopCine client, the real panel client and the real vidsrc chain —
// stage by stage, so a break names the hop that broke.

const (
	tpTMDB = 310
	tpIMDb = "tt0315327"
)

// todoPoderosoPanel is the live panel record, hosts renamed to .test.
const todoPoderosoPanel = `<main><div class="choose-player" id="player-chooser"><h1>SELECIONE UM PLAYER</h1><div id="players">
	<button class="btn" data-show-player="true" data-source="https://actelecup.test/video/2c383b03860c7cc5/iframe" data-subtitles="" data-type="iframe" data-id="1">&nbsp; HD #1 Dublado                        </button>
	<button class="btn" data-show-player="true" data-source="https://letsupload.test/plugins/mediaplayer/site/_embed.php?u=eglW&w" data-subtitles="" data-type="iframe" data-id="2">&nbsp; HD #2 Dublado</button>
	<button class="btn" data-show-player="true" data-source="https://www.fembed.test/v/7qv7nqlr0wo" data-subtitles="" data-type="iframe" data-id="3">&nbsp; HD #3 Dublado</button>
	<button class="btn" data-show-player="true" data-source="https://streamango.test/embed/famdqlffonckmrkm" data-subtitles="" data-type="iframe" data-id="4">&nbsp; HD #4 Dublado</button>
	<button class="btn" data-show-player="true" data-source="https://4toshare.test/embed-zvrbhnrc8etu.html" data-subtitles="" data-type="iframe" data-id="5">&nbsp; HD #5 Dublado</button>
</div></div></main>`

// todoPoderosoSite counts what the app asked of each hop.
type todoPoderosoSite struct {
	srv        *httptest.Server
	langChecks atomic.Int32
	masterHits atomic.Int32
}

// useTodoPoderosoSite serves every hop of the report from one server and
// points the app at it: TopCine and its player by the host the client asked
// for (routeAll), the panel by GOANIME_STARTFLIX_PANEL_URL, and the vidsrc
// chain by GOANIME_VIDSRC_URL. Only the official-record lookup is a stub.
func useTodoPoderosoSite(t *testing.T) *todoPoderosoSite {
	t.Helper()
	site := &todoPoderosoSite{}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, p := r.Header.Get("X-Original-Host"), r.URL.Path
		switch {
		case host == "topcine.test" && p == "/buscar":
			_, _ = fmt.Fprint(w, `<html><body>
				<article class="pagina-busca-card"><a href="/filme/a-volta-do-todo-poderoso"><img src="https://image.tmdb.org/t/p/w342/v.jpg"><h2>A Volta do Todo Poderoso</h2><small>2007 • Filme</small></a></article>
				<article class="pagina-busca-card"><a href="/filme/todo-poderoso"><img src="https://image.tmdb.org/t/p/w342/b.jpg"><h2>Todo Poderoso</h2><small>2003 • Filme</small></a></article>
			</body></html>`)
		case host == "topcine.test" && p == "/filme/todo-poderoso":
			_, _ = fmt.Fprintf(w, `<html><body><h2>Assistir Todo Poderoso Online</h2>
				<iframe src="https://azullog.test/filme/%d" title="Player de Todo Poderoso"></iframe></body></html>`, tpTMDB)
		case host == "azullog.test" && p == fmt.Sprintf("/filme/%d", tpTMDB) && r.URL.Query().Get("acao") == "carregar_idiomas":
			site.langChecks.Add(1)
			_, _ = fmt.Fprintf(w, `{"sucesso":true,"tem_dublado":true,"tem_legendado":true,"url_dublado":"https://1take.test/e/tmdb%ddub","url_legendado":"https://1take.test/e/tmdb%dleg"}`, tpTMDB, tpTMDB)
		case host == "painel.test" && p == "/filme/"+tpIMDb:
			_, _ = fmt.Fprint(w, todoPoderosoPanel)
		case host != "":
			http.NotFound(w, r) // a dead player host, or a hop nobody expects
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
			// The live master: three video variants, one muxed English track.
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

	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := topcine.NewClientForTest(&http.Client{Transport: routeAll{target}}, "https://topcine.test")
	prevClient := tcClientFn
	tcClientFn = func() *topcine.Client { return client }
	t.Cleanup(func() { tcClientFn = prevClient })
	return site
}

// todoPoderosoFromSearch finds the 2003 movie the way the user did.
func todoPoderosoFromSearch(t *testing.T) *models.Anime {
	t.Helper()
	results, err := tcClientFn().Search(t.Context(), "todo poderoso")
	require.NoError(t, err)
	for _, r := range results {
		if r.Title == "Todo Poderoso" && r.Year == "2003" {
			return r.ToAnimeModel()
		}
	}
	t.Fatalf("the 2003 movie is not among %d results", len(results))
	return nil
}

// recordSwaps answers every swap question with answer, recording them.
func recordSwaps(answer bool, err error) *[]audioSwap {
	var asked []audioSwap
	sfxAudioSwapFn = func(swap audioSwap) (bool, error) {
		asked = append(asked, swap)
		return answer, err
	}
	return &asked
}

// stage runs one hop of a cascade; the cascade stops at the first broken hop,
// so the failure names the hop rather than everything downstream of it.
func stage(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	if !t.Run(name, fn) {
		t.Fatalf("cascade stopped: stage %q failed", name)
	}
}

// TestTopCineTodoPoderosoCascade is the report end to end, with the user
// saying "Back" to the original audio: nothing plays, and the message says
// why in plain words.
func TestTopCineTodoPoderosoCascade(t *testing.T) {
	site := useTodoPoderosoSite(t)
	asked := recordSwaps(false, nil)
	util.SetGlobalAudioLanguage("left-behind")

	var anime *models.Anime
	var eps []models.Episode
	stage(t, "1 search lists the 2003 movie", func(t *testing.T) {
		anime = todoPoderosoFromSearch(t)
		assert.Equal(t, topcine.SourceName, anime.Source)
		assert.Equal(t, models.MediaTypeMovie, anime.MediaType)
	})
	stage(t, "2 title page names TMDB 310", func(t *testing.T) {
		title, err := tcClientFn().Title(t.Context(), anime.URL)
		require.NoError(t, err)
		assert.Equal(t, topcine.KindMovie, title.Kind)
		assert.Equal(t, tpTMDB, title.TMDBID)
	})
	stage(t, "3 TopCine's own player has the dub, behind 1take", func(t *testing.T) {
		langs, err := topCineLanguages(anime)
		require.NoError(t, err)
		assert.True(t, langs.Dubbed, "TopCine offers it Dublado: that is the whole report")
		assert.Equal(t, []string{"1take.test"}, langs.Hosts)
	})
	stage(t, "4 listing reaches the panel by IMDb id", func(t *testing.T) {
		var err error
		eps, err = GetTopCineEpisodes(anime)
		require.NoError(t, err)
		require.Len(t, eps, 1)
		assert.Equal(t, "https://painel.test/filme/"+tpIMDb, eps[0].URL)
		assert.Equal(t, tpIMDb, anime.IMDBID)
	})
	stage(t, "5 the panel lists the movie Dublado only", func(t *testing.T) {
		players, err := sfxClientFn().Players(t.Context(), eps[0].URL)
		require.NoError(t, err)
		require.Len(t, players, 5)
		assert.Equal(t, []startflix.Audio{startflix.AudioDubbed}, startflix.ListedAudios(players, ""))
	})
	stage(t, "6 the dub does not play and the original is not passed off as it", func(t *testing.T) {
		got, err := GetTopCineStreamURL(anime, &eps[0], "best")
		require.Error(t, err)
		assert.Empty(t, got, "no stream reaches mpv")
		assert.ErrorIs(t, err, ErrAudioUnavailable)
		assert.Equal(t, int32(1), site.masterHits.Load(), "the fallback did resolve the original audio")

		require.Len(t, *asked, 1, "the user is asked before the original audio replaces the dub")
		swap := (*asked)[0]
		assert.Equal(t, topcine.SourceName, swap.Source)
		assert.Equal(t, "Todo Poderoso", swap.Title)
		assert.Equal(t, startflix.AudioDubbed, swap.Want)
		assert.Equal(t, startflix.AudioSubtitled, swap.Have)
		why := strings.Join(swap.Why, "\n")
		assert.Contains(t, why, "actelecup.test", "the dead Dublado servers are named")
		assert.Contains(t, why, "1take.test", "where TopCine keeps the dub is named")
		assert.Contains(t, why, "without Portuguese subtitles", "the fallback carries none")

		msg := err.Error()
		assert.Contains(t, msg, "Dublado")
		assert.Contains(t, msg, "1take.test")
		assert.NotContains(t, msg, "--audio", "the user answered; no flag advice")
	})
	stage(t, "7 nothing was told to mpv, and the title stays listed", func(t *testing.T) {
		assert.Equal(t, "left-behind", util.GetGlobalAudioLanguage(), "a refused stream sets no audio preference")
		assert.False(t, isPanelDead(anime.URL), "TopCine's page is not the stale part")
	})
}

// TestTopCineTodoPoderosoOriginalOnlyWhenAsked: the user takes the original
// audio. It plays, mpv is told it is the original, and the rest of the
// session does not ask again for the same title.
func TestTopCineTodoPoderosoOriginalOnlyWhenAsked(t *testing.T) {
	useTodoPoderosoSite(t)
	asked := recordSwaps(true, nil)
	anime := todoPoderosoFromSearch(t)
	eps, err := GetTopCineEpisodes(anime)
	require.NoError(t, err)

	got, err := GetTopCineStreamURL(anime, &eps[0], "best")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, "/pl/master.m3u8?token=testtoken"), "the fallback's master: %s", got)
	assert.Len(t, *asked, 1)
	assert.Equal(t, originalALang, util.GetGlobalAudioLanguage(), "mpv is told the stream's real audio, not Portuguese")

	_, err = GetTopCineStreamURL(anime, &eps[0], "best")
	require.NoError(t, err)
	assert.Len(t, *asked, 1, "one answer per title")
}

// TestTopCineTodoPoderosoWithoutTerminalRefuses: a download with no terminal
// to ask on gets no English file named after a Dublado title, and learns the
// flag that would take it.
func TestTopCineTodoPoderosoWithoutTerminalRefuses(t *testing.T) {
	useTodoPoderosoSite(t)
	asked := recordSwaps(false, errors.New("no terminal: stdin is not a TTY"))
	anime := todoPoderosoFromSearch(t)
	eps, err := GetTopCineEpisodes(anime)
	require.NoError(t, err)

	_, err = GetTopCineStreamURL(anime, &eps[0], "best")
	require.ErrorIs(t, err, ErrAudioUnavailable)
	assert.Len(t, *asked, 1)
	assert.Contains(t, err.Error(), "--audio eng")
}

// TestTopCineTodoPoderosoPinnedAudio: an explicit --audio is the answer given
// in advance. The original asked for plays unasked; a pinned Portuguese is
// never silently swapped.
func TestTopCineTodoPoderosoPinnedAudio(t *testing.T) {
	for _, tt := range []struct {
		pin     string
		wantErr bool
	}{
		{pin: "eng", wantErr: false},
		{pin: "pt-BR", wantErr: true},
	} {
		t.Run(tt.pin, func(t *testing.T) {
			useTodoPoderosoSite(t)
			asked := recordSwaps(false, nil)
			util.SetGlobalAudioLanguage(tt.pin)
			util.GlobalAudioLanguageExplicit = true
			anime := todoPoderosoFromSearch(t)
			eps, err := GetTopCineEpisodes(anime)
			require.NoError(t, err)

			_, err = GetTopCineStreamURL(anime, &eps[0], "best")
			if tt.wantErr {
				require.ErrorIs(t, err, ErrAudioUnavailable)
				assert.Len(t, *asked, 1, "a pinned dub is asked about, not swapped")
			} else {
				require.NoError(t, err)
				assert.Empty(t, *asked, "the original was asked for in advance")
			}
			assert.Equal(t, tt.pin, util.GetGlobalAudioLanguage(), "a pinned --audio is the user's, left alone")
		})
	}
}
