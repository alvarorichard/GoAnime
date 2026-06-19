package scraper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadFrontendFixture returns the captured rotating-frontend serie page
// (lospobreflix.site/serie/the-boys/1, season 1, 5 seasons linked).
func loadFrontendFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "superflix_frontend_serie.html"))
	require.NoError(t, err)
	return string(data)
}

func TestParseFrontendEpisodes(t *testing.T) {
	t.Parallel()
	html := loadFrontendFixture(t)
	got := parseFrontendEpisodes(html)
	require.NotEmpty(t, got, "must parse episodes from frontend page")
	// The fixture page is season 1 with 8 episodes.
	require.Contains(t, got, "1")
	assert.Len(t, got["1"], 8)
	assert.Equal(t, "1", got["1"][0].EpiNum.String())
	assert.Equal(t, "8", got["1"][7].EpiNum.String())
}

func TestParseWindowAllEpisodes(t *testing.T) {
	t.Parallel()
	html := loadFrontendFixture(t)
	got := parseWindowAllEpisodes(html)
	require.NotEmpty(t, got, "must parse window.allEpisodes blob")
	// The blob carries all 5 seasons at once (the anchors only cover season 1).
	require.Contains(t, got, "1")
	require.Contains(t, got, "5")
	assert.Len(t, got["1"], 8)
	// Unlike the anchor path, the blob populates air_date + title.
	first := got["1"][0]
	assert.Equal(t, "1", first.EpiNum.String())
	assert.Equal(t, "O Nome do Jogo", first.Title)
	assert.Equal(t, "2019-07-25", first.AirDate)
	// Every returned episode survived the air-date filter with a real date.
	for season, eps := range got {
		for _, ep := range eps {
			assert.NotEmpty(t, ep.AirDate, "air_date populated in season %s", season)
			assert.NotEqual(t, "null", ep.AirDate)
		}
	}
}

func TestParseFrontendSeasons(t *testing.T) {
	t.Parallel()
	html := loadFrontendFixture(t)
	got := parseFrontendSeasons(html)
	// The fixture links seasons 1..5.
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, got)
}

func TestResolveFrontendSeasonURLs(t *testing.T) {
	t.Parallel()
	html := loadFrontendFixture(t)
	got := resolveFrontendSeasonURLs(html, "https://lospobreflix.site/serie/the-boys/1")
	// Seasons 1..5 resolved to absolute URLs on the solved frontend domain.
	require.Len(t, got, 5)
	assert.Equal(t, "https://lospobreflix.site/serie/the-boys/2", got["2"])
	assert.Equal(t, "https://lospobreflix.site/serie/the-boys/5", got["5"])
}

func TestExtractEpisodes_FrontendFallback(t *testing.T) {
	t.Parallel()
	c := NewSuperFlixClient()
	got, err := c.ExtractEpisodes(loadFrontendFixture(t))
	require.NoError(t, err)
	require.Contains(t, got, "1")
	assert.Len(t, got["1"], 8)
}
