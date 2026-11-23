package encoder

// One-pass fast encoder for an input fragment, independent of input history.
// When a backward match is found, the corresponding command and literal codes
// are immediately emitted to the bitstream.

import (
	"math/bits"
	"unsafe"

	"github.com/molecule-man/go-brrr/internal/core"
)

//go:generate go run ../../cmd/genfixedshift -in compress_fragment_fast.go -out compress_fragment_fast_table15.go -func writeCommandsAnyTable -name writeCommandsTable15 -bits 15

// maxDistance is the maximum backward reference distance for window size 18.
// BROTLI_MAX_BACKWARD_LIMIT(18) = (1 << 18) - 16 = 262128.
const maxDistance = (1 << 18) - 16

// hashMul32 is the hash multiplier used for 5-byte match lookups.
const hashMul32 = 0x1E35A7BD

const (
	firstBlockSize   = 3 << 15
	mergeBlockSize   = 1 << 16
	inputMarginBytes = core.WindowGap
	minMatchLen      = 5
)

// minRatio is the acceptable loss for uncompressible speedup (2%).
const minRatio = 980

// cmdHistoSeed is the initial histogram seed for commands. Each compression
// block starts with these counts to provide a non-zero baseline.
var cmdHistoSeed = [128]uint32{
	0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 0, 0, 0,
}

// copyLenCodeInfo packs the per-copyLen encoding for the inner-repeat-match
// fast emit path (copyLen ∈ [5, 133]). Replaces the branch + bits.Len + extra
// shifts in the hot loop with a single table load, freeing the CL register
// from extra shift-count contention in the encoding block.
//
// Layout: bits 0..7 = ccode (in [19, 33]); bits 8..15 = nbits (in [0, 5]);
// bits 16..31 = cextra (in [0, 31]).
var copyLenCodeInfo = func() [134]uint32 {
	var t [134]uint32
	for copyLen := 5; copyLen < 134; copyLen++ {
		var ccode, nbits, cextra uint32
		if copyLen < 10 {
			ccode = uint32(copyLen) + 14
		} else {
			tail := uint32(copyLen - 6)
			nbits = uint32(bits.Len32(tail)) - 2
			prefix := tail >> nbits
			ccode = (nbits << 1) + prefix + 20
			cextra = tail - (prefix << nbits)
		}
		t[copyLen] = ccode | (nbits << 8) | (cextra << 16)
	}
	return t
}()

type fragmentCompressor struct {
	arena          *onePassArena
	b              *bitWriter
	input          []byte
	table          []uint32
	tableBits      uint
	blockSize      int
	totalBlockSize int
	pos            int
	shift          uint
	mlenPos        uint
	nextEmit       int
	ipEnd          int
	literalRatio   uint
	metablockStart int
	isLast         bool
}

func (c *fragmentCompressor) compress() {
	c.tableBits = uint(bits.Len(uint(len(c.table)))) - 1
	initialBitOffset := c.b.bitOffset

	if len(c.input) == 0 {
		assert(c.isLast)
		c.b.writeBits(1, 1) // islast
		c.b.writeBits(1, 1) // isempty
		c.b.byteAlign()
		return
	}

	c.blockSize = min(len(c.input), firstBlockSize)
	c.totalBlockSize = c.blockSize
	c.shift = 64 - c.tableBits
	// Save the bit position of the MLEN field of the meta-block header, so that
	// we can update it later if we decide to extend this meta-block.
	c.mlenPos = c.b.bitOffset + 3

	c.b.writeMetaBlockHeader(c.blockSize, false, false)
	// No block splits, no contexts.
	c.b.writeBits(13, 0)

	c.literalRatio = c.arena.buildAndWriteLiteralPrefixCode(c.input[:c.blockSize], c.b)

	// Store the pre-compressed command and distance prefix codes.
	{
		i := uint(0)
		for i+7 < c.arena.cmdCodeNumBits {
			c.b.writeBits(8, uint64(c.arena.cmdCode[i/8]))
			i += 8
		}
		c.b.writeBits(c.arena.cmdCodeNumBits&7, uint64(c.arena.cmdCode[c.arena.cmdCodeNumBits/8]))
	}

	c.writeCommands()

	// If output is larger than a single uncompressed block, rewrite it.
	if c.b.bitOffset-initialBitOffset > 31+uint(len(c.input))*8 {
		c.writeUncompressedMetaBlock(c.input, initialBitOffset)
	}

	if c.isLast {
		c.b.writeBits(1, 1) // islast
		c.b.writeBits(1, 1) // isempty
		c.b.byteAlign()
	}
}

