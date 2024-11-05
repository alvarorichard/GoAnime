//package main
//
//import (
//	"fmt"
//	"log"
//	"strconv"
//
//	"github.com/alvarorichard/Goanime/internal/api"
//	"github.com/alvarorichard/Goanime/internal/player"
//	"github.com/alvarorichard/Goanime/internal/util"
//)
//
//func main() {
//	animeName, err := util.FlagParser()
//	if err != nil {
//		log.Fatalln(util.ErrorHandler(err))
//	}
//
//	anime, err := api.SearchAnime(animeName)
//	if err != nil {
//		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
//	}
//
//	episodes, err := api.GetAnimeEpisodes(anime.URL)
//	if err != nil || len(episodes) == 0 {
//		if util.IsDebug {
//			log.Fatalln("The selected anime has no episodes on the server:", util.ErrorHandler(err))
//		}
//		log.Fatalln("The selected anime has no episodes on the server.")
//	}
//
//	series, totalEpisodes, err := api.IsSeries(anime.URL)
//	if err != nil {
//		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
//	}
//
//	if series {
//		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
//		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
//		if err != nil {
//			log.Fatalln(util.ErrorHandler(err))
//		}
//
//		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
//		if err != nil {
//			log.Fatalln("Error parsing episode number:", util.ErrorHandler(err))
//		}
//
//		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)
//	} else {
//		fmt.Println("The selected anime is a movie/OVA. Starting direct playback...")
//		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
//	}
//}

