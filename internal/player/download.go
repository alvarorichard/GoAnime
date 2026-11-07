package player

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
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

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/rangedl"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
	"github.com/alvarorichard/Goanime/v2/internal/ytdlpdl"
	"golang.org/x/term"
)

// Pre-compiled regexes for download quality parsing
var (
	digitsRe               = regexp.MustCompile(`\d+`)
	resolutionMp4Re        = regexp.MustCompile(`(\d{3,4})p?\.mp4`)
	downloadPartRetryDelay = 500 * time.Millisecond
)

// downloadUserAgent mirrors the browser User-Agent that AnimeFire's CDN
// (lightspeedst.net) and other token-protected origins expect. Using the
// default Go transport UA causes some CDNs to return HTTP 401.
const downloadUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

const minDownloadedVideoSize int64 = 10 * 1024 * 1024

// validateDownloadedVideo prevents CDN error pages (a few bytes of "security
// error" text served with a 200) from ever being reported as a completed
// download.
func validateDownloadedVideo(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("download failed: output file was not created: %w", err)
	}
	if stat.Size() < minDownloadedVideoSize {
		return fmt.Errorf("download incomplete: file is only %d bytes (%.1f MB), expected at least %.0f MB",
			stat.Size(), float64(stat.Size())/(1024*1024), float64(minDownloadedVideoSize)/(1024*1024))
	}
	return nil
}

// applyDownloadAuthHeaders sets the Referer / User-Agent headers required by
// origin-protected CDNs (AnimeFire's lightspeedst.net, etc.) onto a
// download request. It prefers the per-source referer stored in
// util.GetGlobalReferer (set by the scraper / api layer) and falls back to
// hardcoded values for known hosts so legacy paths keep working even when the
// referer was never set.
func applyDownloadAuthHeaders(req *http.Request, url string) {
	if req == nil {
		return
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", downloadUserAgent)
	}

	if ref := downloadReferer(url); ref != "" {
		req.Header.Set("Referer", ref)
		origin := strings.TrimSuffix(ref, "/")
		if u, err := neturl.Parse(origin); err == nil && u.Scheme != "" && u.Host != "" {
			origin = u.Scheme + "://" + u.Host
		}
		if req.Header.Get("Origin") == "" {
			req.Header.Set("Origin", origin)
		}
		return
	}

	switch {
	case strings.Contains(url, "lightspeedst.net"), strings.Contains(url, "animefire"):
		req.Header.Set("Referer", "https://animefire.io")
		req.Header.Set("Origin", "https://animefire.io")
	}
}

// downloadPart downloads one byte range into its part file, resuming on
// connection drops. The mechanics live in rangedl; this wires in the progress
// model and the auth headers.
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model) error {
	var sink rangedl.Sink
	if m != nil {
		sink = modelSink{m}
	}
	return rangedl.FetchPart(rangedl.Part{
		URL: url, From: from, To: to, Index: part, Dest: destPath,
	}, client, rangedl.Options{
		Prepare:        func(req *http.Request) { applyDownloadAuthHeaders(req, url) },
		Progress:       sink,
		RetryDelay:     downloadPartRetryDelay,
		ExpiredIsFinal: true,
	})
}

// isBloggerProxyURL returns true if the URL points to the local Blogger proxy.
// The proxy runs on 127.0.0.1 and is safe to access without SSRF protection.
func isBloggerProxyURL(u string) bool {
	return strings.Contains(u, "127.0.0.1") && strings.Contains(u, "blogger_proxy")
}

// isAnimeFireHLS reports whether u is AnimeFire's HLS playlist, which is served
// under an image name.
//
// Its CDN hands out
//
//	https://akumast.net/i/<token>/h.jpg      the multivariant master
//	.../<variant>/p.jpg                      the variant playlists inside it
//
// Both answer with application/vnd.apple.mpegurl and a body starting #EXTM3U.
// The extension is camouflage, not a format.
//
// Missing it here sends
// the URL to the plain-MP4 downloader, which saves the ~200-byte playlist as
// the episode and then deletes it for being too small — the download "failing"
// with no sign that anything was misread.
//
// Matched on the final path segment rather than the host, because the host
// rotates and the naming has not. A real picture never reaches this function:
// it is only ever called on a resolved STREAM url, and artwork travels in a
// different field (models.Media.ImageURL).
func isAnimeFireHLS(lowerURL string) bool {
	path := lowerURL
	if before, _, ok := strings.Cut(path, "?"); ok {
		path = before
	}
	return strings.HasSuffix(path, "/h.jpg") || strings.HasSuffix(path, "/p.jpg")
}

// LooksLikeHLS returns true if the URL appears to be an HLS stream.
// Matches .m3u8 extensions and /hls/ path segments commonly used by CDNs
// that serve HLS playlists without a .m3u8 extension.
func LooksLikeHLS(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "m3u8") ||
		strings.Contains(lower, "/hls/") ||
		isAnimeFireHLS(lower)
}

// hasUnsafeExtension returns true if the URL has a file extension that yt-dlp
// will reject as unusual (e.g. .aspx from SharePoint CDNs).
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
	resp, err := safeGetFn(videoURL)
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
	if err := jsonx.Unmarshal(body, &videoResponse); err != nil {
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

	resp, err := safeGetFn(videoAPIURL)
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
	return pickFallbackCandidate(candidates, failedURL)
}

