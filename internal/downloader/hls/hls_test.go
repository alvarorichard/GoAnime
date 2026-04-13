package hls

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Playlist fixtures — reproduce real-world CDN responses that triggered bugs
// ---------------------------------------------------------------------------

// masterPlaylistWithSeparateAudio reproduces the SuperFlix CDN response that
// caused ErrSeparateAudioTracks: a master playlist with #EXT-X-MEDIA TYPE=AUDIO
// pointing to a separate audio track URI.
const masterPlaylistWithSeparateAudio = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="Portuguese",DEFAULT=YES,AUTOSELECT=YES,URI="audio/pt.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720,AUDIO="audio"
720p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080,AUDIO="audio"
1080p/index.m3u8
`

// masterPlaylistMuxed is a normal master playlist without separate audio.
const masterPlaylistMuxed = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=854x480
480p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
1080p.m3u8
`

// mediaPlaylistObfuscatedJS reproduces the CDN that serves segments with .js
// and .html extensions. These are valid MPEG-TS data behind obfuscated names.
// BUG: ffmpeg rejects .js in allowed_segment_extensions → exit code 183.
// FIX: native HLS downloader doesn't care about extensions.
const mediaPlaylistObfuscatedJS = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
https://temquever.top/cdn/down/disk10/abc123/Video/720p/720p_000.html
#EXTINF:10.0,
https://calcinharosa.top/cdn/down/disk10/abc123/Video/720p/720p_001.js
#EXTINF:10.0,
https://calcinharosa.top/cdn/down/disk10/abc123/Video/720p/720p_002.js
#EXTINF:8.5,
https://calcinharosa.top/cdn/down/disk10/abc123/Video/720p/720p_003.js
#EXT-X-ENDLIST
`

// mediaPlaylistNormal uses conventional .ts extensions.
const mediaPlaylistNormal = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:VOD
#EXTINF:6.0, seg1
seg_000.ts
#EXTINF:6.0, seg2
seg_001.ts
#EXTINF:4.5, seg3
seg_002.ts
#EXT-X-ENDLIST
`

