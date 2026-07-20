package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPickEntryFilterValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		item PickItem
		want string
	}{
		{"label and details", PickItem{Label: "Season 2", Details: "12 episodes"}, "Season 2 12 episodes"},
		{"label only", PickItem{Label: "Season 2"}, "Season 2"},
		{"empty", PickItem{}, ""},
		{"strips ansi and controls", PickItem{Label: "\x1b[31mSeason\x1b[0m\t2", Details: "a\nb"}, "Season 2 a b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, pickEntry{item: tt.item}.FilterValue())
		})
	}
}

func TestPickEntryTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		item PickItem
		want string
	}{
		{"plain label", PickItem{Label: "Episode 12 — Pilot"}, "Episode 12 — Pilot"},
		{"empty label falls back", PickItem{}, "Untitled"},
		{"control-only label falls back", PickItem{Label: "\x1b[2J\x07"}, "Untitled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, pickEntry{item: tt.item}.Title())
		})
	}
}

func TestPickEntryDescription(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		item PickItem
		want string
	}{
		{"plain details", PickItem{Details: "12 episodes  •  2019"}, "12 episodes • 2019"},
		{"empty details", PickItem{}, ""},
		{"multiline collapses", PickItem{Details: "a\r\nb"}, "a b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, pickEntry{item: tt.item}.Description())
		})
	}
}

func TestNewPickerModel(t *testing.T) {
	t.Parallel()

	t.Run("defaults applied", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "A"}}, PickOptions{})
		assert.Equal(t, "Select", model.shell.Breadcrumb)
		assert.Equal(t, -1, model.selectedIndex)
		assert.NoError(t, model.err)
		require.Len(t, model.entries.Items(), 1)
		assert.Equal(t, 0, model.entries.Index())
	})

	t.Run("options and initial index", func(t *testing.T) {
		t.Parallel()
		items := []PickItem{{Label: "S1"}, {Label: "S2"}, {Label: "S3"}}
		model := newPickerModel(items, PickOptions{
			Breadcrumb:   "Search > Show > Seasons",
			ItemSingular: "season",
			ItemPlural:   "seasons",
			InitialIndex: 2,
		})
		assert.Equal(t, "Search > Show > Seasons", model.shell.Breadcrumb)
		assert.Equal(t, 2, model.entries.Index())
	})

	t.Run("sanitizes breadcrumb and window title", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "A"}}, PickOptions{
			Breadcrumb:  "Search > \x1b[31mEvil\x1b[0m\nShow",
			WindowTitle: "Go\x07Anime",
		})
		assert.Equal(t, "Search > Evil Show", model.shell.Breadcrumb)
		assert.Equal(t, "Go Anime", model.options.WindowTitle)
		assert.NotContains(t, model.shell.Breadcrumb, "\x1b")
		assert.NotContains(t, model.options.WindowTitle, "\x07")
	})

	t.Run("out of range initial index falls back to first", func(t *testing.T) {
		t.Parallel()
		for _, initial := range []int{-3, 5} {
			model := newPickerModel([]PickItem{{Label: "A"}, {Label: "B"}}, PickOptions{InitialIndex: initial})
			assert.Equal(t, 0, model.entries.Index(), "initial index %d", initial)
		}
	})
}

func TestPickerModelInit(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "A"}}, PickOptions{})
	assert.Nil(t, model.Init())
}

