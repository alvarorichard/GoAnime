package favorites

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// weirdDocuments are the structured seeds of FuzzListsFile: JSON that parses
// but is not what GoAnime writes. They are also run as a plain test.
func weirdDocuments() map[string]string {
	var many strings.Builder
	many.WriteString(`{"version":1,"lists":[`)
	for i := range 1000 {
		if i > 0 {
			many.WriteByte(',')
		}
		many.WriteString(`{"id":"l` + strconv.Itoa(i%300) + `","name":"n","entries":[{"key":"k` + strconv.Itoa(i) + `"}]}`)
	}
	many.WriteString(`]}`)
	return map[string]string{
		"empty lists":        `{"version":1,"entries":[],"lists":[]}`,
		"null lists":         `{"lists":null}`,
		"duplicate ids":      `{"version":1,"lists":[{"id":"a","name":"x"},{"id":"a","name":"y"},{"id":"a","name":"z"}]}`,
		"reserved id":        `{"version":1,"lists":[{"id":"favorites","name":"Favorites","entries":[{"key":"k"}]}]}`,
		"empty names":        `{"version":1,"lists":[{"id":"a","name":""},{"id":"b","name":"  \t "},{"id":"c"}]}`,
		"huge name":          `{"version":1,"lists":[{"id":"a","name":"` + strings.Repeat("ñ", 100000) + `"}]}`,
		"huge id":            `{"version":1,"lists":[{"id":"` + strings.Repeat("i", 100000) + `","name":"x"}]}`,
		"null entries":       `{"version":1,"lists":[{"id":"a","name":"A","entries":null},{"id":"b","name":"B","entries":[null,null,{}]}]}`,
		"null list":          `{"version":1,"lists":[null,null]}`,
		"control names":      `{"version":1,"lists":[{"id":"a","name":"\u0000\u0001\u001b[31m\u007f"},{"id":"b","name":"a\nb"}]}`,
		"bad utf8 name":      "{\"version\":1,\"lists\":[{\"id\":\"a\",\"name\":\"\xff\xfe\"}]}",
		"surrogates":         `{"version":1,"lists":[{"id":"a","name":"𐀀 \ud800"}]}`,
		"dates":              `{"version":1,"lists":[{"id":"a","name":"A","created_at":"9999-12-31T23:59:59Z"},{"id":"b","name":"B","created_at":"0001-01-01T00:00:00Z"}]}`,
		"same keys":          `{"version":1,"entries":[{"key":"k"}],"lists":[{"id":"a","name":"A","entries":[{"key":"k"},{"key":"K"},{"key":" k "}]}]}`,
		"1000 lists":         many.String(),
		"future version":     `{"version":2,"lists":[{"id":"a","name":"A"}]}`,
		"negative version":   `{"version":-1,"lists":[{"id":"a","name":"A"}]}`,
		"wrong shapes":       `{"version":1,"lists":{"a":1}}`,
		"id numbers":         `{"version":1,"lists":[{"id":5,"name":6}]}`,
		"rtl and zero width": `{"version":1,"lists":[{"id":"a","name":"\u202eevil\u200b"},{"id":"b","name":"Favor\u200bites"}]}`,
	}
}

func TestWeirdListDocumentsNeverBreakTheStore(t *testing.T) {
	for name, body := range weirdDocuments() {
		t.Run(name, func(t *testing.T) {
			path := pathOf(t)
			writeRaw(t, path, body)
			checkListsFileUsable(t, path)
		})
	}
}

