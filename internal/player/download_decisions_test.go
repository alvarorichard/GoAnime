package player

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- splitByteRanges ------------------------------------------------------------

// rangeInvariants returns every rule a split of contentLength into n parts
// breaks. The same rules judge the real function and its mutants.
func rangeInvariants(split func(int64, int) []byteRange, contentLength int64, n int) (v []string) {
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
	if got[0].from != 0 {
		v = append(v, "does not start at byte 0")
	}
	if got[len(got)-1].to != contentLength-1 {
		v = append(v, fmt.Sprintf("ends at %d want %d", got[len(got)-1].to, contentLength-1))
	}
	for i, r := range got {
		if r.to < r.from {
			v = append(v, fmt.Sprintf("range %d is empty", i))
		}
		if i > 0 && r.from != got[i-1].to+1 {
			v = append(v, fmt.Sprintf("gap or overlap before range %d", i))
		}
		if i < len(got)-1 && i > 0 && r.to-r.from != got[0].to-got[0].from {
			v = append(v, fmt.Sprintf("range %d has an uneven size", i))
		}
	}
	// The last range takes the remainder, so it is never smaller than the others.
	last := got[len(got)-1]
	if last.to-last.from < got[0].to-got[0].from {
		v = append(v, "remainder not on the last range")
	}
	return v
}

func splitGrid(split func(int64, int) []byteRange) (v []string) {
	for _, length := range []int64{-5, 0, 1, 2, 3, 7, 8, 9, 100, 1023, 1024, 1_000_003, 5_000_000_000} {
		for _, n := range []int{-1, 0, 1, 2, 3, 4, 8, 16, 100} {
			for _, msg := range rangeInvariants(split, length, n) {
				v = append(v, fmt.Sprintf("len=%d n=%d: %s", length, n, msg))
			}
		}
	}
	return v
}

func TestSplitByteRangesInvariants(t *testing.T) {
	assert.Empty(t, splitGrid(splitByteRanges))
}

func TestSplitByteRangesExamples(t *testing.T) {
	assert.Equal(t, []byteRange{{0, 99}}, splitByteRanges(100, 1))
	assert.Equal(t, []byteRange{{0, 49}, {50, 99}}, splitByteRanges(100, 2))
	assert.Equal(t, []byteRange{{0, 32}, {33, 65}, {66, 99}}, splitByteRanges(100, 3), "the remainder goes to the last part")
	assert.Equal(t, []byteRange{{0, 0}, {1, 1}, {2, 2}}, splitByteRanges(3, 8), "never more parts than bytes")
	assert.Equal(t, []byteRange{{0, 9}}, splitByteRanges(10, 0), "no threads means one")
	assert.Nil(t, splitByteRanges(0, 4))
}

func TestSplitByteRangesMeta(t *testing.T) {
	require.Empty(t, splitGrid(splitByteRanges))
	mutants := map[string]func(int64, int) []byteRange{
		"last part drops the remainder": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			if len(r) > 1 {
				r[len(r)-1].to = r[len(r)-1].from + (r[0].to - r[0].from)
			}
			return r
		},
		"off by one end": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			for i := range r {
				r[i].to++
			}
			return r
		},
		"overlap": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			for i := 1; i < len(r); i++ {
				r[i].from--
			}
			return r
		},
		"gap": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			for i := range len(r) - 1 {
				r[i].to--
			}
			return r
		},
		"more parts than bytes": func(l int64, n int) []byteRange {
			if l <= 0 {
				return nil
			}
			n = max(n, 1)
			chunk := l / int64(n)
			out := make([]byteRange, n)
			for i := range out {
				out[i] = byteRange{int64(i) * chunk, int64(i)*chunk + chunk - 1}
			}
			out[n-1].to = l - 1
			return out
		},
		"ignores the thread count": func(l int64, _ int) []byteRange { return splitByteRanges(l, 1) },
		"empty file gets a range": func(l int64, n int) []byteRange {
			if l <= 0 {
				return []byteRange{{0, 0}}
			}
			return splitByteRanges(l, n)
		},
		"zero threads panics": func(l int64, n int) []byteRange {
			if l <= 0 {
				return nil
			}
			chunk := l / int64(n) // the original code's division
			_ = chunk
			return splitByteRanges(l, n)
		},
		"drops the first part": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			if len(r) > 1 {
				return r[1:]
			}
			return r
		},
		"starts at one": func(l int64, n int) []byteRange {
			r := splitByteRanges(l, n)
			if len(r) > 0 {
				r[0].from = 1
			}
			return r
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, splitGrid(m), "mutant %q survived", name)
	}
}

