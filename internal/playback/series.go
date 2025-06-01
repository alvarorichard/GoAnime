package playback

import (
	"log"
	"sync"
	"fmt"
	"strconv"
	
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/api"
	
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
		PlayEpisode(
			anime,
			episodes,
			selectedEpisodeNum,
			selectedEpisodeURL,
			episodeNumberStr,
			discordEnabled,
			&isPaused,
			&animeMutex,
		)

		userInput := GetUserInput()
		if userInput == "q" {
			log.Println("Quitting application as per user request.")
			break
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