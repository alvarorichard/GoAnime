package superflix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// streamCacheEntry is the browser-gated facts for one piece of content: the
// warezcdn player host and the 32-hex content hash. Together they let us hit the
// (ungated) getVideo endpoint over plain HTTP — no browser — for fresh signed
// HLS links on every replay.
type streamCacheEntry struct {
	Host         string              `json:"host"` // e.g. https://xn--kcksk7a2bl5le7b6doc1h3f.com
	Hash         string              `json:"hash"` // 32-hex warezcdn content id
	DefaultAudio []string            `json:"default_audio,omitempty"`
	Subtitles    []SuperFlixSubtitle `json:"subtitles,omitempty"`
	ExtrasCached bool                `json:"extras_cached,omitempty"`
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
	// A native-player host can never replay (its getVideo answers 405), so a
	// central refusal here keeps every put site from poisoning the cache.
	if e.Host == "" || e.Hash == "" || isNativePlayerHost(e.Host) {
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
