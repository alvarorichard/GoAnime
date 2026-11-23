//go:build amd64 && !purego

package encoder

import "unsafe"

// h6b5uFindInBucket scans the bucket of key from newest to oldest, then stores cur.
// It prefetches the bucket of nextKey and data at slot (num[nextKey]-1)&31.
// It updates out only when the score exceeds bestScore.
// Before the first wrap, mask is ^uint(0).
//
//go:noescape
func h6b5uFindInBucket(h *h6b5u, data unsafe.Pointer, key, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	nextKey uint, out *hasherSearchResult)
