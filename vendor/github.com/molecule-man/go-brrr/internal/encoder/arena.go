// Package encoder hosts the brotli streaming/fast encoder implementations
// and shared codec primitives consumed by the root brrr package's decoder.
package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// Pre-allocated scratch buffers for one-pass fast compression, avoiding
// per-call heap allocations.

const defaultCommandCodeNumBits = 448

// Default command prefix codes, used to initialize the arena before the
// first compression call.
var defaultCommandDepths = [128]byte{
	0, 4, 4, 5, 6, 6, 7, 7, 7, 7, 7, 8, 8, 8, 8, 8,
	0, 0, 0, 4, 4, 4, 4, 4, 5, 5, 6, 6, 6, 6, 7, 7,
	7, 7, 10, 10, 10, 10, 10, 10, 0, 4, 4, 5, 5, 5, 6, 6,
	7, 8, 8, 9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
	5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	6, 6, 6, 6, 6, 6, 5, 5, 5, 5, 5, 5, 4, 4, 4, 4,
	4, 4, 4, 5, 5, 5, 5, 5, 5, 6, 6, 7, 7, 7, 8, 10,
	12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12,
}

var defaultCommandBits = [128]uint16{
	0, 0, 8, 9, 3, 35, 7, 71,
	39, 103, 23, 47, 175, 111, 239, 31,
	0, 0, 0, 4, 12, 2, 10, 6,
	13, 29, 11, 43, 27, 59, 87, 55,
	15, 79, 319, 831, 191, 703, 447, 959,
	0, 14, 1, 25, 5, 21, 19, 51,
	119, 159, 95, 223, 479, 991, 63, 575,
	127, 639, 383, 895, 255, 767, 511, 1023,
	14, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	27, 59, 7, 39, 23, 55, 30, 1, 17, 9, 25, 5, 0, 8, 4, 12,
	2, 10, 6, 21, 13, 29, 3, 19, 11, 15, 47, 31, 95, 63, 127, 255,
	767, 2815, 1791, 3839, 511, 2559, 1535, 3583, 1023, 3071, 2047, 4095,
}

var defaultCommandCode = [57]byte{
	0xff, 0x77, 0xd5, 0xbf, 0xe7, 0xde, 0xea, 0x9e, 0x51, 0x5d, 0xde, 0xc6,
	0x70, 0x57, 0xbc, 0x58, 0x58, 0x58, 0xd8, 0xd8, 0x58, 0xd5, 0xcb, 0x8c,
	0xea, 0xe0, 0xc3, 0x87, 0x1f, 0x83, 0xc1, 0x60, 0x1c, 0x67, 0xb2, 0xaa,
	0x06, 0x83, 0xc1, 0x60, 0x30, 0x18, 0xcc, 0xa1, 0xce, 0x88, 0x54, 0x94,
	0x46, 0xe1, 0xb0, 0xd0, 0x4e, 0xb2, 0xf7, 0x04, 0x00,
}

// onePassArena holds pre-allocated scratch buffers for compressFragmentFast,
// avoiding per-call heap allocations. Callers should reuse an arena across
// consecutive calls.
type onePassArena struct {
	litDepth [256]byte
	litBits  [256]uint16

	// Command and distance prefix codes (each 64 symbols, stored back-to-back)
	// used for the next block. The command prefix code is over a smaller alphabet
	// with the following 64 symbols:
	//    0 - 15: insert length code 0, copy length code 0 - 15, same distance
	//   16 - 39: insert length code 0, copy length code 0 - 23
	//   40 - 63: insert length code 0 - 23, copy length code 0
	// Note that symbols 16 and 40 represent the same code in the full alphabet,
	// but neither is used.
	cmdDepth [128]byte
	cmdBits  [128]uint16
	cmdHisto [128]uint32

	// The compressed form of the command and distance prefix codes for the
	// next block.
	cmdCode        [512]byte
	cmdCodeNumBits uint

	tree      [2*core.AlphabetSizeLiteral + 1]huffmanTreeNode
	histogram [256]uint32
	tmpDepth  [core.AlphabetSizeInsertAndCopyLength]byte
	tmpBits   [64]uint16
}

// twoPassArena holds pre-allocated scratch buffers for the two-pass fast
// compression, avoiding per-call heap allocations. Callers should reuse
// an arena across consecutive calls.
type twoPassArena struct {
	litHisto [256]uint32
	litDepth [256]byte
	litBits  [256]uint16

	cmdHisto [128]uint32
	cmdDepth [128]byte
	cmdBits  [128]uint16

	tree     [2*core.AlphabetSizeLiteral + 1]huffmanTreeNode
	tmpDepth [core.AlphabetSizeInsertAndCopyLength]byte
	tmpBits  [64]uint16
}

