// Package scraper provides unified media handling for movie and TV sources.
package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
)

// MediaManager provides the subset of movie/TV helpers still used by download flows.
type MediaManager struct {
	scraperManager *ScraperManager
	flixhqClient   *FlixHQClient
	sflixClient    *SFlixClient
}

// Timeout configurations for movie/TV searches.
const (
	movieSearchTimeout = 6 * time.Second
	earlyReturnWait    = 800 * time.Millisecond
)

type movieSearchResult struct {
	source string
	flixhq []*FlixHQMedia
	sflix  []*SFlixMedia
	err    error
}

// NewMediaManager creates a new MediaManager.
func NewMediaManager() *MediaManager {
	sm := NewScraperManager()

	var flixhqClient *FlixHQClient
	if adapter, ok := sm.scrapers[FlixHQType].(*FlixHQAdapter); ok {
		flixhqClient = adapter.client
	} else {
		flixhqClient = NewFlixHQClient()
	}

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

// SearchMoviesAndTV searches both FlixHQ and SFlix for movies and TV shows.
func (mm *MediaManager) SearchMoviesAndTV(query string) ([]*FlixHQMedia, error) {
	return mm.SearchAllMovieSources(query)
}

// SearchAllMovieSources searches both FlixHQ and SFlix concurrently with timeout.
func (mm *MediaManager) SearchAllMovieSources(query string) ([]*FlixHQMedia, error) {
	ctx, cancel := context.WithTimeout(context.Background(), movieSearchTimeout)
	defer cancel()

	resultChan := make(chan movieSearchResult, 2)

	go func() {
		results, err := mm.flixhqClient.SearchMedia(query)
		resultChan <- movieSearchResult{source: "FlixHQ", flixhq: results, err: err}
	}()

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
						combined = append(combined, mm.ConvertSFlixToFlixHQ(r))
					}
					hasResults = true
				}
				mutex.Unlock()

				if hasResults && earlyTimer == nil {
					earlyTimer = time.After(earlyReturnWait)
				}
			}

			if receivedCount >= 2 {
				goto done
			}
		case <-earlyTimer:
			if hasResults {
				goto done
			}
		case <-ctx.Done():
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

// ConvertSFlixToFlixHQ converts SFlixMedia to FlixHQMedia format for unified handling.
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

// GetTVSeasons gets all seasons for a TV show.
func (mm *MediaManager) GetTVSeasons(mediaID string) ([]FlixHQSeason, error) {
	return mm.flixhqClient.GetSeasons(mediaID)
}

// GetSFlixTVSeasons gets all seasons for a TV show from SFlix.
func (mm *MediaManager) GetSFlixTVSeasons(mediaID string) ([]SFlixSeason, error) {
	return mm.sflixClient.GetSeasons(mediaID)
}

// GetTVEpisodes gets all episodes for a season.
func (mm *MediaManager) GetTVEpisodes(seasonID string) ([]FlixHQEpisode, error) {
	return mm.flixhqClient.GetEpisodes(seasonID)
}

// GetSFlixTVEpisodes gets all episodes for a season from SFlix.
func (mm *MediaManager) GetSFlixTVEpisodes(seasonID string) ([]SFlixEpisode, error) {
	return mm.sflixClient.GetEpisodes(seasonID)
}

// GetSFlixMovieStreamInfo gets stream information for a movie from SFlix.
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

// GetTVEpisodeStreamInfo gets stream information for a TV episode.
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

// GetSFlixTVEpisodeStreamInfo gets stream information for a TV episode from SFlix.
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

// ConvertFlixHQToAnime converts FlixHQ media list to Anime models for unified handling.
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

// GetMovieStreamWithQuality gets the stream URL for a movie with a specific quality.
func (mm *MediaManager) GetMovieStreamWithQuality(mediaID string, quality Quality, subsLanguage string) (*FlixHQStreamInfo, error) {
	return mm.flixhqClient.GetMovieStreamWithQuality(mediaID, quality, subsLanguage)
}
