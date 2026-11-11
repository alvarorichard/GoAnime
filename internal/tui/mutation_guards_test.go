package tui

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPickerStatusTruncatesTheTitleToTheRoomItHas(t *testing.T) {
	s := pickerStatus{prefix: "Added ", work: "Sousou no Frieren", suffix: " to Weekend."}

	assert.Equal(t, "Added Sousou no Frieren to Weekend.", s.text(100), "enough room: the whole title")
	assert.Equal(t, "Added Sous… to Weekend.", s.text(5), "short of room: cut with an ellipsis")
	assert.Equal(t, "Added Sousou no Frieren to Weekend.", s.text(0), "no room to measure: left whole")
	assert.Equal(t, "Added Sousou no Frieren to Weekend.", s.text(-3))
	assert.Equal(t, "Added  to Weekend.", pickerStatus{prefix: "Added ", suffix: " to Weekend."}.text(5), "no title")
}

func TestRowKeyOfARowPastTheEndIsEmptyNotAPanic(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	n := len(m.lists.all)
	is := assert.New(t)

	is.NotEmpty(m.rowKey(homeRow{kind: homeRowList, index: n - 1}), "the last list has a key")
	is.Equal("", m.rowKey(homeRow{kind: homeRowList, index: n}), "one past the last list")
	is.Equal(rowKeyList+m.lists.all[0].ID, m.rowKey(homeRow{kind: homeRowList, index: 0}))

	nf := len(m.favoriteAll)
	m.lists.level = homeLevelLists
	is.NotEmpty(m.rowKey(homeRow{kind: homeRowFavorite, index: nf - 1}), "the last favorite has a key")
	is.Equal("", m.rowKey(homeRow{kind: homeRowFavorite, index: nf}), "one past the last favorite")
}

func TestRestorePointOnAnEmptyFavoritesTabNamesNoRow(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	tabTo(t, m, "FAVORITES")
	assert.Empty(t, m.rows)

	for _, focus := range []bool{true, false} {
		m.focusInput = focus
		m.cursor = 0
		assert.NotPanics(t, func() {
			r := m.restorePoint(SearchHomeResult{})
			assert.Equal(t, TabFavorites, r.Tab)
			assert.Empty(t, r.RowKey)
		})
	}
}

func TestRestorePointOnTheFavoritesTabNamesTheRowAtTheEdges(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	last := len(m.rows) - 1
	m.focusInput = false

	m.cursor = 0
	assert.Equal(t, m.rowKey(m.rows[0]), m.restorePoint(SearchHomeResult{}).RowKey, "the first row")
	m.cursor = last
	assert.Equal(t, m.rowKey(m.rows[last]), m.restorePoint(SearchHomeResult{}).RowKey, "the last row")
	m.cursor = last + 1
	assert.Empty(t, m.restorePoint(SearchHomeResult{}).RowKey, "past the end")
	m.cursor = -1
	assert.Empty(t, m.restorePoint(SearchHomeResult{}).RowKey, "before the start")
}

func TestThePickerCountsFollowWhatIsToggled(t *testing.T) {
	theme := CurrentTheme()
	fake := pkNewLists()
	p := newPickerState(&theme, fake, HomeFavorite{Key: "frieren", Title: "Frieren"}, "Frieren", map[string]bool{})
	p.fit(80, 20)
	weekend := slices.IndexFunc(p.items, func(l HomeList) bool { return l.ID == "weekend" })
	require.GreaterOrEqual(t, weekend, 0)
	before := p.items[weekend].Count

	p.toggle(weekend)
	assert.Equal(t, before+1, p.items[weekend].Count, "added: one more title")
	assert.True(t, p.member["weekend"])

	p.toggle(weekend)
	assert.Equal(t, before, p.items[weekend].Count, "taken off: back to what it was")
	assert.False(t, p.member["weekend"])

	// A repeat of what the list already says changes no count.
	p.applyToggle(weekend, false)
	assert.Equal(t, before, p.items[weekend].Count, "not in it, taken off again: no change")
	p.applyToggle(weekend, true)
	p.applyToggle(weekend, true)
	assert.Equal(t, before+1, p.items[weekend].Count, "in it, added again: no change")
}

func TestThePickerCountsTheListsTheUserMade(t *testing.T) {
	theme := CurrentTheme()
	p := newPickerState(&theme, pkNewLists(), HomeFavorite{Key: "k"}, "T", map[string]bool{})
	assert.Equal(t, 2, p.userLists(), "Weekend and Kids; Favorites is not one of them")

	p.items = append(p.items, HomeList{ID: "odd", Name: "Odd", Auto: true})
	assert.Equal(t, 2, p.userLists(), "an automatic list is not the user's")
	p.items = append(p.items, HomeList{ID: FavoritesListID, Name: "Favorites again"})
	assert.Equal(t, 2, p.userLists(), "the Favorites id is never the user's")
}

func TestOpenPickerNeedsAHighlightedWorkAndLists(t *testing.T) {
	pkPlain(t)
	m, _ := pkResults(t, pkNewLists())
	assert.True(t, m.openPicker())

	m.lists = nil
	assert.False(t, m.openPicker(), "no lists to put it in")
}

func TestUndoNeverHidesAReadFailure(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	f.fail["Works"] = errors.New("flaky")
	hlSend(m, press('u', "u"))

	assert.Contains(t, screen(m), "Could not read your lists", "the failed read is what the screen says")
	assert.NotContains(t, screen(m), "is back in the list")
}

func TestEnterOnTheListsLevelWithNothingUnderTheCursorDoesNothing(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	m.focusInput = false
	for _, c := range []int{-1, len(m.rows)} {
		m.cursor = c
		assert.NotPanics(t, func() { m.openHighlighted() })
		assert.Equal(t, homeLevelLists, m.lists.level)
	}
	m.cursor = len(m.rows) - 1
	assert.NotPanics(t, func() { m.openHighlighted() }, "the last row exists")
}
