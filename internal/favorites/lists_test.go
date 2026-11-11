package favorites

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ------------------------------------------------------------------

var t0 = time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

// pathOf is where a test store keeps its file.
func pathOf(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "favorites.json")
}

// tickStore is a store whose clock moves one second per call, so every stamp
// is unique and ordered.
func tickStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := pathOf(t)
	s := Open(path)
	n := 0
	s.now = func() time.Time { n++; return t0.Add(time.Duration(n) * time.Second) }
	return s, path
}

func mkList(t *testing.T, s *Store, name string) Summary {
	t.Helper()
	sum, err := s.CreateList(name)
	require.NoError(t, err, name)
	return sum
}

func mustToggleIn(t *testing.T, s *Store, id, key string) bool {
	t.Helper()
	on, err := s.ToggleIn(id, Entry{Key: key, Title: strings.ToUpper(key)})
	require.NoError(t, err)
	return on
}

func entriesOf(t *testing.T, s *Store, id string) []string {
	t.Helper()
	list, err := s.Entries(id)
	require.NoError(t, err)
	return keysOf(list)
}

func summaries(t *testing.T, s *Store) []Summary {
	t.Helper()
	out, err := s.Summaries()
	require.NoError(t, err)
	return out
}

func namesOf(sums []Summary) []string {
	out := make([]string, len(sums))
	for i := range sums {
		out[i] = sums[i].Name
	}
	return out
}

func idsOf(sums []Summary) []string {
	out := make([]string, len(sums))
	for i := range sums {
		out[i] = sums[i].ID
	}
	return out
}

// writeDoc writes a document the way an older or a hand-editing hand would.
func writeDoc(t *testing.T, path string, f file) {
	t.Helper()
	data, err := json.MarshalIndent(f, "", "  ")
	require.NoError(t, err)
	writeRaw(t, path, string(data))
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func repeat(s string, n int) string { return strings.Repeat(s, n) }

// ---- the contract's own words -------------------------------------------------

func TestListErrorsAreLowercasePhrasesWithNoFinalPeriod(t *testing.T) {
	want := map[error]string{
		ErrNameEmpty:    "give the list a name",
		ErrNameTooLong:  "a list name can have at most 40 characters",
		ErrNameTaken:    "you already have a list with that name",
		ErrReservedName: "that name is already used by the Favorites list",
		ErrTooManyLists: "you can have at most 50 lists, delete one first",
		ErrNoSuchList:   "that list no longer exists",
		ErrAutoList:     "the Favorites list cannot be renamed or deleted",
	}
	for err, text := range want {
		assert.Equal(t, text, err.Error())
		assert.False(t, strings.HasSuffix(err.Error(), "."))
	}
	assert.Equal(t, 40, maxListNameRunes)
	assert.Equal(t, 50, maxLists)
	assert.Equal(t, "favorites", FavoritesID)
}

// ---- names --------------------------------------------------------------------

func TestCreateListNormalizesTheName(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Weekend marathon", "Weekend marathon"},
		{"  Weekend   marathon  ", "Weekend marathon"},
		{"a\tb\nc\r\nd", "a b c d"},
		{"a\x00b", "ab"},
		{"a \x1b b", "a b"},
		{"\x7fx\x7f", "x"},
		{"\u00a0x\u00a0", "x"},    // no-break spaces are white space
		{"x\u2003\u2003y", "x y"}, // em spaces too
		{"ab\xffc", "abc"},        // not UTF-8: dropped
		{"Maratón de fim", "Maratón de fim"},
		{"Família 👨\u200d👩\u200d👧 hoje", "Família 👨\u200d👩\u200d👧 hoje"},
		{"<b>x</b> & \"q\"", "<b>x</b> & \"q\""}, // stored as given
		{"../../etc/passwd", "../../etc/passwd"},
		{"日本語 のリスト", "日本語 のリスト"},
	} {
		t.Run(strconv.Quote(tc.in), func(t *testing.T) {
			s := newStore(t)
			sum, err := s.CreateList(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, sum.Name)
			got := summaries(t, s)
			require.Len(t, got, 2)
			assert.Equal(t, tc.want, got[1].Name, "what is saved is what Create returned")
		})
	}
}

func TestAnEmptyNameIsRefused(t *testing.T) {
	for _, in := range []string{"", " ", "   ", "\t", "\n\r\n", "\x00", "\x00\x01 \x02", "\u00a0\u2003", "\xff\xfe"} {
		t.Run(strconv.Quote(in), func(t *testing.T) {
			s := newStore(t)
			_, err := s.CreateList(in)
			assert.ErrorIs(t, err, ErrNameEmpty)
			assert.Len(t, summaries(t, s), 1)
			_, statErr := os.Stat(s.path)
			assert.ErrorIs(t, statErr, os.ErrNotExist, "a refused name writes nothing")
		})
	}
}

