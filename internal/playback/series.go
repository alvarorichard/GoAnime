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
)

func HandleSeries(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) {
	fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
	animeMutex := sync.Mutex{}
	isPaused := false

	selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err := SelectInitialEpisode(episodes)
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
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
			newAnime, newEpisodes, err := ChangeAnime()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series and get new total episodes
			series, newTotalEpisodes := CheckIfSeries(anime.URL)
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
		if userInput == "q" {
			log.Println("Quitting application as per user request.")
			break
		}

		// Handle anime change
		if userInput == "c" {
			newAnime, newEpisodes, err := ChangeAnime()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series and get new total episodes
			series, newTotalEpisodes := CheckIfSeries(anime.URL)
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

		selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum = handleUserNavigation(
			userInput,
			episodes,
			selectedEpisodeNum,
			totalEpisodes,
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

func CheckIfSeries(url string) (bool, int) {
	series, totalEpisodes, err := api.IsSeries(url)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}
	return series, totalEpisodes
}

// CheckIfSeriesEnhanced checks if anime is a series using enhanced API
func CheckIfSeriesEnhanced(anime *models.Anime) (bool, int) {
	series, totalEpisodes, err := api.IsSeriesEnhanced(anime)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}
	return series, totalEpisodes
}
