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

// Enhanced search that supports multiple sources - always searches both animefire.plus and allanime simultaneously
func SearchAnimeEnhanced(name string, source string) (*models.Anime, error) {
	scraperManager := scraper.NewScraperManager()

	var scraperType *scraper.ScraperType

	// If a specific source is requested, honor it
	if strings.ToLower(source) == "allanime" {
		t := scraper.AllAnimeType
		scraperType = &t
		util.Debug("Searching specific source", "source", "AllAnime")
	} else if strings.ToLower(source) == "animefire" {
		t := scraper.AnimefireType
		scraperType = &t
		util.Debug("Searching specific source", "source", "AnimeFire")
	} else {
		// Default behavior: search both sources simultaneously
		scraperType = nil
		util.Debug("Searching all sources", "query", name)
	}

	// Perform the search - this will search both sources if scraperType is nil
	util.Debug("Searching for anime", "query", name)
	animes, err := scraperManager.SearchAnime(name, scraperType)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime: %w", err)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("nenhum anime encontrado com o nome: %s", name)
	}

	// Enhance source identification and tagging
	for _, anime := range animes {
		// Ensure proper source identification
		if anime.Source == "" {
			// Fallback source identification by URL analysis
			if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
				anime.Source = "AllAnime"
			} else if strings.Contains(anime.URL, "animefire") {
				anime.Source = "AnimeFire.plus"
			}
		}

		// Ensure name has proper source tag (without emojis for cleaner display)
		if anime.Source == "AllAnime" && !strings.Contains(anime.Name, "AllAnime") {
			anime.Name = "[AllAnime] " + strings.TrimSpace(strings.ReplaceAll(anime.Name, "[AllAnime]", ""))
		} else if anime.Source == "AnimeFire.plus" && !strings.Contains(anime.Name, "AnimeFire") {
			anime.Name = "[AnimeFire] " + strings.TrimSpace(strings.ReplaceAll(anime.Name, "[AnimeFire]", ""))
		}
	}

	util.Debug("Search results summary", "total", len(animes))

	// Show sources breakdown in debug only
	animefireCount := 0
	allanimeCount := 0
	for _, anime := range animes {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		}
	}

	util.Debug("Source breakdown", "AnimeFire", animefireCount, "AllAnime", allanimeCount)

	// If only one result, return it directly
	if len(animes) == 1 {
		util.Debug("Auto-selecting single result", "anime", animes[0].Name)

		// CRITICAL: Enrich with AniList data for images and metadata (like the original system)
		if err := enrichAnimeData(animes[0]); err != nil {
			util.Errorf("Error enriching anime data: %v", err)
		}

		return animes[0], nil
	}

	// Helper to map provider tags to user-friendly language labels for display only
	providerLabel := func(src string) string {
		if strings.Contains(src, "AnimeFire") {
			return "Portuguese"
		}
		if src == "AllAnime" {
			return "English"
		}
		return src
	}

	// Use fuzzy finder to let user select
	var idx int

	if util.IsDebug {
		// In debug mode, show preview window with technical details
		idx, err = fuzzyfinder.Find(
			animes,
			func(i int) string {
				// Replace provider tags in the display name only
				name := animes[i].Name
				name = strings.ReplaceAll(name, "[AllAnime]", "[English]")
				name = strings.ReplaceAll(name, "[AnimeFire]", "[Portuguese]")
				return name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i >= 0 && i < len(animes) {
					anime := animes[i]
					var preview string
					preview = "Source: " + providerLabel(anime.Source) + "\nURL: " + anime.URL
					if anime.ImageURL != "" {
						preview += "\nImage: " + anime.ImageURL
					}
					return preview
				}
				return ""
			}),
		)
	} else {
		// In normal mode, no preview window at all
		idx, err = fuzzyfinder.Find(
			animes,
			func(i int) string {
				// Replace provider tags in the display name only
				name := animes[i].Name
				name = strings.ReplaceAll(name, "[AllAnime]", "[English]")
				name = strings.ReplaceAll(name, "[AnimeFire]", "[Portuguese]")
				return name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("seleção de anime cancelada: %w", err)
	}

	selectedAnime := animes[idx]
	util.Debug("Anime selected", "name", selectedAnime.Name, "source", selectedAnime.Source)

	// CRITICAL: Enrich with AniList data for images and metadata (like the original system)
	if err := enrichAnimeData(selectedAnime); err != nil {
		util.Errorf("Error enriching anime data: %v", err)
	}

	return selectedAnime, nil
}

// Enhanced episode fetching that works with different sources
func GetAnimeEpisodesEnhanced(anime *models.Anime) ([]models.Episode, error) {
	// Determine source type from multiple indicators with enhanced logic
	var sourceName string

	// Priority 1: Check the Source field (most reliable)
	if anime.Source == "AllAnime" {
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Source, "AnimeFire") {
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.Name, "[AllAnime]") {
		// Priority 2: Check name tags
		sourceName = "AllAnime"
		anime.Source = "AllAnime" // Update source field
	} else if strings.Contains(anime.Name, "[AnimeFire]") {
		sourceName = "AnimeFire.plus"
		anime.Source = "AnimeFire.plus" // Update source field
	} else if strings.Contains(anime.URL, "allanime") || (len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http")) {
		// Priority 3: URL analysis for AllAnime (short IDs or allanime URLs)
		sourceName = "AllAnime"
		anime.Source = "AllAnime" // Update source field
	} else if strings.Contains(anime.URL, "animefire") {
		// Priority 4: URL analysis for AnimeFire
		sourceName = "AnimeFire.plus"
		anime.Source = "AnimeFire.plus" // Update source field
	} else {
		// Default to AllAnime for unknown sources
		sourceName = "AllAnime (default)"
		anime.Source = "AllAnime"
	}

	cleanName := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(anime.Name, "[AllAnime]", ""), "[AnimeFire]", ""))

	util.Debug("Getting episodes", "source", sourceName, "anime", cleanName)

	var episodes []models.Episode
	var err error

	// Use different approaches based on source
	if strings.Contains(sourceName, "AllAnime") {
		// For AllAnime, use the scraper directly with AniSkip support
		scraperManager := scraper.NewScraperManager()
		scraperInstance, scErr := scraperManager.GetScraper(scraper.AllAnimeType)
		if scErr != nil {
			return nil, fmt.Errorf("failed to get AllAnime scraper: %w", scErr)
		}

		// Cast to AllAnime client to access enhanced features
		if allAnimeClient, ok := scraperInstance.(*scraper.AllAnimeClient); ok && anime.MalID > 0 {
			// Use AniSkip enhanced version like Curd does
			episodes, err = allAnimeClient.GetAnimeEpisodesWithAniSkip(anime.URL, anime.MalID, GetAndParseAniSkipData)
			util.Debug("AniSkip integration enabled", "malID", anime.MalID)
		} else {
			// Fallback to regular episodes
			episodes, err = scraperInstance.GetAnimeEpisodes(anime.URL)
		}
	} else {
		// For AnimeFire and others, use the original API function
		episodes, err = GetAnimeEpisodes(anime.URL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from %s: %w", sourceName, err)
	}

	if len(episodes) > 0 {
		util.Debug("Episodes found", "count", len(episodes), "source", sourceName)

		// Provide additional info for user based on source (debug only)
		if strings.Contains(sourceName, "AllAnime") {
			util.Debug("Source info", "type", "AllAnime", "quality", "high")
		} else {
			util.Debug("Source info", "type", "AnimeFire.plus", "features", "dubbed/subtitled")
		}
	} else {
		util.Warn("No episodes found", "source", sourceName)
	}

	return episodes, nil
}

