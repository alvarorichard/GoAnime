package api

import "github.com/alvarorichard/Goanime/internal/models"

// IsSeries checks if the given anime URL corresponds to a series (multiple episodes).
// It returns a boolean indicating if the anime has more than one episode, the total number of episodes, and an error if any issues occur.
//
// Parameters:
// - animeURL: the URL of the anime's page.
//
// Returns:
// - bool: true if the anime has more than one episode (i.e., is a series), false otherwise.
// - int: the total number of episodes found.
// - error: an error if the process of retrieving episodes fails.
func IsSeries(animeURL string) (bool, int, error) {
	// Retrieve the list of episodes for the given anime URL.
	episodes, err := GetAnimeEpisodes(animeURL)
	if err != nil {
		// Return false, 0, and the error if there's an issue retrieving episodes.
		return false, 0, err
	}

	// Return true if the anime has more than one episode, along with the episode count.
	return len(episodes) > 1, len(episodes), nil
}

// IsSeriesEnhanced checks if the given anime corresponds to a series using enhanced API
func IsSeriesEnhanced(anime *models.Anime) (bool, int, error) {
	// Use enhanced episode fetching
	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		return false, 0, err
	}

	// Count the total number of episodes retrieved.
	totalEpisodes := len(episodes)

	// Return true if there's more than one episode, indicating it's a series.
	return totalEpisodes > 1, totalEpisodes, nil
}
