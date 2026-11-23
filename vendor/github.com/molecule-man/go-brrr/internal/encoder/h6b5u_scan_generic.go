//go:build !amd64 || purego

package encoder

import (
	"math/bits"
	"unsafe"
)

// h6b5uFindInBucket scans the bucket of key from newest to oldest, then stores cur.
func h6b5uFindInBucket(h *h6b5u, data unsafe.Pointer, key, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	nextKey uint, out *hasherSearchResult,
) {
	// This target has no prefetch. Load from the next bucket to warm the cache.
	nextKey &= h6b5BucketSize - 1
	nextBucket := h.bucketAt(uint32(nextKey))
	h.nextBucket = nextBucket[0]
	if nextN := h.num[nextKey]; nextN > 0 {
		h.nextBucket = uint32(*(*byte)(unsafe.Add(data, uint(nextBucket[(nextN-1)&h6b5BlockMask])&mask)))
	}

	load32 := func(i uint) uint32 { return loadU32LEPtr(unsafe.Add(data, i)) }
	key &= h6b5BucketSize - 1
	bucket := h.bucketAt(uint32(key))
	n := uint(h.num[key])

	down := uint(0)
	if n > h6b5BlockSize {
		down = n - h6b5BlockSize
	}
	score := bestScore
	if curMasked+bestLen > mask {
		down = n
	}
	first4 := load32(curMasked)
	curProbe := load32(curMasked + bestLen - 3)
	for i := n; i > down; {
		i--
		prevRaw := uint(bucket[i&h6b5BlockMask])
		if prevRaw < minPrev {
			break
		}
		prevMasked := prevRaw & mask
		if curProbe != load32(prevMasked+bestLen-3) || prevMasked+bestLen > mask {
			continue
		}
		if first4 != load32(prevMasked) {
			continue
		}
		// Keep this loop here: a helper exceeds the inline budget.
		ml, limit := uint(4), maxLength
		for ml+8 <= limit {
			xor := loadU64LEPtr(unsafe.Add(data, prevMasked+ml)) ^ loadU64LEPtr(unsafe.Add(data, curMasked+ml))
			if xor != 0 {
				ml += uint(bits.TrailingZeros64(xor) / 8)
				break
			}
			ml += 8
		}
		for ml < limit && *(*byte)(unsafe.Add(data, prevMasked+ml)) == *(*byte)(unsafe.Add(data, curMasked+ml)) {
			ml++
		}
		// backward < window, so the mask gives cur-prevRaw.
		backward := (curMasked - prevMasked) & mask
		if s := backwardReferenceScore(ml, backward); s > score {
			score, bestLen = s, ml
			out.len, out.distance, out.score = ml, backward, s
			if curMasked+bestLen > mask {
				break
			}
			curProbe = load32(curMasked + bestLen - 3)
		}
	}

	// Store cur after the scan so its distance cannot be zero.
	bucket[n&h6b5BlockMask] = uint32(cur)
	h.num[key]++
}
