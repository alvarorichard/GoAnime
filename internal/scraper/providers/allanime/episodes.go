// Package scraper provides web scraping functionality for anime sources
package allanime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

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

	if err := netx.CheckHTTPStatus(resp, "AllAnime episodes list"); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := netx.CheckHTMLResponse(resp, body, "AllAnime episodes list"); err != nil {
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

	if err := jsonx.Unmarshal(body, &response); err != nil {
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
