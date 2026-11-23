// Hasher pools and the type-dispatching releaseHasher used by the streaming
// encoder. Pools live next to the hashers so encoder.go stays focused on
// encoder logic.

package encoder

import "sync"

// Hasher pools keyed by concrete type. Using sync.Pool avoids the expensive
// zero-initialization of large bucket arrays (up to 32 MB for h5b8/h6b8) on
// every oneshot compression. The hasher's reset() method only zeroes the small
// num[] counter array, not the full bucket storage.
var (
	poolH2      = sync.Pool{New: func() any { return new(h2) }}
	poolH2u16   = sync.Pool{New: func() any { return new(h2u16) }}
	poolH3      = sync.Pool{New: func() any { return new(h3) }}
	poolH3u16   = sync.Pool{New: func() any { return new(h3u16) }}
	poolH4      = sync.Pool{New: func() any { return new(h4) }}
	poolH4u16   = sync.Pool{New: func() any { return new(h4u16) }}
	poolH5      = sync.Pool{New: func() any { return new(h5) }}
	poolH54     = sync.Pool{New: func() any { return new(h54) }}
	poolH5b5    = sync.Pool{New: func() any { return new(h5b5) }}
	poolH5b6    = sync.Pool{New: func() any { return new(h5b6) }}
	poolH5b6u16 = sync.Pool{New: func() any { return new(h5b6u16) }}
	poolH5b7    = sync.Pool{New: func() any { return new(h5b7) }}
	poolH5b8    = sync.Pool{New: func() any { return new(h5b8) }}
	poolH6      = sync.Pool{New: func() any { return new(h6) }}
	poolH6u     = sync.Pool{New: func() any { return new(h6u) }}
	poolH6b5    = sync.Pool{New: func() any { return new(h6b5) }}
	poolH6b5u   = sync.Pool{New: func() any { return new(h6b5u) }}
	poolH6b6    = sync.Pool{New: func() any { return new(h6b6) }}
	poolH6b7    = sync.Pool{New: func() any { return new(h6b7) }}
	poolH6b8    = sync.Pool{New: func() any { return new(h6b8) }}
	poolH10     = sync.Pool{New: func() any { return new(h10) }}
)

// releaseHasher returns a hasher to its type-specific pool. Hashers without a
// pool (h40, h41, h42) are dropped on the floor — they're allocated
// directly and Go's GC reclaims them.
func releaseHasher(h streamHasher) {
	switch h := h.(type) {
	case *h2:
		poolH2.Put(h)
	case *h2u16:
		poolH2u16.Put(h)
	case *h3:
		poolH3.Put(h)
	case *h3u16:
		poolH3u16.Put(h)
	case *h4:
		poolH4.Put(h)
	case *h4u16:
		poolH4u16.Put(h)
	case *h5:
		poolH5.Put(h)
	case *h54:
		poolH54.Put(h)
	case *h5b5:
		poolH5b5.Put(h)
	case *h5b6:
		poolH5b6.Put(h)
	case *h5b6u16:
		poolH5b6u16.Put(h)
	case *h5b7:
		poolH5b7.Put(h)
	case *h5b8:
		poolH5b8.Put(h)
	case *h6:
		poolH6.Put(h)
	case *h6u:
		poolH6u.Put(h)
	case *h6b5:
		poolH6b5.Put(h)
	case *h6b5u:
		poolH6b5u.Put(h)
	case *h6b6:
		poolH6b6.Put(h)
	case *h6b7:
		poolH6b7.Put(h)
	case *h6b8:
		poolH6b8.Put(h)
	case *h10:
		h.bufs = nil
		poolH10.Put(h)
	}
}
