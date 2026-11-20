package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests from the beginner review of the "Add to a list" screen: one word for
// the things in a list ("titles"), how to finish said on the screen, plain
// marks in words, a letter counter, the 50-list limit said up front, and wide
// characters that never overflow the terminal.
//
// Like favorite_picker_test.go these swap package globals and are not parallel.

func TestTheRowsSayTitlesLikeTheResultsDo(t *testing.T) {
	view := pkScreen(pkOpen(t, pkNewLists()))
	assert.Contains(t, view, "12 titles")
	assert.Contains(t, view, "1 title")
	assert.NotContains(t, view, "work")
}

func TestBeforeAnythingIsDoneTheScreenSaysHowToFinish(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	view := pkScreen(m)
	assert.Contains(t, view, "enter adds it or takes it off. esc when you are done.")
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	view = pkScreen(m)
	assert.NotContains(t, view, "esc when you are done.", "a sentence about what happened takes its place")
	assert.Contains(t, view, `Added "Frieren" to Favorites.`)
	footer := strings.Split(view, "\n")
	assert.Contains(t, footer[len(footer)-1], "esc")
	assert.Contains(t, footer[len(footer)-1], "when you are done", "and the footer keeps saying it")
}

func TestPlainModeSaysNotInThisListInWords(t *testing.T) {
	pkPlain(t)
	m := pkOpen(t, pkNewLists())
	view := pkScreen(m)
	assert.Contains(t, view, "not in this list")
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Create a new list") {
			assert.NotContains(t, line, "in this list", "the create row has no mark")
		}
	}
}

func TestRichModeSaysNothingForListsTheWorkIsNotIn(t *testing.T) {
	view := pkScreen(pkOpen(t, pkNewLists()))
	assert.NotContains(t, view, "not in this list")
}

func TestPlainNarrowRowsDropTheWordsButKeepTheBrackets(t *testing.T) {
	pkPlain(t)
	theme := CurrentTheme()
	out := ansi.Strip(renderPickerRow(&theme, pickerRow{name: "Weekend marathon", count: "12 titles"}, false, 36))
	assert.Contains(t, out, "[ ]")
	assert.NotContains(t, out, "not in this list")
	assert.LessOrEqual(t, ansi.StringWidth(out), 36)
}

func TestTheNameEntryCountsLetters(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	assert.Contains(t, pkScreen(m), "0/40")
	pkType(m, "Weekend")
	view := pkScreen(m)
	assert.Contains(t, view, "7/40")
	assert.NotContains(t, view, "the most a name can have")
	pkType(m, strings.Repeat("x", 33))
	assert.Contains(t, pkScreen(m), "40/40, the most a name can have")
	pkType(m, "more")
	assert.Len(t, []rune(m.picker.input.Value()), 40, "typing stops at 40")
	assert.Contains(t, pkScreen(m), "40/40, the most a name can have")
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Contains(t, pkScreen(m), "39/40")
	assert.NotContains(t, pkScreen(m), "the most a name can have")
}

func TestTheCounterCountsLettersNotBytes(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	pkType(m, "日本語é")
	assert.Contains(t, pkScreen(m), "4/40")
}

func TestAnErrorTakesTheCountersLine(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEnter)
	view := pkScreen(m)
	assert.Contains(t, view, "Give the list a name.")
	assert.NotContains(t, view, "0/40")
}

func TestAtFiftyListsTheScreenSaysSoBeforeAnyNameIsTyped(t *testing.T) {
	fake := pkBigLists(50)
	m := pkOpen(t, fake)
	view := pkScreen(m)
	assert.Contains(t, view, "You have 50 lists, the most there can be. Delete one first.")
	assert.Contains(t, pkHighlighted(t, view), "Create a new list (you have 50, the most there can be)")
	assert.Equal(t, "see why", m.hints()[0].Desc)

	_, cmd := m.Update(pkEnter)

	assert.Nil(t, cmd)
	assert.False(t, m.picker.naming, "no name entry that could only fail")
	assert.Zero(t, fake.createCalls)
	assert.Regexp(t, `! +You have 50 lists, the most there can be\. Delete one first\.`, pkScreen(m))
}

func TestFortyNineListsStillAllowOneMore(t *testing.T) {
	fake := pkBigLists(49)
	m := pkOpen(t, fake)
	assert.NotContains(t, pkScreen(m), "the most there can be")
	assert.Equal(t, "new list", m.hints()[0].Desc)
	_, _ = m.Update(pkEnter)
	assert.True(t, m.picker.naming)
}

func TestFavoritesIsNotOneOfTheFiftyLists(t *testing.T) {
	m := pkOpen(t, pkBigLists(49))
	assert.Equal(t, 49, m.picker.userLists())
	assert.False(t, m.picker.atLimit())
}

func TestWideCharactersNeverOverflowTheTerminalWidth(t *testing.T) {
	wide := strings.Repeat("日", 40) // 40 letters, 80 cells
	for _, size := range [][2]int{{80, 24}, {60, 20}, {40, 12}} {
		fake := pkNewLists()
		fake.lists = append(fake.lists, HomeList{ID: "wide", Name: wide})
		m, _ := pkResults(t, fake)
		_, _ = m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		_, _ = m.Update(pkCtrlL)
		_, _ = m.Update(pkEnd)
		_, _ = m.Update(pkEnter)
		check := func(what string) {
			view := pkScreen(m)
			for _, line := range strings.Split(view, "\n") {
				assert.LessOrEqual(t, ansi.StringWidth(line), size[0], "%v %s: %q", size, what, line)
			}
			assert.LessOrEqual(t, strings.Count(view, "\n")+1, size[1], "%v %s", size, what)
		}
		check("row and status")
		_, _ = m.Update(pkHome)
		_, _ = m.Update(pkEnter)
		pkType(m, wide)
		check("name entry")
		assert.Contains(t, pkScreen(m), "40/40, the most a name can have")
		_, _ = m.Update(pkEnter)
		check("created")
	}
}

func TestAWideTitleIsShortenedByCellsInTheHeading(t *testing.T) {
	theme := CurrentTheme()
	p := newPickerState(&theme, pkNewLists(), HomeFavorite{Key: "k"}, strings.Repeat("日", 60), map[string]bool{})
	p.fit(50, 20)
	for _, line := range strings.Split(p.view(50, 20), "\n") {
		assert.LessOrEqual(t, ansi.StringWidth(line), 50, line)
	}
}

func TestEveryPackageLevelNameOfThePickerIsPrefixed(t *testing.T) {
	// The home tab declares its own helpers in the same package, so the
	// picker's package-level names all start with "picker" (or are the seam or
	// the row renderer the brief names).
	file, err := parser.ParseFile(token.NewFileSet(), "favorite_picker.go", nil, 0)
	require.NoError(t, err)
	prefixed := func(name string) bool {
		return strings.HasPrefix(name, "picker") || strings.HasPrefix(name, "newPicker") ||
			name == "renderPickerRow" || name == "newFavoriteLists"
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				assert.True(t, prefixed(d.Name.Name), "func %s is not prefixed", d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch sp := spec.(type) {
				case *ast.TypeSpec:
					assert.True(t, prefixed(sp.Name.Name), "type %s is not prefixed", sp.Name.Name)
				case *ast.ValueSpec:
					for _, id := range sp.Names {
						assert.True(t, prefixed(id.Name), "%s is not prefixed", id.Name)
					}
				}
			}
		}
	}
}
