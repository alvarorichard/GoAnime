package superflix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as
// "this episode has no source on SuperFlix" rather than retrying.
var ErrSuperFlixNoServers = errors.New("superflix: no servers available for this content")

// ErrSuperFlixRestricted is returned when the browser sniff lands on SuperFlix's
// terminal "Visualização Externa / Acesso Restrito" shell and no stream follows.
//
// It is a CONTENT signal, not a transient gate flake: the page renders a static
// "restricted access" card whose only content is a copy-paste embed iframe, and it
// does not yield a playable stream. Callers must NOT retry it (retrying just burns
// another 90s solve) and should tell the user to try another title or source.
var ErrSuperFlixRestricted = errors.New("superflix: content is access-restricted (external-embed only)")

// ErrSuperFlixNoEpisodeList is returned when a serie page is solved successfully
// but carries no episode list to parse.
//
// This is a scrape failure, not a content signal: SuperFlix increasingly answers
// /serie/<tmdb> with an embed-only shell ("Embed | <name>", a signed player
// iframe and nothing else), and it does so non-deterministically — the same title
// may render the full frontend on one solve and the bare shell on the next. The
// browser-free TVmaze listing is therefore the reliable path; this error tells
// callers the browser fallback found nothing to work with, so they can say so
// instead of reporting a bare "no seasons found".
var ErrSuperFlixNoEpisodeList = errors.New("superflix: page exposed no episode list")

const (
	// SuperFlixBase is the compiled-in fallback for the SuperFlix host, and the
	// first of many discovery seeds (the rest are retiredSuperFlixHosts in
	// host.go). Retired aliases 301-redirect to whichever one is live; Go's
	// http.Client follows the redirect but downgrades the POST to a GET
	// (dropping the body), which makes /player/bootstrap return HTML and break
	// JSON decoding — so requests must target the live host directly.
	//
	// The live host is discovered at runtime (see host.go), so a stale value
	// here is no longer an outage. `.baby` now 301-redirects to `.monster`
	// (confirmed 2026-09-14, issue #199). When rotating this, move the old host
	// to the top of retiredSuperFlixHosts rather than dropping it.
	SuperFlixBase = "https://superflixapi.monster"
	// SuperFlixEmbedHost is the host that serves the Turnstile-gated player
	// embed. The frontend no longer funnels through warezcdn.lat (which now
	// gates behind Google reCAPTCHA + a QR-scan we can't solve); instead the API
	// host itself serves https://superflixapi.monster/{filme|serie}/<tmdb>,
	// which clears Cloudflare Turnstile (handled by the cfBrowserSolver) and
	// then the player returns the signed HLS master. Like SuperFlixBase this is
	// the fallback and first seed for runtime host discovery.
	SuperFlixEmbedHost = "superflixapi.monster"
	// SuperFlixUserAgent is the UA the plain-HTTP path presents before any
	// Cloudflare solve has run. It must describe the browser the solver drives:
	// Cloudflare binds cf_clearance to the User-Agent that solved the
	// challenge, and a client that then sends a different one is re-challenged
	// in a loop. (After a solve, effectiveUserAgent switches to the real UA the
	// browser reported.)
	//
	// It says Chrome because the solver drives Chromium — Playwright's bundled
	// build, or the system Chrome channel. It used to say
	// "X11; Linux x86_64 … Firefox/125.0", left over from when the solver was a
	// Firefox, and that string is now throttled by this host. Measured
	// 2026-09-24 against /pesquisar, requests spaced 25s apart on one IP with an
	// identical Accept-Language:
	//
	//	Firefox/125.0 (X11; Linux x86_64)   429, 429, 429, 429
	//	Firefox/125.0 (Windows NT 10.0)     200
	//	Firefox/121.0 (Windows NT 10.0)     200, 200
	//	Chrome/124 (X11; Linux x86_64)      200
	//
	// So it is neither "Linux" nor "Firefox 125" on its own — it is that exact
	// string. The user-visible symptom was SuperFlix returning nothing from
	// every search while plain curl on the same URL got 200 and three results.
	//
	// Treat this as measured behaviour, not a permanent law: the Accept-Language
	// note below records a rule on this host that reproduced eight times and
	// then stopped. GOANIME_SF_UA overrides it without waiting for a release.
	SuperFlixUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	// superFlixAcceptLanguage is Chrome's ENGLISH ladder, and the "English" part
	// is not a preference — this host refuses the Portuguese one.
	//
	// Measured 2026-09-25 against /pesquisar, same Chrome User-Agent, requests
	// alternating 22s apart:
	//
	//	pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7   429, 429, 429
	//	en-US,en;q=0.9                        200, 200, 200
	//
	// and, in a separate sweep, every other value tried — no header at all,
	// "pt-BR", "en-US", and Firefox's pt-BR ladder — also answered 200. So it is
	// that one string, not Portuguese in general and not the presence of the
	// header.
	//
	// This is the same rule first seen on 2026-09-21, when it reproduced eight
	// times and then stopped; the note below records that. It came back, and it
	// came back because switching the User-Agent to Chrome brought its ladder
	// along and walked straight into it. The value is also exactly what the
	// media CDN demands byte for byte (cdn.go), so one string now satisfies
	// both.
	//
	// It stays coherent with the UA: en-US,en;q=0.9 is the whole header a real
	// English Chrome sends.
	//
	// The previous value was Chrome's ladder under a Firefox UA. On 2026-09-21
	// this host answered 429 to that exact string while serving every other
	// Accept-Language 200, across eight probes interleaved 15s apart — and then
	// stopped reproducing hours later, with the old value accepted again. So the
	// rule is reputation-sensitive, not a fixed blocklist, and this change is
	// hygiene rather than a guaranteed cure. It is kept because a coherent
	// browser costs nothing, and because this host's CDN already matches
	// Accept-Language by value (see cdn.go).
	superFlixAcceptLanguage = netx.ChromeEnglishAcceptLanguage
)

