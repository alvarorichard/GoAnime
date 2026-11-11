package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hlOpen opens the list called name from the first level.
func hlOpen(t *testing.T, m *searchHomeModel, name string) {
	t.Helper()
	hlRowIs(t, m, name)
	hlSend(m, hlEnter)
	require.Equal(t, homeLevelWorks, m.lists.level)
}

// ---- level 2: the works of one list ---------------------------------------------------

func TestAnOpenListShowsItsNameCountAndWorks(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	view := screen(m)

	assert.Contains(t, view, "Weekend marathon  12")
	assert.NotContains(t, view, "FAVORITES  12", "the heading is the list name now")
	assert.NotContains(t, view, "+ Create")
	assert.Contains(t, view, "Weekend marathon 0")
	assert.Contains(t, view, "Weekend marathon 11")
	assert.Len(t, m.rows, 12)
	assert.Equal(t, homeWorksPlaceholder, m.input.Placeholder)
	assert.True(t, m.focusInput)
	assert.Equal(t, "", m.input.Value())
}

func TestTheWorksRowsShowYearTypeAndSource(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 0)
	f.works["a"] = []HomeFavorite{{Key: "k", Title: "Frieren", Year: "2023", Media: "Anime", Source: "Animefire.io"}}
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Mine")
	line := hlLineWith(screen(m), "Frieren")
	assert.Contains(t, line, "2023 · Anime")
	assert.Contains(t, line, "AnimeFire")
}

func TestEnterOnAWorkSearchesItsTitleEverywhere(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 0)
	f.works["a"] = []HomeFavorite{{Key: "p", Title: "Pokémon: Aventuras"}, {Key: "q", Title: "Naruto"}}
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Mine")
	hlGoto(t, m, 0)
	_, cmd := m.Update(hlEnter)
	require.NotNil(t, cmd)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Equal(t, "Pokémon Aventuras", m.result.Query)
	assert.Equal(t, "", m.result.Scope)
}

func TestEnterOnTheSecondWorkSearchesTheSecond(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 3)
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Mine")
	hlGoto(t, m, 1)
	_, _ = m.Update(hlEnter)
	assert.Equal(t, "Mine 1", m.result.Query)
}

func TestEnterInTheBoxOpensTheFirstMatchingWork(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "marathon 7")
	_, cmd := m.Update(hlEnter)
	require.NotNil(t, cmd)
	assert.Equal(t, "Weekend marathon 7", m.result.Query)
}

func TestEnterOnAnEmptyListDoesNothing(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Empty")
	_, cmd := m.Update(hlEnter)
	assert.Nil(t, cmd)
	assert.Equal(t, homeLevelWorks, m.lists.level)
}

func TestTypingFiltersTheWorksOfTheList(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "marathon 1")
	assert.Len(t, m.rows, 3, "1, 10 and 11")
	assert.Contains(t, screen(m), "Weekend marathon  3 of 12")

	hlType(m, "zzz")
	assert.Contains(t, screen(m), `No title matches "marathon 1zzz". Delete a letter to see them all.`)
}

func TestAnEmptyListSaysHowToFillIt(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Empty")
	assert.Contains(t, screen(m), "This list is empty. Search for a title, then press ctrl+l on it to add it here.")
	assert.Contains(t, screen(m), "Empty one")
	assert.Empty(t, m.rows)
}

func TestAnEmptyFavoritesListSaysCtrlF(t *testing.T) {
	m := homeListsScreen(t, newHomeListsFake(), 100, 40)
	hlOpen(t, m, "Favorites")
	view := screen(m)
	assert.Contains(t, view, "This list is empty. Search for a title, then press ctrl+f on it to add it here.")
	assert.NotContains(t, view, "ctrl+l")
}

func TestTheFavoritesListShowsItsWorksNewestFirstAsGiven(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Favorites")
	assert.Equal(t, []string{"Fav 0", "Fav 1", "Fav 2"}, hlTitles(m.lists.works))
	assert.Contains(t, screen(m), "Favorites  3")
}

// ---- footer on the works ------------------------------------------------------------------

func TestTheFooterOnAListNamesTheWayBack(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	d := hlHintDescs(m)
	assert.Equal(t, "back to your lists", d["esc"])
	assert.Equal(t, "open", d["enter"])
	assert.NotContains(t, d, "x", "nothing to remove while the box has the focus")
	h := hintsByKey(m)["esc"]
	assert.True(t, h.Essential)
	assert.Equal(t, []string{"enter", "esc", "↑↓", "tab", "?", "ctrl+c"}, hlHintOrder(m))
}

