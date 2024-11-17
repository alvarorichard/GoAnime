//package main
//
//import (
//	"fmt"
//	"github.com/alvarorichard/Goanime/internal/api"
//	"github.com/alvarorichard/Goanime/internal/player"
//	"github.com/alvarorichard/Goanime/internal/util"
//	"github.com/hugolgst/rich-go/client"
//	"log"
//	"strconv"
//	"sync"
//	"time"
//)
//
//const discordClientID = "1302721937717334128" // Your Discord Client ID
//
//func main() {
//	var animeMutex sync.Mutex
//
//	// Parse flags to get the anime name
//	animeName, err := util.FlagParser()
//	if err != nil {
//		log.Fatalln(util.ErrorHandler(err))
//	}
//
//	// Initial login to Discord Rich Presence
//	err = client.Login(discordClientID)
//	if err != nil {
//		log.Fatalln("Failed to initialize Discord Rich Presence:", err)
//	}
//	defer client.Logout() // Ensure logout on exit
//
//	if util.IsDebug {
//		log.Printf("Attempting AniList search with title: %s\n", animeName)
//
//	}
//
//	// Search for the anime
//	anime, err := api.SearchAnime(animeName)
//	if err != nil {
//		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
//	}
//
//	// Fetch anime details, including cover image URL
//	err = api.FetchAnimeDetails(anime)
//	if err != nil {
//		log.Println("Failed to fetch anime details:", err)
//	}
//
//	// Log existing fields (ensure these fields exist in your api.Anime struct)
//	if util.IsDebug {
//		log.Printf("Title: %s, Cover Image URL: %s\n",
//			anime.Details.Title.Romaji, anime.ImageURL)
//	}
//
//	// Fetch episodes for the anime
//	episodes, err := api.GetAnimeEpisodes(anime.URL) // Corrected to use anime.URL
//	if err != nil || len(episodes) == 0 {
//		log.Fatalln("The selected anime does not have episodes on the server.")
//	}
//
//	// Check if the anime is a series or a movie/OVA
//	series, totalEpisodes, err := api.IsSeries(anime.URL) // Corrected to use anime.URL
//	if err != nil {
//		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
//	}
//
//	if series {
//		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
//	} else {
//		fmt.Println("The selected anime is a movie/OVA. Starting playback directly...")
//	}
//
//	// Define a flag to track whether the playback is paused
//	isPaused := false
//
//	// Initialize the player.RichPresenceUpdater
//	//updater := player.NewRichPresenceUpdater(anime, &isPaused, &animeMutex, 15*time.Second)
//	//defer updater.Stop() // Ensure that the updater is stopped on exit
//
//	// Fetch episode duration in seconds (ensure api.GetEpisodeData includes duration data)
//	episodeDuration := time.Duration(anime.Episodes[0].Duration) * time.Second
//
//	// Initialize RichPresenceUpdater with start time and duration
//	//updater := player.NewRichPresenceUpdater(anime, &isPaused, &animeMutex, 15*time.Second, episodeDuration)
//	socketPath := "/tmp/mpvsocket" // Adjust this to the actual socket path
//	updateFreq := 1 * time.Second  // Update every second
//	//updater := player.NewRichPresenceUpdater(anime, &isPaused, &animeMutex, 15*time.Second, episodeDuration, socketPath)
//	//updater := player.NewRichPresenceUpdater(anime, &animeMutex, updateFreq, episodeDuration, socketPath)
//	//isPaused := false // Define isPaused as a bool
//	updater := player.NewRichPresenceUpdater(anime, &isPaused, &animeMutex, updateFreq, episodeDuration, socketPath)
//
//	defer updater.Stop()
//
//	if series {
//		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
//
//		// Start the Rich Presence updater
//		updater.Start()
//
//		for {
//			// Select an episode with fuzzy finder
//			selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
//			if err != nil {
//				log.Fatalln(util.ErrorHandler(err))
//			}
//
//			selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
//			if err != nil {
//				log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
//			}
//
//			// Update the anime struct with the selected episode
//			animeMutex.Lock()
//			anime.Episodes = []api.Episode{
//				{
//					Number: episodeNumberStr,
//					Num:    selectedEpisodeNum,
//					URL:    selectedEpisodeURL,
//				},
//			}
//			animeMutex.Unlock()
//
//			if util.IsDebug {
//				log.Printf("Selected Episode Updated: %+v", anime.Episodes[0])
//
//			}
//
//			// Fetch additional episode details
//			err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime)
//			if err != nil {
//				log.Printf("Error fetching episode data: %v", err)
//			} else {
//				if util.IsDebug {
//					log.Printf("Episode Details Updated: %+v", anime.Episodes[0])
//
//				}
//			}
//
//			// Get the video URL for the selected episode
//			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
//			if err != nil {
//				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//			}
//
//			// Handle download and play, updating the paused state as necessary
//			player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr, updater) // Corrected to use anime.URL
//
//			// The updater will automatically pick up the latest episode details during the next tick
//
//			// Check for next episode input
//			var userInput string
//			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
//			fmt.Scanln(&userInput)
//			if userInput == "q" {
//				log.Println("Quitting application as per user request.")
//				break
//			} else if userInput == "n" {
//				log.Println("User selected next episode.")
//				continue
//			} else if userInput == "p" {
//				log.Println("User selected previous episode.")
//				continue
//			} else {
//				log.Println("Invalid input, continuing current episode.")
//			}
//		}
//
//	} else {
//		// Update the anime struct with the first episode (movie/OVA)
//		animeMutex.Lock()
//		anime.Episodes = []api.Episode{
//			episodes[0],
//		}
//		animeMutex.Unlock()
//
//		if util.IsDebug {
//			log.Printf("Selected Episode Updated: %+v", anime.Episodes[0])
//
//		}
//		// Fetch additional episode details
//		err = api.GetEpisodeData(anime.MalID, 1, anime)
//		if err != nil {
//			log.Printf("Error fetching episode data: %v", err)
//		} else {
//			log.Printf("Episode Details Updated: %+v", anime.Episodes[0])
//		}
//
//		// Start the Rich Presence updater
//		// updater.Start()
//
//		// Get the video URL for the movie/OVA
//		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		// Handle download and play, updating the paused state as necessary
//		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number, updater) // Corrected to use anime.URL
//	}
//
//	// The deferred updater.Stop() will ensure that the updater is stopped before exiting
//}

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1302721937717334128" // Your Discord Client ID

