package tracking

import (
	"testing"
)

func TestHandleTrackingNotice_NoPanic(t *testing.T) {
	// HandleTrackingNotice prints to stdout when CGO is disabled.
	// We just verify it doesn't panic on either code path.
	assert := func(condition bool, msg string) {
		t.Helper()
		if !condition {
			t.Fatal(msg)
		}
	}
	_ = assert

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HandleTrackingNotice panicked: %v", r)
		}
	}()
	HandleTrackingNotice()
}