// ---- pickFallbackCandidate --------------------------------------------------------

var fallbackCases = []struct {
	name       string
	candidates []string
	failed     string
	want       string
	wantErr    bool
}{
	{"first that is not the failed one", []string{"a", "b", "c"}, "a", "b", false},
	{"the failed one in the middle", []string{"a", "b", "c"}, "b", "a", false},
	{"nothing failed yet", []string{"a", "b"}, "", "a", false},
	{"empty entries are skipped", []string{"", "", "c"}, "x", "c", false},
	{"only the failed one", []string{"a"}, "a", "", true},
	{"failed repeated", []string{"a", "a"}, "a", "", true},
	{"no candidates", nil, "a", "", true},
	{"only empties", []string{"", ""}, "", "", true},
}

func fallbackFailures(impl func([]string, string) (string, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range fallbackCases {
		got, err := impl(tc.candidates, tc.failed)
		if got != tc.want || (err != nil) != tc.wantErr {
			f = append(f, fmt.Sprintf("%s: got (%q,%v)", tc.name, got, err))
		}
	}
	return f
}

func TestPickFallbackCandidate(t *testing.T) {
	assert.Empty(t, fallbackFailures(pickFallbackCandidate))
}

func TestPickFallbackCandidateMeta(t *testing.T) {
	require.Empty(t, fallbackFailures(pickFallbackCandidate))
	mutants := map[string]func([]string, string) (string, error){
		"returns the failed one": func(c []string, _ string) (string, error) { return pickFallbackCandidate(c, "\x00") },
		"last instead of first": func(c []string, f string) (string, error) {
			for i := len(c) - 1; i >= 0; i-- {
				if c[i] != "" && c[i] != f {
					return c[i], nil
				}
			}
			return "", errors.New("none")
		},
		"empty accepted": func(c []string, f string) (string, error) {
			for _, x := range c {
				if x != f {
					return x, nil
				}
			}
			return "", errors.New("none")
		},
		"no error when exhausted": func(c []string, f string) (string, error) {
			got, err := pickFallbackCandidate(c, f)
			if err != nil {
				return "", nil //nolint:nilerr // deliberate mutant: no error when exhausted
			}
			return got, nil
		},
		"always first": func(c []string, _ string) (string, error) {
			if len(c) == 0 {
				return "", errors.New("none")
			}
			return c[0], nil
		},
		"stops at the failed one": func(c []string, f string) (string, error) {
			for _, x := range c {
				if x == f {
					break
				}
				if x != "" {
					return x, nil
				}
			}
			return "", errors.New("none")
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, fallbackFailures(m), "mutant %q survived", name)
	}
}

// ---- parseEpisodeRange / validEpisodeNumber -----------------------------------------

var rangeParseCases = []struct {
	start, end string
	ws, we     int
	wantErr    string
}{
	{"1", "12", 1, 12, ""},
	{" 3 ", "\t5\n", 3, 5, ""},
	{"7", "7", 7, 7, ""},
	{"0", "0", 0, 0, ""},
	{"-2", "3", -2, 3, ""},
	{"5", "4", 0, 0, "start cannot be greater"},
	{"x", "4", 0, 0, "invalid number"},
	{"1", "y", 0, 0, "invalid number"},
	{"", "3", 0, 0, "invalid number"},
	{"1.5", "3", 0, 0, "invalid number"},
	{"1", "", 0, 0, "invalid number"},
}

func rangeParseFailures(impl func(string, string) (int, int, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range rangeParseCases {
		s, e, err := impl(tc.start, tc.end)
		if tc.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) || s != 0 || e != 0 {
				f = append(f, fmt.Sprintf("%q-%q: got (%d,%d,%v)", tc.start, tc.end, s, e, err))
			}
			continue
		}
		if err != nil || s != tc.ws || e != tc.we {
			f = append(f, fmt.Sprintf("%q-%q: got (%d,%d,%v)", tc.start, tc.end, s, e, err))
		}
	}
	return f
}

