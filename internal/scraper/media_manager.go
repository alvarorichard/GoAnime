// Package scraper provides unified media handling for anime, movies, and TV shows
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

// MediaManager provides a unified interface for all media types
type MediaManager struct {
	scraperManager *ScraperManager
	flixhqClient   *FlixHQClient
	sflixClient    *SFlixClient
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

	// Get the SFlix client from the adapter
	var sflixClient *SFlixClient
	if adapter, ok := sm.scrapers[SFlixType].(*SFlixAdapter); ok {
		sflixClient = adapter.client
	} else {
		sflixClient = NewSFlixClient()
	}

	return &MediaManager{
		scraperManager: sm,
		flixhqClient:   flixhqClient,
		sflixClient:    sflixClient,
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

// Timeout configurations for movie/TV searches
const (
	// movieSearchTimeout is the maximum time to wait for all sources
	movieSearchTimeout = 6 * time.Second
	// earlyReturnWait is how long to wait after first results before returning
	earlyReturnWait = 800 * time.Millisecond
)

// SearchMoviesAndTV searches both FlixHQ and SFlix for movies and TV shows
// Returns results from both sources combined with timeout and early return
func (mm *MediaManager) SearchMoviesAndTV(query string) ([]*FlixHQMedia, error) {
	return mm.SearchAllMovieSources(query)
}

// SearchFlixHQOnly searches only FlixHQ for movies and TV shows
func (mm *MediaManager) SearchFlixHQOnly(query string) ([]*FlixHQMedia, error) {
	return mm.flixhqClient.SearchMedia(query)
}

// SearchSFlixMoviesAndTV searches SFlix for movies and TV shows
func (mm *MediaManager) SearchSFlixMoviesAndTV(query string) ([]*SFlixMedia, error) {
	return mm.sflixClient.SearchMedia(query)
}

// movieSearchResult holds result from a source
type movieSearchResult struct {
	source string
	flixhq []*FlixHQMedia
	sflix  []*SFlixMedia
	err    error
}

// SearchAllMovieSources searches both FlixHQ and SFlix concurrently with timeout
// Returns as soon as we have results, doesn't wait for slow sources
func (mm *MediaManager) SearchAllMovieSources(query string) ([]*FlixHQMedia, error) {
	ctx, cancel := context.WithTimeout(context.Background(), movieSearchTimeout)
	defer cancel()

	resultChan := make(chan movieSearchResult, 2)

	// Launch FlixHQ search
	go func() {
		results, err := mm.flixhqClient.SearchMedia(query)
		resultChan <- movieSearchResult{source: "FlixHQ", flixhq: results, err: err}
	}()

	// Launch SFlix search
	go func() {
		results, err := mm.sflixClient.SearchMedia(query)
		resultChan <- movieSearchResult{source: "SFlix", sflix: results, err: err}
	}()

	var (
		combined      []*FlixHQMedia
		mutex         sync.Mutex
		receivedCount int
		hasResults    bool
		earlyTimer    <-chan time.Time
		errors        []string
	)

	for {
		select {
		case res := <-resultChan:
			receivedCount++

			if res.err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", res.source, res.err))
			} else {
				mutex.Lock()
				if res.source == "FlixHQ" && len(res.flixhq) > 0 {
					for _, r := range res.flixhq {
						r.Source = "FlixHQ"
						combined = append(combined, r)
					}
					hasResults = true
				} else if res.source == "SFlix" && len(res.sflix) > 0 {
					for _, r := range res.sflix {
						converted := mm.ConvertSFlixToFlixHQ(r)
						combined = append(combined, converted)
					}
					hasResults = true
				}
				mutex.Unlock()

				// Start early return timer on first results
				if hasResults && earlyTimer == nil {
					earlyTimer = time.After(earlyReturnWait)
				}
			}

			// All sources responded
			if receivedCount >= 2 {
				goto done
			}

		case <-earlyTimer:
			// Got results, waited enough, return early
			if hasResults {
				goto done
			}

		case <-ctx.Done():
			// Timeout - return what we have
			goto done
		}
	}

done:
	if len(combined) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("no results found: %s", strings.Join(errors, "; "))
		}
		return nil, fmt.Errorf("no results found for query: %s", query)
	}

	return combined, nil
}

