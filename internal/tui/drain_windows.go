//go:build windows

package tui

// drainStdin is a no-op on Windows.
// Windows terminals handle cursor key modes differently and the DECCKM
// issue only affects Unix-like terminals (xterm, iTerm2, GNOME Terminal, etc.).
func drainStdin() {}
