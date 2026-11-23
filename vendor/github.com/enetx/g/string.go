package g

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/text/unicode/norm"

	"github.com/enetx/g/cmp"
)

// String is a wrapper around the string type.
type String string

// Named is a map-like type that stores key-value pairs for resolving named
// placeholders in Format.
type Named Map[String, any]

// NewString creates a new String from the provided string, rune, byte, rune slice, or byte slice.
func NewString[T ~string | rune | byte | ~[]rune | ~[]byte](str T) String { return String(str) }

// Clone returns a copy of the String.
// It ensures that the returned String does not share underlying memory with the original String,
// making it safe to modify or store independently.
func (s String) Clone() String { return String(strings.Clone(s.Std())) }

// Transform applies a transformation function to the String and returns the result.
func (s String) Transform[U any](fn func(String) U) U { return fn(s) }

// Builder returns a new Builder initialized with the content of the String.
func (s String) Builder() *Builder {
	b := new(Builder)
	b.WriteString(s)
	return b
}

// Min returns the minimum of Strings.
func (s String) Min(b ...String) String { return cmp.Min(append(b, s)...) }

// Max returns the maximum of Strings.
func (s String) Max(b ...String) String { return cmp.Max(append(b, s)...) }

// IsASCII checks if all characters in the String are ASCII bytes.
func (s String) IsASCII() bool {
	for i := range s {
		if s[i] >= 0x80 {
			return false
		}
	}

	return true
}

// IsDigit checks if all characters in the String are digits.
func (s String) IsDigit() bool {
	if s.IsEmpty() {
		return false
	}

	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}

// TryInt tries to parse the String as an int and returns an Int.
func (s String) TryInt() Result[Int] {
	hint, err := strconv.ParseInt(s.Std(), 0, 64)
	if err != nil {
		return Err[Int](Errorf("{:w}: \"{}\"", ErrParseInt, s))
	}

	return Ok(Int(hint))
}

// TryBigInt attempts to convert the String receiver into a Result containing a *big.Int.
// This function assumes the string represents a numerical value, which can be in decimal,
// hexadecimal (prefixed with "0x"), or octal (prefixed with "0") format. The function
// leverages the SetString method of the math/big package, automatically detecting the
// numeric base when set to 0.
//
// If the string is correctly formatted and represents a valid number, TryBigInt returns
// an Ok containing the *big.Int parsed from the string. If the string is empty, contains
// invalid characters, or does not conform to a recognizable numeric format, TryBigInt
// returns an Err describing the invalid input.
//
// Returns:
//   - A Result[*big.Int] encapsulating the conversion result. It returns Ok[*big.Int]
//     with the parsed value if successful, otherwise Err[*big.Int] if the parsing fails.
func (s String) TryBigInt() Result[*big.Int] {
	if bigInt, ok := new(big.Int).SetString(s.Std(), 0); ok {
		return Ok(bigInt)
	}

	return Err[*big.Int](Errorf("{:w}: \"{}\"", ErrParseBigInt, s))
}

// TryFloat tries to parse the String as a float64 and returns an Float.
func (s String) TryFloat() Result[Float] {
	float, err := strconv.ParseFloat(s.Std(), 64)
	if err != nil {
		return Err[Float](Errorf("{:w}: \"{}\"", ErrParseFloat, s))
	}

	return Ok(Float(float))
}

// TryBool tries to parse the String as a bool and returns the result.
// It accepts the values understood by strconv.ParseBool: 1, t, T, TRUE, true,
// True, 0, f, F, FALSE, false, False.
func (s String) TryBool() Result[bool] {
	b, err := strconv.ParseBool(s.Std())
	if err != nil {
		return Err[bool](Errorf("{:w}: \"{}\"", ErrParseBool, s))
	}

	return Ok(b)
}

// TryUint tries to parse the String as an unsigned integer and returns a uint.
// The base is inferred from the prefix (0x, 0o/0, 0b), matching TryInt.
func (s String) TryUint() Result[uint] {
	u, err := strconv.ParseUint(s.Std(), 0, 64)
	if err != nil {
		return Err[uint](Errorf("{:w}: \"{}\"", ErrParseUint, s))
	}

	return Ok(uint(u))
}