// pickFallbackCandidate is the first usable source that is not the one that
// just failed. The list is already ordered by preference.
func pickFallbackCandidate(candidates []string, failedURL string) (string, error) {
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

	// Every episode failed for the same reason (one broken source, typically):
	// say it once instead of repeating it per episode.
	same := true
	for _, failure := range e.Failures[1:] {
		if fmt.Sprint(failure.Err) != fmt.Sprint(e.Failures[0].Err) {
			same = false
			break
		}
	}
	if same {
		return fmt.Sprintf("%d episodes failed, all with the same error: %v", len(e.Failures), e.Failures[0].Err)
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
	if err := validateURLFn(directURL); err != nil {
		return fmt.Errorf("SSRF blocked: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use download client that follows redirects (googlevideo uses 302s)
	headClient := newBloggerClientFn()
	contentLength, err := getContentLength(directURL, headClient)
	if err != nil {
		return fmt.Errorf("failed to get content length: %w", err)
	}
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
	}

	m.setProgressTotal(contentLength)

	err = rangedl.Download(contentLength, numThreads, destPath, func(p rangedl.Part) error {
		err := downloadBloggerChunk(directURL, p.From, p.To, p.Index, destPath, m)
		if err != nil {
			util.Logger.Error("Blogger download chunk failed", "thread", p.Index, "error", err)
		}
		return err
	})
	if err != nil {
		return err
	}

	util.Debug("downloadBloggerDirect completed", "duration", time.Since(start))
	return nil
}

// downloadBloggerChunk downloads a single chunk from the googlevideo CDN.
// Each attempt uses its own fresh client (new TCP+TLS connection) to avoid
// HTTP/2 multiplexing and surf's per-client timeout, and resumes on drops.
func downloadBloggerChunk(url string, from, to int64, part int, destPath string, m *model) error {
	// SSRF protection: reject URLs that resolve to private/internal IPs.
	if err := validateURLFn(url); err != nil {
		return fmt.Errorf("SSRF blocked: %w", err)
	}
	var sink rangedl.Sink
	if m != nil {
		sink = modelSink{m}
	}
	return rangedl.FetchPart(rangedl.Part{
		URL: url, From: from, To: to, Index: part, Dest: destPath,
	}, nil, rangedl.Options{
		Progress:   sink,
		NewClient:  newBloggerClientFn,
		Sleep:      chunkRetrySleep,
		RetryDelay: 500 * time.Millisecond,
	})
}

// downloadTransport is the SSRF-guarded transport for media downloads, except
// that it lets a request through to one of GoAnime's own loopback stream
// proxies (util.IsLocalProxyURL) — the guard refuses loopback, and those
// proxies are how a stream that needs rewriting reaches the downloader as a
// plain file.
func downloadTransport(timeout time.Duration) http.RoundTripper {
	return localProxyAware{
		guarded: api.SafeTransport(timeout),
		local:   &http.Transport{ResponseHeaderTimeout: max(timeout, localProxyHeaderTimeout)},
	}
}

// localProxyHeaderTimeout is how long a download waits for a local stream
// proxy's response headers. The StartFlix proxy sends them only once it holds
// the first part of the range — a 2 MiB CDN fetch it bounds itself at 30 s,
// retries included — so a dead origin is a 502 instead of a truncated 206.
// The general 10 s ran out under a batch's load and restarted ranges.
const localProxyHeaderTimeout = 90 * time.Second

type localProxyAware struct{ guarded, local http.RoundTripper }

func (t localProxyAware) RoundTrip(r *http.Request) (*http.Response, error) {
	if util.IsLocalProxyURL(r.URL.String()) {
		return t.local.RoundTrip(r)
	}
	return t.guarded.RoundTrip(r)
}

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	start := time.Now()
	util.Debug("DownloadVideo started", "url", url)
	destPath = filepath.Clean(destPath)
	httpClient := &http.Client{
		Transport: downloadTransport(10 * time.Second),
	}
	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		return err
	}
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
	}
	// A local stream proxy already fetches its upstream parts in parallel
	// (read-ahead); each extra range would be one more reader with its own
	// read-ahead, multiplying the requests the CDN sees until parts arrive
	// slower than anyone waits for them.
	if numThreads > 1 && util.IsLocalProxyURL(url) {
		numThreads = 1
	}
	err = rangedl.Download(contentLength, numThreads, destPath, func(p rangedl.Part) error {
		err := downloadPart(url, p.From, p.To, p.Index, httpClient, destPath, m)
		if err != nil {
			util.Logger.Error("Download part failed", "thread", p.Index, "error", err)
		}
		return err
	})
	if err != nil {
		return err
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
	// 60-minute timeout: movies from StartFlix/FlixHQ can be 2+ hours of video;
	// a 10-minute timeout was killing yt-dlp mid-download ("signal: killed").
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	if m != nil && util.IsDebug {
		util.Debugf("Preparing yt-dlp engine (first run may take a moment)...")
	}
	engine := newYtdlpEngine()
	if installErr := engine.Install(ctx); installErr != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", installErr)
	}

	if m != nil && util.IsDebug {
		util.Debugf("Starting yt-dlp download...")
	}
	spec := ytdlpdl.Spec{
		URL:         safeURL,
		Output:      safePath,
		Impersonate: ytdlpCanImpersonate(),
		// Forward the stored referer/origin so the CDN accepts the request
		Headers: ytdlpdl.Headers(downloadReferer(url)),
	}
	var sink ytdlpdl.Sink
	if m != nil {
		sink = modelSink{m}
	}
	runErr := ytdlpdl.Run(ctx, engine, spec, sink, ytdlpSleep, func(attempt int) {
		if m != nil && util.IsDebug {
			util.Debugf("Retrying download (attempt %d/%d)...", attempt+1, ytdlpdl.MaxRetries+1)
		}
	})
	return ytdlpdl.Outcome(runErr, func() error {
		util.Debugf("yt-dlp rejected URL extension, falling back to direct HTTP: %s", safeURL)
		return downloadDirectHTTPFn(safeURL, safePath, m)
	})
}

