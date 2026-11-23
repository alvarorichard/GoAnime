// Decoder state machine types, enums, and initialization.

package brrr

import (
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
	"github.com/molecule-man/go-brrr/internal/encoder"
)

// Decoder-specific constants.
const (
	// Maximum block/metablock length (used as default when no block switching).
	blockSizeCap = 1 << 24

	// Maximum Huffman table sizes for the decoder, indexed by alphabet size.
	huffmanMaxSize272 = 646 // for context map symbols (max 272)
	huffmanMaxSize258 = 632 // for block type symbols (max 258)
	huffmanMaxSize26  = 396 // for block count symbols (26)
)

const (
	decoderResultError decoderResult = iota
	decoderResultSuccess
	decoderResultNeedsMoreInput
	decoderResultNeedsMoreOutput
)

const (
	decoderStateUninited decoderState = iota
	decoderStateLargeWindowBits
	decoderStateInitialize
	decoderStateMetablockBegin
	decoderStateMetablockHeader
	decoderStateMetablockHeader2
	decoderStateContextModes
	decoderStateCommandBegin
	decoderStateCommandInner
	decoderStateCommandPostDecodeLiterals
	decoderStateCommandPostWrapCopy
	decoderStateUncompressed
	decoderStateMetadata
	decoderStateCommandInnerWrite
	decoderStateMetablockDone
	decoderStateCommandPostWrite1
	decoderStateCommandPostWrite2
	decoderStateBeforeCompressedMetablockHeader
	decoderStateHuffmanCode0
	decoderStateHuffmanCode1
	decoderStateHuffmanCode2
	decoderStateHuffmanCode3
	decoderStateContextMap1
	decoderStateContextMap2
	decoderStateTreeGroup
	decoderStateBeforeCompressedMetablockBody
	decoderStateDone
)

const (
	metablockHeaderNone metablockHeaderState = iota
	metablockHeaderEmpty
	metablockHeaderNibbles
	metablockHeaderSize
	metablockHeaderUncompressed
	metablockHeaderReserved
	metablockHeaderBytes
	metablockHeaderMetadata
)

const (
	uncompressedNone uncompressedState = iota
	uncompressedWrite
)

const (
	treeGroupNone treeGroupState = iota
	treeGroupLoop
)

const (
	contextMapNone contextMapState = iota
	contextMapReadPrefix
	contextMapHuffman
	contextMapDecode
	contextMapTransform
)

const (
	huffmanNone huffmanState = iota
	huffmanSimpleSize
	huffmanSimpleRead
	huffmanSimpleBuild
	huffmanComplex
	huffmanLengthSymbols
)

const (
	decodeUint8None decodeUint8State = iota
	decodeUint8Short
	decodeUint8Long
)

const (
	readBlockLengthNone readBlockLengthState = iota
	readBlockLengthSuffix
)

// decoderResult indicates the outcome of a decode step.
type decoderResult int

// decoderState enumerates the top-level states of the streaming decoder.
type decoderState int

// metablockHeaderState enumerates states within metablock header parsing.
type metablockHeaderState int

// uncompressedState enumerates states within uncompressed metablock processing.
type uncompressedState int

// treeGroupState enumerates states within Huffman tree group decoding.
type treeGroupState int

// contextMapState enumerates states within context map decoding.
type contextMapState int

// huffmanState enumerates states within Huffman code reading.
type huffmanState int

// decodeUint8State enumerates states within uint8 decoding.
type decodeUint8State int

// readBlockLengthState enumerates states within block length reading.
type readBlockLengthState int

// bitReader holds the bit-level reading state for the decoder input stream.
type bitReader struct {
	inputBase unsafe.Pointer // cached base pointer for bounds-check-free loads
	input     []byte         // full input buffer (set via setInput)
	pos       int            // current byte position in input
	val       uint64         // pre-fetched bits
	bitPos    uint           // number of valid bits in val
	inputLen  int            // cached len(input) to avoid slice header reads
	fastEnd   int            // inputLen - fastInputSlack; checkInputAmount is pos <= fastEnd
}

// huffmanTreeGroup holds a collection of Huffman trees sharing the same
// alphabet size. Used by the decoder to store literal, insert-copy, and
// distance Huffman trees.
type huffmanTreeGroup struct {
	htrees            []int              // indices into codes for each tree's root
	codes             []core.HuffmanCode // flat backing array for all trees
	alphabetSizeMax   uint16
	alphabetSizeLimit uint16
	numHTrees         uint16
}

