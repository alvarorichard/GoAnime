package api

// Season selection UI for seasoned titles (StartFlix).
//
// The picker uses the shared Bubble Tea screen from internal/tui (tui.Pick):
//
//   - single-season titles skip the full-screen picker entirely;
//   - seasons are listed ascending top-down with instant fuzzy filtering;
//   - the previously watched season (media.CurrentSeason) is preselected
//     when the user re-enters the picker;
//   - each row carries its episode counts per audio (Dublado/Legendado).
//
// tui.ErrPickBack / tui.ErrPickCancelled are returned unwrapped so the
// caller can map them to ErrBackToSearch.

import (
	"fmt"
	"strconv"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// seasonDisplayName renders a season key for humans: "0" is the TVDB/TMDB
// convention for specials, numeric keys become "Season N", and non-numeric
// keys (TVmaze year-based "seasons") pass through as-is.
func seasonDisplayName(sn string) string {
	if n, err := strconv.Atoi(sn); err == nil {
		if n == 0 {
			return "Specials"
		}
		return fmt.Sprintf("Season %d", n)
	}
	return "Season " + sn
}

// episodeCountLabel pluralizes the episode count.
func episodeCountLabel(n int) string {
	if n == 1 {
		return "1 episode"
	}
	return fmt.Sprintf("%d episodes", n)
}

// currentSeasonIndex locates media.CurrentSeason in the season keys, falling
// back to the first season when unset or absent.
func currentSeasonIndex(media *models.Anime, seasonNums []string) int {
	if media == nil || media.CurrentSeason <= 0 {
		return 0
	}
	current := strconv.Itoa(media.CurrentSeason)
	for i, sn := range seasonNums {
		if sn == current {
			return i
		}
	}
	return 0
}

// seasonPickFunc matches tui.Pick so tests can inject a headless picker.
type seasonPickFunc func([]tui.PickItem, tui.PickOptions) (int, error)

// selectSeasonWith is the source-independent picker: seasonNums are the keys in
// display order, and items builds their rows (only when a picker is shown).
func selectSeasonWith(pick seasonPickFunc, media *models.Anime, seasonNums []string, items func() []tui.PickItem) (string, error) {
	if len(seasonNums) == 1 {
		only := seasonNums[0]
		util.Infof("Only one season available — %s selected automatically.", seasonDisplayName(only))
		return only, nil
	}

	if pick == nil {
		return "", fmt.Errorf("season picker not configured")
	}
	if len(seasonNums) == 0 {
		return "", fmt.Errorf("no seasons to select")
	}

	name := ""
	if media != nil {
		name = tui.SingleLine(media.Name)
	}
	if name == "" {
		name = "Title"
	}
	idx, err := pick(items(), tui.PickOptions{
		Breadcrumb:   fmt.Sprintf("Search > %s > Seasons", name),
		WindowTitle:  "GoAnime - Seasons",
		ItemSingular: "season",
		ItemPlural:   "seasons",
		InitialIndex: currentSeasonIndex(media, seasonNums),
	})
	if err != nil {
		return "", err
	}
	if idx < 0 || idx >= len(seasonNums) {
		return "", fmt.Errorf("season picker returned invalid index %d", idx)
	}
	return seasonNums[idx], nil
}
