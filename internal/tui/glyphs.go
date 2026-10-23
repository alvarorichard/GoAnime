package tui

import (
	"sync"
)

// The screens draw flavor glyphs apart from the Japanese glosses: the ❯
// prompt, the ▶ row cursor, the ✓ ◐ ● status marks, the ⁖ asking mark, the
// ━ panel rule, the • rating dot and the braille the paper panels' screen
// tone and the search spinner draw. A console font without them draws a box,
// as on a Windows Sandbox conhost, so every draw asks GlyphCovered first:
// drawGlyph answers with the ASCII fallback plain mode already drew there,
// and the probes (glyphs_windows.go) read the fonts the console really
// draws with. #215's glosses hide as whole strings; these single glyphs are
// hidden per rune, so one missing face never stocks a whole screen with
// boxes.

// flavorRunes are the single runes GoAnime draws outside the Japanese
// glosses. A glyph newly used in a screen string must be added here, or the
// probes cannot ask about it and TestGlyphFontCoversEveryDecorativeGlyph
// rejects the change.
var flavorRunes = []rune{'❯', '▶', '✓', '◐', '●', '○', '⁖', '━', '•'}

// spinnerDotFrames are huh spinner.Dots' bubbles frames. Its poster asks
// SpinnerDots and posts spinner.Line — ASCII every console font carries —
// when these have no face.
var spinnerDotFrames = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

func init() {
	// The halftone lives with the panels; its braille dots take part in the
	// same probe as the spinner's.
	flavorRunes = append(flavorRunes, spinnerDotFrames...)
	seen := make(map[rune]bool, len(flavorRunes))
	for _, r := range flavorRunes {
		seen[r] = true
	}
	for _, row := range brailleHalftone {
		for _, r := range row {
			if !seen[r] {
				seen[r] = true
				flavorRunes = append(flavorRunes, r)
			}
		}
	}
}

var (
	glyphMu       sync.RWMutex // guards glyphOverride
	glyphOverride *bool
	glyphOnce     sync.Once
	glyphCoverage map[rune]bool
)

// SetGlyphs forces the flavor glyphs on (true) or off (false) for the
// process, whatever the fonts. Tests pin it so renders do not depend on the
// fonts of the machine running them.
func SetGlyphs(show bool) {
	glyphMu.Lock()
	defer glyphMu.Unlock()
	glyphOverride = &show
}

func glyphPrioritized() bool {
	glyphMu.RLock()
	defer glyphMu.RUnlock()
	return glyphOverride != nil
}

// EnsureGlyphsDetected runs the coverage probe once. Startup calls it before
// the first screen, so the first frame does not pay for the probe.
func EnsureGlyphsDetected() {
	if glyphPrioritized() {
		return
	}
	glyphOnce.Do(func() { glyphCoverage = decorativeCoverage() })
}

// GlyphCovered reports whether the console drawing GoAnime has a font for
// the glyph r. A probe that cannot answer keeps the glyph on, like every
// probe that cannot answer — except classic consoles, which never fall back
// for a missing glyph and so degrade to ASCII.
func GlyphCovered(r rune) bool {
	glyphMu.RLock()
	override := glyphOverride
	glyphMu.RUnlock()
	if override != nil {
		return *override
	}
	glyphOnce.Do(func() { glyphCoverage = decorativeCoverage() })
	covered, known := glyphCoverage[r]
	return !known || covered
}

// drawGlyph returns fancy when the console has a font for its first rune and
// plain otherwise — exactly the fallback plain mode draws, so no console
// shows a box for a glyph GoAnime prints.
func drawGlyph(fancy, plain string) string {
	if Plain() {
		return plain
	}
	if first := []rune(fancy); len(first) > 0 && !GlyphCovered(first[0]) {
		return plain
	}
	return fancy
}

// meterRunes answers the fill and empty runes for the heavy-line bars: the
// heavy rule where the console has its font, and the ASCII pair plain mode
// draws otherwise — either font carries both or the ASCII pair wins, so no
// meter shows as a run of boxes.
func meterRunes() (fill, empty rune) {
	if Plain() || !GlyphCovered('━') {
		return '=', '-'
	}
	return '━', '━'
}

// SpinnerDots reports whether huh spinners may draw their braille frames:
// plain never draws them, and a console without a braille face receives the
// ASCII spinner.Line over them.
func SpinnerDots() bool {
	if Plain() {
		return false
	}
	for _, r := range spinnerDotFrames {
		if !GlyphCovered(r) {
			return false
		}
	}
	return true
}
