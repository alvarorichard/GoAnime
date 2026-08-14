// ===========================================================================
// download_surf_test.go — Regression tests for the surf HTTP client fixes
//
// Issue #193: downloads failed with
//   "surf HTTP/2 failed and cannot retry because req GetBody is nil:
//    surf negotiated ALPN expected h2"
// on Blogger/googlevideo URLs. googlevideo edges omit ALPN, so surf's HTTP/2
// dial fails; requests built with http.NoBody have nil GetBody (see
// net/http.NewRequest, NoBody hits the default branch), which blocks surf's
// HTTP/1.1 fallback.
//
// Fixes under test:
//   1. newSurfDownloadClient() forces HTTP/1.1 (scraper.go) — the Blogger
//      download path never touches the HTTP/2 transport at all.
//   2. newNoBodyRequest() in downloader/hls sets GetBody, so HLS downloads
//      through h2-capable clients can fall back to HTTP/1.1.
//
// Every TLS server here refuses HTTP/2 (only negotiates http/1.1), exactly
// like the googlevideo edges from the issue.
// ===========================================================================

package player

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/enetx/surf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newHTTP11OnlyTLSServer starts an HTTPS test server that only negotiates
// HTTP/1.1, mirroring the googlevideo edges that omit ALPN and triggered
// issue #193. The TLS dial sanity check guarantees the server really refuses
// HTTP/2, otherwise the tests would pass trivially without exercising the
// fallback paths they are meant to guard.
func newHTTP11OnlyTLSServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(handler)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	conn, err := tls.Dial("tcp", srv.Listener.Addr().String(), &tls.Config{
		InsecureSkipVerify: true, // #nosec G402 -- test-only, localhost
		NextProtos:         []string{"h2", "http/1.1"},
	})
	require.NoError(t, err)
	negotiated := conn.ConnectionState().NegotiatedProtocol
	_ = conn.Close()
	require.Equal(t, "http/1.1", negotiated, "test server must offer HTTP/1.1 only")
	return srv
}

// TestSurfClient_NoBody_FailsOnHTTP11OnlyServer is the negative control for
// all the HTTP/1.1-only server tests: a plain h2-capable surf client with an
// http.NoBody request (nil GetBody) must still fail with the exact issue #193
// error against these servers. If this test ever stops failing, the other
// tests in this file no longer reproduce the original bug and must be
// revisited.
func TestSurfClient_NoBody_FailsOnHTTP11OnlyServer(t *testing.T) {
	t.Parallel()
	srv := newHTTP11OnlyTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	client := surf.NewClient().
		Builder().
		Impersonate().Chrome().
		Timeout(10 * time.Second).
		Build().
		Unwrap().
		Std()

	req, err := http.NewRequest(http.MethodHead, srv.URL+"/video.mp4", http.NoBody)
	require.NoError(t, err)

	_, err = client.Do(req)
	require.Error(t, err, "control: h2-capable surf client must fail on HTTP/1.1-only servers")
	assert.Contains(t, err.Error(), "GetBody is nil", "control: must reproduce issue #193 exactly")
	assert.Contains(t, err.Error(), "negotiated ALPN", "control: must reproduce issue #193 exactly")
}

// TestGetContentLength_SurfClient_HTTP11OnlyServer covers the exact failure
// point from issue #193: the Blogger download path's HEAD length probe
// ("failed to get content length: ... surf negotiated ALPN expected h2").
// newSurfDownloadClient forces HTTP/1.1, so the probe must succeed against
// servers that only negotiate http/1.1.
func TestGetContentLength_SurfClient_HTTP11OnlyServer(t *testing.T) {
	t.Parallel()
	srv := newHTTP11OnlyTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "12345")
		w.WriteHeader(http.StatusOK)
	}))

	client := newSurfDownloadClient().Std()
	length, err := getContentLength(srv.URL+"/video.mp4", client)

	require.NoError(t, err, "download client must probe content length over HTTP/1.1-only edges")
	assert.Equal(t, int64(12345), length)
}

// TestGetContentLength_SurfClient_RangeFallback covers the HEAD-not-supported
// fallback: when the server rejects HEAD (405), getContentLength retries with
// a Range GET (bytes=0-0). This must also work through the forced-HTTP/1.1
// download client.
func TestGetContentLength_SurfClient_RangeFallback(t *testing.T) {
	t.Parallel()
	srv := newHTTP11OnlyTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Length", "1")
		w.Header().Set("Content-Range", "bytes 0-0/12345")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte{0})
	}))

	client := newSurfDownloadClient().Std()
	length, err := getContentLength(srv.URL+"/video.mp4", client)

	require.NoError(t, err)
	assert.Equal(t, int64(1), length)
}

// TestGetContentLength_SurfClient_FollowsRedirects covers googlevideo's 302
// redirect behavior (the download client is built to follow redirects, unlike
// newSurfClient).
func TestGetContentLength_SurfClient_FollowsRedirects(t *testing.T) {
	t.Parallel()
	var redirects atomic.Int32
	srv := newHTTP11OnlyTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			redirects.Add(1)
			http.Redirect(w, r, "/video.mp4", http.StatusFound)
			return
		}
		w.Header().Set("Content-Length", "777")
		w.WriteHeader(http.StatusOK)
	}))

	client := newSurfDownloadClient().Std()
	length, err := getContentLength(srv.URL+"/redirect", client)

	require.NoError(t, err)
	assert.Equal(t, int64(777), length)
	assert.Greater(t, redirects.Load(), int32(0), "download client must follow googlevideo-style redirects")
}
