package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsBloggerPlayerURL_2026_09_01 pins the classification that keeps the
// FirePlayer getVideo contract away from Blogger-hosted titles.
//
// Without it, a chosen server that resolved to
// https://www.blogger.com/video.g?token=<token> was split into
// host="https://www.blogger.com" + hash="video.g" (the token, the only thing
// identifying the video, went with the discarded query string), and the
// pipeline POSTed to blogger.com/player/index.php?data=video.g&do=getVideo —
// a 404 reported to the user as "the chosen server failed".
func TestIsBloggerPlayerURL_2026_09_01(t *testing.T) {
	t.Parallel()

	for _, u := range []string{
		"https://www.blogger.com/video.g?token=AD6v5dxWLxemiwKpvfe",
		"https://blogger.com/video.g?token=abc",
		"http://www.blogger.com/video.g?token=abc",
	} {
		assert.Truef(t, isBloggerPlayerURL(u), "%q is a Blogger video page", u)
	}

	for _, u := range []string{
		"",
		"https://www.blogger.com/",
		"https://notblogger.com/video.g?token=abc",
		"https://player.best/video/6f1896bb1bfc3d0bcd2cba28ad968dd4",
		"https://evil.test/?x=https://www.blogger.com/video.g?token=abc",
	} {
		assert.Falsef(t, isBloggerPlayerURL(u), "%q is not a Blogger video page", u)
	}
}

// TestIsReplayablePlayerHost pins which hosts may enter the stream cache. Only
// hosts that implement /player/index.php?do=getVideo can be replayed without a
// browser; caching any other costs a doomed round-trip on every later play.
func TestIsReplayablePlayerHost(t *testing.T) {
	t.Parallel()

	assert.True(t, isReplayablePlayerHost("https://xn--tckasiu6cvova0eb5fua2449g98vg.best"))

	assert.False(t, isReplayablePlayerHost("https://superflixapi.beer/player/native/media/123"),
		"the native player answers getVideo with 405")
	assert.False(t, isReplayablePlayerHost("https://www.blogger.com"),
		"blogger answers getVideo with 404")
	assert.False(t, isReplayablePlayerHost("https://www.blogger.com/video.g?token=abc"))
}

// TestStreamCache_RefusesBloggerEntries is the regression for the poisoned
// cache found on a real user's disk:
//
//	"serie:54728:2:4": {"host": "https://www.blogger.com", "hash": "video.g"}
//
// It looked like a valid pair, so it was written, and every later play replayed
// the same 404 before falling through to a 90s embed sniff that could not match.
func TestStreamCache_RefusesBloggerEntries(t *testing.T) {
	sc := &streamCache{path: t.TempDir() + "/c.json"}

	sc.put("serie:54728:2:4", streamCacheEntry{Host: "https://www.blogger.com", Hash: "video.g"})
	_, ok := sc.get("serie:54728:2:4")
	assert.False(t, ok, "a Blogger pair must never be cached")

	sc.put("filme:603", streamCacheEntry{Host: "https://player.best", Hash: "deadbeef"})
	_, ok = sc.get("filme:603")
	assert.True(t, ok, "a real FirePlayer pair still caches")
}

// TestStreamFromCache_SelfHealsPoisonedEntry covers the entries already written
// to disk before the guard existed: they must be dropped on read, not replayed.
func TestStreamFromCache_SelfHealsPoisonedEntry(t *testing.T) {
	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: t.TempDir() + "/c.json"}

	// Seed directly, bypassing put's guard, the way an old build left it.
	defaultStreamCache.mu.Lock()
	defaultStreamCache.ensureLoaded()
	defaultStreamCache.entries["serie:54728:2:4"] = streamCacheEntry{
		Host: "https://www.blogger.com", Hash: "video.g",
	}
	defaultStreamCache.mu.Unlock()

	c := NewClientForTest("https://sf.test")
	res, ok := c.streamFromCache(t.Context(), "serie:54728:2:4")
	require.False(t, ok, "a poisoned entry must not be replayed")
	assert.Nil(t, res)

	_, still := defaultStreamCache.get("serie:54728:2:4")
	assert.False(t, still, "and it must be evicted so it never costs another play")
}

// TestBloggerStreamResult pins what the caller hands the player layer: the
// Blogger URL untouched, and no CDN Referer/User-Agent, since none of the
// FirePlayer CDN's rules apply to it.
func TestBloggerStreamResult(t *testing.T) {
	t.Parallel()
	const u = "https://www.blogger.com/video.g?token=AD6v5dx"

	c := NewClientForTest("https://sf.test")
	res := c.bloggerStreamResult(u, &SuperFlixTokens{Title: "Beyblade"}, "")

	require.NotNil(t, res)
	assert.Equal(t, u, res.StreamURL, "the token must survive — it identifies the video")
	assert.Equal(t, "Beyblade", res.Title)
	assert.Empty(t, res.Referer)
	assert.Empty(t, res.UserAgent)
}

// A nil token set (no server list) must not panic the Blogger path.
func TestBloggerStreamResult_NilTokens(t *testing.T) {
	t.Parallel()
	c := NewClientForTest("https://sf.test")
	res := c.bloggerStreamResult("https://www.blogger.com/video.g?token=x", nil, "")
	require.NotNil(t, res)
	assert.Empty(t, res.Title)
}
