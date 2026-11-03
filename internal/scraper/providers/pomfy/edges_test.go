package pomfy

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The edge matrix: every branch left after the flow and circuit files —
// the devalue guards, the PoW oracle edges, the search pool's degradation
// path and the small shapers' failures.

// --- devalue guards ---

func TestDevalueRefusals(t *testing.T) {
	arr := []json.RawMessage{
		json.RawMessage(`{"a":1}`),
		json.RawMessage(`[1,2]`),
		json.RawMessage(`"just a string"`),
		json.RawMessage(`42`),
	}
	g := newDevalueGraph(arr)

	// out-of-range: every accessor answers its zero value, no panic.
	if _, ok := g.object(99); ok {
		t.Error("out-of-range object must not resolve")
	}
	if g.array(99) != nil {
		t.Error("out-of-range array must not resolve")
	}
	if s, ok := g.stringRef(99); ok || s != "" {
		t.Errorf("out-of-range string = %q ok=%v, want empty/false", s, ok)
	}
	if got := g.integerRef(99); got != 0 {
		t.Errorf("out-of-range integer = %d, want 0", got)
	}

	// shape mismatches: a string where an object is expected, and the like.
	if _, ok := g.object(2); ok {
		t.Error("a string at that index must not be an object")
	}
	if g.array(0) != nil {
		t.Error("an object at that index must not be an array")
	}
	if _, ok := g.stringRef(3); ok {
		t.Error("a number at that index must not be a string")
	}
	if got := g.integerRef(2); got != 0 {
		t.Errorf("a string at that index must not be an integer: %d", got)
	}
}

// --- PoW oracle edges ---

func TestBitsLeadingZerosZeroWord(t *testing.T) {
	if got := bitsLeadingZeros(0); got != 32 {
		t.Errorf("zeros = %d, want 32", got)
	}
	if got := bitsLeadingZeros(0x80000000); got != 0 {
		t.Errorf("top-bit word = %d, want 0", got)
	}
	digest := [8]uint32{0, 0x80000000}
	if got := leadingZeroBits(digest); got != 32 {
		t.Errorf("zero word then bit = %d, want 32", got)
	}
}

func TestCheckNonceDifficultyZero(t *testing.T) {
	if !checkNonce("anything", "", 0) {
		t.Error("difficulty 0 accepts everything")
	}
}

func TestTimeoutErrorText(t *testing.T) {
	if msg := errPoWTimeout.Error(); msg == "" {
		t.Error("the timeout error must speak")
	}
}

func TestSolveNonceDeadlineBranch(t *testing.T) {
	start := time.Now()
	_, err := solveNonce(context.Background(), "deadline-case", 25, 1*time.Millisecond)
	if !errors.Is(err, errPoWTimeout) {
		t.Fatalf("err = %v, want the package timeout error", err)
	}
	if time.Since(start) > time.Second {
		t.Error("the deadline fired late")
	}
}

// --- the search pool's degradation path ---

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.EOF }

func TestShuffledIndexesCryptoDeath(t *testing.T) {
	prior := rand.Reader
	rand.Reader = failingReader{}
	out := shuffledIndexes(4)
	rand.Reader = prior
	if len(out) != 4 {
		t.Errorf("degraded shuffle = %v, want a 4-slot identity", out)
	}
	for i, v := range out {
		if v != i {
			t.Fatalf("the degraded path must keep the identity order: %v", out)
		}
	}
}

func TestSearchMultiWithKeyTransportAndGarbage(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := dead.URL
	dead.Close()
	resetSearchCache(t)
	old := tmdbBase
	tmdbBase = deadURL
	if _, err := searchMultiWithKey(PublicKeyForTesting(0), "x"); err == nil {
		t.Error("a dead base must error the search")
	}
	tmdbBase = old

	garbage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>not tmdb</html>`))
	}))
	t.Cleanup(garbage.Close)
	resetSearchCache(t)
	tmdbBase = garbage.URL
	if _, err := searchMultiWithKey(PublicKeyForTesting(0), "x"); err == nil {
		t.Error("garbage for the search parser must error")
	}
	tmdbBase = old
}

// --- the small shapers ---

func TestDecryptBysePlaybackDecryptedGarbage(t *testing.T) {
	seal := SealForTest([]byte("this is not the record's JSON"))
	var wrapped struct {
		Playback playbackSeal `json:"playback"`
	}
	if err := json.Unmarshal(seal, &wrapped); err != nil {
		t.Fatal(err)
	}
	if _, err := decryptBysePlayback(&wrapped.Playback); err == nil ||
		!strings.Contains(err.Error(), "decrypted payload") {
		t.Fatalf("err = %v, want the decrypted-payload parse error", err)
	}
}

func TestParseMovieTitleWithoutTipo(t *testing.T) {
	shape := compact(`{"nodes":[
		{"data":[]},
		{"data":[
			{"no":1},
			{"":2}
		]}
	]}`)
	if _, err := parseMovieTitle([]byte(shape)); err == nil {
		t.Fatal("a payload with no tipo node must fail loudly")
	}
}

func TestDecodeTitleNodeShortYearAndNoImage(t *testing.T) {
	arr := []json.RawMessage{
		json.RawMessage(`{"tipo":1}`),
		json.RawMessage(`{"tipo":2,"title":3,"year":4,"overview":5,"image":6,"availableSeasons":7}`),
		json.RawMessage("x"),
		json.RawMessage(`"Title"`),
		json.RawMessage(`"199"`),
		json.RawMessage(`"o"`),
		json.RawMessage(`""`),
		json.RawMessage(`[0]`),
	}
	out := decodeTitleNode(arr, 1)
	if out.Year != "" {
		t.Errorf("a 3-char year must not ride: %q", out.Year)
	}
	if out.Poster != "" {
		t.Errorf("an empty image must land empty: %q", out.Poster)
	}
}

func TestUserTMDBClientFnStillReal(t *testing.T) {
	if userTMDBClientFn() == nil {
		t.Error("the default factory must hand a real client back")
	}
}
