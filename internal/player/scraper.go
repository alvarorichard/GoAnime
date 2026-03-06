package player

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	//"github.com/Microsoft/go-winio"
	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
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
	// Check if this is a URL that might not have Content-Length header
	isKnownStreamURL := strings.Contains(url, "sharepoint.com") ||
		strings.Contains(url, "wixmp.com") ||
		strings.Contains(url, "master.m3u8") ||
		strings.Contains(url, ".m3u8") ||
		strings.Contains(url, "allanime.pro") ||
		strings.Contains(url, "animefire") ||
		strings.Contains(url, "blogger.com") ||
		strings.Contains(url, "animesfire") ||
		strings.Contains(url, "repackager.wixmp.com")

	// Attempts to create an HTTP HEAD request to retrieve headers without downloading the body.
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Returns 0 and the error if the request creation fails.
		return 0, err
	}

	// Sends the HEAD request to the server.
	resp, err := client.Do(req) // #nosec G704
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
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
	req, err := http.NewRequest("GET", url, nil)
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

// SelectEpisodeWithFuzzyFinder allows the user to select an episode using fuzzy finder
func SelectEpisodeWithFuzzyFinder(episodes []models.Episode) (string, string, error) {
	if len(episodes) == 0 {
		return "", "", errors.New("no episodes provided")
	}

	// Create a list with back option at the beginning
	backOption := "← Back"
	displayList := make([]string, len(episodes)+1)
	displayList[0] = backOption
	for i, ep := range episodes {
		displayList[i+1] = ep.Number
	}

	idx, err := fuzzyfinder.Find(
		displayList,
		func(i int) string {
			return displayList[i]
		},
		fuzzyfinder.WithPromptString("Select the episode: "),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to select episode with go-fuzzyfinder: %w", err)
	}

	if idx < 0 || idx >= len(displayList) {
		return "", "", errors.New("invalid index returned by fuzzyfinder")
	}

	// Check if back was selected
	if idx == 0 {
		return "", "", ErrBackRequested
	}

	// Adjust index for episodes (subtract 1 for the back option)
	episodeIdx := idx - 1
	return episodes[episodeIdx].URL, episodes[episodeIdx].Number, nil
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

// GetVideoURLForEpisodeEnhanced gets the video URL using the enhanced API with AllAnime navigation support
func GetVideoURLForEpisodeEnhanced(episode *models.Episode, anime *models.Anime) (string, error) {
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

	// Try AnimeDrive enhanced navigation if applicable
	if isAnimeDriveSourcePlayer(anime) {
		streamURL, err := api.GetEpisodeStreamURL(episode, anime, util.GlobalQuality)
		if err == nil {
			return streamURL, nil
		}
		// Check if user requested to go back from server selection
		if errors.Is(err, scraper.ErrBackRequested) {
			return "", ErrBackToEpisodeSelection
		}
		// For AnimeDrive, return the error instead of trying legacy method
		return "", fmt.Errorf("failed to get AnimeDrive stream URL: %w", err)
	}

	// Try FlixHQ for movies/TV shows
	if isFlixHQSourcePlayer(anime) {
		util.Debug("FlixHQ source detected", "source", anime.Source, "mediaType", anime.MediaType, "episodeURL", episode.URL)
		streamURL, err := api.GetEpisodeStreamURL(episode, anime, util.GlobalQuality)
		if err == nil {
			util.Debug("FlixHQ stream URL obtained", "url", streamURL)
			return streamURL, nil
		}
		util.Debug("FlixHQ stream URL failed", "error", err)
		// For FlixHQ, return the error - legacy method won't work with DataIDs
		return "", fmt.Errorf("failed to get FlixHQ stream URL: %w", err)
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

	// The enhanced API may return intermediate URLs (Blogger embeds, AnimeFire
	// video JSON API) that are not directly playable. Resolve them to actual
	// CDN video URLs before returning.
	if needsVideoExtraction(streamURL) {
		resolved, err := extractActualVideoURL(streamURL)
		if err == nil && resolved != "" {
			return resolved, nil
		}
		// If resolution failed, fall back to the original URL so yt-dlp can try
		util.Debug("Could not resolve intermediate URL, using as-is", "url", streamURL, "err", err)
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
		!strings.Contains(anime.URL, "http") &&
		!strings.Contains(anime.URL, "animesdrive") {
		return true
	}

	return false
}

// Helper function to check if anime is from AnimeDrive source (player module)
func isAnimeDriveSourcePlayer(anime *models.Anime) bool {
	if anime == nil {
		return false
	}
	if anime.Source == "AnimeDrive" {
		return true
	}
	if strings.Contains(anime.Name, "[AnimeDrive]") {
		return true
	}
	if strings.Contains(anime.URL, "animesdrive") {
		return true
	}
	return false
}

// Helper function to check if anime is from FlixHQ source (player module)
func isFlixHQSourcePlayer(anime *models.Anime) bool {
	if anime == nil {
		return false
	}
	if anime.Source == "FlixHQ" {
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
	pattern := `https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 0 {
		return matches[0], nil
	} else {
		return "", errors.New("no blogger video link found in the content")
	}
}

// extractBloggerVideoURL starts a local Python proxy that uses curl_cffi
// (Chrome TLS impersonation) to extract the googlevideo URL via Blogger's
// batchexecute API and stream the video to mpv.
// The entire chain (page load, batchexecute, video streaming) uses Chrome TLS
// so that Google's CDN does not reject the requests.
func extractBloggerVideoURL(bloggerURL string) (string, error) {
	if util.IsDebug {
		util.Debugf("Extracting actual video URL from Blogger page: %s", bloggerURL)
	}

	// Validate that it's a Blogger URL with a token
	tokenRe := regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)
	tokenMatch := tokenRe.FindStringSubmatch(bloggerURL)
	if len(tokenMatch) < 2 {
		return "", fmt.Errorf("could not extract token from Blogger URL: %s", bloggerURL)
	}

	// Start the proxy with the Blogger page URL; the Python script
	// handles batchexecute extraction internally using curl_cffi.
	proxyURL, err := startBloggerProxy(bloggerURL)
	if err != nil {
		return "", fmt.Errorf("failed to start Blogger proxy: %w", err)
	}

	return proxyURL, nil
}

// bloggerProxy holds the state of the running proxy process.
var bloggerProxy struct {
	mu   sync.Mutex
	cmd  *exec.Cmd
	port string
}

// StopBloggerProxy terminates any running Blogger video proxy.
func StopBloggerProxy() {
	bloggerProxy.mu.Lock()
	defer bloggerProxy.mu.Unlock()
	if bloggerProxy.cmd != nil && bloggerProxy.cmd.Process != nil {
		util.Debugf("Stopping Blogger proxy (PID %d)", bloggerProxy.cmd.Process.Pid)
		_ = bloggerProxy.cmd.Process.Kill()
		_ = bloggerProxy.cmd.Wait()
		bloggerProxy.cmd = nil
		bloggerProxy.port = ""
	}
}

// bloggerProxyScript is the Python script that:
//  1. Loads the Blogger page and calls batchexecute (WcwnYd) to extract the
//     googlevideo.com stream URL — all via curl_cffi Chrome TLS impersonation.
//  2. Starts a local HTTP proxy that streams the video to mpv using the same
//     Chrome TLS fingerprint.
//
// Using a single TLS identity end-to-end prevents Google's CDN from rejecting
// requests with 403 due to fingerprint mismatch.
const bloggerProxyScript = `
import sys, os, re, json, socket
from http.server import HTTPServer, BaseHTTPRequestHandler
from socketserver import ThreadingMixIn
from urllib.parse import quote as urlquote
from curl_cffi import requests as cffi_requests

def extract_video_url(blogger_url):
    s = cffi_requests.Session(impersonate='chrome')
    r = s.get(blogger_url)
    sid_m = re.search(r'"FdrFJe"\s*:\s*"([^"]+)"', r.text)
    bh_m = re.search(r'"cfb2h"\s*:\s*"([^"]+)"', r.text)
    tok_m = re.search(r'token=([A-Za-z0-9_-]+)', blogger_url)
    if not sid_m or not bh_m or not tok_m:
        raise RuntimeError('Failed to extract session params from Blogger page')
    sid, bh, token = sid_m.group(1), bh_m.group(1), tok_m.group(1)
    sys.stderr.write(f'PROXY extract: SID={sid}, build={bh}\n'); sys.stderr.flush()
    inner = json.dumps([token, '', 0])
    freq = json.dumps([[['WcwnYd', inner, None, 'generic']]])
    post_body = 'f.req=' + urlquote(freq) + '&'
    batch_url = (
        'https://www.blogger.com/_/BloggerVideoPlayerUi/data/batchexecute'
        '?rpcids=WcwnYd&source-path=%2Fvideo.g'
        f'&f.sid={urlquote(sid)}&bl={urlquote(bh)}'
        '&hl=en-US&_reqid=100001&rt=c'
    )
    r2 = s.post(batch_url, data=post_body, headers={
        'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8',
        'X-Same-Domain': '1',
        'Origin': 'https://www.blogger.com',
        'Referer': 'https://www.blogger.com/',
    })
    if r2.status_code != 200:
        raise RuntimeError(f'batchexecute returned {r2.status_code}')
    video_url = None
    for line in r2.text.split('\n'):
        if 'wrb.fr' not in line:
            continue
        outer = json.loads(line)
        for entry in outer:
            if entry[0] != 'wrb.fr' or entry[1] != 'WcwnYd':
                continue
            data = json.loads(entry[2])
            for stream in data[2]:
                u = stream[0]
                if 'mime=video%2Fmp4' in u or 'mime=video/mp4' in u:
                    video_url = u
                    break
            if not video_url:
                video_url = data[2][0][0]
            break
        break
    if not video_url:
        raise RuntimeError('No video URL found in batchexecute response')
    sys.stderr.write(f'PROXY extract: video URL obtained ({len(video_url)} chars)\n')
    sys.stderr.flush()
    return video_url

BLOGGER_URL = sys.argv[1]
VIDEO_URL = extract_video_url(BLOGGER_URL)

class TS(ThreadingMixIn, HTTPServer):
    daemon_threads = True

class H(BaseHTTPRequestHandler):
    def _s(self):
        return cffi_requests.Session(impersonate='chrome')
    def do_HEAD(self):
        try:
            r = self._s().head(VIDEO_URL)
            self.send_response(r.status_code)
            for k in ('Content-Type','Content-Length','Accept-Ranges'):
                v = r.headers.get(k)
                if v: self.send_header(k, v)
            self.end_headers()
            sys.stderr.write(f'PROXY HEAD -> {r.status_code}\n'); sys.stderr.flush()
        except Exception as e:
            sys.stderr.write(f'PROXY HEAD err: {e}\n'); sys.stderr.flush()
            self.send_error(502)
    def do_GET(self):
        try:
            h = {}
            rng = self.headers.get('Range')
            if rng: h['Range'] = rng
            r = self._s().get(VIDEO_URL, headers=h, stream=True)
            self.send_response(r.status_code)
            for k in ('Content-Type','Content-Length','Content-Range','Accept-Ranges'):
                v = r.headers.get(k)
                if v: self.send_header(k, v)
            self.end_headers()
            sys.stderr.write(f'PROXY GET Range={rng} -> {r.status_code}\n'); sys.stderr.flush()
            w = 0
            for c in r.iter_content(chunk_size=65536):
                self.wfile.write(c); w += len(c)
            sys.stderr.write(f'PROXY GET done: {w}b\n'); sys.stderr.flush()
        except (BrokenPipeError, ConnectionResetError):
            sys.stderr.write('PROXY GET: disconn\n'); sys.stderr.flush()
        except Exception as e:
            sys.stderr.write(f'PROXY GET err: {e}\n'); sys.stderr.flush()
            try: self.send_error(502)
            except Exception: pass
    def log_message(self, *a): pass

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
sock.bind(('127.0.0.1', 0))
port = sock.getsockname()[1]
sock.close()
srv = TS(('127.0.0.1', port), H)
sys.stdout.write(f'{port}\n'); sys.stdout.flush()
srv.serve_forever()
`

// startBloggerProxy starts a local Python proxy that extracts the video URL
// from a Blogger page via batchexecute and streams it with Chrome TLS.
func startBloggerProxy(bloggerURL string) (string, error) {
	// Stop any existing proxy
	StopBloggerProxy()

	bloggerProxy.mu.Lock()
	defer bloggerProxy.mu.Unlock()

	// #nosec G204 -- bloggerURL comes from the trusted Blogger iframe src attribute
	cmd := exec.Command("python3", "-c", bloggerProxyScript, bloggerURL)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr from the proxy for debug logging
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start proxy: %w", err)
	}

	// Forward proxy stderr to debug log in background
	go func() {
		sc := bufio.NewScanner(stderrPipe)
		for sc.Scan() {
			util.Debugf("BloggerProxy: %s", sc.Text())
		}
	}()

	// Read the port number from the first line of stdout
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		_ = cmd.Process.Kill()
		return "", errors.New("proxy did not output port number")
	}
	port := strings.TrimSpace(scanner.Text())
	if port == "" {
		_ = cmd.Process.Kill()
		return "", errors.New("proxy returned empty port")
	}

	bloggerProxy.cmd = cmd
	bloggerProxy.port = port

	proxyURL := fmt.Sprintf("http://127.0.0.1:%s/blogger_proxy", port)
	util.Debugf("Blogger proxy started on port %s (PID %d)", port, cmd.Process.Pid)

	// Readiness check: verify the proxy responds before returning
	client := &http.Client{Timeout: 5 * 1000000000} // 5 seconds
	headResp, headErr := client.Head(proxyURL)
	if headErr != nil {
		util.Debugf("Proxy readiness check failed: %v", headErr)
	} else {
		util.Debugf("Proxy readiness check: status=%d content-type=%s content-length=%s",
			headResp.StatusCode, headResp.Header.Get("Content-Type"), headResp.Header.Get("Content-Length"))
		_ = headResp.Body.Close()
	}

	return proxyURL, nil
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

			// Prompt user for quality selection
			// Create options for huh.Select with back option first
			var options []huh.Option[string]
			options = append(options, huh.NewOption("← Back", "back"))
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

			// Handle back selection
			if selectedSrc == "back" {
				return "", ErrBackRequested
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
		videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
		re := regexp.MustCompile(videoURLPattern)
		matches := re.FindString(string(body))
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
	Data  []VideoData `json:"data"`
	Token string      `json:"token"`
}