func TestPickerModelUpdate(t *testing.T) {
	t.Parallel()

	newModel := func(t *testing.T) *pickerModel {
		t.Helper()
		return newPickerModel([]PickItem{
			{Label: "Season 1", Details: "8 episodes"},
			{Label: "Season 2", Details: "12 episodes"},
			{Label: "Specials", Details: "2 episodes"},
		}, PickOptions{})
	}

	t.Run("resize keeps full width list", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		for _, width := range []int{80, 99, 100, 120} {
			_, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: 24})
			assert.Equal(t, width, model.entries.Width(), "no side panel — list uses full width")
		}
	})

	t.Run("escape goes back", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrPickBack)
	})

	t.Run("ctrl c cancels even while filtering", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		require.Equal(t, list.Filtering, model.entries.FilterState())
		_, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrPickCancelled)
	})

	t.Run("typing digit starts filter like fzf", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		require.Equal(t, list.Unfiltered, model.entries.FilterState())

		_, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})

		assert.Equal(t, list.Filtering, model.entries.FilterState())
		assert.Equal(t, "2", model.entries.FilterInput.Value())
		assert.NoError(t, model.err)
		assert.Equal(t, -1, model.selectedIndex)
	})

	t.Run("type episode number then select without arrow keys", func(t *testing.T) {
		t.Parallel()
		items := make([]PickItem, 200)
		for i := range items {
			items[i] = PickItem{Label: fmt.Sprintf("%d — Episode title", i+1)}
		}
		model := newPickerModel(items, PickOptions{})

		// fzf-style: type "150" directly, no "/" and no arrow spam.
		for _, ch := range []struct {
			code rune
			text string
		}{{'1', "1"}, {'5', "5"}, {'0', "0"}} {
			_, _ = model.Update(tea.KeyPressMsg{Code: ch.code, Text: ch.text})
		}
		// Apply pending filter matches the same way the program loop would.
		if model.filterPending {
			model.entries.SetFilterText(model.entries.FilterInput.Value())
			model.filterPending = false
		}
		require.NotEmpty(t, model.entries.VisibleItems())

		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Equal(t, 149, model.selectedIndex,
			"typing 150 + Enter must land on episode 150 without scrolling")
	})

	t.Run("j still navigates without opening filter", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		assert.Equal(t, list.Unfiltered, model.entries.FilterState())
		assert.Equal(t, 1, model.entries.Index())
	})

	t.Run("arrow down navigates like vim j", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		assert.Equal(t, list.Unfiltered, model.entries.FilterState())
		assert.Equal(t, 1, model.entries.Index())
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		assert.Equal(t, 0, model.entries.Index())
	})

	t.Run("g and G jump first last", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
		assert.Equal(t, 2, model.entries.Index())
		_, _ = model.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
		assert.Equal(t, 0, model.entries.Index())
		assert.Equal(t, list.Unfiltered, model.entries.FilterState())
	})

	t.Run("enter selects current row", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Equal(t, 1, model.selectedIndex)
		assert.NoError(t, model.err)
	})

	t.Run("q types into filter instead of quitting", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		// Unfiltered: q is type-to-filter (fzf), not quit.
		_, _ = model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		assert.NoError(t, model.err)
		assert.Equal(t, list.Filtering, model.entries.FilterState())
		assert.Equal(t, "q", model.entries.FilterInput.Value())
	})

	t.Run("enter with pending filter does not select", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		_, _ = model.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		_, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		require.True(t, model.filterPending)
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.Equal(t, -1, model.selectedIndex)
		assert.Equal(t, list.Filtering, model.entries.FilterState())
	})

	t.Run("filter matches message clears pending flag", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		model.filterPending = true
		_, _ = model.Update(list.FilterMatchesMsg{})
		assert.False(t, model.filterPending)
	})

	t.Run("filtered selection maps to original index", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		model.entries.SetFilterText("Specials")
		require.Len(t, model.entries.VisibleItems(), 1)
		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Equal(t, 2, model.selectedIndex,
			"selection must return the index in the original items slice, not the filtered position")
	})

	t.Run("empty filtered result cannot select", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		model.entries.SetFilterText("does-not-exist")
		require.Empty(t, model.entries.VisibleItems())
		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.Nil(t, cmd)
		assert.Equal(t, -1, model.selectedIndex)
	})

	t.Run("escape with applied filter clears filter before backing out", func(t *testing.T) {
		t.Parallel()
		model := newModel(t)
		model.entries.SetFilterText("Season")
		require.NotEqual(t, list.Unfiltered, model.entries.FilterState())
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.NoError(t, model.err)

		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrPickBack)
	})

	t.Run("large list filters and selects correctly", func(t *testing.T) {
		t.Parallel()
		items := make([]PickItem, 5000)
		for i := range items {
			items[i] = PickItem{Label: fmt.Sprintf("Episode %04d", i+1)}
		}
		model := newPickerModel(items, PickOptions{})
		model.entries.SetFilterText("Episode 4321")
		require.NotEmpty(t, model.entries.VisibleItems())
		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Equal(t, 4320, model.selectedIndex)
	})
}

