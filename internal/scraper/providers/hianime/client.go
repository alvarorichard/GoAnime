package hianime

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

const (
	// hianimeBase is the public host. Pinned here so a rotation fails loudly in
	// TestHostIsPinned instead of silently scraping the wrong site.
	hianimeBase = "https://hianime.at"

	// restPath is the prefix the site's own frontend uses for the two endpoints
	// this scraper calls. It is a Laravel app, so a wrong path answers 404 with
	// a JSON {"message":"The route … could not be found."} rather than an HTML
	// error page — which is how the old aniwatch-style /ajax/v2/… guesses were
	// ruled out on 2026-09-22.
	restPath = "/api/theme/"

	// sourceLabel is the name carried in diagnostics and in models.Anime.Source.
	sourceLabel = "HiAnime"

	// Response caps for jsonx.Decode. The episode list of a long-running show
	// arrives as an HTML fragment inside the JSON envelope, so it is larger than
	// a plain API payload: Naruto's 220 episodes weigh ~210KB. 4MB leaves room
	// for the longest shows on the site without accepting a payload we would
	// never want.
	maxJSONResponseBytes = 4 << 20
	maxHTMLResponseBytes = 8 << 20
)

// Package-level regexes: compiled once, per the provider contract.
var (
	// animeHrefRe matches an anime permalink and captures slug and numeric id.
	// Host-agnostic on purpose so a domain rotation only needs the const above.
	// The optional /watch prefix is the same anime under its player URL.
	animeHrefRe = regexp.MustCompile(`(?:^|/)(?:watch/)?([a-z0-9][a-z0-9-]*?)-(\d+)/?(?:\?|$)`)
	// episodeIDRe captures the episode id out of a watch URL's ?ep= parameter.
	episodeIDRe = regexp.MustCompile(`[?&]ep=(\d+)`)
	// variantRe reads one #EXT-X-STREAM-INF line's resolution height.
	variantRe = regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	// qualityDigitsRe pulls "1080" out of "1080p".
	qualityDigitsRe = regexp.MustCompile(`(\d{3,4})`)
)

// HiAnimeClient handles interactions with hianime.at.
type HiAnimeClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewHiAnimeClient creates a new hianime.at client. Performs no network I/O: it
// runs under sync.Once in the adapter.
func NewHiAnimeClient() *HiAnimeClient {
	return &HiAnimeClient{
		client:     util.NewFastClient(),
		baseURL:    hianimeBase,
		userAgent:  netx.UserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

// NewClientForTest returns a client pointed at a test server with retries
// disabled. Only for tests.
func NewClientForTest(serverURL string) *HiAnimeClient {
	c := NewHiAnimeClient()
	c.baseURL = strings.TrimSuffix(serverURL, "/")
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// restURL builds one of the frontend endpoints under restPath.
func (c *HiAnimeClient) restURL(suffix string) string {
	return c.baseURL + restPath + suffix
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func (c *HiAnimeClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", netx.EnglishAcceptLanguage)
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *HiAnimeClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

// sleep waits out the retry delay, or returns early when ctx is cancelled.
// A plain time.Sleep here would keep a cancelled search alive for another
// retryDelay per attempt, which is exactly what the dispatcher cannot afford.
func (c *HiAnimeClient) sleep(ctx context.Context) {
	if c.retryDelay <= 0 {
		return
	}
	timer := time.NewTimer(c.retryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// getBody performs a GET with the client's retry policy and returns the body.
// The response is fully drained and closed here, so callers get bytes only.
//
// decorate runs after the standard headers and may be nil; it is how the embed
// fetch swaps in the Referer its host expects.
func (c *HiAnimeClient) getBody(ctx context.Context, rawURL, layer string, decorate func(*http.Request)) ([]byte, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, netx.NewParserError(sourceLabel, layer, "bad request URL", err)
		}
		c.decorateRequest(req)
		if decorate != nil {
			decorate(req)
		}

		resp, err := c.client.Do(req) // #nosec G704 -- URL is built from the pinned base
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return nil, netx.NewParserError(sourceLabel, layer, "request failed", lastErr)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxHTMLResponseBytes))
		status := resp.StatusCode
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return nil, netx.NewParserError(sourceLabel, layer, "failed to read response", lastErr)
		}

		if status != http.StatusOK {
			if status >= 500 && c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return nil, netx.NewHTTPStatusError(sourceLabel, layer, status)
		}
		return body, nil
	}
}

// fragmentResponse is the envelope both frontend endpoints answer with: a
// status flag and a chunk of rendered HTML.
type fragmentResponse struct {
	Status     bool   `json:"status"`
	TotalItems int    `json:"totalItems"`
	HTML       string `json:"html"`
}

// getFragment calls one of the frontend endpoints and returns its HTML payload.
//
// The endpoints are XHR-only in the site's own use, so they get the header that
// says so; without it a rotation could plausibly start answering with the full
// page instead of the fragment.
func (c *HiAnimeClient) getFragment(ctx context.Context, rawURL, layer, referer string) (string, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return "", netx.NewParserError(sourceLabel, layer, "bad request URL", err)
		}
		c.decorateRequest(req)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		if referer != "" {
			req.Header.Set("Referer", referer)
		}

		resp, err := c.client.Do(req) // #nosec G704 -- URL is built from the pinned base
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return "", netx.NewParserError(sourceLabel, layer, "request failed", lastErr)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			if resp.StatusCode >= 500 && c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return "", netx.NewHTTPStatusError(sourceLabel, layer, resp.StatusCode)
		}

		var payload fragmentResponse
		decodeErr := jsonx.Decode(resp.Body, maxJSONResponseBytes, &payload)
		_ = resp.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return "", netx.NewParserError(sourceLabel, layer, "malformed JSON response", lastErr)
		}
		if !payload.Status || strings.TrimSpace(payload.HTML) == "" {
			// A 200 that carries no fragment: the route still exists but the id
			// meant nothing to it. Saying so beats handing an empty document to
			// a parser that would then report "layout changed".
			return "", netx.NewParserError(sourceLabel, layer,
				"the endpoint answered with no content for this id", nil)
		}
		return payload.HTML, nil
	}
}
