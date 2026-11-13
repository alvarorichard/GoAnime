//go:build reallists

package tui

// The cascade of "lists beside Favorites", end to end on the REAL store and
// the REAL screens:
//
//	results model -> ctrl+l picker -> file on disk -> home model (Lists set)
//	-> FAVORITES tab level 1 (the lists) -> open one -> the work is there
//	-> enter searches its clean title.
//
// It needs the store (internal/favorites), the picker (favorite_picker.go,
// the newFavoriteLists seam) and the lists level of FAVORITES (home_lists.go)
// all merged. Run: go test -tags reallists -run RealCascade ./internal/tui
//
// The key presses of the two new screens are not part of the written
// contract, so every guess about them lives in the helpers at the bottom
// (addViaPicker, createViaPicker, openListOnHome). Adjust those if a screen
// uses other keys; the assertions about what ends up on disk and on the
// screen do not change.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// productionFavoriteLists is the seam's real default, kept before TestMain
// switches the saved lists off for the tests.
var productionFavoriteLists = newFavoriteLists

var (
	cascCtrlL = tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl}
	cascSpace = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	cascEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	cascDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	cascTab   = tea.KeyPressMsg{Code: tea.KeyTab}
)

// cascadeWorld is one temp favorites file, the lists over it, and the results
// screen wired to it through the seam.
type cascadeWorld struct {
	store *favorites.Store
	path  string
	lists FavoriteLists
}

func newCascadeWorld(t *testing.T) *cascadeWorld {
	t.Helper()
	path := filepath.Join(t.TempDir(), "favorites.json")
	w := &cascadeWorld{store: favorites.Open(path), path: path}
	w.lists = StoreLists(w.store)
	prev := newFavoriteLists
	newFavoriteLists = func() FavoriteLists { return w.lists }
	prevFavs := newFavorites
	newFavorites = func() favoritesAPI { return w.store }
	t.Cleanup(func() {
		newFavoriteLists = prev
		newFavorites = prevFavs
	})
	return w
}

func (w *cascadeWorld) results() *animeResultsModel {
	return newAnimeResultsModel([]*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[PT-BR] Naruto", Source: "Animefire.io", Year: "2002", MediaType: models.MediaTypeAnime},
	})
}

// home is a fresh home model over what is on disk now, reopened the way the
// handler does it: a new Lists over the same file.
func (w *cascadeWorld) home() *searchHomeModel {
	return newSearchHomeModel(SearchHomeOptions{Lists: StoreLists(favorites.Open(w.path))})
}

func cascView(m interface{ View() tea.View }) string { return ansi.Strip(m.View().Content) }

func (w *cascadeWorld) summary(t *testing.T, name string) favorites.Summary {
	t.Helper()
	sums, err := favorites.Open(w.path).Summaries()
	require.NoError(t, err)
	for _, s := range sums {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("no list named %q on disk: %+v", name, sums)
	return favorites.Summary{}
}

func TestRealCascadeResultsPickerDiskHomeListOpenSearch(t *testing.T) {
	w := newCascadeWorld(t)
	made, err := w.lists.Create("Cozy nights")
	require.NoError(t, err)

	// 1 results screen: ctrl+l puts the highlighted work in the list.
	require.True(t, addViaPicker(t, w, "frieren", made.ID), "the picker adds the highlighted work to the list")

	// 2 the file on disk has it.
	disk := favorites.Open(w.path)
	entries, err := disk.Entries(made.ID)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "frieren", entries[0].Key)
	assert.Equal(t, "Frieren", entries[0].Title)
	assert.Equal(t, "2023", entries[0].Year)
	assert.Equal(t, "Anime", entries[0].Media)
	assert.False(t, disk.Has("frieren"), "being in a list is not being in Favorites")

	// 3 the home model with Lists set: FAVORITES tab level 1 shows the list
	// with its count and date.
	home := w.home()
	toFavoritesTab(t, home)
	level1 := cascView(home)
	assert.Contains(t, level1, "Cozy nights")
	assert.Contains(t, level1, "1", "the list's count")
	assert.Contains(t, level1, "created "+w.summary(t, "Cozy nights").CreatedAt.Format(homeListDateLayout), "the list's date")

	// 4 open it: the work is there.
	home = openListOnHome(t, w, "Cozy nights", "Frieren")
	assert.Contains(t, cascView(home), "Frieren")

	// 5 enter on the work searches its clean title, in every scope.
	_, cmd := home.Update(cascEnter)
	require.NotNil(t, cmd, "enter leaves the screen")
	assert.Equal(t, HomeSearch, home.result.Action)
	assert.Equal(t, "Frieren", home.result.Query, "the clean title, no tags")
	assert.Empty(t, home.result.Scope)
}

func TestRealCascadeCreateAListFromThePickerAndAddTheWork(t *testing.T) {
	w := newCascadeWorld(t)

	require.True(t, createViaPicker(t, w, "Weekend"), "the picker can make a list")

	s := w.summary(t, "Weekend")
	assert.False(t, s.Auto)
	assert.WithinDuration(t, time.Now(), s.CreatedAt, time.Minute)
	if s.Count == 0 { // a screen that only creates: add the work as the next step
		require.True(t, addViaPicker(t, w, "frieren", s.ID))
	}
	assert.Equal(t, 1, w.summary(t, "Weekend").Count)

	home := w.home()
	toFavoritesTab(t, home)
	assert.Contains(t, cascView(home), "Weekend")
}