// ConvertSFlixToFlixHQ converts SFlixMedia to FlixHQMedia format for unified handling
func (mm *MediaManager) ConvertSFlixToFlixHQ(sflix *SFlixMedia) *FlixHQMedia {
	return &FlixHQMedia{
		ID:          sflix.ID,
		Title:       sflix.Title,
		URL:         sflix.URL,
		ImageURL:    sflix.ImageURL,
		Type:        sflix.Type,
		Year:        sflix.Year,
		ReleaseDate: sflix.ReleaseDate,
		Quality:     sflix.Quality,
		Duration:    sflix.Duration,
		Description: sflix.Description,
		Genres:      sflix.Genres,
		Country:     sflix.Country,
		Production:  sflix.Production,
		Casts:       sflix.Casts,
		Source:      "SFlix",
	}
}

// GetTrendingMovies gets trending movies from both FlixHQ and SFlix
func (mm *MediaManager) GetTrendingMovies() ([]*FlixHQMedia, error) {
	return mm.GetAllTrendingMovies()
}

// GetFlixHQTrendingMovies gets trending movies from FlixHQ only
func (mm *MediaManager) GetFlixHQTrendingMovies() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetTrending()
}

// GetSFlixTrendingMovies gets trending movies from SFlix
func (mm *MediaManager) GetSFlixTrendingMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetTrending()
}

// GetAllTrendingMovies gets trending movies from both sources with fast return
func (mm *MediaManager) GetAllTrendingMovies() ([]*FlixHQMedia, error) {
	return mm.fetchFromBothSources(
		func() ([]*FlixHQMedia, error) { return mm.flixhqClient.GetTrending() },
		func() ([]*SFlixMedia, error) { return mm.sflixClient.GetTrending() },
	)
}

// GetRecentMovies gets recent movies from both FlixHQ and SFlix
func (mm *MediaManager) GetRecentMovies() ([]*FlixHQMedia, error) {
	return mm.GetAllRecentMovies()
}

// GetFlixHQRecentMovies gets recent movies from FlixHQ only
func (mm *MediaManager) GetFlixHQRecentMovies() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetRecentMovies()
}

// GetSFlixRecentMovies gets recent movies from SFlix
func (mm *MediaManager) GetSFlixRecentMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentMovies()
}

// GetAllRecentMovies gets recent movies from both sources with fast return
func (mm *MediaManager) GetAllRecentMovies() ([]*FlixHQMedia, error) {
	return mm.fetchFromBothSources(
		func() ([]*FlixHQMedia, error) { return mm.flixhqClient.GetRecentMovies() },
		func() ([]*SFlixMedia, error) { return mm.sflixClient.GetRecentMovies() },
	)
}

// GetRecentTV gets recent TV shows from both FlixHQ and SFlix
func (mm *MediaManager) GetRecentTV() ([]*FlixHQMedia, error) {
	return mm.GetAllRecentTV()
}

// GetFlixHQRecentTV gets recent TV shows from FlixHQ only
func (mm *MediaManager) GetFlixHQRecentTV() ([]*FlixHQMedia, error) {
	return mm.flixhqClient.GetRecentTV()
}

// GetSFlixRecentTV gets recent TV shows from SFlix
func (mm *MediaManager) GetSFlixRecentTV() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentTV()
}

// GetAllRecentTV gets recent TV shows from both sources with fast return
func (mm *MediaManager) GetAllRecentTV() ([]*FlixHQMedia, error) {
	return mm.fetchFromBothSources(
		func() ([]*FlixHQMedia, error) { return mm.flixhqClient.GetRecentTV() },
		func() ([]*SFlixMedia, error) { return mm.sflixClient.GetRecentTV() },
	)
}

