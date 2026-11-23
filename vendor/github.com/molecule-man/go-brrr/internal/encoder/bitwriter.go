package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// symbolBits is the number of bits used for the symbol value in the
// packed RLE representation:
//
//	bits [31..9]  →  extra data (only for zero-run symbols)
//	bits [8..0]   →  the Huffman symbol
const symbolBits = 9

// Brotli bitstream encoding: bit-level packing to match the spec format.
// No compression algorithms; just the right ordering of bits.
// Callers must ensure sufficient space in the output buffer.

type bitWriter struct {
	buf       []byte
	bitOffset uint
}

// writeBits is defined in bitwriter_le_unsafe.go (little-endian platforms)
// or bitwriter_le_generic.go (other platforms).

// writeHuffmanTree writes a complex prefix code to the bitstream.
// RFC 7932 Section 3.5.
//
// depths contains the code length (bit depth) for each symbol in the data
// alphabet. tree is scratch space for building the code length prefix code.
//
// This implements the three-layer encoding scheme described in huffman.go:
//
//	Bitstream output:
//	┌──────────────────────────────┬──────────────────────────────┐
//	│ Layer 0: HSKIP + code length │ Layer 1: code length symbols │
//	│ depths in transmission order │ with RLE extra bits          │
//	└──────────────────────────────┴──────────────────────────────┘
func (b *bitWriter) writeHuffmanTree(depths []byte, tree []huffmanTreeNode) {
	// RLE-encode the data alphabet code lengths into code length alphabet
	// symbols (0–17). Symbols 0–15 are literal lengths; 16 repeats the
	// previous non-zero length; 17 repeats zero.
	treeArray := [core.AlphabetSizeInsertAndCopyLength]byte{}
	extraBitsArray := [core.AlphabetSizeInsertAndCopyLength]byte{}
	treeSize := encodeHuffmanTree(depths, treeArray[:], extraBitsArray[:])

	treeBuf := treeArray[:treeSize]

	// Count frequency of each code length alphabet symbol (0–17).
	histogram := [core.AlphabetSizeCodeLengths]uint32{}
	for _, v := range treeBuf {
		histogram[v]++
	}

	// Count distinct code length symbols. If only one symbol is used,
	// record which one — single-symbol alphabets need special handling.
	numCodes := 0
	code := 0
	for i, v := range histogram {
		if v != 0 {
			if numCodes == 0 {
				code = i
				numCodes = 1
			} else if numCodes == 1 {
				numCodes = 2
				break
			}
		}
	}

	// Build the prefix code for the code length alphabet (max depth 5).
	bitDepthArray := [core.AlphabetSizeCodeLengths]byte{}
	bitDepthSymbolsArray := [core.AlphabetSizeCodeLengths]uint16{}
	createHuffmanTree(histogram[:], 5, tree, bitDepthArray[:])
	convertBitDepthsToSymbols(bitDepthArray[:], bitDepthSymbolsArray[:])

	// Write Layer 0: HSKIP field + code length depths in transmission order.
	b.writeHuffmanTreeOfHuffmanTreeToBitmask(numCodes, bitDepthArray[:])

	// Single-symbol code length alphabets: the spec expects depth 0
	// (zero bits per symbol) since no actual decoding choice exists.
	if numCodes == 1 {
		bitDepthArray[code] = 0
	}

	// Write Layer 1: each code length symbol encoded with the code length
	// prefix code, plus extra bits for RLE symbols (16 and 17).
	b.writeHuffmanTreeToBitmask(treeBuf, extraBitsArray[:treeSize], bitDepthArray[:], bitDepthSymbolsArray[:])
}

