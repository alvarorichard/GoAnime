package api

// Season selection UI for SuperFlix/FlixHQ-style seasoned titles.
//
// The picker uses the shared Bubble Tea screen from internal/tui (tui.Pick):
//
//   - single-season titles skip the full-screen picker entirely;
//   - seasons are listed ascending top-down with instant fuzzy filtering;
//   - the previously watched season (media.CurrentSeason) is preselected
//     when the user re-enters the picker;
//   - each row carries the episode count and air-year range in the details line.
//
// tui.ErrPickBack / tui.ErrPickCancelled are returned unwrapped so the
// caller can map them to ErrBackToSearch.

import (
	"fmt"
	"strconv"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
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

// airYear extracts the 4-digit year prefix from an ISO-ish air date
// ("2019-04-01" → "2019"). Returns "" when the date is missing or malformed.
func airYear(d string) string {
	if len(d) < 4 {
		return ""
	}
	y := d[:4]
	for _, r := range y {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return y
}

// seasonYearRange summarizes a season's air years as "2019" or "2019-2021".
// Episodes without a usable air date are ignored; all-unknown yields "".
func seasonYearRange(eps []superflix.SuperFlixEpisode) string {
	first, last := "", ""
	for _, ep := range eps {
		y := airYear(ep.AirDate)
		if y == "" {
			continue
		}
		if first == "" || y < first {
			first = y
		}
		if last == "" || y > last {
			last = y
		}
	}
	switch first {
	case "":
		return ""
	case last:
		return first
	default:
		return first + "-" + last
	}
}

// seasonPickItems maps ascending season keys to picker rows.
func seasonPickItems(seasonNums []string, allEpisodes map[string][]superflix.SuperFlixEpisode) []tui.PickItem {
	items := make([]tui.PickItem, len(seasonNums))
	for i, sn := range seasonNums {
		details := episodeCountLabel(len(allEpisodes[sn]))
		if yr := seasonYearRange(allEpisodes[sn]); yr != "" {
			details += "  •  " + yr
		}
		items[i] = tui.PickItem{
			Label:   seasonDisplayName(sn),
			Details: details,
		}
	}
	return items
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

// selectSuperFlixSeason asks the user to pick a season and returns its key.
// Single-season titles are selected automatically without opening the picker.
// tui.ErrPickBack / tui.ErrPickCancelled are returned unwrapped so the caller
// can map them to ErrBackToSearch.
func selectSuperFlixSeason(media *models.Anime, seasonNums []string, allEpisodes map[string][]superflix.SuperFlixEpisode) (string, error) {
	return selectSuperFlixSeasonWith(tui.Pick, media, seasonNums, allEpisodes)
}

// selectSuperFlixSeasonWith isolates picker execution for deterministic tests.
func selectSuperFlixSeasonWith(pick seasonPickFunc, media *models.Anime, seasonNums []string, allEpisodes map[string][]superflix.SuperFlixEpisode) (string, error) {
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
	idx, err := pick(seasonPickItems(seasonNums, allEpisodes), tui.PickOptions{
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
