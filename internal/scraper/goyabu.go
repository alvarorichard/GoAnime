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

// Pre-compiled regexes for Goyabu scraping (package-level like other scrapers)
var (
	// Episode extraction patterns — ordered by likelihood
	goyabuEpisodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?:const|let|var)\s+allEpisodes\s*=\s*(\[[\s\S]*?\])\s*;`),
		regexp.MustCompile(`episodes\s*[:=]\s*(\[[\s\S]*?\])`),
		regexp.MustCompile(`"episodes"\s*:\s*(\[[\s\S]*?\])`),
		regexp.MustCompile(`episodeList\s*[:=]\s*(\[[\s\S]*?\])`),
		regexp.MustCompile(`episodios\s*[:=]\s*(\[[\s\S]*?\])`),
	}
	// Cleanup regex: matches unquoted JS keys, preserving the delimiter before them.
	// Group 1 = delimiter ({, [, , or start), Group 2 = key name.
	goyabuUnquotedKeyRe = regexp.MustCompile(`([,{\[\s]|^)(\w+)\s*:`)
	// Nonce extraction
	goyabuNonceRe = regexp.MustCompile(`"nonce"\s*:\s*"([a-f0-9]+)"`)
	// Blogger token patterns
	goyabuBloggerPatterns = []*regexp.Regexp{
		regexp.MustCompile(`blogger_token\s*[:=]\s*["']([^"']+)["']`),
		regexp.MustCompile(`data-blogger-token\s*=\s*["']([^"']+)["']`),
		regexp.MustCompile(`"blogger_token"\s*:\s*"([^"]+)"`),
	}
	// Video URL patterns
	goyabuVideoPatterns = []*regexp.Regexp{
		regexp.MustCompile(`"file"\s*:\s*"(https?://[^"]+\.m3u8[^"]*)"`),
		regexp.MustCompile(`"file"\s*:\s*"(https?://[^"]+\.mp4[^"]*)"`),
		regexp.MustCompile(`src\s*[:=]\s*["'](https?://[^"']+\.m3u8[^"']*)"`),
		regexp.MustCompile(`src\s*[:=]\s*["'](https?://[^"']+\.mp4[^"']*)"`),
		regexp.MustCompile(`source\s*[:=]\s*["'](https?://[^"']+\.m3u8[^"']*)"`),
	}
	// playersData extraction: Goyabu embeds a JS array with player info including Blogger URLs
	goyabuPlayersDataRe = regexp.MustCompile(`var\s+playersData\s*=\s*(\[.*?\])\s*;`)
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

