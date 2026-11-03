package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pomfy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pomfyProviderWith returns a provider whose adapter is already built.
func pomfyProviderWith(ad scraper.UnifiedScraper) *pomfyProvider {
	p := &pomfyProvider{}
	p.once.Do(func() { p.adapter.UnifiedScraper = ad })
	return p
}

// pomfySearchAdapter serves the catalog list the TMDB filter checks against,
// all routed to one test server; the TMDB answers come from the fake seam.
func pomfySearchAdapter(t *testing.T) (scraper.UnifiedScraper, func(fakes []pomfy.Media)) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/api/catalogo-ids"):
			_, _ = w.Write([]byte(`{"filmes":[550],"series":[1396]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: routeAllTo{target}}

	client := pomfy.NewClientForTest(hc, "https://pomfy.test")
	setFakes := func(fakes []pomfy.Media) {
		client.SetSearchSeam(func(query string) ([]pomfy.Media, error) {
			_ = query
			return fakes, nil
		})
	}
	return scraper.NewPomfyAdapterWithClient(client), setFakes
}

func TestPomfyProvider_SearchTagsResults(t *testing.T) {
	t.Parallel()
	ad, setFakes := pomfySearchAdapter(t)
	setFakes([]pomfy.Media{
		{Title: "Clube da Luta", TMDBID: 550, Kind: pomfy.KindMovie, Year: "1999"},
		{Title: "Breaking Bad", TMDBID: 1396, Kind: pomfy.KindSeries, Year: "2008"},
	})
	p := pomfyProviderWith(ad)

	results, err := p.Search(context.Background(), "clube")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "[Movie] [PT-BR] Clube da Luta", results[0].Name)
	assert.Equal(t, "[TV] [PT-BR] Breaking Bad", results[1].Name)
	for _, r := range results {
		assert.Equal(t, "Pomfy", r.Source)
		kind, _ := source.Resolve(r)
		require.NotNil(t, kind)
		assert.Equal(t, source.Pomfy, kind.Describe().Kind, "a result routes back to Pomfy")
	}
}

func TestPomfyProvider_SearchCancelled(t *testing.T) {
	t.Parallel()
	ad, setFakes := pomfySearchAdapter(t)
	setFakes([]pomfy.Media{{Title: "x", TMDBID: 550, Kind: pomfy.KindMovie}})
	p := pomfyProviderWith(ad)
	_, err := p.Search(cancelledCtx(), "x")
	require.ErrorIs(t, err, context.Canceled)
}

// TestPomfyProvider_FetchEpisodesMovie pins the one-episode movie listing and
// the stream path driving the watch-sources endpoint.
func TestPomfyProvider_FetchEpisodesMovie(t *testing.T) {
	t.Parallel()

	// Swap the api layer's pomfy client seam to a full fake site.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/api/catalogo-ids"):
			_, _ = w.Write([]byte(`{"filmes":[550],"series":[]}`))
		case strings.HasSuffix(r.URL.Path, "/api/watch-sources"):
			_, _ = w.Write([]byte(`{"mode":"direct","sources":[{"kind":"byse","label":"Servidor 1","embedUrl":"https://pomfy-cdn.test/e/gkcq8gbiwh4r","token":"x"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: routeAllTo{target}}

	orig := api.PomfyClientFn
	api.PomfyClientFn = func() *pomfy.Client { return pomfy.NewClientForTest(hc, "https://pomfy.test") }
	t.Cleanup(func() { api.PomfyClientFn = orig })

	p := pomfyProviderWith(scraper.NewPomfyAdapterWithClient(pomfy.NewClientForTest(hc, "https://pomfy.test")))
	media := &models.Anime{Name: "Clube da Luta", URL: "https://pomfy.test/assistir/550", Source: "Pomfy", MediaType: models.MediaTypeMovie, TMDBID: 550}

	eps, err := p.FetchEpisodes(context.Background(), media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "1", eps[0].Number)
	assert.Equal(t, 550, media.TMDBID) // untouched: the listing reads only
}
