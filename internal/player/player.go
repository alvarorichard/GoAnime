package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	neturl "net/url"
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
	"github.com/cheggaaa/pb/v3"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

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
	// Extract numeric part of the quality labels (assuming format "720p", "1080p", etc.)
	quality1Value, _ := strconv.Atoi(strings.TrimRight(quality1, "p"))
	quality2Value, _ := strconv.Atoi(strings.TrimRight(quality2, "p"))
	return quality1Value > quality2Value
}

func isValidURL(url string) bool {
	// Parse the URL to check for validity and to extract the hostname
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return false
	}

	// Check if the hostname is an IP address
	ip := net.ParseIP(parsedURL.Hostname())
	if ip != nil {
		// If it's an IP address, check if it's disallowed
		return !api.IsDisallowedIP(ip.String())
	}

	// If the hostname is not an IP address, it's considered valid for this example
	// You might want to add additional checks here depending on your requirements
	return true
}

// downloadVideo downloads a video from a URL to a destination path using multiple threads
func downloadVideo(url string, destPath string, numThreads int) error {
	// Certifique-se de que o caminho de destino é validado para evitar a travessia de diretório
	destPath = filepath.Clean(destPath)

	// Crie um cliente HTTP seguro usando api.SafeTransport
	const clientConnectTimeout = 10 * time.Second
	httpClient := &http.Client{
		Transport: api.SafeTransport(clientConnectTimeout),
	}

	// Faça uma solicitação HEAD para obter o tamanho do conteúdo
	resp, err := httpClient.Head(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	// Verifique se o servidor suporta conteúdo parcial
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return errors.New(fmt.Sprintf("o servidor não suporta conteúdo parcial: código de status %d", resp.StatusCode))
	}

	// Obtenha o tamanho do conteúdo
	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	// Calcule o tamanho de cada parte
	chunkSize := contentLength / numThreads
	var wg sync.WaitGroup

	// Crie barras de progresso para cada parte
	bars := make([]*pb.ProgressBar, numThreads)
	for i := range bars {
		bars[i] = pb.Full.Start64(int64(chunkSize))
	}
	pool, err := pb.StartPool(bars...)
	if err != nil {
		return err
	}

	// Faça o download de cada parte em uma goroutine separada
	for i := 0; i < numThreads; i++ {
		from := i * chunkSize
		to := from + chunkSize - 1
		if i == numThreads-1 {
			to = contentLength - 1
		}

		wg.Add(1)
		go func(from, to, part int, bar *pb.ProgressBar) {
			defer wg.Done()

			if !isValidURL(url) {
				if util.IsDebug {
					log.Printf("Thread %d: unsafe URL detected, aborting request\n", part)
				}
				return
			}

			// Crie uma solicitação GET com um cabeçalho Range
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				if util.IsDebug {
					log.Printf("Thread %d: erro ao criar a solicitação: %v\n", part, err)
				}
				return
			}
			rangeHeader := fmt.Sprintf("bytes=%d-%d", from, to)
			req.Header.Add("Range", rangeHeader)

			// Faça a solicitação usando o cliente HTTP seguro
			resp, err := httpClient.Do(req)
			if err != nil {
				if util.IsDebug {
					log.Printf("Thread %d: erro na solicitação: %v\n", part, err)
				}
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Printf("Thread %d: erro ao fechar o corpo da resposta: %v\n", part, err)
				}
			}(resp.Body)

			// Crie um arquivo para a parte baixada
			partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), part)
			partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
			file, err := os.Create(partFilePath)
			if err != nil {
				if util.IsDebug {
					log.Printf("Thread %d: erro ao criar o arquivo: %v\n", part, err)
				}
				return
			}

			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Printf("Thread %d: erro ao fechar o arquivo: %v\n", part, err)
				}
			}(file)

			// Escreva os dados no arquivo e atualize a barra de progresso
			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					_, writeErr := file.Write(buf[:n])
					if writeErr != nil {
						if util.IsDebug {
							log.Printf("Thread %d: erro ao escrever no arquivo: %v\n", part, writeErr)
						}
						return
					}
					bar.Add(n)
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					if util.IsDebug {
						log.Printf("Thread %d: erro ao ler o corpo da resposta: %v\n", part, err)
					}
					return
				}
			}
			bar.Finish()
		}(from, to, i, bars[i])
	}

	// Aguarde todas as goroutines terminarem
	wg.Wait()
	err = pool.Stop()
	if err != nil {
		return err
	}

	// Combine todas as partes em um único arquivo
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err = outFile.Close()
		if err != nil {
			log.Printf("Erro ao fechar o arquivo de saída: %v\n", err)
		}
	}(outFile)

	for i := 0; i < numThreads; i++ {
		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)

		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, partFile)
		err = partFile.Close()
		if err != nil {
			return err
		}
		err = os.Remove(partFilePath)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}

	return nil
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

// ExtractEpisodeNumber extracts the episode number from the episode string
func ExtractEpisodeNumber(episodeStr string) string {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeStr)
	if numStr == "" {
		return "1" // Retorna "1" para filmes/OVAs
	}
	return numStr
}

// extractVideoURL extracts the video URL from the HTML of the episode page
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
	if videoSrc == "" { // If the data-video-src attribute is not found, try to find the video source URL in the Blogger video player
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

// fetchContent fetches the content to send it to the findBloggerLink function
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

// findBloggerLink extracts the video link for Blogger uploaded videos
func findBloggerLink(content string) (string, error) {
	// Regex to match the link pattern
	pattern := `https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 0 {
		// matches[0] would contain the whole matched string
		return matches[0], nil
	} else {
		return "", errors.New("no blogger video link found in the content")
	}
}

// GetVideoURLForEpisode extracts the actual video URL from the video source URL
func GetVideoURLForEpisode(episodeURL string) (string, error) {
	// Assuming extractVideoURL and extractActualVideoURL functions are defined elsewhere
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
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

	// Function to compare video quality labels and return the highest quality video URL
	highestQualityVideoURL := selectHighestQualityVideo(videoResponse.Data)
	if highestQualityVideoURL == "" {
		return "", errors.New("no suitable video quality found")
	}

	return highestQualityVideoURL, nil
}

// askForDownload asks the user if they want to download the episode or play it online
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

// downloadFolderFormatter formats the anime URL to be used as the download folder name
func downloadFolderFormatter(str string) string {
	regex := regexp.MustCompile(`https://animefire\.plus/video/([^/?]+)`)
	match := regex.FindStringSubmatch(str)
	if len(match) > 1 {
		finalStep := match[1]
		return finalStep
	}
	return ""
}

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(videoURL string, episodes []api.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string) {
	if askForDownload() {
		currentUser, err := user.Current()
		if err != nil {
			log.Panicln("Failed to get current user:", util.ErrorHandler(err))
		}

		downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", downloadFolderFormatter(animeURL))
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
				if err := downloadVideo(videoURL, episodePath, numThreads); err != nil {
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

// playVideo plays the video using the VLC player and allows the user to navigate between episodes
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

	// Find the index of the current episode based on Num
	currentEpisodeIndex := -1
	for i, ep := range episodes {
		if ep.Num == currentEpisodeNum {
			currentEpisodeIndex = i
			break
		}
	}

	// If the current episode was not found, return an error or handle appropriately
	if currentEpisodeIndex == -1 {
		if util.IsDebug {
			log.Printf("Current episode number %d not found", currentEpisodeNum)
		}
		return errors.New("current episode not found")
	}

	// Command listener for navigating episodes
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
				wg.Wait() // Wait for the current video to stop
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
				wg.Wait() // Wait for the current video to stop
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