func (c *fragmentCompressor) writeCommands() {
	if c.shift == 64-15 {
		c.writeCommandsTable15()
		return
	}
	c.writeCommandsAnyTable()
}

// The generator uses writeCommandsAnyTable as the source for writeCommandsTable15.
func (c *fragmentCompressor) writeCommandsAnyTable() {
	// Initialize the command and distance histograms.
	c.arena.cmdHisto = cmdHistoSeed
	input := c.input
	tbl := unsafe.Pointer(unsafe.SliceData(c.table))
	ip := c.pos
	shift := c.shift
	lastDistance := -1
	c.ipEnd = c.pos + c.blockSize

	if c.blockSize < inputMarginBytes {
		c.writeRemainder()
		return
	}

	// For the last block, keep a 16-byte margin so all distances are at
	// most window size - 16. For other blocks, keep a 5-byte margin.
	lenLimit := min(c.blockSize-minMatchLen, len(input)-c.pos-inputMarginBytes)
	ipLimit := c.pos + lenLimit

	ip++
	nextLoad := loadU64LE(input, uint(ip))
	nextHash := hashBytesAtOffset(nextLoad, 0, shift)

	for {
		// Step 1: Scan forward in the input looking for a 5-byte-long match.
		// If we find one whose distance is too large, keep scanning.
		skip := uint32(32)
		nextIP := ip
		var candidate int

		for {
			slot := (*uint32)(unsafe.Add(tbl, uintptr(nextHash)*4))
			ipBytes := nextLoad
			bytesBetweenHashLookups := skip >> 5
			skip++
			ip = nextIP
			nextIP = ip + int(bytesBetweenHashLookups)
			if nextIP > ipLimit {
				c.writeRemainder()
				return
			}
			nextLoad = loadU64LE(input, uint(nextIP))
			nextHash = hashBytesAtOffset(nextLoad, 0, shift)

			candidate = ip - lastDistance
			// uint(candidate) < uint(ip) folds the two-step `candidate >= 0 &&
			// candidate < ip` validity check into a single unsigned compare:
			// when candidate is negative (initial state) its unsigned form is
			// larger than uint(ip), so the compare fails.
			if uint(candidate) < uint(ip) && (loadU64LE(input, uint(candidate))^ipBytes)<<24 == 0 {
				*slot = uint32(ip)
				if ip-candidate <= maxDistance {
					break
				}
				continue
			}

			candidate = int(*slot)
			*slot = uint32(ip)
			if (loadU64LE(input, uint(candidate))^ipBytes)<<24 == 0 {
				if ip-candidate <= maxDistance {
					break
				}
				continue
			}
		}

		// Step 2: Emit the found match together with the literal bytes from
		// nextEmit to the bitstream, and then see if we can find a next match
		// immediately afterwards.
		{
			base := ip
			matched := 5 + matchLenAt(
				input, uint(candidate+5), uint(ip+5), c.ipEnd-ip-5)
			distance := base - candidate
			insert := uint(base - c.nextEmit)
			ip += matched

			switch {
			case insert < 6:
				code := insert + 40
				c.b.writeBits(uint(c.arena.cmdDepth[code]), uint64(c.arena.cmdBits[code]))
				c.arena.cmdHisto[code]++
			case insert < 6210:
				c.writeInsertLen(insert)
			case c.shouldUseUncompressedMode(insert):
				c.writeUncompressedMetaBlock(
					input[c.metablockStart:base], c.mlenPos-3)
				c.pos = base
				c.nextEmit = c.pos
				c.nextBlock()
				return
			default:
				c.writeLongInsertLen(insert)
			}
			c.writeLiterals(input[c.nextEmit : c.nextEmit+int(insert)])
			if distance == lastDistance {
				c.b.writeBits(uint(c.arena.cmdDepth[64]), uint64(c.arena.cmdBits[64]))
				c.arena.cmdHisto[64]++
				c.writeCopyLenLastDistance(uint(matched))
			} else {
				c.writeDistanceAndCopyLenLastDistance(uint(distance), uint(matched))
				lastDistance = distance
			}

			c.nextEmit = ip
			if ip >= ipLimit {
				c.writeRemainder()
				return
			}
		}

		// Try to find another match immediately. writeCopyAndDistance is
		// inlined here so the inner-repeat-match loop's hottest call site
		// avoids per-iteration prologue/epilogue overhead. The function is
		// otherwise too large to inline automatically (cost ~400 vs the
		// 80-instruction budget) yet it's the only caller.
		for {
			inputBytes := loadU64LE(input, uint(ip-3))
			prevHash := hashBytesAtOffset(inputBytes, 0, shift)
			curHash := hashBytesAtOffset(inputBytes, 3, shift)
			*(*uint32)(unsafe.Add(tbl, uintptr(prevHash)*4)) = uint32(ip - 3)
			prevHash = hashBytesAtOffset(inputBytes, 1, shift)
			*(*uint32)(unsafe.Add(tbl, uintptr(prevHash)*4)) = uint32(ip - 2)
			prevHash = hashBytesAtOffset(inputBytes, 2, shift)
			*(*uint32)(unsafe.Add(tbl, uintptr(prevHash)*4)) = uint32(ip - 1)
			curPtr := (*uint32)(unsafe.Add(tbl, uintptr(curHash)*4))
			candidate = int(*curPtr)
			*curPtr = uint32(ip)
			if !isMatch(input, uint(ip), uint(candidate)) {
				break
			}
			base := ip
			matched := 5 + matchLenAt(
				input, uint(candidate+5), uint(ip+5), c.ipEnd-ip-5)
			if ip-candidate > maxDistance {
				break
			}
			ip += matched
			lastDistance = base - candidate
			distance := uint(lastDistance)
			copyLen := uint(matched)
			if copyLen >= 134 {
				c.writeCopyLen(copyLen)
				c.writeDistance(distance)
			} else {
				depth := c.arena.cmdDepth[:]
				cmdBits := c.arena.cmdBits[:]
				histo := c.arena.cmdHisto[:]

				// Distance: 2*(nbits-1)+prefix+80 ∈ [80,111]; total ≤ 31 bits.
				dd := distance + 3
				dNbits := (uint(bits.Len(dd)) - 2) & 63
				dPrefix := (dd >> dNbits) & 1
				dOffset := (2 + dPrefix) << dNbits
				distcode := 2*(dNbits-1) + dPrefix + 80
				dDepth := uint(depth[distcode]) & 63
				dBits := uint64(cmdBits[distcode]) | uint64(dd-dOffset)<<dDepth
				dLen := dDepth + dNbits

				// Copy length: lookup precomputed (ccode, nbits, cextra) so the
				// hot path is one table load + one depth/bits load instead of
				// the bits.Len + shift sequence (frees CL across the block).
				info := copyLenCodeInfo[copyLen]
				ccode := uint(info & 0xFF)
				nbits := uint((info >> 8) & 0xFF)
				cextra := uint(info >> 16)
				d := uint(depth[ccode]) & 63
				cBits := uint64(cmdBits[ccode]) | uint64(cextra)<<d
				cLen := d + nbits

				// Combined ≤ 51 bits, safely under the 56-bit writeBits limit.
				c.b.writeBits(cLen+dLen, cBits|dBits<<(cLen&63))
				histo[ccode]++
				histo[distcode]++
			}

			c.nextEmit = ip
			if ip >= ipLimit {
				c.writeRemainder()
				return
			}
		}

		ip++
		nextLoad = loadU64LE(input, uint(ip))
		nextHash = hashBytesAtOffset(nextLoad, 0, shift)
	}
}

