package superflix

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/playwright-community/playwright-go"
)

// ErrPlaywrightUnavailable is returned when the Playwright driver or its bundled
// Chromium can't be initialized (first run needs network to download them).

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

// sfDirectMediaRe matches actual media traffic (HLS playlists/segments, MP4)
// the player itself fetches once it starts playing. Narrower than sfMediaRe:
// it must NOT match the getVideo/securedLink API URLs, only real media, since
// it feeds SniffEmbedStream's last-resort capture below.
var sfDirectMediaRe = regexp.MustCompile(`(?i)\.m3u8(\?|$|#)|\.mp4(\?|$|#)|/hls/|master\.txt`)

// SniffEmbedStream loads a SuperFlix embed URL (e.g. https://superflixapi.pro/
// filme/1048794 or /serie/76479/1/1) inside a genuine cross-origin iframe so it
// runs in iframe Sec-Fetch context (how the embed is meant to be served), lets
// the persistent profile auto-clear Turnstile, then captures the player's
// `do=getVideo` JSON response and returns its signed HLS master URL.
//
// This is the live extraction path (verified 2026-07-04): the embed funnels to a
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

	// Last-resort capture: the raw media traffic the player emits once it starts
	// playing. If the (rotating) player host changes the getVideo contract —
	// different endpoint, POST body params, new JSON shape — the OnResponse
	// capture below goes blind while the video visibly plays in the solver
	// window. The first HLS/MP4 request IS the playable URL (mpv replays the
	// master.m3u8 directly), so record it and let the wait loop adopt it after
	// giving getVideo a grace period to produce the preferred signed link.
	var fbURL, fbRef, fbUA string
	var fbAt time.Time
	page.OnRequest(func(r playwright.Request) {
		u := r.URL()
		if !sfDirectMediaRe.MatchString(u) {
			return
		}
		mu.Lock()
		if fbURL == "" {
			fbURL = u
			fbAt = time.Now()
			if h, hErr := r.AllHeaders(); hErr == nil {
				fbRef = h["referer"]
				fbUA = h["user-agent"]
			}
		}
		mu.Unlock()
	})

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
		// Adopt the raw media URL only after getVideo has had a grace period
		// to deliver the preferred signed link — the media request fires right
		// after getVideo answers, so if getVideo capture works it always wins.
		if got == "" && fbURL != "" && time.Since(fbAt) > 8*time.Second {
			streamURL = fbURL
			referer = fbRef
			if ua == "" {
				ua = fbUA
			}
			got = streamURL
			util.Debug("SuperFlix getVideo capture missed; adopting raw media URL sniffed from player traffic", "url", fbURL)
		}
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
	if streamURL == "" && fbURL != "" {
		// Timed out waiting for getVideo but the player did fetch media —
		// nothing better is coming, so take what it played.
		streamURL = fbURL
		referer = fbRef
		if ua == "" {
			ua = fbUA
		}
		util.Debug("SuperFlix getVideo capture missed; adopting raw media URL sniffed from player traffic", "url", fbURL)
	}
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