// fetchFromBothSources is a generic helper that fetches from both sources with timeout
func (mm *MediaManager) fetchFromBothSources(
	flixhqFn func() ([]*FlixHQMedia, error),
	sflixFn func() ([]*SFlixMedia, error),
) ([]*FlixHQMedia, error) {
	ctx, cancel := context.WithTimeout(context.Background(), movieSearchTimeout)
	defer cancel()

	type result struct {
		source string
		flixhq []*FlixHQMedia
		sflix  []*SFlixMedia
		err    error
	}

	resultChan := make(chan result, 2)

	go func() {
		results, err := flixhqFn()
		resultChan <- result{source: "FlixHQ", flixhq: results, err: err}
	}()

	go func() {
		results, err := sflixFn()
		resultChan <- result{source: "SFlix", sflix: results, err: err}
	}()

	var (
		combined      []*FlixHQMedia
		receivedCount int
		hasResults    bool
		earlyTimer    <-chan time.Time
	)

	for {
		select {
		case res := <-resultChan:
			receivedCount++

			if res.err == nil {
				if res.source == "FlixHQ" && len(res.flixhq) > 0 {
					for _, r := range res.flixhq {
						r.Source = "FlixHQ"
						combined = append(combined, r)
					}
					hasResults = true
				} else if res.source == "SFlix" && len(res.sflix) > 0 {
					for _, r := range res.sflix {
						converted := mm.ConvertSFlixToFlixHQ(r)
						combined = append(combined, converted)
					}
					hasResults = true
				}

				if hasResults && earlyTimer == nil {
					earlyTimer = time.After(earlyReturnWait)
				}
			}

			if receivedCount >= 2 {
				return combined, nil
			}

		case <-earlyTimer:
			if hasResults {
				return combined, nil
			}

		case <-ctx.Done():
			if len(combined) > 0 {
				return combined, nil
			}
			return nil, fmt.Errorf("timeout fetching data")
		}
	}
}

// GetTVSeasons gets all seasons for a TV show
func (mm *MediaManager) GetTVSeasons(mediaID string) ([]FlixHQSeason, error) {
	return mm.flixhqClient.GetSeasons(mediaID)
}

// GetSFlixTVSeasons gets all seasons for a TV show from SFlix
func (mm *MediaManager) GetSFlixTVSeasons(mediaID string) ([]SFlixSeason, error) {
	return mm.sflixClient.GetSeasons(mediaID)
}

// GetTVEpisodes gets all episodes for a season
func (mm *MediaManager) GetTVEpisodes(seasonID string) ([]FlixHQEpisode, error) {
	return mm.flixhqClient.GetEpisodes(seasonID)
}