func TestParseEpisodeRange(t *testing.T) {
	assert.Empty(t, rangeParseFailures(parseEpisodeRange))
}

func TestParseEpisodeRangeMeta(t *testing.T) {
	require.Empty(t, rangeParseFailures(parseEpisodeRange))
	mutants := map[string]func(string, string) (int, int, error){
		"start may exceed end": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err != nil && strings.Contains(err.Error(), "greater") {
				return 5, 4, nil
			}
			return s, e, err
		},
		"equal is rejected": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err == nil && s == e {
				return 0, 0, errors.New("start cannot be greater than end")
			}
			return s, e, err
		},
		"no trimming": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err == nil && (strings.TrimSpace(a) != a || strings.TrimSpace(b) != b) {
				return 0, 0, errors.New("invalid number")
			}
			return s, e, err
		},
		"bad number ignored as zero": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err != nil && strings.Contains(err.Error(), "invalid number") {
				return 0, 0, nil
			}
			return s, e, err
		},
		"swapped": func(a, b string) (int, int, error) { return parseEpisodeRange(b, a) },
		"values on error": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err != nil {
				return 1, 1, err
			}
			return s, e, err
		},
		"negative rejected": func(a, b string) (int, int, error) {
			s, e, err := parseEpisodeRange(a, b)
			if err == nil && s < 0 {
				return 0, 0, errors.New("invalid number")
			}
			return s, e, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, rangeParseFailures(m), "mutant %q survived", name)
	}
}

func TestValidEpisodeNumber(t *testing.T) {
	for _, ok := range []string{"1", " 12 ", "-3", "0", "007"} {
		assert.NoError(t, validEpisodeNumber(ok), ok)
	}
	for _, bad := range []string{"", " ", "a", "1.5", "1 2", "0x10"} {
		assert.EqualError(t, validEpisodeNumber(bad), "invalid number", "%q", bad)
	}
}

// ---- the "watch what you downloaded" menu -----------------------------------------------

func epsWith(titles map[int][2]string, nums ...int) []models.Episode {
	var out []models.Episode
	for _, n := range nums {
		e := models.Episode{Num: n, Number: fmt.Sprint(n)}
		if t, ok := titles[n]; ok {
			e.Title = models.TitleDetails{English: t[0], Romaji: t[1]}
		}
		out = append(out, e)
	}
	return out
}

func labelsOf(es []watchMenuEntry) (l, v []string) {
	for _, e := range es {
		l = append(l, e.label)
		v = append(v, e.value)
	}
	return l, v
}

var menuCases = []struct {
	name       string
	eps        []models.Episode
	wantLabels []string
	wantValues []string
}{
	{"plain", epsWith(nil, 1, 2), []string{"Episode 1", "Episode 2", "Don't watch anything"}, []string{"1", "2", "exit"}},
	{"english title wins", epsWith(map[int][2]string{3: {"The Test", "Tesuto"}}, 3), []string{"Episode 3: The Test", "Don't watch anything"}, []string{"3", "exit"}},
	{"romaji when no english", epsWith(map[int][2]string{4: {"", "Tesuto"}}, 4), []string{"Episode 4: Tesuto", "Don't watch anything"}, []string{"4", "exit"}},
	{"no episodes still offers the exit", nil, []string{"Don't watch anything"}, []string{"exit"}},
}

