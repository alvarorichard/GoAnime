// Portable fallback for platforms without unaligned little-endian access
// (big-endian, strict-alignment archs such as arm/mips/riscv, or purego builds):
// safe-Go load primitives and bit writer hot loops that avoid unsafe pointer
// arithmetic.

//go:build purego || !(amd64 || 386 || arm64 || loong64 || ppc64le || wasm)

package encoder

import "encoding/binary"

//go:nosplit
func loadByte(b []byte, i uint) byte {
	return b[i]
}

//go:nosplit
func loadU32LE(b []byte, i uint) uint32 {
	_ = b[i+3]
	return uint32(b[i]) | uint32(b[i+1])<<8 | uint32(b[i+2])<<16 | uint32(b[i+3])<<24
}

//go:nosplit
func loadU64LE(b []byte, i uint) uint64 {
	_ = b[i+7]
	return uint64(b[i]) | uint64(b[i+1])<<8 | uint64(b[i+2])<<16 | uint64(b[i+3])<<24 |
		uint64(b[i+4])<<32 | uint64(b[i+5])<<40 | uint64(b[i+6])<<48 | uint64(b[i+7])<<56
}

func copy16(dst []byte, d uint, src []byte, s uint) {
	copy(dst[d:d+16], src[s:s+16])
}

// writeBits packs value into the bitstream and advances the bit position.
// Up to 56 bits may be written at a time.
func (b *bitWriter) writeBits(nbits uint, value uint64) {
	bytePos := b.bitOffset >> 3
	bitOff := b.bitOffset & 7
	p := b.buf[bytePos : bytePos+8]
	v := uint64(p[0]) | value<<bitOff
	binary.LittleEndian.PutUint64(p, v)
	b.bitOffset += nbits
}

// writeBitsAt is the variant of writeBits used by hot loops (huffmanBlock.
// writeData) that keep both the output buffer and the bit position in
// locals across many calls. It takes the buffer and current bitOffset and
// returns the updated bitOffset, so the caller can hold the slice header and
// the offset in registers instead of round-tripping them through the
// bitWriter on every writeBits.
//
//go:nosplit
func writeBitsAt(buf []byte, bitOffset, nbits uint, value uint64) uint {
	bytePos := bitOffset >> 3
	bitOff := bitOffset & 7
	p := buf[bytePos : bytePos+8]
	v := uint64(p[0]) | value<<bitOff
	binary.LittleEndian.PutUint64(p, v)
	return bitOffset + nbits
}

func (b *bitWriter) writeLiteralBits(input []byte, depths *[256]byte, bits *[256]uint16) {
	b.bitOffset = writeLiteralBitsAt(b.buf, b.bitOffset, input, depths, bits)
}

func writeLiteralBitsAt(buf []byte, bitOffset uint, input []byte, depths *[256]byte, bits *[256]uint16) uint {
	for _, lit := range input {
		bytePos := bitOffset >> 3
		bitOff := bitOffset & 7
		p := buf[bytePos : bytePos+8]
		v := uint64(p[0]) | uint64(bits[lit])<<bitOff
		binary.LittleEndian.PutUint64(p, v)
		bitOffset += uint(depths[lit])
	}
	return bitOffset
}
