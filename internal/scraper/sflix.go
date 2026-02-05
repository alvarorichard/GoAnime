// Package scraper provides web scraping functionality for SFlix movies and TV shows
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	SFlixBase      = "https://sflix.ps"
	SFlixAPI       = "https://dec.eatmynerds.live" // Same API as FlixHQ for extraction
	SFlixUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

// SFlixClient handles interactions with SFlix
type SFlixClient struct {
	client      *http.Client
	baseURL     string
	apiURL      string
	userAgent   string
	maxRetries  int
	retryDelay  time.Duration
	searchCache sync.Map
	infoCache   sync.Map
	serverCache sync.Map
}

// SFlixMedia represents a movie or TV show from SFlix
type SFlixMedia struct {
	ID          string
	Title       string
	Type        MediaType
	Year        string
	ImageURL    string
	URL         string
	Seasons     []SFlixSeason
	Duration    string
	Quality     string
	Description string
	Genres      []string
	Rating      string
	ReleaseDate string
	Country     string
	Production  string
	Casts       []string
	Episodes    []SFlixEpisode // For storing episodes directly
}

// SFlixSeason represents a TV show season
type SFlixSeason struct {
	ID       string
	Number   int
	Title    string
	Episodes []SFlixEpisode
}

// SFlixEpisode represents a TV show episode
type SFlixEpisode struct {
	ID         string
	DataID     string
	Title      string
	Number     int
	Season     int
	SeasonID   string
	EpisodeURL string
	MediaID    string // Store the parent media ID for reference
}

// SFlixServer represents a streaming server
type SFlixServer struct {
	Name ServerName
	ID   string
	URL  string
}

// SFlixStreamInfo contains streaming information
type SFlixStreamInfo struct {
	VideoURL   string
	Quality    string
	Subtitles  []SFlixSubtitle
	Referer    string
	SourceName string
	StreamType StreamType
	IsM3U8     bool
	Headers    map[string]string
	Qualities  []SFlixQualityOption
}

// SFlixQualityOption represents an available quality option
type SFlixQualityOption struct {
	Quality Quality
	URL     string
	IsM3U8  bool
}

// SFlixSubtitle represents a subtitle track
type SFlixSubtitle struct {
	URL       string
	Language  string
	Label     string
	IsForced  bool
	IsDefault bool
}

// SFlixVideoSources contains parsed video sources
type SFlixVideoSources struct {
	Sources   []SFlixSource
	Subtitles []SFlixSubtitle
}

// SFlixSource represents a video source
type SFlixSource struct {
	URL     string
	Quality string
	IsM3U8  bool
	Referer string
}

// NewSFlixClient creates a new SFlix client
func NewSFlixClient() *SFlixClient {
	return &SFlixClient{
		client:     util.GetFastClient(),
		baseURL:    SFlixBase,
		apiURL:     SFlixAPI,
		userAgent:  SFlixUserAgent,
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

// NewSFlixClientWithContext creates a new SFlix client with custom settings
func NewSFlixClientWithContext(timeout time.Duration, maxRetries int) *SFlixClient {
	return &SFlixClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    SFlixBase,
		apiURL:     SFlixAPI,
		userAgent:  SFlixUserAgent,
		maxRetries: maxRetries,
		retryDelay: 300 * time.Millisecond,
	}
}

// SearchMedia searches for movies and TV shows on SFlix
func (c *SFlixClient) SearchMedia(query string) ([]*SFlixMedia, error) {
	return c.SearchMediaWithContext(context.Background(), query)
}

// SearchMediaWithContext searches for movies and TV shows on SFlix with context support
func (c *SFlixClient) SearchMediaWithContext(ctx context.Context, query string) ([]*SFlixMedia, error) {
	// Check cache first
	cacheKey := strings.ToLower(strings.TrimSpace(query))
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*SFlixMedia), nil
	}

	// SFlix uses dashes instead of spaces in search URLs
	searchQuery := strings.ReplaceAll(query, " ", "-")
	searchURL := fmt.Sprintf("%s/search/%s", c.baseURL, searchQuery)

	util.Debug("SFlix search", "query", query, "url", searchURL)

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
			lastErr = errors.New("SFlix returned a challenge page (try VPN or wait)")
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
	return nil, errors.New("failed to retrieve results from SFlix")
}

// GetTrending gets trending movies and TV shows
func (c *SFlixClient) GetTrending() ([]*SFlixMedia, error) {
	return c.getMediaFromPath("home")
}

// GetRecentMovies gets recent movies
func (c *SFlixClient) GetRecentMovies() ([]*SFlixMedia, error) {
	return c.getMediaFromPath("movie")
}

