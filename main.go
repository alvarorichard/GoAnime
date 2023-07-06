package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/manifoldco/promptui"
	
)

const baseURL = "https://animefire.net"

type Anime struct {
	Name     string
	URL      string
	Episodes []Episode
}

type Episode struct {
	Number string
	URL    string
}

func main() {
	animeName := getUserInput("Enter the name of the anime")

	animeURL, err := searchAnime(animeName)
	if err != nil {
		log.Fatalf("Failed to search anime: %v", err)
	}

	episodes, err := getAnimeEpisodes(animeURL)
	if err != nil {
		log.Fatalf("Failed to get anime episodes: %v", err)
	}

	if len(episodes) == 0 {
		log.Fatalf("No episodes found for the selected anime")
	}

	episodeIndex := selectEpisode(episodes)
	selectedEpisode := episodes[episodeIndex]

	playVideo(selectedEpisode.URL)
}

func getUserInput(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Failed to get user input: %v", err)
	}
	return result
}

func searchAnime(animeName string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseURL, strings.ReplaceAll(animeName, " ", "-"))

	for {
		resp, err := http.Get(currentPageURL)
		if err != nil {
			return "", fmt.Errorf("failed to perform search request: %v", err)
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to parse search results: %v", err)
		}

		animes := make([]Anime, 0)
		doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
			anime := Anime{
				Name: strings.TrimSpace(s.Text()),
				URL:  s.AttrOr("href", ""),
			}
			animes = append(animes, anime)
		})

		if len(animes) > 0 {
			// Select the anime
			index := selectAnime(animes)
			selectedAnime := animes[index]

			return selectedAnime.URL, nil
		}

		// Check if there's a next page
		nextPageURL, exists := doc.Find(".pagination .next a").Attr("href")
		if !exists || nextPageURL == "" {
			return "", fmt.Errorf("no anime found with the given name")
		}

		currentPageURL = baseURL + nextPageURL
	}
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
		log.Fatalf("Failed to select anime: %v", err)
	}

	return index
}

func getAnimeEpisodes(animeURL string) ([]Episode, error) {
	resp, err := http.Get(animeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime details: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse anime details: %v", err)
	}

	episodesContainer := doc.Find(".container.secao.anime.secao_episodes ul li")

	episodes := make([]Episode, 0)
	episodesContainer.Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Find("div.ep_num").Text()
		episodeURL, _ := s.Find("a").Attr("href")

		episode := Episode{
			Number: strings.TrimSpace(episodeNum),
			URL:    episodeURL,
		}
		episodes = append(episodes, episode)
	})

	return episodes, nil
}

func selectEpisode(episodes []Episode) int {
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
	}

	return index
}

func playVideo(videoURL string) {
	cmd := exec.Command("vlc", videoURL)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start video player: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("Video player failed: %v", err)
	}
}
