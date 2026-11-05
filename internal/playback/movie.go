package playback

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/appflow"
	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// HandleMovie gerencia a reprodução de filmes/OVAs
func HandleMovie(ctx context.Context, anime *models.Anime, episodes []models.Episode, discordEnabled bool) error {
	return runMovie(realFlow(), ctx, anime, episodes, discordEnabled)
}

// runMovie is HandleMovie's loop over the environment env provides.
func runMovie(env flowEnv, ctx context.Context, anime *models.Anime, episodes []models.Episode, discordEnabled bool) error { //nolint:revive // env first keeps call sites readable next to the real wiring
	// Prepare the mpv path while metadata/stream work is in progress.
	env.preWarmMPV()
	env.setEpisodeListCtx(anime.Name, anime.Source, anime.Audio)

	for {
		animeMutex := sync.Mutex{}
		isPaused := false

		animeMutex.Lock()
		anime.Episodes = []models.Episode{episodes[0]}
		animeMutex.Unlock()

		movieData := fetchMovieData(anime)

		// Use static log instead of spinner while fetching video URL
		// to avoid terminal UI contention if a quality picker opens.
		util.Infof("Loading video stream...")
		videoURL, videoErr := env.streamURL(ctx, &episodes[0], anime)

		// Keep the metadata only if it is already here; playback never waits
		// for it.
		select {
		case eps := <-movieData:
			animeMutex.Lock()
			anime.Episodes = eps
			animeMutex.Unlock()
		default:
		}

		if videoErr != nil {
			// The anime-side scheme: say why this copy failed, name the other
			// sources that list the title, and let the next result screen
			// open with the notice above the list.
			env.reportStreamFailed(anime, videoErr)
			// Return to anime selection
			return player.ErrBackToAnimeSelection
		}

		episodeDuration := time.Duration(episodes[0].Duration) * time.Second
		updater := env.newUpdater(anime, &isPaused, &animeMutex, episodeDuration, discordEnabled)

		// Route downloads to the correct directory (anime/ vs movies/) using exact media type
		env.setExactMediaType(string(anime.MediaType))

		// Store external IDs for Plex/Jellyfin-compatible folder naming
		env.setMediaMeta(&util.MediaMeta{
			OfficialTitle: anime.OfficialTitle(),
			Year:          anime.Year,
			TMDBID:        anime.TMDBID,
			IMDBID:        anime.IMDBID,
			AnilistID:     anime.AnilistID,
			MalID:         anime.MalID,
		})

		playErr := env.downloadAndPlay(
			videoURL,
			episodes,
			1,
			anime.URL,
			episodes[0].Number,
			anime.MalID,
			anime.AnilistID,
			updater,
			anime.Name,
			anime.CurrentSeason,
			anime,
		)

		if updater != nil {
			updater.Stop()
		}

		action := movieActionAfterPlayback(playErr)
		if action == movieAskUser {
			if playErr != nil {
				log.Printf("Error during movie playback: %v", util.LogSafe(playErr))
			}
			// Ask user what to do next after movie finishes
			action = movieActionAfterPrompt(env.userInput(true))
		}

		switch action {
		case movieQuit:
			log.Println("Quitting application as per user request.")
			return nil
		case movieReplay:
			// For movies, other navigation options don't make much sense, so just continue playing the same movie
			log.Println("Replaying the same movie...")
		case movieChange:
			change := resolveAnimeChange()
			switch change.outcome {
			case changeBack:
				// A CONTINUE row pends for the session loop to reopen.
				return player.ErrBackToAnimeSelection
			case changeStay:
				continue // Stay with current anime if change fails
			case changeToSeries:
				anime, episodes = change.anime, change.episodes
				log.Printf("Switched to series: %s with %d episodes.\n", util.LogSafe(anime.Name), len(episodes))
				if seriesErr := env.handleSeries(ctx, anime, episodes, len(episodes), discordEnabled); errors.Is(seriesErr, player.ErrBackToAnimeSelection) {
					return seriesErr
				}
				return nil
			case changeToMovie:
				anime, episodes = change.anime, change.episodes
				fmt.Printf("Switched to movie: %s\n", anime.Name)
				continue // Continue with new movie
			}
		}
	}
}

// getMovieData is api.GetMovieData; a var so tests can stand in a slow or
// failing Jikan.
var getMovieData = api.GetMovieData

