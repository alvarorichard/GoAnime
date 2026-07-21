package playback

import (
	"errors"

	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// menuItem maps a display label to the short code returned by GetUserInput.
type menuItem struct {
	Label string
	Value string
}

// findMenuFunc is a package-level indirection over the fancy picker so tests
// can drive GetUserInput without opening a TTY.
var findMenuFunc = func(items []menuItem) (int, error) {
	labels := make([]string, len(items))
	for i, it := range items {
		labels[i] = it.Label
	}
	return tui.PickLabels(labels, tui.PickOptions{
		Breadcrumb:   "Playback > Next",
		WindowTitle:  "GoAnime - Menu",
		ItemSingular: "option",
		ItemPlural:   "options",
	})
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

	idx, err := findMenuFunc(items)
	if err != nil {
		// Esc/back from the fancy picker is an intentional exit path.
		if errors.Is(err, tui.ErrPickBack) {
			return "back"
		}
		// A broken menu must not auto-advance: returning "n" here made
		// HandleSeries/HandleMovie auto-play forever on non-TTY terminals.
		util.Errorf("Error showing menu: %v", err)
		return "q"
	}
	if idx < 0 || idx >= len(items) {
		return "q"
	}
	return items[idx].Value
}
