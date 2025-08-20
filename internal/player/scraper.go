package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	//"github.com/Microsoft/go-winio"
	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
	"github.com/ktr0731/go-fuzzyfinder"
)

// DownloadFolderFormatter formats the anime URL to create a download folder name.
//
// This function extracts a specific part of the anime video URL to use it as the name
// for the download folder. It uses a regular expression to capture the part of the URL
// after "/video/", which is often unique and suitable as a folder name.
//
// Steps:
// 1. Compiles a regular expression that matches URLs of the form "https://<domain>/video/<unique-part>".
// 2. Extracts the "<unique-part>" from the URL.
// 3. If the match is successful, it returns the extracted part as the folder name.
// 4. If no match is found, it returns an empty string.
//
// Parameters:
// - str: The anime video URL as a string.
//
// Returns:
// - A string representing the formatted folder name, or an empty string if no match is found.
func DownloadFolderFormatter(str string) string {
	// Regular expression to capture the unique part after "/video/"
	regex := regexp.MustCompile(`https?://[^/]+/video/([^/?]+)`)

	// Apply the regex to the input URL
	match := regex.FindStringSubmatch(str)

	// If a match is found, return the captured group (folder name)
	if len(match) > 1 {
		finalStep := match[1]
		return finalStep
	}

	// If no match, return an empty string
	return ""
}

// getContentLength retrieves the content length of the given URL.
func getContentLength(url string, client *http.Client) (int64, error) {
	// Check if this is an AllAnime URL that might not have Content-Length header
	isAllAnimeURL := strings.Contains(url, "sharepoint.com") ||
		strings.Contains(url, "wixmp.com") ||
		strings.Contains(url, "master.m3u8") ||
		strings.Contains(url, "allanime.pro")

	// Attempts to create an HTTP HEAD request to retrieve headers without downloading the body.
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Returns 0 and the error if the request creation fails.
		return 0, err
	}

	// Sends the HEAD request to the server.
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		// If the HEAD request fails or is not supported, fall back to a GET request.
		req.Method = "GET"
		req.Header.Set("Range", "bytes=0-0") // Requests only the first byte to minimize data transfer.
		resp, err = client.Do(req)           // Sends the modified GET request.
		if err != nil {
			// Returns 0 and the error if the GET request fails.
			return 0, err
		}
	}

	// Ensures that the response body is closed after it is used to avoid resource leaks.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Logs a warning if closing the response body fails.
			util.Debugf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	// Checks if the server responded with a 200 OK or 206 Partial Content status.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		// Returns an error if the server does not support partial content (required for ranged requests).
		return 0, fmt.Errorf("server does not support partial content: status code %d", resp.StatusCode)
	}

	// Retrieves the "Content-Length" header from the response.
	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		// For AllAnime URLs that might not have Content-Length, return a default size
		if isAllAnimeURL {
			util.Debugf("Content-Length header missing for AllAnime URL, using fallback method")
			// Try to estimate content length or use a default for streaming
			return estimateContentLengthForAllAnime(url, client)
		}
		// Returns an error if the "Content-Length" header is missing for non-AllAnime URLs.
		return 0, fmt.Errorf("Content-Length header is missing")
	}

	// Converts the "Content-Length" header from a string to an int64.
	contentLength, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	if err != nil {
		// Returns 0 and an error if the conversion fails.
		return 0, err
	}

	// Returns the content length in bytes.
	return contentLength, nil
}

