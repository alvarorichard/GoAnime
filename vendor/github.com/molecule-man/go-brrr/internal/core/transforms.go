// RFC 7932 transforms for static dictionary words.

package core

import "unsafe"

// Transform types from RFC 7932.
const (
	TransformIdentity       = iota // 0
	TransformOmitLast1             // 1
	TransformOmitLast2             // 2
	TransformOmitLast3             // 3
	TransformOmitLast4             // 4
	TransformOmitLast5             // 5
	TransformOmitLast6             // 6
	TransformOmitLast7             // 7
	TransformOmitLast8             // 8
	TransformOmitLast9             // 9
	TransformUppercaseFirst        // 10
	TransformUppercaseAll          // 11
	TransformOmitFirst1            // 12
	TransformOmitFirst2            // 13
	TransformOmitFirst3            // 14
	TransformOmitFirst4            // 15
	TransformOmitFirst5            // 16
	TransformOmitFirst6            // 17
	TransformOmitFirst7            // 18
	TransformOmitFirst8            // 19
	TransformOmitFirst9            // 20
	TransformShiftFirst            // 21
	TransformShiftAll              // 22
)

// NumTransforms is the count of static dictionary transforms defined by
// RFC 7932 Section 8.
const NumTransforms = 121

// Prefix and suffix strings used by transforms. Indices are referenced by
// TransformTriplets.
var transformPrefixSuffix = [50]string{
	" ",        // 0
	", ",       // 1
	" of the ", // 2
	" of ",     // 3
	"s ",       // 4
	".",        // 5
	" and ",    // 6
	" in ",     // 7
	"\"",       // 8
	" to ",     // 9
	"\">",      // 10
	"\n",       // 11
	". ",       // 12
	"]",        // 13
	" for ",    // 14
	" a ",      // 15
	" that ",   // 16
	"'",        // 17
	" with ",   // 18
	" from ",   // 19
	" by ",     // 20
	"(",        // 21
	". The ",   // 22
	" on ",     // 23
	" as ",     // 24
	" is ",     // 25
	"ing ",     // 26
	"\n\t",     // 27
	":",        // 28
	"ed ",      // 29
	"=\"",      // 30
	" at ",     // 31
	"ly ",      // 32
	",",        // 33
	"='",       // 34
	".com/",    // 35
	". This ",  // 36
	" not ",    // 37
	"er ",      // 38
	"al ",      // 39
	"ful ",     // 40
	"ive ",     // 41
	"less ",    // 42
	"est ",     // 43
	"ize ",     // 44
	"\xc2\xa0", // 45 (U+00A0 non-breaking space)
	"ous ",     // 46
	" the ",    // 47
	"e ",       // 48
	"",         // 49
}

