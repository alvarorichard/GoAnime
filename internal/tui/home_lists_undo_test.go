package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemovingATitleSaysHowToPutItBack(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))

	assert.Contains(t, screen(m), "Removed Weekend marathon 0. Press u to put it back.")
	assert.Len(t, m.lists.works, 11)
}

func TestUPutsTheRemovedTitleBack(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	hlSend(m, press('u', "u"))

	assert.Len(t, m.lists.works, 12)
	assert.Contains(t, screen(m), "Weekend marathon 0 is back in the list.")
	assert.Nil(t, m.lists.removed, "one undo, not two")
}

func TestUDoesNothingWhenNothingWasRemoved(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('u', "u"))

	assert.Len(t, m.lists.works, 12)
	for _, c := range f.calls() {
		assert.NotContains(t, c, "ToggleWork")
	}
}

func TestUInTheBoxIsJustAFilterLetter(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	hlType(m, "u")

	for _, c := range f.calls() {
		assert.NotContains(t, c, "ToggleWork")
	}
	assert.Len(t, m.lists.works, 11, "typing u in the box does not undo")
}

func TestLeavingTheListForgetsTheUndo(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyEscape})

	require.Equal(t, homeLevelLists, m.lists.level)
	assert.Nil(t, m.lists.removed)
}

func TestAFailedUndoIsSaid(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	f.fail["ToggleWork"] = errors.New("disk full")
	hlSend(m, press('u', "u"))

	assert.Contains(t, screen(m), "Could not put the title back.")
}

func TestADamagedFileNoticeStaysOnTheFavoritesTab(t *testing.T) {
	f := hlThreeLists()
	m := newSearchHomeModel(SearchHomeOptions{Lists: f, Notice: "Your saved lists were damaged and set aside."})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	tabTo(t, m, "FAVORITES")
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyDown})
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyDown})

	assert.Contains(t, screen(m), "Your saved lists were damaged")
}

func TestADamagedFileNoticeDoesNotFollowToOtherTabs(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Lists: hlThreeLists(), Notice: "Your saved lists were damaged and set aside."})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyTab})

	assert.NotContains(t, screen(m), "damaged")
}
