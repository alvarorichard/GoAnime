package tui

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Edge cases for FAVORITOS on the result screen and the search home, written
// from a hand-run mutation pass: each test here fails on a realistic mutant of
// anime_results.go or search_home.go that the first tests let through.
//
// These tests swap package globals (newFavorites, the plain mode) and so are
// not parallel; the shared temp store TestMain installs is never touched.

var (
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keyRight = tea.KeyPressMsg{Code: tea.KeyRight}
)

func storedFavorites(t *testing.T, store *favorites.Store) []favorites.Entry {
	t.Helper()
	saved, err := store.List()
	require.NoError(t, err)
	return saved
}

// ---- result screen -------------------------------------------------------------

func TestFavoriteRemembersTheVersionOfTheHighlightedWorkNotTheFirst(t *testing.T) {
	m, store := resultsWithFavorites(t)
	_, _ = m.Update(keyRight) // Frieren: pick the HiAnime copy
	_, _ = m.Update(keyDown)  // on to Naruto, which only has AnimeFire
	require.Contains(t, m.detailsContent().title, "Naruto")

	_, _ = m.Update(ctrlF)

	saved := storedFavorites(t, store)
	require.Len(t, saved, 1)
	assert.Equal(t, "Naruto", saved[0].Title)
	assert.Equal(t, "naruto", saved[0].Key)
	assert.Equal(t, "2002", saved[0].Year)
	assert.Equal(t, "AnimeFire", saved[0].Source, "Naruto's own version, not Frieren's HiAnime pick")
}

func TestFavoriteKeyIsTheNormalizedTitleSoPunctuationAndCaseDoNotSplitAWork(t *testing.T) {
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	prev := newFavorites
	newFavorites = func() favoritesAPI { return store }
	t.Cleanup(func() { newFavorites = prev })
	m := newAnimeResultsModel([]*models.Anime{
		{Name: "[PT-BR] Naruto: Shippuden (Legendado)", Source: "Animefire.io", Year: "2007", MediaType: models.MediaTypeAnime},
	})

	_, _ = m.Update(ctrlF)

	saved := storedFavorites(t, store)
	require.Len(t, saved, 1)
	assert.Equal(t, "naruto shippuden", saved[0].Key, "no colon, no tag, one space")
	assert.Equal(t, "Naruto: Shippuden", saved[0].Title, "the title keeps its punctuation")
	assert.True(t, store.Has("naruto shippuden"))
	assert.True(t, strings.HasPrefix(m.detailsContent().title, "★ ") || strings.HasPrefix(m.detailsContent().title, "* "),
		"the panel marks it though the stored key differs from the raw title: %q", m.detailsContent().title)
}

func TestAllTheVersionsOfAWorkShareOneFavoriteMark(t *testing.T) {
	m, store := resultsWithFavorites(t)
	_, _ = m.Update(ctrlF)
	require.True(t, store.Has("frieren"))
	_, _ = m.Update(keyRight) // another version of the same work
	assert.NotEqual(t, "Frieren", m.detailsContent().title, "still marked on the other version")
	_, _ = m.Update(keyDown) // another work
	assert.Equal(t, "Naruto", m.detailsContent().title, "and not marked on a different work")
}

func TestATitleWithNoLettersCannotBeFavoritedButDoesNotBreakTheScreen(t *testing.T) {
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	prev := newFavorites
	newFavorites = func() favoritesAPI { return store }
	t.Cleanup(func() { newFavorites = prev })
	m := newAnimeResultsModel([]*models.Anime{{Name: "!!!", Source: "Animefire.io", MediaType: models.MediaTypeAnime}})

	_, cmd := m.Update(ctrlF)

	assert.Nil(t, cmd)
	assert.Contains(t, m.notice, "Could not update FAVORITOS")
	assert.Empty(t, storedFavorites(t, store))
}

func TestCtrlFWithoutAFavoritesStoreDoesNothing(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	m.favs = nil

	assert.NotPanics(t, func() {
		_, cmd := m.Update(ctrlF)
		assert.Nil(t, cmd)
	})
	assert.Equal(t, "Frieren", m.detailsContent().title, "nothing to mark with")
	assert.Empty(t, m.notice)
}

func TestCtrlFOnAnEmptyListDoesNothing(t *testing.T) {
	m, store := resultsWithFavorites(t)
	_, _ = m.Update(press('z', "z")) // filter that matches nothing
	for _, r := range "zzzz" {
		_, _ = m.Update(press(r, string(r)))
	}
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // apply the filter
	assert.NotPanics(t, func() { _, _ = m.Update(ctrlF) })
	assert.Empty(t, storedFavorites(t, store))
}

