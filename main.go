package main

import (
	//"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	//"time"
	"io"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	//"github.com/cavaliergopher/grab/v3"
	"github.com/cheggaaa/pb/v3"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ktr0731/go-fuzzyfinder"
)

const baseSiteURL string = "https://animefire.plus/"

type Episode struct {
	Number string
	URL    string
}

type Anime struct {
	Name     string
	URL      string
	Episodes []Episode
}

type VideoResponse struct {
	Data []VideoData `json:"data"`
}

type VideoData struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

func databaseFormatter(str string) string {
	regex := regexp.MustCompile(`\s*\([^)]*\)|\bn/a\b|\s+\d+(\.\d+)?$`)
	result := regex.ReplaceAllString(str, "")
	result = strings.TrimSpace(result)
	result = strings.ToLower(result)
	return result
}

func DownloadFolderFormatter(str string) string {
	regex := regexp.MustCompile(`https:\/\/animefire\.plus\/video\/([^\/?]+)`)
	match := regex.FindStringSubmatch(str)
	if len(match) > 1 {
		finalStep := match[1]
		return finalStep
	}
	return ""
}

func listAnimeNamesFromDB(db *sql.DB) error {
	query := `
		SELECT name FROM anime
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Anime names:")
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return err
		}
		fmt.Println(name)
	}

	return nil
}

func initializeDB() (*sql.DB, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	dirPath := filepath.Join(currentUser.HomeDir, ".local", "goanime")

	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", filepath.Join(dirPath, "anime.db"))
	if err != nil {
		return nil, err
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS anime(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE
  );
  `

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return db, nil
	}

	return db, nil
}

func addAnimeNamesToDB(db *sql.DB, animeNames []string) error {
	insertSQL := `
		INSERT OR IGNORE INTO anime (name) VALUES (?)
	`

	for _, name := range animeNames {
		_, err := db.Exec(insertSQL, databaseFormatter(name))
		if err != nil {
			return err
		}
	}

	return nil
}

func extractVideoURL(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(response.Body)

	videoElements := doc.Find("video")

	if videoElements.Length() > 0 {
		oldDataVideo, _ := videoElements.Attr("data-video-src")
		return oldDataVideo, nil
	} else {
		videoElements = doc.Find("div")
		if videoElements.Length() > 0 {
			oldDataVideo, _ := videoElements.Attr("data-video-src")
			return oldDataVideo, nil
		}
	}

	return "", nil
}

func extractActualVideoURL(videoSrc string) (string, error) {
	fmt.Println(videoSrc)
	response, err := http.Get(videoSrc)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var videoResponse VideoResponse
	err = json.Unmarshal(body, &videoResponse)
	if err != nil {
		return "", err
	}

	if len(videoResponse.Data) == 0 {
		return "", fmt.Errorf("no video data found")
	}

	return videoResponse.Data[0].Src, nil
}

func PlayVideo(videoURL string) {
	cmd := exec.Command("vlc", "-vvv", videoURL)
	err := cmd.Start()

	if err != nil {
		log.Fatalf("Failed to start video player: %v", err)
	}

	err = cmd.Wait()

	if err != nil {
		log.Fatalf("Failed to play video: %v", err)
	}
}

// func selectWithFZF(items []string) (string, error) {
// 	cmd := exec.Command("go-fuzzyfinder")
// 	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))

// 	var out bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &stderr

// 	err := cmd.Run()
// 	if err != nil {
// 		return "", fmt.Errorf("Failed to select item with FZF: %v. Stderr: %s", err, stderr.String())
// 	}

// 	return strings.TrimSpace(out.String()), nil
// }

func selectWithGoFuzzyFinder(items []string) (string, error) {
	idx, err := fuzzyfinder.Find(
		items,
		func(i int) string {
			return items[i]
		},
	)
	if err != nil {
		return "", fmt.Errorf("Failed to select item with go-fuzzyfinder: %v", err)
	}

	return items[idx], nil
}

func selectAnimeWithGoFuzzyFinder(animes []Anime) string {
	animeNames := make([]string, len(animes))
	for i, anime := range animes {
		animeNames[i] = anime.Name
	}

	idx, err := fuzzyfinder.Find(
		animeNames,
		func(i int) string {
			return animeNames[i]
		},
	)
	if err != nil {
		log.Fatalf("Failed to select anime with go-fuzzyfinder: %v", err)
	}

	return animes[idx].Name
}

