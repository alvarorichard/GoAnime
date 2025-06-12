package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/download"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/updater"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/version"
)

func main() {
	animeName, err := util.FlagParser()
	if err != nil {
		// Check if error is update request
		if err == util.ErrUpdateRequested {
			if updateErr := handleUpdateRequest(); updateErr != nil {
				log.Fatalln(util.ErrorHandler(updateErr))
			}
			return
		}
		// Check if error is download request
		if err == util.ErrDownloadRequested {
			if downloadErr := handleDownloadRequest(); downloadErr != nil {
				log.Fatalln(util.ErrorHandler(downloadErr))
			}
			return
		}
		// For help and version requests, just exit silently
		if err == util.ErrHelpRequested {
			return
		}
		log.Fatalln(util.ErrorHandler(err))
	}

	// Handle normal playback mode
	handlePlaybackMode(animeName)
}

// handleUpdateRequest processes update requests
func handleUpdateRequest() error {
	// Initialize logger for update process
	util.InitLogger()
	util.Info("Checking for updates...")
	if updateErr := updater.CheckAndPromptUpdate(); updateErr != nil {
		return fmt.Errorf("update failed: %w", updateErr)
	}
	return nil
}

// handleDownloadRequest processes download requests
func handleDownloadRequest() error {
	// Initialize logger for download process
	util.InitLogger()

	if util.GlobalDownloadRequest == nil {
		return fmt.Errorf("download request is nil")
	}

	if err := download.HandleDownloadRequest(util.GlobalDownloadRequest); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}

// handlePlaybackMode processes normal anime playback
func handlePlaybackMode(animeName string) {
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
	episodes := appflow.GetAnimeEpisodes(anime.URL)

	util.Debugf("[PERF] Full boot in %v", time.Since(startAll))

	series, totalEpisodes := playback.CheckIfSeries(anime.URL)
	if series {
		playback.HandleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
	} else {
		playback.HandleMovie(anime, episodes, discordManager.IsEnabled())
	}
}