func TestTheNameLimitIsFortyRunesNotBytes(t *testing.T) {
	for _, tc := range []struct {
		label, name string
		ok          bool
	}{
		{"ascii 40", repeat("a", 40), true},
		{"ascii 41", repeat("a", 41), false},
		{"accents 40 (80 bytes)", repeat("é", 40), true},
		{"accents 41", repeat("é", 41), false},
		{"emoji 40 (160 bytes)", repeat("🎬", 40), true},
		{"emoji 41", repeat("🎬", 41), false},
		{"mixed 40", repeat("a", 20) + repeat("ã", 10) + repeat("🎬", 10), true},
		{"mixed 41", repeat("a", 20) + repeat("ã", 10) + repeat("🎬", 11), false},
		{"japanese 40", repeat("日", 40), true},
		{"japanese 41", repeat("日", 41), false},
		// the spaces are collapsed before they are counted
		{"40 after collapsing", repeat("a", 19) + "     " + repeat("b", 20), true},
		{"41 after collapsing", repeat("a", 20) + "     " + repeat("b", 20), false},
		{"padding does not count", "   " + repeat("a", 40) + "   ", true},
		// combining accents are runes of their own: 20 pairs are 40, 21 are 42
		{"decomposed 20 pairs", repeat("e\u0301", 20), true},
		{"decomposed 21 pairs", repeat("e\u0301", 21), false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			s := newStore(t)
			sum, err := s.CreateList(tc.name)
			if tc.ok {
				require.NoError(t, err)
				assert.Equal(t, cleanName(tc.name), sum.Name, "a name at the limit is kept whole")
				return
			}
			assert.ErrorIs(t, err, ErrNameTooLong)
			assert.Len(t, summaries(t, s), 1)
		})
	}
}

func TestTheReservedNameInAnyDisguise(t *testing.T) {
	for _, name := range []string{
		"favorites", "Favorites", "FAVORITES", "  Favorites  ", "FaVoRiTeS",
		"Favorités", "FÁVORITES", "fàvôrítès", "Fav orites", "F a V o R i T e S", "fa\tvorites",
		"favorites\x00", "Favori\u0301tes", // "i" with a combining accent
	} {
		t.Run(strconv.Quote(name), func(t *testing.T) {
			s := newStore(t)
			_, err := s.CreateList(name)
			assert.ErrorIs(t, err, ErrReservedName)
			assert.Len(t, summaries(t, s), 1)

			sum := mkList(t, s, "Weekend")
			assert.ErrorIs(t, s.RenameList(sum.ID, name), ErrReservedName)
			assert.Equal(t, "Weekend", summaries(t, s)[1].Name)
		})
	}
	for _, name := range []string{"Favorites 2", "My favorites", "Favorite", "favoritess", "Favorito", "Fav"} {
		t.Run("not "+name, func(t *testing.T) {
			_, err := newStore(t).CreateList(name)
			assert.NoError(t, err)
		})
	}
}

