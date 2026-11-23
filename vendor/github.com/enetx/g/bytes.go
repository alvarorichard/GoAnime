package g

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/enetx/g/cmp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// Bytes is a wrapper around the []byte type.
type Bytes []byte

var (
	lower = cases.Lower(language.Und)
	upper = cases.Upper(language.Und)
	title = cases.Title(language.Und)
)

// NewBytes creates a Bytes from the provided string or byte slice, mirroring
// NewString. For an empty pre-sized buffer use make(Bytes, n) or
// make(Bytes, n, cap) directly.
func NewBytes[T ~string | ~[]byte](b T) Bytes { return Bytes(b) }

// Transform applies a transformation function to the Bytes and returns the result.
func (bs Bytes) Transform[U any](fn func(Bytes) U) U { return fn(bs) }

// Min returns the minimum of Bytes.
func (bs Bytes) Min(b ...Bytes) Bytes { return cmp.MinBy(Bytes.Cmp, append(b, bs)...) }

// Max returns the maximum of Bytes.
func (bs Bytes) Max(b ...Bytes) Bytes { return cmp.MaxBy(Bytes.Cmp, append(b, bs)...) }

// Reverse reverses bytes for ASCII or invalid UTF-8; for valid UTF-8 it reverses by runes.
func (bs Bytes) Reverse() Bytes {
	n := len(bs)
	out := make(Bytes, n)
	ascii := true

	for _, b := range bs {
		if b >= utf8.RuneSelf {
			ascii = false
			break
		}
	}

	if ascii || !utf8.Valid(bs) {
		for i, b := range bs {
			out[n-1-i] = b
		}

		return out
	}

	w := 0

	for i := n; i > 0; {
		r, size := utf8.DecodeLastRune(bs[:i])
		w += utf8.EncodeRune(out[w:], r)
		i -= size
	}

	return out
}

// Replace replaces the first 'n' occurrences of 'oldB' with 'newB' in the Bytes.
func (bs Bytes) Replace(oldB, newB Bytes, n Int) Bytes { return bytes.Replace(bs, oldB, newB, n.Std()) }

// ReplaceAll replaces all occurrences of 'oldB' with 'newB' in the Bytes.
func (bs Bytes) ReplaceAll(oldB, newB Bytes) Bytes { return bytes.ReplaceAll(bs, oldB, newB) }

// Trim trims leading and trailing white space from the Bytes.
func (bs Bytes) Trim() Bytes { return bytes.TrimSpace(bs) }

// TrimStart removes leading white space from the Bytes.
func (bs Bytes) TrimStart() Bytes { return bytes.TrimLeftFunc(bs, unicode.IsSpace) }

// TrimEnd removes trailing white space from the Bytes.
func (bs Bytes) TrimEnd() Bytes { return bytes.TrimRightFunc(bs, unicode.IsSpace) }

// TrimSet trims the specified set of characters from both the beginning and end of the Bytes.
func (bs Bytes) TrimSet(cutset String) Bytes { return bytes.Trim(bs, cutset.Std()) }

// TrimStartSet removes the specified set of characters from the beginning of the Bytes.
func (bs Bytes) TrimStartSet(cutset String) Bytes { return bytes.TrimLeft(bs, cutset.Std()) }

// TrimEndSet removes the specified set of characters from the end of the Bytes.
func (bs Bytes) TrimEndSet(cutset String) Bytes { return bytes.TrimRight(bs, cutset.Std()) }

