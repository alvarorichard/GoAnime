package player

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"

	// Blank import: the providers package self-registers every live Source in
	// its init(). Without it the registry is empty and dispatch resolves
	// nothing. Consolidated into a single wiring file in a later phase (S3).
	_ "github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
	g "github.com/enetx/g"
	"github.com/enetx/surf"
)

// Pre-compiled regexes for player scraper (avoid per-call compilation)
var (
	downloadFolderRe    = regexp.MustCompile(`https?://[^/]+/video/([^/?]+)`)
	videoURLPatternRe   = regexp.MustCompile(`https?://[^\s<>"]+?\.(?:mp4|m3u8)`)
	bloggerPatternRe    = regexp.MustCompile(`^https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)$`)
	tokenRe             = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)
	sidRe               = regexp.MustCompile(`"FdrFJe"\s*:\s*"([^"]+)"`)
	bhRe                = regexp.MustCompile(`"cfb2h"\s*:\s*"([^"]+)"`)
	atRe                = regexp.MustCompile(`"SNlM0e"\s*:\s*"([^"]+)"`)
	extractResolutionRe = regexp.MustCompile(`(\d+)p?`)
	googleVideoRe       = regexp.MustCompile(`https://[^"\\]+\.googlevideo\.com/[^"\\]+`)
	episodePatternREs   = []*regexp.Regexp{
		regexp.MustCompile(`(?i)epis[oó]dio\s+(\d+)`),
		regexp.MustCompile(`(?i)episode\s+(\d+)`),
		regexp.MustCompile(`(?i)ep\.?\s*(\d+)`),
		regexp.MustCompile(`(?i)cap[íi]tulo\s+(\d+)`),
		regexp.MustCompile(`\b(\d+)\b`),
	}
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
	match := downloadFolderRe.FindStringSubmatch(str)

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
	// Check if this is a URL that might not have Content-Length header
	isKnownStreamURL := strings.Contains(url, "sharepoint.com") ||
		strings.Contains(url, "wixmp.com") ||
		strings.Contains(url, "master.m3u8") ||
		strings.Contains(url, ".m3u8") ||
		strings.Contains(url, "allanime.pro") ||
		strings.Contains(url, "allanime.day") ||
		strings.Contains(url, "animefire") ||
		strings.Contains(url, "blogger.com") ||
		strings.Contains(url, "animesfire") ||
		strings.Contains(url, "repackager.wixmp.com")

	// Attempts to create an HTTP HEAD request to retrieve headers without downloading the body.
	req, err := http.NewRequest("HEAD", url, http.NoBody)
	if err != nil {
		// Returns 0 and the error if the request creation fails.
		return 0, err
	}
	applyDownloadAuthHeaders(req, url)

	// Sends the HEAD request to the server.
	resp, err := client.Do(req) // #nosec G704
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				util.Debugf("Failed to close HEAD response body: %v", closeErr)
			}
		}
		// If the HEAD request fails or is not supported, fall back to a GET request.
		req.Method = "GET"
		req.Header.Set("Range", "bytes=0-0") // Requests only the first byte to minimize data transfer.
		resp, err = client.Do(req)           // #nosec G704 -- Sends the modified GET request.
		if err != nil {
			// Returns 0 and the error if the GET request fails.
			return 0, err
		}
	}

	// Ensures that the response body is closed after it is used to avoid resource leaks.
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			// Logs a warning if closing the response body fails.
			util.Debugf("Failed to close response body: %v", err)
		}
	}()

	// Checks if the server responded with a 200 OK or 206 Partial Content status.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return 0, netx.NewDownloadExpiredError("Download", "content-length", resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
		}
		// Returns an error if the server does not support partial content (required for ranged requests).
		return 0, fmt.Errorf("server does not support partial content: status code %d", resp.StatusCode)
	}

	// Retrieves the "Content-Length" header from the response.
	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		// For known streaming URLs that might not have Content-Length, return a default size
		if isKnownStreamURL {
			util.Debugf("Content-Length header missing for streaming URL, using fallback method")
			return estimateContentLengthForAllAnime(url, client)
		}
		// For any other URL without Content-Length, use a reasonable default instead of failing
		util.Debugf("Content-Length header missing, using default estimate")
		return 200 * 1024 * 1024, nil // 200MB default
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
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return 0, err
	}

	// Request only first few KB to check response
	req.Header.Set("Range", "bytes=0-4095")
	resp, err := client.Do(req) // #nosec G704
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

// ErrBackRequested is returned when user selects the back option
var ErrBackRequested = errors.New("back requested")

// ErrBackToAnimeSelection is returned when user wants to go back to anime selection
var ErrBackToAnimeSelection = errors.New("back to anime selection")

// ErrBackToEpisodeSelection is returned when user wants to go back to episode selection
var ErrBackToEpisodeSelection = errors.New("back to episode selection")

// episodeDisplayTitle picks the best available episode title.
func episodeDisplayTitle(ep models.Episode) string {
	if ep.Title.Romaji != "" {
		return ep.Title.Romaji
	}
	return ep.Title.English
}

// episodeLabel builds the picker row label: "Episode 12 — Title".
func episodeLabel(ep models.Episode) string {
	title := episodeDisplayTitle(ep)
	if title == "" {
		return ep.Number
	}
	return ep.Number + " — " + title
}

// episodeDetails builds the secondary row: air date and filler/recap badges.
func episodeDetails(ep models.Episode) string {
	parts := make([]string, 0, 3)
	if aired := strings.TrimSpace(ep.Aired); aired != "" {
		parts = append(parts, aired)
	}
	if ep.IsFiller {
		parts = append(parts, "Filler")
	}
	if ep.IsRecap {
		parts = append(parts, "Recap")
	}
	return strings.Join(parts, "  •  ")
}

