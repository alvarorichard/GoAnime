package main

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/alvarorichard/Goanime/internal/api"
)

func main() {
	a := app.New()
	w := a.NewWindow("GoAnime GUI")

	// UI Elements
	label := widget.NewLabel("Digite o nome do anime:")
	entry := widget.NewEntry()
	resultLabel := widget.NewLabel("Selecione um anime:")
	episodeLabel := widget.NewLabel("Selecione um episódio:")

	// Lists for anime and episodes
	var animes []api.GUIAnime
	var episodes []api.Episode

	animeList := widget.NewList(
		func() int { return len(animes) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(animes[i].Name)
		},
	)

	episodeList := widget.NewList(
		func() int { return len(episodes) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(fmt.Sprintf("Episódio %s", episodes[i].Number))
		},
	)

	button := widget.NewButton("Buscar", func() {
		animeName := strings.TrimSpace(entry.Text)
		if animeName == "" {
			resultLabel.SetText("Digite um nome válido!")
			return
		}

		// ✅ Fetch anime results
		foundAnimes, err := api.SearchAnimeGUI(animeName)
		if err != nil || len(foundAnimes) == 0 {
			resultLabel.SetText("Nenhum anime encontrado!")
			log.Println("Erro:", err)
			return
		}

		// ✅ Update anime list
		animes = foundAnimes
		animeList.Refresh()
		resultLabel.SetText("Selecione um anime:")
	})

	// ✅ Handle anime selection to fetch episodes
	animeList.OnSelected = func(id widget.ListItemID) {
		selectedAnime := animes[id]
		resultLabel.SetText(fmt.Sprintf("Selecionado: %s", selectedAnime.Name))

		// ✅ Fetch episodes for the selected anime
		foundEpisodes, err := api.GetAnimeEpisodes(selectedAnime.URL)
		if err != nil || len(foundEpisodes) == 0 {
			episodeLabel.SetText("Nenhum episódio encontrado!")
			log.Println("Erro ao buscar episódios:", err)
			episodes = nil // Ensure empty list
			episodeList.Refresh()
			return
		}

		// ✅ Update episode list dynamically
		episodes = foundEpisodes
		episodeList.Length = func() int { return len(episodes) }
		episodeList.Refresh()
		episodeLabel.SetText("Selecione um episódio:")
	}

	// ✅ Handle episode selection (copy URL to clipboard)
	episodeList.OnSelected = func(id widget.ListItemID) {
		selectedEpisode := episodes[id]
		episodeLabel.SetText(fmt.Sprintf("Episódio %s copiado para a área de transferência!", selectedEpisode.Number))

		// Copy episode URL to clipboard
		w.Clipboard().SetContent(selectedEpisode.URL)
		log.Println("📋 Copied URL:", selectedEpisode.URL)
	}

	// Layout
	content := container.NewVBox(
		label, entry, button,
		resultLabel, animeList,
		episodeLabel, episodeList,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 600))
	w.ShowAndRun()
}
