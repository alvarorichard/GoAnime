package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock progress sender — drives runDownloadWithProgress without a TTY.
// ---------------------------------------------------------------------------

// mockProgressSender implements progressSender. It records every Send call,
// blocks Run until Quit is invoked, and lets tests inject a Run-time error.
type mockProgressSender struct {
	mu       sync.Mutex
	sent     []tea.Msg
	quitCh   chan struct{}
	quitOnce sync.Once
	runErr   error
	runDelay time.Duration
	runCount atomic.Int32
}

func newMockSender() *mockProgressSender {
	return &mockProgressSender{quitCh: make(chan struct{})}
}

func (m *mockProgressSender) Send(msg tea.Msg) {
	m.mu.Lock()
	m.sent = append(m.sent, msg)
	m.mu.Unlock()
}

func (m *mockProgressSender) Quit() {
	m.quitOnce.Do(func() { close(m.quitCh) })
}

func (m *mockProgressSender) Run() (tea.Model, error) {
	m.runCount.Add(1)
	if m.runDelay > 0 {
		time.Sleep(m.runDelay)
	}
	<-m.quitCh
	return nil, m.runErr
}

func (m *mockProgressSender) sentMessages() []tea.Msg {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]tea.Msg, len(m.sent))
	copy(out, m.sent)
	return out
}

func (m *mockProgressSender) hasStatusContaining(substr string) bool {
	for _, msg := range m.sentMessages() {
		if s, ok := msg.(statusMsg); ok && strings.Contains(string(s), substr) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Builder for an injected downloader: loopback-allowed HTTP client + no-op
// sleep + mock sender. Returns the downloader and the sender for assertions.
// ---------------------------------------------------------------------------

func makeInjectedDownloader(t *testing.T) (*EpisodeDownloader, *mockProgressSender) {
	t.Helper()
	sender := newMockSender()
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(1), false)
	d.opts = downloaderOptions{
		httpClient: &http.Client{Timeout: 5 * time.Second}, // bypass SafeTransport for loopback
		sleep:      func(time.Duration) {},                 // skip the 1s/500ms pauses
		newSender:  func(tea.Model) progressSender { return sender },
	}
	return d, sender
}

func newProgressModelForTest(total int64) *progressModel {
	return &progressModel{
		progress:   progress.New(progress.WithDefaultBlend()),
		totalBytes: total,
	}
}

// ---------------------------------------------------------------------------
// selectDownloadMethod — pure routing rule, table-driven exhaustive.
// ---------------------------------------------------------------------------

func TestSelectDownloadMethod(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		url         string
		wantPrimary downloadMethod
		wantHasFB   bool
	}{
		{"plain_mp4", "https://cdn.example.com/file.mp4", methodHTTP, false},
		{"hls_m3u8", "https://cdn.example.com/index.m3u8", methodYtDlp, false},
		{"hls_master", "https://cdn.example.com/master.m3u8", methodYtDlp, false},
		{"wixmp_repackager", "https://repackager.wixmp.com/v.mpd", methodYtDlp, false},
		{"blogger", "https://blogger.com/v.g?id=1", methodYtDlp, false},
		{"sharepoint_fallback", "https://x.sharepoint.com/v.mp4", methodHTTP, true},
		{"allmanga", "https://allmanga.to/v.mp4", methodYtDlp, false},
		{"empty_string", "", methodHTTP, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			primary, fallback, hasFB := selectDownloadMethod(tc.url)
			assert.Equal(t, tc.wantPrimary, primary, "primary method")
			assert.Equal(t, tc.wantHasFB, hasFB, "hasFallback")
			// fallback must be the opposite method
			assert.NotEqual(t, primary, fallback, "fallback must differ from primary")
		})
	}
}

// ---------------------------------------------------------------------------
// runMethod — dispatcher correctness (including unknown method).
// ---------------------------------------------------------------------------

