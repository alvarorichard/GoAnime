package player

import (
	"context"
	"encoding/json"
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
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/lrstanley/go-ytdlp"
	"github.com/manifoldco/promptui"
)

// downloadPart downloads a part of the video file using HTTP Range Requests.
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))
	resp, err := client.Do(req) // #nosec G704
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger.Warn("Error closing response body", "error", err)
		}
	}()
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
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
			m.mu.Lock()
			m.received += int64(n)
			m.mu.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
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
	for i := 0; i < numThreads; i++ {
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

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	start := time.Now()
	if util.IsDebug {
		util.Logger.Debug("DownloadVideo started", "url", url)
	}
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
	for i := 0; i < numThreads; i++ {
		from := int64(i) * chunkSize
		to := from + chunkSize - 1
		if i == numThreads-1 {
			to = contentLength - 1
		}
		downloadWg.Add(1)
		go func(from, to int64, part int, httpClient *http.Client) {
			defer downloadWg.Done()
			err := downloadPart(url, from, to, part, httpClient, destPath, m)
			if err != nil {
				util.Logger.Error("Download part failed", "thread", part, "error", err)
			}
		}(from, to, i, httpClient)
	}
	downloadWg.Wait()
	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}
	if util.IsDebug {
		util.Logger.Debug("DownloadVideo completed", "url", url, "duration", time.Since(start))
	}
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

	// Use go-ytdlp library (no external binary required on PATH)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if m != nil && util.IsDebug {
		fmt.Println("Preparing yt-dlp engine (first run may take a moment)...")
	}

	// Try to install yt-dlp with timeout and error handling
	_, installErr := ytdlp.Install(ctx, nil)
	if installErr != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", installErr)
	}

	if m != nil && util.IsDebug {
		fmt.Println("Starting yt-dlp download...")
	}

	// Use typed API for ALL flags so they are placed before the URL by go-ytdlp.
	// This is critical for --downloader-args to be processed correctly.
	// We need --downloader ffmpeg for "live" HLS streams (no #EXT-X-ENDLIST),
	// and --downloader-args with -allowed_extensions ALL to make ffmpeg accept
	// obfuscated segment extensions (.jpg, .png) from CDNs like AllAnime.
	dl := ytdlp.New().
		Output(safePath).
		Format("bestvideo+bestaudio/best").
		Downloader("ffmpeg").
		DownloaderArgs("ffmpeg_i:-allowed_extensions ALL").
		ConcurrentFragments(4).
		FragmentRetries("5").
		Retries("5").
		SocketTimeout(30).
		Impersonate("chrome")

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
	var lastReportedBytes int64
	var lastProgressFile string
	if m != nil {
		dl.ProgressFunc(200*time.Millisecond, func(update ytdlp.ProgressUpdate) {
			if update.Status == ytdlp.ProgressStatusPostProcessing ||
				update.Status == ytdlp.ProgressStatusFinished {
				return
			}

			m.mu.Lock()
			defer m.mu.Unlock()

			if update.Filename != "" && update.Filename != lastProgressFile {
				lastProgressFile = update.Filename
				lastReportedBytes = 0
			}

			downloaded := int64(update.DownloadedBytes)
			if delta := downloaded - lastReportedBytes; delta > 0 {
				m.received += delta
				lastReportedBytes = downloaded
			}

			if update.TotalBytes > 0 && m.totalBytes < int64(update.TotalBytes) {
				m.totalBytes = int64(update.TotalBytes)
			}
		})
	}

	// Run with --hls-use-mpegts as raw arg (no typed method available) + retry logic
	var runErr error
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if m != nil && util.IsDebug {
				fmt.Printf("Retrying download (attempt %d/%d)...\n", attempt+1, maxRetries+1)
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
		return fmt.Errorf("go-ytdlp download failed: %w", runErr)
	}

	return nil
}