func (c *fragmentCompressor) writeRemainder() {
	input := c.input
	c.pos += c.blockSize
	c.blockSize = min(len(input)-c.pos, mergeBlockSize)

	// Decide if we want to continue this meta-block instead of emitting the
	// last insert-only command.
	if c.pos < len(input) &&
		c.totalBlockSize+c.blockSize <= (1<<20) &&
		c.arena.shouldMergeBlock(input[c.pos:], c.blockSize, c.arena.litDepth[:]) {
		// Update the size of the current meta-block and continue emitting
		// commands. We can do this because the current size and the new size
		// both have 5 nibbles.
		c.totalBlockSize += c.blockSize
		updateBits(c.b.buf, 20, uint32(c.totalBlockSize-1), c.mlenPos)
		c.writeCommands()
		return
	}

	// Emit the remaining bytes as literals.
	if c.nextEmit < c.ipEnd {
		insert := uint(c.ipEnd - c.nextEmit)
		switch {
		case insert < 6210:
			c.writeInsertLen(insert)
			c.writeLiterals(input[c.nextEmit:c.ipEnd])
		case c.shouldUseUncompressedMode(insert):
			c.writeUncompressedMetaBlock(
				input[c.metablockStart:c.ipEnd], c.mlenPos-3)
		default:
			c.writeLongInsertLen(insert)
			c.writeLiterals(input[c.nextEmit:c.ipEnd])
		}
	}
	c.nextEmit = c.ipEnd
	c.nextBlock()
}

