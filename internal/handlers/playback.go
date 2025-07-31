package handlers

import (
	"time"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/playback"
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

	anime := appflow.SearchAnime(animeName)
	appflow.FetchAnimeDetails(anime)
	episodes := appflow.GetAnimeEpisodes(anime)

	util.Debugf("[PERF] Full boot in %v", time.Since(startAll))

	series, totalEpisodes := playback.CheckIfSeries(anime.URL)
	if series {
		playback.HandleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
	} else {
		playback.HandleMovie(anime, episodes, discordManager.IsEnabled())
	}
}
