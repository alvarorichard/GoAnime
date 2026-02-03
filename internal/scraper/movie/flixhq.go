// Package movie provides movie-specific scraping functionality
// This is a thin wrapper around the main scraper package for backwards compatibility
package movie

import (
	"context"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// Re-export types from main scraper package for backwards compatibility
type (
	MediaType           = scraper.MediaType
	Quality             = scraper.Quality
	StreamType          = scraper.StreamType
	ServerName          = scraper.ServerName
	FlixHQMedia         = scraper.FlixHQMedia
	FlixHQSeason        = scraper.FlixHQSeason
	FlixHQEpisode       = scraper.FlixHQEpisode
	FlixHQServer        = scraper.FlixHQServer
	FlixHQStreamInfo    = scraper.FlixHQStreamInfo
	FlixHQQualityOption = scraper.FlixHQQualityOption
	FlixHQSubtitle      = scraper.FlixHQSubtitle
	FlixHQVideoSources  = scraper.FlixHQVideoSources
	FlixHQSource        = scraper.FlixHQSource
)

// Re-export constants
const (
	MediaTypeMovie = scraper.MediaTypeMovie
	MediaTypeTV    = scraper.MediaTypeTV

	QualityAuto = scraper.QualityAuto
	Quality360  = scraper.Quality360
	Quality480  = scraper.Quality480
	Quality720  = scraper.Quality720
	Quality1080 = scraper.Quality1080
	QualityBest = scraper.QualityBest

	StreamTypeHLS = scraper.StreamTypeHLS
	StreamTypeMP4 = scraper.StreamTypeMP4

	ServerVidcloud  = scraper.ServerVidcloud
	ServerUpCloud   = scraper.ServerUpCloud
	ServerVoe       = scraper.ServerVoe
	ServerMixDrop   = scraper.ServerMixDrop
	ServerFilelions = scraper.ServerFilelions
)

// MovieClient wraps the main FlixHQ client for movie-specific operations
type MovieClient struct {
	client *scraper.FlixHQClient
}

// NewMovieClient creates a new movie client
func NewMovieClient() *MovieClient {
	return &MovieClient{
		client: scraper.NewFlixHQClient(),
	}
}

// SearchMovies searches for movies only
func (mc *MovieClient) SearchMovies(query string) ([]*FlixHQMedia, error) {
	return mc.SearchMoviesWithContext(context.Background(), query)
}

// SearchMoviesWithContext searches for movies with context
func (mc *MovieClient) SearchMoviesWithContext(ctx context.Context, query string) ([]*FlixHQMedia, error) {
	results, err := mc.client.SearchMediaWithContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Filter to movies only
	var movies []*FlixHQMedia
	for _, r := range results {
		if r.Type == MediaTypeMovie {
			movies = append(movies, r)
		}
	}
	return movies, nil
}

// GetMovieInfo gets detailed info for a movie
func (mc *MovieClient) GetMovieInfo(id string) (*FlixHQMedia, error) {
	return mc.client.GetInfo(id)
}

// GetMovieInfoWithContext gets detailed info with context
func (mc *MovieClient) GetMovieInfoWithContext(ctx context.Context, id string) (*FlixHQMedia, error) {
	return mc.client.GetInfoWithContext(ctx, id)
}

// GetMovieServers gets available servers for a movie
func (mc *MovieClient) GetMovieServers(mediaID string) ([]FlixHQServer, error) {
	return mc.client.GetServers(mediaID, true)
}

// GetMovieServersWithContext gets movie servers with context
func (mc *MovieClient) GetMovieServersWithContext(ctx context.Context, mediaID string) ([]FlixHQServer, error) {
	return mc.client.GetServersWithContext(ctx, mediaID, true)
}

// GetMovieSources gets video sources for a movie
func (mc *MovieClient) GetMovieSources(mediaID string) (*FlixHQVideoSources, error) {
	return mc.client.GetSources(mediaID, true)
}

// GetMovieSourcesWithContext gets movie sources with context
func (mc *MovieClient) GetMovieSourcesWithContext(ctx context.Context, mediaID string) (*FlixHQVideoSources, error) {
	return mc.client.GetSourcesWithContext(ctx, mediaID, true)
}

// GetAvailableQualities returns available video qualities for a movie
func (mc *MovieClient) GetAvailableQualities(mediaID string) ([]Quality, error) {
	return mc.client.GetAvailableQualities(mediaID, true)
}

// GetAvailableQualitiesWithContext returns qualities with context
func (mc *MovieClient) GetAvailableQualitiesWithContext(ctx context.Context, mediaID string) ([]Quality, error) {
	return mc.client.GetAvailableQualitiesWithContext(ctx, mediaID, true)
}

// GetStreamURL gets the stream URL for a movie with specified quality
func (mc *MovieClient) GetStreamURL(mediaID string, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	media := &FlixHQMedia{
		ID:   mediaID,
		Type: MediaTypeMovie,
	}
	return mc.client.GetStreamURL(media, nil, "Vidcloud", quality, subsLanguage)
}

// GetStreamURLWithContext gets stream URL with context
func (mc *MovieClient) GetStreamURLWithContext(ctx context.Context, mediaID string, quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	media := &FlixHQMedia{
		ID:   mediaID,
		Type: MediaTypeMovie,
	}
	return mc.client.GetStreamURLWithContext(ctx, media, nil, "Vidcloud", quality, subsLanguage)
}

// SelectBestQuality selects the best available quality from sources
func (mc *MovieClient) SelectBestQuality(sources *FlixHQVideoSources, preferred Quality) *FlixHQSource {
	return mc.client.SelectBestQuality(sources, preferred)
}

// HealthCheck checks if FlixHQ is accessible
func (mc *MovieClient) HealthCheck(ctx context.Context) error {
	return mc.client.HealthCheck(ctx)
}

// ClearCache clears all caches
func (mc *MovieClient) ClearCache() {
	mc.client.ClearCache()
}

// GetTrending gets trending movies
func (mc *MovieClient) GetTrending() ([]*FlixHQMedia, error) {
	results, err := mc.client.GetTrending()
	if err != nil {
		return nil, err
	}

	// Filter to movies only
	var movies []*FlixHQMedia
	for _, r := range results {
		if r.Type == MediaTypeMovie {
			movies = append(movies, r)
		}
	}
	return movies, nil
}

// GetRecent gets recent movies
func (mc *MovieClient) GetRecent() ([]*FlixHQMedia, error) {
	return mc.client.GetRecentMovies()
}

// ToAnimeModel converts FlixHQMedia to models.Anime
func ToAnimeModel(m *FlixHQMedia) *models.Anime {
	return m.ToAnimeModel()
}

// ToMedia converts FlixHQMedia to models.Media
func ToMedia(m *FlixHQMedia) *models.Media {
	return m.ToMedia()
}

// ToStreamInfo converts FlixHQStreamInfo to models.StreamInfo
func ToStreamInfo(s *FlixHQStreamInfo) *models.StreamInfo {
	return s.ToStreamInfo()
}
