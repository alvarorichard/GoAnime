// Package hls provides HLS (HTTP Live Streaming) download functionality
package hls

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSeparateAudioTracks is returned when a master playlist contains separate
// audio tracks (#EXT-X-MEDIA:TYPE=AUDIO with a URI). The native HLS downloader
// only handles muxed streams; callers should fall back to yt-dlp which properly
// merges separate video and audio tracks.
var ErrSeparateAudioTracks = errors.New("master playlist has separate audio tracks; use yt-dlp for proper audio/video merging")

// ErrUnsupportedEncryption is returned for segments encrypted with a method
// other than AES-128 (SAMPLE-AES and its variants): concatenating them cannot
// produce a playable file, so the caller should hand the stream to yt-dlp.
// It is returned while parsing, before any output file exists.
var ErrUnsupportedEncryption = errors.New("HLS segments use an encryption method the native downloader cannot undo; use yt-dlp")

// newNoBodyRequest creates a body-less GET request whose GetBody returns
// http.NoBody. Surf's HTTP/2->HTTP/1.1 fallback (and Go's HTTP/2 transport
// auto-retry) refuses to retry requests with a non-nil Body but a nil GetBody,
// failing with "cannot retry because req.GetBody is nil" whenever a CDN does
// not negotiate h2 (e.g. some googlevideo edges). A GetBody that reproduces
// the empty body lets those retries succeed.
func newNoBodyRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
	return req, nil
}

// Segment represents a single HLS segment
type Segment struct {
	URL      string
	Index    int
	Duration float64
	Title    string
	// Key is the AES-128 key in force for the segment, nil when it is plain.
	Key *SegmentKey
	// Sequence is the segment's media sequence number, the implicit IV.
	Sequence uint64
}

// SegmentKey is an #EXT-X-KEY with METHOD=AES-128 (RFC 8216 §4.3.2.4): the
// segment is AES-128-CBC encrypted with PKCS#7 padding under the 16-byte key
// served at URI.
type SegmentKey struct {
	URI string // absolute
	IV  []byte // nil: the segment's media sequence number is the IV
}

// M3U8Playlist represents the HLS playlist structure
type M3U8Playlist struct {
	Version        string
	TargetDuration float64
	MediaSequence  int
	Segments       []Segment
	EndList        bool
	PlaylistType   string
	// InitSegmentURL is the #EXT-X-MAP initialisation segment, for playlists
	// whose media segments are fragmented MP4 rather than MPEG-TS.
	//
	// An fMP4 segment starts at a `moof` box and carries no headers of its own:
	// the track definitions live once, in this init segment. Concatenating only
	// the media segments therefore produces a file nothing can open —
	// "trun track id unknown, no tfhd was found / error reading header" — which
	// is exactly what AnimeFire's downloads turned into. The bytes were all
	// there; the header never was.
	//
	// Empty for MPEG-TS playlists, which are self-describing.
	InitSegmentURL string
	// InitSegmentKey is the AES-128 key in force for the init segment, if any.
	InitSegmentKey *SegmentKey
}

// Downloader handles HLS downloads
type Downloader struct {
	client *http.Client
}

// NewDownloader creates a new HLS downloader with a plain Go HTTP client.
// Prefer NewDownloaderWithClient to supply a surf-backed client with Chrome
// TLS fingerprinting, which is required by most CDNs.
func NewDownloader() *Downloader {
	// Force HTTP/1.1 by disabling HTTP/2.  CDN servers often reset
	// multiplexed HTTP/2 streams with INTERNAL_ERROR when many segments
	// are fetched concurrently over a single connection.  HTTP/1.1 opens
	// a separate TCP connection per request, avoiding this issue.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		// Setting TLSNextProto to an empty map disables HTTP/2
		TLSNextProto:        make(map[string]func(string, *tls.Conn) http.RoundTripper),
		MaxIdleConns:        48,
		MaxIdleConnsPerHost: 24,
		IdleConnTimeout:     90 * time.Second,
		// SSRF protection: validates resolved IPs before connecting
		DialContext: api.SafeDialContext(30 * time.Second),
	}

	return &Downloader{
		client: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: transport,
		},
	}
}

// NewDownloaderWithClient creates an HLS downloader using the provided
// *http.Client. Pass a surf-backed client (surf.NewClient().…Build().…Std())
// so that requests carry a real browser TLS fingerprint and bypass CDN
// anti-bot checks that reject plain Go clients.
func NewDownloaderWithClient(client *http.Client) *Downloader {
	return &Downloader{client: client}
}

