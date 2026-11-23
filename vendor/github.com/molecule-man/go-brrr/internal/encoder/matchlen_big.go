//go:build ppc64 || s390x || mips || mips64

package encoder

import (
	"math/bits"
	"unsafe"
)

func matchLenSIMD(dataPtr unsafe.Pointer, a, b uint, limit int) int {
	i := 0
	for ; i <= limit-8; i += 8 {
		xor := loadU64LEPtr(unsafe.Add(dataPtr, a+uint(i))) ^ loadU64LEPtr(unsafe.Add(dataPtr, b+uint(i)))
		if xor != 0 {
			return i + bits.TrailingZeros64(xor)/8
		}
	}
	for ; i < limit &&
		*(*byte)(unsafe.Add(dataPtr, a+uint(i))) == *(*byte)(unsafe.Add(dataPtr, b+uint(i))); i++ {
	}
	return i
}
