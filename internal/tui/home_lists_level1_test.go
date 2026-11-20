package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hlThreeLists is Favorites with 3 titles, a list of 12, one of 1 and an empty one.
func hlThreeLists() *homeListsFake {
	return newHomeListsFake().withFavorites(3).
		withList("a", "Weekend marathon", 12).
		withList("b", "Short", 1).
		withList("c", "Empty one", 0)
}

func hlHintDescs(m *searchHomeModel) map[string]string {
	out := map[string]string{}
	for _, h := range m.hints() {
		out[h.Key] = h.Desc
	}
	return out
}

func hlHintOrder(m *searchHomeModel) []string {
	var out []string
	for _, h := range m.hints() {
		out = append(out, h.Key)
	}
	return out
}

// hlLineWith is the screen line that holds text.
func hlLineWith(view, text string) string {
	for line := range strings.SplitSeq(view, "\n") {
		if strings.Contains(line, text) {
			return line
		}
	}
	return ""
}

func hlTitles(ws []HomeFavorite) []string {
	var out []string
	for _, w := range ws {
		out = append(out, w.Title)
	}
	return out
}

func hlBackspaces(m *searchHomeModel, n int) {
	for range n {
		hlSend(m, press(tea.KeyBackspace, ""))
	}
}

// hlHighlighted is the name of the list the highlight is on.
func hlHighlighted(m *searchHomeModel) string {
	return m.lists.all[m.rows[m.cursor].index].Name
}

// ---- level 1: your lists ---------------------------------------------------------------

func TestTheTabOpensAtYourListsWithCreateOnTop(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)

	require.Len(t, m.rows, 5)
	assert.Equal(t, homeRowNewList, m.rows[0].kind)
	for i := 1; i < 5; i++ {
		assert.Equal(t, homeRowList, m.rows[i].kind)
	}
	assert.Equal(t, "Favorites", m.lists.all[m.rows[1].index].Name, "in the order Lists() gave")
	assert.Equal(t, "Empty one", m.lists.all[m.rows[4].index].Name)
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.False(t, m.focusInput, "the highlight starts on a list, not in the box")
	assert.Equal(t, 1, m.cursor)
	assert.Equal(t, "Favorites", m.lists.all[m.rows[m.cursor].index].Name)
	assert.Equal(t, homeListsPlaceholder, m.input.Placeholder)
	assert.Contains(t, m.input.Placeholder, "filter your lists")
}

func TestEachRowSaysItsCountAndDate(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 80, 24)
	view := screen(m)

	assert.Contains(t, view, "YOUR LISTS  4", "the heading counts the lists, not the Create row")
	assert.Contains(t, view, "+ Create a new list")
	assert.Contains(t, hlLineWith(view, "Favorites"), "3 titles")
	assert.Contains(t, hlLineWith(view, "Weekend marathon"), "12 titles")
	assert.Contains(t, hlLineWith(view, "Weekend marathon"), "created 2026-11-10")
	assert.Contains(t, hlLineWith(view, "Short"), "1 title")
	assert.NotContains(t, hlLineWith(view, "Short"), "1 titles")
	assert.Contains(t, hlLineWith(view, "Empty one"), "empty")
	assert.NotContains(t, hlLineWith(view, "Favorites"), "created", "the automatic list was not created by anyone")
}

func TestTheFavoritesRowHasTheStarAndPlainModeDrawsItAsText(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 80, 24)
	assert.Contains(t, hlLineWith(screen(m), "Favorites"), "★ Favorites")
	assert.NotContains(t, hlLineWith(screen(m), "Weekend"), "★")

	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	assert.Contains(t, hlLineWith(screen(m), "Favorites"), "* Favorites")
	assert.NotContains(t, screen(m), "★")
}

func TestAListsDateIsItsCreationDay(t *testing.T) {
	f := hlThreeLists()
	f.order[1].Created = time.Date(2025, 1, 5, 23, 59, 0, 0, time.UTC)
	m := homeListsScreen(t, f, 80, 24)
	assert.Contains(t, hlLineWith(screen(m), "Weekend marathon"), "created 2025-01-05")
}

func TestAListWithNoDateShowsNoStrayWord(t *testing.T) {
	f := hlThreeLists()
	f.order[1].Created = time.Time{}
	m := homeListsScreen(t, f, 80, 24)
	assert.NotContains(t, hlLineWith(screen(m), "Weekend marathon"), "created")
}

func TestListNamesAreDrawnOnOneLine(t *testing.T) {
	f := newHomeListsFake().withList("a", "Two\nlines\x1b[31m", 1)
	m := homeListsScreen(t, f, 80, 24)
	assert.NotContains(t, m.View().Content, "\x1b[31m")
	assert.Contains(t, screen(m), "Two lines")
}

func TestTabbingAwayAndBackReturnsToYourLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Weekend")
	hlSend(m, hlEnter)
	require.Equal(t, homeLevelWorks, m.lists.level)

	hlSend(m, hlTab) // wraps to ALL
	require.Equal(t, "ALL", m.scopes[m.scopeIdx].Name)
	tabTo(t, m, "FAVORITES")

	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Equal(t, 1, m.cursor, "back on Favorites, a visible row")
	assert.False(t, m.focusInput)
	assert.Contains(t, screen(m), "+ Create a new list")
	assert.Len(t, m.rows, 5)
}

