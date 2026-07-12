package superflix

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/playwright-community/playwright-go"
)

// ErrPlaywrightUnavailable is returned when the Playwright driver or its bundled
// Chromium can't be initialized (first run needs network to download them).

// moveWindow positions the solver's OS window onscreen (Turnstile needs it
// rendered to auto-pass). Best-effort.
func moveWindow(page playwright.Page, x, y int) {
	_, _ = page.Evaluate(fmt.Sprintf(
		`() => { try { window.moveTo(%d, %d); window.resizeTo(1100, 800); } catch (e) {} }`, x, y))
}

// embedHostParentURL is the ungated homepage of the embed host
// (superflixapi.pro) — the same-origin parent the player iframe is injected
// under.
func embedHostParentURL(embedURL string) string {
	parentURL := "https://" + SuperFlixEmbedHost + "/"
	if pu, err := neturl.Parse(embedURL); err == nil && pu.Host != "" {
		parentURL = pu.Scheme + "://" + pu.Host + "/"
	}
	return parentURL
}

// warmGateTopLevel performs a COLD-profile warm-up: a top-level navigation to
// the embed URL so SuperFlix's managed Turnstile auto-passes ONCE at the top
// level and seeds the first-party `__sf_turnstile_pass` cookie. Verified live
// 2026-06-17: a fresh profile does NOT auto-clear when the embed is injected
// straight into an iframe (the same-origin embed blanks before the cookie is
// set), but a top-level visit auto-passes in ~6s — the gate redirects through
// serie|filme/<id>?cfv=<JWT> to the real content. Once the cookie is seeded the
// iframe phase reuses it and clears silently.
//
// Best-effort: returns as soon as the gate clears (cfv redirect, or the
// challenge markup is gone) or the budget elapses; the caller proceeds either
// way. On an already-warm profile it returns almost immediately.
func warmGateTopLevel(page playwright.Page, embedURL string, budget time.Duration) {
	if _, err := page.Goto(embedURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(float64(budget.Milliseconds())),
	}); err != nil {
		return
	}
	_ = page.BringToFront()
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		if hasVerificationParam(page.URL()) {
			return // passing: bouncing through the cfv verification redirect
		}
		if c, err := page.Content(); err == nil && c != "" && !bodyHasChallengeMarker([]byte(c)) {
			return // gate markup gone — cookie seeded
		}
		humanize(page)
		clickTurnstile(page)
		// The warm path normally clears in a few seconds.  Poll often enough to
		// hand control to the iframe as soon as it does, without busy-looping the
		// browser process.
		time.Sleep(350 * time.Millisecond)
	}
}

// injectEmbedSameOrigin loads the embed host homepage and injects the player as a
// SAME-ORIGIN iframe. Same-origin lets the iframe reuse warezcdn's first-party
// `__sf_turnstile_pass` cookie, so a warm profile auto-clears Turnstile in ~7s —
// the fast path. Its cold-profile risk (the embed can navigate the parent away) is
// handled by SniffEmbedStream's blank-out detection + cross-origin fallback.
func injectEmbedSameOrigin(page playwright.Page, embedURL string) error {
	if _, err := page.Goto(embedHostParentURL(embedURL), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(40000),
	}); err != nil {
		return fmt.Errorf("load embed-host parent: %w", err)
	}
	if _, err := page.Evaluate(`(src) => {
		document.body.innerHTML = '<iframe src="' + src + '" allow="autoplay; encrypted-media; fullscreen; picture-in-picture" style="position:fixed;inset:0;width:100%;height:100%;border:0"></iframe>';
	}`, embedURL); err != nil {
		return fmt.Errorf("inject embed iframe: %w", err)
	}
	return nil
}

