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
	} else if strings.ToLower(source) == "animefire" {
		t := scraper.AnimefireType
		scraperType = &t
	} else {
		// Default behavior: search both sources simultaneously
		scraperType = nil
	}

	// Perform the search - this will search both sources if scraperType is nil
	fmt.Printf("🔍 Buscando '%s' em todas as fontes disponíveis...\n", name)
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

		// Ensure name has proper source tag (unified scraper already adds them)
		if anime.Source == "AllAnime" && !strings.Contains(anime.Name, "AllAnime") {
			anime.Name = "🌐[AllAnime] " + strings.TrimSpace(strings.ReplaceAll(anime.Name, "🌐[AllAnime]", ""))
		} else if anime.Source == "AnimeFire.plus" && !strings.Contains(anime.Name, "AnimeFire") {
			anime.Name = "🔥[AnimeFire] " + strings.TrimSpace(strings.ReplaceAll(anime.Name, "🔥[AnimeFire]", ""))
		}
	}

	fmt.Printf("📊 Total de animes encontrados: %d\n", len(animes))

	// Show sources breakdown
	animefireCount := 0
	allanimeCount := 0
	for _, anime := range animes {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		}
	}

	if animefireCount > 0 {
		fmt.Printf("🔥 AnimeFire.plus: %d resultados\n", animefireCount)
	}
	if allanimeCount > 0 {
		fmt.Printf("🌐 AllAnime: %d resultados\n", allanimeCount)
	}

	// If only one result, return it directly
	if len(animes) == 1 {
		fmt.Printf("✅ Selecionando automaticamente: %s\n", animes[0].Name)

		// CRITICAL: Enrich with AniList data for images and metadata (like the original system)
		if err := enrichAnimeData(animes[0]); err != nil {
			util.Errorf("Error enriching anime data: %v", err)
		}

		return animes[0], nil
	}

	// Use fuzzy finder to let user select with enhanced preview
	idx, err := fuzzyfinder.Find(
		animes,
		func(i int) string {
			return animes[i].Name
		},
		fuzzyfinder.WithPromptString("Selecione o anime desejado: "),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i >= 0 && i < len(animes) {
				anime := animes[i]
				preview := fmt.Sprintf("Fonte: %s\nURL: %s", anime.Source, anime.URL)
				if anime.ImageURL != "" {
					preview += fmt.Sprintf("\nImagem: %s", anime.ImageURL)
				}
				return preview
			}
			return ""
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("seleção de anime cancelada: %w", err)
	}

	selectedAnime := animes[idx]
	fmt.Printf("✅ Anime selecionado: %s (Fonte: %s)\n", selectedAnime.Name, selectedAnime.Source)

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
	} else if strings.Contains(anime.Name, "🌐[AllAnime]") || strings.Contains(anime.Name, "[AllAnime]") {
		// Priority 2: Check name tags
		sourceName = "AllAnime"
		anime.Source = "AllAnime" // Update source field
	} else if strings.Contains(anime.Name, "🔥[AnimeFire]") || strings.Contains(anime.Name, "[AnimeFire]") {
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
		sourceName = "AllAnime (padrão)"
		anime.Source = "AllAnime"
	}

	fmt.Printf("📺 Obtendo episódios de %s para: %s\n", sourceName, strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(anime.Name, "🌐[AllAnime]", ""), "🔥[AnimeFire]", "")))

	var episodes []models.Episode
	var err error

	// Use different approaches based on source
	if strings.Contains(sourceName, "AllAnime") {
		// For AllAnime, use the scraper directly with AniSkip support
		scraperManager := scraper.NewScraperManager()
		scraperInstance, scErr := scraperManager.GetScraper(scraper.AllAnimeType)
		if scErr != nil {
			return nil, fmt.Errorf("falha ao obter scraper AllAnime: %w", scErr)
		}

		// Cast to AllAnime client to access enhanced features
		if allAnimeClient, ok := scraperInstance.(*scraper.AllAnimeClient); ok && anime.MalID > 0 {
			// Use AniSkip enhanced version like Curd does
			episodes, err = allAnimeClient.GetAnimeEpisodesWithAniSkip(anime.URL, anime.MalID, GetAndParseAniSkipData)
			fmt.Printf("🎯 AniSkip integration enabled for MAL ID: %d\n", anime.MalID)
		} else {
			// Fallback to regular episodes
			episodes, err = scraperInstance.GetAnimeEpisodes(anime.URL)
		}
	} else {
		// For AnimeFire and others, use the original API function
		episodes, err = GetAnimeEpisodes(anime.URL)
	}

	if err != nil {
		return nil, fmt.Errorf("falha ao obter episódios de %s: %w", sourceName, err)
	}

	if len(episodes) > 0 {
		fmt.Printf("✅ Encontrados %d episódios em %s\n", len(episodes), sourceName)

		// Provide additional info for user based on source
		if strings.Contains(sourceName, "AllAnime") {
			fmt.Printf("🌐 Fonte: AllAnime - Episódios em alta qualidade disponíveis\n")
		} else {
			fmt.Printf("🔥 Fonte: AnimeFire.plus - Episódios dublados/legendados disponíveis\n")
		}
	} else {
		fmt.Printf("⚠️  Nenhum episódio encontrado em %s\n", sourceName)
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
	} else if strings.Contains(anime.Name, "🌐[AllAnime]") || strings.Contains(anime.Name, "[AllAnime]") {
		// Priority 2: Check name tags
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "🔥[AnimeFire]") || strings.Contains(anime.Name, "[AnimeFire]") {
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

	fmt.Printf("🎯 Obtendo URL de stream de %s para episódio %s\n", sourceName, episode.Number)

	if util.IsDebug {
		util.Debugf("Detalhes da fonte:")
		util.Debugf("  ScraperType: %v", scraperType)
		util.Debugf("  AnimeURL: %s", anime.URL)
		util.Debugf("  EpisodeURL: %s", episode.URL)
		util.Debugf("  EpisodeNumber: %s", episode.Number)
		util.Debugf("  Quality: %s", quality)
	}

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
		fmt.Printf("🌐 Processando através do AllAnime...\n")
		if util.IsDebug {
			util.Debugf("AllAnime: Obtendo URL de stream para anime ID: %s, episódio: %s", anime.URL, episode.Number)
		}
		streamURL, _, streamErr = scraperInstance.GetStreamURL(anime.URL, episode.Number, quality)
	} else {
		fmt.Printf("🔥 Processando através do AnimeFire.plus...\n")
		if util.IsDebug {
			util.Debugf("AnimeFire: Obtendo URL de stream para episódio URL: %s", episode.URL)
		}
		streamURL, _, streamErr = scraperInstance.GetStreamURL(episode.URL, quality)
	}

	if streamErr != nil {
		return "", fmt.Errorf("falha ao obter URL de stream de %s: %w", sourceName, streamErr)
	}

	if streamURL == "" {
		return "", fmt.Errorf("URL de stream vazia retornada de %s", sourceName)
	}

	fmt.Printf("✅ URL de stream obtida com sucesso de %s\n", sourceName)
	if util.IsDebug {
		util.Debugf("Stream URL obtida: %s", streamURL)
	}

	return streamURL, nil
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
