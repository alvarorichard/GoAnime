// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// LinkPriorities defines the order of priority for link domains (from Curd project)
var LinkPriorities = []string{
	"sharepoint.com",
	"wixmp.com",
	"dropbox.com",
	"wetransfer.com",
	"gogoanime.com",
}

// GetEpisodeURL gets the streaming URL for a specific episode using priority-based selection.
//
// Transport (updated 2026-07-22 per ani-cli PR #1779):
//  1. Try the persisted-query GET path with `Origin: https://mkissa.to`.
//     This is the only path AllAnime serves `tobeparsed` blobs on.
//  2. Fall back to the legacy POST when the GET response lacks usable source data
//     (older mirrors and edge nodes still serve the legacy shape over POST).
func (c *AllAnimeClient) GetEpisodeURL(animeID, episodeNo, mode, quality string) (string, map[string]string, error) {
	if mode == "" {
		mode = "sub"
	}
	if quality == "" {
		quality = "best"
	}

	varsMap := map[string]string{
		"showId":          animeID,
		"translationType": mode,
		"episodeString":   episodeNo,
	}
	varsBytes, err := json.Marshal(varsMap)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	entries, err := c.fetchEpisodeEntries(varsBytes)
	if err != nil {
		return "", nil, err
	}
	if len(entries) == 0 {
		return "", nil, fmt.Errorf("no source URLs found for episode %s", episodeNo)
	}
	return c.processSourceEntriesConcurrent(entries, quality, animeID, episodeNo)
}

// fetchEpisodeEntries tries the persisted-query GET first, then falls back to
// the legacy POST when the GET response yields no source entries (empty body,
// stripped response, or transport failure).
func (c *AllAnimeClient) fetchEpisodeEntries(varsBytes []byte) ([]sourceEntry, error) {
	if body, err := c.tryPersistedQueryGET(varsBytes); err == nil {
		if entries := c.extractSourceEntries(string(body)); len(entries) > 0 {
			return entries, nil
		}
		util.Debugf("AllAnime GET response yielded no source entries; falling back to POST")
	} else {
		util.Debugf("AllAnime GET path failed (%v); falling back to POST", err)
	}

	body, err := c.legacyPOST(varsBytes)
	if err != nil {
		return nil, err
	}
	return c.extractSourceEntries(string(body)), nil
}

// tryPersistedQueryGET issues the Apollo persistedQuery GET request. AllAnime
// only serves the encrypted `tobeparsed` blob on this path with the mkissa.to
// Origin header and a valid aaReq token.
func (c *AllAnimeClient) tryPersistedQueryGET(varsBytes []byte) ([]byte, error) {
	keys, err := c.getAAKeys()
	if err != nil {
		return nil, fmt.Errorf("aaReq keys: %w", err)
	}
	aaReq, err := buildAAReq(allAnimePersistedQueryHash, keys)
	if err != nil {
		return nil, err
	}
	extBytes, err := json.Marshal(map[string]any{
		"persistedQuery": map[string]any{
			"version":    1,
			"sha256Hash": allAnimePersistedQueryHash,
		},
		"aaReq": aaReq,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal persisted query extensions: %w", err)
	}

	getURL := c.apiBase + "?variables=" + url.QueryEscape(string(varsBytes)) +
		"&extensions=" + url.QueryEscape(string(extBytes))

	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Origin", allAnimePersistedQueryOrigin)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("GET failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AllAnime episode URL (GET)"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read GET response: %w", err)
	}
	if err := netx.CheckHTMLResponse(resp, body, "AllAnime episode URL (GET)"); err != nil {
		return nil, err
	}
	return body, nil
}

