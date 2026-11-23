// Encoder-side Huffman tree construction and bit-depth → symbol code conversion.

package encoder

import (
	"math"
	"math/bits"
	"slices"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// Compile-time assertion: createHuffmanTree reinterprets []huffmanTreeNode as
// []uint64 for sorting, so the struct must be exactly 8 bytes.
var _ [unsafe.Sizeof(huffmanTreeNode{})]struct{} = [8]struct{}{}

// huffmanTreeNode is a node in a Huffman tree used during tree construction.
type huffmanTreeNode struct {
	totalCount   uint32
	left         int16
	rightOrValue int16
}

// --- Tree construction ---

// createHuffmanTree builds a Huffman tree from symbol frequencies and writes
// the resulting bit depths into depth[].
//
// The tree cannot be arbitrarily deep. Brotli specifies a maximum depth of
// 15 bits for "code trees" and 7 bits for "code length code trees."
//
// countLimit is faked as the minimum value and raised until the tree fits
// within treeLimit.
//
// This algorithm is not of excellent performance for very long data blocks,
// especially when population counts are longer than 2**treeLimit, but
// we are not planning to use this with extremely long blocks.
//
// See http://en.wikipedia.org/wiki/Huffman_coding
func createHuffmanTree(data []uint32, treeLimit int, tree []huffmanTreeNode, depth []byte) {
	sentinel := huffmanTreeNode{totalCount: math.MaxUint32, left: -1, rightOrValue: -1}

	// For block sizes below 64 kB, we never need to do a second iteration
	// of this loop. Probably all of our block sizes will be smaller than
	// that, so this loop is mostly of academic interest. If we actually
	// would need this, we would be better off with the Katajainen algorithm.
	for countLimit := uint32(1); ; countLimit *= 2 {
		n := 0
		for i := len(data) - 1; i >= 0; i-- {
			c := data[i]
			if c != 0 {
				tree[n] = huffmanTreeNode{
					totalCount:   max(c, countLimit),
					left:         -1,
					rightOrValue: int16(i),
				}
				n++
			}
		}

		if n == 1 {
			depth[tree[0].rightOrValue] = 1 // Only one element.
			break
		}
		if n == 2 {
			depth[tree[0].rightOrValue] = 1
			depth[tree[1].rightOrValue] = 1
			break
		}

		// Sort leaf nodes by (totalCount ASC, rightOrValue DESC).
		// Pack each node into a uint64 sort key so we can use slices.Sort
		// (direct integer comparison, no closure overhead) instead of
		// slices.SortFunc. huffmanTreeNode is 8 bytes == uint64, so we
		// reinterpret the same backing memory.
		sortKeys := unsafe.Slice((*uint64)(unsafe.Pointer(&tree[0])), n)
		for i := 0; i < n; i++ {
			tc := tree[i].totalCount
			rv := uint16(tree[i].rightOrValue)
			sortKeys[i] = uint64(tc)<<32 | uint64(math.MaxInt16-rv)<<16 | uint64(rv)
		}
		slices.Sort(sortKeys)
		for i := 0; i < n; i++ {
			packed := sortKeys[i]
			tree[i] = huffmanTreeNode{
				totalCount:   uint32(packed >> 32),
				left:         -1,
				rightOrValue: int16(packed & 0xFFFF),
			}
		}

		// The nodes are:
		//   [0, n): the sorted leaf nodes that we start with.
		//   [n]: we add a sentinel here.
		//   [n + 1, 2n): new parent nodes are added here, starting from
		//                (n+1). These are naturally in ascending order.
		//   [2n]: we add a sentinel at the end as well.
		//   There will be (2n+1) elements at the end.
		tree[n] = sentinel
		tree[n+1] = sentinel

		i := 0     // Points to the next leaf node.
		j := n + 1 // Points to the next non-leaf node.
		for k := n - 1; k != 0; k-- {
			var left, right int
			if tree[i].totalCount <= tree[j].totalCount {
				left = i
				i++
			} else {
				left = j
				j++
			}

			if tree[i].totalCount <= tree[j].totalCount {
				right = i
				i++
			} else {
				right = j
				j++
			}

			// The sentinel node becomes the parent node.
			jEnd := 2*n - k
			tree[jEnd] = huffmanTreeNode{
				totalCount:   tree[left].totalCount + tree[right].totalCount,
				left:         int16(left),
				rightOrValue: int16(right),
			}

			// Add back the last sentinel node.
			tree[jEnd+1] = sentinel
		}

		if setDepth(tree, depth, 2*n-1, treeLimit) {
			// We need to pack the Huffman tree in treeLimit bits. If this was not
			// successful, add fake entities to the lowest values and retry.
			break
		}
	}
}

// --- Tree encoding (RLE + code length serialization) ---

// optimizeHuffmanCountsForRLE adjusts population counts so that the
// subsequent Huffman tree compression (especially its RLE part) is more
// likely to compress efficiently.
//
// goodForRLEBuf is a reusable scratch buffer to avoid per-call allocation;
// callers pass the same pointer across calls so the backing array is reused.
func optimizeHuffmanCountsForRLE(counts []uint32, goodForRLEBuf *[]bool) {
	streakLimit := 1240

	// Single forward pass: find trimmed length, count nonzeros, find smallest.
	length := 0
	nonzeroCount := 0
	smallestNonzero := uint32(1 << 30)
	for i, c := range counts {
		if c != 0 {
			length = i + 1
			nonzeroCount++
			if c < smallestNonzero {
				smallestNonzero = c
			}
		}
	}

	if nonzeroCount < 16 {
		return
	}

	if smallestNonzero < 4 {
		zeros := length - nonzeroCount
		if zeros < 6 {
			for i := 1; i < length-1; i++ {
				if counts[i-1] != 0 && counts[i] == 0 && counts[i+1] != 0 {
					counts[i] = 1
				}
			}
		}
	}

	if nonzeroCount < 28 {
		return
	}

	// Mark all population counts that already can be encoded with an RLE code.
	var goodForRLE []bool
	if cap(*goodForRLEBuf) < length {
		*goodForRLEBuf = make([]bool, length)
	} else {
		*goodForRLEBuf = (*goodForRLEBuf)[:length]
		clear(*goodForRLEBuf)
	}
	goodForRLE = *goodForRLEBuf

	// Don't spoil any of the existing good RLE codes.
	// Mark any seq of 0's longer than 5 as goodForRLE.
	// Mark any seq of non-0's longer than 7 as goodForRLE.
	sym := counts[0]
	step := 0
	for i := 0; i <= length; i++ {
		if i == length || counts[i] != sym {
			if (sym == 0 && step >= 5) || (sym != 0 && step >= 7) {
				for k := 0; k < step; k++ {
					goodForRLE[i-k-1] = true
				}
			}

			step = 1
			if i != length {
				sym = counts[i]
			}
		} else {
			step++
		}
	}

	// Replace population counts that lead to more RLE codes.
	// Math here is in 24.8 fixed point representation.
	stride := 0
	limit := int(256*(counts[0]+counts[1]+counts[2])/3 + 420)
	sum := 0
	for i := 0; i <= length; i++ {
		breakStride := i == length
		if !breakStride {
			val := int(256*counts[i]) - limit
			breakStride = goodForRLE[i] || (i != 0 && goodForRLE[i-1]) || val < -streakLimit || val >= streakLimit
		}
		if breakStride {
			if stride >= 4 || (stride >= 3 && sum == 0) {
				count := (sum + stride/2) / stride
				// The stride must end, collapse what we have, if we have enough (4).
				if count == 0 {
					count = 1
				}
				if sum == 0 {
					// Don't make an all zeros stride to be upgraded to ones.
					count = 0
				}

				for k := 0; k < stride; k++ {
					// We don't want to change value at counts[i],
					// that is already belonging to the next stride. Thus - 1.
					counts[i-k-1] = uint32(count)
				}
			}

			stride = 0
			sum = 0
			switch {
			case i < length-2:
				// All interesting strides have a count of at least 4,
				// at least when non-zeros.
				limit = int(256*(counts[i]+counts[i+1]+counts[i+2])/3 + 420)
			case i < length:
				limit = int(256 * counts[i])
			default:
				limit = 0
			}
		}

		stride++
		if i != length {
			sum += int(counts[i])
			if stride >= 4 {
				// float64 division is ~3× faster than IDIVQ for variable
				// divisors, and exact for our input range (numerator < 2^25,
				// denominator < 2^9, quotient < 2^17 — all exactly
				// representable in float64's 52-bit mantissa).
				limit = int(float64(256*sum+stride/2) / float64(stride))
			}
			if stride == 4 {
				limit += 120
			}
		}
	}
}

// encodeHuffmanTreeRepetitions encodes repetitions of value into tree/extraBitsData
// starting at position 0. Returns the number of elements written.
func encodeHuffmanTreeRepetitions(tree, extraBitsData []byte, prevValue, value byte, repetitions int) int {
	assert(repetitions > 0)
	n := 0
	if prevValue != value {
		tree[n] = value
		extraBitsData[n] = 0
		n++
		repetitions--
	}

	if repetitions == 7 {
		tree[n] = value
		extraBitsData[n] = 0
		n++
		repetitions--
	}

	if repetitions < 3 {
		for range repetitions {
			tree[n] = value
			extraBitsData[n] = 0
			n++
		}
	} else {
		start := n
		repetitions -= 3
		for {
			tree[n] = core.RepeatPreviousCodeLength
			extraBitsData[n] = byte(repetitions & 0x3)
			n++
			repetitions >>= 2
			if repetitions == 0 {
				break
			}
			repetitions--
		}

		slices.Reverse(tree[start:n])
		slices.Reverse(extraBitsData[start:n])
	}
	return n
}

// encodeHuffmanTreeRepetitionsZeros encodes repetitions of zero into tree/extraBitsData
// starting at position 0. Returns the number of elements written.
func encodeHuffmanTreeRepetitionsZeros(tree, extraBitsData []byte, repetitions int) int {
	n := 0
	if repetitions == 11 {
		tree[n] = 0
		extraBitsData[n] = 0
		n++
		repetitions--
	}

	if repetitions < 3 {
		for range repetitions {
			tree[n] = 0
			extraBitsData[n] = 0
			n++
		}
	} else {
		start := n
		repetitions -= 3
		for {
			tree[n] = alphabetSizeRepeatZeroCodeLength
			extraBitsData[n] = byte(repetitions & 0x7)
			n++
			repetitions >>= 3
			if repetitions == 0 {
				break
			}
			repetitions--
		}

		slices.Reverse(tree[start:n])
		slices.Reverse(extraBitsData[start:n])
	}
	return n
}

// findRunEnd returns the first index in depth[start:] where the value
// changes from depth[start-1], i.e., the exclusive end of a run. It scans
// 8 bytes at a time using loadU64LE for speed, then handles any remainder
// byte-by-byte. Equivalent to:
//
//	i := start
//	for i < len(depth) && depth[i] == value { i++ }
//	return i
func findRunEnd(depth []byte, start int, value byte) int {
	v64 := uint64(value) * 0x0101010101010101
	n := start
	for ; n+8 <= len(depth); n += 8 {
		diff := loadU64LE(depth, uint(n)) ^ v64
		if diff != 0 {
			return n + bits.TrailingZeros64(diff)/8
		}
	}
	for ; n < len(depth) && depth[n] == value; n++ {
	}
	return n
}

// decideOverRLEUse examines depth and returns whether RLE encoding
// should be used for nonzero and zero runs respectively.
func decideOverRLEUse(depth []byte) (useNonZero, useZero bool) {
	totalRepsZero := 0
	totalRepsNonZero := 0
	countRepsZero := 1
	countRepsNonZero := 1

	for i := 0; i < len(depth); {
		value := depth[i]
		end := findRunEnd(depth, i+1, value)
		reps := end - i

		if reps >= 3 && value == 0 {
			totalRepsZero += reps
			countRepsZero++
		}
		if reps >= 4 && value != 0 {
			totalRepsNonZero += reps
			countRepsNonZero++
		}

		i = end
	}

	return totalRepsNonZero > countRepsNonZero*2, totalRepsZero > countRepsZero*2
}

// encodeHuffmanTree encodes a Huffman tree from bit depths into the bit-stream
// representation of a Huffman tree. The generated Huffman tree is to be
// compressed once more using a Huffman tree. Returns the number of elements written.
func encodeHuffmanTree(depth, tree, extraBitsData []byte) int {
	prevValue := byte(core.InitialRepeatedCodeLength)
	useRLENonZero := false
	useRLEZero := false

	// Trim trailing zeros.
	newLength := len(depth)
	for newLength > 0 && depth[newLength-1] == 0 {
		newLength--
	}

	// First gather statistics on if it is a good idea to do RLE.
	if len(depth) > 50 {
		// Find RLE coding for longer codes.
		// Shorter codes seem not to benefit from RLE.
		useRLENonZero, useRLEZero = decideOverRLEUse(depth[:newLength])
	}

	// Actual RLE coding.
	size := 0
	for i := 0; i < newLength; {
		value := depth[i]
		reps := 1
		if (value != 0 && useRLENonZero) || (value == 0 && useRLEZero) {
			end := findRunEnd(depth[:newLength], i+1, value)
			reps = end - i
		}

		if value == 0 {
			size += encodeHuffmanTreeRepetitionsZeros(tree[size:], extraBitsData[size:], reps)
		} else {
			size += encodeHuffmanTreeRepetitions(tree[size:], extraBitsData[size:], prevValue, value, reps)
			prevValue = value
		}

		i += reps
	}
	return size
}

// --- Bit depth / symbol conversion ---

func reverseBits(numBits int, val uint16) uint16 {
	return bits.Reverse16(val) >> uint(16-numBits)
}

// convertBitDepthsToSymbols generates the actual bit values for a tree of
// bit depths. In Brotli, all bit depths are [1..15]; 0 means the symbol
// does not exist.
func convertBitDepthsToSymbols(depth []byte, symbols []uint16) {
	var blCount [maxHuffmanBits]uint16
	var nextCode [maxHuffmanBits]uint16

	// Build the histogram using 4-way interleaving to reduce store-to-load
	// forwarding stalls when consecutive depth values map to the same bucket.
	// Each sub-histogram (bc1..bc3) is independent, so the CPU can pipeline
	// four scatter-increments per iteration without RAW hazards.
	code := 0
	{
		var bc1, bc2, bc3 [maxHuffmanBits]uint16
		i, n := 0, len(depth)
		for ; i+3 < n; i += 4 {
			blCount[depth[i]&0x0F]++
			bc1[depth[i+1]&0x0F]++
			bc2[depth[i+2]&0x0F]++
			bc3[depth[i+3]&0x0F]++
		}
		for ; i < n; i++ {
			blCount[depth[i]&0x0F]++
		}
		for j := range blCount {
			blCount[j] += bc1[j] + bc2[j] + bc3[j]
		}
	}

	blCount[0] = 0
	nextCode[0] = 0
	for i := 1; i < maxHuffmanBits; i++ {
		code = (code + int(blCount[i-1])) << 1
		nextCode[i] = uint16(code)
	}

	// Narrow symbols to the exact length so the compiler sees
	// len(symbols)==len(depth) and can eliminate the symbols[i] check.
	symbols = symbols[:len(depth)]
	for i, d := range depth {
		if d != 0 {
			dm := d & 0x0F
			symbols[i] = reverseBits(int(dm), nextCode[dm])
			nextCode[dm]++
		}
	}
}

// setDepth walks the Huffman tree rooted at pool[p0] and writes the
// depth of each leaf into depth[]. Returns false if any leaf exceeds maxDepth.
func setDepth(pool []huffmanTreeNode, depth []byte, p0, maxDepth int) bool {
	var stack [16]int
	level := 0
	p := p0
	assert(maxDepth <= 15)
	stack[0] = -1
	for {
		if pool[p].left >= 0 {
			level++
			if level > maxDepth {
				return false
			}
			stack[level] = int(pool[p].rightOrValue)
			p = int(pool[p].left)
			continue
		}
		depth[pool[p].rightOrValue] = byte(level)

		for level >= 0 && stack[level] == -1 {
			level--
		}
		if level < 0 {
			return true
		}
		p = stack[level]
		stack[level] = -1
	}
}
