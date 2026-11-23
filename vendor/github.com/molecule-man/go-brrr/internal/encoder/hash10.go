// H10 binary tree hasher for quality 10–11 (Zopfli).
//
// H10 is a hash table where each bucket contains a binary search tree of
// sequences whose first 4 bytes share the same hash code. Each sequence is
// up to h10MaxTreeCompLength (128) bytes long and is identified by its
// starting position in the input data. The binary tree is sorted by the
// lexicographic order of the sequences, and it is also a max-heap with
// respect to starting positions (newer positions are always ancestors of
// older ones).
//
// Unlike the bucket-chain hashers (H5/H6) which return the single best
// match, H10 returns all matches at a position sorted by increasing length.
// This match set is consumed by the Zopfli optimal parsing algorithm.

package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// H10 configuration constants.
const (
	h10BucketBits = 17
	h10BucketSize = 1 << h10BucketBits // 131,072

	// h10MaxTreeSearchDepth is the maximum number of tree nodes examined
	// per storeAndFindMatches call, bounding worst-case search time.
	h10MaxTreeSearchDepth = 64

	// h10MaxTreeCompLength is the maximum number of bytes compared per
	// tree node. Also used as the StoreLookahead for H10: positions can
	// only be inserted into the tree when at least this many bytes remain.
	h10MaxTreeCompLength = 128

	// h10MaxNumMatches is the maximum number of matches returned by
	// findAllMatches (64 short-range + 64 tree matches).
	h10MaxNumMatches = 128

	h10HashShift = 32 - h10BucketBits // 15
)

// h10 is the H10 binary tree hasher.
//
// Each hash bucket is the root of a binary search tree keyed by the
// lexicographic order of the byte sequences at stored positions. The tree
// is also a max-heap on position: every node's position is greater than its
// children's, so a single root-to-leaf traversal both searches for matches
// and re-roots the tree at the current position.
//
// The forest stores left/right child pointers for every position in the
// sliding window: forest[2*(pos & windowMask)] is the left child,
// forest[2*(pos & windowMask)+1] is the right child.
type h10 struct {
	bufs    *q10Bufs // reusable scratch buffers for Zopfli DP
	forest  []uint32
	lgwin   int
	quality int
	hasherCommon
	windowMask uint32
	invalidPos uint32
	buckets    [h10BucketSize]uint32
	skipDict   bool
}

func (h *h10) common() *hasherCommon {
	return &h.hasherCommon
}

// reset initializes the hasher for a new compression session.
// All bucket roots are set to invalidPos (sentinel for empty tree).
func (h *h10) reset(_ bool, inputSize uint, _ []byte) {
	lgwin := h.lgwin
	h.windowMask = (1 << lgwin) - 1
	h.invalidPos = 0 - h.windowMask

	numNodes := min(uint(1)<<lgwin, inputSize)
	if len(h.forest) < int(2*numNodes) {
		h.forest = make([]uint32, 2*numNodes)
	}

	for i := range h.buckets {
		h.buckets[i] = h.invalidPos
	}
	h.ready = true
}

// storeAndFindMatches is the core operation of the binary tree hasher.
// In a single tree traversal it simultaneously:
//  1. Searches for matches longer than *bestLen
//  2. Re-roots the binary tree at curIx
//  3. Appends found matches to the matches slice
//
// When maxLength < h10MaxTreeCompLength, the tree is searched but not
// modified because the incomplete sequence cannot be correctly ordered.
//
// Returns the number of matches written to matches.
func (h *h10) storeAndFindMatches(
	data []byte, curIx, ringBufferMask, maxLength, maxBackward uint,
	bestLen *uint, matches []backwardMatch,
) int {
	curIxMasked := curIx & ringBufferMask

	key := h.hash(data, curIxMasked)
	prevIx := uint(h.buckets[key])
	h.buckets[key] = uint32(curIx)

	// Hoisted so the loop keeps the forest header and mask in registers:
	// without restrict, Go re-loads them from h after every forest store.
	forest := h.forest
	mask := uint(h.windowMask)
	invalidPos := h.invalidPos

	// nodeLeft/nodeRight track where to attach subtrees as the tree is
	// re-rooted. They are forest indices, not positions.
	nodeLeft := 2 * (curIx & mask)
	nodeRight := nodeLeft + 1

	// bestLenLeft/bestLenRight are the known match lengths of the
	// boundary nodes of the left and right subtrees being built.
	var bestLenLeft, bestLenRight uint

	nMatches := 0

	for depth := h10MaxTreeSearchDepth; ; depth-- {
		backward := curIx - prevIx
		prevIxMasked := prevIx & ringBufferMask

		if backward == 0 || backward > maxBackward || depth == 0 {
			forest[nodeLeft] = invalidPos
			forest[nodeRight] = invalidPos
			break
		}

		curLen := min(bestLenLeft, bestLenRight)
		length := curLen + uint(matchLenAt(
			data,
			curIxMasked+curLen,
			prevIxMasked+curLen,
			int(maxLength-curLen),
		))

		if length > *bestLen {
			*bestLen = length
			matches[nMatches] = newBackwardMatch(backward, length)
			nMatches++
		}

		if length >= h10MaxTreeCompLength {
			// Full match up to comparison limit: steal the old node's children.
			forest[nodeLeft] = forest[2*(prevIx&mask)]
			forest[nodeRight] = forest[2*(prevIx&mask)+1]
			break
		}

		// Lexicographic comparison determines left vs right subtree placement.
		if data[curIxMasked+length] > data[prevIxMasked+length] {
			bestLenLeft = length
			forest[nodeLeft] = uint32(prevIx)
			nodeLeft = 2*(prevIx&mask) + 1
			prevIx = uint(forest[nodeLeft])
		} else {
			bestLenRight = length
			forest[nodeRight] = uint32(prevIx)
			nodeRight = 2 * (prevIx & mask)
			prevIx = uint(forest[nodeRight])
		}
	}

	return nMatches
}

