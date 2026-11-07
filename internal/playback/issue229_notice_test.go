package playback

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #229: Wind Breaker on AnimeFire has no video files. GoAnime found that
// out and logged it, and then opened the episode list again over the message:
// the user saw the list come back "after a few milliseconds" with no word of
// why, and no hint that ESC leaves it. The reason now sits on the list itself.
// Not parallel: the notice seam is package-level.

// captureNotices swaps the picker-notice seam and returns what was said.
func captureNotices(t *testing.T) *[]string {
	t.Helper()
	var said []string
	prev := noticeOnNextPicker
	noticeOnNextPicker = func(text string) { said = append(said, text) }
	t.Cleanup(func() { noticeOnNextPicker = prev })
	return &said
}

func TestSourceVideoRemovedMessage(t *testing.T) {
	problem, hint := sourceVideoRemovedMessage(&models.Anime{Source: "Animefire.io"}, "Episódio 01")

	assert.Equal(t, "Episódio 01 is no longer playable on Animefire.io: the host deleted the video file (the episode is still listed, the media behind it is gone).", problem)
	assert.Equal(t, "Press ESC to leave the episode list and pick the same title from another source: StartFlix.", hint)
}

func TestSourceVideoRemovedMessageFallsBackToWhatItKnows(t *testing.T) {
	t.Run("no anime and a blank label", func(t *testing.T) {
		problem, hint := sourceVideoRemovedMessage(nil, "  ")
		assert.Equal(t, "This episode is no longer playable on this source: the host deleted the video file (the episode is still listed, the media behind it is gone).", problem)
		assert.Contains(t, hint, "pick the same title from another source:", "an unknown source can still be rescued by the others")
	})
	t.Run("no other source to offer", func(t *testing.T) {
		t.Setenv("GOANIME_DISABLED_SOURCES", "animefire,startflix")
		_, hint := sourceVideoRemovedMessage(&models.Anime{Source: "Goyabu"}, "1")
		assert.Equal(t, "Press ESC to leave the episode list and search again.", hint)
	})
}

func TestReportSourceVideoRemovedPutsTheReasonOnTheList(t *testing.T) {
	said := captureNotices(t)

	reportSourceVideoRemoved(&models.Anime{Name: "Wind Breaker", Source: "Animefire.io"}, "1")

	require.Len(t, *said, 1, "one notice for the list that opens next")
	assert.Contains(t, (*said)[0], "1 is no longer playable on Animefire.io")
	assert.Contains(t, (*said)[0], "StartFlix", "and where else the title may be")
	assert.Contains(t, (*said)[0], "Press ESC", "and how to leave")
	assert.NotContains(t, (*said)[0], "\n")
}

func TestReportEpisodeVideoErrorPicksTheNoticeByCause(t *testing.T) {
	anime := &models.Anime{Name: "Wind Breaker", Source: "Animefire.io"}

	t.Run("a removed video explains the removal", func(t *testing.T) {
		said := captureNotices(t)
		reportEpisodeVideoError(anime, "1", fmt.Errorf("resolving: %w", player.ErrSourceVideoRemoved))
		require.Len(t, *said, 1)
		assert.Contains(t, (*said)[0], "no longer playable")
	})
	t.Run("the user's own back says nothing", func(t *testing.T) {
		said := captureNotices(t)
		reportEpisodeVideoError(anime, "1", player.ErrBackToEpisodeSelection)
		assert.Empty(t, *said)
	})
	t.Run("any other failure says what failed", func(t *testing.T) {
		said := captureNotices(t)
		reportEpisodeVideoError(anime, "7", errors.New("the stream host timed out"))
		require.Len(t, *said, 1)
		assert.Contains(t, (*said)[0], "7 did not load")
		assert.Contains(t, (*said)[0], "the stream host timed out")
		assert.Contains(t, (*said)[0], "Pick another episode, or press ESC to go back.")
	})
}

