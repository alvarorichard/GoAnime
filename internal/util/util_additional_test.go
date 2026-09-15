package util

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// snapshotCleanup saves and restores the global cleanup list. Tests touching
// cleanup state must run serially because cleanupFuncs is a package singleton.
func snapshotCleanup(t *testing.T) {
	t.Helper()
	cleanupMu.Lock()
	prev := append([]func(){}, cleanupFuncs...)
	cleanupMu.Unlock()
	t.Cleanup(func() {
		cleanupMu.Lock()
		cleanupFuncs = prev
		cleanupMu.Unlock()
	})
}

func snapshotGlobalRequest(t *testing.T) {
	t.Helper()
	prevDl := GlobalDownloadRequest
	prevUp := GlobalUpscaleRequest
	prevOut := GlobalOutputDir
	prevSrc := GlobalSource
	prevQual := GlobalQuality
	prevMT := GlobalMediaType
	prevSub := GlobalSubsLanguage
	prevAud := GlobalAudioLanguage
	prevNoSubs := GlobalNoSubs
	prevDebug := IsDebug
	prevPerf := PerfEnabled
	prevLogPath := LogFilePath
	t.Cleanup(func() {
		GlobalDownloadRequest = prevDl
		GlobalUpscaleRequest = prevUp
		GlobalOutputDir = prevOut
		GlobalSource = prevSrc
		GlobalQuality = prevQual
		GlobalMediaType = prevMT
		GlobalSubsLanguage = prevSub
		GlobalAudioLanguage = prevAud
		GlobalNoSubs = prevNoSubs
		IsDebug = prevDebug
		PerfEnabled = prevPerf
		LogFilePath = prevLogPath
	})
}

func TestRegisterCleanup_AppendsFunction(t *testing.T) {
	snapshotCleanup(t)
	cleanupMu.Lock()
	cleanupFuncs = nil
	cleanupMu.Unlock()

	var called atomic.Int32
	RegisterCleanup(func() { called.Add(1) })

	cleanupMu.Lock()
	got := len(cleanupFuncs)
	cleanupMu.Unlock()
	assert.Equal(t, 1, got, "RegisterCleanup must append exactly one function")
}

func TestRegisterCleanup_ConcurrentSafe(t *testing.T) {
	snapshotCleanup(t)
	cleanupMu.Lock()
	cleanupFuncs = nil
	cleanupMu.Unlock()

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			RegisterCleanup(func() {})
		})
	}
	wg.Wait()

	cleanupMu.Lock()
	got := len(cleanupFuncs)
	cleanupMu.Unlock()
	assert.Equal(t, 50, got)
}

func TestRunCleanup_InvokesAllRegisteredAndDisablesPerfPrint(t *testing.T) {
	snapshotCleanup(t)
	prevPerf := PerfEnabled
	t.Cleanup(func() { PerfEnabled = prevPerf })

	cleanupMu.Lock()
	cleanupFuncs = nil
	cleanupMu.Unlock()

	var n atomic.Int32
	RegisterCleanup(func() { n.Add(1) })
	RegisterCleanup(func() { n.Add(1) })
	RegisterCleanup(func() { n.Add(1) })

	PerfEnabled = false // avoid printing the perf report

	RunCleanup()
	assert.Equal(t, int32(3), n.Load())
}

func TestRunCleanup_EmptyListIsNoOp(t *testing.T) {
	snapshotCleanup(t)
	cleanupMu.Lock()
	cleanupFuncs = nil
	cleanupMu.Unlock()
	RunCleanup() // must not panic
}

func TestErrorHandler_DebugWithLogPath(t *testing.T) {
	snapshotGlobalRequest(t)
	IsDebug = true
	LogFilePath = "/tmp/debug.log"
	out := ErrorHandler(errors.New("boom"))
	assert.Contains(t, out, "boom")
	assert.Contains(t, out, "/tmp/debug.log")
}

func TestErrorHandler_DebugWithoutLogPath(t *testing.T) {
	snapshotGlobalRequest(t)
	IsDebug = true
	LogFilePath = ""
	out := ErrorHandler(errors.New("kaboom"))
	assert.Contains(t, out, "kaboom")
	assert.NotContains(t, out, "Debug log saved to")
}

func TestErrorHandler_NonDebugIncludesDebugHint(t *testing.T) {
	snapshotGlobalRequest(t)
	IsDebug = false
	out := ErrorHandler(errors.New("oops"))
	assert.Contains(t, out, "oops")
	assert.Contains(t, out, "--debug")
}

