package superflix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/mxschmitt/playwright-go"
)

// ErrPlaywrightUnavailable is returned when the Playwright driver or its bundled
// Chromium can't be initialized (first run needs network to download them).
var ErrPlaywrightUnavailable = errors.New("superflix: playwright chromium unavailable (first run needs network to download it)")

// CFSolveResult holds what the browser captured after solving the CF challenge.
type CFSolveResult struct {
	Cookies   []*http.Cookie
	HTML      string
	FinalURL  string
	UserAgent string // the real browser UA; cf_clearance is bound to it
}

// cfBrowserSolver drives Playwright's OWN bundled Chromium (downloaded on first
// run) to clear Cloudflare Turnstile.
//
// Why Playwright's Chromium (not the user's system browser): the tool must work
// for end users who may not have Chrome installed — Playwright downloads a
// self-contained Chromium into the user cache. We launch it HEADED via a
// PERSISTENT context with the automation fingerprints stripped
// (IgnoreDefaultArgs --enable-automation, --disable-blink-features=
// AutomationControlled, navigator.webdriver masked) so Turnstile treats it as a
// normal browser. Persistent profile means a solved challenge is cached and
// reused across runs; headed means the user can complete a checkbox if one
// appears.
type cfBrowserSolver struct {
	// mu serializes solves (one challenge at a time).
	mu sync.Mutex

	// lifeMu guards the lifecycle handles, independent of mu, so Close() (from
	// the SIGINT cleanup) can tear everything down even while a solve holds mu.
	lifeMu            sync.Mutex
	pw                *playwright.Playwright
	pctx              playwright.BrowserContext
	cleanupRegistered bool
}

var defaultCFSolver = &cfBrowserSolver{}

// webdriverMaskScript patches the common Playwright fingerprint tells (empty
// plugins/mimeTypes, missing window.chrome, a permissions.query that disagrees
// with Notification.permission, the "Google Inc." WebGL UNMASKED_VENDOR, low
// core/memory counts). Every patch is wrapped so a failure can't abort the
// site's own scripts.
//
// IMPORTANT: this is OPT-IN ONLY (GOANIME_SF_MASK) and OFF by default — see the
// launchSolverContext caller. Against SuperFlix's managed Turnstile the masking
// is counter-productive: defining getters on navigator is itself detected and
// makes the challenge refuse to initialize. The bare browser (launch args alone)
// auto-passes; the mask is retained solely as an escape hatch for environments
// where the unmasked fingerprint is the thing that gets rejected.
const webdriverMaskScript = `(() => {
  const def = (obj, prop, val) => { try { Object.defineProperty(obj, prop, { get: () => val, configurable: true }); } catch (_) {} };
  def(navigator, 'webdriver', undefined);
  def(navigator, 'languages', ['pt-BR', 'pt', 'en-US', 'en']);
  // A non-empty plugin list (headless/automation reports zero).
  try {
    const fake = [{ name: 'Chrome PDF Plugin' }, { name: 'Chrome PDF Viewer' }, { name: 'Native Client' }];
    def(navigator, 'plugins', fake);
    def(navigator, 'mimeTypes', [{ type: 'application/pdf' }]);
  } catch (_) {}
  // A plausible consumer-machine core/memory count; automation defaults betray
  // throttled/odd values the managed challenge scores against.
  try { def(navigator, 'hardwareConcurrency', 8); } catch (_) {}
  try { def(navigator, 'deviceMemory', 8); } catch (_) {}
  // window.chrome exists on real Chrome; its absence is a strong automation tell.
  // Provide the app + runtime shapes real Chrome exposes, not a bare {}.
  try { if (!window.chrome) { window.chrome = { runtime: {}, app: { isInstalled: false }, csi: function () {}, loadTimes: function () {} }; } } catch (_) {}
  // permissions.query for 'notifications' must agree with Notification.permission.
  try {
    const orig = window.navigator.permissions && window.navigator.permissions.query;
    if (orig) {
      window.navigator.permissions.query = (p) =>
        p && p.name === 'notifications'
          ? Promise.resolve({ state: Notification.permission })
          : orig(p);
    }
  } catch (_) {}
  // Real GPU vendor/renderer instead of Chromium's SwiftShader / "Google Inc.".
  try {
    const patch = (proto) => {
      const gp = proto.getParameter;
      proto.getParameter = function (p) {
        if (p === 37445) return 'Intel Inc.';                  // UNMASKED_VENDOR_WEBGL
        if (p === 37446) return 'Intel Iris OpenGL Engine';    // UNMASKED_RENDERER_WEBGL
        return gp.apply(this, arguments);
      };
    };
    if (window.WebGLRenderingContext) patch(WebGLRenderingContext.prototype);
    if (window.WebGL2RenderingContext) patch(WebGL2RenderingContext.prototype);
  } catch (_) {}
})();`

