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
const TerminalResetSequence = "" +
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
	"\x1b[?25h" + //   DECTCEM: show the cursor
	"\x1b[0m" + //     SGR 0: reset all character attributes/colors
	"\r" //            carriage return to column 0

// RestoreTerminalState writes TerminalResetSequence to w, returning sane
// interactive terminal state on program exit. It is safe to call multiple times
// and on any exit path.
func RestoreTerminalState(w io.Writer) {
	_, _ = io.WriteString(w, TerminalResetSequence)
}

// RestoreTerminalStdout restores the terminal on stdout. Convenience wrapper for
// the exit-cleanup path.
func RestoreTerminalStdout() {
	RestoreTerminalState(os.Stdout)
}
