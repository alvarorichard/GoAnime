package superflix

import (
	"context"
	"fmt"
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

func (genericGateSolver) SolveChallenge(ctx context.Context, targetURL string, timeout time.Duration, visible bool) (*netx.ChallengeSolveResult, error) {
	return defaultCFSolver.solveGate(ctx, targetURL, timeout, visible)
}

// solveGate drives the shared browser through an ordinary Cloudflare
// interstitial and returns the clearance.
//
// It reuses everything that already works — the persistent low-fingerprint
// profile, the behavioural nudges, the trusted-click Turnstile handling — and
// stops at "the challenge markup is gone", which is all a plain HTTP scraper
// needs before retrying its own request.
func (s *cfBrowserSolver) solveGate(ctx context.Context, targetURL string, timeout time.Duration, visible bool) (*netx.ChallengeSolveResult, error) {
	bctx, err := s.init()
	if err != nil {
		return nil, err
	}

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
		// Release the browser once the clearance is in hand.
		//
		// Minimizing is not enough: a minimized window is still a window, still
		// in the taskbar, still a live Chrome — reported as "não está fechando o
		// navegador sozinho", with a screenshot of it sitting on the cleared
		// page. Closing the context is what makes it go away, and it is what the
		// SuperFlix resolve path has always done (ReleaseSharedBrowser).
		//
		// Nothing is lost by it. The cookies are already extracted, the caller
		// holds them, and the Cloudflare state lives in the on-disk profile — so
		// the next solve relaunches warm in about a second.
		//
		// Safe here: every solve path takes s.mu, which we still hold, so no
		// other solve can be in flight against the context we are closing.
		s.closeContext()
	}()

	// A caller that needs an on-screen window gets one, whatever --sf-offscreen
	// says. Measured 2026-09-21 against goyabu.io from a challenged IP, profile
	// wiped between runs: minimized never cleared in 70s, visible cleared every
	// time. Hiding it would be honouring a preference at the cost of the result
	// the user actually asked for.
	if visible {
		showSolverWindow(page, bctx)
		_ = page.BringToFront()
		// A new page starts at about:blank, and the window is on screen from
		// here until the navigation paints. Say what it is meanwhile, instead
		// of showing a blank window that reads as a hung browser.
		brandSolverPage(page)
	} else {
		hideSolverWindow(page, bctx)
	}

	if _, err := page.Goto(targetURL, playwright.PageGotoOptions{
		Timeout:   new(float64(timeout.Milliseconds())),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	deadline := time.Now().Add(timeout)
	cleared := false
	for time.Now().Before(deadline) {
		content, cErr := page.Content()
		if cErr == nil && content != "" && !bodyHasChallengeMarker([]byte(content)) {
			cleared = true
			break
		}
		// Same nudges the SuperFlix path uses: a little pointer movement, and a
		// trusted click if the challenge degrades to a checkbox.
		humanize(page)
		clickTurnstile(page)
		if clickChallengeRetry(page) {
			util.Debug("challenge solver: pressed the page's retry control", "url", targetURL)
		}
		if !visible {
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
		"url", targetURL, "finalURL", finalURL, "cookies", len(rawCookies), "visible", visible)

	return &netx.ChallengeSolveResult{
		Cookies:   convertPlaywrightCookies(rawCookies),
		UserAgent: ua,
		FinalURL:  finalURL,
	}, nil
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
