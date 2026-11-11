package tui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The beginner round: words, the first highlight, the n key, Caps Lock, the
// counter and the limits.

func TestTheHeadingIsYourListsNotFavorites(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	view := screen(m)
	assert.Contains(t, view, "YOUR LISTS  4")
	assert.NotContains(t, view, "FAVORITES  4", "it would read as four favorites")
}

func TestTheWordIsTitlesEverywhere(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	view := screen(m)
	for _, bad := range []string{" works", " work ", "Works"} {
		assert.NotContains(t, view, bad)
	}
	hlOpen(t, m, "Weekend")
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"))
	assert.Contains(t, hlFlat(screen(m)), `Clear all 12 titles from "Weekend marathon"?`)
}

// hlFlat is text with its line breaks and runs of spaces made single spaces.
func hlFlat(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestTheFavoritesRowSaysWhatTheListIs(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 80, 24)
	assert.Contains(t, hlLineWith(screen(m), "Favorites"), "what you save with ctrl+f")
	assert.NotContains(t, hlLineWith(screen(m), "Weekend"), "ctrl+f")

	narrow := homeListsScreen(t, hlThreeLists(), 60, 20)
	assert.NotContains(t, screen(narrow), "what you save", "no room on a narrow row")
	assert.Contains(t, hlLineWith(screen(narrow), "Favorites"), "3 titles")

	edge := homeListsScreen(t, hlThreeLists(), 70, 24)
	assert.Contains(t, screen(edge), "what you save with ctrl+f", "from 70 columns")
	below := homeListsScreen(t, hlThreeLists(), 69, 24)
	assert.NotContains(t, screen(below), "what you save with ctrl+f")
}

func TestTheTabOpensOnTheFavoritesRowSoKeysActOnWhatIsSeen(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 80, 24)
	assert.False(t, m.focusInput)
	assert.Contains(t, hlLineWith(screen(m), "Favorites"), "▶", "the cursor mark is on Favorites")
	assert.Equal(t, "open", hlHintDescs(m)["enter"])

	hlSend(m, hlEnter)
	assert.Equal(t, FavoritesListID, m.lists.open.ID)
	assert.Empty(t, f.calls())
}

func TestTheFirstHighlightIsTheFirstListEvenWithATypedFilterLeftOver(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 80, 24)
	hlType(m, "short")
	hlSend(m, hlTab)
	tabTo(t, m, "FAVORITES")
	assert.Equal(t, "Short", hlHighlighted(m))
	assert.False(t, m.focusInput)
}

func TestWithNoListsAtAllTheBoxKeepsTheFocus(t *testing.T) {
	f := newHomeListsFake()
	f.order = nil
	m := homeListsScreen(t, f, 80, 24)
	assert.True(t, m.focusInput, "nothing to highlight but Create")
	assert.Equal(t, homeRowNewList, m.rows[0].kind)
}

func TestRAndXOnCreateAskToPickAList(t *testing.T) {
	for _, key := range []string{"r", "x"} {
		f := hlThreeLists()
		m := homeListsScreen(t, f, 100, 40)
		hlGoto(t, m, 0)
		hlSend(m, press(rune(key[0]), key))
		assert.Contains(t, screen(m), "Pick a list first.", key)
		assert.Empty(t, f.calls())
		assert.Equal(t, homeEntryNone, m.lists.entry)
	}
}

func TestEAndShiftXOnCreateStaySilent(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlGoto(t, m, 0)
	hlSend(m, press('e', "e"), press('X', "X"))
	assert.NotContains(t, screen(m), "Pick a list first.")
}

// ---- the n key and its hint --------------------------------------------------------------

func TestTheFooterOffersNewListWhileARowIsHighlighted(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	for _, row := range []int{0, 1, 2} {
		hlGoto(t, m, row)
		assert.Equal(t, "new list", hlHintDescs(m)["n"], "row %d", row)
	}
	hlFocusBox(m)
	assert.NotContains(t, hlHintDescs(m), "n")
}

func TestTheNewListHintIsNotEssentialAndSitsBeforeHelp(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	order := hlHintOrder(m)
	assert.Less(t, indexOfString(order, "r"), indexOfString(order, "n"))
	assert.Less(t, indexOfString(order, "n"), indexOfString(order, "?"))
	assert.False(t, hintsByKey(m)["n"].Essential)
}

func indexOfString(in []string, s string) int {
	for i, v := range in {
		if v == s {
			return i
		}
	}
	return -1
}

func TestTheHelpNamesTheNKey(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlSend(m, press('?', "?"))
	assert.Contains(t, screen(m), "make a new list")
}

