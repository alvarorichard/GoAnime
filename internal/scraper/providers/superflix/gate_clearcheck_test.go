package superflix

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A cleared Cloudflare gate always leaves cookies, and a page with nothing in
// it is not a page.
//
// Both halves of the clear check were satisfiable by nothing at all. The check
// is "the document has no challenge markers", and an empty document has none —
// so a failed navigation that left the tab blank read as success. The solver
// then reported "gate cleared … cookies=0", the caller replayed a clearance
// that did not exist, got 403, and the 403 was read for hours as a clearance
// that would not transfer.

func TestPageLooksReal_RejectsAnEmptyDocument(t *testing.T) {
	t.Parallel()
	blank := []string{
		"",
		"   \n\t ",
		"<html><head></head><body></body></html>",
		"<!DOCTYPE html><html><head></head><body></body></html>",
	}
	for _, html := range blank {
		assert.Falsef(t, pageLooksReal(html),
			"an empty document has no challenge markers either, so it read as a cleared gate: %q", html)
	}
}

func TestPageLooksReal_AcceptsARealPage(t *testing.T) {
	t.Parallel()
	// Even a small real page is far heavier than a blank one; an interstitial
	// is heavier still (measured ~5.5KB, and ~28KB once rendered).
	page := "<!DOCTYPE html><html><body>" + strings.Repeat("<div class=\"card\">anime</div>", 40) + "</body></html>"
	assert.True(t, pageLooksReal(page))

	interstitial := "<html><head><title>Just a moment...</title></head><body>" +
		strings.Repeat("<script>x</script>", 60) + "</body></html>"
	assert.True(t, pageLooksReal(interstitial),
		"an interstitial is a real document; it is the MARKER check that must reject it, not this one")
	assert.True(t, bodyHasChallengeMarker([]byte(interstitial)),
		"and the marker check is what rejects it")
}
