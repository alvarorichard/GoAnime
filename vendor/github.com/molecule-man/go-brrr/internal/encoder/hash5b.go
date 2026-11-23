// H5 hashers for qualities 7 and 8.
//
// The bucket array length sets the search depth and remains a compile-time constant.
// Bucket types have no methods, which prevents generic dictionary calls in hot paths.

package encoder

import (
	"math/bits"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

const (
	h5bBucketBits = 15
	h5bBucketSize = 1 << h5bBucketBits // 32768
	h5bHashShift  = 32 - h5bBucketBits // 17

	// h5bHashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h5bHashTypeLength = 4

	h5b6BlockBits = 6
	h5b6BlockSize = 1 << h5b6BlockBits // 64
	h5b6BlockMask = h5b6BlockSize - 1
	h5b7BlockSize = 1 << 7 // 128
)

type h5bBlock interface {
	~[h5b6BlockSize]uint32 | ~[h5b7BlockSize]uint32
}

// h5b stores one position ring per hash bucket. B sets the search depth.
//
//nolint:govet // fieldalignment counts B as pointer data; every B is [N]uint32
type h5b[B h5bBlock] struct {
	num        [h5bBucketSize]uint16 // entry count per bucket
	buckets    [h5bBucketSize]B      // position ring buffers
	nextBucket uint32                // speculative load to warm cache
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions exceeding mask+1, after which the no-wrap fast path is
	// disabled because stored bucket values may then encode positions
	// outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

// h5b6 serves quality 7.
type h5b6 = h5b[[h5b6BlockSize]uint32]

// h5b7 serves quality 8.
type h5b7 = h5b[[h5b7BlockSize]uint32]

func (h *h5b[B]) common() *hasherCommon { return &h.hasherCommon }

func h5bHash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h5bHashShift
}

// reset zeroes the entry counts before use.
// When oneShot is true and the input is small, only the touched buckets
// are cleared (partial prepare). Otherwise the full count array is zeroed.
func (h *h5b[B]) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h5bBucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h5bHash(data, i)
			h.num[key] = 0
		}
	} else {
		h.num = [h5bBucketSize]uint16{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the ring buffer for the 4-byte sequence at
// data[pos & mask].
func (h *h5b[B]) store(data []byte, mask, pos uint) {
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	key := h5bHash(data, pos&mask)
	offset := uint(h.num[key])&(blockSize-1) + uint(key)*blockSize
	h.num[key]++
	*(*uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), offset<<2)) = uint32(pos)
}

// Keep the store loop here. A helper exceeds the inline budget.
func (h *h5b[B]) storeRange(data []byte, mask, start, end uint) {
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	for i := start; i < end; i++ {
		key := h5bHash(data, i&mask)
		offset := uint(h.num[key])&(blockSize-1) + uint(key)*blockSize
		h.num[key]++
		*(*uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), offset<<2)) = uint32(i)
	}
}

func (h *h5b[B]) storeNoWrap(data []byte, pos uint) {
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	key := h5bHash(data, pos)
	offset := uint(h.num[key])&(blockSize-1) + uint(key)*blockSize
	h.num[key]++
	*(*uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), offset<<2)) = uint32(pos)
}

func (h *h5b[B]) storeRangeNoWrap(data []byte, start, end uint) {
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	for i := start; i < end; i++ {
		key := h5bHash(data, i)
		offset := uint(h.num[key])&(blockSize-1) + uint(key)*blockSize
		h.num[key]++
		*(*uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), offset<<2)) = uint32(i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h5b[B]) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h5bHashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch checks cached distances, the hash bucket, and the static dictionary.
