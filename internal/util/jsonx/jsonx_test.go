package jsonx

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

type sample struct {
	ID    int      `json:"id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

func TestUnmarshalBasic(t *testing.T) {
	var got sample
	if err := Unmarshal([]byte(`{"id":7,"title":"Naruto","tags":["a","b"]}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != 7 || got.Title != "Naruto" || len(got.Tags) != 2 {
		t.Fatalf("got %+v", got)
	}
}

// The three leniency options are the contract of this package; each gets a test
// so a future option change cannot pass silently.
func TestUnmarshalIsCaseInsensitive(t *testing.T) {
	var got sample
	if err := Unmarshal([]byte(`{"ID":7,"TITLE":"Naruto"}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != 7 || got.Title != "Naruto" {
		t.Fatalf("case-insensitive matching lost: %+v", got)
	}
}

func TestUnmarshalAllowsDuplicateNames(t *testing.T) {
	var got sample
	if err := Unmarshal([]byte(`{"id":1,"id":2}`), &got); err != nil {
		t.Fatalf("duplicate names must be accepted: %v", err)
	}
	if got.ID != 2 {
		t.Fatalf("last duplicate should win, got %d", got.ID)
	}
}

func TestUnmarshalAllowsInvalidUTF8(t *testing.T) {
	// 0xff is not valid UTF-8; v1 replaced it with U+FFFD and so must we.
	payload := []byte(`{"title":"na` + "\xff" + `ruto"}`)
	var got sample
	if err := Unmarshal(payload, &got); err != nil {
		t.Fatalf("invalid UTF-8 must be tolerated: %v", err)
	}
	if !strings.Contains(got.Title, "ruto") {
		t.Fatalf("unexpected title %q", got.Title)
	}
}

func TestUnmarshalRejectsMalformed(t *testing.T) {
	var got sample
	if err := Unmarshal([]byte(`{"id":`), &got); err == nil {
		t.Fatal("expected an error for truncated JSON")
	}
}

func TestDecodeStreamsWithinLimit(t *testing.T) {
	body := `{"id":42,"title":"Bleach","tags":["x"]}`
	var got sample
	if err := Decode(strings.NewReader(body), 1<<20, &got); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.ID != 42 || got.Title != "Bleach" {
		t.Fatalf("got %+v", got)
	}
}

func TestDecodeExactLimitIsAccepted(t *testing.T) {
	body := `{"id":1}`
	var got sample
	if err := Decode(strings.NewReader(body), int64(len(body)), &got); err != nil {
		t.Fatalf("a body of exactly limit bytes must be accepted: %v", err)
	}
	if got.ID != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestDecodeRejectsOversizedBody(t *testing.T) {
	body := `{"id":1,"title":"` + strings.Repeat("x", 4096) + `"}`
	var got sample
	err := Decode(strings.NewReader(body), 128, &got)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestDecodeOversizedBodyIsNotSilentlyTruncated(t *testing.T) {
	// The pre-1.27 io.ReadAll(io.LimitReader(...)) pattern turned an oversized
	// body into a truncated buffer and then an opaque syntax error. Decode must
	// name the real problem instead.
	body := `[` + strings.Repeat(`{"id":1},`, 500) + `{"id":2}]`
	var got []sample
	err := Decode(strings.NewReader(body), 64, &got)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
	if !strings.Contains(err.Error(), "64 bytes") {
		t.Errorf("error should mention the limit, got %q", err)
	}
}

func TestDecodeUnlimited(t *testing.T) {
	var got sample
	if err := Decode(strings.NewReader(`{"id":3}`), 0, &got); err != nil {
		t.Fatalf("Decode with no limit: %v", err)
	}
	if got.ID != 3 {
		t.Fatalf("got %+v", got)
	}
}

func TestDecodePropagatesReaderError(t *testing.T) {
	want := errors.New("boom")
	var got sample
	err := Decode(io.MultiReader(strings.NewReader(`{"id":`), errReader{want}), 1<<20, &got)
	if err == nil {
		t.Fatal("expected the reader error to surface")
	}
	if errors.Is(err, ErrTooLarge) {
		t.Fatalf("a reader failure must not be reported as ErrTooLarge: %v", err)
	}
}

func TestDecodeRejectsTrailingGarbage(t *testing.T) {
	var got sample
	if err := Decode(strings.NewReader(`{"id":1} trailing`), 1<<20, &got); err == nil {
		t.Fatal("expected an error for trailing data")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	in := sample{ID: 9, Title: "Frieren", Tags: []string{"fantasy"}}
	data, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sample
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("round trip mismatch: %+v vs %+v", out, in)
	}
}

type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }

// countingReader yields an endless JSON string and records how much of it was
// consumed.
type countingReader struct {
	read int64
	sent bool
}

func (c *countingReader) Read(p []byte) (int, error) {
	if !c.sent {
		c.sent = true
		n := copy(p, `{"title":"`)
		c.read += int64(n)
		return n, nil
	}
	for i := range p {
		p[i] = 'x'
	}
	c.read += int64(len(p))
	return len(p), nil
}

// TestDecodeStopsReadingAtLimit is the memory-safety guarantee behind the
// migration away from io.ReadAll: an endless response must be abandoned right
// after the cap, not drained.
func TestDecodeStopsReadingAtLimit(t *testing.T) {
	const limit = 64 << 10
	r := &countingReader{}
	var got sample
	err := Decode(r, limit, &got)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
	if r.read > limit+1 {
		t.Fatalf("read %d bytes for a %d byte limit; the cap is not being enforced", r.read, limit)
	}
}