func TestEveryVisitReadsTheListsAgain(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	f.withList("d", "Added elsewhere", 2)
	hlSend(m, hlTab)
	tabTo(t, m, "FAVORITES")
	assert.Contains(t, screen(m), "Added elsewhere")
}

func TestOpeningWithTheFavoritesTabByIndexShowsTheLists(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Lists: hlThreeLists(), ScopeIndex: 4})
	assert.Len(t, m.rows, 5)
	assert.Equal(t, homeListsPlaceholder, m.input.Placeholder)
}

func TestTypingFiltersTheListsByName(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "ShO")

	require.Len(t, m.rows, 1, "Create is hidden while filtering")
	assert.Equal(t, "Short", m.lists.all[m.rows[0].index].Name)
	view := screen(m)
	assert.NotContains(t, view, "+ Create")
	assert.Contains(t, view, "YOUR LISTS  1 of 4")
}

func TestTheListFilterNeedsEveryWordAndIgnoresAccentsAndCase(t *testing.T) {
	f := newHomeListsFake().withList("a", "Fim de Semana", 1).withList("b", "Semana", 1)
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "SEMANA fim")
	require.Len(t, m.rows, 1)
	assert.Equal(t, "Fim de Semana", hlHighlightedAt(m, 0))

	m2 := homeListsScreen(t, newHomeListsFake().withList("a", "Ação", 1), 100, 40)
	hlType(m2, "acao")
	assert.Len(t, m2.rows, 1)
}

func hlHighlightedAt(m *searchHomeModel, row int) string {
	return m.lists.all[m.rows[row].index].Name
}

func TestTheListFilterDoesNotLookAtTheDateOrCount(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "2026")
	assert.Empty(t, m.rows)
	assert.Contains(t, screen(m), `No list matches "2026". Delete a letter to see them all.`)
}

func TestCreateComesBackWhenTheFilterIsEmptied(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "short")
	require.Len(t, m.rows, 1)
	hlBackspaces(m, 5)
	assert.Len(t, m.rows, 5)
	assert.Equal(t, homeRowNewList, m.rows[0].kind)
}

func TestEscClearsTheListFilterAndStaysOnTheLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "short")
	hlSend(m, hlEsc)
	assert.Equal(t, "", m.input.Value())
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Len(t, m.rows, 5)
}

func TestEnterInTheBoxOpensTheFirstRow(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "week")
	hlSend(m, hlEnter)
	assert.Equal(t, homeLevelWorks, m.lists.level)
	assert.Equal(t, "a", m.lists.open.ID)

	// With the box focused and nothing typed, enter opens the first list and
	// never the name entry: making a list takes the n key or a highlighted
	// "+ Create a new list".
	m2 := homeListsScreen(t, f, 100, 40)
	hlFocusBox(m2)
	require.True(t, m2.focusInput)
	hlSend(m2, hlEnter)
	assert.Equal(t, homeEntryNone, m2.lists.entry)
	assert.Equal(t, homeLevelWorks, m2.lists.level)
	assert.Equal(t, FavoritesListID, m2.lists.open.ID)
}

func TestEnterNeverCreatesFromTheBox(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlFocusBox(m)
	hlSend(m, hlEnter, hlEsc)
	hlSend(m, hlEnter)
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Empty(t, f.calls())
}

func TestEnterWithOnlyCreateToOpenDoesNothingFromTheBox(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "zzz")
	hlSend(m, hlEnter)
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, homeLevelLists, m.lists.level)
}

func TestEnterWithNothingToOpenDoesNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "zzz")
	hlSend(m, hlEnter)
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, "", m.result.Query)
}

func TestEnterOnAListOpensItsWorks(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, hlEnter)

	assert.Equal(t, homeLevelWorks, m.lists.level)
	assert.Equal(t, "b", m.lists.open.ID)
	assert.Equal(t, []string{"Short 0"}, hlTitles(m.lists.works))
	assert.Contains(t, f.log, "Works(b)")
	assert.Contains(t, screen(m), "Short  1")
	assert.Equal(t, "", m.result.Query, "opening a list never leaves the screen")
}

func TestEnterOnTheAutomaticListOpensIt(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Favorites")
	hlSend(m, hlEnter)
	assert.Equal(t, FavoritesListID, m.lists.open.ID)
	assert.Len(t, m.lists.works, 3)
}

func TestPagingAndEndsMoveThroughManyLists(t *testing.T) {
	m := homeListsScreen(t, hlManyLists(40), 100, 24)
	hlGoto(t, m, 0)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, homePageRows, m.cursor)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyEnd})
	assert.Equal(t, 41, m.cursor, "Create, Favorites and 40 lists, 0-based")
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 41-homePageRows, m.cursor)
	hlSend(m, tea.KeyPressMsg{Code: tea.KeyHome})
	assert.Equal(t, 0, m.cursor)
}

// ---- footer words ---------------------------------------------------------------------

func TestFooterOnCreateSaysEnterCreate(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlGoto(t, m, 0)
	d := hlHintDescs(m)
	assert.Equal(t, "create", d["enter"])
	assert.NotContains(t, d, "x")
	assert.NotContains(t, d, "r")
	assert.Equal(t, "move", d["↑↓"])
	assert.Equal(t, "scope", d["tab"])
}

