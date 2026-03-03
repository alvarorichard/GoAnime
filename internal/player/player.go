package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/upscaler"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
)

// lastAnimeURL stores the most recent anime URL/ID to support navigation when no updater is present
var lastAnimeURL string

// lastAnimeName stores the most recent anime name for Plex-compatible download naming
var lastAnimeName string

// lastAnimeSeason stores the most recent anime season number for download naming
var lastAnimeSeason int

// lastIsMovieOrTV indicates whether the current content is a movie/TV show (non-anime)
var lastIsMovieOrTV bool

// lastMediaType stores the exact media type for intelligent path organization
var lastMediaType string // "movie", "tv", or "anime"

// SetAnimeName sets the anime name and season for Plex-compatible download file naming.
// Call this before any download operations to ensure proper naming.
func SetAnimeName(name string, season int) {
	lastAnimeName = name
	lastAnimeSeason = season
	if season < 1 {
		lastAnimeSeason = 1
	}
}

// SetMediaType marks whether the current content is a movie/TV show (true) or anime (false).
// This determines whether downloads go to the movies or anime directory.
func SetMediaType(isMovieOrTV bool) {
	lastIsMovieOrTV = isMovieOrTV
}

// SetExactMediaType stores the exact media type ("movie", "tv", "anime") for
// intelligent download path organization. Movies get flat paths, TV shows and
// anime get season/episode structures.
func SetExactMediaType(mediaType string) {
	lastMediaType = mediaType
	lastIsMovieOrTV = (mediaType == "movie" || mediaType == "tv")
}

// GetExactMediaType returns the current exact media type.
func GetExactMediaType() string {
	return lastMediaType
}

// IsCurrentMediaMovie returns true if the current content is a standalone movie.
func IsCurrentMediaMovie() bool {
	return lastMediaType == "movie"
}

const (
	padding = 2
)

// tickMsg is a message for the tick command
type tickMsg time.Time

// statusMsg is a message to update the status
type statusMsg string

// model represents the Bubble Tea model for the progress bar and status
type model struct {
	progress   progress.Model
	totalBytes int64
	received   int64
	done       bool
	doneFrames int // frames elapsed since done; allows 100% to render before quit
	status     string
	mu         sync.Mutex
	keys       keyMap
}

type keyMap struct {
	quit key.Binding
}

// Init initializes the Bubble Tea model
func (m *model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.progress.Init())
}

// StartVideo opens mpv with a socket for IPC
// Modify the StartVideo function in player.go
func StartVideo(link string, args []string) (string, error) {
	// Verify MPV is installed using platform-specific search
	mpvPath, err := findMPVPath()
	if err != nil {
		return "", fmt.Errorf("mpv not found: %w\nPlease install mpv: https://mpv.io/installation/", err)
	}

	randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
	var socketPath string

	if runtime.GOOS == "windows" {
		socketPath = fmt.Sprintf(`\\.\pipe\goanime_mpvsocket_%s`, randomNumber)
	} else {
		// Use os.TempDir() for cross-platform compatibility
		// macOS uses /var/folders/... accessed via $TMPDIR
		// filepath.Join handles trailing slashes correctly (fixes macOS double-slash issue)
		socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("goanime_mpvsocket_%s", randomNumber))
	}

	mpvArgs := []string{
		"--no-terminal",
		"--quiet",
		"--force-window=yes",
		fmt.Sprintf("--input-ipc-server=%s", socketPath),
	}
	// Validate and filter any additional args before passing to mpv
	mpvArgs = append(mpvArgs, filterMPVArgs(args)...)

	// Sanitize media target (URL or local file path)
	safeLink, err := sanitizeMediaTarget(link)
	if err != nil {
		return "", fmt.Errorf("invalid media target: %w", err)
	}
	mpvArgs = append(mpvArgs, safeLink)

	util.Debugf("Starting mpv with arguments: %v", mpvArgs)

	// #nosec G204: mpvArgs are validated via filterMPVArgs and sanitizeMediaTarget
	cmd := exec.Command(mpvPath, mpvArgs...)
	setProcessGroup(cmd) // Handle OS-specific process groups

	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mpv: %w (stderr: %s)", err, stderr.String())
	}

	util.Debugf("mpv started, waiting for socket creation: %s", socketPath)

	// Wait for socket creation with adaptive timeout and exponential backoff
	// Total max wait time: ~10 seconds (accommodates slow network streams)
	// Initial intervals are short for fast local files, then back off for streams
	maxWaitTime := 10 * time.Second
	initialInterval := 50 * time.Millisecond
	maxInterval := 500 * time.Millisecond
	currentInterval := initialInterval

	for time.Since(startTime) < maxWaitTime {
		util.Debugf("Attempt at %.2fs: checking socket connection...", time.Since(startTime).Seconds())

		// Try to connect to the socket instead of checking file existence
		// This works for both Unix sockets and Windows named pipes
		conn, err := dialMPVSocket(socketPath)
		if err == nil {
			_ = conn.Close() // Close immediately, we just wanted to verify connectivity
			util.Debugf("Socket connected successfully after %.2fs", time.Since(startTime).Seconds())
			return socketPath, nil
		}

		util.Debugf("Connection attempt failed: %v", err)

		// Check if MPV process is still running
		if cmd.Process == nil {
			return "", fmt.Errorf("mpv process not started properly: %s", stderr.String())
		}

		// Check if process exited prematurely
		// Note: ProcessState is nil until the process exits, so we need a different check
		select {
		case <-time.After(currentInterval):
			// Apply exponential backoff
			currentInterval = min(time.Duration(float64(currentInterval)*1.5), maxInterval)
		default:
			time.Sleep(currentInterval)
			currentInterval = min(time.Duration(float64(currentInterval)*1.5), maxInterval)
		}
	}

	elapsed := time.Since(startTime)
	util.Debugf("Timeout after %.2fs waiting for mpv socket", elapsed.Seconds())

	// Cleanup if timeout occurs
	if killErr := cmd.Process.Kill(); killErr != nil {
		util.Debugf("Failed to kill mpv process: %v", killErr)
	}

	return "", fmt.Errorf("timeout waiting for mpv socket after %.1fs. Possible issues:\n1. Slow network connection - video source may be unresponsive\n2. MPV installation corrupted\n3. Firewall blocking IPC\n4. Invalid video URL\nCheck debug logs with -debug flag", elapsed.Seconds())
}

