package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPobreflixGatedDubPlaysWithoutAnySolver is the title that stopped playing
// when the site moved to the RedeFlixApi player: the page lists a player with
// no Turnstile. No solver is registered, yet the dub plays, the panel is never
// touched, and the ticket is traded the way the player's own script trades it.
func TestPobreflixGatedDubPlaysWithoutAnySolver(t *testing.T) {
	const media = "https://cdn.example/videos/todo-poderoso.mp4?X-Amz-Signature=sig"
	netx.RegisterChallengeSolver(nil)

	var panelHits, posts atomic.Int32
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/filme/todo-poderoso":
			_, _ = fmt.Fprintf(w, `<script id="initialMovieData" type="application/json">{"type":"movie","item":{"id":310,"tmdb_id":310,"name":"Todo Poderoso","slug":"todo-poderoso"},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"%[1]s/filme/310"}],"leg":[{"name":"Player Principal Legendado","quality":"HD","url":"%[1]s/filme/310"}]}}</script>`, srv.URL)
		case r.URL.Path == "/filme/310" && r.Method == http.MethodGet:
			_, _ = fmt.Fprintf(w, `<script>var playbackResolveTicket = "v1.TICKET";
var playbackResolveUrl = "%s/playback-resolve.php";</script>`, srv.URL)
		case r.URL.Path == "/playback-resolve.php":
			posts.Add(1)
			assert.Equal(t, srv.URL, r.Header.Get("Origin"))
			var body map[string]string
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "default", body["server"], "the dub is the track the page lists as default")
			_ = json.NewEncoder(w).Encode(map[string]string{"launchTicket": "v1.LAUNCH"})
		case r.URL.Path == "/playerkys/source.php":
			posts.Add(1)
			assert.Equal(t, srv.URL, r.Header.Get("Origin"))
			_ = json.NewEncoder(w).Encode(map[string]string{"url": media})
		default:
			panelHits.Add(1) // any panel hop lands here
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	t.Setenv("GOANIME_POBREFLIX_OPEN_PLAYER_HOSTS", "127.0.0.1")
	prev := pbClientFn
	pbClientFn = func() *pobreflix.Client {
		return pobreflix.NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	}
	t.Cleanup(func() { pbClientFn = prev })
	asked := recordSwaps(false, nil)

	anime := &models.Anime{
		Name:      "Todo Poderoso",
		URL:       srv.URL + "/filme/todo-poderoso",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	got, err := GetPobreflixStreamURL(anime, &eps[0], "best")

	require.NoError(t, err)
	assert.Equal(t, media, got)
	assert.Equal(t, int32(2), posts.Load(), "resolve, then source")
	assert.Zero(t, panelHits.Load(), "the panel is not a hop when the player answers")
	assert.Empty(t, *asked, "the dub was asked for and served; nothing to confirm")
}

// redeflixFake is a RedeFlix player on its own host: the page, the resolve
// endpoint (which can be made to fail) and the source endpoint.
type redeflixFake struct {
	srv          *httptest.Server
	media        string
	failResolve  bool
	resolveCalls atomic.Int32
}

