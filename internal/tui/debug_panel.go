package tui

import "github.com/charmbracelet/x/ansi"

// DebugPanel draws the notice shown at startup with --debug: where the log
// file is and how to follow it. It is the same paper panel the pickers use,
// with its crimson shadow and the label glossed in Japanese, so the debug
// screen no longer looks like the plain-text tool it used to be.
//
// width is the room the panel may take (shadow included). It returns "" when
// there is not enough, and the caller prints a plain line instead. That
// includes a command that would have to be cut: the user copies it, so it is
// shown whole or not at all.
func DebugPanel(logPath, followCmd string, width int) string {
	theme := CurrentTheme()
	return debugPanel(&theme, logPath, followCmd, width)
}

func debugPanel(theme *Theme, logPath, followCmd string, width int) string {
	const (
		shadow    = 1 // the column the shadow takes
		padding   = 4 // two cells each side
		codeStart = 3 // " $ " before the command
	)
	if width-shadow-padding < codeStart+ansi.StringWidth(followCmd) {
		return ""
	}
	return fitPanel(theme, &panelContent{
		label:   "debug",
		gloss:   "詳細", // details: the font has no バ or グ, so デバッグ would be boxes
		title:   "Debug log",
		fields:  []detailField{{label: "FILE", value: logPath}},
		lines:   []string{"Run this in another terminal to follow it live."},
		command: followCmd,
	}, width, 14)
}
