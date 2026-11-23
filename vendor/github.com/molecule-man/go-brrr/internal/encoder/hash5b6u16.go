// H5b6 hasher variant for inputs bounded to 64 KiB.
//
// h5b6u16 mirrors h5b6 (quality 7) with two adaptations specific to the
// "size hint says small" branch:
//  1. uint16 bucket slots: stored positions never exceed 65535, so the bucket
//     ring buffer halves from 8 MiB to 4 MiB. The smaller table fits in L3 on
//     more CPUs and roughly halves Phase 2 cache-line traffic per bucket scan.
//  2. No speculative cur+1 prefetch: the hash table working set is small
//     enough at this size hint that the prefetch's extra hash + load chain
//     costs more than the latency it hides. It is unconditionally elided.
//
// Because the size hint guarantees position never wraps modulo mask+1, the
// findLongestMatch / createBackwardReferences "wrap" and "small buf" paths
// are absent here — only the no-wrap fast path remains. If buffered input
// crosses the 64 KiB boundary mid-stream (when the hint underestimated the
// real total), encoderCore.maybePromoteHasher transparently swaps to h5b6.

package encoder

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// h5b6u16 is the h5b6 hasher with uint16 bucket slots, dispatched only when
// the encoder expects the input to fit in 64 KiB. Configuration constants
// (bucket bits, block bits, hash type length, distance-cache layout) match
// h5b6 verbatim — only the storage width and the auxiliary state differ.
type h5b6u16 struct {
	num     [h5bBucketSize]uint16                 // entry count per bucket
	buckets [h5bBucketSize * h5b6BlockSize]uint16 // position ring buffers
	hasherCommon
}

func (h *h5b6u16) common() *hasherCommon { return &h.hasherCommon }

// bucketAt returns a fixed-size ring pointer so scan loops need no bounds check.
func (h *h5b6u16) bucketAt(key uint32) *[h5b6BlockSize]uint16 {
	return (*[h5b6BlockSize]uint16)(unsafe.Add(unsafe.Pointer(&h.buckets), uintptr(key)<<(h5b6BlockBits+1)))
}

func (h *h5b6u16) hash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h5bHashShift
}

// reset clears touched counts for small one-shot inputs to avoid a 64 KiB memset.
func (h *h5b6u16) reset(oneShot bool, inputSize uint, data []byte) {
	partialPrepareThreshold := h5bBucketSize >> 6
	if oneShot && inputSize <= uint(partialPrepareThreshold) {
		for i := range inputSize {
			key := h.hash(data, i)
			h.num[key] = 0
		}
	} else {
		h.num = [h5bBucketSize]uint16{}
	}
	h.ready = true
}

// store records position pos in the ring buffer for the 4-byte sequence at
// data[pos & mask]. uint16 truncation is lossless under the 64 KiB
// dispatch contract.
func (h *h5b6u16) store(data []byte, mask, pos uint) {
	key := h.hash(data, pos&mask)
	minorIx := h.num[key] & h5b6BlockMask
	offset := uint(minorIx) + uint(key)<<h5b6BlockBits
	h.num[key]++
	h.buckets[offset] = uint16(pos)
}

func (h *h5b6u16) storeNoWrap(data []byte, pos uint) {
	key := h.hash(data, pos)
	minorIx := h.num[key] & h5b6BlockMask
	offset := uint(minorIx) + uint(key)<<h5b6BlockBits
	h.num[key]++
	h.buckets[offset] = uint16(pos)
}

func (h *h5b6u16) storeRangeNoWrap(data []byte, start, end uint) {
	for i := start; i < end; i++ {
		h.storeNoWrap(data, i)
	}
}

// stitchToPreviousBlock seeds the hash table with the last 3 positions of
// the previous block so that cross-block matches can be found.
func (h *h5b6u16) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	if numBytes >= h5bHashTypeLength-1 && position >= 3 {
		h.store(ringBuffer, ringBufferMask, position-3)
		h.store(ringBuffer, ringBufferMask, position-2)
		h.store(ringBuffer, ringBufferMask, position-1)
	}
}

