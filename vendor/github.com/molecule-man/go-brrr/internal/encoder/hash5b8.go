// H5 hasher variant with bucketBits=15, blockBits=8 for quality 9.
//
// Compared to H5b7 (quality 8), this variant doubles the per-bucket depth
// (256 vs 128 entries) and expands the distance cache to 16 entries (4 base
// plus 6 derived near-miss entries for each of dist[0] and dist[1]).

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// h5b8 configuration constants for quality 9.
const (
	h5b8BucketBits = 15
	h5b8BucketSize = 1 << h5b8BucketBits // 32768
	h5b8BlockBits  = 8
	h5b8BlockSize  = 1 << h5b8BlockBits // 256
	h5b8BlockMask  = h5b8BlockSize - 1
	h5b8HashShift  = 32 - h5b8BucketBits // 17

	// h5b8HashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h5b8HashTypeLength = 4
)

// h5b8 is the H5 hasher with bucketBits=15 and blockBits=8: a forgetful hash
// table where each of 32K buckets holds a ring buffer of up to 256 positions.
type h5b8 struct {
	num        [h5b8BucketSize]uint16                 // entry count per bucket
	buckets    [h5b8BucketSize * h5b8BlockSize]uint32 // position ring buffers
	nextBucket uint32                                 // speculative load to warm cache
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions reaching or exceeding mask+1, after which the no-wrap
	// fast path is disabled because stored bucket values may then encode
	// positions outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

func (h *h5b8) common() *hasherCommon { return &h.hasherCommon }

// hash computes a 15-bit bucket index from 4 bytes at data[i:i+4].
func (h *h5b8) hash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h5b8HashShift
}

// reset zeroes the entry counts before use.
// When oneShot is true and the input is small, only the touched buckets
// are cleared (partial prepare). Otherwise the full count array is zeroed.
func (h *h5b8) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h5b8BucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h.hash(data, i)
			h.num[key] = 0
		}
	} else {
		h.num = [h5b8BucketSize]uint16{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the ring buffer for the 4-byte sequence at
// data[pos & mask].
func (h *h5b8) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	minorIx := h.num[key] & h5b8BlockMask
	offset := uint(minorIx) + uint(key)<<h5b8BlockBits
	h.num[key]++
	h.buckets[offset] = uint32(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h5b8) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.store(data, mask, i)
	}
}

func (h *h5b8) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	minorIx := h.num[key] & h5b8BlockMask
	offset := uint(minorIx) + uint(key)<<h5b8BlockBits
	h.num[key]++
	h.buckets[offset] = uint32(pos)
}