// TryComplex tries to parse the String as a complex number and returns a complex128.
func (s String) TryComplex() Result[complex128] {
	c, err := strconv.ParseComplex(s.Std(), 128)
	if err != nil {
		return Err[complex128](Errorf("{:w}: \"{}\"", ErrParseComplex, s))
	}

	return Ok(c)
}

// Title converts the String to title case.
func (s String) Title() String { return String(title.String(s.Std())) }

// Lower returns the String in lowercase.
func (s String) Lower() String { return s.BytesUnsafe().Lower().StringUnsafe() }

// Upper returns the String in uppercase.
func (s String) Upper() String { return s.BytesUnsafe().Upper().StringUnsafe() }

// IsLower checks if the String consists only of lowercase letters.
func (s String) IsLower() bool { return s.BytesUnsafe().IsLower() }

// IsUpper checks if the String consists only of uppercase letters.
func (s String) IsUpper() bool { return s.BytesUnsafe().IsUpper() }

// IsTitle checks if the String is in title case.
func (s String) IsTitle() bool { return s.BytesUnsafe().IsTitle() }

// Trim removes leading and trailing white space from the String.
func (s String) Trim() String { return String(strings.TrimSpace(s.Std())) }

// TrimStart removes leading white space from the String.
func (s String) TrimStart() String { return String(strings.TrimLeftFunc(s.Std(), unicode.IsSpace)) }

// TrimEnd removes trailing white space from the String.
func (s String) TrimEnd() String { return String(strings.TrimRightFunc(s.Std(), unicode.IsSpace)) }

// TrimSet removes the specified set of characters from both the beginning and end of the String.
func (s String) TrimSet(cutset String) String { return String(strings.Trim(s.Std(), cutset.Std())) }

// TrimStartSet removes the specified set of characters from the beginning of the String.
func (s String) TrimStartSet(cutset String) String {
	return String(strings.TrimLeft(s.Std(), cutset.Std()))
}

// TrimEndSet removes the specified set of characters from the end of the String.
func (s String) TrimEndSet(cutset String) String {
	return String(strings.TrimRight(s.Std(), cutset.Std()))
}

// StripPrefix trims the specified prefix from the String.
func (s String) StripPrefix(prefix String) String {
	return String(strings.TrimPrefix(s.Std(), prefix.Std()))
}

// StripSuffix trims the specified suffix from the String.
func (s String) StripSuffix(suffix String) String {
	return String(strings.TrimSuffix(s.Std(), suffix.Std()))
}

// Replace replaces the 'oldS' String with the 'newS' String for the specified number of
// occurrences.
func (s String) Replace(oldS, newS String, n Int) String {
	return String(strings.Replace(s.Std(), oldS.Std(), newS.Std(), n.Std()))
}

// ReplaceAll replaces all occurrences of the 'oldS' String with the 'newS' String.
func (s String) ReplaceAll(oldS, newS String) String {
	return String(strings.ReplaceAll(s.Std(), oldS.Std(), newS.Std()))
}

// ReplaceMulti creates a custom replacer to perform multiple string replacements.
//
// Parameters:
//
// - oldnew ...String: Pairs of strings to be replaced. Specify as many pairs as needed.
//
// Returns:
//
// - String: A new string with replacements applied using the custom replacer.
//
// Example usage:
//
//	original := g.String("Hello, world! This is a test.")
//	replaced := original.ReplaceMulti(
//	    "Hello", "Greetings",
//	    "world", "universe",
//	    "test", "example",
//	)
//	// replaced contains "Greetings, universe! This is a example."
func (s String) ReplaceMulti(oldnew ...String) String {
	pairs := make([]string, len(oldnew))
	for i, str := range oldnew {
		pairs[i] = str.Std()
	}

	return String(strings.NewReplacer(pairs...).Replace(s.Std()))
}

