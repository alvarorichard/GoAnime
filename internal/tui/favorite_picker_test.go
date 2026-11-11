package tui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The "Add to a list" screen of the results: ctrl+l opens an overlay with the
// saved lists, enter puts the highlighted work in a list or takes it off, and
// "+ Create a new list" names a new one.
//
// These tests swap package globals (newFavoriteLists, the plain mode) and so
// are not parallel. They use fakes only: no file and no HOME is touched.

var (
	pkCtrlL = tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl}
	pkCtrlC = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	pkEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	pkEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	pkUp    = tea.KeyPressMsg{Code: tea.KeyUp}
	pkPgUp  = tea.KeyPressMsg{Code: tea.KeyPgUp}
	pkPgDn  = tea.KeyPressMsg{Code: tea.KeyPgDown}
	pkHome  = tea.KeyPressMsg{Code: tea.KeyHome}
	pkEnd   = tea.KeyPressMsg{Code: tea.KeyEnd}
)

// pickerLists is a FavoriteLists that behaves like the real store (a work is
// toggled in or out of a list, names are checked) and counts what it is asked.
type pickerLists struct {
	lists []HomeList
	works map[string][]string // list ID -> keys of its works

	listsErr, createErr, toggleErr error

	listCalls, memberCalls, createCalls, toggleCalls, otherCalls int
	// calls records every Create and ToggleWork in order: "create:NAME" and
	// "toggle:ID:KEY".
	calls  []string
	nextID int
}

func pkNewLists() *pickerLists {
	return &pickerLists{
		lists: []HomeList{
			{ID: FavoritesListID, Name: "Favorites", Auto: true},
			{ID: "weekend", Name: "Weekend"},
			{ID: "kids", Name: "Kids"},
		},
		works: map[string][]string{
			FavoritesListID: {"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"},
			"weekend":       {"x"},
		},
	}
}

func (f *pickerLists) Lists() ([]HomeList, error) {
	f.listCalls++
	if f.listsErr != nil {
		return nil, f.listsErr
	}
	out := make([]HomeList, len(f.lists))
	for i, l := range f.lists {
		l.Count = len(f.works[l.ID])
		out[i] = l
	}
	return out, nil
}

func (f *pickerLists) Membership(key string) map[string]bool {
	f.memberCalls++
	out := map[string]bool{}
	for id, keys := range f.works {
		for _, k := range keys {
			if k == key {
				out[id] = true
			}
		}
	}
	return out
}

func (f *pickerLists) Create(name string) (HomeList, error) {
	f.createCalls++
	f.calls = append(f.calls, "create:"+name)
	if f.createErr != nil {
		return HomeList{}, f.createErr
	}
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return HomeList{}, favorites.ErrNameEmpty
	}
	for _, l := range f.lists {
		if strings.EqualFold(l.Name, name) {
			return HomeList{}, errors.New("you already have a list with that name")
		}
	}
	f.nextID++
	list := HomeList{ID: fmt.Sprintf("new%d", f.nextID), Name: name}
	f.lists = append(f.lists, list)
	return list, nil
}

func (f *pickerLists) ToggleWork(id string, w HomeFavorite) (bool, error) {
	f.toggleCalls++
	f.calls = append(f.calls, "toggle:"+id+":"+w.Key)
	if f.toggleErr != nil {
		return false, f.toggleErr
	}
	for i, k := range f.works[id] {
		if k == w.Key {
			f.works[id] = append(f.works[id][:i:i], f.works[id][i+1:]...)
			return false, nil
		}
	}
	if f.works == nil {
		f.works = map[string][]string{}
	}
	f.works[id] = append(f.works[id], w.Key)
	return true, nil
}

func (f *pickerLists) Works(string) ([]HomeFavorite, error) { f.otherCalls++; return nil, nil }
func (f *pickerLists) Rename(string, string) error          { f.otherCalls++; return nil }
func (f *pickerLists) Delete(string) error                  { f.otherCalls++; return nil }
func (f *pickerLists) RemoveWork(string, string) error      { f.otherCalls++; return nil }
func (f *pickerLists) Clear(string) error                   { f.otherCalls++; return nil }

// has reports whether the fake holds the work key in the list.
func (f *pickerLists) has(id, key string) bool {
	for _, k := range f.works[id] {
		if k == key {
			return true
		}
	}
	return false
}

// reads is every call that reads the saved lists.
func (f *pickerLists) reads() int { return f.listCalls + f.memberCalls + f.otherCalls }

func pkSampleResults() []*models.Anime {
	return []*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[PT-BR] Naruto", Source: "Animefire.io", Year: "2002", MediaType: models.MediaTypeAnime},
	}
}

// pkResults is the results screen with fake as its saved lists and a
// temporary FAVORITES store behind ctrl+f, measured 80 by 24.
func pkResults(t *testing.T, fake FavoriteLists) (*animeResultsModel, *favorites.Store) {
	t.Helper()
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	prevFavs, prevLists := newFavorites, newFavoriteLists
	newFavorites = func() favoritesAPI { return store }
	newFavoriteLists = func() FavoriteLists { return fake }
	t.Cleanup(func() { newFavorites, newFavoriteLists = prevFavs, prevLists })
	m := newAnimeResultsModel(pkSampleResults())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m, store
}

func pkPlain(t *testing.T) {
	t.Helper()
	plainMu.Lock()
	prev := plainOverride
	plainMu.Unlock()
	SetPlain(true)
	t.Cleanup(func() {
		plainMu.Lock()
		plainOverride = prev
		plainMu.Unlock()
	})
}

func pkScreen(m *animeResultsModel) string { return ansi.Strip(m.View().Content) }

