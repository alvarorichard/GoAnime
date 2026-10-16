package tracking

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestTracker(t *testing.T) *LocalTracker {
	t.Helper()
	tracker := NewLocalTracker(filepath.Join(t.TempDir(), "test.db"))
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	t.Cleanup(func() { _ = tracker.Close() })
	return tracker
}

func TestAnimeContextRoundTrip(t *testing.T) {
	tracker := openTestTracker(t)

	entry := Anime{
		AllanimeID:    "https://animefire.io/video/x:ep5",
		EpisodeNumber: 5,
		PlaybackTime:  372,
		Duration:      580,
		Title:         "Urashima Tarou",
		AnimeTitle:    "Mukashi Banashi",
		Source:        "Animefire.io",
		Audio:         "Legendado",
		MediaType:     "anime",
		LastUpdated:   time.Now(),
	}
	require.NoError(t, tracker.UpdateProgress(entry))

	got, err := tracker.GetAnime(0, entry.AllanimeID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Mukashi Banashi", got.AnimeTitle)
	assert.Equal(t, "Animefire.io", got.Source)
	assert.Equal(t, "Legendado", got.Audio)
	assert.Equal(t, "Urashima Tarou", got.Title)
	assert.Equal(t, "anime", got.MediaType, "media type survives the round trip for continue routing")

	all, err := tracker.GetAllAnime()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "Mukashi Banashi", all[0].AnimeTitle)
}

func TestAnimeContextUpdateKeepsKnownValues(t *testing.T) {
	tracker := openTestTracker(t)

	require.NoError(t, tracker.UpdateProgress(Anime{
		AllanimeID: "key", EpisodeNumber: 1, PlaybackTime: 10, Duration: 100,
		Title: "Ep", AnimeTitle: "Show", Source: "HiAnime", Audio: "Sub",
		LastUpdated: time.Now(),
	}))
	// A later write without context (duration-only row at episode start)
	// must not blank what is already known.
	require.NoError(t, tracker.UpdateProgress(Anime{
		AllanimeID: "key", EpisodeNumber: 1, PlaybackTime: 0, Duration: 100,
		Title: "Ep", LastUpdated: time.Now(),
	}))

	got, err := tracker.GetAnime(0, "key")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Show", got.AnimeTitle)
	assert.Equal(t, "HiAnime", got.Source)
	assert.Equal(t, "Sub", got.Audio)
}

func TestMigrateOldDatabaseGainsColumns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.db")

	db, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE media_progress (
		allanime_id TEXT PRIMARY KEY NOT NULL, anilist_id INTEGER DEFAULT 0,
		episode_number INTEGER NOT NULL, playback_time INTEGER NOT NULL,
		duration INTEGER NOT NULL, title TEXT, media_type TEXT DEFAULT 'anime',
		last_updated INTEGER NOT NULL)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO media_progress VALUES ('k',0,1,60,600,'Ep','anime',?)`, time.Now().Unix())
	require.NoError(t, err)
	require.NoError(t, db.Close())

	tracker := NewLocalTracker(path)
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	t.Cleanup(func() { _ = tracker.Close() })

	got, err := tracker.GetAnime(0, "k")
	require.NoError(t, err)
	require.NotNil(t, got, "old rows stay readable after migration")
	assert.Equal(t, "Ep", got.Title)
	assert.Empty(t, got.AnimeTitle)

	require.NoError(t, tracker.UpdateProgress(Anime{
		AllanimeID: "k2", EpisodeNumber: 2, PlaybackTime: 5, Duration: 600,
		Title: "Ep2", AnimeTitle: "Show", Source: "HiAnime", LastUpdated: time.Now(),
	}))
	got2, err := tracker.GetAnime(0, "k2")
	require.NoError(t, err)
	require.NotNil(t, got2)
	assert.Equal(t, "Show", got2.AnimeTitle)
}

func TestRecentSearches(t *testing.T) {
	tracker := openTestTracker(t)

	_, err := tracker.ListRecentSearches(5)
	require.NoError(t, err)

	require.NoError(t, tracker.RecordRecentSearch("momotarou", 3))
	require.NoError(t, tracker.RecordRecentSearch("namakura gatana", 5))
	require.NoError(t, tracker.RecordRecentSearch("  ", 0), "blank queries are ignored")
	require.NoError(t, tracker.RecordRecentSearch("momotarou", 4), "re-search updates the count")

	got, err := tracker.ListRecentSearches(5)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "momotarou", got[0].Query, "newest first")
	assert.Equal(t, 4, got[0].ResultCount)
	assert.Equal(t, "namakura gatana", got[1].Query)
	assert.Equal(t, 5, got[1].ResultCount)

	one, err := tracker.ListRecentSearches(1)
	require.NoError(t, err)
	assert.Len(t, one, 1)

	none, err := tracker.ListRecentSearches(0)
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestDBPath(t *testing.T) {
	assert.NotEmpty(t, DBPath())
	assert.Contains(t, DBPath(), "progress.db")
}