func menuFailures(impl func([]models.Episode) []watchMenuEntry) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range menuCases {
		l, v := labelsOf(impl(tc.eps))
		if fmt.Sprint(l) != fmt.Sprint(tc.wantLabels) || fmt.Sprint(v) != fmt.Sprint(tc.wantValues) {
			f = append(f, fmt.Sprintf("%s: %v %v", tc.name, l, v))
		}
	}
	return f
}

func TestWatchMenuEntries(t *testing.T) {
	assert.Empty(t, menuFailures(watchMenuEntries))
}

func TestWatchMenuEntriesMeta(t *testing.T) {
	require.Empty(t, menuFailures(watchMenuEntries))
	mutants := map[string]func([]models.Episode) []watchMenuEntry{
		"no exit row": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			return r[:len(r)-1]
		},
		"exit row first": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			return append([]watchMenuEntry{r[len(r)-1]}, r[:len(r)-1]...)
		},
		"romaji preferred": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			for i := range e {
				if e[i].Title.Romaji != "" {
					r[i].label = fmt.Sprintf("Episode %d: %s", e[i].Num, e[i].Title.Romaji)
				}
			}
			return r
		},
		"titles ignored": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			for i := range e {
				r[i].label = fmt.Sprintf("Episode %d", e[i].Num)
			}
			return r
		},
		"value is the index": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			for i := range e {
				r[i].value = fmt.Sprint(i)
			}
			return r
		},
		"reversed": func(e []models.Episode) []watchMenuEntry {
			r := watchMenuEntries(e)
			for i, j := 0, len(r)-2; i < j; i, j = i+1, j-1 {
				r[i], r[j] = r[j], r[i]
			}
			return r
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, menuFailures(m), "mutant %q survived", name)
	}
}

func TestWatchPickError(t *testing.T) {
	assert.Equal(t, ErrUserQuit, watchPickError(tui.ErrPickBack))
	assert.Equal(t, ErrUserQuit, watchPickError(tui.ErrPickCancelled))
	assert.Equal(t, ErrUserQuit, watchPickError(fmt.Errorf("wrapped: %w", tui.ErrPickBack)))
	boom := errors.New("terminal gone")
	err := watchPickError(boom)
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, ErrUserQuit, "a real failure is not a quit")
	assert.Contains(t, err.Error(), "episode selection error")
}

// menuHarness mocks the disk, the picker and the player.
type menuHarness struct {
	onDisk     map[int]bool
	pathErr    map[int]error
	pickIdx    int
	pickErr    error
	playErr    error
	asked      [][]string
	breadcrumb string
	played     []string
	playedNum  int
}

func (h *menuHarness) install(t *testing.T) {
	t.Helper()
	oldPath, oldExists, oldPick, oldPlay := episodePathFn, pathExistsFn, pickLabelsFn, playLocalFn
	t.Cleanup(func() { episodePathFn, pathExistsFn, pickLabelsFn, playLocalFn = oldPath, oldExists, oldPick, oldPlay })
	episodePathFn = func(_ string, n int) (string, error) {
		if err := h.pathErr[n]; err != nil {
			return "", err
		}
		return filepath.Join("lib", fmt.Sprintf("ep%d.mp4", n)), nil
	}
	pathExistsFn = func(p string) bool {
		for n, ok := range h.onDisk {
			if ok && p == filepath.Join("lib", fmt.Sprintf("ep%d.mp4", n)) {
				return true
			}
		}
		return false
	}
	pickLabelsFn = func(labels []string, o tui.PickOptions) (int, error) {
		h.asked = append(h.asked, labels)
		h.breadcrumb = o.Breadcrumb
		return h.pickIdx, h.pickErr
	}
	playLocalFn = func(path string, _ []models.Episode, n int) error {
		h.played = append(h.played, path)
		h.playedNum = n
		return h.playErr
	}
}

