package tui

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeListStore is a listStore that says what it is told and writes down what
// it is asked. It is not the real store: these tests are about the adapter.
type fakeListStore struct {
	sums    []favorites.Summary
	entries []favorites.Entry
	created favorites.Summary
	on      bool
	member  map[string]bool
	err     error

	gotID    string
	gotName  string
	gotKey   string
	gotEntry favorites.Entry
	calls    []string
}

func (f *fakeListStore) Summaries() ([]favorites.Summary, error) {
	f.calls = append(f.calls, "Summaries")
	return f.sums, f.err
}

func (f *fakeListStore) CreateList(name string) (favorites.Summary, error) {
	f.calls = append(f.calls, "CreateList")
	f.gotName = name
	return f.created, f.err
}

func (f *fakeListStore) RenameList(id, name string) error {
	f.calls = append(f.calls, "RenameList")
	f.gotID, f.gotName = id, name
	return f.err
}

func (f *fakeListStore) DeleteList(id string) error {
	f.calls = append(f.calls, "DeleteList")
	f.gotID = id
	return f.err
}

func (f *fakeListStore) Entries(id string) ([]favorites.Entry, error) {
	f.calls = append(f.calls, "Entries")
	f.gotID = id
	return f.entries, f.err
}

func (f *fakeListStore) ToggleIn(id string, e favorites.Entry) (bool, error) {
	f.calls = append(f.calls, "ToggleIn")
	f.gotID, f.gotEntry = id, e
	return f.on, f.err
}

func (f *fakeListStore) RemoveFrom(id, key string) error {
	f.calls = append(f.calls, "RemoveFrom")
	f.gotID, f.gotKey = id, key
	return f.err
}

func (f *fakeListStore) ClearList(id string) error {
	f.calls = append(f.calls, "ClearList")
	f.gotID = id
	return f.err
}

func (f *fakeListStore) Membership(key string) map[string]bool {
	f.calls = append(f.calls, "Membership")
	f.gotKey = key
	return f.member
}

var errStore = errors.New("you already have a list with that name")

func TestStoreListsMapsEverySummaryField(t *testing.T) {
	made := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	f := &fakeListStore{sums: []favorites.Summary{
		{ID: "favorites", Name: "Favorites", Auto: true, Count: 3, Titles: []string{"C", "B", "A"}},
		{ID: "l1", Name: "Weekend", CreatedAt: made, Count: 2, Titles: []string{"Y", "X"}},
	}}

	got, err := StoreLists(f).Lists()

	require.NoError(t, err)
	assert.Equal(t, []HomeList{
		{ID: "favorites", Name: "Favorites", Auto: true, Count: 3, Titles: []string{"C", "B", "A"}},
		{ID: "l1", Name: "Weekend", Created: made, Count: 2, Titles: []string{"Y", "X"}},
	}, got)
}

func TestStoreListsPutsTheAutomaticListFirstAndKeepsTheRestInOrder(t *testing.T) {
	f := &fakeListStore{sums: []favorites.Summary{
		{ID: "b", Name: "B"}, {ID: "a", Name: "A"}, {ID: "favorites", Name: "Favorites", Auto: true}, {ID: "c", Name: "C"},
	}}

	got, err := StoreLists(f).Lists()

	require.NoError(t, err)
	ids := make([]string, 0, len(got))
	for _, l := range got {
		ids = append(ids, l.ID)
	}
	assert.Equal(t, []string{"favorites", "b", "a", "c"}, ids, "Favorites leads; the others keep the store's order")
}

func TestStoreListsCopiesTheTitlesSoTheScreenCannotChangeTheStore(t *testing.T) {
	f := &fakeListStore{sums: []favorites.Summary{{ID: "l1", Titles: []string{"One", "Two"}}}}
	got, err := StoreLists(f).Lists()
	require.NoError(t, err)

	got[0].Titles[0] = "changed"

	assert.Equal(t, "One", f.sums[0].Titles[0])
}