func (c *fragmentCompressor) nextBlock() {
	input := c.input
	// If we have more data, write a new meta-block header and prefix codes
	// and then continue emitting commands.
	if c.pos < len(input) {
		c.metablockStart = c.pos
		c.blockSize = min(len(input)-c.pos, firstBlockSize)
		c.totalBlockSize = c.blockSize
		c.mlenPos = c.b.bitOffset + 3
		c.b.writeMetaBlockHeader(c.blockSize, false, false)
		// No block splits, no contexts.
		c.b.writeBits(13, 0)
		c.literalRatio = c.arena.buildAndWriteLiteralPrefixCode(
			input[c.pos:c.pos+c.blockSize], c.b)
		c.arena.buildAndWriteCommandPrefixCode(c.b)
		c.writeCommands()
		return
	}

	if !c.isLast {
		// If this is not the last block, update the command and distance
		// prefix codes for the next block and store the compressed forms.
		c.arena.cmdCode[0] = 0
		c.arena.cmdCodeNumBits = 0
		cmdCodeStream := &bitWriter{buf: c.arena.cmdCode[:], bitOffset: 0}
		c.arena.buildAndWriteCommandPrefixCode(cmdCodeStream)
		c.arena.cmdCodeNumBits = cmdCodeStream.bitOffset
	}
}

// writeInsertLen writes an insert length code to the bitstream.
// Requires: insertLen < 6210
func (c *fragmentCompressor) writeInsertLen(insertLen uint) {
	depth := c.arena.cmdDepth[:]
	cmdBits := c.arena.cmdBits[:]
	histo := c.arena.cmdHisto[:]
	switch {
	case insertLen < 6:
		code := insertLen + 40
		c.b.writeBits(uint(depth[code]), uint64(cmdBits[code]))
		histo[code]++
	case insertLen < 130:
		tail := insertLen - 2
		nbits := uint(bits.Len(tail)) - 2
		prefix := tail >> nbits
		inscode := (nbits << 1) + prefix + 42
		d := uint(depth[inscode])
		// Pack Huffman code and extra bits into one writeBits.
		// Max total = 15 + 5 = 20 bits.
		c.b.writeBits(d+nbits, uint64(cmdBits[inscode])|uint64(tail-(prefix<<nbits))<<d)
		histo[inscode]++
	case insertLen < 2114:
		tail := insertLen - 66
		nbits := uint(bits.Len(tail)) - 1
		code := nbits + 50
		d := uint(depth[code])
		// Max total = 15 + 11 = 26 bits.
		c.b.writeBits(d+nbits, uint64(cmdBits[code])|uint64(tail-(1<<nbits))<<d)
		histo[code]++
	default:
		d := uint(depth[61])
		// Max total = 15 + 12 = 27 bits.
		c.b.writeBits(d+12, uint64(cmdBits[61])|uint64(insertLen-2114)<<d)
		histo[61]++
	}
}

