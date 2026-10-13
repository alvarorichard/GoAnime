package player

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func progressEpisodes() []models.Episode {
	return []models.Episode{
		{Number: "1", URL: "http://show"},
		{Number: "2", URL: "http://show"},
		{Number: "3", URL: "http://show"},
		{Number: "4", URL: "http://show"},
	}
}

func TestMatchEpisodeProgress(t *testing.T) {
	t.Parallel()
	at := time.Unix(1_700_000_000, 0)
	saved := []tracking.Anime{
		// Written by updateTracking, keyed per episode.
		{AllanimeID: trackingKey("http://show", 1), EpisodeNumber: 1, PlaybackTime: 1400, Duration: 1440, LastUpdated: at},
		{AllanimeID: trackingKey("http://show", 2), EpisodeNumber: 2, PlaybackTime: 300, Duration: 1440, LastUpdated: at},
		// Written when an episode starts: a duration and no position.
		{AllanimeID: trackingKey("http://show", 3), EpisodeNumber: 3, Duration: 1440},
		// A row whose key names one episode and whose number another.
		{AllanimeID: trackingKey("http://show", 4), EpisodeNumber: 7, PlaybackTime: 60, Duration: 1440},
		// The legacy key without the episode number is not a per-episode row.
		{AllanimeID: "http://show", EpisodeNumber: 4, PlaybackTime: 900, Duration: 1440},
	}

	got := matchEpisodeProgress(progressEpisodes(), saved)

	require.Len(t, got, 2)
	assert.Equal(t, episodeWatch{position: 1400, duration: 1440, updated: at}, got[0])
	assert.Equal(t, 300, got[1].position)
}

func TestEpisodeWatchFinished(t *testing.T) {
	t.Parallel()
	assert.True(t, episodeWatch{position: 1300, duration: 1440}.finished())
	assert.False(t, episodeWatch{position: 1200, duration: 1440}.finished())
	assert.False(t, episodeWatch{position: 10}.finished(), "no duration, no way to tell")
}

func TestLastWatched(t *testing.T) {
	t.Parallel()
	at := time.Unix(1_700_000_000, 0)

	_, ok := lastWatched(nil)
	assert.False(t, ok)

	last, ok := lastWatched(map[int]episodeWatch{0: {updated: at}, 2: {updated: at.Add(time.Hour)}, 1: {updated: at}})
	assert.True(t, ok)
	assert.Equal(t, 2, last)

	last, _ = lastWatched(map[int]episodeWatch{0: {updated: at}, 3: {updated: at}})
	assert.Equal(t, 3, last, "a tie goes to the later episode, whatever the map order")
}

func TestApplyEpisodeProgress(t *testing.T) {
	t.Parallel()
	at := time.Unix(1_700_000_000, 0)
	episodes := progressEpisodes()

	t.Run("stopped partway: marks, cursor and panel", func(t *testing.T) {
		t.Parallel()
		items := episodePickItems(episodes)
		opts := tui.PickOptions{}
		applyEpisodeProgress(items, &opts, episodes, map[int]episodeWatch{
			0: {position: 1400, duration: 1440, updated: at},
			1: {position: 872, duration: 1440, updated: at.Add(time.Hour)},
		})

		assert.Equal(t, tui.PickWatched, items[0].Status)
		assert.Equal(t, tui.PickInProgress, items[1].Status)
		assert.Equal(t, "stopped at 14:32", items[1].Note)
		assert.Equal(t, tui.PickNone, items[2].Status)
		assert.Equal(t, 1, opts.InitialIndex)
		require.NotNil(t, opts.Panel)
		assert.Equal(t, "2", opts.Panel.Title)
		position, duration, ok := opts.Panel.Progress()
		assert.True(t, ok)
		assert.InDelta(t, 872, position, 0)
		assert.InDelta(t, 1440, duration, 0)
	})

	t.Run("finished last: the next episode is preselected", func(t *testing.T) {
		t.Parallel()
		items := episodePickItems(episodes)
		opts := tui.PickOptions{}
		applyEpisodeProgress(items, &opts, episodes, map[int]episodeWatch{
			1: {position: 300, duration: 1440, updated: at},
			2: {position: 1430, duration: 1440, updated: at.Add(time.Hour)},
		})
		assert.Equal(t, 3, opts.InitialIndex)
		assert.Nil(t, opts.Panel, "an older unfinished episode is marked, not offered")
		assert.Equal(t, tui.PickInProgress, items[1].Status)
	})

	t.Run("finished the last episode: the cursor stays", func(t *testing.T) {
		t.Parallel()
		items := episodePickItems(episodes)
		opts := tui.PickOptions{}
		applyEpisodeProgress(items, &opts, episodes, map[int]episodeWatch{3: {position: 1430, duration: 1440, updated: at}})
		assert.Equal(t, 0, opts.InitialIndex)
		assert.Equal(t, tui.PickWatched, items[3].Status)
	})

	t.Run("nothing watched", func(t *testing.T) {
		t.Parallel()
		items := episodePickItems(episodes)
		opts := tui.PickOptions{}
		applyEpisodeProgress(items, &opts, episodes, nil)
		assert.Equal(t, 0, opts.InitialIndex)
		assert.Nil(t, opts.Panel)
	})
}

func TestSelectEpisodeWithPicker_ProgressReachesThePicker(t *testing.T) {
	t.Parallel()
	episodes := progressEpisodes()
	watched := map[int]episodeWatch{1: {position: 60, duration: 1440, updated: time.Unix(1_700_000_000, 0)}}

	_, num, err := selectEpisodeWithPicker(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
		assert.Equal(t, "play", opts.Action)
		assert.Equal(t, 1, opts.InitialIndex)
		assert.Equal(t, tui.PickInProgress, items[1].Status)
		return opts.InitialIndex, nil
	}, episodes, watched)

	require.NoError(t, err)
	assert.Equal(t, "2", num)
}
