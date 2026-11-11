package favorites

// Regression net for the favorites store as it is TODAY (before lists live
// beside Favorites). Everything here freezes what the legacy single list does
// to favorites.json and what it answers: the exact bytes, every error text,
// the ordering rules, the file shapes it accepts and rejects, the name of the
// set-aside copy of a damaged file, the 500 cap and the permission modes.
//
// Where today's behavior is arguably wrong the assertion still freezes it and
// carries a "frozen:" comment. The next feature steps may change such an
// assertion ONLY on purpose.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func regStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "favorites.json")
	return Open(path), path
}

func regRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func regWrite(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func regDay(day int) time.Time { return time.Date(2024, 1, day, 12, 0, 0, 0, time.UTC) }

func regKeys(list []Entry) []string {
	out := make([]string, 0, len(list))
	for i := range list {
		out = append(out, list[i].Key)
	}
	return out
}

func regSkipPerms(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores permission bits")
	}
}

// ---- golden bytes --------------------------------------------------------------

func TestRegressionStoreFileBytes(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)

	// Missing file: nothing is created by reading.
	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "reading never creates the file")

	// 1. The key is normalized, the title is kept verbatim, empty fields leave out.
	on, err := s.Toggle(Entry{Key: "  Frieren ", Title: " Frieren ", AddedAt: regDay(1)})
	require.NoError(t, err)
	assert.True(t, on)
	// frozen: no trailing newline after the closing brace.
	assert.Equal(t, "{\n"+
		"  \"version\": 1,\n"+
		"  \"entries\": [\n"+
		"    {\n"+
		"      \"key\": \"frieren\",\n"+
		"      \"title\": \" Frieren \",\n"+
		"      \"added_at\": \"2024-01-01T12:00:00Z\"\n"+
		"    }\n"+
		"  ]\n"+
		"}", regRead(t, path))

	// 2. Every field set, a zone offset and nanoseconds.
	zone := time.FixedZone("X", 3600)
	on, err = s.Toggle(Entry{Key: "Naruto", Title: "Naruto", Year: "2002", Media: "Anime", Source: "AnimeFire",
		AddedAt: time.Date(2024, 1, 2, 3, 4, 5, 123456789, zone)})
	require.NoError(t, err)
	assert.True(t, on)
	assert.Equal(t, "{\n"+
		"  \"version\": 1,\n"+
		"  \"entries\": [\n"+
		"    {\n"+
		"      \"key\": \"frieren\",\n"+
		"      \"title\": \" Frieren \",\n"+
		"      \"added_at\": \"2024-01-01T12:00:00Z\"\n"+
		"    },\n"+
		"    {\n"+
		"      \"key\": \"naruto\",\n"+
		"      \"title\": \"Naruto\",\n"+
		"      \"year\": \"2002\",\n"+
		"      \"media\": \"Anime\",\n"+
		"      \"source\": \"AnimeFire\",\n"+
		"      \"added_at\": \"2024-01-02T03:04:05.123456789+01:00\"\n"+
		"    }\n"+
		"  ]\n"+
		"}", regRead(t, path))

	// 3. A third one, then the first is toggled off: the file keeps the order
	// the works were added in (it is List that sorts, not the file).
	_, err = s.Toggle(Entry{Key: "Movie One", Title: "Movie One", Year: "1999", Media: "Movie", AddedAt: regDay(3)})
	require.NoError(t, err)
	on, err = s.Toggle(Entry{Key: "FRIEREN"})
	require.NoError(t, err)
	assert.False(t, on, "the same key takes it off, whatever the case and the other fields")
	assert.Equal(t, "{\n"+
		"  \"version\": 1,\n"+
		"  \"entries\": [\n"+
		"    {\n"+
		"      \"key\": \"naruto\",\n"+
		"      \"title\": \"Naruto\",\n"+
		"      \"year\": \"2002\",\n"+
		"      \"media\": \"Anime\",\n"+
		"      \"source\": \"AnimeFire\",\n"+
		"      \"added_at\": \"2024-01-02T03:04:05.123456789+01:00\"\n"+
		"    },\n"+
		"    {\n"+
		"      \"key\": \"movie one\",\n"+
		"      \"title\": \"Movie One\",\n"+
		"      \"year\": \"1999\",\n"+
		"      \"media\": \"Movie\",\n"+
		"      \"added_at\": \"2024-01-03T12:00:00Z\"\n"+
		"    }\n"+
		"  ]\n"+
		"}", regRead(t, path))

	// 4. Remove by a differently spelled key.
	require.NoError(t, s.Remove("  NARUTO"))
	assert.Equal(t, "{\n"+
		"  \"version\": 1,\n"+
		"  \"entries\": [\n"+
		"    {\n"+
		"      \"key\": \"movie one\",\n"+
		"      \"title\": \"Movie One\",\n"+
		"      \"year\": \"1999\",\n"+
		"      \"media\": \"Movie\",\n"+
		"      \"added_at\": \"2024-01-03T12:00:00Z\"\n"+
		"    }\n"+
		"  ]\n"+
		"}", regRead(t, path))

	// 5. Removing the last one through Remove leaves an empty array...
	require.NoError(t, s.Remove("movie one"))
	assert.Equal(t, "{\n  \"version\": 1,\n  \"entries\": []\n}", regRead(t, path))

	// 6. ...while Clear writes null.
	_, err = s.Toggle(Entry{Key: "x", Title: "X", AddedAt: regDay(4)})
	require.NoError(t, err)
	require.NoError(t, s.Clear())
	// frozen: Clear saves a nil slice, so the file says "entries": null where
	// Remove-the-last says []. Both read back as an empty list.
	assert.Equal(t, "{\n  \"version\": 1,\n  \"entries\": null\n}", regRead(t, path))
	list, err = s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRegressionStoreClearCreatesTheFileAndItsDirectory(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "a", "b")
	path := filepath.Join(dir, "favorites.json")
	s := Open(path)

	require.NoError(t, s.Clear())

	// frozen: clearing an absent file still writes one.
	assert.Equal(t, "{\n  \"version\": 1,\n  \"entries\": null\n}", regRead(t, path))
}