// Seams for the yt-dlp download; tests replace them so no binary is installed
// or run.
var (
	newYtdlpEngine       = func() ytdlpdl.Engine { return ytdlpdl.GoEngine{} }
	ytdlpCanImpersonate  = util.YtdlpCanImpersonate
	ytdlpSleep           = time.Sleep
	downloadDirectHTTPFn = downloadDirectHTTP
)

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

	// The referer the stream's resolver gave (per URL in a batch, else global)
	referer := downloadReferer(streamURL)
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
		origin := downloadOrigin(streamURL)
		if origin == "" {
			origin = strings.TrimSuffix(referer, "/")
		}
		headers["Origin"] = origin
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
		Transport: downloadTransport(10 * time.Minute),
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
			Transport: downloadTransport(10 * time.Minute),
		}
	}
	req, err := http.NewRequest("GET", safeURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if ref := downloadReferer(videoURL); ref != "" {
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
			return netx.NewDownloadExpiredError("Download", "http", resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Look at what came back before creating the file: a 200 is not proof of
	// video. This runs as the fallback after a failed native HLS download, so
	// it is often handed the .m3u8 itself, and CDNs serve error pages as 200.
	body := bufio.NewReaderSize(resp.Body, 512)
	if err := refuseNonMedia(resp.Header.Get("Content-Type"), body); err != nil {
		return err
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
		n, readErr := body.Read(buf)
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

// errNotMedia marks a response that is a playlist or a web page rather than
// the video, so the caller moves on to its next way of fetching it.
var errNotMedia = errors.New("response is not a media file")

// refuseNonMedia rejects a playlist or HTML page served in place of the
// video, by its Content-Type or, when the type is missing or generic, by how
// the body starts. Video containers (MP4, MPEG-TS, Matroska) never begin with
// "#EXTM3U" or markup.
func refuseNonMedia(contentType string, body *bufio.Reader) error {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	for _, bad := range []string{"mpegurl", "text/html", "application/xhtml"} {
		if strings.Contains(ct, bad) {
			return fmt.Errorf("%w: served as %s", errNotMedia, contentType)
		}
	}
	head, _ := body.Peek(512)
	lead := strings.ToLower(strings.TrimLeft(string(bytes.TrimPrefix(head, []byte("\xef\xbb\xbf"))), " \t\r\n"))
	for _, marker := range []string{"#extm3u", "<!doctype html", "<html", "<head", "<body"} {
		if strings.HasPrefix(lead, marker) {
			return fmt.Errorf("%w: body starts with %q", errNotMedia, marker)
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
		resp, err := safeGetFn(videoSrc)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				util.Logger.Warn("Error closing response body", "error", err)
			}
		}()
		var videoResponse videoEnvelope
		// Streamed: the response bytes are only used for this decode.
		if err := jsonx.Decode(resp.Body, 5*1024*1024, &videoResponse); err == nil && len(videoResponse.Data) > 0 {
			return sourcesFromEntries(videoResponse.Data), nil
		}
	}
	var respStruct videoEnvelope
	if err := jsonx.Unmarshal([]byte(videoSrc), &respStruct); err == nil && len(respStruct.Data) > 0 {
		return sourcesFromEntries(respStruct.Data), nil
	}
	return []struct {
		Quality int
		URL     string
	}{{Quality: qualityOfURL(videoSrc), URL: videoSrc}}, nil
}

// videoEntry is one quality in a source's video JSON.
type videoEntry struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

// videoEnvelope is the JSON a video endpoint answers with.
type videoEnvelope struct {
	Data []videoEntry
}

// qualityOfLabel is the number in a quality label ("720p" is 720); 0 when the
// label has none ("Auto", "FHD").
func qualityOfLabel(label string) int {
	digits := digitsRe.FindString(label)
	if digits == "" {
		return 0
	}
	q, _ := strconv.Atoi(digits)
	return q
}

// qualityOfURL is the resolution a direct MP4 link spells in its name
// ("..._720p.mp4"), 0 when it spells none.
func qualityOfURL(u string) int {
	matches := resolutionMp4Re.FindStringSubmatch(u)
	if len(matches) > 1 {
		q, _ := strconv.Atoi(matches[1])
		return q
	}
	return 0
}

// sourcesFromEntries lists the entries' links with the quality their labels
// state, in the order given.
func sourcesFromEntries(entries []videoEntry) []struct {
	Quality int
	URL     string
} {
	sources := make([]struct {
		Quality int
		URL     string
	}, 0, len(entries))
	for _, v := range entries {
		sources = append(sources, struct {
			Quality int
			URL     string
		}{Quality: qualityOfLabel(v.Label), URL: v.Src})
	}
	return sources
}

// getBestQualityStream resolves an episode to the best stream available, with
// its playback context (referer, subtitles, ...) in the Stream rather than in
// the player's globals. It uses the anime context to route to the correct
// source-specific stream resolver.
func getBestQualityStream(episode models.Episode, anime *models.Anime) (source.Stream, error) {
	// Source-aware routing: use anime.Source to determine the correct resolver.
	// Sources like StartFlix, FlixHQ, 9Anime, Goyabu, and AnimeDrive use
	// source-specific identifiers or HTTP pages that need scraper-specific
	// resolution instead of the legacy generic page extractor.
	src := anime.Source
	if src != "" {
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

		util.Debugf("getBestQualityStream: using source-aware resolver for %s (episode %s)", src, epNumber)
		stream, err := providers.FetchStream(context.Background(), ep, anime, util.GlobalQuality)
		if err != nil {
			return source.Stream{}, fmt.Errorf("failed to get stream URL from %s for episode %s: %w", src, epNumber, err)
		}
		if stream.URL == "" {
			return source.Stream{}, fmt.Errorf("empty stream URL from %s for episode %s", src, epNumber)
		}
		return stream, nil
	}

	// Legacy HTTP page URL path. Keep this after source-aware routing so
	// provider-backed episode pages are not parsed with the generic extractor.
	if strings.HasPrefix(strings.ToLower(episode.URL), "http://") || strings.HasPrefix(strings.ToLower(episode.URL), "https://") {
		sources, err := ExtractVideoSources(episode.URL)
		if err != nil {
			return source.Stream{}, fmt.Errorf("failed to extract video sources: %w", err)
		}
		if len(sources) == 0 {
			return source.Stream{}, fmt.Errorf("no video sources available")
		}
		best := sources[0]
		for _, s := range sources {
			if s.Quality > best.Quality {
				best = s
			}
		}
		return source.Stream{URL: best.URL}, nil
	}

	return source.Stream{}, fmt.Errorf("unsupported episode identifier: %s", episode.URL)
}

// buildQualityMenu sorts sources highest-quality-first and produces stable,
// unambiguous labels. Sources with Quality == 0 (label digits not parseable)
// render as "Auto"; same-quality duplicates get "(mirror N)" suffixes so the
// fuzzyfinder choices are 1:1 with the returned slice and indexing the
// selection is unambiguous.
func buildQualityMenu(sources []struct {
	Quality int
	URL     string
}) (sortedSources []struct {
	Quality int
	URL     string
}, labels []string) {
	sorted := make([]struct {
		Quality int
		URL     string
	}, len(sources))
	copy(sorted, sources)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Quality > sorted[j].Quality
	})

	items := make([]string, len(sorted))
	seen := make(map[string]int, len(sorted))
	for i, s := range sorted {
		base := "Auto"
		if s.Quality > 0 {
			base = fmt.Sprintf("%dp", s.Quality)
		}
		seen[base]++
		if seen[base] > 1 {
			items[i] = fmt.Sprintf("%s (mirror %d)", base, seen[base])
		} else {
			items[i] = base
		}
	}
	return sorted, items
}

// ExtractVideoSourcesWithPrompt allows the user to choose video quality.
//
// 2026-04-28 fix: the previous implementation had three UX bugs.
//  1. When two sources shared a numeric quality (e.g. mirror A and mirror B
//     both 720p), the post-prompt loop matched on the rendered label and
//     always returned the *first* matching source — the user's actual
//     selection was silently discarded. Indexing directly into the sorted
//     slice fixes this.
//  2. Cancelling the prompt (Esc / Ctrl-C → fuzzyfinder.ErrAbort) silently
//     played whichever source happened to be first. We now propagate
//     ErrBackRequested so callers can route back to the menu.
//  3. Sources with an unparseable label rendered as "0p", and the order
//     matched whatever the upstream returned (often ascending). We now
//     sort descending by quality and substitute a useful fallback label
//     when the resolution is unknown.
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

	sorted, items := buildQualityMenu(sources)
	resolutions := make([]int, len(sorted))
	for i, s := range sorted {
		resolutions[i] = s.Quality
	}

	// Fancy list picker (type-to-filter + arrows/vim). Pure-logic fixes
	// (descending sort, mirror-N disambiguation, index-based selection)
	// live in buildQualityMenu and are pinned by quality_menu_test.go.
	idx, err := tui.Pick(qualityPickItems(items, resolutions), tui.PickOptions{
		Breadcrumb:   "Playback > Quality",
		WindowTitle:  "GoAnime - Quality",
		ItemSingular: "quality",
		ItemPlural:   "qualities",
		Action:       "play",
	})
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return "", ErrBackRequested
		}
		return "", fmt.Errorf("failed to select quality: %w", err)
	}
	if idx < 0 || idx >= len(sorted) {
		return "", fmt.Errorf("invalid quality selection: index %d out of range", idx)
	}
	return sorted[idx].URL, nil
}

