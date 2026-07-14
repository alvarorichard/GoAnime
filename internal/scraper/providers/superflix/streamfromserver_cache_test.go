package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
