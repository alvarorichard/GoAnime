package player

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	"github.com/manifoldco/promptui"
)

// Funções de download extraídas de player.go

// downloadPart baixa um pedaço do arquivo de vídeo usando HTTP Range Requests.
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
			log.Printf("Error closing response body: %v", err)
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
			log.Printf("Error closing file: %v", err)
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

// combineParts combina partes baixadas em um único arquivo final.
func combineParts(destPath string, numThreads int) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Printf("Error closing output file: %v", err)
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
				log.Printf("Error closing part file: %v", closeErr)
			}
			return err
		}
		if err := partFile.Close(); err != nil {
			log.Printf("Error closing part file: %v", err)
		}
		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}
	return nil
}

// DownloadVideo baixa um vídeo usando múltiplas threads.
func DownloadVideo(url, destPath string, numThreads int, m *model) error {
	start := time.Now()
	if util.IsDebug {
		log.Printf("[PERF] DownloadVideo iniciado para %s", url)

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
				log.Printf("Thread %d: download part failed: %v\n", part, err)
			}
		}(from, to, i, httpClient)
	}
	downloadWg.Wait()
	err = combineParts(destPath, numThreads)
	if err != nil {
		return fmt.Errorf("failed to combine parts: %v", err)
	}
	if util.IsDebug {
		log.Printf("[PERF] DownloadVideo finalizado para %s em %v", url, time.Since(start))

	}
	return nil
}

// downloadWithYtDlp baixa um vídeo usando yt-dlp.
func downloadWithYtDlp(url, path string) error {
	cmd := exec.Command("yt-dlp", "--no-progress", "-f", "best", "-o", path, url)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("yt-dlp error: %v\n%s", err, string(output))
	}
	return nil
}

// ExtractVideoSources retorna as fontes de vídeo disponíveis para um episódio.
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
				log.Printf("Error closing response body: %v", err)
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

// getBestQualityURL retorna a melhor qualidade disponível para um episódio.
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

// ExtractVideoSourcesWithPrompt permite ao usuário escolher a qualidade do vídeo.
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

// HandleBatchDownload faz o download em lote de episódios.
func HandleBatchDownload(episodes []models.Episode, animeURL string) error {
	start := time.Now()
	if util.IsDebug {
		log.Printf("[PERF] HandleBatchDownload iniciado para %s", animeURL)

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
	)
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
	if util.IsDebug {
		log.Printf("[PERF] HandleBatchDownload finalizado para %s em %v", animeURL, time.Since(start))

	}
	return nil
}

// getEpisodeRange pede ao usuário o intervalo de episódios para download.
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

// findEpisode retorna o struct do episódio pelo número.
func findEpisode(episodes []models.Episode, episodeNum int) (models.Episode, bool) {
	for _, ep := range episodes {
		if ep.Num == episodeNum {
			return ep, true
		}
	}
	return models.Episode{}, false
}

// createEpisodePath cria o caminho do arquivo para o episódio baixado.
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