func TestRegressionStoreRemoveOfAnAbsentKeyWritesNothing(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)

	require.NoError(t, s.Remove("ghost"))
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "no file is created for a no-op remove")

	// A hand-written file in another layout stays byte for byte when nothing is removed.
	hand := `{"version":1,"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z"}]}`
	regWrite(t, path, hand)
	require.NoError(t, s.Remove("ghost"))
	assert.Equal(t, hand, regRead(t, path))

	// Removing a present key rewrites it in the canonical layout.
	require.NoError(t, s.Remove(" A"))
	assert.Equal(t, "{\n  \"version\": 1,\n  \"entries\": []\n}", regRead(t, path))
}

func TestRegressionStoreRewriteDropsUnknownFieldsAndRewritesKeys(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	regWrite(t, path, `{"version":1,"future":true,"entries":[{"key":"  Mixed Case ","title":"T","added_at":"2024-01-01T00:00:00Z","extra":1}]}`)

	_, err := s.Toggle(Entry{Key: "other", Title: "O", AddedAt: regDay(2)})
	require.NoError(t, err)

	// frozen: fields this build does not know are dropped on the next write
	// (only a newer *version* is protected), and a stored key is rewritten
	// normalized.
	assert.Equal(t, "{\n"+
		"  \"version\": 1,\n"+
		"  \"entries\": [\n"+
		"    {\n"+
		"      \"key\": \"mixed case\",\n"+
		"      \"title\": \"T\",\n"+
		"      \"added_at\": \"2024-01-01T00:00:00Z\"\n"+
		"    },\n"+
		"    {\n"+
		"      \"key\": \"other\",\n"+
		"      \"title\": \"O\",\n"+
		"      \"added_at\": \"2024-01-02T12:00:00Z\"\n"+
		"    }\n"+
		"  ]\n"+
		"}", regRead(t, path))
}