//package main
//
//import (
//	"context"
//	"fmt"
//	"log"
//	"strconv"
//	"sync"
//	"time"
//
//	"github.com/alvarorichard/Goanime/internal/api"
//	"github.com/alvarorichard/Goanime/internal/player"
//	"github.com/alvarorichard/Goanime/internal/util"
//	"github.com/hugolgst/rich-go/client"
//)
//
//const discordClientID = "1302721937717334128" // Seu Client ID do Discord
//
//// RichPresenceManager manages the Discord Rich Presence updates.
//type RichPresenceManager struct {
//	mu     sync.Mutex
//	cancel context.CancelFunc
//	wg     sync.WaitGroup
//}
//
//// Start initiates a new Rich Presence update goroutine.
//// It first cancels any existing goroutine before starting a new one.
//func (rpm *RichPresenceManager) Start(discordClientID string, anime *api.Anime, isPaused *bool, animeMutex *sync.Mutex, episodeDetail string) error {
//	rpm.mu.Lock()
//	defer rpm.mu.Unlock()
//
//	// If there's an existing goroutine, cancel it
//	if rpm.cancel != nil {
//		rpm.cancel()
//		rpm.cancel = nil
//	}
//
//	// Create a new context for the new goroutine
//	ctx, cancel := context.WithCancel(context.Background())
//	rpm.cancel = cancel
//
//	// Start the Rich Presence update in a new goroutine
//	rpm.wg.Add(1)
//	go func() {
//		defer rpm.wg.Done()
//		if err := updateDiscordPresence(ctx, discordClientID, anime, isPaused, animeMutex, episodeDetail); err != nil {
//			log.Println("Erro ao atualizar Discord Rich Presence:", err)
//		}
//	}()
//
//	return nil
//}
//
//// Stop cancels any running Rich Presence update goroutine.
//func (rpm *RichPresenceManager) Stop() {
//	rpm.mu.Lock()
//	defer rpm.mu.Unlock()
//
//	if rpm.cancel != nil {
//		rpm.cancel()
//		rpm.cancel = nil
//	}
//	// Wait for the goroutine to finish
//	rpm.wg.Wait()
//}
//
//// updateDiscordPresence updates the Discord Rich Presence.
//// It runs in a goroutine and periodically updates the presence until the context is canceled.
//func updateDiscordPresence(ctx context.Context, discordClientID string, anime *api.Anime, isPaused *bool, animeMutex *sync.Mutex, episodeDetail string) error {
//	animeMutex.Lock()
//	originalRomaji := anime.Details.Title.Romaji
//	anime.Details.Title.Romaji = originalRomaji + " | Ep " + episodeDetail // Modify Title.Romaji to include episode detail
//	animeMutex.Unlock()
//
//	startTimestamp := time.Now().Unix()
//
//	// Initial presence update
//	err := setDiscordPresence(discordClientID, anime, isPaused, startTimestamp)
//	if err != nil {
//		log.Println("Erro ao atualizar o Discord Rich Presence:", err)
//		return err
//	} else {
//		log.Println("Discord Rich Presence atualizado para o episódio:", episodeDetail)
//	}
//
//	// Restore the original Title.Romaji after updating
//	animeMutex.Lock()
//	anime.Details.Title.Romaji = originalRomaji
//	animeMutex.Unlock()
//
//	// Create a ticker for periodic updates (e.g., every 15 seconds)
//	ticker := time.NewTicker(15 * time.Second)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ctx.Done():
//			log.Println("Atualização do Rich Presence cancelada.")
//			return nil
//		case <-ticker.C:
//			err := setDiscordPresence(discordClientID, anime, isPaused, startTimestamp)
//			if err != nil {
//				log.Println("Erro ao atualizar o Discord Rich Presence:", err)
//			} else {
//				log.Println("Discord Rich Presence atualizado.")
//			}
//		}
//	}
//}
//
//// setDiscordPresence sets the Rich Presence data.
//func setDiscordPresence(discordClientID string, anime *api.Anime, isPaused *bool, startTimestamp int64) error {
//	// No need for additional mutex here as it's handled by the caller
//	return api.DiscordPresence(discordClientID, *anime, *isPaused, startTimestamp)
//}
//
//func main() {
//	var animeMutex sync.Mutex
//
//	// Initialize the Rich Presence Manager
//	rpm := &RichPresenceManager{}
//	defer rpm.Stop() // Ensure that any running Rich Presence goroutine is stopped on exit
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
//		log.Fatalln("Falha ao inicializar Discord Rich Presence:", err)
//	}
//	defer client.Logout() // Ensure logout on exit
//
//	// Search for the anime
//	anime, err := api.SearchAnime(animeName)
//	if err != nil {
//		log.Fatalln("Falha ao buscar anime:", util.ErrorHandler(err))
//	}
//
//	// Fetch anime details, including cover image URL
//	err = api.FetchAnimeDetails(anime)
//	if err != nil {
//		log.Println("Falha ao buscar detalhes do anime:", err)
//	}
//
//	// Fetch episodes for the anime
//	episodes, err := api.GetAnimeEpisodes(anime.URL)
//	if err != nil || len(episodes) == 0 {
//		log.Fatalln("O anime selecionado não possui episódios no servidor.")
//	}
//
//	// Check if the anime is a series or a movie/OVA
//	series, totalEpisodes, err := api.IsSeries(anime.URL)
//	if err != nil {
//		log.Fatalln("Erro ao verificar se o anime é uma série:", util.ErrorHandler(err))
//	}
//
//	// Define a flag to track whether the playback is paused
//	isPaused := false
//
//	if series {
//		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)
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
//				log.Fatalln("Erro ao converter número do episódio:", util.ErrorHandler(err))
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
//			// Stop any existing Rich Presence update
//			rpm.Stop()
//
//			// Start a new Rich Presence update for the new episode
//			log.Printf("Iniciando atualização do Rich Presence para o episódio %s...\n", episodeNumberStr)
//			err = rpm.Start(discordClientID, anime, &isPaused, &animeMutex, episodeNumberStr)
//			if err != nil {
//				log.Printf("Erro ao iniciar o Discord Rich Presence para o episódio %s: %v\n", episodeNumberStr, err)
//			} else {
//				log.Printf("Discord Rich Presence iniciado para o episódio: %s\n", episodeNumberStr)
//			}
//
//			// Get the video URL for the selected episode
//			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
//			if err != nil {
//				log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
//			}
//
//			// Handle download and play, updating the paused state as necessary
//			player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)
//
//			// Check for next episode input
//			var userInput string
//			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
//			fmt.Scanln(&userInput)
//			if userInput == "q" {
//				break
//			} else if userInput == "n" {
//				continue
//			} else if userInput == "p" {
//				// Implement previous episode logic if needed
//				continue
//			} else {
//				log.Println("Entrada inválida, continuando episódio atual.")
//			}
//		}
//
//	} else {
//		fmt.Println("O anime selecionado é um filme/OVA. Iniciando reprodução direta...")
//
//		// Update the anime struct with the first episode (movie/OVA)
//		animeMutex.Lock()
//		anime.Episodes = []api.Episode{
//			episodes[0],
//		}
//		animeMutex.Unlock()
//
//		// Start Rich Presence update for the movie/OVA
//		log.Println("Iniciando atualização do Rich Presence para o filme/OVA...")
//		err = rpm.Start(discordClientID, anime, &isPaused, &animeMutex, "OVA")
//		if err != nil {
//			log.Println("Erro ao iniciar o Discord Rich Presence para o filme/OVA:", err)
//		} else {
//			log.Println("Discord Rich Presence iniciado para o filme/OVA.")
//		}
//
//		// Get the video URL for the movie/OVA
//		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
//		if err != nil {
//			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
//		}
//
//		// Handle download and play, updating the paused state as necessary
//		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
//	}
//
//	// The deferred rpm.Stop() will ensure that any running Rich Presence goroutine is stopped before exiting
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

