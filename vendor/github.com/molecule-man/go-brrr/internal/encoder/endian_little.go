//go:build !(ppc64 || s390x || mips || mips64)

package encoder

import "unsafe"

func storeU64LEPtr(p unsafe.Pointer, v uint64) { *(*uint64)(p) = v }

// Match the byte order of native probe loads.
func le32(v uint32) uint32 { return v }
