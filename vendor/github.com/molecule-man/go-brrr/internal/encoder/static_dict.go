// Static dictionary search for the streaming encoder (quality >= 2).

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// Packed transform IDs for omit-last-N transforms (N=0..9).
// Bits [cut*6 +: 6] hold the 6-bit base for transform N.
const (
	cutoffTransformsCount = 10
	cutoffTransforms      = uint64(0x071B520ADA2D3200)
)

// staticDictHashEntries packs (firstByte, length, wordIndex) for each hash
// bucket so a single load brings all three filter inputs into the same cache
// line. Layout: bits [0:8] firstByte, [8:16] length, [16:32] wordIndex.
// Built at startup from the unpacked arrays.
var staticDictHashEntries [32768]uint32

func init() {
	for i := range staticDictHashEntries {
		staticDictHashEntries[i] = uint32(staticDictHashFirstBytes[i]) |
			uint32(staticDictHashLengths[i])<<8 |
			uint32(staticDictHashWords[i])<<16
	}
}

// hash14 computes a 14-bit hash from 4 bytes for dictionary hash table lookup.
func hash14(data []byte) uint32 {
	return (loadU32LE(data, 0) * hashMul32) >> 18
}

// findMatchLenDict returns the length of the common prefix between data and
// a dictionary word. Both must be at least len(word) bytes.
func findMatchLenDict(data []byte, word string) int {
	for i := range len(word) {
		if data[i] != word[i] {
			return i
		}
	}
	return len(word)
}

// matchStaticDictEntry checks whether the dictionary word at
// (wordLen, wordIndex) matches the input data with a score beating minScore.
// The caller must ensure wordLen <= len(data).
func matchStaticDictEntry(data []byte, wordLen, wordIndex, maxBackward, maxDistance, minScore uint) (hasherSearchResult, bool) {
	offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*wordIndex

	ml := uint(findMatchLenDict(data[:wordLen], core.DictData[offset:offset+wordLen]))
	if ml+cutoffTransformsCount <= wordLen || ml == 0 {
		return hasherSearchResult{}, false
	}

	cut := wordLen - ml
	transformID := (cut << 2) + uint((cutoffTransforms>>(cut*6))&0x3F)
	backward := maxBackward + 1 + wordIndex + (transformID << core.DictSizeBitsByLength[wordLen])
	if backward > maxDistance {
		return hasherSearchResult{}, false
	}

	score := backwardReferenceScore(ml, backward)
	if score < minScore {
		return hasherSearchResult{}, false
	}

	return hasherSearchResult{
		len:          ml,
		lenCodeDelta: int(wordLen) - int(ml),
		distance:     backward,
		score:        score,
	}, true
}

func findMatchLenDictAt(data []byte, pos uint, word string) int {
	for i := range len(word) {
		if loadByte(data, pos+uint(i)) != word[i] {
			return i
		}
	}
	return len(word)
}

func matchStaticDictEntryAt(data []byte, pos, wordLen, wordIndex, maxBackward, minScore uint, out *hasherSearchResult) bool {
	offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*wordIndex

	ml := uint(findMatchLenDictAt(data, pos, core.DictData[offset:offset+wordLen]))
	if ml+cutoffTransformsCount <= wordLen || ml == 0 {
		return false
	}

	cut := wordLen - ml
	transformID := (cut << 2) + uint((cutoffTransforms>>(cut*6))&0x3F)
	backward := maxBackward + 1 + wordIndex + (transformID << core.DictSizeBitsByLength[wordLen])
	if backward > maxBackwardDistance {
		return false
	}

	score := backwardReferenceScore(ml, backward)
	if score < minScore {
		return false
	}

	out.len = ml
	out.lenCodeDelta = int(wordLen) - int(ml)
	out.distance = backward
	out.score = score
	return true
}

// searchStaticDictionary searches the RFC 7932 static dictionary for a match
// at data[0:]. The adaptive heuristic skips the search when matches are rare.
//
// This is the shallow variant (quality 2): it checks one hash entry only.
func searchStaticDictionary(data []byte, maxLength, maxBackward, maxDistance uint,
	dictNumLookups, dictNumMatches *uint, minScore uint) (hasherSearchResult, bool) {
	if *dictNumMatches < (*dictNumLookups >> 7) {
		return hasherSearchResult{}, false
	}

	key := hash14(data) << 1

	*dictNumLookups++
	if data[0] == staticDictHashFirstBytes[key] &&
		staticDictHashLengths[key] != 0 && uint(staticDictHashLengths[key]) <= maxLength {
		if m, ok := matchStaticDictEntry(data, uint(staticDictHashLengths[key]), uint(staticDictHashWords[key]),
			maxBackward, maxDistance, minScore); ok {
			*dictNumMatches++
			return m, true
		}
	}
	return hasherSearchResult{}, false
}

