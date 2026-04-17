// Package playback provides episode navigation for AllAnime source
package playback

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// navigatorCache caches AllAnimeNavigator instances keyed by anime ID
// so that episode lists are fetched once and reused across next/prev calls.
var (
	navigatorCache   = make(map[string]*AllAnimeNavigator)
	navigatorCacheMu sync.Mutex
)

// AllAnimeNavigator handles navigation between episodes for AllAnime content
type AllAnimeNavigator struct {
	animeID  string
	episodes []string
	client   *scraper.AllAnimeClient
}

// NewAllAnimeNavigator creates a new navigator for AllAnime content
func NewAllAnimeNavigator(anime *models.Anime) (*AllAnimeNavigator, error) {
	if !api.IsAllAnimeSource(anime) {
		return nil, fmt.Errorf("this navigator only works with AllAnime sources")
	}

	animeID := api.ExtractAllAnimeID(anime.URL)
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

// HandleAllAnimeEpisodeNavigation handles episode navigation for AllAnime
func HandleAllAnimeEpisodeNavigation(anime *models.Anime, currentEpisodeNumber, direction string) (*models.Episode, error) {
	// Use cached navigator to avoid re-fetching the entire episode list
	animeID := api.ExtractAllAnimeID(anime.URL)
	navigatorCacheMu.Lock()
	navigator, ok := navigatorCache[animeID]
	navigatorCacheMu.Unlock()

	if !ok {
		var err error
		navigator, err = NewAllAnimeNavigator(anime)
		if err != nil {
			return nil, fmt.Errorf("failed to create AllAnime navigator: %w", err)
		}
		navigatorCacheMu.Lock()
		navigatorCache[animeID] = navigator
		navigatorCacheMu.Unlock()
	}

	var targetEpisodeNumber string
	var err error
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
