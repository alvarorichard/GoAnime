// Package tui provides terminal UI helpers that wrap go-fuzzyfinder with
// proper terminal state management.
//
// tcell (used by go-fuzzyfinder) enables "application cursor key mode"
// (DECCKM, ESC[?1h) during initialization.  Its Fini() is supposed to
// restore the original mode, but in practice the terminal is sometimes left
// in application mode.  When a second fuzzyfinder instance (or any other
// readline-style UI) runs afterwards it receives SS3-encoded arrow keys
// (ESC O A) instead of the expected CSI sequences (ESC [ A) and prints
// them as raw text.
//
// The Find wrapper in this package explicitly resets the terminal to normal
// cursor key mode after every fuzzyfinder call and drains any stale bytes
// from stdin so that subsequent interactive prompts work correctly.
package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
)

// ResetTerminal sends ANSI sequences to reset terminal state after tcell
// and drains any stale bytes from stdin that tcell may have left behind.
func ResetTerminal() {
	// Reset DECCKM (normal cursor keys) + reset keypad numeric mode + show cursor
	// These match the exact sequences tcell's ExitKeypad should send but
	// sometimes fails to:
	//   \033[?1l  — reset DECCKM (normal cursor keys)
	//   \033>     — numeric keypad mode
	//   \033[?25h — show cursor
	fmt.Fprint(os.Stdout, "\033[?1l\033>\033[?25h")

	// Drain any stale bytes from stdin (platform-specific implementation). A
	// short raw/no-echo window also catches late terminal capability responses
	// such as OSC 11 and DECRQM ?2026/?2027 before they can be echoed.
	DrainTerminalResponses(50 * time.Millisecond)
}

// Find is a drop-in replacement for fuzzyfinder.Find that resets the
// terminal's cursor key mode after the finder exits.
func Find[T any](slice []T, itemFunc func(i int) string, opts ...fuzzyfinder.Option) (int, error) {
	idx, err := fuzzyfinder.Find(slice, itemFunc, opts...)
	ResetTerminal()
	return idx, err
}
