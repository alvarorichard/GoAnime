package superflix

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newGatedTransport builds the transport under test with a gate of its own.
//
// The production gate is process-wide on purpose — two clients must not each
// spend a request on a host that is already blocked — but inside one test
// binary that would make these tests share state: they all provoke a 429 on the
// same synthetic host, so one arming the gate would suppress the next one's
// request. A fresh gate per test keeps them independent without weakening what
// each one checks.
func newGatedTransport(base http.RoundTripper) *cfFallbackTransport {
	return &cfFallbackTransport{base: base, gate: newRateLimitGate("")}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	mk := func(v string) *http.Response {
		h := http.Header{}
		if v != "" {
			h.Set("Retry-After", v)
		}
		return &http.Response{Header: h}
	}

	assert.Equal(t, 120*time.Second, parseRetryAfter(mk("120")))
	assert.Equal(t, time.Duration(0), parseRetryAfter(mk("")), "absent header → 0")
	assert.Equal(t, time.Duration(0), parseRetryAfter(mk("0")), "zero → 0")
	assert.Equal(t, time.Duration(0), parseRetryAfter(mk("-5")), "negative → 0")
	assert.Equal(t, time.Duration(0), parseRetryAfter(mk("soon")), "garbage → 0")

	// HTTP-date in the future → a positive, bounded duration.
	future := time.Now().Add(3 * time.Second).UTC().Format(http.TimeFormat)
	d := parseRetryAfter(mk(future))
	assert.Greater(t, d, time.Duration(0))
	assert.LessOrEqual(t, d, 3*time.Second)

	// HTTP-date in the past → 0 (don't wait).
	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	assert.Equal(t, time.Duration(0), parseRetryAfter(mk(past)))
}

func TestHonorRetryAfter429_WaitsThenRetries(t *testing.T) {
	t.Parallel()
	// The Retry-After wait is real time; inside a synctest bubble the fake clock
	// skips it while still proving the transport waited the full second.
	synctest.Test(t, func(t *testing.T) {
		testHonorRetryAfter429WaitsThenRetries(t)
	})
}

func testHonorRetryAfter429WaitsThenRetries(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	tr := newGatedTransport(srv.Client().Transport)
	client := &http.Client{Transport: tr}

	start := time.Now()
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
	assert.Equal(t, int32(2), hits.Load(), "server hit twice (429 + honored retry)")
	assert.GreaterOrEqual(t, time.Since(start), time.Second, "waited out the Retry-After")
}

func TestHonorRetryAfter429_GivesUpWithinBudget(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		testHonorRetryAfter429GivesUpWithinBudget(t)
	})
}

func testHonorRetryAfter429GivesUpWithinBudget(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("still limited"))
	}))

	tr := newGatedTransport(srv.Client().Transport)
	client := &http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Persistent 429: bounded retries, then the 429 is returned (not hung forever).
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.LessOrEqual(t, hits.Load(), int32(maxRetryAfterTries+1))
	assert.GreaterOrEqual(t, hits.Load(), int32(2), "retried at least once")
}

func TestHonorRetryAfter429_POSTNotRetried(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	tr := newGatedTransport(http.DefaultTransport)
	client := &http.Client{Transport: tr}

	resp, err := client.Post(srv.URL, "text/plain", strings.NewReader("x"))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(1), hits.Load(), "POST 429 not transparently retried")
}

func TestHonorRetryAfter429_ContextCancel(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		testHonorRetryAfter429ContextCancel(t)
	})
}

// A context with no deadline still waits out the Retry-After (there is nothing
// to say the wait is futile), so a cancel arriving mid-wait must abort it
// rather than sleep out the full 30s.
func testHonorRetryAfter429ContextCancel(t *testing.T) {
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))

	tr := newGatedTransport(srv.Client().Transport)
	client := &http.Client{Transport: tr}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(200*time.Millisecond, cancel)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)

	_, err := client.Do(req)
	require.Error(t, err, "context cancellation during the wait must abort")
}

// Regression for the "dexter" search (2026-09-20): SuperFlix answered 429 with
// Retry-After 10s while the multi-source search allows 12s per source. The
// transport slept 10s, retried into another 429, and was still sleeping when
// the search gave up — so the user waited the full budget to be told it timed
// out, and the abandoned loop kept requesting against a host that was
// rate-limiting precisely because of that traffic.
//
// A best-effort request (the same class that may not open the headed browser)
// no longer waits at all: the 429 goes straight back, once.
func TestHonorRetryAfter429_BestEffortRequestDoesNotWait(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(http.StatusTooManyRequests)
		}))

		tr := newGatedTransport(srv.Client().Transport)
		client := &http.Client{Transport: tr}

		ctx, cancel := context.WithTimeout(WithoutBrowserSolve(context.Background()), 12*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)

		start := time.Now()
		resp, err := client.Do(req)
		require.NoError(t, err, "the 429 must come back as a response, not a timeout")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, int32(1), hits.Load(), "a doomed retry must not add load to a rate-limited host")
		assert.Less(t, time.Since(start), time.Second, "search must not spend its budget on a wait it cannot finish")
	})
}

// The play path — where the user explicitly chose SuperFlix content and the
// budget is 210s — still waits the rate limit out, because that is what
// actually clears it.
func TestHonorRetryAfter429_PlayPathStillWaits(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hits.Add(1) == 1 {
				w.Header().Set("Retry-After", "10")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		tr := newGatedTransport(srv.Client().Transport)
		client := &http.Client{Transport: tr}

		ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "the retry after the wait must be taken")
		assert.Equal(t, int32(2), hits.Load())
	})
}

// A wait that cannot fit in the caller's remaining deadline is skipped even on
// the play path: sleeping into a cancellation returns nothing and still costs
// the host a request.
func TestHonorRetryAfter429_SkipsWaitLongerThanDeadline(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
		}))

		tr := newGatedTransport(srv.Client().Transport)
		client := &http.Client{Transport: tr}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)

		resp, err := client.Do(req)
		require.NoError(t, err, "a futile wait must return the 429, not a timeout")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, int32(1), hits.Load())
	})
}
