// Package api provides enhanced episode URL fetching with AllAnime navigation support
package api

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// GetEpisodeStreamURLEnhanced gets streaming URL with AllAnime navigation support
func GetEpisodeStreamURLEnhanced(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	// Determine source type and use appropriate method
	sourceName := "Unknown"
	scraperType := scraper.AllAnimeType // Default

	// Enhanced source detection like in enhanced.go
	if anime.Source != "" {
		sourceName = anime.Source
		if strings.Contains(anime.Source, "AllAnime") {
			scraperType = scraper.AllAnimeType
		} else if strings.Contains(anime.Source, "AnimeFire") {
			scraperType = scraper.AnimefireType
		}
	} else if strings.Contains(anime.Name, "[AllAnime]") {
		// Priority 1: Name tag detection
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "[AnimeFire]") {
		// Priority 2: AnimeFire tag detection
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
		// Priority 3: URL analysis for AllAnime (short IDs)
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.URL, "animefire") {
		// Priority 4: URL analysis for AnimeFire
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.URL, "allanime") {
		// Priority 5: AllAnime full URLs
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	}

	util.Debug("Enhanced episode URL fetch",
		"source", sourceName,
		"episode", episode.Number,
		"quality", quality)

	// Use AllAnime enhanced navigation if applicable
	if scraperType == scraper.AllAnimeType {
		url, metadata, err := GetAllAnimeEpisodeURLDirect(anime, episode.Number, quality)
		if err != nil {
			return "", fmt.Errorf("failed to get AllAnime episode URL: %w", err)
		}

		util.Debug("AllAnime episode URL retrieved via direct method",
			"episode", episode.Number,
			"quality", metadata["quality"],
			"priority", metadata["priority"])

		return url, nil
	}

	// Fallback to regular enhanced API
	return GetEpisodeStreamURL(episode, anime, quality)
}

// GetAllAnimeEpisodeURLDirect gets streaming URL directly without circular dependencies
func GetAllAnimeEpisodeURLDirect(anime *models.Anime, episodeNumber string, quality string) (string, map[string]string, error) {
	if !isAllAnimeSourceAPI(anime) {
		return "", nil, fmt.Errorf("this function is only for AllAnime sources")
	}

	// Create AllAnime client directly
	client := scraper.NewAllAnimeClient()
	animeID := extractAllAnimeIDAPI(anime.URL)

	if animeID == "" {
		return "", nil, fmt.Errorf("could not extract anime ID from URL: %s", anime.URL)
	}

	mode := "sub" // Default
	if quality == "" {
		quality = "best"
	}

	url, metadata, err := client.GetEpisodeURL(animeID, episodeNumber, mode, quality)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get episode URL: %w", err)
	}

	// Add additional metadata
	metadata["navigator"] = "allanime"
	metadata["anime_id"] = animeID
	metadata["mode"] = mode

	util.Debug("AllAnime episode URL retrieved directly",
		"episode", episodeNumber,
		"quality", quality,
		"url_length", len(url))

	return url, metadata, nil
}

// Helper function to check if anime is from AllAnime source (API module)
func isAllAnimeSourceAPI(anime *models.Anime) bool {
	if anime.Source == "AllAnime" {
		return true
	}

	if strings.Contains(anime.URL, "allanime") {
		return true
	}

	if len(anime.URL) < 30 &&
		strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") &&
		!strings.Contains(anime.URL, "http") {
		return true
	}

	return false
}

// Helper function to extract AllAnime ID from URL (API module)
func extractAllAnimeIDAPI(url string) string {
	// For AllAnime, the URL is often just the anime ID
	if !strings.Contains(url, "http") && len(url) < 30 {
		return url
	}

	// Extract ID from full AllAnime URLs if needed
	if strings.Contains(url, "allanime") {
		parts := strings.Split(url, "/")
		for _, part := range parts {
			if len(part) > 5 && len(part) < 30 &&
				strings.ContainsAny(part, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") {
				return part
			}
		}
	}

	return url // Return as-is if can't extract
}