func TestStoreListsNoListsIsAnEmptySliceNotAnError(t *testing.T) {
	got, err := StoreLists(&fakeListStore{}).Lists()
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestStoreListsListsErrorPassesThroughUnchanged(t *testing.T) {
	f := &fakeListStore{err: errStore, sums: []favorites.Summary{{ID: "x"}}}
	got, err := StoreLists(f).Lists()
	assert.Same(t, errStore, err)
	assert.Nil(t, got, "no half answer next to an error")
}

func TestStoreListsWorksAreNewestFirstWhateverOrderTheStoreUses(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &fakeListStore{entries: []favorites.Entry{
		{Key: "old", Title: "Old", AddedAt: base},
		{Key: "new", Title: "New", AddedAt: base.Add(2 * time.Hour)},
		{Key: "mid", Title: "Mid", AddedAt: base.Add(time.Hour)},
	}}

	got, err := StoreLists(f).Works("l1")

	require.NoError(t, err)
	keys := make([]string, 0, len(got))
	for _, w := range got {
		keys = append(keys, w.Key)
	}
	assert.Equal(t, []string{"new", "mid", "old"}, keys)
	assert.Equal(t, "l1", f.gotID, "the list asked for is the list the store is asked")
	assert.Equal(t, "old", f.entries[0].Key, "the store's own slice is not reordered")
}

func TestStoreListsWorksWithEqualTimesKeepTheStoresOrder(t *testing.T) {
	f := &fakeListStore{entries: []favorites.Entry{{Key: "a"}, {Key: "b"}, {Key: "c"}}}
	got, err := StoreLists(f).Works("favorites")
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, []string{"a", "b", "c"}, []string{got[0].Key, got[1].Key, got[2].Key})
}

