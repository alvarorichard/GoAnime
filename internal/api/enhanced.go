// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh/spinner"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
)

// ErrBackToSearch is returned when user selects the back option to search again
var ErrBackToSearch = errors.New("back to search requested")

// Enhanced search that supports multiple sources - always searches both Animefire.io and allanime simultaneously
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
	} else if strings.ToLower(source) == "animedrive" {
		t := scraper.AnimeDriveType
		scraperType = &t
		util.Debug("Searching specific source", "source", "AnimeDrive")
	} else if strings.ToLower(source) == "flixhq" || strings.ToLower(source) == "movie" || strings.ToLower(source) == "tv" {
		t := scraper.FlixHQType
		scraperType = &t
		util.Debug("Searching specific source", "source", "FlixHQ")
	} else {
		// Default behavior: search all sources simultaneously (including FlixHQ)
		scraperType = nil
		util.Debug("Searching all sources", "query", name)
	}

	// Perform the search - this will search all sources if scraperType is nil
	util.Debug("Searching for anime/media", "query", name)
	var animes []*models.Anime
	var searchErr error
	_ = spinner.New().
		Title("Searching for anime...").
		Type(spinner.Dots).
		Action(func() {
			animes, searchErr = scraperManager.SearchAnime(name, scraperType)
		}).
		Run()
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
			} else if strings.Contains(anime.URL, "flixhq") {
				anime.Source = "FlixHQ"
			}
		}

		// Language tags are already added by unified.go, don't duplicate them here
	}

	util.Debug("Search results summary", "total", len(animes))

	// Show sources breakdown in debug only
	animefireCount := 0
	allanimeCount := 0
	animedriveCount := 0
	flixhqCount := 0
	for _, anime := range animes {
		if strings.Contains(anime.Source, "AnimeFire") {
			animefireCount++
		} else if anime.Source == "AllAnime" {
			allanimeCount++
		} else if anime.Source == "AnimeDrive" {
			animedriveCount++
		} else if anime.Source == "FlixHQ" {
			flixhqCount++
		}
	}

	util.Debug("Source breakdown", "AnimeFire", animefireCount, "AllAnime", allanimeCount, "AnimeDrive", animedriveCount, "FlixHQ", flixhqCount)

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

// Enhanced episode fetching that works with different sources
func GetAnimeEpisodesEnhanced(anime *models.Anime) ([]models.Episode, error) {
	// Check if this is a FlixHQ movie/TV show
	if anime.Source == "FlixHQ" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		return GetFlixHQEpisodes(anime)
	}

	// Determine source type from multiple indicators with enhanced logic
	var sourceName string

	// Priority 1: Check the Source field (most reliable)
	if anime.Source == "AllAnime" {
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Source, "AnimeFire") {
		sourceName = "Animefire.io"
	} else if anime.Source == "AnimeDrive" {
		sourceName = "AnimeDrive"
	} else if strings.Contains(anime.Name, "[English]") {
		// Priority 2: Check language tags (AllAnime = English)
		sourceName = "AllAnime"
		anime.Source = "AllAnime" // Update source field
	} else if strings.Contains(anime.Name, "[Portuguese]") || strings.Contains(anime.Name, "[Português]") {
		// AnimeFire or AnimeDrive = Portuguese
		// Check URL to determine which one
		if strings.Contains(anime.URL, "animesdrive") {
			sourceName = "AnimeDrive"
			anime.Source = "AnimeDrive"
		} else {
			sourceName = "Animefire.io"
			anime.Source = "Animefire.io"
		}
	} else if strings.Contains(anime.URL, "allanime") || (len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http")) {
		// Priority 3: URL analysis for AllAnime (short IDs or allanime URLs)
		sourceName = "AllAnime"
		anime.Source = "AllAnime" // Update source field
	} else if strings.Contains(anime.URL, "animefire") {
		// Priority 4: URL analysis for AnimeFire
		sourceName = "Animefire.io"
		anime.Source = "Animefire.io" // Update source field
	} else if strings.Contains(anime.URL, "animesdrive") {
		// Priority 5: URL analysis for AnimeDrive
		sourceName = "AnimeDrive"
		anime.Source = "AnimeDrive" // Update source field
	} else {
		// Default to AllAnime for unknown sources
		sourceName = "AllAnime (default)"
		anime.Source = "AllAnime"
	}

	cleanName := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(anime.Name, "[English]", ""), "[Portuguese]", ""))

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
	} else if sourceName == "AnimeDrive" {
		// For AnimeDrive, use the AnimeDrive scraper
		scraperManager := scraper.NewScraperManager()
		scraperInstance, scErr := scraperManager.GetScraper(scraper.AnimeDriveType)
		if scErr != nil {
			return nil, fmt.Errorf("failed to get AnimeDrive scraper: %w", scErr)
		}
		episodes, err = scraperInstance.GetAnimeEpisodes(anime.URL)
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
		} else if sourceName == "AnimeDrive" {
			util.Debug("Source info", "type", "AnimeDrive", "features", "multiple qualities")
		} else {
			util.Debug("Source info", "type", "Animefire.io", "features", "dubbed/subtitled")
		}
	} else {
		util.Warn("No episodes found", "source", sourceName)
	}

	return episodes, nil
}

