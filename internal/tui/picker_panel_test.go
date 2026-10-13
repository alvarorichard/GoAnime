package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func menuItems() []PickItem {
	return []PickItem{
		{Label: "Next episode", Key: "n"},
		{Label: "Previous episode", Key: "p"},
		{Label: "← Back", Key: "esc"},
		{Label: "Exit", Key: "q"},
	}
}

func TestPickerShortcuts(t *testing.T) {
	t.Parallel()

	t.Run("a key picks its row at once", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(menuItems(), PickOptions{})
		_, cmd := model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		require.NotNil(t, cmd)
		assert.Equal(t, 1, model.selectedIndex)
	})

	t.Run("other characters still filter", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(menuItems(), PickOptions{})
		_, _ = model.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
		assert.Equal(t, -1, model.selectedIndex)
		assert.Equal(t, list.Filtering, model.entries.FilterState())
	})

	t.Run("while filtering a key is typed, not picked", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(menuItems(), PickOptions{})
		_, _ = model.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
		_, _ = model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		assert.Equal(t, -1, model.selectedIndex)
		assert.Equal(t, "en", model.entries.FilterInput.Value())
	})

	t.Run("esc is shown, not bound: it already goes back", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(menuItems(), PickOptions{})
		assert.Equal(t, []string{"n", "p", "q"}, model.keys)
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.ErrorIs(t, model.err, ErrPickBack)
	})

	t.Run("the footer lists the keys instead of type-to-filter", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(menuItems(), PickOptions{Action: "choose"})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
		footer := ansi.Strip(renderHints(&model.theme, model.hints(), 100))
		assert.Contains(t, footer, "n p q  shortcuts")
		assert.Contains(t, footer, "enter  choose")
		assert.NotContains(t, footer, "filter")
	})

	t.Run("a repeated key belongs to the first row", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "A", Key: "a"}, {Label: "B", Key: "a"}}, PickOptions{})
		_, _ = model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
		assert.Equal(t, 0, model.selectedIndex)
	})
}

func panelItems() []PickItem {
	items := make([]PickItem, 12)
	for i := range items {
		items[i] = PickItem{Label: "Option"}
	}
	return items
}

func TestPickerPanelLayout(t *testing.T) {
	t.Parallel()
	panel := &PickPanel{Label: "Now playing", Title: "Momotarou", Lines: []string{"Episode 1"}}

	t.Run("beside the list on a wide terminal", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(panelItems(), PickOptions{Panel: panel})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
		assert.Equal(t, resultsListWidth(120)-resultsGutter, model.entries.Width())
		view := model.View().Content
		assert.Contains(t, ansi.Strip(view), "NOW PLAYING")
		assert.LessOrEqual(t, lipgloss.Width(view), 120)
		assert.LessOrEqual(t, lipgloss.Height(view), 30)
	})

	t.Run("above the list on a narrow one", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(panelItems(), PickOptions{Panel: panel})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 70, Height: 30})
		assert.Equal(t, 70, model.entries.Width())
		view := ansi.Strip(model.View().Content)
		require.Contains(t, view, "NOW PLAYING")
		assert.Less(t, strings.Index(view, "NOW PLAYING"), strings.Index(view, "Option"))
		assert.LessOrEqual(t, lipgloss.Height(model.View().Content), 30)
	})

	t.Run("left out when the list would have no room", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(panelItems(), PickOptions{Panel: panel})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 70, Height: 12})
		assert.NotContains(t, ansi.Strip(model.View().Content), "NOW PLAYING")
		assert.Equal(t, 8, model.entries.Height(), "the list keeps the whole body")
	})

	t.Run("the list keeps enough rows: the panel goes compact first", func(t *testing.T) {
		t.Parallel()
		live := &PickPanel{Label: "Now playing", Title: "Momotarou", Lines: []string{"Episode 1"},
			Progress: func() (float64, float64, bool) { return 492, 1420, true }}
		model := newPickerModel(panelItems()[:8], PickOptions{Panel: live})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 90, Height: 25})
		view := ansi.Strip(model.View().Content)
		require.Contains(t, view, "NOW PLAYING")
		assert.Equal(t, 8, strings.Count(view, "Option"), "every option is on screen")
		assert.LessOrEqual(t, lipgloss.Height(model.View().Content), 25)
	})

	t.Run("never out of bounds", func(t *testing.T) {
		t.Parallel()
		for _, size := range []tea.WindowSizeMsg{{Width: 20, Height: 6}, {Width: 50, Height: 14}, {Width: 99, Height: 24}, {Width: 100, Height: 10}, {Width: 160, Height: 50}} {
			model := newPickerModel(panelItems(), PickOptions{Panel: panel})
			_, _ = model.Update(size)
			view := model.View().Content
			assert.LessOrEqual(t, lipgloss.Width(view), size.Width, "%dx%d", size.Width, size.Height)
			assert.LessOrEqual(t, lipgloss.Height(view), size.Height, "%dx%d", size.Width, size.Height)
		}
	})
}