// isRetryableError checks if an error is retryable (network timeouts, connection issues)
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

	if m != nil && util.IsDebug {
		util.Logger.Debug("Starting native HLS download", "streamURL", safeURL)
	}

	// Get referer from global storage (set from embed URL in GetFlixHQStreamURL)
	referer := util.GetGlobalReferer()
	if referer == "" {
		referer = extractRefererFromURL(safeURL)
	}

	if util.IsDebug {
		util.Logger.Debug("Native HLS download using referer", "referer", referer, "streamURL", safeURL)
	}

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

	// Real byte-based progress via the HLS callback.
	// The callback now reports (bytesWritten, segmentsWritten, totalSegments).
	// bytesWritten = actual bytes flushed to disk.
	// We use bytesWritten directly for m.received, and dynamically estimate
	// m.totalBytes from the average bytes per written segment.
	err = hls.DownloadToFile(ctx, safeURL, safePath, headers, func(bytesWritten int64, segmentsWritten, totalSegments int) {
		if m == nil || totalSegments <= 0 {
			return
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		// Update received with real bytes on disk
		m.received = bytesWritten

		// Dynamically estimate total file size from average bytes per segment
		if segmentsWritten >= 3 {
			avgBytesPerSeg := bytesWritten / int64(segmentsWritten)
			estimatedTotal := avgBytesPerSeg * int64(totalSegments)
			// Only increase estimate (never shrink it to prevent bar going backwards)
			if estimatedTotal > m.totalBytes {
				m.totalBytes = estimatedTotal
			}
		}

		// Cap at 98% to prevent showing 100% while write buffer still flushing
		if m.totalBytes > 0 && m.received >= m.totalBytes {
			m.received = int64(float64(m.totalBytes) * 0.98)
		}
	})

	if err != nil {
		return fmt.Errorf("native HLS download failed: %w", err)
	}

	// Set real 100% from actual file size now that download is truly complete
	if m != nil {
		if fi, statErr := os.Stat(safePath); statErr == nil && fi.Size() > 0 {
			m.mu.Lock()
			m.totalBytes = fi.Size()
			m.received = fi.Size()
			m.mu.Unlock()
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
		body, err := io.ReadAll(resp.Body)
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
				labelDigits := regexp.MustCompile(`\d+`).FindString(v.Label)
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
			labelDigits := regexp.MustCompile(`\d+`).FindString(v.Label)
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
	re := regexp.MustCompile(`(\d{3,4})p?\\.mp4`)
	matches := re.FindStringSubmatch(videoSrc)
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
// For AllAnime episodes (non-HTTP identifiers), resolve via enhanced API using episode.Number and animeID.
func getBestQualityURL(episode models.Episode, animeURL string) (string, error) {
	// Non-AllAnime HTTP page URL path
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
	if isAllAnime(animeURL) {
		anime := &models.Anime{URL: animeURL, Source: "AllAnime", Name: "AllAnime"}
		// Build minimal episode with proper number and AllAnime context URL
		ep := &models.Episode{Number: episode.Number, URL: animeURL}
		if url, err := api.GetEpisodeStreamURLEnhanced(ep, anime, util.GlobalQuality); err == nil && url != "" {
			return url, nil
		}
		if url, err := api.GetEpisodeStreamURL(ep, anime, util.GlobalQuality); err == nil && url != "" {
			return url, nil
		}
		return "", fmt.Errorf("failed to resolve AllAnime stream URL")
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
	prompt := promptui.Select{
		Label: "Select video quality",
		Items: items,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return sources[0].URL, nil
	}
	for _, s := range sources {
		if fmt.Sprintf("%dp", s.Quality) == result {
			return s.URL, nil
		}
	}
	return sources[0].URL, nil
}

// HandleBatchDownload performs batch download of episodes.
func HandleBatchDownload(episodes []models.Episode, animeURL string) error {
	start := time.Now()
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownload started", "animeURL", animeURL)
	}
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
	)

	// First pass: check which episodes need downloading and calculate total bytes
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			continue
		}

		// Check if episode already exists
		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			continue
		}
		if fileExists(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		// Resolve URL first; only queue episodes we can actually download
		videoURL, err := getBestQualityURL(episode, animeURL)
		if err != nil || videoURL == "" {
			util.Logger.Warn("Skipping episode (no stream)", "episode", episodeNum, "error", err)
			continue
		}

		// Episode needs downloading
		episodesToDownload = append(episodesToDownload, episodeNum)
		// Include HLS estimate when Content-Length is not available so progress accumulates realistically
		if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
			totalBytes += sz
		} else if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "master.m3u8") || strings.Contains(videoURL, "wixmp.com") || strings.Contains(videoURL, "repackager.wixmp.com") {
			totalBytes += 500 * 1024 * 1024
		} else {
			totalBytes += 200 * 1024 * 1024
		}
	}

	// Check if any episodes need downloading
	if len(episodesToDownload) == 0 {
		// All episodes in range already exist, offer to play one of them
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

	if totalBytes > 0 {
		m = &model{
			progress: progress.New(progress.WithDefaultGradient()),
			keys: keyMap{
				quit: key.NewBinding(
					key.WithKeys("ctrl+c"),
					key.WithHelp("ctrl+c", "quit"),
				),
			},
			totalBytes: totalBytes,
		}
		p = tea.NewProgram(m)
	}
	downloadErrChan := make(chan error)
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
					util.Logger.Warn("Episode not found in batch", "episode", epNum)
					return
				}
				videoURL, err := getBestQualityURL(episode, animeURL)
				if err != nil {
					util.Logger.Warn("Skipping episode in batch", "episode", epNum, "error", err)
					return
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					util.Logger.Error("Episode path error", "episode", epNum, "error", err)
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
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
				if strings.Contains(videoURL, ".m3u8") {
					err = downloadWithNativeHLS(videoURL, episodePath, m)
					if err != nil {
						util.Logger.Warn("Native HLS failed, falling back to yt-dlp", "episode", epNum, "error", err)
						err = downloadWithYtDlp(videoURL, episodePath, m)
					}
				} else if strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") || strings.Contains(videoURL, "blogger.com") {
					err = downloadWithYtDlp(videoURL, episodePath, m)
				} else {
					err = DownloadVideo(videoURL, episodePath, 4, m)
				}
				if err != nil {
					util.Logger.Error("Failed episode download", "episode", epNum, "error", err)
				} else {
					// Verify the downloaded file is a reasonable size for a video
					const minEpSize int64 = 10 * 1024 * 1024 // 10 MB
					if stat, statErr := os.Stat(episodePath); statErr == nil && stat.Size() < minEpSize {
						util.Logger.Warn("Downloaded file too small, removing partial file",
							"episode", epNum, "size_mb", fmt.Sprintf("%.1f", float64(stat.Size())/(1024*1024)))
						_ = os.Remove(episodePath)
					}
				}
			}(epNum)
		}
		wg.Wait()
		// Signal that all downloads are complete
		if m != nil {
			// Send final completion message first
			if p != nil {
				p.Send(statusMsg("All downloads completed!"))
			}
			// Small delay to ensure the user sees the completion message
			time.Sleep(500 * time.Millisecond)

			m.mu.Lock()
			m.done = true
			m.mu.Unlock()
		}

		downloadErrChan <- nil
	}()
	if p != nil {
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("progress UI error: %w", err)
		}
	}
	if err := <-downloadErrChan; err != nil {
		return err
	}
	fmt.Println("\nAll episodes downloaded successfully!")
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownload completed", "animeURL", animeURL, "duration", time.Since(start))
	}

	// Ask user which episode from the downloaded range they want to play
	return askAndPlayDownloadedEpisode(episodes, animeURL, startNum, endNum)
}

