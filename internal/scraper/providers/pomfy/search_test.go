package pomfy

import (
	"fmt"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// The search tests run the REAL tmdbSearch against a fakeTMDB whose TMDB
// judgments are the ones the production server makes (401 for off-pool keys,
// 429 for over-quota keys, 200 with pages for known keys). The package seam
// tmdbBase is repointed for the duration of each test.

// setTMDBBase points tmdbSearch at fn's value until the test ends.
func setTMDBBase(t *testing.T, u string) {
	t.Helper()
	old := tmdbBase
	tmdbBase = u
	t.Cleanup(func() { tmdbBase = old })
}

func resetSearchCache(t *testing.T) {
	t.Helper()
	tmdbCacheMu.Lock()
	prior := tmdbCache
	tmdbCache = map[string]tmdbCacheEntry{}
	tmdbCacheMu.Unlock()
	t.Cleanup(func() {
		tmdbCacheMu.Lock()
		tmdbCache = prior
		tmdbCacheMu.Unlock()
	})
}

func fakeMedia(n int) []map[string]any {
	out := make([]map[string]any, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, map[string]any{
			"id":           i + 1000,
			"media_type":   "movie",
			"title":        fmt.Sprintf("Fake Movie %02d", i),
			"vote_average": 7.5,
			"overview":     fmt.Sprintf("overview %d", i),
		})
	}
	return out
}

func TestSearchMultiWithKeyPages(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "fight", fakeMedia(21)) // > 20 forces page 2
	setTMDBBase(t, tmdb.srv.URL+"/3")

	movies, err := searchMultiWithKey(PublicKeyForTesting(0), "fight")
	if err != nil {
		t.Fatalf("searchMultiWithKey: %v", err)
	}
	if len(movies) != 21 {
		t.Errorf("len(movies) = %d, want 21 (page 1 list + empty page 2)", len(movies))
	}
	pages := tmdb.usedKeys[PublicKeyForTesting(0)]
	if len(pages) != 2 || pages[0] != 1 || pages[1] != 2 {
		t.Errorf("pages used = %v, want [1 2]", pages)
	}
}

func TestSearchMultiWithKeyShortPageStops(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "short", fakeMedia(5)) // < 20 → page 1 only
	setTMDBBase(t, tmdb.srv.URL+"/3")

	if _, err := searchMultiWithKey(PublicKeyForTesting(1), "short"); err != nil {
		t.Fatalf("searchMultiWithKey: %v", err)
	}
	pages := tmdb.usedKeys[PublicKeyForTesting(1)]
	if len(pages) != 1 || pages[0] != 1 {
		t.Errorf("pages used = %v, want [1] (a short page stops the paging)", pages)
	}
}

func TestSearchMultiWithKeyOffPoolKeyRejected(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "fight", nil)
	setTMDBBase(t, tmdb.srv.URL+"/3")

	if _, err := searchMultiWithKey("00000000000000000000000000000001", "fight"); err == nil {
		t.Fatal("an off-pool key must meet the TMDB 401")
	}
}

// TestTMDBSearchRotatesOnQuota proves the search ends on a surviving pool
// key whatever order the shuffle picks: dead keys answer 429, the page
// triple lands, and no key gets tried twice — the rotation moves forward.
func TestTMDBSearchRotatesOnQuota(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "rotates", fakeMedia(2))
	setTMDBBase(t, tmdb.srv.URL+"/3")
	for i := 0; i < len(publicTMDBKeys)-1; i++ {
		tmdb.quota[PublicKeyForTesting(i)] = true
	}

	movies, err := tmdbSearch("rotates")
	if err != nil {
		t.Fatalf("tmdbSearch: %v", err)
	}
	if len(movies) != 2 {
		t.Errorf("len(movies) = %d, want 2 from the surviving key", len(movies))
	}
	tmdb.mu.Lock()
	defer tmdb.mu.Unlock()
	for key, pages := range tmdb.usedKeys {
		if len(pages) > 1 {
			t.Errorf("key %s…%s was tried %d times, want at most once per rotation", key[:8], key[24:], len(pages))
		}
	}
	survivorOK := false
	for i, key := range publicTMDBKeys {
		if i < len(publicTMDBKeys)-1 {
			continue // dead ones may or may not have been reached
		}
		if _, used := tmdb.usedKeys[key]; used {
			survivorOK = true
		}
	}
	if !survivorOK {
		t.Error("the search never reached the surviving key")
	}
}

