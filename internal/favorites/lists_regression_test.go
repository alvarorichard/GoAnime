package favorites

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// preListsFile is exactly what the commit before lists wrote: MarshalIndent of
// {Version, Entries}, two spaces, no "lists" field.
const preListsFile = `{
  "version": 1,
  "entries": [
    {
      "key": "frieren",
      "title": "Frieren",
      "year": "2023",
      "media": "Anime",
      "source": "AnimeFire",
      "added_at": "2024-03-04T05:06:07Z"
    },
    {
      "key": "dune",
      "title": "Dune",
      "year": "2021",
      "media": "Movie",
      "added_at": "2024-03-05T05:06:07Z"
    }
  ]
}`

func rawField(t *testing.T, path, field string) string {
	t.Helper()
	var doc map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(readBytes(t, path), &doc))
	return string(doc[field])
}

// ---- (a) a file from before lists --------------------------------------------

func TestAPreListsFileLoadsAndEveryOldCallStillWorks(t *testing.T) {
	path := pathOf(t)
	writeRaw(t, path, preListsFile)
	s := Open(path)

	got := summaries(t, s)
	require.Len(t, got, 1, "no lists in an old file")
	assert.Equal(t, Summary{ID: FavoritesID, Name: "Favorites", Auto: true, Count: 2, Titles: []string{"Dune", "Frieren"}}, got[0])
	assert.Equal(t, []string{"dune", "frieren"}, keysOf(mustList(t, s)))
	assert.True(t, s.Has("Frieren"))
	assert.Equal(t, map[string]bool{"frieren": true, "dune": true}, s.Keys())
	assert.Equal(t, map[string]bool{FavoritesID: true}, s.Membership("dune"))
	assert.Empty(t, s.TakeDamaged())

	// the old operations keep working on it
	on, err := s.Toggle(Entry{Key: "naruto", Title: "Naruto", AddedAt: time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)
	assert.True(t, on)
	require.NoError(t, s.Remove("dune"))
	assert.Equal(t, []string{"naruto", "frieren"}, keysOf(mustList(t, s)))

	// and the file written by them has no "lists" at all: old builds read it
	doc := readFileJSON(t, path)
	assert.NotContains(t, doc, "lists")
	assert.EqualValues(t, 1, doc["version"])
}

func TestAfterCreatingAListTheOldFavoritesAreStillThere(t *testing.T) {
	path := pathOf(t)
	writeRaw(t, path, preListsFile)
	s := Open(path)
	sum := mkList(t, s, "Weekend")

	assert.Equal(t, []string{"dune", "frieren"}, keysOf(mustList(t, Open(path))))
	got := summaries(t, Open(path))
	require.Len(t, got, 2)
	assert.Equal(t, 2, got[0].Count)
	assert.Equal(t, sum.ID, got[1].ID)
	// the old entries are kept whole, field by field
	list := mustList(t, s)
	assert.Equal(t, Entry{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "AnimeFire", AddedAt: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)}, list[1])
	assert.Equal(t, Entry{Key: "dune", Title: "Dune", Year: "2021", Media: "Movie", AddedAt: time.Date(2024, 3, 5, 5, 6, 7, 0, time.UTC)}, list[0])
}

func TestAnOldFileWithoutAVersionStillLoadsAndIsWrittenWithOne(t *testing.T) {
	path := pathOf(t)
	writeRaw(t, path, `{"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z"}]}`)
	s := Open(path)
	require.Equal(t, []string{"a"}, keysOf(mustList(t, s)))
	mkList(t, s, "L")
	assert.EqualValues(t, 1, readFileJSON(t, path)["version"])
	assert.Equal(t, []string{"a"}, keysOf(mustList(t, s)))
}

// ---- (b) neither kind of change drops the other ------------------------------

func TestSavingFavoritesChangesKeepsTheListsAsTheyWere(t *testing.T) {
	s, path := tickStore(t)
	a := mkList(t, s, "Alpha")
	b := mkList(t, s, "Beta ünïcode")
	mustToggleIn(t, s, a.ID, "x")
	mustToggleIn(t, s, a.ID, "y")
	mustToggleIn(t, s, b.ID, "x")
	want := rawField(t, path, "lists")
	require.NotEqual(t, "null", want)

	check := func(label string) {
		t.Helper()
		assert.Equal(t, want, rawField(t, path, "lists"), "after %s the lists are byte for byte the same", label)
	}
	_, err := s.Toggle(Entry{Key: "f1", Title: "F1"})
	require.NoError(t, err)
	check("Toggle on")
	_, err = s.Toggle(Entry{Key: "x", Title: "X"}) // same key as in the lists
	require.NoError(t, err)
	check("Toggle on a key the lists have")
	_, err = s.Toggle(Entry{Key: "f1"})
	require.NoError(t, err)
	check("Toggle off")
	require.NoError(t, s.Remove("x"))
	check("Remove")
	require.NoError(t, s.Clear())
	check("Clear")
	require.NoError(t, s.RemoveFrom(FavoritesID, "nothing"))
	check("a Remove that changes nothing")

	// and a fresh store, as the next run, reads them the same
	got := summaries(t, Open(path))
	assert.Equal(t, []string{FavoritesID, a.ID, b.ID}, idsOf(got))
	assert.Equal(t, 2, got[1].Count)
	assert.Equal(t, 1, got[2].Count)
}

