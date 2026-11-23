package pobreflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// On 2026-11-23 the site was rebuilt on WordPress. Its /sitemap.xml index
// stopped listing sitemap-movies.xml / sitemap-series.xml and lists one
// sitemap per post type and year instead; the client found "no movie/series
// files" and the corpus fallback was dead.

const wordpressSitemapIndex = `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<sitemap><loc>{{BASE}}/sitemap-posttype-filme.2026.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-posttype-serie.2026.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-posttype-filme.2025.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-posttype-page.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-taxonomy-genero_filme.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-taxonomy-ano_serie.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-author.xml</loc></sitemap>
<sitemap><loc>{{BASE}}/sitemap-news.xml</loc></sitemap>
</sitemapindex>`

const wordpressMovieShard = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url><loc>https://pobreflixtvfree.cc/filmes/online/resident-evil-dublado-1423191</loc></url>
<url><loc>https://pobreflixtvfree.cc/filmes/online/nu-dublado-433251</loc></url>
</urlset>`

const wordpressSeriesShard = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url><loc>https://pobreflixtvfree.cc/series/online/typhoon-family-dublado-280946</loc></url>
</urlset>`

func wordpressSite(t *testing.T) *Client {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sitemap.xml":
			// The index names its shards by absolute URL: point them here, so
			// the test never reaches the real site.
			_, _ = w.Write([]byte(strings.ReplaceAll(wordpressSitemapIndex, "{{BASE}}", srv.URL)))
		case "/sitemap-posttype-filme.2026.xml":
			_, _ = w.Write([]byte(wordpressMovieShard))
		case "/sitemap-posttype-serie.2026.xml":
			_, _ = w.Write([]byte(wordpressSeriesShard))
		case "/sitemap-posttype-filme.2025.xml":
			_, _ = w.Write([]byte(`<urlset></urlset>`))
		default:
			http.NotFound(w, r) // including /sitemap.php, the old index
		}
	}))
	t.Cleanup(srv.Close)
	return NewClientForTest(http.DefaultClient, srv.URL, t.TempDir()+"/corpus.json")
}

func TestSitemapShardsReadsThePostTypeLayout(t *testing.T) {
	t.Parallel()
	c := wordpressSite(t)
	shards, err := c.sitemapShards(context.Background())
	require.NoError(t, err, "the index lists content sitemaps; this must not be 'no movie/series files'")
	var names []string
	for _, s := range shards {
		names = append(names, s.name)
	}
	assert.Equal(t, []string{"filme-2025", "filme-2026", "serie-2026"}, names,
		"movies before series, older years first; pages, taxonomies, authors and news are not content")
}

func TestSitemapSearchFindsTitlesInThePostTypeLayout(t *testing.T) {
	t.Parallel()
	c := wordpressSite(t)
	movies, err := c.Search(context.Background(), "resident evil")
	require.NoError(t, err)
	require.Len(t, movies, 1)
	assert.Equal(t, "Resident Evil", movies[0].Title)
	assert.Equal(t, KindMovie, movies[0].Kind)
	assert.Equal(t, 1423191, movies[0].VideoID)
	assert.Equal(t, AudioDubbed, movies[0].Audio)

	series, err := c.Search(context.Background(), "typhoon family")
	require.NoError(t, err)
	require.Len(t, series, 1)
	assert.Equal(t, KindSeries, series[0].Kind)
}

func TestCurrentSitemapReMatches(t *testing.T) {
	t.Parallel()
	for path, want := range map[string]bool{
		"/sitemap-movies.xml":                     true,
		"/sitemap-series-2.xml":                   true,
		"/sitemap-filmes.xml":                     true,
		"/sitemap-posttype-filme.2026.xml":        true,
		"/sitemap-posttype-serie.2026.xml":        true,
		"/sitemap-posttype-serie.xml":             true,
		"/sitemap-posttype-page.xml":              false,
		"/sitemap-taxonomy-genero_filme.xml":      false,
		"/sitemap-taxonomy-ano_serie.xml":         false,
		"/sitemap-author.xml":                     false,
		"/sitemap-news.xml":                       false,
		"/sitemap-posttype-filme.2026.xml.backup": false,
	} {
		assert.Equal(t, want, currentSitemapRe.MatchString(path), path)
	}
}
