// Package playback provides episode navigation for AllAnime source
package playback

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// AllAnimeNavigator handles navigation between episodes for AllAnime content
type AllAnimeNavigator struct {
	animeID  string
	episodes []string
	client   *scraper.AllAnimeClient
}

// NewAllAnimeNavigator creates a new navigator for AllAnime content
func NewAllAnimeNavigator(anime *models.Anime) (*AllAnimeNavigator, error) {
	if !isAllAnimeSource(anime) {
		return nil, fmt.Errorf("this navigator only works with AllAnime sources")
	}

	animeID := extractAllAnimeID(anime.URL)
	if animeID == "" {
		return nil, fmt.Errorf("could not extract anime ID from URL: %s", anime.URL)
	}

	client := scraper.NewAllAnimeClient()
	navigator := &AllAnimeNavigator{
		animeID: animeID,
		client:  client,
	}

	// Fetch episodes list
	episodes, err := client.GetEpisodesList(animeID, "sub")
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes list: %w", err)
	}

	navigator.episodes = episodes
	util.Debugf("AllAnime Navigator initialized animeID=%s episodes=%d", animeID, len(episodes))
	return navigator, nil
}

// GetNextEpisode returns the next episode after the current one
func (nav *AllAnimeNavigator) GetNextEpisode(currentEpisode string) (string, error) {
	current, err := strconv.Atoi(currentEpisode)
	if err != nil {
		return "", fmt.Errorf("invalid current episode number: %w", err)
	}

	next := current + 1
	if next > len(nav.episodes) {
		return "", fmt.Errorf("no next episode available (current: %d, total: %d)", current, len(nav.episodes))
	}

	util.Debugf("Navigating to next episode from=%d to=%d", current, next)
	return strconv.Itoa(next), nil
}

// GetPreviousEpisode returns the previous episode before the current one
func (nav *AllAnimeNavigator) GetPreviousEpisode(currentEpisode string) (string, error) {
	current, err := strconv.Atoi(currentEpisode)
	if err != nil {
		return "", fmt.Errorf("invalid current episode number: %w", err)
	}

	prev := current - 1
	if prev < 1 {
		return "", fmt.Errorf("no previous episode available (current: %d)", current)
	}

	util.Debugf("Navigating to previous episode from=%d to=%d", current, prev)
	return strconv.Itoa(prev), nil
}

// GetTotalEpisodes returns the total number of episodes
func (nav *AllAnimeNavigator) GetTotalEpisodes() int {
	return len(nav.episodes)
}

// ListAllEpisodes returns all available episode numbers
func (nav *AllAnimeNavigator) ListAllEpisodes() []string {
	result := make([]string, len(nav.episodes))
	for i := range nav.episodes {
		result[i] = strconv.Itoa(i + 1)
	}
	return result
}

// Helper function to check if anime is from AllAnime source
func isAllAnimeSource(anime *models.Anime) bool {
	if anime.Source == "AllAnime" {
		return true
	}

	if strings.Contains(anime.URL, "allanime") {
		return true
	}

	// Check if URL is a short ID (AllAnime typically uses short IDs)
	if len(anime.URL) < 30 &&
		strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") &&
		!strings.Contains(anime.URL, "http") {
		return true
	}

	return false
}

// Helper function to extract AllAnime ID from URL
func extractAllAnimeID(url string) string {
	// For AllAnime, the URL is often just the anime ID
	if !strings.Contains(url, "http") && len(url) < 30 {
		return url
	}

	// Extract ID from full AllAnime URLs if needed
	if strings.Contains(url, "allanime") {
		parts := strings.Split(url, "/")
		for _, part := range parts {
			if len(part) > 5 && len(part) < 30 &&
				strings.ContainsAny(part, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") {
				return part
			}
		}
	}

	return url // Return as-is if can't extract
}

// HandleAllAnimeEpisodeNavigation handles episode navigation for AllAnime
func HandleAllAnimeEpisodeNavigation(anime *models.Anime, currentEpisodeNumber string, direction string) (*models.Episode, error) {
	navigator, err := NewAllAnimeNavigator(anime)
	if err != nil {
		return nil, fmt.Errorf("failed to create AllAnime navigator: %w", err)
	}

	var targetEpisodeNumber string
	switch direction {
	case "next":
		targetEpisodeNumber, err = navigator.GetNextEpisode(currentEpisodeNumber)
		if err != nil {
			return nil, err
		}
	case "previous":
		targetEpisodeNumber, err = navigator.GetPreviousEpisode(currentEpisodeNumber)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid direction: %s (use 'next' or 'previous')", direction)
	}

	// Create episode object
	episode := &models.Episode{
		Number: targetEpisodeNumber,
		URL:    anime.URL, // For AllAnime, episode URL is the anime ID
		Num:    0,         // Will be set by caller
	}

	// Convert episode number to int
	if num, err := strconv.Atoi(targetEpisodeNumber); err == nil {
		episode.Num = num
	}

	return episode, nil
}
