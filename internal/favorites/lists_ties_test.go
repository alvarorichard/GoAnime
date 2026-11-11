package favorites

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ties are where an unstable sort or a different insertion point shows: small
// slices hide it (the sort falls back to insertion sort), so these use more
// than a dozen of them.

func TestListsMadeAtTheSameInstantKeepTheOrderTheyWereMade(t *testing.T) {
	s := newStore(t)
	s.now = func() time.Time { return t0 }
	var want []string
	for i := range 3 {
		want = append(want, mkList(t, s, "list "+strconv.Itoa(i)).ID)
	}
	assert.Equal(t, append([]string{FavoritesID}, want...), idsOf(summaries(t, s)))
}

func TestManyTiedListsKeepTheFileOrder(t *testing.T) {
	path := pathOf(t)
	doc := file{Version: 1}
	want := []string{FavoritesID}
	for i := range 40 {
		id := "l" + strconv.Itoa(1000-i) // not sorted by ID, so only the file order explains it
		doc.Lists = append(doc.Lists, list{ID: id, Name: "n" + strconv.Itoa(i), CreatedAt: t0})
		want = append(want, id)
	}
	writeDoc(t, path, doc)
	assert.Equal(t, want, idsOf(summaries(t, Open(path))))
}

func TestManyTiedWorksKeepTheFileOrder(t *testing.T) {
	path := pathOf(t)
	full := list{ID: "la", Name: "A", CreatedAt: t0}
	doc := file{Version: 1}
	var want, wantTitles []string
	for i := range 40 {
		e := Entry{Key: "k" + strconv.Itoa(40-i), Title: "t" + strconv.Itoa(i), AddedAt: t0}
		doc.Entries = append(doc.Entries, e)
		full.Entries = append(full.Entries, e)
		want = append(want, e.Key)
		wantTitles = append(wantTitles, e.Title)
	}
	doc.Lists = []list{full}
	writeDoc(t, path, doc)
	s := Open(path)
	assert.Equal(t, want, entriesOf(t, s, "la"))
	assert.Equal(t, want, entriesOf(t, s, FavoritesID))
	sums := summaries(t, s)
	assert.Equal(t, wantTitles, sums[0].Titles)
	assert.Equal(t, wantTitles, sums[1].Titles)
}

func TestTheCapDropsTheOldestWhenManyAreTied(t *testing.T) {
	path := pathOf(t)
	full := list{ID: "la", Name: "A", CreatedAt: t0}
	for i := range maxEntries {
		full.Entries = append(full.Entries, Entry{Key: "k" + strconv.Itoa(i), AddedAt: t0})
	}
	writeDoc(t, path, file{Version: 1, Lists: []list{full}})
	s := Open(path)
	mustToggleInAt(t, s, "la", "extra", t0.Add(time.Hour))
	got := entriesOf(t, s, "la")
	require.Len(t, got, maxEntries)
	assert.Equal(t, "extra", got[0])
	// among equal dates the cap keeps the file order: the last one in goes
	assert.Equal(t, "k0", got[1])
	assert.Equal(t, "k498", got[maxEntries-1])
}

func TestConcurrentChangesToOneListLoseNothing(t *testing.T) {
	s := newStore(t)
	a := mkList(t, s, "Shared")
	const n = 24
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, err := s.ToggleIn(a.ID, Entry{Key: "k" + strconv.Itoa(i)})
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()
			_, err := s.ToggleIn(a.ID, Entry{Key: "gone" + strconv.Itoa(i)})
			assert.NoError(t, err)
			assert.NoError(t, s.RemoveFrom(a.ID, "gone"+strconv.Itoa(i)))
		}()
		go func() {
			defer wg.Done()
			_, err := s.CreateList("other " + strconv.Itoa(i))
			assert.NoError(t, err)
			assert.NoError(t, s.RenameList(a.ID, "Shared"))
		}()
	}
	wg.Wait()
	got := summaries(t, s)
	require.Len(t, got, n+2)
	assert.Equal(t, n, got[1].Count, "every work added to the shared list is there")
	assert.Equal(t, "Shared", got[1].Name)
	ids := map[string]bool{}
	for _, sum := range got {
		ids[sum.ID] = true
	}
	assert.Len(t, ids, n+2)
}
