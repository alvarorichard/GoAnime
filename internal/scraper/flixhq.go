// Package scraper provides web scraping functionality for FlixHQ movies and TV shows
package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
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

// Quality represents video quality levels
type Quality string

const (
	QualityAuto Quality = "auto"
	Quality360  Quality = "360"
	Quality480  Quality = "480"
	Quality720  Quality = "720"
	Quality1080 Quality = "1080"
	QualityBest Quality = "best"
)

// StreamType represents the type of stream
type StreamType string

const (
	StreamTypeHLS StreamType = "hls"
	StreamTypeMP4 StreamType = "mp4"
)

// ServerName represents known streaming servers
type ServerName string

const (
	ServerVidcloud  ServerName = "Vidcloud"
	ServerUpCloud   ServerName = "UpCloud"
	ServerVoe       ServerName = "Voe"
	ServerMixDrop   ServerName = "MixDrop"
	ServerFilelions ServerName = "Filelions"
)

// DefaultServerPriority defines the preferred server order
var DefaultServerPriority = []ServerName{
	ServerVidcloud,
	ServerUpCloud,
	ServerVoe,
	ServerMixDrop,
	ServerFilelions,
}

// FlixHQClient handles interactions with FlixHQ
type FlixHQClient struct {
	client      *http.Client
	baseURL     string
	apiURL      string
	userAgent   string
	maxRetries  int
	retryDelay  time.Duration
	searchCache sync.Map // Caches search results
	infoCache   sync.Map // Caches media info
	serverCache sync.Map // Caches server lists
}

// FlixHQMedia represents a movie or TV show from FlixHQ
type FlixHQMedia struct {
	ID          string
	Title       string
	Type        MediaType
	Year        string
	ImageURL    string
	URL         string
	Seasons     []FlixHQSeason
	Duration    string // For movies
	Quality     string
	Description string
	Genres      []string
	Rating      float64
	ReleaseDate string
	Country     string
	Production  string
	Casts       []string
	Source      string // Source identifier (FlixHQ, SFlix, etc.)
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
	Season     int
	SeasonID   string
	EpisodeURL string
}

// FlixHQServer represents a streaming server
type FlixHQServer struct {
	Name ServerName
	ID   string
	URL  string
}

// FlixHQStreamInfo contains streaming information
type FlixHQStreamInfo struct {
	VideoURL   string
	Quality    string
	Subtitles  []FlixHQSubtitle
	Referer    string
	SourceName string
	StreamType StreamType
	IsM3U8     bool
	Headers    map[string]string
	Qualities  []FlixHQQualityOption // Available quality options
}

// FlixHQQualityOption represents an available quality option
type FlixHQQualityOption struct {
	Quality Quality
	URL     string
	IsM3U8  bool
}

// FlixHQSubtitle represents a subtitle track
type FlixHQSubtitle struct {
	URL       string
	Language  string
	Label     string
	IsForced  bool
	IsDefault bool
}

// FlixHQVideoSources contains parsed video sources
type FlixHQVideoSources struct {
	Sources   []FlixHQSource
	Subtitles []FlixHQSubtitle
}

// FlixHQSource represents a video source
type FlixHQSource struct {
	URL     string
	Quality string
	IsM3U8  bool
	Referer string
}

