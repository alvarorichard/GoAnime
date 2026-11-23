// Shared hasher types, constants, and scoring functions.

package encoder

import (
	"math/bits"
)

// Hash configuration constants shared by H2, H3, and similar hashers.
const (
	bucketBits = 16
	bucketSize = 1 << bucketBits // 65536
	bucketMask = bucketSize - 1
	hashLen    = 5

	// hashTypeLength is the minimum number of bytes required to compute
	// a hash and verify a match (StoreLookahead in the C reference).
	hashTypeLength = 8

	// hashMul64 is the 64-bit hash multiplier from hash_base.h.
	hashMul64        = 0x1FE35A7BD3579BD3
	hashMul64Shifted = 0x7BD3579BD3000000 // hashMul64 << 24, truncated to 64 bits.
)

// Backward reference scoring constants from hash.h.
const (
	literalByteScore   = 135
	distanceBitPenalty = 30
	scoreBase          = distanceBitPenalty * 8 * 8 // 8 bytes (64-bit) → 1920
	minScore           = scoreBase + 100            // 2020
)

// hasherCommon holds state shared by all hashers.
type hasherCommon struct {
	ready bool // false until reset() zeroes the table; cleared on 32-bit position wrap
}

// hasherSearchResult holds the best match found by findLongestMatch.
type hasherSearchResult struct {
	len          uint
	distance     uint
	score        uint
	lenCodeDelta int
}

// streamHasher is the interface for all streaming hashers (H2–H54, H5, H5b5,
// H5b6, H5b7, H5b8, H6, H6u, H6b5, H6b5u, H6b6, H6b7, H6b8, H40, H41, H42,
// H10). The hot match-finding loop lives inside createBackwardReferences so that
// findLongestMatch/store/storeRange remain direct (non-virtual) calls — only
// the per-metablock entry point is virtual.
type streamHasher interface {
	common() *hasherCommon
	reset(oneShot bool, inputSize uint, data []byte)
	stitchToPreviousBlock(numBytes, position uint, ringBuffer []byte, ringBufferMask uint)
	createBackwardReferences(s *encodeState, bytes, wrappedPos uint32)
}

// hashBytes computes a 16-bit hash from 5 bytes at data[offset:offset+8].
// The caller must ensure len(data) >= offset+8.
func hashBytes(data []byte, offset uint) uint32 {
	v := loadU64LE(data, offset)
	h := v * hashMul64Shifted
	return uint32(h >> (64 - bucketBits))
}

// backwardReferenceScore computes a score for a match of copyLength bytes
// at the given backward distance. Higher is better.
func backwardReferenceScore(copyLength, backwardReferenceOffset uint) uint {
	return scoreBase + literalByteScore*copyLength -
		distanceBitPenalty*uint(bits.Len(backwardReferenceOffset)-1)
}

// backwardReferenceScoreUsingLastDistance computes a score for a match
// that reuses the most recent backward distance (cache[0]).
func backwardReferenceScoreUsingLastDistance(copyLength uint) uint {
	return literalByteScore*copyLength + scoreBase + 15
}
