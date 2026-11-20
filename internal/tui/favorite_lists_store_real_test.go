//go:build reallists

package tui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests need the real lists in internal/favorites. Run them with
// go test -tags reallists ./internal/tui

func realListsStore(t *testing.T) (*favorites.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "favorites.json")
	return favorites.Open(path), path
}

func TestRealListsConstantsAgree(t *testing.T) {
	assert.Equal(t, favorites.FavoritesID, FavoritesListID, "the screens and the store name the automatic list alike")
}

func TestRealListsFavoritesLeadsAndIsAutomatic(t *testing.T) {
	store, _ := realListsStore(t)
	lists := StoreLists(store)

	all, err := lists.Lists()

	require.NoError(t, err)
	require.NotEmpty(t, all)
	assert.Equal(t, FavoritesListID, all[0].ID)
	assert.True(t, all[0].Auto)
	assert.True(t, all[0].Created.IsZero())
	assert.Equal(t, 0, all[0].Count)
}

func TestRealListsLifeOfAListOnDisk(t *testing.T) {
	store, path := realListsStore(t)
	lists := StoreLists(store)

	made, err := lists.Create("  Weekend   watch ")
	require.NoError(t, err)
	assert.Equal(t, "Weekend watch", made.Name)
	assert.False(t, made.Auto)
	assert.WithinDuration(t, time.Now(), made.Created, time.Minute)
	assert.NotEqual(t, FavoritesListID, made.ID)

	on, err := lists.ToggleWork(made.ID, HomeFavorite{Title: "[PT-BR] Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)
	assert.True(t, on)
	time.Sleep(2 * time.Millisecond)
	_, err = lists.ToggleWork(made.ID, HomeFavorite{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime"})
	require.NoError(t, err)

	// A second handle on the same file sees the same lists: it is on disk.
	again := StoreLists(favorites.Open(path))
	all, err := again.Lists()
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Equal(t, made.ID, all[1].ID)
	assert.Equal(t, 2, all[1].Count)
	assert.Equal(t, []string{"Naruto", "[PT-BR] Frieren"}, all[1].Titles)
	assert.WithinDuration(t, made.Created, all[1].Created, time.Second)

	works, err := again.Works(made.ID)
	require.NoError(t, err)
	require.Len(t, works, 2)
	assert.Equal(t, HomeFavorite{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime"}, works[0], "newest first")
	assert.Equal(t, HomeFavorite{Key: "frieren", Title: "[PT-BR] Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"}, works[1],
		"a missing key was made from the title")

	assert.Equal(t, map[string]bool{made.ID: true}, again.Membership("frieren"))

	require.NoError(t, lists.Rename(made.ID, "Cozy"))
	require.NoError(t, lists.RemoveWork(made.ID, "frieren"))
	all, _ = again.Lists()
	assert.Equal(t, "Cozy", all[1].Name)
	assert.Equal(t, 1, all[1].Count)
	assert.Empty(t, again.Membership("frieren"))

	require.NoError(t, lists.Clear(made.ID))
	works, _ = again.Works(made.ID)
	assert.Empty(t, works)
	all, _ = again.Lists()
	assert.Len(t, all, 2, "clearing keeps the list")

	require.NoError(t, lists.Delete(made.ID))
	all, _ = again.Lists()
	assert.Len(t, all, 1)
	_, err = again.Works(made.ID)
	assert.ErrorIs(t, err, favorites.ErrNoSuchList)
}

func TestRealListsFavoritesListIsWhatCtrlFFills(t *testing.T) {
	store, _ := realListsStore(t)
	lists := StoreLists(store)

	on, err := lists.ToggleWork(FavoritesListID, HomeFavorite{Title: "Frieren", Year: "2023", Media: "Anime"})
	require.NoError(t, err)
	assert.True(t, on)

	assert.True(t, store.Has("frieren"), "the legacy API sees what the lists API added")
	assert.Equal(t, map[string]bool{FavoritesListID: true}, lists.Membership("frieren"))
	all, _ := lists.Lists()
	assert.Equal(t, 1, all[0].Count)

	_, err = store.Toggle(favorites.Entry{Key: "bleach", Title: "Bleach"})
	require.NoError(t, err)
	works, err := lists.Works(FavoritesListID)
	require.NoError(t, err)
	require.Len(t, works, 2)
	assert.Equal(t, "Bleach", works[0].Title, "newest first")
}

func TestRealListsErrorsAreThePhrasesTheScreensSoften(t *testing.T) {
	store, _ := realListsStore(t)
	lists := StoreLists(store)
	made, err := lists.Create("Mine")
	require.NoError(t, err)

	_, err = lists.Create("  ")
	require.ErrorIs(t, err, favorites.ErrNameEmpty)
	assert.Equal(t, "Give the list a name.", friendly(err))

	_, err = lists.Create("mine")
	require.ErrorIs(t, err, favorites.ErrNameTaken)
	assert.Equal(t, "You already have a list with that name.", friendly(err))

	_, err = lists.Create("Favorites")
	require.ErrorIs(t, err, favorites.ErrReservedName)

	require.ErrorIs(t, lists.Rename(FavoritesListID, "x"), favorites.ErrAutoList)
	require.ErrorIs(t, lists.Delete(FavoritesListID), favorites.ErrAutoList)
	require.ErrorIs(t, lists.Delete("nope"), favorites.ErrNoSuchList)
	_, err = lists.ToggleWork("nope", HomeFavorite{Title: "T"})
	require.ErrorIs(t, err, favorites.ErrNoSuchList)
	require.ErrorIs(t, lists.Rename(made.ID, ""), favorites.ErrNameEmpty)
}

func TestRealListsWorkWithoutAnyKeyOrTitleIsRefusedNotStored(t *testing.T) {
	store, _ := realListsStore(t)
	lists := StoreLists(store)
	made, err := lists.Create("Mine")
	require.NoError(t, err)

	_, err = lists.ToggleWork(made.ID, HomeFavorite{Title: "!!!"})

	require.Error(t, err, "a work with nothing to key it by is an error, not a silent blank row")
	works, _ := lists.Works(made.ID)
	assert.Empty(t, works)
}

func TestRealListsAreNotAvailableWithoutAFile(t *testing.T) {
	var none *favorites.Store
	_, err := StoreLists(none).Lists()
	assert.EqualError(t, err, "your lists are not available")
}
