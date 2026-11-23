//go:build linux

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// A user who puts the binary on PATH and runs `goanime --version` in a
// terminal must see the answer. Every exit used to clear the viewport, so the
// answer was pushed into the scrollback the moment it was printed and the
// screen came back blank (--version, -h, a flag error alike). Here the real
// binary runs on a real pseudo-terminal, as in a shell, and what it writes to
// that terminal is read back.

// clearScreen is the erase the exit cleanup writes (tui.ClearViewport).
const clearScreen = "\x1b[2J"

// openPTY returns the controlling and the terminal side of a new
// pseudo-terminal.
func openPTY(t *testing.T) (ptm, pts *os.File) {
	t.Helper()
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Skipf("no pseudo-terminal here: %v", err)
	}
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		_ = unix.Close(fd)
		t.Fatalf("unlock pty: %v", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		_ = unix.Close(fd)
		t.Fatalf("pty number: %v", err)
	}
	ptm = os.NewFile(uintptr(fd), "/dev/ptmx")
	pts, err = os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		_ = ptm.Close()
		t.Fatalf("open pts: %v", err)
	}
	if err := unix.IoctlSetWinsize(int(pts.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 40, Col: 120}); err != nil {
		t.Fatalf("pty size: %v", err)
	}
	t.Cleanup(func() { _ = ptm.Close(); _ = pts.Close() })
	return ptm, pts
}

// buildGoAnime builds this command once per test binary.
var builtGoAnime string

func buildGoAnime(t *testing.T) string {
	t.Helper()
	if builtGoAnime != "" {
		return builtGoAnime
	}
	bin := filepath.Join(os.TempDir(), fmt.Sprintf("goanime-exit-screen-%d", os.Getpid()))
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	builtGoAnime = bin
	return bin
}

func TestMain(m *testing.M) {
	code := m.Run()
	if builtGoAnime != "" {
		_ = os.Remove(builtGoAnime)
	}
	os.Exit(code)
}

// isolatedEnv keeps a run away from the user's history, settings and caches.
func isolatedEnv(home string) []string {
	return []string{
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"XDG_DATA_HOME=" + filepath.Join(home, ".local", "share"),
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"PATH=" + os.Getenv("PATH"),
		// No background tool downloads from a test.
		"GOANIME_CFSOLVE_AUTODOWNLOAD=0",
	}
}

// runOnTerminal runs goanime with args on a pseudo-terminal of its own, in a
// session it leads, and returns everything it wrote to the terminal.
func runOnTerminal(t *testing.T, args ...string) string {
	t.Helper()
	// Not skipped under -short: CI runs -short, and this is the only test that
	// sees what a user's terminal shows (a build and a few runs, ~3 s).
	bin := buildGoAnime(t)
	ptm, pts := openPTY(t)

	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = pts, pts, pts
	cmd.Env = append(isolatedEnv(t.TempDir()), "TERM=xterm-256color")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	_ = pts.Close() // the child holds its own

	var out bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(&out, ptm) // ends with EIO once the child is gone
	}()

	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	select {
	case err := <-waited:
		var exit *exec.ExitError
		if err != nil && !errors.As(err, &exit) {
			t.Fatalf("wait: %v", err)
		}
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("goanime %s did not exit; it wrote:\n%q", strings.Join(args, " "), out.String())
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	return out.String()
}

// visibleAfterExit is what is left on the screen: the output after the last
// erase of the screen.
func visibleAfterExit(out string) string {
	if i := strings.LastIndex(out, clearScreen); i >= 0 {
		return out[i+len(clearScreen):]
	}
	return out
}

func TestWhatACommandPrintsStaysOnTheScreen(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--version"}, "GoAnime v"},
		{[]string{"-h"}, "Beautiful Anime Streaming CLI"},
		{[]string{"--help"}, "Beautiful Anime Streaming CLI"},
		{[]string{"--no-such-flag"}, "flag provided but not defined"},
		{[]string{"--source", "naoexiste", "naruto"}, `unknown source "naoexiste"`},
		{[]string{"-d", "--source", "animfire", "naruto", "1"}, `unknown source "animfire"`},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			out := runOnTerminal(t, tc.args...)
			if !strings.Contains(out, tc.want) {
				t.Fatalf("goanime %s never printed %q; it wrote:\n%q", strings.Join(tc.args, " "), tc.want, out)
			}
			if !strings.Contains(visibleAfterExit(out), tc.want) {
				t.Fatalf("goanime %s printed %q and then cleared the screen, so the user sees a blank terminal", strings.Join(tc.args, " "), tc.want)
			}
		})
	}
}