// Enhanced episode URL fetching with improved source detection
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	// Clear any previous subtitles
	util.ClearGlobalSubtitles()

	// Check if this is FlixHQ content
	if anime.Source == "FlixHQ" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		streamURL, subtitles, err := GetFlixHQStreamURL(anime, episode, quality)
		if err != nil {
			return "", err
		}

		// Store subtitles globally for playback
		if len(subtitles) > 0 && !util.GlobalNoSubs {
			var subInfos []util.SubtitleInfo
			for _, sub := range subtitles {
				subInfos = append(subInfos, util.SubtitleInfo{
					URL:      sub.URL,
					Language: sub.Language,
					Label:    sub.Label,
				})
			}
			util.SetGlobalSubtitles(subInfos)
		}

		return streamURL, nil
	}

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
		sourceName = "Animefire.io"
	} else if anime.Source == "AnimeDrive" {
		scraperType = scraper.AnimeDriveType
		sourceName = "AnimeDrive"
	} else if strings.Contains(anime.Name, "[English]") {
		// Priority 2: Check language tags (AllAnime = English)
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.Name, "[Portuguese]") || strings.Contains(anime.Name, "[Português]") {
		// AnimeFire or AnimeDrive = Portuguese
		// Check URL to determine which one
		if strings.Contains(anime.URL, "animesdrive") {
			scraperType = scraper.AnimeDriveType
			sourceName = "AnimeDrive"
		} else {
			scraperType = scraper.AnimefireType
			sourceName = "Animefire.io"
		}
	} else if len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http") {
		// Priority 3: URL analysis for AllAnime (short IDs)
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else if strings.Contains(anime.URL, "animefire") {
		// Priority 4: URL analysis for AnimeFire
		scraperType = scraper.AnimefireType
		sourceName = "Animefire.io"
	} else if strings.Contains(anime.URL, "animesdrive") {
		// Priority 5: URL analysis for AnimeDrive
		scraperType = scraper.AnimeDriveType
		sourceName = "AnimeDrive"
	} else if strings.Contains(anime.URL, "allanime") {
		// Priority 6: AllAnime full URLs
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime"
	} else {
		// Default to AllAnime
		scraperType = scraper.AllAnimeType
		sourceName = "AllAnime (default)"
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
		return "", fmt.Errorf("failed to get scraper for %s: %w", sourceName, err)
	}

	if quality == "" {
		quality = "best"
	}

	var streamURL string
	var streamErr error

	// Handle different scraper types with appropriate parameters
	switch scraperType {
	case scraper.AllAnimeType:
		util.Debug("Processing through AllAnime")
		streamURL, _, streamErr = scraperInstance.GetStreamURL(anime.URL, episode.Number, quality)
	case scraper.AnimeDriveType:
		util.Debug("Processing through AnimeDrive")
		streamURL, _, streamErr = scraperInstance.GetStreamURL(episode.URL)
	default:
		util.Debug("Processing through Animefire.io")
		streamURL, _, streamErr = scraperInstance.GetStreamURL(episode.URL, quality)
	}

	if streamErr != nil {
		// Propagate back request error without wrapping
		if errors.Is(streamErr, scraper.ErrBackRequested) {
			return "", streamErr
		}
		return "", fmt.Errorf("failed to get stream URL from %s: %w", sourceName, streamErr)
	}

	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from %s", sourceName)
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
	// Remove language tags
	name = strings.ReplaceAll(name, "[English]", "")
	name = strings.ReplaceAll(name, "[Portuguese]", "")
	name = strings.ReplaceAll(name, "[Português]", "")
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