func TestFooterWhileTheBoxIsFocusedNamesWhatEnterDoes(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlFocusBox(m)
	assert.Equal(t, "open", hlHintDescs(m)["enter"], "with the box focused enter opens, never creates")
	assert.NotContains(t, hlHintDescs(m), "n", "n types into the box")
	hlType(m, "short")
	assert.Equal(t, "open", hlHintDescs(m)["enter"])
}

func TestFooterOnAUserListOffersDeleteAndRename(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	d := hlHintDescs(m)
	assert.Equal(t, "open", d["enter"])
	assert.Equal(t, "delete", d["x"])
	assert.Equal(t, "rename", d["r"])
	assert.NotContains(t, d, "X", "there is no clear-all on this level")
	for _, h := range m.hints() {
		if h.Key == "x" || h.Key == "r" || h.Key == "enter" {
			assert.True(t, h.Essential, h.Key)
		}
	}
	assert.Equal(t, []string{"enter", "↑↓", "tab", "x", "r", "n", "?", "ctrl+c"}, hlHintOrder(m))
}

func TestFooterOnTheAutomaticListOffersNeitherDeleteNorRename(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Favorites")
	d := hlHintDescs(m)
	assert.Equal(t, "open", d["enter"])
	assert.NotContains(t, d, "x")
	assert.NotContains(t, d, "r")
	assert.NotContains(t, d, "X")
}

func TestFooterEssentialsSurviveANarrowTerminal(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 40, 20)
	hlRowIs(t, m, "Short")
	foot := ansi.Strip(m.View().Content)
	last := foot[strings.LastIndex(foot, "\n")+1:]
	assert.Contains(t, last, "enter  open")
	assert.Contains(t, last, "x  delete")
}

func TestPlainFooterSaysUpDn(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	assert.Contains(t, hlHintOrder(m), "up/dn")
}

func TestTheHelpListsTheKeysOfTheLists(t *testing.T) {
	for _, size := range [][2]int{{100, 40}, {80, 24}} {
		m := homeListsScreen(t, hlThreeLists(), size[0], size[1])
		hlGoto(t, m, 1)
		hlSend(m, press('?', "?"))
		require.True(t, m.showHelp)
		view := screen(m)
		assert.Contains(t, view, "back to your lists", "%v", size)
		assert.Contains(t, view, "rename a list", "%v", size)
		assert.Contains(t, view, "delete a list or remove a title", "%v", size)
		hlSend(m, press('?', "?"))
		assert.False(t, m.showHelp)
	}
}

// ---- creating a list --------------------------------------------------------------------

func hlOpenCreate(t *testing.T, m *searchHomeModel) {
	t.Helper()
	hlGoto(t, m, 0)
	hlSend(m, hlEnter)
	require.Equal(t, homeEntryCreate, m.lists.entry)
}

func TestEnterOnCreateOpensTheNameEntryBelowTheTabs(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	view := screen(m)

	assert.Contains(t, view, "Name your list")
	assert.Contains(t, view, "e.g. Weekend marathon", "the placeholder shows an example")
	assert.Contains(t, view, "enter create · esc cancel")
	assert.Greater(t, indexOf(view, "Name your list"), indexOf(view, "in ALL"), "below the tab row")
	assert.Less(t, indexOf(view, "Name your list"), indexOf(view, "YOUR LISTS  4"))
	assert.Equal(t, homeListNameLimit, m.lists.name.CharLimit)
	assert.False(t, m.input.Focused(), "the filter box does not keep a caret beside the entry")
	assert.Equal(t, []string{"enter", "esc", "ctrl+c"}, hlHintOrder(m))
	assert.Equal(t, "create", hlHintDescs(m)["enter"])
	assert.Equal(t, "cancel", hlHintDescs(m)["esc"])
}

func TestTypingInTheEntryDoesNotFilterOrMoveTheLists(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "xyz")
	assert.Equal(t, "xyz", m.lists.name.Value())
	assert.Equal(t, "", m.input.Value())
	assert.Len(t, m.rows, 5)
	assert.Contains(t, screen(m), "xyz")
}

func TestEverySpecialKeyBelongsToTheEntryWhileItIsOpen(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlSend(m, hlTab, press('x', "x"), press('r', "r"), press('?', "?"), tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, homeEntryCreate, m.lists.entry)
	assert.Equal(t, "FAVORITES", m.scopes[m.scopeIdx].Name, "tab did not leave")
	assert.False(t, m.showHelp)
	assert.Nil(t, m.sources)
	assert.Equal(t, "xr?", m.lists.name.Value())
	assert.Empty(t, f.calls())
}

func TestEnterCreatesTheListAndOpensItEmpty(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Trip home")
	hlSend(m, hlEnter)

	assert.Equal(t, []string{"Create(Trip home)"}, f.calls())
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, homeLevelWorks, m.lists.level, "it opens the new list")
	assert.Equal(t, "Trip home", m.lists.open.Name)
	assert.Empty(t, m.lists.works)
	view := screen(m)
	assert.Contains(t, view, "Trip home")
	assert.Contains(t, view, "This list is empty. Search for a title, then press ctrl+l on it to add it here.")
	assert.NotContains(t, view, "Name your list")
	assert.True(t, m.focusInput)
}

func TestCreateSendsWhatWasTypedAndTheListsCleanIt(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "  Many   spaces ")
	hlSend(m, hlEnter)
	assert.Equal(t, []string{"Create(  Many   spaces )"}, f.calls())
	assert.Equal(t, "Many spaces", m.lists.open.Name)
}