// mediaPlaylistLive has no #EXT-X-ENDLIST (live stream).
const mediaPlaylistLive = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:42
#EXTINF:6.0,
seg_042.ts
#EXTINF:6.0,
seg_043.ts
`

// nonHLSResponse simulates a CDN returning HTML instead of M3U8.
const nonHLSResponse = `<!DOCTYPE html><html><head><title>403 Forbidden</title></head><body>Access Denied</body></html>`

// ---------------------------------------------------------------------------
// Helper: mock CDN server
// ---------------------------------------------------------------------------

// mockCDN creates an httptest.Server that serves different playlists and
// segment data based on the request path. Returns the server and its base URL.
func mockCDN(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/master_separate_audio.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, masterPlaylistWithSeparateAudio)
		case r.URL.Path == "/master_muxed.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, masterPlaylistMuxed)
		case r.URL.Path == "/obfuscated.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, mediaPlaylistObfuscatedJS)
		case r.URL.Path == "/normal.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, mediaPlaylistNormal)
		case r.URL.Path == "/live.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, mediaPlaylistLive)
		case r.URL.Path == "/not_hls.m3u8":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, nonHLSResponse)
		case r.URL.Path == "/1080p.m3u8":
			// Media playlist returned from master playlist resolution
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, mediaPlaylistNormal)
		case strings.HasSuffix(r.URL.Path, ".ts") ||
			strings.HasSuffix(r.URL.Path, ".js") ||
			strings.HasSuffix(r.URL.Path, ".html"):
			// Simulate MPEG-TS segment data (188-byte TS packets, simplified)
			w.Header().Set("Content-Type", "video/mp2t")
			w.Write(fakeSegmentData(1024))
		case r.URL.Path == "/error_500":
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal Server Error")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// fakeSegmentData creates fake MPEG-TS-like data of the given size.
func fakeSegmentData(size int) []byte {
	data := make([]byte, size)
	// Write sync bytes at packet boundaries (every 188 bytes)
	for i := 0; i < size; i += 188 {
		data[i] = 0x47 // MPEG-TS sync byte
	}
	return data
}

// ---------------------------------------------------------------------------
// Tests: ErrSeparateAudioTracks detection
// ---------------------------------------------------------------------------

// TestParsePlaylist_SeparateAudioTracks verifies that a master playlist with
// separate audio tracks (#EXT-X-MEDIA TYPE=AUDIO with URI) returns
// ErrSeparateAudioTracks. This is the exact scenario from the SuperFlix CDN
// that caused the original bug: native HLS would download video-only,
// producing a silent movie.
func TestParsePlaylist_SeparateAudioTracks(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	_, err := dl.parsePlaylist(ctx, srv.URL+"/master_separate_audio.m3u8", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSeparateAudioTracks),
		"expected ErrSeparateAudioTracks, got: %v", err)
}

// TestParsePlaylist_MuxedAudio verifies that a master playlist with muxed
// audio (no separate #EXT-X-MEDIA TYPE=AUDIO URI) succeeds and selects
// the highest bandwidth stream.
func TestParsePlaylist_MuxedAudio(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	playlist, err := dl.parsePlaylist(ctx, srv.URL+"/master_muxed.m3u8", nil)
	require.NoError(t, err)
	assert.NotNil(t, playlist)
	// Should have resolved to the 1080p media playlist (highest bandwidth=5000000)
	assert.Greater(t, len(playlist.Segments), 0, "expected segments from resolved media playlist")
}

// ---------------------------------------------------------------------------
// Tests: Obfuscated segment extensions (.js, .html)
// ---------------------------------------------------------------------------

// TestParsePlaylist_ObfuscatedExtensions verifies that the HLS parser
// correctly handles segment URLs with .js and .html extensions.
// BUG: ffmpeg rejected these with "not in allowed_segment_extensions".
// FIX: native HLS downloads segments via raw HTTP — no extension check.
func TestParsePlaylist_ObfuscatedExtensions(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	playlist, err := dl.parsePlaylist(ctx, srv.URL+"/obfuscated.m3u8", nil)
	require.NoError(t, err)
	require.Len(t, playlist.Segments, 4)

	// Verify the parser preserved obfuscated URLs without mangling them
	assert.Contains(t, playlist.Segments[0].URL, "720p_000.html",
		"segment 0 must keep .html extension")
	assert.Contains(t, playlist.Segments[1].URL, "720p_001.js",
		"segment 1 must keep .js extension")
	assert.Contains(t, playlist.Segments[3].URL, "720p_003.js",
		"segment 3 must keep .js extension")

	// Verify durations are parsed
	assert.InDelta(t, 10.0, playlist.Segments[0].Duration, 0.01)
	assert.InDelta(t, 8.5, playlist.Segments[3].Duration, 0.01)
	assert.True(t, playlist.EndList, "obfuscated playlist should have ENDLIST")
}

// TestDownloadWithProgress_ObfuscatedExtensions performs an end-to-end
// download of an HLS stream where segments have .js/.html extensions.
// This is the core regression test for the original bug.
func TestDownloadWithProgress_ObfuscatedExtensions(t *testing.T) {
	t.Parallel()

	// Create a mock CDN that serves the obfuscated playlist + segments
	var segmentRequests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/stream.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			// Use relative URLs to test URL resolution
			fmt.Fprint(w, `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