// MpvSendCommand is a wrapper function to expose mpvSendCommand to other packages
func MpvSendCommand(socketPath string, command []any) (any, error) {
	return mpvSendCommand(socketPath, command)
}

// filterMPVArgs whitelists allowed mpv flags to avoid passing unexpected parameters.
func filterMPVArgs(args []string) []string {
	allowedNoValue := map[string]struct{}{
		"--no-config": {},
	}
	allowedWithValuePrefixes := []string{
		"--hwdec=",
		"--vo=",
		"--gpu-context=",
		"--profile=",
		"--cache=",
		"--demuxer-max-bytes=",
		"--demuxer-readahead-secs=",
		"--video-latency-hacks=",
		"--audio-display=",
		"--start=",
		"--alang=",              // Audio language preference
		"--slang=",              // Subtitle language preference
		"--aid=",                // Audio track ID
		"--sid=",                // Subtitle track ID
		"--sub-file=",           // External subtitle file (single)
		"--sub-files=",          // External subtitle files (multiple, colon-separated)
		"--audio-file=",         // External audio file
		"--http-header-fields=", // HTTP headers for HLS streams
		"--stream-lavf-o=",      // FFmpeg/lavf options for streaming protocols
		"--referrer=",           // HTTP referrer for streaming
		"--user-agent=",         // HTTP user agent for streaming
		// Anime4K real-time upscaling shaders
		"--glsl-shader=",          // GLSL shader for video processing
		"--glsl-shaders=",         // Multiple GLSL shaders (colon-separated)
		"--gpu-shader-cache-dir=", // Shader cache directory
		"--gpu-api=",              // GPU API selection (auto, opengl, vulkan, d3d11)
		// yt-dlp integration for Cloudflare-protected streams (9Anime, etc.)
		"--script-opts=",             // mpv script options (e.g. ytdl_hook-try_ytdl_first)
		"--ytdl-raw-options-append=", // Pass raw options to yt-dlp backend
		"--ytdl-format=",             // yt-dlp format / quality selection
		// Add more allowed prefixes here if needed in the future
	}

	var filtered []string
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			// ignore positional args; media target is handled separately
			continue
		}
		if _, ok := allowedNoValue[a]; ok {
			filtered = append(filtered, a)
			continue
		}
		for _, p := range allowedWithValuePrefixes {
			if strings.HasPrefix(a, p) {
				filtered = append(filtered, a)
				break
			}
		}
	}
	return filtered
}

// sanitizeMediaTarget ensures the media target is a safe http(s) URL or a cleaned file path
func sanitizeMediaTarget(link string) (string, error) {
	l := strings.TrimSpace(link)
	if l == "" {
		return "", fmt.Errorf("empty link")
	}
	if strings.ContainsAny(l, "\x00\n\r") {
		return "", fmt.Errorf("invalid control characters in link")
	}
	if strings.HasPrefix(l, "-") {
		return "", fmt.Errorf("media target must not start with '-' (looks like a flag)")
	}
	// Treat as URL only if it contains "://". This avoids misclassifying Windows
	// paths like "C:\\..." as having scheme "c".
	if strings.Contains(l, "://") {
		u, err := url.Parse(l)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			return l, nil
		default:
			return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
		}
	}
	// Treat as local path
	cleaned := filepath.Clean(l)
	return cleaned, nil
}

// sanitizeOutputPath validates an output path to avoid directory traversal and disallow leading '-'
func sanitizeOutputPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty output path")
	}
	if strings.ContainsAny(p, "\x00\n\r") {
		return "", fmt.Errorf("invalid control characters in output path")
	}
	if strings.HasPrefix(p, "-") {
		return "", fmt.Errorf("output path must not start with '-' (looks like a flag)")
	}
	cleaned := filepath.Clean(p)
	// Verify the resolved path stays within user home to prevent path traversal
	userHome, err := os.UserHomeDir()
	if err == nil {
		abs, absErr := filepath.Abs(cleaned)
		if absErr == nil && !strings.HasPrefix(abs, userHome) {
			return "", fmt.Errorf("output path escapes user home directory")
		}
	}
	return cleaned, nil
}