func TestRealCascadeDeleteRenameRemovePathsReachTheScreen(t *testing.T) {
	w := newCascadeWorld(t)
	made, err := w.lists.Create("Old name")
	require.NoError(t, err)
	for _, title := range []string{"Frieren", "Naruto"} {
		_, err = w.lists.ToggleWork(made.ID, HomeFavorite{Title: title, Year: "2020", Media: "Anime"})
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
	}

	// remove one work: level 2 no longer shows it.
	require.NoError(t, w.lists.RemoveWork(made.ID, "frieren"))
	home := openListOnHome(t, w, "Old name", "Naruto")
	assert.NotContains(t, cascView(home), "Frieren")

	// rename: level 1 shows the new name, not the old.
	require.NoError(t, w.lists.Rename(made.ID, "New name"))
	home = w.home()
	toFavoritesTab(t, home)
	level1 := cascView(home)
	assert.Contains(t, level1, "New name")
	assert.NotContains(t, level1, "Old name")

	// clear: the list stays, empty.
	require.NoError(t, w.lists.Clear(made.ID))
	assert.Equal(t, 0, w.summary(t, "New name").Count)

	// delete: it is gone from level 1.
	require.NoError(t, w.lists.Delete(made.ID))
	home = w.home()
	toFavoritesTab(t, home)
	assert.NotContains(t, cascView(home), "New name")
}

func TestRealCascadeFavoritesListStillHoldsWhatCtrlFAdds(t *testing.T) {
	w := newCascadeWorld(t)
	m := w.results()

	_, _ = m.Update(ctrlF)

	assert.True(t, favorites.Open(w.path).Has("frieren"), "ctrl+f still fills Favorites")
	home := w.home()
	toFavoritesTab(t, home)
	assert.Contains(t, cascView(home), "Favorites")
	home = openListOnHome(t, w, "Favorites", "Frieren")
	assert.Contains(t, cascView(home), "Frieren")
}

func TestRealCascadeTheResultsScreenDefaultIsTheRealStore(t *testing.T) {
	// The seam default, not the test's replacement.
	// The cascade tests above put the seam back when they finish.
	got := productionFavoriteLists()
	sl, ok := got.(*storeLists)
	require.True(t, ok, "the default is the store adapter, not nil and not a fake: %T", got)
	assert.Same(t, favorites.Default(), sl.store, "it reads and writes the user's own favorites file")
}

// ---- helpers: every guess about the picker's and the home's keys is here ----

// toFavoritesTab presses tab until the FAVORITES tab shows.
func toFavoritesTab(t *testing.T, m *searchHomeModel) {
	t.Helper()
	for i := 0; i < len(homeScopes) && !m.onFavorites(); i++ {
		_, _ = m.Update(cascTab)
	}
	require.True(t, m.onFavorites(), "tab reaches FAVORITES")
}

// addViaPicker opens the picker for the highlighted result with ctrl+l, moves
// to the list and toggles it, trying the usual ways a checklist is toggled.
// It reports whether the work with key ended up in the list.
func addViaPicker(t *testing.T, w *cascadeWorld, key, listID string) bool {
	t.Helper()
	lists, err := w.lists.Lists()
	require.NoError(t, err)
	idx := -1
	for i, l := range lists {
		if l.ID == listID {
			idx = i
		}
	}
	require.GreaterOrEqual(t, idx, 0, "the list exists")

	for _, keys := range [][]tea.KeyPressMsg{{cascSpace}, {cascEnter}, {cascSpace, cascEnter}} {
		_ = w.lists.RemoveWork(listID, key)
		m := w.results() // the first result is Frieren
		_, _ = m.Update(cascCtrlL)
		require.NotContains(t, cascView(m), "Could not", "the picker opens")
		for range idx + 1 { // the first row is "+ Create a new list"
			_, _ = m.Update(cascDown)
		}
		for _, k := range keys {
			_, _ = m.Update(k)
		}
		if w.lists.Membership(key)[listID] {
			return true
		}
	}
	return false
}

// createViaPicker makes a list from the picker: "n" for a new list, the name,
// enter. It reports whether a list with that name exists afterwards.
func createViaPicker(t *testing.T, w *cascadeWorld, name string) bool {
	t.Helper()
	m := w.results()
	_, _ = m.Update(cascCtrlL)
	_, _ = m.Update(cascEnter) // on "+ Create a new list", the first row
	for _, r := range name {
		_, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	_, _ = m.Update(cascEnter)
	lists, err := w.lists.Lists()
	require.NoError(t, err)
	for _, l := range lists {
		if l.Name == name {
			return true
		}
	}
	return false
}

// openListOnHome returns a fresh home model on the FAVORITES tab with a list
// opened (level 2), found by moving down until enter changes the screen to one
// that shows wantInView; it fails when no row does that.
func openListOnHome(t *testing.T, w *cascadeWorld, name, wantInView string) *searchHomeModel {
	t.Helper()
	for downs := 0; downs <= 8; downs++ {
		m := w.home()
		toFavoritesTab(t, m)
		before := cascView(m)
		for range downs {
			_, _ = m.Update(cascDown)
		}
		_, _ = m.Update(cascEnter)
		after := cascView(m)
		if m.result.Query == "" && after != before && strings.Contains(after, wantInView) {
			return m
		}
	}
	t.Fatalf("could not open the list %q on the FAVORITES tab (adjust openListOnHome)", name)
	return nil
}
