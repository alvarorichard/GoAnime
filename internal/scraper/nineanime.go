// Package scraper provides web scraping functionality for 9animetv.to
package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	NineAnimeBase      = "https://9animetv.to"
	NineAnimeUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// NineAnimeClient handles interactions with 9animetv.to
type NineAnimeClient struct {
	client      *http.Client
	baseURL     string
	userAgent   string
	maxRetries  int
	retryDelay  time.Duration
	searchCache sync.Map
	serverCache sync.Map
}

// NineAnimeResult represents a single search result from 9anime
type NineAnimeResult struct {
	Title   string
	URL     string // e.g. /watch/naruto-677
	AnimeID string // e.g. 677
	Extra   string // e.g. "SUB DUB  Ep 220/220"
	Image   string // poster/thumbnail URL
}

// NineAnimeEpisode represents a single episode entry
type NineAnimeEpisode struct {
	Number    int
	Title     string
	EpisodeID string // data-id attribute used in AJAX calls
	URL       string // relative URL with ?ep= param
}

// NineAnimeServer represents a streaming server option
type NineAnimeServer struct {
	Name      string // e.g. Vidstreaming, Vidcloud, DouVideo
	DataID    string // the data-id used in /ajax/episode/sources
	ServerID  string // data-server-id
	AudioType string // "sub" or "dub"
}

// NineAnimeStreamSource represents resolved stream information
type NineAnimeStreamSource struct {
	EmbedURL   string
	ServerType int
}

// NineAnimeSubtitleTrack represents a subtitle/caption track
type NineAnimeSubtitleTrack struct {
	Label   string // e.g. "English", "Portuguese - Brazilian Portuguese"
	File    string // VTT URL
	Kind    string // "captions" or "subtitles"
	Default bool
}

// NineAnimeStreamInfo contains full resolved stream: video URL + subtitles + skip info
type NineAnimeStreamInfo struct {
	M3U8URL    string
	Tracks     []NineAnimeSubtitleTrack
	IntroStart int
	IntroEnd   int
	OutroStart int
	OutroEnd   int
	EmbedURL   string
}

// NewNineAnimeClient creates a new 9anime client (singleton for connection reuse)
var (
	nineAnimeClientInstance *NineAnimeClient
	nineAnimeClientOnce     sync.Once
)

func NewNineAnimeClient() *NineAnimeClient {
	nineAnimeClientOnce.Do(func() {
		nineAnimeClientInstance = &NineAnimeClient{
			client:     util.GetFastClient(),
			baseURL:    NineAnimeBase,
			userAgent:  NineAnimeUserAgent,
			maxRetries: 2,
			retryDelay: 300 * time.Millisecond,
		}
	})
	return nineAnimeClientInstance
}

// decorateRequest adds standard headers to a request
func (c *NineAnimeClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
}

// decorateAJAXRequest adds headers specific to AJAX/JSON requests
func (c *NineAnimeClient) decorateAJAXRequest(req *http.Request) {
	c.decorateRequest(req)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
}

func (c *NineAnimeClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *NineAnimeClient) sleep() {
	if c.retryDelay > 0 {
		time.Sleep(c.retryDelay)
	}
}

func (c *NineAnimeClient) isChallengePage(doc *goquery.Document) bool {
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

// ─── Search ──────────────────────────────────────────────────────────────────

// SearchAnime searches for anime by keyword on 9animetv.to
func (c *NineAnimeClient) SearchAnime(query string) ([]*models.Anime, error) {
	return c.SearchAnimeWithContext(context.Background(), query)
}

// SearchAnimeWithContext searches with context support
func (c *NineAnimeClient) SearchAnimeWithContext(ctx context.Context, query string) ([]*models.Anime, error) {
	// Check cache
	cacheKey := strings.ToLower(strings.TrimSpace(query))
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*models.Anime), nil
	}

	searchURL := fmt.Sprintf("%s/search?keyword=%s", c.baseURL, strings.ReplaceAll(query, " ", "+"))
	util.Debug("9anime search", "query", query, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
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

		resp, err := c.client.Do(req) // #nosec G704 -- URL built from known base domain
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
			lastErr = errors.New("9anime returned a challenge page (try VPN or wait)")
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		results := c.extractSearchResults(doc)
		animes := make([]*models.Anime, 0, len(results))
		for _, r := range results {
			anime := &models.Anime{
				Name:      r.Title,
				URL:       r.AnimeID, // Store anime ID for subsequent episode listing
				ImageURL:  "",
				Source:    "9Anime",
				MediaType: models.MediaTypeAnime,
			}
			if r.Extra != "" {
				anime.Name = fmt.Sprintf("%s (%s)", r.Title, r.Extra)
			}
			animes = append(animes, anime)
		}

		c.searchCache.Store(cacheKey, animes)
		return animes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve results from 9anime")
}