// ---- Caps Lock ------------------------------------------------------------------------------

func TestCapitalYAndNAnswerEveryQuestion(t *testing.T) {
	// delete a list
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"), press('Y', "Y"))
	assert.Equal(t, []string{"Delete(b)"}, f.calls())

	// keep a list
	f2 := hlThreeLists()
	m2 := homeListsScreen(t, f2, 100, 40)
	hlRowIs(t, m2, "Short")
	hlSend(m2, press('x', "x"), press('N', "N"))
	assert.Empty(t, f2.calls())
	assert.Equal(t, clearNothing, m2.clearing)

	// clear a list's titles
	f3 := hlThreeLists()
	m3 := homeListsScreen(t, f3, 100, 40)
	hlOpen(t, m3, "Weekend")
	hlGoto(t, m3, 0)
	hlSend(m3, press('X', "X"), press('Y', "Y"))
	assert.Equal(t, []string{"Clear(a)"}, f3.calls())
}

func TestCapitalYConfirmsTheOldClearQuestionToo(t *testing.T) {
	m, f := hlPlainHome(t, 100, 40)
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('Y', "Y"))
	assert.Equal(t, 1, f.cleared)
	assert.Empty(t, m.favoriteAll)
}

func TestCapitalNKeepsWhatTheOldClearQuestionAsked(t *testing.T) {
	m, f := hlPlainHome(t, 100, 40)
	hlGoto(t, m, 0)
	hlSend(m, press('X', "X"), press('N', "N"))
	assert.Zero(t, f.cleared)
	assert.Len(t, m.favoriteAll, 2)
	assert.Equal(t, clearNothing, m.clearing)
}

// ---- the empty states ---------------------------------------------------------------------

func TestWithNoListsOfYourOwnTheTabSaysWhatToDoNext(t *testing.T) {
	for _, size := range [][2]int{{100, 40}, {80, 24}, {60, 20}} {
		f := newHomeListsFake().withFavorites(2)
		m := homeListsScreen(t, f, size[0], size[1])
		view := hlFlat(hlBody(m))
		assert.Contains(t, view, `You have no lists of your own yet. Press n to make one, for example "Weekend marathon".`, "%v", size)
		assert.LessOrEqual(t, strings.Count(hlBody(m), "\n")+1, m.shellHeight(), "%v", size)
	}
}

func TestTheTeachingLineGoesWithTheFirstList(t *testing.T) {
	f := newHomeListsFake().withList("a", "Mine", 1)
	m := homeListsScreen(t, f, 100, 40)
	assert.NotContains(t, hlFlat(screen(m)), "You have no lists of your own yet")
}

func TestTheTeachingLineStaysQuietWhileFilteringOrAfterAFailedRead(t *testing.T) {
	f := newHomeListsFake()
	m := homeListsScreen(t, f, 100, 40)
	hlType(m, "fav")
	assert.NotContains(t, screen(m), "You have no lists of your own yet")

	f2 := newHomeListsFake()
	f2.fail["Lists"] = fmt.Errorf("boom")
	m2 := homeListsScreen(t, f2, 100, 40)
	assert.NotContains(t, screen(m2), "You have no lists of your own yet")
}

func TestTheTeachingLineDoesNotHideTheHighlightedRowOnASmallScreen(t *testing.T) {
	f := newHomeListsFake()
	m := homeListsScreen(t, f, 60, 12)
	assert.Contains(t, hlBody(m), "Favorites")
}

// ---- the name entry: counter and limits ------------------------------------------------------

func TestTheEntryCountsWhatIsTyped(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	assert.Contains(t, screen(m), "0/40")
	hlTypeName(m, "Summer trip!")
	assert.Contains(t, screen(m), "12/40")
	assert.NotContains(t, screen(m), "longest")
}

func TestHomeListsCounterCountsLettersNotBytes(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, "Ação")
	assert.Contains(t, screen(m), "4/40")
}

func TestAtTheLimitTheEntryStopsAndSaysSoInWords(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlTypeName(m, strings.Repeat("a", 45))
	assert.Len(t, []rune(m.lists.name.Value()), 40)
	view := screen(m)
	assert.Contains(t, view, "40/40")
	assert.Contains(t, view, "that is the longest a name can be")
}

func TestTheCounterShowsOnARenameToo(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('r', "r"))
	assert.Contains(t, screen(m), "5/40")
}

func TestTheCounterFitsASmallScreen(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 60, 20)
	hlOpenCreate(t, m)
	hlTypeName(m, strings.Repeat("b", 40))
	view := screen(m)
	assert.Contains(t, view, "40/40")
	for _, line := range strings.Split(view, "\n") {
		assert.LessOrEqual(t, lipgloss.Width(line), 60, line)
	}
}

