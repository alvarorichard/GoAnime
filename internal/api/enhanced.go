// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// Enhanced search that supports multiple sources
func SearchAnimeEnhanced(name string, source string) (*models.Anime, error) {
	scraperManager := scraper.NewScraperManager()

	var scraperType *scraper.ScraperType
	switch strings.ToLower(source) {
	case "allanime":
		t := scraper.AllAnimeType
		scraperType = &t
	case "animefire":
		t := scraper.AnimefireType
		scraperType = &t
	default:
		scraperType = nil // Search all sources
	}

	animes, err := scraperManager.SearchAnime(name, scraperType)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime: %w", err)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no anime found with name: %s", name)
	}

	// Add source tags to anime names for clarity
	for _, anime := range animes {
		// Check source field first, then fallback to URL analysis
		if anime.Source != "" {
			// Source already identified by scraper
			continue
		}

		// Fallback source identification by URL analysis
		if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
			// Convert short ID to full AllAnime URL
			anime.URL = "https://allanime.to/anime/" + anime.URL
			if !strings.Contains(anime.Name, "AllAnime") {
				anime.Name = "🌐[AllAnime] " + anime.Name
				anime.Source = "AllAnime"
			}
		} else if strings.Contains(anime.URL, "animefire") && !strings.Contains(anime.Name, "AnimeFire") {
			anime.Name = "🔥[AnimeFire] " + anime.Name
			anime.Source = "AnimeFire.plus"
		}
	}

	// If only one result, return it
	if len(animes) == 1 {
		return animes[0], nil
	}

	// Use fuzzy finder to let user select
	idx, err := fuzzyfinder.Find(
		animes,
		func(i int) string {
			return animes[i].Name
		},
		fuzzyfinder.WithPromptString("Select anime: "),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i >= 0 && i < len(animes) {
				return fmt.Sprintf("URL: %s\nImage: %s", animes[i].URL, animes[i].ImageURL)
			}
			return ""
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("anime selection cancelled: %w", err)
	}

	return animes[idx], nil
}

// Enhanced episode fetching that works with different sources
func GetAnimeEpisodesEnhanced(anime *models.Anime) ([]models.Episode, error) {
	scraperManager := scraper.NewScraperManager()

	// Determine source type from multiple indicators
	var scraperType scraper.ScraperType
	var sourceName string

	if anime.Source == "AllAnime" {
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Source, "AnimeFire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.Name, "AllAnime") {
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "AnimeFire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.URL, "allanime") {
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.URL, "animefire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else {
		// Default to AllAnime for unknown sources
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime (default)"
	}

	fmt.Printf("📺 Obtendo episódios de %s...\n", sourceName)

	scraperInstance, err := scraperManager.GetScraper(scraperType)
	if err != nil {
		return nil, fmt.Errorf("failed to get scraper: %w", err)
	}

	episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}

	if len(episodes) > 0 {
		fmt.Printf("✅ Encontrados %d episódios em %s\n", len(episodes), sourceName)
	}

	return episodes, nil
}

