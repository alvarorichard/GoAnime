// H40 forgetful chain hasher for quality 5/6 with small windows (lgwin <= 16).
//
// H40 uses a linked-list chain structure with banked slot storage instead of
// H5's per-bucket ring buffers. A tiny hash table provides quick rejection
// for distance cache candidates. With 32K buckets (vs H5's 16K) and chain
// traversal via delta-linked slots, H40 trades memory layout for better
// match quality on small windows.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// H40 configuration constants.
const (
	h40BucketBits = 15
	h40BucketSize = 1 << h40BucketBits // 32768
	h40BankBits   = 16
	h40BankSize   = 1 << h40BankBits // 65536
	h40NumBanks   = 1
	h40HashShift  = 32 - h40BucketBits // 17

	h40NumLastDistances = 4

	// h40HashTypeLength is the minimum number of bytes needed to compute
	// the hash and verify a match (StoreLookahead in C).
	h40HashTypeLength = 4
)

// h40Slot is one node in the forgetful chain.
type h40Slot struct {
	delta uint16
	next  uint16
}

// h40PackedSlot stores h40Slot as one 32-bit value for h40's hot chain walk.
// The low 16 bits hold delta; the high 16 bits hold next.
type h40PackedSlot uint32

// h40 is the H40 forgetful chain hasher. Each bucket maps to a linked list
// of slots stored in a single bank of 64K entries.
type h40 struct {
	maxHops     uint                  // Q5=16, Q6=32
	addr        [h40BucketSize]uint32 // position at bucket head
	head        [h40BucketSize]uint16 // index of head slot in bank
	tinyHash    [65536]uint8          // quick rejection for distance cache
	slots       [h40BankSize]h40PackedSlot
	freeSlotIdx uint16 // monotonically increasing, wraps
	hasherCommon
}

func (h *h40) common() *hasherCommon { return &h.hasherCommon }

// hash computes a 15-bit bucket index from 4 bytes at data[i:i+4].
func (h *h40) hash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h40HashShift
}

// reset prepares the hasher for use. addr stores positions as one's complement
// (^uint32(pos)) so a zeroed slot decodes to 0xFFFFFFFF — far enough in the
// "future" that `cur - decoded` always exceeds maxBackward. That lets the
// full-sweep path use clear()/memclr instead of a scalar fill of 0xCCCCCCCC.
func (h *h40) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h40BucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			bucket := h.hash(data, i)
			h.addr[bucket] = 0
		}
	} else {
		clear(h.addr[:])
		h.head = [h40BucketSize]uint16{}
	}
	h.tinyHash = [65536]uint8{}
	h.freeSlotIdx = 0
	h.ready = true
}

// store records position ix in the chain for the 4-byte sequence at
// data[ix & mask]. Positions are stored as one's complement (^uint32(ix))
// so the cleared/zero-init state decodes as a sentinel — see reset.
func (h *h40) store(data []byte, mask, ix uint) {
	key := h.hash(data, ix&mask)
	idx := h.freeSlotIdx
	h.freeSlotIdx++
	delta := ix - uint(^h.addr[key])
	h.tinyHash[uint16(ix)] = uint8(key)
	if delta > 0xFFFF {
		delta = 0xFFFF
	}
	h.slots[idx] = h40PackedSlot(uint32(delta) | uint32(h.head[key])<<16)
	h.addr[key] = ^uint32(ix)
	h.head[key] = idx
}

func (h *h40) storeWithKey(key uint32, ix uint) {
	idx := h.freeSlotIdx
	h.freeSlotIdx++
	delta := ix - uint(^h.addr[key])
	h.tinyHash[uint16(ix)] = uint8(key)
	if delta > 0xFFFF {
		delta = 0xFFFF
	}
	h.slots[idx] = h40PackedSlot(uint32(delta) | uint32(h.head[key])<<16)
	h.addr[key] = ^uint32(ix)
	h.head[key] = idx
}

