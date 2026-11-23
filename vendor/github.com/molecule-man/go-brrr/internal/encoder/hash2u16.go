// H2 variant for inputs bounded to 64 KiB.
// Stores uint16 positions in buckets, halving the table from 256 KB to 128 KB.
// The dispatch contract guarantees position never exceeds 2^16, so
// `uint16(pos)` is lossless — no aliasing, no stale-probe overhead. If buffered
// input crosses the 64 KiB boundary mid-stream (when sizeHint underestimated
// the real total), encoderCore.maybePromoteHasher transparently swaps to h2.

package encoder

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// h2u16 is the H2 hasher with uint16 bucket slots, dispatched only when the
// encoder expects the input to fit in 64 KiB (either via a user-supplied
// sizeHint or via the one-shot isLast guarantee). Within that bound every
// stored position is in [0, 65535], so `uint16(pos)` storage is lossless and
// `position - prev` (uint subtraction) is directly the real backward distance
// — no modular truncation needed. Semantics match h2 at half the memory; the
// lookup arithmetic is the same shape.
type h2u16 struct {
	buckets    [bucketSize]uint16
	nextBucket uint16 // speculative load to warm cache for the next match lookup
	// everWrapped is sticky: false until any createBackwardReferences call
	// has positions reaching or exceeding mask+1, after which the no-wrap
	// fast path is disabled because stored bucket values may then encode
	// positions outside the ring buffer's modular window.
	everWrapped bool
	hasherCommon
}

func (h *h2u16) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
//
// The partial path views buckets through a uint32 alias and zeroes pairs of
// adjacent uint16 slots per store. This avoids the length-changing-prefix
// (66h) predecode stalls that `MOVW $0, mem` triggers when fetched from
// MITE — clearing the partner slot in each pair is harmless: it would be
// either stale (and is meant to be cleared) or already zero.
func (h *h2u16) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := bucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		buckets32 := (*[bucketSize / 2]uint32)(unsafe.Pointer(&h.buckets))
		for i := range inputSize {
			buckets32[hashBytes(data, i)>>1] = 0
		}
	} else {
		h.buckets = [bucketSize]uint16{}
	}
	h.everWrapped = false
	h.ready = true
}

// store records position pos in the bucket for the 5-byte sequence at
// data[pos & mask].
func (h *h2u16) store(data []byte, mask, pos uint) {
	key := hashBytes(data, pos&mask)
	h.buckets[key] = uint16(pos)
}

// storeRange records positions [start, end) in the hash table.
func (h *h2u16) storeRange(data []byte, mask, start, end uint) {
	for i := start; i < end; i++ {
		h.store(data, mask, i)
	}
}