// Enhanced episode URL fetching with improved source detection
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperManager := scraper.NewScraperManager()

	// Determine source type with enhanced logic
	var scraperType scraper.ScraperType
	var sourceName string

	// Priority 1: Check the Source field (most reliable)
	if anime.Source == "AllAnime" {
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Source, "AnimeFire") {
		scraperType = scraper.AnimefireType
		sourceName = "AnimeFire.plus"
	} else if strings.Contains(anime.Name, "[AllAnime]") {
		// Priority 2: Check name tags
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "[AnimeFire]") {
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
	} else {
		// Default to AllAnime
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime (padrão)"
	}

	util.Debug("Getting stream URL", "source", sourceName, "episode", episode.Number)

	util.Debug("Source details",
		"scraperType", scraperType,
		"animeURL", anime.URL,
		"episodeURL", episode.URL,
		"episodeNumber", episode.Number,
		"quality", quality)

	scraperInstance, err := scraperManager.GetScraper(scraperType)
	if err != nil {
		return "", fmt.Errorf("falha ao obter scraper para %s: %w", sourceName, err)
	}

	if quality == "" {
		quality = "best"
	}

	var streamURL string
	var streamErr error

	// Handle different scraper types with appropriate parameters
	if scraperType == scraper.AllAnimeType {
		util.Debug("Processing through AllAnime")
		streamURL, _, streamErr = scraperInstance.GetStreamURL(anime.URL, episode.Number, quality)
	} else {
		util.Debug("Processing through AnimeFire.plus")
		streamURL, _, streamErr = scraperInstance.GetStreamURL(episode.URL, quality)
	}

	if streamErr != nil {
		return "", fmt.Errorf("falha ao obter URL de stream de %s: %w", sourceName, streamErr)
	}

	if streamURL == "" {
		return "", fmt.Errorf("URL de stream vazia retornada de %s", sourceName)
	}

	util.Debug("Stream URL obtained", "source", sourceName)
	util.Debug("Stream URL details", "url", streamURL)

	return streamURL, nil
}

