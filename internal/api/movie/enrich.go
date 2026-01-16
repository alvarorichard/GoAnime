// Package movie provides media enrichment functions for movies and TV shows
package movie

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// EnrichMedia enriches a media item with metadata from TMDB or OMDb
// Falls back to OMDb if TMDB API key is not configured
func EnrichMedia(media *models.Media) error {
	if media.MediaType != models.MediaTypeMovie && media.MediaType != models.MediaTypeTV {
		return nil // Only enrich movies and TV shows
	}

	tmdbClient := NewTMDBClient()

	// If TMDB is not configured, fall back to OMDb
	if !tmdbClient.IsConfigured() {
		util.Debug("TMDB not configured, using OMDb fallback")
		return EnrichWithOMDb(media)
	}

	// Clean the name for search
	cleanName := CleanMediaName(media.Name)
	util.Debug("Searching TMDB for", "name", cleanName)

	var searchResult *models.TMDBSearchResult
	var err error

	// Search based on media type
	if media.MediaType == models.MediaTypeMovie {
		searchResult, err = tmdbClient.SearchMovies(cleanName)
	} else {
		searchResult, err = tmdbClient.SearchTV(cleanName)
	}

	if err != nil {
		util.Debug("TMDB search failed, trying OMDb fallback", "error", err)
		return EnrichWithOMDb(media)
	}

	if len(searchResult.Results) == 0 {
		util.Debug("No TMDB results found, trying OMDb fallback", "name", cleanName)
		return EnrichWithOMDb(media)
	}

	// Use the first result (best match)
	tmdbMedia := searchResult.Results[0]

	// Enrich the media object
	media.TMDBID = tmdbMedia.ID
	media.Rating = tmdbMedia.VoteAverage
	media.Overview = tmdbMedia.Overview

	if tmdbMedia.PosterPath != "" {
		media.ImageURL = tmdbClient.GetImageURL(tmdbMedia.PosterPath, "w500")
	}

	if media.Year == "" {
		media.Year = tmdbMedia.GetReleaseYear()
	}

	// Get detailed information
	var details *models.TMDBDetails
	if media.MediaType == models.MediaTypeMovie {
		details, err = tmdbClient.GetMovieDetails(tmdbMedia.ID)
	} else {
		details, err = tmdbClient.GetTVDetails(tmdbMedia.ID)
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

// EnrichWithOMDb enriches a media item with OMDb data
func EnrichWithOMDb(media *models.Media) error {
	if media.MediaType != models.MediaTypeMovie && media.MediaType != models.MediaTypeTV {
		return nil // Only enrich movies and TV shows
	}

	client := NewOMDbClient()
	if !client.IsConfigured() {
		util.Debug("OMDb not configured, skipping enrichment")
		return nil
	}

	// Clean the name for search
	cleanName := CleanMediaName(media.Name)
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
			// Not finding metadata is not critical - just log and continue
			util.Debug("OMDb search failed or no results, continuing without enrichment", "error", searchErr)
			return nil // Return nil instead of error - metadata is optional
		}
		// Get details for the first result
		omdbMedia, err = client.GetByIMDBID(searchResult.Search[0].IMDBID)
		if err != nil {
			util.Debug("Failed to get OMDb details, continuing without enrichment", "error", err)
			return nil // Return nil instead of error - metadata is optional
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

// CleanMediaName removes tags and cleans the media name for search
func CleanMediaName(name string) string {
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

// FormatMediaInfo formats media info for display
func FormatMediaInfo(media *models.Media) string {
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