func pkType(m *animeResultsModel, text string) {
	for _, r := range text {
		_, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

// pkOpen is the results screen with the picker open on Frieren.
func pkOpen(t *testing.T, fake *pickerLists) *animeResultsModel {
	t.Helper()
	m, _ := pkResults(t, fake)
	_, _ = m.Update(pkCtrlL)
	require.NotNil(t, m.picker, "ctrl+l opens the picker")
	return m
}

// highlighted is the picker row that carries the cursor, on screen.
func pkHighlighted(t *testing.T, view string) string {
	t.Helper()
	var found []string
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "▶") || strings.Contains(line, " > ") && !strings.Contains(line, "GOANIME") {
			found = append(found, line)
		}
	}
	require.Len(t, found, 1, "exactly one row carries the cursor:\n%s", view)
	return found[0]
}

// ---- the feature is absent without lists -------------------------------------------

func TestWithoutListsTheFeatureIsAbsent(t *testing.T) {
	m, _ := pkResults(t, nil)
	require.Nil(t, m.lists)
	before := pkScreen(m)

	_, cmd := m.Update(pkCtrlL)

	assert.Nil(t, cmd)
	assert.Nil(t, m.picker)
	assert.Equal(t, before, pkScreen(m), "ctrl+l does nothing")
	for _, h := range m.hints() {
		assert.NotEqual(t, "ctrl+l", h.Key)
	}
	assert.NotContains(t, m.detailsContent().tip, "ctrl+l")
	_, _ = m.Update(press('?', "?"))
	assert.NotContains(t, pkScreen(m), "ctrl+l")
}

func TestWithoutListsTheExistingScreenIsExactlyWhatItWas(t *testing.T) {
	m, _ := pkResults(t, nil)

	assert.Equal(t, []Hint{
		{Key: "enter", Desc: "open", Essential: true},
		{Key: "←→", Desc: "version"},
		{Key: "ctrl+f", Desc: "favorite"},
		{Key: "↑↓", Desc: "move"},
		{Key: "esc", Desc: "edit search", Essential: true},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}, m.hints())
	assert.Equal(t, "ctrl+f adds it to FAVORITES", m.detailsContent().tip)

	_, _ = m.Update(ctrlF)
	assert.Equal(t, "ctrl+f takes it off FAVORITES", m.detailsContent().tip)
	assert.Equal(t, "unfavorite", m.hints()[2].Desc)

	m.help()
	view := pkScreen(m)
	assert.Contains(t, view, "choose a version")
	assert.Contains(t, view, "add to or remove from FAVORITES")
	assert.Contains(t, view, "edit the search")
	assert.NotContains(t, view, "put it in a list")
}

func TestWithoutListsOrFavoritesThereIsNoTip(t *testing.T) {
	m, _ := pkResults(t, nil)
	m.favs = nil
	assert.Empty(t, m.detailsContent().tip)
}

// help opens the KEYS overlay.
func (m *animeResultsModel) help() { _, _ = m.Update(press('?', "?")) }

func TestWithoutListsCtrlLIsNotATypedLetter(t *testing.T) {
	m, _ := pkResults(t, nil)
	_, _ = m.Update(pkCtrlL)
	assert.Equal(t, 0, int(m.results.FilterState()), "ctrl+l did not start the filter")
}

// ---- opening and closing -------------------------------------------------------------

func TestCtrlLOpensThePickerOverTheList(t *testing.T) {
	fake := pkNewLists()
	m, _ := pkResults(t, fake)
	closed := pkScreen(m)

	_, cmd := m.Update(pkCtrlL)

	assert.Nil(t, cmd)
	view := pkScreen(m)
	assert.Contains(t, view, `Add "Frieren" to a list`)
	assert.Contains(t, view, "+ Create a new list")
	assert.Contains(t, view, "★ Favorites")
	assert.Contains(t, view, "Weekend")
	assert.Contains(t, view, "Kids")
	assert.NotContains(t, view, "type to filter", "the picker replaces the list")
	assert.NotEqual(t, closed, view)
}

func TestEscClosesThePickerAndTheListIsBackAsItWas(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	closed := pkScreen(m)
	_, _ = m.Update(pkCtrlL)

	_, cmd := m.Update(pkEsc)

	assert.Nil(t, cmd, "esc closes the picker, not the screen")
	assert.Nil(t, m.picker)
	assert.NoError(t, m.err)
	assert.Equal(t, closed, pkScreen(m))
}

func TestQClosesThePickerWithoutQuitting(t *testing.T) {
	m := pkOpen(t, pkNewLists())

	_, cmd := m.Update(press('q', "q"))

	assert.Nil(t, cmd)
	assert.Nil(t, m.picker)
	assert.NoError(t, m.err)
}

func TestCtrlCInThePickerQuitsLikeTheResults(t *testing.T) {
	m := pkOpen(t, pkNewLists())

	_, cmd := m.Update(pkCtrlC)

	require.NotNil(t, cmd)
	assert.Equal(t, ErrSelectionCancelled, m.err)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestCtrlCInTheNameEntryQuitsToo(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	require.True(t, m.picker.naming)

	_, cmd := m.Update(pkCtrlC)

	require.NotNil(t, cmd)
	assert.Equal(t, ErrSelectionCancelled, m.err)
}

func TestCtrlLWhileFilteringIsIgnored(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	_, _ = m.Update(press('n', "n"))
	require.Equal(t, 1, int(m.results.FilterState()))

	_, _ = m.Update(pkCtrlL)

	assert.Nil(t, m.picker)
}

func TestCtrlLWithNothingHighlightedIsIgnored(t *testing.T) {
	fake := pkNewLists()
	m, _ := pkResults(t, fake)
	m.results.SetItems(nil)

	_, _ = m.Update(pkCtrlL)

	assert.Nil(t, m.picker)
	assert.Zero(t, fake.reads())
}

func TestCtrlLIsIgnoredInTheHelpOverlay(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	m.help()
	_, _ = m.Update(pkCtrlL)
	assert.Nil(t, m.picker)
	assert.True(t, m.showHelp)
}

func TestOtherKeysAreNotTypedIntoTheResultsBehindThePicker(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(press('n', "n"))
	_, _ = m.Update(ctrlF)
	assert.Equal(t, 0, int(m.results.FilterState()))
	assert.NotNil(t, m.picker)
	assert.False(t, m.favKeys["frieren"], "ctrl+f is not heard behind the picker")
}

func TestThePickerOpensOnTheCreateRow(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "+ Create a new list")
}

