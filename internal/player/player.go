package player

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/hugolgst/rich-go/client"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ktr0731/go-fuzzyfinder"
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

type RichPresenceUpdater struct {
	anime           *api.Anime
	isPaused        *bool
	animeMutex      *sync.Mutex
	updateFreq      time.Duration
	done            chan bool
	wg              sync.WaitGroup
	startTime       time.Time     // Start time of playback
	episodeDuration time.Duration // Total duration of the episode
	episodeStarted  bool          // Whether the episode has started
	socketPath      string        // Path to mpv IPC socket
}

func NewRichPresenceUpdater(anime *api.Anime, isPaused *bool, animeMutex *sync.Mutex, updateFreq time.Duration, episodeDuration time.Duration, socketPath string) *RichPresenceUpdater {
	return &RichPresenceUpdater{
		anime:           anime,
		isPaused:        isPaused,
		animeMutex:      animeMutex,
		updateFreq:      updateFreq, // Make sure updateFreq is actually used in the struct
		done:            make(chan bool),
		startTime:       time.Now(),
		episodeDuration: episodeDuration,
		episodeStarted:  false,
		socketPath:      socketPath,
	}
}

func (rpu *RichPresenceUpdater) getCurrentPlaybackPosition() (time.Duration, error) {
	position, err := mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "time-pos"})
	if err != nil {
		return 0, err
	}

	// Convert position to float64 and then to time.Duration
	posSeconds, ok := position.(float64)
	if !ok {
		return 0, fmt.Errorf("failed to parse playback position")
	}

	return time.Duration(posSeconds) * time.Second, nil
}

// Start begins the periodic Rich Presence updates.
func (rpu *RichPresenceUpdater) Start() {
	rpu.wg.Add(1)
	go func() {
		defer rpu.wg.Done()
		ticker := time.NewTicker(rpu.updateFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				go rpu.updateDiscordPresence() // Run update asynchronously
			case <-rpu.done:
				if util.IsDebug {
					log.Println("Rich Presence updater received stop signal.")
				}
				return
			}
		}
	}()
	if util.IsDebug {
		log.Println("Rich Presence updater started.")
	}
}

// Stop signals the updater to stop and waits for the goroutine to finish.
func (rpu *RichPresenceUpdater) Stop() {
	close(rpu.done)
	rpu.wg.Wait()
	if util.IsDebug {
		log.Println("Rich Presence updater stopped.")

	}
}

func (rpu *RichPresenceUpdater) updateDiscordPresence() {
	rpu.animeMutex.Lock()
	defer rpu.animeMutex.Unlock()

	currentPosition, err := rpu.getCurrentPlaybackPosition()
	if err != nil {
		if util.IsDebug {
			log.Printf("Error fetching playback position: %v\n", err)
		}
		return
	}

	// Debug log to check episode duration
	if util.IsDebug {
		log.Printf("Episode Duration in updateDiscordPresence: %v seconds (%v minutes)\n", rpu.episodeDuration.Seconds(), rpu.episodeDuration.Minutes())

	}

	// Convert episode duration to minutes and seconds format
	totalMinutes := int(rpu.episodeDuration.Minutes())
	totalSeconds := int(rpu.episodeDuration.Seconds()) % 60 // Remaining seconds after full minutes

	// Format the current playback position as minutes and seconds
	timeInfo := fmt.Sprintf("%02d:%02d / %02d:%02d",
		int(currentPosition.Minutes()), int(currentPosition.Seconds())%60,
		totalMinutes, totalSeconds,
	)

	// Create the activity with updated Details
	activity := client.Activity{
		Details:    fmt.Sprintf("%s | Episode %s | %s / %d min", rpu.anime.Details.Title.Romaji, rpu.anime.Episodes[0].Number, timeInfo, totalMinutes),
		State:      "Watching",
		LargeImage: rpu.anime.ImageURL,
		LargeText:  rpu.anime.Details.Title.Romaji,
		Buttons: []*client.Button{
			{Label: "View on AniList", Url: fmt.Sprintf("https://anilist.co/anime/%d", rpu.anime.AnilistID)},
			{Label: "View on MAL", Url: fmt.Sprintf("https://myanimelist.net/anime/%d", rpu.anime.MalID)},
		},
	}

	// Set the activity in Discord Rich Presence
	if err := client.SetActivity(activity); err != nil {
		if util.IsDebug {
			log.Printf("Error updating Discord Rich Presence: %v\n", err)
		} else {
			log.Printf("Discord Rich Presence updated with elapsed time: %s\n", timeInfo)
		}
	}
}