func TestRegressionStoreToggleStampsNowOnlyWhenNoTimeIsGiven(t *testing.T) {
	t.Parallel()
	s, _ := regStore(t)

	before := time.Now().Add(-time.Second)
	_, err := s.Toggle(Entry{Key: "stamped", Title: "S"})
	require.NoError(t, err)
	after := time.Now().Add(time.Second)
	_, err = s.Toggle(Entry{Key: "given", Title: "G", AddedAt: regDay(5)})
	require.NoError(t, err)

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 2)
	byKey := map[string]Entry{}
	for _, e := range list {
		byKey[e.Key] = e
	}
	assert.False(t, byKey["stamped"].AddedAt.Before(before) || byKey["stamped"].AddedAt.After(after), "stamped with the current time")
	assert.True(t, byKey["given"].AddedAt.Equal(regDay(5)), "a given time is kept")
}

func TestRegressionStoreRoundTripKeepsEveryField(t *testing.T) {
	t.Parallel()
	s, _ := regStore(t)
	zone := time.FixedZone("X", -3*3600)
	want := Entry{Key: "k", Title: "Title  Spaced", Year: "2001", Media: "TV", Source: "Pobreflix", AddedAt: time.Date(2024, 5, 6, 7, 8, 9, 10, zone)}
	_, err := s.Toggle(want)
	require.NoError(t, err)

	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	got := list[0]
	assert.Equal(t, "k", got.Key)
	assert.Equal(t, "Title  Spaced", got.Title, "the title is not normalized")
	assert.Equal(t, "2001", got.Year)
	assert.Equal(t, "TV", got.Media)
	assert.Equal(t, "Pobreflix", got.Source)
	assert.True(t, got.AddedAt.Equal(want.AddedAt))
}

// ---- error texts ----------------------------------------------------------------