// What the user sees once the prompt is back is a terminal's business, not a
// byte stream's: tmux is the terminal here. Each run prints a line before
// GoAnime and one after it, as the shell's prompt would come back; both must
// be on the screen, in that order, with what GoAnime printed between them and
// nothing written over any of it. This is what caught the reset of the scroll
// region sending the cursor to the top of the screen, so the prompt landed on
// GoAnime's last lines.
func TestTheScreenAfterExitAsAUserSeesIt(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux to look at the screen with")
	}
	bin := buildGoAnime(t)
	for _, tc := range []struct {
		name string
		args []string
		// ready is on the screen once the run waits for a key; then keys are sent.
		ready, keys string
		want        string
	}{
		{name: "version", args: []string{"--version"}, want: "GoAnime v"},
		{name: "help", args: []string{"-h"}, want: "Beautiful Anime Streaming CLI"},
		{name: "flag error", args: []string{"--no-such-flag"}, want: "flag provided but not defined"},
		{name: "unknown source", args: []string{"--source", "naoexiste", "naruto"}, want: `unknown source "naoexiste"`},
		// The search home is a full-screen TUI on the alternate screen.
		{name: "search home closed", args: nil, ready: "quit", keys: "C-c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			screen := screenAfterExit(t, bin, tc.args, tc.ready, tc.keys)
			before := strings.Index(screen, "\n__BEFORE__\n")
			after := strings.Index(screen, "\n__AFTER__\n")
			if before < 0 || after < 0 || after < before {
				t.Fatalf("the lines around the run were written over; the screen is:\n%s", screen)
			}
			if tc.want != "" && !strings.Contains(screen[before:after], tc.want) {
				t.Fatalf("%q is not on the screen between the lines around the run; the screen is:\n%s", tc.want, screen)
			}
		})
	}
}

// screenAfterExit runs goanime in a tmux pane of its own server, sends keys
// once ready shows (when keys is set), waits for the run to end and returns
// the visible screen with a newline before every line. The pane is tall
// enough for the whole --help (about 200 lines).
func screenAfterExit(t *testing.T, bin string, args []string, ready, keys string) string {
	t.Helper()
	dir := t.TempDir()
	socket := fmt.Sprintf("goanime-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	tmux := func(a ...string) (string, error) {
		out, err := exec.Command("tmux", append([]string{"-L", socket, "-f", "/dev/null"}, a...)...).CombinedOutput()
		return string(out), err
	}
	t.Cleanup(func() { _, _ = tmux("kill-server") })

	quoted := make([]string, 0, len(args))
	for _, a := range args {
		quoted = append(quoted, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
	}
	env := isolatedEnv(filepath.Join(dir, "home"))
	script := "clear\necho __BEFORE__\n" +
		"env -i " + strings.Join(append(env, "TERM=xterm-256color"), " ") + " " + bin + " " + strings.Join(quoted, " ") + "\n" +
		"echo __AFTER__\nsleep 600\n"
	path := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if out, err := tmux("new-session", "-d", "-s", "run", "-x", "120", "-y", "250", "sh "+path); err != nil {
		t.Fatalf("tmux: %v\n%s", err, out)
	}
	capture := func() string {
		out, _ := tmux("capture-pane", "-t", "run", "-p")
		return "\n" + out
	}
	waitFor := func(text string) string {
		deadline := time.Now().Add(30 * time.Second)
		for {
			s := capture()
			if strings.Contains(s, text) {
				return s
			}
			if time.Now().After(deadline) {
				t.Fatalf("%q never showed; the screen is:\n%s", text, s)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if keys != "" {
		waitFor(ready)
		time.Sleep(300 * time.Millisecond) // let the screen take its first keys
		if out, err := tmux("send-keys", "-t", "run", keys); err != nil {
			t.Fatalf("tmux send-keys: %v\n%s", err, out)
		}
	}
	waitFor("__AFTER__") // anywhere: written over another line, it still ends the wait
	time.Sleep(200 * time.Millisecond)
	return capture()
}
