// Package offlinecache remembers titles a source still lists but that the user
// has already watched fail with no copy anywhere else, so later searches can
// hide them.
//
// Nothing here probes the network: a title is recorded only after playback
// failed with "video removed" and the fallback found the episode on no other
// source. Entries expire, because a host may restore deleted files. The store
// lives in the user's cache directory and is best-effort: any I/O problem
// degrades to an in-memory store and never fails a search or a playback.
package offlinecache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// TTL is how long a recorded title stays hidden.
	TTL = 6 * time.Hour
	// EnvHide switches hiding off when set to 0, false, no or off. Titles are
	// then still listed (and still recorded, so turning it back on works).
	EnvHide = "GOANIME_HIDE_OFFLINE"

	fileName = "offline_titles.json"
)

// Cache is a TTL map of (source, title URL) pairs. It is safe for concurrent use.
type Cache struct {
	mu      sync.Mutex
	path    string // empty means memory-only
	ttl     time.Duration
	now     func() time.Time
	entries map[string]int64 // key -> expiry, unix seconds
}

// New returns a store at path (memory-only when path is empty) whose entries
// stay valid for ttl. A ttl <= 0 makes every entry expire at once.
func New(path string, ttl time.Duration) *Cache {
	return &Cache{path: path, ttl: ttl, now: time.Now, entries: map[string]int64{}}
}

// Default builds the store the app uses: <user cache dir>/goanime/offline_titles.json.
func Default() *Cache {
	path := ""
	if dir, err := os.UserCacheDir(); err == nil {
		path = filepath.Join(dir, "goanime", fileName)
	}
	return New(path, TTL)
}

var (
	sharedOnce sync.Once
	shared     *Cache
)

// Shared is the app-wide store: playback records titles in it and search reads
// it. It is built on first use.
func Shared() *Cache {
	sharedOnce.Do(func() { shared = Default() })
	return shared
}

// HidingEnabled reports whether searches should hide recorded titles.
func HidingEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvHide))) {
	case "0", "false", "no", "off":
		return false
	}
	return true
}

func key(source, titleURL string) string {
	return strings.ToLower(source) + "|" + titleURL
}

// Set returns the recorded titles of source among urls, reading the file once.
func (c *Cache) Set(source string, urls []string) map[string]bool {
	out := map[string]bool{}
	if c == nil {
		return out
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mergeDiskLocked()
	now := c.now().Unix()
	for _, u := range urls {
		if exp, ok := c.entries[key(source, u)]; ok && now < exp {
			out[u] = true
		}
	}
	return out
}

// MarkOffline records the title and persists the store. The returned error is
// informational: the record is kept in memory either way.
func (c *Cache) MarkOffline(source, titleURL string) error {
	if c == nil || titleURL == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mergeDiskLocked()
	c.entries[key(source, titleURL)] = c.now().Add(c.ttl).Unix()
	return c.saveLocked()
}

// mergeDiskLocked folds the file's live entries into memory, keeping the later
// expiry, so two processes do not erase each other's records.
func (c *Cache) mergeDiskLocked() {
	if c.path == "" {
		return
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	var onDisk map[string]int64
	if err := json.Unmarshal(data, &onDisk); err != nil {
		return // corrupt: the next save replaces it
	}
	for k, exp := range onDisk {
		if exp > c.entries[k] {
			c.entries[k] = exp
		}
	}
}

func (c *Cache) saveLocked() error {
	if c.path == "" {
		return nil
	}
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
	// Temp file + rename, so a crash never leaves a half-written store.
	tmp, err := os.CreateTemp(filepath.Dir(c.path), fileName+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	fail := func(err error) error {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		return fail(err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fail(err)
	}
	if err := os.Rename(tmpName, c.path); err != nil {
		return fail(err)
	}
	return nil
}
