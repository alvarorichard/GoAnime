package api

import "github.com/alvarorichard/Goanime/internal/models"

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
