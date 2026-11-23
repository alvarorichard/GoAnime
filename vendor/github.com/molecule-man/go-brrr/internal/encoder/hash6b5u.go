// H6b5u stores positions without tags for quality 6 and large inputs.
//
// A bucket scan reads data for each entry. The encoder converts h6b5u to
// h6b5 when many entries have different tags.
//
// The encoder selects h6b5u when quality=6, sizeHint >= 1MiB, and lgwin >= 19.

package encoder

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// h6b5u converts to h6b5 after h6b5uBusyCalls consecutive calls exceed
// h6b5uBusyRatio tag rejects per store. It samples every h6b5uSampleRate
// positions outside lazy match searches.
const (
	h6b5uBusyRatio  = 5
	h6b5uBusyCalls  = 3
	h6b5uSampleRate = 256
)

// h6b5u stores up to h6b5BlockSize (32) positions per bucket.
type h6b5u struct {
	num     [h6b5BucketSize]uint16                 // entry count per bucket
	buckets [h6b5BucketSize * h6b5BlockSize]uint32 // position ring buffers
	// rejects and stores hold the current call's switch metrics.
	// busyCalls counts consecutive calls above h6b5uBusyRatio.
	rejects, stores uint
	busyCalls       int
	nextBucket      uint32 // speculative load to warm cache
	// everWrapped disables the direct position path after the first ring wrap.
	everWrapped bool
	hasherCommon
}

func (h *h6b5u) common() *hasherCommon { return &h.hasherCommon }

// bucketAt returns the ring buffer for key.
// The hash limits key to h6b5BucketSize. The array pointer removes a bounds
// check from each scan step.
func (h *h6b5u) bucketAt(key uint32) *[h6b5BlockSize]uint32 {
	return (*[h6b5BlockSize]uint32)(unsafe.Add(unsafe.Pointer(&h.buckets), uintptr(key)<<(h6b5BlockBits+2)))
}

// hash computes a 15-bit bucket index from 8 bytes at data[i:i+8].
func (h *h6b5u) hash(data []byte, i uint) uint32 {
	return uint32((loadU64LE(data, i) * h6b5HashMul) >> h6b5HashShift)
}

// reset zeroes the entry counts before use.
// When oneShot is true and the input is small, only the touched buckets
// are cleared (partial prepare). Otherwise the full count array is zeroed.
func (h *h6b5u) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h6b5BucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h.hash(data, i)
			h.num[key] = 0
		}
	} else {
		h.num = [h6b5BucketSize]uint16{}
	}
	h.everWrapped = false
	h.rejects, h.stores, h.busyCalls = 0, 0, 0
	h.ready = true
}

// store records position pos in the ring buffer for the 8-byte sequence at
// data[pos & mask].
func (h *h6b5u) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	minorIx := h.num[key] & h6b5BlockMask
	offset := uint(minorIx) + uint(key)<<h6b5BlockBits
	h.num[key]++
	h.buckets[offset] = uint32(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h6b5u) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.store(data, mask, i)
	}
}

func (h *h6b5u) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	minorIx := h.num[key] & h6b5BlockMask
	offset := uint(minorIx) + uint(key)<<h6b5BlockBits
	h.num[key]++
	h.buckets[offset] = uint32(pos)
}

