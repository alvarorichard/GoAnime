package favorites

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that pin one behavior of the store each, written from a hand-run
// mutation pass over favorites.go: every test here fails on at least one
// realistic mutant of the code it covers.

func skipIfNoPermissions(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores permission bits")
	}
}

func readFileJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func keysOf(list []Entry) []string {
	out := make([]string, len(list))
	for i := range list {
		out[i] = list[i].Key
	}
	return out
}

// ---- keys ---------------------------------------------------------------------

func TestEveryEntryPointNormalizesTheKey(t *testing.T) {
	for _, raw := range []string{"Naruto", "  naruto", "naruto\t", " NARUTO ", "\nNaRuTo\r\n"} {
		t.Run(strings.TrimSpace(strconv.Quote(raw)), func(t *testing.T) {
			s := newStore(t)
			on, err := s.Toggle(Entry{Key: raw, Title: "Naruto"})
			require.NoError(t, err)
			require.True(t, on)

			list, err := s.List()
			require.NoError(t, err)
			require.Len(t, list, 1)
			assert.Equal(t, "naruto", list[0].Key, "stored lower-cased and trimmed")
			assert.Equal(t, "Naruto", list[0].Title, "the title is left as given")

			assert.True(t, s.Has(raw), "Has normalizes the way Toggle does")
			assert.True(t, s.Has("naruto"))

			on, err = s.Toggle(Entry{Key: raw})
			require.NoError(t, err)
			assert.False(t, on, "the same spelling toggles it off")
			assert.False(t, s.Has("naruto"))

			_, err = s.Toggle(Entry{Key: "naruto", Title: "Naruto"})
			require.NoError(t, err)
			require.NoError(t, s.Remove(raw), "Remove normalizes the way Toggle does")
			assert.False(t, s.Has("naruto"))
		})
	}
}

func TestDifferentSpellingsAreDifferentWorksOnlyWhenTheyDiffer(t *testing.T) {
	s := newStore(t)
	for _, k := range []string{"one piece", "onepiece", "one  piece"} {
		_, err := s.Toggle(Entry{Key: k, Title: k})
		require.NoError(t, err)
	}
	list, err := s.List()
	require.NoError(t, err)
	assert.Len(t, list, 3, "only case and the outer spaces are folded; the tui folds the rest")
}

func TestToggleWithABlankKeyWritesNothing(t *testing.T) {
	s := newStore(t)
	for _, k := range []string{"", " ", "\t\n", " "} { // NBSP is a space to TrimSpace
		on, err := s.Toggle(Entry{Key: k, Title: "x"})
		assert.Error(t, err, "key %q", k)
		assert.False(t, on)
	}
	_, statErr := os.Stat(s.path)
	assert.ErrorIs(t, statErr, os.ErrNotExist, "a refused toggle never creates the file")
}

// ---- Toggle / Remove / Clear persist ------------------------------------------

func TestToggleOnAndOffAreVisibleToAFreshStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	a, b := Open(path), Open(path)

	on, err := a.Toggle(Entry{Key: "x", Title: "X"})
	require.NoError(t, err)
	require.True(t, on)
	require.True(t, b.Has("x"), "adding is on disk, not in a cache")

	on, err = a.Toggle(Entry{Key: "x"})
	require.NoError(t, err)
	require.False(t, on)
	assert.False(t, b.Has("x"), "removing is on disk too")
	assert.False(t, Open(path).Has("x"))
}

func TestRemoveAndClearAreVisibleToAFreshStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	s := Open(path)
	for _, k := range []string{"a", "b", "c"} {
		_, err := s.Toggle(Entry{Key: k, Title: k})
		require.NoError(t, err)
	}

	require.NoError(t, s.Remove("b"))
	list, err := Open(path).List()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "c"}, keysOf(list), "Remove reached the disk and took only b")

	require.NoError(t, s.Clear())
	list, err = Open(path).List()
	require.NoError(t, err)
	assert.Empty(t, list, "Clear reached the disk")
	assert.False(t, Open(path).Has("a"))
}

