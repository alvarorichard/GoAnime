package superflix

import (
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

// TestHideSolverWindow_NoOpWhenVisible pins that minimizing only engages for a
// hidden-mode solve, and never after a deliberate reveal — re-minimizing a
// window the user was just asked to interact with would take the captcha away
// mid-solve.
func TestHideSolverWindow_NoOpWhenVisible(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "0")
	// nil page/context would panic if the guard let it reach CDP.
	hideSolverWindow(nil, nil)
}

func TestHideSolverWindow_NoOpAfterReveal(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	var page playwright.Page // nil key is fine
	revealedPages.Store(page, struct{}{})
	t.Cleanup(func() { forgetRevealedPage(page) })

	hideSolverWindow(page, nil) // must return before touching the nil context
}

// TestSolverWindowPositionArg_2026_09_01 pins the difference between "the user
// cannot see it" and "the browser is not rendering".
//
// Cloudflare Turnstile rejects headless outright — measured against the live
// gate: a headless run (bundled Chromium and the new headless mode alike) sat
// on the "Verificação" page past 150s, while a headed run cleared it in under
// 5s. What it checks is that a real browser renders, not that a human watches:
// the same solve with the window parked outside the desktop cleared the gate
// and captured the stream in 7.4s from a cold profile.
func TestSolverWindowPositionArg_2026_09_01(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "--window-position=60,60", solverWindowPositionArg(false),
		"default is on screen, so a stalled challenge can still be clicked")
	assert.Equal(t, "--window-position=-32000,-32000", solverWindowPositionArg(true))
}

// TestLoadSuperflixConfig_Offscreen pins the default and its opt-out. Hiding is
// on unless explicitly disabled, and only an explicit falsey value disables it —
// a typo must not silently pop a browser window back into the user's face.
func TestLoadSuperflixConfig_Offscreen(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "")
	assert.True(t, loadSuperflixConfig().Offscreen, "hidden by default")

	for _, off := range []string{"0", "false", "FALSE", "no", "off", " off "} {
		t.Setenv("GOANIME_SF_OFFSCREEN", off)
		assert.Falsef(t, loadSuperflixConfig().Offscreen, "%q must disable it", off)
	}
	for _, on := range []string{"1", "true", "yes", "banana"} {
		t.Setenv("GOANIME_SF_OFFSCREEN", on)
		assert.Truef(t, loadSuperflixConfig().Offscreen, "%q must leave the default alone", on)
	}
}

// Offscreen and headless are independent: --sf-offscreen must NOT imply the
// headless mode the gate rejects.
func TestOffscreenIsNotHeadless(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	t.Setenv("GOANIME_SF_HEADLESS", "")

	cfg := loadSuperflixConfig()
	assert.True(t, cfg.Offscreen)
	assert.False(t, cfg.Headless, "the window must still be a real, rendering browser")
}

// TestChallengeRetrySelectors_2026_09_01 pins the recovery control for a
// failure mode the solver could not see before: the Turnstile widget mounts,
// sits on "Verifying…", and then the page reports "Não foi possível carregar a
// verificação" with a retry button. There is no checkbox to tick, so
// clickTurnstile has nothing to do and the solve burns its whole budget.
//
// The button is hidden until that happens, so it is matched by class rather
// than by its Portuguese label, and a visible bounding box is what says "press
// me".
func TestChallengeRetrySelectors_2026_09_01(t *testing.T) {
	t.Parallel()
	assert.Contains(t, challengeRetrySelectors, "button.captcha-retry")
	assert.Contains(t, challengeRetrySelectors, "button[id^='cfr-']")
	for _, sel := range challengeRetrySelectors {
		assert.NotContains(t, sel, "Tentar novamente",
			"match the control by class, not by a localized label")
	}
}

// TestTurnstileSelectors_ExcludeTheEmptyMount guards a regression this fix
// introduced and then backed out.
//
// SuperFlix renamed its mount to div.cf-turnstile-placeholder, and adding that
// to turnstileSelectors looked like the fix. It is not: the container is empty
// until Cloudflare injects its iframe, so clicking it does nothing while
// consuming clickTurnstile's one allowed click — and a click landing on a
// widget that is still verifying keeps it spinning, so it never reaches the
// failed state whose retry button IS the recovery.
func TestTurnstileSelectors_ExcludeTheEmptyMount(t *testing.T) {
	t.Parallel()
	for _, sel := range turnstileSelectors {
		assert.NotContains(t, sel, "cf-turnstile-placeholder",
			"the mount is not the checkbox; clicking it blocks the real recovery")
		assert.NotContains(t, sel, "cfw-",
			"the per-render mount id is not the checkbox either")
	}
	assert.Contains(t, turnstileSelectors, "iframe[src*='challenges.cloudflare.com']",
		"the real checkbox lives in Cloudflare's own iframe")
}

