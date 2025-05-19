package test_util_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseAnimes(t *testing.T) {
	html := `
        <div class="row ml-1 mr-1">
            <a href="/anime/1">Anime One</a>
            <a href="/anime/2">Anime Two</a>
            <a href="/anime/3">Anime Three</a>
        </div>
    `
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document from HTML string: %v", err)
	}

	// Set the base URL for absolute URL resolution
	doc.Url, _ = url.Parse("https://animefire.plus")

	expectedAnimes := []models.Anime{
		{Name: "Anime One", URL: "https://animefire.plus/anime/1"},
		{Name: "Anime Two", URL: "https://animefire.plus/anime/2"},
		{Name: "Anime Three", URL: "https://animefire.plus/anime/3"},
	}

	animes := api.ParseAnimes(doc)

	assert.Equal(t, expectedAnimes, animes, "Parsed animes do not match expected animes")
}
