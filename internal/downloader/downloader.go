package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/go-ytdlp"
)

// DownloadConfig holds configuration for download operations
type DownloadConfig struct {
	AnimeURL   string
	OutputDir  string
	NumThreads int
	Concurrent int // Number of concurrent episode downloads
}

// EpisodeDownloader handles episode download operations
type EpisodeDownloader struct {
	config   DownloadConfig
	episodes []models.Episode
	anime    *models.Anime // Store anime data for enhanced API calls
}

// NewEpisodeDownloader creates a new episode downloader
func NewEpisodeDownloader(episodes []models.Episode, animeURL string) *EpisodeDownloader {
	return NewEpisodeDownloaderWithAnime(episodes, animeURL, nil)
}

// NewEpisodeDownloaderWithAnime creates a new episode downloader with anime data for enhanced API support
func NewEpisodeDownloaderWithAnime(episodes []models.Episode, animeURL string, anime *models.Anime) *EpisodeDownloader {
	userHome, _ := os.UserHomeDir()
	safeAnimeName := strings.ReplaceAll(player.DownloadFolderFormatter(animeURL), " ", "_")
	outputDir := filepath.Join(userHome, ".local", "goanime", "downloads", "anime", safeAnimeName)

	return &EpisodeDownloader{
		config: DownloadConfig{
			AnimeURL:   animeURL,
			OutputDir:  outputDir,
			NumThreads: 4,
			Concurrent: 3, // Download max 3 episodes concurrently
		},
		episodes: episodes,
		anime:    anime,
	}
}