// writeLongInsertLen writes a long insert length code to the bitstream.
func (c *fragmentCompressor) writeLongInsertLen(insertLen uint) {
	depth := c.arena.cmdDepth[:]
	cmdBits := c.arena.cmdBits[:]
	histo := c.arena.cmdHisto[:]
	if insertLen < 22594 {
		d := uint(depth[62])
		// Max total = 15 + 14 = 29 bits.
		c.b.writeBits(d+14, uint64(cmdBits[62])|uint64(insertLen-6210)<<d)
		histo[62]++
	} else {
		d := uint(depth[63])
		// Max total = 15 + 24 = 39 bits.
		c.b.writeBits(d+24, uint64(cmdBits[63])|uint64(insertLen-22594)<<d)
		histo[63]++
	}
}

// writeCopyLen writes a copy length code to the bitstream.
func (c *fragmentCompressor) writeCopyLen(copyLen uint) {
	depth := c.arena.cmdDepth[:]
	cmdBits := c.arena.cmdBits[:]
	histo := c.arena.cmdHisto[:]
	switch {
	case copyLen < 10:
		c.b.writeBits(uint(depth[copyLen+14]), uint64(cmdBits[copyLen+14]))
		histo[copyLen+14]++
	case copyLen < 134:
		tail := copyLen - 6
		nbits := uint(bits.Len(tail)) - 2
		prefix := tail >> nbits
		code := (nbits << 1) + prefix + 20
		d := uint(depth[code])
		// Pack Huffman code and extra bits into one writeBits.
		// Max total = 15 + 5 = 20 bits.
		c.b.writeBits(d+nbits, uint64(cmdBits[code])|uint64(tail-(prefix<<nbits))<<d)
		histo[code]++
	case copyLen < 2118:
		tail := copyLen - 70
		nbits := uint(bits.Len(tail)) - 1
		code := nbits + 28
		d := uint(depth[code])
		// Max total = 15 + 11 = 26 bits.
		c.b.writeBits(d+nbits, uint64(cmdBits[code])|uint64(tail-(1<<nbits))<<d)
		histo[code]++
	default:
		d := uint(depth[39])
		// Max total = 15 + 24 = 39 bits.
		c.b.writeBits(d+24, uint64(cmdBits[39])|uint64(copyLen-2118)<<d)
		histo[39]++
	}
}