// launchSolverContext launches a headed, persistent, low-fingerprint browser
// solverHoldingPage is what the headed solver window shows before it navigates,
// instead of a bare about:blank that looks like a crash. Plain language only — the
// person seeing it did not ask for a browser and should not meet jargon.
const solverHoldingPage = `<!doctype html><html lang="pt-br"><head><meta charset="utf-8">` +
	`<title>GoAnime</title></head>` +
	`<body style="margin:0;height:100vh;display:flex;align-items:center;justify-content:center;` +
	`background:#0f1116;color:#e6e6e6;font-family:system-ui,sans-serif;text-align:center">` +
	`<div><div style="font-size:40px;margin-bottom:16px">🌙 GoAnime</div>` +
	`<div style="font-size:18px">Preparando o vídeo do SuperFlix…</div>` +
	`<div style="font-size:14px;color:#9aa0aa;margin-top:10px">` +
	`Esta janela é normal e vai fechar sozinha. Se aparecer uma caixa “sou humano”, clique nela.</div>` +
	`</div></body></html>`

// brandSolverPage paints the holding page. Best-effort and non-fatal: it is
// cosmetic, so a failure must never stop a solve.
func brandSolverPage(page playwright.Page) {
	_ = page.SetContent(solverHoldingPage, playwright.PageSetContentOptions{
		Timeout: playwright.Float(2000),
	})
}

// context. A non-empty channel selects a system browser distribution (e.g.
// "chrome"); empty uses Playwright's bundled Chromium.
func launchSolverContext(pw *playwright.Playwright, profileDir, channel string, headless bool) (playwright.BrowserContext, error) {
	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(headless),
		// Strip the "Chrome is being controlled by automated test software"
		// switch — its presence is a Turnstile tell.
		IgnoreDefaultArgs: []string{"--enable-automation"},
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--no-first-run",
			"--no-default-browser-check",
			// We close the context abruptly between plays (to make the window
			// disappear during playback), which Chrome flags as a dirty shutdown and
			// greets the next launch with a "Restore pages?" bubble. That bubble
			// overlays the solver page and can block the challenge, so suppress it.
			"--hide-crash-restore-bubble",
			"--disable-session-crashed-bubble",
			// The window MUST be a real, onscreen, non-headless browser for
			// Cloudflare Turnstile to auto-pass: headless is detected and an
			// offscreen/occluded window gets throttled so the challenge stalls
			// (both verified to fail). It's only shown on a COLD solve — the
			// persistent profile pass cookie + on-disk stream cache make repeat
			// plays browser-free.
			"--window-position=60,60",
			"--window-size=1100,800",
		},
	}
	if channel != "" {
		opts.Channel = playwright.String(channel)
	}
	return pw.Chromium.LaunchPersistentContext(profileDir, opts)
}