// The ways a batch fetches an HLS episode, in the order it tries them; seams
// for tests.
var (
	nativeHLSDownloadFn  = downloadWithNativeHLS
	directHTTPDownloadFn = downloadDirectHTTP
	ytdlpDownloadFn      = downloadWithYtDlp
)

// downloadHLSWithFallbacks fetches one HLS episode for a batch: the native
// downloader first; yt-dlp straight away when the stream needs muxing
// (separate audio) or decryption (SAMPLE-AES) the native one cannot do;
// otherwise direct HTTP, then yt-dlp. Direct HTTP refuses a playlist or web
// page served in place of the video (refuseNonMedia), so a failed native
// download always reaches yt-dlp instead of "succeeding" with the .m3u8 text.
func downloadHLSWithFallbacks(videoURL, path string, m *model, epNum int) error {
	err := nativeHLSDownloadFn(videoURL, path, m)
	if err == nil {
		return nil
	}
	if errors.Is(err, hls.ErrSeparateAudioTracks) || errors.Is(err, hls.ErrUnsupportedEncryption) {
		util.Debugf("Episode %d: native HLS cannot handle this stream, using yt-dlp: %v", epNum, err)
		m.resetProgressReceived()
		return ytdlpDownloadFn(videoURL, path, m)
	}
	util.Debugf("Episode %d: Native HLS failed, trying direct HTTP: %v", epNum, err)
	m.resetProgressReceived()
	if err = directHTTPDownloadFn(videoURL, path, m); err == nil {
		return nil
	}
	util.Debugf("Episode %d: Direct HTTP failed, falling back to yt-dlp: %v", epNum, err)
	m.resetProgressReceived()
	return ytdlpDownloadFn(videoURL, path, m)
}

// HandleBatchDownload asks the user for an episode range, downloads it, and
// offers to play one of the downloaded episodes.
func HandleBatchDownload(episodes []models.Episode, anime *models.Anime) error {
	startNum, endNum, err := getEpisodeRange()
	if err != nil {
		return fmt.Errorf("invalid episode range: %w", err)
	}
	return runBatchDownload(episodes, anime, startNum, endNum, batchOptions{
		finish: func() error { return askAndPlayDownloadedEpisode(episodes, anime.URL, startNum, endNum) },
	})
}

// downloadRangeFrom downloads startNum and on, asking only for the end.
// The episodes screen's D key presets the highlighted episode as the start.
func downloadRangeFrom(startNum int, episodes []models.Episode, anime *models.Anime) error {
	endNum, err := promptRangeEnd(startNum)
	if err != nil {
		return err
	}
	return runBatchDownload(episodes, anime, startNum, endNum, batchOptions{
		finish: func() error { return askAndPlayDownloadedEpisode(episodes, anime.URL, startNum, endNum) },
	})
}

