package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
