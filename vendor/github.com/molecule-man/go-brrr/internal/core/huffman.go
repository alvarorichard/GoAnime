package core

// Huffman (prefix code) construction, encoding, and decoding table building.
//
// A symbol is an element of an alphabet. Each alphabet has its own Huffman tree
// (prefix code), and decoding one entry from that tree produces one symbol.
// The RFC uses "symbol" and "code" interchangeably; the meaning depends on
// which alphabet is in context:
//
//   - Literal symbols:            0–255   (raw bytes)
//   - Insert-and-copy symbols:    0–703   (Section 5)
//   - Distance symbols:           0–N     (Section 4, N depends on parameters)
//   - Block count symbols:        0–25    (Section 6.3)
//   - Block type symbols:         0–N     (Section 6, N = number of block types + 1)
//   - Code length symbols:        0–17    (Section 3.5, used to build other trees)
//
// Some symbols map directly to values (e.g., literal symbols are byte values).
// Others are intermediary codes that require extra bits to produce a final value
// (see coderange.go).
//
// Brotli transmits Huffman trees ("prefix codes") using a three-layer encoding.
// The layers, from outermost (closest to raw bits) to innermost (used on data):
//
//	Layer 0: Fixed code (hardcoded, never transmitted)
//	    Encodes values 0-5, which represent possible code lengths
//	    for code length alphabet symbols.
//
//	        Value   Bits
//	          0       00
//	          1     0111
//	          2      011
//	          3       01
//	          4       10
//	          5     1111
//
//	Layer 1: Code length alphabet (symbols 0-17)
//	    Symbols 0-15 are literal code lengths.
//	    Symbol 16 repeats the previous non-zero code length (RLE).
//	    Symbol 17 repeats zero (RLE for unused symbols).
//	    Each symbol's own code length (0-5) is transmitted using Layer 0.
//
//	Layer 2: Data alphabet (literals, insert-and-copy, distances)
//	    Each symbol's code length (0-15) is transmitted using the
//	    Huffman tree built from Layer 1.

import (
	"math/bits"
	"unsafe"
)

// Huffman coding limits from the brotli specification.
const (
	HuffmanMaxCodeLength           = 15
	HuffmanMaxCodeLengthCodeLength = 5
)

// reverseBitsWidth is the bit width used for canonical-code reversal (8 = one byte).
const reverseBitsWidth = 8

// reverseBitsHigh is the highest bit in a reverseBitsWidth-wide field.
const reverseBitsHigh = uint64(1) << (reverseBitsWidth - 1)

// SymbolList provides access to a uint16 slice with a base offset,
// supporting the negative-index linked-list pattern used during Huffman table
// construction.
type SymbolList struct {
	Storage []uint16
	Offset  int
}

func (sl SymbolList) get(i int) uint16 {
	return *(*uint16)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(sl.Storage)), uintptr(sl.Offset+i)*2))
}

// replicateValue fills table[0], table[step], table[2*step], ..., table[end-step] with code.
// Uses unsafe to pack the HuffmanCode into a single uint32 store per iteration,
// avoiding separate byte+uint16 stores and per-iteration bounds checks.
func replicateValue(table []HuffmanCode, code HuffmanCode, step, end int) {
	base := unsafe.Pointer(unsafe.SliceData(table))
	packed := uint32(code.Bits) | uint32(code.Value)<<16
	for {
		end -= step
		*(*uint32)(unsafe.Add(base, uintptr(end)*4)) = packed
		if end <= 0 {
			break
		}
	}
}

// nextTableBitSize returns the width of the next 2nd-level table.
// count is the histogram of bit lengths for remaining symbols,
// length is the code length of the next processed symbol.
func nextTableBitSize(count []uint16, length, rootBits int) int {
	left := 1 << uint(length-rootBits)
	for length < HuffmanMaxCodeLength {
		left -= int(count[length])
		if left <= 0 {
			break
		}
		length++
		left <<= 1
	}
	return length - rootBits
}

