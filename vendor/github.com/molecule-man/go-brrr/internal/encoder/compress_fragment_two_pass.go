package encoder

// Two-pass fast encoder for an input fragment, independent of input history.
// In the first pass, backward matches and literal bytes are saved into
// buffers. In the second pass, they are emitted into the bitstream using
// prefix codes built from actual command and literal histograms.

import (
	"math/bits"
	"unsafe"
)

//go:generate go run ../../cmd/genfixedshift -in compress_fragment_two_pass.go -out compress_fragment_two_pass_table17.go -func createCommandsMinMatch6 -name createCommandsTable17 -bits 17

const twoPassBlockSize = 1 << 17

// sampleRate is the byte sampling interval used by shouldCompress to
// estimate entropy without scanning every byte.
const sampleRate = 43

// numExtraBits maps each command code (0-127) to the number of extra bits
// that follow it in the bitstream.
var numExtraBits = [128]uint{
	0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5,
	6, 7, 8, 9, 10, 12, 14, 24, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 2, 2, 3, 3, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 7, 8, 9, 10, 24,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8,
	9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 14, 14, 15, 15, 16, 16,
	17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22, 23, 23, 24, 24,
}

// insertOffset maps each insert length code (0-23) to the base insert length.
var insertOffset = [24]uint{
	0, 1, 2, 3, 4, 5, 6, 8, 10, 14, 18, 26,
	34, 50, 66, 98, 130, 194, 322, 578, 1090, 2114, 6210, 22594,
}

// copyLenCodeQ1Small precomputes the q1 copy-length command word for every
// copyLen in [4, 133] - the dominant range on real-world corpora. The packed
// uint32 is exactly the value to store into commands[]: bits 0..7 are the
// command code (43..47 for cl<10, 48..57 for cl<134) and bits 8..23 are the
// extra-bits payload. This replaces the bits.Len + shift chain in the
// inner-repeat-match emit loop with one table load . ccode is recovered as `cmd & 0x7F`.
var copyLenCodeQ1Small = func() [134]uint32 {
	var t [134]uint32
	for cl := 4; cl < 134; cl++ {
		if cl < 10 {
			t[cl] = uint32(cl + 38)
			continue
		}
		tail := uint32(cl - 6)
		nbits := uint32(bits.Len32(tail)) - 2
		prefix := tail >> nbits
		ccode := (nbits << 1) + prefix + 44
		extra := tail - (prefix << nbits)
		t[cl] = ccode | extra<<8
	}
	return t
}()

// insertLenCodeQ1Small stores the command word for each insertLen below 130.
// Bits 0..6 contain the command code.
var insertLenCodeQ1Small = func() [130]uint32 {
	var t [130]uint32
	var h [128]uint32
	for n := range uint(130) {
		encodeInsertLen(t[n:], n, &h)
	}
	return t
}()

// copyLenLastDistanceCodeQ1Small stores each single-slot command word below 72.
// Bits 0..6 contain the command code.
var copyLenLastDistanceCodeQ1Small = func() [72]uint32 {
	var t [72]uint32
	var h [128]uint32
	for n := uint(4); n < 72; n++ {
		encodeCopyLenLastDistance(t[n:], n, &h)
	}
	return t
}()

type twoPassCompressor struct {
	arena      *twoPassArena
	b          *bitWriter
	input      []byte
	table      []uint32
	commandBuf []uint32
	literalBuf []byte
	tableBits  uint
	minMatch   int
	isLast     bool
}

func (c *twoPassCompressor) compress() {
	c.tableBits = uint(bits.Len(uint(len(c.table)))) - 1
	if c.tableBits <= 15 {
		c.minMatch = 4
	} else {
		c.minMatch = 6
	}
	initialBitOffset := c.b.bitOffset

	input := c.input
	pos := 0

	inputSize := len(input)
	for inputSize > 0 {
		blockSize := min(inputSize, twoPassBlockSize)
		commands := c.commandBuf
		literals := c.literalBuf

		numCommands, numLiterals := c.createCommands(
			input, pos, blockSize, inputSize, commands, literals)

		if c.shouldCompress(input[pos:pos+blockSize], numLiterals) {
			c.b.writeMetaBlockHeader(blockSize, false, false)
			// No block splits, no contexts.
			c.b.writeBits(13, 0)
			c.writeCommands(literals[:numLiterals], commands[:numCommands])
		} else {
			// Since we did not find many backward references and the entropy of
			// the data is close to 8 bits, we can simply emit an uncompressed
			// block. This makes compression of uncompressible data about 3x
			// faster.
			c.b.writeUncompressedMetaBlock(input[pos : pos+blockSize])
		}
		pos += blockSize
		inputSize -= blockSize
	}

	// If output is larger than a single uncompressed block, rewrite it.
	if c.b.bitOffset-initialBitOffset > 31+uint(len(input))*8 {
		c.b.rewindTo(initialBitOffset)
		c.b.writeUncompressedMetaBlock(input)
	}

	if c.isLast {
		c.b.writeBits(1, 1) // islast
		c.b.writeBits(1, 1) // isempty
		c.b.byteAlign()
	}
}

