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
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// routeAllTo sends every request to one test server whatever its host.
type routeAllTo struct{ target *url.URL }

func (r routeAllTo) RoundTrip(req *http.Request) (*http.Response, error) {
	out := req.Clone(req.Context())
	out.URL.Scheme, out.URL.Host, out.Host = r.target.Scheme, r.target.Host, r.target.Host
	return http.DefaultTransport.RoundTrip(out)
}

// topCineProviderWith returns a provider whose adapter is already built.
func topCineProviderWith(ad scraper.UnifiedScraper) *topCineProvider {
	p := &topCineProvider{}
	p.once.Do(func() { p.adapter.UnifiedScraper = ad })
	return p
}

func topCineSearchAdapter(t *testing.T, h http.HandlerFunc) scraper.UnifiedScraper {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: routeAllTo{target}}
	return scraper.NewTopCineAdapterWithClients(
		topcine.NewClientForTest(hc, "https://topcine.test"),
		startflix.NewClientForTest(hc, "https://www.startflix.test"),
	)
}

func TestTopCineProvider_SearchTagsResults(t *testing.T) {
	t.Parallel()
	p := topCineProviderWith(topCineSearchAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "dexter", r.URL.Query().Get("q"))
		_, _ = w.Write([]byte(`
			<article class="pagina-busca-card"><a href="/serie/dexter"><h2>Dexter</h2><small>2006 • Série</small></a></article>
			<article class="pagina-busca-card"><a href="/filme/dexter-o-filme"><h2>Dexter</h2><small>2020 • Filme</small></a></article>`))
	}))

	results, err := p.Search(context.Background(), "dexter")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "[TV] [PT-BR] Dexter", results[0].Name)
	assert.Equal(t, "[Movie] [PT-BR] Dexter", results[1].Name)
	for _, r := range results {
		assert.Equal(t, "TopCine", r.Source)
		kind, _ := source.Resolve(r)
		require.NotNil(t, kind)
		assert.Equal(t, source.TopCine, kind.Describe().Kind, "a result routes back to TopCine")
	}
}

func TestTopCineProvider_SearchError(t *testing.T) {
	t.Parallel()
	p := topCineProviderWith(topCineSearchAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	_, err := p.Search(context.Background(), "x")
	require.Error(t, err)
}

func TestTopCineProvider_SearchCancelled(t *testing.T) {
	t.Parallel()
	p := topCineProviderWith(topCineSearchAdapter(t, func(http.ResponseWriter, *http.Request) {
		t.Error("searched with a cancelled context")
	}))
	_, err := p.Search(cancelledCtx(), "x")
	require.ErrorIs(t, err, context.Canceled)
}

// plainScraper is an adapter that cannot take a context.
type plainScraper struct {
	results []*models.Anime
	err     error
}

func (s plainScraper) SearchAnime(string, ...any) ([]*models.Anime, error) { return s.results, s.err }
func (plainScraper) GetAnimeEpisodes(string) ([]models.Episode, error)     { return nil, nil }
func (plainScraper) GetStreamURL(string, ...any) (streamURL string, metadata map[string]string, err error) {
	return "", nil, nil
}
func (plainScraper) GetType() scraper.ScraperType { return scraper.TopCineType }

func TestTopCineProvider_SearchWithoutContextSupport(t *testing.T) {
	t.Parallel()
	p := topCineProviderWith(plainScraper{results: []*models.Anime{{Name: "X", MediaType: models.MediaTypeMovie}}})
	results, err := p.Search(context.Background(), "x")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "[Movie] [PT-BR] X", results[0].Name)

	_, err = topCineProviderWith(plainScraper{err: assert.AnError}).Search(context.Background(), "x")
	require.ErrorIs(t, err, assert.AnError)
}

func TestTopCineProvider_SearchWithoutAdapter(t *testing.T) {
	t.Parallel()
	p := &topCineProvider{}
	p.once.Do(func() {}) // the build ran and produced nothing
	_, err := p.Search(context.Background(), "x")
	require.ErrorContains(t, err, "no adapter")
}

