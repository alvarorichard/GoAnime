//go:build windows

package tui

import (
	"os"

	"golang.org/x/sys/windows"
)

// EnableVirtualTerminal turns on ANSI/VT processing for stdout and stderr.
//
// Classic cmd.exe on Windows 10 leaves ENABLE_VIRTUAL_TERMINAL_PROCESSING off.
// Without it, TrueColor/ANSI sequences from lipgloss/log/tcell print as raw
// garbage (e.g. ←[38;2;255;255;255m). Returns true when at least one of
// stdout/stderr has VT processing active after the call.
func EnableVirtualTerminal() bool {
	outOK := enableVT(os.Stdout)
	errOK := enableVT(os.Stderr)
	return outOK || errOK
}

// HasVirtualTerminal reports whether f is a console with VT processing on.
func HasVirtualTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return consoleHasVT(windows.Handle(f.Fd()))
}

func enableVT(f *os.File) bool {
	if f == nil {
		return false
	}
	handle := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	if err := windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return false
	}
	return consoleHasVT(handle)
}

func consoleHasVT(handle windows.Handle) bool {
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	return mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0
}
