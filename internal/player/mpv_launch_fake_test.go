//go:build !windows

package player

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A stand-in for mpv, built for the test: a program that does what it is told
// by a file next to it. Not parallel: the choice of mpv is process-wide.

const fakeMPVSource = `package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func appendLine(path, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err == nil {
		_, _ = f.WriteString(line + "\n")
		_ = f.Close()
	}
}

func fail(logFile string, code int, msg string) {
	if logFile != "" {
		_ = os.WriteFile(logFile, []byte("[  0.003][e][cplayer] "+msg+"\n"), 0o600)
	}
	os.Exit(code)
}

func main() {
	exe, _ := os.Executable()
	args := os.Args[1:]
	var sock, logFile string
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--input-ipc-server="); ok {
			sock = v
		}
		if v, ok := strings.CutPrefix(a, "--log-file="); ok {
			logFile = v
		}
	}
	if rec := os.Getenv("FAKE_MPV_RECORD"); rec != "" {
		appendLine(rec, exe+"\x1f"+strings.Join(args, "\x1f"))
	}
	appendLine(exe+".pids", strconv.Itoa(os.Getpid()))

	raw, _ := os.ReadFile(exe + ".behavior")
	behavior := strings.TrimSpace(string(raw))
	switch {
	case strings.HasPrefix(behavior, "reject="):
		for _, bad := range strings.Split(strings.TrimPrefix(behavior, "reject="), ",") {
			for _, a := range args {
				if bad != "" && strings.HasPrefix(a, bad) {
					fail(logFile, 1, "Error parsing option "+a+" (option not found)")
				}
			}
		}
	case strings.HasPrefix(behavior, "exit="):
		code, _ := strconv.Atoi(strings.TrimPrefix(behavior, "exit="))
		fail(logFile, code, "exiting as told")
	case behavior == "hang":
		time.Sleep(time.Minute)
		return
	case behavior == "hangnoisy":
		// Up, never opening its socket, and writing to stderr all the while.
		for i := 0; i < 60000; i++ {
			_, _ = os.Stderr.WriteString("still starting up, driver is slow\n")
			time.Sleep(time.Millisecond)
		}
		return
	}
	ln, err := net.Listen("unix", sock)
	if err != nil {
		fail(logFile, 1, "listen: "+err.Error())
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_ = c.Close()
	}
}
`

type fakeMPVs struct {
	t      *testing.T
	dir    string
	binary string
	record string
}

type fakeLaunch struct {
	exe  string
	args []string
}

// newFakeMPVs compiles the stand-in and records every launch of it.
func newFakeMPVs(t *testing.T) *fakeMPVs {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build a stand-in mpv with")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(src, []byte(fakeMPVSource), 0o600))
	binary := filepath.Join(dir, "fakempv")
	build := exec.Command(goBin, "build", "-o", binary, src) // #nosec G204 -- the go toolchain, building a file this test wrote
	build.Dir = dir
	out, err := build.CombinedOutput()
	require.NoError(t, err, "building the stand-in mpv:\n%s", out)

	f := &fakeMPVs{t: t, dir: dir, binary: binary, record: filepath.Join(dir, "launches")}
	t.Setenv("FAKE_MPV_RECORD", f.record)

	// mpv's log file and IPC socket go in the temp directory. A directory of
	// their own keeps them out of the shared one, and has to be short: a unix
	// socket path is limited to about 100 bytes, which t.TempDir() can exceed.
	tmp, err := os.MkdirTemp("", "g")
	require.NoError(t, err)
	t.Setenv("TMPDIR", tmp)

	resetMPVChoice()
	old := mpvStartTimeout
	t.Cleanup(func() {
		mpvStartTimeout = old
		resetMPVChoice()
		f.killAll()
		// Each launch's log is removed by a goroutine once its mpv is gone;
		// wait for that so none is left behind.
		for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
			if left, _ := filepath.Glob(filepath.Join(tmp, "goanime_mpv_*.log")); len(left) == 0 {
				break
			}
		}
		_ = os.RemoveAll(tmp)
	})
	return f
}

