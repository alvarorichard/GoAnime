package encoder

import (
	"bytes"
	"math/bits"
)

// Byte-level prefix matching for LZ77.

// matchLenLongBlock is the block compared per bytes.Equal call, and the span
// below which matchLenAt is the better choice: this function is not inlinable,
// so a short span would pay a call to reach the same scalar loop.
const matchLenLongBlock = 4096

// matchLen returns the number of bytes common to the start of a and b,
// examining at most limit bytes. Both slices must be at least limit bytes long.
func matchLen(a, b []byte, limit int) int {
	i := 0
	for ; i <= limit-8; i += 8 {
		xor := loadU64LE(a, uint(i)) ^ loadU64LE(b, uint(i))
		if xor != 0 {
			return i + bits.TrailingZeros64(xor)/8
		}
	}

	for ; i < limit && a[i] == b[i]; i++ {
	}

	return i
}

// matchLenAt compares data[a:] against data[b:] for up to limit bytes using
// unsafe loads, avoiding sub-slice creation and its bounds checks.
func matchLenAt(data []byte, a, b uint, limit int) int {
	i := 0
	for ; i <= limit-8; i += 8 {
		xor := loadU64LE(data, a+uint(i)) ^ loadU64LE(data, b+uint(i))
		if xor != 0 {
			return i + bits.TrailingZeros64(xor)/8
		}
	}

	for ; i < limit && data[a+uint(i)] == data[b+uint(i)]; i++ {
	}

	return i
}

// matchLenAtNoInline compares data[a:] against data[b:] for up to limit bytes.
// Equivalent to matchLen(data[a:], data[b:], limit) but avoids sub-slice
// creation at the call site, reducing caller code size.
//
// The tail loop uses loadByte (unsafe) to avoid bounds-check calls, keeping
// this a leaf function and eliminating frame-pointer save/restore overhead.
//
//go:noinline
func matchLenAtNoInline(data []byte, a, b uint, limit int) int {
	i := 0
	for ; i <= limit-8; i += 8 {
		xor := loadU64LE(data, a+uint(i)) ^ loadU64LE(data, b+uint(i))
		if xor != 0 {
			return i + bits.TrailingZeros64(xor)/8
		}
	}

	for ; i < limit && loadByte(data, a+uint(i)) == loadByte(data, b+uint(i)); i++ {
	}

	return i
}

// matchLenAtLong returns the number of bytes common to data[a:] and data[b:],
// examining at most limit bytes, for spans long enough that a full match is the
// common case.
//
// extendLastCommand chunks its copy to the ring-buffer end, so it arrives here
// with limit in the tens or hundreds of kilobytes and, on repetitive input,
// matches all of it. bytes.Equal lowers to the runtime's memequal, which is
// vectorised far wider than the eight bytes per iteration matchLenAt manages;
// only the first block that differs is scanned byte-wise.
//
// matchLenAt stays the right choice everywhere else: its ~90 hasher call sites
// see a mean match of roughly sixteen bytes, where this loses to the extra call.
func matchLenAtLong(data []byte, a, b uint, limit int) int {
	i := 0
	for i+matchLenLongBlock <= limit {
		if !bytes.Equal(data[a+uint(i):a+uint(i)+matchLenLongBlock], data[b+uint(i):b+uint(i)+matchLenLongBlock]) {
			break
		}
		i += matchLenLongBlock
	}
	return i + matchLenAt(data, a+uint(i), b+uint(i), limit-i)
}
