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

const baseSiteUrl string = "https://animefire.net"

type Episode struct {
  Number string
  Url string
}

type Anime struct {
  Name string
  Url string
  Episodes []Episode
}

func PlayVideo(Url string){
  cmd := exec.Command("vlc", "-vvv", Url)
  err := cmd.Run()

  if err != nil{
    log.Fatalf("Failed to start video player: %v", err)
    return
  }
}

func selectEpisode(episodes []Episode) string{
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
  if err != nil{
    log.Fatalf("Failed to select episode: %v", err)
    return ""
  }

  return episodes[index].Url
}

func getAnimeEpisodes(animeUrl string) ([]Episode, error){
  resp, err := http.Get(animeUrl)
  
  if err != nil {
		return nil, fmt.Errorf("failed to get anime details: %v", err)
	}
	defer resp.Body.Close()
  
  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil{
    return nil, fmt.Errorf("failed to parse anime details: %v", err)
  }

  episodeContainer := doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex")
  
  Episodes := make([]Episode, 0)

  episodeContainer.Each(func(i int, s* goquery.Selection){
    episodeNum := s.Text()
    episodeURL, _ := s.Attr("href")

    episode := Episode{
      Number: episodeNum,
      Url: episodeURL,
    }
    Episodes = append(Episodes, episode)
  })
  return Episodes, nil
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
    return 0
	}

  return index
}

func searchAnime(animeName string) (string, error){
  currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteUrl, animeName)

  for {
    response, err := http.Get(currentPageURL)
    if err != nil{
      return "", fmt.Errorf("failed to perform search resquest: %v", err)
    }

    defer response.Body.Close()

    doc, err := goquery.NewDocumentFromReader(response.Body)

    if err != nil{
      return "", fmt.Errorf("failed to parse response: %v", err)
    }

    animes := make([]Anime, 0)

    doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection){
      anime := Anime{
        Name: strings.TrimSpace(s.Text()),
        Url: s.AttrOr("href", ""),
      }

      animes = append(animes, anime)
    })

    if len(animes) > 0{
      index := selectAnime(animes)
      selectedAnime := animes[index]

      return selectedAnime.Url, nil
    }

    nextPage, exists := doc.Find(".pagination .next a").Attr("href")
    if !exists || nextPage == ""{
      return "", fmt.Errorf("no anime found with the given name")
    }

    currentPageURL = baseSiteUrl + nextPage
  }
}

func treatingAnimeName(animeName string) string{
  loweredName := strings.ToLower(animeName)
  spacelessName := strings.ReplaceAll(loweredName," ", "-")
  return spacelessName
}

func getUserInput(label string) string{
  prompt := promptui.Prompt{
    Label: label,
  }

  result, err := prompt.Run()

  if err != nil{
    log.Fatal(err)
    return ""
  }

  return result
}

func main(){
  animeName := getUserInput("Enter anime name")
  animeURL, err := searchAnime(treatingAnimeName(animeName))

  if err != nil{
    log.Fatalf("Failed to get anime episodes: %v", err)
    return
  }

  episodes, err := getAnimeEpisodes(animeURL)

  if err != nil || len(episodes) <= 0{
    log.Fatalln("Failed to catching episodes from selected anime")
    return
  }

  selectedEpisode := selectEpisode(episodes)
  PlayVideo(selectedEpisode)
}