func TestDownloadedEpisodesInRange(t *testing.T) {
	h := &menuHarness{onDisk: map[int]bool{1: true, 3: true, 4: true, 9: true}, pathErr: map[int]error{4: errors.New("no dir")}}
	h.install(t)
	eps := epsWith(nil, 1, 2, 3, 4, 5, 9)

	got := downloadedEpisodesInRange(eps, "url", 1, 5)
	var nums []int
	for _, e := range got {
		nums = append(nums, e.Num)
	}
	assert.Equal(t, []int{1, 3}, nums, "missing files, path errors and out-of-range numbers are left out")

	assert.Empty(t, downloadedEpisodesInRange(eps, "url", 6, 8), "nothing in a gap")
	assert.Len(t, downloadedEpisodesInRange(eps, "url", 9, 9), 1, "single-episode range, both ends inclusive")
	assert.Empty(t, downloadedEpisodesInRange(eps, "url", 5, 1), "reversed range is empty")
}

func downloadedFailures(impl func([]models.Episode, string, int, int) []models.Episode) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	nums := func(es []models.Episode) string {
		var s []string
		for _, e := range es {
			s = append(s, fmt.Sprint(e.Num))
		}
		return strings.Join(s, ",")
	}
	eps := epsWith(nil, 1, 2, 3, 4, 5, 9)
	for _, tc := range []struct {
		a, b int
		want string
	}{{1, 5, "1,3"}, {3, 3, "3"}, {1, 1, "1"}, {2, 2, ""}, {9, 9, "9"}, {5, 1, ""}, {4, 4, ""}, {0, 100, "1,3,9"}} {
		if got := nums(impl(eps, "u", tc.a, tc.b)); got != tc.want {
			f = append(f, fmt.Sprintf("%d-%d: got %q want %q", tc.a, tc.b, got, tc.want))
		}
	}
	return f
}

func TestDownloadedEpisodesInRangeMeta(t *testing.T) {
	h := &menuHarness{onDisk: map[int]bool{1: true, 3: true, 4: true, 9: true}, pathErr: map[int]error{4: errors.New("no dir")}}
	h.install(t)
	require.Empty(t, downloadedFailures(downloadedEpisodesInRange))
	mutants := map[string]func([]models.Episode, string, int, int) []models.Episode{
		"end exclusive": func(e []models.Episode, u string, a, b int) []models.Episode {
			return downloadedEpisodesInRange(e, u, a, b-1)
		},
		"start exclusive": func(e []models.Episode, u string, a, b int) []models.Episode {
			return downloadedEpisodesInRange(e, u, a+1, b)
		},
		"ignores the disk": func(e []models.Episode, u string, a, b int) []models.Episode {
			var out []models.Episode
			for n := a; n <= b; n++ {
				if ep, ok := findEpisode(e, n); ok {
					out = append(out, ep)
				}
			}
			return out
		},
		"path errors count as present": func(e []models.Episode, u string, a, b int) []models.Episode {
			var out []models.Episode
			for n := a; n <= b; n++ {
				ep, ok := findEpisode(e, n)
				if !ok {
					continue
				}
				if p, err := episodePathFn(u, n); err != nil || pathExistsFn(p) {
					out = append(out, ep)
				}
			}
			return out
		},
		"descending": func(e []models.Episode, u string, a, b int) []models.Episode {
			r := downloadedEpisodesInRange(e, u, a, b)
			for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
				r[i], r[j] = r[j], r[i]
			}
			return r
		},
		"first only": func(e []models.Episode, u string, a, b int) []models.Episode {
			r := downloadedEpisodesInRange(e, u, a, b)
			if len(r) > 1 {
				return r[:1]
			}
			return r
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, downloadedFailures(m), "mutant %q survived", name)
	}
}