// Remove removes all occurrences of specified substrings from the String.
//
// Parameters:
//
// - matches ...String: Substrings to be removed from the string. Specify as many substrings as needed.
//
// Returns:
//
// - String: A new string with all specified substrings removed.
//
// Example usage:
//
//	original := g.String("Hello, world! This is a test.")
//	modified := original.Remove(
//	    "Hello",
//	    "test",
//	)
//	// modified contains ", world! This is a ."
func (s String) Remove(matches ...String) String {
	if len(matches) == 0 {
		return s
	}

	pairs := make([]string, len(matches)*2)
	for i, match := range matches {
		pairs[i*2] = match.Std()
		pairs[i*2+1] = ""
	}

	return String(strings.NewReplacer(pairs...).Replace(s.Std()))
}

// ReplaceNth returns a new String instance with the nth occurrence of oldS
// replaced with newS. If there aren't enough occurrences of oldS, the
// original String is returned. If n is less than -1, the original String
// is also returned. If n is -1, the last occurrence of oldS is replaced with newS.
//
// Returns:
//
// - A new String instance with the nth occurrence of oldS replaced with newS.
//
// Example usage:
//
//	s := g.String("The quick brown dog jumped over the lazy dog.")
//	result := s.ReplaceNth("dog", "fox", 2)
//	fmt.Println(result)
//
// Output: "The quick brown dog jumped over the lazy fox.".
func (s String) ReplaceNth(oldS, newS String, n Int) String {
	if n < -1 || len(oldS) == 0 {
		return s
	}

	count, i := Int(0), Int(0)

	for {
		pos := s[i:].Index(oldS)
		if pos == -1 {
			break
		}

		pos += i
		count++

		if count == n || (n == -1 && s[pos+oldS.Len():].Index(oldS) == -1) {
			var b Builder
			b.WriteString(s[:pos])
			b.WriteString(newS)
			b.WriteString(s[pos+oldS.Len():])

			return b.String()
		}

		i = pos + oldS.Len()
	}

	return s
}

// Contains checks if the String contains the specified substring.
func (s String) Contains(substr String) bool { return strings.Contains(s.Std(), substr.Std()) }

// ContainsAny checks if the String contains any of the specified substrings.
func (s String) ContainsAny(substrs ...String) bool {
	return slices.ContainsFunc(substrs, s.Contains)
}

// ContainsAll checks if the given String contains all the specified substrings.
func (s String) ContainsAll(substrs ...String) bool {
	for _, substr := range substrs {
		if !s.Contains(substr) {
			return false
		}
	}

	return true
}

// ContainsAnyChars checks if the String contains any characters from the specified String.
func (s String) ContainsAnyChars(chars String) bool {
	return strings.ContainsAny(s.Std(), chars.Std())
}

// StartsWith checks if the String starts with the specified prefix.
func (s String) StartsWith(prefix String) bool { return strings.HasPrefix(s.Std(), prefix.Std()) }

// StartsWithAny checks if the String starts with any of the provided prefixes.
// The method accepts a variable number of arguments, allowing for checking against multiple
// prefixes at once. It iterates over the provided prefixes and uses the HasPrefix function from
// the strings package to check if the String starts with each prefix.
// The function returns true if the String starts with any of the prefixes, and false otherwise.
//
// Example usage:
//
//	s := g.String("http://example.com")
//	if s.StartsWithAny("http://", "https://") {
//	   // do something
//	}
func (s String) StartsWithAny(prefixes ...String) bool {
	return slices.ContainsFunc(prefixes, s.StartsWith)
}

// EndsWith checks if the String ends with the specified suffix.
func (s String) EndsWith(suffix String) bool { return strings.HasSuffix(s.Std(), suffix.Std()) }

// EndsWithAny checks if the String ends with any of the provided suffixes.
// The method accepts a variable number of arguments, allowing for checking against multiple
// suffixes at once. It iterates over the provided suffixes and uses the HasSuffix function from
// the strings package to check if the String ends with each suffix.
// The function returns true if the String ends with any of the suffixes, and false otherwise.
//
// Example usage:
//
//	s := g.String("example.com")
//	if s.EndsWithAny(".com", ".net") {
//	   // do something
//	}
func (s String) EndsWithAny(suffixes ...String) bool {
	return slices.ContainsFunc(suffixes, s.EndsWith)
}