func TestTheNewListIsOnTheFirstLevelWhenYouGoBack(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Zeta")
	hlSend(m, hlEnter, hlEsc)

	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Contains(t, screen(m), "YOUR LISTS  5")
	assert.Equal(t, "Zeta", hlHighlighted(m), "the highlight is on the list just left")
}

func TestEachRefusedNameStaysInTheEntryWithTheReason(t *testing.T) {
	cases := []struct{ name, typed, want string }{
		{"empty", "", "Give your list a name."},
		{"taken", "short", "You already have a list with that name."},
		{"reserved", "FAVORITES", "That name is reserved."},
		{"too long", "x", "That name is too long."},
	}
	for _, c := range cases {
		f := hlThreeLists()
		if c.name == "too long" {
			f.fail["Create"] = errors.New("that name is too long")
		}
		m := homeListsScreen(t, f, 100, 40)
		hlOpenCreate(t, m)
		hlTypeName(m, c.typed)
		hlSend(m, hlEnter)

		assert.Equal(t, homeEntryCreate, m.lists.entry, c.name)
		assert.Equal(t, c.typed, m.lists.name.Value(), "%s: the text is kept", c.name)
		assert.Contains(t, screen(m), c.want, c.name)
		assert.Equal(t, homeLevelLists, m.lists.level, c.name)
		assert.Len(t, f.order, 4, "%s: nothing was created", c.name)
	}
}

func TestTheReasonGoesAwayWhenTheNameIsEdited(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlSend(m, hlEnter)
	require.Contains(t, screen(m), "Give your list a name.")
	hlTypeName(m, "a")
	assert.NotContains(t, screen(m), "Give your list a name.")
}

func TestARefusedNameCanBeFixedAndSent(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Short")
	hlSend(m, hlEnter)
	require.Contains(t, screen(m), "already have a list")
	hlTypeName(m, " two")
	hlSend(m, hlEnter)
	assert.Equal(t, []string{"Create(Short)", "Create(Short two)"}, f.calls())
	assert.Equal(t, "Short two", m.lists.open.Name)
}

func TestTheEntryStopsAtFortyCharacters(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, strings.Repeat("ab", 30))
	assert.Len(t, []rune(m.lists.name.Value()), 40)
}

func TestEscCancelsTheEntryAndCreatesNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Never made")
	hlSend(m, hlEsc)

	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Empty(t, f.calls())
	assert.NotContains(t, screen(m), "Name your list")
	assert.Equal(t, homeLevelLists, m.lists.level)
	hlSend(m, hlEsc)
	assert.Equal(t, homeLevelLists, m.lists.level, "a second esc does nothing on this level")
}

func TestTheEntryOpensEmptyEachTime(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "left over")
	hlSend(m, hlEsc)
	hlGoto(t, m, 0)
	hlSend(m, hlEnter)
	assert.Equal(t, "", m.lists.name.Value())
}

func TestTheNKeyOpensTheNameEntryFromAnyHighlightedRow(t *testing.T) {
	for _, row := range []int{0, 1, 2, 4} {
		m := homeListsScreen(t, hlThreeLists(), 100, 40)
		hlGoto(t, m, row)
		hlSend(m, press('n', "n"))
		assert.Equal(t, homeEntryCreate, m.lists.entry, "row %d", row)
		assert.Equal(t, "", m.input.Value(), "the n was a command, not typing")
		hlSend(m, hlEsc)
		assert.Equal(t, row, m.cursor, "esc leaves the highlight where it was")
	}
}

func TestTheNKeyTypesIntoTheBoxWhileItHasTheFocus(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlFocusBox(m)
	hlSend(m, press('n', "n"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, "n", m.input.Value())
}

func TestTheNKeyDoesNothingInsideAList(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('n', "n"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, "n", m.input.Value(), "inside a list n is just a letter for the filter")
}

func TestCreateFromTheBoxGivesTheCaretBackAfterEsc(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlFocusBox(m)
	require.True(t, m.focusInput)
	m.openEntry(homeEntryCreate, HomeList{})
	hlSend(m, hlEsc)
	assert.True(t, m.focusInput)
	assert.True(t, m.input.Focused())
	hlType(m, "short")
	assert.Len(t, m.rows, 1)
}

func TestCreateFromAHighlightedRowKeepsTheHighlightAfterEsc(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlSend(m, hlEsc)
	assert.False(t, m.focusInput)
	assert.False(t, m.input.Focused())
	assert.Equal(t, 0, m.cursor)
}

func TestCtrlCQuitsFromTheEntry(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}

func TestACreatedListThatCannotBeReadOpensTheListsWithTheReason(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Fresh")
	f.fail["Works"] = errors.New("disk is full")
	hlSend(m, hlEnter)

	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Contains(t, screen(m), "Could not read your lists. Disk is full.")
	assert.Equal(t, "Fresh", hlHighlighted(m), "the new list is highlighted")
}

// ---- renaming ---------------------------------------------------------------------------

func TestRenameOpensTheEntryHoldingTheName(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('r', "r"))

	assert.Equal(t, homeEntryRename, m.lists.entry)
	assert.Equal(t, "Short", m.lists.name.Value())
	view := screen(m)
	assert.Contains(t, view, "Name your list")
	assert.Contains(t, view, "enter rename · esc cancel")
	assert.Equal(t, "rename", hlHintDescs(m)["enter"])
}

