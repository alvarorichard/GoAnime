package tracking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleTrackingNotice_NoPanic(t *testing.T) {
	// HandleTrackingNotice prints to stdout when CGO is disabled.
	// We just verify it doesn't panic on either code path.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HandleTrackingNotice panicked: %v", r)
		}
	}()
	HandleTrackingNotice()
}

func TestHandleTrackingNotice_CgoDisabled(t *testing.T) {
	// Force IsCgoEnabled = false to exercise the print body of HandleTrackingNotice.
	// Cannot be parallel — modifies the package-level var.
	prev := IsCgoEnabled
	IsCgoEnabled = false
	t.Cleanup(func() { IsCgoEnabled = prev })

	assert.NotPanics(t, func() { HandleTrackingNotice() })
}

func TestHandleTrackingNotice_CgoEnabled(t *testing.T) {
	// IsCgoEnabled = true → function body is skipped (only the condition is evaluated).
	prev := IsCgoEnabled
	IsCgoEnabled = true
	t.Cleanup(func() { IsCgoEnabled = prev })

	assert.NotPanics(t, func() { HandleTrackingNotice() })
}
