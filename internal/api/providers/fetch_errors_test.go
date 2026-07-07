package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
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

// restoreStreamFns resets the stream-fetch indirections after a test stubbed
// them. Tests that stub these globals must NOT run in parallel.
func restoreStreamFns(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		allAnimeEnhancedStreamFn = api.GetEpisodeStreamURLEnhanced
		fallbackStreamFn = api.GetEpisodeStreamURL
		superFlixStreamFn = api.GetSuperFlixStreamURL
	})
}

func TestAllAnimeProvider_FetchStreamURL(t *testing.T) {
	// Stubs package-level fn indirections — not parallel.
	p := &allAnimeProvider{}
	anime := &models.Anime{URL: "hHjXnUTda", Source: "AllAnime"}
	ep := &models.Episode{Number: "1", Num: 1, URL: "hHjXnUTda"}

	t.Run("enhanced path wins", func(t *testing.T) {
		restoreStreamFns(t)
		allAnimeEnhancedStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			return "https://cdn.example/enhanced.m3u8", nil
		}
		fallbackStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			t.Error("fallback must not run when enhanced succeeds")
			return "", nil
		}
		url, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example/enhanced.m3u8", url)
	})

	t.Run("falls back to regular enhanced API", func(t *testing.T) {
		restoreStreamFns(t)
		allAnimeEnhancedStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			return "", assert.AnError
		}
		fallbackStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			return "https://cdn.example/regular.m3u8", nil
		}
		url, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example/regular.m3u8", url)
	})

	t.Run("both paths failing surfaces error", func(t *testing.T) {
		restoreStreamFns(t)
		allAnimeEnhancedStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			return "", assert.AnError
		}
		fallbackStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			return "", assert.AnError
		}
		_, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "allAnime stream")
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		restoreStreamFns(t)
		allAnimeEnhancedStreamFn = func(_ *models.Episode, _ *models.Anime, _ string) (string, error) {
			t.Error("must not fetch with cancelled context")
			return "", nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := p.FetchStreamURL(ctx, ep, anime, "best")
		require.ErrorIs(t, err, context.Canceled)
	})
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
	// Mutates util globals (ClearGlobalSubtitles/SetGlobalAnimeSource) — not parallel.
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
	// Mutates util globals (ClearGlobalSubtitles/SetGlobalAnimeSource) — not parallel.
	t.Cleanup(ResetForTesting)
	p := &goyabuProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://goyabu.io/anime/naruto", Source: "Goyabu"}
	ep := &models.Episode{Number: "1", Num: 1, URL: "https://goyabu.io/ep/1"}
	_, err := p.FetchStreamURL(context.Background(), ep, anime, "")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// superFlixProvider.FetchEpisodes / FetchStreamURL
// ---------------------------------------------------------------------------

func TestSuperFlixProvider_FetchEpisodes_ScraperNotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(ResetForTesting)
	p := &superFlixProvider{sm: emptyManager()}
	anime := &models.Anime{URL: "https://superflix.gs/serie/1234", Source: "SuperFlix"}
	_, err := p.FetchEpisodes(context.Background(), anime)
	require.Error(t, err)
}

func TestSuperFlixProvider_FetchStreamURL(t *testing.T) {
	// Stubs package-level fn indirections and reads global anime source — not parallel.
	p := &superFlixProvider{}
	anime := &models.Anime{URL: "1234", Source: "SuperFlix", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "3", Num: 3, URL: "1234", SeasonID: "2"}

	t.Run("delegates to the full SuperFlix UX path", func(t *testing.T) {
		restoreStreamFns(t)
		var gotAnime *models.Anime
		var gotEp *models.Episode
		var gotQuality string
		superFlixStreamFn = func(a *models.Anime, e *models.Episode, q string) (string, error) {
			gotAnime, gotEp, gotQuality = a, e, q
			return "https://cdn.example/sf.m3u8", nil
		}
		url, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example/sf.m3u8", url)
		assert.Same(t, anime, gotAnime)
		assert.Same(t, ep, gotEp)
		assert.Equal(t, "best", gotQuality)
		assert.Equal(t, "SuperFlix", util.GetGlobalAnimeSource(), "entry side effect must tag the global source")
	})

	t.Run("error passthrough", func(t *testing.T) {
		restoreStreamFns(t)
		superFlixStreamFn = func(_ *models.Anime, _ *models.Episode, _ string) (string, error) {
			return "", assert.AnError
		}
		_, err := p.FetchStreamURL(context.Background(), ep, anime, "")
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		restoreStreamFns(t)
		superFlixStreamFn = func(_ *models.Anime, _ *models.Episode, _ string) (string, error) {
			t.Error("must not fetch with cancelled context")
			return "", nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := p.FetchStreamURL(ctx, ep, anime, "best")
		require.ErrorIs(t, err, context.Canceled)
	})
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