func TestStoreListsWorkMapsEveryEntryFieldOneToOne(t *testing.T) {
	f := &fakeListStore{entries: []favorites.Entry{{
		Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime", AddedAt: time.Now(),
	}}}

	got, err := StoreLists(f).Works("l1")

	require.NoError(t, err)
	assert.Equal(t, []HomeFavorite{{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"}}, got)
}

func TestStoreListsWorksEmptyAndError(t *testing.T) {
	got, err := StoreLists(&fakeListStore{}).Works("l1")
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = StoreLists(&fakeListStore{err: favorites.ErrNoSuchList, entries: []favorites.Entry{{Key: "k"}}}).Works("gone")
	assert.Same(t, favorites.ErrNoSuchList, err)
	assert.Nil(t, got)
}

func TestStoreListsCreateMapsTheNewListAndHandsTheNameOver(t *testing.T) {
	made := time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)
	f := &fakeListStore{created: favorites.Summary{ID: "l7", Name: "Cozy  nights", CreatedAt: made}}

	got, err := StoreLists(f).Create("  Cozy  nights ")

	require.NoError(t, err)
	assert.Equal(t, HomeList{ID: "l7", Name: "Cozy  nights", Created: made}, got)
	assert.Equal(t, "  Cozy  nights ", f.gotName, "trimming and collapsing is the store's job, not the adapter's")
}

func TestStoreListsCreateErrorPassesThroughAndReturnsNoList(t *testing.T) {
	f := &fakeListStore{err: favorites.ErrNameTaken, created: favorites.Summary{ID: "l7", Name: "x"}}
	got, err := StoreLists(f).Create("x")
	assert.Same(t, favorites.ErrNameTaken, err)
	assert.Equal(t, HomeList{}, got)
}

func TestStoreListsRenameDeleteRemoveClearReachTheStoreWithTheirArguments(t *testing.T) {
	f := &fakeListStore{}
	lists := StoreLists(f)

	require.NoError(t, lists.Rename("l1", "New name"))
	assert.Equal(t, []string{"RenameList"}, f.calls)
	assert.Equal(t, "l1", f.gotID)
	assert.Equal(t, "New name", f.gotName)

	f.calls = nil
	require.NoError(t, lists.Delete("l2"))
	assert.Equal(t, []string{"DeleteList"}, f.calls)
	assert.Equal(t, "l2", f.gotID)

	f.calls = nil
	require.NoError(t, lists.RemoveWork("l3", "frieren"))
	assert.Equal(t, []string{"RemoveFrom"}, f.calls)
	assert.Equal(t, "l3", f.gotID)
	assert.Equal(t, "frieren", f.gotKey)

	f.calls = nil
	require.NoError(t, lists.Clear("l4"))
	assert.Equal(t, []string{"ClearList"}, f.calls)
	assert.Equal(t, "l4", f.gotID)
}

func TestStoreListsMutationsPassErrorsThroughUnchanged(t *testing.T) {
	tests := []struct {
		name string
		call func(FavoriteLists) error
	}{
		{"rename", func(l FavoriteLists) error { return l.Rename("a", "b") }},
		{"delete", func(l FavoriteLists) error { return l.Delete("a") }},
		{"remove", func(l FavoriteLists) error { return l.RemoveWork("a", "k") }},
		{"clear", func(l FavoriteLists) error { return l.Clear("a") }},
		{"toggle", func(l FavoriteLists) error { _, err := l.ToggleWork("a", HomeFavorite{Title: "T"}); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(StoreLists(&fakeListStore{err: favorites.ErrAutoList}))
			assert.Same(t, favorites.ErrAutoList, err)
			assert.Equal(t, "The Favorites list cannot be renamed or deleted.", friendly(err), "the screens can make it a sentence")
		})
	}
}

func TestStoreListsToggleWorkMapsEveryFieldAndStampsTheTime(t *testing.T) {
	stamp := time.Date(2026, 9, 9, 9, 9, 9, 0, time.UTC)
	f := &fakeListStore{on: true}
	lists := StoreLists(f)
	lists.(*storeLists).now = func() time.Time { return stamp }

	on, err := lists.ToggleWork("l1", HomeFavorite{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"})

	require.NoError(t, err)
	assert.True(t, on)
	assert.Equal(t, "l1", f.gotID)
	assert.Equal(t, favorites.Entry{
		Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "HiAnime", AddedAt: stamp,
	}, f.gotEntry)
}

func TestStoreListsToggleWorkReportsWhatTheStoreSays(t *testing.T) {
	f := &fakeListStore{on: false}
	on, err := StoreLists(f).ToggleWork("l1", HomeFavorite{Key: "k", Title: "T"})
	require.NoError(t, err)
	assert.False(t, on)
	f.on = true
	on, err = StoreLists(f).ToggleWork("l1", HomeFavorite{Key: "k", Title: "T"})
	require.NoError(t, err)
	assert.True(t, on)
}

func TestStoreListsToggleWorkStampsTheRealClockByDefault(t *testing.T) {
	f := &fakeListStore{}
	before := time.Now()
	_, err := StoreLists(f).ToggleWork("l1", HomeFavorite{Key: "k", Title: "T"})
	require.NoError(t, err)
	assert.False(t, f.gotEntry.AddedAt.Before(before), "a work added now is newest")
	assert.WithinDuration(t, time.Now(), f.gotEntry.AddedAt, time.Minute)
}

func TestStoreListsToggleWorkNormalizesAMissingKeyFromTheTitle(t *testing.T) {
	tests := []struct{ name, key, title, want string }{
		{"empty key, tagged title", "", "[PT-BR] Frieren: Beyond Journey's End!", "frieren beyond journeys end"},
		{"blank key", "   ", "Naruto  Shippuden", "naruto shippuden"},
		{"tab key", "\t", "Pokémon", "pokémon"},
		{"given key is kept as is", "My-Key", "Other Title", "My-Key"},
		{"nothing to normalize", "", "!!!", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeListStore{}
			_, err := StoreLists(f).ToggleWork("l1", HomeFavorite{Key: tt.key, Title: tt.title})
			require.NoError(t, err)
			assert.Equal(t, tt.want, f.gotEntry.Key)
			assert.Equal(t, tt.title, f.gotEntry.Title, "the title itself is not changed")
		})
	}
}

func TestStoreListsMembershipCopiesAndDropsFalseEntries(t *testing.T) {
	f := &fakeListStore{member: map[string]bool{"favorites": true, "l1": false, "l2": true}}

	got := StoreLists(f).Membership("frieren")

	assert.Equal(t, map[string]bool{"favorites": true, "l2": true}, got)
	assert.Equal(t, "frieren", f.gotKey)
	got["l9"] = true
	assert.NotContains(t, f.member, "l9", "the screen's map is its own")
}

func TestStoreListsMembershipOfNothingIsAnEmptyMapNotNil(t *testing.T) {
	got := StoreLists(&fakeListStore{member: nil}).Membership("k")
	require.NotNil(t, got)
	assert.Empty(t, got)
	got["x"] = true // must not panic
}

func TestStoreListsKeepWeirdTitlesWhole(t *testing.T) {
	weird := []string{
		"日本語のタイトル", "Pokémon: Mewtwo Strikes Back", "Emoji 🎬 title", "Tab\there", "New\nline",
		"Esc\x1b[31mred", "  padded  ", strings.Repeat("long ", 200), "Ünïcödé ﬁ ligature", "",
	}
	for _, title := range weird {
		f := &fakeListStore{
			sums:    []favorites.Summary{{ID: "l1", Name: title, Titles: []string{title}}},
			entries: []favorites.Entry{{Key: "k", Title: title, Source: title, Media: title, Year: title}},
		}
		lists := StoreLists(f)

		got, err := lists.Lists()
		require.NoError(t, err)
		assert.Equal(t, title, got[0].Name)
		assert.Equal(t, []string{title}, got[0].Titles)

		works, err := lists.Works("l1")
		require.NoError(t, err)
		assert.Equal(t, HomeFavorite{Key: "k", Title: title, Source: title, Media: title, Year: title}, works[0],
			"sanitising is the screens' job; the adapter hands over what is saved")
	}
}

func TestStoreListsWithoutAStoreFailsPolitelyEverywhere(t *testing.T) {
	var nilConcrete *favorites.Store
	for name, lists := range map[string]FavoriteLists{
		"nil interface":     StoreLists(nil),
		"nil *Store inside": StoreLists(nilConcrete),
	} {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, lists)
			const want = "your lists are not available"

			l, err := lists.Lists()
			assert.EqualError(t, err, want)
			assert.Nil(t, l)

			w, err := lists.Works("favorites")
			assert.EqualError(t, err, want)
			assert.Nil(t, w)

			made, err := lists.Create("x")
			assert.EqualError(t, err, want)
			assert.Equal(t, HomeList{}, made)

			assert.EqualError(t, lists.Rename("a", "b"), want)
			assert.EqualError(t, lists.Delete("a"), want)
			assert.EqualError(t, lists.RemoveWork("a", "k"), want)
			assert.EqualError(t, lists.Clear("a"), want)

			on, err := lists.ToggleWork("a", HomeFavorite{Title: "T"})
			assert.False(t, on)
			assert.EqualError(t, err, want)

			m := lists.Membership("k")
			require.NotNil(t, m)
			assert.Empty(t, m)
			assert.Equal(t, "Your lists are not available.", friendly(err))
		})
	}
}

