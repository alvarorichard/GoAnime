// Package scraper provides unified media handling for anime, movies, and TV shows
package scraper

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// MediaManager provides a unified interface for all media types
type MediaManager struct {
	scraperManager *ScraperManager
	flixhqClient   *FlixHQClient
}

// NewMediaManager creates a new MediaManager
func NewMediaManager() *MediaManager {
	sm := NewScraperManager()

	// Get the FlixHQ client from the adapter
	var flixhqClient *FlixHQClient
	if adapter, ok := sm.scrapers[FlixHQType].(*FlixHQAdapter); ok {
		flixhqClient = adapter.client
	} else {
		flixhqClient = NewFlixHQClient()
	}

	return &MediaManager{
		scraperManager: sm,
		flixhqClient:   flixhqClient,
	}
}

// SearchAll searches across all sources (anime + movies/TV)
func (mm *MediaManager) SearchAll(query string) ([]*models.Anime, error) {
	return mm.scraperManager.SearchAnime(query, nil)
}

// SearchAnimeOnly searches only anime sources
func (mm *MediaManager) SearchAnimeOnly(query string) ([]*models.Anime, error) {
	var allResults []*models.Anime

	// Search AllAnime
	allAnimeType := AllAnimeType
	animeResults, err := mm.scraperManager.SearchAnime(query, &allAnimeType)
	if err == nil {
		allResults = append(allResults, animeResults...)
	}

	// Search AnimeFire
	animefireType := AnimefireType
	animefireResults, err := mm.scraperManager.SearchAnime(query, &animefireType)
	if err == nil {
		allResults = append(allResults, animefireResults...)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no anime found with name: %s", query)
	}

	return allResults, nil
}

// SearchMoviesAndTV searches only FlixHQ for movies and TV shows
func (mm *MediaManager) SearchMoviesAndTV(query string) ([]*FlixHQMedia, error) {
	return mm.flixhqClient.SearchMedia(query)
}

// GetTrendingMovies gets trending movies from FlixHQ
func (mm *MediaManager) GetTrendingMovies() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetTrending()
}

// GetRecentMovies gets recent movies from FlixHQ
func (mm *MediaManager) GetRecentMovies() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetRecentMovies()
}

// GetRecentTV gets recent TV shows from FlixHQ
func (mm *MediaManager) GetRecentTV() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetRecentTV()
}

// GetTVSeasons gets all seasons for a TV show
func (mm *MediaManager) GetTVSeasons(mediaID string) ([]FlixHQSeason, error) {
	return mm.flixhqClient.GetSeasons(mediaID)
}

// GetTVEpisodes gets all episodes for a season
func (mm *MediaManager) GetTVEpisodes(seasonID string) ([]FlixHQEpisode, error) {
	return mm.flixhqClient.GetEpisodes(seasonID)
}

// GetMovieStreamInfo gets stream information for a movie
func (mm *MediaManager) GetMovieStreamInfo(mediaID, provider, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.flixhqClient.GetMovieServerID(mediaID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie server: %w", err)
	}

	embedLink, err := mm.flixhqClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.flixhqClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
}

// GetTVEpisodeStreamInfo gets stream information for a TV episode
func (mm *MediaManager) GetTVEpisodeStreamInfo(dataID, provider, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.flixhqClient.GetEpisodeServerID(dataID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode server: %w", err)
	}

	embedLink, err := mm.flixhqClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.flixhqClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
}

// GetAnimeStreamURL gets stream URL for anime episodes
func (mm *MediaManager) GetAnimeStreamURL(anime *models.Anime, episodeNum string, quality, mode string) (string, map[string]string, error) {
	source := strings.ToLower(anime.Source)

	util.Debug("Getting stream URL", "source", source, "anime", anime.Name, "episode", episodeNum)

	switch {
	case strings.Contains(source, "allanime"):
		scraper, err := mm.scraperManager.GetScraper(AllAnimeType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	case strings.Contains(source, "animefire"):
		scraper, err := mm.scraperManager.GetScraper(AnimefireType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	case strings.Contains(source, "animedrive"):
		scraper, err := mm.scraperManager.GetScraper(AnimeDriveType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	default:
		return "", nil, fmt.Errorf("unknown source: %s", anime.Source)
	}
}

// ConvertFlixHQToAnime converts FlixHQ media list to Anime models for unified handling
func ConvertFlixHQToAnime(media []*FlixHQMedia) []*models.Anime {
	var animes []*models.Anime
	for _, m := range media {
		anime := m.ToAnimeModel()
		if m.Type == MediaTypeMovie {
			anime.MediaType = models.MediaTypeMovie
		} else {
			anime.MediaType = models.MediaTypeTV
		}
		anime.Year = m.Year
		animes = append(animes, anime)
	}
	return animes
}

// ConvertFlixHQEpisodesToEpisodes converts FlixHQ episodes to Episode models
func ConvertFlixHQEpisodesToEpisodes(episodes []FlixHQEpisode) []models.Episode {
	var eps []models.Episode
	for _, e := range episodes {
		eps = append(eps, e.ToEpisodeModel())
	}
	return eps
}

// GetScraperManager returns the underlying scraper manager for advanced usage
func (mm *MediaManager) GetScraperManager() *ScraperManager {
	return mm.scraperManager
}

// GetFlixHQClient returns the FlixHQ client for direct access
func (mm *MediaManager) GetFlixHQClient() *FlixHQClient {
	return mm.flixhqClient
}
