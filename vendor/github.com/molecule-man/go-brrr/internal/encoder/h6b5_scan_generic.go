//go:build !amd64 || purego

package encoder

import (
	"math/bits"
	"unsafe"
)

// h6b5FindInBucket scans matching tags from newest to oldest, then stores cur.
func h6b5FindInBucket(h *h6b5, data unsafe.Pointer, keyTag, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	next *h6b5Block, out *hasherSearchResult,
) {
	// This target has no prefetch. Load the next block to warm the cache.
	h.nextWarm = uint32(next.tags[0]) + next.pos[16]

	load32 := func(i uint) uint32 { return loadU32LEPtr(unsafe.Add(data, i)) }
	key, tag := keyTag>>8&(h6b5BucketSize-1), uint8(keyTag)
	block := &h.blocks[key]
	n := uint(h.num[key])

	// Bit j of matches: slot (head+j) has the same tag. Bit 0 is the newest entry.
	head := (n + 1) & h6b5BlockMask
	matches := bits.RotateLeft32(tagMask32(&block.tags, tag), -int(head))
	// Mask the empty slots.
	if stored := 0xFFFF - n; stored < h6b5BlockSize {
		matches &= uint32(1)<<stored - 1
	}

	score := bestScore
	if curMasked+bestLen > mask {
		matches = 0
	}
	first4 := load32(curMasked)
	curProbe := load32(curMasked + bestLen - 3)
	for ; matches != 0; matches &= matches - 1 {
		prevRaw := uint(block.pos[(head+uint(bits.TrailingZeros32(matches)))&h6b5BlockMask])
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
	slot := n & h6b5BlockMask
	block.pos[slot] = uint32(cur)
	block.tags[slot] = tag
	h.num[key]--
}
