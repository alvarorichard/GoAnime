// Exhaustive static dictionary search for the H10 binary tree hasher.
//
// Unlike the shallow/deep search variants used by H5/H6 (which return the
// single best match), this function returns matches at ALL lengths, enabling
// the Zopfli optimal parser to consider all dictionary match options.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// maxStaticDictMatchLen is the maximum match length for static dictionary
// entries including transforms. Matches BROTLI_MAX_STATIC_DICTIONARY_MATCH_LEN.
const maxStaticDictMatchLen = 37

// invalidMatch is the sentinel value for no match at a given length.
const invalidMatch = 0xFFFFFFF

// hash15 computes a 15-bit hash from 4 bytes for the dictionary bucket lookup.
func hash15(data []byte) uint32 {
	return (loadU32LE(data, 0) * hashMul32) >> (32 - 15)
}

// addMatch records a dictionary match at the given length. Only the minimum
// distance (i.e. best match) for each length is kept.
func addMatch(distance, length, lenCode uint, matches []uint32) {
	match := uint32((distance << 5) + lenCode)
	if match < matches[length] {
		matches[length] = match
	}
}

// dictMatchLength returns the number of matching bytes between data and the
// dictionary word at (id, wordLen), up to maxLen bytes.
func dictMatchLength(data []byte, id, wordLen, maxLen uint) uint {
	offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*id
	limit := min(wordLen, maxLen)
	for i := range limit {
		if data[i] != core.DictData[offset+i] {
			return i
		}
	}
	return limit
}

// isMatch checks whether the dictionary word w fully matches data[0:w.len].
// For transform==10 (uppercase first), the first byte is case-flipped.
// For transform==11 (uppercase all), all lowercase ASCII bytes are flipped.
func isDictWordMatch(w dictWord, data []byte, maxLength uint) bool {
	wordLen := uint(w.len & 0x1F)
	if wordLen > maxLength {
		return false
	}
	offset := uint(core.DictOffsetsByLength[wordLen]) + wordLen*uint(w.idx)

	switch w.transform {
	case 0:
		// Identity transform: exact match.
		for i := range wordLen {
			if data[i] != core.DictData[offset+i] {
				return false
			}
		}
		return true
	case 10:
		// Uppercase first: first byte must be uppercase version.
		d := core.DictData[offset]
		if d < 'a' || d > 'z' {
			return false
		}
		if (d ^ 32) != data[0] {
			return false
		}
		for i := uint(1); i < wordLen; i++ {
			if data[i] != core.DictData[offset+i] {
				return false
			}
		}
		return true
	default:
		// Uppercase all: all lowercase ASCII letters are uppercased.
		for i := range wordLen {
			d := core.DictData[offset+i]
			if d >= 'a' && d <= 'z' {
				if (d ^ 32) != data[i] {
					return false
				}
			} else {
				if d != data[i] {
					return false
				}
			}
		}
		return true
	}
}