// StartVideo opens mpv with a socket for IPC
func StartVideo(link string, args []string) (string, error) {
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	randomNumber := fmt.Sprintf("%x", randomBytes)

	var socketPath string
	if runtime.GOOS == "windows" {
		socketPath = fmt.Sprintf(`\\.\pipe\goanime_mpvsocket_%s`, randomNumber)
	} else {
		socketPath = fmt.Sprintf("/tmp/goanime_mpvsocket_%s", randomNumber)
	}

	mpvArgs := append([]string{"--no-terminal", "--quiet", fmt.Sprintf("--input-ipc-server=%s", socketPath), link}, args...)
	cmd := exec.Command("mpv", mpvArgs...)
	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start mpv: %w", err)
	}

	return socketPath, nil
}

// mpvSendCommand sends a JSON command to MPV via the IPC socket and receives the response.
// mpvSendCommand sends a JSON command to mpv via a socket and reads the response.
func mpvSendCommand(socketPath string, command []interface{}) (interface{}, error) {
	conn, err := dialMPVSocket(socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

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

// Update handles updates to the Bubble Tea model.
//
// This function processes incoming messages (`tea.Msg`) and updates the model's state accordingly.
// It locks the model's mutex to ensure thread safety, especially when modifying shared data like
// `m.received`, `m.totalBytes`, and other stateful properties.
//
// The function processes different message types, including:
//
// 1. `tickMsg`: A periodic message that triggers the progress update. If the download is complete
// (`m.done` is `true`), the program quits. Otherwise, it calculates the percentage of bytes received
// and updates the progress bar. It then schedules the next tick.
//
// 2. `statusMsg`: Updates the status string in the model, which can be used to display custom messages
// to the user, such as "Downloading..." or "Download complete".
//
// 3. `progress.FrameMsg`: Handles frame updates for the progress bar. It delegates the update to the
// internal `progress.Model` and returns any commands necessary to refresh the UI.
//
// 4. `tea.KeyMsg`: Responds to key events, such as quitting the program when "Ctrl+C" is pressed.
// If the user requests to quit, the program sets `m.done` to `true` and returns the quit command.
//
// For unhandled message types, it returns the model unchanged.
//
// Returns:
// - Updated `tea.Model` representing the current state of the model.
// - A `tea.Cmd` that specifies the next action the program should perform.

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch msg := msg.(type) {
	case tickMsg:
		if m.done {
			return m, tea.Quit
		}
		if m.totalBytes > 0 {
			cmd := m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
			return m, tea.Batch(cmd, tickCmd())
		}
		return m, tickCmd()

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case progress.FrameMsg:
		var cmd tea.Cmd
		var newModel tea.Model
		newModel, cmd = m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.quit) {
			m.done = true
			return m, tea.Quit
		}
		return m, nil

	default:
		return m, nil
	}
}

// View renders the Bubble Tea model
// View renders the user interface for the Bubble Tea model.
//
// This function generates the visual output that is displayed to the user. It includes the status message,
// the progress bar, and a quit instruction. The layout is formatted with padding for proper alignment.
//
// Steps:
// 1. Adds padding to each line using spaces.
// 2. Styles the status message (m.status) with an orange color (#FFA500).
// 3. Displays the progress bar using the progress model.
// 4. Shows a message instructing the user to press "Ctrl+C" to quit.
//
// Returns:
// - A formatted string that represents the UI for the current state of the model.
func (m *model) View() string {
	// Creates padding spaces for consistent layout
	pad := strings.Repeat(" ", padding)

	// Styles the status message with an orange color
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

	// Returns the UI layout: status message, progress bar, and quit instruction
	return "\n" +
		pad + statusStyle.Render(m.status) + "\n\n" + // Render the styled status message
		pad + m.progress.View() + "\n\n" + // Render the progress bar
		pad + "Press Ctrl+C to quit" // Show quit instruction
}

// tickCmd returns a command that triggers a "tick" every 100 milliseconds.
//
// This function sets up a recurring event (tick) that fires every 100 milliseconds.
// Each tick sends a `tickMsg` with the current time (`t`) as a message, which can be
// handled by the update function to trigger actions like updating the progress bar.
//
// Returns:
// - A `tea.Cmd` that schedules a tick every 100 milliseconds and sends a `tickMsg`.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// statusUpdateCmd returns a command to update the status

