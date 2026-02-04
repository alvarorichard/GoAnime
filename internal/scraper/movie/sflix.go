// Package movie provides movie-specific scraping functionality
// This file provides SFlix movie client as a thin wrapper around the main scraper package
package movie

import (
	"context"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// Re-export SFlix types from main scraper package for backwards compatibility
type (
	SFlixMedia         = scraper.SFlixMedia
	SFlixSeason        = scraper.SFlixSeason
	SFlixEpisode       = scraper.SFlixEpisode
	SFlixServer        = scraper.SFlixServer
	SFlixStreamInfo    = scraper.SFlixStreamInfo
	SFlixQualityOption = scraper.SFlixQualityOption
	SFlixSubtitle      = scraper.SFlixSubtitle
	SFlixVideoSources  = scraper.SFlixVideoSources
	SFlixSource        = scraper.SFlixSource
)

// SFlixMovieClient wraps the main SFlix client for movie-specific operations
type SFlixMovieClient struct {
	client *scraper.SFlixClient
}

// NewSFlixMovieClient creates a new SFlix movie client
func NewSFlixMovieClient() *SFlixMovieClient {
	return &SFlixMovieClient{
		client: scraper.NewSFlixClient(),
	}
}

// SearchMovies searches for movies only
func (sc *SFlixMovieClient) SearchMovies(query string) ([]*SFlixMedia, error) {
	return sc.SearchMoviesWithContext(context.Background(), query)
}

// SearchMoviesWithContext searches for movies with context
func (sc *SFlixMovieClient) SearchMoviesWithContext(ctx context.Context, query string) ([]*SFlixMedia, error) {
	results, err := sc.client.SearchMediaWithContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Filter to movies only
	var movies []*SFlixMedia
	for _, r := range results {
		if r.Type == MediaTypeMovie {
			movies = append(movies, r)
		}
	}
	return movies, nil
}

// GetMovieInfo gets detailed info for a movie
func (sc *SFlixMovieClient) GetMovieInfo(id string) (*SFlixMedia, error) {
	return sc.client.GetInfo(id)
}

// GetMovieInfoWithContext gets detailed info with context
func (sc *SFlixMovieClient) GetMovieInfoWithContext(ctx context.Context, id string) (*SFlixMedia, error) {
	return sc.client.GetInfoWithContext(ctx, id)
}

// GetMovieServers gets available servers for a movie
func (sc *SFlixMovieClient) GetMovieServers(mediaID string) ([]SFlixServer, error) {
	return sc.client.GetServers(mediaID, true)
}

// GetMovieServersWithContext gets movie servers with context
func (sc *SFlixMovieClient) GetMovieServersWithContext(ctx context.Context, mediaID string) ([]SFlixServer, error) {
	return sc.client.GetServersWithContext(ctx, mediaID, true)
}

// GetMovieSources gets video sources for a movie
func (sc *SFlixMovieClient) GetMovieSources(mediaID string) (*SFlixVideoSources, error) {
	return sc.client.GetSources(mediaID, true)
}

// GetMovieSourcesWithContext gets movie sources with context
func (sc *SFlixMovieClient) GetMovieSourcesWithContext(ctx context.Context, mediaID string) (*SFlixVideoSources, error) {
	return sc.client.GetSourcesWithContext(ctx, mediaID, true)
}

// GetAvailableQualities returns available video qualities for a movie
func (sc *SFlixMovieClient) GetAvailableQualities(mediaID string) ([]Quality, error) {
	return sc.client.GetAvailableQualities(mediaID, true)
}

// GetAvailableQualitiesWithContext returns qualities with context
func (sc *SFlixMovieClient) GetAvailableQualitiesWithContext(ctx context.Context, mediaID string) ([]Quality, error) {
	return sc.client.GetAvailableQualitiesWithContext(ctx, mediaID, true)
}

// GetStreamURL gets the stream URL for a movie with specified quality
func (sc *SFlixMovieClient) GetStreamURL(mediaID string, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	media := &SFlixMedia{
		ID:   mediaID,
		Type: MediaTypeMovie,
	}
	return sc.client.GetStreamURL(media, nil, "Vidcloud", quality, subsLanguage)
}

// GetStreamURLWithContext gets stream URL with context
func (sc *SFlixMovieClient) GetStreamURLWithContext(ctx context.Context, mediaID string, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	media := &SFlixMedia{
		ID:   mediaID,
		Type: MediaTypeMovie,
	}
	return sc.client.GetStreamURLWithContext(ctx, media, nil, "Vidcloud", quality, subsLanguage)
}

// SelectBestQuality selects the best available quality from sources
func (sc *SFlixMovieClient) SelectBestQuality(sources *SFlixVideoSources, preferred Quality) *SFlixSource {
	return sc.client.SelectBestQuality(sources, preferred)
}

// HealthCheck checks if SFlix is accessible
func (sc *SFlixMovieClient) HealthCheck(ctx context.Context) error {
	return sc.client.HealthCheck(ctx)
}

// ClearCache clears all caches
func (sc *SFlixMovieClient) ClearCache() {
	sc.client.ClearCache()
}

// GetTrending gets trending movies
func (sc *SFlixMovieClient) GetTrending() ([]*SFlixMedia, error) {
	results, err := sc.client.GetTrending()
	if err != nil {
		return nil, err
	}

	// Filter to movies only
	var movies []*SFlixMedia
	for _, r := range results {
		if r.Type == MediaTypeMovie {
			movies = append(movies, r)
		}
	}
	return movies, nil
}

// GetRecent gets recent movies
func (sc *SFlixMovieClient) GetRecent() ([]*SFlixMedia, error) {
	return sc.client.GetRecentMovies()
}

// SFlixToAnimeModel converts SFlixMedia to models.Anime
func SFlixToAnimeModel(m *SFlixMedia) *models.Anime {
	return m.ToAnimeModel()
}

// SFlixToMedia converts SFlixMedia to models.Media
func SFlixToMedia(m *SFlixMedia) *models.Media {
	return m.ToMedia()
}

// SFlixToStreamInfo converts SFlixStreamInfo to models.StreamInfo
func SFlixToStreamInfo(s *SFlixStreamInfo) *models.StreamInfo {
	return s.ToStreamInfo()
}
