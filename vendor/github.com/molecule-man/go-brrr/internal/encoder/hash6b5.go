// H6b5 uses 32 positions per bucket for quality 6 and large inputs.
//
// The encoder converts h6b5u to h6b5 when many entries have different tags.

package encoder

import (
	"math/bits"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// H6b5 configuration constants for quality 6.
const (
	h6b5BucketBits = 15
	h6b5BucketSize = 1 << h6b5BucketBits // 32768
	h6b5BlockBits  = 5
	h6b5BlockSize  = 1 << h6b5BlockBits // 32
	h6b5BlockMask  = h6b5BlockSize - 1
	h6b5HashShift  = 64 - h6b5BucketBits // 49

	// h6b5HashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h6b5HashTypeLength = 8

	// h6b5NumLastDistances is the number of distance cache entries to check.
	// For quality 6 (< 7), the C reference uses 4.
	h6b5NumLastDistances = 4

	// h6b5HashMul is the hash multiplier: kHashMul64 << (64 - 5*8).
	// Pre-computed because the untyped shift overflows Go constant arithmetic.
	h6b5HashMul uint64 = 0x7BD3579BD3000000
)

// h6b5 is the H6 hasher with blockBits=5: a forgetful hash table where each
// bucket holds a ring buffer of up to h6b5BlockSize (32) tagged positions.
type h6b5 struct {
	num    [h6b5BucketSize]uint16 // 0xFFFF minus entry count, per bucket
	blocks [h6b5BucketSize]h6b5Block
	// everWrapped disables the direct position path after the first ring wrap.
	everWrapped bool
	hasherCommon
	nextWarm uint32 //nolint:unused // generic scan loads the next bucket into the cache
}

// h6b5Block stores tags beside positions in each bucket.
type h6b5Block struct {
	tags [h6b5BlockSize]uint8  // low hash bits of each entry
	pos  [h6b5BlockSize]uint32 // position ring buffer
}

func (h *h6b5) common() *hasherCommon { return &h.hasherCommon }

// bucketAt returns the position ring buffer for key.
//
// key is always < h6b5BucketSize. The mask lets the compiler drop the
// bounds check. The fixed-size array pointer lets it prove that
// `i & h6b5BlockMask` is in range in the scan loops.
func (h *h6b5) bucketAt(key uint32) *[h6b5BlockSize]uint32 {
	return &h.blocks[key&(h6b5BucketSize-1)].pos
}

// hash computes a 15-bit bucket index from 8 bytes at data[i:i+8].
func (h *h6b5) hash(data []byte, i uint) uint32 {
	return uint32((loadU64LE(data, i) * h6b5HashMul) >> h6b5HashShift)
}

// hashTag computes the bucket key and the 8-bit tag of the entry at data[i:i+8].
func (h *h6b5) hashTag(data []byte, i uint) (uint32, uint8) {
	v := (loadU64LE(data, i) * h6b5HashMul) >> (h6b5HashShift - 8)
	return uint32(v >> 8), uint8(v)
}

// tagsAt returns the tags of the block for key. See bucketAt.
func (h *h6b5) tagsAt(key uint32) *[h6b5BlockSize]uint8 {
	return &h.blocks[key&(h6b5BucketSize-1)].tags
}

// reset marks all buckets as empty before use.
// When oneShot is true and the input is small, only the touched buckets
// are cleared (partial prepare). Otherwise all buckets are cleared.
func (h *h6b5) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h6b5BucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h.hash(data, i)
			h.num[key] = 0xFFFF
		}
	} else {
		for i := range h.num {
			h.num[i] = 0xFFFF
		}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the ring buffer for the 8-byte sequence at
// data[pos & mask].
func (h *h6b5) store(data []byte, mask, pos uint) {
	key, tag := h.hashTag(data, pos&mask)
	minorIx := h.num[key] & h6b5BlockMask
	h.num[key]--
	b := &h.blocks[key&(h6b5BucketSize-1)]
	b.pos[minorIx] = uint32(pos)
	b.tags[minorIx] = tag
}

// storeRange records positions [start, end) in the hash table.
func (h *h6b5) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.store(data, mask, i)
	}
}

