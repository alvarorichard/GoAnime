package tlscache

import (
	"container/list"
	"crypto/tls"
	"slices"
	"sync"
	"time"

	"github.com/enetx/surf/internal/tlsstate"
)

// Cache is a concurrency-safe ticket queue. Get consumes TLS 1.3 tickets;
// Put(key, nil) removes all remembered state for that key. A recently consumed
// ticket cannot be resurrected by a TLS early-data rejection update.
type Cache struct {
	mu       sync.Mutex
	policy   Builder
	keys     map[string]*list.Element
	lru      *list.List
	bytes    int
	sequence uint64
}

type entry struct {
	key, group string
	queue      []ticket
	consumed   [][32]byte
}

type ticket struct {
	state                  *tls.ClientSessionState
	version                uint16
	from, until            time.Time
	certificateUntil       time.Time // Separate safety pruning from deferred session validation.
	rank                   time.Time // Session deadline, independent of certificate and hint expiry.
	offerFrom, offerUntil  time.Time // Additional engine/NSS wire-offer bounds.
	deferSessionValidation bool      // Preserve live insertion ordering until lookup.
	id                     [32]byte
	encoded                []byte // Replaces state when compressed storage is enabled.
	cost                   int    // Charged payload and per-ticket key bytes.
	sequence               uint64 // Deterministic tie-break for equal global expirations.
}

func newCache(policy Builder) *Cache {
	return &Cache{policy: policy, keys: make(map[string]*list.Element), lru: list.New()}
}

// Put stores a valid ticket, replacing incompatible queued state when Group
// changes. Nil removes the key and its consumption history. Recently consumed
// identities are ignored, including TLS early-data rejection updates.
func (c *Cache) Put(key string, state *tls.ClientSessionState) {
	if state == nil {
		c.mu.Lock()
		if element := c.keys[key]; element != nil {
			c.remove(element)
		}
		c.mu.Unlock()
		return
	}

	info, ok := tlsstate.SessionInfo(state)
	if !ok {
		return
	}
	now := c.policy.clock()
	record, ok := c.validity(info, now)
	if !ok {
		return
	}

	group := ""
	if c.policy.group != nil {
		group = c.policy.group(state)
	}

	encoded, size, err := encodeRecord(state, c.policy.brotli)
	if err != nil {
		return
	}

	stored := state
	if encoded != nil {
		stored = nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	element := c.keys[key]
	if element == nil {
		if len(c.keys) >= c.policy.capacity {
			c.remove(c.lru.Back())
		}
		element = c.lru.PushFront(&entry{key: key, group: group})
		c.keys[key] = element
	}
	c.lru.MoveToFront(element)

	e := element.Value.(*entry)
	if slices.Contains(e.consumed, record.id) {
		return
	}

	if e.group != group {
		c.clearQueue(e)
		e.group = group
	}
	for i := len(e.queue) - 1; i >= 0; i-- {
		if t := e.queue[i]; t.id == record.id || !now.Before(t.certificateUntil) ||
			!t.deferSessionValidation && !now.Before(t.until) {
			c.removeTicket(e, i)
		}
	}
	if c.policy.reuseTLS12 && len(e.queue) > 0 && e.queue[0].version < tls.VersionTLS13 {
		c.removeTicket(e, 0)
	}
	if len(e.queue) == c.policy.tickets {
		if c.policy.order == EarliestExpiryFirst {
			c.removeTicket(e, 0)
		} else {
			c.removeTicket(e, len(e.queue)-1)
		}
	}

	c.sequence++
	record.state, record.encoded, record.cost = stored, encoded, len(key)+size
	record.sequence = c.sequence
	e.queue = append([]ticket{record}, e.queue...)
	c.bytes += record.cost
	if c.policy.order == EarliestExpiryFirst {
		// Stable insertion leaves a newly received equal-expiry token first,
		// matching Firefox TokenCacheEntry::AddRecord.
		slices.SortStableFunc(e.queue, func(a, b ticket) int { return a.rank.Compare(b.rank) })
	}

	c.evict()
}

// Bytes reports charged queued-record storage, including the key for each
// ticket. It excludes object overhead and bounded consumed-ticket history;
// it is a storage budget, not a measurement of the Go process heap.
func (c *Cache) Bytes() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.bytes
}

// Get selects the next usable ticket according to Order and the expiry policy.
// TLS 1.3 tickets are consumed once; ReuseTLS12 controls legacy ticket reuse.
func (c *Cache) Get(key string) (*tls.ClientSessionState, bool) {
	return c.lookup(key, true)
}

