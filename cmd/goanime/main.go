//package main
//
//import (
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
//	// Initialize Discord Rich Presence
//	discordEnabled := true
//	if err := client.Login(discordClientID); err != nil {
//		if util.IsDebug {
//			log.Println("Failed to initialize Discord Rich Presence:", err)
//
//		}
//		discordEnabled = false
//	} else {
//		defer client.Logout() // Ensure logout on exit
//	}
//
//	// Search for the anime
//	anime, err := api.SearchAnime(animeName)
//	if err != nil {
//		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
//	}
//
//	// Fetch anime details, including cover image URL
//	if err = api.FetchAnimeDetails(anime); err != nil {
//		log.Println("Failed to fetch anime details:", err)
//	}
//
//	// Fetch episodes for the anime
//	episodes, err := api.GetAnimeEpisodes(anime.URL)
//	if err != nil || len(episodes) == 0 {
//		log.Fatalln("The selected anime does not have episodes on the server.")
//	}
//
//	// Check if the anime is a series or a movie/OVA
//	series, totalEpisodes, err := api.IsSeries(anime.URL)
//	if err != nil {
//		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
//	}
//
//	// Define a flag to track if the playback is paused
//	isPaused := false
//	socketPath := "/tmp/mpvsocket" // Adjust socket path as per your setup
//	updateFreq := 1 * time.Second  // Update frequency for Rich Presence
//	episodeDuration := time.Duration(episodes[0].Duration) * time.Second
//
//	if series {
//		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
//
//		for {
//			// Select an episode using fuzzy finder
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
//			// Lock anime struct and update with selected episode
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
//			// Fetch episode details and AniSkip data
//			if err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime); err != nil {
//				log.Printf("Error fetching episode data: %v", err)
//			}
//
//			// Retrieve video URL for the selected episode
//			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
//			if err != nil {
//				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//			}
//
//			// Initialize a new RichPresenceUpdater for this episode if Discord is enabled
//			var updater *player.RichPresenceUpdater
//			if discordEnabled {
//				updater = player.NewRichPresenceUpdater(
//					anime,
//					&isPaused,
//					&animeMutex,
//					updateFreq,
//					episodeDuration,
//					socketPath,
//				)
//				defer updater.Stop() // Ensure updater is stopped when done
//			} else {
//				updater = nil
//			}
//
//			// Handle download and playback, updating paused state as necessary
//			player.HandleDownloadAndPlay(
//				videoURL,
//				episodes,
//				selectedEpisodeNum,
//				anime.URL,
//				episodeNumberStr,
//				anime.MalID, // Pass the animeMalID here
//				updater,
//			)
//
//			// Prompt user for next action
//			var userInput string
//			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
//			fmt.Scanln(&userInput)
//			if userInput == "q" {
//				log.Println("Quitting application as per user request.")
//				break
//			} else if userInput == "n" || userInput == "p" {
//				continue // loop continues for next or previous episode
//			} else {
//				log.Println("Invalid input, continuing current episode.")
//			}
//		}
//
//	} else {
//		// Handle movie/OVA playback
//
//		// Lock anime struct and update with the first episode
//		animeMutex.Lock()
//		anime.Episodes = []api.Episode{episodes[0]}
//		animeMutex.Unlock()
//
//		// Fetch details and AniSkip data for the movie/OVA
//		if err = api.GetMovieData(anime.MalID, anime); err != nil {
//			log.Printf("Error fetching movie/OVA data: %v", err)
//		}
//
//		// Get the video URL for the movie/OVA
//		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		// Initialize a new RichPresenceUpdater for the movie if Discord is enabled
//		var updater *player.RichPresenceUpdater
//		if discordEnabled {
//			updater = player.NewRichPresenceUpdater(
//				anime,
//				&isPaused,
//				&animeMutex,
//				updateFreq,
//				episodeDuration,
//				socketPath,
//			)
//			defer updater.Stop()
//		} else {
//			updater = nil
//		}
//
//		// Handle download and play, with Rich Presence updates
//		player.HandleDownloadAndPlay(
//			videoURL,
//			episodes,
//			1, // Episode number for movies/OVAs
//			anime.URL,
//			episodes[0].Number,
//			anime.MalID, // Pass the animeMalID here
//			updater,
//		)
//	}
//
//	// No need to call updater.Stop() here as it's deferred after each initialization
//}

package main

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1302721937717334128"

var (
	statusLabel     *widget.Label
	animeList       *widget.List
	episodeList     *widget.List
	searchEntry     *widget.Entry
	playButton      *widget.Button
	contentArea     *fyne.Container
	currentAnime    *api.Anime
	currentEpisodes []api.Episode
	selectedEpisode int = -1
	discordEnabled  bool
	animeMutex      sync.Mutex
	animes          []api.Anime
)

func updateStatus(message string) {
	statusLabel.SetText(message)
}

func main() {
	// Inicializa o Rich Presence do Discord
	if err := client.Login(discordClientID); err != nil {
		discordEnabled = false
	} else {
		discordEnabled = true
		defer client.Logout()
	}

	// Inicializa o app Fyne
	myApp := app.New()
	mainWindow := myApp.NewWindow("GoAnime")
	mainWindow.Resize(fyne.NewSize(800, 600))

	// Widgets principais
	statusLabel = widget.NewLabel("Digite o nome do anime e clique em 'Pesquisar'.")
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Digite o nome do anime...")
	playButton = widget.NewButton("Reproduzir/Download", func() {
		showDownloadOptions(mainWindow)
	})
	playButton.Disable()

	contentArea = container.NewMax()

	// Layout da interface
	mainContainer := container.NewBorder(
		container.NewVBox(
			searchEntry,
			widget.NewButton("Pesquisar", func() {
				searchAnime()
			}),
		),
		container.NewVBox(statusLabel, playButton),
		nil, nil,
		contentArea,
	)

	mainWindow.SetContent(mainContainer)
	mainWindow.ShowAndRun()
}

