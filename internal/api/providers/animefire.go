package providers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// AnimeFireProvider handles episode fetching and stream resolution for AnimeFire sources.
type AnimeFireProvider struct {
	manager *scraper.ScraperManager
}

func NewAnimeFireProvider() *AnimeFireProvider {
	return &AnimeFireProvider{manager: scraper.NewScraperManager()}
}

func (p *AnimeFireProvider) Name() string { return "Animefire.io" }

func (p *AnimeFireProvider) HasSeasons() bool { return false }

func (p *AnimeFireProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AnimefireType)
	if err != nil {
		return nil, fmt.Errorf("failed to get AnimeFire scraper: %w", err)
	}
	episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get AnimeFire episodes: %w", err)
	}
	return episodes, nil
}

func (p *AnimeFireProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AnimefireType)
	if err != nil {
		return "", fmt.Errorf("failed to get AnimeFire scraper: %w", err)
	}

	if quality == "" {
		quality = "best"
	}

	streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL, quality)
	if streamErr != nil {
		return "", fmt.Errorf("failed to get AnimeFire stream URL: %w", streamErr)
	}
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from AnimeFire")
	}
	return streamURL, nil
}
