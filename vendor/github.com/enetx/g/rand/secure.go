package rand

import (
	crand "crypto/rand"
	"encoding/binary"

	"github.com/enetx/g"
	"github.com/enetx/g/constraints"
)

// SecureBytes returns length cryptographically secure random bytes drawn from
// crypto/rand. A zero or negative length yields nil.
//
// Unlike the rest of this package, the result is safe for keys, tokens and
// other security-sensitive material.
func SecureBytes[I constraints.Integer](length I) g.Bytes {
	if length <= 0 {
		return nil
	}

	buf := make(g.Bytes, length)
	if _, err := crand.Read(buf); err != nil {
		panic(err) // crypto/rand.Read does not fail on supported platforms
	}

	return buf
}

// SecureString generates a cryptographically secure random String of the
// specified length, selecting characters from predefined sets. If additional
// character sets are provided, only those are used; the default set
// (g.ASCIILetters and g.Digits) is excluded unless explicitly provided.
//
// If length is zero or negative, an empty String is returned. If an explicit
// letter set is provided but resolves to empty, an empty String is returned as
// well.
//
// Characters are drawn from crypto/rand with rejection sampling, so the
// selection is uniform (no modulo bias). Unlike [String], the result is safe
// for tokens, one-time codes and other security-sensitive material.
//
//	rand.SecureString(32)          // 32 alphanumeric characters
//	rand.SecureString(6, g.Digits) // 6-digit one-time code
func SecureString[I constraints.Integer](length I, letters ...g.String) g.String {
	if length <= 0 {
		return ""
	}

	var chars []rune
	if len(letters) != 0 {
		var buf g.Builder
		for _, set := range letters {
			_, _ = buf.WriteString(set)
		}

		chars = buf.String().Runes()
	} else {
		chars = (g.ASCIILetters + g.Digits).Runes()
	}

	n := len(chars)
	if n == 0 {
		return ""
	}

	var b g.Builder
	b.Grow(g.Int(length))

	for range int(length) {
		b.WriteRune(chars[secureIndex(n)])
	}

	return b.String()
}

// secureIndex returns a uniform random index in [0, n) drawn from crypto/rand,
// using rejection sampling to avoid modulo bias.
func secureIndex(n int) int {
	if n == 1 {
		return 0
	}

	// Values at or above limit fall into the biased tail and are redrawn.
	// A limit of zero means 2^32 is an exact multiple of n — nothing to reject.
	limit := uint32((1 << 32 / uint64(n)) * uint64(n))

	var buf [4]byte

	for {
		if _, err := crand.Read(buf[:]); err != nil {
			panic(err) // crypto/rand.Read does not fail on supported platforms
		}

		v := binary.BigEndian.Uint32(buf[:])
		if limit == 0 || v < limit {
			return int(v % uint32(n))
		}
	}
}
