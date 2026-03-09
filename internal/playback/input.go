package playback

import (
	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/util"
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
			huh.NewOption("Change anime", "c"),
			huh.NewOption("← Back", "back"),
			huh.NewOption("Quit", "q"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing menu: %v", err)
		return "n" // Default to next episode on error
	}

	return choice
}