// mpvSendCommand sends a JSON command to MPV via the IPC socket and receives the response.
func mpvSendCommand(socketPath string, command []any) (any, error) {
	conn, err := dialMPVSocket(socketPath)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing mpv socket")
		}
	}(conn)

	commandJSON, err := json.Marshal(map[string]any{
		"command": command,
	})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(commandJSON, '\n'))
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	util.Debugf("Raw response from mpv: %s", string(buffer[:n]))

	// Tratar múltiplos JSONs na mesma resposta
	responses := bytes.SplitSeq(buffer[:n], []byte("\n"))
	for resp := range responses {
		if len(bytes.TrimSpace(resp)) == 0 {
			continue
		}
		var response map[string]any
		err = json.Unmarshal(resp, &response)
		if err != nil {
			util.Debugf("Error when unmarshaling: %v", err)
			continue
		}
		if errStr, ok := response["error"].(string); ok && errStr == "property unavailable" {
			// Propriedade ainda não disponível, ignore sem erro
			util.Debugf("Property not yet available, ignoring...")
			continue
		}
		// Check for success response (set_property returns {"error":"success"} without data)
		if errStr, ok := response["error"].(string); ok && errStr == "success" {
			// Command succeeded, return nil data
			if data, exists := response["data"]; exists {
				return data, nil
			}
			return nil, nil
		}
		if data, exists := response["data"]; exists {
			return data, nil
		}
	}
	return nil, errors.New("no data field in mpv response")
}

// windows
// dialMPVSocket creates a connection to mpv's socket.
//func dialMPVSocket(socketPath string) (net.Conn, error) {
//	if runtime.GOOS == "windows" {
// Attempt named pipe on Windows
//		return net.Dial("unix", socketPath)
//	} else {
// Unix-like system uses Unix sockets
//		return net.Dial("unix", socketPath)
//	}
//}

// Funções de download extraídas de player.go
// downloadPart, combineParts, DownloadVideo, downloadWithYtDlp, ExtractVideoSources, getBestQualityURL, ExtractVideoSourcesWithPrompt, HandleBatchDownload, getEpisodeRange, findEpisode, createEpisodePath, fileExists
// As implementações completas estão agora em download.go

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(
	videoURL string,
	episodes []models.Episode,
	selectedEpisodeNum int,
	animeURL string,
	episodeNumberStr string,
	animeMalID int,
	updater *discord.RichPresenceUpdater,
	animeName string,
) error {
	util.Debug("HandleDownloadAndPlay called", "videoURL", videoURL, "episodeNum", selectedEpisodeNum)

	// Persist the anime URL/ID to aid episode switching when updater is nil (e.g., Discord disabled)
	lastAnimeURL = animeURL

	// Store anime name for Plex-compatible download file naming
	if animeName != "" {
		season := 1
		if util.GlobalDownloadRequest != nil && util.GlobalDownloadRequest.SeasonNum > 0 {
			season = util.GlobalDownloadRequest.SeasonNum
		}
		SetAnimeName(animeName, season)
	}

	// Check if this is an HLS stream (for proper handling later)
	isHLSStream := strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "m3u8")
	util.Debug("Stream type", "isHLS", isHLSStream)

	for {
		downloadOption := askForDownload()
		switch downloadOption {
		case 0:
			// User wants to go back to server selection
			return ErrBackToEpisodeSelection
		case 1:
			// Download the current episode
			if isHLSStream {
				// HLS streams need special download handling
				util.Debugf("HLS download requested - using stream URL")
			}
			err := downloadAndPlayEpisode(
				videoURL,
				episodes,
				selectedEpisodeNum,
				animeURL,
				episodeNumberStr,
				animeMalID,
				updater,
			)
			if err != nil {
				if errors.Is(err, ErrBackToDownloadOptions) {
					continue // Go back to download options menu
				}
				return err
			}
			return nil
		case 2:
			// Download episodes in a range
			if err := HandleBatchDownload(episodes, animeURL); err != nil {
				return err
			}
			return nil
		case 3:
			// Upscale video with Anime4K
			if err := handleUpscaleFromMenu(); err != nil {
				util.Errorf("Upscale error: %v", err)
			}
			continue // Return to menu after upscaling
		default:
			// Play online - determine the best approach based on URL type
			videoURLToPlay := ""

			if isHLSStream {
				// HLS streams are already resolved, play directly
				videoURLToPlay = videoURL
				if util.IsDebug {
					util.Debugf("HLS stream detected, playing directly: %s", videoURLToPlay)
				}
			} else if videoURL != "" && needsVideoExtraction(videoURL) {
				// Intermediate URL (e.g. animefire.io/video/) needs resolution
				// to obtain the final CDN video URL.
				if util.IsDebug {
					util.Debugf("Intermediate URL detected, resolving: %s", videoURL)
				}
				if resolved, err := extractActualVideoURL(videoURL); err == nil && resolved != "" {
					videoURLToPlay = resolved
				}
			} else if videoURL != "" && strings.HasPrefix(videoURL, "http") {
				// The enhanced API already resolved a direct stream URL (CDN,
				// mp4, etc.). Use it directly — re-extracting may trigger
				// duplicate quality prompts or cause CDN URLs to expire.
				videoURLToPlay = videoURL
				if util.IsDebug {
					util.Debugf("Using resolved stream URL directly: %s", videoURLToPlay)
				}
			} else if videoURL != "" {
				// Non-HTTP URL (e.g. episode ID). Try legacy extraction.
				if len(episodes) > 0 && selectedEpisodeNum > 0 {
					selectedEp, found := findEpisode(episodes, selectedEpisodeNum)
					if found {
						if util.IsDebug {
							util.Debugf("Extracting URL from episode page: %s", selectedEp.URL)
						}
						if url, err := ExtractVideoSourcesWithPrompt(selectedEp.URL); err == nil && url != "" {
							videoURLToPlay = url
						}
					}
				}
				// Fallback: try to extract from original videoURL
				if videoURLToPlay == "" {
					if util.IsDebug {
						util.Debugf("Fallback: extracting from original URL: %s", videoURL)
					}
					if url, err := ExtractVideoSourcesWithPrompt(videoURL); err == nil && url != "" {
						videoURLToPlay = url
					}
				}
			}

			// Final validation
			if videoURLToPlay == "" {
				util.Debugf("No valid video URL found")
				return fmt.Errorf("no valid video URL found")
			}

			if util.IsDebug {
				util.Debugf("Final video URL: %s", videoURLToPlay)
			}

			err := playVideo(
				videoURLToPlay,
				episodes,
				selectedEpisodeNum,
				animeMalID,
				updater,
			)
			if err != nil {
				if errors.Is(err, ErrBackToDownloadOptions) {
					continue // Go back to download options menu
				}
				return err
			}
			return nil
		}
	}
}

