package favorites

// Stability of the newest-first order, found by the mutation pass: ties must
// keep the order of the file even when there are enough entries for an
// unstable sort to shuffle them.

import (
	"encoding/json"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func regInterleaved(n int) []Entry {
	entries := make([]Entry, 0, n)
	for i := range n {
		// Three dates, interleaved: many ties, so a stable sort is the only
		// one that keeps the file's order inside each date.
		entries = append(entries, Entry{Key: "k" + strconv.Itoa(i), Title: "T", AddedAt: regDay(1 + (i*7)%3)})
	}
	return entries
}

func regStableNewestFirst(entries []Entry) []string {
	cp := append([]Entry(nil), entries...)
	sort.SliceStable(cp, func(i, j int) bool { return cp[i].AddedAt.After(cp[j].AddedAt) })
	return regKeys(cp)
}

func TestRegressionStoreListIsStableOnManyTies(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	entries := regInterleaved(60)
	data, err := json.MarshalIndent(file{Version: 1, Entries: entries}, "", "  ")
	require.NoError(t, err)
	regWrite(t, path, string(data))

	list, err := s.List()

	require.NoError(t, err)
	assert.Equal(t, regStableNewestFirst(entries), regKeys(list))
}

func TestRegressionStoreCapKeepsTheFileOrderInsideTies(t *testing.T) {
	t.Parallel()
	s, path := regStore(t)
	entries := regInterleaved(500)
	data, err := json.MarshalIndent(file{Version: 1, Entries: entries}, "", "  ")
	require.NoError(t, err)
	regWrite(t, path, string(data))

	_, err = s.Toggle(Entry{Key: "fresh", Title: "F", AddedAt: regDay(9)})
	require.NoError(t, err)

	var f file
	require.NoError(t, json.Unmarshal([]byte(regRead(t, path)), &f))
	want := regStableNewestFirst(append(append([]Entry(nil), entries...), Entry{Key: "fresh", AddedAt: regDay(9)}))[:500]
	assert.Equal(t, want, regKeys(f.Entries), "after the cap the file is the stable newest-first order of the first 500")
}