// intFromBytes parses up to 8 bytes into Int using the given byte order.
// For BE: uses the last 8 bytes if longer; pads on the left if shorter.
// For LE: uses the first 8 bytes if longer; pads on the right if shorter.
func intFromBytes(bs Bytes, order binary.ByteOrder) Int {
	var buf [8]byte

	switch order {
	case binary.BigEndian:
		if len(bs) > 8 {
			bs = bs[len(bs)-8:]
		}

		copy(buf[8-len(bs):], bs)

		if len(bs) > 0 && bs[0]&0x80 != 0 && len(bs) < 8 {
			for i := 0; i < 8-len(bs); i++ {
				buf[i] = 0xFF
			}
		}
	case binary.LittleEndian:
		if len(bs) > 8 {
			bs = bs[:8]
		}

		copy(buf[:len(bs)], bs)

		if len(bs) > 0 && bs[len(bs)-1]&0x80 != 0 && len(bs) < 8 {
			for i := len(bs); i < 8; i++ {
				buf[i] = 0xFF
			}
		}
	}

	return Int(int64(order.Uint64(buf[:])))
}

// IntBE interprets the Bytes as a signed 64-bit integer in BigEndian order.
// If the Bytes length is less than 8, the value is sign-extended to 64 bits
// (the most-significant byte's high bit determines the sign).
// If the Bytes length is greater than 8, only the last 8 bytes are used.
func (bs Bytes) IntBE() Int { return intFromBytes(bs, binary.BigEndian) }

// IntLE interprets the Bytes as a signed 64-bit integer in LittleEndian order.
// If the Bytes length is less than 8, the value is sign-extended to 64 bits
// (the most-significant byte's high bit determines the sign).
// If the Bytes length is greater than 8, only the first 8 bytes are used.
func (bs Bytes) IntLE() Int { return intFromBytes(bs, binary.LittleEndian) }

// FloatBE interprets the Bytes as an IEEE-754 64-bit float in BigEndian order.
// If the Bytes length is not exactly 8, returns 0.
func (bs Bytes) FloatBE() Float {
	if len(bs) != 8 {
		return 0
	}

	bits := binary.BigEndian.Uint64(bs)

	return Float(math.Float64frombits(bits))
}

// FloatLE interprets the Bytes as an IEEE-754 64-bit float in LittleEndian order.
// If the Bytes length is not exactly 8, returns 0.
func (bs Bytes) FloatLE() Float {
	if len(bs) != 8 {
		return 0
	}

	bits := binary.LittleEndian.Uint64(bs)

	return Float(math.Float64frombits(bits))
}

// StripPrefix trims the specified Bytes prefix from the Bytes.
func (bs Bytes) StripPrefix(cutset Bytes) Bytes { return bytes.TrimPrefix(bs, cutset) }

// StripSuffix trims the specified Bytes suffix from the Bytes.
func (bs Bytes) StripSuffix(cutset Bytes) Bytes { return bytes.TrimSuffix(bs, cutset) }

// StartsWith checks if the Bytes starts with the specified prefix.
func (bs Bytes) StartsWith(prefix Bytes) bool { return bytes.HasPrefix(bs, prefix) }

// StartsWithAny checks if the Bytes starts with any of the provided prefixes.
// The method accepts a variable number of arguments, allowing for checking against multiple
// prefixes at once. It iterates over the provided prefixes and uses the HasPrefix function from
// the bytes package to check if the Bytes starts with each prefix.
// The function returns true if the Bytes starts with any of the prefixes, and false otherwise.
func (bs Bytes) StartsWithAny(prefixes ...Bytes) bool {
	for _, prefix := range prefixes {
		if bytes.HasPrefix(bs, prefix) {
			return true
		}
	}

	return false
}

// EndsWith checks if the Bytes ends with the specified suffix.
func (bs Bytes) EndsWith(suffix Bytes) bool { return bytes.HasSuffix(bs, suffix) }

// EndsWithAny checks if the Bytes ends with any of the provided suffixes.
// The method accepts a variable number of arguments, allowing for checking against multiple
// suffixes at once. It iterates over the provided suffixes and uses the HasSuffix function from
// the bytes package to check if the Bytes ends with each suffix.
// The function returns true if the Bytes ends with any of the suffixes, and false otherwise.
func (bs Bytes) EndsWithAny(suffixes ...Bytes) bool {
	for _, suffix := range suffixes {
		if bytes.HasSuffix(bs, suffix) {
			return true
		}
	}

	return false
}