// Lines splits the String by lines, with trailing whitespace trimmed per
// line. The substrings share the receiver's backing memory.
//
// It returns a plain []String — deliberately not a Seq or Slice: a named
// generic type in the signature would make every package that names g.String
// compile the whole container machinery. Convert with g.SliceOf(parts...) for
// chaining; for lazy streaming over files use fs.File.Lines.
func (s String) Lines() []String {
	var result []String

	for line := range strings.Lines(s.Std()) {
		result = append(result, String(line).TrimEnd())
	}

	return result
}

// Fields splits the String around whitespace. See [String.Lines] for why the
// return type is a plain slice.
func (s String) Fields() []String {
	return castStrings(strings.Fields(s.Std()))
}

// FieldsBy splits the String using a custom function to determine the field
// boundaries. See [String.Lines] for why the return type is a plain slice.
func (s String) FieldsBy(fn func(r rune) bool) []String {
	return castStrings(strings.FieldsFunc(s.Std(), fn))
}

// Split splits the String by the specified separator. If sep is empty, the
// String is split after each UTF-8 rune. See [String.Lines] for why the return
// type is a plain slice.
func (s String) Split(sep String) []String {
	return castStrings(strings.Split(s.Std(), sep.Std()))
}

// SplitAfter splits the String after each instance of the specified separator.
// See [String.Lines] for why the return type is a plain slice.
func (s String) SplitAfter(sep String) []String {
	return castStrings(strings.SplitAfter(s.Std(), sep.Std()))
}

// SplitN splits the String into substrings using the provided separator and
// returns a plain []String of the results (convert with Slice[String] for
// chaining). The n parameter controls the number of substrings to return:
// - If n is negative, there is no limit on the number of substrings returned.
// - If n is zero, an empty slice is returned.
// - If n is positive, at most n substrings are returned.
func (s String) SplitN(sep String, n Int) []String {
	parts := strings.SplitN(s.Std(), sep.Std(), n.Std())

	result := make([]String, len(parts))
	for i, p := range parts {
		result[i] = String(p)
	}

	return result
}

// Chunks splits the String into chunks of the specified size.
//
// This function iterates through the String, creating new String chunks of the specified size.
// If size is less than or equal to 0 or the String is empty, it returns nil.
// If size is greater than or equal to the length of the String, it returns the
// original String as the only chunk.
//
// Parameters:
//
// - size (Int): The size of the chunks to split the String into.
//
// Returns:
//
// - []String: the String chunks of the specified size.
//
// Example usage:
//
//	text := g.String("Hello, World!")
//	chunks := text.Chunks(4)
//
// chunks contains {"Hell", "o, W", "orld", "!"}.
func (s String) Chunks(size Int) []String {
	if size.Lte(0) || s.IsEmpty() {
		return nil
	}

	if s.IsASCII() {
		n := size.Std()
		l := len(s)

		if n >= l {
			return []String{s}
		}

		result := make([]String, 0, (l+n-1)/n)
		for i := 0; i < l; i += n {
			result = append(result, s[i:min(i+n, l)])
		}

		return result
	}

	n := size.Std()

	var result []String

	rest := s
	for !rest.IsEmpty() {
		i := 0
		str := rest.Std()
		for count := 0; count < n && i < len(rest); count++ {
			_, sz := utf8.DecodeRuneInString(str[i:])
			i += sz
		}

		result = append(result, rest[:i])
		rest = rest[i:]
	}

	return result
}