func TestRemoveFromTheMiddleKeepsTheRestInOrder(t *testing.T) {
	s := newStore(t)
	base := time.Now().Add(-time.Hour)
	for i, k := range []string{"a", "b", "c", "d", "e"} {
		_, err := s.Toggle(Entry{Key: k, Title: k, AddedAt: base.Add(time.Duration(i) * time.Minute)})
		require.NoError(t, err)
	}
	// toggle off the middle one, and the last one, and the first one
	for _, k := range []string{"c", "e", "a"} {
		on, err := s.Toggle(Entry{Key: k})
		require.NoError(t, err)
		require.False(t, on)
	}
	list, err := s.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"d", "b"}, keysOf(list), "exactly the toggled ones went, the others keep their dates")
}

func TestRemoveOfAMissingKeyDoesNotTouchTheDisk(t *testing.T) {
	s := newStore(t)
	require.NoError(t, s.Remove("ghost"))
	_, err := os.Stat(s.path)
	assert.ErrorIs(t, err, os.ErrNotExist, "no file is created for a no-op")

	_, err = s.Toggle(Entry{Key: "real", Title: "Real"})
	require.NoError(t, err)
	before, err := os.Stat(s.path)
	require.NoError(t, err)
	// make the file recognizable: a rewrite would replace this mtime
	old := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(s.path, old, old))

	require.NoError(t, s.Remove("ghost"))

	after, err := os.Stat(s.path)
	require.NoError(t, err)
	assert.True(t, after.ModTime().Equal(old), "a no-op Remove does not rewrite the file (was %v)", before.ModTime())
}

func TestRemoveDropsEveryCopyOfAKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	// A hand-edited file may carry the same key twice.
	writeRaw(t, path, `{"version":1,"entries":[
		{"key":"dup","title":"A","added_at":"2024-01-01T00:00:00Z"},
		{"key":"keep","title":"K","added_at":"2024-01-02T00:00:00Z"},
		{"key":"dup","title":"B","added_at":"2024-01-03T00:00:00Z"}]}`)
	s := Open(path)
	require.NoError(t, s.Remove("DUP"))
	list, err := s.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"keep"}, keysOf(list))
}

func TestToggleOffWithDuplicateKeysInTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	writeRaw(t, path, `{"version":1,"entries":[
		{"key":"dup","title":"A","added_at":"2024-01-01T00:00:00Z"},
		{"key":"dup","title":"B","added_at":"2024-01-03T00:00:00Z"}]}`)
	s := Open(path)
	on, err := s.Toggle(Entry{Key: "dup"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.False(t, s.Has("dup"), "toggled off means gone")
}

func writeRaw(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

// ---- AddedAt -----------------------------------------------------------------

func TestAddedAtIsKeptWhenGivenAndStampedWhenNot(t *testing.T) {
	s := newStore(t)
	given := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
	_, err := s.Toggle(Entry{Key: "old", Title: "Old", AddedAt: given})
	require.NoError(t, err)

	before := time.Now().Add(-time.Second)
	_, err = s.Toggle(Entry{Key: "new", Title: "New"})
	require.NoError(t, err)
	after := time.Now().Add(time.Second)

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 2)
	byKey := map[string]Entry{}
	for _, e := range list {
		byKey[e.Key] = e
	}
	assert.True(t, byKey["old"].AddedAt.Equal(given), "a date the caller gave is not replaced: %v", byKey["old"].AddedAt)
	assert.True(t, byKey["new"].AddedAt.After(before) && byKey["new"].AddedAt.Before(after),
		"an undated entry is stamped with now: %v", byKey["new"].AddedAt)
	assert.Equal(t, []string{"new", "old"}, keysOf(list), "newest first")
}

// ---- ordering ----------------------------------------------------------------

func TestListOrderIsByDateNotByInsertion(t *testing.T) {
	s := newStore(t)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range []struct {
		k   string
		day int
	}{{"mid", 5}, {"newest", 9}, {"oldest", 1}, {"second", 7}} {
		_, err := s.Toggle(Entry{Key: p.k, Title: p.k, AddedAt: base.AddDate(0, 0, p.day)})
		require.NoError(t, err)
	}
	list, err := s.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"newest", "second", "mid", "oldest"}, keysOf(list))
}