// sanitizeOutputPath validates and cleans the output path to prevent directory traversal
func sanitizeOutputPath(path string) (string, error) {
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path contains directory traversal: %s", path)
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Resolve to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

// Download downloads an HLS stream to the specified output file with concurrent segment downloads
func (d *Downloader) Download(ctx context.Context, url, output string, headers map[string]string) error {
	return d.DownloadWithProgress(ctx, url, output, headers, nil)
}

// parsePlaylist downloads and parses the M3U8 playlist
func (d *Downloader) parsePlaylist(ctx context.Context, url string, headers map[string]string) (*M3U8Playlist, error) {
	req, err := newNoBodyRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Some HLS streams require a proper User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	}

	resp, err := d.client.Do(req) // #nosec G704
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Peek at the response to verify it's an M3U8 playlist before scanning.
	// Non-HLS responses (binary video, HTML) cause "bufio.Scanner: token too long".
	peekBuf := make([]byte, 64)
	peekN, _ := io.ReadFull(resp.Body, peekBuf)
	peek := string(peekBuf[:peekN])
	if peekN > 0 && !strings.Contains(peek, "#EXTM3U") && !strings.Contains(peek, "#EXT-X-") {
		snip := peek
		if len(snip) > 32 {
			snip = snip[:32]
		}
		return nil, fmt.Errorf("response is not an HLS playlist (starts with %q)", snip)
	}
	body := io.MultiReader(strings.NewReader(peek), resp.Body)
	scanner := bufio.NewScanner(body)

	// Check if this is a master playlist by looking for STREAM-INF tags
	isMasterPlaylist := false
	var masterPlaylistLines []string

	// First pass: collect all lines and determine if it's a master playlist
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		masterPlaylistLines = append(masterPlaylistLines, line)

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			isMasterPlaylist = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if isMasterPlaylist {
		// Check for separate audio tracks — these require ffmpeg muxing that
		// only yt-dlp handles. Downloading just the video playlist would
		// produce a file without audio.
		for _, line := range masterPlaylistLines {
			if strings.HasPrefix(line, "#EXT-X-MEDIA:") && strings.Contains(line, "TYPE=AUDIO") && strings.Contains(line, "URI=") {
				return nil, ErrSeparateAudioTracks
			}
		}

		// For master playlists, select the highest quality stream (largest bandwidth)
		selectedMediaPlaylistURL := d.selectBestStream(masterPlaylistLines, url)
		if selectedMediaPlaylistURL != "" {
			// Recursively parse the selected media playlist
			return d.parseMediaPlaylist(ctx, selectedMediaPlaylistURL, headers)
		}
		return nil, fmt.Errorf("no suitable stream found in master playlist")
	}

	// It's a media playlist, parse it directly
	return d.parseMediaPlaylistLines(masterPlaylistLines, url)
}

// bandwidthRe extracts the BANDWIDTH attribute from #EXT-X-STREAM-INF tags.
var bandwidthRe = regexp.MustCompile(`BANDWIDTH=(\d+)`)

// resolutionRe extracts the picture height from a RESOLUTION=WxH attribute.
var resolutionRe = regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)

// selectBestStream finds the highest quality stream from a master playlist:
// the tallest picture, and among equally tall ones the highest bandwidth.
// Height comes first because bandwidth alone can rank a 720p variant in an
// older codec above a 1080p one in a more efficient codec.
func (d *Downloader) selectBestStream(lines []string, baseURL string) string {
	type StreamInfo struct {
		URL       string
		Height    int
		Bandwidth int
	}

	var streams []StreamInfo

	for i, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			info := StreamInfo{}
			if bwMatch := bandwidthRe.FindStringSubmatch(line); len(bwMatch) > 1 {
				info.Bandwidth, _ = strconv.Atoi(bwMatch[1])
			}
			if resMatch := resolutionRe.FindStringSubmatch(line); len(resMatch) > 1 {
				info.Height, _ = strconv.Atoi(resMatch[1])
			}

			// Next non-tag line should be the URL
			if i+1 < len(lines) {
				if urlLine := strings.TrimSpace(lines[i+1]); urlLine != "" && !strings.HasPrefix(urlLine, "#") {
					info.URL = resolveReference(baseURL, urlLine)
					streams = append(streams, info)
				}
			}
		}
	}

	if len(streams) == 0 {
		return ""
	}
	best := streams[0]
	for _, s := range streams[1:] {
		if s.Height > best.Height || (s.Height == best.Height && s.Bandwidth > best.Bandwidth) {
			best = s
		}
	}
	return best.URL
}

