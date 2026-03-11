// Package scraper provides web scraping functionality for goyabu.io
package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	GoyabuBase = "https://goyabu.io"
)

// GoyabuClient handles interactions with goyabu.io
type GoyabuClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewGoyabuClient creates a new Goyabu client
func NewGoyabuClient() *GoyabuClient {
	return &GoyabuClient{
		client:     util.GetFastClient(),
		baseURL:    GoyabuBase,
		userAgent:  UserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

// goyabuSearchResult represents a search result from the WP REST API
type goyabuSearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

// SearchAnime searches for anime on goyabu.io
// Uses the WordPress REST API search endpoint
func (c *GoyabuClient) SearchAnime(query string) ([]*models.Anime, error) {
	util.Debug("Goyabu search", "query", query)

	// First, fetch the homepage to get the nonce for API authentication
	nonce, err := c.fetchNonce()
	if err != nil {
		util.Debug("Goyabu nonce fetch failed, trying HTML fallback", "error", err)
		return c.searchAnimeHTML(query)
	}

	// Use the WP REST API search endpoint
	searchURL := fmt.Sprintf("%s/wp-json/animeonline/search/?keyword=%s&nonce=%s",
		c.baseURL, url.QueryEscape(query), nonce)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		util.Debug("Goyabu API search failed, trying HTML fallback", "error", err)
		return c.searchAnimeHTML(query)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.Debug("Goyabu API returned non-200, trying HTML fallback", "status", resp.StatusCode)
		return c.searchAnimeHTML(query)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var results []goyabuSearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		// API might return a different structure, try HTML fallback
		util.Debug("Goyabu API parse failed, trying HTML fallback", "error", err)
		return c.searchAnimeHTML(query)
	}

	var animes []*models.Anime
	for _, r := range results {
		if r.Title != "" && r.URL != "" {
			animes = append(animes, &models.Anime{
				Name:     r.Title,
				URL:      c.resolveURL(c.baseURL, r.URL),
				ImageURL: r.Image,
			})
		}
	}

	if len(animes) == 0 {
		return []*models.Anime{}, nil
	}

	return animes, nil
}

// fetchNonce gets the CSRF nonce from the homepage
func (c *GoyabuClient) fetchNonce() (string, error) {
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return "", err
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract nonce from glosAP config: "nonce":"xxxxx"
	re := regexp.MustCompile(`"nonce"\s*:\s*"([a-f0-9]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) >= 2 {
		return string(matches[1]), nil
	}

	return "", errors.New("nonce not found in page")
}

// searchAnimeHTML fallback: search by scraping the anime list page
func (c *GoyabuClient) searchAnimeHTML(query string) ([]*models.Anime, error) {
	// Try searching via the anime list with query parameter
	searchURL := fmt.Sprintf("%s/?s=%s", c.baseURL, url.QueryEscape(query))

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
			lastErr = fmt.Errorf("server returned: %s", resp.Status)
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
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}

		return c.extractSearchResults(doc), nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to search Goyabu")
}

// extractSearchResults parses search results from HTML
func (c *GoyabuClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	var animes []*models.Anime

	// Look for anime card links with images and titles
	doc.Find("article a, .anime-item a, .post a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Only include anime pages
		if !strings.Contains(href, "/anime/") {
			return
		}

		title := ""
		// Try h3, h2, or img alt for the title
		if h := s.Find("h3"); h.Length() > 0 {
			title = strings.TrimSpace(h.Text())
		} else if h := s.Find("h2"); h.Length() > 0 {
			title = strings.TrimSpace(h.Text())
		} else if img := s.Find("img"); img.Length() > 0 {
			title, _ = img.Attr("alt")
			title = strings.TrimSpace(title)
		}

		if title == "" {
			return
		}

		imgURL := ""
		if img := s.Find("img"); img.Length() > 0 {
			imgURL, _ = img.Attr("src")
			if imgURL == "" {
				imgURL, _ = img.Attr("data-src")
			}
		}

		animes = append(animes, &models.Anime{
			Name:     title,
			URL:      c.resolveURL(c.baseURL, href),
			ImageURL: imgURL,
		})
	})

	// Deduplicate by URL
	seen := make(map[string]bool)
	var unique []*models.Anime
	for _, a := range animes {
		if !seen[a.URL] {
			seen[a.URL] = true
			unique = append(unique, a)
		}
	}

	return unique
}

// goyabuEpisode represents episode data from the page's JavaScript
type goyabuEpisode struct {
	ID       string `json:"id"`
	Episodio string `json:"episodio"`
	Thumb    string `json:"thumb"`
}

// GetAnimeEpisodes fetches the episode list for a given anime page
func (c *GoyabuClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("Goyabu episodes", "url", animeURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequest("GET", animeURL, nil)
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
			lastErr = fmt.Errorf("server returned: %s", resp.Status)
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		episodes := c.parseEpisodesFromJS(string(body))

		// Sort episodes by number ascending
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Num < episodes[j].Num
		})

		return episodes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve episodes from Goyabu")
}

// parseEpisodesFromJS extracts episode data from the page's JavaScript
// The episodes are stored as a JS array like: episodes = [{id:"69013",episodio:"1",...}]
func (c *GoyabuClient) parseEpisodesFromJS(html string) []models.Episode {
	var episodes []models.Episode

	// Try to find the episodes JSON array in the page source
	// Pattern: episodes data in various JS formats
	patterns := []string{
		`episodes\s*[:=]\s*(\[[\s\S]*?\])`,
		`"episodes"\s*:\s*(\[[\s\S]*?\])`,
		`episodeList\s*[:=]\s*(\[[\s\S]*?\])`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			jsonStr := matches[1]
			// Clean up JS object notation to valid JSON
			// Convert unquoted keys: {id:"69013"} -> {"id":"69013"}
			jsonStr = regexp.MustCompile(`(\w+)\s*:`).ReplaceAllString(jsonStr, `"$1":`)
			// Fix single quotes to double quotes
			jsonStr = strings.ReplaceAll(jsonStr, "'", "\"")

			var epData []goyabuEpisode
			if err := json.Unmarshal([]byte(jsonStr), &epData); err != nil {
				util.Debug("Goyabu episode JSON parse error", "error", err)
				continue
			}

			for i, ep := range epData {
				num := i + 1
				if ep.Episodio != "" {
					if parsed, err := strconv.Atoi(ep.Episodio); err == nil {
						num = parsed
					}
				}

				epURL := fmt.Sprintf("%s/%s", c.baseURL, ep.ID)

				episodes = append(episodes, models.Episode{
					Number: fmt.Sprintf("Episódio %d", num),
					Num:    num,
					URL:    epURL,
				})
			}

			if len(episodes) > 0 {
				return episodes
			}
		}
	}

	// Fallback: parse episode-item elements if present in static HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return episodes
	}

	doc.Find(".episode-item, .boxEP").Each(func(i int, s *goquery.Selection) {
		epNum, _ := s.Attr("data-episode-number")
		link := s.Find("a")
		href, exists := link.Attr("href")
		if !exists {
			return
		}

		num := i + 1
		if epNum != "" {
			if parsed, err := strconv.Atoi(epNum); err == nil {
				num = parsed
			}
		}

		episodes = append(episodes, models.Episode{
			Number: fmt.Sprintf("Episódio %d", num),
			Num:    num,
			URL:    c.resolveURL(c.baseURL, href),
		})
	})

	return episodes
}

// GetEpisodeStreamURL gets the streaming URL for a specific episode
// Goyabu uses a blogger player with token-based URL decoding via AJAX
func (c *GoyabuClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	util.Debug("Goyabu stream URL", "url", episodeURL)

	// Step 1: Fetch the episode page to get the blogger token
	req, err := http.NewRequest("GET", episodeURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch episode page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	pageHTML := string(body)

	// Strategy 1: Look for direct video URL in the page
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageHTML))
	if err == nil {
		// Check for iframe with video embed
		if src, exists := doc.Find("iframe").Attr("src"); exists && src != "" {
			return src, nil
		}

		// Check for video element
		if src, exists := doc.Find("video source").Attr("src"); exists && src != "" {
			return src, nil
		}
		if src, exists := doc.Find("video[data-video-src]").Attr("data-video-src"); exists && src != "" {
			return src, nil
		}
	}

	// Strategy 2: Extract blogger token and decode via AJAX
	bloggerToken := c.extractBloggerToken(pageHTML)
	if bloggerToken != "" {
		streamURL, err := c.decodeBloggerToken(bloggerToken)
		if err == nil && streamURL != "" {
			return streamURL, nil
		}
		util.Debug("Blogger token decode failed", "error", err)
	}

	// Strategy 3: Look for direct video URLs in script tags
	videoURLPatterns := []string{
		`"file"\s*:\s*"(https?://[^"]+\.m3u8[^"]*)"`,
		`"file"\s*:\s*"(https?://[^"]+\.mp4[^"]*)"`,
		`src\s*[:=]\s*["'](https?://[^"']+\.m3u8[^"']*)"`,
		`src\s*[:=]\s*["'](https?://[^"']+\.mp4[^"']*)"`,
		`source\s*[:=]\s*["'](https?://[^"']+\.m3u8[^"']*)"`,
	}

	for _, pattern := range videoURLPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageHTML)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not find stream URL in episode page")
}

