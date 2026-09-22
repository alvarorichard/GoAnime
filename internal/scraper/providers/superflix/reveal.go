package superflix

import (
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/mxschmitt/playwright-go"
)

// Smart offscreen: the solver window is hidden while it works and surfaces only
// if the challenge turns out to need a human.
//
// Hiding is done by MINIMIZING the window, not by parking it off the desktop,
// and it is partial rather than absolute. What was measured live 2026-09-01,
// step by step:
//
//	launch at -32000,-32000            window starts there
//	  → first navigation               compositor drags it back to 0,0
//	  → CDP push back to -32000        ignored; lands at 0,0
//	minimize via CDP                   minimized
//	  → navigate to the embed host     STILL minimized
//	  → inject the embed iframe        restored to normal
//
// So minimizing survives ordinary navigation where an offscreen position does
// not, and it still renders enough for Cloudflare (the gate cleared in 3s
// minimized, where headless never clears at all) — but the SuperFlix embed
// itself raises the window when it loads. Neutralizing window.focus/moveTo/
// resizeTo did not stop either behaviour, so it is not the site's own scripts.
//
// The honest summary: --sf-offscreen keeps the window out of the way for the
// launch and warm-up, then it surfaces when the embed loads. Fully invisible
// needs a virtual display (Xvfb), which is not assumed to be installed.
//
// Revealing matters because a permanently hidden window is a trap. Cloudflare's
// widget does sometimes fail to auto-pass — measured on this machine, it can sit
// on "Não foi possível carregar a verificação" indefinitely — and if the user
// cannot see the window, they cannot rescue it and the play just fails.

// solverWindowBounds is where a revealed window lands: the same spot a normal
// (non-offscreen) solve uses, so the experience is identical from there on.
var solverWindowBounds = map[string]any{
	"left": 60, "top": 60, "width": 1100, "height": 800, "windowState": "normal",
}

// revealedPages tracks the pages already pulled on screen, so a reveal happens
// at most once per page no matter how often the wait loop asks.
var revealedPages sync.Map // playwright.Page -> struct{}

// solverWindowID resolves the OS window behind a page.
func solverWindowID(page playwright.Page, ctx playwright.BrowserContext) (playwright.CDPSession, any, bool) {
	cdp, err := ctx.NewCDPSession(page)
	if err != nil {
		util.Debug("SuperFlix: no CDP session for the solver window", "err", err)
		return nil, nil, false
	}
	win, err := cdp.Send("Browser.getWindowForTarget", nil)
	if err != nil {
		util.Debug("SuperFlix: cannot locate the solver window", "err", err)
		return nil, nil, false
	}
	m, ok := win.(map[string]any)
	if !ok || m["windowId"] == nil {
		return nil, nil, false
	}
	return cdp, m["windowId"], true
}

// hideSolverWindow minimizes the solver window so the user never sees it.
//
// No-op unless --sf-offscreen is on, and never undoes a deliberate reveal.
// Best-effort: if minimizing fails the solve still runs, just visibly.
func hideSolverWindow(page playwright.Page, ctx playwright.BrowserContext) {
	if !loadSuperflixConfig().Offscreen {
		return
	}
	if _, revealed := revealedPages.Load(page); revealed {
		return
	}
	cdp, id, ok := solverWindowID(page, ctx)
	if !ok {
		return
	}
	if _, err := cdp.Send("Browser.setWindowBounds", map[string]any{
		"windowId": id,
		"bounds":   map[string]any{"windowState": "minimized"},
	}); err != nil {
		util.Debug("SuperFlix: could not minimize the solver window", "err", err)
		return
	}
	util.Debug("SuperFlix: solver window minimized (--sf-offscreen)")
}

// hideReassertInterval throttles keepSolverWindowHidden. Each re-assert is a
// CDP round-trip and the sniff loop ticks every 500ms, so re-minimizing on
// every tick would triple the loop's protocol traffic for no gain; the window
// is raised by a page load, which is a rare event on that timescale.
const hideReassertInterval = 2 * time.Second