// DownloadSingleEpisode downloads a specific episode by number
func (d *EpisodeDownloader) DownloadSingleEpisode(episodeNum int) error {
	episode, found := d.findEpisodeByNumber(episodeNum)
	if !found {
		return fmt.Errorf("episode %d not found", episodeNum)
	}

	// Create output directory
	if err := os.MkdirAll(d.config.OutputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", episodeNum))

	// Check if episode already exists
	if d.fileExists(episodePath) {
		fmt.Printf("Episode %d already exists at: %s\n", episodeNum, episodePath)
		return d.promptPlayExisting(episodeNum, episodePath)
	}

	// Get video URL using enhanced method if possible, fallback to regular method
	videoURL, err := d.getBestQualityURL(episode.URL)
	if err != nil {
		return fmt.Errorf("failed to get video URL: %w", err)
	}

	// Download with progress
	return d.downloadWithProgress(videoURL, episodePath, episodeNum)
}

// DownloadEpisodeRange downloads a range of episodes
func (d *EpisodeDownloader) DownloadEpisodeRange(startEp, endEp int) error {
	if startEp > endEp {
		return fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
	}

	// Create output directory
	if err := os.MkdirAll(d.config.OutputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect episodes to download
	var episodesToDownload []int
	var existingEpisodes []int
	for epNum := startEp; epNum <= endEp; epNum++ {
		_, found := d.findEpisodeByNumber(epNum)
		if !found {
			util.Warnf("Episode %d not found, skipping", epNum)
			continue
		}
		episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", epNum))
		if d.fileExists(episodePath) {
			existingEpisodes = append(existingEpisodes, epNum)
		} else {
			episodesToDownload = append(episodesToDownload, epNum)
		}
	}
	// Handle case where all episodes already exist
	if len(episodesToDownload) == 0 {
		fmt.Printf("All episodes in range %d-%d already exist!\n", startEp, endEp)
		return d.promptPlayExistingRangeHuh(existingEpisodes)
	}
	fmt.Printf("Found %d episode(s) to download (episodes %d-%d)\n",
		len(episodesToDownload), startEp, endEp)
	// Download episodes concurrently with progress UI
	return d.downloadConcurrentWithProgress(episodesToDownload)
}

// downloadConcurrentWithProgress downloads multiple episodes with proper Bubble Tea progress UI
func (d *EpisodeDownloader) downloadConcurrentWithProgress(episodeNums []int) error {
	if len(episodeNums) == 0 {
		return nil
	}

	// Create progress model for overall progress
	m := &progressModel{
		progress: progress.New(progress.WithDefaultGradient()),
	}

	// Calculate total bytes for all episodes
	var totalBytes int64
	episodeInfos := make(map[int]struct {
		videoURL string
		path     string
		size     int64
	})

	fmt.Println("Calculating download sizes...")
	for _, epNum := range episodeNums {
		episode, found := d.findEpisodeByNumber(epNum)
		if !found {
			continue
		}

		videoURL, err := d.getBestQualityURL(episode.URL)
		if err != nil {
			util.Warnf("Failed to get video URL for episode %d: %v", epNum, err)
			continue
		}

		episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", epNum))

		// Get content length
		size, err := d.getContentLength(videoURL)
		if err != nil {
			util.Warnf("Failed to get content length for episode %d: %v", epNum, err)
			size = 100 * 1024 * 1024 // Default to 100MB estimate
		}

		episodeInfos[epNum] = struct {
			videoURL string
			path     string
			size     int64
		}{videoURL, episodePath, size}

		totalBytes += size
	}

	m.totalBytes = totalBytes
	p := tea.NewProgram(m)

	// Start downloads with progress tracking
	downloadComplete := make(chan error, 1)
	go func() {
		defer func() {
			m.mu.Lock()
			m.done = true
			m.mu.Unlock()
			time.Sleep(500 * time.Millisecond)
			p.Send(statusMsg("All downloads completed!"))
			time.Sleep(200 * time.Millisecond)
			p.Quit()
		}()

		err := d.downloadMultipleWithProgress(episodeNums, episodeInfos, m, p)
		downloadComplete <- err
	}()

	// Run progress bar
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("progress display error: %w", err)
	}

	// Wait for download completion
	if err := <-downloadComplete; err != nil {
		return err
	}

	fmt.Printf("\nAll %d episodes downloaded successfully!\n", len(episodeNums))
	return d.promptPlayDownloadedRangeHuh(episodeNums)
}

// downloadMultipleWithProgress performs concurrent downloads with progress updates
func (d *EpisodeDownloader) downloadMultipleWithProgress(episodeNums []int, episodeInfos map[int]struct {
	videoURL string
	path     string
	size     int64
}, progressModel *progressModel, program *tea.Program) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, d.config.Concurrent) // Limit concurrent downloads
	errChan := make(chan error, len(episodeNums))

	// Shared progress tracking
	var totalReceived int64
	var mu sync.Mutex

	for _, epNum := range episodeNums {
		info, exists := episodeInfos[epNum]
		if !exists {
			continue
		}

		wg.Add(1)
		go func(episodeNum int, info struct {
			videoURL string
			path     string
			size     int64
		}) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			program.Send(statusMsg(fmt.Sprintf("Downloading episode %d...", episodeNum)))

			// Create a simple download progress tracker
			episodeReceived := int64(0)

			err := d.downloadEpisodeWithSharedProgress(info.videoURL, info.path, &episodeReceived, &totalReceived, &mu, progressModel, program)
			if err != nil {
				errChan <- fmt.Errorf("episode %d: download failed: %w", episodeNum, err)
				return
			}

			program.Send(statusMsg(fmt.Sprintf("Episode %d completed!", episodeNum)))
		}(epNum, info)
	}

	// Wait for all downloads to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		fmt.Printf("Some downloads failed:\n")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("%d download(s) failed", len(errors))
	}

	return nil
}