// Split splits the Bytes by the specified separator. If sep is empty, the
// Bytes are split after each UTF-8 rune. See [String.Lines] for why the return
// type is a plain slice.
func (bs Bytes) Split(sep Bytes) []Bytes {
	return castBytesSlices(bytes.Split(bs, sep))
}

// SplitAfter splits the Bytes after each instance of the specified separator.
// See [String.Lines] for why the return type is a plain slice.
func (bs Bytes) SplitAfter(sep Bytes) []Bytes {
	return castBytesSlices(bytes.SplitAfter(bs, sep))
}

// SplitN splits the Bytes into subslices using the provided separator and
// returns a plain []Bytes of the results (convert with Slice[Bytes] for
// chaining). The n parameter controls the number of subslices to return:
// - If n is negative, there is no limit on the number of subslices returned.
// - If n is zero, an empty slice is returned.
// - If n is positive, at most n subslices are returned.
func (bs Bytes) SplitN(sep Bytes, n Int) []Bytes {
	parts := bytes.SplitN(bs, sep, n.Std())

	result := make([]Bytes, len(parts))
	for i, p := range parts {
		result[i] = Bytes(p)
	}

	return result
}

// Lines splits the Bytes by lines, with trailing whitespace trimmed per line.
// See [String.Lines] for why the return type is a plain slice.
func (bs Bytes) Lines() []Bytes {
	var result []Bytes

	for line := range bytes.Lines(bs) {
		result = append(result, Bytes(line).TrimEnd())
	}

	return result
}

// Fields splits the Bytes around whitespace. See [String.Lines] for why the
// return type is a plain slice.
func (bs Bytes) Fields() []Bytes {
	return castBytesSlices(bytes.Fields(bs))
}

// FieldsBy splits the Bytes using a custom function to determine the field
// boundaries. See [String.Lines] for why the return type is a plain slice.
func (bs Bytes) FieldsBy(fn func(r rune) bool) []Bytes {
	return castBytesSlices(bytes.FieldsFunc(bs, fn))
}

// Append appends the given Bytes to the current Bytes.
//
// Warning: like the builtin append, this may reuse and mutate the receiver's
// backing array when it has spare capacity, so the returned Bytes can alias bs.
// This is asymmetric with Prepend (which always copies) and with the immutable
// String.Append. Clone the receiver first if it must remain unchanged.
func (bs Bytes) Append(obs Bytes) Bytes { return append(bs, obs...) }

// Prepend prepends the given Bytes to the current Bytes.
func (bs Bytes) Prepend(obs Bytes) Bytes {
	if len(obs) == 0 {
		return bs
	}

	if len(bs) == 0 {
		return obs
	}

	out := make(Bytes, len(obs)+len(bs))
	copy(out, obs)
	copy(out[len(obs):], bs)

	return out
}

// Std returns the Bytes as a byte slice.
func (bs Bytes) Std() []byte { return bs }

// Clone creates a new Bytes instance with the same content as the current Bytes.
func (bs Bytes) Clone() Bytes { return bytes.Clone(bs) }

// Cmp compares the Bytes with another Bytes and returns an cmp.Ordering.
func (bs Bytes) Cmp(obs Bytes) cmp.Ordering { return cmp.Ordering(bytes.Compare(bs, obs)) }

// Contains checks if the Bytes contains the specified Bytes.
func (bs Bytes) Contains(obs Bytes) bool { return bytes.Contains(bs, obs) }

// ContainsAny checks if the Bytes contains any of the specified Bytes.
func (bs Bytes) ContainsAny(obss ...Bytes) bool {
	for _, obs := range obss {
		if bytes.Contains(bs, obs) {
			return true
		}
	}

	return false
}