// createCommands performs the first pass: scan input for backward matches and
// write insert/copy/distance commands into commandBuf and literal bytes into
// literalBuf. Returns the number of commands and literals produced.
//
// input is the full input buffer; pos is the start of the current block.
// blockSize is the size of the current block; inputSize is the remaining
// input size from pos onward.
func (c *twoPassCompressor) createCommands(
	input []byte, pos, blockSize, inputSize int,
	commands []uint32, literals []byte,
) (numCommands, numLiterals int) {
	// Zero the command histogram before this block. createCommands builds the
	// histogram incrementally as it emits commands, saving the rescan pass that
	// writeCommands used to do.
	c.arena.cmdHisto = [128]uint32{}

	if c.minMatch == 6 {
		if c.tableBits == 17 {
			return c.createCommandsTable17(input, pos, blockSize, inputSize, commands, literals)
		}
		return c.createCommandsMinMatch6(input, pos, blockSize, inputSize, commands, literals)
	}

	cmdHisto := &c.arena.cmdHisto
	ip := pos
	shift := 64 - c.tableBits
	ipEnd := pos + blockSize
	nextEmit := pos
	lastDistance := -1
	table := c.table
	tbl := unsafe.Pointer(unsafe.SliceData(table))
	const minMatch = 4
	cmdPos := 0
	litPos := 0

	if blockSize < inputMarginBytes {
		goto encodeRemainder
	}

	{
		// For the last block, keep a 16-byte margin so all distances are at
		// most window size - 16. For other blocks, keep a margin of minMatch
		// bytes so we don't go past the block size with a copy.
		lenLimit := min(blockSize-minMatch, inputSize-inputMarginBytes)
		ipLimit := pos + lenLimit

		ip++

		for {
			var candidate int
			var found bool
			ip, candidate, found = findMatchMinMatch4(input, tbl, ip, ipLimit, lastDistance, shift)
			if !found {
				goto encodeRemainder
			}

			// Step 2: Emit the found match together with the literal bytes from
			// nextEmit, and then see if we can find a next match immediately
			// afterwards.
			{
				base := ip
				matched := minMatch + matchLenAt(
					input, uint(candidate+minMatch), uint(ip+minMatch), ipEnd-ip-minMatch)
				distance := base - candidate
				insert := base - nextEmit
				ip += matched

				if u := uint(insert); u < 130 {
					cmd := insertLenCodeQ1Small[u]
					commands[cmdPos] = cmd
					cmdHisto[cmd&0x7F]++
					cmdPos++
				} else {
					cmdPos += encodeInsertLen(commands[cmdPos:], u, cmdHisto)
				}
				// The scan margin makes the 16-byte source read safe.
				if insert <= 16 && litPos+16 <= len(literals) {
					copy16(literals, uint(litPos), input, uint(nextEmit))
				} else {
					copy(literals[litPos:], input[nextEmit:nextEmit+insert])
				}
				litPos += insert
				if distance == lastDistance {
					commands[cmdPos] = 64
					cmdHisto[64]++
					cmdPos++
				} else {
					cmdPos += encodeDistance(commands[cmdPos:], uint(distance), cmdHisto)
					lastDistance = distance
				}
				if cl := uint(matched); cl < 72 {
					cmd := copyLenLastDistanceCodeQ1Small[cl]
					commands[cmdPos] = cmd
					cmdHisto[cmd&0x7F]++
					cmdPos++
				} else {
					cmdPos += encodeCopyLenLastDistance(commands[cmdPos:], cl, cmdHisto)
				}

				nextEmit = ip
				if ip >= ipLimit {
					goto encodeRemainder
				}

				candidate = c.updateHashTableTwoPass(input, table, ip, shift, minMatch, 0)
			}

			// Try to find another match immediately.
			for ip-candidate <= maxDistance &&
				isMatchTwoPass4At(input, uint(ip), uint(candidate)) {
				base := ip
				matched := minMatch + matchLenAt(
					input, uint(candidate+minMatch), uint(ip+minMatch), ipEnd-ip-minMatch)
				ip += matched
				lastDistance = base - candidate

				// Fast path for cl < 134 (the dominant range on real corpora):
				// one table load delivers both the command word and ccode,
				// replacing the bits.Len + shift chain.
				var ccode uint
				if cl := uint(matched); cl < 134 {
					cmd := copyLenCodeQ1Small[cl]
					commands[cmdPos] = cmd
					ccode = uint(cmd & 0x7F)
				} else if cl < 2118 {
					tail := cl - 70
					nbits := uint(bits.Len(tail)) - 1
					ccode = nbits + 52
					extra := tail - (1 << nbits)
					commands[cmdPos] = uint32(ccode) | uint32(extra)<<8
				} else {
					ccode = 63
					extra := cl - 2118
					commands[cmdPos] = uint32(ccode) | uint32(extra)<<8
				}
				cmdHisto[ccode]++
				cmdPos++

				cmdPos += encodeDistance(commands[cmdPos:], uint(lastDistance), cmdHisto)

				nextEmit = ip
				if ip >= ipLimit {
					goto encodeRemainder
				}

				candidate = c.updateHashTableTwoPass(input, table, ip, shift, minMatch, 2)
			}

			ip++
		}
	} // close block scope for ipLimit, lenLimit

encodeRemainder:
	// Emit the remaining bytes as literals.
	if nextEmit < ipEnd {
		insert := ipEnd - nextEmit
		cmdPos += encodeInsertLen(commands[cmdPos:], uint(insert), cmdHisto)
		copy(literals[litPos:], input[nextEmit:ipEnd])
		litPos += insert
	}
	return cmdPos, litPos
}

