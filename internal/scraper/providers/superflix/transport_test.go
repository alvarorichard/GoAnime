package superflix

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCFFallbackTransport_WithoutBrowserSolveSkipsSolver pins the search-path
// contract: a request carrying WithoutBrowserSolve never escalates a CF
// challenge to the headed browser — the challenge response is returned as-is.
func TestCFFallbackTransport_WithoutBrowserSolveSkipsSolver(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<html><body>blocked</body></html>`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequestWithContext(WithoutBrowserSolve(t.Context()), "GET", srv.URL, http.NoBody)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 403, resp.StatusCode, "challenge must be handed back untouched")
	assert.Contains(t, string(body), "blocked")
	assert.Equal(t, int32(0), solver.calls.Load(), "browser solver must NOT run for no-solve requests")
}

// TestCFFallbackTransport_SolverStillRunsWithoutOptOut proves the same
// challenge DOES escalate when the caller did not opt out (play path).
func TestCFFallbackTransport_SolverStillRunsWithoutOptOut(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<html><body>blocked</body></html>`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{html: "<html>real content</html>"}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, int32(1), solver.calls.Load(), "solver must run when no opt-out is present")
}

// TestSearchMedia_ChallengedGateNeverOpensBrowser pins the end-to-end search
// behavior: a challenged SuperFlix search fails as a plain error, without the
// solver ever being invoked.
func TestSearchMedia_ChallengedGateNeverOpensBrowser(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<html><body>blocked</body></html>`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{}
	c := NewClientForTest(srv.URL)
	c.client = &http.Client{Transport: &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}}

	_, err := c.SearchMedia("naruto")
	require.Error(t, err, "challenged search must fail instead of opening a browser")
	assert.Equal(t, int32(0), solver.calls.Load(), "search must never invoke the browser solver")
}

// TestBrowserSolveForbidden_ContextPropagation pins the opt-out predicate:
// a plain context allows solving, WithoutBrowserSolve forbids it, and the
// flag is inherited by child contexts (so a derived request context — e.g.
// one with a timeout added downstream — still honors the search opt-out).
func TestBrowserSolveForbidden_ContextPropagation(t *testing.T) {
	t.Parallel()

	assert.False(t, browserSolveForbidden(context.Background()),
		"a plain context must allow the browser solver (play path)")

	noSolve := WithoutBrowserSolve(context.Background())
	assert.True(t, browserSolveForbidden(noSolve),
		"WithoutBrowserSolve must forbid the browser solver")

	child, cancel := context.WithCancel(noSolve)
	defer cancel()
	assert.True(t, browserSolveForbidden(child),
		"the opt-out must be inherited by derived contexts")
}

// TestGetStreamURL_PlayPathStillAllowsBrowser pins that the play path does NOT
// carry the opt-out — the browser stays available where the user explicitly
// chose SuperFlix content. It drives GetStreamURL against a challenged server
// with a fake solver and asserts the solver IS reached (regression guard so a
// future change doesn't over-apply WithoutBrowserSolve to the play path).
func TestGetStreamURL_PlayPathStillAllowsBrowser(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<html><body>blocked</body></html>`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{html: "<html>content</html>"}
	c := NewClientForTest(srv.URL)
	c.client = &http.Client{Transport: &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}}

	// GetStreamURL will fail overall (fake content yields no stream), but the
	// point is the solver must have been invoked — the play path is browser-eligible.
	_, _ = c.GetStreamURL(context.Background(), "filme", "1234", "", "")
	assert.Positive(t, solver.calls.Load(), "play path must remain able to open the browser solver")
}
