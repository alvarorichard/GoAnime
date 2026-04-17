// Package download provides high-level download workflow management
package download

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/movie"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// HandleDownloadRequest processes a download request from command line
func HandleDownloadRequest(request *util.DownloadRequest) error {
	util.Info("Starting enhanced download mode...")

	// Use source preference if specified
	source := request.Source
	quality := request.Quality
	if quality == "" {
		quality = "best"
	}

	util.Infof("Using source: %s, quality: %s", source, quality)

	// Try enhanced search with retry logic
	anime, err := appflow.SearchAnimeWithRetry(request.AnimeName)
	if err != nil {
		util.Errorf("Failed to search for anime: %v", err)
		return err
	}

	// Set anime name for Plex-compatible download file naming
	season := 1
	if request.SeasonNum > 0 {
		season = request.SeasonNum
	}
	player.SetAnimeName(anime.Name, season)
	// Route downloads to the correct directory (anime/ vs movies/) using exact media type
	player.SetExactMediaType(string(anime.MediaType))

	// Build and store external IDs for Plex/Jellyfin-compatible folder naming
	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	// Enrich with AniList metadata for per-episode season resolution
	enricher := metadata.NewEnricher()
	seasonMap, _ := enricher.EnrichAnime(context.Background(), anime)
	player.SetSeasonMap(seasonMap)

	// Update metadata after enrichment (AniList may have populated IDs)
	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	// If this is a movie from FlixHQ/SFlix, redirect to the movie download workflow
	// Movies should not go through the episode-based download path
	if anime.IsMovie() {
		util.Infof("Detected movie content: %s — redirecting to movie download workflow", anime.Name)
		movieRequest := &util.DownloadRequest{
			AnimeName:    request.AnimeName,
			IsMovie:      true,
			Quality:      quality,
			SubsLanguage: request.SubsLanguage,
			OutputDir:    request.OutputDir,
		}
		return HandleMovieDownloadRequest(movieRequest)
	}

	// If this is 9Anime content, use the dedicated 9anime downloader
	// 9Anime episodes use data-id based resolution that is incompatible with legacy downloaders
	if anime.Source == "9Anime" {
		util.Infof("Detected 9Anime content: %s — using 9Anime downloader", anime.Name)
		nad := downloader.NewNineAnimeDownloader(downloader.NineAnimeDownloadConfig{
			AnimeName:    anime.Name,
			Quality:      quality,
			OutputDir:    request.OutputDir,
			Season:       season,
			SubsLanguage: request.SubsLanguage,
		})
		if request.IsAll {
			return nad.DownloadAllEpisodes(anime)
		}
		if request.IsRange {
			return nad.DownloadEpisodeRange(anime, request.StartEpisode, request.EndEpisode)
		}
		if request.EpisodeNum <= 0 {
			// No specific episode requested — download all episodes
			return nad.DownloadAllEpisodes(anime)
		}
		return nad.DownloadSingleEpisode(anime, request.EpisodeNum)
	}

	// Download-all mode: fetch all episodes and download them
	if request.IsAll {
		util.Infof("Downloading ALL episodes of %s", anime.Name)

		// Try enhanced episode fetch first, fallback to legacy
		eps, err := api.GetAnimeEpisodesEnhanced(anime)
		if err == nil && len(eps) > 0 {
			dlErr := player.HandleBatchDownload(eps, anime)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}

		// Fallback to legacy downloader
		episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadAllEpisodes()
	}

	if request.IsRange {
		util.Infof("Downloading episodes %d-%d of %s",
			request.StartEpisode, request.EndEpisode, anime.Name)

		// Exclusive AllAnime Smart Range
		if request.AllAnimeSmart && (anime.Source == "AllAnime" || source == "allanime" || source == "AllAnime") {
			util.Info("AllAnime Smart Range enabled: mirror priority + AniSkip integration + progress UI")
			// Use player batch downloader with provided range to get consistent progress UI
			eps, err := api.GetAnimeEpisodesEnhanced(anime)
			if err == nil && len(eps) > 0 {
				dlErr := player.HandleBatchDownloadRange(eps, anime, request.StartEpisode, request.EndEpisode)
				if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
					return nil
				}
				// Fall through to API-based smart range if UI path fails
				util.Infof("Progress UI path failed, falling back to API smart range: %v", dlErr)
			} else if err != nil {
				util.Infof("Enhanced episodes fetch failed for progress path: %v", err)
			}
			if err := api.DownloadAllAnimeSmartRange(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
				util.Errorf("AllAnime Smart Range failed: %v", err)
				// Fallback to normal enhanced
				if err := api.DownloadEpisodeRangeEnhanced(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
					util.Infof("Enhanced download failed, falling back to legacy: %v", err)
					// Fallback to legacy downloader
					episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
					if legacyErr != nil {
						return fmt.Errorf("legacy episode fetch also failed: %w", legacyErr)
					}
					dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
					return dl.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
				}
				return nil
			}
			return nil
		}

		// Try batch downloader with progress UI first (works for AllAnime and other sources)
		eps, err := api.GetAnimeEpisodesEnhanced(anime)
		if err == nil && len(eps) > 0 {
			dlErr := player.HandleBatchDownloadRange(eps, anime, request.StartEpisode, request.EndEpisode)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}
		// Fallback to legacy downloader
		episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
	} else {
		util.Infof("Downloading episode %d of %s",
			request.EpisodeNum, anime.Name)

		// Enhanced download is a placeholder - use legacy downloader
		util.Infof("Using legacy downloader for episode %d", request.EpisodeNum)
		episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadSingleEpisode(request.EpisodeNum)
	}
}

