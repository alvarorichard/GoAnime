package tracking

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolatedHome points HOME/USERPROFILE at a temp dir, resets the global
// tracker and restores both afterwards, so the package-level helpers
// (RecordSearch, RecentSearches, ContinueEntries) run against a throwaway DB.
func isolatedHome(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("DBPath on Windows follows LOCALAPPDATA")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	resetGlobalTracker(t)
	t.Cleanup(func() { resetGlobalTracker(t) })
	return home
}

func needSQLite(t *testing.T) {
	t.Helper()
	if !IsCgoEnabled || NewLocalTracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
}

func withoutCGO(t *testing.T) {
	t.Helper()
	prev := IsCgoEnabled
	IsCgoEnabled = false
	t.Cleanup(func() { IsCgoEnabled = prev })
}

// ---- DBPath ------------------------------------------------------------------

func TestDBPathFollowsHome(t *testing.T) {
	home := isolatedHome(t)
	assert.Equal(t, filepath.Join(home, ".local", "goanime", "tracking", "progress.db"), DBPath())
}

func TestDBPathEmptyWithoutHome(t *testing.T) {
	isolatedHome(t)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	assert.Empty(t, DBPath(), "no home directory, no path: callers treat \"\" as unavailable")
}

// ---- sharedTracker -----------------------------------------------------------

func TestSharedTrackerNilWithoutCGO(t *testing.T) {
	isolatedHome(t)
	withoutCGO(t)
	assert.Nil(t, sharedTracker())
}

func TestSharedTrackerNilWithoutHome(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	assert.Nil(t, sharedTracker())
}

func TestSharedTrackerReusesTheGlobalOne(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	first := sharedTracker()
	require.NotNil(t, first)
	assert.Same(t, first, GetGlobalTracker(), "opening caches it globally")
	assert.Same(t, first, sharedTracker(), "second call reuses it, no second open")
}

func TestSharedTrackerCreatesTheDatabaseUnderHome(t *testing.T) {
	needSQLite(t)
	home := isolatedHome(t)
	require.NotNil(t, sharedTracker())
	assert.FileExists(t, filepath.Join(home, ".local", "goanime", "tracking", "progress.db"))
}

// ---- RecordSearch / RecentSearches -------------------------------------------

func TestRecordSearchRoundTrip(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)

	require.NoError(t, RecordSearch("naruto", 12))
	require.NoError(t, RecordSearch("  one   piece\t", 7), "whitespace is collapsed")
	require.NoError(t, RecordSearch("   ", 3), "blank query is ignored, not an error")
	require.NoError(t, RecordSearch("", 3))

	got, err := RecentSearches(10)
	require.NoError(t, err)
	require.Len(t, got, 2, "blank queries were not stored")
	byQuery := map[string]int{}
	for _, r := range got {
		byQuery[r.Query] = r.ResultCount
		assert.WithinDuration(t, time.Now(), r.SearchedAt, time.Minute)
	}
	assert.Equal(t, map[string]int{"naruto": 12, "one piece": 7}, byQuery)
}

func TestRecordSearchUpdatesInsteadOfDuplicating(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	require.NoError(t, RecordSearch("bleach", 1))
	require.NoError(t, RecordSearch("bleach", 9))
	got, err := RecentSearches(10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 9, got[0].ResultCount)
}

