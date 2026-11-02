package pobreflix

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

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSiteKey = "0x4AAAA-test-sitekey"
	testToken   = "0.test-token"
	testProof   = "1793605772350.test-proof"
)

// fakeWidgetSolver answers both gates so it can stand in for cfsolve.
type fakeWidgetSolver struct {
	mu       sync.Mutex
	requests []netx.TurnstileRequest
	token    string
	err      error
}

func (s *fakeWidgetSolver) SolveChallenge(context.Context, string, time.Duration, netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	return &netx.ChallengeSolveResult{}, nil
}

func (s *fakeWidgetSolver) SolveTurnstile(_ context.Context, r netx.TurnstileRequest, _ time.Duration) (string, error) {
	s.mu.Lock()
	s.requests = append(s.requests, r)
	s.mu.Unlock()
	if s.err != nil {
		return "", s.err
	}
	if s.token == "" {
		return testToken, nil
	}
	return s.token, nil
}

func (s *fakeWidgetSolver) lastRequest() (netx.TurnstileRequest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) == 0 {
		return netx.TurnstileRequest{}, false
	}
	return s.requests[len(s.requests)-1], true
}

func registerWidgetSolver(t *testing.T, s netx.ChallengeSolver) {
	t.Helper()
	netx.RegisterChallengeSolver(s)
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
}

func clearChallengeSolver(t *testing.T) {
	t.Helper()
	netx.RegisterChallengeSolver(nil)
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
}

// playerPageWith renders the solved page's __PLAYER_PARAMS__ string from an
// url.Values built by the test.
func playerPageWith(sources []playerSource) string {
	encoded, _ := json.Marshal(sources)
	params := url.Values{}
	params.Set("sources", string(encoded))
	if len(sources) > 0 {
		params.Set("url", sources[0].URL)
	}
	return `<!doctype html><html><body><script>window.__PLAYER_PARAMS__ = ` +
		strconv.Quote(params.Encode()) + `;</script></body></html>`
}

func TestResolvePlayerClearsTheTurnstileAndKeepsTheProofAsReferer(t *testing.T) {
	clearChallengeSolver(t)
	solver := &fakeWidgetSolver{}
	registerWidgetSolver(t, solver)

	proxy := "/api/proxy?t=abc&ext=mp4"
	var gateHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/filme/310":
			gateHits.Add(1)
			if r.URL.Query().Get("tsv") == "" {
				fmt.Fprintf(w, `<!doctype html><html><body>
<div class="cf-turnstile" data-sitekey="%s" data-action="embed" data-cdata="cd-1"></div>
</body></html>`, testSiteKey)
				return
			}
			_, _ = fmt.Fprint(w, playerPageWith([]playerSource{{
				Kind: "stream", Label: "Dublado", URL: proxy,
			}}))
		case "/api/turnstile-verify":
			assert.Equal(t, http.MethodPost, r.Method)
			var body struct {
				Token string `json:"token"`
			}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, testToken, body.Token)
			_, _ = fmt.Fprintf(w, `{"success":true,"proof":%q}`, testProof)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	stream, err := c.ResolvePlayer(context.Background(), srv.URL+"/filme/310",
		PlayerTarget{TMDBID: 310, Kind: KindMovie}, AudioDubbed)

	require.NoError(t, err)
	assert.Equal(t, srv.URL+proxy, stream.URL)
	assert.Equal(t, srv.URL+"/filme/310?tsv="+testProof, stream.Referer)
	assert.Equal(t, AudioDubbed, stream.Audio)
	assert.Equal(t, int32(2), gateHits.Load(), "the gate page is fetched once, then once with the proof")

	req, ok := solver.lastRequest()
	require.True(t, ok, "the widget was not sent to the solver")
	assert.Equal(t, srv.URL+"/filme/310", req.PageURL)
	assert.Equal(t, testSiteKey, req.SiteKey)
	assert.Equal(t, "embed", req.Action)
	assert.Equal(t, "cd-1", req.CData)
}

func TestResolvePlayerWithoutSolverTouchesNothing(t *testing.T) {
	clearChallengeSolver(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))

	_, err := c.ResolvePlayer(context.Background(), srv.URL+"/filme/310",
		PlayerTarget{TMDBID: 310, Kind: KindMovie}, AudioDubbed)

	require.ErrorIs(t, err, ErrPlayerSolverMissing)
	assert.Zero(t, hits.Load(), "no solver means no request at all")
}