// Example usage functions for documentation

// ExampleSingleDownload demonstrates single episode download
//func ExampleSingleDownload() {
//	// Command: goanime -d "My Hero Academia" 15
//	// This would create a DownloadRequest like:
//	request := &util.DownloadRequest{
//		AnimeName:  "My Hero Academia",
//		EpisodeNum: 15,
//		IsRange:    false,
//	}
//
//	if err := HandleDownloadRequest(request); err != nil {
//		log.Printf("Download failed: %v", err)
//	}
//}

//// ExampleRangeDownload demonstrates episode range download
//func ExampleRangeDownload() {
//	// Command: goanime -d -r "Attack on Titan" 1-5
//	// This would create a DownloadRequest like:
//	request := &util.DownloadRequest{
//		AnimeName:    "Attack on Titan",
//		IsRange:      true,
//		StartEpisode: 1,
//		EndEpisode:   5,
//	}
//
//	if err := HandleDownloadRequest(request); err != nil {
//		log.Printf("Range download failed: %v", err)
//	}
//}

// HandleMovieDownloadRequest processes movie/TV download requests from FlixHQ and SFlix
func HandleMovieDownloadRequest(request *util.DownloadRequest) error {
	util.Info("Starting movie/TV download mode...")

	quality := request.Quality
	if quality == "" {
		quality = "1080"
	}

	subsLanguage := request.SubsLanguage
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	util.Infof("Searching for: %s (quality: %s)", request.AnimeName, quality)

	// Create media manager and search
	mediaManager := scraper.NewMediaManager()
	results, err := mediaManager.SearchMoviesAndTV(request.AnimeName)
	if err != nil {
		return fmt.Errorf("failed to search for movie/TV: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no results found for: %s", request.AnimeName)
	}

	// Let user select from results
	selectedMedia, err := selectMovieFromResults(results, request.IsMovie, request.IsTV)
	if err != nil {
		return fmt.Errorf("failed to select media: %w", err)
	}

	// Convert to models.Anime for compatibility with downloader
	anime := selectedMedia.ToAnimeModel()
	anime.Source = selectedMedia.Source

	// Enrich with TMDB/OMDb metadata to get official title, year, and external IDs.
	// This is essential for Plex/Jellyfin-compatible folder naming — without it,
	// folders use the scraped (often localized) name instead of the official title.
	if err := movie.EnrichMedia(anime); err != nil {
		util.Debugf("TMDB/OMDb enrichment failed (non-critical): %v", err)
	}

	// Set exact media type for intelligent path organization
	if selectedMedia.Type == scraper.MediaTypeMovie {
		player.SetExactMediaType("movie")
	} else {
		player.SetExactMediaType("tv")
	}
	player.SetAnimeName(anime.Name, request.SeasonNum)

	// Build and store external IDs for Plex/Jellyfin-compatible folder naming
	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	// Create movie downloader
	md := downloader.NewMovieDownloaderWithConfig(downloader.MovieDownloadConfig{
		Quality:      scraper.Quality(quality),
		SubsLanguage: subsLanguage,
		Provider:     "Vidcloud",
	})

	if request.IsMovie || selectedMedia.Type == scraper.MediaTypeMovie {
		// Download movie
		util.Infof("Downloading movie: %s", anime.Name)
		return md.DownloadMovie(anime)

	} else if request.IsTV || selectedMedia.Type == scraper.MediaTypeTV {
		// Download TV episode(s)

		// Download-all mode: get every season and every episode
		if request.IsAll {
			util.Infof("Downloading ALL seasons and episodes of %s", anime.Name)
			return md.DownloadAllSeasons(anime)
		}

		mediaID := extractIDFromURL(selectedMedia.URL)

		// Interactive mode: no specific season/episode/range/all flags were set
		if !request.IsRange && request.SeasonNum == 0 && request.EpisodeNum == 0 {
			var downloadMode string
			modeForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Download mode for: "+anime.Name).
						Options(
							huh.NewOption("Download ALL seasons and episodes", "all_seasons"),
							huh.NewOption("Download all episodes in a season", "all_in_season"),
							huh.NewOption("Download a single episode", "single"),
							huh.NewOption("Download a range of episodes", "range"),
						).
						Value(&downloadMode),
				),
			)

			if err := modeForm.Run(); err != nil {
				return fmt.Errorf("download mode selection cancelled: %w", err)
			}

			switch downloadMode {
			case "all_seasons":
				util.Infof("Downloading ALL seasons and episodes of %s", anime.Name)
				return md.DownloadAllSeasons(anime)

			case "all_in_season":
				seasonNum, sErr := selectSeason(mediaManager, mediaID)
				if sErr != nil {
					return fmt.Errorf("failed to select season: %w", sErr)
				}
				epCount, cErr := getSeasonEpisodeCount(mediaManager, mediaID, seasonNum)
				if cErr != nil {
					return fmt.Errorf("failed to get episode count: %w", cErr)
				}
				util.Infof("Downloading all %d episodes of %s Season %d", epCount, anime.Name, seasonNum)
				return md.DownloadTVEpisodeRange(anime, seasonNum, 1, epCount)

			case "single":
				seasonNum, sErr := selectSeason(mediaManager, mediaID)
				if sErr != nil {
					return fmt.Errorf("failed to select season: %w", sErr)
				}
				episodeNum, eErr := selectEpisode(mediaManager, mediaID, seasonNum)
				if eErr != nil {
					return fmt.Errorf("failed to select episode: %w", eErr)
				}
				util.Infof("Downloading %s S%02dE%02d", anime.Name, seasonNum, episodeNum)
				return md.DownloadTVEpisode(anime, seasonNum, episodeNum)

			case "range":
				seasonNum, sErr := selectSeason(mediaManager, mediaID)
				if sErr != nil {
					return fmt.Errorf("failed to select season: %w", sErr)
				}
				var startStr, endStr string
				rangeForm := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Start episode").
							Description("First episode number").
							Value(&startStr).
							Validate(func(v string) error {
								if n, parseErr := strconv.Atoi(v); parseErr != nil || n < 1 {
									return fmt.Errorf("enter a valid positive number")
								}
								return nil
							}),
						huh.NewInput().
							Title("End episode").
							Description("Last episode number").
							Value(&endStr).
							Validate(func(v string) error {
								if n, parseErr := strconv.Atoi(v); parseErr != nil || n < 1 {
									return fmt.Errorf("enter a valid positive number")
								}
								return nil
							}),
					),
				)
				if err := rangeForm.Run(); err != nil {
					return fmt.Errorf("range input cancelled: %w", err)
				}
				startEp, _ := strconv.Atoi(startStr)
				endEp, _ := strconv.Atoi(endStr)
				if startEp > endEp {
					return fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
				}
				util.Infof("Downloading %s S%02d E%02d-%02d", anime.Name, seasonNum, startEp, endEp)
				return md.DownloadTVEpisodeRange(anime, seasonNum, startEp, endEp)

			default:
				return fmt.Errorf("unknown download mode selected")
			}
		}

		seasonNum := request.SeasonNum
		if seasonNum == 0 {
			// Let user select season
			seasonNum, err = selectSeason(mediaManager, mediaID)
			if err != nil {
				return fmt.Errorf("failed to select season: %w", err)
			}
		}

		if request.IsRange {
			util.Infof("Downloading %s S%02d E%02d-%02d", anime.Name, seasonNum, request.StartEpisode, request.EndEpisode)
			return md.DownloadTVEpisodeRange(anime, seasonNum, request.StartEpisode, request.EndEpisode)
		} else {
			episodeNum := request.EpisodeNum
			if episodeNum == 0 {
				// Let user select episode
				episodeNum, err = selectEpisode(mediaManager, mediaID, seasonNum)
				if err != nil {
					return fmt.Errorf("failed to select episode: %w", err)
				}
			}
			util.Infof("Downloading %s S%02dE%02d", anime.Name, seasonNum, episodeNum)
			return md.DownloadTVEpisode(anime, seasonNum, episodeNum)
		}
	}

	return fmt.Errorf("could not determine media type for download")
}

