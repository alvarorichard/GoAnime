package player

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/lrstanley/go-ytdlp"
)

// Pre-compiled regexes for download quality parsing
var (
	digitsRe               = regexp.MustCompile(`\d+`)
	resolutionMp4Re        = regexp.MustCompile(`(\d{3,4})p?\.mp4`)
	downloadPartRetryDelay = 500 * time.Millisecond
)

// downloadPart downloads a part of the video file using HTTP Range Requests.
// Automatically retries with resume on connection drops (up to 20 stale retries).
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model) error {
	partFilePath, err := safePartPath(destPath, part)
	if err != nil {
		return err
	}
	// #nosec G304: path validated by safePartPath to remain within destination directory
	file, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			util.Logger.Warn("Error closing file", "error", err)
		}
	}()

	current := from
	maxStaleRetries := 20
	staleRetries := 0

	for attempt := 0; current <= to; attempt++ {
		if staleRetries >= maxStaleRetries {
			return fmt.Errorf("part %d: max retries (%d) exceeded at offset %d/%d", part, maxStaleRetries, current, to)
		}
		if attempt > 0 {
			util.Debugf("Download part %d: resuming at byte %d (attempt %d)", part, current, attempt+1)
			time.Sleep(downloadPartRetryDelay)
		}

		beforeRead := current

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", current, to))
		if strings.Contains(url, "allanime.day") || strings.Contains(url, "allanime.pro") {
			req.Header.Set("Referer", "https://allanime.to")
		}

		resp, err := client.Do(req) // #nosec G704
		if err != nil {
			util.Debugf("Download part %d: request error: %v", part, err)
			staleRetries++
			continue
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			statusCode := resp.StatusCode
			status := resp.Status
			if cErr := resp.Body.Close(); cErr != nil {
				util.Logger.Warn("Error closing response body", "error", cErr)
			}
			if statusCode == http.StatusForbidden || statusCode == http.StatusNotFound {
				return scraper.NewDownloadExpiredError("Download", "http-range", statusCode, fmt.Errorf("HTTP %d: %s", statusCode, status))
			}
			util.Debugf("Download part %d: unexpected status %d", part, statusCode)
			staleRetries++
			continue
		}

		buf := make([]byte, 256*1024)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := file.Write(buf[:n]); writeErr != nil {
					if cErr := resp.Body.Close(); cErr != nil {
						util.Logger.Warn("Error closing response body", "error", cErr)
					}
					return writeErr
				}
				current += int64(n)
				m.addProgressReceived(int64(n))
			}
			if readErr == io.EOF {
				if cErr := resp.Body.Close(); cErr != nil {
					util.Logger.Warn("Error closing response body", "error", cErr)
				}
				break
			}
			if readErr != nil {
				if cErr := resp.Body.Close(); cErr != nil {
					util.Logger.Warn("Error closing response body", "error", cErr)
				}
				util.Debugf("Download part %d: read error after %d bytes: %v", part, current-from, readErr)
				break
			}
		}

		if current > to {
			break
		}

		if current == beforeRead {
			staleRetries++
		} else {
			staleRetries = 0
		}
	}

	return nil
}

// combineParts combines downloaded parts into a single final file.
func combineParts(destPath string, numThreads int) error {
	outFile, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			util.Logger.Warn("Error closing output file", "error", err)
		}
	}()
	for i := range numThreads {
		partFilePath, err := safePartPath(destPath, i)
		if err != nil {
			return err
		}
		// #nosec G304: path validated by safePartPath to remain within destination directory
		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, partFile); err != nil {
			if closeErr := partFile.Close(); closeErr != nil {
				util.Logger.Warn("Error closing part file", "error", closeErr)
			}
			return err
		}
		if err := partFile.Close(); err != nil {
			util.Logger.Warn("Error closing part file", "error", err)
		}
		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}
	return nil
}

// safePartPath builds the part file path and ensures it stays within the destination directory
func safePartPath(destPath string, part int) (string, error) {
	dir := filepath.Clean(filepath.Dir(destPath))
	base := filepath.Base(destPath)
	name := fmt.Sprintf("%s.part%d", base, part)
	joined := filepath.Join(dir, name)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absDir, absFile)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid part path: %s", joined)
	}
	return joined, nil
}

// isBloggerProxyURL returns true if the URL points to the local Blogger proxy.
// The proxy runs on 127.0.0.1 and is safe to access without SSRF protection.
func isBloggerProxyURL(u string) bool {
	return strings.Contains(u, "127.0.0.1") && strings.Contains(u, "blogger_proxy")
}

// LooksLikeHLS returns true if the URL appears to be an HLS stream.
// Matches .m3u8 extensions and /hls/ path segments commonly used by CDNs
// that serve HLS playlists without a .m3u8 extension.
func LooksLikeHLS(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, ".m3u8") ||
		strings.Contains(lower, "m3u8") ||
		strings.Contains(lower, "/hls/")
}

// hasUnsafeExtension returns true if the URL has a file extension that yt-dlp
// will reject as unusual (e.g. .aspx from SharePoint/AllAnime CDNs).
func hasUnsafeExtension(u string) bool {
	// Extract path from URL, ignore query string
	path := u
	if before, _, ok := strings.Cut(u, "?"); ok {
		path = before
	}
	lower := strings.ToLower(path)
	for _, ext := range []string{".aspx", ".asp", ".php", ".jsp", ".cgi"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func isAnimeFireVideoAPIURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "animefire.io/video/") ||
		strings.Contains(lower, "animefire.plus/video/")
}

func resolveDownloadURL(videoURL string) (string, error) {
	if !isAnimeFireVideoAPIURL(videoURL) {
		return videoURL, nil
	}

	util.Debug("Resolving AnimeFire video API URL for download", "url", videoURL)
	resp, err := api.SafeGet(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch AnimeFire video API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Warn("Error closing AnimeFire video API response", "error", err)
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read AnimeFire video API: %w", err)
	}

	selected, err := selectAnimeFireDownloadSource(body, util.GlobalQuality)
	if err != nil {
		return "", err
	}
	util.Debug("Resolved AnimeFire download URL", "quality", util.GlobalQuality, "url", selected)
	return selected, nil
}

func selectAnimeFireDownloadSource(body []byte, quality string) (string, error) {
	candidates, err := selectAnimeFireDownloadCandidates(body, quality)
	if err != nil {
		return "", err
	}
	return candidates[0], nil
}

func selectAnimeFireDownloadCandidates(body []byte, quality string) ([]string, error) {
	var videoResponse VideoResponse
	if err := json.Unmarshal(body, &videoResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AnimeFire video API: %w", err)
	}
	if len(videoResponse.Data) > 0 {
		candidates := orderAnimeFireSources(videoResponse.Data, quality)
		if len(candidates) == 0 {
			return nil, errors.New("AnimeFire video API returned no selectable source")
		}
		return candidates, nil
	}
	if strings.Contains(videoResponse.Token, "blogger.com") {
		return []string{videoResponse.Token}, nil
	}
	return nil, errors.New("AnimeFire video API returned no sources")
}

func orderAnimeFireSources(videoData []VideoData, quality string) []string {
	if len(videoData) == 0 {
		return nil
	}

	preferred := selectQualityFromOptions(videoData, quality)
	ordered := append([]VideoData(nil), videoData...)
	preferredQuality := strings.ToLower(strings.TrimSpace(quality))
	targetResolution := extractResolution(preferredQuality)

	sort.SliceStable(ordered, func(i, j int) bool {
		left := extractResolution(ordered[i].Label)
		right := extractResolution(ordered[j].Label)
		switch {
		case preferredQuality == "worst":
			return left < right
		case preferredQuality == "best" || preferredQuality == "":
			return left > right
		case targetResolution > 0:
			leftDiff := abs(left - targetResolution)
			rightDiff := abs(right - targetResolution)
			if leftDiff == rightDiff {
				return left > right
			}
			return leftDiff < rightDiff
		default:
			return false
		}
	})

	var candidates []string
	seen := make(map[string]struct{}, len(videoData))
	addCandidate := func(src string) {
		if src == "" {
			return
		}
		if _, ok := seen[src]; ok {
			return
		}
		seen[src] = struct{}{}
		candidates = append(candidates, src)
	}

	addCandidate(preferred)
	for _, v := range ordered {
		addCandidate(v.Src)
	}
	return candidates
}

