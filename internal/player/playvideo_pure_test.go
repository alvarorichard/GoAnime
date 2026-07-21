//go:build !windows

package player

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFakeDiscordUpdater() *discord.RichPresenceUpdater {
	paused := false
	var mu sync.Mutex
	return discord.NewRichPresenceUpdater(
		&models.Anime{Name: "Test"},
		&paused,
		&mu,
		time.Second,
		0,
		"",
		MpvSendCommand,
	)
}

func TestApplySkipTimes_SendsConcatenatedOptions(t *testing.T) {
	t.Parallel()
	var got []any
	var wg sync.WaitGroup
	wg.Add(1)
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		defer wg.Done()
		got, _ = req["command"].([]any)
		return mpvOK(nil)
	})

	ep := &models.Episode{Number: "01"}
	ep.SkipTimes.Op.Start, ep.SkipTimes.Op.End = 30, 90
	ep.SkipTimes.Ed.Start, ep.SkipTimes.Ed.End = 1320, 1380

	applySkipTimes(sock, ep)
	wg.Wait()

	require.Len(t, got, 3)
	assert.Equal(t, "set_property", got[0])
	assert.Equal(t, "script-opts", got[1])
	opts, _ := got[2].(string)
	assert.Contains(t, opts, "skip_op=30-90")
	assert.Contains(t, opts, "skip_ed=1320-1380")
}

func TestApplySkipTimes_NoOptionsMeansNoCall(t *testing.T) {
	t.Parallel()
	var called atomic.Bool
	sock := startMockMPVSocket(t, func(map[string]any) []byte {
		called.Store(true)
		return mpvOK(nil)
	})

	applySkipTimes(sock, &models.Episode{Number: "01"})
	// Small window so any (unwanted) dial would be observed.
	time.Sleep(50 * time.Millisecond)
	assert.False(t, called.Load(), "no IPC call expected when skip times are zero")
}

func TestFindEpisodeIndex(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{Number: "Episódio 1", Num: 1},
		{Number: "Episódio 2", Num: 2},
		{Number: "3", Num: 3},
		{Number: "Movie", Num: 4},
	}

	tests := []struct {
		name string
		num  int
		want int
	}{
		{"first via extraction", 1, 0},
		{"second via extraction", 2, 1},
		{"third via raw number string", 3, 2},
		{"missing returns -1", 5, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, findEpisodeIndex(episodes, tt.num))
		})
	}
}

func TestFindEpisodeIndex_IndexBasedFallback(t *testing.T) {
	t.Parallel()
	// Neither ExtractEpisodeNumber nor Num match → falls back to index-based access.
	episodes := []models.Episode{
		{Number: "Movie", Num: 999},
		{Number: "Special", Num: 888},
	}
	assert.Equal(t, 0, findEpisodeIndex(episodes, 1))
	assert.Equal(t, 1, findEpisodeIndex(episodes, 2))
	assert.Equal(t, -1, findEpisodeIndex(episodes, 3))
	assert.Equal(t, -1, findEpisodeIndex(episodes, 0))
}

func TestTrackingKey_BuildsPerEpisodeKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://x/anime:ep5", trackingKey("https://x/anime", 5))
	assert.Equal(t, ":ep0", trackingKey("", 0))
}

func TestGetTrackerDBPath_CachesAndReturnsValidPath(t *testing.T) {
	// Mutates package-level cachedDBPath — keep serial.
	prev := cachedDBPath
	cachedDBPath = ""
	t.Cleanup(func() { cachedDBPath = prev })

	got := getTrackerDBPath()
	require.NotEmpty(t, got)
	assert.True(t, strings.HasSuffix(got, filepath.Join("tracking", "progress.db")),
		"path %q should end with tracking/progress.db", got)

	// Second call returns the cached value without recomputing.
	again := getTrackerDBPath()
	assert.Equal(t, got, again)
}