// genfixedshift uses createCommandsMinMatch6 to generate createCommandsTable17.
//
// createCommandsMinMatch6 is the quality-1 large-table path. Keeping the
// 6-byte minimum match as a constant removes minMatch branches from the hot
// scan and hash-table update loops used for real large inputs.
func (c *twoPassCompressor) createCommandsMinMatch6(
	input []byte, pos, blockSize, inputSize int,
	commands []uint32, literals []byte,
) (numCommands, numLiterals int) {
	const minMatch = 6

	cmdHisto := &c.arena.cmdHisto
	ip := pos
	shift := 64 - c.tableBits
	ipEnd := pos + blockSize
	nextEmit := pos
	lastDistance := -1
	tbl := unsafe.Pointer(unsafe.SliceData(c.table))
	cmdPos := 0
	litPos := 0

	if blockSize < inputMarginBytes {
		goto encodeRemainder
	}

	{
		// For the last block, keep a 16-byte margin so all distances are at
		// most window size - 16. For other blocks, keep a margin of minMatch
		// bytes so we don't go past the block size with a copy.
		lenLimit := min(blockSize-minMatch, inputSize-inputMarginBytes)
		ipLimit := pos + lenLimit

		ip++

		for {
			var candidate int
			var found bool
			ip, candidate, found = findMatchMinMatch6(input, tbl, ip, ipLimit, lastDistance, shift)
			if !found {
				goto encodeRemainder
			}

			// Step 2: Emit the found match together with the literal bytes from
			// nextEmit, and then see if we can find a next match immediately
			// afterwards.
			{
				base := ip
				matched := minMatch + matchLenAt(
					input, uint(candidate+minMatch), uint(ip+minMatch), ipEnd-ip-minMatch)
				distance := base - candidate
				insert := base - nextEmit
				ip += matched

				if u := uint(insert); u < 130 {
					cmd := insertLenCodeQ1Small[u]
					commands[cmdPos] = cmd
					cmdHisto[cmd&0x7F]++
					cmdPos++
				} else {
					cmdPos += encodeInsertLen(commands[cmdPos:], u, cmdHisto)
				}
				// The scan margin makes the 16-byte source read safe.
				if insert <= 16 && litPos+16 <= len(literals) {
					copy16(literals, uint(litPos), input, uint(nextEmit))
				} else {
					copy(literals[litPos:], input[nextEmit:nextEmit+insert])
				}
				litPos += insert
				if distance == lastDistance {
					commands[cmdPos] = 64
					cmdHisto[64]++
					cmdPos++
				} else {
					cmdPos += encodeDistance(commands[cmdPos:], uint(distance), cmdHisto)
					lastDistance = distance
				}
				if cl := uint(matched); cl < 72 {
					cmd := copyLenLastDistanceCodeQ1Small[cl]
					commands[cmdPos] = cmd
					cmdHisto[cmd&0x7F]++
					cmdPos++
				} else {
					cmdPos += encodeCopyLenLastDistance(commands[cmdPos:], cl, cmdHisto)
				}
			}

			// Try to find another match immediately.
			for {
				nextEmit = ip
				if ip >= ipLimit {
					goto encodeRemainder
				}

				lo := loadU64LE(input, uint(ip-5))
				hi := loadU64LE(input, uint(ip-2))
				*(*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(lo, 0, shift))*4)) = uint32(ip - 5)
				*(*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(lo, 1, shift))*4)) = uint32(ip - 4)
				*(*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(lo, 2, shift))*4)) = uint32(ip - 3)
				*(*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(hi, 0, shift))*4)) = uint32(ip - 2)
				*(*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(hi, 1, shift))*4)) = uint32(ip - 1)
				curPtr := (*uint32)(unsafe.Add(tbl, uintptr(hashBytesAtOffsetTwoPass6(hi, 2, shift))*4))
				candidate = int(*curPtr)
				*curPtr = uint32(ip)
				if ip-candidate > maxDistance || (loadU64LE(input, uint(candidate))^(hi>>16))<<16 != 0 {
					break
				}

				base := ip
				matched := minMatch + matchLenAt(
					input, uint(candidate+minMatch), uint(ip+minMatch), ipEnd-ip-minMatch)
				ip += matched
				lastDistance = base - candidate

				// Fast path for cl < 134 (the dominant range on real corpora):
				// one table load delivers both the command word and ccode,
				// replacing the bits.Len + shift chain.
				var ccode uint
				if cl := uint(matched); cl < 134 {
					cmd := copyLenCodeQ1Small[cl]
					commands[cmdPos] = cmd
					ccode = uint(cmd & 0x7F)
				} else if cl < 2118 {
					tail := cl - 70
					nbits := uint(bits.Len(tail)) - 1
					ccode = nbits + 52
					extra := tail - (1 << nbits)
					commands[cmdPos] = uint32(ccode) | uint32(extra)<<8
				} else {
					ccode = 63
					extra := cl - 2118
					commands[cmdPos] = uint32(ccode) | uint32(extra)<<8
				}
				cmdHisto[ccode]++
				cmdPos++

				cmdPos += encodeDistance(commands[cmdPos:], uint(lastDistance), cmdHisto)
			}

			ip++
		}
	} // close block scope for ipLimit, lenLimit