// extractBloggerToken extracts the blogger_token from the page HTML
func (c *GoyabuClient) extractBloggerToken(html string) string {
	patterns := []string{
		`blogger_token\s*[:=]\s*["']([^"']+)["']`,
		`data-blogger-token\s*=\s*["']([^"']+)["']`,
		`"blogger_token"\s*:\s*"([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			return matches[1]
		}
	}
	return ""
}

// decodeBloggerToken calls the AJAX endpoint to decode the blogger video token
func (c *GoyabuClient) decodeBloggerToken(token string) (string, error) {
	ajaxURL := fmt.Sprintf("%s/wp-admin/admin-ajax.php", c.baseURL)

	data := url.Values{}
	data.Set("action", "decode_blogger_video")
	data.Set("blogger_token", token)

	req, err := http.NewRequest("POST", ajaxURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create AJAX request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AJAX request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AJAX response: %w", err)
	}

	// Try to parse the response as JSON with video URLs
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// Maybe it returned a direct URL string
		urlStr := strings.TrimSpace(string(body))
		if strings.HasPrefix(urlStr, "http") {
			return urlStr, nil
		}
		return "", fmt.Errorf("failed to parse AJAX response: %w", err)
	}

	// Look for video URL in response
	for _, key := range []string{"url", "file", "src", "video_url", "stream_url"} {
		if val, ok := result[key]; ok {
			if urlStr, ok := val.(string); ok && urlStr != "" {
				return urlStr, nil
			}
		}
	}

	// Check for quality options array
	if qualities, ok := result["qualities"]; ok {
		if qArr, ok := qualities.([]interface{}); ok && len(qArr) > 0 {
			// Pick highest quality
			for _, q := range qArr {
				if qMap, ok := q.(map[string]interface{}); ok {
					if src, ok := qMap["src"].(string); ok && src != "" {
						return src, nil
					}
					if file, ok := qMap["file"].(string); ok && file != "" {
						return file, nil
					}
				}
			}
		}
	}

	return "", errors.New("no video URL found in AJAX response")
}

func (c *GoyabuClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *GoyabuClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *GoyabuClient) sleep() {
	if c.retryDelay <= 0 {
		return
	}
	time.Sleep(c.retryDelay)
}

func (c *GoyabuClient) resolveURL(base, ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	if strings.HasPrefix(ref, "/") {
		return base + ref
	}
	return base + "/" + ref
}
