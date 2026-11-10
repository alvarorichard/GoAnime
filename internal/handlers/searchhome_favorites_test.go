package handlers

import (
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useFavorites(t *testing.T) *favorites.Store {
	t.Helper()
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	prev := favoritesStore
	favoritesStore = func() *favorites.Store { return store }
	t.Cleanup(func() { favoritesStore = prev })
	return store
}

func TestHomeFavoritesAreNewestFirstAndBounded(t *testing.T) {
	store := useFavorites(t)
	base := time.Now().Add(-time.Hour)
	for i := range maxHomeFavorites + 3 {
		_, err := store.Toggle(favorites.Entry{
			Key: "k" + strconv.Itoa(i), Title: "Title " + strconv.Itoa(i),
			Year: "2020", Media: "Anime", Source: "HiAnime",
			AddedAt: base.Add(time.Duration(i) * time.Minute),
		})
		require.NoError(t, err)
	}

	rows := homeFavorites()

	require.Len(t, rows, maxHomeFavorites)
	assert.Equal(t, "Title "+strconv.Itoa(maxHomeFavorites+2), rows[0].Title, "newest first")
	assert.Equal(t, "k"+strconv.Itoa(maxHomeFavorites+2), rows[0].Key)
	assert.Equal(t, "2020", rows[0].Year)
	assert.Equal(t, "HiAnime", rows[0].Source)
}

func TestHomeFavoritesSanitizeTitlesAndSkipBlankOnes(t *testing.T) {
	store := useFavorites(t)
	_, err := store.Toggle(favorites.Entry{Key: "a", Title: "Bad\x1b[31m\ntitle"})
	require.NoError(t, err)
	_, err = store.Toggle(favorites.Entry{Key: "b", Title: "  "})
	require.NoError(t, err)

	rows := homeFavorites()

	require.Len(t, rows, 1)
	assert.NotContains(t, rows[0].Title, "\x1b")
	assert.NotContains(t, rows[0].Title, "\n")
}

func TestHomeFavoritesWithoutAStoreAreEmpty(t *testing.T) {
	prev := favoritesStore
	favoritesStore = func() *favorites.Store { return favorites.Open("") }
	t.Cleanup(func() { favoritesStore = prev })
	assert.Empty(t, homeFavorites())
}

func TestForgetAndClearFavoritesReachTheStore(t *testing.T) {
	store := useFavorites(t)
	for _, k := range []string{"a", "b"} {
		_, err := store.Toggle(favorites.Entry{Key: k, Title: k})
		require.NoError(t, err)
	}
	require.NoError(t, forgetFavorite("a"))
	assert.False(t, store.Has("a"))
	assert.True(t, store.Has("b"))
	require.NoError(t, clearFavorites())
	assert.False(t, store.Has("b"))
}

func TestOpenSearchHomePassesFavoritesOn(t *testing.T) {
	store := useFavorites(t)
	_, err := store.Toggle(favorites.Entry{Key: "naruto", Title: "Naruto"})
	require.NoError(t, err)

	var gotRows int
	var hasRemove, hasClear bool
	_, err = openSearchHome("", func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		gotRows = len(o.Favorites)
		hasRemove, hasClear = o.DeleteFavorite != nil, o.ClearFavorites != nil
		return tui.SearchHomeResult{}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, gotRows)
	assert.True(t, hasRemove)
	assert.True(t, hasClear)
}
