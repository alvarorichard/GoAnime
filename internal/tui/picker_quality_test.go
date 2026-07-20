package tui

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Quality / regression suite for the fancy picker: long catalogs (One Piece),
// dual nav (arrows + vim), type-to-filter (fzf), bounds, and parallel isolation.

func TestIsTypeToFilterKey_CodeOnlyPrintable(t *testing.T) {
	t.Parallel()
	// Some terminals deliver printable keys via Code without Text.
	assert.True(t, isTypeToFilterKey(tea.KeyPressMsg{Code: '5'}))
	assert.True(t, isTypeToFilterKey(tea.KeyPressMsg{Code: 'z'}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: 0}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: 127}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeyPgUp}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeyPgDown}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeyHome}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeyEnd}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeyTab}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}))
}

func TestIsTypeToFilterKey_ControlTextRejected(t *testing.T) {
	t.Parallel()
	// DEL / below-space text must never open the filter.
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Text: "\x1f"}))
	assert.False(t, isTypeToFilterKey(tea.KeyPressMsg{Text: string(rune(127))}))
}

func TestPickerModelUpdate_SlashStillOpensFilter(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "A"}, {Label: "B"}}, PickOptions{})
	_, _ = model.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assert.Equal(t, list.Filtering, model.entries.FilterState())
	assert.NoError(t, model.err)
}

func TestPickerModelUpdate_SpaceDoesNotOpenFilter(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "A"}, {Label: "B"}}, PickOptions{})
	_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	assert.Equal(t, list.Unfiltered, model.entries.FilterState())
}

func TestPickerModelUpdate_EnterWithNoSelectionStaysOpen(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "Only"}}, PickOptions{})
	model.entries.SetFilterText("zzz-no-match")
	require.Empty(t, model.entries.VisibleItems())
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Equal(t, -1, model.selectedIndex)
	assert.NoError(t, model.err)
}

func TestPickerModelUpdate_CtrlNPNavigate(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{
		{Label: "1"}, {Label: "2"}, {Label: "3"},
	}, PickOptions{})
	_, _ = model.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	assert.Equal(t, list.Unfiltered, model.entries.FilterState())
	assert.Equal(t, 1, model.entries.Index())
	_, _ = model.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	assert.Equal(t, 0, model.entries.Index())
}

func TestPickerModelUpdate_OnePieceStyleJump(t *testing.T) {
	t.Parallel()
	const total = 1169
	items := make([]PickItem, total)
	for i := range items {
		items[i] = PickItem{Label: fmt.Sprintf("One Piece - Episódio %d", i+1)}
	}
	model := newPickerModel(items, PickOptions{
		ItemSingular: "episode",
		ItemPlural:   "episodes",
	})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	assert.Equal(t, 100, model.entries.Width(), "full width — no blue panel")

	// Jump near the end without scrolling.
	for _, ch := range "1100" {
		_, _ = model.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}
	if model.filterPending {
		model.entries.SetFilterText(model.entries.FilterInput.Value())
		model.filterPending = false
	}
	require.NotEmpty(t, model.entries.VisibleItems())
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	assert.Equal(t, 1099, model.selectedIndex, "One Piece ep 1100")
}

func TestPickerModelUpdate_FilterThenClearThenBack(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{
		{Label: "Alpha"}, {Label: "Beta"}, {Label: "Gamma"},
	}, PickOptions{})
	model.entries.SetFilterText("Beta")
	require.Len(t, model.entries.VisibleItems(), 1)

	// First esc clears filter; second esc goes back.
	_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.NoError(t, model.err)
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, model.err, ErrPickBack)
}

func TestPickerModelUpdate_UnknownMsgIsNoop(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "A"}}, PickOptions{})
	before := model.entries.Index()
	_, cmd := model.Update(struct{ n int }{n: 1})
	assert.Nil(t, cmd)
	assert.Equal(t, before, model.entries.Index())
	assert.Equal(t, -1, model.selectedIndex)
}

func TestPickerModelView_ShowsCountAndNoPanelBorder(t *testing.T) {
	t.Parallel()
	items := make([]PickItem, 50)
	for i := range items {
		items[i] = PickItem{Label: fmt.Sprintf("Ep %02d", i+1)}
	}
	model := newPickerModel(items, PickOptions{
		WindowTitle:  "GoAnime - Episodes",
		ItemSingular: "episode",
		ItemPlural:   "episodes",
	})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	view := model.View()
	assert.Contains(t, view.Content, "episodes")
	assert.Contains(t, view.Content, "Ep 01")
	// Rounded panel borders from the old side preview must stay gone.
	assert.NotContains(t, view.Content, "╭")
	assert.NotContains(t, view.Content, "╰")
	assert.LessOrEqual(t, lipgloss.Width(view.Content), 120)
	assert.LessOrEqual(t, lipgloss.Height(view.Content), 30)
}