func TestRenameSendsTheListAndTheNewName(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('r', "r"))
	hlBackspaces(m, 5)
	hlTypeName(m, "Tiny")
	hlSend(m, hlEnter)

	assert.Equal(t, []string{"Rename(b,Tiny)"}, f.calls())
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, homeLevelLists, m.lists.level, "renaming stays on the lists")
	assert.Equal(t, "Tiny", hlHighlighted(m), "the highlight follows the renamed list")
	assert.False(t, m.focusInput)
	for _, l := range m.lists.all {
		assert.NotEqual(t, "Short", l.Name)
	}
}

func TestRenameRefusedStaysOpenWithTheTextKept(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('r', "r"))
	hlBackspaces(m, 5)
	hlTypeName(m, "weekend marathon")
	hlSend(m, hlEnter)

	assert.Equal(t, homeEntryRename, m.lists.entry)
	assert.Equal(t, "weekend marathon", m.lists.name.Value())
	assert.Contains(t, screen(m), "You already have a list with that name.")
	assert.Equal(t, "Short", f.order[2].Name)
}

func TestEscCancelsARenameWithoutCalling(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('r', "r"))
	hlTypeName(m, "zzz")
	hlSend(m, hlEsc)
	assert.Empty(t, f.calls())
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Contains(t, screen(m), "Short")
}

func TestRenameWhileFilteringFindsTheListAgain(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "short")
	hlGoto(t, m, 0)
	hlSend(m, press('r', "r"))
	hlBackspaces(m, 5)
	hlTypeName(m, "Brief")
	hlSend(m, hlEnter)
	assert.Equal(t, "", m.input.Value(), "the filter would hide the new name")
	assert.Equal(t, "Brief", hlHighlighted(m))
}

func TestRenameOnTheAutomaticListSaysWhyInWords(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Favorites")
	hlSend(m, press('r', "r"))

	assert.Equal(t, homeEntryNone, m.lists.entry, "no entry opens")
	assert.Contains(t, screen(m), "The favorites list cannot be renamed.")
	assert.Equal(t, "Favorites", f.order[0].Name)
	assert.Equal(t, []string{"Rename(favorites,Favorites)"}, f.calls(), "the same name: nothing could change")
}

type hlNoRenameError struct{ *homeListsFake }

func (hlNoRenameError) Rename(string, string) error { return nil }

func TestRenameOnTheAutomaticListHasAFallbackSentence(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Favorites")
	m.lists.api = hlNoRenameError{m.lists.api.(*homeListsFake)}
	hlSend(m, press('r', "r"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Contains(t, screen(m), "The Favorites list cannot be renamed.")
}

func TestTheNoticeGoesAtTheNextKey(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Favorites")
	hlSend(m, press('x', "x"))
	require.Contains(t, screen(m), "cannot be deleted")
	hlSend(m, keyDown)
	assert.NotContains(t, screen(m), "cannot be deleted")
}

func TestRAndEAndXOnCreateDoNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlGoto(t, m, 0)
	hlSend(m, press('r', "r"), press('e', "e"), press('x', "x"), press('X', "X"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, clearNothing, m.clearing)
	assert.Empty(t, f.calls())
	assert.Equal(t, "", m.input.Value(), "the letters were commands, not typing")
}

func TestEAndShiftXOnAListDoNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('e', "e"), press('X', "X"))
	assert.Equal(t, clearNothing, m.clearing)
	assert.Empty(t, f.calls())
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Equal(t, "", m.result.Query)
}

func TestLettersTypeIntoTheBoxWhileItHasTheFocus(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "rex")
	assert.Equal(t, "rex", m.input.Value())
	assert.Equal(t, clearNothing, m.clearing)
	assert.Equal(t, homeEntryNone, m.lists.entry)
}

// ---- deleting ----------------------------------------------------------------------------

func TestXAsksBeforeDeletingAndNamesTheCount(t *testing.T) {
	cases := map[string]string{
		"Weekend marathon": `Delete the list "Weekend marathon"? The 12 titles stay available, only the list goes. y to delete, n to keep.`,
		"Short":            `Delete the list "Short"? The title stays available, only the list goes. y to delete, n to keep.`,
		"Empty one":        `Delete the empty list "Empty one"? y to delete, n to keep.`,
	}
	for name, want := range cases {
		f := hlThreeLists()
		m := homeListsScreen(t, f, 100, 40)
		hlRowIs(t, m, name)
		hlSend(m, press('x', "x"))
		assert.Contains(t, strings.Join(strings.Fields(screen(m)), " "), want, name)
		assert.Empty(t, f.calls(), "%s: nothing is deleted before the answer", name)
		assert.Equal(t, "delete", hlHintDescs(m)["y"])
		assert.Equal(t, "keep", hlHintDescs(m)["n"])
		assert.Equal(t, []string{"y", "n", "ctrl+c"}, hlHintOrder(m))
	}
}

func TestYDeletesTheList(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"), press('y', "y"))

	assert.Equal(t, []string{"Delete(b)"}, f.calls())
	assert.Equal(t, clearNothing, m.clearing)
	view := screen(m)
	assert.NotContains(t, view, "Short")
	assert.Contains(t, view, "YOUR LISTS  3")
	assert.Equal(t, "Empty one", hlHighlighted(m), "the next list slides under the highlight")
}

