package playback

import (
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
)

func GetUserInput() string {
	var choice string

	menu := huh.NewSelect[string]().
		Title("Playback Control").
		Description("What would you like to do next?").
		Options(
			huh.NewOption("Next episode", "n"),
			huh.NewOption("Previous episode", "p"),
			huh.NewOption("Select episode", "e"),
			huh.NewOption("Quit", "q"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing menu: %v", err)
		return "n" // Default to next episode on error
	}

	return choice
}
