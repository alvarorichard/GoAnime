package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests added after hand-applied mutation testing found a behaviour nothing
// pinned down.

func TestThePlainPanelIsSeparatedFromTheRowsByABlankLine(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Weekend")
	assert.Contains(t, hlBody(m), "\n\nLIST\n")
}

func TestTheWidePanelStaysBesideTheListWithinTheTerminalWidth(t *testing.T) {
	for _, w := range []int{100, 120} {
		m := homeListsScreen(t, hlThreeLists(), w, 40)
		hlRowIs(t, m, "Weekend")
		panelAt := -1
		for _, line := range strings.Split(hlBody(m), "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), w, "width %d: %q", w, line)
			if i := strings.Index(line, "Weekend marathon 0"); i >= 0 {
				panelAt = lipgloss.Width(line[:i])
			}
		}
		assert.GreaterOrEqual(t, panelAt, resultsListWidth(w), "width %d: the panel starts after the list column", w)
	}
}

func TestAListWithTitlesMissingFromItsSummaryStillSaysHowManyMore(t *testing.T) {
	// Count and Titles can disagree when the lists cap the titles they send.
	capped := HomeList{Name: "Big", Count: 40, Titles: []string{"A", "B"}}
	assert.Equal(t, []string{"A", "+39 more"}, homeListPanelContent(&capped, 1).lines)

	noTitles := HomeList{Name: "Big", Count: 5}
	c := homeListPanelContent(&noTitles, 8)
	assert.NotContains(t, c.lines, "Nothing in it yet.", "five titles are not nothing")
	assert.Empty(t, c.lines)

	odd := HomeList{Name: "Odd", Count: 1, Titles: []string{"A", "B", "C"}}
	assert.Equal(t, []string{"A", "+2 more"}, homeListPanelContent(&odd, 1).lines, "the longer of the two wins")

	listed := HomeList{Name: "Odd", Count: 0, Titles: []string{"A"}}
	assert.Equal(t, []string{"A"}, homeListPanelContent(&listed, 8).lines, "titles that exist are listed")
}

func TestAReasonDoesNotFollowTheEntryToItsNextOpening(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	hlSend(m, hlEnter)
	require.Contains(t, screen(m), "Give your list a name.")
	hlSend(m, hlEsc)
	assert.False(t, m.lists.name.Focused(), "a closed entry holds no caret")
	hlSend(m, press('n', "n"))
	require.Equal(t, homeEntryCreate, m.lists.entry)
	assert.NotContains(t, screen(m), "Give your list a name.")
}

func TestEscOnTheListsLevelWithAnEmptyBoxChangesNothing(t *testing.T) {
	f := hlThreeLists()
	m := homeListsScreen(t, f, 100, 40)
	hlRowIs(t, m, "Short")
	reads := len(f.log)
	hlSend(m, hlEsc)
	assert.Equal(t, "Short", hlHighlighted(m), "the highlight stays where it was")
	assert.False(t, m.focusInput)
	assert.Equal(t, reads, len(f.log), "and nothing is read again")
}

func TestTheBoxHasNoCaretWhileTheHighlightIsOnAListAfterGoingBack(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Short")
	require.True(t, m.input.Focused(), "inside a list the box has the focus")
	hlSend(m, hlEsc)
	assert.False(t, m.focusInput)
	assert.False(t, m.input.Focused(), "the highlight and a caret are never both on screen")
}

func TestAFailedMutationReadsNothingAgain(t *testing.T) {
	cases := map[string]func(m *searchHomeModel){
		"delete": func(m *searchHomeModel) { hlRowIs(t, m, "Short"); hlSend(m, press('x', "x")) },
		"clear":  func(m *searchHomeModel) { hlOpen(t, m, "Weekend"); hlGoto(t, m, 0); hlSend(m, press('X', "X")) },
		"remove": func(m *searchHomeModel) { hlOpen(t, m, "Weekend"); hlGoto(t, m, 0) },
	}
	for name, setup := range cases {
		f := hlThreeLists()
		f.fail["Delete"] = errors.New("read-only")
		f.fail["Clear"] = errors.New("read-only")
		f.fail["RemoveWork"] = errors.New("read-only")
		m := homeListsScreen(t, f, 100, 40)
		setup(m)
		reads := len(f.log)
		if name == "remove" {
			hlSend(m, press('x', "x"))
			assert.Equal(t, reads+1, len(f.log), "%s: only the removal itself is called", name)
			continue
		}
		hlSend(m, press('y', "y"))
		assert.Equal(t, reads+1, len(f.log), "%s: only the call itself, no reading again", name)
	}
}

func TestTheAnswerKeysOfEveryListQuestionSurviveANarrowFooter(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlRowIs(t, m, "Short")
	hlSend(m, press('x', "x"))
	for _, key := range []string{"y", "n"} {
		assert.True(t, hintsByKey(m)[key].Essential, "delete question: %s", key)
	}
	hlSend(m, hlEsc)
	hlSend(m, press('r', "r"))
	for _, key := range []string{"enter", "esc"} {
		assert.True(t, hintsByKey(m)[key].Essential, "rename entry: %s", key)
	}
	hlSend(m, hlEsc)
	hlSend(m, press('n', "n"))
	for _, key := range []string{"enter", "esc"} {
		assert.True(t, hintsByKey(m)[key].Essential, "create entry: %s", key)
	}
}

