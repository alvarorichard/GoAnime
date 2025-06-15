package player

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/manifoldco/promptui"
)

// downloadPart downloads a part of the video file using HTTP Range Requests.
func downloadPart(url string, from, to int64, part int, client *http.Client, destPath string, m *model) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger.Warn("Error closing response body", "error", err)
		}
	}()
	partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), part)
	partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
	file, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			util.Logger.Warn("Error closing file", "error", err)
		}
	}()
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
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

// combineParts combines downloaded parts into a single final file.
func combineParts(destPath string, numThreads int) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			util.Logger.Warn("Error closing output file", "error", err)
		}
	}()
	for i := 0; i < numThreads; i++ {
		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, partFile); err != nil {
			if closeErr := partFile.Close(); closeErr != nil {
				util.Logger.Warn("Error closing part file", "error", closeErr)
			}
			return err
		}
		if err := partFile.Close(); err != nil {
			util.Logger.Warn("Error closing part file", "error", err)
		}
		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}
	return nil
}

// DownloadVideo downloads a video using multiple threads.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	start := time.Now()
	if util.IsDebug {
		util.Logger.Debug("DownloadVideo started", "url", url)
	}
	destPath = filepath.Clean(destPath)
	httpClient := &http.Client{
		Transport: api.SafeTransport(10 * time.Second),
	}
	chunkSize := int64(0)
	var contentLength int64
	contentLength, err := getContentLength(url, httpClient)
	if err != nil {
		return err
	}
	if contentLength == 0 {
		return fmt.Errorf("content length is zero")
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
		go func(from, to int64, part int, httpClient *http.Client) {
			defer downloadWg.Done()
			err := downloadPart(url, from, to, part, httpClient, destPath, m)
			if err != nil {
				util.Logger.Error("Download part failed", "thread", part, "error", err)
			}
		}(from, to, i, httpClient)
	}
	downloadWg.Wait()
	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}
	if util.IsDebug {
		util.Logger.Debug("DownloadVideo completed", "url", url, "duration", time.Since(start))
	}
	return nil
}

// downloadWithYtDlp downloads a video using yt-dlp.
func downloadWithYtDlp(url, path string) error {
	cmd := exec.Command("yt-dlp", "--no-progress", "-f", "best", "-o", path, url)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("yt-dlp error: %v\n%s", err, string(output))
	}
	return nil
}

// ExtractVideoSources returns the available video sources for an episode.
func ExtractVideoSources(episodeURL string) ([]struct {
	Quality int
	URL     string
}, error) {
	videoSrc, err := extractVideoURL(episodeURL)
	if err != nil {
		return nil, err
	}
	if strings.Contains(videoSrc, "animefire.plus/video/") {
		resp, err := api.SafeGet(videoSrc)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				util.Logger.Warn("Error closing response body", "error", err)
			}
		}()
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
				sources = append(sources, struct {
					Quality int
					URL     string
				}{Quality: q, URL: v.Src})
			}
			return sources, nil
		}
	}
	var respStruct struct {
		Data []struct {
			Src   string `json:"src"`
			Label string `json:"label"`
		}
	}
	if err := json.Unmarshal([]byte(videoSrc), &respStruct); err == nil && len(respStruct.Data) > 0 {
		var sources []struct {
			Quality int
			URL     string
		}
		for _, v := range respStruct.Data {
			labelDigits := regexp.MustCompile(`\d+`).FindString(v.Label)
			q := 0
			if labelDigits != "" {
				q, _ = strconv.Atoi(labelDigits)
			}
			sources = append(sources, struct {
				Quality int
				URL     string
			}{Quality: q, URL: v.Src})
		}
		return sources, nil
	}
	re := regexp.MustCompile(`(\d{3,4})p?\\.mp4`)
	matches := re.FindStringSubmatch(videoSrc)
	if len(matches) > 1 {
		q, _ := strconv.Atoi(matches[1])
		return []struct {
			Quality int
			URL     string
		}{{Quality: q, URL: videoSrc}}, nil
	}
	return []struct {
		Quality int
		URL     string
	}{{Quality: 0, URL: videoSrc}}, nil
}

// getBestQualityURL returns the best available quality for an episode.
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

