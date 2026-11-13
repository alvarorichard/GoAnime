package favorites

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The model is the specification, written without the code under test: a
// name pool where each name is labelled by hand with what it must do, and
// plain slices for the lists.

type nameCase struct {
	raw   string
	kind  string // "ok", "empty", "long", "reserved"
	canon string // two ok names are the same name when their canon is equal
}

var namePool = []nameCase{
	{"Weekend", "ok", "weekend"},
	{"weekend  ", "ok", "weekend"},
	{" WEEKÉND", "ok", "weekend"},
	{"Familia", "ok", "familia"},
	{"Família", "ok", "familia"},
	{"Maratón de anime", "ok", "maraton de anime"},
	{"maraton  DE anime", "ok", "maraton de anime"},
	{"F A V", "ok", "f a v"},
	{"Favorites 2", "ok", "favorites 2"},
	{"🎬 night", "ok", "🎬 night"},
	{strings.Repeat("é", 40), "ok", strings.Repeat("e", 40)},
	{strings.Repeat("e", 40), "ok", strings.Repeat("e", 40)},
	{"", "empty", ""},
	{"  \t ", "empty", ""},
	{"favorites", "reserved", ""},
	{"Fávorites", "reserved", ""},
	{" FAVORITES ", "reserved", ""},
	{strings.Repeat("é", 41), "long", ""},
	{strings.Repeat("🎬", 41), "long", ""},
}

type mList struct {
	id, name, canon string
	keys            []string // oldest first: every stamp comes from one rising clock
}

type model struct {
	fav   []string
	lists []*mList
}

func (m *model) find(id string) *mList {
	for _, l := range m.lists {
		if l.id == id {
			return l
		}
	}
	return nil
}

// keysFor returns the slice that holds a list's keys.
func (m *model) keysFor(id string) (*[]string, error) {
	if id == FavoritesID {
		return &m.fav, nil
	}
	if l := m.find(id); l != nil {
		return &l.keys, nil
	}
	return nil, ErrNoSuchList
}

// checkName is the expected result of validating a name for the list except.
func (m *model) checkName(c nameCase, except string) error {
	switch c.kind {
	case "empty":
		return ErrNameEmpty
	case "long":
		return ErrNameTooLong
	case "reserved":
		return ErrReservedName
	}
	for _, l := range m.lists {
		if l.id != except && l.canon == c.canon {
			return ErrNameTaken
		}
	}
	return nil
}

func newestFirst(keys []string) []string {
	out := append([]string{}, keys...)
	slices.Reverse(out)
	return out
}

