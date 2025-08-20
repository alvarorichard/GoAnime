// Package scraper provides web scraping functionality for animefire.plus
package scraper

import (
	"fmt"
	"net/http"
	"net/url"
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

// SearchAnime searches for anime on Animefire.plus using the original logic
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("access restricted: VPN may be required")
		}
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var animes []*models.Anime

	// Use the same parsing logic as the original system
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		if urlPath, exists := s.Attr("href"); exists {
			name := strings.TrimSpace(s.Text())
			if name != "" {
				fullURL := c.resolveURL(c.baseURL, urlPath)
				anime := &models.Anime{
					Name: name,
					URL:  fullURL,
				}
				animes = append(animes, anime)
			}
		}
	})

	// If no results with the primary selector, try the card-based selector as fallback
	if len(animes) == 0 {
		doc.Find(".card_ani").Each(func(i int, s *goquery.Selection) {
			titleElem := s.Find(".ani_name a")
			title := strings.TrimSpace(titleElem.Text())
			link, exists := titleElem.Attr("href")

			if exists && title != "" {
				// Get image URL
				imgElem := s.Find(".div_img img")
				imgURL, _ := imgElem.Attr("src")
				if imgURL != "" {
					imgURL = c.resolveURL(c.baseURL, imgURL)
				}

				anime := &models.Anime{
					Name:     title,
					URL:      c.resolveURL(c.baseURL, link),
					ImageURL: imgURL,
				}

				animes = append(animes, anime)
			}
		})
	}

	return animes, nil
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

// GetEpisodeStreamURL gets the streaming URL for a specific episode
// This is specific to scraper functionality, not duplicated from API
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	// Implementation for getting stream URLs - specific to scraper
	return "", fmt.Errorf("stream URL extraction not implemented")
}

// GetAnimeDetails - placeholder method, details are fetched by API layer
func (c *AnimefireClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}
