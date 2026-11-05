package player

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- formatMPVEarlyExitError -------------------------------------------------

// earlyExitFailures is the contract of the early-exit error, judged on the
// message alone so mutants (plain functions) can be judged by the same rules.
func earlyExitFailures(impl func(error, string, string) error) []string {
	var f []string
	cause := errors.New("exit status 2")
	path := `C:\mpv\mpv.exe`

	with := impl(cause, "  Failed to open VO  \n", path)
	if with == nil {
		return []string{"nil error with stderr"}
	}
	m := with.Error()
	for _, want := range []string{"exit status 2", "Failed to open VO", path, "VO fallback chain"} {
		if !strings.Contains(m, want) {
			f = append(f, "with stderr: missing "+want)
		}
	}
	if strings.Contains(m, "no stderr captured") {
		f = append(f, "with stderr: claims no stderr")
	}
	if strings.Contains(m, "  Failed") || strings.HasSuffix(strings.Split(m, "\nmpv path")[0], "\n") {
		f = append(f, "with stderr: stderr not trimmed")
	}

	for _, blank := range []string{"", "   \n\t "} {
		without := impl(cause, blank, path)
		if without == nil {
			return append(f, "nil error without stderr")
		}
		m = without.Error()
		for _, want := range []string{"exit status 2", "no stderr captured", path, `--version`} {
			if !strings.Contains(m, want) {
				f = append(f, fmt.Sprintf("blank stderr %q: missing %s", blank, want))
			}
		}
		if strings.Contains(m, "mpv stderr:") {
			f = append(f, "blank stderr: prints an empty stderr section")
		}
	}
	return f
}

func TestFormatMPVEarlyExitError(t *testing.T) {
	assert.Empty(t, earlyExitFailures(formatMPVEarlyExitError))
}

func TestFormatMPVEarlyExitErrorHasNoStdlibWrap(t *testing.T) {
	cause := errors.New("boom")
	err := formatMPVEarlyExitError(cause, "x", "/usr/bin/mpv")
	require.Error(t, err)
	assert.NotErrorIs(t, err, cause, "the cause is quoted with %v, not wrapped")
}

func TestFormatMPVEarlyExitErrorMeta(t *testing.T) {
	mutants := map[string]func(error, string, string) error{
		"always nil": func(error, string, string) error { return nil },
		"stderr untrimmed": func(w error, s, p string) error {
			if strings.TrimSpace(s) != "" {
				return fmt.Errorf("mpv exited before IPC socket was ready: %v\nmpv stderr: %s\nmpv path: %s\nHint: VO fallback chain", w, s, p)
			}
			return formatMPVEarlyExitError(w, s, p)
		},
		"blank stderr treated as present": func(w error, s, p string) error {
			return formatMPVEarlyExitError(w, "x", p)
		},
		"drops the path": func(w error, s, p string) error { return formatMPVEarlyExitError(w, s, "") },
		"drops the cause": func(_ error, s, p string) error {
			return formatMPVEarlyExitError(errors.New("other"), s, p)
		},
		"drops the stderr": func(w error, _ string, p string) error { return formatMPVEarlyExitError(w, "", p) },
		"always the stderr branch": func(w error, s, p string) error {
			return fmt.Errorf("mpv exited before IPC socket was ready: %v\nmpv stderr: %s\nmpv path: %s\nHint: VO fallback chain", w, strings.TrimSpace(s), p)
		},
	}
	require.Empty(t, earlyExitFailures(formatMPVEarlyExitError))
	for name, m := range mutants {
		assert.NotEmpty(t, earlyExitFailures(m), "mutant %q survived", name)
	}
}

// ---- takeRangeStart ----------------------------------------------------------

func TestTakeRangeStart(t *testing.T) {
	prev := currentMenuRequest
	t.Cleanup(func() { currentMenuRequest = prev })

	currentMenuRequest = EpisodeMenuRequest{}
	assert.Equal(t, 0, takeRangeStart(), "nothing preset")

	currentMenuRequest = EpisodeMenuRequest{DownloadCode: 2, RangeStart: 7}
	assert.Equal(t, 7, takeRangeStart(), "returns the preset start")
	assert.Equal(t, 0, currentMenuRequest.RangeStart, "consumed")
	assert.Equal(t, 2, currentMenuRequest.DownloadCode, "other request fields are left alone")
	assert.Equal(t, 0, takeRangeStart(), "second call finds nothing")
}

func rangeStartFailures(take func() int, set func(int), peek func() int) []string {
	var f []string
	set(0)
	if take() != 0 {
		f = append(f, "empty request returned non-zero")
	}
	set(9)
	if got := take(); got != 9 {
		f = append(f, fmt.Sprintf("returned %d want 9", got))
	}
	if peek() != 0 {
		f = append(f, "not consumed")
	}
	if take() != 0 {
		f = append(f, "second take not zero")
	}
	return f
}