// init launches a headed, persistent, low-fingerprint browser context.
//
// It PREFERS the user's system Google Chrome (Playwright channel "chrome"): no
// browser download is needed and, being a real consumer browser, it's less likely
// to trip Cloudflare's managed challenge than Playwright's bundled Chromium.
// Skipping the browser download also removes the "BEWARE: your OS is not
// officially supported by Playwright; downloading fallback build …" noise on
// unsupported distros (Fedora, etc.). If Chrome isn't installed the launch fails
// and we fall back to the self-contained bundled Chromium (downloaded on demand)
// so users without Chrome still work. Set GOANIME_SF_BUNDLED=1 to force bundled.
//
// It is REBUILD-CAPABLE: if a prior context died (the user closed the offscreen
// window, a crash, etc.) s.pctx is nil'd by the OnClose hook, and the next call
// launches a fresh one instead of handing back a dead handle. The old sync.Once
// design left a closed context cached forever → every later solve failed with
// "target closed: Target page, context or browser has been closed".
func (s *cfBrowserSolver) init() (playwright.BrowserContext, error) {
	s.lifeMu.Lock()
	defer s.lifeMu.Unlock()

	if s.pctx != nil {
		return s.pctx, nil
	}

	cfg := loadSuperflixConfig()

	// Install the Playwright driver (node). Skip the browser download when we can
	// drive system Chrome. Reuse the driver across context rebuilds.
	if s.pw == nil {
		if err := installPlaywright(!cfg.ForceBundled); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, err)
		}
		pw, err := playwright.Run()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, err)
		}
		s.pw = pw
	}

	cache, _ := os.UserCacheDir()

	channel := cfg.resolveChannel()

	// Try system Chrome first; on failure (not installed), download + use bundled
	// Chromium. Separate profile dirs so a Chrome-created and a Chromium-created
	// profile never clash.
	pctx, err := launchSolverContext(s.pw, solverProfileDir(cache, channel), channel, cfg.Headless)
	if err != nil && channel != "" {
		util.Debug("SuperFlix: system Chrome unavailable, falling back to bundled Chromium", "err", err)
		if !bundledChromiumInstalled() {
			// This download is ~150MB and installPlaywright discards its
			// progress output, so without this line the app looks frozen
			// behind the spinner for minutes.
			util.Info("⏳ Chrome not found — downloading a helper browser (~150 MB, one time only). This can take a few minutes…")
		}
		if instErr := installPlaywright(false); instErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, instErr)
		}
		pctx, err = launchSolverContext(s.pw, solverProfileDir(cache, ""), "", cfg.Headless)
	}
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}
	// Fingerprint mask: OFF by default. Counter-intuitively it BREAKS the solve.
	// Verified live 2026-06-17: with the mask injected, SuperFlix's Turnstile
	// renders its 300x70 widget box but never loads the inner
	// challenges.cloudflare.com challenge iframe (no token, gate never clears) —
	// the navigator getter tampering (Object.defineProperty on webdriver/plugins/
	// permissions) is itself a signal the managed challenge flags, so it refuses
	// to initialize. Without the mask the launch args alone already yield
	// navigator.webdriver === false and the managed challenge AUTO-PASSES in ~7s
	// (top page redirects through serie/<id>?cfv=<JWT> to the real content).
	// Kept behind GOANIME_SF_MASK as an opt-in escape hatch for hosts/future
	// Turnstile builds where the bare fingerprint is rejected instead.
	if cfg.Mask {
		_ = pctx.AddInitScript(playwright.Script{Content: playwright.String(webdriverMaskScript)})
	}

	// Auto-close ad pop-unders. The warezcdn/fireplayer player spawns ad tabs
	// via window.open on interaction (e.g. aboveboardcomplicate.com). Any page
	// with an opener is such a popup — our own pages (NewPage) have no opener —
	// so closing opener!=nil tabs kills the ads without touching legit pages,
	// regardless of the (rotating) ad host. Close off the dispatch goroutine.
	pctx.OnPage(func(p playwright.Page) {
		go func() {
			defer func() { _ = recover() }()
			if op, _ := p.Opener(); op != nil {
				_ = p.Close()
			}
		}()
	})

	// If the context dies (window closed, crash), forget it so the next init()
	// rebuilds. Clear in a goroutine: OnClose fires on Playwright's dispatch
	// goroutine and lifeMu may be held by Close() mid-teardown — grabbing it
	// inline could deadlock.
	pctx.OnClose(func(playwright.BrowserContext) {
		go func() {
			s.lifeMu.Lock()
			if s.pctx == pctx {
				s.pctx = nil
			}
			s.lifeMu.Unlock()
		}()
	})

	s.pctx = pctx

	// Record first successful setup so the one-time "preparing browser" notice
	// (BrowserSetupPending) is not shown on later runs.
	markBrowserReady()

	// Close the browser/driver on program exit (SIGINT/normal). Register once.
	if !s.cleanupRegistered {
		s.cleanupRegistered = true
		util.RegisterCleanup(s.Close)
	}
	return s.pctx, nil
}

