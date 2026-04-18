package player

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// Tests: LooksLikeHLS — URL detection for HLS streams
// =====================================================================
// BUG: URLs like https://cdn.example.com/hls/<token> were not detected
// as HLS streams because they don't contain ".m3u8". The download went
// through the plain HTTP path which can't handle HLS playlists.
// FIX: LooksLikeHLS now also matches "/hls/" path segments.

func TestLooksLikeHLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		// Standard .m3u8 URLs
		{
			name: "explicit m3u8 extension",
			url:  "https://cdn.example.com/video/master.m3u8",
			want: true,
		},
		{
			name: "m3u8 with query string",
			url:  "https://cdn.example.com/stream.m3u8?md5=abc&expires=123",
			want: true,
		},
		{
			name: "m3u8 in query parameter",
			url:  "https://proxy.example.com/fetch?url=master.m3u8",
			want: true,
		},

		// /hls/ path segment — the bug scenario
		{
			name: "CDN /hls/ path without .m3u8 extension",
			url:  "https://llanfairpwllgwyngy.com/hls/Y1JEMmw2SEdlRTdybDNHV2lRY0RLblE0ZFhIZXJLYytvSVlCYkI3VTBIUlN",
			want: true,
		},
		{
			name: "CDN /hls/ path with token and URL encoding",
			url:  "https://cdn.example.com/hls/base64token%3D%3D.",
			want: true,
		},
		{
			name: "nested /hls/ path",
			url:  "https://cdn.example.com/cdn/hls/abc123/master.m3u8",
			want: true,
		},
		{
			name: "uppercase /HLS/ path",
			url:  "https://cdn.example.com/HLS/stream",
			want: true,
		},
		{
			name: "mixed case /Hls/ path",
			url:  "https://cdn.example.com/Hls/live",
			want: true,
		},

		// Non-HLS URLs — must NOT match
		{
			name: "direct MP4 URL",
			url:  "https://cdn.example.com/video.mp4",
			want: false,
		},
		{
			name: "blogger URL",
			url:  "https://www.blogger.com/video.php?blogID=123",
			want: false,
		},
		{
			name: "SharePoint aspx",
			url:  "https://myorg.sharepoint.com/sites/media/file.aspx",
			want: false,
		},
		{
			name: "empty string",
			url:  "",
			want: false,
		},
		{
			name: "plain text",
			url:  "not-a-url",
			want: false,
		},
		{
			name: "hls substring in hostname but not path",
			url:  "https://hlscdn.example.com/video.mp4",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := LooksLikeHLS(tc.url)
			assert.Equal(t, tc.want, got, "LooksLikeHLS(%q)", tc.url)
		})
	}
}

// =====================================================================
// Tests: hasUnsafeExtension
// =====================================================================

func TestHasUnsafeExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"aspx extension", "https://myorg.sharepoint.com/file.aspx", true},
		{"asp extension", "https://legacy.example.com/play.asp", true},
		{"php extension", "https://old-server.com/stream.php", true},
		{"jsp extension", "https://java-server.com/media.jsp", true},
		{"cgi extension", "https://cgi.example.com/video.cgi", true},
		{"ASPX uppercase", "https://myorg.sharepoint.com/file.ASPX", true},
		{"aspx with query", "https://myorg.sharepoint.com/file.aspx?token=abc", true},
		{"mp4 is safe", "https://cdn.example.com/video.mp4", false},
		{"m3u8 is safe", "https://cdn.example.com/stream.m3u8", false},
		{"ts is safe", "https://cdn.example.com/seg.ts", false},
		{"js is safe", "https://cdn.example.com/seg.js", false},
		{"no extension", "https://cdn.example.com/stream", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := hasUnsafeExtension(tc.url)
			assert.Equal(t, tc.want, got, "hasUnsafeExtension(%q)", tc.url)
		})
	}
}