// Pre-compiled regexes for SuperFlix scraper
var (
	sfCSRFTokenRe   = regexp.MustCompile(`var CSRF_TOKEN\s*=\s*"([^"]+)"`)
	sfPageTokenRe   = regexp.MustCompile(`var PAGE_TOKEN\s*=\s*"([^"]+)"`)
	sfContentIDRe   = regexp.MustCompile(`var INITIAL_CONTENT_ID\s*=\s*(\d+)`)
	sfContentTypeRe = regexp.MustCompile(`var CONTENT_TYPE\s*=\s*"([^"]+)"`)
	sfTitleRe       = regexp.MustCompile(`<title>(?:Player \| )?(.+?)</title>`)
	sfAllEpisodesRe = regexp.MustCompile(`var ALL_EPISODES\s*=\s*(\{.+?\});`)
	// Current rotating frontend injects the full per-season dataset (with
	// air_date, title, epi_num) as `window.allEpisodes = {...};` and renders the
	// anchors client-side from it. The blob carries metadata the anchors don't.
	sfWindowAllEpisodesRe = regexp.MustCompile(`window\.allEpisodes\s*=\s*(\{.+?\});`)
	sfDefaultAudioRe      = regexp.MustCompile(`var defaultAudio\s*=\s*(\[.+?\]);`)
	sfSubtitleRe          = regexp.MustCompile(`var playerjsSubtitle\s*=\s*"(.+?)";`)
	sfSubPartRe           = regexp.MustCompile(`\[(.+?)\](https?://.+)`)
)

// SuperFlixClient handles interactions with SuperFlix
type SuperFlixClient struct {
	client    *http.Client
	baseURL   string
	userAgent string
	// browserSolver drives the headed browser for episode discovery on the
	// rotating, gated frontend. nil in tests (SetTestConfig) so GetEpisodes
	// falls back to the plain HTTP path against an httptest server.
	browserSolver cfSolver
	maxRetries    int
	retryDelay    time.Duration
	searchCache   sync.Map
}

var (
	sharedClientOnce sync.Once
	sharedClient     *SuperFlixClient
)

// SharedSuperFlixClient returns the process-wide client used by the interactive
// API flow. Its HTTP connection pool, cookie jar and solved browser UA survive
// episode selection and subsequent playback, avoiding a fresh TLS connection and
// Cloudflare-cookie handoff for every play. SuperFlixClient contains no
// request-specific mutable state, so net/http's concurrent-safe client/jar can
// safely serve independent metadata and stream requests in parallel.
func SharedSuperFlixClient() *SuperFlixClient {
	sharedClientOnce.Do(func() { sharedClient = NewSuperFlixClient() })
	return sharedClient
}

