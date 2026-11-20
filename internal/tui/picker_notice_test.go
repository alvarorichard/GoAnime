package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The episode list is a picker, and a step that sends the user back to it
// (a removed video, a title with no files) has to say why on the picker
// itself: what is printed to the terminal before it opens is hidden behind it.
// Not parallel where the notice is package-level.

const removedNotice = "Episódio 01 is no longer playable on Animefire.io: the host deleted the video. Press ESC to pick the same title from StartFlix."

func episodeItems(n int) []PickItem {
	items := make([]PickItem, n)
	for i := range items {
		items[i] = PickItem{Label: fmt.Sprintf("Episode %d", i+1)}
	}
	return items
}

func sized(model *pickerModel, width, height int) string {
	_, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return model.View().Content
}

func TestPickerShowsItsNoticeAboveTheList(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 60, Height: 14}, {Width: 80, Height: 12}, {Width: 120, Height: 18}, {Width: 160, Height: 40}} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice, Action: "resume"})

			view := sized(model, size.Width, size.Height)
			text := ansi.Strip(view)

			assert.Contains(t, text, "no longer playable", "the reason is on the screen")
			assert.Contains(t, text, "Episode 1", "and the list is still there")
			assert.Less(t, strings.Index(text, "no longer playable"), strings.Index(text, "Episode 1"), "the notice sits above the list")
			assert.LessOrEqual(t, lipgloss.Height(view), size.Height, "the notice does not push the screen past the terminal")
			assert.LessOrEqual(t, lipgloss.Width(view), size.Width)
		})
	}
}

func TestPickerWithoutANoticeDrawsNothingExtra(t *testing.T) {
	with := ansi.Strip(sized(newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice}), 100, 24))
	without := ansi.Strip(sized(newPickerModel(episodeItems(25), PickOptions{}), 100, 24))

	assert.NotContains(t, without, "no longer playable")
	assert.NotContains(t, without, "[!]")
	assert.NotEqual(t, with, without)
}

// rowsShown is how many "Episode N" rows a rendered picker lists.
func rowsShown(view string) int { return strings.Count(ansi.Strip(view), "Episode ") }

// The notice takes its rows from the list, exactly: the list neither runs under
// it (and loses its last rows to the cut at the terminal's height) nor leaves
// a gap.
func TestPickerNoticeTakesExactlyItsRowsFromTheList(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 16}, {Width: 100, Height: 24}, {Width: 120, Height: 40}} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			with := newPickerModel(episodeItems(60), PickOptions{Notice: removedNotice})
			without := newPickerModel(episodeItems(60), PickOptions{})
			withView := sized(with, size.Width, size.Height)
			withoutView := sized(without, size.Width, size.Height)

			width, _ := with.shell.ContentSize()
			reserved := with.noticeHeight(width)
			require.Positive(t, reserved, "a notice takes at least one row")

			assert.Equal(t, rowsShown(withoutView)-reserved, rowsShown(withView))
			assert.Contains(t, ansi.Strip(withView), "esc", "and the footer is still there")
		})
	}
}

func TestPickerNoticeSurvivesNavigationAndFiltering(t *testing.T) {
	model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice})
	_ = sized(model, 100, 24)

	for _, msg := range []tea.Msg{
		tea.KeyPressMsg{Code: tea.KeyDown},
		tea.KeyPressMsg{Code: tea.KeyDown},
		tea.WindowSizeMsg{Width: 90, Height: 20},
	} {
		_, _ = model.Update(msg)
		assert.Contains(t, ansi.Strip(model.View().Content), "no longer playable", "after %T", msg)
	}
}

func TestPickerNoticeIsOneLineWhateverItHolds(t *testing.T) {
	model := newPickerModel(episodeItems(5), PickOptions{Notice: "Broke\nacross\r\nlines\twith\ttabs"})

	assert.Equal(t, singleLine("Broke\nacross\r\nlines\twith\ttabs"), model.notice)
	assert.NotContains(t, model.notice, "\n")
}

func TestPickerNoticeIsCutRatherThanFlooding(t *testing.T) {
	long := strings.Repeat("The host deleted every video of this title. ", 40)
	model := newPickerModel(episodeItems(25), PickOptions{Notice: long})

	view := sized(model, 80, 20)

	assert.LessOrEqual(t, lipgloss.Height(view), 20)
	assert.Contains(t, ansi.Strip(view), "Episode 1", "even a huge notice leaves the list")
	assert.Contains(t, ansi.Strip(view), "…", "and is cut where it runs out of room")
}

func TestPickerNoticeLeavesAListOnATinyTerminal(t *testing.T) {
	model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice})

	view := sized(model, 40, 8)

	assert.NotPanics(t, func() { _ = model.View() })
	assert.LessOrEqual(t, lipgloss.Width(view), 40)
}

func TestPickerHelpOverlayTakesTheWholeScreen(t *testing.T) {
	model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice, HelpHint: true})
	_ = sized(model, 100, 30)

	_, _ = model.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	require.True(t, model.showHelp)
	view := model.View().Content

	assert.NotContains(t, ansi.Strip(view), "no longer playable", "the notice is for the list, not the keys overlay")
	assert.LessOrEqual(t, lipgloss.Height(view), 30)
}