// promptRangeEnd asks for the range's end episode, which must not precede
// its start. Split from getEpisodeRange so the D key presets the start.
func promptRangeEnd(startNum int) (int, error) {
	if fd := os.Stdin.Fd(); fd <= uintptr(math.MaxInt) && !term.IsTerminal(int(fd)) {
		return 0, fmt.Errorf("invalid end episode: stdin is not a terminal")
	}

	var endStr string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Download from episode %d to...", startNum)).
				Value(&endStr).
				Validate(func(v string) error {
					end, e := strconv.Atoi(strings.TrimSpace(v))
					if e != nil {
						return fmt.Errorf("invalid number")
					}
					if end < startNum {
						return fmt.Errorf("end cannot precede episode %d", startNum)
					}
					return nil
				}),
		),
	)

	if err := tui.RunClean(form.WithTheme(tui.FormTheme()).Run); err != nil {
		return 0, err
	}
	endNum, _ := strconv.Atoi(strings.TrimSpace(endStr))
	return endNum, nil
}

// HandleBatchDownloadRange downloads the episodes numbered startNum..endNum.
// A number missing from the list is reported as a failure: the user asked
// for it by number.
func HandleBatchDownloadRange(episodes []models.Episode, anime *models.Anime, startNum, endNum int) error {
	return runBatchDownload(episodes, anime, startNum, endNum, batchOptions{
		// For programmatic range downloads, exit without further prompts.
		finish: func() error { return ErrUserQuit },
	})
}

// HandleDownloadAll downloads every listed episode. It spans the episode
// numbers the list actually has, not 1..len(list): a season whose list starts
// past 1 or skips a number (a dubbed list missing an episode) used to lose
// its last episodes and report the gap as a failed download.
func HandleDownloadAll(episodes []models.Episode, anime *models.Anime) error {
	first, last, ok := episodeNumberSpan(episodes)
	if !ok {
		return errors.New("no numbered episodes to download")
	}
	return runBatchDownload(episodes, anime, first, last, batchOptions{
		skipMissing: true,
		finish:      func() error { return ErrUserQuit },
	})
}

// episodeNumberSpan returns the lowest and highest episode numbers listed.
func episodeNumberSpan(episodes []models.Episode) (first, last int, ok bool) {
	for _, ep := range episodes {
		if ep.Num < 1 {
			continue
		}
		if !ok || ep.Num < first {
			first = ep.Num
		}
		if !ok || ep.Num > last {
			last = ep.Num
		}
		ok = true
	}
	return first, last, ok
}

type batchOptions struct {
	skipMissing bool         // numbers absent from the list are gaps, not failures
	finish      func() error // the result once every episode downloaded
}

// batchEpisode is one episode of a batch, with everything its download needs
// captured right after its stream was resolved. Resolvers hand the referer
// and subtitle tracks over in process-wide globals, and a batch resolves
// every episode before downloading any, so by download time those globals
// hold the last episode's: every episode used to be fetched with that
// episode's referer and muxed with its subtitles.
type batchEpisode struct {
	num       int
	episode   models.Episode
	path      string
	url       string // what is downloaded
	sourceURL string // the stream URL before resolveDownloadURL (AnimeFire fallback)
	referer   string
	origin    string
	subtitles []util.SubtitleInfo
}

// Per-episode stream resolution and subtitle muxing; seams for tests.
var (
	resolveEpisodeStreamFn = getBestQualityStream
	embedSubtitlesFn       = downloadSubtitleFiles
)

// batchStatusf reports the batch's first pass. Resolving a stream can take
// seconds per episode, and the progress bar only appears once every episode
// is resolved; a nine-episode StartFlix season used to sit for a minute and a
// half with nothing on screen. A seam for tests.
var batchStatusf = func(format string, a ...any) { fmt.Printf(format, a...) }

// downloadContext is the request context a download's resolver gave: the
// referer, and the origin when the resolver stated one.
type downloadContext struct {
	referer, origin string
	hasOrigin       bool // origin is the resolver's own answer, empty included
}

// downloadContexts maps a download URL to the context its resolver gave, so
// concurrent batch downloads each send their own instead of the global.
var downloadContexts sync.Map // url -> downloadContext

// rememberDownloadReferer pins referer to url until the returned func is
// called, so a download keeps the referer its resolver gave even when the
// global one changes under it. The origin keeps following the global one.
func rememberDownloadReferer(url, referer string) (forget func()) {
	downloadContexts.Store(url, downloadContext{referer: referer})
	return func() { downloadContexts.Delete(url) }
}

// rememberDownloadContext pins both the referer and the origin to url until
// the returned func is called. An empty origin is pinned too: the resolver
// gave none, so the global one is not stood in for it.
func rememberDownloadContext(url, referer, origin string) (forget func()) {
	downloadContexts.Store(url, downloadContext{referer: referer, origin: origin, hasOrigin: true})
	return func() { downloadContexts.Delete(url) }
}

// downloadReferer is the referer to send when downloading url.
func downloadReferer(url string) string {
	if v, ok := downloadContexts.Load(url); ok {
		return v.(downloadContext).referer
	}
	return util.GetGlobalReferer()
}

// downloadOrigin is the explicit Origin to send when downloading url, "" when
// the resolver stated none.
func downloadOrigin(url string) string {
	if v, ok := downloadContexts.Load(url); ok {
		if c := v.(downloadContext); c.hasOrigin {
			return c.origin
		}
	}
	return util.GetGlobalOrigin()
}

// episodeAlreadyDownloaded reports whether path holds a finished episode.
// A file below the minimum video size is what a failed or interrupted
// download leaves behind; it is removed so the episode is fetched again
// rather than skipped as "already exists" on every later run.
func episodeAlreadyDownloaded(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.Size() >= minDownloadedVideoSize {
		return true
	}
	util.Warn("Removing an incomplete earlier download", "path", path, "bytes", info.Size())
	_ = os.Remove(path)
	return false
}