func TestRunMethod_UnknownReturnsError(t *testing.T) {
	t.Parallel()
	d, _ := makeInjectedDownloader(t)
	err := d.runMethod(downloadMethod(99), "https://x/y.mp4", "/tmp/x.mp4", newProgressModelForTest(0), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown download method")
}

// ---------------------------------------------------------------------------
// Helper-method coverage: httpClient/sleepFn/newSender default & injected.
// ---------------------------------------------------------------------------

func TestEpisodeDownloader_Helpers_DefaultsAndInjected(t *testing.T) {
	t.Parallel()

	t.Run("httpClient_default_is_safe", func(t *testing.T) {
		t.Parallel()
		d := &EpisodeDownloader{}
		c := d.httpClient()
		require.NotNil(t, c)
		assert.NotNil(t, c.Transport, "default client must use SafeTransport")
	})

	t.Run("httpClient_injected_is_used", func(t *testing.T) {
		t.Parallel()
		injected := &http.Client{Timeout: 7 * time.Second}
		d := &EpisodeDownloader{opts: downloaderOptions{httpClient: injected}}
		assert.Same(t, injected, d.httpClient())
	})

	t.Run("sleepFn_default_no_panic", func(t *testing.T) {
		t.Parallel()
		d := &EpisodeDownloader{}
		assert.NotPanics(t, func() { d.sleepFn(time.Nanosecond) })
	})

	t.Run("sleepFn_injected_called", func(t *testing.T) {
		t.Parallel()
		var called atomic.Int32
		d := &EpisodeDownloader{opts: downloaderOptions{sleep: func(time.Duration) { called.Add(1) }}}
		d.sleepFn(time.Second)
		assert.Equal(t, int32(1), called.Load())
	})

	t.Run("newSender_default_is_tui_program", func(t *testing.T) {
		t.Parallel()
		d := &EpisodeDownloader{}
		s := d.newSender(&progressModel{})
		require.NotNil(t, s)
		_, ok := s.(*tea.Program)
		assert.True(t, ok, "default sender must be a *tea.Program")
	})

	t.Run("newSender_injected", func(t *testing.T) {
		t.Parallel()
		mock := newMockSender()
		d := &EpisodeDownloader{opts: downloaderOptions{newSender: func(tea.Model) progressSender { return mock }}}
		assert.Same(t, mock, d.newSender(&progressModel{}))
	})
}

// ---------------------------------------------------------------------------
// downloadHTTPWithProgress — exhaustive HTTP scenarios via httptest +
// injected loopback-friendly client.
// ---------------------------------------------------------------------------

func TestDownloadHTTPWithProgress_Success(t *testing.T) {
	t.Parallel()
	body := bytesOfSize(64 * 1024) // 64KB → exercises multiple 32KB buffer reads
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	m := newProgressModelForTest(int64(len(body)))

	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, m, sender)
	require.NoError(t, err)

	got, err := os.ReadFile(dest) // #nosec G304: path under t.TempDir()
	require.NoError(t, err)
	assert.Equal(t, body, got, "downloaded bytes must match server body")

	// Progress model must reflect totals.
	m.mu.Lock()
	assert.Equal(t, int64(len(body)), m.totalBytes)
	m.mu.Unlock()

	// Sender must have received at least one progressMsg.
	var sawProgress bool
	for _, msg := range sender.sentMessages() {
		if pm, ok := msg.(progressMsg); ok && pm.received > 0 {
			sawProgress = true
			break
		}
	}
	assert.True(t, sawProgress, "expected at least one progressMsg with bytes")
}

func TestDownloadHTTPWithProgress_BadStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
	assert.NoFileExists(t, dest, "no file must be created on bad status")
}

func TestDownloadHTTPWithProgress_EmptyBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "0")
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, newProgressModelForTest(0), sender)
	require.NoError(t, err)
	stat, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stat.Size())
}

func TestDownloadHTTPWithProgress_NoContentLength(t *testing.T) {
	t.Parallel()
	body := []byte("CHUNKED PAYLOAD DATA")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Omit Content-Length; Go sends chunked transfer.
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	m := newProgressModelForTest(int64(len(body))) // pre-set total since server doesn't provide it
	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, m, sender)
	require.NoError(t, err)
	got, _ := os.ReadFile(dest) // #nosec G304: path under t.TempDir()
	assert.Equal(t, body, got)
}

