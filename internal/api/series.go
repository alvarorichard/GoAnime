package api

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
