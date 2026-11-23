// Bit-level reading primitives for the decoder input stream.

package brrr

import "unsafe"

// fastInputSlack is the minimum bytes of unconsumed input required for
// fast-path bit reading (162 bits + 11 bytes of margin for the 8-byte refill).
const fastInputSlack = 32

// bitReaderState is a snapshot of a bitReader, used to save and restore
// position for speculative reads that may need to be rolled back.
type bitReaderState struct {
	pos    int
	val    uint64
	bitPos uint
}

// init resets the bit reader to a clean initial state.
func (br *bitReader) init() {
	br.val = 0
	br.bitPos = 0
	br.input = nil
	br.inputLen = 0
	br.pos = 0
	br.inputBase = nil
	br.fastEnd = -1 // no input → always fail fast-path check
}

// setInput sets the input buffer for the bit reader. The reader starts
// consuming from the beginning of data.
func (br *bitReader) setInput(data []byte) {
	br.input = data
	br.inputLen = len(data)
	br.pos = 0
	br.inputBase = unsafe.Pointer(unsafe.SliceData(data))
	br.fastEnd = len(data) - fastInputSlack
}

// warmup pulls bytes until the accumulator is non-empty.
// Returns false if data is required but there is no input available.
func (br *bitReader) warmup() bool {
	if br.bitPos == 0 {
		br.val = 0
		if !br.pullByte() {
			return false
		}
	}
	return true
}

// saveState returns a snapshot of the current reader position.
func (br *bitReader) saveState() bitReaderState {
	return bitReaderState{
		pos:    br.pos,
		val:    br.val,
		bitPos: br.bitPos,
	}
}

// restoreState restores the reader to a previously saved state.
func (br *bitReader) restoreState(s bitReaderState) {
	br.pos = s.pos
	br.val = s.val
	br.bitPos = s.bitPos
}

// availIn returns the number of unconsumed input bytes (excluding bits
// already loaded into the accumulator).
func (br *bitReader) availIn() int {
	return br.inputLen - br.pos
}

// availBits returns the number of valid bits in the accumulator.
func (br *bitReader) availBits() uint {
	return br.bitPos
}

// remainingBytes returns the number of unread bytes, including whole bytes
// buffered in the accumulator. Capped at 1<<30.
func (br *bitReader) remainingBytes() int {
	const remainingCap = 1 << 30
	avail := br.availIn()
	if avail > remainingCap {
		return remainingCap
	}
	return avail + int(br.bitPos>>3)
}

// checkInputAmount reports whether enough unconsumed input remains for
// fast-path bit reading. Uses precomputed fastEnd to avoid a subtraction.
func (br *bitReader) checkInputAmount() bool {
	return br.pos <= br.fastEnd
}

// checkInputAmountAndBits reports whether the fast input window and at least n bits are available.
// The bitwise OR combines both signed boundary checks.
func (br *bitReader) checkInputAmountAndBits(n uint) bool {
	return (br.fastEnd-br.pos)|(int(br.bitPos)-int(n)) >= 0
}

// fillBitWindow ensures that at least nBits+1 bits are available in the
// accumulator. If the accumulator has fewer than 32 valid bits, 4 bytes are
// loaded from input. nBits must be in the range [1..24].
func (br *bitReader) fillBitWindow(nBits uint) {
	_ = nBits // used for documentation; the 64-bit path handles up to 32
	if br.bitPos < 32 {
		br.val |= uint64(le32(*(*uint32)(unsafe.Add(br.inputBase, br.pos)))) << br.bitPos
		br.bitPos += 32
		br.pos += 4
	}
}

// fillBitWindow16 ensures at least 17 bits are available.
func (br *bitReader) fillBitWindow16() {
	br.fillBitWindow(17)
}

// bitsUnmasked returns the raw accumulator value. The number of valid bits
// is given by availBits.
func (br *bitReader) bitsUnmasked() uint64 {
	return br.val
}

// getBits returns n bits from the accumulator without advancing. The
// accumulator is filled first if needed.
func (br *bitReader) getBits(n uint) uint64 {
	br.fillBitWindow(n)
	return br.val & bitMask(n)
}

// dropBits discards the lowest n bits from the accumulator.
// n must be less than 64; in practice it is at most 15 (max Huffman code length).
// The & 63 mask is a no-op on the value but lets the compiler emit a plain
// shift instead of Go's safe-shift sequence (SHRQ+CMPQ+SBBQ+ANDQ → SHRQ).
func (br *bitReader) dropBits(n uint) {
	br.bitPos -= n
	br.val >>= n & 63
}

