// Package movie provides TMDB (The Movie Database) API integration for movie/TV metadata
package movie

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// TMDB API base URL
	TMDBBaseURL = "https://api.themoviedb.org/3"
	// TMDB image base URL
	TMDBImageBaseURL = "https://image.tmdb.org/t/p"
)

// TMDBClient handles interactions with TMDB API
type TMDBClient struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	imageBase string
}

// NewTMDBClient creates a new TMDB client
// Requires TMDB_API_KEY environment variable to be set
// Get your free API key at https://www.themoviedb.org/settings/api
func NewTMDBClient() *TMDBClient {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		util.Debug("TMDB_API_KEY not set, TMDB enrichment will be disabled")
	}

	return &TMDBClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey:    apiKey,
		baseURL:   TMDBBaseURL,
		imageBase: TMDBImageBaseURL,
	}
}

// IsConfigured returns true if the TMDB API key is configured
func (c *TMDBClient) IsConfigured() bool {
	return c.apiKey != ""
}

// SearchMulti searches for both movies and TV shows
func (c *TMDBClient) SearchMulti(query string) (*models.TMDBSearchResult, error) {
	endpoint := fmt.Sprintf("%s/search/multi?query=%s&include_adult=false&language=en-US&page=1",
		c.baseURL, url.QueryEscape(query))

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("TMDB search failed: %w", err)
	}

	var result models.TMDBSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TMDB response: %w", err)
	}

	// Filter only movies and TV shows
	var filtered []models.TMDBMedia
	for _, item := range result.Results {
		if item.MediaType == "movie" || item.MediaType == "tv" {
			filtered = append(filtered, item)
		}
	}
	result.Results = filtered

	return &result, nil
}

// SearchMovies searches for movies only
func (c *TMDBClient) SearchMovies(query string) (*models.TMDBSearchResult, error) {
	endpoint := fmt.Sprintf("%s/search/movie?query=%s&include_adult=false&language=en-US&page=1",
		c.baseURL, url.QueryEscape(query))

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("TMDB movie search failed: %w", err)
	}

	var result models.TMDBSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TMDB response: %w", err)
	}

	// Mark all results as movies
	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	return &result, nil
}

// SearchTV searches for TV shows only
func (c *TMDBClient) SearchTV(query string) (*models.TMDBSearchResult, error) {
	endpoint := fmt.Sprintf("%s/search/tv?query=%s&include_adult=false&language=en-US&page=1",
		c.baseURL, url.QueryEscape(query))

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("TMDB TV search failed: %w", err)
	}

	var result models.TMDBSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TMDB response: %w", err)
	}

	// Mark all results as TV
	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	return &result, nil
}

// GetMovieDetails gets detailed information about a movie
func (c *TMDBClient) GetMovieDetails(movieID int) (*models.TMDBDetails, error) {
	endpoint := fmt.Sprintf("%s/movie/%d?language=en-US", c.baseURL, movieID)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie details: %w", err)
	}

	var details models.TMDBDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to parse movie details: %w", err)
	}

	return &details, nil
}

// GetTVDetails gets detailed information about a TV show
func (c *TMDBClient) GetTVDetails(tvID int) (*models.TMDBDetails, error) {
	endpoint := fmt.Sprintf("%s/tv/%d?language=en-US", c.baseURL, tvID)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV details: %w", err)
	}

	var details models.TMDBDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to parse TV details: %w", err)
	}

	return &details, nil
}

// GetTVSeasons gets season information for a TV show
func (c *TMDBClient) GetTVSeasons(tvID int) ([]models.TMDBSeason, error) {
	endpoint := fmt.Sprintf("%s/tv/%d?language=en-US", c.baseURL, tvID)
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV seasons: %w", err)
	}

	var result struct {
		Seasons []models.TMDBSeason `json:"seasons"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse seasons: %w", err)
	}

	return result.Seasons, nil
}

// GetSeasonEpisodes gets episodes for a specific season
func (c *TMDBClient) GetSeasonEpisodes(tvID, seasonNumber int) ([]models.TMDBEpisode, error) {
	endpoint := fmt.Sprintf("%s/tv/%d/season/%d?language=en-US", c.baseURL, tvID, seasonNumber)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get season episodes: %w", err)
	}

	var result struct {
		Episodes []models.TMDBEpisode `json:"episodes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse episodes: %w", err)
	}

	return result.Episodes, nil
}

// GetCredits gets cast and crew for a movie or TV show
func (c *TMDBClient) GetCredits(mediaType string, mediaID int) (*models.TMDBCredits, error) {
	endpoint := fmt.Sprintf("%s/%s/%d/credits?language=en-US", c.baseURL, mediaType, mediaID)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get credits: %w", err)
	}

	var credits models.TMDBCredits
	if err := json.Unmarshal(body, &credits); err != nil {
		return nil, fmt.Errorf("failed to parse credits: %w", err)
	}

	return &credits, nil
}

// FindByIMDBID finds a movie or TV show by IMDB ID
func (c *TMDBClient) FindByIMDBID(imdbID string) (*models.TMDBMedia, error) {
	endpoint := fmt.Sprintf("%s/find/%s?external_source=imdb_id", c.baseURL, imdbID)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to find by IMDB ID: %w", err)
	}

	var result struct {
		MovieResults []models.TMDBMedia `json:"movie_results"`
		TVResults    []models.TMDBMedia `json:"tv_results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse find response: %w", err)
	}

	if len(result.MovieResults) > 0 {
		result.MovieResults[0].MediaType = "movie"
		return &result.MovieResults[0], nil
	}
	if len(result.TVResults) > 0 {
		result.TVResults[0].MediaType = "tv"
		return &result.TVResults[0], nil
	}

	return nil, fmt.Errorf("no results found for IMDB ID: %s", imdbID)
}

// GetTrending gets trending movies and TV shows
func (c *TMDBClient) GetTrending(mediaType string, timeWindow string) (*models.TMDBSearchResult, error) {
	if mediaType == "" {
		mediaType = "all"
	}
	if timeWindow == "" {
		timeWindow = "week"
	}

	endpoint := fmt.Sprintf("%s/trending/%s/%s?language=en-US", c.baseURL, mediaType, timeWindow)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending: %w", err)
	}

	var result models.TMDBSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse trending response: %w", err)
	}

	return &result, nil
}

// GetPopular gets popular movies or TV shows
func (c *TMDBClient) GetPopular(mediaType string) (*models.TMDBSearchResult, error) {
	endpoint := fmt.Sprintf("%s/%s/popular?language=en-US&page=1", c.baseURL, mediaType)

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular: %w", err)
	}

	var result models.TMDBSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse popular response: %w", err)
	}

	// Mark media type
	for i := range result.Results {
		result.Results[i].MediaType = mediaType
	}

	return &result, nil
}

// makeRequest performs an authenticated request to TMDB API
func (c *TMDBClient) makeRequest(endpoint string) ([]byte, error) {
	// Add API key as query parameter
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	endpointWithKey := endpoint + separator + "api_key=" + c.apiKey

	req, err := http.NewRequest("GET", endpointWithKey, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// GetImageURL returns the full URL for a TMDB image
func (c *TMDBClient) GetImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	if size == "" {
		size = "w500"
	}
	return fmt.Sprintf("%s/%s%s", c.imageBase, size, path)
}
