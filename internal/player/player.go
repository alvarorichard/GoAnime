package player

import (
	"bufio"
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
		go func(from, to int64, part int) {
			defer downloadWg.Done() // Marks the thread as done when it finishes.

			// Downloads the part of the file corresponding to the byte range (from, to).
			err := downloadPart(url, from, to, part, httpClient, destPath, m)
			if err != nil {
				// Logs an error if the download of this part fails.
				log.Printf("Thread %d: download part failed: %v\n", part, err)
			}
		}(from, to, i) // Passes the byte range and part number to the goroutine.
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

//// HandleDownloadAndPlay handles the download and playback of the video
//func HandleDownloadAndPlay(videoURL string, episodes []models.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string, updater *RichPresenceUpdater) {
//	downloadOption := askForDownload()
//	switch downloadOption {
//	case 1:
//		// Download the current episode
//		downloadAndPlayEpisode(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr, updater)
//	case 2:
//		// Download episodes in a range
//		if err := HandleBatchDownload(episodes, animeURL); err != nil {
//			log.Panicln("Failed to download episodes:", util.ErrorHandler(err))
//		}
//	default:
//		// Play online
//		if err := playVideo(videoURL, episodes, selectedEpisodeNum, updater); err != nil {
//			log.Panicln("Failed to play video:", util.ErrorHandler(err))
//		}
//	}
//}

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
		if err := playVideo(
			videoURL,
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

func HandleBatchDownload(episodes []models.Episode, animeURL string) error {
	// Get the start and end episode numbers from the user
	prompt := promptui.Prompt{
		Label: "Enter the start episode number",
	}
	startStr, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("error acquiring start episode number: %v", err)
	}

	prompt = promptui.Prompt{
		Label: "Enter the end episode number",
	}
	endStr, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("error acquiring end episode number: %v", err)
	}

	// Convert to integers
	startNum, err := strconv.Atoi(startStr)
	if err != nil {
		return fmt.Errorf("invalid start episode number: %v", err)
	}
	endNum, err := strconv.Atoi(endStr)
	if err != nil {
		return fmt.Errorf("invalid end episode number: %v", err)
	}

	if startNum > endNum {
		return fmt.Errorf("start episode number cannot be greater than end episode number")
	}

	// Initialize variables for progress bar
	var m *model
	var p *tea.Program
	useProgressBar := false // Flag to determine if progress bar is needed

	// Prepare to calculate total content length
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}

	m = &model{
		progress: progress.New(progress.WithDefaultGradient()),
		keys: keyMap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
	}
	p = tea.NewProgram(m)

	// Calculate total content length
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		// Find the episode in the 'episodes' slice
		var episode models.Episode
		found := false
		for _, ep := range episodes {
			// Extract numeric part from ep.Number
			epNumStr := ExtractEpisodeNumber(ep.Number)
			epNum, err := strconv.Atoi(epNumStr)
			if err != nil {
				continue
			}
			if epNum == episodeNum {
				episode = ep
				found = true
				break
			}
		}
		if !found {
			log.Printf("Episode %d not found\n", episodeNum)
			continue
		}

		// Get video URL
		videoURL, err := GetVideoURLForEpisode(episode.URL)
		if err != nil {
			log.Printf("Failed to get video URL for episode %d: %v\n", episodeNum, err)
			continue
		}

		// Check if the video URL is from Blogger
		if strings.Contains(videoURL, "blogger.com") {
			// Skip adding content length for episodes using yt-dlp
			continue
		}

		// Get content length
		contentLength, err := getContentLength(videoURL, httpClient)
		if err != nil {
			log.Printf("Failed to get content length for episode %d: %v\n", episodeNum, err)
			continue
		}

		m.totalBytes += contentLength
		useProgressBar = true
	}

	// Start the Bubble Tea program in the main goroutine if needed
	if useProgressBar {
		// Start the download in a separate goroutine
		downloadErrChan := make(chan error)

		go func() {
			var overallWg sync.WaitGroup

			// Now start downloads
			for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
				// Find the episode in the 'episodes' slice
				var episode models.Episode
				found := false
				for _, ep := range episodes {
					// Extract numeric part from ep.Number
					epNumStr := ExtractEpisodeNumber(ep.Number)
					epNum, err := strconv.Atoi(epNumStr)
					if err != nil {
						continue
					}
					if epNum == episodeNum {
						episode = ep
						found = true
						break
					}
				}
				if !found {
					log.Printf("Episode %d not found\n", episodeNum)
					continue
				}

				// Get video URL
				videoURL, err := GetVideoURLForEpisode(episode.URL)
				if err != nil {
					log.Printf("Failed to get video URL for episode %d: %v\n", episodeNum, err)
					continue
				}

				// Build download path
				currentUser, err := user.Current()
				if err != nil {
					log.Panicln("Failed to get current user:", util.ErrorHandler(err))
				}

				downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
				episodeNumberStr := strconv.Itoa(episodeNum)
				episodePath := filepath.Join(downloadPath, episodeNumberStr+".mp4")

				if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
					if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
						log.Panicln("Failed to create download directory:", util.ErrorHandler(err))
					}
				}

				if _, err := os.Stat(episodePath); os.IsNotExist(err) {
					numThreads := 4 // Define the number of threads for downloading

					overallWg.Add(1)
					go func(videoURL, episodePath, episodeNumberStr string) {
						defer overallWg.Done()

						// Check if the video URL is from Blogger
						if strings.Contains(videoURL, "blogger.com") {
							// Use yt-dlp to download the video from Blogger
							fmt.Printf("Downloading episode %s with yt-dlp...\n", episodeNumberStr)
							cmd := exec.Command("yt-dlp", "--no-progress", "-o", episodePath, videoURL)
							if err := cmd.Run(); err != nil {
								log.Printf("Failed to download video using yt-dlp: %v\n", err)
							} else {
								fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
							}
						} else {
							// Update status
							p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))

							if err := DownloadVideo(videoURL, episodePath, numThreads, m); err != nil {
								log.Printf("Failed to download episode %s: %v\n", episodeNumberStr, err)
							}
						}
					}(videoURL, episodePath, episodeNumberStr)
				} else {
					log.Printf("Episode %d already downloaded.\n", episodeNum)
				}
			}

			overallWg.Wait()
			if useProgressBar {
				m.mu.Lock()
				m.done = true
				m.mu.Unlock()

				// Final status update
				p.Send(statusMsg("All videos downloaded successfully!"))
			} else {
				fmt.Println("All videos downloaded successfully!")
			}

			downloadErrChan <- nil
		}()

		// Run the Bubble Tea program in the main goroutine
		if _, err := p.Run(); err != nil {
			log.Fatalf("error running progress bar: %v", err)
		}

		// Wait for the download goroutine to finish
		if err := <-downloadErrChan; err != nil {
			return err
		}
	} else {
		// No need for progress bar; just proceed with downloads
		// Similar logic without progress bar
		var overallWg sync.WaitGroup

		for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
			// Find the episode in the 'episodes' slice
			var episode models.Episode
			found := false
			for _, ep := range episodes {
				// Extract numeric part from ep.Number
				epNumStr := ExtractEpisodeNumber(ep.Number)
				epNum, err := strconv.Atoi(epNumStr)
				if err != nil {
					continue
				}
				if epNum == episodeNum {
					episode = ep
					found = true
					break
				}
			}
			if !found {
				log.Printf("Episode %d not found\n", episodeNum)
				continue
			}

			// Get video URL
			videoURL, err := GetVideoURLForEpisode(episode.URL)
			if err != nil {
				log.Printf("Failed to get video URL for episode %d: %v\n", episodeNum, err)
				continue
			}

			// Build download path
			currentUser, err := user.Current()
			if err != nil {
				log.Panicln("Failed to get current user:", util.ErrorHandler(err))
			}

			downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
			episodeNumberStr := strconv.Itoa(episodeNum)
			episodePath := filepath.Join(downloadPath, episodeNumberStr+".mp4")

			if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
				if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
					log.Panicln("Failed to create download directory:", util.ErrorHandler(err))
				}
			}

			if _, err := os.Stat(episodePath); os.IsNotExist(err) {
				numThreads := 4 // Define the number of threads for downloading

				overallWg.Add(1)
				go func(videoURL, episodePath, episodeNumberStr string) {
					defer overallWg.Done()

					// Check if the video URL is from Blogger
					if strings.Contains(videoURL, "blogger.com") {
						// Use yt-dlp to download the video from Blogger
						fmt.Printf("Downloading episode %s with yt-dlp...\n", episodeNumberStr)
						cmd := exec.Command("yt-dlp", "--no-progress", "-o", episodePath, videoURL)
						if err := cmd.Run(); err != nil {
							log.Printf("Failed to download video using yt-dlp: %v\n", err)
						} else {
							fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
						}
					} else {
						// Use standard download method without progress bar
						fmt.Printf("Downloading episode %s...\n", episodeNumberStr)
						if err := DownloadVideo(videoURL, episodePath, numThreads, nil); err != nil {
							log.Printf("Failed to download episode %s: %v\n", episodeNumberStr, err)
						}
						fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
					}
				}(videoURL, episodePath, episodeNumberStr)
			} else {
				log.Printf("Episode %d already downloaded.\n", episodeNum)
			}
		}

		overallWg.Wait()
		fmt.Println("All videos downloaded successfully!")
	}

	return nil
}

