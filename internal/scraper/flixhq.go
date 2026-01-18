// Package scraper provides web scraping functionality for FlixHQ movies and TV shows
package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	FlixHQBase      = "https://flixhq.to"
	FlixHQAPI       = "https://dec.eatmynerds.live"
	FlixHQUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"
)

// MediaType represents the type of media (movie or TV show)
type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

// FlixHQClient handles interactions with FlixHQ
type FlixHQClient struct {
	client     *http.Client
	baseURL    string
	apiURL     string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// FlixHQMedia represents a movie or TV show from FlixHQ
type FlixHQMedia struct {
	ID       string
	Title    string
	Type     MediaType
	Year     string
	ImageURL string
	URL      string
	Seasons  []FlixHQSeason
	Duration string // For movies
	Quality  string
}

// FlixHQSeason represents a TV show season
type FlixHQSeason struct {
	ID       string
	Number   int
	Title    string
	Episodes []FlixHQEpisode
}

// FlixHQEpisode represents a TV show episode
type FlixHQEpisode struct {
	ID         string
	DataID     string
	Title      string
	Number     int
	SeasonID   string
	EpisodeURL string
}

// FlixHQStreamInfo contains streaming information
type FlixHQStreamInfo struct {
	VideoURL   string
	Quality    string
	Subtitles  []FlixHQSubtitle
	Referer    string
	SourceName string
}

// FlixHQSubtitle represents a subtitle track
type FlixHQSubtitle struct {
	URL      string
	Language string
	Label    string
}

// NewFlixHQClient creates a new FlixHQ client
func NewFlixHQClient() *FlixHQClient {
	return &FlixHQClient{
		client:     util.GetFastClient(), // Use shared fast client
		baseURL:    FlixHQBase,
		apiURL:     FlixHQAPI,
		userAgent:  FlixHQUserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond, // Reduced from 500ms
	}
}

// SearchMedia searches for movies and TV shows on FlixHQ
func (c *FlixHQClient) SearchMedia(query string) ([]*FlixHQMedia, error) {
	normalizedQuery := strings.ReplaceAll(strings.TrimSpace(query), " ", "-")
	searchURL := fmt.Sprintf("%s/search/%s", c.baseURL, url.PathEscape(normalizedQuery))

	util.Debug("FlixHQ search", "query", query, "url", searchURL)

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
			lastErr = fmt.Errorf("failed to parse HTML: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		if c.isChallengePage(doc) {
			lastErr = errors.New("FlixHQ returned a challenge page (try VPN or wait)")
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		media := c.extractSearchResults(doc)
		return media, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve results from FlixHQ")
}

// GetTrending gets trending movies and TV shows
func (c *FlixHQClient) GetTrending() ([]*FlixHQMedia, error) {
	return c.getMediaFromSection("home", "trending-movies")
}

// GetRecentMovies gets recent movies
func (c *FlixHQClient) GetRecentMovies() ([]*FlixHQMedia, error) {
	return c.getMediaFromPath("movie")
}

// GetRecentTV gets recent TV shows
func (c *FlixHQClient) GetRecentTV() ([]*FlixHQMedia, error) {
	return c.getMediaFromPath("tv-show")
}

func (c *FlixHQClient) getMediaFromSection(path, section string) ([]*FlixHQMedia, error) {
	pageURL := fmt.Sprintf("%s/%s", c.baseURL, path)

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var media []*FlixHQMedia
	sectionSelector := fmt.Sprintf("#%s", section)
	doc.Find(sectionSelector).Find(".flw-item").Each(func(i int, s *goquery.Selection) {
		if m := c.parseMediaItem(s); m != nil {
			media = append(media, m)
		}
	})

	return media, nil
}

func (c *FlixHQClient) getMediaFromPath(path string) ([]*FlixHQMedia, error) {
	pageURL := fmt.Sprintf("%s/%s", c.baseURL, path)

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return c.extractSearchResults(doc), nil
}

// GetSeasons gets all seasons for a TV show
func (c *FlixHQClient) GetSeasons(mediaID string) ([]FlixHQSeason, error) {
	seasonsURL := fmt.Sprintf("%s/ajax/v2/tv/seasons/%s", c.baseURL, mediaID)

	req, err := http.NewRequest("GET", seasonsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var seasons []FlixHQSeason
	seasonNum := 1
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Extract season ID from href pattern like "javascript:;" data-id="123"
		dataID, _ := s.Attr("data-id")
		if dataID == "" {
			// Try to extract from href
			re := regexp.MustCompile(`-(\d+)$`)
			matches := re.FindStringSubmatch(href)
			if len(matches) > 1 {
				dataID = matches[1]
			}
		}

		title := strings.TrimSpace(s.Text())
		if title != "" && dataID != "" {
			seasons = append(seasons, FlixHQSeason{
				ID:     dataID,
				Number: seasonNum,
				Title:  title,
			})
			seasonNum++
		}
	})

	return seasons, nil
}

// GetEpisodes gets all episodes for a season
func (c *FlixHQClient) GetEpisodes(seasonID string) ([]FlixHQEpisode, error) {
	episodesURL := fmt.Sprintf("%s/ajax/v2/season/episodes/%s", c.baseURL, seasonID)

	req, err := http.NewRequest("GET", episodesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var episodes []FlixHQEpisode
	episodeNum := 1
	doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
		dataID, exists := s.Attr("data-id")
		if !exists {
			return
		}

		title, _ := s.Attr("title")
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}

		if dataID != "" {
			episodes = append(episodes, FlixHQEpisode{
				DataID:   dataID,
				Title:    title,
				Number:   episodeNum,
				SeasonID: seasonID,
			})
			episodeNum++
		}
	})

	return episodes, nil
}

