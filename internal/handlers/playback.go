package handlers

import (
	"errors"
	"time"

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
	startAll := time.Now()

	// Initialize the beautiful logger
	util.InitLogger()

	tracking.HandleTrackingNotice()
	util.Debugf("[PERF] starting Goanime v%s", version.Version)

	discordManager := discord.NewManager()
	if err := discordManager.Initialize(); err != nil {
		util.Debug("Failed to initialize Discord Rich Presence:", "error", err)
	} else {
		defer discordManager.Shutdown()
	}

	currentAnimeName := animeName

	for {
		// Use enhanced search with retry logic
		anime, err := appflow.SearchAnimeWithRetry(currentAnimeName)
		if err != nil {
			util.Errorf("Failed to search for anime: %v", err)
			return
		}

		appflow.FetchAnimeDetails(anime)
		episodes := appflow.GetAnimeEpisodes(anime)

		util.Debugf("[PERF] Full boot in %v", time.Since(startAll))

		series, totalEpisodes := playback.CheckIfSeriesEnhanced(anime)
		var playbackErr error
		if series {
			playbackErr = playback.HandleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
		} else {
			playbackErr = playback.HandleMovie(anime, episodes, discordManager.IsEnabled())
		}

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
