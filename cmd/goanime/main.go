package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/alvarorichard/Goanime/internal/handlers"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"golang.org/x/term"
)

func main() {
	// Save terminal state so we can restore it on exit.
	// Libraries like promptui (readline) and go-fuzzyfinder (tcell) put the
	// terminal into raw mode; if the process is interrupted or exits abnormally
	// the terminal can be left in a broken state.
	if fd := os.Stdin.Fd(); fd <= math.MaxInt && term.IsTerminal(int(fd)) {
		intFd := int(fd)
		if origState, err := term.GetState(intFd); err == nil {
			util.RegisterCleanup(func() {
				_ = term.Restore(intFd, origState)
				// Ensure the terminal is in a clean state for the shell:
				// - \033[0m  resets all ANSI attributes (bold, color, etc.)
				// - \033[?25h re-shows the cursor (spinners may hide it)
				// - \r\n moves cursor to column 0 on a fresh line
				fmt.Fprint(os.Stdout, "\033[0m\033[?25h\r\n")
			})
		}
	}

	// Catch panics and log them instead of crashing silently
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			util.Errorf("GoAnime crashed: %v\n%s", r, stack)
			fmt.Fprintf(os.Stderr, "\nGoAnime crashed unexpectedly: %v\nStack trace logged to debug log.\n", r)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		util.RunCleanup()
		os.Exit(0)
	}()

	// Ensure cleanup runs on normal exit
	defer util.RunCleanup()

	// Start total execution timer
	timer := util.StartTimer("TotalExecution")
	defer timer.Stop()

	// Initialize tracker early in background to avoid delays when playing movies
	player.InitTrackerAsync()

	// Pre-warm mpv binary lookup so StartVideo doesn't block on filesystem search
	player.PreWarmMPVPath()

	// Pre-initialize HTTP clients and scraper manager in background so the
	// first search doesn't pay the Chrome TLS + scraper setup cost
	util.PreWarmClients()
	util.PreWarmConnections()
	scraper.PreWarmScraperManager()

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

	// Handle normal playback mode
	handlers.HandlePlaybackMode(animeName)
}