func TestRegressionStoreErrorTexts(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "favorites: no data directory", errNoPath.Error())
	assert.Equal(t, "favorites: favorites.json was written by a newer GoAnime; update to change it", ErrNewerFile.Error())

	t.Run("no data directory on every call", func(t *testing.T) {
		t.Parallel()
		s := Open("")
		_, err := s.List()
		assert.Equal(t, "favorites: no data directory", errString(err))
		on, err := s.Toggle(Entry{Key: "k", Title: "K"})
		assert.False(t, on)
		assert.Equal(t, "favorites: no data directory", errString(err))
		assert.Equal(t, "favorites: no data directory", errString(s.Remove("k")))
		assert.Equal(t, "favorites: no data directory", errString(s.Clear()))
		assert.False(t, s.Has("k"), "any failure reads as no")
		keys := s.Keys()
		assert.NotNil(t, keys)
		assert.Empty(t, keys)
		assert.Empty(t, s.TakeDamaged())
	})

	t.Run("a work needs a key", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		for _, key := range []string{"", "   ", "\t\n"} {
			on, err := s.Toggle(Entry{Key: key, Title: "T"})
			assert.False(t, on)
			assert.Equal(t, "favorites: a work needs a key", errString(err))
		}
		_, statErr := os.Stat(path)
		assert.True(t, os.IsNotExist(statErr), "nothing is written for a refused work")
	})

	t.Run("a newer file is refused everywhere and left alone", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		newer := `{"version":2,"entries":[{"key":"future","title":"Future","added_at":"2024-01-01T00:00:00Z","new_field":1}]}`
		regWrite(t, path, newer)
		want := "favorites: favorites.json was written by a newer GoAnime; update to change it"

		_, err := s.List()
		assert.ErrorIs(t, err, ErrNewerFile)
		assert.Equal(t, want, errString(err))
		on, err := s.Toggle(Entry{Key: "k", Title: "K"})
		assert.False(t, on)
		assert.ErrorIs(t, err, ErrNewerFile)
		assert.ErrorIs(t, s.Remove("future"), ErrNewerFile)
		assert.ErrorIs(t, s.Clear(), ErrNewerFile)
		assert.False(t, s.Has("future"), "an unreadable list has nothing")
		assert.Empty(t, s.Keys())
		assert.Empty(t, s.TakeDamaged(), "a newer file is not a damaged one")
		assert.Equal(t, newer, regRead(t, path), "the file is byte for byte as it was")
		entries, err := os.ReadDir(filepath.Dir(path))
		require.NoError(t, err)
		assert.Len(t, entries, 1, "no copy was set aside, no temp file is left")
	})

	t.Run("the file is a directory", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "favorites.json")
		require.NoError(t, os.Mkdir(path, 0o700))
		s := Open(path)
		_, err := s.List()
		// frozen: the os error is cut to its operation and reason (the path
		// is noise on a screen).
		assert.Equal(t, "favorites: read: is a directory", errString(err))
		assert.False(t, s.Has("x"))
	})

	t.Run("cannot create the directory", func(t *testing.T) {
		regSkipPerms(t)
		t.Parallel()
		root := t.TempDir()
		require.NoError(t, os.Chmod(root, 0o500))
		t.Cleanup(func() { _ = os.Chmod(root, 0o700) })
		s := Open(filepath.Join(root, "sub", "favorites.json"))
		on, err := s.Toggle(Entry{Key: "k", Title: "K"})
		// frozen: a failed save still reports true ("it is a favorite
		// afterwards"); only the error says otherwise.
		assert.True(t, on)
		assert.Equal(t, "favorites: mkdir: permission denied", errString(err))
	})

	t.Run("cannot create the temp file", func(t *testing.T) {
		regSkipPerms(t)
		t.Parallel()
		root := t.TempDir()
		require.NoError(t, os.Chmod(root, 0o500))
		t.Cleanup(func() { _ = os.Chmod(root, 0o700) })
		s := Open(filepath.Join(root, "favorites.json"))
		on, err := s.Toggle(Entry{Key: "k", Title: "K"})
		assert.True(t, on, "frozen: true even though the save failed")
		assert.Equal(t, "favorites: open: permission denied", errString(err))
		assert.Equal(t, "favorites: open: permission denied", errString(s.Clear()))
	})

	t.Run("an unreadable file", func(t *testing.T) {
		regSkipPerms(t)
		t.Parallel()
		s, path := regStore(t)
		regWrite(t, path, `{"version":1,"entries":[]}`)
		require.NoError(t, os.Chmod(path, 0o000))
		t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
		_, err := s.List()
		assert.Equal(t, "favorites: open: permission denied", errString(err))
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ---- ordering ---------------------------------------------------------------------

func TestRegressionStoreListIsNewestFirstAndStableOnTies(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	regWrite(t, path, `{"version":1,"entries":[
	 {"key":"old","title":"Old","added_at":"2024-01-01T00:00:00Z"},
	 {"key":"tie-a","title":"Tie A","added_at":"2024-01-03T00:00:00Z"},
	 {"key":"newest","title":"Newest","added_at":"2024-01-05T00:00:00Z"},
	 {"key":"tie-b","title":"Tie B","added_at":"2024-01-03T00:00:00Z"},
	 {"key":"tie-c","title":"Tie C","added_at":"2024-01-03T00:00:00Z"}
	]}`)

	list, err := s.List()
	require.NoError(t, err)

	assert.Equal(t, []string{"newest", "tie-a", "tie-b", "tie-c", "old"}, regKeys(list),
		"newest first; equal dates keep the order of the file")
}

func TestRegressionStoreDedupeRules(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	// Position of the first appearance, content of the newest, empty keys dropped,
	// keys normalized before comparing.
	regWrite(t, path, `{"version":1,"entries":[
	 {"key":"Dup","title":"first","added_at":"2024-01-01T00:00:00Z"},
	 {"key":"  ","title":"no key","added_at":"2024-01-09T00:00:00Z"},
	 {"key":"solo","title":"Solo","added_at":"2024-01-02T00:00:00Z"},
	 {"key":" dup ","title":"newer","added_at":"2024-01-04T00:00:00Z"},
	 {"key":"DUP","title":"older","added_at":"2024-01-03T00:00:00Z"},
	 {"key":"","title":"empty","added_at":"2024-01-09T00:00:00Z"},
	 {"key":"same","title":"same-1","added_at":"2024-01-06T00:00:00Z"},
	 {"key":"same","title":"same-2","added_at":"2024-01-06T00:00:00Z"}
	]}`)

	list, err := s.List()
	require.NoError(t, err)

	got := make([]string, 0, len(list))
	for _, e := range list {
		got = append(got, e.Key+"="+e.Title)
	}
	// frozen: on an exact date tie the FIRST entry of the file wins.
	assert.Equal(t, []string{"same=same-1", "dup=newer", "solo=Solo"}, got)
	keys := s.Keys()
	assert.Equal(t, map[string]bool{"dup": true, "solo": true, "same": true}, keys)
}

func TestRegressionStoreHasAndKeysNormalize(t *testing.T) {
	t.Parallel()
	s, _ := regStore(t)
	assert.False(t, s.Has("anything"))
	assert.NotNil(t, s.Keys())
	assert.Empty(t, s.Keys())

	_, err := s.Toggle(Entry{Key: "Dungeon Meshi", Title: "Dungeon Meshi", AddedAt: regDay(1)})
	require.NoError(t, err)

	assert.True(t, s.Has("dungeon meshi"))
	assert.True(t, s.Has("  DUNGEON MESHI\t"))
	assert.False(t, s.Has("dungeon"))
	assert.False(t, s.Has(""))
	assert.Equal(t, map[string]bool{"dungeon meshi": true}, s.Keys())

	// Keys is a fresh map each time.
	keys := s.Keys()
	keys["mutated"] = true
	assert.Equal(t, map[string]bool{"dungeon meshi": true}, s.Keys())
}

func TestRegressionStoreToggleReportsTheStateAfterwards(t *testing.T) {
	t.Parallel()
	s, _ := regStore(t)
	on, err := s.Toggle(Entry{Key: "a", Title: "A", AddedAt: regDay(1)})
	require.NoError(t, err)
	assert.True(t, on)
	on, err = s.Toggle(Entry{Key: " A ", Title: "other title"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.False(t, s.Has("a"))
	on, err = s.Toggle(Entry{Key: "a", Title: "A again", AddedAt: regDay(2)})
	require.NoError(t, err)
	assert.True(t, on)
	list, err := s.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "A again", list[0].Title, "a work put back is a new entry")
}

// ---- v1 file shapes -----------------------------------------------------------------

func TestRegressionStoreV1FileShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
		keys []string // accepted: the keys List returns
		// aside is true when the file is damaged and set aside.
		aside bool
		newer bool
	}{
		{name: "empty object is version 0 and empty", body: `{}`},
		{name: "version 1 with entries", body: `{"version":1,"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z"}]}`, keys: []string{"a"}},
		{name: "version 1 with null entries", body: `{"version":1,"entries":null}`},
		{name: "version 1 without entries", body: `{"version":1}`},
		{name: "version 0 with entries is read", body: `{"version":0,"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z"}]}`, keys: []string{"a"}},
		{name: "negative version is read", body: `{"version":-5,"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z"}]}`, keys: []string{"a"}},
		{name: "unknown fields are ignored", body: `{"version":1,"zzz":[1,2],"entries":[{"key":"a","title":"A","added_at":"2024-01-01T00:00:00Z","what":1}]}`, keys: []string{"a"}},
		{name: "JSON null is an empty list", body: `null`},
		{name: "entry without a date is read", body: `{"version":1,"entries":[{"key":"a","title":"A"}]}`, keys: []string{"a"}},
		{name: "entry without a title is read", body: `{"version":1,"entries":[{"key":"a"}]}`, keys: []string{"a"}},
		{name: "version 2 is newer", body: `{"version":2,"entries":[]}`, newer: true},
		{name: "version 99 is newer", body: `{"version":99}`, newer: true},
		{name: "a newer version is judged after the JSON parses", body: `{"version":2,"entries":[{"key":"a"}`, aside: true},
		{name: "empty file is damaged", body: ``, aside: true},
		{name: "whitespace file is damaged", body: " \n", aside: true},
		{name: "truncated JSON is damaged", body: `{"version":1,"entries":[`, aside: true},
		{name: "top-level array is damaged", body: `[]`, aside: true},
		{name: "top-level string is damaged", body: `"favorites"`, aside: true},
		{name: "version as a string is damaged", body: `{"version":"1","entries":[]}`, aside: true},
		{name: "fractional version is damaged", body: `{"version":1.5,"entries":[]}`, aside: true},
		{name: "entries as an object is damaged", body: `{"version":1,"entries":{}}`, aside: true},
		{name: "entry as a string is damaged", body: `{"version":1,"entries":["a"]}`, aside: true},
		{name: "year as a number is damaged", body: `{"version":1,"entries":[{"key":"a","year":2020}]}`, aside: true},
		{name: "bad date is damaged", body: `{"version":1,"entries":[{"key":"a","added_at":"yesterday"}]}`, aside: true},
		{name: "trailing garbage is damaged", body: `{"version":1,"entries":[]} x`, aside: true},
		{name: "binary garbage is damaged", body: "\x00\x01\x02", aside: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, path := regStore(t)
			regWrite(t, path, tt.body)

			list, err := s.List()

			switch {
			case tt.newer:
				require.ErrorIs(t, err, ErrNewerFile)
				assert.Equal(t, tt.body, regRead(t, path))
			case tt.aside:
				require.NoError(t, err)
				assert.Empty(t, list)
				assert.NotEmpty(t, s.TakeDamaged())
				_, statErr := os.Stat(path)
				assert.True(t, os.IsNotExist(statErr), "the damaged file was moved away")
			default:
				require.NoError(t, err)
				assert.Equal(t, append([]string{}, tt.keys...), regKeys(list))
				assert.Empty(t, s.TakeDamaged())
				assert.Equal(t, tt.body, regRead(t, path), "reading never rewrites")
			}
		})
	}
}

