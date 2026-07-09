// Package download provides high-level download workflow management
package download

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

// workflowSearchFn is the anime search function used by HandleDownloadRequest.
// Tests may override it to avoid spawning a real TUI search.
var workflowSearchFn = appflow.SearchAnimeWithRetry

// HandleDownloadRequest processes a download request from command line
func HandleDownloadRequest(request *util.DownloadRequest) error {
	util.Info("Starting enhanced download mode...")

	source := request.Source
	quality := request.Quality
	if quality == "" {
		quality = "best"
	}

	util.Infof("Using source: %s, quality: %s", source, quality)

	anime, err := workflowSearchFn(request.AnimeName)
	if err != nil {
		util.Errorf("Failed to search for anime: %v", err)
		return err
	}

	season := 1
	if request.SeasonNum > 0 {
		season = request.SeasonNum
	}
	player.SetAnimeName(anime.Name, season)
	player.SetExactMediaType(string(anime.MediaType))

	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	enricher := metadata.NewEnricher()
	seasonMap, _ := enricher.EnrichAnime(context.Background(), anime)
	player.SetSeasonMap(seasonMap)

	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	if request.IsAll {
		util.Infof("Downloading ALL episodes of %s", anime.Name)
		eps, err := providers.FetchEpisodes(context.Background(), anime)
		if err == nil && len(eps) > 0 {
			dlErr := player.HandleBatchDownload(eps, anime)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}

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

		if request.AllAnimeSmart && (anime.Source == "AllAnime" || source == "allanime" || source == "AllAnime") {
			util.Info("AllAnime Smart Range enabled: mirror priority + AniSkip integration + progress UI")
			eps, err := providers.FetchEpisodes(context.Background(), anime)
			if err == nil && len(eps) > 0 {
				dlErr := player.HandleBatchDownloadRange(eps, anime, request.StartEpisode, request.EndEpisode)
				if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
					return nil
				}
				util.Infof("Progress UI path failed, falling back to API smart range: %v", dlErr)
			} else if err != nil {
				util.Infof("Enhanced episodes fetch failed for progress path: %v", err)
			}
			if err := api.DownloadAllAnimeSmartRange(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
				util.Errorf("AllAnime Smart Range failed: %v", err)
				if err := api.DownloadEpisodeRangeEnhanced(anime, request.StartEpisode, request.EndEpisode, quality); err != nil {
					util.Infof("Enhanced download failed, falling back to legacy: %v", err)
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

		eps, err := providers.FetchEpisodes(context.Background(), anime)
		if err == nil && len(eps) > 0 {
			dlErr := player.HandleBatchDownloadRange(eps, anime, request.StartEpisode, request.EndEpisode)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}
		episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
	}

	util.Infof("Downloading episode %d of %s", request.EpisodeNum, anime.Name)
	episodes, legacyErr := appflow.GetAnimeEpisodesLegacy(anime.URL)
	if legacyErr != nil {
		return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
	}
	dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
	return dl.DownloadSingleEpisode(request.EpisodeNum)
}

// HandleMovieDownloadRequest is a stub kept for callers — movie/TV scrapers
// have all been removed (SFlix/FlixHQ deleted because they went offline).
// Always returns an error.
func HandleMovieDownloadRequest(request *util.DownloadRequest) error {
	_ = request
	return fmt.Errorf("movie/TV download is no longer supported: scrapers have been removed")
}
