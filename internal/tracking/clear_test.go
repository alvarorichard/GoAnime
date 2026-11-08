package tracking

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Forgetting history: one recent search, every recent search, every watched
// episode. The two lists live in two tables and clearing one must never touch
// the other.

func TestNormalizeQuery(t *testing.T) {
	for in, want := range map[string]string{
		"naruto":           "naruto",
		"  naruto  ":       "naruto",
		"one   piece":      "one piece",
		"\tboku\nno hero ": "boku no hero",
		"":                 "",
		"   ":              "",
	} {
		assert.Equal(t, want, normalizeQuery(in), "%q", in)
	}
}

// fakeResult is a sql.Result that answers what a test says.
type fakeResult struct {
	rows int64
	err  error
}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }

func (r fakeResult) RowsAffected() (int64, error) { return r.rows, r.err }

func TestAffected(t *testing.T) {
	boom := errors.New("database is locked")

	n, err := affected(fakeResult{rows: 3}, nil)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = affected(fakeResult{rows: 3}, boom)
	require.ErrorIs(t, err, boom, "a failed statement is an error")
	assert.Zero(t, n, "and touched nothing as far as the caller can tell")

	n, err = affected(fakeResult{rows: 3, err: boom}, nil)
	require.ErrorIs(t, err, boom, "a driver that cannot count says so")
	assert.Equal(t, 3, n)
}

func seedHistory(t *testing.T, tracker *LocalTracker) {
	t.Helper()
	for _, q := range []string{"naruto", "one piece", "bleach"} {
		require.NoError(t, tracker.RecordRecentSearch(q, 4))
	}
	for i, id := range []string{"ep1", "ep2"} {
		require.NoError(t, tracker.UpdateProgress(Anime{
			AllanimeID: id, EpisodeNumber: i + 1, PlaybackTime: 60, Duration: 1400,
			AnimeTitle: "Show", MediaType: "anime", LastUpdated: time.Now(),
		}))
	}
}

func recentQueries(t *testing.T, tracker *LocalTracker) []string {
	t.Helper()
	list, err := tracker.ListRecentSearches(50)
	require.NoError(t, err)
	var out []string
	for _, r := range list {
		out = append(out, r.Query)
	}
	return out
}

func TestDeleteRecentSearch(t *testing.T) {
	tracker := openTestTracker(t)
	seedHistory(t, tracker)

	gone, err := tracker.DeleteRecentSearch("one piece")
	require.NoError(t, err)
	assert.True(t, gone)
	assert.ElementsMatch(t, []string{"naruto", "bleach"}, recentQueries(t, tracker), "only that search goes")

	all, err := tracker.GetAllAnime()
	require.NoError(t, err)
	assert.Len(t, all, 2, "watched episodes are not searches")
}

func TestDeleteRecentSearchFindsItBothWaysItIsSpelled(t *testing.T) {
	tracker := openTestTracker(t)
	require.NoError(t, tracker.RecordRecentSearch("one   piece", 1), "stored with single spaces")

	gone, err := tracker.DeleteRecentSearch("  one  piece ")

	require.NoError(t, err)
	assert.True(t, gone, "the row on screen and the one typed are the same search")
	assert.Empty(t, recentQueries(t, tracker))
}

func TestDeleteRecentSearchThatIsNotThere(t *testing.T) {
	tracker := openTestTracker(t)
	seedHistory(t, tracker)

	gone, err := tracker.DeleteRecentSearch("dragon ball")

	require.NoError(t, err, "forgetting what was never there is not a failure")
	assert.False(t, gone)
	assert.Len(t, recentQueries(t, tracker), 3)

	gone, err = tracker.DeleteRecentSearch("   ")
	require.NoError(t, err)
	assert.False(t, gone, "a blank search matches nothing")
	assert.Len(t, recentQueries(t, tracker), 3)
}

func TestClearRecentSearches(t *testing.T) {
	tracker := openTestTracker(t)
	seedHistory(t, tracker)

	n, err := tracker.ClearRecentSearches()

	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Empty(t, recentQueries(t, tracker))
	all, err := tracker.GetAllAnime()
	require.NoError(t, err)
	assert.Len(t, all, 2, "CONTINUE is not RECENT")

	n, err = tracker.ClearRecentSearches()
	require.NoError(t, err)
	assert.Zero(t, n, "nothing left to clear")
}

func TestClearProgress(t *testing.T) {
	tracker := openTestTracker(t)
	seedHistory(t, tracker)

	n, err := tracker.ClearProgress()

	require.NoError(t, err)
	assert.Equal(t, 2, n)
	all, err := tracker.GetAllAnime()
	require.NoError(t, err)
	assert.Empty(t, all)
	assert.Len(t, recentQueries(t, tracker), 3, "RECENT is not CONTINUE")

	n, err = tracker.ClearProgress()
	require.NoError(t, err)
	assert.Zero(t, n)
}

func TestClearingOnATrackerThatIsNotThere(t *testing.T) {
	var none *LocalTracker
	closed := &LocalTracker{}

	for name, tracker := range map[string]*LocalTracker{"nil": none, "no database": closed} {
		_, err := tracker.DeleteRecentSearch("x")
		assert.ErrorIs(t, err, ErrTrackerNotInited, name)
		_, err = tracker.ClearRecentSearches()
		assert.ErrorIs(t, err, ErrTrackerNotInited, name)
		_, err = tracker.ClearProgress()
		assert.ErrorIs(t, err, ErrTrackerNotInited, name)
	}
}

func TestClearingOnAClosedDatabaseReportsTheError(t *testing.T) {
	tracker := NewLocalTracker(t.TempDir() + "/closed.db")
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	require.NoError(t, tracker.Close())

	_, err := tracker.DeleteRecentSearch("x")
	assert.Error(t, err)
	_, err = tracker.ClearRecentSearches()
	assert.Error(t, err)
	_, err = tracker.ClearProgress()
	assert.Error(t, err)
}

// ---- the package-level helpers the screen and the command line call ----------

func TestForgetSearchClearRecentAndClearContinueOnTheSharedDatabase(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	require.NoError(t, RecordSearch("naruto", 2))
	require.NoError(t, RecordSearch("bleach", 5))
	tracker := sharedTracker()
	require.NotNil(t, tracker)
	require.NoError(t, tracker.UpdateProgress(Anime{AllanimeID: "ep1", EpisodeNumber: 1, PlaybackTime: 60, Duration: 1400, AnimeTitle: "Show", MediaType: "anime", LastUpdated: time.Now()}))

	require.NoError(t, ForgetSearch("naruto"))
	left, err := RecentSearches(10)
	require.NoError(t, err)
	require.Len(t, left, 1)
	assert.Equal(t, "bleach", left[0].Query)

	n, err := ClearRecent()
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	left, err = RecentSearches(10)
	require.NoError(t, err)
	assert.Empty(t, left)
	entries, err := ContinueEntries()
	require.NoError(t, err)
	assert.Len(t, entries, 1, "clearing RECENT leaves CONTINUE")

	n, err = ClearContinue()
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	entries, err = ContinueEntries()
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestClearingWithoutHistoryIsReportedNotSwallowed(t *testing.T) {
	needSQLite(t)
	isolatedHome(t)
	withoutCGO(t)

	assert.ErrorIs(t, ForgetSearch("naruto"), ErrHistoryUnavailable)
	_, err := ClearRecent()
	assert.ErrorIs(t, err, ErrHistoryUnavailable)
	_, err = ClearContinue()
	assert.ErrorIs(t, err, ErrHistoryUnavailable)
}
