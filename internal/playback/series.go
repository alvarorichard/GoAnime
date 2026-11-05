package playback

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/appflow"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// printEpisodeNotFoundMsg prints a user-friendly warning when the selected
// episode doesn't exist on the current source.
func printEpisodeNotFoundMsg() {
	util.Warnf("This episode does not exist in this source. Try another episode.")
}

func HandleSeries(ctx context.Context, anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) error {
	tui.ResetTerminal()
	player.SetEpisodeListContext(anime.Name, anime.Source, anime.Audio)
	if anime.IsTV() {
		fmt.Printf("The selected TV show has %d episodes.\n", totalEpisodes)
	} else {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
	}
	animeMutex := sync.Mutex{}
	isPaused := false

	sel, err := selectFirstEpisode(episodes)
	if err != nil {
		return err
	}

	for {
		playErr := PlayEpisode(
			ctx,
			anime,
			episodes,
			sel.num,
			sel.url,
			sel.label,
			discordEnabled,
			&isPaused,
			&animeMutex,
		)

		var userInput string
		action := seriesActionAfterPlayback(playErr)
		if action == seriesAskUser {
			if playErr != nil {
				log.Printf("Error during episode playback: %v", util.LogSafe(playErr))
			}
			userInput = GetUserInput()
			action = seriesActionAfterPrompt(userInput)
		}

		switch action {
		case seriesQuit:
			log.Println("Quitting application as per user request.")
			return nil

		case seriesReselectFromServer:
			// Back from server selection: the picker opens again, and backing
			// out of it leaves for anime selection.
			next, err := reselectEpisode(episodes, sel, true)
			if err != nil {
				return err
			}
			sel = next

		case seriesPickEpisode:
			// Back or failure keeps the current (valid) selection.
			sel, _ = reselectEpisode(episodes, sel, false)

		case seriesChange:
			change := resolveAnimeChange()
			switch change.outcome {
			case changeBack:
				// A CONTINUE row pends for the session loop to reopen.
				return player.ErrBackToAnimeSelection
			case changeStay:
				continue // Stay with current anime if change fails
			case changeToMovie:
				anime, episodes = change.anime, change.episodes
				log.Println("Switched to a movie/OVA, handling as single episode.")
				if movieErr := HandleMovie(ctx, anime, episodes, discordEnabled); errors.Is(movieErr, player.ErrBackToAnimeSelection) {
					return movieErr
				}
				return nil
			case changeToSeries:
				anime, episodes, totalEpisodes = change.anime, change.episodes, len(change.episodes)
				// There is no valid selection to fall back to (the old one
				// belongs to the previous anime), so a failed pick bounces to
				// anime selection instead of committing zeros.
				next, err := pickEpisode(episodes)
				if err != nil {
					log.Printf("Error selecting episode for new anime: %v", util.LogSafe(err))
					return player.ErrBackToAnimeSelection
				}
				sel = next
				fmt.Printf("Switched to anime: %s with %d episodes.\n", anime.Name, totalEpisodes)
			}

		case seriesNavigate:
			url, label, num := handleUserNavigationEnhanced(userInput, episodes, sel.num, totalEpisodes, anime)
			next, moved := navigationResult(sel, url, label, num)
			if !moved {
				// Navigation failed (e.g. fuzzy finder error) — retry selection
				log.Println("Episode navigation failed, please select again.")
			}
			sel = next
		}
	}
}

// selectEpisodeFuncType matches player.SelectEpisodeWithFuzzyFinder so tests
// can swap in a mock that doesn't open a TUI.
type selectEpisodeFuncType func(episodes []models.Episode) (string, string, error)

// extractEpisodeNumberFuncType matches player.ExtractEpisodeNumber for
// symmetric injection — keeps the entire SelectInitialEpisode pipeline
// driveable without touching the player package state.
type extractEpisodeNumberFuncType func(s string) string

// selectEpisodeFunc and extractEpisodeNumberFunc are package-level
// indirections injected by tests. Production code never touches them.
var (
	selectEpisodeFunc        selectEpisodeFuncType        = player.SelectEpisodeWithFuzzyFinder
	extractEpisodeNumberFunc extractEpisodeNumberFuncType = player.ExtractEpisodeNumber
)