// Solve drives the real Chrome through the CF gate for targetURL.
//
// Flow:
//  1. Reuse the persistent-profile browser context (so a cached cf_clearance
//     is already attached) and its page.
//  2. Navigate to targetURL.
//  3. Poll page.Content() until the gate markup is gone. The timeout is
//     generous so the user can click the Turnstile checkbox if it doesn't
//     auto-pass; once solved, the persistent profile usually clears it
//     automatically on later runs.
//  4. Capture cookies, HTML, and the real UA (cf_clearance is UA-bound).
func (s *cfBrowserSolver) Solve(ctx context.Context, targetURL string, timeout time.Duration) (*CFSolveResult, error) {
	bctx, err := s.init()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	// Reuse an existing page (the persistent context's tab) or open one.
	var page playwright.Page
	if pages := bctx.Pages(); len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = bctx.NewPage()
		if err != nil {
			return nil, fmt.Errorf("create page: %w", err)
		}
	}

	// This window is headed and onscreen (Turnstile requires it). Until we
	// navigate, its tab sits at a stark blank "about:blank" that reads as a hung or
	// broken browser — the #184-adjacent report was exactly that confusion. Paint a
	// branded holding page so the user knows GoAnime opened it on purpose and that
	// it will close itself.
	brandSolverPage(page)

	if _, err := page.Goto(targetURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(float64(timeout.Milliseconds())),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	// Wait for the REAL page. Two phases happen after navigation:
	//   1. Turnstile gate ("Verificação" / cf-turnstile-form / Turnstile script)
	//      — cleared automatically or by a manual checkbox click.
	//   2. On success the gate redirects through a transient verification URL
	//      (serie/<id>?cfv=<JWT>) that briefly renders a processing page with
	//      NO gate markers but also NO real content. If we capture there we get
	//      a useless ~20KB page (no CSRF_TOKEN / ALL_EPISODES).
	//
	// So "gate markers gone" is necessary but not sufficient. We keep the latest
	// gate-free HTML as a best-effort fallback, but only declare success once
	// the page has left the verification redirect (URL no longer carries a
	// verification param) AND either shows a real SuperFlix marker or its HTML
	// has stabilized between two reads.
	_ = page.BringToFront() // surface the window so the challenge renders/focuses
	deadline := time.Now().Add(timeout)
	var html, prevContent string
	var pastGate, settled bool
	for time.Now().Before(deadline) {
		content, cErr := page.Content()
		if cErr == nil && content != "" && !bodyHasChallengeMarker([]byte(content)) {
			pastGate = true
			html = content // best-effort: latest gate-free HTML
			u := page.URL()
			realContent := strings.Contains(content, "CSRF_TOKEN") || strings.Contains(content, "ALL_EPISODES")
			// The restricted "Visualização Externa / Acesso Restrito" page is a
			// TERMINAL post-gate state, not a transient verification redirect — its
			// URL keeps ?cfv=<JWT> permanently (scope=embed token). Treating cfv as
			// "still verifying" made the loop spin the full timeout here, leaving the
			// window parked on the restricted page (looks like it's waiting for the
			// user) and exhausting the deadline before the embed read could run. So
			// break out immediately when the real content OR the restricted page is
			// reached; only the genuinely transient cfv processing page (no markers
			// at all) keeps waiting for a stable read.
			restricted := isRestrictedEmbedPage([]byte(content))
			if realContent || restricted || (!hasVerificationParam(u) && content == prevContent) {
				settled = true
				break
			}
			prevContent = content
		} else if cErr == nil {
			// Still gated: feed behavioral signals and auto-click the Turnstile
			// checkbox (trusted OS click) if the challenge demands interaction,
			// so the gate clears with no human.
			humanize(page)
			clickTurnstile(page)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	if !pastGate {
		return nil, fmt.Errorf("CF gate not cleared within %s (auto-solve attempted; if a Turnstile checkbox is still showing in the Chrome window, click it)", timeout)
	}
	if !settled {
		util.Debug("SuperFlix CF solve: gate cleared but page did not settle on real content; using best-effort HTML")
	}

	// Read cookies filtered by the target URL. A bare bctx.Cookies() uses CDP
	// Storage.getCookies, which a CDP-connected (external-browser) context
	// rejects with "Browser context management is not supported"; passing a URL
	// routes through Network.getCookies instead. Non-fatal: we can still return
	// the player HTML even if the cookie snapshot fails.
	rawCookies, err := bctx.Cookies(targetURL, SuperFlixBase)
	if err != nil {
		util.Debug("SuperFlix CF solve: cookie read failed (continuing)", "err", err)
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

	// If we did NOT land on real player/episode content (e.g. the restricted
	// "Visualização Externa" page served once the gate is warm), the content
	// lives in an embed iframe whose src carries a fresh cfv token and
	// scope=embed. The server only serves it in IFRAME Sec-Fetch context, so
	// load that embed URL inside a genuine cross-origin iframe and read the
	// child frame's HTML.
	if !isRealPlayerHTML(html) {
		if embed := extractSuperFlixEmbedURL(html); embed != "" {
			util.Debug("SuperFlix CF solve: reading player from embed iframe", "embed", embed)
			// Fresh deadline: the settle loop above may have consumed most/all of
			// the shared `deadline` (especially on the restricted page), which used
			// to leave readEmbeddedPlayer no time to run at all. ctx still bounds
			// the overall budget.
			// A restricted shell is not a normal slow player load.  Give its
			// cross-origin iframe a short chance to produce the real page, then
			// return control to the caller.  The previous 45s allowance made the
			// server-list enhancement look hung while Chrome stayed visibly parked
			// on "Acesso Restrito".
			embedBudget := 45 * time.Second
			if isRestrictedEmbedPage([]byte(html)) {
				embedBudget = restrictedShellGrace
			}
			embedDeadline := time.Now().Add(embedBudget)
			if pf, fErr := readEmbeddedPlayer(ctx, page, embed, embedDeadline); fErr == nil && pf != "" {
				html = pf
				finalURL = embed
			} else if fErr != nil {
				util.Debug("SuperFlix embed iframe read failed", "err", fErr)
			}
		}
	}

	util.Debug("SuperFlix CF solve result",
		"finalURL", finalURL,
		"htmlLen", len(html),
		"hasALL_EPISODES", strings.Contains(html, "ALL_EPISODES"),
		"hasCSRF_TOKEN", strings.Contains(html, "CSRF_TOKEN"),
		"cookies", len(rawCookies),
		"ua", ua,
	)

	return &CFSolveResult{
		Cookies:   convertPlaywrightCookies(rawCookies),
		HTML:      html,
		FinalURL:  finalURL,
		UserAgent: ua,
	}, nil
}