// downloadEpisodeWithSharedProgress downloads an episode while updating shared progress
func (d *EpisodeDownloader) downloadEpisodeWithSharedProgress(videoURL, destPath string, episodeReceived, totalReceived *int64, mu *sync.Mutex, progressModel *progressModel, program *tea.Program) error {
	if strings.Contains(videoURL, "blogger.com") {
		return d.downloadWithYtDlp(videoURL, destPath)
	}

	// Create HTTP client with longer timeout for video downloads
	client := &http.Client{
		Transport: api.SafeTransport(10 * time.Minute), // Much longer transport timeout
		Timeout:   0,                                   // No overall timeout - let it download completely
	}

	// Get the file
	resp, err := client.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create destination file
	safeDest, err := d.sanitizeDestPath(destPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}
	// #nosec G304: dest path validated by sanitizeDestPath to remain within configured OutputDir
	out, err := os.Create(safeDest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			util.Warnf("Failed to close output file: %v", closeErr)
		}
	}()

	// Copy with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Write to file
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress tracking
			mu.Lock()
			*episodeReceived += int64(n)
			*totalReceived += int64(n)

			// Update the progress model
			progressModel.mu.Lock()
			progressModel.received = *totalReceived
			progressModel.mu.Unlock()

			// Send progress update
			program.Send(progressMsg{
				received:   *totalReceived,
				totalBytes: progressModel.totalBytes,
			})
			mu.Unlock()
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read from response: %w", err)
		}
	}

	return nil
}

// Helper methods

func (d *EpisodeDownloader) findEpisodeByNumber(num int) (models.Episode, bool) {
	for _, ep := range d.episodes {
		if ep.Num == num {
			return ep, true
		}
	}
	return models.Episode{}, false
}

func (d *EpisodeDownloader) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// sanitizeDestPath ensures the destination path stays within the configured OutputDir
func (d *EpisodeDownloader) sanitizeDestPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty destination path")
	}
	cleaned := filepath.Clean(p)
	outDir := filepath.Clean(d.config.OutputDir)
	absDir, err := filepath.Abs(outDir)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(cleaned)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absDir, absFile)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("destination escapes output directory: %s", cleaned)
	}
	return absFile, nil
}

func (d *EpisodeDownloader) getBestQualityURL(episodeURL string) (string, error) {
	// Use existing player functionality to get video URL
	videoURL, err := player.GetVideoURLForEpisode(episodeURL)
	if err != nil {
		return "", err
	}
	return videoURL, nil
}

func (d *EpisodeDownloader) getContentLength(url string) (int64, error) {
	// Check if this is an AllAnime URL that might not have Content-Length header
	// Based on ani-cli patterns
	isAllAnimeURL := strings.Contains(url, "sharepoint.com") ||
		strings.Contains(url, "wixmp.com") ||
		strings.Contains(url, "repackager.wixmp.com") ||
		strings.Contains(url, "master.m3u8") ||
		strings.Contains(url, ".m3u8") ||
		strings.Contains(url, "allanime.pro") ||
		strings.Contains(url, "blogger.com")

	// For streaming URLs that we know won't have Content-Length, return estimate immediately
	if strings.Contains(url, ".m3u8") || strings.Contains(url, "master.m3u8") {
		fmt.Println("HLS stream detected, using estimated size")
		return 400 * 1024 * 1024, nil // 400MB estimate for HLS streams
	}

	// Simple HTTP HEAD request to get content length
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		if isAllAnimeURL {
			fmt.Printf("HEAD request failed for AllAnime URL, using estimate: %v\n", err)
			return 300 * 1024 * 1024, nil // 300MB default for AllAnime
		}
		return 0, err
	}

	// Add referer for AllAnime URLs (like ani-cli does)
	if isAllAnimeURL {
		req.Header.Set("Referer", "https://allmanga.to")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		if isAllAnimeURL {
			fmt.Printf("HEAD request failed for AllAnime URL, using estimate: %v\n", err)
			return 300 * 1024 * 1024, nil // 300MB default for AllAnime
		}
		return 0, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		// For AllAnime URLs that might not have Content-Length, use fallback
		if isAllAnimeURL {
			fmt.Println("Content-Length header missing for AllAnime URL, using fallback estimate")
			return d.estimateContentLengthForAllAnime(url, httpClient)
		}
		return 0, fmt.Errorf("content-length header missing")
	}
	return strconv.ParseInt(contentLength, 10, 64)
}

