package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	version = "1.1.0"
)

func main() {
	startAll := time.Now()
	versionFlag := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *versionFlag || hasVersionArg() {
		showVersion()
		return
	}

	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	handleTrackingNotice()
	if util.IsDebug {
		log.Printf("[PERF] starting Goanime v%s", version)
	}

	discordManager := discord.NewManager()
	if err := discordManager.Initialize(); err != nil {
		if util.IsDebug {
			log.Println("Failed to initialize Discord Rich Presence:", err)
		}
	} else {
		defer discordManager.Shutdown()
	}

	anime := searchAnime(animeName)
	fetchAnimeDetails(anime)
	episodes := getAnimeEpisodes(anime.URL)

	if util.IsDebug {
		log.Printf("[PERF] Full boot in %v", time.Since(startAll))
	}

	series, totalEpisodes := checkIfSeries(anime.URL)
	if series {
		handleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
	} else {
		handleMovie(anime, episodes, discordManager.IsEnabled())
	}
}

func hasVersionArg() bool {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		return arg == "--version" || arg == "-version"
	}
	return false
}

func showVersion() {
	fmt.Printf("GoAnime v%s", version)
	if tracking.IsCgoEnabled {
		fmt.Println(" (with SQLite tracking)")
	} else {
		fmt.Println(" (without SQLite tracking)")
	}
}

func handleTrackingNotice() {
	if !tracking.IsCgoEnabled {
		fmt.Println("Notice: Anime progress tracking disabled (CGO not available)")
		fmt.Println("Episode progress and resume features will not be available.")
		fmt.Println()
	}
}

func searchAnime(name string) *models.Anime {
	searchStart := time.Now()
	anime, err := api.SearchAnime(name)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}
	if util.IsDebug {
		log.Printf("[PERF] Busca de anime em %v", time.Since(searchStart))
	}
	return anime
}

func fetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()
	if err := api.FetchAnimeDetails(anime); err != nil {
		log.Println("Failed to fetch anime details:", err)
	}
	if util.IsDebug {
		log.Printf("[PERF] Search in details %v", time.Since(detailsStart))
	}
}

func getAnimeEpisodes(url string) []models.Episode {
	episodesStart := time.Now()
	episodes, err := api.GetAnimeEpisodes(url)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}
	if util.IsDebug {
		log.Printf("[PERF] Search Episode in %v", time.Since(episodesStart))
	}
	return episodes
}

func checkIfSeries(url string) (bool, int) {
	series, totalEpisodes, err := api.IsSeries(url)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}
	return series, totalEpisodes
}

func getSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\goanime_mpvsocket`
	}
	return "/tmp/mpvsocket"
}

func createUpdater(anime *models.Anime, isPaused *bool, animeMutex *sync.Mutex, episodeDuration time.Duration, discordEnabled bool) *discord.RichPresenceUpdater {
	if !discordEnabled {
		return nil
	}
	return discord.NewRichPresenceUpdater(
		anime,
		isPaused,
		animeMutex,
		1*time.Second,
		episodeDuration,
		getSocketPath(),
		player.MpvSendCommand,
	)
}

func handleSeries(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) {
	fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
	animeMutex := sync.Mutex{}
	isPaused := false

	selectedEpisodeURL, episodeNumberStr, selectedEpisodeNum, err := selectInitialEpisode(episodes)
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	for {
		playEpisode(
			anime,
			episodes,
			selectedEpisodeNum,
			selectedEpisodeURL,
			episodeNumberStr,
			discordEnabled,
			&isPaused,
			&animeMutex,
		)

		userInput := getUserInput()
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

func selectInitialEpisode(episodes []models.Episode) (string, string, int, error) {
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
		return selectEpisodeWithFuzzy(episodes)
	case "p":
		newNum := currentNum - 1
		if newNum < 1 {
			newNum = 1
		}
		return findEpisodeByNumber(episodes, newNum)
	default: // 'n' or default
		newNum := currentNum + 1
		if newNum > totalEpisodes {
			newNum = totalEpisodes
		}
		return findEpisodeByNumber(episodes, newNum)
	}
}

func selectEpisodeWithFuzzy(episodes []models.Episode) (string, string, int) {
	url, numStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}
	epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(numStr))
	if err != nil {
		log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
	}
	return url, numStr, epNum
}

func findEpisodeByNumber(episodes []models.Episode, num int) (string, string, int) {
	for _, ep := range episodes {
		if epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(ep.Number)); err == nil && epNum == num {
			return ep.URL, ep.Number, num
		}
	}
	log.Printf("Warning: Episode number %d not found. Re-selecting.", num)
	return selectEpisodeWithFuzzy(episodes)
}

func playEpisode(
	anime *models.Anime,
	episodes []models.Episode,
	episodeNum int,
	episodeURL string,
	episodeNumberStr string,
	discordEnabled bool,
	isPaused *bool,
	animeMutex *sync.Mutex,
) {
	animeMutex.Lock()
	anime.Episodes = []models.Episode{{
		Number: episodeNumberStr,
		Num:    episodeNum,
		URL:    episodeURL,
	}}
	animeMutex.Unlock()

	if err := api.GetEpisodeData(anime.MalID, episodeNum, anime); err != nil {
		log.Printf("Error fetching episode data: %v", err)
	}

	videoURL, err := player.GetVideoURLForEpisode(episodeURL)
	if err != nil {
		log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
	}

	episodeDuration := time.Duration(episodes[0].Duration) * time.Second
	updater := createUpdater(anime, isPaused, animeMutex, episodeDuration, discordEnabled)

	player.HandleDownloadAndPlay(
		videoURL,
		episodes,
		episodeNum,
		anime.URL,
		episodeNumberStr,
		anime.MalID,
		updater,
	)

	if updater != nil {
		updater.Stop()
	}
}

func getUserInput() string {
	fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'e' to select episode, 'q' to quit: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		if err.Error() == "unexpected newline" {
			log.Println("No input detected, continuing playback")
			return "n"
		}
		log.Printf("Error reading input: %v - defaulting to continue", util.ErrorHandler(err))
		return "n"
	}
	return input
}

func handleMovie(anime *models.Anime, episodes []models.Episode, discordEnabled bool) {
	animeMutex := sync.Mutex{}
	isPaused := false

	animeMutex.Lock()
	anime.Episodes = []models.Episode{episodes[0]}
	animeMutex.Unlock()

	if err := api.GetMovieData(anime.MalID, anime); err != nil {
		log.Printf("Error fetching movie/OVA data: %v", err)
	}

	videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
	if err != nil {
		log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
	}

	episodeDuration := time.Duration(episodes[0].Duration) * time.Second
	updater := createUpdater(anime, &isPaused, &animeMutex, episodeDuration, discordEnabled)

	player.HandleDownloadAndPlay(
		videoURL,
		episodes,
		1,
		anime.URL,
		episodes[0].Number,
		anime.MalID,
		updater,
	)

	if updater != nil {
		updater.Stop()
	}
}
