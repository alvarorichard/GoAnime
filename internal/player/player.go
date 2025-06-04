package player

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/lrstanley/go-ytdlp"
	"github.com/pkg/errors"
)

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
	// Verify MPV is installed
	if _, err := exec.LookPath("mpv"); err != nil {
		return "", fmt.Errorf("mpv not found in PATH. Please install mpv: https://mpv.io/installation/")
	}

	randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
	var socketPath string

	if runtime.GOOS == "windows" {
		socketPath = fmt.Sprintf(`\\.\pipe\goanime_mpvsocket_%s`, randomNumber)
	} else {
		tmpDir := "/tmp"
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create tmp directory: %w", err)
		}
		socketPath = fmt.Sprintf("%s/goanime_mpvsocket_%s", tmpDir, randomNumber)
	}

	mpvArgs := []string{
		"--no-terminal",
		"--quiet",
		fmt.Sprintf("--input-ipc-server=%s", socketPath),
	}
	mpvArgs = append(mpvArgs, args...)
	mpvArgs = append(mpvArgs, link)

	if util.IsDebug {
		fmt.Printf("[DEBUG] Starting mpv with arguments: %v\n", mpvArgs)
	}

	cmd := exec.Command("mpv", mpvArgs...)
	setProcessGroup(cmd) // Handle OS-specific process groups

	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mpv: %w (stderr: %s)", err, stderr.String())
	}

	if util.IsDebug {
		fmt.Printf("[DEBUG]mpv started, waiting for socket creation: %s\n", socketPath)
	}

	// Wait for socket creation with longer timeout
	maxAttempts := 30 // 3 seconds total
	for i := 0; i < maxAttempts; i++ {
		if util.IsDebug {
			fmt.Printf("[DEBUG] Try %d/%d: checking existence of socket...\n", i+1, maxAttempts)
		}
		if runtime.GOOS == "windows" {
			// Special handling for Windows named pipes
			_, err := os.Stat(`\\.\pipe\` + strings.TrimPrefix(socketPath, `\\.\pipe\`))
			if err == nil {
				if util.IsDebug {
					fmt.Printf("[DEBUG] Socket found after %.2fs\n", time.Since(startTime).Seconds())
				}
				return socketPath, nil
			}
		} else {
			if _, err := os.Stat(socketPath); err == nil {
				if util.IsDebug {
					fmt.Printf("[DEBUG] Socket found after %.2fs\n", time.Since(startTime).Seconds())
				}
				return socketPath, nil
			}
		}

		// Check if MPV process is still running
		if cmd.Process == nil || cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			return "", fmt.Errorf("mpv process exited prematurely: %s", stderr.String())
		}

		time.Sleep(100 * time.Millisecond)
	}

	if util.IsDebug {
		fmt.Printf("[DEBUG] Timeout after %.2fs waiting  socket of mpv\n", time.Since(startTime).Seconds())
	}
	// Cleanup if timeout occurs
	err := cmd.Process.Kill()
	if err != nil {

		return "", err
	}
	return "", fmt.Errorf("timeout waiting for mpv socket. Possible issues:\n1. MPV installation corrupted\n2. Firewall blocking IPC\n3. Invalid video URL\nCheck debug logs with -debug flag")
}

// MpvSendCommand is a wrapper function to expose mpvSendCommand to other packages
func MpvSendCommand(socketPath string, command []interface{}) (interface{}, error) {
	return mpvSendCommand(socketPath, command)
}

// mpvSendCommand sends a JSON command to MPV via the IPC socket and receives the response.
func mpvSendCommand(socketPath string, command []interface{}) (interface{}, error) {
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

	commandJSON, err := json.Marshal(map[string]interface{}{
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

	if util.IsDebug {
		fmt.Printf("[DEBUG]Raw response from mpv: %s\n", string(buffer[:n]))
	}

	// Tratar múltiplos JSONs na mesma resposta
	responses := bytes.Split(buffer[:n], []byte("\n"))
	for _, resp := range responses {
		if len(bytes.TrimSpace(resp)) == 0 {
			continue
		}
		var response map[string]interface{}
		err = json.Unmarshal(resp, &response)
		if err != nil {
			if util.IsDebug {
				fmt.Printf("[DEBUG]Error when unmarshaling: %v\n", err)
			}
			continue
		}
		if errStr, ok := response["error"].(string); ok && errStr == "property unavailable" {
			// Propriedade ainda não disponível, ignore sem erro
			if util.IsDebug {
				fmt.Println("[DEBUG] Property not yet available, ignoring...")
			}
			continue
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
) error {
	downloadOption := askForDownload()
	switch downloadOption {
	case 1:
		// Download the current episode
		if err := downloadAndPlayEpisode(
			videoURL,
			episodes,
			selectedEpisodeNum,
			animeURL,
			episodeNumberStr,
			animeMalID,
			updater,
		); err != nil {
			return err
		}
	case 2:
		// Download episodes in a range
		if err := HandleBatchDownload(episodes, animeURL); err != nil {
			util.Fatal("Failed to download episodes:", err)
		}
	default:
		// Play online
		videoURLToPlay := ""
		// Always use the episode page URL for streaming, so the user can select quality
		if len(episodes) > 0 && selectedEpisodeNum > 0 {
			// Find the selected episode struct
			selectedEp, found := findEpisode(episodes, selectedEpisodeNum)
			if found {
				if url, err := ExtractVideoSourcesWithPrompt(selectedEp.URL); err == nil {
					videoURLToPlay = url
				}
			}
		}
		if videoURLToPlay == "" {
			// fallback: try original videoURL
			if url, err := ExtractVideoSourcesWithPrompt(videoURL); err == nil {
				videoURLToPlay = url
			}
		}
		if err := playVideo(
			videoURLToPlay,
			episodes,
			selectedEpisodeNum,
			animeMalID,
			updater,
		); err != nil {
			return err
		}
	}
	return nil
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
	currentUser, err := user.Current()
	if err != nil {
		util.Fatal("Failed to get current user:", err)
	}

	downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
	episodePath := filepath.Join(downloadPath, episodeNumberStr+".mp4")

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
			util.Fatal("Failed to create download directory:", err)
		}
	}

	if _, err := os.Stat(episodePath); os.IsNotExist(err) {
		numThreads := 4 // Define the number of threads for downloading

		// Check if the video URL is from Blogger
		if strings.Contains(videoURL, "blogger.com") {
			// Use yt-dlp to download the video from Blogger
			fmt.Printf("Downloading episode %s with yt-dlp...\n", episodeNumberStr)

			// Ensure yt-dlp is installed
			ytdlp.MustInstall(context.Background(), nil)

			// Configure downloader
			dl := ytdlp.New().
				//Quiet(true).          // --no-progress
				Output(episodePath) // -o <episodePath>

			// Execute download
			if _, err := dl.Run(context.Background(), videoURL); err != nil {
				util.Error("Failed to download video using yt-dlp:", err)
			}
			fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)

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
				util.Fatal("Failed to get content length:", err)
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
		}
	} else {
		fmt.Println("Video already downloaded.")
	}

	if askForPlayOffline() {
		if err := playVideo(episodePath, episodes, selectedEpisodeNum, animeMalID, updater); err != nil {
			return err
		}
	}
	return nil
}

// askForDownload presents a prompt for the user to choose a download option.
func askForDownload() int {
	var choice string

	menu := huh.NewSelect[string]().
		Title("Download Options").
		Description("Choose how you want to proceed:").
		Options(
			huh.NewOption("Download this episode", "download_single"),
			huh.NewOption("Download episodes in a range", "download_range"),
			huh.NewOption("No download (play online)", "play_online"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing download menu: %v", err)
		return 3 // Default to play online on error
	}

	// Determines the selected option based on the choice value
	switch choice {
	case "download_single":
		return 1
	case "download_range":
		return 2
	default:
		return 3
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
	_, err := mpvSendCommand(socketPath, []interface{}{
		"cycle",
		"sub-visibility",
	})
	return err
}

// GetPlaybackStats returns current playback statistics
func GetPlaybackStats(socketPath string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

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
		value, err := mpvSendCommand(socketPath, []interface{}{"get_property", prop})
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", prop, err)
		}
		stats[prop] = value
	}

	return stats, nil
}

// SetPlaybackSpeed sets the video playback speed
func SetPlaybackSpeed(socketPath string, speed float64) error {
	_, err := mpvSendCommand(socketPath, []interface{}{
		"set_property",
		"speed",
		speed,
	})
	return err
}