// ---- damaged file ---------------------------------------------------------------------

var regAsideName = regexp.MustCompile(`^favorites\.json\.corrupt\.(\d{8}-\d{6})$`)

func TestRegressionStoreDamagedFileIsSetAside(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	regWrite(t, path, `{not json at all`)

	before := time.Now().Add(-2 * time.Second)
	list, err := s.List()
	after := time.Now().Add(2 * time.Second)

	require.NoError(t, err)
	assert.Empty(t, list, "a damaged file reads as empty")
	aside := s.TakeDamaged()
	require.NotEmpty(t, aside)
	assert.Equal(t, filepath.Dir(path), filepath.Dir(aside), "it stays beside the original")
	m := regAsideName.FindStringSubmatch(filepath.Base(aside))
	require.NotNil(t, m, "name is favorites.json.corrupt.YYYYMMDD-HHMMSS, got %q", filepath.Base(aside))
	stamp, perr := time.ParseInLocation("20060102-150405", m[1], time.Local)
	require.NoError(t, perr)
	assert.False(t, stamp.Before(before.Truncate(time.Second)) || stamp.After(after), "the stamp is the local time of the move")
	assert.Equal(t, `{not json at all`, regRead(t, aside), "the copy is the damaged file byte for byte")
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
	assert.Empty(t, s.TakeDamaged(), "it is told once")

	// The feature goes on with a clean file.
	on, err := s.Toggle(Entry{Key: "k", Title: "K", AddedAt: regDay(1)})
	require.NoError(t, err)
	assert.True(t, on)
	assert.Empty(t, s.TakeDamaged(), "the set-aside copy is not reported again")
	assert.Equal(t, `{not json at all`, regRead(t, aside))
}

