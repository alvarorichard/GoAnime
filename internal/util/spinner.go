package util

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

// RunWithSpinner runs the action with a spinner if stdout is a terminal,
// otherwise runs the action directly. This ensures CI and non-interactive
// environments work correctly since huh/v2 spinner may skip the Action
// callback when no terminal is attached.
func RunWithSpinner(title string, action func()) {
	if isStdoutTerminal() {
		_ = spinner.New().
			Title(title).
			Type(spinner.Dots).
			Action(action).
			Run()
	} else {
		action()
	}
}
