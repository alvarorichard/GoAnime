package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withFreshStreamCache swaps the package-global stream cache for an isolated,
// temp-file-backed one, so a test's writes don't touch the user's real cache or
// leak between tests.
func withFreshStreamCache(t *testing.T) {
	t.Helper()
	prev := defaultStreamCache
	defaultStreamCache = &streamCache{path: t.TempDir() + "/cache.json"}
	t.Cleanup(func() { defaultStreamCache = prev })
}

// sfStreamServer stands up a fake SuperFlix player pipeline: source → redirect →
// video page → getVideo. Returns the base URL.
func sfStreamServer(t *testing.T) string {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/player/source", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/hash123"}}`, srv.URL)
	})
	mux.HandleFunc("/video/hash123", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, realPlayerPage)
	})
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"securedLink":"https://cdn.example/master.m3u8"}`)
	})
	return srv.URL
}

// StreamFromServer must cache (host, hash) so the next play of the same episode
// replays over plain HTTP — the whole point of the fast path.
func TestStreamFromServer_PopulatesCache(t *testing.T) {
	withFreshStreamCache(t)
	base := sfStreamServer(t)

	c := NewClientForTest(base)
	tokens := &SuperFlixTokens{ContentID: "1", PageToken: "tok"}

	res, err := c.StreamFromServer(context.Background(), tokens, "159462", "serie", "42821", "1", "3")
	require.NoError(t, err)
	assert.Contains(t, res.StreamURL, "master.m3u8")

	// The exact episode key must now be cached, pointing at this player host/hash.
	ent, ok := defaultStreamCache.get(streamCacheKey("serie", "42821", "1", "3"))
	require.True(t, ok, "StreamFromServer must cache the resolved (host, hash)")
	assert.Equal(t, base, ent.Host)
	assert.Equal(t, "hash123", ent.Hash)
}

// TryCachedStream replays purely from the cache — no source/redirect calls, only
// the ungated getVideo + player-extras — and reports a miss when nothing is
// cached.
func TestTryCachedStream(t *testing.T) {
	withFreshStreamCache(t)
	base := sfStreamServer(t)
	c := NewClientForTest(base)

	// Miss: nothing cached yet.
	_, ok := c.TryCachedStream(context.Background(), "serie", "42821", "1", "3")
	assert.False(t, ok, "an unseen episode is a cache miss")

	// Seed the cache as a prior resolution would.
	defaultStreamCache.put(streamCacheKey("serie", "42821", "1", "3"),
		streamCacheEntry{Host: base, Hash: "hash123"})

	res, ok := c.TryCachedStream(context.Background(), "serie", "42821", "1", "3")
	require.True(t, ok, "a seen episode must replay from cache")
	assert.Contains(t, res.StreamURL, "master.m3u8")
	assert.Equal(t, base+"/", res.Referer)
	// The extras (audio tracks + subtitles) must come along on the fast path too.
	assert.NotEmpty(t, res.DefaultAudio)
	require.Len(t, res.Subtitles, 1)
	assert.Equal(t, "Portuguese", res.Subtitles[0].Lang)
}

// A server that resolves to SuperFlix's NATIVE player (…/player/native/media/…)
// must fail fast — no getVideo attempt (it answers 405) — and must NOT cache the
// pair, which can never replay. Regression for the 2026-07-22 rollout that
// poisoned user caches and cost a wasted solve on every play.
func TestStreamFromServer_NativePlayerFailsFastAndIsNotCached(t *testing.T) {
	withFreshStreamCache(t)

	var getVideoHit atomic.Bool
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/player/source", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/player/native/media/185176?embed=1&mt=tok"}}`, srv.URL)
	})
	mux.HandleFunc("/player/native/media/185176", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "<html>native player shell</html>")
	})
	mux.HandleFunc("/player/native/media/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		getVideoHit.Store(true)
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	c := NewClientForTest(srv.URL)
	tokens := &SuperFlixTokens{ContentID: "1", PageToken: "tok"}

	_, err := c.StreamFromServer(context.Background(), tokens, "159462", "filme", "10214", "", "")
	require.Error(t, err, "the native player has no getVideo API; the caller must fall back to the sniff")
	assert.False(t, getVideoHit.Load(), "must not fire the doomed getVideo round-trip")

	_, ok := defaultStreamCache.get(streamCacheKey("filme", "10214", "", ""))
	assert.False(t, ok, "a native-player (host, hash) can never replay and must not be cached")
}

// A poisoned cache entry pointing at the native player (written before the
// rollout was recognized) must self-heal: reported as a miss, deleted, and no
// network round-trip attempted.
func TestTryCachedStream_NativePlayerEntrySelfHeals(t *testing.T) {
	withFreshStreamCache(t)

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	t.Cleanup(srv.Close)

	key := streamCacheKey("filme", "10214", "", "")
	defaultStreamCache.put(key, streamCacheEntry{Host: srv.URL + "/player/native/media", Hash: "185176"})
	require.False(t, HasCachedStream("filme", "10214", "", ""),
		"put must already refuse a native-player host")

	// Simulate an entry that predates the put guard (written by an old build).
	defaultStreamCache.mu.Lock()
	defaultStreamCache.ensureLoaded()
	defaultStreamCache.entries[key] = streamCacheEntry{Host: srv.URL + "/player/native/media", Hash: "185176"}
	defaultStreamCache.mu.Unlock()

	c := NewClientForTest(srv.URL)
	_, ok := c.TryCachedStream(context.Background(), "filme", "10214", "", "")
	assert.False(t, ok, "a native-player entry must be a miss")
	assert.Zero(t, requests.Load(), "no round-trip may be wasted on a never-replayable entry")

	_, stillThere := defaultStreamCache.get(key)
	assert.False(t, stillThere, "the poisoned entry must be deleted (self-heal)")
}

// A cached host that has rotated out (getVideo fails) must be reported as a miss
// so the caller re-resolves, rather than returning a dead stream.
func TestTryCachedStream_StaleHostIsMiss(t *testing.T) {
	withFreshStreamCache(t)

	// Point the cache at a server that 500s on getVideo.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/index.php") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprint(w, "x")
	}))
	t.Cleanup(dead.Close)

	defaultStreamCache.put(streamCacheKey("filme", "999", "", ""),
		streamCacheEntry{Host: dead.URL, Hash: "gone"})

	c := NewClientForTest(dead.URL)
	_, ok := c.TryCachedStream(context.Background(), "filme", "999", "", "")
	assert.False(t, ok, "a stale/rotated cached host must be a miss, not a dead stream")
}
