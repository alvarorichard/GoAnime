// Block encoder for quality >= 4 metablock encoding.
//
// Manages per-block-type Huffman codes and block switch signaling for the
// three symbol categories (literal, command, distance).

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// blockTypeCodeCalc tracks the last two block types to compute the
// type code for block switches.
type blockTypeCodeCalc struct {
	lastType       uint
	secondLastType uint
}

// blockSplitCode stores the Huffman codes needed to encode block type
// and block length switches within a block category.
type blockSplitCode struct {
	typeCalc     blockTypeCodeCalc
	typeDepths   [maxBlockTypeSymbols]byte
	typeBits     [maxBlockTypeSymbols]uint16
	lengthDepths [core.AlphabetSizeBlockCount]byte
	lengthBits   [core.AlphabetSizeBlockCount]uint16
}

// blockEncoder manages the encoding of one block category (literal, command,
// or distance) within a metablock. It tracks block boundaries, emits block
// switch codes, and stores symbols with per-block-type Huffman codes.
type blockEncoder struct {
	blockTypes   []byte   // reference to blockSplit.types
	blockLengths []uint32 // reference to blockSplit.lengths
	depths       []byte   // Huffman depths, all block types concatenated
	bits         []uint16 // Huffman codes, all block types concatenated

	splitCode     blockSplitCode // Huffman codes for type/length switches
	histogramLen  int            // alphabet size (256, 704, or distAlphabetSize)
	numBlockTypes int            // from blockSplit.numTypes
	numBlocks     int            // len(blockTypes)
	blockIdx      int            // current block index
	blockLen      int            // remaining symbols in current block
	entropyIdx    int            // current histogram offset = blockType * histogramLen
}

// newBlockEncoder creates a block encoder from a block split.
func newBlockEncoder(histogramLen, numBlockTypes int, types []byte, lengths []uint32) blockEncoder {
	numBlocks := len(types)
	blockLen := 0
	if numBlocks > 0 {
		blockLen = int(lengths[0])
	}
	return blockEncoder{
		histogramLen:  histogramLen,
		numBlockTypes: numBlockTypes,
		blockTypes:    types,
		blockLengths:  lengths,
		numBlocks:     numBlocks,
		blockLen:      blockLen,
	}
}

// initBlockTypeCodeCalc initializes the calculator. The C reference starts with
// lastType=1, secondLastType=0 so that the first block type (always 0) can
// use code 0 (second-last type).
func initBlockTypeCodeCalc() blockTypeCodeCalc {
	return blockTypeCodeCalc{lastType: 1, secondLastType: 0}
}

// nextBlockTypeCode returns the type code for switching to blockType and
// advances the state.
//
//	Code 0 = switch to second-last type
//	Code 1 = switch to (last type + 1) mod numTypes
//	Code 2+ = switch to type (code - 2)
func (c *blockTypeCodeCalc) nextBlockTypeCode(blockType byte) uint {
	var typeCode uint
	switch {
	case uint(blockType) == c.lastType+1:
		typeCode = 1
	case uint(blockType) == c.secondLastType:
		typeCode = 0
	default:
		typeCode = uint(blockType) + 2
	}
	c.secondLastType = c.lastType
	c.lastType = uint(blockType)
	return typeCode
}

// storeBlockSwitch writes a block type switch command to the bitstream.
func (code *blockSplitCode) storeBlockSwitch(blockLen uint32, blockType byte, isFirstBlock bool, b *bitWriter) {
	typeCode := code.typeCalc.nextBlockTypeCode(blockType)
	if !isFirstBlock {
		b.writeBits(uint(code.typeDepths[typeCode]), uint64(code.typeBits[typeCode]))
	}
	lenCode, lenNExtra, lenExtra := getBlockLengthPrefixCode(blockLen)
	b.writeBits(uint(code.lengthDepths[lenCode]), uint64(code.lengthBits[lenCode]))
	b.writeBits(uint(lenNExtra), uint64(lenExtra))
}

// buildAndStoreBlockSwitchEntropyCodes builds the block type/length Huffman
// codes and writes them to the bitstream.
func (enc *blockEncoder) buildAndStoreBlockSwitchEntropyCodes(tree []huffmanTreeNode, b *bitWriter) {
	buildAndStoreBlockSplitCode(enc.blockTypes, enc.blockLengths,
		enc.numBlocks, enc.numBlockTypes, tree, &enc.splitCode, b)
}

// buildAndStoreEntropyCodes builds a Huffman tree for each block type's
// histogram and writes all the trees to the bitstream. The depths and bits
// arrays are allocated here and stored for use by storeSymbol.
func (enc *blockEncoder) buildAndStoreEntropyCodes(
	histograms []uint32, numHistograms, alphabetSize int,
	tree []huffmanTreeNode, b *bitWriter,
	depthsBuf []byte, bitsBuf []uint16,
) ([]byte, []uint16) {
	tableSize := numHistograms * enc.histogramLen

	if cap(depthsBuf) < tableSize {
		depthsBuf = make([]byte, tableSize)
	} else {
		depthsBuf = depthsBuf[:tableSize]
	}

	if cap(bitsBuf) < tableSize {
		bitsBuf = make([]uint16, tableSize)
	} else {
		bitsBuf = bitsBuf[:tableSize]
	}

	enc.depths = depthsBuf
	enc.bits = bitsBuf

	for i := range numHistograms {
		ix := i * enc.histogramLen
		b.buildAndWriteHuffmanTree(
			histograms[ix:ix+enc.histogramLen],
			uint(alphabetSize),
			tree,
			enc.depths[ix:ix+enc.histogramLen],
			enc.bits[ix:ix+enc.histogramLen],
		)
	}

	return depthsBuf, bitsBuf
}