// GetRecentTV gets recent TV shows
func (c *SFlixClient) GetRecentTV() ([]*SFlixMedia, error) {
	return c.getMediaFromPath("tv-show")
}

func (c *SFlixClient) getMediaFromPath(path string) ([]*SFlixMedia, error) {
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
func (c *SFlixClient) GetSeasons(mediaID string) ([]SFlixSeason, error) {
	seasonsURL := fmt.Sprintf("%s/ajax/season/list/%s", c.baseURL, mediaID)

	req, err := http.NewRequest("GET", seasonsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var seasons []SFlixSeason
	seasonIdx := 0

	doc.Find(".ss-item").Each(func(i int, s *goquery.Selection) {
		dataID, exists := s.Attr("data-id")
		if !exists {
			return
		}

		seasonNumber := seasonIdx + 1
		seasonText := strings.TrimSpace(s.Text())

		// Try to extract season number from text like "Season 1"
		if strings.Contains(seasonText, "Season ") {
			parts := strings.Split(seasonText, "Season ")
			if len(parts) > 1 {
				if num, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					seasonNumber = num
				}
			}
		}

		seasons = append(seasons, SFlixSeason{
			ID:     dataID,
			Number: seasonNumber,
			Title:  seasonText,
		})
		seasonIdx++
	})

	return seasons, nil
}

// GetEpisodes gets all episodes for a season
func (c *SFlixClient) GetEpisodes(seasonID string) ([]SFlixEpisode, error) {
	episodesURL := fmt.Sprintf("%s/ajax/season/episodes/%s", c.baseURL, seasonID)

	req, err := http.NewRequest("GET", episodesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var episodes []SFlixEpisode
	epIdx := 0

	doc.Find(".eps-item").Each(func(i int, s *goquery.Selection) {
		dataID, exists := s.Attr("data-id")
		if !exists {
			return
		}

		epNumber := epIdx + 1
		epTitle := ""

		// Try to get episode number from the episode-number div
		epNumberText := s.Find(".episode-number").Text()
		epNumberText = strings.TrimSpace(strings.TrimPrefix(epNumberText, "Episode "))
		epNumberText = strings.TrimSuffix(epNumberText, ":")
		if parsedNum, err := strconv.Atoi(epNumberText); err == nil {
			epNumber = parsedNum
		}

		// Get episode title
		epTitle = strings.TrimSpace(s.Find(".film-name a").Text())
		if epTitle == "" {
			epTitle = strings.TrimSpace(s.Find("a").Text())
		}

		episodes = append(episodes, SFlixEpisode{
			ID:       dataID,
			DataID:   dataID,
			Title:    epTitle,
			Number:   epNumber,
			SeasonID: seasonID,
		})
		epIdx++
	})

	return episodes, nil
}

// GetInfo fetches detailed info for a movie/show
func (c *SFlixClient) GetInfo(id string) (*SFlixMedia, error) {
	return c.GetInfoWithContext(context.Background(), id)
}

// GetInfoWithContext fetches detailed info for a movie/show with context support
func (c *SFlixClient) GetInfoWithContext(ctx context.Context, id string) (*SFlixMedia, error) {
	// Check cache first
	if cached, ok := c.infoCache.Load(id); ok {
		return cached.(*SFlixMedia), nil
	}

	// Determine media type from id
	var mediaType MediaType
	var infoURL string

	if strings.HasPrefix(id, "movie/") {
		mediaType = MediaTypeMovie
		infoURL = fmt.Sprintf("%s/%s", c.baseURL, id)
	} else if strings.HasPrefix(id, "tv/") {
		mediaType = MediaTypeTV
		infoURL = fmt.Sprintf("%s/%s", c.baseURL, id)
	} else {
		// Try to fetch and detect type
		infoURL = fmt.Sprintf("%s/movie/%s", c.baseURL, id)
		mediaType = MediaTypeMovie
	}

	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Try TV show URL if movie fails
		if resp != nil {
			_ = resp.Body.Close()
		}
		infoURL = fmt.Sprintf("%s/tv/%s", c.baseURL, id)
		mediaType = MediaTypeTV

		req, err = http.NewRequestWithContext(ctx, "GET", infoURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err = c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch info: %w", err)
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info request returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract clean media ID
	cleanMediaID := id
	if !strings.Contains(id, "/") {
		cleanMediaID = string(mediaType) + "/" + id
	}

	info := &SFlixMedia{
		ID:       cleanMediaID,
		URL:      infoURL,
		Episodes: []SFlixEpisode{},
		Genres:   []string{},
		Type:     mediaType,
	}

	// Extract title
	info.Title = strings.TrimSpace(doc.Find("h2.heading-name").Text())
	if info.Title == "" {
		info.Title = strings.TrimSpace(doc.Find(".heading-name a").First().Text())
	}

	// Extract image
	if img, exists := doc.Find("img.film-poster-img").Attr("src"); exists {
		info.ImageURL = img
	}
	if info.ImageURL == "" {
		if img, exists := doc.Find(".m_i-d-poster img").Attr("src"); exists {
			info.ImageURL = img
		}
	}

	// Extract description
	info.Description = strings.TrimSpace(doc.Find("div.description").Text())

	// Extract rating from IMDB field
	ratingText := strings.TrimSpace(doc.Find("span.imdb").Text())
	if ratingText != "" {
		ratingText = strings.TrimPrefix(ratingText, "IMDB:")
		info.Rating = strings.TrimSpace(ratingText)
	}

	// Extract release date and genres from elements section
	doc.Find("div.elements .row-line").Each(func(i int, sel *goquery.Selection) {
		text := sel.Text()
		label := strings.ToLower(strings.TrimSpace(sel.Find("strong, span.type").Text()))

		switch {
		case strings.Contains(text, "Released:") || strings.Contains(label, "released"):
			parts := strings.Split(text, "Released:")
			if len(parts) > 1 {
				info.ReleaseDate = strings.TrimSpace(parts[1])
			}
			// Extract year from release date
			if len(info.ReleaseDate) >= 4 {
				// Try to find a 4-digit year
				re := regexp.MustCompile(`\d{4}`)
				if match := re.FindString(info.ReleaseDate); match != "" {
					info.Year = match
				}
			}

		case strings.Contains(text, "Genre:") || strings.Contains(label, "genre"):
			sel.Find("a").Each(func(j int, genreSel *goquery.Selection) {
				genre := strings.TrimSpace(genreSel.Text())
				if genre != "" && !strings.Contains(strings.ToLower(genre), "http") {
					info.Genres = append(info.Genres, genre)
				}
			})

		case strings.Contains(label, "country"):
			info.Country = strings.TrimSpace(sel.Find("a").First().Text())

		case strings.Contains(label, "production"):
			info.Production = strings.TrimSpace(sel.Find("a").First().Text())

		case strings.Contains(label, "duration"):
			info.Duration = strings.TrimSpace(strings.TrimPrefix(text, "Duration:"))

		case strings.Contains(label, "cast"):
			sel.Find("a").Each(func(j int, cast *goquery.Selection) {
				castName := strings.TrimSpace(cast.Text())
				if castName != "" {
					info.Casts = append(info.Casts, castName)
				}
			})
		}
	})

	// Extract data-id for fetching episodes
	dataID, exists := doc.Find(".detail_page-watch").Attr("data-id")
	if !exists {
		dataID, _ = doc.Find("#watch").Attr("data-id")
	}

	// For movies, create single episode with mediaID stored
	if mediaType == MediaTypeMovie && dataID != "" {
		info.Episodes = []SFlixEpisode{
			{
				ID:      dataID,
				DataID:  dataID,
				Number:  1,
				Title:   info.Title,
				MediaID: cleanMediaID,
			},
		}
	} else if mediaType == MediaTypeTV && dataID != "" {
		// For TV shows, fetch episode list
		episodes, err := c.fetchTVEpisodes(dataID, cleanMediaID)
		if err == nil && len(episodes) > 0 {
			info.Episodes = episodes

			// Build seasons from episodes
			seasonsMap := make(map[int][]SFlixEpisode)
			for _, ep := range episodes {
				season := ep.Season
				if season == 0 {
					season = 1
				}
				seasonsMap[season] = append(seasonsMap[season], ep)
			}

			for sNum, eps := range seasonsMap {
				info.Seasons = append(info.Seasons, SFlixSeason{
					ID:       fmt.Sprintf("%s|%d", dataID, sNum),
					Number:   sNum,
					Title:    fmt.Sprintf("Season %d", sNum),
					Episodes: eps,
				})
			}

			// Sort seasons
			sort.Slice(info.Seasons, func(i, j int) bool {
				return info.Seasons[i].Number < info.Seasons[j].Number
			})
		}
	}

	// Cache the result
	c.infoCache.Store(id, info)
	return info, nil
}

// fetchTVEpisodes fetches all episodes for a TV show
func (c *SFlixClient) fetchTVEpisodes(showID string, mediaID string) ([]SFlixEpisode, error) {
	// Step 1: Get all seasons
	seasonURL := fmt.Sprintf("%s/ajax/season/list/%s", c.baseURL, showID)

	req, err := http.NewRequest("GET", seasonURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create season list request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch season list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	seasonBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read season list response: %w", err)
	}

	seasonDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(seasonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse season HTML: %w", err)
	}

	var episodes []SFlixEpisode

	// Step 2: For each season, fetch its episodes
	seasonDoc.Find(".ss-item").Each(func(seasonIdx int, seasonSel *goquery.Selection) {
		seasonID, exists := seasonSel.Attr("data-id")
		if !exists {
			return
		}

		seasonNumber := seasonIdx + 1
		seasonText := strings.TrimSpace(seasonSel.Text())
		if strings.Contains(seasonText, "Season ") {
			parts := strings.Split(seasonText, "Season ")
			if len(parts) > 1 {
				if num, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					seasonNumber = num
				}
			}
		}

		// Fetch episodes for this season
		episodeURL := fmt.Sprintf("%s/ajax/season/episodes/%s", c.baseURL, seasonID)

		epReq, err := http.NewRequest("GET", episodeURL, nil)
		if err != nil {
			return
		}

		c.decorateRequest(epReq)
		epReq.Header.Set("X-Requested-With", "XMLHttpRequest")

		epResp, err := c.client.Do(epReq)
		if err != nil {
			return
		}
		defer func() { _ = epResp.Body.Close() }()

		epBody, err := io.ReadAll(epResp.Body)
		if err != nil {
			return
		}

		epDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(epBody)))
		if err != nil {
			return
		}

		// Find all episodes in this season
		epDoc.Find(".eps-item").Each(func(epIdx int, epSel *goquery.Selection) {
			epID, exists := epSel.Attr("data-id")
			if !exists {
				return
			}

			epNumber := epIdx + 1
			epTitle := ""

			// Try to get episode number from the episode-number div
			epNumberText := epSel.Find(".episode-number").Text()
			epNumberText = strings.TrimSpace(strings.TrimPrefix(epNumberText, "Episode "))
			epNumberText = strings.TrimSuffix(epNumberText, ":")
			if parsedNum, err := strconv.Atoi(epNumberText); err == nil {
				epNumber = parsedNum
			}

			// Get episode title
			epTitle = strings.TrimSpace(epSel.Find(".film-name a").Text())
			if epTitle == "" {
				epTitle = strings.TrimSpace(epSel.Find("a").Text())
			}

			episodes = append(episodes, SFlixEpisode{
				ID:       epID,
				DataID:   epID,
				Number:   epNumber,
				Season:   seasonNumber,
				Title:    epTitle,
				SeasonID: seasonID,
				MediaID:  mediaID,
			})
		})
	})

	return episodes, nil
}

