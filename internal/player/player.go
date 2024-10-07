package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
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
	"github.com/charmbracelet/lipgloss" // For styling
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

// Update handles updates to the Bubble Tea model
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch msg := msg.(type) {
	case tickMsg:
		if m.done {
			return m, tea.Quit
		}
		cmd := m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
		return m, tea.Batch(cmd, tickCmd())

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
func (m *model) View() string {
	pad := strings.Repeat(" ", padding)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	return "\n" +
		pad + statusStyle.Render(m.status) + "\n\n" +
		pad + m.progress.View() + "\n\n" +
		pad + "Press Ctrl+C to quit"
}

// tickCmd returns a command to tick every 100 milliseconds
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// statusUpdateCmd returns a command to update the status
func statusUpdateCmd(s string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(s)
	}
}

// DownloadFolderFormatter formats the anime URL to be used as the download folder name
func DownloadFolderFormatter(str string) string {
	regex := regexp.MustCompile(`https?://[^/]+/video/([^/?]+)`)
	match := regex.FindStringSubmatch(str)
	if len(match) > 1 {
		finalStep := match[1]
		return finalStep
	}
	return ""
}

// getContentLength gets the content length of the URL.
func getContentLength(url string, client *http.Client) (int64, error) {
	resp, err := client.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("server does not support partial content: status code %d", resp.StatusCode)
	}

	contentLength, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return 0, err
	}

	return contentLength, nil
}

// downloadPart downloads a part of the video file.
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model, p *tea.Program) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), part)
	partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
	file, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
			// Update received progress
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

// combineParts combines downloaded parts into a single file.
func combineParts(destPath string, numThreads int) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for i := 0; i < numThreads; i++ {
		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)

		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, partFile); err != nil {
			partFile.Close()
			return err
		}
		partFile.Close()

		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}

	return nil
}

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int, m *model, p *tea.Program) error {
	destPath = filepath.Clean(destPath)

	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}

	chunkSize := int64(0)
	var contentLength int64

	// Get content length
	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		return err
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
		go func(from, to int64, part int) {
			defer downloadWg.Done()
			err := downloadPart(url, from, to, part, httpClient, destPath, m, p)
			if err != nil {
				log.Printf("Thread %d: download part failed: %v\n", part, err)
			}
		}(from, to, i)
	}

	downloadWg.Wait()

	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}

	return nil
}

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string) {
	downloadOption := askForDownload()
	switch downloadOption {
	case 1:
		// Download the current episode
		downloadAndPlayEpisode(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr)
	case 2:
		// Download episodes in a range
		if err := HandleBatchDownload(episodes, animeURL); err != nil {
			log.Panicln("Failed to download episodes:", util.ErrorHandler(err))
		}
	default:
		// Play online
		if err := playVideo(videoURL, episodes, selectedEpisodeNum); err != nil {
			log.Panicln("Failed to play video:", util.ErrorHandler(err))
		}
	}
}

func downloadAndPlayEpisode(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string) {
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

		// Run the Bubble Tea program in a separate goroutine
		go func() {
			if _, err := p.Run(); err != nil {
				log.Fatalf("error running progress bar: %v", err)
			}
		}()

		// Update status
		p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))

		// Check if the video URL is from Blogger
		if strings.Contains(videoURL, "blogger.com") {
			// Use yt-dlp to download the video from Blogger
			p.Send(statusMsg(fmt.Sprintf("Downloading episode %s with yt-dlp...", episodeNumberStr)))
			cmd := exec.Command("yt-dlp", "-o", episodePath, videoURL)
			if err := cmd.Run(); err != nil {
				log.Panicln("Failed to download video using yt-dlp:", util.ErrorHandler(err))
			}
		} else {
			// Use the standard download method for other video sources
			if err := DownloadVideo(videoURL, episodePath, numThreads, m, p); err != nil {
				log.Panicln("Failed to download video:", util.ErrorHandler(err))
			}
		}

		m.mu.Lock()
		m.done = true
		m.mu.Unlock()

		// Final status update
		p.Send(statusMsg("Download completed!"))

	} else {
		fmt.Println("Video already downloaded.")
	}

	if askForPlayOffline() {
		if err := playVideo(episodePath, episodes, selectedEpisodeNum); err != nil {
			log.Panicln("Failed to play video:", util.ErrorHandler(err))
		}
	}
}