func TestSavingListChangesKeepsFavoritesAsTheyWere(t *testing.T) {
	s, path := tickStore(t)
	mustToggleIn(t, s, FavoritesID, "f1")
	mustToggleIn(t, s, FavoritesID, "f2")
	want := rawField(t, path, "entries")
	require.NotEqual(t, "null", want)
	check := func(label string) {
		t.Helper()
		assert.Equal(t, want, rawField(t, path, "entries"), "after %s Favorites are byte for byte the same", label)
	}

	a := mkList(t, s, "Alpha")
	check("CreateList")
	mustToggleIn(t, s, a.ID, "f1")
	check("ToggleIn")
	mustToggleIn(t, s, a.ID, "z")
	check("ToggleIn again")
	require.NoError(t, s.RenameList(a.ID, "Beta"))
	check("RenameList")
	require.NoError(t, s.RemoveFrom(a.ID, "f1"))
	check("RemoveFrom")
	require.NoError(t, s.ClearList(a.ID))
	check("ClearList")
	b := mkList(t, s, "Other")
	require.NoError(t, s.DeleteList(b.ID))
	check("DeleteList")
	require.NoError(t, s.DeleteList(a.ID))
	check("DeleteList of the last list")
	assert.NotContains(t, readFileJSON(t, path), "lists", "no lists: no field, the old shape")
}

// ---- (c) the new layout, pinned ----------------------------------------------

func TestTheNewFileLayoutIsPinned(t *testing.T) {
	s, path := tickStore(t)
	a := mkList(t, s, "Weekend marathon")
	b := mkList(t, s, "Empty one")
	require.Equal(t, "l"+strconv.FormatInt(t0.Add(time.Second).UnixNano(), 36), a.ID)
	_, err := s.ToggleIn(a.ID, Entry{Key: "dune", Title: "Dune", Year: "2021", Media: "Movie", Source: "StartFlix", AddedAt: time.Date(2024, 3, 5, 5, 6, 7, 0, time.UTC)})
	require.NoError(t, err)
	_, err = s.Toggle(Entry{Key: "frieren", Title: "Frieren", AddedAt: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)})
	require.NoError(t, err)
	_ = b

	want := `{
  "version": 1,
  "entries": [
    {
      "key": "frieren",
      "title": "Frieren",
      "added_at": "2024-03-04T05:06:07Z"
    }
  ],
  "lists": [
    {
      "id": "` + a.ID + `",
      "name": "Weekend marathon",
      "created_at": "2024-01-02T03:04:06Z",
      "entries": [
        {
          "key": "dune",
          "title": "Dune",
          "year": "2021",
          "media": "Movie",
          "source": "StartFlix",
          "added_at": "2024-03-05T05:06:07Z"
        }
      ]
    },
    {
      "id": "` + b.ID + `",
      "name": "Empty one",
      "created_at": "2024-01-02T03:04:07Z",
      "entries": []
    }
  ]
}`
	assert.Equal(t, want, string(readBytes(t, path)))
}

func TestTheNewLayoutReadsBackFromItsGolden(t *testing.T) {
	path := pathOf(t)
	writeRaw(t, path, `{
  "version": 1,
  "entries": [],
  "lists": [
    {"id": "lone", "name": "One", "created_at": "2024-01-02T03:04:06Z",
     "entries": [{"key": "dune", "title": "Dune", "added_at": "2024-03-05T05:06:07Z"}]}
  ]
}`)
	s := Open(path)
	got := summaries(t, s)
	require.Len(t, got, 2)
	assert.Equal(t, Summary{ID: "lone", Name: "One", CreatedAt: time.Date(2024, 1, 2, 3, 4, 6, 0, time.UTC), Count: 1, Titles: []string{"Dune"}}, got[1])
}

// ---- (d) a newer file is refused and never rewritten -------------------------