// selectMovieFromResults presents a selection UI for movie/TV results
func selectMovieFromResults(results []*scraper.FlixHQMedia, preferMovie, preferTV bool) (*scraper.FlixHQMedia, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	// Filter results if preference is set
	var filtered []*scraper.FlixHQMedia
	if preferMovie {
		for _, r := range results {
			if r.Type == scraper.MediaTypeMovie {
				filtered = append(filtered, r)
			}
		}
	} else if preferTV {
		for _, r := range results {
			if r.Type == scraper.MediaTypeTV {
				filtered = append(filtered, r)
			}
		}
	}

	// If filtering removed all results, use all results
	if len(filtered) == 0 {
		filtered = results
	}

	// If only one result, return it
	if len(filtered) == 1 {
		return filtered[0], nil
	}

	// Prepare display items
	var items []string
	for _, r := range filtered {
		typeTag := "[Movie]"
		if r.Type == scraper.MediaTypeTV {
			typeTag = "[TV]"
		}
		year := ""
		if r.Year != "" {
			year = fmt.Sprintf(" (%s)", r.Year)
		}
		source := ""
		if r.Source != "" {
			source = fmt.Sprintf(" - %s", r.Source)
		}
		items = append(items, fmt.Sprintf("%s %s%s%s", typeTag, r.Title, year, source))
	}

	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select movie/TV show to download: "))
	if err != nil {
		return nil, err
	}

	return filtered[idx], nil
}

