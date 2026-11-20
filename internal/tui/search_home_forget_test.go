package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Forgetting history from the SEARCH home screen: x removes the highlighted
// row of either list, X asks (y/n) and clears the whole list the highlighted
// row is in. The history itself is a fake here: the tests say what it answers
// and read back what the screen asked of it.

// historyFake stands in for the stored history.
type historyFake struct {
	deletedEntries  []string
	deletedQueries  []string
	clearedContinue int
	clearedRecent   int
	err             error
}

func (h *historyFake) options(continueAll []HomeContinueEntry, recent []HomeRecentSearch) SearchHomeOptions {
	return SearchHomeOptions{
		Continue: continueAll,
		Recent:   recent,
		DeleteEntry: func(key string) error {
			h.deletedEntries = append(h.deletedEntries, key)
			return h.err
		},
		DeleteRecent: func(query string) error {
			h.deletedQueries = append(h.deletedQueries, query)
			return h.err
		},
		ClearContinue: func() error { h.clearedContinue++; return h.err },
		ClearRecent:   func() error { h.clearedRecent++; return h.err },
	}
}

// homeWithHistory is a screen with 3 CONTINUE rows and 2 RECENT rows, the box
// focused, and the fake history behind it.
func homeWithHistory(t *testing.T) (*searchHomeModel, *historyFake) {
	t.Helper()
	continueAll, recent := homeTestEntries()
	h := &historyFake{}
	return newSearchHomeModel(h.options(continueAll, recent)), h
}

// downTo presses down until row is highlighted.
func downTo(t *testing.T, m *searchHomeModel, row int) {
	t.Helper()
	for i := 0; i < 2*len(m.rows)+2 && (m.focusInput || m.cursor != row); i++ {
		_, _ = m.Update(press(tea.KeyDown, ""))
	}
	require.False(t, m.focusInput)
	require.Equal(t, row, m.cursor)
}

const (
	firstContinue = 0
	firstRecent   = 3
)

func screen(m *searchHomeModel) string { return ansi.Strip(m.View().Content) }

// ---- x on a RECENT row --------------------------------------------------------

func TestXRemovesTheHighlightedRecentSearch(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstRecent)

	_, cmd := m.Update(press('x', "x"))

	assert.Nil(t, cmd, "removing stays on the screen")
	assert.Equal(t, []string{"namakura gatana"}, h.deletedQueries)
	assert.Empty(t, h.deletedEntries, "a search is not an episode")
	require.Len(t, m.recentAll, 1)
	assert.Equal(t, "momotarou", m.recentAll[0].Query)
	assert.Len(t, m.continueAll, 3, "CONTINUE is untouched")
	assert.Len(t, m.rows, 4)
}

func TestXRemovesTheSecondRecentSearchWithoutTouchingTheFirst(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstRecent+1)

	_, _ = m.Update(press('x', "x"))

	assert.Equal(t, []string{"momotarou"}, h.deletedQueries)
	require.Len(t, m.recentAll, 1)
	assert.Equal(t, "namakura gatana", m.recentAll[0].Query)
	assert.Equal(t, firstRecent, m.cursor, "the highlight lands on the row that is left")
}

func TestXOnTheLastRecentSearchHandsTheBoxItsFocusBack(t *testing.T) {
	h := &historyFake{}
	_, recent := homeTestEntries()
	m := newSearchHomeModel(h.options(nil, recent[:1]))
	downTo(t, m, 0)

	_, _ = m.Update(press('x', "x"))

	assert.Empty(t, m.recentAll)
	assert.Empty(t, m.rows)
	assert.True(t, m.focusInput, "nothing left to pick")
	assert.True(t, m.input.Focused())
}

func TestXOnARecentSearchThatCannotBeRemovedDoesNothing(t *testing.T) {
	_, recent := homeTestEntries()
	m := newSearchHomeModel(SearchHomeOptions{Recent: recent})
	downTo(t, m, 0)

	_, cmd := m.Update(press('x', "x"))

	assert.Nil(t, cmd)
	assert.Len(t, m.recentAll, 2, "no remover, no removal")
	assert.Empty(t, m.notice, "and nothing to apologise for")
}

func TestARecentSearchThatFailsToGoStaysAndTheScreenSaysWhy(t *testing.T) {
	m, h := homeWithHistory(t)
	h.err = errors.New("database is locked")
	downTo(t, m, firstRecent)

	_, _ = m.Update(press('x', "x"))

	assert.Len(t, m.recentAll, 2)
	assert.Contains(t, screen(m), "Could not remove the search: database is locked")
}

