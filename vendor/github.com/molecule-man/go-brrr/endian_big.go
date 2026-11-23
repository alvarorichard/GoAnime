//go:build ppc64 || s390x || mips || mips64

package brrr

import "math/bits"

// Convert native loads of stream bytes to little-endian values.
func le32(v uint32) uint32 { return bits.ReverseBytes32(v) }

func le64(v uint64) uint64 { return bits.ReverseBytes64(v) }
