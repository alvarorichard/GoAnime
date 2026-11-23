// Backward match descriptor for the H10 binary tree hasher.
//
// A backwardMatch encodes a (distance, length) pair found by the H10 hasher's
// findAllMatches method. The length and an optional length-code are packed into
// a single uint32 so that the match array stays compact and allocation-free.

package encoder

// backwardMatch describes a single backward reference found by the H10
// binary tree hasher. Matches are returned sorted by strictly increasing
// length and non-strictly increasing distance.
type backwardMatch struct {
	distance      uint32
	lengthAndCode uint32 // length << 5 | lengthCode, or 0 when lengthCode == length
}

// newBackwardMatch creates a match with the given backward distance and
// length. The length code equals the length (no dictionary transform).
func newBackwardMatch(distance, length uint) backwardMatch {
	return backwardMatch{
		distance:      uint32(distance),
		lengthAndCode: uint32(length << 5),
	}
}

// newDictionaryBackwardMatch creates a match against a static dictionary
// entry. The length code may differ from the length when a dictionary
// transform omits trailing bytes.
func newDictionaryBackwardMatch(distance, length, lenCode uint) backwardMatch {
	code := lenCode
	if lenCode == length {
		code = 0
	}
	return backwardMatch{
		distance:      uint32(distance),
		lengthAndCode: uint32((length << 5) | code),
	}
}

// matchLength returns the match length in bytes.
func (m backwardMatch) matchLength() uint {
	return uint(m.lengthAndCode >> 5)
}

// matchLengthCode returns the length code, which equals the match length
// for normal matches but may differ for dictionary matches with transforms.
func (m backwardMatch) matchLengthCode() uint {
	if code := uint(m.lengthAndCode) & 31; code != 0 {
		return code
	}
	return uint(m.lengthAndCode >> 5)
}