// install makes an mpv at name that behaves as told: "" comes up,
// "reject=--vo=,--profile=" dies on those options, "exit=2" dies with that
// code, "hang" never opens its socket, "hangnoisy" does not either and keeps
// writing to stderr.
func (f *fakeMPVs) install(name, behavior string) string {
	f.t.Helper()
	dir := filepath.Join(f.dir, name)
	require.NoError(f.t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "mpv")
	data, err := os.ReadFile(f.binary)
	require.NoError(f.t, err)
	require.NoError(f.t, os.WriteFile(path, data, 0o755)) // #nosec G306 -- a test executable
	require.NoError(f.t, os.WriteFile(path+".behavior", []byte(behavior), 0o600))
	return path
}

func (f *fakeMPVs) launches() []fakeLaunch {
	f.t.Helper()
	raw, err := os.ReadFile(f.record)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	require.NoError(f.t, err)
	var out []fakeLaunch
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		parts := strings.Split(line, "\x1f")
		out = append(out, fakeLaunch{exe: parts[0], args: parts[1:]})
	}
	return out
}

func (f *fakeMPVs) launched() []string {
	var names []string
	for _, l := range f.launches() {
		names = append(names, filepath.Base(filepath.Dir(l.exe)))
	}
	return names
}

func (f *fakeMPVs) killAll() {
	matches, _ := filepath.Glob(filepath.Join(f.dir, "*", "mpv.pids"))
	for _, m := range matches {
		raw, _ := os.ReadFile(m)
		for _, line := range strings.Fields(string(raw)) {
			if pid, err := strconv.Atoi(line); err == nil {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	}
}

func hasArg(args []string, prefix string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			return true
		}
	}
	return false
}

var playbackArgs = []string{
	"--cache=yes", "--demuxer-max-bytes=300M", "--hwdec=auto-safe", "--vo=gpu-next,gpu,direct3d,sdl", "--profile=fast",
	"--no-config", "--http-header-fields=Referer: https://streameeeeee.site/", "--force-media-title=Show S02E07",
}

const streamLink = "https://cdn.test/stream/07.mp4/index.m3u8"

func TestStartVideoAcrossStartsTheFirstMPVThatComesUp(t *testing.T) {
	f := newFakeMPVs(t)
	good := f.install("good", "")

	socket, err := startVideoAcross([]string{good}, streamLink, playbackArgs)

	require.NoError(t, err)
	assert.NotEmpty(t, socket)
	require.Len(t, f.launches(), 1)
	launch := f.launches()[0]
	assert.Equal(t, streamLink, launch.args[len(launch.args)-1], "mpv is given the link")
	assert.True(t, hasArg(launch.args, "--vo="), "a working mpv gets GoAnime's video options")
	assert.True(t, hasArg(launch.args, "--input-ipc-server="))
}

// The reported case: PATH hands over a stub that exits with status 1, and the
// mpv the installer shipped is further down the list.
func TestStartVideoAcrossFallsThroughToTheNextMPVWhenThePathOneExits(t *testing.T) {
	f := newFakeMPVs(t)
	stub := f.install("WindowsApps", "exit=1")
	bundled := f.install("bundled", "")

	socket, err := startVideoAcross([]string{stub, bundled}, streamLink, playbackArgs)

	require.NoError(t, err)
	assert.NotEmpty(t, socket)
	assert.Equal(t, []string{"WindowsApps", "bundled"}, f.launched(), "the stub once, with the options as they were, then the one that works")
	assert.True(t, hasArg(f.launches()[1].args, "--vo="), "and the working one is not stripped of its options")
}

func TestStartVideoAcrossRemembersTheMPVThatWorkedSoTheStubIsNotTriedAgain(t *testing.T) {
	f := newFakeMPVs(t)
	stub := f.install("stub", "exit=1")
	bundled := f.install("bundled", "")
	paths := []string{stub, bundled}

	_, err := startVideoAcross(paths, streamLink, playbackArgs)
	require.NoError(t, err)
	before := len(f.launches())

	_, err = startVideoAcross(paths, streamLink, playbackArgs)
	require.NoError(t, err)

	assert.Equal(t, before+1, len(f.launches()), "the next episode goes straight to the mpv that worked")
	assert.Equal(t, "bundled", f.launched()[len(f.launched())-1])
}