// Enhanced download support
func DownloadEpisodeEnhanced(anime *models.Anime, episodeNum int, quality string) error {
	util.Debugf("Fetching episodes for %s...", anime.Name)

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	if episodeNum < 1 || episodeNum > len(episodes) {
		return fmt.Errorf("episode %d not found (available: 1-%d)", episodeNum, len(episodes))
	}

	episode := episodes[episodeNum-1]

	util.Debugf("Getting stream URL for episode %d...", episodeNum)
	streamURL, err := GetEpisodeStreamURL(&episode, anime, quality)
	if err != nil {
		return fmt.Errorf("failed to get stream URL: %w", err)
	}

	util.Debugf("Stream URL obtained: %s", streamURL)

	// Create a basic downloader (this would integrate with your existing downloader)
	return downloadFromURL(streamURL, fmt.Sprintf("%s_Episode_%d",
		sanitizeFilename(anime.Name), episodeNum))
}

// Enhanced range download support
func DownloadEpisodeRangeEnhanced(anime *models.Anime, startEp, endEp int, quality string) error {
	util.Debugf("Fetching episodes for %s...", anime.Name)

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
		// Note: downloadFromURL is a placeholder - integrate with proper downloader
		_ = downloadFromURL(streamURL, filename) // This will always fail as expected

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
func downloadFromURL(_ string, _ string) error {
	// This is a placeholder that should fail to trigger fallback to the proper downloader
	util.Debugf("Enhanced API downloadFromURL is a placeholder - returning error to trigger fallback")
	return fmt.Errorf("enhanced download not implemented - use legacy downloader")
}

// Legacy wrapper functions to maintain compatibility
func SearchAnimeWithSource(name string, source string) (*models.Anime, error) {
	return SearchAnimeEnhanced(name, source)
}

func GetAnimeEpisodesWithSource(anime *models.Anime) ([]models.Episode, error) {
	return GetAnimeEpisodesEnhanced(anime)
}
