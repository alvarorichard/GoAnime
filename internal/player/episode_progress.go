package player

import (
	"strconv"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// episodeWatch is what the tracker saved for one episode.
type episodeWatch struct {
	position, duration int // seconds
	updated            time.Time
}

// finished reports whether the episode was watched into its last tenth,
// where the end credits run.
func (w episodeWatch) finished() bool {
	return w.duration > 0 && w.position >= w.duration*9/10
}

// loadEpisodeProgress reads the position the tracker saved for each episode,
// keyed by the episode's index. Episodes never played are absent, and so is
// everything when the build has no SQLite (no CGO).
func loadEpisodeProgress(episodes []models.Episode) map[int]episodeWatch {
	if !tracking.IsCgoEnabled || len(episodes) == 0 {
		return nil
	}
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		path := getTrackerDBPath()
		if path == "" {
			return nil
		}
		tracker = tracking.NewLocalTracker(path)
	}
	if tracker == nil {
		return nil
	}
	saved, err := tracker.GetAllAnime()
	if err != nil {
		util.Debugf("Episode progress: %v", err)
		return nil
	}
	return matchEpisodeProgress(episodes, saved)
}

// matchEpisodeProgress pairs saved rows with episodes by the key playback
// writes them under (trackingKey). Rows without a position are skipped: the
// row written when an episode starts holds its duration only.
func matchEpisodeProgress(episodes []models.Episode, saved []tracking.Anime) map[int]episodeWatch {
	byKey := make(map[string]tracking.Anime, len(saved))
	for _, a := range saved {
		byKey[a.AllanimeID] = a
	}
	watched := make(map[int]episodeWatch)
	for i, ep := range episodes {
		num, err := strconv.Atoi(ExtractEpisodeNumber(ep.Number))
		if err != nil {
			continue
		}
		a, ok := byKey[trackingKey(ep.URL, num)]
		if !ok || a.EpisodeNumber != num || a.PlaybackTime <= 0 {
			continue
		}
		watched[i] = episodeWatch{position: a.PlaybackTime, duration: a.Duration, updated: a.LastUpdated}
	}
	return watched
}

// lastWatched returns the index of the episode played most recently, if any.
func lastWatched(watched map[int]episodeWatch) (int, bool) {
	last, found := -1, false
	for i, w := range watched {
		if !found || w.updated.After(watched[last].updated) || (w.updated.Equal(watched[last].updated) && i > last) {
			last, found = i, true
		}
	}
	return last, found
}

// applyEpisodeProgress marks the watched episodes, opens the list where the
// user left off and, when that episode was stopped partway, says so in a
// panel: the episode played last if it was not finished, else the one after
// it.
func applyEpisodeProgress(items []tui.PickItem, opts *tui.PickOptions, episodes []models.Episode, watched map[int]episodeWatch) {
	for i, w := range watched {
		if i >= len(items) {
			continue
		}
		if w.finished() {
			items[i].Status = tui.PickWatched
			continue
		}
		items[i].Status = tui.PickInProgress
		items[i].Note = "stopped at " + tui.FormatClock(float64(w.position))
	}

	last, ok := lastWatched(watched)
	if !ok || last >= len(episodes) {
		return
	}
	w := watched[last]
	if w.finished() {
		if last+1 < len(episodes) {
			opts.InitialIndex = last + 1
		}
		return
	}
	opts.InitialIndex = last
	opts.Panel = &tui.PickPanel{
		Label: "Continue",
		Gloss: "続き",
		Title: episodeLabel(episodes[last]),
		Lines: []string{"Playing it again offers to resume from where it stopped."},
		Progress: func() (position, duration float64, ok bool) {
			return float64(w.position), float64(w.duration), true
		},
	}
}