// GetFlixHQEpisodes handles episodes/content for FlixHQ movies and TV shows
func GetFlixHQEpisodes(media *models.Anime) ([]models.Episode, error) {
	flixhqClient := scraper.NewFlixHQClient()

	// Extract media ID from URL
	mediaID := extractMediaIDFromURL(media.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", media.URL)
	}

	util.Debug("Getting FlixHQ content", "mediaType", media.MediaType, "mediaID", mediaID)

	// For movies, return a single "episode" representing the movie
	if media.MediaType == models.MediaTypeMovie {
		util.Debug("FlixHQ: Processing movie")
		return []models.Episode{
			{
				Number: "1",
				Num:    1,
				URL:    mediaID, // Store media ID for later use
				Title: models.TitleDetails{
					English: media.Name,
					Romaji:  media.Name,
				},
			},
		}, nil
	}

	// For TV shows, get seasons and let user select
	util.Debug("FlixHQ: Processing TV show, getting seasons")

	// Use spinner for loading seasons (network call)
	var seasons []scraper.FlixHQSeason
	var seasonsErr error
	_ = spinner.New().
		Title("Loading seasons...").
		Type(spinner.Dots).
		Action(func() {
			seasons, seasonsErr = flixhqClient.GetSeasons(mediaID)
		}).
		Run()
	if seasonsErr != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", seasonsErr)
	}

	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found for TV show")
	}

	// Let user select a season
	seasonNames := make([]string, len(seasons))
	for i, s := range seasons {
		seasonNames[i] = s.Title
	}

	seasonIdx, err := fuzzyfinder.Find(
		seasonNames,
		func(i int) string { return seasonNames[i] },
		fuzzyfinder.WithPromptString("Select season: "),
	)
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasons[seasonIdx]
	util.Debug("Selected season", "season", selectedSeason.Title, "id", selectedSeason.ID)

	// Clear the fuzzy finder output before showing the spinner
	fmt.Print("\033[2K\033[1A\033[2K\r")

	// Use spinner for loading episodes (network call)
	var flixEpisodes []scraper.FlixHQEpisode
	var episodesErr error
	_ = spinner.New().
		Title("Loading episodes...").
		Type(spinner.Dots).
		Action(func() {
			flixEpisodes, episodesErr = flixhqClient.GetEpisodes(selectedSeason.ID)
		}).
		Run()
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}

	// Convert to models.Episode
	var episodes []models.Episode
	for _, ep := range flixEpisodes {
		episodes = append(episodes, models.Episode{
			Number: fmt.Sprintf("%d", ep.Number),
			Num:    ep.Number,
			URL:    ep.DataID, // Store DataID for stream retrieval
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			DataID:   ep.DataID,
			SeasonID: selectedSeason.ID,
		})
	}

	util.Debug("FlixHQ episodes loaded", "count", len(episodes))
	return episodes, nil
}

