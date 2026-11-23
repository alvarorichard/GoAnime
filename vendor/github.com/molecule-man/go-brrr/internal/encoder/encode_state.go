// Shared encoder state for the streaming encoder (quality >= 2).

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// encodeState holds ring-buffer, bitstream, distance cache, and position
// tracking shared across all quality-specific encoders (Q2, Q3, Q4).
type encodeState struct {
	// Pointer-containing fields grouped first (reduces GC scan area).
	data         []byte          // ring buffer (for q>=2, points into ringBufAlloc at offset 2)
	commands     []command       // per-meta-block (reusable buffer)
	ringBufAlloc []byte          // [2-byte prefix | size | tailSize | 7 slack]
	outBuf       []byte          // reusable output scratch
	hashers      []*hasherCommon // cleared on 32-bit position wrap to force re-reset
	// litHisto/cmdHisto/distHisto point into the encoderArena's metablockArena
	// arrays when qualities 2 and 3 accumulate histograms inline. Non-nil from
	// the first createBackwardReferences call of a metablock until
	// writeMetaBlockFast / writeMetaBlockTrivial consumes them, allowing the
	// tally second-pass to be skipped entirely.
	litHisto  []uint32
	cmdHisto  []uint32
	distHisto []uint32
	b         bitWriter
	compound  compoundDictionary

	// distCache is the 4-entry ring buffer of recently used backward distances
	// (RFC 7932 Section 4). The format defines 16 "distance short codes" that
	// reference these entries with optional small offsets, allowing repeated
	// distances to be encoded compactly. Entry [0] is the most recent, [3] the
	// oldest. The spec mandates initial values {4, 11, 15, 16}.
	//
	// Updated by the encoder when a match uses a genuinely new distance (not
	// already in the cache): entries shift down and the new distance becomes [0].
	// Matches that hit the cache do not update it.
	distCache [4]uint

	// savedDistCache snapshots distCache before encoding a metablock. If the
	// encoder falls back to an uncompressed metablock, it restores from this
	// snapshot — uncompressed blocks do not modify the distance cache per the
	// spec, so the decoder's cache stays in sync.
	savedDistCache [4]uint

	lgblock               int
	sizeHint              uint   // expected total input size (for context modeling heuristics)
	dictNumLookups        uint   // adaptive heuristic counters (persist across metablocks)
	lastFlushPos          uint64 // stream position at last metablock flush
	inputPos              uint64 // total bytes written to ring buffer
	lastProcessedPos      uint64 // last position consumed by encodeData
	dictNumMatches        uint
	quality               int
	distAlphabetSizeMax   uint
	distAlphabetSizeLimit uint

	// Per-metablock counters (quality >= 2).
	numLiterals   uint
	numCommands   uint
	lastInsertLen uint

	lgwin       int
	mask        uint32 // ring-buffer mask
	ringBufSize uint32
	ringBufPos  uint32

	// distParams holds the current metablock distance parameters (Q10+).
	// Placed after uint32 fields for alignment (5 × uint32 = 20 bytes).
	distParams distanceParams

	lastBytes     uint16
	lastBytesBits uint8

	prevByte  byte
	prevByte2 byte

	// userSizeHint records whether the caller supplied a non-zero sizeHint
	// (vs. the auto-detected value set by updateSizeHint from the first
	// Write). Routing decisions that trust the hint at face value (e.g. the
	// h*u16 fast path) require this; the auto-detected value can mislead
	// when the first chunk is small but the stream is large.
	userSizeHint bool
}

