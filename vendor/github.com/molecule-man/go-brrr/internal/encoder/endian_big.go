//go:build ppc64 || s390x || mips || mips64

package encoder

import (
	"math/bits"
	"unsafe"
)

// Byte stores permit an unaligned destination.
func storeU64LEPtr(p unsafe.Pointer, v uint64) {
	b := (*[8]byte)(p)
	b[0], b[1], b[2], b[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	b[4], b[5], b[6], b[7] = byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56)
}

// Match the byte order of native probe loads.
func le32(v uint32) uint32 { return bits.ReverseBytes32(v) }