func resolveAnimeFireFallbackDownloadURL(videoAPIURL, failedURL string) (string, error) {
	if !isAnimeFireVideoAPIURL(videoAPIURL) {
		return "", errors.New("source URL is not an AnimeFire video API URL")
	}

	resp, err := api.SafeGet(videoAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch AnimeFire fallback sources: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Warn("Error closing AnimeFire fallback response", "error", err)
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read AnimeFire fallback sources: %w", err)
	}

	candidates, err := selectAnimeFireDownloadCandidates(body, util.GlobalQuality)
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		if candidate != "" && candidate != failedURL {
			return candidate, nil
		}
	}
	return "", errors.New("AnimeFire video API returned no fallback source")
}

type directDownloadFunc func(string, string, *model) error
type fallbackResolveFunc func(string, string) (string, error)

func downloadAnimeFireDirectWithFallback(videoAPIURL, videoURL, path string, m *model) error {
	// lightspeedst.net (AnimeFire CDN) requires Referer: https://animefire.io to authorise
	// token-signed requests. Ensure it is set before the download client sends any request.
	if util.GetGlobalReferer() == "" {
		util.SetGlobalReferer("https://animefire.io")
	}
	return runAnimeFireDirectDownloadWithFallback(
		videoAPIURL,
		videoURL,
		path,
		m,
		downloadDirectHTTP,
		resolveAnimeFireFallbackDownloadURL,
	)
}

func runAnimeFireDirectDownloadWithFallback(videoAPIURL, videoURL, path string, m *model, download directDownloadFunc, resolveFallback fallbackResolveFunc) error {
	err := download(videoURL, path, m)
	if err == nil {
		return nil
	}
	if !isHTTPStatusError(err, http.StatusNotFound) || !isAnimeFireVideoAPIURL(videoAPIURL) {
		return err
	}

	fallbackURL, fallbackErr := resolveFallback(videoAPIURL, videoURL)
	if fallbackErr != nil || fallbackURL == "" {
		util.Debug("AnimeFire fallback source unavailable", "url", videoURL, "error", fallbackErr)
		return err
	}

	util.Debug("AnimeFire source returned 404, retrying fallback source", "failed_url", videoURL, "fallback_url", fallbackURL)
	m.resetProgressReceived()
	if retryErr := download(fallbackURL, path, m); retryErr != nil {
		return fmt.Errorf("%w; AnimeFire fallback failed: %v", err, retryErr)
	}
	return nil
}

func isHTTPStatusError(err error, status int) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status))
}

type batchDownloadFailure struct {
	Episode int
	Err     error
}

type batchDownloadError struct {
	Failures []batchDownloadFailure
}

func (e batchDownloadError) Error() string {
	if len(e.Failures) == 0 {
		return "batch download failed"
	}
	if len(e.Failures) == 1 {
		failure := e.Failures[0]
		return fmt.Sprintf("1 episode failed: episode %d: %v", failure.Episode, failure.Err)
	}

	const maxDetails = 5
	parts := make([]string, 0, min(len(e.Failures), maxDetails))
	for i, failure := range e.Failures {
		if i >= maxDetails {
			break
		}
		parts = append(parts, fmt.Sprintf("episode %d: %v", failure.Episode, failure.Err))
	}
	if len(e.Failures) > maxDetails {
		parts = append(parts, fmt.Sprintf("%d more", len(e.Failures)-maxDetails))
	}
	return fmt.Sprintf("%d episodes failed: %s", len(e.Failures), strings.Join(parts, "; "))
}

func recordBatchDownloadFailure(mu *sync.Mutex, failures *[]batchDownloadFailure, episode int, err error) {
	if err == nil {
		return
	}
	mu.Lock()
	*failures = append(*failures, batchDownloadFailure{Episode: episode, Err: err})
	mu.Unlock()
}

func newBatchDownloadError(failures []batchDownloadFailure) error {
	if len(failures) == 0 {
		return nil
	}
	copied := append([]batchDownloadFailure(nil), failures...)
	sort.Slice(copied, func(i, j int) bool {
		return copied[i].Episode < copied[j].Episode
	})
	return batchDownloadError{Failures: copied}
}

// downloadBloggerDirect downloads a video directly from googlevideo CDN
// using multiple independent surf clients with Chrome TLS impersonation.
// Each thread creates its own TCP+TLS connection, avoiding HTTP/2 multiplexing
// and surf's internal 30s timeout that kills single long-lived connections.
func downloadBloggerDirect(directURL, destPath string, numThreads int, m *model) error {
	start := time.Now()
	util.Debug("downloadBloggerDirect started", "url_len", len(directURL), "threads", numThreads)
	destPath = filepath.Clean(destPath)

	// SSRF protection: validate the URL resolves to a public IP before using
	// the Chrome-impersonation surf client (whose transport we cannot replace).
	if err := api.ValidateExternalURL(directURL); err != nil {
		return fmt.Errorf("SSRF blocked: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use download client that follows redirects (googlevideo uses 302s)
	headClient := newSurfDownloadClient().Std()
	contentLength, err := getContentLength(directURL, headClient)
	if err != nil {
		return fmt.Errorf("failed to get content length: %w", err)
	}
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
	}

	m.setProgressTotal(contentLength)

	chunkSize := contentLength / int64(numThreads)
	var downloadWg sync.WaitGroup
	errChan := make(chan error, numThreads)

	for i := range numThreads {
		from := int64(i) * chunkSize
		to := from + chunkSize - 1
		if i == numThreads-1 {
			to = contentLength - 1
		}
		downloadWg.Add(1)
		go func(from, to int64, part int) {
			defer downloadWg.Done()
			if err := downloadBloggerChunk(directURL, from, to, part, destPath, m); err != nil {
				util.Logger.Error("Blogger download chunk failed", "thread", part, "error", err)
				errChan <- err
			}
		}(from, to, i)
	}

	downloadWg.Wait()
	close(errChan)

	// Check if any chunk failed
	for err := range errChan {
		if err != nil {
			return fmt.Errorf("chunk download failed: %w", err)
		}
	}

	if err := combineParts(destPath, numThreads); err != nil {
		return fmt.Errorf("failed to combine parts: %w", err)
	}

	util.Debug("downloadBloggerDirect completed", "duration", time.Since(start))
	return nil
}

// downloadBloggerChunk downloads a single chunk from googlevideo CDN.
// Creates its own independent surf client (fresh TCP+TLS connection) to avoid
// HTTP/2 multiplexing and surf's per-client timeout.
// Automatically retries with resume on connection drops.
func downloadBloggerChunk(url string, from, to int64, part int, destPath string, m *model) error {
	// SSRF protection: reject URLs that resolve to private/internal IPs.
	if err := api.ValidateExternalURL(url); err != nil {
		return fmt.Errorf("SSRF blocked: %w", err)
	}

	partFilePath, err := safePartPath(destPath, part)
	if err != nil {
		return err
	}

	// Create/truncate the part file
	file, err := os.Create(partFilePath) /* #nosec G304 -- path validated by safePartPath (traversal checked) */
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			util.Logger.Warn("Error closing part file", "error", err)
		}
	}()

	current := from
	maxStaleRetries := 20 // Only count retries where no data was received
	staleRetries := 0

	for attempt := 0; current <= to; attempt++ {
		if staleRetries >= maxStaleRetries {
			return fmt.Errorf("chunk %d: max retries (%d) exceeded at offset %d/%d", part, maxStaleRetries, current, to)
		}
		if attempt > 0 {
			util.Debugf("Blogger chunk %d: resuming at byte %d (attempt %d)", part, current, attempt+1)
			time.Sleep(500 * time.Millisecond) // Brief pause before retry
		}

		beforeRead := current

		// Fresh download client per attempt — follows redirects and has a long timeout
		client := newSurfDownloadClient().Std()

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", current, to))

		resp, err := client.Do(req)
		if err != nil {
			util.Debugf("Blogger chunk %d: request error: %v", part, err)
			continue // Retry
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			util.Debugf("Blogger chunk %d: unexpected status %d", part, resp.StatusCode)
			continue
		}

		// Read data from this connection until it drops
		buf := make([]byte, 64*1024) // 64KB buffer
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := file.Write(buf[:n]); writeErr != nil {
					_ = resp.Body.Close()
					return writeErr
				}
				current += int64(n)
				m.addProgressReceived(int64(n))
			}
			if readErr == io.EOF {
				_ = resp.Body.Close()
				break // Done with this connection
			}
			if readErr != nil {
				_ = resp.Body.Close()
				util.Debugf("Blogger chunk %d: read error after %d bytes: %v", part, current-from, readErr)
				break // Will retry from current offset
			}
		}

		// Check if we got all the data
		if current > to {
			break
		}

		// Only count as stale retry if no progress was made
		if current == beforeRead {
			staleRetries++
		} else {
			staleRetries = 0 // Reset on progress
		}
	}

	return nil
}

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	start := time.Now()
	util.Debug("DownloadVideo started", "url", url)
	destPath = filepath.Clean(destPath)
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}
	chunkSize := int64(0)
	var contentLength int64
	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		return err
	}
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
	}
	chunkSize = contentLength / int64(numThreads)
	var downloadWg sync.WaitGroup
	errChan := make(chan error, numThreads)
	for i := range numThreads {
		from := int64(i) * chunkSize
		to := from + chunkSize - 1
		if i == numThreads-1 {
			to = contentLength - 1
		}
		downloadWg.Add(1)
		go func(from, to int64, part int, httpClient *http.Client) {
			defer downloadWg.Done()
			if err := downloadPart(url, from, to, part, httpClient, destPath, m); err != nil {
				util.Logger.Error("Download part failed", "thread", part, "error", err)
				errChan <- err
			}
		}(from, to, i, httpClient)
	}
	downloadWg.Wait()
	close(errChan)

	// Check if any chunk failed
	for err := range errChan {
		if err != nil {
			return fmt.Errorf("chunk download failed: %w", err)
		}
	}

	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}
	util.Debug("DownloadVideo completed", "url", url, "duration", time.Since(start))
	return nil
}

