package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTerminalResetSequenceCompleteness locks in the exact set of resets the exit
// cleanup must emit. A broken shell prompt after GoAnime exits is caused by one
// of these private terminal modes being left enabled; if a future change drops
// one of these sequences, this test fails before the regression can ship.
func TestTerminalResetSequenceCompleteness(t *testing.T) {
	required := []struct {
		name string
		seq  string
	}{
		{"leave alternate screen", "\x1b[?1049l"},
		{"disable bracketed paste", "\x1b[?2004l"},
		{"disable X10/normal mouse tracking", "\x1b[?1000l"},
		{"disable button-event mouse tracking", "\x1b[?1002l"},
		{"disable any-event mouse tracking", "\x1b[?1003l"},
		{"disable SGR mouse coordinates", "\x1b[?1006l"},
		{"disable urxvt mouse coordinates", "\x1b[?1015l"},
		{"normal (non-application) cursor keys", "\x1b[?1l"},
		{"normal keypad", "\x1b>"},
		{"re-enable autowrap", "\x1b[?7h"},
		{"disable left/right margin mode", "\x1b[?69l"},
		{"reset scroll region", "\x1b[r"},
		{"show cursor", "\x1b[?25h"},
		{"reset character attributes", "\x1b[0m"},
		{"carriage return to column 0", "\r"},
	}
	for _, r := range required {
		if !strings.Contains(TerminalResetSequence, r.seq) {
			t.Errorf("TerminalResetSequence is missing the sequence to %s (%q)", r.name, r.seq)
		}
	}
}

// TestTerminalResetSequenceNonDestructive guards that restoring the terminal never
// wipes the user's screen or scrollback — only mode/attribute resets are allowed.
func TestTerminalResetSequenceNonDestructive(t *testing.T) {
	forbidden := []struct {
		name string
		seq  string
	}{
		{"full reset (RIS)", "\x1bc"},
		{"erase entire screen", "\x1b[2J"},
		{"erase scrollback", "\x1b[3J"},
	}
	for _, f := range forbidden {
		if strings.Contains(TerminalResetSequence, f.seq) {
			t.Errorf("TerminalResetSequence must not contain the destructive sequence %s (%q)", f.name, f.seq)
		}
	}
}

// TestRestoreTerminalStateWrites verifies the helper writes the full sequence to
// the provided writer (the exit cleanup relies on this to reach stdout).
func TestRestoreTerminalStateWrites(t *testing.T) {
	var buf bytes.Buffer
	RestoreTerminalState(&buf)
	if got := buf.String(); got != TerminalResetSequence {
		t.Errorf("RestoreTerminalState wrote %q, want %q", got, TerminalResetSequence)
	}
}

// TestRestoreTerminalStateEndsAtColumnZero ensures the sequence ends by returning
// the cursor to column 0 so the next shell prompt starts flush left (the visible
// "shifted prompt" symptom).
func TestRestoreTerminalStateEndsAtColumnZero(t *testing.T) {
	if !strings.HasSuffix(TerminalResetSequence, "\r") {
		t.Errorf("TerminalResetSequence must end with a carriage return so the prompt starts at column 0")
	}
}

// Resetting the scroll region moves the cursor to the top-left corner. Unless
// the cursor is saved before and restored after, the shell's prompt is written
// over the top of the screen, on top of what GoAnime printed last.
func TestRestoreTerminalStateLeavesTheCursorWhereItWas(t *testing.T) {
	seq := TerminalResetSequence
	save, region, restore := strings.Index(seq, "\x1b7"), strings.Index(seq, "\x1b[r"), strings.Index(seq, "\x1b8")
	if region < 0 {
		t.Fatal("the scroll region is no longer reset; update this test with the sequence")
	}
	if save < 0 || save > region {
		t.Errorf("the cursor must be saved (ESC 7) before the scroll region is reset (ESC [r) in %q", seq)
	}
	if restore < 0 || restore < region {
		t.Errorf("the cursor must be restored (ESC 8) after the scroll region is reset (ESC [r) in %q", seq)
	}
	if leave := strings.Index(seq, "\x1b[?1049l"); leave < save {
		t.Errorf("the cursor must be saved before the alternate screen is left: leaving it restores the cursor of when the last full-screen view opened, above what was printed since: %q", seq)
	}
	for _, home := range []string{"\x1b[H", "\x1b[1;1H", "\x1b[9999;1H"} {
		if strings.Contains(seq, home) {
			t.Errorf("the reset moves the cursor (%q): the prompt would land away from the last line printed", home)
		}
	}
}

// The exit must never take away what the run printed last: no erase of the
// screen, and nothing that scrolls it out of view.
func TestTheExitClearsNothing(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "cmd", "goanime", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, clear := range []string{"ClearViewport", "ClearScreen", `\x1b[2J`, `\033[2J`, `\x1bc`} {
		if strings.Contains(string(src), clear) {
			t.Errorf("cmd/goanime/main.go clears the screen (%s): --version, --help and the last error would be pushed out of sight", clear)
		}
	}
}