seg_000.html
#EXTINF:10.0,
seg_001.js
#EXTINF:8.5,
seg_002.js
#EXT-X-ENDLIST
`)
		case strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".html"):
			segmentRequests.Add(1)
			w.Header().Set("Content-Type", "video/mp2t")
			w.Write(fakeSegmentData(2048))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Download to a temp file
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.ts")

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	// Track progress callbacks
	var lastBytesWritten int64
	var lastSegmentsWritten int
	var callbackCount atomic.Int32

	err := dl.DownloadWithProgress(ctx, srv.URL+"/stream.m3u8", outFile, nil,
		func(bytesWritten int64, segmentsWritten, totalSegments int) {
			callbackCount.Add(1)
			lastBytesWritten = bytesWritten
			lastSegmentsWritten = segmentsWritten
			assert.Equal(t, 3, totalSegments, "total segments should be 3")
		})

	require.NoError(t, err, "download must succeed despite .js/.html segment extensions")

	// Verify all segments were fetched
	assert.Equal(t, int32(3), segmentRequests.Load(),
		"all 3 segments (with .js/.html extensions) must be downloaded")

	// Verify output file exists and has data
	info, err := os.Stat(outFile)
	require.NoError(t, err)
	assert.Equal(t, int64(3*2048), info.Size(),
		"output file should contain all 3 segments (3 × 2048 bytes)")

	// Verify progress reporting
	assert.Greater(t, callbackCount.Load(), int32(0), "progress callback must be invoked")
	assert.Equal(t, int64(3*2048), lastBytesWritten,
		"final bytesWritten should match file size")
	assert.Equal(t, 3, lastSegmentsWritten,
		"final segmentsWritten should be 3")
}

// ---------------------------------------------------------------------------
// Tests: Normal playlist parsing
// ---------------------------------------------------------------------------

func TestParsePlaylist_NormalSegments(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	playlist, err := dl.parsePlaylist(ctx, srv.URL+"/normal.m3u8", nil)
	require.NoError(t, err)
	require.Len(t, playlist.Segments, 3)
	assert.Equal(t, "3", playlist.Version)
	assert.InDelta(t, 6.0, playlist.TargetDuration, 0.01)
	assert.Equal(t, "VOD", playlist.PlaylistType)
	assert.True(t, playlist.EndList)

	// Verify relative URL resolution
	assert.True(t, strings.HasSuffix(playlist.Segments[0].URL, "/seg_000.ts"))
	assert.True(t, strings.HasSuffix(playlist.Segments[2].URL, "/seg_002.ts"))

	// Verify segment titles
	assert.Equal(t, "seg1", playlist.Segments[0].Title)
	assert.Equal(t, "seg3", playlist.Segments[2].Title)
}

func TestParsePlaylist_LiveStream(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	playlist, err := dl.parsePlaylist(ctx, srv.URL+"/live.m3u8", nil)
	require.NoError(t, err)
	require.Len(t, playlist.Segments, 2)
	assert.False(t, playlist.EndList, "live stream must not have ENDLIST")
	assert.Equal(t, 42, playlist.MediaSequence)
}

// ---------------------------------------------------------------------------
// Tests: Non-HLS response detection
// ---------------------------------------------------------------------------

// TestParsePlaylist_NonHLSResponse verifies the downloader rejects HTML/binary
// responses that are not HLS playlists, instead of causing scanner errors.
func TestParsePlaylist_NonHLSResponse(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	_, err := dl.parsePlaylist(ctx, srv.URL+"/not_hls.m3u8", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an HLS playlist")
}

// ---------------------------------------------------------------------------
// Tests: HTTP error handling
// ---------------------------------------------------------------------------

func TestParsePlaylist_ServerError(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	_, err := dl.parsePlaylist(ctx, srv.URL+"/error_500", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestParsePlaylist_NotFound(t *testing.T) {
	t.Parallel()
	srv := mockCDN(t)
	defer srv.Close()

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	_, err := dl.parsePlaylist(ctx, srv.URL+"/nonexistent.m3u8", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// ---------------------------------------------------------------------------
// Tests: Custom headers forwarding
// ---------------------------------------------------------------------------

// TestDownload_HeadersForwarded verifies that custom headers (Referer, Origin)
// are forwarded to segment requests. CDNs reject requests without proper Referer.
func TestDownload_HeadersForwarded(t *testing.T) {
	t.Parallel()

	var capturedReferer atomic.Value
	var capturedOrigin atomic.Value

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture headers from segment requests
		if strings.HasSuffix(r.URL.Path, ".ts") {
			capturedReferer.Store(r.Header.Get("Referer"))
			capturedOrigin.Store(r.Header.Get("Origin"))
			w.Write(fakeSegmentData(512))
			return
		}
		// Serve a minimal playlist
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
seg.ts
#EXT-X-ENDLIST
`)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "headers_test.ts")

	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	headers := map[string]string{
		"Referer": "https://llanfairpwllgwyngy.com/",
		"Origin":  "https://llanfairpwllgwyngy.com",
	}

	err := dl.Download(ctx, srv.URL+"/playlist.m3u8", outFile, headers)
	require.NoError(t, err)

	ref, _ := capturedReferer.Load().(string)
	origin, _ := capturedOrigin.Load().(string)
	assert.Equal(t, "https://llanfairpwllgwyngy.com/", ref,
		"Referer header must be forwarded to segment requests")
	assert.Equal(t, "https://llanfairpwllgwyngy.com", origin,
		"Origin header must be forwarded to segment requests")
}

// ---------------------------------------------------------------------------
// Tests: NewDownloaderWithClient
// ---------------------------------------------------------------------------