func TestEqualDatesKeepTheirFileOrder(t *testing.T) {
	s := newStore(t)
	at := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var want []string
	for i := range 60 {
		k := "k" + strconv.Itoa(1000+i)
		want = append(want, k)
		_, err := s.Toggle(Entry{Key: k, Title: k, AddedAt: at})
		require.NoError(t, err)
	}
	list, err := s.List()
	require.NoError(t, err)
	assert.Equal(t, want, keysOf(list), "a stable sort: ties stay in the order they were added")
}

// ---- the cap -----------------------------------------------------------------

func fill(t *testing.T, s *Store, n int, base time.Time) {
	t.Helper()
	for i := range n {
		_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i), Title: "t", AddedAt: base.Add(time.Duration(i) * time.Second)})
		require.NoError(t, err)
	}
}

func TestTheCapIsExactlyMaxEntries(t *testing.T) {
	s := newStore(t)
	base := time.Now().Add(-24 * time.Hour)
	fill(t, s, maxEntries, base)

	list, err := s.List()
	require.NoError(t, err)
	assert.Len(t, list, maxEntries, "exactly the cap is kept whole")
	assert.True(t, s.Has("k0"), "nothing is dropped until the cap is passed")

	_, err = s.Toggle(Entry{Key: "one-more", Title: "t", AddedAt: base.Add(time.Hour)})
	require.NoError(t, err)
	list, err = s.List()
	require.NoError(t, err)
	assert.Len(t, list, maxEntries, "one past the cap drops exactly one")
	assert.False(t, s.Has("k0"), "the oldest went")
	assert.True(t, s.Has("k1"), "the second oldest stayed")
	assert.True(t, s.Has("one-more"))
	assert.Equal(t, "one-more", list[0].Key)
}

func TestTheCapDropsTheOldestByDateNotByPosition(t *testing.T) {
	s := newStore(t)
	base := time.Now().Add(-24 * time.Hour)
	// Insert the OLDEST one in the middle of the file and the rest newest-last.
	fill(t, s, maxEntries, base.Add(time.Hour))
	_, err := s.Toggle(Entry{Key: "ancient", Title: "t", AddedAt: base})
	require.NoError(t, err)
	assert.Len(t, mustList(t, s), maxEntries)
	assert.False(t, s.Has("ancient"), "the oldest by date is the one that goes, wherever it sits")
	assert.True(t, s.Has("k0"))
}

func mustList(t *testing.T, s *Store) []Entry {
	t.Helper()
	list, err := s.List()
	require.NoError(t, err)
	return list
}

// ---- reading the file --------------------------------------------------------

func TestTheFileFormatIsPinned(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	writeRaw(t, path, `{"version":1,"entries":[{"key":"frieren","title":"Frieren","year":"2023","media":"Anime","source":"AnimeFire","added_at":"2024-03-04T05:06:07Z"}]}`)
	s := Open(path)

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, Entry{
		Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "AnimeFire",
		AddedAt: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC),
	}, list[0], "files written by another version keep reading")

	// and what we write has the same field names and a version
	_, err = s.Toggle(Entry{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)
	doc := readFileJSON(t, path)
	assert.EqualValues(t, 1, doc["version"])
	entries, ok := doc["entries"].([]any)
	require.True(t, ok)
	require.Len(t, entries, 2)
	for _, e := range entries {
		m, ok := e.(map[string]any)
		require.True(t, ok)
		for _, field := range []string{"key", "title", "year", "media", "source", "added_at"} {
			assert.Contains(t, m, field)
		}
	}
}

func TestAnEmptyListIsWrittenAsAnArrayOrNullButReadsBack(t *testing.T) {
	s := newStore(t)
	_, err := s.Toggle(Entry{Key: "a", Title: "A"})
	require.NoError(t, err)
	require.NoError(t, s.Clear())
	doc := readFileJSON(t, s.path)
	assert.EqualValues(t, 1, doc["version"], "a cleared file is still a versioned file")
	assert.Contains(t, doc, "entries")
	assert.Empty(t, mustList(t, s))
	on, err := s.Toggle(Entry{Key: "b", Title: "B"})
	require.NoError(t, err)
	assert.True(t, on)
}

