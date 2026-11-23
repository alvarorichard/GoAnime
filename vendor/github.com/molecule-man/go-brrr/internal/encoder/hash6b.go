// H6 hasher family for qualities 7-8 on large inputs with large windows.
//
// The bucket array type sets the search depth. One instantiation per depth.
// The depth, the shift and the mask come from the array length, so each
// instantiation compiles them as constants. The bucket type has no methods,
// so no hot-path call needs a generic dictionary.
//
// Selected when quality is 7 or 8, sizeHint >= 1MiB, and lgwin >= 19.

package encoder

import (
	"math/bits"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

const (
	h6bBucketBits = 15
	h6bBucketSize = 1 << h6bBucketBits // 32768
	h6bHashShift  = 64 - h6bBucketBits // 49

	// h6bHashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h6bHashTypeLength = 8

	// h6bNumLastDistances is the number of distance cache entries to check.
	// For quality 7-8, the C reference uses 10.
	h6bNumLastDistances = 10

	// Search depths, in positions per bucket.
	h6b6BlockSize = 64
	h6b7BlockSize = 128
)

// h6bHashMul is the hash multiplier: kHashMul64 << (64 - 5*8).
// Pre-computed because the untyped shift overflows Go constant arithmetic.
const h6bHashMul uint64 = 0x7BD3579BD3000000

// h6bBlock constrains a bucket to one of the supported search depths.
type h6bBlock interface {
	~[h6b6BlockSize]uint32 | ~[h6b7BlockSize]uint32
}

// bucketRing is the position ring buffer of one bucket. Access is unchecked.
// Callers mask the index with the depth, so the index stays in range.
type bucketRing struct{ base unsafe.Pointer }

// h6b is a forgetful hash table. Each of the 32K buckets holds one ring
// buffer B of positions. The length of B is the search depth.
//
//nolint:govet // fieldalignment counts B as pointer data; every B is [N]uint32
type h6b[B h6bBlock] struct {
	num        [h6bBucketSize]uint16 // entry count per bucket
	buckets    [h6bBucketSize]B      // position ring buffers
	nextBucket uint32                // speculative load to warm cache
	// everWrapped disables the direct position path after the first ring wrap.
	everWrapped bool
	hasherCommon
}

// h6b6 searches 64 positions per bucket (quality 7).
type h6b6 = h6b[[h6b6BlockSize]uint32]

// h6b7 searches 128 positions per bucket (quality 8).
type h6b7 = h6b[[h6b7BlockSize]uint32]

func (b bucketRing) at(i uint) uint32       { return *(*uint32)(unsafe.Add(b.base, i<<2)) }
func (b bucketRing) put(i uint, pos uint32) { *(*uint32)(unsafe.Add(b.base, i<<2)) = pos }

func (h *h6b[B]) common() *hasherCommon { return &h.hasherCommon }

// bucketRingAt returns the ring buffer for key.
//
// h6bHash shifts its product right by h6bHashShift, so key < h6bBucketSize
// and the whole ring lies inside buckets.
func bucketRingAt(buckets unsafe.Pointer, key uint32, shift uint) bucketRing {
	return bucketRing{unsafe.Add(buckets, uintptr(key)<<(shift+2))}
}

// h6bHash computes a 15-bit bucket index from 8 bytes at data[i:i+8].
func h6bHash(data []byte, i uint) uint32 {
	return uint32((loadU64LE(data, i) * h6bHashMul) >> h6bHashShift)
}

// reset zeroes the entry counts before use.
// When oneShot is true and the input is small, only the touched buckets
// are cleared (partial prepare). Otherwise the full count array is zeroed.
func (h *h6b[B]) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h6bBucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h6bHash(data, i)
			h.num[key] = 0
		}
	} else {
		h.num = [h6bBucketSize]uint16{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the ring buffer for the 8-byte sequence at
// data[pos & mask].
func (h *h6b[B]) store(data []byte, mask, pos uint) {
	blockSize := uint(unsafe.Sizeof(h.buckets)) / h6bBucketSize / 4
	key := h6bHash(data, pos&mask)
	offset := uint(h.num[key])&(blockSize-1) + uint(key)*blockSize
	h.num[key]++
	*(*uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), offset<<2)) = uint32(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h6b[B]) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.store(data, mask, i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h6b[B]) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h6bHashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch searches for the best backward reference at position cur
// in the ring buffer, then stores cur in the hash table.
//
// The search has three phases:
//  1. Distance cache: try the last 10 cached distances (4 base entries plus
//     6 derived near-miss entries for dist[0]). Accept length >= 3, or
//     length == 2 for the first two cache entries.
//  2. Hash bucket scan: walk the ring buffer of the bucket. Reject candidates
//     with a 4-byte quick comparison, accept length >= 4.
//  3. Static dictionary fallback: when neither phase produced a match,
//     search the static dictionary with deep search.
func (h *h6b[B]) findLongestMatch(
	data []byte, ringBufferMask uint,
	distCache *[16]int,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	if ringBufferMask >= uint(len(data)) {
		h.findLongestMatchSmallBuf(data, ringBufferMask, distCache,
			cur, maxLength, maxBackward, dictDistance,
			dictNumLookups, dictNumMatches, out)
		return
	}

	// --- fast path: ringBufferMask < len(data) ---
	_ = data[ringBufferMask]

	// Depth, shift and mask fold to constants in each instantiation.
	// Do not move this into a helper. A generic method that calls a generic
	// function loads a sub-dictionary on every call, also when the callee
	// inlines.
	blockSize := uint(unsafe.Sizeof(h.buckets)) / h6bBucketSize / 4
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h6bHash(data, curMasked)

	nextKey := h6bHash(data, (cur+1)&ringBufferMask)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	if hasPrefetch {
		prefetch2(unsafe.Add(nextBucket.base, (uint(nextN-1)&blockMask)<<2),
			unsafe.Add(nextBucket.base, (uint(nextN-17)&blockMask)<<2))
	} else {
		h.nextBucket = nextBucket.at(0)
		if nextN > 0 {
			p := uint(nextBucket.at(uint((nextN-1)&uint16(blockMask)))) & ringBufferMask
			h.nextBucket = uint32(data[p])
		}
	}

	out.len = 0
	out.lenCodeDelta = 0

	lim := ringLimit(ringBufferMask, curMasked, bestLen)

	// Phase 1: try cached distances.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	curByte := loadByte(data, curMasked+bestLen)
	backward := uint(distCache[0])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[1])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						lim = ringLimit(ringBufferMask, curMasked, bestLen)
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = uint(distCache[2])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+43 < score {
					bestScore = score - 43
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[3])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+43 < score {
					bestScore = score - 43
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[4])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+39 < score {
					bestScore = score - 39
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[5])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+39 < score {
					bestScore = score - 39
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[6])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+47 < score {
					bestScore = score - 47
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[7])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+47 < score {
					bestScore = score - 47
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[8])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+49 < score {
					bestScore = score - 49
					bestLen = ml
					lim = ringLimit(ringBufferMask, curMasked, bestLen)
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[9])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+49 < score {
					bestScore = score - 49
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
				}
			}
		}
	}

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: cur is stored after this scan.
	// minPrev avoids a subtraction for each bucket entry. maxBackward <= cur.
	// If bucket moves above phase 1, it spills into phase 1.
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)
	n := h.num[key]
	down := uint(0)
	if uint(n) > blockSize {
		down = uint(n) - blockSize
	}
	minPrev := cur - maxBackward
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	if curMasked+bestLen > ringBufferMask {
		down = uint(n)
	}
	for i := uint(n); i > down; {
		i--
		prevRaw := uint(bucket.at(i & blockMask))
		if prevRaw < minPrev {
			break
		}
		prevMasked := prevRaw & ringBufferMask
		if curProbe != loadU32LE(data, prevMasked+bestLen-3) || prevMasked+bestLen > ringBufferMask {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prevMasked, curMasked, int(maxLength)))
		if ml >= 4 {
			backward := cur - prevRaw
			score := backwardReferenceScore(ml, backward)
			if bestScore < score {
				bestScore = score
				bestLen = ml
				out.len = bestLen
				out.distance = backward
				out.score = bestScore
				if curMasked+bestLen > ringBufferMask {
					break
				}
				curProbe = loadU32LE(data, curMasked+bestLen-3)
			}
		}
	}

	bucket.put(uint(h.num[key])&blockMask, uint32(cur))
	h.num[key]++

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// findLongestMatchNoWrap uses direct indexes before the first ring wrap.
// Both cur and stored positions fit inside the ring.
func (h *h6b[B]) findLongestMatchNoWrap(
	data []byte,
	distCache *[16]int,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	// Keep this code in the method. A helper loads a generic dictionary per call.
	blockSize := uint(unsafe.Sizeof(h.buckets)) / h6bBucketSize / 4
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1
	bestScore := out.score
	bestLen := out.len
	key := h6bHash(data, cur)

	nextKey := h6bHash(data, cur+1)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	if hasPrefetch {
		prefetch2(unsafe.Add(nextBucket.base, (uint(nextN-1)&blockMask)<<2),
			unsafe.Add(nextBucket.base, (uint(nextN-17)&blockMask)<<2))
	} else {
		h.nextBucket = nextBucket.at(0)
		if nextN > 0 {
			p := uint(nextBucket.at(uint((nextN - 1) & uint16(blockMask))))
			h.nextBucket = uint32(data[p])
		}
	}

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	curByte := loadByte(data, cur+bestLen)
	backward := uint(distCache[0])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[1])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = uint(distCache[2])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+43 < score {
					bestScore = score - 43
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[3])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+43 < score {
					bestScore = score - 43
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[4])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+39 < score {
					bestScore = score - 39
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[5])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+39 < score {
					bestScore = score - 39
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[6])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+47 < score {
					bestScore = score - 47
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[7])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+47 < score {
					bestScore = score - 47
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[8])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+49 < score {
					bestScore = score - 49
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = uint(distCache[9])
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore+49 < score {
					bestScore = score - 49
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
				}
			}
		}
	}

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: cur is stored after this scan.
	//
	// minPrev = cur - maxBackward is equivalent to the backward > maxBackward break
	// condition but avoids computing backward = cur - prev on every iteration.
	// maxBackward = min(cur, maxBackwardLimit) <= cur so the subtraction never
	// wraps. backward is then computed lazily only when ml >= 4 (rare path).
	//
	// Do not hoist bucket above phase 1. It spills there, and the reload
	// lands at the head of the phase 1 loop.
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)
	n := h.num[key]
	down := uint(0)
	if uint(n) > blockSize {
		down = uint(n) - blockSize
	}
	minPrev := cur - maxBackward
	curProbe := loadU32LE(data, cur+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prevRaw := uint(bucket.at(i & blockMask))
		if prevRaw < minPrev {
			break
		}
		if curProbe != loadU32LE(data, prevRaw+bestLen-3) {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prevRaw, cur, int(maxLength)))
		if ml >= 4 {
			backward := cur - prevRaw
			score := backwardReferenceScore(ml, backward)
			if bestScore < score {
				bestScore = score
				bestLen = ml
				out.len = bestLen
				out.distance = backward
				out.score = bestScore
				curProbe = loadU32LE(data, cur+bestLen-3)
			}
		}
	}

	// Store current position in the bucket.
	bucket.put(uint(h.num[key])&blockMask, uint32(cur))
	h.num[key]++

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[cur:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// findLongestMatchSmallBuf is the version of findLongestMatch used when the
// ring buffer backing array is smaller than ringBufferMask+1.
func (h *h6b[B]) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache *[16]int,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	// Depth, shift and mask fold to constants in each instantiation.
	// Do not move this into a helper. A generic method that calls a generic
	// function loads a sub-dictionary on every call, also when the callee
	// inlines.
	blockSize := uint(unsafe.Sizeof(h.buckets)) / h6bBucketSize / 4
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h6bHash(data, curMasked)

	nextKey := h6bHash(data, (cur+1)&ringBufferMask)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	if hasPrefetch {
		prefetch2(unsafe.Add(nextBucket.base, (uint(nextN-1)&blockMask)<<2),
			unsafe.Add(nextBucket.base, (uint(nextN-17)&blockMask)<<2))
	} else {
		h.nextBucket = nextBucket.at(0)
		if nextN > 0 {
			p := uint(nextBucket.at(uint((nextN-1)&uint16(blockMask)))) & ringBufferMask
			h.nextBucket = uint32(data[p])
		}
	}

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	//
	// Cache entries 0 and 1 are unrolled: they accept ml >= 2 and use a fixed
	// penalty (0 for entry 0, 39 for entry 1) instead of the lookup-table
	// penalty used for entries 2..9.
	backward := uint(distCache[0])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curMasked+bestLen <= ringBufferMask &&
			prev+bestLen <= ringBufferMask &&
			data[curMasked+bestLen] == data[prev+bestLen] {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
				}
			}
		}
	}
	backward = uint(distCache[1])
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curMasked+bestLen <= ringBufferMask &&
			prev+bestLen <= ringBufferMask &&
			data[curMasked+bestLen] == data[prev+bestLen] {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
					}
				}
			}
		}
	}
	for i := uint(2); i < h6bNumLastDistances; i++ {
		backward := uint(distCache[i])
		if backward-1 >= maxBackward {
			continue
		}
		prev := (cur - backward) & ringBufferMask

		if curMasked+bestLen > ringBufferMask {
			break
		}
		if prev+bestLen > ringBufferMask ||
			data[curMasked+bestLen] != data[prev+bestLen] {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
		if ml >= 3 {
			score := backwardReferenceScoreUsingLastDistance(ml)
			if bestScore < score {
				score -= backwardReferencePenaltyUsingLastDistance(i)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
				}
			}
		}
	}

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: cur is stored after this scan.
	//
	// minPrev = cur - maxBackward avoids the per-iteration backward = cur - prev
	// subtraction; backward is only computed when ml >= 4 (rare path).
	//
	// Do not hoist bucket above phase 1. It spills there, and the reload
	// lands at the head of the phase 1 loop.
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)
	n := h.num[key]
	down := uint(0)
	if uint(n) > blockSize {
		down = uint(n) - blockSize
	}
	minPrev := cur - maxBackward
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prevRaw := uint(bucket.at(i & blockMask))
		if prevRaw < minPrev {
			break
		}
		prevMasked := prevRaw & ringBufferMask
		if curMasked+bestLen > ringBufferMask {
			break
		}
		if prevMasked+bestLen > ringBufferMask ||
			curProbe != loadU32LE(data, prevMasked+bestLen-3) {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prevMasked, curMasked, int(maxLength)))
		if ml >= 4 {
			backward := cur - prevRaw
			score := backwardReferenceScore(ml, backward)
			if bestScore < score {
				bestScore = score
				bestLen = ml
				out.len = bestLen
				out.distance = backward
				out.score = bestScore
				curProbe = loadU32LE(data, curMasked+bestLen-3)
			}
		}
	}

	// Store current position in the bucket.
	bucket.put(uint(h.num[key])&blockMask, uint32(cur))
	h.num[key]++

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct: the receiver is a concrete type, not a type parameter.
func (h *h6b[B]) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	mask := uint(s.mask)
	if !h.everWrapped && uint(wrappedPos)+uint(bytes) <= mask {
		h.createBackwardReferencesNoWrap(s, bytes, wrappedPos)
		return
	}
	h.everWrapped = true
	data := s.data
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h6bHashTypeLength {
		storeEnd = posEnd - h6bHashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 10 derived entries.
	var distCache [16]int
	for i, d := range s.distCache {
		distCache[i] = int(d)
	}
	prepareDistanceCache10(&distCache)

	for position+h6bHashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		h.findLongestMatch(data, mask, &distCache,
			position, maxLength, maxDistance, maxDistance+gap,
			&s.dictNumLookups, &s.dictNumMatches, &sr)
		if hasCompound {
			s.compound.lookupMatch(data, mask,
				&s.distCache, position, maxLength,
				maxDistance, &sr)
		}

		if sr.score > minScore {
			delayedBackwardReferencesInRow := 0
			maxLength--
			for {
				const costDiffLazy = 175
				var sr2 hasherSearchResult
				sr2.score = minScore
				maxDistance = min(position+1, maxBackwardLimit)

				h.findLongestMatch(data, mask, &distCache,
					position+1, maxLength, maxDistance, maxDistance+gap,
					&s.dictNumLookups, &s.dictNumMatches, &sr2)
				if hasCompound {
					s.compound.lookupMatch(data, mask,
						&s.distCache, position+1, maxLength,
						maxDistance, &sr2)
				}

				if sr2.score >= sr.score+costDiffLazy {
					position++
					insertLength++
					sr = sr2
					delayedBackwardReferencesInRow++
					if delayedBackwardReferencesInRow < 4 &&
						position+h6bHashTypeLength < posEnd {
						maxLength--
						continue
					}
				}
				break
			}

			applyRandomHeuristics = position + 2*sr.len + randomHeuristicsWindowSize

			maxDistance = min(position, maxBackwardLimit)
			distanceCode := computeDistanceCode(sr.distance, maxDistance+gap, &s.distCache)
			if sr.distance <= maxDistance+gap && distanceCode > 0 {
				s.distCache[3] = s.distCache[2]
				s.distCache[2] = s.distCache[1]
				s.distCache[1] = s.distCache[0]
				s.distCache[0] = sr.distance
			}

			s.pushCommandSimpleDist(insertLength, sr.len, sr.lenCodeDelta, distanceCode)
			s.numLiterals += insertLength
			insertLength = 0

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			h.storeRange(data, mask, rangeStart, rangeEnd)

			position += sr.len

			// Re-expand distance cache after updating it.
			for i, d := range s.distCache {
				distCache[i] = int(d)
			}
			prepareDistanceCache10(&distCache)
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h6bHashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h6bHashTypeLength-1))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 2
						position += 2
					}
				}
			}
		}
	}

	insertLength += posEnd - position
	s.lastInsertLen = insertLength
	s.numCommands += uint(len(s.commands)) - origCmdCount
}

