package superflix

import (
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #202: "abre o chrome test no superflix, e imediatamente fecha".
//
// --sf-offscreen minimizes the solver window and re-asserts that after every
// navigation the solver performs, but the SuperFlix embed raises the window
// itself when it finishes loading — asynchronously, inside the sniff's wait
// loop, after the last navigation has returned. The window then stayed up for
// the rest of the solve and vanished when the context was released.
//
// keepSolverWindowHidden re-asserts inside that loop. These tests pin the
// properties that make it safe: it respects the offscreen switch, never fights
// a deliberate reveal, and throttles so a 500ms loop does not become 500ms of
// CDP traffic.
//
// As in offscreen_test.go, a nil page/context is the probe: reaching the CDP
// call would panic, so a clean return proves the guard short-circuited.

func TestKeepSolverWindowHidden_NoOpWhenNotOffscreen(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "0")
	var page playwright.Page
	t.Cleanup(func() { forgetRevealedPage(page) })

	keepSolverWindowHidden(page, nil)

	_, seen := lastHideAt.Load(page)
	assert.False(t, seen, "a visible-mode run must not record a hide attempt")
}

func TestKeepSolverWindowHidden_SkipsRevealedPages(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	var page playwright.Page
	revealedPages.Store(page, struct{}{})
	t.Cleanup(func() { forgetRevealedPage(page) })

	keepSolverWindowHidden(page, nil)

	_, seen := lastHideAt.Load(page)
	assert.False(t, seen, "a window revealed for manual verification must never be minimized again")
}

func TestKeepSolverWindowHidden_ThrottlesRepeatCalls(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	var page playwright.Page
	t.Cleanup(func() { forgetRevealedPage(page) })

	// Seed a recent attempt; the throttle must swallow the call before it
	// reaches the (nil-context) CDP path.
	seeded := time.Now()
	lastHideAt.Store(page, seeded)

	keepSolverWindowHidden(page, nil)

	at, ok := lastHideAt.Load(page)
	require.True(t, ok, "throttle entry disappeared")
	assert.Equal(t, seeded, at, "the suppressed call must not refresh the throttle")
}

func TestForgetRevealedPage_ClearsThrottleState(t *testing.T) {
	var page playwright.Page
	revealedPages.Store(page, struct{}{})
	lastHideAt.Store(page, time.Now())

	forgetRevealedPage(page)

	_, revealed := revealedPages.Load(page)
	assert.False(t, revealed, "reveal state leaked across plays")
	_, throttled := lastHideAt.Load(page)
	assert.False(t, throttled, "throttle state leaked across plays; the map would grow unbounded")
}

// The solver closes its whole context between plays, so window state for the
// context's own tab (which is never closed individually) has to be dropped
// there or it leaks one entry per play and carries a reveal into the next,
// fresh window.
func TestForgetSolverWindowState_ClearsEveryPage(t *testing.T) {
	var page playwright.Page
	revealedPages.Store(page, struct{}{})
	lastHideAt.Store(page, time.Now())

	forgetSolverWindowState()

	revealedPages.Range(func(k, _ any) bool {
		t.Fatalf("reveal state survived a context close: %v", k)
		return false
	})
	lastHideAt.Range(func(k, _ any) bool {
		t.Fatalf("throttle state survived a context close: %v", k)
		return false
	})
}
