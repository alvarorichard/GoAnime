package test_util_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
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

	expectedAnimes := []api.Anime{
		{Name: "Anime One", URL: "/anime/1"},
		{Name: "Anime Two", URL: "/anime/2"},
		{Name: "Anime Three", URL: "/anime/3"},
	}

	animes := api.ParseAnimes(doc)

	assert.Equal(t, expectedAnimes, animes, "Parsed animes do not match expected animes")
}
