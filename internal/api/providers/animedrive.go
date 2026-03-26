package providers

import (
	"errors"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// AnimeDriveProvider handles episode fetching and stream resolution for AnimeDrive sources.
type AnimeDriveProvider struct {
	manager *scraper.ScraperManager
}

func NewAnimeDriveProvider() *AnimeDriveProvider {
	return &AnimeDriveProvider{manager: scraper.NewScraperManager()}
}

func (p *AnimeDriveProvider) Name() string { return "AnimeDrive" }

func (p *AnimeDriveProvider) HasSeasons() bool { return false }

func (p *AnimeDriveProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return nil, fmt.Errorf("failed to get AnimeDrive scraper: %w", err)
	}
	episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get AnimeDrive episodes: %w", err)
	}
	return episodes, nil
}

func (p *AnimeDriveProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return "", fmt.Errorf("failed to get AnimeDrive scraper: %w", err)
	}

	// "auto" skips interactive server selection (runs inside spinner context)
	streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL, "auto")
	if streamErr != nil {
		if errors.Is(streamErr, scraper.ErrBackRequested) {
			return "", streamErr
		}
		return "", fmt.Errorf("failed to get AnimeDrive stream URL: %w", streamErr)
	}
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from AnimeDrive")
	}
	return streamURL, nil
}
