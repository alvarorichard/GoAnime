package providers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// GoyabuProvider handles episode fetching and stream resolution for Goyabu sources.
type GoyabuProvider struct {
	manager *scraper.ScraperManager
}

func NewGoyabuProvider() *GoyabuProvider {
	return &GoyabuProvider{manager: scraper.NewScraperManager()}
}

func (p *GoyabuProvider) Name() string { return "Goyabu" }

func (p *GoyabuProvider) HasSeasons() bool { return false }

func (p *GoyabuProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.GoyabuType)
	if err != nil {
		return nil, fmt.Errorf("failed to get Goyabu scraper: %w", err)
	}
	episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get Goyabu episodes: %w", err)
	}
	return episodes, nil
}

func (p *GoyabuProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.GoyabuType)
	if err != nil {
		return "", fmt.Errorf("failed to get Goyabu scraper: %w", err)
	}

	streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL)
	if streamErr != nil {
		return "", fmt.Errorf("failed to get Goyabu stream URL: %w", streamErr)
	}
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from Goyabu")
	}
	return streamURL, nil
}
