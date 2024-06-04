package util

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type Anime struct {
	Name string
	URL  string
}

func TestSearchAnimeOnPage(t *testing.T) {
	// Mock server setup
	handler := http.NewServeMux()
	handler.HandleFunc("/pesquisar/", func(w http.ResponseWriter, r *http.Request) {
		page := `
		<html>
		<head><title>Search Results</title></head>
		<body>
			<a href="/anime/1">Anime 1</a>
			<a href="/anime/2">Anime 2</a>
			<a href="/anime/3">Anime 3</a>
			<div class="pagination">
				<a class="next" href="/pesquisar/nextpage">Next</a>
			</div>
		</body>
		</html>
		`
		w.Write([]byte(page))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	url, nextPage, err := searchAnimeOnPage(server.URL + "/pesquisar/")
	assert.NoError(t, err)
	assert.Equal(t, server.URL+"/anime/1", url)
	assert.Equal(t, server.URL+"/pesquisar/nextpage", nextPage)
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
			return "", "", errors.New("Connection refused: You need to be in Brazil or use a VPN to access the server.")
		}
		return "", "", errors.Errorf("Search failed, the server returned the error: %s", response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse response")
	}

	animes := parseAnimes(doc)
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

	nextPage, exists := doc.Find(".pagination .next").Attr("href")
	if !exists {
		return "", "", nil
	}

	return "", nextPage, nil
}

// parseAnimes and selectAnimeWithGoFuzzyFinder are mocks for the actual implementation
func parseAnimes(doc *goquery.Document) []Anime {
	var animes []Anime
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			animes = append(animes, Anime{
				Name: s.Text(),
				URL:  href,
			})
		}
	})
	return animes
}

func selectAnimeWithGoFuzzyFinder(animes []Anime) (string, error) {
	// For testing purposes, we select the first anime
	if len(animes) > 0 {
		return animes[0].Name, nil
	}
	return "", errors.New("no animes found")
}