// storeRange records positions [start, end) in the hash table.
func (h *h40) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.storeWithKey(h.hash(data, i&mask), i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h40) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h40HashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// findLongestMatch searches for the best backward reference at position cur
// in the ring buffer, then stores cur in the hash table.
//
// The search has three phases:
//  1. Distance cache: try 4 entries, use tinyHash for i>0 rejection,
//     accept length >= 2 for all entries.
//  2. Chain walk: traverse slot chain up to maxHops, 4-byte quick reject,
//     accept length >= 4.
//  3. Static dictionary fallback.
func (h *h40) findLongestMatch(
	data []byte, ringBufferMask uint,
	distCache []int,
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
	minScore := out.score
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, curMasked)
	tinyHash := uint8(key)
	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances. The first cache entry is hot and does
	// not need the tinyHash rejection used by the other entries.
	{
		backward := uint(distCache[0])
		prevIx := cur - backward
		if prevIx < cur && backward <= maxBackward {
			prevIx &= ringBufferMask
			if loadByte(data, prevIx) == loadByte(data, curMasked) &&
				loadByte(data, prevIx+1) == loadByte(data, curMasked+1) {
				ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
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
	}

	{
		backward := uint(distCache[1])
		prevIx := cur - backward
		if h.tinyHash[uint16(prevIx)] == tinyHash {
			if prevIx < cur && backward <= maxBackward {
				prevIx &= ringBufferMask
				if loadByte(data, prevIx) == loadByte(data, curMasked) &&
					loadByte(data, prevIx+1) == loadByte(data, curMasked+1) {
					ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
					if ml >= 2 {
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
		}
	}
	{
		backward := uint(distCache[2])
		prevIx := cur - backward
		if h.tinyHash[uint16(prevIx)] == tinyHash {
			if prevIx < cur && backward <= maxBackward {
				prevIx &= ringBufferMask
				if loadByte(data, prevIx) == loadByte(data, curMasked) &&
					loadByte(data, prevIx+1) == loadByte(data, curMasked+1) {
					ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
					if ml >= 2 {
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
		}
	}
	{
		backward := uint(distCache[3])
		prevIx := cur - backward
		if h.tinyHash[uint16(prevIx)] == tinyHash {
			if prevIx < cur && backward <= maxBackward {
				prevIx &= ringBufferMask
				if loadByte(data, prevIx) == loadByte(data, curMasked) &&
					loadByte(data, prevIx+1) == loadByte(data, curMasked+1) {
					ml := uint(matchLenAtNoInline(data, prevIx, curMasked, int(maxLength)))
					if ml >= 2 {
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
		}
	}

	// Raise bestLen floor to 3 so phase 2 only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: walk the chain.
	// Capture the old chain head/addr, then store cur *before* the walk so
	// storeWithKey's writes to addr/head/slots/tinyHash pipeline against the
	// chain's serial slot loads. The walk still traverses the old chain — the
	// newly-stored slot is only reachable via the new head, which we don't use.
	{
		oldAddr := uint(^h.addr[key])
		oldHead := h.head[key]

		newIdx := h.freeSlotIdx
		h.freeSlotIdx++
		storeDelta := cur - oldAddr
		h.tinyHash[uint16(cur)] = uint8(key)
		if storeDelta > 0xFFFF {
			storeDelta = 0xFFFF
		}
		h.slots[newIdx] = h40PackedSlot(uint32(storeDelta) | uint32(oldHead)<<16)
		h.addr[key] = ^uint32(cur)
		h.head[key] = newIdx

		backward := uint(0)
		hops := h.maxHops
		delta := cur - oldAddr
		slot := oldHead
		for hops > 0 {
			hops--
			backward += delta
			if backward > maxBackward {
				break
			}
			prevIx := (cur - backward) & ringBufferMask
			slotEntry := uint32(h.slots[uint(slot)])
			slot = uint16(slotEntry >> 16)
			delta = uint(uint16(slotEntry))
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
	}

	// Phase 3: static dictionary fallback when no match was found.
	if out.score == minScore {
		searchStaticDictionaryDeep(data[curMasked:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}

// findLongestMatchSmallBuf is the generic version of findLongestMatch used
// when the ring buffer backing array is smaller than ringBufferMask+1 (i.e.
// the first small write hasn't triggered a full allocation yet). It keeps
// all runtime bounds checks and is only called for small initial payloads.
func (h *h40) findLongestMatchSmallBuf(
	data []byte, ringBufferMask uint,
	distCache []int,
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
	for i := range uint(h40NumLastDistances) {
		backward := uint(distCache[i])
		prevIx := cur - backward
		if i > 0 && h.tinyHash[uint16(prevIx)] != tinyHash {
			continue
		}
		if prevIx >= cur || backward > maxBackward {
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
		bank := key & (h40NumBanks - 1)
		backward := uint(0)
		hops := h.maxHops
		delta := cur - uint(^h.addr[key])
		slot := h.head[key]
		slotBase := uint(bank) * h40BankSize
		for hops > 0 {
			hops--
			backward += delta
			if backward > maxBackward {
				break
			}
			prevIx := (cur - backward) & ringBufferMask
			slotEntry := uint32(h.slots[slotBase+uint(slot)])
			slot = uint16(slotEntry >> 16)
			delta = uint(uint16(slotEntry))
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
		h.storeWithKey(key, cur)
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
func (h *h40) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h40HashTypeLength {
		storeEnd = posEnd - h40HashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 10 derived entries.
	var distCache [16]int
	for i, d := range s.distCache {
		distCache[i] = int(d)
	}
	prepareDistanceCache(distCache[:])

	for position+h40HashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		h.findLongestMatch(data, mask, distCache[:],
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

				h.findLongestMatch(data, mask, distCache[:],
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
						position+h40HashTypeLength < posEnd {
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
					posJump := min(position+16, posEnd-max(h40HashTypeLength-1, 4))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h40HashTypeLength-1))
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