func TestALongNameMadeOfTheReservedWordIsTooLongNotReserved(t *testing.T) {
	// a name made only of "favorites" is far over 40 runes: too long wins,
	// because it is not the reserved name.
	s := newStore(t)
	_, err := s.CreateList(repeat("favorites", 6))
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestATakenNameIgnoresCaseAndAccentsButNotSpaces(t *testing.T) {
	s := newStore(t)
	mkList(t, s, "Weekend")
	for _, name := range []string{"weekend", "WEEKEND", " Weekend ", "Weekénd", "WÉÉKÉND", "Weeke\u0301nd", "wEEkEnD"} {
		_, err := s.CreateList(name)
		assert.ErrorIs(t, err, ErrNameTaken, name)
	}
	_, err := s.CreateList("Wee\u200bkend")
	assert.ErrorIs(t, err, ErrNameTaken, "an invisible character does not make a new name")
	for _, name := range []string{"Weekend 2", "Week end", "Weekends", "end", "Weekend·"} {
		_, err := s.CreateList(name)
		assert.NoError(t, err, name)
	}
}

func TestAnAccentedNameIsTakenByItsPlainSpelling(t *testing.T) {
	s := newStore(t)
	mkList(t, s, "Maratón")
	_, err := s.CreateList("maraton")
	assert.ErrorIs(t, err, ErrNameTaken)
	_, err = s.CreateList("MARATÓN")
	assert.ErrorIs(t, err, ErrNameTaken)
	_, err = s.CreateList("Maratões")
	assert.NoError(t, err)
}

func TestTheFoldIsNeverEmptyForARealName(t *testing.T) {
	// a name of only combining marks folds to "" but is still a name that
	// compares equal to itself; it must not equal every other name.
	s := newStore(t)
	mkList(t, s, "\u0301\u0301")
	_, err := s.CreateList("Weekend")
	assert.NoError(t, err)
}

// ---- Create -------------------------------------------------------------------

func TestCreateListReturnsWhatIsSaved(t *testing.T) {
	s, path := tickStore(t)
	sum, err := s.CreateList("Weekend marathon")
	require.NoError(t, err)
	assert.Equal(t, "Weekend marathon", sum.Name)
	assert.False(t, sum.Auto)
	assert.Equal(t, 0, sum.Count)
	assert.Empty(t, sum.Titles)
	assert.Equal(t, t0.Add(time.Second), sum.CreatedAt)
	assert.Equal(t, time.UTC, sum.CreatedAt.Location())
	assert.NotEmpty(t, sum.ID)

	got := summaries(t, Open(path))
	require.Len(t, got, 2)
	assert.Equal(t, sum, got[1], "a fresh store sees the same list")
}

func TestCreateListStampsTheRealClockWhenThereIsNone(t *testing.T) {
	s := newStore(t)
	before := time.Now()
	sum := mkList(t, s, "Now")
	assert.False(t, sum.CreatedAt.Before(before.Add(-time.Second)))
	assert.False(t, sum.CreatedAt.After(time.Now().Add(time.Second)))
}

var urlSafeID = regexp.MustCompile(`^l[0-9a-z]+$`)

func TestListIDsAreUniqueUrlSafeAndNeverFavorites(t *testing.T) {
	s := newStore(t)
	s.now = func() time.Time { return t0 } // every list made at the same instant
	seen := map[string]bool{FavoritesID: true}
	for i := range maxLists {
		sum := mkList(t, s, "list "+strconv.Itoa(i))
		assert.Regexp(t, urlSafeID, sum.ID)
		assert.False(t, seen[sum.ID], "ID %s repeats", sum.ID)
		seen[sum.ID] = true
	}
	assert.Len(t, seen, maxLists+1)
}

func TestIDsAreUniqueUnderRapidCreateAndDelete(t *testing.T) {
	s := newStore(t) // real clock: successive calls may share a nanosecond reading
	seen := map[string]bool{}
	for i := range 100 {
		sum := mkList(t, s, "n"+strconv.Itoa(i))
		assert.False(t, seen[sum.ID], "ID %s repeats", sum.ID)
		seen[sum.ID] = true
		assert.NotEqual(t, FavoritesID, sum.ID)
		if i%2 == 0 {
			require.NoError(t, s.DeleteList(sum.ID))
		}
	}
}

func TestAnIDIsBumpedPastOneAlreadyInUse(t *testing.T) {
	s := newStore(t)
	s.now = func() time.Time { return time.Unix(0, 36*36) } // base 36: "100"
	a := mkList(t, s, "a")
	b := mkList(t, s, "b")
	c := mkList(t, s, "c")
	assert.Equal(t, "l100", a.ID)
	assert.Equal(t, "l101", b.ID)
	assert.Equal(t, "l102", c.ID)
	require.NoError(t, s.DeleteList(b.ID))
	d := mkList(t, s, "d")
	assert.Equal(t, "l101", d.ID, "the first free ID is taken again")
}

func TestAnIDIsFreeOfAnInstantBeforeTheEpoch(t *testing.T) {
	s := newStore(t)
	s.now = func() time.Time { return time.Unix(-5, 0) }
	a := mkList(t, s, "a")
	b := mkList(t, s, "b")
	assert.Regexp(t, urlSafeID, a.ID)
	assert.Regexp(t, urlSafeID, b.ID)
	assert.NotEqual(t, a.ID, b.ID)
}

func TestFiftyListsAreAllowedAndTheFiftyFirstIsNot(t *testing.T) {
	s := newStore(t)
	for i := range maxLists {
		mkList(t, s, "list "+strconv.Itoa(i))
	}
	assert.Len(t, summaries(t, s), maxLists+1)

	before := readBytes(t, s.path)
	_, err := s.CreateList("one more")
	assert.ErrorIs(t, err, ErrTooManyLists)
	assert.Equal(t, before, readBytes(t, s.path), "nothing is written")

	// a bad name is still reported as the name's fault
	_, err = s.CreateList("")
	assert.ErrorIs(t, err, ErrNameEmpty)
	_, err = s.CreateList("list 3")
	assert.ErrorIs(t, err, ErrNameTaken)
	_, err = s.CreateList("favorites")
	assert.ErrorIs(t, err, ErrReservedName)
	_, err = s.CreateList(repeat("x", 41))
	assert.ErrorIs(t, err, ErrNameTooLong)

	// room again after one is deleted
	sums := summaries(t, s)
	require.NoError(t, s.DeleteList(sums[7].ID))
	_, err = s.CreateList("one more")
	assert.NoError(t, err)
	_, err = s.CreateList("and another")
	assert.ErrorIs(t, err, ErrTooManyLists)
}

// ---- Rename -------------------------------------------------------------------

func TestRenameListChangesOnlyThatName(t *testing.T) {
	s, path := tickStore(t)
	a := mkList(t, s, "Alpha")
	b := mkList(t, s, "Beta")
	mustToggleIn(t, s, a.ID, "w")
	require.NoError(t, s.RenameList(a.ID, "  Gamma   ray "))

	got := summaries(t, Open(path))
	assert.Equal(t, []string{"Favorites", "Gamma ray", "Beta"}, namesOf(got))
	assert.Equal(t, []string{FavoritesID, a.ID, b.ID}, idsOf(got), "IDs and order are kept")
	assert.Equal(t, a.CreatedAt, got[1].CreatedAt)
	assert.Equal(t, 1, got[1].Count, "the works are kept")
}

func TestRenameToItsOwnNameIsAllowed(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "Alpha")
	mkList(t, s, "Beta")
	require.NoError(t, s.RenameList(a.ID, "Alpha"))
	require.NoError(t, s.RenameList(a.ID, "ALPHA"), "only the case changes")
	assert.Equal(t, "ALPHA", summaries(t, s)[1].Name)
	require.NoError(t, s.RenameList(a.ID, "Alphá"), "only the accent changes")
	assert.Equal(t, "Alphá", summaries(t, s)[1].Name)
	require.NoError(t, s.RenameList(a.ID, "  alpha  "))
	assert.Equal(t, "alpha", summaries(t, s)[1].Name)
}

func TestRenameToAnotherListsNameIsRefused(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "Alpha")
	b := mkList(t, s, "Beta")
	before := readBytes(t, s.path)
	for _, name := range []string{"Beta", "beta", "BÉTA", "bêta"} {
		assert.ErrorIs(t, s.RenameList(a.ID, name), ErrNameTaken, name)
	}
	assert.Equal(t, before, readBytes(t, s.path), "nothing is written")
	assert.ErrorIs(t, s.RenameList(b.ID, "ALPHA"), ErrNameTaken)
	require.NoError(t, s.RenameList(b.ID, "Beta 2"))
	require.NoError(t, s.RenameList(a.ID, "Beta"), "the name is free once it is released")
}