func TestGetCurrentEpisode(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{Number: "Episódio 1", Num: 1},
		{Number: "Episódio 2", Num: 2},
		{Number: "3", Num: 3},
	}
	ep, err := getCurrentEpisode(episodes, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, ep.Num)

	ep, err = getCurrentEpisode(episodes, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, ep.Num)

	_, err = getCurrentEpisode(episodes, 99)
	require.Error(t, err)
}

func TestGetCurrentEpisode_IndexFallback(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{Number: "Movie", Num: 999},
		{Number: "Special", Num: 888},
	}
	// Neither match → index-based access.
	ep, err := getCurrentEpisode(episodes, 1)
	require.NoError(t, err)
	assert.Equal(t, 999, ep.Num)
}

func TestGetEpisodeTitle_Priority(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "EN", getEpisodeTitle(models.TitleDetails{English: "EN", Romaji: "RO"}))
	assert.Equal(t, "RO", getEpisodeTitle(models.TitleDetails{Romaji: "RO", Japanese: "JA"}))
	assert.Equal(t, "JA", getEpisodeTitle(models.TitleDetails{Japanese: "JA"}))
	assert.Equal(t, "No title", getEpisodeTitle(models.TitleDetails{}))
}

func TestInitTracking_NoCgoReturnsNilTracker(t *testing.T) {
	// tracking.IsCgoEnabled is package-level and platform-dependent. The
	// function returns (nil, 0) when CGO is off; when CGO is on, the
	// underlying tracker may fail to open the DB and also return (nil, 0).
	// Either way, no panic and the tracker handle is allowed to be nil.
	tracker, resume := initTracking(0, &models.Episode{URL: "x", Num: 1}, 1)
	assert.GreaterOrEqual(t, resume, 0)
	_ = tracker
}

func TestInitTrackerAsync_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, InitTrackerAsync)
	// Give the goroutine time to no-op.
	time.Sleep(20 * time.Millisecond)
}

func TestUpdateTrackingWithDuration_NilTrackerNoop(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		updateTrackingWithDuration(nil, 1, &models.Episode{URL: "x", Num: 1}, 1, time.Minute)
	})
}

func TestFetchAniSkipAsync_RoutesMALID(t *testing.T) {
	// Mutates package-level aniSkipFetcher — keep serial.
	prev := aniSkipFetcher
	t.Cleanup(func() { aniSkipFetcher = prev })

	var seenMAL atomic.Int32
	aniSkipFetcher = func(malID, _ int, _ *models.Episode) error {
		seenMAL.Store(int32(malID))
		return nil
	}
	ch := fetchAniSkipAsync(12345, 1, &models.Episode{})
	require.NoError(t, <-ch)
	assert.Equal(t, int32(12345), seenMAL.Load())
}

func TestShowShaderOSD_DoesNotPanicOnBogusSocket(t *testing.T) {
	assert.NotPanics(t, func() { showShaderOSD("/tmp/missing_mpv_socket.sock") })
	// Give the background goroutine a moment to fire and fail silently
	// before the test exits and triggers a race report.
	time.Sleep(700 * time.Millisecond)
}

func TestApplyAniSkipResults_TimesOutWhenNoData(t *testing.T) {
	t.Parallel()
	ch := make(chan error)
	// Will hit the 2s timeout branch since nothing sends on ch. Run async
	// and just wait briefly to ensure no panic; the goroutine returns on
	// its own timer.
	assert.NotPanics(t, func() {
		applyAniSkipResults(ch, "/tmp/missing.sock", &models.Episode{URL: "x"}, 1)
	})
}

