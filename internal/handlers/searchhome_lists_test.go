package handlers

import (
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheHomeIsHandedTheLists(t *testing.T) {
	useFavorites(t)

	opts := captureHomeOptions(t)

	require.NotNil(t, opts.Lists, "the FAVORITES tab needs its lists")
}

func TestTheHomeListsAreBackedByTheFavoritesStoreTheHandlerUses(t *testing.T) {
	// A store with a file answers; one with no data directory cannot read, so
	// its lists fail. That tells the handler's store from any other.
	useFavorites(t)
	_, err := captureHomeOptions(t).Lists.Lists()
	require.NoError(t, err, "a file backs the lists")

	prev := favoritesStore
	favoritesStore = func() *favorites.Store { return favorites.Open("") }
	t.Cleanup(func() { favoritesStore = prev })

	lists := captureHomeOptions(t).Lists
	require.NotNil(t, lists)
	_, err = lists.Lists()
	assert.Error(t, err, "no file, no lists")
	assert.NotPanics(t, func() { _ = lists.Membership("k") })
}

func TestTheListsAreReadAfterTheDamagedFileIsSetAside(t *testing.T) {
	// Reading a damaged file earns the notice once; handing over the lists
	// must not eat it or read the file first.
	path := t.TempDir() + "/favorites.json"
	require.NoError(t, os.WriteFile(path, []byte("{broken"), 0o600))
	prev := favoritesStore
	store := favorites.Open(path)
	favoritesStore = func() *favorites.Store { return store }
	t.Cleanup(func() { favoritesStore = prev })

	opts := captureHomeOptions(t)

	assert.Contains(t, opts.Notice, "damaged")
	assert.NotNil(t, opts.Lists)
}

func TestListsDoNotChangeTheLegacyFavoritesOptions(t *testing.T) {
	store := useFavorites(t)
	for _, k := range []string{"a", "b", "c"} {
		_, err := store.Toggle(favorites.Entry{Key: k, Title: "Title " + k, Year: "2020", Media: "Anime", Source: "HiAnime"})
		require.NoError(t, err)
	}

	opts := captureHomeOptions(t)

	require.Len(t, opts.Favorites, 3)
	for _, f := range opts.Favorites {
		assert.NotEmpty(t, f.Key)
		assert.Equal(t, "Anime", f.Media)
		assert.Equal(t, "HiAnime", f.Source)
		assert.Equal(t, "2020", f.Year)
	}
	require.NotNil(t, opts.DeleteFavorite)
	require.NotNil(t, opts.ClearFavorites)
	require.NoError(t, opts.DeleteFavorite("a"))
	assert.False(t, store.Has("a"))
	require.NoError(t, opts.ClearFavorites())
	assert.False(t, store.Has("b"))
	assert.Empty(t, opts.Notice)
	assert.NotNil(t, opts.Sources)
	assert.NotNil(t, opts.SetSource)
	assert.NotNil(t, opts.Suggest)
}

func TestTheOptionBlockKeepsItsAlignmentForTheSourcesGuard(t *testing.T) {
	// The sourcesguard mutation test deletes this exact line. If gofmt ever
	// realigns the block (a longer name in the first block), the mutant
	// stops applying and the guard goes quiet. Lists sits in the second block.
	src, err := os.ReadFile("searchhome.go")
	require.NoError(t, err)
	text := strings.ReplaceAll(string(src), "\r\n", "\n") // a Windows checkout may use CRLF

	assert.Contains(t, text, "\t\tSources:       homeSources,\n")
	assert.Contains(t, text, "\t\tSetSource:     setHomeSource,\n\n\t\tFavorites:", "the second block starts after a blank line")
	assert.Contains(t, text, "\t\tLists:          tui.StoreLists(favoritesStore()),\n", "Lists sits in the second block, aligned with ClearFavorites")
	first, second, ok := strings.Cut(text, "\t\tSetSource:")
	require.True(t, ok)
	assert.NotContains(t, first, "Lists:", "not in the block pinned by the guard")
	block, _, ok := strings.Cut(second, "\t})")
	require.True(t, ok)
	assert.Contains(t, block, "Lists:", "Lists is in the same block as Favorites, Notice and the legacy callbacks")
}