func downloadAndPlayEpisode(
	videoURL string,
	episodes []models.Episode,
	selectedEpisodeNum int,
	animeURL string,
	episodeNumberStr string,
	animeMalID int, // Added animeMalID parameter
	updater *discord.RichPresenceUpdater,
) error {
	// Check if video URL is valid
	if videoURL == "" {
		return fmt.Errorf("empty video URL provided for episode %s", episodeNumberStr)
	}

	// Resolve intermediate URLs (e.g., animefire.io/video/ JSON API) to actual CDN URLs
	if strings.Contains(videoURL, "animefire.io/video/") {
		util.Debug("Resolving AnimeFire video API URL before download", "url", videoURL)
		resolved, err := extractActualVideoURL(videoURL)
		if err != nil {
			return fmt.Errorf("failed to resolve AnimeFire video URL: %w", err)
		}
		if resolved != "" {
			util.Debug("Resolved AnimeFire URL", "resolved", resolved)
			videoURL = resolved
		}
	}

	currentUser, err := user.Current()
	if err != nil {
		util.Fatal("Failed to get current user:", err)
	}

	// Use Plex-compatible naming when anime name is available
	var downloadPath, episodePath string
	if lastAnimeName != "" {
		// Route to the correct base directory: movies/ for movies/TV, anime/ for anime
		var baseDir string
		if lastIsMovieOrTV {
			baseDir = util.DefaultMovieDownloadDir()
		} else {
			baseDir = util.DefaultDownloadDir()
		}

		// Check if this is a standalone movie (flat path) vs TV/anime (season structure)
		if IsCurrentMediaMovie() {
			// Movies: flat structure <baseDir>/<MovieName>/
			downloadPath = util.FormatPlexMovieDir(baseDir, lastAnimeName)
			episodePath = util.FormatPlexMoviePath(baseDir, lastAnimeName, "")
		} else {
			// TV Shows and Anime: season/episode structure
			season := lastAnimeSeason
			if season < 1 {
				season = 1
			}
			// Use the int episode number directly; fall back to parsing the string only if needed
			epNum := selectedEpisodeNum
			if epNum < 1 {
				parsed, _ := strconv.Atoi(episodeNumberStr)
				if parsed > 0 {
					epNum = parsed
				} else {
					epNum = 1
				}
			}
			downloadPath = util.FormatPlexEpisodeDir(baseDir, lastAnimeName, season)
			episodePath = util.FormatPlexEpisodePath(baseDir, lastAnimeName, season, epNum)
		}
		util.Debugf("Download routing: mediaType=%s, isMovieOrTV=%v, baseDir=%s, path=%s", lastMediaType, lastIsMovieOrTV, baseDir, episodePath)
	} else {
		// Fallback: route based on media type even without anime name
		var fallbackBase string
		if lastIsMovieOrTV {
			fallbackBase = util.DefaultMovieDownloadDir()
		} else {
			fallbackBase = filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime")
		}
		downloadPath = filepath.Join(fallbackBase, DownloadFolderFormatter(animeURL))
		episodePath = filepath.Join(downloadPath, episodeNumberStr+".mp4")
	}

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadPath, 0700); err != nil {
			util.Fatal("Failed to create download directory:", err)
		}
	}

	// Prompt user to select subtitle language BEFORE download starts
	// (stdin is free here — no Bubble Tea running yet)
	// For 9Anime, ALWAYS use the mandatory language prompt regardless of track count.
	if util.Is9AnimeSource() {
		util.PromptSubtitleLanguage()
	} else if len(util.GlobalSubtitles) > 0 {
		util.SelectSubtitles()
	}

	if _, err := os.Stat(episodePath); os.IsNotExist(err) {
		numThreads := 4 // Define the number of threads for downloading

		// Check URL type and use appropriate download method
		if strings.Contains(videoURL, "blogger.com") ||
			strings.Contains(videoURL, ".m3u8") ||
			strings.Contains(videoURL, "wixmp.com") ||
			strings.Contains(videoURL, "sharepoint.com") {
			// Use yt-dlp with progress bar
			m := &model{
				progress: progress.New(progress.WithDefaultGradient()),
				keys: keyMap{
					quit: key.NewBinding(
						key.WithKeys("ctrl+c"),
						key.WithHelp("ctrl+c", "quit"),
					),
				},
			}
			p := tea.NewProgram(m)

			// Estimate/obtain total size for progress percentage
			httpClient := &http.Client{Transport: api.SafeTransport(10 * time.Second)}
			if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
				m.totalBytes = sz
			} else {
				// Fallback for HLS
				m.totalBytes = 500 * 1024 * 1024
			}

			go func() {
				p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.jpg, .png) and "live" HLS (no #EXT-X-ENDLIST) that break yt-dlp.
				// yt-dlp is only used for non-HLS streams.
				var dlErr error
				if strings.Contains(videoURL, ".m3u8") {
					dlErr = downloadWithNativeHLS(videoURL, episodePath, m)
					if dlErr != nil {
						util.Logger.Warn("Native HLS failed, falling back to yt-dlp", "error", dlErr)
						dlErr = downloadWithYtDlp(videoURL, episodePath, m)
					}
				} else {
					dlErr = downloadWithYtDlp(videoURL, episodePath, m)
				}
				if dlErr != nil {
					util.Fatal("Failed to download video:", dlErr)
				}
				// Update progress to reflect real file size so bar shows accurate 100%
				if fi, statErr := os.Stat(episodePath); statErr == nil && fi.Size() > 0 {
					m.mu.Lock()
					m.totalBytes = fi.Size()
					m.received = fi.Size()
					m.mu.Unlock()
				}
				m.mu.Lock()
				m.done = true
				m.mu.Unlock()
				p.Send(statusMsg("Download completed!"))
			}()

			if _, err := p.Run(); err != nil {
				util.Fatal("Error running progress bar:", err)
			}

			// Verify the file was actually downloaded
			if _, err := os.Stat(episodePath); os.IsNotExist(err) {
				return fmt.Errorf("download failed: file was not created")
			}

			// Verify the file is a reasonable size for a video episode.
			// HLS episodes are typically at least 20 MB; anything below 10 MB
			// almost certainly indicates a truncated or failed download.
			const minEpisodeSize int64 = 10 * 1024 * 1024 // 10 MB
			if stat, err := os.Stat(episodePath); err == nil && stat.Size() < minEpisodeSize {
				_ = os.Remove(episodePath) // remove partial file so retry won't skip it
				return fmt.Errorf("download incomplete: file is only %d bytes (%.1f MB), expected at least %.0f MB",
					stat.Size(), float64(stat.Size())/(1024*1024), float64(minEpisodeSize)/(1024*1024))
			}

			fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)

			// Download selected subtitles alongside the video file
			downloadSubtitleFiles(episodePath)

		} else {
			// Initialize progress model
			m := &model{
				progress: progress.New(progress.WithDefaultGradient()),
				keys: keyMap{
					quit: key.NewBinding(
						key.WithKeys("ctrl+c"),
						key.WithHelp("ctrl+c", "quit"),
					),
				},
			}
			p := tea.NewProgram(m)

			// Get content length
			httpClient := &http.Client{
				Transport: api.SafeTransport(10 * time.Second),
			}
			contentLength, err := getContentLength(videoURL, httpClient)
			if err != nil {
				util.Warnf("Failed to get content length: %v, using fallback estimate", err)
				contentLength = 200 * 1024 * 1024 // 200MB fallback
			}
			m.totalBytes = contentLength

			// Start the download in a separate goroutine
			go func() {
				// Update status
				p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))

				if err := DownloadVideo(videoURL, episodePath, numThreads, m); err != nil {
					util.Fatal("Failed to download video:", err)
				}

				m.mu.Lock()
				m.done = true
				m.mu.Unlock()

				// Final status update
				p.Send(statusMsg("Download completed!"))
			}()

			// Run the Bubble Tea program in the main goroutine
			if _, err := p.Run(); err != nil {
				util.Fatal("Error running progress bar:", err)
			}

			// Download selected subtitles alongside the video file
			downloadSubtitleFiles(episodePath)
		}
	} else {
		fmt.Println("Video already downloaded.")
		// Check if the file is actually valid (not empty)
		if stat, err := os.Stat(episodePath); err == nil {
			if stat.Size() < 1024 {
				fmt.Println("File is too small, re-downloading...")
				if removeErr := os.Remove(episodePath); removeErr != nil {
					util.Warnf("Failed to remove invalid file: %v", removeErr)
				}
				return downloadAndPlayEpisode(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr, animeMalID, updater)
			}
		}
	}

	if askForPlayOffline() {
		if err := playVideo(episodePath, episodes, selectedEpisodeNum, animeMalID, updater); err != nil {
			return err
		}
		return nil
	}
	// User chose not to watch; terminate flow cleanly
	return ErrUserQuit
}