// NewFlixHQClient creates a new FlixHQ client
func NewFlixHQClient() *FlixHQClient {
	return &FlixHQClient{
		client:     util.GetFastClient(), // Use shared fast client
		baseURL:    FlixHQBase,
		apiURL:     FlixHQAPI,
		userAgent:  FlixHQUserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

// NewFlixHQClientWithContext creates a new FlixHQ client with custom settings
func NewFlixHQClientWithContext(timeout time.Duration, maxRetries int) *FlixHQClient {
	return &FlixHQClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    FlixHQBase,
		apiURL:     FlixHQAPI,
		userAgent:  FlixHQUserAgent,
		maxRetries: maxRetries,
		retryDelay: 300 * time.Millisecond,
	}
}

// SearchMedia searches for movies and TV shows on FlixHQ
func (c *FlixHQClient) SearchMedia(query string) ([]*FlixHQMedia, error) {
	return c.SearchMediaWithContext(context.Background(), query)
}

// SearchMediaWithContext searches for movies and TV shows on FlixHQ with context support
func (c *FlixHQClient) SearchMediaWithContext(ctx context.Context, query string) ([]*FlixHQMedia, error) {
	// Check cache first
	cacheKey := strings.ToLower(strings.TrimSpace(query))
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*FlixHQMedia), nil
	}

	// Replace non-word characters with hyphens (matching greg implementation)
	re := regexp.MustCompile(`[\W_]+`)
	cleanQuery := re.ReplaceAllString(query, "-")
	searchURL := fmt.Sprintf("%s/search/%s", c.baseURL, url.PathEscape(cleanQuery))

	util.Debug("FlixHQ search", "query", query, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
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

		// Cache results
		c.searchCache.Store(cacheKey, media)
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

	util.Debug("FlixHQ get episode servers", "url", serversURL, "provider", provider)

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

	// Try multiple providers in order of preference
	providers := []string{provider, "Vidcloud", "UpCloud", "Voe", "MixDrop"}
	var episodeID string

	for _, prov := range providers {
		doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
			if episodeID != "" {
				return // Already found
			}
			serverTitle, _ := s.Attr("title")
			if strings.EqualFold(serverTitle, prov) {
				id, exists := s.Attr("data-id")
				if exists && id != "" {
					episodeID = id
					util.Debug("FlixHQ found server", "provider", serverTitle, "id", id)
				}
			}
		})
		if episodeID != "" {
			break
		}
	}

	if episodeID == "" {
		// Fallback to first available server
		doc.Find(".nav-item a").First().Each(func(i int, s *goquery.Selection) {
			id, exists := s.Attr("data-id")
			if exists {
				episodeID = id
				title, _ := s.Attr("title")
				util.Debug("FlixHQ using fallback server", "provider", title, "id", id)
			}
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

	util.Debug("FlixHQ get movie servers", "url", movieURL, "provider", provider)

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

	// Try multiple providers in order of preference
	providers := []string{provider, "Vidcloud", "UpCloud", "Voe", "MixDrop"}
	var episodeID string

	for _, prov := range providers {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if episodeID != "" {
				return // Already found
			}
			serverTitle, _ := s.Attr("title")
			href, _ := s.Attr("href")

			if strings.EqualFold(serverTitle, prov) {
				// Extract episode ID from href like /watch-movie-123.456
				re := regexp.MustCompile(`\.(\d+)$`)
				matches := re.FindStringSubmatch(href)
				if len(matches) > 1 {
					episodeID = matches[1]
					util.Debug("FlixHQ found movie server", "provider", serverTitle, "id", episodeID)
				}
			}
		})
		if episodeID != "" {
			break
		}
	}

	if episodeID == "" {
		// Fallback to first available server
		doc.Find("a").First().Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			re := regexp.MustCompile(`\.(\d+)$`)
			matches := re.FindStringSubmatch(href)
			if len(matches) > 1 {
				episodeID = matches[1]
				title, _ := s.Attr("title")
				util.Debug("FlixHQ using fallback movie server", "provider", title, "id", episodeID)
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

	util.Debug("FlixHQ get embed link", "url", sourcesURL, "episodeID", episodeID)

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

	// Read full body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	util.Debug("FlixHQ embed response", "body", string(bodyBytes))

	var result struct {
		Link  string `json:"link"`
		Type  string `json:"type"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w (body: %s)", err, string(bodyBytes))
	}

	if result.Error != "" {
		return "", fmt.Errorf("embed API error: %s", result.Error)
	}

	if result.Link == "" {
		return "", errors.New("no embed link found")
	}

	util.Debug("FlixHQ embed link found", "link", result.Link)
	return result.Link, nil
}

// ExtractStreamInfo extracts video URL and subtitles from embed link
func (c *FlixHQClient) ExtractStreamInfo(embedLink string, preferredQuality string, subsLanguage string) (*FlixHQStreamInfo, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, url.QueryEscape(embedLink))

	util.Debug("FlixHQ API request", "url", apiURL, "embed", embedLink)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	// Read the full response body for better error handling
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	util.Debug("FlixHQ API response", "body", string(bodyBytes))

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
		// Additional fields that might be returned
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(bodyBytes))
	}

	// Check for error messages
	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}
	if result.Message != "" && result.File == "" && len(result.Sources) == 0 {
		return nil, fmt.Errorf("API message: %s", result.Message)
	}

	streamInfo := &FlixHQStreamInfo{
		Qualities: make([]FlixHQQualityOption, 0),
	}

	// Build quality options from sources
	if result.File != "" {
		streamInfo.Qualities = append(streamInfo.Qualities, FlixHQQualityOption{
			Quality: QualityAuto,
			URL:     result.File,
			IsM3U8:  strings.Contains(result.File, ".m3u8"),
		})
	}
	for _, source := range result.Sources {
		streamInfo.Qualities = append(streamInfo.Qualities, FlixHQQualityOption{
			Quality: parseQuality(source.Quality),
			URL:     source.File,
			IsM3U8:  strings.Contains(source.Type, "hls") || strings.Contains(source.File, ".m3u8"),
		})
	}

	// Get video URL
	if result.File != "" {
		streamInfo.VideoURL = result.File
		util.Debug("FlixHQ got file", "url", result.File)
		// Apply quality preference
		if preferredQuality != "" && preferredQuality != "auto" && preferredQuality != "best" {
			streamInfo.VideoURL = strings.Replace(streamInfo.VideoURL, "/playlist.m3u8",
				fmt.Sprintf("/%s/index.m3u8", preferredQuality), 1)
		}
	} else if len(result.Sources) > 0 {
		util.Debug("FlixHQ got sources", "count", len(result.Sources))
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
	return c.GetStreamURLWithContext(context.Background(), media, episode, provider, quality, subsLanguage)
}

// GetStreamURLWithContext is a convenience method with context support
func (c *FlixHQClient) GetStreamURLWithContext(ctx context.Context, media *FlixHQMedia, episode *FlixHQEpisode, provider, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
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

	streamInfo, err := c.ExtractStreamInfoWithContext(ctx, embedLink, quality, subsLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract stream info: %w", err)
	}

	return streamInfo, nil
}

// GetInfo fetches detailed info for a movie/show (matching greg interface)
func (c *FlixHQClient) GetInfo(id string) (*FlixHQMedia, error) {
	return c.GetInfoWithContext(context.Background(), id)
}

// GetInfoWithContext fetches detailed info for a movie/show with context support
func (c *FlixHQClient) GetInfoWithContext(ctx context.Context, id string) (*FlixHQMedia, error) {
	// Check cache first
	if cached, ok := c.infoCache.Load(id); ok {
		return cached.(*FlixHQMedia), nil
	}

	// Construct info URL
	infoURL := c.baseURL + "/" + id
	if strings.HasPrefix(id, "/") {
		infoURL = c.baseURL + id
	}

	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info request returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	info := &FlixHQMedia{
		ID:  id,
		URL: infoURL,
	}

	// Extract title
	info.Title = strings.TrimSpace(doc.Find(".heading-name a").First().Text())
	if info.Title == "" {
		info.Title = strings.TrimSpace(doc.Find("h2.heading-name").Text())
	}

	// Extract image
	if img, exists := doc.Find(".m_i-d-poster img").Attr("src"); exists {
		info.ImageURL = img
	}

	// Extract description
	info.Description = strings.TrimSpace(doc.Find(".description").Text())

	// Extract details from row-lines
	doc.Find(".row-line").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("strong, span.type").Text())
		label = strings.ToLower(label)

		switch {
		case strings.Contains(label, "released"):
			info.ReleaseDate = strings.TrimSpace(s.Find("a").First().Text())
			if info.ReleaseDate == "" {
				text := strings.TrimSpace(s.Text())
				text = strings.TrimPrefix(text, "Released:")
				info.ReleaseDate = strings.TrimSpace(text)
			}
			// Extract year from release date
			if len(info.ReleaseDate) >= 4 {
				info.Year = info.ReleaseDate[len(info.ReleaseDate)-4:]
			}
		case strings.Contains(label, "genre"):
			s.Find("a").Each(func(j int, genre *goquery.Selection) {
				genreText := strings.TrimSpace(genre.Text())
				if genreText != "" {
					info.Genres = append(info.Genres, genreText)
				}
			})
		case strings.Contains(label, "country"):
			info.Country = strings.TrimSpace(s.Find("a").First().Text())
		case strings.Contains(label, "production"):
			info.Production = strings.TrimSpace(s.Find("a").First().Text())
		case strings.Contains(label, "duration"):
			info.Duration = strings.TrimSpace(strings.TrimPrefix(s.Text(), "Duration:"))
		case strings.Contains(label, "cast"):
			s.Find("a").Each(func(j int, cast *goquery.Selection) {
				castName := strings.TrimSpace(cast.Text())
				if castName != "" {
					info.Casts = append(info.Casts, castName)
				}
			})
		}
	})

	// Determine if it's a movie or TV series
	if strings.Contains(strings.ToLower(info.Title), "season") ||
		doc.Find("#episodes-content").Length() > 0 ||
		doc.Find(".ss-list").Length() > 0 {
		info.Type = MediaTypeTV
	} else {
		info.Type = MediaTypeMovie
	}

	// Cache the result
	c.infoCache.Store(id, info)
	return info, nil
}

// GetServers fetches available servers for an episode or movie
func (c *FlixHQClient) GetServers(episodeID string, isMovie bool) ([]FlixHQServer, error) {
	return c.GetServersWithContext(context.Background(), episodeID, isMovie)
}

// GetServersWithContext fetches available servers with context support
func (c *FlixHQClient) GetServersWithContext(ctx context.Context, episodeID string, isMovie bool) ([]FlixHQServer, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s_%t", episodeID, isMovie)
	if cached, ok := c.serverCache.Load(cacheKey); ok {
		return cached.([]FlixHQServer), nil
	}

	var serverURL string
	if isMovie {
		serverURL = fmt.Sprintf("%s/ajax/movie/episodes/%s", c.baseURL, episodeID)
	} else {
		serverURL = fmt.Sprintf("%s/ajax/v2/episode/servers/%s", c.baseURL, episodeID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var servers []FlixHQServer
	var parseErr error

	if isMovie {
		servers, parseErr = c.parseMovieServers(string(body))
	} else {
		// Try JSON first for TV series
		var jsonResponse map[string]interface{}
		if err := json.Unmarshal(body, &jsonResponse); err == nil {
			if htmlContent, ok := jsonResponse["html"].(string); ok && htmlContent != "" {
				servers, parseErr = c.parseTVServers(htmlContent)
			} else {
				servers, parseErr = c.parseTVServers(string(body))
			}
		} else {
			servers, parseErr = c.parseTVServers(string(body))
		}
	}

	if parseErr != nil {
		return nil, parseErr
	}

	// Cache the servers
	c.serverCache.Store(cacheKey, servers)
	return servers, nil
}

// parseMovieServers parses server list from movie episodes HTML
func (c *FlixHQClient) parseMovieServers(htmlContent string) ([]FlixHQServer, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var servers []FlixHQServer

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		serverName, _ := s.Attr("title")
		href, exists := s.Attr("href")

		if exists && serverName != "" && href != "" {
			re := regexp.MustCompile(`\.(\d+)$`)
			if matches := re.FindStringSubmatch(href); len(matches) > 1 {
				servers = append(servers, FlixHQServer{
					Name: ServerName(serverName),
					ID:   matches[1],
					URL:  fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, matches[1]),
				})
			}
		}
	})

	return servers, nil
}

// parseTVServers parses server list from TV episode HTML
func (c *FlixHQClient) parseTVServers(htmlContent string) ([]FlixHQServer, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var servers []FlixHQServer

	doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
		serverName := strings.TrimSpace(s.Text())
		if serverName == "" {
			serverName, _ = s.Attr("title")
		}
		serverID, exists := s.Attr("data-id")

		if exists && serverName != "" {
			servers = append(servers, FlixHQServer{
				Name: ServerName(serverName),
				ID:   serverID,
				URL:  fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, serverID),
			})
		}
	})

	return servers, nil
}

// GetSources fetches video sources from all servers
func (c *FlixHQClient) GetSources(episodeID string, isMovie bool) (*FlixHQVideoSources, error) {
	return c.GetSourcesWithContext(context.Background(), episodeID, isMovie)
}

// GetSourcesWithContext fetches video sources with context support
func (c *FlixHQClient) GetSourcesWithContext(ctx context.Context, episodeID string, isMovie bool) (*FlixHQVideoSources, error) {
	servers, err := c.GetServersWithContext(ctx, episodeID, isMovie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	if len(servers) == 0 {
		return &FlixHQVideoSources{
			Sources:   []FlixHQSource{},
			Subtitles: []FlixHQSubtitle{},
		}, nil
	}

	// Sort servers by priority
	sortedServers := c.sortServersByPriority(servers)

	// Try each server until we get valid sources
	var lastErr error
	for _, server := range sortedServers {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sources, err := c.extractSourcesFromServer(ctx, server)
		if err != nil {
			lastErr = err
			util.Debug("Server failed", "server", server.Name, "error", err)
			continue
		}

		if len(sources.Sources) > 0 {
			return sources, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to extract sources from all servers: %w", lastErr)
	}

	return &FlixHQVideoSources{
		Sources:   []FlixHQSource{},
		Subtitles: []FlixHQSubtitle{},
	}, nil
}

// sortServersByPriority sorts servers based on DefaultServerPriority
func (c *FlixHQClient) sortServersByPriority(servers []FlixHQServer) []FlixHQServer {
	sorted := make([]FlixHQServer, len(servers))
	copy(sorted, servers)

	priorityMap := make(map[ServerName]int)
	for i, name := range DefaultServerPriority {
		priorityMap[name] = i
	}

	sort.Slice(sorted, func(i, j int) bool {
		pi, okI := priorityMap[sorted[i].Name]
		pj, okJ := priorityMap[sorted[j].Name]

		if !okI {
			pi = 100
		}
		if !okJ {
			pj = 100
		}
		return pi < pj
	})

	return sorted
}

// extractSourcesFromServer extracts video sources from a specific server
func (c *FlixHQClient) extractSourcesFromServer(ctx context.Context, server FlixHQServer) (*FlixHQVideoSources, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sources: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response to get embed URL
	var jsonResponse map[string]interface{}
	embedURL := ""

	if err := json.Unmarshal(body, &jsonResponse); err == nil {
		if link, ok := jsonResponse["link"].(string); ok {
			embedURL = link
		} else if link, ok := jsonResponse["embed"].(string); ok {
			embedURL = link
		} else if link, ok := jsonResponse["url"].(string); ok {
			embedURL = link
		}
	}

	if embedURL == "" {
		return nil, fmt.Errorf("no embed URL found in response")
	}

	// Extract sources from embed URL using the API
	return c.extractFromEmbedURL(ctx, embedURL)
}

// extractFromEmbedURL extracts video sources from an embed URL
func (c *FlixHQClient) extractFromEmbedURL(ctx context.Context, embedURL string) (*FlixHQVideoSources, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, url.QueryEscape(embedURL))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

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
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	videoSources := &FlixHQVideoSources{}

	// Parse sources
	if result.File != "" {
		videoSources.Sources = append(videoSources.Sources, FlixHQSource{
			URL:     result.File,
			Quality: "auto",
			IsM3U8:  strings.Contains(result.File, ".m3u8"),
			Referer: c.baseURL,
		})
	}

	for _, src := range result.Sources {
		videoSources.Sources = append(videoSources.Sources, FlixHQSource{
			URL:     src.File,
			Quality: src.Quality,
			IsM3U8:  strings.Contains(src.Type, "hls") || strings.Contains(src.File, ".m3u8"),
			Referer: c.baseURL,
		})
	}

	// Parse subtitles
	for _, track := range result.Tracks {
		if track.Kind == "captions" || track.Kind == "subtitles" {
			videoSources.Subtitles = append(videoSources.Subtitles, FlixHQSubtitle{
				URL:       track.File,
				Label:     track.Label,
				Language:  c.extractLanguageFromLabel(track.Label),
				IsDefault: track.Default,
			})
		}
	}

	return videoSources, nil
}

// GetAvailableQualities returns available video qualities for an episode
func (c *FlixHQClient) GetAvailableQualities(episodeID string, isMovie bool) ([]Quality, error) {
	return c.GetAvailableQualitiesWithContext(context.Background(), episodeID, isMovie)
}

// GetAvailableQualitiesWithContext returns available video qualities with context support
func (c *FlixHQClient) GetAvailableQualitiesWithContext(ctx context.Context, episodeID string, isMovie bool) ([]Quality, error) {
	sources, err := c.GetSourcesWithContext(ctx, episodeID, isMovie)
	if err != nil {
		return nil, err
	}

	qualitySet := make(map[Quality]bool)
	var qualities []Quality

	for _, src := range sources.Sources {
		q := parseQuality(src.Quality)
		if !qualitySet[q] {
			qualitySet[q] = true
			qualities = append(qualities, q)
		}
	}

	// Sort qualities from highest to lowest
	sort.Slice(qualities, func(i, j int) bool {
		return qualityToInt(qualities[i]) > qualityToInt(qualities[j])
	})

	return qualities, nil
}

// SelectBestQuality selects the best available quality from sources
func (c *FlixHQClient) SelectBestQuality(sources *FlixHQVideoSources, preferred Quality) *FlixHQSource {
	if len(sources.Sources) == 0 {
		return nil
	}

	// If preferred is auto or best, return the first (usually auto/best)
	if preferred == QualityAuto || preferred == QualityBest {
		for _, src := range sources.Sources {
			if src.Quality == "auto" || src.Quality == "" {
				return &src
			}
		}
		return &sources.Sources[0]
	}

	// Try to find exact match
	for _, src := range sources.Sources {
		if parseQuality(src.Quality) == preferred {
			return &src
		}
	}

	// Find closest quality
	preferredInt := qualityToInt(preferred)
	var best *FlixHQSource
	bestDiff := 10000

	for i := range sources.Sources {
		src := &sources.Sources[i]
		q := qualityToInt(parseQuality(src.Quality))
		diff := abs(q - preferredInt)
		if diff < bestDiff {
			bestDiff = diff
			best = src
		}
	}

	return best
}

// QualityOption represents a quality option for display
type QualityOption struct {
	Quality    Quality
	Label      string
	URL        string
	IsM3U8     bool
	Resolution string
}

// GetMovieQualities fetches available qualities for a movie
func (c *FlixHQClient) GetMovieQualities(mediaID string) ([]QualityOption, error) {
	return c.GetMovieQualitiesWithContext(context.Background(), mediaID)
}

// GetMovieQualitiesWithContext fetches available qualities for a movie with context support
func (c *FlixHQClient) GetMovieQualitiesWithContext(ctx context.Context, mediaID string) ([]QualityOption, error) {
	sources, err := c.GetSourcesWithContext(ctx, mediaID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie sources: %w", err)
	}

	return c.sourcesToQualityOptions(sources), nil
}

// GetMovieStreamWithQuality gets the stream URL for a movie with a specific quality
func (c *FlixHQClient) GetMovieStreamWithQuality(mediaID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return c.GetMovieStreamWithQualityContext(context.Background(), mediaID, quality, subsLanguage)
}

// GetMovieStreamWithQualityContext gets the stream URL for a movie with context support
func (c *FlixHQClient) GetMovieStreamWithQualityContext(ctx context.Context, mediaID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	sources, err := c.GetSourcesWithContext(ctx, mediaID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie sources: %w", err)
	}

	if len(sources.Sources) == 0 {
		return nil, errors.New("no video sources available for this movie")
	}

	// Select the appropriate source based on quality
	selectedSource := c.SelectBestQuality(sources, quality)
	if selectedSource == nil {
		return nil, errors.New("no matching quality found")
	}

	streamInfo := &FlixHQStreamInfo{
		VideoURL:  selectedSource.URL,
		Quality:   selectedSource.Quality,
		Referer:   c.baseURL,
		IsM3U8:    selectedSource.IsM3U8,
		Headers:   make(map[string]string),
		Qualities: c.sourcesToFlixHQQualityOptions(sources),
		Subtitles: sources.Subtitles,
	}
	streamInfo.Headers["Referer"] = c.baseURL

	if streamInfo.IsM3U8 {
		streamInfo.StreamType = StreamTypeHLS
	} else {
		streamInfo.StreamType = StreamTypeMP4
	}

	// Filter subtitles by language if specified
	if subsLanguage != "" && len(streamInfo.Subtitles) > 0 {
		streamInfo.Subtitles = c.filterSubtitlesByLanguage(streamInfo.Subtitles, subsLanguage)
	}

	return streamInfo, nil
}

// GetQualityLabels returns human-readable labels for available qualities
func (c *FlixHQClient) GetQualityLabels(qualities []Quality) []string {
	labels := make([]string, 0, len(qualities))
	for _, q := range qualities {
		labels = append(labels, c.QualityToLabel(q))
	}
	return labels
}

// QualityToLabel converts a Quality to a human-readable label
func (c *FlixHQClient) QualityToLabel(q Quality) string {
	switch q {
	case QualityAuto:
		return "Auto (Adaptive)"
	case Quality360:
		return "360p (SD)"
	case Quality480:
		return "480p (SD)"
	case Quality720:
		return "720p (HD)"
	case Quality1080:
		return "1080p (Full HD)"
	case QualityBest:
		return "Best Available"
	default:
		return string(q)
	}
}

// LabelToQuality converts a label back to Quality
func (c *FlixHQClient) LabelToQuality(label string) Quality {
	label = strings.ToLower(label)
	switch {
	case strings.Contains(label, "auto"):
		return QualityAuto
	case strings.Contains(label, "360"):
		return Quality360
	case strings.Contains(label, "480"):
		return Quality480
	case strings.Contains(label, "720"):
		return Quality720
	case strings.Contains(label, "1080"):
		return Quality1080
	case strings.Contains(label, "best"):
		return QualityBest
	default:
		return QualityAuto
	}
}

// sourcesToQualityOptions converts video sources to quality options for display
func (c *FlixHQClient) sourcesToQualityOptions(sources *FlixHQVideoSources) []QualityOption {
	options := make([]QualityOption, 0, len(sources.Sources))
	seen := make(map[Quality]bool)

	for _, src := range sources.Sources {
		q := parseQuality(src.Quality)
		if seen[q] {
			continue
		}
		seen[q] = true

		options = append(options, QualityOption{
			Quality:    q,
			Label:      c.QualityToLabel(q),
			URL:        src.URL,
			IsM3U8:     src.IsM3U8,
			Resolution: c.qualityToResolution(q),
		})
	}

	// Sort by quality (highest first)
	sort.Slice(options, func(i, j int) bool {
		return qualityToInt(options[i].Quality) > qualityToInt(options[j].Quality)
	})

	return options
}

// sourcesToFlixHQQualityOptions converts sources to FlixHQQualityOption slice
func (c *FlixHQClient) sourcesToFlixHQQualityOptions(sources *FlixHQVideoSources) []FlixHQQualityOption {
	options := make([]FlixHQQualityOption, 0, len(sources.Sources))
	for _, src := range sources.Sources {
		options = append(options, FlixHQQualityOption{
			Quality: parseQuality(src.Quality),
			URL:     src.URL,
			IsM3U8:  src.IsM3U8,
		})
	}
	return options
}

// qualityToResolution returns the resolution string for a quality
func (c *FlixHQClient) qualityToResolution(q Quality) string {
	switch q {
	case Quality360:
		return "640x360"
	case Quality480:
		return "854x480"
	case Quality720:
		return "1280x720"
	case Quality1080:
		return "1920x1080"
	default:
		return ""
	}
}

// filterSubtitlesByLanguage filters subtitles by language preference
func (c *FlixHQClient) filterSubtitlesByLanguage(subs []FlixHQSubtitle, language string) []FlixHQSubtitle {
	language = strings.ToLower(language)
	var filtered []FlixHQSubtitle

	for _, sub := range subs {
		if strings.Contains(strings.ToLower(sub.Language), language) ||
			strings.Contains(strings.ToLower(sub.Label), language) {
			filtered = append(filtered, sub)
		}
	}

	if len(filtered) > 0 {
		return filtered
	}
	return subs // Return original if no match
}

// SelectQualityInteractive returns quality options formatted for interactive selection
func (c *FlixHQClient) SelectQualityInteractive(mediaID string, isMovie bool) ([]string, []Quality, error) {
	return c.SelectQualityInteractiveWithContext(context.Background(), mediaID, isMovie)
}

// SelectQualityInteractiveWithContext returns quality options with context support
func (c *FlixHQClient) SelectQualityInteractiveWithContext(ctx context.Context, mediaID string, isMovie bool) ([]string, []Quality, error) {
	qualities, err := c.GetAvailableQualitiesWithContext(ctx, mediaID, isMovie)
	if err != nil {
		return nil, nil, err
	}

	if len(qualities) == 0 {
		// Return default options if none found
		defaultQualities := []Quality{QualityAuto, Quality720, Quality1080}
		return c.GetQualityLabels(defaultQualities), defaultQualities, nil
	}

	return c.GetQualityLabels(qualities), qualities, nil
}

// HealthCheck checks if FlixHQ is accessible
func (c *FlixHQClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ClearCache clears all caches
func (c *FlixHQClient) ClearCache() {
	c.searchCache = sync.Map{}
	c.infoCache = sync.Map{}
	c.serverCache = sync.Map{}
}

// ExtractStreamInfoWithContext extracts stream info with context support
func (c *FlixHQClient) ExtractStreamInfoWithContext(ctx context.Context, embedLink string, preferredQuality string, subsLanguage string) (*FlixHQStreamInfo, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, url.QueryEscape(embedLink))

	util.Debug("FlixHQ API request", "url", apiURL, "embed", embedLink)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	util.Debug("FlixHQ API response", "body", string(bodyBytes))

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
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(bodyBytes))
	}

	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}
	if result.Message != "" && result.File == "" && len(result.Sources) == 0 {
		return nil, fmt.Errorf("API message: %s", result.Message)
	}

	streamInfo := &FlixHQStreamInfo{
		Headers: make(map[string]string),
	}
	streamInfo.Headers["Referer"] = c.baseURL

	// Build quality options
	if result.File != "" {
		streamInfo.Qualities = append(streamInfo.Qualities, FlixHQQualityOption{
			Quality: QualityAuto,
			URL:     result.File,
			IsM3U8:  strings.Contains(result.File, ".m3u8"),
		})
	}
	for _, src := range result.Sources {
		streamInfo.Qualities = append(streamInfo.Qualities, FlixHQQualityOption{
			Quality: parseQuality(src.Quality),
			URL:     src.File,
			IsM3U8:  strings.Contains(src.Type, "hls") || strings.Contains(src.File, ".m3u8"),
		})
	}

	// Get video URL based on quality preference
	if result.File != "" {
		streamInfo.VideoURL = result.File
		streamInfo.IsM3U8 = strings.Contains(result.File, ".m3u8")
		util.Debug("FlixHQ got file", "url", result.File)

		if preferredQuality != "" && preferredQuality != "auto" && preferredQuality != "best" {
			streamInfo.VideoURL = strings.Replace(streamInfo.VideoURL, "/playlist.m3u8",
				fmt.Sprintf("/%s/index.m3u8", preferredQuality), 1)
		}
	} else if len(result.Sources) > 0 {
		util.Debug("FlixHQ got sources", "count", len(result.Sources))
		// Find preferred quality or use first
		for _, source := range result.Sources {
			if source.Quality == preferredQuality {
				streamInfo.VideoURL = source.File
				streamInfo.IsM3U8 = strings.Contains(source.Type, "hls") || strings.Contains(source.File, ".m3u8")
				break
			}
		}
		if streamInfo.VideoURL == "" {
			streamInfo.VideoURL = result.Sources[0].File
			streamInfo.IsM3U8 = strings.Contains(result.Sources[0].Type, "hls") || strings.Contains(result.Sources[0].File, ".m3u8")
		}
	}

	if streamInfo.VideoURL == "" {
		return nil, errors.New("no video URL found")
	}

	// Determine stream type
	if streamInfo.IsM3U8 {
		streamInfo.StreamType = StreamTypeHLS
	} else {
		streamInfo.StreamType = StreamTypeMP4
	}

	// Get subtitles
	for _, track := range result.Tracks {
		if track.Kind == "captions" || track.Kind == "subtitles" {
			sub := FlixHQSubtitle{
				URL:       track.File,
				Label:     track.Label,
				Language:  c.extractLanguageFromLabel(track.Label),
				IsDefault: track.Default,
			}
			streamInfo.Subtitles = append(streamInfo.Subtitles, sub)
		}
	}

	// Filter subtitles by preferred language
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

// Helper methods

// parseQuality converts a quality string to Quality type
func parseQuality(q string) Quality {
	q = strings.ToLower(strings.TrimSpace(q))
	switch {
	case q == "auto" || q == "":
		return QualityAuto
	case strings.Contains(q, "360"):
		return Quality360
	case strings.Contains(q, "480"):
		return Quality480
	case strings.Contains(q, "720"):
		return Quality720
	case strings.Contains(q, "1080"):
		return Quality1080
	case q == "best":
		return QualityBest
	default:
		return Quality(q)
	}
}

// qualityToInt converts Quality to int for comparison
func qualityToInt(q Quality) int {
	switch q {
	case Quality360:
		return 360
	case Quality480:
		return 480
	case Quality720:
		return 720
	case Quality1080:
		return 1080
	case QualityBest:
		return 9999
	default:
		return 0
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.ImageURL,
		Source:   "FlixHQ",
		Year:     m.Year,
		Quality:  m.Quality,
	}

	// Set media type
	if m.Type == MediaTypeMovie {
		anime.MediaType = models.MediaTypeMovie
	} else {
		anime.MediaType = models.MediaTypeTV
	}

	// Set additional fields if available
	if m.Description != "" {
		anime.Overview = m.Description
	}
	if len(m.Genres) > 0 {
		anime.Genres = m.Genres
	}
	if m.Rating > 0 {
		anime.Rating = m.Rating
	}

	return anime
}

// ToMedia converts FlixHQMedia to models.Media
func (m *FlixHQMedia) ToMedia() *models.Media {
	media := &models.Media{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.ImageURL,
		Source:   "FlixHQ",
		Year:     m.Year,
		Quality:  m.Quality,
		Overview: m.Description,
		Genres:   m.Genres,
		Rating:   m.Rating,
	}

	if m.Type == MediaTypeMovie {
		media.MediaType = models.MediaTypeMovie
	} else {
		media.MediaType = models.MediaTypeTV
	}

	return media
}

// ToEpisodeModel converts FlixHQEpisode to models.Episode for compatibility
func (e *FlixHQEpisode) ToEpisodeModel() models.Episode {
	return models.Episode{
		Number: fmt.Sprintf("%d", e.Number),
		Num:    e.Number,
		URL:    e.EpisodeURL,
		DataID: e.DataID,
		Title: models.TitleDetails{
			English: e.Title,
			Romaji:  e.Title,
		},
	}
}

// ToStreamInfo converts FlixHQStreamInfo to models.StreamInfo
func (s *FlixHQStreamInfo) ToStreamInfo() *models.StreamInfo {
	streamInfo := &models.StreamInfo{
		VideoURL:   s.VideoURL,
		Quality:    s.Quality,
		Referer:    s.Referer,
		SourceName: s.SourceName,
		Headers:    s.Headers,
	}

	for _, sub := range s.Subtitles {
		streamInfo.Subtitles = append(streamInfo.Subtitles, models.Subtitle{
			URL:      sub.URL,
			Language: sub.Language,
			Label:    sub.Label,
			IsForced: sub.IsForced,
		})
	}

	return streamInfo
}
