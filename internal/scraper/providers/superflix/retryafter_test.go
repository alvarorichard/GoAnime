package superflix

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	tr := &cfFallbackTransport{base: http.DefaultTransport}
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
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("still limited"))
	}))
	t.Cleanup(srv.Close)

	tr := &cfFallbackTransport{base: http.DefaultTransport}
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

	tr := &cfFallbackTransport{base: http.DefaultTransport}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(srv.URL, "text/plain", strings.NewReader("x"))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(1), hits.Load(), "POST 429 not transparently retried")
}

func TestHonorRetryAfter429_ContextCancel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	tr := &cfFallbackTransport{base: http.DefaultTransport}
	client := &http.Client{Transport: tr}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)

	_, err := client.Do(req)
	require.Error(t, err, "context cancellation during the wait must abort")
}
