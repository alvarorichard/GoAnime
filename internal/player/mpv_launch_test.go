package player

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #229: on a Windows PATH, WindowsApps\mpv.exe (a Microsoft Store app
// alias) comes before the mpv the GoAnime installer ships. It exited with
// status 1 straight away, and GoAnime never looked at the mpv it had installed
// itself. These tests cover the pieces of starting mpv that do not need a
// process; mpv_launch_fake_test.go starts a stand-in mpv.

func TestEssentialMPVArgsKeepsWhatPlaybackNeeds(t *testing.T) {
	args := []string{
		"--cache=yes", "--demuxer-max-bytes=300M", "--demuxer-readahead-secs=20", "--audio-display=no",
		"--no-config", "--hwdec=auto-safe", "--vo=gpu-next,gpu", "--profile=fast", "--video-latency-hacks=yes",
		"--http-header-fields=Referer: https://r.test/", "--demuxer-lavf-o=allowed_extensions=ALL",
		"--force-media-title=Show S01E01", "--sub-file=/tmp/a.srt", "--start=90", "--alang=por",
		"--glsl-shaders=/s/a.glsl", "--gpu-shader-cache-dir=/c",
	}

	kept, dropped := essentialMPVArgs(args)

	assert.Equal(t, []string{
		"--no-config", "--http-header-fields=Referer: https://r.test/", "--demuxer-lavf-o=allowed_extensions=ALL",
		"--force-media-title=Show S01E01", "--sub-file=/tmp/a.srt", "--start=90", "--alang=por",
	}, kept, "what mpv needs to play what it was asked stays, in order")
	assert.Equal(t, []string{
		"--cache=yes", "--demuxer-max-bytes=300M", "--demuxer-readahead-secs=20", "--audio-display=no",
		"--hwdec=auto-safe", "--vo=gpu-next,gpu", "--profile=fast", "--video-latency-hacks=yes",
		"--glsl-shaders=/s/a.glsl", "--gpu-shader-cache-dir=/c",
	}, dropped, "the tuning goes, in order")
}

func TestEssentialMPVArgsMatchesOptionNamesNotPrefixesOfLongerOnes(t *testing.T) {
	kept, dropped := essentialMPVArgs([]string{"--volume=50", "--vo=gpu", "--cache-secs=10", "--cache=yes", "--profile-restore=copy", "--profile=fast"})

	assert.Equal(t, []string{"--volume=50", "--cache-secs=10", "--profile-restore=copy"}, kept,
		"--volume is not --vo, and --cache-secs is not --cache")
	assert.Equal(t, []string{"--vo=gpu", "--cache=yes", "--profile=fast"}, dropped)
}

func TestEssentialMPVArgsEdges(t *testing.T) {
	kept, dropped := essentialMPVArgs(nil)
	assert.Empty(t, kept)
	assert.Empty(t, dropped)

	kept, dropped = essentialMPVArgs([]string{"--no-config", "--start=1"})
	assert.Equal(t, []string{"--no-config", "--start=1"}, kept)
	assert.Empty(t, dropped, "nothing to leave out means nothing to retry")

	kept, dropped = essentialMPVArgs([]string{"--vo=gpu", "--hwdec=no"})
	assert.Empty(t, kept)
	assert.Len(t, dropped, 2)
}

// Every tuning option has to be one GoAnime can actually pass mpv, or the list
// names something that can never be dropped.
func TestEveryTuningOptionIsOneGoAnimeCanPass(t *testing.T) {
	for _, prefix := range tuningMPVArgs {
		assert.Equal(t, []string{prefix + "x"}, filterMPVArgs([]string{prefix + "x"}),
			"%s is dropped on a retry but the argument filter would never let it through", prefix)
	}
}

func TestFailedToLaunch(t *testing.T) {
	early := &mpvEarlyExitError{code: 1, err: errors.New("exited")}
	start := &mpvStartError{err: errors.New("not runnable")}

	assert.True(t, failedToLaunch(early))
	assert.True(t, failedToLaunch(start))
	assert.True(t, failedToLaunch(fmt.Errorf("wrapped: %w", early)), "a wrapped one still counts")
	assert.False(t, failedToLaunch(errors.New("timeout waiting for mpv socket")), "an mpv that came up and was silent is not another mpv's to fix")
	assert.False(t, failedToLaunch(nil))
}

func TestMPVErrorsKeepTheirMessage(t *testing.T) {
	inner := errors.New("mpv exited before IPC socket was ready: exit status 1")
	early := &mpvEarlyExitError{code: 1, err: inner}
	start := &mpvStartError{err: inner}

	assert.Equal(t, inner.Error(), early.Error())
	assert.Equal(t, inner.Error(), start.Error())
	assert.ErrorIs(t, early, inner)
	assert.ErrorIs(t, start, inner)
}

func TestIsStoreAlias(t *testing.T) {
	for path, want := range map[string]bool{
		`C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`:  true,
		`c:\users\n\appdata\local\microsoft\windowsapps\mpv.exe`:  true,
		`C:/Users/n/AppData/Local/Microsoft/WindowsApps/mpv.exe`:  true,
		`C:\Program Files\GoAnime\bin\mpv.exe`:                    false,
		`C:\Users\n\scoop\shims\mpv.exe`:                          false,
		`/usr/bin/mpv`:                                            false,
		`C:\Users\n\WindowsApps\mpv.exe`:                          false,
		`C:\Users\n\AppData\Local\Microsoft\WindowsAppsX\mpv.exe`: false,
		``: false,
	} {
		assert.Equal(t, want, isStoreAlias(path), path)
	}
}