func TestRegressionStoreEveryEntryPointSetsADamagedFileAside(t *testing.T) {
	t.Parallel()
	calls := map[string]func(s *Store){
		"List":   func(s *Store) { _, _ = s.List() },
		"Has":    func(s *Store) { _ = s.Has("a") },
		"Keys":   func(s *Store) { _ = s.Keys() },
		"Toggle": func(s *Store) { _, _ = s.Toggle(Entry{Key: "a", Title: "A", AddedAt: regDay(1)}) },
		"Remove": func(s *Store) { _ = s.Remove("a") },
		"Clear":  func(s *Store) { _ = s.Clear() },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s, path := regStore(t)
			regWrite(t, path, `oops`)
			call(s)
			aside := s.TakeDamaged()
			require.NotEmpty(t, aside, "%s reads the file, so it sets a damaged one aside", name)
			assert.Equal(t, `oops`, regRead(t, aside))
		})
	}
}

func TestRegressionStoreDamagedFileThatCannotBeSetAside(t *testing.T) {
	regSkipPerms(t)
	t.Parallel()
	s, path := regStore(t)
	regWrite(t, path, `broken`)
	dir := filepath.Dir(path)
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	_, err := s.List()

	require.Error(t, err)
	// Updated on purpose by the lists work: the rename's *LinkError used to
	// carry both paths onto the screen; like every other store error it now
	// says only what failed and why.
	assert.EqualError(t, err, "favorites: favorites.json is damaged and could not be set aside: rename: permission denied")
	assert.NotContains(t, err.Error(), path, "no file path on the screen")
	assert.Empty(t, s.TakeDamaged())
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr, "an unmovable damaged file is never overwritten")
	assert.False(t, s.Has("x"))
	on, terr := s.Toggle(Entry{Key: "x", Title: "X"})
	assert.False(t, on)
	assert.Error(t, terr)
	require.NoError(t, os.Chmod(dir, 0o700))
	assert.Equal(t, `broken`, regRead(t, path))
}