// TransformTriplets holds [prefix_id, transform_type, suffix_id] for each of
// the 121 RFC 7932 transforms.
var TransformTriplets = [NumTransforms * 3]byte{
	49, TransformIdentity, 49, // 0
	49, TransformIdentity, 0, // 1
	0, TransformIdentity, 0, // 2
	49, TransformOmitFirst1, 49, // 3
	49, TransformUppercaseFirst, 0, // 4
	49, TransformIdentity, 47, // 5
	0, TransformIdentity, 49, // 6
	4, TransformIdentity, 0, // 7
	49, TransformIdentity, 3, // 8
	49, TransformUppercaseFirst, 49, // 9
	49, TransformIdentity, 6, // 10
	49, TransformOmitFirst2, 49, // 11
	49, TransformOmitLast1, 49, // 12
	1, TransformIdentity, 0, // 13
	49, TransformIdentity, 1, // 14
	0, TransformUppercaseFirst, 0, // 15
	49, TransformIdentity, 7, // 16
	49, TransformIdentity, 9, // 17
	48, TransformIdentity, 0, // 18
	49, TransformIdentity, 8, // 19
	49, TransformIdentity, 5, // 20
	49, TransformIdentity, 10, // 21
	49, TransformIdentity, 11, // 22
	49, TransformOmitLast3, 49, // 23
	49, TransformIdentity, 13, // 24
	49, TransformIdentity, 14, // 25
	49, TransformOmitFirst3, 49, // 26
	49, TransformOmitLast2, 49, // 27
	49, TransformIdentity, 15, // 28
	49, TransformIdentity, 16, // 29
	0, TransformUppercaseFirst, 49, // 30
	49, TransformIdentity, 12, // 31
	5, TransformIdentity, 49, // 32
	0, TransformIdentity, 1, // 33
	49, TransformOmitFirst4, 49, // 34
	49, TransformIdentity, 18, // 35
	49, TransformIdentity, 17, // 36
	49, TransformIdentity, 19, // 37
	49, TransformIdentity, 20, // 38
	49, TransformOmitFirst5, 49, // 39
	49, TransformOmitFirst6, 49, // 40
	47, TransformIdentity, 49, // 41
	49, TransformOmitLast4, 49, // 42
	49, TransformIdentity, 22, // 43
	49, TransformUppercaseAll, 49, // 44
	49, TransformIdentity, 23, // 45
	49, TransformIdentity, 24, // 46
	49, TransformIdentity, 25, // 47
	49, TransformOmitLast7, 49, // 48
	49, TransformOmitLast1, 26, // 49
	49, TransformIdentity, 27, // 50
	49, TransformIdentity, 28, // 51
	0, TransformIdentity, 12, // 52
	49, TransformIdentity, 29, // 53
	49, TransformOmitFirst9, 49, // 54
	49, TransformOmitFirst7, 49, // 55
	49, TransformOmitLast6, 49, // 56
	49, TransformIdentity, 21, // 57
	49, TransformUppercaseFirst, 1, // 58
	49, TransformOmitLast8, 49, // 59
	49, TransformIdentity, 31, // 60
	49, TransformIdentity, 32, // 61
	47, TransformIdentity, 3, // 62
	49, TransformOmitLast5, 49, // 63
	49, TransformOmitLast9, 49, // 64
	0, TransformUppercaseFirst, 1, // 65
	49, TransformUppercaseFirst, 8, // 66
	5, TransformIdentity, 21, // 67
	49, TransformUppercaseAll, 0, // 68
	49, TransformUppercaseFirst, 10, // 69
	49, TransformIdentity, 30, // 70
	0, TransformIdentity, 5, // 71
	35, TransformIdentity, 49, // 72
	47, TransformIdentity, 2, // 73
	49, TransformUppercaseFirst, 17, // 74
	49, TransformIdentity, 36, // 75
	49, TransformIdentity, 33, // 76
	5, TransformIdentity, 0, // 77
	49, TransformUppercaseFirst, 21, // 78
	49, TransformUppercaseFirst, 5, // 79
	49, TransformIdentity, 37, // 80
	0, TransformIdentity, 30, // 81
	49, TransformIdentity, 38, // 82
	0, TransformUppercaseAll, 0, // 83
	49, TransformIdentity, 39, // 84
	0, TransformUppercaseAll, 49, // 85
	49, TransformIdentity, 34, // 86
	49, TransformUppercaseAll, 8, // 87
	49, TransformUppercaseFirst, 12, // 88
	0, TransformIdentity, 21, // 89
	49, TransformIdentity, 40, // 90
	0, TransformUppercaseFirst, 12, // 91
	49, TransformIdentity, 41, // 92
	49, TransformIdentity, 42, // 93
	49, TransformUppercaseAll, 17, // 94
	49, TransformIdentity, 43, // 95
	0, TransformUppercaseFirst, 5, // 96
	49, TransformUppercaseAll, 10, // 97
	0, TransformIdentity, 34, // 98
	49, TransformUppercaseFirst, 33, // 99
	49, TransformIdentity, 44, // 100
	49, TransformUppercaseAll, 5, // 101
	45, TransformIdentity, 49, // 102
	0, TransformIdentity, 33, // 103
	49, TransformUppercaseFirst, 30, // 104
	49, TransformUppercaseAll, 30, // 105
	49, TransformIdentity, 46, // 106
	49, TransformUppercaseAll, 1, // 107
	49, TransformUppercaseFirst, 34, // 108
	0, TransformUppercaseFirst, 33, // 109
	0, TransformUppercaseAll, 30, // 110
	0, TransformUppercaseAll, 1, // 111
	49, TransformUppercaseAll, 33, // 112
	49, TransformUppercaseAll, 21, // 113
	49, TransformUppercaseAll, 12, // 114
	0, TransformUppercaseAll, 5, // 115
	49, TransformUppercaseAll, 34, // 116
	0, TransformUppercaseAll, 12, // 117
	0, TransformUppercaseFirst, 30, // 118
	0, TransformUppercaseAll, 34, // 119
	0, TransformUppercaseFirst, 34, // 120
}

// TransformCutOffs holds the indices of transforms ["", omit-last-N, ""]
// for N=0..9, where N=0 means identity. Used by the encoder for fast
// dictionary matching.
var TransformCutOffs = [10]int16{0, 12, 27, 23, 42, 63, 56, 48, 59, 64}