// createBackwardReferencesNoWrap searches before the first ring wrap.
func (h *h6b[B]) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h6bHashTypeLength {
		storeEnd = posEnd - h6bHashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Add six derived distances to the four cache entries.
	var distCache [16]int
	for i, d := range s.distCache {
		distCache[i] = int(d)
	}
	prepareDistanceCache10(&distCache)

	for position+h6bHashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		h.findLongestMatchNoWrap(data, &distCache,
			position, maxLength, maxDistance, maxDistance+gap,
			&s.dictNumLookups, &s.dictNumMatches, &sr)
		if hasCompound {
			s.compound.lookupMatch(data, mask,
				&s.distCache, position, maxLength,
				maxDistance, &sr)
		}

		if sr.score > minScore {
			delayedBackwardReferencesInRow := 0
			maxLength--
			for {
				const costDiffLazy = 175
				var sr2 hasherSearchResult
				sr2.score = minScore
				maxDistance = min(position+1, maxBackwardLimit)

				h.findLongestMatchNoWrap(data, &distCache,
					position+1, maxLength, maxDistance, maxDistance+gap,
					&s.dictNumLookups, &s.dictNumMatches, &sr2)
				if hasCompound {
					s.compound.lookupMatch(data, mask,
						&s.distCache, position+1, maxLength,
						maxDistance, &sr2)
				}

				if sr2.score >= sr.score+costDiffLazy {
					position++
					insertLength++
					sr = sr2
					delayedBackwardReferencesInRow++
					if delayedBackwardReferencesInRow < 4 &&
						position+h6bHashTypeLength < posEnd {
						maxLength--
						continue
					}
				}
				break
			}

			applyRandomHeuristics = position + 2*sr.len + randomHeuristicsWindowSize

			maxDistance = min(position, maxBackwardLimit)
			distanceCode := computeDistanceCode(sr.distance, maxDistance+gap, &s.distCache)
			if sr.distance <= maxDistance+gap && distanceCode > 0 {
				s.distCache[3] = s.distCache[2]
				s.distCache[2] = s.distCache[1]
				s.distCache[1] = s.distCache[0]
				s.distCache[0] = sr.distance
			}

			s.pushCommandSimpleDist(insertLength, sr.len, sr.lenCodeDelta, distanceCode)
			s.numLiterals += insertLength
			insertLength = 0

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			h.storeRange(data, mask, rangeStart, rangeEnd)

			position += sr.len

			// Update the derived distances after a cache change.
			for i, d := range s.distCache {
				distCache[i] = int(d)
			}
			prepareDistanceCache10(&distCache)
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h6bHashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h6bHashTypeLength-1))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 2
						position += 2
					}
				}
			}
		}
	}

	insertLength += posEnd - position
	s.lastInsertLen = insertLength
	s.numCommands += uint(len(s.commands)) - origCmdCount
}

// prepareDistanceCache10 expands the 4 base distances into 6 near-miss
// entries derived from distCache[0].
func prepareDistanceCache10(distCache *[16]int) {
	last := distCache[0]
	distCache[4] = last - 1
	distCache[5] = last + 1
	distCache[6] = last - 2
	distCache[7] = last + 2
	distCache[8] = last - 3
	distCache[9] = last + 3
}