// estimateContentLengthForAllAnime provides a fallback method to estimate content length for AllAnime URLs
func (d *EpisodeDownloader) estimateContentLengthForAllAnime(url string, client *http.Client) (int64, error) {
	// For streaming URLs (.m3u8), we can't get exact size, so return a reasonable estimate
	if strings.Contains(url, ".m3u8") {
		util.Debugf("HLS stream detected, using estimated size for download")
		// Return an estimated size for a typical episode (500MB)
		return 500 * 1024 * 1024, nil
	}

	// For other AllAnime URLs, try to get partial content to estimate size
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Request only first few KB to check response
	req.Header.Set("Range", "bytes=0-4095")
	resp, err := client.Do(req)
	if err != nil {
		// If range request fails, return default size
		util.Debugf("Range request failed, using default size estimate")
		return 300 * 1024 * 1024, nil // 300MB default
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	// Check Content-Range header for total size
	contentRange := resp.Header.Get("Content-Range")
	if contentRange != "" {
		// Parse "bytes 0-4095/12345678" format
		parts := strings.Split(contentRange, "/")
		if len(parts) == 2 && parts[1] != "*" {
			if totalSize, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return totalSize, nil
			}
		}
	}

	// Fallback to default size estimate
	util.Debugf("Could not determine exact size, using default estimate")
	return 300 * 1024 * 1024, nil // 300MB default
}

// downloadWithProgress downloads a single episode with progress bar
func (d *EpisodeDownloader) downloadWithProgress(videoURL, episodePath string, episodeNum int) error {
	// Create progress model
	m := &progressModel{
		progress: progress.New(progress.WithDefaultGradient()),
	}

	// Get content length for progress tracking
	contentLength, err := d.getContentLength(videoURL)
	if err != nil {
		fmt.Printf("Warning: Failed to get content length: %v, using fallback\n", err)
		// Use a reasonable fallback size for progress tracking
		contentLength = 200 * 1024 * 1024 // 200MB fallback
	}
	m.totalBytes = contentLength

	fmt.Printf("Download setup - Content Length: %d MB\n", contentLength/(1024*1024))

	p := tea.NewProgram(m)

	// Start download in goroutine with proper progress tracking
	downloadComplete := make(chan error, 1)
	go func() {
		// Use the existing player download functionality with progress tracking
		err := d.downloadEpisodeWithProgress(videoURL, episodePath, m, p)

		// Verify the file was actually downloaded before marking as complete
		if err == nil && !d.fileExists(episodePath) {
			err = fmt.Errorf("download failed: file was not created")
		}

		// Send completion status and wait before quitting
		if err == nil {
			p.Send(statusMsg("Download completed!"))
			// Give time for final progress update to show 100%
			time.Sleep(1 * time.Second)
		} else {
			p.Send(statusMsg(fmt.Sprintf("Download failed: %v", err)))
			time.Sleep(500 * time.Millisecond)
		}

		// Mark as done and quit
		m.mu.Lock()
		m.done = true
		m.mu.Unlock()
		p.Quit()

		downloadComplete <- err
	}()

	// Run progress bar - this will block until download is complete
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("progress display error: %w", err)
	}

	// Wait for download completion
	if err := <-downloadComplete; err != nil {
		return err
	}

	// Double-check the file exists and has reasonable size
	if !d.fileExists(episodePath) {
		return fmt.Errorf("download verification failed: file does not exist")
	}

	if stat, err := os.Stat(episodePath); err == nil && stat.Size() < 1024 {
		return fmt.Errorf("download verification failed: file is too small (%d bytes)", stat.Size())
	}

	fmt.Printf("\nEpisode %d downloaded successfully!\n", episodeNum)
	return d.promptPlayDownloaded(episodeNum, episodePath)
}

