// Package goanime provides a public API for anime scraping and searching functionality.
// This package can be used as a library in other Go projects.
package goanime

import (
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

// Client is the main client for interacting with anime sources
type Client struct {
	manager *scraper.ScraperManager
}

// NewClient creates a new GoAnime client with all available scrapers
func NewClient() *Client {
	return &Client{
		manager: scraper.NewScraperManager(),
	}
}

// SearchAnime searches for anime across all sources or a specific source.
// If source is nil, searches all available sources.
// Returns a list of anime results or an error.
func (c *Client) SearchAnime(query string, source *types.Source) ([]*types.Anime, error) {
	var scraperType *scraper.ScraperType
	if source != nil {
		st := source.ToScraperType()
		scraperType = &st
	}

	results, err := c.manager.SearchAnime(query, scraperType)
	if err != nil {
		return nil, err
	}

	// Convert internal models to public types
	return types.FromInternalAnimeList(results), nil
}

// GetAnimeEpisodes retrieves all episodes for a specific anime.
// The animeURL should be obtained from a SearchAnime result.
func (c *Client) GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error) {
	scraper, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return nil, err
	}

	episodes, err := scraper.GetAnimeEpisodes(animeURL)
	if err != nil {
		return nil, err
	}

	return types.FromInternalEpisodeList(episodes), nil
}

// GetStreamURL retrieves the streaming URL and headers for a specific episode.
// The episodeURL should be obtained from GetAnimeEpisodes.
func (c *Client) GetStreamURL(episodeURL string, source types.Source, options ...interface{}) (string, map[string]string, error) {
	scraper, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return "", nil, err
	}

	return scraper.GetStreamURL(episodeURL, options...)
}

// GetAvailableSources returns a list of all available scraper sources.
func (c *Client) GetAvailableSources() []types.Source {
	return []types.Source{
		types.SourceAllAnime,
		types.SourceAnimeFire,
	}
}