func TestStartVideoAcrossFallsThroughWhenAnMPVCannotBeStartedAtAll(t *testing.T) {
	f := newFakeMPVs(t)
	notRunnable := filepath.Join(f.dir, "broken-mpv")
	require.NoError(t, os.WriteFile(notRunnable, []byte("not a program"), 0o600)) // no exec bit
	good := f.install("good", "")

	socket, err := startVideoAcross([]string{notRunnable, good}, streamLink, playbackArgs)

	require.NoError(t, err)
	assert.NotEmpty(t, socket)
	assert.Equal(t, []string{"good"}, f.launched())
}

func TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects(t *testing.T) {
	f := newFakeMPVs(t)
	picky := f.install("picky", "reject=--vo=")

	socket, err := startVideoAcross([]string{picky}, streamLink, playbackArgs)

	require.NoError(t, err, "the same mpv plays once the option it rejects is left out")
	assert.NotEmpty(t, socket)
	launches := f.launches()
	require.Len(t, launches, 2)
	assert.True(t, hasArg(launches[0].args, "--vo="), "first as GoAnime wants it")
	assert.False(t, hasArg(launches[1].args, "--vo="), "then without the tuning")
	assert.False(t, hasArg(launches[1].args, "--hwdec="))
	assert.False(t, hasArg(launches[1].args, "--cache="))
	assert.True(t, hasArg(launches[1].args, "--http-header-fields="), "but with what the stream needs")
	assert.True(t, hasArg(launches[1].args, "--force-media-title="))
	assert.True(t, hasArg(launches[1].args, "--no-config"))
	assert.Equal(t, streamLink, launches[1].args[len(launches[1].args)-1])
}

func TestStartVideoAcrossGoesStraightToTheMinimalOptionsForAnMPVThatNeedsThem(t *testing.T) {
	f := newFakeMPVs(t)
	picky := f.install("picky", "reject=--vo=")

	_, err := startVideoAcross([]string{picky}, streamLink, playbackArgs)
	require.NoError(t, err)
	before := len(f.launches())

	_, err = startVideoAcross([]string{picky}, streamLink, playbackArgs)
	require.NoError(t, err)

	require.Len(t, f.launches(), before+1, "no failed attempt first, the second time")
	assert.False(t, hasArg(f.launches()[before].args, "--vo="))
}

func TestStartVideoAcrossTriesEveryMPVBeforeStrippingOptions(t *testing.T) {
	// A second mpv that plays is a better answer than the first one without
	// its tuning: the stripped retry comes last.
	f := newFakeMPVs(t)
	picky := f.install("picky", "reject=--vo=")
	good := f.install("good", "")

	_, err := startVideoAcross([]string{picky, good}, streamLink, playbackArgs)

	require.NoError(t, err)
	assert.Equal(t, []string{"picky", "good"}, f.launched(), "no stripped retry of the first while another mpv is left")
	assert.True(t, hasArg(f.launches()[1].args, "--vo="))
}

func TestStartVideoAcrossDoesNotStripOptionsAfterAnExitThatIsNotAnInitFailure(t *testing.T) {
	f := newFakeMPVs(t)
	cannotPlay := f.install("cannotplay", "exit=2") // mpv's "could not play the file"

	_, err := startVideoAcross([]string{cannotPlay}, streamLink, playbackArgs)

	require.Error(t, err)
	assert.Equal(t, []string{"cannotplay"}, f.launched(), "fewer options would not make the file playable")
}

func TestStartVideoAcrossHasNothingToStripWhenThereIsNoTuning(t *testing.T) {
	f := newFakeMPVs(t)
	stub := f.install("stub", "exit=1")

	_, err := startVideoAcross([]string{stub}, streamLink, []string{"--no-config", "--start=90"})

	require.Error(t, err)
	assert.Equal(t, []string{"stub"}, f.launched(), "the retry would be the same launch")
}

func TestStartVideoAcrossReportsWhatEachMPVSaid(t *testing.T) {
	f := newFakeMPVs(t)
	// --no-config is something playback needs, so leaving the tuning out does
	// not help this one.
	first := f.install("first", "reject=--no-config")
	second := f.install("second", "exit=3")

	_, err := startVideoAcross([]string{first, second}, streamLink, playbackArgs)

	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "mpv exited before IPC socket was ready", "the first mpv's own failure leads")
	assert.Contains(t, msg, "Error parsing option --no-config", "with what mpv itself logged")
	assert.Contains(t, msg, "Also tried "+second, "then the other mpv")
	assert.Contains(t, msg, "did not help", "and that a retry without the tuning was made")
	assert.Contains(t, msg, "--vo=gpu-next,gpu,direct3d,sdl", "naming what was left out")
	assert.Equal(t, []string{"first", "second", "first"}, f.launched(), "each mpv as asked, then the first again without the tuning")
	var early *mpvEarlyExitError
	assert.ErrorAs(t, err, &early)
	assert.Equal(t, 1, early.code, "the first mpv's own exit code")
}

