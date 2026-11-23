// Hash table for the streaming encoder (quality 3).
// BUCKET_BITS=16, BUCKET_SWEEP=2, HASH_LEN=5, USE_DICTIONARY=0.

package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// H3-specific configuration constants.
const (
	h3BucketSweep    = 2
	h3BucketSweepMsk = (h3BucketSweep - 1) << 3 // 8
)

// h3 is the H3 hasher: a forgetful hash table of fixed size that maps
// 5-byte sequences to positions. Each logical bucket stores two uint32
// positions spread across adjacent slots (BUCKET_SWEEP=2).
type h3 struct {
	buckets    [bucketSize]uint32
	nextBucket uint32 // speculative load to warm cache for the next match lookup
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions reaching or exceeding mask+1, after which the no-wrap
	// fast path is disabled because stored bucket values may then encode
	// positions outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

func (h *h3) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
func (h *h3) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := bucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := hashBytes(data, i)
			h.buckets[key] = 0
			h.buckets[(key+8)&bucketMask] = 0
		}
	} else {
		h.buckets = [bucketSize]uint32{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the bucket for the 5-byte sequence at
// data[pos & mask]. Uses the sweep offset to distribute entries.
func (h *h3) store(data []byte, mask, pos uint) {
	key := hashBytes(data, pos&mask)
	off := uint32(pos) & h3BucketSweepMsk
	h.buckets[(key+off)&bucketMask] = uint32(pos)
}

func (h *h3) storeNoWrap(data []byte, pos uint) {
	key := hashBytes(data, pos)
	off := uint32(pos) & h3BucketSweepMsk
	h.buckets[(key+off)&bucketMask] = uint32(pos)
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h3) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
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
// H3 uses BUCKET_SWEEP=2 and USE_DICTIONARY=0, so no static dictionary
// lookups are performed.
//
// When no compound dictionary is configured and the call's
// [wrappedPos, wrappedPos+bytes) range fits entirely within the ring buffer
// (no modular wrap) and no past call has wrapped, dispatch to
// createBackwardReferencesNoCompoundNoWrap which drops both the per-iteration
// compound branch and the `& mask` ops on stored bucket values.
func (h *h3) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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

	// StoreLookahead for H3 is 8 (same as H2).
	storeEnd := position
	if uint(bytes) >= hashTypeLength {
		storeEnd = posEnd - hashTypeLength + 1
	}

	// Random heuristics for uncompressible data. Quality < 9 → window = 64.
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
		// (cost 671 > budget 80), so we do it here to eliminate call overhead.
		{
			lastDistance := s.distCache[0]
			curMasked := position & mask
			// Fuse the guard byte and hash load — both read from data[curMasked].
			// Doing one 8-byte load and deriving both keeps loadByte off the
			// dependency chain and hands the compiler a single fused access.
			curWord := loadU64LE(data, curMasked)
			guardByte := byte(curWord) // sr.len == 0 → bestLenIn == 0
			key := uint32(((curWord << (64 - 8*hashLen)) * hashMul64) >> (64 - bucketBits))
			bestScore := sr.score // minScore, type uint
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & bucketMask
			keyOut := hkey0
			if (position & h3BucketSweepMsk) != 0 {
				keyOut = hkey1
			}

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])

			{
				prev := position - lastDistance
				if prev < position && lastDistance <= maxDistance {
					prev &= mask
					if guardByte == loadByte(data, prev+bestLen) {
						length := matchLenAt(data, prev, curMasked, int(maxLength))
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
				if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev0, curMasked, int(maxLength))
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
				if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
					length := matchLenAt(data, prev1, curMasked, int(maxLength))
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

		if hasCompound {
			s.compound.lookupMatch(data, mask,
				&s.distCache, position, maxLength,
				maxDistance, &sr)
		}

		if sr.score > minScore {
			// Found a match. Try lazy matching: look one position ahead.
			delayedBackwardReferencesInRow := 0
			maxLength--
			for {
				const costDiffLazy = 175
				var sr2 hasherSearchResult
				// quality < MIN_QUALITY_FOR_EXTENSIVE_REFERENCE_SEARCH (5):
				// cap sr2.len to sr.len-1.
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
					bestLen2 := sr2.len
					guardByte := loadByte(data, curMasked+bestLen2)
					key := hashBytes(data, curMasked)
					bestScore := sr2.score // minScore

					hkey0 := key
					hkey1 := (key + 8) & bucketMask
					keyOut := hkey0
					if (cur2 & h3BucketSweepMsk) != 0 {
						keyOut = hkey1
					}

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])

					{
						prev := cur2 - lastDistance
						if prev < cur2 && lastDistance <= maxDistance {
							prev &= mask
							if guardByte == loadByte(data, prev+bestLen2) {
								length := matchLenAt(data, prev, curMasked, int(maxLength))
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										bestLen2 = uint(length)
										sr2.len = bestLen2
										sr2.distance = lastDistance
										sr2.score = score
										bestScore = score
										guardByte = loadByte(data, curMasked+bestLen2)
									}
								}
							}
						}
					}

					{
						backward := cur2 - prev0
						prev0 &= mask
						if guardByte == loadByte(data, prev0+bestLen2) && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev0, curMasked, int(maxLength))
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen2 = uint(length)
									sr2.len = bestLen2
									guardByte = loadByte(data, curMasked+bestLen2)
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
						if guardByte == loadByte(data, prev1+bestLen2) && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev1, curMasked, int(maxLength))
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

				if hasCompound {
					s.compound.lookupMatch(data, mask,
						&s.distCache, position+1, maxLength,
						maxDistance, &sr2)
				}

				if sr2.score >= sr.score+costDiffLazy {
					// Better match one byte ahead. Write one literal.
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

			// Compute distance code and update distance cache.
			// Recompute maxDistance after the lazy loop because position may
			// have advanced. This matches the C reference's dictionary_start.
			maxDistance = min(position, maxBackwardLimit)
			distanceCode := computeDistanceCode(sr.distance, maxDistance+gap, &s.distCache)
			if sr.distance <= maxDistance+gap && distanceCode > 0 {
				s.distCache[3] = s.distCache[2]
				s.distCache[2] = s.distCache[1]
				s.distCache[1] = s.distCache[0]
				s.distCache[0] = sr.distance
				// H3 PrepareDistanceCache is a no-op (BUCKET_SWEEP<=2).
			}

			{
				delta := uint32(uint8(int8(sr.lenCodeDelta)))
				distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
				effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
				insCode := getInsertLenCode(insertLength)
				copyCode := getCopyLenCode(effectiveCopyLen)
				cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
				s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
				// Inline histogram update: accumulate symbol counts while the
				// insert-literal bytes and command fields are hot in registers,
				// avoiding the separate tally pass in writeMetaBlockTrivial.
				if s.cmdHisto != nil {
					s.cmdHisto[cmdPrefix]++
					// Distance symbol is written when cmdPrefix >= 128
					// (!usesLastDistance). This matches tally's condition:
					//   copyLen != 0 && !cmd.usesLastDistance()
					// Note: distPrefix&0x3FF may be 0 even when cmdPrefix >= 128
					// (last-distance copy with long insert/copy length codes).
					if cmdPrefix >= 128 {
						s.distHisto[distPrefix&0x3FF]++
					}
					basePos := position - insertLength
					for j := uint(0); j < insertLength; j++ {
						s.litHisto[data[(basePos+j)&mask]]++
					}
				}
			}
			s.numLiterals += insertLength
			insertLength = 0

			// Pre-warm the hash bucket for the next main-loop iteration at
			// position+sr.len. After a match the position jumps by sr.len, leaving
			// that bucket cold (potential L3 miss). Issuing the load here lets the
			// store loop below hide the cache-miss latency.
			nextMatchPos := position + sr.len
			if nextMatchPos+hashTypeLength < posEnd {
				h.nextBucket = buckets[hashBytes(data, nextMatchPos&mask)]
			}

			// Store hash entries for the matched range, avoiding RLE poisoning.
			// Direct loop (vs. h.storeRange) lets the compiler use the local
			// buckets pointer, improving alias analysis for surrounding reads.
			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			for i := rangeStart; i < rangeEnd; i++ {
				key := hashBytes(data, i&mask)
				off := uint32(i) & h3BucketSweepMsk
				buckets[(key+off)&bucketMask] = uint32(i)
			}

			position += sr.len
		} else {
			// No good match. Insert a literal.
			insertLength++
			position++

			// Random heuristics: skip match lookups in uncompressible regions.
			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					// Very long run without matches: stride by 4, store every 4th.
					posJump := min(position+16, posEnd-(hashTypeLength-1))
					for position < posJump {
						h.store(data, mask, position)
						insertLength += 4
						position += 4
					}
				} else {
					// Moderate run without matches: stride by 2, store every 2nd.
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

// createBackwardReferencesNoCompoundNoWrap is the H3 fast path used by
// createBackwardReferences when no compound dictionary is configured, the
// call's position range fits entirely within mask+1, and no past call has
// wrapped. It drops both the per-iteration compound lookup branch and the
// `& mask` ops on stored bucket values (each redundant when stored values
// are < mask+1). The paired no-wrap store helpers also skip redundant
// position masking while preserving the same bucket update.
func (h *h3) createBackwardReferencesNoCompoundNoWrap(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	// StoreLookahead for H3 is 8 (same as H2).
	storeEnd := position
	if uint(bytes) >= hashTypeLength {
		storeEnd = posEnd - hashTypeLength + 1
	}

	// Random heuristics for uncompressible data. Quality < 9 → window = 64.
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
		// (cost 671 > budget 80), so we do it here to eliminate call overhead.
		{
			lastDistance := s.distCache[0]
			curMasked := position
			// Fuse the guard byte and hash load — both read from data[curMasked].
			// Doing one 8-byte load and deriving both keeps loadByte off the
			// dependency chain and hands the compiler a single fused access.
			curWord := loadU64LE(data, curMasked)
			guardByte := byte(curWord) // sr.len == 0 → bestLenIn == 0
			key := uint32(((curWord << (64 - 8*hashLen)) * hashMul64) >> (64 - bucketBits))
			bestScore := sr.score // minScore, type uint
			bestLen := uint(0)

			hkey0 := key
			hkey1 := (key + 8) & bucketMask
			keyOut := hkey0
			if (position & h3BucketSweepMsk) != 0 {
				keyOut = hkey1
			}

			prev0 := uint(buckets[hkey0])
			prev1 := uint(buckets[hkey1])

			{
				prev := position - lastDistance
				if prev < position && lastDistance <= maxDistance {
					if guardByte == loadByte(data, prev+bestLen) {
						// Manually inline first 8-byte chunk of matchLenAt,
						// reusing curWord (= loadU64LE(data, curMasked)) instead
						// of issuing a second load for the b side. Short matches
						// (< 8 bytes) return without a function call to the
						// matchLenAt tail.
						length := 0
						xor := loadU64LE(data, prev) ^ curWord
						if xor != 0 {
							length = bits.TrailingZeros64(xor) / 8
						} else {
							length = 8 + matchLenAt(data, prev+8, curMasked+8, int(maxLength)-8)
						}
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
				if guardByte == loadByte(data, prev0+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev0) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev0+8, curMasked+8, int(maxLength)-8)
					}
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
				if guardByte == loadByte(data, prev1+bestLen) && backward != 0 && backward <= maxDistance {
					length := 0
					xor := loadU64LE(data, prev1) ^ curWord
					if xor != 0 {
						length = bits.TrailingZeros64(xor) / 8
					} else {
						length = 8 + matchLenAt(data, prev1+8, curMasked+8, int(maxLength)-8)
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
			// Found a match. Try lazy matching: look one position ahead.
			delayedBackwardReferencesInRow := 0
			maxLength--
			for {
				const costDiffLazy = 175
				var sr2 hasherSearchResult
				// quality < MIN_QUALITY_FOR_EXTENSIVE_REFERENCE_SEARCH (5):
				// cap sr2.len to sr.len-1.
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
					bestLen2 := sr2.len
					guardByte := loadByte(data, curMasked+bestLen2)
					key := hashBytes(data, curMasked)
					bestScore := sr2.score // minScore

					hkey0 := key
					hkey1 := (key + 8) & bucketMask
					keyOut := hkey0
					if (cur2 & h3BucketSweepMsk) != 0 {
						keyOut = hkey1
					}

					prev0 := uint(buckets[hkey0])
					prev1 := uint(buckets[hkey1])

					{
						prev := cur2 - lastDistance
						if prev < cur2 && lastDistance <= maxDistance {
							if guardByte == loadByte(data, prev+bestLen2) {
								length := matchLenAt(data, prev, curMasked, int(maxLength))
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										bestLen2 = uint(length)
										sr2.len = bestLen2
										sr2.distance = lastDistance
										sr2.score = score
										bestScore = score
										guardByte = loadByte(data, curMasked+bestLen2)
									}
								}
							}
						}
					}

					{
						backward := cur2 - prev0
						if guardByte == loadByte(data, prev0+bestLen2) && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev0, curMasked, int(maxLength))
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									bestLen2 = uint(length)
									sr2.len = bestLen2
									guardByte = loadByte(data, curMasked+bestLen2)
									bestScore = score
									sr2.score = score
									sr2.distance = backward
								}
							}
						}
					}

					{
						backward := cur2 - prev1
						if guardByte == loadByte(data, prev1+bestLen2) && backward != 0 && backward <= maxDistance {
							length := matchLenAt(data, prev1, curMasked, int(maxLength))
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
					// Better match one byte ahead. Write one literal.
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

			// Compute distance code and update distance cache.
			// Recompute maxDistance after the lazy loop because position may
			// have advanced. This matches the C reference's dictionary_start.
			maxDistance = min(position, maxBackwardLimit)
			distanceCode := computeDistanceCode(sr.distance, maxDistance, &s.distCache)
			if sr.distance <= maxDistance && distanceCode > 0 {
				s.distCache[3] = s.distCache[2]
				s.distCache[2] = s.distCache[1]
				s.distCache[1] = s.distCache[0]
				s.distCache[0] = sr.distance
				// H3 PrepareDistanceCache is a no-op (BUCKET_SWEEP<=2).
			}

			{
				delta := uint32(uint8(int8(sr.lenCodeDelta)))
				distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
				effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
				insCode := getInsertLenCode(insertLength)
				copyCode := getCopyLenCode(effectiveCopyLen)
				cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
				s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
				// Inline histogram update: accumulate symbol counts while the
				// insert-literal bytes and command fields are hot in registers,
				// avoiding the separate tally pass in writeMetaBlockTrivial.
				if s.cmdHisto != nil {
					s.cmdHisto[cmdPrefix]++
					// Distance symbol is written when cmdPrefix >= 128
					// (!usesLastDistance). This matches tally's condition:
					//   copyLen != 0 && !cmd.usesLastDistance()
					// Note: distPrefix&0x3FF may be 0 even when cmdPrefix >= 128
					// (last-distance copy with long insert/copy length codes).
					if cmdPrefix >= 128 {
						s.distHisto[distPrefix&0x3FF]++
					}
					basePos := position - insertLength
					for j := uint(0); j < insertLength; j++ {
						s.litHisto[data[basePos+j]]++
					}
				}
			}
			s.numLiterals += insertLength
			insertLength = 0

			// Pre-warm the hash bucket for the next main-loop iteration at
			// position+sr.len. After a match the position jumps by sr.len, leaving
			// that bucket cold (potential L3 miss). Issuing the load here lets the
			// store loop below hide the cache-miss latency.
			nextMatchPos := position + sr.len
			if nextMatchPos+hashTypeLength < posEnd {
				h.nextBucket = buckets[hashBytes(data, nextMatchPos)]
			}

			// Store hash entries for the matched range, avoiding RLE poisoning.
			// Direct loop (vs. h.storeRange) lets the compiler use the local
			// buckets pointer, improving alias analysis for surrounding reads.
			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			if rangeEnd >= rangeStart+2 {
				last := rangeEnd - 2
				i := rangeStart
				for ; i < last; i += 2 {
					buckets[(hashBytes(data, i)+uint32(i)&h3BucketSweepMsk)&bucketMask] = uint32(i)
					buckets[(hashBytes(data, i+1)+uint32(i+1)&h3BucketSweepMsk)&bucketMask] = uint32(i + 1)
				}
				buckets[(hashBytes(data, last)+uint32(last)&h3BucketSweepMsk)&bucketMask] = uint32(last)
				buckets[(hashBytes(data, last+1)+uint32(last+1)&h3BucketSweepMsk)&bucketMask] = uint32(last + 1)
			} else {
				for i := rangeStart; i < rangeEnd; i++ {
					key := hashBytes(data, i)
					off := uint32(i) & h3BucketSweepMsk
					buckets[(key+off)&bucketMask] = uint32(i)
				}
			}

			position += sr.len
		} else {
			// No good match. Insert a literal.
			insertLength++
			position++

			// Random heuristics: skip match lookups in uncompressible regions.
			if position > applyRandomHeuristics {
				if position > applyRandomHeuristics+4*randomHeuristicsWindowSize {
					// Very long run without matches: stride by 4, store every 4th.
					posJump := min(position+16, posEnd-(hashTypeLength-1))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					// Moderate run without matches: stride by 2, store every 2nd.
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