func TestWatchDownloadedRangePlaysTheChosenEpisode(t *testing.T) {
	h := &menuHarness{onDisk: map[int]bool{2: true, 4: true}, pickIdx: 1}
	h.install(t)
	eps := epsWith(nil, 1, 2, 3, 4)

	err := watchDownloadedRange(eps, "url", 1, 4, "Download > Watch")
	require.NoError(t, err)
	require.Len(t, h.asked, 1)
	assert.Equal(t, []string{"Episode 2", "Episode 4", "Don't watch anything"}, h.asked[0])
	assert.Equal(t, "Download > Watch", h.breadcrumb)
	assert.Equal(t, []string{filepath.Join("lib", "ep4.mp4")}, h.played, "the second row is episode 4")
	assert.Equal(t, 4, h.playedNum)
}

func TestWatchDownloadedRangeOutcomes(t *testing.T) {
	boom := errors.New("terminal gone")
	playBoom := errors.New("mpv missing")
	tests := []struct {
		name       string
		h          menuHarness
		wantErr    error
		wantErrHas string
		wantPlayed int
		wantAsked  int
	}{
		{"exit row", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 1}, ErrUserQuit, "", 0, 1},
		{"back", menuHarness{onDisk: map[int]bool{1: true}, pickErr: tui.ErrPickBack}, ErrUserQuit, "", 0, 1},
		{"cancelled", menuHarness{onDisk: map[int]bool{1: true}, pickErr: tui.ErrPickCancelled}, ErrUserQuit, "", 0, 1},
		{"picker failure", menuHarness{onDisk: map[int]bool{1: true}, pickErr: boom}, boom, "episode selection error", 0, 1},
		{"index past the end", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 7}, nil, "selected episode not found", 0, 1},
		{"index just past the last row", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 2}, nil, "selected episode not found", 0, 1},
		{"negative index", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: -1}, nil, "selected episode not found", 0, 1},
		{"nothing on disk", menuHarness{onDisk: map[int]bool{}}, nil, "", 0, 0},
		{"player fails", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 0, playErr: playBoom}, playBoom, "", 1, 1},
		{"plays", menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 0}, nil, "", 1, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := tc.h
			h.install(t)
			err := watchDownloadedRange(epsWith(nil, 1, 2), "url", 1, 2, "bc")
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrHas != "":
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrHas)
			default:
				assert.NoError(t, err)
			}
			assert.Len(t, h.played, tc.wantPlayed)
			assert.Len(t, h.asked, tc.wantAsked, "the picker opens only when something is on disk")
		})
	}
}

func TestWatchDownloadedRangePathFailureAfterPickingIsReported(t *testing.T) {
	h := &menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 0}
	h.install(t)
	calls := 0
	base := episodePathFn
	episodePathFn = func(u string, n int) (string, error) {
		calls++
		if calls > 1 { // the listing worked; resolving the chosen file does not
			return "", errors.New("disk full")
		}
		return base(u, n)
	}
	err := watchDownloadedRange(epsWith(nil, 1), "url", 1, 1, "bc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episode path")
	assert.Empty(t, h.played)
}

func TestTheTwoEntryPointsDifferOnlyInTheBreadcrumb(t *testing.T) {
	for entry, want := range map[string]string{"existing": "Download > Watch", "range": "Download > Watch Range"} {
		h := &menuHarness{onDisk: map[int]bool{1: true}, pickErr: tui.ErrPickBack}
		h.install(t)
		eps := epsWith(nil, 1)
		var err error
		if entry == "existing" {
			err = handleExistingEpisodes(eps, "url", 1, 1)
		} else {
			err = askAndPlayDownloadedEpisode(eps, "url", 1, 1)
		}
		assert.ErrorIs(t, err, ErrUserQuit, entry)
		assert.Equal(t, want, h.breadcrumb, entry)
	}
}