// metablockHeaderArena holds temporary state used during metablock header
// decoding. Reused across metablocks.
type metablockHeaderArena struct {
	// List of heads of symbol chains.
	symbolLists []uint16

	substateTreeGroup  treeGroupState
	substateContextMap contextMapState
	substateHuffman    huffmanState

	subLoopCounter uint

	repeatCodeLen uint
	prevCodeLen   uint

	// For ReadHuffmanCode.
	symbol uint
	repeat uint
	space  uint

	// Huffman table for code length histograms.
	table            [32]core.HuffmanCode
	symbolListsArray [core.HuffmanMaxCodeLength + 1 + core.AlphabetSizeInsertAndCopyLength]uint16
	// Tails of symbol chains.
	nextSymbol            [32]int
	codeLengthCodeLengths [core.AlphabetSizeCodeLengths]byte
	// Population counts for the code lengths.
	codeLengthHisto [16]uint16

	// For HuffmanTreeGroupDecode.
	htreeIndex int
	next       int // index into codes slice

	// For DecodeContextMap.
	contextIndex       uint
	maxRunLengthPrefix uint
	code               uint
	contextMapTable    [huffmanMaxSize272]core.HuffmanCode
}

// metablockBodyArena holds temporary state used during metablock body
// decoding. Not used simultaneously with metablockHeaderArena.
type metablockBodyArena struct {
	distExtraBits [core.NumHistogramDistanceSymbols]byte
	distOffset    [core.NumHistogramDistanceSymbols]uint

	// Cache key for calculateDistanceLut: skip recomputation when params match.
	cachedPostfix         uint
	cachedDirect          uint
	cachedAlphabetSizeLim uint
	cachedValid           bool
}

// decoderCompoundDictionary holds compound dictionary chunks for the decoder.
// The decoder uses these chunks as a virtual LZ77 prefix: backward references
// beyond the ring buffer are resolved against the concatenated chunk data.
type decoderCompoundDictionary struct {
	chunks       [16][]byte
	chunkOffsets [17]int // cumulative byte offsets; chunkOffsets[0] = 0
	numChunks    int
	totalSize    int
	blockBits    int       // -1 until lazily initialized
	blockMap     [256]byte // maps address >> blockBits to chunk index
	// Copy state for suspension across ringbuffer writes.
	brIndex  int
	brOffset int
	brLength int
	brCopied int
}

