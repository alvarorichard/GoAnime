package anidb

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
	// anidbBase is the public host. Pinned here so a rotation fails loudly in
	// TestHostIsPinned instead of silently scraping the wrong site.
	anidbBase = "https://anidb.app"

	// sourceLabel is the name carried in diagnostics and in models.Anime.Source.
	sourceLabel = "AniDB"

	// Response caps for jsonx.Decode. The episode list of a long-running show is
	// a few hundred KB at most; anything larger is not a payload we want.
	maxJSONResponseBytes = 4 << 20
	maxHTMLResponseBytes = 8 << 20
)

// Package-level regexes: compiled once, per the provider contract.
var (
	// animeHrefRe matches an anime permalink and captures slug and numeric id.
	// Host-agnostic on purpose so a domain rotation only needs the const above.
	animeHrefRe = regexp.MustCompile(`/anime/([a-z0-9-]+?)-(\d+)/?$`)
	// episodeHrefRe captures the numeric episode id out of a canonical URL.
	episodeHrefRe = regexp.MustCompile(`/episode/(\d+)/?$`)
	// masterPlaylistRe pulls the HLS master URL out of the embed page's player
	// config: file: 'https://…/master.m3u8'
	masterPlaylistRe = regexp.MustCompile(`file:\s*['"]([^'"]+\.m3u8[^'"]*)['"]`)
	// variantRe reads one #EXT-X-STREAM-INF line's resolution height.
	variantRe = regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	// qualityDigitsRe pulls "1080" out of "1080p".
	qualityDigitsRe = regexp.MustCompile(`(\d{3,4})`)
)

// AniDBClient handles interactions with anidb.app.
type AniDBClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewAniDBClient creates a new anidb.app client. Performs no network I/O: it
// runs under sync.Once in the adapter.
func NewAniDBClient() *AniDBClient {
	return &AniDBClient{
		client:     util.NewFastClient(),
		baseURL:    anidbBase,
		userAgent:  netx.UserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

// NewClientForTest returns a client pointed at a test server with retries
// disabled. Only for tests.
func NewClientForTest(serverURL string) *AniDBClient {
	c := NewAniDBClient()
	c.baseURL = strings.TrimSuffix(serverURL, "/")
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func (c *AniDBClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *AniDBClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

// sleep waits out the retry delay, or returns early when ctx is cancelled.
// A plain time.Sleep here would keep a cancelled search alive for another
// retryDelay per attempt, which is exactly what the dispatcher cannot afford.
func (c *AniDBClient) sleep(ctx context.Context) {
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
func (c *AniDBClient) getBody(ctx context.Context, rawURL, layer string) ([]byte, error) {
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

// getJSON decodes a JSON endpoint straight off the wire, bounded.
func (c *AniDBClient) getJSON(ctx context.Context, rawURL, layer string, dst any) error {
	var lastErr error
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return netx.NewParserError(sourceLabel, layer, "bad request URL", err)
		}
		c.decorateRequest(req)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req) // #nosec G704 -- URL is built from the pinned base
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return netx.NewParserError(sourceLabel, layer, "request failed", lastErr)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			if resp.StatusCode >= 500 && c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return netx.NewHTTPStatusError(sourceLabel, layer, resp.StatusCode)
		}

		decodeErr := jsonx.Decode(resp.Body, maxJSONResponseBytes, dst)
		_ = resp.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			if c.shouldRetry(attempt) {
				c.sleep(ctx)
				continue
			}
			return netx.NewParserError(sourceLabel, layer, "malformed JSON response", lastErr)
		}
		return nil
	}
}