// reset (re)initializes the encoder state for streaming compression (quality >= 2).
func (s *encodeState) reset(quality, lgwin int, sizeHint uint) {
	s.quality = quality
	s.lgwin = lgwin
	s.sizeHint = sizeHint
	s.userSizeHint = sizeHint > 0

	// ComputeLgBlock (quality.h):
	//   quality < 4    → lgblock = 14 (no block splitting)
	//   quality 4–8    → lgblock = 16
	//   quality >= 9   → lgblock = min(18, max(16, lgwin))
	switch {
	case quality < 4:
		s.lgblock = 14
	case quality < 9:
		s.lgblock = 16
	default:
		s.lgblock = min(18, max(16, lgwin))
	}

	// ComputeRbBits: 1 + max(lgwin, lgblock).
	rbBits := 1 + max(lgwin, s.lgblock)
	s.ringBufSize = 1 << rbBits
	s.mask = s.ringBufSize - 1
	s.ringBufPos = 0
	s.inputPos = 0
	s.lastProcessedPos = 0
	s.lastFlushPos = 0
	s.prevByte = 0
	s.prevByte2 = 0

	s.distCache = [4]uint{4, 11, 15, 16}
	s.savedDistCache = s.distCache

	// NDIRECT=0, NPOSTFIX=0, MAXNBITS=24 → distAlphabetSizeMax = 16 + 2*24 = 64.
	s.distParams = initDistanceParams(0, 0)
	s.distAlphabetSizeMax = uint(s.distParams.alphabetSizeMax)
	s.distAlphabetSizeLimit = uint(s.distParams.alphabetSizeLimit)

	// Stream header: EncodeWindowBits for the given lgwin.
	s.lastBytes, s.lastBytesBits = encodeWindowBits(lgwin)

	s.commands = s.commands[:0]
	s.numLiterals = 0
	s.numCommands = 0
	s.lastInsertLen = 0
	s.litHisto = nil
	s.cmdHisto = nil
	s.distHisto = nil

	s.dictNumLookups = 0
	s.dictNumMatches = 0

	s.compound = compoundDictionary{}

	// Ring buffer, command buffer, and output buffer are allocated lazily
	// on first use (copyInputToRingBuffer, prepareMetaBlock, initMetaBlockOutput).
	// This avoids multi-megabyte upfront allocations that dominate cost for
	// small inputs. On Reset the existing capacities are retained automatically.
	//
	// Exception: quality >= 10 uses Zopfli backward references whose BCE
	// hints require the full ring buffer to be addressable via the mask,
	// so pre-allocate for those quality levels.
	if quality >= 10 {
		tailSize := uint32(1) << s.lgblock
		fullBufLen := s.ringBufSize + tailSize
		needed := int(2 + fullBufLen + 7)
		if cap(s.ringBufAlloc) < needed {
			if v := ringBufPool.Get(); v != nil {
				if bp := v.(*[]byte); cap(*bp) >= needed {
					s.ringBufAlloc = (*bp)[:0]
				}
			}
		}
		if cap(s.ringBufAlloc) < needed {
			s.ringBufAlloc = make([]byte, 0, needed)
		}
	}
}

// extendLastCommand extends the most recent command's copy length by comparing
// ring-buffer bytes at the cached distance. Returns the updated remaining byte
// count and wrapped position after extension.
func (s *encodeState) extendLastCommand(length, wrappedPos uint32) (remainingLen, newPos uint32) {
	cmd := &s.commands[s.numCommands-1]
	data := s.data
	mask := s.mask
	maxBackwardDistance := (uint64(1) << s.lgwin) - core.WindowGap
	lastCopyLen := uint64(cmd.copyLen & 0x1FFFFFF)
	lastProcessedPos := s.lastProcessedPos - lastCopyLen
	maxDistance := min(lastProcessedPos, maxBackwardDistance)
	cmdDist := uint64(s.distCache[0])

	distanceCode := cmd.distanceCode(uint(s.distParams.numDirectCodes), uint(s.distParams.postfixBits))

	if distanceCode < core.NumDistanceShortCodes ||
		uint64(distanceCode-(core.NumDistanceShortCodes-1)) == cmdDist {
		if cmdDist <= maxDistance {
			for length != 0 {
				dst := wrappedPos & mask
				src := (wrappedPos - uint32(cmdDist)) & mask
				// Chunk to the ring-buffer end so neither cursor wraps inside a
				// compare, which is what lets matchLenAt read eight bytes at a
				// time instead of one.
				n := min(length, mask+1-dst, mask+1-src)
				var m uint32
				if n >= matchLenLongBlock {
					m = uint32(matchLenAtLong(data, uint(src), uint(dst), int(n)))
				} else {
					m = uint32(matchLenAt(data, uint(src), uint(dst), int(n)))
				}
				cmd.copyLen += m
				length -= m
				wrappedPos += m
				if m != n {
					break
				}
			}
		}
		// Compound dictionary extension: extend copy into dictionary data
		// when the distance points beyond the ring buffer.
		if cd := &s.compound; cd.numChunks > 0 &&
			(cmdDist-maxDistance-1) < uint64(cd.totalSize) &&
			lastCopyLen < cmdDist-maxDistance {
			address := uint64(cd.totalSize) - (cmdDist - maxDistance) + lastCopyLen
			brIndex := 0
			for address >= uint64(cd.chunkOffsets[brIndex+1]) {
				brIndex++
			}
			brOffset := uint(address) - cd.chunkOffsets[brIndex]
			chunk := cd.chunkSource[brIndex]
			chunkLen := cd.chunkOffsets[brIndex+1] - cd.chunkOffsets[brIndex]
			for length != 0 && data[wrappedPos&mask] == chunk[brOffset] {
				cmd.copyLen++
				length--
				wrappedPos++
				brOffset++
				if brOffset == chunkLen {
					brIndex++
					brOffset = 0
					if brIndex != cd.numChunks {
						chunk = cd.chunkSource[brIndex]
						chunkLen = cd.chunkOffsets[brIndex+1] - cd.chunkOffsets[brIndex]
					} else {
						break
					}
				}
			}
		}

		// Recalculate cmdPrefix after extending the copy length.
		copyLenCode := cmd.copyLenCode()
		insCode := getInsertLenCode(uint(cmd.insertLen))
		copyCode := getCopyLenCode(uint(copyLenCode))
		cmd.cmdPrefix = combineLengthCodes(insCode, copyCode, cmd.distPrefixCode() == 0)
	}
	return length, wrappedPos
}

