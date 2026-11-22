//go:build linux

package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TerminalResetSequence is written at every exit. What it does to the cursor
// is a terminal's business, so it is played to one (tmux) after what a run
// writes, and the screen is read back: everything printed must still be there
// and the shell's next line must come below it, written over nothing.
//
// The full-screen views open on the alternate screen. Leaving it restores the
// cursor saved when it was entered, even when the view closed long before the
// exit, and resetting the scroll region homes the cursor: either put the
// prompt on top of what was printed last.
func TestTheResetLeavesThePromptBelowWhatWasPrinted(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux to look at the screen with")
	}
	view := "\x1b[?1049h\x1b[Hthe view\x1b[?1049l" // a full-screen view, opened and closed
	for _, tc := range []struct {
		name, run string
	}{
		{"only printed", "GoAnime v2\n"},
		{"printed after a full-screen view closed", view + "Search cancelled.\nError: no episodes\n"},
		{"two views, then printed", view + "Going back\n" + view + "Search cancelled.\n"},
		{"a view closed, nothing printed", view},
	} {
		t.Run(tc.name, func(t *testing.T) {
			screen := playToTmux(t, "__BEFORE__\n"+tc.run+TerminalResetSequence)
			lines := strings.Split(strings.TrimRight(screen, "\n"), "\n")
			want := []string{"__BEFORE__"}
			for _, l := range strings.Split(strings.ReplaceAll(tc.run, view, ""), "\n") {
				if l != "" {
					want = append(want, l)
				}
			}
			want = append(want, "__AFTER__")
			if len(lines) != len(want) {
				t.Fatalf("the screen is\n%s\nwant the lines %q", screen, want)
			}
			for i := range want {
				if lines[i] != want[i] {
					t.Fatalf("line %d is %q, want %q: something was written over it; the screen is\n%s", i+1, lines[i], want[i], screen)
				}
			}
		})
	}
}

// playToTmux writes out to a fresh tmux pane, then a line as the shell's
// prompt would come back, and returns the non-empty lines of the screen.
func playToTmux(t *testing.T, out string) string {
	t.Helper()
	dir := t.TempDir()
	data := filepath.Join(dir, "out")
	if err := os.WriteFile(data, []byte(out), 0o600); err != nil {
		t.Fatal(err)
	}
	socket := fmt.Sprintf("goanime-reset-%d-%d", os.Getpid(), time.Now().UnixNano())
	tmux := func(a ...string) (string, error) {
		b, err := exec.Command("tmux", append([]string{"-L", socket, "-f", "/dev/null"}, a...)...).CombinedOutput()
		return string(b), err
	}
	t.Cleanup(func() { _, _ = tmux("kill-server") })
	script := "clear; cat " + data + "; echo __AFTER__; sleep 600"
	if b, err := tmux("new-session", "-d", "-s", "p", "-x", "80", "-y", "24", script); err != nil {
		t.Fatalf("tmux: %v\n%s", err, b)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		b, _ := tmux("capture-pane", "-t", "p", "-p")
		if strings.Contains(b, "__AFTER__") {
			time.Sleep(100 * time.Millisecond)
			b, _ = tmux("capture-pane", "-t", "p", "-p")
			var kept []string
			for _, l := range strings.Split(b, "\n") {
				if strings.TrimSpace(l) != "" {
					kept = append(kept, strings.TrimRight(l, " "))
				}
			}
			return strings.Join(kept, "\n") + "\n"
		}
		if time.Now().After(deadline) {
			t.Fatalf("the run never ended; the screen is\n%s", b)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
