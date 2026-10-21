package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The id-keyed fallback is the piece that makes a dead panel record watchable
// again. These tests drive the real api flow: a Pobreflix-style movie whose
// panel lists only long-dead player hosts now plays through
// sfxVIDSRCFallbackFn, and the title stays listed because its own page is
// blameless.

const fakeVIDSRCMaster = `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=3905000,RESOLUTION=1920x1080
/master/index.m3u8
`

// TestPobreflixDeadPanelPlaysThroughVIDSRC is the "Todo Poderoso" scenario:
// every panel server is dead, the fallback resolves the movie by IMDb id on
// its own chain, and playback succeeds.
func TestPobreflixDeadPanelPlaysThroughVIDSRC(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case p == "/filmes/online/todo-poderoso-dublado-450/":
			_, _ = w.Write([]byte(fakeTodoPoderosoPage))
		case p == "/filme/tt1900": // the movie panel by IMDb id: dead players
			_, _ = w.Write([]byte(deadPlayersPanel))
		case strings.HasPrefix(p, "/vs_src.php"):
			_, _ = w.Write([]byte(`{"src":"` + srv.URL + `/embed/movie/tt1900?vs=abc"}`))
		case strings.HasPrefix(p, "/embed/player/"): // the real player page
			_, _ = w.Write([]byte(`<script>window.CONFIG = {"mediaType":"movie","api":"` + srv.URL +
				`/data/api.php?type=movie&imdb=tt1900","apiToken":"tok1","apiTokenRefresh":false};</script>`))
		case strings.HasPrefix(p, "/embed/"): // landing → playerUrl
			_, _ = w.Write([]byte(`<script>window.CFG = {"playerUrl":"/embed/player/movie/tt1900?vs=def"}</script>`))
		case strings.HasPrefix(p, "/data/api.php"):
			assert.Equal(t, "tt1900", r.URL.Query().Get("imdb"), "the fallback keys the movie by its IMDb id")
			assert.Equal(t, "tok1", r.URL.Query().Get("api_token"), "the single-use token rides along")
			_, _ = w.Write([]byte(`{"status_code":"200","data":{"stream_urls":["` + srv.URL + `/pl/master.m3u8"]}}`))
		case strings.HasSuffix(p, "/generate.php"):
			_, _ = w.Write([]byte("testtoken"))
		case strings.HasPrefix(p, "/pl/"):
			assert.Equal(t, "testtoken", r.URL.Query().Get("token"), "the master request carries the host token")
			_, _ = w.Write([]byte(fakeVIDSRCMaster))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("GOANIME_VIDSRC_URL", srv.URL)
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
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
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", srv.URL)

	anime := &models.Anime{
		Name:      "Todo Poderoso",
		URL:       "https://www.pobreflixtv.sale/filmes/online/todo-poderoso-dublado-450/",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	url, err := GetPobreflixStreamURL(anime, &eps[0], "best")
	require.NoError(t, err, "the fallback must save the title its panel could not")
	assert.Equal(t, srv.URL+"/pl/master.m3u8?token=testtoken", url, "the resolved master is token-stamped")
	assert.False(t, isPanelDead(anime.URL), "a title the fallback saved stays listed")
}

// TestStartFlixDeadPanelAndFallbackMissKeepsMessage covers when BOTH the panel
// players and the fallback fail: the user still sees the honest dead-servers
// message, and a StartFlix listing still hides (its page is the stale part).
func TestStartFlixDeadPanelAndFallbackMissKeepsMessage(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch p {
		case "/filmes/deadfilm/":
			_, _ = w.Write([]byte(`<html><body><iframe src="` + srv.URL + `/filme/tt1900"></iframe></body></html>`))
		case "/filme/tt1900":
			_, _ = w.Write([]byte(deadPlayersPanel))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	// The fallback misses, as it would when the embed source is off air:
	// the original panel error must survive for the user.
	prev := sfxVIDSRCFallbackFn
	sfxVIDSRCFallbackFn = func(funcCtx context.Context, media *models.Anime, episode *models.Episode) (*startflix.Stream, error) {
		return nil, fmt.Errorf("least fortress: the fallback is off air")
	}
	t.Cleanup(func() { sfxVIDSRCFallbackFn = prev })

	anime := &models.Anime{
		Name:      "Dead Film",
		URL:       srv.URL + "/filmes/deadfilm/",
		Source:    "StartFlix",
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetStartFlixEpisodes(anime)
	require.NoError(t, err)

	_, err = GetStartFlixStreamURL(anime, &eps[0], "best")
	require.Error(t, err)
	// The user-facing message still names the listed dead hosts (StartFlix's
	// own wording, unlike Pobreflix's movie-aware "no longer answer").
	msg := err.Error()
	assert.Contains(t, msg, "only offers this")
	assert.Contains(t, msg, "actelecup.test")
	assert.True(t, isPanelDead(anime.URL), "a StartFlix listing still hides when the fallback also misses")
}
