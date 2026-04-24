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
	// TMDBBaseURL is the TMDB REST API base URL.
	TMDBBaseURL = "https://api.themoviedb.org/3"
	// TMDBImageBaseURL is the TMDB image CDN base URL.
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
			Timeout:   15 * time.Second,
			Transport: safeMovieTransport(15 * time.Second),
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

	return io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
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
