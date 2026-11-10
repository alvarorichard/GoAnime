package handlers

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain keeps the whole package off the real FAVORITOS file: tests that open
// the search home without choosing a store would otherwise read (and, on a
// corrupt file, rename) the developer's own favorites.json.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "goanime-handlers-favs-")
	if err != nil {
		panic(err)
	}
	store := favorites.Open(filepath.Join(dir, "favorites.json"))
	favoritesStore = func() *favorites.Store { return store }
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func captureHomeOptions(t *testing.T) tui.SearchHomeOptions {
	t.Helper()
	var got tui.SearchHomeOptions
	_, err := openSearchHome("", func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		got = o
		return tui.SearchHomeResult{}, nil
	})
	require.NoError(t, err)
	return got
}

func TestTheHomeShowsExactlyEightFavorites(t *testing.T) {
	store := useFavorites(t)
	base := time.Now().Add(-time.Hour)
	for i := range 12 {
		_, err := store.Toggle(favorites.Entry{Key: "k" + strconv.Itoa(i), Title: "T" + strconv.Itoa(i), AddedAt: base.Add(time.Duration(i) * time.Minute)})
		require.NoError(t, err)
	}
	rows := homeFavorites()
	require.Len(t, rows, 8, "the home list is a short list: the 8 newest")
	assert.Equal(t, "k11", rows[0].Key)
	assert.Equal(t, "k4", rows[7].Key)
}

func TestFewerThanTheCapShowAll(t *testing.T) {
	store := useFavorites(t)
	for i := range maxHomeFavorites {
		_, err := store.Toggle(favorites.Entry{Key: "k" + strconv.Itoa(i), Title: "T" + strconv.Itoa(i)})
		require.NoError(t, err)
	}
	assert.Len(t, homeFavorites(), maxHomeFavorites, "exactly the cap is not cut")
}

