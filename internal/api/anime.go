package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

const baseSiteURL = "https://animefire.plus/"

type Anime struct {
	Name     string
	URL      string
	Episodes []Episode
}

type Episode struct {
	Number string
	Num    int
	URL    string
}

func SearchAnime(animeName string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, animeName)
	if util.IsDebug {
		log.Printf("Searching for anime with URL: %s", currentPageURL)
	}

	for {
		animeURL, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			return "", err
		}
		if animeURL != "" {
			return animeURL, nil
		}
		if nextPageURL == "" {
			return "", errors.New("no anime found with the given name")
		}
		currentPageURL = baseSiteURL + nextPageURL
	}
}

func searchAnimeOnPage(url string) (string, string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to perform search request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusForbidden {
			return "", "", errors.New("Connection refused: You need be in Brazil or use a VPN to access the server.")
		}
		return "", "", errors.Errorf("Search failed, the server returned the error: %s", response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse response")
	}

	animes := ParseAnimes(doc)
	if len(animes) > 0 {
		selectedAnimeName, err := selectAnimeWithGoFuzzyFinder(animes)
		if err != nil {
			return "", "", err
		}
		for _, anime := range animes {
			if anime.Name == selectedAnimeName {
				return anime.URL, "", nil
			}
		}
	}

	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
	if !exists {
		return "", "", nil
	}

	return "", nextPage, nil
}

func sortAnimes(animeList []Anime) []Anime {
	sort.Slice(animeList, func(i, j int) bool {
		return animeList[i].Name < animeList[j].Name
	})

	return animeList
}

func ParseAnimes(doc *goquery.Document) []Anime {
	var animes []Anime
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		animes = append(animes, Anime{
			Name: strings.TrimSpace(s.Text()),
			URL:  s.AttrOr("href", ""),
		})
	})
	return animes
}

func selectAnimeWithGoFuzzyFinder(animes []Anime) (string, error) {
	if len(animes) == 0 {
		return "", errors.New("no anime provided")
	}

	animeNames := make([]string, len(animes))
	for i, anime := range sortAnimes(animes) {
		animeNames[i] = anime.Name
	}

	idx, err := fuzzyfinder.Find(
		animeNames,
		func(i int) string {
			return animeNames[i]
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
	}

	if idx < 0 || idx >= len(animes) {
		return "", errors.New("invalid index returned by fuzzyfinder")
	}

	return animes[idx].Name, nil
}
