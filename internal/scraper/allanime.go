// Package scraper provides web scraping functionality for anime sources
package scraper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	AllAnimeReferer = "https://allmanga.to"
	AllAnimeBase    = "allanime.day"
	AllAnimeAPI     = "https://api.allanime.day/api"
	UserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"
	// allAnimeKeyPhrase is the passphrase used to derive the AES-256 key via SHA-256.
	// Updated 2026-04-24: AllAnime rotated the key to "Xot36i3lK3:v1".
	// Matches: printf '%s' 'Xot36i3lK3:v1' | openssl dgst -sha256 -binary
	allAnimeKeyPhrase = "Xot36i3lK3:v1"
)

// Pre-compiled regexes for AllAnime scraper (avoid per-call compilation)
var (
	allAnimeSourceURLFallbackRe = regexp.MustCompile(`"sourceUrl":"--([^"]*)".*?"sourceName":"([^"]*)"`)
	allAnimeVideoLinkRe         = regexp.MustCompile(`"link":"([^"]*)".*?"resolutionStr":"([^"]*)"`)
	allAnimeM3U8Re              = regexp.MustCompile(`"hls":true.*?"link":"([^"]*)"`)
)

// AllAnimeClient handles interactions with AllAnime API
type AllAnimeClient struct {
	client    *http.Client
	referer   string
	apiBase   string
	userAgent string
}

// allAnimeClientInstance is a singleton for connection reuse
var (
	allAnimeClientInstance *AllAnimeClient
	allAnimeClientOnce     sync.Once
)

// allAnimeKey returns the AES-256 key derived from SHA-256(allAnimeKeyPhrase).
// Computed once and cached.
var allAnimeKey = func() []byte {
	h := sha256.Sum256([]byte(allAnimeKeyPhrase))
	return h[:]
}()

// sourceInfo holds a decoded source URL and its provider name.
type sourceInfo struct {
	sourceName string
	sourceURL  string
}

