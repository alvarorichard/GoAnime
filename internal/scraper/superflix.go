// Package scraper provides web scraping functionality for SuperFlix movies, TV shows, animes and doramas
package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	SuperFlixBase      = "https://superflixapi.rest"
	SuperFlixUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Pre-compiled regexes for SuperFlix scraper
var (
	sfCSRFTokenRe    = regexp.MustCompile(`var CSRF_TOKEN\s*=\s*"([^"]+)"`)
	sfPageTokenRe    = regexp.MustCompile(`var PAGE_TOKEN\s*=\s*"([^"]+)"`)
	sfContentIDRe    = regexp.MustCompile(`var INITIAL_CONTENT_ID\s*=\s*(\d+)`)
	sfContentTypeRe  = regexp.MustCompile(`var CONTENT_TYPE\s*=\s*"([^"]+)"`)
	sfTitleRe        = regexp.MustCompile(`<title>(?:Player \| )?(.+?)</title>`)
	sfAllEpisodesRe  = regexp.MustCompile(`var ALL_EPISODES\s*=\s*(\{.+?\});`)
	sfDefaultAudioRe = regexp.MustCompile(`var defaultAudio\s*=\s*(\[.+?\]);`)
	sfSubtitleRe     = regexp.MustCompile(`var playerjsSubtitle\s*=\s*"(.+?)";`)
	sfSubPartRe      = regexp.MustCompile(`\[(.+?)\](https?://.+)`)
)

// SuperFlixTokens holds the tokens extracted from a SuperFlix player page
type SuperFlixTokens struct {
	CSRF        string
	PageToken   string
	ContentID   string
	ContentType string
	Title       string
}

// SuperFlixServer represents a streaming server option
type SuperFlixServer struct {
	ID   json.RawMessage `json:"ID"`
	Name string          `json:"name"`
}

// SuperFlixSubtitle represents a subtitle track
type SuperFlixSubtitle struct {
	Lang string
	URL  string
}

// SuperFlixStreamResult holds the final stream extraction result
type SuperFlixStreamResult struct {
	StreamURL    string
	Title        string
	Referer      string
	Subtitles    []SuperFlixSubtitle
	DefaultAudio []string
	Thumb        string
}

// SuperFlixEpisode represents a single episode in a season
type SuperFlixEpisode struct {
	EpiNum  json.Number `json:"epi_num"`
	Title   string      `json:"title"`
	AirDate string      `json:"air_date"`
}

// SuperFlixMedia represents a search result from SuperFlix
type SuperFlixMedia struct {
	Title    string
	Year     string
	Type     string // "Filme", "Série", etc.
	SFType   string // "filme" or "serie"
	TMDBID   string
	IMDBID   string
	ImageURL string // Cover image URL from search results
}

// SuperFlixClient handles interactions with SuperFlix
type SuperFlixClient struct {
	client      *http.Client
	baseURL     string
	userAgent   string
	maxRetries  int
	retryDelay  time.Duration
	searchCache sync.Map
}

// NormalizeSuperFlixImageURL converts SuperFlix CloudFront proxy URLs to direct TMDB image URLs.
// Discord's image proxy cannot handle the double-URL format used by SuperFlix:
//
//	https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster.jpg
//
// This extracts the embedded TMDB URL and upgrades to w500 quality:
//
//	https://image.tmdb.org/t/p/w500/poster.jpg
func NormalizeSuperFlixImageURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}
	const tmdbPrefix = "https://image.tmdb.org/t/p/"
	if idx := strings.Index(imageURL, tmdbPrefix); idx > 0 {
		direct := imageURL[idx:]
		// Upgrade thumbnail size for Discord display
		direct = strings.Replace(direct, "/w342/", "/w500/", 1)
		direct = strings.Replace(direct, "/w185/", "/w500/", 1)
		direct = strings.Replace(direct, "/w154/", "/w500/", 1)
		return direct
	}
	return imageURL
}

// NewSuperFlixClient creates a new SuperFlix client
func NewSuperFlixClient() *SuperFlixClient {
	return &SuperFlixClient{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: safeScraperTransport(30 * time.Second),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		baseURL:    SuperFlixBase,
		userAgent:  SuperFlixUserAgent,
		maxRetries: 2,
		retryDelay: 200 * time.Millisecond,
	}
}

func (c *SuperFlixClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
}

// SetTestConfig overrides the base URL and HTTP client for testing.
// This should only be used in test code.
func (c *SuperFlixClient) SetTestConfig(baseURL string, httpClient *http.Client) {
	c.baseURL = baseURL
	c.client = httpClient
	c.maxRetries = 0
	c.retryDelay = 0
}