// ContainsAll checks if the Bytes contains all of the specified Bytes.
func (bs Bytes) ContainsAll(obss ...Bytes) bool {
	for _, obs := range obss {
		if !bytes.Contains(bs, obs) {
			return false
		}
	}

	return true
}

// ContainsAnyChars checks if the given Bytes contains any characters from the input String.
func (bs Bytes) ContainsAnyChars(chars String) bool { return bytes.ContainsAny(bs, chars.Std()) }

// ContainsRune checks if the Bytes contains the specified rune.
func (bs Bytes) ContainsRune(r rune) bool { return bytes.ContainsRune(bs, r) }

// Count counts the number of occurrences of the specified Bytes in the Bytes.
func (bs Bytes) Count(obs Bytes) Int { return Int(bytes.Count(bs, obs)) }

// IsEmpty checks if the Bytes is empty.
func (bs Bytes) IsEmpty() bool { return len(bs) == 0 }

// Eq checks if the Bytes is equal to another Bytes.
func (bs Bytes) Eq(obs Bytes) bool { return bs.Cmp(obs).IsEq() }

// EqFold compares two Bytes slices case-insensitively.
func (bs Bytes) EqFold(obs Bytes) bool { return bytes.EqualFold(bs, obs) }

// Gt checks if the Bytes is greater than another Bytes.
func (bs Bytes) Gt(obs Bytes) bool { return bs.Cmp(obs).IsGt() }

// Gte checks if the Bytes is greater than or equal to another Bytes.
func (bs Bytes) Gte(obs Bytes) bool { return !bs.Cmp(obs).IsLt() }

// String returns the Bytes as an String.
func (bs Bytes) String() String { return String(bs) }

// StringUnsafe converts the Bytes into a String without copying memory.
// Warning: the resulting String shares the same underlying memory as the original Bytes.
// If the Bytes is modified later, the String will reflect those changes and may cause undefined behavior.
func (bs Bytes) StringUnsafe() String { return String(unsafe.String(unsafe.SliceData(bs), len(bs))) }

// TryInt parses the Bytes as an integer, mirroring String.TryInt.
func (bs Bytes) TryInt() Result[Int] { return bs.StringUnsafe().TryInt() }

// TryUint parses the Bytes as an unsigned integer, mirroring String.TryUint.
func (bs Bytes) TryUint() Result[uint] { return bs.StringUnsafe().TryUint() }

// TryFloat parses the Bytes as a float, mirroring String.TryFloat.
func (bs Bytes) TryFloat() Result[Float] { return bs.StringUnsafe().TryFloat() }

// TryBool parses the Bytes as a bool, mirroring String.TryBool.
func (bs Bytes) TryBool() Result[bool] { return bs.StringUnsafe().TryBool() }

// TryComplex parses the Bytes as a complex number, mirroring String.TryComplex.
func (bs Bytes) TryComplex() Result[complex128] { return bs.StringUnsafe().TryComplex() }

// TryBigInt parses the Bytes as a *big.Int, mirroring String.TryBigInt.
func (bs Bytes) TryBigInt() Result[*big.Int] { return bs.StringUnsafe().TryBigInt() }

// Index returns the index of the first instance of obs in bs, or -1 if obs is not present in bs.
func (bs Bytes) Index(obs Bytes) Int { return Int(bytes.Index(bs, obs)) }

// LastIndex returns the index of the last instance of obs in bs, or -1 if obs is not present in bs.
func (bs Bytes) LastIndex(obs Bytes) Int { return Int(bytes.LastIndex(bs, obs)) }

// IndexByte returns the index of the first instance of the byte b in bs, or -1 if b is not
// present in bs.
func (bs Bytes) IndexByte(b byte) Int { return Int(bytes.IndexByte(bs, b)) }

// LastIndexByte returns the index of the last instance of the byte b in bs, or -1 if b is not
// present in bs.
func (bs Bytes) LastIndexByte(b byte) Int { return Int(bytes.LastIndexByte(bs, b)) }

