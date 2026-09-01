package superflix

import (
	"context"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
	"github.com/mxschmitt/playwright-go"
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
	loadBudget := 15 * time.Second
	if remaining := time.Until(deadline); remaining < loadBudget {
		loadBudget = remaining
	}
	if loadBudget <= 0 {
		return "", fmt.Errorf("embedded player frame deadline expired")
	}
	if err := page.SetContent(wrapper, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   new(float64(loadBudget.Milliseconds())),
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

// playerRefererFor builds the Referer the CDN requires for a signed stream URL:
// the player's own /video/<hash> page, NOT the player host's root.
//
// The distinction is not cosmetic — it decides whether anything plays at all.
// Verified live 2026-08-26 against a freshly signed master.txt, two requests:
//
//	Referer: https://<player>/                 -> 403 Forbidden
//	Referer: https://<player>/video/<hash>     -> 200 OK
//
// The browser sends the full path because the player document IS
// /video/<hash> and the request is same-origin; under the default
// strict-origin-when-cross-origin policy only the CROSS-origin segment
// fetches fall back to the bare origin. Sending the bare origin for the
// same-origin playlist fetch is a request no real player ever makes, and the
// CDN rejects it.
//
// With the root Referer the damage landed before mpv ever started:
// streamURLDead probes the signed URL with this exact value and maps 403 to
// "host rotated out", so a perfectly good solve was discarded as a dead host.
//
// hash is empty only on the raw-media fallback capture (no getVideo URL to
// read it from); there the origin is the best available guess.
func playerRefererFor(playerHost, hash string) string {
	playerHost = strings.TrimSuffix(playerHost, "/")
	if playerHost == "" {
		return ""
	}
	if hash == "" {
		return playerHost + "/"
	}
	return playerHost + "/video/" + hash
}

// sfPlayerVideoPageRe matches a player's own document URL,
// https://<player-host>/video/<hash>, and captures the two halves.
var sfPlayerVideoPageRe = regexp.MustCompile(`^(https?://[^/]+)/video/([0-9a-zA-Z]+)`)

// playerIdentityFromReferer recovers the (playerHost, videoHash) pair from the
// Referer a media request carried.
//
// The pair is normally read off the getVideo XHR
// (…/player/index.php?data=<hash>&do=getVideo), but as of 2026-08-31 the
// current player no longer calls getVideo at all — its /video/<hash> document
// fetches master.txt directly. The raw-media fallback still captures that
// request, and its Referer IS the player document, so the same two facts are
// recoverable from it. Without this the pair stayed empty, which cost the
// stream cache (every play re-solved through the browser) and left
// getStreamViaBrowser reporting a blank player host.
func playerIdentityFromReferer(referer string) (playerHost, videoHash string) {
	m := sfPlayerVideoPageRe.FindStringSubmatch(referer)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

// fallbackGraceFor reports how long to keep waiting for a getVideo capture
// before settling for the raw media URL sniffed off the player's own traffic.
//
// getVideo is preferred only because it names the player host and content hash
// that the raw capture used to lack. Once the capture's Referer is a
// /video/<hash> page those are already known (playerIdentityFromReferer), so
// there is nothing left to wait for.
func fallbackGraceFor(fallbackReferer string) time.Duration {
	if host, hash := playerIdentityFromReferer(fallbackReferer); host != "" && hash != "" {
		return 0
	}
	return 8 * time.Second
}

// bloggerFrameURL returns the URL of a Blogger video page loaded in any of the
// page's frames, or "" when none is.
//
// A Blogger-hosted title is invisible to the media sniffer: its player fetches
// the stream through a batchexecute RPC and never issues a request matching
// sfDirectMediaRe, so the sniff would run out its full budget (90s, retried
// once) and fail on a title that plays fine. The player document itself is the
// stream reference — the player layer resolves it — so finding the frame IS
// finding the stream.
func bloggerFrameURL(page playwright.Page) string {
	for _, fr := range page.Frames() {
		if isBloggerPlayerURL(fr.URL()) {
			return fr.URL()
		}
	}
	return ""
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

// preferredGetVideoURL keeps every extraction path on the same upstream
// contract. videoSource is the working unsigned HLS playlist; securedLink is
// retained only as a compatibility fallback for responses that omit it.
func preferredGetVideoURL(response getVideoResponse) string {
	if response.VideoSource != "" {
		return response.VideoSource
	}
	return response.SecuredLink
}

// sfGetVideoRe matches the player's getVideo XHR whose JSON body carries the
// real (signed) HLS URL.
var sfGetVideoRe = regexp.MustCompile(`(?i)/player/index\.php\?.*do=getVideo`)

// sfDirectMediaRe matches actual media traffic (HLS playlists/segments, MP4)
// the player itself fetches once it starts playing. Narrower than sfMediaRe:
// it must NOT match the getVideo/securedLink API URLs, only real media, since
// it feeds SniffEmbedStream's last-resort capture below.
var sfDirectMediaRe = regexp.MustCompile(`(?i)\.m3u8(\?|$|#)|\.mp4(\?|$|#)|/hls/|master\.txt`)

// SniffEmbedStream loads a SuperFlix embed URL (e.g. https://superflixapi.sbs/
// filme/1048794 or /serie/76479/1/1) inside a genuine cross-origin iframe so it
// runs in iframe Sec-Fetch context (how the embed is meant to be served), lets
// the persistent profile auto-clear Turnstile, then captures the player's
// `do=getVideo` JSON response and returns its signed HLS master URL.
//
// This is the live extraction path (verified 2026-07-04): the embed funnels to a
// rotating player host (currently a punycode domain) that answers getVideo with
// {"securedLink":"https://…/cdn/hls/<hash>/master.m3u8?md5=…&expires=…"}, a plain
// multivariant HLS playlist mpv plays directly (no cookie, referer optional).
// restrictedShellGrace is how long the sniff keeps trying after it first sees the
// "Acesso Restrito" shell — enough for the cross-origin embed read to yield a
// stream if one is coming, short enough that a genuinely restricted title fails in
// ~12s instead of spinning the full 90s (×2 solve attempts) the user was hitting.
const restrictedShellGrace = 12 * time.Second

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
	defer func() { _ = page.Close() }()
	// Onscreen during the solve — Turnstile only auto-passes when the page truly
	// renders (headless & offscreen both stall it). Close only this tab afterwards:
	// keeping the persistent context alive retains Chromium's process, connection
	// pools and first-party challenge state for the next episode. Re-launching the
	// whole context here was the largest avoidable delay between plays.
	moveWindow(page, 60, 60)

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
		claimed := fbURL == ""
		if claimed {
			fbURL = u
			fbAt = time.Now()
		}
		mu.Unlock()
		if !claimed {
			return
		}
		// AllHeaders() is a protocol round-trip, and this handler runs on
		// Playwright's dispatch goroutine — the one that reads every message
		// off the driver pipe. Calling it inline blocks the driver waiting on
		// itself: observed hanging the whole sniff until the test timeout
		// (280s) rather than failing. Same rule the OnPopup and OnResponse
		// handlers above already follow.
		go func() {
			h, hErr := r.AllHeaders()
			if hErr != nil {
				return
			}
			mu.Lock()
			fbRef, fbUA = h["referer"], h["user-agent"]
			mu.Unlock()
			select {
			case found <- struct{}{}:
			default:
			}
		}()
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
			if jsonx.Unmarshal(body, &gv) != nil {
				return
			}
			// As of 2026-08-26 the player returns the SAME master.txt URL in
			// both fields, so the choice no longer matters in practice — but it
			// did (securedLink was a dead signed master.m3u8 while videoSource
			// worked), and the fields can diverge again on the next rotation.
			// Keep preferring videoSource: it is what the player itself plays.
			link := preferredGetVideoURL(gv)
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
					videoHash = pu.Query().Get("data")
					referer = playerRefererFor(playerHost, videoHash)
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
	// This is an optimization, never a second timeout.  Keeping it inside the
	// attempt deadline leaves the iframe recovery enough time to work and makes
	// a failed warm-up fail within the advertised sniff budget instead of adding
	// another 45 seconds before that budget even begins.
	deadline := time.Now().Add(timeout)
	warmBudget := min(timeout/3, 20*time.Second)
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
	embedSeen := false // have we ever observed a live embed frame?
	var restrictedSince time.Time
	_ = page.BringToFront() // surface the solve window once so the challenge renders/focuses

	for time.Now().Before(deadline) {
		mu.Lock()
		got := streamURL
		// Adopt the raw media URL once getVideo has had its grace period to
		// deliver the preferred signed link — the media request fires right
		// after getVideo answers, so if getVideo capture works it always wins.
		//
		// The grace collapses to nothing as soon as the captured Referer is a
		// player /video/<hash> page, because then the fallback already carries
		// everything getVideo would have added (player host, content hash, the
		// exact Referer). Waiting the full 8s in that case buys nothing and
		// costs 8s on EVERY play — which is what the current player does, since
		// its getVideo endpoint is gone and the grace could only ever expire.
		if got == "" && fbURL != "" && time.Since(fbAt) > fallbackGraceFor(fbRef) {
			streamURL = fbURL
			referer = fbRef
			playerHost, videoHash = playerIdentityFromReferer(fbRef)
			if ua == "" {
				ua = fbUA
			}
			got = streamURL
			util.Debug("SuperFlix getVideo capture missed; adopting raw media URL sniffed from player traffic",
				"url", fbURL, "host", playerHost, "hash", videoHash)
		}
		mu.Unlock()
		if got != "" {
			break
		}

		// A Blogger-hosted title emits no media request at all, so the capture
		// above can never fire. Its player document is the stream reference.
		if bu := bloggerFrameURL(page); bu != "" {
			mu.Lock()
			streamURL, referer, playerHost, videoHash = bu, "", "", ""
			mu.Unlock()
			util.Debug("SuperFlix sniff: Blogger player detected; using its video page as the stream", "url", bu)
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

		// Third failure mode: the terminal "Acesso Restrito" shell. It is neither
		// blank nor a live player, so the two detectors above miss it and the loop
		// would spin the full timeout. Note when we first see it so we can recover
		// once and then bail fast.
		restricted := pageShowsRestrictedShell(page)
		if restricted && restrictedSince.IsZero() {
			restrictedSince = time.Now()
			util.Debug("SuperFlix sniff: landed on the restricted 'Acesso Restrito' shell")
		}

		if !recovered && (pageBlankedOut(page) || (embedSeen && !alive) || restricted) {
			recovered = true
			util.Debug("SuperFlix CF solve: recovering via cross-origin embed", "restricted", restricted)
			if err := injectEmbedCrossOrigin(page, embedURL); err != nil {
				util.Debug("SuperFlix cross-origin re-inject failed", "err", err)
			}
		}

		// Fast bail on the restricted shell: the cross-origin embed read got a grace
		// window to produce a stream; if none came, none is coming from this
		// content. Return a terminal error so the caller shows a clear message and
		// does NOT burn another 90s solve retrying.
		if !restrictedSince.IsZero() && time.Since(restrictedSince) > restrictedShellGrace {
			return nil, fmt.Errorf("%w (%s)", ErrSuperFlixRestricted, embedURL)
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
		case <-time.After(500 * time.Millisecond):
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if streamURL == "" && fbURL != "" {
		// Timed out waiting for getVideo but the player did fetch media —
		// nothing better is coming, so take what it played.
		streamURL = fbURL
		referer = fbRef
		playerHost, videoHash = playerIdentityFromReferer(fbRef)
		if ua == "" {
			ua = fbUA
		}
		util.Debug("SuperFlix getVideo capture missed; adopting raw media URL sniffed from player traffic",
			"url", fbURL, "host", playerHost, "hash", videoHash)
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