func TestRecordSearchIsBestEffortWithoutSQLite(t *testing.T) {
	isolatedHome(t)
	withoutCGO(t)
	assert.NoError(t, RecordSearch("naruto", 1), "never fails the search")
	got, err := RecentSearches(5)
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestRecentSearchesHonoursLimit(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	for i := range 5 {
		require.NoError(t, RecordSearch(fmt.Sprintf("q%d", i), i))
	}
	for limit, want := range map[int]int{1: 1, 3: 3, 5: 5, 50: 5, 0: 0, -1: 0} {
		got, err := RecentSearches(limit)
		require.NoError(t, err)
		assert.Len(t, got, want, "limit %d", limit)
	}
}

// ---- ContinueEntries ---------------------------------------------------------

func TestContinueEntriesNewestFirst(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	tr := sharedTracker()
	require.NotNil(t, tr)
	base := time.Unix(1_700_000_000, 0)
	for i, id := range []string{"a", "b", "c", "d"} {
		// Write out of order so the sort, not insertion, decides.
		offsets := []int{2, 0, 3, 1}
		require.NoError(t, tr.UpdateProgress(Anime{
			AnilistID: i + 1, AllanimeID: id, EpisodeNumber: i + 1, PlaybackTime: 10, Duration: 100,
			Title: "t" + id, AnimeTitle: "Show " + id, LastUpdated: base.Add(time.Duration(offsets[i]) * time.Hour),
		}))
	}
	got, err := ContinueEntries()
	require.NoError(t, err)
	var ids []string
	for _, e := range got {
		ids = append(ids, e.AllanimeID)
	}
	assert.Equal(t, []string{"c", "a", "d", "b"}, ids)
	assert.Equal(t, "Show c", got[0].AnimeTitle)
}

func TestContinueEntriesEmptyAndNilSafe(t *testing.T) {
	isolatedHome(t)
	withoutCGO(t)
	got, err := ContinueEntries()
	assert.NoError(t, err)
	assert.Empty(t, got)

	IsCgoEnabled = true
	needSQLite(t)
	got, err = ContinueEntries()
	assert.NoError(t, err)
	assert.Empty(t, got, "no history yet")
}

// ---- metatests ---------------------------------------------------------------

// normalizeQuery mirrors the rule RecordSearch applies; recordNormalizationFailures
// judges any implementation of it by the same table the real code must satisfy.
func recordNormalizationFailures(norm func(string) string) []string {
	var f []string
	for in, want := range map[string]string{
		"naruto": "naruto", "  one   piece\t": "one piece", "   ": "", "": "",
		"a\nb": "a b", "x  y  z": "x y z", " Death   Note ": "Death Note", "ナルト  疾風伝": "ナルト 疾風伝",
	} {
		if got := norm(in); got != want {
			f = append(f, fmt.Sprintf("%q: got %q want %q", in, got, want))
		}
	}
	return f
}

func TestRecordSearchNormalizationMeta(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	real := func(in string) string {
		// Observe the normalisation through the public behaviour: store, read back.
		resetGlobalTracker(t)
		t.Setenv("HOME", t.TempDir()) // a fresh database for every observation
		if err := RecordSearch(in, 1); err != nil {
			return "ERR"
		}
		got, _ := RecentSearches(1)
		if len(got) == 0 {
			return ""
		}
		return got[0].Query
	}
	require.Empty(t, recordNormalizationFailures(real), "the real RecordSearch satisfies the contract")

	mutants := map[string]func(string) string{
		"no normalisation": func(s string) string { return s },
		"trim only":        strings.TrimSpace,
		"lowercases":       func(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) },
		"first word only": func(s string) string {
			if f := strings.Fields(s); len(f) > 0 {
				return f[0]
			}
			return ""
		},
		"joins without space": func(s string) string { return strings.Join(strings.Fields(s), "") },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, recordNormalizationFailures(m), "mutant %q survived", name)
	}
}

// newestFirst is the ordering contract of ContinueEntries.
func orderingFailures(order func([]Anime) []Anime) []string {
	base := time.Unix(1_700_000_000, 0)
	in := []Anime{
		{AllanimeID: "mid", LastUpdated: base.Add(time.Hour)},
		{AllanimeID: "old", LastUpdated: base},
		{AllanimeID: "new", LastUpdated: base.Add(2 * time.Hour)},
	}
	got := order(append([]Anime(nil), in...))
	var ids []string
	for _, a := range got {
		ids = append(ids, a.AllanimeID)
	}
	if fmt.Sprint(ids) != "[new mid old]" {
		return []string{fmt.Sprintf("order %v want [new mid old]", ids)}
	}
	return nil
}

func TestContinueEntriesOrderingMeta(t *testing.T) {
	realOrder := func(in []Anime) []Anime {
		sort.SliceStable(in, func(i, j int) bool { return in[i].LastUpdated.After(in[j].LastUpdated) })
		return in
	}
	require.Empty(t, orderingFailures(realOrder))
	mutants := map[string]func([]Anime) []Anime{
		"oldest first": func(in []Anime) []Anime {
			sort.SliceStable(in, func(i, j int) bool { return in[i].LastUpdated.Before(in[j].LastUpdated) })
			return in
		},
		"unsorted": func(in []Anime) []Anime { return in },
		"by id": func(in []Anime) []Anime {
			sort.Slice(in, func(i, j int) bool { return in[i].AllanimeID < in[j].AllanimeID })
			return in
		},
		"drops one":     func(in []Anime) []Anime { return realOrder(in)[:2] },
		"reverse input": func(in []Anime) []Anime { sort.SliceStable(in, func(i, j int) bool { return i > j }); return in },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, orderingFailures(m), "mutant %q survived", name)
	}
}