// TransformDictionaryWord applies transform transformIdx to word, writing the
// result into dst. Returns the number of bytes written. dst must be large
// enough to hold the result (word length + longest prefix + longest suffix).
func TransformDictionaryWord(dst []byte, word string, transformIdx int) int {
	prefix := transformPrefixSuffix[TransformTriplets[transformIdx*3]]
	t := TransformTriplets[transformIdx*3+1]
	suffix := transformPrefixSuffix[TransformTriplets[transformIdx*3+2]]

	// Use unsafe copies to avoid runtime.memmove overhead for small data.
	// Safety: dst is backed by the ring buffer which has ≥542 bytes of
	// slack, and word comes from the static dictionary (≥120 KB). All
	// reads and writes within 32 bytes of the starting offsets are in
	// bounds.
	dstBase := unsafe.Pointer(unsafe.SliceData(dst))

	idx := len(prefix)
	if idx > 0 {
		*(*[8]byte)(dstBase) = *(*[8]byte)(unsafe.Pointer(unsafe.StringData(prefix)))
	}

	wordLen := len(word)
	wordStart := 0
	if t <= TransformOmitLast9 {
		wordLen -= int(t)
		if wordLen < 0 {
			wordLen = 0
		}
	} else if t >= TransformOmitFirst1 && t <= TransformOmitFirst9 {
		skip := min(int(t)-(TransformOmitFirst1-1), wordLen)
		wordStart = skip
		wordLen -= skip
	}

	if wordLen > 0 {
		wordSrc := unsafe.Pointer(unsafe.StringData(word))
		dstWord := unsafe.Add(dstBase, idx)
		*(*[16]byte)(dstWord) = *(*[16]byte)(unsafe.Add(wordSrc, wordStart))
		if wordLen > 16 {
			*(*[16]byte)(unsafe.Add(dstWord, 16)) = *(*[16]byte)(unsafe.Add(wordSrc, wordStart+16))
		}
	}
	idx += wordLen

	switch t {
	case TransformUppercaseFirst:
		toUpperCase(dst[idx-wordLen:])
	case TransformUppercaseAll:
		p := dst[idx-wordLen:]
		remaining := wordLen
		for remaining > 0 {
			step := toUpperCase(p)
			p = p[step:]
			remaining -= step
		}
	}

	suffixLen := len(suffix)
	if suffixLen > 0 {
		*(*[8]byte)(unsafe.Add(dstBase, idx)) = *(*[8]byte)(unsafe.Pointer(unsafe.StringData(suffix)))
	}
	idx += suffixLen
	return idx
}

// toUpperCase uppercases the first UTF-8 character in p using the simplified
// model from RFC 7932. Returns the byte length of the character processed.
func toUpperCase(p []byte) int {
	if p[0] < 0xC0 {
		if p[0] >= 'a' && p[0] <= 'z' {
			p[0] ^= 32
		}
		return 1
	}
	// Simplified 2-byte UTF-8 uppercase: flip bit 5 of the second byte.
	if p[0] < 0xE0 {
		p[1] ^= 32
		return 2
	}
	// Arbitrary transform for 3-byte UTF-8: XOR the third byte with 5.
	p[2] ^= 5
	return 3
}

// shiftUTF8 applies a scalar shift to the first UTF-8 character in word.
// Returns the byte length of the character. Used by SHIFT_FIRST and SHIFT_ALL
// transforms, which are not present in the 121 static RFC 7932 transforms but
// are part of the spec.
func shiftUTF8(word []byte, wordLen int, param uint16) int {
	// Limited sign extension: scalar < (1 << 24).
	scalar := uint32(param&0x7FFF) + (0x1000000 - uint32(param&0x8000))
	switch {
	case word[0] < 0x80:
		// 1-byte / ASCII / 7-bit scalar.
		scalar += uint32(word[0])
		word[0] = byte(scalar & 0x7F)
		return 1
	case word[0] < 0xC0:
		// Continuation byte.
		return 1
	case word[0] < 0xE0:
		// 2-byte / 11-bit scalar.
		if wordLen < 2 {
			return 1
		}
		scalar += uint32(word[1]&0x3F) | uint32(word[0]&0x1F)<<6
		word[0] = byte(0xC0 | ((scalar >> 6) & 0x1F))
		word[1] = byte(uint32(word[1]&0xC0) | (scalar & 0x3F))
		return 2
	case word[0] < 0xF0:
		// 3-byte / 16-bit scalar.
		if wordLen < 3 {
			return wordLen
		}
		scalar += uint32(word[2]&0x3F) | uint32(word[1]&0x3F)<<6 | uint32(word[0]&0x0F)<<12
		word[0] = byte(0xE0 | ((scalar >> 12) & 0x0F))
		word[1] = byte(uint32(word[1]&0xC0) | ((scalar >> 6) & 0x3F))
		word[2] = byte(uint32(word[2]&0xC0) | (scalar & 0x3F))
		return 3
	case word[0] < 0xF8:
		// 4-byte / 21-bit scalar.
		if wordLen < 4 {
			return wordLen
		}
		scalar += uint32(word[3]&0x3F) | uint32(word[2]&0x3F)<<6 |
			uint32(word[1]&0x3F)<<12 | uint32(word[0]&0x07)<<18
		word[0] = byte(0xF0 | ((scalar >> 18) & 0x07))
		word[1] = byte(uint32(word[1]&0xC0) | ((scalar >> 12) & 0x3F))
		word[2] = byte(uint32(word[2]&0xC0) | ((scalar >> 6) & 0x3F))
		word[3] = byte(uint32(word[3]&0xC0) | (scalar & 0x3F))
		return 4
	default:
		return 1
	}
}
