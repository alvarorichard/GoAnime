//go:build !linux && !windows

package tui

// japaneseFontProbe asks fc-list, the same fontconfig question Linux asks.
func japaneseFontProbe() fontState { return fontconfigState() }

// glyphFontState counts any font fontconfig finds: GoAnime installs its own
// only on Linux, so none of them is GoAnime Glyphs.
func glyphFontState([]string) fontState { return fontPresent }

// installGlyphFont does nothing outside Linux and Windows. macOS already has
// Japanese fonts, and the classic Windows console draws with its one font
// whatever is installed — Windows installs GoAnime Glyphs instead.
func installGlyphFont() string { return "" }