// estimateContentLengthForAllAnime provides a fallback method to estimate content length for AllAnime URLs
func estimateContentLengthForAllAnime(url string, client *http.Client) (int64, error) {
	// For streaming URLs (.m3u8), we can't get exact size, so return a reasonable estimate
	if strings.Contains(url, ".m3u8") {
		util.Debugf("HLS stream detected, using estimated size for download")
		// Return an estimated size for a typical episode (500MB)
		return 500 * 1024 * 1024, nil
	}

	// For other AllAnime URLs, try to get partial content to estimate size
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Request only first few KB to check response
	req.Header.Set("Range", "bytes=0-4095")
	resp, err := client.Do(req)
	if err != nil {
		// If range request fails, return default size
		util.Debugf("Range request failed, using default size estimate")
		return 300 * 1024 * 1024, nil // 300MB default
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			util.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	// Check Content-Range header for total size
	contentRange := resp.Header.Get("Content-Range")
	if contentRange != "" {
		// Parse "bytes 0-4095/12345678" format
		parts := strings.Split(contentRange, "/")
		if len(parts) == 2 && parts[1] != "*" {
			if totalSize, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return totalSize, nil
			}
		}
	}

	// Fallback to default size estimate
	util.Debugf("Could not determine exact size, using default estimate")
	return 300 * 1024 * 1024, nil // 300MB default
}

// SelectEpisodeWithFuzzyFinder allows the user to select an episode using fuzzy finder
func SelectEpisodeWithFuzzyFinder(episodes []models.Episode) (string, string, error) {
	if len(episodes) == 0 {
		return "", "", errors.New("no episodes provided")
	}

	idx, err := fuzzyfinder.Find(
		episodes,
		func(i int) string {
			return episodes[i].Number
		},
		fuzzyfinder.WithPromptString("Select the episode"),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to select episode with go-fuzzyfinder: %w", err)
	}

	if idx < 0 || idx >= len(episodes) {
		return "", "", errors.New("invalid index returned by fuzzyfinder")
	}

	return episodes[idx].URL, episodes[idx].Number, nil
}

// ExtractEpisodeNumber extracts the numeric part of an episode string
// func ExtractEpisodeNumber(episodeStr string) string {
// 	numRe := regexp.MustCompile(`(?i)epis[oó]dio\s+(\d+)`)
// 	matches := numRe.FindStringSubmatch(episodeStr)

// 	if len(matches) >= 2 {
// 		return matches[1]
// 	}
// 	return "1"
// }

