package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// esc on the TRY screen edits first, then goes back: with text in the box it
// clears it, from the suggestions it returns to the box, and on an empty box
// it leaves for the SEARCH home (ErrSelectionBack). It never quits.

func escHint(m *tryModel) string {
	for _, h := range m.hints() {
		if h.Key == "esc" {
			return h.Desc
		}
	}
	return ""
}

func TestTryEscOnAnEmptyBoxGoesBack(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())

	assert.Equal(t, "clear", escHint(m), "with the query in the box, esc clears it")
	_, cmd := m.Update(tryPress(tea.KeyEscape, ""))
	assert.Nil(t, cmd, "the first esc only clears")
	require.NoError(t, m.err)
	assert.Empty(t, m.input.Value())

	assert.Equal(t, "back", escHint(m), "on an empty box the footer says where esc goes")
	_, cmd = m.Update(tryPress(tea.KeyEscape, ""))
	require.NotNil(t, cmd, "the screen closes")
	assert.ErrorIs(t, m.err, ErrSelectionBack, "back, not cancelled: the program goes on")
	assert.NotErrorIs(t, m.err, ErrSelectionCancelled)
	assert.False(t, m.decided, "nothing was chosen")
}

func TestTryEscFromTheSuggestionsReturnsToTheBoxEvenWhenItIsEmpty(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())
	m.input.SetValue("")
	_, _ = m.Update(tryPress(tea.KeyDown, ""))
	require.False(t, m.focusBox)
	assert.Equal(t, "edit", escHint(m))

	_, cmd := m.Update(tryPress(tea.KeyEscape, ""))
	assert.Nil(t, cmd, "it does not leave from the suggestions")
	assert.True(t, m.focusBox)
	assert.NoError(t, m.err)
}

func TestTryEscWithHelpOpenOnlyClosesTheHelp(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())
	m.input.SetValue("")
	_, _ = m.Update(tryPress(tea.KeyDown, ""))
	_, _ = m.Update(tryPress('?', "?"))
	require.True(t, m.showHelp)
	m.focusBox = true // as if the box held focus behind the overlay

	_, cmd := m.Update(tryPress(tea.KeyEscape, ""))
	assert.Nil(t, cmd)
	assert.False(t, m.showHelp)
	assert.NoError(t, m.err, "esc closes the overlay; it does not also go back")
}

func TestTryCtrlCStillQuits(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSelectionCancelled, "ctrl+c is the one key that quits")
}

func TestTrySearchReturnsBackToItsCaller(t *testing.T) {
	t.Parallel()
	_, err := trySearchWithRunner(tryTestOptions(), func(model tea.Model) (tea.Model, error) {
		m := model.(*tryModel)
		m.input.SetValue("")
		_, _ = m.Update(tryPress(tea.KeyEscape, ""))
		return m, nil
	})
	assert.ErrorIs(t, err, ErrSelectionBack)
}

func TestTryHelpSaysWhereEscGoes(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())
	_, _ = m.Update(tryPress(tea.KeyDown, ""))
	_, _ = m.Update(tryPress('?', "?"))
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "back to search")
	assert.Contains(t, view, "esc never quits", "the promise still holds")
}

func TestTryFooterListsEscBetweenTabAndHelp(t *testing.T) {
	t.Parallel()
	m := newTryModelWithOptions(tryTestOptions())
	var keys []string
	for _, h := range m.hints() {
		keys = append(keys, h.Key)
	}
	require.Len(t, keys, 6)
	assert.Contains(t, []string{"↓", "dn"}, keys[1], "the move key, drawn plain on a plain terminal")
	keys[1] = "move"
	assert.Equal(t, []string{"enter", "move", "tab", "esc", "?", "ctrl+c"}, keys)
}

func TestTryHelpOverlayKeys(t *testing.T) {
	t.Parallel()
	open := func() *tryModel {
		m := newTryModelWithOptions(tryTestOptions())
		_, _ = m.Update(tryPress(tea.KeyDown, ""))
		_, _ = m.Update(tryPress('?', "?"))
		require.True(t, m.showHelp)
		return m
	}

	m := open()
	_, cmd := m.Update(tryPress('x', "x"))
	assert.Nil(t, cmd, "other keys do nothing behind the overlay")
	assert.True(t, m.showHelp)
	assert.NoError(t, m.err)

	m = open()
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd, "ctrl+c quits from the overlay too")
	assert.ErrorIs(t, m.err, ErrSelectionCancelled)
}