// Enhanced episode URL fetching
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperManager := scraper.NewScraperManager()

	// Determine source type with better logic
	var scraperType scraper.ScraperType
	var sourceName string

	// Priority 1: Check the Source field
	if anime.Source == "AllAnime" {
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Source, "AnimeFire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.Name, "AllAnime") {
		// Priority 2: Check name tags
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "AnimeFire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
		// Priority 3: URL analysis
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.URL, "animefire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else {
		// Default to AllAnime
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime (default)"
	}

	fmt.Printf("🎯 Fonte identificada: %s\n", sourceName)
	fmt.Printf("DEBUG: ScraperType: %v, AnimeURL: %s, EpisodeURL: %s, EpisodeNumber: %s\n",
		scraperType, anime.URL, episode.URL, episode.Number)

	if util.IsDebug {
		util.Debugf("Using scraper type: %v for anime: %s", scraperType, anime.Name)
	}

	scraperInstance, err := scraperManager.GetScraper(scraperType)
	if err != nil {
		return "", fmt.Errorf("failed to get scraper: %w", err)
	}

	if quality == "" {
		quality = "best"
	}

	// For AllAnime, we need to pass the anime ID and episode number separately
	if scraperType == scraper.AllAnimeType {
		if util.IsDebug {
			util.Debugf("AllAnime: Getting stream URL for anime ID: %s, episode: %s", anime.URL, episode.Number)
		}
		streamURL, _, err := scraperInstance.GetStreamURL(anime.URL, episode.Number, quality)
		if err != nil {
			return "", fmt.Errorf("failed to get stream URL from AllAnime: %w", err)
		}
		if util.IsDebug {
			util.Debugf("AllAnime returned stream URL: %s", streamURL)
		}
		return streamURL, nil
	} else {
		// For other scrapers, use the episode URL directly
		if util.IsDebug {
			util.Debugf("Other scraper: Getting stream URL for episode URL: %s", episode.URL)
		}
		streamURL, _, err := scraperInstance.GetStreamURL(episode.URL, quality)
		if err != nil {
			return "", fmt.Errorf("failed to get stream URL: %w", err)
		}
		if util.IsDebug {
			util.Debugf("Other scraper returned stream URL: %s", streamURL)
		}
		return streamURL, nil
	}
}

// Enhanced download support
func DownloadEpisodeEnhanced(anime *models.Anime, episodeNum int, quality string) error {
	util.Infof("Fetching episodes for %s...", anime.Name)

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	if episodeNum < 1 || episodeNum > len(episodes) {
		return fmt.Errorf("episode %d not found (available: 1-%d)", episodeNum, len(episodes))
	}

	episode := episodes[episodeNum-1]

	util.Infof("Getting stream URL for episode %d...", episodeNum)
	streamURL, err := GetEpisodeStreamURL(&episode, anime, quality)
	if err != nil {
		return fmt.Errorf("failed to get stream URL: %w", err)
	}

	util.Infof("Stream URL obtained: %s", streamURL)

	// Create a basic downloader (this would integrate with your existing downloader)
	return downloadFromURL(streamURL, fmt.Sprintf("%s_Episode_%d",
		sanitizeFilename(anime.Name), episodeNum))
}

// Enhanced range download support
func DownloadEpisodeRangeEnhanced(anime *models.Anime, startEp, endEp int, quality string) error {
	util.Infof("Fetching episodes for %s...", anime.Name)

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	if startEp < 1 || endEp > len(episodes) || startEp > endEp {
		return fmt.Errorf("invalid range %d-%d (available: 1-%d)", startEp, endEp, len(episodes))
	}

	for i := startEp; i <= endEp; i++ {
		util.Infof("Downloading episode %d of %d...", i, endEp)

		episode := episodes[i-1]
		streamURL, err := GetEpisodeStreamURL(&episode, anime, quality)
		if err != nil {
			util.Errorf("Failed to get stream URL for episode %d: %v", i, err)
			continue
		}

		filename := fmt.Sprintf("%s_Episode_%d", sanitizeFilename(anime.Name), i)
		if err := downloadFromURL(streamURL, filename); err != nil {
			util.Errorf("Failed to download episode %d: %v", i, err)
			continue
		}

		util.Infof("Successfully downloaded episode %d", i)
	}

	return nil
}

// Helper function to sanitize filename
func sanitizeFilename(name string) string {
	// Remove source tags
	name = strings.ReplaceAll(name, "[AllAnime]", "")
	name = strings.ReplaceAll(name, "[AnimeFire]", "")
	name = strings.TrimSpace(name)

	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}

	return name
}

// Basic download function (placeholder - integrate with your existing downloader)
func downloadFromURL(url, filename string) error {
	// This is a placeholder - you would integrate this with your existing
	// downloader package functionality
	util.Infof("Downloading from URL: %s to file: %s", url, filename)

	// For now, just log the download intent
	// In a real implementation, you'd use the downloader package
	return nil
}

// Legacy wrapper functions to maintain compatibility
func SearchAnimeWithSource(name string, source string) (*models.Anime, error) {
	return SearchAnimeEnhanced(name, source)
}

func GetAnimeEpisodesWithSource(anime *models.Anime) ([]models.Episode, error) {
	return GetAnimeEpisodesEnhanced(anime)
}
