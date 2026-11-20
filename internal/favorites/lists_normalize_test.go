package favorites

import (
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// What a file read from disk is repaired to (see normalizeDocument). These
// tests pin the decisions: nothing is dropped but works without a key and
// duplicate keys; a list with a bad ID gets a fresh one.

func loadRaw(t *testing.T, body string) (*Store, []Summary) {
	t.Helper()
	path := pathOf(t)
	writeRaw(t, path, body)
	s := Open(path)
	return s, summaries(t, s)
}

func TestDuplicateListIDsKeepEveryListAndTheFirstKeepsTheID(t *testing.T) {
	s, got := loadRaw(t, `{"version":1,"lists":[
	  {"id":"lx","name":"First","created_at":"2024-01-01T00:00:00Z","entries":[{"key":"a","title":"A"}]},
	  {"id":"lx","name":"Second","created_at":"2024-01-02T00:00:00Z","entries":[{"key":"b","title":"B"}]},
	  {"id":"lx","name":"Third","created_at":"2024-01-03T00:00:00Z","entries":[]}]}`)
	assert.Equal(t, []string{FavoritesID, "lx", "r1", "r2"}, idsOf(got))
	assert.Equal(t, []string{"Favorites", "First", "Second", "Third"}, namesOf(got))
	assert.Equal(t, []string{"a"}, entriesOf(t, s, "lx"))
	assert.Equal(t, []string{"b"}, entriesOf(t, s, "r1"), "the works go with their list")
}

func TestAListWithTheReservedIDGetsAnotherOne(t *testing.T) {
	s, got := loadRaw(t, `{"version":1,"entries":[{"key":"fav","title":"F"}],"lists":[
	  {"id":"favorites","name":"Impostor","created_at":"2024-01-01T00:00:00Z","entries":[{"key":"mine","title":"M"}]}]}`)
	require.Len(t, got, 2)
	assert.Equal(t, FavoritesID, got[0].ID)
	assert.True(t, got[0].Auto)
	assert.Equal(t, 1, got[0].Count, "the automatic list is the file's own entries")
	assert.Equal(t, "r1", got[1].ID)
	assert.Equal(t, "Impostor", got[1].Name)
	assert.Equal(t, []string{"fav"}, entriesOf(t, s, FavoritesID))
	assert.Equal(t, []string{"mine"}, entriesOf(t, s, "r1"))
	assert.ErrorIs(t, s.DeleteList(FavoritesID), ErrAutoList)
	require.NoError(t, s.DeleteList("r1"))
	assert.Equal(t, []string{"fav"}, entriesOf(t, s, FavoritesID))
}

func TestEmptyAndMissingIDsAreRepaired(t *testing.T) {
	_, got := loadRaw(t, `{"version":1,"lists":[{"name":"No id"},{"id":"","name":"Empty id"},null,{"id":"keep","name":"Kept"}]}`)
	assert.Equal(t, []string{FavoritesID, "r1", "r2", "r3", "keep"}, idsOf(got))
	assert.Equal(t, []string{"Favorites", "No id", "Empty id", "Untitled list", "Kept"}, namesOf(got))
}

func TestRepairedIDsNeverCollideWithAnIDTheFileHas(t *testing.T) {
	_, got := loadRaw(t, `{"version":1,"lists":[
	  {"id":"x","name":"a"},{"id":"x","name":"b"},{"id":"r1","name":"c"},{"id":"r3","name":"d"},{"id":"","name":"e"}]}`)
	assert.Equal(t, []string{FavoritesID, "x", "r2", "r1", "r3", "r4"}, idsOf(got))
}

func TestRepairIsTheSameEveryTimeTheFileIsRead(t *testing.T) {
	path := pathOf(t)
	writeRaw(t, path, `{"version":1,"lists":[{"id":"d","name":"a"},{"id":"d","name":"b"},{"id":"favorites","name":"c"},{"name":"d"}]}`)
	first := summaries(t, Open(path))
	for range 3 {
		assert.Equal(t, first, summaries(t, Open(path)))
	}
	// once saved, the repaired IDs are the file's own and stay the same
	s := Open(path)
	mkList(t, s, "Fresh")
	after := summaries(t, Open(path))
	assert.Equal(t, idsOf(first), idsOf(after)[:len(first)])
}

func TestRepairedNamesAreCleanAndNeverEmptyOrTooLong(t *testing.T) {
	long := repeat("é", 100)
	_, got := loadRaw(t, `{"version":1,"lists":[
	  {"id":"a","name":""},
	  {"id":"b","name":"   "},
	  {"id":"c","name":"  two   spaces\tand\u0007bell "},
	  {"id":"d","name":"`+long+`"},
	  {"id":"e","name":"`+repeat("🎬", 39)+` `+repeat("x", 9)+`"},
	  {"id":"f"}]}`)
	assert.Equal(t, "Untitled list", got[1].Name)
	assert.Equal(t, "Untitled list", got[2].Name)
	assert.Equal(t, "two spaces andbell", got[3].Name)
	assert.Equal(t, repeat("é", 40), got[4].Name)
	assert.Equal(t, repeat("🎬", 39), got[5].Name, "a cut that ends on a space drops the space")
	assert.Equal(t, "Untitled list", got[6].Name)
	for _, sum := range got[1:] {
		assert.LessOrEqual(t, utf8.RuneCountInString(sum.Name), maxListNameRunes)
		assert.NotEmpty(t, sum.Name)
	}
}

func TestDuplicateAndReservedNamesInAFileAreLeftAlone(t *testing.T) {
	s, got := loadRaw(t, `{"version":1,"lists":[
	  {"id":"a","name":"Same","created_at":"2024-01-01T00:00:00Z"},
	  {"id":"b","name":"same","created_at":"2024-01-02T00:00:00Z"},
	  {"id":"c","name":"Favorites","created_at":"2024-01-03T00:00:00Z"}]}`)
	assert.Equal(t, []string{"Favorites", "Same", "same", "Favorites"}, namesOf(got))
	// the user can still repair them: the list's own name is allowed, a
	// different list's name is not
	assert.ErrorIs(t, s.RenameList("b", "SAME"), ErrNameTaken)
	require.NoError(t, s.RenameList("b", "Different"))
	assert.ErrorIs(t, s.RenameList("c", "favorites"), ErrReservedName)
	require.NoError(t, s.RenameList("c", "Other"))
}

func TestMoreThanFiftyListsInAFileAreAllKeptAndNoneCanBeAdded(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"version":1,"lists":[`)
	for i := range 1000 {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"id":"l` + strconv.Itoa(i) + `","name":"List ` + strconv.Itoa(i) + `","created_at":"2024-01-01T00:00:00Z","entries":[{"key":"k","title":"K"}]}`)
	}
	b.WriteString(`]}`)
	s, got := loadRaw(t, b.String())
	assert.Len(t, got, 1001)
	_, err := s.CreateList("one more")
	assert.ErrorIs(t, err, ErrTooManyLists)
	require.NoError(t, s.DeleteList("l5"))
	assert.Len(t, summaries(t, s), 1000, "deleting works, and nothing else is lost")
	assert.Len(t, s.Membership("k"), 999)
}