// parseMediaPlaylist fetches and parses a media playlist (not master)
func (d *Downloader) parseMediaPlaylist(ctx context.Context, url string, headers map[string]string) (*M3U8Playlist, error) {
	req, err := newNoBodyRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Some HLS streams require a proper User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	}

	resp, err := d.client.Do(req) // #nosec G704
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Peek at the response to verify it's an M3U8 playlist.
	peekBuf := make([]byte, 64)
	peekN, _ := io.ReadFull(resp.Body, peekBuf)
	peek := string(peekBuf[:peekN])
	if peekN > 0 && !strings.Contains(peek, "#EXTM3U") && !strings.Contains(peek, "#EXT-X-") {
		snip := peek
		if len(snip) > 32 {
			snip = snip[:32]
		}
		return nil, fmt.Errorf("response is not an HLS playlist (starts with %q)", snip)
	}
	body := io.MultiReader(strings.NewReader(peek), resp.Body)
	scanner := bufio.NewScanner(body)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return d.parseMediaPlaylistLines(lines, url)
}

// parseMediaPlaylistLines parses lines from a media playlist
// extXMapURIRe pulls the URI out of an #EXT-X-MAP attribute list.
//
// The tag is an attribute list (URI, optionally BYTERANGE), so the value is
// read by name rather than by position — BYTERANGE may precede it, and a
// future attribute must not shift what we read.
var extXMapURIRe = regexp.MustCompile(`(?i)URI="([^"]*)"`)

