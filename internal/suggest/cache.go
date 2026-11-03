package suggest

import (
	"container/list"
	"slices"
	"sync"
	"time"
)

// cache is a small LRU with a TTL; it holds provider answers by query.
type cache struct {
	mu    sync.Mutex
	max   int
	ttl   time.Duration
	now   func() time.Time
	order *list.List // front is the most recently used
	byKey map[string]*list.Element
}

type cacheEntry struct {
	key     string
	items   []Suggestion
	expires time.Time
}

func newCache(size int, ttl time.Duration, now func() time.Time) *cache {
	return &cache{max: size, ttl: ttl, now: now, order: list.New(), byKey: make(map[string]*list.Element)}
}

// get returns a copy of the live entry for key and marks it recently used; an
// expired entry is dropped.
func (c *cache) get(key string) ([]Suggestion, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.byKey[key]
	if !ok {
		return nil, false
	}
	entry := el.Value.(*cacheEntry)
	if !c.now().Before(entry.expires) {
		c.order.Remove(el)
		delete(c.byKey, key)
		return nil, false
	}
	c.order.MoveToFront(el)
	return slices.Clone(entry.items), true
}

// put stores a copy of items under key, evicting the least recently used
// entry when the cache is full.
func (c *cache) put(key string, items []Suggestion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expires := c.now().Add(c.ttl)
	if el, ok := c.byKey[key]; ok {
		entry := el.Value.(*cacheEntry)
		entry.items, entry.expires = slices.Clone(items), expires
		c.order.MoveToFront(el)
		return
	}
	c.byKey[key] = c.order.PushFront(&cacheEntry{key: key, items: slices.Clone(items), expires: expires})
	if c.order.Len() > c.max {
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.byKey, oldest.Value.(*cacheEntry).key)
	}
}

// len is the number of cached entries.
func (c *cache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}
