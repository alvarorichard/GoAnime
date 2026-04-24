// Package providers registers and dispatches to anime source providers.
package providers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// goyabuProvider is the explicit HTTP-only provider for Goyabu.
// It delegates to the existing scraper layer and never uses headless browser automation.
type goyabuProvider struct {
	sm scraperLookup
}

func init() {
	RegisterProvider(source.Goyabu, func(sm *scraper.ScraperManager) Provider {
		return &goyabuProvider{sm: sm}
	})
}

func (p *goyabuProvider) Kind() source.SourceKind { return source.Goyabu }
func (p *goyabuProvider) HasSeasons() bool        { return false }

func (p *goyabuProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.GoyabuType)
	if err != nil {
		return nil, err
	}

	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *goyabuProvider) FetchStreamURL(_ context.Context, episode *models.Episode, _ *models.Anime, _ string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.GoyabuType)
	if err != nil {
		return "", err
	}

	const maxAttempts = 4
	for attempt := 0; attempt < maxAttempts; attempt++ {
		url, _, err := adapter.GetStreamURL(episode.URL)
		if err == nil {
			return url, nil
		}

		if !errors.Is(err, scraper.ErrSourceUnavailable) || attempt == maxAttempts-1 {
			return "", fmt.Errorf("goyabu stream: %w", err)
		}

		time.Sleep(time.Duration(attempt+1) * 1200 * time.Millisecond)
	}

	return "", fmt.Errorf("goyabu stream: exhausted retries")
}
