package superflix

import (
	"errors"
	"fmt"
	"net/http"
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

const (
	// SuperFlixBase is the canonical SuperFlix host. Previous hosts
	// (`superflixapi.rest`, `superflixapi.online`, `superflixapi.best`,
	// `superflixapi.fit`, `superflixapi.cyou`, `superflixapi.lifestyle`)
	// 301-redirect to whichever alias is live; Go's http.Client follows the
	// redirect but downgrades the POST to a GET (dropping the body), which makes
	// /player/bootstrap return HTML 404 and break JSON decoding — so we target
	// the live host directly. `.lifestyle` redirects to `.pro` (the current
	// canonical host, confirmed 2026-07-04 via the embed's `cfv` session token,
	// which carries `"host":"superflixapi.pro"`).
	SuperFlixBase = "https://superflixapi.pro"
	// SuperFlixEmbedHost is the host that serves the Turnstile-gated player
	// embed. The frontend no longer funnels through warezcdn.lat (which now
	// gates behind Google reCAPTCHA + a QR-scan we can't solve); instead the API
	// host itself serves https://superflixapi.pro/{filme|serie}/<tmdb>, which
	// clears Cloudflare Turnstile (handled by the cfBrowserSolver) and then the
	// player's getVideo endpoint returns the signed HLS master. Confirmed live
	// 2026-07-04 for both /filme and /serie.
	SuperFlixEmbedHost = "superflixapi.pro"
	// SuperFlixUserAgent MUST match the UA the CF solver's Firefox presents
	// (see cfBrowserSolver.Solve). Cloudflare binds the cf_clearance cookie to
	// the User-Agent that solved the challenge; if the HTTP client then sends a
	// different UA, CF rejects the clearance and re-challenges in a loop. A
	// Firefox-on-Linux UA is used because the solver drives a real Firefox.
	SuperFlixUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"
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

func (c *SuperFlixClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
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
	looksHTML := len(trimmed) > 0 && trimmed[0] == '<'

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
