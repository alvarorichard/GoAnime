package pomfy

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// The Byse captcha's proof-of-work is a custom 256-bit hash the player bundle
// calls gr(), not SHA-256 (the API's "algorithm" field mislabels it). The
// code below is a line-by-line translation of that bundle so its output
// matches bit for bit; pow_test.go locks it with reference vectors produced
// from the original JavaScript.
//
// The task: for a server nonce pow_nonce and difficulty D, find s >= 0 whose
// hash of "pow_nonce:s" starts with D zero bits. The API hands out 16-18,
// which means about 2^D hashes — a second or two in Go.

const (
	qrMixLR  = 2654435761 // the bundle's lr, multiplied into the referrer word
	qrMixHR  = 2246822519 // hr, multiplied into the blended words
	grWords  = 512        // be: the size of the shuffled table
	grRounds = 2          // dr: how many shuffling passes run
)

// grInit is the bundle's initial four-word state; the same constants ChaCha20
// seeds with.
var grInit = [4]uint32{1779033703, 3144134277, 1013904242, 2773480762}

// rotl32 is the bundle's re(): a left rotation on 32-bit words.
func rotl32(x uint32, n uint) uint32 { return x<<n | x>>(32-n) }

// chachaQuarterRound is the bundle's ye(): ChaCha20's quarter round with the
// rotates in the (16, 12, 8, 7) ladder.
func chachaQuarterRound(e *[4]uint32) {
	e[0] += e[1]
	e[3] = rotl32(e[3]^e[0], 16)
	e[2] += e[3]
	e[1] = rotl32(e[1]^e[2], 12)
	e[0] += e[1]
	e[3] = rotl32(e[3]^e[0], 8)
	e[2] += e[3]
	e[1] = rotl32(e[1]^e[2], 7)
}

// grHash is the bundle's gr(): a 256-bit digest of q. Every byte stirs into
// e[0] (added, rotated 7, then a quarter round over the whole state); the
// state then fills a 512-word table one quarter round per entry, the table is
// shuffled twice in place (each entry feeding from one its neighbors name,
// with e[0] absorbing the result), and finally eight quarter rounds blend 64
// table words each into an output word, every one XORed against e[2].
func grHash(q []byte) [8]uint32 {
	e := grInit
	for _, b := range q {
		e[0] += uint32(b)
		e[0] = rotl32(e[0], 7)
		chachaQuarterRound(&e)
	}
	for i := 0; i < 8; i++ {
		chachaQuarterRound(&e)
	}
	var r [grWords]uint32
	for i := range r {
		chachaQuarterRound(&e)
		r[i] = e[0] ^ e[2]
	}
	for i := 0; i < grRounds; i++ {
		for s := 0; s < grWords; s++ {
			a := r[s] & (grWords - 1) // the bundle's lt = 511
			c := r[s] + r[a]
			c = rotl32(c, 13)
			c ^= r[(s+1)&(grWords-1)] * qrMixLR
			r[s] = c
			e[0] ^= c
			chachaQuarterRound(&e)
		}
	}
	var out [8]uint32
	const span = grWords / 8
	for i := range out {
		chachaQuarterRound(&e)
		s := e[0]
		for c := 0; c < span; c++ {
			d := r[i*span+c]
			s += d
			s = rotl32(s, 5)
			s ^= d * qrMixHR
		}
		out[i] = s ^ e[2]
	}
	return out
}

// leadingZeroBits counts the zero bits in front of the digest, reading it the
// way the bundle's wr() does: every word contributes its leading zeros, a zero
// word counts 32 and the scan continues into the next one.
func leadingZeroBits(digest [8]uint32) int {
	total := 0
	for _, w := range digest {
		total += bitsLeadingZeros(w)
		if w != 0 {
			break
		}
	}
	return total
}

// bitsLeadingZeros is bits.LeadingZeros32 spelled out so the hot path keeps no
// new imports: successive shifts until the top bit sets.
func bitsLeadingZeros(w uint32) int {
	if w == 0 {
		return 32
	}
	n := 0
	for w&(1<<31) == 0 {
		n++
		w <<= 1
	}
	return n
}

// solveNonce scans candidates s = 0, 1, 2, … until the digest of prefix:s
// carries at least difficulty leading zero bits. ctx cancels the hunt and the
// timeout caps it even without a live context deadline. The server accepts
// any meeting nonce, so whatever lands first is returned.
func solveNonce(ctx context.Context, prefix string, difficulty int, timeout time.Duration) (string, error) {
	if difficulty <= 0 {
		return "0", nil
	}
	deadline := time.Now().Add(timeout)
	until := []byte{}
	until = append(until, prefix...)
	failed := ctx.Err() != nil
	if failed {
		return "", ctx.Err()
	}
	for s := 0; ; s++ {
		h := grHash(appendDecimal(until, s))
		if leadingZeroBits(h) >= difficulty {
			return itoa(s), nil
		}
		// One grHash is thousands of quarter rounds (and far slower under the
		// race detector), so reading the clock after every candidate costs
		// nothing next to it; batching the check let a deadline slip by
		// seconds.
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if time.Now().After(deadline) {
			return "", errPoWTimeout
		}
	}
}

// timeoutError marks a solver that ran out of budget without finding a nonce
// that met the difficulty. Callers surface it as a retriable failure.
type timeoutError struct{ msg string }

func (e *timeoutError) Error() string { return e.msg }

var errPoWTimeout error = &timeoutError{msg: "the byse proof-of-work took too long to solve"} //nolint:err113 // fixed package error

// solveHashVector is the test entry point: one lookup of a prefix's digest and
// its leading zero count, as the reference vectors name them.
func solveHashVector(prefix string) ([8]uint32, int) {
	h := grHash([]byte(prefix))
	return h, leadingZeroBits(h)
}

// checkNonce is the test oracle for one candidate: does the digest of
// prefix:s meet the difficulty?
func checkNonce(prefix, s string, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}
	return leadingZeroBits(grHash([]byte(prefix+s))) >= difficulty
}

// randomHex returns n random bytes as lowercase hex, seeding the fake
// fingerprint fields with. crypto/rand.Read never fails on the supported
// platforms (its 1.24+ contract), so no fallback dance runs here.
func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// appendDecimal appends v's decimal digits to buf (the solver's hot append).
func appendDecimal(buf []byte, v int) []byte {
	if v == 0 {
		return append(buf, '0')
	}
	var digits [12]byte
	n := 0
	for v > 0 {
		digits[n] = byte('0' + v%10)
		v /= 10
		n++
	}
	for n > 0 {
		n--
		buf = append(buf, digits[n])
	}
	return buf
}

// itoa is appendDecimal on a fresh buffer.
func itoa(v int) string { return string(appendDecimal(nil, v)) }
