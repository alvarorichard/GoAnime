package playback

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
)

func HandleSeries(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) {
	fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
	animeMutex := sync.Mutex{}
	isPaused := false

	selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err := SelectInitialEpisode(episodes)
	if err != nil {
		log.Printf("Episode selection error: %v", util.ErrorHandler(err))
		return
	}

	for {
		err := PlayEpisode(
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

		// Check if user requested to change anime during video playback
		if errors.Is(err, player.ErrChangeAnime) {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series and get new total episodes
			series, newTotalEpisodes := CheckIfSeriesEnhanced(anime)
			totalEpisodes = newTotalEpisodes

			if !series {
				// If new anime is a movie, handle it differently
				log.Println("Switched to a movie/OVA, handling as single episode.")
				HandleMovie(anime, episodes, discordEnabled)
				break
			}

			// Select initial episode for the new anime
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err = SelectInitialEpisode(episodes)
			if err != nil {
				log.Printf("Error selecting episode for new anime: %v", err)
				continue
			}

			fmt.Printf("Switched to anime: %s with %d episodes.\n", anime.Name, totalEpisodes)
			continue // Skip normal navigation and start playing the new anime
		}

		// Handle other errors
		if err != nil {
			log.Printf("Error during episode playback: %v", err)
		}

		userInput := GetUserInput()
		if userInput == "q" || userInput == "quit" {
			log.Println("Quitting application as per user request.")
			break
		}

		// Handle anime change
		if userInput == "c" {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series and get new total episodes
			series, newTotalEpisodes := CheckIfSeriesEnhanced(anime)
			totalEpisodes = newTotalEpisodes

			if !series {
				// If new anime is a movie, handle it differently
				log.Println("Switched to a movie/OVA, handling as single episode.")
				HandleMovie(anime, episodes, discordEnabled)
				break
			}

			// Select initial episode for the new anime
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err = SelectInitialEpisode(episodes)
			if err != nil {
				log.Printf("Error selecting episode for new anime: %v", err)
				continue
			}

			fmt.Printf("Switched to anime: %s with %d episodes.\n", anime.Name, totalEpisodes)
			continue // Skip normal navigation and start playing the new anime
		}

		// Handle episode selection
		if userInput == "e" {
			selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err = SelectInitialEpisode(episodes)
			if err != nil {
				log.Printf("Error selecting episode: %v", err)
				continue
			}
			continue
		}

		selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = handleUserNavigationEnhanced(
			userInput,
			episodes,
			selectedEpisodeNum,
			totalEpisodes,
			anime,
		)
	}
}

func SelectInitialEpisode(episodes []models.Episode) (string, string, int, error) {
	selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		return "", "", 0, err
	}
	selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
	if err != nil {
		return "", "", 0, err
	}
	return selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, nil
}

func handleUserNavigation(input string, episodes []models.Episode, currentNum, totalEpisodes int) (string, string, int) {
	switch input {
	case "e":
		return SelectEpisodeWithFuzzy(episodes)
	case "p":
		newNum := currentNum - 1
		if newNum < 1 {
			newNum = 1
		}
		return FindEpisodeByNumber(episodes, newNum)
	default: // 'n' or default
		newNum := currentNum + 1
		if newNum > totalEpisodes {
			newNum = totalEpisodes
		}
		return FindEpisodeByNumber(episodes, newNum)
	}
}

// Enhanced navigation handler that supports AllAnime-specific navigation
func handleUserNavigationEnhanced(input string, episodes []models.Episode, currentNum, totalEpisodes int, anime *models.Anime) (string, string, int) {
	// Check if this is an AllAnime source and use enhanced navigation
	if isAllAnimeSource(anime) {
		return handleAllAnimeNavigation(input, episodes, currentNum, totalEpisodes, anime)
	}

	// Fallback to regular navigation for other sources
	return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
}

// AllAnime-specific navigation handler
func handleAllAnimeNavigation(input string, episodes []models.Episode, currentNum, totalEpisodes int, anime *models.Anime) (string, string, int) {
	// Find current episode string
	currentEpisodeStr := ""
	for _, ep := range episodes {
		if ep.Num == currentNum {
			currentEpisodeStr = ep.Number
			break
		}
	}

	if currentEpisodeStr == "" {
		util.Debug("Current episode not found, falling back to regular navigation", "currentNum", currentNum)
		return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
	}

	switch input {
	case "e":
		return SelectEpisodeWithFuzzy(episodes)
	case "p":
		// Use AllAnime navigator for previous episode
		nextEp, err := HandleAllAnimeEpisodeNavigation(anime, currentEpisodeStr, "previous")
		if err != nil {
			util.Debug("AllAnime previous navigation failed, using fallback", "error", err.Error())
			return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
		}
		return nextEp.URL, nextEp.Number, nextEp.Num
	case "n":
		// Use AllAnime navigator for next episode
		nextEp, err := HandleAllAnimeEpisodeNavigation(anime, currentEpisodeStr, "next")
		if err != nil {
			util.Debug("AllAnime next navigation failed, using fallback", "error", err.Error())
			return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
		}
		return nextEp.URL, nextEp.Number, nextEp.Num
	default:
		return handleUserNavigation(input, episodes, currentNum, totalEpisodes)
	}
}

func CheckIfSeries(url string) (bool, int) {
	series, totalEpisodes, err := api.IsSeries(url)
	if err != nil {
		// Instead of killing the app, assume series unknown -> treat as single episode (movie)
		log.Printf("Error checking if the anime is a series: %v", util.ErrorHandler(err))
		return false, 1
	}
	return series, totalEpisodes
}

// CheckIfSeriesEnhanced checks if anime is a series using enhanced API
func CheckIfSeriesEnhanced(anime *models.Anime) (bool, int) {
	series, totalEpisodes, err := api.IsSeriesEnhanced(anime)
	if err != nil {
		log.Printf("Error checking if the anime is a series: %v", util.ErrorHandler(err))
		return false, 1
	}
	return series, totalEpisodes
}

// ChangeAnimeLocal allows the user to search for and select a new anime (local implementation to avoid circular imports)
func ChangeAnimeLocal() (*models.Anime, []models.Episode, error) {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		var animeName string

		prompt := huh.NewInput().
			Title("Change Anime").
			Description("Enter the name of the anime you want to watch:").
			Value(&animeName).
			Validate(func(v string) error {
				if len(v) < 2 {
					return fmt.Errorf("anime name must be at least 2 characters")
				}
				return nil
			})

		if err := prompt.Run(); err != nil {
			return nil, nil, err
		}

		// Use the enhanced API to search for anime
		anime, err := api.SearchAnimeEnhanced(animeName, "")
		if err != nil || anime == nil {
			if i < maxRetries-1 {
				util.Errorf("No anime found with the name: %s", animeName)
				util.Infof("Please try again with a different search term. (Attempt %d/%d)", i+2, maxRetries)
				continue
			}
			return nil, nil, fmt.Errorf("failed to find anime after %d attempts", maxRetries)
		}

		// Get episodes for the new anime using enhanced API
		episodes, err := api.GetAnimeEpisodesEnhanced(anime)
		if err != nil {
			if i < maxRetries-1 {
				util.Errorf("Failed to get episodes for: %s", anime.Name)
				util.Infof("Please try searching for a different anime. (Attempt %d/%d)", i+2, maxRetries)
				continue
			}
			return nil, nil, fmt.Errorf("failed to get episodes after %d attempts", maxRetries)
		}

		return anime, episodes, nil
	}

	return nil, nil, fmt.Errorf("failed to change anime after %d attempts", maxRetries)
}
