package favorites

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// modelStore is the specification: a map from normalized key to entry.
type modelStore map[string]Entry

func (m modelStore) sortedKeys() []string {
	entries := make([]Entry, 0, len(m))
	for _, e := range m {
		entries = append(entries, e)
	}
	// newest first; every AddedAt in the test is unique
	slices.SortFunc(entries, func(a, b Entry) int { return b.AddedAt.Compare(a.AddedAt) })
	return keysOf(entries)
}

// spell gives the same work in different spellings, to exercise normalization.
func spell(r *rand.Rand, base string) string {
	switch r.IntN(5) {
	case 0:
		return strings.ToUpper(base)
	case 1:
		return "  " + base + " \t"
	case 2:
		return strings.ToUpper(base[:1]) + base[1:]
	default:
		return base
	}
}

// TestStoreMatchesAMapModel runs random Toggle/Remove/Clear/Has/List
// sequences against several Store instances that share one file (used one at
// a time, like successive runs of the program) and a plain map.
func TestStoreMatchesAMapModel(t *testing.T) {
	pool := make([]string, 12)
	for i := range pool {
		pool[i] = "work " + strconv.Itoa(i)
	}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seeds := uint64(12)
	if testing.Short() {
		seeds = 4
	}
	for seed := uint64(1); seed <= seeds; seed++ {
		t.Run("seed"+strconv.FormatUint(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewPCG(seed, seed*7919))
			path := filepath.Join(t.TempDir(), "d", "favorites.json")
			stores := []*Store{Open(path), Open(path), Open(path)}
			model := modelStore{}
			tick := 0

			for step := range 400 {
				s := stores[r.IntN(len(stores))]
				key := pool[r.IntN(len(pool))]
				switch op := r.IntN(12); {
				case op < 5: // Toggle
					tick++
					e := Entry{Key: spell(r, key), Title: "T " + key, Year: strconv.Itoa(2000 + r.IntN(25)),
						AddedAt: base.Add(time.Duration(tick) * time.Minute)}
					_, in := model[key]
					got, err := s.Toggle(e)
					require.NoError(t, err, "step %d", step)
					require.Equal(t, !in, got, "step %d Toggle(%q)", step, e.Key)
					if in {
						delete(model, key)
					} else {
						e.Key = key
						model[key] = e
					}
				case op < 7: // Remove
					require.NoError(t, s.Remove(spell(r, key)), "step %d", step)
					delete(model, key)
				case op == 7: // Clear (rarely)
					if r.IntN(4) == 0 {
						require.NoError(t, s.Clear(), "step %d", step)
						clear(model)
					}
				case op < 10: // Has
					_, want := model[key]
					require.Equal(t, want, s.Has(spell(r, key)), "step %d Has(%q)", step, key)
				default: // List
					list, err := s.List()
					require.NoError(t, err, "step %d", step)
					require.Equal(t, model.sortedKeys(), keysOf(list), "step %d List order", step)
					for _, e := range list {
						require.Equal(t, model[e.Key].Title, e.Title, "step %d fields survive", step)
						require.Equal(t, model[e.Key].Year, e.Year)
						require.True(t, model[e.Key].AddedAt.Equal(e.AddedAt))
					}
				}
			}

			// the whole thing, once more, from a store that never saw an op
			list, err := Open(path).List()
			require.NoError(t, err)
			assert.Equal(t, model.sortedKeys(), keysOf(list))
			_, statErr := os.Stat(path + ".corrupt")
			assert.ErrorIs(t, statErr, os.ErrNotExist)
			assert.Equal(t, []string{"favorites.json"}, dirNames(t, filepath.Dir(path)))
		})
	}
}

// TestCapHoldsUnderRandomInsertion adds more than the cap in random date order
// and checks the list never exceeds it and ends sorted.
func TestCapHoldsUnderRandomInsertion(t *testing.T) {
	r := rand.New(rand.NewPCG(42, 43))
	s := newStore(t)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	total := maxEntries + 40
	for _, i := range r.Perm(total) {
		_, err := s.Toggle(Entry{Key: "k" + strconv.Itoa(i), Title: "t", AddedAt: base.Add(time.Duration(i) * time.Second)})
		require.NoError(t, err)
		require.LessOrEqual(t, len(mustList(t, s)), maxEntries)
	}
	list := mustList(t, s)
	require.Len(t, list, maxEntries)
	assert.True(t, slices.IsSortedFunc(list, func(a, b Entry) int { return b.AddedAt.Compare(a.AddedAt) }))
	// Nothing newer than the oldest survivor was dropped.
	oldest := list[len(list)-1].AddedAt
	for i := range total {
		if base.Add(time.Duration(i) * time.Second).After(oldest) {
			assert.True(t, s.Has("k"+strconv.Itoa(i)), "k%d is newer than the oldest survivor", i)
		}
	}
}
