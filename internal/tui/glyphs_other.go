//go:build !windows

package tui

// decorativeCoverage has nothing to probe outside Windows: macOS ships
// Hiragino and every other system's terminals fall back over the installed
// fonts, so none of the flavor glyphs needs hiding. A nil map means every
// rune answered.
func decorativeCoverage() map[rune]bool {
	return nil
}
