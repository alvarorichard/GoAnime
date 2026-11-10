package favorites

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return Open(filepath.Join(t.TempDir(), "sub", "favorites.json"))
}

func TestToggleAddsThenRemoves(t *testing.T) {
	s := newStore(t)
	on, err := s.Toggle(Entry{Key: " Naruto ", Title: "Naruto", Year: "2002", Media: "Anime", Source: "AnimeFire"})
	require.NoError(t, err)
	assert.True(t, on)
	assert.True(t, s.Has("naruto"), "keys are case and space insensitive")

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Naruto", list[0].Title)
	assert.Equal(t, "naruto", list[0].Key)
	assert.False(t, list[0].AddedAt.IsZero())

	on, err = s.Toggle(Entry{Key: "naruto"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.False(t, s.Has("naruto"))
}

func TestListIsNewestFirst(t *testing.T) {
	s := newStore(t)
	now := time.Now()
	for i, k := range []string{"a", "b", "c"} {
		_, err := s.Toggle(Entry{Key: k, Title: k, AddedAt: now.Add(time.Duration(i) * time.Hour)})
		require.NoError(t, err)
	}
	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, []string{"c", "b", "a"}, []string{list[0].Key, list[1].Key, list[2].Key})
}

func TestPersistsAcrossStores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	_, err := Open(path).Toggle(Entry{Key: "x", Title: "X"})
	require.NoError(t, err)
	assert.True(t, Open(path).Has("x"))
}

func TestToggleNeedsAKey(t *testing.T) {
	_, err := newStore(t).Toggle(Entry{Title: "no key", Key: "  "})
	assert.Error(t, err)
}

func TestRemoveAndClear(t *testing.T) {
	s := newStore(t)
	for _, k := range []string{"a", "b"} {
		_, err := s.Toggle(Entry{Key: k, Title: k})
		require.NoError(t, err)
	}
	require.NoError(t, s.Remove("A"))
	assert.False(t, s.Has("a"))
	assert.True(t, s.Has("b"))
	require.NoError(t, s.Remove("missing"), "removing what is not there is fine")

	require.NoError(t, s.Clear())
	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestCorruptFileIsMovedAsideNotFatal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))
	s := Open(path)

	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
	aside, globErr := filepath.Glob(path + ".corrupt.*")
	require.NoError(t, globErr)
	require.Len(t, aside, 1, "the bad file is kept for inspection")
	kept, readErr := os.ReadFile(aside[0])
	require.NoError(t, readErr)
	assert.Equal(t, "{not json", string(kept))

	on, err := s.Toggle(Entry{Key: "ok", Title: "Ok"})
	require.NoError(t, err)
	assert.True(t, on)
}

func TestNoPathFailsClearly(t *testing.T) {
	s := Open("")
	_, err := s.List()
	assert.ErrorIs(t, err, errNoPath)
	_, err = s.Toggle(Entry{Key: "a"})
	assert.ErrorIs(t, err, errNoPath)
	assert.False(t, s.Has("a"))
}

func TestCapDropsTheOldest(t *testing.T) {
	s := newStore(t)
	base := time.Now().Add(-24 * time.Hour)
	for i := range maxEntries + 3 {
		_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i), Title: "t", AddedAt: base.Add(time.Duration(i) * time.Second)})
		require.NoError(t, err)
	}
	list, err := s.List()
	require.NoError(t, err)
	assert.Len(t, list, maxEntries)
	assert.False(t, s.Has("k0"), "the oldest went")
	assert.True(t, s.Has("k"+strconv.Itoa(maxEntries+2)))
}

func TestConcurrentTogglesDoNotCorruptTheFile(t *testing.T) {
	s := newStore(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i), Title: "t"})
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	list, err := s.List()
	require.NoError(t, err)
	assert.Len(t, list, 20)
}

func TestANewerFileIsNeverRewritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	original := `{"version":2,"entries":[{"key":"a","title":"A","added_at":"2026-01-01T00:00:00Z","future":1}]}`
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))
	s := Open(path)

	_, err := s.Toggle(Entry{Key: "b", Title: "B"})
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.ErrorIs(t, s.Remove("a"), ErrNewerFile)
	assert.ErrorIs(t, s.Clear(), ErrNewerFile)
	_, err = s.List()
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.Empty(t, s.Keys())

	after, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, original, string(after), "nothing was lost")
}

func TestACorruptFileThatCannotBeSetAsideIsNotOverwritten(t *testing.T) {
	if os.Geteuid() == 0 || runtime.GOOS == "windows" {
		t.Skip("needs a read-only directory that is enforced")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	require.NoError(t, os.WriteFile(path, []byte("{broken but precious"), 0o600))
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	_, err := Open(path).Toggle(Entry{Key: "a", Title: "A"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "damaged")
	require.NoError(t, os.Chmod(dir, 0o700))
	kept, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "{broken but precious", string(kept))
}

func TestErrorsCarryNoPathOrTempName(t *testing.T) {
	if os.Geteuid() == 0 || runtime.GOOS == "windows" {
		t.Skip("needs a read-only directory that is enforced")
	}
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	_, err := Open(filepath.Join(dir, "favorites.json")).Toggle(Entry{Key: "a", Title: "A"})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), dir)
	assert.NotContains(t, err.Error(), ".tmp")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestDuplicateKeysInAHandEditedFileCollapse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	data := `{"version":1,"entries":[
	 {"key":"a","title":"old","added_at":"2026-01-01T00:00:00Z"},
	 {"key":" A ","title":"new","added_at":"2026-02-01T00:00:00Z"},
	 {"key":"","title":"nameless","added_at":"2026-03-01T00:00:00Z"}]}`
	require.NoError(t, os.WriteFile(path, []byte(data), 0o600))
	s := Open(path)

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "new", list[0].Title, "the newest copy wins")

	on, err := s.Toggle(Entry{Key: "a"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.False(t, s.Has("a"), "one toggle takes it off, not just one of the copies")
}

func TestKeysIsOneSnapshotOfEveryFavorite(t *testing.T) {
	s := newStore(t)
	for _, k := range []string{"a", "b"} {
		_, err := s.Toggle(Entry{Key: k, Title: k})
		require.NoError(t, err)
	}
	assert.Equal(t, map[string]bool{"a": true, "b": true}, s.Keys())
	assert.Empty(t, Open("").Keys())
}
