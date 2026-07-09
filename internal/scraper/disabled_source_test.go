package scraper

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchAllScrapers_SkipsDisabledSource pins the S1 kill-switch on the
// concurrent search path: a source named in GOANIME_DISABLED_SOURCES is never
// queried, while the others still run.
func TestSearchAllScrapers_SkipsDisabledSource(t *testing.T) {
	// Uses t.Setenv — not parallel.
	allAnime := &MockScraper{searchFunc: func(string) ([]*models.Anime, error) {
		return []*models.Anime{{Name: "AA", URL: "id1"}}, nil
	}}
	animeFire := &MockScraper{searchFunc: func(string) ([]*models.Anime, error) {
		return []*models.Anime{{Name: "AF", URL: "https://animefire.io/1"}}, nil
	}}
	mgr := createTestManager(allAnime, animeFire)

	t.Setenv("GOANIME_DISABLED_SOURCES", "AllAnime")

	results, err := mgr.SearchAnime("x", nil)
	require.NoError(t, err)

	assert.Equal(t, int32(0), allAnime.searchCallCount.Load(), "disabled source must not be queried")
	assert.Equal(t, int32(1), animeFire.searchCallCount.Load(), "enabled source must still be queried")
	for _, r := range results {
		assert.NotEqual(t, "AA", r.Name, "disabled source must contribute no results")
	}
}

// TestSearchSpecificScraper_DisabledReturnsError pins that explicitly targeting
// a disabled source fails with a clear reason instead of querying it.
func TestSearchSpecificScraper_DisabledReturnsError(t *testing.T) {
	// Uses t.Setenv — not parallel.
	allAnime := &MockScraper{searchFunc: func(string) ([]*models.Anime, error) {
		return []*models.Anime{{Name: "AA"}}, nil
	}}
	mgr := createTestManager(allAnime, nil)

	t.Setenv("GOANIME_DISABLED_SOURCES", "AllAnime")

	st := AllAnimeType
	_, err := mgr.SearchAnime("x", &st)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
	assert.Equal(t, int32(0), allAnime.searchCallCount.Load(), "disabled source must not be queried")
}
