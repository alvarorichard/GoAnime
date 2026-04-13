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
		client:     util.GetFastClient(),
		baseURL:    AnimefireBase,
		userAgent:  UserAgent,
		maxRetries: 2,
		retryDelay: 100 * time.Millisecond,
	}
}

// SearchAnime searches for anime on Animefire.io using the original logic.
func (c *AnimefireClient) SearchAnime(query string) ([]*models.Anime, error) {
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

		if err := checkHTTPStatus(resp, "AnimeFire search"); err != nil {
			lastErr = err
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

		if err := checkChallengeDocument(doc, "AnimeFire search"); err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		animes := c.extractSearchResults(doc)
		if len(animes) == 0 {
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

func (c *AnimefireClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *AnimefireClient) sleep() {
	if c.retryDelay <= 0 {
		return
	}
	time.Sleep(c.retryDelay)
}

func (c *AnimefireClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	var animes []*models.Anime

	doc.Find(".row.ml-1.mr-1 a").Each(func(_ int, s *goquery.Selection) {
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

	doc.Find(".card_ani").Each(func(_ int, s *goquery.Selection) {
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

// resolveURL resolves relative URLs to absolute URLs.
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

		if err := checkHTTPStatus(resp, "AnimeFire episodes"); err != nil {
			lastErr = err
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

		if err := checkChallengeDocument(doc, "AnimeFire episodes"); err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		episodes := c.parseEpisodes(doc)
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

// parseEpisodes extracts a list of Episode structs from the given document.
func (c *AnimefireClient) parseEpisodes(doc *goquery.Document) []models.Episode {
	var episodes []models.Episode
	doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex").Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		num := i + 1
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

// GetEpisodeStreamURL gets the streaming URL for a specific episode from AnimeFire.
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

		if err := checkHTTPStatus(resp, "AnimeFire episode page"); err != nil {
			lastErr = err
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

		if err := checkChallengeDocument(doc, "AnimeFire episode page"); err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return "", lastErr
		}

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

// extractVideoURL extracts the video URL from an AnimeFire episode page.
func (c *AnimefireClient) extractVideoURL(doc *goquery.Document) (string, error) {
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

		validatedURL, err := validateStreamURL(src, "AnimeFire")
		if err != nil {
			return
		}

		label, _ := s.Attr("data-quality")
		sources = append(sources, videoSource{url: validatedURL, quality: qualityRanks[strings.ToLower(label)]})
	})
	if len(sources) > 0 {
		best := sources[0]
		for _, source := range sources[1:] {
			if source.quality > best.quality {
				best = source
			}
		}
		util.Debugf("AnimeFire: selected quality rank %d url %s from %d sources", best.quality, best.url, len(sources))
		return best.url, nil
	}

	if videoSrc, exists := doc.Find("video source").Attr("src"); exists && videoSrc != "" {
		return validateStreamURL(videoSrc, "AnimeFire")
	}
	if videoSrc, exists := doc.Find("video").Attr("src"); exists && videoSrc != "" {
		return validateStreamURL(videoSrc, "AnimeFire")
	}

	iframeSrc := ""
	doc.Find("iframe").Each(func(_ int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if strings.Contains(src, "blogger.com") || strings.Contains(src, "blogspot.com") {
				iframeSrc = src
			}
		}
	})
	if iframeSrc != "" {
		util.Debug("Found Blogger iframe", "src", iframeSrc)
		return validateStreamURL(iframeSrc, "AnimeFire")
	}

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
				return validateStreamURL(val, "AnimeFire")
			}
		}
	}

	html, err := doc.Html()
	if err == nil {
		if matches := extractAnimefireBloggerURL(html); matches != "" {
			return validateStreamURL(matches, "AnimeFire")
		}

		for _, re := range []*regexp.Regexp{animefireMp4Re, animefireM3U8Re} {
			if re.MatchString(html) {
				if matches := re.FindString(html); matches != "" {
					return validateStreamURL(matches, "AnimeFire")
				}
			}
		}
	}

	return "", errors.New("no video source found in the page")
}

func extractAnimefireBloggerURL(html string) string {
	const marker = "https://www.blogger.com/video.g?token="

	search := html
	offset := 0
	for {
		start := strings.Index(search, marker)
		if start < 0 {
			return ""
		}

		start += offset
		candidate := html[start:]
		if end := strings.IndexAny(candidate, "\"' <>\r\n\t"); end >= 0 {
			candidate = candidate[:end]
		}

		if isValidAnimefireBloggerURL(candidate) {
			return candidate
		}

		next := start + len(marker)
		if next >= len(html) {
			return ""
		}
		search = html[next:]
		offset = next
	}
}

func isValidAnimefireBloggerURL(rawValue string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawValue))
	if err != nil {
		return false
	}

	if parsed.Scheme != "https" || parsed.Host != "www.blogger.com" || parsed.Path != "/video.g" {
		return false
	}

	token := parsed.Query().Get("token")
	if token == "" {
		return false
	}

	for _, r := range token {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return false
	}

	return true
}

// GetAnimeDetails is a placeholder method; details are fetched by the API layer.
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}
