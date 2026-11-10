package favorites

import (
	"os"
	"path/filepath"
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
	_, statErr := os.Stat(path + ".corrupt")
	assert.NoError(t, statErr, "the bad file is kept for inspection")

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