func TestTheFooterOnAWorkOffersRemoveAndClear(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	d := hlHintDescs(m)
	assert.Equal(t, "remove", d["x"])
	assert.Equal(t, "clear list", d["X"])
	assert.True(t, hintsByKey(m)["x"].Essential)
	assert.Equal(t, "back to your lists", d["esc"])
	assert.NotContains(t, d, "r")
}

func TestTheFooterSaysClearFilterWhileThereIsText(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "mar")
	assert.Equal(t, "clear filter", hlHintDescs(m)["esc"])
	hlSend(m, hlEsc)
	assert.Equal(t, "back to your lists", hlHintDescs(m)["esc"])
}

func TestTheFooterKeepsWayBackOnANarrowTerminal(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 40, 20)
	hlOpen(t, m, "Weekend")
	foot := m.View().Content
	assert.Contains(t, foot, "back to your lists")
}

func TestTheClearQuestionKeepsTheStandardFooter(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"))
	assert.Equal(t, []string{"y", "n", "ctrl+c"}, hlHintOrder(m))
	assert.Equal(t, "clear all", hlHintDescs(m)["y"])
}

// ---- esc paths ----------------------------------------------------------------------------

func TestEscWithAnEmptyBoxGoesBackToYourLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Short")
	hlSend(m, hlEsc)

	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Equal(t, homeListsPlaceholder, m.input.Placeholder)
	assert.Equal(t, "Short", hlHighlighted(m), "the highlight is on the list just left")
	assert.False(t, m.focusInput)
	assert.Contains(t, screen(m), "YOUR LISTS  4")
	assert.Contains(t, screen(m), "+ Create a new list")
}

func TestEscKeepsTheHighlightOnEachListLeft(t *testing.T) {
	for _, name := range []string{"Favorites", "Weekend marathon", "Short", "Empty one"} {
		m := homeListsScreen(t, hlThreeLists(), 80, 24)
		hlOpen(t, m, name)
		hlSend(m, hlEsc)
		assert.Equal(t, name, hlHighlighted(m))
		assert.Contains(t, hlBody(m), name)
	}
}

func TestEscWithTextClearsTheTextFirst(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "mar")
	hlSend(m, hlEsc)
	assert.Equal(t, homeLevelWorks, m.lists.level, "the first esc only clears the text")
	assert.Equal(t, "", m.input.Value())
	assert.Len(t, m.rows, 12)
	hlSend(m, hlEsc)
	assert.Equal(t, homeLevelLists, m.lists.level)
}

func TestEscOnAHighlightedWorkWithAnEmptyBoxGoesBack(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 3)
	hlSend(m, hlEsc)
	assert.Equal(t, homeLevelLists, m.lists.level)
}

func TestGoingBackReadsTheCountsAgain(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"), hlEsc)
	assert.Contains(t, hlLineWith(screen(m), "Weekend marathon"), "11 titles")
}

func TestOpeningAListAgainStartsAtItsTopWithAnEmptyBox(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "mar")
	hlGoto(t, m, 2)
	hlSend(m, hlEsc, hlEsc)
	require.Equal(t, homeLevelLists, m.lists.level)
	hlSend(m, hlEnter)
	assert.Equal(t, homeLevelWorks, m.lists.level)
	assert.Equal(t, "", m.input.Value())
	assert.Equal(t, 0, m.cursor)
}

func TestTabInsideAListLeavesAndComesBackToTheLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlSend(m, hlTab)
	assert.Equal(t, "ALL", m.scopes[m.scopeIdx].Name)
	assert.Contains(t, m.input.Placeholder, "title, or")
	tabTo(t, m, "FAVORITES")
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Equal(t, homeListsPlaceholder, m.input.Placeholder)
}

func TestRAndEDoNothingOnAWork(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('r', "r"), press('e', "e"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Empty(t, f.calls())
	assert.Equal(t, "", m.result.Query)
}

// ---- removing and clearing -----------------------------------------------------------------

func TestXRemovesTheHighlightedWorkFromTheOpenList(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 2)
	hlSend(m, press('x', "x"))

	assert.Equal(t, []string{"RemoveWork(a,weekend marathon-2)"}, f.calls())
	assert.Len(t, m.rows, 11)
	assert.Len(t, m.lists.works, 11)
	assert.NotContains(t, screen(m), "Weekend marathon 2\n")
	assert.Contains(t, screen(m), "Weekend marathon  11")
	assert.Equal(t, 2, m.cursor, "the next work slides under the highlight")
	assert.Equal(t, "Weekend marathon 3", m.lists.works[m.rows[m.cursor].index].Title)
}

func TestXWhileFilteringRemovesTheFilteredWork(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlType(m, "marathon 1")
	hlGoto(t, m, 1)
	hlSend(m, press('x', "x"))
	assert.Equal(t, []string{"RemoveWork(a,weekend marathon-10)"}, f.calls())
	assert.Len(t, m.rows, 2)
}

