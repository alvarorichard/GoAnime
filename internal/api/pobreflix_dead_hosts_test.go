package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadPlayersPanel is a movie panel whose every player sits on a host no
// resolver covers and whose DNS is dead — the shape the live "Todo Poderoso
// (2003)" case hit (checked 2026-10-17: actelecup.com, letsupload.co,
// fembed.com, streamango.com, 4toshare.com — two NXDOMAIN, three gone TLS).
// Under tests the yt-dlp fallback stays off, so ResolveStream returns a
// NoStreamError naming the hosts and nothing else: the Unsupported-only
// shape isPanelDeadErr reads as "will not heal this session".
const deadPlayersPanel = `<main><div class="choose-player" id="player-chooser"><div id="players">
	<button class="btn" data-show-player="true" data-source="https://actelecup.test/video/x/iframe" data-subtitles="" data-type="iframe" data-id="4891">HD #1 Dublado</button>
	<button class="btn" data-show-player="true" data-source="https://letsupload.test/_embed.php?u=eglW" data-subtitles="" data-type="iframe" data-id="4892">HD #2 Dublado</button>
</div></div></main>`

const fakeTodoPoderosoPage = `<!DOCTYPE html><html><head><title>Assistir Todo Poderoso Online</title>
<link rel="canonical" href="https://www.pobreflixtv.sale/filmes/online/todo-poderoso-dublado-450/">
</head><body><h2 class="one">Todo Poderoso</h2>
<div id="view" data-video-id="450" data-tmdb-id="310" data-temp="" data-ep=""></div></body></html>`

// TestPobreflixDeadHostsKeepsTitleAndExplains drives the user's report:
// "Todo Poderoso (2003)" (https://www.pobreflixtv.sale/filmes/online/todo-poderoso-dublado-450/)
// is a Pobreflix-only movie whose panel lists five long-dead player hosts.
// Two things had to hold and did not:
//
//  1. the failure hid the title from every later search in the session
//     (markPanelDead keyed the Pobreflix page URL) — after one failed
//     attempt the app could not find the movie again, though its page
//     stays as real as ever;
//  2. the message told the user to "try another episode" of a movie,
//     which has none.
//
// pobreflixDeadServersMessage has its own unit test below; this one runs the
// whole api flow.
func TestPobreflixDeadHostsKeepsTitleAndExplains(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/filmes/online/todo-poderoso-dublado-450/":
			_, _ = w.Write([]byte(fakeTodoPoderosoPage))
		case r.URL.Path == "/filme/tt1900": // the movie panel by IMDb id
			_, _ = w.Write([]byte(deadPlayersPanel))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	sfxEnrichByIDFn = func(m *models.Media, _ bool) error {
		m.IMDBID = "tt1900"
		m.TMDBDetails = &models.TMDBDetails{Title: "Bruce Almighty"}
		m.Year = "2003"
		return nil
	}
	prevPB := pbClientFn
	pbClientFn = func() *pobreflix.Client { return pbTitleClient(t, srv) }
	t.Cleanup(func() { pbClientFn = prevPB })
	// The panel base is where MoviePanel points Pobreflix movies by IMDb id.
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", srv.URL)

	anime := &models.Anime{
		Name:      "Todo Poderoso",
		URL:       "https://www.pobreflixtv.sale/filmes/online/todo-poderoso-dublado-450/",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err, "listing a movie needs no player hosts")
	require.Len(t, eps, 1)
	assert.Equal(t, models.MediaTypeMovie, anime.MediaType, "the panel names the kind")

	_, err = GetPobreflixStreamURL(anime, &eps[0], "best")
	require.Error(t, err)
	require.ErrorIs(t, err, startflix.ErrNoSupportedServer, "the dead servers surface as an unsupported-hosts error")

	// The message names the dead servers and gives up for today without
	// sending the user after a movie's "other episode".
	msg := err.Error()
	assert.Contains(t, msg, "no longer answer")
	assert.Contains(t, msg, "actelecup.test", "the message names the listed hosts: %s", msg)
	assert.NotContains(t, msg, "another episode", "a movie has no other episode")

	// The title must stay in the search results: the panel's stale record is
	// not the Pobreflix page's fault.
	assert.False(t, isPanelDead(anime.URL), "a Pobreflix page is never the failing listing")
	kept := filterPanelDead([]*models.Anime{anime})
	require.Len(t, kept, 1)
	assert.Equal(t, anime.URL, kept[0].URL)
}

// pbTitleClient builds the Pobreflix client test seam routes through the
// fake site: the title page keeps its real-world pobreflixtv.sale URL (that
// is what a search result carries) and the transport redirects it there.
func pbTitleClient(t *testing.T, srv *httptest.Server) *pobreflix.Client {
	t.Helper()
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return pobreflix.NewClientForTest(&http.Client{Transport: routeAll{target}}, srv.URL,
		filepath.Join(t.TempDir(), "pobreflix-search.cache"))
}

// usePanelTransport repoints the StartFlix client seam at srv, so the panel
// fetch (movie: /filme/<imdb>) lands on this test's server instead of the
// useFakeStartFlix one, which returns ErrNotOnPanel answers.
func usePanelTransport(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := startflix.NewClientForTest(&http.Client{Transport: routeAll{target}}, "https://www.startflix.test")
	prev := sfxClientFn
	sfxClientFn = func() *startflix.Client { return client }
	t.Cleanup(func() { sfxClientFn = prev })
}

// TestStartFlixDeadMovieStaysHidden guards the other half of the split: for a
// StartFlix listing the dead-panel title hides as before, because there the
// listing page itself is the stale artifact (see markPanelDeadErr).
func TestStartFlixDeadMovieStaysHidden(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/filmes/deadfilm/":
			_, _ = w.Write([]byte(`<html><body><iframe src="` + srv.URL + `/filme/tt1900"></iframe></body></html>`))
		case r.URL.Path == "/filme/tt1900":
			_, _ = w.Write([]byte(deadPlayersPanel))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)

	anime := &models.Anime{
		Name:      "Dead Film",
		URL:       srv.URL + "/filmes/deadfilm/",
		Source:    "StartFlix",
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetStartFlixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.False(t, isPanelDead(anime.URL), "listing does not reach the panel")

	_, err = GetStartFlixStreamURL(anime, &eps[0], "best")
	require.Error(t, err)
	var noStream *startflix.NoStreamError
	require.ErrorAs(t, err, &noStream)
	require.Empty(t, noStream.Failures, "the players are unsupported, not failed")
	require.NotEmpty(t, noStream.Unsupported)

	assert.True(t, isPanelDead(anime.URL), "a StartFlix listing still hides on the dead-player failure")
}

// TestPobreflixDeadHostsMessageSaysTitleOrMovie pins the noun: a movie says
// "movie" (the reported case), anything else stays "title".
func TestPobreflixDeadHostsMessageSaysTitleOrMovie(t *testing.T) {
	t.Parallel()
	movieMedia := &models.Anime{Name: "Todo Poderoso", MediaType: models.MediaTypeMovie}
	msg := pobreflixDeadServersMessage(movieMedia, []string{"actelecup.test", "letsupload.test"})
	assert.Contains(t, msg, "this movie only on servers that no longer answer")
	assert.Contains(t, msg, "actelecup.test, letsupload.test")
	assert.Contains(t, msg, "Try again later")

	seriesMedia := &models.Anime{Name: "Show", MediaType: models.MediaTypeTV}
	series := pobreflixDeadServersMessage(seriesMedia, []string{"x.test"})
	assert.True(t, strings.Contains(series, "this title only on servers"), "a series keeps the generic noun: %s", series)
}