// TestRevealSolverWindow_NoOpWhenVisible pins that the reveal machinery only
// engages for a hidden window. Without --sf-offscreen the window is already on
// screen, so a reveal would be a pointless CDP round-trip on the play path.
func TestRevealSolverWindow_NoOpWhenVisible(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "0")
	// A nil page/context would panic if the function got past the guard.
	revealSolverWindow(nil, nil, "should not reach CDP")
}

// TestOffscreenRevealAfter pins the patience budget. A warm profile clears the
// gate in under five seconds and a cold one in well under fifteen, so this has
// to sit above a healthy solve and far below the sniff budget — otherwise the
// window either flashes up on every play or never appears in time to help.
func TestOffscreenRevealAfter(t *testing.T) {
	t.Parallel()
	assert.Greater(t, offscreenRevealAfter, 10*time.Second,
		"must outlast a healthy cold solve, or the window pops up for nothing")
	assert.Less(t, offscreenRevealAfter, 45*time.Second,
		"must leave the user time to actually solve it inside the sniff budget")
}

// TestForgetRevealedPage keeps the reveal-once bookkeeping from leaking across
// plays: the solver reuses one browser context for many pages.
func TestForgetRevealedPage(t *testing.T) {
	var page playwright.Page // nil is fine — the map only uses it as a key
	revealedPages.Store(page, struct{}{})
	_, ok := revealedPages.Load(page)
	require.True(t, ok)

	forgetRevealedPage(page)
	_, ok = revealedPages.Load(page)
	assert.False(t, ok, "a closed page must not keep its reveal state forever")
}

// TestFocusSolverPage_SuppressedWhileHidden pins why BringToFront is not called
// directly any more.
//
// page.BringToFront() does not merely select a tab: the window manager activates
// the window, which drags it onto the current desktop. That alone was enough to
// put a hidden solver window at 0,0 right after the warm-up. So focusing is
// suppressed while the window is deliberately hidden, and allowed again once it
// has been revealed on purpose.
func TestFocusSolverPage_SuppressedWhileHidden(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	var page playwright.Page // nil: any call through to Playwright would panic

	focusSolverPage(page) // hidden and not revealed -> must not touch the page
}

func TestFocusSolverPage_AllowedWhenVisible(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "0")
	// With hiding off there is nothing to protect, so the call goes through;
	// a nil page would panic, which is what proves it was not short-circuited.
	assert.Panics(t, func() { focusSolverPage(nil) },
		"with the window visible, focusing must not be suppressed")
}

// TestSolverWindowBounds_RestoresFullRectangle pins the shape of the reveal.
// A minimized window has no usable position to fall back on, so restoring it
// needs the whole rectangle — sending windowState alone leaves it wherever the
// compositor decides.
func TestSolverWindowBounds_RestoresFullRectangle(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"left", "top", "width", "height", "windowState"} {
		assert.Containsf(t, solverWindowBounds, k, "reveal bounds must carry %q", k)
	}
	assert.Equal(t, "normal", solverWindowBounds["windowState"],
		"revealing must un-minimize, not just move")
	assert.Equal(t, solverOnscreenX, solverWindowBounds["left"],
		"a revealed window lands where a normal solve would open")
	assert.Equal(t, solverOnscreenY, solverWindowBounds["top"])
}

// TestRevealIsOncePerPage: the wait loops ask on every tick, but the window may
// only be yanked around once — repeatedly re-raising it would fight the user
// while they are typing into a captcha.
func TestRevealIsOncePerPage(t *testing.T) {
	t.Setenv("GOANIME_SF_OFFSCREEN", "1")
	var page playwright.Page

	// Claim the slot the way a real reveal does, then confirm a second attempt
	// short-circuits before it would need a browser context.
	revealedPages.Store(page, struct{}{})
	t.Cleanup(func() { forgetRevealedPage(page) })

	revealSolverWindow(page, nil, "second attempt") // nil ctx: must not be used

	_, still := revealedPages.Load(page)
	assert.True(t, still, "the reveal record must survive repeated asks")
}