// Cut returns two String values. The first String contains the remainder of the
// original String after the cut. The second String contains the text between the
// first occurrences of the 'start' and 'end' strings, with tags removed if specified.
//
// The function searches for the 'start' and 'end' strings within the String.
// If both are found, it returns the first String containing the remainder of the
// original String after the cut, followed by the second String containing the text
// between the first occurrences of 'start' and 'end' with tags removed if specified.
//
// If either 'start' or 'end' is empty or not found in the String, it returns the
// original String as the first String, and an empty String as the second.
//
// Parameters:
//
// - start (String): The String marking the beginning of the text to be cut.
//
// - end (String): The String marking the end of the text to be cut.
//
//   - rmtags (bool, optional): An optional boolean parameter indicating whether
//     to remove the matched region (including the 'start' and 'end' tags) from the
//     remainder. Defaults to false, in which case the remainder equals the original
//     String and only the cut content is extracted.
//
// Returns:
//
//   - String: The first String containing the remainder of the original String.
//     When rmtags is true the matched region is removed from it; otherwise it is
//     the original String. Returns the original String if 'start' or 'end' is empty or not found.
//
//   - String: The second String containing the text between the first occurrences of
//     'start' and 'end', or an empty String if 'start' or 'end' is empty or not found.
//
// Example usage:
//
//	s := g.String("Hello, [world]! How are you?")
//	remainder, cut := s.Cut("[", "]", true)
//	// remainder: "Hello, ! How are you?"
//	// cut: "world"
func (s String) Cut(start, end String, rmtags ...bool) (String, String) {
	if start.IsEmpty() || end.IsEmpty() {
		return s, ""
	}

	startIndex := s.Index(start)
	if startIndex == -1 {
		return s, ""
	}

	startEnd := startIndex + start.Len()
	endIndex := s[startEnd:].Index(end)
	if endIndex == -1 {
		return s, ""
	}

	cut := s[startEnd : startEnd+endIndex]

	if len(rmtags) == 0 || !rmtags[0] {
		return s, cut
	}

	return s[:startIndex] + s[startEnd+endIndex+end.Len():], cut
}

// Similarity calculates the similarity between two Strings using the
// Levenshtein distance algorithm and returns the similarity percentage as an Float.
//
// The function compares two Strings using the Levenshtein distance,
// which measures the difference between two sequences by counting the number
// of single-character edits required to change one sequence into the other.
// The similarity is then calculated by normalizing the distance by the maximum
// length of the two input Strings.
//
// Parameters:
//
// - str (String): The String to compare with s.
//
// Returns:
//
// - Float: The similarity percentage between the two Strings as a value between 0 and 100.
//
// Example usage:
//
//	s1 := g.String("kitten")
//	s2 := g.String("sitting")
//	similarity := s1.Similarity(s2) // 57.14285714285714
func (s String) Similarity(str String) Float {
	if s.Eq(str) {
		return 100
	}

	if s.IsEmpty() || str.IsEmpty() {
		return 0
	}

	s1 := s.Runes()
	s2 := str.Runes()

	n1, n2 := len(s1), len(s2)

	if n1 > n2 {
		s1, s2, n1, n2 = s2, s1, n2, n1
	}

	distance := make([]int, n1+1)

	for i, r2 := range s2 {
		prev := i + 1

		for j, r1 := range s1 {
			current := distance[j]
			if r2 != r1 {
				current = min(distance[j]+1, min(prev+1, distance[j+1]+1))
			}

			distance[j], prev = prev, current
		}

		distance[n1] = prev
	}

	return Float(1-float64(distance[n1])/float64(max(n1, n2))) * 100
}

// Cmp compares two Strings and returns an cmp.Ordering indicating their relative order.
// The result will be cmp.Equal if s==str, cmp.Less if s < str, and cmp.Greater if s > str.
func (s String) Cmp(str String) cmp.Ordering { return cmp.Cmp(s, str) }

// Append appends the specified String to the current String.
func (s String) Append(str String) String { return s + str }

// Prepend prepends the specified String to the current String.
func (s String) Prepend(str String) String { return str + s }

// ContainsRune checks if the String contains the specified rune.
func (s String) ContainsRune(r rune) bool { return strings.ContainsRune(s.Std(), r) }

// Count returns the number of non-overlapping instances of the substring in the String.
func (s String) Count(substr String) Int { return Int(strings.Count(s.Std(), substr.Std())) }

// IsEmpty checks if the String is empty.
func (s String) IsEmpty() bool { return len(s) == 0 }

