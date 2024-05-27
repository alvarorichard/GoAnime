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
	"github.com/w1tchCrafter/arrays/pkg/arrays"
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

	animes := parseAnimes(doc)
	if animes.Len() > 0 {
		selectedAnimeName, err := selectAnimeWithGoFuzzyFinder(animes)
		if err != nil {
			return "", "", err
		}
		
    findAnime, _ := animes.Find(func(i int, a Anime) bool {
      return a.Name == selectedAnimeName
    })

    return findAnime.URL, "", nil
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

func parseAnimes(doc *goquery.Document) arrays.Array[Anime] {
	animes := arrays.New[Anime]()
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		animes.Push(Anime{
			Name: strings.TrimSpace(s.Text()),
			URL:  s.AttrOr("href", ""),
		})
	})

	return animes
}

func selectAnimeWithGoFuzzyFinder(animes arrays.Array[Anime]) (string, error) {
	if animes.Len() == 0 {
		return "", errors.New("no anime provided")
	}

  slicedAnimes, _ := animes.ToSlice(arrays.FULL_COPY) 

	animeNames := arrays.New[string]()
	for _, anime := range sortAnimes(slicedAnimes) {
		animeNames.Push(anime.Name)
	}

	slicedAnimeName, _ := animeNames.ToSlice(arrays.FULL_COPY)

	idx, err := fuzzyfinder.Find(
		slicedAnimeName,
		func(i int) string {
			selected, _ := animeNames.At(i)
			return selected
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
	}

	if idx < 0 || idx >= animes.Len() {
		return "", errors.New("invalid index returned by fuzzyfinder")
	}

  animesSelected, err :=  animes.At(idx)

  return animesSelected.Name, err
}