// TestNewDownloaderWithClient verifies that the injected client is used for
// segment requests (not the default Go client with rejected TLS fingerprint).
func TestNewDownloaderWithClient(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
data.ts
#EXT-X-ENDLIST
`)
			return
		}
		w.Write(fakeSegmentData(256))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "client_test.ts")

	// Use the test server's client (follows its TLS)
	dl := NewDownloaderWithClient(srv.Client())
	ctx := context.Background()

	err := dl.Download(ctx, srv.URL+"/stream.m3u8", outFile, nil)
	require.NoError(t, err)

	// Both playlist + segment requests should go through the injected client
	assert.Equal(t, int32(2), requestCount.Load(),
		"injected client must be used for both playlist and segment requests")
}

// ---------------------------------------------------------------------------
// Tests: DownloadToFileWithClient convenience function
// ---------------------------------------------------------------------------

func TestDownloadToFileWithClient(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
chunk.ts
#EXT-X-ENDLIST
`)
			return
		}
		w.Write(fakeSegmentData(1024))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "convenience_test.ts")

	var progressCalled atomic.Bool
	err := DownloadToFileWithClient(
		context.Background(),
		srv.Client(),
		srv.URL+"/video.m3u8",
		outFile,
		nil,
		func(bytesWritten int64, segmentsWritten, totalSegments int) {
			progressCalled.Store(true)
		},
	)
	require.NoError(t, err)
	assert.True(t, progressCalled.Load(), "progress callback must be invoked")

	info, err := os.Stat(outFile)
	require.NoError(t, err)
	assert.Equal(t, int64(1024), info.Size())
}

// ---------------------------------------------------------------------------
// Tests: Segment retry on transient errors
// ---------------------------------------------------------------------------

// TestDownloadSegment_RetryOnTransientError verifies that segment downloads
// retry on transient HTTP errors (5xx) and eventually succeed.
func TestDownloadSegment_RetryOnTransientError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:2
#EXTINF:2.0,
retry_seg.ts
#EXT-X-ENDLIST
`)
			return
		}
		n := attempts.Add(1)
		if n <= 2 {
			// First 2 attempts: return 503 to trigger retry
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// 3rd attempt: succeed
		w.Write(fakeSegmentData(512))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "retry_test.ts")

	dl := NewDownloaderWithClient(srv.Client())
	err := dl.Download(context.Background(), srv.URL+"/retry.m3u8", outFile, nil)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, attempts.Load(), int32(3),
		"segment download should have been retried after transient errors")

	info, err := os.Stat(outFile)
	require.NoError(t, err)
	assert.Equal(t, int64(512), info.Size())
}

// ---------------------------------------------------------------------------
// Tests: Context cancellation
// ---------------------------------------------------------------------------

func TestDownload_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
seg1.ts
#EXTINF:5.0,
seg2.ts
#EXT-X-ENDLIST
`)
			return
		}
		// Delay segment response — context should be cancelled before completion
		<-r.Context().Done()
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "cancel_test.ts")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	dl := NewDownloaderWithClient(srv.Client())
	err := dl.Download(ctx, srv.URL+"/cancel.m3u8", outFile, nil)
	require.Error(t, err, "download must fail when context is cancelled")
}

// ---------------------------------------------------------------------------
// Tests: Partial segment failure tolerance
// ---------------------------------------------------------------------------

// TestDownload_PartialSegmentFailure verifies that a small number of failed
// segments (<5%) does not cause the entire download to fail.
func TestDownload_PartialSegmentFailure(t *testing.T) {
	t.Parallel()

	// Build a playlist with 25 segments, 1 will always fail (4%)
	var playlistBuilder strings.Builder
	playlistBuilder.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:5\n")
	for i := range 25 {
		fmt.Fprintf(&playlistBuilder, "#EXTINF:5.0,\nseg_%03d.ts\n", i)
	}
	playlistBuilder.WriteString("#EXT-X-ENDLIST\n")
	playlist := playlistBuilder.String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, playlist)
			return
		}
		// seg_012 always fails — simulates 1/25 = 4% failure rate
		if strings.Contains(r.URL.Path, "seg_012") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(fakeSegmentData(256))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "partial_fail.ts")

	dl := NewDownloaderWithClient(srv.Client())
	err := dl.Download(context.Background(), srv.URL+"/partial.m3u8", outFile, nil)
	require.NoError(t, err, "4% failure rate should be tolerated (threshold is 5%)")
}

// TestDownload_TooManySegmentsFail verifies that >5% segment failures
// cause the download to fail.
func TestDownload_TooManySegmentsFail(t *testing.T) {
	t.Parallel()

	// Build a playlist with 10 segments, 2 will always fail (20%)
	var playlistBuilder strings.Builder
	playlistBuilder.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:5\n")
	for i := range 10 {
		fmt.Fprintf(&playlistBuilder, "#EXTINF:5.0,\nseg_%03d.ts\n", i)
	}
	playlistBuilder.WriteString("#EXT-X-ENDLIST\n")
	playlist := playlistBuilder.String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprint(w, playlist)
			return
		}
		// seg_003 and seg_007 always fail — 20% failure rate
		if strings.Contains(r.URL.Path, "seg_003") || strings.Contains(r.URL.Path, "seg_007") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(fakeSegmentData(256))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "too_many_fail.ts")

	dl := NewDownloaderWithClient(srv.Client())
	err := dl.Download(context.Background(), srv.URL+"/fail.m3u8", outFile, nil)
	require.Error(t, err, "20% failure rate should cause download to fail")
	assert.Contains(t, err.Error(), "download incomplete")
}