// Eq checks if two Strings are equal.
func (s String) Eq(str String) bool { return s == str }

// EqFold compares two String strings case-insensitively.
func (s String) EqFold(str String) bool { return strings.EqualFold(s.Std(), str.Std()) }

// Gt checks if the String is greater than the specified String.
func (s String) Gt(str String) bool { return s > str }

// Gte checks if the String is greater than or equal to the specified String.
func (s String) Gte(str String) bool { return s >= str }

// Bytes returns the String as an Bytes.
func (s String) Bytes() Bytes { return Bytes(s) }

// BytesUnsafe converts the String into Bytes without copying memory.
// Warning: the resulting Bytes shares the same underlying memory as the original String.
// If the original String is modified through unsafe operations (rare), or if it is garbage collected,
// the Bytes may become invalid or cause undefined behavior.
func (s String) BytesUnsafe() Bytes { return unsafe.Slice(unsafe.StringData(s.Std()), len(s)) }

// Index returns the index of the first instance of the specified substring in the String, or -1
// if substr is not present in s.
func (s String) Index(substr String) Int { return Int(strings.Index(s.Std(), substr.Std())) }

// LastIndex returns the index of the last instance of the specified substring in the String, or -1
// if substr is not present in s.
func (s String) LastIndex(substr String) Int { return Int(strings.LastIndex(s.Std(), substr.Std())) }

// IndexRune returns the index of the first instance of the specified rune in the String.
func (s String) IndexRune(r rune) Int { return Int(strings.IndexRune(s.Std(), r)) }

// IndexByte returns the index of the first instance of the specified byte in the String, or -1
// if b is not present in s.
func (s String) IndexByte(b byte) Int { return Int(strings.IndexByte(s.Std(), b)) }

// LastIndexByte returns the index of the last instance of the specified byte in the String, or -1
// if b is not present in s.
func (s String) LastIndexByte(b byte) Int { return Int(strings.LastIndexByte(s.Std(), b)) }

// Len returns the length of the String.
func (s String) Len() Int { return Int(len(s)) }

// LenRunes returns the number of runes in the String.
func (s String) LenRunes() Int { return Int(utf8.RuneCountInString(s.Std())) }

// Lt checks if the String is less than the specified String.
func (s String) Lt(str String) bool { return s < str }

// Lte checks if the String is less than or equal to the specified String.
func (s String) Lte(str String) bool { return s <= str }

// Map applies the provided function to all runes in the String and returns the resulting String.
func (s String) Map(fn func(rune) rune) String { return String(strings.Map(fn, s.Std())) }

// Ne checks if two Strings are not equal.
func (s String) Ne(str String) bool { return !s.Eq(str) }

// Reader returns a *strings.Reader initialized with the content of String.
func (s String) Reader() *strings.Reader { return strings.NewReader(s.Std()) }

// Repeat returns a new String consisting of the specified count of the original String.
func (s String) Repeat(count Int) String { return String(strings.Repeat(s.Std(), count.Std())) }

// Reverse reverses the String.
func (s String) Reverse() String { return s.BytesUnsafe().Reverse().StringUnsafe() }

// Runes returns the String as a plain slice of runes.
func (s String) Runes() []rune { return []rune(s) }

// Chars splits the String into individual UTF-8 characters, equivalent to
// s.Split(""). Prefer Runes when only code points are needed.
func (s String) Chars() []String { return s.Split("") }

