// Package scraper provides a unified interface for different anime sources
package scraper

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ScraperType represents different scraper types
type ScraperType int

const (
	AllAnimeType ScraperType = iota
	AnimefireType
)

// UnifiedScraper provides a common interface for all scrapers
type UnifiedScraper interface {
	SearchAnime(query string, options ...interface{}) ([]*models.Anime, error)
	GetAnimeEpisodes(animeURL string) ([]models.Episode, error)
	GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error)
	GetType() ScraperType
}

// ScraperManager manages multiple scrapers
type ScraperManager struct {
	scrapers map[ScraperType]UnifiedScraper
}

// NewScraperManager creates a new scraper manager
func NewScraperManager() *ScraperManager {
	manager := &ScraperManager{
		scrapers: make(map[ScraperType]UnifiedScraper),
	}

	// Initialize scrapers
	manager.scrapers[AllAnimeType] = &AllAnimeAdapter{client: NewAllAnimeClient()}
	manager.scrapers[AnimefireType] = &AnimefireAdapter{client: NewAnimefireClient()}

	return manager
}

// SearchAnime searches across all available scrapers with enhanced Portuguese messaging
func (sm *ScraperManager) SearchAnime(query string, scraperType *ScraperType) ([]*models.Anime, error) {
	var allResults []*models.Anime

	if scraperType != nil {
		// Search using specific scraper
		if scraper, exists := sm.scrapers[*scraperType]; exists {
			util.Debug("Searching specific scraper", "scraper", sm.getScraperDisplayName(*scraperType))

			results, err := scraper.SearchAnime(query)
			if err != nil {
				return nil, fmt.Errorf("busca falhou em %s: %w", sm.getScraperDisplayName(*scraperType), err)
			}

			// Add source tags even for specific searches
			for _, anime := range results {
				sourceName := sm.getScraperDisplayName(*scraperType)
				sourceTag := sm.getSourceTag(*scraperType)

				if !strings.Contains(anime.Name, fmt.Sprintf("[%s]", sourceName)) && !strings.Contains(anime.Name, sourceTag) {
					anime.Name = fmt.Sprintf("%s %s", sourceTag, anime.Name)
				}
				// Add metadata to identify the source
				anime.Source = sourceName
			}

			if len(results) > 0 {
				util.Debug("Search completed", "scraper", sm.getScraperDisplayName(*scraperType), "results", len(results))
			}

			return results, nil
		}
		return nil, fmt.Errorf("tipo de scraper %v não encontrado", *scraperType)
	}

	// Search across all scrapers simultaneously
	util.Debug("Starting simultaneous search", "query", query)

	for scraperType, scraper := range sm.scrapers {
		util.Debug("Searching in source", "source", sm.getScraperDisplayName(scraperType))

		results, err := scraper.SearchAnime(query)
		if err != nil {
			// Log error but continue with other scrapers
			util.Debug("Search error", "source", sm.getScraperDisplayName(scraperType), "error", err)
			continue
		}

		util.Debug("Search results", "source", sm.getScraperDisplayName(scraperType), "count", len(results))

		// Add source information to results with enhanced formatting
		for _, anime := range results {
			sourceName := sm.getScraperDisplayName(scraperType)
			sourceTag := sm.getSourceTag(scraperType)

			if !strings.Contains(anime.Name, sourceTag) {
				anime.Name = fmt.Sprintf("%s %s", sourceTag, anime.Name)
			}

			// Add metadata to identify the source
			anime.Source = sourceName
		}

		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		util.Debug("No anime found", "query", query)
		return nil, fmt.Errorf("no anime found with name: %s", query)
	}

	// Count results by source for summary
	animefireCount := 0
	allanimeCount := 0
	for _, anime := range allResults {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		}
	}

	if util.IsDebug {
		util.Debug("Search summary",
			"animeFire", animefireCount,
			"allAnime", allanimeCount,
			"total", len(allResults))
	}

	return allResults, nil
}

// GetScraper returns a specific scraper by type
func (sm *ScraperManager) GetScraper(scraperType ScraperType) (UnifiedScraper, error) {
	if scraper, exists := sm.scrapers[scraperType]; exists {
		return scraper, nil
	}
	return nil, fmt.Errorf("scraper type %v not found", scraperType)
}

// getScraperDisplayName returns a Portuguese display name for the scraper type
func (sm *ScraperManager) getScraperDisplayName(scraperType ScraperType) string {
	switch scraperType {
	case AllAnimeType:
		return "AllAnime"
	case AnimefireType:
		return "AnimeFire.plus"
	default:
		return "Desconhecido"
	}
}

// getSourceTag returns a colored tag for the source
func (sm *ScraperManager) getSourceTag(scraperType ScraperType) string {
	switch scraperType {
	case AllAnimeType:
		return "[AllAnime]"
	case AnimefireType:
		return "[AnimeFire]"
	default:
		return "[Unknown]"
	}
}

// AllAnimeAdapter adapts AllAnimeClient to UnifiedScraper interface
type AllAnimeAdapter struct {
	client *AllAnimeClient
}

func (a *AllAnimeAdapter) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
	// mode is now hardcoded in the new implementation
	return a.client.SearchAnime(query)
}

func (a *AllAnimeAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// For AllAnime, animeURL is actually the anime ID
	episodes, err := a.client.GetEpisodesList(animeURL, "sub")
	if err != nil {
		return nil, err
	}

	var episodeModels []models.Episode
	for i, ep := range episodes {
		episodeModels = append(episodeModels, models.Episode{
			Number: ep,
			Num:    i + 1,
			URL:    animeURL, // Store the anime ID in URL field
			Title: models.TitleDetails{
				Romaji: fmt.Sprintf("Episode %s", ep),
			},
		})
	}

	return episodeModels, nil
}

func (a *AllAnimeAdapter) GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error) {
	// For AllAnime, episodeURL contains the anime ID
	animeID := episodeURL

	// Parse options to get episode number
	episodeNo := "1"
	if len(options) > 0 {
		if ep, ok := options[0].(string); ok {
			episodeNo = ep
		}
	}

	quality := "best"
	if len(options) > 1 {
		if q, ok := options[1].(string); ok {
			quality = q
		}
	}

	mode := "sub"
	if len(options) > 2 {
		if m, ok := options[2].(string); ok {
			mode = m
		}
	}

	return a.client.GetEpisodeURL(animeID, episodeNo, mode, quality)
}

func (a *AllAnimeAdapter) GetType() ScraperType {
	return AllAnimeType
}

// AnimefireAdapter adapts AnimefireClient to UnifiedScraper interface
type AnimefireAdapter struct {
	client *AnimefireClient
}

func (a *AnimefireAdapter) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *AnimefireAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimefireAdapter) GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error) {
	url, err := a.client.GetEpisodeStreamURL(episodeURL)
	metadata := make(map[string]string)
	metadata["source"] = "animefire"
	return url, metadata, err
}

func (a *AnimefireAdapter) GetType() ScraperType {
	return AnimefireType
}
