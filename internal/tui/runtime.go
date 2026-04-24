// Package tui wraps fuzzyfinder and form utilities for interactive terminal prompts.
package tui

import (
	"fmt"
	"math"
	"os"
	"sync"

	"charm.land/huh/v2/spinner"
	"github.com/ktr0731/go-fuzzyfinder"
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

// RunWithSpinner executes the action with a spinner in interactive terminals
// and falls back to a plain call in non-interactive environments.
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

// ChooseIndex returns the selected index for an interactive list and
// auto-selects the first item when there is only one option or no TTY.
func ChooseIndex[T any](items []T, render func(int) string, prompt string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items available")
	}
	if len(items) == 1 || !isStdoutTerminal() {
		return 0, nil
	}

	return Find(items, render, fuzzyfinder.WithPromptString(prompt))
}