// The live catalog currently lists /api/player?id=<tmdb>, which answers 404
// "Vídeo indisponível". The resolver must rebuild the player page from the
// TMDB id instead of failing the title.
func TestResolvePlayerRebuildsADeadCatalogPlayerEndpoint(t *testing.T) {
	clearChallengeSolver(t)
	solver := &fakeWidgetSolver{}
	registerWidgetSolver(t, solver)

	proxy := "/api/proxy?t=dead&ext=mp4"
	var fallbackHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/player":
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, "Vídeo indisponível")
		case "/filme/310":
			fallbackHits.Add(1)
			if r.URL.Query().Get("tsv") == "" {
				fmt.Fprintf(w, `<div class="cf-turnstile" data-sitekey="%s"></div>`, testSiteKey)
				return
			}
			_, _ = fmt.Fprint(w, playerPageWith([]playerSource{{
				Kind: "stream", Label: "Dublado", URL: proxy,
			}}))
		case "/api/turnstile-verify":
			_, _ = fmt.Fprintf(w, `{"success":true,"proof":%q}`, testProof)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	c.playerBase = srv.URL
	stream, err := c.ResolvePlayer(context.Background(), srv.URL+"/api/player?id=310",
		PlayerTarget{TMDBID: 310, Kind: KindMovie}, AudioDubbed)

	require.NoError(t, err)
	assert.Equal(t, srv.URL+proxy, stream.URL)
	assert.Equal(t, int32(2), fallbackHits.Load(), "the constructed page is the one that is gated and solved")
}

func TestResolvePlayerPicksTheSourceForTheWantedAudio(t *testing.T) {
	clearChallengeSolver(t)
	solver := &fakeWidgetSolver{}
	registerWidgetSolver(t, solver)

	sources := []playerSource{
		{Kind: "stream", Label: "Legendado", URL: "/api/proxy?t=leg&ext=mp4"},
		{Kind: "stream", Label: "Dublado", URL: "/api/proxy?t=dub&ext=mp4"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/filme/310" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, playerPageWith(sources))
	}))
	t.Cleanup(srv.Close)

	for _, tt := range []struct {
		want  Audio
		path  string
		audio Audio
	}{
		{AudioSubtitled, "/api/proxy?t=leg&ext=mp4", AudioSubtitled},
		{AudioDubbed, "/api/proxy?t=dub&ext=mp4", AudioDubbed},
	} {
		t.Run(string(tt.want), func(t *testing.T) {
			c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
			stream, err := c.ResolvePlayer(context.Background(), srv.URL+"/filme/310",
				PlayerTarget{TMDBID: 310, Kind: KindMovie}, tt.want)
			require.NoError(t, err)
			assert.Equal(t, srv.URL+tt.path, stream.URL)
			assert.Equal(t, tt.audio, stream.Audio)
		})
	}
	_, called := solver.lastRequest()
	assert.False(t, called, "a page already carrying the player params needs no solve")
}

func TestPickPlayerSourceRefusesEmbeds(t *testing.T) {
	embed := playerSource{Kind: "iframe", URL: "https://player.test/embed/7"}
	file := playerSource{Kind: "stream", URL: "https://cdn.test/a.mp4"}

	if _, ok := pickPlayerSource([]playerSource{embed}, AudioDubbed); ok {
		t.Fatal("an iframe was chosen as a stream")
	}
	got, ok := pickPlayerSource([]playerSource{embed, file}, AudioSubtitled)
	require.True(t, ok)
	assert.Equal(t, file, got, "the playable source wins over the embed")
}

func TestParseTurnstileAcceptsWidgetAttributesAndScriptRendering(t *testing.T) {
	siteKey, action, cdata, ok := parseTurnstile([]byte(`<div class="cf-turnstile" data-sitekey="0xWIDGET" data-action="embed" data-cdata="cd-1"></div>`))
	require.True(t, ok)
	assert.Equal(t, "0xWIDGET", siteKey)
	assert.Equal(t, "embed", action)
	assert.Equal(t, "cd-1", cdata)

	siteKey, action, cdata, ok = parseTurnstile([]byte(`<script>turnstile.render(el, {sitekey: "0xSCRIPT", action: "login", cdata: "cd-2"});</script>`))
	require.True(t, ok)
	assert.Equal(t, "0xSCRIPT", siteKey)
	assert.Equal(t, "login", action)
	assert.Equal(t, "cd-2", cdata)

	_, _, _, ok = parseTurnstile([]byte(`<html><body>no widget here</body></html>`))
	assert.False(t, ok)
}
