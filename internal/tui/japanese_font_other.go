//go:build !linux

package tui

// glyphFontState counts any font fontconfig finds: GoAnime installs its own
// only on Linux, so none of them is GoAnime Glyphs.
func glyphFontState([]string) fontState { return fontPresent }

// installGlyphFont does nothing outside Linux. Windows Terminal and macOS
// already have Japanese fonts, and the classic Windows console draws with
// its one font whatever is installed.
func installGlyphFont() string { return "" }
