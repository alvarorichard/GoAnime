// H42 forgetful chain hasher for quality 9 with small windows (lgwin <= 16).
//
// H42 shares the same chain algorithm as H40/H41 but uses 512 banks of 512
// slots each (vs H41's single bank of 64K slots). This distributes chains
// across more independent storage areas for better cache locality. The
// distance cache search is expanded to 16 entries (4 base plus 6 derived
// near-miss entries for each of dist[0] and dist[1]).

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// H42 configuration constants.
const (
	h42BucketBits = 15
	h42BucketSize = 1 << h42BucketBits // 32768
	h42BankBits   = 9
	h42BankSize   = 1 << h42BankBits // 512
	h42NumBanks   = 512
	h42HashShift  = 32 - h42BucketBits // 17

	// h42NumLastDistances is the number of distance cache entries to check.
	// For quality 9, the C reference uses 16 (the 4 base entries plus 6
	// derived near-miss entries for each of dist[0] and dist[1]).
	h42NumLastDistances = 16

	// h42HashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h42HashTypeLength = 4
)

// h42 is the H42 forgetful chain hasher. Each bucket maps to a linked list
// of slots distributed across 512 banks of 512 entries each.
type h42 struct {
	addr        [h42BucketSize]uint32             // position at bucket head
	head        [h42BucketSize]uint16             // index of head slot in bank
	tinyHash    [65536]uint8                      // quick rejection for distance cache
	banks       [h42NumBanks][h42BankSize]h40Slot // 512 banks × 512 slots
	freeSlotIdx [h42NumBanks]uint16               // per-bank monotonically increasing index
	hasherCommon
}

func (h *h42) common() *hasherCommon { return &h.hasherCommon }

// hash computes a 15-bit bucket index from 4 bytes at data[i:i+4].
func (h *h42) hash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h42HashShift
}

// reset prepares the hasher for use. addr stores positions as one's complement
// (^uint32(pos)) so a zeroed slot decodes to 0xFFFFFFFF — far enough in the
// "future" that `cur - decoded` always exceeds maxBackward. That lets the
// full-sweep path use clear()/memclr instead of a scalar fill of 0xCCCCCCCC,
// which dominated reset on the q=9 small-window benchmarks.
func (h *h42) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h42BucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			bucket := h.hash(data, i)
			h.addr[bucket] = 0
		}
	} else {
		clear(h.addr[:])
		h.head = [h42BucketSize]uint16{}
	}
	h.tinyHash = [65536]uint8{}
	h.freeSlotIdx = [h42NumBanks]uint16{}
	h.ready = true
}

// store records position ix in the chain for the 4-byte sequence at
// data[ix & mask]. Positions are stored as one's complement (^uint32(ix))
// so the cleared/zero-init state decodes as a sentinel — see reset.
func (h *h42) store(data []byte, mask, ix uint) {
	key := h.hash(data, ix&mask)
	bank := key & (h42NumBanks - 1)
	idx := h.freeSlotIdx[bank] & (h42BankSize - 1)
	h.freeSlotIdx[bank]++
	delta := ix - uint(^h.addr[key])
	h.tinyHash[uint16(ix)] = uint8(key)
	if delta > 0xFFFF {
		delta = 0xFFFF
	}
	h.banks[bank][idx].delta = uint16(delta)
	h.banks[bank][idx].next = h.head[key]
	h.addr[key] = ^uint32(ix)
	h.head[key] = idx
}

// storeRange records positions [start, end) in the hash table.
func (h *h42) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		key := h.hash(data, i&mask)
		bank := key & (h42NumBanks - 1)
		idx := h.freeSlotIdx[bank] & (h42BankSize - 1)
		h.freeSlotIdx[bank]++
		delta := i - uint(^h.addr[key])
		h.tinyHash[uint16(i)] = uint8(key)
		if delta > 0xFFFF {
			delta = 0xFFFF
		}
		h.banks[bank][idx].delta = uint16(delta)
		h.banks[bank][idx].next = h.head[key]
		h.addr[key] = ^uint32(i)
		h.head[key] = idx
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h42) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h42HashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch searches for the best backward reference at position cur
// in the ring buffer, then stores cur in the hash table.
//
// The search has three phases:
//  1. Distance cache: try 16 entries (4 base + 6 derived near-miss for each
//     of dist[0] and dist[1]), use tinyHash for i>0 rejection, accept
//     length >= 2 for all entries.
//  2. Chain walk: traverse slot chain up to 224 hops, 4-byte quick reject,
//     accept length >= 4.
//  3. Static dictionary fallback.
func (h *h42) findLongestMatch(
	data []byte, ringBufferMask uint,
	distCache *[16]int,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	curMasked := cur & ringBufferMask
	minScore := out.score
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)
	tinyHash := uint8(key)
	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	for i := range uint(h42NumLastDistances) {
		backward := uint(distCache[i])
		if backward > maxBackward {
			continue
		}
		prevIx := cur - backward
		if i > 0 && h.tinyHash[uint16(prevIx)] != tinyHash {
			continue
		}
		if prevIx >= cur {
			continue
		}
		prevIx &= ringBufferMask

		ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
		if ml >= 2 {
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

	// Phase 2: walk the chain.
	{
		bank := key & (h42NumBanks - 1)
		backward := uint(0)
		hops := uint(224) // 7 << (9 - 4) for quality 9
		delta := cur - uint(^h.addr[key])
		slot := h.head[key]
		for hops > 0 {
			hops--
			backward += delta
			if backward > maxBackward {
				break
			}
			prevIx := (cur - backward) & ringBufferMask
			nextSlot := h.banks[bank][slot].next
			nextDelta := uint(h.banks[bank][slot].delta)
			slot = nextSlot
			delta = nextDelta
			if curMasked+bestLen > ringBufferMask ||
				prevIx+bestLen > ringBufferMask ||
				loadU32LE(data, curMasked+bestLen-3) != loadU32LE(data, prevIx+bestLen-3) {
				continue
			}

			ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
			if ml >= 4 {
				score := backwardReferenceScore(ml, backward)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
				}
			}
		}
		h.store(data, ringBufferMask, cur)
	}

	// Phase 3: static dictionary fallback when no match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct (non-virtual) since the receiver is concrete.
func (h *h42) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h42HashTypeLength {
		storeEnd = posEnd - h42HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 512
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 16 derived entries.
	var distCache [16]int
	for i, d := range s.distCache {
		distCache[i] = int(d)
	}
	prepareDistanceCache(distCache[:])

	for position+h42HashTypeLength < posEnd {
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
						position+h42HashTypeLength < posEnd {
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

			for i, d := range s.distCache {
				distCache[i] = int(d)
			}
			prepareDistanceCache(distCache[:])
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-max(h42HashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h42HashTypeLength-1))
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
