//go:build !(ppc64 || s390x || mips || mips64)

package core

// HuffmanCode is a single entry in a Huffman lookup table.
// Its native uint32 form has Bits in the low byte and Value in the high bytes.
type HuffmanCode struct {
	Bits  byte
	Value uint16
}