func TestTopCineProvider_FetchStreamURL(t *testing.T) {
	// Stubs package-level fn indirections and reads global state — not parallel.
	p := &topCineProvider{}
	anime := &models.Anime{URL: "https://topcine3.site/serie/x", Source: "TopCine", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "1", Num: 1, URL: "https://www.painel-aso.sbs/episodio/1"}
	t.Cleanup(func() { topCineStreamFn = api.GetTopCineStreamURL })

	t.Run("delegates to the api path and resets per-episode state", func(t *testing.T) {
		util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://old.test/prev.vtt"}})
		var gotAnime *models.Anime
		var gotEp *models.Episode
		var gotQuality string
		var subsAtCall int
		topCineStreamFn = func(a *models.Anime, e *models.Episode, q string) (string, error) {
			gotAnime, gotEp, gotQuality = a, e, q
			subsAtCall = len(util.GetGlobalSubtitles())
			return "https://cdn.example/tc.m3u8", nil
		}
		got, err := p.FetchStreamURL(context.Background(), ep, anime, "720")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example/tc.m3u8", got)
		assert.Same(t, anime, gotAnime)
		assert.Same(t, ep, gotEp)
		assert.Equal(t, "720", gotQuality)
		assert.Zero(t, subsAtCall, "the previous episode's subtitles must not leak in")
		assert.Equal(t, "TopCine", util.GetGlobalAnimeSource())
	})

	t.Run("error passthrough", func(t *testing.T) {
		topCineStreamFn = func(*models.Anime, *models.Episode, string) (string, error) { return "", assert.AnError }
		_, err := p.FetchStreamURL(context.Background(), ep, anime, "")
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("an anime without a source leaves the global alone", func(t *testing.T) {
		util.SetGlobalAnimeSource("Previous")
		topCineStreamFn = func(*models.Anime, *models.Episode, string) (string, error) { return "u", nil }
		_, err := p.FetchStreamURL(context.Background(), ep, &models.Anime{}, "")
		require.NoError(t, err)
		assert.Equal(t, "Previous", util.GetGlobalAnimeSource())
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		topCineStreamFn = func(*models.Anime, *models.Episode, string) (string, error) {
			t.Error("must not fetch with a cancelled context")
			return "", nil
		}
		_, err := p.FetchStreamURL(cancelledCtx(), ep, anime, "best")
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestTopCineProvider_FetchEpisodes(t *testing.T) {
	// Stubs a package-level fn indirection — not parallel.
	p := &topCineProvider{}
	anime := &models.Anime{URL: "https://topcine3.site/filme/x", Source: "TopCine", MediaType: models.MediaTypeMovie}
	t.Cleanup(func() { topCineEpisodesFn = api.GetTopCineEpisodes })

	t.Run("delegates to the interactive api listing", func(t *testing.T) {
		var got *models.Anime
		topCineEpisodesFn = func(a *models.Anime) ([]models.Episode, error) {
			got = a
			return []models.Episode{{Number: "1", Num: 1}}, nil
		}
		eps, err := p.FetchEpisodes(context.Background(), anime)
		require.NoError(t, err)
		assert.Len(t, eps, 1)
		assert.Same(t, anime, got)
	})

	t.Run("error passthrough", func(t *testing.T) {
		topCineEpisodesFn = func(*models.Anime) ([]models.Episode, error) { return nil, assert.AnError }
		_, err := p.FetchEpisodes(context.Background(), anime)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		topCineEpisodesFn = func(*models.Anime) ([]models.Episode, error) {
			t.Error("must not list episodes with a cancelled context")
			return nil, nil
		}
		_, err := p.FetchEpisodes(cancelledCtx(), anime)
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestTopCineProvider_DefaultSeams(t *testing.T) {
	// Production must point at the TopCine api paths, not StartFlix's.
	assert.NotNil(t, topCineStreamFn)
	assert.NotNil(t, topCineEpisodesFn)
	_, err := topCineEpisodesFn(nil)
	require.ErrorContains(t, err, "no TopCine page")
}