func TestClearOnAMissingFileLeavesAValidFile(t *testing.T) {
	s := newStore(t)
	require.NoError(t, s.Clear())
	_, err := os.Stat(s.path)
	require.NoError(t, err)
	assert.Empty(t, mustList(t, s))
}

func TestNoPathFailsEveryCall(t *testing.T) {
	s := Open("")
	assert.ErrorIs(t, s.Remove("a"), errNoPath)
	assert.ErrorIs(t, s.Clear(), errNoPath)
	on, err := s.Toggle(Entry{Key: "a"})
	assert.ErrorIs(t, err, errNoPath)
	assert.False(t, on, "a failed toggle does not claim the work is a favorite")
	_, err = s.List()
	assert.ErrorIs(t, err, errNoPath)
}

func TestAReadErrorIsAnErrorNotAnEmptyList(t *testing.T) {
	// A directory where the file should be: ReadFile fails with something that
	// is not "not found", which must not read as an empty list.
	dir := filepath.Join(t.TempDir(), "favorites.json")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "child"), 0o700))
	s := Open(dir)

	_, err := s.List()
	require.Error(t, err)
	assert.NotErrorIs(t, err, os.ErrNotExist)
	assert.False(t, s.Has("x"))

	on, err := s.Toggle(Entry{Key: "x", Title: "x"})
	assert.Error(t, err)
	assert.False(t, on, "no toggle happened")
	assert.Error(t, s.Remove("x"))
	assert.Error(t, s.Clear())
	_, statErr := os.Stat(filepath.Join(dir, "child"))
	assert.NoError(t, statErr, "nothing was damaged")
	_, statErr = os.Stat(dir + ".corrupt")
	assert.ErrorIs(t, statErr, os.ErrNotExist, "an unreadable file is not 'corrupt'")
}

func TestUnreadableFileIsNeverOverwritten(t *testing.T) {
	skipIfNoPermissions(t)
	path := filepath.Join(t.TempDir(), "favorites.json")
	s := Open(path)
	_, err := s.Toggle(Entry{Key: "keep", Title: "Keep"})
	require.NoError(t, err)
	original, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.Chmod(path, 0))
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err = s.List()
	assert.Error(t, err)
	on, err := s.Toggle(Entry{Key: "other", Title: "Other"})
	assert.Error(t, err, "cannot read, so cannot add")
	assert.False(t, on)
	assert.Error(t, s.Remove("keep"))
	assert.Error(t, s.Clear(), "Clear must not wipe what it could not read")

	require.NoError(t, os.Chmod(path, 0o600))
	now, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original, now, "the file is byte-identical")
	_, statErr := os.Stat(path + ".corrupt")
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestCorruptFileIsMovedAsideOnEveryEntryPoint(t *testing.T) {
	cases := map[string]func(*Store) error{
		"List":   func(s *Store) error { _, err := s.List(); return err },
		"Toggle": func(s *Store) error { _, err := s.Toggle(Entry{Key: "a", Title: "a"}); return err },
		"Remove": func(s *Store) error { return s.Remove("a") },
		"Clear":  func(s *Store) error { return s.Clear() },
		"Has":    func(s *Store) error { s.Has("a"); return nil },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "favorites.json")
			writeRaw(t, path, "\x00\x01 definitely not json")
			require.NoError(t, call(Open(path)), "a bad file is not an error")
			aside, err := filepath.Glob(path + ".corrupt.*")
			require.NoError(t, err)
			require.Len(t, aside, 1, "kept for inspection")
			bad, err := os.ReadFile(aside[0])
			require.NoError(t, err)
			assert.Equal(t, "\x00\x01 definitely not json", string(bad))
		})
	}
}