// NewSuperFlixClient creates a new SuperFlix client.
//
// The HTTP client is wrapped with cfFallbackTransport: on a 403/503/429
// response carrying Cloudflare-challenge markers, the request is replayed
// through a real, headed Firefox (driven via Playwright) to obtain a
// cf_clearance cookie, which is then attached to the retried request and
// every subsequent request for the same host via the cookie jar.
func NewSuperFlixClient() *SuperFlixClient {
	jar, _ := newCookieJar()
	base := netx.SafeScraperTransport(30 * time.Second)
	transport := &cfFallbackTransport{
		base:   base,
		solver: defaultCFSolver,
		jar:    jar,
		// Solve budget. The CF gate may need a manual Turnstile checkbox click
		// in the real Chrome window, so allow ~3min. After the first solve the
		// persistent Chrome profile usually clears it automatically in seconds.
		timeout: 180 * time.Second,
	}
	return &SuperFlixClient{
		client: &http.Client{
			// Wall-clock cap on the ENTIRE Do, including a CF browser solve.
			// Must exceed the solve budget above. Fast (non-challenged)
			// requests still return immediately — this only raises the ceiling.
			Timeout:   210 * time.Second,
			Transport: transport,
			Jar:       jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		baseURL:       SuperFlixBase,
		userAgent:     SuperFlixUserAgent,
		browserSolver: defaultCFSolver,
		maxRetries:    2,
		retryDelay:    200 * time.Millisecond,
	}
}

// NewClientForTest returns a client pointed at a test server, bypassing the
// SSRF-safe transport (localhost is blocked by it) and the browser solver so
// the plain-HTTP paths can be driven against httptest. Only for tests.
func NewClientForTest(serverURL string) *SuperFlixClient {
	c := NewSuperFlixClient()
	c.baseURL = serverURL
	c.client = &http.Client{Timeout: 5 * time.Second, Transport: http.DefaultTransport}
	c.browserSolver = nil
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// base returns the origin every request should target.
//
// A test client (SetTestConfig / NewClientForTest) points at its own httptest
// server and is returned verbatim. A production client seeded with the
// compiled-in default instead resolves through runtime host discovery, so a
// domain rotation is followed automatically rather than 301-ing every POST into
// a body-less GET. Discovery runs at most once per process and falls back to
// the constant, so this stays a cheap atomic load after the first call.
func (c *SuperFlixClient) base() string {
	if c.baseURL != SuperFlixBase {
		return c.baseURL
	}
	if c.browserSolver == defaultCFSolver {
		// Real solver == real site: safe to spend one round trip discovering
		// the live host. Tests swap in a scripted solver (or nil) and never
		// reach the network here.
		ensureLiveHost(context.Background())
	}
	return liveBase()
}

// effectiveUserAgent returns the User-Agent this client's requests actually go
// out with.
//
// It is NOT always c.userAgent: once a Cloudflare solve has run through
// cfFallbackTransport, RoundTrip rewrites every request to carry the solving
// browser's UA so the UA-bound cf_clearance cookie stays valid. Any result that
// reports a User-Agent has to report that one — the player CDN binds a signed
// URL to the UA that fetched it, so handing mpv c.userAgent after the transport
// signed with a different one gets every fetch 403'd.
func (c *SuperFlixClient) effectiveUserAgent() string {
	if t, ok := c.client.Transport.(*cfFallbackTransport); ok {
		if ua := t.getSolvedUA(); ua != "" {
			return ua
		}
	}
	return c.userAgent
}

// userAgentOverrideEnv lets an installed binary route around a UA block without
// waiting for a release — the same escape hatch GOANIME_SF_HOST is for a host
// rotation. This host has blocked a User-Agent string once; it can do it again,
// and the person it happens to should not have to wait for us.
const userAgentOverrideEnv = "GOANIME_SF_UA"

// resolveUserAgent returns the UA to present on plain HTTP requests.
func resolveUserAgent(configured string) string {
	if v := strings.TrimSpace(os.Getenv(userAgentOverrideEnv)); v != "" {
		return v
	}
	return configured
}

func (c *SuperFlixClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", resolveUserAgent(c.userAgent))
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", superFlixAcceptLanguage)
}

// SetTestConfig overrides the base URL and HTTP client for testing.
// This should only be used in test code.
func (c *SuperFlixClient) SetTestConfig(baseURL string, httpClient *http.Client) {
	c.baseURL = baseURL
	c.client = httpClient
	c.browserSolver = nil // force the plain-HTTP episode path against httptest
	c.maxRetries = 0
	c.retryDelay = 0
}

// ensureJSONResponse fails fast when a SuperFlix API endpoint replies with an
// HTML body or a non-2xx status. Without this, callers get the unhelpful
// `invalid character '<' looking for beginning of value` JSON error — which
// hides real causes like the host having moved (the .rest → .online 301 that
// silently downgrades POST → GET) or a Cloudflare/captcha interstitial.
//
// Trust the body, not the Content-Type header. Some upstream players (e.g.
// firevideoplayer.com behind llanfairpwllgwyngy.com) serve real JSON with
// `Content-Type: text/html`, so a header-only check would reject valid
// responses.
func ensureJSONResponse(label string, resp *http.Response, body []byte) error {
	trimmed := strings.TrimLeft(string(body), " \t\r\n\ufeff")
	looksHTML := trimmed != "" && trimmed[0] == '<'

	if looksHTML {
		finalURL := ""
		if resp.Request != nil && resp.Request.URL != nil {
			finalURL = resp.Request.URL.String()
		}
		return fmt.Errorf("%s endpoint returned HTML (status %d, url=%q) — provider may have moved or is blocking the request", label, resp.StatusCode, finalURL)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s endpoint returned status %d", label, resp.StatusCode)
	}
	return nil
}

// Helper: split string by separator and trim each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