// ExtractVideoSourcesWithPrompt allows the user to choose video quality.
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
		return sources[0].URL, nil
	}
	for _, s := range sources {
		if fmt.Sprintf("%dp", s.Quality) == result {
			return s.URL, nil
		}
	}
	return sources[0].URL, nil
}

// HandleBatchDownload performs batch download of episodes.
func HandleBatchDownload(episodes []models.Episode, animeURL string) error {
	start := time.Now()
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownload started", "animeURL", animeURL)
	}
	startNum, endNum, err := getEpisodeRange()
	if err != nil {
		return fmt.Errorf("invalid episode range: %w", err)
	}
	var (
		m          *model
		p          *tea.Program
		totalBytes int64
		httpClient = &http.Client{
			Transport: api.SafeTransport(10 * time.Second),
		}
		episodesToDownload []int
	)

	// First pass: check which episodes need downloading and calculate total bytes
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			util.Logger.Warn("Episode not found", "episode", episodeNum)
			continue
		}

		// Check if episode already exists
		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			util.Logger.Error("Episode path error", "episode", episodeNum, "error", err)
			continue
		}
		if fileExists(episodePath) {
			util.Logger.Info("Episode already exists", "episode", episodeNum)
			continue
		}

		// Episode needs downloading
		episodesToDownload = append(episodesToDownload, episodeNum)

		videoURL, err := getBestQualityURL(episode.URL)
		if err != nil {
			util.Logger.Warn("Skipping episode", "episode", episodeNum, "error", err)
			continue
		}
		contentLength, err := getContentLength(videoURL, httpClient)
		if err == nil {
			totalBytes += contentLength
		}
	}

	// Check if any episodes need downloading
	if len(episodesToDownload) == 0 {
		// All episodes in range already exist, offer to play one of them
		return handleExistingEpisodes(episodes, animeURL, startNum, endNum)
	}

	fmt.Printf("Found %d episode(s) to download...\n", len(episodesToDownload))

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
		sem := make(chan struct{}, 4)
		for _, epNum := range episodesToDownload {
			sem <- struct{}{}
			wg.Add(1)
			go func(epNum int) {
				defer func() {
					<-sem
					wg.Done()
				}()
				episode, found := findEpisode(episodes, epNum)
				if !found {
					util.Logger.Warn("Episode not found in batch", "episode", epNum)
					return
				}
				videoURL, err := getBestQualityURL(episode.URL)
				if err != nil {
					util.Logger.Warn("Skipping episode in batch", "episode", epNum, "error", err)
					return
				}
				episodePath, err := createEpisodePath(animeURL, epNum)
				if err != nil {
					util.Logger.Error("Episode path error", "episode", epNum, "error", err)
					return
				}

				// Double-check if file still doesn't exist (race condition protection)
				if fileExists(episodePath) {
					if p != nil {
						p.Send(statusMsg(fmt.Sprintf("Episode %d already exists, skipping...", epNum)))
					}
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
					util.Logger.Error("Failed episode download", "episode", epNum, "error", err)
				}
			}(epNum)
		}
		wg.Wait()
		// Signal that all downloads are complete
		if m != nil {
			// Send final completion message first
			if p != nil {
				p.Send(statusMsg("All downloads completed!"))
			}
			// Small delay to ensure the user sees the completion message
			time.Sleep(500 * time.Millisecond)

			m.mu.Lock()
			m.done = true
			m.mu.Unlock()
		}

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
	if util.IsDebug {
		util.Logger.Debug("HandleBatchDownload completed", "animeURL", animeURL, "duration", time.Since(start))
	}

	// Ask user which episode from the downloaded range they want to play
	return askAndPlayDownloadedEpisode(episodes, animeURL, startNum, endNum)
}

// getEpisodeRange asks the user for the episode range for download.
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

// findEpisode returns the episode struct by number.
func findEpisode(episodes []models.Episode, episodeNum int) (models.Episode, bool) {
	for _, ep := range episodes {
		if ep.Num == episodeNum {
			return ep, true
		}
	}
	return models.Episode{}, false
}

// createEpisodePath creates the file path for the downloaded episode.
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

