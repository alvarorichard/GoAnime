// Package scraper provides web scraping functionality for anime sources
package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	AllAnimeReferer = "https://allanime.to"
	AllAnimeBase    = "allanime.day"
	AllAnimeAPI     = "https://api.allanime.day/api"
	UserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"
)

// AllAnimeClient handles interactions with AllAnime API
type AllAnimeClient struct {
	client    *http.Client
	referer   string
	apiBase   string
	userAgent string
}

// NewAllAnimeClient creates a new AllAnime client
func NewAllAnimeClient() *AllAnimeClient {
	return &AllAnimeClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		referer:   AllAnimeReferer,
		apiBase:   AllAnimeAPI,
		userAgent: UserAgent,
	}
}

// SearchResponse represents the API response structure for anime search
type SearchResponse struct {
	Data struct {
		Shows struct {
			Edges []struct {
				ID                string `json:"_id"`
				Name              string `json:"name"`
				AvailableEpisodes struct {
					Sub int `json:"sub"`
					Dub int `json:"dub"`
				} `json:"availableEpisodes"`
			} `json:"edges"`
		} `json:"shows"`
	} `json:"data"`
}

// EpisodeResponse represents the API response for episode details
type EpisodeResponse struct {
	Data struct {
		Episode struct {
			EpisodeString string `json:"episodeString"`
			SourceUrls    []struct {
				SourceName string `json:"sourceName"`
				SourceUrl  string `json:"sourceUrl"`
			} `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
}

// EpisodesListResponse represents the API response for episodes list
type EpisodesListResponse struct {
	Data struct {
		Show struct {
			ID                      string                 `json:"_id"`
			AvailableEpisodesDetail map[string]interface{} `json:"availableEpisodesDetail"`
		} `json:"show"`
	} `json:"data"`
}

// SearchAnime searches for anime using AllAnime API (based on Curd implementation)
func (c *AllAnimeClient) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
	// Use the exact same GraphQL query as Curd
	searchGql := `query($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
		shows(search: $search, limit: $limit, page: $page, translationType: $translationType, countryOrigin: $countryOrigin) {
			edges {
				_id
				name
				englishName
				availableEpisodes
				__typename
			}
		}
	}`

	// Prepare the GraphQL variables exactly like Curd
	variables := map[string]interface{}{
		"search": map[string]interface{}{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":           40,
		"page":            1,
		"translationType": "sub",
		"countryOrigin":   "ALL",
	}

	// Marshal the variables to JSON
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	// Build the request URL exactly like Curd
	reqURL := fmt.Sprintf("%s?variables=%s&query=%s", c.apiBase, url.QueryEscape(string(variablesJSON)), url.QueryEscape(searchGql))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse using a simple structure like Curd
	var response struct {
		Data struct {
			Shows struct {
				Edges []struct {
					ID                string      `json:"_id"`
					Name              string      `json:"name"`
					EnglishName       string      `json:"englishName"`
					AvailableEpisodes interface{} `json:"availableEpisodes"`
				} `json:"edges"`
			} `json:"shows"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var animes []*models.Anime
	for _, edge := range response.Data.Shows.Edges {
		var episodesStr string
		if episodes, ok := edge.AvailableEpisodes.(map[string]interface{}); ok {
			if subEpisodes, ok := episodes["sub"].(float64); ok {
				episodesStr = fmt.Sprintf("(%d episodes)", int(subEpisodes))
			} else {
				episodesStr = "(Unknown episodes)"
			}
		}

		// Use English name if available, otherwise use default name
		displayName := edge.Name
		if edge.EnglishName != "" {
			displayName = edge.EnglishName
		}

		anime := &models.Anime{
			Name: fmt.Sprintf("%s %s", displayName, episodesStr),
			URL:  edge.ID, // For AllAnime, the "URL" is actually the anime ID
		}
		animes = append(animes, anime)
	}

	return animes, nil
}

// GetEpisodesList gets the list of available episodes for an anime (based on Curd implementation)
func (c *AllAnimeClient) GetEpisodesList(animeID string, mode string) ([]string, error) {
	if mode == "" {
		mode = "sub"
	}

	episodesListGql := `query ($showId: String!) { show( _id: $showId ) { _id availableEpisodesDetail }}`

	// Correctly URL encode the parameters like Curd does
	variables := fmt.Sprintf(`{"showId":"%s"}`, animeID)
	reqURL := fmt.Sprintf("%s?variables=%s&query=%s",
		c.apiBase,
		url.QueryEscape(variables),
		url.QueryEscape(episodesListGql))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Use the same response structure as Curd
	var response struct {
		Data struct {
			Show struct {
				ID                      string                 `json:"_id"`
				AvailableEpisodesDetail map[string]interface{} `json:"availableEpisodesDetail"`
			} `json:"show"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract and sort the episodes exactly like Curd
	episodes := extractEpisodes(response.Data.Show.AvailableEpisodesDetail, mode)
	return episodes, nil
}

// Helper function to get map keys for debugging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// extractEpisodes extracts the episodes list from the availableEpisodesDetail field (from Curd)
func extractEpisodes(availableEpisodesDetail map[string]interface{}, mode string) []string {
	var episodes []float64

	// Check if the mode (e.g., "sub") exists in the map
	if eps, ok := availableEpisodesDetail[mode].([]interface{}); ok {
		for _, ep := range eps {
			var epNum float64
			if n, _ := fmt.Sscanf(fmt.Sprintf("%v", ep), "%f", &epNum); n == 1 {
				episodes = append(episodes, epNum)
			}
		}
	}

	// Sort episodes numerically
	sort.Float64s(episodes)

	// Convert to string and return
	var episodesStr []string
	for _, ep := range episodes {
		episodesStr = append(episodesStr, fmt.Sprintf("%v", ep))
	}
	return episodesStr
}

// GetAnimeEpisodes converts AllAnime episode list to models.Episode format
func (c *AllAnimeClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// Extract anime ID from URL (animeURL should be the anime ID for AllAnime)
	animeID := animeURL

	// Get episode list using existing function
	episodeStrings, err := c.GetEpisodesList(animeID, "sub")
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes list: %w", err)
	}

	// Convert to models.Episode format
	var episodes []models.Episode
	for _, epStr := range episodeStrings {
		episodes = append(episodes, models.Episode{
			Number: epStr,
			Num:    parseEpisodeNum(epStr),
			URL:    epStr, // For AllAnime, the episode "URL" is just the episode number
		})
	}

	return episodes, nil
}

// GetAnimeEpisodesWithAniSkip converts AllAnime episode list to models.Episode format and enriches with AniSkip data (like Curd)
func (c *AllAnimeClient) GetAnimeEpisodesWithAniSkip(animeURL string, malID int, aniSkipFunc func(int, int, *models.Episode) error) ([]models.Episode, error) {
	// Get basic episodes first
	episodes, err := c.GetAnimeEpisodes(animeURL)
	if err != nil {
		return nil, err
	}

	// Enrich with AniSkip data for each episode (like Curd does)
	if malID > 0 && aniSkipFunc != nil {
		for i := range episodes {
			episodeNum := episodes[i].Num
			if episodeNum > 0 {
				// Try to get AniSkip data for this episode
				if err := aniSkipFunc(malID, episodeNum, &episodes[i]); err != nil {
					// Not an error if AniSkip data is not found, just log it
					util.Debugf("AniSkip data not found for episode %d: %v", episodeNum, err)
				}
			}
		}
	}

	return episodes, nil
}

// SendSkipTimesToMPV sends OP and ED timings to MPV as chapter markers (based on Curd implementation)
func (c *AllAnimeClient) SendSkipTimesToMPV(episode *models.Episode, socketPath string, mpvSendCommand func(string, []interface{}) (interface{}, error)) error {
	// Only proceed if we have valid skip times
	if episode.SkipTimes.Op.Start == 0 && episode.SkipTimes.Op.End == 0 &&
		episode.SkipTimes.Ed.Start == 0 && episode.SkipTimes.Ed.End == 0 {
		return fmt.Errorf("no skip times available for episode")
	}

	// Create chapter list exactly like Curd does
	chapterList := []map[string]interface{}{}

	// Pre-Opening chapter
	if episode.SkipTimes.Op.Start > 0 {
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Pre-Opening",
			"time":  0.0,
			"end":   float64(episode.SkipTimes.Op.Start),
		})
	}

	// Opening chapter
	if episode.SkipTimes.Op.Start > 0 && episode.SkipTimes.Op.End > episode.SkipTimes.Op.Start {
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Opening",
			"time":  float64(episode.SkipTimes.Op.Start),
			"end":   float64(episode.SkipTimes.Op.End),
		})
	}

	// Main content chapter
	mainStart := float64(episode.SkipTimes.Op.End)
	if mainStart == 0 {
		mainStart = 0.0
	}
	mainEnd := float64(episode.SkipTimes.Ed.Start)
	if mainEnd == 0 {
		// If no ending skip time, don't set an end for main content
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Main",
			"time":  mainStart,
		})
	} else {
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Main",
			"time":  mainStart,
			"end":   mainEnd,
		})
	}

	// Ending chapter
	if episode.SkipTimes.Ed.Start > 0 && episode.SkipTimes.Ed.End > episode.SkipTimes.Ed.Start {
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Ending",
			"time":  float64(episode.SkipTimes.Ed.Start),
			"end":   float64(episode.SkipTimes.Ed.End),
		})
	}

	// Post-Credits chapter
	if episode.SkipTimes.Ed.End > 0 {
		chapterList = append(chapterList, map[string]interface{}{
			"title": "Post-Credits",
			"time":  float64(episode.SkipTimes.Ed.End),
		})
	}

	// Send chapter list to MPV exactly like Curd does
	_, err := mpvSendCommand(socketPath, []interface{}{
		"set_property",
		"chapter-list",
		chapterList,
	})
	if err != nil {
		return fmt.Errorf("error sending chapter list to MPV: %w", err)
	}

	util.Debug("AniSkip chapter markers sent to MPV successfully")
	return nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseEpisodeNum converts episode string to integer
func parseEpisodeNum(epStr string) int {
	// Try to extract number from string
	var num int
	fmt.Sscanf(epStr, "%d", &num)
	if num == 0 {
		num = 1 // Default to 1 if parsing fails
	}
	return num
}

// GetAnimeDetails - placeholder method for interface consistency
func (c *AllAnimeClient) GetAnimeDetails(animeURL string) (*models.Anime, error) {
	return nil, fmt.Errorf("anime details should be fetched using API layer, not scraper")
}

// LinkPriorities defines the order of priority for link domains (from Curd project)
var LinkPriorities = []string{
	"sharepoint.com",
	"wixmp.com",
	"dropbox.com",
	"wetransfer.com",
	"gogoanime.com",
}

// GetEpisodeURL gets the streaming URL for a specific episode using priority-based selection
func (c *AllAnimeClient) GetEpisodeURL(animeID string, episodeNo string, mode string, quality string) (string, map[string]string, error) {
	if mode == "" {
		mode = "sub"
	}
	if quality == "" {
		quality = "best"
	}

	episodeEmbedGQL := `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) { episode( showId: $showId translationType: $translationType episodeString: $episodeString ) { episodeString sourceUrls }}`
	variables := fmt.Sprintf(`{"showId":"%s","translationType":"%s","episodeString":"%s"}`, animeID, mode, episodeNo)

	req, err := http.NewRequest("GET", c.apiBase+"/api", nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("variables", variables)
	q.Add("query", episodeEmbedGQL)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response to extract source URLs
	sourceURLs := c.extractSourceURLs(string(body))
	if len(sourceURLs) == 0 {
		return "", nil, fmt.Errorf("no source URLs found for episode %s", episodeNo)
	}

	// Process URLs concurrently like Curd does
	return c.processSourceURLsConcurrent(sourceURLs, quality, animeID, episodeNo)
}

// processSourceURLsConcurrent processes source URLs with concurrent requests and priority-based selection
func (c *AllAnimeClient) processSourceURLsConcurrent(sourceURLs []string, quality string, animeID string, episodeNo string) (string, map[string]string, error) {
	type result struct {
		index     int
		links     map[string]string
		err       error
		sourceURL string
	}

	results := make(chan result, len(sourceURLs))
	highPriorityLink := make(chan string, 1)

	// Rate limiter like in Curd
	rateLimiter := time.NewTicker(50 * time.Millisecond)
	defer rateLimiter.Stop()

	// Launch goroutines for concurrent processing
	for i, sourceURL := range sourceURLs {
		go func(idx int, url string) {
			<-rateLimiter.C // Rate limit the requests

			links, err := c.getLinks(url)
			if err != nil {
				results <- result{index: idx, err: err, sourceURL: url}
				return
			}

			// Check for high priority links first
			for _, link := range links {
				for _, domain := range LinkPriorities[:3] { // Check top 3 priority domains
					if strings.Contains(link, domain) {
						// Found high priority link, send it immediately
						select {
						case highPriorityLink <- link:
						default:
							// Channel already has a high priority link
						}
						break
					}
				}
			}

			results <- result{index: idx, links: links, sourceURL: url}
		}(i, sourceURL)
	}

	// First, try to get a high priority link quickly
	select {
	case link := <-highPriorityLink:
		// Found high priority link, return it immediately
		metadata := map[string]string{
			"quality":  quality,
			"anime_id": animeID,
			"episode":  episodeNo,
			"priority": "high",
		}
		return link, metadata, nil
	case <-time.After(2 * time.Second): // Wait briefly for high priority link
		// No high priority link found quickly, proceed with normal collection
	}

	// Collect results with timeout
	timeout := time.After(10 * time.Second)
	successCount := 0
	var bestURL string
	var bestMetadata map[string]string

	for successCount < len(sourceURLs) {
		select {
		case res := <-results:
			if res.err != nil {
				continue
			}

			// Select quality from the links
			selectedURL, metadata := c.selectQuality(res.links, quality)
			if selectedURL != "" {
				// Check if this is a prioritized link
				priority := c.getPriorityScore(selectedURL)
				if priority > 0 || bestURL == "" {
					bestURL = selectedURL
					bestMetadata = metadata
					bestMetadata["source_url"] = res.sourceURL
					bestMetadata["anime_id"] = animeID
					bestMetadata["episode"] = episodeNo

					if priority > 0 {
						// Found a priority link, return immediately
						return bestURL, bestMetadata, nil
					}
				}
			}
			successCount++

		case <-timeout:
			if bestURL != "" {
				return bestURL, bestMetadata, nil
			}
			return "", nil, fmt.Errorf("timeout waiting for results")
		}
	}

	if bestURL != "" {
		return bestURL, bestMetadata, nil
	}

	return "", nil, fmt.Errorf("no suitable quality found from any source")
}

// getPriorityScore returns the priority score of a URL based on domain
func (c *AllAnimeClient) getPriorityScore(url string) int {
	for i, domain := range LinkPriorities {
		if strings.Contains(url, domain) {
			return len(LinkPriorities) - i // Higher index means higher priority
		}
	}
	return 0
}

// extractSourceURLs extracts source URLs from the API response
func (c *AllAnimeClient) extractSourceURLs(response string) []string {
	// Parse the response as JSON to extract sourceUrls properly
	var episodeResp EpisodeResponse
	if err := json.Unmarshal([]byte(response), &episodeResp); err == nil {
		var urls []string
		for _, sourceUrl := range episodeResp.Data.Episode.SourceUrls {
			if strings.HasPrefix(sourceUrl.SourceUrl, "--") {
				// This is an encoded URL that needs decoding
				encoded := strings.TrimPrefix(sourceUrl.SourceUrl, "--")
				decoded := c.decodeSourceURL(encoded)
				urls = append(urls, decoded)
			} else {
				// Direct URL
				urls = append(urls, sourceUrl.SourceUrl)
			}
		}
		return urls
	}

	// Fallback to regex-based extraction if JSON parsing fails
	re := regexp.MustCompile(`"sourceUrl":"--([^"]*)".*?"sourceName":"([^"]*)"`)
	matches := re.FindAllStringSubmatch(response, -1)

	var urls []string
	for _, match := range matches {
		if len(match) >= 2 {
			// Decode the URL using the complex decoding logic from ani-cli
			decodedURL := c.decodeSourceURL(match[1])
			urls = append(urls, decodedURL)
		}
	}

	return urls
}

// decodeSourceURL decodes the encoded source URL using the exact logic from Curd
func (c *AllAnimeClient) decodeSourceURL(encoded string) string {
	// Handle the case where the encoded string might contain a colon and port
	parts := strings.Split(encoded, ":")
	mainPart := parts[0]
	port := ""
	if len(parts) > 1 {
		port = ":" + parts[1]
	}

	// Create mapping exactly like Curd's decodeProviderID function
	replacements := map[string]string{
		"01": "9", "08": "0", "05": "=", "0a": "2", "0b": "3", "0c": "4", "07": "?",
		"00": "8", "5c": "d", "0f": "7", "5e": "f", "17": "/", "54": "l", "09": "1",
		"48": "p", "4f": "w", "0e": "6", "5b": "c", "5d": "e", "0d": "5", "53": "k",
		"1e": "&", "5a": "b", "59": "a", "4a": "r", "4c": "t", "4e": "v", "57": "o",
		"51": "i",
	}

	// Split the string into pairs of characters
	re := regexp.MustCompile("..")
	pairs := re.FindAllString(mainPart, -1)

	// Perform the replacement
	for i, pair := range pairs {
		if val, exists := replacements[pair]; exists {
			pairs[i] = val
		}
	}

	// Join the modified pairs back into a single string
	result := strings.Join(pairs, "") + port

	// Replace "/clock" with "/clock.json" like in Curd
	result = strings.ReplaceAll(result, "/clock", "/clock.json")

	// If it starts with /, it's a path that needs the AllAnime base
	if strings.HasPrefix(result, "/") {
		result = fmt.Sprintf("https://%s%s", AllAnimeBase, result)
	}

	return result
}

// getLinks extracts video links from the source URL with proper headers
func (c *AllAnimeClient) getLinks(sourceURL string) (map[string]string, error) {
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use the same headers as Curd for better compatibility
	req.Header.Set("Referer", "https://allanime.to")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	links := c.extractVideoLinks(string(body))

	// Apply priority-based link selection
	return c.prioritizeLinks(links), nil
}

// prioritizeLinks applies priority-based sorting to video links
func (c *AllAnimeClient) prioritizeLinks(links map[string]string) map[string]string {
	prioritizedLinks := make(map[string]string)

	// First, add high priority links
	for quality, link := range links {
		for _, domain := range LinkPriorities {
			if strings.Contains(link, domain) {
				prioritizedLinks[quality+"_priority"] = link
			}
		}
	}

	// Then add regular links
	for quality, link := range links {
		prioritizedLinks[quality] = link
	}

	return prioritizedLinks
}

// extractVideoLinks extracts video links from the response with debug logging
func (c *AllAnimeClient) extractVideoLinks(response string) map[string]string {
	links := make(map[string]string)

	// Debug: log response structure
	util.Debugf("Response length: %d", len(response))

	// Parse JSON response
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
		// Extract links from JSON structure
		if linksInterface, ok := jsonData["links"].([]interface{}); ok {
			for _, linkInterface := range linksInterface {
				if linkMap, ok := linkInterface.(map[string]interface{}); ok {
					if link, ok := linkMap["link"].(string); ok {
						quality := "unknown"
						if resStr, ok := linkMap["resolutionStr"].(string); ok {
							quality = resStr
						} else if hls, ok := linkMap["hls"].(bool); ok && hls {
							quality = "hls"
						}

						link = strings.ReplaceAll(link, "\\", "")
						links[quality] = link
						util.Debugf("Found link - Quality: %s, URL: %s", quality, link)
					}
				}
			}
		}
	}

	// Fallback: Extract mp4 links with quality information using regex
	re := regexp.MustCompile(`"link":"([^"]*)".*?"resolutionStr":"([^"]*)"`)
	matches := re.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			quality := match[2]
			link := match[1]
			// Clean up the link
			link = strings.ReplaceAll(link, "\\", "")
			links[quality] = link
			util.Debugf("Regex found link - Quality: %s, URL: %s", quality, link)
		}
	}

	// Extract m3u8 links
	m3u8Re := regexp.MustCompile(`"hls":true.*?"link":"([^"]*)"`)
	m3u8Matches := m3u8Re.FindAllStringSubmatch(response, -1)

	for _, match := range m3u8Matches {
		if len(match) >= 2 {
			link := match[1]
			link = strings.ReplaceAll(link, "\\", "")
			links["hls"] = link
			util.Debugf("Found HLS link: %s", link)
		}
	}

	util.Debugf("Total links found: %d", len(links))
	return links
}

// selectQuality selects the appropriate quality from available links with priority consideration
func (c *AllAnimeClient) selectQuality(links map[string]string, requestedQuality string) (string, map[string]string) {
	metadata := make(map[string]string)

	// First, try to find priority links matching requested quality
	switch requestedQuality {
	case "best":
		for _, qualityLevel := range []string{"1080p", "720p", "480p", "360p"} {
			// Check for priority version first
			if url, exists := links[qualityLevel+"_priority"]; exists {
				metadata["quality"] = qualityLevel
				metadata["priority"] = "high"
				return url, metadata
			}
		}
		// Then check regular links
		for _, qualityLevel := range []string{"1080p", "720p", "480p", "360p"} {
			if url, exists := links[qualityLevel]; exists {
				metadata["quality"] = qualityLevel
				return url, metadata
			}
		}
	case "worst":
		for _, qualityLevel := range []string{"360p", "480p", "720p", "1080p"} {
			// Check for priority version first
			if url, exists := links[qualityLevel+"_priority"]; exists {
				metadata["quality"] = qualityLevel
				metadata["priority"] = "high"
				return url, metadata
			}
		}
		// Then check regular links
		for _, qualityLevel := range []string{"360p", "480p", "720p", "1080p"} {
			if url, exists := links[qualityLevel]; exists {
				metadata["quality"] = qualityLevel
				return url, metadata
			}
		}
	default:
		// Try exact match with priority first
		if url, exists := links[requestedQuality+"_priority"]; exists {
			metadata["quality"] = requestedQuality
			metadata["priority"] = "high"
			return url, metadata
		}
		// Then try exact match regular
		if url, exists := links[requestedQuality]; exists {
			metadata["quality"] = requestedQuality
			return url, metadata
		}
	}

	// Fallback to HLS if available (with priority check)
	if url, exists := links["hls_priority"]; exists {
		metadata["quality"] = "hls"
		metadata["type"] = "m3u8"
		metadata["priority"] = "high"
		return url, metadata
	}
	if url, exists := links["hls"]; exists {
		metadata["quality"] = "hls"
		metadata["type"] = "m3u8"
		return url, metadata
	}

	// Return first priority link available
	for quality, url := range links {
		if strings.HasSuffix(quality, "_priority") {
			actualQuality := strings.TrimSuffix(quality, "_priority")
			metadata["quality"] = actualQuality
			metadata["priority"] = "high"
			return url, metadata
		}
	}

	// Return first available if nothing else works
	for quality, url := range links {
		if !strings.HasSuffix(quality, "_priority") {
			metadata["quality"] = quality
			return url, metadata
		}
	}

	return "", metadata
}

// GetStreamURL implements the UnifiedScraper interface
func (c *AllAnimeClient) GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error) {
	// For AllAnime, episodeURL contains episode ID, we need anime ID and episode number
	// This is a simplified implementation - in practice you'd need to parse more context
	return "", map[string]string{}, fmt.Errorf("GetStreamURL not fully implemented for AllAnime - use GetEpisodeURL instead")
}

// GetType implements the UnifiedScraper interface
func (c *AllAnimeClient) GetType() ScraperType {
	return AllAnimeType
}
