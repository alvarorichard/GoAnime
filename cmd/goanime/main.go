package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/alvarorichard/Goanime/v2/internal/handlers"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/updater"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"golang.org/x/term"
)

func main() {
	// Exit code for scripts: 0 is success or a deliberate quit (help,
	// version, user cancel), 1 is any genuine error. Registered FIRST so
	// it runs LAST, after the cleanup below.
	exitCode := 0
	defer func() { os.Exit(exitCode) }()
	fail := func(err error) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", util.ErrorHandler(err))
		exitCode = 1
	}

	// A double-click on Windows lands in conhost, which has no font fallback
	// and limited VT support; the TUI would degrade to ASCII. Hand a standalone
	// launch to Windows Terminal (installing it once if needed) and let this
	// process return. Shells and other ConPTY hosts keep their own terminal.
	if tui.RelaunchInWindowsTerminal() {
		return
	}

	// Finish housekeeping from a previous --update: remove the backup and
	// staged files only after this binary has started and its digest matches
	// the one recorded by the updater. No marker means nothing to clean up.
	updater.CleanupAfterUpdate()

	// Enable ANSI/VT processing on Windows consoles (classic cmd.exe leaves it
	// off). Must run before any colored log/TUI output or users see raw escape
	// codes like ←[38;2;...m instead of colors. If enable fails, color paths
	// fall back to ASCII via tui.ConsoleColorProfile / SupportsANSI.
	_ = tui.EnableVirtualTerminal()
	// Decode classic consoles as UTF-8: GoAnime's UTF-8 bytes otherwise meet
	// the console's OEM code page (850/932/…) and Japanese or accented titles
	// arrive as mojibake. The restore hands the old page back on exit.
	restoreConsoleCP := tui.EnableUTF8Console()
	defer restoreConsoleCP()

	// Save terminal state so we can restore it on exit.
	// Libraries like promptui (readline) and go-fuzzyfinder (tcell) put the
	// terminal into raw mode; if the process is interrupted or exits abnormally
	// the terminal can be left in a broken state.
	//
	// restoreTerminal must run as the LAST terminal write on every exit path. On
	// Ctrl+C the huh spinner / Bubble Tea keeps rendering (re-hiding the cursor,
	// re-setting modes) while RunCleanup tears down the browser, so resetting
	// before that teardown lets the spinner re-corrupt the prompt. Running it after
	// RunCleanup — immediately before os.Exit — guarantees ours is the final write.
	restoreTerminal := func() {}
	if fd := os.Stdin.Fd(); fd <= math.MaxInt && term.IsTerminal(int(fd)) {
		intFd := int(fd)
		if origState, err := term.GetState(intFd); err == nil {
			restoreTerminal = func() {
				// Restore the saved raw-mode state, then emit a full set of
				// ANSI/DEC resets so an abnormal exit can't leave the shell prompt
				// broken — shifted by a leftover scroll region/margin, cursor
				// hidden, raw mouse bytes, or garbled colors. See
				// tui.TerminalResetSequence.
				_ = term.Restore(intFd, origState)
				tui.RestoreTerminalStdout()
				// Hand the shell a blank screen: leftover TUI frames (spinner,
				// fuzzyfinder, progress bars) otherwise stay glued around the
				// next prompt. Scrolls the old content into scrollback rather
				// than erasing it, so final messages remain reachable.
				tui.ClearViewportStdout()
			}
		}
	}

	// Catch panics and log them instead of crashing silently. Restore the terminal
	// first so the crash message renders cleanly.
	defer func() {
		if r := recover(); r != nil {
			restoreTerminal()
			stack := debug.Stack()
			util.Errorf("GoAnime crashed: %v\n%s", r, stack)
			fmt.Fprintf(os.Stderr, "\nGoAnime crashed unexpectedly: %v\nStack trace logged to debug log.\n", r)
			exitCode = 1
		}
	}()

	// Setup signal handling for graceful shutdown. Restore the terminal LAST, after
	// RunCleanup, so the spinner/browser teardown can't re-corrupt it afterward.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		util.RunCleanup()
		restoreTerminal()
		os.Exit(130)
	}()

	// On normal return: RunCleanup runs first (registered later → LIFO), then
	// restoreTerminal runs last (registered earlier → LIFO).
	defer restoreTerminal()
	defer util.RunCleanup()

	// Start total execution timer
	timer := util.StartTimer("TotalExecution")
	defer timer.Stop()

	// Initialize tracker early in background to avoid delays when playing movies
	player.InitTrackerAsync()

	// Pre-warm mpv binary lookup so StartVideo doesn't block on filesystem search
	player.PreWarmMPVPath()

	// Pre-initialize HTTP clients in background so the first search doesn't pay
	// the Chrome TLS setup cost. The per-source scraper adapters are cheap, lazy
	// structs the Model B providers build on first use — no pre-warm needed.
	util.PreWarmClients()
	util.PreWarmConnections()

	animeName, err := util.FlagParser()
	if err != nil {
		// Check if error is update request
		if err == util.ErrUpdateRequested {
			if updateErr := handlers.HandleUpdateRequest(); updateErr != nil {
				fail(updateErr)
			}
			return

		}
		// Check if error is download request
		if err == util.ErrDownloadRequested {
			if downloadErr := handlers.HandleDownloadRequest(); downloadErr != nil {
				fail(downloadErr)
			}
			return
		}
		// Check if error is movie download request (FlixHQ/SFlix)
		if err == util.ErrMovieDownloadRequested {
			// Movie/TV download is currently a permanent stub that always
			// returns an explanatory error (scrapers removed) — log it
			// unconditionally.
			fail(handlers.HandleMovieDownloadRequest())
			return
		}
		// Check if error is upscale request
		if err == util.ErrUpscaleRequested {
			if upscaleErr := handlers.HandleUpscaleRequest(); upscaleErr != nil {
				fail(upscaleErr)
			}
			return
		}
		// For help and version requests, just exit silently
		if err == util.ErrHelpRequested {
			return
		}
		fail(err)
		return
	}

	// Guarantee the Japanese font decision on every real launch. The probes
	// run once (sync.Once): install GoAnime Glyphs when no Japanese face is
	// accounted for, and log the note. No handler is trusted to be the only
	// possible path — the search home, a named search and the download and
	// update flags all ride through here. Plain mode does not probe, and
	// help/version already returned above, so a terminal-less or scripted
	// run pays nothing.
	tui.EnsureJapaneseDetected()

	// Handle normal playback mode: an empty name opens the SEARCH home
	// screen (tela 1) instead of searching blank.
	if animeName == "" {
		decision, homeErr := handlers.HandleSearchHome()
		if homeErr != nil {
			fail(homeErr)
			return
		}
		if decision.Cancelled {
			return
		}
		if decision.Entry != nil {
			handlers.HandleContinue(decision)
			return
		}
		if decision.Scope != "" {
			util.GlobalSource = decision.Scope
		}
		animeName = util.TreatingAnimeName(decision.Query)
	}
	handlers.HandlePlaybackMode(animeName)
}
