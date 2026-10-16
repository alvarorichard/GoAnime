package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func actionTestOptions() PickOptions {
	return PickOptions{
		Breadcrumb:        "Search > Show > Episodes",
		Action:            "resume",
		BackDesc:          "results",
		HelpHint:          true,
		FilterPlaceholder: "type a number to jump, / to search titles",
		Actions: []PickAction{
			{Key: "r", Label: "from the start", Hint: "from start"},
			{Key: "d", Label: "this episode", Hint: "download", Section: true},
			{Key: "D", Label: "a range, from here to ...", Section: true},
			{Key: "A", Label: "all 12 episodes", Section: true},
			{Key: "m", Label: "every option, upscale included", Section: true},
		},
		ActionsLabel: "Download",
		ActionsGloss: "ダウンロード",
	}
}

func TestPickerModelActions(t *testing.T) {
	t.Parallel()

	t.Run("action keys registered", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "Ep 1"}, {Label: "Ep 2"}}, actionTestOptions())
		assert.Len(t, model.actionKeys, 5)
		assert.Empty(t, model.actionKey)
	})

	t.Run("action quits with highlighted row", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "Ep 1"}, {Label: "Ep 2"}}, actionTestOptions())
		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		updated, cmd := model.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
		require.NotNil(t, cmd)
		m := updated.(*pickerModel)
		assert.Equal(t, "d", m.actionKey)
		assert.Equal(t, 1, m.selectedIndex, "download acts on the highlighted row")
	})

	t.Run("actions win over filtering", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "Drama"}}, actionTestOptions())
		updated, cmd := model.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
		require.NotNil(t, cmd, "m opens the menu instead of filtering for it")
		assert.Equal(t, "m", updated.(*pickerModel).actionKey)
	})

	t.Run("unknown keys still filter", func(t *testing.T) {
		t.Parallel()
		model := newPickerModel([]PickItem{{Label: "Ep 1"}}, actionTestOptions())
		_, _ = model.Update(tea.KeyPressMsg{Code: 'z', Text: "z"})
		assert.Equal(t, "z", model.entries.FilterValue())
	})
}

func TestPickerHintsWithActions(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "Ep 1"}}, actionTestOptions())

	var got []string
	for _, h := range model.hints() {
		got = append(got, h.Key+" "+h.Desc)
	}
	assert.Equal(t, []string{
		"enter resume", "r from start", "d download", "↑↓ move", "? help", "esc results",
	}, got, "tela 5 footer order")
}

func TestPickerActionsSection(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "Ep 1"}}, actionTestOptions())

	section := ansi.Strip(model.actionsSection())
	assert.Contains(t, section, "DOWNLOAD")
	assert.Contains(t, section, "d")
	assert.Contains(t, section, "this episode")
	assert.Contains(t, section, "every option, upscale included")
	assert.NotContains(t, section, "from the start", "r resumes; it is no download")

	plain := newPickerModel([]PickItem{{Label: "Ep 1"}}, PickOptions{})
	assert.Equal(t, "", plain.actionsSection(), "no label, no section")
}

func TestPickerHelpWithActions(t *testing.T) {
	t.Parallel()
	model := newPickerModel([]PickItem{{Label: "Ep 1"}}, actionTestOptions())
	extra := model.helpExtra()
	require.Contains(t, extra, "DOWNLOAD")
	var keys []string
	for _, row := range extra["DOWNLOAD"] {
		keys = append(keys, row.Key)
	}
	assert.Equal(t, []string{"r", "d", "D", "A", "m"}, keys)
}

func TestPickActioned(t *testing.T) {
	t.Parallel()
	items := []PickItem{{Label: "Ep 1"}, {Label: "Ep 2"}}

	t.Run("selection returns empty action", func(t *testing.T) {
		t.Parallel()
		index, action, err := pickActionedWithRunner(items, PickOptions{}, func(model tea.Model) (tea.Model, error) {
			m := model.(*pickerModel)
			m.selectedIndex = 1
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, index)
		assert.Equal(t, "", action)
	})

	t.Run("action returns key with index", func(t *testing.T) {
		t.Parallel()
		index, action, err := pickActionedWithRunner(items, actionTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*pickerModel)
			m.selectedIndex = 0
			m.actionKey = "A"
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 0, index)
		assert.Equal(t, "A", action)
	})

	t.Run("plain Pick drops the action", func(t *testing.T) {
		t.Parallel()
		index, err := pickWithRunner(items, actionTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*pickerModel)
			m.actionKey = "d"
			m.selectedIndex = 1
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("nil runner", func(t *testing.T) {
		t.Parallel()
		_, _, err := pickActionedWithRunner(items, PickOptions{}, nil)
		assert.ErrorContains(t, err, "runner not configured")
	})

	t.Run("runner error", func(t *testing.T) {
		t.Parallel()
		want := errors.New("no tty")
		_, _, err := pickActionedWithRunner(items, PickOptions{}, func(tea.Model) (tea.Model, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
	})
}

func TestPickerFilterPlaceholder(t *testing.T) {
	t.Parallel()
	custom := newPickerModel([]PickItem{{Label: "Ep 1"}}, actionTestOptions())
	assert.Equal(t, "type a number to jump, / to search titles", custom.entries.FilterInput.Placeholder)

	fallback := newPickerModel([]PickItem{{Label: "Ep 1"}}, PickOptions{})
	assert.Equal(t, "type number or title…", fallback.entries.FilterInput.Placeholder)
}