// storeOnly is the match-free twin of storeAndFindMatches for store and
// stitchToPreviousBlock, which re-root the tree at ix without reporting
// matches. Dropping the match bookkeeping keeps four fewer values live in
// the tree walk.
func (h *h10) storeOnly(data []byte, curIx, ringBufferMask, maxBackward uint) {
	curIxMasked := curIx & ringBufferMask

	key := h.hash(data, curIxMasked)
	prevIx := uint(h.buckets[key])
	h.buckets[key] = uint32(curIx)

	forest := h.forest
	mask := uint(h.windowMask)
	invalidPos := h.invalidPos

	nodeLeft := 2 * (curIx & mask)
	nodeRight := nodeLeft + 1

	var bestLenLeft, bestLenRight uint

	for depth := h10MaxTreeSearchDepth; ; depth-- {
		backward := curIx - prevIx
		prevIxMasked := prevIx & ringBufferMask

		if backward == 0 || backward > maxBackward || depth == 0 {
			forest[nodeLeft] = invalidPos
			forest[nodeRight] = invalidPos
			break
		}

		curLen := min(bestLenLeft, bestLenRight)
		length := curLen + uint(matchLenAt(
			data,
			curIxMasked+curLen,
			prevIxMasked+curLen,
			h10MaxTreeCompLength-int(curLen),
		))

		if length >= h10MaxTreeCompLength {
			forest[nodeLeft] = forest[2*(prevIx&mask)]
			forest[nodeRight] = forest[2*(prevIx&mask)+1]
			break
		}

		if data[curIxMasked+length] > data[prevIxMasked+length] {
			bestLenLeft = length
			forest[nodeLeft] = uint32(prevIx)
			nodeLeft = 2*(prevIx&mask) + 1
			prevIx = uint(forest[nodeLeft])
		} else {
			bestLenRight = length
			forest[nodeRight] = uint32(prevIx)
			nodeRight = 2 * (prevIx & mask)
			prevIx = uint(forest[nodeRight])
		}
	}
}

// findMatchesNoStore is the search-only twin of storeAndFindMatches for the
// last positions of a block, where fewer than h10MaxTreeCompLength bytes
// remain: the sequence cannot be ordered, so the tree is read but never
// re-rooted, and the comparison limit is the remaining length itself.
func (h *h10) findMatchesNoStore(
	data []byte, curIx, ringBufferMask, maxLength, maxBackward uint,
	bestLen *uint, matches []backwardMatch,
) int {
	curIxMasked := curIx & ringBufferMask

	key := h.hash(data, curIxMasked)
	prevIx := uint(h.buckets[key])

	forest := h.forest
	mask := uint(h.windowMask)

	var bestLenLeft, bestLenRight uint

	nMatches := 0

	for depth := h10MaxTreeSearchDepth; ; depth-- {
		backward := curIx - prevIx
		prevIxMasked := prevIx & ringBufferMask

		if backward == 0 || backward > maxBackward || depth == 0 {
			break
		}

		curLen := min(bestLenLeft, bestLenRight)
		length := curLen + uint(matchLenAt(
			data,
			curIxMasked+curLen,
			prevIxMasked+curLen,
			int(maxLength-curLen),
		))

		if length > *bestLen {
			*bestLen = length
			matches[nMatches] = newBackwardMatch(backward, length)
			nMatches++
		}

		if length >= maxLength {
			break
		}

		if data[curIxMasked+length] > data[prevIxMasked+length] {
			bestLenLeft = length
			prevIx = uint(forest[2*(prevIx&mask)+1])
		} else {
			bestLenRight = length
			prevIx = uint(forest[2*(prevIx&mask)])
		}
	}

	return nMatches
}