encodeRemainder:
	// Emit the remaining bytes as literals.
	if nextEmit < ipEnd {
		insert := ipEnd - nextEmit
		cmdPos += encodeInsertLen(commands[cmdPos:], uint(insert), cmdHisto)
		copy(literals[litPos:], input[nextEmit:ipEnd])
		litPos += insert
	}
	return cmdPos, litPos
}

// findMatchMinMatch4 returns the first valid 4-byte match at or after ip.
// It returns false when the scan passes ipLimit.
// A separate function keeps the scan state in registers.
//
//go:noinline
func findMatchMinMatch4(input []byte, tbl unsafe.Pointer, ip, ipLimit, lastDistance int, shift uint) (matchIP, candidate int, found bool) {
	nextHash := hashTwoPass4At(input, uint(ip), shift)
	skip := uint32(32)
	nextIP := ip
	for {
		hash := nextHash
		ip = nextIP
		nextIP = ip + int(skip>>5)
		skip++
		if nextIP > ipLimit {
			return 0, 0, false
		}
		nextHash = hashTwoPass4At(input, uint(nextIP), shift)

		slot := (*uint32)(unsafe.Add(tbl, uintptr(hash)*4))
		candidate = ip - lastDistance
		// The unsigned compare is candidate >= 0 && candidate < ip.
		if uint(candidate) < uint(ip) && isMatchTwoPass4At(input, uint(ip), uint(candidate)) {
			*slot = uint32(ip)
			if ip-candidate <= maxDistance {
				return ip, candidate, true
			}
			continue
		}

		candidate = int(*slot)
		*slot = uint32(ip)
		if isMatchTwoPass4At(input, uint(ip), uint(candidate)) && ip-candidate <= maxDistance {
			return ip, candidate, true
		}
	}
}