// downloadEpisodeWithProgress downloads an episode with progress model and Bubble Tea program
func (d *EpisodeDownloader) downloadEpisodeWithProgress(videoURL, destPath string, progressModel *progressModel, program *tea.Program) error {
	// Check if URL is empty or invalid
	if videoURL == "" {
		return fmt.Errorf("empty video URL provided")
	}

	// Inspired by ani-cli download logic
	// Check URL type and use appropriate download method

	// For m3u8 streams (HLS) - use yt-dlp like ani-cli
	if strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "master.m3u8") {
		fmt.Println("Detected HLS stream, using yt-dlp download (ani-cli style)")
		return d.downloadM3U8WithYtDlp(videoURL, destPath, progressModel, program)
	}

	// For wixmp.com URLs (common in AllAnime) - use yt-dlp
	if strings.Contains(videoURL, "wixmp.com") || strings.Contains(videoURL, "repackager.wixmp.com") {
		fmt.Println("Detected wixmp URL, using yt-dlp download")
		return d.downloadM3U8WithYtDlp(videoURL, destPath, progressModel, program)
	}

	// For blogger.com URLs - use yt-dlp
	if strings.Contains(videoURL, "blogger.com") {
		fmt.Println("Detected blogger URL, using yt-dlp download")
		return d.downloadM3U8WithYtDlp(videoURL, destPath, progressModel, program)
	}

	// For sharepoint URLs (AllAnime) - try HTTP first, fallback to yt-dlp
	if strings.Contains(videoURL, "sharepoint.com") {
		fmt.Println("Detected SharePoint URL, trying HTTP download first")
		err := d.downloadHTTPWithProgress(videoURL, destPath, progressModel, program)
		if err != nil {
			fmt.Printf("HTTP download failed: %v, trying yt-dlp fallback\n", err)
			return d.downloadM3U8WithYtDlp(videoURL, destPath, progressModel, program)
		}
		return nil
	}

	// For any AllAnime URL, try yt-dlp as default
	if strings.Contains(videoURL, "allanime") || strings.Contains(videoURL, "allmanga") {
		fmt.Println("Detected AllAnime URL, using yt-dlp download")
		return d.downloadM3U8WithYtDlp(videoURL, destPath, progressModel, program)
	}

	// For regular MP4 URLs - use HTTP download
	fmt.Println("Using HTTP download for regular MP4 URL")
	return d.downloadHTTPWithProgress(videoURL, destPath, progressModel, program)
}

// downloadHTTPWithProgress downloads via HTTP with progress tracking
func (d *EpisodeDownloader) downloadHTTPWithProgress(videoURL, destPath string, progressModel *progressModel, program *tea.Program) error {
	// Create HTTP client with longer timeout for video downloads
	client := &http.Client{
		Transport: api.SafeTransport(10 * time.Minute), // Much longer transport timeout
		Timeout:   0,                                   // No overall timeout - let it download completely
	}

	// Get the file
	resp, err := client.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Ensure directory exists and validate destination path
	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	safeDest, err := d.sanitizeDestPath(destPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}
	// #nosec G304: dest path validated by sanitizeDestPath to remain within configured OutputDir
	out, err := os.Create(safeDest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			util.Warnf("Failed to close output file: %v", closeErr)
		}
	}()

	// Get actual content length from response if available
	actualContentLength := resp.ContentLength
	if actualContentLength > 0 {
		progressModel.mu.Lock()
		progressModel.totalBytes = actualContentLength
		progressModel.mu.Unlock()
	}

	// Copy with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	var totalReceived int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Write to file
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress tracking
			totalReceived += int64(n)

			// Update the progress model
			progressModel.mu.Lock()
			progressModel.received = totalReceived
			// Update total bytes if we got it from response and it's more accurate
			if actualContentLength > 0 {
				progressModel.totalBytes = actualContentLength
			}
			progressModel.mu.Unlock()

			// Send progress update
			program.Send(progressMsg{
				received:   totalReceived,
				totalBytes: progressModel.totalBytes,
			})
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read from response: %w", err)
		}
	}

	// Send final progress update to ensure 100% is shown
	progressModel.mu.Lock()
	progressModel.received = totalReceived
	progressModel.mu.Unlock()

	program.Send(progressMsg{
		received:   totalReceived,
		totalBytes: progressModel.totalBytes,
	})

	fmt.Printf("HTTP download completed: %d bytes downloaded\n", totalReceived)
	return nil
}