func TestPickerModelView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "minimum compact", width: 20, height: 5},
		{name: "compact", width: 80, height: 8},
		{name: "wide", width: 120, height: 18},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := newPickerModel([]PickItem{{
				Label:   "Season 2 with a very long label that must never overflow the terminal frame",
				Details: "12 episodes  •  2019",
			}}, PickOptions{WindowTitle: "GoAnime - Seasons"})
			_, _ = model.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})

			view := model.View()

			assert.True(t, view.AltScreen)
			assert.Equal(t, "GoAnime - Seasons", view.WindowTitle)
			if tt.width >= 9 {
				assert.Contains(t, view.Content, "GOANIME")
			}
			// Full-width list only — no side preview panel.
			assert.NotContains(t, view.Content, "╭")
			assert.LessOrEqual(t, lipgloss.Width(view.Content), tt.width)
			assert.LessOrEqual(t, lipgloss.Height(view.Content), tt.height)
		})
	}

	t.Run("default window title", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "A"}}, PickOptions{})
		assert.Equal(t, "GoAnime", model.View().WindowTitle)
	})

	t.Run("wide layout still shows list rows only", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "Season 9", Details: "3 episodes"}}, PickOptions{})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
		content := model.View().Content
		assert.Contains(t, content, "Season 9")
		assert.Contains(t, content, "3 episodes")
		assert.Equal(t, 120, model.entries.Width())
	})
}

func TestPick(t *testing.T) {
	t.Parallel()
	// The TTY path cannot run headlessly; the guard clauses shared with
	// pickWithRunner are pinned here.
	idx, err := Pick(nil, PickOptions{})
	assert.Equal(t, -1, idx)
	assert.ErrorIs(t, err, ErrNoPickItems)
}

func TestIsTypeToFilterKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want bool
	}{
		{"digit", tea.KeyPressMsg{Code: '9', Text: "9"}, true},
		{"letter", tea.KeyPressMsg{Code: 'a', Text: "a"}, true},
		{"punctuation", tea.KeyPressMsg{Code: '-', Text: "-"}, true},
		{"down arrow", tea.KeyPressMsg{Code: tea.KeyDown}, false},
		{"up arrow", tea.KeyPressMsg{Code: tea.KeyUp}, false},
		{"j vim", tea.KeyPressMsg{Code: 'j', Text: "j"}, false},
		{"k vim", tea.KeyPressMsg{Code: 'k', Text: "k"}, false},
		{"h vim", tea.KeyPressMsg{Code: 'h', Text: "h"}, false},
		{"l vim", tea.KeyPressMsg{Code: 'l', Text: "l"}, false},
		{"ctrl n emacs", tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl}, false},
		{"ctrl p emacs", tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, false},
		{"ctrl d page", tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl}, false},
		{"ctrl u page", tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}, false},
		{"g first", tea.KeyPressMsg{Code: 'g', Text: "g"}, false},
		{"G last", tea.KeyPressMsg{Code: 'G', Text: "G"}, false},
		{"slash filter chord", tea.KeyPressMsg{Code: '/', Text: "/"}, false},
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}, false},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isTypeToFilterKey(tt.msg))
		})
	}
}

