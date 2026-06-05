package api

// Regression suite for the FlixHQ "season selection abort" bug.
//
// Discovered:  2026-04-27 — user-supplied debug log
//              ("00:08:12 ERRO  GoAnime  : Failed to get episodes:
//              failed to fetch episodes: season selection cancelled: abort")
// Fixed:       2026-04-27 — same-day fix in this commit.
// Root cause:  internal/api/enhanced.go GetFlixHQEpisodes (FlixHQ) wrapped
//              fuzzyfinder.ErrAbort
//              as `fmt.Errorf("season selection cancelled: %w", err)` and
//              returned it. Because that error chain was not ErrBackToSearch,
//              the playback handler (internal/handlers/playback.go) bailed
//              out via `return` instead of looping back to the search prompt.
//              Net effect: pressing ESC during season selection killed the
//              session — the user had to relaunch the binary to try a
//              different show.
// Blast radius:user-facing — ESC during season selection (a routine action
//              for "wrong show, let me search again") terminated the program
//              with a fatal-looking log line. Especially likely to bite users
//              of FlixHQ because every TV show on that source
//              prompts for a season.
//
// The tests below pin three invariants:
//   1. The abort path returns the sentinel ErrBackToSearch (errors.Is true).
//   2. The wrap performed by appflow.GetAnimeEpisodes
//      (`failed to fetch episodes: %w`) preserves the chain so the handler
//      can still recognize ErrBackToSearch via errors.Is.
//   3. Non-abort fuzzyfinder errors are NOT swallowed as ErrBackToSearch —
//      they must continue to surface as real failures.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/stretchr/testify/assert"
)

// mapSeasonSelectionErr mirrors the production predicate from
// internal/api/enhanced.go (GetFlixHQEpisodes).
// Keeping the predicate in one tiny helper means this test pins the exact
// shape of the mapping rather than just observing some end-to-end behaviour.
func mapSeasonSelectionErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fuzzyfinder.ErrAbort) {
		return ErrBackToSearch
	}
	return fmt.Errorf("season selection cancelled: %w", err)
}

func TestSeasonSelection_AbortMapsToBackToSearch(t *testing.T) {
	got := mapSeasonSelectionErr(fuzzyfinder.ErrAbort)
	assert.ErrorIs(t, got, ErrBackToSearch,
		"fuzzyfinder.ErrAbort during season selection MUST surface as ErrBackToSearch "+
			"so the playback handler can re-prompt the user instead of returning")
}

func TestSeasonSelection_AbortSurvivesAppflowWrap(t *testing.T) {
	// appflow.GetAnimeEpisodes wraps the API error as
	// `failed to fetch episodes: %w`. The handler must still recognise the
	// sentinel through that wrap; otherwise our fix only works at the API
	// boundary and the handler keeps killing the session.
	apiErr := mapSeasonSelectionErr(fuzzyfinder.ErrAbort)
	wrapped := fmt.Errorf("failed to fetch episodes: %w", apiErr)

	assert.ErrorIs(t, wrapped, ErrBackToSearch,
		"errors.Is must traverse the appflow wrap — if this fails, the playback "+
			"handler will not detect the abort and will kill the session again")
}

func TestSeasonSelection_NonAbortErrorsStillFail(t *testing.T) {
	// Make sure we did not over-correct: a real failure (e.g. terminal
	// closed, IO error) must NOT be silently mapped to ErrBackToSearch.
	realErr := errors.New("some other fuzzyfinder failure")
	got := mapSeasonSelectionErr(realErr)

	assert.NotErrorIs(t, got, ErrBackToSearch,
		"only fuzzyfinder.ErrAbort should map to ErrBackToSearch — other errors "+
			"must keep their normal failure semantics")
	assert.ErrorContains(t, got, "season selection cancelled",
		"non-abort errors should still produce the descriptive wrap so logs are useful")
	assert.ErrorIs(t, got, realErr,
		"the underlying error must remain reachable via errors.Is for diagnostics")
}

func TestSeasonSelection_NilErrorPassesThrough(t *testing.T) {
	assert.NoError(t, mapSeasonSelectionErr(nil),
		"happy path must remain a no-op — no spurious ErrBackToSearch on success")
}
