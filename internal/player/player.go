package player

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/go-ytdlp"
	"github.com/manifoldco/promptui"
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
		fmt.Printf("Starting mpv with arguments: %v\n", mpvArgs)
	}

	cmd := exec.Command("mpv", mpvArgs...)
	setProcessGroup(cmd) // Handle OS-specific process groups

	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mpv: %w (stderr: %s)", err, stderr.String())
	}

	// Wait for socket creation with longer timeout
	maxAttempts := 30 // 3 seconds total
	for i := 0; i < maxAttempts; i++ {
		if runtime.GOOS == "windows" {
			// Special handling for Windows named pipes
			_, err := os.Stat(`\\.\pipe\` + strings.TrimPrefix(socketPath, `\\.\pipe\`))
			if err == nil {
				return socketPath, nil
			}
		} else {
			if _, err := os.Stat(socketPath); err == nil {
				return socketPath, nil
			}
		}

		// Check if MPV process is still running
		if cmd.Process == nil || cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			return "", fmt.Errorf("mpv process exited prematurely: %s", stderr.String())
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Cleanup if timeout occurs
	cmd.Process.Kill()
	return "", fmt.Errorf("timeout waiting for mpv socket. Possible issues:\n1. MPV installation corrupted\n2. Firewall blocking IPC\n3. Invalid video URL\nCheck debug logs with -debug flag")
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

	var response map[string]interface{}
	err = json.Unmarshal(buffer[:n], &response)
	if err != nil {
		return nil, err
	}

	if data, exists := response["data"]; exists {
		return data, nil
	}
	return nil, errors.New("no data field in mpv response")
}

// dialMPVSocket creates a connection to mpv's socket.
func dialMPVSocket(socketPath string) (net.Conn, error) {
	if runtime.GOOS == "windows" {
		// Attempt named pipe on Windows
		return net.Dial("unix", socketPath)
	} else {
		// Unix-like system uses Unix sockets
		return net.Dial("unix", socketPath)
	}
}

// downloadPart downloads a part of the video file.
//
// This function downloads a specific part (or chunk) of a video file using HTTP ranged requests.
// It saves the downloaded part as a temporary file and updates the progress state as data is received.
//
// Parameters:
// - url: The URL of the video file to download.
// - from: The starting byte of the file part to download.
// - to: The ending byte of the file part to download.
// - part: The part number, used to name the temporary file.
// - client: The HTTP client used to make the request.
// - destPath: The destination path where the downloaded file part will be saved.
// - m: The model containing the progress and state information.
//
// Returns:
// - An error if the download fails, or nil if it succeeds.
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model) error {
	// Creates a new HTTP GET request for the specified URL.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Returns the error if the request creation fails.
		return err
	}

	// Adds a "Range" header to specify the byte range to download (from 'from' to 'to').
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))

	// Sends the HTTP request using the provided client.
	resp, err := client.Do(req)
	if err != nil {
		// Returns the error if the request fails.
		return err
	}

	// Ensures that the response body is closed after the function finishes to avoid resource leaks.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// Constructs the file name and path for the current part (e.g., video.mp4.part0).
	partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), part)
	partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)

	// Creates a new file to store the downloaded part.
	file, err := os.Create(partFilePath)
	if err != nil {
		// Returns the error if file creation fails.
		return err
	}

	// Ensures that the file is closed properly after writing the data.
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close file: %v\n", err)
		}
	}(file)

	// Creates a buffer of 32 KB to read the response data in chunks.
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		// Reads data from the response body into the buffer.
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// If data is read, write it to the file.
			if _, err := file.Write(buf[:n]); err != nil {
				// Returns the error if writing to the file fails.
				return err
			}

			// Updates the received byte count in the model.
			m.mu.Lock()
			m.received += int64(n) // Updates the progress with the number of bytes received.
			m.mu.Unlock()
		}

		// If EOF is reached (end of file), the download for this part is complete.
		if err == io.EOF {
			break
		}

		// If another error occurs during reading, return the error.
		if err != nil {
			return err
		}
	}

	// Returns nil if the download part completes successfully.
	return nil
}

// combineParts combines downloaded parts into a single file.
//
// This function merges multiple downloaded parts of a file into one complete file. Each part is saved
// as a temporary file (e.g., video.mp4.part0, video.mp4.part1) and is combined sequentially into the
// final destination file. After merging, the temporary part files are deleted.
//
// Parameters:
// - destPath: The path where the final combined file will be saved.
// - numThreads: The number of parts (or threads) that were used to download the file.
//
// Returns:
// - An error if the merging process fails, or nil if successful.
func combineParts(destPath string, numThreads int) error {
	// Creates the final output file where all parts will be merged.
	outFile, err := os.Create(destPath)
	if err != nil {
		// Returns an error if the final file cannot be created.
		return err
	}

	// Ensures that the output file is closed after all parts are written.
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			// Logs an error if closing the output file fails.
			log.Printf("Failed to close output file: %v\n", err)
		}
	}(outFile)

	// Loops through each part that was downloaded.
	for i := 0; i < numThreads; i++ {
		// Constructs the file name for the current part (e.g., video.mp4.part0).
		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
		// Builds the full path to the part file.
		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)

		// Opens the part file for reading.
		partFile, err := os.Open(partFilePath)
		if err != nil {
			// Logs an error and returns it if the part file cannot be opened.
			fmt.Println("Failed to open part file:", err)
			return err
		}

		// Copies the contents of the part file into the final output file.
		if _, err := io.Copy(outFile, partFile); err != nil {
			// If copying fails, ensures the part file is closed before returning an error.
			err := partFile.Close()
			if err != nil {
				fmt.Printf("Failed to close part file: %v\n", err)
				return err
			}
			// Returns the error if the copy operation fails.
			return err
		}

		// Closes the part file after it has been copied to the final file.
		err = partFile.Close()
		if err != nil {
			// Logs an error if closing the part file fails.
			fmt.Printf("Failed to close part file: %v\n", err)
			return err
		}

		// Deletes the part file after it has been successfully copied and closed.
		if err := os.Remove(partFilePath); err != nil {
			// Returns an error if the part file cannot be deleted.
			return err
		}
	}

	// Returns nil to indicate success after all parts are combined and deleted.
	return nil
}

// DownloadVideo downloads a video using multiple threads.
//
// This function downloads a video file in parallel using multiple threads. It divides the file
// into chunks, downloads each chunk concurrently, and then combines the parts into a single file
// at the destination path.
//
// Parameters:
// - url: The URL of the video file to download.
// - destPath: The destination path where the video file will be saved.
// - numThreads: The number of threads (or parts) to use for downloading the video.
// - m: The model used to track the progress and status of the download.
//
// Returns:
// - An error if the download or combination of parts fails, or nil if successful.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	// Cleans the destination path to ensure it is valid and well-formed.
	destPath = filepath.Clean(destPath)

	// Creates an HTTP client with custom transport that includes a 10-second timeout.
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}

	chunkSize := int64(0)   // Variable to store the size of each download chunk.
	var contentLength int64 // Variable to store the total content length of the file.

	// Retrieves the content length of the file from the URL.
	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		// Returns an error if the content length cannot be determined.
		return err
	}

	// Returns an error if the content length is zero, indicating an invalid or empty file.
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
	}

	// Calculates the size of each chunk based on the total content length and the number of threads.
	chunkSize = contentLength / int64(numThreads)

	var downloadWg sync.WaitGroup // WaitGroup to synchronize the completion of all download threads.

	// Loops over the number of threads to create a concurrent download for each chunk.
	for i := 0; i < numThreads; i++ {
		from := int64(i) * chunkSize // Starting byte for the current chunk.
		to := from + chunkSize - 1   // Ending byte for the current chunk.

		// For the last chunk, ensure that the 'to' value covers the remainder of the file.
		if i == numThreads-1 {
			to = contentLength - 1
		}

		// Adds one to the WaitGroup to track this download thread.
		downloadWg.Add(1)

		// Starts a new goroutine for each chunk download.
		go func(from, to int64, part int, httpClient *http.Client) {
			defer downloadWg.Done() // Marks the thread as done when it finishes.

			// Downloads the part of the file corresponding to the byte range (from, to).
			err := downloadPart(url, from, to, part, httpClient, destPath, m)
			if err != nil {
				// Logs an error if the download of this part fails.
				log.Printf("Thread %d: download part failed: %v\n", part, err)
			}
		}(from, to, i, httpClient) // Passes the byte range, part number, and httpClient to the goroutine.
	}

	// Waits for all download threads to complete before proceeding.
	downloadWg.Wait()

	// Combines all the downloaded parts into a single file.
	err = combineParts(destPath, numThreads)
	if err != nil {
		// Returns an error if combining the parts fails.
		return fmt.Errorf("failed to combine parts: %v", err)
	}

	// Returns nil to indicate that the download and combination were successful.
	return nil
}

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(
	videoURL string,
	episodes []models.Episode,
	selectedEpisodeNum int,
	animeURL string,
	episodeNumberStr string,
	animeMalID int,
	updater *RichPresenceUpdater,
) {
	downloadOption := askForDownload()
	switch downloadOption {
	case 1:
		// Download the current episode
		downloadAndPlayEpisode(
			videoURL,
			episodes,
			selectedEpisodeNum,
			animeURL,
			episodeNumberStr,
			animeMalID,
			updater,
		)
	case 2:
		// Download episodes in a range
		if err := HandleBatchDownload(episodes, animeURL); err != nil {
			log.Panicln("Failed to download episodes:", util.ErrorHandler(err))
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
			log.Panicln("Failed to play video:", util.ErrorHandler(err))
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
	updater *RichPresenceUpdater,
) {
	currentUser, err := user.Current()
	if err != nil {
		log.Panicln("Failed to get current user:", util.ErrorHandler(err))
	}

	downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
	episodePath := filepath.Join(downloadPath, episodeNumberStr+".mp4")

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
			log.Panicln("Failed to create download directory:", util.ErrorHandler(err))
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
				log.Printf("Failed to download video using yt-dlp: %v\n", err)
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
				log.Panicln("Failed to get content length:", util.ErrorHandler(err))
			}
			m.totalBytes = contentLength

			// Start the download in a separate goroutine
			go func() {
				// Update status
				p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))

				if err := DownloadVideo(videoURL, episodePath, numThreads, m); err != nil {
					log.Panicln("Failed to download video:", util.ErrorHandler(err))
				}

				m.mu.Lock()
				m.done = true
				m.mu.Unlock()

				// Final status update
				p.Send(statusMsg("Download completed!"))
			}()

			// Run the Bubble Tea program in the main goroutine
			if _, err := p.Run(); err != nil {
				log.Fatalf("error running progress bar: %v", err)
			}
		}
	} else {
		fmt.Println("Video already downloaded.")
	}

	if askForPlayOffline() {
		if err := playVideo(episodePath, episodes, selectedEpisodeNum, animeMalID, updater); err != nil {
			log.Panicln("Failed to play video:", util.ErrorHandler(err))
		}
	}
}

// askForDownload presents a prompt for the user to choose a download option.
//
// This function displays a menu with options for downloading a single episode,
// downloading a range of episodes, or skipping the download and playing online.
// Based on the user's selection, it returns a corresponding integer code.
//
// Returns:
// - 1 if the user selects "Download this episode".
// - 2 if the user selects "Download episodes in a range".
// - 3 if the user selects "No download (play online)" or an invalid option.
func askForDownload() int {
	// Creates a prompt using the promptui.Select widget with a label and three options.
	prompt := promptui.Select{
		Label: "Choose an option",                                                                             // The label displayed at the top of the menu.
		Items: []string{"Download this episode", "Download episodes in a range", "No download (play online)"}, // The menu items to select from.
	}
	//

	// Runs the prompt and captures the selected result and any potential error.
	_, result, err := prompt.Run()
	if err != nil {
		// If an error occurs while acquiring user input, it logs the error and terminates the program using Panic.
		log.Panicln("Error acquiring user input:", util.ErrorHandler(err))
	}

	// Converts the user's input to lowercase and determines the selected option.
	switch strings.ToLower(result) {
	case "download this episode":
		// Returns 1 if the user selected "Download this episode".
		return 1
	case "download episodes in a range":
		// Returns 2 if the user selected "Download episodes in a range".
		return 2
	default:
		// Returns 3 for any other selection, including "No download (play online)".
		return 3
	}
}

func askForPlayOffline() bool {
	prompt := promptui.Select{
		Label: "Do you want to play the downloaded version offline?",
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Panicln("Error acquiring user input:", util.ErrorHandler(err))
	}
	return strings.ToLower(result) == "yes"
}

// Helper functions for batch download with best quality and concurrency
func getEpisodeRange() (startNum, endNum int, err error) {
	prompt := promptui.Prompt{Label: "Enter start episode number"}
	startStr, err := prompt.Run()
	if err != nil {
		return 0, 0, err
	}

	prompt.Label = "Enter end episode number"
	endStr, err := prompt.Run()
	if err != nil {
		return 0, 0, err
	}

	startNum, _ = strconv.Atoi(startStr)
	endNum, _ = strconv.Atoi(endStr)
	if startNum > endNum {
		return 0, 0, fmt.Errorf("start cannot be greater than end")
	}

	return startNum, endNum, nil
}

// findEpisode returns the episode struct for a given episode number
func findEpisode(episodes []models.Episode, episodeNum int) (models.Episode, bool) {
	for _, ep := range episodes {
		if ep.Num == episodeNum {
			return ep, true
		}
	}
	return models.Episode{}, false
}

func createEpisodePath(animeURL string, epNum int) (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	safeAnimeName := strings.ReplaceAll(DownloadFolderFormatter(animeURL), " ", "_")
	downloadDir := filepath.Join(userHome, ".local", "goanime", "downloads", "anime", safeAnimeName)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(downloadDir, fmt.Sprintf("%d.mp4", epNum)), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func downloadWithYtDlp(url, path string) error {
	cmd := exec.Command("yt-dlp",
		"--no-progress",
		"-f", "best",
		"-o", path,
		url,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("yt-dlp error: %v\n%s", err, string(output))
	}
	return nil
}

// ExtractVideoSources returns a list of available video sources (quality and URL) for an episode URL.
// This is a helper for batch download best quality selection.
func ExtractVideoSources(episodeURL string) ([]struct {
	Quality int
	URL     string
}, error) {
	// Step 1: Extract the raw video source URL (not the final best quality URL)
	videoSrc, err := extractVideoURL(episodeURL)
	if err != nil {
		return nil, err
	}

	// Step 2: If it's an AnimeFire video page, fetch and parse the JSON
	if strings.Contains(videoSrc, "animefire.plus/video/") {
		resp, err := api.SafeGet(videoSrc)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
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
				if util.IsDebug {
					log.Printf("Found quality: label='%s' parsed=%d url=%s", v.Label, q, v.Src)
				}
				sources = append(sources, struct {
					Quality int
					URL     string
				}{Quality: q, URL: v.Src})
			}
			return sources, nil
		}
	}

	// Step 3: Try to parse as JSON (for other sources that may return JSON directly)
	var resp struct {
		Data []struct {
			Src   string `json:"src"`
			Label string `json:"label"`
		}
	}
	if err := json.Unmarshal([]byte(videoSrc), &resp); err == nil && len(resp.Data) > 0 {
		var sources []struct {
			Quality int
			URL     string
		}
		for _, v := range resp.Data {
			labelDigits := regexp.MustCompile(`\d+`).FindString(v.Label)
			q := 0
			if labelDigits != "" {
				q, _ = strconv.Atoi(labelDigits)
			}
			if util.IsDebug {
				log.Printf("Found quality: label='%s' parsed=%d url=%s", v.Label, q, v.Src)
			}
			sources = append(sources, struct {
				Quality int
				URL     string
			}{Quality: q, URL: v.Src})
		}
		return sources, nil
	}

	// Step 4: Fallback: try to extract quality from URL
	re := regexp.MustCompile(`(\d{3,4})p?\\.mp4`)
	matches := re.FindStringSubmatch(videoSrc)
	if len(matches) > 1 {
		q, _ := strconv.Atoi(matches[1])
		if util.IsDebug {
			log.Printf("Fallback: found quality in URL: %d for %s", q, videoSrc)
		}
		return []struct {
			Quality int
			URL     string
		}{{Quality: q, URL: videoSrc}}, nil
	}

	// Step 5: If no quality info, return as is with 0 quality
	if util.IsDebug {
		log.Printf("No quality info found, returning as is: %s", videoSrc)
	}
	return []struct {
		Quality int
		URL     string
	}{{Quality: 0, URL: videoSrc}}, nil
}

func getBestQualityURL(episodeURL string) (string, error) {
	sources, err := ExtractVideoSources(episodeURL)
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

// ExtractVideoSourcesWithPrompt returns a list of available video sources and allows user to select quality for streaming.
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
	// Prompt user to select quality
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
		return sources[0].URL, nil // fallback to best
	}
	// Find the selected quality
	for _, s := range sources {
		if fmt.Sprintf("%dp", s.Quality) == result {
			return s.URL, nil
		}
	}
	return sources[0].URL, nil // fallback
}

// Nova implementação de HandleBatchDownload
func HandleBatchDownload(episodes []models.Episode, animeURL string) error {
	// Get episode range from user
	startNum, endNum, err := getEpisodeRange()
	if err != nil {
		return fmt.Errorf("invalid episode range: %w", err)
	}

	// Initialize progress tracking
	var (
		m          *model
		p          *tea.Program
		totalBytes int64
		httpClient = &http.Client{
			Transport: api.SafeTransport(10 * time.Second),
		}
	)

	// Calculate total size with best quality sources
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			log.Printf("Episode %d not found\n", episodeNum)
			continue
		}
		videoURL, err := getBestQualityURL(episode.URL)
		if err != nil {
			log.Printf("Skipping episode %d: %v\n", episodeNum, err)
			continue
		}
		contentLength, err := getContentLength(videoURL, httpClient)
		if err == nil {
			totalBytes += contentLength
		}
	}

	// Initialize UI components if we have downloadable content
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
		sem := make(chan struct{}, 4) // Concurrent download limiter
		for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
			sem <- struct{}{}
			wg.Add(1)
			go func(epNum int) {
				defer func() {
					<-sem
					wg.Done()
				}()
				episode, found := findEpisode(episodes, epNum)
				if !found {
					log.Printf("Episode %d not found\n", epNum)
					return
				}
				videoURL, err := getBestQualityURL(episode.URL)
				if err != nil {
					log.Printf("Skipping episode %d: %v\n", epNum, err)
					return
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					log.Printf("Episode %d path error: %v\n", epNum, err)
					return
				}
				if fileExists(episodePath) {
					log.Printf("Episode %d already exists\n", epNum)
					return
				}
				if p != nil {
					p.Send(statusMsg(fmt.Sprintf("Downloading episode %d...", epNum)))
				}
				if strings.Contains(videoURL, "blogger.com") {
					err = downloadWithYtDlp(videoURL, episodePath)
				} else {
					err = DownloadVideo(videoURL, episodePath, 4, m)
				}
				if err != nil {
					log.Printf("Failed episode %d: %v\n", epNum, err)
				}
			}(episodeNum)
		}
		wg.Wait()
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
	return nil
}

func mpvSendPosition(sock string) (int, error) {
	val, err := mpvSendCommand(sock, []interface{}{"get_property", "time-pos"})
	if err != nil || val == nil {
		return -1, err
	}
	if f, ok := val.(float64); ok {
		return int(f + 0.5), nil // arredonda
	}
	return -1, nil
}

// salva posição final antes de sair/trocar
func saveFinalPosition(tr *tracking.LocalTracker, sock string,
	animeID int, url string, epNum int,
	title string, durationSec int) {

	if pos, _ := mpvSendPosition(sock); pos >= 0 {
		_ = tr.UpdateProgress(tracking.Anime{
			AnilistID:     animeID,
			AllanimeID:    url,
			EpisodeNumber: epNum,
			PlaybackTime:  pos,
			Duration:      durationSec,
			Title:         title,
			LastUpdated:   time.Now(),
		})
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return "Unknown Title"
}

func promptYesNo(label string) (bool, error) {
	p := promptui.Select{Label: label, Items: []string{"Sim", "Não"}}
	_, res, err := p.Run()
	return strings.ToLower(res) == "sim", err
}

func nextIndex(cmd rune, cur, total int) int {
	if cmd == 'n' && cur+1 < total {
		return cur + 1
	}
	if cmd == 'p' && cur > 0 {
		return cur - 1
	}
	return cur
}

func getURL(ep models.Episode) string {
	u, _ := GetVideoURLForEpisode(ep.URL)
	return u
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