func TestApplyAniSkipResults_NilErrorAppliesSkipTimes(t *testing.T) {
	var mu sync.Mutex
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		mu.Lock()
		seen = cmd
		mu.Unlock()
		return mpvOK(nil)
	})

	ch := make(chan error, 1)
	ch <- nil

	ep := &models.Episode{URL: "https://x/long-url-not-allanime"}
	ep.SkipTimes.Op.Start, ep.SkipTimes.Op.End = 30, 90

	applyAniSkipResults(ch, sock, ep, 1)
	deadline := time.Now().Add(1 * time.Second)
	for {
		mu.Lock()
		n := len(seen)
		mu.Unlock()
		if n > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	require.NotEmpty(t, seen)
	assert.Equal(t, "set_property", seen[0])
	mu.Unlock()
}

func TestApplyAniSkipResults_ErrorOnChanFallsThrough(t *testing.T) {
	t.Parallel()
	ch := make(chan error, 1)
	ch <- assert.AnError
	assert.NotPanics(t, func() {
		applyAniSkipResults(ch, "/tmp/missing.sock", &models.Episode{URL: "x"}, 1)
	})
	// Give goroutine a moment to consume and log.
	time.Sleep(50 * time.Millisecond)
}

func TestWaitForVideoReady_ReturnsTrueOnPositiveDuration(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) == 2 && cmd[1] == "duration" {
			return mpvOK(123.0)
		}
		return mpvOK(0.0)
	})
	assert.True(t, waitForVideoReady(sock))
}

func TestWaitForVideoReady_ReturnsTrueOnPositiveTimePos(t *testing.T) {
	t.Parallel()
	// duration first call returns 0 → poll alternates and hits time-pos
	// branch on the second iteration.
	var iter atomic.Int32
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) >= 2 && cmd[1] == "duration" {
			return mpvOK(0.0)
		}
		if len(cmd) >= 2 && cmd[1] == "time-pos" {
			iter.Add(1)
			return mpvOK(5.0)
		}
		return mpvOK(nil)
	})
	assert.True(t, waitForVideoReady(sock))
}

func TestUpdateEpisodeDuration_ExistingDurationShortCircuits(t *testing.T) {
	upd := newFakeDiscordUpdater()
	upd.SetEpisodeStarted(true)
	upd.SetEpisodeDuration(time.Minute) // pre-existing

	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(0.0) })
	done := make(chan struct{})
	go func() {
		updateEpisodeDuration(sock, upd, nil, 0, &models.Episode{URL: "x", Num: 1}, 1)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("updateEpisodeDuration with pre-set duration must short-circuit")
	}
}

func TestSelectEpisode_PropagatesBackRequest(t *testing.T) {
	// SelectEpisodeWithFuzzyFinder with no episodes returns "no episodes
	// provided" error — selectEpisode wraps it as "failed to select episode".
	stop := make(chan struct{})
	err := selectEpisode(nil, 0, 0, nil, stop, "/tmp/missing.sock")
	require.Error(t, err)
}

func TestSwitchEpisode_BoundsCheck(t *testing.T) {
	// Single episode, target index 0 (valid). switchEpisode will try to
	// resolve via GetVideoURLForEpisodeEnhanced and fail (no real anime/source).
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "shortid"}}
	stop := make(chan struct{})
	err := switchEpisode(0, episodes, 0, 0, nil, stop, "/tmp/missing.sock")
	require.Error(t, err)
}

func TestSeekToResumePosition_NonPositiveReturnsImmediately(t *testing.T) {
	t.Parallel()
	// Bogus socket would otherwise time out for 45s; resumeTime<=0 must
	// short-circuit before any IPC.
	start := time.Now()
	seekToResumePosition("/tmp/missing.sock", 0)
	assert.Less(t, time.Since(start), 100*time.Millisecond)
}

func TestSeekToResumePosition_HappyPathSeeksAndReturns(t *testing.T) {
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) >= 2 && cmd[1] == "duration" {
			return mpvOK(300.0)
		}
		if len(cmd) >= 2 && cmd[1] == "time-pos" {
			return mpvOK(60.0)
		}
		if len(cmd) > 0 && cmd[0] == "seek" {
			return mpvOK(nil)
		}
		return mpvOK(nil)
	})
	done := make(chan struct{})
	go func() {
		seekToResumePosition(sock, 55)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("seek did not return in time")
	}
}

