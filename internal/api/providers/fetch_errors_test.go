package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyManager creates a ScraperManager with no scrapers registered so every
// GetScraper call returns an error — exercises the error branches of FetchEpisodes
// and FetchStreamURL without any network activity.
func emptyManager() *scraper.ScraperManager {
	return scraper.NewScraperManagerForTest()
}

// ---------------------------------------------------------------------------
// allAnimeProvider.FetchEpisodes / FetchStreamURL
// ---------------------------------------------------------------------------

func TestAllAnimeProvider_FetchEpisodes_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &allAnimeProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://allanime.to/anime/abc123", Source: "AllAnime"}
	_, err := p.FetchEpisodes(context.Background(), anime)
	require.Error(t, err)
}

func TestAllAnimeProvider_FetchStreamURL_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &allAnimeProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://allanime.to/anime/abc123", Source: "AllAnime"}
	ep := &models.Episode{Number: "1", Num: 1, URL: "ep-url"}
	_, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// animeFireProvider.FetchEpisodes / FetchStreamURL
// ---------------------------------------------------------------------------

func TestAnimeFireProvider_FetchEpisodes_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &animeFireProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://animefire.io/anime/naruto", Source: "AnimeFire"}
	_, err := p.FetchEpisodes(context.Background(), anime)
	require.Error(t, err)
}

func TestAnimeFireProvider_FetchStreamURL_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &animeFireProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://animefire.io/anime/naruto", Source: "AnimeFire"}
	ep := &models.Episode{Number: "1", Num: 1, URL: "https://animefire.io/ep/1"}
	_, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// goyabuProvider.FetchEpisodes / FetchStreamURL
// ---------------------------------------------------------------------------

func TestGoyabuProvider_FetchEpisodes_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &goyabuProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://goyabu.io/anime/naruto", Source: "Goyabu"}
	_, err := p.FetchEpisodes(context.Background(), anime)
	require.Error(t, err)
}

func TestGoyabuProvider_FetchStreamURL_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &goyabuProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://goyabu.io/anime/naruto", Source: "Goyabu"}
	ep := &models.Episode{Number: "1", Num: 1, URL: "https://goyabu.io/ep/1"}
	_, err := p.FetchStreamURL(context.Background(), ep, anime, "")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// ForKind integration (verifies factory + error path for unknown kind)
// ---------------------------------------------------------------------------

func TestForKind_UnknownKind_ReturnsError(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	_, err := ForKind(source.SourceKind("__nonexistent_source__"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no provider registered")
}
