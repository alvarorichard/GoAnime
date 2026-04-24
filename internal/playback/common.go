package playback

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// PlayEpisode loads and starts playback for a single episode.
func PlayEpisode(
	anime *models.Anime,
	episodes []models.Episode,
	episodeNum int,
	episodeURL string,
	episodeNumberStr string,
	discordEnabled bool,
	isPaused *bool,
	animeMutex *sync.Mutex,
) error {
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
		episodeURLForCreation := episodeURL
		if api.IsAllAnimeSource(anime) {
			episodeURLForCreation = anime.URL
		}

		currentEpisode = &models.Episode{
			Number: episodeNumberStr,
			Num:    episodeNum,
			URL:    episodeURLForCreation,
		}
	}

	// Fetch episode metadata and stream URL in parallel under a single spinner
	// GetEpisodeData (Jikan/AniList metadata) and GetVideoURLForEpisodeEnhanced (scraper)
	// are independent operations — running them concurrently saves a full round-trip
	var videoURL string
	var videoErr error
	currentEpisodeCopy := currentEpisode // capture for goroutine

	tui.RunWithSpinner("Loading episode...", func() {
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := api.GetEpisodeData(anime.MalID, episodeNum, anime); err != nil {
				util.Debugf("Error fetching episode data: %v", err)
			}
		}()

		go func() {
			defer wg.Done()
			videoURL, videoErr = player.GetVideoURLForEpisodeEnhanced(currentEpisodeCopy, anime)
		}()

		wg.Wait()
	})

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

	playErr := player.HandleDownloadAndPlay(
		videoURL,
		episodes,
		episodeNum,
		anime.URL,
		episodeNumberStr,
		anime.MalID,
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

// SelectEpisodeWithFuzzy presents a fuzzy-finder UI and returns the chosen episode URL, number string, and number.
func SelectEpisodeWithFuzzy(episodes []models.Episode) (string, string, int, error) {
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

// FindEpisodeByNumber returns the URL, number string, and number for the episode matching num.
func FindEpisodeByNumber(episodes []models.Episode, num int) (string, string, int, error) {
	for _, ep := range episodes {
		if epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(ep.Number)); err == nil && epNum == num {
			return ep.URL, ep.Number, num, nil
		}
	}
	log.Printf("Warning: Episode number %d not found. Re-selecting.", num)
	return SelectEpisodeWithFuzzy(episodes)
}
