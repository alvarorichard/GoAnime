// ===========================================================================
// download_ffmpeg_test.go — Regression tests for the SuperFlix HLS fixes
//
// Issue #193: SuperFlix downloads failed with
//   "ffmpeg is required for SuperFlix HLS downloads:
//    exec: \"ffmpeg\": executable file not found in %PATH%"
// on Windows, where ffmpeg is not installed.
//
// Fix under test: when ffmpeg is missing from PATH, downloadWithFFmpegHLS
// installs a bundled static build via the go-ytdlp cache (installFFmpegFunc /
// installFFprobeFunc) instead of failing. The native HLS downloader cannot
// substitute: SuperFlix serves video-only segments with a separate audio
// group (ErrSeparateAudioTracks), and yt-dlp treats the .txt master playlist
// as a plain file.
//
// NOTE: these tests mutate package-level installer vars and PATH, so they
// must NOT call t.Parallel() (non-parallel tests never overlap parallel ones,
// so this is race-free).
// ===========================================================================

package player

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeFFmpegScript = `#!/bin/sh
out=""
for arg in "$@"; do
    out="$arg"
done
printf 'FAKE-MP4-DATA-0123456789' > "$out"
printf 'out_time_us=1000000\nprogress=continue\n'
`

const fakeFFprobeScript = `#!/bin/sh
printf '123.5\n'
`

const fakeSuperFlixMasterURL = "https://cdn.example.com/cdn/hls/token/master.txt"

// TestMain doubles as the Windows stub for the fake ffmpeg/ffprobe binaries:
// writeFakeExecutable copies the test binary itself, so on Windows (where
// shell scripts are not executable) invoking the fake re-runs this process,
// which detects its own argv[0] basename and behaves like the real tool.
// On Unix the script-based fakes are used instead.
func TestMain(m *testing.M) {
	switch {
	case strings.HasPrefix(filepath.Base(os.Args[0]), "ffprobe"):
		if out := os.Getenv("GOANIME_FAKE_STDOUT"); out != "" {
			fmt.Print(out)
		} else {
			fmt.Println("123.5")
		}
		os.Exit(0)
	case strings.HasPrefix(filepath.Base(os.Args[0]), "ffmpeg"):
		_ = os.WriteFile(os.Args[len(os.Args)-1], []byte("FAKE-MP4-DATA-0123456789"), 0o644)
		fmt.Println("out_time_us=1000000")
		fmt.Println("progress=continue")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func writeFakeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		name += ".exe"
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(os.Args[0])
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, data, 0o755))
		if strings.HasPrefix(name, "ffprobe") {
			stdout := content
			if start := strings.Index(stdout, "'"); start >= 0 {
				if end := strings.LastIndex(stdout, "'"); end > start {
					stdout = stdout[start+1 : end]
				}
			}
			t.Setenv("GOANIME_FAKE_STDOUT", stdout)
		}
		return path
	}
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	return path
}

// assertRefersToSameFile compares by file identity (os.SameFile) instead of
// path strings. Path equality is unreliable across platforms: macOS firmlinks
// (/var -> /private/var), Windows 8.3 short names (RUNNER~1 vs runneradmin)
// and case folding all break string comparisons while resolving to the same
// file. Production normalizes with filepath.EvalSymlinks, which handles some
// of these but not all.
func assertRefersToSameFile(t *testing.T, want, got string) {
	t.Helper()
	wantStat, err := os.Stat(want)
	require.NoError(t, err, "expected path %q does not exist", want)
	gotStat, err := os.Stat(got)
	require.NoError(t, err, "resolved path %q does not exist", got)
	assert.True(t, os.SameFile(wantStat, gotStat), "paths %q and %q must refer to the same file", want, got)
}

// fakeFFmpegHLSOutputDir returns a directory inside the user home, which
// sanitizeOutputPath requires (it rejects paths escaping the home dir).
func fakeFFmpegHLSOutputDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	dir := filepath.Join(home, ".goanime-test-ffmpeg")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// TestResolveFFmpeg_OnPath_SkipsInstaller: an ffmpeg on PATH is used as-is
// and the installer must never be called.
func TestResolveFFmpeg_OnPath_SkipsInstaller(t *testing.T) {
	binDir := t.TempDir()
	fake := writeFakeExecutable(t, binDir, "ffmpeg", fakeFFmpegScript)
	t.Setenv("PATH", binDir)

	installed := false
	path, err := resolveFFmpeg(func(ctx context.Context) (string, error) {
		installed = true
		return "", errors.New("installer must not be called when ffmpeg is on PATH")
	})

	require.NoError(t, err)
	assert.False(t, installed)
	assertRefersToSameFile(t, fake, path)
}