// attach appends a dictionary chunk.
// decodeState holds the full state of a streaming brotli decoder.
//
// Fields are ordered with all pointer-containing types first (slices,
// interfaces, structs embedding slices) to minimize the GC scan range.
type decodeState struct {
	// Pointer-containing fields.
	br               bitReader
	err              error
	ringbuffer       []byte
	htreeCommand     []core.HuffmanCode // slice into insertCopyHGroup.codes
	htreeCommandLut  []uint64           // command LUT entries for the active tree
	cmdLuts          []uint64           // command LUT entries for all insert-and-copy trees
	contextLookup    []byte
	distContextMap   []byte
	literalHTree     []core.HuffmanCode // slice into literalHGroup.codes
	contextMap       []byte
	contextModes     []byte
	blockTypeTrees   []core.HuffmanCode
	blockLenTrees    []core.HuffmanCode
	literalHGroup    huffmanTreeGroup
	insertCopyHGroup huffmanTreeGroup
	distanceHGroup   huffmanTreeGroup
	compoundDict     *decoderCompoundDictionary
	sink             *chunkWriter
	outChunks        chunkWriter

	// literalCodesPtrs caches a precomputed unsafe.Pointer for each of the
	// 64 literal contexts. Each entry is codesBase + literalCodesOffsets[ctx]*4,
	// eliminating a multiply and add from the critical path in
	// decodeLiteralsContextBatch. Always populated alongside literalCodesOffsets
	// by prepareLiteralDecoding; never zeroed by initForReuse.
	// Placed here (in the pointer section) to keep all pointer-containing
	// fields contiguous and avoid extending the GC scan range into the
	// large scalar arrays.
	literalCodesPtrs [64]unsafe.Pointer

	// In C these are a union; Go embeds both since they are never used
	// simultaneously. headerArena contains a slice so it must precede
	// scalar-only fields.
	headerArena metablockHeaderArena

	// Scalar fields (no pointers beyond this point).
	state       decoderState
	loopCounter int // reused for several disjoint loops

	// Temporary storage for remaining input during streaming. When the
	// decoder runs out of input mid-operation, unread bytes from br.input
	// are stashed here. The bitReader is then repointed at this buffer to
	// resume decoding. 64 bits are enough to make progress in decoding.
	bufferLength uint
	buffer       [8]byte

	pos                   int
	maxBackwardDistance   int
	maxDistance           int
	ringbufferSize        int
	ringbufferMask        int
	distRBIdx             int
	distRB                [4]int
	safeCmdMemento        bitReaderState // saved before safe command symbol read
	metaBlockRemainingLen int
	// Indices into contextMap and distContextMap slices.
	contextMapSliceIdx     int
	distContextMapSliceIdx int

	// True if the literal context map histogram type always matches the block
	// type, so context tracking is not needed (faster decoding).
	trivialLiteralContext int
	// Actual after command decode, before distance computation; then reused
	// as a temporary variable.
	distanceContext        int
	blockLength            [3]uint
	blockLengthIndex       uint
	numBlockTypes          [3]uint
	blockTypeRB            [6]uint
	distancePostfixBits    uint
	numDirectDistanceCodes uint
	numDistHTrees          uint
	numLiteralHTrees       uint

	// For partial write operations.
	rbRoundtrips  uint // how many times we wrapped around the ring buffer
	partialPosOut uint // total output to the user

	// For InverseMoveToFrontTransform.
	mtfUpperBound uint

	copyLength   int
	distanceCode int

	// States inside function calls.
	substateMetablockHeader metablockHeaderState
	substateUncompressed    uncompressedState
	substateDecodeUint8     decodeUint8State
	substateReadBlockLength readBlockLengthState

	newRingbufferSize int
	windowBits        int
	sizeNibbles       int

	trivialLiteralContexts [8]uint32 // 256 bits

	isLastMetablock           bool
	isUncompressed            bool
	isMetadata                bool
	cannyRingbufferAllocation bool
	distHTreeIndex            byte
	distCodesOffset           int // cached s.distanceHGroup.htrees[distHTreeIndex]

	// distCodesCache maps each of the 4 distance contexts to the
	// corresponding htrees offset for the current distance block type.
	// Avoids two dependent memory loads (distContextMap → htrees)
	// per command in processCommands.
	distCodesCache [4]int

	// The C reference exposes BrotliDecoderSetMetadataCallbacks (decode.h)
	// which lets callers receive metadata block contents via a start/chunk
	// callback pair. The decoder already skips metadata blocks correctly;
	// exposing their contents is a future enhancement. The idiomatic Go
	// approach would be a MetadataHandler interface or io.Writer rather than
	// C-style function pointers.

	bodyArena metablockBodyArena

	// literalCodesOffsets maps each of the 64 literal contexts to the
	// corresponding codes offset for the current literal block type.
	// Avoids two dependent memory loads (contextMap → htrees)
	// per literal in the non-trivial context path of processCommands.
	// Placed at end of struct to avoid shifting hot fields.
	literalCodesOffsets [64]int
}

func (g *huffmanTreeGroup) init(alphabetSizeMax, alphabetSizeLimit, ntrees uint16) {
	maxTableSize := int(alphabetSizeLimit) + 376
	g.alphabetSizeMax = alphabetSizeMax
	g.alphabetSizeLimit = alphabetSizeLimit
	g.numHTrees = ntrees
	g.codes = reuseHuffmanCodes(g.codes, int(ntrees)*maxTableSize)
	g.htrees = reuseInts(g.htrees, int(ntrees))
}

func (cd *decoderCompoundDictionary) attach(data []byte) error {
	if len(data) == 0 {
		return encoder.ErrEmptyDict
	}
	if cd.numChunks == 15 {
		return encoder.ErrTooManyDicts
	}
	cd.chunks[cd.numChunks] = data
	cd.numChunks++
	cd.totalSize += len(data)
	cd.chunkOffsets[cd.numChunks] = cd.totalSize
	return nil
}

// ensureInit lazily builds the block map for O(1) chunk lookup.
func (cd *decoderCompoundDictionary) ensureInit() {
	if cd.blockBits != -1 {
		return
	}
	blockBits := 8
	for (cd.totalSize-1)>>blockBits != 0 {
		blockBits++
	}
	blockBits -= 8
	cd.blockBits = blockBits

	cursor := 0
	index := 0
	for cursor < cd.totalSize {
		for cd.chunkOffsets[index+1] < cursor {
			index++
		}
		cd.blockMap[cursor>>blockBits] = byte(index)
		cursor += 1 << blockBits
	}
}

