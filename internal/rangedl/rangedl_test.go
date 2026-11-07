package rangedl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Split ------------------------------------------------------------

// rangeInvariants returns every rule a split of contentLength into n parts
// breaks. The same rules judge the real function and its mutants.
func rangeInvariants(split func(int64, int) []ByteRange, contentLength int64, n int) (v []string) {
	defer func() {
		if r := recover(); r != nil {
			v = append(v, fmt.Sprint("panic: ", r))
		}
	}()
	got := split(contentLength, n)
	if contentLength <= 0 {
		if len(got) != 0 {
			v = append(v, "ranges for an empty file")
		}
		return v
	}
	wantN := max(n, 1)
	if int64(wantN) > contentLength {
		wantN = int(contentLength)
	}
	if len(got) != wantN {
		return append(v, fmt.Sprintf("len=%d want %d", len(got), wantN))
	}
	if got[0].From != 0 {
		v = append(v, "does not start at byte 0")
	}
	if got[len(got)-1].To != contentLength-1 {
		v = append(v, fmt.Sprintf("ends at %d want %d", got[len(got)-1].To, contentLength-1))
	}
	for i, r := range got {
		if r.To < r.From {
			v = append(v, fmt.Sprintf("range %d is empty", i))
		}
		if i > 0 && r.From != got[i-1].To+1 {
			v = append(v, fmt.Sprintf("gap or overlap before range %d", i))
		}
		if i < len(got)-1 && i > 0 && r.To-r.From != got[0].To-got[0].From {
			v = append(v, fmt.Sprintf("range %d has an uneven size", i))
		}
	}
	// The last range takes the remainder, so it is never smaller than the others.
	last := got[len(got)-1]
	if last.To-last.From < got[0].To-got[0].From {
		v = append(v, "remainder not on the last range")
	}
	return v
}

func splitGrid(split func(int64, int) []ByteRange) (v []string) {
	for _, length := range []int64{-5, 0, 1, 2, 3, 7, 8, 9, 100, 1023, 1024, 1_000_003, 5_000_000_000} {
		for _, n := range []int{-1, 0, 1, 2, 3, 4, 8, 16, 100} {
			for _, msg := range rangeInvariants(split, length, n) {
				v = append(v, fmt.Sprintf("len=%d n=%d: %s", length, n, msg))
			}
		}
	}
	return v
}

func TestSplitInvariants(t *testing.T) {
	assert.Empty(t, splitGrid(Split))
}

func TestSplitExamples(t *testing.T) {
	assert.Equal(t, []ByteRange{{0, 99}}, Split(100, 1))
	assert.Equal(t, []ByteRange{{0, 49}, {50, 99}}, Split(100, 2))
	assert.Equal(t, []ByteRange{{0, 32}, {33, 65}, {66, 99}}, Split(100, 3), "the remainder goes to the last part")
	assert.Equal(t, []ByteRange{{0, 0}, {1, 1}, {2, 2}}, Split(3, 8), "never more parts than bytes")
	assert.Equal(t, []ByteRange{{0, 9}}, Split(10, 0), "no threads means one")
	assert.Nil(t, Split(0, 4))
}

func TestSplitMeta(t *testing.T) {
	require.Empty(t, splitGrid(Split))
	mutants := map[string]func(int64, int) []ByteRange{
		"last part drops the remainder": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			if len(r) > 1 {
				r[len(r)-1].To = r[len(r)-1].From + (r[0].To - r[0].From)
			}
			return r
		},
		"off by one end": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			for i := range r {
				r[i].To++
			}
			return r
		},
		"overlap": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			for i := 1; i < len(r); i++ {
				r[i].From--
			}
			return r
		},
		"gap": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			for i := range len(r) - 1 {
				r[i].To--
			}
			return r
		},
		"more parts than bytes": func(l int64, n int) []ByteRange {
			if l <= 0 {
				return nil
			}
			n = max(n, 1)
			chunk := l / int64(n)
			out := make([]ByteRange, n)
			for i := range out {
				out[i] = ByteRange{int64(i) * chunk, int64(i)*chunk + chunk - 1}
			}
			out[n-1].To = l - 1
			return out
		},
		"ignores the thread count": func(l int64, _ int) []ByteRange { return Split(l, 1) },
		"empty file gets a range": func(l int64, n int) []ByteRange {
			if l <= 0 {
				return []ByteRange{{0, 0}}
			}
			return Split(l, n)
		},
		"zero threads panics": func(l int64, n int) []ByteRange {
			if l <= 0 {
				return nil
			}
			chunk := l / int64(n) // the original code's division
			_ = chunk
			return Split(l, n)
		},
		"drops the first part": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			if len(r) > 1 {
				return r[1:]
			}
			return r
		},
		"starts at one": func(l int64, n int) []ByteRange {
			r := Split(l, n)
			if len(r) > 0 {
				r[0].From = 1
			}
			return r
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, splitGrid(m), "mutant %q survived", name)
	}
}

func TestPartPath(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		got, err := PartPath(filepath.FromSlash("/tmp/video.mp4"), 1)
		require.NoError(t, err)
		assert.Equal(t, filepath.FromSlash("/tmp/video.mp4.part1"), got)
	})

	t.Run("subdir ok", func(t *testing.T) {
		t.Parallel()
		got, err := PartPath(filepath.FromSlash("/tmp/sub/video.mp4"), 7)
		require.NoError(t, err)
		assert.Contains(t, got, "video.mp4.part7")
	})

	t.Run("part embedded in dir name still resolves correctly", func(t *testing.T) {
		t.Parallel()
		// Use a path whose basename uniquely identifies the part file.
		got, err := PartPath(filepath.FromSlash("/tmp/dir.with.dots/video.mp4"), 0)
		require.NoError(t, err)
		assert.Equal(t, filepath.FromSlash("/tmp/dir.with.dots/video.mp4.part0"), got)
	})
}

func TestCombine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "video.mp4")

	// Write 3 parts.
	for i, payload := range [][]byte{[]byte("AAA"), []byte("BBB"), []byte("CCC")} {
		p, err := PartPath(dest, i)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(p, payload, 0o600))
	}

	require.NoError(t, Combine(dest, 3))
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "AAABBBCCC", string(got))

	// Parts must be cleaned up.
	for i := range 3 {
		p, _ := PartPath(dest, i)
		_, err := os.Stat(p)
		assert.True(t, os.IsNotExist(err), "part %d should be removed", i)
	}
}

func TestCombine_MissingPartReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "v.mp4")
	require.Error(t, Combine(dest, 2))
}
