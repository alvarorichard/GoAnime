package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamCacheKey verifies the cache key shape for movies and episodes,
// including the season/episode defaulting.
func TestStreamCacheKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                              string
		mediaType, mediaID, season, epNum string
		want                              string
	}{
		{"movie", "filme", "1048794", "", "", "filme:1048794"},
		{"movie ignores s/e", "filme", "1048794", "3", "4", "filme:1048794"},
		{"episode explicit", "serie", "76479", "2", "5", "serie:76479:2:5"},
		{"episode defaults", "serie", "76479", "", "", "serie:76479:1:1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := streamCacheKey(tt.mediaType, tt.mediaID, tt.season, tt.epNum)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStreamCache_PutGet verifies the on-disk cache round-trips and rejects
// incomplete entries.
func TestStreamCache_PutGet(t *testing.T) {
	t.Parallel()
	sc := &streamCache{path: filepath.Join(t.TempDir(), "cache.json")}

	_, ok := sc.get("filme:1")
	assert.False(t, ok, "empty cache must miss")

	sc.put("filme:1", streamCacheEntry{Host: "https://h.com", Hash: "abc"})
	got, ok := sc.get("filme:1")
	require.True(t, ok)
	assert.Equal(t, "https://h.com", got.Host)
	assert.Equal(t, "abc", got.Hash)

	// Incomplete entries are ignored.
	sc.put("filme:2", streamCacheEntry{Host: "https://h.com"})
	_, ok = sc.get("filme:2")
	assert.False(t, ok, "entry without hash must not be stored")

	// A fresh cache pointed at the same file reloads persisted entries.
	sc2 := &streamCache{path: sc.path}
	got2, ok := sc2.get("filme:1")
	require.True(t, ok, "persisted entry must reload from disk")
	assert.Equal(t, "abc", got2.Hash)
}

// failingEmbedSolver implements both cfSolver and embedStreamSolver but fails the
// test if SniffEmbedStream is ever called — proving the cache path never touches
// the browser.
type failingEmbedSolver struct{ t *testing.T }

func (f *failingEmbedSolver) Solve(context.Context, string, time.Duration) (*CFSolveResult, error) {
	return nil, fmt.Errorf("Solve must not be called")
}

func (f *failingEmbedSolver) SniffEmbedStream(context.Context, string, time.Duration) (*CFStreamResult, error) {
	f.t.Fatal("SniffEmbedStream (browser) must NOT be called on a cache hit")
	return nil, nil
}

// TestGetStreamURL_CacheHitSkipsBrowser is the core proof for the browser-free
// replay: with a cached (host, hash) the stream is resolved over plain HTTP via
// getVideo, and the browser solver is never invoked.
func TestGetStreamURL_CacheHitSkipsBrowser(t *testing.T) {
	// Not parallel: mutates the package-global defaultStreamCache.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "do=getVideo")
		assert.Equal(t, "data123", r.URL.Query().Get("data"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"hls":true,"securedLink":"%s/cdn/hls/x/master.m3u8?md5=z&expires=9","videoImage":"%s/thumb.jpg"}`, srv0(r), srv0(r))
	}))
	t.Cleanup(srv.Close)

	// Point the cached host at the httptest server.
	key := streamCacheKey("filme", "999001", "", "")
	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}
	defaultStreamCache.put(key, streamCacheEntry{Host: srv.URL, Hash: "data123"})

	client := NewSuperFlixClient()
	// Plain client so getVideo reaches the httptest server (the production
	// netx.SafeScraperTransport blocks loopback as an anti-SSRF measure).
	client.client = srv.Client()
	client.browserSolver = &failingEmbedSolver{t: t}

	res, err := client.GetStreamURL(context.Background(), "filme", "999001", "", "")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.StreamURL, "master.m3u8")
	assert.Equal(t, srv.URL+"/", res.Referer)
}

// srv0 returns the request's own scheme+host (the httptest base) so the fixture
// JSON points back at the test server.
func srv0(r *http.Request) string {
	return "http://" + r.Host
}