// RichPresenceUpdater manages periodic updates to Discord Rich Presence.
type RichPresenceUpdater struct {
	anime      *api.Anime
	isPaused   *bool
	animeMutex *sync.Mutex
	updateFreq time.Duration
	done       chan bool
	wg         sync.WaitGroup
}

// NewRichPresenceUpdater initializes a new RichPresenceUpdater.
func NewRichPresenceUpdater(anime *api.Anime, isPaused *bool, animeMutex *sync.Mutex, updateFreq time.Duration) *RichPresenceUpdater {
	return &RichPresenceUpdater{
		anime:      anime,
		isPaused:   isPaused,
		animeMutex: animeMutex,
		updateFreq: updateFreq,
		done:       make(chan bool),
	}
}

// Start begins the periodic Rich Presence updates.
func (rpu *RichPresenceUpdater) Start() {
	rpu.wg.Add(1)
	go func() {
		defer rpu.wg.Done()
		ticker := time.NewTicker(rpu.updateFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rpu.updateDiscordPresence()
			case <-rpu.done:
				log.Println("Rich Presence updater received stop signal.")
				return
			}
		}
	}()
	log.Println("Rich Presence updater started.")
}

// Stop signals the updater to stop and waits for the goroutine to finish.
func (rpu *RichPresenceUpdater) Stop() {
	close(rpu.done)
	rpu.wg.Wait()
	log.Println("Rich Presence updater stopped.")
}

// updateDiscordPresence fetches the latest episode details and updates Discord Rich Presence.
func (rpu *RichPresenceUpdater) updateDiscordPresence() {
	rpu.animeMutex.Lock()
	defer rpu.animeMutex.Unlock()

	if len(rpu.anime.Episodes) == 0 {
		log.Println("No episodes available to update Rich Presence.")
		return
	}

	currentEpisode := rpu.anime.Episodes[0] // Assuming the first element is the current episode
	log.Printf("Current Episode in Updater: %+v", currentEpisode)

	animeTitle := rpu.anime.Details.Title.Romaji
	episodeTitle := currentEpisode.Number // e.g., "Black Clover - Episódio 2 - A Promessa dos Meninos"
	combinedTitle := fmt.Sprintf("%s | %s", animeTitle, episodeTitle)

	// Log the update details
	log.Printf("Updating Rich Presence with anime title: %s\n", combinedTitle)

	// Create a copy to avoid mutating the original anime struct
	copiedAnime := *rpu.anime
	copiedAnime.Details.Title.Romaji = combinedTitle

	// Log the data being sent
	log.Printf("Sending Rich Presence: %+v\n", copiedAnime)

	// Update Rich Presence
	err := api.DiscordPresence(discordClientID, copiedAnime, *rpu.isPaused, time.Now().Unix())
	if err != nil {
		log.Println("Error updating Discord Rich Presence:", err)
	} else {
		log.Printf("Discord Rich Presence updated for episode: %s\n", episodeTitle)
	}
}

