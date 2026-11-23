package g

import (
	"encoding/base64"
	"encoding/hex"
	"html"
	"net/url"
	"strconv"

	json "encoding/json/v2"
)

type (
	// A struct that wraps Bytes for encoding.
	bencode struct{ bytes Bytes }

	// A struct that wraps Bytes for decoding.
	bdecode struct{ bytes Bytes }
)

// Encode returns a bencode struct wrapping the given Bytes.
func (bs Bytes) Encode() bencode { return bencode{bs} }

// Decode returns a bdecode struct wrapping the given Bytes.
func (bs Bytes) Decode() bdecode { return bdecode{bs} }

// Base64 encodes the wrapped Bytes using standard Base64 (with padding).
func (e bencode) Base64() Bytes {
	out := make(Bytes, base64.StdEncoding.EncodedLen(len(e.bytes)))
	base64.StdEncoding.Encode(out, e.bytes)
	return out
}

// Base64Raw encodes the wrapped Bytes using standard Base64 without padding.
func (e bencode) Base64Raw() Bytes {
	out := make(Bytes, base64.RawStdEncoding.EncodedLen(len(e.bytes)))
	base64.RawStdEncoding.Encode(out, e.bytes)
	return out
}

// Base64URL encodes the wrapped Bytes using URL-safe Base64 (with padding).
func (e bencode) Base64URL() Bytes {
	out := make(Bytes, base64.URLEncoding.EncodedLen(len(e.bytes)))
	base64.URLEncoding.Encode(out, e.bytes)
	return out
}

// Base64RawURL encodes the wrapped Bytes using URL-safe Base64 without padding.
func (e bencode) Base64RawURL() Bytes {
	out := make(Bytes, base64.RawURLEncoding.EncodedLen(len(e.bytes)))
	base64.RawURLEncoding.Encode(out, e.bytes)
	return out
}

// Base64 decodes the wrapped Bytes as standard Base64 (with padding) and returns Result[Bytes].
func (d bdecode) Base64() Result[Bytes] {
	out := make(Bytes, base64.StdEncoding.DecodedLen(len(d.bytes)))
	n, err := base64.StdEncoding.Decode(out, d.bytes)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(out[:n])
}

// Base64Raw decodes the wrapped Bytes as standard Base64 without padding and returns Result[Bytes].
func (d bdecode) Base64Raw() Result[Bytes] {
	out := make(Bytes, base64.RawStdEncoding.DecodedLen(len(d.bytes)))
	n, err := base64.RawStdEncoding.Decode(out, d.bytes)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(out[:n])
}

// Base64URL decodes the wrapped Bytes as URL-safe Base64 (with padding) and returns Result[Bytes].
func (d bdecode) Base64URL() Result[Bytes] {
	out := make(Bytes, base64.URLEncoding.DecodedLen(len(d.bytes)))
	n, err := base64.URLEncoding.Decode(out, d.bytes)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(out[:n])
}

// Base64RawURL decodes the wrapped Bytes as URL-safe Base64 without padding and returns Result[Bytes].
func (d bdecode) Base64RawURL() Result[Bytes] {
	out := make(Bytes, base64.RawURLEncoding.DecodedLen(len(d.bytes)))
	n, err := base64.RawURLEncoding.Decode(out, d.bytes)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(out[:n])
}

// Hex hex-encodes the wrapped Bytes and returns the result as Bytes.
func (e bencode) Hex() Bytes {
	out := make(Bytes, hex.EncodedLen(len(e.bytes)))
	hex.Encode(out, e.bytes)
	return out
}

// Hex hex-decodes the wrapped Bytes and returns the decoded result as Result[Bytes].
func (d bdecode) Hex() Result[Bytes] {
	out := make(Bytes, hex.DecodedLen(len(d.bytes)))
	n, err := hex.Decode(out, d.bytes)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(out[:n])
}

