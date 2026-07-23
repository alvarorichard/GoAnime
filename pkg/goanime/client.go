// Package goanime provides a public API for anime scraping and searching functionality.
// This package can be used as a library in other Go projects.
package goanime

import (
	"context"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/api/providers"
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

// Manager abstracts the scraper backend used by Client. registryManager (the
// default) satisfies it on top of the Model B registry.
type Manager interface {
	SearchAnime(query string, scraperType *scraper.ScraperType) ([]*models.Anime, error)
	GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error)
}

// Client is the main client for interacting with anime sources
type Client struct {
	manager Manager
}

// NewClient creates a new GoAnime client backed by the Model B registry.
func NewClient() *Client {
	return &Client{
		manager: registryManager{},
	}
}

// registryManager implements Manager over the Model B registry: searches fan out
// through providers.SearchAll and per-source access builds a standalone adapter
// via scraper.NewAdapter — the same paths the CLI uses, so the SDK stays in sync
// with app behavior (kill-switch, tagging, circuit breaker) for free.
type registryManager struct{}

func (registryManager) SearchAnime(query string, scraperType *scraper.ScraperType) ([]*models.Anime, error) {
	if scraperType != nil {
		kind, ok := scraperKind(*scraperType)
		if !ok {
			return nil, fmt.Errorf("unknown scraper type %v", *scraperType)
		}
		return providers.SearchAll(context.Background(), query, kind)
	}
	return providers.SearchAll(context.Background(), query)
}

func (registryManager) GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error) {
	return scraper.NewAdapter(scraperType)
}

// scraperKind maps an internal ScraperType to its Model B SourceKind.
func scraperKind(st scraper.ScraperType) (apisource.SourceKind, bool) {
	switch st {
	case scraper.AllAnimeType:
		return apisource.AllAnime, true
	case scraper.AnimefireType:
		return apisource.AnimeFire, true
	case scraper.GoyabuType:
		return apisource.Goyabu, true
	case scraper.SuperFlixType:
		return apisource.SuperFlix, true
	default:
		return "", false
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
	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return nil, err
	}

	episodes, err := scr.GetAnimeEpisodes(animeURL)
	if err != nil {
		return nil, err
	}

	// For AllAnime, we need to store the anime ID in episodes for later stream URL retrieval
	if source == types.SourceAllAnime {
		for i := range episodes {
			episodes[i].URL = animeURL // Store anime ID in URL field
		}
	}

	return types.FromInternalEpisodeList(episodes), nil
}

// GetStreamURL retrieves the streaming URL and headers for a specific episode.
// The episodeURL should be obtained from GetAnimeEpisodes.
//
// Deprecated: Use GetEpisodeStreamURL instead for better control over quality and mode.
func (c *Client) GetStreamURL(episodeURL string, source types.Source, options ...any) (streamURL string, metadata map[string]string, err error) {
	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return "", nil, err
	}

	return scr.GetStreamURL(episodeURL, options...)
}

// StreamOptions contains options for retrieving stream URLs
type StreamOptions struct {
	// Quality can be "best", "worst", "1080p", "720p", "480p", "360p"
	Quality string
	// Mode can be "sub" (subtitled) or "dub" (dubbed)
	Mode string
}

// DefaultStreamOptions returns default stream options
func DefaultStreamOptions() StreamOptions {
	return StreamOptions{
		Quality: "best",
		Mode:    "sub",
	}
}

// GetEpisodeStreamURL retrieves the streaming URL for a specific episode with full control.
// This is the recommended method to get playback URLs.
//
// Parameters:
//   - anime: The anime object from SearchAnime
//   - episode: The episode object from GetAnimeEpisodes
//   - options: Optional StreamOptions (uses defaults if nil)
//
// Returns:
//   - streamURL: Direct URL for video playback
//   - metadata: Additional info like quality, source, etc.
//   - error: Any error that occurred
func (c *Client) GetEpisodeStreamURL(anime *types.Anime, episode *types.Episode, options *StreamOptions) (streamURL string, metadata map[string]string, err error) {
	source, err := types.ParseSource(anime.Source)
	if err != nil {
		return "", nil, err
	}

	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return "", nil, err
	}

	// Set default options if not provided
	opts := DefaultStreamOptions()
	if options != nil {
		if options.Quality != "" {
			opts.Quality = options.Quality
		}
		if options.Mode != "" {
			opts.Mode = options.Mode
		}
	}

	// For AllAnime, we need to pass: animeID (URL), episodeNumber, quality, mode
	if source == types.SourceAllAnime {
		return scr.GetStreamURL(anime.URL, episode.Number, opts.Quality, opts.Mode)
	}

	// For AnimeFire, the episode URL is the direct episode page
	return scr.GetStreamURL(episode.URL)
}

// NewClientForTest creates a Client backed by a custom Manager. Only for tests.
func NewClientForTest(manager Manager) *Client {
	return &Client{manager: manager}
}

// GetAvailableSources returns a list of all available scraper sources.
func (c *Client) GetAvailableSources() []types.Source {
	return []types.Source{
		types.SourceAllAnime,
		types.SourceAnimeFire,
	}
}