// ---- rows ----------------------------------------------------------------------------

func TestRowsAreTheListsInOrderWithTheirCounts(t *testing.T) {
	fake := pkNewLists()
	fake.works["kids"] = nil
	m := pkOpen(t, fake)
	view := pkScreen(m)

	create := strings.Index(view, "+ Create a new list")
	favs := strings.Index(view, "Favorites")
	weekend := strings.Index(view, "Weekend")
	kids := strings.Index(view, "Kids")
	require.True(t, create >= 0 && create < favs && favs < weekend && weekend < kids, view)
	rows := strings.Split(view, "\n")
	find := func(name string) string {
		for _, r := range rows {
			if strings.Contains(r, name) {
				return r
			}
		}
		t.Fatalf("no row %q in\n%s", name, view)
		return ""
	}
	assert.Contains(t, find("Favorites"), "12 titles")
	assert.Contains(t, find("Weekend"), "1 title")
	assert.NotContains(t, find("Weekend"), "1 titles")
	assert.Contains(t, find("Kids"), "empty")
}

func TestOnlyTheListsHoldingTheWorkSayIt(t *testing.T) {
	fake := pkNewLists()
	fake.works["weekend"] = []string{"frieren"}
	m := pkOpen(t, fake)
	view := pkScreen(m)

	for _, row := range strings.Split(view, "\n") {
		switch {
		case strings.Contains(row, "Weekend"):
			assert.Contains(t, row, "✓ in this list")
		case strings.Contains(row, "Favorites"), strings.Contains(row, "Kids"), strings.Contains(row, "Create"):
			assert.NotContains(t, row, "in this list")
		}
	}
}

func TestPlainModeMarksRowsWithBracketsInText(t *testing.T) {
	pkPlain(t)
	fake := pkNewLists()
	fake.works["weekend"] = []string{"frieren"}
	m := pkOpen(t, fake)
	view := pkScreen(m)

	for _, row := range strings.Split(view, "\n") {
		switch {
		case strings.Contains(row, "Weekend"):
			assert.Regexp(t, `\[x\] +Weekend`, row)
			assert.Contains(t, row, "in this list")
			assert.NotContains(t, row, "not in this list")
		case strings.Contains(row, "Kids"):
			assert.Regexp(t, `\[ \] +Kids`, row)
			assert.Contains(t, row, "not in this list", "in words, not only brackets")
		case strings.Contains(row, "Favorites"):
			assert.Regexp(t, `\[ \] \* Favorites`, row)
		}
	}
	for _, glyph := range []string{"▶", "★", "✓", "↑", "↓", "·", "❯"} {
		assert.NotContains(t, view, glyph, "plain mode draws ASCII only")
	}
	assert.Contains(t, pkHighlighted(t, view), "+ Create a new list")
	assert.Contains(t, view, "> ", "the cursor is a >")
}

func TestPlainModeCursorFollowsTheKeys(t *testing.T) {
	pkPlain(t)
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(keyDown)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Favorites")
}

// ---- adding and removing -------------------------------------------------------------

func TestEnterOnAListAddsTheWork(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(keyDown) // Favorites
	_, _ = m.Update(keyDown) // Weekend

	_, cmd := m.Update(pkEnter)

	assert.Nil(t, cmd)
	assert.Equal(t, []string{"toggle:weekend:frieren"}, fake.calls)
	assert.True(t, fake.has("weekend", "frieren"))
	view := pkScreen(m)
	assert.Contains(t, view, `Added "Frieren" to Weekend.`)
	assert.NotNil(t, m.picker, "the picker stays open")
	assert.Contains(t, pkHighlighted(t, view), "2 titles")
	assert.Contains(t, pkHighlighted(t, view), "✓ in this list")
}

func TestTheWorkIsSentWithItsDetails(t *testing.T) {
	var got HomeFavorite
	fake := &pickerRecordingLists{pickerLists: pkNewLists(), onToggle: func(w HomeFavorite) { got = w }}
	m, _ := pkResults(t, fake)
	_, _ = m.Update(keyRight) // the English copy
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)

	assert.Equal(t, HomeFavorite{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"}, got)
}

type pickerRecordingLists struct {
	*pickerLists
	onToggle func(HomeFavorite)
}

func (r *pickerRecordingLists) ToggleWork(id string, w HomeFavorite) (bool, error) {
	r.onToggle(w)
	return r.pickerLists.ToggleWork(id, w)
}

func TestEnterAgainTakesTheWorkOff(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)

	_, _ = m.Update(pkEnter)

	assert.False(t, fake.has("weekend", "frieren"))
	view := pkScreen(m)
	assert.Contains(t, view, `Removed "Frieren" from Weekend.`)
	assert.NotContains(t, view, `Added "Frieren"`)
	assert.Contains(t, pkHighlighted(t, view), "1 title")
	assert.NotContains(t, pkHighlighted(t, view), "in this list")
}

func TestSeveralListsCanBeChosenInARow(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)

	assert.True(t, fake.has("weekend", "frieren"))
	assert.True(t, fake.has("kids", "frieren"))
	require.NotNil(t, m.picker)
	view := pkScreen(m)
	assert.Contains(t, view, `Added "Frieren" to Kids.`)
	assert.Equal(t, 2, strings.Count(view, "in this list"))
}