// ---- dbPathFor (the per-OS layout) --------------------------------------------

func dbPathFailures(impl func(goos, home, local string) string) []string {
	var f []string
	check := func(name, got, want string) {
		if got != want {
			f = append(f, fmt.Sprintf("%s: got %q want %q", name, got, want))
		}
	}
	j := filepath.Join
	check("linux", impl("linux", "/h", ""), j("/h", ".local", "goanime", "tracking", "progress.db"))
	check("darwin", impl("darwin", "/h", ""), j("/h", ".local", "goanime", "tracking", "progress.db"))
	check("linux ignores LOCALAPPDATA", impl("linux", "/h", "/l"), j("/h", ".local", "goanime", "tracking", "progress.db"))
	check("windows LOCALAPPDATA", impl("windows", "/h", "/l"), j("/l", "GoAnime", "tracking", "progress.db"))
	check("windows fallback", impl("windows", "/h", ""), j("/h", "AppData", "Local", "GoAnime", "tracking", "progress.db"))
	return f
}

func TestDBPathFor(t *testing.T) {
	assert.Empty(t, dbPathFailures(dbPathFor))
}

func TestDBPathForMeta(t *testing.T) {
	require.Empty(t, dbPathFailures(dbPathFor))
	mutants := map[string]func(string, string, string) string{
		"windows treated as unix": func(_, h, l string) string { return dbPathFor("linux", h, l) },
		"unix treated as windows": func(_, h, l string) string { return dbPathFor("windows", h, l) },
		"ignores LOCALAPPDATA":    func(g, h, _ string) string { return dbPathFor(g, h, "") },
		"fallback ignores home":   func(g, _, l string) string { return dbPathFor(g, "/other", l) },
		"wrong file name": func(g, h, l string) string {
			return strings.Replace(dbPathFor(g, h, l), "progress.db", "tracking.db", 1)
		},
		"always set LOCALAPPDATA": func(g, h, l string) string { return dbPathFor(g, h, "/forced") },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, dbPathFailures(m), "mutant %q survived", name)
	}
}

// ---- nil / empty trackers ------------------------------------------------------

func TestTrackerMethodsRejectNilAndUninitialised(t *testing.T) {
	var nilTracker *LocalTracker
	empty := &LocalTracker{}
	for name, tr := range map[string]*LocalTracker{"nil": nilTracker, "zero value": empty} {
		assert.ErrorIs(t, tr.RecordRecentSearch("x", 1), ErrTrackerNotInited, name)
		_, err := tr.ListRecentSearches(5)
		assert.ErrorIs(t, err, ErrTrackerNotInited, name)
	}
}

func TestRecordSearchWithUninitialisedGlobalIsNoOp(t *testing.T) {
	isolatedHome(t)
	trackerMutex.Lock()
	globalTracker = &LocalTracker{}
	trackerMutex.Unlock()
	t.Cleanup(func() { trackerMutex.Lock(); globalTracker = nil; trackerMutex.Unlock() })
	assert.NoError(t, RecordSearch("naruto", 1), "no statements prepared: skip quietly, never panic")
}

func TestRecentSearchTimestampIsWholeSeconds(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	require.NoError(t, RecordSearch("bleach", 1))
	got, err := RecentSearches(1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 0, got[0].SearchedAt.Nanosecond(), "stored as unix seconds")
}

// A tracker that has a database handle but no prepared statements (a failed
// prepareStatements) must report "not initialised", not dereference nil.
func TestTrackerWithDBButNoStatementsIsNotInitialised(t *testing.T) {
	needSQLite(t)
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	half := &LocalTracker{db: db}

	assert.ErrorIs(t, half.RecordRecentSearch("x", 1), ErrTrackerNotInited)
	_, err = half.ListRecentSearches(5)
	assert.ErrorIs(t, err, ErrTrackerNotInited)

	// The package-level helper treats it as "no history": best effort, no panic.
	isolatedHome(t)
	trackerMutex.Lock()
	globalTracker = half
	trackerMutex.Unlock()
	t.Cleanup(func() { trackerMutex.Lock(); globalTracker = nil; trackerMutex.Unlock() })
	assert.NoError(t, RecordSearch("naruto", 1))
}