func TestPickWithRunner_OutOfRangeSelectionCancelled(t *testing.T) {
	t.Parallel()
	items := []PickItem{{Label: "A"}}
	idx, err := pickWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
		picker := model.(*pickerModel)
		picker.selectedIndex = 99 // corrupt / impossible index
		return picker, nil
	})
	assert.Equal(t, -1, idx)
	assert.ErrorIs(t, err, ErrPickCancelled)
}

func TestPickWithRunner_CancelErrorFromModel(t *testing.T) {
	t.Parallel()
	items := []PickItem{{Label: "A"}}
	idx, err := pickWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
		picker := model.(*pickerModel)
		_, _ = picker.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		return picker, nil
	})
	assert.Equal(t, -1, idx)
	assert.ErrorIs(t, err, ErrPickCancelled)
}

func TestPickWithRunner_ParallelIsolation(t *testing.T) {
	t.Parallel()
	const workers = 32
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(want int) {
			defer wg.Done()
			items := []PickItem{
				{Label: "zero"},
				{Label: "one"},
				{Label: "two"},
			}
			got, err := pickWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
				picker := model.(*pickerModel)
				picker.entries.Select(want)
				_, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				return picker, nil
			})
			if err != nil {
				errs <- err
				return
			}
			if got != want {
				errs <- fmt.Errorf("want %d got %d", want, got)
			}
		}(w % 3)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}

func TestStartFilterWithKey_ThenTypeDigitsViaUpdate(t *testing.T) {
	t.Parallel()
	items := make([]PickItem, 300)
	for i := range items {
		items[i] = PickItem{Label: fmt.Sprintf("%d — title", i+1)}
	}
	model := newPickerModel(items, PickOptions{})
	// First printable opens filter (fzf).
	_, _ = model.startFilterWithKey(tea.KeyPressMsg{Code: '2', Text: "2"})
	assert.Equal(t, list.Filtering, model.entries.FilterState())
	// Further digits go through normal Update while already filtering.
	_, _ = model.Update(tea.KeyPressMsg{Code: '5', Text: "5"})
	_, _ = model.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	model.entries.SetFilterText(model.entries.FilterInput.Value())
	model.filterPending = false
	require.NotEmpty(t, model.entries.VisibleItems())
	_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, 249, model.selectedIndex)
}

func TestNewPickerModel_EmptyItemNamesDefault(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "x"}}, PickOptions{
		ItemSingular: "",
		ItemPlural:   "",
	})
	// Status bar item names fall back to item/items — exercised via SetStatusBarItemName.
	assert.NotNil(t, model)
	assert.Equal(t, "Select", model.shell.Breadcrumb)
}

func TestFancyListKeyMap_PageBindings(t *testing.T) {
	t.Parallel()
	km := fancyListKeyMap()
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyPgUp}, km.PrevPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyPgDown}, km.NextPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyLeft}, km.PrevPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyRight}, km.NextPage))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyHome}, km.GoToStart))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyEnd}, km.GoToEnd))
}

func TestPickerFilterValue_NeverContainsControls(t *testing.T) {
	t.Parallel()
	entry := pickEntry{item: PickItem{
		Label:   "Ep\x1b[31m 12\x00",
		Details: "Filler\r\nRecap",
	}}
	got := entry.FilterValue()
	assert.NotContains(t, got, "\x1b")
	assert.NotContains(t, got, "\x00")
	assert.NotContains(t, got, "\r")
	assert.NotContains(t, got, "\n")
	for _, r := range got {
		assert.False(t, unicode.IsControl(r), "control rune %U in %q", r, got)
	}
}

func TestPick_EmptySliceAndEmptyLiteral(t *testing.T) {
	t.Parallel()
	for _, items := range [][]PickItem{nil, {}} {
		idx, err := Pick(items, PickOptions{})
		assert.Equal(t, -1, idx)
		assert.ErrorIs(t, err, ErrNoPickItems)
	}
}

func TestPickerModelUpdate_ResizePreservesSelection(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{
		{Label: "a"}, {Label: "b"}, {Label: "c"},
	}, PickOptions{})
	model.entries.Select(2)
	for _, w := range []int{40, 80, 120, 200} {
		_, _ = model.Update(tea.WindowSizeMsg{Width: w, Height: 24})
		assert.Equal(t, 2, model.entries.Index())
		assert.Equal(t, w, model.entries.Width())
	}
}

func TestPickerModelUpdate_RapidTypeDoesNotQuit(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "query match"}}, PickOptions{})
	start := time.Now()
	for _, ch := range "query" {
		_, _ = model.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		assert.NoError(t, model.err)
		assert.Equal(t, -1, model.selectedIndex)
	}
	assert.Less(t, time.Since(start), time.Second)
	assert.Equal(t, list.Filtering, model.entries.FilterState())
}