func TestDeletingTheLastListMovesTheHighlightUp(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Empty one")
	hlSend(m, press('x', "x"), press('y', "y"))
	assert.Equal(t, "Short", hlHighlighted(m))
	assert.False(t, m.focusInput)
}

func TestNAndEscKeepTheUserList(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{press('n', "n"), hlEsc} {
		f := hlThreeLists()
		m := homeListsScreen(t, f, 100, 40)
		hlRowIs(t, m, "Short")
		hlSend(m, press('x', "x"), key)
		assert.Equal(t, clearNothing, m.clearing)
		assert.Empty(t, f.calls())
		assert.Contains(t, screen(m), "Short")
		assert.NotContains(t, screen(m), "Delete the list")
	}
}

func TestStrayKeysDuringTheDeleteQuestionTouchNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"), hlEnter, keyDown, hlTab, press('z', "z"), press('Z', "Z"), press('r', "r"))
	assert.Equal(t, clearDeleteList, m.clearing)
	assert.Equal(t, "FAVORITES", m.scopes[m.scopeIdx].Name)
	assert.Empty(t, f.calls())
	assert.Equal(t, "", m.input.Value())
}

func TestCtrlCQuitsFromTheDeleteQuestion(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"))
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}

func TestXOnTheAutomaticListSaysItCannotBeDeletedAndAsksNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Favorites")
	hlSend(m, press('x', "x"))

	assert.Equal(t, clearNothing, m.clearing)
	assert.Contains(t, screen(m), "The Favorites list cannot be deleted.")
	hlSend(m, press('y', "y"))
	assert.Empty(t, f.calls(), "Delete is never called for it")
}

func TestAFailedDeleteKeepsTheListAndSaysWhy(t *testing.T) {
	f := hlThreeLists()
	f.fail["Delete"] = errors.New("the file is read-only")
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"), press('y', "y"))
	view := screen(m)
	assert.Contains(t, view, "Could not delete the list. The file is read-only.")
	assert.Contains(t, view, "Short")
	assert.Equal(t, clearNothing, m.clearing)
}

func TestXWhileFilteringDeletesTheListHighlightedNotTheOneAtThatPosition(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "empty")
	hlGoto(t, m, 0)
	hlSend(m, press('x', "x"), press('y', "y"))
	assert.Equal(t, []string{"Delete(c)"}, f.calls())
}

func TestDeletingAUserListDoesNotTouchTheAllSection(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	before := slices.Clone(m.favoriteAll)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"), press('y', "y"))
	assert.Equal(t, before, m.favoriteAll)
	assert.NotContains(t, f.log, "Works(favorites)")
}

// ---- the details panel ---------------------------------------------------------------------

func TestTheWidePanelDescribesTheHighlightedList(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Weekend")
	view := screen(m)

	assert.Contains(t, view, "LIST")
	assert.Contains(t, view, "Titles")
	assert.Contains(t, view, "Created")
	assert.Contains(t, view, "2026-11-10")
	for i := range 8 {
		assert.Contains(t, view, fmt.Sprintf("Weekend marathon %d", i))
	}
	assert.NotContains(t, view, "Weekend marathon 8", "only the first eight")
	assert.Contains(t, view, "+4 more")
	assert.Contains(t, view, "enter opens it")
}

func TestThePanelFollowsTheHighlight(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	view := screen(m)
	assert.Contains(t, view, "Short 0")
	assert.NotContains(t, view, "Weekend marathon 0")
	assert.NotContains(t, view, "more")
}

func TestThePanelOfTheAutomaticListSaysAlwaysThere(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Favorites")
	view := screen(m)
	assert.Contains(t, view, "Always there")
	assert.Contains(t, view, "Fav 0")
}

func TestThePanelCountsExactlyEightWithoutAMoreLine(t *testing.T) {
	f := newHomeListsFake().withList("a", "Eight", 8).withList("b", "Nine", 9)
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Eight")
	assert.NotContains(t, screen(m), "more")
	assert.Contains(t, screen(m), "Eight 7")
	hlRowIs(t, m, "Nine")
	assert.Contains(t, screen(m), "+1 more")
	assert.NotContains(t, screen(m), "Nine 8")
}

func TestThePanelOfAnEmptyListSaysSo(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Empty one")
	view := screen(m)
	assert.Contains(t, view, "Nothing in it yet.")
	assert.Contains(t, view, "enter opens it")
}

func TestThereIsNoPanelOnCreateOrOnNarrowTerminals(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlGoto(t, m, 0)
	assert.NotContains(t, screen(m), "enter opens it")
	assert.NotContains(t, screen(m), "Created  ")

	for _, w := range []int{99, 80, 60} {
		n := homeListsScreen(t, hlThreeLists(), w, 40)
		hlRowIs(t, n, "Weekend")
		view := screen(n)
		assert.NotContains(t, view, "enter opens it", "width %d", w)
		assert.Contains(t, view, "12 titles", "width %d: the row carries the count", w)
		assert.Contains(t, view, "created 2026-11-10", "width %d: and the date", w)
	}
}

func TestThePanelShowsWhileTheBoxHasTheFocusOnAMatch(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlType(m, "weekend")
	assert.Contains(t, screen(m), "Weekend marathon 0")
}

