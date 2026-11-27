package miruro

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

const (
	// miruroBase is the public host. Pinned here so a rotation fails loudly in
	// TestHostIsPinned instead of silently scraping the wrong site.
	miruroBase = "https://miruro.cx"

	// sourceLabel is the name carried in diagnostics and in models.Anime.Source.
	sourceLabel = "Miruro"

	// dataSuffix turns a page path into its SvelteKit data endpoint:
	// /search?q=x is the page, /search/__data.json?q=x its load data.
	dataSuffix = "/__data.json"

	// maxDataBytes caps one data response. One Piece's watch page, the
	// largest seen (1180 episodes), weighs about 950KB.
	maxDataBytes = 16 << 20
	// maxPlaylistBytes caps a playlist read while picking a stream.
	maxPlaylistBytes = 2 << 20

	// maxRedirects bounds how many SvelteKit redirects one fetch follows. A
	// watch URL with the wrong slug costs exactly one.
	maxRedirects = 2
)

// Client handles interactions with miruro.cx.
type Client struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
	// probeTimeout bounds each stream candidate's playlist check, so a dead
	// mirror costs seconds and the next one gets its turn.
	probeTimeout time.Duration
}

// NewClient creates a new miruro.cx client. Performs no network I/O: it
// runs under sync.Once in the adapter.
func NewClient() *Client {
	return &Client{
		client:       util.NewFastClient(),
		baseURL:      miruroBase,
		userAgent:    netx.UserAgent,
		maxRetries:   2,
		retryDelay:   300 * time.Millisecond,
		probeTimeout: 6 * time.Second,
	}
}

// NewClientForTest returns a client pointed at a test server with retries
// disabled. Only for tests.
func NewClientForTest(serverURL string) *Client {
	c := NewClient()
	c.baseURL = strings.TrimSuffix(serverURL, "/")
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func (c *Client) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, */*;q=0.8")
	req.Header.Set("Accept-Language", netx.EnglishAcceptLanguage)
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *Client) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

// sleep waits out the retry delay, or returns early when ctx is cancelled.
func (c *Client) sleep(ctx context.Context) {
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

// getData fetches the SvelteKit load data of one of the site's pages, path
// being the page ("/search", "/watch/<id>/<slug>") and query its query
// string. A redirect the app answers with (a watch URL whose slug is not the
// canonical one) is followed, but only to another path on the same site.
func (c *Client) getData(ctx context.Context, path string, query url.Values, layer string) (dataResponse, error) {
	for hop := 0; ; hop++ {
		rawURL := c.baseURL + strings.TrimSuffix(path, "/") + dataSuffix
		if len(query) > 0 {
			rawURL += "?" + query.Encode()
		}
		body, err := c.getBody(ctx, rawURL, layer)
		if err != nil {
			return dataResponse{}, err
		}
		resp, err := parseDataResponse(body)
		if err != nil {
			return dataResponse{}, netx.NewParserError(sourceLabel, layer, "unreadable page data", err)
		}
		if resp.Type != "redirect" {
			resp.Path = path
			return resp, nil
		}
		if hop >= maxRedirects {
			return dataResponse{}, netx.NewParserError(sourceLabel, layer, "too many redirects", nil)
		}
		next, err := sitePath(resp.Location)
		if err != nil {
			return dataResponse{}, netx.NewParserError(sourceLabel, layer, "redirect leaves the site", err)
		}
		util.Debug("Miruro redirect", "from", path, "to", next.Path)
		path, query = next.Path, next.Query()
	}
}

// sitePath accepts a redirect target only when it is a path on this site:
// "/watch/x/y?ep=1" yes, "https://elsewhere/..." and "//elsewhere" no.
func sitePath(location string) (*url.URL, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "" || u.Host != "" || !strings.HasPrefix(u.Path, "/") {
		return nil, fmt.Errorf("%s: %w", location, errOffSite)
	}
	return u, nil
}

var errOffSite = errors.New("not a path on miruro.cx")

// getBody performs a GET with the client's retry policy and returns the body.
func (c *Client) getBody(ctx context.Context, rawURL, layer string) ([]byte, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, netx.NewParserError(sourceLabel, layer, "bad request URL", err)
		}
		c.decorateRequest(req)

		resp, err := c.client.Do(req) // #nosec G704 -- URL is built from the pinned base
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return nil, netx.NewParserError(sourceLabel, layer, "request failed", lastErr)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxDataBytes))
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

		// Cloudflare sits in front of the site. Its interstitial is HTML with
		// a 403 or 503, and the data endpoints never answer HTML otherwise.
		if isChallenge(resp.Header.Get("Content-Type"), body) {
			return nil, netx.NewBlockedChallengeError(sourceLabel, layer, "Cloudflare challenge instead of page data", nil)
		}
		if status != http.StatusOK {
			if status >= 500 && c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return nil, netx.NewHTTPStatusError(sourceLabel, layer, status)
		}
		if err := netx.CheckHTMLResponse(resp, body, sourceLabel); err != nil {
			return nil, err
		}
		return body, nil
	}
}

// isChallenge reports a Cloudflare interstitial.
func isChallenge(contentType string, body []byte) bool {
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return false
	}
	lower := bytes.ToLower(body)
	return bytes.Contains(lower, []byte("<title>just a moment")) ||
		bytes.Contains(lower, []byte("cf-chl-")) ||
		bytes.Contains(lower, []byte("attention required! | cloudflare"))
}