func TestCtrlFWithAGrouplessSelectionIsIgnored(t *testing.T) {
	m, store := resultsWithFavorites(t)
	m.results.SetItems([]list.Item{animeGroupItem{group: nil}})

	assert.NotPanics(t, func() { m.toggleFavorite() })
	assert.Empty(t, storedFavorites(t, store))
}

func TestAFailedFavoriteNoticeIsOneLine(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	m.favs = errFavorites{errors.New("disk\nfull\x1b[31m!\x1b[0m\r\nreally")}

	_, _ = m.Update(ctrlF)

	assert.Equal(t, "Could not update FAVORITOS: disk full! really", m.notice)
}

type errFavorites struct{ err error }

func (errFavorites) Keys() map[string]bool                  { return map[string]bool{} }
func (e errFavorites) Toggle(favorites.Entry) (bool, error) { return false, e.err }

func TestAFailedFavoriteTakesItsRowsFromTheList(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	before := m.results.Height()
	m.favs = errFavorites{errors.New("disk full")}

	_, _ = m.Update(ctrlF)

	require.NotEmpty(t, m.notice)
	assert.Less(t, m.results.Height(), before, "the notice's rows come off the saved, like the other notices")
	assert.Contains(t, ansi.Strip(m.View().Content), "Could not update FAVORITOS")
}

func TestFavoritingInPlainModeUpdatesTheReservedPanel(t *testing.T) {
	plainMu.Lock()
	prevOverride := plainOverride
	plainMu.Unlock()
	SetPlain(true)
	t.Cleanup(func() {
		plainMu.Lock()
		plainOverride = prevOverride
		plainMu.Unlock()
	})
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	require.NotEmpty(t, m.plainPanel, "plain mode draws the panel under the list")
	require.NotContains(t, ansi.Strip(m.plainPanel), "★")

	_, _ = m.Update(ctrlF)

	assert.Regexp(t, `[★*] Frieren`, ansi.Strip(m.plainPanel), "the panel is redrawn with the mark")
}

func TestHelpOverlayListsCtrlF(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, _ = m.Update(press('?', "?"))
	require.True(t, m.showHelp)
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "ctrl+f")
	assert.Contains(t, view, "FAVORITOS")
}

func TestFooterHintsKeepCtrlFBeforeTheBackAndQuitKeys(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	var keys []string
	for _, h := range m.hints() {
		keys = append(keys, h.Key)
	}
	assert.Contains(t, keys, "ctrl+f")
	assert.Contains(t, keys, "enter", "the favorite hint did not push the main ones out")
}

// ---- search home -----------------------------------------------------------------

func homeWith(rows []HomeFavorite, nCont, nRecent int) (*searchHomeModel, *favoritesFake) {
	continueAll, recent := homeTestEntries()
	f := &favoritesFake{}
	return newSearchHomeModel(f.options(rows, continueAll[:nCont], recent[:nRecent])), f
}

// threeFavorites() is a function: the home model removes rows in place, which
// would otherwise corrupt a shared slice between tests.
func threeFavorites() []HomeFavorite {
	return []HomeFavorite{
		{Key: "a", Title: "Alpha", Year: "2001", Media: "Anime", Source: "HiAnime"},
		{Key: "b", Title: "Bravo", Year: "2002", Media: "Movie", Source: "StartFlix"},
		{Key: "c", Title: "Charlie", Year: "2003", Media: "TV", Source: "Animefire.io"},
	}
}

func TestFavoritesAloneAreTheWholeList(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: threeFavorites()})
	require.Len(t, m.rows, 3)
	assert.True(t, m.focusInput, "the box has the focus first")
	downTo(t, m, 0)
	downTo(t, m, 2)
	_, _ = m.Update(keyDown)
	assert.True(t, m.focusInput || m.cursor == 2, "past the end it wraps to the box or stays on the last row")
}

func TestEnterOnEachFavoriteSearchesThatTitle(t *testing.T) {
	for i, want := range []string{"Alpha", "Bravo", "Charlie"} {
		m, _ := homeWith(threeFavorites(), 2, 2)
		downTo(t, m, 2+i) // after two CONTINUE rows
		require.Equal(t, homeRowFavorite, m.rows[m.cursor].kind)

		_, cmd := m.Update(press(tea.KeyEnter, ""))

		require.NotNil(t, cmd, want)
		assert.Equal(t, SearchHomeResult{Action: HomeSearch, Query: want}, m.result, "title, not key; every scope")
	}
}