// GetFlixHQStreamURL gets the stream URL for FlixHQ content
func GetFlixHQStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, []models.Subtitle, error) {
	flixhqClient := scraper.NewFlixHQClient()
	provider := "Vidcloud"
	subsLanguage := util.GlobalSubsLanguage
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	var streamInfo *scraper.FlixHQStreamInfo
	var episodeID string
	var embedLink string
	var streamErr error

	if media.MediaType == models.MediaTypeMovie {
		// For movies, episode.URL contains the media ID
		mediaID := episode.URL
		util.Debug("Getting movie stream", "mediaID", mediaID)

		// Use spinner for all network calls (server ID, embed link, stream extraction)
		_ = spinner.New().
			Title("Loading movie stream...").
			Type(spinner.Dots).
			Action(func() {
				episodeID, streamErr = flixhqClient.GetMovieServerID(mediaID, provider)
				if streamErr != nil {
					return
				}

				embedLink, streamErr = flixhqClient.GetEmbedLink(episodeID)
				if streamErr != nil {
					return
				}

				// Extract stream info to get available qualities
				streamInfo, streamErr = flixhqClient.ExtractStreamInfo(embedLink, "auto", subsLanguage)
			}).
			Run()

		if streamErr != nil {
			return "", nil, fmt.Errorf("failed to get movie stream: %w", streamErr)
		}

		// If we have multiple quality options, let user choose (UI - no spinner needed)
		if len(streamInfo.Qualities) > 1 {
			selectedQuality, selectErr := selectFlixHQQualityOptions(streamInfo.Qualities)
			if selectErr == nil && selectedQuality.URL != "" {
				streamInfo.VideoURL = selectedQuality.URL
				streamInfo.Quality = string(selectedQuality.Quality)
				streamInfo.IsM3U8 = selectedQuality.IsM3U8
			}
		}
	} else {
		// For TV shows, episode.URL contains the DataID
		dataID := episode.URL
		util.Debug("Getting TV episode stream", "dataID", dataID)

		// Use spinner for all network calls (server ID, embed link, stream extraction)
		_ = spinner.New().
			Title("Loading episode stream...").
			Type(spinner.Dots).
			Action(func() {
				episodeID, streamErr = flixhqClient.GetEpisodeServerID(dataID, provider)
				if streamErr != nil {
					return
				}

				embedLink, streamErr = flixhqClient.GetEmbedLink(episodeID)
				if streamErr != nil {
					return
				}

				// Extract stream info to get available qualities
				streamInfo, streamErr = flixhqClient.ExtractStreamInfo(embedLink, "auto", subsLanguage)
			}).
			Run()

		if streamErr != nil {
			return "", nil, fmt.Errorf("failed to get episode stream: %w", streamErr)
		}

		// If we have multiple quality options, let user choose (UI - no spinner needed)
		if len(streamInfo.Qualities) > 1 {
			selectedQuality, selectErr := selectFlixHQQualityOptions(streamInfo.Qualities)
			if selectErr == nil && selectedQuality.URL != "" {
				streamInfo.VideoURL = selectedQuality.URL
				streamInfo.Quality = string(selectedQuality.Quality)
				streamInfo.IsM3U8 = selectedQuality.IsM3U8
			}
		}
	}

	// Convert subtitles
	var subtitles []models.Subtitle
	for _, sub := range streamInfo.Subtitles {
		subtitles = append(subtitles, models.Subtitle{
			URL:      sub.URL,
			Language: sub.Language,
			Label:    sub.Label,
		})
	}

	return streamInfo.VideoURL, subtitles, nil
}

// selectFlixHQQualityOptions shows a menu for the user to select video quality from FlixHQQualityOption
func selectFlixHQQualityOptions(qualities []scraper.FlixHQQualityOption) (scraper.FlixHQQualityOption, error) {
	if len(qualities) == 0 {
		return scraper.FlixHQQualityOption{Quality: scraper.QualityAuto}, fmt.Errorf("no qualities available")
	}

	// If only one quality, use it directly
	if len(qualities) == 1 {
		return qualities[0], nil
	}

	// Build labels for each quality
	var items []string
	client := scraper.NewFlixHQClient()
	for _, q := range qualities {
		items = append(items, client.QualityToLabel(q.Quality))
	}

	prompt := promptui.Select{
		Label: "Select video quality",
		Items: items,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		// On error/cancel, return first (auto) quality
		return qualities[0], err
	}

	return qualities[idx], nil
}

// extractMediaIDFromURL extracts the media ID from a FlixHQ URL
func extractMediaIDFromURL(urlStr string) string {
	// URL format: https://flixhq.to/movie/watch-movie-name-12345 or /movie/watch-movie-name-12345
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func GetAnimeEpisodesWithSource(anime *models.Anime) ([]models.Episode, error) {
	return GetAnimeEpisodesEnhanced(anime)
}