// selectSeason presents a selection UI for TV seasons
func selectSeason(mm *scraper.MediaManager, mediaID string) (int, error) {
	seasons, err := mm.GetTVSeasons(mediaID)
	if err != nil {
		return 0, fmt.Errorf("failed to get seasons: %w", err)
	}

	if len(seasons) == 0 {
		return 0, fmt.Errorf("no seasons found")
	}

	if len(seasons) == 1 {
		return 1, nil
	}

	var items []string
	for _, s := range seasons {
		items = append(items, s.Title)
	}

	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select season: "))
	if err != nil {
		return 0, err
	}

	return idx + 1, nil
}

// selectEpisode presents a selection UI for TV episodes
func selectEpisode(mm *scraper.MediaManager, mediaID string, seasonNum int) (int, error) {
	seasons, err := mm.GetTVSeasons(mediaID)
	if err != nil {
		return 0, fmt.Errorf("failed to get seasons: %w", err)
	}

	if seasonNum > len(seasons) {
		return 0, fmt.Errorf("season %d not found", seasonNum)
	}

	season := seasons[seasonNum-1]
	episodes, err := mm.GetTVEpisodes(season.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get episodes: %w", err)
	}

	if len(episodes) == 0 {
		return 0, fmt.Errorf("no episodes found")
	}

	var items []string
	for _, e := range episodes {
		items = append(items, fmt.Sprintf("Episode %d: %s", e.Number, e.Title))
	}

	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select episode: "))
	if err != nil {
		return 0, err
	}

	return idx + 1, nil
}

// getSeasonEpisodeCount returns the number of episodes in a given season
func getSeasonEpisodeCount(mm *scraper.MediaManager, mediaID string, seasonNum int) (int, error) {
	seasons, err := mm.GetTVSeasons(mediaID)
	if err != nil {
		return 0, fmt.Errorf("failed to get seasons: %w", err)
	}

	if seasonNum > len(seasons) || seasonNum < 1 {
		return 0, fmt.Errorf("season %d not found (have %d seasons)", seasonNum, len(seasons))
	}

	season := seasons[seasonNum-1]
	episodes, err := mm.GetTVEpisodes(season.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get episodes for season %d: %w", seasonNum, err)
	}

	return len(episodes), nil
}

// extractIDFromURL extracts the media ID from a FlixHQ/SFlix URL
func extractIDFromURL(urlStr string) string {
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ExampleMovieDownload demonstrates movie download
func ExampleMovieDownload() {
	// Command: goanime -dm "Inception"
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName: "Inception",
		IsMovie:   true,
		Quality:   "1080",
	}

	if err := HandleMovieDownloadRequest(request); err != nil {
		log.Printf("Movie download failed: %v", err)
	}
}

// ExampleTVDownload demonstrates TV episode download
func ExampleTVDownload() {
	// Command: goanime -dm --type tv "Breaking Bad" 1 1
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:  "Breaking Bad",
		IsTV:       true,
		SeasonNum:  1,
		EpisodeNum: 1,
		Quality:    "1080",
	}

	if err := HandleMovieDownloadRequest(request); err != nil {
		log.Printf("TV download failed: %v", err)
	}
}

// ExampleTVRangeDownload demonstrates TV episode range download
func ExampleTVRangeDownload() {
	// Command: goanime -dm -r "Game of Thrones" 1 1-5
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:    "Game of Thrones",
		IsTV:         true,
		IsRange:      true,
		SeasonNum:    1,
		StartEpisode: 1,
		EndEpisode:   5,
		Quality:      "1080",
	}

	if err := HandleMovieDownloadRequest(request); err != nil {
		log.Printf("TV range download failed: %v", err)
	}
}