// TestHelper_WritesToStdout — Helper writes a styled help block to stdout.
// The pipe is drained concurrently to prevent the pipe-buffer deadlock on
// Windows: ShowBeautifulHelp emits ~12 KB in one shot, which exceeds the
// default 4 KB kernel pipe buffer and causes WriteFile to block forever if
// the read end is only opened after the write returns.
func TestHelper_WritesToStdout(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	ch := make(chan int64, 1)
	go func() {
		n, _ := io.Copy(io.Discard, r)
		ch <- n
	}()

	Helper()

	os.Stdout = orig
	require.NoError(t, w.Close())

	n := <-ch
	r.Close()
	assert.Greater(t, n, int64(0), "Helper must produce output")
}

// snapshotOsArgs captures and restores os.Args around a test.
func snapshotOsArgs(t *testing.T) {
	t.Helper()
	prev := os.Args
	t.Cleanup(func() { os.Args = prev })
}

func TestFlagParser_PlainName(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "naruto"}

	name, err := FlagParser()
	require.NoError(t, err)
	assert.Equal(t, "naruto", name)
}

func TestFlagParser_TooShortNameFails(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "abc"}

	_, err := FlagParser()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestFlagParser_HelpFlagReturnsHelpRequested(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "--help"}

	// Helper() → ShowBeautifulHelp writes ~12 KB to stdout in one shot.
	// Without a concurrent drainer the pipe buffer fills and WriteFile
	// blocks forever on Windows (same deadlock as TestHelper_WritesToStdout).
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	go func() { _, _ = io.Copy(io.Discard, r) }()
	t.Cleanup(func() { os.Stdout = orig; _ = r.Close() })

	_, err = FlagParser()
	_ = w.Close()
	require.ErrorIs(t, err, ErrHelpRequested)
}

func TestFlagParser_UpdateFlagReturnsUpdateRequested(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "--update"}

	_, err := FlagParser()
	require.ErrorIs(t, err, ErrUpdateRequested)
}

func TestFlagParser_DownloadModeStoresRequest(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "-d", "naruto", "5"}

	name, err := FlagParser()
	require.ErrorIs(t, err, ErrDownloadRequested)
	assert.Equal(t, "naruto", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.Equal(t, 5, GlobalDownloadRequest.EpisodeNum)
	assert.False(t, GlobalDownloadRequest.IsRange)
}

func TestFlagParser_DownloadModeRangeStoresRequest(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "-d", "-r", "demon slayer", "1-3"}

	name, err := FlagParser()
	require.ErrorIs(t, err, ErrDownloadRequested)
	assert.Equal(t, "demon-slayer", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsRange)
	assert.Equal(t, 1, GlobalDownloadRequest.StartEpisode)
	assert.Equal(t, 3, GlobalDownloadRequest.EndEpisode)
}

func TestFlagParser_DownloadAllSetsAllFlag(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "-d", "-a", "one piece"}

	_, err := FlagParser()
	require.ErrorIs(t, err, ErrDownloadRequested)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsAll)
}

func TestFlagParser_MovieDownloadStoresRequest(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "-dm", "inception"}

	_, err := FlagParser()
	require.ErrorIs(t, err, ErrMovieDownloadRequested)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsMovie)
}

func TestFlagParser_UpscaleStoresRequest(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)

	tmp := filepath.Join(t.TempDir(), "video.mp4")
	require.NoError(t, os.WriteFile(tmp, []byte("x"), 0o600))
	os.Args = []string{"goanime", "--upscale", tmp}

	path, err := FlagParser()
	require.ErrorIs(t, err, ErrUpscaleRequested)
	assert.Equal(t, tmp, path)
	require.NotNil(t, GlobalUpscaleRequest)
	assert.Equal(t, tmp, GlobalUpscaleRequest.InputPath)
}

func TestFlagParser_PerfFlagEnablesPerfAndDebug(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	os.Args = []string{"goanime", "-perf", "naruto"}

	_, err := FlagParser()
	require.NoError(t, err)
	assert.True(t, PerfEnabled)
	assert.True(t, IsDebug)
}

