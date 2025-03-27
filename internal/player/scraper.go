package player

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

// WINDOWS RELEASE

//func dialMPVSocket(socketPath string) (net.Conn, error) {
//	if runtime.GOOS == "windows" {
//		//Attempt to connect using named pipe on Windows
//		conn, err := winio.DialPipe(socketPath, nil)
//		if err != nil {
//			return nil, fmt.Errorf("failed to connect to named pipe: %w", err)
//		}
//		return conn, nil
//	} else {
//		// Unix-like system uses Unix sockets
//		conn, err := net.Dial("unix", socketPath)
//		if err != nil {
//			return nil, fmt.Errorf("failed to connect to Unix socket: %w", err)
//		}
//		return conn, nil
//	}
//}

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
			log.Printf("Failed to close response body: %v\n", err)
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
		// Returns an error if the "Content-Length" header is missing.
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

// SelectEpisodeWithFuzzyFinder allows the user to select an episode using fuzzy finder
func SelectEpisodeWithFuzzyFinder(episodes []api.Episode) (string, string, error) {
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
func ExtractEpisodeNumber(episodeStr string) string {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeStr)
	if numStr == "" {
		return "1"
	}
	return numStr
}

// GetVideoURLForEpisode gets the video URL for a given episode URL
func GetVideoURLForEpisode(episodeURL string) (string, error) {

	if util.IsDebug {
		log.Printf("Tentando extrair URL de vídeo para o episódio: %s", episodeURL)
	}
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	return extractActualVideoURL(videoURL)
}

func extractVideoURL(url string) (string, error) {
	if util.IsDebug {
		log.Printf("Extraindo URL de vídeo da página: %s", url)
	}

	response, err := api.SafeGet(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to fetch URL: %+v", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(response.Body)

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to parse HTML: %+v", err))
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
				log.Printf("Found elements with selector: %s", selector)
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
						log.Printf("Found video URL in attribute %s: %s", attr, videoSrc)
					}
					return videoSrc, nil
				}
			}
		}
	}

	// If no video element found, try to find in page content
	if util.IsDebug {
		log.Printf("No video elements found, searching in page content")
	}

	urlBody, err := fetchContent(url)
	if err != nil {
		return "", err
	}

	// Try to find blogger link
	videoSrc, err = findBloggerLink(urlBody)
	if err == nil && videoSrc != "" {
		if util.IsDebug {
			log.Printf("Found blogger link: %s", videoSrc)
		}
		return videoSrc, nil
	}

	// Try to find direct video URL in content
	videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
	re := regexp.MustCompile(videoURLPattern)
	matches := re.FindString(urlBody)
	if matches != "" {
		if util.IsDebug {
			log.Printf("Found direct video URL: %s", matches)
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
			log.Printf("Failed to close response body: %v\n", err)
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

// selectVideoQuality allows the user to select a video quality
func selectVideoQuality(videos []VideoData) (string, error) {
	if len(videos) == 0 {
		return "", errors.New("no video qualities available")
	}

	// Create a list of quality options
	var options []string
	for _, video := range videos {
		options = append(options, video.Label)
	}

	if util.IsDebug {
		log.Printf("Available qualities: %v", options)
	}

	// Use fuzzy finder to select quality
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i]
		},
		fuzzyfinder.WithPromptString("Select video quality"),
	)
	if err != nil {
		return "", fmt.Errorf("failed to select quality: %w", err)
	}

	if idx < 0 || idx >= len(videos) {
		return "", errors.New("invalid quality selection")
	}

	selectedQuality := videos[idx]
	if util.IsDebug {
		log.Printf("Selected quality: %s -> %s", selectedQuality.Label, selectedQuality.Src)
	}

	// Ensure the URL matches the selected quality
	url := selectedQuality.Src
	qualityPattern := regexp.MustCompile(`/(\d+)p\.mp4`)
	matches := qualityPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		// Replace the quality in the URL with the selected one
		url = qualityPattern.ReplaceAllString(url, fmt.Sprintf("/%sp.mp4", selectedQuality.Label))
		if util.IsDebug {
			log.Printf("Adjusted URL to match selected quality: %s", url)
		}
	}

	return url, nil
}

func extractActualVideoURL(videoSrc string) (string, error) {
	if util.IsDebug {
		log.Printf("Processing video source: %s", videoSrc)
	}

	if strings.Contains(videoSrc, "blogger.com") {
		return videoSrc, nil
	}

	// If the URL is from animefire.plus, we need to fetch it first
	if strings.Contains(videoSrc, "animefire.plus/video/") {
		if util.IsDebug {
			log.Printf("Found animefire.plus video URL, fetching content...")
		}

		// Fetch the video page
		response, err := api.SafeGet(videoSrc)
		if err != nil {
			return "", fmt.Errorf("failed to fetch video page: %w", err)
		}
		defer response.Body.Close()

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read video page: %w", err)
		}

		// Try to parse as JSON first
		var videoResponse VideoResponse
		err = json.Unmarshal(body, &videoResponse)
		if err == nil && len(videoResponse.Data) > 0 {
			if util.IsDebug {
				log.Printf("Found video data with %d qualities", len(videoResponse.Data))
				for _, v := range videoResponse.Data {
					log.Printf("Available quality: %s -> %s", v.Label, v.Src)
				}
			}
			// If we have multiple qualities, let the user select
			if len(videoResponse.Data) > 1 {
				return selectVideoQuality(videoResponse.Data)
			}
			// If only one quality, use it
			return videoResponse.Data[0].Src, nil
		}

		// Try to find a direct video URL in the content
		videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
		re := regexp.MustCompile(videoURLPattern)
		matches := re.FindString(string(body))
		if matches != "" {
			if util.IsDebug {
				log.Printf("Found direct video URL: %s", matches)
			}
			return matches, nil
		}

		// Try to find a blogger link
		videoSrc, err = findBloggerLink(string(body))
		if err == nil && videoSrc != "" {
			if util.IsDebug {
				log.Printf("Found blogger link: %s", videoSrc)
			}
			return videoSrc, nil
		}
	}

	// Try to parse as JSON first
	var videoResponse VideoResponse
	err := json.Unmarshal([]byte(videoSrc), &videoResponse)
	if err == nil && len(videoResponse.Data) > 0 {
		// If we have multiple qualities, let the user select
		if len(videoResponse.Data) > 1 {
			return selectVideoQuality(videoResponse.Data)
		}
		// If only one quality, use it
		return videoResponse.Data[0].Src, nil
	}

	// If not JSON, try to find a direct video URL
	videoURLPattern := `https?://[^\s<>"]+?\.(?:mp4|m3u8)`
	re := regexp.MustCompile(videoURLPattern)
	matches := re.FindString(videoSrc)

	if matches == "" {
		return "", errors.New("no valid video URL found")
	}

	return matches, nil
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

// selectHighestQualityVideo selects the highest quality video available
// func selectHighestQualityVideo(videos []VideoData) string {
// 	var highestQuality int
// 	var highestQualityURL string
// 	for _, video := range videos {
// 		qualityValue, _ := strconv.Atoi(strings.TrimRight(video.Label, "p"))
// 		if qualityValue > highestQuality {
// 			highestQuality = qualityValue
// 			highestQualityURL = video.Src
// 		}
// 	}
// 	return highestQualityURL
// }
