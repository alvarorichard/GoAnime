package superflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The multi-source dispatcher gives each source a deadline and abandons the
// goroutine when it fires, so a search whose HTTP request ignores the context
// keeps running unseen. For SuperFlix that meant the transport's Retry-After
// loop slept and re-requested for ~30s after the search had already given up,
// against the very host that was rate-limiting it — and the next search stacked
// another one on top ("dexter", 2026-09-20).
//
// SearchMediaWithContext must therefore honor cancellation end to end.
func TestSearchMediaWithContext_HonorsCancellation(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	var served atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served.Add(1)
		select {
		case <-release:
		case <-r.Context().Done(): // client went away
		}
	}))
	defer srv.Close()
	defer close(release)

	c := NewClientForTest(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.SearchMediaWithContext(ctx, "dexter")

	require.Error(t, err, "a cancelled search must return, not hang on the request")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second, "the request must be torn down with the context")
	assert.Equal(t, int32(1), served.Load())
}

// The adapter is what the registry provider calls, and it is where the context
// used to be dropped (context.Background() via SearchMedia). Pin that the
// contextual entry point exists and threads the caller's context through.
func TestSuperFlixClient_SearchMediaHasContextualEntryPoint(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body></body></html>`))
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)

	// An already-cancelled context must not produce a request at all.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.SearchMediaWithContext(ctx, "dexter")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