// getUserInput drives a huh form that requires a TTY. Without one,
// tui.RunClean returns an error immediately — we use that to verify the
// function returns (empty name, non-nil error) instead of hanging.
// This is the only practical way to exercise the function without a real TTY.
func TestGetUserInput_NoTTYReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows, bubbletea uses ReadConsole (the native Windows console
		// API) rather than reading from os.Stdin. Redirecting os.Stdin to
		// os.DevNull has no effect — the form still blocks on the real console
		// handle indefinitely, making the test hang instead of error.
		t.Skip("Windows console API bypasses os.Stdin redirection; no-TTY error path not testable without process-level console detachment")
	}
	if os.Getenv("CI") == "" && os.Getenv("GOANIME_RUN_TTY_TESTS") == "1" {
		t.Skip("explicit TTY mode requested by env — skipping non-TTY assertion")
	}
	// Redirect stdin to /dev/null to guarantee no TTY.
	devnull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = devnull.Close() })
	prev := os.Stdin
	os.Stdin = devnull
	t.Cleanup(func() { os.Stdin = prev })

	name, err := getUserInput("Test prompt")
	// Without a TTY huh.Run typically errors → empty name + non-nil err.
	assert.Empty(t, name)
	assert.Error(t, err)
}

func TestTreatingAnimeName_LowersAndDashes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"Naruto", "naruto"},
		{"Attack On Titan", "attack-on-titan"},
		{"  spaces  ", "--spaces--"},
		{"", ""},
		{"ALL CAPS NAME", "all-caps-name"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, TreatingAnimeName(tt.in))
		})
	}
}