// askForDownload presents a prompt for the user to choose a download option.
func askForDownload() int {
	var choice string

	// Build the upscale option label with current status
	upscaleStatus := upscaler.GetShaderModeName(upscaler.CurrentShaderMode)
	upscaleLabel := fmt.Sprintf("Real-time Upscale [%s]", upscaleStatus)

	menu := huh.NewSelect[string]().
		Title("Download Options").
		Description("Choose how you want to proceed:").
		Options(
			huh.NewOption("← Back", "back"),
			huh.NewOption("Download this episode", "download_single"),
			huh.NewOption("Download episodes in a range", "download_range"),
			huh.NewOption(upscaleLabel, "upscale"),
			huh.NewOption("No download (play online)", "play_online"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing download menu: %v", err)
		return 4 // Default to play online on error
	}

	// Determines the selected option based on the choice value
	switch choice {
	case "back":
		return 0
	case "download_single":
		return 1
	case "download_range":
		return 2
	case "upscale":
		return 3
	default:
		return 4
	}
}

func askForPlayOffline() bool {
	var choice string

	menu := huh.NewSelect[string]().
		Title("Offline Playback").
		Description("Do you want to play the downloaded version offline?").
		Options(
			huh.NewOption("Yes", "yes"),
			huh.NewOption("No", "no"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing offline playback menu: %v", err)
		return false // Default to no on error
	}

	return choice == "yes"
}

// playVideo has been moved to playvideo.go

// ToggleSubtitle toggles subtitle visibility
func ToggleSubtitle(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"sub-visibility",
	})
	return err
}