func TestRemovingTheLastWorkGivesTheBoxTheFocusAndSaysTheListIsEmpty(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 1)
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Mine")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	assert.True(t, m.focusInput)
	assert.Contains(t, screen(m), "This list is empty.")
}

func TestRemovingTheLastRowMovesTheHighlightUp(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 11)
	hlSend(m, press('x', "x"))
	assert.Equal(t, 10, m.cursor)
	assert.False(t, m.focusInput)
}

func TestARemovalThatFailsKeepsTheWorkAndSaysWhy(t *testing.T) {
	f := hlThreeLists()
	f.fail["RemoveWork"] = errors.New("the file is read-only")
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	assert.Contains(t, screen(m), "Could not remove the title. The file is read-only.")
	assert.Len(t, m.lists.works, 12)
}

func TestRemovingFromTheFavoritesListRefreshesTheAllTabSection(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	require.Len(t, m.favoriteAll, 3)
	hlOpen(t, m, "Favorites")
	hlGoto(t, m, 1)
	hlSend(m, press('x', "x"))

	assert.Contains(t, f.log, "Works(favorites)")
	assert.Equal(t, []string{"Fav 0", "Fav 2"}, hlTitles(m.favoriteAll))
	hlSend(m, hlTab)
	tabTo(t, m, "ALL")
	view := screen(m)
	assert.Contains(t, view, "Fav 0")
	assert.NotContains(t, view, "Fav 1")
	assert.Contains(t, view, "Fav 2")
}

func TestRemovingFromAUserListLeavesTheAllSectionAlone(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	before := slices.Clone(m.favoriteAll)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"))
	assert.Equal(t, before, m.favoriteAll)
	assert.NotContains(t, f.log, "Works(favorites)")
}

func TestTheAllSectionAndTheTabNeverDisagree(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	// Another part of the program added a favorite meanwhile.
	f.works[FavoritesListID] = append([]HomeFavorite{{Key: "new", Title: "Brand new"}}, f.works[FavoritesListID]...)
	hlOpen(t, m, "Favorites")
	assert.Contains(t, hlTitles(m.lists.works), "Brand new")
	hlGoto(t, m, 3)
	hlSend(m, press('x', "x"))
	assert.Equal(t, hlTitles(m.lists.works), hlTitles(m.favoriteAll))
}

func TestShiftXAsksThenYClearsTheOpenList(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 4)
	hlSend(m, press('X', "X"))

	assert.Contains(t, screen(m), `Clear all 12 titles from "Weekend marathon"? y to clear, n to keep.`)
	assert.Empty(t, f.calls(), "nothing before the answer")
	hlSend(m, press('y', "y"))
	assert.Equal(t, []string{"Clear(a)"}, f.calls())
	assert.Empty(t, m.lists.works)
	assert.Contains(t, screen(m), "This list is empty.")
	assert.True(t, m.focusInput)
	assert.Equal(t, clearNothing, m.clearing)
}

func TestShiftXOnOneWorkSaysOneWork(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 1)
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Mine")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"))
	assert.Contains(t, screen(m), `Clear all 1 title from "Mine"?`)
}

func TestNAndEscKeepTheWorks(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{press('n', "n"), hlEsc} {
		f := hlThreeLists()
		m := homeListsScreen(t, f, 100, 40)
		hlOpen(t, m, "Weekend")
		hlGoto(t, m, 0)
		hlSend(m, press('X', "X"), key)
		assert.Equal(t, clearNothing, m.clearing)
		assert.Empty(t, f.calls())
		assert.Equal(t, homeLevelWorks, m.lists.level, "esc answers the question, it does not go back")
		assert.Len(t, m.lists.works, 12)
	}
}

func TestStrayKeysDuringTheListClearQuestionTouchNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), hlEnter, keyDown, hlTab, press('x', "x"), press('z', "z"))
	assert.Equal(t, clearListWorks, m.clearing)
	assert.Empty(t, f.calls())
	assert.Equal(t, "FAVORITES", m.scopes[m.scopeIdx].Name)
}

func TestShiftXOnAnEmptyListAsksNothing(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Empty")
	hlSend(m, press('X', "X"))
	assert.Equal(t, clearNothing, m.clearing)
}

func TestShiftXTypesIntoTheBoxWhileItHasTheFocus(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlSend(m, press('X', "X"))
	assert.Equal(t, clearNothing, m.clearing)
	assert.Equal(t, "X", m.input.Value())
}

