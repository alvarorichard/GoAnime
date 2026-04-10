// Package scraper provides web scraping functionality for animefire.io
package scraper

import (
	"errors"
	"fmt"
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
	AnimefireBase = "https://animefire.io"
)

// Pre-compiled regexes for AnimeFire scraper (avoid per-call compilation)
var (
	animefireBloggerRe = regexp.MustCompile(`https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`)
	animefireMp4Re     = regexp.MustCompile(`(https?://[^"'\s<>]+\.mp4(?:\?[^"'\s<>]*)?)`)
	animefireM3U8Re    = regexp.MustCompile(`(https?://[^"'\s<>]+\.m3u8(?:\?[^"'\s<>]*)?)`)
	animefireEpisodeRe = regexp.MustCompile(`(?i)epis[oó]dio\s+(\d+)`)
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
		retryDelay: 100 * time.Millisecond,
	}
}

// SearchAnime searches for anime on Animefire.io using the original logic
func (c *AnimefireClient) SearchAnime(query string) ([]*models.Anime, error) {
	// AnimeFire expects spaces as hyphens in the URL
	normalizedQuery := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(query)), " ", "-")
	searchURL := fmt.Sprintf("%s/pesquisar/%s", c.baseURL, url.PathEscape(normalizedQuery))

	util.Debug("AnimeFire search", "query", query, "normalized", normalizedQuery, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req) // #nosec G704
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
			lastErr = fmt.Errorf("animefire challenge page blocked: %w", ErrSourceUnavailable)
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

// GetAnimeEpisodes fetches and parses the list of episodes for a given anime.
// It returns a sorted slice of Episode structs, ordered by episode number.
func (c *AnimefireClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("AnimeFire episodes", "url", animeURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
		req, err := http.NewRequest("GET", animeURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req) // #nosec G704
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
			lastErr = fmt.Errorf("animefire challenge page blocked: %w", ErrSourceUnavailable)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		episodes := c.parseEpisodes(doc)

		// Sort episodes by number ascending
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Num < episodes[j].Num
		})

		return episodes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve episodes from AnimeFire")
}

// parseEpisodes extracts a list of Episode structs from the given goquery.Document.
func (c *AnimefireClient) parseEpisodes(doc *goquery.Document) []models.Episode {
	var episodes []models.Episode
	doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex").Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		num := i + 1 // default to index-based numbering
		matches := animefireEpisodeRe.FindStringSubmatch(episodeNum)
		if len(matches) >= 2 {
			parsed, err := strconv.Atoi(matches[1])
			if err != nil {
				util.Debug("Error parsing episode number", "text", episodeNum, "error", err)
				return
			}
			num = parsed
		}

		episodes = append(episodes, models.Episode{
			Number: episodeNum,
			Num:    num,
			URL:    c.resolveURL(c.baseURL, episodeURL),
		})
	})

	return episodes
}

// GetEpisodeStreamURL gets the streaming URL for a specific episode from AnimeFire
// This handles various video sources including Blogger embeds
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	util.Debug("AnimeFire stream URL extraction", "episodeURL", episodeURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
		req, err := http.NewRequest("GET", episodeURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req) // #nosec G704
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return "", lastErr
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = c.handleStatusError(resp)
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return "", lastErr
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to parse HTML: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return "", lastErr
		}

		if c.isChallengePage(doc) {
			lastErr = fmt.Errorf("animefire episode page blocked: %w", ErrSourceUnavailable)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return "", lastErr
		}

		// Try to find video URL using multiple methods
		videoURL, err := c.extractVideoURL(doc)
		if err == nil && videoURL != "" {
			util.Debug("AnimeFire video URL found", "url", videoURL)
			return videoURL, nil
		}

		lastErr = err
		if c.shouldRetry(attempt) {
			c.sleep()
			continue
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("failed to extract video URL from AnimeFire")
}

// extractVideoURL extracts the video URL from an AnimeFire episode page
func (c *AnimefireClient) extractVideoURL(doc *goquery.Document) (string, error) {
	// Method 1: Collect all [data-video-src] elements and return the highest-quality URL.
	// qualityRanks maps common AnimeFire quality labels to a numeric rank (higher = better).
	qualityRanks := map[string]int{"1080p": 5, "720p": 4, "480p": 3, "360p": 2, "240p": 1}
	type videoSource struct {
		url     string
		quality int
	}
	var sources []videoSource
	doc.Find("[data-video-src]").Each(func(_ int, s *goquery.Selection) {
		src, exists := s.Attr("data-video-src")
		if !exists || src == "" {
			return
		}
		label, _ := s.Attr("data-quality")
		sources = append(sources, videoSource{url: src, quality: qualityRanks[strings.ToLower(label)]})
	})
	if len(sources) > 0 {
		best := sources[0]
		for _, s := range sources[1:] {
			if s.quality > best.quality {
				best = s
			}
		}
		util.Debugf("AnimeFire: selected quality rank %d url %s from %d sources", best.quality, best.url, len(sources))
		return best.url, nil
	}

	// Method 2: Look for video element with src attribute
	if videoSrc, exists := doc.Find("video source").Attr("src"); exists && videoSrc != "" {
		return videoSrc, nil
	}
	if videoSrc, exists := doc.Find("video").Attr("src"); exists && videoSrc != "" {
		return videoSrc, nil
	}

	// Method 3: Look for iframe with Blogger video
	iframeSrc := ""
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if strings.Contains(src, "blogger.com") || strings.Contains(src, "blogspot.com") {
				iframeSrc = src
			}
		}
	})
	if iframeSrc != "" {
		util.Debug("Found Blogger iframe", "src", iframeSrc)
		return iframeSrc, nil
	}

	// Method 4: Look for data-video, data-src, data-url attributes in various elements.
	// Note: div[data-video-src] is already handled by Method 1 with quality ranking.
	selectors := []string{
		"div[data-video]",
		"div[data-src]",
		"div[data-url]",
		"[data-player]",
	}
	attrs := []string{"data-video", "data-src", "data-url", "data-player"}

	for i, selector := range selectors {
		if elem := doc.Find(selector); elem.Length() > 0 {
			if val, exists := elem.Attr(attrs[i]); exists && val != "" {
				return val, nil
			}
		}
	}

	// Method 5: Search in HTML content for video URLs
	html, err := doc.Html()
	if err == nil {
		// Look for Blogger video links
		if animefireBloggerRe.MatchString(html) {
			if matches := animefireBloggerRe.FindString(html); matches != "" {
				return matches, nil
			}
		}

		// Look for direct video URLs
		for _, re := range []*regexp.Regexp{animefireMp4Re, animefireM3U8Re} {
			if re.MatchString(html) {
				if matches := re.FindString(html); matches != "" {
					return matches, nil
				}
			}
		}
	}

	return "", errors.New("no video source found in the page")
}

// GetAnimeDetails - placeholder method, details are fetched by API layer
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}
