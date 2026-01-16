// Package api provides OMDb API integration for movie/TV metadata
// OMDb API is free with limited usage or with API key for full access
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// OMDb API base URL
	OMDbBaseURL = "https://www.omdbapi.com"
)

// OMDbSearchResult represents a search result from OMDb
type OMDbSearchResult struct {
	Search       []OMDbMedia `json:"Search"`
	TotalResults string      `json:"totalResults"`
	Response     string      `json:"Response"`
	Error        string      `json:"Error"`
}

// OMDbMedia represents a movie or series from OMDb
type OMDbMedia struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	IMDBID string `json:"imdbID"`
	Type   string `json:"Type"` // "movie", "series", "episode"
	Poster string `json:"Poster"`
	// Detailed fields (only in full response)
	Rated        string `json:"Rated"`
	Released     string `json:"Released"`
	Runtime      string `json:"Runtime"`
	Genre        string `json:"Genre"`
	Director     string `json:"Director"`
	Writer       string `json:"Writer"`
	Actors       string `json:"Actors"`
	Plot         string `json:"Plot"`
	Language     string `json:"Language"`
	Country      string `json:"Country"`
	Awards       string `json:"Awards"`
	Metascore    string `json:"Metascore"`
	IMDBRating   string `json:"imdbRating"`
	IMDBVotes    string `json:"imdbVotes"`
	BoxOffice    string `json:"BoxOffice"`
	Production   string `json:"Production"`
	Website      string `json:"Website"`
	TotalSeasons string `json:"totalSeasons"` // For series
	Response     string `json:"Response"`
	Error        string `json:"Error"`
}

// OMDbClient handles interactions with OMDb API
type OMDbClient struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewOMDbClient creates a new OMDb client
// If OMDB_API_KEY is set, uses authenticated requests
// Otherwise uses limited public access
func NewOMDbClient() *OMDbClient {
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		// Use a demo key - limited to 1000 requests per day
		// Users can get their own key at https://www.omdbapi.com/apikey.aspx
		apiKey = "trilogy" // Public demo key
	}

	return &OMDbClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: OMDbBaseURL,
	}
}

// IsConfigured returns true if OMDb client is ready
func (c *OMDbClient) IsConfigured() bool {
	return c.apiKey != ""
}

// SearchByTitle searches for movies/series by title
func (c *OMDbClient) SearchByTitle(title string, mediaType string) (*OMDbSearchResult, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("s", title)

	if mediaType != "" {
		// "movie", "series", or "episode"
		params.Set("type", mediaType)
	}

	endpoint := fmt.Sprintf("%s/?%s", c.baseURL, params.Encode())

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("OMDb search failed: %w", err)
	}

	var result OMDbSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse OMDb response: %w", err)
	}

	if result.Response == "False" {
		return nil, fmt.Errorf("OMDb error: %s", result.Error)
	}

	return &result, nil
}

// GetByIMDBID gets detailed information by IMDB ID
func (c *OMDbClient) GetByIMDBID(imdbID string) (*OMDbMedia, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("i", imdbID)
	params.Set("plot", "short")

	endpoint := fmt.Sprintf("%s/?%s", c.baseURL, params.Encode())

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("OMDb get failed: %w", err)
	}

	var media OMDbMedia
	if err := json.Unmarshal(body, &media); err != nil {
		return nil, fmt.Errorf("failed to parse OMDb response: %w", err)
	}

	if media.Response == "False" {
		return nil, fmt.Errorf("OMDb error: %s", media.Error)
	}

	return &media, nil
}

// GetByTitle gets detailed information by exact title
func (c *OMDbClient) GetByTitle(title string, year string) (*OMDbMedia, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("t", title)
	params.Set("plot", "short")

	if year != "" {
		params.Set("y", year)
	}

	endpoint := fmt.Sprintf("%s/?%s", c.baseURL, params.Encode())

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("OMDb get failed: %w", err)
	}

	var media OMDbMedia
	if err := json.Unmarshal(body, &media); err != nil {
		return nil, fmt.Errorf("failed to parse OMDb response: %w", err)
	}

	if media.Response == "False" {
		return nil, fmt.Errorf("OMDb error: %s", media.Error)
	}

	return &media, nil
}

// makeRequest performs an HTTP request to OMDb API
func (c *OMDbClient) makeRequest(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
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
		return nil, fmt.Errorf("OMDb API returned status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// GetRuntimeMinutes parses runtime string to minutes
func (m *OMDbMedia) GetRuntimeMinutes() int {
	// Runtime is usually "136 min"
	runtime := strings.TrimSuffix(m.Runtime, " min")
	runtime = strings.TrimSpace(runtime)
	minutes, _ := strconv.Atoi(runtime)
	return minutes
}

// GetRating parses IMDB rating to float
func (m *OMDbMedia) GetRating() float64 {
	rating, _ := strconv.ParseFloat(m.IMDBRating, 64)
	return rating
}

// GetGenres splits genre string into list
func (m *OMDbMedia) GetGenres() []string {
	if m.Genre == "" || m.Genre == "N/A" {
		return nil
	}
	genres := strings.Split(m.Genre, ", ")
	return genres
}

// EnrichMediaWithOMDb enriches a media item with OMDb data
func EnrichMediaWithOMDb(media *models.Anime) error {
	if media.MediaType != models.MediaTypeMovie && media.MediaType != models.MediaTypeTV {
		return nil // Only enrich movies and TV shows
	}

	client := NewOMDbClient()
	if !client.IsConfigured() {
		util.Debug("OMDb not configured, skipping enrichment")
		return nil
	}

	// Clean the name for search
	cleanName := cleanMediaName(media.Name)
	util.Debug("Searching OMDb for", "name", cleanName)

	// Determine media type for search
	var searchType string
	switch media.MediaType {
	case models.MediaTypeMovie:
		searchType = "movie"
	case models.MediaTypeTV:
		searchType = "series"
	}

	// Try exact title match first
	omdbMedia, err := client.GetByTitle(cleanName, media.Year)
	if err != nil {
		// Fall back to search
		util.Debug("Exact title match failed, trying search", "error", err)
		searchResult, searchErr := client.SearchByTitle(cleanName, searchType)
		if searchErr != nil || len(searchResult.Search) == 0 {
			util.Debug("OMDb search failed", "error", searchErr)
			return searchErr
		}
		// Get details for the first result
		omdbMedia, err = client.GetByIMDBID(searchResult.Search[0].IMDBID)
		if err != nil {
			return err
		}
	}

	// Enrich the media object
	if omdbMedia.IMDBID != "" {
		media.IMDBID = omdbMedia.IMDBID
	}

	if rating := omdbMedia.GetRating(); rating > 0 {
		media.Rating = rating
	}

	if omdbMedia.Plot != "" && omdbMedia.Plot != "N/A" {
		media.Overview = omdbMedia.Plot
	}

	if media.Year == "" && omdbMedia.Year != "" && omdbMedia.Year != "N/A" {
		media.Year = omdbMedia.Year
	}

	if runtime := omdbMedia.GetRuntimeMinutes(); runtime > 0 {
		media.Runtime = runtime
	}

	if genres := omdbMedia.GetGenres(); len(genres) > 0 {
		media.Genres = genres
	}

	if omdbMedia.Poster != "" && omdbMedia.Poster != "N/A" && media.ImageURL == "" {
		media.ImageURL = omdbMedia.Poster
	}

	util.Debug("OMDb enrichment successful",
		"imdb", media.IMDBID,
		"rating", media.Rating,
		"year", media.Year)

	return nil
}
