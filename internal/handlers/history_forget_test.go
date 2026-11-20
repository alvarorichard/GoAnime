package handlers

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The CONTINUE list shows one row per show, but the history keeps one row per
// episode watched. Removing a show's row must remove the show.

// historyHome points HOME at a throwaway directory so the real history code
// runs against a database of its own.
func historyHome(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("DBPath on Windows follows LOCALAPPDATA")
	}
	if !tracking.IsCgoEnabled {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	require.NoError(t, tracking.CloseGlobalTracker())
	t.Cleanup(func() { _ = tracking.CloseGlobalTracker() })
}

func watched(t *testing.T, show, key string, episode int, at time.Time) {
	t.Helper()
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		tracker = tracking.NewLocalTracker(tracking.DBPath())
	}
	require.NotNil(t, tracker)
	require.NoError(t, tracker.UpdateProgress(tracking.Anime{
		AllanimeID: key, EpisodeNumber: episode, PlaybackTime: 120, Duration: 1400,
		AnimeTitle: show, Source: "Animefire.io", MediaType: "anime", LastUpdated: at,
	}))
}

func homeRows(t *testing.T) []tui.HomeContinueEntry {
	t.Helper()
	rows, err := tracking.ContinueEntries()
	require.NoError(t, err)
	return groupContinueEntries(rows)
}

func TestRemovingAShowFromContinueRemovesEveryEpisodeOfIt(t *testing.T) {
	historyHome(t)
	now := time.Now()
	watched(t, "Wind Breaker", "wb:ep1", 1, now.Add(-3*time.Hour))
	watched(t, "Wind Breaker", "wb:ep2", 2, now.Add(-2*time.Hour))
	watched(t, "Wind Breaker", "wb:ep3", 3, now.Add(-1*time.Hour))
	watched(t, "Naruto", "naruto:ep1", 1, now.Add(-5*time.Hour))
	rows := homeRows(t)
	require.Len(t, rows, 2, "one row per show")
	require.Equal(t, "wb:ep3", rows[0].Key, "the latest episode stands for the show")

	require.NoError(t, deleteContinueEntry(rows[0].Key))

	left := homeRows(t)
	require.Len(t, left, 1, "Wind Breaker does not come back with an older episode")
	assert.Equal(t, "Naruto", left[0].AnimeTitle)
}

func TestKeysOfShow(t *testing.T) {
	row := func(key, show, source string) tracking.Anime {
		return tracking.Anime{AllanimeID: key, AnimeTitle: show, Source: source}
	}
	rows := []tracking.Anime{
		row("wb:ep1", "Wind Breaker", "Animefire.io"),
		row("wb:ep2", "Wind Breaker", "Animefire.io"),
		row("wb-dub:ep1", "Wind Breaker", "Goyabu"),
		row("naruto:ep1", "Naruto", "Animefire.io"),
		row("https://site/series/x:ep1", "", ""),
		row("https://site/series/x:ep2", "", ""),
		row("https://site/series/y:ep1", "", ""),
	}

	assert.ElementsMatch(t, []string{"wb:ep1", "wb:ep2"}, keysOfShow(rows, "wb:ep2"), "the show on that source")
	assert.Equal(t, []string{"wb-dub:ep1"}, keysOfShow(rows, "wb-dub:ep1"), "the same title on another source is another row")
	assert.Equal(t, []string{"naruto:ep1"}, keysOfShow(rows, "naruto:ep1"))
	assert.ElementsMatch(t, []string{"https://site/series/x:ep1", "https://site/series/x:ep2"}, keysOfShow(rows, "https://site/series/x:ep2"), "rows with no show title group by their URL")
	assert.Equal(t, []string{"gone"}, keysOfShow(rows, "gone"), "a key that is not in the history is just itself")
	assert.Equal(t, []string{"anything"}, keysOfShow(nil, "anything"))
}

// historyFake stands in for the stored history.
type historyFake struct {
	rows       []tracking.Anime
	listErr    error
	deleteErr  map[string]error
	deleted    []string
	forgotten  []string
	forgetErr  error
	clearedCon int
	clearedRec int
	clearErr   error
}

func useHistoryFake(t *testing.T, h *historyFake) {
	t.Helper()
	prevRows, prevDelete, prevForget, prevCon, prevRec := continueRows, deleteTracked, forgetSearchFn, clearContinueFn, clearRecentFn
	continueRows = func() ([]tracking.Anime, error) { return h.rows, h.listErr }
	deleteTracked = func(key string) error {
		h.deleted = append(h.deleted, key)
		return h.deleteErr[key]
	}
	forgetSearchFn = func(query string) error { h.forgotten = append(h.forgotten, query); return h.forgetErr }
	clearContinueFn = func() (int, error) { h.clearedCon++; return 7, h.clearErr }
	clearRecentFn = func() (int, error) { h.clearedRec++; return 4, h.clearErr }
	t.Cleanup(func() {
		continueRows, deleteTracked, forgetSearchFn, clearContinueFn, clearRecentFn = prevRows, prevDelete, prevForget, prevCon, prevRec
	})
}

