package player

import (
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

	"bufio"
	"github.com/PuerkitoBio/goquery"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

const (
	padding = 2
)

// tickMsg is a message for the tick command
type tickMsg time.Time

// model represents the Bubble Tea model for the progress bar
type model struct {
	progress   progress.Model
	totalBytes int64
	received   int64
	done       bool
	mu         sync.Mutex
	keys       keyMap
}

type keyMap struct {
	quit key.Binding
}

// Init initializes the Bubble Tea model
func (m *model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles updates to the Bubble Tea model
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tickMsg:
		if m.done {
			return m, tea.Quit
		}
		cmd := m.progress.SetPercent(float64(m.received) / float64(m.totalBytes))
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
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
	return "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + "Press Ctrl+C to quit"
}

// tickCmd returns a command to tick every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// DownloadFolderFormatter formats the anime URL to be used as the download folder name
func DownloadFolderFormatter(str string) string {
	regex := regexp.MustCompile(`https://animefire\.plus/video/([^/?]+)`)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

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
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, p *tea.Program, m *model) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), part)
	partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
	file, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close file: %v\n", err)
		}
	}(file)

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
			// Atualizar o progresso recebido
			m.mu.Lock()
			m.received += int64(n)
			m.mu.Unlock()
			p.Send(tickMsg(time.Now())) // Notify progress
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
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			log.Printf("Failed to close output file: %v\n", err)
		}
	}(outFile)

	for i := 0; i < numThreads; i++ {
		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)

		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, partFile); err != nil {
			err := partFile.Close()
			if err != nil {
				return err
			}
			return err
		}
		err = partFile.Close()
		if err != nil {
			return err
		}

		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}

	return nil
}

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int) error {
	destPath = filepath.Clean(destPath)

	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}

	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		return err
	}

	chunkSize := contentLength / int64(numThreads)
	var wg sync.WaitGroup

	m := model{
		progress:   progress.New(progress.WithDefaultGradient()),
		totalBytes: contentLength,
		received:   0,
		done:       false,
		keys: keyMap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
	}

	p := tea.NewProgram(&m)

	// Run the Bubble Tea program in a separate goroutine
	go func() {
		if _, err := p.Run(); err != nil {
			log.Fatalf("error running progress bar: %v", err)
		}
	}()

	for i := 0; i < numThreads; i++ {
		from := int64(i) * chunkSize
		to := from + chunkSize - 1
		if i == numThreads-1 {
			to = contentLength - 1
		}

		wg.Add(1)
		go func(from, to int64, part int) {
			defer wg.Done()
			err := downloadPart(url, from, to, part, httpClient, destPath, p, &m)
			if err != nil {
				log.Printf("Thread %d: download part failed: %v\n", part, err)
			}
		}(from, to, i)
	}

	wg.Wait()
	m.mu.Lock()
	m.done = true
	m.mu.Unlock()
	p.Send(tickMsg(time.Now())) // Final progress update

	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}

	return nil
}

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string) {
	if askForDownload() {
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
			fmt.Println("Downloading the video...")

			// Check if the video URL is from Blogger
			if strings.Contains(videoURL, "blogger.com") {
				// Use yt-dlp para baixar o vídeo do Blogger
				fmt.Println("Using yt-dlp to download Blogger video...")
				cmd := exec.Command("yt-dlp", "-o", episodePath, videoURL)
				if err := cmd.Run(); err != nil {
					log.Panicln("Failed to download video using yt-dlp:", util.ErrorHandler(err))
				}
			} else {
				// Use the standard download method for other video sources
				fmt.Println("Using standard download method...")
				numThreads := 4 // Define the number of threads for downloading
				if err := DownloadVideo(videoURL, episodePath, numThreads); err != nil {
					log.Panicln("Failed to download video:", util.ErrorHandler(err))
				}
			}
			fmt.Println("Video downloaded successfully!")
		} else {
			fmt.Println("Video already downloaded.")
		}

		if askForPlayOffline() {
			if err := playVideo(episodePath, episodes, selectedEpisodeNum); err != nil {
				log.Panicln("Failed to play video:", util.ErrorHandler(err))
			}
		}
	} else {
		if err := playVideo(videoURL, episodes, selectedEpisodeNum); err != nil {
			log.Panicln("Failed to play video:", util.ErrorHandler(err))
		}
	}
}

func askForDownload() bool {
	prompt := promptui.Select{
		Label: "Do you want to download the episode? If not, it will be played online.",
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Panicln("Error acquiring user input:", util.ErrorHandler(err))
	}
	return strings.ToLower(result) == "yes"
}

func askForPlayOffline() bool {
	prompt := promptui.Select{
		Label: "Do you want to play the downloaded version offline",
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Panicln("Error acquiring user input:", util.ErrorHandler(err))
	}
	return strings.ToLower(result) == "yes"
}

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

func ExtractEpisodeNumber(episodeStr string) string {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeStr)
	if numStr == "" {
		return "1"
	}
	return numStr
}

// Given an Anime Fire URL, this funcion will extract the correct video
// provider and return the URL (*string*) of the `.mp4` **OR** the Blogger
// link with the Youtube video.
func GetVideoURLForEpisode(episodeURL string) (string, error) {
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
}

func extractVideoURL(url string) (string, error) {
	response, err := api.SafeGet(url)
	if util.IsDebug {
		log.Printf("Fetching URL: %s - Error Status: %s\n", url, err)
	}
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
	if util.IsDebug {
		log.Printf("Parsing HTML from URL: %s - Error Status: %s\n", url, err)
	}
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to parse HTML: %+v", err))
	}

	videoElements := doc.Find("video")
	if videoElements.Length() == 0 {
		videoElements = doc.Find("div")
	}
	if util.IsDebug {
		log.Printf("Found %d video elements in the HTML\n", videoElements.Length())
	}

	if videoElements.Length() == 0 {
		return "", errors.New("no video elements found in the HTML")
	}

	videoSrc, _ := videoElements.Attr("data-video-src")
	if videoSrc == "" {
		urlBody, err := fetchContent(url)
		if err != nil {
			return "", err
		}
		videoSrc, err = findBloggerLink(urlBody)
		if err != nil {
			return "", err
		}
	}
	if util.IsDebug {
		log.Printf("Found video source URL: %s\n", videoSrc)
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

// selectHighestQualityVideo Assumes that the quality label contains resolution information (e.g., "1080p").  This function can be adapted based on the actual format of the quality labels.
func selectHighestQualityVideo(videos []VideoData) string {
	var highestQuality string
	var highestQualityURL string
	for _, video := range videos {
		if isHigherQuality(video.Label, highestQuality) {
			highestQuality = video.Label
			highestQualityURL = video.Src
		}
	}
	return highestQualityURL
}

// isHigherQuality Compares two quality labels and returns true if the first is of higher quality than the second.
func isHigherQuality(quality1, quality2 string) bool {
	quality1Value, _ := strconv.Atoi(strings.TrimRight(quality1, "p"))
	quality2Value, _ := strconv.Atoi(strings.TrimRight(quality2, "p"))
	return quality1Value > quality2Value
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
		if ep.Num == currentEpisodeNum {
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
				return playVideo(nextVideoURL, episodes, nextEpisode.Num)
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
				return playVideo(prevVideoURL, episodes, prevEpisode.Num)
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
