// Brotli decompression via a suspendable state machine (dec/decode.c).

package brrr

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

// Decoder-specific constants.
const (
	huffmanTableBits          = 8
	huffmanTableMask          = 0xFF
	ringBufferWriteAheadSlack = 542
	maxAllowedDistance        = 0x7FFFFFFC

	// symListBase is the offset into symbolListsArray that makes C-style
	// negative indices in nextSymbol[] addressable.
	symListBase = core.HuffmanMaxCodeLength + 1
)

// Static prefix code for the complex code length code lengths (RFC 7932 §3.5).
var codeLengthPrefixLength = [16]byte{
	2, 2, 2, 3, 2, 2, 2, 4, 2, 2, 2, 3, 2, 2, 2, 4,
}

var codeLengthPrefixValue = [16]byte{
	0, 4, 3, 2, 0, 4, 3, 1, 0, 4, 3, 2, 0, 4, 3, 5,
}

// ErrExcessiveInput reports bytes that follow finalized brotli stream.
var ErrExcessiveInput = errors.New("brotli: excessive input")

// Sentinel errors returned by Decompress.
var errPadding = errors.New("brotli: non-zero padding bits")

var decodeStatePool = sync.Pool{
	New: func() any {
		s := new(decodeState)
		s.init()
		s.outChunks = chunkWriter{pool: &decodeChunkPool, size: decodeChunkSize}
		return s
	},
}

// decRingBufPool caches decoder ring buffer slices to reduce allocation
// pressure in streaming (Reader) decompression, where a new decodeState is
// created per Reader. The oneshot Decompress path already pools the entire
// decodeState (including its ring buffer).
var decRingBufPool sync.Pool

var cmdLutPacked = packCmdLut()

func packCmdLut() (t [core.AlphabetSizeInsertAndCopyLength]uint64) {
	for i := range core.CmdLut {
		e := &core.CmdLut[i]
		t[i] = uint64(e.InsertLenExtraBits) | uint64(e.CopyLenExtraBits)<<8 | uint64(uint8(e.DistanceCode))<<16 |
			uint64(e.Context)<<24 | uint64(e.InsertLenOffset)<<32 | uint64(e.CopyLenOffset)<<48
	}
	return t
}

// resetCmdLuts clears one command LUT entry for each root slot in every
// insert-and-copy tree. The command decoder caches entries on first use.
// Zero marks an empty slot because cmdLutPacked contains no zero entries.
func (s *decodeState) resetCmdLuts() {
	n := int(s.insertCopyHGroup.numHTrees) << huffmanTableBits
	if cap(s.cmdLuts) < n {
		s.cmdLuts = make([]uint64, n)
	}
	s.cmdLuts = s.cmdLuts[:n]
	clear(s.cmdLuts)
}

func getDecRingBuf(size int) []byte {
	if v := decRingBufPool.Get(); v != nil {
		bp := v.(*[]byte)
		if cap(*bp) >= size {
			return (*bp)[:size]
		}
		// Wrong size — let GC reclaim it.
	}
	return make([]byte, size)
}

func putDecRingBuf(buf []byte) {
	if buf != nil {
		b := buf[:cap(buf)]
		decRingBufPool.Put(&b)
	}
}

// Decompress decodes the brotli-compressed data and returns the original bytes.
//
// It returns [ErrExcessiveInput] if bytes follow the stream.
func Decompress(data []byte) ([]byte, error) {
	s := decodeStatePool.Get().(*decodeState)
	s.initForReuse()
	s.br.setInput(data)
	s.sink = &s.outChunks
	err := s.decodeAll()
	s.sink = nil
	if err != nil {
		s.outChunks.discard()
		decodeStatePool.Put(s)
		return nil, err
	}
	out := s.outChunks.take()
	decodeStatePool.Put(s)
	return out, nil
}

func (s *decodeState) decodeAll() error {
	var output []byte
	for {
		switch s.decompressStream(&output) {
		case decoderResultSuccess:
			if s.excessiveInput() {
				return ErrExcessiveInput
			}
			s.flushOutput(output)
			return nil
		case decoderResultError:
			return s.err
		case decoderResultNeedsMoreInput:
			return decompressError("truncated input")
		case decoderResultNeedsMoreOutput:
			output = s.flushOutput(output)
		}
	}
}

// decompressStream advances the decoder state machine. Returns a result
// indicating success, error, or that more input/output is needed.
func (s *decodeState) decompressStream(output *[]byte) decoderResult {
	for {
		switch s.state {
		case decoderStateUninited:
			if !s.br.warmup() {
				return decoderResultNeedsMoreInput
			}
			if err := s.decodeWindowBits(); err != nil {
				s.err = err
				return decoderResultError
			}
			s.state = decoderStateInitialize
			fallthrough

		case decoderStateInitialize:
			s.maxBackwardDistance = (1 << s.windowBits) - core.WindowGap
			allTrees := reuseHuffmanCodes(s.blockTypeTrees, 3*(huffmanMaxSize258+huffmanMaxSize26))
			s.blockTypeTrees = allTrees[:3*huffmanMaxSize258]
			s.blockLenTrees = allTrees[3*huffmanMaxSize258:]
			s.state = decoderStateMetablockBegin
			fallthrough

		case decoderStateMetablockBegin:
			s.metablockBegin()
			s.state = decoderStateMetablockHeader
			fallthrough

		case decoderStateMetablockHeader:
			result := s.decodeMetaBlockLength()
			if result != decoderResultSuccess {
				return result
			}
			if s.isMetadata || s.isUncompressed {
				if !s.br.jumpToByteBoundary() {
					s.err = errPadding
					return decoderResultError
				}
			}
			if s.isMetadata {
				s.state = decoderStateMetadata
				break
			}
			if s.metaBlockRemainingLen == 0 {
				s.state = decoderStateMetablockDone
				break
			}
			s.calculateRingBufferSize()
			if s.isUncompressed {
				s.state = decoderStateUncompressed
				break
			}
			s.state = decoderStateBeforeCompressedMetablockHeader
			fallthrough

		case decoderStateBeforeCompressedMetablockHeader:
			h := &s.headerArena
			s.loopCounter = 0
			h.subLoopCounter = 0
			h.symbolLists = h.symbolListsArray[core.HuffmanMaxCodeLength+1:]
			h.substateHuffman = huffmanNone
			h.substateTreeGroup = treeGroupNone
			h.substateContextMap = contextMapNone
			s.state = decoderStateHuffmanCode0
			fallthrough

		case decoderStateHuffmanCode0:
			if s.loopCounter >= 3 {
				s.state = decoderStateMetablockHeader2
				break
			}
			result := s.decodeVarLenUint8(&s.numBlockTypes[s.loopCounter])
			if result != decoderResultSuccess {
				return result
			}
			s.numBlockTypes[s.loopCounter]++
			if s.numBlockTypes[s.loopCounter] < 2 {
				s.loopCounter++
				break
			}
			s.state = decoderStateHuffmanCode1
			fallthrough

		case decoderStateHuffmanCode1:
			alphabetSize := s.numBlockTypes[s.loopCounter] + 2
			treeOffset := s.loopCounter * huffmanMaxSize258
			result := s.readHuffmanCode(alphabetSize, alphabetSize,
				s.blockTypeTrees[treeOffset:], nil)
			if result != decoderResultSuccess {
				return result
			}
			s.state = decoderStateHuffmanCode2
			fallthrough

		case decoderStateHuffmanCode2:
			treeOffset := s.loopCounter * huffmanMaxSize26
			result := s.readHuffmanCode(core.AlphabetSizeBlockCount, core.AlphabetSizeBlockCount,
				s.blockLenTrees[treeOffset:], nil)
			if result != decoderResultSuccess {
				return result
			}
			s.state = decoderStateHuffmanCode3
			fallthrough

		case decoderStateHuffmanCode3:
			treeOffset := s.loopCounter * huffmanMaxSize26
			bl, ok := safeReadBlockLength(s, s.blockLenTrees[treeOffset:])
			if !ok {
				return decoderResultNeedsMoreInput
			}
			s.blockLength[s.loopCounter] = bl
			s.loopCounter++
			s.state = decoderStateHuffmanCode0

		case decoderStateMetablockHeader2:
			bits, ok := s.br.safeReadBits(6)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			s.distancePostfixBits = uint(bits & bitMask(2))
			bits >>= 2
			s.numDirectDistanceCodes = uint(bits) << s.distancePostfixBits
			s.contextModes = reuseBytes(s.contextModes, int(s.numBlockTypes[0]))
			s.loopCounter = 0
			s.state = decoderStateContextModes
			fallthrough

		case decoderStateContextModes:
			result := s.readContextModes()
			if result != decoderResultSuccess {
				return result
			}
			s.state = decoderStateContextMap1
			fallthrough

		case decoderStateContextMap1:
			result := s.decodeContextMap(s.numBlockTypes[0]<<core.LiteralContextBits, &s.contextMap, &s.numLiteralHTrees)
			if result != decoderResultSuccess {
				return result
			}
			s.detectTrivialLiteralBlockTypes()
			s.state = decoderStateContextMap2
			fallthrough

		case decoderStateContextMap2:
			npostfix := s.distancePostfixBits
			ndirect := s.numDirectDistanceCodes
			distAlphabetSizeMax := uint(core.NumDistanceShortCodes) + ndirect + (uint(core.MaxDistanceBits) << (npostfix + 1))
			distAlphabetSizeLimit := distAlphabetSizeMax

			result := s.decodeContextMap(s.numBlockTypes[2]<<core.DistanceContextBits, &s.distContextMap, &s.numDistHTrees)
			if result != decoderResultSuccess {
				return result
			}

			s.literalHGroup.init(uint16(core.AlphabetSizeLiteral), uint16(core.AlphabetSizeLiteral), uint16(s.numLiteralHTrees))
			s.insertCopyHGroup.init(uint16(core.AlphabetSizeInsertAndCopyLength), uint16(core.AlphabetSizeInsertAndCopyLength), uint16(s.numBlockTypes[1]))
			s.distanceHGroup.init(uint16(distAlphabetSizeMax), uint16(distAlphabetSizeLimit), uint16(s.numDistHTrees))

			s.loopCounter = 0
			s.state = decoderStateTreeGroup
			fallthrough

		case decoderStateTreeGroup:
			var hgroup *huffmanTreeGroup
			switch s.loopCounter {
			case 0:
				hgroup = &s.literalHGroup
			case 1:
				hgroup = &s.insertCopyHGroup
			case 2:
				hgroup = &s.distanceHGroup
			default:
				s.err = decompressError("unreachable tree group index")
				return decoderResultError
			}
			result := s.huffmanTreeGroupDecode(hgroup)
			if result != decoderResultSuccess {
				return result
			}
			s.loopCounter++
			if s.loopCounter < 3 {
				break
			}
			s.state = decoderStateBeforeCompressedMetablockBody
			fallthrough

		case decoderStateBeforeCompressedMetablockBody:
			s.prepareLiteralDecoding()
			s.distContextMapSliceIdx = 0
			s.updateDistCodesCache()
			s.resetCmdLuts()
			s.htreeCommand = s.insertCopyHGroup.codes[s.insertCopyHGroup.htrees[0]:]
			s.htreeCommandLut = s.cmdLuts
			s.ensureRingBuffer()
			s.calculateDistanceLut()
			s.state = decoderStateCommandBegin
			fallthrough

		case decoderStateCommandBegin,
			decoderStateCommandInner,
			decoderStateCommandPostDecodeLiterals,
			decoderStateCommandPostWrapCopy:
			result := s.processCommands()
			if result != decoderResultSuccess {
				// The caller may write the ring buffer out next. C refuses that
				// after a command ran past the metablock (BLOCK_LENGTH_1).
				if result != decoderResultError && s.metaBlockRemainingLen < 0 {
					s.err = decompressError("negative remaining metablock length")
					return decoderResultError
				}
				return result
			}
			// processCommands sets s.state to MetablockDone on success.

		case decoderStateCommandInnerWrite,
			decoderStateCommandPostWrite1,
			decoderStateCommandPostWrite2:
			s.writeRingBuffer(output)
			switch s.state {
			case decoderStateCommandPostWrite1:
				if cd := s.compoundDict; cd != nil && cd.brLength != cd.brCopied {
					s.pos += cd.copyTo(s, s.pos)
					if s.pos >= s.ringbufferSize {
						continue
					}
				}
				if s.metaBlockRemainingLen == 0 {
					s.state = decoderStateMetablockDone
				} else {
					s.state = decoderStateCommandBegin
				}
			case decoderStateCommandPostWrite2:
				s.state = decoderStateCommandPostWrapCopy
			default: // CommandInnerWrite
				if s.loopCounter == 0 {
					if s.metaBlockRemainingLen == 0 {
						s.state = decoderStateMetablockDone
					} else {
						s.state = decoderStateCommandPostDecodeLiterals
					}
				} else {
					s.state = decoderStateCommandInner
				}
			}

		case decoderStateMetadata:
			result := s.skipMetadataBlock()
			if result != decoderResultSuccess {
				return result
			}
			s.state = decoderStateMetablockDone

		case decoderStateUncompressed:
			result := s.copyUncompressedBlock(output)
			if result != decoderResultSuccess {
				return result
			}
			s.state = decoderStateMetablockDone

		case decoderStateMetablockDone:
			if s.metaBlockRemainingLen < 0 {
				s.err = decompressError("negative remaining metablock length")
				return decoderResultError
			}
			if !s.isLastMetablock {
				s.state = decoderStateMetablockBegin
				break
			}
			if !s.br.jumpToByteBoundary() {
				s.err = errPadding
				return decoderResultError
			}
			s.state = decoderStateDone
			fallthrough

		case decoderStateDone:
			return decoderResultSuccess

		default:
			s.err = decompressError("unhandled decoder state")
			return decoderResultError
		}
	}
}