// findMatchMinMatch6 returns the first valid 6-byte match at or after ip.
// It returns false when the scan passes ipLimit.
// A separate function keeps the scan state in registers.
//
//go:noinline
func findMatchMinMatch6(input []byte, tbl unsafe.Pointer, ip, ipLimit, lastDistance int, shift uint) (matchIP, candidate int, found bool) {
	nextLoad := loadU64LE(input, uint(ip))
	nextHash := uint32(((nextLoad << 16) * hashMul32) >> (shift & 63))
	skip := uint32(32)
	nextIP := ip
	for {
		hash := nextHash
		ipBytes := nextLoad
		ip = nextIP
		nextIP = ip + int(skip>>5)
		skip++
		if nextIP > ipLimit {
			return 0, 0, false
		}
		nextLoad = loadU64LE(input, uint(nextIP))
		nextHash = uint32(((nextLoad << 16) * hashMul32) >> (shift & 63))

		slot := (*uint32)(unsafe.Add(tbl, uintptr(hash)*4))
		candidate = ip - lastDistance
		// The unsigned compare is candidate >= 0 && candidate < ip.
		if uint(candidate) < uint(ip) &&
			(loadU64LE(input, uint(candidate))^ipBytes)<<16 == 0 {
			*slot = uint32(ip)
			if ip-candidate <= maxDistance {
				return ip, candidate, true
			}
			continue
		}

		candidate = int(*slot)
		*slot = uint32(ip)
		if (loadU64LE(input, uint(candidate))^ipBytes)<<16 == 0 && ip-candidate <= maxDistance {
			return ip, candidate, true
		}
	}
}

// updateHashTableTwoPass updates the hash table with positions from the last
// copy and returns the next candidate position.
//
// thirdOffset controls the byte offset used for hashing position ip-1.
// The C reference uses offset 0 after the first match (Step 2) and offset 2
// in the consecutive-match loop.
func (c *twoPassCompressor) updateHashTableTwoPass(input []byte, table []uint32, ip int, shift uint, minMatch int, thirdOffset uint) int {
	var curHash uint32
	if minMatch == 4 {
		inputBytes := loadU64LE(input, uint(ip-3))
		curHash = hashBytesAtOffsetTwoPass(inputBytes, 3, shift, minMatch)
		prevHash := hashBytesAtOffsetTwoPass(inputBytes, 0, shift, minMatch)
		table[prevHash] = uint32(ip - 3)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, 1, shift, minMatch)
		table[prevHash] = uint32(ip - 2)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, thirdOffset, shift, minMatch)
		table[prevHash] = uint32(ip - 1)
	} else {
		inputBytes := loadU64LE(input, uint(ip-5))
		prevHash := hashBytesAtOffsetTwoPass(inputBytes, 0, shift, minMatch)
		table[prevHash] = uint32(ip - 5)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, 1, shift, minMatch)
		table[prevHash] = uint32(ip - 4)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, 2, shift, minMatch)
		table[prevHash] = uint32(ip - 3)
		inputBytes = loadU64LE(input, uint(ip-2))
		curHash = hashBytesAtOffsetTwoPass(inputBytes, 2, shift, minMatch)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, 0, shift, minMatch)
		table[prevHash] = uint32(ip - 2)
		prevHash = hashBytesAtOffsetTwoPass(inputBytes, 1, shift, minMatch)
		table[prevHash] = uint32(ip - 1)
	}

	candidate := int(table[curHash])
	table[curHash] = uint32(ip)
	return candidate
}