// GetPlaybackStats returns current playback statistics
func GetPlaybackStats(socketPath string) (map[string]any, error) {
	stats := make(map[string]any)

	// Get various playback properties
	properties := []string{
		"time-pos",
		"duration",
		"speed",
		"volume",
		"pause",
		"filename",
	}

	for _, prop := range properties {
		value, err := mpvSendCommand(socketPath, []any{"get_property", prop})
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", prop, err)
		}
		stats[prop] = value
	}

	return stats, nil
}

// SetPlaybackSpeed sets the video playback speed
func SetPlaybackSpeed(socketPath string, speed float64) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"speed",
		speed,
	})
	return err
}

// CycleAudioTrack cycles through available audio tracks
func CycleAudioTrack(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"aid",
	})
	return err
}

// CycleSubtitleTrack cycles through available subtitle tracks
func CycleSubtitleTrack(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"sid",
	})
	return err
}

// SetAudioTrack sets a specific audio track by ID
func SetAudioTrack(socketPath string, trackID int) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"aid",
		trackID,
	})
	return err
}

// SetSubtitleTrack sets a specific subtitle track by ID
func SetSubtitleTrack(socketPath string, trackID int) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"sid",
		trackID,
	})
	return err
}

// GetAudioTracks returns list of available audio tracks
func GetAudioTracks(socketPath string) ([]map[string]any, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "track-list"})
	if err != nil {
		return nil, err
	}

	tracks, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected track-list format")
	}

	var audioTracks []map[string]any
	for _, t := range tracks {
		track, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if trackType, ok := track["type"].(string); ok && trackType == "audio" {
			audioTracks = append(audioTracks, track)
		}
	}
	return audioTracks, nil
}

// GetSubtitleTracks returns list of available subtitle tracks
func GetSubtitleTracks(socketPath string) ([]map[string]any, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "track-list"})
	if err != nil {
		return nil, err
	}

	tracks, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected track-list format")
	}

	var subTracks []map[string]any
	for _, t := range tracks {
		track, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if trackType, ok := track["type"].(string); ok && trackType == "sub" {
			subTracks = append(subTracks, track)
		}
	}
	return subTracks, nil
}

// GetCurrentAudioTrack returns the current audio track ID
func GetCurrentAudioTrack(socketPath string) (int, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "aid"})
	if err != nil {
		return 0, err
	}
	if id, ok := result.(float64); ok {
		return int(id), nil
	}
	return 0, fmt.Errorf("unexpected aid format")
}

// GetCurrentSubtitleTrack returns the current subtitle track ID
func GetCurrentSubtitleTrack(socketPath string) (int, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "sid"})
	if err != nil {
		return 0, err
	}
	if id, ok := result.(float64); ok {
		return int(id), nil
	}
	// "no" means no subtitle is selected
	if _, ok := result.(string); ok {
		return 0, nil
	}
	return 0, fmt.Errorf("unexpected sid format")
}