// createBackwardReferences finds backward reference matches and populates
// s.commands. By dispatch contract the position range never wraps the ring
// buffer, so a single straight-line variant suffices — there is no separate
// "no-wrap" fast path or "small ringbuffer" fallback.
func (h *h5b6u16) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	data := s.data
	mask := uint(s.mask)
	maxBackwardLimit := (uint(1) << s.lgwin) - core.WindowGap
	gap := s.compound.totalSize
	hasCompound := s.compound.numChunks > 0

	insertLength := s.lastInsertLen
	position := uint(wrappedPos)
	posEnd := position + uint(bytes)

	storeEnd := position
	if uint(bytes) >= h5bHashTypeLength {
		storeEnd = posEnd - h5bHashTypeLength + 1
	}

	const randomHeuristicsWindowSize = 64
	applyRandomHeuristics := position + randomHeuristicsWindowSize

	origCmdCount := uint(len(s.commands))

	// Expand the 4-entry distance cache to 10 derived entries.
	var distCache [16]uint
	d0 := s.distCache[0]
	distCache[0] = d0
	distCache[1] = s.distCache[1]
	distCache[2] = s.distCache[2]
	distCache[3] = s.distCache[3]
	distCache[4] = d0 - 1
	distCache[5] = d0 + 1
	distCache[6] = d0 - 2
	distCache[7] = d0 + 2
	distCache[8] = d0 - 3
	distCache[9] = d0 + 3

	for position+h5bHashTypeLength < posEnd {
		maxLength := posEnd - position
		maxDistance := min(position, maxBackwardLimit)

		var sr hasherSearchResult
		sr.score = minScore

		h.findLongestMatch(data, &distCache,
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

				h.findLongestMatch(data, &distCache,
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
						position+h5bHashTypeLength < posEnd {
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
				distCache[3] = distCache[2]
				distCache[2] = distCache[1]
				distCache[1] = d0
				d0 = sr.distance
				distCache[0] = d0
				distCache[4] = d0 - 1
				distCache[5] = d0 + 1
				distCache[6] = d0 - 2
				distCache[7] = d0 + 2
				distCache[8] = d0 - 3
				distCache[9] = d0 + 3
			}

			{
				delta := uint32(uint8(int8(sr.lenCodeDelta)))
				distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
				effectiveCopyLen := uint(int(sr.len) + sr.lenCodeDelta)
				insCode := getInsertLenCode(insertLength)
				copyCode := getCopyLenCode(effectiveCopyLen)
				cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
				s.appendCommand(uint32(insertLength), uint32(sr.len)|(delta<<25), distExtra, cmdPrefix, distPrefix)
			}
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
					posJump := min(position+16, posEnd-max(h5bHashTypeLength-1, 4))
					for position < posJump {
						h.storeNoWrap(data, position)
						insertLength += 4
						position += 4
					}
				} else {
					posJump := min(position+8, posEnd-(h5bHashTypeLength-1))
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

// findLongestMatch is the no-wrap, no-prefetch hot path. It mirrors
// h5b6.findLongestMatchNoWrap structurally — Phase 1 unrolled distance
// cache, Phase 2 bucket scan, Phase 3 dictionary fallback — but elides the
// speculative cur+1 prefetch (cache pressure is low at this size hint) and
// reads bucket entries as uint16 (lossless under the dispatch contract).
func (h *h5b6u16) findLongestMatch(
	data []byte,
	distCache *[16]uint,
	cur, maxLength, maxBackward, dictDistance uint,
	dictNumLookups, dictNumMatches *uint,
	out *hasherSearchResult,
) {
	bestScore := out.score
	bestLen := out.len
	key := h.hash(data, cur)
	bucket := h.bucketAt(key)
	n := h.num[key]

	out.len = 0
	out.lenCodeDelta = 0

	// Phase 1: try cached distances.
	// curByte caches loadByte(data, cur+bestLen) so the byte pre-check
	// reuses a register across iterations; refresh it whenever bestLen changes.
	curByte := loadByte(data, cur+bestLen)
	backward := distCache[0]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					bestScore = score
					bestLen = ml
					out.len = bestLen
					out.distance = backward
					out.score = bestScore
					curByte = loadByte(data, cur+bestLen)
				}
			}
		}
	}
	backward = distCache[1]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 || ml == 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[2]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[3]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 43
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[4]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[5]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 39
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[6]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[7]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 47
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[8]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.distance = backward
						out.score = bestScore
						curByte = loadByte(data, cur+bestLen)
					}
				}
			}
		}
	}
	backward = distCache[9]
	if backward-1 < maxBackward {
		prev := cur - backward
		if curByte == loadByte(data, prev+bestLen) {
			ml := uint(matchLenAtNoInline(data, prev, cur, int(maxLength)))
			if ml >= 3 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					score -= 49
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

	// Phase 2: scan hash bucket entries. uint16 entries — zero-extension is
	// lossless because positions never exceed the 64 KiB dispatch contract.
	down := uint(0)
	if uint(n) > h5b6BlockSize {
		down = uint(n) - h5b6BlockSize
	}
	minPrev := cur - maxBackward
	curProbe := loadU32LE(data, cur+bestLen-3)
	for i := uint(n); i > down; {
		i--
		prevRaw := uint(bucket[i&h5b6BlockMask])
		if prevRaw < minPrev {
			break
		}
		if curProbe != loadU32LE(data, prevRaw+bestLen-3) {
			continue
		}

		ml := uint(matchLenAtNoInline(data, prevRaw, cur, int(maxLength)))
		if ml >= 4 {
			backward := cur - prevRaw
			score := backwardReferenceScore(ml, backward)
			if bestScore < score {
				bestScore = score
				bestLen = ml
				out.len = bestLen
				out.distance = backward
				out.score = bestScore
				curProbe = loadU32LE(data, cur+bestLen-3)
			}
		}
	}

	// Store current position in the bucket.
	bucket[h.num[key]&h5b6BlockMask] = uint16(cur)
	h.num[key]++

	// Phase 3: static dictionary fallback when no hash match was found.
	if bestScore == minScore {
		searchStaticDictionaryDeep(data[cur:], maxLength, dictDistance, maxBackwardDistance,
			dictNumLookups, dictNumMatches, out)
	}
}