func main() {
	var animeMutex sync.Mutex

	// Parse flags to get the anime name
	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	// Initial login to Discord Rich Presence
	err = client.Login(discordClientID)
	if err != nil {
		log.Fatalln("Failed to initialize Discord Rich Presence:", err)
	}
	defer client.Logout() // Ensure logout on exit

	log.Printf("Attempting AniList search with title: %s\n", animeName)

	// Search for the anime
	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	// Fetch anime details, including cover image URL
	err = api.FetchAnimeDetails(anime)
	if err != nil {
		log.Println("Failed to fetch anime details:", err)
	}

	// Log existing fields (ensure these fields exist in your api.Anime struct)
	log.Printf("Title: %s, Cover Image URL: %s\n",
		anime.Details.Title.Romaji, anime.ImageURL)

	// Fetch episodes for the anime
	episodes, err := api.GetAnimeEpisodes(anime.URL) // Corrected to use anime.URL
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	// Check if the anime is a series or a movie/OVA
	series, totalEpisodes, err := api.IsSeries(anime.URL) // Corrected to use anime.URL
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
	} else {
		fmt.Println("The selected anime is a movie/OVA. Starting playback directly...")
	}

	// Define a flag to track whether the playback is paused
	isPaused := false

	// Initialize the RichPresenceUpdater
	updater := NewRichPresenceUpdater(anime, &isPaused, &animeMutex, 15*time.Second)
	defer updater.Stop() // Ensure that the updater is stopped on exit

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)

		// Start the Rich Presence updater
		updater.Start()

		for {
			// Select an episode with fuzzy finder
			selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
			if err != nil {
				log.Fatalln(util.ErrorHandler(err))
			}

			selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
			if err != nil {
				log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
			}

			// Update the anime struct with the selected episode
			animeMutex.Lock()
			anime.Episodes = []api.Episode{
				{
					Number: episodeNumberStr,
					Num:    selectedEpisodeNum,
					URL:    selectedEpisodeURL,
				},
			}
			animeMutex.Unlock()

			log.Printf("Selected Episode Updated: %+v", anime.Episodes[0])

			// Fetch additional episode details
			err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime)
			if err != nil {
				log.Printf("Error fetching episode data: %v", err)
			} else {
				log.Printf("Episode Details Updated: %+v", anime.Episodes[0])
			}

			// Get the video URL for the selected episode
			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
			if err != nil {
				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
			}

			// Handle download and play, updating the paused state as necessary
			player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr) // Corrected to use anime.URL

			// The updater will automatically pick up the latest episode details during the next tick

			// Check for next episode input
			var userInput string
			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
			fmt.Scanln(&userInput)
			if userInput == "q" {
				log.Println("Quitting application as per user request.")
				break
			} else if userInput == "n" {
				log.Println("User selected next episode.")
				continue
			} else if userInput == "p" {
				log.Println("User selected previous episode.")
				continue
			} else {
				log.Println("Invalid input, continuing current episode.")
			}
		}

	} else {
		// Update the anime struct with the first episode (movie/OVA)
		animeMutex.Lock()
		anime.Episodes = []api.Episode{
			episodes[0],
		}
		animeMutex.Unlock()
		log.Printf("Selected Episode Updated: %+v", anime.Episodes[0])

		// Fetch additional episode details
		err = api.GetEpisodeData(anime.MalID, 1, anime)
		if err != nil {
			log.Printf("Error fetching episode data: %v", err)
		} else {
			log.Printf("Episode Details Updated: %+v", anime.Episodes[0])
		}

		// Start the Rich Presence updater
		updater.Start()

		// Get the video URL for the movie/OVA
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		// Handle download and play, updating the paused state as necessary
		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number) // Corrected to use anime.URL
	}

	// The deferred updater.Stop() will ensure that the updater is stopped before exiting
}

