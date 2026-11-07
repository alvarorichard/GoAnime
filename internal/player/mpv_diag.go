package player

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// lockedBuffer is mpv's stderr: os/exec copies into it from its own goroutine
// while the launch may already be reading it (a timeout reads it with mpv still
// running), so a bare bytes.Buffer would be a data race.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *lockedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

// mpvLogMaxLines bounds how much of mpv's own log is kept for an error or the
// debug log; a verbose startup is a few hundred lines.
const mpvLogMaxLines = 300

// mpvExit is how a launched mpv ended, with its own log read before the file
// is removed.
type mpvExit struct {
	err     error
	code    int // -1 when mpv did not report one (signal, wait error)
	log     string
	runtime time.Duration
}

// mpvExitCode extracts the process exit code from a Wait error.
func mpvExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// tailLines keeps the last n lines of s.
func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// mpvProblemLines is what goes into the user-facing error: mpv's fatal, error
// and warning lines ("[f]", "[e]", "[w]"), which verbose mode buries among
// hundreds of debug lines. With none, the plain tail.
func mpvProblemLines(log string, n int) string {
	var keep []string
	for _, l := range strings.Split(log, "\n") {
		if strings.Contains(l, "][f]") || strings.Contains(l, "][e]") || strings.Contains(l, "][w]") {
			keep = append(keep, l)
		}
	}
	if len(keep) == 0 {
		return tailLines(log, n)
	}
	return tailLines(strings.Join(keep, "\n"), n)
}

var (
	mpvInfoOnce sync.Once
	mpvInfo     string
)

// mpvVersionInfo runs `mpv --version` once (5 s cap) so a log shows which build
// failed: the version, its ffmpeg/libplacebo and the build flags.
func mpvVersionInfo(mpvPath string) string {
	mpvInfoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, mpvPath, "--no-config", "--version").CombinedOutput() // #nosec G204 -- the mpv we already located, fixed args
		mpvInfo = strings.TrimSpace(string(out))
		if err != nil {
			mpvInfo = strings.TrimSpace(mpvInfo + "\n(mpv --version failed: " + err.Error() + ")")
		}
	})
	return mpvInfo
}

// logMPVLaunch writes the facts that explain most "mpv does not open" reports.
func logMPVLaunch(mpvPath string, args []string) {
	if !util.IsDebug {
		return
	}
	util.Debugf("mpv launch: os=%s/%s path=%s", runtime.GOOS, runtime.GOARCH, mpvPath)
	if fi, err := os.Lstat(mpvPath); err == nil {
		util.Debugf("mpv binary: size=%d mode=%v modtime=%s", fi.Size(), fi.Mode(), fi.ModTime().Format(time.RFC3339))
		if fi.Mode()&os.ModeSymlink != 0 || fi.Mode()&os.ModeIrregular != 0 || (runtime.GOOS == "windows" && fi.Size() == 0) {
			// A Windows "App Execution Alias" (WindowsApps\mpv.exe) is a zero-byte
			// reparse point, not the real executable.
			util.Debugf("mpv binary is a link/alias, not a plain file; real target: %s", resolvedPath(mpvPath))
		}
	} else {
		util.Debugf("mpv binary stat failed: %v", err)
	}
	for _, line := range strings.Split(mpvVersionInfo(mpvPath), "\n") {
		util.Debugf("mpv version: %s", line)
	}
	env := []string{"DISPLAY", "WAYLAND_DISPLAY", "XDG_SESSION_TYPE", "XDG_RUNTIME_DIR", "MPV_HOME", "SESSIONNAME"}
	var set []string
	for _, k := range env {
		if v := os.Getenv(k); v != "" {
			set = append(set, k+"="+v)
		} else {
			set = append(set, k+"=<unset>")
		}
	}
	util.Debugf("mpv environment: %s", strings.Join(set, " "))
	util.Debugf("mpv args (%d): %q", len(args), args)
}

func resolvedPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return "<unresolved>"
}

// reapMPV waits for cmd in the background, reads mpv's log file when it ends,
// logs how it ended and removes the file. The result is delivered on the
// returned channel (buffered: nobody has to be listening).
func reapMPV(cmd *exec.Cmd, logPath string, stderr fmt.Stringer, started time.Time) <-chan mpvExit {
	ch := make(chan mpvExit, 1)
	go func() {
		err := cmd.Wait()
		ex := mpvExit{err: err, code: mpvExitCode(err), runtime: time.Since(started)}
		if logPath != "" {
			if data, rerr := os.ReadFile(logPath); rerr == nil { // #nosec G304 -- our own temp file
				ex.log = tailLines(string(data), mpvLogMaxLines)
			}
			_ = os.Remove(logPath)
		}
		util.Debugf("mpv exited: code=%d err=%v after %s", ex.code, err, ex.runtime.Round(time.Millisecond))
		if s := strings.TrimSpace(stderr.String()); s != "" {
			util.Debugf("mpv stderr:\n%s", s)
		}
		if ex.log != "" && (util.IsDebug || ex.code != 0) {
			util.Debugf("mpv log (last lines):\n%s", ex.log)
		}
		ch <- ex
	}()
	return ch
}
