package g

import (
	"fmt"
	"html"
	"net/url"
	"strconv"
	"unicode/utf8"

	json "encoding/json/v2"
)

type (
	// A struct that wraps an String for encoding.
	encode struct{ str String }

	// A struct that wraps an String for decoding.
	decode struct{ str String }
)

// Encode returns an encode struct wrapping the given String.
func (s String) Encode() encode { return encode{s} }

// Decode returns a decode struct wrapping the given String.
func (s String) Decode() decode { return decode{s} }

// Base64 encodes the wrapped String using standard Base64 (with padding).
func (e encode) Base64() String { return e.str.BytesUnsafe().Encode().Base64().StringUnsafe() }

// Base64Raw encodes the wrapped String using standard Base64 without padding.
func (e encode) Base64Raw() String { return e.str.BytesUnsafe().Encode().Base64Raw().StringUnsafe() }

// Base64URL encodes the wrapped String using URL-safe Base64 (with padding).
func (e encode) Base64URL() String { return e.str.BytesUnsafe().Encode().Base64URL().StringUnsafe() }

// Base64RawURL encodes the wrapped String using URL-safe Base64 without padding.
func (e encode) Base64RawURL() String {
	return e.str.BytesUnsafe().Encode().Base64RawURL().StringUnsafe()
}

// Base64 decodes the wrapped String using standard Base64 (with padding).
func (d decode) Base64() Result[String] {
	return d.str.BytesUnsafe().Decode().Base64().Map(Bytes.String)
}

// Base64Raw decodes the wrapped String using standard Base64 without padding.
func (d decode) Base64Raw() Result[String] {
	return d.str.BytesUnsafe().Decode().Base64Raw().Map(Bytes.String)
}

// Base64URL decodes the wrapped String using URL-safe Base64 (with padding).
func (d decode) Base64URL() Result[String] {
	return d.str.BytesUnsafe().Decode().Base64URL().Map(Bytes.String)
}

// Base64RawURL decodes the wrapped String using URL-safe Base64 without padding.
func (d decode) Base64RawURL() Result[String] {
	return d.str.BytesUnsafe().Decode().Base64RawURL().Map(Bytes.String)
}

// Hex hex-encodes the wrapped String and returns the encoded result as an String.
func (e encode) Hex() String { return e.str.BytesUnsafe().Encode().Hex().StringUnsafe() }

// Hex hex-decodes the wrapped String and returns the decoded result as Result[String].
func (d decode) Hex() Result[String] {
	return d.str.BytesUnsafe().Decode().Hex().Map(Bytes.String)
}

// XOR encodes the wrapped String using a repeating-key XOR cipher with the given key.
//
// WARNING: XOR is NOT a security primitive. A repeating-key XOR cipher provides no
// confidentiality against any serious analysis and offers no integrity or
// authentication. Use it only for lightweight obfuscation, never to protect
// sensitive data.
func (e encode) XOR(key String) String {
	return String(e.str.BytesUnsafe().Encode().XOR(key.BytesUnsafe()))
}

// XOR decodes the wrapped String using a repeating-key XOR cipher with the given key.
//
// WARNING: XOR is NOT a security primitive. See encode.XOR for details.
func (d decode) XOR(key String) String { return d.str.Encode().XOR(key) }

// Binary converts the wrapped String to its binary representation.
func (e encode) Binary() String { return e.str.BytesUnsafe().Encode().Binary().StringUnsafe() }

// Binary converts the wrapped binary String back to its original String.
func (d decode) Binary() Result[String] {
	return d.str.BytesUnsafe().Decode().Binary().Map(Bytes.String)
}

// JSON encodes the provided string as a JSON string using encoding/json/v2 and
// returns the result as Result[String].
//
// Breaking change (v2 semantics): a String containing invalid UTF-8 now yields
// Err. Previously (encoding/json v1) invalid sequences were silently replaced
// with the Unicode replacement character (U+FFFD), making the encoding lossy.
// Unlike encoding/json v1, the output does not HTML-escape '<', '>', '&' or the
// line separators U+2028/U+2029 — they are emitted raw. Escape the output
// yourself before embedding it in HTML or <script> contexts.
func (e encode) JSON() Result[String] {
	jsonData, err := json.Marshal(e.str)
	if err != nil {
		return Err[String](err)
	}

	return Ok(String(jsonData))
}

