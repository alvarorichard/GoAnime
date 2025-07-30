package playback

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/models"
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
			huh.NewOption("Change anime", "c"),
			huh.NewOption("Quit", "q"),
		).
		Value(&choice)

	if err := menu.Run(); err != nil {
		util.Errorf("Error showing menu: %v", err)
		return "n" // Default to next episode on error
	}

	return choice
}

// ChangeAnime allows the user to search for and select a new anime
func ChangeAnime() (*models.Anime, []models.Episode, error) {
	var animeName string

	prompt := huh.NewInput().
		Title("Change Anime").
		Description("Enter the name of the anime you want to watch:").
		Value(&animeName)

	if err := prompt.Run(); err != nil {
		return nil, nil, err
	}

	if len(animeName) < 2 {
		util.Errorf("Anime name too short")
		return nil, nil, fmt.Errorf("anime name must be at least 2 characters")
	}

	// Search for the new anime
	anime := appflow.SearchAnime(animeName)
	if anime == nil {
		return nil, nil, fmt.Errorf("failed to find anime")
	}

	// Fetch anime details
	appflow.FetchAnimeDetails(anime)

	// Get episodes for the new anime
	episodes := appflow.GetAnimeEpisodes(anime.URL)

	return anime, episodes, nil
}
