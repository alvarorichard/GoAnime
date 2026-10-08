package player

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLocalProxyWaitsLongEnoughForHeaders: the StartFlix proxy answers a
// range only once it holds the first part of it — a 2 MiB fetch from the CDN,
// which it bounds itself (30 s with retries) — so that a dead origin is a 502
// rather than a truncated 206. The downloader gave it 10 s: under a batch's
// load that ran out ("net/http: timeout awaiting response headers"), the
// range was retried from scratch, and the work already done was thrown away.
func TestLocalProxyWaitsLongEnoughForHeaders(t *testing.T) {
	t.Parallel()
	tr, ok := downloadTransport(10 * time.Second).(localProxyAware)
	if !ok {
		t.Fatal("downloadTransport is not the local-proxy-aware transport")
	}
	local, ok := tr.local.(*http.Transport)
	if !ok {
		t.Fatal("local transport is not an *http.Transport")
	}
	// Twice the proxy's own 30 s bound on fetching a part, retries included.
	if got := local.ResponseHeaderTimeout; got < 60*time.Second {
		t.Errorf("local proxy header timeout = %v, want at least 60s", got)
	}
}

// TestDownloadVideoReadsALocalProxyAsOneRange: the proxy fetches an Abyss
// file's parts in parallel already (read-ahead), and every extra range is
// another reader with its own read-ahead. Four ranges per episode, four
// episodes at once, put ~28 concurrent requests on the CDN, each part got
// slower than the header timeout, and readers kept restarting.
func TestDownloadVideoReadsALocalProxyAsOneRange(t *testing.T) {
	restore := installDownloadRangeTestState(t.TempDir())
	defer restore()
	home := t.TempDir()
	t.Setenv("HOME", home)
	body := []byte(strings.Repeat("v", 1<<16))
	var mu sync.Mutex
	var ranges []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.Header.Get("Range") != "" {
			mu.Lock()
			ranges = append(ranges, r.Header.Get("Range"))
			mu.Unlock()
		}
		http.ServeContent(w, r, "v.mp4", time.Time{}, strings.NewReader(string(body)))
	}))
	t.Cleanup(srv.Close)
	u, _ := url.Parse(srv.URL)
	util.RegisterLocalProxyHost(u.Host)

	out := filepath.Join(home, "dl", "ep.mp4")
	if err := os.MkdirAll(filepath.Dir(out), 0o700); err != nil { // createEpisodePath does this in a batch
		t.Fatal(err)
	}
	if err := DownloadVideo(srv.URL+"/abyss/x.mp4", out, 4, nil); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(ranges) != 1 {
		t.Errorf("the local proxy was read as %d ranges %v, want one", len(ranges), ranges)
	}
}