func TestRenameValidatesTheNameLikeCreate(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "Alpha")
	before := readBytes(t, s.path)
	assert.ErrorIs(t, s.RenameList(a.ID, ""), ErrNameEmpty)
	assert.ErrorIs(t, s.RenameList(a.ID, " \t\x00 "), ErrNameEmpty)
	assert.ErrorIs(t, s.RenameList(a.ID, repeat("é", 41)), ErrNameTooLong)
	assert.ErrorIs(t, s.RenameList(a.ID, "Favorites"), ErrReservedName)
	assert.Equal(t, before, readBytes(t, s.path))
	require.NoError(t, s.RenameList(a.ID, repeat("🎬", 40)))
	assert.Equal(t, repeat("🎬", 40), summaries(t, s)[1].Name)
}

func TestRenameOfTheAutomaticOrAMissingList(t *testing.T) {
	s := newStore(t)
	mkList(t, s, "Alpha")
	assert.ErrorIs(t, s.RenameList(FavoritesID, "Mine"), ErrAutoList)
	assert.ErrorIs(t, s.RenameList(FavoritesID, ""), ErrAutoList, "the list is the problem, not the name")
	assert.ErrorIs(t, s.RenameList("nope", "Mine"), ErrNoSuchList)
	assert.ErrorIs(t, s.RenameList("", "Mine"), ErrNoSuchList)
	assert.ErrorIs(t, s.RenameList("L1", "Mine"), ErrNoSuchList, "IDs are exact")
	assert.Equal(t, []string{"Favorites", "Alpha"}, namesOf(summaries(t, s)))

	empty := newStore(t) // no file at all
	assert.ErrorIs(t, empty.RenameList("x", "y"), ErrNoSuchList)
	_, err := os.Stat(empty.path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

// ---- Delete -------------------------------------------------------------------

func TestDeleteListRemovesTheListAndItsWorksOnly(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "Alpha")
	b := mkList(t, s, "Beta")
	c := mkList(t, s, "Gamma")
	mustToggleIn(t, s, FavoritesID, "w")
	mustToggleIn(t, s, a.ID, "w")
	mustToggleIn(t, s, b.ID, "w")
	mustToggleIn(t, s, b.ID, "only-b")

	require.NoError(t, s.DeleteList(b.ID))
	got := summaries(t, s)
	assert.Equal(t, []string{FavoritesID, a.ID, c.ID}, idsOf(got))
	assert.Equal(t, map[string]bool{FavoritesID: true, a.ID: true}, s.Membership("w"))
	assert.Empty(t, s.Membership("only-b"))
	_, err := s.Entries(b.ID)
	assert.ErrorIs(t, err, ErrNoSuchList)
	assert.Equal(t, []string{"w"}, entriesOf(t, s, FavoritesID), "Favorites is untouched")
	assert.Equal(t, []string{"w"}, entriesOf(t, s, a.ID))
}

func TestDeleteTheOnlyAndTheLastAndTheFirst(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "A")
	require.NoError(t, s.DeleteList(a.ID))
	assert.Len(t, summaries(t, s), 1)

	a = mkList(t, s, "A")
	b := mkList(t, s, "B")
	c := mkList(t, s, "C")
	require.NoError(t, s.DeleteList(c.ID))
	assert.Equal(t, []string{FavoritesID, a.ID, b.ID}, idsOf(summaries(t, s)))
	require.NoError(t, s.DeleteList(a.ID))
	assert.Equal(t, []string{FavoritesID, b.ID}, idsOf(summaries(t, s)))
}

func TestDeleteOfTheAutomaticOrAMissingList(t *testing.T) {
	s := newStore(t)
	mustToggleIn(t, s, FavoritesID, "w")
	assert.ErrorIs(t, s.DeleteList(FavoritesID), ErrAutoList)
	assert.ErrorIs(t, s.DeleteList("nope"), ErrNoSuchList)
	assert.ErrorIs(t, s.DeleteList(""), ErrNoSuchList)
	a := mkList(t, s, "A")
	require.NoError(t, s.DeleteList(a.ID))
	assert.ErrorIs(t, s.DeleteList(a.ID), ErrNoSuchList, "the second delete finds nothing")
	assert.Equal(t, []string{"w"}, entriesOf(t, s, FavoritesID))
}

func TestDeleteThenRecreateTheSameName(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "Weekend")
	mustToggleIn(t, s, a.ID, "w")
	require.NoError(t, s.DeleteList(a.ID))
	b := mkList(t, s, "Weekend")
	assert.NotEqual(t, a.ID, b.ID, "a new list, not the old one back")
	assert.Equal(t, 0, summaries(t, s)[1].Count)
	assert.Empty(t, s.Membership("w"))
	_, err := s.Entries(a.ID)
	assert.ErrorIs(t, err, ErrNoSuchList)
}

// ---- Entries / ToggleIn ------------------------------------------------------

func TestEntriesAreNewestFirstInEveryList(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	for i, k := range []string{"x", "y", "z"} {
		_, err := s.ToggleIn(a.ID, Entry{Key: k, AddedAt: t0.Add(time.Duration(i) * time.Hour)})
		require.NoError(t, err)
		_, err = s.ToggleIn(FavoritesID, Entry{Key: k + "f", AddedAt: t0.Add(time.Duration(-i) * time.Hour)})
		require.NoError(t, err)
	}
	assert.Equal(t, []string{"z", "y", "x"}, entriesOf(t, s, a.ID))
	assert.Equal(t, []string{"xf", "yf", "zf"}, entriesOf(t, s, FavoritesID))
	assert.Equal(t, mustList(t, s), mustEntries(t, s, FavoritesID), "List is Entries(FavoritesID)")
}