func TestWrongShapesCountAsCorrupt(t *testing.T) {
	for name, content := range map[string]string{
		"array":            `[]`,
		"entries a string": `{"entries":"x"}`,
		"entry a number":   `{"entries":[1]}`,
		"truncated":        `{"version":1,"entries":[{"key":"a"`,
		"empty file":       ``,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "favorites.json")
			writeRaw(t, path, content)
			s := Open(path)
			assert.Empty(t, mustList(t, s))
			on, err := s.Toggle(Entry{Key: "ok", Title: "ok"})
			require.NoError(t, err)
			assert.True(t, on)
			assert.True(t, Open(path).Has("ok"))
		})
	}
}

// ---- writing the file --------------------------------------------------------

func TestSaveLeavesOnlyTheFile(t *testing.T) {
	s := newStore(t)
	for i := range 10 {
		_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i), Title: "t"})
		require.NoError(t, err)
	}
	require.NoError(t, s.Remove("k3"))
	require.NoError(t, s.Clear())
	names := dirNames(t, filepath.Dir(s.path))
	assert.Equal(t, []string{"favorites.json"}, names, "no temp files are left behind")
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	require.NoError(t, err)
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	return out
}

func TestSaveCreatesTheDirectoryPrivately(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on windows")
	}
	s := Open(filepath.Join(t.TempDir(), "a", "b", "favorites.json"))
	_, err := s.Toggle(Entry{Key: "x", Title: "x"})
	require.NoError(t, err)
	for _, p := range []string{filepath.Dir(s.path), filepath.Dir(filepath.Dir(s.path))} {
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Zero(t, info.Mode().Perm()&0o077, "%s must not be open to group or others: %v", p, info.Mode().Perm())
	}
	info, err := os.Stat(s.path)
	require.NoError(t, err)
	assert.Zero(t, info.Mode().Perm()&0o077, "the file is private: %v", info.Mode().Perm())
}

func TestSaveIntoAFileInsteadOfADirectoryFails(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "plain-file")
	require.NoError(t, os.WriteFile(parent, []byte("x"), 0o600))
	s := Open(filepath.Join(parent, "favorites.json"))

	_, err := s.Toggle(Entry{Key: "x", Title: "x"})
	require.Error(t, err)
	assert.Error(t, s.Clear())
	assert.False(t, s.Has("x"))
}

func TestAReadOnlyDirectoryFailsCleanlyAndKeepsTheOldFile(t *testing.T) {
	skipIfNoPermissions(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	s := Open(path)
	_, err := s.Toggle(Entry{Key: "keep", Title: "Keep"})
	require.NoError(t, err)
	original, err := os.ReadFile(path)
	require.NoError(t, err)

	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	on, err := s.Toggle(Entry{Key: "new", Title: "New"})
	require.Error(t, err, "the temp file cannot be created")
	assert.Contains(t, err.Error(), "favorites:")
	_ = on
	assert.Error(t, s.Remove("keep"))
	assert.Error(t, s.Clear())
	assert.True(t, s.Has("keep"), "reading still works")

	require.NoError(t, os.Chmod(dir, 0o700))
	now, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original, now, "the old file is intact")
	assert.Equal(t, []string{"favorites.json"}, dirNames(t, dir), "no temp file left behind")

	on, err = s.Toggle(Entry{Key: "new", Title: "New"})
	require.NoError(t, err, "it recovers once the directory is writable")
	assert.True(t, on)
}

