// Huffman-coded meta-block data: commands, literals, and distances.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// huffmanBlock holds input data, commands, and Huffman codes for the three
// prefix code alphabets (literal, insert-and-copy, distance).
type huffmanBlock struct {
	input    []byte
	commands []command

	litDepth  []byte
	litBits   []uint16
	cmdDepth  []byte
	cmdBits   []uint16
	distDepth []byte
	distBits  []uint16

	startPos uint
	mask     uint
}

// writeData encodes a sequence of commands into the bitstream using the
// provided Huffman codes. For each command it writes:
//  1. The insert-and-copy length symbol + extra bits
//  2. The literal bytes covered by the insert length
//  3. The distance symbol + extra bits (if the command has a backward reference)
func (block huffmanBlock) writeData(b *bitWriter) { //nolint:gocritic // stack copy avoids heap escape; called once per metablock
	pos := block.startPos
	// Hoist b.bitOffset and b.buf to locals across the whole loop. Each
	// writeBits call reads and writes b.bitOffset and re-derives the buffer
	// base pointer; the compiler can't keep them in registers across calls
	// because the writes into b.buf could in theory alias the bitWriter
	// fields. Holding them locally collapses 4–5 redundant load/stores per
	// command into a single load at entry and store at exit.
	buf := b.buf
	bitOffset := b.bitOffset
	for _, cmd := range block.commands {
		cmdCode := cmd.cmdPrefix
		// Fold the cmd Huffman code and the insert/copy extra bits into a
		// single writeBitsAt call. cmdDepth ≤ 15 and InsertLenExtraBits +
		// CopyLenExtraBits ≤ 48, so the combined width is ≤ 63 bits. For
		// typical commands (short inserts/copies) the combined width is well
		// under 56; fall back to two writes when it would otherwise overflow
		// the 56-bit writeBitsAt limit.
		cmdDepth := uint(block.cmdDepth[cmdCode])
		cmdBits := uint64(block.cmdBits[cmdCode])
		// Use the command prefix lookup table to avoid recomputing the
		// insert/copy length prefix classes for every command.
		lut := core.CmdLut[cmdCode]
		copyLenCode := cmd.copyLen
		// Most copy commands have no encoded length delta; avoid the
		// sign-extension path unless the high delta bits are present.
		if copyLenCode>>25 != 0 {
			copyLenCode = cmd.copyLenCode()
		}
		effCopyLen := uint(copyLenCode)
		insNumExtra := uint(lut.InsertLenExtraBits)
		extraBitsLen := insNumExtra + uint(lut.CopyLenExtraBits)
		extraBits := ((uint64(effCopyLen) - uint64(lut.CopyLenOffset)) << insNumExtra) |
			(uint64(cmd.insertLen) - uint64(lut.InsertLenOffset))
		if cmdDepth+extraBitsLen <= 56 {
			bitOffset = writeBitsAt(buf, bitOffset,
				cmdDepth+extraBitsLen, cmdBits|extraBits<<cmdDepth)
		} else {
			bitOffset = writeBitsAt(buf, bitOffset, cmdDepth, cmdBits)
			bitOffset = writeBitsAt(buf, bitOffset, extraBitsLen, extraBits)
		}
		// Literal encoding loop: three consecutive literals are packed into a
		// single writeBits call. Each Brotli literal code is at most 15 bits,
		// so three codes total at most 45 bits — well within the 56-bit limit.
		// This reduces writeBits call count ~3x in the common case.
		j := cmd.insertLen
		for j != 0 {
			lit0 := block.input[pos&block.mask]
			n0 := uint(block.litDepth[lit0])
			v0 := uint64(block.litBits[lit0])
			pos++
			j--
			if j == 0 {
				bitOffset = writeBitsAt(buf, bitOffset, n0, v0)
				break
			}
			lit1 := block.input[pos&block.mask]
			n1 := uint(block.litDepth[lit1])
			v1 := uint64(block.litBits[lit1])
			pos++
			j--
			if j == 0 {
				bitOffset = writeBitsAt(buf, bitOffset, n0+n1, v0|v1<<n0)
				break
			}
			lit2 := block.input[pos&block.mask]
			n2 := uint(block.litDepth[lit2])
			pos++
			j--
			bitOffset = writeBitsAt(buf, bitOffset, n0+n1+n2,
				v0|v1<<n0|uint64(block.litBits[lit2])<<(n0+n1))
		}
		copyLen := cmd.copyLen & 0x1FFFFFF
		pos += uint(copyLen)
		if copyLen != 0 && cmdCode >= 128 {
			distCode := cmd.distPrefix & 0x3FF
			distBitsLen := uint(block.distDepth[distCode])
			distNumExtra := uint(cmd.distPrefix >> 10)
			// Merge the distance Huffman code and its extra bits into one
			// writeBits call. distBitsLen ≤ 15, distNumExtra ≤ ~24, total ≤ 39 bits.
			bitOffset = writeBitsAt(buf, bitOffset, distBitsLen+distNumExtra,
				uint64(block.distBits[distCode])|(uint64(cmd.distExtra)<<distBitsLen))
		}
	}
	b.bitOffset = bitOffset
}