// extXMapURI returns the initialisation segment URI from an #EXT-X-MAP
// attribute list, or "" when there is none.
func extXMapURI(attrs string) string {
	m := extXMapURIRe.FindStringSubmatch(attrs)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// resolveReference turns a reference in a playlist — a variant, segment, init
// segment or key URI — into an absolute URL against the playlist's own URL,
// as RFC 3986 §5.2 (and every player) does: "/x" is host-relative, "../x"
// climbs, "//host/x" keeps the scheme, and a "/" inside the playlist URL's
// query is not a directory.
func resolveReference(playlistURL, ref string) string {
	r, refErr := neturl.Parse(ref)
	if refErr == nil && r.IsAbs() {
		return ref
	}
	if base, err := neturl.Parse(playlistURL); err == nil && base.IsAbs() && refErr == nil {
		return base.ResolveReference(r).String()
	}
	// No absolute base to resolve against (playlists are always fetched by
	// absolute URL, so only tests get here): join at the last slash.
	if dir, _, ok := strings.CutLast(playlistURL, "/"); ok {
		return dir + "/" + ref
	}
	return playlistURL + "/" + ref
}

var (
	extXKeyMethodRe = regexp.MustCompile(`(?i)METHOD=([A-Z0-9-]+)`)
	extXKeyIVRe     = regexp.MustCompile(`(?i)IV=0x([0-9a-f]+)`)
)

// parseKey reads an #EXT-X-KEY attribute list: nil for METHOD=NONE, the key
// for AES-128, ErrUnsupportedEncryption for anything else.
func parseKey(attrs, playlistURL string) (*SegmentKey, error) {
	m := extXKeyMethodRe.FindStringSubmatch(attrs)
	if m == nil {
		return nil, fmt.Errorf("%w: EXT-X-KEY without METHOD", ErrUnsupportedEncryption)
	}
	switch method := strings.ToUpper(m[1]); method {
	case "NONE":
		return nil, nil
	case "AES-128":
		uri := extXMapURI(attrs)
		if uri == "" {
			return nil, fmt.Errorf("%w: AES-128 key without URI", ErrUnsupportedEncryption)
		}
		key := &SegmentKey{URI: resolveReference(playlistURL, uri)}
		if ivm := extXKeyIVRe.FindStringSubmatch(attrs); ivm != nil {
			raw := ivm[1]
			if len(raw) > 2*aes.BlockSize {
				return nil, fmt.Errorf("%w: IV longer than 128 bits", ErrUnsupportedEncryption)
			}
			if len(raw)%2 == 1 {
				raw = "0" + raw
			}
			iv, err := hex.DecodeString(raw)
			if err != nil {
				return nil, fmt.Errorf("%w: bad IV: %w", ErrUnsupportedEncryption, err)
			}
			// A shorter hex sequence is the same 128-bit number, zero-padded.
			key.IV = append(make([]byte, aes.BlockSize-len(iv)), iv...)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("%w: METHOD=%s", ErrUnsupportedEncryption, method)
	}
}

func (d *Downloader) parseMediaPlaylistLines(lines []string, url string) (*M3U8Playlist, error) {
	playlist := &M3U8Playlist{
		Segments: make([]Segment, 0),
	}

	segmentIndex := 0
	var key *SegmentKey
	var nextSeq uint64 // media sequence number of the next segment
	for i, line := range lines {
		if strings.HasPrefix(line, "#EXTM3U") {
			continue // Header
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-VERSION:"); ok {
			playlist.Version = after
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-TARGETDURATION:"); ok {
			durationStr := after
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err == nil {
				playlist.TargetDuration = duration
			}
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"); ok {
			seqStr := after
			seq, err := strconv.Atoi(seqStr)
			if err == nil {
				playlist.MediaSequence = seq
			}
			// The spec's decimal-integer is 0..2^64-1: read it unsigned.
			if u, err := strconv.ParseUint(strings.TrimSpace(seqStr), 10, 64); err == nil {
				nextSeq = u
			}
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-PLAYLIST-TYPE:"); ok {
			playlist.PlaylistType = after
		} else if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			playlist.EndList = true
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-KEY:"); ok {
			k, err := parseKey(after, url)
			if err != nil {
				return nil, err
			}
			key = k
		} else if after, ok := strings.CutPrefix(line, "#EXT-X-MAP:"); ok {
			if uri := extXMapURI(after); uri != "" {
				playlist.InitSegmentURL = resolveReference(url, uri)
				playlist.InitSegmentKey = key
			}
		} else if after, ok := strings.CutPrefix(line, "#EXTINF:"); ok {
			// Parse duration and title
			infLine := after
			parts := strings.SplitN(infLine, ",", 2)
			var duration float64
			if len(parts) > 0 {
				duration, _ = strconv.ParseFloat(strings.TrimRight(parts[0], ", "), 64)
			}

			var title string
			if len(parts) > 1 {
				title = strings.TrimSpace(parts[1])
			}

			// Next line should be the URL
			if i+1 < len(lines) {
				segmentURL := strings.TrimSpace(lines[i+1])
				if segmentURL != "" && !strings.HasPrefix(segmentURL, "#") {
					playlist.Segments = append(playlist.Segments, Segment{
						URL:      resolveReference(url, segmentURL),
						Index:    segmentIndex,
						Duration: duration,
						Title:    title,
						Key:      key,
						Sequence: nextSeq,
					})
					nextSeq++
					segmentIndex++
				}
			}
		}
	}

	return playlist, nil
}

// downloadSegment downloads a single segment
// segmentBackoff is the progressive delay before retrying a failed segment:
// 1s, 2s, 3s…
func segmentBackoff(attempt int) time.Duration {
	return time.Duration(attempt+1) * time.Second
}

// waitBackoff sleeps for d unless ctx is cancelled first.
//
// This used to be a plain time.Sleep, which kept a cancelled download alive for
// up to another 5 seconds per in-flight segment while the user waited for the
// UI to come back.
func waitBackoff(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (d *Downloader) downloadSegment(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	maxRetries := 5
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := newNoBodyRequest(ctx, url)
		if err != nil {
			return nil, err
		}

		// Add custom headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Some HLS streams require a proper User-Agent
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		}

		resp, err := d.client.Do(req) // #nosec G704
		if err != nil {
			if attempt < maxRetries {
				if werr := waitBackoff(ctx, segmentBackoff(attempt)); werr != nil {
					return nil, werr
				}
				continue
			}
			return nil, err
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 30*1024*1024))
		_ = resp.Body.Close()

		if err != nil {
			if attempt < maxRetries {
				if werr := waitBackoff(ctx, segmentBackoff(attempt)); werr != nil {
					return nil, werr
				}
				continue
			}
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			if attempt < maxRetries {
				if werr := waitBackoff(ctx, segmentBackoff(attempt)); werr != nil {
					return nil, werr
				}
				continue
			}
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed to download segment after %d attempts", maxRetries+1)
}

// ProgressCallback reports download progress.
// bytesWritten  = cumulative bytes flushed to the output file (real disk size).
// segmentsWritten = number of segments sequentially written to disk.
// totalSegments   = total number of segments in the playlist.
type ProgressCallback func(bytesWritten int64, segmentsWritten, totalSegments int)

// DownloadWithProgress downloads HLS content with progress reporting
func (d *Downloader) DownloadWithProgress(ctx context.Context, url, output string, headers map[string]string, progressCallback ProgressCallback) error {
	playlist, err := d.parsePlaylist(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("failed to parse playlist: %w", err)
	}

	if len(playlist.Segments) == 0 {
		return fmt.Errorf("playlist has no segments to download")
	}

	// Sanitize output path to prevent directory traversal (G304)
	sanitizedOutput, err := sanitizeOutputPath(output)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	output = sanitizedOutput

	// Create output directory if it doesn't exist
	if err = os.MkdirAll(filepath.Dir(output), 0o750); err != nil { // #nosec G301 - directory needs to be accessible
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open the output file for writing with buffered I/O for better throughput
	outFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 - path sanitized above
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()
	bufferedWriter := bufio.NewWriterSize(outFile, 512*1024) // 512KB write buffer
	defer func() { _ = bufferedWriter.Flush() }()

	totalSegments := len(playlist.Segments)
	var downloadedSegments atomic.Int32
	var bytesWritten int64 // cumulative bytes flushed to disk

	// The initialisation segment goes first, before any media segment.
	//
	// For a fragmented-MP4 playlist it is the only place the track definitions
	// exist; every media segment after it is a bare `moof`+`mdat` fragment.
	// Writing the media segments alone produced a file that ffmpeg and mpv both
	// refused — "no tfhd was found", "Failed to recognize file format" — so the
	// download looked like it had worked (100+ MB on disk, well past the size
	// check) and nothing could play it.
	//
	// Fetched synchronously and written before the workers start, because its
	// position in the file is not negotiable.
	keys := &keyCache{keys: map[string][]byte{}}
	if playlist.InitSegmentURL != "" {
		initData, initErr := d.downloadSegment(ctx, playlist.InitSegmentURL, headers)
		if initErr == nil && playlist.InitSegmentKey != nil {
			// The init section's IV must be explicit; the first segment's
			// sequence number is the only sensible stand-in when it is not.
			initData, initErr = d.decrypt(ctx, initData, playlist.InitSegmentKey, playlist.Segments[0].Sequence, headers, keys)
		}
		if initErr != nil {
			return fmt.Errorf("failed to download HLS init segment: %w", initErr)
		}
		n, werr := bufferedWriter.Write(initData)
		if werr != nil {
			return fmt.Errorf("failed to write HLS init segment: %w", werr)
		}
		bytesWritten += int64(n)
		util.Debug("HLS init segment written", "bytes", n, "url", playlist.InitSegmentURL)
	}

	// Report initial progress
	if progressCallback != nil {
		progressCallback(0, 0, totalSegments)
	}

	// Concurrent download configuration
	// 16 workers provides high parallelism for fast downloads
	const maxWorkers = 24

	type job struct {
		index   int
		segment Segment
	}
	jobs := make(chan job, totalSegments)

	type result struct {
		index int
		data  []byte
		err   error
	}
	results := make(chan result, totalSegments)

	// Start workers
	var wg sync.WaitGroup
	for range maxWorkers {
		wg.Go(func() {
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				data, err := d.downloadSegment(ctx, j.segment.URL, headers)
				if err == nil && j.segment.Key != nil {
					data, err = d.decrypt(ctx, data, j.segment.Key, j.segment.Sequence, headers, keys)
				}
				results <- result{index: j.index, data: data, err: err}
			}
		})
	}

	// Fill job queue
	for i, segment := range playlist.Segments {
		jobs <- job{index: i, segment: segment}
	}
	close(jobs)

	// Collect results and write in order
	segmentBuffer := make(map[int][]byte)
	nextIndex := 0
	var failedSegments int
	var firstErr error

	for range totalSegments {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res := <-results:
			if res.err != nil {
				failedSegments++
				if firstErr == nil {
					firstErr = res.err
				}
				// Still increment downloadedSegments so the write-loop
				// doesn't block forever waiting for this index.
				// We write an empty slice for the gap so subsequent
				// segments can still be flushed in order.
				segmentBuffer[res.index] = nil
			} else {
				segmentBuffer[res.index] = res.data
			}
			downloadedSegments.Add(1)

			// Write available sequential segments
			for {
				data, ok := segmentBuffer[nextIndex]
				if !ok {
					break
				}

				if data != nil {
					n, werr := bufferedWriter.Write(data)
					if werr != nil {
						if firstErr == nil {
							firstErr = fmt.Errorf("failed to write segment %d: %w", nextIndex, werr)
						}
					} else {
						bytesWritten += int64(n)
					}
				}

				delete(segmentBuffer, nextIndex)
				nextIndex++
			}

			// Report progress based on segments WRITTEN to disk (not just downloaded)
			if progressCallback != nil {
				progressCallback(bytesWritten, nextIndex, totalSegments)
			}
		}
	}

	wg.Wait()

	// Fail if too many segments were lost (>5%).
	// A handful of missing segments (<5%) in a long stream is tolerable —
	// the video will have brief glitches but is otherwise watchable.
	if failedSegments > 0 {
		failRatio := float64(failedSegments) / float64(totalSegments)
		if failRatio > 0.05 {
			return fmt.Errorf("download incomplete: %d/%d segments failed (%.0f%%): %w",
				failedSegments, totalSegments, failRatio*100, firstErr)
		}
		// Log minor losses but don't fail — use structured logger instead of
		// fmt.Printf because bubbletea's progress bar owns stdout at this point.
		util.Debug("HLS segments partially failed",
			"failed", failedSegments,
			"total", totalSegments,
			"failPercent", fmt.Sprintf("%.1f%%", failRatio*100))
	}

	return nil
}

