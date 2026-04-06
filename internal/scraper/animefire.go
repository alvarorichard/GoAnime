// Package scraper provides web scraping functionality for animefire.io
package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	AnimefireBase = "https://animefire.io"
)

// AnimefireClient handles interactions with Animefire.io
type AnimefireClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewAnimefireClient creates a new Animefire client
func NewAnimefireClient() *AnimefireClient {
	return &AnimefireClient{
		client:     util.GetFastClient(), // Use shared fast client
		baseURL:    AnimefireBase,
		userAgent:  UserAgent,
		maxRetries: 2,
		retryDelay: 250 * time.Millisecond, // Reduced from 350ms
	}
}

// SearchAnime searches for anime on Animefire.io using the original logic
func (c *AnimefireClient) SearchAnime(query string) ([]*models.Anime, error) {
	// AnimeFire expects spaces as hyphens in the URL
	normalizedQuery := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(query)), " ", "-")
	searchURL := fmt.Sprintf("%s/pesquisar/%s", c.baseURL, normalizedQuery)

	util.Debug("AnimeFire search", "query", query, "normalized", normalizedQuery, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = c.handleStatusError(resp)
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to parse HTML: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		if c.isChallengePage(doc) {
			lastErr = errors.New("animefire returned a challenge page (try VPN or wait)")
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		animes := c.extractSearchResults(doc)
		if len(animes) == 0 {
			// Legitimate empty result set – return without error
			return []*models.Anime{}, nil
		}

		return animes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve results from AnimeFire")
}

func (c *AnimefireClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *AnimefireClient) handleStatusError(resp *http.Response) error {
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("access restricted: VPN may be required")
	}
	return fmt.Errorf("server returned: %s", resp.Status)
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

func (c *AnimefireClient) isChallengePage(doc *goquery.Document) bool {
	title := strings.ToLower(strings.TrimSpace(doc.Find("title").First().Text()))
	if strings.Contains(title, "just a moment") {
		return true
	}

	if doc.Find("#cf-wrapper").Length() > 0 || doc.Find("#challenge-form").Length() > 0 {
		return true
	}

	body := strings.ToLower(doc.Text())
	return strings.Contains(body, "cf-error") || strings.Contains(body, "cloudflare")
}

func (c *AnimefireClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	var animes []*models.Anime

	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		if urlPath, exists := s.Attr("href"); exists {
			name := strings.TrimSpace(s.Text())
			if name != "" {
				animes = append(animes, &models.Anime{
					Name: name,
					URL:  c.resolveURL(c.baseURL, urlPath),
				})
			}
		}
	})

	if len(animes) > 0 {
		return animes
	}

	doc.Find(".card_ani").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find(".ani_name a")
		title := strings.TrimSpace(titleElem.Text())
		link, exists := titleElem.Attr("href")

		if exists && title != "" {
			imgElem := s.Find(".div_img img")
			imgURL, _ := imgElem.Attr("src")
			if imgURL != "" {
				imgURL = c.resolveURL(c.baseURL, imgURL)
			}

			animes = append(animes, &models.Anime{
				Name:     title,
				URL:      c.resolveURL(c.baseURL, link),
				ImageURL: imgURL,
			})
		}
	})

	return animes
}

// resolveURL resolves relative URLs to absolute URLs
func (c *AnimefireClient) resolveURL(base, ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	if strings.HasPrefix(ref, "/") {
		return base + ref
	}
	return base + "/" + ref
}

// GetAnimeEpisodes - placeholder method, actual episodes are fetched by API layer
func (c *AnimefireClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return nil, fmt.Errorf("episodes should be fetched using API layer, not scraper")
}

// qualityOrder maps common AnimeFire quality labels to a numeric rank.
// Higher is better.
var qualityOrder = map[string]int{
	"1080p": 5, "720p": 4, "480p": 3, "360p": 2, "240p": 1,
}

// qualityRank returns the numeric rank for a quality string, defaulting to 0.
func qualityRank(q string) int {
	if r, ok := qualityOrder[strings.ToLower(q)]; ok {
		return r
	}
	return 0
}

// GetEpisodeStreamURL fetches the episode page, reads all [data-video-src]
// elements and returns the URL with the highest quality label.
// Returns ErrSourceUnavailable if the page is a challenge / block page.
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	req, err := http.NewRequest("GET", episodeURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch episode page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleStatusError(resp)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse episode page: %w", err)
	}

	if c.isChallengePage(doc) {
		return "", fmt.Errorf("animefire episode page blocked: %w", ErrSourceUnavailable)
	}

	type videoSource struct {
		url     string
		quality string
	}

	var sources []videoSource
	doc.Find("[data-video-src]").Each(func(_ int, s *goquery.Selection) {
		src, exists := s.Attr("data-video-src")
		if !exists || src == "" {
			return
		}
		quality, _ := s.Attr("data-quality")
		sources = append(sources, videoSource{url: src, quality: quality})
	})

	if len(sources) == 0 {
		return "", errors.New("no video sources found on episode page (data-video-src missing)")
	}

	// Pick the highest-quality source.
	best := sources[0]
	for _, s := range sources[1:] {
		if qualityRank(s.quality) > qualityRank(best.quality) {
			best = s
		}
	}

	util.Debugf("AnimeFire stream: selected %s (%s) from %d sources", best.url, best.quality, len(sources))
	return best.url, nil
}

// GetAnimeDetails - placeholder method, details are fetched by API layer
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}