// SearchMedia searches SuperFlix for movies/series/animes
func (c *SuperFlixClient) SearchMedia(query string) ([]*SuperFlixMedia, error) {
	return c.SearchMediaWithContext(context.Background(), query)
}

// SearchMediaWithContext searches with context support
func (c *SuperFlixClient) SearchMediaWithContext(ctx context.Context, query string) ([]*SuperFlixMedia, error) {
	cacheKey := strings.ToLower(strings.TrimSpace(query))
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*SuperFlixMedia), nil
	}

	searchURL := fmt.Sprintf("%s/pesquisar?s=%s", c.baseURL, url.QueryEscape(query))
	util.Debug("SuperFlix search", "query", query, "url", searchURL)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
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
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	results := c.parseCards(doc)
	c.searchCache.Store(cacheKey, results)
	return results, nil
}

// parseCards extracts media cards from SuperFlix HTML
func (c *SuperFlixClient) parseCards(doc *goquery.Document) []*SuperFlixMedia {
	var results []*SuperFlixMedia
	seen := make(map[string]bool)

	doc.Find("div.group\\/card").Each(func(i int, card *goquery.Selection) {
		var title, imageURL string
		if img := card.Find("img"); img.Length() > 0 {
			title, _ = img.Attr("alt")
			// Extract cover image URL from src, data-src, or srcset
			if src, ok := img.Attr("src"); ok && src != "" && !strings.HasPrefix(src, "data:") {
				imageURL = src
			}
			if imageURL == "" {
				if dataSrc, ok := img.Attr("data-src"); ok && dataSrc != "" {
					imageURL = dataSrc
				}
			}
			if imageURL == "" {
				if srcset, ok := img.Attr("srcset"); ok && srcset != "" {
					// Take the first URL from srcset (format: "url size, url size, ...")
					parts := strings.Fields(strings.Split(srcset, ",")[0])
					if len(parts) > 0 {
						imageURL = parts[0]
					}
				}
			}
		}
		if title == "" {
			if h3 := card.Find("h3"); h3.Length() > 0 {
				title = strings.TrimSpace(h3.Text())
			}
		}
		if title == "" {
			return
		}

		var tmdbID, imdbID, linkURL string

		card.Find("button").Each(func(_ int, btn *goquery.Selection) {
			msg, _ := btn.Attr("data-msg")
			copyVal, _ := btn.Attr("data-copy")
			if strings.Contains(msg, "TMDB") {
				tmdbID = copyVal
			} else if strings.Contains(msg, "IMDB") {
				imdbID = copyVal
			} else if strings.Contains(msg, "Link") {
				linkURL = copyVal
			}
		})

		// Extract type and year from metadata
		meta := card.Find("div.mt-3")
		metaText := strings.TrimSpace(meta.Text())
		metaParts := splitAndTrim(metaText, "|")

		var tipo, year string
		if len(metaParts) > 0 {
			tipo = metaParts[len(metaParts)-1]
		}
		if len(metaParts) > 1 {
			year = metaParts[1]
		}

		sfType := "serie"
		if strings.Contains(linkURL, "/filme/") {
			sfType = "filme"
		}

		key := tmdbID
		if key == "" {
			key = title
		}
		if seen[key] {
			return
		}
		seen[key] = true

		if tipo == "" {
			if sfType == "filme" {
				tipo = "Filme"
			} else {
				tipo = "Série"
			}
		}

		results = append(results, &SuperFlixMedia{
			Title:    title,
			Year:     year,
			Type:     tipo,
			SFType:   sfType,
			TMDBID:   tmdbID,
			IMDBID:   imdbID,
			ImageURL: NormalizeSuperFlixImageURL(imageURL),
		})
	})

	return results
}

// GetPlayerPage loads the player page for a given content
func (c *SuperFlixClient) GetPlayerPage(ctx context.Context, mediaType, mediaID, season, episode string) (string, error) {
	path := fmt.Sprintf("/%s/%s", mediaType, mediaID)
	if season != "" {
		path += "/" + season
	}
	if episode != "" {
		path += "/" + episode
	}

	pageURL := c.baseURL + path
	util.Debug("SuperFlix player page", "url", pageURL)

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// ExtractTokens extracts CSRF_TOKEN, PAGE_TOKEN, etc. from player HTML
func (c *SuperFlixClient) ExtractTokens(html string) *SuperFlixTokens {
	tokens := &SuperFlixTokens{}
	if m := sfCSRFTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.CSRF = m[1]
	}
	if m := sfPageTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.PageToken = m[1]
	}
	if m := sfContentIDRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentID = m[1]
	}
	if m := sfContentTypeRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentType = m[1]
	}
	if m := sfTitleRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.Title = m[1]
	}
	return tokens
}