// storeSymbol writes the next symbol using the Huffman code of the current
// block type. At block boundaries, emits a block switch command.
func (enc *blockEncoder) storeSymbol(symbol uint, b *bitWriter) {
	if enc.blockLen == 0 {
		enc.emitBlockSwitch(b)
	}
	enc.blockLen--
	ix := enc.entropyIdx + int(symbol)
	b.writeBits(uint(enc.depths[ix]), uint64(enc.bits[ix]))
}

// emitBlockSwitch advances to the next block and writes the block switch
// command. Factored out of storeSymbol so the hot path remains small enough
// for the compiler to inline.
//
//go:noinline
func (enc *blockEncoder) emitBlockSwitch(b *bitWriter) {
	enc.blockIdx++
	blockLen := enc.blockLengths[enc.blockIdx]
	blockType := enc.blockTypes[enc.blockIdx]
	enc.blockLen = int(blockLen)
	enc.entropyIdx = int(blockType) * enc.histogramLen
	enc.splitCode.storeBlockSwitch(blockLen, blockType, false, b)
}

// storeSymbolWithContext writes the next symbol using the Huffman code
// selected by the context map. At block boundaries, emits a block switch
// command and updates entropyIdx to blockType << contextBits.
//
// Unlike storeSymbol (which indexes Huffman codes by blockType * histogramLen),
// this method uses the context map to select among clustered histograms:
//
//	histoIdx = contextMap[entropyIdx + context]
//	codeIdx  = histoIdx * histogramLen + symbol
func (enc *blockEncoder) storeSymbolWithContext(symbol, context uint, contextMap []uint32, contextBits uint, b *bitWriter) {
	if enc.blockLen == 0 {
		enc.blockIdx++
		blockLen := enc.blockLengths[enc.blockIdx]
		blockType := enc.blockTypes[enc.blockIdx]
		enc.blockLen = int(blockLen)
		enc.entropyIdx = int(blockType) << contextBits
		enc.splitCode.storeBlockSwitch(blockLen, blockType, false, b)
	}
	enc.blockLen--
	histoIdx := contextMap[uint(enc.entropyIdx)+context]
	ix := int(histoIdx)*enc.histogramLen + int(symbol)
	b.writeBits(uint(enc.depths[ix]), uint64(enc.bits[ix]))
}

// buildAndStoreBlockSplitCode builds Huffman codes for block type and block
// length symbols, writes the number of block types and the Huffman trees to
// the bitstream, and stores the first block switch.
func buildAndStoreBlockSplitCode(types []byte, lengths []uint32, numBlocks, numTypes int, tree []huffmanTreeNode, code *blockSplitCode, b *bitWriter) {
	var typeHisto [maxBlockTypeSymbols]uint32
	var lengthHisto [core.AlphabetSizeBlockCount]uint32

	calc := initBlockTypeCodeCalc()
	for i := range numBlocks {
		typeCode := calc.nextBlockTypeCode(types[i])
		if i != 0 {
			typeHisto[typeCode]++
		}
		lengthHisto[blockLengthPrefixCode(lengths[i])]++
	}

	b.storeVarLenUint8(uint(numTypes) - 1)
	if numTypes > 1 {
		alphabetSize := uint(numTypes + 2)
		b.buildAndWriteHuffmanTree(typeHisto[:alphabetSize], alphabetSize, tree,
			code.typeDepths[:], code.typeBits[:])
		b.buildAndWriteHuffmanTree(lengthHisto[:], core.AlphabetSizeBlockCount, tree,
			code.lengthDepths[:], code.lengthBits[:])
		code.typeCalc = initBlockTypeCodeCalc()
		code.storeBlockSwitch(lengths[0], types[0], true, b)
	}
}

// blockLengthPrefixCode returns just the prefix code for a block length
// (without extra bits).
func blockLengthPrefixCode(length uint32) uint32 {
	code, _, _ := getBlockLengthPrefixCode(length)
	return code
}

// storeTrivialContextMap encodes a context map where the histogram type
// is always the block type (identity mapping). Used at Q4 where no
// context modeling is performed.
func storeTrivialContextMap(numTypes, contextBits uint, tree []huffmanTreeNode, b *bitWriter) {
	b.storeVarLenUint8(numTypes - 1)
	if numTypes > 1 {
		repeatCode := contextBits - 1
		repeatBits := (uint64(1) << repeatCode) - 1
		alphabetSize := numTypes + repeatCode

		// Max alphabetSize = maxNumberOfBlockTypes + core.LiteralContextBits - 1 = 261.
		var histogramBuf [maxNumberOfBlockTypes + 5]uint32
		var depthsBuf [maxNumberOfBlockTypes + 5]byte
		var bitsBuf [maxNumberOfBlockTypes + 5]uint16
		histogram := histogramBuf[:alphabetSize]
		depths := depthsBuf[:alphabetSize]
		bits := bitsBuf[:alphabetSize]

		// Write RLEMAX.
		b.writeBits(1, 1)
		b.writeBits(4, uint64(repeatCode-1))

		histogram[repeatCode] = uint32(numTypes)
		histogram[0] = 1
		for i := contextBits; i < alphabetSize; i++ {
			histogram[i] = 1
		}

		b.buildAndWriteHuffmanTree(histogram, alphabetSize, tree, depths, bits)

		for i := range numTypes {
			code := uint(0)
			if i > 0 {
				code = i + contextBits - 1
			}
			b.writeBits(uint(depths[code]), uint64(bits[code]))
			b.writeBits(uint(depths[repeatCode]), uint64(bits[repeatCode]))
			b.writeBits(repeatCode, repeatBits)
		}
		// Write IMTF (inverse-move-to-front) bit.
		b.writeBits(1, 1)
	}
}