// decodeToBeParsed decrypts the "tobeparsed" blob from the AllAnime API.
//
// Blob format (updated 2026-04-24): [1-byte version 0x01][12-byte nonce][ciphertext][16-byte GCM tag]
// The cipher is AES-256-GCM; the key is SHA-256("Xot36i3lK3:v1").
// Minimum valid size: 1 + 12 + 0 + 16 = 29 bytes (empty plaintext), so we require ≥ 30.
//
// ani-cli reference: https://github.com/pystardust/ani-cli/commit/e5523a9b480f67ee878a0cc075043313cc58e07d
func decodeToBeParsed(blob string) ([]sourceInfo, error) {
	util.Debugf("AllAnime tobeparsed raw blob (first 60 chars): %q", blob[:min(60, len(blob))])

	// Try standard base64 first, then URL-safe (AllAnime may use either)
	data, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(blob)
		if err != nil {
			data, err = base64.RawURLEncoding.DecodeString(blob)
			if err != nil {
				return nil, fmt.Errorf("base64 decode failed: %w", err)
			}
		}
	}

	util.Debugf("AllAnime tobeparsed decoded length: %d bytes, first 16 bytes: %x", len(data), data[:min(16, len(data))])

	// 1 (version) + 12 (nonce) + 16 (GCM tag) + at least 1 byte plaintext = 30
	if len(data) < 30 {
		return nil, fmt.Errorf("tobeparsed blob too short (%d bytes)", len(data))
	}

	// Blob format: [0x01 version byte][12-byte nonce][ciphertext + 16-byte GCM tag]
	nonce := data[1:13]
	ciphertextWithTag := data[13:]
	util.Debugf("AllAnime tobeparsed nonce: %x", nonce)

	block, err := aes.NewCipher(allAnimeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	// Parse the decrypted JSON to extract sourceUrl/sourceName pairs.
	// The bash script does:
	//   sed -nE 's|.*"sourceUrl":"--([^"]*)".*"sourceName":"([^"]*)".*|\2 :\1|p'
	// We parse each sourceUrls entry from the JSON structure.
	var result struct {
		Data struct {
			Episode struct {
				SourceUrls []struct {
					SourceURL  string `json:"sourceUrl"`
					SourceName string `json:"sourceName"`
				} `json:"sourceUrls"`
			} `json:"episode"`
		} `json:"data"`
	}

	util.Debugf("AllAnime tobeparsed decrypted (first 200 bytes): %q", string(plaintext[:min(200, len(plaintext))]))

	// The plaintext may contain the full GraphQL response or just the sourceUrls array.
	// Try parsing as the full response first.
	if err := json.Unmarshal(plaintext, &result); err == nil && len(result.Data.Episode.SourceUrls) > 0 {
		var sources []sourceInfo
		for _, su := range result.Data.Episode.SourceUrls {
			url := su.SourceURL
			url = strings.TrimPrefix(url, "--")
			sources = append(sources, sourceInfo{
				sourceName: su.SourceName,
				sourceURL:  url,
			})
		}
		return sources, nil
	}

	// Fallback: try to extract using regex (like the bash sed pattern).
	// The plaintext might not be perfectly structured JSON.
	re := regexp.MustCompile(`"sourceUrl"\s*:\s*"--([^"]*)"[^}]*"sourceName"\s*:\s*"([^"]*)"`)
	matches := re.FindAllSubmatch(plaintext, -1)
	if len(matches) == 0 {
		// Also try reverse order (sourceName before sourceUrl)
		re2 := regexp.MustCompile(`"sourceName"\s*:\s*"([^"]*)"[^}]*"sourceUrl"\s*:\s*"--([^"]*)"`)
		matches = re2.FindAllSubmatch(plaintext, -1)
		if len(matches) == 0 {
			return nil, fmt.Errorf("no source URLs found in decrypted tobeparsed data")
		}
		var sources []sourceInfo
		for _, m := range matches {
			sources = append(sources, sourceInfo{
				sourceName: string(m[1]),
				sourceURL:  string(m[2]),
			})
		}
		return sources, nil
	}

	var sources []sourceInfo
	for _, m := range matches {
		sources = append(sources, sourceInfo{
			sourceName: string(m[2]),
			sourceURL:  string(m[1]),
		})
	}
	return sources, nil
}