func TestFancyListKeyMap(t *testing.T) {
	t.Parallel()
	km := fancyListKeyMap()
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyDown}, km.CursorDown))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'j', Text: "j"}, km.CursorDown))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyUp}, km.CursorUp))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'k', Text: "k"}, km.CursorUp))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'h', Text: "h"}, km.PrevPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'l', Text: "l"}, km.NextPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'g', Text: "g"}, km.GoToStart))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'G', Text: "G"}, km.GoToEnd))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl}, km.NextPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}, km.PrevPage))
	assert.False(t, km.Quit.Enabled(), "quit owned by picker, not list")
	assert.False(t, km.ForceQuit.Enabled())
}

func TestFancyListHelpKeys(t *testing.T) {
	t.Parallel()
	keys := fancyListHelpKeys()
	require.NotEmpty(t, keys)
	helps := make([]string, 0, len(keys))
	for _, b := range keys {
		helps = append(helps, b.Help().Key+" "+b.Help().Desc)
	}
	joined := strings.Join(helps, " | ")
	assert.Contains(t, joined, "↑↓/jk")
	assert.Contains(t, joined, "type")
	assert.Contains(t, joined, "g/G")
	assert.Contains(t, joined, "enter")
	assert.Contains(t, joined, "esc")
}

func TestStartFilterWithKey(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{
		{Label: "1 — First"},
		{Label: "42 — Answer"},
		{Label: "100 — Century"},
	}, PickOptions{})
	_, _ = model.startFilterWithKey(tea.KeyPressMsg{Code: '4', Text: "4"})
	assert.Equal(t, list.Filtering, model.entries.FilterState())
	assert.Equal(t, "4", model.entries.FilterInput.Value())
	assert.True(t, model.filterPending)
}

func TestPickWithRunner(t *testing.T) {
	t.Parallel()
	items := []PickItem{{Label: "Season 1"}, {Label: "Season 2"}, {Label: "Specials"}}

	t.Run("empty items", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(nil, PickOptions{}, func(model tea.Model) (tea.Model, error) {
			t.Fatal("runner must not be called for empty items")
			return model, nil
		})
		assert.Equal(t, -1, idx)
		assert.ErrorIs(t, err, ErrNoPickItems)
	})

	t.Run("nil runner", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{}, nil)
		assert.Equal(t, -1, idx)
		assert.ErrorContains(t, err, "picker runner not configured")
	})

	t.Run("runner error is wrapped", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("terminal exploded")
		idx, err := pickWithRunner(items, PickOptions{}, func(tea.Model) (tea.Model, error) {
			return nil, boom
		})
		assert.Equal(t, -1, idx)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("unexpected final model type", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{}, func(tea.Model) (tea.Model, error) {
			return nil, nil
		})
		assert.Equal(t, -1, idx)
		assert.ErrorContains(t, err, "unexpected picker model")
	})

	t.Run("typed nil final model", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{}, func(tea.Model) (tea.Model, error) {
			return (*pickerModel)(nil), nil
		})
		assert.Equal(t, -1, idx)
		assert.ErrorContains(t, err, "unexpected picker model")
	})

	t.Run("back propagates", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
			picker, ok := model.(*pickerModel)
			require.True(t, ok)
			_, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			return picker, nil
		})
		assert.Equal(t, -1, idx)
		assert.ErrorIs(t, err, ErrPickBack)
	})

	t.Run("quit without selection is cancelled", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
			return model, nil
		})
		assert.Equal(t, -1, idx)
		assert.ErrorIs(t, err, ErrPickCancelled)
	})

	t.Run("cascade filter then selection returns original index", func(t *testing.T) {
		t.Parallel()
		idx, err := pickWithRunner(items, PickOptions{InitialIndex: 1}, func(model tea.Model) (tea.Model, error) {
			picker, ok := model.(*pickerModel)
			require.True(t, ok)
			_, _ = picker.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
			picker.entries.SetFilterText("Specials")
			_, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			return picker, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, idx)
	})
}