// extractSearchResults parses search result HTML
func (c *NineAnimeClient) extractSearchResults(doc *goquery.Document) []NineAnimeResult {
	var results []NineAnimeResult
	animeIDRegex := regexp.MustCompile(`-(\d+)$`)

	doc.Find(".film_list-wrap .flw-item").Each(func(i int, item *goquery.Selection) {
		linkTag := item.Find("h3.film-name a, .film-name a").First()
		if linkTag.Length() == 0 {
			return
		}

		title := strings.TrimSpace(linkTag.Text())
		href, _ := linkTag.Attr("href")

		animeID := ""
		if matches := animeIDRegex.FindStringSubmatch(href); len(matches) > 1 {
			animeID = matches[1]
		}

		// Grab extra info (episode count, sub/dub badges)
		var extraParts []string
		item.Find(".tick-item, .tick-sub, .tick-dub, .tick-eps").Each(func(j int, badge *goquery.Selection) {
			text := strings.TrimSpace(badge.Text())
			if text != "" {
				extraParts = append(extraParts, text)
			}
		})

		// Try to get image URL
		imgURL := ""
		img := item.Find("img").First()
		if img.Length() > 0 {
			imgURL, _ = img.Attr("data-src")
			if imgURL == "" {
				imgURL, _ = img.Attr("src")
			}
		}

		if title != "" && animeID != "" {
			results = append(results, NineAnimeResult{
				Title:   title,
				URL:     href,
				AnimeID: animeID,
				Extra:   strings.Join(extraParts, " "),
				Image:   imgURL,
			})
		}
	})

	return results
}

// ─── Episodes ────────────────────────────────────────────────────────────────

// GetEpisodes fetches the full episode list for an anime via the AJAX endpoint
func (c *NineAnimeClient) GetEpisodes(animeID string) ([]NineAnimeEpisode, error) {
	return c.GetEpisodesWithContext(context.Background(), animeID)
}

