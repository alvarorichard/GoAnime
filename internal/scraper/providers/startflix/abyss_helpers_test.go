package startflix

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// TestOrderedChunkHosts pins the sibling-failover order: primary first, then
// every offered host deduplicated, blanks dropped.
func TestOrderedChunkHosts(t *testing.T) {
	t.Parallel()
	c := &abyssChunkSource{chunkHost: "a.test"}
	if got := c.orderedChunkHosts(); len(got) != 1 || got[0] != "a.test" {
		t.Fatalf("primary only = %v", got)
	}
	c = &abyssChunkSource{
		chunkHost: "a.test",
		domains:   []string{"b.test", "", "a.test", "c.test", "b.test"},
	}
	want := []string{"a.test", "b.test", "c.test"}
	got := c.orderedChunkHosts()
	if len(got) != len(want) {
		t.Fatalf("hosts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("hosts = %v, want %v", got, want)
		}
	}
}

// TestIsTruncationError pins which failures fall back to sliced Range
// re-fetching: truncated transports yes, origin statuses no (those already
// had their bounded whole-part retries).
func TestIsTruncationError(t *testing.T) {
	t.Parallel()
	truncation := []error{
		&url.Error{Op: "Get", URL: "abyss-chunk", Err: io.ErrUnexpectedEOF},
		fmt.Errorf("abyss chunk returned %d bytes, expected %d", 100, 655360),
	}
	for _, err := range truncation {
		if !isTruncationError(err) {
			t.Errorf("%v: want truncation", err)
		}
	}
	notTruncation := []error{
		netx.NewHTTPStatusError(SourceName, "abyss chunk", http.StatusInternalServerError),
		netx.NewHTTPStatusError(SourceName, "abyss chunk", http.StatusNotFound),
		context.Canceled,
		context.DeadlineExceeded,
	}
	for _, err := range notTruncation {
		if isTruncationError(err) {
			t.Errorf("%v: want no truncation fallback", err)
		}
	}
}