// It stores cur after the search.
func (h *h5b[B]) findLongestMatch(
	data []byte, ringBufferMask uint,
	distCache *[16]uint,
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

	// Keep these values local. A helper adds a generic dictionary load.
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1

	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h5bHash(data, curMasked)
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)
	// Issue the Phase 2 num[] load early so its (often L3-miss) latency is
	// hidden by Phase 1. n is held in a register until Phase 2.
	n := h.num[key]

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h5bHash(data, (cur+1)&ringBufferMask)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	h.nextBucket = nextBucket.at(0)
	if nextN > 0 {
		p := uint(nextBucket.at(uint((nextN-1)&uint16(blockMask)))) & ringBufferMask
		h.nextBucket = uint32(data[p])
	}

	out.len = 0
	out.lenCodeDelta = 0

	lim := ringLimit(ringBufferMask, curMasked, bestLen)

	// Phase 1: try cached distances.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	// The penalty constants below are backwardReferencePenaltyUsingLastDistance(i).
	// curByte caches loadByte(data, curMasked+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, curMasked+bestLen)
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 || ml == 2 {
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
	backward = distCache[1]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 || ml == 2 {
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
	backward = distCache[2]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
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
	backward = distCache[3]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
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
	backward = distCache[4]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
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
	backward = distCache[5]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
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
	backward = distCache[6]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
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
	backward = distCache[7]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
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
	backward = distCache[8]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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
	backward = distCache[9]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4
	// (the 4-byte quick rejection compares bestLen-3 .. bestLen).
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: we store cur after the loop, so all
	// bucket entries refer to strictly earlier positions.
	//
	// minPrev = cur - maxBackward is equivalent to the backward > maxBackward break
	// condition but avoids computing backward = cur - prev on every iteration.
	// maxBackward = min(cur, maxBackwardLimit) ≤ cur so the subtraction never
	// wraps. backward is then computed lazily only when ml ≥ 4 (rare path).
	// n was loaded near the top of the function so its latency overlaps Phase 1.
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

	// Store current position in the bucket.
	bucket.put(uint(h.num[key])&blockMask, uint32(cur))
	h.num[key]++

	// Phase 3: static dictionary fallback when no hash match was found.
	if bestScore == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// findLongestMatchSmallBuf handles backing arrays smaller than ringBufferMask+1.
func (h *h5b[B]) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache *[16]uint,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	// Keep these values local. A helper adds a generic dictionary load.
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1

	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h5bHash(data, curMasked)
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h5bHash(data, (cur+1)&ringBufferMask)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	h.nextBucket = nextBucket.at(0)
	if nextN > 0 {
		p := uint(nextBucket.at(uint((nextN-1)&uint16(blockMask)))) & ringBufferMask
		h.nextBucket = uint32(data[p])
	}

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// Note: the wrap-around guards (curMasked+bestLen > ringBufferMask and
	// prev+bestLen > ringBufferMask) are omitted. This function is only
	// reached when ringBufferMask >= len(data), which only occurs for small
	// one-shot inputs allocated via initRingBuffer(n) with n < tailSize.
	// In that case posEnd ≤ n ≤ ringBufferMask and prev < cur, so both
	// curMasked+bestLen and prev+bestLen stay strictly below ringBufferMask.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	// The penalty constants below are backwardReferencePenaltyUsingLastDistance(i).
	//
	// curByte = data[curMasked+bestLen] is hoisted across iterations. The
	// compiler can't CSE these loads because it can't prove writes through
	// *out don't alias *data, so it reloads them on each Phase 1 branch.
	// Refreshed only when bestLen changes (rare: common case is no match).
	curByte := loadByte(data, curMasked+bestLen)
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, curMasked+bestLen)
				}
			}
		}
	}
	backward = distCache[1]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[2]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[3]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[4]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[5]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[6]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[7]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[8]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, curMasked+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[9]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4
	// (the 4-byte quick rejection compares bestLen-3 .. bestLen).
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: we store cur after the loop, so all
	// bucket entries refer to strictly earlier positions.
	// Wrap-around guards omitted for the same reason as Phase 1.
	// minPrev replaces the backward > maxBackward break check; see findLongestMatch.
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
		if curProbe != loadU32LE(data, prevMasked+bestLen-3) {
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
	if bestScore == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// createBackwardReferences adds matches to s.commands.
// It uses the no-wrap path while stored positions cannot exceed mask+1.
func (h *h5b[B]) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
	if uint(bytes) >= h5bHashTypeLength {
		storeEnd = posEnd - h5bHashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 10 derived entries.
	var distCache [16]uint
	d0 := s.distCache[0]
	distCache[0] = d0
	distCache[1] = s.distCache[1]
	distCache[2] = s.distCache[2]
	distCache[3] = s.distCache[3]
	distCache[4] = d0 - 1
	distCache[5] = d0 + 1
	distCache[6] = d0 - 2
	distCache[7] = d0 + 2
	distCache[8] = d0 - 3
	distCache[9] = d0 + 3

	for position+h5bHashTypeLength < posEnd {
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
						position+h5bHashTypeLength < posEnd {
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
				// Update the local cache without reads from s.distCache.
				distCache[3] = distCache[2]
				distCache[2] = distCache[1]
				distCache[1] = d0
				d0 = sr.distance
				distCache[0] = d0
				distCache[4] = d0 - 1
				distCache[5] = d0 + 1
				distCache[6] = d0 - 2
				distCache[7] = d0 + 2
				distCache[8] = d0 - 3
				distCache[9] = d0 + 3
			}
			// distanceCode == 0 leaves both caches unchanged.

			// Keep this code inline to avoid a function call and a command copy.
			{
				delta := uint32(uint8(int8(sr.lenCodeDelta)))
				distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
				effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
				insCode := getInsertLenCode(insertLength)
				copyCode := getCopyLenCode(effectiveCopyLen)
				cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
				s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
			}
			s.numLiterals += insertLength
			insertLength = 0

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			h.storeRange(data, mask, rangeStart, rangeEnd)

			position += sr.len
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h5bHashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h5bHashTypeLength-1))
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

// createBackwardReferencesNoWrap omits position masks while all positions remain below mask+1.
func (h *h5b[B]) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h5bHashTypeLength {
		storeEnd = posEnd - h5bHashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 10 derived entries.
	var distCache [16]uint
	d0 := s.distCache[0]
	distCache[0] = d0
	distCache[1] = s.distCache[1]
	distCache[2] = s.distCache[2]
	distCache[3] = s.distCache[3]
	distCache[4] = d0 - 1
	distCache[5] = d0 + 1
	distCache[6] = d0 - 2
	distCache[7] = d0 + 2
	distCache[8] = d0 - 3
	distCache[9] = d0 + 3

	for position+h5bHashTypeLength < posEnd {
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
						position+h5bHashTypeLength < posEnd {
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
				distCache[3] = distCache[2]
				distCache[2] = distCache[1]
				distCache[1] = d0
				d0 = sr.distance
				distCache[0] = d0
				distCache[4] = d0 - 1
				distCache[5] = d0 + 1
				distCache[6] = d0 - 2
				distCache[7] = d0 + 2
				distCache[8] = d0 - 3
				distCache[9] = d0 + 3
			}

			{
				delta := uint32(uint8(int8(sr.lenCodeDelta)))
				distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
				effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
				insCode := getInsertLenCode(insertLength)
				copyCode := getCopyLenCode(effectiveCopyLen)
				cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
				s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
			}
			s.numLiterals += insertLength
			insertLength = 0

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			h.storeRangeNoWrap(data, rangeStart, rangeEnd)

			position += sr.len
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h5bHashTypeLength-1, 4))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h5bHashTypeLength-1))
					for position < posJump {
						h.storeNoWrap(data, position)
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

// findLongestMatchNoWrap omits position masks while cur and stored positions remain below mask+1.
func (h *h5b[B]) findLongestMatchNoWrap(
	data []byte,
	distCache *[16]uint,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	// Keep these values local. A helper adds a generic dictionary load.
	blockSize := uint(unsafe.Sizeof(h.buckets[0])) >> 2
	blockShift := uint(bits.TrailingZeros(blockSize))
	blockMask := blockSize - 1

	bestScore := out.score
	bestLen := out.len
	key := h5bHash(data, cur)
	bucket := bucketRingAt(unsafe.Pointer(&h.buckets), key, blockShift)
	// Issue the Phase 2 num[] load early so its (often L3-miss) latency is
	// hidden by Phase 1.
	n := h.num[key]

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h5bHash(data, cur+1)
	nextBucket := bucketRingAt(unsafe.Pointer(&h.buckets), nextKey, blockShift)
	nextN := h.num[nextKey]
	h.nextBucket = nextBucket.at(0)
	if nextN > 0 {
		p := uint(nextBucket.at(uint((nextN - 1) & uint16(blockMask))))
		h.nextBucket = uint32(data[p])
	}

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// curByte caches loadByte(data, cur+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, cur+bestLen)
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
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
	backward = distCache[1]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
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
	backward = distCache[2]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
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
	backward = distCache[3]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
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
	backward = distCache[4]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
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
	backward = distCache[5]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
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
	backward = distCache[6]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
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
	backward = distCache[7]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
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
	backward = distCache[8]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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
	backward = distCache[9]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: scan hash bucket entries.
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

		ml := uint(matchLenAt(data, prevRaw, cur, int(maxLength)))
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
	if bestScore == minScore {
		searchStaticDictionaryDeep(data[cur:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}
