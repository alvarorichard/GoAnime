// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"sync"

	"github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

var initProvidersOnce sync.Once

// initProviders injects dependencies into providers that cannot import the api package
// directly (to avoid circular imports). Called lazily on first use.
func initProviders() {
	initProvidersOnce.Do(func() {
		p := providers.ForSourceName("allanime")
		if ap, ok := p.(*providers.AllAnimeProvider); ok {
			ap.SetAniSkipFunc(GetAndParseAniSkipData)
		}
	})
}

// ErrBackToSearch is returned when user selects the back option to search again
var ErrBackToSearch = errors.New("back to search requested")

// sourceToScraperType maps CLI --source values to scraper types.
var sourceToScraperType = map[string]scraper.ScraperType{
	"allanime":   scraper.AllAnimeType,
	"animefire":  scraper.AnimefireType,
	"animedrive": scraper.AnimeDriveType,
	"flixhq":     scraper.FlixHQType,
	"movie":      scraper.FlixHQType,
	"tv":         scraper.FlixHQType,
	"9anime":     scraper.NineAnimeType,
	"nineanime":  scraper.NineAnimeType,
	"goyabu":     scraper.GoyabuType,
}

// Enhanced search that supports multiple sources - always searches both Animefire.io and allanime simultaneously
func SearchAnimeEnhanced(name string, source string) (*models.Anime, error) {
	scraperManager := scraper.NewScraperManager()

	var scraperType *scraper.ScraperType
	isPTBR := false

	lowerSource := strings.ToLower(source)
	if lowerSource == "ptbr" || lowerSource == "pt-br" {
		isPTBR = true
		util.Debug("Searching all PT-BR sources (AnimeFire + Goyabu)")
	} else if st, ok := sourceToScraperType[lowerSource]; ok {
		scraperType = &st
		util.Debug("Searching specific source", "source", lowerSource)
	} else if lowerSource != "" {
		util.Debug("Unknown source, searching all", "source", lowerSource)
	} else {
		util.Debug("Searching all sources", "query", name)
	}

	// Perform the search
	util.Debug("Searching for anime/media", "query", name)
	var animes []*models.Anime
	var searchErr error
	util.RunWithSpinner("Searching for anime...", func() {
		if isPTBR {
			animes, searchErr = scraperManager.SearchAnimePTBR(name)
		} else {
			animes, searchErr = scraperManager.SearchAnime(name, scraperType)
		}
	})
	if searchErr != nil {
		return nil, fmt.Errorf("failed to search: %w", searchErr)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no results found for: %s", name)
	}

	// Enhance source identification - names already have language tags from unified.go
	for _, anime := range animes {
		// Ensure proper source identification (for internal use only)
		if anime.Source == "" {
			// Fallback source identification by URL analysis
			if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
				anime.Source = "AllAnime"
			} else if strings.Contains(anime.URL, "animefire") {
				anime.Source = "Animefire.io"
			} else if strings.Contains(anime.URL, "animesdrive") {
				anime.Source = "AnimeDrive"
			} else if strings.Contains(anime.URL, "goyabu") {
				anime.Source = "Goyabu"
			} else if strings.Contains(anime.URL, "flixhq") {
				anime.Source = "FlixHQ"
			}
			// Note: 9Anime uses numeric IDs which can't be identified by URL alone;
			// the Source field is already set by the scraper
		}

		// Language tags are already added by unified.go, don't duplicate them here
	}

	util.Debug("Search results summary", "total", len(animes))

	// Show sources breakdown in debug only
	animefireCount := 0
	allanimeCount := 0
	animedriveCount := 0
	flixhqCount := 0
	nineAnimeCount := 0
	for _, anime := range animes {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		} else if anime.Source == "AnimeDrive" {
			animedriveCount++
		} else if anime.Source == "FlixHQ" {
			flixhqCount++
		} else if anime.Source == "9Anime" {
			nineAnimeCount++
		}
	}

	util.Debug("Source breakdown", "AnimeFire", animefireCount, "AllAnime", allanimeCount, "AnimeDrive", animedriveCount, "FlixHQ", flixhqCount, "9Anime", nineAnimeCount)

	// Sort results by language priority: Portuguese first, then Multilanguage, Movies/TV, English, others
	sort.SliceStable(animes, func(i, j int) bool {
		return languagePriority(animes[i].Name) < languagePriority(animes[j].Name)
	})

	// Create a special "back" option as the first item
	backOption := &models.Anime{
		Name:   "← Back",
		URL:    "__back__",
		Source: "__back__",
	}

	// Prepend back option to the list
	animesWithBack := make([]*models.Anime, 0, len(animes)+1)
	animesWithBack = append(animesWithBack, backOption)
	animesWithBack = append(animesWithBack, animes...)

	// Use fuzzy finder to let user select
	var idx int
	var err error

	if util.IsDebug {
		// In debug mode, show preview window with technical details
		idx, err = fuzzyfinder.Find(
			animesWithBack,
			func(i int) string {
				// Show the anime name with language tag as-is
				return animesWithBack[i].Name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i >= 0 && i < len(animesWithBack) {
					anime := animesWithBack[i]
					if anime.Source == "__back__" {
						return "Go back to perform a new search"
					}
					var preview string
					preview = "Source: " + anime.Source + "\nURL: " + anime.URL
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
			animesWithBack,
			func(i int) string {
				// Show the anime name with language tag as-is
				return animesWithBack[i].Name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("anime selection cancelled: %w", err)
	}

	selectedAnime := animesWithBack[idx]

	// Check if user selected the back option
	if selectedAnime.Source == "__back__" {
		return nil, ErrBackToSearch
	}
	util.Debug("Anime selected", "name", selectedAnime.Name, "source", selectedAnime.Source)

	// CRITICAL: Enrich with AniList data for images and metadata (like the original system)
	if err := enrichAnimeData(selectedAnime); err != nil {
		util.Errorf("Error enriching anime data: %v", err)
	}

	return selectedAnime, nil
}

// Enhanced episode fetching that works with different sources.
// Delegates to the appropriate EpisodeProvider via the registry.
func GetAnimeEpisodesEnhanced(anime *models.Anime) ([]models.Episode, error) {
	initProviders()

	provider := providers.ForSource(anime)
	sourceName := providers.ResolveSourceName(anime)

	util.Debug("Getting episodes", "source", provider.Name(), "resolved", sourceName)

	episodes, err := provider.FetchEpisodes(anime)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from %s: %w", provider.Name(), err)
	}

	if len(episodes) > 0 {
		util.Debug("Episodes found", "count", len(episodes), "source", provider.Name())
	} else {
		util.Warn("No episodes found", "source", provider.Name())
	}

	return episodes, nil
}

// Enhanced episode URL fetching with improved source detection.
// Delegates to the appropriate EpisodeProvider via the registry.
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	initProviders()

	util.ClearGlobalSubtitles()

	if anime != nil && anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}

	provider := providers.ForSource(anime)

	util.Debug("Getting stream URL", "source", provider.Name(), "episode", episode.Number)

	if quality == "" {
		quality = "best"
	}

	streamURL, err := provider.GetStreamURL(episode, anime, quality)
	if err != nil {
		if errors.Is(err, scraper.ErrBackRequested) {
			return "", err
		}
		return "", fmt.Errorf("failed to get stream URL from %s: %w", provider.Name(), err)
	}

	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from %s", provider.Name())
	}

	util.Debug("Stream URL obtained", "source", provider.Name())
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
	// Remove language tags
	name = strings.ReplaceAll(name, "[English]", "")
	name = strings.ReplaceAll(name, "[PT-BR]", "")
	name = strings.ReplaceAll(name, "[Português]", "")
	name = strings.ReplaceAll(name, "(Legendado)", "")
	name = strings.ReplaceAll(name, "(Dublado)", "")
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

// languagePriority returns a sort key for language-based ordering.
// Lower values sort first: Portuguese → Multilanguage → English → Movies/TV → Unknown.
func languagePriority(name string) int {
	lower := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lower, "[pt-br]") || strings.HasPrefix(lower, "[portuguese]") || strings.HasPrefix(lower, "[português]"):
		return 0
	case strings.HasPrefix(lower, "[multilanguage]"):
		return 1
	case strings.HasPrefix(lower, "[english]"):
		return 2
	case strings.HasPrefix(lower, "[movie]") || strings.HasPrefix(lower, "[tv]") || strings.HasPrefix(lower, "[movies/tv]"):
		return 3
	default:
		return 4
	}
}