// lastHideAt records the last re-assert per page, so the throttle is per solve
// and cleared with the page (forgetRevealedPage).
var lastHideAt sync.Map // playwright.Page -> time.Time

// keepSolverWindowHidden re-minimizes a window that the page raised behind our
// back. hideSolverWindow is already re-asserted after every navigation we
// perform, but the SuperFlix embed raises the window itself when it finishes
// loading — which happens asynchronously, inside the sniff's wait loop, long
// after the last navigation returned. Without this the window then stayed on
// screen for the rest of the solve, which is the "Chrome opens and closes"
// users report (issue #202).
//
// Carries hideSolverWindow's guards: no-op unless --sf-offscreen is on, and
// never undoes a deliberate reveal, so the manual-verification rescue path is
// unaffected.
func keepSolverWindowHidden(page playwright.Page, ctx playwright.BrowserContext) {
	if !loadSuperflixConfig().Offscreen {
		return
	}
	if _, revealed := revealedPages.Load(page); revealed {
		return
	}
	now := time.Now()
	if prev, ok := lastHideAt.Load(page); ok {
		if at, isTime := prev.(time.Time); isTime && now.Sub(at) < hideReassertInterval {
			return
		}
	}
	lastHideAt.Store(page, now)
	hideSolverWindow(page, ctx)
}

// revealSolverWindow pulls the solver window onto the desktop and focuses it.
//
// No-op unless --sf-offscreen is on (otherwise the window is already visible)
// and no-op after the first call for a given page. Best-effort throughout: a
// failure to reveal must never abort a solve that might still succeed on its
// own.
func revealSolverWindow(page playwright.Page, ctx playwright.BrowserContext, reason string) {
	if !loadSuperflixConfig().Offscreen {
		return
	}
	if _, already := revealedPages.LoadOrStore(page, struct{}{}); already {
		return
	}

	cdp, id, ok := solverWindowID(page, ctx)
	if !ok {
		util.Debug("SuperFlix: solver window not addressable; leaving it hidden")
		return
	}
	// Restoring needs the full rectangle, not just windowState: a minimized
	// window has no usable position of its own to fall back to.
	if _, err := cdp.Send("Browser.setWindowBounds", map[string]any{
		"windowId": id,
		"bounds":   solverWindowBounds,
	}); err != nil {
		util.Debug("SuperFlix: failed to move the solver window on screen", "err", err)
		return
	}
	_ = page.BringToFront()
	util.Debug("SuperFlix: solver window revealed for manual verification", "reason", reason)
	util.Info("SuperFlix: a verificação precisa de você — resolva o captcha na janela que acabou de aparecer.")
}

// focusSolverPage raises the solver window, unless it is deliberately hidden.
//
// page.BringToFront() does not merely select the tab: the window manager
// activates the window, which drags it onto the current desktop. Observed
// directly — a context launched at -32000,-32000 was sitting at 0,0 right after
// the warm-up, purely because of a BringToFront call. So under --sf-offscreen it
// is suppressed until revealSolverWindow has deliberately surfaced the window,
// after which focusing it is exactly what we want.
func focusSolverPage(page playwright.Page) {
	if loadSuperflixConfig().Offscreen {
		if _, revealed := revealedPages.Load(page); !revealed {
			return
		}
	}
	_ = page.BringToFront()
}

// forgetRevealedPage drops a page's reveal state. Called when the solver closes
// a page so the map does not grow across plays.
func forgetRevealedPage(page playwright.Page) {
	revealedPages.Delete(page)
	lastHideAt.Delete(page)
}

// forgetSolverWindowState drops every page's window state at once. The solver
// closes its whole context between plays, which invalidates all of its pages —
// including the persistent context's own tab, which no forgetRevealedPage call
// covers because the solver never closes it individually. Clearing here keeps
// both maps bounded across a long session and, just as importantly, makes each
// new context start hidden again rather than inheriting a reveal from the
// previous play.
func forgetSolverWindowState() {
	revealedPages.Clear()
	lastHideAt.Clear()
}
