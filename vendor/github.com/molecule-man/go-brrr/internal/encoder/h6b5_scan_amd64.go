//go:build amd64 && !purego

package encoder

import "unsafe"

// h6b5FindInBucket scans matching tags from newest to oldest, then stores cur.
// It prefetches next.
// It updates out only when the score exceeds bestScore.
// keyTag packs the bucket key above the tag. Before the first wrap, mask is ^uint(0).
//
//go:noescape
func h6b5FindInBucket(h *h6b5, data unsafe.Pointer, keyTag, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	next *h6b5Block, out *hasherSearchResult)
