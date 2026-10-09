package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: the old *_ScraperNotFound tests injected an empty ScraperManager via the
// (now-deleted) legacy Provider factory to exercise a "scraper not found" error.
// Providers now own their adapter directly (scraper.NewAdapter), so that error
// path no longer exists. The remaining unit-testable contract is entry-time
// context cancellation and StartFlix's delegation to the api UX path.

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestProviders_FetchStreamURL_CancelledContext(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{URL: "x", Source: "X"}
	ep := &models.Episode{Number: "1", URL: "x"}

	for _, p := range []interface {
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	}{
		&hianimeProvider{},
		&animeFireProvider{},
		&goyabuProvider{},
		&startFlixProvider{},
		&topCineProvider{},
	} {
		_, err := p.FetchStreamURL(cancelledCtx(), ep, anime, "best")
		require.ErrorIs(t, err, context.Canceled)
	}
}

// restoreStreamFns resets the StartFlix stream indirection after a test stubbed
// it. Tests that stub this global must NOT run in parallel.
func restoreStreamFns(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		startFlixStreamFn = api.GetStartFlixStreamURL
	})
}

func TestStartFlixProvider_FetchStreamURL(t *testing.T) {
	// Stubs package-level fn indirections and reads global anime source — not parallel.
	p := &startFlixProvider{}
	anime := &models.Anime{URL: "https://www.startflix.biz/series/x/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "3", Num: 3, URL: "https://www.painel-aso.sbs/episodio/1", SeasonID: "2"}

	t.Run("delegates to the full StartFlix UX path", func(t *testing.T) {
		restoreStreamFns(t)
		var gotAnime *models.Anime
		var gotEp *models.Episode
		var gotQuality string
		startFlixStreamFn = func(a *models.Anime, e *models.Episode, q string) (string, error) {
			gotAnime, gotEp, gotQuality = a, e, q
			return "https://cdn.example/sfx.m3u8", nil
		}
		url, err := p.FetchStreamURL(context.Background(), ep, anime, "best")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example/sfx.m3u8", url)
		assert.Same(t, anime, gotAnime)
		assert.Same(t, ep, gotEp)
		assert.Equal(t, "best", gotQuality)
		assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource(), "entry side effect must tag the global source")
	})

	t.Run("error passthrough", func(t *testing.T) {
		restoreStreamFns(t)
		startFlixStreamFn = func(_ *models.Anime, _ *models.Episode, _ string) (string, error) {
			return "", assert.AnError
		}
		_, err := p.FetchStreamURL(context.Background(), ep, anime, "")
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		restoreStreamFns(t)
		startFlixStreamFn = func(_ *models.Anime, _ *models.Episode, _ string) (string, error) {
			t.Error("must not fetch with cancelled context")
			return "", nil
		}
		_, err := p.FetchStreamURL(cancelledCtx(), ep, anime, "best")
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestStartFlixProvider_FetchEpisodes(t *testing.T) {
	// Stubs a package-level fn indirection — not parallel.
	p := &startFlixProvider{}
	anime := &models.Anime{URL: "https://www.startflix.biz/filmes/x/", Source: "StartFlix", MediaType: models.MediaTypeMovie}
	t.Cleanup(func() { startFlixEpisodesFn = api.GetStartFlixEpisodes })

	t.Run("delegates to the interactive api listing", func(t *testing.T) {
		var got *models.Anime
		startFlixEpisodesFn = func(a *models.Anime) ([]models.Episode, error) {
			got = a
			return []models.Episode{{Number: "1", Num: 1}}, nil
		}
		eps, err := p.FetchEpisodes(context.Background(), anime)
		require.NoError(t, err)
		assert.Len(t, eps, 1)
		assert.Same(t, anime, got)
	})

	t.Run("error passthrough", func(t *testing.T) {
		startFlixEpisodesFn = func(*models.Anime) ([]models.Episode, error) { return nil, assert.AnError }
		_, err := p.FetchEpisodes(context.Background(), anime)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		startFlixEpisodesFn = func(*models.Anime) ([]models.Episode, error) {
			t.Error("must not list episodes with a cancelled context")
			return nil, nil
		}
		_, err := p.FetchEpisodes(cancelledCtx(), anime)
		require.ErrorIs(t, err, context.Canceled)
	})
}
