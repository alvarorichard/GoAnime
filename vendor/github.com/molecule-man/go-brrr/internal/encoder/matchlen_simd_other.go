//go:build (!amd64 || purego) && !(ppc64 || s390x || mips || mips64)

package encoder

import (
	"math/bits"
	"unsafe"
)

func matchLenSIMD(dataPtr unsafe.Pointer, a, b uint, limit int) int {
	i := 0
	for ; i <= limit-8; i += 8 {
		pa := (*uint64)(unsafe.Add(dataPtr, a+uint(i)))
		pb := (*uint64)(unsafe.Add(dataPtr, b+uint(i)))
		xor := *pa ^ *pb
		if xor != 0 {
			return i + bits.TrailingZeros64(xor)/8
		}
	}
	for ; i < limit &&
		*(*byte)(unsafe.Add(dataPtr, a+uint(i))) == *(*byte)(unsafe.Add(dataPtr, b+uint(i))); i++ {
	}
	return i
}