// SubString extracts a substring from the String starting at the 'start' index and ending before the 'end' index.
// The function also supports an optional 'step' parameter to define the increment between indices in the substring.
// If 'start' or 'end' index is negative, they represent positions relative to the end of the String:
// - A negative 'start' index indicates the position from the end of the String, moving backward.
// - A negative 'end' index indicates the position from the end of the String.
// The function ensures that indices are adjusted to fall within the valid range of the String's length.
// Out-of-bounds indices are clamped to the String's bounds instead of panicking;
// if 'start' exceeds 'end' (for a positive step) the result is an empty String.
func (s String) SubString(start, end Int, step ...Int) String {
	runes := s.Runes()
	n := Int(len(runes))

	clamp := func(i Int) Int {
		if i < 0 {
			i += n
		}

		if i < 0 {
			return 0
		}

		if i > n {
			return n
		}

		return i
	}

	start, end = clamp(start), clamp(end)

	st := Int(1)
	if len(step) > 0 {
		st = step[0]
	}

	// For a negative step the iteration starts AT start and moves down,
	// so a start clamped to n must begin at the last element.
	if st < 0 && start == n {
		start--
	}

	if st == 1 {
		if start >= end {
			return ""
		}

		return String(runes[start:end])
	}

	if (start >= end && st > 0) || (start <= end && st < 0) || st == 0 {
		return ""
	}

	var out []rune

	if st > 0 {
		for i := start; i < end; i += st {
			out = append(out, runes[i])
		}
	} else {
		for i := start; i > end; i += st {
			out = append(out, runes[i])
		}
	}

	return String(out)
}

// NormalizeNFC returns a new String with its Unicode characters normalized using the NFC form.
func (s String) NormalizeNFC() String { return String(norm.NFC.String(s.Std())) }

// Std returns the String as a string.
func (s String) Std() string { return string(s) }

// Format applies a specified format to the String object.
func (s String) Format(template String) String { return Format(template, s) }

// Truncate shortens the String to the specified maximum length. If the String exceeds the
// specified length, it is truncated, and an ellipsis ("...") is appended to indicate the truncation.
//
// If the length of the String is less than or equal to the specified maximum length, the
// original String is returned unchanged.
//
// The method respects Unicode characters and truncates based on the number of runes,
// not bytes.
//
// Parameters:
//   - max: The maximum number of runes allowed in the resulting String.
//
// Returns:
//   - A new String truncated to the specified maximum length with "..." appended
//     if truncation occurs. Otherwise, returns the original String.
//
// Example usage:
//
//	s := g.String("Hello, World!")
//	result := s.Truncate(5)
//	// result: "Hello..."
//
//	s2 := g.String("Short")
//	result2 := s2.Truncate(10)
//	// result2: "Short"
//
//	s3 := g.String("😊😊😊😊😊")
//	result3 := s3.Truncate(3)
//	// result3: "😊😊😊..."
func (s String) Truncate(max Int) String {
	if max.IsNegative() {
		return s
	}

	if s.IsASCII() {
		if Int(len(s)) <= max {
			return s
		}

		return s[:max].Append("...")
	}

	i := 0
	str := s.Std()
	for count := Int(0); i < len(s); count++ {
		if count == max {
			return s[:i].Append("...")
		}

		_, sz := utf8.DecodeRuneInString(str[i:])
		i += sz
	}

	return s
}

// LeftJustify justifies the String to the left by adding padding to the right, up to the
// specified length. If the length of the String is already greater than or equal to the specified
// length, or the pad is empty, the original String is returned.
//
// The padding String is repeated as necessary to fill the remaining length.
// The padding is added to the right of the String.
//
// Parameters:
//   - length: The desired length of the resulting justified String.
//   - pad: The String used as padding.
//
// Example usage:
//
//	s := g.String("Hello")
//	result := s.LeftJustify(10, "...")
//	// result: "Hello....."
func (s String) LeftJustify(length Int, pad String) String {
	rlen := s.LenRunes()
	if rlen >= length || pad.IsEmpty() {
		return s
	}

	var b Builder
	padlen := pad.LenRunes()
	b.Grow(s.Len() + paddingCapacity(pad, padlen, length-rlen))

	_, _ = b.WriteString(s)
	writePadding(&b, pad, padlen, length-rlen)

	return b.String()
}

