//go:build !windows

package tui

// RelaunchInWindowsTerminal is a Windows-only bootstrap. Other systems start
// GoAnime in the terminal the user already chose, so there is nothing to
// relaunch.
func RelaunchInWindowsTerminal() bool { return false }