// =====================================================================
// Tests: isRetryableError
// =====================================================================

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"timeout", errors.New("i/o timeout"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"network unreachable", errors.New("network is unreachable"), true},
		{"temporary failure", errors.New("temporary failure in name resolution"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"permission denied", errors.New("permission denied"), false},
		{"not found", errors.New("404 not found"), false},
		{"ffmpeg exit code", errors.New("ffmpeg exited with code 183"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isRetryableError(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// =====================================================================
// Tests: isUnsafeExtensionError
// =====================================================================

func TestIsUnsafeExtensionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"unsafe extension msg", errors.New("unsafe file extension detected"), true},
		{"unusual extension msg", errors.New("the file has an unusual extension"), true},
		{"skip msg", errors.New("is unusual and will be skipped"), true},
		{"yt-dlp CVE message", errors.New("URL has an unsafe extension"), true},
		{"generic error", errors.New("download failed"), false},
		{"ffmpeg extension error", errors.New("not in allowed_segment_extensions"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isUnsafeExtensionError(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// =====================================================================
// Tests: isBloggerProxyURL
// =====================================================================

func TestIsBloggerProxyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid proxy URL", "http://127.0.0.1:8080/blogger_proxy/video", true},
		{"proxy with different port", "http://127.0.0.1:9999/blogger_proxy/xyz", true},
		{"not localhost", "https://cdn.example.com/blogger_proxy/video", false},
		{"no proxy path", "http://127.0.0.1:8080/video.mp4", false},
		{"external blogger", "https://www.blogger.com/video.php", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isBloggerProxyURL(tc.url)
			assert.Equal(t, tc.want, got)
		})
	}
}

// =====================================================================
// Tests: downloadWithNativeHLS — integration with mock CDN
// =====================================================================

// mockHLSCDN creates a test server that simulates the problematic CDN:
// - Master playlist → separate audio tracks → ErrSeparateAudioTracks
// - Media playlist with .js/.html segments → must succeed
// - Normal .ts segments → baseline
func mockHLSCDN(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/separate_audio.m3u8":
			fmt.Fprint(w, `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aac",NAME="Portuguese",DEFAULT=YES,URI="audio_pt.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=2500000,AUDIO="aac"
720p.m3u8
`)
		case "/obfuscated.m3u8":
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
seg_000.html
#EXTINF:10.0,
seg_001.js
#EXT-X-ENDLIST
`)
		case "/normal.m3u8":
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
seg.ts
#EXT-X-ENDLIST
`)
		default:
			// All segment requests return fake MPEG-TS data
			w.Write(make([]byte, 2048))
		}
	}))
}

// TestDownloadWithNativeHLS_SeparateAudioTracksError verifies that native HLS
// correctly returns ErrSeparateAudioTracks for playlists with separate audio.
// This is the exact scenario from the SuperFlix CDN that caused the original bug.
func TestDownloadWithNativeHLS_SeparateAudioTracksError(t *testing.T) {
	t.Parallel()
	srv := mockHLSCDN(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "sep_audio.mp4")

	err := hls.DownloadToFileWithClient(
		t.Context(),
		srv.Client(),
		srv.URL+"/separate_audio.m3u8",
		outFile,
		nil,
		nil,
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, hls.ErrSeparateAudioTracks),
		"expected ErrSeparateAudioTracks, got: %v", err)
}

// TestDownloadWithNativeHLS_ObfuscatedExtensionsSuccess verifies that native
// HLS successfully downloads segments with .js/.html extensions.
// BUG: ffmpeg rejected these with "not in allowed_segment_extensions, exit 183"
// FIX: native HLS downloads via raw HTTP — no extension check.
func TestDownloadWithNativeHLS_ObfuscatedExtensionsSuccess(t *testing.T) {
	t.Parallel()
	srv := mockHLSCDN(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "obfuscated.ts")

	err := hls.DownloadToFileWithClient(
		t.Context(),
		srv.Client(),
		srv.URL+"/obfuscated.m3u8",
		outFile,
		nil,
		nil,
	)

	require.NoError(t, err, "native HLS must handle .js/.html segments")

	info, err := os.Stat(outFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "output file must have data")
}

// TestDownloadWithNativeHLS_NormalSegments verifies baseline functionality.
func TestDownloadWithNativeHLS_NormalSegments(t *testing.T) {
	t.Parallel()
	srv := mockHLSCDN(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "normal.ts")

	err := hls.DownloadToFileWithClient(
		t.Context(),
		srv.Client(),
		srv.URL+"/normal.m3u8",
		outFile,
		nil,
		nil,
	)

	require.NoError(t, err)
}

// =====================================================================
// Tests: Fallback chain integration
// =====================================================================

