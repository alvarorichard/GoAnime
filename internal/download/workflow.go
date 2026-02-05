// Package download provides high-level download workflow management
package download

import (
	"fmt"
	"log"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/manifoldco/promptui"
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

	if request.IsRange {
		util.Infof("Downloading episodes %d-%d of %s",
			request.StartEpisode, request.EndEpisode, anime.Name)

		// Exclusive AllAnime Smart Range
		if request.AllAnimeSmart && (anime.Source == "AllAnime" || source == "allanime" || source == "AllAnime") {
			util.Info("AllAnime Smart Range enabled: mirror priority + AniSkip integration + progress UI")
			// Use player batch downloader with provided range to get consistent progress UI
			eps, err := api.GetAnimeEpisodesEnhanced(anime)
			if err == nil && len(eps) > 0 {
				if err := player.HandleBatchDownloadRange(eps, anime.URL, request.StartEpisode, request.EndEpisode); err == nil {
					return nil
				}
				// Fall through to API-based smart range if UI path fails
				util.Infof("Progress UI path failed, falling back to API smart range: %v", err)
			} else if err != nil {
				util.Infof("Enhanced episodes fetch failed for progress path: %v", err)
			}
			if err := api.DownloadAllAnimeSmartRange(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
				util.Errorf("AllAnime Smart Range failed: %v", err)
				// Fallback to normal enhanced
				if err := api.DownloadEpisodeRangeEnhanced(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
					util.Infof("Enhanced download failed, falling back to legacy: %v", err)
					// Fallback to legacy downloader
					episodes := appflow.GetAnimeEpisodesLegacy(anime.URL)
					downloader := downloader.NewEpisodeDownloader(episodes, anime.URL)
					return downloader.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
				}
				return nil
			}
			return nil
		}

		// Try enhanced download first
		if err := api.DownloadEpisodeRangeEnhanced(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
			util.Infof("Enhanced download failed, falling back to legacy: %v", err)
			// Fallback to legacy downloader
			episodes := appflow.GetAnimeEpisodesLegacy(anime.URL)
			downloader := downloader.NewEpisodeDownloader(episodes, anime.URL)
			return downloader.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
		}
		return nil
	} else {
		util.Infof("Downloading episode %d of %s",
			request.EpisodeNum, anime.Name)

		// Enhanced download is a placeholder - use legacy downloader
		util.Infof("Using legacy downloader for episode %d", request.EpisodeNum)
		episodes := appflow.GetAnimeEpisodesLegacy(anime.URL)
		downloader := downloader.NewEpisodeDownloader(episodes, anime.URL)
		return downloader.DownloadSingleEpisode(request.EpisodeNum)
	}
}

// Example usage functions for documentation

// ExampleSingleDownload demonstrates single episode download
func ExampleSingleDownload() {
	// Command: goanime -d "My Hero Academia" 15
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:  "My Hero Academia",
		EpisodeNum: 15,
		IsRange:    false,
	}

	if err := HandleDownloadRequest(request); err != nil {
		log.Printf("Download failed: %v", err)
	}
}

// ExampleRangeDownload demonstrates episode range download
func ExampleRangeDownload() {
	// Command: goanime -d -r "Attack on Titan" 1-5
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:    "Attack on Titan",
		IsRange:      true,
		StartEpisode: 1,
		EndEpisode:   5,
	}

	if err := HandleDownloadRequest(request); err != nil {
		log.Printf("Range download failed: %v", err)
	}
}

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
		seasonNum := request.SeasonNum
		if seasonNum == 0 {
			// Let user select season
			seasonNum, err = selectSeason(mediaManager, extractIDFromURL(selectedMedia.URL))
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
				episodeNum, err = selectEpisode(mediaManager, extractIDFromURL(selectedMedia.URL), seasonNum)
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

	prompt := promptui.Select{
		Label: "Select movie/TV show to download",
		Items: items,
		Size:  15,
	}

	idx, _, err := prompt.Run()
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

	prompt := promptui.Select{
		Label: "Select season",
		Items: items,
	}

	idx, _, err := prompt.Run()
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

	prompt := promptui.Select{
		Label: "Select episode",
		Items: items,
		Size:  15,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return idx + 1, nil
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