// fetchMovieData starts the Jikan lookup for a movie/OVA in the background
// and returns where its episodes will arrive. Nothing is sent when the lookup
// does not apply or fails.
//
// The lookup is best-effort, and nothing on the way to playback reads it. It
// used to run before the stream was resolved, so an unreachable Jikan held the
// screen blank until its request timed out and then printed a raw error. It
// now runs alongside stream resolution, as the episode lookup in PlayEpisode
// does, and a failure only reaches the debug log. GetMovieData writes into
// Episodes, so it works on its own copy.
func fetchMovieData(anime *models.Anime) <-chan []models.Episode {
	out := make(chan []models.Episode, 1)
	// Only anime has Jikan data; movies and TV get theirs from TMDB/OMDb.
	if anime.IsMovieOrTV() || anime.MalID <= 0 {
		return out
	}

	lookup, malID := getMovieData, anime.MalID
	data := *anime
	data.Episodes = append([]models.Episode(nil), anime.Episodes...)
	go func() {
		if err := lookup(malID, &data); err != nil {
			util.Debugf("Movie metadata unavailable: %v", err)
			return
		}
		out <- data.Episodes
	}()
	return out
}

// createUpdater cria um atualizador de Discord Rich Presence se estiver habilitado
func createUpdater(anime *models.Anime, isPaused *bool, animeMutex *sync.Mutex, episodeDuration time.Duration, discordEnabled bool) *discord.RichPresenceUpdater {
	if !discordEnabled {
		return nil
	}
	// Use 3 second update interval for precise timing sync
	// Discord timestamps need regular updates to stay accurate
	return discord.NewRichPresenceUpdater(
		anime,
		isPaused,
		animeMutex,
		3*time.Second,
		episodeDuration,
		getSocketPath(),
		player.MpvSendCommand,
	)
}

// getSocketPath retorna o caminho do socket MPV baseado no sistema operacional
func getSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\goanime_mpvsocket`
	}
	// Use os.TempDir() for macOS compatibility
	return filepath.Join(os.TempDir(), "mpvsocket")
}

// movieAction is what HandleMovie does once playback returns.
type movieAction int

const (
	movieAskUser movieAction = iota // playback ended on its own: ask what next
	movieQuit
	movieChange // search for another title
	movieReplay
)

// movieActionAfterPlayback reads the outcome of playback. The user quitting or
// asking for another title from inside the player acts at once; anything else,
// including a playback error, falls through to the post-playback menu.
func movieActionAfterPlayback(playErr error) movieAction {
	switch {
	case errors.Is(playErr, player.ErrUserQuit):
		return movieQuit
	case errors.Is(playErr, player.ErrChangeAnime):
		return movieChange
	default:
		return movieAskUser
	}
}

// movieActionAfterPrompt maps the post-playback menu's answer. Both "c" and
// "back" lead to a new search; a movie has no episodes to navigate, so every
// other answer replays it.
func movieActionAfterPrompt(input string) movieAction {
	switch input {
	case "q":
		return movieQuit
	case "c", "back":
		return movieChange
	default:
		return movieReplay
	}
}

// changeOutcome is how a change-of-title request ended.
type changeOutcome int

const (
	changeStay     changeOutcome = iota // it failed: keep the current title
	changeBack                          // a CONTINUE row is pending: return to the session loop
	changeToSeries                      // the new title has episodes to navigate
	changeToMovie                       // the new title is a single movie
)

// changeAnimeFn is ChangeAnimeLocal; a var so tests can answer for the TUI.
var changeAnimeFn = ChangeAnimeLocal

type animeChange struct {
	outcome  changeOutcome
	anime    *models.Anime
	episodes []models.Episode
}

// resolveAnimeChange asks for another title and decides where it goes.
func resolveAnimeChange() animeChange {
	newAnime, newEpisodes, err := changeAnimeFn()
	if errors.Is(err, appflow.ErrHomeContinue) {
		return animeChange{outcome: changeBack}
	}
	if err != nil {
		log.Printf("Error changing anime: %v", util.LogSafe(err))
		return animeChange{outcome: changeStay}
	}
	out := animeChange{outcome: changeToMovie, anime: newAnime, episodes: newEpisodes}
	if isSeriesSwitch(newAnime, len(newEpisodes)) {
		out.outcome = changeToSeries
	}
	return out
}

// isSeriesSwitch tells a series from a movie by media type first and episode
// count second. It avoids re-fetching episodes, which would repeat the season
// selection for FlixHQ.
func isSeriesSwitch(a *models.Anime, totalEpisodes int) bool {
	return !a.IsMovie() && totalEpisodes > 1
}
