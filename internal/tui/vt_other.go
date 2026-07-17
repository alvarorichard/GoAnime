//go:build !windows

package tui

import "os"

// EnableVirtualTerminal is a no-op outside Windows. Unix terminals interpret
// ANSI escape sequences by default. Always returns true.
func EnableVirtualTerminal() bool { return true }

// HasVirtualTerminal is always true outside Windows: hosts render ANSI natively.
func HasVirtualTerminal(f *os.File) bool {
	return f != nil
}
