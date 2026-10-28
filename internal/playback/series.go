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

	var selectedEpisodeURL, episodeNumberStr string
	var selectedEpisodeNum int
	const maxEpisodeRetries = 3
	var selErr error
	for attempt := range maxEpisodeRetries {
		selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, selErr = SelectInitialEpisode(episodes)
		if selErr == nil {
			break
		}
		if errors.Is(selErr, player.ErrBackRequested) {
			return player.ErrBackToAnimeSelection
		}
		printEpisodeNotFoundMsg()
		util.Warnf("Episode selection failed (attempt %d/%d): %v", attempt+1, maxEpisodeRetries, selErr)
	}
	if selErr != nil {
		return fmt.Errorf("episode selection failed after %d attempts: %w", maxEpisodeRetries, selErr)
	}

	for {
		err := PlayEpisode(
			ctx,
			anime,
			episodes,
			selectedEpisodeNum,
			selectedEpisodeURL,
			episodeNumberStr,
			discordEnabled,
			&isPaused,
			&animeMutex,
		)

		// Check if user quit during video playback
		if errors.Is(err, player.ErrUserQuit) {
			log.Println("Quitting application as per user request.")
			break
		}

		// Check if user requested to go back to episode selection (from server selection)
		if errors.Is(err, player.ErrBackToEpisodeSelection) {
			newURL, newNumStr, newNum, selErr := SelectInitialEpisode(episodes)
			if selErr != nil {
				// If user selected back at episode selection, go back to anime selection
				if errors.Is(selErr, player.ErrBackRequested) {
					return player.ErrBackToAnimeSelection
				}
				// Keep the previous selection: committing the zero values of a
				// failed selection made the loop replay a fabricated empty episode.
				log.Printf("Error selecting episode: %v", util.LogSafe(selErr))
				continue
			}
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = newURL, newNumStr, newNum
			continue
		}

		// Check if user requested to change anime during video playback
		if errors.Is(err, player.ErrChangeAnime) {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if errors.Is(err, appflow.ErrHomeContinue) {
				// A CONTINUE row pends for the session loop to reopen.
				return player.ErrBackToAnimeSelection
			}
			if err != nil {
				log.Printf("Error changing anime: %v", util.LogSafe(err))
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series using media type first, then episode count as fallback
			// This avoids re-fetching episodes which would cause duplicate season selection for FlixHQ
			newTotalEpisodes := len(newEpisodes)
			totalEpisodes = newTotalEpisodes
			series := !newAnime.IsMovie() && newTotalEpisodes > 1

			if !series {
				// If new anime is a movie, handle it differently
				log.Println("Switched to a movie/OVA, handling as single episode.")
				if err := HandleMovie(ctx, anime, episodes, discordEnabled); err != nil {
					if errors.Is(err, player.ErrBackToAnimeSelection) {
						return err
					}
				}
				break
			}

			// Select initial episode for the new anime. On failure there is no
			// valid selection to fall back to (the old one belongs to the previous
			// anime), so bounce to anime selection instead of committing zeros.
			newURL, newNumStr, newNum, selErr := SelectInitialEpisode(episodes)
			if selErr != nil {
				log.Printf("Error selecting episode for new anime: %v", util.LogSafe(selErr))
				return player.ErrBackToAnimeSelection
			}
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = newURL, newNumStr, newNum

			fmt.Printf("Switched to anime: %s with %d episodes.\n", anime.Name, totalEpisodes)
			continue // Skip normal navigation and start playing the new anime
		}

		// Handle other errors
		if err != nil {
			log.Printf("Error during episode playback: %v", util.LogSafe(err))
		}

		userInput := GetUserInput()
		if userInput == "q" || userInput == "quit" {
			log.Println("Quitting application as per user request.")
			break
		}

		// Handle back/change anime - both options allow searching for a new anime
		if userInput == "c" || userInput == "back" {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if errors.Is(err, appflow.ErrHomeContinue) {
				// A CONTINUE row pends for the session loop to reopen.
				return player.ErrBackToAnimeSelection
			}
			if err != nil {
				log.Printf("Error changing anime: %v", util.LogSafe(err))
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series using media type first, then episode count as fallback
			// This avoids re-fetching episodes which would cause duplicate season selection for FlixHQ
			newTotalEpisodes := len(newEpisodes)
			totalEpisodes = newTotalEpisodes
			series := !newAnime.IsMovie() && newTotalEpisodes > 1

			if !series {
				// If new anime is a movie, handle it differently
				log.Println("Switched to a movie/OVA, handling as single episode.")
				if err := HandleMovie(ctx, anime, episodes, discordEnabled); err != nil {
					if errors.Is(err, player.ErrBackToAnimeSelection) {
						return err
					}
				}
				break
			}

			// Select initial episode for the new anime. On failure there is no
			// valid selection to fall back to (the old one belongs to the previous
			// anime), so bounce to anime selection instead of committing zeros.
			newURL, newNumStr, newNum, selErr := SelectInitialEpisode(episodes)
			if selErr != nil {
				log.Printf("Error selecting episode for new anime: %v", util.LogSafe(selErr))
				return player.ErrBackToAnimeSelection
			}
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = newURL, newNumStr, newNum

			fmt.Printf("Switched to anime: %s with %d episodes.\n", anime.Name, totalEpisodes)
			continue // Skip normal navigation and start playing the new anime
		}

		// Handle episode selection
		if userInput == "e" {
			newURL, newNumStr, newNum, selErr := SelectInitialEpisode(episodes)
			if selErr != nil {
				// Back or failure: keep the current (valid) selection instead
				// of committing the zero values of a failed selection.
				if !errors.Is(selErr, player.ErrBackRequested) {
					log.Printf("Error selecting episode: %v", util.LogSafe(selErr))
				}
				continue
			}
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = newURL, newNumStr, newNum
			continue
		}

		newURL, newNumStr, newNum := handleUserNavigationEnhanced(
			userInput,
			episodes,
			selectedEpisodeNum,
			totalEpisodes,
			anime,
		)
		if newURL == "" {
			// Navigation failed (e.g. fuzzy finder error) — retry selection
			log.Println("Episode navigation failed, please select again.")
			continue
		}
		selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = newURL, newNumStr, newNum
	}
	return nil
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
