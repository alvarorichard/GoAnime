package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type blockedStreamSniffer struct {
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
	once    sync.Once
	result  *CFStreamResult
}

func (s *blockedStreamSniffer) SniffEmbedStream(ctx context.Context, _ string, _ time.Duration) (*CFStreamResult, error) {
	if call := s.calls.Add(1); call > 1 {
		return nil, fmt.Errorf("unexpected duplicate browser resolve: call %d", call)
	}
	s.once.Do(func() { close(s.started) })
	select {
	case <-s.release:
		return s.result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (*blockedStreamSniffer) Solve(context.Context, string, time.Duration) (*CFSolveResult, error) {
	return nil, fmt.Errorf("unexpected generic challenge solve")
}

// The cache recheck after acquiring the per-content gate is the user-visible
// payoff: a simultaneous request must reuse the first browser resolution.
func TestGetStreamURLConcurrentSameEpisodeReusesResolvedCache(t *testing.T) {
	withFreshStreamCache(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/master.txt":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\n"))
		case "/video/hash123":
			_, _ = w.Write([]byte("<html><title>Test</title></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	solver := &blockedStreamSniffer{
		started: make(chan struct{}),
		release: make(chan struct{}),
		result: &CFStreamResult{
			StreamURL:  srv.URL + "/master.txt",
			Referer:    srv.URL + "/video/hash123",
			UserAgent:  SuperFlixUserAgent,
			PlayerHost: srv.URL,
			VideoHash:  "hash123",
		},
	}
	client := NewSuperFlixClient()
	client.browserSolver = solver
	client.baseURL = "https://superflix.test"
	client.client = &http.Client{Timeout: 5 * time.Second, Transport: http.DefaultTransport}

	type resolveResult struct {
		result *SuperFlixStreamResult
		err    error
	}
	firstDone := make(chan resolveResult, 1)
	go func() {
		result, err := client.GetStreamURL(context.Background(), "filme", "42", "", "")
		firstDone <- resolveResult{result: result, err: err}
	}()

	select {
	case <-solver.started:
	case <-time.After(time.Second):
		close(solver.release)
		t.Fatal("first request did not start browser resolution")
	}

	secondCtx := newObservedDoneContext(context.Background())
	secondDone := make(chan resolveResult, 1)
	go func() {
		result, err := client.GetStreamURL(secondCtx, "filme", "42", "", "")
		secondDone <- resolveResult{result: result, err: err}
	}()
	select {
	case <-secondCtx.observed:
	case <-time.After(time.Second):
		close(solver.release)
		t.Fatal("second request did not reach the per-content gate")
	}

	close(solver.release)
	for name, done := range map[string]<-chan resolveResult{"first": firstDone, "second": secondDone} {
		select {
		case got := <-done:
			if got.err != nil {
				t.Errorf("%s resolve failed: %v", name, got.err)
				continue
			}
			if got.result == nil || !strings.HasSuffix(got.result.StreamURL, "/master.txt") {
				t.Errorf("%s resolve returned an unexpected stream: %#v", name, got.result)
			}
		case <-time.After(3 * time.Second):
			t.Errorf("%s resolve did not finish", name)
		}
	}
	if calls := solver.calls.Load(); calls != 1 {
		t.Fatalf("browser resolve calls = %d, want 1", calls)
	}
}
