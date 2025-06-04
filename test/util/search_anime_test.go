package test_util_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Define a variable for the base site URL for testing purposes
var testBaseSiteURL string

func TestSearchAnime(t *testing.T) {
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
		</body>
		</html>
		`
		_, err := w.Write([]byte(page))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
			return
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Substitute the base URL with our mock server URL
	testBaseSiteURL = server.URL

	animeName := "anime"
	expectedURL := server.URL + "/anime/1"

	url, err := searchAnimeWithBaseURL(animeName, testBaseSiteURL)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
}

func searchAnimeWithBaseURL(animeName, baseURL string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseURL, animeName)

	for {
		animeURL, nextPageURL, err := searchAnimeOnPage(currentPageURL, baseURL)
		if err != nil {
			return "", err
		}
		if animeURL != "" {
			return animeURL, nil
		}
		if nextPageURL == "" {
			return "", errors.New("no anime found with the given name")
		}
		currentPageURL = baseURL + nextPageURL
	}
}

func searchAnimeOnPage(currentPageURL, baseURL string) (string, string, error) {
	resp, err := http.Get(currentPageURL)
	if err != nil {
		return "", "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Log the error but don't return it
			util.Debugf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return "", "", errors.New("failed to fetch page")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	var animeURL string
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists {
			animeURL = strings.TrimRight(baseURL, "/") + href
			return false
		}
		return true
	})

	// In this mock, we won't handle pagination
	return animeURL, "", nil
}
