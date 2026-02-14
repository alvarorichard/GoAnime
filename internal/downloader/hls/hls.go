// Package hls provides HLS (HTTP Live Streaming) download functionality
package hls

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Segment represents a single HLS segment
type Segment struct {
	URL      string
	Index    int
	Duration float64
	Title    string
}

// M3U8Playlist represents the HLS playlist structure
type M3U8Playlist struct {
	Version        string
	TargetDuration float64
	MediaSequence  int
	Segments       []Segment
	EndList        bool
	PlaylistType   string
}

// Downloader handles HLS downloads
type Downloader struct {
	client *http.Client
}

// NewDownloader creates a new HLS downloader
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
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Downloader{
		client: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: transport,
		},
	}
}

// sanitizeOutputPath validates and cleans the output path to prevent directory traversal
func sanitizeOutputPath(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Resolve to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Ensure no directory traversal attempts
	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("path contains directory traversal")
	}

	return absPath, nil
}

// Download downloads an HLS stream to the specified output file with concurrent segment downloads
func (d *Downloader) Download(ctx context.Context, url, output string, headers map[string]string) error {
	return d.DownloadWithProgress(ctx, url, output, headers, nil)
}

// parsePlaylist downloads and parses the M3U8 playlist
func (d *Downloader) parsePlaylist(ctx context.Context, url string, headers map[string]string) (*M3U8Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	scanner := bufio.NewScanner(resp.Body)

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

// selectBestStream finds the highest quality stream from a master playlist
func (d *Downloader) selectBestStream(lines []string, baseURL string) string {
	type StreamInfo struct {
		URL       string
		Bandwidth int
	}

	var streams []StreamInfo

	for i, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// Parse bandwidth from the tag
			bandwidth := 0
			if bwMatch := regexp.MustCompile(`BANDWIDTH=(\d+)`).FindStringSubmatch(line); len(bwMatch) > 1 {
				if bw, err := strconv.Atoi(bwMatch[1]); err == nil {
					bandwidth = bw
				}
			}

			// Next non-tag line should be the URL
			if i+1 < len(lines) {
				urlLine := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(urlLine, "http") {
					streams = append(streams, StreamInfo{URL: urlLine, Bandwidth: bandwidth})
				} else {
					// Handle relative URL
					if idx := strings.LastIndex(baseURL, "/"); idx != -1 {
						streams = append(streams, StreamInfo{
							URL:       baseURL[:idx+1] + urlLine,
							Bandwidth: bandwidth,
						})
					}
				}
			}
		}
	}

	// Select the stream with the highest bandwidth
	if len(streams) > 0 {
		best := streams[0]
		for _, s := range streams[1:] {
			if s.Bandwidth > best.Bandwidth {
				best = s
			}
		}
		return best.URL
	}

	return ""
}

// parseMediaPlaylist fetches and parses a media playlist (not master)
func (d *Downloader) parseMediaPlaylist(ctx context.Context, url string, headers map[string]string) (*M3U8Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	scanner := bufio.NewScanner(resp.Body)
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
func (d *Downloader) parseMediaPlaylistLines(lines []string, url string) (*M3U8Playlist, error) {
	playlist := &M3U8Playlist{
		Segments: make([]Segment, 0),
	}

	segmentIndex := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "#EXTM3U") {
			continue // Header
		} else if strings.HasPrefix(line, "#EXT-X-VERSION:") {
			playlist.Version = strings.TrimPrefix(line, "#EXT-X-VERSION:")
		} else if strings.HasPrefix(line, "#EXT-X-TARGETDURATION:") {
			durationStr := strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:")
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err == nil {
				playlist.TargetDuration = duration
			}
		} else if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			seqStr := strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:")
			seq, err := strconv.Atoi(seqStr)
			if err == nil {
				playlist.MediaSequence = seq
			}
		} else if strings.HasPrefix(line, "#EXT-X-PLAYLIST-TYPE:") {
			playlist.PlaylistType = strings.TrimPrefix(line, "#EXT-X-PLAYLIST-TYPE:")
		} else if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			playlist.EndList = true
		} else if strings.HasPrefix(line, "#EXTINF:") {
			// Parse duration and title
			infLine := strings.TrimPrefix(line, "#EXTINF:")
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
					// Handle relative URLs
					if !strings.HasPrefix(segmentURL, "http") {
						baseURL := url
						if idx := strings.LastIndex(baseURL, "/"); idx != -1 {
							baseURL = baseURL[:idx+1]
						} else {
							baseURL = baseURL + "/"
						}
						segmentURL = baseURL + segmentURL
					}

					playlist.Segments = append(playlist.Segments, Segment{
						URL:      segmentURL,
						Index:    segmentIndex,
						Duration: duration,
						Title:    title,
					})
					segmentIndex++
				}
			}
		}
	}

	return playlist, nil
}

// downloadSegment downloads a single segment
func (d *Downloader) downloadSegment(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	maxRetries := 5
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
				time.Sleep(time.Duration(attempt+1) * time.Second) // progressive backoff: 1s, 2s, 3s…
				continue
			}
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed to download segment after %d attempts", maxRetries+1)
}

// ProgressCallback is a function that reports download progress
type ProgressCallback func(downloaded, total int)

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
	if err = os.MkdirAll(filepath.Dir(output), 0750); err != nil { // #nosec G301 - directory needs to be accessible
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open the output file for writing
	outFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600) // #nosec G304 - path sanitized above
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	totalSegments := len(playlist.Segments)
	var downloadedSegments int32

	// Report initial progress
	if progressCallback != nil {
		progressCallback(0, totalSegments)
	}

	// Concurrent download configuration
	// 8 workers provides good parallelism without triggering most CDN rate limits
	const maxWorkers = 8

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
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				data, err := d.downloadSegment(ctx, j.segment.URL, headers)
				results <- result{index: j.index, data: data, err: err}
			}
		}()
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

	for i := 0; i < totalSegments; i++ {
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
			atomic.AddInt32(&downloadedSegments, 1)

			// Write available sequential segments
			for {
				data, ok := segmentBuffer[nextIndex]
				if !ok {
					break
				}

				if data != nil {
					if _, err := outFile.Write(data); err != nil {
						if firstErr == nil {
							firstErr = fmt.Errorf("failed to write segment %d: %w", nextIndex, err)
						}
					}
				}

				delete(segmentBuffer, nextIndex)
				nextIndex++
			}

			// Report progress
			if progressCallback != nil {
				progressCallback(int(downloadedSegments), totalSegments)
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
		// Log minor losses but don't fail
		if util.IsDebug {
			fmt.Printf("Note: %d/%d segments could not be downloaded (%.1f%%), minor glitches possible\n",
				failedSegments, totalSegments, failRatio*100)
		}
	}

	return nil
}

// DownloadToFile is a convenience function to download HLS to a file
func DownloadToFile(ctx context.Context, streamURL, outputPath string, headers map[string]string, progressCallback ProgressCallback) error {
	downloader := NewDownloader()
	return downloader.DownloadWithProgress(ctx, streamURL, outputPath, headers, progressCallback)
}