// initCopy sets up a copy operation from compound dictionary at address for
// length bytes. Updates the distance ring buffer and decrements
// metaBlockRemainingLen. Returns false if the address+length exceeds the
// total dictionary size.
func (cd *decoderCompoundDictionary) initCopy(s *decodeState, address, length int) bool {
	cd.ensureInit()
	index := int(cd.blockMap[address>>cd.blockBits])
	for address >= cd.chunkOffsets[index+1] {
		index++
	}
	if cd.totalSize < address+length {
		return false
	}
	s.distRB[s.distRBIdx&3] = s.distanceCode
	s.distRBIdx++
	s.metaBlockRemainingLen -= length
	cd.brIndex = index
	cd.brOffset = address - cd.chunkOffsets[index]
	cd.brLength = length
	cd.brCopied = 0
	return true
}

// copyTo copies bytes from compound dictionary chunks into s.ringbuffer[pos:],
// handling chunk boundaries and stopping when the ringbuffer is full.
// Returns the number of bytes copied.
func (cd *decoderCompoundDictionary) copyTo(s *decodeState, pos int) int {
	origPos := pos
	for cd.brLength != cd.brCopied {
		copySrc := cd.chunks[cd.brIndex][cd.brOffset:]
		space := s.ringbufferSize - pos
		remChunkLen := (cd.chunkOffsets[cd.brIndex+1] - cd.chunkOffsets[cd.brIndex]) - cd.brOffset
		length := min(cd.brLength-cd.brCopied, min(remChunkLen, space))
		copy(s.ringbuffer[pos:], copySrc[:length])
		pos += length
		cd.brOffset += length
		cd.brCopied += length
		if length == remChunkLen {
			cd.brIndex++
			cd.brOffset = 0
		}
		if pos == s.ringbufferSize {
			break
		}
	}
	return pos - origPos
}

// init resets s to a clean initial state, ready to begin decoding.
func (s *decodeState) init() {
	*s = decodeState{}
	s.initDefaults()
}

func (s *decodeState) initDefaults() {

	s.state = decoderStateUninited
	s.cannyRingbufferAllocation = true

	s.distRB = [4]int{16, 15, 11, 4}

	s.mtfUpperBound = 63
}

func (s *decodeState) initForReuse() {
	// Reset the decoder state for reuse, preserving heap-allocated backing
	// arrays. Unlike the previous *s = decodeState{} approach, this avoids
	// zeroing ~10 KB of arena arrays (headerArena.contextMapTable,
	// headerArena.symbolListsArray, bodyArena, literalCodesOffsets) that are
	// always fully initialized before being read.

	// Reset only the scalar bitReader fields that must be zero.
	// inputBase and input are stale but will be overwritten by setInput
	// before any read, so skipping those two pointer-bearing fields avoids
	// two GC write barriers per Decompress call.
	s.br.val = 0
	s.br.bitPos = 0
	s.br.pos = 0
	s.br.inputLen = 0
	s.br.fastEnd = 0
	s.err = nil

	// Nil out slices that are derived from preserved slices or set during
	// decompression. The saved backing slices (ringbuffer, blockTypeTrees,
	// contextMap, etc.) are kept; derived slices are re-derived later.
	s.htreeCommand = nil
	s.htreeCommandLut = nil
	s.contextLookup = nil
	s.literalHTree = nil
	s.blockLenTrees = nil
	s.compoundDict = nil

	// Zero the huffmanTreeGroup metadata (preserved slice headers stay).
	s.literalHGroup.alphabetSizeMax = 0
	s.literalHGroup.alphabetSizeLimit = 0
	s.literalHGroup.numHTrees = 0
	s.insertCopyHGroup.alphabetSizeMax = 0
	s.insertCopyHGroup.alphabetSizeLimit = 0
	s.insertCopyHGroup.numHTrees = 0
	s.distanceHGroup.alphabetSizeMax = 0
	s.distanceHGroup.alphabetSizeLimit = 0
	s.distanceHGroup.numHTrees = 0

	// Zero only the small scalar fields in headerArena; the large arrays
	// (table, symbolListsArray, nextSymbol, codeLengthCodeLengths,
	// codeLengthHisto, contextMapTable) are always initialized before use.
	s.headerArena.symbolLists = nil
	s.headerArena.substateTreeGroup = 0
	s.headerArena.substateContextMap = 0
	s.headerArena.substateHuffman = 0
	s.headerArena.subLoopCounter = 0
	s.headerArena.repeatCodeLen = 0
	s.headerArena.prevCodeLen = 0
	s.headerArena.symbol = 0
	s.headerArena.repeat = 0
	s.headerArena.space = 0
	s.headerArena.htreeIndex = 0
	s.headerArena.next = 0
	s.headerArena.contextIndex = 0
	s.headerArena.maxRunLengthPrefix = 0
	s.headerArena.code = 0

	// Bulk-zero the contiguous scalar block (state through distCodesCache).
	// This region contains no pointers, so raw memclr is safe.
	// bodyArena and literalCodesOffsets are NOT zeroed — they are always
	// fully written by calculateDistanceLut / prepareLiteralDecoding
	// before being read.
	start := unsafe.Pointer(&s.state)
	size := uintptr(unsafe.Pointer(&s.distCodesCache)) - uintptr(start) + unsafe.Sizeof(s.distCodesCache)
	clear(unsafe.Slice((*byte)(start), size))

	s.initDefaults()
}