// ExtractEpisodeNumber extracts the episode number from the episode string
func ExtractEpisodeNumber(episodeStr string) string {
	// Try to extract numeric episode number from various patterns
	patterns := []string{
		`(?i)epis[oó]dio\s+(\d+)`, // "Episódio 1"
		`(?i)episode\s+(\d+)`,     // "Episode 1"
		`(?i)ep\.?\s*(\d+)`,       // "Ep 1" or "Ep. 1"
		`(?i)cap[íi]tulo\s+(\d+)`, // "Capítulo 1"
		`\b(\d+)\b`,               // Any standalone number
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(episodeStr)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	// Handle simple numeric cases by splitting
	parts := strings.Split(strings.TrimSpace(episodeStr), " ")
	for _, part := range parts {
		// Clean common separators and try to parse
		cleanPart := strings.Trim(part, "()[]{}:.-_")
		if num, err := strconv.Atoi(cleanPart); err == nil && num > 0 {
			return fmt.Sprintf("%d", num)
		}
	}

	// For movies, OVAs, and specials that don't have episode numbers, return "1"
	lowerStr := strings.ToLower(episodeStr)
	if strings.Contains(lowerStr, "filme") ||
		strings.Contains(lowerStr, "movie") ||
		strings.Contains(lowerStr, "ova") ||
		strings.Contains(lowerStr, "special") {
		return "1"
	}

	// If no number found at all, return "1" as fallback (for single episodes/movies)
	return "1"
}

// GetVideoURLForEpisode gets the video URL for a given episode URL
func GetVideoURLForEpisode(episodeURL string) (string, error) {
	// Check if this looks like an AllAnime ID instead of URL
	if len(episodeURL) < 30 && !strings.Contains(episodeURL, "http") {
		return "", fmt.Errorf("GetVideoURLForEpisode called with AllAnime ID '%s' instead of HTTP URL - use enhanced API", episodeURL)
	}

	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
}

// GetVideoURLForEpisodeEnhanced gets the video URL using the enhanced API with AllAnime navigation support
func GetVideoURLForEpisodeEnhanced(episode *models.Episode, anime *models.Anime) (string, error) {
	// If we don't have anime context, decide safely how to resolve
	if anime == nil {
		// If it's a normal HTTP URL, use legacy extraction
		if strings.Contains(episode.URL, "http") {
			if util.IsDebug {
				util.Debugf("No anime context; using legacy extraction for HTTP URL, episode %s", episode.Number)
			}
			return GetVideoURLForEpisode(episode.URL)
		}

		// If episode.URL looks like an AllAnime ID, synthesize minimal anime context
		if isLikelyAllAnimeID(episode.URL) {
			if util.IsDebug {
				util.Debugf("No anime context; detected AllAnime ID '%s'. Using enhanced API with synthetic anime context.", episode.URL)
			}
			tmpAnime := &models.Anime{
				URL:    episode.URL,
				Source: "AllAnime",
				Name:   "[AllAnime]",
			}
			// Ensure episode number is set
			if episode.Number == "" && episode.Num > 0 {
				episode.Number = fmt.Sprintf("%d", episode.Num)
			}
			if episode.Number == "" {
				episode.Number = "1"
			}
			return api.GetEpisodeStreamURLEnhanced(episode, tmpAnime, util.GlobalQuality)
		}

		// If it's likely just an episode number without anime context, we cannot resolve via enhanced API
		return "", fmt.Errorf("cannot resolve stream without anime context for episode %s; missing anime identifier", episode.Number)
	}

	// Try AllAnime enhanced navigation first if applicable
	if isAllAnimeSourcePlayer(anime) {
		streamURL, err := api.GetEpisodeStreamURLEnhanced(episode, anime, util.GlobalQuality)
		if err == nil {
			return streamURL, nil
		}
	}

	// Use the regular enhanced API to get stream URL
	streamURL, err := api.GetEpisodeStreamURL(episode, anime, util.GlobalQuality)
	if err != nil {
		// Only use legacy fallback for non-AllAnime sources
		if !isAllAnimeSourcePlayer(anime) {
			return GetVideoURLForEpisode(episode.URL)
		}
		// For AllAnime, return the error instead of trying legacy method
		return "", fmt.Errorf("failed to get AllAnime stream URL: %w", err)
	}

	return streamURL, nil
}

// Helper function to check if anime is from AllAnime source (player module)
func isAllAnimeSourcePlayer(anime *models.Anime) bool {
	if anime == nil {
		return false
	}
	if anime.Source == "AllAnime" {
		return true
	}

	if strings.Contains(anime.URL, "allanime") {
		return true
	}

	if len(anime.URL) < 30 &&
		strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") &&
		!strings.Contains(anime.URL, "http") {
		return true
	}

	return false
}

// Helper: detect if a string is purely numeric (e.g., "12" or "12.5")
func isNumericString(s string) bool {
	if s == "" {
		return false
	}
	re := regexp.MustCompile(`^\d+(?:\.\d+)?$`)
	return re.MatchString(s)
}

// Helper: detect if the value looks like an AllAnime ID (short, non-HTTP, alphanumeric with letters)
func isLikelyAllAnimeID(s string) bool {
	if strings.Contains(s, "http") {
		return false
	}
	if isNumericString(s) {
		return false
	}
	// Typical AllAnime IDs are short-ish alphanumeric strings
	if len(s) >= 6 && len(s) < 30 {
		// Must contain at least one letter
		re := regexp.MustCompile(`[A-Za-z]`)
		return re.MatchString(s)
	}
	return false
}

func extractVideoURL(url string) (string, error) {
	if util.IsDebug {
		util.Debugf("Extracting video URL from page: %s", url)
	}

	response, err := api.SafeGet(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %+v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v\n", err)
		}
	}(response.Body)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %+v", err)
	}

	// Try different selectors for video elements
	selectors := []string{
		"video",
		"div[data-video-src]",
		"div[data-src]",
		"div[data-url]",
		"div[data-video]",
		"div[data-player]",
		"iframe[src*='video']",
		"iframe[src*='player']",
	}

	var videoSrc string
	var exists bool

	for _, selector := range selectors {
		elements := doc.Find(selector)
		if elements.Length() > 0 {
			if util.IsDebug {
				util.Debugf("Found elements with selector: %s", selector)
			}

			// Try different attribute names
			attributes := []string{
				"data-video-src",
				"data-src",
				"data-url",
				"data-video",
				"src",
			}

			for _, attr := range attributes {
				videoSrc, exists = elements.Attr(attr)
				if exists && videoSrc != "" {
					if util.IsDebug {
						util.Debugf("Found video URL in attribute %s: %s", attr, videoSrc)
					}
					return videoSrc, nil
				}
			}
		}
	}

	// If no video element found, try to find in page content
	if util.IsDebug {
		util.Debugf("No video elements found, searching in page content")
	}

	urlBody, err := fetchContent(url)
	if err != nil {
		return "", err
	}

	// Try to find blogger link
	videoSrc, err = findBloggerLink(urlBody)
	if err == nil && videoSrc != "" {
		if util.IsDebug {
			util.Debugf("Found blogger link: %s", videoSrc)
		}
		return videoSrc, nil
	}

	// Try to find direct video URL in content
	videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
	re := regexp.MustCompile(videoURLPattern)
	matches := re.FindString(urlBody)
	if matches != "" {
		if util.IsDebug {
			util.Debugf("Found direct video URL: %s", matches)
		}
		return matches, nil
	}

	return "", errors.New("no video source found in the page")
}