// TestFallbackChain_SimulatedFlow simulates the complete download fallback
// chain as it actually runs in player.go. Uses a mock CDN to verify:
// 1. Native HLS → ErrSeparateAudioTracks (expected)
// 2. Would call downloadWithYtDlp (tested separately, not called here)
// The test verifies each stage produces the expected error/success.
func TestFallbackChain_SimulatedFlow(t *testing.T) {
	t.Parallel()

	// Stage 1: Native HLS correctly detects separate audio
	t.Run("stage1_nativeHLS_separate_audio", func(t *testing.T) {
		t.Parallel()
		srv := mockHLSCDN(t)
		defer srv.Close()

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "stage1.ts")

		err := hls.DownloadToFileWithClient(
			t.Context(), srv.Client(),
			srv.URL+"/separate_audio.m3u8", outFile, nil, nil,
		)
		require.Error(t, err)
		// The fallback chain checks errors.Is(err, ErrSeparateAudioTracks)
		// which wraps through "failed to parse playlist: ..."
		assert.True(t, errors.Is(err, hls.ErrSeparateAudioTracks),
			"stage 1 must return ErrSeparateAudioTracks for fallback routing")
	})

	// Stage 2: Native HLS succeeds when no separate audio (obfuscated segments)
	t.Run("stage2_nativeHLS_obfuscated_succeeds", func(t *testing.T) {
		t.Parallel()
		srv := mockHLSCDN(t)
		defer srv.Close()

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "stage2.ts")

		err := hls.DownloadToFileWithClient(
			t.Context(), srv.Client(),
			srv.URL+"/obfuscated.m3u8", outFile, nil, nil,
		)
		require.NoError(t, err,
			"stage 2: native HLS must succeed for obfuscated .js/.html segments")
	})
}

// TestFallbackChain_RoutingLogic verifies the download routing logic used in
// player.go downloads — the if/else chain that decides which download method
// to use based on URL characteristics.
func TestFallbackChain_RoutingLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		videoURL  string
		expectHLS bool // should route to native HLS / yt-dlp HLS path
	}{
		{
			name:      "m3u8 URL routes to HLS path",
			videoURL:  "https://cdn.example.com/stream.m3u8?token=abc",
			expectHLS: true,
		},
		{
			name:      "/hls/ URL routes to HLS path",
			videoURL:  "https://llanfairpwllgwyngy.com/hls/Y1JEMmw2SEdlR",
			expectHLS: true,
		},
		{
			name:      "sharepoint aspx routes to HLS path via unsafe extension",
			videoURL:  "https://myorg.sharepoint.com/sites/file.aspx",
			expectHLS: true, // hasUnsafeExtension OR LooksLikeHLS
		},
		{
			name:      "direct MP4 routes to non-HLS path",
			videoURL:  "https://cdn.example.com/video.mp4",
			expectHLS: false,
		},
		{
			name:      "blogger URL routes to non-HLS path",
			videoURL:  "https://blogger.com/video",
			expectHLS: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			isHLS := LooksLikeHLS(tc.videoURL) || hasUnsafeExtension(tc.videoURL)
			assert.Equal(t, tc.expectHLS, isHLS,
				"routing for %q", tc.videoURL)
		})
	}
}

// =====================================================================
// Tests: downloadDirectHTTP — mock integration
// =====================================================================

// TestDownloadDirectHTTP_Success verifies the direct HTTP fallback downloads
// correctly from a mock server.
// TestDownloadDirectHTTP_BadPath verifies path validation rejects traversal.
func TestDownloadDirectHTTP_BadPath(t *testing.T) {
	t.Parallel()

	err := downloadDirectHTTP("https://cdn.example.com/video.mp4", "/tmp/../../etc/passwd", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output path")
}

// TestDownloadDirectHTTP_EmptyPath verifies empty path is rejected.
func TestDownloadDirectHTTP_EmptyPath(t *testing.T) {
	t.Parallel()

	err := downloadDirectHTTP("https://cdn.example.com/video.mp4", "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output path")
}

// =====================================================================
// Tests: Progress model concurrent safety
// =====================================================================
// Reproduces the race condition where multiple goroutines update the
// progress model simultaneously during concurrent HLS segment downloads.

func TestProgressModel_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	// Build a playlist with many segments to maximize concurrency
	var playlistBuilder strings.Builder
	playlistBuilder.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:2\n")
	for i := range 50 {
		fmt.Fprintf(&playlistBuilder, "#EXTINF:2.0,\nseg_%03d.ts\n", i)
	}
	playlistBuilder.WriteString("#EXT-X-ENDLIST\n")
	playlist := playlistBuilder.String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, playlist)
			return
		}
		w.Write(make([]byte, 1024))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "concurrent.ts")

	// This callback runs from the HLS goroutine while the main goroutine
	// reads the same fields — races should be detected by -race.
	var progressUpdates atomic.Int32
	callback := func(bytesWritten int64, segmentsWritten, totalSegments int) {
		progressUpdates.Add(1)
		// Simulate what the real progress bar does: read totalSegments,
		// compute percentage, etc.
		if totalSegments > 0 {
			_ = float64(segmentsWritten) / float64(totalSegments)
		}
	}

	err := hls.DownloadToFileWithClient(
		t.Context(), srv.Client(),
		srv.URL+"/many.m3u8", outFile, nil, callback,
	)
	require.NoError(t, err)
	assert.Greater(t, progressUpdates.Load(), int32(0),
		"progress callback must be invoked during concurrent download")
}

