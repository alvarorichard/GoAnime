package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/version"
)

// HandlePlaybackMode processes normal anime playback
func HandlePlaybackMode(animeName string) {
	timer := util.StartTimer("PlaybackMode:Total")
	defer timer.Stop()

	// Root context for the playback session. Today it is Background; once the
	// dispatch path honors ctx end to end, this becomes the single place to
	// hook signal-aware cancellation (signal.NotifyContext).
	ctx := context.Background()

	// Initialize the beautiful logger
	util.InitLogger()

	// Confirm the manual kill-switch (S1) visibly: if the user disabled any
	// source via GOANIME_DISABLED_SOURCES, say so once at startup so a turned-
	// off source is never a silent surprise (R5).
	if disabled := source.DisabledSources(); len(disabled) > 0 {
		names := make([]string, len(disabled))
		for i, k := range disabled {
			names[i] = string(k)
		}
		util.Warnf("Sources disabled by config (GOANIME_DISABLED_SOURCES): %s", strings.Join(names, ", "))
	}

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
			// The user quit the result screen or the search prompt. That is an
			// exit, not a failure, so leave without an error banner.
			if errors.Is(err, api.ErrSearchAborted) {
				util.Infof("Search cancelled.")
				return
			}
			util.Errorf("Failed to search for anime: %v", err)
			return
		}

		// Fetch details before episodes. Both operations access the same mutable
		// Media, and some episode flows also open an interactive picker, so
		// concurrent execution would race in memory and contend for the terminal.
		var episodes []models.Episode
		var epErr error

		fetchTimer := util.StartTimer("FetchDetails+Episodes:Sequential")
		detailsTimer := util.StartTimer("FetchAnimeDetails")
		appflow.FetchAnimeDetails(anime)
		detailsTimer.Stop()

		episodesTimer := util.StartTimer("GetAnimeEpisodes")
		episodes, epErr = appflow.GetAnimeEpisodes(anime)
		episodesTimer.Stop()
		fetchTimer.Stop()

		// User aborted season selection (FlixHQ/StartFlix ESC) — go back to a
		// fresh search prompt instead of killing the session.
		if errors.Is(epErr, api.ErrBackToSearch) {
			util.Infof("Going back to new search...")
			currentAnimeName = ""
			continue
		}

		// A title that cannot be opened sends the user back to the results
		// instead of ending the session: another result, often the same title
		// on another source, may well play. Ending here used to look like a
		// silent crash, because the terminal is cleared on exit.
		if epErr != nil || len(episodes) == 0 {
			reportEpisodesUnavailable(anime, epErr)
			util.Infof("Going back to anime selection...")
			continue
		}

		util.PerfCount("anime_loaded")

		// Determine if this is a movie or series using the media type first,
		// then fall back to episode count for anime sources that don't set media type.
		totalEpisodes := len(episodes)
		series := !anime.IsMovie() && totalEpisodes > 1
		var playbackErr error

		playbackTimer := util.StartTimer("Playback:Handle")
		if series {
			playbackErr = playback.HandleSeries(ctx, anime, episodes, totalEpisodes, discordManager.IsEnabled())
		} else {
			playbackErr = playback.HandleMovie(ctx, anime, episodes, discordManager.IsEnabled())
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

// reportEpisodesUnavailable says why a title could not be opened, both in the
// terminal and on the result screen the user is sent back to, where anything
// printed before it opens would be hidden.
func reportEpisodesUnavailable(anime *models.Anime, err error) {
	notice, failed := episodesUnavailableNotice(anime, err)
	if failed {
		util.Errorf("Failed to get episodes: %v", err)
	} else {
		util.Warn(notice)
	}
	tui.NoticeOnNextResults(notice)
}

// episodesUnavailableNotice words the result-screen notice for a title whose
// episodes could not be listed. failed is false when the source simply has no
// episodes for it, such as an announced season that has not aired: that is
// content state, not an error.
func episodesUnavailableNotice(anime *models.Anime, err error) (notice string, failed bool) {
	name := tui.SingleLine(anime.Name)
	where := ""
	if src := tui.SingleLine(anime.Source); src != "" {
		where = " on " + src
	}
	if err == nil || errors.Is(err, netx.ErrNoEpisodes) {
		return fmt.Sprintf("%s has no episodes%s yet.", name, where), false
	}
	return fmt.Sprintf("Could not load the episodes of %s%s. Try another result.", name, where), true
}