// findAllStaticDictionaryMatches searches the RFC 7932 static dictionary for
// matches at all lengths from minLength to maxLength. Results are stored in
// matches[], indexed by match length. Each entry encodes (distance << 5 | lenCode);
// entries equal to invalidMatch indicate no match at that length.
//
// Returns true if any match was found.
func findAllStaticDictionaryMatches(data []byte, minLength, maxLength uint, matches []uint32) bool {
	hasFoundMatch := false

	// Primary bucket lookup: Hash15(data).
	offset := uint(staticDictBuckets[hash15(data)])
	for offset != 0 {
		w := staticDictWords[offset]
		offset++
		l := uint(w.len & 0x1F)
		n := uint(1) << core.DictSizeBitsByLength[l]
		id := uint(w.idx)
		end := w.len&0x80 != 0
		w.len = uint8(l)

		if w.transform == 0 {
			matchLen := dictMatchLength(data, id, l, maxLength)
			// Transform "" + IDENTITY + ""
			if matchLen == l {
				addMatch(id, l, l, matches)
				hasFoundMatch = true
			}
			// Transforms "" + OMIT_LAST_1 + "" and
			//            "" + OMIT_LAST_1 + "ing "
			if matchLen >= l-1 {
				addMatch(id+12*n, l-1, l, matches)
				if l+2 < maxLength &&
					data[l-1] == 'i' && data[l] == 'n' && data[l+1] == 'g' &&
					data[l+2] == ' ' {
					addMatch(id+49*n, l+3, l, matches)
				}
				hasFoundMatch = true
			}
			// Transform "" + OMIT_LAST_# + "" (# = 2..9)
			omitMinLen := minLength
			if l > 9 {
				omitMinLen = max(omitMinLen, l-9)
			}
			omitMaxLen := min(matchLen, l-2)
			for length := omitMinLen; length <= omitMaxLen; length++ {
				cut := l - length
				transformID := (cut << 2) + uint((cutoffTransforms>>(cut*6))&0x3F)
				addMatch(id+transformID*n, length, l, matches)
				hasFoundMatch = true
			}
			if matchLen < l || l+6 >= maxLength {
				if end {
					break
				}
				continue
			}
			s := data[l:]
			// Transforms "" + IDENTITY + <suffix>
			switch s[0] {
			case ' ':
				addMatch(id+n, l+1, l, matches)
				switch s[1] {
				case 'a':
					switch s[2] {
					case ' ':
						addMatch(id+28*n, l+3, l, matches)
					case 's':
						if s[3] == ' ' {
							addMatch(id+46*n, l+4, l, matches)
						}
					case 't':
						if s[3] == ' ' {
							addMatch(id+60*n, l+4, l, matches)
						}
					case 'n':
						if s[3] == 'd' && s[4] == ' ' {
							addMatch(id+10*n, l+5, l, matches)
						}
					}
				case 'b':
					if s[2] == 'y' && s[3] == ' ' {
						addMatch(id+38*n, l+4, l, matches)
					}
				case 'i':
					switch s[2] {
					case 'n':
						if s[3] == ' ' {
							addMatch(id+16*n, l+4, l, matches)
						}
					case 's':
						if s[3] == ' ' {
							addMatch(id+47*n, l+4, l, matches)
						}
					}
				case 'f':
					switch s[2] {
					case 'o':
						if s[3] == 'r' && s[4] == ' ' {
							addMatch(id+25*n, l+5, l, matches)
						}
					case 'r':
						if s[3] == 'o' && s[4] == 'm' && s[5] == ' ' {
							addMatch(id+37*n, l+6, l, matches)
						}
					}
				case 'o':
					switch s[2] {
					case 'f':
						if s[3] == ' ' {
							addMatch(id+8*n, l+4, l, matches)
						}
					case 'n':
						if s[3] == ' ' {
							addMatch(id+45*n, l+4, l, matches)
						}
					}
				case 'n':
					if s[2] == 'o' && s[3] == 't' && s[4] == ' ' {
						addMatch(id+80*n, l+5, l, matches)
					}
				case 't':
					switch s[2] {
					case 'h':
						switch s[3] {
						case 'e':
							if s[4] == ' ' {
								addMatch(id+5*n, l+5, l, matches)
							}
						case 'a':
							if s[4] == 't' && s[5] == ' ' {
								addMatch(id+29*n, l+6, l, matches)
							}
						}
					case 'o':
						if s[3] == ' ' {
							addMatch(id+17*n, l+4, l, matches)
						}
					}
				case 'w':
					if s[2] == 'i' && s[3] == 't' && s[4] == 'h' && s[5] == ' ' {
						addMatch(id+35*n, l+6, l, matches)
					}
				}
			case '"':
				addMatch(id+19*n, l+1, l, matches)
				if s[1] == '>' {
					addMatch(id+21*n, l+2, l, matches)
				}
			case '.':
				addMatch(id+20*n, l+1, l, matches)
				if s[1] == ' ' {
					addMatch(id+31*n, l+2, l, matches)
					if s[2] == 'T' && s[3] == 'h' {
						switch s[4] {
						case 'e':
							if s[5] == ' ' {
								addMatch(id+43*n, l+6, l, matches)
							}
						case 'i':
							if s[5] == 's' && s[6] == ' ' {
								addMatch(id+75*n, l+7, l, matches)
							}
						}
					}
				}
			case ',':
				addMatch(id+76*n, l+1, l, matches)
				if s[1] == ' ' {
					addMatch(id+14*n, l+2, l, matches)
				}
			case '\n':
				addMatch(id+22*n, l+1, l, matches)
				if s[1] == '\t' {
					addMatch(id+50*n, l+2, l, matches)
				}
			case ']':
				addMatch(id+24*n, l+1, l, matches)
			case '\'':
				addMatch(id+36*n, l+1, l, matches)
			case ':':
				addMatch(id+51*n, l+1, l, matches)
			case '(':
				addMatch(id+57*n, l+1, l, matches)
			case '=':
				switch s[1] {
				case '"':
					addMatch(id+70*n, l+2, l, matches)
				case '\'':
					addMatch(id+86*n, l+2, l, matches)
				}
			case 'a':
				if s[1] == 'l' && s[2] == ' ' {
					addMatch(id+84*n, l+3, l, matches)
				}
			case 'e':
				switch s[1] {
				case 'd':
					if s[2] == ' ' {
						addMatch(id+53*n, l+3, l, matches)
					}
				case 'r':
					if s[2] == ' ' {
						addMatch(id+82*n, l+3, l, matches)
					}
				case 's':
					if s[2] == 't' && s[3] == ' ' {
						addMatch(id+95*n, l+4, l, matches)
					}
				}
			case 'f':
				if s[1] == 'u' && s[2] == 'l' && s[3] == ' ' {
					addMatch(id+90*n, l+4, l, matches)
				}
			case 'i':
				switch s[1] {
				case 'v':
					if s[2] == 'e' && s[3] == ' ' {
						addMatch(id+92*n, l+4, l, matches)
					}
				case 'z':
					if s[2] == 'e' && s[3] == ' ' {
						addMatch(id+100*n, l+4, l, matches)
					}
				}
			case 'l':
				switch s[1] {
				case 'e':
					if s[2] == 's' && s[3] == 's' && s[4] == ' ' {
						addMatch(id+93*n, l+5, l, matches)
					}
				case 'y':
					if s[2] == ' ' {
						addMatch(id+61*n, l+3, l, matches)
					}
				}
			case 'o':
				if s[1] == 'u' && s[2] == 's' && s[3] == ' ' {
					addMatch(id+106*n, l+4, l, matches)
				}
			}
		} else {
			// Uppercase first (transform==10) or uppercase all (transform==11).
			isAllCaps := w.transform != 10
			if !isDictWordMatch(w, data, maxLength) {
				if end {
					break
				}
				continue
			}
			// Transform "" + kUppercase{First,All} + ""
			if isAllCaps {
				addMatch(id+44*n, l, l, matches)
			} else {
				addMatch(id+9*n, l, l, matches)
			}
			hasFoundMatch = true
			if l+1 >= maxLength {
				if end {
					break
				}
				continue
			}
			// Transforms "" + kUppercase{First,All} + <suffix>
			s := data[l:]
			switch s[0] {
			case ' ':
				if isAllCaps {
					addMatch(id+68*n, l+1, l, matches)
				} else {
					addMatch(id+4*n, l+1, l, matches)
				}
			case '"':
				if isAllCaps {
					addMatch(id+87*n, l+1, l, matches)
				} else {
					addMatch(id+66*n, l+1, l, matches)
				}
				if s[1] == '>' {
					if isAllCaps {
						addMatch(id+97*n, l+2, l, matches)
					} else {
						addMatch(id+69*n, l+2, l, matches)
					}
				}
			case '.':
				if isAllCaps {
					addMatch(id+101*n, l+1, l, matches)
				} else {
					addMatch(id+79*n, l+1, l, matches)
				}
				if s[1] == ' ' {
					if isAllCaps {
						addMatch(id+114*n, l+2, l, matches)
					} else {
						addMatch(id+88*n, l+2, l, matches)
					}
				}
			case ',':
				if isAllCaps {
					addMatch(id+112*n, l+1, l, matches)
				} else {
					addMatch(id+99*n, l+1, l, matches)
				}
				if s[1] == ' ' {
					if isAllCaps {
						addMatch(id+107*n, l+2, l, matches)
					} else {
						addMatch(id+58*n, l+2, l, matches)
					}
				}
			case '\'':
				if isAllCaps {
					addMatch(id+94*n, l+1, l, matches)
				} else {
					addMatch(id+74*n, l+1, l, matches)
				}
			case '(':
				if isAllCaps {
					addMatch(id+113*n, l+1, l, matches)
				} else {
					addMatch(id+78*n, l+1, l, matches)
				}
			case '=':
				switch s[1] {
				case '"':
					if isAllCaps {
						addMatch(id+105*n, l+2, l, matches)
					} else {
						addMatch(id+104*n, l+2, l, matches)
					}
				case '\'':
					if isAllCaps {
						addMatch(id+116*n, l+2, l, matches)
					} else {
						addMatch(id+108*n, l+2, l, matches)
					}
				}
			}
		}

		if end {
			break
		}
	}

	// Transforms with prefixes " " and "."
	if maxLength >= 5 && (data[0] == ' ' || data[0] == '.') {
		isSpace := data[0] == ' '
		offset := uint(staticDictBuckets[hash15(data[1:])])
		for offset != 0 {
			w := staticDictWords[offset]
			offset++
			l := uint(w.len & 0x1F)
			n := uint(1) << core.DictSizeBitsByLength[l]
			id := uint(w.idx)
			end := w.len&0x80 != 0
			w.len = uint8(l)

			if w.transform == 0 {
				if !isDictWordMatch(w, data[1:], maxLength-1) {
					if end {
						break
					}
					continue
				}
				// Transforms " " + IDENTITY + "" and "." + IDENTITY + ""
				if isSpace {
					addMatch(id+6*n, l+1, l, matches)
				} else {
					addMatch(id+32*n, l+1, l, matches)
				}
				hasFoundMatch = true
				if l+2 >= maxLength {
					if end {
						break
					}
					continue
				}
				s := data[l+1:]
				switch s[0] {
				case ' ':
					if isSpace {
						addMatch(id+2*n, l+2, l, matches)
					} else {
						addMatch(id+77*n, l+2, l, matches)
					}
				case '(':
					if isSpace {
						addMatch(id+89*n, l+2, l, matches)
					} else {
						addMatch(id+67*n, l+2, l, matches)
					}
				case ',':
					if isSpace {
						addMatch(id+103*n, l+2, l, matches)
						if s[1] == ' ' {
							addMatch(id+33*n, l+3, l, matches)
						}
					}
				case '.':
					if isSpace {
						addMatch(id+71*n, l+2, l, matches)
						if s[1] == ' ' {
							addMatch(id+52*n, l+3, l, matches)
						}
					}
				case '=':
					if isSpace {
						switch s[1] {
						case '"':
							addMatch(id+81*n, l+3, l, matches)
						case '\'':
							addMatch(id+98*n, l+3, l, matches)
						}
					}
				}
			} else if isSpace {
				// Uppercase first/all with " " prefix.
				isAllCaps := w.transform != 10
				if !isDictWordMatch(w, data[1:], maxLength-1) {
					if end {
						break
					}
					continue
				}
				if isAllCaps {
					addMatch(id+85*n, l+1, l, matches)
				} else {
					addMatch(id+30*n, l+1, l, matches)
				}
				hasFoundMatch = true
				if l+2 >= maxLength {
					if end {
						break
					}
					continue
				}
				s := data[l+1:]
				switch s[0] {
				case ' ':
					if isAllCaps {
						addMatch(id+83*n, l+2, l, matches)
					} else {
						addMatch(id+15*n, l+2, l, matches)
					}
				case ',':
					if !isAllCaps {
						addMatch(id+109*n, l+2, l, matches)
					}
					if s[1] == ' ' {
						if isAllCaps {
							addMatch(id+111*n, l+3, l, matches)
						} else {
							addMatch(id+65*n, l+3, l, matches)
						}
					}
				case '.':
					if isAllCaps {
						addMatch(id+115*n, l+2, l, matches)
					} else {
						addMatch(id+96*n, l+2, l, matches)
					}
					if s[1] == ' ' {
						if isAllCaps {
							addMatch(id+117*n, l+3, l, matches)
						} else {
							addMatch(id+91*n, l+3, l, matches)
						}
					}
				case '=':
					switch s[1] {
					case '"':
						if isAllCaps {
							addMatch(id+110*n, l+3, l, matches)
						} else {
							addMatch(id+118*n, l+3, l, matches)
						}
					case '\'':
						if isAllCaps {
							addMatch(id+119*n, l+3, l, matches)
						} else {
							addMatch(id+120*n, l+3, l, matches)
						}
					}
				}
			}

			if end {
				break
			}
		}
	}

	if maxLength >= 6 {
		// Transforms with prefixes "e ", "s ", ", " and "\xC2\xA0"
		if (data[1] == ' ' &&
			(data[0] == 'e' || data[0] == 's' || data[0] == ',')) ||
			(data[0] == 0xC2 && data[1] == 0xA0) {
			offset := uint(staticDictBuckets[hash15(data[2:])])
			for offset != 0 {
				w := staticDictWords[offset]
				offset++
				l := uint(w.len & 0x1F)
				n := uint(1) << core.DictSizeBitsByLength[l]
				id := uint(w.idx)
				end := w.len&0x80 != 0
				w.len = uint8(l)

				if w.transform == 0 && isDictWordMatch(w, data[2:], maxLength-2) {
					if data[0] == 0xC2 {
						addMatch(id+102*n, l+2, l, matches)
						hasFoundMatch = true
					} else if l+2 < maxLength && data[l+2] == ' ' {
						var t uint
						switch data[0] {
						case 'e':
							t = 18
						case 's':
							t = 7
						default:
							t = 13
						}
						addMatch(id+t*n, l+3, l, matches)
						hasFoundMatch = true
					}
				}

				if end {
					break
				}
			}
		}
	}

	if maxLength >= 9 {
		// Transforms with prefixes " the " and ".com/"
		if (data[0] == ' ' && data[1] == 't' && data[2] == 'h' &&
			data[3] == 'e' && data[4] == ' ') ||
			(data[0] == '.' && data[1] == 'c' && data[2] == 'o' &&
				data[3] == 'm' && data[4] == '/') {
			offset := uint(staticDictBuckets[hash15(data[5:])])
			for offset != 0 {
				w := staticDictWords[offset]
				offset++
				l := uint(w.len & 0x1F)
				n := uint(1) << core.DictSizeBitsByLength[l]
				id := uint(w.idx)
				end := w.len&0x80 != 0
				w.len = uint8(l)

				if w.transform == 0 && isDictWordMatch(w, data[5:], maxLength-5) {
					if data[0] == ' ' {
						addMatch(id+41*n, l+5, l, matches)
					} else {
						addMatch(id+72*n, l+5, l, matches)
					}
					hasFoundMatch = true
					if l+5 < maxLength {
						s := data[l+5:]
						if data[0] == ' ' {
							if l+8 < maxLength &&
								s[0] == ' ' && s[1] == 'o' && s[2] == 'f' && s[3] == ' ' {
								addMatch(id+62*n, l+9, l, matches)
								if l+12 < maxLength &&
									s[4] == 't' && s[5] == 'h' && s[6] == 'e' && s[7] == ' ' {
									addMatch(id+73*n, l+13, l, matches)
								}
							}
						}
					}
				}

				if end {
					break
				}
			}
		}
	}

	return hasFoundMatch
}