// BuildCodeLengthsHuffmanTable populates table with the Huffman lookup
// entries for the code-length alphabet using the given per-symbol code
// lengths and length histogram.
func BuildCodeLengthsHuffmanTable(table []HuffmanCode, codeLengths []byte, count []uint16) {
	assert(HuffmanMaxCodeLengthCodeLength <= reverseBitsWidth)

	var sorted [AlphabetSizeCodeLengths]int
	var offset [HuffmanMaxCodeLengthCodeLength + 1]int

	// Generate offsets into sorted symbol table by code length.
	sym := -1
	for bitLen := 1; bitLen <= HuffmanMaxCodeLengthCodeLength; bitLen++ {
		sym += int(count[bitLen])
		offset[bitLen] = sym
	}

	// Symbols with code length 0 are placed after all other symbols.
	offset[0] = AlphabetSizeCodeLengths - 1

	// Sort symbols by length, by symbol order within each length.
	sym = AlphabetSizeCodeLengths
	for {
		for range 6 {
			sym--
			sorted[offset[codeLengths[sym]]] = sym
			offset[codeLengths[sym]]--
		}
		if sym == 0 {
			break
		}
	}

	tableSize := 1 << HuffmanMaxCodeLengthCodeLength

	// Special case: all symbols but one have 0 code length.
	if offset[0] == 0 {
		code := HuffmanCode{Bits: 0, Value: uint16(sorted[0])}
		table[0] = code
		for i := 1; i < tableSize; i *= 2 {
			copy(table[i:tableSize], table[:i])
		}
		return
	}

	// Fill in table.
	key := uint64(0)
	keyStep := reverseBitsHigh
	sym = 0
	step := 2
	for bitLen := 1; bitLen <= HuffmanMaxCodeLengthCodeLength; bitLen++ {
		for n := int(count[bitLen]); n > 0; n-- {
			code := HuffmanCode{Bits: byte(bitLen), Value: uint16(sorted[sym])}
			sym++
			replicateValue(table[bits.Reverse8(byte(key)):], code, step, tableSize)
			key += keyStep
		}
		step <<= 1
		keyStep >>= 1
	}
}

// BuildHuffmanTable populates rootTable with a two-level lookup table for
// a Huffman code described by symbols (per-length linked lists) and count
// (length histogram). Returns the total number of entries written.
func BuildHuffmanTable(rootTable []HuffmanCode, rootBits int, symbols SymbolList, count []uint16) uint32 {
	assert(rootBits <= reverseBitsWidth)
	assert(HuffmanMaxCodeLength-rootBits <= reverseBitsWidth)

	maxLength := -1
	for symbols.get(maxLength) == 0xFFFF {
		maxLength--
	}
	maxLength += HuffmanMaxCodeLength + 1

	table := rootTable
	tableBits := rootBits
	tableSize := 1 << uint(tableBits)
	totalSize := tableSize

	// Reduce the root table size if possible, and create repetitions by copy.
	if tableBits > maxLength {
		tableBits = maxLength
		tableSize = 1 << uint(tableBits)
	}

	key := uint64(0)
	keyStep := reverseBitsHigh
	step := 2
	for bitLen := 1; bitLen <= tableBits; bitLen++ {
		sym := bitLen - (HuffmanMaxCodeLength + 1)
		for n := int(count[bitLen]); n > 0; n-- {
			sym = int(symbols.get(sym))
			code := HuffmanCode{Bits: byte(bitLen), Value: uint16(sym)}
			replicateValue(table[bits.Reverse8(byte(key)):], code, step, tableSize)
			key += keyStep
		}
		step <<= 1
		keyStep >>= 1
	}

	// If rootBits != tableBits then replicate to fill the remaining slots.
	for totalSize != tableSize {
		copy(table[tableSize:], table[:uint(tableSize)])
		tableSize <<= 1
	}

	// Fill in 2nd level tables and add pointers to root table.
	keyStep = reverseBitsHigh >> uint(rootBits-1)
	subKey := reverseBitsHigh << 1
	subKeyStep := reverseBitsHigh
	step = 2
	for codeLen := rootBits + 1; codeLen <= maxLength; codeLen++ {
		sym := codeLen - (HuffmanMaxCodeLength + 1)
		for ; count[codeLen] != 0; count[codeLen]-- {
			if subKey == reverseBitsHigh<<1 {
				table = table[tableSize:]
				tableBits = nextTableBitSize(count, codeLen, rootBits)
				tableSize = 1 << uint(tableBits)
				totalSize += tableSize
				subKey = uint64(bits.Reverse8(byte(key)))
				key += keyStep
				rootTable[subKey] = HuffmanCode{
					Bits:  byte(tableBits + rootBits),
					Value: uint16(uint64(uint(-cap(table)+cap(rootTable))) - subKey),
				}
				subKey = 0
			}

			sym = int(symbols.get(sym))
			code := HuffmanCode{Bits: byte(codeLen - rootBits), Value: uint16(sym)}
			replicateValue(table[bits.Reverse8(byte(subKey)):], code, step, tableSize)
			subKey += subKeyStep
		}
		step <<= 1
		subKeyStep >>= 1
	}

	return uint32(totalSize)
}