// handleUpscaleFromMenu shows the real-time upscaling options menu
func handleUpscaleFromMenu() error {
	var choice string

	// Check if shaders are installed
	shadersInstalled := upscaler.ShadersInstalled()
	currentMode := upscaler.GetShaderModeName(upscaler.CurrentShaderMode)

	description := fmt.Sprintf("Current: %s", currentMode)
	if !shadersInstalled {
		description = "Shaders not installed - select 'Setup shaders' first"
	}

	options := []huh.Option[string]{
		huh.NewOption("Back", "back"),
		huh.NewOption("Off (no upscaling)", "off"),
	}

	if shadersInstalled {
		options = append(options,
			huh.NewOption("Performance (weak GPU)", "performance"),
			huh.NewOption("Fast (Mode A - text-heavy)", "fast"),
			huh.NewOption("Balanced (Mode B - general)", "balanced"),
			huh.NewOption("Quality (Mode C - films)", "quality"),
			huh.NewOption("Ultra (Max Enhancement - SD sources)", "ultra"),
			huh.NewOption("--- Advanced Modes ---", "separator"),
			huh.NewOption("A+A (Max Perceptual - 1080p)", "advanced_aa"),
			huh.NewOption("B+B (720p Optimized)", "advanced_bb"),
			huh.NewOption("C+A (Upscaled/Downscaled Content)", "advanced_ca"),
		)

		// Add GAN UUL option (check if GAN shaders are installed)
		if upscaler.GANShadersInstalled() {
			options = append(options,
				huh.NewOption("GAN UUL (360p to 4K - HEAVY)", "gan_uul"),
			)
		} else {
			options = append(options,
				huh.NewOption("GAN UUL (not installed)", "setup_gan"),
			)
		}
	}

	options = append(options, huh.NewOption("Setup shaders (download)", "setup"))

	menu := huh.NewSelect[string]().
		Title("Real-time Anime4K Upscaling").
		Description(description).
		Options(options...).
		Value(&choice)

	if err := menu.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	switch choice {
	case "back":
		return nil
	case "separator":
		// Do nothing for separator, show menu again
		return handleUpscaleFromMenu()
	case "off":
		upscaler.SetShaderMode(upscaler.ShaderModeOff)
		util.Info("Real-time upscaling disabled")
	case "performance":
		upscaler.SetShaderMode(upscaler.ShaderModePerformance)
		util.Info("Real-time upscaling: Performance mode (minimal shaders)")
	case "fast":
		upscaler.SetShaderMode(upscaler.ShaderModeFast)
		util.Info("Real-time upscaling: Fast mode (Mode A - good for subtitled anime)")
	case "balanced":
		upscaler.SetShaderMode(upscaler.ShaderModeBalanced)
		util.Info("Real-time upscaling: Balanced mode (Mode B - general purpose)")
	case "quality":
		upscaler.SetShaderMode(upscaler.ShaderModeQuality)
		util.Info("Real-time upscaling: Quality mode (Mode C - best for films)")
	case "ultra":
		upscaler.SetShaderMode(upscaler.ShaderModeUltra)
		util.Info("Real-time upscaling: Ultra mode (Maximum enhancement for SD sources)")
	case "advanced_aa":
		upscaler.SetShaderMode(upscaler.ShaderModeAdvancedAA)
		util.Info("Real-time upscaling: Advanced A+A (highest perceptual quality, may cause ringing)")
	case "advanced_bb":
		upscaler.SetShaderMode(upscaler.ShaderModeAdvancedBB)
		util.Info("Real-time upscaling: Advanced B+B (optimized for 720p with aliasing)")
	case "advanced_ca":
		upscaler.SetShaderMode(upscaler.ShaderModeAdvancedCA)
		util.Info("Real-time upscaling: Advanced C+A (quality + restore for downscaled content)")
	case "gan_uul":
		upscaler.SetShaderMode(upscaler.ShaderModeGAN_UUL)
		util.Info("Real-time upscaling: GAN UUL mode (360p→4K - requires powerful GPU!)")
		util.Warn("This mode is VERY heavy! If you experience lag, switch to a lighter mode.")
	case "setup_gan":
		util.Info("Setting up experimental GAN UUL shaders...")
		if err := upscaler.InstallGANShaders(); err != nil {
			return fmt.Errorf("failed to install GAN shaders: %w", err)
		}
		util.Info("GAN UUL shaders installed! Select 'GAN UUL' to enable 360p→4K upscaling.")
	case "setup":
		util.Info("Setting up Anime4K shaders...")
		if err := upscaler.InstallShaders(); err != nil {
			return fmt.Errorf("failed to install shaders: %w", err)
		}
		util.Info("Anime4K shaders installed! Select a mode to enable upscaling.")
	}

	return nil
}

