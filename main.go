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
    os.Exit(1)
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
    os.Exit(1)
  }

  return episodes[index].Url
}

func getAnimeEpisodes(animeUrl string) ([]Episode, error){
  resp, err := http.Get(animeUrl)
  
  if err != nil {
    log.Fatalf("failed to get anime details: %v\n", err)
    os.Exit(1)
	}
	defer resp.Body.Close()
  
  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil{
    log.Fatalf("failed to parse anime details: %v\n", err)
    os.Exit(1)
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
		log.Fatalf("Failed to select anime: %v\n", err)
    os.Exit(1)
	}

  return index
}

func searchAnime(animeName string) (string, error){
  currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteUrl, animeName)

  for {
    response, err := http.Get(currentPageURL)
    if err != nil{
      log.Fatalf("failed to perform search resquest: %v\n", err)
      os.Exit(1)
    }

    defer response.Body.Close()

    doc, err := goquery.NewDocumentFromReader(response.Body)

    if err != nil{
      log.Fatalf("failed to parse response: %v\n", err)
      os.Exit(1)
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
      log.Fatalf("no anime found with the given name")
      os.Exit(1)
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
    log.Fatalf("Error from adquire user input: %v\n", err)
    os.Exit(1)
  }

  return result
}

func main(){
  animeName := getUserInput("Enter anime name")
  animeURL, err := searchAnime(treatingAnimeName(animeName))

  if err != nil{
    log.Fatalf("Failed to get anime episodes: %v", err)
    os.Exit(1)
  }

  episodes, err := getAnimeEpisodes(animeURL)

  if err != nil || len(episodes) <= 0{
    log.Fatalln("Failed to catching episodes from selected anime")
    os.Exit(1)
  }

  selectedEpisode := selectEpisode(episodes)
  PlayVideo(selectedEpisode)
}