// playVideo handles the online playback of a video and user interaction.
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	animeMalID int,
	updater *RichPresenceUpdater,
) error {
	// Initialize mpv arguments
	mpvArgs := make([]string, 0)

	// Fix the URL if it has a double 'p' in the quality suffix
	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)

	// Fetch AniSkip data for the current episode
	if util.IsDebug {
		log.Printf("Video URL: %s", videoURL)
	}

	currentEpisode := &episodes[currentEpisodeNum-1]
	err := api.GetAndParseAniSkipData(animeMalID, currentEpisodeNum, currentEpisode)
	if err != nil {
		log.Printf("AniSkip data not available for episode %d: %v\n", currentEpisodeNum, err)
	} else if util.IsDebug {
		log.Printf("AniSkip data for episode %d: %+v\n", currentEpisodeNum, currentEpisode.SkipTimes)
	}

	// Add skip times to mpv arguments if available
	if currentEpisode.SkipTimes.Op.Start > 0 || currentEpisode.SkipTimes.Op.End > 0 {
		opStart, opEnd := currentEpisode.SkipTimes.Op.Start, currentEpisode.SkipTimes.Op.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_op=%d-%d", opStart, opEnd))
	}
	if currentEpisode.SkipTimes.Ed.Start > 0 || currentEpisode.SkipTimes.Ed.End > 0 {
		edStart, edEnd := currentEpisode.SkipTimes.Ed.Start, currentEpisode.SkipTimes.Ed.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_ed=%d-%d", edStart, edEnd))
	}

	// Initialize local tracker
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	databaseFile := filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "anime_progress.csv")
	tracker := tracking.NewLocalTracker(databaseFile)
	if tracker == nil {
		return fmt.Errorf("failed to initialize tracker")
	}

	// Check for existing progress
	existingAnime, err := tracker.GetAnime(animeMalID, currentEpisode.URL)
	if err != nil {
		log.Printf("Failed to get existing progress: %v", err)
	} else if existingAnime != nil {
		// Add seek position to mpv arguments if we have existing progress
		if existingAnime.PlaybackTime > 0 {
			mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", existingAnime.PlaybackTime))
			if util.IsDebug {
				log.Printf("Resuming from position: %d seconds", existingAnime.PlaybackTime)
			}
		}
	}

	// Start mpv with IPC support
	socketPath, err := StartVideo(videoURL, mpvArgs)
	if err != nil {
		return fmt.Errorf("failed to start video with IPC: %w", err)
	}

	// Only proceed with Rich Presence updates if updater is not nil
	if updater != nil {
		// Wait for the episode to start before retrieving the duration
		go func() {
			for {
				// Get current playback time
				timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
				if err != nil {
					if util.IsDebug {
						log.Printf("Error getting playback time: %v", err)
					}
				}

				// Check if playback has started
				if timePos != nil {
					if !updater.episodeStarted {
						updater.episodeStarted = true
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()

		// Retrieve the video duration once the episode has started
		go func() {
			for {
				if updater.episodeStarted && updater.episodeDuration == 0 {
					// Retrieve video duration
					durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
					if err != nil {
						log.Printf("Error getting video duration: %v", err)
					} else if durationPos != nil {
						if duration, ok := durationPos.(float64); ok {
							// Set episodeDuration correctly in seconds
							updater.episodeDuration = time.Duration(duration * float64(time.Second))
							if util.IsDebug {
								log.Printf("Retrieved Video duration: %v seconds", updater.episodeDuration.Seconds())
							}

							// Validate duration
							if updater.episodeDuration < time.Second {
								log.Printf("Warning: Retrieved episode duration is very small (%v). Setting a default duration.", updater.episodeDuration)
								updater.episodeDuration = 24 * time.Minute // Set a reasonable default duration if necessary
							}

							// Update local tracking with initial duration
							anime := tracking.Anime{
								AnilistID:     animeMalID,
								AllanimeID:    currentEpisode.URL,
								EpisodeNumber: currentEpisodeNum,
								Duration:      int(updater.episodeDuration.Seconds()),
								Title:         currentEpisode.Title.English,
								LastUpdated:   time.Now(),
							}
							if err := tracker.UpdateProgress(anime); err != nil {
								log.Printf("Failed to update local tracking: %v", err)
							}
						} else {
							log.Printf("Error: duration is not a float64")
						}
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()

		// Set up the Rich Presence updater and start it
		updater.socketPath = socketPath
		updater.Start()
		defer updater.Stop()
	}

	// Locate the index of the current episode
	currentEpisodeIndex := -1
	for i, ep := range episodes {
		if ExtractEpisodeNumber(ep.Number) == strconv.Itoa(currentEpisodeNum) {
			currentEpisodeIndex = i
			break
		}
	}
	if currentEpisodeIndex == -1 {
		return fmt.Errorf("current episode number %d not found", currentEpisodeNum)
	}

	// Command loop for user interaction
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press 'n' for next episode, 'p' for previous episode, 'q' to quit, 's' to skip intro:")

	// Start a goroutine to periodically update local tracking
	stopTracking := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get current playback time
				timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
				if err != nil {
					if util.IsDebug {
						log.Printf("Error getting playback time for tracking: %v", err)
					}
					continue
				}

				if timePos != nil {
					if position, ok := timePos.(float64); ok {
						// Update local tracking
						anime := tracking.Anime{
							AnilistID:     animeMalID,
							AllanimeID:    currentEpisode.URL,
							EpisodeNumber: currentEpisodeNum,
							PlaybackTime:  int(position),
							Duration:      int(updater.episodeDuration.Seconds()),
							Title:         currentEpisode.Title.English,
							LastUpdated:   time.Now(),
						}
						if err := tracker.UpdateProgress(anime); err != nil {
							log.Printf("Failed to update local tracking: %v", err)
						}
					}
				}
			case <-stopTracking:
				return
			}
		}
	}()

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Printf("Failed to read command: %v\n", err)
			break
		}

		switch char {
		case 'n': // Next episode
			if currentEpisodeIndex+1 < len(episodes) {
				nextEpisode := episodes[currentEpisodeIndex+1]
				if updater != nil {
					updater.Stop()
				}
				nextVideoURL, err := GetVideoURLForEpisode(nextEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for next episode: %v\n", err)
					continue
				}
				// Set duration for the next episode
				nextEpisodeDuration := time.Duration(nextEpisode.Duration) * time.Second
				var newUpdater *RichPresenceUpdater
				if updater != nil {
					newUpdater = NewRichPresenceUpdater(
						updater.anime,
						updater.isPaused,
						updater.animeMutex,
						updater.updateFreq,
						nextEpisodeDuration,
						"",
					)
					updater.episodeStarted = false
				}
				close(stopTracking)
				return playVideo(nextVideoURL, episodes, currentEpisodeNum+1, animeMalID, newUpdater)
			} else {
				fmt.Println("Already at the last episode.")
			}
		case 'p': // Previous episode
			if currentEpisodeIndex > 0 {
				prevEpisode := episodes[currentEpisodeIndex-1]
				if updater != nil {
					updater.Stop()
				}
				prevVideoURL, err := GetVideoURLForEpisode(prevEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for previous episode: %v\n", err)
					continue
				}
				// Set duration for the previous episode
				prevEpisodeDuration := time.Duration(prevEpisode.Duration) * time.Second
				var newUpdater *RichPresenceUpdater
				if updater != nil {
					newUpdater = NewRichPresenceUpdater(
						updater.anime,
						updater.isPaused,
						updater.animeMutex,
						updater.updateFreq,
						prevEpisodeDuration,
						"",
					)
					updater.episodeStarted = false
				}
				close(stopTracking)
				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1, animeMalID, newUpdater)
			} else {
				fmt.Println("Already at the first episode.")
			}
		case 'q': // Quit
			fmt.Println("Quitting video playback.")
			close(stopTracking)
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return nil
		case 's': // Skip intro (OP)
			if currentEpisode.SkipTimes.Op.End > 0 {
				fmt.Printf("Skipping intro to %d seconds.\n", currentEpisode.SkipTimes.Op.End)
				_, _ = mpvSendCommand(socketPath, []interface{}{"seek", currentEpisode.SkipTimes.Op.End, "absolute"})
			} else {
				fmt.Println("No intro skip data available for this episode.")
			}
		}
	}

	return nil
}

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