// GetEpisodesWithContext fetches episodes with context support
func (c *NineAnimeClient) GetEpisodesWithContext(ctx context.Context, animeID string) ([]NineAnimeEpisode, error) {
	url := fmt.Sprintf("%s/ajax/episode/list/%s", c.baseURL, animeID)
	util.Debug("9anime get episodes", "animeID", animeID, "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateAJAXRequest(req)

	resp, err := c.client.Do(req) // #nosec G704 -- URL built from known base domain
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ajaxResp struct {
		Status bool   `json:"status"`
		HTML   string `json:"html"`
	}
	if err := json.Unmarshal(body, &ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if ajaxResp.HTML == "" {
		return nil, errors.New("empty episode list HTML")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ajaxResp.HTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse episode HTML: %w", err)
	}

	var episodes []NineAnimeEpisode
	doc.Find("a.ep-item").Each(func(i int, a *goquery.Selection) {
		epNumberStr, _ := a.Attr("data-number")
		epID, _ := a.Attr("data-id")
		epTitle, _ := a.Attr("title")
		epURL, _ := a.Attr("href")

		epNumber, _ := strconv.Atoi(epNumberStr)
		if epTitle == "" {
			epTitle = fmt.Sprintf("Episode %d", epNumber)
		}

		if epID != "" {
			episodes = append(episodes, NineAnimeEpisode{
				Number:    epNumber,
				Title:     epTitle,
				EpisodeID: epID,
				URL:       epURL,
			})
		}
	})

	util.Debug("9anime episodes found", "animeID", animeID, "count", len(episodes))
	return episodes, nil
}

// GetAnimeEpisodes returns episodes in the models.Episode format
func (c *NineAnimeClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// animeURL is actually the anime ID for 9anime
	episodes, err := c.GetEpisodes(animeURL)
	if err != nil {
		return nil, err
	}

	var modelEpisodes []models.Episode
	for _, ep := range episodes {
		modelEpisodes = append(modelEpisodes, models.Episode{
			Number: strconv.Itoa(ep.Number),
			Num:    ep.Number,
			URL:    ep.EpisodeID, // Store episode ID for server resolution
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			DataID: ep.EpisodeID,
		})
	}

	return modelEpisodes, nil
}

// ─── Servers ─────────────────────────────────────────────────────────────────

// GetServers fetches available streaming servers for an episode
func (c *NineAnimeClient) GetServers(episodeID string) ([]NineAnimeServer, error) {
	return c.GetServersWithContext(context.Background(), episodeID)
}

// GetServersWithContext fetches servers with context support
func (c *NineAnimeClient) GetServersWithContext(ctx context.Context, episodeID string) ([]NineAnimeServer, error) {
	// Check cache
	cacheKey := "servers:" + episodeID
	if cached, ok := c.serverCache.Load(cacheKey); ok {
		return cached.([]NineAnimeServer), nil
	}

	url := fmt.Sprintf("%s/ajax/episode/servers?episodeId=%s", c.baseURL, episodeID)
	util.Debug("9anime get servers", "episodeID", episodeID, "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateAJAXRequest(req)

	resp, err := c.client.Do(req) // #nosec G704 -- URL built from known base domain
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ajaxResp struct {
		Status bool   `json:"status"`
		HTML   string `json:"html"`
	}
	if err := json.Unmarshal(body, &ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if ajaxResp.HTML == "" {
		return nil, errors.New("empty server list HTML")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ajaxResp.HTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse server HTML: %w", err)
	}

	var servers []NineAnimeServer
	doc.Find(".server-item").Each(func(i int, item *goquery.Selection) {
		name := strings.TrimSpace(item.Text())
		dataID, _ := item.Attr("data-id")
		serverID, _ := item.Attr("data-server-id")
		audioType, _ := item.Attr("data-type")
		if audioType == "" {
			audioType = "sub"
		}

		if dataID != "" && name != "" {
			servers = append(servers, NineAnimeServer{
				Name:      name,
				DataID:    dataID,
				ServerID:  serverID,
				AudioType: audioType,
			})
		}
	})

	util.Debug("9anime servers found", "episodeID", episodeID, "count", len(servers))
	c.serverCache.Store(cacheKey, servers)
	return servers, nil
}

// ─── Source / Stream Resolution ──────────────────────────────────────────────

// GetSource resolves the embed/iframe URL for a given server data-id
func (c *NineAnimeClient) GetSource(serverDataID string) (*NineAnimeStreamSource, error) {
	return c.GetSourceWithContext(context.Background(), serverDataID)
}

// GetSourceWithContext resolves the embed URL with context support
func (c *NineAnimeClient) GetSourceWithContext(ctx context.Context, serverDataID string) (*NineAnimeStreamSource, error) {
	url := fmt.Sprintf("%s/ajax/episode/sources?id=%s", c.baseURL, serverDataID)
	util.Debug("9anime get source", "serverDataID", serverDataID, "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateAJAXRequest(req)

	resp, err := c.client.Do(req) // #nosec G704 -- URL built from known base domain
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data struct {
		Link   string `json:"link"`
		Server int    `json:"server"`
		Type   string `json:"type"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse source JSON: %w", err)
	}

	if data.Link == "" {
		return nil, errors.New("no embed URL returned by 9anime source API")
	}

	util.Debug("9anime source resolved", "embedURL", data.Link, "server", data.Server)
	return &NineAnimeStreamSource{
		EmbedURL:   data.Link,
		ServerType: data.Server,
	}, nil
}

// GetStreamInfo resolves full stream info from a rapid-cloud embed URL
func (c *NineAnimeClient) GetStreamInfo(embedURL string) (*NineAnimeStreamInfo, error) {
	return c.GetStreamInfoWithContext(context.Background(), embedURL)
}

// GetStreamInfoWithContext resolves stream info with context support
func (c *NineAnimeClient) GetStreamInfoWithContext(ctx context.Context, embedURL string) (*NineAnimeStreamInfo, error) {
	if embedURL == "" {
		return nil, errors.New("empty embed URL")
	}

	// Attempt 1: rapid-cloud getSources API
	rcRegex := regexp.MustCompile(`/embed-2/v2/e-1/([^?/]+)`)
	if matches := rcRegex.FindStringSubmatch(embedURL); len(matches) > 1 {
		videoID := matches[1]
		domainRegex := regexp.MustCompile(`^(https?://[^/]+)`)
		domainMatches := domainRegex.FindStringSubmatch(embedURL)
		baseDomain := "https://rapid-cloud.co"
		if len(domainMatches) > 1 {
			baseDomain = domainMatches[1]
		}

		sourcesURL := fmt.Sprintf("%s/embed-2/v2/e-1/getSources?id=%s", baseDomain, videoID)
		util.Debug("9anime rapid-cloud getSources", "url", sourcesURL)

		req, err := http.NewRequestWithContext(ctx, "GET", sourcesURL, nil)
		if err == nil {
			req.Header.Set("Referer", embedURL)
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
			req.Header.Set("User-Agent", c.userAgent)

			resp, err := c.client.Do(req) // #nosec G704 -- URL built from known rapid-cloud domain
			if err == nil {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode == http.StatusOK {
					body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
					if err == nil {
						info := c.parseRapidCloudResponse(body, embedURL)
						if info != nil {
							return info, nil
						}
					}
				}
			}
		}
	}

	// Attempt 2: scrape embed page for m3u8 URL
	info := c.scrapeEmbedPage(ctx, embedURL)
	if info != nil {
		return info, nil
	}

	return nil, errors.New("could not resolve stream from 9anime embed URL")
}

// parseRapidCloudResponse parses the rapid-cloud getSources API response
func (c *NineAnimeClient) parseRapidCloudResponse(body []byte, embedURL string) *NineAnimeStreamInfo {
	var data struct {
		Sources json.RawMessage `json:"sources"`
		Tracks  []struct {
			Label   string `json:"label"`
			File    string `json:"file"`
			Kind    string `json:"kind"`
			Default bool   `json:"default"`
		} `json:"tracks"`
		Intro struct {
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"intro"`
		Outro struct {
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"outro"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}

	// Parse sources - can be an array of objects or a raw string
	var m3u8URL string
	var sources []struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(data.Sources, &sources); err == nil && len(sources) > 0 {
		m3u8URL = sources[0].File
	} else {
		// Try as bare string
		var rawStr string
		if err := json.Unmarshal(data.Sources, &rawStr); err == nil && rawStr != "" {
			m3u8URL = rawStr
		}
	}

	if m3u8URL == "" {
		return nil
	}

	// Parse subtitle tracks
	var tracks []NineAnimeSubtitleTrack
	for _, t := range data.Tracks {
		if t.Kind == "captions" || t.Kind == "subtitles" {
			tracks = append(tracks, NineAnimeSubtitleTrack{
				Label:   t.Label,
				File:    t.File,
				Kind:    t.Kind,
				Default: t.Default,
			})
		}
	}

	util.Debug("9anime stream resolved via rapid-cloud",
		"m3u8", m3u8URL[:min(len(m3u8URL), 80)],
		"tracks", len(tracks),
		"introEnd", data.Intro.End,
	)

	return &NineAnimeStreamInfo{
		M3U8URL:    m3u8URL,
		Tracks:     tracks,
		IntroStart: data.Intro.Start,
		IntroEnd:   data.Intro.End,
		OutroStart: data.Outro.Start,
		OutroEnd:   data.Outro.End,
		EmbedURL:   embedURL,
	}
}

// scrapeEmbedPage scrapes the embed page HTML for m3u8 URLs
func (c *NineAnimeClient) scrapeEmbedPage(ctx context.Context, embedURL string) *NineAnimeStreamInfo {
	req, err := http.NewRequestWithContext(ctx, "GET", embedURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req) // #nosec G704 -- URL from embed page resolution
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil
	}

	text := string(body)

	// Try direct m3u8 URL pattern
	m3u8Regex := regexp.MustCompile(`(https?://[^\s"'<>]+\.m3u8[^\s"'<>]*)`)
	if matches := m3u8Regex.FindStringSubmatch(text); len(matches) > 1 {
		util.Debug("9anime m3u8 found via regex scrape", "url", matches[1][:min(len(matches[1]), 80)])
		return &NineAnimeStreamInfo{
			M3U8URL:  matches[1],
			EmbedURL: embedURL,
		}
	}

	// Try JSON "file" pattern
	fileRegex := regexp.MustCompile(`"file"\s*:\s*"(https?://[^"]+\.m3u8[^"]*)"`)
	if matches := fileRegex.FindStringSubmatch(text); len(matches) > 1 {
		util.Debug("9anime m3u8 found via JSON file pattern", "url", matches[1][:min(len(matches[1]), 80)])
		return &NineAnimeStreamInfo{
			M3U8URL:  matches[1],
			EmbedURL: embedURL,
		}
	}

	return nil
}

// GetStreamURL resolves the stream URL for an episode, trying multiple servers
// This is the main method used by the UnifiedScraper interface
func (c *NineAnimeClient) GetStreamURL(episodeID string, options ...any) (string, map[string]string, error) {
	// Parse options
	preferredAudio := "sub"
	if len(options) > 0 {
		if audio, ok := options[0].(string); ok {
			preferredAudio = audio
		}
	}

	// Get servers
	servers, err := c.GetServers(episodeID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get servers: %w", err)
	}
	if len(servers) == 0 {
		return "", nil, errors.New("no servers available for this episode")
	}

	// Filter by preferred audio type, fallback to all servers
	filtered := make([]NineAnimeServer, 0)
	for _, s := range servers {
		if s.AudioType == preferredAudio {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		filtered = servers
	}

	// Try each server until we get a working stream
	var lastErr error
	for _, server := range filtered {
		source, err := c.GetSource(server.DataID)
		if err != nil {
			lastErr = err
			util.Debug("9anime server failed", "server", server.Name, "error", err)
			continue
		}

		streamInfo, err := c.GetStreamInfo(source.EmbedURL)
		if err != nil {
			lastErr = err
			util.Debug("9anime stream resolution failed", "server", server.Name, "error", err)
			continue
		}

		if streamInfo.M3U8URL == "" {
			lastErr = errors.New("empty m3u8 URL")
			continue
		}

		metadata := map[string]string{
			"source":     "9anime",
			"server":     server.Name,
			"audio_type": server.AudioType,
		}

		// Add embed URL referer domain
		domainRegex := regexp.MustCompile(`^(https?://[^/]+)`)
		if domainMatches := domainRegex.FindStringSubmatch(source.EmbedURL); len(domainMatches) > 1 {
			metadata["referer"] = domainMatches[1] + "/"
		}

		// Add subtitle URLs to metadata
		if len(streamInfo.Tracks) > 0 {
			var subURLs []string
			for _, t := range streamInfo.Tracks {
				subURLs = append(subURLs, t.File)
			}
			metadata["subtitles"] = strings.Join(subURLs, ",")

			// Add subtitle labels
			var subLabels []string
			for _, t := range streamInfo.Tracks {
				subLabels = append(subLabels, t.Label)
			}
			metadata["subtitle_labels"] = strings.Join(subLabels, ",")
		}

		// Add skip times
		if streamInfo.IntroEnd > 0 {
			metadata["intro_start"] = strconv.Itoa(streamInfo.IntroStart)
			metadata["intro_end"] = strconv.Itoa(streamInfo.IntroEnd)
		}
		if streamInfo.OutroStart > 0 {
			metadata["outro_start"] = strconv.Itoa(streamInfo.OutroStart)
			metadata["outro_end"] = strconv.Itoa(streamInfo.OutroEnd)
		}

		util.Debug("9anime stream resolved successfully",
			"server", server.Name,
			"audio", server.AudioType,
			"url", streamInfo.M3U8URL[:min(len(streamInfo.M3U8URL), 60)],
		)

		return streamInfo.M3U8URL, metadata, nil
	}

	if lastErr != nil {
		return "", nil, fmt.Errorf("all servers failed, last error: %w", lastErr)
	}
	return "", nil, errors.New("no working stream found")
}

// ToAnimeModel converts NineAnimeResult to models.Anime for compatibility
func (r *NineAnimeResult) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:      r.Title,
		URL:       r.AnimeID,
		Source:    "9Anime",
		MediaType: models.MediaTypeAnime,
	}
	return anime
}
