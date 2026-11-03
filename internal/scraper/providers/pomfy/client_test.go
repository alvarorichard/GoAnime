package pomfy

import (
	"testing"
)

// TestWatchSourcesSeriesQuery pins the two series-specific query parameters
// the watch-sources call carries (temporada/episodio), recorded 2026-11-02.
func TestWatchSourcesSeriesQuery(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)

	if _, err := c.WatchSources(t.Context(), 1396, KindSeries, 2, 1); err != nil {
		t.Fatalf("WatchSources series: %v", err)
	}
	if site.moviesWCalls != 1 {
		t.Errorf("watch-sources calls = %d, want 1", site.moviesWCalls)
	}
}

// TestByseCodePaths pins the /e/<code> code extraction the embed URLs use.
func TestByseCodePaths(t *testing.T) {
	cases := []struct{ raw, want string }{
		{"https://pomfy-cdn.shop/e/gkcq8gbiwh4r", "gkcq8gbiwh4r"},
		{"https://pomfy-cdn.shop/e/u5zkyenlctsj/heart-of-the-beast", "u5zkyenlctsj"},
		{"", ""},
		{"https://pomfy-cdn.shop/e/", ""},
		{"https://cdn.test/not/a/code", ""},
		{"https://cdn.test/e/too-short", ""},
	}
	for _, tc := range cases {
		if got := ByseCode(tc.raw); got != tc.want {
			t.Errorf("ByseCode(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// TestSeamSearchFilter pins the catalog filter with the search seam swapped:
// TMDB hits not on the catalog drop out, kept ones get their URL, and the
// models carry ids/types.
func TestSeamSearchFilter(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	c.SetSearchSeam(func(query string) ([]Media, error) {
		return []Media{
			{Title: "Clube da Luta", TMDBID: 550, TMDBType: "movie"},
			{Title: "Forrest Gump", TMDBID: 13, TMDBType: "movie"},
			{Title: "Breaking Bad", TMDBID: 1396, TMDBType: "tv"},
		}, nil
	})
	results, err := c.Search(t.Context(), "clube da luta")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 || results[0].TMDBID != 550 || results[1].TMDBID != 1396 {
		t.Fatalf("results = %+v, want the two catalog titles", results)
	}
	if results[0].URL != c.BaseURL()+"/assistir/550" {
		t.Errorf("movie URL = %q", results[0].URL)
	}
	if results[1].URL != c.BaseURL()+"/serie/1396" {
		t.Errorf("series URL = %q", results[1].URL)
	}
	for _, r := range results {
		anime := r.ToAnimeModel()
		if anime.TMDBID == 0 || (anime.MediaType != "movie" && anime.MediaType != "tv") {
			t.Errorf("ToAnimeModel lost ids/types: %+v", anime)
		}
	}
}