// resolveBatchEpisode resolves one episode's stream and keeps the referer and
// subtitles it came with. They travel in the Stream, so resolving the next
// episode cannot replace them.
func resolveBatchEpisode(episode models.Episode, anime *models.Anime) (batchEpisode, error) {
	nineAnime := util.Is9AnimeSource()
	if !nineAnime {
		// A batch leaves the player no tracks of its own unless the user picks
		// some (chooseBatchSubtitles); the ones an earlier episode left behind
		// must not be offered for a downloaded file. (9Anime keeps the user's
		// language choice there instead.) This is the one write to the player's
		// state the batch still makes besides the source tag: it goes when the
		// player stops reading the tracks from there.
		util.ClearGlobalSubtitles()
	}
	// The tracking of what is watched afterwards, a downloaded episode
	// included, reads the current source; resolving used to set it as a side
	// effect.
	if anime != nil && anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	be := batchEpisode{num: episode.Num, episode: episode}
	stream, err := resolveEpisodeStreamFn(episode, anime)
	if err != nil || stream.URL == "" {
		if err == nil {
			err = errors.New("empty stream URL")
		}
		return be, fmt.Errorf("failed to resolve stream: %w", err)
	}
	be.sourceURL = stream.URL
	if be.url, err = resolveDownloadURL(stream.URL); err != nil || be.url == "" {
		if err == nil {
			err = errors.New("empty download URL")
		}
		return be, err
	}
	be.referer, be.origin = stream.Referer, stream.Origin
	if !nineAnime {
		be.subtitles = streamSubtitles(stream)
	}
	return be, nil
}

// streamSubtitles is the stream's subtitle tracks as the player holds them.
func streamSubtitles(s source.Stream) []util.SubtitleInfo {
	if len(s.Subtitles) == 0 {
		return nil
	}
	subs := make([]util.SubtitleInfo, 0, len(s.Subtitles))
	for _, sub := range s.Subtitles {
		subs = append(subs, util.SubtitleInfo{URL: sub.URL, Language: sub.Language, Label: sub.Label})
	}
	return subs
}

// chooseBatchSubtitles asks once, for the whole batch, which subtitle tracks
// to embed, offering the first episode that has any as the example. It
// returns the chosen tracks, matched against every episode by label and
// language (keepChosenSubtitles), or nil when nothing should be embedded.
func chooseBatchSubtitles(eps map[int]*batchEpisode, order []int) []util.SubtitleInfo {
	for _, n := range order {
		if subs := eps[n].subtitles; len(subs) > 0 {
			util.SetGlobalSubtitles(subs)
			util.SelectSubtitles()
			if util.GlobalNoSubs {
				return nil
			}
			return util.GetGlobalSubtitles()
		}
	}
	return nil
}