func TestListsMatchAnInMemoryModel(t *testing.T) {
	keyPool := make([]string, 8)
	for i := range keyPool {
		keyPool[i] = "work " + strconv.Itoa(i)
	}
	seeds := uint64(10)
	if testing.Short() {
		seeds = 5
	}
	steps := 260
	for seed := uint64(1); seed <= seeds; seed++ {
		t.Run("seed"+strconv.FormatUint(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewPCG(seed, seed*104729))
			path := pathOf(t)
			tick := 0
			clock := func() time.Time { tick++; return t0.Add(time.Duration(tick) * time.Second) }
			stores := make([]*Store, 3)
			for i := range stores {
				stores[i] = Open(path)
				stores[i].now = clock
			}
			createHeavy := seed%3 == 0 // these seeds climb to the 50-list limit
			m := &model{}
			genN := 0

			for step := range steps {
				s := stores[r.IntN(len(stores))]
				ids := []string{FavoritesID, "ghost"}
				for _, l := range m.lists {
					ids = append(ids, l.id)
				}
				anyID := func() string { return ids[r.IntN(len(ids))] }
				userID := func() string {
					if len(m.lists) > 0 && r.IntN(8) != 0 {
						return m.lists[r.IntN(len(m.lists))].id
					}
					return []string{FavoritesID, "ghost"}[r.IntN(2)]
				}
				key := keyPool[r.IntN(len(keyPool))]
				spelled := key
				switch r.IntN(4) {
				case 0:
					spelled = strings.ToUpper(key)
				case 1:
					spelled = "  " + key + "\t"
				}
				label := "seed " + strconv.FormatUint(seed, 10) + " step " + strconv.Itoa(step)

				op := r.IntN(100)
				if createHeavy && op < 30 {
					op = 0 // a create
				}
				switch {
				case op < 14: // CreateList
					c := namePool[r.IntN(len(namePool))]
					if r.IntN(3) == 0 {
						genN++
						c = nameCase{"gen " + strconv.Itoa(genN), "ok", "gen " + strconv.Itoa(genN)}
					}
					want := m.checkName(c, "")
					if want == nil && len(m.lists) >= maxLists {
						want = ErrTooManyLists
					}
					sum, err := s.CreateList(c.raw)
					if want != nil {
						require.ErrorIs(t, err, want, label+" create %q", c.raw)
						break
					}
					require.NoError(t, err, label)
					assert.Equal(t, strings.Join(strings.Fields(c.raw), " "), sum.Name, label)
					m.lists = append(m.lists, &mList{id: sum.ID, name: sum.Name, canon: c.canon})
				case op < 24: // RenameList
					id := userID()
					c := namePool[r.IntN(len(namePool))]
					var want error
					switch l := m.find(id); {
					case id == FavoritesID:
						want = ErrAutoList
					case l == nil:
						want = ErrNoSuchList
					default:
						want = m.checkName(c, id)
					}
					err := s.RenameList(id, c.raw)
					if want != nil {
						require.ErrorIs(t, err, want, label+" rename %q", c.raw)
						break
					}
					require.NoError(t, err, label)
					l := m.find(id)
					l.name, l.canon = strings.Join(strings.Fields(c.raw), " "), c.canon
				case op < 32: // DeleteList
					id := userID()
					var want error
					switch {
					case id == FavoritesID:
						want = ErrAutoList
					case m.find(id) == nil:
						want = ErrNoSuchList
					}
					err := s.DeleteList(id)
					if want != nil {
						require.ErrorIs(t, err, want, label)
						break
					}
					require.NoError(t, err, label)
					m.lists = slices.DeleteFunc(m.lists, func(l *mList) bool { return l.id == id })
				case op < 66: // ToggleIn / Toggle
					id := anyID()
					keys, want := m.keysFor(id)
					var on bool
					var err error
					if id == FavoritesID && r.IntN(2) == 0 {
						on, err = s.Toggle(Entry{Key: spelled, Title: "T:" + key})
					} else {
						on, err = s.ToggleIn(id, Entry{Key: spelled, Title: "T:" + key})
					}
					if want != nil {
						require.ErrorIs(t, err, want, label)
						break
					}
					require.NoError(t, err, label)
					if i := slices.Index(*keys, key); i >= 0 {
						*keys = slices.Delete(*keys, i, i+1)
						assert.False(t, on, label)
					} else {
						*keys = append(*keys, key)
						assert.True(t, on, label)
					}
				case op < 80: // RemoveFrom / Remove
					id := anyID()
					keys, want := m.keysFor(id)
					var err error
					if id == FavoritesID && r.IntN(2) == 0 {
						err = s.Remove(spelled)
					} else {
						err = s.RemoveFrom(id, spelled)
					}
					if want != nil {
						require.ErrorIs(t, err, want, label)
						break
					}
					require.NoError(t, err, label)
					if i := slices.Index(*keys, key); i >= 0 {
						*keys = slices.Delete(*keys, i, i+1)
					}
				case op < 88: // ClearList / Clear
					id := anyID()
					keys, want := m.keysFor(id)
					var err error
					if id == FavoritesID && r.IntN(2) == 0 {
						err = s.Clear()
					} else {
						err = s.ClearList(id)
					}
					if want != nil {
						require.ErrorIs(t, err, want, label)
						break
					}
					require.NoError(t, err, label)
					*keys = nil
				default: // reads
					compareModel(t, s, m, keyPool, label)
				}
				if step%20 == 0 {
					compareModel(t, stores[r.IntN(len(stores))], m, keyPool, label)
				}
			}
			compareModel(t, Open(path), m, keyPool, "final, fresh store")
		})
	}
}

func compareModel(t *testing.T, s *Store, m *model, keyPool []string, label string) {
	t.Helper()
	got, err := s.Summaries()
	require.NoError(t, err, label)
	require.Len(t, got, len(m.lists)+1, label)
	assert.Equal(t, FavoritesID, got[0].ID, label)
	assert.True(t, got[0].Auto, label)
	assert.Equal(t, len(m.fav), got[0].Count, label)
	for i, l := range m.lists {
		sum := got[i+1] // the clock rises, so creation order is the model's order
		assert.Equal(t, l.id, sum.ID, label)
		assert.Equal(t, l.name, sum.Name, label)
		assert.False(t, sum.Auto, label)
		assert.Equal(t, len(l.keys), sum.Count, label)
		want := []string{}
		for _, k := range newestFirst(l.keys) {
			want = append(want, "T:"+k)
		}
		assert.Equal(t, want, sum.Titles, label)
	}
	check := func(id string, keys []string) {
		entries, err := s.Entries(id)
		require.NoError(t, err, label)
		assert.Equal(t, newestFirst(keys), append([]string{}, keysOf(entries)...), label+" "+id)
	}
	check(FavoritesID, m.fav)
	for _, l := range m.lists {
		check(l.id, l.keys)
	}
	for _, k := range keyPool {
		want := map[string]bool{}
		if slices.Contains(m.fav, k) {
			want[FavoritesID] = true
		}
		for _, l := range m.lists {
			if slices.Contains(l.keys, k) {
				want[l.id] = true
			}
		}
		assert.Equal(t, want, s.Membership(k), label+" membership of "+k)
		assert.Equal(t, slices.Contains(m.fav, k), s.Has(k), label)
	}
	assert.Len(t, s.Keys(), len(m.fav), label)
}
