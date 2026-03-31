// Package providers contains the implementations for various anime sources
// and a registry to manage them.
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

// NewGoyabuProvider creates a new instance of GoyabuProvider.
func NewGoyabuProvider() *GoyabuProvider {
	return &GoyabuProvider{manager: scraper.NewScraperManager()}
}

// Name returns the provider's identifier.
func (p *GoyabuProvider) Name() string { return "Goyabu" }

// HasSeasons returns false since Goyabu doesn't use seasons.
func (p *GoyabuProvider) HasSeasons() bool { return false }

// FetchEpisodes fetches episodes for a given anime.
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

// GetStreamURL fetches the stream URL for an episode.
func (p *GoyabuProvider) GetStreamURL(episode *models.Episode, _ *models.Anime, quality string) (string, error) {
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