// HandleBatchDownloadRange performs batch download of episodes using a provided range.
// It mirrors HandleBatchDownload but skips prompting for the range and enables optional
// AniSkip sidecar generation when AllAnime Smart is enabled.
func HandleBatchDownloadRange(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	start := time.Now()
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownloadRange started", "animeURL", animeURL, "start", startNum, "end", endNum)
	}

	if startNum < 1 || endNum < startNum {
		return fmt.Errorf("invalid episode range: %d-%d", startNum, endNum)
	}

	var (
		m                  *model
		p                  *tea.Program
		totalBytes         int64
		httpClient         = &http.Client{Transport: api.SafeTransport(10 * time.Second)}
		episodesToDownload []int
	)

	// First pass: check which episodes need downloading and calculate total bytes
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			continue
		}
		if fileExists(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		// Resolve URL first; only queue episodes we can actually download
		videoURL, err := getBestQualityURL(episode, animeURL)
		if err != nil || videoURL == "" {
			util.Logger.Warn("Skipping episode (no stream)", "episode", episodeNum, "error", err)
			continue
		}

		episodesToDownload = append(episodesToDownload, episodeNum)
		if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
			totalBytes += sz
		} else if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "master.m3u8") || strings.Contains(videoURL, "wixmp.com") || strings.Contains(videoURL, "repackager.wixmp.com") {
			totalBytes += 500 * 1024 * 1024
		} else {
			totalBytes += 200 * 1024 * 1024
		}
	}

	if len(episodesToDownload) == 0 {
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

	if totalBytes > 0 {
		m = &model{
			progress:   progress.New(progress.WithDefaultGradient()),
			keys:       keyMap{quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
			totalBytes: totalBytes,
		}
		p = tea.NewProgram(m)
	}

	downloadErrChan := make(chan error)
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
					util.Logger.Warn("Episode not found in batch", "episode", epNum)
					return
				}

				videoURL, err := getBestQualityURL(episode, animeURL)
				if err != nil {
					util.Logger.Warn("Skipping episode in batch", "episode", epNum, "error", err)
					return
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					util.Logger.Error("Episode path error", "episode", epNum, "error", err)
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
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
				if strings.Contains(videoURL, ".m3u8") {
					dlErr = downloadWithNativeHLS(videoURL, episodePath, m)
					if dlErr != nil {
						util.Logger.Warn("Native HLS failed, falling back to yt-dlp", "episode", epNum, "error", dlErr)
						dlErr = downloadWithYtDlp(videoURL, episodePath, m)
					}
				} else if strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") || strings.Contains(videoURL, "blogger.com") {
					dlErr = downloadWithYtDlp(videoURL, episodePath, m)
				} else {
					dlErr = DownloadVideo(videoURL, episodePath, 4, m)
				}
				if dlErr != nil {
					util.Logger.Error("Failed episode download", "episode", epNum, "error", dlErr)
					return
				}

				// Optional: write AniSkip sidecar when AllAnime Smart is enabled
				if util.GlobalDownloadRequest != nil && util.GlobalDownloadRequest.AllAnimeSmart {
					// Basic heuristic for AllAnime
					if strings.Contains(strings.ToLower(animeURL), "allanime") || (len(animeURL) < 30 && !strings.Contains(animeURL, "http")) {
						_ = api.WriteAniSkipSidecar(episodePath, &episode)
					}
				}
			}(epNum)
		}
		wg.Wait()

		if m != nil {
			if p != nil {
				p.Send(statusMsg("All downloads completed!"))
			}
			time.Sleep(500 * time.Millisecond)
			m.mu.Lock()
			m.done = true
			m.mu.Unlock()
		}
		downloadErrChan <- nil
	}()

	if p != nil {
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("progress UI error: %w", err)
		}
	}
	if err := <-downloadErrChan; err != nil {
		return err
	}
	fmt.Println("\nAll episodes downloaded successfully!")
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownloadRange completed", "animeURL", animeURL, "duration", time.Since(start))
	}
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

	if err := form.Run(); err != nil {
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
func createEpisodePath(animeURL string, epNum int) (string, error) {
	// Route to the correct base directory: movies/ for movies/TV, anime/ for anime
	var baseDir string
	if lastIsMovieOrTV {
		baseDir = util.DefaultMovieDownloadDir()
	} else {
		baseDir = util.DefaultDownloadDir()
	}

	// Use Plex-compatible naming when anime name is available
	if lastAnimeName != "" {
		season := lastAnimeSeason
		if season < 1 {
			season = 1
		}
		var fullPath string
		if lastIsMovieOrTV && season <= 0 {
			// Movies: <baseDir>/<MovieName>/<MovieName>.mp4 (no season hierarchy)
			safeName := util.SanitizeForFilename(lastAnimeName)
			if safeName == "" {
				safeName = "Unknown"
			}
			fullPath = filepath.Join(baseDir, safeName, safeName+".mp4")
		} else {
			fullPath = util.FormatPlexEpisodePath(baseDir, lastAnimeName, season, epNum)
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
	if lastIsMovieOrTV {
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
	var options []huh.Option[string]
	for _, ep := range existingEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		options = append(options, huh.NewOption(title, strconv.Itoa(ep.Num)))
	}

	// Add option to not watch anything
	options = append(options, huh.NewOption("Don't watch anything", "exit"))

	var selectedEpisode string
	err := huh.NewSelect[string]().
		Title("Which episode would you like to watch?").
		Options(options...).
		Value(&selectedEpisode).
		Run()

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

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
	var options []huh.Option[string]
	for _, ep := range downloadedEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		options = append(options, huh.NewOption(title, strconv.Itoa(ep.Num)))
	}

	// Add option to not watch anything
	options = append(options, huh.NewOption("Don't watch anything", "exit"))

	var selectedEpisode string
	err := huh.NewSelect[string]().
		Title("Which episode would you like to watch?").
		Options(options...).
		Value(&selectedEpisode).
		Run()

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

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