func TestAWorkAlreadyInAListIsMarkedWhenThePickerOpens(t *testing.T) {
	fake := pkNewLists()
	fake.works["kids"] = []string{"frieren"}
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnd)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "in this list")

	_, _ = m.Update(pkEnter)
	assert.Contains(t, pkScreen(m), `Removed "Frieren" from Kids.`)
}

func TestTheStatusFollowsTheLatestAction(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	require.Contains(t, pkScreen(m), "to Favorites.")
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	view := pkScreen(m)
	assert.Contains(t, view, `Added "Frieren" to Weekend.`)
	assert.NotContains(t, view, "to Favorites.")
}

func TestAListThatCannotBeChangedSaysWhy(t *testing.T) {
	fake := pkNewLists()
	fake.toggleErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)

	_, _ = m.Update(pkEnter)

	view := pkScreen(m)
	assert.Contains(t, view, `Could not change "Weekend". The disk is full.`)
	assert.NotContains(t, view, "Added")
	require.NotNil(t, m.picker, "the picker stays open")
	assert.Contains(t, pkHighlighted(t, view), "1 title", "nothing changed in the count")
	assert.NotContains(t, pkHighlighted(t, view), "in this list")
	assert.False(t, m.favKeys["frieren"])
}

func TestEnterMoveKeysClampAtTheEnds(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkUp)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Create a new list")
	_, _ = m.Update(pkEnd)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Kids")
	_, _ = m.Update(keyDown)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Kids")
	_, _ = m.Update(pkUp)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Weekend")
	_, _ = m.Update(pkHome)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Create a new list")
	_, _ = m.Update(pkPgDn)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Kids", "page down goes as far as there is")
	_, _ = m.Update(pkPgUp)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Create a new list")
}

func TestEnterWordTellsWhatEnterWillDo(t *testing.T) {
	fake := pkNewLists()
	fake.works["weekend"] = []string{"frieren"}
	m := pkOpen(t, fake)
	desc := func() string { return m.hints()[0].Desc }

	assert.Equal(t, "new list", desc())
	_, _ = m.Update(keyDown)
	assert.Equal(t, "add", desc())
	_, _ = m.Update(keyDown)
	assert.Equal(t, "take off", desc())
	assert.Equal(t, "enter", m.hints()[0].Key)
}

func TestPickerFooterNamesEnterMoveAndClose(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	hints := m.hints()
	var keys []string
	for _, h := range hints {
		keys = append(keys, h.Key)
	}
	assert.Equal(t, []string{"enter", "↑↓", "esc"}, keys)
	assert.True(t, hints[0].Essential)
	assert.True(t, hints[2].Essential)
	assert.Equal(t, "when you are done", hints[2].Desc)
	footer := strings.Split(pkScreen(m), "\n")
	assert.Contains(t, footer[len(footer)-1], "esc")
	assert.Contains(t, footer[len(footer)-1], "when you are done")
}

func TestPlainPickerFooterSaysUpDn(t *testing.T) {
	pkPlain(t)
	m := pkOpen(t, pkNewLists())
	assert.Equal(t, "up/dn", m.hints()[1].Key)
}

// ---- creating a list -----------------------------------------------------------------

func TestCreateRowOpensTheNameEntry(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)

	_, _ = m.Update(pkEnter)

	view := pkScreen(m)
	assert.Contains(t, view, "Name your list")
	assert.Contains(t, view, "e.g. Weekend marathon")
	assert.Contains(t, view, "enter create · esc back")
	assert.Contains(t, view, `"Frieren"`, "it says what the list starts with")
	assert.Zero(t, fake.createCalls, "nothing is made until enter on a name")
	assert.Equal(t, []Hint{
		{Key: "enter", Desc: "create", Essential: true},
		{Key: "esc", Desc: "back", Essential: true},
	}, m.hints())
}

func TestNamingAListCreatesItAndPutsTheWorkIn(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Weekend marathon")
	assert.Contains(t, pkScreen(m), "Weekend marathon")

	_, _ = m.Update(pkEnter)

	assert.Equal(t, []string{"create:Weekend marathon", "toggle:new1:frieren"}, fake.calls)
	require.False(t, m.picker.naming, "back to the lists")
	view := pkScreen(m)
	assert.Contains(t, view, `Created "Weekend marathon" and added "Frieren".`)
	row := pkHighlighted(t, view)
	assert.Contains(t, row, "Weekend marathon", "the new list is highlighted")
	assert.Contains(t, row, "1 title")
	assert.Contains(t, row, "✓ in this list")
	assert.NotContains(t, view, "Name your list")
}

func TestTheNameIsLimitedToFortyLetters(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	pkType(m, strings.Repeat("é", 50))
	assert.Equal(t, 40, len([]rune(m.picker.input.Value())))
	assert.Equal(t, 40, m.picker.input.CharLimit)
	assert.Equal(t, strings.Repeat("é", 40), m.picker.input.Value())
}

func TestEscInTheNameEntryGoesBackWithoutCreating(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Nope")

	_, cmd := m.Update(pkEsc)

	assert.Nil(t, cmd)
	require.NotNil(t, m.picker, "esc goes back, it does not close the picker")
	assert.False(t, m.picker.naming)
	assert.Zero(t, fake.createCalls)
	assert.Zero(t, fake.toggleCalls)
	view := pkScreen(m)
	assert.Contains(t, view, "+ Create a new list")
	assert.NotContains(t, view, "Nope")
	assert.Contains(t, pkHighlighted(t, view), "Create a new list")

	_, _ = m.Update(pkEnter)
	assert.Empty(t, m.picker.input.Value(), "the entry starts empty again")
}

func TestAnEmptyNameSaysSoAndKeepsTheEntryOpen(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)

	_, _ = m.Update(pkEnter)

	assert.True(t, m.picker.naming)
	assert.Contains(t, pkScreen(m), "Give the list a name.")
	assert.Equal(t, []string{"create:"}, fake.calls, "the store decides what a good name is")
	assert.Zero(t, fake.toggleCalls)
}

