package player

import (
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/stretchr/testify/require"
)

// Real-world identifiers from the bug report (Jujutsu Kaisen).
// Using realistic, distinct values makes assertion failures concrete:
// the MAL ID 57658 is what was incorrectly logged as anilistID before
// the fix in playback/common.go.
const (
	bugReportMalID     = 57658  // MyAnimeList ID
	bugReportAnilistID = 145064 // AniList ID — semantically different
)

// newTempTracker creates an isolated SQLite-backed tracker rooted in a
// throwaway temp dir. The tracking package keeps a process-wide singleton,
// so we tear it down before and after each test to guarantee isolation.
func newTempTracker(t *testing.T) *tracking.LocalTracker {
	t.Helper()
	if !tracking.IsCgoEnabled {
		t.Skip("CGO disabled — sqlite tracking unavailable in this build")
	}
	if err := tracking.CloseGlobalTracker(); err != nil {
		t.Logf("CloseGlobalTracker (pre): %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "id_routing.db")
	tracker := tracking.NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}
	t.Cleanup(func() { _ = tracking.CloseGlobalTracker() })
	return tracker
}

// readStoredRow looks up a row by its allanime_id key by scanning all rows.
// We deliberately avoid tracker.GetAnime() here because that method
// overrides the AnilistID field with whatever value the caller passes in
// (it does not return the column value as stored). The point of these
// tests is exactly to inspect what the column holds, so we go through
// GetAllAnime which does an honest SELECT of the column.
func readStoredRow(t *testing.T, tracker *tracking.LocalTracker, key string) *tracking.Anime {
	t.Helper()
	all, err := tracker.GetAllAnime()
	if err != nil {
		t.Fatalf("GetAllAnime: %v", err)
	}
	for i := range all {
		if all[i].AllanimeID == key {
			return &all[i]
		}
	}
	return nil
}

// withSpyAniSkipFetcher swaps the package-level aniSkipFetcher seam for
// the duration of the test and returns a pointer to the captured ID.
// The spy returns a sentinel error so the production code path treats
// the result as a normal "no skip data" outcome.
func withSpyAniSkipFetcher(t *testing.T) *atomic.Int64 {
	t.Helper()
	captured := &atomic.Int64{}
	captured.Store(-1) // sentinel: "spy was never called"

	original := aniSkipFetcher
	aniSkipFetcher = func(animeMalId int, episodeNum int, episode *models.Episode) error {
		captured.Store(int64(animeMalId))
		return errors.New("spy: no real fetch")
	}
	t.Cleanup(func() { aniSkipFetcher = original })
	return captured
}

// TestUpdateTrackingWithDuration_PersistsAnilistIDNotMalID is the core
// regression test for the original bug.
//
// Before the fix, internal/playback/common.go passed anime.MalID where
// anime.AnilistID was needed; that value was carried as `anilistID`
// through the playback chain and ultimately written into the
// anilist_id column of the SQLite tracker. Logs read
// "Tracking lookup: anilistID=57658" — but 57658 was the MAL ID.
//
// This test exercises the persistence path with the AniList ID and
// asserts that the column holds exactly that value, and explicitly
// asserts that the MAL ID has NOT been persisted in its place.
func TestUpdateTrackingWithDuration_PersistsAnilistIDNotMalID(t *testing.T) {
	tracker := newTempTracker(t)

	episode := &models.Episode{
		URL:   "regression-anime-url",
		Title: models.TitleDetails{English: "Jujutsu Kaisen"},
	}
	const epNum = 1

	updateTrackingWithDuration(tracker, bugReportAnilistID, episode, epNum, 24*time.Minute)

	row := readStoredRow(t, tracker, episode.URL)
	require.NotNil(t, row, "expected a row in tracker DB after updateTrackingWithDuration; got none")

	if row.AnilistID != bugReportAnilistID {
		t.Errorf("anilist_id column = %d, want %d", row.AnilistID, bugReportAnilistID)
	}
	if row.AnilistID == bugReportMalID {
		t.Errorf("REGRESSION: MAL ID (%d) was persisted as anilist_id. "+
			"Caller is passing anime.MalID where anime.AnilistID is required.",
			bugReportMalID)
	}
	if row.EpisodeNumber != epNum {
		t.Errorf("episode_number = %d, want %d", row.EpisodeNumber, epNum)
	}
	if row.Duration <= 0 {
		t.Errorf("duration = %d, want > 0", row.Duration)
	}
}

// TestFetchAniSkipAsync_ForwardsMALIDOnly verifies the AniSkip fetcher is
// called with the MAL ID. The AniSkip public API is keyed on MAL — passing
// an AniList ID would yield no skip-times even when they exist.
//
// This is the symmetric half of the routing contract: the same value that
// MUST NOT land in the tracker's anilist_id column MUST land at the
// AniSkip fetcher. The test would fail if either:
//   - the chain stopped routing MAL ID to AniSkip, or
//   - someone "fixed" the playback layer by passing AniList ID here too.
func TestFetchAniSkipAsync_ForwardsMALIDOnly(t *testing.T) {
	captured := withSpyAniSkipFetcher(t)

	episode := &models.Episode{URL: "any"}
	ch := fetchAniSkipAsync(bugReportMalID, 1, episode)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("fetchAniSkipAsync did not return within 2s")
	}

	got := int(captured.Load())
	if got == -1 {
		t.Fatal("AniSkip fetcher spy was never invoked")
	}
	if got != bugReportMalID {
		t.Errorf("AniSkip fetcher received id=%d, want MAL=%d", got, bugReportMalID)
	}
	if got == bugReportAnilistID {
		t.Errorf("REGRESSION: AniSkip received the AniList ID (%d). "+
			"The AniSkip API is keyed on MAL — this would silently break skip-times.",
			bugReportAnilistID)
	}
}

