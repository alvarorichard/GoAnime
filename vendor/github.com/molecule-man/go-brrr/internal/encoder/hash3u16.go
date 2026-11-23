// H3 variant for inputs bounded to 64 KiB.
// Stores uint16 positions in buckets, halving the table from 256 KB to 128 KB.
// The dispatch contract guarantees position never exceeds 2^16, so
// `uint16(pos)` is lossless — no aliasing, no stale-probe overhead. If buffered
// input crosses the 64 KiB boundary mid-stream (when sizeHint underestimated
// the real total), encoderCore.maybePromoteHasher transparently swaps to h3.

package encoder

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// h3u16 is the H3 hasher with uint16 bucket slots, dispatched only when the
// encoder expects the input to fit in 64 KiB (either via a user-supplied
// sizeHint or via the one-shot isLast guarantee). Within that bound every
// stored position is in [0, 65535], so `uint16(pos)` storage is lossless and
// `position - prev` (uint subtraction) is directly the real backward distance
// — no modular truncation needed. Semantics match h3 at half the memory; the
// lookup arithmetic is the same shape.
type h3u16 struct {
	buckets    [bucketSize]uint16
	nextBucket uint16 // speculative load to warm cache for the next match lookup
	hasherCommon
}

func (h *h3u16) common() *hasherCommon { return &h.hasherCommon }

// reset clears or selectively zeroes the hash table before use.
//
// The partial path views buckets through a uint32 alias and zeroes pairs of
// adjacent uint16 slots per store. This avoids the length-changing-prefix
// (66h) predecode stalls that `MOVW $0, mem` triggers when fetched from
// MITE, and is exact for the two sweep slots after rounding the hash key
// down to even — clearing the partner slot in each pair is harmless: it
// would be either stale (and is meant to be cleared) or already zero.
func (h *h3u16) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := bucketSize >> 5
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		const mask32 = bucketSize/2 - 1
		buckets32 := (*[bucketSize / 2]uint32)(unsafe.Pointer(&h.buckets))
		for i := range inputSize {
			k := hashBytes(data, i) >> 1
			buckets32[k] = 0
			buckets32[(k+4)&mask32] = 0
		}
	} else {
		h.buckets = [bucketSize]uint16{}
	}
	h.ready = true
}

// store records position pos in the bucket for the 5-byte sequence at
// data[pos & mask]. Uses the sweep offset to distribute entries.
func (h *h3u16) store(data []byte, mask, pos uint) {
	key := hashBytes(data, pos&mask)
	off := uint32(pos) & h3BucketSweepMsk
	h.buckets[(key+off)&bucketMask] = uint16(pos)
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h3u16) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= hashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// createBackwardReferences finds backward reference matches using this hasher
// and populates s.commands. The hot findLongestMatch/store calls are direct
// (non-virtual) since the receiver is concrete.
func (h *h3u16) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
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
			curMasked := position & mask
			curWord := loadU64LE(data, curMasked)
			guardByte := byte(curWord)
			key := uint32(((curWord << (64 - 8*hashLen)) * hashMul64) >> (64 - bucketBits))
			bestScore := sr.score
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

			buckets[keyOut] = uint16(position)
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
					bestLen2 := sr2.len
					guardByte := loadByte(data, curMasked+bestLen2)
					key := hashBytes(data, curMasked)
					bestScore := sr2.score

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

					buckets[keyOut] = uint16(cur2)
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

			{
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
			for i := rangeStart; i < rangeEnd; i++ {
				key := hashBytes(data, i&mask)
				off := uint32(i) & h3BucketSweepMsk
				buckets[(key+off)&bucketMask] = uint16(i)
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