// RightJustify justifies the String to the right by adding padding to the left, up to the
// specified length. If the length of the String is already greater than or equal to the specified
// length, or the pad is empty, the original String is returned.
//
// The padding String is repeated as necessary to fill the remaining length.
// The padding is added to the left of the String.
//
// Parameters:
//   - length: The desired length of the resulting justified String.
//   - pad: The String used as padding.
//
// Example usage:
//
//	s := g.String("Hello")
//	result := s.RightJustify(10, "...")
//	// result: ".....Hello"
func (s String) RightJustify(length Int, pad String) String {
	rlen := s.LenRunes()
	if rlen >= length || pad.IsEmpty() {
		return s
	}

	var b Builder
	padlen := pad.LenRunes()
	b.Grow(s.Len() + paddingCapacity(pad, padlen, length-rlen))

	writePadding(&b, pad, padlen, length-rlen)
	_, _ = b.WriteString(s)

	return b.String()
}

// Center justifies the String by adding padding on both sides, up to the specified length.
// If the length of the String is already greater than or equal to the specified length, or the
// pad is empty, the original String is returned.
//
// The padding String is repeated as necessary to evenly distribute the remaining length on both
// sides.
// The padding is added to the left and right of the String.
//
// Parameters:
//   - length: The desired length of the resulting justified String.
//   - pad: The String used as padding.
//
// Example usage:
//
//	s := g.String("Hello")
//	result := s.Center(10, "...")
//	// result: "..Hello..."
func (s String) Center(length Int, pad String) String {
	slen := s.LenRunes()
	if slen >= length || pad.IsEmpty() {
		return s
	}

	var b Builder

	padlen := pad.LenRunes()
	remains := length - slen
	b.Grow(s.Len() + paddingCapacity(pad, padlen, remains))

	writePadding(&b, pad, padlen, remains/2)
	_, _ = b.WriteString(s)
	writePadding(&b, pad, padlen, (remains+1)/2)

	return b.String()
}

func paddingCapacity(pad String, padlen, runes Int) Int {
	return ((runes + padlen - 1) / padlen) * pad.Len()
}

// writePadding writes the padding String to the output Builder to fill the remaining length.
// It repeats the padding String as necessary and appends any remaining runes from the padding
// String.
func writePadding(b *Builder, pad String, padlen, remains Int) {
	for range remains / padlen {
		_, _ = b.WriteString(pad)
	}

	rem := remains % padlen
	if rem == 0 {
		return
	}

	if pad.IsASCII() {
		for i := range rem {
			b.WriteByte(pad[i])
		}

		return
	}

	i := 0
	str := pad.Std()
	for range rem {
		r, sz := utf8.DecodeRuneInString(str[i:])
		_, _ = b.WriteRune(r)
		i += sz
	}
}

// Print writes the content of the String to the standard output (console)
// and returns the String unchanged.
func (s String) Print() String { fmt.Print(s); return s }

// Println writes the content of the String to the standard output (console) with a newline
// and returns the String unchanged.
func (s String) Println() String { fmt.Println(s); return s }

// Scan implements the database/sql.Scanner interface for g.String.
//
// Behavior:
//   - If src is nil, the String is set to an empty string.
//   - If src is a string, it is directly assigned.
//   - If src is a []byte, it is converted to a string.
//   - Otherwise, an error is returned.
//
// Supported SQL types (common):
//   - TEXT / VARCHAR → string
//   - BLOB / BYTEA  → []byte (converted to string)
//
// Notes:
//   - This method allows g.String to be used directly with database/sql and compatible drivers.
func (s *String) Scan(src any) error {
	if src == nil {
		*s = ""
		return nil
	}

	switch v := src.(type) {
	case string:
		*s = String(v)
		return nil
	case []byte:
		*s = String(v)
		return nil
	default:
		return fmt.Errorf("g.String.Scan: cannot scan %T into g.String", src)
	}
}

// Value implements the database/sql/driver.Valuer interface for g.String.
//
// Behavior:
//   - Returns the underlying string value, ready for database insertion.
//   - Always returns a value compatible with SQL TEXT / VARCHAR types.
func (s String) Value() (driver.Value, error) { return string(s), nil }

// castStrings reinterprets a []string as []String without copying: String is
// defined as `type String string`, so the two slice types share one memory
// layout.
func castStrings(ss []string) []String {
	return unsafe.Slice((*String)(unsafe.SliceData(ss)), len(ss))
}