// PreselectEpisodeNum skips the episode picker once: the next
// SelectInitialEpisode opens this episode number directly instead of asking.
// The SEARCH home screen's CONTINUE list sets it to reopen a tracked episode;
// a missing episode falls back to the picker. Zero (the default) asks.
var PreselectEpisodeNum int

// SelectInitialEpisode runs the fuzzy-finder selector then parses the result.
// The TUI call goes through selectEpisodeFunc, so tests inject a mock.
func SelectInitialEpisode(episodes []models.Episode) (episodeURL, episodeNumber string, episodeIndex int, err error) {
	if PreselectEpisodeNum > 0 {
		num := PreselectEpisodeNum
		PreselectEpisodeNum = 0
		url, numStr, epNum, findErr := FindEpisodeByNumber(episodes, num)
		if findErr == nil {
			util.Debugf("[TRACE] SelectInitialEpisode: preselected episode %d", num)
			return url, numStr, epNum, nil
		}
		util.Debugf("[TRACE] SelectInitialEpisode: preselected episode %d missing, asking", num)
	}
	util.Debugf("[TRACE] SelectInitialEpisode: calling selector with %d episodes", len(episodes))
	url, numStr, err := selectEpisodeFunc(episodes)
	util.Debugf("[TRACE] SelectInitialEpisode: returned url=%q, num=%q, err=%v", url, numStr, err)
	return parseEpisodeSelection(url, numStr, err)
}

// parseEpisodeSelection is the pure post-processing of a fuzzy-finder result.
// All branches (back, generic error, atoi success, atoi failure) are exposed
// here so they can be table-tested without TUI involvement.
func parseEpisodeSelection(url, numStr string, fuzzyErr error) (episodeURL, episodeNumber string, episodeIndex int, err error) {
	if fuzzyErr != nil {
		if errors.Is(fuzzyErr, player.ErrBackRequested) {
			return "", "", -1, player.ErrBackRequested
		}
		return "", "", 0, fuzzyErr
	}
	epNum, err := strconv.Atoi(extractEpisodeNumberFunc(numStr))
	if err != nil {
		return "", "", 0, err
	}
	return url, numStr, epNum, nil
}

func handleUserNavigation(input string, episodes []models.Episode, currentNum, totalEpisodes int) (episodeURL, episodeNumber string, episodeIndex int) {
	var url, numStr string
	var epNum int
	var err error

	switch input {
	case "e":
		url, numStr, epNum, err = SelectEpisodeWithFuzzy(episodes)
	case "p":
		newNum := max(currentNum-1, 1)
		url, numStr, epNum, err = FindEpisodeByNumber(episodes, newNum)
	default: // 'n' or default
		newNum := min(currentNum+1, totalEpisodes)
		url, numStr, epNum, err = FindEpisodeByNumber(episodes, newNum)
	}
	if err != nil {
		log.Printf("Navigation error: %v", util.LogSafe(err))
		return "", "", currentNum
	}
	return url, numStr, epNum
}

// handleUserNavigationEnhanced is the navigation entry point for every source.
//
// It used to branch to an AllAnime-specific navigator that walked the episode
// list through the AllAnime API. That source is gone, and no remaining source
// needs out-of-band navigation: every one of them returns a complete episode
// list up front, so index-based navigation is both sufficient and cheaper.
func handleUserNavigationEnhanced(input string, episodes []models.Episode, currentNum, totalEpisodes int, _ *models.Anime) (episodeURL, episodeNumber string, episodeIndex int) {
	return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
}

func CheckIfSeries(url string) (isSeries bool, episodeCount int) {
	series, totalEpisodes, err := api.IsSeries(url)
	if err != nil {
		// Instead of killing the app, assume series unknown -> treat as single episode (movie)
		log.Printf("Error checking if the anime is a series: %v", util.LogSafe(util.ErrorHandler(err)))
		return false, 1
	}
	return series, totalEpisodes
}

// CheckIfSeriesEnhanced checks if anime is a series using enhanced API
func CheckIfSeriesEnhanced(anime *models.Anime) (isSeries bool, episodeCount int) {
	series, totalEpisodes, err := api.IsSeriesEnhanced(anime)
	if err != nil {
		log.Printf("Error checking if the anime is a series: %v", util.LogSafe(util.ErrorHandler(err)))
		return false, 1
	}
	return series, totalEpisodes
}

