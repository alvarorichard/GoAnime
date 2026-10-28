//go:build !windows

package tui

import "os"

// EnableVirtualTerminal is a no-op outside Windows. Unix terminals interpret
// ANSI escape sequences by default. Always returns true.
func EnableVirtualTerminal() bool { return true }

// EnableUTF8Console is a no-op outside Windows: Unix terminals always read
// UTF-8. The restore func does nothing.
func EnableUTF8Console() func() { return func() {} }

// HasVirtualTerminal is always true outside Windows: hosts render ANSI natively.
func HasVirtualTerminal(f *os.File) bool {
	return f != nil
}
