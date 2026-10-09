// Package download provides high-level download workflow management
package download

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

// workflowSearchFn is the anime search function used by HandleDownloadRequest.
// Tests may override it to avoid spawning a real TUI search.
var workflowSearchFn = appflow.SearchAnimeWithRetry

// workflowEnrichFn resolves the AniList season mapping for the found anime.
// Tests may override it to avoid the real AniList request.
var workflowEnrichFn = func(ctx context.Context, anime *models.Anime) ([]metadata.SeasonMapping, error) {
	return metadata.NewEnricher().EnrichAnime(ctx, anime)
}

// Episode listing and the player's batch downloads; seams for tests.
var (
	workflowFetchEpisodesFn  = providers.FetchEpisodes
	workflowLegacyEpisodesFn = appflow.GetAnimeEpisodesLegacy
	workflowDownloadAllFn    = player.HandleDownloadAll
	workflowBatchRangeFn     = player.HandleBatchDownloadRange
)

// The movie/TV sources (StartFlix and TopCine, searched together) for the -dm
// downloads; seams for tests. TopCine titles list through the same panel flow,
// so the StartFlix listing functions serve both.
var (
	movieSearchFn = func(name string) (*models.Anime, error) {
		return api.SearchAnimeEnhanced(name, api.MovieTVSources)
	}
	seriesSeasonsFn  = api.GetStartFlixSeasonNumbers
	seriesEpisodesFn = api.GetStartFlixSeasonEpisodes
)

// fileUnderSeason points the player's download paths at the season being
// downloaded and refreshes the naming metadata. Called after listing, which
// is where StartFlix's season picker records the season the user chose.
func fileUnderSeason(anime *models.Anime, season int) {
	if anime.CurrentSeason > 0 {
		season = anime.CurrentSeason
	}
	player.SetAnimeName(anime.Name, max(season, 1))
	player.SetExactMediaType(string(anime.MediaType))
	setMediaMeta(anime) // listing can reveal ids and the official title (StartFlix)
}

// quitIsDone maps the batch's "finished, do not prompt further" signal to
// success.
func quitIsDone(err error) error {
	if errors.Is(err, player.ErrUserQuit) {
		return nil
	}
	return err
}

// setMediaMeta hands the player what download paths are named from: the
// official title, year and external ids the anime carries right now.
func setMediaMeta(anime *models.Anime) {
	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})
}

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

	setMediaMeta(anime)

	seasonMap, _ := workflowEnrichFn(context.Background(), anime)
	player.SetSeasonMap(seasonMap)

	setMediaMeta(anime)

	if request.IsAll {
		util.Infof("Downloading ALL episodes of %s", anime.Name)
		eps, err := workflowFetchEpisodesFn(context.Background(), anime)
		if err == nil && len(eps) > 0 {
			fileUnderSeason(anime, season)
			// Every listed episode, without asking for a range: "-d -a" used
			// to open HandleBatchDownload's start/end form.
			dlErr := workflowDownloadAllFn(eps, anime)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}

		episodes, legacyErr := workflowLegacyEpisodesFn(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadAllEpisodes()
	}

	if request.IsRange {
		util.Infof("Downloading episodes %d-%d of %s",
			request.StartEpisode, request.EndEpisode, anime.Name)

		eps, err := workflowFetchEpisodesFn(context.Background(), anime)
		if err == nil && len(eps) > 0 {
			fileUnderSeason(anime, season)
			dlErr := workflowBatchRangeFn(eps, anime, request.StartEpisode, request.EndEpisode)
			if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
				return nil
			}
			util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
		} else if err != nil {
			util.Infof("Enhanced episodes fetch failed: %v", err)
		}
		episodes, legacyErr := workflowLegacyEpisodesFn(anime.URL)
		if legacyErr != nil {
			return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
		}
		dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
		return dl.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
	}

	util.Infof("Downloading episode %d of %s", request.EpisodeNum, anime.Name)
	// The source's own listing first, like the range and all-episodes paths:
	// the legacy one only reads AnimeFire pages, so a StartFlix episode could
	// not be downloaded on its own.
	eps, err := workflowFetchEpisodesFn(context.Background(), anime)
	if err == nil && len(eps) > 0 {
		fileUnderSeason(anime, season)
		dlErr := workflowBatchRangeFn(eps, anime, request.EpisodeNum, request.EpisodeNum)
		if dlErr == nil || errors.Is(dlErr, player.ErrUserQuit) {
			return nil
		}
		util.Infof("Batch download path failed, falling back to legacy: %v", dlErr)
	} else if err != nil {
		util.Infof("Enhanced episodes fetch failed: %v", err)
	}
	episodes, legacyErr := workflowLegacyEpisodesFn(anime.URL)
	if legacyErr != nil {
		return fmt.Errorf("failed to fetch episodes: %w", legacyErr)
	}
	dl := downloader.NewEpisodeDownloaderWithAnime(episodes, anime.URL, anime)
	return dl.DownloadSingleEpisode(request.EpisodeNum)
}