// A timeout reads mpv's stderr while mpv is still writing it. Under -race this
// is a data race if the buffer is not locked.
func TestStartVideoAcrossReadsStderrSafelyWhileMPVIsStillWritingIt(t *testing.T) {
	f := newFakeMPVs(t)
	mpvStartTimeout = 500 * time.Millisecond
	noisy := f.install("noisy", "hangnoisy")

	_, err := startVideoAcross([]string{noisy}, streamLink, playbackArgs)

	require.Error(t, err)
	assert.ErrorContains(t, err, "timeout waiting for mpv socket")
	assert.ErrorContains(t, err, "still starting up, driver is slow", "what mpv said while it waited is in the error")
}

func TestStartVideoAcrossDoesNotWaitOnASecondMPVWhenTheFirstCameUpAndWentSilent(t *testing.T) {
	f := newFakeMPVs(t)
	mpvStartTimeout = 400 * time.Millisecond
	slow := f.install("slow", "hang")
	good := f.install("good", "")

	start := time.Now()
	_, err := startVideoAcross([]string{slow, good}, streamLink, playbackArgs)

	require.Error(t, err)
	assert.ErrorContains(t, err, "timeout waiting for mpv socket")
	assert.Equal(t, []string{"slow"}, f.launched(), "another mpv would not be quicker, and the wait is already spent")
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestStartVideoAcrossRefusesABadLinkBeforeLaunchingAnything(t *testing.T) {
	f := newFakeMPVs(t)
	good := f.install("good", "")

	_, err := startVideoAcross([]string{good}, "http://bad\nurl", playbackArgs)

	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid media target")
	assert.Empty(t, f.launches())
}

func TestStartVideoAcrossLeavesTheCallersArgsAlone(t *testing.T) {
	f := newFakeMPVs(t)
	picky := f.install("picky", "reject=--vo=")
	args := append([]string(nil), playbackArgs...)

	_, err := startVideoAcross([]string{picky}, streamLink, args)

	require.NoError(t, err)
	assert.Equal(t, playbackArgs, args, "stripping builds a new list")
}

func TestStartVideoAcrossDoesNotRetryAMinimalLaunchWithTheMinimalOptions(t *testing.T) {
	f := newFakeMPVs(t)
	picky := f.install("picky", "reject=--vo=")
	paths := []string{picky}
	_, err := startVideoAcross(paths, streamLink, playbackArgs) // learns that it needs the minimal options
	require.NoError(t, err)
	before := len(f.launches())

	// Now it fails whatever it is given.
	require.NoError(t, os.WriteFile(picky+".behavior", []byte("exit=1"), 0o600))
	_, err = startVideoAcross(paths, streamLink, playbackArgs)

	require.Error(t, err)
	assert.Equal(t, before+1, len(f.launches()), "the launch was already the minimal one; the same again would change nothing")
}

// With a stub first on PATH and an mpv that only plays without the tuning
// second, the next episode must not start by trying the stub again.
func TestStartVideoAcrossRemembersWhichMPVNeededTheMinimalOptions(t *testing.T) {
	f := newFakeMPVs(t)
	stub := f.install("stub", "exit=1")
	picky := f.install("picky", "reject=--vo=")
	paths := []string{stub, picky}

	_, err := startVideoAcross(paths, streamLink, playbackArgs)
	require.NoError(t, err)
	assert.Equal(t, []string{"stub", "picky", "stub", "picky"}, f.launched(),
		"each as asked, then each again without the tuning; the second mpv comes up")
	before := len(f.launches())

	_, err = startVideoAcross(paths, streamLink, playbackArgs)

	require.NoError(t, err)
	assert.Equal(t, []string{"picky"}, f.launched()[before:], "straight to the mpv that works, with the options it needs")
	assert.False(t, hasArg(f.launches()[before].args, "--vo="))
}