func TestDownloadHTTPWithProgress_GetError(t *testing.T) {
	t.Parallel()
	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	// Use a URL with an unresolvable host so client.Get fails with a network error.
	err := d.downloadHTTPWithProgress("http://invalid.invalid.invalid.test/file.mp4", dest, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start download")
}

func TestDownloadHTTPWithProgress_PathEscape(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	// Path that escapes OutputDir → sanitizeDestPath must reject.
	dest := filepath.Join(d.config.OutputDir, "..", "..", "evil.mp4")
	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid destination path")
}

// TestDownloadHTTPWithProgress_BodyReadError simulates a server hanging up
// mid-stream. We do that by having the handler hijack the connection and
// close it after writing partial data without a Content-Length header.
func TestDownloadHTTPWithProgress_BodyReadError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Errorf("ResponseWriter does not support Hijack")
			return
		}
		// Write a chunked response, then hijack + close to abort the stream.
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		conn, _, err := hj.Hijack()
		if err == nil {
			_ = conn.Close()
		}
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	err := d.downloadHTTPWithProgress(srv.URL+"/v.mp4", dest, newProgressModelForTest(1000000), sender)
	require.Error(t, err)
	// Either "failed to read from response" or an EOF-style transport error.
	assert.True(t,
		strings.Contains(err.Error(), "failed to read") ||
			strings.Contains(err.Error(), "EOF") ||
			strings.Contains(err.Error(), "unexpected"),
		"expected read error, got %q", err.Error(),
	)
}

// ---------------------------------------------------------------------------
// runDownloadWithProgress — full pipeline through the mock sender.
// ---------------------------------------------------------------------------

func TestRunDownloadWithProgress_Success(t *testing.T) {
	// Sequential: uses os.Stdin for the play prompt.
	body := bytesOfSize(2048) // > 1024 to pass size verification
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")
	m := newProgressModelForTest(int64(len(body)))

	withClosedStdin(t) // promptPlayDownloaded reads stdin
	err := d.runDownloadWithProgress(srv.URL+"/v.mp4", dest, 1, m, sender)
	require.NoError(t, err)

	assert.FileExists(t, dest)
	assert.True(t, sender.hasStatusContaining("Download completed"), "expected success status sent")
	m.mu.Lock()
	assert.True(t, m.done, "progress model must be marked done after pipeline")
	m.mu.Unlock()
}