// GetServers fetches available servers for an episode or movie
func (c *SFlixClient) GetServers(episodeID string, isMovie bool) ([]SFlixServer, error) {
	return c.GetServersWithContext(context.Background(), episodeID, isMovie)
}

// GetServersWithContext fetches available servers with context support
func (c *SFlixClient) GetServersWithContext(ctx context.Context, episodeID string, isMovie bool) ([]SFlixServer, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s_%t", episodeID, isMovie)
	if cached, ok := c.serverCache.Load(cacheKey); ok {
		return cached.([]SFlixServer), nil
	}

	// Check if episodeID contains mediaID (format: "id|mediaID")
	actualEpisodeID := episodeID
	var mediaID string
	if parts := strings.Split(episodeID, "|"); len(parts) == 2 {
		actualEpisodeID = parts[0]
		mediaID = parts[1]
	}

	// Determine endpoint based on whether it's a movie or TV show
	var endpoint string
	if isMovie || strings.Contains(mediaID, "movie") {
		endpoint = fmt.Sprintf("%s/ajax/episode/list/%s", c.baseURL, actualEpisodeID)
	} else {
		endpoint = fmt.Sprintf("%s/ajax/episode/servers/%s", c.baseURL, actualEpisodeID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
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

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var servers []SFlixServer

	// Find all server items - SFlix uses different selectors
	doc.Find(".ulclear > li, .nav-item").Each(func(i int, sel *goquery.Selection) {
		var dataID string
		var serverName string

		// Try getting data-id from anchor
		anchor := sel.Find("a")
		if anchor.Length() > 0 {
			dataID, _ = anchor.Attr("data-id")
			serverName = strings.TrimSpace(anchor.Find("span").Text())
			if serverName == "" {
				serverName = strings.TrimSpace(anchor.Text())
			}
		} else {
			dataID, _ = sel.Attr("data-id")
			serverName = strings.TrimSpace(sel.Text())
		}

		if dataID == "" || serverName == "" {
			return
		}

		servers = append(servers, SFlixServer{
			Name: ServerName(strings.ToLower(serverName)),
			ID:   dataID,
			URL:  fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, dataID),
		})
	})

	// Cache the servers
	c.serverCache.Store(cacheKey, servers)
	return servers, nil
}

// GetSources fetches video sources from all servers
func (c *SFlixClient) GetSources(episodeID string, isMovie bool) (*SFlixVideoSources, error) {
	return c.GetSourcesWithContext(context.Background(), episodeID, isMovie)
}

// GetSourcesWithContext fetches video sources with context support
func (c *SFlixClient) GetSourcesWithContext(ctx context.Context, episodeID string, isMovie bool) (*SFlixVideoSources, error) {
	servers, err := c.GetServersWithContext(ctx, episodeID, isMovie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	if len(servers) == 0 {
		return &SFlixVideoSources{
			Sources:   []SFlixSource{},
			Subtitles: []SFlixSubtitle{},
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
			util.Debug("SFlix server failed", "server", server.Name, "error", err)
			continue
		}

		if len(sources.Sources) > 0 {
			return sources, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to extract sources from all servers: %w", lastErr)
	}

	return &SFlixVideoSources{
		Sources:   []SFlixSource{},
		Subtitles: []SFlixSubtitle{},
	}, nil
}

// sortServersByPriority sorts servers based on DefaultServerPriority
func (c *SFlixClient) sortServersByPriority(servers []SFlixServer) []SFlixServer {
	sorted := make([]SFlixServer, len(servers))
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
func (c *SFlixClient) extractSourcesFromServer(ctx context.Context, server SFlixServer) (*SFlixVideoSources, error) {
	// Use /ajax/episode/sources/{serverID} endpoint
	sourcesURL := server.URL
	if sourcesURL == "" {
		sourcesURL = fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, server.ID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", sourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch embed URL: %w", err)
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
	var jsonResponse struct {
		Link string `json:"link"`
	}

	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if jsonResponse.Link == "" {
		return nil, fmt.Errorf("no embed link found in response")
	}

	// Extract sources from embed URL using the API
	return c.extractFromEmbedURL(ctx, jsonResponse.Link)
}

// extractFromEmbedURL extracts video sources from an embed URL
func (c *SFlixClient) extractFromEmbedURL(ctx context.Context, embedURL string) (*SFlixVideoSources, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, embedURL)

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

	videoSources := &SFlixVideoSources{}

	// Parse sources
	if result.File != "" {
		videoSources.Sources = append(videoSources.Sources, SFlixSource{
			URL:     result.File,
			Quality: "auto",
			IsM3U8:  strings.Contains(result.File, ".m3u8"),
			Referer: c.baseURL,
		})
	}

	for _, src := range result.Sources {
		videoSources.Sources = append(videoSources.Sources, SFlixSource{
			URL:     src.File,
			Quality: src.Quality,
			IsM3U8:  strings.Contains(src.Type, "hls") || strings.Contains(src.File, ".m3u8"),
			Referer: c.baseURL,
		})
	}

	// Parse subtitles
	for _, track := range result.Tracks {
		if track.Kind == "captions" || track.Kind == "subtitles" {
			videoSources.Subtitles = append(videoSources.Subtitles, SFlixSubtitle{
				URL:       track.File,
				Label:     track.Label,
				Language:  c.extractLanguageFromLabel(track.Label),
				IsDefault: track.Default,
			})
		}
	}

	return videoSources, nil
}

// GetEmbedLink gets the embed link for streaming
func (c *SFlixClient) GetEmbedLink(episodeID string) (string, error) {
	sourcesURL := fmt.Sprintf("%s/ajax/episode/sources/%s", c.baseURL, episodeID)

	util.Debug("SFlix get embed link", "url", sourcesURL, "episodeID", episodeID)

	req, err := http.NewRequest("GET", sourcesURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	util.Debug("SFlix embed response", "body", string(bodyBytes))

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

	util.Debug("SFlix embed link found", "link", result.Link)
	return result.Link, nil
}

// GetMovieServerID gets the server ID for a movie
func (c *SFlixClient) GetMovieServerID(mediaID, provider string) (string, error) {
	movieURL := fmt.Sprintf("%s/ajax/episode/list/%s", c.baseURL, mediaID)

	util.Debug("SFlix get movie servers", "url", movieURL, "provider", provider)

	req, err := http.NewRequest("GET", movieURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Try multiple providers in order of preference
	providers := []string{provider, "Vidcloud", "UpCloud", "Megacloud", "Voe", "MixDrop"}
	var episodeID string

	for _, prov := range providers {
		doc.Find(".ulclear > li a, a").Each(func(i int, s *goquery.Selection) {
			if episodeID != "" {
				return // Already found
			}
			serverName := strings.TrimSpace(s.Find("span").Text())
			if serverName == "" {
				serverName = strings.TrimSpace(s.Text())
			}

			if strings.EqualFold(serverName, prov) {
				id, exists := s.Attr("data-id")
				if exists && id != "" {
					episodeID = id
					util.Debug("SFlix found server", "provider", serverName, "id", id)
				}
			}
		})
		if episodeID != "" {
			break
		}
	}

	if episodeID == "" {
		// Fallback to first available server
		doc.Find(".ulclear > li a, a").First().Each(func(i int, s *goquery.Selection) {
			id, exists := s.Attr("data-id")
			if exists && id != "" {
				episodeID = id
				serverName := strings.TrimSpace(s.Find("span").Text())
				util.Debug("SFlix using fallback server", "provider", serverName, "id", id)
			}
		})
	}

	if episodeID == "" {
		return "", errors.New("no server found for movie")
	}

	return episodeID, nil
}

// GetEpisodeServerID gets the server ID for an episode
func (c *SFlixClient) GetEpisodeServerID(dataID, provider string) (string, error) {
	serversURL := fmt.Sprintf("%s/ajax/episode/servers/%s", c.baseURL, dataID)

	util.Debug("SFlix get episode servers", "url", serversURL, "provider", provider)

	req, err := http.NewRequest("GET", serversURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Try multiple providers in order of preference
	providers := []string{provider, "Vidcloud", "UpCloud", "Megacloud", "Voe", "MixDrop"}
	var episodeID string

	for _, prov := range providers {
		doc.Find(".ulclear > li a, .nav-item a, a").Each(func(i int, s *goquery.Selection) {
			if episodeID != "" {
				return
			}
			serverName := strings.TrimSpace(s.Find("span").Text())
			if serverName == "" {
				serverName, _ = s.Attr("title")
			}
			if serverName == "" {
				serverName = strings.TrimSpace(s.Text())
			}

			if strings.EqualFold(serverName, prov) {
				id, exists := s.Attr("data-id")
				if exists && id != "" {
					episodeID = id
					util.Debug("SFlix found server", "provider", serverName, "id", id)
				}
			}
		})
		if episodeID != "" {
			break
		}
	}

	if episodeID == "" {
		// Fallback to first available server
		doc.Find(".ulclear > li a, .nav-item a, a").First().Each(func(i int, s *goquery.Selection) {
			id, exists := s.Attr("data-id")
			if exists && id != "" {
				episodeID = id
				serverName := strings.TrimSpace(s.Find("span").Text())
				util.Debug("SFlix using fallback server", "provider", serverName, "id", id)
			}
		})
	}

	if episodeID == "" {
		return "", errors.New("no server found for episode")
	}

	return episodeID, nil
}

// ExtractStreamInfo extracts video URL and subtitles from embed link
func (c *SFlixClient) ExtractStreamInfo(embedLink string, preferredQuality string, subsLanguage string) (*SFlixStreamInfo, error) {
	return c.ExtractStreamInfoWithContext(context.Background(), embedLink, preferredQuality, subsLanguage)
}

// ExtractStreamInfoWithContext extracts stream info with context support
func (c *SFlixClient) ExtractStreamInfoWithContext(ctx context.Context, embedLink string, preferredQuality string, subsLanguage string) (*SFlixStreamInfo, error) {
	apiURL := fmt.Sprintf("%s/?url=%s", c.apiURL, embedLink)

	util.Debug("SFlix API request", "url", apiURL, "embed", embedLink)

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

	util.Debug("SFlix API response", "body", string(bodyBytes))

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

	streamInfo := &SFlixStreamInfo{
		Headers: make(map[string]string),
	}
	streamInfo.Headers["Referer"] = c.baseURL

	// Build quality options
	if result.File != "" {
		streamInfo.Qualities = append(streamInfo.Qualities, SFlixQualityOption{
			Quality: QualityAuto,
			URL:     result.File,
			IsM3U8:  strings.Contains(result.File, ".m3u8"),
		})
	}
	for _, src := range result.Sources {
		streamInfo.Qualities = append(streamInfo.Qualities, SFlixQualityOption{
			Quality: parseQuality(src.Quality),
			URL:     src.File,
			IsM3U8:  strings.Contains(src.Type, "hls") || strings.Contains(src.File, ".m3u8"),
		})
	}

	// Get video URL based on quality preference
	if result.File != "" {
		streamInfo.VideoURL = result.File
		streamInfo.IsM3U8 = strings.Contains(result.File, ".m3u8")

		if preferredQuality != "" && preferredQuality != "auto" && preferredQuality != "best" {
			streamInfo.VideoURL = strings.Replace(streamInfo.VideoURL, "/playlist.m3u8",
				fmt.Sprintf("/%s/index.m3u8", preferredQuality), 1)
		}
	} else if len(result.Sources) > 0 {
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
			sub := SFlixSubtitle{
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
		var filteredSubs []SFlixSubtitle
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

	// Extract referer from the embed link - this is critical for avoiding 403 errors
	// The streaming server expects the Referer to be the origin of the embed URL
	// e.g., https://videostr.net/embed-1/abc123 -> https://videostr.net/
	if parsedURL, parseErr := url.Parse(embedLink); parseErr == nil && parsedURL.Host != "" {
		streamInfo.Referer = fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host)
		streamInfo.Headers["Referer"] = streamInfo.Referer
		streamInfo.Headers["Origin"] = strings.TrimSuffix(streamInfo.Referer, "/")
		util.Debug("SFlix set stream referer", "referer", streamInfo.Referer, "embedLink", embedLink)
	}

	return streamInfo, nil
}

// GetStreamURL is a convenience method that combines all steps to get a stream URL
func (c *SFlixClient) GetStreamURL(media *SFlixMedia, episode *SFlixEpisode, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	return c.GetStreamURLWithContext(context.Background(), media, episode, provider, quality, subsLanguage)
}

// GetStreamURLWithContext is a convenience method with context support
func (c *SFlixClient) GetStreamURLWithContext(ctx context.Context, media *SFlixMedia, episode *SFlixEpisode, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
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

// GetAvailableQualities returns available video qualities for an episode
func (c *SFlixClient) GetAvailableQualities(episodeID string, isMovie bool) ([]Quality, error) {
	return c.GetAvailableQualitiesWithContext(context.Background(), episodeID, isMovie)
}

// GetAvailableQualitiesWithContext returns available video qualities with context support
func (c *SFlixClient) GetAvailableQualitiesWithContext(ctx context.Context, episodeID string, isMovie bool) ([]Quality, error) {
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
func (c *SFlixClient) SelectBestQuality(sources *SFlixVideoSources, preferred Quality) *SFlixSource {
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
	var best *SFlixSource
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

// HealthCheck checks if SFlix is accessible
func (c *SFlixClient) HealthCheck(ctx context.Context) error {
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
func (c *SFlixClient) ClearCache() {
	c.searchCache = sync.Map{}
	c.infoCache = sync.Map{}
	c.serverCache = sync.Map{}
}

// Helper methods

func (c *SFlixClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *SFlixClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *SFlixClient) sleep() {
	if c.retryDelay > 0 {
		time.Sleep(c.retryDelay)
	}
}

func (c *SFlixClient) isChallengePage(doc *goquery.Document) bool {
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

func (c *SFlixClient) extractSearchResults(doc *goquery.Document) []*SFlixMedia {
	var media []*SFlixMedia

	doc.Find(".flw-item, div.flw-item").Each(func(i int, s *goquery.Selection) {
		if m := c.parseMediaItem(s); m != nil {
			media = append(media, m)
		}
	})

	return media
}

func (c *SFlixClient) parseMediaItem(s *goquery.Selection) *SFlixMedia {
	// Get link and extract media info
	linkElem := s.Find("h2.film-name a, .film-name a")
	href, exists := linkElem.Attr("href")
	if !exists {
		return nil
	}

	// Get title from the title attribute first (cleaner), then fallback to text
	title, _ := linkElem.Attr("title")
	if title == "" {
		// Clone the element and remove any nested elements that might contain "Watch now"
		title = strings.TrimSpace(linkElem.Text())
	}

	// Clean up title - remove "Watch now" and similar patterns using regex
	watchNowRegex := regexp.MustCompile(`(?i)\s*Watch\s*now\s*$`)
	title = watchNowRegex.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)

	if title == "" {
		return nil
	}

	// Get image URL
	imgElem := s.Find("img")
	imageURL, _ := imgElem.Attr("data-src")
	if imageURL == "" {
		imageURL, _ = imgElem.Attr("src")
	}

	// Determine media type and ID from URL
	var mediaType MediaType
	var mediaID string

	// Extract full path ID (e.g., "movie/free-inception-hd-19764" or "tv/free-stranger-things-hd-39444")
	parts := strings.Split(strings.TrimPrefix(href, "/"), "/")
	if len(parts) >= 2 {
		switch parts[0] {
		case "tv", "tv-show":
			mediaType = MediaTypeTV
		case "movie":
			mediaType = MediaTypeMovie
		default:
			// Try to detect type from other indicators
			mediaType = c.detectMediaType(s, href)
		}
		mediaID = parts[0] + "/" + parts[1]
	} else if len(parts) == 1 {
		mediaID = parts[0]
		// Try to detect type from other indicators
		mediaType = c.detectMediaType(s, href)
	} else {
		return nil
	}

	// If we still couldn't determine the type, skip this item
	if mediaID == "" {
		return nil
	}

	// Get year/info from fdi-item
	year := ""
	s.Find(".fdi-item").Each(func(i int, item *goquery.Selection) {
		text := strings.TrimSpace(item.Text())
		// Check if it looks like a year (4 digits)
		if len(text) == 4 {
			if _, err := strconv.Atoi(text); err == nil {
				year = text
			}
		}
	})

	return &SFlixMedia{
		ID:       mediaID,
		Title:    title,
		Type:     mediaType,
		Year:     year,
		ImageURL: imageURL,
		URL:      c.resolveURL(href),
	}
}

// detectMediaType tries to detect media type from various HTML indicators
func (c *SFlixClient) detectMediaType(s *goquery.Selection, href string) MediaType {
	// Check for type badge/indicator in the item
	// SFlix often has spans or elements with "TV" or "Movie" text
	fdType := strings.ToLower(s.Find(".fd-type, .fdi-type, .type").Text())
	if strings.Contains(fdType, "tv") || strings.Contains(fdType, "series") {
		return MediaTypeTV
	}
	if strings.Contains(fdType, "movie") {
		return MediaTypeMovie
	}

	// Check for episode count indicator (TV shows have "Eps" badge)
	if s.Find(".tick-eps, .tick-sub").Length() > 0 {
		return MediaTypeTV
	}

	// Check the fdi-item for indicators
	var hasEps bool
	s.Find(".fdi-item").Each(func(i int, item *goquery.Selection) {
		text := strings.ToLower(strings.TrimSpace(item.Text()))
		if strings.Contains(text, "eps") || strings.Contains(text, "ss") {
			hasEps = true
		}
	})
	if hasEps {
		return MediaTypeTV
	}

	// Check URL patterns
	hrefLower := strings.ToLower(href)
	if strings.Contains(hrefLower, "/tv/") || strings.Contains(hrefLower, "tv-show") {
		return MediaTypeTV
	}
	if strings.Contains(hrefLower, "/movie/") {
		return MediaTypeMovie
	}

	// Check for quality badge (movies typically show quality like "HD")
	qualityBadge := s.Find(".quality, .fdi-item strong").Text()
	if qualityBadge != "" && !strings.Contains(strings.ToLower(qualityBadge), "eps") {
		return MediaTypeMovie
	}

	// Default to movie if we can't determine
	return MediaTypeMovie
}

func (c *SFlixClient) resolveURL(ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	if strings.HasPrefix(ref, "/") {
		return c.baseURL + ref
	}
	return c.baseURL + "/" + ref
}

func (c *SFlixClient) extractLanguageFromLabel(label string) string {
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

// QualityToLabel converts a Quality to a human-readable label
func (c *SFlixClient) QualityToLabel(q Quality) string {
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

// ToAnimeModel converts SFlixMedia to models.Anime for compatibility
func (m *SFlixMedia) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.ImageURL,
		Source:   "SFlix",
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

	return anime
}

// ToMedia converts SFlixMedia to models.Media
func (m *SFlixMedia) ToMedia() *models.Media {
	media := &models.Media{
		Name:     m.Title,
		URL:      m.URL,
		ImageURL: m.ImageURL,
		Source:   "SFlix",
		Year:     m.Year,
		Quality:  m.Quality,
		Overview: m.Description,
		Genres:   m.Genres,
	}

	if m.Type == MediaTypeMovie {
		media.MediaType = models.MediaTypeMovie
	} else {
		media.MediaType = models.MediaTypeTV
	}

	return media
}

// ToEpisodeModel converts SFlixEpisode to models.Episode for compatibility
func (e *SFlixEpisode) ToEpisodeModel() models.Episode {
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

// ToStreamInfo converts SFlixStreamInfo to models.StreamInfo
func (s *SFlixStreamInfo) ToStreamInfo() *models.StreamInfo {
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