func TestANewerFileWithListsIsRefusedByEveryCallAndNeverRewritten(t *testing.T) {
	path := pathOf(t)
	const body = `{"version":2,"entries":[{"key":"a","title":"A"}],"lists":[{"id":"lx","name":"X","created_at":"2024-01-01T00:00:00Z","entries":[]}],"future":{"x":1}}`
	writeRaw(t, path, body)
	s := Open(path)

	_, err := s.Summaries()
	assert.ErrorIs(t, err, ErrNewerFile)
	_, err = s.CreateList("New")
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.ErrorIs(t, s.RenameList("lx", "Y"), ErrNewerFile)
	assert.ErrorIs(t, s.DeleteList("lx"), ErrNewerFile)
	_, err = s.Entries("lx")
	assert.ErrorIs(t, err, ErrNewerFile)
	_, err = s.Entries(FavoritesID)
	assert.ErrorIs(t, err, ErrNewerFile)
	_, err = s.ToggleIn("lx", Entry{Key: "k"})
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.ErrorIs(t, s.RemoveFrom("lx", "k"), ErrNewerFile)
	assert.ErrorIs(t, s.ClearList("lx"), ErrNewerFile)
	_, err = s.Toggle(Entry{Key: "k"})
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.ErrorIs(t, s.Remove("a"), ErrNewerFile)
	assert.ErrorIs(t, s.Clear(), ErrNewerFile)
	_, err = s.List()
	assert.ErrorIs(t, err, ErrNewerFile)
	assert.False(t, s.Has("a"))
	assert.Empty(t, s.Keys())
	assert.Empty(t, s.Membership("a"))
	assert.Empty(t, s.TakeDamaged(), "a newer file is not damaged")

	assert.Equal(t, body, string(readBytes(t, path)), "the file is untouched")
	assert.Equal(t, []string{"favorites.json"}, dirNames(t, filepath.Dir(path)), "and nothing was set aside")
}

// ---- (e) a damaged file is set aside, with lists too --------------------------

func TestADamagedFileIsSetAsideOnEveryListEntryPoint(t *testing.T) {
	calls := map[string]func(*Store) error{
		"Summaries":  func(s *Store) error { _, err := s.Summaries(); return err },
		"CreateList": func(s *Store) error { _, err := s.CreateList("A"); return err },
		"Rename":     func(s *Store) error { return s.RenameList("x", "y") },
		"Delete":     func(s *Store) error { return s.DeleteList("x") },
		"Entries":    func(s *Store) error { _, err := s.Entries(FavoritesID); return err },
		"ToggleIn":   func(s *Store) error { _, err := s.ToggleIn(FavoritesID, Entry{Key: "k"}); return err },
		"RemoveFrom": func(s *Store) error { return s.RemoveFrom(FavoritesID, "k") },
		"ClearList":  func(s *Store) error { return s.ClearList(FavoritesID) },
		"Membership": func(s *Store) error { s.Membership("k"); return nil },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			path := pathOf(t)
			const junk = `{"version":1,"lists":[{"id":"l1","name":`
			writeRaw(t, path, junk)
			s := Open(path)
			_ = call(s) // a missing list is fine here: the point is the file
			aside := s.TakeDamaged()
			require.NotEmpty(t, aside, "the screen can say where it went")
			assert.True(t, strings.HasPrefix(aside, path+".corrupt."))
			assert.Equal(t, junk, string(readBytes(t, aside)), "the bad file is kept whole")
			assert.Empty(t, s.TakeDamaged(), "said once")
			// and the feature works again
			_, err := s.CreateList("Fresh")
			require.NoError(t, err)
			got := summaries(t, s)
			assert.Equal(t, "Fresh", got[len(got)-1].Name)
			assert.Len(t, got, 2+btoi(name == "CreateList"), "CreateList's own list is there too")
		})
	}
}

func TestADamagedFileThatCannotBeSetAsideIsNotOverwrittenByLists(t *testing.T) {
	skipIfNoPermissions(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	writeRaw(t, path, `{not json`)
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	s := Open(path)
	_, err := s.CreateList("A")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "damaged")
	assert.NotContains(t, err.Error(), dir)
	require.NoError(t, os.Chmod(dir, 0o700))
	assert.Equal(t, `{not json`, string(readBytes(t, path)))
}

func TestAFileWithWrongListShapesCountsAsDamaged(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"lists":{"a":1}}`,
		`{"version":1,"lists":"x"}`,
		`{"version":1,"lists":[1,2]}`,
		`{"version":1,"lists":[{"id":5}]}`,
		`{"version":1,"lists":[{"entries":{"a":1}}]}`,
		`{"version":1,"lists":[{"created_at":"yesterday"}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			path := pathOf(t)
			writeRaw(t, path, body)
			s := Open(path)
			got := summaries(t, s)
			assert.Len(t, got, 1)
			assert.NotEmpty(t, s.TakeDamaged())
		})
	}
}

// ---- the old suite is untouched ----------------------------------------------

// The existing tests of this package (favorites_*_test.go) were not edited and
// still pass; this guards the one visible difference a refactor could cause:
// what Clear and an empty Favorites write.
func TestClearStillWritesTheOldShape(t *testing.T) {
	path := pathOf(t)
	s := Open(path)
	require.NoError(t, s.Clear())
	assert.JSONEq(t, `{"version":1,"entries":null}`, string(readBytes(t, path)))
	_, err := s.Toggle(Entry{Key: "a", AddedAt: t0})
	require.NoError(t, err)
	require.NoError(t, s.Remove("a"))
	assert.JSONEq(t, `{"version":1,"entries":[]}`, string(readBytes(t, path)))
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