func TestRemovingTheSecondFavoriteKeepsTheFirstAndThird(t *testing.T) {
	m, f := homeWith(threeFavorites(), 0, 0)
	downTo(t, m, 1)

	_, _ = m.Update(press('x', "x"))

	assert.Equal(t, []string{"b"}, f.removed, "the key of the highlighted row")
	require.Len(t, m.favoriteAll, 2)
	assert.Equal(t, "Alpha", m.favoriteAll[0].Title)
	assert.Equal(t, "Charlie", m.favoriteAll[1].Title)
	require.Len(t, m.rows, 2, "the rows follow the list")
	assert.Equal(t, 1, m.cursor, "the cursor stays where it was, now on Charlie")
	assert.NotContains(t, screen(m), "Bravo")
}

func TestRemovingTheLastRowMovesTheCursorUp(t *testing.T) {
	m, _ := homeWith(threeFavorites(), 0, 0)
	downTo(t, m, 2)

	_, _ = m.Update(press('x', "x"))

	require.Len(t, m.rows, 2)
	assert.Equal(t, 1, m.cursor, "clamped to the new last row, not left past the end")
	row, ok := m.currentRow()
	require.True(t, ok, "the highlighted row still exists")
	assert.Equal(t, homeRowFavorite, row.kind)
	_, cmd := m.Update(press(tea.KeyEnter, ""))
	require.NotNil(t, cmd)
	assert.Equal(t, "Bravo", m.result.Query)
}

func TestRemovingTheLastFavoriteInTheMiddleOfTheListClampsToTheListEnd(t *testing.T) {
	m, _ := homeWith(threeFavorites()[:1], 3, 2) // 3 continue, 1 favorite, 2 recent
	downTo(t, m, 3)
	_, _ = m.Update(press('x', "x"))
	require.Len(t, m.rows, 5)
	assert.Equal(t, 3, m.cursor, "the cursor stays on the same index: the first RECENT row")
	assert.Equal(t, homeRowRecent, m.rows[m.cursor].kind)
}

func TestRemovingTheOnlyFavoriteGivesTheFocusBackToTheBox(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(threeFavorites()[:1], nil, nil))
	downTo(t, m, 0)
	_, _ = m.Update(press('x', "x"))
	assert.Empty(t, m.rows)
	assert.True(t, m.focusInput)
	assert.Contains(t, screen(m), "Nothing here yet", "the section stays, with its hint, so it can be found again")
}

func TestAStartupWithNoRowsFocusesTheBox(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	assert.True(t, m.focusInput)
	assert.Empty(t, m.rows)
}

func TestFavoriteRowsShowTitleYearTypeAndSource(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: threeFavorites()})
	view := screen(m)
	for _, want := range []string{"Alpha", "2001 · Anime", "HiAnime", "Bravo", "2002 · Movie", "StartFlix", "Charlie", "2003 · TV", "AnimeFire"} {
		assert.Contains(t, view, want)
	}
	assert.NotContains(t, view, "Animefire.io", "the source reads the way the result screen names it")
}

func TestFavoriteRowsWithoutYearOrTypeShowNoStraySeparator(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: []HomeFavorite{
		{Key: "a", Title: "OnlyTitle"}, {Key: "b", Title: "WithYear", Year: "1999"}, {Key: "c", Title: "WithType", Media: "TV"},
	}})
	view := screen(m)
	assert.Contains(t, view, "1999")
	assert.Contains(t, view, "TV")
	assert.NotContains(t, view, "· ", "a missing half leaves no dangling dot")
	assert.NotContains(t, view, " ·\n")
}

func TestFavoriteTitlesAreDrawnOnOneLineWithoutEscapes(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: []HomeFavorite{
		{Key: "a", Title: "Bad\x1b[31m\nTitle\x07", Year: "20\n20", Media: "A\x1bnime", Source: "S\nrc"},
	}})
	raw := m.View().Content
	assert.NotContains(t, raw, "\x07")
	assert.NotContains(t, ansi.Strip(raw), "\x1b")
	assert.Contains(t, ansi.Strip(raw), "Bad Title")
}

func TestTheCursorMarkIsOnTheHighlightedFavoriteOnly(t *testing.T) {
	m, _ := homeWith(threeFavorites(), 1, 1)
	downTo(t, m, 2) // second favorite
	var marked []string
	for _, line := range strings.Split(screen(m), "\n") {
		if strings.Contains(line, "▶") || strings.Contains(line, "> ") {
			marked = append(marked, line)
		}
	}
	require.NotEmpty(t, marked, screen(m))
	joined := strings.Join(marked, "\n")
	assert.Contains(t, joined, "Bravo", "the mark is on the highlighted favorite")
	for _, other := range []string{"Alpha", "Charlie", "Mukashi", "namakura"} {
		assert.NotContains(t, joined, other)
	}
}