func TestFirstLine(t *testing.T) {
	assert.Equal(t, "one", firstLine("one\ntwo"))
	assert.Equal(t, "one", firstLine("  one  \r\ntwo"))
	assert.Equal(t, "only", firstLine("only"))
	assert.Empty(t, firstLine(""))
}

func TestOrderedMPVPutsTheOneThatWorkedFirst(t *testing.T) {
	resetMPVChoice()
	t.Cleanup(resetMPVChoice)
	paths := []string{"/a", "/b", "/c"}

	assert.Equal(t, paths, orderedMPV(paths), "nothing remembered: the order they were found in")

	rememberWorkingMPV("/c")
	assert.Equal(t, []string{"/c", "/a", "/b"}, orderedMPV(paths))

	rememberWorkingMPV("/gone")
	assert.Equal(t, paths, orderedMPV(paths), "a remembered mpv that is no longer found leads nothing")

	assert.Equal(t, paths, paths, "the caller's list is not reordered")
}

func TestMPVChoiceRemembersWhichBuildsNeedTheMinimalOptions(t *testing.T) {
	resetMPVChoice()
	t.Cleanup(resetMPVChoice)

	assert.False(t, mpvNeedsMinimal("/a"))
	rememberMinimalMPV("/a")
	assert.True(t, mpvNeedsMinimal("/a"))
	assert.False(t, mpvNeedsMinimal("/b"), "only that build")

	resetMPVChoice()
	assert.False(t, mpvNeedsMinimal("/a"))
}

func TestDescribeMPVFailures(t *testing.T) {
	mk := func(msg string) error { return &mpvEarlyExitError{code: 1, err: errors.New(msg)} }

	t.Run("one failure is returned as it is", func(t *testing.T) {
		only := mk("mpv exited\nmpv stderr: boom")
		got := describeMPVFailures([]mpvAttempt{{path: "/usr/bin/mpv", err: only}})
		assert.Same(t, only, got)
	})
	t.Run("the first mpv's error leads and the others follow", func(t *testing.T) {
		got := describeMPVFailures([]mpvAttempt{
			{path: "/a/mpv", err: mk("first failed\nwith detail")},
			{path: "/b/mpv", err: mk("second failed\nwith detail")},
			{path: "/c/mpv", err: mk("third failed")},
		})
		assert.Equal(t, "first failed\nwith detail\nAlso tried /b/mpv: second failed\nAlso tried /c/mpv: third failed", got.Error())
		var early *mpvEarlyExitError
		assert.ErrorAs(t, got, &early, "the failure is still the first mpv's kind")
	})
	t.Run("a Store alias gets the one thing that can be done", func(t *testing.T) {
		alias := `C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`
		got := describeMPVFailures([]mpvAttempt{{path: alias, err: mk("exited")}}).Error()
		assert.Contains(t, got, alias+" is a Microsoft Store app alias")
		assert.Contains(t, got, "App execution aliases")
		assert.Contains(t, got, "the GoAnime installer includes one")
	})
	t.Run("an alias that is not the first is not blamed", func(t *testing.T) {
		got := describeMPVFailures([]mpvAttempt{
			{path: `C:\Program Files\GoAnime\bin\mpv.exe`, err: mk("exited")},
			{path: `C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`, err: mk("exited")},
		}).Error()
		assert.NotContains(t, got, "Microsoft Store app alias")
	})
	t.Run("a Store alias the others also failed is still named", func(t *testing.T) {
		alias := `C:\Users\n\AppData\Local\Microsoft\WindowsApps\mpv.exe`
		got := describeMPVFailures([]mpvAttempt{{path: alias, err: mk("exited")}, {path: "/b/mpv", err: mk("also exited")}}).Error()
		assert.Contains(t, got, "Also tried /b/mpv: also exited")
		assert.Contains(t, got, "Microsoft Store app alias")
	})
}

func TestStartVideoAcrossRefusesABadLinkBeforeTryingAnyMPV(t *testing.T) {
	resetMPVChoice()
	t.Cleanup(resetMPVChoice)

	_, err := startVideoAcross([]string{"/definitely/not/mpv"}, "http://bad\nurl", nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid media target")
	assert.False(t, failedToLaunch(err), "the link is at fault, not an mpv")
}

func TestLeadWith(t *testing.T) {
	assert.Equal(t, []string{"/a"}, leadWith("/a", nil), "the cached one is all there is")
	assert.Equal(t, []string{"/a", "/b"}, leadWith("/a", []string{"/a", "/b"}), "already first: unchanged")
	assert.Equal(t, []string{"/c", "/a", "/b"}, leadWith("/c", []string{"/a", "/b"}), "not in the list: put in front")
	assert.Equal(t, []string{"/b", "/a", "/c"}, leadWith("/b", []string{"/a", "/b", "/c"}), "found later: moved to the front, not repeated")
	paths := []string{"/a", "/b"}
	_ = leadWith("/b", paths)
	assert.Equal(t, []string{"/a", "/b"}, paths, "the caller's list is not touched")
}
