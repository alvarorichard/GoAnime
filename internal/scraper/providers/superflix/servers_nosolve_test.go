package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// solveFlagTransport records, per request, whether the browser solve was
// forbidden on that request's context. It is how we prove — without a live
// Cloudflare gate — that the server-list path can never trigger a headed-browser
// solve.
type solveFlagTransport struct {
	base http.RoundTripper
	mu   sync.Mutex
	seen map[string]bool // url path -> browserSolveForbidden(ctx)
}

func (t *solveFlagTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	if t.seen == nil {
		t.seen = map[string]bool{}
	}
	// Record the strictest observation: if any request on a path was NOT
	// forbidden, that path is unsafe.
	if prev, ok := t.seen[req.URL.Path]; !ok || prev {
		t.seen[req.URL.Path] = browserSolveForbidden(req.Context())
	}
	t.mu.Unlock()
	return t.base.RoundTrip(req)
}

// TestGetServers_AllowsBrowserSolve pins the correction of a mistaken "fix":
// getting the server list REQUIRES the Cloudflare solve, because the tokened
// player page is gated (measured live — without the solve the page never carries
// tokens and the whole feature is dead). An earlier version forbade the solve here
// to dodge a hang; the hang was really a client-rebuild bug (fixed separately), so
// forbidding the solve only removed the feature. GetServers must therefore NOT
// mark its requests solve-forbidden.
func TestGetServers_AllowsBrowserSolve(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/bootstrap") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, sfBootstrapJSON)
			return
		}
		_, _ = fmt.Fprint(w, sfRealPlayerPage)
	}))
	t.Cleanup(srv.Close)

	tr := &solveFlagTransport{base: http.DefaultTransport}
	c := NewClientForTest(srv.URL)
	c.client = &http.Client{Transport: tr}

	_, _, err := c.GetServers(context.Background(), "serie", "103913", "1", "1")
	require.NoError(t, err)

	tr.mu.Lock()
	defer tr.mu.Unlock()
	require.NotEmpty(t, tr.seen, "GetServers must have made requests")
	for path, forbidden := range tr.seen {
		assert.False(t, forbidden, "request to %s must be allowed to solve — the tokened page is gated", path)
	}
}

// StreamFromServer resolves the actual payload, so it MUST stay free to solve the
// gate if it has to. Locking the enhancement's no-solve rule must not accidentally
// gag the stream path.
func TestStreamFromServer_DoesNotForbidBrowserSolve(t *testing.T) {
	// Not parallel: swaps the global stream cache. Without the swap this test
	// wrote its httptest (host, hash) into the USER'S real on-disk cache.
	withFreshStreamCache(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/player/source"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data":{"video_url":"%s/video/hash123"}}`, srvBase(r))
		case strings.HasPrefix(r.URL.Path, "/video/"):
			_, _ = fmt.Fprint(w, realPlayerPage)
		case strings.Contains(r.URL.Path, "/player/index.php"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"securedLink":"https://cdn/x.m3u8"}`)
		default:
			_, _ = fmt.Fprint(w, "ok")
		}
	}))
	t.Cleanup(srv.Close)

	tr := &solveFlagTransport{base: http.DefaultTransport}
	c := NewClientForTest(srv.URL)
	c.client = &http.Client{Transport: tr}

	tokens := &SuperFlixTokens{ContentID: "1", PageToken: "tok"}
	_, err := c.StreamFromServer(context.Background(), tokens, "159462", "serie", "1", "1", "1")
	require.NoError(t, err)

	tr.mu.Lock()
	defer tr.mu.Unlock()
	for path, forbidden := range tr.seen {
		assert.False(t, forbidden, "stream request to %s must be allowed to solve the gate", path)
	}
}

// srvBase returns scheme+host of the test server from a request.
func srvBase(r *http.Request) string {
	return "http://" + r.Host
}
