//go:build !amd64 || purego

package encoder

import (
	"math/bits"
	"unsafe"
)

// h6FindInBucket scans matching tags newest first, then stores cur.
func h6FindInBucket(h *h6, data unsafe.Pointer, keyTag, cur, curMasked,
	mask, minPrev, maxLength, bestLen, bestScore uint,
	_ *h6Block, out *hasherSearchResult,
) {
	load32 := func(i uint) uint32 { return loadU32LEPtr(unsafe.Add(data, i)) }
	key, tag := keyTag>>8&(h6BucketSize-1), uint8(keyTag)
	block := &h.blocks[key]
	n := uint(h.num[key])

	// Bit j of matches: slot (head+j) has the same tag. Bit 0 is the newest entry.
	head := (n + 1) & h6BlockMask
	splat := uint64(tag) * 0x0101010101010101
	matches := uint32(bits.RotateLeft16(uint16(tagEqualMask8(loadU64LE(block.tags[:], 0)^splat)|
		tagEqualMask8(loadU64LE(block.tags[:], 8)^splat)<<8), -int(head)))
	// Mask the empty slots.
	if stored := 0xFFFF - n; stored < h6BlockSize {
		matches &= uint32(1)<<stored - 1
	}

	score := bestScore
	if curMasked+bestLen > mask {
		matches = 0
	}
	first4 := load32(curMasked)
	curProbe := load32(curMasked + bestLen - 3)
	for ; matches != 0; matches &= matches - 1 {
		prevRaw := uint(block.pos[(head+uint(bits.TrailingZeros32(matches)))&h6BlockMask])
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
	slot := n & h6BlockMask
	block.pos[slot] = uint32(cur)
	block.tags[slot] = tag
	h.num[key]--
}

// loadU32LEPtr reads unaligned bytes in little-endian order.
func loadU32LEPtr(p unsafe.Pointer) uint32 {
	b := (*[4]byte)(p)
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// loadU64LEPtr reads unaligned bytes in little-endian order.
func loadU64LEPtr(p unsafe.Pointer) uint64 {
	b := (*[8]byte)(p)
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}