// searchStaticDictAt checks the heuristic and first-byte/length filter for a
// static dictionary lookup at data[pos]. It increments *dictNumLookups and
// returns the match candidate (wordLen, wordIdx, true) when the first-byte
// and length filters pass; otherwise returns (0, 0, false).
//
// The caller is responsible for calling matchStaticDictEntryAt and incrementing
// *dictNumMatches on success. Separating the fast filter from the full match
// check keeps this function within the inliner budget so the hot path in the
// backward-reference loop is free of a function call.
func searchStaticDictAt(data []byte, pos, maxLength uint, dictNumLookups, dictNumMatches *uint) (wordLen, wordIdx uint, ok bool) {
	if *dictNumMatches < (*dictNumLookups >> 7) {
		return
	}
	key := ((loadU32LE(data, pos) * hashMul32) >> 18) << 1
	*dictNumLookups++
	e := staticDictHashEntries[key]
	lb := uint(byte(e >> 8))
	if loadByte(data, pos) == byte(e) && lb != 0 && lb <= maxLength {
		return lb, uint(e >> 16), true
	}
	return
}

// searchStaticDictionaryDeep searches the RFC 7932 static dictionary for a
// match at data[0:], checking two consecutive hash entries (key and key+1).
// This reduces the chance of missing a match due to hash collisions.
//
// Unlike the shallow variant, this writes directly into out (matching the C
// semantics where SearchInStaticDictionary mutates the HasherSearchResult).
// Used by quality >= 5 hashers that call the dictionary search internally.
func searchStaticDictionaryDeep(data []byte, maxLength, maxBackward, maxDistance uint,
	dictNumLookups, dictNumMatches *uint, out *hasherSearchResult) {
	if *dictNumMatches < (*dictNumLookups >> 7) {
		return
	}

	v := loadU32LE(data, 0)
	key := ((v * hashMul32) >> 18) << 1
	firstByte := byte(v)
	*dictNumLookups += 2

	e := staticDictHashEntries[key]
	wordLen := uint(byte(e >> 8))
	if firstByte == byte(e) && wordLen != 0 && wordLen <= maxLength {
		wordIndex := uint(e >> 16)
		offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*wordIndex

		ml := uint(findMatchLenDict(data[:wordLen], core.DictData[offset:offset+wordLen]))
		if ml+cutoffTransformsCount > wordLen && ml != 0 {
			cut := wordLen - ml
			transformID := (cut << 2) + uint((cutoffTransforms>>(cut*6))&0x3F)
			backward := maxBackward + 1 + wordIndex + (transformID << core.DictSizeBitsByLength[wordLen])
			if backward <= maxDistance {
				score := backwardReferenceScore(ml, backward)
				if score >= out.score {
					out.len = ml
					out.lenCodeDelta = int(wordLen) - int(ml)
					out.distance = backward
					out.score = score
					*dictNumMatches++
				}
			}
		}
	}

	e = staticDictHashEntries[key+1]
	wordLen = uint(byte(e >> 8))
	if firstByte == byte(e) && wordLen != 0 && wordLen <= maxLength {
		wordIndex := uint(e >> 16)
		offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*wordIndex

		ml := uint(findMatchLenDict(data[:wordLen], core.DictData[offset:offset+wordLen]))
		if ml+cutoffTransformsCount > wordLen && ml != 0 {
			cut := wordLen - ml
			transformID := (cut << 2) + uint((cutoffTransforms>>(cut*6))&0x3F)
			backward := maxBackward + 1 + wordIndex + (transformID << core.DictSizeBitsByLength[wordLen])
			if backward <= maxDistance {
				score := backwardReferenceScore(ml, backward)
				if score >= out.score {
					out.len = ml
					out.lenCodeDelta = int(wordLen) - int(ml)
					out.distance = backward
					out.score = score
					*dictNumMatches++
				}
			}
		}
	}
}
