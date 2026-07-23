package scraper

import (
	"sync/atomic"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
)

// MockScraper implements UnifiedScraper for testing the adapter layer and
// source health diagnostics without real network I/O.
type MockScraper struct {
	searchFunc      func(query string) ([]*models.Anime, error)
	episodesFunc    func(url string) ([]models.Episode, error)
	streamURLFunc   func(url string) (string, map[string]string, error)
	scraperType     ScraperType
	searchCallCount atomic.Int32
	searchDelay     time.Duration
}

func (m *MockScraper) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	m.searchCallCount.Add(1)
	if m.searchDelay > 0 {
		time.Sleep(m.searchDelay)
	}
	if m.searchFunc != nil {
		return m.searchFunc(query)
	}
	return nil, nil
}

func (m *MockScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	if m.episodesFunc != nil {
		return m.episodesFunc(animeURL)
	}
	return nil, nil
}

func (m *MockScraper) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	if m.streamURLFunc != nil {
		return m.streamURLFunc(episodeURL)
	}
	return "", nil, nil
}

func (m *MockScraper) GetType() ScraperType {
	return m.scraperType
}
