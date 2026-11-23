package encoder

import "github.com/molecule-man/go-brrr/internal/core"

const (
	// Code length alphabet symbol used for the repeat-zero RLE code
	// (RFC 7932 Section 3.5).
	alphabetSizeRepeatZeroCodeLength = 17

	// Maximum distance alphabet size across all valid parameter combinations.
	// Covers NPOSTFIX=0..3, NDIRECT=0..120 (RFC 7932 Section 4).
	alphabetSizeDistance = 140

	maxHuffmanBits = 16 // 0..15 are values for bits

	// Maximum number of block types (RFC 7932 Section 6).
	maxNumberOfBlockTypes = 256
	maxBlockTypeSymbols   = maxNumberOfBlockTypes + 2 // 258

	// Absolute max backward distance expressible in the brotli bitstream
	// (NDIRECT=0, NPOSTFIX=0, MAXNBITS=24). RFC 7932 Section 4.
	maxBackwardDistance = 0x3FFFFFC
)

// Static Huffman code for encoding code length bit depths:
//
//	Symbol   Code
//	  0        00
//	  1      1110
//	  2       110
//	  3        01
//	  4        10
//	  5      1111
var codeLengthCodeSymbols = [...]byte{0, 7, 3, 2, 1, 15}
var codeLengthCodeBitLengths = [...]byte{2, 4, 3, 2, 2, 4}

// Insert/copy length code extra-bit lookup tables (RFC 7932 Section 5).
// Indexed by insert length code (0–23) or copy length code (0–23).
var insertExtra = [24]uint32{
	0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3,
	4, 4, 5, 5, 6, 7, 8, 9, 10, 12, 14, 24,
}

var copyExtra = [24]float32{
	0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 2,
	3, 3, 4, 4, 5, 5, 6, 7, 8, 9, 10, 24,
}

// getBlockLengthPrefixCode returns the prefix code, number of extra bits,
// and extra bits value for a block length.
func getBlockLengthPrefixCode(length uint32) (code, nExtra, extra uint32) {
	// Fast jump to approximate code, then linear scan.
	switch {
	case length >= 753:
		code = 20
	case length >= 177:
		code = 14
	case length >= 41:
		code = 7
	default:
		code = 0
	}
	for code < core.AlphabetSizeBlockCount-1 && length >= core.BlockLengthOffset[code+1] {
		code++
	}
	nExtra = core.BlockLengthNBits[code]
	extra = length - core.BlockLengthOffset[code]
	return code, nExtra, extra
}
