// H4 variant for inputs bounded to 64 KiB.
// Stores uint16 positions in buckets, halving the table from 512 KB to 256 KB.
// The dispatch contract guarantees position never exceeds 2^16, so
// `uint16(pos)` is lossless — no aliasing, no stale-probe overhead. If buffered
// input crosses the 64 KiB boundary mid-stream (when sizeHint underestimated
// the real total), encoderCore.maybePromoteHasher transparently swaps to h4.

package encoder

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// h4u16 is the H4 hasher with uint16 bucket slots, dispatched only when the
// encoder expects the input to fit in 64 KiB (either via a user-supplied
// sizeHint or via the one-shot isLast guarantee). Within that bound every
// stored position is in [0, 65535], so `uint16(pos)` storage is lossless and
// `position - prev` (uint subtraction) is directly the real backward distance
// — no modular truncation needed. Semantics match h4 at half the memory; the
// lookup arithmetic is the same shape.
type h4u16 struct {
	buckets [h4BucketSize]uint16
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions reaching or exceeding mask+1, after which the no-wrap
	// fast path is disabled because stored bucket values may then encode
	// positions outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

func (h *h4u16) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
//
// The partial path views buckets through a uint32 alias and zeroes pairs of
// adjacent uint16 slots per store. This avoids the length-changing-prefix
// (66h) predecode stalls that `MOVW $0, mem` triggers when fetched from
// MITE, and is exact for the four sweep slots after rounding the hash key
// down to even — clearing the partner slot in each pair is harmless: it
// would be either stale (and is meant to be cleared) or already zero.
func (h *h4u16) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h4BucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		const mask32 = h4BucketSize/2 - 1
		buckets32 := (*[h4BucketSize / 2]uint32)(unsafe.Pointer(&h.buckets))
		for i := range inputSize {
			k := h.hash(data, i) >> 1
			buckets32[k] = 0
			buckets32[(k+4)&mask32] = 0
			buckets32[(k+8)&mask32] = 0
			buckets32[(k+12)&mask32] = 0
		}
	} else {
		h.buckets = [h4BucketSize]uint16{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the bucket for the 5-byte sequence at
// data[pos & mask]. Uses the sweep offset to distribute entries.
func (h *h4u16) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	off := uint32(pos) & h4BucketSweepMsk
	h.buckets[(key+off)&h4BucketMask] = uint16(pos)
}

func (h *h4u16) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	off := uint32(pos) & h4BucketSweepMsk
	h.buckets[(key+off)&h4BucketMask] = uint16(pos)
}