func TestRAndEDoNothingOnARecentRow(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstRecent)

	for _, key := range []rune{'r', 'e'} {
		_, cmd := m.Update(press(key, string(key)))
		assert.Nil(t, cmd, string(key))
	}
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Empty(t, h.deletedQueries)
}

// ---- x on a CONTINUE row: a failure now says so -------------------------------

func TestAnEpisodeThatFailsToGoStaysAndTheScreenSaysWhy(t *testing.T) {
	m, h := homeWithHistory(t)
	h.err = errors.New("db locked")
	downTo(t, m, firstContinue)

	_, _ = m.Update(press('x', "x"))

	assert.Len(t, m.continueAll, 3)
	assert.Equal(t, []string{"k1"}, h.deletedEntries)
	assert.Contains(t, screen(m), "Could not remove the episode: db locked")
}

func TestTheFailureNoticeGoesAwayWithTheNextKey(t *testing.T) {
	m, h := homeWithHistory(t)
	h.err = errors.New("db locked")
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('x', "x"))
	require.NotEmpty(t, m.notice)

	_, _ = m.Update(press(tea.KeyDown, ""))

	assert.Empty(t, m.notice)
	assert.NotContains(t, screen(m), "Could not")
}

func TestOnlyTheFailingRemovalIsNoticed(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstContinue)

	_, _ = m.Update(press('x', "x"))

	assert.Empty(t, m.notice, "a removal that worked says nothing")
	assert.NotContains(t, screen(m), "Could not")
	assert.Equal(t, []string{"k1"}, h.deletedEntries)
}

// ---- X: ask, then clear the whole list ----------------------------------------

func TestXOnAContinueRowAsksBeforeClearingAnything(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstContinue)

	_, cmd := m.Update(press('X', "X"))

	assert.Nil(t, cmd)
	assert.Equal(t, clearContinueRows, m.clearing)
	assert.Zero(t, h.clearedContinue, "asking clears nothing")
	assert.Contains(t, screen(m), "Clear all 3 episodes from CONTINUE? y to clear, n to keep.")
}

func TestXOnARecentRowAsksAboutRecent(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstRecent)

	_, _ = m.Update(press('X', "X"))

	assert.Equal(t, clearRecentRows, m.clearing)
	assert.Zero(t, h.clearedRecent)
	assert.Contains(t, screen(m), "Clear all 2 searches from RECENT? y to clear, n to keep.")
}

func TestTheQuestionSaysOneInTheSingular(t *testing.T) {
	continueAll, recent := homeTestEntries()
	h := &historyFake{}
	m := newSearchHomeModel(h.options(continueAll[:1], recent[:1]))

	downTo(t, m, 0)
	_, _ = m.Update(press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 1 episode from CONTINUE?")

	_, _ = m.Update(press('n', "n"))
	_, _ = m.Update(press(tea.KeyDown, ""))
	_, _ = m.Update(press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 1 search from RECENT?")
}

func TestYClearsTheContinueListAndOnlyThat(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('X', "X"))

	_, cmd := m.Update(press('y', "y"))

	assert.Nil(t, cmd)
	assert.Equal(t, 1, h.clearedContinue)
	assert.Zero(t, h.clearedRecent, "RECENT is another list")
	assert.Empty(t, m.continueAll)
	assert.Len(t, m.recentAll, 2)
	assert.Len(t, m.rows, 2)
	assert.Equal(t, clearNothing, m.clearing)
	assert.Empty(t, m.notice)
	assert.NotContains(t, screen(m), "Clear all")
}

func TestYClearsTheRecentListAndOnlyThat(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstRecent)
	_, _ = m.Update(press('X', "X"))

	_, _ = m.Update(press('y', "y"))

	assert.Equal(t, 1, h.clearedRecent)
	assert.Zero(t, h.clearedContinue, "CONTINUE is another list")
	assert.Empty(t, m.recentAll)
	assert.Len(t, m.continueAll, 3)
	assert.Len(t, m.rows, 3)
}

