// Package scraper provides a unified interface for different anime sources
package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ScraperType represents different scraper types
type ScraperType int

// searchTimeout is the maximum time to wait for all scrapers
const searchTimeout = 15 * time.Second

const (
	AllAnimeType ScraperType = iota
	AnimefireType
	AnimeDriveType
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
	
	// AnimeDrive - Currently on standby
	// Reason: Site is protected by Cloudflare, no bypass solution found yet
	// TODO: Revisit when Cloudflare protection is removed or bypass method is discovered
	// manager.scrapers[AnimeDriveType] = &AnimeDriveAdapter{client: NewAnimeDriveClient()}

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

			// Add language tags (not source names)
			for _, anime := range results {
				sourceName := sm.getScraperDisplayName(*scraperType)
				languageTag := sm.getLanguageTag(*scraperType)

				// Check if the anime name already has any language tag
				hasLanguageTag := strings.Contains(anime.Name, "[English]") || 
					strings.Contains(anime.Name, "[Portuguese]") ||
					strings.Contains(anime.Name, "[Português]")

				if !hasLanguageTag {
					anime.Name = fmt.Sprintf("%s %s", languageTag, anime.Name)
				}
				// Add metadata to identify the source (internal use only)
				anime.Source = sourceName
			}

			if len(results) > 0 {
				util.Debug("Search completed", "scraper", sm.getScraperDisplayName(*scraperType), "results", len(results))
			}

			return results, nil
		}
		return nil, fmt.Errorf("scraper type %v not found", *scraperType)
	}

	// Search across all scrapers concurrently for better performance
	util.Debug("Starting concurrent search across all sources", "query", query)

	type searchResult struct {
		scraperType ScraperType
		results     []*models.Anime
		err         error
	}

	// Create context with timeout to prevent hanging on slow scrapers
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	resultChan := make(chan searchResult, len(sm.scrapers))
	var wg sync.WaitGroup

	// Launch concurrent searches
	for sType, scraper := range sm.scrapers {
		wg.Add(1)
		go func(st ScraperType, s UnifiedScraper) {
			defer wg.Done()
			util.Debug("Searching in source", "source", sm.getScraperDisplayName(st))

			// Create a channel for this individual search result
			done := make(chan struct{})
			var results []*models.Anime
			var err error

			go func() {
				results, err = s.SearchAnime(query)
				close(done)
			}()

			// Wait for result or context cancellation
			select {
			case <-done:
				resultChan <- searchResult{
					scraperType: st,
					results:     results,
					err:         err,
				}
			case <-ctx.Done():
				util.Debug("Search timeout", "source", sm.getScraperDisplayName(st))
				resultChan <- searchResult{
					scraperType: st,
					results:     nil,
					err:         fmt.Errorf("search timed out after %v", searchTimeout),
				}
			}
		}(sType, scraper)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and track errors
	var searchErrors []string
	for res := range resultChan {
		if res.err != nil {
			// Log error and track it for user feedback
			sourceName := sm.getScraperDisplayName(res.scraperType)
			util.Debug("Search error", "source", sourceName, "error", res.err)
			searchErrors = append(searchErrors, fmt.Sprintf("%s: %v", sourceName, res.err))
			continue
		}

		util.Debug("Search results", "source", sm.getScraperDisplayName(res.scraperType), "count", len(res.results))

		// Add language tags to results (not source names)
		for _, anime := range res.results {
			sourceName := sm.getScraperDisplayName(res.scraperType)
			languageTag := sm.getLanguageTag(res.scraperType)

			// Check if the anime name already has any language tag
			hasLanguageTag := strings.Contains(anime.Name, "[English]") || 
				strings.Contains(anime.Name, "[Portuguese]") ||
				strings.Contains(anime.Name, "[Português]")

			if !hasLanguageTag {
				anime.Name = fmt.Sprintf("%s %s", languageTag, anime.Name)
			}

			// Add metadata to identify the source (internal use only)
			anime.Source = sourceName
		}

		allResults = append(allResults, res.results...)
	}

	// Log warnings for failed sources at INFO level so users can see them
	if len(searchErrors) > 0 {
		for _, errMsg := range searchErrors {
			util.Warn("Search source unavailable", "details", errMsg)
		}
	}

	if len(allResults) == 0 {
		util.Debug("No anime found", "query", query)
		if len(searchErrors) > 0 {
			return nil, fmt.Errorf("no anime found with name: %s (some sources failed: %s)", query, strings.Join(searchErrors, "; "))
		}
		return nil, fmt.Errorf("no anime found with name: %s", query)
	}

	// Count results by source for summary
	animefireCount := 0
	allanimeCount := 0
	animedriveCount := 0
	for _, anime := range allResults {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		} else if anime.Source == "AnimeDrive" {
			animedriveCount++
		}
	}

	if util.IsDebug {
		util.Debug("Search summary",
			"animeFire", animefireCount,
			"allAnime", allanimeCount,
			"animeDrive", animedriveCount,
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
		return "Animefire.io"
	case AnimeDriveType:
		return "AnimeDrive"
	default:
		return "Desconhecido"
	}
}

// getLanguageTag returns a language tag for the source
func (sm *ScraperManager) getLanguageTag(scraperType ScraperType) string {
	switch scraperType {
	case AllAnimeType:
		return "[English]"
	case AnimefireType:
		return "[Portuguese]"
	case AnimeDriveType:
		return "[Portuguese]"
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

// AnimeDriveAdapter adapts AnimeDriveClient to UnifiedScraper interface
type AnimeDriveAdapter struct {
	client *AnimeDriveClient
}

func (a *AnimeDriveAdapter) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *AnimeDriveAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimeDriveAdapter) GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error) {
	// Check if server selection is requested via options
	selectServer := true // Default to showing server selection
	for _, opt := range options {
		if s, ok := opt.(string); ok && s == "auto" {
			selectServer = false
			break
		}
		if b, ok := opt.(bool); ok {
			selectServer = b
		}
	}

	if selectServer {
		return a.client.GetStreamURLWithSelection(episodeURL)
	}
	return a.client.GetStreamURL(episodeURL)
}

func (a *AnimeDriveAdapter) GetType() ScraperType {
	return AnimeDriveType
}