// metablockBegin resets per-metablock state at the start of each metablock.
func (s *decodeState) metablockBegin() {
	s.metaBlockRemainingLen = 0
	s.blockLength = [3]uint{blockSizeCap, blockSizeCap, blockSizeCap}
	s.numBlockTypes = [3]uint{1, 1, 1}
	s.blockTypeRB = [6]uint{1, 0, 1, 0, 1, 0}
	s.contextMap = s.contextMap[:0]
	s.contextModes = s.contextModes[:0]
	s.distContextMap = s.distContextMap[:0]
	s.contextMapSliceIdx = 0
	s.distContextMapSliceIdx = 0
	s.distHTreeIndex = 0
	s.distCodesOffset = 0
	s.distCodesCache = [4]int{}
	// literalHTree and contextLookup are always overwritten by
	// prepareLiteralDecoding before use — skipping the nil assignments here
	// avoids two GC write barriers per metablock.
	// literalHGroup/insertCopyHGroup/distanceHGroup .codes and .htrees are
	// always overwritten by huffmanTreeGroup.init via reuseHuffmanCodes /
	// reuseInts (which only check cap, not len) before use — skipping the
	// [:0] reslices here avoids six more GC write barriers per metablock.
}

// updateDistCodesCache fills distCodesCache with the htrees offset for each
// of the 4 distance contexts at the current distContextMapSliceIdx.
func (s *decodeState) updateDistCodesCache() {
	base := s.distContextMapSliceIdx
	for ctx := range s.distCodesCache {
		idx := s.distContextMap[base+ctx]
		s.distCodesCache[ctx] = s.distanceHGroup.htrees[idx]
	}
}

// stashTail copies unread bytes from br.input into the internal buffer
// for later resumption. Called when the decoder needs more input while
// reading directly from caller-supplied data.
func (s *decodeState) stashTail() {
	s.br.unload()
	avail := min(s.br.availIn(), len(s.buffer))
	copy(s.buffer[:avail], s.br.input[s.br.pos:s.br.pos+avail])
	s.bufferLength = uint(avail)
}

// useStashedTail repoints the bitReader at the internal buffer so
// decoding can resume from the stashed tail bytes.
func (s *decodeState) useStashedTail() {
	s.br.setInput(s.buffer[:s.bufferLength])
}

// compoundDictSize returns the total size of attached compound dictionaries,
// or 0 if none are attached.
func (s *decodeState) compoundDictSize() int {
	if s.compoundDict == nil {
		return 0
	}
	return s.compoundDict.totalSize
}

// attachCompoundDict appends a compound dictionary chunk to the decoder state.
// Lazily allocates the decoderCompoundDictionary on first call.
func (s *decodeState) attachCompoundDict(data []byte) error {
	if s.compoundDict == nil {
		s.compoundDict = &decoderCompoundDictionary{blockBits: -1}
	}
	return s.compoundDict.attach(data)
}

// No explicit cleanup method is needed for decodeState: all resources are
// Go-managed (slices, pointers), and Reader.Close already zeroes the state.

func reuseBytes(buf []byte, size int) []byte {
	if cap(buf) >= size {
		return buf[:size]
	}
	return make([]byte, size)
}

func reuseInts(buf []int, size int) []int {
	if cap(buf) >= size {
		return buf[:size]
	}
	return make([]int, size)
}

func reuseHuffmanCodes(buf []core.HuffmanCode, size int) []core.HuffmanCode {
	if cap(buf) >= size {
		return buf[:size]
	}
	return make([]core.HuffmanCode, size)
}
