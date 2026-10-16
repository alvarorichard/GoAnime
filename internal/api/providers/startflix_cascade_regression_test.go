package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. The StartFlix provider owns search while
// episodes/stream delegate to the api UX path, so this cascade drives a real
// adapter search against a stub, then the delegation contract: tagging,
// routing back, globals, and error passthrough.

// startFlixSearchAdapter points a real StartFlix adapter at a stub catalog.
func startFlixSearchAdapter(t *testing.T, h http.HandlerFunc) scraper.UnifiedScraper {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: routeAllTo{target}}
	return scraper.NewStartFlixAdapterWithClient(
		startflix.NewClientForTest(hc, "https://www.startflix.test"),
	)
}

func startFlixProviderWith(ad scraper.UnifiedScraper) *startFlixProvider {
	p := &startFlixProvider{}
	p.once.Do(func() { p.adapter.UnifiedScraper = ad })
	return p
}

// TestStartFlixCascade_SearchToFetch wires search → tagging → routing →
// FetchEpisodes/FetchStreamURL delegation in one flow.
func TestStartFlixCascade_SearchToFetch(t *testing.T) {
	// Stubs the Fetch fn indirections and reads globals — not parallel.
	p := startFlixProviderWith(startFlixSearchAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "dexter", r.URL.Query().Get("s"))
		_, _ = w.Write([]byte(`
			<div class="result-item"><article>
				<div class="image"><img src="https://img.test/t/p/w92/d.jpg" /></div>
				<div class="details"><div class="title"><a href="https://www.startflix.test/series/dexter/">Dexter</a></div>
				<div class="meta"><span class="year">2006</span></div></div></article></div>`))
	}))
	ctx := context.Background()

	results, err := p.Search(ctx, "dexter")
	require.NoError(t, err)
	require.Len(t, results, 1)
	anime := results[0]
	assert.Equal(t, "[TV] [PT-BR] Dexter", anime.Name)
	assert.Equal(t, "StartFlix", anime.Source)

	_, resolved := source.Resolve(anime)
	assert.Equal(t, source.StartFlix, resolved.Kind, "reason: %s", resolved.Reason)

	t.Cleanup(func() {
		startFlixEpisodesFn = api.GetStartFlixEpisodes
		restoreStreamFns(t)
	})
	ep := &models.Episode{Number: "1", Num: 1, URL: "https://painel.test/episodio/11"}
	startFlixEpisodesFn = func(a *models.Anime) ([]models.Episode, error) {
		require.Same(t, anime, a, "the tagged result must reach the listing untouched")
		return []models.Episode{*ep}, nil
	}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	startFlixStreamFn = func(a *models.Anime, e *models.Episode, q string) (string, error) {
		require.Same(t, anime, a)
		require.Same(t, &eps[0], e)
		assert.Equal(t, "best", q)
		return "https://cdn.example/sfx.m3u8", nil
	}
	got, err := p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/sfx.m3u8", got)
	assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource())

	startFlixStreamFn = func(*models.Anime, *models.Episode, string) (string, error) {
		return "", assert.AnError
	}
	_, err = p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.ErrorIs(t, err, assert.AnError, "a dead episode must fail, not return empty")
}

// TestStartFlixCascade_SearchFailurePropagation pins a catalog outage as an
// error, not an empty success.
func TestStartFlixCascade_SearchFailurePropagation(t *testing.T) {
	t.Parallel()

	p := startFlixProviderWith(startFlixSearchAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	_, err := p.Search(context.Background(), "dexter")
	require.Error(t, err)

	_, err = p.Search(cancelledCtx(), "dexter")
	require.ErrorIs(t, err, context.Canceled)
}
