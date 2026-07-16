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

// findMenuFunc is a package-level indirection over tui.Find so tests can
// drive GetUserInput without opening a TUI.
var findMenuFunc = func(items []menuItem, itemFunc func(i int) string, opts ...fuzzyfinder.Option) (int, error) {
	return tui.Find(items, itemFunc, opts...)
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

	idx, err := findMenuFunc(items, func(i int) string {
		return items[i].Label
	}, fuzzyfinder.WithPromptString("What would you like to do next? "))
	if err != nil {
		// A broken or aborted menu must not auto-advance: returning "n" here
		// made HandleSeries/HandleMovie auto-play forever on non-TTY
		// terminals and turned Esc into "next episode".
		util.Errorf("Error showing menu: %v", err)
		return "q"
	}

	return items[idx].Value
}