func fetchContent(url string) (string, error) {
	resp, err := api.SafeGet(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func findBloggerLink(content string) (string, error) {
	pattern := `https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 0 {
		return matches[0], nil
	} else {
		return "", errors.New("no blogger video link found in the content")
	}
}

// extractActualVideoURL processes the video source and allows the user to select quality
func extractActualVideoURL(videoSrc string) (string, error) {
	if util.IsDebug {
		util.Debugf("Processing video source: %s", videoSrc)
	}

	if strings.Contains(videoSrc, "blogger.com") {
		return videoSrc, nil
	}

	// If the URL is from animefire.plus, fetch the content
	if strings.Contains(videoSrc, "animefire.plus/video/") {
		if util.IsDebug {
			util.Debugf("Found animefire.plus video URL, fetching content...")
		}

		// Fetch the video page
		response, err := api.SafeGet(videoSrc)
		if err != nil {
			return "", fmt.Errorf("failed to fetch video page: %w", err)
		}
		defer func() {
			if err := response.Body.Close(); err != nil {
				util.Debugf("Error closing response body: %v", err)
			}
		}()

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read video page: %w", err)
		}

		// Try to parse as JSON
		var videoResponse VideoResponse
		err = json.Unmarshal(body, &videoResponse)
		if err == nil && len(videoResponse.Data) > 0 {
			if util.IsDebug {
				util.Debugf("Found video data with %d qualities", len(videoResponse.Data))
				for _, v := range videoResponse.Data {
					util.Debugf("Available quality: %s -> %s", v.Label, v.Src)
				}
			}

			// If only one quality, use it directly
			if len(videoResponse.Data) == 1 {
				return videoResponse.Data[0].Src, nil
			}

			// Use global quality preference only if it's a specific quality (not "best")
			if util.GlobalQuality != "" && util.GlobalQuality != "best" {
				selectedSrc := selectQualityFromOptions(videoResponse.Data, util.GlobalQuality)
				if selectedSrc != "" {
					if util.IsDebug {
						util.Debugf("Using global quality preference: %s -> %s", util.GlobalQuality, selectedSrc)
					}
					return selectedSrc, nil
				}
			}

			// Always prompt user for quality selection to maintain the 360p, 720p, 1080p options
			// Create options for huh.Select
			var options []huh.Option[string]
			for _, v := range videoResponse.Data {
				options = append(options, huh.NewOption(v.Label, v.Src))
			}

			// Present quality options to the user
			var selectedSrc string
			err := huh.NewSelect[string]().
				Title("Select Video Quality").
				Options(options...).
				Value(&selectedSrc).
				Run()
			if err != nil {
				return "", fmt.Errorf("failed to select quality: %w", err)
			}

			// Store the selected quality for future use in this session
			if util.GlobalQuality == "" || util.GlobalQuality == "best" {
				// Extract quality label from selected option
				for _, v := range videoResponse.Data {
					if v.Src == selectedSrc {
						util.GlobalQuality = strings.ToLower(v.Label)
						if util.IsDebug {
							util.Debugf("Storing selected quality for session: %s", util.GlobalQuality)
						}
						break
					}
				}
			}

			// Return the selected source URL
			return selectedSrc, nil
		}

		// Fallback: Try to find a direct video URL in the content
		videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
		re := regexp.MustCompile(videoURLPattern)
		matches := re.FindString(string(body))
		if matches != "" {
			if util.IsDebug {
				util.Debugf("Found direct video URL: %s", matches)
			}
			return matches, nil
		}

		// Try to find a blogger link
		videoSrc, err := findBloggerLink(string(body))
		if err == nil && videoSrc != "" {
			if util.IsDebug {
				util.Debugf("Found blogger link: %s", videoSrc)
			}
			return videoSrc, nil
		}
	}

	// If not JSON or no qualities found, return an error
	return "", errors.New("no valid video URL found")
}

// selectQualityFromOptions selects the best matching quality from available options
func selectQualityFromOptions(videoData []VideoData, preferredQuality string) string {
	if len(videoData) == 0 {
		return ""
	}

	// Normalize preferred quality string
	preferredQuality = strings.ToLower(strings.TrimSpace(preferredQuality))

	// Handle "best" quality preference
	if preferredQuality == "best" || preferredQuality == "" {
		// Find the highest quality (assume higher resolution numbers are better)
		bestQuality := videoData[0]
		for _, v := range videoData {
			if extractResolution(v.Label) > extractResolution(bestQuality.Label) {
				bestQuality = v
			}
		}
		return bestQuality.Src
	}

	// Handle "worst" quality preference
	if preferredQuality == "worst" {
		// Find the lowest quality
		worstQuality := videoData[0]
		for _, v := range videoData {
			if extractResolution(v.Label) < extractResolution(worstQuality.Label) {
				worstQuality = v
			}
		}
		return worstQuality.Src
	}

	// Try to find exact match first
	for _, v := range videoData {
		if strings.Contains(strings.ToLower(v.Label), preferredQuality) {
			return v.Src
		}
	}

	// If no exact match, try to find the closest resolution
	targetResolution := extractResolution(preferredQuality)
	if targetResolution > 0 {
		closestMatch := videoData[0]
		minDiff := abs(extractResolution(closestMatch.Label) - targetResolution)

		for _, v := range videoData {
			diff := abs(extractResolution(v.Label) - targetResolution)
			if diff < minDiff {
				minDiff = diff
				closestMatch = v
			}
		}
		return closestMatch.Src
	}

	// Fallback to first option if nothing matches
	return videoData[0].Src
}

// extractResolution extracts numeric resolution from quality label (e.g., "1080p" -> 1080)
func extractResolution(label string) int {
	re := regexp.MustCompile(`(\d+)p?`)
	matches := re.FindStringSubmatch(strings.ToLower(label))
	if len(matches) > 1 {
		if res, err := strconv.Atoi(matches[1]); err == nil {
			return res
		}
	}
	return 0
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// VideoData represents the video data structure, with a source URL and a label
type VideoData struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

// VideoResponse represents the video response structure with a slice of VideoData
type VideoResponse struct {
	Data []VideoData `json:"data"`
}