func (h *h2u16) storeNoWrap(data []byte, pos uint) {
	key := hashBytes(data, pos)
	h.buckets[key] = uint16(pos)
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h2u16) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= hashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. Mirrors h2.createBackwardReferences with uint16
// bucket slots; positions in [0, 65535] make the storage lossless.
//
// When the call's [wrappedPos, wrappedPos+bytes) range fits entirely within
// the ring buffer (no modular wrap) and no past call has wrapped, dispatch
// to createBackwardReferencesNoWrap which omits the per-iteration & mask
// ops (each redundant when stored bucket values are < mask+1).
func (h *h2u16) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
	const directStoreRangeMinBytes = 4096
	directStoreRange := bytes >= directStoreRangeMinBytes

	for position+hashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.len = 0
		sr.lenCodeDelta = 0
		sr.distance = 0
		sr.score = minScore

		var dictEligible bool
		{
			lastDistance := s.distCache[0]
			curMasked := position & mask
			guardByte := loadByte(data, curMasked)
			key := hashBytes(data, curMasked)
			bestScore := sr.score

			lastDistanceHit := false
			{
				prev := position - lastDistance
				if prev < position {
					prev &= mask
					if guardByte == loadByte(data, prev) {
						length := matchLenAt(data, prev, curMasked, int(maxLength))
						if length >= 4 {
							score := backwardReferenceScoreUsingLastDistance(uint(length))
							if bestScore < score {
								sr.len = uint(length)
								sr.distance = lastDistance
								sr.score = score
								buckets[key] = uint16(position)
								lastDistanceHit = true
							}
						}
					}
				}
			}

			if !lastDistanceHit {
				prev := uint(buckets[key])
				buckets[key] = uint16(position)
				backward := position - prev
				prev &= mask
				if guardByte == loadByte(data, prev) && backward != 0 && backward <= maxDistance {
					dictEligible = true
					length := matchLenAt(data, prev, curMasked, int(maxLength))
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							sr.len = uint(length)
							sr.distance = backward
							sr.score = score
						}
					}
				}
			}
		}

		if dictEligible && sr.score == minScore {
			if m, ok := searchStaticDictionary(data[position&mask:], maxLength, maxDistance+gap, maxBackwardDistance,
				&s.dictNumLookups, &s.dictNumMatches, sr.score); ok {
				sr = m
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

				var dictEligible2 bool
				{
					cur2 := position + 1
					lastDistance := s.distCache[0]
					curMasked := cur2 & mask
					bestLen := sr2.len
					guardByte := loadByte(data, curMasked+bestLen)
					key := hashBytes(data, curMasked)
					bestScore := sr2.score

					lastDistanceHit := false
					{
						prev := cur2 - lastDistance
						if prev < cur2 {
							prev &= mask
							if guardByte == loadByte(data, prev+bestLen) {
								length := matchLenAt(data, prev, curMasked, int(maxLength))
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										sr2.len = uint(length)
										sr2.distance = lastDistance
										sr2.score = score
										buckets[key] = uint16(cur2)
										lastDistanceHit = true
									}
								}
							}
						}
					}

					if !lastDistanceHit {
						prev := uint(buckets[key])
						buckets[key] = uint16(cur2)
						backward := cur2 - prev
						prev &= mask
						if guardByte == loadByte(data, prev+bestLen) && backward != 0 && backward <= maxDistance {
							dictEligible2 = true
							length := matchLenAt(data, prev, curMasked, int(maxLength))
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									sr2.len = uint(length)
									sr2.distance = backward
									sr2.score = score
								}
							}
						}
					}
				}

				if dictEligible2 && sr2.score == minScore {
					if m, ok := searchStaticDictionary(data[(position+1)&mask:], maxLength, maxDistance+gap, maxBackwardDistance,
						&s.dictNumLookups, &s.dictNumMatches, sr2.score); ok {
						sr2 = m
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
			if s.cmdHisto != nil {
				s.cmdHisto[cmdPrefix]++
				if cmdPrefix >= 128 {
					s.distHisto[distPrefix&0x3FF]++
				}
				basePos := position - insertLength
				for j := uint(0); j < insertLength; j++ {
					s.litHisto[data[(basePos+j)&mask]]++
				}
			}
			s.numLiterals += insertLength
			insertLength = 0

			nextMatchPos := position + sr.len
			if nextMatchPos+hashTypeLength < posEnd {
				h.nextBucket = buckets[hashBytes(data, nextMatchPos&mask)]
			}

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			if directStoreRange {
				if rangeEnd >= rangeStart+4 {
					last := rangeEnd - 4
					i := rangeStart
					for ; i < last; i += 4 {
						buckets[hashBytes(data, i&mask)] = uint16(i)
						buckets[hashBytes(data, (i+1)&mask)] = uint16(i + 1)
						buckets[hashBytes(data, (i+2)&mask)] = uint16(i + 2)
						buckets[hashBytes(data, (i+3)&mask)] = uint16(i + 3)
					}
					buckets[hashBytes(data, last&mask)] = uint16(last)
					buckets[hashBytes(data, (last+1)&mask)] = uint16(last + 1)
					buckets[hashBytes(data, (last+2)&mask)] = uint16(last + 2)
					buckets[hashBytes(data, (last+3)&mask)] = uint16(last + 3)
				} else {
					for i := rangeStart; i < rangeEnd; i++ {
						buckets[hashBytes(data, i&mask)] = uint16(i)
					}
				}
			} else {
				h.storeRange(data, mask, rangeStart, rangeEnd)
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
// `position & mask` ops are redundant and elided here. The paired
// no-wrap store helpers also skip redundant position masking while
// preserving the same bucket update.
func (h *h2u16) createBackwardReferencesNoWrap(s *encodeState, bytes, wrappedPos uint32) {
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

		var dictEligible bool
		{
			lastDistance := s.distCache[0]
			guardByte := loadByte(data, position)
			key := hashBytes(data, position)
			bestScore := sr.score

			lastDistanceHit := false
			{
				prev := position - lastDistance
				if prev < position {
					if guardByte == loadByte(data, prev) {
						length := matchLenAt(data, prev, position, int(maxLength))
						if length >= 4 {
							score := backwardReferenceScoreUsingLastDistance(uint(length))
							if bestScore < score {
								sr.len = uint(length)
								sr.distance = lastDistance
								sr.score = score
								buckets[key] = uint16(position)
								lastDistanceHit = true
							}
						}
					}
				}
			}

			if !lastDistanceHit {
				prev := uint(buckets[key])
				buckets[key] = uint16(position)
				backward := position - prev
				if guardByte == loadByte(data, prev) && backward != 0 && backward <= maxDistance {
					dictEligible = true
					length := matchLenAt(data, prev, position, int(maxLength))
					if length >= 4 {
						score := backwardReferenceScore(uint(length), backward)
						if bestScore < score {
							sr.len = uint(length)
							sr.distance = backward
							sr.score = score
						}
					}
				}
			}
		}

		if dictEligible && sr.score == minScore {
			if m, ok := searchStaticDictionary(data[position:], maxLength, maxDistance+gap, maxBackwardDistance,
				&s.dictNumLookups, &s.dictNumMatches, sr.score); ok {
				sr = m
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

				var dictEligible2 bool
				{
					cur2 := position + 1
					lastDistance := s.distCache[0]
					bestLen := sr2.len
					guardByte := loadByte(data, cur2+bestLen)
					key := hashBytes(data, cur2)
					bestScore := sr2.score

					lastDistanceHit := false
					{
						prev := cur2 - lastDistance
						if prev < cur2 {
							if guardByte == loadByte(data, prev+bestLen) {
								length := matchLenAt(data, prev, cur2, int(maxLength))
								if length >= 4 {
									score := backwardReferenceScoreUsingLastDistance(uint(length))
									if bestScore < score {
										sr2.len = uint(length)
										sr2.distance = lastDistance
										sr2.score = score
										buckets[key] = uint16(cur2)
										lastDistanceHit = true
									}
								}
							}
						}
					}

					if !lastDistanceHit {
						prev := uint(buckets[key])
						buckets[key] = uint16(cur2)
						backward := cur2 - prev
						if guardByte == loadByte(data, prev+bestLen) && backward != 0 && backward <= maxDistance {
							dictEligible2 = true
							length := matchLenAt(data, prev, cur2, int(maxLength))
							if length >= 4 {
								score := backwardReferenceScore(uint(length), backward)
								if bestScore < score {
									sr2.len = uint(length)
									sr2.distance = backward
									sr2.score = score
								}
							}
						}
					}
				}

				if dictEligible2 && sr2.score == minScore {
					if m, ok := searchStaticDictionary(data[position+1:], maxLength, maxDistance+gap, maxBackwardDistance,
						&s.dictNumLookups, &s.dictNumMatches, sr2.score); ok {
						sr2 = m
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
			if s.cmdHisto != nil {
				s.cmdHisto[cmdPrefix]++
				if cmdPrefix >= 128 {
					s.distHisto[distPrefix&0x3FF]++
				}
				basePos := position - insertLength
				for j := uint(0); j < insertLength; j++ {
					s.litHisto[data[basePos+j]]++
				}
			}
			s.numLiterals += insertLength
			insertLength = 0

			nextMatchPos := position + sr.len
			if nextMatchPos+hashTypeLength < posEnd {
				h.nextBucket = buckets[hashBytes(data, nextMatchPos)]
			}

			rangeStart := position + 2
			rangeEnd := min(position+sr.len, storeEnd)
			if sr.distance < sr.len>>2 {
				rangeStart = min(rangeEnd, max(rangeStart, position+sr.len-(sr.distance<<2)))
			}
			if rangeEnd >= rangeStart+4 {
				last := rangeEnd - 4
				i := rangeStart
				for ; i < last; i += 4 {
					buckets[hashBytes(data, i)] = uint16(i)
					buckets[hashBytes(data, i+1)] = uint16(i + 1)
					buckets[hashBytes(data, i+2)] = uint16(i + 2)
					buckets[hashBytes(data, i+3)] = uint16(i + 3)
				}
				buckets[hashBytes(data, last)] = uint16(last)
				buckets[hashBytes(data, last+1)] = uint16(last + 1)
				buckets[hashBytes(data, last+2)] = uint16(last + 2)
				buckets[hashBytes(data, last+3)] = uint16(last + 3)
			} else {
				for i := rangeStart; i < rangeEnd; i++ {
					buckets[hashBytes(data, i)] = uint16(i)
				}
			}

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