// aqui funciona sem autalizar a cada 5sec

// package main

// import (
// 	"fmt"
// 	"log"
// 	"strconv"

// 	"github.com/alvarorichard/Goanime/internal/api"
// 	"github.com/alvarorichard/Goanime/internal/player"
// 	"github.com/alvarorichard/Goanime/internal/util"
// 	"github.com/hugolgst/rich-go/client"
// )

// const discordClientID = "1302721937717334128" // Seu Client ID do Discord

// // Function to update Discord Rich Presence only once
// func updateDiscordPresence(anime *api.Anime, isPaused bool) {
// 	// Update the Discord presence
// 	err := api.DiscordPresence(discordClientID, *anime, isPaused)
// 	if err != nil {
// 		log.Println("Erro ao atualizar o Discord Rich Presence:", err)
// 	}
// }

// func main() {
// 	// Parse flags to get the anime name
// 	animeName, err := util.FlagParser()
// 	if err != nil {
// 		log.Fatalln(util.ErrorHandler(err))
// 	}

// 	// Initialize Discord Rich Presence
// 	err = client.Login(discordClientID)
// 	if err != nil {
// 		log.Fatalln("Falha ao inicializar Discord Rich Presence:", err)
// 	}
// 	defer client.Logout() // Ensure logout on exit

// 	// Search for the anime
// 	anime, err := api.SearchAnime(animeName)
// 	if err != nil {
// 		log.Fatalln("Falha ao buscar anime:", util.ErrorHandler(err))
// 	}

// 	// Fetch anime details, including cover image URL
// 	err = api.FetchAnimeDetails(anime)
// 	if err != nil {
// 		log.Println("Falha ao buscar detalhes do anime:", err)
// 	}

// 	// Fetch episodes for the anime
// 	episodes, err := api.GetAnimeEpisodes(anime.URL)
// 	if err != nil || len(episodes) == 0 {
// 		log.Fatalln("O anime selecionado não possui episódios no servidor.")
// 	}

// 	// Check if the anime is a series or a movie/OVA
// 	series, totalEpisodes, err := api.IsSeries(anime.URL)
// 	if err != nil {
// 		log.Fatalln("Erro ao verificar se o anime é uma série:", util.ErrorHandler(err))
// 	}

// 	// Define a flag to track whether the playback is paused
// 	isPaused := false

// 	if series {
// 		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)

// 		// Select an episode with fuzzy finder
// 		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
// 		if err != nil {
// 			log.Fatalln(util.ErrorHandler(err))
// 		}

// 		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
// 		if err != nil {
// 			log.Fatalln("Erro ao converter número do episódio:", util.ErrorHandler(err))
// 		}

// 		// Update the anime struct with the selected episode
// 		anime.Episodes = []api.Episode{
// 			{
// 				Number: episodeNumberStr,
// 				Num:    selectedEpisodeNum,
// 				URL:    selectedEpisodeURL,
// 			},
// 		}

// 		// Update Discord presence only once
// 		updateDiscordPresence(anime, isPaused)

// 		// Get the video URL for the selected episode
// 		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
// 		if err != nil {
// 			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
// 		}

// 		// Handle download and play, updating the paused state as necessary
// 		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)

// 	} else {
// 		fmt.Println("O anime selecionado é um filme/OVA. Iniciando reprodução direta...")

// 		// Update the anime struct with the first episode (movie/OVA)
// 		anime.Episodes = []api.Episode{
// 			episodes[0],
// 		}

// 		// Update Discord presence only once
// 		updateDiscordPresence(anime, isPaused)

