package pomfy

import (
	"context"
	"testing"
	"time"
)

// The reference vectors below were produced from the player bundle's own
// gr()/wr() implementation (the Go code is a translation of it), stamped
// 2026-11-02. If grHash drifts from the bundle, the captcha verify answers
// wrong and playback refuses every video, so any hash change must re-run the
// generator against a fresh bundle — these numbers are the contract.

type hashVector struct {
	msg  string
	hash string
	bits int
}

var hashVectors = []hashVector{
	{"nonce:0", "bf927bf123cf93eb9d08d02b221b6c3fb3d0efe7c79bfc617bb66c6a0b06f18a", 0},
	{"nonce:1", "395dc6f43ce60e57026f5fd4ea785f7337a5cd1b4f4adf400f9cfe3e858f3880", 2},
	{"nonce:199307", "776ad655df7ab2ef7b781de5b6d800f521a0a4ef4b9f2575180b23cb7efe2ac7", 1},
	{"abc:def", "ce39b93921d8e9f0d36a239aff4f0956cd30ded150fc1be07b51b7c522bfa81a", 0},
	{"a", "943cfd8c92c17927a34d9eb5602fdae4517a8003380b53018337d0491a31f1ed", 0},
	{"b420ebfcbb9b33a8b574d3b7d10071a5:42", "8cdf87239348da5c802a18ec19355f3b437cefb6a4d1ee7cc053f891c3e38880", 0},
	{"p/t-_.~:9", "165695965ac08e7a0205abf689d23dc1ed2d7189287058cba46f1d73f1fed450", 3},
}

type solveVector struct {
	prefix     string
	difficulty int
	solution   int
}

var solveVectors = []solveVector{
	{"ref:0", 6, 105},
	{"batata:0", 8, 184},
	{"pow-nonce_x:0", 5, 53},
	{"021da150ea1d:0", 7, 5},
}

func TestGrHashReferenceVectors(t *testing.T) {
	for _, v := range hashVectors {
		digest, bits := solveHashVector(v.msg)
		got := ""
		for _, w := range digest {
			got += hexWord(w)
		}
		if got != v.hash {
			t.Errorf("grHash(%q) = %s, want %s", v.msg, got, v.hash)
		}
		if bits != v.bits {
			t.Errorf("leadingZeroBits(%q) = %d, want %d", v.msg, bits, v.bits)
		}
	}
}

func hexWord(w uint32) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		out[7-i] = digits[w&0xF]
		w >>= 4
	}
	return string(out)
}

func TestSolveNonceReferenceVectors(t *testing.T) {
	for _, v := range solveVectors {
		// The happy path must answer immediately; no guard would mask a
		// hang, but a generous cap keeps a broken loop from pinning tests.
		ctx := t.Context()
		got, err := solveNonce(ctx, v.prefix, v.difficulty, 5*time.Second)
		if err != nil {
			t.Errorf("solveNonce(%q, %d) failed: %v", v.prefix, v.difficulty, err)
			continue
		}
		want := itoa(v.solution)
		if got != want {
			t.Errorf("solveNonce(%q, %d) = %q, want %q", v.prefix, v.difficulty, got, want)
		}
	}
}

func TestSolveNonceDifficultyZero(t *testing.T) {
	got, err := solveNonce(t.Context(), "anything", 0, time.Second)
	if err != nil || got != "0" {
		t.Errorf("difficulty 0 must answer 0 instantly, got %q err %v", got, err)
	}
}

func TestSolveNonceTimeout(t *testing.T) {
	// Difficulty 24 needs ~16M hashes — far beyond two milliseconds.
	start := time.Now()
	_, err := solveNonce(t.Context(), "timeout-case", 24, 5*time.Millisecond)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("solver ignored its deadline: %v", elapsed)
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := solveNonce(ctx, "cancel-case", 24, time.Second)
	if err == nil {
		t.Fatal("expected a cancellation error")
	}
}