func TestClearQuestionCountsOnlyTheFavorites(t *testing.T) {
	m, f := homeWith(threeFavorites()[:1], 3, 2)
	downTo(t, m, 3)
	_, _ = m.Update(press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 1 favorite from FAVORITOS?", "one favorite, not the 2 searches or 3 episodes")
	assert.Zero(t, f.cleared)

	m, _ = homeWith(threeFavorites(), 1, 2)
	downTo(t, m, 1)
	_, _ = m.Update(press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 3 favorites from FAVORITOS?")
}

func TestShiftXDoesNothingWithoutAClearerOrWithoutRows(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: threeFavorites()}) // no ClearFavorites
	downTo(t, m, 0)
	_, _ = m.Update(press('X', "X"))
	assert.Equal(t, clearNothing, m.clearing, "the host cannot clear: no question")

	m = newSearchHomeModel((&favoritesFake{}).options(nil, nil, nil))
	_, _ = m.Update(press('X', "X"))
	assert.Equal(t, clearNothing, m.clearing)
}

func TestEscAlsoKeepsTheFavorites(t *testing.T) {
	m, f := homeWith(threeFavorites(), 0, 0)
	downTo(t, m, 0)
	_, _ = m.Update(press('X', "X"))
	require.Equal(t, clearFavoriteRows, m.clearing)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Equal(t, clearNothing, m.clearing)
	assert.Zero(t, f.cleared)
	assert.Len(t, m.favoriteAll, 3)
}

func TestStrayKeysDuringTheClearQuestionTouchNothing(t *testing.T) {
	m, f := homeWith(threeFavorites(), 0, 0)
	downTo(t, m, 0)
	_, _ = m.Update(press('X', "X"))
	for _, k := range []tea.KeyPressMsg{press('x', "x"), press('r', "r"), press('a', "a"), {Code: tea.KeyEnter}, {Code: tea.KeyDown}} {
		_, cmd := m.Update(k)
		assert.Nil(t, cmd, k.String())
	}
	assert.Equal(t, clearFavoriteRows, m.clearing, "still asking")
	assert.Empty(t, f.removed, "x did not remove the row behind the question")
	assert.Zero(t, f.cleared)
	assert.Len(t, m.favoriteAll, 3)
	assert.Empty(t, m.input.Value(), "nothing typed into the box")
}

func TestAFailedClearKeepsTheFavoritesAndSaysWhy(t *testing.T) {
	f := &favoritesFake{err: errors.New("read-only disk")}
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel(f.options(threeFavorites(), continueAll, recent))
	downTo(t, m, 3)
	_, _ = m.Update(press('X', "X"))
	_, _ = m.Update(press('y', "y"))

	assert.Equal(t, 1, f.cleared, "it was tried")
	assert.Len(t, m.favoriteAll, 3, "and the rows stay")
	assert.Equal(t, clearNothing, m.clearing)
	assert.Contains(t, m.notice, "clear FAVORITOS")
	assert.Contains(t, m.notice, "read-only disk")
}

func TestClearingFavoritesFromTheListEndLeavesAValidCursor(t *testing.T) {
	m, _ := homeWith(threeFavorites(), 1, 0)
	downTo(t, m, 3) // last row overall
	_, _ = m.Update(press('X', "X"))
	_, _ = m.Update(press('y', "y"))
	require.Len(t, m.rows, 1)
	assert.Equal(t, 0, m.cursor, "clamped onto the one CONTINUE row left")
	_, ok := m.currentRow()
	assert.True(t, ok)
}

func TestHomeOptionsWireTheFavoritesAndTheirActions(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(threeFavorites(), nil, nil))
	assert.Len(t, m.favoriteAll, 3)
	require.NotNil(t, m.deleteFavorite)
	require.NotNil(t, m.clearFavorites)
	require.NoError(t, m.deleteFavorite("a"))
	require.NoError(t, m.clearFavorites())
	assert.Equal(t, []string{"a"}, f.removed)
	assert.Equal(t, 1, f.cleared)
}

func TestFavoritosSectionIsOnTheMenuWhateverTheOtherListsHold(t *testing.T) {
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})
	view := screen(m)
	assert.Contains(t, view, "FAVORITOS")
	assert.Contains(t, view, "Nothing here yet")
	assert.Less(t, indexOf(view, "CONTINUE"), indexOf(view, "FAVORITOS"))
	assert.Less(t, indexOf(view, "FAVORITOS"), indexOf(view, "RECENT"))
}