// findAllMatches finds all backward matches at curIx and stores curIx in the
// hash table. Matches are sorted by strictly increasing length and
// non-strictly increasing distance.
//
// The search proceeds in three phases:
//  1. Short-match scan: linear brute-force backward search for 2-byte prefix
//     matches (up to 16 positions for Q10, 64 for Q11).
//  2. Tree search: calls storeAndFindMatches for longer matches via the
//     binary search tree.
//  3. Static dictionary: searches the RFC 7932 static dictionary for matches
//     longer than the best LZ77 match.
//
// The matches slice must have capacity for at least h10MaxNumMatches entries.
// Returns the number of matches found.
func (h *h10) findAllMatches(
	data []byte, ringBufferMask, curIx, maxLength, maxBackward, dictionaryDistance uint,
	quality int, matches []backwardMatch,
) uint {
	curIxMasked := curIx & ringBufferMask
	bestLen := uint(1)
	nMatches := 0

	// Phase 1: Short-match brute-force scan.
	// Quality 10 searches 16 positions back; quality 11 searches 64.
	shortMatchMaxBackward := uint(16)
	if quality == 11 {
		shortMatchMaxBackward = 64
	}

	stop := uint(0)
	if curIx > shortMatchMaxBackward {
		stop = curIx - shortMatchMaxBackward
	}

	if prefix2Mask64Available &&
		curIxMasked >= 64 && curIx > 64 && maxBackward >= shortMatchMaxBackward-1 {
		mask := prefix2Mask64(&data[curIxMasked-64], data[curIxMasked], data[curIxMasked+1])
		mask &= ^uint64(0) << (65 - shortMatchMaxBackward)
		for mask != 0 && bestLen <= 2 {
			j := uint(63 - bits.LeadingZeros64(mask))
			mask &^= 1 << j
			backward := 64 - j
			prevIxMasked := curIxMasked - backward
			length := uint(matchLenAt(data, prevIxMasked, curIxMasked, int(maxLength)))
			if length > bestLen {
				bestLen = length
				matches[nMatches] = newBackwardMatch(backward, length)
				nMatches++
			}
		}
	} else {
		for i := curIx - 1; i > stop && bestLen <= 2; i-- {
			backward := curIx - i
			if backward > maxBackward {
				break
			}
			prevIxMasked := i & ringBufferMask
			if data[curIxMasked] != data[prevIxMasked] ||
				data[curIxMasked+1] != data[prevIxMasked+1] {
				continue
			}
			length := uint(matchLenAt(data, prevIxMasked, curIxMasked, int(maxLength)))
			if length > bestLen {
				bestLen = length
				matches[nMatches] = newBackwardMatch(backward, length)
				nMatches++
			}
		}
	}

	// Phase 2: Tree search for longer matches.
	if bestLen < maxLength {
		if maxLength >= h10MaxTreeCompLength {
			nMatches += h.storeAndFindMatches(
				data, curIx, ringBufferMask, maxLength, maxBackward,
				&bestLen, matches[nMatches:],
			)
		} else {
			nMatches += h.findMatchesNoStore(
				data, curIx, ringBufferMask, maxLength, maxBackward,
				&bestLen, matches[nMatches:],
			)
		}
	}

	// Phase 3: Static dictionary search.
	// Search the RFC 7932 static dictionary for matches at all lengths
	// longer than the best LZ77 match found so far. Each length's best
	// dictionary match is converted to a backwardMatch.
	if !h.skipDict {
		nMatches += staticDictBackwardMatches(data, curIxMasked, bestLen, maxLength, dictionaryDistance, matches[nMatches:])
	}

	return uint(nMatches)
}

