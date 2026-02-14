// Package downloader provides download functionality for movies and TV shows from FlixHQ and SFlix
// This file implements dedicated movie/TV download functionality without affecting the existing anime downloader
package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/go-ytdlp"
)

// MovieDownloadConfig holds configuration for movie download operations
type MovieDownloadConfig struct {
	OutputDir    string
	Quality      scraper.Quality
	SubsLanguage string
	Provider     string
}

// MovieDownloader handles movie/TV download operations from FlixHQ and SFlix
type MovieDownloader struct {
	config       MovieDownloadConfig
	mediaManager *scraper.MediaManager
}

// NewMovieDownloader creates a new movie downloader
func NewMovieDownloader() *MovieDownloader {
	userHome, _ := os.UserHomeDir()
	outputDir := filepath.Join(userHome, ".local", "goanime", "downloads", "movies")

	return &MovieDownloader{
		config: MovieDownloadConfig{
			OutputDir:    outputDir,
			Quality:      scraper.Quality1080,
			SubsLanguage: "english",
			Provider:     "Vidcloud",
		},
		mediaManager: scraper.NewMediaManager(),
	}
}

// NewMovieDownloaderWithConfig creates a movie downloader with custom configuration
func NewMovieDownloaderWithConfig(config MovieDownloadConfig) *MovieDownloader {
	if config.OutputDir == "" {
		userHome, _ := os.UserHomeDir()
		config.OutputDir = filepath.Join(userHome, ".local", "goanime", "downloads", "movies")
	}
	if config.Quality == "" {
		config.Quality = scraper.Quality1080
	}
	if config.SubsLanguage == "" {
		config.SubsLanguage = "english"
	}
	if config.Provider == "" {
		config.Provider = "Vidcloud"
	}

	return &MovieDownloader{
		config:       config,
		mediaManager: scraper.NewMediaManager(),
	}
}