// injectEmbedCrossOrigin loads the embed in a CROSS-ORIGIN iframe under an opaque
// about:blank parent. Because the iframe's origin no longer matches the top
// document, the embed cannot reach up to navigate the parent — which is exactly
// the cold same-origin failure mode (Cloudflare's managed challenge blanks the
// page to about:blank before it can settle). The trade-off: the pass cookie is
// CHIPS-partitioned and not reused, so this is the recovery path, not the fast one.
func injectEmbedCrossOrigin(page playwright.Page, embedURL string) error {
	if _, err := page.Goto("about:blank"); err != nil {
		return fmt.Errorf("reset to about:blank: %w", err)
	}
	wrapper := fmt.Sprintf(
		`<!doctype html><meta charset="utf-8"><body style="margin:0;background:#000">`+
			`<iframe src=%q allow="autoplay; encrypted-media; fullscreen; picture-in-picture" `+
			`style="position:fixed;inset:0;width:100%%;height:100%%;border:0"></iframe></body>`,
		embedURL,
	)
	if err := page.SetContent(wrapper, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   playwright.Float(20000),
	}); err != nil {
		return fmt.Errorf("set cross-origin iframe: %w", err)
	}
	return nil
}

// pageBlankedOut reports whether the solver page has been navigated to about:blank
// — the cold same-origin failure mode where the embed nukes the parent before the
// challenge settles. (Only meaningful BEFORE the cross-origin fallback, which uses
// an about:blank parent by design.)
func pageBlankedOut(page playwright.Page) bool {
	u := page.URL()
	return u == "" || u == "about:blank"
}

// pageShowsRestrictedShell reports whether the solver page is parked on
// SuperFlix's "Visualização Externa / Acesso Restrito" shell.
//
// This is a THIRD failure mode distinct from the two blank-out cases: the page is
// neither blank nor a live player — it renders a real, static "restricted access"
// card whose only real content is a copy-paste embed iframe. The blank-out
// detectors miss it, so the sniff loop would otherwise spin its full timeout
// staring at it (exactly the >40s hang users hit on some titles).
//
// It scans EVERY frame, not just the top document: the sniff injects the embed as
// a same-origin iframe, which moves the restricted card off the top document and
// into the child frame — checking only the top page missed it entirely (the cause
// of the >40s hang). Cross-origin frames simply fail Content() and are skipped,
// which is fine: by then restrictedSince is already set and the grace timer bails.
func pageShowsRestrictedShell(page playwright.Page) bool {
	for _, fr := range page.Frames() {
		if c, err := fr.Content(); err == nil && c != "" && isRestrictedEmbedPage([]byte(c)) {
			return true
		}
	}
	return false
}

// embedFrameLive reports whether the solver page still has a live (non-blank)
// child frame — the injected player embed.
//
// There are TWO same-origin cold failure modes, not one. pageBlankedOut catches
// the first (the embed navigates the TOP page to about:blank). But the one
// observed against warezcdn is subtler: the TOP page stays on the embed host
// homepage while the embed nukes its OWN iframe to about:blank — so the top-URL
// check misses it and the cross-origin recovery never fires. A healthy injected
// embed always keeps a non-blank child frame; its disappearance is the signal.
func embedFrameLive(page playwright.Page) bool {
	main := page.MainFrame()
	for _, fr := range page.Frames() {
		if fr == main {
			continue
		}
		if u := fr.URL(); u != "" && u != "about:blank" {
			return true
		}
	}
	return false
}

// humanize feeds Cloudflare's managed challenge a behavioral signal it scores: a
// little real pointer movement across the viewport. Combined with the stealthed
// fingerprint and the foregrounded window, this is what lets a managed challenge
// auto-pass with no interaction. Best-effort.
func humanize(page playwright.Page) {
	if m := page.Mouse(); m != nil {
		_ = m.Move(float64(200+secureIntn(500)), float64(180+secureIntn(360)))
	}
}