// writeCommands performs the second pass: build prefix codes from the actual
// histograms and write commands and literals to the bitstream.
func (c *twoPassCompressor) writeCommands(literals []byte, commands []uint32) {
	s := c.arena
	b := c.b
	cmdDepth := s.cmdDepth[:]
	cmdBits := s.cmdBits[:]

	s.resetHistograms()

	// Build literal histogram and Huffman code.
	for _, lit := range literals {
		s.litHisto[lit]++
	}
	c.b.buildAndWriteHuffmanTreeFast(s.tree[:], s.litHisto[:],
		uint(len(literals)), 8, s.litDepth[:], s.litBits[:])

	// The command histogram has already been built incrementally during
	// createCommands. Ensure some baseline counts for codes that must exist.
	s.cmdHisto[1]++
	s.cmdHisto[2]++
	s.cmdHisto[64]++
	s.cmdHisto[84]++
	s.buildAndWriteCommandPrefixCode(b)

	// Emit commands and interleaved literals. Inline writeBits with both
	// bitOffset and the buffer base pointer hoisted to locals so the
	// compiler can keep them in registers across iterations; the regular
	// b.writeBits cannot avoid reloading b.bitOffset and re-deriving the
	// buffer base on every call because writes into b.buf could alias the
	// bitWriter fields. writeLiteralBitsAt requires the 14-bit depth limit
	// that encodeHuffmanTree applies to the literal Huffman tree.
	bufBase := unsafe.Pointer(unsafe.SliceData(b.buf))
	bitOffset := b.bitOffset
	litIdx := 0
	for _, cmd := range commands {
		// Brotli insert-and-copy command codes are in [0, 128); masking
		// lets the compiler drop the bounds checks on cmdDepth/cmdBits/
		// numExtraBits which would otherwise hit on every iteration.
		code := uint(cmd) & 0x7F
		extra := uint64(cmd >> 8)
		depth := uint(cmdDepth[code])
		{
			nbits := depth + numExtraBits[code]
			value := uint64(cmdBits[code]) | extra<<depth
			p := unsafe.Add(bufBase, bitOffset>>3)
			storeU64LEPtr(p, uint64(*(*byte)(p))|value<<(bitOffset&7))
			bitOffset += nbits
		}
		if code < 24 {
			j := int(insertOffset[code]) + int(extra)
			bitOffset = writeLiteralBitsAt(b.buf, bitOffset, literals[litIdx:litIdx+j], &s.litDepth, &s.litBits)
			litIdx += j
		}
	}
	b.bitOffset = bitOffset
}

// shouldCompress decides whether to use compressed or uncompressed mode
// for the given block.
func (c *twoPassCompressor) shouldCompress(input []byte, numLiterals int) bool {
	corpusSize := float64(len(input))
	if float64(numLiterals) < 0.98*corpusSize {
		return true
	}
	maxTotalBitCost := corpusSize * 8 * 0.98 / sampleRate
	s := c.arena
	s.litHisto = [256]uint32{}
	for i := 0; i < len(input); i += sampleRate {
		s.litHisto[input[i]]++
	}
	return bitsEntropy(s.litHisto[:]) < maxTotalBitCost
}

// TODO: take *encoder instead of *bitWriter
// compressFragmentTwoPass compresses the input as one or more complete
// meta-blocks using two-pass processing and writes them to the bitstream.
// If isLast is true, an additional empty last meta-block is emitted.
//
// The arena s is used for scratch space.
//
// commandBuf and literalBuf must each have capacity for at least
// twoPassBlockSize elements. table must be zeroed on the first call.
// Its length must be a power of two.
func compressFragmentTwoPass(
	s *twoPassArena,
	input []byte,
	isLast bool,
	commandBuf []uint32,
	literalBuf []byte,
	table []uint32,
	b *bitWriter,
) {
	c := &twoPassCompressor{
		arena:      s,
		b:          b,
		input:      input,
		table:      table,
		commandBuf: commandBuf,
		literalBuf: literalBuf,
		isLast:     isLast,
	}
	c.compress()
}

// hashTwoPass4At computes a 4-byte hash of input[i:], shifted right by
// shift bits (64 - tableBits). Taking the raw byte slice and index avoids
// the sub-slice bounds check at the call site in the inner scan loop.
func hashTwoPass4At(input []byte, i, shift uint) uint32 {
	h := (loadU64LE(input, i) << 32) * hashMul32
	return uint32(h >> shift)
}

// hashBytesAtOffsetTwoPass computes a hash of minMatch bytes within a
// 64-bit value, starting at the given byte offset.
func hashBytesAtOffsetTwoPass(v uint64, offset, shift uint, minMatch int) uint32 {
	h := ((v >> (8 * offset)) << uint((8-minMatch)*8)) * hashMul32
	return uint32(h >> shift)
}

