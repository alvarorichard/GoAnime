// Package rand is the single home for randomness over g types, built on
// math/rand/v2.
//
//	rand.N(10)                    // Int-like value in [0, 10)
//	rand.Range(5, 10)             // half-open [5, 10)
//	rand.RangeInclusive(1, 6)     // closed [1, 6]
//	rand.Float()                  // Float in [0, 1)
//	rand.Chance(0.25)             // true with probability 0.25
//	rand.String(10)               // 10 alphanumeric characters
//	rand.Choice(users)            // Option with a random element
//	rand.Sample(deck, 5)          // 5 distinct random elements
//	rand.Shuffle(deck)            // in place; accepts Slice, MapOrd, plain slices
//	rand.SecureString(32)         // crypto/rand-backed token
//
// Every function is generic over its numeric arguments: integer parameters
// accept any integer type (int, g.Int, uint32, ...), float parameters any
// float type, and the result follows the argument's type where one exists.
//
// The container types deliberately carry NO random methods — everything lives
// here, one way to do it.
//
// The generators come from math/rand/v2 and are NOT cryptographically
// secure. For keys, tokens and anything security-sensitive use [SecureBytes]
// and [SecureString], which draw from crypto/rand.
package rand

import (
	"math/rand/v2"
	"strings"

	"github.com/enetx/g"
	"github.com/enetx/g/constraints"
)

// N returns a random integer in [0, n). It panics if n <= 0, matching
// math/rand/v2.N.
func N[T constraints.Integer](n T) T { return rand.N(n) }

// Range returns a random integer in the half-open interval [lo, hi).
// It panics if hi <= lo.
func Range[T constraints.Integer](lo, hi T) T {
	if hi <= lo {
		panic("rand.Range: hi must be greater than lo")
	}

	return lo + rand.N(hi-lo)
}

// RangeInclusive returns a random integer in the closed interval [lo, hi].
// The order of bounds does not matter (it normalizes to
// [min, max]); it works for negative bounds and the full int64 range without
// overflow or bias.
func RangeInclusive[T constraints.Integer](lo, hi T) T {
	if lo > hi {
		lo, hi = hi, lo
	}

	if lo == hi {
		return lo
	}

	const bias = uint64(1) << 63 // 2^63 = 9223372036854775808

	ulo := uint64(lo) + bias
	uhi := uint64(hi) + bias

	w := uhi - ulo + 1

	if w == 0 {
		return T(rand.Uint64())
	}

	randv := rand.N(w)
	result := int64((ulo + randv) - bias)

	return T(result)
}

// Float returns a random Float in [0, 1).
func Float() g.Float { return g.Float(rand.Float64()) }

// Uniform returns a random value in [lo, hi) of the arguments' float type.
func Uniform[T constraints.Float](lo, hi T) T { return lo + (hi-lo)*T(rand.Float64()) }

// NormFloat returns a normally distributed Float with mean 0 and standard
// deviation 1.
func NormFloat() g.Float { return g.Float(rand.NormFloat64()) }

// Bool returns true or false with equal probability.
func Bool() bool { return rand.N(2) == 0 }

// Chance returns true with probability p. Values outside [0, 1] clamp to
// always-false / always-true.
func Chance[T constraints.Float](p T) bool {
	if p <= 0 {
		return false
	}

	if p >= 1 {
		return true
	}

	return rand.Float64() < float64(p)
}

// Perm returns a random permutation of the integers [0, n) as a Slice of the
// argument's integer type.
func Perm[T constraints.Integer](n T) g.Slice[T] {
	if n <= 0 {
		return nil
	}

	result := make(g.Slice[T], n)
	for i, v := range rand.Perm(int(n)) {
		result[i] = T(v)
	}

	return result
}

// Choice returns a random element of the slice. An
// empty slice yields None. It accepts any slice-shaped type (Slice, MapOrd,
// plain slices).
func Choice[S ~[]E, E any](sl S) g.Option[E] {
	if len(sl) == 0 {
		return g.None[E]()
	}

	return g.Some(sl[rand.N(len(sl))])
}

// Choices returns k elements drawn WITH replacement.
// An empty source or non-positive k yields an empty result.
func Choices[S ~[]E, E any, I constraints.Integer](sl S, k I) S {
	if len(sl) == 0 || k <= 0 {
		return nil
	}

	result := make(S, k)
	for i := range result {
		result[i] = sl[rand.N(len(sl))]
	}

	return result
}

// Sample returns k distinct elements drawn WITHOUT replacement. If k is not
// less than the slice length, a shuffled copy of the whole
// slice is returned. The source slice is not modified.
func Sample[S ~[]E, E any, I constraints.Integer](sl S, k I) S {
	n := len(sl)

	if n == 0 || k <= 0 {
		return nil
	}

	if uint64(k) >= uint64(n) {
		out := make(S, n)
		copy(out, sl)
		Shuffle(out)

		return out
	}

	kk := int(k)

	// For small samples, track displaced indices in a map instead of copying
	// the whole slice: O(k) time and space.
	if float64(kk) < float64(n)*0.25 {
		result := make(S, kk)
		swapped := make(map[int]int, kk)

		for i := range kk {
			j := i + rand.N(n-i)

			vi, foundI := swapped[i]
			if !foundI {
				vi = i
			}

			vj, foundJ := swapped[j]
			if !foundJ {
				vj = j
			}

			swapped[i] = vj
			if i != j {
				swapped[j] = vi
			}

			result[i] = sl[vj]
		}

		return result
	}

	out := make(S, n)
	copy(out, sl)
	Shuffle(out)

	return out[:kk]
}

// Shuffle permutes the slice in place. It accepts any
// slice-shaped type (Slice, MapOrd, plain slices).
func Shuffle[S ~[]E, E any](sl S) {
	rand.Shuffle(len(sl), func(i, j int) { sl[i], sl[j] = sl[j], sl[i] })
}

// Bytes returns length random bytes. NOT cryptographically secure — use
// [SecureBytes] for keys and tokens.
func Bytes[I constraints.Integer](length I) g.Bytes {
	if length <= 0 {
		return nil
	}

	buf := make(g.Bytes, length)
	for i := range buf {
		buf[i] = byte(rand.N(256))
	}

	return buf
}

// String generates a random String of the specified length, selecting
// characters from predefined sets. If additional character sets are provided,
// only those are used; the default set (g.ASCIILetters and g.Digits) is
// excluded unless explicitly provided.
//
// If length is zero or negative, an empty String is returned. If an explicit
// letter set is provided but resolves to empty, an empty String is returned as
// well.
//
//	rand.String(10)          // 10 alphanumeric characters
//	rand.String(6, g.Digits) // 6-digit code
func String[I constraints.Integer](length I, letters ...g.String) g.String {
	if length <= 0 {
		return ""
	}

	if len(letters) != 0 {
		var buf g.Builder
		for _, set := range letters {
			_, _ = buf.WriteString(set)
		}

		chars := buf.String().Runes()
		n := len(chars)
		if n == 0 {
			return ""
		}

		var b strings.Builder
		b.Grow(int(length))
		for range int(length) {
			b.WriteRune(chars[rand.N(n)])
		}

		return g.String(b.String())
	}

	const charset = g.ASCIILetters + g.Digits
	n := len(charset)
	buf := make(g.Bytes, length)
	for i := range buf {
		buf[i] = charset[rand.N(n)]
	}

	return buf.StringUnsafe()
}