// ExtractEpisodes extracts ALL_EPISODES from the player page JS
func (c *SuperFlixClient) ExtractEpisodes(html string) (map[string][]SuperFlixEpisode, error) {
	m := sfAllEpisodesRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil, nil
	}

	var result map[string][]SuperFlixEpisode
	if err := json.Unmarshal([]byte(m[1]), &result); err != nil {
		return nil, fmt.Errorf("failed to parse ALL_EPISODES: %w", err)
	}
	return result, nil
}

// Bootstrap calls /player/bootstrap to get server list
func (c *SuperFlixClient) Bootstrap(ctx context.Context, tokens *SuperFlixTokens) ([]SuperFlixServer, error) {
	bootstrapURL := c.baseURL + "/player/bootstrap"

	form := url.Values{
		"contentid":  {tokens.ContentID},
		"type":       {tokens.ContentType},
		"_token":     {tokens.CSRF},
		"page_token": {tokens.PageToken},
		"pageToken":  {tokens.PageToken},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Data struct {
			Options []SuperFlixServer `json:"options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode bootstrap response: %w", err)
	}

	return result.Data.Options, nil
}

// GetSourceURL calls /player/source to get the redirect URL for a video
func (c *SuperFlixClient) GetSourceURL(ctx context.Context, videoID string, tokens *SuperFlixTokens) (string, error) {
	sourceURL := c.baseURL + "/player/source"

	form := url.Values{
		"video_id":   {videoID},
		"page_token": {tokens.PageToken},
		"host":       {""},
		"site":       {""},
		"_token":     {tokens.CSRF},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sourceURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Data struct {
			VideoURL string `json:"video_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode source response: %w", err)
	}

	if result.Data.VideoURL == "" {
		return "", fmt.Errorf("no video URL in source response")
	}

	return result.Data.VideoURL, nil
}

// ResolveRedirect follows the SuperFlix redirect to get the external player URL
func (c *SuperFlixClient) ResolveRedirect(ctx context.Context, redirectURL string) (baseURL, videoHash, playerHTML string, err error) {
	// Use the client's transport if available, otherwise fall back to safe transport
	transport := c.client.Transport
	if transport == nil {
		transport = safeScraperTransport(30 * time.Second)
	}

	// Use a client that does NOT follow redirects automatically
	noRedirectClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", redirectURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	location := redirectURL
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location = resp.Header.Get("Location")
		if location == "" {
			return "", "", "", fmt.Errorf("redirect with no Location header")
		}
	}

	// Follow to the final page
	req2, err := http.NewRequestWithContext(ctx, "GET", location, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create follow request: %w", err)
	}
	c.decorateRequest(req2)
	req2.Header.Set("Referer", c.baseURL+"/")

	followClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	resp2, err := followClient.Do(req2)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to follow redirect: %w", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp2.Body, 5*1024*1024))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read player page: %w", err)
	}

	finalURL := resp2.Request.URL.String()

	if strings.Contains(finalURL, "/video/") {
		parts := strings.SplitN(finalURL, "/video/", 2)
		baseURL = parts[0]
		videoHash = strings.SplitN(parts[1], "?", 2)[0]
		videoHash = strings.SplitN(videoHash, "#", 2)[0]
	} else {
		idx := strings.LastIndex(finalURL, "/")
		if idx > 0 {
			baseURL = finalURL[:idx]
			videoHash = strings.SplitN(finalURL[idx+1:], "?", 2)[0]
		}
	}

	return baseURL, videoHash, string(body), nil
}

// ExtractPlayerExtras extracts defaultAudio and subtitles from the external player HTML
func (c *SuperFlixClient) ExtractPlayerExtras(html string) (defaultAudio []string, subtitles []SuperFlixSubtitle) {
	if m := sfDefaultAudioRe.FindStringSubmatch(html); len(m) > 1 {
		_ = json.Unmarshal([]byte(m[1]), &defaultAudio)
	}

	if m := sfSubtitleRe.FindStringSubmatch(html); len(m) > 1 {
		for part := range strings.SplitSeq(m[1], ",") {
			sm := sfSubPartRe.FindStringSubmatch(part)
			if len(sm) > 2 {
				subtitles = append(subtitles, SuperFlixSubtitle{
					Lang: sm[1],
					URL:  sm[2],
				})
			}
		}
	}
	return
}

