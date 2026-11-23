package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// combineLengthCodesBase holds the base offset for each (insCode>>3, copyCode>>3)
// cell of the RFC 7932 Section 5 insert-and-copy length code grid. Both codes
// are in [0,23], so the index into each dimension is in [0,2].
//
//	               Copy code
//	               0..7  8..15  16..23
//	Insert  0..7   128   192    384
//	        8..15  256   320    512
//	        16..23 448   576    640
var combineLengthCodesBase = [3][3]uint16{
	{128, 192, 384},
	{256, 320, 512},
	{448, 576, 640},
}

var insertLenCodeLUT = initInsertLenCodeLUT()
var copyLenCodeLUT = initCopyLenCodeLUT()
var cmdCodeLUT = initCmdCodeLUT()

// commandConfig holds the inputs for constructing a command from a LZ77 match.
type commandConfig struct {
	insertLen uint
	copyLen   uint

	// copyLenDelta is the difference between the copy length code and
	// the actual copy length (used when distance short-codes remap lengths).
	copyLenDelta int

	// distanceCode is an intermediate code: one of the 16 short codes (ring
	// buffer references) or the actual distance increased by
	// core.NumDistanceShortCodes - 1.
	distanceCode uint

	// numDirectCodes and postfixBits are the distance encoding parameters for
	// the meta-block (RFC 7932 Section 4).
	numDirectCodes uint
	postfixBits    uint
}

// command represents a single brotli command in the compressed stream.
// Each command consists of two parts: a sequence of literal bytes (of strings
// that have not been detected as duplicated within the sliding window) and a
// pointer to a duplicated string, which is represented as a pair
//
//	<length, backward distance>
//
// There may be zero literal bytes in the command.
//
// One meta-block command then appears as a sequence of prefix codes:
//
//	Insert-and-copy length, literal, literal, ..., literal, distance
//
// where the insert-and-copy length defines an insertion length and a copy
// length. The insertion length determines the number of literals that
// immediately follow. The distance defines how far back to go for the copy and
// the copy length determines the number of bytes to copy.
//
// The resulting uncompressed data is the sequence of bytes:
//
//		literal, literal, ..., literal, copy, copy, ..., copy
//	     |______ insertLen ________|  |____ copyLen ___|
//
// The last command in the meta-block may end with the last literal if the total
// uncompressed length of the meta-block has been satisfied. In that case, there
// is no distance in the last command, and the copy length is ignored.
//
// Literal bytes are not present in the struct (they are stored in separate ring
// buffer) to make the struct allocate-able on stack.
// The struct is 16 bytes so that 4 commands fit per 64-byte cache line.
type command struct {
	insertLen  uint32 // number of literal bytes
	copyLen    uint32 // copy length (low 25 bits) + copy code delta (high 7 bits)
	distExtra  uint32 // distance extra bits
	cmdPrefix  uint16 // the combined insert-and-copy length code (Section 5)
	distPrefix uint16 // distance code (low 10 bits) + extra bits length (high 6 bits)
}

func initInsertLenCodeLUT() [130]uint16 {
	var lut [130]uint16
	for i := range lut {
		lut[i] = getInsertLenCodeSlow(uint(i))
	}
	return lut
}

func initCopyLenCodeLUT() [134]uint16 {
	var lut [134]uint16
	for i := uint(2); i < uint(len(lut)); i++ {
		lut[i] = getCopyLenCodeSlow(i)
	}
	return lut
}

func initCmdCodeLUT() [2][24][32]uint16 {
	var lut [2][24][32]uint16
	for insCode := range uint16(24) {
		for copyCode := range uint16(24) {
			lut[0][insCode][copyCode] = combineLengthCodes(insCode, copyCode, false)
			lut[1][insCode][copyCode] = combineLengthCodes(insCode, copyCode, true)
		}
	}
	return lut
}

// newCommand creates a command for a literal insertion followed by a backward
// reference copy.
func newCommand(cfg commandConfig) command {
	delta := uint32(uint8(int8(cfg.copyLenDelta)))
	distPrefix, distExtra := prefixEncodeCopyDistance(cfg.distanceCode, cfg.numDirectCodes, cfg.postfixBits)
	effectiveCopyLen := uint(int(cfg.copyLen) + cfg.copyLenDelta)
	insCode := getInsertLenCode(cfg.insertLen)
	copyCode := getCopyLenCode(effectiveCopyLen)
	cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
	return command{
		insertLen:  uint32(cfg.insertLen),
		copyLen:    uint32(cfg.copyLen) | (delta << 25),
		distExtra:  distExtra,
		cmdPrefix:  cmdPrefix,
		distPrefix: distPrefix,
	}
}

func (s *encodeState) pushCommandSimpleDist(insertLen, copyLen uint, copyLenDelta int, distanceCode uint) {
	delta := uint32(uint8(int8(copyLenDelta)))
	distPrefix, distExtra := prefixEncodeSimpleDistance(distanceCode)
	effectiveCopyLen := uint(int(copyLen) + copyLenDelta)
	insCode := getInsertLenCode(insertLen)
	copyCode := getCopyLenCode(effectiveCopyLen)
	cmdPrefix := combineLengthCodes(insCode, copyCode, (distPrefix&0x3FF) == 0)
	s.appendCommand(uint32(insertLen), uint32(copyLen)|(delta<<25), distExtra, cmdPrefix, distPrefix)
}