func TestStoreListsRealStoreSatisfiesTheInterface(t *testing.T) {
	var _ listStore = (*favorites.Store)(nil)
}

// A stateful in-memory listStore: the cascade of a list's life through the
// adapter, without the real file.
type memListStore struct {
	next  int
	lists []*memList
}

type memList struct {
	id, name string
	made     time.Time
	entries  []favorites.Entry
}

func (m *memListStore) find(id string) *memList {
	for _, l := range m.lists {
		if l.id == id {
			return l
		}
	}
	return nil
}

func (m *memListStore) Summaries() ([]favorites.Summary, error) {
	out := []favorites.Summary{{ID: favorites.FavoritesID, Name: "Favorites", Auto: true}}
	for _, l := range m.lists {
		s := favorites.Summary{ID: l.id, Name: l.name, CreatedAt: l.made, Count: len(l.entries)}
		for i := len(l.entries) - 1; i >= 0; i-- {
			s.Titles = append(s.Titles, l.entries[i].Title)
		}
		out = append(out, s)
	}
	return out, nil
}

func (m *memListStore) CreateList(name string) (favorites.Summary, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return favorites.Summary{}, favorites.ErrNameEmpty
	}
	m.next++
	l := &memList{id: "l" + string(rune('0'+m.next)), name: name, made: time.Date(2026, 1, m.next, 0, 0, 0, 0, time.UTC)}
	m.lists = append(m.lists, l)
	return favorites.Summary{ID: l.id, Name: l.name, CreatedAt: l.made}, nil
}