// keyCache holds the AES-128 keys a download has fetched, by URI. A key is
// fetched once however many segments use it; the lock is held during the
// fetch so concurrent workers wait for it instead of each asking the server.
type keyCache struct {
	mu   sync.Mutex
	keys map[string][]byte
}

// decrypt undoes AES-128 (RFC 8216 §5.2) on one segment: CBC under the key at
// k.URI, with k.IV or else the segment's sequence number as a big-endian
// 128-bit IV, then PKCS#7 padding removed.
func (d *Downloader) decrypt(ctx context.Context, data []byte, k *SegmentKey, seq uint64, headers map[string]string, cache *keyCache) ([]byte, error) {
	cache.mu.Lock()
	key, ok := cache.keys[k.URI]
	if !ok {
		var err error
		key, err = d.downloadSegment(ctx, k.URI, headers)
		if err == nil && len(key) != aes.BlockSize {
			err = fmt.Errorf("key is %d bytes, want %d", len(key), aes.BlockSize)
		}
		if err != nil {
			cache.mu.Unlock()
			return nil, fmt.Errorf("HLS key %s: %w", k.URI, err)
		}
		cache.keys[k.URI] = key
	}
	cache.mu.Unlock()

	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("encrypted segment is %d bytes, not a whole number of AES blocks", len(data))
	}
	iv := k.IV
	if iv == nil {
		iv = make([]byte, aes.BlockSize)
		binary.BigEndian.PutUint64(iv[8:], seq)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, data)
	pad := int(plain[len(plain)-1])
	if pad == 0 || pad > aes.BlockSize || !bytes.Equal(plain[len(plain)-pad:], bytes.Repeat([]byte{byte(pad)}, pad)) {
		return nil, errors.New("decrypted segment has invalid PKCS#7 padding (wrong key or IV)")
	}
	return plain[:len(plain)-pad], nil
}

// DownloadToFile is a convenience function to download HLS to a file
func DownloadToFile(ctx context.Context, streamURL, outputPath string, headers map[string]string, progressCallback ProgressCallback) error {
	downloader := NewDownloader()
	return downloader.DownloadWithProgress(ctx, streamURL, outputPath, headers, progressCallback)
}

// DownloadToFileWithClient is like DownloadToFile but uses the provided
// *http.Client (e.g. a surf-backed client with Chrome TLS fingerprinting).
func DownloadToFileWithClient(ctx context.Context, client *http.Client, streamURL, outputPath string, headers map[string]string, progressCallback ProgressCallback) error {
	downloader := NewDownloaderWithClient(client)
	return downloader.DownloadWithProgress(ctx, streamURL, outputPath, headers, progressCallback)
}
