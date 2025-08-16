package player

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
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
	resp, err := client.Do(req)
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

	// Determine an estimated size for per-episode tracking (do NOT overwrite m.totalBytes here)
	var epTotal int64
	if m != nil {
		client := &http.Client{Transport: api.SafeTransport(10 * time.Second)}
		if sz, e := getContentLength(safeURL, client); e == nil && sz > 0 {
			epTotal = sz
		} else if strings.Contains(safeURL, ".m3u8") || strings.Contains(safeURL, "master.m3u8") || strings.Contains(safeURL, "wixmp.com") || strings.Contains(safeURL, "repackager.wixmp.com") {
			epTotal = 500 * 1024 * 1024 // 500MB default for HLS-like streams
		} else {
			epTotal = 200 * 1024 * 1024 // 200MB generic fallback
		}
		// Keep stdout clean; log only in debug
		if util.IsDebug {
			util.Logger.Debug("Starting download", "estimate_mb", fmt.Sprintf("%.1f", float64(epTotal)/(1024*1024)))
		}
	}

	// Start a progress goroutine: aggressively poll file size during the download for real-time updates
	done := make(chan struct{})
	if m != nil && epTotal > 0 {
		go func(total int64, outPath string) {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			var lastLocal int64 // last local bytes accounted to the global aggregator

			// Precompute base name to aggregate temp files created by yt-dlp/ffmpeg
			dir := filepath.Dir(outPath)
			base := filepath.Base(outPath)
			prefix := strings.TrimSuffix(base, filepath.Ext(base))

			// helper: measure current local bytes by summing sizes of temp files for this output
			measureLocal := func() int64 {
				var sum int64
				// Check the final file straight away
				if fi, err := os.Stat(outPath); err == nil {
					// If final file exists, that's the authoritative size
					return fi.Size()
				}
				// Otherwise, sum temp files that start with the prefix in the same dir
				entries, err := os.ReadDir(dir)
				if err != nil {
					return 0
				}
				for _, e := range entries {
					name := e.Name()
					if !strings.HasPrefix(name, prefix) {
						continue
					}
					// Skip the intended final filename
					if name == base {
						continue
					}
					if fi, err := os.Stat(filepath.Join(dir, name)); err == nil {
						sum += fi.Size()
					}
				}
				return sum
			}

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					// Measure local progress as bytes for this single episode
					cur := measureLocal()
					// Convert to a conservative percent against estimate, capped until completion
					// But for aggregation we only care about delta bytes
					if cur < 0 {
						cur = 0
					}
					// Cap contribution to avoid runaway when estimate is small
					if float64(cur) > float64(total)*0.98 {
						cur = int64(float64(total) * 0.98)
					}
					// Apply only positive deltas to the global model
					if delta := cur - lastLocal; delta > 0 {
						m.mu.Lock()
						m.received += delta
						m.mu.Unlock()
						lastLocal = cur
					}
				}
			}
		}(epTotal, safePath)
	}

	// Use go-ytdlp library (no external binary required on PATH)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // Increased timeout for slow connections
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

	dl := ytdlp.New().
		Output(safePath)

	// Run the download with HLS-friendly options and retry logic
	var runErr error
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if m != nil && util.IsDebug {
				fmt.Printf("Retrying download (attempt %d/%d)...\n", attempt+1, maxRetries+1)
			}
			time.Sleep(time.Duration(attempt*2) * time.Second) // Progressive backoff
		}

		_, runErr = dl.Run(ctx, safeURL,
			"--downloader", "ffmpeg",
			"--hls-use-mpegts",
			"--fragment-retries", "3",
			"--retries", "3",
			"--socket-timeout", "30")

		if runErr == nil {
			break // Success, exit retry loop
		}

		// Check if this is a retryable error
		if attempt < maxRetries && isRetryableError(runErr) {
			continue
		} else {
			break // Either max retries reached or non-retryable error
		}
	}

	// Stop progress goroutine and finalize remaining delta
	close(done)

	if runErr != nil {
		return fmt.Errorf("go-ytdlp download failed: %w", runErr)
	}

	if m != nil && epTotal > 0 {
		// Apply a small remaining delta so UI reaches the batch total smoothly
		tail := int64(float64(epTotal) * 0.02)
		if tail < 0 {
			tail = 0
		}
		m.mu.Lock()
		m.received += tail
		if m.totalBytes > 0 && m.received > m.totalBytes {
			m.received = m.totalBytes
		}
		m.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
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

// ExtractVideoSources returns the available video sources for an episode.
func ExtractVideoSources(episodeURL string) ([]struct {
	Quality int
	URL     string
}, error) {
	videoSrc, err := extractVideoURL(episodeURL)
	if err != nil {
		return nil, err
	}
	if strings.Contains(videoSrc, "animefire.plus/video/") {
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
				// Use yt-dlp for HLS/DASH playlists and hosters that require it
				if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") {
					err = downloadWithYtDlp(videoURL, episodePath, m)
				} else if strings.Contains(videoURL, "blogger.com") {
					err = downloadWithYtDlp(videoURL, episodePath, m)
				} else {
					err = DownloadVideo(videoURL, episodePath, 4, m)
				}
				if err != nil {
					util.Logger.Error("Failed episode download", "episode", epNum, "error", err)
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
				if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, ".mpd") || strings.Contains(videoURL, "repackager.wixmp.com") || strings.Contains(videoURL, "blogger.com") {
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

// createEpisodePath creates the file path for the downloaded episode.
func createEpisodePath(animeURL string, epNum int) (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	safeAnimeName := strings.ReplaceAll(DownloadFolderFormatter(animeURL), " ", "_")
	downloadDir := filepath.Join(userHome, ".local", "goanime", "downloads", "anime", safeAnimeName)
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
