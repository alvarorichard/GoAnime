package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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
		// The cache path makes two plain-HTTP calls: getVideo for a fresh signed
		// link, and the player page for the subtitle/audio tracks.
		if strings.HasPrefix(r.URL.Path, "/video/") {
			assert.Equal(t, "/video/data123", r.URL.Path)
			_, _ = fmt.Fprint(w, realPlayerPage)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/cdn/") {
			// Live CDN: the cache path probes the freshly signed master.m3u8
			// before trusting it, so serve a 200 here.
			_, _ = fmt.Fprint(w, "#EXTM3U")
			return
		}
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
	// The CDN 403s the signed playlist for anything but the player's own
	// /video/<hash> page, so the cached replay must rebuild that exact URL —
	// this used to pin the bare origin and shipped the defect. See
	// playerRefererFor.
	assert.Equal(t, srv.URL+"/video/data123", res.Referer)

	// The cache path used to return a bare stream — no subtitles, no audio info —
	// so a replayed episode silently lost both. It must carry them now.
	assert.Equal(t, []string{"por", "und", "eng", "spa", "kor", "jpn", "chi", "und"}, res.DefaultAudio,
		"a cached replay must still know which audio tracks the stream has")
	require.Len(t, res.Subtitles, 1, "a cached replay must still carry the subtitle track")
	assert.Equal(t, "Portuguese", res.Subtitles[0].Lang)
}

func TestStreamCache_PersistsExtrasForTheNextReplay(t *testing.T) {
	// Not parallel: swaps the package-level cache.
	var playerPageHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/video/") {
			playerPageHits++
			_, _ = fmt.Fprint(w, realPlayerPage)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"securedLink":"%s/cdn/master.m3u8"}`, srv0(r))
	}))
	t.Cleanup(srv.Close)

	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}
	key := streamCacheKey("filme", "42", "", "")
	defaultStreamCache.put(key, streamCacheEntry{Host: srv.URL, Hash: "data42"})

	c := NewClientForTest(srv.URL)
	first, ok := c.TryCachedStream(context.Background(), "filme", "42", "", "")
	require.True(t, ok)
	require.NotEmpty(t, first.Subtitles, "first replay must wait for subtitle extraction")
	assert.Equal(t, 1, playerPageHits)

	second, ok := c.TryCachedStream(context.Background(), "filme", "42", "", "")
	require.True(t, ok)
	require.NotEmpty(t, second.Subtitles, "cached subtitle metadata must survive the next replay")
	assert.Equal(t, 1, playerPageHits, "second replay must reuse cached extras, not refetch them")
}

// srv0 returns the request's own scheme+host (the httptest base) so the fixture
// JSON points back at the test server.
func srv0(r *http.Request) string {
	return "http://" + r.Host
}

// TestStreamCache_Del verifies eviction removes only the target entry and
// persists the deletion, and that deleting a missing key is a safe no-op.
func TestStreamCache_Del(t *testing.T) {
	t.Parallel()
	sc := &streamCache{path: filepath.Join(t.TempDir(), "cache.json")}
	sc.put("filme:1", streamCacheEntry{Host: "https://h.com", Hash: "abc"})
	sc.put("filme:2", streamCacheEntry{Host: "https://h.com", Hash: "def"})

	sc.del("filme:1")
	_, ok := sc.get("filme:1")
	assert.False(t, ok, "deleted entry must miss")
	_, ok = sc.get("filme:2")
	assert.True(t, ok, "unrelated entry must survive")

	// A fresh cache on the same file must not see the deleted entry.
	sc2 := &streamCache{path: sc.path}
	_, ok = sc2.get("filme:1")
	assert.False(t, ok, "deletion must persist to disk")

	// Deleting a missing key is a no-op (no panic).
	sc.del("filme:missing")
}

// TestStreamURLDead verifies the CDN-liveness probe: a definitive rejection
// (403/404/410) is dead; a playable or ambiguous response is not.
func TestStreamURLDead(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"ok", http.StatusOK, false},
		{"partial content", http.StatusPartialContent, false},
		{"forbidden host rotated", http.StatusForbidden, true},
		{"not found", http.StatusNotFound, true},
		{"gone", http.StatusGone, true},
		{"server error is ambiguous", http.StatusInternalServerError, false},
		{"too many requests is ambiguous", http.StatusTooManyRequests, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "bytes=0-1", r.Header.Get("Range"), "probe must be a tiny ranged GET")
				assert.Equal(t, srv0(r)+"/", r.Header.Get("Referer"))
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(srv.Close)
			c := NewSuperFlixClient()
			c.client = srv.Client()
			got := c.streamURLDead(context.Background(), srv.URL+"/cdn/hls/x/master.m3u8", srv.URL+"/", SuperFlixUserAgent)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("unreachable host is not dead", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		client := srv.Client()
		deadURL := srv.URL + "/cdn/x"
		srv.Close() // connection refused — ambiguous, must not evict.
		c := NewSuperFlixClient()
		c.client = client
		assert.False(t, c.streamURLDead(context.Background(), deadURL, "http://x/", SuperFlixUserAgent))
	})
}

// TestTryCachedStream_StaleCDNInvalidatesEntry is the regression for the
// dead-cache bug: getVideo keeps signing links on a rotated-out host, the CDN
// answers 403, and mpv dies. The probe must catch that, report a cache miss,
// and evict the entry so the next play re-resolves through the browser.
func TestTryCachedStream_StaleCDNInvalidatesEntry(t *testing.T) {
	// Not parallel: swaps the package-global defaultStreamCache.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/video/") {
			_, _ = fmt.Fprint(w, realPlayerPage)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/cdn/") {
			// Host rotated out of the CDN: signed link is rejected.
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"securedLink":"%s/cdn/hls/x/master.m3u8?md5=z&expires=9"}`, srv0(r))
	}))
	t.Cleanup(srv.Close)

	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}
	key := streamCacheKey("filme", "555", "", "")
	defaultStreamCache.put(key, streamCacheEntry{Host: srv.URL, Hash: "data555"})

	c := NewClientForTest(srv.URL)
	res, ok := c.TryCachedStream(context.Background(), "filme", "555", "", "")
	assert.False(t, ok, "a 403 from the CDN must be treated as a cache miss")
	assert.Nil(t, res)

	_, ok = defaultStreamCache.get(key)
	assert.False(t, ok, "the stale entry must be evicted so the next play re-resolves via the browser")
}