func main() {
	var animeMutex sync.Mutex

	// Parse flags to get the anime name
	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	// Initialize Discord Rich Presence
	if err := client.Login(discordClientID); err != nil {
		log.Fatalln("Failed to initialize Discord Rich Presence:", err)
	}
	defer client.Logout() // Ensure logout on exit

	// Search for the anime
	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	// Fetch anime details, including cover image URL
	if err = api.FetchAnimeDetails(anime); err != nil {
		log.Println("Failed to fetch anime details:", err)
	}

	// Fetch episodes for the anime
	episodes, err := api.GetAnimeEpisodes(anime.URL)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	// Check if the anime is a series or a movie/OVA
	series, totalEpisodes, err := api.IsSeries(anime.URL)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}

	// Define a flag to track if the playback is paused
	isPaused := false
	socketPath := "/tmp/mpvsocket" // Adjust socket path as per your setup
	updateFreq := 1 * time.Second  // Update frequency for Rich Presence
	episodeDuration := time.Duration(episodes[0].Duration) * time.Second

	// Initialize RichPresenceUpdater
	updater := player.NewRichPresenceUpdater(anime, &isPaused, &animeMutex, updateFreq, episodeDuration, socketPath)
	defer updater.Stop()

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
		updater.Start()

		for {
			// Select an episode using fuzzy finder
			selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
			if err != nil {
				log.Fatalln(util.ErrorHandler(err))
			}

			selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
			if err != nil {
				log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
			}

			// Lock anime struct and update with selected episode
			animeMutex.Lock()
			anime.Episodes = []api.Episode{
				{
					Number: episodeNumberStr,
					Num:    selectedEpisodeNum,
					URL:    selectedEpisodeURL,
				},
			}
			animeMutex.Unlock()

			// Fetch episode details and AniSkip data
			if err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime); err != nil {
				log.Printf("Error fetching episode data: %v", err)
			}

			// Retrieve video URL for the selected episode
			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
			if err != nil {
				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
			}

			// Handle download and playback, updating paused state as necessary
			player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr, updater)

			// Prompt user for next action
			var userInput string
			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
			fmt.Scanln(&userInput)
			if userInput == "q" {
				log.Println("Quitting application as per user request.")
				break
			} else if userInput == "n" || userInput == "p" {
				continue // loop continues for next or previous episode
			} else {
				log.Println("Invalid input, continuing current episode.")
			}
		}

	} else {
		// Handle movie/OVA playback
		animeMutex.Lock()
		anime.Episodes = []api.Episode{episodes[0]}
		animeMutex.Unlock()

		// Fetch details and AniSkip data for the movie/OVA
		if err = api.GetMovieData(anime.MalID, anime); err != nil {
			log.Printf("Error fetching movie/OVA data: %v", err)
		}

		// Get the video URL for the movie/OVA
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		// Handle download and play, with Rich Presence updates
		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number, updater)
	}

	// Deferred updater.Stop() will ensure cleanup on exit
}