// downloadWithYtDlp downloads a video using yt-dlp and updates the progress model if provided.
func downloadWithYtDlp(url, path string, m *model) error {
	// Sanitize inputs
	safeURL, err := sanitizeMediaTarget(url)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	safePath, err := sanitizeOutputPath(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(safePath), 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use go-ytdlp library (no external binary required on PATH).
	// 60-minute timeout: movies from SuperFlix/FlixHQ can be 2+ hours of video;
	// a 10-minute timeout was killing yt-dlp mid-download ("signal: killed").
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	if m != nil && util.IsDebug {
		util.Debugf("Preparing yt-dlp engine (first run may take a moment)...")
	}

	// Try to install yt-dlp with timeout and error handling
	_, installErr := ytdlp.Install(ctx, nil)
	if installErr != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", installErr)
	}

	if m != nil && util.IsDebug {
		util.Debugf("Starting yt-dlp download...")
	}

	// Use yt-dlp's native HLS downloader (not ffmpeg) so that obfuscated
	// segment extensions (.js, .html, .jpg) from CDNs are accepted without
	// issues. ffmpeg's allowed_segment_extensions check rejects them.
	dl := ytdlp.New().
		Output(safePath).
		Format("bestvideo+bestaudio/best").
		ConcurrentFragments(24).
		BufferSize("32M").
		FragmentRetries("5").
		Retries("5").
		SocketTimeout(30)

	if util.YtdlpCanImpersonate() {
		dl.Impersonate("chrome")
	}

	// Forward the stored referer/origin so the CDN accepts the request
	if ref := util.GetGlobalReferer(); ref != "" {
		dl.AddHeaders("Referer:" + ref)
		origin := strings.TrimSuffix(ref, "/")
		if u, e := neturl.Parse(origin); e == nil {
			origin = u.Scheme + "://" + u.Host
		}
		dl.AddHeaders("Origin:" + origin)
	}

	// Real-time progress via yt-dlp's native callback.
	// Track per-file totals so video+audio sizes are summed correctly
	// (yt-dlp downloads them as separate files then merges).
	var lastReportedBytes int64
	var lastProgressFile string
	fileTotals := make(map[string]int64)
	if m != nil {
		dl.ProgressFunc(200*time.Millisecond, func(update ytdlp.ProgressUpdate) {
			if update.Status == ytdlp.ProgressStatusPostProcessing ||
				update.Status == ytdlp.ProgressStatusFinished {
				return
			}

			if update.Filename != "" && update.Filename != lastProgressFile {
				lastProgressFile = update.Filename
				lastReportedBytes = 0
			}

			downloaded := int64(update.DownloadedBytes)
			if delta := downloaded - lastReportedBytes; delta > 0 {
				m.addProgressReceived(delta)
				lastReportedBytes = downloaded
			}

			// Sum totals across all files (video + audio) for accurate progress.
			if update.TotalBytes > 0 {
				fname := update.Filename
				if fname == "" {
					fname = "_default"
				}
				fileTotals[fname] = int64(update.TotalBytes)
				var sum int64
				for _, v := range fileTotals {
					sum += v
				}
				m.setProgressTotal(sum)
			} else if update.FragmentCount > 0 && update.FragmentIndex > 0 {
				// HLS streams: TotalBytes is 0 because the size is unknown
				// upfront. Use fragment index/count as a monotonically
				// increasing progress ratio so the bar never goes backward.
				pct := float64(update.FragmentIndex) / float64(update.FragmentCount)
				if pct > 0.99 {
					pct = 0.99 // reserve 100% for actual completion
				}
				m.setProgressPeak(pct)
			}
		})
	}

	// Run with --hls-use-mpegts as raw arg (no typed method available) + retry logic
	var runErr error
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if m != nil && util.IsDebug {
				util.Debugf("Retrying download (attempt %d/%d)...", attempt+1, maxRetries+1)
			}
			time.Sleep(time.Duration(attempt*2) * time.Second)
			lastReportedBytes = 0
			lastProgressFile = ""
		}

		_, runErr = dl.Run(ctx, safeURL, "--hls-use-mpegts")

		if runErr == nil {
			break
		}

		if attempt < maxRetries && isRetryableError(runErr) {
			continue
		} else {
			break
		}
	}

	if runErr != nil {
		// yt-dlp rejects URLs with unusual extensions (.aspx, etc.) as a security
		// measure (CVE-2024-38519). Fall back to direct HTTP download.
		if isUnsafeExtensionError(runErr) {
			util.Debugf("yt-dlp rejected URL extension, falling back to direct HTTP: %s", safeURL)
			return downloadDirectHTTP(safeURL, safePath, m)
		}
		return fmt.Errorf("go-ytdlp download failed: %w", runErr)
	}

	return nil
}

// isUnsafeExtensionError returns true if the error is yt-dlp's unsafe extension check.
func isUnsafeExtensionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unsafe") && strings.Contains(s, "extension") ||
		strings.Contains(s, "unusual") && strings.Contains(s, "extension") ||
		strings.Contains(s, "is unusual and will be skipped")
}
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "temporary") ||
		strings.Contains(errStr, "reset") ||
		strings.Contains(errStr, "refused")
}