// pendingOutput aliases unread ring-buffer bytes until decoding resumes.
// writeRingBuffer handles bytes in the write-ahead slack.
func (s *decodeState) pendingOutput() []byte {
	// After an overrun the ring buffer holds bytes past the metablock.
	if s.metaBlockRemainingLen < 0 {
		return nil
	}
	pos := min(s.pos, s.ringbufferSize)
	flushed := int(s.partialPosOut) - int(s.rbRoundtrips)*s.ringbufferSize
	if pos <= flushed {
		return nil
	}
	return s.ringbuffer[flushed:pos]
}

func (s *decodeState) consumeOutput(n int) {
	s.partialPosOut += uint(n)
}

func (s *decodeState) flushOutput(output []byte) []byte {
	pending := s.pendingOutput()
	if len(pending) == 0 {
		return output
	}
	if s.sink != nil {
		s.sink.write(pending)
	} else {
		output = append(output, pending...)
	}
	s.consumeOutput(len(pending))
	return output
}

// writeRingBuffer flushes the ring buffer when it fills, appending to output.
func (s *decodeState) writeRingBuffer(output *[]byte) {
	if s.pos >= s.ringbufferSize {
		if s.ringbufferSize < 1<<s.windowBits {
			// Ring buffer is smaller than the window. Grow it to fit pos
			// without jumping to full window size, to avoid prematurely
			// setting maxDistance = maxBackwardDistance (which would corrupt
			// static dictionary address computation).
			s.ensureRingBuffer()
			newSize := s.ringbufferSize
			for newSize <= s.pos {
				newSize = min(newSize<<1, 1<<s.windowBits)
			}
			if newSize != s.ringbufferSize {
				if cap(s.ringbuffer) >= newSize+ringBufferWriteAheadSlack {
					s.ringbuffer = s.ringbuffer[:newSize+ringBufferWriteAheadSlack]
				} else {
					newBuf := getDecRingBuf(newSize + ringBufferWriteAheadSlack)
					copy(newBuf, s.ringbuffer[:s.pos])
					putDecRingBuf(s.ringbuffer)
					s.ringbuffer = newBuf
				}
				s.ringbufferSize = newSize
				s.ringbufferMask = newSize - 1
				s.newRingbufferSize = newSize
			}
			if s.pos < s.ringbufferSize {
				return
			}
		}
		if s.sink != nil {
			s.sink.write(s.ringbuffer[int(s.partialPosOut)-int(s.rbRoundtrips)*s.ringbufferSize : s.ringbufferSize])
		} else {
			*output = append(*output, s.ringbuffer[int(s.partialPosOut)-int(s.rbRoundtrips)*s.ringbufferSize:s.ringbufferSize]...)
		}
		s.partialPosOut = s.rbRoundtrips*uint(s.ringbufferSize) + uint(s.ringbufferSize)
		s.pos -= s.ringbufferSize
		s.rbRoundtrips++
		if s.pos > 0 {
			copy(s.ringbuffer, s.ringbuffer[s.ringbufferSize:s.ringbufferSize+s.pos])
		}
		s.maxDistance = s.maxBackwardDistance
	}
}

// --- Window bits ---

func (s *decodeState) decodeWindowBits() error {
	br := &s.br
	n := br.takeBits(1)
	if n == 0 {
		s.windowBits = 16
		return nil
	}
	n = br.takeBits(3)
	if n != 0 {
		s.windowBits = int(17+n) & 63
		return nil
	}
	n = br.takeBits(3)
	if n == 1 {
		return decompressError("invalid window bits")
	}
	if n != 0 {
		s.windowBits = int(8+n) & 63
		return nil
	}
	s.windowBits = 17
	return nil
}

// --- Variable-length uint8 ---

func (s *decodeState) decodeVarLenUint8(value *uint) decoderResult {
	br := &s.br
	switch s.substateDecodeUint8 {
	case decodeUint8None:
		bits, ok := br.safeReadBits(1)
		if !ok {
			return decoderResultNeedsMoreInput
		}
		if bits == 0 {
			*value = 0
			return decoderResultSuccess
		}
		fallthrough

	case decodeUint8Short:
		bits, ok := br.safeReadBits(3)
		if !ok {
			s.substateDecodeUint8 = decodeUint8Short
			return decoderResultNeedsMoreInput
		}
		if bits == 0 {
			*value = 1
			s.substateDecodeUint8 = decodeUint8None
			return decoderResultSuccess
		}
		// Use *value as temporary storage for nbits (must be persisted).
		*value = uint(bits)
		fallthrough

	case decodeUint8Long:
		bits, ok := br.safeReadBits(*value)
		if !ok {
			s.substateDecodeUint8 = decodeUint8Long
			return decoderResultNeedsMoreInput
		}
		*value = (1 << *value) + uint(bits)
		s.substateDecodeUint8 = decodeUint8None
		return decoderResultSuccess
	}
	s.err = decompressError("unreachable decodeVarLenUint8 state")
	return decoderResultError
}

// --- Metablock header ---