func TestATakenNameShowsTheErrorAndKeepsWhatWasTyped(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "weekend")

	_, _ = m.Update(pkEnter)

	view := pkScreen(m)
	assert.Contains(t, view, "You already have a list with that name.")
	assert.Contains(t, view, "weekend", "the name is still there")
	assert.Equal(t, "weekend", m.picker.input.Value())
	assert.True(t, m.picker.naming)
	assert.Zero(t, fake.toggleCalls)
}

func TestTheNameErrorGoesAwayWhenTheNameChanges(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEnter)
	require.Contains(t, pkScreen(m), "Give the list a name.")

	pkType(m, "K")

	assert.NotContains(t, pkScreen(m), "Give the list a name.")
}

func TestACreateThatFailsShowsTheStoresWords(t *testing.T) {
	fake := pkNewLists()
	fake.createErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")

	_, _ = m.Update(pkEnter)

	assert.Contains(t, pkScreen(m), "The disk is full.")
	assert.Equal(t, "Trip", m.picker.input.Value())
	assert.True(t, m.picker.naming)
	assert.Zero(t, fake.toggleCalls)
	assert.Len(t, fake.lists, 3, "no list appeared")
}

func TestAListMadeButNotFilledCanBeRetriedWithoutASecondList(t *testing.T) {
	fake := pkNewLists()
	fake.toggleErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)

	view := pkScreen(m)
	assert.Contains(t, view, `Made the list "Trip", but could not add "Frieren".`)
	assert.Contains(t, view, "The disk is full.")
	assert.True(t, m.picker.naming, "the entry stays open")
	assert.Equal(t, "Trip", m.picker.input.Value())

	fake.toggleErr = nil
	_, _ = m.Update(pkEnter)

	assert.Equal(t, 1, fake.createCalls, "enter again does not make a second Trip")
	assert.True(t, fake.has("new1", "frieren"))
	assert.False(t, m.picker.naming)
	assert.Contains(t, pkScreen(m), `Created "Trip" and added "Frieren".`)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "Trip")
}

func TestChangingTheNameAfterAHalfDoneCreateMakesAnotherList(t *testing.T) {
	fake := pkNewLists()
	fake.toggleErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)
	fake.toggleErr = nil
	pkType(m, "2")

	_, _ = m.Update(pkEnter)

	assert.Equal(t, 2, fake.createCalls)
	assert.True(t, fake.has("new2", "frieren"))
	assert.Contains(t, pkScreen(m), `Created "Trip2" and added "Frieren".`)
}

func TestEscAfterAHalfDoneCreateShowsTheNewListEmpty(t *testing.T) {
	fake := pkNewLists()
	fake.toggleErr = errors.New("the disk is full")
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)

	_, _ = m.Update(pkEsc)

	view := pkScreen(m)
	assert.Contains(t, view, "Trip")
	assert.Contains(t, view, "empty")
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)
	assert.Contains(t, pkScreen(m), "You already have a list with that name.", "esc forgot the half-made list")
}

func TestCreatedNameKeepsTheStoresSpelling(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnter)
	pkType(m, "  Late   night ")
	_, _ = m.Update(pkEnter)

	assert.Equal(t, []string{"create:  Late   night "}, fake.calls[:1], "the name goes to the store as typed")
	assert.Contains(t, pkScreen(m), `Created "Late night" and added "Frieren".`, "and is shown as the store spells it")
}

func TestAPastedNameLandsInTheEntry(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)

	_, _ = m.Update(tea.PasteMsg{Content: "Pasted name"})

	assert.Equal(t, "Pasted name", m.picker.input.Value())
}

func TestTheNameEntryIsFocusedAndGetsTheCursor(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	cmd := func() tea.Cmd { _, c := m.Update(pkEnter); return c }()
	assert.True(t, m.picker.input.Focused())
	assert.NotNil(t, cmd, "the cursor starts blinking")
}

// ---- lists that cannot be read ---------------------------------------------------------

func TestListsThatCannotBeReadSayWhyAndCanBeClosed(t *testing.T) {
	fake := pkNewLists()
	fake.listsErr = errors.New("the file is unreadable")
	m, _ := pkResults(t, fake)

	_, _ = m.Update(pkCtrlL)

	require.NotNil(t, m.picker, "the picker still opens")
	view := pkScreen(m)
	assert.Contains(t, view, "Could not read your lists. The file is unreadable.")
	assert.NotContains(t, view, "+ Create a new list")
	assert.Zero(t, fake.memberCalls, "no list, no membership")
	assert.Equal(t, []Hint{{Key: "esc", Desc: "close", Essential: true}}, m.hints())

	for _, key := range []tea.KeyPressMsg{keyDown, pkEnter, pkEnd, pkUp, pkPgDn} {
		_, _ = m.Update(key)
	}
	assert.NotNil(t, m.picker)
	assert.Zero(t, fake.toggleCalls+fake.createCalls)

	_, _ = m.Update(pkEsc)
	assert.Nil(t, m.picker)
}

// ---- Favorites and the star on the row -------------------------------------------------

func TestAddingToFavoritesInThePickerStarsTheResultRow(t *testing.T) {
	fake := pkNewLists()
	m := pkOpen(t, fake)
	require.False(t, m.favKeys["frieren"])
	require.NotContains(t, pkScreen(m), "★ Frieren")
	_, _ = m.Update(keyDown) // Favorites
	_, _ = m.Update(pkEnter)

	assert.True(t, m.favKeys["frieren"], "the row's star follows")
	_, _ = m.Update(pkEsc)
	assert.Contains(t, ansi.Strip(m.results.View()), "★")
	assert.Contains(t, m.detailsContent().tip, "ctrl+f takes it off FAVORITES")
	assert.Equal(t, "unfavorite", m.hints()[2].Desc)
}