func TestEpisodeLoadFailedNotice(t *testing.T) {
	assert.Empty(t, episodeLoadFailedNotice("1", nil), "no error, no notice")

	t.Run("names the episode and the cause", func(t *testing.T) {
		assert.Equal(t, "Episódio 06 did not load: offline. Pick another episode, or press ESC to go back.",
			episodeLoadFailedNotice("Episódio 06", errors.New("offline")))
	})
	t.Run("a blank label still reads", func(t *testing.T) {
		assert.Equal(t, "This episode did not load: offline. Pick another episode, or press ESC to go back.",
			episodeLoadFailedNotice("  ", errors.New("offline")))
	})
	t.Run("a cause that already ends in a period is not doubled", func(t *testing.T) {
		assert.Equal(t, "1 did not load: The host is down. Pick another episode, or press ESC to go back.",
			episodeLoadFailedNotice("1", errors.New("The host is down.")))
	})
	t.Run("a multi-line cause keeps every line (the list flattens them)", func(t *testing.T) {
		got := episodeLoadFailedNotice("1", errors.New("first\nsecond"))
		assert.Contains(t, got, "first")
		assert.Contains(t, got, "second")
	})
	t.Run("an empty cause still says something", func(t *testing.T) {
		assert.Contains(t, episodeLoadFailedNotice("1", errors.New("")), "no reason was given")
	})
}

func TestReportEpisodeLoadFailedNeverPanicsOnAMissingEpisodeLabel(t *testing.T) {
	said := captureNotices(t)
	assert.NotPanics(t, func() { reportEpisodeLoadFailed("", errors.New("x")) })
	assert.Len(t, *said, 1)
}

// ---- PlayEpisode: the report reaches the list the user is sent back to -----------

// failedPlay runs PlayEpisode with a resolver that fails with err and returns
// what the user would be told on the list that opens next.
func failedPlay(t *testing.T, err error) (said []string, playErr error) {
	t.Helper()
	notices := captureNotices(t)
	prevVideo, prevData := videoURLFor, episodeDataFor
	videoURLFor = func(context.Context, *models.Episode, *models.Anime) (string, error) { return "", err }
	episodeDataFor = func(int, int, *models.Anime) error { return nil }
	t.Cleanup(func() { videoURLFor, episodeDataFor = prevVideo, prevData })

	var paused bool
	var mu sync.Mutex
	anime := &models.Anime{Name: "Wind Breaker", Source: "Animefire.io", URL: "https://animefire.one/anime/E0CBXaPxqvY"}
	eps := []models.Episode{{Number: "1", Num: 1, URL: "https://animefire.one/episode/wxpZ0IO2NVY"}}

	playErr = PlayEpisode(context.Background(), anime, eps, 1, eps[0].URL, "1", false, &paused, &mu)
	return *notices, playErr
}

func TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo(t *testing.T) {
	said, err := failedPlay(t, fmt.Errorf("%w: AnimeFire episode wxpZ0IO2NVY: media offline", player.ErrSourceVideoRemoved))

	require.ErrorIs(t, err, player.ErrBackToEpisodeSelection, "the user is sent back to the list")
	require.Len(t, said, 1, "and the list opens with the reason, which is what the log of #229 lacked")
	assert.Contains(t, said[0], "1 is no longer playable on Animefire.io")
	assert.Contains(t, said[0], "Press ESC")
}

func TestPlayEpisodeExplainsAnyOtherLoadFailureToo(t *testing.T) {
	said, err := failedPlay(t, errors.New("the stream host timed out"))

	require.ErrorIs(t, err, player.ErrBackToEpisodeSelection)
	require.Len(t, said, 1)
	assert.Contains(t, said[0], "1 did not load")
	assert.Contains(t, said[0], "the stream host timed out")
}

func TestPlayEpisodeSaysNothingWhenTheUserWentBack(t *testing.T) {
	said, err := failedPlay(t, player.ErrBackToEpisodeSelection)

	require.ErrorIs(t, err, player.ErrBackToEpisodeSelection)
	assert.Empty(t, said, "the user asked to go back; there is nothing to explain")
}