func TestDeleteContinueEntryRemovesEveryEpisodeOfTheShow(t *testing.T) {
	h := &historyFake{rows: []tracking.Anime{
		{AllanimeID: "a:ep1", AnimeTitle: "A"}, {AllanimeID: "a:ep2", AnimeTitle: "A"}, {AllanimeID: "b:ep1", AnimeTitle: "B"},
	}}
	useHistoryFake(t, h)

	require.NoError(t, deleteContinueEntry("a:ep2"))

	assert.ElementsMatch(t, []string{"a:ep1", "a:ep2"}, h.deleted)
}

func TestDeleteContinueEntryStopsAtTheFirstFailure(t *testing.T) {
	h := &historyFake{
		rows:      []tracking.Anime{{AllanimeID: "a:ep1", AnimeTitle: "A"}, {AllanimeID: "a:ep2", AnimeTitle: "A"}},
		deleteErr: map[string]error{"a:ep1": errors.New("database is locked")},
	}
	useHistoryFake(t, h)

	err := deleteContinueEntry("a:ep2")

	require.ErrorContains(t, err, "database is locked", "the screen keeps the row and says why")
	assert.Len(t, h.deleted, 1, "no more rows are touched after one failed")
}

func TestDeleteContinueEntryReportsAHistoryItCannotRead(t *testing.T) {
	h := &historyFake{listErr: errors.New("no such table")}
	useHistoryFake(t, h)

	err := deleteContinueEntry("a:ep1")

	require.ErrorContains(t, err, "no such table")
	assert.Empty(t, h.deleted, "nothing is deleted blind")
}

func TestForgetRecentSearchAndTheClearsPassThroughWhatHistoryAnswers(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)

	require.NoError(t, forgetRecentSearch("naruto"))
	require.NoError(t, clearContinueHistory())
	require.NoError(t, clearRecentHistory())

	assert.Equal(t, []string{"naruto"}, h.forgotten)
	assert.Equal(t, 1, h.clearedCon)
	assert.Equal(t, 1, h.clearedRec)

	h.forgetErr, h.clearErr = errors.New("locked"), errors.New("locked")
	assert.ErrorContains(t, forgetRecentSearch("x"), "locked")
	assert.ErrorContains(t, clearContinueHistory(), "locked")
	assert.ErrorContains(t, clearRecentHistory(), "locked")
}

func TestClearContinueAndRecentAreSeparateCalls(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)

	require.NoError(t, clearContinueHistory())
	assert.Equal(t, 0, h.clearedRec)

	require.NoError(t, clearRecentHistory())
	assert.Equal(t, 1, h.clearedCon)
}

func TestTheHomeScreenIsGivenTheHistoryActions(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)
	isolateHome(t)
	var got tui.SearchHomeOptions

	_, err := openSearchHome("", func(opts tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		got = opts
		return tui.SearchHomeResult{}, nil
	})
	require.NoError(t, err)

	require.NotNil(t, got.DeleteEntry)
	require.NotNil(t, got.DeleteRecent)
	require.NotNil(t, got.ClearContinue)
	require.NotNil(t, got.ClearRecent)

	require.NoError(t, got.DeleteRecent("bleach"))
	assert.Equal(t, []string{"bleach"}, h.forgotten, "x on a RECENT row forgets that search")
	assert.Zero(t, h.clearedCon+h.clearedRec)

	require.NoError(t, got.ClearContinue())
	assert.Equal(t, 1, h.clearedCon)
	assert.Zero(t, h.clearedRec, "X on CONTINUE clears CONTINUE only")

	require.NoError(t, got.ClearRecent())
	assert.Equal(t, 1, h.clearedRec)
	assert.Equal(t, 1, h.clearedCon, "X on RECENT clears RECENT only")

	require.NoError(t, got.DeleteEntry("a:ep1"))
	assert.Equal(t, []string{"a:ep1"}, h.deleted, "x on a CONTINUE row goes through the show lookup")
}

func TestDeleteTrackedRowRemovesTheRow(t *testing.T) {
	historyHome(t)
	watched(t, "Naruto", "naruto:ep1", 1, time.Now())
	require.Len(t, homeRows(t), 1)

	require.NoError(t, deleteTrackedRow("naruto:ep1"))

	assert.Empty(t, homeRows(t))
}

func TestDeleteTrackedRowWithoutHistoryReportsIt(t *testing.T) {
	isolateHome(t)
	require.NoError(t, tracking.CloseGlobalTracker())
	prev := tracking.NewLocalTracker
	tracking.NewLocalTracker = func(string) *tracking.LocalTracker { return nil }
	t.Cleanup(func() { tracking.NewLocalTracker = prev })

	err := deleteTrackedRow("x")

	require.ErrorContains(t, err, "watch history unavailable")
}