// XOR encodes the wrapped Bytes using XOR cipher with the given key.
//
// Warning: a repeating-key XOR cipher is not a security primitive and provides
// no real confidentiality. Use it only for lightweight obfuscation, never to
// protect secrets.
func (e bencode) XOR(key Bytes) Bytes {
	if len(key) == 0 {
		return e.bytes.Clone()
	}

	out := make(Bytes, len(e.bytes))
	for i, b := range e.bytes {
		out[i] = b ^ key[i%len(key)]
	}

	return out
}

// XOR decodes the wrapped Bytes using XOR cipher with the given key.
func (d bdecode) XOR(key Bytes) Bytes { return d.bytes.Encode().XOR(key) }

// Binary converts the wrapped Bytes to its binary representation as Bytes.
func (e bencode) Binary() Bytes {
	var b Builder
	b.Grow(e.bytes.Len() * 8)

	for _, c := range e.bytes {
		for bit := 7; bit >= 0; bit-- {
			b.WriteByte('0' + (c>>uint(bit))&1)
		}
	}

	return b.String().Bytes()
}

// Binary converts the wrapped binary Bytes back to raw Bytes as Result[Bytes].
func (d bdecode) Binary() Result[Bytes] {
	if len(d.bytes)%8 != 0 {
		return Err[Bytes](ErrInvalidBinaryLength)
	}

	out := make(Bytes, 0, len(d.bytes)/8)

	for i := 0; i+8 <= len(d.bytes); i += 8 {
		var b byte
		for j := range 8 {
			c := d.bytes[i+j]
			if c != '0' && c != '1' {
				return Err[Bytes](ErrInvalidBinaryDigit)
			}
			b = b<<1 | (c - '0')
		}

		out = append(out, b)
	}

	return Ok(out)
}

// JSON encodes the wrapped Bytes as a JSON string using encoding/json/v2 and
// returns the result as Result[Bytes].
//
// The bytes are treated as text, mirroring String.Encode().JSON.
//
// v2 semantics: Bytes containing invalid UTF-8 yield Err — use Base64/Hex
// encoding for arbitrary binary data.
// Unlike encoding/json v1, the output does not HTML-escape '<', '>', '&' or the
// line separators U+2028/U+2029 — they are emitted raw. Escape the output
// yourself before embedding it in HTML or <script> contexts.
func (e bencode) JSON() Result[Bytes] {
	jsonData, err := json.Marshal(string(e.bytes))
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(Bytes(jsonData))
}

// JSON decodes the wrapped JSON string using encoding/json/v2 and returns the
// result as Result[Bytes].
//
// v2 semantics: a JSON string containing invalid UTF-8 yields Err instead of
// being decoded with U+FFFD replacements.
func (d bdecode) JSON() Result[Bytes] {
	var data String
	err := json.Unmarshal(d.bytes, &data)
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(data.Bytes())
}

// URL encodes the wrapped Bytes, leaving the RFC 2396 reserved characters
// (";/?:@&=+$,") unescaped and query-escaping the rest. If safe characters are
// provided, they replace that default set and will not be encoded.
//
// Unlike String.Encode().URL, which matches safe characters by rune, matching
// here is byte-wise; with ASCII-only safe sets the output is identical to the
// String version for valid UTF-8 input. Non-UTF-8 bytes are percent-encoded
// verbatim, so the encoding is lossless for arbitrary binary input.
func (e bencode) URL(safe ...Bytes) Bytes {
	reserved := Bytes(";/?:@&=+$,")
	if len(safe) != 0 {
		reserved = safe[0]
	}

	out := make(Bytes, 0, len(e.bytes))

	for _, c := range e.bytes {
		if reserved.IndexByte(c) != -1 {
			out = append(out, c)
			continue
		}

		out = appendQueryEscaped(out, c)
	}

	return out
}

const upperhex = "0123456789ABCDEF"