func (s *decodeState) decodeMetaBlockLength() decoderResult {
	br := &s.br
	for {
		switch s.substateMetablockHeader {
		case metablockHeaderNone:
			bits, ok := br.safeReadBits(1)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			s.isLastMetablock = bits != 0
			s.metaBlockRemainingLen = 0
			s.isUncompressed = false
			s.isMetadata = false
			if !s.isLastMetablock {
				s.substateMetablockHeader = metablockHeaderNibbles
				break
			}
			s.substateMetablockHeader = metablockHeaderEmpty
			fallthrough

		case metablockHeaderEmpty:
			bits, ok := br.safeReadBits(1)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			if bits != 0 {
				// Empty last metablock.
				s.substateMetablockHeader = metablockHeaderNone
				return decoderResultSuccess
			}
			s.substateMetablockHeader = metablockHeaderNibbles
			fallthrough

		case metablockHeaderNibbles:
			bits, ok := br.safeReadBits(2)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			s.sizeNibbles = int(bits + 4)
			s.loopCounter = 0
			if bits == 3 {
				s.isMetadata = true
				s.substateMetablockHeader = metablockHeaderReserved
				break
			}
			s.substateMetablockHeader = metablockHeaderSize
			fallthrough

		case metablockHeaderSize:
			i := s.loopCounter
			for i < s.sizeNibbles {
				bits, ok := br.safeReadBits(4)
				if !ok {
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
				if i+1 == s.sizeNibbles && s.sizeNibbles > 4 && bits == 0 {
					s.err = decompressError("exuberant nibble")
					return decoderResultError
				}
				s.metaBlockRemainingLen |= int(bits) << (i * 4)
				i++
			}
			s.substateMetablockHeader = metablockHeaderUncompressed
			fallthrough

		case metablockHeaderUncompressed:
			if !s.isLastMetablock {
				bits, ok := br.safeReadBits(1)
				if !ok {
					return decoderResultNeedsMoreInput
				}
				s.isUncompressed = bits != 0
			}
			s.metaBlockRemainingLen++
			s.substateMetablockHeader = metablockHeaderNone
			return decoderResultSuccess

		case metablockHeaderReserved:
			bits, ok := br.safeReadBits(1)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			if bits != 0 {
				s.err = decompressError("reserved bit set in metadata header")
				return decoderResultError
			}
			s.substateMetablockHeader = metablockHeaderBytes
			fallthrough

		case metablockHeaderBytes:
			bits, ok := br.safeReadBits(2)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			if bits == 0 {
				s.substateMetablockHeader = metablockHeaderNone
				return decoderResultSuccess
			}
			s.sizeNibbles = int(bits)
			s.loopCounter = 0
			s.substateMetablockHeader = metablockHeaderMetadata
			fallthrough

		case metablockHeaderMetadata:
			i := s.loopCounter
			for i < s.sizeNibbles {
				bits, ok := br.safeReadBits(8)
				if !ok {
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
				if i+1 == s.sizeNibbles && s.sizeNibbles > 1 && bits == 0 {
					s.err = decompressError("exuberant metadata nibble")
					return decoderResultError
				}
				s.metaBlockRemainingLen |= int(bits) << (i * 8)
				i++
			}
			s.metaBlockRemainingLen++
			s.substateMetablockHeader = metablockHeaderNone
			return decoderResultSuccess

		default:
			s.err = decompressError("unreachable metablock header state")
			return decoderResultError
		}
	}
}

// --- Metadata skip ---

func (s *decodeState) skipMetadataBlock() decoderResult {
	br := &s.br
	for s.metaBlockRemainingLen > 0 {
		avail := br.remainingBytes()
		if avail == 0 {
			return decoderResultNeedsMoreInput
		}
		n := min(avail, s.metaBlockRemainingLen)
		// Skip n bytes.
		for range n {
			if br.bitPos >= 8 {
				br.dropBits(8)
			} else {
				br.pos++
			}
		}
		s.metaBlockRemainingLen -= n
	}
	return decoderResultSuccess
}

// --- Ring buffer ---

func (s *decodeState) calculateRingBufferSize() {
	windowSize := 1 << s.windowBits
	newSize := windowSize
	if s.ringbufferSize == windowSize {
		return
	}
	if s.isMetadata {
		return
	}

	outputSize := s.pos + s.metaBlockRemainingLen
	minSize := s.ringbufferSize
	if minSize == 0 {
		minSize = 1024
	}
	if minSize < outputSize {
		minSize = outputSize
	}

	if s.cannyRingbufferAllocation {
		for (newSize >> 1) >= minSize {
			newSize >>= 1
		}
	}
	s.newRingbufferSize = newSize
}

func (s *decodeState) ensureRingBuffer() {
	if s.ringbufferSize == s.newRingbufferSize {
		return
	}
	newSize := s.newRingbufferSize + ringBufferWriteAheadSlack
	if cap(s.ringbuffer) >= newSize {
		s.ringbuffer = s.ringbuffer[:newSize]
	} else {
		newBuf := getDecRingBuf(newSize)
		if s.ringbuffer != nil {
			copy(newBuf, s.ringbuffer[:s.pos])
			putDecRingBuf(s.ringbuffer)
		}
		s.ringbuffer = newBuf
	}
	s.ringbuffer[s.newRingbufferSize-2] = 0
	s.ringbuffer[s.newRingbufferSize-1] = 0
	s.ringbufferSize = s.newRingbufferSize
	s.ringbufferMask = s.newRingbufferSize - 1
}

// --- Uncompressed block ---

func (s *decodeState) copyUncompressedBlock(output *[]byte) decoderResult {
	s.ensureRingBuffer()
	br := &s.br
	for {
		if s.substateUncompressed == uncompressedWrite {
			s.writeRingBuffer(output)
			s.substateUncompressed = uncompressedNone
			continue
		}
		if s.metaBlockRemainingLen == 0 {
			return decoderResultSuccess
		}
		avail := br.remainingBytes()
		if avail == 0 {
			return decoderResultNeedsMoreInput
		}
		n := min(avail, s.metaBlockRemainingLen)
		if s.pos+n > s.ringbufferSize {
			n = s.ringbufferSize - s.pos
		}
		br.copyBytes(s.ringbuffer[s.pos : s.pos+n])
		s.pos += n
		s.metaBlockRemainingLen -= n
		if s.pos >= s.ringbufferSize {
			s.substateUncompressed = uncompressedWrite
			return decoderResultNeedsMoreOutput
		}
	}
}

// --- Block type switching ---

// decodeBlockTypeAndLength reads a block type and new block length.
// Returns true if a block switch occurred (i.e., num_block_types > 1).
func (s *decodeState) decodeBlockTypeAndLength(safe bool, treeType int) bool {
	maxBlockType := s.numBlockTypes[treeType]
	if maxBlockType <= 1 {
		return false
	}

	typeTree := s.blockTypeTrees[treeType*huffmanMaxSize258:]
	lenTree := s.blockLenTrees[treeType*huffmanMaxSize26:]
	br := &s.br
	rb := &s.blockTypeRB

	var blockType uint
	if !safe {
		br.fillBitWindow(16)
		blockType = decodeSymbol(br.val, typeTree, br)
		s.blockLength[treeType] = readBlockLength(lenTree, br)
	} else {
		memento := br.saveState()
		var ok bool
		blockType, ok = safeReadSymbol(typeTree, br)
		if !ok {
			return false
		}
		bl, ok := safeReadBlockLength(s, lenTree)
		if !ok {
			s.substateReadBlockLength = readBlockLengthNone
			br.restoreState(memento)
			return false
		}
		s.blockLength[treeType] = bl
	}

	switch blockType {
	case 1:
		blockType = rb[treeType*2+1] + 1
	case 0:
		blockType = rb[treeType*2]
	default:
		blockType -= 2
	}
	if blockType >= maxBlockType {
		blockType -= maxBlockType
	}
	rb[treeType*2] = rb[treeType*2+1]
	rb[treeType*2+1] = blockType
	return true
}

// --- Trivial literal context detection ---

func (s *decodeState) detectTrivialLiteralBlockTypes() {
	s.trivialLiteralContexts = [8]uint32{}
	for i := range s.numBlockTypes[0] {
		offset := i << core.LiteralContextBits
		sample := s.contextMap[offset]
		trivial := true
		for j := range uint(1 << core.LiteralContextBits) {
			if s.contextMap[offset+j] != sample {
				trivial = false
				break
			}
		}
		if trivial {
			s.trivialLiteralContexts[i>>5] |= 1 << (i & 31)
		}
	}
}

func (s *decodeState) prepareLiteralDecoding() {
	blockType := s.blockTypeRB[1]
	contextOffset := blockType << core.LiteralContextBits
	s.contextMapSliceIdx = int(contextOffset)
	trivial := s.trivialLiteralContexts[blockType>>5]
	s.trivialLiteralContext = int((trivial >> (blockType & 31)) & 1)
	s.literalHTree = s.literalHGroup.codes[s.literalHGroup.htrees[s.contextMap[s.contextMapSliceIdx]]:]
	contextMode := s.contextModes[blockType] & 3
	s.contextLookup = core.ContextLookupTable[uint(contextMode)<<9:]
	if s.trivialLiteralContext == 0 {
		ctxMap := s.contextMap[s.contextMapSliceIdx:]
		htrees := s.literalHGroup.htrees
		codesBase := unsafe.Pointer(unsafe.SliceData(s.literalHGroup.codes))
		for ctx := range 64 {
			offset := htrees[ctxMap[ctx]]
			s.literalCodesOffsets[ctx] = offset
			s.literalCodesPtrs[ctx] = unsafe.Add(codesBase, uintptr(offset)*4)
		}
	}
}

// --- Compressed metablock header sub-functions ---

// --- Context modes ---

func (s *decodeState) readContextModes() decoderResult {
	br := &s.br
	i := s.loopCounter
	for uint(i) < s.numBlockTypes[0] {
		bits, ok := br.safeReadBits(2)
		if !ok {
			s.loopCounter = i
			return decoderResultNeedsMoreInput
		}
		s.contextModes[i] = byte(bits)
		i++
	}
	s.loopCounter = i
	return decoderResultSuccess
}

// --- Context map ---

func (s *decodeState) decodeContextMap(contextMapSize uint, contextMap *[]byte, numHTrees *uint) decoderResult {
	br := &s.br
	h := &s.headerArena

	for {
		switch h.substateContextMap {
		case contextMapNone:
			result := s.decodeVarLenUint8(numHTrees)
			if result != decoderResultSuccess {
				return result
			}
			(*numHTrees)++

			*contextMap = reuseBytes(*contextMap, int(contextMapSize))
			if *numHTrees <= 1 {
				// Trivial context map: all zeros.
				clear(*contextMap)
				return decoderResultSuccess
			}
			h.substateContextMap = contextMapReadPrefix
			fallthrough

		case contextMapReadPrefix:
			bits, ok := br.safeGetBits(5)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			if bits&1 != 0 {
				h.maxRunLengthPrefix = uint(bits>>1) + 1
				br.dropBits(5)
			} else {
				h.maxRunLengthPrefix = 0
				br.dropBits(1)
			}
			alphabetSize := *numHTrees + h.maxRunLengthPrefix
			h.substateHuffman = huffmanNone
			h.substateContextMap = contextMapHuffman
			// Save code for RLE decoding (h.code stores maxRunLengthPrefix).
			h.code = h.maxRunLengthPrefix
			_ = alphabetSize // used in next case
			fallthrough

		case contextMapHuffman:
			alphabetSize := *numHTrees + h.code
			result := s.readHuffmanCode(alphabetSize, alphabetSize,
				h.contextMapTable[:], nil)
			if result != decoderResultSuccess {
				return result
			}
			h.contextIndex = 0
			h.code = 0xFFFF // sentinel: no pending RLE
			h.substateContextMap = contextMapDecode
			fallthrough

		case contextMapDecode:
			code := h.code
			skipPreamble := code != 0xFFFF
			for h.contextIndex < contextMapSize || skipPreamble {
				if !skipPreamble {
					var ok bool
					code, ok = safeReadSymbol(h.contextMapTable[:], br)
					if !ok {
						h.code = 0xFFFF
						return decoderResultNeedsMoreInput
					}
					if code == 0 {
						(*contextMap)[h.contextIndex] = 0
						h.contextIndex++
						continue
					}
					if code > h.maxRunLengthPrefix {
						(*contextMap)[h.contextIndex] = byte(code - h.maxRunLengthPrefix)
						h.contextIndex++
						continue
					}
				} else {
					skipPreamble = false
				}
				// RLE: zero-fill.
				reps, ok := br.safeReadBits(code)
				if !ok {
					h.code = code
					return decoderResultNeedsMoreInput
				}
				reps += 1 << code
				if h.contextIndex+uint(reps) > contextMapSize {
					s.err = decompressError("context map repeat overflow")
					return decoderResultError
				}
				clear((*contextMap)[h.contextIndex : h.contextIndex+uint(reps)])
				h.contextIndex += uint(reps)
			}
			h.substateContextMap = contextMapTransform
			fallthrough

		case contextMapTransform:
			bits, ok := br.safeReadBits(1)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			if bits != 0 {
				s.inverseMoveToFrontTransform(*contextMap)
			}
			h.substateContextMap = contextMapNone
			return decoderResultSuccess

		default:
			s.err = decompressError("unreachable context map state")
			return decoderResultError
		}
	}
}

// inverseMoveToFrontTransform applies the inverse MTF transform in-place.
func (s *decodeState) inverseMoveToFrontTransform(v []byte) {
	var list [256]byte
	for i := range 256 {
		list[i] = byte(i)
	}

	upperBound := uint(0)
	for i := range v {
		idx := int(v[i])
		val := list[idx]
		upperBound |= uint(idx)
		v[i] = val
		copy(list[1:idx+1], list[:idx])
		list[0] = val
	}
	s.mtfUpperBound = upperBound >> 2
}

// --- Huffman tree group decode ---

func (s *decodeState) huffmanTreeGroupDecode(group *huffmanTreeGroup) decoderResult {
	h := &s.headerArena
	if h.substateTreeGroup != treeGroupLoop {
		h.htreeIndex = 0
		h.next = 0
		h.substateTreeGroup = treeGroupLoop
	}
	for h.htreeIndex < int(group.numHTrees) {
		var tableSize uint32
		result := s.readHuffmanCode(uint(group.alphabetSizeMax), uint(group.alphabetSizeLimit),
			group.codes[h.next:], &tableSize)
		if result != decoderResultSuccess {
			return result
		}
		group.htrees[h.htreeIndex] = h.next
		h.next += int(tableSize)
		h.htreeIndex++
	}
	h.substateTreeGroup = treeGroupNone
	return decoderResultSuccess
}

// --- Read Huffman code ---

func (s *decodeState) readHuffmanCode(alphabetSizeMax, alphabetSizeLimit uint, table []core.HuffmanCode, optTableSize *uint32) decoderResult {
	br := &s.br
	h := &s.headerArena

	for {
		switch h.substateHuffman {
		case huffmanNone:
			bits, ok := br.safeReadBits(2)
			if !ok {
				return decoderResultNeedsMoreInput
			}
			h.subLoopCounter = uint(bits)
			if bits != 1 {
				// Complex prefix code.
				h.space = 32
				h.repeat = 0
				h.codeLengthHisto = [16]uint16{}
				h.codeLengthCodeLengths = [core.AlphabetSizeCodeLengths]byte{}
				h.substateHuffman = huffmanComplex
				continue
			}
			fallthrough

		case huffmanSimpleSize:
			bits, ok := br.safeReadBits(2)
			if !ok {
				h.substateHuffman = huffmanSimpleSize
				return decoderResultNeedsMoreInput
			}
			h.symbol = uint(bits) // num_symbols (0..3 means 1..4)
			h.subLoopCounter = 0
			fallthrough

		case huffmanSimpleRead:
			result := s.readSimpleHuffmanSymbols(alphabetSizeMax, alphabetSizeLimit)
			if result != decoderResultSuccess {
				return result
			}
			fallthrough

		case huffmanSimpleBuild:
			if h.symbol == 3 {
				bits, ok := br.safeReadBits(1)
				if !ok {
					h.substateHuffman = huffmanSimpleBuild
					return decoderResultNeedsMoreInput
				}
				h.symbol += uint(bits)
			}
			tableSize := core.BuildSimpleHuffmanTable(table, huffmanTableBits,
				h.symbolListsArray[:4], uint32(h.symbol))
			if optTableSize != nil {
				*optTableSize = tableSize
			}
			h.substateHuffman = huffmanNone
			return decoderResultSuccess

		case huffmanComplex:
			result := s.readCodeLengthCodeLengths()
			if result != decoderResultSuccess {
				return result
			}
			core.BuildCodeLengthsHuffmanTable(h.table[:], h.codeLengthCodeLengths[:], h.codeLengthHisto[:])
			h.codeLengthHisto = [16]uint16{}

			// Initialize symbol lists for core.BuildHuffmanTable.
			h.symbolLists = h.symbolListsArray[:]
			for i := 0; i <= core.HuffmanMaxCodeLength; i++ {
				h.nextSymbol[i] = i - symListBase
				h.symbolListsArray[i] = 0xFFFF
			}

			h.symbol = 0
			h.prevCodeLen = core.InitialRepeatedCodeLength
			h.repeat = 0
			h.repeatCodeLen = 0
			h.space = 32768
			h.substateHuffman = huffmanLengthSymbols
			fallthrough

		case huffmanLengthSymbols:
			result := s.readSymbolCodeLengths(alphabetSizeLimit)
			if result != decoderResultSuccess {
				return result
			}
			if h.space != 0 {
				s.err = decompressError("huffman space not zero")
				return decoderResultError
			}
			sl := core.SymbolList{Storage: h.symbolLists, Offset: symListBase}
			tableSize := core.BuildHuffmanTable(table, huffmanTableBits, sl, h.codeLengthHisto[:])
			if optTableSize != nil {
				*optTableSize = tableSize
			}
			h.substateHuffman = huffmanNone
			return decoderResultSuccess

		default:
			s.err = decompressError("unreachable huffman state")
			return decoderResultError
		}
	}
}

// readSimpleHuffmanSymbols reads 1..4 symbol values for simple Huffman codes.
func (s *decodeState) readSimpleHuffmanSymbols(alphabetSizeMax, alphabetSizeLimit uint) decoderResult {
	br := &s.br
	h := &s.headerArena

	maxBits := uint(0)
	if alphabetSizeMax > 1 {
		v := alphabetSizeMax - 1
		for v > 0 {
			maxBits++
			v >>= 1
		}
	}

	i := h.subLoopCounter
	numSymbols := h.symbol
	for i <= numSymbols {
		v, ok := br.safeReadBits(maxBits)
		if !ok {
			h.subLoopCounter = i
			h.substateHuffman = huffmanSimpleRead
			return decoderResultNeedsMoreInput
		}
		if uint(v) >= alphabetSizeLimit {
			s.err = decompressError("simple huffman alphabet overflow")
			return decoderResultError
		}
		h.symbolListsArray[i] = uint16(v)
		i++
	}

	// Check for duplicates.
	for i := range numSymbols {
		for k := i + 1; k <= numSymbols; k++ {
			if h.symbolListsArray[i] == h.symbolListsArray[k] {
				s.err = decompressError("duplicate simple huffman symbols")
				return decoderResultError
			}
		}
	}

	return decoderResultSuccess
}

// readCodeLengthCodeLengths reads code length code lengths using the static
// prefix code (RFC 7932 §3.5). Can suspend and resume via h.subLoopCounter.
func (s *decodeState) readCodeLengthCodeLengths() decoderResult {
	br := &s.br
	h := &s.headerArena
	numCodes := h.repeat
	space := h.space
	i := h.subLoopCounter
	for i < core.AlphabetSizeCodeLengths {
		codeLenIdx := core.CodeLengthCodeOrder[i]
		ix, prefixAvailable := br.safeGetBits(4)
		if !prefixAvailable {
			avail := br.availBits()
			if avail != 0 {
				br.normalize()
				ix = br.bitsUnmasked() & 0xF
			} else {
				ix = 0
			}
			if uint(codeLengthPrefixLength[ix]) > avail {
				h.subLoopCounter = i
				h.repeat = numCodes
				h.space = space
				h.substateHuffman = huffmanComplex
				return decoderResultNeedsMoreInput
			}
		}
		v := codeLengthPrefixValue[ix]
		br.dropBits(uint(codeLengthPrefixLength[ix]))
		h.codeLengthCodeLengths[codeLenIdx] = v
		if v != 0 {
			space -= 32 >> v
			numCodes++
			h.codeLengthHisto[v]++
			if space-1 >= 32 {
				break
			}
		}
		i++
	}
	if numCodes != 1 && space != 0 {
		s.err = decompressError("invalid code length code space")
		return decoderResultError
	}
	return decoderResultSuccess
}

// processSingleCodeLength processes a single code length symbol.
func (h *metablockHeaderArena) processSingleCodeLength(codeLen uint) {
	h.repeat = 0
	if codeLen != 0 {
		h.symbolListsArray[h.nextSymbol[codeLen]+symListBase] = uint16(h.symbol)
		h.nextSymbol[codeLen] = int(h.symbol)
		h.prevCodeLen = codeLen
		h.space -= 32768 >> codeLen
		h.codeLengthHisto[codeLen]++
	}
	h.symbol++
}

// processRepeatedCodeLength processes a repeated code length (16 or 17).
func (h *metablockHeaderArena) processRepeatedCodeLength(codeLen, repeatDelta, alphabetSize uint) {
	extraBits := uint(3)
	newLen := uint(0)
	if codeLen == core.RepeatPreviousCodeLength {
		newLen = h.prevCodeLen
		extraBits = 2
	}
	if h.repeatCodeLen != newLen {
		h.repeat = 0
		h.repeatCodeLen = newLen
	}
	oldRepeat := h.repeat
	if h.repeat > 0 {
		h.repeat -= 2
		h.repeat <<= extraBits
	}
	h.repeat += repeatDelta + 3
	delta := h.repeat - oldRepeat
	if h.symbol+delta > alphabetSize {
		h.symbol = alphabetSize
		h.space = 0xFFFFF
		return
	}
	if h.repeatCodeLen != 0 {
		sym := h.symbol
		last := sym + delta
		next := h.nextSymbol[h.repeatCodeLen]
		for sym < last {
			h.symbolListsArray[next+symListBase] = uint16(sym)
			next = int(sym)
			sym++
		}
		h.symbol = last
		h.nextSymbol[h.repeatCodeLen] = next
		h.space -= delta << (15 - h.repeatCodeLen)
		h.codeLengthHisto[h.repeatCodeLen] += uint16(delta)
	} else {
		h.symbol += delta
	}
}

func (s *decodeState) readSymbolCodeLengths(alphabetSize uint) decoderResult {
	br := &s.br
	h := &s.headerArena

	if !br.warmup() {
		return decoderResultNeedsMoreInput
	}

	// Batched fast path: hoist bit reader state into locals and process
	// single code lengths in a tight loop without per-symbol method calls.
	if br.checkInputAmount() {
		val := br.val
		bitPos := br.bitPos
		brPos := br.pos
		inputBase := br.inputBase
		tableBase := unsafe.Pointer(&h.table[0])
		symbol := h.symbol

		for symbol < alphabetSize && h.space > 0 {
			// fillBitWindow inline
			if bitPos < 32 {
				if brPos > br.fastEnd {
					break
				}
				val |= uint64(le32(*(*uint32)(unsafe.Add(inputBase, brPos)))) << bitPos
				bitPos += 32
				brPos += 4
			}

			// Decode code length from table — the table has exactly 32 entries
			// (1 << core.HuffmanMaxCodeLengthCodeLength), so masking with 0x1F is
			// always in bounds.
			raw := *(*uint32)(unsafe.Add(tableBase, (val&0x1F)*4))
			drop := uint(raw & 0xFF)
			codeLen := uint(raw >> 16)

			bitPos -= drop
			val >>= drop & 63

			if codeLen >= core.RepeatPreviousCodeLength {
				// Handle repeat code: inline processRepeatedCodeLength
				// to avoid function call overhead and unnecessary br write-backs.
				extraBits := uint(2)
				newLen := uint(0)
				if codeLen == core.RepeatPreviousCodeLength {
					newLen = h.prevCodeLen
				} else {
					extraBits = 3
				}
				repeatDelta := uint(val & bitMask(extraBits))
				bitPos -= extraBits
				val >>= extraBits & 63

				if h.repeatCodeLen != newLen {
					h.repeat = 0
					h.repeatCodeLen = newLen
				}
				oldRepeat := h.repeat
				if h.repeat > 0 {
					h.repeat -= 2
					h.repeat <<= extraBits
				}
				h.repeat += repeatDelta + 3
				delta := h.repeat - oldRepeat
				switch {
				case symbol+delta > alphabetSize:
					symbol = alphabetSize
					h.space = 0xFFFFF
				case newLen != 0:
					last := symbol + delta
					next := h.nextSymbol[newLen]
					for symbol < last {
						h.symbolListsArray[next+symListBase] = uint16(symbol)
						next = int(symbol)
						symbol++
					}
					h.nextSymbol[newLen] = next
					h.space -= delta << (15 - newLen)
					h.codeLengthHisto[newLen] += uint16(delta)
				default:
					symbol += delta
				}

				if brPos > br.fastEnd {
					h.symbol = symbol
					br.val = val
					br.bitPos = bitPos
					br.pos = brPos
					goto slow
				}
				continue
			}

			// processSingleCodeLength inline
			h.repeat = 0
			if codeLen != 0 {
				h.symbolListsArray[h.nextSymbol[codeLen]+symListBase] = uint16(symbol)
				h.nextSymbol[codeLen] = int(symbol)
				h.prevCodeLen = codeLen
				h.space -= 32768 >> codeLen
				h.codeLengthHisto[codeLen]++
			}
			symbol++
		}

		h.symbol = symbol
		br.val = val
		br.bitPos = bitPos
		br.pos = brPos
	}

slow:
	for h.symbol < alphabetSize && h.space > 0 {
		p := h.table[:]
		if br.checkInputAmount() {
			// Fast path.
			br.fillBitWindow16()
			entry := p[br.bitsUnmasked()&bitMask(core.HuffmanMaxCodeLengthCodeLength)]
			br.dropBits(uint(entry.Bits))
			codeLen := uint(entry.Value)
			if codeLen < core.RepeatPreviousCodeLength {
				h.processSingleCodeLength(codeLen)
			} else {
				extraBits := uint(2)
				if codeLen != core.RepeatPreviousCodeLength {
					extraBits = 3
				}
				repeatDelta := uint(br.bitsUnmasked() & bitMask(extraBits))
				br.dropBits(extraBits)
				h.processRepeatedCodeLength(codeLen, repeatDelta, alphabetSize)
			}
		} else {
			// Safe path.
			availBits := br.availBits()
			var bits uint64
			if availBits != 0 {
				br.normalize()
				bits = br.bitsUnmasked()
			}
			entry := p[bits&bitMask(core.HuffmanMaxCodeLengthCodeLength)]
			if uint(entry.Bits) > availBits {
				if !br.pullByte() {
					return decoderResultNeedsMoreInput
				}
				continue
			}
			codeLen := uint(entry.Value)
			if codeLen < core.RepeatPreviousCodeLength {
				br.dropBits(uint(entry.Bits))
				h.processSingleCodeLength(codeLen)
			} else {
				extraBits := codeLen - 14
				repeatDelta := uint((bits >> uint(entry.Bits)) & bitMask(extraBits))
				if availBits < uint(entry.Bits)+extraBits {
					if !br.pullByte() {
						return decoderResultNeedsMoreInput
					}
					continue
				}
				br.dropBits(uint(entry.Bits) + extraBits)
				h.processRepeatedCodeLength(codeLen, repeatDelta, alphabetSize)
			}
		}
	}
	return decoderResultSuccess
}

// --- Distance LUT ---

func (s *decodeState) calculateDistanceLut() {
	b := &s.bodyArena
	npostfix := s.distancePostfixBits
	ndirect := s.numDirectDistanceCodes
	alphabetSizeLimit := uint(s.distanceHGroup.alphabetSizeLimit)

	// Skip recomputation when the parameters haven't changed.
	if b.cachedValid &&
		b.cachedPostfix == npostfix &&
		b.cachedDirect == ndirect &&
		b.cachedAlphabetSizeLim == alphabetSizeLimit {
		return
	}

	postfix := uint(1) << npostfix

	i := uint(core.NumDistanceShortCodes)

	// Fill direct codes.
	for j := range ndirect {
		b.distExtraBits[i] = 0
		b.distOffset[i] = j + 1
		i++
	}

	// Fill regular distance codes.
	bits := uint(1)
	half := uint(0)
	for i < alphabetSizeLimit {
		base := ndirect + ((((2 + half) << bits) - 4) << npostfix) + 1
		for j := range postfix {
			b.distExtraBits[i] = byte(bits)
			b.distOffset[i] = base + j
			i++
		}
		bits += half
		half ^= 1
	}

	b.cachedPostfix = npostfix
	b.cachedDirect = ndirect
	b.cachedAlphabetSizeLim = alphabetSizeLimit
	b.cachedValid = true
}

// copyOverlappingPattern fills n bytes at pos by repeating the dist-byte pattern
// that ends there, for dist < n.
//
// The first loop doubles the usable gap: after storing dist correct bytes the
// valid run behind p is twice as long, so once it reaches 16 every 16-byte load
// lies entirely inside already-final data and the second loop can run flat out.
// Both loops store 16 bytes at a time and so may write up to 15 bytes past
// pos+n, which the ring buffer's 542 bytes of slack absorb — the same contract
// the eager 16-byte store in processCommands already relies on.
func copyOverlappingPattern(base unsafe.Pointer, pos, dist, n int) {
	p, end := pos, pos+n
	for dist < 16 && p < end {
		*(*[16]byte)(unsafe.Add(base, p)) = *(*[16]byte)(unsafe.Add(base, p-dist))
		p += dist
		dist += dist
	}
	for p < end {
		*(*[16]byte)(unsafe.Add(base, p)) = *(*[16]byte)(unsafe.Add(base, p-dist))
		p += 16
	}
}

func (s *decodeState) processCommands() decoderResult {
	br := &s.br
	pos := s.pos
	i := s.loopCounter
	var cmdCode uint
	var ok bool
	var insertLenExtra uint64
	var literal uint
	var p1, p2 byte

	// Dispatch to the correct entry point based on state.
	switch s.state {
	case decoderStateCommandInner:
		goto commandInner
	case decoderStateCommandPostDecodeLiterals:
		goto commandPostDecodeLiterals
	case decoderStateCommandPostWrapCopy:
		goto commandPostWrapCopy
	}

commandBegin:
	if s.blockLength[1] == 0 && s.numBlockTypes[1] > 1 {
		if br.checkInputAmount() {
			s.decodeBlockTypeAndLength(false, 1)
		} else if !s.decodeBlockTypeAndLength(true, 1) {
			s.state = decoderStateCommandBegin
			s.pos = pos
			return decoderResultNeedsMoreInput
		}
		s.htreeCommand = s.insertCopyHGroup.codes[s.insertCopyHGroup.htrees[s.blockTypeRB[3]]:]
		s.htreeCommandLut = s.cmdLuts[s.blockTypeRB[3]<<huffmanTableBits:]
	}

	// Read the command. The fast path requires enough bits for a root index.
	// It computes the index before the refill, which removes a load dependency.
	if br.checkInputAmountAndBits(huffmanTableBits) {
		val, bitPos := br.val, br.bitPos
		idx := val & huffmanTableMask
		val |= le64(*(*uint64)(unsafe.Add(br.inputBase, br.pos))) << (bitPos & 63)
		br.pos += int((63 - bitPos) >> 3)
		bitPos |= 56
		// Decode the command inline to avoid a call and bounds checks.
		cmdTableBase := unsafe.Pointer(unsafe.SliceData(s.htreeCommand))
		raw := *(*uint32)(unsafe.Add(cmdTableBase, idx*4))
		lutSlot := (*uint64)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(s.htreeCommandLut)), idx*8))
		lut := *lutSlot
		cmdDrop := uint(raw & 0xFF)
		if cmdDrop > huffmanTableBits {
			nbits := cmdDrop - huffmanTableBits
			idx2 := idx + uint64(raw>>16) + ((val >> huffmanTableBits) & bitMask(nbits))
			raw = *(*uint32)(unsafe.Add(cmdTableBase, idx2*4))
			cmdDrop = huffmanTableBits + uint(raw&0xFF)
			lut = *(*uint64)(unsafe.Add(unsafe.Pointer(&cmdLutPacked[0]), uintptr(raw>>16)*8))
		} else if lut == 0 {
			lut = *(*uint64)(unsafe.Add(unsafe.Pointer(&cmdLutPacked[0]), uintptr(raw>>16)*8))
			*lutSlot = lut
		}

		s.distanceCode = int(int8(lut >> 16))
		s.distanceContext = int(uint8(lut >> 24))
		s.distCodesOffset = s.distCodesCache[s.distanceContext&3]
		i = int(uint16(lut >> 32))
		insertBits := uint(uint8(lut))
		copyBits := uint(uint8(lut >> 8))

		// Combined drop + readBits for insertLenExtra + copyLenExtra:
		// use local val/bitPos to avoid redundant stores/loads of br.val
		// and br.bitPos across intervening struct writes.
		val >>= cmdDrop & 63
		bitPos -= cmdDrop

		// The refill leaves at least 56 bits. The command code uses at most 15 bits.
		// The copy extra bits need another refill only after the insert extra bits.
		insertLenExtra = 0
		if insertBits != 0 {
			insertLenExtra = val & bitMask(insertBits)
			val >>= insertBits & 63
			bitPos -= insertBits
			val |= le64(*(*uint64)(unsafe.Add(br.inputBase, br.pos))) << (bitPos & 63)
			br.pos += int((63 - bitPos) >> 3)
			bitPos |= 56
		}

		copyExtra := val & bitMask(copyBits)
		val >>= copyBits & 63
		bitPos -= copyBits

		br.val = val
		br.bitPos = bitPos

		s.copyLength = int(lut>>48) + int(copyExtra)
	} else {
		cmdMemento := br.saveState()
		cmdCode, ok = safeReadSymbol(s.htreeCommand, br)
		if !ok {
			s.state = decoderStateCommandBegin
			s.pos = pos
			return decoderResultNeedsMoreInput
		}
		s.safeCmdMemento = cmdMemento

		v := core.CmdLut[cmdCode]
		s.distanceCode = int(v.DistanceCode)
		s.distanceContext = int(v.Context)
		s.distCodesOffset = s.distCodesCache[s.distanceContext]
		i = int(v.InsertLenOffset)

		insertLenExtra = 0
		if v.InsertLenExtraBits != 0 {
			insertLenExtra, ok = br.safeReadBits(uint(v.InsertLenExtraBits))
			if !ok {
				br.restoreState(s.safeCmdMemento)
				s.state = decoderStateCommandBegin
				s.pos = pos
				return decoderResultNeedsMoreInput
			}
		}
		var val uint64
		val, ok = br.safeReadBits(uint(v.CopyLenExtraBits))
		if !ok {
			br.restoreState(s.safeCmdMemento)
			s.state = decoderStateCommandBegin
			s.pos = pos
			return decoderResultNeedsMoreInput
		}
		s.copyLength = int(v.CopyLenOffset) + int(val)
	}
	s.blockLength[1]--
	i += int(insertLenExtra)

	s.metaBlockRemainingLen -= i
	if i == 0 {
		goto commandPostDecodeLiterals
	}

