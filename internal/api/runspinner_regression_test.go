package api

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Regression tests (added 2026-05-01)
//
// Symptom: after "Going back to anime selection..." the user's debug log
// (goanime_2026-05-01_00-41-46.log lines 161-195) showed all 4 fast sources
// logging "Search results received" (which fires after the per-source results
// are appended to allResults inside searchAllScrapersConcurrent), yet the
// main thread reported "No anime found with the name: naruto" only 2s in —
// far below the 15s searchTimeout, so the for-select loop had not exited.
//
// Root cause: the huh spinner's Run() can return before its Action goroutine
// completes (a tea.Interrupt fires for buffered stdin bytes left over from
// the just-closed fuzzyfinder). The closure inside SearchAnimeEnhanced was
// still running and mutating `animes`/`searchErr`, but runWithSpinner had
// already returned with both at zero values, so `if len(animes) == 0`
// returned "no results found for: naruto".
//
// Pin: awaitActionThroughRunner must (1) run action exactly once and
// (2) block until that single execution has returned, even when the runner
// exits before invoking the wrapped function it was given.

func TestAwaitActionThroughRunner_BlocksWhileRunnerExitsBeforeWrappedFinishes_2026_05_01(t *testing.T) {
	t.Parallel()
	// Simulate the bug: the runner schedules the wrapped action in a
	// goroutine but returns before that goroutine has finished.
	var got atomic.Int64
	runner := func(wrapped func()) {
		go func() {
			time.Sleep(50 * time.Millisecond)
			wrapped()
		}()
	}
	awaitActionThroughRunner(func() {
		got.Store(42)
	}, runner)
	assert.Equal(t, int64(42), got.Load(),
		"awaitActionThroughRunner must wait for the action to finish even when the runner returns first")
}

func TestAwaitActionThroughRunner_RunsActionWhenRunnerNeverInvokesWrapped_2026_05_01(t *testing.T) {
	t.Parallel()
	// If the spinner's Run() returns before its Action command goroutine
	// even starts (early tea.Interrupt), wrapped is never invoked. The
	// trailing safety call must still run the action so the caller does
	// not observe zero values.
	var got atomic.Int64
	noopRunner := func(wrapped func()) { /* never calls wrapped */ }
	awaitActionThroughRunner(func() {
		got.Store(7)
	}, noopRunner)
	assert.Equal(t, int64(7), got.Load(),
		"action must run via the trailing safety call when the runner never invokes wrapped")
}

func TestAwaitActionThroughRunner_RunsActionExactlyOnce_2026_05_01(t *testing.T) {
	t.Parallel()
	// sync.Once must guard against the race where both the runner-launched
	// goroutine and the trailing safety call attempt to invoke action.
	var calls atomic.Int64
	runner := func(wrapped func()) {
		var wg sync.WaitGroup
		for range 5 {
			wg.Go(func() {
				wrapped()
			})
		}
		// Return immediately so the trailing safety call also races.

	}
	awaitActionThroughRunner(func() {
		calls.Add(1)
	}, runner)
	// Allow any straggler goroutines to attempt their wrapped() — once.Do
	// must still cap the count at 1.
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int64(1), calls.Load(),
		"action must execute exactly once across runner-spawned goroutines and the trailing safety call")
}