// Peek inspects the next valid ticket without consuming it. The returned state
// is read-only; callers must not use it to bypass single-use TLS selection.
func (c *Cache) Peek(key string) (*tls.ClientSessionState, bool) {
	return c.lookup(key, false)
}

func (c *Cache) lookup(key string, consume bool) (*tls.ClientSessionState, bool) {
	now := c.policy.clock()
	c.mu.Lock()
	defer c.mu.Unlock()
	element := c.keys[key]
	if element == nil {
		return nil, false
	}
	c.lru.MoveToFront(element)
	e := element.Value.(*entry)
	for len(e.queue) > 0 {
		t := e.queue[0]
		if now.Before(t.from) || !now.Before(t.until) {
			if c.policy.discardOnExpiry {
				c.remove(element)
				return nil, false
			}
			c.removeTicket(e, 0)
			if !c.policy.skipExpired {
				return nil, false
			}
			continue
		}
		if now.Before(t.offerFrom) {
			// BoringSSL can reject a future rebased session while Chrome's
			// one-second cache tolerance still retains it for a later lookup.
			return nil, false
		}

		if !t.offerUntil.IsZero() && !now.Before(t.offerUntil) {
			c.removeTicket(e, 0)
			if !c.policy.skipExpired {
				return nil, false
			}
			continue
		}

		state, err := decodeRecord(t)
		if err != nil {
			c.removeTicket(e, 0)
			continue
		}
		if consume && (t.version >= tls.VersionTLS13 || !c.policy.reuseTLS12) {
			c.removeTicket(e, 0)
			e.consumed = append(e.consumed, t.id)
			if len(e.consumed) > 2*c.policy.tickets {
				e.consumed = e.consumed[1:]
			}
		}
		return state, true
	}
	return nil, false
}

// Clear releases all queued and consumed state.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	clear(c.keys)
	c.lru.Init()
	c.bytes = 0
}

// RejectEarlyData preserves queued TLS resumption while disabling their early
// data, when the configured policy enables that behavior. Selected tickets are
// already consumed and are not inserted again.
func (c *Cache) RejectEarlyData(key string) {
	if !c.policy.clearEarlyData {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	element := c.keys[key]
	if element == nil {
		return
	}

	e := element.Value.(*entry)
	kept := e.queue[:0]
	for _, t := range e.queue {
		c.bytes -= t.cost
		original, err := decodeRecord(t)
		if err != nil {
			continue
		}
		label, state, err := original.ResumptionState()
		if err != nil {
			continue
		}
		data, err := state.Bytes()
		if err != nil {
			continue
		}
		clone, err := tls.ParseSessionState(data)
		if err != nil {
			continue
		}
		clone.EarlyData = false
		updated, err := tls.NewResumptionState(label, clone)
		if err != nil {
			continue
		}
		t.state = updated
		t.encoded, t.cost, err = encodeRecord(updated, c.policy.brotli)
		if err != nil {
			continue
		}
		if t.encoded != nil {
			t.state = nil
		}
		t.cost += len(key)
		c.bytes += t.cost
		kept = append(kept, t)
	}
	clear(e.queue[len(kept):])
	e.queue = kept
	c.evict()
}

func (c *Cache) remove(element *list.Element) {
	e := element.Value.(*entry)
	c.clearQueue(e)
	delete(c.keys, e.key)
	c.lru.Remove(element)
}

func (c *Cache) clearQueue(e *entry) {
	for _, t := range e.queue {
		c.bytes -= t.cost
	}
	e.queue = nil
}

func (c *Cache) removeTicket(e *entry, index int) {
	c.bytes -= e.queue[index].cost
	e.queue = slices.Delete(e.queue, index, index+1)
}

func (c *Cache) evict() {
	for c.policy.byteCapacity > 0 && c.bytes > c.policy.byteCapacity {
		if c.policy.eviction == EvictLeastRecentlyUsed {
			c.remove(c.lru.Back())
			continue
		}

		var selected *entry
		var index int
		for _, element := range c.keys {
			e := element.Value.(*entry)
			for i, t := range e.queue {
				if selected == nil || t.rank.Before(selected.queue[index].rank) ||
					t.rank.Equal(selected.queue[index].rank) && t.sequence < selected.queue[index].sequence {
					selected, index = e, i
				}
			}
		}

		// NSS does not specify global ordering for equal-expiry tickets.
		// Sequence provides a deterministic local tie-break.
		c.removeTicket(selected, index)
	}
}
