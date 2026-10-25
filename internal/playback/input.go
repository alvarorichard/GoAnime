package playback

import (
	"errors"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// menuItem maps a display label to the short code returned by GetUserInput,
// and to the key that picks it at once.
type menuItem struct {
	Label string
	Value string
	Key   string
}

// findMenuFunc is a package-level indirection over the fancy picker so tests
// can drive GetUserInput without opening a TTY.
var findMenuFunc = func(items []menuItem) (int, error) {
	pickItems := make([]tui.PickItem, len(items))
	for i, it := range items {
		pickItems[i] = tui.PickItem{Label: it.Label, Key: it.Key}
	}
	return tui.Pick(pickItems, tui.PickOptions{
		Breadcrumb:   "Playback > Next",
		WindowTitle:  "GoAnime - Menu",
		ItemSingular: "option",
		ItemPlural:   "options",
		Action:       "choose",
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
			{"Replay movie", "n", "r"},
			{"Change movie", "c", "c"},
			{"← Back", "back", "esc"},
			{"Quit", "q", "q"},
		}
	} else {
		// TV series / anime: full episode navigation
		items = []menuItem{
			{"Next episode", "n", "n"},
			{"Previous episode", "p", "p"},
			{"Select episode", "e", "e"},
			{"Change anime", "c", "c"},
			{"← Back", "back", "esc"},
			{"Quit", "q", "q"},
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