func staticDictBackwardMatches(data []byte, curIxMasked, bestLen, maxLength, dictionaryDistance uint, matches []backwardMatch) int {
	nMatches := 0
	minLen := max(uint(4), bestLen+1)
	maxLen := min(uint(maxStaticDictMatchLen), maxLength)
	if minLen <= maxLen {
		var dictMatches [maxStaticDictMatchLen + 1]uint32
		for i := range dictMatches {
			dictMatches[i] = invalidMatch
		}
		if findAllStaticDictionaryMatches(data[curIxMasked:], minLen, maxLength, dictMatches[:]) {
			for l := minLen; l <= maxLen; l++ {
				dictID := dictMatches[l]
				if dictID < invalidMatch {
					distance := dictionaryDistance + uint(dictID>>5) + 1
					if distance <= maxBackwardDistance {
						matches[nMatches] = newDictionaryBackwardMatch(distance, l, uint(dictID&31))
						nMatches++
					}
				}
			}
		}
	}
	return nMatches
}

// store records position ix in the binary tree without returning matches.
// Requires that at least h10MaxTreeCompLength bytes are available at ix.
func (h *h10) store(data []byte, mask, ix uint) {
	// Maximum distance is window size - 16 (RFC 7932 Section 9.1).
	maxBackward := uint(h.windowMask) - core.WindowGap + 1
	h.storeOnly(data, ix, mask, maxBackward)
}

// storeRange stores positions ixStart..ixEnd-1 in the binary tree.
// For large ranges, a sparse prefix (every 8th position) is stored first,
// followed by a dense tail of the last 63 positions.
func (h *h10) storeRange(data []byte, mask, ixStart, ixEnd uint) {
	i := ixStart
	j := ixStart

	// Dense tail: always store the last 63 positions.
	if ixStart+63 <= ixEnd {
		i = ixEnd - 63
	}
	// Sparse prefix: store every 8th position if the range is large enough.
	if ixStart+512 <= i {
		for ; j < i; j += 8 {
			h.store(data, mask, j)
		}
	}
	// Dense tail.
	for ; i < ixEnd; i++ {
		h.store(data, mask, i)
	}
}

// stitchToPreviousBlock stores positions from the end of the previous block
// that could not be stored earlier because they required data from the
// current block (the sequence at those positions spans the block boundary).
func (h *h10) stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint) {
	h.growForest(position + numBytes)
	// Need at least 3 bytes (hashTypeLength - 1 = 4 - 1) and the position
	// must be past the initial StoreLookahead region.
	if numBytes < 3 || position < h10MaxTreeCompLength {
		return
	}

	iStart := position - h10MaxTreeCompLength + 1
	iEnd := min(position, iStart+numBytes)

	for i := iStart; i < iEnd; i++ {
		// Maximum distance is window size - 16 (RFC 7932 Section 9.1).
		// Also ensure we don't look further back than the start of the
		// current block to avoid reading overwritten ring buffer data.
		maxBackward := uint(h.windowMask) - max(core.WindowGap-1, position-i)
		h.storeOnly(ringBuffer, i, ringBufferMask, maxBackward)
	}
}

func (h *h10) growForest(end uint) {
	window := uint(h.windowMask) + 1
	need := 2 * min(window, end)
	if uint(len(h.forest)) >= need {
		return
	}
	forest := make([]uint32, min(max(need, 2*uint(len(h.forest))), 2*window))
	copy(forest, h.forest)
	h.forest = forest
}

// createBackwardReferences runs the Zopfli optimal parsing algorithm to
// find backward references for the given input range.
//
// This bridges the streamHasher interface to the Zopfli DP entry point,
// converting the encodeState's [4]uint distance cache to the []int form
// that the Zopfli functions use.
func (h *h10) createBackwardReferences(s *encodeState, bytes, wrappedPos uint32) {
	distCache := [4]int{int(s.distCache[0]), int(s.distCache[1]), int(s.distCache[2]), int(s.distCache[3])}

	origCmdCount := len(s.commands)
	gap := s.compound.totalSize
	if s.quality == 11 || h.bufs.parallel {
		createHqZopfliBackwardReferences(uint(bytes), uint(wrappedPos), s.data, uint(s.mask),
			s.quality, s.lgwin, gap, &s.compound, distCache[:], h, &s.lastInsertLen, &s.commands, &s.numLiterals, h.bufs)
	} else {
		createZopfliBackwardReferences(uint(bytes), uint(wrappedPos), s.data, uint(s.mask),
			s.quality, s.lgwin, gap, &s.compound, distCache[:], h, &s.lastInsertLen, &s.commands, &s.numLiterals, h.bufs)
	}

	s.distCache = [4]uint{uint(distCache[0]), uint(distCache[1]), uint(distCache[2]), uint(distCache[3])}
	s.numCommands += uint(len(s.commands) - origCmdCount)
}

// hash computes a 17-bit bucket index from 4 bytes at data[i:i+4].
func (h *h10) hash(data []byte, i uint) uint32 {
	return (loadU32LE(data, i) * hashMul32) >> h10HashShift
}
