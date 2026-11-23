package encoder

// tagEqualMask8 returns a mask with bit k set when byte k of c is zero.
// With c = tags ^ splat(tag), bit k marks tags[k] == tag. Port of the
// portable path of C GetMatchingTagMask.
func tagEqualMask8(c uint64) uint32 {
	const (
		x01 = 0x0101010101010101
		x80 = x01 << 7
		// Moves the low bit of byte k to bit 56+k.
		gather = 0x0102040810204080
	)
	zero := ^((((c | x80) - x01) | c) & x80)
	return uint32((zero & x80 >> 7) * gather >> 56)
}