func TestHomeFavoriteRowsCarryEveryField(t *testing.T) {
	store := useFavorites(t)
	_, err := store.Toggle(favorites.Entry{Key: " Frieren ", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)

	assert.Equal(t, []tui.HomeFavorite{{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"}}, homeFavorites())
}

func TestHomeFavoriteTitlesAreOneCleanLine(t *testing.T) {
	store := useFavorites(t)
	_, err := store.Toggle(favorites.Entry{Key: "a", Title: "Bad\x1b[31m\ntitle\t\x07  here"})
	require.NoError(t, err)

	rows := homeFavorites()

	require.Len(t, rows, 1)
	assert.Equal(t, "Bad title here", rows[0].Title)
	assert.Equal(t, "a", rows[0].Key, "the key is not touched: it is what removal uses")
}

func TestHomeFavoritesKeepTheirOrderWhenDatesTie(t *testing.T) {
	store := useFavorites(t)
	at := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, k := range []string{"first", "second", "third"} {
		_, err := store.Toggle(favorites.Entry{Key: k, Title: k, AddedAt: at})
		require.NoError(t, err)
	}
	var got []string
	for _, r := range homeFavorites() {
		got = append(got, r.Key)
	}
	assert.Equal(t, []string{"first", "second", "third"}, got)
}

func TestACorruptFavoritesFileOpensTheHomeWithoutFavorites(t *testing.T) {
	store := useFavorites(t)
	path := filepath.Join(t.TempDir(), "favorites.json")
	require.NoError(t, os.WriteFile(path, []byte("{broken"), 0o600))
	favoritesStore = func() *favorites.Store { return favorites.Open(path) }

	opts := captureHomeOptions(t)

	assert.Empty(t, opts.Favorites)
	assert.NotNil(t, opts.DeleteFavorite, "the actions are still wired")
	_, err := os.Stat(path + ".corrupt")
	assert.NoError(t, err, "kept for inspection")
	_ = store
}

func TestAnUnreadableFavoritesFileOpensTheHomeWithoutFavorites(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permission bits and a non-root user")
	}
	useFavorites(t)
	path := filepath.Join(t.TempDir(), "favorites.json")
	s := favorites.Open(path)
	_, err := s.Toggle(favorites.Entry{Key: "a", Title: "A"})
	require.NoError(t, err)
	require.NoError(t, os.Chmod(path, 0))
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	favoritesStore = func() *favorites.Store { return favorites.Open(path) }

	opts := captureHomeOptions(t)

	assert.Empty(t, opts.Favorites, "the home opens; the list is just empty")
	assert.Error(t, opts.DeleteFavorite("a"))
	assert.Error(t, opts.ClearFavorites())
}

func TestForgetAndClearReportTheStoresError(t *testing.T) {
	useFavorites(t)
	favoritesStore = func() *favorites.Store { return favorites.Open("") }
	assert.Error(t, forgetFavorite("a"))
	assert.Error(t, clearFavorites())
}

func TestForgetOnlyDropsThatKeyAndClearDropsAll(t *testing.T) {
	store := useFavorites(t)
	for _, k := range []string{"a", "b", "c"} {
		_, err := store.Toggle(favorites.Entry{Key: k, Title: k})
		require.NoError(t, err)
	}
	require.NoError(t, forgetFavorite("b"))
	assert.True(t, store.Has("a"))
	assert.False(t, store.Has("b"))
	assert.True(t, store.Has("c"))
	require.NoError(t, clearFavorites())
	assert.False(t, store.Has("a"))
	assert.False(t, store.Has("c"))
}

// The file the result screen writes (see internal/tui) is what the home reads.
// This is a literal copy of what ctrl+f produces, so the two packages cannot
// drift apart on field names without one of the suites noticing.
const fileWrittenByTheResultScreen = `{
  "version": 1,
  "entries": [
    {"key": "frieren", "title": "Frieren", "year": "2023", "media": "Anime", "source": "HiAnime", "added_at": "2024-03-04T05:06:07.123456789Z"},
    {"key": "a origem", "title": "A Origem", "year": "2010", "media": "Movie", "source": "StartFlix", "added_at": "2024-03-04T05:06:09Z"}
  ]
}`

func TestTheHomeReadsWhatTheResultScreenWrites(t *testing.T) {
	useFavorites(t)
	path := filepath.Join(t.TempDir(), "goanime", "favorites.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(fileWrittenByTheResultScreen), 0o600))
	favoritesStore = func() *favorites.Store { return favorites.Open(path) }

	opts := captureHomeOptions(t)

	assert.Equal(t, []tui.HomeFavorite{
		{Key: "a origem", Title: "A Origem", Year: "2010", Media: "Movie", Source: "StartFlix"},
		{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"},
	}, opts.Favorites, "newest first")

	// the actions the screen is handed act on that same file
	require.NoError(t, opts.DeleteFavorite(opts.Favorites[0].Key))
	assert.Equal(t, []string{"frieren"}, keysIn(t, path))
	require.NoError(t, opts.ClearFavorites())
	assert.Empty(t, keysIn(t, path))
	assert.Empty(t, captureHomeOptions(t).Favorites, "and the next home is empty")
}

func keysIn(t *testing.T, path string) []string {
	t.Helper()
	list, err := favorites.Open(path).List()
	require.NoError(t, err)
	var out []string
	for _, e := range list {
		out = append(out, e.Key)
	}
	return out
}

func TestTheHomeActionsFailWithAReadOnlyDataDirectory(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permission bits and a non-root user")
	}
	useFavorites(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	s := favorites.Open(path)
	_, err := s.Toggle(favorites.Entry{Key: "a", Title: "A"})
	require.NoError(t, err)
	favoritesStore = func() *favorites.Store { return favorites.Open(path) }
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	opts := captureHomeOptions(t)
	require.Len(t, opts.Favorites, 1, "reading works")
	assert.Error(t, opts.DeleteFavorite("a"), "the screen is told so it keeps the row")
	assert.Error(t, opts.ClearFavorites())

	require.NoError(t, os.Chmod(dir, 0o700))
	assert.True(t, favorites.Open(path).Has("a"), "nothing was lost")
}