func (s *encodeState) appendCommand(insertLen, copyLen, distExtra uint32, cmdPrefix, distPrefix uint16) {
	s.commands = append(s.commands, command{
		insertLen:  insertLen,
		copyLen:    copyLen,
		distExtra:  distExtra,
		cmdPrefix:  cmdPrefix,
		distPrefix: distPrefix,
	})
}

// newInsertCommand creates a command that contains only literal insertions
// with no backward reference.
func newInsertCommand(insertLen uint) command {
	insCode := getInsertLenCode(insertLen)
	copyCode := getCopyLenCode(4)
	cmdPrefix := combineLengthCodes(insCode, copyCode, false)
	return command{
		insertLen:  uint32(insertLen),
		copyLen:    4 << 25,
		distPrefix: core.NumDistanceShortCodes,
		cmdPrefix:  cmdPrefix,
	}
}

func (c command) usesLastDistance() bool {
	return c.cmdPrefix < 128
}

// distanceContext returns a 2-bit distance context (0–3) for this command,
// used to select among four distance Huffman codes per block type.
//
// The context is derived from the insert-and-copy length code grid
// (RFC 7932 Section 5). Context 0–2 indicates that a short distance
// code is likely (copy length is small relative to the grid cell);
// context 3 indicates a general distance code is expected.
func (c command) distanceContext() uint32 {
	r := c.cmdPrefix >> 6
	col := c.cmdPrefix & 7
	if (r == 0 || r == 2 || r == 4 || r == 7) && col <= 2 {
		return uint32(col)
	}
	return 3
}

func (c command) copyLength() uint32 {
	return c.copyLen & 0x1FFFFFF
}

func (c command) distPrefixCode() uint16 {
	return c.distPrefix & 0x3FF
}

// copyLenCode returns the effective copy length code, applying the
// signed delta stored in the high 7 bits of copyLen.
func (c command) copyLenCode() uint32 {
	modifier := c.copyLen >> 25
	delta := int8(uint8(modifier | ((modifier & 0x40) << 1)))
	return uint32(int32(c.copyLen&0x1FFFFFF) + int32(delta))
}

// distanceCode reconstructs the distance code from the packed distPrefix
// and distExtra fields. Inverse of prefixEncodeCopyDistance.
func (c command) distanceCode(numDirectCodes, postfixBits uint) uint32 {
	dcode := uint32(c.distPrefix & 0x3FF)
	if dcode < core.NumDistanceShortCodes+uint32(numDirectCodes) {
		return dcode
	}
	nbits := uint32(c.distPrefix >> 10)
	extra := c.distExtra
	postfixMask := (uint32(1) << postfixBits) - 1
	hcode := (dcode - uint32(numDirectCodes) - core.NumDistanceShortCodes) >> postfixBits
	lcode := (dcode - uint32(numDirectCodes) - core.NumDistanceShortCodes) & postfixMask
	offset := ((2 + (hcode & 1)) << nbits) - 4
	return ((offset+extra)<<postfixBits + lcode +
		uint32(numDirectCodes) + core.NumDistanceShortCodes)
}

// getInsertLenCode returns the insert length prefix code (0–23) for a given
// insertion length, per the table in RFC 7932 Section 5:
//
//	Code — symbol written to the compressed stream via a Huffman code.
//	Extra — number of raw bits written after the code to pinpoint the exact length.
//	Range — the set of lengths that code + extra bits can represent.
//
//	Code  Extra  Range         Code  Extra  Range
//	0–5   0      0–5           14    5      66–97
//	6     1      6–7           15    5      98–129
//	7     1      8–9           16    6      130–193
//	8     2      10–13         17    7      194–321
//	9     2      14–17         18    8      322–577
//	10    3      18–25         19    9      578–1089
//	11    3      26–33         20    10     1090–2113
//	12    4      34–49         21    12     2114–6209
//	13    4      50–65         22    14     6210–22593
//	                           23    24     22594–16799809
func getInsertLenCode(insertLen uint) uint16 {
	if insertLen < uint(len(insertLenCodeLUT)) {
		return insertLenCodeLUT[insertLen]
	}
	return getInsertLenCodeSlow(insertLen)
}

func getInsertLenCodeSlow(insertLen uint) uint16 {
	switch {
	case insertLen < 6: // codes 0–5: one-to-one mapping
		return uint16(insertLen)
	case insertLen < 130: // codes 6–15: 1–5 extra bits
		nbits := uint(bits.Len(insertLen-2) - 2)
		return uint16((nbits << 1) + ((insertLen - 2) >> nbits) + 2)
	case insertLen < 2114: // codes 16–20: 6–10 extra bits
		return uint16(bits.Len(insertLen-66) + 9)
	case insertLen < 6210: // code 21: 12 extra bits
		return 21
	case insertLen < 22594: // code 22: 14 extra bits
		return 22
	default: // code 23: 24 extra bits
		return 23
	}
}

