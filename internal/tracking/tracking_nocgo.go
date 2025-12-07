//go:build !cgo

package tracking

import (
	"fmt"
)

// NoopTracker is a tracker that does nothing but logs messages if debug is enabled
// Used when CGO is disabled and SQLite is unavailable
type NoopTracker struct{}

// NewNoopTracker creates a new NoopTracker that satisfies the LocalTracker interface
func NewNoopTracker() *LocalTracker {
	fmt.Println("Notice: Anime progress tracking disabled (CGO support not available)")
	return nil
}

// init function for build without CGO
func init() {
	// Override NewLocalTracker to return NoopTracker when CGO is disabled
	originalNewLocalTracker := NewLocalTracker
	NewLocalTracker = func(dbPath string) *LocalTracker {
		return NewNoopTracker()
	}

	// Enable local_test.go to still use the real implementation during tests
	// by keeping a reference to the original function
	_ = originalNewLocalTracker

	// Set the global flag that CGO is disabled
	IsCgoEnabled = false
}
