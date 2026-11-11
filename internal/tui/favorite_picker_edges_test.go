package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Edge cases for the "Add to a list" screen, written from a hand-run mutation
// pass over favorite_picker.go and its hooks in anime_results.go: each test
// here fails on a realistic mutant that the first tests let through.
//
// Like favorite_picker_test.go these swap package globals and are not parallel.

func TestTheSentenceOfAnEarlierActionIsGoneAfterNamingAList(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	require.Contains(t, pkScreen(m), "to Favorites.")
	_, _ = m.Update(pkHome)
	_, _ = m.Update(pkEnter)

	_, _ = m.Update(pkEsc)

	assert.NotContains(t, pkScreen(m), "to Favorites.", "opening the name entry starts a fresh status")
}

func TestAFailedChangeIsMarkedWithTheNoticeBlock(t *testing.T) {
	fake := pkNewLists()
	fake.toggleErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	assert.Regexp(t, `! +Could not change`, pkScreen(m))

	pkPlain(t)
	_, _ = m.Update(pkEnter)
	assert.Contains(t, pkScreen(m), "[!] Could not change", "in plain mode the block is in brackets")
}

func TestAFailedNameIsMarkedWithTheNoticeBlock(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEnter)
	assert.Regexp(t, `! +Give the list a name\.`, pkScreen(m))
}

func TestThePickerSentencesKeepTheMarginOfTheResultList(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	lines := strings.Split(pkScreen(m), "\n")
	assert.True(t, strings.HasPrefix(lines[2], ` Add "Frieren"`), lines[2])
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	assert.Contains(t, pkScreen(m), "\n Added \"Frieren\"")
	_, _ = m.Update(pkHome)
	_, _ = m.Update(pkEnter)
	view := pkScreen(m)
	assert.Contains(t, view, "\n Name your list")
	assert.Contains(t, view, "\n enter create")
}

func TestANoticeAboveTheResultsTakesRowsFromThePicker(t *testing.T) {
	NoticeOnNextResults("Could not open that title, so you are back at the results.")
	fake := pkBigLists(40)
	m, _ := pkResults(t, fake)
	require.NotEmpty(t, m.notice)
	_, _ = m.Update(pkCtrlL)

	for range 42 {
		view := pkScreen(m)
		assert.LessOrEqual(t, strings.Count(view, "\n")+1, 24, "the notice and the picker share the body")
		assert.Contains(t, view, "Could not open that title")
		assert.Equal(t, 1, strings.Count(view, "▶"))
		_, _ = m.Update(keyDown)
	}
	_, _ = m.Update(pkEnd)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "List 40")
}

func TestStarringInPlainModeRefreshesThePanelUnderTheList(t *testing.T) {
	pkPlain(t)
	m, _ := pkResults(t, pkNewLists())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	require.NotEmpty(t, m.plainPanel)
	require.NotContains(t, ansi.Strip(m.plainPanel), "* Frieren")
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(keyDown)

	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEsc)

	assert.Contains(t, ansi.Strip(m.plainPanel), "* Frieren")
	assert.Contains(t, ansi.Strip(m.plainPanel), "ctrl+f takes it off FAVORITES")
}

func TestThePickerNeedsNoFavoritesSetToWork(t *testing.T) {
	theme := CurrentTheme()
	fake := pkNewLists()
	p := newFavoritePicker(&theme, fake, HomeFavorite{Key: "frieren", Title: "Frieren"}, "Frieren", nil)
	p.fit(80, 20)
	p.move(1)
	assert.NotPanics(t, func() { p.choose() })
	assert.True(t, fake.has(FavoritesListID, "frieren"))
}

func TestTheNameEntryKeepsItsBlankLinesOnlyWhenThereIsRoom(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	roomy := strings.Split(m.picker.view(60, 12), "\n")
	assert.Len(t, roomy, 7, "heading, hint, blank, entry, error line, blank, instructions")
	assert.Equal(t, "", strings.TrimSpace(ansi.Strip(roomy[2])))
	assert.Contains(t, ansi.Strip(roomy[3]), "Weekend marathon")
	assert.Contains(t, ansi.Strip(roomy[6]), "enter create")

	tight := strings.Split(m.picker.view(60, 5), "\n")
	assert.Len(t, tight, 5, "no blank lines when the body is short")
	assert.Contains(t, ansi.Strip(tight[4]), "enter create · esc back", "the instructions are still there")
}

func TestTheNameErrorSitsRightUnderTheEntry(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEnter)
	lines := strings.Split(m.picker.view(60, 12), "\n")
	assert.Contains(t, ansi.Strip(lines[4]), "Give the list a name.")
}

func TestTheLoadErrorFitsEverySize(t *testing.T) {
	fake := pkNewLists()
	fake.listsErr = errors.New("the file is unreadable and this sentence is long enough to wrap around")
	theme := CurrentTheme()
	for height := 1; height <= 24; height++ {
		for _, width := range []int{24, 40, 80} {
			p := newFavoritePicker(&theme, fake, HomeFavorite{Key: "frieren"}, "Frieren", map[string]bool{})
			p.fit(width, height)
			view := p.view(width, height)
			for _, line := range strings.Split(view, "\n") {
				require.LessOrEqual(t, ansi.StringWidth(line), width, "%dx%d %q", width, height, line)
			}
			if height >= 9 {
				require.LessOrEqual(t, strings.Count(view, "\n")+1, height, "%dx%d\n%s", width, height, view)
				require.Contains(t, strings.Join(strings.Fields(ansi.Strip(view)), " "), "Could not read your lists.")
			}
		}
	}
}

func TestOpeningBeforeAnyResizeStillLaysThePickerOut(t *testing.T) {
	fake := pkNewLists()
	prev := newFavoriteLists
	newFavoriteLists = func() FavoriteLists { return fake }
	t.Cleanup(func() { newFavoriteLists = prev })
	m := newAnimeResultsModel(pkSampleResults())
	_, _ = m.Update(pkCtrlL)
	require.NotNil(t, m.picker)
	assert.Positive(t, m.picker.room)
}

func TestLettersAreLettersInTheNameEntry(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(press('q', "q"))
	assert.Equal(t, "q", m.picker.input.Value(), "q is a letter in the name entry")
	assert.True(t, m.picker.naming)
	assert.NotNil(t, m.picker)

	_, _ = m.Update(press('j', "j"))
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnd)
	assert.Equal(t, "qj", m.picker.input.Value())
	assert.Equal(t, 0, m.picker.cursor)
}

func TestTheListRowsAreAlignedWhetherOrNotTheWorkIsIn(t *testing.T) {
	theme := CurrentTheme()
	in := ansi.Strip(renderPickerRow(&theme, pickerRow{name: "Weekend", count: "1 work", in: true}, false, 60))
	out := ansi.Strip(renderPickerRow(&theme, pickerRow{name: "Weekend", count: "1 work"}, false, 60))
	assert.Equal(t, strings.Index(in, "Weekend"), strings.Index(out, "Weekend"))
	assert.Equal(t, 60, ansi.StringWidth(in), "rows fill the width")
	assert.Equal(t, 60, ansi.StringWidth(out), "rows fill the width")
}