func TestAReadOnlyParentOfAMissingDirectoryFails(t *testing.T) {
	skipIfNoPermissions(t)
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0o500))
	t.Cleanup(func() { _ = os.Chmod(root, 0o700) })
	s := Open(filepath.Join(root, "sub", "favorites.json"))
	_, err := s.Toggle(Entry{Key: "x", Title: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mkdir", "the missing directory could not be created, and that is what is reported")
	assert.False(t, s.Has("x"))
}

// ---- odd but portable paths --------------------------------------------------

func TestPathsWithSpacesAndUnicodeWork(t *testing.T) {
	path := filepath.Join(t.TempDir(), "my data", "ração é ありがとう", "favorites.json")
	s := Open(path)
	_, err := s.Toggle(Entry{Key: "Fullmetal Alchemist: Brotherhood", Title: "Fullmetal Alchemist: Brotherhood"})
	require.NoError(t, err)
	assert.True(t, Open(path).Has("fullmetal alchemist: brotherhood"))
}

func TestUnicodeKeysFoldCase(t *testing.T) {
	s := newStore(t)
	_, err := s.Toggle(Entry{Key: "  ÀÇÃO  ", Title: "Ação"})
	require.NoError(t, err)
	assert.True(t, s.Has("àção"))
	assert.True(t, s.Has("ÀÇÃO"))
}

func TestARelativePathStillWorks(t *testing.T) {
	t.Chdir(t.TempDir())
	s := Open("favorites.json")
	_, err := s.Toggle(Entry{Key: "x", Title: "x"})
	require.NoError(t, err)
	assert.True(t, s.Has("x"))
}

// ---- concurrency -------------------------------------------------------------

func TestMixedConcurrentWritersLoseNothing(t *testing.T) {
	// One Store shared by many goroutines, mixing every writer. The final
	// state is fully determined, so a missing lock shows as a lost update.
	const n = 16
	for round := range 8 {
		s := newStore(t)
		for i := range n {
			_, err := s.Toggle(Entry{Key: "gone" + strconv.Itoa(i), Title: "t"})
			require.NoError(t, err)
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := range n {
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, err := s.Toggle(Entry{Key: "stay" + strconv.Itoa(i), Title: "t"})
				assert.NoError(t, err)
			}()
			go func() {
				defer wg.Done()
				<-start
				assert.NoError(t, s.Remove("gone"+strconv.Itoa(i)))
			}()
		}
		close(start)
		wg.Wait()
		list, err := s.List()
		require.NoError(t, err)
		require.Len(t, list, n, "round %d: every add landed and every remove did", round)
		for _, e := range list {
			require.True(t, strings.HasPrefix(e.Key, "stay"), "round %d: %q should have been removed", round, e.Key)
		}
	}
}

func TestConcurrentReadersSeeWholeFiles(t *testing.T) {
	s := newStore(t)
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				list, err := s.List()
				assert.NoError(t, err, "a reader never sees a half-written file")
				for _, e := range list {
					assert.NotEmpty(t, e.Key)
				}
				s.Has("k1")
			}
		}()
	}
	for i := range 60 {
		_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i%7), Title: "t"})
		require.NoError(t, err)
		if i%10 == 0 {
			require.NoError(t, s.Clear())
		}
	}
	close(stop)
	wg.Wait()
	_, statErr := os.Stat(s.path + ".corrupt")
	assert.ErrorIs(t, statErr, os.ErrNotExist, "no reader ever saw the file as corrupt")
}

// ---- Default -----------------------------------------------------------------

func TestDefaultLivesBesideTheWatchHistory(t *testing.T) {
	// Default is computed once per process, so this does not pick the home;
	// it checks the file's name and neighbours, not where the home is.
	s := Default()
	require.Same(t, s, Default(), "one shared store")
	require.NotEmpty(t, s.path)
	assert.Equal(t, "favorites.json", filepath.Base(s.path))
	assert.Equal(t, filepath.Base(filepath.Dir(s.path)), "tracking", "beside progress.db")
}

// With no home directory Default has no path, so every call fails clearly
// instead of writing into the working directory. It needs a process of its own:
// Default is computed once.
func TestDefaultWithoutAHomeHasNoPath(t *testing.T) {
	if os.Getenv("GOANIME_FAV_NOHOME") == "1" {
		s := Default()
		if s.path != "" {
			t.Fatalf("Default() has a path without a home directory: %q", s.path)
		}
		if _, err := s.Toggle(Entry{Key: "x"}); err == nil {
			t.Fatal("Toggle on a path-less default store must fail")
		}
		return
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("os.UserHomeDir falls back to other variables here")
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestDefaultWithoutAHomeHasNoPath$", "-test.count=1")
	cmd.Env = append(envWithout("HOME"), "HOME=", "GOANIME_FAV_NOHOME=1")
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	entries, err := os.ReadDir(cmd.Dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "nothing was written into the working directory")
}

func envWithout(name string) []string {
	var out []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, name+"=") {
			out = append(out, kv)
		}
	}
	return out
}
