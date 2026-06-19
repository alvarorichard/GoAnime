package scraper

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/playwright-community/playwright-go"
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

// installPlaywright installs the Playwright driver (node) and, unless skipped, the
// bundled Chromium. The installer's stdout/stderr are discarded and Verbose is off
// so its progress and compatibility notices — notably the "BEWARE: your OS is not
// officially supported by Playwright; downloading fallback build …" line on
// unsupported distros — can't bleed into and corrupt the loading spinner. Real
// failures still propagate through the returned error.
func installPlaywright(skipBrowsers bool) error {
	opts := &playwright.RunOptions{
		SkipInstallBrowsers: skipBrowsers,
		Verbose:             false,
		Stdout:              io.Discard,
		Stderr:              io.Discard,
	}
	if !skipBrowsers {
		opts.Browsers = []string{"chromium"}
	}
	return playwright.Install(opts)
}

// solverProfileDir returns the persistent profile directory for the given channel.
// System Chrome and bundled Chromium get separate dirs so a profile created by one
// is never opened by the other (which can refuse or corrupt it). Empty channel ==
// bundled Chromium, keeping the historical "cf-playwright-profile" path.
func solverProfileDir(cache, channel string) string {
	name := "cf-playwright-profile"
	if channel != "" {
		name = "cf-" + sanitizeProfileSegment(channel) + "-profile"
	}
	dir := filepath.Join(cache, "goanime", name)
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// sanitizeProfileSegment reduces an arbitrary string (the browser channel, which
// can come from the GOANIME_SF_CHROME_CHANNEL env var) to a single safe path
// segment. Channel feeds a directory name, so without this a value like
// "../../etc" would let filepath.Join escape the cache dir. We keep only
// [A-Za-z0-9-_] and collapse everything else; an empty result falls back to a
// fixed token so the path is always a valid single component.
func sanitizeProfileSegment(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

// launchSolverContext launches a headed, persistent, low-fingerprint browser
// context. A non-empty channel selects a system browser distribution (e.g.
// "chrome"); empty uses Playwright's bundled Chromium.
func launchSolverContext(pw *playwright.Playwright, profileDir, channel string) (playwright.BrowserContext, error) {
	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(os.Getenv("GOANIME_SF_HEADLESS") != ""),
		// Strip the "Chrome is being controlled by automated test software"
		// switch — its presence is a Turnstile tell.
		IgnoreDefaultArgs: []string{"--enable-automation"},
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--no-first-run",
			"--no-default-browser-check",
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

	forceBundled := os.Getenv("GOANIME_SF_BUNDLED") != ""

	// Install the Playwright driver (node). Skip the browser download when we can
	// drive system Chrome. Reuse the driver across context rebuilds.
	if s.pw == nil {
		if err := installPlaywright(!forceBundled); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, err)
		}
		pw, err := playwright.Run()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, err)
		}
		s.pw = pw
	}

	cache, _ := os.UserCacheDir()

	channel := os.Getenv("GOANIME_SF_CHROME_CHANNEL")
	if channel == "" && !forceBundled {
		channel = "chrome"
	}

	// Try system Chrome first; on failure (not installed), download + use bundled
	// Chromium. Separate profile dirs so a Chrome-created and a Chromium-created
	// profile never clash.
	pctx, err := launchSolverContext(s.pw, solverProfileDir(cache, channel), channel)
	if err != nil && channel != "" {
		util.Debug("SuperFlix: system Chrome unavailable, falling back to bundled Chromium", "err", err)
		if instErr := installPlaywright(false); instErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrPlaywrightUnavailable, instErr)
		}
		pctx, err = launchSolverContext(s.pw, solverProfileDir(cache, ""), "")
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
	if os.Getenv("GOANIME_SF_MASK") != "" {
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
			embedDeadline := time.Now().Add(45 * time.Second)
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

// sfEmbedURLRe extracts the embed iframe src (carrying a ?cfv= token) from a
// real <iframe src="..."> attribute.
var sfEmbedURLRe = regexp.MustCompile(`(?i)src=["']([^"']*\?cfv=[^"']+)["']`)

// sfCfvURLRe finds the embed URL anywhere in the page — including the restricted
// page's "EMBED CODE" box, where the iframe is rendered as escaped TEXT
// (src=&quot;…&quot;) rather than a live attribute, so sfEmbedURLRe misses it.
// The cfv value is a JWT (base64url segments + dots).
var sfCfvURLRe = regexp.MustCompile(`https?://[a-zA-Z0-9.\-]+/(?:serie|filme)/[A-Za-z0-9/_\-]+\?cfv=[A-Za-z0-9._\-]+`)

// extractSuperFlixEmbedURL pulls the player embed URL out of the restricted
// page HTML. It first tries a live iframe attribute, then falls back to a raw
// URL scan after unescaping the HTML entities the EMBED CODE box uses.
func extractSuperFlixEmbedURL(rawHTML string) string {
	s := rawHTML
	for _, r := range []struct{ from, to string }{
		{"&amp;", "&"}, {"&#38;", "&"},
		{"&quot;", `"`}, {"&#34;", `"`}, {"&#039;", "'"}, {"&#39;", "'"},
	} {
		s = strings.ReplaceAll(s, r.from, r.to)
	}
	if m := sfEmbedURLRe.FindStringSubmatch(s); len(m) >= 2 {
		return m[1]
	}
	if u := sfCfvURLRe.FindString(s); u != "" {
		return u
	}
	return ""
}

// readEmbeddedPlayer loads embedURL inside a genuine cross-origin iframe and
// returns the child frame's HTML once it contains the real player markers.
//
// We first navigate to about:blank so the wrapper page has an opaque origin —
// that makes the iframe load cross-site (Sec-Fetch-Site: cross-site), matching
// how SuperFlix is meant to be embedded. The server then serves the player
// (CSRF_TOKEN / ALL_EPISODES) instead of the restricted page.
func readEmbeddedPlayer(ctx context.Context, page playwright.Page, embedURL string, deadline time.Time) (string, error) {
	if _, err := page.Goto("about:blank"); err != nil {
		return "", fmt.Errorf("reset to about:blank: %w", err)
	}
	wrapper := fmt.Sprintf(
		`<!doctype html><meta charset="utf-8"><body style="margin:0">`+
			`<iframe src=%q style="width:1280px;height:720px;border:0" `+
			`allow="autoplay; encrypted-media; fullscreen"></iframe></body>`,
		embedURL,
	)
	if err := page.SetContent(wrapper, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   playwright.Float(15000),
	}); err != nil {
		return "", fmt.Errorf("set iframe wrapper: %w", err)
	}

	for time.Now().Before(deadline) {
		for _, fr := range page.Frames() {
			c, cErr := fr.Content()
			if cErr == nil && isRealPlayerHTML(c) {
				return c, nil
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return "", fmt.Errorf("embedded player frame did not yield content before deadline")
}

// isRealPlayerHTML reports whether html is actual SuperFlix content (legacy
// player ALL_EPISODES, or the rotating frontend's episode anchors) rather than
// a gate/restricted-embed shell.
func isRealPlayerHTML(html string) bool {
	return strings.Contains(html, "ALL_EPISODES") || strings.Contains(html, "data-episode-id")
}

// hasVerificationParam reports whether a URL still carries a Cloudflare/SuperFlix
// verification-redirect token. SuperFlix's gate, on success, bounces through
// serie/<id>?cfv=<JWT> before landing on the real page; while that param is
// present we have NOT reached real content yet.
func hasVerificationParam(rawURL string) bool {
	for _, p := range []string{"cfv=", "__cf_chl", "cf_chl_", "cf_chl_rt"} {
		if strings.Contains(rawURL, p) {
			return true
		}
	}
	return false
}

// CFStreamResult holds a sniffed media URL and the headers mpv needs to play it.
//
// PlayerHost and VideoHash identify the warezcdn player content. They are the
// only browser-gated facts in the whole chain — once captured they can be cached
// and replayed forever via a pure-HTTP getVideo call (no browser), since the
// player host's getVideo endpoint is NOT Cloudflare-gated. See
// SuperFlixClient.getStreamViaBrowser / the on-disk stream cache.
type CFStreamResult struct {
	StreamURL  string
	Referer    string
	UserAgent  string
	PlayerHost string // e.g. https://xn--kcksk7a2bl5le7b6doc1h3f.com
	VideoHash  string // 32-hex warezcdn content id
}

// sfMediaRe matches the network requests that carry the actual video (HLS
// playlist, MP4, or the players' getVideo/securedLink endpoints).
var sfMediaRe = regexp.MustCompile(`(?i)\.m3u8(\?|$|#)|\.mp4(\?|$|#)|/getVideo|videoSource|securedLink|/hls/|master\.txt`)

// SniffStream drives the real browser through a player embed and captures the
// first media URL the player fetches, plus the Referer/UA needed to replay it
// in mpv.
//
// SuperFlix's player providers are all Cloudflare-gated and serve content only
// in iframe context, so we load embedURL inside a genuine cross-origin iframe
// (the persistent profile auto-clears Turnstile), then nudge the nested player
// to start (muted autoplay + clicking common play controls) while a network
// listener watches every frame for a media request.
//
// Foundation: returns the first matching media URL. Some providers need extra
// play interaction or de-obfuscation that can be layered on later.
func (s *cfBrowserSolver) SniffStream(ctx context.Context, embedURL string, timeout time.Duration) (*CFStreamResult, error) {
	bctx, err := s.init()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	page, err := bctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	defer func() { _ = page.Close() }()

	var mu sync.Mutex
	var hitURL, hitRef, hitUA string
	found := make(chan struct{}, 1)
	page.OnRequest(func(r playwright.Request) {
		u := r.URL()
		if !sfMediaRe.MatchString(u) {
			return
		}
		mu.Lock()
		if hitURL == "" {
			hitURL = u
			if h, herr := r.AllHeaders(); herr == nil {
				hitRef = h["referer"]
				hitUA = h["user-agent"]
			}
			select {
			case found <- struct{}{}:
			default:
			}
		}
		mu.Unlock()
	})

	// Load the embed in a cross-origin iframe so it runs in iframe Sec-Fetch
	// context (how the player is meant to be served).
	if _, err := page.Goto("about:blank"); err != nil {
		return nil, fmt.Errorf("reset page: %w", err)
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
		return nil, fmt.Errorf("load embed iframe: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := hitURL
		mu.Unlock()
		if got != "" {
			break
		}
		triggerPlay(page)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-found:
		case <-time.After(2 * time.Second):
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if hitURL == "" {
		return nil, fmt.Errorf("no media URL sniffed within %s", timeout)
	}
	if hitUA == "" {
		hitUA = SuperFlixUserAgent
	}
	util.Debug("SuperFlix sniffed stream", "url", hitURL, "referer", hitRef)
	return &CFStreamResult{StreamURL: hitURL, Referer: hitRef, UserAgent: hitUA}, nil
}

// getVideoResponse models the JSON the warezcdn player's
// /player/index.php?do=getVideo endpoint returns. securedLink is the signed HLS
// master (md5+expires); videoSource is the unsigned master.txt fallback.
type getVideoResponse struct {
	HLS         bool   `json:"hls"`
	SecuredLink string `json:"securedLink"`
	VideoSource string `json:"videoSource"`
	VideoImage  string `json:"videoImage"`
}

// sfGetVideoRe matches the player's getVideo XHR whose JSON body carries the
// real (signed) HLS URL.
var sfGetVideoRe = regexp.MustCompile(`(?i)/player/index\.php\?.*do=getVideo`)

// SniffEmbedStream loads a warezcdn embed URL (e.g. https://warezcdn.lat/filme/
// 1048794 or /serie/76479/1/1) inside a genuine cross-origin iframe so it runs
// in iframe Sec-Fetch context (how the embed is meant to be served), lets the
// persistent profile auto-clear Turnstile, then captures the player's
// `do=getVideo` JSON response and returns its signed HLS master URL.
//
// This is the live extraction path (verified 2026-06-09): the embed funnels to a
// rotating player host (currently a punycode domain) that answers getVideo with
// {"securedLink":"https://…/cdn/hls/<hash>/master.m3u8?md5=…&expires=…"}, a plain
// multivariant HLS playlist mpv plays directly (no cookie, referer optional).
func (s *cfBrowserSolver) SniffEmbedStream(ctx context.Context, embedURL string, timeout time.Duration) (*CFStreamResult, error) {
	bctx, err := s.init()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	page, err := bctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	// Onscreen during the solve — Turnstile only auto-passes when the page truly
	// renders (headless & offscreen both stall it). On exit, tear the WHOLE
	// context down so no window lingers (the persistent context's default
	// about:blank tab would otherwise stay visible after we close our page).
	// init() rebuilds the context on the next solve; the on-disk profile keeps
	// the warm pass cookie, and pw is reused, so the rebuild is fast.
	moveWindow(page, 60, 60)
	defer s.closeContext()

	// Close ad popunders the embed spawns via window.open so they don't steal the
	// context. Close in a goroutine — Close() is a protocol round-trip and must
	// not run on Playwright's dispatch goroutine.
	page.OnPopup(func(p playwright.Page) {
		go func() { _ = p.Close() }()
	})

	var mu sync.Mutex
	var streamURL, referer, ua, playerHost, videoHash string
	found := make(chan struct{}, 1)

	page.OnResponse(func(resp playwright.Response) {
		u := resp.URL()
		if !sfGetVideoRe.MatchString(u) {
			return
		}
		// Body() is a round-trip — read it off the dispatch goroutine.
		go func() {
			body, bErr := resp.Body()
			if bErr != nil {
				return
			}
			var gv getVideoResponse
			if json.Unmarshal(body, &gv) != nil {
				return
			}
			link := gv.SecuredLink
			if link == "" {
				link = gv.VideoSource
			}
			if link == "" {
				return
			}
			mu.Lock()
			if streamURL == "" {
				streamURL = link
				// Player origin + content hash from the getVideo URL
				// (…/player/index.php?data=<hash>&do=getVideo). These two are the
				// only browser-gated facts — cache them for browser-free replays.
				if pu, pErr := neturl.Parse(u); pErr == nil {
					playerHost = pu.Scheme + "://" + pu.Host
					referer = playerHost + "/"
					videoHash = pu.Query().Get("data")
				}
				select {
				case found <- struct{}{}:
				default:
				}
			}
			mu.Unlock()
		}()
	})

	// Phase 0 — COLD warm-up. A fresh profile does NOT auto-clear when the embed
	// is injected straight into an iframe (verified live 2026-06-17: the embed
	// blanks before the first-party pass cookie is ever set). A TOP-LEVEL visit
	// to the gate, however, auto-passes in ~6s and seeds `__sf_turnstile_pass`.
	// So warm the cookie at the top level first; the same-origin iframe phase
	// below then reuses it and clears silently. Cheap on a warm profile (the
	// gate is already cleared, so it returns almost immediately).
	warmBudget := timeout / 2
	if warmBudget > 45*time.Second {
		warmBudget = 45 * time.Second
	}
	warmGateTopLevel(page, embedURL, warmBudget)

	// Phase 1 — SAME-ORIGIN (fast path). Navigate the parent to warezcdn's own
	// (ungated) homepage and inject the player as a same-origin iframe so it reuses
	// the stable FIRST-PARTY `__sf_turnstile_pass` cookie. A warm profile auto-
	// clears Turnstile here silently in ~7s.
	if err := injectEmbedSameOrigin(page, embedURL); err != nil {
		return nil, err
	}
	if v, uErr := page.Evaluate("() => navigator.userAgent"); uErr == nil {
		if str, ok := v.(string); ok {
			ua = str
		}
	}

	// Watch for the COLD failure mode: when the managed challenge doesn't auto-pass,
	// the same-origin embed navigates the PARENT to about:blank within ~7s and the
	// solve can never recover (verified: every frame collapses to about:blank and
	// stays there). Detect that blank-out and re-inject CROSS-ORIGIN under an opaque
	// about:blank parent the embed cannot navigate — that keeps the page alive so
	// the managed challenge has time to settle in place. Done at most once.
	//
	// Throughout, feed the challenge the behavioral signals it scores (window in the
	// foreground + a little pointer movement) which, with the stealthed fingerprint,
	// is what tips a managed challenge into auto-passing without interaction.
	recovered := false
	embedSeen := false      // have we ever observed a live embed frame?
	_ = page.BringToFront() // surface the solve window once so the challenge renders/focuses

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := streamURL
		mu.Unlock()
		if got != "" {
			break
		}

		// Detect BOTH same-origin failure modes before recovering once:
		//   (1) the embed navigated the TOP page to about:blank (pageBlankedOut), and
		//   (2) the embed nuked its OWN iframe while the top stayed on the host —
		//       caught by embedSeen && !embedFrameLive (the live child frame vanished).
		// embedSeen gates (2) so we don't false-trigger before the iframe loads.
		alive := embedFrameLive(page)
		if alive {
			embedSeen = true
		}
		if !recovered && (pageBlankedOut(page) || (embedSeen && !alive)) {
			recovered = true
			util.Debug("SuperFlix CF solve: same-origin embed blanked the page; retrying cross-origin")
			if err := injectEmbedCrossOrigin(page, embedURL); err != nil {
				util.Debug("SuperFlix cross-origin re-inject failed", "err", err)
			}
		}

		// Feed behavioral signals every tick while we're still gated. warezcdn's
		// managed challenge ("Validação segura") is served directly (no
		// challenges.cloudflare.com iframe), so gating this on challengeFramePresent
		// would never fire — drive it whenever no stream has been captured yet.
		// clickTurnstile handles the degraded case where the challenge stops
		// auto-passing and demands an interactive checkbox: a trusted OS click
		// completes it with no human.
		humanize(page)
		clickTurnstile(page)
		triggerPlay(page)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-found:
		case <-time.After(2 * time.Second):
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if streamURL == "" {
		return nil, fmt.Errorf("no getVideo stream captured within %s", timeout)
	}
	if ua == "" {
		ua = SuperFlixUserAgent
	}
	util.Debug("SuperFlix sniffed embed stream", "url", streamURL, "referer", referer, "host", playerHost, "hash", videoHash)
	return &CFStreamResult{
		StreamURL:  streamURL,
		Referer:    referer,
		UserAgent:  ua,
		PlayerHost: playerHost,
		VideoHash:  videoHash,
	}, nil
}

// moveWindow positions the solver's OS window onscreen (Turnstile needs it
// rendered to auto-pass). Best-effort.
func moveWindow(page playwright.Page, x, y int) {
	_, _ = page.Evaluate(fmt.Sprintf(
		`() => { try { window.moveTo(%d, %d); window.resizeTo(1100, 800); } catch (e) {} }`, x, y))
}

// embedHostParentURL is the ungated homepage of the embed host (warezcdn.lat) —
// the same-origin parent the player iframe is injected under.
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
		time.Sleep(800 * time.Millisecond)
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

// closeContext closes only the browser context (the visible window) after a
// solve, keeping the Playwright driver (s.pw) so init() can relaunch a fresh
// context quickly on the next solve. This is why no window lingers after a
// stream resolves. Uses lifeMu (not the solve mutex). Safe if already nil.
func (s *cfBrowserSolver) closeContext() {
	s.lifeMu.Lock()
	defer s.lifeMu.Unlock()
	if s.pctx != nil {
		_ = s.pctx.Close()
		s.pctx = nil
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