// JSON decodes the provided JSON string using encoding/json/v2 and returns the
// result as Result[String].
//
// v2 semantics: a JSON string containing invalid UTF-8 yields Err instead of
// being decoded with U+FFFD replacements.
func (d decode) JSON() Result[String] {
	var data String
	err := json.Unmarshal(d.str.BytesUnsafe(), &data)
	if err != nil {
		return Err[String](err)
	}

	return Ok(data)
}

// URL encodes the input string, leaving the RFC 2396 reserved characters
// (";/?:@&=+$,") unescaped and query-escaping the rest. If safe characters are
// provided, they replace that default set and will not be encoded.
//
// Note: the default reserved set leaves '+' unescaped, while decoding maps '+'
// to a space (application/x-www-form-urlencoded), so Encode -> Decode is lossy
// for input containing a literal '+'. Pass a custom safe set to escape it.
func (e encode) URL(safe ...String) String {
	reserved := String(";/?:@&=+$,")
	if len(safe) != 0 {
		reserved = safe[0]
	}

	out := make(Bytes, 0, e.str.Len())

	for _, r := range e.str {
		if reserved.ContainsRune(r) {
			out = Bytes(utf8.AppendRune(out, r))
			continue
		}

		if r < utf8.RuneSelf {
			out = appendQueryEscaped(out, byte(r))
			continue
		}

		var enc [utf8.UTFMax]byte
		n := utf8.EncodeRune(enc[:], r)
		for _, c := range enc[:n] {
			out = appendQueryEscaped(out, c)
		}
	}

	return out.StringUnsafe()
}

// URL URL-decodes the wrapped String and returns the decoded result as Result[String].
func (d decode) URL() Result[String] {
	result, err := url.QueryUnescape(d.str.Std())
	if err != nil {
		return Err[String](err)
	}

	return Ok(String(result))
}

// HTML HTML-encodes the wrapped String.
func (e encode) HTML() String { return String(html.EscapeString(e.str.Std())) }

// HTML HTML-decodes the wrapped String.
func (d decode) HTML() String { return String(html.UnescapeString(d.str.Std())) }

// Rot13 encodes the wrapped String using the ROT13 cipher.
//
// WARNING: ROT13 is NOT a security primitive. It is a fixed letter-substitution
// cipher with no key and is trivially reversible. Use it only for obfuscation.
func (e encode) Rot13() String {
	rot := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		default:
			return r
		}
	}

	return e.str.Map(rot)
}

// Rot13 decodes the wrapped String using ROT13 cipher.
func (d decode) Rot13() String { return d.str.Encode().Rot13() }

// Octal returns the octal representation of the encoded string.
func (e encode) Octal() String {
	var b Builder
	var tmp [7]byte

	first := true

	for _, char := range e.str {
		if !first {
			b.WriteByte(' ')
		}

		_, _ = b.Write(strconv.AppendInt(tmp[:0], int64(char), 8))
		first = false
	}

	return b.String()
}

// Octal decodes the octal representation back to String.
// An empty input returns an empty String, mirroring encode.Octal("").
// Each token must represent a valid Unicode code point in the range [0, MaxRune].
func (d decode) Octal() Result[String] {
	if d.str.IsEmpty() {
		return Ok(String(""))
	}

	var b Builder

	for _, v := range d.str.Split(" ") {
		n, err := strconv.ParseUint(v.Std(), 8, 32)
		if err != nil {
			return Err[String](err)
		}

		if n > utf8.MaxRune || (n >= 0xD800 && n <= 0xDFFF) {
			return Err[String](fmt.Errorf("g.decode.Octal: invalid code point %d", n))
		}

		b.WriteRune(rune(n))
	}

	return Ok(b.String())
}