func TestTakingTheWorkOffFavoritesInThePickerUnstarsTheRow(t *testing.T) {
	fake := pkNewLists()
	fake.works[FavoritesListID] = append(fake.works[FavoritesListID], "frieren")
	m, _ := pkResults(t, fake)
	m.favKeys["frieren"] = true // what the store read said at the start
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(keyDown)
	require.Contains(t, pkHighlighted(t, pkScreen(m)), "in this list")

	_, _ = m.Update(pkEnter)

	assert.False(t, m.favKeys["frieren"])
	assert.Contains(t, pkScreen(m), `Removed "Frieren" from Favorites.`)
	_, _ = m.Update(pkEsc)
	assert.NotContains(t, ansi.Strip(m.results.View()), "★")
	assert.Equal(t, "favorite", m.hints()[2].Desc)
}

func TestOtherListsLeaveTheStarAlone(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	assert.Empty(t, m.favKeys)
}

func TestCreatingAListLeavesTheStarAlone(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)
	assert.Empty(t, m.favKeys)
}

func TestCtrlFStillWorksAfterThePickerClosed(t *testing.T) {
	fake := pkNewLists()
	m, store := pkResults(t, fake)
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(pkEsc)

	_, cmd := m.Update(ctrlF)

	assert.Nil(t, cmd)
	assert.True(t, store.Has("frieren"))
	assert.True(t, m.favKeys["frieren"])
	assert.Equal(t, "ctrl+f takes it off FAVORITES · ctrl+l puts it in a list", m.detailsContent().tip)
	assert.Zero(t, fake.toggleCalls, "ctrl+f changes nothing else")
	assert.Zero(t, fake.createCalls)
}

func TestFavoritesStarsStayInStepWhenReopened(t *testing.T) {
	fake := pkNewLists()
	m, _ := pkResults(t, fake)
	_, _ = m.Update(ctrlF) // FAVORITES via ctrl+f
	fake.works[FavoritesListID] = append(fake.works[FavoritesListID], "frieren")
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(keyDown)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "in this list")
}

// ---- reads ---------------------------------------------------------------------------

func TestOneReadOfTheListsAndTheMembershipPerOpening(t *testing.T) {
	fake := pkNewLists()
	m, _ := pkResults(t, fake)
	require.Zero(t, fake.reads(), "the results screen reads nothing from the lists on its own")

	_, _ = m.Update(pkCtrlL)

	assert.Equal(t, 1, fake.listCalls)
	assert.Equal(t, 1, fake.memberCalls)
	assert.Zero(t, fake.otherCalls)

	// Drawing, moving, resizing, toggling and creating read nothing more.
	for range 3 {
		_ = m.View()
	}
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkEnter)
	_, _ = m.Update(pkHome)
	_, _ = m.Update(pkEnter)
	pkType(m, "Trip")
	_, _ = m.Update(pkEnter)
	_ = m.View()
	assert.Equal(t, 1, fake.listCalls)
	assert.Equal(t, 1, fake.memberCalls)
	assert.Zero(t, fake.otherCalls)

	_, _ = m.Update(pkEsc)
	_, _ = m.Update(pkCtrlL)
	assert.Equal(t, 2, fake.listCalls, "another opening reads again, once")
	assert.Equal(t, 2, fake.memberCalls)
}

func TestMembershipIsAskedForTheHighlightedWork(t *testing.T) {
	fake := pkNewLists()
	m, _ := pkResults(t, fake)
	_, _ = m.Update(keyDown) // Naruto
	fake.works["kids"] = []string{"naruto"}
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(pkEnd)
	assert.Contains(t, pkScreen(m), `Add "Naruto" to a list`)
	assert.Contains(t, pkHighlighted(t, pkScreen(m)), "in this list")
}

// ---- discoverability -----------------------------------------------------------------

func TestWithListsTheFooterOffersCtrlLRightAfterCtrlF(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	hints := m.hints()
	var keys []string
	for _, h := range hints {
		keys = append(keys, h.Key)
	}
	assert.Equal(t, []string{"enter", "←→", "ctrl+f", "ctrl+l", "↑↓", "esc", "?", "q"}, keys)
	assert.Equal(t, Hint{Key: "ctrl+l", Desc: "lists"}, hints[3])
	assert.False(t, hints[3].Essential)
}

func TestWithListsTheTipSaysBothKeysInWords(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	assert.Equal(t, "ctrl+f adds it to FAVORITES · ctrl+l puts it in a list", m.detailsContent().tip)
	_, _ = m.Update(ctrlF)
	assert.Equal(t, "ctrl+f takes it off FAVORITES · ctrl+l puts it in a list", m.detailsContent().tip)
}

func TestWithListsButNoFavoritesStoreTheTipNamesCtrlL(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	m.favs = nil
	assert.Equal(t, "ctrl+l puts it in a list", m.detailsContent().tip)
}

func TestTheTipIsStillTheFirstThingDroppedOnAShortScreen(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	theme := CurrentTheme()
	content := m.detailsContent()
	bare := *content
	bare.tip = ""
	const width = 60
	need := len(panelLines(&theme, &bare, width-1-4, false)) + 3

	withTip := ansi.Strip(fitPanel(&theme, content, width, 40))
	assert.Contains(t, withTip, "ctrl+l puts it in a list")
	short := ansi.Strip(fitPanel(&theme, content, width, need))
	assert.NotEmpty(t, short, "the panel stays")
	assert.NotContains(t, short, "ctrl+l")
	assert.Contains(t, short, "Frieren")
}

func TestTheDetailsPanelShowsTheTipOnAWideScreen(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	view := pkScreen(m)
	assert.Contains(t, view, "ctrl+l puts")
	assert.Contains(t, view, "it in a list")
}

func TestThePlainScreenShowsTheTipBelowTheList(t *testing.T) {
	pkPlain(t)
	m, _ := pkResults(t, pkNewLists())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	assert.Contains(t, pkScreen(m), "ctrl+l puts it in a list")
}

