package playback

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Style definitions for beautiful series UI
	seriesTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true).
				Underline(true)

	seriesInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Bold(true)

	seriesSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	// seriesWarningStyle = lipgloss.NewStyle().
	// 			Foreground(lipgloss.Color("#FFD700")).
	// 			Bold(true)

	seriesBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(1, 2)
)

func HandleSeries(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) {
	// Create beautiful series header
	seriesHeader := seriesTitleStyle.Render(fmt.Sprintf("📺 %s", anime.Details.Title.Romaji))
	seriesInfo := seriesInfoStyle.Render(fmt.Sprintf("🎬 Series • %d Episodes Available", totalEpisodes))

	headerBox := seriesBoxStyle.Render(
		seriesHeader + "\n" +
			seriesInfo,
	)
	fmt.Println("\n" + headerBox)

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
			// Display beautiful goodbye message
			goodbyeMsg := seriesSuccessStyle.Render("🚪 ✨ Thanks for watching! Goodbye!")
			fmt.Println("\n" + goodbyeMsg)
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