// `shift & 63` is a no-op (shift = 64 - tableBits ∈ [47, 48] on the q=1
// minMatch=6 path) that lets the compiler elide the variable-shift safety
// mask. The scan-loop hashes in createCommandsMinMatch6 use the same trick
// inline.
func hashBytesAtOffsetTwoPass6(v uint64, offset, shift uint) uint32 {
	h := ((v >> (8 * offset)) << 16) * hashMul32
	return uint32(h >> (shift & 63))
}

// isMatchTwoPass4At compares 4 bytes of input at positions a and b. Taking
// the raw byte slice and indices avoids the sub-slice bounds check at the
// call site in the inner scan loop.
func isMatchTwoPass4At(input []byte, a, b uint) bool {
	return loadU32LE(input, a) == loadU32LE(input, b)
}

// encodeInsertLen encodes an insert length command into commands and returns
// the number of command slots used (always 1).
// encodeInsertLen encodes an insert length command into commands and updates
// cmdHisto for the emitted code. Returns the number of command slots used
// (always 1).
func encodeInsertLen(commands []uint32, insertLen uint, cmdHisto *[128]uint32) int {
	var code uint
	switch {
	case insertLen < 6:
		code = insertLen
		commands[0] = uint32(code)
	case insertLen < 130:
		tail := insertLen - 2
		nbits := uint(bits.Len(tail)) - 2
		prefix := tail >> nbits
		code = (nbits << 1) + prefix + 2
		extra := tail - (prefix << nbits)
		commands[0] = uint32(code) | uint32(extra)<<8
	case insertLen < 2114:
		tail := insertLen - 66
		nbits := uint(bits.Len(tail)) - 1
		code = nbits + 10
		extra := tail - (1 << nbits)
		commands[0] = uint32(code) | uint32(extra)<<8
	case insertLen < 6210:
		code = 21
		extra := insertLen - 2114
		commands[0] = uint32(code) | uint32(extra)<<8
	case insertLen < 22594:
		code = 22
		extra := insertLen - 6210
		commands[0] = uint32(code) | uint32(extra)<<8
	default:
		code = 23
		extra := insertLen - 22594
		commands[0] = uint32(code) | uint32(extra)<<8
	}
	cmdHisto[code]++
	return 1
}

// encodeCopyLenLastDistance encodes a copy-with-last-distance command into
// commands, updates cmdHisto for each emitted code, and returns the number of
// command slots used (1 or 2).
func encodeCopyLenLastDistance(commands []uint32, copyLen uint, cmdHisto *[128]uint32) int {
	switch {
	case copyLen < 12:
		code := copyLen + 20
		commands[0] = uint32(code)
		cmdHisto[code]++
		return 1
	case copyLen < 72:
		tail := copyLen - 8
		nbits := uint(bits.Len(tail)) - 2
		prefix := tail >> nbits
		code := (nbits << 1) + prefix + 28
		extra := tail - (prefix << nbits)
		commands[0] = uint32(code) | uint32(extra)<<8
		cmdHisto[code]++
		return 1
	case copyLen < 136:
		tail := copyLen - 8
		code := (tail >> 5) + 54
		extra := tail & 31
		commands[0] = uint32(code) | uint32(extra)<<8
		commands[1] = 64
		cmdHisto[code]++
		cmdHisto[64]++
		return 2
	case copyLen < 2120:
		tail := copyLen - 72
		nbits := uint(bits.Len(tail)) - 1
		code := nbits + 52
		extra := tail - (1 << nbits)
		commands[0] = uint32(code) | uint32(extra)<<8
		commands[1] = 64
		cmdHisto[code]++
		cmdHisto[64]++
		return 2
	default:
		extra := copyLen - 2120
		commands[0] = 63 | uint32(extra)<<8
		commands[1] = 64
		cmdHisto[63]++
		cmdHisto[64]++
		return 2
	}
}

// encodeDistance encodes a distance command into commands, updates cmdHisto
// for the emitted code, and returns the number of command slots used (always
// 1).
func encodeDistance(commands []uint32, distance uint, cmdHisto *[128]uint32) int {
	d := distance + 3
	nbits := uint(bits.Len(d)) - 2
	prefix := (d >> nbits) & 1
	offset := (2 + prefix) << nbits
	distcode := 2*(nbits-1) + prefix + 80
	extra := d - offset
	commands[0] = uint32(distcode) | uint32(extra)<<8
	cmdHisto[distcode&0x7F]++
	return 1
}