commandInner:
	// CommandInner: read i literal bytes.
	if s.trivialLiteralContext != 0 {
		for i > 0 {
			if s.blockLength[0] == 0 && s.numBlockTypes[0] > 1 {
				if br.checkInputAmount() {
					s.decodeBlockTypeAndLength(false, 0)
				} else if !s.decodeBlockTypeAndLength(true, 0) {
					s.state = decoderStateCommandInner
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
				s.prepareLiteralDecoding()
				if s.trivialLiteralContext == 0 {
					goto commandInner
				}
			}

			// Batch decode: when we have enough input, ringbuffer space,
			// and block length, decode multiple literals without per-symbol checks.
			n := min(i, int(s.blockLength[0]), s.ringbufferSize-pos, (br.availIn()-8)/2)
			if n == 1 {
				// Decode one symbol here. The call overhead of
				// decodeLiteralsBatch costs more than one symbol saves.
				// n == 1 implies availIn >= 10, so the 4-byte load in
				// fillBitWindow stays in bounds.
				br.fillBitWindow(16)
				s.ringbuffer[pos] = byte(decodeSymbol(br.val, s.literalHTree, br))
				pos++
				i--
				s.blockLength[0]--
				if pos == s.ringbufferSize {
					s.state = decoderStateCommandInnerWrite
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreOutput
				}
				continue
			} else if n > 1 {
				decodeLiteralsBatch(s.ringbuffer[pos:], n, s.literalHTree, br)
				pos += n
				i -= n
				s.blockLength[0] -= uint(n)
				if pos == s.ringbufferSize {
					s.state = decoderStateCommandInnerWrite
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreOutput
				}
				continue
			}

			// Slow path: not enough input for batch, decode one symbol at a time.
			if br.checkInputAmount() {
				br.fillBitWindow(16)
				literal = decodeSymbol(br.val, s.literalHTree, br)
			} else {
				literal, ok = safeReadSymbol(s.literalHTree, br)
				if !ok {
					s.state = decoderStateCommandInner
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
			}
			s.ringbuffer[pos] = byte(literal)
			s.blockLength[0]--
			pos++
			if pos == s.ringbufferSize {
				s.state = decoderStateCommandInnerWrite
				i--
				s.pos = pos
				s.loopCounter = i
				return decoderResultNeedsMoreOutput
			}
			i--
		}
	} else {
		p1 = s.ringbuffer[(pos-1)&s.ringbufferMask]
		p2 = s.ringbuffer[(pos-2)&s.ringbufferMask]
		for i > 0 {
			if s.blockLength[0] == 0 && s.numBlockTypes[0] > 1 {
				if br.checkInputAmount() {
					s.decodeBlockTypeAndLength(false, 0)
				} else if !s.decodeBlockTypeAndLength(true, 0) {
					s.state = decoderStateCommandInner
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
				s.prepareLiteralDecoding()
				if s.trivialLiteralContext != 0 {
					goto commandInner
				}
			}

			// Batch decode: when we have enough input, ringbuffer space,
			// and block length, decode multiple context-dependent literals
			// without per-symbol bounds checks.
			n := min(i, int(s.blockLength[0]), s.ringbufferSize-pos, (br.availIn()-8)/2)
			if n > 0 {
				p1, p2 = decodeLiteralsContextBatch(
					s.ringbuffer[pos:], n,
					&s.literalCodesPtrs,
					s.contextLookup, p1, p2, br,
				)
				pos += n
				i -= n
				s.blockLength[0] -= uint(n)
				if pos == s.ringbufferSize {
					s.state = decoderStateCommandInnerWrite
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreOutput
				}
				continue
			}

			ctx := s.contextLookup[p1] | s.contextLookup[256+int(p2)]
			hc := s.literalHGroup.codes[s.literalCodesOffsets[ctx]:]
			p2 = p1
			if br.checkInputAmount() {
				br.fillBitWindow(16)
				p1 = byte(decodeSymbol(br.val, hc, br))
			} else {
				literal, ok = safeReadSymbol(hc, br)
				if !ok {
					s.state = decoderStateCommandInner
					s.pos = pos
					s.loopCounter = i
					return decoderResultNeedsMoreInput
				}
				p1 = byte(literal)
			}
			s.ringbuffer[pos] = p1
			s.blockLength[0]--
			pos++
			if pos == s.ringbufferSize {
				s.state = decoderStateCommandInnerWrite
				i--
				s.pos = pos
				s.loopCounter = i
				return decoderResultNeedsMoreOutput
			}
			i--
		}
	}

	if s.metaBlockRemainingLen <= 0 {
		s.state = decoderStateMetablockDone
		s.pos = pos
		s.loopCounter = i
		return decoderResultSuccess
	}

commandPostDecodeLiterals:
	if s.distanceCode >= 0 {
		// Implicit distance.
		if s.distanceCode != 0 {
			s.distanceContext = 0
		} else {
			s.distanceContext = 1
		}
		s.distRBIdx--
		s.distanceCode = s.distRB[s.distRBIdx&3]
	} else {
		// Read distance code.
		if s.blockLength[2] == 0 && s.numBlockTypes[2] > 1 {
			if br.checkInputAmount() {
				s.decodeBlockTypeAndLength(false, 2)
			} else if !s.decodeBlockTypeAndLength(true, 2) {
				s.state = decoderStateCommandPostDecodeLiterals
				s.pos = pos
				s.loopCounter = i
				return decoderResultNeedsMoreInput
			}
			s.distContextMapSliceIdx = int(s.blockTypeRB[5]) << core.DistanceContextBits
			s.updateDistCodesCache()
			s.distCodesOffset = s.distCodesCache[s.distanceContext]
		}
		// Decode the distance inline because readDistance exceeds the compiler's inline budget.
		if br.checkInputAmountAndBits(huffmanTableBits) {
			val, bitPos := br.val, br.bitPos
			// Load the root entry before the refill.
			raw := distanceSymbolEntryFast(val, s.distanceHGroup.codes, s.distCodesOffset)
			val |= le64(*(*uint64)(unsafe.Add(br.inputBase, br.pos))) << (bitPos & 63)
			br.pos += int((63 - bitPos) >> 3)
			bitPos |= 56
			drop := uint(raw & 0xFF)
			code := uint(raw >> 16)
			if drop > huffmanTableBits {
				code, drop = decodeDistanceSymbolSecondLevel(val, s.distanceHGroup.codes, s.distCodesOffset, code, drop)
			}
			// Combined dropBits + readBits32: use locals to avoid
			// redundant stores/loads of br.val and br.bitPos.
			val >>= drop & 63
			bitPos -= drop
			s.blockLength[2]--
			s.distanceContext = 0
			if code&^0xF == 0 {
				br.val = val
				br.bitPos = bitPos
				s.distanceCode = int(code)
				s.takeDistanceFromRingBuffer()
			} else {
				b := &s.bodyArena
				nExtra := uint(*(*byte)(unsafe.Add(unsafe.Pointer(&b.distExtraBits[0]), uintptr(code))))
				offset := *(*uint)(unsafe.Add(unsafe.Pointer(&b.distOffset[0]), uintptr(code)*unsafe.Sizeof(uint(0))))
				bits := val & bitMask(nExtra)
				br.val = val >> (nExtra & 63)
				br.bitPos = bitPos - nExtra
				s.distanceCode = int(offset + uint(bits)<<s.distancePostfixBits)
			}
		} else {
			result := s.readDistance(br)
			if result != decoderResultSuccess {
				s.state = decoderStateCommandPostDecodeLiterals
				s.pos = pos
				s.loopCounter = i
				return result
			}
		}
	}

	if s.maxDistance != s.maxBackwardDistance {
		s.maxDistance = min(pos, s.maxBackwardDistance)
	}

	i = s.copyLength
	if s.distanceCode > s.maxDistance {
		if s.distanceCode > maxAllowedDistance {
			s.err = decompressError("invalid backward reference distance")
			return decoderResultError
		}

		compoundSize := s.compoundDictSize()

		switch {
		case s.distanceCode-s.maxDistance <= compoundSize:
			// Compound dictionary reference.
			address := compoundSize - (s.distanceCode - s.maxDistance)
			if !s.compoundDict.initCopy(s, address, i) {
				s.err = decompressError("invalid compound dictionary reference")
				return decoderResultError
			}
			pos += s.compoundDict.copyTo(s, pos)
			if pos >= s.ringbufferSize {
				s.state = decoderStateCommandPostWrite1
				s.pos = pos
				s.loopCounter = 0
				return decoderResultNeedsMoreOutput
			}

		case i >= core.DictMinWordLength && i <= core.DictMaxWordLength:
			// Static dictionary reference.
			address := s.distanceCode - s.maxDistance - 1 - compoundSize
			shift := uint(core.DictSizeBitsByLength[i])
			if shift == 0 {
				s.err = decompressError("invalid static dictionary reference")
				return decoderResultError
			}
			wordIdx := address & ((1 << shift) - 1)
			transformIdx := address >> shift
			if transformIdx >= core.NumTransforms {
				s.err = decompressError("invalid dictionary transform")
				return decoderResultError
			}
			offset := int(core.DictOffsetsByLength[i]) + wordIdx*i
			word := core.DictData[offset : offset+i]

			// Compensate double distance-ring-buffer roll.
			s.distRBIdx += s.distanceContext

			if transformIdx == int(core.TransformCutOffs[0]) {
				copy(s.ringbuffer[pos:], word)
			} else {
				i = core.TransformDictionaryWord(s.ringbuffer[pos:], word, transformIdx)
				// Omit transforms may produce an empty word; retain C's distance guard.
				if i == 0 && s.distanceCode <= 120 {
					s.err = decompressError("invalid dictionary transform")
					return decoderResultError
				}
			}
			pos += i
			s.metaBlockRemainingLen -= i
			if pos >= s.ringbufferSize {
				s.state = decoderStateCommandPostWrite1
				s.pos = pos
				s.loopCounter = 0
				return decoderResultNeedsMoreOutput
			}

		default:
			s.err = decompressError("invalid dictionary word length")
			return decoderResultError
		}

		if s.metaBlockRemainingLen <= 0 {
			s.state = decoderStateMetablockDone
			s.pos = pos
			s.loopCounter = 0
			return decoderResultSuccess
		}
		goto commandBegin
	}

	// Copy from ring buffer.
	s.distRB[s.distRBIdx&3] = s.distanceCode
	s.distRBIdx++
	s.metaBlockRemainingLen -= i

	{
		srcStart := (pos - s.distanceCode) & s.ringbufferMask
		dstEnd := pos + i
		srcEnd := srcStart + i

		// Eagerly copy 16 bytes using direct pointer arithmetic.
		// The ring buffer has 542 bytes of slack past ringbufferSize,
		// so pos+16 and srcStart+16 are always within bounds.
		base := unsafe.Pointer(unsafe.SliceData(s.ringbuffer))
		*(*[16]byte)(unsafe.Add(base, pos)) = *(*[16]byte)(unsafe.Add(base, srcStart))

		if dstEnd >= s.ringbufferSize || srcEnd >= s.ringbufferSize {
			// Wrapping — the byte loop is what yields to the state machine.
			goto commandPostWrapCopy
		}
		if srcEnd > pos && dstEnd > srcStart {
			// Overlapping but not wrapping. Only a source behind pos can be
			// pattern-replicated; when the ring-buffer subtraction wrapped,
			// srcStart is ahead of pos and the masked byte loop is still needed.
			if dist := pos - srcStart; dist > 0 {
				copyOverlappingPattern(base, pos, dist, i)
				pos += i
				goto postCopy
			}
			goto commandPostWrapCopy
		}

		pos += i
		if i > 16 {
			// Use fixed-size 16-byte copies for sizes up to 96 to avoid
			// the overhead of a memmove function call on moderate copies.
			src16 := srcStart + 16
			dst16 := pos - i + 16
			*(*[16]byte)(unsafe.Add(base, dst16)) = *(*[16]byte)(unsafe.Add(base, src16))
			if i > 32 {
				*(*[16]byte)(unsafe.Add(base, dst16+16)) = *(*[16]byte)(unsafe.Add(base, src16+16))
				if i > 48 {
					*(*[16]byte)(unsafe.Add(base, dst16+32)) = *(*[16]byte)(unsafe.Add(base, src16+32))
					if i > 64 {
						*(*[16]byte)(unsafe.Add(base, dst16+48)) = *(*[16]byte)(unsafe.Add(base, src16+48))
						if i > 80 {
							*(*[16]byte)(unsafe.Add(base, dst16+64)) = *(*[16]byte)(unsafe.Add(base, src16+64))
							if i > 96 {
								copy(s.ringbuffer[dst16+80:pos], s.ringbuffer[src16+80:srcEnd])
							}
						}
					}
				}
			}
		}

		goto postCopy
	}

commandPostWrapCopy:
	for i > 0 {
		srcStart := (pos - s.distanceCode) & s.ringbufferMask
		// Take the copy in chunks that stop at the ring-buffer end, so neither
		// cursor wraps inside one. ringbufferMask is ringbufferSize-1, so both
		// remainders are at least 1 and the loop always advances.
		n := min(i, s.ringbufferSize-pos, s.ringbufferSize-srcStart)
		if dist := pos - srcStart; dist > 0 {
			if dist >= n {
				copy(s.ringbuffer[pos:pos+n], s.ringbuffer[srcStart:srcStart+n])
			} else {
				// Stores up to 15 bytes past pos+n, which cannot exceed
				// ringbufferSize+15 and so stays inside the 542 bytes of slack.
				copyOverlappingPattern(unsafe.Pointer(unsafe.SliceData(s.ringbuffer)),
					pos, dist, n)
			}
			pos += n
			i -= n
		} else {
			// The ring-buffer subtraction wrapped, putting the source ahead of
			// pos; that has no contiguous run to copy, so step one byte.
			s.ringbuffer[pos] = s.ringbuffer[srcStart]
			pos++
			i--
		}
		if pos == s.ringbufferSize {
			s.state = decoderStateCommandPostWrite2
			s.pos = pos
			s.loopCounter = i
			return decoderResultNeedsMoreOutput
		}
	}

postCopy:

	if s.metaBlockRemainingLen <= 0 {
		s.state = decoderStateMetablockDone
		s.pos = pos
		s.loopCounter = i
		return decoderResultSuccess
	}
	goto commandBegin
}

// --- Distance reading ---

func (s *decodeState) readDistance(br *bitReader) decoderResult {
	b := &s.bodyArena
	distTree := s.distanceHGroup.codes[s.distCodesOffset:]

	if br.checkInputAmount() {
		br.fillBitWindow(16)
		code := decodeSymbol(br.val, distTree, br)
		s.blockLength[2]--
		s.distanceContext = 0
		if code&^0xF == 0 {
			s.distanceCode = int(code)
			s.takeDistanceFromRingBuffer()
			return decoderResultSuccess
		}
		// After the first checkInputAmount (availIn >= 28) and one
		// fillBitWindow (consuming at most 4 bytes), at least 24 bytes
		// remain — enough for readBits32's inner fillBitWindow.
		bits := br.readBits32(uint(b.distExtraBits[code]))
		s.distanceCode = int(b.distOffset[code] + uint(bits)<<s.distancePostfixBits)
		return decoderResultSuccess
	}

	// Slow path: not enough input for fast-path bit reading.
	memento := br.saveState()
	code, ok := safeReadSymbol(distTree, br)
	if !ok {
		return decoderResultNeedsMoreInput
	}
	s.blockLength[2]--

	s.distanceContext = 0
	if code&^0xF == 0 {
		s.distanceCode = int(code)
		s.takeDistanceFromRingBuffer()
		return decoderResultSuccess
	}

	var bits uint64
	if b.distExtraBits[code] != 0 {
		bits, ok = br.safeReadBits32(uint(b.distExtraBits[code]))
		if !ok {
			s.blockLength[2]++
			br.restoreState(memento)
			return decoderResultNeedsMoreInput
		}
	}
	s.distanceCode = int(b.distOffset[code] + uint(bits)<<s.distancePostfixBits)
	return decoderResultSuccess
}

func (s *decodeState) takeDistanceFromRingBuffer() {
	offset := s.distanceCode - 3
	if s.distanceCode <= 3 {
		s.distanceContext = 1 >> s.distanceCode
		s.distanceCode = s.distRB[(s.distRBIdx-offset)&3]
		s.distRBIdx -= s.distanceContext
	} else {
		indexDelta := 3
		base := s.distanceCode - 10
		if s.distanceCode < 10 {
			base = s.distanceCode - 4
		} else {
			indexDelta = 2
		}
		delta := ((0x605142 >> (4 * base)) & 0xF) - 3
		s.distanceCode = s.distRB[(s.distRBIdx+indexDelta)&0x3] + delta
		if s.distanceCode <= 0 {
			s.distanceCode = 0x7FFFFFFF
		}
	}
}

func (s *decodeState) excessiveInput() bool {
	s.br.unload()
	return s.br.availIn() > 0
}

// decompressError formats a decode-stage error.
func decompressError(msg string) error {
	return errors.New("brotli: " + msg)
}

// --- Huffman symbol reading ---

// decodeSymbol decodes a Huffman symbol from the bit reader using a two-level table.
// Requires at least 15 bits available.
func decodeSymbol(bits uint64, table []core.HuffmanCode, br *bitReader) uint {
	idx := bits & huffmanTableMask
	entry := table[idx]
	drop := uint(entry.Bits)
	if drop > huffmanTableBits {
		nbits := drop - huffmanTableBits
		entry = table[idx+uint64(entry.Value)+((bits>>huffmanTableBits)&bitMask(nbits))]
		// Second-level entry.Bits stores codeLen-rootBits; combine both drops
		// into a single shift to reduce overhead on the hot path.
		drop = huffmanTableBits + uint(entry.Bits)
	}
	br.dropBits(drop)
	return uint(entry.Value)
}

// distanceSymbolEntryFast loads the first-level distance Huffman table entry.
func distanceSymbolEntryFast(bits uint64, table []core.HuffmanCode, offset int) uint32 {
	tableBase := unsafe.Add(
		unsafe.Pointer(unsafe.SliceData(table)),
		uintptr(offset)*unsafe.Sizeof(core.HuffmanCode{}),
	)
	idx := bits & huffmanTableMask
	return *(*uint32)(unsafe.Add(tableBase, idx*4))
}

// decodeDistanceSymbolSecondLevel decodes a second-level distance Huffman entry.
func decodeDistanceSymbolSecondLevel(bits uint64, table []core.HuffmanCode, offset int, code, drop uint) (uint, uint) {
	tableBase := unsafe.Add(
		unsafe.Pointer(unsafe.SliceData(table)),
		uintptr(offset)*unsafe.Sizeof(core.HuffmanCode{}),
	)
	nbits := drop - huffmanTableBits
	idx := (bits & huffmanTableMask) + uint64(code) + ((bits >> huffmanTableBits) & bitMask(nbits))
	raw := *(*uint32)(unsafe.Add(tableBase, idx*4))
	return uint(raw >> 16), huffmanTableBits + uint(raw&0xFF)
}

// decodeLiteralsBatch decodes n literal symbols from br into dst using table.
// It hoists the bitReader state into local variables to keep them in registers
// and avoid repeated struct field access on the hot path.
// The fill is on demand: a symbol consumes at most core.HuffmanMaxCodeLength
// bits, so bits are only loaded once fewer than that remain. One fill leaves
// ≥56 bits, which covers about nine to fourteen symbols at typical code
// lengths, so the branch is taken rarely and predicts well.
func decodeLiteralsBatch(dst []byte, n int, table []core.HuffmanCode, br *bitReader) {
	val := br.val
	bitPos := br.bitPos
	brPos := br.pos
	inputBase := br.inputBase
	tableBase := unsafe.Pointer(unsafe.SliceData(table))

	for j := range n {
		if bitPos < core.HuffmanMaxCodeLength {
			val |= le64(*(*uint64)(unsafe.Add(inputBase, brPos))) << (bitPos & 63)
			brPos += int((63 - bitPos) >> 3)
			bitPos |= 56
		}

		idx := val & huffmanTableMask
		raw := *(*uint32)(unsafe.Add(tableBase, idx*4))
		drop := uint(raw & 0xFF)
		value := uint(raw >> 16)
		if drop > huffmanTableBits {
			nbits := drop - huffmanTableBits
			idx2 := idx + uint64(value) + ((val >> huffmanTableBits) & bitMask(nbits))
			raw = *(*uint32)(unsafe.Add(tableBase, idx2*4))
			drop = huffmanTableBits + uint(raw&0xFF)
			value = uint(raw >> 16)
		}
		bitPos -= drop
		val >>= drop & 63
		dst[j] = byte(value)
	}

	br.val = val
	br.bitPos = bitPos
	br.pos = brPos
}

// decodeLiteralsContextBatch decodes n literals using context-dependent Huffman
// trees. It hoists bitReader state into locals and uses unsafe pointer arithmetic
// to avoid per-symbol bounds checks. Returns updated (p1, p2) context bytes.
func decodeLiteralsContextBatch(
	dst []byte, n int,
	ptrs *[64]unsafe.Pointer,
	contextLookup []byte, p1, p2 byte,
	br *bitReader,
) (byte, byte) {
	val := br.val
	bitPos := br.bitPos
	brPos := br.pos
	inputBase := br.inputBase
	ctxBase := unsafe.Pointer(unsafe.SliceData(contextLookup))
	// Use an incrementing pointer instead of an index to avoid spilling the
	// loop counter onto the stack — on x86-64 the shift in fillBitWindow
	// requires CX, which would otherwise evict j every iteration.
	dstPtr := unsafe.Pointer(unsafe.SliceData(dst))
	ptrsBase := unsafe.Pointer(ptrs)
	c1, c2 := uint(p1), uint(p2)

	for n > 0 {
		n--
		if bitPos < core.HuffmanMaxCodeLength {
			val |= le64(*(*uint64)(unsafe.Add(inputBase, brPos))) << (bitPos & 63)
			brPos += int((63 - bitPos) >> 3)
			bitPos |= 56
		}

		ctx := uint(*(*byte)(unsafe.Add(ctxBase, c1))) | uint(*(*byte)(unsafe.Add(ctxBase, 256+c2)))
		tableBase := *(*unsafe.Pointer)(unsafe.Add(ptrsBase, uintptr(ctx)*unsafe.Sizeof(unsafe.Pointer(nil))))

		// decodeSymbol inline
		idx := val & huffmanTableMask
		raw := *(*uint32)(unsafe.Add(tableBase, idx*4))
		drop := uint(raw & 0xFF)
		value := uint(raw >> 16)
		if drop > huffmanTableBits {
			nbits := drop - huffmanTableBits
			idx2 := idx + uint64(value) + ((val >> huffmanTableBits) & bitMask(nbits))
			raw = *(*uint32)(unsafe.Add(tableBase, idx2*4))
			drop = huffmanTableBits + uint(raw&0xFF)
			value = uint(raw >> 16)
		}
		// dropBits inline
		bitPos -= drop
		val >>= drop & 63

		c2 = c1
		c1 = value
		*(*byte)(dstPtr) = byte(value)
		dstPtr = unsafe.Add(dstPtr, 1)
	}

	br.val = val
	br.bitPos = bitPos
	br.pos = brPos
	return byte(c1), byte(c2)
}

// safeDecodeSymbol is a fallback when fewer than 15 bits may be available.
func safeDecodeSymbol(table []core.HuffmanCode, br *bitReader) (uint, bool) {
	availBits := br.availBits()
	if availBits == 0 {
		if table[0].Bits == 0 {
			return uint(table[0].Value), true
		}
		return 0, false
	}
	br.normalize()
	val := br.bitsUnmasked()
	idx := val & huffmanTableMask
	entry := table[idx]
	if uint(entry.Bits) <= huffmanTableBits {
		if uint(entry.Bits) <= availBits {
			br.dropBits(uint(entry.Bits))
			return uint(entry.Value), true
		}
		return 0, false
	}
	if availBits <= huffmanTableBits {
		return 0, false
	}
	sub := (val & bitMask(uint(entry.Bits))) >> huffmanTableBits
	availBits -= huffmanTableBits
	entry2 := table[idx+uint64(entry.Value)+sub]
	if availBits < uint(entry2.Bits) {
		return 0, false
	}
	br.dropBits(huffmanTableBits + uint(entry2.Bits))
	return uint(entry2.Value), true
}

// safeReadSymbol tries the fast path (15 bits), falls back to safe decode.
func safeReadSymbol(table []core.HuffmanCode, br *bitReader) (uint, bool) {
	val, ok := br.safeGetBits(15)
	if ok {
		return decodeSymbol(val, table, br), true
	}
	return safeDecodeSymbol(table, br)
}

// --- Block length reading ---

func readBlockLength(table []core.HuffmanCode, br *bitReader) uint {
	br.fillBitWindow(16)
	code := decodeSymbol(br.val, table, br)
	nbits := core.BlockLengthNBits[code]
	return uint(core.BlockLengthOffset[code]) + uint(br.readBits(uint(nbits)))
}

func safeReadBlockLength(s *decodeState, table []core.HuffmanCode) (uint, bool) {
	br := &s.br
	var index uint
	if s.substateReadBlockLength == readBlockLengthNone {
		var ok bool
		index, ok = safeReadSymbol(table, br)
		if !ok {
			return 0, false
		}
	} else {
		index = s.blockLengthIndex
	}
	nbits := uint(core.BlockLengthNBits[index])
	bits, ok := br.safeReadBits(nbits)
	if !ok {
		s.blockLengthIndex = index
		s.substateReadBlockLength = readBlockLengthSuffix
		return 0, false
	}
	s.substateReadBlockLength = readBlockLengthNone
	return uint(core.BlockLengthOffset[index]) + uint(bits), true
}