// 🔍 Pesquisa animes e exibe na GUI
func searchAnime() {
	query := searchEntry.Text
	if query == "" {
		updateStatus("Por favor, insira um nome de anime.")
		return
	}

	updateStatus("Pesquisando...")

	go func() {
		result, err := api.SearchAnime(query)
		if err != nil {
			updateStatus("Erro na busca: " + err.Error())
			return
		}

		animes = []api.Anime{*result}
		if len(animes) == 0 {
			updateStatus("Nenhum resultado encontrado.")
			return
		}

		// Criando a lista de animes na GUI
		animeList = widget.NewList(
			func() int { return len(animes) },
			func() fyne.CanvasObject { return widget.NewLabel("Anime") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(animes[id].Name)
			},
		)

		animeList.OnSelected = func(id widget.ListItemID) {
			currentAnime = &animes[id]
			updateStatus("Carregando episódios para: " + currentAnime.Name)
			loadEpisodes(currentAnime)
		}

		contentArea.Objects = []fyne.CanvasObject{animeList}
		contentArea.Refresh()
		updateStatus(fmt.Sprintf("Encontrados %d resultados.", len(animes)))
	}()
}

// 🎥 Carrega a lista de episódios do anime selecionado
func loadEpisodes(anime *api.Anime) {
	go func() {
		episodes, err := api.GetAnimeEpisodes(anime.URL)
		if err != nil || len(episodes) == 0 {
			updateStatus("Erro ao carregar episódios.")
			return
		}

		currentEpisodes = episodes
		selectedEpisode = -1
		playButton.Disable()

		// Criando a lista de episódios na GUI
		episodeList = widget.NewList(
			func() int { return len(episodes) },
			func() fyne.CanvasObject { return widget.NewLabel("Episódio") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(fmt.Sprintf("Episódio %d", id+1))
			},
		)

		episodeList.OnSelected = func(id widget.ListItemID) {
			selectedEpisode = id
			playButton.Enable()
		}

		contentArea.Objects = []fyne.CanvasObject{episodeList}
		contentArea.Refresh()
		updateStatus(fmt.Sprintf("Carregados %d episódios.", len(episodes)))
	}()
}

// 📥 Exibe as opções de download/reprodução
func showDownloadOptions(win fyne.Window) {
	if currentAnime == nil || selectedEpisode < 0 || selectedEpisode >= len(currentEpisodes) {
		updateStatus("Seleção inválida.")
		return
	}

	options := []string{"Assistir online", "Baixar este episódio", "Baixar um intervalo de episódios"}
	radio := widget.NewRadioGroup(options, func(selected string) {
		switch selected {
		case "Assistir online":
			playEpisode()
		case "Baixar este episódio":
			downloadEpisode(selectedEpisode)
		case "Baixar um intervalo de episódios":
			showDownloadRangeDialog(win)
		}
	})

	dialog.ShowCustom("Escolha uma opção", "OK", container.NewVBox(radio), win)
}

// 🆕 ⬇️ Exibe um diálogo para baixar múltiplos episódios
func showDownloadRangeDialog(win fyne.Window) {
	startEntry := widget.NewEntry()
	startEntry.SetPlaceHolder("Episódio inicial")

	endEntry := widget.NewEntry()
	endEntry.SetPlaceHolder("Episódio final")

	form := widget.NewForm(
		widget.NewFormItem("Início", startEntry),
		widget.NewFormItem("Fim", endEntry),
	)

	dialog.ShowCustomConfirm("Baixar Intervalo", "Baixar", "Cancelar", form, func(confirm bool) {
		if !confirm {
			return
		}

		start, err1 := strconv.Atoi(startEntry.Text)
		end, err2 := strconv.Atoi(endEntry.Text)
		if err1 != nil || err2 != nil || start < 1 || end > len(currentEpisodes) || start > end {
			updateStatus("Intervalo inválido.")
			return
		}

		for i := start - 1; i < end; i++ {
			downloadEpisode(i)
		}
	}, win)
}

// ⬇️ Baixar um único episódio
func downloadEpisode(epIndex int) {
	ep := currentEpisodes[epIndex]
	updateStatus(fmt.Sprintf("Baixando Episódio %d...", epIndex+1))

	go func() {
		videoURL, err := player.GetVideoURLForEpisode(ep.URL)
		if err != nil {
			updateStatus("Erro ao obter URL do vídeo para download.")
			return
		}

		err = player.DownloadVideo(videoURL, fmt.Sprintf("episodio_%d.mp4", epIndex+1), 0, nil)
		if err != nil {
			updateStatus("Erro no download: " + err.Error())
		} else {
			updateStatus(fmt.Sprintf("Episódio %d baixado com sucesso!", epIndex+1))
		}
	}()
}

// ▶️ Reproduz o episódio selecionado
func playEpisode() {
	ep := currentEpisodes[selectedEpisode]
	updateStatus(fmt.Sprintf("Reproduzindo Episódio %d...", selectedEpisode+1))

	go func() {
		videoURL, err := player.GetVideoURLForEpisode(ep.URL)
		if err != nil {
			updateStatus("Erro ao obter URL do vídeo.")
			return
		}

		go player.HandleDownloadAndPlay(videoURL, []api.Episode{ep}, selectedEpisode+1, currentAnime.URL, ep.Number, currentAnime.MalID, nil)
		updateStatus(fmt.Sprintf("Reproduzindo Episódio %d...", selectedEpisode+1))
	}()
}
