//go:build reallists

package handlers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests need the real lists in internal/favorites. Run them with
// go test -tags reallists ./internal/handlers

func TestRealListsCascadeThroughTheHandlersOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	prev := favoritesStore
	favoritesStore = func() *favorites.Store { return favorites.Open(path) }
	t.Cleanup(func() { favoritesStore = prev })

	// First launch: the home screen's Lists make a list and fill it.
	lists := captureHomeOptions(t).Lists
	made, err := lists.Create("Weekend")
	require.NoError(t, err)
	_, err = lists.ToggleWork(made.ID, tui.HomeFavorite{Title: "[PT-BR] Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)
	_, err = lists.ToggleWork(tui.FavoritesListID, tui.HomeFavorite{Title: "Naruto", Year: "2002", Media: "Anime"})
	require.NoError(t, err)

	// The file on disk has both.
	disk := favorites.Open(path)
	entries, err := disk.Entries(made.ID)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "frieren", entries[0].Key)
	assert.True(t, disk.Has("naruto"))

	// Second launch: new options over the same file see the same lists, and
	// the legacy FAVORITES rows are what the automatic list holds.
	opts := captureHomeOptions(t)
	all, err := opts.Lists.Lists()
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Equal(t, tui.FavoritesListID, all[0].ID)
	assert.True(t, all[0].Auto)
	assert.Equal(t, 1, all[0].Count)
	assert.Equal(t, "Weekend", all[1].Name)
	assert.Equal(t, 1, all[1].Count)
	assert.WithinDuration(t, time.Now(), all[1].Created, time.Minute)
	require.Len(t, opts.Favorites, 1)
	assert.Equal(t, "Naruto", opts.Favorites[0].Title)

	works, err := opts.Lists.Works(made.ID)
	require.NoError(t, err)
	require.Len(t, works, 1)
	assert.Equal(t, "[PT-BR] Frieren", works[0].Title)
	assert.Equal(t, map[string]bool{made.ID: true}, opts.Lists.Membership("frieren"))

	// The legacy callbacks and the lists act on the same file.
	require.NoError(t, opts.DeleteFavorite("naruto"))
	all, _ = captureHomeOptions(t).Lists.Lists()
	assert.Equal(t, 0, all[0].Count)

	// Rename, remove, delete.
	require.NoError(t, opts.Lists.Rename(made.ID, "Cozy"))
	require.NoError(t, opts.Lists.RemoveWork(made.ID, "frieren"))
	all, _ = captureHomeOptions(t).Lists.Lists()
	assert.Equal(t, "Cozy", all[1].Name)
	assert.Equal(t, 0, all[1].Count)
	require.NoError(t, opts.Lists.Delete(made.ID))
	all, _ = captureHomeOptions(t).Lists.Lists()
	assert.Len(t, all, 1)
}

func TestRealListsTheHandlersStoreIsTheOneTheListsUse(t *testing.T) {
	// Two stores, two files: the lists follow favoritesStore(), not a default.
	store := useFavorites(t)
	made, err := captureHomeOptions(t).Lists.Create("Mine")
	require.NoError(t, err)

	sums, err := store.Summaries()
	require.NoError(t, err)
	var names []string
	for _, s := range sums {
		names = append(names, s.Name)
	}
	assert.Contains(t, names, "Mine")
	assert.NotEmpty(t, made.ID)
}
