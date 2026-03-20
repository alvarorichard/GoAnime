package handlers

import (
	"errors"
	"sync"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/version"
)

// HandlePlaybackMode processes normal anime playback
func HandlePlaybackMode(animeName string) {
	timer := util.StartTimer("PlaybackMode:Total")
	defer timer.Stop()

	// Initialize the beautiful logger
	util.InitLogger()

	// Pre-warm connections are now started in main() so they run while the
	// user is still typing the anime name. This call is a noop (sync.Once).
	util.PreWarmConnections()

	tracking.HandleTrackingNotice()
	util.Debugf("[PERF] starting Goanime v%s", version.Version)

	// Discord init runs in background - doesn't block startup
	discordManager := discord.NewManager()
	_ = discordManager.Initialize() // Non-blocking, runs async
	defer discordManager.Shutdown()

	currentAnimeName := animeName

	for {
		// Use enhanced search with retry logic
		searchTimer := util.StartTimer("SearchAnime:WithRetry")
		anime, err := appflow.SearchAnimeWithRetry(currentAnimeName)
		searchTimer.Stop()

		if err != nil {
			util.Errorf("Failed to search for anime: %v", err)
			return
		}

		// Fetch details and episodes in parallel — they are independent
		// Details come from AniList/TMDB, episodes from the source scraper
		var episodes []models.Episode
		var wg sync.WaitGroup

		parallelTimer := util.StartTimer("FetchDetails+Episodes:Parallel")

		wg.Add(2)
		go func() {
			defer wg.Done()
			detailsTimer := util.StartTimer("FetchAnimeDetails")
			appflow.FetchAnimeDetails(anime)
			detailsTimer.Stop()
		}()
		go func() {
			defer wg.Done()
			episodesTimer := util.StartTimer("GetAnimeEpisodes")
			var epErr error
			episodes, epErr = appflow.GetAnimeEpisodes(anime)
			if epErr != nil {
				util.Errorf("Failed to get episodes: %v", epErr)
			}
			episodesTimer.Stop()
		}()

		wg.Wait()
		parallelTimer.Stop()

		if len(episodes) == 0 {
			util.Errorf("No episodes found for this anime. Try a different search.")
			return
		}

		util.PerfCount("anime_loaded")

		// Use length of already-fetched episodes to determine if it's a series
		// This avoids re-fetching episodes which would cause duplicate season selection for FlixHQ
		totalEpisodes := len(episodes)
		series := totalEpisodes > 1
		var playbackErr error

		playbackTimer := util.StartTimer("Playback:Handle")
		if series {
			playbackErr = playback.HandleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
		} else {
			playbackErr = playback.HandleMovie(anime, episodes, discordManager.IsEnabled())
		}
		playbackTimer.Stop()

		// Check if user wants to go back to anime selection
		if errors.Is(playbackErr, player.ErrBackToAnimeSelection) {
			util.Infof("Going back to anime selection...")
			// Keep the same search term to show the anime list again
			continue
		}

		// Normal exit or other errors
		break
	}
}
