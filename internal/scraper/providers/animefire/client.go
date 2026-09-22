// Package scraper provides web scraping functionality for animefire.io
package animefire

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// AnimefireBase is the site's origin. It moved from animefire.io to
	// animefire.one when the site was rebuilt as a single-page app; the old
	// host still redirects, but naming the live one keeps the redirect off the
	// critical path.
	AnimefireBase = siteBase
)

// AnimefireClient handles interactions with Animefire.io
type AnimefireClient struct {
	client  *http.Client
	baseURL string
	// apiBase is the JSON API this client reads. A field rather than the
	// constant so a test can point it at an httptest server.
	apiBase    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewAnimefireClient creates a new Animefire client
func NewAnimefireClient() *AnimefireClient {
	return &AnimefireClient{
		client:     util.NewFastClient(),
		baseURL:    AnimefireBase,
		apiBase:    resolveAPIBase(),
		userAgent:  netx.UserAgent,
		maxRetries: 2,
		retryDelay: 100 * time.Millisecond,
	}
}

// retrying runs an API call under the client's existing retry policy.
//
// The three public methods used to each carry their own copy of this loop,
// interleaved with HTML fetching and parsing. With the parsing gone there is
// one shape left, so there is one loop.
func retrying[T any](c *AnimefireClient, call func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := range c.maxRetries + 1 {
		got, err := call()
		if err == nil {
			return got, nil
		}
		lastErr = err
		if c.shouldRetry(attempt) {
			c.sleep()
			continue
		}
		break
	}
	if lastErr == nil {
		lastErr = errors.New("AnimeFire: request failed")
	}
	return zero, lastErr
}

func (c *AnimefireClient) SearchAnime(query string) ([]*models.Anime, error) {
	util.Debug("AnimeFire search", "query", query)
	return retrying(c, func() ([]*models.Anime, error) { return c.searchAPI(query) })
}

func (c *AnimefireClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", netx.AcceptLanguage)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *AnimefireClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *AnimefireClient) sleep() {
	if c.retryDelay <= 0 {
		return
	}
	time.Sleep(c.retryDelay)
}

// GetAnimeEpisodes fetches and parses the list of episodes for a given anime.
func (c *AnimefireClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("AnimeFire episodes", "url", animeURL)
	return retrying(c, func() ([]models.Episode, error) { return c.episodesAPI(animeURL) })
}

// GetEpisodeStreamURL gets the streaming URL for a specific episode from AnimeFire.
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	util.Debug("AnimeFire stream URL", "episodeURL", episodeURL)
	return retrying(c, func() (string, error) { return c.streamAPI(episodeURL) })
}

// GetAnimeDetails is a placeholder method; details are fetched by the API layer.
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}

// NewClientForTest returns a client pointed at a test server with retries
// disabled. Only for tests.
func NewClientForTest(serverURL string) *AnimefireClient {
	c := NewAnimefireClient()
	c.baseURL = serverURL
	c.apiBase = serverURL
	c.client = &http.Client{Timeout: 5 * time.Second}
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}