// shouldCompress estimates whether the given block will benefit from
// compression. It returns false for very short blocks or blocks whose
// sampled literal entropy is close to 8 bits per byte (incompressible).
func (s *encodeState) shouldCompress(length, numLiterals, numCommands int) bool {
	if length <= 2 {
		return false
	}
	if numCommands < (length>>8)+2 && float64(numLiterals) > 0.99*float64(length) {
		var literalHisto [256]uint32
		const sampleRate = 13
		const invSampleRate = 1.0 / 13.0
		const minEntropy = 7.92
		bitCostThreshold := float64(length) * minEntropy * invSampleRate
		t := (length + sampleRate - 1) / sampleRate
		mask := s.mask
		pos := uint32(s.lastFlushPos)
		for range t {
			literalHisto[s.data[pos&mask]]++
			pos += sampleRate
		}
		if bitsEntropy(literalHisto[:]) > bitCostThreshold {
			return false
		}
	}
	return true
}

// writeUncompressedMetaBlock writes an uncompressed meta-block from a ring
// buffer. Unlike bitWriter.writeUncompressedMetaBlock (which takes a linear
// slice), this handles ring-buffer wrapping and the isLast flag.
func (s *encodeState) writeUncompressedMetaBlock(length int, isLast bool) {
	b := &s.b
	data := s.data
	mask := uint(s.mask)

	// Header: ISLAST=0, MLEN, ISUNCOMPRESSED=1.
	b.writeMetaBlockHeader(length, false, true)
	b.byteAlign()

	maskedPos := wrapPosition(s.lastFlushPos) & mask
	if maskedPos+uint(length) > mask+1 {
		len1 := mask + 1 - maskedPos
		b.writeBytes(data[maskedPos : maskedPos+len1])
		length -= int(len1)
		maskedPos = 0
	}
	b.writeBytes(data[maskedPos : maskedPos+uint(length)])

	// An uncompressed block cannot itself be the last block, so append an
	// empty final block if this is the end of the stream.
	if isLast {
		b.writeBits(1, 1) // ISLAST
		b.writeBits(1, 1) // ISEMPTY
		b.byteAlign()
	}
}

// updateSizeHint auto-estimates the total input size when no explicit hint
// was provided. Called once on the first Write; a no-op when sizeHint is
// already set.
func (s *encodeState) updateSizeHint(availableIn uint) {
	if s.sizeHint != 0 {
		return
	}
	total := s.unprocessedInputSize() + uint64(availableIn)
	const limit = 1 << 30
	if total > limit {
		total = limit
	}
	s.sizeHint = uint(total)
}

func (s *encodeState) inputBlockSize() uint {
	return 1 << s.lgblock
}

func (s *encodeState) unprocessedInputSize() uint64 {
	return s.inputPos - s.lastProcessedPos
}

func (s *encodeState) remainingInputBlockSize() uint {
	delta := s.unprocessedInputSize()
	blockSize := s.inputBlockSize()
	if delta >= uint64(blockSize) {
		return 0
	}
	return blockSize - uint(delta)
}

// updateLastProcessedPos marks all input as processed. When the 32-bit
// wrapped position rolls over, it resets all registered hashers so they
// are re-prepared on the next block.
func (s *encodeState) updateLastProcessedPos() {
	wrappedLast := uint32(wrapPosition(s.lastProcessedPos))
	wrappedCur := uint32(wrapPosition(s.inputPos))
	s.lastProcessedPos = s.inputPos
	if wrappedCur < wrappedLast {
		for _, h := range s.hashers {
			h.ready = false
		}
	}
}

func (s *encodeState) maxMetablockSize() uint {
	n := min(1+max(s.lgwin, s.lgblock), 24)
	return 1 << n
}

func (s *encodeState) trailingBits() (uint16, uint8) {
	return s.lastBytes, s.lastBytesBits
}

func (s *encodeState) clearTrailingBits() {
	s.lastBytes = 0
	s.lastBytesBits = 0
}

func (s *encodeState) attachDictionary(pd *PreparedDictionary) error {
	return s.compound.attach(pd)
}