func (h *h4u16) storeRangeNoWrap(data []byte, start, end uint) {
	buckets := &h.buckets
	for i := start; i < end; i++ {
		key := h.hash(data, i)
		off := uint32(i) & h4BucketSweepMsk
		buckets[(key+off)&h4BucketMask] = uint16(i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h4u16) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= hashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store calls are direct
// (non-virtual) since the receiver is concrete.
//
// H4 uses BUCKET_SWEEP=4 and USE_DICTIONARY=1, so static dictionary lookups
// are performed when no hash-table match is found.
//
// When the call's [wrappedPos, wrappedPos+bytes) range fits entirely within
// the ring buffer (no modular wrap) and no past call has wrapped, dispatch
// to createBackwardReferencesNoWrap which omits the per-iteration & mask
// ops (each redundant when stored bucket values are < mask+1).
func (h *h4u16) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	mask := uint(s.mask)
	if !h.everWrapped && uint(wrappedPos)+uint(bytes) <= mask+1 {
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
	if uint(bytes) >= hashTypeLength {
		storeEnd = posEnd - hashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))
	buckets := &h.buckets

	for position+hashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.len = 0
		sr.lenCodeDelta = 0
		sr.distance = 0
		sr.score = minScore

		{
			lastDistance := s.distCache[0]
			curMasked := position & mask
			guardByte := loadByte(data, curMasked)
			key := h.hash(data, curMasked)
			bestScore := sr.score
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & h4BucketMask
			hkey2 := (key + 16) & h4BucketMask
			hkey3 := (key + 24) & h4BucketMask

			keyOut := (key + uint32(position&h4BucketSweepMsk)) & h4BucketMask

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])
			prev2 := uint(buckets[hkey2])
			prev3 := uint(buckets[hkey3])

			limit := int(maxLength)

			{
				prev := position - lastDistance
				if prev < position && lastDistance <= maxDistance {
					prev &= mask
					if loadByte(data, prev+bestLen) == guardByte {
						length := matchLenAt(data, prev, curMasked, limit)
						if length >= 4 {
							score := backwardReferenceScoreUsingLastDistance(uint(length))
							if bestScore < score {
								bestLen = uint(length)
								sr.len = bestLen
								sr.distance = lastDistance
								sr.score = score
								bestScore = score
								guardByte = loadByte(data, curMasked+bestLen)
							}
						}
					}
				}
			}

			{
				backward := position - prev0
				prev0 &= mask
				if loadByte(data, prev0+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev0, curMasked, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, curMasked+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev1
				prev1 &= mask
				if loadByte(data, prev1+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev1, curMasked, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, curMasked+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev2
				prev2 &= mask
				if loadByte(data, prev2+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev2, curMasked, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, curMasked+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev3
				prev3 &= mask
				if loadByte(data, prev3+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev3, curMasked, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							sr.len = uint(length)
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			buckets[keyOut] = uint16(position)
		}

		if sr.score == minScore {
			posM := position & mask
			if wl, wi, ok := searchStaticDictAt(data, posM, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok {
				if matchStaticDictEntryAt(data, posM, wl, wi, maxDistance+gap, sr.score, &sr) {
					s.dictNumMatches++
				}
			}
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
				sr2.len = min(sr.len-1, maxLength)
				sr2.lenCodeDelta = 0
				sr2.distance = 0
				sr2.score = minScore
				maxDistance = min(position+1, maxBackwardLimit)

				{
					cur2 := position + 1
					lastDistance := s.distCache[0]
					curMasked := cur2 & mask
					bestLen := sr2.len
					guardByte := loadByte(data, curMasked+bestLen)
					key := h.hash(data, curMasked)
					bestScore := sr2.score

					hkey0 := key
					hkey1 := (key + 8) & h4BucketMask
					hkey2 := (key + 16) & h4BucketMask
					hkey3 := (key + 24) & h4BucketMask

					keyOut := (key + uint32(cur2&h4BucketSweepMsk)) & h4BucketMask

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])
					prev2 := uint(buckets[hkey2])
					prev3 := uint(buckets[hkey3])

					limit := int(maxLength)

					{
						prev := cur2 - lastDistance
						if prev < cur2 && lastDistance <= maxDistance {
							prev &= mask
							if loadByte(data, prev+bestLen) == guardByte {
								length := matchLenAt(data, prev, curMasked, limit)
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										bestLen = uint(length)
										sr2.len = bestLen
										sr2.distance = lastDistance
										sr2.score = score
										bestScore = score
										guardByte = loadByte(data, curMasked+bestLen)
									}
								}
							}
						}
					}

					{
						backward := cur2 - prev0
						prev0 &= mask
						if loadByte(data, prev0+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev0, curMasked, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, curMasked+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev1
						prev1 &= mask
						if loadByte(data, prev1+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev1, curMasked, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, curMasked+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev2
						prev2 &= mask
						if loadByte(data, prev2+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev2, curMasked, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, curMasked+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev3
						prev3 &= mask
						if loadByte(data, prev3+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev3, curMasked, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									sr2.len = uint(length)
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					buckets[keyOut] = uint16(cur2)
				}

				if sr2.score == minScore {
					posM2 := (position + 1) & mask
					if wl2, wi2, ok2 := searchStaticDictAt(data, posM2, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok2 {
						if matchStaticDictEntryAt(data, posM2, wl2, wi2, maxDistance+gap, sr2.score, &sr2) {
							s.dictNumMatches++
						}
					}
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
						position+hashTypeLength < posEnd {
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

			delta := uint32(uint8(int8(sr.lenCodeDelta)))
			distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
			effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
			insCode := getInsertLenCode(insertLength)
			copyCode := getCopyLenCode(effectiveCopyLen)
			cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
			s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
			s.numLiterals += insertLength
			insertLength = 0

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			for i := rangeStart; i < rangeEnd; i++ {
				key := h.hash(data, i&mask)
				off := uint32(i) & h4BucketSweepMsk
				buckets[(key+off)&h4BucketMask] = uint16(i)
			}

			position += sr.len
		} else {
			insertLength++
			position++

			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					posJump := min(position+16, posEnd-(hashTypeLength-1))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(hashTypeLength-1))
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
// `position & mask` ops are redundant and elided here. The paired no-wrap
// store helpers also skip redundant position masking while preserving the
// same bucket update.
func (h *h4u16) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= hashTypeLength {
		storeEnd = posEnd - hashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))
	buckets := &h.buckets

	for position+hashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.len = 0
		sr.lenCodeDelta = 0
		sr.distance = 0
		sr.score = minScore

		{
			lastDistance := s.distCache[0]
			guardByte := loadByte(data, position)
			key := h.hash(data, position)
			bestScore := sr.score
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & h4BucketMask
			hkey2 := (key + 16) & h4BucketMask
			hkey3 := (key + 24) & h4BucketMask

			keyOut := (key + uint32(position&h4BucketSweepMsk)) & h4BucketMask

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])
			prev2 := uint(buckets[hkey2])
			prev3 := uint(buckets[hkey3])

			limit := int(maxLength)

			{
				prev := position - lastDistance
				if prev < position && lastDistance <= maxDistance {
					if loadByte(data, prev+bestLen) == guardByte {
						length := matchLenAt(data, prev, position, limit)
						if length >= 4 {
							score := backwardReferenceScoreUsingLastDistance(uint(length))
							if bestScore < score {
								bestLen = uint(length)
								sr.len = bestLen
								sr.distance = lastDistance
								sr.score = score
								bestScore = score
								guardByte = loadByte(data, position+bestLen)
							}
						}
					}
				}
			}

			{
				backward := position - prev0
				if loadByte(data, prev0+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev0, position, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, position+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev1
				if loadByte(data, prev1+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev1, position, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, position+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev2
				if loadByte(data, prev2+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev2, position, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							bestLen = uint(length)
							sr.len = bestLen
							guardByte = loadByte(data, position+bestLen)
							bestScore = score
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			{
				backward := position - prev3
				if loadByte(data, prev3+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev3, position, limit)
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							sr.len = uint(length)
							sr.score = score
							sr.distance = backward
						}
					}
				}
			}

			buckets[keyOut] = uint16(position)
		}

		if sr.score == minScore {
			if wl, wi, ok := searchStaticDictAt(data, position, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok {
				if matchStaticDictEntryAt(data, position, wl, wi, maxDistance+gap, sr.score, &sr) {
					s.dictNumMatches++
				}
			}
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
				sr2.len = min(sr.len-1, maxLength)
				sr2.lenCodeDelta = 0
				sr2.distance = 0
				sr2.score = minScore
				maxDistance = min(position+1, maxBackwardLimit)

				{
					cur2 := position + 1
					lastDistance := s.distCache[0]
					bestLen := sr2.len
					guardByte := loadByte(data, cur2+bestLen)
					key := h.hash(data, cur2)
					bestScore := sr2.score

					hkey0 := key
					hkey1 := (key + 8) & h4BucketMask
					hkey2 := (key + 16) & h4BucketMask
					hkey3 := (key + 24) & h4BucketMask

					keyOut := (key + uint32(cur2&h4BucketSweepMsk)) & h4BucketMask

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])
					prev2 := uint(buckets[hkey2])
					prev3 := uint(buckets[hkey3])

					limit := int(maxLength)

					{
						prev := cur2 - lastDistance
						if prev < cur2 && lastDistance <= maxDistance {
							if loadByte(data, prev+bestLen) == guardByte {
								length := matchLenAt(data, prev, cur2, limit)
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										bestLen = uint(length)
										sr2.len = bestLen
										sr2.distance = lastDistance
										sr2.score = score
										bestScore = score
										guardByte = loadByte(data, cur2+bestLen)
									}
								}
							}
						}
					}

					{
						backward := cur2 - prev0
						if loadByte(data, prev0+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev0, cur2, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, cur2+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev1
						if loadByte(data, prev1+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev1, cur2, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, cur2+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev2
						if loadByte(data, prev2+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev2, cur2, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen = uint(length)
									sr2.len = bestLen
									guardByte = loadByte(data, cur2+bestLen)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev3
						if loadByte(data, prev3+bestLen) == guardByte && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev3, cur2, limit)
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									sr2.len = uint(length)
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					buckets[keyOut] = uint16(cur2)
				}

				if sr2.score == minScore {
					if wl2, wi2, ok2 := searchStaticDictAt(data, position+1, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok2 {
						if matchStaticDictEntryAt(data, position+1, wl2, wi2, maxDistance+gap, sr2.score, &sr2) {
							s.dictNumMatches++
						}
					}
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
						position+hashTypeLength < posEnd {
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

			delta := uint32(uint8(int8(sr.lenCodeDelta)))
			distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
			effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
			insCode := getInsertLenCode(insertLength)
			copyCode := getCopyLenCode(effectiveCopyLen)
			cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
			s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
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
					posJump := min(position+16, posEnd-(hashTypeLength-1))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(hashTypeLength-1))
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

// hash computes the h4 17-bit hash from 5 bytes at data[off:off+8]
// using an unsafe load, avoiding sub-slice creation and bounds checks.
func (*h4u16) hash(data []byte, off uint) uint32 {
	v := loadU64LE(data, off)
	h := (v << (64 - 8*hashLen)) * hashMul64
	return uint32(h >> (64 - h4BucketBits))
}