// writeHuffmanTreeOfHuffmanTreeToBitmask writes Layer 0 of the prefix code
// encoding: the HSKIP field followed by each code length alphabet symbol's
// depth, in core.CodeLengthCodeOrder transmission order, encoded with the fixed
// variable-length code from RFC 7932 Section 3.5. Trailing zero-depth entries
// are omitted.
func (b *bitWriter) writeHuffmanTreeOfHuffmanTreeToBitmask(numCodes int, bitDepth []byte) {
	codesToStore := uint(core.AlphabetSizeCodeLengths)

	// Trim trailing zeros in storage order.
	if numCodes > 1 {
		for codesToStore > 0 && bitDepth[core.CodeLengthCodeOrder[codesToStore-1]] == 0 {
			codesToStore--
		}
	}

	// Determine how many leading zero-depth entries to skip (2 or 3).
	var skip uint
	if bitDepth[core.CodeLengthCodeOrder[0]] == 0 && bitDepth[core.CodeLengthCodeOrder[1]] == 0 {
		skip = 2
		if bitDepth[core.CodeLengthCodeOrder[2]] == 0 {
			skip = 3
		}
	}

	b.writeBits(2, uint64(skip))
	for i := skip; i < codesToStore; i++ {
		depth := bitDepth[core.CodeLengthCodeOrder[i]]
		b.writeBits(
			uint(codeLengthCodeBitLengths[depth]),
			uint64(codeLengthCodeSymbols[depth]),
		)
	}
}

// writeHuffmanTreeToBitmask writes Layer 1 of the prefix code encoding: each
// code length symbol encoded with the code length prefix code built from
// Layer 0. RLE symbols carry extra bits: symbol 16 (repeat previous non-zero)
// has 2 extra bits; symbol 17 (repeat zero) has 3 extra bits.
func (b *bitWriter) writeHuffmanTreeToBitmask(tree, extraBits, bitDepth []byte, symbols []uint16) {
	for i, code := range tree {
		b.writeBits(uint(bitDepth[code]), uint64(symbols[code]))

		switch code {
		case core.RepeatPreviousCodeLength:
			b.writeBits(2, uint64(extraBits[i]))
		case alphabetSizeRepeatZeroCodeLength:
			b.writeBits(3, uint64(extraBits[i]))
		}
	}
}

// writeSimpleHuffmanTree writes a simple prefix code to the bitstream.
// RFC 7932 Section 3.4.
//
// Simple prefix codes encode alphabets with 2–4 active symbols. (NSYM=1 is
// handled directly in encodeHuffmanTree as a degenerate case.)
//
// Bitstream format:
//
//	┌───────┬────────┬──────────────────────────────────┬────────────┐
//	│ HSKIP │ NSYM-1 │ NSYM symbol values               │ tree-select│
//	│ 2 bits│ 2 bits │ each maxBits wide, sorted by     │ 1 bit      │
//	│ = 1   │        │ code length then value           │ (NSYM=4    │
//	│       │        │                                  │  only)     │
//	└───────┴────────┴──────────────────────────────────┴────────────┘
//
// maxBits (ALPHABET_BITS in the spec) is the smallest bit width that can
// represent any symbol in the alphabet.
//
// Implied code depths by NSYM:
//
//	NSYM=2: depths {1, 1}
//	NSYM=3: depths {1, 2, 2}
//	NSYM=4: tree-select=0 → depths {2, 2, 2, 2}
//	        tree-select=1 → depths {1, 2, 3, 3}
//
// depths contains code lengths for the full alphabet. symbols holds the active
// symbol indices. numSymbols is NSYM (2–4). maxBits is ALPHABET_BITS.
func (b *bitWriter) writeSimpleHuffmanTree(depths []byte, symbols [4]uint, numSymbols, maxBits uint) {
	// HSKIP=1 distinguishes simple prefix codes from complex ones.
	// Complex codes use HSKIP values 0, 2, or 3 (the skip count).
	b.writeBits(2, 1)
	// NSYM-1 in 2 bits: the decoder adds 1 to recover NSYM.
	b.writeBits(2, uint64(numSymbols-1))

	// Symbols must appear sorted by code length (then by value within
	// equal lengths) for canonical Huffman code reconstruction.
	for i := range numSymbols {
		for j := i + 1; j < numSymbols; j++ {
			if depths[symbols[j]] < depths[symbols[i]] {
				symbols[i], symbols[j] = symbols[j], symbols[i]
			}
		}
	}

	switch numSymbols {
	case 2: // implied depths {1, 1}
		b.writeBits(maxBits, uint64(symbols[0]))
		b.writeBits(maxBits, uint64(symbols[1]))
	case 3: // implied depths {1, 2, 2}
		b.writeBits(maxBits, uint64(symbols[0]))
		b.writeBits(maxBits, uint64(symbols[1]))
		b.writeBits(maxBits, uint64(symbols[2]))
	default: // NSYM=4
		b.writeBits(maxBits, uint64(symbols[0]))
		b.writeBits(maxBits, uint64(symbols[1]))
		b.writeBits(maxBits, uint64(symbols[2]))
		b.writeBits(maxBits, uint64(symbols[3]))
		// tree-select: 0 → depths {2,2,2,2}, 1 → depths {1,2,3,3}.
		v := uint64(0)
		if depths[symbols[0]] == 1 {
			v = 1
		}
		b.writeBits(1, v)
	}
}