// IndexRune returns the index of the first instance of the rune r in bs, or -1 if r is not
// present in bs.
func (bs Bytes) IndexRune(r rune) Int { return Int(bytes.IndexRune(bs, r)) }

// Len returns the length of the Bytes.
func (bs Bytes) Len() Int { return Int(len(bs)) }

// LenRunes returns the number of runes in the Bytes.
func (bs Bytes) LenRunes() Int { return Int(utf8.RuneCount(bs)) }

// Lt checks if the Bytes is less than another Bytes.
func (bs Bytes) Lt(obs Bytes) bool { return bs.Cmp(obs).IsLt() }

// Lte checks if the Bytes is less than or equal to another Bytes.
func (bs Bytes) Lte(obs Bytes) bool { return !bs.Cmp(obs).IsGt() }

// Map applies a function to each rune in the Bytes and returns the modified Bytes.
func (bs Bytes) Map(fn func(rune) rune) Bytes { return bytes.Map(fn, bs) }

// Ne checks if the Bytes is not equal to another Bytes.
func (bs Bytes) Ne(obs Bytes) bool { return !bs.Eq(obs) }

// Reader returns a *bytes.Reader initialized with the content of Bytes.
func (bs Bytes) Reader() *bytes.Reader { return bytes.NewReader(bs) }

// Repeat returns a new Bytes consisting of the current Bytes repeated 'count' times.
func (bs Bytes) Repeat(count Int) Bytes { return bytes.Repeat(bs, count.Std()) }

// Title converts the Bytes to title case.
func (bs Bytes) Title() Bytes { return title.Bytes(bs) }

// NormalizeNFC returns a new Bytes with its Unicode characters normalized using the NFC form.
func (bs Bytes) NormalizeNFC() Bytes { return norm.NFC.Bytes(bs) }

// Reset resets the length of the Bytes slice to zero, preserving its capacity.
func (bs *Bytes) Reset() { *bs = (*bs)[:0] }

// Runes returns the Bytes as a plain slice of runes.
func (bs Bytes) Runes() []rune { return bytes.Runes(bs) }

// Chars splits the Bytes into individual UTF-8 characters, equivalent to
// bs.Split(Bytes("")) and mirroring String.Chars.
func (bs Bytes) Chars() []Bytes { return bs.Split(Bytes("")) }

func convertCase(bs Bytes, from byte, diff int8, ucFn func([]byte) []byte) Bytes {
	for i, b := range bs {
		if b >= utf8.RuneSelf {
			return ucFn(bs)
		}

		if from <= b && b <= from+25 {
			for _, c := range bs[i+1:] {
				if c >= utf8.RuneSelf {
					return ucFn(bs)
				}
			}

			out := make(Bytes, len(bs))
			copy(out, bs[:i])
			out[i] = byte(int8(b) + diff)

			for j, c := range bs[i+1:] {
				if from <= c && c <= from+25 {
					out[i+1+j] = byte(int8(c) + diff)
				} else {
					out[i+1+j] = c
				}
			}

			return out
		}
	}

	return bs
}

// Lower converts the Bytes to lowercase.
func (bs Bytes) Lower() Bytes { return convertCase(bs, 'A', 'a'-'A', lower.Bytes) }

// Upper converts the Bytes to uppercase.
func (bs Bytes) Upper() Bytes { return convertCase(bs, 'a', 'A'-'a', upper.Bytes) }

// IsLower reports whether bs contains at least one letter and no uppercase letters.
func (bs Bytes) IsLower() bool {
	letter := false

	for i, b := range bs {
		if b >= utf8.RuneSelf {
			rest := bs[i:]
			for len(rest) > 0 {
				r, size := utf8.DecodeRune(rest)
				rest = rest[size:]
				if r == utf8.RuneError && size == 1 {
					continue
				}

				if unicode.IsLetter(r) {
					letter = true
					if unicode.IsUpper(r) {
						return false
					}
				}
			}

			return letter
		}

		if 'A' <= b && b <= 'Z' {
			return false
		}

		if 'a' <= b && b <= 'z' {
			letter = true
		}
	}

	return letter
}