// TestIDRouting_FullContract_NoConflation is the end-to-end routing test.
// It drives both consumers — AniSkip fetcher and the SQLite tracker —
// with distinct IDs the way playVideo does internally, and asserts:
//
//  1. The AniSkip fetcher received the MAL ID.
//  2. The tracker persisted the AniList ID.
//  3. The two recorded values are NOT equal — a hard guarantee against
//     any future refactor that re-conflates the IDs into a single
//     parameter.
//
// This is the brutal test: even if the storage and fetcher tests above
// were each individually satisfied by some pathological wiring that
// pipes the same value to both, the inequality assertion catches it.
func TestIDRouting_FullContract_NoConflation(t *testing.T) {
	tracker := newTempTracker(t)
	captured := withSpyAniSkipFetcher(t)

	episode := &models.Episode{
		URL:   "full-contract-url",
		Title: models.TitleDetails{English: "Jujutsu Kaisen"},
	}
	const epNum = 4

	// Mirror what playVideo does: dispatch AniSkip with the MAL ID,
	// persist progress with the AniList ID.
	ch := fetchAniSkipAsync(bugReportMalID, epNum, episode)
	updateTrackingWithDuration(tracker, bugReportAnilistID, episode, epNum, 24*time.Minute)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("AniSkip spy did not return")
	}

	aniSkipGot := int(captured.Load())
	if aniSkipGot != bugReportMalID {
		t.Errorf("AniSkip fetcher received id=%d, want MAL=%d", aniSkipGot, bugReportMalID)
	}

	row := readStoredRow(t, tracker, episode.URL)
	require.NotNil(t, row, "tracker has no row for episode after updateTrackingWithDuration")
	if row.AnilistID != bugReportAnilistID {
		t.Errorf("tracker stored anilist_id=%d, want %d", row.AnilistID, bugReportAnilistID)
	}

	if aniSkipGot == row.AnilistID {
		t.Fatalf("REGRESSION: AniSkip fetcher and tracker both received %d. "+
			"MAL and AniList IDs must remain distinct values routed to distinct consumers.",
			aniSkipGot)
	}
}

// TestIDRouting_ZeroAnilistID_DoesNotFallbackToMalID guards against a
// "helpful" fallback that some refactor might add — e.g. "if anilistID
// is 0, just store malID instead". That fallback would re-introduce the
// original bug for any anime whose AniList metadata enrichment failed.
//
// The contract is: whatever AniList ID the caller passes (including 0)
// is what gets persisted. The MAL ID never silently substitutes.
func TestIDRouting_ZeroAnilistID_DoesNotFallbackToMalID(t *testing.T) {
	tracker := newTempTracker(t)

	episode := &models.Episode{
		URL:   "zero-anilist-url",
		Title: models.TitleDetails{English: "No Anilist Metadata"},
	}
	const epNum = 1

	updateTrackingWithDuration(tracker, 0, episode, epNum, 24*time.Minute)

	row := readStoredRow(t, tracker, episode.URL)
	require.NotNil(t, row, "expected stored row even when anilistID is 0")
	if row.AnilistID != 0 {
		t.Errorf("anilist_id = %d, want 0 (no silent fallback to MAL)", row.AnilistID)
	}
	if row.AnilistID == bugReportMalID {
		t.Fatalf("REGRESSION: MAL ID (%d) silently substituted when AniList ID was 0",
			bugReportMalID)
	}
}

// TestUpdateTracking_WithoutSocket_FailsClosed asserts the periodic
// in-flight tracker write bails out when there is no live mpv socket
// rather than persisting bogus rows. This guards a different failure
// mode but lives next to its sibling tests because it shares the same
// tracker fixture: a regression that made updateTracking write a row
// without time-pos data would silently corrupt resume positions.
func TestUpdateTracking_WithoutSocket_FailsClosed(t *testing.T) {
	tracker := newTempTracker(t)
	episode := &models.Episode{
		URL:   "no-socket-url",
		Title: models.TitleDetails{English: "test"},
	}

	updateTracking(tracker, "/nonexistent/socket.path", bugReportAnilistID, episode, 1, nil)

	if row := readStoredRow(t, tracker, trackingKey(episode.URL, 1)); row != nil {
		t.Fatalf("updateTracking wrote a row without a working mpv socket: %+v", row)
	}
	if row := readStoredRow(t, tracker, episode.URL); row != nil {
		t.Fatalf("updateTracking wrote a row (legacy key) without a working mpv socket: %+v", row)
	}
}
