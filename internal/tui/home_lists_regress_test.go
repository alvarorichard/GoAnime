package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With Options.Lists nil the FAVORITES tab is what it was before lists
// existed: these tests pin its exact words and keys.

func hlPlainHome(t *testing.T, w, h int) (*searchHomeModel, *favoritesFake) {
	t.Helper()
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(hlCloneFav(sampleFavorites), nil, nil))
	_, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	tabTo(t, m, "FAVORITES")
	return m, f
}

func hlCloneFav(in []HomeFavorite) []HomeFavorite { return append([]HomeFavorite(nil), in...) }

func TestWithoutListsTheModelHasNoListState(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	assert.Nil(t, m.lists)
	assert.False(t, m.listsActive())
	assert.False(t, m.listsLevelOne())
	assert.False(t, m.listsLevelTwo())
	assert.Equal(t, sampleFavorites, m.tabFavorites())
}

func TestWithoutListsTheTabIsTheOldFlatList(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	view := screen(m)
	assert.Equal(t, "type to filter your favorites, enter opens the first one", m.input.Placeholder)
	assert.Contains(t, view, "FAVORITES  2")
	assert.Contains(t, view, "Frieren")
	assert.Contains(t, view, "A Origem")
	assert.NotContains(t, view, "+ Create")
	assert.NotContains(t, view, "back to your lists")
	for _, row := range m.rows {
		assert.Equal(t, homeRowFavorite, row.kind)
	}
}

func TestWithoutListsTheHintsAreTheOldOnes(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	assert.Equal(t, []Hint{
		{Key: "enter", Desc: "open", Essential: true},
		{Key: "↑↓", Desc: "move"},
		{Key: "tab", Desc: "scope"},
		{Key: "?", Desc: "help"},
		{Key: "ctrl+c", Desc: "quit", Essential: true},
	}, m.hints(), "box focused")

	downTo(t, m, 0)
	assert.Equal(t, []Hint{
		{Key: "enter", Desc: "open", Essential: true},
		{Key: "↑↓", Desc: "move"},
		{Key: "tab", Desc: "scope"},
		{Key: "x", Desc: "remove", Essential: true},
		{Key: "X", Desc: "clear list"},
		{Key: "?", Desc: "help"},
		{Key: "ctrl+c", Desc: "quit", Essential: true},
	}, m.hints(), "row highlighted")

	hlSend(m, press('X', "X"))
	assert.Equal(t, []Hint{
		{Key: "y", Desc: "clear all", Essential: true},
		{Key: "n", Desc: "keep", Essential: true},
		{Key: "ctrl+c", Desc: "quit"},
	}, m.hints(), "clear question")
}

func TestWithoutListsTheHelpIsTheOldOne(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	downTo(t, m, 0)
	hlSend(m, press('?', "?"))
	view := screen(m)
	for _, want := range []string{"search, resume", "change scope; FAVORITES lists your favorites", "choose sources", "from the start", "episodes", "remove", "clear the list"} {
		assert.Contains(t, view, want)
	}
	for _, not := range []string{"back to your lists", "rename a list", "FAVORITES: "} {
		assert.NotContains(t, view, not)
	}
}

func TestWithoutListsTheEmptyAndNoMatchWordsAreTheOldOnes(t *testing.T) {
	empty := newSearchHomeModel(SearchHomeOptions{ScopeIndex: 4})
	_, _ = empty.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	assert.Contains(t, screen(empty), "You have no favorites yet. Search for a title, then press ctrl+f on it to add it here.")
	assert.Equal(t, "You have no favorites yet. Search for a title, then press ctrl+f on it to add it here.", empty.tabEmptyText())

	m, _ := hlPlainHome(t, 100, 40)
	typeText(m, "zzz")
	assert.Contains(t, screen(m), `No favorite matches "zzz". Delete a letter to see them all.`)
	assert.Equal(t, `No favorite matches "zzz". Delete a letter to see them all.`, m.tabNoMatchText())
}