// GetSFlixTVEpisodes gets all episodes for a season from SFlix
func (mm *MediaManager) GetSFlixTVEpisodes(seasonID string) ([]SFlixEpisode, error) {
	return mm.sflixClient.GetEpisodes(seasonID)
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

// GetSFlixMovieStreamInfo gets stream information for a movie from SFlix
func (mm *MediaManager) GetSFlixMovieStreamInfo(mediaID, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.sflixClient.GetMovieServerID(mediaID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie server: %w", err)
	}

	embedLink, err := mm.sflixClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.sflixClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
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

// GetSFlixTVEpisodeStreamInfo gets stream information for a TV episode from SFlix
func (mm *MediaManager) GetSFlixTVEpisodeStreamInfo(dataID, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.sflixClient.GetEpisodeServerID(dataID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode server: %w", err)
	}

	embedLink, err := mm.sflixClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.sflixClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
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

// ConvertSFlixToAnime converts SFlix media list to Anime models for unified handling
func ConvertSFlixToAnime(media []*SFlixMedia) []*models.Anime {
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

// ConvertSFlixEpisodesToEpisodes converts SFlix episodes to Episode models
func ConvertSFlixEpisodesToEpisodes(episodes []SFlixEpisode) []models.Episode {
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

// GetSFlixClient returns the SFlix client for direct access
func (mm *MediaManager) GetSFlixClient() *SFlixClient {
	return mm.sflixClient
}

// GetMovieInfo gets detailed info for a movie or TV show
func (mm *MediaManager) GetMovieInfo(id string) (*FlixHQMedia, error) {
	return mm.flixhqClient.GetInfo(id)
}

// GetSFlixMovieInfo gets detailed info for a movie or TV show from SFlix
func (mm *MediaManager) GetSFlixMovieInfo(id string) (*SFlixMedia, error) {
	return mm.sflixClient.GetInfo(id)
}

// GetMovieInfoWithContext gets detailed info with context support
func (mm *MediaManager) GetMovieInfoWithContext(ctx context.Context, id string) (*FlixHQMedia, error) {
	return mm.flixhqClient.GetInfoWithContext(ctx, id)
}

// GetSFlixMovieInfoWithContext gets detailed info with context support
func (mm *MediaManager) GetSFlixMovieInfoWithContext(ctx context.Context, id string) (*SFlixMedia, error) {
	return mm.sflixClient.GetInfoWithContext(ctx, id)
}

// GetServers gets available streaming servers
func (mm *MediaManager) GetServers(episodeID string, isMovie bool) ([]FlixHQServer, error) {
	return mm.flixhqClient.GetServers(episodeID, isMovie)
}

// GetSFlixServers gets available streaming servers from SFlix
func (mm *MediaManager) GetSFlixServers(episodeID string, isMovie bool) ([]SFlixServer, error) {
	return mm.sflixClient.GetServers(episodeID, isMovie)
}

// GetServersWithContext gets available streaming servers with context
func (mm *MediaManager) GetServersWithContext(ctx context.Context, episodeID string, isMovie bool) ([]FlixHQServer, error) {
	return mm.flixhqClient.GetServersWithContext(ctx, episodeID, isMovie)
}

// GetSources gets video sources from all servers
func (mm *MediaManager) GetSources(episodeID string, isMovie bool) (*FlixHQVideoSources, error) {
	return mm.flixhqClient.GetSources(episodeID, isMovie)
}

// GetSourcesWithContext gets video sources with context support
func (mm *MediaManager) GetSourcesWithContext(ctx context.Context, episodeID string, isMovie bool) (*FlixHQVideoSources, error) {
	return mm.flixhqClient.GetSourcesWithContext(ctx, episodeID, isMovie)
}

// GetAvailableQualities returns available video qualities
func (mm *MediaManager) GetAvailableQualities(episodeID string, isMovie bool) ([]Quality, error) {
	return mm.flixhqClient.GetAvailableQualities(episodeID, isMovie)
}

// GetAvailableQualitiesWithContext returns available qualities with context
func (mm *MediaManager) GetAvailableQualitiesWithContext(ctx context.Context, episodeID string, isMovie bool) ([]Quality, error) {
	return mm.flixhqClient.GetAvailableQualitiesWithContext(ctx, episodeID, isMovie)
}

// GetStreamWithQuality gets stream URL with specific quality
func (mm *MediaManager) GetStreamWithQuality(episodeID string, isMovie bool, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return mm.GetStreamWithQualityWithContext(context.Background(), episodeID, isMovie, quality, subsLanguage)
}

// GetStreamWithQualityWithContext gets stream URL with specific quality and context
func (mm *MediaManager) GetStreamWithQualityWithContext(ctx context.Context, episodeID string, isMovie bool, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	sources, err := mm.flixhqClient.GetSourcesWithContext(ctx, episodeID, isMovie)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sources.Sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	// Select best quality based on preference
	selectedSource := mm.flixhqClient.SelectBestQuality(sources, quality)
	if selectedSource == nil {
		return nil, fmt.Errorf("no suitable quality found")
	}

	// Build stream info
	streamInfo := &FlixHQStreamInfo{
		VideoURL: selectedSource.URL,
		Quality:  string(quality),
		Referer:  mm.flixhqClient.baseURL,
		IsM3U8:   selectedSource.IsM3U8,
		Headers:  make(map[string]string),
	}
	streamInfo.Headers["Referer"] = mm.flixhqClient.baseURL

	if selectedSource.IsM3U8 {
		streamInfo.StreamType = StreamTypeHLS
	} else {
		streamInfo.StreamType = StreamTypeMP4
	}

	// Add subtitles
	for _, sub := range sources.Subtitles {
		// Filter by language if specified
		if subsLanguage != "" {
			if !strings.Contains(strings.ToLower(sub.Language), strings.ToLower(subsLanguage)) &&
				!strings.Contains(strings.ToLower(sub.Label), strings.ToLower(subsLanguage)) {
				continue
			}
		}
		streamInfo.Subtitles = append(streamInfo.Subtitles, sub)
	}

	// If filtering removed all subs, add them all back
	if subsLanguage != "" && len(streamInfo.Subtitles) == 0 {
		streamInfo.Subtitles = sources.Subtitles
	}

	return streamInfo, nil
}

// HealthCheck checks if FlixHQ is accessible
func (mm *MediaManager) HealthCheck(ctx context.Context) error {
	return mm.flixhqClient.HealthCheck(ctx)
}

// ClearCache clears all caches
func (mm *MediaManager) ClearCache() {
	mm.flixhqClient.ClearCache()
}

// GetMovieQualities fetches available qualities for a movie
func (mm *MediaManager) GetMovieQualities(mediaID string) ([]QualityOption, error) {
	return mm.flixhqClient.GetMovieQualities(mediaID)
}

// GetMovieQualitiesWithContext fetches available qualities for a movie with context
func (mm *MediaManager) GetMovieQualitiesWithContext(ctx context.Context, mediaID string) ([]QualityOption, error) {
	return mm.flixhqClient.GetMovieQualitiesWithContext(ctx, mediaID)
}

// GetEpisodeQualities fetches available qualities for a TV episode
func (mm *MediaManager) GetEpisodeQualities(dataID string) ([]QualityOption, error) {
	return mm.GetEpisodeQualitiesWithContext(context.Background(), dataID)
}

// GetEpisodeQualitiesWithContext fetches available qualities for a TV episode with context
func (mm *MediaManager) GetEpisodeQualitiesWithContext(ctx context.Context, dataID string) ([]QualityOption, error) {
	sources, err := mm.flixhqClient.GetSourcesWithContext(ctx, dataID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode sources: %w", err)
	}
	return mm.flixhqClient.sourcesToQualityOptions(sources), nil
}

// GetMovieStreamWithQuality gets the stream URL for a movie with a specific quality
func (mm *MediaManager) GetMovieStreamWithQuality(mediaID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return mm.flixhqClient.GetMovieStreamWithQuality(mediaID, quality, subsLanguage)
}

// GetMovieStreamWithQualityContext gets the stream URL for a movie with context support
func (mm *MediaManager) GetMovieStreamWithQualityContext(ctx context.Context, mediaID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return mm.flixhqClient.GetMovieStreamWithQualityContext(ctx, mediaID, quality, subsLanguage)
}

// GetEpisodeStreamWithQuality gets the stream URL for an episode with a specific quality
func (mm *MediaManager) GetEpisodeStreamWithQuality(dataID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return mm.GetEpisodeStreamWithQualityContext(context.Background(), dataID, quality, subsLanguage)
}

// GetEpisodeStreamWithQualityContext gets the stream URL for an episode with context support
func (mm *MediaManager) GetEpisodeStreamWithQualityContext(ctx context.Context, dataID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	sources, err := mm.flixhqClient.GetSourcesWithContext(ctx, dataID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode sources: %w", err)
	}

	if len(sources.Sources) == 0 {
		return nil, fmt.Errorf("no video sources available for this episode")
	}

	selectedSource := mm.flixhqClient.SelectBestQuality(sources, quality)
	if selectedSource == nil {
		return nil, fmt.Errorf("no matching quality found")
	}

	streamInfo := &FlixHQStreamInfo{
		VideoURL:  selectedSource.URL,
		Quality:   selectedSource.Quality,
		Referer:   mm.flixhqClient.baseURL,
		IsM3U8:    selectedSource.IsM3U8,
		Headers:   make(map[string]string),
		Qualities: mm.flixhqClient.sourcesToFlixHQQualityOptions(sources),
		Subtitles: sources.Subtitles,
	}
	streamInfo.Headers["Referer"] = mm.flixhqClient.baseURL

	if streamInfo.IsM3U8 {
		streamInfo.StreamType = StreamTypeHLS
	} else {
		streamInfo.StreamType = StreamTypeMP4
	}

	// Filter subtitles by language if specified
	if subsLanguage != "" && len(streamInfo.Subtitles) > 0 {
		streamInfo.Subtitles = mm.flixhqClient.filterSubtitlesByLanguage(streamInfo.Subtitles, subsLanguage)
	}

	return streamInfo, nil
}
