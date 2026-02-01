package handlers

import (
	"errors"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
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

	tracking.HandleTrackingNotice()
	util.Debugf("[PERF] starting Goanime v%s", version.Version)

	discordTimer := util.StartTimer("Discord:Initialize")
	discordManager := discord.NewManager()
	if err := discordManager.Initialize(); err != nil {
		util.Debug("Failed to initialize Discord Rich Presence:", "error", err)
	} else {
		defer discordManager.Shutdown()
	}
	discordTimer.Stop()

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

		detailsTimer := util.StartTimer("FetchAnimeDetails")
		appflow.FetchAnimeDetails(anime)
		detailsTimer.Stop()

		episodesTimer := util.StartTimer("GetAnimeEpisodes")
		episodes := appflow.GetAnimeEpisodes(anime)
		episodesTimer.Stop()

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
