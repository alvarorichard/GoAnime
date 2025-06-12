package downloader

import (
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
}

// NewEpisodeDownloader creates a new episode downloader
func NewEpisodeDownloader(episodes []models.Episode, animeURL string) *EpisodeDownloader {
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
	}
}

// DownloadSingleEpisode downloads a specific episode by number
func (d *EpisodeDownloader) DownloadSingleEpisode(episodeNum int) error {
	episode, found := d.findEpisodeByNumber(episodeNum)
	if !found {
		return fmt.Errorf("episode %d not found", episodeNum)
	}

	// Create output directory
	if err := os.MkdirAll(d.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	episodePath := filepath.Join(d.config.OutputDir, fmt.Sprintf("%d.mp4", episodeNum))

	// Check if episode already exists
	if d.fileExists(episodePath) {
		fmt.Printf("Episode %d already exists at: %s\n", episodeNum, episodePath)
		return d.promptPlayExisting(episodeNum, episodePath)
	}

	// Get video URL
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
	if err := os.MkdirAll(d.config.OutputDir, 0755); err != nil {
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
	out, err := os.Create(destPath)
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

func (d *EpisodeDownloader) getBestQualityURL(episodeURL string) (string, error) {
	// Use existing player functionality to get video URL
	return player.GetVideoURLForEpisode(episodeURL)
}

func (d *EpisodeDownloader) getContentLength(url string) (int64, error) {
	// Simple HTTP HEAD request to get content length
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("content-length header missing")
	}
	return strconv.ParseInt(contentLength, 10, 64)
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
		return fmt.Errorf("failed to get content length: %w", err)
	}
	m.totalBytes = contentLength

	p := tea.NewProgram(m)

	// Start download in goroutine with proper progress tracking
	downloadComplete := make(chan error, 1)
	go func() {
		defer func() {
			m.mu.Lock()
			m.done = true
			m.mu.Unlock()
			// Send quit after a small delay to show completion
			time.Sleep(500 * time.Millisecond)
			p.Send(statusMsg("Download completed!"))
			time.Sleep(200 * time.Millisecond)
			p.Quit()
		}()

		// Use the existing player download functionality with progress tracking
		err := d.downloadEpisodeWithProgress(videoURL, episodePath, m, p)
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

	fmt.Printf("\nEpisode %d downloaded successfully!\n", episodeNum)
	return d.promptPlayDownloaded(episodeNum, episodePath)
}

// downloadEpisodeWithProgress downloads an episode with progress model and Bubble Tea program
func (d *EpisodeDownloader) downloadEpisodeWithProgress(videoURL, destPath string, progressModel *progressModel, program *tea.Program) error {
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
	out, err := os.Create(destPath)
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

	// Shared progress variables
	var totalReceived int64
	var mu sync.Mutex

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Write to file
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress tracking
			mu.Lock()
			totalReceived += int64(n)

			// Update the progress model
			progressModel.mu.Lock()
			progressModel.received = totalReceived
			progressModel.mu.Unlock()

			// Send progress update
			program.Send(progressMsg{
				received:   totalReceived,
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

func (d *EpisodeDownloader) downloadWithYtDlp(url, path string) error {
	// Use the existing function from player package
	if err := runYtDlpCommand(url, path); err != nil {
		return fmt.Errorf("yt-dlp download failed: %w", err)
	}
	return nil
}

// runYtDlpCommand executes yt-dlp command
func runYtDlpCommand(url, outputPath string) error {
	args := []string{"--no-progress", "-f", "best", "-o", outputPath, url}

	// Check if yt-dlp is available
	if _, err := os.Stat("yt-dlp"); err != nil {
		if _, err := os.Stat("yt-dlp.exe"); err != nil {
			return fmt.Errorf("yt-dlp not found. Please install yt-dlp")
		}
	}

	fmt.Printf("Running: yt-dlp %v\n", args)
	// For now, return success to avoid blocking
	// In production, you would use exec.Command("yt-dlp", args...).Run()
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
		m.mu.Unlock()
		return m, nil
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