func (m *memListStore) RenameList(id, name string) error {
	l := m.find(id)
	if l == nil {
		return favorites.ErrNoSuchList
	}
	l.name = name
	return nil
}

func (m *memListStore) DeleteList(id string) error {
	for i, l := range m.lists {
		if l.id == id {
			m.lists = append(m.lists[:i], m.lists[i+1:]...)
			return nil
		}
	}
	return favorites.ErrNoSuchList
}

func (m *memListStore) Entries(id string) ([]favorites.Entry, error) {
	l := m.find(id)
	if l == nil {
		return nil, favorites.ErrNoSuchList
	}
	return append([]favorites.Entry(nil), l.entries...), nil
}

func (m *memListStore) ToggleIn(id string, e favorites.Entry) (bool, error) {
	l := m.find(id)
	if l == nil {
		return false, favorites.ErrNoSuchList
	}
	for i := range l.entries {
		if l.entries[i].Key == e.Key {
			l.entries = append(l.entries[:i], l.entries[i+1:]...)
			return false, nil
		}
	}
	l.entries = append(l.entries, e)
	return true, nil
}

func (m *memListStore) RemoveFrom(id, key string) error {
	l := m.find(id)
	if l == nil {
		return favorites.ErrNoSuchList
	}
	kept := l.entries[:0]
	for _, e := range l.entries {
		if e.Key != key {
			kept = append(kept, e)
		}
	}
	l.entries = kept
	return nil
}

func (m *memListStore) ClearList(id string) error {
	l := m.find(id)
	if l == nil {
		return favorites.ErrNoSuchList
	}
	l.entries = nil
	return nil
}

func (m *memListStore) Membership(key string) map[string]bool {
	out := map[string]bool{}
	for _, l := range m.lists {
		for _, e := range l.entries {
			if e.Key == key {
				out[l.id] = true
			}
		}
	}
	return out
}