func TestPickerPanelProgress(t *testing.T) {
	t.Parallel()

	t.Run("a still panel is read once, up front", func(t *testing.T) {
		t.Parallel()
		calls := 0
		model := newPickerModel(panelItems(), PickOptions{Panel: &PickPanel{Label: "Continue", Progress: func() (float64, float64, bool) {
			calls++
			return 872, 1440, true
		}}})
		assert.Nil(t, model.Init())
		assert.Equal(t, 1, calls)
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
		assert.Contains(t, ansi.Strip(model.View().Content), "14:32 / 24:00")
	})

	t.Run("a live panel polls through commands", func(t *testing.T) {
		t.Parallel()
		position := 10.0
		model := newPickerModel(panelItems(), PickOptions{Panel: &PickPanel{Label: "Now playing", Live: true, Progress: func() (float64, float64, bool) {
			return position, 1440, true
		}}})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
		assert.Contains(t, ansi.Strip(model.View().Content), "--:-- / --:--", "nothing is asked before the program starts")

		poll := model.Init()
		require.NotNil(t, poll)
		_, tick := model.Update(poll())
		assert.Contains(t, ansi.Strip(model.View().Content), "00:10 / 24:00")
		require.NotNil(t, tick, "the next poll is scheduled")

		position = 70
		_, next := model.Update(progressTickMsg{})
		require.NotNil(t, next)
		_, _ = model.Update(next())
		assert.Contains(t, ansi.Strip(model.View().Content), "01:10 / 24:00")
	})

	t.Run("a failed poll keeps the last answer", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel(panelItems(), PickOptions{Panel: &PickPanel{Live: true, Progress: func() (float64, float64, bool) { return 0, 0, false }}})
		model.progress = panelProgress{position: 30, duration: 60}
		_, _ = model.Update(progressMsg{ok: false})
		assert.Equal(t, panelProgress{position: 30, duration: 60}, model.progress)
	})
}

func TestRenderNotice(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	notice := "Could not load the episodes of Kuro Nyago on AnimeFire. Try another result."

	assert.Empty(t, renderNotice(&theme, "", 80))

	one := strings.Split(strings.TrimSuffix(ansi.Strip(renderNotice(&theme, notice, 120)), "\n"), "\n")
	assert.Equal(t, []string{" !  " + notice}, one)

	two := strings.Split(strings.TrimSuffix(ansi.Strip(renderNotice(&theme, notice, 60)), "\n"), "\n")
	require.Len(t, two, 2)
	assert.True(t, strings.HasPrefix(two[1], "    "), "the second line hangs under the text")
	for _, line := range two {
		assert.LessOrEqual(t, lipgloss.Width(line), 60)
	}

	cut := strings.Split(strings.TrimSuffix(ansi.Strip(renderNotice(&theme, notice+" "+notice, 40)), "\n"), "\n")
	require.Len(t, cut, maxNoticeLines)
	assert.True(t, strings.HasSuffix(cut[1], "…"))
}