func TestHandleDownloadModeWithSmart_EmptyArgsErrors(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleDownloadModeWithSmart(nil, false, false, "", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires anime name")
}

func TestHandleDownloadModeWithSmart_AllSetsRequest(t *testing.T) {
	snapshotGlobalRequest(t)
	name, err := handleDownloadModeWithSmart([]string{"naruto"}, false, true, "src", "1080p")
	require.ErrorIs(t, err, ErrDownloadRequested)
	assert.Equal(t, "naruto", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsAll)
	assert.Equal(t, "src", GlobalDownloadRequest.Source)
	assert.Equal(t, "1080p", GlobalDownloadRequest.Quality)
}

func TestHandleDownloadModeWithSmart_RangeValid(t *testing.T) {
	snapshotGlobalRequest(t)
	name, err := handleDownloadModeWithSmart([]string{"naruto", "1-3"}, true, false, "", "best")
	require.ErrorIs(t, err, ErrDownloadRequested)
	assert.Equal(t, "naruto", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsRange)
	assert.Equal(t, 1, GlobalDownloadRequest.StartEpisode)
	assert.Equal(t, 3, GlobalDownloadRequest.EndEpisode)
}

func TestHandleDownloadModeWithSmart_RangeMissingRangeArg(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleDownloadModeWithSmart([]string{"naruto"}, true, false, "", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "range")
}

func TestHandleDownloadModeWithSmart_RangeInvalidFormat(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleDownloadModeWithSmart([]string{"naruto", "abc"}, true, false, "", "best")
	require.Error(t, err)
}

func TestHandleDownloadModeWithSmart_RangeStartAfterEnd(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleDownloadModeWithSmart([]string{"naruto", "5-2"}, true, false, "", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be greater")
}

func TestHandleDownloadModeWithSmart_RangeNonPositive(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleDownloadModeWithSmart([]string{"naruto", "0-3"}, true, false, "", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestHandleDownloadModeWithSmart_SingleNumericArg(t *testing.T) {
	snapshotGlobalRequest(t)
	name, err := handleDownloadModeWithSmart([]string{"naruto", "5"}, false, false, "", "best")
	require.ErrorIs(t, err, ErrDownloadRequested)
	assert.Equal(t, "naruto", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.Equal(t, 5, GlobalDownloadRequest.EpisodeNum)
	assert.False(t, GlobalDownloadRequest.IsRange)
}

// handleDownloadModeWithSmart with non-numeric trailing arg falls through to
// the interactive menu, which needs a TTY. Without one it returns an error.
func TestHandleDownloadModeWithSmart_NoEpisodeNumberFallsThroughToMenu(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows bubbletea uses ReadConsole (raw console handle) instead of
		// os.Stdin, so redirecting stdin to os.DevNull has no effect and the
		// form blocks on the real console indefinitely.
		t.Skip("Windows console API bypasses os.Stdin redirection; no-TTY error path not testable without process-level console detachment")
	}
	snapshotGlobalRequest(t)
	devnull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = devnull.Close() })
	prev := os.Stdin
	os.Stdin = devnull
	t.Cleanup(func() { os.Stdin = prev })

	_, err = handleDownloadModeWithSmart([]string{"naruto"}, false, false, "", "best")
	assert.Error(t, err)
}

func TestHandleUpscaleMode_EmptyArgsErrors(t *testing.T) {
	snapshotGlobalRequest(t)
	// Build minimal FlagSet with no positional args.
	fs := newEmptyFlagSet(t)
	_, err := handleUpscaleMode(fs, "", 2, 2, false, false, false, "8M", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input file path")
}

func TestHandleUpscaleMode_NonexistentInputErrors(t *testing.T) {
	snapshotGlobalRequest(t)
	fs := newFlagSetWithArgs(t, "/no/such/file.mp4")
	_, err := handleUpscaleMode(fs, "", 2, 2, false, false, false, "8M", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHandleUpscaleMode_ValidStoresRequestWithDefaults(t *testing.T) {
	snapshotGlobalRequest(t)
	tmp := filepath.Join(t.TempDir(), "v.mp4")
	require.NoError(t, os.WriteFile(tmp, []byte("x"), 0o600))
	fs := newFlagSetWithArgs(t, tmp)

	// scale 9 is out of range → defaults to 2; passes 0 → defaults to 2.
	path, err := handleUpscaleMode(fs, "out.mp4", 9, 0, true, false, true, "12M", 4)
	require.ErrorIs(t, err, ErrUpscaleRequested)
	assert.Equal(t, tmp, path)
	require.NotNil(t, GlobalUpscaleRequest)
	assert.Equal(t, 2, GlobalUpscaleRequest.ScaleFactor)
	assert.Equal(t, 2, GlobalUpscaleRequest.Passes)
	assert.True(t, GlobalUpscaleRequest.FastMode)
	assert.True(t, GlobalUpscaleRequest.UseGPU)
	assert.Equal(t, "12M", GlobalUpscaleRequest.VideoBitrate)
	assert.Equal(t, 4, GlobalUpscaleRequest.Workers)
}

func TestHandleUpscaleMode_HQModeOverridesPassesAndStrength(t *testing.T) {
	snapshotGlobalRequest(t)
	tmp := filepath.Join(t.TempDir(), "v.mp4")
	require.NoError(t, os.WriteFile(tmp, []byte("x"), 0o600))
	fs := newFlagSetWithArgs(t, tmp)

	_, err := handleUpscaleMode(fs, "", 2, 2, false, true, false, "8M", 0)
	require.ErrorIs(t, err, ErrUpscaleRequested)
	require.NotNil(t, GlobalUpscaleRequest)
	assert.Equal(t, 4, GlobalUpscaleRequest.Passes, "hqMode forces passes=4")
	assert.True(t, GlobalUpscaleRequest.HighQuality)
	assert.InDelta(t, 0.4, GlobalUpscaleRequest.StrengthColor, 0.001)
}

func TestHandleMovieDownloadMode_EmptyArgsErrors(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode(nil, false, false, "best", "english", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires movie/TV name")
}

func TestHandleMovieDownloadMode_Movie(t *testing.T) {
	snapshotGlobalRequest(t)
	name, err := handleMovieDownloadMode([]string{"inception"}, false, false, "best", "english", "")
	require.ErrorIs(t, err, ErrMovieDownloadRequested)
	assert.Equal(t, "inception", name)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsMovie)
}

func TestHandleMovieDownloadMode_TVAllSeasons(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"breaking", "bad"}, false, true, "best", "english", "tv")
	require.ErrorIs(t, err, ErrMovieDownloadRequested)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsAll)
	assert.True(t, GlobalDownloadRequest.IsTV)
}

func TestHandleMovieDownloadMode_TVRangeValid(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "1-5"}, true, false, "best", "english", "tv")
	require.ErrorIs(t, err, ErrMovieDownloadRequested)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsTV)
	assert.True(t, GlobalDownloadRequest.IsRange)
	assert.Equal(t, 1, GlobalDownloadRequest.SeasonNum)
	assert.Equal(t, 1, GlobalDownloadRequest.StartEpisode)
	assert.Equal(t, 5, GlobalDownloadRequest.EndEpisode)
}

func TestHandleMovieDownloadMode_TVRangeMissingArgs(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got"}, true, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVRangeInvalidSeason(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "x", "1-3"}, true, false, "best", "english", "tv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "season number")
}