// checkListsFileUsable is the property of FuzzListsFile: whatever the bytes in
// path are, no call panics and, unless the file is from a newer version, the
// loaded lists are sane and the store keeps working.
func checkListsFileUsable(t *testing.T, path string) {
	t.Helper()
	s := Open(path)
	sums, err := s.Summaries()
	if errors.Is(err, ErrNewerFile) {
		// refused: every call says so and the file is not touched
		before, _ := os.ReadFile(path)
		_, cerr := s.CreateList("fz")
		require.ErrorIs(t, cerr, ErrNewerFile)
		require.Empty(t, s.Membership("fz"))
		after, _ := os.ReadFile(path)
		require.Equal(t, string(before), string(after), "a newer file is never rewritten")
		return
	}
	require.NoError(t, err, "Summaries on a readable file must not fail")
	require.NotEmpty(t, sums)
	require.Equal(t, FavoritesID, sums[0].ID)
	require.True(t, sums[0].Auto)

	seen := map[string]bool{FavoritesID: true}
	for _, sum := range sums[1:] {
		require.False(t, sum.Auto)
		require.False(t, seen[sum.ID], "duplicate list ID %q", sum.ID)
		seen[sum.ID] = true
		require.NotEmpty(t, sum.ID)
		require.NotEmpty(t, sum.Name)
		require.LessOrEqual(t, utf8.RuneCountInString(sum.Name), maxListNameRunes)
		require.True(t, utf8.ValidString(sum.Name))
		for _, r := range sum.Name {
			require.False(t, unicode.IsControl(r), "control character in %q", sum.Name)
		}
		require.Len(t, sum.Titles, sum.Count)
		entries, eerr := s.Entries(sum.ID)
		require.NoError(t, eerr)
		keys := map[string]bool{}
		for _, e := range entries {
			require.NotEmpty(t, e.Key)
			require.Equal(t, normalize(e.Key), e.Key)
			require.False(t, keys[e.Key], "duplicate key %q in list %q", e.Key, sum.ID)
			keys[e.Key] = true
		}
		require.Len(t, entries, sum.Count)
	}
	for i := 2; i < len(sums); i++ {
		require.False(t, sums[i].CreatedAt.Before(sums[i-1].CreatedAt), "user lists are oldest first")
	}

	// Using the store changes nothing it had no reason to change: the same
	// lists, with the same IDs and names, are read back after a save.
	_, terr := s.ToggleIn(FavoritesID, Entry{Key: " FZ ", Title: "fuzz"})
	require.NoError(t, terr)
	again, err := Open(path).Summaries()
	require.NoError(t, err)
	require.Equal(t, idsOf(sums), idsOf(again))
	require.Equal(t, namesOf(sums), namesOf(again))
	require.True(t, Open(path).Has("fz"))
	require.Contains(t, Open(path).Membership("fz"), FavoritesID)

	// Every list can be worked on, then deleted, and the store is usable.
	for _, sum := range sums[1:] {
		_, terr := s.ToggleIn(sum.ID, Entry{Key: "fz-list"})
		require.NoError(t, terr)
		require.NoError(t, s.RemoveFrom(sum.ID, "fz-list"))
		require.NoError(t, s.ClearList(sum.ID))
		if len(sums) > 3 { // leave the big files whole after a few
			break
		}
	}
	for _, sum := range sums[1:] {
		require.NoError(t, s.DeleteList(sum.ID))
	}
	made, err := s.CreateList("fuzz made")
	require.NoError(t, err)
	require.NotEqual(t, FavoritesID, made.ID)
	require.ErrorIs(t, s.RenameList(FavoritesID, "x"), ErrAutoList)
	require.NoError(t, s.RenameList(made.ID, "fuzz renamed"))
	got, err := Open(path).Summaries()
	require.NoError(t, err)
	require.Equal(t, []string{FavoritesID, made.ID}, idsOf(got))
	assert.Equal(t, "fuzz renamed", got[1].Name)
}

// FuzzListsFile feeds arbitrary bytes in as favorites.json, with weird lists
// as seeds. Nothing may panic; the loaded lists are normalised (see
// normalizeDocument: duplicate, empty and reserved IDs are repaired, names are
// cleaned and cut, works are deduplicated) and the store stays usable.
func FuzzListsFile(f *testing.F) {
	for _, body := range weirdDocuments() {
		if len(body) < 20000 { // keep the corpus small; the big ones run as a plain test
			f.Add([]byte(body))
		}
	}
	for _, seed := range []string{``, `{}`, `null`, `[]`, "\x00\x01", `{"version":`, "\xef\xbb\xbf{}", preListsFile} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		path := pathOf(t)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Skip()
		}
		checkListsFileUsable(t, path)
	})
}