// ---------------------------------------------------------------------------
// Tests: selectBestStream
// ---------------------------------------------------------------------------

func TestSelectBestStream(t *testing.T) {
	t.Parallel()

	dl := &Downloader{}
	baseURL := "https://cdn.example.com/video/master.m3u8"

	lines := []string{
		"#EXTM3U",
		"#EXT-X-STREAM-INF:BANDWIDTH=500000,RESOLUTION=640x360",
		"360p.m3u8",
		"#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=854x480",
		"480p.m3u8",
		"#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080",
		"1080p.m3u8",
		"#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720",
		"720p.m3u8",
	}

	best := dl.selectBestStream(lines, baseURL)
	assert.Equal(t, "https://cdn.example.com/video/1080p.m3u8", best,
		"must select the stream with highest bandwidth (1080p)")
}

func TestSelectBestStream_AbsoluteURLs(t *testing.T) {
	t.Parallel()

	dl := &Downloader{}
	lines := []string{
		"#EXTM3U",
		"#EXT-X-STREAM-INF:BANDWIDTH=3000000",
		"https://cdn1.example.com/720p.m3u8",
		"#EXT-X-STREAM-INF:BANDWIDTH=6000000",
		"https://cdn2.example.com/1080p.m3u8",
	}

	best := dl.selectBestStream(lines, "https://origin.example.com/master.m3u8")
	assert.Equal(t, "https://cdn2.example.com/1080p.m3u8", best,
		"should preserve absolute CDN URLs")
}

func TestSelectBestStream_Empty(t *testing.T) {
	t.Parallel()

	dl := &Downloader{}
	best := dl.selectBestStream([]string{"#EXTM3U"}, "https://example.com/master.m3u8")
	assert.Empty(t, best, "empty playlist should return empty string")
}

// ---------------------------------------------------------------------------
// Tests: Concurrent segment download ordering
// ---------------------------------------------------------------------------

// TestDownload_SegmentOrder verifies that segments are written to the output
// file in the correct order regardless of concurrent download timing.
func TestDownload_SegmentOrder(t *testing.T) {
	t.Parallel()

	// Each segment has a unique 4-byte marker for order verification
	markerA := []byte{0xAA, 0xAA, 0xAA, 0xAA}
	markerB := []byte{0xBB, 0xBB, 0xBB, 0xBB}
	markerC := []byte{0xCC, 0xCC, 0xCC, 0xCC}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/order.m3u8":
			fmt.Fprint(w, `#EXTM3U
#EXT-X-TARGETDURATION:5
#EXTINF:5.0,
seg_a.ts
#EXTINF:5.0,
seg_b.ts
#EXTINF:5.0,
seg_c.ts
#EXT-X-ENDLIST
`)
		case "/seg_a.ts":
			w.Write(markerA)
		case "/seg_b.ts":
			w.Write(markerB)
		case "/seg_c.ts":
			w.Write(markerC)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "order_test.ts")

	dl := NewDownloaderWithClient(srv.Client())
	err := dl.Download(context.Background(), srv.URL+"/order.m3u8", outFile, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(outFile) // #nosec G304 - test file
	require.NoError(t, err)
	require.Len(t, data, 12) // 3 segments × 4 bytes

	// Verify segments are written in order: A, B, C
	assert.Equal(t, markerA, data[0:4], "segment A must be first")
	assert.Equal(t, markerB, data[4:8], "segment B must be second")
	assert.Equal(t, markerC, data[8:12], "segment C must be third")
}

// ---------------------------------------------------------------------------
// Tests: Output path sanitization
// ---------------------------------------------------------------------------

func TestSanitizeOutputPath_DirectoryTraversal(t *testing.T) {
	t.Parallel()

	_, err := sanitizeOutputPath("../../etc/passwd")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "traversal")
}

func TestSanitizeOutputPath_Valid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := filepath.Join(dir, "output.ts")
	result, err := sanitizeOutputPath(input)
	require.NoError(t, err)

	expected, _ := filepath.Abs(input)
	assert.Equal(t, expected, result)
}