// BuildSimpleHuffmanTable populates table with the lookup entries for a
// "simple" prefix code (RFC 7932 Section 3.4) carrying numSymbols+1
// symbols listed in val. Returns the number of table entries written.
func BuildSimpleHuffmanTable(table []HuffmanCode, rootBits int, val []uint16, numSymbols uint32) uint32 {
	tableSize := uint32(1)
	goalSize := uint32(1) << uint(rootBits)

	switch numSymbols {
	case 0:
		table[0] = HuffmanCode{Bits: 0, Value: val[0]}

	case 1:
		if val[1] > val[0] {
			table[0] = HuffmanCode{Bits: 1, Value: val[0]}
			table[1] = HuffmanCode{Bits: 1, Value: val[1]}
		} else {
			table[0] = HuffmanCode{Bits: 1, Value: val[1]}
			table[1] = HuffmanCode{Bits: 1, Value: val[0]}
		}
		tableSize = 2

	case 2:
		table[0] = HuffmanCode{Bits: 1, Value: val[0]}
		table[2] = HuffmanCode{Bits: 1, Value: val[0]}
		if val[2] > val[1] {
			table[1] = HuffmanCode{Bits: 2, Value: val[1]}
			table[3] = HuffmanCode{Bits: 2, Value: val[2]}
		} else {
			table[1] = HuffmanCode{Bits: 2, Value: val[2]}
			table[3] = HuffmanCode{Bits: 2, Value: val[1]}
		}
		tableSize = 4

	case 3:
		for i := range 3 {
			for k := i + 1; k < 4; k++ {
				if val[k] < val[i] {
					val[k], val[i] = val[i], val[k]
				}
			}
		}
		table[0] = HuffmanCode{Bits: 2, Value: val[0]}
		table[2] = HuffmanCode{Bits: 2, Value: val[1]}
		table[1] = HuffmanCode{Bits: 2, Value: val[2]}
		table[3] = HuffmanCode{Bits: 2, Value: val[3]}
		tableSize = 4

	case 4:
		if val[3] < val[2] {
			val[3], val[2] = val[2], val[3]
		}
		table[0] = HuffmanCode{Bits: 1, Value: val[0]}
		table[1] = HuffmanCode{Bits: 2, Value: val[1]}
		table[2] = HuffmanCode{Bits: 1, Value: val[0]}
		table[3] = HuffmanCode{Bits: 3, Value: val[2]}
		table[4] = HuffmanCode{Bits: 1, Value: val[0]}
		table[5] = HuffmanCode{Bits: 2, Value: val[1]}
		table[6] = HuffmanCode{Bits: 1, Value: val[0]}
		table[7] = HuffmanCode{Bits: 3, Value: val[3]}
		tableSize = 8
	}

	for tableSize != goalSize {
		copy(table[tableSize:], table[:uint(tableSize)])
		tableSize <<= 1
	}

	return goalSize
}

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}
