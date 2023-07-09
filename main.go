package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/manifoldco/promptui"
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

func extractVideoUrl(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make the request: %v\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Fatalf("Failed to fetch the URL, status code: %d", resp.StatusCode)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse the document: %v", err)
		os.Exit(1)
	}

	var videoURL string

	doc.Find("div#main_div_video video#my-video").Each(func(i int, s *goquery.Selection) {
		videoURL, _ = s.Attr("data-video-src")
	})
	return videoURL
}

func PlayVideo(videoURL string) {
	cmd := exec.Command("vlc", "-vvv", videoURL)
	err := cmd.Start()

	if err != nil {
		log.Fatalf("Failed to start video player: %v", err)
		os.Exit(1)
	}

	err = cmd.Wait()

	if err != nil {
		log.Fatalf("Failed to play video: %v", err)
		os.Exit(1)
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

func selectAnime(animes []Anime) int {
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

func searchAnime(animeName string) (string, error) {
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

			animes = append(animes, anime)
		})

		if len(animes) > 0 {
			index := selectAnime(animes)
			selectedAnime := animes[index]

			return selectedAnime.Url, nil
		}

		nextPage, exists := doc.Find(".pagination .next a").Attr("href")
		if !exists || nextPage == "" {
			log.Fatalln("No anime found with the given name")
			os.Exit(1)
		}

		currentPageURL = baseSiteUrl + nextPage
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
	animeName := getUserInput("Enter anime name")
	animeURL, err := searchAnime(treatingAnimeName(animeName))

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
	videoURL := extractVideoUrl(selectedEpisodeURL)

	if videoURL == "" {
		log.Fatalln("No video URL found")
		os.Exit(1)
	}

	PlayVideo(videoURL)
}