func TestStartingOnTheFavoritesTabAlsoLandsOnTheFirstList(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Lists: hlThreeLists(), ScopeIndex: 4})
	assert.False(t, m.focusInput)
	assert.Equal(t, 1, m.cursor)
	assert.False(t, m.input.Focused())
}

func TestTheTeachingLineIsWrappedToTheWidthAndCountedInTheHeight(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {80, 24}, {100, 24}, {100, 20}, {120, 30}, {100, 16}} {
		f := newHomeListsFake().withFavorites(30)
		m := homeListsScreen(t, f, size[0], size[1])
		view := hlBody(m)
		require.Contains(t, hlFlat(view), "press n to make one", "%v", size)
		for _, line := range strings.Split(view, "\n") {
			if !strings.HasPrefix(line, "in ALL") { // the scope row is the shell's to cut
				assert.LessOrEqual(t, lipgloss.Width(line), size[0], "%v: %q", size, line)
			}
		}
		assert.LessOrEqual(t, strings.Count(view, "\n")+1, m.shellHeight(), "%v: the panel leaves room for the line", size)
	}
}

func TestTheWindowMathIsExact(t *testing.T) {
	cases := []struct {
		total, cursor, capacity, start, end int
	}{
		{10, 0, 4, 0, 4},
		{10, 3, 4, 1, 5},
		{10, 5, 4, 3, 7},
		{10, 9, 4, 6, 10},
		{4, 2, 4, 0, 4},
		{3, 2, 4, 0, 3},
		{10, 4, 1, 4, 5},
	}
	for _, c := range cases {
		w := windowAround(c.total, c.cursor, c.capacity, true)
		assert.Equal(t, c.start, w.start, "%+v", c)
		assert.Equal(t, c.end, w.end, "%+v", c)
	}
}

func TestOneRowAboveOrBelowIsStillCounted(t *testing.T) {
	m := homeListsScreen(t, hlManyLists(6), 100, 40)
	items := []homeRowItem{{title: "a"}, {title: "b"}}
	view := ansi.Strip(m.windowLines("HEAD", items, rowWindow{start: 1, end: 3, arrows: true}, 80))
	assert.Contains(t, view, "↑ 1 more above")
	assert.Contains(t, view, "more below")
	flat := ansi.Strip(m.windowLines("HEAD", items, rowWindow{start: 0, end: len(m.rows), arrows: true}, 80))
	assert.NotContains(t, flat, "more above")
	assert.NotContains(t, flat, "more below")
	hidden := ansi.Strip(m.windowLines("HEAD", items, rowWindow{start: 2, end: 4, arrows: false}, 80))
	assert.NotContains(t, hidden, "more", "no arrows when there was no room for them")
}

func TestNoRowIsMarkedWhileTheBoxHasTheFocus(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	items := []homeRowItem{{title: "a"}, {title: "b"}}
	m.focusInput = false
	assert.Contains(t, ansi.Strip(m.windowLines("HEAD", items, rowWindow{start: 0, end: 2}, 80)), "▶")
	m.focusInput = true
	assert.NotContains(t, ansi.Strip(m.windowLines("HEAD", items, rowWindow{start: 0, end: 2}, 80)), "▶")
}

func TestTheNameEntryFollowsTheTerminalWhenItIsResized(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpenCreate(t, m)
	wide := m.lists.name.Width()
	hlSend(m, tea.WindowSizeMsg{Width: 50, Height: 30})
	assert.Less(t, m.lists.name.Width(), wide)
	assert.Equal(t, 50-lipgloss.Width(m.lists.name.Prompt)-2, m.lists.name.Width())
}

func TestClearAllIsNeverAskedOfAnEmptyList(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlOpen(t, m, "Empty")
	m.listsRowKey("X") // not reachable from the keyboard, where the box has the focus
	assert.Equal(t, clearNothing, m.clearing)
}

func TestTheNameEntryTakesTheCaretFromTheBox(t *testing.T) {
	m := homeListsScreen(t, hlThreeLists(), 100, 40)
	hlFocusBox(m)
	require.True(t, m.input.Focused())
	m.openEntry(homeEntryCreate, HomeList{})
	assert.False(t, m.input.Focused(), "two carets would show at once")
	assert.True(t, m.lists.name.Focused())
	hlSend(m, hlEsc)
	assert.True(t, m.input.Focused())
}

func TestTheTeachingLineIsNotShownBesideAReadError(t *testing.T) {
	f := newHomeListsFake().withFavorites(1)
	m := homeListsScreen(t, f, 100, 40)
	require.Contains(t, screen(m), "You have none of your own yet")
	f.fail["Lists"] = errors.New("flaky")
	hlSend(m, hlTab)
	tabTo(t, m, "FAVORITES")
	view := screen(m)
	assert.Contains(t, view, "Could not read your lists. Flaky.")
	assert.NotContains(t, view, "You have none of your own yet", "the lists are in doubt, so say nothing about having none")
}