// ---- the whole path, results screen to home screen -------------------------------

// TestFavoritesFlowFromTheResultScreenToTheHomeScreen toggles works on the real
// result model, reads the file the way the handlers do (newest first, same
// fields), puts the rows on the real home model and walks them with the keys.
func TestFavoritesFlowFromTheResultScreenToTheHomeScreen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goanime", "favorites.json")
	store := favorites.Open(path)
	prev := newFavorites
	newFavorites = func() favoritesAPI { return favorites.Open(path) } // a fresh instance, like a new run
	t.Cleanup(func() { newFavorites = prev })

	m := newAnimeResultsModel([]*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[PT-BR] Naruto", Source: "Animefire.io", Year: "2002", MediaType: models.MediaTypeAnime},
		{Name: "[Movie] A Origem", Source: "StartFlix", Year: "2010", MediaType: models.MediaTypeMovie},
	})
	// favorite Frieren (HiAnime copy), then A Origem; Naruto stays out
	_, _ = m.Update(keyRight)
	_, _ = m.Update(ctrlF)
	time.Sleep(2 * time.Millisecond) // distinct AddedAt, newest = A Origem
	_, _ = m.Update(keyDown)
	_, _ = m.Update(keyDown)
	require.Contains(t, m.detailsContent().title, "Origem")
	_, _ = m.Update(ctrlF)

	// disk
	saved, err := store.List()
	require.NoError(t, err)
	require.Len(t, saved, 2)
	assert.Equal(t, []string{"a origem", "frieren"}, []string{saved[0].Key, saved[1].Key}, "newest first")
	assert.Equal(t, "HiAnime", saved[1].Source)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"added_at"`)

	// home rows as the handlers build them
	rows := make([]HomeFavorite, 0, len(saved))
	for _, f := range saved {
		rows = append(rows, HomeFavorite{Key: f.Key, Title: SingleLine(f.Title), Year: f.Year, Media: f.Media, Source: f.Source})
	}
	removed := ""
	home := newSearchHomeModel(SearchHomeOptions{
		Favorites:      rows,
		DeleteFavorite: func(k string) error { removed = k; return store.Remove(k) },
		ClearFavorites: store.Clear,
	})
	view := screen(home)
	assert.Contains(t, view, "FAVORITOS")
	assert.Less(t, indexOf(view, "A Origem"), indexOf(view, "Frieren"), "newest on top")
	assert.Contains(t, view, "2023 · Anime")
	assert.Contains(t, view, "2010 · Movie")

	// enter on the second row searches Frieren in every scope
	home.scopeIdx = 1
	downTo(t, home, 1)
	_, cmd := home.Update(press(tea.KeyEnter, ""))
	require.NotNil(t, cmd)
	assert.Equal(t, SearchHomeResult{Action: HomeSearch, Query: "Frieren"}, home.result)

	// x on the first row removes A Origem from the file
	home = newSearchHomeModel(SearchHomeOptions{
		Favorites:      rows,
		DeleteFavorite: func(k string) error { removed = k; return store.Remove(k) },
	})
	downTo(t, home, 0)
	_, _ = home.Update(press('x', "x"))
	assert.Equal(t, "a origem", removed)
	assert.False(t, store.Has("a origem"))
	assert.True(t, store.Has("frieren"))
	assert.True(t, favorites.Open(path).Has("frieren"), "a later run still sees Frieren")

	// and ctrl+f on the result screen again toggles the survivor off
	m2 := newAnimeResultsModel([]*models.Anime{{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime}})
	require.NotEqual(t, "Frieren", m2.detailsContent().title, "a new run opens it already marked")
	_, _ = m2.Update(ctrlF)
	assert.False(t, store.Has("frieren"))
}

func TestResultScreenAgainstAReadOnlyDirectoryShowsTheNotice(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permission bits and a non-root user")
	}
	dir := t.TempDir()
	prev := newFavorites
	newFavorites = func() favoritesAPI { return favorites.Open(filepath.Join(dir, "favorites.json")) }
	t.Cleanup(func() { newFavorites = prev })
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	m := newAnimeResultsModel([]*models.Anime{{Name: "[PT-BR] Frieren", Source: "Animefire.io", MediaType: models.MediaTypeAnime}})

	_, cmd := m.Update(ctrlF)

	assert.Nil(t, cmd)
	assert.Contains(t, m.notice, "Could not update FAVORITOS: favorites:")
	assert.Equal(t, "Frieren", m.detailsContent().title, "not marked: it was not saved")
}