// takeBits returns the lowest n bits and discards them.
func (br *bitReader) takeBits(n uint) uint64 {
	v := br.val & bitMask(n)
	br.dropBits(n)
	return v
}

// readBits reads up to 24 bits from the input. The accumulator is filled
// first if needed.
func (br *bitReader) readBits(n uint) uint64 {
	br.fillBitWindow(n)
	return br.takeBits(n)
}

// readBits32 reads up to 32 bits from the input. On 64-bit platforms this
// is identical to readBits.
func (br *bitReader) readBits32(n uint) uint64 {
	br.fillBitWindow(n)
	return br.takeBits(n)
}

// pullByte loads one byte from input into the accumulator. Returns false
// if no input is available.
func (br *bitReader) pullByte() bool {
	if br.pos >= br.inputLen {
		return false
	}
	br.val |= uint64(br.input[br.pos]) << br.bitPos
	br.bitPos += 8
	br.pos++
	return true
}

// safeGetBits peeks at n bits, pulling bytes as needed. Returns the value
// and true on success, or 0 and false if insufficient input.
func (br *bitReader) safeGetBits(n uint) (uint64, bool) {
	for br.bitPos < n {
		if !br.pullByte() {
			return 0, false
		}
	}
	return br.val & bitMask(n), true
}

// safeReadBits reads up to 24 bits, pulling bytes as needed. Returns the
// value and true on success, or 0 and false if insufficient input.
func (br *bitReader) safeReadBits(n uint) (uint64, bool) {
	for br.bitPos < n {
		if !br.pullByte() {
			return 0, false
		}
	}
	return br.takeBits(n), true
}

// safeReadBits32 reads up to 32 bits, pulling bytes as needed. On 64-bit
// platforms this is identical to safeReadBits for n <= 24; for n > 24 it
// falls back to safeReadBits32Slow.
func (br *bitReader) safeReadBits32(n uint) (uint64, bool) {
	if n <= 24 {
		return br.safeReadBits(n)
	}
	return br.safeReadBits32Slow(n)
}

// safeReadBits32Slow handles safe reads of more than 24 bits by splitting
// into two 16-bit reads with state save/restore on failure.
func (br *bitReader) safeReadBits32Slow(n uint) (uint64, bool) {
	memento := br.saveState()
	lowVal, ok := br.safeReadBits(16)
	if !ok {
		br.restoreState(memento)
		return 0, false
	}
	highVal, ok := br.safeReadBits(n - 16)
	if !ok {
		br.restoreState(memento)
		return 0, false
	}
	return lowVal | (highVal << 16), true
}

// normalize clears any bits above bitPos in the accumulator (Spectre
// mitigation for cases where bytes are skipped without being loaded).
func (br *bitReader) normalize() {
	if br.bitPos < 64 {
		br.val &= bitMask(br.bitPos)
	}
}

// unload pushes unused whole bytes from the accumulator back into the
// unconsumed input by decrementing pos.
func (br *bitReader) unload() {
	unusedBytes := br.bitPos >> 3
	unusedBits := unusedBytes << 3
	if unusedBytes > 0 {
		br.pos -= int(unusedBytes)
	}
	br.bitPos -= unusedBits
	br.normalize()
}

// jumpToByteBoundary advances to the next byte boundary and verifies that
// any skipped padding bits are zero. Returns false if non-zero padding is
// found.
func (br *bitReader) jumpToByteBoundary() bool {
	padBitsCount := br.bitPos & 0x7
	padBits := uint64(0)
	if padBitsCount != 0 {
		padBits = br.takeBits(padBitsCount)
	}
	br.normalize()
	return padBits == 0
}

// dropBytes advances the byte position by n. The accumulator must be empty.
func (br *bitReader) dropBytes(n int) {
	br.pos += n
}

// copyBytes copies num bytes from the reader to dest. Drains the
// accumulator first, then copies directly from input.
func (br *bitReader) copyBytes(dest []byte) {
	i := 0
	num := len(dest)
	for br.bitPos >= 8 && i < num {
		dest[i] = byte(br.val)
		br.dropBits(8)
		i++
	}
	br.normalize()
	if i < num {
		copy(dest[i:], br.input[br.pos:br.pos+num-i])
		br.pos += num - i
	}
}

// bitMask returns a mask with the lowest n bits set.
// Uses shift-subtract instead of a table lookup to avoid a memory load
// and reduce L1 cache pressure in the decode hot path.
// The & 63 lets the compiler prove the shift is in-range, eliminating
// the Go overflow check. All callers pass n ≤ 24.
func bitMask(n uint) uint64 {
	return (uint64(1) << (n & 63)) - 1
}