// goyabuSearchResult represents a search result from the WP REST API.
// The API returns a map keyed by post ID, with "img" for the image field.
type goyabuSearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"img"`
}

// SearchAnime searches for anime on goyabu.io
// Uses the WordPress REST API search endpoint
func (c *GoyabuClient) SearchAnime(query string) ([]*models.Anime, error) {
	// Normalize query: replace hyphens/underscores with spaces for WordPress search
	// (CLI args arrive hyphenated like "cavaleiro-do-zodiaco" but Goyabu needs spaces)
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, "-", " ")
	query = strings.ReplaceAll(query, "_", " ")
	for strings.Contains(query, "  ") {
		query = strings.ReplaceAll(query, "  ", " ")
	}

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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		util.Debug("Goyabu API returned non-200, trying HTML fallback", "status", resp.StatusCode)
		return c.searchAnimeHTML(query)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The API returns an object keyed by post ID, not an array.
	// Error responses mix string values (e.g. {"error":"no_posts","title":"Sem resultados"})
	// with the same map shape, so decode to json.RawMessage first and skip non-object entries.
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawMap); err != nil {
		util.Debug("Goyabu API parse failed, trying HTML fallback", "error", err)
		return c.searchAnimeHTML(query)
	}

	var animes []*models.Anime
	for _, raw := range rawMap {
		// Skip string values like "no_posts" or "Sem resultados"
		if len(raw) == 0 || raw[0] == '"' {
			continue
		}
		var r goyabuSearchResult
		if err := json.Unmarshal(raw, &r); err != nil {
			continue
		}
		if r.Title != "" && r.URL != "" {
			animes = append(animes, &models.Anime{
				Name:     r.Title,
				URL:      c.resolveURL(c.baseURL, r.URL),
				ImageURL: r.Image,
			})
		}
	}

	if len(animes) == 0 {
		util.Debug("Goyabu JSON API returned 0 results, trying HTML fallback", "query", query)
		return c.searchAnimeHTML(query)
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
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract nonce from glosAP config: "nonce":"xxxxx"
	matches := goyabuNonceRe.FindSubmatch(body)
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

	for attempt := range attempts {
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

// goyabuEpisode represents episode data from the page's JavaScript.
// The API returns ID as an integer, so we use json.Number for flexibility.
type goyabuEpisode struct {
	ID       json.Number `json:"id"`
	Episodio string      `json:"episodio"`
	Link     string      `json:"link"`
	Thumb    string      `json:"thumb"`
}

// GetAnimeEpisodes fetches the episode list for a given anime page
func (c *GoyabuClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("Goyabu episodes", "url", animeURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
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

// parseEpisodesFromJS extracts episode data from the page's JavaScript.
// Goyabu stores episodes as: const allEpisodes = [{"id":41414,"episodio":"1",...}]
func (c *GoyabuClient) parseEpisodesFromJS(html string) []models.Episode {
	var episodes []models.Episode

	// Try to find the episodes JSON array in the page source.
	// The site uses `const allEpisodes = [...]` with valid JSON.
	for _, re := range goyabuEpisodePatterns {
		matches := re.FindStringSubmatch(html)
		if len(matches) < 2 {
			continue
		}

		jsonStr := matches[1]

		// Try parsing as valid JSON first (Goyabu returns proper JSON)
		var epData []goyabuEpisode
		if err := json.Unmarshal([]byte(jsonStr), &epData); err != nil {
			// Only if direct parse fails, try cleaning JS notation to JSON:
			// Convert unquoted keys ({id:1} -> {"id":1}) but skip already-quoted ones
			cleaned := goyabuUnquotedKeyRe.ReplaceAllString(jsonStr, `$1"$2":`)
			cleaned = strings.ReplaceAll(cleaned, "'", "\"")
			if err2 := json.Unmarshal([]byte(cleaned), &epData); err2 != nil {
				util.Debug("Goyabu episode JSON parse error", "error", err2)
				continue
			}
		}

		for i, ep := range epData {
			num := i + 1
			if ep.Episodio != "" {
				if parsed, err := strconv.Atoi(ep.Episodio); err == nil {
					num = parsed
				}
			}

			// Goyabu episode IDs are WordPress post IDs; use /?p=ID
			epURL := fmt.Sprintf("%s/?p=%s", c.baseURL, ep.ID.String())

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

	// Fallback: parse episode link elements from static HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return episodes
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}
		// Only follow links that look like episode pages (contain /?p= or /episode/)
		if !strings.Contains(href, "/?p=") && !strings.Contains(href, "/episode/") {
			return
		}
		if !strings.Contains(href, c.baseURL) && !strings.HasPrefix(href, "/") {
			return
		}

		num := len(episodes) + 1
		if epNum, _ := s.Attr("data-episode-number"); epNum != "" {
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
	defer func() { _ = resp.Body.Close() }()

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

	// Strategy 2: Extract playersData JSON and decode blogger token via AJAX
	bloggerToken, bloggerURL := c.extractPlayerData(pageHTML)
	if bloggerToken != "" {
		streamURL, err := c.decodeBloggerToken(bloggerToken)
		if err == nil && streamURL != "" {
			return streamURL, nil
		}
		util.Debug("Blogger token decode failed", "error", err)
	}

	// Strategy 3: Look for direct video URLs in script tags
	for _, re := range goyabuVideoPatterns {
		matches := re.FindStringSubmatch(pageHTML)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	// Strategy 4: Return Blogger embed URL as last resort (video player can handle it)
	if bloggerURL != "" {
		util.Debug("Using Blogger embed URL as fallback", "url", bloggerURL)
		return bloggerURL, nil
	}

	return "", fmt.Errorf("could not find stream URL in episode page")
}

// extractPlayerData extracts the blogger_token and Blogger embed URL from the page.
// It first tries the structured playersData JS variable, then falls back to regex patterns.
func (c *GoyabuClient) extractPlayerData(html string) (token, bloggerURL string) {
	// Try structured playersData first: var playersData = [{...}];
	if matches := goyabuPlayersDataRe.FindStringSubmatch(html); len(matches) >= 2 {
		var players []struct {
			BloggerToken string `json:"blogger_token"`
			URL          string `json:"url"`
		}
		if err := json.Unmarshal([]byte(matches[1]), &players); err == nil && len(players) > 0 {
			token = players[0].BloggerToken
			bloggerURL = players[0].URL
			util.Debug("Extracted playersData", "hasToken", token != "", "hasURL", bloggerURL != "")
		}
	}

	// Fallback: extract blogger_token from other patterns if not found
	if token == "" {
		for _, re := range goyabuBloggerPatterns {
			if m := re.FindStringSubmatch(html); len(m) >= 2 {
				token = m[1]
				break
			}
		}
	}
	return token, bloggerURL
}

// decodeBloggerToken calls the AJAX endpoint to decode the blogger video token.
// The server expects action=decode_blogger_video with token=<base64> and returns
// {"success":true,"data":{"play":[{"src":"url","size":720,"type":"video/mp4"},...]}}.
func (c *GoyabuClient) decodeBloggerToken(token string) (string, error) {
	ajaxURL := fmt.Sprintf("%s/wp-admin/admin-ajax.php", c.baseURL)

	data := url.Values{}
	data.Set("action", "decode_blogger_video")
	data.Set("token", token)

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
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AJAX response: %w", err)
	}

	// Try to parse the response as JSON with video URLs
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		// Maybe it returned a direct URL string
		urlStr := strings.TrimSpace(string(body))
		if strings.HasPrefix(urlStr, "http") {
			return urlStr, nil
		}
		return "", fmt.Errorf("failed to parse AJAX response: %w", err)
	}

	// Goyabu wraps the response in {"success":bool,"data":{...}}
	dataObj, _ := result["data"].(map[string]any)

	// Check for play array: data.play[].src (Goyabu's actual response format)
	if dataObj != nil {
		if play, ok := dataObj["play"].([]any); ok {
			bestURL, bestSize := "", float64(0)
			for _, item := range play {
				if m, ok := item.(map[string]any); ok {
					src, _ := m["src"].(string)
					size, _ := m["size"].(float64)
					if src != "" && size >= bestSize {
						bestURL, bestSize = src, size
					}
				}
			}
			if bestURL != "" {
				return bestURL, nil
			}
		}
	}

	// Fallback: look for video URL at top level or in data
	for _, obj := range []map[string]any{result, dataObj} {
		if obj == nil {
			continue
		}
		for _, key := range []string{"url", "file", "src", "video_url", "stream_url"} {
			if val, ok := obj[key]; ok {
				if urlStr, ok := val.(string); ok && strings.HasPrefix(urlStr, "http") {
					return urlStr, nil
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