// ---- the 500 cap -------------------------------------------------------------------------

func regFill(t *testing.T, path string, n int) {
	t.Helper()
	entries := make([]Entry, 0, n)
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range n {
		entries = append(entries, Entry{Key: "k" + strconv.Itoa(i), Title: "T" + strconv.Itoa(i), AddedAt: base.Add(time.Duration(i) * time.Minute)})
	}
	data, err := json.MarshalIndent(file{Version: 1, Entries: entries}, "", "  ")
	require.NoError(t, err)
	regWrite(t, path, string(data))
}

func TestRegressionStoreCapIs500(t *testing.T) {
	t.Parallel()

	t.Run("exactly 500 keeps all, in the order added", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		regFill(t, path, 499)
		on, err := s.Toggle(Entry{Key: "last", Title: "Last", AddedAt: regDay(1)})
		require.NoError(t, err)
		assert.True(t, on)
		list, err := s.List()
		require.NoError(t, err)
		assert.Len(t, list, 500)
		assert.True(t, s.Has("k0"))
		// The file is in add order: the new one is last.
		var f file
		require.NoError(t, json.Unmarshal([]byte(regRead(t, path)), &f))
		assert.Equal(t, "k0", f.Entries[0].Key)
		assert.Equal(t, "last", f.Entries[499].Key)
	})

	t.Run("the 501st drops the oldest and the file becomes newest first", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		regFill(t, path, 500)
		on, err := s.Toggle(Entry{Key: "fresh", Title: "Fresh", AddedAt: regDay(1)})
		require.NoError(t, err)
		assert.True(t, on)
		list, err := s.List()
		require.NoError(t, err)
		assert.Len(t, list, 500)
		assert.False(t, s.Has("k0"), "the oldest went")
		assert.True(t, s.Has("k1"))
		assert.True(t, s.Has("fresh"))
		assert.Equal(t, "fresh", list[0].Key)
		var f file
		require.NoError(t, json.Unmarshal([]byte(regRead(t, path)), &f))
		require.Len(t, f.Entries, 500)
		// frozen: past the cap the file is rewritten newest first (sorted),
		// below it the file keeps the order the works were added in.
		assert.Equal(t, "fresh", f.Entries[0].Key)
		assert.Equal(t, "k499", f.Entries[1].Key)
		assert.Equal(t, "k1", f.Entries[499].Key)
	})

	t.Run("a work older than all 500 is not kept although Toggle says it is", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		regFill(t, path, 500)
		on, err := s.Toggle(Entry{Key: "ancient", Title: "Ancient", AddedAt: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)})
		require.NoError(t, err)
		// frozen: Toggle reports true although the entry was the one dropped.
		assert.True(t, on)
		assert.False(t, s.Has("ancient"))
		assert.True(t, s.Has("k0"))
	})

	t.Run("on a tie the newcomer is the one dropped", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		entries := make([]Entry, 0, 500)
		for i := range 500 {
			entries = append(entries, Entry{Key: "t" + strconv.Itoa(i), Title: "T", AddedAt: regDay(1)})
		}
		data, err := json.MarshalIndent(file{Version: 1, Entries: entries}, "", "  ")
		require.NoError(t, err)
		regWrite(t, path, string(data))
		on, err := s.Toggle(Entry{Key: "late", Title: "Late", AddedAt: regDay(1)})
		require.NoError(t, err)
		// frozen: same date, stable sort: the appended entry sorts last and goes.
		assert.True(t, on)
		assert.False(t, s.Has("late"))
		assert.True(t, s.Has("t499"))
	})

	t.Run("a hand-edited file over the cap is trimmed only by an add", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		regFill(t, path, 510)
		list, err := s.List()
		require.NoError(t, err)
		assert.Len(t, list, 510, "reading does not trim")
		require.NoError(t, s.Remove("k0"))
		list, err = s.List()
		require.NoError(t, err)
		assert.Len(t, list, 509, "removing does not trim")
		_, err = s.Toggle(Entry{Key: "fresh", Title: "F", AddedAt: regDay(1)})
		require.NoError(t, err)
		list, err = s.List()
		require.NoError(t, err)
		assert.Len(t, list, 500)
		assert.Equal(t, "fresh", list[0].Key)
	})

	t.Run("taking one off at the cap needs no trim", func(t *testing.T) {
		t.Parallel()
		s, path := regStore(t)
		regFill(t, path, 500)
		on, err := s.Toggle(Entry{Key: "k250"})
		require.NoError(t, err)
		assert.False(t, on)
		list, err := s.List()
		require.NoError(t, err)
		assert.Len(t, list, 499)
	})
}

