package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvarorichard/Goanime/internal/handlers"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

func main() {
	// FIRST: Parse flags and get user input IMMEDIATELY
	// This ensures UI appears as fast as possible
	animeName, err := util.FlagParser()
	if err != nil {
		// Check if error is update request
		if err == util.ErrUpdateRequested {
			if updateErr := handlers.HandleUpdateRequest(); updateErr != nil {
				log.Fatalln(util.ErrorHandler(updateErr))
			}
			return
		}
		// Check if error is download request
		if err == util.ErrDownloadRequested {
			if downloadErr := handlers.HandleDownloadRequest(); downloadErr != nil {
				log.Fatalln(util.ErrorHandler(downloadErr))
			}
			return
		}
		// Check if error is movie download request (FlixHQ/SFlix)
		if err == util.ErrMovieDownloadRequested {
			if movieDownloadErr := handlers.HandleMovieDownloadRequest(); movieDownloadErr != nil {
				log.Fatalln(util.ErrorHandler(movieDownloadErr))
			}
			return
		}
		// Check if error is upscale request
		if err == util.ErrUpscaleRequested {
			if upscaleErr := handlers.HandleUpscaleRequest(); upscaleErr != nil {
				log.Fatalln(util.ErrorHandler(upscaleErr))
			}
			return
		}
		// For help and version requests, just exit silently
		if err == util.ErrHelpRequested {
			return
		}
		log.Fatalln(util.ErrorHandler(err))
	}

	// Setup signal handling for graceful shutdown (after UI is shown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		util.RunCleanup()
		os.Exit(0)
	}()

	// Ensure cleanup runs on normal exit
	defer util.RunCleanup()

	// Start total execution timer (after UI is shown)
	timer := util.StartTimer("TotalExecution")
	defer timer.Stop()

	// NOW: Initialize heavy stuff AFTER UI is shown
	// Initialize tracker early in background to avoid delays when playing movies
	player.InitTrackerAsync()

	// Pre-warm mpv binary lookup so StartVideo doesn't block on filesystem search
	player.PreWarmMPVPath()

	// Handle normal playback mode
	handlers.HandlePlaybackMode(animeName)
}
