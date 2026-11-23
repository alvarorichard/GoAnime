// Hash table for the streaming encoder (quality 4, large inputs).
// BUCKET_BITS=20, BUCKET_SWEEP=4, HASH_LEN=7, USE_DICTIONARY=0.

package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// H54-specific configuration constants.
const (
	h54BucketBits     = 20
	h54BucketSize     = 1 << h54BucketBits // 1048576
	h54BucketMask     = h54BucketSize - 1
	h54BucketSweep    = 4
	h54BucketSweepMsk = (h54BucketSweep - 1) << 3 // 24
	h54HashLen        = 7
)

// h54 is the H54 hasher: a forgetful hash table that maps 7-byte sequences
// to positions. Each logical bucket stores four uint32 positions spread
// across adjacent slots (BUCKET_SWEEP=4). USE_DICTIONARY=0.
type h54 struct {
	buckets     [h54BucketSize]uint32
	everWrapped bool
	hasherCommon
}

func (h *h54) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
func (h *h54) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h54BucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h.hash(data, i)
			for j := range uint32(h54BucketSweep) {
				h.buckets[(key+(j<<3))&h54BucketMask] = 0
			}
		}
	} else {
		h.buckets = [h54BucketSize]uint32{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the bucket for the 7-byte sequence at
// data[pos & mask]. Uses the sweep offset to distribute entries.
func (h *h54) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	off := uint32(pos) & h54BucketSweepMsk
	h.buckets[(key+off)&h54BucketMask] = uint32(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h54) storeRange(data []byte, mask, start, end uint) {
	buckets := &h.buckets
	for i := start; i < end; i++ {
		key := h.hash(data, i&mask)
		off := uint32(i) & h54BucketSweepMsk
		buckets[(key+off)&h54BucketMask] = uint32(i)
	}
}

func (h *h54) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	off := uint32(pos) & h54BucketSweepMsk
	h.buckets[(key+off)&h54BucketMask] = uint32(pos)
}

func (h *h54) storeRangeNoWrap(data []byte, start, end uint) {
	buckets := &h.buckets
	for i := start; i < end; i++ {
		key := h.hash(data, i)
		off := uint32(i) & h54BucketSweepMsk
		buckets[(key+off)&h54BucketMask] = uint32(i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h54) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= hashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store/storeRange calls
// are direct (non-virtual) since the receiver is concrete.
//
// H54 uses BUCKET_SWEEP=4 and USE_DICTIONARY=0, so no static dictionary
// lookups are performed.
func (h *h54) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	mask := uint(s.mask)
	if !h.everWrapped && uint(wrappedPos)+uint(bytes) <= mask+1 {
		h.createBackwardReferencesNoWrap(s, bytes, wrappedPos)
		return
	}
	h.everWrapped = true
	data := s.data
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize

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

		// Manually inlined findLongestMatch — the compiler cannot inline it
		// (cost 1096 > budget 80), so we do it here to eliminate call overhead.
		{
			lastDistance := s.distCache[0]
			curMasked := position & mask
			guardByte := loadByte(data, curMasked) // sr.len == 0 → bestLenIn == 0
			key := h.hash(data, curMasked)
			bestScore := sr.score // minScore, type uint
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & h54BucketMask
			hkey2 := (key + 16) & h54BucketMask
			hkey3 := (key + 24) & h54BucketMask

			keyOut := (key + uint32(position&h54BucketSweepMsk)) & h54BucketMask

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])
			prev2 := uint(buckets[hkey2])
			prev3 := uint(buckets[hkey3])

			limit := int(maxLength)

			{
				prev := position - lastDistance
				if prev < position {
					prev &= mask
					if guardByte == loadByte(data, prev+bestLen) {
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

			// Entry 0.
			{
				backward := position - prev0
				prev0 &= mask
				if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
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

			// Entry 1.
			{
				backward := position - prev1
				prev1 &= mask
				if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
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

			// Entry 2.
			{
				backward := position - prev2
				prev2 &= mask
				if guardByte == loadByte(data, prev2+bestLen) && backward != 0 && backward <= maxDistance {
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

			// Entry 3.
			{
				backward := position - prev3
				prev3 &= mask
				if guardByte == loadByte(data, prev3+bestLen) && backward != 0 && backward <= maxDistance {
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

			buckets[keyOut] = uint32(position)
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

				// Manually inlined findLongestMatch for lazy match at position+1.
				{
					cur2 := position + 1
					lastDistance := s.distCache[0]
					curMasked := cur2 & mask
					bestLen := sr2.len
					guardByte := loadByte(data, curMasked+bestLen)
					key := h.hash(data, curMasked)
					bestScore := sr2.score // minScore

					hkey0 := key
					hkey1 := (key + 8) & h54BucketMask
					hkey2 := (key + 16) & h54BucketMask
					hkey3 := (key + 24) & h54BucketMask

					keyOut := (key + uint32(cur2&h54BucketSweepMsk)) & h54BucketMask

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])
					prev2 := uint(buckets[hkey2])
					prev3 := uint(buckets[hkey3])

					limit := int(maxLength)

					{
						prev := cur2 - lastDistance
						if prev < cur2 {
							prev &= mask
							if guardByte == loadByte(data, prev+bestLen) {
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

					// Entry 0.
					{
						backward := cur2 - prev0
						prev0 &= mask
						if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
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

					// Entry 1.
					{
						backward := cur2 - prev1
						prev1 &= mask
						if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
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

					// Entry 2.
					{
						backward := cur2 - prev2
						prev2 &= mask
						if guardByte == loadByte(data, prev2+bestLen) && backward != 0 && backward <= maxDistance {
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

					// Entry 3.
					{
						backward := cur2 - prev3
						prev3 &= mask
						if guardByte == loadByte(data, prev3+bestLen) && backward != 0 && backward <= maxDistance {
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

					buckets[keyOut] = uint32(cur2)
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

			// Recompute maxDistance after the lazy loop because position may
			// have advanced. This matches the C reference's dictionary_start.
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

func (h *h54) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize

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
			curWord := loadU64LE(data, position)
			guardByte := byte(curWord)
			key := uint32(((curWord << (64 - 8*h54HashLen)) * hashMul64) >> (64 - h54BucketBits))
			bestScore := sr.score
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & h54BucketMask
			hkey2 := (key + 16) & h54BucketMask
			hkey3 := (key + 24) & h54BucketMask

			keyOut := (key + uint32(position&h54BucketSweepMsk)) & h54BucketMask

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])
			prev2 := uint(buckets[hkey2])
			prev3 := uint(buckets[hkey3])

			limit := int(maxLength)

			{
				prev := position - lastDistance
				if prev < position {
					if guardByte == loadByte(data, prev+bestLen) {
						length := 0
						xor := loadU64LE(data, prev) ^ curWord
						if xor != 0 {
							length = bits.TrailingZeros64(xor) / 8
						} else {
							length = 8 + matchLenAt(data, prev+8, position+8, limit-8)
						}
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
				if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev0) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev0+8, position+8, limit-8)
					}
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
				if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev1) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev1+8, position+8, limit-8)
					}
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
				if guardByte == loadByte(data, prev2+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev2) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev2+8, position+8, limit-8)
					}
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
				if guardByte == loadByte(data, prev3+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev3) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev3+8, position+8, limit-8)
					}
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

			buckets[keyOut] = uint32(position)
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
					curWord := loadU64LE(data, cur2)
					key := uint32(((curWord << (64 - 8*h54HashLen)) * hashMul64) >> (64 - h54BucketBits))
					bestScore := sr2.score

					hkey0 := key
					hkey1 := (key + 8) & h54BucketMask
					hkey2 := (key + 16) & h54BucketMask
					hkey3 := (key + 24) & h54BucketMask

					keyOut := (key + uint32(cur2&h54BucketSweepMsk)) & h54BucketMask

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])
					prev2 := uint(buckets[hkey2])
					prev3 := uint(buckets[hkey3])

					limit := int(maxLength)

					{
						prev := cur2 - lastDistance
						if prev < cur2 {
							if guardByte == loadByte(data, prev+bestLen) {
								length := 0
								xor := loadU64LE(data, prev) ^ curWord
								if xor != 0 {
									length = bits.TrailingZeros64(xor) / 8
								} else {
									length = 8 + matchLenAt(data, prev+8, cur2+8, limit-8)
								}
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
						if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
							length := 0
							xor := loadU64LE(data, prev0) ^ curWord
							if xor != 0 {
								length = bits.TrailingZeros64(xor) / 8
							} else {
								length = 8 + matchLenAt(data, prev0+8, cur2+8, limit-8)
							}
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
						if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
							length := 0
							xor := loadU64LE(data, prev1) ^ curWord
							if xor != 0 {
								length = bits.TrailingZeros64(xor) / 8
							} else {
								length = 8 + matchLenAt(data, prev1+8, cur2+8, limit-8)
							}
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
						if guardByte == loadByte(data, prev2+bestLen) && backward != 0 && backward <= maxDistance {
							length := 0
							xor := loadU64LE(data, prev2) ^ curWord
							if xor != 0 {
								length = bits.TrailingZeros64(xor) / 8
							} else {
								length = 8 + matchLenAt(data, prev2+8, cur2+8, limit-8)
							}
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
						if guardByte == loadByte(data, prev3+bestLen) && backward != 0 && backward <= maxDistance {
							length := 0
							xor := loadU64LE(data, prev3) ^ curWord
							if xor != 0 {
								length = bits.TrailingZeros64(xor) / 8
							} else {
								length = 8 + matchLenAt(data, prev3+8, cur2+8, limit-8)
							}
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

					buckets[keyOut] = uint32(cur2)
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

// hash computes the h54 20-bit hash from 7 bytes at data[off:off+8]
// using an unsafe load, avoiding sub-slice creation and bounds checks.
func (*h54) hash(data []byte, off uint) uint32 {
	v := loadU64LE(data, off)
	x := (v << (64 - 8*h54HashLen)) * hashMul64
	return uint32(x >> (64 - h54BucketBits))
}
