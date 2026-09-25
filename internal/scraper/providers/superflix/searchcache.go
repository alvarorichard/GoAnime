package superflix

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Searching this host twice for the same thing must not cost two requests.
//
// Its /pesquisar endpoint has a small allowance and a sticky guard: go over it
// and the endpoint refuses, and every request sent during the refusal renews it
// (see backoff.go for the measurements). The client already had a search cache
// and it was a sync.Map — per process. GoAnime is run as a command: every
// invocation started with an empty cache, so re-running it and typing the same
// title again was a fresh request every time. Watching a user hunt for one
// movie across a few runs, that was most of the traffic that put them over the
// line, spent re-fetching answers we already had.
//
// So the cache goes to disk, next to the stream and host caches.

const (
	searchCacheFileName = "superflix-search-cache.json"

	// searchCacheTTL is how long an answer is reused.
	//
	// A catalogue entry does not appear and vanish within the hour, and the
	// cost of being slightly stale is one title missing from a search that the
	// user can repeat later. The cost of NOT caching is measured above.
	searchCacheTTL = 45 * time.Minute

	// searchCacheMaxEntries bounds the file. Oldest go first — the point is to
	// absorb a session's repeats, not to remember a year of them.
	searchCacheMaxEntries = 300
)

// searchCacheEntry is one query's answer and when it was taken.
type searchCacheEntry struct {
	At      int64             `json:"at"`
	Results []*SuperFlixMedia `json:"results"`
}

func (e searchCacheEntry) fresh() bool {
	return time.Since(time.Unix(e.At, 0)) <= searchCacheTTL
}

// searchCache is the on-disk query→results map.
type searchCache struct {
	mu      sync.Mutex
	path    string
	entries map[string]searchCacheEntry
	loaded  bool
}

var defaultSearchCache = &searchCache{}

// file returns the backing path, or "" when there is nowhere to write.
//
// Test binaries get "" so a test that happens to search cannot seed the
// developer's real cache — and cannot read an answer from it either, which
// would make a parser test pass on yesterday's data. Same rule the back-off
// gate uses.
func (sc *searchCache) file() string {
	if sc.path != "" {
		return sc.path
	}
	if testing.Testing() {
		return ""
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	d := filepath.Join(dir, "goanime")
	if mkErr := os.MkdirAll(d, 0o700); mkErr != nil {
		return ""
	}
	sc.path = filepath.Join(d, searchCacheFileName)
	return sc.path
}

func (sc *searchCache) ensureLoadedLocked() {
	if sc.loaded {
		return
	}
	sc.loaded = true
	sc.entries = map[string]searchCacheEntry{}

	path := sc.file()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is built from the user cache dir
	if err != nil {
		return
	}
	_ = jsonx.Unmarshal(data, &sc.entries)
}

// get returns a cached answer for key when one is still fresh.
func (sc *searchCache) get(key string) ([]*SuperFlixMedia, bool) {
	if sc == nil || key == "" {
		return nil, false
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	// No file means no cache at all, not an in-memory one. This layer exists
	// only to survive the process; the client keeps its own map for repeats
	// inside one run. Holding entries here with nowhere to write them would
	// make a package-level map shared by every test in the binary — which is
	// exactly what it did, and four tests that count requests started seeing
	// each other's answers.
	if sc.file() == "" {
		return nil, false
	}
	sc.ensureLoadedLocked()

	e, ok := sc.entries[key]
	if !ok || !e.fresh() {
		return nil, false
	}
	return e.Results, true
}

// put records an answer and writes the file.
//
// An empty result is cached too, on purpose: "this host does not have it" is an
// answer, and re-asking for a title the catalogue does not carry is exactly the
// repeat that costs an allowance nobody has to spare.
func (sc *searchCache) put(key string, results []*SuperFlixMedia) {
	if sc == nil || key == "" {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	path := sc.file()
	if path == "" {
		return
	}
	sc.ensureLoadedLocked()

	sc.entries[key] = searchCacheEntry{At: time.Now().Unix(), Results: results}
	sc.evictLocked()

	blob, err := jsonx.Marshal(sc.entries)
	if err != nil {
		return
	}
	if writeErr := os.WriteFile(path, blob, 0o600); writeErr != nil {
		util.Debug("SuperFlix: could not persist the search cache", "err", writeErr)
	}
}

// evictLocked drops stale entries, then the oldest, until the map fits.
func (sc *searchCache) evictLocked() {
	for k, e := range sc.entries {
		if !e.fresh() {
			delete(sc.entries, k)
		}
	}
	if len(sc.entries) <= searchCacheMaxEntries {
		return
	}
	keys := make([]string, 0, len(sc.entries))
	for k := range sc.entries {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return sc.entries[keys[i]].At < sc.entries[keys[j]].At
	})
	for _, k := range keys[:len(sc.entries)-searchCacheMaxEntries] {
		delete(sc.entries, k)
	}
}