func mustEntries(t *testing.T, s *Store, id string) []Entry {
	t.Helper()
	list, err := s.Entries(id)
	require.NoError(t, err)
	return list
}

func TestEntriesTiesKeepTheFileOrder(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	for _, k := range []string{"one", "two", "three", "four"} {
		_, err := s.ToggleIn(a.ID, Entry{Key: k, AddedAt: t0})
		require.NoError(t, err)
	}
	assert.Equal(t, []string{"one", "two", "three", "four"}, entriesOf(t, s, a.ID))
}

func TestEntriesOfAnEmptyOrMissingList(t *testing.T) {
	s := newStore(t)
	assert.Empty(t, entriesOf(t, s, FavoritesID), "no file yet")
	_, err := s.Entries("nope")
	assert.ErrorIs(t, err, ErrNoSuchList)
	a := mkList(t, s, "A")
	assert.Empty(t, entriesOf(t, s, a.ID))
	_, err = s.Entries("")
	assert.ErrorIs(t, err, ErrNoSuchList)
	_, err = s.Entries("FAVORITES")
	assert.ErrorIs(t, err, ErrNoSuchList, "IDs are exact")
}

func TestToggleInAddsThenRemoves(t *testing.T) {
	s, path := tickStore(t)
	a := mkList(t, s, "A")
	on, err := s.ToggleIn(a.ID, Entry{Key: " Naruto ", Title: "Naruto", Year: "2002", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)
	assert.True(t, on)

	list := mustEntries(t, Open(path), a.ID)
	require.Len(t, list, 1)
	assert.Equal(t, "naruto", list[0].Key, "the key is normalized")
	assert.Equal(t, Entry{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime", Source: "HiAnime", AddedAt: list[0].AddedAt}, list[0])
	assert.False(t, list[0].AddedAt.IsZero(), "AddedAt is stamped from the clock")
	assert.False(t, Open(path).Has("naruto"), "and Favorites does not have it")

	on, err = s.ToggleIn(a.ID, Entry{Key: "NARUTO"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.Empty(t, entriesOf(t, Open(path), a.ID))
}

func TestToggleInStampsOnlyAZeroAddedAt(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "A")
	given := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
	mustToggleInAt(t, s, a.ID, "old", given)
	mustToggleIn(t, s, a.ID, "new")
	list := mustEntries(t, s, a.ID)
	require.Len(t, list, 2)
	assert.Equal(t, "new", list[0].Key)
	assert.True(t, list[0].AddedAt.After(t0))
	assert.True(t, given.Equal(list[1].AddedAt))
}

func mustToggleInAt(t *testing.T, s *Store, id, key string, at time.Time) {
	t.Helper()
	_, err := s.ToggleIn(id, Entry{Key: key, AddedAt: at})
	require.NoError(t, err)
}

func TestToggleInNeedsAKeyAndWritesNothing(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	before := readBytes(t, s.path)
	for _, key := range []string{"", "   ", "\t\n"} {
		on, err := s.ToggleIn(a.ID, Entry{Key: key, Title: "x"})
		require.Error(t, err)
		assert.False(t, on)
		assert.NotErrorIs(t, err, ErrNoSuchList)
		_, err = s.ToggleIn(FavoritesID, Entry{Key: key})
		require.Error(t, err)
	}
	assert.Equal(t, before, readBytes(t, s.path))
}

func TestToggleInOnAMissingListIsAnErrorAndWritesNothing(t *testing.T) {
	s := newStore(t)
	on, err := s.ToggleIn("nope", Entry{Key: "w"})
	assert.ErrorIs(t, err, ErrNoSuchList)
	assert.False(t, on)
	_, statErr := os.Stat(s.path)
	assert.ErrorIs(t, statErr, os.ErrNotExist)

	a := mkList(t, s, "A")
	require.NoError(t, s.DeleteList(a.ID))
	_, err = s.ToggleIn(a.ID, Entry{Key: "w"})
	assert.ErrorIs(t, err, ErrNoSuchList, "a list deleted under the screen")
	assert.ErrorIs(t, errors.Unwrap(fmt.Errorf("x: %w", err)), ErrNoSuchList)
}

func TestToggleInFavoritesIDIsToggle(t *testing.T) {
	s := newStore(t)
	on, err := s.ToggleIn(FavoritesID, Entry{Key: " Frieren ", Title: "Frieren"})
	require.NoError(t, err)
	assert.True(t, on)
	assert.True(t, s.Has("frieren"))
	assert.Equal(t, map[string]bool{"frieren": true}, s.Keys())
	on, err = s.Toggle(Entry{Key: "frieren"})
	require.NoError(t, err)
	assert.False(t, on, "Toggle and ToggleIn(FavoritesID) act on the same list")
	assert.False(t, s.Has("frieren"))
}

func TestAListHoldsExactlyFiveHundredWorksAndDropsTheOldest(t *testing.T) {
	path := pathOf(t)
	doc := file{Version: 1, Lists: []list{{ID: "lfull", Name: "Full", CreatedAt: t0}}}
	for i := range maxEntries {
		doc.Lists[0].Entries = append(doc.Lists[0].Entries, Entry{Key: "k" + strconv.Itoa(i), Title: "t", AddedAt: t0.Add(time.Duration(i) * time.Second)})
	}
	writeDoc(t, path, doc)
	s := Open(path)

	assert.Len(t, mustEntries(t, s, "lfull"), maxEntries, "500 fit")
	on, err := s.ToggleIn("lfull", Entry{Key: "extra", AddedAt: t0.Add(time.Hour)})
	require.NoError(t, err)
	assert.True(t, on)
	got := mustEntries(t, s, "lfull")
	require.Len(t, got, maxEntries, "the 501st pushes the oldest out")
	assert.Equal(t, "extra", got[0].Key)
	assert.Equal(t, "k499", got[1].Key)
	assert.Equal(t, "k1", got[maxEntries-1].Key)
	assert.NotContains(t, keysOf(got), "k0")

	// removing one makes room, with no loss
	require.NoError(t, s.RemoveFrom("lfull", "k250"))
	on, err = s.ToggleIn("lfull", Entry{Key: "extra2", AddedAt: t0.Add(2 * time.Hour)})
	require.NoError(t, err)
	assert.True(t, on)
	got = mustEntries(t, s, "lfull")
	require.Len(t, got, maxEntries)
	assert.Contains(t, keysOf(got), "k1")
}

func TestTheOldestIsByDateNotByPositionInAList(t *testing.T) {
	path := pathOf(t)
	doc := file{Version: 1, Lists: []list{{ID: "lfull", Name: "Full", CreatedAt: t0}}}
	for i := range maxEntries { // the file order is newest first: reversed
		doc.Lists[0].Entries = append(doc.Lists[0].Entries, Entry{Key: "k" + strconv.Itoa(i), AddedAt: t0.Add(time.Duration(maxEntries-i) * time.Second)})
	}
	writeDoc(t, path, doc)
	s := Open(path)
	mustToggleInAt(t, s, "lfull", "extra", t0.Add(time.Hour))
	got := keysOf(mustEntries(t, s, "lfull"))
	assert.Len(t, got, maxEntries)
	assert.NotContains(t, got, "k499", "the oldest by date went, not the last in the file")
	assert.Contains(t, got, "k0")
}

func TestTheCapOfOneListDoesNotTouchAnother(t *testing.T) {
	path := pathOf(t)
	full := list{ID: "lfull", Name: "Full", CreatedAt: t0}
	for i := range maxEntries {
		full.Entries = append(full.Entries, Entry{Key: "k" + strconv.Itoa(i), AddedAt: t0.Add(time.Duration(i) * time.Second)})
	}
	writeDoc(t, path, file{Version: 1, Entries: []Entry{{Key: "fav", AddedAt: t0}}, Lists: []list{full, {ID: "lsmall", Name: "Small", CreatedAt: t0.Add(time.Second)}}})
	s := Open(path)
	mustToggleInAt(t, s, "lfull", "extra", t0.Add(time.Hour))
	mustToggleInAt(t, s, "lsmall", "extra", t0.Add(time.Hour))
	assert.Equal(t, []string{"extra"}, entriesOf(t, s, "lsmall"))
	assert.Equal(t, []string{"fav"}, entriesOf(t, s, FavoritesID))
}

func TestTheSameWorkCanBeInManyLists(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	b := mkList(t, s, "B")
	assert.True(t, mustToggleIn(t, s, a.ID, "w"))
	assert.True(t, mustToggleIn(t, s, b.ID, "w"), "being in A does not make it in B")
	assert.True(t, mustToggleIn(t, s, FavoritesID, "w"))
	assert.Equal(t, map[string]bool{FavoritesID: true, a.ID: true, b.ID: true}, s.Membership("w"))

	assert.False(t, mustToggleIn(t, s, a.ID, "w"), "toggling in A takes it off A only")
	assert.Equal(t, map[string]bool{FavoritesID: true, b.ID: true}, s.Membership("w"))
	assert.Equal(t, []string{"w"}, entriesOf(t, s, b.ID))
}

func TestTheSameWorkTwiceInOneListIsOneWork(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	assert.True(t, mustToggleIn(t, s, a.ID, "Dup"))
	assert.False(t, mustToggleIn(t, s, a.ID, " dup "), "the second toggle takes it off")
	assert.Empty(t, entriesOf(t, s, a.ID))
}

// ---- RemoveFrom / ClearList ---------------------------------------------------

func TestRemoveFromTakesTheKeyOffThatListOnly(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	b := mkList(t, s, "B")
	for _, id := range []string{FavoritesID, a.ID, b.ID} {
		mustToggleIn(t, s, id, "w")
		mustToggleIn(t, s, id, "keep")
	}
	require.NoError(t, s.RemoveFrom(a.ID, "  W "))
	assert.Equal(t, []string{"keep"}, entriesOf(t, s, a.ID))
	assert.Equal(t, []string{"keep", "w"}, sortedCopy(entriesOf(t, s, b.ID)))
	assert.Equal(t, []string{"keep", "w"}, sortedCopy(entriesOf(t, s, FavoritesID)))

	require.NoError(t, s.RemoveFrom(FavoritesID, "w"))
	assert.False(t, s.Has("w"))
	assert.Equal(t, map[string]bool{b.ID: true}, s.Membership("w"))
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func TestRemoveFromAMissingKeyIsNotAnErrorAndWritesNothing(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	mustToggleIn(t, s, a.ID, "w")
	before := readBytes(t, s.path)
	require.NoError(t, s.RemoveFrom(a.ID, "not there"))
	require.NoError(t, s.RemoveFrom(a.ID, ""))
	require.NoError(t, s.RemoveFrom(FavoritesID, "not there"))
	assert.Equal(t, before, readBytes(t, s.path))
}

func TestRemoveFromAMissingList(t *testing.T) {
	s := newStore(t)
	assert.ErrorIs(t, s.RemoveFrom("nope", "w"), ErrNoSuchList)
	assert.ErrorIs(t, s.ClearList("nope"), ErrNoSuchList)
	_, err := os.Stat(s.path)
	assert.ErrorIs(t, err, os.ErrNotExist, "nothing is created for a list that is not there")
}

func TestClearListEmptiesItButKeepsTheListAndTheOthers(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	b := mkList(t, s, "B")
	for _, id := range []string{FavoritesID, a.ID, b.ID} {
		mustToggleIn(t, s, id, "w")
		mustToggleIn(t, s, id, "x")
	}
	require.NoError(t, s.ClearList(a.ID))
	assert.Empty(t, entriesOf(t, s, a.ID))
	assert.Equal(t, []string{"A", "B"}, namesOf(summaries(t, s)[1:]), "the list stays")
	assert.Len(t, entriesOf(t, s, b.ID), 2)
	assert.Len(t, entriesOf(t, s, FavoritesID), 2)

	require.NoError(t, s.ClearList(FavoritesID))
	assert.Empty(t, mustList(t, s))
	assert.Len(t, entriesOf(t, s, b.ID), 2, "clearing Favorites keeps the lists")
	assert.Len(t, summaries(t, s), 3)

	require.NoError(t, s.ClearList(a.ID), "clearing an empty list is fine")
	assert.True(t, mustToggleIn(t, s, a.ID, "again"), "and it can be filled again")
}

func TestAClearedListIsWrittenAsAnEmptyArray(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	mustToggleIn(t, s, a.ID, "w")
	require.NoError(t, s.ClearList(a.ID))
	doc := readFileJSON(t, s.path)
	lists, ok := doc["lists"].([]any)
	require.True(t, ok)
	require.Len(t, lists, 1)
	entries, ok := lists[0].(map[string]any)["entries"].([]any)
	assert.True(t, ok, "entries is [], not null")
	assert.Empty(t, entries)
}

// ---- Summaries ----------------------------------------------------------------

func TestSummariesAreFavoritesFirstThenOldestFirst(t *testing.T) {
	s := newStore(t)
	assert.Equal(t, []Summary{{ID: FavoritesID, Name: "Favorites", Auto: true, Titles: []string{}}}, summaries(t, s),
		"with no file there is still the automatic list")

	path := s.path
	writeDoc(t, path, file{Version: 1, Lists: []list{
		{ID: "lc", Name: "Third", CreatedAt: t0.Add(3 * time.Hour)},
		{ID: "la", Name: "First", CreatedAt: t0.Add(1 * time.Hour)},
		{ID: "lb", Name: "Second", CreatedAt: t0.Add(2 * time.Hour)},
	}})
	got := summaries(t, s)
	assert.Equal(t, []string{FavoritesID, "la", "lb", "lc"}, idsOf(got))
	assert.Equal(t, []string{"Favorites", "First", "Second", "Third"}, namesOf(got))
}

func TestSummariesTiesKeepTheFileOrder(t *testing.T) {
	writeTies := func(t *testing.T, ids ...string) *Store {
		path := pathOf(t)
		doc := file{Version: 1}
		for _, id := range ids {
			doc.Lists = append(doc.Lists, list{ID: id, Name: "n-" + id, CreatedAt: t0})
		}
		writeDoc(t, path, doc)
		return Open(path)
	}
	assert.Equal(t, []string{FavoritesID, "lz", "la", "lm"}, idsOf(summaries(t, writeTies(t, "lz", "la", "lm"))))
	assert.Equal(t, []string{FavoritesID, "lm", "la", "lz"}, idsOf(summaries(t, writeTies(t, "lm", "la", "lz"))))
}

func TestSummariesCarryTheFieldsTheScreensShow(t *testing.T) {
	s, _ := tickStore(t)
	a := mkList(t, s, "A")
	mkList(t, s, "B")
	for i, title := range []string{"Old", "Middle", "New"} {
		_, err := s.ToggleIn(a.ID, Entry{Key: "k" + strconv.Itoa(i), Title: title, AddedAt: t0.Add(time.Duration(i) * time.Hour)})
		require.NoError(t, err)
	}
	_, err := s.Toggle(Entry{Key: "f1", Title: "Fav one", AddedAt: t0})
	require.NoError(t, err)
	_, err = s.Toggle(Entry{Key: "f2", Title: "Fav two", AddedAt: t0.Add(time.Hour)})
	require.NoError(t, err)

	got := summaries(t, s)
	require.Len(t, got, 3)
	assert.Equal(t, Summary{ID: FavoritesID, Name: "Favorites", Auto: true, Count: 2, Titles: []string{"Fav two", "Fav one"}}, got[0])
	assert.Equal(t, a.ID, got[1].ID)
	assert.Equal(t, "A", got[1].Name)
	assert.False(t, got[1].Auto)
	assert.Equal(t, a.CreatedAt, got[1].CreatedAt)
	assert.Equal(t, 3, got[1].Count)
	assert.Equal(t, []string{"New", "Middle", "Old"}, got[1].Titles, "titles are newest first")
	assert.False(t, got[2].Auto)
	assert.Equal(t, 0, got[2].Count)
	assert.Empty(t, got[2].Titles)
	assert.True(t, got[0].CreatedAt.IsZero())
}

func TestSummariesHaveNoAliasWithTheFile(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	mustToggleIn(t, s, a.ID, "w")
	got := summaries(t, s)
	got[1].Titles[0] = "mutated"
	got[1].Name = "mutated"
	assert.Equal(t, []string{"W"}, summaries(t, s)[1].Titles)
}

func TestUntrustedTitlesAndNamesNeverPanic(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	for i, title := range []string{"", "\x00\x01", "\xff\xfe", repeat("x", 100000), "<script>", "line\nbreak", "🎬"} {
		_, err := s.ToggleIn(a.ID, Entry{Key: "k" + strconv.Itoa(i), Title: title})
		require.NoError(t, err)
	}
	got := summaries(t, s)
	assert.Equal(t, 7, got[1].Count)
	assert.Len(t, got[1].Titles, 7)
}

// ---- Membership --------------------------------------------------------------

func TestMembershipFollowsToggleRemoveClearAndDelete(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "A")
	b := mkList(t, s, "B")
	assert.Empty(t, s.Membership("w"))
	assert.NotNil(t, s.Membership("w"), "never nil")

	mustToggleIn(t, s, FavoritesID, "w")
	assert.Equal(t, map[string]bool{FavoritesID: true}, s.Membership("w"), "Favorites counts")
	mustToggleIn(t, s, a.ID, "W")
	mustToggleIn(t, s, b.ID, " w ")
	assert.Equal(t, map[string]bool{FavoritesID: true, a.ID: true, b.ID: true}, s.Membership("w"))
	assert.Equal(t, s.Membership("w"), s.Membership("  W"), "the key is normalized")
	assert.Empty(t, s.Membership("other"))
	assert.Empty(t, s.Membership(""))

	mustToggleIn(t, s, a.ID, "w") // off
	assert.Equal(t, map[string]bool{FavoritesID: true, b.ID: true}, s.Membership("w"))
	require.NoError(t, s.RemoveFrom(FavoritesID, "w"))
	assert.Equal(t, map[string]bool{b.ID: true}, s.Membership("w"))
	require.NoError(t, s.ClearList(b.ID))
	assert.Empty(t, s.Membership("w"))

	mustToggleIn(t, s, a.ID, "w")
	require.NoError(t, s.DeleteList(a.ID))
	assert.Empty(t, s.Membership("w"), "a deleted list holds nothing")
}

// ---- failures -----------------------------------------------------------------

func TestNoPathFailsEveryListCall(t *testing.T) {
	s := Open("")
	_, err := s.Summaries()
	assert.ErrorIs(t, err, errNoPath)
	_, err = s.CreateList("x")
	assert.ErrorIs(t, err, errNoPath)
	assert.ErrorIs(t, s.RenameList("x", "y"), errNoPath)
	assert.ErrorIs(t, s.DeleteList("x"), errNoPath)
	_, err = s.Entries(FavoritesID)
	assert.ErrorIs(t, err, errNoPath)
	_, err = s.ToggleIn(FavoritesID, Entry{Key: "k"})
	assert.ErrorIs(t, err, errNoPath)
	assert.ErrorIs(t, s.RemoveFrom(FavoritesID, "k"), errNoPath)
	assert.ErrorIs(t, s.ClearList(FavoritesID), errNoPath)
	assert.Empty(t, s.Membership("k"))
	assert.NotNil(t, s.Membership("k"))
}

func TestListCallsOnADeletedAutoListIDAreChecked(t *testing.T) {
	// the automatic-list errors win over the file: they need no read.
	s := Open("")
	assert.ErrorIs(t, s.RenameList(FavoritesID, "x"), ErrAutoList)
	assert.ErrorIs(t, s.DeleteList(FavoritesID), ErrAutoList)
}

func TestAReadOnlyDirectoryFailsListWritesCleanlyAndKeepsTheFile(t *testing.T) {
	skipIfNoPermissions(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.json")
	s := Open(path)
	a := mkList(t, s, "A")
	before := readBytes(t, path)
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	_, err := s.CreateList("B")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), dir, "no path on a screen")
	assert.NotContains(t, err.Error(), ".tmp")
	require.Error(t, s.RenameList(a.ID, "C"))
	require.Error(t, s.DeleteList(a.ID))
	_, err = s.ToggleIn(a.ID, Entry{Key: "k"})
	require.Error(t, err)
	require.Error(t, s.ClearList(FavoritesID))

	require.NoError(t, os.Chmod(dir, 0o700))
	assert.Equal(t, before, readBytes(t, path), "the old file is intact")
	assert.Equal(t, []string{"Favorites", "A"}, namesOf(summaries(t, s)))
	assert.Equal(t, []string{"favorites.json"}, dirNames(t, dir), "no temp file is left")
}

func TestAnUnreadableFileFailsListCallsAndIsNotOverwritten(t *testing.T) {
	skipIfNoPermissions(t)
	path := pathOf(t)
	s := Open(path)
	a := mkList(t, s, "A")
	before := readBytes(t, path)
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err := s.Summaries()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), path)
	_, err = s.CreateList("B")
	require.Error(t, err)
	assert.Empty(t, s.Membership("k"))
	_, err = s.Entries(a.ID)
	require.Error(t, err)
	require.NoError(t, os.Chmod(path, 0o600))
	assert.Equal(t, before, readBytes(t, path))
}

func TestConcurrentListWritersLoseNothing(t *testing.T) {
	s := newStore(t)
	const workers = 8
	done := make(chan error, workers)
	for w := range workers {
		go func() {
			sum, err := s.CreateList("w" + strconv.Itoa(w))
			if err == nil {
				for i := range 5 {
					_, err = s.ToggleIn(sum.ID, Entry{Key: "k" + strconv.Itoa(i)})
					if err != nil {
						break
					}
				}
			}
			done <- err
		}()
	}
	for range workers {
		require.NoError(t, <-done)
	}
	got := summaries(t, s)
	require.Len(t, got, workers+1)
	ids := map[string]bool{}
	for _, sum := range got[1:] {
		assert.Equal(t, 5, sum.Count)
		ids[sum.ID] = true
	}
	assert.Len(t, ids, workers, "the IDs are all different")
}
