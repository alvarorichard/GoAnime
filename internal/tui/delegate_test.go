package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderRow draws the item at index of a list holding items, width cells wide.
func renderRow(t *testing.T, d rowDelegate, items []list.Item, index, width int) string {
	t.Helper()
	l := list.New(items, d, width, 20)
	var b strings.Builder
	d.Render(&b, l, index, items[index])
	return b.String()
}

func TestRowDelegate_OneLine(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	d := rowDelegate{theme: &theme, marks: true}
	items := []list.Item{
		pickEntry{index: 0, item: PickItem{Label: "Episode 1", Status: PickWatched}},
		pickEntry{index: 1, item: PickItem{Label: "Episode 2", Status: PickInProgress, Note: "stopped at 14:32"}},
		pickEntry{index: 2, item: PickItem{Label: "Episode 3", Details: "2019-04-01", Tags: []string{"Filler"}}},
		pickEntry{index: 3, item: PickItem{Label: "Next episode", Key: "n"}},
		pickEntry{index: 4, item: PickItem{Label: "1080p", Meter: 1}},
	}

	selected := renderRow(t, d, items, 0, 50)
	assert.Equal(t, " ▶ ✓ Episode 1", strings.TrimRight(ansi.Strip(selected), " "))
	assert.Equal(t, 50, lipgloss.Width(selected), "the cursor row is filled to the list's edge")

	assert.Equal(t, "   ◐ Episode 2", ansi.Strip(renderRow(t, d, items, 1, 50))[:len("   ◐ Episode 2")])
	assert.True(t, strings.HasSuffix(ansi.Strip(renderRow(t, d, items, 1, 50)), "stopped at 14:32 "))

	tagged := ansi.Strip(renderRow(t, d, items, 2, 50))
	assert.Contains(t, tagged, "Episode 3  FILLER ")
	assert.True(t, strings.HasSuffix(tagged, "2019-04-01 "))

	assert.True(t, strings.HasSuffix(ansi.Strip(renderRow(t, d, items, 3, 50)), " n "), "the key sits at the far right")

	meter := ansi.Strip(renderRow(t, d, items, 4, 50))
	assert.Contains(t, meter, strings.Repeat("━", 12))

	for i := range items {
		assert.Equal(t, 50, lipgloss.Width(renderRow(t, d, items, i, 50)), "row %d", i)
	}
}

func TestRowDelegate_NarrowGivesUpTheRightSideFirst(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	d := rowDelegate{theme: &theme}
	items := []list.Item{
		pickEntry{item: PickItem{Label: "A long episode title that needs room", Details: "2019-04-01", Tags: []string{"Recap"}, Key: "x"}},
	}

	got := ansi.Strip(renderRow(t, d, items, 0, 24))
	assert.Equal(t, 24, lipgloss.Width(got))
	assert.NotContains(t, got, "2019")
	assert.NotContains(t, got, "RECAP")
	assert.Contains(t, got, "A long", "the title keeps its room")
	assert.True(t, strings.HasSuffix(got, " x "), "a shortcut stays")
}

func TestRowDelegate_TwoLineResults(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	d := rowDelegate{theme: &theme, twoLine: true, spacing: 1}
	items := []list.Item{animeGroupItem{number: "01"}}

	got := renderRow(t, d, items, 0, 40)
	lines := strings.Split(ansi.Strip(got), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, " ▶ 01  Unknown title", strings.TrimRight(lines[0], " "))
	assert.Equal(t, "       Information unavailable", strings.TrimRight(lines[1], " "))
	assert.Equal(t, 2, d.Height())
	assert.Equal(t, 1, d.Spacing())
}

func TestRowDelegate_FilterMatchesAreStyled(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	items := []list.Item{pickEntry{item: PickItem{Label: "Mikanbune"}}, pickEntry{index: 1, item: PickItem{Label: "Kuro Nyago"}}}
	l := list.New(items, rowDelegate{theme: &theme}, 40, 20)
	l.SetFilterText("kan")
	require.Len(t, l.VisibleItems(), 1)

	var b strings.Builder
	rowDelegate{theme: &theme}.Render(&b, l, 0, l.VisibleItems()[0])
	assert.Contains(t, ansi.Strip(b.String()), "Mikanbune")
	assert.Contains(t, b.String(), theme.Selected.Underline(true).Render("kan"), "a match on the cursor row is underlined")
}

func TestStatusMark(t *testing.T) {
	t.Parallel()
	assert.Equal(t, " ", statusMark(PickNone))
	assert.Equal(t, "✓", statusMark(PickWatched))
	assert.Equal(t, "◐", statusMark(PickInProgress))
	assert.Equal(t, "●", statusMark(PickCurrent))
}