// downloadSubtitleFiles downloads the user-selected subtitle tracks alongside
// the downloaded video file. Uses the subtitles stored in util.GlobalSubtitles
// (already filtered by util.SelectSubtitles).
func downloadSubtitleFiles(videoPath string) {
	subs := util.GlobalSubtitles
	if len(subs) == 0 {
		return
	}

	// Check ffmpeg availability — required for muxing subtitles into the video
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		util.Warnf("ffmpeg not found — cannot embed subtitles into the video file")
		return
	}
	// Resolve symlinks and validate the binary path to prevent PATH-based injection.
	ffmpegPath, err = filepath.EvalSymlinks(ffmpegPath)
	if err != nil {
		util.Warnf("failed to resolve ffmpeg path: %v", err)
		return
	}
	if !filepath.IsAbs(ffmpegPath) {
		util.Warnf("ffmpeg resolved to a non-absolute path — refusing to execute")
		return
	}
	if fi, statErr := os.Stat(ffmpegPath); statErr != nil || fi.IsDir() {
		util.Warnf("ffmpeg path is not a valid file: %s", ffmpegPath)
		return
	}

	dir := filepath.Dir(videoPath)
	client := &http.Client{
		Transport: api.SafeTransport(30 * time.Second),
		Timeout:   60 * time.Second,
	}

	// Collect subtitle files to mux
	type subEntry struct {
		tmpPath  string
		label    string
		langCode string
	}
	var entries []subEntry

	for _, sub := range subs {
		if sub.URL == "" {
			continue
		}

		// Determine extension
		ext := "vtt"
		lower := strings.ToLower(sub.URL)
		if strings.Contains(lower, ".srt") {
			ext = "srt"
		} else if strings.Contains(lower, ".ass") {
			ext = "ass"
		}

		lang := util.SanitizeForFilename(sub.Label)
		if lang == "" {
			lang = util.SanitizeForFilename(sub.Language)
		}
		if lang == "" {
			lang = "unknown"
		}

		// Download to a temp file
		tmpPath := filepath.Join(dir, fmt.Sprintf(".tmp_sub_%s.%s", lang, ext))
		req, reqErr := http.NewRequest("GET", sub.URL, nil)
		if reqErr != nil {
			util.Warnf("Failed to create subtitle request (%s): %v", sub.Label, reqErr)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, respErr := client.Do(req) // #nosec G107 G704
		if respErr != nil {
			util.Warnf("Failed to download subtitle (%s): %v", sub.Label, respErr)
			continue
		}

		func() {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				util.Warnf("Subtitle download failed (%s): HTTP %d", sub.Label, resp.StatusCode)
				return
			}
			out, oErr := os.Create(filepath.Clean(tmpPath))
			if oErr != nil {
				util.Warnf("Failed to create temp subtitle file (%s): %v", sub.Label, oErr)
				return
			}
			defer func() { _ = out.Close() }()
			if _, cpErr := io.Copy(out, resp.Body); cpErr != nil {
				util.Warnf("Failed to write subtitle (%s): %v", sub.Label, cpErr)
				return
			}
			langCode := sub.Language
			if langCode == "" {
				langCode = lang
			}
			entries = append(entries, subEntry{tmpPath: tmpPath, label: sub.Label, langCode: langCode})
		}()
	}

	if len(entries) == 0 {
		return
	}

	// --- Mux subtitles into the video container ---
	fmt.Printf("Embedding %d subtitle(s) into video...\n", len(entries))

	// buildMuxArgs builds the ffmpeg arguments for a given subtitle codec and output path.
	buildMuxArgs := func(subCodec, outPath string) []string {
		a := []string{"-y", "-fflags", "+genpts", "-i", filepath.Clean(videoPath)}
		for _, e := range entries {
			a = append(a, "-i", filepath.Clean(e.tmpPath))
		}
		// Map only video and audio from input — skip data streams like timed_id3
		// which are present in MPEG-TS from HLS downloads and crash MP4/MKV muxing.
		a = append(a, "-map", "0:v", "-map", "0:a")
		for i := range entries {
			a = append(a, "-map", fmt.Sprintf("%d", i+1))
		}
		a = append(a, "-c:v", "copy", "-c:a", "copy", "-c:s", subCodec)
		for i, e := range entries {
			a = append(a, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", e.langCode))
			a = append(a, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", e.label))
		}
		a = append(a, filepath.Clean(outPath))
		return a
	}

	// runMux executes ffmpeg and captures stderr for diagnostics.
	runMux := func(subCodec, outPath string) error {
		args := buildMuxArgs(subCodec, outPath)
		util.Debugf("ffmpeg mux cmd: %s %v", ffmpegPath, args)
		cmd := exec.Command(ffmpegPath, args...) // #nosec G204 -- ffmpegPath is validated: resolved via EvalSymlinks, confirmed absolute and a regular file
		var stderrBuf bytes.Buffer
		cmd.Stdout = nil
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			util.Debugf("ffmpeg mux failed: %v\nstderr: %s", err, stderrBuf.String())
			_ = os.Remove(outPath)
			return err
		}
		return nil
	}

	embedded := false

	// Attempt 1: MP4 container with mov_text subtitle codec
	tmpMP4 := videoPath + ".muxing.mp4"
	if err := runMux("mov_text", tmpMP4); err == nil {
		if renErr := os.Rename(tmpMP4, videoPath); renErr != nil {
			util.Warnf("Failed to replace video: %v", renErr)
			_ = os.Remove(tmpMP4)
		} else {
			embedded = true
		}
	} else {
		util.Debugf("MP4 mux failed: %v — trying MKV fallback", err)
	}

	// Attempt 2: MKV container (more tolerant of various subtitle formats / TS inputs)
	if !embedded {
		mkvPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".mkv"
		tmpMKV := mkvPath + ".tmp.mkv"
		if err := runMux("srt", tmpMKV); err == nil {
			if renErr := os.Rename(tmpMKV, mkvPath); renErr != nil {
				util.Warnf("Failed to save MKV: %v", renErr)
				_ = os.Remove(tmpMKV)
			} else {
				embedded = true
				fmt.Printf("Note: saved as .mkv for better subtitle compatibility\n")
			}
		}
	}

	if embedded {
		fmt.Printf("Subtitles embedded successfully!\n")
	} else {
		util.Warnf("Could not embed subtitles — both MP4 and MKV muxing failed")
	}

	// Clean up temp subtitle files
	for _, e := range entries {
		_ = os.Remove(e.tmpPath)
	}
}
