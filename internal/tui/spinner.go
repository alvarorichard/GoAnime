// Package tui provides utility functions like spinners and UI helpers.
package tui

import (
	"math"
	"os"
	"sync"

	"charm.land/huh/v2/spinner"
	"golang.org/x/term"
)

var (
	stdoutIsTerminal     bool
	stdoutIsTerminalOnce sync.Once
)

func isStdoutTerminal() bool {
	stdoutIsTerminalOnce.Do(func() {
		fd := os.Stdout.Fd()
		stdoutIsTerminal = fd <= math.MaxInt && term.IsTerminal(int(fd))
	})
	return stdoutIsTerminal
}

// RunWithSpinner runs the action with a spinner if stdout is a terminal.
// In non-interactive environments it executes the action directly.
func RunWithSpinner(title string, action func()) {
	if isStdoutTerminal() {
		_ = spinner.New().
			Title(title).
			Type(spinner.Dots).
			Action(action).
			Run()
		return
	}

	action()
}
