package superflix

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/mxschmitt/playwright-go"
)

// The browser this package drives is not SuperFlix's private property.
//
// Goyabu began serving a Cloudflare managed challenge (403, cf-mitigated:
// challenge, "Just a moment…") that no header tuning gets past. The same Chrome
// that clears SuperFlix's gate clears that one too, so it is offered through
// netx.ChallengeSolver and any scraper can ask for it — without a provider
// importing a sibling provider.
//
// Solve() is deliberately NOT reused for this. It carries SuperFlix's own
// settle rules (CSRF_TOKEN / ALL_EPISODES, the cfv verification redirect, the
// restricted-shell embed read), none of which mean anything elsewhere, and a
// foreign page would be judged against them. SolveChallenge answers the one
// question a generic caller has: is the interstitial gone, and what cookies
// prove it.

func init() { netx.RegisterChallengeSolver(genericGateSolver{}) }

// genericGateSolver adapts cfBrowserSolver to netx.ChallengeSolver.
type genericGateSolver struct{}

func (genericGateSolver) SolveChallenge(ctx context.Context, targetURL string, timeout time.Duration, reveal netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	return defaultCFSolver.solveGate(ctx, targetURL, timeout, reveal)
}

// solveGate drives the shared browser through an ordinary Cloudflare
// interstitial and returns the clearance.
//
// It reuses everything that already works — the persistent low-fingerprint
// profile, the behavioural nudges, the trusted-click Turnstile handling — and
// stops at "the challenge markup is gone", which is all a plain HTTP scraper
// needs before retrying its own request.
func (s *cfBrowserSolver) solveGate(ctx context.Context, targetURL string, timeout time.Duration, reveal netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	bctx, release, err := s.acquire()
	if err != nil {
		return nil, err
	}
	defer release()

	s.mu.Lock()
	defer s.mu.Unlock()

	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	// Reuse the context's existing tab, exactly as Solve does.
	//
	// Creating a fresh page instead leaves that original tab untouched — and a
	// persistent context always has one, sitting at about:blank in its own
	// window from the moment the browser launched. Nothing in this path ever
	// hid it, so a solve put a second window on screen and left the first one
	// there: a stark white about:blank that stayed behind afterwards. That is
	// the window in the screenshot, and it was never the page being solved.
	var page playwright.Page
	ownPage := false
	// Declared before the teardown defer below, which reads it to decide whether
	// there is a window to take off the screen.
	revealed := false
	if pages := bctx.Pages(); len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = bctx.NewPage()
		if err != nil {
			return nil, fmt.Errorf("create page: %w", err)
		}
		ownPage = true
	}
	defer func() {
		// A window put on screen for this solve has to LEAVE the screen when the
		// solve ends. Closing only our own page does not do that: the persistent
		// context keeps its original tab, which then sits there as a stark white
		// about:blank in a window we raised and never lowered — reported with a
		// screenshot of exactly that.
		//
		// Put it back down while we still hold a page to address the window
		// through. Tearing the whole CONTEXT down here instead looked tidier and
		// was wrong: the next solve relaunches onto the same persistent profile
		// directory while the previous instance still holds its lock, so every
		// following solve failed (measured: eight failed solves, 32s per search).
		forgetRevealedPage(page)
		if ownPage {
			_ = page.Close()
		}
		// Get it off the screen NOW if we ever put it there, so the user sees it
		// go the moment the gate falls rather than when the context is torn
		// down. Hiding is only half the job — a minimized window is still a
		// window — but the other half is the idle watchdog's, below.
		if revealed {
			hideSolverWindow(page, bctx)
		}
		// Teardown is NOT done here any more.
		//
		// This used to call s.closeContext() outright, which was correct about
		// the window needing to disappear and wrong about who gets to decide.
		// A search runs its sources concurrently: closing the shared context at
		// the end of THIS solve tears down the browser another source may be in
		// the middle of using, and that source then relaunches onto the same
		// persistent profile while the old instance still holds its lock.
		//
		// The release handed back by acquire() drops this operation's claim
		// instead. When it was the last one, the watchdog in idle.go closes the
		// context on its own; when it was not, the window stays for whoever is
		// still using it. Either way nobody has to remember anything.
	}()

	// Start hidden, whatever the policy, and earn the right to be seen.
	//
	// This used to take a bare visible=true from the caller and put a window on
	// screen for every solve. It was not wrong about the hard case — a cold
	// profile genuinely cannot clear hidden — but it charged the user for the
	// hard case on every single search, including the overwhelmingly common one
	// where the profile is warm and the gate falls in under two seconds.
	// Measured 2026-09-24, three consecutive solves against goyabu.io: cold and
	// hidden never cleared in 90s; warm and hidden cleared in 1.8s; warm and
	// visible cleared in 1.6s. Hidden costs nothing when it works, and the loop
	// below notices within revealWhenStuckAfter when it does not.
	surface := func(why string) {
		if revealed || reveal == netx.RevealNever {
			return
		}
		revealed = true
		showSolverWindow(page, bctx)
		_ = page.BringToFront()
		// The window is on screen from here until the page paints something of
		// its own. Say what it is meanwhile, instead of showing a blank window
		// that reads as a hung browser.
		brandSolverPage(page)
		util.Info("A verificação do site precisa de um toque seu — abri a janela do navegador. Assim que ela passar, fecho sozinho.")
		util.Debug("challenge solver: revealed the window", "url", targetURL, "reason", why, "policy", reveal.String())
	}

	hideSolverWindow(page, bctx)
	if reveal == netx.RevealAlways {
		surface("caller asked for a visible window")
	}

	if _, err := page.Goto(targetURL, playwright.PageGotoOptions{
		Timeout:   new(float64(timeout.Milliseconds())),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	deadline := time.Now().Add(timeout)
	revealAt := time.Now().Add(revealWhenStuckAfter)
	cleared := false
	for time.Now().Before(deadline) {
		content, cErr := page.Content()
		if cErr == nil && pageLooksReal(content) && !bodyHasChallengeMarker([]byte(content)) {
			cleared = true
			break
		}
		// Same nudges the SuperFlix path uses: a little pointer movement, and a
		// trusted click if the challenge degrades to a checkbox.
		humanize(page)
		clickTurnstile(page)
		// A retry control on screen is the definitive "this is not going to
		// auto-pass" signal, so it surfaces the window immediately rather than
		// waiting out the timer.
		if clickChallengeRetry(page) {
			surface("challenge reported a load failure")
		}
		if time.Now().After(revealAt) {
			surface("gate did not clear on its own")
		}
		if !revealed {
			keepSolverWindowHidden(page, bctx)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	if !cleared {
		return nil, fmt.Errorf("challenge not cleared within %s for %s", timeout, targetURL)
	}

	rawCookies, err := bctx.Cookies(targetURL)
	if err != nil {
		util.Debug("challenge solver: cookie read failed", "url", targetURL, "err", err)
		rawCookies = nil
	}

	// Clearing a Cloudflare gate always leaves cookies. Reporting success with
	// none is reporting a clearance that does not exist, and the caller then
	// replays nothing and blames the replay — which is exactly what happened:
	// "challenge cleared cookies=0" followed by a 403, read for hours as the
	// clearance failing to transfer when there was no clearance to transfer.
	//
	// Failing here instead lets the caller's cooldown do its job and puts the
	// truth in the log.
	if len(rawCookies) == 0 {
		return nil, fmt.Errorf("gate reported clear for %s but produced no cookies", targetURL)
	}

	ua := ""
	if v, uErr := page.Evaluate("() => navigator.userAgent"); uErr == nil {
		if str, ok := v.(string); ok {
			ua = str
		}
	}

	finalURL := page.URL()
	if finalURL == "" {
		finalURL = targetURL
	}

	util.Debug("challenge solver: gate cleared",
		"url", targetURL, "finalURL", finalURL, "cookies", len(rawCookies),
		"policy", reveal.String(), "windowShown", revealed)

	return &netx.ChallengeSolveResult{
		Cookies:   convertPlaywrightCookies(rawCookies),
		UserAgent: ua,
		FinalURL:  finalURL,
	}, nil
}

// revealWhenStuckAfter is how long a hidden solve is given before the window is
// handed to the user.
//
// Deliberately the same budget SuperFlix's own warm-up uses (offscreenRevealAfter):
// a warm profile clears in under two seconds, so anything past ten is a gate
// that is not going to pass on its own, and every further second hidden is a
// second the only person who can solve it is not looking at it.
const revealWhenStuckAfter = 10 * time.Second

// pageLooksReal rejects a document with nothing in it.
//
// The clear check is "no challenge markers", and a blank page has none — so
// about:blank, or a navigation that failed and left the tab empty, read as a
// cleared gate. An interstitial is heavier than this; so is any real page.
func pageLooksReal(html string) bool {
	const minRealPage = 512
	return len(strings.TrimSpace(html)) >= minRealPage
}

// showSolverWindow puts the window on screen at the same spot a revealed one
// uses, and marks the page as revealed so the offscreen machinery does not
// minimize it back down mid-solve.
func showSolverWindow(page playwright.Page, ctx playwright.BrowserContext) {
	revealedPages.Store(page, struct{}{})

	cdp, id, ok := solverWindowID(page, ctx)
	if !ok {
		util.Debug("challenge solver: window not addressable; continuing without placing it")
		return
	}
	if _, err := cdp.Send("Browser.setWindowBounds", map[string]any{
		"windowId": id,
		"bounds":   solverWindowBounds,
	}); err != nil {
		util.Debug("challenge solver: could not place the window on screen", "err", err)
	}
}
