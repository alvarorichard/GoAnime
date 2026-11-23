package core

// Alphabet sizes from the brotli specification.
// https://datatracker.ietf.org/doc/html/rfc7932#section-3.3
const (
	AlphabetSizeLiteral             = 256
	AlphabetSizeInsertAndCopyLength = 704
	AlphabetSizeBlockCount          = 26

	// NumDistanceShortCodes is the number of distance short codes that reference
	// the ring buffer of recent distances (RFC 7932 Section 4).
	NumDistanceShortCodes = 16

	// AlphabetSizeCodeLengths is the size of the code length alphabet
	// (RFC 7932 Section 3.5): 18 = 16 literal lengths + repeat-previous (16)
	// + repeat-zero (17).
	AlphabetSizeCodeLengths = 18

	// NumHistogramDistanceSymbols is the fixed histogram bin count used for
	// distance histograms in the block splitter and histogram clustering.
	// Matches BROTLI_NUM_HISTOGRAM_DISTANCE_SYMBOLS in the C reference.
	NumHistogramDistanceSymbols = 544

	// MaxDistanceBits is the maximum number of extra bits in a distance code
	// (RFC 7932 Section 4). The distance alphabet size formula uses this as
	// the MAXNBITS parameter.
	MaxDistanceBits = 24

	// WindowGap is the safety margin between the window size and the maximum
	// backward distance. Reserves space for ring-buffer read-ahead in the
	// streaming encoder.
	WindowGap = 16

	// RepeatPreviousCodeLength and InitialRepeatedCodeLength are entropy
	// coding constants from RFC 7932 Section 3.5.
	// https://datatracker.ietf.org/doc/html/rfc7932#section-3.5
	RepeatPreviousCodeLength  = 16 // used for repeating previous non-zero code length
	InitialRepeatedCodeLength = 8  // "code length of 8 is repeated"
)

// CodeLengthCodeOrder is the order in which code length codes are transmitted
// (RFC 7932 section 3.5).
var CodeLengthCodeOrder = [...]byte{
	1, 2, 3, 4, 0, 5, 17, 6, 16, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

// BlockLengthOffset and BlockLengthNBits are block length prefix code ranges
// (RFC 7932 Section 6). Each entry is {offset, nbits}; the block length for
// code c is:
//
//	offset[c] + ReadBits(nbits[c])
var BlockLengthOffset = [AlphabetSizeBlockCount]uint32{
	1, 5, 9, 13, 17, 25, 33, 41, 49, 65, 81, 97,
	113, 145, 177, 209, 241, 305, 369, 497, 753, 1265, 2289, 4337,
	8433, 16625,
}

// BlockLengthNBits is the extra-bit count for each block-length prefix code.
var BlockLengthNBits = [AlphabetSizeBlockCount]uint32{
	2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4,
	5, 5, 5, 5, 6, 6, 7, 8, 9, 10, 11, 12,
	13, 24,
}