func TestListWorksAreDedupedAndKeyless(t *testing.T) {
	s, _ := loadRaw(t, `{"version":1,"lists":[{"id":"a","name":"A","entries":[
	  {"key":"x","title":"old","added_at":"2024-01-01T00:00:00Z"},
	  {"key":" X ","title":"new","added_at":"2024-02-01T00:00:00Z"},
	  {"key":"x","title":"older","added_at":"2023-01-01T00:00:00Z"},
	  {"key":"","title":"no key"},{"key":"  "},null,{"key":"y"}]}]}`)
	got := mustEntries(t, s, "a")
	assert.Equal(t, []string{"x", "y"}, sortedCopy(keysOf(got)))
	for _, e := range got {
		if e.Key == "x" {
			assert.Equal(t, "new", e.Title, "the newest copy wins")
		}
	}
	// the same key is not shared with Favorites
	assert.False(t, s.Has("x"))
}

func TestNullListsAndNullEntriesAreFine(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"lists":null}`, `{"version":1,"lists":[]}`, `{"version":1,"lists":[null]}`,
		`{"version":1,"lists":[{"id":"a","name":"A","entries":null}]}`,
		`{"version":1,"lists":[{"id":"a","name":"A","entries":[null,null]}]}`,
		`{"version":1,"entries":null,"lists":null}`,
	} {
		t.Run(body, func(t *testing.T) {
			s, _ := loadRaw(t, body)
			mkList(t, s, "Fresh")
			assert.True(t, mustToggleIn(t, s, FavoritesID, "k"))
		})
	}
}

func TestListsOrderByCreationEvenWhenTheFileIsOutOfOrder(t *testing.T) {
	_, got := loadRaw(t, `{"version":1,"lists":[
	  {"id":"c","name":"C","created_at":"2024-03-01T00:00:00Z"},
	  {"id":"nodate","name":"None"},
	  {"id":"a","name":"A","created_at":"2024-01-01T00:00:00Z"},
	  {"id":"b","name":"B","created_at":"2024-02-01T00:00:00Z"}]}`)
	assert.Equal(t, []string{FavoritesID, "nodate", "a", "b", "c"}, idsOf(got), "a list without a date is the oldest")
}

// ---- cleanName / foldName directly --------------------------------------------

func TestCleanNameNeverLeavesControlsOrOuterSpace(t *testing.T) {
	for _, in := range []string{"", " ", "a", " a ", "a\x00", "\x00a", "a\x00\x00 b", "\xff", "a\xffb", "\u0085a\u0085", "a\u2028b", "tab\tbed"} {
		out := cleanName(in)
		assert.True(t, utf8.ValidString(out), in)
		assert.Equal(t, strings.TrimSpace(out), out, in)
		assert.NotContains(t, out, "  ", in)
		for _, r := range out {
			assert.False(t, unicode.IsControl(r), "%q in %q", r, in)
		}
		assert.Equal(t, out, cleanName(out), "cleaning twice changes nothing")
	}
	assert.Equal(t, "a b", cleanName("a\u2028b"))
}

func TestFoldNameTable(t *testing.T) {
	for in, want := range map[string]string{
		"Maratón": "maraton", "ÀÉÎÕÜ": "aeiou", "Çedilha": "cedilha", "naïve café": "naive cafe",
		"e\u0301": "e", "İstanbul": "istanbul", "日本": "日本", "": "", "ABC": "abc",
	} {
		assert.Equal(t, want, foldName(in), in)
	}
}

func TestOnlyTheReservedWordIsReserved(t *testing.T) {
	for in, want := range map[string]bool{
		"favorites": true, "Favorites": true, "fav orites": true, "FÁVÓRÍTÈS": true,
		"": false, "favorite": false, "favorites1": false, "my favorites": false, "favorit es s": false,
	} {
		assert.Equal(t, want, isReserved(in), in)
	}
}