// writeCopyLenLastDistance writes a copy length code with last distance reuse.
func (c *fragmentCompressor) writeCopyLenLastDistance(copyLen uint) {
	depth := c.arena.cmdDepth[:]
	cmdBits := c.arena.cmdBits[:]
	histo := c.arena.cmdHisto[:]
	switch {
	case copyLen < 12:
		c.b.writeBits(uint(depth[copyLen-4]), uint64(cmdBits[copyLen-4]))
		histo[copyLen-4]++
	case copyLen < 72:
		tail := copyLen - 8
		nbits := uint(bits.Len(tail)) - 2
		prefix := tail >> nbits
		code := (nbits << 1) + prefix + 4
		d := uint(depth[code])
		// Pack Huffman code and extra bits into one writeBits.
		// Max total = 15 + 5 = 20 bits.
		c.b.writeBits(d+nbits, uint64(cmdBits[code])|uint64(tail-(prefix<<nbits))<<d)
		histo[code]++
	case copyLen < 136:
		tail := copyLen - 8
		code := (tail >> 5) + 30
		d := uint(depth[code])
		d64 := uint(depth[64])
		// Pack copy code, 5 extra bits, and last-distance code into one
		// writeBits. Max total = 15 + 5 + 15 = 35 bits.
		c.b.writeBits(d+5+d64,
			uint64(cmdBits[code])|uint64(tail&31)<<d|uint64(cmdBits[64])<<(d+5))
		histo[code]++
		histo[64]++
	case copyLen < 2120:
		tail := copyLen - 72
		nbits := uint(bits.Len(tail)) - 1
		code := nbits + 28
		d := uint(depth[code])
		d64 := uint(depth[64])
		// Max total = 15 + 11 + 15 = 41 bits.
		c.b.writeBits(d+nbits+d64,
			uint64(cmdBits[code])|uint64(tail-(1<<nbits))<<d|uint64(cmdBits[64])<<(d+nbits))
		histo[code]++
		histo[64]++
	default:
		d := uint(depth[39])
		d64 := uint(depth[64])
		// Max total = 15 + 24 + 15 = 54 bits, just under the 56-bit limit.
		c.b.writeBits(d+24+d64,
			uint64(cmdBits[39])|uint64(copyLen-2120)<<d|uint64(cmdBits[64])<<(d+24))
		histo[39]++
		histo[64]++
	}
}

// writeDistanceAndCopyLenLastDistance emits a fresh distance code immediately
// followed by a copy length code that uses the implicit last-distance reuse
// alphabet. For copyLen < 72 (the common first-match case on real corpora)
// both Huffman codes and their extra bits fit in a single 56-bit writeBits
// call. Folding the two writes into one cuts the bitstream load-modify-store
// dependency chain in half for the q0 first-match path, mirroring the
// inner-repeat-match fold inlined directly into writeCommands.
func (c *fragmentCompressor) writeDistanceAndCopyLenLastDistance(distance, copyLen uint) {
	if copyLen >= 72 {
		c.writeDistance(distance)
		c.writeCopyLenLastDistance(copyLen)
		return
	}

	depth := c.arena.cmdDepth[:]
	cmdBits := c.arena.cmdBits[:]
	histo := c.arena.cmdHisto[:]

	// Distance: 2*(nbits-1)+prefix+80 ∈ [80,111]; total bits ≤ 15+16 = 31.
	// Masking shift amounts with 63 lets the compiler elide the variable-shift
	// safety mask (dNbits ≤ 62, dDepth ≤ 15 in practice).
	dd := distance + 3
	dNbits := (uint(bits.Len(dd)) - 2) & 63
	dPrefix := (dd >> dNbits) & 1
	dOffset := (2 + dPrefix) << dNbits
	distcode := 2*(dNbits-1) + dPrefix + 80
	dDepth := uint(depth[distcode]) & 63
	dBits := uint64(cmdBits[distcode]) | uint64(dd-dOffset)<<dDepth
	dLen := dDepth + dNbits

	// Copy length with last-distance reuse: either depth-only (copyLen < 12)
	// or depth+nbits (≤ 20 bits).
	var cBits uint64
	var cLen uint
	var ccode uint
	if copyLen < 12 {
		ccode = copyLen - 4
		cBits = uint64(cmdBits[ccode])
		cLen = uint(depth[ccode])
	} else {
		tail := copyLen - 8
		nbits := (uint(bits.Len(tail)) - 2) & 63
		prefix := tail >> nbits
		ccode = (nbits << 1) + prefix + 4
		d := uint(depth[ccode]) & 63
		cBits = uint64(cmdBits[ccode]) | uint64(tail-(prefix<<nbits))<<d
		cLen = d + nbits
	}

	// Combined ≤ 31 + 20 = 51 bits, safely under the 56-bit writeBits limit.
	c.b.writeBits(dLen+cLen, dBits|cBits<<(dLen&63))
	histo[ccode]++
	histo[distcode]++
}

