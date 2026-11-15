// Package offlinecache remembers which titles a source lists but no longer
// hosts, so a search can drop them without probing the network again.
//
// Entries expire: a source that deleted a file may restore it, so a verdict is
// only trusted for a limited time. The cache lives in the user's cache
// directory and is best-effort throughout: any I/O problem degrades to an
// in-memory cache and never fails a search.
package offlinecache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultTTL is how long an "offline" verdict is trusted.
	DefaultTTL = 6 * time.Hour
	// EnvTTL overrides DefaultTTL with a Go duration such as "30m" or "24h".
	// A zero or negative value disables the cache.
	EnvTTL = "GOANIME_OFFLINE_CACHE_TTL"

	fileName = "offline_titles.json"
)

// Cache is a TTL map of (source, title URL) pairs known to be offline.
// It is safe for concurrent use.
type Cache struct {
	mu      sync.Mutex
	path    string // empty means memory-only
	ttl     time.Duration
	now     func() time.Time
	entries map[string]int64 // key -> expiry, unix seconds
	loaded  bool
}

// New returns a cache stored at path (memory-only when path is empty) whose
// entries stay valid for ttl. A ttl <= 0 disables the cache entirely.
func New(path string, ttl time.Duration) *Cache {
	return &Cache{path: path, ttl: ttl, now: time.Now, entries: map[string]int64{}}
}

// Default builds the cache the app uses: <user cache dir>/goanime/offline_titles.json,
// with the TTL taken from GOANIME_OFFLINE_CACHE_TTL when set and valid.
func Default() *Cache {
	ttl := DefaultTTL
	if v := os.Getenv(EnvTTL); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			ttl = d
		}
	}
	path := ""
	if dir, err := os.UserCacheDir(); err == nil {
		path = filepath.Join(dir, "goanime", fileName)
	}
	return New(path, ttl)
}

func key(source, titleURL string) string { return source + "|" + titleURL }

// Enabled reports whether the cache records and serves verdicts.
func (c *Cache) Enabled() bool { return c != nil && c.ttl > 0 }

// IsOffline reports whether the title was recorded offline and the verdict has
// not expired yet.
func (c *Cache) IsOffline(source, titleURL string) bool {
	if !c.Enabled() {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	exp, ok := c.entries[key(source, titleURL)]
	if !ok {
		return false
	}
	if c.now().Unix() >= exp {
		delete(c.entries, key(source, titleURL))
		return false
	}
	return true
}

// MarkOffline records the title as offline for the cache's TTL. Call Save to
// persist a batch of marks.
func (c *Cache) MarkOffline(source, titleURL string) {
	if !c.Enabled() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	c.entries[key(source, titleURL)] = c.now().Add(c.ttl).Unix()
}

// Save writes the cache to disk, dropping expired entries first. It is a no-op
// for a memory-only or disabled cache, and errors are returned for callers that
// care but are safe to ignore.
func (c *Cache) Save() error {
	if !c.Enabled() || c.path == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()

	nowUnix := c.now().Unix()
	for k, exp := range c.entries {
		if nowUnix >= exp {
			delete(c.entries, k)
		}
	}

	data, err := json.Marshal(c.entries)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	// Write to a temp file and rename so a crash never leaves a half-written cache.
	tmp, err := os.CreateTemp(filepath.Dir(c.path), fileName+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, c.path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// loadLocked reads the file once. A missing or corrupt file means an empty
// cache; the next Save replaces it.
func (c *Cache) loadLocked() {
	if c.loaded {
		return
	}
	c.loaded = true
	if c.path == "" {
		return
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	var onDisk map[string]int64
	if err := json.Unmarshal(data, &onDisk); err != nil {
		return
	}
	for k, exp := range onDisk {
		c.entries[k] = exp
	}
}
