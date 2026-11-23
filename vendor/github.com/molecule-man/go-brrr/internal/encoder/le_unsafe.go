// Unsafe fast path for little-endian platforms: load primitives and bit
// writer hot loops that rely on unaligned uint32/uint64 reads and writes.

//go:build !purego && (amd64 || 386 || arm64 || loong64 || ppc64le || wasm)

package encoder

import "unsafe"

//go:nosplit
func loadByte(b []byte, i uint) byte {
	return *(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b)), i))
}

//go:nosplit
func loadU32LE(b []byte, i uint) uint32 {
	return *(*uint32)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b)), i))
}

//go:nosplit
func loadU64LE(b []byte, i uint) uint64 {
	return *(*uint64)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b)), i))
}

// copy16 copies 16 bytes from src[s:] to dst[d:]. The caller checks the bounds.
func copy16(dst []byte, d uint, src []byte, s uint) {
	sp := unsafe.Add(unsafe.Pointer(unsafe.SliceData(src)), s)
	dp := unsafe.Add(unsafe.Pointer(unsafe.SliceData(dst)), d)
	*(*uint64)(dp) = *(*uint64)(sp)
	*(*uint64)(unsafe.Add(dp, 8)) = *(*uint64)(unsafe.Add(sp, 8))
}

// writeBits packs value into the bitstream and advances the bit position.
// Up to 56 bits may be written at a time.
func (b *bitWriter) writeBits(nbits uint, value uint64) {
	bytePos := b.bitOffset >> 3
	bitOff := b.bitOffset & 7
	p := (*uint64)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b.buf)), bytePos))
	*p = uint64(*(*byte)(unsafe.Pointer(p))) | value<<bitOff
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
	p := (*uint64)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(buf)), bytePos))
	*p = uint64(*(*byte)(unsafe.Pointer(p))) | value<<bitOff
	return bitOffset + nbits
}

// writeLiteralBits requires depths of at most 14 bits.
func (b *bitWriter) writeLiteralBits(input []byte, depths *[256]byte, bits *[256]uint16) {
	b.bitOffset = writeLiteralBitsAt(b.buf, b.bitOffset, input, depths, bits)
}

// Four literal codes use at most 56 bits. They fit in one 64-bit store with a seven-bit offset.
// A large caller spills the loop state to the stack if the compiler inlines this function.
//
//go:noinline
func writeLiteralBitsAt(buf []byte, bitOffset uint, input []byte, depths *[256]byte, bits *[256]uint16) uint {
	bufBase := unsafe.Pointer(unsafe.SliceData(buf))
	src := unsafe.Pointer(unsafe.SliceData(input))
	d, c := depths[:], bits[:]
	n := uint(len(input))
	i := uint(0)
	for ; i+4 <= n; i += 4 {
		l0 := *(*byte)(unsafe.Add(src, i))
		l1 := *(*byte)(unsafe.Add(src, i+1))
		l2 := *(*byte)(unsafe.Add(src, i+2))
		l3 := *(*byte)(unsafe.Add(src, i+3))
		n0 := uint(d[l0])
		n01 := n0 + uint(d[l1])
		n012 := n01 + uint(d[l2])
		v := uint64(c[l0]) | uint64(c[l1])<<(n0&63) |
			uint64(c[l2])<<(n01&63) | uint64(c[l3])<<(n012&63)
		p := (*uint64)(unsafe.Add(bufBase, bitOffset>>3))
		*p = uint64(*(*byte)(unsafe.Pointer(p))) | v<<(bitOffset&7)
		bitOffset += n012 + uint(d[l3])
	}
	for ; i < n; i++ {
		lit := *(*byte)(unsafe.Add(src, i))
		p := (*uint64)(unsafe.Add(bufBase, bitOffset>>3))
		*p = uint64(*(*byte)(unsafe.Pointer(p))) | uint64(c[lit])<<(bitOffset&7)
		bitOffset += uint(d[lit])
	}
	return bitOffset
}