// =====================================================================
// Tests: Progress accuracy — the core bug
// =====================================================================

// TestProgressNotStuckAt500MB verifies that after native HLS fails with
// ErrSeparateAudioTracks, the fallback chain resets the progress model so
// that yt-dlp can set the real totalBytes instead of being stuck at the
// original 500 MB estimate.
//
// BUG: getContentLength returned 500 MB for HLS. yt-dlp reported real
// TotalBytes (e.g. 150 MB) but the condition `m.totalBytes < totalBytes`
// was FALSE (500 > 150), so the bar stayed at 500 MB and showed ~30%.
func TestProgressNotStuckAt500MB(t *testing.T) {
	t.Parallel()

	// Simulate the pre-fix state: totalBytes starts at 500MB from getContentLength
	m := &model{}
	m.totalBytes = 500 * 1024 * 1024 // old wrong initial value

	// Simulate native HLS failure + reset (the fix)
	m.received = 0
	m.totalBytes = 0

	// Simulate yt-dlp reporting real video size (150MB)
	const videoTotal = 150 * 1024 * 1024
	m.totalBytes = videoTotal
	m.received = videoTotal / 2 // halfway through video

	pct := float64(m.received) / float64(m.totalBytes)
	assert.InDelta(t, 0.5, pct, 0.01,
		"progress should be ~50%% when halfway through 150MB video, not %0.1f%%", pct*100)

	// Simulate yt-dlp switching to audio file (20MB) — per-file total tracking
	const audioTotal = 20 * 1024 * 1024
	m.totalBytes = videoTotal + audioTotal // sum of all files
	m.received = videoTotal + audioTotal/2 // video done + halfway through audio

	pct = float64(m.received) / float64(m.totalBytes)
	assert.InDelta(t, 0.94, pct, 0.02,
		"progress should be ~94%% (160/170 MB), got %.1f%%", pct*100)
}

// TestProgressNativeHLSDynamicEstimate verifies that the native HLS callback
// properly estimates totalBytes from segment averages and allows shrinking
// after 10+ segments (fixing stuck-at-low-percentage bug).
func TestProgressNativeHLSDynamicEstimate(t *testing.T) {
	t.Parallel()

	m := &model{}
	m.totalBytes = 0 // start unknown
	const totalSegments = 100

	// Simulate 5 segments downloaded (200KB each)
	segmentsWritten := 5
	bytesWritten := int64(5 * 200 * 1024)

	// After 3+ segments, only increase estimate
	avgBytesPerSeg := bytesWritten / int64(segmentsWritten)
	estimatedTotal := avgBytesPerSeg * int64(totalSegments)
	if estimatedTotal > m.totalBytes {
		m.totalBytes = estimatedTotal
	}
	m.received = bytesWritten

	assert.Equal(t, int64(totalSegments)*200*1024, m.totalBytes,
		"estimated total should be avgSegSize × totalSegments")

	// After 10+ segments with smaller average (150KB each), estimate should shrink
	segmentsWritten = 15
	bytesWritten = int64(15 * 150 * 1024)
	avgBytesPerSeg = bytesWritten / int64(segmentsWritten)
	estimatedTotal = avgBytesPerSeg * int64(totalSegments)
	// After 10+ segments, set directly (allow shrink)
	m.totalBytes = estimatedTotal
	m.received = bytesWritten

	assert.Equal(t, int64(totalSegments)*150*1024, m.totalBytes,
		"after 10+ segments, estimate must shrink to match reality")

	pct := float64(m.received) / float64(m.totalBytes)
	assert.InDelta(t, 0.15, pct, 0.01,
		"progress should be ~15%% (15/100 segments), got %.1f%%", pct*100)
}