func (h *h6b5) storeNoWrap(data []byte, pos uint) {
	key, tag := h.hashTag(data, pos)
	minorIx := h.num[key] & h6b5BlockMask
	h.num[key]--
	b := &h.blocks[key&(h6b5BucketSize-1)]
	b.pos[minorIx] = uint32(pos)
	b.tags[minorIx] = tag
}

func (h *h6b5) storeRangeNoWrap(data []byte, start, end uint) {
	for i := start; i < end; i++ {
		h.storeNoWrap(data, i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h6b5) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
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
//  2. Hash bucket scan: walk the up to 32 positions of the bucket whose
//     tag matches, newest first. Reject candidates with a 4-byte quick
//     comparison, accept length >= 4.
//
// The caller searches the static dictionary if neither phase finds a match.
func (h *h6b5) findLongestMatch(
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
	_ = data[ringBufferMask]

	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key, tag := h.hashTag(data, curMasked)

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

	// Phase 2: scan the bucket entries whose tag matches, newest first.
	// backward == 0 is impossible here: cur is stored after this scan.
	// Keep the scan last so no values remain live across the call.
	minPrev := cur - maxBackward
	h6b5FindInBucket(h, unsafe.Pointer(unsafe.SliceData(data)), uint(key)<<8|uint(tag),
		cur, curMasked, ringBufferMask, minPrev, maxLength, bestLen, bestScore,
		&h.blocks[h.hash(data, (cur+1)&ringBufferMask)], out)
}

// findLongestMatchSmallBuf is the generic version of findLongestMatch used
// when the ring buffer backing array is smaller than ringBufferMask+1.
func (h *h6b5) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache *[4]uint,
	cur, maxLength, maxBackward uint,
	out *hasherSearchResult,
) {
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len
	key, tag := h.hashTag(data, curMasked)
	bucket := h.bucketAt(key)
	// Bit j of matches: slot (head+j) has the same tag. Bit 0 is the newest entry.
	n := h.num[key]
	head := uint(n+1) & h6b5BlockMask
	tags := h.tagsAt(key)
	splat := uint64(tag) * 0x0101010101010101
	matches := bits.RotateLeft32(tagEqualMask8(loadU64LE(tags[:], 0)^splat)|
		tagEqualMask8(loadU64LE(tags[:], 8)^splat)<<8|
		tagEqualMask8(loadU64LE(tags[:], 16)^splat)<<16|
		tagEqualMask8(loadU64LE(tags[:], 24)^splat)<<24, -int(head))
	// Drop the slots that hold no entry yet.
	if stored := 0xFFFF - n; stored < h6b5BlockSize {
		matches &= uint32(1)<<stored - 1
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

	// Phase 2: scan the bucket entries whose tag matches, newest first.
	// backward == 0 is impossible here: cur is stored after this scan.
	curProbe := loadU32LE(data, curMasked+bestLen-3)
	for ; matches != 0; matches &= matches - 1 {
		prev := uint(bucket[(head+uint(bits.TrailingZeros32(matches)))&h6b5BlockMask])
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
	slot := uint(h.num[key]) & h6b5BlockMask
	bucket[slot] = uint32(cur)
	tags[slot] = tag
	h.num[key]--
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct (non-virtual) since the receiver is concrete.
//
// If this call and past calls stay within the ring, use the path without masks.
func (h *h6b5) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
}

// createBackwardReferencesNoWrap uses direct positions before the first ring wrap.
// All stored positions fit within the ring, so scans and stores omit masks.
func (h *h6b5) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
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
}

// findLongestMatchNoWrap searches before the first ring wrap.
// Current and stored positions fit within the ring, so the scan omits masks.
// The caller searches the static dictionary if this search finds no match.
func (h *h6b5) findLongestMatchNoWrap(
	data []byte,
	distCache *[4]uint,
	cur, maxLength, maxBackward uint,
	out *hasherSearchResult,
) {
	bestScore := out.score
	bestLen := out.len
	key, tag := h.hashTag(data, cur)

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

	// Phase 2: scan the bucket entries whose tag matches, newest first.
	// Keep the scan last so no values remain live across the call.
	minPrev := cur - maxBackward
	h6b5FindInBucket(h, unsafe.Pointer(unsafe.SliceData(data)), uint(key)<<8|uint(tag),
		cur, cur, ^uint(0), minPrev, maxLength, bestLen, bestScore,
		&h.blocks[h.hash(data, cur+1)], out)
}
