package playback

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

func PlayEpisode(
	ctx context.Context,
	anime *models.Anime,
	episodes []models.Episode,
	episodeNum int,
	episodeURL string,
	episodeNumberStr string,
	discordEnabled bool,
	isPaused *bool,
	animeMutex *sync.Mutex,
) error {
	// The executable lookup is independent from stream scraping.  Start it now
	// so the final handoff to mpv never pays filesystem/PATH probing latency.
	player.PreWarmMPVPath()

	animeMutex.Lock()
	anime.Episodes = []models.Episode{{
		Number: episodeNumberStr,
		Num:    episodeNum,
		URL:    episodeURL,
	}}
	animeMutex.Unlock()

	// Find the specific episode to pass to enhanced API (pure sync, no network)
	var currentEpisode *models.Episode
	util.Debug("PlayEpisode searching for episode", "episodeNumberStr", episodeNumberStr, "totalEpisodes", len(episodes))
	for i := range episodes {
		util.Debug("Checking episode", "index", i, "epNumber", episodes[i].Number, "epURL", episodes[i].URL)
		if episodes[i].Number == episodeNumberStr {
			currentEpisode = &episodes[i]
			util.Debug("Found matching episode", "URL", currentEpisode.URL, "DataID", currentEpisode.DataID)
			break
		}
	}

	if currentEpisode == nil {
		// Create episode if not found
		// Sources addressed by a bare id (rather than an episode page URL) key
		// their tracking rows off the anime id.
		episodeURLForCreation := episodeURL
		if len(anime.URL) < 30 && !strings.Contains(anime.URL, "http") && !strings.Contains(anime.URL, "animesdrive") {
			episodeURLForCreation = anime.URL
		}

		currentEpisode = &models.Episode{
			Number: episodeNumberStr,
			Num:    episodeNum,
			URL:    episodeURLForCreation,
		}
	}

	// Fetch episode metadata and stream URL in parallel.
	//
	// 2026-04-28: removed the huh/v2 Bubble Tea spinner that previously
	// wrapped this block. GetVideoURLForEpisodeEnhanced may invoke a
	// tcell-based fuzzyfinder quality picker (AnimeFire's multi-quality
	// response). The Bubble Tea spinner and tcell racing for stdin/stdout
	// caused two user-visible bugs: arrow keys needed multiple presses to
	// register (input contention) and the spinner's redraw clipped the
	// first character of the picker's prompt ("S" of "Select"). A static
	// log line is the smaller evil — animation is a nice-to-have, the
	// picker working is not.
	util.Infof("Loading episode...")

	var videoURL string
	var videoErr error
	var episodeDataErr error
	currentEpisodeCopy := currentEpisode
	episodeDataAnime := *anime
	episodeDataAnime.Episodes = append([]models.Episode(nil), anime.Episodes...)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// Metadata providers mutate Episodes. Keep that mutation isolated from
		// stream resolution, which concurrently reads and may enrich anime.
		episodeDataErr = api.GetEpisodeData(anime.MalID, episodeNum, &episodeDataAnime)
		if episodeDataErr != nil {
			util.Debugf("Error fetching episode data: %v", episodeDataErr)
		}
	}()

	go func() {
		defer wg.Done()
		videoURL, videoErr = player.GetVideoURLForEpisodeEnhanced(ctx, currentEpisodeCopy, anime)
	}()

	wg.Wait()
	if episodeDataErr == nil {
		animeMutex.Lock()
		anime.Episodes = append([]models.Episode(nil), episodeDataAnime.Episodes...)
		animeMutex.Unlock()
	}

	if videoErr != nil {
		// Any video URL failure means the episode is not available on this source.
		// Route user back to episode selection so they can pick another one.
		if !errors.Is(videoErr, player.ErrBackToEpisodeSelection) {
			util.Warnf("Failed to extract video URL: %v", videoErr)
		}
		return player.ErrBackToEpisodeSelection
	}

	// Guard against empty or missing durations
	var episodeDuration time.Duration
	if len(episodes) > 0 && episodes[0].Duration > 0 {
		episodeDuration = time.Duration(episodes[0].Duration) * time.Second
	} else {
		episodeDuration = 0
	}
	updater := createUpdater(anime, isPaused, animeMutex, episodeDuration, discordEnabled)

	// Route downloads to the correct directory (anime/ vs movies/) using exact media type
	player.SetExactMediaType(string(anime.MediaType))

	// Store external IDs for Plex/Jellyfin-compatible folder naming
	player.SetMediaMeta(&util.MediaMeta{
		OfficialTitle: anime.OfficialTitle(),
		Year:          anime.Year,
		TMDBID:        anime.TMDBID,
		IMDBID:        anime.IMDBID,
		AnilistID:     anime.AnilistID,
		MalID:         anime.MalID,
	})

	// Enrich anime with AniList metadata for per-episode season resolution.
	// This populates the season map so episodes like Black Clover ep 52 go to
	// Season 02 instead of Season 01.
	enricher := metadata.NewEnricher()
	seasonMap, _ := enricher.EnrichAnime(ctx, anime)
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

	playErr := player.HandleDownloadAndPlay(
		videoURL,
		episodes,
		episodeNum,
		anime.URL,
		episodeNumberStr,
		anime.MalID,
		anime.AnilistID,
		updater,
		anime.Name,
		anime.CurrentSeason,
		anime,
	)

	if updater != nil {
		updater.Stop()
	}
	return playErr
}

func SelectEpisodeWithFuzzy(episodes []models.Episode) (episodeURL, episodeNumber string, episodeIndex int, err error) {
	url, numStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		// If user selected back, return empty values to signal back request
		if errors.Is(err, player.ErrBackRequested) {
			return "", "back", -1, nil
		}
		return "", "", 0, fmt.Errorf("episode selection failed: %w", err)
	}
	epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(numStr))
	if err != nil {
		return "", "", 0, fmt.Errorf("error converting episode number: %w", err)
	}
	return url, numStr, epNum, nil
}

func FindEpisodeByNumber(episodes []models.Episode, num int) (episodeURL, episodeNumber string, episodeIndex int, err error) {
	for _, ep := range episodes {
		if epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(ep.Number)); err == nil && epNum == num {
			return ep.URL, ep.Number, num, nil
		}
	}
	log.Printf("Warning: Episode number %d not found. Re-selecting.", num)
	return SelectEpisodeWithFuzzy(episodes)
}