func TestWithoutListsTheKeysBehaveAsBefore(t *testing.T) {
	m, f := hlPlainHome(t, 100, 40)

	downTo(t, m, 0)
	hlSend(m, hlEsc)
	assert.Equal(t, 0, m.cursor, "esc on a highlighted favorite keeps the highlight")
	assert.False(t, m.focusInput)

	hlSend(m, press('r', "r"), press('e', "e"))
	assert.Empty(t, f.removed)
	assert.Len(t, m.favoriteAll, 2)

	hlSend(m, press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 2 favorites from FAVORITES? y to clear, n to keep.")
	hlSend(m, press('n', "n"))

	hlSend(m, press('x', "x"))
	assert.Equal(t, []string{"frieren"}, f.removed)
	assert.Len(t, m.favoriteAll, 1)

	_, cmd := m.Update(hlEnter)
	require.NotNil(t, cmd)
	assert.Equal(t, "A Origem", m.result.Query)
	assert.Equal(t, "", m.result.Scope)
}

func TestWithoutListsEnterInTheBoxOpensTheFirstMatch(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	typeText(m, "orig")
	_, cmd := m.Update(hlEnter)
	require.NotNil(t, cmd)
	assert.Equal(t, "A Origem", m.result.Query)
}

func TestWithoutListsTheHelpKeyAndSourcesStillWork(t *testing.T) {
	m, _ := hlPlainHome(t, 100, 40)
	downTo(t, m, 0)
	hlSend(m, press('?', "?"))
	require.True(t, m.showHelp)
	hlSend(m, press('z', "z"))
	assert.True(t, m.showHelp, "other keys do nothing while the help shows")
	hlSend(m, hlEsc)
	assert.False(t, m.showHelp)
	hlSend(m, press('?', "?"))
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}

// ---- Lists set: the other tabs are untouched --------------------------------------------

func TestTheOtherTabsLookTheSameWithOrWithoutLists(t *testing.T) {
	continueAll, recent := homeTestEntries()
	f := &favoritesFake{}
	plain := newSearchHomeModel(f.options(hlCloneFav(sampleFavorites), continueAll, recent))
	opts := f.options(hlCloneFav(sampleFavorites), continueAll, recent)
	opts.Lists = newHomeListsFake().withFavorites(2)
	with := newSearchHomeModel(opts)
	for _, m := range []*searchHomeModel{plain, with} {
		_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	}
	for range 4 {
		assert.Equal(t, screen(plain), screen(with), plain.scopes[plain.scopeIdx].Name)
		assert.Equal(t, plain.hints(), with.hints())
		hlSend(plain, hlTab)
		hlSend(with, hlTab)
	}
}

func TestWithListsTheAllSectionStillUsesTheOptionsFavorites(t *testing.T) {
	continueAll, recent := homeTestEntries()
	f := &favoritesFake{}
	opts := f.options(hlCloneFav(sampleFavorites), continueAll, recent)
	opts.Lists = newHomeListsFake()
	m := newSearchHomeModel(opts)
	assert.Contains(t, screen(m), "Frieren")
	hlRowIs(t, m, "Frieren")
	hlSend(m, press('x', "x"))
	assert.Equal(t, []string{"frieren"}, f.removed, "x on the ALL section is the old callback")
}

func TestWithListsTheSearchTabsStillSuggestAndSearch(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Lists: newHomeListsFake()})
	typeText(m, "naruto")
	_, cmd := m.Update(hlEnter)
	require.NotNil(t, cmd)
	assert.Equal(t, "naruto", m.result.Query)
	assert.Equal(t, HomeSearch, m.result.Action)
}

func TestWithListsEscOnTheSearchTabsStillClearsTheBox(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Lists: newHomeListsFake()})
	typeText(m, "abc")
	hlSend(m, hlEsc)
	assert.Equal(t, "", m.input.Value())
	hlSend(m, press('r', "r"))
	assert.Equal(t, "r", m.input.Value(), "r types into the box on a search tab")
}
