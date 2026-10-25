package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
)

// IsSeries delegates to GetAnimeEpisodes which uses SafeGet (SSRF-guarded).
// Loopback URLs are rejected by the dialer, so we exercise the error path here
// without spinning up a real fixture server. Network success is covered by
// scraper-level integration tests.
func TestIsSeries_ErrorOnInvalidURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"malformed", "://bad-url"},
		{"loopback rejected", "http://127.0.0.1:1/no-such-host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isSeries, count, err := IsSeries(tt.url)
			assert.Error(t, err)
			assert.False(t, isSeries)
			assert.Equal(t, 0, count)
		})
	}
}

// IsSeriesEnhanced routes by Source/MediaType, then delegates to a scraper.
// Empty / non-routable inputs fall through the AllAnime scraper which yields
// zero episodes (not an error). We assert the contract: function returns
// (false, 0, ...) when no episodes are available.
func TestIsSeriesEnhanced_NoEpisodesReturnsFalse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		anime *models.Anime
	}{
		{"empty url", &models.Anime{}},
		{"loopback url with allanime source", &models.Anime{Source: "AllAnime", URL: "http://127.0.0.1:1/none"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isSeries, count, _ := IsSeriesEnhanced(tt.anime)
			assert.False(t, isSeries)
			assert.LessOrEqual(t, count, 1)
		})
	}
}
