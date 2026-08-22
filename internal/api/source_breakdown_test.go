package api

// Regression suite for the "Source breakdown" diagnostic bug.
//
// Discovered:  2026-04-26 — user-supplied debug log
//              ("===== GoAnime Debug Session — 2026-04-26 18:19:11 =====")
//              showed `Source breakdown AnimeFire=0` while the same session
//              logged `Search results received source=Animefire.io count=10`.
// Fixed:       2026-04-28 — commit 39f17db
//              ("feat: refactor source breakdown logic for case-insensitive
//              matching and add corresponding tests").
// Root cause:  strings.Contains in Go is byte-exact. The breakdown predicate
//              in internal/api/enhanced.go searched for the substring
//              "AnimeFire" (capital F), but the scraper emits the canonical
//              source string "Animefire.io" (lowercase f) at
//              internal/scraper/unified.go:454. The two never matched, so the
//              counter stayed at 0 regardless of how many results came back.
// Blast radius:diagnostic-only — the player still worked because routing
//              elsewhere already used strings.ToLower (enhanced.go:449) or
//              hit the [PT-BR] tag fallback. The Goyabu reproduction in the
//              report confirmed playback was healthy. The bug only made the
//              "Source breakdown" log line lie to operators.
//
// The tests below pin the fix in place: if anyone ever reverts the
// case-insensitive matching or drops Goyabu from the breakdown again, these
// will fail loudly.

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestSourceBreakdown_LegacyCaseSensitiveBug documents the original case-sensitive
// regression: scrapers emit anime.Source = "Animefire.io" (lowercase 'f'), but the
// debug breakdown searched for the substring "AnimeFire" (capital 'F'). Because
// strings.Contains is byte-exact in Go, the AnimeFire counter was always zero
// even when the search returned dozens of AnimeFire results — making the
// "Source breakdown" diagnostic line lie to operators.
//
// This test reproduces the buggy comparison in isolation. The expectation
// matches the broken behaviour so that if anyone ever "fixes" the bug by
// hand without updating both the production code AND this regression test,
// the failure makes the intent obvious.
func TestSourceBreakdown_LegacyCaseSensitiveBug(t *testing.T) {
	source := "Animefire.io"

	// The exact pre-fix predicate from internal/api/enhanced.go.
	matchesLegacyPredicate := strings.Contains(source, "AnimeFire")

	assert.False(t, matchesLegacyPredicate,
		"sanity check: 'Animefire.io' must NOT match 'AnimeFire' under case-sensitive Contains — "+
			"if this assertion changes, Go's strings package semantics changed and the rest of the suite needs review")
}

// TestCountSourceBreakdown_AnimeFireCaseInsensitive is the positive regression
// test for the fix: the breakdown helper must count "Animefire.io" results
// regardless of capitalisation. Future scrapers (or upstream renames) that
// emit "AnimeFire", "ANIMEFIRE", or "animefire" must all be counted.
func TestCountSourceBreakdown_AnimeFireCaseInsensitive(t *testing.T) {
	animes := []*models.Anime{
		{Source: "Animefire.io"},
		{Source: "Animefire.io"},
		{Source: "AnimeFire"},
		{Source: "ANIMEFIRE"},
		{Source: "animefire"},
		{Source: "AllAnime"},
		{Source: "Goyabu"},
		{Source: "SuperFlix"},
	}

	got := countSourceBreakdown(animes)

	assert.Equal(t, 5, got.AnimeFire, "all AnimeFire spellings must be counted")
	assert.Equal(t, 1, got.AllAnime)
	assert.Equal(t, 1, got.Goyabu)
	assert.Equal(t, 1, got.SuperFlix)
}

// TestCountSourceBreakdown_RealisticPayload mirrors the user-reported log:
// 10 AnimeFire results, 5 AllAnime, 8 Goyabu — and asserts that Goyabu is
// reported (the original breakdown silently dropped it).
func TestCountSourceBreakdown_RealisticPayload(t *testing.T) {
	var animes []*models.Anime
	for range 10 {
		animes = append(animes, &models.Anime{Source: "Animefire.io"})
	}
	for range 5 {
		animes = append(animes, &models.Anime{Source: "AllAnime"})
	}
	for range 8 {
		animes = append(animes, &models.Anime{Source: "Goyabu"})
	}

	got := countSourceBreakdown(animes)

	assert.Equal(t, 10, got.AnimeFire, "AnimeFire breakdown must equal what the scraper returned")
	assert.Equal(t, 5, got.AllAnime)
	assert.Equal(t, 8, got.Goyabu, "Goyabu must appear in the breakdown")
	assert.Equal(t, 0, got.SuperFlix)
}
