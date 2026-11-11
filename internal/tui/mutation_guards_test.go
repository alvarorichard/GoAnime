package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	require := assert.New(t)

	require.NotEmpty(m.rowKey(homeRow{kind: homeRowList, index: n - 1}), "the last list has a key")
	require.Equal("", m.rowKey(homeRow{kind: homeRowList, index: n}), "one past the last list")
	require.Equal(rowKeyList+m.lists.all[0].ID, m.rowKey(homeRow{kind: homeRowList, index: 0}))

	nf := len(m.favoriteAll)
	m.lists.level = homeLevelLists
	require.NotEmpty(m.rowKey(homeRow{kind: homeRowFavorite, index: nf - 1}), "the last favorite has a key")
	require.Equal("", m.rowKey(homeRow{kind: homeRowFavorite, index: nf}), "one past the last favorite")
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