// TestResolveFFmpeg_NotOnPath_Installs: a missing ffmpeg triggers the
// installer and the installed executable path is used.
func TestResolveFFmpeg_NotOnPath_Installs(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	fake := writeFakeExecutable(t, t.TempDir(), "ffmpeg", fakeFFmpegScript)

	installerCalled := false
	path, err := resolveFFmpeg(func(ctx context.Context) (string, error) {
		installerCalled = true
		return fake, nil
	})

	require.NoError(t, err)
	assert.True(t, installerCalled, "missing ffmpeg must trigger the bundled install")
	assertRefersToSameFile(t, fake, path)
}

// TestResolveFFmpeg_InstallerFailure_ReturnsClearError: if the bundled
// install fails, the error must explain that ffmpeg is required (the exact
// message users saw in issue #193) and wrap the underlying cause.
func TestResolveFFmpeg_InstallerFailure_ReturnsClearError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := resolveFFmpeg(func(ctx context.Context) (string, error) {
		return "", errors.New("network down")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ffmpeg is required for SuperFlix HLS downloads")
	assert.Contains(t, err.Error(), "network down")
}

// TestResolveFFmpeg_InstallerReturnsBrokenPath: an installer returning a
// non-existent or relative path must not be used.
func TestResolveFFmpeg_InstallerReturnsBrokenPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := resolveFFmpeg(func(ctx context.Context) (string, error) {
		return filepath.Join(t.TempDir(), "nonexistent-ffmpeg"), nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ffmpeg executable path")
}

// TestResolveFFprobe_OnPath: ffprobe from PATH wins, installer is skipped.
func TestResolveFFprobe_OnPath(t *testing.T) {
	binDir := t.TempDir()
	fake := writeFakeExecutable(t, binDir, "ffprobe", fakeFFprobeScript)
	t.Setenv("PATH", binDir)

	installed := false
	got := resolveFFprobe(func(ctx context.Context) (string, error) {
		installed = true
		return "", errors.New("installer must not be called when ffprobe is on PATH")
	})

	assert.False(t, installed)
	assertRefersToSameFile(t, fake, got)
}

// TestResolveFFprobe_NotOnPath_Installs: missing ffprobe triggers the
// bundled install (shipped in the same archive as ffmpeg).
func TestResolveFFprobe_NotOnPath_Installs(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	fake := writeFakeExecutable(t, t.TempDir(), "ffprobe", fakeFFprobeScript)

	got := resolveFFprobe(func(ctx context.Context) (string, error) {
		return fake, nil
	})

	assertRefersToSameFile(t, fake, got)
}

// TestResolveFFprobe_Nowhere_ReturnsEmpty: when neither PATH nor the
// installer yields ffprobe, resolution must return "" (the download still
// proceeds; only duration-based progress is lost).
func TestResolveFFprobe_Nowhere_ReturnsEmpty(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	got := resolveFFprobe(func(ctx context.Context) (string, error) {
		return "", errors.New("no ffprobe available")
	})

	assert.Empty(t, got)
}

// TestProbeHLSDuration_WithExplicitFFprobePath: probeHLSDuration must accept
// the ffprobe path resolved by resolveFFprobe (bundled build path, not on
// PATH) and parse the duration it prints.
func TestProbeHLSDuration_WithExplicitFFprobePath(t *testing.T) {
	fake := writeFakeExecutable(t, t.TempDir(), "ffprobe", fakeFFprobeScript)

	d, err := probeHLSDuration(t.Context(), fakeSuperFlixMasterURL, "https://cdn.example.com/", fake)

	require.NoError(t, err)
	assert.Equal(t, 123*time.Second+500*time.Millisecond, d)
}

// TestProbeHLSDuration_InvalidOutput: garbage on ffprobe stdout must be
// rejected, not silently used.
func TestProbeHLSDuration_InvalidOutput(t *testing.T) {
	fake := writeFakeExecutable(t, t.TempDir(), "ffprobe", "#!/bin/sh\nprintf 'not-a-number\\n'\n")

	_, err := probeHLSDuration(t.Context(), fakeSuperFlixMasterURL, "", fake)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HLS duration")
}

// TestProbeHLSDuration_InvalidPath: a non-existent or relative ffprobe path
// must be rejected.
func TestProbeHLSDuration_InvalidPath(t *testing.T) {
	_, err := probeHLSDuration(t.Context(), fakeSuperFlixMasterURL, "", filepath.Join(t.TempDir(), "nope"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ffprobe executable path")
}

// TestProbeHLSDuration_NoPathAndNotOnPATH: with no explicit path and no
// ffprobe on PATH, probing fails (callers treat this as non-fatal).
func TestProbeHLSDuration_NoPathAndNotOnPATH(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := probeHLSDuration(t.Context(), fakeSuperFlixMasterURL, "", "")

	require.Error(t, err)
}

// TestDownloadWithFFmpegHLS_EndToEnd runs the whole SuperFlix HLS download
// with a fake ffmpeg found on PATH: it must write the output, parse the
// progress pipe, rename the .part file into place and clean up after itself.
func TestDownloadWithFFmpegHLS_EndToEnd(t *testing.T) {
	binDir := t.TempDir()
	writeFakeExecutable(t, binDir, "ffmpeg", fakeFFmpegScript)
	writeFakeExecutable(t, binDir, "ffprobe", fakeFFprobeScript)
	t.Setenv("PATH", binDir)

	outFile := filepath.Join(fakeFFmpegHLSOutputDir(t), "episode.mp4")

	err := downloadWithFFmpegHLS(fakeSuperFlixMasterURL, outFile, nil)

	require.NoError(t, err, "download must succeed with ffmpeg on PATH")
	data, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Equal(t, "FAKE-MP4-DATA-0123456789", string(data), "output must contain the media produced by ffmpeg")

	_, err = os.Stat(partialMediaPath(outFile))
	assert.True(t, os.IsNotExist(err), "the .part file must be renamed away after success")
}

// TestDownloadWithFFmpegHLS_InstallsBundledFFmpegWhenMissing is the direct
// regression test for issue #193's second error: on a machine without
// ffmpeg (PATH with no ffmpeg binary), the download must still succeed by
// installing the bundled build instead of failing with "ffmpeg is required".
func TestDownloadWithFFmpegHLS_InstallsBundledFFmpegWhenMissing(t *testing.T) {
	binDir := t.TempDir()
	writeFakeExecutable(t, binDir, "ffprobe", fakeFFprobeScript)
	t.Setenv("PATH", binDir)

	fakeFFmpeg := writeFakeExecutable(t, t.TempDir(), "ffmpeg", fakeFFmpegScript)
	origInstall := installFFmpegFunc
	installFFmpegFunc = func(ctx context.Context) (string, error) {
		return fakeFFmpeg, nil
	}
	t.Cleanup(func() { installFFmpegFunc = origInstall })

	outFile := filepath.Join(fakeFFmpegHLSOutputDir(t), "episode.mp4")

	err := downloadWithFFmpegHLS(fakeSuperFlixMasterURL, outFile, nil)

	require.NoError(t, err, "missing ffmpeg must trigger the bundled install and still download")
	data, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Equal(t, "FAKE-MP4-DATA-0123456789", string(data))
}

// TestDownloadWithFFmpegHLS_InstallFailure: when the bundled ffmpeg install
// itself fails, the download must fail with the "ffmpeg is required" message
// (never a silent success, never a native-HLS fallback that would produce a
// video-only file for SuperFlix's separate-audio master).
func TestDownloadWithFFmpegHLS_InstallFailure(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	origInstall := installFFmpegFunc
	installFFmpegFunc = func(ctx context.Context) (string, error) {
		return "", errors.New("failed to download ffmpeg build")
	}
	t.Cleanup(func() { installFFmpegFunc = origInstall })

	outFile := filepath.Join(fakeFFmpegHLSOutputDir(t), "episode.mp4")

	err := downloadWithFFmpegHLS(fakeSuperFlixMasterURL, outFile, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ffmpeg is required for SuperFlix HLS downloads")
	assert.Contains(t, err.Error(), "failed to download ffmpeg build")
	_, statErr := os.Stat(outFile)
	assert.True(t, os.IsNotExist(statErr), "no output file must be produced when ffmpeg cannot be resolved")
}