func (h *h6b5u) storeRangeNoWrap(data []byte, start, end uint) {
	for i := start; i < end; i++ {
		h.storeNoWrap(data, i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h6b5u) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h6b5HashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch searches for the best backward reference at position cur
// in the ring buffer, then stores cur in the hash table.
//
// The search has two phases:
//  1. Distance cache: try the last 4 cached distances (and 6 derived
//     near-miss distances for the first two). Accept length >= 3, or
//     length == 2 for the first two cache entries.
//  2. Hash bucket scan: walk the ring buffer of up to 32 positions for the
//     bucket. Reject candidates with a 4-byte quick comparison, accept
//     length >= 4.
//
// The caller searches the static dictionary if neither phase finds a match.
func (h *h6b5u) findLongestMatch(
	data []byte, ringBufferMask uint,
	distCache *[4]uint,
	cur, maxLength, maxBackward uint,
	out *hasherSearchResult,
) {
	if ringBufferMask >= uint(len(data)) {
		h.findLongestMatchSmallBuf(data, ringBufferMask, distCache,
			cur, maxLength, maxBackward, out)
		return
	}

	// --- fast path: ringBufferMask < len(data) ---
	// The mirrored tail permits probe loads past the ring end.
	// The scan rejects those matches after the probe.
	_ = data[ringBufferMask]

	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)

	out.len = 0
	out.lenCodeDelta = 0

	lim := ringLimit(ringBufferMask, curMasked, bestLen)

	// Phase 1: try cached distances. Unrolled so the per-entry conditions
	// (penalty index, ml >= 2 acceptance) are compile-time constants.
	curByte := loadByte(data, curMasked+bestLen)
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := (cur - backward) & ringBufferMask
		if curByte == loadByte(data, prev+bestLen) && prev < lim {
			ml := uint(matchLenAt(data, prev, curMasked, int(maxLength)))
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
			ml := uint(matchLenAt(data, prev, curMasked, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(1)
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
			ml := uint(matchLenAt(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(2)
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
			ml := uint(matchLenAt(data, prev, curMasked, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(3)
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
	// backward == 0 is impossible here: cur is stored after this scan.
	// Keep the scan last so no values remain live across the call.
	minPrev := cur - maxBackward
	h6b5uFindInBucket(h, unsafe.Pointer(unsafe.SliceData(data)), uint(key),
		cur, curMasked, ringBufferMask, minPrev, maxLength, bestLen, bestScore,
		uint(h.hash(data, (cur+1)&ringBufferMask)), out)
}

// findLongestMatchSmallBuf is the generic version of findLongestMatch used
// when the ring buffer backing array is smaller than ringBufferMask+1.
func (h *h6b5u) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache *[4]uint,
	cur, maxLength, maxBackward uint,
	out *hasherSearchResult,
) {
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)
	bucket := h.bucketAt(key)

	// Speculatively load from the next position's bucket to warm the cache.
	nextKey := h.hash(data, (cur+1)&ringBufferMask)
	nextBucket := h.bucketAt(nextKey)
	nextN := h.num[nextKey]
	h.nextBucket = nextBucket[0]
	if nextN > 0 {
		p := uint(nextBucket[(nextN-1)&h6b5BlockMask]) & ringBufferMask
		h.nextBucket = uint32(data[p])
	}

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// backward-1 >= maxBackward is a single check replacing both
	// "prev >= cur" (backward==0) and "backward > maxBackward".
	for i := range uint(h6b5NumLastDistances) {
		backward := distCache[i]
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
		if ml >= 3 || (ml == 2 && i < 2) {
			score := backwardReferenceScoreUsingLastDistance(ml)
			if bestScore < score {
				if i != 0 {
					score -= backwardReferencePenaltyUsingLastDistance(i)
				}
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
	n := h.num[key]
	down := uint(0)
	if uint(n) > h6b5BlockSize {
		down = uint(n) - h6b5BlockSize
	}
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prev := uint(bucket[i&h6b5BlockMask])
		backward := cur - prev
		if backward > maxBackward {
			break
		}
		prev &= ringBufferMask
		if curMasked+bestLen > ringBufferMask {
			break
		}
		if prev+bestLen > ringBufferMask ||
			curProbe != loadU32LE(data, prev+bestLen-3) {
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
	bucket[h.num[key]&h6b5BlockMask] = uint32(cur)
	h.num[key]++
}

// sampleRejects adds to h.rejects the bucket entries with another tag
// than the sequence at data[i], times h6b5uSampleRate.
//
//go:noinline
func (h *h6b5u) sampleRejects(data []byte, mask, i uint) {
	key, tag := h.hashTag(data, i)
	bucket := h.bucketAt(key)
	n := uint(h.num[key])
	for j := range min(n, h6b5BlockSize) {
		p := uint(bucket[(n-1-j)&h6b5BlockMask]) & mask
		if p+8 > uint(len(data)) {
			continue
		}
		if _, t := h.hashTag(data, p); t != tag {
			h.rejects += h6b5uSampleRate
		}
	}
}

// hashTag computes the bucket key and the 8-bit tag of the entry at data[i:i+8].
func (h *h6b5u) hashTag(data []byte, i uint) (uint32, uint8) {
	v := (loadU64LE(data, i) * h6b5HashMul) >> (h6b5HashShift - 8)
	return uint32(v >> 8), uint8(v)
}

// countBusyCall ends the statistics of a createBackwardReferences call.
func (h *h6b5u) countBusyCall() {
	if h.rejects > h6b5uBusyRatio*h.stores {
		h.busyCalls++
	} else {
		h.busyCalls = 0
	}
	h.rejects, h.stores = 0, 0
}

// wantsTags reports whether the scan of h6b5 is faster for the input.
func (h *h6b5u) wantsTags() bool { return h.busyCalls >= h6b5uBusyCalls }

// copyTo copies the hash table into n, which then continues the stream.
// It computes the tags from data. The tag of an entry outside the window
// can be wrong: the h6b5 scan stops before such an entry or skips it,
// which gives the same matches.
func (h *h6b5u) copyTo(n *h6b5, data []byte, mask uint) {
	for key := range h.num {
		// h6b5 counts down and fills the slots in the other direction.
		n.num[key] = 0xFFFF - h.num[key]
		b := &n.blocks[key]
		for slot, p := range h.bucketAt(uint32(key)) {
			pos := &b.pos[h6b5BlockMask-slot]
			*pos = p
			if i := uint(p) & mask; i+8 <= uint(len(data)) {
				_, b.tags[h6b5BlockMask-slot] = h.hashTag(data, i)
			}
		}
	}
	n.everWrapped = h.everWrapped
	n.ready = h.ready
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct (non-virtual) since the receiver is concrete.
//
// If this call and past calls stay within the ring, use the path without masks.
func (h *h6b5u) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
	if uint(bytes) >= h6b5HashTypeLength {
		storeEnd = posEnd - h6b5HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	distCache := &s.distCache

	for position+h6b5HashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		if position%h6b5uSampleRate == 0 {
			h.sampleRejects(data, mask, position&mask)
		}
		h.findLongestMatch(data, mask, distCache, position, maxLength, maxDistance, &sr)
		if sr.score == minScore {
			searchStaticDictionaryDeep(data[position&mask:], maxLength, maxDistance+gap, maxBackwardDistance,
				&s.dictNumLookups, &s.dictNumMatches, &sr)
		}
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

				h.findLongestMatch(data, mask, distCache, position+1, maxLength, maxDistance, &sr2)
				if sr2.score == minScore {
					searchStaticDictionaryDeep(data[(position+1)&mask:], maxLength, maxDistance+gap, maxBackwardDistance,
						&s.dictNumLookups, &s.dictNumMatches, &sr2)
				}
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
						position+h6b5HashTypeLength < posEnd {
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
			h.stores += rangeEnd - rangeStart
			h.storeRange(data, mask, rangeStart, rangeEnd)

			position += sr.len
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h6b5HashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h6b5HashTypeLength-1))
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
	h.countBusyCall()
}

// createBackwardReferencesNoWrap uses direct positions before the first ring wrap.
// All stored positions fit within the ring, so scans and stores omit masks.
func (h *h6b5u) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h6b5HashTypeLength {
		storeEnd = posEnd - h6b5HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	distCache := &s.distCache

	for position+h6b5HashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		if position%h6b5uSampleRate == 0 {
			h.sampleRejects(data, mask, position&mask)
		}
		h.findLongestMatchNoWrap(data, distCache, position, maxLength, maxDistance, &sr)
		if sr.score == minScore {
			searchStaticDictionaryDeep(data[position:], maxLength, maxDistance+gap, maxBackwardDistance,
				&s.dictNumLookups, &s.dictNumMatches, &sr)
		}
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

				h.findLongestMatchNoWrap(data, distCache, position+1, maxLength, maxDistance, &sr2)
				if sr2.score == minScore {
					searchStaticDictionaryDeep(data[position+1:], maxLength, maxDistance+gap, maxBackwardDistance,
						&s.dictNumLookups, &s.dictNumMatches, &sr2)
				}
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
						position+h6b5HashTypeLength < posEnd {
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
			h.stores += rangeEnd - rangeStart
			h.storeRangeNoWrap(data, rangeStart, rangeEnd)

			position += sr.len
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h6b5HashTypeLength-1, 4))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h6b5HashTypeLength-1))
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
	h.countBusyCall()
}

// findLongestMatchNoWrap searches before the first ring wrap.
// Current and stored positions fit within the ring, so the scan omits masks.
// The caller searches the static dictionary if this search finds no match.
func (h *h6b5u) findLongestMatchNoWrap(
	data []byte,
	distCache *[4]uint,
	cur, maxLength, maxBackward uint,
	out *hasherSearchResult,
) {
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, cur)

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances. Unrolled so the per-entry conditions
	// (penalty index, ml >= 2 acceptance) are compile-time constants.
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := cur - backward
		if loadByte(data, cur+bestLen) == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAt(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
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

	backward = distCache[1]
	if backward-1 < maxBackward {
		prev := cur - backward
		if loadByte(data, cur+bestLen) == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAt(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(1)
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

	backward = distCache[2]
	if backward-1 < maxBackward {
		prev := cur - backward
		if loadByte(data, cur+bestLen) == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAt(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(2)
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

	backward = distCache[3]
	if backward-1 < maxBackward {
		prev := cur - backward
		if loadByte(data, cur+bestLen) == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAt(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= backwardReferencePenaltyUsingLastDistance(3)
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
	// Keep the scan last so no values remain live across the call.
	minPrev := cur - maxBackward
	h6b5uFindInBucket(h, unsafe.Pointer(unsafe.SliceData(data)), uint(key),
		cur, cur, ^uint(0), minPrev, maxLength, bestLen, bestScore,
		uint(h.hash(data, cur+1)), out)
}