func TestAClearThatFailsKeepsTheWorksAndSaysWhy(t *testing.T) {
	f := hlThreeLists()
	f.fail["Clear"] = errors.New("the file is read-only")
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('y', "y"))
	assert.Contains(t, screen(m), "Could not clear the list. The file is read-only.")
	assert.Len(t, m.lists.works, 12)
	assert.Equal(t, clearNothing, m.clearing)
}

func TestClearingTheFavoritesListEmptiesTheAllSection(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Favorites")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('y', "y"))
	assert.Equal(t, []string{"Clear(favorites)"}, f.calls())
	assert.Empty(t, m.favoriteAll)
	hlSend(m, hlTab)
	tabTo(t, m, "ALL")
	assert.Contains(t, screen(m), "Nothing here yet")
}

func TestClearingAUserListLeavesTheAllSectionAlone(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('y', "y"))
	assert.Len(t, m.favoriteAll, 3)
}

func TestACompletedClearShowsTheNewCountOnTheLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('y', "y"), hlEsc)
	assert.Contains(t, hlLineWith(screen(m), "Weekend marathon"), "empty")
}

func TestAWorksReadFailureAfterARemovalGoesBackToTheListsWithTheReason(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	f.fail["Works"] = errors.New("flaky")
	hlSend(m, press('x', "x"))
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Contains(t, screen(m), "Could not read your lists. Flaky.")
	assert.Equal(t, "Weekend marathon", hlHighlighted(m))
}

func TestAFavoritesRefreshFailureIsSaidAndTheTabStillUpdates(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpen(t, m, "Favorites")
	hlGoto(t, m, 0)
	f.fail["Works"] = nil
	api := &hlFailSecondWorks{homeListsFake: f}
	m.lists.api = api
	hlSend(m, press('x', "x"))
	assert.Contains(t, screen(m), "Could not read your lists. Second read failed.")
	assert.Len(t, m.lists.works, 2, "the tab itself shows the removal")
	assert.Len(t, m.favoriteAll, 3, "the ALL section keeps the last list it was given")
}

// hlFailSecondWorks answers the first Works call and fails the next ones.
type hlFailSecondWorks struct {
	*homeListsFake
	reads int
}

func (f *hlFailSecondWorks) Works(id string) ([]HomeFavorite, error) {
	f.reads++
	if f.reads == 1 {
		return f.homeListsFake.Works(id)
	}
	return nil, errors.New("second read failed")
}

// ---- scrolling the works -------------------------------------------------------------------

func TestTheWorksScrollAtEverySizeAndKeepTheHighlightInView(t *testing.T) {
	t.Cleanup(func() { SetPlain(false) })
	for _, plain := range []bool{false, true} {
		SetPlain(plain)
		for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
			for _, question := range []bool{false, true} {
				f := newHomeListsFake().withList("a", "Long", 60)
				m := homeListsScreen(t, f, size[0], size[1])
				hlOpen(t, m, "Long")
				hlGoto(t, m, 40)
				if question {
					hlSend(m, press('X', "X"))
				}
				view := hlBody(m)
				label := fmt.Sprintf("%v plain=%v question=%v", size, plain, question)
				assert.Contains(t, view, "Long 40", label)
				assert.Contains(t, view, "more above", label)
				assert.Contains(t, view, "more below", label)
				assert.NotContains(t, view, "Long 0 ", label)
				assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight(), "%s: it fits", label)
			}
		}
	}
}

func TestPagingMovesThroughTheWorks(t *testing.T) {
	f := newHomeListsFake().withList("a", "Long", 60)
	m := homeListsScreen(t, f, 100, 24)
	hlOpen(t, m, "Long")
	hlGoto(t, m, 0)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, homePageRows, m.cursor)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyEnd})
	assert.Equal(t, 59, m.cursor)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyHome})
	assert.Equal(t, 0, m.cursor)
}

func TestTheWorksWindowMatchesTheFavoritesWindow(t *testing.T) {
	// Level 2 reuses the tab's windowing: the same rows show as on a tab
	// without lists holding the same works.
	works := manyFavorites(60)
	f := newHomeListsFake().withList("a", "Long", 0)
	f.works["a"] = works
	withLists := homeListsScreen(t, f, 80, 24)
	hlOpen(t, withLists, "Long")
	hlGoto(t, withLists, 33)
	plain := newSearchHomeModel((&favoritesFake{}).options(works, nil, nil))
	hlSend(plain, tea.WindowSizeMsg{Width: 80, Height: 24})
	tabTo(t, plain, "FAVORITES")
	hlGoto(t, plain, 33)

	a, b := strings.Split(hlBody(withLists), "\n"), strings.Split(hlBody(plain), "\n")
	require.Equal(t, len(b), len(a))
	assert.Equal(t, b[len(b)-4:], a[len(a)-4:], "the last rows and the arrows are the same")
}
