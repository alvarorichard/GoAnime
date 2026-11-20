package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// HandleClearRequest runs --clear-continue, --clear-recent and
// --clear-history: it empties the lists the request names in the history the
// SEARCH home screen reads, after asking.
func HandleClearRequest() error {
	return runClear(util.GlobalClearRequest, os.Stdin, os.Stdout, stdinIsTerminal())
}

// runClear is HandleClearRequest on any input and output. interactive says
// whether there is somebody to ask: without one the question cannot be
// answered, so nothing is cleared unless the request already said yes.
func runClear(req *util.ClearRequest, in io.Reader, out io.Writer, interactive bool) error {
	if req == nil || (!req.Continue && !req.Recent) {
		return errors.New("nothing to clear: use --clear-continue, --clear-recent or --clear-history")
	}
	what := describeClear(req)

	if !req.Yes {
		if !interactive {
			return fmt.Errorf("not clearing %s: there is no terminal to ask on; pass --yes to clear without asking", what)
		}
		_, _ = fmt.Fprintf(out, "Clear %s? This cannot be undone. [y/N] ", what)
		if !answeredYes(in) {
			_, _ = fmt.Fprintln(out, "Nothing was cleared.")
			return nil
		}
	}

	var failures []error
	if req.Continue {
		failures = appendClear(failures, out, "CONTINUE", "episode", "episodes", clearContinueFn)
	}
	if req.Recent {
		failures = appendClear(failures, out, "RECENT", "search", "searches", clearRecentFn)
	}
	return errors.Join(failures...)
}

// describeClear names what a request clears, for the question and the refusal.
func describeClear(req *util.ClearRequest) string {
	switch {
	case req.Continue && req.Recent:
		return "the CONTINUE list and the RECENT searches"
	case req.Continue:
		return "the CONTINUE list"
	default:
		return "the RECENT searches"
	}
}

// answeredYes reads one line from in and reports whether it is a yes. Anything
// else, and no answer at all, is a no.
func answeredYes(in io.Reader) bool {
	line, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes", "s", "sim":
		return true
	}
	return false
}

// appendClear runs one list's clear, tells the user what it did and adds the
// failure, if any, to failures.
func appendClear(failures []error, out io.Writer, list, one, many string, run func() (int, error)) []error {
	n, err := run()
	if err != nil {
		return append(failures, fmt.Errorf("clear %s: %w", list, err))
	}
	if n == 0 {
		_, _ = fmt.Fprintf(out, "%s was already empty.\n", list)
		return failures
	}
	noun := many
	if n == 1 {
		noun = one
	}
	_, _ = fmt.Fprintf(out, "Cleared %d %s from %s.\n", n, noun, list)
	return failures
}