func newRedeflixFake(t *testing.T, media string, failResolve bool) *redeflixFake {
	t.Helper()
	f := &redeflixFake{media: media, failResolve: failResolve}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			_, _ = fmt.Fprintf(w, `<script>var playbackResolveTicket = "v1.T";
var playbackResolveUrl = "%s/playback-resolve.php";</script>`, f.srv.URL)
		case r.URL.Path == "/playback-resolve.php":
			f.resolveCalls.Add(1)
			if f.failResolve {
				w.WriteHeader(http.StatusForbidden)
				_, _ = fmt.Fprint(w, `{"error":"Origem de reprodução não autorizada."}`)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"launchTicket": "v1.L"})
		case r.URL.Path == "/playerkys/source.php":
			_ = json.NewEncoder(w).Encode(map[string]string{"url": f.media})
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// pobreflixCatalog serves a title page listing the given player URLs for the
// dub, best first as the site lists them.
func pobreflixCatalog(t *testing.T, players ...string) *httptest.Server {
	t.Helper()
	var list []string
	for i, u := range players {
		quality := "HD"
		if i == 0 {
			quality = "FHD" // the first scores higher, as a Turnstile player usually does
		}
		list = append(list, fmt.Sprintf(`{"name":"Player %d","quality":%q,"url":%q}`, i, quality, u))
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/filme/todo-poderoso" {
			_, _ = fmt.Fprintf(w, `<script id="initialMovieData" type="application/json">{"type":"movie","item":{"id":310,"tmdb_id":310,"name":"Todo Poderoso","slug":"todo-poderoso"},"servers":{"dub":[%s],"leg":[]}}</script>`, strings.Join(list, ","))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func playPobreflixMovie(t *testing.T, catalog *httptest.Server) (string, error) {
	t.Helper()
	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, catalog)
	prev := pbClientFn
	pbClientFn = func() *pobreflix.Client {
		return pobreflix.NewClientForTest(catalog.Client(), catalog.URL, filepath.Join(t.TempDir(), "cache.json"))
	}
	t.Cleanup(func() { pbClientFn = prev })
	recordSwaps(false, nil)
	anime := &models.Anime{Name: "Todo Poderoso", URL: catalog.URL + "/filme/todo-poderoso", Source: pobreflix.SourceName, MediaType: models.MediaTypeMovie}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	return GetPobreflixStreamURL(anime, &eps[0], "best")
}

// With no solver installed, a Turnstile player listed first (and scoring
// higher) must not hide the RedeFlix player listed after it: the feature is
// for exactly this page.
func TestPobreflixNoSolverSkipsTheTurnstilePlayerForTheOpenOne(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	open := newRedeflixFake(t, "https://cdn.example/open.mp4", false)
	// "localhost" and 127.0.0.1 are different hosts: only the latter is open.
	turnstile := "http://localhost:1/filme/310"
	t.Setenv("GOANIME_POBREFLIX_OPEN_PLAYER_HOSTS", "127.0.0.1")

	got, err := playPobreflixMovie(t, pobreflixCatalog(t, turnstile, open.srv.URL+"/filme/310"))

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/open.mp4", got)
}

// A player that fails (403, an expired ticket, a timeout) is not the end of the
// title when another one is listed.
func TestPobreflixTriesTheNextPlayerWhenOneFails(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	broken := newRedeflixFake(t, "https://cdn.example/never.mp4", true)
	working := newRedeflixFake(t, "https://cdn.example/second.mp4", false)
	t.Setenv("GOANIME_POBREFLIX_OPEN_PLAYER_HOSTS", "127.0.0.1")

	got, err := playPobreflixMovie(t, pobreflixCatalog(t, broken.srv.URL+"/filme/310", working.srv.URL+"/filme/310"))

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/second.mp4", got)
	assert.Equal(t, int32(1), broken.resolveCalls.Load(), "the broken one was tried once, not retried")
}

// The site lists the same player URL for both audios; trying it twice would
// double the requests for nothing.
func TestPobreflixDoesNotTryTheSamePlayerTwice(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	broken := newRedeflixFake(t, "https://cdn.example/never.mp4", true)
	t.Setenv("GOANIME_POBREFLIX_OPEN_PLAYER_HOSTS", "127.0.0.1")
	url := broken.srv.URL + "/filme/310"

	_, err := playPobreflixMovie(t, pobreflixCatalog(t, url, url))

	require.Error(t, err)
	assert.Equal(t, int32(1), broken.resolveCalls.Load())
}

func TestRankPobreflixGatedServersOrdersByAudioThenQualityOncePerURL(t *testing.T) {
	servers := []pobreflix.Server{
		{Name: "files", URL: "https://cdn.example/f.mp4", Quality: "FHD"}, // a file, not a player page
		{Name: "sub-hd", URL: "https://p.example/sub", Quality: "HD", Audio: pobreflix.AudioSubtitled},
		{Name: "dub-hd", URL: "https://p.example/dub", Quality: "HD", Audio: pobreflix.AudioDubbed},
		{Name: "dub-hd-again", URL: "https://p.example/dub", Quality: "HD", Audio: pobreflix.AudioDubbed}, // same URL
		{Name: "dub-fhd", URL: "https://p.example/dub-fhd", Quality: "FHD", Audio: pobreflix.AudioDubbed},
	}

	got := rankPobreflixGatedServers(servers, pobreflix.AudioDubbed, "best")

	var names []string
	for _, s := range got {
		names = append(names, s.Name)
	}
	assert.Equal(t, []string{"dub-fhd", "dub-hd", "sub-hd"}, names,
		"wanted audio first, then quality; a file is not a player and one URL is listed once")
	assert.Equal(t, "dub-hd", got[1].Name, "the first listing of a repeated URL is the one kept")
}
