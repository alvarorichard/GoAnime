package api

func IsSeries(animeURL string) (bool, int, error) {
	episodes, err := GetAnimeEpisodes(animeURL)
	if err != nil {
		return false, 0, err
	}
	return len(episodes) > 1, len(episodes), nil
}
