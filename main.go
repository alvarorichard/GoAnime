package main

import (
	"database/sql"
  "encoding/json"
  "fmt"
  "io/ioutil"
	"log"
	"net/http"
	"os"
  "os/user"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/manifoldco/promptui"
  _ "github.com/mattn/go-sqlite3"
)

const baseSiteUrl string = "https://animefire.net"

type Episode struct {
	Number string
	Url    string
}

type Anime struct {
	Name     string
	Url      string
	Episodes []Episode
}

type VideoResponse struct {
	Data     []VideoData `json:"data"`
	Resposta struct {
		Status string `json:"status"`
		Text   string `json:"text"`
	} `json:"resposta"`
}

type VideoData struct {
	Src   string `json:"src"`
	Label string `json:"label"`
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

func initializeDB() (*sql.DB, error){
  currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
  
  dirPath := currentUser.HomeDir + "/.local/goanime"
    
  err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

  db, err := sql.Open("sqlite3", dirPath+"/anime.db")
  if err != nil{
    return nil, err
  }

  createTableSQL := `
    CREATE TABLE IF NOT EXISTS anime(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE
  );
  `

  _, err = db.Exec(createTableSQL)
  if err != nil{
    return db, nil
  }

  return db, nil
}

func addAnimeNamesToDB(db *sql.DB, animeNames []string) error {
	insertSQL := `
		INSERT OR IGNORE INTO anime (name) VALUES (?)
	`

	for _, name := range animeNames {
		_, err := db.Exec(insertSQL, name)
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
	// Check the response status code
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", response.Status)
	}

	// Convert the response body to a string
  doc, _ := goquery.NewDocumentFromReader(response.Body)

  videoElements := doc.Find("video")

  if videoElements.Length() > 0{
    oldDataVideo, _ := videoElements.Attr("data-video-src")
    return oldDataVideo, nil
  }else{
    videoElements = doc.Find("div")
    if videoElements.Length() > 0{
      oldDataVideo, _ := videoElements.Attr("data-video-src")
      return oldDataVideo, nil
    }
  }
  
	return "", nil
}

func extractActualVideoURL(videoSrc string) (string,error) {
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

func selectEpisode(episodes []Episode) string {
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

	return episodes[index].Url
}

func getAnimeEpisodes(animeUrl string) ([]Episode, error) {
	resp, err := http.Get(animeUrl)

	if err != nil {
		log.Fatalf("Failed to get anime details: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse anime details: %v\n", err)
		os.Exit(1)
	}

	episodeContainer := doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex")

	episodes := make([]Episode, 0)

	episodeContainer.Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		episode := Episode{
			Number: episodeNum,
			Url:    episodeURL,
		}
		episodes = append(episodes, episode)
	})
	return episodes, nil
}

func selectAnime(db *sql.DB,animes []Anime) int {

  animesName := make([]string, 0)
  for i := range(animes){
    animesName = append(animesName, animes[i].Name) 
  }
  addAnimeNamesToDB(db, animesName)
	
  templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▶ {{ .Name | cyan }}",
		Inactive: "  {{ .Name | white }}",
		Selected: "▶ {{ .Name | cyan | underline }}",
	}

	prompt := promptui.Select{
		Label:     "Select the anime",
		Items:     animes,
		Templates: templates,
	}
  
  
	index, _, err := prompt.Run()

	if err != nil {
		log.Fatalf("Failed to select anime: %v\n", err)
		os.Exit(1)
	}

	return index
}

func searchAnime(db *sql.DB,animeName string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteUrl, animeName)

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
				Url:  s.AttrOr("href", ""),
			}
      animeName = strings.TrimSpace(s.Text())

			animes = append(animes, anime)
		})

		if len(animes) > 0 {
			index := selectAnime(db, animes)
			selectedAnime := animes[index]

			return selectedAnime.Url, nil
		}

		nextPage, exists := doc.Find(".pagination .next a").Attr("href")
		if !exists || nextPage == "" {
			log.Fatalln("No anime found with the given name")
			os.Exit(1)
		}

		currentPageURL = baseSiteUrl + nextPage
    if err != nil{
      log.Fatalf("Failed to add anime names to the database: %v", err)
    }
	}
}

func treatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	spacelessName := strings.ReplaceAll(loweredName, " ", "-")
	return spacelessName
}

func getUserInput(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Error acquiring user input: %v\n", err)
		os.Exit(1)
	}

	return result
}

func main() {
  db, err := initializeDB()
  if err != nil{
    log.Fatal(err)
  }

  defer db.Close()

	animeName := getUserInput("Enter anime name")
	animeURL, err := searchAnime(db,treatingAnimeName(animeName))
  
	if err != nil {
		log.Fatalf("Failed to get anime episodes: %v", err)
		os.Exit(1)
	}
  
	episodes, err := getAnimeEpisodes(animeURL)

	if err != nil || len(episodes) <= 0 {
		log.Fatalln("Failed to fetch episodes from selected anime")
		os.Exit(1)
	}
  
	selectedEpisodeURL := selectEpisode(episodes)
	videoURL,err := extractVideoURL(selectedEpisodeURL)
	
  if err != nil {
		log.Fatalf("Failed to extract video URL: %v", err)
	}
  
  videoURL, err = extractActualVideoURL(videoURL)

  if err != nil {
    log.Fatal("Failed to extract the api")
  }

  PlayVideo(videoURL)
}
