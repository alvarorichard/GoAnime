// Package leak_test runs GoAnime's concurrent flows end to end and asserts that
// none of them strands a goroutine, using Go 1.27's goroutineleak profile.
//
// This is the CI gate: it lives in its own package so the profile is taken over
// a binary that only contains the flows under test, keeping the baseline clean.
package leak_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/leakcheck"
)

// hlsServer serves a small multi-segment playlist.
func hlsServer(t *testing.T, segments int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			var b strings.Builder
			b.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:5\n")
			for i := range segments {
				fmt.Fprintf(&b, "#EXTINF:5.0,\nchunk%d.ts\n", i)
			}
			b.WriteString("#EXT-X-ENDLIST\n")
			_, _ = w.Write([]byte(b.String()))
			return
		}
		_, _ = w.Write(make([]byte, 4096))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestHLSDownloadLeavesNoLeakedGoroutines exercises the segment fan-out, which
// is the most goroutine-heavy path in the app.
func TestHLSDownloadLeavesNoLeakedGoroutines(t *testing.T) {
	baseline := leakcheck.Count(t)
	srv := hlsServer(t, 12)

	out := filepath.Join(t.TempDir(), "out.ts")
	err := hls.DownloadToFileWithClient(context.Background(), srv.Client(),
		srv.URL+"/video.m3u8", out, nil, func(int64, int, int) {})
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("expected a non-empty output file, got %v (%v)", info, err)
	}

	leakcheck.AssertNoNewLeaks(t, baseline)
}

// TestCancelledHLSDownloadLeavesNoLeakedGoroutines is the important half: the
// happy path rarely leaks, abandoned work does. The server stalls so the
// download is still in flight when the context is cancelled.
func TestCancelledHLSDownloadLeavesNoLeakedGoroutines(t *testing.T) {
	baseline := leakcheck.Count(t)

	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			var b strings.Builder
			b.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:5\n")
			for i := range 40 {
				fmt.Fprintf(&b, "#EXTINF:5.0,\nchunk%d.ts\n", i)
			}
			b.WriteString("#EXT-X-ENDLIST\n")
			_, _ = w.Write([]byte(b.String()))
			return
		}
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	out := filepath.Join(t.TempDir(), "cancelled.ts")

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = hls.DownloadToFileWithClient(ctx, srv.Client(),
			srv.URL+"/video.m3u8", out, nil, nil)
	}()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("cancelled download did not return")
	}
	close(release)

	leakcheck.AssertNoNewLeaks(t, baseline)
}

// TestResponseCacheLeavesNoLeakedGoroutines covers the shared cache used by the
// metadata paths, hammered concurrently.
func TestResponseCacheLeavesNoLeakedGoroutines(t *testing.T) {
	baseline := leakcheck.Count(t)

	cache := util.NewResponseCache(200*time.Millisecond, 64)
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Go(func() {
			key := fmt.Sprintf("key-%d", i%8)
			cache.Set(key, []byte("payload"))
			if _, ok := cache.Get(key); !ok {
				// A miss right after a Set is possible once entries expire;
				// the assertion here is about goroutines, not cache hits.
				_ = ok
			}
		})
	}
	wg.Wait()

	leakcheck.AssertNoNewLeaks(t, baseline)
}

// TestConcurrentFlowsLeaveNoLeakedGoroutines runs the flows together, which is
// how they actually run in the app.
func TestConcurrentFlowsLeaveNoLeakedGoroutines(t *testing.T) {
	baseline := leakcheck.Count(t)
	srv := hlsServer(t, 6)

	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() {
			out := filepath.Join(t.TempDir(), fmt.Sprintf("out-%d.ts", i))
			if err := hls.DownloadToFileWithClient(context.Background(), srv.Client(),
				srv.URL+"/video.m3u8", out, nil, nil); err != nil {
				t.Errorf("download %d: %v", i, err)
			}
		})
	}
	wg.Wait()

	leakcheck.AssertNoNewLeaks(t, baseline)
}