func hlFiftyLists(own int) *homeListsFake {
	f := newHomeListsFake().withFavorites(1)
	for i := range own {
		f.withList(fmt.Sprintf("l%d", i), fmt.Sprintf("List %02d", i), 0)
	}
	return f
}

func TestAtFiftyListsTheTabSaysSoBeforeAskingForAName(t *testing.T) {
	f := hlFiftyLists(homeListMax)
	m := homeListsScreen(t, f, 100, 40)
	assert.Contains(t, screen(m), "you have 50, the most there can be", "on the Create row")

	hlGoto(t, m, 0)
	hlSend(m, hlEnter)
	assert.Equal(t, homeEntryNone, m.lists.entry, "no name is asked for")
	assert.Contains(t, hlFlat(screen(m)), "You have 50 lists, the most there can be. Delete one to make a new one.")

	hlSend(m, press('n', "n"))
	assert.Equal(t, homeEntryNone, m.lists.entry)
	assert.Empty(t, f.calls())
}

func TestJustUnderFiftyListsStillOpensTheEntry(t *testing.T) {
	f := hlFiftyLists(homeListMax - 1)
	m := homeListsScreen(t, f, 100, 40)
	assert.NotContains(t, screen(m), "the most there can be")
	hlSend(m, press('n', "n"))
	assert.Equal(t, homeEntryCreate, m.lists.entry)
}

func TestFavoritesDoesNotCountTowardsFifty(t *testing.T) {
	m := homeListsScreen(t, hlFiftyLists(homeListMax-1), 100, 40)
	assert.False(t, m.listsFull())
	m2 := homeListsScreen(t, hlFiftyLists(homeListMax), 100, 40)
	assert.True(t, m2.listsFull())
}

// ---- display width ------------------------------------------------------------------------------

var hlWideName = strings.Repeat("漢", 40)

func TestFortyWideCharactersFitEverySizeOnTheListsAndInsideTheList(t *testing.T) {
	require.Equal(t, 40, len([]rune(hlWideName)))
	require.Equal(t, 80, lipgloss.Width(hlWideName))
	for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
		f := newHomeListsFake().withFavorites(2).withList("w", hlWideName, 3)
		m := homeListsScreen(t, f, size[0], size[1])
		hlRowIs(t, m, "漢漢")
		for _, line := range strings.Split(screen(m), "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), size[0], "%v lists: %q", size, line)
		}
		hlSend(m, press('x', "x"))
		for _, line := range strings.Split(screen(m), "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), size[0], "%v question: %q", size, line)
		}
		hlSend(m, press('n', "n"))
		hlSend(m, hlEnter)
		hlSend(m, hlEsc)
		hlSend(m, hlEnter) // opens the list
		for _, line := range strings.Split(screen(m), "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), size[0], "%v inside: %q", size, line)
		}
		assert.Contains(t, hlBody(m), "漢漢", "%v the name is cut, not dropped", size)
	}
}

func TestTheOpenListHeadingIsCutByDisplayWidth(t *testing.T) {
	f := newHomeListsFake().withList("w", hlWideName, 3)
	m := homeListsScreen(t, f, 60, 20)
	hlOpen(t, m, "漢漢")
	head := hlLineWith(hlBody(m), "漢漢")
	assert.LessOrEqual(t, lipgloss.Width(head), 60)
	assert.Contains(t, head, "…")
	assert.Contains(t, head, "3", "the count stays")
}

func TestAQuoteInAListNameIsNotEscapedInTheQuestion(t *testing.T) {
	f := newHomeListsFake().withList("q", `Say "hi" ação`, 2)
	m := homeListsScreen(t, f, 120, 40)
	hlRowIs(t, m, "Say")
	hlSend(m, press('x', "x"))
	view := hlFlat(screen(m))
	assert.Contains(t, view, `Delete the list "Say "hi" ação"?`)
	assert.NotContains(t, view, `\"`)
}

func TestTheDeleteQuestionKeepsItsAnswerKeysOnASmallScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}} {
		m := homeListsScreen(t, hlThreeLists(), size[0], size[1])
		hlRowIs(t, m, "Weekend")
		hlSend(m, press('x', "x"))
		view := hlFlat(screen(m))
		assert.Contains(t, view, "n to keep", "%v: the question is not cut before the answer keys", size)
		assert.Contains(t, view, "y delete", "%v: the footer names the keys", size)
	}
}