// buildAndWriteHuffmanTree builds a Huffman tree from histogram, writes the
// prefix code to the bitstream, and returns the code lengths and code values.
//
// This is the full-quality path used by higher-quality encoding (quality 3+).
// It uses createHuffmanTree with treeLimit=15 (RFC 7932 §3.5), compared to
// buildAndWriteHuffmanTreeFast which uses a treeLimit of 14 and a simplified
// static code length shortcut.
//
// Three encoding paths are selected based on the number of distinct symbols:
//
//   - 0–1 symbols: degenerate simple prefix code (§3.4, NSYM=1, depth 0).
//   - 2–4 symbols: simple prefix code via writeSimpleHuffmanTree (§3.4).
//   - 5+  symbols: full three-layer complex prefix code via writeHuffmanTree (§3.5).
//
// alphabetSize may exceed len(histogram). For example, distance histograms
// have len=140 but alphabetSize=num_distance_symbols. maxBits (ALPHABET_BITS)
// is computed from alphabetSize, not len(histogram).
//
// The tree slice must hold at least 2*len(histogram)+1 nodes.
func (b *bitWriter) buildAndWriteHuffmanTree(
	histogram []uint32,
	alphabetSize uint,
	tree []huffmanTreeNode,
	depth []byte, codes []uint16,
) {
	n := len(histogram)
	clear(depth[:n])
	clear(codes[:n])

	// Scan for up to 4+1 distinct non-zero symbols.
	var count uint
	var s4 [4]uint
	for i, v := range histogram {
		if v != 0 {
			if count < 4 {
				s4[count] = uint(i)
			} else if count > 4 {
				break
			}
			count++
		}
	}

	// ALPHABET_BITS: smallest bit width that can represent any symbol.
	maxBits := uint(0)
	for v := alphabetSize - 1; v != 0; v >>= 1 {
		maxBits++
	}

	// Degenerate simple prefix code: single-symbol alphabet.
	if count <= 1 {
		b.writeBits(4, 1)
		b.writeBits(maxBits, uint64(s4[0]))
		depth[s4[0]] = 0
		codes[s4[0]] = 0
		return
	}

	createHuffmanTree(histogram, 15, tree, depth)
	convertBitDepthsToSymbols(depth[:n], codes[:n])

	if count <= 4 {
		b.writeSimpleHuffmanTree(depth, s4, count, maxBits)
	} else {
		b.writeHuffmanTree(depth[:n], tree)
	}
}

// buildAndWriteHuffmanTreeFast builds a Huffman tree from histogram,
// writes the encoded tree to the bitstream, and fills depth and codes
// (both len(histogram)) with the code lengths and code values respectively.
//
// histogramTotal is the sum of all histogram counts; it allows the initial
// scan to stop early once all non-zero entries have been found, avoiding
// iteration over trailing zeros.
//
// The tree slice must have space for at least 2*core.AlphabetSizeLiteral+1 nodes.
// maxBits is the bit width for encoding symbol values in simple codes (typically 8).
func (b *bitWriter) buildAndWriteHuffmanTreeFast(tree []huffmanTreeNode,
	histogram []uint32, histogramTotal, maxBits uint,
	depth []byte, codes []uint16) {

	var count uint
	var symbols [4]uint
	var length uint
	total := histogramTotal
	for total != 0 {
		if histogram[length] != 0 {
			if count < 4 {
				symbols[count] = length
			}
			count++
			total -= uint(histogram[length])
		}
		length++
	}

	b.encodeHuffmanTree(tree, histogram, depth, codes, symbols, count, length, maxBits)
}

