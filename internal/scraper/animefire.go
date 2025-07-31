// Package scraper provides web scraping functionality for animefire.plus
package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
)

const (
	AnimefireBase = "https://animefire.plus"
)

// AnimefireClient handles interactions with Animefire.plus
type AnimefireClient struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

// NewAnimefireClient creates a new Animefire client
func NewAnimefireClient() *AnimefireClient {
	return &AnimefireClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   AnimefireBase,
		userAgent: UserAgent,
	}
}

// SearchAnime searches for anime on Animefire.plus
func (c *AnimefireClient) SearchAnime(query string) ([]*models.Anime, error) {
	searchURL := fmt.Sprintf("%s/pesquisar/%s", c.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var animes []*models.Anime

	// Parse search results
	doc.Find(".card_ani").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find(".ani_name a")
		title := strings.TrimSpace(titleElem.Text())
		link, exists := titleElem.Attr("href")

		if exists && title != "" {
			// Get image URL
			imgElem := s.Find(".div_img img")
			imgURL, _ := imgElem.Attr("src")

			// Get episode count from description if available
			description := s.Find(".ani_desc").Text()
			episodeCount := c.extractEpisodeCount(description)

			anime := &models.Anime{
				Name:     title,
				URL:      c.baseURL + link,
				ImageURL: imgURL,
			}

			if episodeCount > 0 {
				anime.Name = fmt.Sprintf("%s (%d episodes)", title, episodeCount)
			}

			animes = append(animes, anime)
		}
	})

	return animes, nil
}

// GetAnimeEpisodes gets the list of episodes for an anime
func (c *AnimefireClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var episodes []models.Episode

	// Find episodes list
	doc.Find(".lcp_catlist li").Each(func(i int, s *goquery.Selection) {
		linkElem := s.Find("a")
		title := strings.TrimSpace(linkElem.Text())
		episodeURL, exists := linkElem.Attr("href")

		if exists && title != "" {
			// Extract episode number from title
			episodeNum := c.extractEpisodeNumber(title)

			episode := models.Episode{
				Number: fmt.Sprintf("%d", episodeNum),
				Num:    episodeNum,
				URL:    episodeURL,
				Title: models.TitleDetails{
					Romaji: title,
				},
			}

			episodes = append(episodes, episode)
		}
	})

	// Alternative parsing for different page structures
	if len(episodes) == 0 {
		doc.Find(".ep_box").Each(func(i int, s *goquery.Selection) {
			linkElem := s.Find("a")
			title := strings.TrimSpace(linkElem.Text())
			episodeURL, exists := linkElem.Attr("href")

			if exists && title != "" {
				episodeNum := c.extractEpisodeNumber(title)

				episode := models.Episode{
					Number: fmt.Sprintf("%d", episodeNum),
					Num:    episodeNum,
					URL:    episodeURL,
					Title: models.TitleDetails{
						Romaji: title,
					},
				}

				episodes = append(episodes, episode)
			}
		})
	}

	return episodes, nil
}

// GetEpisodeStreamURL gets the streaming URL for a specific episode
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	req, err := http.NewRequest("GET", episodeURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Look for video sources
	var streamURL string

	// Check for iframe sources
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && strings.Contains(src, "player") {
			streamURL = src
			return
		}
	})

	// Check for direct video sources
	if streamURL == "" {
		doc.Find("video source").Each(func(i int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if exists {
				streamURL = src
				return
			}
		})
	}

	// Check for script-embedded URLs
	if streamURL == "" {
		streamURL = c.extractStreamURLFromScript(doc)
	}

	if streamURL == "" {
		return "", fmt.Errorf("no stream URL found")
	}

	// Ensure URL is absolute
	if strings.HasPrefix(streamURL, "/") {
		streamURL = c.baseURL + streamURL
	} else if !strings.HasPrefix(streamURL, "http") {
		streamURL = c.baseURL + "/" + streamURL
	}

	return streamURL, nil
}

// extractEpisodeCount extracts episode count from description text
func (c *AnimefireClient) extractEpisodeCount(description string) int {
	re := regexp.MustCompile(`(\d+)\s*episódios?`)
	matches := re.FindStringSubmatch(strings.ToLower(description))
	if len(matches) >= 2 {
		count := regexp.MustCompile(`\d+`).FindString(matches[1])
		if count != "" {
			if num, err := strconv.Atoi(count); err == nil {
				return num
			}
		}
	}
	return 0
}

// extractEpisodeNumber extracts episode number from title
func (c *AnimefireClient) extractEpisodeNumber(title string) int {
	// Common patterns for episode numbers
	patterns := []string{
		`episódio\s*(\d+)`,
		`ep\.?\s*(\d+)`,
		`episode\s*(\d+)`,
		`#(\d+)`,
		`(\d+)$`, // Number at the end
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) >= 2 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return num
			}
		}
	}

	return 1 // Default to episode 1 if not found
}

// extractStreamURLFromScript extracts stream URL from JavaScript code
func (c *AnimefireClient) extractStreamURLFromScript(doc *goquery.Document) string {
	var streamURL string

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()

		// Look for common video URL patterns in JavaScript
		patterns := []string{
			`["']([^"']*\.mp4[^"']*)["']`,
			`["']([^"']*\.m3u8[^"']*)["']`,
			`src:\s*["']([^"']+)["']`,
			`file:\s*["']([^"']+)["']`,
			`video_url["']\s*:\s*["']([^"']+)["']`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(scriptContent)
			if len(matches) >= 2 {
				url := matches[1]
				// Basic validation for video URLs
				if strings.Contains(url, ".mp4") || strings.Contains(url, ".m3u8") || strings.Contains(url, "player") {
					streamURL = url
					return
				}
			}
		}
	})

	return streamURL
}

// GetAnimeDetails fetches detailed information about an anime
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	anime := &models.Anime{
		URL: animeURL,
	}

	// Get title
	titleElem := doc.Find("h1, .ani_title, .anime_title").First()
	anime.Name = strings.TrimSpace(titleElem.Text())

	// Get image
	imgElem := doc.Find(".ani_capa img, .anime_img img").First()
	if imgURL, exists := imgElem.Attr("src"); exists {
		anime.ImageURL = imgURL
	}

	// Get description and other details from the page
	descElem := doc.Find(".ani_desc, .anime_description, .sinopse").First()
	description := strings.TrimSpace(descElem.Text())

	// Create basic details structure
	anime.Details = models.AniListDetails{
		Description: description,
	}

	return anime, nil
}