func TestTheHelpOverlayListsCtrlL(t *testing.T) {
	m, _ := pkResults(t, pkNewLists())
	m.help()
	require.True(t, m.showHelp)
	view := pkScreen(m)
	assert.Contains(t, view, "THIS SCREEN")
	assert.Contains(t, view, "ctrl+l")
	assert.Contains(t, view, "put it in a list")
	assert.Contains(t, view, "ctrl+f")
	assert.Contains(t, view, "quit")
	assert.LessOrEqual(t, strings.Count(view, "\n")+1, 24)
}

// ---- layout ---------------------------------------------------------------------------

func TestResizeRelaysOutTheOpenPicker(t *testing.T) {
	fake := pkBigLists(40)
	m := pkOpen(t, fake)
	_, _ = m.Update(pkEnd)

	for _, size := range [][2]int{{100, 40}, {60, 14}, {80, 24}, {60, 20}} {
		_, _ = m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := pkScreen(m)
		assert.LessOrEqual(t, strings.Count(view, "\n")+1, size[1], "%v", size)
		assert.Contains(t, pkHighlighted(t, view), "List 40", "%v: the last list is on screen", size)
		assert.Equal(t, pickerLayoutFor(size[1]-shellChromeHeight).rows, m.picker.room, "%v", size)
	}
}

func TestResizeKeepsWhatWasTypedInTheNameEntry(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	_, _ = m.Update(pkEnter)
	pkType(m, "Late")
	_, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	assert.Equal(t, "Late", m.picker.input.Value())
	assert.Contains(t, pkScreen(m), "Late")
	assert.True(t, m.picker.naming)
}

// pkBigLists is a fake with n user lists besides Favorites.
func pkBigLists(n int) *pickerLists {
	fake := pkNewLists()
	fake.lists = fake.lists[:1]
	for i := 1; i <= n; i++ {
		fake.lists = append(fake.lists, HomeList{ID: fmt.Sprintf("l%d", i), Name: fmt.Sprintf("List %d", i)})
	}
	return fake
}

func TestManyListsScrollAndTheHighlightStaysOnScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			fake := pkBigLists(40)
			m, _ := pkResults(t, fake)
			_, _ = m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			_, _ = m.Update(pkCtrlL)
			require.NotNil(t, m.picker)

			for row := 0; row <= 41; row++ {
				view := pkScreen(m)
				assert.LessOrEqual(t, strings.Count(view, "\n")+1, size[1], "row %d", row)
				want := "Create a new list"
				if row > 0 {
					want = fmt.Sprintf("List %d", row-1)
					if row == 1 {
						want = "Favorites"
					}
				}
				assert.Contains(t, pkHighlighted(t, view), want, "row %d:\n%s", row, view)
				for _, line := range strings.Split(view, "\n") {
					assert.LessOrEqual(t, ansi.StringWidth(line), size[0])
				}
				_, _ = m.Update(keyDown)
			}
		})
	}
}

func TestScrollingNotesSayHowManyRowsAreOutOfSight(t *testing.T) {
	m := pkOpen(t, pkBigLists(40))
	view := pkScreen(m)
	assert.NotContains(t, view, "more above")
	assert.Regexp(t, `↓ \d+ more below`, view)

	_, _ = m.Update(pkEnd)
	view = pkScreen(m)
	assert.Regexp(t, `↑ \d+ more above`, view)
	assert.NotContains(t, view, "more below")

	_, _ = m.Update(pkHome)
	_, _ = m.Update(pkPgDn)
	_, _ = m.Update(pkPgDn)
	view = pkScreen(m)
	assert.Regexp(t, `↑ \d+ more above`, view)
	assert.Regexp(t, `↓ \d+ more below`, view)
}

func TestScrollingNotesInPlainModeAreAscii(t *testing.T) {
	pkPlain(t)
	m := pkOpen(t, pkBigLists(40))
	_, _ = m.Update(pkPgDn)
	_, _ = m.Update(pkPgDn)
	view := pkScreen(m)
	assert.Regexp(t, `v \d+ more below`, view)
	assert.Regexp(t, `\^ \d+ more above`, view)
}

func TestNoScrollNotesWhenEveryRowFits(t *testing.T) {
	view := pkScreen(pkOpen(t, pkNewLists()))
	assert.NotContains(t, view, "more above")
	assert.NotContains(t, view, "more below")
}

func TestPageKeysMoveAWindowAtATime(t *testing.T) {
	m := pkOpen(t, pkBigLists(40))
	room := m.picker.room
	require.Greater(t, room, 4)
	_, _ = m.Update(pkPgDn)
	assert.Equal(t, room-1, m.picker.cursor)
	_, _ = m.Update(pkPgDn)
	assert.Equal(t, 2*(room-1), m.picker.cursor)
	_, _ = m.Update(pkPgUp)
	assert.Equal(t, room-1, m.picker.cursor)
}

func TestThePickerNeverOutgrowsAnyBodySize(t *testing.T) {
	for _, naming := range []bool{false, true} {
		for height := 1; height <= 30; height++ {
			for _, width := range []int{12, 24, 40, 60, 80, 120} {
				fake := pkBigLists(12)
				theme := CurrentTheme()
				p := newPickerState(&theme, fake, HomeFavorite{Key: "frieren"}, strings.Repeat("A very long title ", 8), map[string]bool{})
				p.fit(width, height)
				if naming {
					_ = p.startNaming()
					p.nameErr = "Something went wrong with a rather long sentence to wrap around."
				}
				for range 14 {
					view := p.view(width, height)
					if lines := strings.Count(view, "\n") + 1; !naming && lines > height {
						t.Fatalf("%dx%d: %d lines\n%s", width, height, lines, view)
					}
					for _, line := range strings.Split(view, "\n") {
						if w := ansi.StringWidth(line); w > width {
							t.Fatalf("%dx%d: a line is %d wide: %q", width, height, w, line)
						}
					}
					if !naming && height >= 1 && width >= 24 {
						require.Contains(t, ansi.Strip(view), "▶", "%dx%d", width, height)
					}
					p.move(3)
					p.fit(width, height)
				}
			}
		}
	}
}

