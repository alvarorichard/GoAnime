// Package api provides enhanced episode URL fetching with AllAnime navigation support
package api

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// GetEpisodeStreamURLEnhanced gets streaming URL with AllAnime navigation support.
// Uses the provider registry for source detection, then applies AllAnime-specific
// direct resolution when applicable.
func GetEpisodeStreamURLEnhanced(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	util.Debug("Enhanced episode URL fetch",
		"source", providers.ResolveSourceName(anime),
		"episode", episode.Number,
		"quality", quality)

	if providers.IsAllAnime(anime) {
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

	return GetEpisodeStreamURL(episode, anime, quality)
}

// GetAllAnimeEpisodeURLDirect gets streaming URL directly without circular dependencies
func GetAllAnimeEpisodeURLDirect(anime *models.Anime, episodeNumber string, quality string) (string, map[string]string, error) {
	if !isAllAnimeSourceAPI(anime) {
		return "", nil, fmt.Errorf("this function is only for AllAnime sources")
	}

	// Use the cached scraper manager to get the AllAnime client (avoids re-creating each time)
	sm := scraper.NewScraperManager()
	scraperInstance, scErr := sm.GetScraper(scraper.AllAnimeType)
	var client *scraper.AllAnimeClient
	if scErr == nil {
		if adapter, ok := scraperInstance.(interface {
			Client() *scraper.AllAnimeClient
		}); ok {
			client = adapter.Client()
		}
	}
	if client == nil {
		// Fallback: create directly (shouldn't happen with singleton manager)
		client = scraper.NewAllAnimeClient()
	}
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
		parts := strings.SplitSeq(url, "/")
		for part := range parts {
			if len(part) > 5 && len(part) < 30 &&
				strings.ContainsAny(part, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") {
				return part
			}
		}
	}

	return url // Return as-is if can't extract
}