// downloadM3U8WithYtDlp downloads m3u8/HLS streams using go-ytdlp library
func (d *EpisodeDownloader) downloadM3U8WithYtDlp(videoURL, destPath string, progressModel *progressModel, program *tea.Program) error {
	program.Send(statusMsg("Starting yt-dlp download (using go-ytdlp library)..."))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Start a goroutine to simulate progress for yt-dlp downloads
	done := make(chan bool, 1)
	go func() {
		// Simulate progress updates since yt-dlp doesn't give us real-time progress easily
		for i := 0; i < 100; i++ {
			select {
			case <-done:
				return
			default:
				time.Sleep(300 * time.Millisecond) // Update every 300ms

				// Simulate gradual progress
				simulatedReceived := int64(float64(progressModel.totalBytes) * float64(i) / 100.0)

				progressModel.mu.Lock()
				progressModel.received = simulatedReceived
				progressModel.mu.Unlock()

				program.Send(progressMsg{
					received:   simulatedReceived,
					totalBytes: progressModel.totalBytes,
				})

				if i%10 == 0 { // Update status every 3 seconds
					program.Send(statusMsg(fmt.Sprintf("Downloading with yt-dlp... %d%%", i)))
				}
			}
		}
	}()

	// Ensure yt-dlp is installed
	ctx := context.Background()
	ytdlp.MustInstall(ctx, nil)

	// Configure downloader using the basic API that we know works
	dl := ytdlp.New().
		Output(destPath) // -o destPath

	// Execute download
	_, err := dl.Run(ctx, videoURL)
	if err != nil {
		done <- true // Stop progress simulation
		return fmt.Errorf("go-ytdlp download failed: %w", err)
	}

	// Stop progress simulation
	done <- true

	// Verify the file was created
	if !d.fileExists(destPath) {
		// List files in directory to see what was created
		if dir := filepath.Dir(destPath); dir != "" {
			if files, err := os.ReadDir(dir); err == nil {
				util.Infof("Files in directory %s:", dir)
				for _, file := range files {
					util.Infof("  - %s", file.Name())
				}
			}
		}
		return fmt.Errorf("download failed: file was not created at %s", destPath)
	}

	// Check file size
	if stat, err := os.Stat(destPath); err == nil {
		if stat.Size() < 1024 {
			return fmt.Errorf("download failed: file is too small (%d bytes)", stat.Size())
		}
	}

	// Update progress to 100%
	progressModel.mu.Lock()
	progressModel.received = progressModel.totalBytes
	progressModel.mu.Unlock()

	program.Send(progressMsg{
		received:   progressModel.totalBytes,
		totalBytes: progressModel.totalBytes,
	})

	program.Send(statusMsg("yt-dlp download completed successfully!"))
	fmt.Printf("Download completed successfully: %s\n", destPath)

	return nil
}

func (d *EpisodeDownloader) downloadWithYtDlp(url, path string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Use go-ytdlp library instead of command line
	ctx := context.Background()
	ytdlp.MustInstall(ctx, nil)

	// Configure downloader
	dl := ytdlp.New().
		Output(path) // -o path

	fmt.Printf("Running go-ytdlp for: %s\n", url)

	// Execute download
	if _, err := dl.Run(ctx, url); err != nil {
		return fmt.Errorf("go-ytdlp error: %w", err)
	}

	// Verify the file was actually downloaded
	if !d.fileExists(path) {
		return fmt.Errorf("download failed: file was not created at %s", path)
	}

	return nil
}

func (d *EpisodeDownloader) promptPlayExisting(episodeNum int, episodePath string) error {
	fmt.Printf("Would you like to play episode %d? (y/n): ", episodeNum)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		util.Warnf("Failed to read input: %v", err)
		return nil
	}

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		return d.playEpisode(episodePath, episodeNum)
	}
	return nil
}

func (d *EpisodeDownloader) promptPlayDownloaded(episodeNum int, episodePath string) error {
	fmt.Printf("Would you like to play the downloaded episode %d? (y/n): ", episodeNum)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		util.Warnf("Failed to read input: %v", err)
		return nil
	}

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		return d.playEpisode(episodePath, episodeNum)
	}
	return nil
}