func TestThePlainPanelIsTextBelowTheList(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	for _, size := range [][2]int{{100, 40}, {80, 24}, {60, 20}} {
		m := homeListsScreen(t, hlThreeLists(), size[0], size[1])
		hlRowIs(t, m, "Weekend")
		view := hlBody(m)

		require.Contains(t, view, "\nLIST\n", "%v", size)
		assert.Contains(t, view, "Titles: 12", "%v", size)
		assert.Contains(t, view, "Created: 2026-11-10", "%v", size)
		assert.Contains(t, view, "enter opens it", "%v", size)
		if size[1] >= 24 {
			assert.Contains(t, view, "Weekend marathon 0", "%v", size)
		}
		assert.Greater(t, indexOf(view, "\nLIST\n"), indexOf(view, "Weekend marathon  "), "%v: below the rows", size)
		assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight(), "%v: it fits", size)
		assert.Contains(t, hlLineWith(view, "Weekend marathon  "), "> ", "%v: the highlight is drawn", size)
		assert.NotContains(t, view, "█", "%v", size)
		assert.NotContains(t, view, "⠁", "%v: no braille", size)
	}
}

func TestThePlainPanelNamesFewerTitlesWhenTheScreenIsShort(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	m := homeListsScreen(t, hlThreeLists(), 60, 20)
	hlRowIs(t, m, "Weekend")
	view := hlBody(m)
	assert.Contains(t, view, "more")
	assert.Less(t, strings.Count(view, "Weekend marathon "), 9, "not all of its titles")
	assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight())
}

func TestThePanelNamesFewerTitlesOnAShortWideScreen(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 24)
	hlRowIs(t, m, "Weekend")
	view := hlBody(m)
	assert.Contains(t, view, "enter opens it", "a short screen still gets its panel")
	assert.Contains(t, view, "Weekend marathon 0")
	assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight())
}

func TestAPanelThatCannotFitIsLeftOut(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 12)
	hlRowIs(t, m, "Weekend")
	view := hlBody(m)
	assert.NotContains(t, view, "enter opens it")
	assert.Contains(t, view, "Weekend marathon")
}

func TestListPanelContent(t *testing.T) {
	l := HomeList{Name: "Mine", Count: 3, Created: time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC), Titles: []string{"A", "B", "C"}}
	c := homeListPanelContent(&l, 2)
	assert.Equal(t, "List", c.label)
	assert.Equal(t, "Mine", c.title)
	assert.Equal(t, []detailField{{"Titles", "3"}, {"Created", "2026-03-04"}}, c.fields)
	assert.Equal(t, []string{"A", "B", "+1 more"}, c.lines)
	assert.Equal(t, "enter opens it", c.tip)

	l.Count, l.Created = 0, time.Time{}
	l.Titles = nil
	c = homeListPanelContent(&l, 8)
	assert.Equal(t, "Unknown", c.fields[1].value, "a list with no date")
	assert.Equal(t, []string{"Nothing in it yet."}, c.lines)

	auto := HomeList{Name: "Favorites", Auto: true, Count: 12, Titles: []string{"A"}}
	c = homeListPanelContent(&auto, 8)
	assert.Equal(t, "Always there", c.fields[1].value)
	assert.Equal(t, []string{"A", "+11 more"}, c.lines, "Count wins when the titles were capped")

	none := HomeList{Name: "X", Count: 2, Titles: []string{"A", "B"}}
	assert.Equal(t, []string(nil), homeListPanelContent(&none, 0).lines, "no titles shown: no lone +N line")
}

// ---- scrolling and size -------------------------------------------------------------------

func hlManyLists(n int) *homeListsFake {
	f := newHomeListsFake().withFavorites(2)
	for i := range n {
		f.withList(fmt.Sprintf("l%d", i), fmt.Sprintf("Listing %02d", i), i%5)
	}
	return f
}

func TestTheHighlightedListIsAlwaysOnScreen(t *testing.T) {
	t.Cleanup(func() { SetPlain(false) })
	for _, plain := range []bool{false, true} {
		for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}, {100, 24}} {
			for _, mode := range []string{"plain", "question", "entry", "notice"} {
				SetPlain(plain)
				m := homeListsScreen(t, hlManyLists(40), size[0], size[1])
				hlRowIs(t, m, "Listing 25")
				switch mode {
				case "question":
					hlSend(m, press('x', "x"))
				case "entry":
					hlSend(m, press('r', "r"))
				case "notice":
					m.notice = "Could not do a thing. Nothing happened."
				}
				view := hlBody(m)
				label := fmt.Sprintf("%v plain=%v %s", size, plain, mode)
				if mode == "entry" {
					assert.Contains(t, view, "Name your list", label)
				}
				assert.Contains(t, view, "Listing 25", label)
				assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight(), "%s: the body fits", label)
				assert.Contains(t, view, "more above", label)
				assert.Contains(t, view, "more below", label)
			}
		}
	}
}

func TestTheFirstAndLastListsAreReachable(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
		m := homeListsScreen(t, hlManyLists(40), size[0], size[1])
		hlGoto(t, m, 0)
		assert.Contains(t, hlBody(m), "+ Create a new list")
		assert.NotContains(t, hlBody(m), "more above")
		hlSend(m, tea.KeyPressMsg{Code: tea.KeyEnd})
		view := hlBody(m)
		assert.Contains(t, view, "Listing 39", "%v", size)
		assert.NotContains(t, view, "more below", "%v", size)
		assert.Contains(t, view, "more above", "%v", size)
	}
}

