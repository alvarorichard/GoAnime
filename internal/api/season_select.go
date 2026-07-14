package api

// Season selection UI for SuperFlix/FlixHQ-style seasoned titles.
//
// The picker stays on go-fuzzyfinder (a prior huh.Select migration caused a
// rendering regression — see internal/player/download.go), but fixes the
// ergonomics around it:
//
//   - single-season titles skip the full-screen picker entirely;
//   - fuzzyfinder draws its list bottom-up (first item next to the prompt),
//     so the seasons are fed in DESCENDING order to make the list read
//     ascending top-down, with the cursor starting at the top;
//   - the previously watched season (media.CurrentSeason) is preselected
//     when the user re-enters the picker;
//   - each row carries the episode count and air-year range, and a preview
//     pane lists the highlighted season's episodes.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
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

// seasonLabel builds the picker row: "Season 2  ·  12 episodes  ·  2019".
func seasonLabel(sn string, eps []superflix.SuperFlixEpisode) string {
	label := seasonDisplayName(sn) + "  ·  " + episodeCountLabel(len(eps))
	if yr := seasonYearRange(eps); yr != "" {
		label += "  ·  " + yr
	}
	return label
}

// formatEpisodeLine renders one episode for the preview pane:
// "E01  Pilot  (2019-04-01)". Missing title/date segments are dropped.
func formatEpisodeLine(ep superflix.SuperFlixEpisode) string {
	num := ep.EpiNum.String()
	if len(num) == 1 {
		num = "0" + num
	}
	line := "E" + num
	if t := strings.TrimSpace(ep.Title); t != "" {
		line += "  " + t
	}
	if d := strings.TrimSpace(ep.AirDate); d != "" {
		line += "  (" + d + ")"
	}
	return line
}

// seasonPreviewText builds the preview pane content for one season, trimmed
// to the pane height with a "... and N more" tail when the list is cut.
func seasonPreviewText(sn string, eps []superflix.SuperFlixEpisode, height int) string {
	var b strings.Builder
	b.WriteString(seasonDisplayName(sn))
	if yr := seasonYearRange(eps); yr != "" {
		b.WriteString(" (")
		b.WriteString(yr)
		b.WriteString(")")
	}
	b.WriteString(" — ")
	b.WriteString(episodeCountLabel(len(eps)))
	b.WriteString("\n\n")

	// Header + blank line + borders eat ~4 rows of the pane.
	maxLines := max(height-4, 1)
	for i, ep := range eps {
		if i >= maxLines {
			fmt.Fprintf(&b, "  ... and %d more\n", len(eps)-i)
			break
		}
		b.WriteString("  ")
		b.WriteString(formatEpisodeLine(ep))
		b.WriteString("\n")
	}
	return b.String()
}

// seasonDisplayOrder reverses the ascending season keys. fuzzyfinder renders
// its list bottom-up (index 0 sits next to the prompt at the bottom), so a
// descending feed is what makes the visible list read Season 1 → N top-down.
func seasonDisplayOrder(seasonNums []string) []string {
	display := make([]string, len(seasonNums))
	for i, sn := range seasonNums {
		display[len(seasonNums)-1-i] = sn
	}
	return display
}

// selectSuperFlixSeason asks the user to pick a season and returns its key.
// Single-season titles are selected automatically without opening the picker.
// A fuzzyfinder.ErrAbort (ESC) is returned unwrapped so the caller can map it
// to ErrBackToSearch.
func selectSuperFlixSeason(media *models.Anime, seasonNums []string, allEpisodes map[string][]superflix.SuperFlixEpisode) (string, error) {
	if len(seasonNums) == 1 {
		only := seasonNums[0]
		util.Infof("Only one season available — %s selected automatically.", seasonDisplayName(only))
		return only, nil
	}

	display := seasonDisplayOrder(seasonNums)
	labels := make([]string, len(display))
	for i, sn := range display {
		labels[i] = seasonLabel(sn, allEpisodes[sn])
	}

	header := fmt.Sprintf("%s — %d seasons  ·  ↑/↓ move · Enter select · ESC back", media.Name, len(seasonNums))
	current := strconv.Itoa(media.CurrentSeason)

	idx, err := tui.Find(labels, func(i int) string { return labels[i] },
		fuzzyfinder.WithPromptString("Select season: "),
		fuzzyfinder.WithHeader(header),
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionTop),
		fuzzyfinder.WithPreselected(func(i int) bool {
			return media.CurrentSeason > 0 && display[i] == current
		}),
		fuzzyfinder.WithPreviewWindow(func(i, _, h int) string {
			if i < 0 || i >= len(display) {
				return ""
			}
			return seasonPreviewText(display[i], allEpisodes[display[i]], h)
		}),
	)
	if err != nil {
		return "", err
	}
	return display[idx], nil
}