func TestClearingTheOnlyListHandsTheBoxItsFocusBack(t *testing.T) {
	continueAll, _ := homeTestEntries()
	h := &historyFake{}
	m := newSearchHomeModel(h.options(continueAll, nil))
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('X', "X"))

	_, _ = m.Update(press('y', "y"))

	assert.True(t, m.focusInput)
	assert.True(t, m.input.Focused())
	assert.Empty(t, m.rows)
}

func TestNAndEscKeepTheList(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{press('n', "n"), press(tea.KeyEscape, "")} {
		m, h := homeWithHistory(t)
		downTo(t, m, firstContinue)
		_, _ = m.Update(press('X', "X"))

		_, cmd := m.Update(key)

		assert.Nil(t, cmd, key.String())
		assert.Equal(t, clearNothing, m.clearing, key.String())
		assert.Zero(t, h.clearedContinue, key.String())
		assert.Len(t, m.continueAll, 3, key.String())
	}
}

func TestOtherKeysDoNotAnswerTheQuestionNorLeakThrough(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('X', "X"))

	for _, key := range []tea.KeyPressMsg{
		press('q', "q"), press('Z', "Z"), press(tea.KeyEnter, ""), press(tea.KeyDown, ""),
		press('x', "x"), press('r', "r"), press(tea.KeyTab, ""), press('?', "?"),
	} {
		_, cmd := m.Update(key)
		assert.Nil(t, cmd, key.String())
		assert.Equal(t, clearContinueRows, m.clearing, "%s leaves the question open", key.String())
	}
	assert.Zero(t, h.clearedContinue)
	assert.Empty(t, h.deletedEntries, "x does not act on the row behind the question")
	assert.Zero(t, m.scopeIdx, "tab does not change scope")
	assert.Equal(t, firstContinue, m.cursor)
	assert.False(t, m.showHelp)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Empty(t, m.input.Value(), "nothing typed into the box")
}

func TestCtrlCQuitsFromTheQuestion(t *testing.T) {
	m, h := homeWithHistory(t)
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('X', "X"))

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
	assert.Zero(t, h.clearedContinue)
}

func TestAListThatFailsToClearStaysAndTheScreenSaysWhy(t *testing.T) {
	for _, tc := range []struct {
		row  int
		what string
	}{
		{firstContinue, "Could not clear CONTINUE: disk full"},
		{firstRecent, "Could not clear RECENT: disk full"},
	} {
		m, h := homeWithHistory(t)
		h.err = errors.New("disk full")
		downTo(t, m, tc.row)
		_, _ = m.Update(press('X', "X"))

		_, _ = m.Update(press('y', "y"))

		assert.Len(t, m.continueAll, 3, tc.what)
		assert.Len(t, m.recentAll, 2, tc.what)
		assert.Equal(t, clearNothing, m.clearing, "the question is closed either way")
		assert.Contains(t, screen(m), tc.what)
	}
}

func TestXAsksNothingWhenTheListCannotBeClearedOrIsEmpty(t *testing.T) {
	continueAll, recent := homeTestEntries()

	t.Run("no clearer on the host", func(t *testing.T) {
		m := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})
		downTo(t, m, firstContinue)
		_, _ = m.Update(press('X', "X"))
		assert.Equal(t, clearNothing, m.clearing)
		downTo(t, m, firstRecent)
		_, _ = m.Update(press('X', "X"))
		assert.Equal(t, clearNothing, m.clearing)
	})

	t.Run("a clearer for the other list only", func(t *testing.T) {
		m := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent, ClearContinue: func() error { return nil }})
		downTo(t, m, firstRecent)
		_, _ = m.Update(press('X', "X"))
		assert.Equal(t, clearNothing, m.clearing, "ClearContinue is not RECENT's clearer")

		m = newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent, ClearRecent: func() error { return nil }})
		downTo(t, m, firstContinue)
		_, _ = m.Update(press('X', "X"))
		assert.Equal(t, clearNothing, m.clearing, "ClearRecent is not CONTINUE's clearer")
	})

	t.Run("nothing to clear", func(t *testing.T) {
		m := newSearchHomeModel(SearchHomeOptions{ClearContinue: func() error { return nil }, ClearRecent: func() error { return nil }})
		m.askToClear(homeRowContinue)
		assert.Equal(t, clearNothing, m.clearing)
		m.askToClear(homeRowRecent)
		assert.Equal(t, clearNothing, m.clearing)
	})
}

