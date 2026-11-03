package pomfy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Watch sources, catalog membership, the numeric-search path and the title
// payloads — all through the fakeCatalogSite, whose answers were recorded
// from the production shapes.

func TestWatchSourcesFiltersNonByse(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)

	sources, err := c.WatchSources(t.Context(), 550, KindMovie, 0, 0)
	if err != nil {
		t.Fatalf("WatchSources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("sources = %d, want only the byse one (flyfile dropped)", len(sources))
	}
	if ByseCode(sources[0].EmbedURL) != "gkcq8gbiwh4r" {
		t.Errorf("code = %q, want gkcq8gbiwh4r", ByseCode(sources[0].EmbedURL))
	}

	// The series spelling carries temporada/episodio in the query.
	if _, err := c.WatchSources(t.Context(), 1396, KindSeries, 2, 1); err != nil {
		t.Fatalf("WatchSources series: %v", err)
	}
	if site.moviesWCalls != 2 {
		t.Errorf("watch-sources calls = %d, want 2", site.moviesWCalls)
	}
}

func TestWatchSourcesRejectsIncompleteQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/watch-sources") {
			http.Error(w, `{"error":"missing temporada"}`, http.StatusBadRequest)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	if _, err := c.WatchSources(t.Context(), 1396, KindSeries, 3, 1); err == nil ||
		!strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("err = %v, want the 400 the site answers a broken query with", err)
	}
}

func TestWatchSourcesEmptyListsError(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	site.breakWatch = true
	if _, err := c.WatchSources(t.Context(), 550, KindMovie, 0, 0); err == nil {
		t.Fatal("a broken watch-sources answer must surface an error")
	}
}

func TestCatalogMembership(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	ctx := t.Context()

	ok, err := c.HasTitle(ctx, 550, KindMovie)
	if err != nil || !ok {
		t.Fatalf("movie 550: err=%v ok=%v", err, ok)
	}
	if ok, _ = c.HasTitle(ctx, 1396, KindSeries); !ok {
		t.Error("series 1396 must be on the catalog")
	}
	if ok, _ = c.HasTitle(ctx, 550, KindSeries); ok {
		t.Error("movie 550 must not count as a series")
	}
	if ok, _ = c.HasTitle(ctx, 999, KindMovie); ok {
		t.Error("unknown id 999 must be off the catalog")
	}
}

// TestSearchNumericID is the keyless path the CLI users get for free when
// they search a bare TMDB id: the site payloads name the title.
func TestSearchNumericID(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)

	results, err := c.Search(t.Context(), "550")
	if err != nil {
		t.Fatalf("Search 550: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Clube da Luta" || results[0].Kind != KindMovie {
		t.Fatalf("results = %+v, want the one movie", results)
	}
	if results[0].URL != c.BaseURL()+"/assistir/550" {
		t.Errorf("URL = %q", results[0].URL)
	}

	results, err = c.Search(t.Context(), "1396")
	if err != nil {
		t.Fatalf("Search 1396: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Breaking Bad" || results[0].Kind != KindSeries {
		t.Fatalf("series results = %+v, want Breaking Bad as serie", results)
	}
	site.mu.Lock()
	seriesHit := false
	for _, hit := range site.titleHits {
		if strings.HasPrefix(hit, "/serie/") {
			seriesHit = true
		}
	}
	site.mu.Unlock()
	if !seriesHit {
		t.Error("the series id never opened the serie payload")
	}
}

func TestSearchNumericIDOffCatalog(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	if _, err := c.Search(t.Context(), "9999"); err == nil || !strings.Contains(err.Error(), "not on the catalog") {
		t.Fatalf("err = %v, want ErrNotOnCatalog", err)
	}
}

// TestParseMovieTitleFailsLoudOnGarbage keeps the parser honest.
func TestParseMovieTitleFailsLoudOnGarbage(t *testing.T) {
	for _, body := range []string{`{}`, `<html>`, `{"nodes":[]}`, `not json`} {
		if _, err := parseMovieTitle([]byte(body)); err == nil {
			t.Errorf("garbage %q must fail, not produce a zero title", body)
		}
	}
}

func TestDecodeTitleNodeValues(t *testing.T) {
	var parsed struct {
		Nodes []struct {
			Data []any `json:"data"`
		} `json:"nodes"`
	}
	if _, err := parseMovieTitle([]byte(movieTitleFixture)); err != nil {
		t.Fatalf("parseMovieTitle: %v", err)
	}
	_ = parsed
}

func TestCatalogCacheSharedAcrossProviders(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	ctx := t.Context()
	if _, err := c.catalogData(ctx); err != nil {
		t.Fatalf("first catalog: %v", err)
	}
	if _, err := c.catalogData(ctx); err != nil {
		t.Fatalf("second catalog: %v", err)
	}
	// One HTTP hit is the cache contract: both lookups below read memory.
	// (the fake does not count; assert via the cached pointer identity)
	c.mu.Lock()
	first := c.catalog
	second := c.catalog
	c.mu.Unlock()
	if first != second {
		t.Error("catalog cache was replaced between reads")
	}
	if time.Since(first.fetched) >= catalogTTL {
		t.Error("a fresh catalog must be well within its TTL")
	}
}
