package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The search cache has to survive the process, because GoAnime IS a process.
//
// It was a sync.Map, so every invocation started empty and every repeat of a
// title was a fresh request against /pesquisar — the endpoint whose small,
// sticky allowance is what makes searches here fail. A user hunting one movie
// across a few runs spent most of their allowance re-fetching answers the tool
// already had.

func tempSearchCache(t *testing.T) *searchCache {
	t.Helper()
	return &searchCache{path: filepath.Join(t.TempDir(), "search.json")}
}

// countingOrigin serves one result and counts how many times it was asked.
func countingOrigin(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><div class="group/card">`+
			`<img alt="Cowboy Bebop" src="https://img.example/1.jpg">`+
			`<button data-msg="TMDB ID copiado!" data-copy="30991"></button>`+
			`<button data-msg="Link copiado!" data-copy="https://x/serie/30991"></button>`+
			`<div class="mt-3"><span>1998</span><span>Anime</span></div>`+
			`</div></body></html>`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func cachedClient(t *testing.T, url string, sc *searchCache) *SuperFlixClient {
	t.Helper()
	c := NewSuperFlixClient()
	c.SetTestConfig(url, &http.Client{Timeout: 5 * time.Second})
	c.persistentSearch = sc
	return c
}

// A second RUN — a new client with an empty in-memory map — must not re-ask.
func TestSearchCache_SurvivesANewProcess(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := countingOrigin(t, &hits)
	sc := tempSearchCache(t)

	first, err := cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, int32(1), hits.Load())

	// A fresh client is what a fresh `goanime` invocation looks like.
	second, err := cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	assert.Len(t, second, 1)
	assert.Equal(t, int32(1), hits.Load(),
		"the second run asked the host again for an answer already on disk")
	assert.Equal(t, first[0].TMDBID, second[0].TMDBID)
}

// The slug and the words are the same question, so they must be one entry.
func TestSearchCache_NormalisesTheQuery(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := countingOrigin(t, &hits)
	sc := tempSearchCache(t)

	_, err := cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "Cowboy-Bebop")
	require.NoError(t, err)
	_, err = cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "  cowboy   bebop ")
	require.NoError(t, err)

	assert.Equal(t, int32(1), hits.Load(), "the same query in another spelling cost a second request")
}

// "This host does not have it" is an answer worth keeping. Re-asking for a
// title the catalogue does not carry is exactly the repeat that burns the
// allowance.
func TestSearchCache_RemembersAnEmptyAnswer(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><p>Nenhum resultado encontrado</p></body></html>`)
	}))
	defer srv.Close()
	sc := tempSearchCache(t)

	for range 3 {
		res, err := cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "nao existe")
		require.NoError(t, err)
		assert.Empty(t, res)
	}
	assert.Equal(t, int32(1), hits.Load(), "an empty answer must be remembered like any other")
}

// A stale entry is refetched rather than served forever.
func TestSearchCache_RefetchesAfterTheTTL(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := countingOrigin(t, &hits)
	sc := tempSearchCache(t)

	_, err := cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	require.Equal(t, int32(1), hits.Load())

	// Age the entry past the TTL.
	sc.mu.Lock()
	for k, e := range sc.entries {
		e.At = time.Now().Add(-searchCacheTTL - time.Minute).Unix()
		sc.entries[k] = e
	}
	sc.mu.Unlock()

	_, err = cachedClient(t, srv.URL, sc).SearchMediaWithContext(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	assert.Equal(t, int32(2), hits.Load(), "a stale answer must be refetched")
}

// A test binary must never read or write the developer's real cache: it would
// seed their next search, and let a parser test pass on yesterday's data.
func TestSearchCache_IsDisabledInTests(t *testing.T) {
	t.Parallel()
	assert.Empty(t, (&searchCache{}).file(),
		"a test wrote to the real user cache directory")
}

// The file must not grow without bound, and the oldest entries are the ones to
// lose.
func TestSearchCache_EvictsTheOldest(t *testing.T) {
	t.Parallel()
	sc := tempSearchCache(t)

	for i := range searchCacheMaxEntries + 20 {
		sc.mu.Lock()
		sc.ensureLoadedLocked()
		sc.entries[fmt.Sprintf("q%04d", i)] = searchCacheEntry{At: int64(i + 1)}
		sc.mu.Unlock()
	}
	sc.put("newest", nil)

	sc.mu.Lock()
	defer sc.mu.Unlock()
	assert.LessOrEqual(t, len(sc.entries), searchCacheMaxEntries)
	_, keptNewest := sc.entries["newest"]
	assert.True(t, keptNewest, "the entry just written was evicted")
	_, keptOldest := sc.entries["q0000"]
	assert.False(t, keptOldest, "the oldest entry survived while newer ones were dropped")
}

// And it really is a file: a corrupt one must degrade to a miss, not a crash.
func TestSearchCache_SurvivesACorruptFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "search.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))

	sc := &searchCache{path: path}
	_, ok := sc.get("anything")
	assert.False(t, ok, "a corrupt cache must read as empty")

	sc.put("x", nil)
	got, ok := sc.get("x")
	assert.True(t, ok, "and must still be writable afterwards")
	assert.Empty(t, got)
}