func searchAnime(db *sql.DB, animeName string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, animeName)

	for {
		response, err := http.Get(currentPageURL)
		if err != nil {
			log.Fatalf("Failed to perform search request: %v\n", err)
			os.Exit(1)
		}

		defer response.Body.Close()

		doc, err := goquery.NewDocumentFromReader(response.Body)

		if err != nil {
			log.Fatalf("Failed to parse response: %v\n", err)
			os.Exit(1)
		}

		animes := make([]Anime, 0)
		doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
			anime := Anime{
				Name: strings.TrimSpace(s.Text()),
				URL:  s.AttrOr("href", ""),
			}
			animeName = strings.TrimSpace(s.Text())

			animes = append(animes, anime)
		})

		if len(animes) > 0 {
			selectedAnimeName := selectAnimeWithGoFuzzyFinder(animes)
			for _, anime := range animes {
				if anime.Name == selectedAnimeName {
					return anime.URL, nil
				}
			}
		}

		nextPage, exists := doc.Find(".pagination .next a").Attr("href")
		if !exists || nextPage == "" {
			log.Fatalln("No anime found with the given name")
			os.Exit(1)
		}

		currentPageURL = baseSiteURL + nextPage
		if err != nil {
			log.Fatalf("Failed to add anime names to the database: %v", err)
		}
	}
}

func treatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	spacelessName := strings.ReplaceAll(loweredName, " ", "-")
	return spacelessName
}

func getUserInput(label string, db *sql.DB) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Error acquiring user input: %v", err)
		os.Exit(1)
	}

	return result
}

func askForDownload() bool {
	prompt := promptui.Select{
		Label: "Do you want to download the episode",
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Error acquiring user input: %v", err)
		os.Exit(1)
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
		log.Fatalf("Error acquiring user input: %v", err)
		os.Exit(1)
	}

	return strings.ToLower(result) == "yes"
}
func getAnimeEpisodes(animeURL string) ([]Episode, error) {
	resp, err := http.Get(animeURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to get anime details: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse anime details: %v", err)
	}

	episodeContainer := doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex")

	var episodes []Episode
	episodeContainer.Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		episode := Episode{
			Number: episodeNum,
			URL:    episodeURL,
		}
		episodes = append(episodes, episode)
	})
	return episodes, nil
}

func selectEpisode(episodes []Episode) (string, string) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▶ {{ .Number | cyan }}",
		Inactive: "  {{ .Number | white }}",
		Selected: "▶ {{ .Number | cyan | underline }}",
	}

	prompt := promptui.Select{
		Label:     "Select the episode",
		Items:     episodes,
		Templates: templates,
	}

	index, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Failed to select episode: %v", err)
		os.Exit(1)
	}

	return episodes[index].URL, episodes[index].Number
}

// func DownloadVideo(urls []string, destPath string) error {
// 	var wg sync.WaitGroup
// 	client := grab.NewClient()

// 	for _, url := range urls {
// 		req, _ := grab.NewRequest(destPath, url)
// 		resp := client.Do(req)

// 		// Crie uma nova barra de progresso para cada download.
// 		bar := pb.Full.Start64(resp.Size())
// 		bar.Set("prefix", fmt.Sprintf("Downloading %s: ", filepath.Base(req.URL().Path)))

// 		wg.Add(1)
// 		go func(resp *grab.Response, bar *pb.ProgressBar) {
// 			defer wg.Done()
// 			defer bar.Finish() // Finalize a barra ao completar este download.

// 			ticker := time.NewTicker(500 * time.Millisecond)
// 			defer ticker.Stop()

// 			for {
// 				select {
// 				case <-ticker.C:
// 					bar.SetCurrent(resp.BytesComplete())
// 				case <-resp.Done:
// 					return
// 				}
// 			}
// 		}(resp, bar) // Passe a barra como argumento para a goroutine.
// 	}

// 	wg.Wait()

// 	return nil
// }