func TestTakeRangeStartMeta(t *testing.T) {
	prev := currentMenuRequest
	t.Cleanup(func() { currentMenuRequest = prev })
	set := func(n int) { currentMenuRequest.RangeStart = n }
	peek := func() int { return currentMenuRequest.RangeStart }

	require.Empty(t, rangeStartFailures(takeRangeStart, set, peek))
	mutants := map[string]func() int{
		"does not consume": func() int { return currentMenuRequest.RangeStart },
		"consumes but returns zero": func() int {
			currentMenuRequest.RangeStart = 0
			return 0
		},
		"returns constant": func() int { currentMenuRequest.RangeStart = 0; return 1 },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, rangeStartFailures(m, set, peek), "mutant %q survived", name)
	}
}

// ---- PeekResumeOverride ------------------------------------------------------

func resetResumeOverride(t *testing.T) {
	t.Helper()
	ClearResumeOverride()
	t.Cleanup(ClearResumeOverride)
}

func TestPeekResumeOverride(t *testing.T) {
	type result struct {
		pos       int
		fromStart bool
		ok        bool
	}
	tests := []struct {
		name  string
		setup func()
		want  result
	}{
		{"nothing pending", func() {}, result{0, false, false}},
		{"forced resume position", func() { ForceResumeNext(95) }, result{95, false, true}},
		{"position zero is a valid override", func() { ForceResumeNext(0) }, result{0, false, true}},
		{"forced start", func() { ForceStartNext() }, result{0, true, true}},
		{"start then resume: resume wins (last call)", func() { ForceStartNext(); ForceResumeNext(30) }, result{30, false, true}},
		{"resume then start: start wins (last call)", func() { ForceResumeNext(30); ForceStartNext() }, result{0, true, true}},
		{"cleared", func() { ForceResumeNext(30); ClearResumeOverride() }, result{0, false, false}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetResumeOverride(t)
			tc.setup()
			pos, fromStart, ok := PeekResumeOverride()
			assert.Equal(t, tc.want, result{pos, fromStart, ok})

			// Peeking never consumes: a second look and the real take agree.
			pos2, fs2, ok2 := PeekResumeOverride()
			assert.Equal(t, tc.want, result{pos2, fs2, ok2}, "peek is idempotent")
			tpos, tfs, tok := takeResumeOverride()
			assert.Equal(t, tc.want, result{tpos, tfs, tok}, "peek reports exactly what take will consume")
			_, _, after := PeekResumeOverride()
			assert.False(t, after, "take consumed it")
		})
	}
}

// peekFailures judges a peek implementation against the real take: for every
// setup, peeking must equal taking and must leave the state untouched.
func peekFailures(peek func() (int, bool, bool)) []string {
	var f []string
	setups := map[string]func(){
		"none":   func() { ClearResumeOverride() },
		"resume": func() { ForceResumeNext(42) },
		"zero":   func() { ForceResumeNext(0) },
		"start":  func() { ForceStartNext() },
	}
	for name, setup := range setups {
		setup()
		p, fs, ok := peek()
		p2, fs2, ok2 := peek()
		if p != p2 || fs != fs2 || ok != ok2 {
			f = append(f, name+": second peek differs (consumed)")
		}
		tp, tfs, tok := takeResumeOverride()
		if p != tp || fs != tfs || ok != tok {
			f = append(f, fmt.Sprintf("%s: peek=(%d,%v,%v) take=(%d,%v,%v)", name, p, fs, ok, tp, tfs, tok))
		}
		ClearResumeOverride()
	}
	return f
}

func TestPeekResumeOverrideMeta(t *testing.T) {
	resetResumeOverride(t)
	require.Empty(t, peekFailures(PeekResumeOverride))

	mutants := map[string]func() (int, bool, bool){
		"consumes like take": takeResumeOverride,
		"ignores fromStart": func() (int, bool, bool) {
			p, _, ok := PeekResumeOverride()
			return p, false, ok
		},
		"treats zero as none": func() (int, bool, bool) {
			p, fs, ok := PeekResumeOverride()
			if p == 0 && !fs {
				return 0, false, false
			}
			return p, fs, ok
		},
		"always none": func() (int, bool, bool) { return 0, false, false },
		"drops the position": func() (int, bool, bool) {
			_, fs, ok := PeekResumeOverride()
			return 0, fs, ok
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, peekFailures(m), "mutant %q survived", name)
	}
}

// ---- GetAnimeSeason ----------------------------------------------------------

func TestGetAnimeSeason(t *testing.T) {
	prevName, prevSeason := gMedia.animeName, gMedia.animeSeason
	t.Cleanup(func() { SetAnimeName(prevName, prevSeason) })

	for _, tc := range []struct{ set, want int }{{1, 1}, {3, 3}, {12, 12}, {0, 1}, {-4, 1}} {
		SetAnimeName("Show", tc.set)
		assert.Equal(t, tc.want, GetAnimeSeason(), "SetAnimeName(_, %d)", tc.set)
	}
}