// ---- permissions -----------------------------------------------------------------------------

func TestRegressionStoreModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes")
	}
	t.Parallel()

	t.Run("a new directory is 0700 and the file 0600", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		dir := filepath.Join(root, "x", "tracking")
		path := filepath.Join(dir, "favorites.json")
		s := Open(path)
		_, err := s.Toggle(Entry{Key: "k", Title: "K", AddedAt: regDay(1)})
		require.NoError(t, err)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
		parent, err := os.Stat(filepath.Join(root, "x"))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), parent.Mode().Perm(), "every directory it makes is 0700")
		info, err = os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, entries, 1, "no temp file is left")
		assert.Equal(t, "favorites.json", entries[0].Name())
	})

	t.Run("an existing directory keeps its mode and a 0644 file is replaced by a 0600 one", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.Chmod(dir, 0o755))
		path := filepath.Join(dir, "favorites.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"version":1,"entries":[]}`), 0o644))
		require.NoError(t, os.Chmod(path, 0o644))
		s := Open(path)
		_, err := s.Toggle(Entry{Key: "k", Title: "K", AddedAt: regDay(1)})
		require.NoError(t, err)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
		info, err = os.Stat(path)
		require.NoError(t, err)
		// frozen: a write tightens the file to 0600.
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("the set-aside copy keeps the damaged file's mode", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "favorites.json")
		require.NoError(t, os.WriteFile(path, []byte(`junk`), 0o640))
		require.NoError(t, os.Chmod(path, 0o640))
		s := Open(path)
		_, err := s.List()
		require.NoError(t, err)
		aside := s.TakeDamaged()
		require.NotEmpty(t, aside)
		info, err := os.Stat(aside)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	})
}

// ---- Default ---------------------------------------------------------------------------------

func TestRegressionStoreDefaultIsOneSharedStoreBesideTheHistory(t *testing.T) {
	t.Parallel()
	a, b := Default(), Default()
	assert.Same(t, a, b, "one shared store")
	if a.path != "" {
		assert.Equal(t, "favorites.json", filepath.Base(a.path))
		assert.Equal(t, "tracking", filepath.Base(filepath.Dir(a.path)))
	}
	assert.Equal(t, 500, maxEntries)
	assert.Equal(t, 1, schemaVersion)
}

func TestRegressionStoreEntryJSONFieldNames(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(Entry{Key: "k", Title: "t", Year: "y", Media: "m", Source: "s", AddedAt: regDay(1)})
	require.NoError(t, err)
	assert.Equal(t, `{"key":"k","title":"t","year":"y","media":"m","source":"s","added_at":"2024-01-01T12:00:00Z"}`, string(data))
	data, err = json.Marshal(Entry{Key: "k"})
	require.NoError(t, err)
	assert.Equal(t, `{"key":"k","title":"","added_at":"0001-01-01T00:00:00Z"}`, string(data), "year, media and source leave out when empty; the rest do not")
	data, err = json.Marshal(file{Version: 1})
	require.NoError(t, err)
	assert.Equal(t, `{"version":1,"entries":null}`, string(data))
	assert.False(t, errors.Is(errNoPath, ErrNewerFile))
	assert.True(t, strings.HasPrefix(errNoPath.Error(), "favorites: "))
}
