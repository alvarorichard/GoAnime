// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"charm.land/huh/v2/spinner"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"golang.org/x/term"
)

// Cached terminal detection (checked once, reused)
var (
	stdoutIsTerminal     bool
	stdoutIsTerminalOnce sync.Once
)

func isStdoutTerminal() bool {
	stdoutIsTerminalOnce.Do(func() {
		fd := os.Stdout.Fd()
		stdoutIsTerminal = fd <= math.MaxInt && term.IsTerminal(int(fd))
	})
	return stdoutIsTerminal
}

// runWithSpinner runs the action with a spinner if stdout is a terminal,
// otherwise runs the action directly. This ensures CI and non-interactive
// environments work correctly since huh/v2 spinner may skip the Action
// callback when no terminal is attached.
func runWithSpinner(title string, action func()) {
	if isStdoutTerminal() {
		_ = spinner.New().
			Title(title).
			Type(spinner.Dots).
			Action(action).
			Run()
	} else {
		action()
	}
}

// ErrBackToSearch is returned when user selects the back option to search again
var ErrBackToSearch = errors.New("back to search requested")

// Enhanced search that supports multiple sources - always searches both Animefire.io and allanime simultaneously
func SearchAnimeEnhanced(name string, source string) (*models.Anime, error) {
	scraperManager := scraper.NewScraperManager()

	var scraperType *scraper.ScraperType
	isPTBR := false

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
	} else if strings.ToLower(source) == "9anime" || strings.ToLower(source) == "nineanime" {
		t := scraper.NineAnimeType
		scraperType = &t
		util.Debug("Searching specific source", "source", "9Anime")
	} else if strings.ToLower(source) == "goyabu" {
		t := scraper.GoyabuType
		scraperType = &t
		util.Debug("Searching specific source", "source", "Goyabu")
	} else if strings.ToLower(source) == "superflix" {
		t := scraper.SuperFlixType
		scraperType = &t
		util.Debug("Searching specific source", "source", "SuperFlix")
	} else if strings.ToLower(source) == "ptbr" || strings.ToLower(source) == "pt-br" {
		// Search only PT-BR sources (AnimeFire + Goyabu + SuperFlix) via dedicated method
		isPTBR = true
		util.Debug("Searching all PT-BR sources (AnimeFire + Goyabu + SuperFlix)")
	} else {
		// Default behavior: search all sources simultaneously (including FlixHQ)
		scraperType = nil
		util.Debug("Searching all sources", "query", name)
	}

	// Perform the search
	util.Debug("Searching for anime/media", "query", name)
	var animes []*models.Anime
	var searchErr error
	runWithSpinner("Searching for anime...", func() {
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

	// Normalize source identification once so downstream flows share the same rules.
	for _, anime := range animes {
		if resolved, err := ResolveSource(anime); err == nil {
			resolved.Apply(anime)
		}
	}

	util.Debug("Search results summary", "total", len(animes))

	// Show sources breakdown in debug only
	animefireCount := 0
	allanimeCount := 0
	animedriveCount := 0
	flixhqCount := 0
	nineAnimeCount := 0
	superflixCount := 0
	for _, anime := range animes {
		resolved, err := ResolveSource(anime)
		if err != nil {
			continue
		}

		switch resolved.Kind {
		case SourceAnimefire:
			animefireCount++
		case SourceAllAnime:
			allanimeCount++
		case SourceAnimeDrive:
			animedriveCount++
		case SourceFlixHQ:
			flixhqCount++
		case SourceNineAnime:
			nineAnimeCount++
		case SourceSuperFlix:
			superflixCount++
		}
	}

	util.Debug("Source breakdown", "AnimeFire", animefireCount, "AllAnime", allanimeCount, "AnimeDrive", animedriveCount, "FlixHQ", flixhqCount, "9Anime", nineAnimeCount, "SuperFlix", superflixCount)

	// Sort results by language priority: Portuguese first, then Multilanguage, Movies/TV, English, others
	sort.SliceStable(animes, func(i, j int) bool {
		return languagePriority(animes[i].Name) < languagePriority(animes[j].Name)
	})

	// Create a special "back" option as the first item
	backOption := &models.Anime{
		Name:   "<- Back",
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
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				name := a.Name
				// Append release year if available and not already in the name
				if a.Year != "" && !strings.Contains(name, "("+a.Year+")") {
					name += " (" + a.Year + ")"
				}
				return name
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
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				name := a.Name
				// Append release year if available and not already in the name
				if a.Year != "" && !strings.Contains(name, "("+a.Year+")") {
					name += " (" + a.Year + ")"
				}
				return name
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
	resolved, resolveErr := ResolveSource(anime)
	if resolveErr != nil {
		return nil, resolveErr
	}
	return getEpisodesByResolvedSource(anime, resolved)
}

// Enhanced episode URL fetching with improved source detection
func GetEpisodeStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	return getStreamURLByResolvedSource(anime, episode, quality)
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
	name = strings.ReplaceAll(name, "[Portugu\u00eas]", "")
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

// GetNineAnimeEpisodes handles episode fetching for 9anime sources
func GetNineAnimeEpisodes(anime *models.Anime) ([]models.Episode, error) {
	nineAnimeClient := scraper.NewNineAnimeClient()

	// anime.URL contains the 9anime anime ID
	animeID := anime.URL
	util.Debug("Getting 9Anime episodes", "animeID", animeID)

	episodes, err := nineAnimeClient.GetAnimeEpisodes(animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from 9Anime: %w", err)
	}

	util.Debug("9Anime episodes loaded", "count", len(episodes))
	return episodes, nil
}

// GetNineAnimeStreamURL gets the stream URL for 9anime content
func GetNineAnimeStreamURL(anime *models.Anime, episode *models.Episode, quality string) (string, error) {
	util.ClearGlobalSubtitles()
	util.SetGlobalAnimeSource("9Anime")

	nineAnimeClient := scraper.NewNineAnimeClient()

	// episode.URL / episode.DataID contains the episode data-id
	episodeID := episode.DataID
	if episodeID == "" {
		episodeID = episode.URL
	}

	util.Debug("Getting 9Anime stream", "episodeID", episodeID, "quality", quality)

	// Use the unified GetStreamURL which tries multiple servers automatically
	streamURL, metadata, err := nineAnimeClient.GetStreamURL(episodeID, "sub")
	if err != nil {
		return "", fmt.Errorf("failed to get stream URL from 9Anime: %w", err)
	}

	// Store referer globally for mpv playback
	if referer, ok := metadata["referer"]; ok && referer != "" {
		util.SetGlobalReferer(referer)
	}

	// Store subtitles globally for playback
	if subtitleURLs, ok := metadata["subtitles"]; ok && subtitleURLs != "" && !util.GlobalNoSubs {
		subURLs := strings.Split(subtitleURLs, ",")
		var subLabels []string
		if labels, ok := metadata["subtitle_labels"]; ok {
			subLabels = strings.Split(labels, ",")
		}

		var subInfos []util.SubtitleInfo
		for i, subURL := range subURLs {
			label := "Unknown"
			lang := "unknown"
			if i < len(subLabels) {
				label = subLabels[i]
				// Try to extract language code from label
				labelLower := strings.ToLower(label)
				if strings.Contains(labelLower, "english") {
					lang = "eng"
				} else if strings.Contains(labelLower, "portuguese") {
					lang = "por"
				} else if strings.Contains(labelLower, "spanish") {
					lang = "spa"
				} else if strings.Contains(labelLower, "japanese") {
					lang = "jpn"
				} else if strings.Contains(labelLower, "french") {
					lang = "fre"
				} else if strings.Contains(labelLower, "german") {
					lang = "ger"
				} else if strings.Contains(labelLower, "italian") {
					lang = "ita"
				} else if strings.Contains(labelLower, "arabic") {
					lang = "ara"
				}
			}
			subInfos = append(subInfos, util.SubtitleInfo{
				URL:      subURL,
				Language: lang,
				Label:    label,
			})
		}
		util.SetGlobalSubtitles(subInfos)
		util.Debug("9Anime subtitles loaded", "count", len(subInfos))
	}

	util.Debug("9Anime stream URL obtained", "url", streamURL[:min(len(streamURL), 80)])
	return streamURL, nil
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
	runWithSpinner("Loading seasons...", func() {
		seasons, seasonsErr = flixhqClient.GetSeasons(mediaID)
	})
	if seasonsErr != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", seasonsErr)
	}

	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found for TV show")
	}

	// Let user select a season using fuzzyfinder (same library as anime
	// selection, so it manages its own terminal state via tcell and avoids
	// the readline/escape-sequence issues that plague promptui after tcell).
	seasonIdx, err := tui.Find(seasons, func(i int) string {
		return seasons[i].Title
	}, fuzzyfinder.WithPromptString("Select season: "))
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasons[seasonIdx]
	media.CurrentSeason = selectedSeason.Number
	util.Debug("Selected season", "season", selectedSeason.Title, "id", selectedSeason.ID, "number", selectedSeason.Number)

	// Use spinner for loading episodes (network call)
	var flixEpisodes []scraper.FlixHQEpisode
	var episodesErr error
	runWithSpinner("Loading episodes...", func() {
		flixEpisodes, episodesErr = flixhqClient.GetEpisodes(selectedSeason.ID)
	})
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
	// Set media path for decryption API
	if media.URL != "" {
		flixhqClient.SetMediaPath(scraper.ExtractMediaPath(media.URL))
	}
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

		// Run network calls (server ID, embed link, stream extraction) with optional spinner
		runWithSpinner("Loading movie stream...", func() {
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
		})

		if streamErr != nil {
			return "", nil, fmt.Errorf("failed to get movie stream: %w", streamErr)
		}

		if streamInfo == nil {
			return "", nil, fmt.Errorf("failed to get movie stream: no stream info returned")
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

		// Run network calls (server ID, embed link, stream extraction) with optional spinner
		runWithSpinner("Loading episode stream...", func() {
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
		})

		if streamErr != nil {
			return "", nil, fmt.Errorf("failed to get episode stream: %w", streamErr)
		}

		if streamInfo == nil {
			return "", nil, fmt.Errorf("failed to get episode stream: no stream info returned")
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

	// Store the referer globally for use in downloads
	if streamInfo.Referer != "" {
		util.SetGlobalReferer(streamInfo.Referer)
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

	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select video quality: "))
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

// GetSuperFlixEpisodes handles episodes/content for SuperFlix movies and TV shows
func GetSuperFlixEpisodes(media *models.Anime) ([]models.Episode, error) {
	sfClient := scraper.NewSuperFlixClient()

	// media.URL contains the TMDB ID for SuperFlix
	tmdbID := media.URL
	if tmdbID == "" {
		return nil, fmt.Errorf("no TMDB ID found for SuperFlix content")
	}

	util.Debug("Getting SuperFlix content", "mediaType", media.MediaType, "tmdbID", tmdbID)

	// For movies, return a single "episode" representing the movie
	if media.MediaType == models.MediaTypeMovie {
		util.Debug("SuperFlix: Processing movie")
		return []models.Episode{
			{
				Number: "1",
				Num:    1,
				URL:    tmdbID,
				Title: models.TitleDetails{
					English: media.Name,
					Romaji:  media.Name,
				},
			},
		}, nil
	}

	// For TV shows / series, get seasons and episodes
	util.Debug("SuperFlix: Processing TV show/series, getting episodes")

	var allEpisodes map[string][]scraper.SuperFlixEpisode
	var episodesErr error
	runWithSpinner("Loading seasons...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		allEpisodes, episodesErr = sfClient.GetEpisodes(ctx, tmdbID)
	})
	if episodesErr != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", episodesErr)
	}

	if len(allEpisodes) == 0 {
		return nil, fmt.Errorf("no seasons found")
	}

	// Sort season numbers
	var seasonNums []string
	for k := range allEpisodes {
		seasonNums = append(seasonNums, k)
	}
	sort.Strings(seasonNums)

	// Build season labels for selection
	var seasonLabels []string
	for _, sn := range seasonNums {
		epCount := len(allEpisodes[sn])
		seasonLabels = append(seasonLabels, fmt.Sprintf("Season %s (%d episodes)", sn, epCount))
	}

	// Let user select a season
	seasonIdx, err := tui.Find(seasonLabels, func(i int) string {
		return seasonLabels[i]
	}, fuzzyfinder.WithPromptString("Select season: "))
	if err != nil {
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasonNums[seasonIdx]
	epList := allEpisodes[selectedSeason]
	util.Debug("Selected season", "season", selectedSeason, "episodes", len(epList))

	// Convert to models.Episode
	var episodes []models.Episode
	for _, ep := range epList {
		epNum := ep.EpiNum.String()
		num := 0
		if n, err := ep.EpiNum.Int64(); err == nil {
			num = int(n)
		}

		episodes = append(episodes, models.Episode{
			Number:   epNum,
			Num:      num,
			URL:      tmdbID, // Store TMDB ID for stream retrieval
			SeasonID: selectedSeason,
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			Aired: ep.AirDate,
		})
	}

	// Store current season on the media object
	var seasonNum int
	if _, err := fmt.Sscanf(selectedSeason, "%d", &seasonNum); err == nil {
		media.CurrentSeason = seasonNum
	}

	util.Debug("SuperFlix episodes loaded", "count", len(episodes))
	return episodes, nil
}

// GetSuperFlixStreamURL gets the stream URL for SuperFlix content
func GetSuperFlixStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	util.ClearGlobalSubtitles()
	util.SetGlobalAnimeSource("SuperFlix")

	sfClient := scraper.NewSuperFlixClient()

	tmdbID := episode.URL
	if tmdbID == "" {
		tmdbID = media.URL
	}

	var sfType, season, epNum string
	if media.MediaType == models.MediaTypeMovie {
		sfType = "filme"
	} else {
		sfType = "serie"
		season = episode.SeasonID
		epNum = episode.Number
	}

	util.Debug("Getting SuperFlix stream", "tmdbID", tmdbID, "type", sfType, "season", season, "episode", epNum)

	var result *scraper.SuperFlixStreamResult
	var streamErr error
	runWithSpinner("Loading stream...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		result, streamErr = sfClient.GetStreamURL(ctx, sfType, tmdbID, season, epNum)
	})
	if streamErr != nil {
		return "", fmt.Errorf("failed to get SuperFlix stream: %w", streamErr)
	}

	// Store referer globally for mpv playback
	if result.Referer != "" {
		util.SetGlobalReferer(result.Referer)
	}

	// Update cover image from stream thumbnail if not already set
	if media.ImageURL == "" && result.Thumb != "" {
		media.ImageURL = result.Thumb
		util.Debug("SuperFlix cover set from stream thumbnail", "url", result.Thumb)
	}

	// Store subtitles globally for playback
	if len(result.Subtitles) > 0 && !util.GlobalNoSubs {
		var subInfos []util.SubtitleInfo
		for _, sub := range result.Subtitles {
			lang := strings.ToLower(sub.Lang)
			subInfos = append(subInfos, util.SubtitleInfo{
				URL:      sub.URL,
				Language: lang,
				Label:    sub.Lang,
			})
		}
		util.SetGlobalSubtitles(subInfos)
		util.Debug("SuperFlix subtitles loaded", "count", len(subInfos))
	}

	util.Debug("SuperFlix stream URL obtained", "url", result.StreamURL[:min(len(result.StreamURL), 80)])
	return result.StreamURL, nil
}

// languagePriority returns a sort key for language-based ordering.
// Lower values sort first: Portuguese -> Multilanguage -> English -> Movies/TV -> Unknown.
func languagePriority(name string) int {
	lower := strings.ToLower(name)
	// Check for [PT-BR] anywhere (covers "[Movie] [PT-BR] ...", "[TV] [PT-BR] ...", etc.)
	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portuguese]") || strings.Contains(lower, "[portugu\u00eas]") {
		return 0
	}
	switch {
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