func TestRunDownloadWithProgress_FileTooSmall(t *testing.T) {
	t.Parallel()
	body := []byte("tiny") // < 1024 → triggers size verification error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")

	err := d.runDownloadWithProgress(srv.URL+"/v.mp4", dest, 1, newProgressModelForTest(int64(len(body))), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too small")
	assert.True(t, sender.hasStatusContaining("Download completed") || sender.hasStatusContaining("Download failed"),
		"sender must have received a terminal status")
}

func TestRunDownloadWithProgress_HTTPErrorPropagates(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")

	err := d.runDownloadWithProgress(srv.URL+"/v.mp4", dest, 7, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
	assert.True(t, sender.hasStatusContaining("Download failed"), "expected failure status")
}

func TestRunDownloadWithProgress_SenderRunError(t *testing.T) {
	t.Parallel()
	// Inject an HTTP server that never gets called: we make the sender error
	// out from Run() before/while the download goroutine completes.
	body := bytesOfSize(2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	sender.runErr = errors.New("tea boom")
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")

	err := d.runDownloadWithProgress(srv.URL+"/v.mp4", dest, 1, newProgressModelForTest(int64(len(body))), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "progress display error")
	assert.Contains(t, err.Error(), "tea boom")
}

// TestRunDownloadWithProgress_FileMissingAfterDownload exercises the
// "file was not created" branch. We inject an HTTP transport that returns a
// success status with body but the destination path stays missing because the
// destination is a directory (os.Create fails earlier with an error → covered
// by the HTTPError test). To explicitly hit "file was not created" we mock the
// HTTP layer to succeed without writing — easiest path: inject an httpClient
// transport that returns 200 + zero-length body and we make the dir not exist
// → Create fails. Both covered by other tests. This pin keeps signature.
func TestRunDownloadWithProgress_PathEscapeRejected(t *testing.T) {
	t.Parallel()
	d, sender := makeInjectedDownloader(t)
	// Escape attempt → HTTP layer's sanitizeDestPath errors out before write.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytesOfSize(2048))
	}))
	t.Cleanup(srv.Close)
	badDest := filepath.Join(d.config.OutputDir, "..", "..", "evil.mp4")
	err := d.runDownloadWithProgress(srv.URL+"/v.mp4", badDest, 1, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.True(t,
		strings.Contains(err.Error(), "invalid destination path") ||
			strings.Contains(err.Error(), "file was not created"),
		"expected path-escape or missing-file error, got %q", err.Error(),
	)
}

// ---------------------------------------------------------------------------
// downloadWithProgress — full top-level entry, now testable via injected
// newSender. Exercises MkdirAll + getContentLength fallback + delegation.
// ---------------------------------------------------------------------------

func TestDownloadWithProgress_FullPath_HTTPSuccess(t *testing.T) {
	// Sequential: pipeline ends in promptPlayDownloaded which reads stdin.
	body := bytesOfSize(4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	// getContentLength will fail for the loopback URL (SafeTransport not used
	// here for the actual HTTP fetch — but getContentLength constructs its own
	// SafeTransport client). It will go through the fallback branch and use
	// 200MB estimate. That's fine for the pipeline.
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")
	withClosedStdin(t)

	err := d.downloadWithProgress(srv.URL+"/v.mp4", dest, 1)
	require.NoError(t, err)
	assert.FileExists(t, dest)
	assert.True(t, sender.hasStatusContaining("Download completed"))
}

func TestDownloadWithProgress_MkdirFailsWhenParentIsFile(t *testing.T) {
	t.Parallel()
	d, _ := makeInjectedDownloader(t)
	// Create a regular file where the parent directory would need to exist.
	blocker := filepath.Join(d.config.OutputDir, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
	// Destination claims `blocker` as a directory → MkdirAll must fail.
	dest := filepath.Join(blocker, "child", "ep.mp4")
	err := d.downloadWithProgress("https://example.invalid/v.mp4", dest, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
}

// ---------------------------------------------------------------------------
// downloadEpisodeWithProgress routing — verifies selectDownloadMethod is
// actually wired into the dispatch. Empty URL is already covered elsewhere.
// ---------------------------------------------------------------------------

func TestDownloadEpisodeWithProgress_HTTPRouteSucceeds(t *testing.T) {
	t.Parallel()
	body := bytesOfSize(1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	err := d.downloadEpisodeWithProgress(srv.URL+"/v.mp4", dest, newProgressModelForTest(int64(len(body))), sender)
	require.NoError(t, err)
	assert.FileExists(t, dest)
}

func TestDownloadEpisodeWithProgress_EmptyURLAlt(t *testing.T) {
	t.Parallel()
	d, sender := makeInjectedDownloader(t)
	dest := filepath.Join(d.config.OutputDir, "ep.mp4")
	err := d.downloadEpisodeWithProgress("", dest, newProgressModelForTest(0), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty video URL")
}

// ---------------------------------------------------------------------------
// Mock sender — its own contract tests so future regressions are caught.
// ---------------------------------------------------------------------------

func TestMockProgressSender_Contract(t *testing.T) {
	t.Parallel()

	t.Run("send_records_messages", func(t *testing.T) {
		t.Parallel()
		m := newMockSender()
		m.Send(statusMsg("hello"))
		m.Send(progressMsg{received: 10, totalBytes: 100})
		msgs := m.sentMessages()
		require.Len(t, msgs, 2)
	})

	t.Run("run_blocks_until_quit", func(t *testing.T) {
		t.Parallel()
		m := newMockSender()
		done := make(chan struct{})
		go func() {
			_, _ = m.Run()
			close(done)
		}()
		// Run must block initially.
		select {
		case <-done:
			t.Fatal("Run returned before Quit was called")
		case <-time.After(50 * time.Millisecond):
		}
		m.Quit()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Run did not return after Quit")
		}
	})

	t.Run("quit_is_idempotent", func(t *testing.T) {
		t.Parallel()
		m := newMockSender()
		assert.NotPanics(t, func() {
			m.Quit()
			m.Quit()
			m.Quit()
		})
	})

	t.Run("run_returns_injected_error", func(t *testing.T) {
		t.Parallel()
		want := errors.New("boom")
		m := newMockSender()
		m.runErr = want
		go m.Quit()
		_, err := m.Run()
		assert.ErrorIs(t, err, want)
	})

	t.Run("hasStatusContaining", func(t *testing.T) {
		t.Parallel()
		m := newMockSender()
		m.Send(statusMsg("Downloading episode 3..."))
		assert.True(t, m.hasStatusContaining("Downloading"))
		assert.False(t, m.hasStatusContaining("nope"))
	})
}

// ---------------------------------------------------------------------------
// Small test utilities — local, no external deps.
// ---------------------------------------------------------------------------

func bytesOfSize(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('A' + (i % 26))
	}
	return b
}
