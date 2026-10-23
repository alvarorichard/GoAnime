//go:build windows

package tui

import "golang.org/x/sys/windows"

var procGetConsoleWindow = modkernel32.NewProc("GetConsoleWindow")

// classicConsole reports whether the console window is conhost's own. Under a
// ConPTY host (Windows Terminal, VS Code, WezTerm) it is a hidden
// "PseudoConsoleWindow" instead; no window at all means pipes or mintty.
func classicConsole() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	var class [64]uint16
	n, err := windows.GetClassName(windows.HWND(hwnd), &class[0], int32(len(class)))
	if err != nil || n <= 0 {
		return false
	}
	return windows.UTF16ToString(class[:n]) == "ConsoleWindowClass"
}

// cjkCodePage reports whether Windows runs the Japanese (932), Simplified
// Chinese (936), Korean (949) or Traditional Chinese (950) code page. The
// classic console's default font then carries kana and kanji.
func cjkCodePage() bool {
	switch windows.GetACP() {
	case 932, 936, 949, 950:
		return true
	}
	return false
}

// probeConsoleFontCovers asks gdi whether the classic console's one chosen
// font maps every rune of runes — the exact coverage conhost draws with,
// instead of a guess from the code page. A lost answer keeps ascetics: no
// gloss, so a misread font never prints boxes.
func probeConsoleFontCovers(runes []rune) bool {
	face := consoleFontFace()
	if face == "" {
		return false
	}
	dc, done := screenDC()
	if dc == 0 {
		return false
	}
	defer done()
	drawn, ok := gdiFaceGlyphs(dc, face, runes)
	if !ok {
		return false
	}
	return every(drawn)
}