// getCopyLenCode returns the copy length prefix code (0–23) for a given
// copy length, per the table in RFC 7932 Section 5:
//
//	Code — symbol written to the compressed stream via a Huffman code.
//	Extra — number of raw bits written after the code to pinpoint the exact length.
//	Range — the set of lengths that code + extra bits can represent.
//
//	Code  Extra  Range         Code  Extra  Range
//	0–7   0      2–9           14    4      38–53
//	8     1      10–11         15    4      54–69
//	9     1      12–13         16    5      70–101
//	10    2      14–17         17    5      102–133
//	11    2      18–21         18    6      134–197
//	12    3      22–29         19    7      198–325
//	13    3      30–37         20    8      326–581
//	                           21    9      582–1093
//	                           22    10     1094–2117
//	                           23    24     2118–16779333
func getCopyLenCode(copyLen uint) uint16 {
	if copyLen < uint(len(copyLenCodeLUT)) && copyLen >= 2 {
		return copyLenCodeLUT[copyLen]
	}
	return getCopyLenCodeSlow(copyLen)
}

func getCopyLenCodeSlow(copyLen uint) uint16 {
	switch {
	case copyLen < 10: // codes 0–7: one-to-one mapping (offset by 2)
		return uint16(copyLen - 2)
	case copyLen < 134: // codes 8–17: 1–5 extra bits
		nbits := uint(bits.Len(copyLen-6) - 2)
		return uint16((nbits << 1) + ((copyLen - 6) >> nbits) + 4)
	case copyLen < 2118: // codes 18–22: 6–10 extra bits
		return uint16(bits.Len(copyLen-70) + 11)
	default: // code 23: 24 extra bits
		return 23
	}
}

// prefixEncodeSimpleDistance is a specialization of prefixEncodeCopyDistance
// for numDirectCodes=0 and postfixBits=0 (the default distance parameters).
// Small enough to be inlined at hot call sites.
func prefixEncodeSimpleDistance(distanceCode uint) (code uint16, extraBits uint32) {
	if distanceCode < core.NumDistanceShortCodes {
		return uint16(distanceCode), 0
	}

	dist := 4 + distanceCode - core.NumDistanceShortCodes
	bucket := uint(bits.Len(dist)) - 2
	prefix := (dist >> bucket) & 1
	offset := (2 + prefix) << bucket

	code = uint16((bucket << 10) |
		(core.NumDistanceShortCodes + 2*(bucket-1) + prefix))
	extraBits = uint32(dist - offset)
	return code, extraBits
}

// prefixEncodeCopyDistance encodes a distance into a prefix code and extra bits
// for the distance symbol alphabet (RFC 7932 Section 4).
//
// distanceCode is an intermediate code: one of the 16 short codes (ring buffer
// references) or the actual distance increased by numDirectCodes - 1.
//
// The returned code packs the distance symbol in the low 10 bits and the number
// of extra bits in the high 6 bits. This matches the distPrefix field layout.
func prefixEncodeCopyDistance(distanceCode, numDirectCodes, postfixBits uint) (code uint16, extraBits uint32) {
	if distanceCode < core.NumDistanceShortCodes+numDirectCodes {
		return uint16(distanceCode), 0
	}

	dist := (uint(1) << (postfixBits + 2)) +
		(distanceCode - core.NumDistanceShortCodes - numDirectCodes)
	bucket := uint(bits.Len(dist)) - 2 // Log2FloorNonZero(dist) - 1
	postfixMask := (uint(1) << postfixBits) - 1
	postfix := dist & postfixMask
	prefix := (dist >> bucket) & 1
	offset := (2 + prefix) << bucket
	nbits := bucket - postfixBits

	code = uint16((nbits << 10) |
		(core.NumDistanceShortCodes + numDirectCodes +
			((2*(nbits-1) + prefix) << postfixBits) + postfix))
	extraBits = uint32((dist - offset) >> postfixBits)
	return code, extraBits
}

// combineLengthCodes produces the insert-and-copy length code (Section 5) from
// separate insert and copy length codes.
//
// Each cell in the RFC 7932 Section 5 grid holds 64 combined codes. The cell's
// start value determines the base, and bits 0–2 / 3–5 of the combined code
// select the exact copy / insert code within the cell's 8-wide ranges.
//
//	When useLastDistance is true and both codes are small (ins < 8, copy < 16),
//	the first two rows are used (codes 0–127), signaling distance-symbol reuse:
//
//	                  Copy code
//	                  0..7       8..15
//	  Insert  0..7    0..63      64..127
//
//	Otherwise the 3×3 grid is used (see combineLengthCodesBase).
func combineLengthCodes(insCode, copyCode uint16, useLastDistance bool) uint16 {
	bits64 := (copyCode & 0x7) | ((insCode & 0x7) << 3)
	if useLastDistance && insCode < 8 && copyCode < 16 {
		if copyCode < 8 {
			return bits64
		}
		return bits64 | 64
	}
	return combineLengthCodesBase[insCode>>3][copyCode>>3] | bits64
}