func watchOutcomeFailures(impl func([]models.Episode, string, int, int, string) error) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	run := func(h *menuHarness) error {
		oldPath, oldExists, oldPick, oldPlay := episodePathFn, pathExistsFn, pickLabelsFn, playLocalFn
		defer func() { episodePathFn, pathExistsFn, pickLabelsFn, playLocalFn = oldPath, oldExists, oldPick, oldPlay }()
		install(h)
		return impl(epsWith(nil, 1, 2), "url", 1, 2, "bc")
	}
	if err := run(&menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 1}); !errors.Is(err, ErrUserQuit) {
		f = append(f, fmt.Sprintf("exit row: %v", err))
	}
	if err := run(&menuHarness{onDisk: map[int]bool{1: true}, pickErr: tui.ErrPickBack}); !errors.Is(err, ErrUserQuit) {
		f = append(f, fmt.Sprintf("back: %v", err))
	}
	if err := run(&menuHarness{onDisk: map[int]bool{}}); err != nil {
		f = append(f, fmt.Sprintf("nothing on disk: %v", err))
	}
	h := &menuHarness{onDisk: map[int]bool{1: true, 2: true}, pickIdx: 1}
	if err := run(h); err != nil || h.playedNum != 2 {
		f = append(f, fmt.Sprintf("plays: %v num=%d", err, h.playedNum))
	}
	if err := run(&menuHarness{onDisk: map[int]bool{1: true}, pickIdx: 9}); err == nil {
		f = append(f, "out-of-range index accepted")
	}
	return f
}

// install is menuHarness.install without a *testing.T, for the metatest.
func install(h *menuHarness) {
	episodePathFn = func(_ string, n int) (string, error) { return filepath.Join("lib", fmt.Sprintf("ep%d.mp4", n)), nil }
	pathExistsFn = func(p string) bool {
		for n, ok := range h.onDisk {
			if ok && p == filepath.Join("lib", fmt.Sprintf("ep%d.mp4", n)) {
				return true
			}
		}
		return false
	}
	pickLabelsFn = func([]string, tui.PickOptions) (int, error) { return h.pickIdx, h.pickErr }
	playLocalFn = func(_ string, _ []models.Episode, n int) error { h.playedNum = n; return h.playErr }
}

func TestWatchDownloadedRangeMeta(t *testing.T) {
	require.Empty(t, watchOutcomeFailures(watchDownloadedRange))
	mutants := map[string]func([]models.Episode, string, int, int, string) error{
		"exit plays the first": func(e []models.Episode, u string, a, b int, c string) error {
			err := watchDownloadedRange(e, u, a, b, c)
			if errors.Is(err, ErrUserQuit) {
				return playLocalFn("x", e, 1)
			}
			return err
		},
		"back is an error": func(e []models.Episode, u string, a, b int, c string) error {
			err := watchDownloadedRange(e, u, a, b, c)
			if errors.Is(err, ErrUserQuit) {
				return errors.New("episode selection error")
			}
			return err
		},
		"empty range is an error": func(e []models.Episode, u string, a, b int, c string) error {
			if len(downloadedEpisodesInRange(e, u, a, b)) == 0 {
				return errors.New("nothing")
			}
			return watchDownloadedRange(e, u, a, b, c)
		},
		"plays the wrong row": func(e []models.Episode, u string, a, b int, c string) error {
			return watchDownloadedRange(e, u, a, a, c)
		},
		"bad index swallowed": func(e []models.Episode, u string, a, b int, c string) error {
			err := watchDownloadedRange(e, u, a, b, c)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return nil
			}
			return err
		},
		"never plays": func(e []models.Episode, u string, a, b int, c string) error {
			if err := watchDownloadedRange(e, u, a, b, c); err != nil {
				return err
			}
			return nil
		},
	}
	// "never plays" is only a mutant if it suppresses the player call:
	mutants["never plays"] = func(e []models.Episode, u string, a, b int, c string) error {
		old := playLocalFn
		playLocalFn = func(string, []models.Episode, int) error { return nil }
		defer func() { playLocalFn = old }()
		return watchDownloadedRange(e, u, a, b, c)
	}
	for name, m := range mutants {
		assert.NotEmpty(t, watchOutcomeFailures(m), "mutant %q survived", name)
	}
}