// GetEpisodeServerID gets the server ID for an episode
func (c *FlixHQClient) GetEpisodeServerID(dataID, provider string) (string, error) {
	serversURL := fmt.Sprintf("%s/ajax/v2/episode/servers/%s", c.baseURL, dataID)

	req, err := http.NewRequest("GET", serversURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var episodeID string
	doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
		serverTitle, _ := s.Attr("title")
		if strings.EqualFold(serverTitle, provider) {
			episodeID, _ = s.Attr("data-id")
		}
	})

	if episodeID == "" {
		// Fallback to first available server
		doc.Find(".nav-item a").First().Each(func(i int, s *goquery.Selection) {
			episodeID, _ = s.Attr("data-id")
		})
	}

	if episodeID == "" {
		return "", errors.New("no server found for episode")
	}

	return episodeID, nil
}

// GetMovieServerID gets the server ID for a movie
func (c *FlixHQClient) GetMovieServerID(mediaID, provider string) (string, error) {
	movieURL := fmt.Sprintf("%s/ajax/movie/episodes/%s", c.baseURL, mediaID)

	req, err := http.NewRequest("GET", movieURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var episodeID string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		serverTitle, _ := s.Attr("title")
		href, _ := s.Attr("href")

		if strings.EqualFold(serverTitle, provider) {
			// Extract episode ID from href like /watch-movie-123.456
			re := regexp.MustCompile(`\.(\d+)$`)
			matches := re.FindStringSubmatch(href)
			if len(matches) > 1 {
				episodeID = matches[1]
			}
		}
	})

	if episodeID == "" {
		// Fallback to first available server
		doc.Find("a").First().Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			re := regexp.MustCompile(`\.(\d+)$`)
			matches := re.FindStringSubmatch(href)
			if len(matches) > 1 {
				episodeID = matches[1]
			}
		})
	}

	if episodeID == "" {
		return "", errors.New("no server found for movie")
	}

	return episodeID, nil
}

// GetEmbedLink gets the embed link for streaming
func (c *FlixHQClient) GetEmbedLink(episodeID string) (string, error) {
	sourcesURL := fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, episodeID)

	req, err := http.NewRequest("GET", sourcesURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Link string `json:"link"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Link == "" {
		return "", errors.New("no embed link found")
	}

	return result.Link, nil
}

// ExtractStreamInfo extracts video URL and subtitles from embed link
func (c *FlixHQClient) ExtractStreamInfo(embedLink string, preferredQuality string, subsLanguage string) (*FlixHQStreamInfo, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, url.QueryEscape(embedLink))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		File    string `json:"file"`
		Sources []struct {
			File    string `json:"file"`
			Type    string `json:"type"`
			Quality string `json:"quality"`
		} `json:"sources"`
		Tracks []struct {
			File    string `json:"file"`
			Label   string `json:"label"`
			Kind    string `json:"kind"`
			Default bool   `json:"default"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	streamInfo := &FlixHQStreamInfo{}

	// Get video URL
	if result.File != "" {
		streamInfo.VideoURL = result.File
		// Apply quality preference
		if preferredQuality != "" && preferredQuality != "auto" {
			streamInfo.VideoURL = strings.Replace(streamInfo.VideoURL, "/playlist.m3u8",
				fmt.Sprintf("/%s/index.m3u8", preferredQuality), 1)
		}
	} else if len(result.Sources) > 0 {
		// Find preferred quality or use first
		for _, source := range result.Sources {
			if source.Quality == preferredQuality {
				streamInfo.VideoURL = source.File
				break
			}
		}
		if streamInfo.VideoURL == "" {
			streamInfo.VideoURL = result.Sources[0].File
		}
	}

	if streamInfo.VideoURL == "" {
		return nil, errors.New("no video URL found")
	}

	// Get subtitles matching preferred language
	for _, track := range result.Tracks {
		if track.Kind == "captions" || track.Kind == "subtitles" {
			sub := FlixHQSubtitle{
				URL:      track.File,
				Label:    track.Label,
				Language: c.extractLanguageFromLabel(track.Label),
			}
			streamInfo.Subtitles = append(streamInfo.Subtitles, sub)
		}
	}

	// Filter subtitles by preferred language if specified
	if subsLanguage != "" && len(streamInfo.Subtitles) > 0 {
		var filteredSubs []FlixHQSubtitle
		for _, sub := range streamInfo.Subtitles {
			if strings.Contains(strings.ToLower(sub.Language), strings.ToLower(subsLanguage)) ||
				strings.Contains(strings.ToLower(sub.Label), strings.ToLower(subsLanguage)) {
				filteredSubs = append(filteredSubs, sub)
			}
		}
		if len(filteredSubs) > 0 {
			streamInfo.Subtitles = filteredSubs
		}
	}

	return streamInfo, nil
}

