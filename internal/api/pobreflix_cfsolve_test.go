package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTurnstileSolver is cfsolve's role: answer the player's widget and hand
// the token back. Recording the request proves the page and sitekey reached it.
type fakeTurnstileSolver struct {
	mu       sync.Mutex
	requests []netx.TurnstileRequest
}

func (s *fakeTurnstileSolver) SolveChallenge(context.Context, string, time.Duration, netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	return &netx.ChallengeSolveResult{}, nil
}

func (s *fakeTurnstileSolver) SolveTurnstile(_ context.Context, r netx.TurnstileRequest, _ time.Duration) (string, error) {
	s.mu.Lock()
	s.requests = append(s.requests, r)
	s.mu.Unlock()
	return "0.the-token", nil
}

func (s *fakeTurnstileSolver) requestsSnapshot() []netx.TurnstileRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]netx.TurnstileRequest(nil), s.requests...)
}

// TestPobreflixGatedDubPlaysThroughCfsolve is the reported title end to end:
// the current page lists /api/player (404), the player host is reconstructed
// from the TMDB id, its Turnstile is answered, and the dubbed proxy plays with
// the proof-bearing page as Referer. The panel and the audio-swap prompt are
// never reached.
func TestPobreflixGatedDubPlaysThroughCfsolve(t *testing.T) {
	const (
		siteKey = "0x4AAAA-pobreflix"
		token   = "0.the-token"
		proof   = "1793605772350.proof"
		proxy   = "/api/proxy?t=cf&ext=mp4"
	)

	solver := &fakeTurnstileSolver{}
	netx.RegisterChallengeSolver(solver)
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })

	var (
		gateHits  atomic.Int32
		panelHits atomic.Int32
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/filme/todo-poderoso":
			_, _ = fmt.Fprint(w, `<script id="initialMovieData" type="application/json">{"type":"movie","item":{"id":310,"tmdb_id":310,"name":"Todo Poderoso","slug":"todo-poderoso"},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"/api/player?id=310"}],"leg":[{"name":"Player Principal Legendado","quality":"HD","url":"/api/player?id=310"}]}}</script>`)
		case "/api/player":
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, "Vídeo indisponível")
		case "/filme/310":
			gateHits.Add(1)
			if r.URL.Query().Get("tsv") == "" {
				fmt.Fprintf(w, `<div class="cf-turnstile" data-sitekey="%s"></div>`, siteKey)
				return
			}
			params := url.Values{}
			params.Set("sources", `[{"kind":"stream","label":"Dublado","url":"`+proxy+`"}]`)
			params.Set("url", proxy)
			_, _ = fmt.Fprint(w, `<script>window.__PLAYER_PARAMS__ = `+strconv.Quote(params.Encode())+`;</script>`)
		case "/api/turnstile-verify":
			var body struct {
				Token string `json:"token"`
			}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, token, body.Token)
			_, _ = fmt.Fprintf(w, `{"success":true,"proof":%q}`, proof)
		default:
			// Any panel hop lands here. It must not.
			panelHits.Add(1)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	t.Setenv("GOANIME_POBREFLIX_PLAYER_URL", srv.URL)
	prev := pbClientFn
	pbClientFn = func() *pobreflix.Client {
		return pobreflix.NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	}
	t.Cleanup(func() { pbClientFn = prev })

	asked := recordSwaps(false, nil)
	prevReferer := util.GetGlobalReferer()
	t.Cleanup(func() {
		if prevReferer == "" {
			util.ClearGlobalReferer()
			return
		}
		util.SetGlobalReferer(prevReferer)
	})

	anime := &models.Anime{
		Name:      "Todo Poderoso",
		URL:       srv.URL + "/filme/todo-poderoso",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}

	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1, "the gated movie is listed directly, not through the panel")
	assert.Equal(t, anime.URL, eps[0].URL)

	got, err := GetPobreflixStreamURL(anime, &eps[0], "best")

	require.NoError(t, err)
	assert.Equal(t, srv.URL+proxy, got)
	assert.Equal(t, srv.URL+"/filme/310?tsv="+proof, util.GetGlobalReferer(),
		"the proof-bearing page is what the media endpoint accepts as Referer")
	assert.Empty(t, *asked, "the dub is the audio the player served; nothing to ask")
	assert.Zero(t, panelHits.Load(), "the panel is not a hop once the gate is cleared")

	reqs := solver.requestsSnapshot()
	require.Len(t, reqs, 1)
	assert.Equal(t, srv.URL+"/filme/310", reqs[0].PageURL)
	assert.Equal(t, siteKey, reqs[0].SiteKey)
	assert.Equal(t, int32(2), gateHits.Load(), "gate page, then the same page with ?tsv=")
}