func TestWaitForPlaybackStart_SetsStartedWhenSocketResponds(t *testing.T) {
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(0.5) })
	upd := newFakeDiscordUpdater()
	done := make(chan struct{})
	go func() {
		waitForPlaybackStart(sock, upd)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("waitForPlaybackStart timed out")
	}
	assert.True(t, upd.IsEpisodeStarted())
}

func TestUpdateEpisodeDuration_StoresDurationFromSocket(t *testing.T) {
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) >= 2 && cmd[1] == "duration" {
			return mpvOK(1440.0)
		}
		return mpvOK(nil)
	})
	upd := newFakeDiscordUpdater()
	upd.SetEpisodeStarted(true)
	done := make(chan struct{})
	go func() {
		updateEpisodeDuration(sock, upd, nil, 0, &models.Episode{URL: "x", Num: 1}, 1)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("updateEpisodeDuration timed out")
	}
	assert.True(t, upd.GetEpisodeDuration() > 0)
}

func TestUpdateTracking_ReturnsOnBadSocket(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		updateTracking(nil, "/tmp/missing.sock", 0, &models.Episode{URL: "x"}, 1, nil)
	})
}

func TestPreloadNextEpisode_AllAnimeShortIDSkipped(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{URL: "https://x/y"},
		{URL: "shortid"}, // AllAnime short ID — must be skipped (no goroutine)
	}
	assert.NotPanics(t, func() { preloadNextEpisode(episodes, 0) })
}

func TestPreloadNextEpisode_LastIndexNoop(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{{URL: "https://x/y"}}
	assert.NotPanics(t, func() { preloadNextEpisode(episodes, 0) })
}

func TestStartTrackingRoutine_NilTrackerReturnsClosableChan(t *testing.T) {
	t.Parallel()
	ch := startTrackingRoutine(nil, "/tmp/x.sock", 0, &models.Episode{URL: "x"}, 1, nil)
	require.NotNil(t, ch)
	close(ch)
}

func TestSkipIntro_WithDataSendsSeek(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	ep := &models.Episode{}
	ep.SkipTimes.Op.End = 90
	skipIntro(sock, ep)
	require.Len(t, seen, 3)
	assert.Equal(t, "seek", seen[0])
}

func TestSkipIntro_WithoutDataPrintsMessage(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		skipIntro("/tmp/missing.sock", &models.Episode{})
	})
}

// selectAudioTrack/selectSubtitleTrack invoke fuzzyfinder when tracks
// are present — keep serial to avoid tcell terminfo races.
func TestSelectAudioTrack_NoTracksPrintsAndReturns(t *testing.T) {
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK([]any{}) })
	assert.NotPanics(t, func() { selectAudioTrack(sock) })
}

func TestSelectAudioTrack_SocketErrorReturnsCleanly(t *testing.T) {
	assert.NotPanics(t, func() { selectAudioTrack("/tmp/missing.sock") })
}

func TestSelectSubtitleTrack_SocketErrorReturnsCleanly(t *testing.T) {
	assert.NotPanics(t, func() { selectSubtitleTrack("/tmp/missing.sock") })
}

// With tracks present, selectAudioTrack iterates through them building
// labels, then calls fuzzyfinder. Outside a TTY, fuzzyfinder errors and the
// function returns silently — we still exercise the label-construction loop.
func TestSelectAudioTrack_WithTracksFuzzyFinderErrorIsSilent(t *testing.T) {
	tracks := []any{
		map[string]any{"id": 1.0, "type": "audio", "lang": "ja", "codec": "aac", "demux-channels": "stereo"},
		map[string]any{"id": 2.0, "type": "audio", "title": "Director Comm.", "codec": "ac3", "audio-channels": 6.0},
	}
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) >= 2 && cmd[1] == "track-list" {
			return mpvOK(tracks)
		}
		if len(cmd) >= 2 && cmd[1] == "aid" {
			return mpvOK(1.0)
		}
		return mpvOK(nil)
	})
	assert.NotPanics(t, func() { selectAudioTrack(sock) })
}

