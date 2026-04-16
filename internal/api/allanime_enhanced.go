// Package api provides enhanced episode URL fetching with AllAnime navigation support
package api

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// GetEpisodeStreamURLEnhanced gets streaming URL with AllAnime navigation support
func GetEpisodeStreamURLEnhanced(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	resolved, err := ResolveSource(anime)
	if err != nil {
		return "", err
	}

	util.Debug("Enhanced episode URL fetch",
		"source", resolved.Name,
		"episode", episode.Number,
		"quality", quality)

	if resolved.Kind != SourceAllAnime {
		return GetEpisodeStreamURL(episode, anime, quality)
	}

	url, metadata, err := GetAllAnimeEpisodeURLDirect(anime, providerEpisodeNumber(episode), quality)
	if err != nil {
		return "", fmt.Errorf("failed to get AllAnime episode URL: %w", err)
	}

	util.Debug("AllAnime episode URL retrieved via direct method",
		"episode", providerEpisodeNumber(episode),
		"quality", metadata["quality"],
		"priority", metadata["priority"])

	return url, nil
}

// GetAllAnimeEpisodeURLDirect gets streaming URL directly without circular dependencies
func GetAllAnimeEpisodeURLDirect(anime *models.Anime, episodeNumber, quality string) (string, map[string]string, error) {
	if !IsAllAnimeSource(anime) {
		return "", nil, fmt.Errorf("this function is only for AllAnime sources")
	}

	// Use the cached scraper manager to get the AllAnime client (avoids re-creating each time)
	scraperInstance, scErr := getScraperForKind(SourceAllAnime)
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
	animeID := ExtractAllAnimeID(anime.URL)

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