// legacyPOST issues the original POST request against the GraphQL endpoint.
// Kept as a fallback for mirrors that still serve the old response shape.
func (c *AllAnimeClient) legacyPOST(varsBytes []byte) ([]byte, error) {
	episodeEmbedGQL := `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) { episode( showId: $showId translationType: $translationType episodeString: $episodeString ) { episodeString sourceUrls }}`

	keys, err := c.getAAKeys()
	if err != nil {
		return nil, fmt.Errorf("aaReq keys: %w", err)
	}
	aaReq, err := buildAAReq(allAnimePersistedQueryHash, keys)
	if err != nil {
		return nil, err
	}
	reqBody, err := json.Marshal(map[string]any{
		"variables": json.RawMessage(varsBytes),
		"query":     episodeEmbedGQL,
		"extensions": map[string]any{
			"aaReq": aaReq,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiBase, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("POST failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AllAnime episode URL (POST)"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read POST response: %w", err)
	}
	if err := netx.CheckHTMLResponse(resp, body, "AllAnime episode URL (POST)"); err != nil {
		return nil, err
	}
	return body, nil
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

// extractSourceEntries extracts source URLs from the API response while
// preserving the provider name. Required for filemoon dispatch — generic
// source URL extraction drops the name and routes everything through the
// default getLinks path, which silently fails for "Fm-mp4" sources.
func (c *AllAnimeClient) extractSourceEntries(response string) []sourceEntry {
	if strings.Contains(response, `"tobeparsed"`) {
		blob := extractToBeParsedBlob(response)
		if blob != "" {
			if keys, kerr := c.getAAKeys(); kerr != nil {
				util.Debugf("AllAnime aaKeys unavailable for tobeparsed decode: %v", kerr)
			} else {
				sources, err := decodeToBeParsed(blob, keys.key)
				if err == nil && len(sources) > 0 {
					util.Debugf("Decoded %d sources from tobeparsed blob", len(sources))
					entries := make([]sourceEntry, 0, len(sources))
					for _, src := range sources {
						entries = append(entries, sourceEntry{
							URL:  c.decodeSourceURL(src.sourceURL),
							Name: src.sourceName,
						})
					}
					return entries
				}
				util.Debugf("Failed to decode tobeparsed blob: %v", err)
			}
		}
	}

	var episodeResp EpisodeResponse
	if err := json.Unmarshal([]byte(response), &episodeResp); err == nil {
		entries := make([]sourceEntry, 0, len(episodeResp.Data.Episode.SourceUrls))
		for _, sourceUrl := range episodeResp.Data.Episode.SourceUrls {
			if after, ok := strings.CutPrefix(sourceUrl.SourceUrl, "--"); ok {
				entries = append(entries, sourceEntry{
					URL:  c.decodeSourceURL(after),
					Name: sourceUrl.SourceName,
				})
			} else {
				entries = append(entries, sourceEntry{
					URL:  sourceUrl.SourceUrl,
					Name: sourceUrl.SourceName,
				})
			}
		}
		if len(entries) > 0 {
			return entries
		}
	}

	matches := allAnimeSourceURLFallbackRe.FindAllStringSubmatch(response, -1)
	entries := make([]sourceEntry, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			entries = append(entries, sourceEntry{
				URL:  c.decodeSourceURL(match[1]),
				Name: match[2],
			})
		}
	}
	return entries
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
				if keys, kerr := c.getAAKeys(); kerr != nil {
					util.Debugf("AllAnime aaKeys unavailable for tobeparsed decode: %v", kerr)
				} else {
					sources, err := decodeToBeParsed(blob, keys.key)
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

	if err := netx.CheckHTTPStatus(resp, "AllAnime links"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := netx.CheckHTMLResponse(resp, body, "AllAnime links"); err != nil {
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

// filemoonResponse is the wire shape of a "Fm-mp4" source endpoint.
//
//	{"iv":"<b64url>","payload":"<b64url>","key_parts":["<b64url>","<b64url>"]}
//
// Mirrors ani-cli commit 156bf9b7. base64url is used without padding; the
// key is split across two `key_parts` entries that concatenate to a 32-byte
// AES-256 key. Payload is AES-256-CTR-encrypted (counter = iv||0x00000002)
// followed by 16 trailing bytes that are discarded — the same convention as
// the main `tobeparsed` blob.
type filemoonResponse struct {
	IV       string   `json:"iv"`
	Payload  string   `json:"payload"`
	KeyParts []string `json:"key_parts"`
}

// filemoonSources is the decrypted plaintext shape inside a filemoon payload.
type filemoonSources struct {
	Sources []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
	} `json:"sources"`
}

// getFilemoonLinks fetches and decrypts a filemoon ("Fm-mp4") source endpoint.
// AllAnime added this layer 2026-04-25; without it every Fm-mp4 source returns
// nothing through the generic JSON link parser.
func (c *AllAnimeClient) getFilemoonLinks(sourceURL string) (map[string]string, error) {
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("filemoon: failed to create request: %w", err)
	}
	req.Header.Set("Referer", c.referer)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("filemoon: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AllAnime filemoon"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("filemoon: failed to read response: %w", err)
	}
	if err := netx.CheckHTMLResponse(resp, body, "AllAnime filemoon"); err != nil {
		return nil, err
	}

	var wrapper filemoonResponse
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("filemoon: malformed JSON wrapper: %w", err)
	}
	if len(wrapper.KeyParts) < 2 {
		return nil, fmt.Errorf("filemoon: expected ≥2 key_parts, got %d", len(wrapper.KeyParts))
	}

	plaintext, err := decryptFilemoonPayload(wrapper)
	if err != nil {
		return nil, err
	}

	var parsed filemoonSources
	if err := json.Unmarshal(plaintext, &parsed); err != nil {
		return nil, fmt.Errorf("filemoon: decrypted payload is not valid JSON: %w", err)
	}

	links := make(map[string]string, len(parsed.Sources))
	for _, src := range parsed.Sources {
		if src.URL == "" || src.Height <= 0 {
			continue
		}
		links[fmt.Sprintf("%dp", src.Height)] = src.URL
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("filemoon: no usable sources after decrypt")
	}
	return c.prioritizeLinks(links), nil
}

// processSourceEntriesConcurrent processes source entries with provider-aware
// dispatch. "Fm-mp4" entries route to getFilemoonLinks; everything else uses
// the generic getLinks path. Otherwise behaves identically to
// processSourceURLsConcurrent (priority short-circuit + 6s collection window).
func (c *AllAnimeClient) processSourceEntriesConcurrent(entries []sourceEntry, quality, animeID, episodeNo string) (string, map[string]string, error) {
	type result struct {
		index     int
		links     map[string]string
		err       error
		sourceURL string
	}
	results := make(chan result, len(entries))

	type highPriorityResult struct {
		url      string
		metadata map[string]string
	}
	highPriorityCh := make(chan highPriorityResult, 1)

	for i, entry := range entries {
		go func(idx int, e sourceEntry) {
			if c.isDirectProviderURL(e.URL) {
				results <- result{
					index:     idx,
					sourceURL: e.URL,
					links:     map[string]string{"direct": e.URL},
				}
				return
			}

			var (
				links map[string]string
				err   error
			)
			if strings.EqualFold(e.Name, "Fm-mp4") {
				links, err = c.getFilemoonLinks(e.URL)
			} else {
				links, err = c.getLinks(e.URL)
			}
			if err != nil {
				results <- result{index: idx, err: err, sourceURL: e.URL}
				return
			}

			selectedURL, meta := c.selectQuality(links, quality)
			if selectedURL != "" && c.getPriorityScore(selectedURL) > 0 {
				select {
				case highPriorityCh <- highPriorityResult{url: selectedURL, metadata: meta}:
				default:
				}
			}
			results <- result{index: idx, links: links, sourceURL: e.URL}
		}(i, entry)
	}

	select {
	case hp := <-highPriorityCh:
		hp.metadata["anime_id"] = animeID
		hp.metadata["episode"] = episodeNo
		hp.metadata["priority"] = "high"
		return hp.url, hp.metadata, nil
	case <-time.After(500 * time.Millisecond):
	}

	timeout := time.After(6 * time.Second)
	processedCount := 0
	var bestURL string
	var bestMetadata map[string]string
	var firstErr error

	for processedCount < len(entries) {
		select {
		case res := <-results:
			processedCount++
			if res.err != nil {
				if firstErr == nil {
					firstErr = res.err
				}
				continue
			}
			selectedURL, metadata := c.selectQuality(res.links, quality)
			if selectedURL != "" {
				priority := c.getPriorityScore(selectedURL)
				if priority > 0 || bestURL == "" {
					bestURL = selectedURL
					bestMetadata = metadata
					bestMetadata["source_url"] = res.sourceURL
					bestMetadata["anime_id"] = animeID
					bestMetadata["episode"] = episodeNo
					if priority > 0 {
						return bestURL, bestMetadata, nil
					}
				}
			}
		case <-timeout:
			if bestURL != "" {
				return bestURL, bestMetadata, nil
			}
			if firstErr != nil {
				return "", nil, fmt.Errorf("timeout waiting for results after %d/%d sources: %w", processedCount, len(entries), firstErr)
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
