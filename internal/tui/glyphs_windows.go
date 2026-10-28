//go:build windows

package tui

import "slices"

// decorativeCoverage asks the fonts the console really draws with about the
// flavor runes, answering rune by rune. A classic console draws with its one
// chosen face and never falls back for a missing glyph, so its own answer is
// final and a probe that fails hides everything rather than risk boxes. A
// ConPTY host falls back over every face installed, so uncovered understand
// as "no face on this machine has it".
func decorativeCoverage() map[rune]bool {
	asked := make(map[rune]bool, len(flavorRunes))
	for _, r := range flavorRunes {
		asked[r] = false
	}
	dc, done := screenDC()
	if dc == 0 {
		return asked
	}
	defer done()
	if classicConsole() {
		face := consoleFontFace()
		if face == "" {
			return asked
		}
		drawn, ok := gdiFaceGlyphs(dc, face, flavorRunes)
		if !ok {
			return asked
		}
		for r, yes := range drawn {
			asked[r] = yes
		}
		return asked
	}
	faces, ok := enumFontFaces()
	if !ok {
		return asked
	}
	remaining := slices.Clone(flavorRunes)
	for _, face := range faces {
		if len(remaining) == 0 {
			break
		}
		drawn, ok := gdiFaceGlyphs(dc, face, remaining)
		if !ok {
			continue
		}
		still := remaining[:0]
		for _, r := range remaining {
			if drawn[r] {
				asked[r] = true
			} else {
				still = append(still, r)
			}
		}
		remaining = still
	}
	return asked
}