// IsUpper reports whether bs contains at least one letter and no lowercase letters.
func (bs Bytes) IsUpper() bool {
	letter := false

	for i, b := range bs {
		if b >= utf8.RuneSelf {
			rest := bs[i:]
			for len(rest) > 0 {
				r, size := utf8.DecodeRune(rest)
				rest = rest[size:]
				if r == utf8.RuneError && size == 1 {
					continue
				}

				if unicode.IsLetter(r) {
					letter = true
					if unicode.IsLower(r) {
						return false
					}
				}
			}

			return letter
		}

		if 'a' <= b && b <= 'z' {
			return false
		}

		if 'A' <= b && b <= 'Z' {
			letter = true
		}
	}

	return letter
}

// IsTitle reports whether bs is in title case: the first letter of each word
// is uppercase (or titlecase), the remaining letters are lowercase.
// Non-letter characters act as word separators. Returns false if bs has no letters.
func (bs Bytes) IsTitle() bool {
	letter := false
	prevLetter := false

	for i, b := range bs {
		if b >= utf8.RuneSelf {
			rest := bs[i:]
			for len(rest) > 0 {
				r, size := utf8.DecodeRune(rest)
				rest = rest[size:]

				if r == utf8.RuneError && size == 1 {
					prevLetter = false
					continue
				}

				if unicode.IsLetter(r) {
					letter = true
					if prevLetter && !unicode.IsLower(r) {
						return false
					}
					if !prevLetter && !unicode.IsUpper(r) && !unicode.IsTitle(r) {
						return false
					}
					prevLetter = true
				} else {
					prevLetter = false
				}
			}

			return letter
		}

		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') {
			letter = true
			if prevLetter == (b <= 'Z') {
				return false
			}
			prevLetter = true
		} else {
			prevLetter = false
		}
	}

	return letter
}

// Print writes the content of the Bytes to the standard output (console)
// and returns the Bytes unchanged.
func (bs Bytes) Print() Bytes { fmt.Print(bs); return bs }

// Println writes the content of the Bytes to the standard output (console) with a newline
// and returns the Bytes unchanged.
func (bs Bytes) Println() Bytes { fmt.Println(bs); return bs }

// Scan implements the database/sql.Scanner interface for g.Bytes.
//
// Behavior:
//   - If src is nil, the Bytes slice is set to nil (SQL NULL).
//   - If src is a []byte, a copy is stored (database/sql may reuse the driver's
//     buffer on the next row, so the bytes must not be retained by reference).
//   - Otherwise, an error is returned.
//
// Supported SQL types (common):
//   - BLOB / BYTEA → []byte
//
// Notes:
//   - This allows g.Bytes to be used directly with database/sql and compatible drivers.
func (bs *Bytes) Scan(src any) error {
	if src == nil {
		*bs = nil
		return nil
	}

	if b, ok := src.([]byte); ok {
		*bs = append(Bytes(nil), b...)
		return nil
	}

	return fmt.Errorf("g.Bytes.Scan: cannot scan %T into g.Bytes", src)
}

// Value implements the database/sql/driver.Valuer interface for g.Bytes.
//
// Behavior:
//   - Returns the underlying byte slice, ready for database insertion.
//   - Always returns a value compatible with SQL BLOB / BYTEA types.
func (bs Bytes) Value() (driver.Value, error) { return []byte(bs), nil }

// castBytesSlices reinterprets a [][]byte as []Bytes without copying: Bytes is
// defined as `type Bytes []byte`, so the two slice types share one memory
// layout.
func castBytesSlices(bss [][]byte) []Bytes {
	return unsafe.Slice((*Bytes)(unsafe.SliceData(bss)), len(bss))
}
