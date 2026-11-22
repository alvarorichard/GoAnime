package tui

import (
	"io"
	"os"
)

// TerminalResetSequence is the full set of ANSI/DEC sequences that return a
// terminal to a sane interactive state on program exit.
//
// Interactive UIs in this app (Bubble Tea v2 progress bars, the huh spinner,
// tcell-backed fuzzyfinder, readline prompts) enable private terminal modes —
// the alternate screen, mouse tracking, bracketed paste, a scroll region, left/
// right margins, application cursor keys — and hide the cursor. Their normal
// teardown restores these, but an abnormal exit (SIGINT, a crash, the process
// dying while a TUI is mid-render, or playback being quit) skips it and leaves
// the mode set. The visible result is a broken shell prompt afterwards (shifted
// to the right from a leftover scroll region/margin, invisible cursor, raw mouse
// bytes, or garbled colors).
//
// Emitting all of these unconditionally is safe: re-disabling a mode that is
// already off is a no-op, and NONE of these clear the screen or scrollback (no
// RIS \033c, no ED \033[2J/\033[3J) — the user's history is preserved.
//
// Order: leave the alternate screen first (so the rest applies to the primary
// buffer), reset modes, then show the cursor / reset colors / return to column 0.
//
// The cursor stays on the line it is on, which two of these would move:
//   - Leaving the alternate screen restores the cursor saved when it was
//     entered, and it does so even when the program left it long ago: the
//     cursor went back to where the last full-screen view had opened, above
//     everything printed since ("Search cancelled.", an error).
//   - Resetting the scroll region (DECSTBM) moves the cursor to the top-left
//     corner.
//
// Either way the shell's prompt was written over what GoAnime printed last.
// That once passed for TUI leftovers and was hidden by clearing the screen at
// every exit, which in turn pushed every final message (--version, --help, an
// error) out of sight. So the cursor is saved first, which also makes it the
// one that leaving the alternate screen restores, and restored at the end.
//
// A process killed while a full-screen view was still up (rare: the views
// leave the alternate screen themselves, on a panic too) gets the cursor of
// that view's screen; the prompt then lands wherever that was.
const TerminalResetSequence = "" +
	"\x1b7" + //       DECSC: save the cursor; leaving the alternate screen restores this one
	"\x1b[?1049l" + // DECRST 1049: leave alternate screen buffer
	"\x1b[?2004l" + // disable bracketed paste
	"\x1b[?1000l" + // disable X10/normal mouse tracking
	"\x1b[?1002l" + // disable button-event mouse tracking
	"\x1b[?1003l" + // disable any-event mouse tracking
	"\x1b[?1006l" + // disable SGR mouse extended coordinates
	"\x1b[?1015l" + // disable urxvt mouse extended coordinates
	"\x1b[?1l" + //    DECCKM: normal (non-application) cursor keys
	"\x1b>" + //       DECKPNM: normal keypad
	"\x1b[?7h" + //    DECAWM: re-enable autowrap
	"\x1b[?69l" + //   DECLRMM: disable left/right margin mode (clears any side margins)
	"\x1b[r" + //      DECSTBM: reset the scroll region to the full screen
	"\x1b8" + //       DECRC: put the cursor back where it was
	"\x1b[?25h" + //   DECTCEM: show the cursor
	"\x1b[0m" + //     SGR 0: reset all character attributes/colors
	"\r" //            carriage return to column 0

// RestoreTerminalState writes TerminalResetSequence to w, returning sane
// interactive terminal state on program exit. It is safe to call multiple times
// and on any exit path.
//
// On Windows without VT, the sequence is skipped — emitting it to classic
// cmd.exe prints raw escape garbage instead of restoring the console.
func RestoreTerminalState(w io.Writer) {
	if f, ok := w.(*os.File); ok && !SupportsANSI(f) {
		return
	}
	_, _ = io.WriteString(w, TerminalResetSequence)
}

// RestoreTerminalStdout restores the terminal on stdout. Convenience wrapper for
// the exit-cleanup path.
func RestoreTerminalStdout() {
	RestoreTerminalState(os.Stdout)
}