// promptPlayDownloadedRangeHuh shows a proper UI for episode selection using huh
func (d *EpisodeDownloader) promptPlayDownloadedRangeHuh(episodeNums []int) error {
	if len(episodeNums) == 0 {
		return nil
	}

	// For now, use a simple console prompt until we can properly import huh
	fmt.Printf("Which episode would you like to play? (")
	for i, epNum := range episodeNums {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%d", epNum)
	}
	fmt.Print(", or 0 to exit): ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if choice == 0 {
		return nil
	}

	// Check if choice is in the downloaded episodes
	for _, epNum := range episodeNums {
		if epNum == choice {
			episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", epNum))
			return d.playEpisode(episodePath, epNum)
		}
	}

	fmt.Printf("Episode %d not found in downloaded episodes.\n", choice)
	return nil
}

// promptPlayExistingRangeHuh shows a proper UI for existing episode selection
func (d *EpisodeDownloader) promptPlayExistingRangeHuh(episodeNums []int) error {
	if len(episodeNums) == 0 {
		return nil
	}

	// For now, use a simple console prompt
	fmt.Printf("Which episode would you like to play? (1-%d, or 0 to exit): ", episodeNums[len(episodeNums)-1])

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if choice == 0 {
		return nil
	}

	// Check if choice is in the existing episodes
	for _, epNum := range episodeNums {
		if epNum == choice {
			episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", epNum))
			return d.playEpisode(episodePath, epNum)
		}
	}

	fmt.Printf("Episode %d not found in downloaded episodes.\n", choice)
	return nil
}

func (d *EpisodeDownloader) playEpisode(episodePath string, episodeNum int) error {
	fmt.Printf("Playing episode %d from: %s\n", episodeNum, episodePath)

	// Use StartVideo to play the local file with mpv
	socketPath, err := player.StartVideo(episodePath, []string{})
	if err != nil {
		return fmt.Errorf("failed to start video: %w", err)
	}
	fmt.Printf("Started video playback for episode %d\n", episodeNum)
	fmt.Printf("MPV socket: %s\n", socketPath)
	return nil
}

// tickMsg represents a periodic update message
type tickMsg time.Time

// statusMsg represents a status update message
type statusMsg string

// progressMsg represents a progress update message
type progressMsg struct {
	received   int64
	totalBytes int64
}

// progressModel for tea progress display
type progressModel struct {
	progress   progress.Model
	totalBytes int64
	received   int64
	status     string
	done       bool
	mu         sync.Mutex
}

// tickCmd returns a command that sends a tick message after a delay
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *progressModel) Init() tea.Cmd {
	return tickCmd()
}

func (m *progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case tickMsg:
		if m.done {
			return m, tea.Quit
		}
		m.mu.Lock()
		if m.totalBytes > 0 && m.received > 0 {
			cmd := m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
			m.mu.Unlock()
			return m, tea.Batch(cmd, tickCmd())
		}
		m.mu.Unlock()
		return m, tickCmd()
	case statusMsg:
		m.status = string(msg)
		return m, nil
	case progressMsg:
		m.mu.Lock()
		m.received = msg.received
		m.totalBytes = msg.totalBytes
		// Immediately update the progress bar when we get a progress message
		var cmd tea.Cmd
		if m.totalBytes > 0 {
			cmd = m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
		}
		m.mu.Unlock()
		return m, cmd
	case progress.FrameMsg:
		var cmd tea.Cmd
		newModel, cmd := m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m *progressModel) View() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	percent := 0.0
	if m.totalBytes > 0 {
		percent = float64(m.received) / float64(m.totalBytes) * 100
	}

	status := m.status
	if status == "" {
		status = fmt.Sprintf("Progress: %.1f%%", percent)
	}

	return fmt.Sprintf("Source: %s\n%s\n\nPress Ctrl+C to cancel\n%s",
		"downloading...", // We'll update this with actual URL if needed
		m.progress.View(),
		status)
}
