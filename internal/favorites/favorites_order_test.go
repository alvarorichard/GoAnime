package favorites

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ties, raw file order and Clear against writers: survivors of the first
// mutation pass.

func TestExactlyTheCapIsNotReorderedOnDisk(t *testing.T) {
	s := newStore(t)
	base := time.Now().Add(-24 * time.Hour)
	fill(t, s, maxEntries, base)

	doc := readFileJSON(t, s.path)
	entries, ok := doc["entries"].([]any)
	require.True(t, ok)
	require.Len(t, entries, maxEntries)
	first, ok := entries[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "k0", first["key"], "under the cap the file keeps insertion order; only passing it sorts and trims")
}

func TestManyTiedDatesKeepFileOrderInTheList(t *testing.T) {
	s := newStore(t)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var all []Entry
	for i := range 400 {
		e := Entry{Key: "k" + strconv.Itoa(1000+i), Title: "t", AddedAt: base.Add(time.Duration(i%5) * time.Minute)}
		all = append(all, e)
		_, err := s.Toggle(e)
		require.NoError(t, err)
	}
	// newest first, ties in the order they were added
	var want []string
	for bucket := 4; bucket >= 0; bucket-- {
		for _, e := range all {
			if e.AddedAt.Equal(base.Add(time.Duration(bucket) * time.Minute)) {
				want = append(want, e.Key)
			}
		}
	}
	assert.Equal(t, want, keysOf(mustList(t, s)))
}

func TestClearRacingWithAddsNeverResurrectsOlderEntries(t *testing.T) {
	// Whatever the order, a Clear that really ran leaves either nothing or
	// only entries added after it. A Clear that does not hold the lock can be
	// overwritten by an add that had already read the old list.
	for round := range 60 {
		s := newStore(t)
		for i := range 30 {
			_, err := s.Toggle(Entry{Key: "old" + strconv.Itoa(i), Title: "t"})
			require.NoError(t, err)
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, err := s.Toggle(Entry{Key: "new" + strconv.Itoa(i), Title: "t"})
				assert.NoError(t, err)
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			assert.NoError(t, s.Clear())
		}()
		close(start)
		wg.Wait()
		for _, e := range mustList(t, s) {
			require.True(t, strings.HasPrefix(e.Key, "new"), "round %d: %q survived a Clear", round, e.Key)
		}
	}
}