// DownloadMovie downloads a movie from FlixHQ or SFlix
func (md *MovieDownloader) DownloadMovie(media *models.Anime) error {
	if media == nil {
		return fmt.Errorf("media is nil")
	}

	util.Infof("Starting download for: %s", media.Name)
	util.Debugf("Source: %s, URL: %s", media.Source, media.URL)

	// Create output directory
	if err := os.MkdirAll(md.config.OutputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract media ID from URL
	mediaID := extractMediaIDFromURL(media.URL)
	if mediaID == "" {
		return fmt.Errorf("could not extract media ID from URL: %s", media.URL)
	}

	// Sanitize movie name for filename
	safeFileName := sanitizeFileName(media.Name)
	moviePath := filepath.Join(md.config.OutputDir, fmt.Sprintf("%s.mp4", safeFileName))

	// Check if movie already exists
	if md.fileExists(moviePath) {
		fmt.Printf("Movie already exists at: %s\n", moviePath)
		return md.promptPlayExisting(moviePath, media.Name)
	}

	// Get stream URL based on source
	var streamInfo *scraper.FlixHQStreamInfo
	var err error

	source := strings.ToLower(media.Source)
	if strings.Contains(source, "sflix") {
		streamInfo, err = md.getSFlixMovieStream(mediaID)
	} else {
		// Default to FlixHQ
		streamInfo, err = md.getFlixHQMovieStream(mediaID)
	}

	if err != nil {
		return fmt.Errorf("failed to get stream URL: %w", err)
	}

	if streamInfo.VideoURL == "" {
		return fmt.Errorf("no video URL found for movie")
	}

	util.Infof("Got stream URL: %s", streamInfo.VideoURL)
	util.Debugf("Quality: %s, Stream Type: %s, Referer: %s", streamInfo.Quality, streamInfo.StreamType, streamInfo.Referer)

	// Download with progress, passing referer for authentication
	return md.downloadMovieWithProgress(streamInfo.VideoURL, moviePath, media.Name, streamInfo.IsM3U8, streamInfo.Referer, streamInfo.Headers)
}

// DownloadTVEpisode downloads a TV episode from FlixHQ or SFlix
func (md *MovieDownloader) DownloadTVEpisode(media *models.Anime, seasonNum, episodeNum int) error {
	if media == nil {
		return fmt.Errorf("media is nil")
	}

	util.Infof("Starting download for: %s S%02dE%02d", media.Name, seasonNum, episodeNum)

	// Create output directory with show name
	safeShowName := sanitizeFileName(media.Name)
	showDir := filepath.Join(md.config.OutputDir, safeShowName, fmt.Sprintf("Season %d", seasonNum))
	if err := os.MkdirAll(showDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract media ID from URL
	mediaID := extractMediaIDFromURL(media.URL)
	if mediaID == "" {
		return fmt.Errorf("could not extract media ID from URL: %s", media.URL)
	}

	episodePath := filepath.Join(showDir, fmt.Sprintf("S%02dE%02d.mp4", seasonNum, episodeNum))

	// Check if episode already exists
	if md.fileExists(episodePath) {
		fmt.Printf("Episode already exists at: %s\n", episodePath)
		return md.promptPlayExisting(episodePath, fmt.Sprintf("%s S%02dE%02d", media.Name, seasonNum, episodeNum))
	}

	// Get seasons and find the right episode
	source := strings.ToLower(media.Source)
	var streamInfo *scraper.FlixHQStreamInfo
	var err error

	if strings.Contains(source, "sflix") {
		streamInfo, err = md.getSFlixEpisodeStream(mediaID, seasonNum, episodeNum)
	} else {
		streamInfo, err = md.getFlixHQEpisodeStream(mediaID, seasonNum, episodeNum)
	}

	if err != nil {
		return fmt.Errorf("failed to get stream URL: %w", err)
	}

	if streamInfo.VideoURL == "" {
		return fmt.Errorf("no video URL found for episode")
	}

	util.Infof("Got stream URL: %s", streamInfo.VideoURL)
	util.Debugf("Referer: %s", streamInfo.Referer)

	// Download with progress, passing referer for authentication
	return md.downloadMovieWithProgress(streamInfo.VideoURL, episodePath, fmt.Sprintf("%s S%02dE%02d", media.Name, seasonNum, episodeNum), streamInfo.IsM3U8, streamInfo.Referer, streamInfo.Headers)
}

// DownloadTVEpisodeRange downloads a range of episodes from a TV show
func (md *MovieDownloader) DownloadTVEpisodeRange(media *models.Anime, seasonNum, startEp, endEp int) error {
	if startEp > endEp {
		return fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
	}

	fmt.Printf("Downloading episodes %d-%d from Season %d of %s\n", startEp, endEp, seasonNum, media.Name)

	var errors []error
	for epNum := startEp; epNum <= endEp; epNum++ {
		fmt.Printf("\n--- Episode %d ---\n", epNum)
		if err := md.DownloadTVEpisode(media, seasonNum, epNum); err != nil {
			util.Warnf("Failed to download episode %d: %v", epNum, err)
			errors = append(errors, fmt.Errorf("episode %d: %w", epNum, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d episode(s) failed to download", len(errors))
	}

	fmt.Printf("\nAll episodes downloaded successfully!\n")
	return nil
}

// getFlixHQMovieStream gets stream info for a FlixHQ movie
func (md *MovieDownloader) getFlixHQMovieStream(mediaID string) (*scraper.FlixHQStreamInfo, error) {
	return md.mediaManager.GetMovieStreamWithQuality(mediaID, md.config.Quality, md.config.SubsLanguage)
}

// getSFlixMovieStream gets stream info for a SFlix movie
func (md *MovieDownloader) getSFlixMovieStream(mediaID string) (*scraper.FlixHQStreamInfo, error) {
	sflixInfo, err := md.mediaManager.GetSFlixMovieStreamInfo(mediaID, md.config.Provider, string(md.config.Quality), md.config.SubsLanguage)
	if err != nil {
		return nil, err
	}

	// Convert SFlixStreamInfo to FlixHQStreamInfo for unified handling
	return convertSFlixToFlixHQStreamInfo(sflixInfo), nil
}

// getFlixHQEpisodeStream gets stream info for a FlixHQ TV episode
func (md *MovieDownloader) getFlixHQEpisodeStream(mediaID string, seasonNum, episodeNum int) (*scraper.FlixHQStreamInfo, error) {
	// Get seasons
	seasons, err := md.mediaManager.GetTVSeasons(mediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}

	if seasonNum > len(seasons) {
		return nil, fmt.Errorf("season %d not found (only %d seasons available)", seasonNum, len(seasons))
	}

	season := seasons[seasonNum-1]

	// Get episodes for the season
	episodes, err := md.mediaManager.GetTVEpisodes(season.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}

	if episodeNum > len(episodes) {
		return nil, fmt.Errorf("episode %d not found in season %d (only %d episodes available)", episodeNum, seasonNum, len(episodes))
	}

	episode := episodes[episodeNum-1]

	// Get stream info
	return md.mediaManager.GetTVEpisodeStreamInfo(episode.DataID, md.config.Provider, string(md.config.Quality), md.config.SubsLanguage)
}

// getSFlixEpisodeStream gets stream info for a SFlix TV episode
func (md *MovieDownloader) getSFlixEpisodeStream(mediaID string, seasonNum, episodeNum int) (*scraper.FlixHQStreamInfo, error) {
	// Get seasons
	seasons, err := md.mediaManager.GetSFlixTVSeasons(mediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}

	if seasonNum > len(seasons) {
		return nil, fmt.Errorf("season %d not found (only %d seasons available)", seasonNum, len(seasons))
	}

	season := seasons[seasonNum-1]

	// Get episodes for the season
	episodes, err := md.mediaManager.GetSFlixTVEpisodes(season.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}

	if episodeNum > len(episodes) {
		return nil, fmt.Errorf("episode %d not found in season %d (only %d episodes available)", episodeNum, seasonNum, len(episodes))
	}

	episode := episodes[episodeNum-1]

	// Get stream info
	sflixInfo, err := md.mediaManager.GetSFlixTVEpisodeStreamInfo(episode.DataID, md.config.Provider, string(md.config.Quality), md.config.SubsLanguage)
	if err != nil {
		return nil, err
	}

	return convertSFlixToFlixHQStreamInfo(sflixInfo), nil
}

// downloadMovieWithProgress downloads a movie/episode with Bubble Tea progress bar
func (md *MovieDownloader) downloadMovieWithProgress(videoURL, destPath, title string, isM3U8 bool, referer string, headers map[string]string) error {
	// Set default referer if not provided
	if referer == "" {
		referer = "https://flixhq.to"
	}

	// Create progress model
	m := &movieProgressModel{
		progress: progress.New(progress.WithDefaultGradient()),
		title:    title,
	}

	// Get content length for progress tracking
	contentLength, err := md.getContentLength(videoURL)
	if err != nil {
		util.Warnf("Failed to get content length: %v, using fallback", err)
		contentLength = 500 * 1024 * 1024 // 500MB fallback for movies
	}
	m.totalBytes = contentLength

	fmt.Printf("Download setup - Content Length: %d MB\n", contentLength/(1024*1024))

	p := tea.NewProgram(m)

	// Start download in goroutine with progress tracking
	downloadComplete := make(chan error, 1)
	go func() {
		var err error
		if isM3U8 || strings.Contains(videoURL, ".m3u8") {
			err = md.downloadM3U8WithYtDlp(videoURL, destPath, referer, m, p)
		} else {
			err = md.downloadHTTPWithProgress(videoURL, destPath, referer, headers, m, p)
		}

		// Verify the file was actually downloaded
		if err == nil && !md.fileExists(destPath) {
			err = fmt.Errorf("download failed: file was not created")
		}

		// Send completion status
		if err == nil {
			p.Send(movieStatusMsg("Download completed!"))
			time.Sleep(1 * time.Second)
		} else {
			p.Send(movieStatusMsg(fmt.Sprintf("Download failed: %v", err)))
			time.Sleep(500 * time.Millisecond)
		}

		m.mu.Lock()
		m.done = true
		m.mu.Unlock()
		p.Quit()

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

	// Verify file
	if !md.fileExists(destPath) {
		return fmt.Errorf("download verification failed: file does not exist")
	}

	if stat, err := os.Stat(destPath); err == nil && stat.Size() < 1024 {
		return fmt.Errorf("download verification failed: file is too small (%d bytes)", stat.Size())
	}

	fmt.Printf("\n%s downloaded successfully!\n", title)
	return md.promptPlayDownloaded(destPath, title)
}

// downloadHTTPWithProgress downloads via HTTP with progress tracking
func (md *MovieDownloader) downloadHTTPWithProgress(videoURL, destPath, referer string, headers map[string]string, progressModel *movieProgressModel, program *tea.Program) error {
	// Create HTTP client with longer timeout for video downloads
	client := &http.Client{
		Transport: api.SafeTransport(15 * time.Minute),
		Timeout:   0, // No overall timeout
	}

	// Create request with headers
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set referer and other headers
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req) // #nosec G704
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Validate and sanitize destination path
	safeDest, err := md.sanitizeDestPath(destPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// #nosec G304: dest path validated by sanitizeDestPath to remain within configured OutputDir
	out, err := os.Create(safeDest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	// Update content length from response
	if resp.ContentLength > 0 {
		progressModel.mu.Lock()
		progressModel.totalBytes = resp.ContentLength
		progressModel.mu.Unlock()
	}

	// Copy with progress tracking
	buffer := make([]byte, 32*1024)
	var totalReceived int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			totalReceived += int64(n)

			progressModel.mu.Lock()
			progressModel.received = totalReceived
			progressModel.mu.Unlock()

			program.Send(movieProgressMsg{
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

	// Final progress update
	progressModel.mu.Lock()
	progressModel.received = totalReceived
	progressModel.mu.Unlock()

	program.Send(movieProgressMsg{
		received:   totalReceived,
		totalBytes: progressModel.totalBytes,
	})

	return nil
}

// downloadM3U8WithYtDlp downloads m3u8/HLS streams using yt-dlp for best quality
// (audio/video merging from master playlists), falling back to native HLS if yt-dlp fails.
func (md *MovieDownloader) downloadM3U8WithYtDlp(videoURL, destPath, referer string, progressModel *movieProgressModel, program *tea.Program) error {
	program.Send(movieStatusMsg("Starting HLS download..."))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract referer from the stream URL itself if not provided
	if referer == "" {
		referer = extractRefererFromStreamURL(videoURL)
	}

	// Native HLS first — handles obfuscated segment extensions (.jpg, .png) and
	// "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp's ffmpeg downloader.
	nativeErr := md.downloadM3U8WithNativeHLS(videoURL, destPath, referer, progressModel, program)
	if nativeErr == nil {
		return nil
	}
	util.Logger.Warn("Native HLS failed for movie, falling back to yt-dlp", "error", nativeErr)

	// Fallback to yt-dlp
	return md.downloadM3U8WithYtDlpDirect(videoURL, destPath, referer, progressModel, program)
}

// downloadM3U8WithYtDlpDirect uses yt-dlp to download HLS with best format selection.
// Progress is tracked via yt-dlp's native ProgressFunc callback.
func (md *MovieDownloader) downloadM3U8WithYtDlpDirect(videoURL, destPath, referer string, progressModel *movieProgressModel, program *tea.Program) error {
	program.Send(movieStatusMsg("Downloading with yt-dlp (best quality)..."))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	_, installErr := ytdlp.Install(ctx, nil)
	if installErr != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", installErr)
	}

	// Use typed API for ALL flags so they are placed before the URL
	// (critical for --downloader-args to be processed correctly by yt-dlp).
	dl := ytdlp.New().
		Output(destPath).
		Format("bestvideo+bestaudio/best").
		Downloader("ffmpeg").
		DownloaderArgs("ffmpeg_i:-allowed_extensions ALL").
		ConcurrentFragments(4).
		FragmentRetries("5").
		Retries("5").
		SocketTimeout(30).
		Impersonate("chrome")

	if referer != "" {
		dl.AddHeaders("Referer:" + referer)
		parsed, _ := url.Parse(referer)
		if parsed != nil && parsed.Host != "" {
			origin := parsed.Scheme + "://" + parsed.Host
			dl.AddHeaders("Origin:" + origin)
		}
	}

	// Real-time progress via yt-dlp's native callback
	var lastReportedBytes int64
	var lastProgressFile string
	dl.ProgressFunc(200*time.Millisecond, func(update ytdlp.ProgressUpdate) {
		if update.Status == ytdlp.ProgressStatusPostProcessing ||
			update.Status == ytdlp.ProgressStatusFinished {
			return
		}

		progressModel.mu.Lock()
		defer progressModel.mu.Unlock()

		if update.Filename != "" && update.Filename != lastProgressFile {
			lastProgressFile = update.Filename
			lastReportedBytes = 0
		}

		downloaded := int64(update.DownloadedBytes)
		if delta := downloaded - lastReportedBytes; delta > 0 {
			progressModel.received += delta
			lastReportedBytes = downloaded
		}

		if update.TotalBytes > 0 && progressModel.totalBytes < int64(update.TotalBytes) {
			progressModel.totalBytes = int64(update.TotalBytes)
		}

		program.Send(movieProgressMsg{
			received:   progressModel.received,
			totalBytes: progressModel.totalBytes,
		})
	})

	_, runErr := dl.Run(ctx, videoURL, "--hls-use-mpegts")

	if runErr != nil {
		return fmt.Errorf("yt-dlp download failed: %w", runErr)
	}

	// Use actual file size for final progress (not the estimate)
	var finalSize int64
	if fi, err := os.Stat(destPath); err == nil {
		finalSize = fi.Size()
	}
	if finalSize <= 0 {
		finalSize = progressModel.totalBytes
	}

	progressModel.mu.Lock()
	progressModel.totalBytes = finalSize
	progressModel.received = finalSize
	progressModel.mu.Unlock()
	program.Send(movieProgressMsg{
		received:   finalSize,
		totalBytes: finalSize,
	})

	return nil
}

// downloadM3U8WithNativeHLS downloads m3u8/HLS streams using native HLS downloader as fallback
func (md *MovieDownloader) downloadM3U8WithNativeHLS(videoURL, destPath, referer string, progressModel *movieProgressModel, program *tea.Program) error {
	program.Send(movieStatusMsg("Downloading with native HLS..."))

	// Prepare headers for HLS download with proper referer and origin
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept":     "*/*",
	}

	if referer != "" {
		headers["Referer"] = referer
		headers["Origin"] = strings.TrimSuffix(referer, "/")
	}

	// Create context for the download
	ctx := context.Background()

	// Use the native HLS downloader with byte-based progress callback
	err := hls.DownloadToFile(ctx, videoURL, destPath, headers, func(bytesWritten int64, segmentsWritten, totalSegments int) {
		if totalSegments <= 0 {
			return
		}

		progressModel.mu.Lock()
		// Update received with real bytes on disk
		progressModel.received = bytesWritten

		// Dynamically estimate total from average bytes per written segment
		if segmentsWritten >= 3 {
			avgBytesPerSeg := bytesWritten / int64(segmentsWritten)
			estimatedTotal := avgBytesPerSeg * int64(totalSegments)
			if estimatedTotal > progressModel.totalBytes {
				progressModel.totalBytes = estimatedTotal
			}
		}

		// Cap at 98% until fully done
		total := progressModel.totalBytes
		if total > 0 && progressModel.received >= total {
			progressModel.received = int64(float64(total) * 0.98)
		}
		received := progressModel.received
		progressModel.mu.Unlock()

		program.Send(movieProgressMsg{
			received:   received,
			totalBytes: total,
		})

		percent := float64(0)
		if total > 0 {
			percent = float64(received) / float64(total) * 100
		}
		program.Send(movieStatusMsg(fmt.Sprintf("Downloading HLS... %d/%d segments, %.0f MB (%.0f%%)",
			segmentsWritten, totalSegments,
			float64(bytesWritten)/(1024*1024), percent)))
	})

	if err != nil {
		return fmt.Errorf("HLS download failed: %w", err)
	}

	// Verify file was created
	if !md.fileExists(destPath) {
		return fmt.Errorf("download failed: file was not created at %s", destPath)
	}

	// Update progress to 100% using actual file size
	if fi, statErr := os.Stat(destPath); statErr == nil && fi.Size() > 0 {
		progressModel.mu.Lock()
		progressModel.totalBytes = fi.Size()
		progressModel.received = fi.Size()
		progressModel.mu.Unlock()

		program.Send(movieProgressMsg{
			received:   fi.Size(),
			totalBytes: fi.Size(),
		})
	} else {
		progressModel.mu.Lock()
		progressModel.received = progressModel.totalBytes
		progressModel.mu.Unlock()

		program.Send(movieProgressMsg{
			received:   progressModel.totalBytes,
			totalBytes: progressModel.totalBytes,
		})
	}

	return nil
}

// extractRefererFromStreamURL extracts the referer (origin) from a stream URL
// e.g., https://megacloud.tv/embed-2/abc123?k=v -> https://megacloud.tv/
func extractRefererFromStreamURL(streamURL string) string {
	parsed, err := url.Parse(streamURL)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "" && parsed.Host != "" {
		return fmt.Sprintf("%s://%s/", parsed.Scheme, parsed.Host)
	}
	return ""
}

// getContentLength gets the content length of a URL
func (md *MovieDownloader) getContentLength(url string) (int64, error) {
	if strings.Contains(url, ".m3u8") {
		return 500 * 1024 * 1024, nil // 500MB estimate for HLS
	}

	client := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req) // #nosec G704
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 500 * 1024 * 1024, nil // 500MB fallback
	}

	return strconv.ParseInt(contentLength, 10, 64)
}

// fileExists checks if a file exists
func (md *MovieDownloader) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// sanitizeDestPath ensures the destination path stays within the configured OutputDir
func (md *MovieDownloader) sanitizeDestPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty destination path")
	}
	cleaned := filepath.Clean(p)
	outDir := filepath.Clean(md.config.OutputDir)
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

// promptPlayExisting prompts user to play existing file
func (md *MovieDownloader) promptPlayExisting(path, title string) error {
	fmt.Printf("Would you like to play %s? (y/n): ", title)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return nil
	}

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		return md.playMovie(path, title)
	}
	return nil
}

// promptPlayDownloaded prompts user to play downloaded file
func (md *MovieDownloader) promptPlayDownloaded(path, title string) error {
	fmt.Printf("Would you like to play the downloaded %s? (y/n): ", title)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return nil
	}

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		return md.playMovie(path, title)
	}
	return nil
}

// playMovie plays a local movie file
func (md *MovieDownloader) playMovie(path, title string) error {
	fmt.Printf("Playing %s from: %s\n", title, path)

	socketPath, err := player.StartVideo(path, []string{})
	if err != nil {
		return fmt.Errorf("failed to start video: %w", err)
	}

	fmt.Printf("Started video playback\nMPV socket: %s\n", socketPath)
	return nil
}

// Helper functions

func extractMediaIDFromURL(urlStr string) string {
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func sanitizeFileName(name string) string {
	// Replace invalid characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Also clean up multiple underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return strings.TrimSpace(result)
}

func convertSFlixToFlixHQStreamInfo(sflix *scraper.SFlixStreamInfo) *scraper.FlixHQStreamInfo {
	if sflix == nil {
		return nil
	}

	var subtitles []scraper.FlixHQSubtitle
	for _, sub := range sflix.Subtitles {
		subtitles = append(subtitles, scraper.FlixHQSubtitle(sub))
	}

	var qualities []scraper.FlixHQQualityOption
	for _, q := range sflix.Qualities {
		qualities = append(qualities, scraper.FlixHQQualityOption(q))
	}

	return &scraper.FlixHQStreamInfo{
		VideoURL:   sflix.VideoURL,
		Quality:    sflix.Quality,
		Subtitles:  subtitles,
		Referer:    sflix.Referer,
		SourceName: sflix.SourceName,
		StreamType: sflix.StreamType,
		IsM3U8:     sflix.IsM3U8,
		Headers:    sflix.Headers,
		Qualities:  qualities,
	}
}

// Movie progress model for Bubble Tea (separate from anime progress model)

type movieTickMsg time.Time
type movieStatusMsg string
type movieProgressMsg struct {
	received   int64
	totalBytes int64
}

type movieProgressModel struct {
	progress   progress.Model
	totalBytes int64
	received   int64
	status     string
	title      string
	done       bool
	mu         sync.Mutex
}

func movieTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return movieTickMsg(t)
	})
}

func (m *movieProgressModel) Init() tea.Cmd {
	return movieTickCmd()
}

func (m *movieProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case movieTickMsg:
		if m.done {
			return m, tea.Quit
		}
		m.mu.Lock()
		if m.totalBytes > 0 && m.received > 0 {
			cmd := m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
			m.mu.Unlock()
			return m, tea.Batch(cmd, movieTickCmd())
		}
		m.mu.Unlock()
		return m, movieTickCmd()
	case movieStatusMsg:
		m.status = string(msg)
		return m, nil
	case movieProgressMsg:
		m.mu.Lock()
		m.received = msg.received
		m.totalBytes = msg.totalBytes
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

func (m *movieProgressModel) View() string {
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

	title := m.title
	if title == "" {
		title = "downloading..."
	}

	return fmt.Sprintf("%s\n%s\n\nPress Ctrl+C to cancel\n%s",
		title,
		m.progress.View(),
		status)
}