// HandleMovieDownloadRequest downloads movies and series from the movie/TV
// sources, StartFlix and TopCine. The -dm forms:
//
//	goanime -dm "Movie"                     the movie
//	goanime -dm --type tv "Show" 2 5        season 2, episode 5
//	goanime -dm -r "Show" 2 1-5             season 2, episodes 1-5
//	goanime -dm -a "Show"                   every season, every episode
//
// Seasons are taken from the command, not picked; the audio (Dublado or
// Legendado) is asked once per title when a season offers both.
func HandleMovieDownloadRequest(request *util.DownloadRequest) error {
	if request == nil {
		return errors.New("download request is nil")
	}
	if strings.TrimSpace(request.AnimeName) == "" {
		return errors.New("no movie or series name to download")
	}
	title, err := movieSearchFn(request.AnimeName)
	if err != nil {
		return fmt.Errorf("failed to search the movie/TV sources: %w", err)
	}
	player.SetSeasonMap(nil) // the panel numbers episodes per season

	if !request.IsTV {
		return downloadStartFlixMovie(title)
	}
	if request.IsAll {
		return downloadStartFlixSeries(title)
	}
	start, end := request.EpisodeNum, request.EpisodeNum
	if request.IsRange {
		start, end = request.StartEpisode, request.EndEpisode
	}
	eps, err := seriesEpisodesFn(title, request.SeasonNum)
	if err != nil {
		return seriesErr(title, err)
	}
	util.Infof("Downloading %s season %d, episode(s) %d-%d", title.Name, request.SeasonNum, start, end)
	fileUnderSeason(title, request.SeasonNum)
	return quitIsDone(workflowBatchRangeFn(eps, title, start, end))
}

func downloadStartFlixMovie(title *models.Anime) error {
	eps, err := workflowFetchEpisodesFn(context.Background(), title)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", title.Name, err)
	}
	if title.MediaType != models.MediaTypeMovie {
		return fmt.Errorf("%q is a series on %s: download it with -dm --type tv \"%s\" <season> <episode>, -dm -r or -dm -a", title.Name, sourceOr(title), title.Name)
	}
	if len(eps) == 0 {
		return fmt.Errorf("%q lists nothing to download", title.Name)
	}
	util.Infof("Downloading %s", title.Name)
	fileUnderSeason(title, 1)
	return quitIsDone(workflowBatchRangeFn(eps, title, eps[0].Num, eps[0].Num))
}

// downloadStartFlixSeries downloads every season, one after another. A season
// with failed episodes does not stop the others; the failures are reported
// together at the end.
func downloadStartFlixSeries(title *models.Anime) error {
	seasons, err := seriesSeasonsFn(title)
	if err != nil {
		return seriesErr(title, err)
	}
	var failed []error
	for _, n := range seasons {
		eps, err := seriesEpisodesFn(title, n)
		if err != nil {
			failed = append(failed, fmt.Errorf("season %d: %w", n, err))
			continue
		}
		util.Infof("Downloading %s season %d (%d episodes)", title.Name, n, len(eps))
		fileUnderSeason(title, n)
		if err := quitIsDone(workflowDownloadAllFn(eps, title)); err != nil {
			failed = append(failed, fmt.Errorf("season %d: %w", n, err))
		}
	}
	return errors.Join(failed...)
}

func seriesErr(title *models.Anime, err error) error {
	if errors.Is(err, api.ErrStartFlixNotSeries) {
		return fmt.Errorf("%q is a movie on %s: download it with -dm \"%s\"", title.Name, sourceOr(title), title.Name)
	}
	return fmt.Errorf("failed to list %q: %w", title.Name, err)
}

// sourceOr names the source a title came from, for messages.
func sourceOr(title *models.Anime) string {
	if title.Source != "" {
		return title.Source
	}
	return "StartFlix"
}