func TestHandleMovieDownloadMode_TVRangeStartAfterEnd(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "9-2"}, true, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVRangeNegativeSeason(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "0", "1-3"}, true, false, "best", "english", "tv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestHandleMovieDownloadMode_TVRangeInvalidEpisodeFormat(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "bad"}, true, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVRangeInvalidStartNumber(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "x-3"}, true, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVRangeInvalidEndNumber(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "1-x"}, true, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVSingleEpisode(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "3"}, false, false, "best", "english", "tv")
	require.ErrorIs(t, err, ErrMovieDownloadRequested)
	require.NotNil(t, GlobalDownloadRequest)
	assert.True(t, GlobalDownloadRequest.IsTV)
	assert.Equal(t, 1, GlobalDownloadRequest.SeasonNum)
	assert.Equal(t, 3, GlobalDownloadRequest.EpisodeNum)
}

func TestHandleMovieDownloadMode_TVSingleMissingArgs(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got"}, false, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVSingleInvalidSeason(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "x", "3"}, false, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVSingleInvalidEpisode(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "1", "x"}, false, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestHandleMovieDownloadMode_TVSingleNonPositive(t *testing.T) {
	snapshotGlobalRequest(t)
	_, err := handleMovieDownloadMode([]string{"got", "0", "1"}, false, false, "best", "english", "tv")
	require.Error(t, err)
}

func TestDefaultDownloadDir_RespectsGlobalOverride(t *testing.T) {
	snapshotGlobalRequest(t)
	GlobalOutputDir = "/custom/path"
	assert.Equal(t, "/custom/path", DefaultDownloadDir())
}

func TestDefaultDownloadDir_DefaultPath(t *testing.T) {
	snapshotGlobalRequest(t)
	GlobalOutputDir = ""
	got := DefaultDownloadDir()
	assert.True(t, filepath.IsAbs(got) || got != "")
	assert.Contains(t, got, filepath.Join(".local", "goanime", "downloads", "anime"))
}

func TestDefaultMovieDownloadDir_RespectsGlobalOverride(t *testing.T) {
	snapshotGlobalRequest(t)
	GlobalOutputDir = "/custom/movies"
	assert.Equal(t, "/custom/movies", DefaultMovieDownloadDir())
}

func TestDefaultMovieDownloadDir_DefaultPath(t *testing.T) {
	snapshotGlobalRequest(t)
	GlobalOutputDir = ""
	got := DefaultMovieDownloadDir()
	assert.Contains(t, got, filepath.Join(".local", "goanime", "downloads", "movies"))
}

func TestFormatPlexMovieDir_BaseCase(t *testing.T) {
	t.Parallel()
	dir := FormatPlexMovieDir("/movies", "Inception", &MediaMeta{Year: "2010", TMDBID: 27205})
	assert.Equal(t, "/movies/Inception (2010) {tmdb-27205}", dir)
}

func TestFormatPlexMovieDir_NoMeta(t *testing.T) {
	t.Parallel()
	dir := FormatPlexMovieDir("/movies", "Inception")
	assert.Equal(t, "/movies/Inception", dir)
}

func TestFormatPlexEpisodeDir_BaseCase(t *testing.T) {
	t.Parallel()
	dir := FormatPlexEpisodeDir("/anime", "Naruto", 2, &MediaMeta{Year: "2007", AnilistID: 21})
	assert.Equal(t, "/anime/Naruto (2007) {anilist-21}/Season 02", dir)
}

func TestFormatPlexEpisodeDir_DefaultsToSeasonOne(t *testing.T) {
	t.Parallel()
	dir := FormatPlexEpisodeDir("/anime", "Naruto", 0)
	assert.Equal(t, "/anime/Naruto/Season 01", dir)
}

func TestFormatPlexEpisodeDir_HighSeason(t *testing.T) {
	t.Parallel()
	dir := FormatPlexEpisodeDir("/anime", "One Piece", 21, &MediaMeta{Year: "1999"})
	assert.Contains(t, dir, "Season 21")
	if runtime.GOOS == "windows" {
		t.Skip("path separators differ on Windows")
	}
	assert.Equal(t, "/anime/One Piece (1999)/Season 21", dir)
}

// newEmptyFlagSet builds a parsed flag.FlagSet with zero positional args.
func newEmptyFlagSet(t *testing.T) *flag.FlagSet {
	t.Helper()
	return newFlagSetWithArgs(t)
}

func newFlagSetWithArgs(t *testing.T, args ...string) *flag.FlagSet {
	t.Helper()
	fs := flag.NewFlagSet("goanime-test", flag.ContinueOnError)
	require.NoError(t, fs.Parse(args))
	return fs
}