// writeDistance writes a distance code to the bitstream.
func (c *fragmentCompressor) writeDistance(distance uint) {
	d := distance + 3
	nbits := uint(bits.Len(d)) - 2
	prefix := (d >> nbits) & 1
	offset := (2 + prefix) << nbits
	distcode := 2*(nbits-1) + prefix + 80
	depth := uint(c.arena.cmdDepth[distcode])
	// Pack Huffman code (depth bits) and extra bits (nbits) into one writeBits
	// call. Max total = 15 + 16 = 31 bits, well under the 56-bit limit.
	c.b.writeBits(depth+nbits, uint64(c.arena.cmdBits[distcode])|uint64(d-offset)<<depth)
	c.arena.cmdHisto[distcode]++
}

// writeLiterals writes literal bytes to the bitstream.
func (c *fragmentCompressor) writeLiterals(input []byte) {
	c.b.writeLiteralBits(input, &c.arena.litDepth, &c.arena.litBits)
}

// writeUncompressedMetaBlock rewrites the output as an uncompressed meta-block.
func (c *fragmentCompressor) writeUncompressedMetaBlock(data []byte, startBitOffset uint) {
	c.b.rewindTo(startBitOffset)
	c.b.writeUncompressedMetaBlock(data)
}

// shouldUseUncompressedMode returns true when the data so far looks
// incompressible enough to emit an uncompressed meta-block.
func (c *fragmentCompressor) shouldUseUncompressedMode(insertLen uint) bool {
	if uint(c.nextEmit-c.metablockStart)*50 > insertLen {
		return false
	}
	return c.literalRatio > minRatio
}

// TODO: take *encoder instead of *bitWriter
// compressFragmentFast compresses the input as one or more complete meta-blocks
// and writes them to the bitstream. If isLast is true, an additional empty last
// meta-block is emitted.
//
// The arena s is used for scratch space and also carries the command prefix code
// state between calls: cmdDepth, cmdBits, cmdCode, cmdCodeNumBits must be
// properly initialized before the first call.
//
// table must be zeroed on the first call. Its length must be a power of two
// with an odd exponent (9, 11, 13, or 15).
func compressFragmentFast(
	s *onePassArena,
	input []byte,
	isLast bool,
	table []uint32,
	b *bitWriter,
) {
	c := &fragmentCompressor{
		arena:  s,
		b:      b,
		input:  input,
		table:  table,
		isLast: isLast,
	}
	c.compress()
}

// hashBytesAtOffset computes a hash of 5 bytes within a 64-bit value,
// starting at the given byte offset (0..3).
func hashBytesAtOffset(v uint64, offset, shift uint) uint32 {
	h := ((v >> (8 * offset)) << 24) * hashMul32
	return uint32(h >> (shift & 63))
}

// isMatch returns true if the 5 bytes of input at positions a and b are equal.
// Taking the raw byte slice and indices avoids the sub-slice bounds check at
// the call site in the inner scan loop.
func isMatch(input []byte, a, b uint) bool {
	return loadU32LE(input, a) == loadU32LE(input, b) &&
		loadByte(input, a+4) == loadByte(input, b+4)
}

// updateBits overwrites nBits bits at bit position pos in array with the
// given value.
func updateBits(array []byte, nBits uint, value uint32, pos uint) {
	for nBits > 0 {
		bytePos := pos / 8
		unchangedBits := pos & 7
		changedBits := min(nBits, 8-unchangedBits)
		totalBits := unchangedBits + changedBits
		mask := (^((uint32(1) << totalBits) - 1)) | ((uint32(1) << unchangedBits) - 1)
		unchanged := uint32(array[bytePos]) & mask
		changed := value & ((uint32(1) << changedBits) - 1)
		array[bytePos] = byte((changed << unchangedBits) | unchanged)
		nBits -= changedBits
		value >>= changedBits
		pos += changedBits
	}
}
