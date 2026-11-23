// Hash table for the streaming encoder (quality 4).
// BUCKET_BITS=17, BUCKET_SWEEP=4, HASH_LEN=5, USE_DICTIONARY=1.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// H4-specific configuration constants.
const (
	h4BucketBits     = 17
	h4BucketSize     = 1 << h4BucketBits // 131072
	h4BucketMask     = h4BucketSize - 1
	h4BucketSweep    = 4
	h4BucketSweepMsk = (h4BucketSweep - 1) << 3 // 24
)

// h4 is the H4 hasher: a forgetful hash table that maps 5-byte sequences
// to positions. Each logical bucket stores four uint32 positions spread
// across adjacent slots (BUCKET_SWEEP=4). USE_DICTIONARY=1.
type h4 struct {
	buckets [h4BucketSize]uint32
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions exceeding mask+1, after which the no-wrap fast path is
	// disabled because stored bucket values may then encode positions
	// outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

func (h *h4) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
func (h *h4) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h4BucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		buckets := &h.buckets
		for i := range inputSize {
			key := h.hash(data, i)
			// Unrolled BUCKET_SWEEP=4 loop. All indices are masked by h4BucketMask
			// (< h4BucketSize = len(buckets)), so no bounds checks are needed.
			buckets[key] = 0
			buckets[(key+8)&h4BucketMask] = 0
			buckets[(key+16)&h4BucketMask] = 0
			buckets[(key+24)&h4BucketMask] = 0
		}
	} else {
		h.buckets = [h4BucketSize]uint32{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the bucket for the 5-byte sequence at
// data[pos & mask]. Uses the sweep offset to distribute entries.
func (h *h4) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	off := uint32(pos) & h4BucketSweepMsk
	h.buckets[(key+off)&h4BucketMask] = uint32(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h4) storeRange(data []byte, mask, start, end uint) {
	buckets := &h.buckets
	for i := start; i < end; i++ {
		key := h.hash(data, i&mask)
		off := uint32(i) & h4BucketSweepMsk
		buckets[(key+off)&h4BucketMask] = uint32(i)
	}
}

func (h *h4) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	off := uint32(pos) & h4BucketSweepMsk
	h.buckets[(key+off)&h4BucketMask] = uint32(pos)
}

func (h *h4) storeRangeNoWrap(data []byte, start, end uint) {
	buckets := &h.buckets
	for i := start; i < end; i++ {
		key := h.hash(data, i)
		off := uint32(i) & h4BucketSweepMsk
		buckets[(key+off)&h4BucketMask] = uint32(i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h4) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
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
// H4 uses BUCKET_SWEEP=4 and USE_DICTIONARY=1, so static dictionary
// lookups are performed when no hash-table match is found.
//
// When no compound dictionary is configured and the call's
// [wrappedPos, wrappedPos+bytes) range fits entirely within the ring buffer
// (no modular wrap) and no past call has wrapped, dispatch to
// createBackwardReferencesNoCompoundNoWrap which drops both the per-iteration
// compound branch and the `& mask` ops on stored bucket values.
func (h *h4) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	mask := uint(s.mask)
	if !h.everWrapped && s.compound.numChunks == 0 &&
		uint(wrappedPos)+uint(bytes) <= mask+1 {
		h.createBackwardReferencesNoCompoundNoWrap(s, bytes, wrappedPos)
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

		// Manually inlined findLongestMatch — the compiler cannot inline it
		// (cost 1098 > budget 80), so we do it here to eliminate call overhead.
		{
			lastDistance := s.distCache[0]
			curMasked := position & mask
			guardByte := loadByte(data, curMasked) // sr.len == 0 → bestLenIn == 0
			key := h.hash(data, curMasked)
			bestScore := sr.score // minScore, type uint
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

			// Entry 0.
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

			// Entry 1.
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

			// Entry 2.
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

			// Entry 3.
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

			buckets[keyOut] = uint32(position)
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

					// Entry 0.
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

					// Entry 1.
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

					// Entry 2.
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

					// Entry 3.
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

					buckets[keyOut] = uint32(cur2)
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

// createBackwardReferencesNoCompoundNoWrap is the H4 fast path used by
// createBackwardReferences when no compound dictionary is configured, the
// call's position range fits entirely within mask+1, and no past call has
// wrapped. It drops both the per-iteration compound lookup branch and the
// `& mask` ops on stored bucket values (each redundant when stored values
// are < mask+1). The paired no-wrap store helpers also skip redundant
// position masking while preserving the same bucket update.
func (h *h4) createBackwardReferencesNoCompoundNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap

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
		// (cost 1098 > budget 80), so we do it here to eliminate call overhead.
		{
			lastDistance := s.distCache[0]
			curMasked := position
			guardByte := loadByte(data, curMasked) // sr.len == 0 → bestLenIn == 0
			key := h.hash(data, curMasked)
			bestScore := sr.score // minScore, type uint
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

			// Entry 1.
			{
				backward := position - prev1
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

			// Entry 2.
			{
				backward := position - prev2
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

			// Entry 3.
			{
				backward := position - prev3
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

			buckets[keyOut] = uint32(position)
		}

		if sr.score == minScore {
			posM := position
			if wl, wi, ok := searchStaticDictAt(data, posM, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok {
				if matchStaticDictEntryAt(data, posM, wl, wi, maxDistance, sr.score, &sr) {
					s.dictNumMatches++
				}
			}
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
					curMasked := cur2
					bestLen := sr2.len
					guardByte := loadByte(data, curMasked+bestLen)
					key := h.hash(data, curMasked)
					bestScore := sr2.score // minScore

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

					// Entry 1.
					{
						backward := cur2 - prev1
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

					// Entry 2.
					{
						backward := cur2 - prev2
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

					// Entry 3.
					{
						backward := cur2 - prev3
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

					buckets[keyOut] = uint32(cur2)
				}

				if sr2.score == minScore {
					posM2 := position + 1
					if wl2, wi2, ok2 := searchStaticDictAt(data, posM2, maxLength, &s.dictNumLookups, &s.dictNumMatches); ok2 {
						if matchStaticDictEntryAt(data, posM2, wl2, wi2, maxDistance, sr2.score, &sr2) {
							s.dictNumMatches++
						}
					}
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
			distanceCode := computeDistanceCode(sr.distance, maxDistance, &s.distCache)
			if sr.distance <= maxDistance && distanceCode > 0 {
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
func (*h4) hash(data []byte, off uint) uint32 {
	v := loadU64LE(data, off)
	h := (v << (64 - 8*hashLen)) * hashMul64
	return uint32(h >> (64 - h4BucketBits))
}
