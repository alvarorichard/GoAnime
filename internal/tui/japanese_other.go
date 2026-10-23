//go:build !windows

package tui

// classicConsole is a Windows question: there is no conhost elsewhere.
func classicConsole() bool { return false }

// cjkCodePage is a Windows question: other systems have no ANSI code page.
func cjkCodePage() bool { return false }