// fileExists verifica se o arquivo existe.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// handleExistingEpisodes handles the case when all episodes in the requested range already exist
func handleExistingEpisodes(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	fmt.Printf("All episodes in range %d-%d already exist!\n\n", startNum, endNum)

	// Collect existing episodes in the range
	var existingEpisodes []models.Episode
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			continue
		}

		if fileExists(episodePath) {
			existingEpisodes = append(existingEpisodes, episode)
		}
	}

	if len(existingEpisodes) == 0 {
		fmt.Println("No downloaded episodes found in the specified range.")
		return nil
	}

	// Create options for the interactive menu
	var options []huh.Option[string]
	for _, ep := range existingEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		options = append(options, huh.NewOption(title, strconv.Itoa(ep.Num)))
	}

	// Add option to not watch anything
	options = append(options, huh.NewOption("Don't watch anything", "exit"))

	var selectedEpisode string
	err := huh.NewSelect[string]().
		Title("Which episode would you like to watch?").
		Options(options...).
		Value(&selectedEpisode).
		Run()

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

	if selectedEpisode == "exit" {
		fmt.Println("No episode selected.")
		return nil
	}

	// Find and play the selected episode
	episodeNum, err := strconv.Atoi(selectedEpisode)
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	// Verify the episode exists in our list
	_, found := findEpisode(existingEpisodes, episodeNum)
	if !found {
		return fmt.Errorf("selected episode not found")
	}

	fmt.Printf("Playing Episode %d...\n", episodeNum)

	// Get the episode path and play it
	episodePath, err := createEpisodePath(animeURL, episodeNum)
	if err != nil {
		return fmt.Errorf("failed to get episode path: %w", err)
	}

	// Play the episode using the existing player logic
	// Note: We use the local file path as the video URL since it's already downloaded
	// anilistID set to 0 since we don't have that context here, updater set to nil
	return playVideo(episodePath, episodes, episodeNum, 0, nil)
}

// askAndPlayDownloadedEpisode asks the user which episode from the downloaded range they want to play
func askAndPlayDownloadedEpisode(episodes []models.Episode, animeURL string, startNum, endNum int) error {
	// Collect downloaded episodes in the range
	var downloadedEpisodes []models.Episode
	for episodeNum := startNum; episodeNum <= endNum; episodeNum++ {
		episode, found := findEpisode(episodes, episodeNum)
		if !found {
			continue
		}

		episodePath, err := createEpisodePath(animeURL, episodeNum)
		if err != nil {
			continue
		}

		if fileExists(episodePath) {
			downloadedEpisodes = append(downloadedEpisodes, episode)
		}
	}

	if len(downloadedEpisodes) == 0 {
		fmt.Println("No downloaded episodes found in the specified range.")
		return nil
	}

	// Create options for the interactive menu
	var options []huh.Option[string]
	for _, ep := range downloadedEpisodes {
		title := fmt.Sprintf("Episode %d", ep.Num)
		if ep.Title.English != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.English)
		} else if ep.Title.Romaji != "" {
			title = fmt.Sprintf("Episode %d: %s", ep.Num, ep.Title.Romaji)
		}
		options = append(options, huh.NewOption(title, strconv.Itoa(ep.Num)))
	}

	// Add option to not watch anything
	options = append(options, huh.NewOption("Don't watch anything", "exit"))

	var selectedEpisode string
	err := huh.NewSelect[string]().
		Title("Which episode would you like to watch?").
		Options(options...).
		Value(&selectedEpisode).
		Run()

	if err != nil {
		return fmt.Errorf("episode selection error: %w", err)
	}

	if selectedEpisode == "exit" {
		fmt.Println("No episode selected.")
		return nil
	}

	// Find and play the selected episode
	episodeNum, err := strconv.Atoi(selectedEpisode)
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	// Verify the episode exists in our list
	_, found := findEpisode(downloadedEpisodes, episodeNum)
	if !found {
		return fmt.Errorf("selected episode not found")
	}

	fmt.Printf("Playing Episode %d...\n", episodeNum)

	// Get the episode path and play it
	episodePath, err := createEpisodePath(animeURL, episodeNum)
	if err != nil {
		return fmt.Errorf("failed to get episode path: %w", err)
	}

	// Play the episode using the existing player logic
	// Note: We use the local file path as the video URL since it's already downloaded
	// anilistID set to 0 since we don't have that context here, updater set to nil
	return playVideo(episodePath, episodes, episodeNum, 0, nil)
}