// DownloadFolderFormatter formats the anime URL to create a download folder name.
//
// This function extracts a specific part of the anime video URL to use it as the name
// for the download folder. It uses a regular expression to capture the part of the URL
// after "/video/", which is often unique and suitable as a folder name.
//
// Steps:
// 1. Compiles a regular expression that matches URLs of the form "https://<domain>/video/<unique-part>".
// 2. Extracts the "<unique-part>" from the URL.
// 3. If the match is successful, it returns the extracted part as the folder name.
// 4. If no match is found, it returns an empty string.
//
// Parameters:
// - str: The anime video URL as a string.
//
// Returns:
// - A string representing the formatted folder name, or an empty string if no match is found.
func DownloadFolderFormatter(str string) string {
	// Regular expression to capture the unique part after "/video/"
	regex := regexp.MustCompile(`https?://[^/]+/video/([^/?]+)`)

	// Apply the regex to the input URL
	match := regex.FindStringSubmatch(str)

	// If a match is found, return the captured group (folder name)
	if len(match) > 1 {
		finalStep := match[1]
		return finalStep
	}

	// If no match, return an empty string
	return ""
}

// getContentLength retrieves the content length of the given URL.
func getContentLength(url string, client *http.Client) (int64, error) {
	// Attempts to create an HTTP HEAD request to retrieve headers without downloading the body.
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Returns 0 and the error if the request creation fails.
		return 0, err
	}

	// Sends the HEAD request to the server.
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		// If the HEAD request fails or is not supported, fall back to a GET request.
		req.Method = "GET"
		req.Header.Set("Range", "bytes=0-0") // Requests only the first byte to minimize data transfer.
		resp, err = client.Do(req)           // Sends the modified GET request.
		if err != nil {
			// Returns 0 and the error if the GET request fails.
			return 0, err
		}
	}

	// Ensures that the response body is closed after it is used to avoid resource leaks.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Logs a warning if closing the response body fails.
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// Checks if the server responded with a 200 OK or 206 Partial Content status.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		// Returns an error if the server does not support partial content (required for ranged requests).
		return 0, fmt.Errorf("server does not support partial content: status code %d", resp.StatusCode)
	}

	// Retrieves the "Content-Length" header from the response.
	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		// Returns an error if the "Content-Length" header is missing.
		return 0, fmt.Errorf("Content-Length header is missing")
	}

	// Converts the "Content-Length" header from a string to an int64.
	contentLength, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	if err != nil {
		// Returns 0 and an error if the conversion fails.
		return 0, err
	}

	// Returns the content length in bytes.
	return contentLength, nil
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

	// Creates an HTTP client with a custom transport that includes a 10-second timeout.
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
//func HandleDownloadAndPlay(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string, updater *RichPresenceUpdater) {
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

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(
	videoURL string,
	episodes []api.Episode,
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

//func downloadAndPlayEpisode(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string, updater *RichPresenceUpdater) {
//	currentUser, err := user.Current()
//	if err != nil {
//		log.Panicln("Failed to get current user:", util.ErrorHandler(err))
//	}
//
//	downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
//	episodePath := filepath.Join(downloadPath, episodeNumberStr+".mp4")
//
//	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
//		if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
//			log.Panicln("Failed to create download directory:", util.ErrorHandler(err))
//		}
//	}
//
//	if _, err := os.Stat(episodePath); os.IsNotExist(err) {
//		numThreads := 4 // Define the number of threads for downloading
//
//		// Check if the video URL is from Blogger
//		if strings.Contains(videoURL, "blogger.com") {
//			// Use yt-dlp to download the video from Blogger
//			fmt.Printf("Downloading episode %s with yt-dlp...\n", episodeNumberStr)
//			cmd := exec.Command("yt-dlp", "--no-progress", "-o", episodePath, videoURL)
//			if err := cmd.Run(); err != nil {
//				log.Panicln("Failed to download video using yt-dlp:", util.ErrorHandler(err))
//			}
//			fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
//		} else {
//			// Initialize progress model
//			m := &model{
//				progress: progress.New(progress.WithDefaultGradient()),
//				keys: keyMap{
//					quit: key.NewBinding(
//						key.WithKeys("ctrl+c"),
//						key.WithHelp("ctrl+c", "quit"),
//					),
//				},
//			}
//			p := tea.NewProgram(m)
//
//			// Get content length
//			httpClient := &http.Client{
//				Transport: api.SafeTransport(10 * time.Second),
//			}
//			contentLength, err := getContentLength(videoURL, httpClient)
//			if err != nil {
//				log.Panicln("Failed to get content length:", util.ErrorHandler(err))
//			}
//			m.totalBytes = contentLength
//
//			// Start the download in a separate goroutine
//			go func() {
//				// Update status
//				p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))
//
//				if err := DownloadVideo(videoURL, episodePath, numThreads, m); err != nil {
//					log.Panicln("Failed to download video:", util.ErrorHandler(err))
//				}
//
//				m.mu.Lock()
//				m.done = true
//				m.mu.Unlock()
//
//				// Final status update
//				p.Send(statusMsg("Download completed!"))
//			}()
//
//			// Run the Bubble Tea program in the main goroutine
//			if _, err := p.Run(); err != nil {
//				log.Fatalf("error running progress bar: %v", err)
//			}
//		}
//	} else {
//		fmt.Println("Video already downloaded.")
//	}
//
//	if askForPlayOffline() {
//		if err := playVideo(episodePath, episodes, selectedEpisodeNum, updater); err != nil {
//			log.Panicln("Failed to play video:", util.ErrorHandler(err))
//		}
//	}
//}

func downloadAndPlayEpisode(
	videoURL string,
	episodes []api.Episode,
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
			cmd := exec.Command("yt-dlp", "--no-progress", "-o", episodePath, videoURL)
			if err := cmd.Run(); err != nil {
				log.Panicln("Failed to download video using yt-dlp:", util.ErrorHandler(err))
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

func HandleBatchDownload(episodes []api.Episode, animeURL string) error {
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
		var episode api.Episode
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
				var episode api.Episode
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
			var episode api.Episode
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

// SelectEpisodeWithFuzzyFinder allows the user to select an episode using fuzzy finder
func SelectEpisodeWithFuzzyFinder(episodes []api.Episode) (string, string, error) {
	if len(episodes) == 0 {
		return "", "", errors.New("no episodes provided")
	}

	idx, err := fuzzyfinder.Find(
		episodes,
		func(i int) string {
			return episodes[i].Number
		},
		fuzzyfinder.WithPromptString("Select the episode"),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to select episode with go-fuzzyfinder: %w", err)
	}

	if idx < 0 || idx >= len(episodes) {
		return "", "", errors.New("invalid index returned by fuzzyfinder")
	}

	return episodes[idx].URL, episodes[idx].Number, nil
}

// ExtractEpisodeNumber extracts the numeric part of an episode string
func ExtractEpisodeNumber(episodeStr string) string {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeStr)
	if numStr == "" {
		return "1"
	}
	return numStr
}

// GetVideoURLForEpisode gets the video URL for a given episode URL
func GetVideoURLForEpisode(episodeURL string) (string, error) {

	if util.IsDebug {
		log.Printf("Tentando extrair URL de vídeo para o episódio: %s", episodeURL)
	}
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
}

func extractVideoURL(url string) (string, error) {

	if util.IsDebug {
		log.Printf("Extraindo URL de vídeo da página: %s", url)
	}

	response, err := api.SafeGet(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to fetch URL: %+v", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(response.Body)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to parse HTML: %+v", err))
	}

	videoElements := doc.Find("video")
	if videoElements.Length() == 0 {
		videoElements = doc.Find("div")
	}

	if videoElements.Length() == 0 {
		return "", errors.New("no video elements found in the HTML")
	}

	videoSrc, exists := videoElements.Attr("data-video-src")
	if !exists || videoSrc == "" {
		urlBody, err := fetchContent(url)
		if err != nil {
			return "", err
		}
		videoSrc, err = findBloggerLink(urlBody)
		if err != nil {
			return "", err
		}
	}

	return videoSrc, nil
}

func fetchContent(url string) (string, error) {
	resp, err := api.SafeGet(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func findBloggerLink(content string) (string, error) {
	pattern := `https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 0 {
		return matches[0], nil
	} else {
		return "", errors.New("no blogger video link found in the content")
	}
}

func extractActualVideoURL(videoSrc string) (string, error) {
	if strings.Contains(videoSrc, "blogger.com") {
		return videoSrc, nil
	}
	response, err := api.SafeGet(videoSrc)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to fetch video source: %+v", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("request failed with status: %s", response.Status))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to read response body: %+v", err))
	}

	var videoResponse VideoResponse
	if err := json.Unmarshal(body, &videoResponse); err != nil {
		return "", errors.New(fmt.Sprintf("failed to unmarshal JSON response: %+v", err))
	}

	if len(videoResponse.Data) == 0 {
		return "", errors.New("no video data found in the response")
	}

	highestQualityVideoURL := selectHighestQualityVideo(videoResponse.Data)
	if highestQualityVideoURL == "" {
		return "", errors.New("no suitable video quality found")
	}

	return highestQualityVideoURL, nil
}

// VideoData represents the video data structure, with a source URL and a label
type VideoData struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

// VideoResponse represents the video response structure with a slice of VideoData
type VideoResponse struct {
	Data []VideoData `json:"data"`
}

// selectHighestQualityVideo selects the highest quality video available
func selectHighestQualityVideo(videos []VideoData) string {
	var highestQuality int
	var highestQualityURL string
	for _, video := range videos {
		qualityValue, _ := strconv.Atoi(strings.TrimRight(video.Label, "p"))
		if qualityValue > highestQuality {
			highestQuality = qualityValue
			highestQualityURL = video.Src
		}
	}
	return highestQualityURL
}

//
//func playVideo(videoURL string, episodes []api.Episode, currentEpisodeNum int, updater *RichPresenceUpdater) error {
//	// Fetch AniSkip data for the current episode
//	if util.IsDebug {
//		log.Printf("Video URL: %s", videoURL)
//	}
//
//	currentEpisode := &episodes[currentEpisodeNum-1]
//	err := api.GetAndParseAniSkipData(updater.anime.MalID, currentEpisodeNum, currentEpisode)
//	if err != nil {
//		log.Printf("AniSkip data not available for episode %d: %v\n", currentEpisodeNum, err)
//	} else if util.IsDebug {
//		log.Printf("AniSkip data for episode %d: %+v\n", currentEpisodeNum, currentEpisode.SkipTimes)
//	}
//
//	// Prepare mpv arguments to automatically skip OP and ED if available
//	var mpvArgs []string
//	if currentEpisode.SkipTimes.Op.Start > 0 || currentEpisode.SkipTimes.Op.End > 0 {
//		opStart, opEnd := currentEpisode.SkipTimes.Op.Start, currentEpisode.SkipTimes.Op.End
//		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_op=%d-%d", opStart, opEnd))
//	}
//	if currentEpisode.SkipTimes.Ed.Start > 0 || currentEpisode.SkipTimes.Ed.End > 0 {
//		edStart, edEnd := currentEpisode.SkipTimes.Ed.Start, currentEpisode.SkipTimes.Ed.End
//		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_ed=%d-%d", edStart, edEnd))
//	}
//
//	// Start mpv with IPC support
//	socketPath, err := StartVideo(videoURL, mpvArgs)
//	if err != nil {
//		return fmt.Errorf("failed to start video with IPC: %w", err)
//	}
//
//	// Wait for the episode to start before retrieving the duration
//	go func() {
//		for {
//			// Get current playback time
//			timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
//			if err != nil {
//				if util.IsDebug {
//					log.Printf("Error getting playback time: %v", err)
//
//				}
//			}
//
//			// Check if playback has started
//			if timePos != nil {
//				if !updater.episodeStarted {
//					updater.episodeStarted = true
//					break
//				}
//			}
//			time.Sleep(1 * time.Second)
//		}
//	}()
//
//	// Retrieve the video duration once the episode has started
//	go func() {
//		for {
//			if updater.episodeStarted && updater.episodeDuration == 0 {
//				// Retrieve video duration
//				durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
//				if err != nil {
//					log.Printf("Error getting video duration: %v", err)
//				} else if durationPos != nil {
//					if duration, ok := durationPos.(float64); ok {
//						// Set episodeDuration correctly in seconds
//						updater.episodeDuration = time.Duration(duration * float64(time.Second))
//						if util.IsDebug {
//							log.Printf("Retrieved Video duration: %v seconds", updater.episodeDuration.Seconds())
//
//						}
//
//						// Validate duration
//						if updater.episodeDuration < time.Second {
//							log.Printf("Warning: Retrieved episode duration is very small (%v). Setting a default duration.", updater.episodeDuration)
//							updater.episodeDuration = 24 * time.Minute // Set a reasonable default duration if necessary
//						}
//					} else {
//						log.Printf("Error: duration is not a float64")
//					}
//				}
//				break
//			}
//			time.Sleep(1 * time.Second)
//		}
//	}()
//
//	// Set up the Rich Presence updater and start it
//	updater.socketPath = socketPath
//	updater.Start()
//	defer updater.Stop()
//
//	// Locate the index of the current episode
//	currentEpisodeIndex := -1
//	for i, ep := range episodes {
//		if ExtractEpisodeNumber(ep.Number) == strconv.Itoa(currentEpisodeNum) {
//			currentEpisodeIndex = i
//			break
//		}
//	}
//	if currentEpisodeIndex == -1 {
//		return fmt.Errorf("current episode number %d not found", currentEpisodeNum)
//	}
//
//	// Command loop for user interaction
//	reader := bufio.NewReader(os.Stdin)
//	fmt.Println("Press 'n' for next episode, 'p' for previous episode, 'q' to quit, 's' to skip intro:")
//
//	for {
//		char, _, err := reader.ReadRune()
//		if err != nil {
//			fmt.Printf("Failed to read command: %v\n", err)
//			break
//		}
//
//		switch char {
//		case 'n': // Next episode
//			if currentEpisodeIndex+1 < len(episodes) {
//				nextEpisode := episodes[currentEpisodeIndex+1]
//				updater.Stop()
//				nextVideoURL, err := GetVideoURLForEpisode(nextEpisode.URL)
//				if err != nil {
//					fmt.Printf("Failed to get video URL for next episode: %v\n", err)
//					continue
//				}
//				// Set duration for the next episode
//				nextEpisodeDuration := time.Duration(nextEpisode.Duration) * time.Second
//				newUpdater := NewRichPresenceUpdater(updater.anime, updater.isPaused, updater.animeMutex, updater.updateFreq, nextEpisodeDuration, "")
//				updater.episodeStarted = false
//				return playVideo(nextVideoURL, episodes, currentEpisodeNum+1, newUpdater)
//			} else {
//				fmt.Println("Already at the last episode.")
//			}
//		case 'p': // Previous episode
//			if currentEpisodeIndex > 0 {
//				prevEpisode := episodes[currentEpisodeIndex-1]
//				updater.Stop()
//				prevVideoURL, err := GetVideoURLForEpisode(prevEpisode.URL)
//				if err != nil {
//					fmt.Printf("Failed to get video URL for previous episode: %v\n", err)
//					continue
//				}
//				// Set duration for the previous episode
//				prevEpisodeDuration := time.Duration(prevEpisode.Duration) * time.Second
//				newUpdater := NewRichPresenceUpdater(updater.anime, updater.isPaused, updater.animeMutex, updater.updateFreq, prevEpisodeDuration, "")
//				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1, newUpdater)
//			} else {
//				fmt.Println("Already at the first episode.")
//			}
//		case 'q': // Quit
//			fmt.Println("Quitting video playback.")
//			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
//			return nil
//		case 's': // Skip intro (OP)
//			if currentEpisode.SkipTimes.Op.End > 0 {
//				fmt.Printf("Skipping intro to %d seconds.\n", currentEpisode.SkipTimes.Op.End)
//				_, _ = mpvSendCommand(socketPath, []interface{}{"seek", currentEpisode.SkipTimes.Op.End, "absolute"})
//			} else {
//				fmt.Println("No intro skip data available for this episode.")
//			}
//		}
//	}
//
//	return nil
//}

// playVideo handles the online playback of a video and user interaction.
func playVideo(
	videoURL string,
	episodes []api.Episode,
	currentEpisodeNum int,
	animeMalID int, // Added animeMalID parameter
	updater *RichPresenceUpdater,
) error {
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

	// Prepare mpv arguments to automatically skip OP and ED if available
	var mpvArgs []string
	if currentEpisode.SkipTimes.Op.Start > 0 || currentEpisode.SkipTimes.Op.End > 0 {
		opStart, opEnd := currentEpisode.SkipTimes.Op.Start, currentEpisode.SkipTimes.Op.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_op=%d-%d", opStart, opEnd))
	}
	if currentEpisode.SkipTimes.Ed.Start > 0 || currentEpisode.SkipTimes.Ed.End > 0 {
		edStart, edEnd := currentEpisode.SkipTimes.Ed.Start, currentEpisode.SkipTimes.Ed.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_ed=%d-%d", edStart, edEnd))
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
				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1, animeMalID, newUpdater)
			} else {
				fmt.Println("Already at the first episode.")
			}
		case 'q': // Quit
			fmt.Println("Quitting video playback.")
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
