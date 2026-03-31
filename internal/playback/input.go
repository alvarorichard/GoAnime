package playback

import (
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// menuItem maps a display label to the short code returned by GetUserInput.
type menuItem struct {
	Label string
	Value string
}

// GetUserInput shows post-playback menu. Pass isMovie=true for movies to show
// a simplified menu without episode navigation options.
func GetUserInput(isMovie ...bool) string {
	movie := len(isMovie) > 0 && isMovie[0]

	var items []menuItem
	if movie {
		// Movie: no episode navigation
		items = []menuItem{
			{"Replay movie", "n"},
			{"Change movie", "c"},
			{"← Back", "back"},
			{"Quit", "q"},
		}
	} else {
		// TV series / anime: full episode navigation
		items = []menuItem{
			{"Next episode", "n"},
			{"Previous episode", "p"},
			{"Select episode", "e"},
			{"Change anime", "c"},
			{"← Back", "back"},
			{"Quit", "q"},
		}
	}

	idx, err := tui.Find(items, func(i int) string {
		return items[i].Label
	}, fuzzyfinder.WithPromptString("What would you like to do next? "))
	if err != nil {
		util.Errorf("Error showing menu: %v", err)
		return "n" // Default to next episode on error
	}

	return items[idx].Value
}