// GetVideoAPI calls the external player's API to get the actual stream URL
func (c *SuperFlixClient) GetVideoAPI(ctx context.Context, playerBaseURL, videoHash, referer string) (streamURL, thumbURL string, err error) {
	apiURL := fmt.Sprintf("%s/player/index.php?data=%s&do=getVideo", playerBaseURL, videoHash)

	form := url.Values{
		"hash": {videoHash},
		"r":    {c.baseURL + "/"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		SecuredLink string `json:"securedLink"`
		VideoSource string `json:"videoSource"`
		VideoImage  string `json:"videoImage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to decode video API response: %w", err)
	}

	if result.SecuredLink != "" {
		streamURL = result.SecuredLink
	} else if result.VideoSource != "" {
		streamURL = result.VideoSource
	} else {
		return "", "", fmt.Errorf("no stream URL in video API response")
	}

	return streamURL, result.VideoImage, nil
}

// GetStreamURL is the full pipeline: player page → tokens → bootstrap → source → redirect → video API
func (c *SuperFlixClient) GetStreamURL(ctx context.Context, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	html, err := c.GetPlayerPage(ctx, mediaType, mediaID, season, episode)
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}

	tokens := c.ExtractTokens(html)
	if tokens.CSRF == "" || tokens.PageToken == "" {
		return nil, fmt.Errorf("failed to extract tokens from player page")
	}

	servers, err := c.Bootstrap(ctx, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap: %w", err)
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Pick first non-fallback server
	var videoIDStr string
	for _, s := range servers {
		var raw string
		if err := json.Unmarshal(s.ID, &raw); err == nil {
			if !strings.HasPrefix(raw, "fallback") {
				videoIDStr = raw
				break
			}
		}
		// Try as number
		var num json.Number
		if err := json.Unmarshal(s.ID, &num); err == nil {
			videoIDStr = num.String()
			break
		}
	}
	if videoIDStr == "" {
		// Fallback: use first server
		var raw string
		if err := json.Unmarshal(servers[0].ID, &raw); err == nil {
			videoIDStr = raw
		} else {
			var num json.Number
			if err := json.Unmarshal(servers[0].ID, &num); err == nil {
				videoIDStr = num.String()
			} else {
				return nil, fmt.Errorf("failed to parse server ID")
			}
		}
	}

	redirectURL, err := c.GetSourceURL(ctx, videoIDStr, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get source URL: %w", err)
	}

	playerBaseURL, videoHash, playerHTML, err := c.ResolveRedirect(ctx, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve redirect: %w", err)
	}

	referer := fmt.Sprintf("%s/video/%s", playerBaseURL, videoHash)
	streamURL, thumbURL, err := c.GetVideoAPI(ctx, playerBaseURL, videoHash, referer)
	if err != nil {
		return nil, fmt.Errorf("failed to get video from API: %w", err)
	}

	result := &SuperFlixStreamResult{
		StreamURL: streamURL,
		Title:     tokens.Title,
		Referer:   playerBaseURL + "/",
		Thumb:     NormalizeSuperFlixImageURL(thumbURL),
	}

	defaultAudio, subtitles := c.ExtractPlayerExtras(playerHTML)
	result.DefaultAudio = defaultAudio
	result.Subtitles = subtitles

	return result, nil
}

// GetEpisodes returns the seasons and episodes for a series
func (c *SuperFlixClient) GetEpisodes(ctx context.Context, tmdbID string) (map[string][]SuperFlixEpisode, error) {
	html, err := c.GetPlayerPage(ctx, "serie", tmdbID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}
	return c.ExtractEpisodes(html)
}

// ToAnimeModel converts SuperFlixMedia to models.Anime for compatibility
func (m *SuperFlixMedia) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.TMDBID, // Store TMDB ID as URL identifier
		Source:   "SuperFlix",
		Year:     m.Year,
		ImageURL: m.ImageURL,
	}

	lowerType := strings.ToLower(m.Type)
	switch {
	case m.SFType == "filme":
		anime.MediaType = models.MediaTypeMovie
	case lowerType == "anime" || lowerType == "dorama":
		anime.MediaType = models.MediaTypeAnime
	default:
		anime.MediaType = models.MediaTypeTV
	}

	if m.IMDBID != "" {
		anime.IMDBID = m.IMDBID
	}

	// Parse TMDB ID for direct API lookups during enrichment
	if m.TMDBID != "" {
		if id, err := strconv.Atoi(m.TMDBID); err == nil {
			anime.TMDBID = id
		}
	}

	util.Debug("SuperFlix ToAnimeModel", "title", m.Title, "tmdbID", m.TMDBID, "imageURL", anime.ImageURL)

	return anime
}

// Helper: split string by separator and trim each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