// NewAllAnimeClient creates a new AllAnime client (returns cached instance for connection reuse)
func NewAllAnimeClient() *AllAnimeClient {
	allAnimeClientOnce.Do(func() {
		allAnimeClientInstance = &AllAnimeClient{
			client:    util.NewFastClient(), // Own client to avoid http2 transport race
			referer:   AllAnimeReferer,
			apiBase:   AllAnimeAPI,
			userAgent: UserAgent,
		}
	})
	return allAnimeClientInstance
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

// SearchAnime searches for anime using AllAnime API (based on Curd implementation)
func (c *AllAnimeClient) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
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
	variables := map[string]any{
		"search": map[string]any{
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

	// Build the POST request body with variables and query
	reqBody, err := json.Marshal(map[string]any{
		"variables": json.RawMessage(variablesJSON),
		"query":     searchGql,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiBase, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkHTTPStatus(resp, "AllAnime search"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := checkHTMLResponse(resp, body, "AllAnime search"); err != nil {
		return nil, err
	}

	// Parse using a simple structure like Curd
	var response struct {
		Data struct {
			Shows struct {
				Edges []struct {
					ID                string `json:"_id"`
					Name              string `json:"name"`
					EnglishName       string `json:"englishName"`
					AvailableEpisodes any    `json:"availableEpisodes"`
				} `json:"edges"`
			} `json:"shows"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Collect results with episode counts for sorting
	type searchResult struct {
		anime   *models.Anime
		epCount int
	}
	var results []searchResult

	for _, edge := range response.Data.Shows.Edges {
		var episodesStr string
		var epCount int
		if episodes, ok := edge.AvailableEpisodes.(map[string]any); ok {
			if subEpisodes, ok := episodes["sub"].(float64); ok {
				epCount = int(subEpisodes)
				episodesStr = fmt.Sprintf("(%d episodes)", epCount)
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
		results = append(results, searchResult{anime: anime, epCount: epCount})
	}

	// Sort by episode count descending so the main series (most episodes) appears first
	sort.Slice(results, func(i, j int) bool {
		return results[i].epCount > results[j].epCount
	})

	animes := make([]*models.Anime, len(results))
	for i, r := range results {
		animes[i] = r.anime
	}

	return animes, nil
}

// GetEpisodesList gets the list of available episodes for an anime (based on Curd implementation)
func (c *AllAnimeClient) GetEpisodesList(animeID, mode string) ([]string, error) {
	if mode == "" {
		mode = "sub"
	}

	episodesListGql := `query ($showId: String!) { show( _id: $showId ) { _id availableEpisodesDetail }}`

	// Use json.Marshal to safely build the variables JSON, preventing injection
	varsBytes, err := json.Marshal(map[string]string{"showId": animeID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	// Build the POST request body
	reqBody, err := json.Marshal(map[string]any{
		"variables": json.RawMessage(varsBytes),
		"query":     episodesListGql,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiBase, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.referer)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkHTTPStatus(resp, "AllAnime episodes list"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := checkHTMLResponse(resp, body, "AllAnime episodes list"); err != nil {
		return nil, err
	}

	// Use the same response structure as Curd
	var response struct {
		Data struct {
			Show struct {
				ID                      string         `json:"_id"`
				AvailableEpisodesDetail map[string]any `json:"availableEpisodesDetail"`
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

// extractEpisodes extracts the episodes list from the availableEpisodesDetail field (from Curd)
func extractEpisodes(availableEpisodesDetail map[string]any, mode string) []string {
	var episodes []float64

	// Check if the mode (e.g., "sub") exists in the map
	if eps, ok := availableEpisodesDetail[mode].([]any); ok {
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
func (c *AllAnimeClient) SendSkipTimesToMPV(episode *models.Episode, socketPath string, mpvSendCommand func(string, []any) (any, error)) error {
	// Only proceed if we have valid skip times
	if episode.SkipTimes.Op.Start == 0 && episode.SkipTimes.Op.End == 0 &&
		episode.SkipTimes.Ed.Start == 0 && episode.SkipTimes.Ed.End == 0 {
		return fmt.Errorf("no skip times available for episode")
	}

	// Create chapter list exactly like Curd does
	chapterList := []map[string]any{}

	// Pre-Opening chapter
	if episode.SkipTimes.Op.Start > 0 {
		chapterList = append(chapterList, map[string]any{
			"title": "Pre-Opening",
			"time":  0.0,
			"end":   float64(episode.SkipTimes.Op.Start),
		})
	}

	// Opening chapter
	if episode.SkipTimes.Op.Start > 0 && episode.SkipTimes.Op.End > episode.SkipTimes.Op.Start {
		chapterList = append(chapterList, map[string]any{
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
		chapterList = append(chapterList, map[string]any{
			"title": "Main",
			"time":  mainStart,
		})
	} else {
		chapterList = append(chapterList, map[string]any{
			"title": "Main",
			"time":  mainStart,
			"end":   mainEnd,
		})
	}

	// Ending chapter
	if episode.SkipTimes.Ed.Start > 0 && episode.SkipTimes.Ed.End > episode.SkipTimes.Ed.Start {
		chapterList = append(chapterList, map[string]any{
			"title": "Ending",
			"time":  float64(episode.SkipTimes.Ed.Start),
			"end":   float64(episode.SkipTimes.Ed.End),
		})
	}

	// Post-Credits chapter
	if episode.SkipTimes.Ed.End > 0 {
		chapterList = append(chapterList, map[string]any{
			"title": "Post-Credits",
			"time":  float64(episode.SkipTimes.Ed.End),
		})
	}

	// Send chapter list to MPV exactly like Curd does
	_, err := mpvSendCommand(socketPath, []any{
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

// parseEpisodeNum converts episode string to integer
func parseEpisodeNum(epStr string) int {
	// Try to extract number from string
	var num int
	_, err := fmt.Sscanf(epStr, "%d", &num)
	if err != nil || num == 0 {
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
func (c *AllAnimeClient) GetEpisodeURL(animeID, episodeNo, mode, quality string) (string, map[string]string, error) {
	if mode == "" {
		mode = "sub"
	}
	if quality == "" {
		quality = "best"
	}

	episodeEmbedGQL := `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) { episode( showId: $showId translationType: $translationType episodeString: $episodeString ) { episodeString sourceUrls }}`

	// Use json.Marshal to safely build the variables JSON, preventing injection
	varsMap := map[string]string{
		"showId":          animeID,
		"translationType": mode,
		"episodeString":   episodeNo,
	}
	varsBytes, err := json.Marshal(varsMap)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	// Build the POST request body
	reqBody, err := json.Marshal(map[string]any{
		"variables": json.RawMessage(varsBytes),
		"query":     episodeEmbedGQL,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiBase, bytes.NewReader(reqBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkHTTPStatus(resp, "AllAnime episode URL"); err != nil {
		return "", nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := checkHTMLResponse(resp, body, "AllAnime episode URL"); err != nil {
		return "", nil, err
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
func (c *AllAnimeClient) processSourceURLsConcurrent(sourceURLs []string, quality, animeID, episodeNo string) (string, map[string]string, error) {
	type result struct {
		index     int
		links     map[string]string
		err       error
		sourceURL string
	}

	results := make(chan result, len(sourceURLs))

	// Launch goroutines for concurrent processing
	type highPriorityResult struct {
		url      string
		metadata map[string]string
	}
	highPriorityCh := make(chan highPriorityResult, 1)

	for i, sourceURL := range sourceURLs {
		go func(idx int, url string) {
			if c.isDirectProviderURL(url) {
				results <- result{
					index:     idx,
					sourceURL: url,
					links: map[string]string{
						"direct": url,
					},
				}
				return
			}

			links, err := c.getLinks(url)
			if err != nil {
				results <- result{index: idx, err: err, sourceURL: url}
				return
			}

			// Check for high priority links, but run them through quality
			// selection so we don't accidentally grab a low-res variant.
			selectedURL, meta := c.selectQuality(links, quality)
			if selectedURL != "" && c.getPriorityScore(selectedURL) > 0 {
				select {
				case highPriorityCh <- highPriorityResult{url: selectedURL, metadata: meta}:
				default:
				}
			}

			results <- result{index: idx, links: links, sourceURL: url}
		}(i, sourceURL)
	}

	// First, try to get a high priority link quickly
	select {
	case hp := <-highPriorityCh:
		// Found high priority link with proper quality selection
		hp.metadata["anime_id"] = animeID
		hp.metadata["episode"] = episodeNo
		hp.metadata["priority"] = "high"
		return hp.url, hp.metadata, nil
	case <-time.After(500 * time.Millisecond): // Wait briefly for high priority link
		// No high priority link found quickly, proceed with normal collection
	}

	// Collect results with timeout
	timeout := time.After(6 * time.Second)
	processedCount := 0
	var bestURL string
	var bestMetadata map[string]string
	var firstErr error

	for processedCount < len(sourceURLs) {
		select {
		case res := <-results:
			processedCount++
			if res.err != nil {
				if firstErr == nil {
					firstErr = res.err
				}
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

		case <-timeout:
			if bestURL != "" {
				return bestURL, bestMetadata, nil
			}
			if firstErr != nil {
				return "", nil, fmt.Errorf("timeout waiting for results after %d/%d sources: %w", processedCount, len(sourceURLs), firstErr)
			}
			return "", nil, fmt.Errorf("timeout waiting for results")
		}
	}

	if bestURL != "" {
		return bestURL, bestMetadata, nil
	}

	if firstErr != nil {
		return "", nil, fmt.Errorf("no suitable quality found from any source: %w", firstErr)
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

func (c *AllAnimeClient) isDirectProviderURL(sourceURL string) bool {
	return strings.Contains(sourceURL, "tools.fast4speed.rsvp")
}

// extractSourceURLs extracts source URLs from the API response
func (c *AllAnimeClient) extractSourceURLs(response string) []string {
	// Check if the response contains a "tobeparsed" blob (AES-encrypted source URLs).
	// This matches the bash script: if printf "%s" "$api_resp" | grep -q '"tobeparsed"'; then ...
	var rawResp map[string]json.RawMessage
	if err := json.Unmarshal([]byte(response), &rawResp); err == nil {
		// Look for "tobeparsed" at any level of the response
		if strings.Contains(response, `"tobeparsed"`) {
			blob := extractToBeParsedBlob(response)
			if blob != "" {
				sources, err := decodeToBeParsed(blob)
				if err == nil && len(sources) > 0 {
					util.Debugf("Decoded %d sources from tobeparsed blob", len(sources))
					var urls []string
					for _, src := range sources {
						decoded := c.decodeSourceURL(src.sourceURL)
						urls = append(urls, decoded)
					}
					return urls
				}
				util.Debugf("Failed to decode tobeparsed blob: %v", err)
			}
		}
	}

	// Standard path: parse the JSON response to extract sourceUrls
	var episodeResp EpisodeResponse
	if err := json.Unmarshal([]byte(response), &episodeResp); err == nil {
		var urls []string
		for _, sourceUrl := range episodeResp.Data.Episode.SourceUrls {
			if after, ok := strings.CutPrefix(sourceUrl.SourceUrl, "--"); ok {
				// This is an encoded URL that needs decoding
				decoded := c.decodeSourceURL(after)
				urls = append(urls, decoded)
			} else {
				// Direct URL
				urls = append(urls, sourceUrl.SourceUrl)
			}
		}
		if len(urls) > 0 {
			return urls
		}
	}

	// Fallback to regex-based extraction if JSON parsing fails
	matches := allAnimeSourceURLFallbackRe.FindAllStringSubmatch(response, -1)

	var urls []string
	for _, match := range matches {
		if len(match) >= 2 {
			decodedURL := c.decodeSourceURL(match[1])
			urls = append(urls, decodedURL)
		}
	}

	return urls
}

// extractToBeParsedBlob extracts the base64 "tobeparsed" value from the API response JSON.
func extractToBeParsedBlob(response string) string {
	re := regexp.MustCompile(`"tobeparsed"\s*:\s*"([^"]*)"`)
	match := re.FindStringSubmatch(response)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// hexSubstitutionTable is the complete hex-pair substitution cipher from ani-cli's provider_init.
// Each 2-char hex pair maps to its decoded ASCII character.
var hexSubstitutionTable = map[string]string{
	// Uppercase letters
	"79": "A", "7a": "B", "7b": "C", "7c": "D", "7d": "E", "7e": "F", "7f": "G",
	"70": "H", "71": "I", "72": "J", "73": "K", "74": "L", "75": "M", "76": "N", "77": "O",
	"68": "P", "69": "Q", "6a": "R", "6b": "S", "6c": "T", "6d": "U", "6e": "V", "6f": "W",
	"60": "X", "61": "Y", "62": "Z",
	// Lowercase letters
	"59": "a", "5a": "b", "5b": "c", "5c": "d", "5d": "e", "5e": "f", "5f": "g",
	"50": "h", "51": "i", "52": "j", "53": "k", "54": "l", "55": "m", "56": "n", "57": "o",
	"48": "p", "49": "q", "4a": "r", "4b": "s", "4c": "t", "4d": "u", "4e": "v", "4f": "w",
	"40": "x", "41": "y", "42": "z",
	// Digits
	"08": "0", "09": "1", "0a": "2", "0b": "3", "0c": "4", "0d": "5", "0e": "6", "0f": "7",
	"00": "8", "01": "9",
	// Special characters
	"15": "-", "16": ".", "67": "_", "46": "~",
	"02": ":", "17": "/", "07": "?", "1b": "#",
	"63": "[", "65": "]", "78": "@",
	"19": "!", "1c": "$", "1e": "&",
	"10": "(", "11": ")", "12": "*", "13": "+", "14": ",",
	"03": ";", "05": "=", "1d": "%",
}

// decodeSourceURL decodes the encoded source URL using the hex substitution cipher from ani-cli
func (c *AllAnimeClient) decodeSourceURL(encoded string) string {
	// Split into 2-char hex pairs and substitute
	var result strings.Builder
	result.Grow(len(encoded))
	for i := 0; i+1 < len(encoded); i += 2 {
		pair := encoded[i : i+2]
		if val, exists := hexSubstitutionTable[pair]; exists {
			result.WriteString(val)
		} else {
			result.WriteString(pair)
		}
	}

	decoded := result.String()

	// Replace "/clock" with "/clock.json" like in ani-cli
	decoded = strings.ReplaceAll(decoded, "/clock", "/clock.json")

	// If it starts with /, it's a path that needs the AllAnime base
	if strings.HasPrefix(decoded, "/") {
		decoded = fmt.Sprintf("https://%s%s", AllAnimeBase, decoded)
	}

	return decoded
}

// getLinks extracts video links from the source URL with proper headers
func (c *AllAnimeClient) getLinks(sourceURL string) (map[string]string, error) {
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Match ani-cli: AllAnime's current providers expect the allmanga referer.
	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0")

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkHTTPStatus(resp, "AllAnime links"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := checkHTMLResponse(resp, body, "AllAnime links"); err != nil {
		return nil, err
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
	maps.Copy(prioritizedLinks, links)

	return prioritizedLinks
}

// extractVideoLinks extracts video links from the response with debug logging
func (c *AllAnimeClient) extractVideoLinks(response string) map[string]string {
	links := make(map[string]string)

	// Debug: log response structure
	util.Debugf("Response length: %d", len(response))

	// Parse JSON response
	var jsonData map[string]any
	if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
		// Extract links from JSON structure
		if linksInterface, ok := jsonData["links"].([]any); ok {
			for _, linkInterface := range linksInterface {
				if linkMap, ok := linkInterface.(map[string]any); ok {
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
	matches := allAnimeVideoLinkRe.FindAllStringSubmatch(response, -1)

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
	m3u8Matches := allAnimeM3U8Re.FindAllStringSubmatch(response, -1)

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

	if url, exists := links["direct"]; exists {
		metadata["quality"] = "direct"
		metadata["type"] = "direct"
		metadata["referer"] = c.referer
		return url, metadata
	}

	// Return first priority link available
	for quality, url := range links {
		if before, ok := strings.CutSuffix(quality, "_priority"); ok {
			actualQuality := before
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
func (c *AllAnimeClient) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// For AllAnime, episodeURL contains episode ID, we need anime ID and episode number
	// This is a simplified implementation - in practice you'd need to parse more context
	return "", map[string]string{}, fmt.Errorf("GetStreamURL not fully implemented for AllAnime - use GetEpisodeURL instead")
}

// GetType implements the UnifiedScraper interface
func (c *AllAnimeClient) GetType() ScraperType {
	return AllAnimeType
}
