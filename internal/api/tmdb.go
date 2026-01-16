// Package api provides TMDB (The Movie Database) integration for movie/TV metadata
package api

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
	details, err := c.GetTVDetails(tvID)
	if err != nil {
		return nil, err
	}

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

	_ = details // Used for validation
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

	resp, err := c.client.Do(req)
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

// EnrichMediaWithTMDB enriches a media item with TMDB data
// Falls back to OMDb if TMDB API key is not configured
func EnrichMediaWithTMDB(media *models.Anime) error {
	if media.MediaType != models.MediaTypeMovie && media.MediaType != models.MediaTypeTV {
		return nil // Only enrich movies and TV shows
	}

	client := NewTMDBClient()

	// If TMDB is not configured, fall back to OMDb
	if !client.IsConfigured() {
		util.Debug("TMDB not configured, using OMDb fallback")
		return EnrichMediaWithOMDb(media)
	}

	// Clean the name for search
	cleanName := cleanMediaName(media.Name)
	util.Debug("Searching TMDB for", "name", cleanName)

	var searchResult *models.TMDBSearchResult
	var err error

	// Search based on media type
	if media.MediaType == models.MediaTypeMovie {
		searchResult, err = client.SearchMovies(cleanName)
	} else {
		searchResult, err = client.SearchTV(cleanName)
	}

	if err != nil {
		util.Debug("TMDB search failed, trying OMDb fallback", "error", err)
		return EnrichMediaWithOMDb(media)
	}

	if len(searchResult.Results) == 0 {
		util.Debug("No TMDB results found, trying OMDb fallback", "name", cleanName)
		return EnrichMediaWithOMDb(media)
	}

	// Use the first result (best match)
	tmdbMedia := searchResult.Results[0]

	// Enrich the media object
	media.TMDBID = tmdbMedia.ID
	media.Rating = tmdbMedia.VoteAverage
	media.Overview = tmdbMedia.Overview

	if tmdbMedia.PosterPath != "" {
		media.ImageURL = client.GetImageURL(tmdbMedia.PosterPath, "w500")
	}

	if media.Year == "" {
		media.Year = tmdbMedia.GetReleaseYear()
	}

	// Get detailed information
	var details *models.TMDBDetails
	if media.MediaType == models.MediaTypeMovie {
		details, err = client.GetMovieDetails(tmdbMedia.ID)
	} else {
		details, err = client.GetTVDetails(tmdbMedia.ID)
	}

	if err == nil && details != nil {
		media.TMDBDetails = details
		media.IMDBID = details.IMDBID
		media.Runtime = details.Runtime

		// Extract genres
		var genres []string
		for _, g := range details.Genres {
			genres = append(genres, g.Name)
		}
		media.Genres = genres
	}

	util.Debug("TMDB enrichment successful",
		"id", media.TMDBID,
		"rating", media.Rating,
		"year", media.Year)

	return nil
}

// cleanMediaName removes tags and cleans the media name for search
func cleanMediaName(name string) string {
	// Remove common tags
	tags := []string{"[Movies/TV]", "[Movie]", "[TV]", "[English]", "[Portuguese]", "[Português]"}
	for _, tag := range tags {
		name = strings.ReplaceAll(name, tag, "")
	}

	// Remove year in parentheses if present
	// e.g., "Movie Name (2024)" -> "Movie Name"
	if idx := strings.LastIndex(name, "("); idx > 0 {
		if endIdx := strings.LastIndex(name, ")"); endIdx > idx {
			possibleYear := strings.TrimSpace(name[idx+1 : endIdx])
			if len(possibleYear) == 4 {
				// Check if it's a year
				isYear := true
				for _, c := range possibleYear {
					if c < '0' || c > '9' {
						isYear = false
						break
					}
				}
				if isYear {
					name = name[:idx]
				}
			}
		}
	}

	return strings.TrimSpace(name)
}

// FormatMediaInfo formats TMDB info for display
func FormatMediaInfo(media *models.Anime) string {
	var parts []string

	if media.Year != "" {
		parts = append(parts, media.Year)
	}

	if media.Rating > 0 {
		parts = append(parts, fmt.Sprintf("★ %.1f", media.Rating))
	}

	if media.Runtime > 0 {
		hours := media.Runtime / 60
		mins := media.Runtime % 60
		if hours > 0 {
			parts = append(parts, fmt.Sprintf("%dh %dm", hours, mins))
		} else {
			parts = append(parts, fmt.Sprintf("%dm", mins))
		}
	}

	if len(media.Genres) > 0 {
		// Show first 3 genres
		maxGenres := 3
		if len(media.Genres) < maxGenres {
			maxGenres = len(media.Genres)
		}
		parts = append(parts, strings.Join(media.Genres[:maxGenres], ", "))
	}

	return strings.Join(parts, " | ")
}