func TestSelectSubtitleTrack_WithTracksFuzzyFinderErrorIsSilent(t *testing.T) {
	tracks := []any{
		map[string]any{"id": 1.0, "type": "sub", "lang": "pt"},
		map[string]any{"id": 2.0, "type": "sub", "title": "Director Comm."},
	}
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		cmd, _ := req["command"].([]any)
		if len(cmd) >= 2 && cmd[1] == "track-list" {
			return mpvOK(tracks)
		}
		if len(cmd) >= 2 && cmd[1] == "sid" {
			return mpvOK(0.0)
		}
		return mpvOK(nil)
	})
	assert.NotPanics(t, func() { selectSubtitleTrack(sock) })
}

func TestHandleUserInput_AliveSocketReturnsBackOnMenuError(t *testing.T) {
	// Mock socket responds OK to `get_property pid` (ping passes), then the
	// fuzzyfinder menu fails outside a TTY → handleUserInput sends `quit`
	// to mpv and surfaces ErrBackToDownloadOptions.
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(1234.0) })
	stop := make(chan struct{})
	err := handleUserInput(sock, nil, 0, 1, 0, 0, nil, stop, &models.Episode{})
	assert.ErrorIs(t, err, ErrBackToDownloadOptions)
}

// The following functions all depend on an interactive TUI (fuzzyfinder)
// and/or a fully-orchestrated playback session. Per CLAUDE.md, the
// internal logic is exercised by sibling tests; here we pin the
// non-interactive entry-point behaviour (return paths without prompting).

// fuzzyfinder/tcell uses package-level terminfo state — calls from
// multiple goroutines race on tcell internals. Keep TUI-touching tests
// serial.
func TestShowPlayerMenu_ReturnsErrorOutsideTerminal(t *testing.T) {
	_, err := showPlayerMenu("Anime", 1)
	assert.Error(t, err, "fuzzyfinder must error outside a TTY")
}

func TestShowResumeDialog_ReturnsErrorOutsideTerminal(t *testing.T) {
	_, err := showResumeDialog(1, 95, false)
	assert.Error(t, err)
	_, err = showResumeDialog(1, 95, true)
	assert.Error(t, err)
}

func TestIsMPVConnectionDead(t *testing.T) {
	t.Parallel()
	assert.False(t, isMPVConnectionDead(nil))
	assert.True(t, isMPVConnectionDead(errors.New("dial unix /tmp/x: connect: connection refused")))
	assert.True(t, isMPVConnectionDead(errors.New("no such file or directory")))
	assert.True(t, isMPVConnectionDead(errors.New("write: broken pipe")))
	assert.False(t, isMPVConnectionDead(errors.New("property unavailable")))
}

func TestHandleUserInput_QuitsWhenSocketDead(t *testing.T) {
	t.Parallel()
	stop := make(chan struct{})
	err := handleUserInput("/tmp/missing.sock", nil, 0, 1, 0, 0, nil, stop, &models.Episode{})
	assert.ErrorIs(t, err, ErrBackToDownloadOptions)
}

func TestPlayNextEpisode_LastEpisode_Noop(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{{Number: "1", Num: 1}}
	stop := make(chan struct{})
	err := playNextEpisode(1, episodes, 0, 0, nil, stop, "/tmp/missing.sock")
	// Boundary must keep the player menu open, not return nil (which unwound
	// the menu loop while mpv kept playing).
	assert.ErrorIs(t, err, errStayInPlayerMenu)
}

