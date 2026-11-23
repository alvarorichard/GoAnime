//go:build amd64 && !purego

package encoder

import "unsafe"

// h6FindInBucket scans matching tags newest first, stores cur, and prefetches next.
// It updates out only when the score exceeds bestScore.
// keyTag packs the bucket key above the tag. Before the first wrap, mask is ^uint(0).
//
//go:noescape
func h6FindInBucket(h *h6, data unsafe.Pointer, keyTag, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	next *h6Block, out *hasherSearchResult)
