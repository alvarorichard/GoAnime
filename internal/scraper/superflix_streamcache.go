package scraper

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// streamCacheEntry is the browser-gated facts for one piece of content: the
// warezcdn player host and the 32-hex content hash. Together they let us hit the
// (ungated) getVideo endpoint over plain HTTP — no browser — for fresh signed
// HLS links on every replay.
type streamCacheEntry struct {
	Host string `json:"host"` // e.g. https://xn--kcksk7a2bl5le7b6doc1h3f.com
	Hash string `json:"hash"` // 32-hex warezcdn content id
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
	_ = json.Unmarshal(data, &sc.entries)
}

func (sc *streamCache) get(key string) (streamCacheEntry, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ensureLoaded()
	e, ok := sc.entries[key]
	return e, ok && e.Host != "" && e.Hash != ""
}

func (sc *streamCache) put(key string, e streamCacheEntry) {
	if e.Host == "" || e.Hash == "" {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ensureLoaded()
	if cur, ok := sc.entries[key]; ok && cur == e {
		return
	}
	sc.entries[key] = e
	if data, err := json.MarshalIndent(sc.entries, "", "  "); err == nil {
		_ = os.WriteFile(sc.file(), data, 0o600)
	}
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
