package appflow

import (
	"sync/atomic"
	"testing"
)

// TestDefaultRunSpinner_RunsActionWithoutTTY guards the no-terminal path: the
// spinner is decoration, the action is the work. Under a test binary stdin is
// not a terminal, so huh's spinner returns without ever invoking its Action —
// the work must still happen, and exactly once (defaultRunSpinner's sync.Once
// makes the spinner path and the direct-call fallback mutually exclusive).
func TestDefaultRunSpinner_RunsActionWithoutTTY(t *testing.T) {
	var calls atomic.Int32
	defaultRunSpinner("working", func() { calls.Add(1) })

	if got := calls.Load(); got != 1 {
		t.Fatalf("action ran %d times, want exactly 1 "+
			"(0 = work silently skipped without a terminal; 2 = ran twice)", got)
	}
}
