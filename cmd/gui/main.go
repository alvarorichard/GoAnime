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
	resultLabel := widget.NewLabel("")

	// List for displaying anime search results
	var animes []api.GUIAnime
	list := widget.NewList(
		func() int { return len(animes) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(animes[i].Name)
		},
	)

	button := widget.NewButton("Buscar", func() {
		animeName := strings.TrimSpace(entry.Text)
		if animeName == "" {
			resultLabel.SetText("Digite um nome válido!")
			return
		}

		// ✅ **Call the GUI-optimized search function**
		foundAnimes, err := api.SearchAnimeGUI(animeName)
		if err != nil || len(foundAnimes) == 0 {
			resultLabel.SetText("Nenhum anime encontrado!")
			log.Println("Erro:", err)
			return
		}

		// ✅ **Update list with real results**
		animes = foundAnimes
		list.Refresh()

		// Handle selection
		list.OnSelected = func(id widget.ListItemID) {
			selectedAnime := animes[id]
			resultLabel.SetText(fmt.Sprintf("Nome: %s\nURL: %s",
				selectedAnime.Name, selectedAnime.URL))
		}
	})

	// Layout
	content := container.NewVBox(
		label,
		entry,
		button,
		list,
		resultLabel,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 500))
	w.ShowAndRun()
}