// appendQueryEscaped appends c to dst using url.QueryEscape semantics:
// unreserved bytes (A-Z, a-z, 0-9, '-', '_', '.', '~') pass through unchanged,
// a space becomes '+', and every other byte is percent-encoded.
func appendQueryEscaped(dst Bytes, c byte) Bytes {
	switch {
	case 'A' <= c && c <= 'Z', 'a' <= c && c <= 'z', '0' <= c && c <= '9',
		c == '-', c == '_', c == '.', c == '~':
		return append(dst, c)
	case c == ' ':
		return append(dst, '+')
	default:
		return append(dst, '%', upperhex[c>>4], upperhex[c&0xF])
	}
}

// URL URL-decodes the wrapped Bytes and returns the decoded result as Result[Bytes].
func (d bdecode) URL() Result[Bytes] {
	result, err := url.QueryUnescape(string(d.bytes))
	if err != nil {
		return Err[Bytes](err)
	}

	return Ok(Bytes(result))
}

// HTML HTML-encodes the wrapped Bytes, escaping the characters <, >, &, ' and ".
// All other bytes, including non-UTF-8 sequences, pass through unchanged.
func (e bencode) HTML() Bytes { return Bytes(html.EscapeString(string(e.bytes))) }

// HTML HTML-decodes the wrapped Bytes, unescaping HTML entities.
// Bytes that are not part of an entity, including non-UTF-8 sequences, pass through unchanged.
func (d bdecode) HTML() Bytes { return Bytes(html.UnescapeString(string(d.bytes))) }

// Rot13 encodes the wrapped Bytes using the ROT13 cipher.
//
// The rotation is byte-wise over the ASCII letters A-Z and a-z; all other
// bytes, including those of multibyte UTF-8 runes, are left untouched, so the
// output is identical to String.Encode().Rot13 for valid UTF-8 input.
//
// WARNING: ROT13 is NOT a security primitive. It is a fixed letter-substitution
// cipher with no key and is trivially reversible. Use it only for obfuscation.
func (e bencode) Rot13() Bytes {
	out := make(Bytes, len(e.bytes))

	for i, c := range e.bytes {
		switch {
		case c >= 'A' && c <= 'Z':
			out[i] = 'A' + (c-'A'+13)%26
		case c >= 'a' && c <= 'z':
			out[i] = 'a' + (c-'a'+13)%26
		default:
			out[i] = c
		}
	}

	return out
}

// Rot13 decodes the wrapped Bytes using the ROT13 cipher.
func (d bdecode) Rot13() Bytes { return d.bytes.Encode().Rot13() }

// Octal returns the octal representation of the wrapped Bytes.
//
// Unlike String.Encode().Octal, which encodes Unicode code points, this
// implementation is byte-wise: each byte is rendered as its octal value
// (0-377), separated by spaces. The two representations match only for
// ASCII input.
func (e bencode) Octal() Bytes {
	var tmp [3]byte

	out := make(Bytes, 0, len(e.bytes)*4)

	for i, c := range e.bytes {
		if i != 0 {
			out = append(out, ' ')
		}

		out = append(out, strconv.AppendUint(tmp[:0], uint64(c), 8)...)
	}

	return out
}

// Octal decodes the octal representation back to Bytes.
// An empty input returns empty Bytes, mirroring bencode.Octal on empty input.
// Each space-separated token must represent a valid byte value in the octal
// range [0, 377]; anything else yields an error.
func (d bdecode) Octal() Result[Bytes] {
	if d.bytes.IsEmpty() {
		return Ok(Bytes(""))
	}

	out := make(Bytes, 0, (len(d.bytes)+1)/2)

	for _, v := range d.bytes.Split(Bytes(" ")) {
		n, err := strconv.ParseUint(v.StringUnsafe().Std(), 8, 8)
		if err != nil {
			return Err[Bytes](err)
		}

		out = append(out, byte(n))
	}

	return Ok(out)
}