// secureIntn returns a uniformly random int in [0, n) using crypto/rand. The
// randomness isn't security-critical (it only jitters a mouse position), but
// sourcing it from crypto/rand keeps the whole package free of math/rand. On
// the (practically impossible) read error it returns 0 — a fixed position still
// works as a behavioral signal.
func secureIntn(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// turnstileSelectors locate a rendered Turnstile widget — either the explicit
// mount the page declares (div.cf-turnstile / #cf-turnstile) or the challenge
// iframe Cloudflare injects (challenges.cloudflare.com, or the cf-chl-widget-*
// container). Checked in every frame because the widget can sit in the top
// document or inside the (same-origin) embed.
var turnstileSelectors = []string{
	"iframe[src*='challenges.cloudflare.com']",
	"div.cf-turnstile",
	"#cf-turnstile",
	"div[id^='cf-chl-widget']",
}

// clickTurnstile clicks a rendered Turnstile checkbox with a REAL, OS-level
// (trusted) mouse click — the last automation step needed for a fully
// hands-off solve when the managed challenge degrades to an interactive
// checkbox instead of auto-passing.
//
// Why a coordinate click (not el.Click / JS .click()): the checkbox lives in a
// cross-origin challenges.cloudflare.com iframe we can't reach into, and
// Cloudflare ignores synthetic events (event.isTrusted === false). Only a
// hardware-style page.Mouse().Click at the widget's viewport coordinates
// produces a trusted event Turnstile accepts. BoundingBox is reported relative
// to the main viewport even for elements in child frames, so the coordinates
// line up regardless of iframe nesting; the click then lands on whatever pixel
// is there (the checkbox), through the cross-origin boundary.
//
// The checkbox sits on the LEFT edge of the widget, ~30px in, vertically
// centered. We click at most ONCE per page (guarded by a window flag) because
// re-clicking a widget that's already verifying can reset the challenge.
func clickTurnstile(page playwright.Page) {
	already, err := page.Evaluate(`() => { if (window.__sfTSClicked) return true; window.__sfTSClicked = true; return false; }`)
	if err == nil {
		if done, ok := already.(bool); ok && done {
			return
		}
	}
	for _, fr := range page.Frames() {
		for _, sel := range turnstileSelectors {
			loc := fr.Locator(sel).First()
			if n, cErr := loc.Count(); cErr != nil || n == 0 {
				continue
			}
			box, bErr := loc.BoundingBox()
			if bErr != nil || box == nil || box.Width == 0 {
				continue
			}
			x := box.X + 30
			y := box.Y + box.Height/2
			if m := page.Mouse(); m != nil {
				_ = m.Move(x, y)
				_ = m.Click(x, y)
				util.Debug("SuperFlix CF solve: clicked Turnstile checkbox", "x", x, "y", y)
			}
			return
		}
	}
	// No widget rendered yet — let it appear and retry next tick. Reset the
	// guard so the click can still fire once the widget shows up.
	_, _ = page.Evaluate(`() => { window.__sfTSClicked = false; }`)
}

// triggerPlay nudges every frame's player to start: unmute-and-play any <video>,
// and click the common big-play-button selectors. Best-effort, errors ignored.
func triggerPlay(page playwright.Page) {
	const js = `() => {
	  const fire = (el) => {
	    if (!el) return;
	    try { el.click(); } catch(e){}
	    // Some handlers only react to real pointer/mouse events, not .click().
	    ['pointerdown','mousedown','mouseup','click'].forEach(t => {
	      try { el.dispatchEvent(new MouseEvent(t, {bubbles:true, cancelable:true, view:window})); } catch(e){}
	    });
	  };
	  try {
	    // warezcdn embed source chooser: auto-pick the primary server ONCE so the
	    // player loads without the user clicking "Servidor Principal". Match the
	    // MOST SPECIFIC element (shortest text) — find() would return an ancestor
	    // (card/body) whose click does nothing. Mark it so we don't click again
	    // (re-clicking the player area spawns ad pop-unders).
	    if (!window.__sfServerPicked) {
	      const all = Array.from(document.querySelectorAll('button,a,div,li,span,[role="button"],[data-server],[data-api]'));
	      let cands = all.filter(el => /servidor\s*principal/i.test((el.textContent||'')));
	      if (!cands.length) cands = all.filter(el => /\bservidor\b/i.test((el.textContent||'')));
	      cands.sort((a,b) => (a.textContent||'').length - (b.textContent||'').length);
	      const target = cands[0] || document.querySelector('[data-server],[data-api]');
	      if (target) {
	        window.__sfServerPicked = true;
	        fire(target);
	        const clickable = target.closest('button,a,[role="button"],[data-server],[data-api],li');
	        if (clickable && clickable !== target) fire(clickable);
	      }
	    }
	    // Muted autoplay only — do NOT click play overlays / the player area:
	    // those are ad-click traps that open pop-unders. getVideo fires on player
	    // load anyway, so we don't need a play click.
	    document.querySelectorAll('video').forEach(v => { try { v.muted = true; const p = v.play && v.play(); if (p && p.catch) p.catch(()=>{}); } catch(e){} });
	  } catch(e){}
	}`
	for _, fr := range page.Frames() {
		_, _ = fr.Evaluate(js)
	}
}

// convertPlaywrightCookies converts Playwright's Cookie type into
// net/http.Cookie so the HTTP path can stuff them into its cookie jar without
// depending on Playwright.
func convertPlaywrightCookies(in []playwright.Cookie) []*http.Cookie {
	out := make([]*http.Cookie, 0, len(in))
	for _, c := range in {
		// We forward the browser's own cookies into the HTTP jar, so every
		// attribute (Secure/HttpOnly/SameSite) is mirrored from the source
		// cookie. Forcing Secure/HttpOnly true here — what G124 wants — would
		// corrupt cookies the server set without them and break the clearance.
		out = append(out, &http.Cookie{ // #nosec G124
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			HttpOnly: c.HttpOnly,
			Secure:   c.Secure,
			SameSite: sameSiteFromPlaywright(c.SameSite),
		})
	}
	return out
}

// sameSiteFromPlaywright maps Playwright's SameSite string onto net/http's
// SameSite enum so forwarded cookies keep the attribute the browser actually
// received, rather than silently dropping it.
func sameSiteFromPlaywright(s *playwright.SameSiteAttribute) http.SameSite {
	if s == nil {
		return http.SameSiteDefaultMode
	}
	switch *s {
	case playwright.SameSiteAttribute("Strict"):
		return http.SameSiteStrictMode
	case playwright.SameSiteAttribute("Lax"):
		return http.SameSiteLaxMode
	case playwright.SameSiteAttribute("None"):
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

// ReleaseSharedBrowser closes the shared solver window once a resolve is done, so
// it does not linger on-screen through playback.
//
// It fully closes the browser CONTEXT (window and tabs) rather than just hiding
// it: moving the window off-screen was unreliable — some window managers clamp the
// coordinates back on-screen, leaving the window visible during playback. Closing
// is the only thing that reliably makes it disappear. The Playwright driver (s.pw)
// and the on-disk warm Cloudflare profile are kept, so the next episode relaunches
// a context in ~1s and still solves fast; a re-watch skips the browser entirely
// via the stream cache.
//
// Called at the API layer AFTER the stream URL is obtained (not inside Solve/the
// sniff), so a single resolve — which may drive several solves — keeps one window
// throughout and closes it exactly once at the end. No-op when no window is open
// (e.g. the cache fast path).
func ReleaseSharedBrowser() {
	defaultCFSolver.closeContext()
}

// closeContext closes only the browser context (the visible window), keeping the
// Playwright driver (s.pw) so init() relaunches a fresh context quickly on the
// next solve.
//
// It detaches the context under lifeMu and then runs the actual Close — a
// protocol round-trip + Chromium teardown — in the BACKGROUND, so the caller (a
// post-resolve defer) returns immediately and playback starts without waiting for
// the window to finish disappearing. Nil'ing s.pctx before the async close means
// a following init() sees no context and launches a fresh one, never handing back
// the one being torn down. Safe if already nil.
func (s *cfBrowserSolver) closeContext() {
	s.lifeMu.Lock()
	pctx := s.pctx
	s.pctx = nil
	s.lifeMu.Unlock()
	if pctx != nil {
		go func() { _ = pctx.Close() }()
	}
}

// Close releases the persistent Chromium context and the Playwright driver.
// Uses lifeMu (not the solve mutex) so a SIGINT cleanup can't deadlock against
// an in-flight solve. Safe to call multiple times.
func (s *cfBrowserSolver) Close() {
	s.lifeMu.Lock()
	defer s.lifeMu.Unlock()
	if s.pctx != nil {
		_ = s.pctx.Close()
		s.pctx = nil
	}
	if s.pw != nil {
		_ = s.pw.Stop()
		s.pw = nil
	}
}