// keepChosenSubtitles keeps the episode's tracks the user chose for the batch.
func keepChosenSubtitles(subs, chosen []util.SubtitleInfo) []util.SubtitleInfo {
	var out []util.SubtitleInfo
	for _, s := range subs {
		for _, c := range chosen {
			if s.Label == c.Label && s.Language == c.Language {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

// downloadBatchEpisode downloads one resolved episode of a batch and embeds
// its own subtitles. A failed download removes whatever it wrote, so the
// next run fetches the episode again instead of skipping a partial file.
func downloadBatchEpisode(be *batchEpisode, anime *models.Anime, progressModel *model, printFn func(string, ...any)) error {
	defer rememberDownloadContext(be.url, be.referer, be.origin)()

	var dlErr error
	videoURL, episodePath, epNum := be.url, be.path, be.num
	// Native HLS first for .m3u8 — handles obfuscated segment extensions
	// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
	// Also for URLs with extensions yt-dlp rejects (.aspx, .php, etc.).
	switch {
	// LooksLikeHLS, not a bare ".m3u8" substring: keying on the extension
	// alone missed playlists served under "/hls/" without one — those fell
	// through to the plain-MP4 downloader.
	case LooksLikeHLS(videoURL) || hasUnsafeExtension(videoURL):
		dlErr = downloadHLSWithFallbacks(videoURL, episodePath, progressModel, epNum)
	case strings.Contains(videoURL, "blogger.com"):
		// Blogger URLs: extract googlevideo CDN URL and download directly
		cdnURL, extractErr := extractBloggerGoogleVideoURL(videoURL)
		if extractErr != nil {
			util.Error("Blogger extraction failed", "episode", epNum, "error", extractErr)
			dlErr = extractErr
		} else {
			dlErr = downloadBloggerDirect(cdnURL, episodePath, 4, progressModel)
		}
	case strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com"):
		dlErr = ytdlpDownloadFn(videoURL, episodePath, progressModel)
	case anime.Source == "Animefire.io" || strings.Contains(videoURL, "lightspeedst.net"):
		dlErr = downloadAnimeFireDirectWithFallback(be.sourceURL, videoURL, episodePath, progressModel)
	default:
		// Plain MP4 (including blogger proxy) — multi-threaded Range download
		dlErr = DownloadVideo(videoURL, episodePath, 4, progressModel)
	}
	if dlErr == nil {
		dlErr = validateDownloadedVideo(episodePath)
	}
	if dlErr != nil {
		_ = os.Remove(episodePath)
		return dlErr
	}

	// Embed this episode's subtitles into the downloaded video file.
	embedSubtitlesFn(episodePath, be.subtitles, printFn)

	// Write the AniSkip sidecar whenever skip times are known. Nothing about
	// the sidecar is source-specific, and WriteAniSkipSidecar is a no-op when
	// there are no skip windows.
	_ = api.WriteAniSkipSidecar(episodePath, &be.episode)
	return nil
}

// runBatchDownload is the batch download behind HandleBatchDownload,
// HandleBatchDownloadRange and HandleDownloadAll: resolve every episode in
// startNum..endNum, then download them four at a time behind a progress bar.
func runBatchDownload(episodes []models.Episode, anime *models.Anime, startNum, endNum int, opts batchOptions) error {
	animeURL := anime.URL
	start := time.Now()
	util.Debug("Batch download started", "animeURL", animeURL, "source", anime.Source, "start", startNum, "end", endNum)
	util.SetSubtitleContext(anime.Name, anime.Source, "", "")

	if startNum < 1 || endNum < startNum {
		return fmt.Errorf("invalid episode range: %d-%d", startNum, endNum)
	}

	var (
		m                  *model
		p                  *tea.Program
		totalBytes         int64
		httpClient         = &http.Client{Transport: downloadTransport(10 * time.Second)}
		episodesToDownload []int
		resolved           = make(map[int]*batchEpisode)
		estimatedSizes     = make(map[int]int64)
		failuresMu         sync.Mutex
		failures           []batchDownloadFailure
	)

	// Throttle AllAnime pre-flight to avoid rate-limiting
	isAllAnimeURL := anime.Source == "AllAnime" || strings.Contains(animeURL, "allanime")

	// First pass: resolve each episode that still needs downloading, with the
	// referer and subtitles that belong to it, and estimate the total size.
	listed := 0
	for n := startNum; n <= endNum; n++ {
		if _, found := findEpisode(episodes, n); found {
			listed++
		}
	}
	checked, reported := 0, false
	for i, episodeNum := 0, startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			if opts.skipMissing {
				continue
			}
			err := fmt.Errorf("episode %d not found in selected range", episodeNum)
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}

		checked++

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}
		if episodeAlreadyDownloaded(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		batchStatusf("\rPreparing episode %d (%d of %d)...", episodeNum, checked, listed)
		reported = true

		// Throttle between AllAnime API calls to avoid rate-limiting
		if isAllAnimeURL && i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		i++

		be, err := resolveBatchEpisode(episode, anime)
		if err != nil {
			util.Logger.Warn("Skipping episode (no stream)", "episode", episodeNum, "error", err)
			recordBatchDownloadFailure(&failuresMu, &failures, episodeNum, err)
			continue
		}
		be.num, be.path = episodeNum, episodePath
		resolved[episodeNum] = &be

		episodesToDownload = append(episodesToDownload, episodeNum)
		// Include an HLS estimate when Content-Length is not available so
		// progress accumulates realistically; the real total is adjusted
		// during download from actual segment sizes.
		var estimatedSize int64
		if sz, err := getContentLength(be.url, httpClient); err == nil && sz > 0 {
			estimatedSize = sz
		} else if strings.Contains(be.url, ".m3u8") || strings.Contains(be.url, "wixmp.com") {
			estimatedSize = 150 * 1024 * 1024
		} else {
			estimatedSize = 100 * 1024 * 1024
		}
		estimatedSizes[episodeNum] = estimatedSize
		totalBytes += estimatedSize
	}

	if reported {
		batchStatusf("\n")
	}

	if len(episodesToDownload) == 0 {
		if batchErr := newBatchDownloadError(failures); batchErr != nil {
			return batchErr
		}
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

	// Subtitles: one choice for the whole batch, applied to each episode's
	// own tracks. For 9Anime the language prompt below stores the choice in
	// the global, which every episode then uses.
	if util.Is9AnimeSource() {
		util.PromptSubtitleLanguage()
		for _, be := range resolved {
			be.subtitles = util.GetGlobalSubtitles()
		}
	} else if chosen := chooseBatchSubtitles(resolved, episodesToDownload); chosen != nil {
		for _, be := range resolved {
			be.subtitles = keepChosenSubtitles(be.subtitles, chosen)
		}
	} else {
		for _, be := range resolved {
			be.subtitles = nil
		}
	}

	if totalBytes > 0 {
		// Fail fast without a TTY: the progress program blocks on Windows
		// console input headless instead of erroring like Linux.
		if !tui.HasTerminal() {
			return fmt.Errorf("no terminal: progress UI requires a TTY")
		}
		m = newDownloadModel(anime.Name, fmt.Sprintf("%d episodes", len(episodesToDownload)), totalBytes)
		m.taskTotals = make(map[string]int64)
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
			go func(be *batchEpisode) {
				defer func() { <-sem; wg.Done() }()
				if p != nil && util.IsDebug {
					p.Send(statusMsg(fmt.Sprintf("Downloading episode %d...", be.num)))
				}
				var progressModel *model
				if m != nil {
					progressModel = m.childProgress(fmt.Sprintf("episode-%d", be.num), estimatedSizes[be.num])
				}
				err := downloadBatchEpisode(be, anime, progressModel, func(format string, a ...any) {
					if p != nil {
						p.Send(statusMsg(strings.TrimSpace(fmt.Sprintf(format, a...))))
					}
				})
				m.endTask(fmt.Sprintf("episode-%d", be.num), err == nil)
				if err != nil {
					util.Error("Failed episode download", "episode", be.num, "error", err)
					recordBatchDownloadFailure(&failuresMu, &failures, be.num, err)
				}
			}(resolved[epNum])
		}
		wg.Wait()
		batchErr := newBatchDownloadError(failures)

		if m != nil {
			if p != nil {
				if batchErr != nil {
					p.Send(statusMsg(fmt.Sprintf("Downloads completed with %d failure(s)", len(failures))))
				} else {
					p.Send(statusMsg("All downloads complete"))
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
	if p == nil {
		// Without a progress screen nothing has said so yet.
		fmt.Println("\nAll episodes downloaded successfully!")
	}
	printBatchDownloadLocation(animeURL, startNum)
	util.Debug("Batch download completed", "animeURL", animeURL, "duration", time.Since(start))
	return opts.finish()
}

// getEpisodeRange asks the user for the episode range for download.
func getEpisodeRange() (startNum, endNum int, err error) {
	// Bail out fast when stdin is not a TTY (CI, piped input, headless tests).
	// Otherwise Bubble Tea's input reader blocks on syscall.ReadConsole on
	// Windows or just hangs on a closed stdin elsewhere.
	if fd := os.Stdin.Fd(); fd <= uintptr(math.MaxInt) && !term.IsTerminal(int(fd)) {
		return 0, 0, fmt.Errorf("invalid episode range: stdin is not a terminal")
	}

	var startStr, endStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter start episode number").
				Value(&startStr).
				Validate(validEpisodeNumber),
			huh.NewInput().
				Title("Enter end episode number").
				Value(&endStr).
				Validate(validEpisodeNumber),
		),
	)

	if err := tui.RunClean(form.WithTheme(tui.FormTheme()).Run); err != nil {
		return 0, 0, err
	}
	return parseEpisodeRange(startStr, endStr)
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
		if err := os.MkdirAll(dir, 0o700); err != nil {
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
	if err := os.MkdirAll(downloadDir, 0o700); err != nil {
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
// and offers to play one of them.
func handleExistingEpisodes(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	fmt.Printf("All episodes in range %d-%d already exist!\n\n", startNum, endNum)
	return watchDownloadedRange(episodes, animeURL, startNum, endNum, "Download > Watch")
}

// askAndPlayDownloadedEpisode asks the user which episode from the downloaded range they want to play
func askAndPlayDownloadedEpisode(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	return watchDownloadedRange(episodes, animeURL, startNum, endNum, "Download > Watch Range")
}

// Seams for the "watch what you downloaded" menu.
var (
	episodePathFn = createEpisodePath
	pathExistsFn  = fileExists
	pickLabelsFn  = tui.PickLabels
	playLocalFn   = func(path string, episodes []models.Episode, num int) error {
		return playVideo(path, episodes, num, 0, 0, nil)
	}
	watchExitValue = "exit"
)

// downloadedEpisodesInRange lists the episodes of [startNum, endNum] that are in
// the list and already on disk, in number order.
func downloadedEpisodesInRange(episodes []models.Episode, animeURL string, startNum, endNum int) []models.Episode {
	var found []models.Episode
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, ok := findEpisode(episodes, episodeNum)
		if !ok {
			continue
		}
		episodePath, err := episodePathFn(animeURL, episodeNum)
		if err != nil {
			continue
		}
		if pathExistsFn(episodePath) {
			found = append(found, episode)
		}
	}
	return found
}

// watchMenuEntry is one row of the downloaded-episodes menu.
type watchMenuEntry struct{ label, value string }

// watchMenuEntries labels each episode ("Episode 3: English title", the
// Romaji one when there is no English one) and closes the list with the exit
// row.
func watchMenuEntries(episodes []models.Episode) []watchMenuEntry {
	entries := make([]watchMenuEntry, 0, len(episodes)+1)
	for _, ep := range episodes {
		label := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			label = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			label = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		entries = append(entries, watchMenuEntry{label: label, value: strconv.Itoa(ep.Num)})
	}
	return append(entries, watchMenuEntry{label: "Don't watch anything", value: watchExitValue})
}

// watchPickError turns the picker's failure into what the caller returns:
// backing out or cancelling leaves like quitting does.
func watchPickError(err error) error {
	if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
		return ErrUserQuit
	}
	return fmt.Errorf("episode selection error: %w", err)
}

// watchDownloadedRange offers the episodes of the range that are on disk and
// plays the one chosen.
func watchDownloadedRange(episodes []models.Episode, animeURL string, startNum, endNum int, breadcrumb string) error {
	downloaded := downloadedEpisodesInRange(episodes, animeURL, startNum, endNum)
	if len(downloaded) == 0 {
		fmt.Println("No downloaded episodes found in the specified range.")
		return nil
	}

	entries := watchMenuEntries(downloaded)
	labels := make([]string, len(entries))
	for i, e := range entries {
		labels[i] = e.label
	}
	idx, err := pickLabelsFn(labels, tui.PickOptions{
		Breadcrumb:   breadcrumb,
		WindowTitle:  "GoAnime - Watch Downloaded",
		ItemSingular: "episode",
		ItemPlural:   "episodes",
	})
	if err != nil {
		return watchPickError(err)
	}
	if idx < 0 || idx >= len(entries) {
		return fmt.Errorf("selected episode not found")
	}

	selected := entries[idx].value
	if selected == watchExitValue {
		fmt.Println("No episode selected.")
		return ErrUserQuit
	}
	episodeNum, err := strconv.Atoi(selected)
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}
	if _, ok := findEpisode(downloaded, episodeNum); !ok {
		return fmt.Errorf("selected episode not found")
	}

	fmt.Printf("Playing Episode %d...\n", episodeNum)
	episodePath, err := episodePathFn(animeURL, episodeNum)
	if err != nil {
		return fmt.Errorf("failed to get episode path: %w", err)
	}
	// The file is already local; the IDs are 0 because there is no tracking
	// context here, and no Discord updater.
	return playLocalFn(episodePath, episodes, episodeNum)
}

// parseEpisodeRange validates the two numbers typed into the range form.
func parseEpisodeRange(startStr, endStr string) (start, end int, err error) {
	start, err = strconv.Atoi(strings.TrimSpace(startStr))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number")
	}
	end, err = strconv.Atoi(strings.TrimSpace(endStr))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number")
	}
	if start > end {
		return 0, 0, fmt.Errorf("start cannot be greater than end")
	}
	return start, end, nil
}

// validEpisodeNumber is the form field's check.
func validEpisodeNumber(v string) error {
	if _, err := strconv.Atoi(strings.TrimSpace(v)); err != nil {
		return fmt.Errorf("invalid number")
	}
	return nil
}

// safeGetFn is the SSRF-guarded GET every page and video-API fetch in this
// package goes through; tests replace it, because the guard refuses the
// loopback servers they stand up.
var safeGetFn = api.SafeGet

// Seams for the Blogger/googlevideo parallel download: tests substitute a
// plain client and a no-op guard, because the real ones impersonate Chrome and
// refuse the loopback servers a test starts.
var (
	// validateURLFn is the SSRF guard applied before any request.
	validateURLFn = api.ValidateExternalURL
	// newBloggerClientFn builds a fresh client for each attempt: independent
	// connections avoid HTTP/2 multiplexing and the per-client timeout.
	newBloggerClientFn = func() *http.Client { return newSurfDownloadClient().Std() }
	// chunkRetrySleep is the pause before a chunk resumes.
	chunkRetrySleep = time.Sleep
)