func DownloadVideo(url string, destPath string, numThreads int) error {
    // Get the size of the file
    resp, err := http.Head(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Check if the server supports partial content
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
        return fmt.Errorf("server does not support partial content: status code %d", resp.StatusCode)
    }

    // Get the content length from the header
    contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
    if err != nil {
        return err
    }

    // Calculate size of each chunk
    chunkSize := contentLength / numThreads
    var wg sync.WaitGroup

    // Create progress bars
    bars := make([]*pb.ProgressBar, numThreads)
    for i := range bars {
        bars[i] = pb.Full.Start64(int64(chunkSize))
    }
    pool, err := pb.StartPool(bars...)
    if err != nil {
        return err
    }

    for i := 0; i < numThreads; i++ {
        from := i * chunkSize
        to := from + chunkSize - 1
        if i == numThreads-1 {
            to = contentLength - 1
        }

        wg.Add(1)
        go func(from, to, part int, bar *pb.ProgressBar) {
            defer wg.Done()

            req, err := http.NewRequest("GET", url, nil)
            if err != nil {
                log.Printf("Thread %d: error creating request: %v\n", part, err)
                return
            }
            rangeHeader := fmt.Sprintf("bytes=%d-%d", from, to)
            req.Header.Add("Range", rangeHeader)

            client := &http.Client{}
            resp, err := client.Do(req)
            if err != nil {
                log.Printf("Thread %d: error on request: %v\n", part, err)
                return
            }
            defer resp.Body.Close()

            partFilePath := fmt.Sprintf("%s.part%d", destPath, part)
            file, err := os.Create(partFilePath)
            if err != nil {
                log.Printf("Thread %d: error creating file: %v\n", part, err)
                return
            }
            defer file.Close()

            buf := make([]byte, 1024) // Buffer for copying
            for {
                n, err := resp.Body.Read(buf)
                if n > 0 {
                    _, writeErr := file.Write(buf[:n])
                    if writeErr != nil {
                        log.Printf("Thread %d: error writing to file: %v\n", part, writeErr)
                        return
                    }
                    bar.Add(n)
                }
                if err == io.EOF {
                    break
                }
                if err != nil {
                    log.Printf("Thread %d: error reading response body: %v\n", part, err)
                    return
                }
            }
            bar.Finish()
        }(from, to, i, bars[i])
    }

    wg.Wait()
    pool.Stop()

    // Combine the file parts
    outFile, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer outFile.Close()

    for i := 0; i < numThreads; i++ {
        partFilePath := fmt.Sprintf("%s.part%d", destPath, i)
        partFile, err := os.Open(partFilePath)
        if err != nil {
            return err
        }

        _, err = io.Copy(outFile, partFile)
        partFile.Close()
        os.Remove(partFilePath) // Clean up part file

        if err != nil {
            return err
        }
    }

    return nil
}



func main() {
	db, err := initializeDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	cyan := "\033[38;5;50m"
	reset := "\033[0m"

	fmt.Println(cyan + `
	$$$$$$\             $$$$$$\             $$\                         
	$$  __$$\           $$  __$$\           \__|                        
	$$ /  \__| $$$$$$\  $$ /  $$ |$$$$$$$\  $$\ $$$$$$\$$$$\   $$$$$$\  
	$$ |$$$$\ $$  __$$\ $$$$$$$$ |$$  __$$\ $$ |$$  _$$  _$$\ $$  __$$\ 
	$$ |\_$$ |$$ /  $$ |$$  __$$ |$$ |  $$ |$$ |$$ / $$ / $$ |$$$$$$$$ |
	$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |$$ | $$ | $$ |$$   ____|
	\$$$$$$  |\$$$$$$  |$$ |  $$ |$$ |  $$ |$$ |$$ | $$ | $$ |\$$$$$$$\ 
	 \______/  \______/ \__|  \__|\__|  \__|\__|\__| \__| \__| \_______|
	` + reset)

	animeName := getUserInput("Enter anime name", db)
	animeURL, err := searchAnime(db, treatingAnimeName(animeName))

	if err != nil {
		log.Fatalf("Failed to get anime episodes: %v", err)
		os.Exit(1)
	}

	episodes, err := getAnimeEpisodes(animeURL)

	if err != nil || len(episodes) <= 0 {
		log.Fatalln("Failed to fetch episodes from selected anime")
		os.Exit(1)
	}

	selectedEpisodeURL, episodeNumber := selectEpisode(episodes)

	videoURL, err := extractVideoURL(selectedEpisodeURL)

	if err != nil {
		log.Fatalf("Failed to extract video URL: %v", err)
	}

	videoURL, err = extractActualVideoURL(videoURL)

	if err != nil {
		log.Fatal("Failed to extract the api")
	}

	if askForDownload() {
		currentUser, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to get current user: %v", err)
		}
	
		downloadPath := filepath.Join(currentUser.HomeDir, ".local", "goanime", "downloads", "anime", DownloadFolderFormatter(animeURL))
		episodePath := filepath.Join(downloadPath, episodeNumber+".mp4")
	
		// Create download directory if it doesn't exist
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			os.MkdirAll(downloadPath, os.ModePerm)
		}
	      // teste
		_, err = os.Stat(episodePath)
		if os.IsNotExist(err) {
			numThreads := 4 // Set the number of threads for downloading
			err = DownloadVideo(videoURL, episodePath, numThreads)
			if err != nil {
				log.Fatalf("Failed to download video: %v", err)
			}
			fmt.Println("Video downloaded successfully!")
		}
	
		if askForPlayOffline() {
			PlayVideo(episodePath)
		}
	}else {
		PlayVideo(videoURL)
	}
}