func seasonFailures(get func() int) []string {
	var f []string
	for _, tc := range []struct{ set, want int }{{1, 1}, {3, 3}, {0, 1}, {-4, 1}} {
		SetAnimeName("Show", tc.set)
		if got := get(); got != tc.want {
			f = append(f, fmt.Sprintf("season %d: got %d want %d", tc.set, got, tc.want))
		}
	}
	return f
}

func TestGetAnimeSeasonMeta(t *testing.T) {
	prevName, prevSeason := gMedia.animeName, gMedia.animeSeason
	t.Cleanup(func() { SetAnimeName(prevName, prevSeason) })
	require.Empty(t, seasonFailures(GetAnimeSeason))
	for name, m := range map[string]func() int{
		"always one": func() int { return 1 },
		"raw field, no clamp": func() int {
			gMedia.mu.Lock()
			defer gMedia.mu.Unlock()
			return gMedia.animeSeason - 1
		},
		"off by one": func() int { return GetAnimeSeason() + 1 },
		"zero":       func() int { return 0 },
	} {
		assert.NotEmpty(t, seasonFailures(m), "mutant %q survived", name)
	}
}

// ---- episodePickNum ----------------------------------------------------------

var episodePickCases = []struct {
	name string
	eps  []models.Episode
	idx  int
	want int
}{
	{"Num wins", []models.Episode{{Num: 12, Number: "99"}}, 0, 12},
	{"Num wins over the text", []models.Episode{{Num: 3, Number: "Episode 7"}}, 0, 3},
	{"parses the text when Num is zero", []models.Episode{{Number: "5"}}, 0, 5},
	{"parses Portuguese label", []models.Episode{{Number: "Episódio 8"}}, 0, 8},
	{"zero in the text falls back to position", []models.Episode{{}, {Number: "0"}}, 1, 2},
	{"negative Num falls back to the text", []models.Episode{{Num: -1, Number: "4"}}, 0, 4},
	{"nothing numeric falls back to position", []models.Episode{{Number: "OVA"}, {Number: "Special"}, {Number: "Final"}}, 2, 3},
	{"a label with no digits is not episode 1", []models.Episode{{Number: "x"}, {Number: "y"}, {Number: "Movie"}, {Number: "z"}}, 3, 4},
	{"Num of 1 is kept (boundary)", []models.Episode{{}, {}, {Num: 1}}, 2, 1},
	{"text of 1 is kept (boundary)", []models.Episode{{}, {}, {Number: "1"}}, 2, 1},
	{"empty falls back to position", []models.Episode{{}}, 0, 1},
	{"uses the indexed episode only", []models.Episode{{Num: 1}, {Num: 2}, {Num: 3}}, 1, 2},
}

func episodePickFailures(impl func([]models.Episode, int) int) []string {
	var f []string
	for _, tc := range episodePickCases {
		if got := impl(tc.eps, tc.idx); got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d want %d", tc.name, got, tc.want))
		}
	}
	return f
}

func TestEpisodePickNum(t *testing.T) {
	assert.Empty(t, episodePickFailures(episodePickNum))
	for _, tc := range episodePickCases {
		assert.Equal(t, tc.want, episodePickNum(tc.eps, tc.idx), tc.name)
	}
}

func TestEpisodePickNumMeta(t *testing.T) {
	mutants := map[string]func([]models.Episode, int) int{
		"ignores Num": func(e []models.Episode, i int) int {
			return episodePickNum([]models.Episode{{Number: e[i].Number}}, 0) + 0*i
		},
		"always position":  func(_ []models.Episode, i int) int { return i + 1 },
		"zero-based":       func(e []models.Episode, i int) int { return episodePickNum(e, i) - 1 },
		"accepts Num of 0": func(e []models.Episode, i int) int { return e[i].Num },
		"trusts the extractor's default 1": func(e []models.Episode, i int) int {
			if e[i].Num > 0 {
				return e[i].Num
			}
			if n, err := strconv.Atoi(ExtractEpisodeNumber(e[i].Number)); err == nil && n > 0 {
				return n
			}
			return i + 1
		},
		"never parses text": func(e []models.Episode, i int) int {
			if e[i].Num > 0 {
				return e[i].Num
			}
			return i + 1
		},
		"accepts zero from text": func(e []models.Episode, i int) int {
			if e[i].Num > 0 {
				return e[i].Num
			}
			var n int
			if _, err := fmt.Sscanf(ExtractEpisodeNumber(e[i].Number), "%d", &n); err == nil {
				return n
			}
			return i + 1
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, episodePickFailures(m), "mutant %q survived", name)
	}
}