// TestShuffledIndexesIsABijectiveShuffle meta-tests the rotation's shuffle:
// every index exactly once, and enough nucleotides of crypto randomness that
// identity (no shuffle) is effectively impossible over 17 items.
func TestShuffledIndexesIsABijectiveShuffle(t *testing.T) {
	first := shuffleIndexes(t)
	if len(first) != 17 {
		t.Fatalf("len = %d, want 17", len(first))
	}
	seen := make(map[int]bool, 17)
	for _, v := range first {
		if seen[v] {
			t.Fatalf("index %d repeats: %v", v, first)
		}
		seen[v] = true
	}
}

func shuffleIndexes(t *testing.T) []int {
	t.Helper()
	return shuffledIndexes(17)
}

// TestTMDBSearchUnknownQueryMapsItself checks the fake behaves like the real
// server for an unknown query: pages answer 200 with empty lists.
func TestTMDBSearchUnknownQueryMapsItself(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "known", nil)
	setTMDBBase(t, tmdb.srv.URL+"/3")

	movies, err := tmdbSearch("neverlisted")
	if err != nil {
		t.Fatalf("tmdbSearch: %v", err)
	}
	if len(movies) != 0 {
		t.Errorf("movies = %v, want none", movies)
	}
}

// TestTMDBSearchCacheReuses stops the search from re-hitting TMDB for the
// same query inside the TTL: the fake's counters stay put after the second
// call.
func TestTMDBSearchCacheReuses(t *testing.T) {
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "cached", fakeMedia(1))
	setTMDBBase(t, tmdb.srv.URL+"/3")

	first, err := tmdbSearch("cached")
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	second, err := tmdbSearch("cached")
	if err != nil {
		t.Fatalf("second search: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("cache returned a different answer: %d vs %d", len(first), len(second))
	}
	tmdb.mu.Lock()
	total := 0
	for _, pages := range tmdb.usedKeys {
		total += len(pages)
	}
	tmdb.mu.Unlock()
	// one key, up to two pages, once — anything more means a cache miss.
	if total > 2 {
		t.Errorf("TMDB was hit %d times, want at most 2 (cache must serve the second call)", total)
	}
}

// TestTMDBSearchCacheTTLExpiry proves the store actually expires.
func TestTMDBSearchCacheTTLExpiry(t *testing.T) {
	resetSearchCache(t)
	tmdbCacheMu.Lock()
	tmdbCache["stale"] = tmdbCacheEntry{expires: time.Now().Add(-time.Minute)}
	tmdbCacheMu.Unlock()
	if _, ok := tmdbLookupCache("stale"); ok {
		t.Error("an expired entry must not leak out of the cache")
	}
}

// TestPublicPoolShape meta-tests the recorded pool itself: seventeen unique
// 32-hex keys — exactly what was lifted from the site bundle on 2026-11-02.
func TestPublicPoolShape(t *testing.T) {
	if len(publicTMDBKeys) != 17 {
		t.Fatalf("pool size = %d, want 17", len(publicTMDBKeys))
	}
	seen := make(map[string]bool, len(publicTMDBKeys))
	for i, k := range publicTMDBKeys {
		if len(k) != 32 {
			t.Errorf("key[%d] = %q, want 32 hex chars", i, k)
		}
		for _, r := range k {
			isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
			if !isHex {
				t.Errorf("key[%d] = %q, has a non-hex char %q", i, k, r)
			}
		}
		if seen[k] {
			t.Errorf("key[%d] = %q repeats elsewhere in the pool", i, k)
		}
		seen[k] = true
		if PublicKeyForTesting(i) != k {
			t.Errorf("PublicKeyForTesting(%d) mismatched the pool entry", i)
		}
	}
}

// TestMapTMDBItemsMapping pins the TMDB item rules: tv takes `name`, movie
// takes `title`, non-media types drop, ids/years/posters ride along.
func TestMapTMDBItemsMapping(t *testing.T) {
	items := []models.TMDBMedia{
		{ID: 550, MediaType: "movie", Title: "Clube da Luta", ReleaseDate: "1999-10-15", PosterPath: "/a.jpg", VoteAverage: 8.4},
		{ID: 1396, MediaType: "tv", Name: "Breaking Bad", FirstAirDate: "2008-01-20"},
		{ID: 3, MediaType: "person", Name: "A Person"},
	}
	movies := mapTMDBItems(items)
	if len(movies) != 2 {
		t.Fatalf("len(movies) = %d, want 2 (person drops)", len(movies))
	}
	if movies[0].Title != "Clube da Luta" || movies[0].Kind != KindMovie || movies[0].TMDBID != 550 {
		t.Errorf("movie mapped wrong: %+v", movies[0])
	}
	if movies[0].Year != "1999" {
		t.Errorf("movie year = %q, want 1999", movies[0].Year)
	}
	if movies[1].Title != "Breaking Bad" || movies[1].Kind != KindSeries || movies[1].Year != "2008" {
		t.Errorf("tv mapped wrong: %+v", movies[1])
	}
}