func TestStoreListsLifeOfAListThroughTheAdapter(t *testing.T) {
	lists := StoreLists(&memListStore{})
	ticks := 0
	lists.(*storeLists).now = func() time.Time {
		ticks++
		return time.Date(2026, 2, 1, 0, ticks, 0, 0, time.UTC)
	}

	// 1 a new list appears after Favorites, empty.
	made, err := lists.Create("  Weekend   watch ")
	require.NoError(t, err)
	assert.Equal(t, "Weekend watch", made.Name)
	all, err := lists.Lists()
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.True(t, all[0].Auto)
	assert.Equal(t, made.ID, all[1].ID)
	assert.Equal(t, 0, all[1].Count)

	// 2 two works go in; they come out newest first; the count and titles follow.
	_, err = lists.ToggleWork(made.ID, HomeFavorite{Title: "[PT-BR] Frieren", Year: "2023", Media: "Anime", Source: "HiAnime"})
	require.NoError(t, err)
	on, err := lists.ToggleWork(made.ID, HomeFavorite{Title: "Naruto", Year: "2002", Media: "Anime"})
	require.NoError(t, err)
	assert.True(t, on)
	works, err := lists.Works(made.ID)
	require.NoError(t, err)
	require.Len(t, works, 2)
	assert.Equal(t, "Naruto", works[0].Title, "newest first")
	assert.Equal(t, "frieren", works[1].Key, "a missing key is made from the title")
	all, _ = lists.Lists()
	assert.Equal(t, 2, all[1].Count)
	assert.Equal(t, []string{"Naruto", "[PT-BR] Frieren"}, all[1].Titles)

	// 3 membership says where a work is; toggling again takes it off.
	assert.Equal(t, map[string]bool{made.ID: true}, lists.Membership("naruto"))
	on, err = lists.ToggleWork(made.ID, HomeFavorite{Title: "Naruto"})
	require.NoError(t, err)
	assert.False(t, on)
	assert.Empty(t, lists.Membership("naruto"))

	// 4 rename, remove, clear, delete.
	require.NoError(t, lists.Rename(made.ID, "Cozy"))
	all, _ = lists.Lists()
	assert.Equal(t, "Cozy", all[1].Name)
	require.NoError(t, lists.RemoveWork(made.ID, "frieren"))
	works, _ = lists.Works(made.ID)
	assert.Empty(t, works)
	_, err = lists.ToggleWork(made.ID, HomeFavorite{Title: "Bleach"})
	require.NoError(t, err)
	require.NoError(t, lists.Clear(made.ID))
	works, _ = lists.Works(made.ID)
	assert.Empty(t, works)
	require.NoError(t, lists.Delete(made.ID))
	all, _ = lists.Lists()
	assert.Len(t, all, 1)

	// 5 a list that is gone says so, and the screens can word it.
	_, err = lists.Works(made.ID)
	require.ErrorIs(t, err, favorites.ErrNoSuchList)
	assert.Equal(t, "That list no longer exists.", friendly(err))
}

func TestStoreListsCountIsTheStoresNotTheNumberOfTitles(t *testing.T) {
	// The store may send only a few titles as a preview of a long list.
	f := &fakeListStore{sums: []favorites.Summary{{ID: "l1", Count: 40, Titles: []string{"A", "B", "C"}}}}

	got, err := StoreLists(f).Lists()

	require.NoError(t, err)
	assert.Equal(t, 40, got[0].Count)
	assert.Len(t, got[0].Titles, 3)
}

func TestStoreListsWorksKeepTheStoresOrderAmongManyEqualTimes(t *testing.T) {
	// A sort that is not stable shuffles equal times once the list is long.
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var entries []favorites.Entry
	for i := range 200 {
		entries = append(entries, favorites.Entry{
			Key: "k" + strconv.Itoa(i), AddedAt: base.Add(time.Duration(i%3) * time.Hour),
		})
	}
	got, err := StoreLists(&fakeListStore{entries: entries}).Works("l1")
	require.NoError(t, err)

	want := make([]string, 0, len(entries))
	for _, bucket := range []int{2, 1, 0} {
		for i := range entries {
			if i%3 == bucket {
				want = append(want, "k"+strconv.Itoa(i))
			}
		}
	}
	have := make([]string, 0, len(got))
	for _, w := range got {
		have = append(have, w.Key)
	}
	assert.Equal(t, want, have)
}

func TestHomeWithListsSetStillOpensTheSearchTab(t *testing.T) {
	// The Lists option must not change what the SEARCH tab shows.
	continueAll, recent := homeTestEntries()
	lists := StoreLists(&memListStore{})

	with := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent, Lists: lists})
	without := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})

	assert.Equal(t, len(without.rows), len(with.rows))
	assert.Equal(t, without.focusInput, with.focusInput)
}