func TestAnyRowCanBeHighlightedAndSeenAtEverySize(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
		m := homeListsScreen(t, hlManyLists(25), size[0], size[1])
		for i := range m.rows {
			hlGoto(t, m, i)
			want := "+ Create a new list"
			if i > 0 {
				want = hlHighlighted(m)
			}
			assert.Contains(t, hlBody(m), want, "%v row %d", size, i)
			assert.LessOrEqual(t, strings.Count(hlBody(m), "\n")+1, m.shellHeight(), "%v row %d", size, i)
		}
	}
}

func TestTheWindowCentersTheHighlightAndCountsWhatIsHidden(t *testing.T) {
	m := homeListsScreen(t, hlManyLists(40), 80, 24)
	hlRowIs(t, m, "Listing 20")
	view := hlBody(m)
	above, below := 0, 0
	_, err := fmt.Sscanf(hlLineWith(view, "more above"), "  ↑ %d more above", &above)
	require.NoError(t, err)
	_, err = fmt.Sscanf(hlLineWith(view, "more below"), "  ↓ %d more below", &below)
	require.NoError(t, err)
	shown := strings.Count(view, "Listing ")
	assert.Equal(t, len(m.rows), above+shown+below, "every row is above, shown or below")
	assert.Equal(t, m.cursor-shown/2, above, "the highlight sits in the middle of the window")
}

func TestThePanelNeverPushesTheBodyPastTheTerminal(t *testing.T) {
	for _, size := range [][2]int{{100, 40}, {100, 24}, {120, 30}, {100, 16}} {
		m := homeListsScreen(t, hlManyLists(30), size[0], size[1])
		hlRowIs(t, m, "Listing 04")
		assert.LessOrEqual(t, strings.Count(hlBody(m), "\n")+1, m.shellHeight(), "%v", size)
		assert.Contains(t, hlBody(m), "Listing 04", "%v", size)
	}
}

func TestATinyTerminalStillShowsTheHighlightedList(t *testing.T) {
	m := homeListsScreen(t, hlManyLists(30), 40, 8)
	hlRowIs(t, m, "Listing 12")
	assert.Contains(t, hlBody(m), "▶", "the highlighted row is on screen even if its title is cut")
	assert.Contains(t, hlBody(m), "List")
}

// ---- errors ---------------------------------------------------------------------------------

func TestWhenTheListsCannotBeReadTheTabSaysSo(t *testing.T) {
	f := hlThreeLists()
	f.fail["Lists"] = errors.New("the saved file is damaged")
	m := homeListsScreen(t, f, 100, 40)
	view := screen(m)

	assert.Contains(t, view, "Could not read your lists. The saved file is damaged.")
	assert.NotContains(t, view, "+ Create", "no way to make a list that could not be checked")
	assert.Empty(t, m.rows)
	assert.Contains(t, view, "tab back here to try again")
}

func TestTheScreenStaysUsableAfterAFailedRead(t *testing.T) {
	f := hlThreeLists()
	f.fail["Lists"] = errors.New("the saved file is damaged")
	m := homeListsScreen(t, f, 100, 40)

	hlSend(m, hlEnter, keyDown, hlEsc) // nothing to act on, and nothing breaks
	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Equal(t, homeEntryNone, m.lists.entry)

	delete(f.fail, "Lists")
	hlSend(m, hlTab)
	tabTo(t, m, "FAVORITES")
	assert.Contains(t, screen(m), "Weekend marathon")
	assert.NotContains(t, screen(m), "Could not read")
}

func TestTabbingAwayFromAFailedReadWorks(t *testing.T) {
	f := hlThreeLists()
	f.fail["Lists"] = errors.New("boom")
	m := homeListsScreen(t, f, 100, 40)
	hlSend(m, hlTab)
	assert.Equal(t, "ALL", m.scopes[m.scopeIdx].Name)
	assert.NotContains(t, screen(m), "Could not read")
}

func TestAListThatCannotBeReadStaysClosedWithTheReason(t *testing.T) {
	f := hlThreeLists()
	f.fail["Works"] = errors.New("the saved file is damaged")
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, hlEnter)

	assert.Equal(t, homeLevelLists, m.lists.level)
	assert.Contains(t, screen(m), "Could not read your lists. The saved file is damaged.")
	assert.Contains(t, screen(m), "Weekend marathon", "the lists are still there")
}

func TestAReadFailureAfterAMutationKeepsTheKnownLists(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"))
	f.fail["Lists"] = errors.New("flaky")
	hlSend(m, press('y', "y"))
	view := screen(m)
	assert.Contains(t, view, "Could not read your lists. Flaky.")
	assert.Contains(t, view, "Weekend marathon")
	assert.NotContains(t, view, "+ Create", "Create is hidden while the lists are in doubt")
}

func TestAnErrorWithNoTextStillSaysSomething(t *testing.T) {
	f := hlThreeLists()
	f.fail["Create"] = errors.New("")
	m := homeListsScreen(t, f, 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "x")
	hlSend(m, hlEnter)
	assert.Contains(t, screen(m), "Something went wrong.")
}