func TestPickerNoticeWithAPanel(t *testing.T) {
	panel := &PickPanel{Label: "Now playing", Title: "Momotarou", Lines: []string{"Episode 1"}}
	for _, size := range []tea.WindowSizeMsg{{Width: 70, Height: 30}, {Width: 160, Height: 30}} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice, Panel: panel})

			view := sized(model, size.Width, size.Height)
			text := ansi.Strip(view)

			assert.Contains(t, text, "no longer playable")
			assert.Contains(t, text, "Episode 1")
			assert.Contains(t, text, "Momotarou", "the panel is not lost to the notice")
			assert.LessOrEqual(t, lipgloss.Height(view), size.Height)
			assert.LessOrEqual(t, lipgloss.Width(view), size.Width)
		})
	}
}

func TestPickerNoticeInPlainMode(t *testing.T) {
	plainMode(t, true)
	model := newPickerModel(episodeItems(10), PickOptions{Notice: removedNotice})

	text := sized(model, 100, 20)

	assert.Contains(t, text, "[!]", "plain mode draws the marker as text")
	assert.Contains(t, text, "no longer playable")
}

// ---- the pending notice -------------------------------------------------------------

func TestNoticeOnNextPickerIsTakenOnce(t *testing.T) {
	NoticeOnNextPicker(removedNotice)

	assert.Equal(t, singleLine(removedNotice), takePickerNotice())
	assert.Empty(t, takePickerNotice(), "a notice is shown once")
}

func TestNoticeOnNextPickerKeepsOnlyTheLatest(t *testing.T) {
	NoticeOnNextPicker("first")
	NoticeOnNextPicker("second")

	assert.Equal(t, "second", takePickerNotice())
}

func TestNoticeOnNextPickerCollapsesNewlines(t *testing.T) {
	NoticeOnNextPicker("one\ntwo")

	assert.Equal(t, "one two", takePickerNotice())
}

func TestPickActionedTakesThePendingNoticeEvenWhenItCannotOpen(t *testing.T) {
	// No terminal in the test run, so the picker refuses to open. The notice
	// was meant for it, so it must not wait around for some later screen.
	NoticeOnNextPicker(removedNotice)

	_, _, err := PickActioned(episodeItems(3), PickOptions{})

	require.Error(t, err)
	assert.Empty(t, takePickerNotice(), "an unopened picker does not leave its notice for another one")
}

func TestPickActionedTakesThePendingNoticeEvenWithNoItems(t *testing.T) {
	NoticeOnNextPicker(removedNotice)

	_, _, err := PickActioned(nil, PickOptions{})

	require.ErrorIs(t, err, ErrNoPickItems)
	assert.Empty(t, takePickerNotice())
}

func TestWithPendingNotice(t *testing.T) {
	t.Run("a picker with none of its own takes the pending one", func(t *testing.T) {
		NoticeOnNextPicker("pending")
		assert.Equal(t, "pending", withPendingNotice(PickOptions{}).Notice)
		assert.Empty(t, takePickerNotice(), "and it is spent")
	})
	t.Run("an explicit notice wins, and the pending one is still spent", func(t *testing.T) {
		NoticeOnNextPicker("pending")
		assert.Equal(t, "explicit", withPendingNotice(PickOptions{Notice: "explicit"}).Notice)
		assert.Empty(t, takePickerNotice(), "it was meant for this picker, not a later one")
	})
	t.Run("nothing pending changes nothing", func(t *testing.T) {
		assert.Equal(t, PickOptions{Action: "play"}, withPendingNotice(PickOptions{Action: "play"}))
	})
	t.Run("the rest of the options pass through", func(t *testing.T) {
		NoticeOnNextPicker("pending")
		opts := PickOptions{Breadcrumb: "Search > Episodes", Action: "resume", InitialIndex: 3, HelpHint: true}
		got := withPendingNotice(opts)
		opts.Notice = "pending"
		assert.Equal(t, opts, got)
	})
}

// A panel is drawn whole or not at all. A notice takes rows, so on a short
// terminal the panel that fit without it may no longer fit: it must then give
// way, never be cut (its bottom edge is a row of ▀ that is the first thing to go).
func TestPickerNoticeNeverCutsThePanel(t *testing.T) {
	panel := &PickPanel{Label: "Now playing", Title: "Momotarou", Lines: []string{"Episode 1", "Resumes at 04:12", "of 23:40"}}
	for _, width := range []int{70, 110, 160} {
		for height := 8; height <= 40; height++ {
			t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
				model := newPickerModel(episodeItems(25), PickOptions{Notice: removedNotice, Panel: panel})

				view := sized(model, width, height)
				text := ansi.Strip(view)

				if strings.Contains(text, "Momotarou") {
					assert.Contains(t, text, "▀▀▀", "the panel is drawn without its bottom edge")
				}
				assert.LessOrEqual(t, lipgloss.Height(view), height)
			})
		}
	}
}