// ChangeAnimeLocal changes the title through the SEARCH home screen (see
// appflow.ChangeAnimeHome): type a query or pick CONTINUE/RECENT. A CONTINUE
// row pends for the caller with appflow.ErrHomeContinue.
func ChangeAnimeLocal() (*models.Anime, []models.Episode, error) {
	anime, episodes, err := appflow.ChangeAnimeHome()
	if err != nil {
		return nil, nil, err
	}
	if anime != nil {
		player.SetEpisodeListContext(anime.Name, anime.Source, anime.Audio)
	}
	return anime, episodes, nil
}

// episodeSel is the episode being played: its page, its label as the source
// words it, and its number.
type episodeSel struct {
	url   string
	label string
	num   int
}

// pickEpisode opens the episode picker once (or takes the CONTINUE preselect).
func pickEpisode(episodes []models.Episode) (episodeSel, error) {
	url, label, num, err := SelectInitialEpisode(episodes)
	if err != nil {
		return episodeSel{}, err
	}
	return episodeSel{url: url, label: label, num: num}, nil
}

// maxEpisodeRetries is how many times the first pick is attempted.
const maxEpisodeRetries = 3

// selectFirstEpisode runs the first pick, retrying a failed one. Backing out
// of the picker leaves for anime selection at once; running out of attempts is
// an error.
func selectFirstEpisode(episodes []models.Episode) (episodeSel, error) {
	var lastErr error
	for attempt := range maxEpisodeRetries {
		sel, err := pickEpisode(episodes)
		if err == nil {
			return sel, nil
		}
		if errors.Is(err, player.ErrBackRequested) {
			return episodeSel{}, player.ErrBackToAnimeSelection
		}
		lastErr = err
		printEpisodeNotFoundMsg()
		util.Warnf("Episode selection failed (attempt %d/%d): %v", attempt+1, maxEpisodeRetries, err)
	}
	return episodeSel{}, fmt.Errorf("episode selection failed after %d attempts: %w", maxEpisodeRetries, lastErr)
}

// reselectEpisode asks for another episode while one is playing. A failed pick
// keeps current: committing the zero values of a failed selection made the
// loop replay a fabricated empty episode. When backLeaves is set, backing out
// of the picker leaves for anime selection instead.
func reselectEpisode(episodes []models.Episode, current episodeSel, backLeaves bool) (episodeSel, error) {
	next, err := pickEpisode(episodes)
	if err == nil {
		return next, nil
	}
	if errors.Is(err, player.ErrBackRequested) {
		if backLeaves {
			return current, player.ErrBackToAnimeSelection
		}
		return current, nil
	}
	log.Printf("Error selecting episode: %v", util.LogSafe(err))
	return current, nil
}

// navigationResult accepts a navigation answer only when it names an episode.
func navigationResult(current episodeSel, url, label string, num int) (episodeSel, bool) {
	if url == "" {
		return current, false
	}
	return episodeSel{url: url, label: label, num: num}, true
}

// seriesAction is what HandleSeries does once an episode's playback returns.
type seriesAction int

const (
	seriesAskUser            seriesAction = iota // playback ended: show the menu
	seriesQuit                                   // leave the application
	seriesReselectFromServer                     // back from server selection: pick the episode again
	seriesChange                                 // search for another title
	seriesPickEpisode                            // the menu's "Select episode"
	seriesNavigate                               // next / previous
)

// seriesActionAfterPlayback reads the outcome of playback. Quit, going back to
// episode selection and changing the title act at once; anything else, including
// an error, falls through to the post-playback menu.
func seriesActionAfterPlayback(playErr error) seriesAction {
	switch {
	case errors.Is(playErr, player.ErrUserQuit):
		return seriesQuit
	case errors.Is(playErr, player.ErrBackToEpisodeSelection):
		return seriesReselectFromServer
	case errors.Is(playErr, player.ErrChangeAnime):
		return seriesChange
	default:
		return seriesAskUser
	}
}

// seriesActionAfterPrompt maps the post-playback menu's answer. Both "c" and
// "back" lead to a new search; anything unrecognised navigates, where an
// unknown answer behaves as "next".
func seriesActionAfterPrompt(input string) seriesAction {
	switch input {
	case "q", "quit":
		return seriesQuit
	case "c", "back":
		return seriesChange
	case "e":
		return seriesPickEpisode
	default:
		return seriesNavigate
	}
}
