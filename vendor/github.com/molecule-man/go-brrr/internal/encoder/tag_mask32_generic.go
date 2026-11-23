//go:build !amd64 || purego

package encoder

// tagMask32 sets bit k when tags[k] equals tag.
func tagMask32(tags *[32]uint8, tag uint8) uint32 {
	splat := uint64(tag) * 0x0101010101010101
	return tagEqualMask8(loadU64LE(tags[:], 0)^splat) |
		tagEqualMask8(loadU64LE(tags[:], 8)^splat)<<8 |
		tagEqualMask8(loadU64LE(tags[:], 16)^splat)<<16 |
		tagEqualMask8(loadU64LE(tags[:], 24)^splat)<<24
}