// metablockArena holds pre-allocated scratch buffers for encoding a metablock
// with writeMetaBlockFast or writeMetaBlockTrivial, avoiding per-call heap
// allocations. Callers should reuse an arena across consecutive calls.
//
// Call resetHistograms before each metablock to zero the histogram arrays.
type metablockArena struct {
	litHisto  [core.AlphabetSizeLiteral]uint32
	cmdHisto  [core.AlphabetSizeInsertAndCopyLength]uint32
	distHisto [alphabetSizeDistance]uint32

	litDepth  [core.AlphabetSizeLiteral]byte
	litBits   [core.AlphabetSizeLiteral]uint16
	cmdDepth  [core.AlphabetSizeInsertAndCopyLength]byte
	cmdBits   [core.AlphabetSizeInsertAndCopyLength]uint16
	distDepth [alphabetSizeDistance]byte
	distBits  [alphabetSizeDistance]uint16

	tree [2*core.AlphabetSizeInsertAndCopyLength + 1]huffmanTreeNode
}

// initCommandPrefixCodes initializes the command and distance prefix codes
// for the first block. This must be called before the first call to
// compressFragmentFast.
func (s *onePassArena) initCommandPrefixCodes() {
	s.cmdDepth = defaultCommandDepths
	s.cmdBits = defaultCommandBits
	copy(s.cmdCode[:], defaultCommandCode[:])
	s.cmdCodeNumBits = defaultCommandCodeNumBits
}

// buildAndWriteLiteralPrefixCode builds a literal prefix code into the arena's
// litDepth/litBits based on the statistics of the input and writes it to the
// bitstream. Returns the estimated compression ratio in millibytes/char.
func (s *onePassArena) buildAndWriteLiteralPrefixCode(
	input []byte, b *bitWriter,
) uint {
	histogram := &s.histogram
	*histogram = [256]uint32{}
	var histogramTotal uint

	if len(input) < (1 << 15) {
		for _, c := range input {
			histogram[c]++
		}
		histogramTotal = uint(len(input))
		for i := range 256 {
			// We weigh the first 11 samples with weight 3 to account for the
			// balancing effect of the LZ77 phase on the histogram.
			adjust := 2 * min(histogram[i], 11)
			histogram[i] += adjust
			histogramTotal += uint(adjust)
		}
	} else {
		const sampleRate = 29
		for i := 0; i < len(input); i += sampleRate {
			histogram[input[i]]++
		}
		histogramTotal = uint((len(input) + sampleRate - 1) / sampleRate)
		for i := range 256 {
			// We add 1 to each population count to avoid 0 bit depths (since
			// this is only a sample and we don't know if the symbol appears or
			// not), and we weigh the first 11 samples with weight 3 to account
			// for the balancing effect of the LZ77 phase on the histogram.
			adjust := 1 + 2*min(histogram[i], 11)
			histogram[i] += adjust
			histogramTotal += uint(adjust)
		}
	}

	b.buildAndWriteHuffmanTreeFast(s.tree[:], histogram[:], histogramTotal,
		8, s.litDepth[:], s.litBits[:])

	var literalRatio uint
	for i := range 256 {
		if histogram[i] != 0 {
			literalRatio += uint(histogram[i]) * uint(s.litDepth[i])
		}
	}
	// Estimated encoding ratio, millibytes per symbol.
	return (literalRatio * 125) / histogramTotal
}

// buildAndWriteCommandPrefixCode builds a command and distance prefix code
// (each 64 symbols) into depth and bits based on the histogram, and stores
// it to the bitstream.
func (s *onePassArena) buildAndWriteCommandPrefixCode(b *bitWriter) {
	depth := &s.cmdDepth
	cmdBits := &s.cmdBits
	tmpDepth := &s.tmpDepth
	tmpBits := &s.tmpBits

	clear(tmpDepth[:core.AlphabetSizeInsertAndCopyLength])

	createHuffmanTree(s.cmdHisto[:64], 15, s.tree[:], depth[:64])
	createHuffmanTree(s.cmdHisto[64:128], 14, s.tree[:], depth[64:128])

	// We have to jump through a few hoops here in order to compute
	// the command bits because the symbols are in a different order than in
	// the full alphabet. This looks complicated, but having the symbols
	// in this order in the command bits saves a few branches in the write
	// functions.
	copy(tmpDepth[:24], depth[:24])
	copy(tmpDepth[24:32], depth[40:48])
	copy(tmpDepth[32:40], depth[24:32])
	copy(tmpDepth[40:48], depth[48:56])
	copy(tmpDepth[48:56], depth[32:40])
	copy(tmpDepth[56:64], depth[56:64])
	convertBitDepthsToSymbols(tmpDepth[:64], tmpBits[:64])
	copy(cmdBits[:24], tmpBits[:24])
	copy(cmdBits[24:32], tmpBits[32:40])
	copy(cmdBits[32:40], tmpBits[48:56])
	copy(cmdBits[40:48], tmpBits[24:32])
	copy(cmdBits[48:56], tmpBits[40:48])
	copy(cmdBits[56:64], tmpBits[56:64])
	convertBitDepthsToSymbols(depth[64:128], cmdBits[64:128])

	// Create the bit length array for the full command alphabet.
	clear(tmpDepth[:64])
	copy(tmpDepth[:8], depth[:8])
	copy(tmpDepth[64:72], depth[8:16])
	copy(tmpDepth[128:136], depth[16:24])
	copy(tmpDepth[192:200], depth[24:32])
	copy(tmpDepth[384:392], depth[32:40])
	for i := range 8 {
		tmpDepth[128+8*i] = depth[40+i]
		tmpDepth[256+8*i] = depth[48+i]
		tmpDepth[448+8*i] = depth[56+i]
	}
	b.writeHuffmanTree(tmpDepth[:core.AlphabetSizeInsertAndCopyLength], s.tree[:])
	b.writeHuffmanTree(depth[64:128], s.tree[:])
}

