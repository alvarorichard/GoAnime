// ===========================================================================
// blogger_proxy_readiness_test.go — Blogger-proxy regression tests
//
// Issues observed: 2026-07-23 (Goyabu debug logs)
//   The FIRST play of an episode failed with:
//     "mpv exited before IPC socket was ready: exit status 2 ... (no stderr)"
//     Hint: bundled mpv may be missing DLLs
//   The second attempt (next episode) played fine. mpv was healthy — the hint
//   was misleading.
//   Black Clover episode 3 later failed with:
//     "no valid video URL found"
//   after surf rejected a googlevideo edge that omitted ALPN:
//     "HTTP/2 failed ... negotiated ALPN \"\", expected h2"
//
// Root causes (player/scraper.go):
//   1. Redirect handling. The proxy's upstream client was built with
//      .NotFollowRedirects(); when the signed googlevideo URL 302-redirected to
//      a CDN node, surf surfaced that as a "net/http: use last response" error,
//      and the handler turned it into HTTP 502.
//   2. Readiness masking. The readiness probe treated ANY transport-level
//      success as ready — a 502 (headErr==nil) counted as ready — so mpv was
//      launched against a proxy serving 502 and exited 2.
//   3. Protocol mismatch. Chrome impersonation selected surf's HTTP/2 path,
//      but some googlevideo edges only serve HTTP/1.1 and omit ALPN.
//
// Fixes:
//   1. proxyClient.CheckRedirect = nil (follow redirects, net/http default).
//   2. Readiness rejects non-2xx upstream status; on the deadline it returns an
//      error so extractBloggerVideoURL's retry loop re-resolves a fresh CDN host
//      instead of handing mpv a dead stream.
//   3. The googlevideo client preserves Chrome TLS impersonation while forcing
//      HTTP/1.1.
//
// Function tested: startBloggerProxyServer (the network-free seam split out of
// startBloggerProxy so the forwarding + readiness gating are testable).
// ===========================================================================

package player

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okVideoHandler answers HEAD with 200 and GET with the given body — a stand-in
// for a healthy googlevideo CDN node.
func okVideoHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, body)
	}
}

// getBloggerProxyBody performs a GET against the proxy and returns the body.
func getBloggerProxyBody(t *testing.T, proxyURL string) string {
	t.Helper()
	resp, err := http.Get(proxyURL) //nolint:noctx // local proxy, test-only
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(b)
}

func TestStartBloggerProxyServer(t *testing.T) {
	// Mutates the bloggerReadiness* package vars and the bloggerProxy global —
	// not parallel, and subtests run sequentially so they don't fight over the
	// single shared proxy.
	prevTimeout, prevInterval := bloggerReadinessTimeout, bloggerReadinessInterval
	bloggerReadinessTimeout = 300 * time.Millisecond
	bloggerReadinessInterval = 10 * time.Millisecond
	t.Cleanup(func() {
		bloggerReadinessTimeout, bloggerReadinessInterval = prevTimeout, prevInterval
		StopBloggerProxy()
	})

	const videoBody = "FAKE-MP4-BYTES"

	t.Run("healthy 200 upstream is served", func(t *testing.T) {
		up := httptest.NewServer(okVideoHandler(videoBody))
		t.Cleanup(up.Close)
		t.Cleanup(StopBloggerProxy)

		proxyURL, err := startBloggerProxyServer(up.URL, &http.Client{})
		require.NoError(t, err)
		assert.Equal(t, videoBody, getBloggerProxyBody(t, proxyURL))
	})

	t.Run("follows an upstream redirect (regression: use-last-response -> 502)", func(t *testing.T) {
		final := httptest.NewServer(okVideoHandler(videoBody))
		t.Cleanup(final.Close)
		redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, final.URL, http.StatusFound)
		}))
		t.Cleanup(redir.Close)
		t.Cleanup(StopBloggerProxy)

		// A default client (CheckRedirect==nil) follows the redirect — this is
		// the production fix. Before it, the proxy 502'd on the redirect.
		proxyURL, err := startBloggerProxyServer(redir.URL, &http.Client{})
		require.NoError(t, err, "a redirect-following client must reach the real 200 video")
		assert.Equal(t, videoBody, getBloggerProxyBody(t, proxyURL))
	})

	t.Run("non-following client never yields a 2xx (the old bug shape)", func(t *testing.T) {
		final := httptest.NewServer(okVideoHandler(videoBody))
		t.Cleanup(final.Close)
		redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, final.URL, http.StatusFound)
		}))
		t.Cleanup(redir.Close)
		t.Cleanup(StopBloggerProxy)

		// Models the removed .NotFollowRedirects(): the proxy forwards the 302
		// (never the final 200), so readiness never observes a 2xx and fails.
		noFollow := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		_, err := startBloggerProxyServer(redir.URL, noFollow)
		require.Error(t, err, "without redirect-follow the proxy never serves a 2xx video")
	})

	t.Run("rejects a 502 upstream instead of masking it", func(t *testing.T) {
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		t.Cleanup(up.Close)
		t.Cleanup(StopBloggerProxy)

		_, err := startBloggerProxyServer(up.URL, &http.Client{})
		require.Error(t, err, "a 502-serving upstream must not be reported as ready")
		assert.Contains(t, err.Error(), "502")
	})

	t.Run("accepts a 206 partial-content upstream", func(t *testing.T) {
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			if r.Method == http.MethodGet {
				_, _ = io.WriteString(w, videoBody)
			}
		}))
		t.Cleanup(up.Close)
		t.Cleanup(StopBloggerProxy)

		_, err := startBloggerProxyServer(up.URL, &http.Client{})
		require.NoError(t, err, "206 is a valid streaming status and must count as ready")
	})

	t.Run("forwards byte ranges and partial-content metadata", func(t *testing.T) {
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Accept-Ranges", "bytes")
			if r.Method == http.MethodHead {
				w.Header().Set("Content-Length", "4096")
				w.WriteHeader(http.StatusOK)
				return
			}

			assert.Equal(t, "bytes=1024-2047", r.Header.Get("Range"))
			w.Header().Set("Content-Range", "bytes 1024-2047/4096")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, videoBody)
		}))
		t.Cleanup(up.Close)
		t.Cleanup(StopBloggerProxy)

		proxyURL, err := startBloggerProxyServer(up.URL, &http.Client{})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, proxyURL, http.NoBody)
		require.NoError(t, err)
		req.Header.Set("Range", "bytes=1024-2047")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusPartialContent, resp.StatusCode)
		assert.Equal(t, "video/mp4", resp.Header.Get("Content-Type"))
		assert.Equal(t, "bytes", resp.Header.Get("Accept-Ranges"))
		assert.Equal(t, "bytes 1024-2047/4096", resp.Header.Get("Content-Range"))
		assert.Equal(t, videoBody, string(body))
	})
}

func TestBloggerProxyClientSupportsHTTP11TLS(t *testing.T) {
	// This models the googlevideo edge from the production log: TLS succeeds,
	// but the server negotiates HTTP/1.1 rather than the h2 ALPN expected by
	// surf's Chrome impersonation transport.
	up := httptest.NewUnstartedServer(okVideoHandler("video"))
	up.EnableHTTP2 = false
	up.TLS = &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{}, // non-nil: make httptest omit ALPN like the affected CDN edge
	}
	up.StartTLS()
	t.Cleanup(up.Close)

	client := newBloggerProxyClient()
	req, err := http.NewRequest(http.MethodHead, up.URL, http.NoBody)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err, "Chrome TLS transport must support an HTTP/1.1-only edge")
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, resp.ProtoMajor)
}