func (h *h5b8) storeRangeNoWrap(data []byte, start, end uint) {
	for i := start; i < end; i++ {
		h.storeNoWrap(data, i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h5b8) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h5b8HashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch searches for the best backward reference at position cur
// in the ring buffer, then stores cur in the hash table.
//
// The search has three phases:
//  1. Distance cache: try the last 16 cached distances (4 base entries plus
//     6 derived near-miss entries for each of dist[0] and dist[1]). Accept
//     length >= 3, or length == 2 for the first two cache entries.
//  2. Hash bucket scan: walk the ring buffer of up to 256 positions for the
//     bucket. Reject candidates with a 4-byte quick comparison, accept
//     length >= 4.
//  3. Static dictionary fallback: when neither phase produced a match,
//     search the static dictionary with deep search.
func (h *h5b8) findLongestMatch(
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

	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)
	bucket := h.buckets[uint(key)<<h5b8BlockBits:]
	// Issue the Phase 2 num[] load early so its (often L3-miss) latency is
	// hidden by Phase 1. n is held in a register until Phase 2.
	n := h.num[key]

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h.hash(data, (cur+1)&ringBufferMask)
	nextBase := uint(nextKey) << h5b8BlockBits
	h.nextBucket = h.buckets[nextBase]

	out.len = 0
	out.lenCodeDelta = 0

	lim := ringLimit(ringBufferMask, curMasked, bestLen)

	// Phase 1: try cached distances (fully unrolled for 16 entries).
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	// Penalty constants from backwardReferencePenaltyUsingLastDistance(i).
	// curByte caches loadByte(data, curMasked+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, curMasked+bestLen)
	{
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
	}
	{
		backward := distCache[1]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 || ml == 2 {
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
	}
	{
		backward := distCache[2]
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
	}
	{
		backward := distCache[3]
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
	}
	{
		backward := distCache[4]
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
	}
	{
		backward := distCache[5]
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
	}
	{
		backward := distCache[6]
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
	}
	{
		backward := distCache[7]
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
	}
	{
		backward := distCache[8]
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
	}
	{
		backward := distCache[9]
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
	}
	{
		backward := distCache[10]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[11]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[12]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[13]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[14]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
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
	{
		backward := distCache[15]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) && prev < lim {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
						bestLen = ml
						lim = ringLimit(ringBufferMask, curMasked, bestLen)
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
		lim = ringLimit(ringBufferMask, curMasked, bestLen)
	}

	// Phase 2: scan hash bucket entries.
	// backward == 0 is impossible here: cur is stored after this scan.
	down := uint(0)
	if uint(n) > h5b8BlockSize {
		down = uint(n) - h5b8BlockSize
	}
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prev := uint(bucket[i&h5b8BlockMask])
		backward := cur - prev
		if backward > maxBackward {
			break
		}
		prev &= ringBufferMask
		if curProbe != loadU32LE(data, prev+bestLen-3) || prev >= lim {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
		if ml >= 4 {
			score := backwardReferenceScore(ml, backward)
			if bestScore < score {
				bestScore = score
				bestLen = ml
				lim = ringLimit(ringBufferMask, curMasked, bestLen)
				out.len = bestLen
				out.distance = backward
				out.score = bestScore
				curProbe = loadU32LE(data, curMasked+bestLen-3)
			}
		}
	}

	// Store current position in the bucket.
	h.buckets[uint(n&h5b8BlockMask)+uint(key)<<h5b8BlockBits] = uint32(cur)
	h.num[key] = n + 1

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// findLongestMatchSmallBuf is the generic version of findLongestMatch used
// when the ring buffer backing array is smaller than ringBufferMask+1.
func (h *h5b8) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache *[16]uint,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)
	bucket := h.buckets[uint(key)<<h5b8BlockBits:]
	// Issue the Phase 2 num[] load early so its (often L3-miss) latency is
	// hidden by Phase 1. n is held in a register until Phase 2.
	n := h.num[key]

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h.hash(data, (cur+1)&ringBufferMask)
	nextBase := uint(nextKey) << h5b8BlockBits
	h.nextBucket = h.buckets[nextBase]

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances (fully unrolled for 16 entries).
	// Wrap-around guards are omitted in this small-buffer path. It is only
	// reached for small one-shot inputs where posEnd <= len(input) <=
	// ringBufferMask, and prev < cur, so curMasked+bestLen and prev+bestLen
	// stay within the copied input plus zero tail.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	// Penalty constants from backwardReferencePenaltyUsingLastDistance(i).
	// curByte caches loadByte(data, curMasked+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, curMasked+bestLen)
	{
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
	}
	{
		backward := distCache[1]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 || ml == 2 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+39 < score {
						bestScore = score - 39
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
	{
		backward := distCache[2]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+43 < score {
						bestScore = score - 43
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
	{
		backward := distCache[3]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+43 < score {
						bestScore = score - 43
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
	{
		backward := distCache[4]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+39 < score {
						bestScore = score - 39
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
	{
		backward := distCache[5]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+39 < score {
						bestScore = score - 39
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
	{
		backward := distCache[6]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+47 < score {
						bestScore = score - 47
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
	{
		backward := distCache[7]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+47 < score {
						bestScore = score - 47
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
	{
		backward := distCache[8]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+49 < score {
						bestScore = score - 49
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
	{
		backward := distCache[9]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+49 < score {
						bestScore = score - 49
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
	{
		backward := distCache[10]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[11]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[12]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[13]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[14]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
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
	{
		backward := distCache[15]
		if backward-1 < maxBackward {
			prev := (cur - backward) & ringBufferMask
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
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
	// Wrap-around guards are omitted for the same reason as Phase 1.
	// backward == 0 is impossible here: cur is stored after this scan.
	down := uint(0)
	if uint(n) > h5b8BlockSize {
		down = uint(n) - h5b8BlockSize
	}
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prev := uint(bucket[i&h5b8BlockMask])
		backward := cur - prev
		if backward > maxBackward {
			break
		}
		prev &= ringBufferMask
		if curProbe != loadU32LE(data, prev+bestLen-3) {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prev, curMasked, int(maxLength)))
		if ml >= 4 {
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
	h.buckets[uint(n&h5b8BlockMask)+uint(key)<<h5b8BlockBits] = uint32(cur)
	h.num[key] = n + 1

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct (non-virtual) since the receiver is concrete.
//
// When the call's [wrappedPos, wrappedPos+bytes) range fits entirely within
// the ring buffer (no modular wrap) and no past call has wrapped, dispatch
// to createBackwardReferencesNoWrap which omits the per-iteration & mask
// ops (each redundant when stored bucket values are < mask+1).
func (h *h5b8) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
	if uint(bytes) >= h5b8HashTypeLength {
		storeEnd = posEnd - h5b8HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 512
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 16 derived entries.
	var distCache [16]uint
	d0 := s.distCache[0]
	d1 := s.distCache[1]
	distCache[0] = d0
	distCache[1] = d1
	distCache[2] = s.distCache[2]
	distCache[3] = s.distCache[3]
	distCache[4] = d0 - 1
	distCache[5] = d0 + 1
	distCache[6] = d0 - 2
	distCache[7] = d0 + 2
	distCache[8] = d0 - 3
	distCache[9] = d0 + 3
	distCache[10] = d1 - 1
	distCache[11] = d1 + 1
	distCache[12] = d1 - 2
	distCache[13] = d1 + 2
	distCache[14] = d1 - 3
	distCache[15] = d1 + 3

	for position+h5b8HashTypeLength < posEnd {
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
						position+h5b8HashTypeLength < posEnd {
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
			d0 = s.distCache[0]
			d1 = s.distCache[1]
			distCache[0] = d0
			distCache[1] = d1
			distCache[2] = s.distCache[2]
			distCache[3] = s.distCache[3]
			distCache[4] = d0 - 1
			distCache[5] = d0 + 1
			distCache[6] = d0 - 2
			distCache[7] = d0 + 2
			distCache[8] = d0 - 3
			distCache[9] = d0 + 3
			distCache[10] = d1 - 1
			distCache[11] = d1 + 1
			distCache[12] = d1 - 2
			distCache[13] = d1 + 2
			distCache[14] = d1 - 3
			distCache[15] = d1 + 3
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h5b8HashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h5b8HashTypeLength-1))
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

// createBackwardReferencesNoWrap is the no-wrap fast path used by
// createBackwardReferences when the call's position range fits entirely
// within mask+1 and no past call has wrapped. Stored bucket values are
// then guaranteed < mask+1, so per-iteration `prev &= mask` and
// `position & mask` ops in findLongestMatch are redundant and elided
// here via findLongestMatchNoWrap. The paired no-wrap store helpers also
// skip redundant position masking while preserving the same bucket update.
func (h *h5b8) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h5b8HashTypeLength {
		storeEnd = posEnd - h5b8HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 512
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 16 derived entries.
	var distCache [16]uint
	d0 := s.distCache[0]
	d1 := s.distCache[1]
	distCache[0] = d0
	distCache[1] = d1
	distCache[2] = s.distCache[2]
	distCache[3] = s.distCache[3]
	distCache[4] = d0 - 1
	distCache[5] = d0 + 1
	distCache[6] = d0 - 2
	distCache[7] = d0 + 2
	distCache[8] = d0 - 3
	distCache[9] = d0 + 3
	distCache[10] = d1 - 1
	distCache[11] = d1 + 1
	distCache[12] = d1 - 2
	distCache[13] = d1 + 2
	distCache[14] = d1 - 3
	distCache[15] = d1 + 3

	for position+h5b8HashTypeLength < posEnd {
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
						position+h5b8HashTypeLength < posEnd {
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
				// Re-expand local distCache using existing local values to avoid
				// re-reading from s.distCache (heap). The shift mirrors s.distCache.
				distCache[3] = distCache[2]
				distCache[2] = d1
				d1 = d0
				d0 = sr.distance
				distCache[0] = d0
				distCache[1] = d1
				distCache[4] = d0 - 1
				distCache[5] = d0 + 1
				distCache[6] = d0 - 2
				distCache[7] = d0 + 2
				distCache[8] = d0 - 3
				distCache[9] = d0 + 3
				distCache[10] = d1 - 1
				distCache[11] = d1 + 1
				distCache[12] = d1 - 2
				distCache[13] = d1 + 2
				distCache[14] = d1 - 3
				distCache[15] = d1 + 3
			}

			s.pushCommandSimpleDist(insertLength, sr.len, sr.lenCodeDelta, distanceCode)
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
					posJump := min(position+16, posEnd-max(h5b8HashTypeLength-1, 4))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h5b8HashTypeLength-1))
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

// findLongestMatchNoWrap is the no-wrap variant of findLongestMatch's fast
// path. It assumes cur < mask+1 and that all stored bucket values are
// < mask+1, which makes `& ringBufferMask` ops redundant — they are elided
// here. Otherwise the search structure (Phase 1 distance cache, Phase 2
// bucket scan, Phase 3 dictionary fallback) matches findLongestMatch.
func (h *h5b8) findLongestMatchNoWrap(
	data []byte,
	distCache *[16]uint,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, cur)
	bucket := h.buckets[uint(key)<<h5b8BlockBits:]
	// Issue the Phase 2 num[] load early so its (often L3-miss) latency is
	// hidden by Phase 1. n is held in a register until Phase 2.
	n := h.num[key]

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances (fully unrolled for 16 entries).
	// curByte caches loadByte(data, cur+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, cur+bestLen)
	{
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
	}
	{
		backward := distCache[1]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 || ml == 2 {
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
	}
	{
		backward := distCache[2]
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
	}
	{
		backward := distCache[3]
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
	}
	{
		backward := distCache[4]
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
	}
	{
		backward := distCache[5]
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
	}
	{
		backward := distCache[6]
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
	}
	{
		backward := distCache[7]
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
	}
	{
		backward := distCache[8]
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
	}
	{
		backward := distCache[9]
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
	}
	{
		backward := distCache[10]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[11]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+41 < score {
						bestScore = score - 41
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
	{
		backward := distCache[12]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[13]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+51 < score {
						bestScore = score - 51
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
	{
		backward := distCache[14]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
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
	{
		backward := distCache[15]
		if backward-1 < maxBackward {
			prev := cur - backward
			if curByte == loadByte(data, prev+bestLen) {
				ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
				if ml >= 3 {
					score := backwardReferenceScoreUsingLastDistance(ml)
					if bestScore+45 < score {
						bestScore = score - 45
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
	// minPrev = cur - maxBackward replaces the backward > maxBackward break check
	// and avoids computing backward = cur - prev on every iteration.
	down := uint(0)
	if uint(n) > h5b8BlockSize {
		down = uint(n) - h5b8BlockSize
	}
	minPrev := cur - maxBackward
	curProbe := loadU32LE(data, cur+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prevRaw := uint(bucket[i&h5b8BlockMask])
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
	h.buckets[uint(n&h5b8BlockMask)+uint(key)<<h5b8BlockBits] = uint32(cur)
	h.num[key] = n + 1

	// Phase 3: static dictionary fallback when no hash match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[cur:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}