func askForDownload() int {
	prompt := promptui.Select{
		Label: "Choose an option",
		Items: []string{"Download this episode", "Download episodes in a range", "No download (play online)"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Panicln("Error acquiring user input:", util.ErrorHandler(err))
	}
	switch strings.ToLower(result) {
	case "download this episode":
		return 1
	case "download episodes in a range":
		return 2
	default:
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
		return fmt.Errorf("Error acquiring start episode number: %v", err)
	}

	prompt = promptui.Prompt{
		Label: "Enter the end episode number",
	}
	endStr, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("Error acquiring end episode number: %v", err)
	}

	// Convert to integers
	startNum, err := strconv.Atoi(startStr)
	if err != nil {
		return fmt.Errorf("Invalid start episode number: %v", err)
	}
	endNum, err := strconv.Atoi(endStr)
	if err != nil {
		return fmt.Errorf("Invalid end episode number: %v", err)
	}

	if startNum > endNum {
		return fmt.Errorf("Start episode number cannot be greater than end episode number")
	}

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

	// Get total content length
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}
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

		// Get content length
		contentLength, err := getContentLength(videoURL, httpClient)
		if err != nil {
			log.Printf("Failed to get content length for episode %d: %v\n", episodeNum, err)
			continue
		}

		m.totalBytes += contentLength
	}

	// Run the Bubble Tea program in a separate goroutine
	go func() {
		if _, err := p.Run(); err != nil {
			log.Fatalf("error running progress bar: %v", err)
		}
	}()

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

				// Update status
				p.Send(statusMsg(fmt.Sprintf("Downloading episode %s...", episodeNumberStr)))

				// Check if the video URL is from Blogger
				if strings.Contains(videoURL, "blogger.com") {
					// Use yt-dlp to download the video from Blogger
					p.Send(statusMsg(fmt.Sprintf("Downloading episode %s with yt-dlp...", episodeNumberStr)))
					cmd := exec.Command("yt-dlp", "-o", episodePath, videoURL)
					if err := cmd.Run(); err != nil {
						log.Printf("Failed to download video using yt-dlp: %v\n", err)
					}
				} else {
					// Use the standard download method for other video sources
					if err := DownloadVideo(videoURL, episodePath, numThreads, m, p); err != nil {
						log.Printf("Failed to download episode %s: %v\n", episodeNumberStr, err)
					}
				}
			}(videoURL, episodePath, episodeNumberStr)
		} else {
			log.Printf("Episode %d already downloaded.\n", episodeNum)
		}
	}

	overallWg.Wait()
	m.mu.Lock()
	m.done = true
	m.mu.Unlock()

	// Final status update
	p.Send(statusMsg("All videos downloaded successfully!"))

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
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
}

func extractVideoURL(url string) (string, error) {
	response, err := api.SafeGet(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to fetch URL: %+v", err))
	}
	defer response.Body.Close()

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
	defer resp.Body.Close()

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
	defer response.Body.Close()

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

func playVideo(videoURL string, episodes []api.Episode, currentEpisodeNum int) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		cmd := exec.Command("mpv", "--fs", "--force-window", "--no-terminal", videoURL)
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start video player: %v\n", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			fmt.Printf("Failed to play video: %v\n", err)
		}
	}()

	currentEpisodeIndex := -1
	for i, ep := range episodes {
		epNumStr := ExtractEpisodeNumber(ep.Number)
		epNum, err := strconv.Atoi(epNumStr)
		if err != nil {
			continue
		}
		if epNum == currentEpisodeNum {
			currentEpisodeIndex = i
			break
		}
	}

	if currentEpisodeIndex == -1 {
		if util.IsDebug {
			log.Printf("Current episode number %d not found", currentEpisodeNum)
		}
		return errors.New("current episode not found")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press 'n' for next episode, 'p' for previous episode, 'q' to quit:")

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Printf("Failed to read command: %v\n", err)
			break
		}

		switch char {
		case 'n':
			if currentEpisodeIndex+1 < len(episodes) {
				nextEpisode := episodes[currentEpisodeIndex+1]
				fmt.Printf("Switching to next episode: %s\n", nextEpisode.Number)
				wg.Wait()
				nextVideoURL, err := GetVideoURLForEpisode(nextEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for next episode: %v\n", err)
					continue
				}
				return playVideo(nextVideoURL, episodes, currentEpisodeNum+1)
			} else {
				fmt.Println("Already at the last episode.")
			}
		case 'p':
			if currentEpisodeIndex > 0 {
				prevEpisode := episodes[currentEpisodeIndex-1]
				fmt.Printf("Switching to previous episode: %s\n", prevEpisode.Number)
				wg.Wait()
				prevVideoURL, err := GetVideoURLForEpisode(prevEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for previous episode: %v\n", err)
					continue
				}
				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1)
			} else {
				fmt.Println("Already at the first episode.")
			}
		case 'q':
			fmt.Println("Quitting video playback.")
			return nil
		}
	}

	wg.Wait()
	return nil
}

// end code
