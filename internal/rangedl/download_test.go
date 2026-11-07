package rangedl

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePart stands in for a fetch: it writes the bytes of its range.
func writePart(t *testing.T, content string) func(Part) error {
	return func(p Part) error {
		path, err := PartPath(p.Dest, p.Index)
		if err != nil {
			return err
		}
		return os.WriteFile(path, []byte(content[p.From:p.To+1]), 0o600)
	}
}

func TestDownloadJoinsEveryPartInOrder(t *testing.T) {
	const content = "the quick brown fox jumps"
	dest := filepath.Join(t.TempDir(), "v.mp4")

	require.NoError(t, Download(int64(len(content)), 4, dest, writePart(t, content)))

	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, content, string(got))
	entries, _ := os.ReadDir(filepath.Dir(dest))
	assert.Len(t, entries, 1, "the part files are gone once joined")
}

func TestDownloadCallsFetchOncePerPartWithItsRangeIndexAndDest(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "v.mp4")
	var mu sync.Mutex
	var seen []Part
	err := Download(10, 3, dest, func(p Part) error {
		mu.Lock()
		seen = append(seen, p)
		mu.Unlock()
		return writePart(t, "0123456789")(p)
	})
	require.NoError(t, err)
	sort.Slice(seen, func(i, j int) bool { return seen[i].Index < seen[j].Index })
	assert.Equal(t, []Part{
		{From: 0, To: 2, Index: 0, Dest: dest},
		{From: 3, To: 5, Index: 1, Dest: dest},
		{From: 6, To: 9, Index: 2, Dest: dest},
	}, seen)
}

func TestDownloadReportsAFailedPartAndDoesNotJoin(t *testing.T) {
	boom := errors.New("boom")
	dest := filepath.Join(t.TempDir(), "v.mp4")
	err := Download(10, 2, dest, func(p Part) error {
		if p.Index == 1 {
			return boom
		}
		return writePart(t, "0123456789")(p)
	})
	require.ErrorIs(t, err, boom)
	assert.ErrorContains(t, err, "chunk download failed")
	_, statErr := os.Stat(dest)
	assert.True(t, os.IsNotExist(statErr), "nothing is joined after a failed part")
	part0, _ := PartPath(dest, 0)
	_, statErr = os.Stat(part0)
	assert.NoError(t, statErr, "finished parts stay for a retry to resume from")
}

func TestDownloadReportsAMissingPart(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "v.mp4")
	err := Download(10, 2, dest, func(Part) error { return nil }) // writes nothing
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to combine parts")
}

func TestDownloadWithNoBytesFetchesNothing(t *testing.T) {
	called := false
	_ = Download(0, 4, filepath.Join(t.TempDir(), "v"), func(Part) error { called = true; return nil })
	assert.False(t, called)
}
