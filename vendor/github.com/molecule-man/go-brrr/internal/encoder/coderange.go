package encoder

// Mapping Huffman-coded symbols to values via extra bits (RFC 7932 sections 4, 5, 6.3).
//
// Brotli often needs to encode a potentially very large range of values, with
// the assumption that small values are more likely to appear than large values.
// Rather than assigning a unique Huffman symbol to every possible value (which
// would require an impractically large alphabet), Brotli groups consecutive
// values into ranges. Each range is assigned a single code (a Huffman symbol),
// and extra bits read from the bitstream pinpoint the exact value within that
// range. The difference between the upper and lower bound of a range is always
// a power of two.
//
// A code range has three properties:
//
//   - Code:   the Huffman symbol identifying this range
//   - Offset: the smallest value in the range (lower bound)
//   - NBits:  number of extra bits to read; the range covers 2^NBits values
//
// Decoding:
//
//	value = Offset + read_bits(NBits)
//
// Example — encoding values 0 through 23 with an 8-symbol alphabet:
//
//	Code  NBits  Offset  Range
//	  0     0       0    [0, 0]
//	  1     0       1    [1, 1]
//	  2     1       2    [2, 3]
//	  3     1       4    [4, 5]
//	  4     2       6    [6, 9]
//	  5     3      10    [10, 17]
//	  6     2      18    [18, 21]
//	  7     1      22    [22, 23]
//
// To encode value 15: it falls in Code 5 (Offset=10, NBits=3).
// Extra bits = 15 − 10 = 5 = 0b101. Write Huffman symbol 5, then bits 101.
//
// To decode: read Huffman symbol → 5, look up Offset=10, NBits=3.
// Read 3 extra bits → 5. Value = 10 + 5 = 15.
//
// Small values (0, 1) require zero extra bits — just the Huffman symbol.
// Larger ranges require more extra bits but share a single symbol, keeping
// the Huffman alphabet compact.
//
// Brotli uses this pattern for:
//   - Insert and copy lengths (Section 5)
//   - Block counts (Section 6.3)
//   - Distance codes (Section 4), where the ranges are parameterized per
//     meta-block by NPOSTFIX and NDIRECT