func TestALongTitleIsShortenedInTheHeadingAndTheStatus(t *testing.T) {
	long := strings.Repeat("Mahou Shoujo ", 6)
	fake := pkNewLists()
	prev := newFavoriteLists
	newFavoriteLists = func() FavoriteLists { return fake }
	t.Cleanup(func() { newFavoriteLists = prev })
	m := newAnimeResultsModel([]*models.Anime{{Name: long, Source: "x"}})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	_, _ = m.Update(pkCtrlL)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)
	_, _ = m.Update(pkEnter)

	view := pkScreen(m)
	assert.Contains(t, view, `…" to a list`)
	assert.Contains(t, view, "to Weekend.", "the list's name is never the part that is cut")
	assert.LessOrEqual(t, strings.Count(view, "\n")+1, 20)
	for _, line := range strings.Split(view, "\n") {
		assert.LessOrEqual(t, ansi.StringWidth(line), 60)
	}
}

func TestEveryPickerScreenFitsTheShellAtThePromisedSizes(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}, {100, 40}} {
		fake := pkNewLists()
		fake.createErr = errors.New("the disk is full")
		m, _ := pkResults(t, fake)
		_, _ = m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		_, _ = m.Update(pkCtrlL)
		check := func(what string) {
			view := pkScreen(m)
			assert.LessOrEqual(t, strings.Count(view, "\n")+1, size[1], "%v %s", size, what)
			for _, line := range strings.Split(view, "\n") {
				assert.LessOrEqual(t, ansi.StringWidth(line), size[0], "%v %s", size, what)
			}
		}
		check("list")
		_, _ = m.Update(pkEnter)
		check("name")
		pkType(m, "Trip")
		_, _ = m.Update(pkEnter)
		check("name error")
		assert.Contains(t, pkScreen(m), "disk is full")
		_, _ = m.Update(pkEsc)
		check("back")
	}
}

func TestRowsKeepTheirCountsAtTheRightEdge(t *testing.T) {
	m := pkOpen(t, pkNewLists())
	for _, line := range strings.Split(pkScreen(m), "\n") {
		if strings.Contains(line, "Weekend") {
			assert.True(t, strings.HasSuffix(strings.TrimRight(line, " "), "1 title"), line)
		}
	}
}

func TestNarrowRowsShortenTheMembershipNoteButKeepTheTick(t *testing.T) {
	theme := CurrentTheme()
	row := pickerRow{name: "Weekend marathon", count: "12 titles", in: true}
	wide := ansi.Strip(renderPickerRow(&theme, row, false, 80))
	narrow := ansi.Strip(renderPickerRow(&theme, row, false, 36))
	assert.Contains(t, wide, "✓ in this list")
	assert.NotContains(t, narrow, "in this list")
	assert.Contains(t, narrow, "✓")
	assert.LessOrEqual(t, ansi.StringWidth(narrow), 36)
}

func TestPlainNarrowRowsStillCarryBrackets(t *testing.T) {
	pkPlain(t)
	theme := CurrentTheme()
	row := pickerRow{name: "Weekend marathon", count: "12 titles", in: true}
	narrow := ansi.Strip(renderPickerRow(&theme, row, true, 36))
	assert.Contains(t, narrow, "[x]")
	assert.NotContains(t, narrow, "✓")
	assert.LessOrEqual(t, ansi.StringWidth(narrow), 36)
}

func TestPickerCountLabelWords(t *testing.T) {
	assert.Equal(t, "empty", pickerCountLabel(0))
	assert.Equal(t, "empty", pickerCountLabel(-1))
	assert.Equal(t, "1 title", pickerCountLabel(1))
	assert.Equal(t, "2 titles", pickerCountLabel(2))
	assert.Equal(t, "40 titles", pickerCountLabel(40))
}

func TestWindowStartKeepsTheCursorInView(t *testing.T) {
	assert.Equal(t, 0, pickerWindowStart(0, 0, 10, 4))
	assert.Equal(t, 0, pickerWindowStart(0, 3, 10, 4))
	assert.Equal(t, 1, pickerWindowStart(0, 4, 10, 4))
	assert.Equal(t, 6, pickerWindowStart(0, 9, 10, 4))
	assert.Equal(t, 6, pickerWindowStart(9, 9, 10, 4))
	assert.Equal(t, 2, pickerWindowStart(5, 2, 10, 4))
	assert.Equal(t, 0, pickerWindowStart(5, 1, 3, 4), "fewer rows than room starts at the top")
	assert.Equal(t, 6, pickerWindowStart(9, 9, 10, 4), "the window never runs past the end")
	assert.Equal(t, 5, pickerWindowStart(8, 5, 10, 0), "a room of 0 counts as 1")
}

func TestLayoutSharesTheHeightOut(t *testing.T) {
	for height := 1; height <= 40; height++ {
		l := pickerLayoutFor(height)
		used := l.rows + l.statusLines
		if l.title {
			used++
		}
		if l.spacers {
			used += 2
		}
		assert.GreaterOrEqual(t, l.rows, 1, height)
		if height >= 9 {
			assert.Equal(t, height, used, "no line wasted at %d", height)
		}
		assert.LessOrEqual(t, used, max(height, 1)+0, height)
	}
	assert.Equal(t, pickerLayout{title: true, spacers: true, statusLines: 2, rows: 4}, pickerLayoutFor(9))
	assert.Equal(t, pickerLayout{title: true, statusLines: 2, rows: 2}, pickerLayoutFor(5))
}