// episodePickItems converts episodes to generic picker rows.
func episodePickItems(episodes []models.Episode) []tui.PickItem {
	items := make([]tui.PickItem, len(episodes))
	for i, ep := range episodes {
		items[i] = tui.PickItem{
			Label:   episodeLabel(ep),
			Details: episodeDetails(ep),
		}
	}
	return items
}

// episodePickFunc matches tui.Pick so tests can inject a headless picker.
type episodePickFunc func([]tui.PickItem, tui.PickOptions) (int, error)

// SelectEpisodeWithFuzzyFinder lets the user pick an episode on the styled
// fuzzy picker. ESC (or quit) is reported as ErrBackRequested.
func SelectEpisodeWithFuzzyFinder(episodes []models.Episode) (episodeURL, episodeNumber string, err error) {
	return selectEpisodeWithPicker(tui.Pick, episodes)
}

// selectEpisodeWithPicker isolates picker execution for deterministic tests.
func selectEpisodeWithPicker(pick episodePickFunc, episodes []models.Episode) (episodeURL, episodeNumber string, err error) {
	if len(episodes) == 0 {
		return "", "", errors.New("no episodes provided")
	}
	if pick == nil {
		return "", "", errors.New("episode picker not configured")
	}

	util.Debugf("[TRACE] SelectEpisodeWithFuzzyFinder: opening picker with %d episodes", len(episodes))
	idx, err := pick(episodePickItems(episodes), tui.PickOptions{
		Breadcrumb:   "Search > Episodes",
		WindowTitle:  "GoAnime - Episodes",
		ItemSingular: "episode",
		ItemPlural:   "episodes",
	})
	util.Debugf("[TRACE] SelectEpisodeWithFuzzyFinder: picker returned idx=%d, err=%v", idx, err)
	if err != nil {
		// Treat back/quit as a back request so callers return to the
		// previous screen instead of failing.
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return "", "", ErrBackRequested
		}
		return "", "", fmt.Errorf("failed to select episode: %w", err)
	}

	if idx < 0 || idx >= len(episodes) {
		return "", "", errors.New("invalid index returned by episode picker")
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
	for _, re := range episodePatternREs {
		matches := re.FindStringSubmatch(episodeStr)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	// Handle simple numeric cases by splitting
	parts := strings.SplitSeq(strings.TrimSpace(episodeStr), " ")
	for part := range parts {
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

// GetVideoURLForEpisodeEnhanced gets the video URL using the enhanced API with AllAnime navigation support.
// ctx is honored at entry today; full downstream propagation lands when this
// becomes a thin wrapper over source.Resolve → FetchStreamURL(ctx, …).
func GetVideoURLForEpisodeEnhanced(ctx context.Context, episode *models.Episode, anime *models.Anime) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.Debug("GetVideoURLForEpisodeEnhanced called", "episodeURL", episode.URL, "episodeNum", episode.Number)
	if anime != nil {
		util.Debug("Anime context", "name", anime.Name, "source", anime.Source, "mediaType", anime.MediaType, "url", anime.URL)
	}

	// If we don't have anime context, decide safely how to resolve
	if anime == nil {
		// If it's a normal HTTP URL, use legacy extraction
		if strings.Contains(episode.URL, "http") {
			if util.IsDebug {
				util.Debugf("No anime context; using legacy extraction for HTTP URL, episode %s", episode.Number)
			}
			return GetVideoURLForEpisode(episode.URL)
		}

		// URL-only resolution goes through the source registry — no more
		// hardcoded fake-AllAnime guess (R3). The minimal anime context is
		// derived from what the registry itself matched, never assumed.
		src, resolved := source.ResolveURL(episode.URL)
		if src == nil {
			return "", fmt.Errorf("cannot resolve stream without anime context for episode %s; missing anime identifier", episode.Number)
		}
		util.Debug("URL-only source resolved", "kind", resolved.Kind, "reason", resolved.Reason)
		urlOnlyAnime := &models.Anime{
			URL:    episode.URL,
			Source: string(resolved.Kind),
		}
		// Ensure episode number is set
		if episode.Number == "" && episode.Num > 0 {
			episode.Number = fmt.Sprintf("%d", episode.Num)
		}
		if episode.Number == "" {
			episode.Number = "1"
		}
		return src.FetchStreamURL(ctx, episode, urlOnlyAnime, util.GlobalQuality)
	}

	// ── Model B dispatch: resolve once via the source registry, then fetch
	// through the resolved Source (each Source delegates to the same api
	// functions the legacy chain called, so behavior is unchanged). The old
	// helpers survive below only as the transitional error/extraction policy;
	// Phase 3 deletes them together with the api-level branching.
	src, resolved := source.Resolve(anime)
	if src == nil {
		// Unknown source at the dispatch boundary — never silent (R4/R5):
		// warn loudly, and only fall back to best-effort AllAnime when the
		// user hasn't opted into strict resolution.
		if util.StrictSourceResolution() {
			return "", fmt.Errorf("unrecognized source for %q (%s); best-effort fallback disabled by GOANIME_STRICT_SOURCE", anime.Name, resolved.Reason)
		}
		util.Warn("unrecognized source; dispatching best-effort AllAnime (set GOANIME_STRICT_SOURCE=1 to fail instead)",
			"anime", anime.Name, "url", anime.URL, "reason", resolved.Reason)
		bestEffort, ok := source.Enabled(resolved.BestEffortKind())
		if !ok {
			return "", fmt.Errorf("no enabled source for %q (%s); it may be turned off via GOANIME_DISABLED_SOURCES", resolved.BestEffortKind(), resolved.Reason)
		}
		src = bestEffort
	}
	util.Debug("Source resolved", "kind", src.Describe().Kind,
		"reason", resolved.Reason, "seasoned", source.IsSeasoned(src), "browserGated", source.IsBrowserGated(src))

	// Model C capability: warm up a browser-gated source before fetching. For a
	// non-gated source this is an explicit (logged) no-op; for SuperFlix on a
	// display-less box it fails fast with a clear reason instead of stalling on
	// a browser that can never appear.
	if err := source.WarmUp(ctx, src); err != nil {
		return "", err
	}

	streamURL, err := src.FetchStreamURL(ctx, episode, anime, util.GlobalQuality)
	if err != nil {
		// Transitional error policy — mirrors the legacy chain exactly.
		if isMovieOrTVSourcePlayer(anime) {
			sourceLabel := anime.Source
			if sourceLabel == "" {
				sourceLabel = "movie/TV"
			}
			util.Debug("Movie/TV stream URL failed", "source", sourceLabel, "error", err)
			return "", fmt.Errorf("failed to get %s stream URL: %w", sourceLabel, err)
		}
		if resolved.Kind == source.AllAnime {
			// For AllAnime, return the error instead of trying legacy method
			return "", fmt.Errorf("failed to get AllAnime stream URL: %w", err)
		}
		// Legacy silent fallback for the remaining sources — removed in Phase 2.
		return GetVideoURLForEpisode(episode.URL)
	}

	// Movie/TV URLs are returned as-is, exactly as the legacy chain did.
	if isMovieOrTVSourcePlayer(anime) {
		util.Debug("Movie/TV stream URL obtained", "source", anime.Source, "url", streamURL)
		return streamURL, nil
	}

	// The enhanced API may return intermediate URLs (Blogger embeds, AnimeFire
	// video JSON API) that are not directly playable. Resolve them to actual
	// CDN video URLs before returning.
	if needsVideoExtraction(streamURL) {
		resolved, err := extractActualVideoURL(streamURL)
		if err == nil && resolved != "" {
			return resolved, nil
		}
		// A dead Blogger token will never become live again — bubble the error
		// up so the caller routes the user back to episode selection instead
		// of letting the player layer redundantly resolve the same dead URL.
		if errors.Is(err, errBloggerVideoUnavailable) {
			return "", fmt.Errorf("video unavailable on this source: %w", err)
		}
		// If resolution failed for an unknown reason, fall back to the original
		// URL so yt-dlp can have a chance.
		util.Debug("Could not resolve intermediate URL, using as-is", "url", streamURL, "err", err)
	}

	return streamURL, nil
}

// Helper function to check if anime is from FlixHQ source (player module)
// isMovieOrTVSourcePlayer routes any movie/TV content through the enhanced API,
// which dispatches by anime.Source (SuperFlix, FlixHQ, ...). Despite the legacy
// name, this is not FlixHQ-specific — SuperFlix selections also reach this
// branch because their MediaType is MediaTypeMovie/MediaTypeTV.
func isMovieOrTVSourcePlayer(anime *models.Anime) bool {
	if anime == nil {
		return false
	}
	if anime.Source == "SFlix" || anime.Source == "SuperFlix" {
		return true
	}
	if anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		return true
	}
	if strings.Contains(anime.URL, "flixhq") {
		return true
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
	defer func() {
		err := response.Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v\n", err)
		}
	}()

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
		"iframe[src*='blogger']",
		"iframe[src*='blogspot']",
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

	// Try to find blogger link and extract the actual video URL
	videoSrc, err = findBloggerLink(urlBody)
	if err == nil && videoSrc != "" {
		if util.IsDebug {
			util.Debugf("Found blogger link: %s, extracting actual video URL...", videoSrc)
		}
		// Note: The actual extraction happens in extractActualVideoURL
		// which is called after this function
		return videoSrc, nil
	}

	// Try to find direct video URL in content
	matches := videoURLPatternRe.FindString(urlBody)
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
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func findBloggerLink(content string) (string, error) {
	prefix := "https://www.blogger.com/video.g?token="
	idx := strings.Index(content, prefix)
	if idx == -1 {
		return "", errors.New("no blogger video link found in the content")
	}

	start := idx
	curr := idx + len(prefix)
	for curr < len(content) {
		c := content[curr]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			curr++
		} else {
			break
		}
	}

	if curr > idx+len(prefix) {
		url := content[start:curr]
		// Validate the extracted URL with the anchored regex to satisfy security constraints
		if bloggerPatternRe.MatchString(url) {
			return url, nil
		}
	}

	return "", errors.New("no blogger video link found in the content")
}

// newSurfClient creates a surf HTTP client with Chrome browser impersonation.
// Uses NotFollowRedirects for proxy streaming.
func newSurfClient() *surf.Client {
	return surf.NewClient().
		Builder().
		Impersonate().Chrome().
		NotFollowRedirects().
		Build().
		Unwrap()
}

// newSurfDownloadClient creates a surf client for downloading large files.
// Unlike newSurfClient, it follows redirects (googlevideo CDN uses 302s)
// and has a 10-minute timeout to handle large video chunks.
//
// Force HTTP/1.1 for the same reason as newBloggerProxyClient: some
// googlevideo edges omit ALPN, so surf's HTTP/2 dial fails before any request
// is sent ("negotiated ALPN \"\", expected h2"). Requests built with
// http.NoBody then trip surf's h2->h1.1 fallback guard ("cannot retry because
// req.GetBody is nil"), failing the download with "failed to get content
// length". HTTP/1.1 handles both the HEAD length probe and chunked Range GETs.
func newSurfDownloadClient() *surf.Client {
	return surf.NewClient().
		Builder().
		Impersonate().Chrome().
		ForceHTTP1().
		Timeout(10 * time.Minute).
		Build().
		Unwrap()
}

// bloggerSessionClient is a reusable surf session client for Blogger batchexecute.
// Creating a new TLS-impersonated client per request adds ~200-400ms of handshake overhead.
var (
	bloggerSessionClient     *surf.Client
	bloggerSessionClientOnce sync.Once
)

func getBloggerSessionClient() *surf.Client {
	bloggerSessionClientOnce.Do(func() {
		bloggerSessionClient = surf.NewClient().
			Builder().
			Impersonate().Chrome().
			Build().
			Unwrap()
	})
	return bloggerSessionClient
}

// extractBloggerGoogleVideoURL uses surf with Chrome browser impersonation
// to extract the googlevideo URL via Blogger's batchexecute API.
func extractBloggerGoogleVideoURL(bloggerURL string) (string, error) {
	tokenMatch := tokenRe.FindStringSubmatch(bloggerURL)
	if len(tokenMatch) < 2 {
		return "", fmt.Errorf("could not extract token from Blogger URL: %s", bloggerURL)
	}
	token := tokenMatch[1]

	// Use cached session client (follows redirects, keeps cookies) for the batchexecute flow
	client := getBloggerSessionClient()

	// Step 1: Load the Blogger page to extract session params
	result := client.Get(g.String(bloggerURL)).Do()
	if result.IsErr() {
		return "", fmt.Errorf("failed to load Blogger page: %w", result.Err())
	}
	resp := result.Ok()

	pageBody, err := io.ReadAll(io.LimitReader(resp.Body.Stream(), 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read Blogger page: %w", err)
	}
	pageText := string(pageBody)

	sidMatch := sidRe.FindStringSubmatch(pageText)
	bhMatch := bhRe.FindStringSubmatch(pageText)
	atMatch := atRe.FindStringSubmatch(pageText)
	if len(sidMatch) < 2 || len(bhMatch) < 2 {
		return "", errors.New("failed to extract session params (FdrFJe/cfb2h) from Blogger page")
	}
	sid := sidMatch[1]
	bh := bhMatch[1]
	at := ""
	if len(atMatch) >= 2 {
		at = atMatch[1]
	}
	util.Debugf("Blogger extract: SID=%s, build=%s, at=%s", sid, bh, at)

	// Step 2: Call batchexecute to get the googlevideo URL
	inner, err := json.Marshal([]any{token, "", 0})
	if err != nil {
		return "", fmt.Errorf("failed to marshal inner data: %w", err)
	}
	freq, err := json.Marshal([][]any{{[]any{"WcwnYd", string(inner), nil, "generic"}}})
	if err != nil {
		return "", fmt.Errorf("failed to marshal freq data: %w", err)
	}
	postData := "f.req=" + neturl.QueryEscape(string(freq))
	if at != "" {
		postData += "&at=" + neturl.QueryEscape(at)
	}

	util.Debugf("Blogger batchexecute postBody: %s", postData)

	batchURL := fmt.Sprintf(
		"https://www.blogger.com/_/BloggerVideoPlayerUi/data/batchexecute?rpcids=WcwnYd&source-path=%%2Fvideo.g&f.sid=%s&bl=%s&hl=en-US&_reqid=100001&rt=c",
		neturl.QueryEscape(sid), neturl.QueryEscape(bh),
	)

	util.Debugf("Blogger batchexecute URL: %s", batchURL)

	batchResult := client.Post(g.String(batchURL)).
		SetHeaders("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8").
		AddHeaders("X-Same-Domain", "1").
		AddHeaders("Origin", "https://www.blogger.com").
		AddHeaders("Referer", bloggerURL).
		Body(postData).
		Do()
	if batchResult.IsErr() {
		return "", fmt.Errorf("batchexecute request failed: %w", batchResult.Err())
	}
	batchResp := batchResult.Ok()

	batchBody, err := io.ReadAll(io.LimitReader(batchResp.Body.Stream(), 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read batchexecute response: %w", err)
	}

	if int(batchResp.StatusCode) != http.StatusOK {
		util.Debugf("Blogger batchexecute failed with status %d, body: %s", batchResp.StatusCode, string(batchBody))
		return "", fmt.Errorf("batchexecute returned status %d", batchResp.StatusCode)
	}

	// Step 3: Parse the batchexecute response to find the googlevideo URL
	videoURL, err := parseBatchexecuteResponse(batchBody)
	if err != nil {
		return "", err
	}

	util.Debugf("Blogger extract: video URL obtained (%d chars)", len(videoURL))
	return videoURL, nil
}

// errBloggerVideoUnavailable is returned when Google's batchexecute responds
// HTTP 200 but with an empty payload — body contains only the `)]}'`
// anti-hijacking prefix (and optional whitespace) with no streams. This is
// Google's signal for "this RPC produced no result", and in practice means
// the Blogger video has been deleted, region-blocked, or the token is dead.
// Retrying with the same token cannot recover this state, so callers should
// fail fast and let the next-source fallback take over.
var errBloggerVideoUnavailable = errors.New("blogger video unavailable upstream")

// bloggerRPCError reports a batchexecute call that answered HTTP 200 with a
// status code in place of a payload:
//
//	["wrb.fr","WcwnYd",null,null,null,[5],"generic"]
//
// The number in that slot is a canonical gRPC status code. Terminal codes wrap
// errBloggerVideoUnavailable so callers fast-fail; transient ones do not, so
// the existing retry loop still gets a chance.
//
// Reported in issue #195: 20 of 24 episodes of one AnimeFire title answered
// with code 5 (NOT_FOUND) because the Blogger videos had been deleted upstream,
// and every one of them surfaced as the opaque "no video URL found in
// batchexecute response" — indistinguishable from a parser bug.
type bloggerRPCError struct {
	Code int
}

func (e bloggerRPCError) Error() string {
	name, terminal := bloggerRPCStatusName(e.Code)
	if terminal {
		return fmt.Sprintf("blogger video unavailable upstream: RPC status %d (%s) — the video was removed or is not accessible", e.Code, name)
	}
	return fmt.Sprintf("blogger batchexecute returned RPC status %d (%s)", e.Code, name)
}

func (e bloggerRPCError) Unwrap() error {
	if _, terminal := bloggerRPCStatusName(e.Code); terminal {
		return errBloggerVideoUnavailable
	}
	return nil
}

// bloggerRPCStatusName maps a canonical gRPC status code to its name and says
// whether retrying the same token could ever help.
func bloggerRPCStatusName(code int) (name string, terminal bool) {
	switch code {
	case 3:
		return "INVALID_ARGUMENT", true
	case 5:
		return "NOT_FOUND", true
	case 7:
		return "PERMISSION_DENIED", true
	case 9:
		return "FAILED_PRECONDITION", true
	case 16:
		return "UNAUTHENTICATED", true
	case 2:
		return "UNKNOWN", false
	case 4:
		return "DEADLINE_EXCEEDED", false
	case 8:
		return "RESOURCE_EXHAUSTED", false
	case 13:
		return "INTERNAL", false
	case 14:
		return "UNAVAILABLE", false
	default:
		return "unrecognised status", false
	}
}

// batchexecuteRPCStatus returns the status code carried by a WcwnYd envelope
// whose payload slot is null, e.g. ["wrb.fr","WcwnYd",null,null,null,[5],...].
func batchexecuteRPCStatus(body []byte) (int, bool) {
	for line := range strings.SplitSeq(string(body), "\n") {
		if !strings.Contains(line, "wrb.fr") {
			continue
		}
		var outer []any
		if err := jsonx.Unmarshal([]byte(line), &outer); err != nil {
			continue
		}
		for _, entry := range outer {
			arr, ok := entry.([]any)
			if !ok || len(arr) < 6 {
				continue
			}
			if fmt.Sprint(arr[0]) != "wrb.fr" || fmt.Sprint(arr[1]) != "WcwnYd" {
				continue
			}
			if arr[2] != nil {
				continue // a payload is present; this is not an error envelope
			}
			status, ok := arr[5].([]any)
			if !ok || len(status) == 0 {
				continue
			}
			switch v := status[0].(type) {
			case float64:
				return int(v), true
			case int:
				return v, true
			}
		}
	}
	return 0, false
}

// parseBatchexecuteResponse extracts the best MP4 video URL from a Google
// batchexecute API response body (WcwnYd RPC).
//
// Fix 2026-04-23: the previous implementation assumed streams were always at
// data[2] inside the inner JSON payload. When Google changed the response
// structure (or returned fewer top-level elements), the hardcoded index caused
// a silent `continue`, and all three retry attempts produced
// "no video URL found in batchexecute response".
//
// The fix iterates every index of data[] looking for the first element that is
// itself an array of arrays (the streams list). A regex fallback is also
// applied over the raw body in case structured parsing fails entirely.
//
// Fix 2026-04-28: distinguish "empty response" (token dead upstream — the body
// is just `)]}'\n`) from "structured response we failed to parse" so callers
// can fast-fail dead tokens instead of burning 6 retries (3 in scraper +
// 3 in player layer) on a guaranteed-failure code path.
func parseBatchexecuteResponse(body []byte) (string, error) {
	var videoURL string
	for line := range strings.SplitSeq(string(body), "\n") {
		if !strings.Contains(line, "wrb.fr") {
			continue
		}
		var outer []any
		if err := jsonx.Unmarshal([]byte(line), &outer); err != nil {
			continue
		}
		for _, entry := range outer {
			arr, ok := entry.([]any)
			if !ok || len(arr) < 3 {
				continue
			}
			if fmt.Sprint(arr[0]) != "wrb.fr" || fmt.Sprint(arr[1]) != "WcwnYd" {
				continue
			}
			var data []any
			if err := jsonx.Unmarshal(fmt.Append(nil, arr[2]), &data); err != nil {
				continue
			}
			// Search all indices for a streams array (resilient to Google index changes).
			var streams []any
			for i, elem := range data {
				if s, ok := elem.([]any); ok && len(s) > 0 {
					if _, isSlice := s[0].([]any); isSlice {
						streams = s
						util.Debugf("Blogger batchexecute: found streams at data[%d]", i)
						break
					}
				}
			}
			if streams == nil {
				util.Debugf("Blogger batchexecute: no streams array found in data (len=%d)", len(data))
				continue
			}
			// Collect MP4 URLs; prefer 720p (itag=22) over 360p (itag=18).
			var mp4URLs []string
			for _, s := range streams {
				stream, ok := s.([]any)
				if !ok || len(stream) < 1 {
					continue
				}
				u, ok := stream[0].(string)
				if !ok {
					continue
				}
				if strings.Contains(u, "mime=video%2Fmp4") || strings.Contains(u, "mime=video/mp4") {
					mp4URLs = append(mp4URLs, u)
				}
			}
			for _, u := range mp4URLs {
				if strings.Contains(u, "itag=22") {
					videoURL = u
					break
				}
			}
			if videoURL == "" && len(mp4URLs) > 0 {
				videoURL = mp4URLs[0]
			}
			if videoURL == "" && len(streams) > 0 {
				if first, ok := streams[0].([]any); ok && len(first) > 0 {
					if u, ok := first[0].(string); ok {
						videoURL = u
					}
				}
			}
			break
		}
		if videoURL != "" {
			break
		}
	}

	// Regex fallback: scan the raw body for any *.googlevideo.com URL.
	if videoURL == "" {
		if match := googleVideoRe.FindString(string(body)); match != "" {
			util.Debugf("Blogger batchexecute: found googlevideo URL via regex fallback")
			videoURL = match
		}
	}

	if videoURL == "" {
		// Strip Google's anti-hijacking prefix and surrounding whitespace; if
		// nothing remains, the RPC returned no payload — token is dead upstream.
		stripped := bytes.TrimSpace(bytes.TrimPrefix(bytes.TrimSpace(body), []byte(")]}'")))
		if len(stripped) == 0 {
			util.Debugf("Blogger batchexecute returned empty body (%d bytes total) — token unavailable upstream", len(body))
			return "", errBloggerVideoUnavailable
		}
		if code, ok := batchexecuteRPCStatus(body); ok {
			err := bloggerRPCError{Code: code}
			util.Debugf("Blogger batchexecute returned no payload: %v", err)
			return "", err
		}
		util.Debugf("Blogger batchexecute response body (%d bytes, first 500): %s", len(body), string(body[:min(500, len(body))]))
		return "", errors.New("no video URL found in batchexecute response")
	}
	return videoURL, nil
}

// extractBloggerVideoURL uses Chrome TLS impersonation for Blogger's
// batchexecute API and for the resolved googlevideo CDN URL.
// Retries up to 3 times since the first batchexecute call often fails before
// the session cookies are fully established.
func extractBloggerVideoURL(bloggerURL string) (string, error) {
	if util.IsDebug {
		util.Debugf("Extracting actual video URL from Blogger page: %s", bloggerURL)
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		proxyURL, err := startBloggerProxy(bloggerURL)
		if err == nil {
			return proxyURL, nil
		}
		lastErr = err
		// A dead Blogger token (HTTP 200 + empty body from batchexecute) is
		// not transient — retrying cannot resurrect a deleted/region-blocked
		// video, and previously this burned ~3s here plus another ~3s in the
		// player layer's redundant resolve.
		if errors.Is(err, errBloggerVideoUnavailable) {
			util.Debugf("Blogger token unavailable upstream — skipping remaining retries")
			return "", err
		}
		if attempt < maxRetries {
			util.Debugf("Blogger extraction attempt %d/%d failed: %v, retrying...", attempt, maxRetries, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	return "", fmt.Errorf("failed to start Blogger proxy after %d attempts: %w", maxRetries, lastErr)
}

// bloggerProxy holds the state of the running Go HTTP proxy server.
var bloggerProxy struct {
	mu       sync.Mutex
	server   *http.Server
	port     string
	videoURL string // direct googlevideo CDN URL
}

// GetBloggerVideoURL returns the extracted googlevideo URL, or empty string if unavailable.
func GetBloggerVideoURL() string {
	bloggerProxy.mu.Lock()
	defer bloggerProxy.mu.Unlock()
	return bloggerProxy.videoURL
}

// StopBloggerProxy terminates any running Blogger video proxy.
func StopBloggerProxy() {
	bloggerProxy.mu.Lock()
	defer bloggerProxy.mu.Unlock()
	if bloggerProxy.server != nil {
		util.Debugf("Stopping Blogger proxy on port %s", bloggerProxy.port)
		_ = bloggerProxy.server.Close()
		bloggerProxy.server = nil
		bloggerProxy.port = ""
	}
}

// bloggerReadinessTimeout / bloggerReadinessInterval bound the post-start probe
// that confirms the upstream CDN actually serves the video before we hand the
// proxy URL to mpv. Package vars (not consts) so tests can shorten them.
var (
	bloggerReadinessTimeout  = 3 * time.Second
	bloggerReadinessInterval = 30 * time.Millisecond
)

// newBloggerProxyClient creates the long-lived client used between the local
// proxy and googlevideo. Chrome TLS impersonation is required because the CDN
// rejects Go's native TLS fingerprint. Force HTTP/1.1 because some googlevideo
// edges omit ALPN; surf's HTTP/2 path rejects those connections before sending
// the request ("negotiated ALPN \"\", expected h2").
func newBloggerProxyClient() *http.Client {
	client := surf.NewClient().
		Builder().
		Impersonate().Chrome().
		ForceHTTP1().
		Build().
		Unwrap().
		Std()
	client.Timeout = 0 // streaming requests must not have a whole-request timeout
	return client
}

// startBloggerProxy starts a local Go HTTP proxy that extracts the video URL
// from a Blogger page via batchexecute and streams it with Chrome browser
// impersonation. No Python required.
func startBloggerProxy(bloggerURL string) (string, error) {
	// Stop any existing proxy
	StopBloggerProxy()

	// Extract the googlevideo URL before starting the proxy
	videoURL, err := extractBloggerGoogleVideoURL(bloggerURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract video URL: %w", err)
	}

	proxyClient := newBloggerProxyClient()
	// Follow redirects with net/http's default policy (up to 10). googlevideo
	// signs a URL that 302-redirects to the actual CDN node; the previous
	// .NotFollowRedirects() made surf surface that redirect as a
	// "net/http: use last response" error, which the handler turned into a 502 —
	// mpv then failed to open the stream and exited status 2 (looking like a
	// missing-mpv/DLL fault). Following the redirect yields the real 200/206
	// video response. Setting CheckRedirect here is deterministic regardless of
	// surf's builder default; net/http preserves the Range header across the
	// same-host CDN redirect, so streaming/seek still works.
	proxyClient.CheckRedirect = nil

	return startBloggerProxyServer(videoURL, proxyClient)
}

// startBloggerProxyServer binds the local proxy that streams videoURL through
// proxyClient and blocks until a readiness probe confirms the upstream serves a
// 2xx status, or bloggerReadinessTimeout elapses. Split from startBloggerProxy
// so the request forwarding + readiness gating are testable without the live
// Blogger batchexecute extract (which needs network).
func startBloggerProxyServer(videoURL string, proxyClient *http.Client) (string, error) {
	// Store the direct URL for download bypass
	bloggerProxy.mu.Lock()
	bloggerProxy.videoURL = videoURL
	bloggerProxy.mu.Unlock()

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to listen on a free port: %w", err)
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/blogger_proxy", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			upReq, err := http.NewRequest(http.MethodHead, videoURL, http.NoBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			upResp, err := proxyClient.Do(upReq)
			if err != nil {
				util.Debugf("BloggerProxy HEAD err: %v", err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer func() { _ = upResp.Body.Close() }()
			for _, k := range []string{"Content-Type", "Content-Length", "Accept-Ranges"} {
				if v := upResp.Header.Get(k); v != "" {
					w.Header().Set(k, v)
				}
			}
			w.WriteHeader(upResp.StatusCode)
			util.Debugf("BloggerProxy HEAD -> %d", upResp.StatusCode)

		case http.MethodGet:
			upReq, err := http.NewRequest(http.MethodGet, videoURL, http.NoBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			if rng := r.Header.Get("Range"); rng != "" {
				upReq.Header.Set("Range", rng)
			}
			upResp, err := proxyClient.Do(upReq)
			if err != nil {
				util.Debugf("BloggerProxy GET err: %v", err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer func() { _ = upResp.Body.Close() }()
			for _, k := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
				if v := upResp.Header.Get(k); v != "" {
					w.Header().Set(k, v)
				}
			}
			w.WriteHeader(upResp.StatusCode)
			util.Debugf("BloggerProxy GET Range=%s -> %d", r.Header.Get("Range"), upResp.StatusCode)
			written, _ := io.Copy(w, upResp.Body)
			util.Debugf("BloggerProxy GET done: %db", written)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	bloggerProxy.mu.Lock()
	bloggerProxy.server = srv
	bloggerProxy.port = port
	bloggerProxy.mu.Unlock()

	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			util.Debugf("BloggerProxy server error: %v", err)
		}
	}()

	proxyURL := fmt.Sprintf("http://127.0.0.1:%s/blogger_proxy", port)
	util.Debugf("Blogger proxy started on port %s", port)

	// Readiness check with fast polling. A transport-level success is not enough:
	// the proxy forwards HEAD to the googlevideo CDN, and a bad CDN node answers
	// 5xx (seen: 502 when the signed URL redirects to a dead node) with
	// headErr==nil. Launching mpv against that produces a bare "exit status 2"
	// that masquerades as a missing-mpv/DLL problem. Treat any non-2xx upstream
	// status as not-ready so extractBloggerVideoURL's retry loop re-resolves a
	// fresh CDN host instead of handing mpv a dead stream.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	deadline := time.After(bloggerReadinessTimeout)
	interval := bloggerReadinessInterval
	lastStatus := 0
	for {
		select {
		case <-deadline:
			if lastStatus != 0 {
				util.Debugf("Proxy readiness check: upstream kept returning HTTP %d — re-resolving", lastStatus)
				return "", fmt.Errorf("blogger CDN not serving video (HTTP %d)", lastStatus)
			}
			util.Debugf("Proxy readiness check timed out with no upstream response")
			return proxyURL, nil // never answered — let mpv try
		default:
			headResp, headErr := httpClient.Head(proxyURL)
			if headErr != nil {
				time.Sleep(interval)
				continue
			}
			status := headResp.StatusCode
			ctype := headResp.Header.Get("Content-Type")
			clen := headResp.Header.Get("Content-Length")
			_ = headResp.Body.Close()
			if status >= 200 && status < 300 {
				util.Debugf("Proxy readiness check: status=%d content-type=%s content-length=%s", status, ctype, clen)
				return proxyURL, nil
			}
			lastStatus = status
			util.Debugf("Proxy readiness check: upstream not ready (status=%d), retrying", status)
			time.Sleep(interval)
		}
	}
}

// isPlayableVideoURL returns true when the URL points to a directly playable
// video resource (e.g. .mp4, .m3u8, .webm) that mpv can handle without further extraction.
func isPlayableVideoURL(videoURL string) bool {
	lower := strings.ToLower(videoURL)
	return strings.HasSuffix(lower, ".mp4") ||
		strings.Contains(lower, ".mp4?") ||
		strings.HasSuffix(lower, ".m3u8") ||
		strings.Contains(lower, ".m3u8?") ||
		strings.HasSuffix(lower, ".webm") ||
		strings.Contains(lower, ".webm?") ||
		strings.Contains(lower, "source=")
}

// needsVideoExtraction returns true when the URL is an intermediate endpoint
// (e.g. AnimeFire video JSON API or a Blogger embed page) that must be resolved
// to the actual CDN video URL before it can be played by mpv.
func needsVideoExtraction(videoURL string) bool {
	lower := strings.ToLower(videoURL)
	return strings.Contains(lower, "animefire.io/video/") ||
		strings.Contains(lower, "animefire.plus/video/") ||
		strings.Contains(lower, "blogger.com/video") ||
		strings.Contains(lower, "blogspot.com/video")
}

// extractActualVideoURL processes the video source and allows the user to select quality
func extractActualVideoURL(videoSrc string) (string, error) {
	if util.IsDebug {
		util.Debugf("Processing video source: %s", videoSrc)
	}

	// If the URL is a Blogger video embed, extract the actual video URL
	if strings.Contains(videoSrc, "blogger.com") ||
		strings.Contains(videoSrc, "blogspot.com") {
		return extractBloggerVideoURL(videoSrc)
	}

	// If the URL is from animefire, fetch the content
	isAnimeFire := strings.Contains(videoSrc, "animefire.io/video/") ||
		strings.Contains(videoSrc, "animefire.plus/video/")
	if isAnimeFire {
		if util.IsDebug {
			util.Debugf("Found animefire.io video URL, fetching content...")
		}

		// lightspeedst.net (AnimeFire CDN) requires Referer: https://animefire.io to authorise
		// token-signed requests. Without it, mpv/yt-dlp get HTTP 401 Unauthorized while the
		// browser (which sends the referer automatically) plays the same URL fine. Set it
		// here so appendPlaybackRefererArgs propagates it to mpv on the play path, mirroring
		// the existing download.go logic.
		if util.GetGlobalReferer() == "" {
			util.SetGlobalReferer("https://animefire.io")
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
		body, err := io.ReadAll(io.LimitReader(response.Body, 10*1024*1024))
		if err != nil {
			return "", fmt.Errorf("failed to read video page: %w", err)
		}

		// Try to parse as JSON
		var videoResponse VideoResponse
		err = jsonx.Unmarshal(body, &videoResponse)
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

			// Auto-select quality only when the user has already picked a
			// specific quality during this session (e.g. "720p").
			// When GlobalQuality is "best" (the default) we still show the
			// interactive prompt so the user can choose manually.
			if util.GlobalQuality != "" && util.GlobalQuality != "best" {
				selectedSrc := selectQualityFromOptions(videoResponse.Data, util.GlobalQuality)
				if selectedSrc != "" {
					if util.IsDebug {
						util.Debugf("Auto-selected quality preference: %s -> %s", util.GlobalQuality, selectedSrc)
					}
					return selectedSrc, nil
				}
			}

			// Prompt user for quality selection via fancy list (type-to-filter).
			// Logic fixes (descending sort, mirror-N disambiguation, friendly
			// fallback labels, abort routing, index-based selection) live in
			// buildAnimeFireQualityItems and are pinned by
			// animefire_quality_menu_test.go.
			sortedData, qualityLabels := buildAnimeFireQualityItems(videoResponse.Data)

			qIdx, err := tui.PickLabels(qualityLabels, tui.PickOptions{
				Breadcrumb:   "Playback > Quality",
				WindowTitle:  "GoAnime - Quality",
				ItemSingular: "quality",
				ItemPlural:   "qualities",
			})
			if err != nil {
				if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
					return "", ErrBackRequested
				}
				return "", fmt.Errorf("failed to select quality: %w", err)
			}
			if qIdx < 0 || qIdx >= len(sortedData) {
				return "", fmt.Errorf("invalid quality selection: index %d out of range", qIdx)
			}
			picked := sortedData[qIdx]
			selectedSrc := picked.Src

			// Store the selected quality for future use in this session.
			// Persist a canonical "Np" form so selectQualityFromOptions can
			// match it deterministically next time, even if the upstream
			// label changes (e.g. "FHD" → "1080p").
			if util.GlobalQuality == "" || util.GlobalQuality == "best" {
				if res := extractResolution(picked.Label); res > 0 {
					util.GlobalQuality = fmt.Sprintf("%dp", res)
				} else {
					util.GlobalQuality = strings.ToLower(strings.TrimSpace(picked.Label))
				}
				if util.IsDebug {
					util.Debugf("Storing selected quality for session: %s", util.GlobalQuality)
				}
			}

			// Return the selected source URL
			return selectedSrc, nil
		}

		// When Data is empty, check the Token field for a Blogger URL
		if err == nil && len(videoResponse.Data) == 0 && videoResponse.Token != "" {
			if util.IsDebug {
				util.Debugf("AnimeFire returned empty data with token: %s", videoResponse.Token)
			}
			if strings.Contains(videoResponse.Token, "blogger.com") {
				return extractBloggerVideoURL(videoResponse.Token)
			}
		}

		// Fallback: Try to find a direct video URL in the content
		matches := videoURLPatternRe.FindString(string(body))
		if matches != "" {
			if util.IsDebug {
				util.Debugf("Found direct video URL: %s", matches)
			}
			return matches, nil
		}

		// Try to find a blogger link and extract the actual video URL
		bloggerLink, err := findBloggerLink(string(body))
		if err == nil && bloggerLink != "" {
			if util.IsDebug {
				util.Debugf("Found blogger link: %s, extracting actual video URL...", bloggerLink)
			}
			return extractBloggerVideoURL(bloggerLink)
		}
	}

	// If not JSON or no qualities found, return an error
	return "", errors.New("no valid video URL found")
}

// buildAnimeFireQualityItems sorts AnimeFire video sources highest-quality
// first and renders unambiguous, user-friendly labels.
//
// Rules:
//   - parsed resolution from VideoData.Label (e.g. "FHD" / "1080p" → 1080)
//     is the sort key, descending.
//   - sources whose resolution can't be parsed render as "Auto" so the
//     menu never shows an empty or raw URL string.
//   - duplicate labels get a "(mirror N)" suffix so the fuzzyfinder index
//     aligns 1:1 with the returned slice and the user's pick is unambiguous.
func buildAnimeFireQualityItems(videoData []VideoData) (sortedSources []VideoData, qualityLabels []string) {
	sorted := append([]VideoData(nil), videoData...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return extractResolution(sorted[i].Label) > extractResolution(sorted[j].Label)
	})

	labels := make([]string, len(sorted))
	seen := make(map[string]int, len(sorted))
	for i, v := range sorted {
		base := "Auto"
		if res := extractResolution(v.Label); res > 0 {
			base = fmt.Sprintf("%dp", res)
		} else if trimmed := strings.TrimSpace(v.Label); trimmed != "" {
			base = trimmed
		}
		seen[base]++
		if seen[base] > 1 {
			labels[i] = fmt.Sprintf("%s (mirror %d)", base, seen[base])
		} else {
			labels[i] = base
		}
	}
	return sorted, labels
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
	matches := extractResolutionRe.FindStringSubmatch(strings.ToLower(label))
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
	Data  []VideoData `json:"data"`
	Token string      `json:"token"`
}