func TestXInTheBoxTypesAnX(t *testing.T) {
	m, h := homeWithHistory(t)
	require.True(t, m.focusInput)

	_, _ = m.Update(press('X', "X"))

	assert.Equal(t, "X", m.input.Value())
	assert.Equal(t, clearNothing, m.clearing, "no question while typing")
	assert.Zero(t, h.clearedContinue)
}

func TestClearConfirmedWithNoQuestionOpenDoesNothing(t *testing.T) {
	m, h := homeWithHistory(t)

	m.clearConfirmed()

	assert.Zero(t, h.clearedContinue+h.clearedRecent)
	assert.Len(t, m.continueAll, 3)
	assert.Len(t, m.recentAll, 2)
}

// ---- what the user is shown ---------------------------------------------------

func TestCountOf(t *testing.T) {
	assert.Equal(t, "0 episodes", countOf(0, "episode", "episodes"))
	assert.Equal(t, "1 episode", countOf(1, "episode", "episodes"))
	assert.Equal(t, "2 episodes", countOf(2, "episode", "episodes"))
	assert.Equal(t, "1 search", countOf(1, "search", "searches"))
	assert.Equal(t, "12 searches", countOf(12, "search", "searches"))
}

func TestQuestionViewIsEmptyWithNothingToSay(t *testing.T) {
	m, _ := homeWithHistory(t)

	assert.Empty(t, m.questionView(80))
}

func TestFailedKeepsTheCauseAndTheScreenDrawsItOnOneLine(t *testing.T) {
	m, _ := homeWithHistory(t)

	m.failed("remove the search", errors.New("line one\nline two"))

	assert.Equal(t, "Could not remove the search: line one\nline two", m.notice)
	view := ansi.Strip(m.questionView(80))
	assert.Contains(t, view, "line one line two")
	assert.NotContains(t, view, "line one\n")
}

func TestTheQuestionHasItsOwnFooter(t *testing.T) {
	m, _ := homeWithHistory(t)
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('X', "X"))

	var keys []string
	for _, h := range m.hints() {
		keys = append(keys, h.Key)
	}

	assert.Equal(t, []string{"y", "n", "ctrl+c"}, keys)
}

func TestTheFooterOffersRemoveAndClearOnlyOverARow(t *testing.T) {
	m, _ := homeWithHistory(t)

	assert.False(t, hintKeys(m)["x"], "in the box x is a letter")
	assert.False(t, hintKeys(m)["X"])

	downTo(t, m, firstContinue)
	assert.True(t, hintKeys(m)["x"])
	assert.True(t, hintKeys(m)["X"])
	for _, h := range m.hints() {
		switch h.Key {
		case "x":
			assert.Equal(t, "remove", h.Desc)
		case "X":
			assert.Equal(t, "clear list", h.Desc)
		}
	}
}

func TestTheHelpScreenListsClearTheList(t *testing.T) {
	m, _ := homeWithHistory(t)
	downTo(t, m, firstContinue)
	_, _ = m.Update(press('?', "?"))
	require.True(t, m.showHelp)

	view := screen(m)

	assert.Contains(t, view, "clear the list")
	assert.Contains(t, view, "remove")
}

func TestSearchHomeWiresHistoryOptionsIntoTheModel(t *testing.T) {
	h := &historyFake{}
	continueAll, recent := homeTestEntries()

	m := newSearchHomeModel(h.options(continueAll, recent))

	require.NotNil(t, m.deleteRecent)
	require.NotNil(t, m.clearContinue)
	require.NotNil(t, m.clearRecent)
	require.NoError(t, m.deleteRecent("q"))
	require.NoError(t, m.clearContinue())
	require.NoError(t, m.clearRecent())
	assert.Equal(t, []string{"q"}, h.deletedQueries)
	assert.Equal(t, 1, h.clearedContinue)
	assert.Equal(t, 1, h.clearedRecent)
}

func TestRowActionWithNoRowHighlightedDoesNothing(t *testing.T) {
	m, h := homeWithHistory(t)
	require.True(t, m.focusInput, "the box has the focus: no row is highlighted")

	for _, key := range []string{"r", "e", "x", "X"} {
		assert.Nil(t, m.rowAction(key), key)
	}

	assert.Equal(t, clearNothing, m.clearing)
	assert.Empty(t, h.deletedEntries)
	assert.Empty(t, h.deletedQueries)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Len(t, m.continueAll, 3)
	assert.Len(t, m.recentAll, 2)
}