// 		// Get the video URL for the movie/OVA
// 		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
// 		if err != nil {
// 			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
// 		}

// 		// Handle download and play, updating the paused state as necessary
// 		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
// 	}
// }

// package main

// import (
// 	"fmt"
// 	"log"
// 	"strconv"

// 	"github.com/alvarorichard/Goanime/internal/api"
// 	"github.com/alvarorichard/Goanime/internal/player"
// 	"github.com/alvarorichard/Goanime/internal/util"
// 	"github.com/hugolgst/rich-go/client"
// )

// const discordClientID = "1302721937717334128" // Seu Client ID do Discord

// // Function to update Discord Rich Presence once in the background
// func updateDiscordPresenceOnce(anime *api.Anime, isPaused bool) {
// 	// Start a goroutine to update Discord Rich Presence
// 	go func() {
// 		err := api.DiscordPresence(discordClientID, *anime, isPaused)
// 		if err != nil {
// 			log.Println("Erro ao atualizar o Discord Rich Presence:", err)
// 		}
// 	}()
// }

// func main() {
// 	// Parse flags to get the anime name
// 	animeName, err := util.FlagParser()
// 	if err != nil {
// 		log.Fatalln(util.ErrorHandler(err))
// 	}

// 	// Initialize Discord Rich Presence
// 	err = client.Login(discordClientID)
// 	if err != nil {
// 		log.Fatalln("Falha ao inicializar Discord Rich Presence:", err)
// 	}
// 	defer client.Logout() // Ensure logout on exit

// 	// Search for the anime
// 	anime, err := api.SearchAnime(animeName)
// 	if err != nil {
// 		log.Fatalln("Falha ao buscar anime:", util.ErrorHandler(err))
// 	}

// 	// Fetch anime details, including cover image URL
// 	err = api.FetchAnimeDetails(anime)
// 	if err != nil {
// 		log.Println("Falha ao buscar detalhes do anime:", err)
// 	}

// 	// Fetch episodes for the anime
// 	episodes, err := api.GetAnimeEpisodes(anime.URL)
// 	if err != nil || len(episodes) == 0 {
// 		log.Fatalln("O anime selecionado não possui episódios no servidor.")
// 	}

// 	// Check if the anime is a series or a movie/OVA
// 	series, totalEpisodes, err := api.IsSeries(anime.URL)
// 	if err != nil {
// 		log.Fatalln("Erro ao verificar se o anime é uma série:", util.ErrorHandler(err))
// 	}

// 	// Define a flag to track whether the playback is paused
// 	isPaused := false

// 	if series {
// 		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)

// 		// Select an episode with fuzzy finder
// 		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
// 		if err != nil {
// 			log.Fatalln(util.ErrorHandler(err))
// 		}

// 		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
// 		if err != nil {
// 			log.Fatalln("Erro ao converter número do episódio:", util.ErrorHandler(err))
// 		}

// 		// Update the anime struct with the selected episode
// 		anime.Episodes = []api.Episode{
// 			{
// 				Number: episodeNumberStr,
// 				Num:    selectedEpisodeNum,
// 				URL:    selectedEpisodeURL,
// 			},
// 		}

// 		// Start background update for Discord presence
// 		updateDiscordPresenceOnce(anime, isPaused)

// 		// Get the video URL for the selected episode
// 		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
// 		if err != nil {
// 			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
// 		}

// 		// Handle download and play, updating the paused state as necessary
// 		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)

// 	} else {
// 		fmt.Println("O anime selecionado é um filme/OVA. Iniciando reprodução direta...")

// 		// Update the anime struct with the first episode (movie/OVA)
// 		anime.Episodes = []api.Episode{
// 			episodes[0],
// 		}

// 		// Start background update for Discord presence
// 		updateDiscordPresenceOnce(anime, isPaused)

// 		// Get the video URL for the movie/OVA
// 		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
// 		if err != nil {
// 			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
// 		}

// 		// Handle download and play, updating the paused state as necessary
// 		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
// 	}
// }