// downloadWithNativeHLS downloads HLS streams using native implementation instead of yt-dlp
// This avoids issues with yt-dlp where ffmpeg rejects obfuscated segment extensions (.jpg, .png)
// and yt-dlp's native downloader rejects "live" HLS (no #EXT-X-ENDLIST).
func downloadWithNativeHLS(streamURL, path string, m *model) error {
	// Sanitize inputs
	safeURL, err := sanitizeMediaTarget(streamURL)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	safePath, err := sanitizeOutputPath(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(safePath), 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if m != nil {
		util.Debug("Starting native HLS download", "streamURL", safeURL)
	}

	// Get referer from global storage (set from embed URL in GetFlixHQStreamURL)
	referer := util.GetGlobalReferer()
	if referer == "" {
		referer = extractRefererFromURL(safeURL)
	}

	util.Debug("Native HLS download using referer", "referer", referer, "streamURL", safeURL)

	// Prepare headers with proper referer and origin
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept":     "*/*",
	}

	if referer != "" {
		headers["Referer"] = referer
		headers["Origin"] = strings.TrimSuffix(referer, "/")
	}

	ctx := context.Background()

	// Use a surf-backed HTTP client with Chrome TLS fingerprinting so the CDN
	// does not reject requests from a plain Go client.
	surfClient := util.GetDownloadClient()

	// Real byte-based progress via the HLS callback.
	// The callback now reports (bytesWritten, segmentsWritten, totalSegments).
	// bytesWritten = actual bytes flushed to disk.
	// We use bytesWritten directly for m.received, and dynamically estimate
	// m.totalBytes from the average bytes per written segment.
	err = hls.DownloadToFileWithClient(ctx, surfClient, safeURL, safePath, headers, func(bytesWritten int64, segmentsWritten, totalSegments int) {
		if m == nil || totalSegments <= 0 {
			return
		}

		// Update received with real bytes on disk
		m.setProgressReceived(bytesWritten)

		// Dynamically estimate total file size from average bytes per segment.
		// After only a few segments, only increase the estimate (to avoid the bar
		// going backwards due to a single outlier segment). Once we have enough
		// segments for a reliable average (10+), allow the estimate to shrink too
		// so the progress bar tracks reality instead of sitting at e.g. 39%.
		if segmentsWritten >= 3 {
			avgBytesPerSeg := bytesWritten / int64(segmentsWritten)
			estimatedTotal := avgBytesPerSeg * int64(totalSegments)
			if segmentsWritten >= 10 {
				// Reliable average — use directly
				m.setProgressTotal(estimatedTotal)
			} else if m.shouldGrowProgressTotal(estimatedTotal) {
				// Early estimate — only grow to avoid flicker
				m.setProgressTotal(estimatedTotal)
			}
		}

		// Cap at 98% to prevent showing 100% while write buffer still flushing
		if total := m.progressTotal(); total > 0 && bytesWritten >= total {
			m.setProgressReceived(int64(float64(total) * 0.98))
		}
	})

	if err != nil {
		return fmt.Errorf("native HLS download failed: %w", err)
	}

	// NOTE: Do NOT set m.received = m.totalBytes here.
	// The caller goroutine sets final 100% together with m.done in a single
	// lock, preventing the tick handler from seeing 100% before done=true
	// (which would cause a visual jump in the progress bar).

	return nil
}

// downloadDirectHTTP downloads a video via plain HTTP streaming.
// Used as a last-resort fallback when both native HLS and yt-dlp fail
// (e.g. SharePoint .aspx URLs that serve direct video content).
func downloadDirectHTTP(videoURL, path string, m *model) error {
	client := &http.Client{
		Transport: api.SafeTransport(10 * time.Minute),
	}
	return downloadDirectHTTPWithClient(videoURL, path, m, client)
}

func downloadDirectHTTPWithClient(videoURL, path string, m *model, client *http.Client) error {
	safeURL, err := sanitizeMediaTarget(videoURL)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	safePath, err := sanitizeOutputPath(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(safePath), 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	util.Debug("Starting direct HTTP download", "url", safeURL)

	if client == nil {
		client = &http.Client{
			Transport: api.SafeTransport(10 * time.Minute),
		}
	}
	req, err := http.NewRequest("GET", safeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if ref := util.GetGlobalReferer(); ref != "" {
		req.Header.Set("Referer", ref)
	}

	resp, err := client.Do(req) // #nosec G107 -- URL validated above
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger.Warn("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return scraper.NewDownloadExpiredError("Download", "http", resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	if m != nil {
		m.setProgressTotal(resp.ContentLength)
	}

	// #nosec G304: path validated by sanitizeOutputPath
	out, err := os.Create(safePath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	buf := make([]byte, 256*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				return wErr
			}
			if m != nil {
				m.addProgressReceived(int64(n))
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	return nil
}

// extractRefererFromURL extracts the referer (origin) from a URL
// e.g., https://megacloud.tv/embed-2/abc123?k=v -> https://megacloud.tv/
func extractRefererFromURL(streamURL string) string {
	parsed, err := neturl.Parse(streamURL)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "" && parsed.Host != "" {
		return fmt.Sprintf("%s://%s/", parsed.Scheme, parsed.Host)
	}
	return ""
}

// ExtractVideoSources returns the available video sources for an episode.
func ExtractVideoSources(episodeURL string) ([]struct {
	Quality int
	URL     string
}, error) {
	videoSrc, err := extractVideoURL(episodeURL)
	if err != nil {
		return nil, err
	}
	if strings.Contains(videoSrc, "animefire.io/video/") {
		resp, err := api.SafeGet(videoSrc)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				util.Logger.Warn("Error closing response body", "error", err)
			}
		}()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
		if err != nil {
			return nil, err
		}
		var videoResponse struct {
			Data []struct {
				Src   string `json:"src"`
				Label string `json:"label"`
			}
		}
		if err := json.Unmarshal(body, &videoResponse); err == nil && len(videoResponse.Data) > 0 {
			var sources []struct {
				Quality int
				URL     string
			}
			for _, v := range videoResponse.Data {
				labelDigits := digitsRe.FindString(v.Label)
				q := 0
				if labelDigits != "" {
					q, _ = strconv.Atoi(labelDigits)
				}
				sources = append(sources, struct {
					Quality int
					URL     string
				}{Quality: q, URL: v.Src})
			}
			return sources, nil
		}
	}
	var respStruct struct {
		Data []struct {
			Src   string `json:"src"`
			Label string `json:"label"`
		}
	}
	if err := json.Unmarshal([]byte(videoSrc), &respStruct); err == nil && len(respStruct.Data) > 0 {
		var sources []struct {
			Quality int
			URL     string
		}
		for _, v := range respStruct.Data {
			labelDigits := digitsRe.FindString(v.Label)
			q := 0
			if labelDigits != "" {
				q, _ = strconv.Atoi(labelDigits)
			}
			sources = append(sources, struct {
				Quality int
				URL     string
			}{Quality: q, URL: v.Src})
		}
		return sources, nil
	}
	matches := resolutionMp4Re.FindStringSubmatch(videoSrc)
	if len(matches) > 1 {
		q, _ := strconv.Atoi(matches[1])
		return []struct {
			Quality int
			URL     string
		}{{Quality: q, URL: videoSrc}}, nil
	}
	return []struct {
		Quality int
		URL     string
	}{{Quality: 0, URL: videoSrc}}, nil
}

// getBestQualityURL returns the best available quality for an episode.
// It uses the anime context to route to the correct source-specific stream resolver.
func getBestQualityURL(episode models.Episode, anime *models.Anime) (string, error) {
	animeURL := anime.URL

	// Source-aware routing: use anime.Source to determine the correct resolver.
	// Sources like SuperFlix, FlixHQ, 9Anime, Goyabu, and AnimeDrive use
	// source-specific identifiers or HTTP pages that need scraper-specific
	// resolution instead of the legacy generic page extractor.
	source := anime.Source
	if source != "" && source != "AllAnime" {
		// Use episode.Number; fall back to Num if Number is empty
		epNumber := episode.Number
		if epNumber == "" && episode.Num > 0 {
			epNumber = strconv.Itoa(episode.Num)
		}
		ep := &models.Episode{
			Number:   epNumber,
			Num:      episode.Num,
			URL:      episode.URL,
			DataID:   episode.DataID,
			SeasonID: episode.SeasonID,
		}

		util.Debugf("getBestQualityURL: using source-aware resolver for %s (episode %s)", source, epNumber)
		url, err := api.GetEpisodeStreamURL(ep, anime, util.GlobalQuality)
		if err != nil {
			return "", fmt.Errorf("failed to get stream URL from %s for episode %s: %w", source, epNumber, err)
		}
		if url == "" {
			return "", fmt.Errorf("empty stream URL from %s for episode %s", source, epNumber)
		}
		return url, nil
	}

	// Legacy HTTP page URL path. Keep this after source-aware routing so
	// provider-backed episode pages are not parsed with the generic extractor.
	if strings.HasPrefix(strings.ToLower(episode.URL), "http://") || strings.HasPrefix(strings.ToLower(episode.URL), "https://") {
		sources, err := ExtractVideoSources(episode.URL)
		if err != nil {
			return "", fmt.Errorf("failed to extract video sources: %w", err)
		}
		if len(sources) == 0 {
			return "", fmt.Errorf("no video sources available")
		}
		best := sources[0]
		for _, s := range sources {
			if s.Quality > best.Quality {
				best = s
			}
		}
		return best.URL, nil
	}

	// AllAnime path: animeURL is AllAnime ID/URL, episode.Number is episode string
	isAllAnime := func(u string) bool {
		return strings.Contains(u, "allanime") || (len(u) < 30 && !strings.Contains(u, "http") && len(u) > 0)
	}
	if source == "AllAnime" || isAllAnime(animeURL) {
		allAnime := &models.Anime{URL: animeURL, Source: "AllAnime", Name: "AllAnime"}

		// Use episode.Number; fall back to Num if Number is empty
		epNumber := episode.Number
		if epNumber == "" && episode.Num > 0 {
			epNumber = strconv.Itoa(episode.Num)
		}

		// Build minimal episode with proper number and AllAnime context URL
		ep := &models.Episode{Number: epNumber, URL: animeURL}

		// Retry with backoff — AllAnime rate-limits rapid sequential requests
		const maxRetries = 3
		for attempt := range maxRetries {
			if attempt > 0 {
				delay := time.Duration(attempt) * 800 * time.Millisecond
				util.Debugf("AllAnime retry %d/%d for episode %s (waiting %v)", attempt+1, maxRetries, epNumber, delay)
				time.Sleep(delay)
			}

			if url, err := api.GetEpisodeStreamURLEnhanced(ep, allAnime, util.GlobalQuality); err == nil && url != "" {
				return url, nil
			} else if err != nil {
				util.Debugf("GetEpisodeStreamURLEnhanced failed for episode %s (attempt %d): %v", epNumber, attempt+1, err)
			}

			if url, err := api.GetEpisodeStreamURL(ep, allAnime, util.GlobalQuality); err == nil && url != "" {
				return url, nil
			} else if err != nil {
				util.Debugf("GetEpisodeStreamURL failed for episode %s (attempt %d): %v", epNumber, attempt+1, err)
			}
		}

		return "", fmt.Errorf("failed to resolve AllAnime stream URL for episode %s after %d attempts", epNumber, maxRetries)
	}

	return "", fmt.Errorf("unsupported episode identifier: %s", episode.URL)
}

// ExtractVideoSourcesWithPrompt allows the user to choose video quality.
func ExtractVideoSourcesWithPrompt(episodeURL string) (string, error) {
	sources, err := ExtractVideoSources(episodeURL)
	if err != nil {
		return "", err
	}
	if len(sources) == 0 {
		return "", fmt.Errorf("no video sources available")
	}
	if len(sources) == 1 {
		return sources[0].URL, nil
	}
	var items []string
	for _, s := range sources {
		items = append(items, fmt.Sprintf("%dp", s.Quality))
	}
	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select video quality: "))
	if err != nil {
		return sources[0].URL, nil
	}
	result := items[idx]
	for _, s := range sources {
		if fmt.Sprintf("%dp", s.Quality) == result {
			return s.URL, nil
		}
	}
	return sources[0].URL, nil
}

// HandleBatchDownload performs batch download of episodes.
func HandleBatchDownload(episodes []models.Episode, anime *models.Anime) error {
	animeURL := anime.URL
	start := time.Now()
	util.Debug("HandleBatchDownload started", "animeURL", animeURL, "source", anime.Source)
	startNum, endNum, err := getEpisodeRange()
	if err != nil {
		return fmt.Errorf("invalid episode range: %w", err)
	}
	var (
		m          *model
		p          *tea.Program
		totalBytes int64
		httpClient = &http.Client{
			Transport: api.SafeTransport(10 * time.Second),
		}
		episodesToDownload []int
		resolvedURLs       = make(map[int]string) // cache URLs from pre-flight
		sourceURLs         = make(map[int]string) // original source URLs used for fallback resolution
		estimatedSizes     = make(map[int]int64)
		failuresMu         sync.Mutex
		failures           []batchDownloadFailure
	)

	// Throttle AllAnime pre-flight to avoid rate-limiting
	isAllAnimeURL := anime.Source == "AllAnime" || strings.Contains(animeURL, "allanime")

	// First pass: check which episodes need downloading and calculate total bytes
	for i, episodeNum := 0, startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			err := fmt.Errorf("episode %d not found in selected range", episodeNum)
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}

		// Check if episode already exists
		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}
		if fileExists(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		// Throttle between AllAnime API calls to avoid rate-limiting
		if isAllAnimeURL && i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		i++

		// Resolve URL first; only queue episodes we can actually download
		videoURL, err := getBestQualityURL(episode, anime)
		if err != nil || videoURL == "" {
			if err == nil {
				err = errors.New("empty stream URL")
			}
			util.Logger.Warn("Skipping episode (no stream)", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, fmt.Errorf("failed to resolve stream: %w", err))
			continue
		}
		sourceURLs[episodeNum] = videoURL
		videoURL, err = resolveDownloadURL(videoURL)
		if err != nil || videoURL == "" {
			if err == nil {
				err = errors.New("empty download URL")
			}
			util.Logger.Warn("Skipping episode (failed to resolve download URL)", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}

		// Cache the resolved URL so download goroutines don't re-resolve
		resolvedURLs[episodeNum] = videoURL

		// Episode needs downloading
		episodesToDownload = append(episodesToDownload, episodeNum)
		// Include HLS estimate when Content-Length is not available so progress accumulates realistically.
		// Use 150MB for HLS (typical ~24min anime episode at moderate bitrate) instead of 500MB.
		// The actual total is dynamically adjusted during download from real segment sizes.
		var estimatedSize int64
		if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
			estimatedSize = sz
		} else if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "master.m3u8") || strings.Contains(videoURL, "wixmp.com") || strings.Contains(videoURL, "repackager.wixmp.com") {
			estimatedSize = 150 * 1024 * 1024
		} else {
			estimatedSize = 100 * 1024 * 1024
		}
		estimatedSizes[episodeNum] = estimatedSize
		totalBytes += estimatedSize
	}

	// Check if any episodes need downloading
	if len(episodesToDownload) == 0 {
		if batchErr := newBatchDownloadError(failures); batchErr != nil {
			return batchErr
		}
		// All episodes in range already exist, offer to play one of them
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

	// For 9Anime, prompt subtitle language selection BEFORE starting batch download.
	// The user's choice is stored in GlobalSubtitles and used after each episode download
	// to embed subtitles directly into the video file.
	if util.Is9AnimeSource() {
		util.PromptSubtitleLanguage()
	} else if len(util.GlobalSubtitles) > 0 {
		util.SelectSubtitles()
	}

	if totalBytes > 0 {
		m = &model{
			progress: progress.New(progress.WithDefaultBlend()),
			keys: keyMap{
				quit: key.NewBinding(
					key.WithKeys("ctrl+c"),
					key.WithHelp("ctrl+c", "quit"),
				),
			},
			totalBytes: totalBytes,
			taskTotals: make(map[string]int64),
		}
		for _, epNum := range episodesToDownload {
			m.taskTotals[fmt.Sprintf("episode-%d", epNum)] = estimatedSizes[epNum]
		}
		p = tui.NewProgram(m)
	}
	downloadErrChan := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 4)
		for _, epNum := range episodesToDownload {
			sem <- struct{}{}
			wg.Add(1)
			go func(epNum int) {
				defer func() {
					<-sem
					wg.Done()
				}()
				episode, found := findEpisode(episodes, epNum)
				if !found {
					err := fmt.Errorf("episode %d not found in selected range", epNum)
					util.Warn("Episode not found in batch", "episode", epNum)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
					return
				}
				// Use cached URL from pre-flight; fall back to re-resolving
				videoURL, ok := resolvedURLs[epNum]
				sourceURL := sourceURLs[epNum]
				if !ok || videoURL == "" {
					var err error
					videoURL, err = getBestQualityURL(episode, anime)
					if err != nil || videoURL == "" {
						if err == nil {
							err = errors.New("empty stream URL")
						}
						util.Warn("Skipping episode in batch", "episode", epNum, "error", err)
						recordBatchDownloadFailure(&failuresMu, &failures, epNum, fmt.Errorf("failed to resolve stream: %w", err))
						return
					}
					sourceURL = videoURL
					videoURL, err = resolveDownloadURL(videoURL)
					if err != nil {
						util.Warn("Skipping episode in batch", "episode", epNum, "error", err)
						recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
						return
					}
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					util.Error("Episode path error", "episode", epNum, "error", err)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
					return
				}

				// Double-check if file still doesn't exist (race condition protection)
				if fileExists(episodePath) {
					if p != nil {
						p.Send(statusMsg(fmt.Sprintf("Episode %d already exists, skipping...", epNum)))
					}
					return
				}

				// Keep UI clean in batch mode; don't spam per-episode status or reset aggregate progress
				if p != nil && util.IsDebug {
					p.Send(statusMsg(fmt.Sprintf("Downloading episode %d...", epNum)))
				}
				var progressModel *model
				if m != nil {
					progressModel = m.childProgress(fmt.Sprintf("episode-%d", epNum), estimatedSizes[epNum])
				}
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
				// Also for URLs with extensions yt-dlp rejects (.aspx, .php, etc.).
				if strings.Contains(videoURL, ".m3u8") || hasUnsafeExtension(videoURL) {
					err = downloadWithNativeHLS(videoURL, episodePath, progressModel)
					if err != nil && errors.Is(err, hls.ErrSeparateAudioTracks) {
						util.Debugf("Episode %d: HLS has separate audio tracks, using yt-dlp: %v", epNum, err)
						progressModel.resetProgressReceived()
						err = downloadWithYtDlp(videoURL, episodePath, progressModel)
					} else if err != nil {
						util.Debugf("Episode %d: Native HLS failed, trying direct HTTP: %v", epNum, err)
						progressModel.resetProgressReceived()
						err = downloadDirectHTTP(videoURL, episodePath, progressModel)
					}
					if err != nil {
						util.Debugf("Episode %d: Direct HTTP failed, falling back to yt-dlp: %v", epNum, err)
						progressModel.resetProgressReceived()
						err = downloadWithYtDlp(videoURL, episodePath, progressModel)
					}
				} else if strings.Contains(videoURL, "blogger.com") {
					// Blogger URLs: extract googlevideo CDN URL and download directly
					cdnURL, extractErr := extractBloggerGoogleVideoURL(videoURL)
					if extractErr != nil {
						util.Error("Blogger extraction failed", "episode", epNum, "error", extractErr)
						err = extractErr
					} else {
						err = downloadBloggerDirect(cdnURL, episodePath, 4, progressModel)
					}
				} else if strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") {
					err = downloadWithYtDlp(videoURL, episodePath, progressModel)
				} else if anime.Source == "Animefire.io" || strings.Contains(videoURL, "lightspeedst.net") {
					err = downloadAnimeFireDirectWithFallback(sourceURL, videoURL, episodePath, progressModel)
				} else {
					// Plain MP4 (including blogger proxy) — multi-threaded Range download
					err = DownloadVideo(videoURL, episodePath, 4, progressModel)
				}
				if err != nil {
					util.Error("Failed episode download", "episode", epNum, "error", err)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
				} else {
					// Verify the downloaded file is a reasonable size for a video
					const minEpSize int64 = 10 * 1024 * 1024 // 10 MB
					if stat, statErr := os.Stat(episodePath); statErr == nil && stat.Size() < minEpSize {
						err := fmt.Errorf("downloaded file too small: %.1f MB", float64(stat.Size())/(1024*1024))
						util.Warn("Downloaded file too small, removing partial file",
							"episode", epNum, "size_mb", fmt.Sprintf("%.1f", float64(stat.Size())/(1024*1024)))
						_ = os.Remove(episodePath)
						recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
					} else {
						// Embed selected subtitles into the downloaded video file
						downloadSubtitleFiles(episodePath, func(format string, a ...any) {
							if p != nil {
								msg := fmt.Sprintf(format, a...)
								p.Send(statusMsg(strings.TrimSpace(msg)))
							}
						})
					}
				}
			}(epNum)
		}
		wg.Wait()
		batchErr := newBatchDownloadError(failures)
		// Signal that all downloads are complete
		if m != nil {
			// Send final completion message first
			if p != nil {
				if batchErr != nil {
					p.Send(statusMsg(fmt.Sprintf("Downloads completed with %d failure(s)", len(failures))))
				} else {
					p.Send(statusMsg("All downloads completed!"))
				}
			}
			// Small delay to ensure the user sees the completion message
			time.Sleep(500 * time.Millisecond)

			m.mu.Lock()
			m.err = batchErr
			m.done = true
			m.mu.Unlock()
		}

		downloadErrChan <- batchErr
	}()
	if p != nil {
		restoreConsoleLogs := util.SuppressConsoleLogging()
		_, err := p.Run()
		restoreConsoleLogs()
		if err != nil {
			return fmt.Errorf("progress UI error: %w", err)
		}
	}
	if err := <-downloadErrChan; err != nil {
		return err
	}
	fmt.Println("\nAll episodes downloaded successfully!")
	printBatchDownloadLocation(animeURL, startNum)
	util.Debug("HandleBatchDownload completed", "animeURL", animeURL, "duration", time.Since(start))

	// Ask user which episode from the downloaded range they want to play
	return askAndPlayDownloadedEpisode(episodes, animeURL, startNum, endNum)
}

// HandleBatchDownloadRange performs batch download of episodes using a provided range.
// It mirrors HandleBatchDownload but skips prompting for the range and enables optional
// AniSkip sidecar generation when AllAnime Smart is enabled.
func HandleBatchDownloadRange(episodes []models.Episode, anime *models.Anime, startNum, endNum int) error {
	animeURL := anime.URL
	start := time.Now()
	util.Debug("HandleBatchDownloadRange started", "animeURL", animeURL, "source", anime.Source, "start", startNum, "end", endNum)

	if startNum < 1 || endNum < startNum {
		return fmt.Errorf("invalid episode range: %d-%d", startNum, endNum)
	}

	var (
		m                  *model
		p                  *tea.Program
		totalBytes         int64
		httpClient         = &http.Client{Transport: api.SafeTransport(10 * time.Second)}
		episodesToDownload []int
		resolvedURLs       = make(map[int]string) // cache URLs from pre-flight
		sourceURLs         = make(map[int]string) // original source URLs used for fallback resolution
		estimatedSizes     = make(map[int]int64)
		failuresMu         sync.Mutex
		failures           []batchDownloadFailure
	)

	// Throttle AllAnime pre-flight to avoid rate-limiting
	isAllAnimeURL := anime.Source == "AllAnime" || strings.Contains(animeURL, "allanime")

	// First pass: check which episodes need downloading and calculate total bytes
	for i, episodeNum := 0, startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			err := fmt.Errorf("episode %d not found in selected range", episodeNum)
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}
		if fileExists(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		// Throttle between AllAnime API calls to avoid rate-limiting
		if isAllAnimeURL && i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		i++

		// Resolve URL first; only queue episodes we can actually download
		videoURL, err := getBestQualityURL(episode, anime)
		if err != nil || videoURL == "" {
			if err == nil {
				err = errors.New("empty stream URL")
			}
			util.Logger.Warn("Skipping episode (no stream)", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, fmt.Errorf("failed to resolve stream: %w", err))
			continue
		}
		sourceURLs[episodeNum] = videoURL
		videoURL, err = resolveDownloadURL(videoURL)
		if err != nil || videoURL == "" {
			if err == nil {
				err = errors.New("empty download URL")
			}
			util.Logger.Warn("Skipping episode (failed to resolve download URL)", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}

		// Cache the resolved URL so download goroutines don't re-resolve
		resolvedURLs[episodeNum] = videoURL

		episodesToDownload = append(episodesToDownload, episodeNum)
		var estimatedSize int64
		if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
			estimatedSize = sz
		} else if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "master.m3u8") || strings.Contains(videoURL, "wixmp.com") || strings.Contains(videoURL, "repackager.wixmp.com") {
			estimatedSize = 150 * 1024 * 1024
		} else {
			estimatedSize = 100 * 1024 * 1024
		}
		estimatedSizes[episodeNum] = estimatedSize
		totalBytes += estimatedSize
	}

	if len(episodesToDownload) == 0 {
		if batchErr := newBatchDownloadError(failures); batchErr != nil {
			return batchErr
		}
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

	// For 9Anime, prompt subtitle language selection BEFORE starting batch download.
	// The user's choice is stored in GlobalSubtitles and used after each episode download
	// to embed subtitles directly into the video file.
	if util.Is9AnimeSource() {
		util.PromptSubtitleLanguage()
	} else if len(util.GlobalSubtitles) > 0 {
		util.SelectSubtitles()
	}

	if totalBytes > 0 {
		m = &model{
			progress:   progress.New(progress.WithDefaultBlend()),
			keys:       keyMap{quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
			totalBytes: totalBytes,
			taskTotals: make(map[string]int64),
		}
		for _, epNum := range episodesToDownload {
			m.taskTotals[fmt.Sprintf("episode-%d", epNum)] = estimatedSizes[epNum]
		}
		p = tui.NewProgram(m)
	}

	downloadErrChan := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 4)
		for _, epNum := range episodesToDownload {
			sem <- struct{}{}
			wg.Add(1)
			go func(epNum int) {
				defer func() { <-sem; wg.Done() }()
				episode, found := findEpisode(episodes, epNum)
				if !found {
					err := fmt.Errorf("episode %d not found in selected range", epNum)
					util.Warn("Episode not found in batch", "episode", epNum)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
					return
				}

				// Use cached URL from pre-flight; fall back to re-resolving
				videoURL, ok := resolvedURLs[epNum]
				sourceURL := sourceURLs[epNum]
				if !ok || videoURL == "" {
					var err error
					videoURL, err = getBestQualityURL(episode, anime)
					if err != nil || videoURL == "" {
						if err == nil {
							err = errors.New("empty stream URL")
						}
						util.Warn("Skipping episode in batch", "episode", epNum, "error", err)
						recordBatchDownloadFailure(&failuresMu, &failures, epNum, fmt.Errorf("failed to resolve stream: %w", err))
						return
					}
					sourceURL = videoURL
					videoURL, err = resolveDownloadURL(videoURL)
					if err != nil {
						util.Warn("Skipping episode in batch", "episode", epNum, "error", err)
						recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
						return
					}
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					util.Error("Episode path error", "episode", epNum, "error", err)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, err)
					return
				}

				if fileExists(episodePath) {
					if p != nil {
						p.Send(statusMsg(fmt.Sprintf("Episode %d already exists, skipping...", epNum)))
					}
					return
				}

				if p != nil && util.IsDebug {
					p.Send(statusMsg(fmt.Sprintf("Downloading episode %d...", epNum)))
				}

				var dlErr error
				var progressModel *model
				if m != nil {
					progressModel = m.childProgress(fmt.Sprintf("episode-%d", epNum), estimatedSizes[epNum])
				}
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
				// Also for URLs with extensions yt-dlp rejects (.aspx, .php, etc.).
				if strings.Contains(videoURL, ".m3u8") || hasUnsafeExtension(videoURL) {
					dlErr = downloadWithNativeHLS(videoURL, episodePath, progressModel)
					if dlErr != nil && errors.Is(dlErr, hls.ErrSeparateAudioTracks) {
						util.Debugf("Episode %d: HLS has separate audio tracks, using yt-dlp: %v", epNum, dlErr)
						progressModel.resetProgressReceived()
						dlErr = downloadWithYtDlp(videoURL, episodePath, progressModel)
					} else if dlErr != nil {
						util.Debugf("Episode %d: Native HLS failed, trying direct HTTP: %v", epNum, dlErr)
						progressModel.resetProgressReceived()
						dlErr = downloadDirectHTTP(videoURL, episodePath, progressModel)
					}
					if dlErr != nil {
						util.Debugf("Episode %d: Direct HTTP failed, falling back to yt-dlp: %v", epNum, dlErr)
						progressModel.resetProgressReceived()
						dlErr = downloadWithYtDlp(videoURL, episodePath, progressModel)
					}
				} else if strings.Contains(videoURL, "blogger.com") {
					// Blogger URLs: extract googlevideo CDN URL and download directly
					cdnURL, extractErr := extractBloggerGoogleVideoURL(videoURL)
					if extractErr != nil {
						util.Error("Blogger extraction failed", "episode", epNum, "error", extractErr)
						dlErr = extractErr
					} else {
						dlErr = downloadBloggerDirect(cdnURL, episodePath, 4, progressModel)
					}
				} else if strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") {
					dlErr = downloadWithYtDlp(videoURL, episodePath, progressModel)
				} else if anime.Source == "Animefire.io" || strings.Contains(videoURL, "lightspeedst.net") {
					dlErr = downloadAnimeFireDirectWithFallback(sourceURL, videoURL, episodePath, progressModel)
				} else {
					// Plain MP4 (including blogger proxy) — multi-threaded Range download
					dlErr = DownloadVideo(videoURL, episodePath, 4, progressModel)
				}
				if dlErr != nil {
					util.Error("Failed episode download", "episode", epNum, "error", dlErr)
					recordBatchDownloadFailure(&failuresMu, &failures, epNum, dlErr)
					return
				}

				// Embed selected subtitles into the downloaded video file
				downloadSubtitleFiles(episodePath, func(format string, a ...any) {
					if p != nil {
						msg := fmt.Sprintf(format, a...)
						p.Send(statusMsg(strings.TrimSpace(msg)))
					}
				})

				// Optional: write AniSkip sidecar when AllAnime Smart is enabled
				if util.GlobalDownloadRequest != nil && util.GlobalDownloadRequest.AllAnimeSmart {
					// Basic heuristic for AllAnime
					if anime.Source == "AllAnime" || strings.Contains(strings.ToLower(animeURL), "allanime") {
						_ = api.WriteAniSkipSidecar(episodePath, &episode)
					}
				}
			}(epNum)
		}
		wg.Wait()
		batchErr := newBatchDownloadError(failures)

		if m != nil {
			if p != nil {
				if batchErr != nil {
					p.Send(statusMsg(fmt.Sprintf("Downloads completed with %d failure(s)", len(failures))))
				} else {
					p.Send(statusMsg("All downloads completed!"))
				}
			}
			time.Sleep(500 * time.Millisecond)
			m.mu.Lock()
			m.err = batchErr
			m.done = true
			m.mu.Unlock()
		}
		downloadErrChan <- batchErr
	}()

	if p != nil {
		restoreConsoleLogs := util.SuppressConsoleLogging()
		_, err := p.Run()
		restoreConsoleLogs()
		if err != nil {
			return fmt.Errorf("progress UI error: %w", err)
		}
	}
	if err := <-downloadErrChan; err != nil {
		return err
	}
	fmt.Println("\nAll episodes downloaded successfully!")
	printBatchDownloadLocation(animeURL, startNum)
	util.Debug("HandleBatchDownloadRange completed", "animeURL", animeURL, "duration", time.Since(start))
	// For programmatic range downloads, exit without further prompts
	return ErrUserQuit
}

// getEpisodeRange asks the user for the episode range for download.
func getEpisodeRange() (startNum, endNum int, err error) {
	var startStr, endStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter start episode number").
				Value(&startStr).
				Validate(func(v string) error {
					if _, e := strconv.Atoi(strings.TrimSpace(v)); e != nil {
						return fmt.Errorf("invalid number")
					}
					return nil
				}),
			huh.NewInput().
				Title("Enter end episode number").
				Value(&endStr).
				Validate(func(v string) error {
					if _, e := strconv.Atoi(strings.TrimSpace(v)); e != nil {
						return fmt.Errorf("invalid number")
					}
					return nil
				}),
		),
	)

	if err := tui.RunClean(form.Run); err != nil {
		return 0, 0, err
	}
	startNum, _ = strconv.Atoi(strings.TrimSpace(startStr))
	endNum, _ = strconv.Atoi(strings.TrimSpace(endStr))
	if startNum > endNum {
		return 0, 0, fmt.Errorf("start cannot be greater than end")
	}
	return startNum, endNum, nil
}

// findEpisode returns the episode struct by number.
func findEpisode(episodes []models.Episode, episodeNum int) (models.Episode, bool) {
	for _, ep := range episodes {
		if ep.Num == episodeNum {
			return ep, true
		}
	}
	return models.Episode{}, false
}

// createEpisodePath creates the file path for the downloaded episode
// using Plex/Jellyfin-compatible naming when anime name is available.
// Intelligently organizes:
//   - Movies: <baseDir>/<MovieName>/<MovieName>.mp4 (flat, no season)
//   - TV Shows: <baseDir>/<ShowName>/Season XX/<ShowName> - sXXeXX.mp4
//   - Anime: <baseDir>/<AnimeName>/Season XX/<AnimeName> - sXXeXX.mp4
func createEpisodePath(animeURL string, epNum int) (string, error) {
	// Snapshot all media state atomically — this function is called from
	// concurrent batch-download goroutines.
	snap := snapshotMedia()

	// Route to the correct base directory: movies/ for movies/TV, anime/ for anime
	var baseDir string
	if snap.IsMovieOrTV {
		baseDir = util.DefaultMovieDownloadDir()
	} else {
		baseDir = util.DefaultDownloadDir()
	}

	// Use Plex-compatible naming when anime name is available
	if snap.AnimeName != "" {
		var fullPath string
		// Check if this is a standalone movie (no season/episode hierarchy)
		if snap.MediaType == "movie" {
			// Movies: flat structure  <baseDir>/<MovieName (Year) {ids}>/<MovieName (Year)>.mp4
			fullPath = util.FormatPlexMoviePath(baseDir, snap.AnimeName, "", snap.Meta)
		} else {
			// TV Shows and Anime: season/episode structure
			season, relEp := resolveSeasonForEpisode(snap, epNum)
			fullPath = util.FormatPlexEpisodePath(baseDir, snap.AnimeName, season, relEp, snap.Meta)
		}
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}
		return fullPath, nil
	}

	// Fallback to URL-based directory for backward compatibility
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	safeAnimeName := strings.ReplaceAll(DownloadFolderFormatter(animeURL), " ", "_")
	var fallbackBase string
	if snap.IsMovieOrTV {
		fallbackBase = filepath.Join(userHome, ".local", "goanime", "downloads", "movies")
	} else {
		fallbackBase = filepath.Join(userHome, ".local", "goanime", "downloads", "anime")
	}
	downloadDir := filepath.Join(fallbackBase, safeAnimeName)
	if err := os.MkdirAll(downloadDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(downloadDir, fmt.Sprintf("%d.mp4", epNum)), nil
}

// printBatchDownloadLocation prints the directory where batch-downloaded episodes
// were saved, derived from the episode path builder.
func printBatchDownloadLocation(animeURL string, sampleEpNum int) {
	ep, err := createEpisodePath(animeURL, sampleEpNum)
	if err != nil {
		return
	}
	dir := filepath.Dir(ep)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	util.PrintSavedLocation("Episodes saved at:", absDir)
}

// fileExists verifica se o arquivo existe.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// handleExistingEpisodes handles the case when all episodes in the requested range already exist
func handleExistingEpisodes(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	fmt.Printf("All episodes in range %d-%d already exist!\n\n", startNum, endNum)

	// Collect existing episodes in the range
	var existingEpisodes []models.Episode
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			continue
		}

		if fileExists(episodePath) {
			existingEpisodes = append(existingEpisodes, episode)
		}
	}

	if len(existingEpisodes) == 0 {
		fmt.Println("No downloaded episodes found in the specified range.")
		return nil
	}

	// Create options for the interactive menu
	type menuOption struct {
		Label string
		Value string
	}
	var menuItems []menuOption
	for _, ep := range existingEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		menuItems = append(menuItems, menuOption{Label: title, Value: strconv.Itoa(ep.Num)})
	}

	// Add option to not watch anything
	menuItems = append(menuItems, menuOption{Label: "Don't watch anything", Value: "exit"})

	idx, err := tui.Find(menuItems, func(i int) string {
		return menuItems[i].Label
	}, fuzzyfinder.WithPromptString("Which episode would you like to watch? "))

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

	selectedEpisode := menuItems[idx].Value
	if selectedEpisode == "exit" {
		fmt.Println("No episode selected.")
		return ErrUserQuit
	}

	// Find and play the selected episode
	episodeNum, err := strconv.Atoi(selectedEpisode)
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	// Verify the episode exists in our list
	_, found := findEpisode(existingEpisodes, episodeNum)
	if !found {
		return fmt.Errorf("selected episode not found")
	}

	fmt.Printf("Playing Episode %d...\n", episodeNum)

	// Get the episode path and play it
	episodePath, err := createEpisodePath(animeURL, episodeNum)
	if err != nil {
		return fmt.Errorf("failed to get episode path: %w", err)
	}

	// Play the episode using the existing player logic
	// Note: We use the local file path as the video URL since it's already downloaded
	// anilistID set to 0 since we don't have that context here, updater set to nil
	return playVideo(episodePath, episodes, episodeNum, 0, nil)
}

// askAndPlayDownloadedEpisode asks the user which episode from the downloaded range they want to play
func askAndPlayDownloadedEpisode(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	// Collect downloaded episodes in the range
	var downloadedEpisodes []models.Episode
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			continue
		}

		if fileExists(episodePath) {
			downloadedEpisodes = append(downloadedEpisodes, episode)
		}
	}

	if len(downloadedEpisodes) == 0 {
		fmt.Println("No downloaded episodes found in the specified range.")
		return nil
	}

	// Create options for the interactive menu
	type menuOption struct {
		Label string
		Value string
	}
	var menuItems []menuOption
	for _, ep := range downloadedEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		menuItems = append(menuItems, menuOption{Label: title, Value: strconv.Itoa(ep.Num)})
	}

	// Add option to not watch anything
	menuItems = append(menuItems, menuOption{Label: "Don't watch anything", Value: "exit"})

	idx, err := tui.Find(menuItems, func(i int) string {
		return menuItems[i].Label
	}, fuzzyfinder.WithPromptString("Which episode would you like to watch? "))

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

	selectedEpisode := menuItems[idx].Value
	if selectedEpisode == "exit" {
		fmt.Println("No episode selected.")
		return ErrUserQuit
	}

	// Find and play the selected episode
	episodeNum, err := strconv.Atoi(selectedEpisode)
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	// Verify the episode exists in our list
	_, found := findEpisode(downloadedEpisodes, episodeNum)
	if !found {
		return fmt.Errorf("selected episode not found")
	}

	fmt.Printf("Playing Episode %d...\n", episodeNum)

	// Get the episode path and play it
	episodePath, err := createEpisodePath(animeURL, episodeNum)
	if err != nil {
		return fmt.Errorf("failed to get episode path: %w", err)
	}

	// Play the episode using the existing player logic
	// Note: We use the local file path as the video URL since it's already downloaded
	// anilistID set to 0 since we don't have that context here, updater set to nil
	return playVideo(episodePath, episodes, episodeNum, 0, nil)
}
