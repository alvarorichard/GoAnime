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

		// Fetch details and episodes.
		// For FlixHQ/movie/TV content, GetAnimeEpisodes may show interactive
		// UI (season selection fuzzyfinder), so we CANNOT run it concurrently
		// with FetchAnimeDetails (which runs a Bubble Tea spinner).  Two
		// programs fighting over the terminal at the same time corrupts
		// terminal state and causes arrow-key escape sequences to be printed
		// as raw text.  For regular anime the episodes fetch is non-interactive,
		// so parallelism is safe there.
		var episodes []models.Episode
		var epErr error

		needsInteractiveEpisodes := anime.Source == "FlixHQ" ||
			anime.MediaType == models.MediaTypeMovie ||
			anime.MediaType == models.MediaTypeTV

		if needsInteractiveEpisodes {
			// Sequential: details first (spinner), then episodes (may show fuzzyfinder)
			parallelTimer := util.StartTimer("FetchDetails+Episodes:Sequential")

			detailsTimer := util.StartTimer("FetchAnimeDetails")
			appflow.FetchAnimeDetails(anime)
			detailsTimer.Stop()

			episodesTimer := util.StartTimer("GetAnimeEpisodes")
			episodes, epErr = appflow.GetAnimeEpisodes(anime)
			if epErr != nil {
				util.Errorf("Failed to get episodes: %v", epErr)
			}
			episodesTimer.Stop()

			parallelTimer.Stop()
		} else {
			// Parallel: safe because GetAnimeEpisodes only uses a spinner (no fuzzyfinder)
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
				episodes, epErr = appflow.GetAnimeEpisodes(anime)
				if epErr != nil {
					util.Errorf("Failed to get episodes: %v", epErr)
				}
				episodesTimer.Stop()
			}()

			wg.Wait()
			parallelTimer.Stop()
		}

		if epErr != nil {
			return
		}

		if len(episodes) == 0 {
			util.Errorf("No episodes found for this anime. Try a different search.")
			return
		}

		util.PerfCount("anime_loaded")

		// Determine if this is a movie or series using the media type first,
		// then fall back to episode count for anime sources that don't set media type.
		totalEpisodes := len(episodes)
		series := !anime.IsMovie() && totalEpisodes > 1
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
