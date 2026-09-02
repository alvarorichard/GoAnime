package superflix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// streamCacheEntry is the browser-gated facts for one piece of content: the
// warezcdn player host and the 32-hex content hash. Together they let us hit the
// (ungated) getVideo endpoint over plain HTTP — no browser — for fresh signed
// HLS links on every replay.
type streamCacheEntry struct {
	Host string `json:"host"` // e.g. https://xn--kcksk7a2bl5le7b6doc1h3f.com
	Hash string `json:"hash"` // 32-hex warezcdn content id
	// StreamURL is the signed media URL captured on the last successful solve,
	// with the Referer and User-Agent the CDN binds it to.
	//
	// The (host, hash) pair above only replays through the player's getVideo
	// endpoint, and the current player has none — it answers 404, so every
	// "cached" play fell through to a full browser solve (~7s). The signed URL
	// itself outlives the solve by a useful margin (measured: still 200 sixteen
	// minutes after capture), so keeping it turns a re-watch or a resume into a
	// single probe. It is never trusted blindly: streamFromCache probes it and
	// falls through to a re-solve on any rejection.
	StreamURL    string              `json:"stream_url,omitempty"`
	StreamRef    string              `json:"stream_referer,omitempty"`
	StreamUA     string              `json:"stream_user_agent,omitempty"`
	SignedAt     int64               `json:"signed_at,omitempty"` // unix seconds
	DefaultAudio []string            `json:"default_audio,omitempty"`
	Subtitles    []SuperFlixSubtitle `json:"subtitles,omitempty"`
	ExtrasCached bool                `json:"extras_cached,omitempty"`
}

// signedURLTTL bounds how long a cached signed URL is worth probing.
//
// The CDN's real expiry is not published; one was measured still alive 16
// minutes after capture. This is deliberately shorter than any observed
// lifetime is likely to be — the probe, not this constant, is the authority on
// whether a URL still works. Its only job is to stop us probing a URL so old
// that a re-solve is the better bet anyway.
const signedURLTTL = 30 * time.Minute

// freshSignedURL returns the cached signed URL and its headers when one is
// present and still within signedURLTTL.
func (e streamCacheEntry) freshSignedURL() (streamURL, referer, userAgent string, ok bool) {
	if e.StreamURL == "" || e.SignedAt == 0 {
		return "", "", "", false
	}
	if time.Since(time.Unix(e.SignedAt, 0)) > signedURLTTL {
		return "", "", "", false
	}
	return e.StreamURL, e.StreamRef, e.StreamUA, true
}

// streamCache persists tmdb→(host,hash) so the headed browser runs only on the
// FIRST play of a title (or when the cached host rotates out and getVideo fails).
// Stored as a small JSON map under the user cache dir.
type streamCache struct {
	mu      sync.Mutex
	path    string
	entries map[string]streamCacheEntry
	loaded  bool
}

var defaultStreamCache = &streamCache{}

func (sc *streamCache) file() string {
	if sc.path != "" {
		return sc.path
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	d := filepath.Join(dir, "goanime")
	_ = os.MkdirAll(d, 0o700)
	sc.path = filepath.Join(d, "superflix-stream-cache.json")
	return sc.path
}

func (sc *streamCache) ensureLoaded() {
	if sc.loaded {
		return
	}
	sc.loaded = true
	sc.entries = map[string]streamCacheEntry{}
	data, err := os.ReadFile(sc.file())
	if err != nil {
		return
	}
	_ = jsonx.Unmarshal(data, &sc.entries)
}

func (sc *streamCache) get(key string) (streamCacheEntry, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ensureLoaded()
	e, ok := sc.entries[key]
	return e, ok && e.Host != "" && e.Hash != ""
}

func (sc *streamCache) put(key string, e streamCacheEntry) {
	// A host that does not implement the getVideo contract can never replay, so
	// a central refusal here keeps every put site from poisoning the cache.
	// Blogger is the case that got through: its player URL split into
	// host="https://www.blogger.com", hash="video.g", which looked like a valid
	// pair and was cached, then failed the same way on every later play.
	if e.Host == "" || e.Hash == "" || !isReplayablePlayerHost(e.Host) {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ensureLoaded()
	sc.entries[key] = e
	if data, err := json.MarshalIndent(sc.entries, "", "  "); err == nil {
		_ = os.WriteFile(sc.file(), data, 0o600)
	}
}

// del removes a cache entry and persists the change. Used when a cached
// (host, hash) still signs URLs but the CDN rejects them (host rotated out),
// so the next play re-resolves through the browser instead of replaying a dead
// link.
func (sc *streamCache) del(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ensureLoaded()
	if _, ok := sc.entries[key]; !ok {
		return
	}
	delete(sc.entries, key)
	if data, err := json.MarshalIndent(sc.entries, "", "  "); err == nil {
		_ = os.WriteFile(sc.file(), data, 0o600)
	}
}

// HasCachedStream reports whether a browser-free replay entry exists for the
// given content. Pure cache lookup — no network I/O — so callers (e.g. the
// next-episode prefetch) can decide cheaply whether a warm-up is needed.
func HasCachedStream(mediaType, mediaID, season, episode string) bool {
	_, ok := defaultStreamCache.get(streamCacheKey(mediaType, mediaID, season, episode))
	return ok
}

// streamCacheKey identifies one playable unit (movie, or a specific episode).
func streamCacheKey(mediaType, mediaID, season, episode string) string {
	if mediaType == "serie" {
		s, e := season, episode
		if s == "" {
			s = "1"
		}
		if e == "" {
			e = "1"
		}
		return "serie:" + mediaID + ":" + s + ":" + e
	}
	return "filme:" + mediaID
}
