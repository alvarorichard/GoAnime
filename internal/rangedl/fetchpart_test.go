package rangedl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countSink struct{ n atomic.Int64 }

func (s *countSink) AddProgressReceived(d int64) { s.n.Add(d) }

const payload = "0123456789abcdefghij"

// rangeServer serves payload honouring Range, and records the Range headers.
func rangeServer(t *testing.T, onRequest func(n int, w http.ResponseWriter, r *http.Request) bool) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		if onRequest != nil && onRequest(n, w, r) {
			return
		}
		var from, to int
		_, err := fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-%d", &from, &to)
		if err != nil {
			http.Error(w, "no range", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte(payload[from : to+1]))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func partFile(t *testing.T, dest string, i int) string {
	t.Helper()
	p, err := PartPath(dest, i)
	require.NoError(t, err)
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

func TestFetchPartWritesItsRangeAndReportsProgress(t *testing.T) {
	srv, _ := rangeServer(t, nil)
	dest := filepath.Join(t.TempDir(), "v.mp4")
	sink := &countSink{}
	prepared := 0

	err := FetchPart(Part{URL: srv.URL, From: 5, To: 14, Index: 1, Dest: dest}, srv.Client(), Options{
		Prepare:  func(r *http.Request) { prepared++; r.Header.Set("X-Test", "1") },
		Progress: sink,
	})

	require.NoError(t, err)
	assert.Equal(t, payload[5:15], partFile(t, dest, 1))
	assert.Equal(t, int64(10), sink.n.Load())
	assert.Equal(t, 1, prepared, "Prepare runs once per request")
}

func TestFetchPartResumesFromWhereTheConnectionDropped(t *testing.T) {
	var seen []string
	srv, calls := rangeServer(t, func(n int, w http.ResponseWriter, r *http.Request) bool {
		seen = append(seen, r.Header.Get("Range"))
		if n == 1 { // promise 10 bytes, send 4, then drop
			w.Header().Set("Content-Length", "10")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte(payload[0:4]))
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		}
		return false
	})
	dest := filepath.Join(t.TempDir(), "v.mp4")
	sink := &countSink{}

	err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: dest}, srv.Client(), Options{Progress: sink, RetryDelay: time.Millisecond})

	require.NoError(t, err)
	assert.Equal(t, payload[:10], partFile(t, dest, 0))
	assert.Equal(t, int32(2), calls.Load())
	assert.Equal(t, []string{"bytes=0-9", "bytes=4-9"}, seen, "the retry asks only for what is missing")
	assert.Equal(t, int64(10), sink.n.Load())
}

func TestFetchPartExpiredLinkFailsFastWithoutRetrying(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		srv, calls := rangeServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) bool {
			w.WriteHeader(status)
			return true
		})
		err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: filepath.Join(t.TempDir(), "v.mp4")}, srv.Client(), Options{RetryDelay: time.Millisecond})
		require.Error(t, err, "status %d", status)
		assert.Equal(t, int32(1), calls.Load(), "status %d is not retried", status)
	}
}

func TestFetchPartGivesUpAfterTheStaleBudget(t *testing.T) {
	srv, calls := rangeServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) bool {
		w.WriteHeader(http.StatusInternalServerError)
		return true
	})
	err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: filepath.Join(t.TempDir(), "v.mp4")}, srv.Client(), Options{RetryDelay: time.Microsecond})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "max retries"), err.Error())
	assert.Equal(t, int32(20), calls.Load())
}

func TestFetchPartProgressResetsTheStaleBudget(t *testing.T) {
	// 15 dead answers, 3 good bytes then a drop, 15 more dead answers, then the
	// rest. Without a reset the 30 failures would blow the budget of 20.
	srv, calls := rangeServer(t, func(n int, w http.ResponseWriter, r *http.Request) bool {
		switch {
		case n <= 15, n >= 17 && n <= 31:
			w.WriteHeader(http.StatusInternalServerError)
			return true
		case n == 16:
			w.Header().Set("Content-Length", "10")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte(payload[0:3]))
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		}
		return false
	})
	dest := filepath.Join(t.TempDir(), "v.mp4")

	err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: dest}, srv.Client(), Options{RetryDelay: time.Microsecond})

	require.NoError(t, err)
	assert.Equal(t, payload[:10], partFile(t, dest, 0))
	assert.Equal(t, int32(32), calls.Load())
}

func TestFetchPartHonoursTheRetryDelay(t *testing.T) {
	srv, _ := rangeServer(t, func(n int, w http.ResponseWriter, _ *http.Request) bool {
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}
		return false
	})
	start := time.Now()
	err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: filepath.Join(t.TempDir(), "v.mp4")}, srv.Client(), Options{RetryDelay: time.Millisecond})
	require.NoError(t, err)
	assert.Less(t, time.Since(start), DefaultRetryDelay/2, "a short RetryDelay must not fall back to the default")
}

func TestFetchPartDefaultsTheRetryDelay(t *testing.T) {
	srv, _ := rangeServer(t, func(n int, w http.ResponseWriter, _ *http.Request) bool {
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}
		return false
	})
	start := time.Now()
	err := FetchPart(Part{URL: srv.URL, From: 0, To: 9, Dest: filepath.Join(t.TempDir(), "v.mp4")}, srv.Client(), Options{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), DefaultRetryDelay, "an unset RetryDelay waits the default")
}