// GetStreamURL is a convenience method that combines all steps to get a stream URL
func (c *FlixHQClient) GetStreamURL(media *FlixHQMedia, episode *FlixHQEpisode, provider, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	var episodeID string
	var err error

	if media.Type == MediaTypeMovie {
		episodeID, err = c.GetMovieServerID(media.ID, provider)
	} else {
		episodeID, err = c.GetEpisodeServerID(episode.DataID, provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get server ID: %w", err)
	}

	embedLink, err := c.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	streamInfo, err := c.ExtractStreamInfo(embedLink, quality, subsLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract stream info: %w", err)
	}

	return streamInfo, nil
}

// Helper methods

func (c *FlixHQClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *FlixHQClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *FlixHQClient) sleep() {
	if c.retryDelay > 0 {
		time.Sleep(c.retryDelay)
	}
}

func (c *FlixHQClient) isChallengePage(doc *goquery.Document) bool {
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

func (c *FlixHQClient) extractSearchResults(doc *goquery.Document) []*FlixHQMedia {
	var media []*FlixHQMedia

	doc.Find(".flw-item").Each(func(i int, s *goquery.Selection) {
		if m := c.parseMediaItem(s); m != nil {
			media = append(media, m)
		}
	})

	return media
}

func (c *FlixHQClient) parseMediaItem(s *goquery.Selection) *FlixHQMedia {
	// Get image URL
	imgElem := s.Find("img")
	imageURL, _ := imgElem.Attr("data-src")
	if imageURL == "" {
		imageURL, _ = imgElem.Attr("src")
	}

	// Get link and extract media info
	linkElem := s.Find(".film-name a, .film-detail a")
	href, exists := linkElem.Attr("href")
	if !exists {
		return nil
	}

	title := strings.TrimSpace(linkElem.Text())
	if title == "" {
		title, _ = linkElem.Attr("title")
	}

	if title == "" {
		return nil
	}

	// Determine media type and ID from URL
	var mediaType MediaType
	var mediaID string

	if strings.Contains(href, "/tv/") {
		mediaType = MediaTypeTV
	} else if strings.Contains(href, "/movie/") {
		mediaType = MediaTypeMovie
	} else {
		return nil
	}

	// Extract ID from URL (e.g., /movie/watch-movie-name-12345 -> 12345)
	re := regexp.MustCompile(`-(\d+)$`)
	matches := re.FindStringSubmatch(href)
	if len(matches) > 1 {
		mediaID = matches[1]
	}

	if mediaID == "" {
		return nil
	}

	// Get year/info from fdi-item
	year := ""
	s.Find(".fdi-item").Each(func(i int, item *goquery.Selection) {
		text := strings.TrimSpace(item.Text())
		if text != "" {
			if i == 0 {
				year = text
			}
		}
	})

	return &FlixHQMedia{
		ID:       mediaID,
		Title:    title,
		Type:     mediaType,
		Year:     year,
		ImageURL: imageURL,
		URL:      c.resolveURL(href),
	}
}

func (c *FlixHQClient) resolveURL(ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	if strings.HasPrefix(ref, "/") {
		return c.baseURL + ref
	}
	return c.baseURL + "/" + ref
}

func (c *FlixHQClient) extractLanguageFromLabel(label string) string {
	label = strings.ToLower(label)
	languages := map[string]string{
		"english":    "en",
		"spanish":    "es",
		"portuguese": "pt",
		"french":     "fr",
		"german":     "de",
		"italian":    "it",
		"japanese":   "ja",
		"korean":     "ko",
		"chinese":    "zh",
		"arabic":     "ar",
		"russian":    "ru",
	}

	for lang, code := range languages {
		if strings.Contains(label, lang) {
			return code
		}
	}

	return label
}

// ToAnimeModel converts FlixHQMedia to models.Anime for compatibility
func (m *FlixHQMedia) ToAnimeModel() *models.Anime {
	return &models.Anime{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.ImageURL,
		Source:   "FlixHQ",
	}
}

// ToEpisodeModel converts FlixHQEpisode to models.Episode for compatibility
func (e *FlixHQEpisode) ToEpisodeModel() models.Episode {
	return models.Episode{
		Number: fmt.Sprintf("%d", e.Number),
		Num:    e.Number,
		URL:    e.EpisodeURL,
		Title: models.TitleDetails{
			English: e.Title,
			Romaji:  e.Title,
		},
	}
}
