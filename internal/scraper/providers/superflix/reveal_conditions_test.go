package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The solver window is hidden by default (--sf-offscreen) and is handed to the
// user only when a human can actually help. Two signals justify that:
//
//   - the challenge page reported a load failure and rendered its retry button
//     (clickChallengeRetry), or
//   - a Turnstile widget is on screen and has not cleared for
//     offscreenRevealAfter.
//
// Elapsed time on its own does not. It used to: the sniff loop revealed the
// window 15s in, whatever the page was doing. A cold solve that had already
// passed the gate and was waiting for the player to emit its media request —
// measured at 22s on this machine — therefore popped a window saying "resolva o
// captcha", with no captcha on it. That is what "o captcha não está resolvendo
// sozinho" describes: the solve was fine, the prompt was wrong.

// TestRevealSelectors_CoverTheWidgetFormsWeCanSee pins the selector list that
// challengeVisible shares with clickTurnstile. A widget form missing from it is
// a widget we would neither click nor reveal for.
func TestRevealSelectors_CoverTheWidgetFormsWeCanSee(t *testing.T) {
	t.Parallel()
	assert.Contains(t, turnstileSelectors, "iframe[src*='challenges.cloudflare.com']",
		"the challenge iframe is the form Cloudflare always injects; without it nothing is detectable")
	assert.Contains(t, turnstileSelectors, "div.cf-turnstile")
	assert.Contains(t, turnstileSelectors, "#cf-turnstile")
	assert.Contains(t, turnstileSelectors, "div[id^='cf-chl-widget']")
}

// The retry button is the one unambiguous "this will not auto-pass" signal, so
// it reveals immediately rather than waiting out the timer.
func TestRevealSelectors_RetryButtonIsDistinctFromTheWidget(t *testing.T) {
	t.Parallel()
	for _, retry := range challengeRetrySelectors {
		assert.NotContains(t, turnstileSelectors, retry,
			"the retry control is a different signal from a rendered widget and must stay separate")
	}
}

// A cold solve on this machine takes ~13s and has been measured at 22s. The
// timer has to leave room for the slow-but-healthy case, since crossing it no
// longer reveals on its own — but it must still be short enough that a genuinely
// stuck challenge reaches the user quickly.
func TestOffscreenRevealAfter_LeavesRoomForASlowButHealthySolve(t *testing.T) {
	t.Parallel()
	assert.GreaterOrEqual(t, offscreenRevealAfter.Seconds(), 10.0,
		"a cold profile routinely needs more than ten seconds; revealing sooner is a false alarm")
	assert.LessOrEqual(t, offscreenRevealAfter.Seconds(), 30.0,
		"a stuck challenge must reach the user while they are still watching")
}

// challengeVisible must be inert on a page it cannot query, because it runs on
// the reveal path: an error there has to mean "nothing to show the user", never
// a panic in the middle of a solve.
func TestChallengeVisible_IsInertOnAnUnusablePage(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		assert.False(t, challengeVisible(nil))
	})
}

// A challenge GoAnime cannot click is still a challenge.
//
// turnstileSelectors lists what clickTurnstile can act on, and deliberately
// excludes SuperFlix's own mount. Gating the hand-over on that list alone left
// the user staring at SuperFlix's "Verificação" page with a minimized window
// and no prompt — reported as "o captcha não resolve sozinho", which was
// exactly right: it could not, and nobody was told.
//
// The page's own gate markers therefore count as well, and they are the same
// set the HTTP path recognises, so both layers agree on what a gate is.
func TestChallengeMarkers_CoverTheGateSuperFlixServesItself(t *testing.T) {
	t.Parallel()
	for _, marker := range []string{
		"cf-turnstile-form",
		"<title>Verificação</title>",
	} {
		assert.Truef(t, bodyHasChallengeMarker([]byte("<html>"+marker+"</html>")),
			"%q is SuperFlix's own gate; the reveal path has to recognise it", marker)
	}

	// And the mount that is deliberately NOT clickable must still not be a
	// clickable selector, or clickTurnstile spends its one click on an empty box.
	for _, sel := range turnstileSelectors {
		assert.NotContains(t, sel, "cf-turnstile-placeholder",
			"the placeholder is a mount, not a checkbox; clicking it burns the single allowed click")
	}
}

// The real player page must never look like a challenge, or every healthy solve
// would pop a window at the 15s mark again.
func TestChallengeMarkers_RealPlayerPageIsNotAChallenge(t *testing.T) {
	t.Parallel()
	const player = `<html><script>var CSRF_TOKEN = "abc"; var ALL_EPISODES = {};</script></html>`

	assert.False(t, bodyHasChallengeMarker([]byte(player)),
		"a page with the player markers is past the gate; revealing there is the false alarm")
	assert.True(t, isRealPlayerHTML(player))
}