// encodeHuffmanTree builds the Huffman tree from pre-scanned histogram data
// and writes the encoded prefix code to the bitstream. count and symbols
// come from the histogram scan; length is 1 + the index of the last non-zero
// histogram entry.
func (b *bitWriter) encodeHuffmanTree(tree []huffmanTreeNode,
	histogram []uint32, depth []byte, codes []uint16,
	symbols [4]uint, count, length, maxBits uint) {

	// Degenerate simple prefix code: a single-symbol alphabet.
	// HSKIP=1 (2 bits) + NSYM-1=0 (2 bits) = 4-bit value 1, followed by
	// the symbol value. Depth is 0 because the decoder emits the symbol
	// without consuming any bits from the data stream.
	if count <= 1 {
		b.writeBits(4, 1)
		b.writeBits(maxBits, uint64(symbols[0]))
		depth[symbols[0]] = 0
		codes[symbols[0]] = 0
		return
	}

	clear(depth[:length])

	for countLimit := uint32(1); ; countLimit *= 2 {
		n := 0
		for l := length; l != 0; {
			l--
			if histogram[l] != 0 {
				tree[n] = huffmanTreeNode{
					totalCount:   max(histogram[l], countLimit),
					left:         -1,
					rightOrValue: int16(l),
				}
				n++
			}
		}

		sentinel := huffmanTreeNode{totalCount: 1<<32 - 1, left: -1, rightOrValue: -1}

		sortHuffmanTree(tree[:n])

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
		for k := n - 1; k > 0; k-- {
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

		if setDepth(tree, depth, 2*n-1, 14) {
			break
		}
	}

	convertBitDepthsToSymbols(depth[:length], codes[:length])
	b.writeHuffmanCode(depth, symbols, count, length, maxBits)
}

// writeHuffmanCode writes a Huffman code to the bitstream. For alphabets with
// 1–4 symbols it dispatches to writeSimpleHuffmanTree (Section 3.4). For larger
// alphabets it uses a simplified complex encoding path: a static code length
// prefix code (all symbols depth 4) followed by the code lengths RLE-encoded
// directly, bypassing the full three-layer writeHuffmanTree machinery.
func (b *bitWriter) writeHuffmanCode(depth []byte, symbols [4]uint, count, length, maxBits uint) {
	if count <= 4 {
		b.writeSimpleHuffmanTree(depth, symbols, count, maxBits)
		return
	}

	b.writeStaticCodeLengthCode()

	previousValue := byte(core.InitialRepeatedCodeLength)
	for i := uint(0); i < length; {
		value := depth[i]
		reps := uint(1)
		for k := i + 1; k < length && depth[k] == value; k++ {
			reps++
		}
		i += reps
		if value == 0 {
			b.writeBits(uint(zeroRepsDepth[reps]), zeroRepsBits[reps])
		} else {
			if previousValue != value {
				b.writeBits(uint(codeLengthDepth[value]), uint64(codeLengthBits[value]))
				reps--
			}
			if reps < 3 {
				for reps != 0 {
					reps--
					b.writeBits(uint(codeLengthDepth[value]), uint64(codeLengthBits[value]))
				}
			} else {
				reps -= 3
				b.writeBits(uint(nonZeroRepsDepth[reps]), nonZeroRepsBits[reps])
			}
			previousValue = value
		}
	}
}

// writeStaticCodeLengthCode writes the pre-computed static code length prefix
// code to the bitstream. This is a flat code where all 18 code length symbols
// have depth 4 (except symbol 15 which is unused, depth 0).
func (b *bitWriter) writeStaticCodeLengthCode() {
	b.writeBits(40, 0x0000FF55555554)
}

// storeVarLenUint8 encodes a value 0–255 using a variable-length prefix code.
//
//	n == 0: writes 1 zero bit.
//	n > 0:  writes 1 one-bit, then 3 bits for floor(log2(n)), then
//	        floor(log2(n)) bits for n - 2^floor(log2(n)).
func (b *bitWriter) storeVarLenUint8(n uint) {
	if n == 0 {
		b.writeBits(1, 0)
		return
	}
	nbits := uint(bits.Len(n)) - 1 // floorLog2(n)
	b.writeBits(1, 1)
	b.writeBits(3, uint64(nbits))
	b.writeBits(nbits, uint64(n)-uint64(uint(1)<<nbits))
}

// encodeContextMap encodes a context map to the bitstream.
// The context map maps (block_type, context) pairs to histogram cluster
// indices. numClusters is the number of distinct clusters.
//
// Encoding pipeline (RFC 7932 Section 7.3):
//  1. Write numClusters-1 as a variable-length uint8.
//  2. If numClusters >= 2:
//     a. Apply move-to-front transform to exploit locality.
//     b. Run-length encode zero runs.
//     c. Write RLE parameters (use_rle flag + max prefix).
//     d. Build and write a Huffman tree for the RLE+MTF alphabet.
//     e. Write all encoded symbols with their extra bits.
//     f. Signal inverse-MTF for the decoder.
func (b *bitWriter) encodeContextMap(contextMap []uint32, numClusters uint, tree []huffmanTreeNode) {
	b.encodeContextMapBuf(contextMap, numClusters, tree, nil)
}

func (b *bitWriter) encodeContextMapBuf(contextMap []uint32, numClusters uint, tree []huffmanTreeNode, buf []uint32) []uint32 {
	b.storeVarLenUint8(numClusters - 1)

	if numClusters <= 1 {
		return buf
	}

	// Work on a copy to avoid mutating the caller's context map.
	buf = growUint32(buf, len(contextMap))
	rleSymbols := buf[:len(contextMap)]
	copy(rleSymbols, contextMap)

	moveToFrontTransform(rleSymbols)
	rleSymbols, maxRunLenPrefix := runLengthCodeZeros(rleSymbols)

	const symbolMask = (1 << symbolBits) - 1

	// Build histogram of the RLE symbols.
	alphabetSize := numClusters + uint(maxRunLenPrefix)
	var histogram [maxNumberOfBlockTypes + 6 + 1]uint32
	for _, sym := range rleSymbols {
		histogram[sym&symbolMask]++
	}

	// Write RLE parameters.
	useRLE := maxRunLenPrefix > 0
	if useRLE {
		b.writeBits(1, 1)
		b.writeBits(4, uint64(maxRunLenPrefix-1))
	} else {
		b.writeBits(1, 0)
	}

	// Build and write the Huffman tree for the context map alphabet.
	var depths [maxNumberOfBlockTypes + 6 + 1]byte
	var codes [maxNumberOfBlockTypes + 6 + 1]uint16
	b.buildAndWriteHuffmanTree(
		histogram[:alphabetSize], alphabetSize,
		tree, depths[:alphabetSize], codes[:alphabetSize],
	)

	// Write encoded symbols with extra bits for zero-run prefixes.
	for _, sym := range rleSymbols {
		rleSym := sym & symbolMask
		extraBitsVal := sym >> symbolBits
		b.writeBits(uint(depths[rleSym]), uint64(codes[rleSym]))
		if rleSym > 0 && rleSym <= maxRunLenPrefix {
			b.writeBits(uint(rleSym), uint64(extraBitsVal))
		}
	}

	// Signal the decoder to apply inverse move-to-front.
	b.writeBits(1, 1)
	return buf
}

// byteAlign advances the bit position to the next byte boundary.
// If already byte-aligned, the position is unchanged.
func (b *bitWriter) byteAlign() {
	b.bitOffset = (b.bitOffset + 7) &^ 7
}

// writeBytes copies data into the bitstream buffer and advances the bit
// position. The bitstream must be byte-aligned before calling.
func (b *bitWriter) writeBytes(data []byte) {
	off := b.bitOffset / 8
	copy(b.buf[off:], data)
	b.bitOffset += uint(len(data)) * 8
	b.buf[b.bitOffset/8] = 0
}

// rewindTo resets the bitstream to the given bit position and clears
// any trailing bits in the current byte.
func (b *bitWriter) rewindTo(bitOffset uint) {
	bitpos := bitOffset & 7
	mask := byte((1 << bitpos) - 1)
	b.buf[bitOffset/8] &= mask
	b.bitOffset = bitOffset
}

// sortHuffmanTree sorts tree nodes by totalCount using the same algorithm as the
// C reference encoder: insertion sort for n < 13, shell sort otherwise. Using
// Go's standard library sort would produce different orderings for equal-count
// nodes, changing the Huffman tree shape and thus the compressed output.
func sortHuffmanTree(items []huffmanTreeNode) {
	n := len(items)
	if n < 13 {
		// Insertion sort.
		for i := 1; i < n; i++ {
			tmp := items[i]
			k := i
			j := i - 1
			for tmp.totalCount < items[j].totalCount {
				items[k] = items[j]
				k = j
				if j == 0 {
					break
				}
				j--
			}
			items[k] = tmp
		}
		return
	}

	// Shell sort with the same gap sequence as the C reference.
	var shellGaps = [6]int{132, 57, 23, 10, 4, 1}
	g := 2
	if n >= 57 {
		g = 0
	}
	for ; g < 6; g++ {
		gap := shellGaps[g]
		for i := gap; i < n; i++ {
			tmp := items[i]
			j := i
			for j >= gap && tmp.totalCount < items[j-gap].totalCount {
				items[j] = items[j-gap]
				j -= gap
			}
			items[j] = tmp
		}
	}
}

// moveToFrontTransform applies the move-to-front transform to v in place.
// Each value is replaced by its position in a dynamically maintained list,
// reducing the effective alphabet for context map values that exhibit locality.
func moveToFrontTransform(v []uint32) {
	if len(v) == 0 {
		return
	}

	// Find the maximum value to size the MTF list.
	maxVal := v[0]
	for _, x := range v[1:] {
		if x > maxVal {
			maxVal = x
		}
	}

	// Initialize MTF list [0..maxVal].
	var mtf [256]byte
	for i := range maxVal + 1 {
		mtf[i] = byte(i)
	}

	for i, x := range v {
		// Find the position of x in the MTF list.
		val := byte(x)
		var pos uint32
		for mtf[pos] != val {
			pos++
		}
		v[i] = pos

		// Move the found element to the front.
		copy(mtf[1:pos+1], mtf[:pos])
		mtf[0] = val
	}
}

// runLengthCodeZeros encodes runs of zeros in v using a prefix code scheme.
// Returns the modified slice (reusing v's storage), and the maximum run
// length prefix used. Non-zero values are shifted up by maxRunLenPrefix.
//
// The prefix code for a zero run of length L is floor(log2(L)), with
// floor(log2(L)) extra bits encoding L - 2^floor(log2(L)). The packed
// representation stores the prefix code in the lower symbolBits and the
// extra bits above that.
func runLengthCodeZeros(v []uint32) (out []uint32, maxRunLenPrefix uint32) {
	// First pass: find the longest zero run.
	var maxReps uint32
	for i := 0; i < len(v); {
		for i < len(v) && v[i] != 0 {
			i++
		}
		var reps uint32
		for i < len(v) && v[i] == 0 {
			reps++
			i++
		}
		if reps > maxReps {
			maxReps = reps
		}
	}

	// Cap the prefix at floor(log2(maxReps)), at most 6.
	if maxReps > 0 {
		maxRunLenPrefix = uint32(bits.Len32(maxReps)) - 1
	}
	if maxRunLenPrefix > 6 {
		maxRunLenPrefix = 6
	}

	// Second pass: encode in place. Output is always <= input length.
	n := 0
	for i := 0; i < len(v); {
		if v[i] != 0 {
			v[n] = v[i] + maxRunLenPrefix
			i++
			n++
		} else {
			// Count the zero run.
			var reps uint32
			for j := i; j < len(v) && v[j] == 0; j++ {
				reps++
			}
			i += int(reps)

			// Emit prefix codes for the zero run.
			for reps != 0 {
				if reps < 2<<maxRunLenPrefix {
					runLenPrefix := uint32(bits.Len32(reps)) - 1
					extra := reps - (1 << runLenPrefix)
					v[n] = runLenPrefix + (extra << symbolBits)
					n++
					break
				}
				extra := (uint32(1) << maxRunLenPrefix) - 1
				v[n] = maxRunLenPrefix + (extra << symbolBits)
				reps -= (2 << maxRunLenPrefix) - 1
				n++
			}
		}
	}

	return v[:n], maxRunLenPrefix
}