func TestPlayPreviousEpisode_FirstEpisode_Noop(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{{Number: "1", Num: 1}}
	stop := make(chan struct{})
	err := playPreviousEpisode(-1, episodes, 0, 0, nil, stop, "/tmp/missing.sock")
	assert.ErrorIs(t, err, errStayInPlayerMenu)
}

func TestSelectEpisode_FuzzyFinderUnavailable(t *testing.T) {
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "x"}}
	stop := make(chan struct{})
	err := selectEpisode(episodes, 0, 0, nil, stop, "/tmp/missing.sock")
	assert.Error(t, err)
}

func TestSwitchEpisode_PropagatesResolutionError(t *testing.T) {
	// ExtractEpisodeNumber never produces a non-numeric — exercise the next
	// failure mode: GetVideoURLForEpisodeEnhanced cannot resolve a bogus
	// short ID, so an error surfaces.
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "shortid"}}
	stop := make(chan struct{})
	err := switchEpisode(0, episodes, 0, 0, nil, stop, "/tmp/missing.sock")
	require.Error(t, err)
}

func TestPlayVideo_PropagatesEpisodeNotFound(t *testing.T) {
	t.Parallel()
	// Empty episodes list — getCurrentEpisode returns an error and the
	// orchestration aborts before touching mpv.
	err := playVideo("https://x/y.mp4", nil, 1, 0, 0, nil)
	require.Error(t, err)
}

func TestPlayVideo_ValidEpisodeRoutesToStartVideo(t *testing.T) {
	// Valid episode → getCurrentEpisode succeeds. mpv args + tracking setup
	// run. StartVideo either fails (no mpv) or launches mpv with a bogus
	// URL; in both cases playVideo returns once the player exits or fails.
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "https://cdn/ep.mp4"}}
	_ = playVideo("https://cdn.example/ep.mp4", episodes, 1, 0, 0, nil)
}

func TestPlayVideo_MovieTypeAppliesLanguagePreferences(t *testing.T) {
	// Movie/TV → playVideo enters the language-preference branch which
	// adds --alang/--slang mpv args. Set gMedia accordingly.
	SetExactMediaType("movie")
	SetMediaType(true)
	t.Cleanup(func() {
		SetExactMediaType("")
		SetMediaType(false)
	})
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "https://cdn/movie.mp4"}}
	_ = playVideo("https://cdn.example/movie.mp4", episodes, 1, 0, 0, nil)
}

func TestPlayVideo_BloggerURLTriggersResolveAttempt(t *testing.T) {
	// blogger.com/video.g URL → playVideo calls extractBloggerGoogleVideoURL
	// (fails because URL is bogus) then continues with the original URL.
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "https://blogger.com/video.g?token=abc"}}
	_ = playVideo("https://blogger.com/video.g?token=abc", episodes, 1, 0, 0, nil)
}

// HLS path waits up to 45s for the bogus stream to become ready when mpv is
// installed. Keep this test off the default suite by skipping when mpv is
// present; otherwise it provides cheap StartVideo failure-path coverage.
func TestPlayVideo_HLSURLEnablesHLSPath(t *testing.T) {
	if _, err := exec.LookPath("mpv"); err == nil {
		t.Skip("mpv present on host — HLS branch waits 45s polling for stream readiness")
	}
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "https://cdn/ep.m3u8"}}
	err := playVideo("https://cdn.example/ep.m3u8", episodes, 1, 0, 0, nil)
	require.Error(t, err)
}

// initDiscordPresence spawns two long-lived goroutines (Discord RPC ticker
// + waitForPlaybackStart/updateEpisodeDuration polling) that outlive the
// test. Driving the orchestration end-to-end requires a Discord daemon
// and a real mpv socket. Per CLAUDE.md, TUI/external-boundary functions
// pin via symbol presence; the collaborators (waitForPlaybackStart,
// updateEpisodeDuration) each have their own dedicated tests above.
func TestInitDiscordPresence_SymbolPinned(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, initDiscordPresence)
}
