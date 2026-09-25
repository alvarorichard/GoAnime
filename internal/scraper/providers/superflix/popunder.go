package superflix

import (
	"sync/atomic"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/mxschmitt/playwright-go"
)

// Clicking the player is how you start a movie, and how you spring an ad trap.
//
// The sniff used to refuse to click anything but a <video> element, with a
// comment saying play overlays are pop-under traps. That was true and it cost
// the feature: on a movie the muted-autoplay nudge does not start the player at
// all, so the sniff sat through its whole 90s budget capturing nothing while
// the user stared at a window and eventually pressed play themselves — the
// thing the automation exists to avoid.
//
// The trap is worth disarming rather than avoiding. A pop-under needs somewhere
// to go: window.open, a target="_blank" anchor, or a new tab. All three are
// reachable from here, so all three get closed off for the length of the sniff,
// and the click becomes safe to make.

// popunderGuardScript neuters the openers from inside the page.
//
// Runs as an init script, so it is in place before the player's own code — a
// handler installed after the fact would be racing it, and the capture-phase
// listeners below only beat the page's handlers because they are registered
// first. Every patch is wrapped because this must never be the reason a player
// fails to load.
const popunderGuardScript = `(() => {
  try { window.open = function () { return null; }; } catch (e) {}
  try {
    // Kill the click on any anchor, in the capture phase, before the page's own
    // handlers see it.
    //
    // Stripping target="_blank" is NOT enough, and the first version of this
    // that only stripped it was worse than nothing: the anchor then navigated
    // the CURRENT document instead of a new tab, so the ad replaced the player
    // we were sniffing. Caught by the guard's own test, which asserted the page
    // still had the player on it afterwards.
    //
    // Safe to blanket: a video player's own controls are buttons and divs, and
    // an anchor inside a player at the moment we click play is an ad.
    document.addEventListener('click', function (ev) {
      try {
        const a = ev.target && ev.target.closest && ev.target.closest('a[href]');
        if (!a) return;
        ev.preventDefault();
        ev.stopPropagation();
      } catch (e) {}
    }, true);
  } catch (e) {}
  try {
    // Forms have the same escape hatch, and no player submits one.
    document.addEventListener('submit', function (ev) {
      try { ev.preventDefault(); ev.stopPropagation(); } catch (e) {}
    }, true);
  } catch (e) {}
})();`

// popunderGuard closes any page the context opens while it is armed.
//
// The in-page script covers what the page can do to itself; this covers what it
// gets past that — a window the browser opens anyway is closed before it can
// paint. Disarmed, the handler is a no-op, which is how it survives being
// attached once per sniff on a shared context.
type popunderGuard struct {
	armed  atomic.Bool
	closed atomic.Int32
}

// install arms the guard and attaches both halves. The returned func disarms
// it; the handler stays attached to the context and does nothing afterwards.
func (g *popunderGuard) install(page playwright.Page, bctx playwright.BrowserContext) func() {
	g.armed.Store(true)

	if err := page.AddInitScript(playwright.Script{Content: playwright.String(popunderGuardScript)}); err != nil {
		util.Debug("SuperFlix: could not install the pop-under guard script", "err", err)
	}

	bctx.OnPage(func(p playwright.Page) {
		if !g.armed.Load() || p == page {
			return
		}
		g.closed.Add(1)
		util.Debug("SuperFlix: closed a pop-under opened during the sniff", "url", p.URL())
		_ = p.Close()
	})

	return func() { g.armed.Store(false) }
}

// blocked is how many pop-unders were closed, for the debug log and for tests.
func (g *popunderGuard) blocked() int { return int(g.closed.Load()) }