// shouldMergeBlock decides whether to extend the current meta-block with a
// new block of data instead of starting a new meta-block.
func (s *onePassArena) shouldMergeBlock(data []byte, length int, depths []byte) bool {
	const sampleRate = 43
	histogram := &s.histogram
	*histogram = [256]uint32{}
	for i := 0; i < length; i += sampleRate {
		histogram[data[i]]++
	}

	total := (length + sampleRate - 1) / sampleRate
	r := float64((fastLog2(total)+0.5)*float64(total)) + 200
	for i := range 256 {
		r -= float64(float64(histogram[i]) * (float64(depths[i]) + fastLog2(int(histogram[i]))))
	}
	return r >= 0.0
}

// resetHistograms zeroes the literal histogram and prefix-code arrays before
// encoding a new block. cmdHisto is built incrementally during createCommands
// and is zeroed there.
func (s *twoPassArena) resetHistograms() {
	s.litHisto = [256]uint32{}
	s.cmdDepth = [128]byte{}
	s.cmdBits = [128]uint16{}
}

// buildAndWriteCommandPrefixCode builds a command and distance prefix code
// (each 64 symbols) into cmdDepth/cmdBits based on cmdHisto, and writes it
// into the bitstream.
func (s *twoPassArena) buildAndWriteCommandPrefixCode(b *bitWriter) {
	// Tree size for building a tree over 64 symbols is 2 * 64 + 1.
	clear(s.tmpDepth[:])

	createHuffmanTree(s.cmdHisto[:64], 15, s.tree[:], s.cmdDepth[:64])
	createHuffmanTree(s.cmdHisto[64:128], 14, s.tree[:], s.cmdDepth[64:128])

	// We have to jump through a few hoops here in order to compute
	// the command bits because the symbols are in a different order than in
	// the full alphabet. This looks complicated, but having the symbols
	// in this order in the command bits saves a few branches in the write
	// functions.
	copy(s.tmpDepth[:24], s.cmdDepth[24:48])
	copy(s.tmpDepth[24:32], s.cmdDepth[:8])
	copy(s.tmpDepth[32:40], s.cmdDepth[48:56])
	copy(s.tmpDepth[40:48], s.cmdDepth[8:16])
	copy(s.tmpDepth[48:56], s.cmdDepth[56:64])
	copy(s.tmpDepth[56:64], s.cmdDepth[16:24])
	convertBitDepthsToSymbols(s.tmpDepth[:64], s.tmpBits[:64])
	copy(s.cmdBits[:8], s.tmpBits[24:32])
	copy(s.cmdBits[8:16], s.tmpBits[40:48])
	copy(s.cmdBits[16:24], s.tmpBits[56:64])
	copy(s.cmdBits[24:48], s.tmpBits[:24])
	copy(s.cmdBits[48:56], s.tmpBits[32:40])
	copy(s.cmdBits[56:64], s.tmpBits[48:56])
	convertBitDepthsToSymbols(s.cmdDepth[64:128], s.cmdBits[64:128])

	// Create the bit length array for the full command alphabet.
	clear(s.tmpDepth[:64]) // only first 64 values were used
	copy(s.tmpDepth[:8], s.cmdDepth[24:32])
	copy(s.tmpDepth[64:72], s.cmdDepth[32:40])
	copy(s.tmpDepth[128:136], s.cmdDepth[40:48])
	copy(s.tmpDepth[192:200], s.cmdDepth[48:56])
	copy(s.tmpDepth[384:392], s.cmdDepth[56:64])
	for i := range 8 {
		s.tmpDepth[128+8*i] = s.cmdDepth[i]
		s.tmpDepth[256+8*i] = s.cmdDepth[8+i]
		s.tmpDepth[448+8*i] = s.cmdDepth[16+i]
	}
	b.writeHuffmanTree(s.tmpDepth[:core.AlphabetSizeInsertAndCopyLength], s.tree[:])
	b.writeHuffmanTree(s.cmdDepth[64:128], s.tree[:])
}

// resetHistograms zeroes the histogram arrays before encoding a new metablock.
func (a *metablockArena) resetHistograms() {
	a.litHisto = [core.AlphabetSizeLiteral]uint32{}
	a.cmdHisto = [core.AlphabetSizeInsertAndCopyLength]uint32{}
	a.distHisto = [alphabetSizeDistance]uint32{}
}
