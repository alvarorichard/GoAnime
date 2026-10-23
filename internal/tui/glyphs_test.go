package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cp1252Upper are the characters Windows decodes from cp1252's upper bytes
// (0x80-0x9F), which the latin-1 range check misses. Every Windows console
// font carries them, so the registration scan never asks for them.
var cp1252Upper = []rune{
	'€', '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', 'Ž',
	'‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', 'ž', 'Ÿ',
}

// wgl4Shapes are the geometric shapes a Windows console font always draws —
// the shapes and blocks of WGL4. Panels' shadows and lists' meters draw
// theirs directly, so the scan never asks for them.
var wgl4Shapes = map[rune]bool{
	'■': true, '□': true, '▪': true, '▫': true, '▲': true, '►': true,
	'▼': true, '◄': true, '◊': true, '○': true, '●': true, '◘': true,
	'◙': true, '◦': true, '▀': true, '▄': true, '█': true,
	'░': true, '▒': true, '▓': true,
}

// dataOnlyRunes are runes that string literals carry as parsed DATA — HTTP
// bodies, protocol markers — and never draw on a screen, so no console face
// needs to cover them.
var dataOnlyRunes = map[rune]bool{
	'\ufeff': true, // the UTF-8 BOM, trimmed from fetched bodies (player/download.go)
	// U+FFFD is what ranging over the invalid-UTF-8 byte strings yields — the
	// UTF-16 BOM "\xff\xfe" of vidsrc.go's body decoder; parsed data only.
	'\ufffd': true,
}

// needsRegistration answers whether a glyph in a screen string must sit in
// flavorRunes so the coverage probes can ask about it: outside Windows every
// rune these classes draw is renderable; anything else — the ❯ prompt, the
// ▶ cursor, the ✓ ◐ marks, the ⁖ asking mark, braille — may have no face on
// the console and so needs the probe and its ASCII fallback.
func needsRegistration(r rune) bool {
	switch {
	case r < 128:
		return false
	case isJapanese(r):
		return false // japanese_test.go's scan asks about these
	case r >= '¡' && r <= 'ÿ':
		return false
	}
	for _, known := range cp1252Upper {
		if r == known {
			return false
		}
	}
	if r >= '←' && r <= '↟' {
		return false // the basic arrows every console font carries
	}
	if r >= '─' && r <= '▓' {
		return false // box drawing, block elements and shades of WGL4
	}
	if wgl4Shapes[r] {
		return false
	}
	if dataOnlyRunes[r] {
		return false
	}
	return true
}

// scanUsedRunes walks dirs relative to this package and maps every
// needsRegistration rune found in a string literal of their non-test .go
// files to the files it sits in. testdata trees never count — nothing there
// draws a screen — and a dir that does not exist is skipped, like a root the
// module holds per-layout. A file that does not parse fails the run: real
// trees compile.
func scanUsedRunes(t testing.TB, dirs ...string) map[rune][]string {
	t.Helper()
	used := map[rune][]string{}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && d.Name() == "testdata" {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if parseErr != nil {
				return parseErr
			}
			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				value, unquoteErr := strconv.Unquote(lit.Value)
				if unquoteErr != nil {
					return true
				}
				for _, r := range value {
					if needsRegistration(r) {
						used[r] = append(used[r], path)
					}
				}
				return true
			})
			return nil
		})
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)
	}
	return used
}

// TestGlyphFontCoversEveryDecorativeGlyph fails when a screen draws a flavor
// glyph the probes do not know: they would keep it on where the console
// draws a box for it, and never hide it as the ASCII fallback. The same
// holds drawn from Go's own code as tests already pin (SetGlyphs(true)).
func TestGlyphFontCoversEveryDecorativeGlyph(t *testing.T) {
	t.Parallel()

	used := scanUsedRunes(t, "../../internal", "../../pkg", "../../cmd")
	require.NotEmpty(t, used, "the scan found the glyphs")
	require.Contains(t, used, '❯', "the scan reads the prompt")
	require.Contains(t, used, '▶', "the scan reads the cursor")

	registered := make(map[rune]bool, len(flavorRunes))
	for _, r := range flavorRunes {
		registered[r] = true
	}
	for r, files := range used {
		require.True(t, registered[r], "%q (in %s) is not in glyphs.go's flavorRunes: the console coverage probe cannot ask about it and a font without it draws a box; add the rune and its ASCII drawGlyph pair", r, files[0])
	}

	// The spinner's frames and the halftone art register themselves from
	// their sources; the flavor set answers everything it draws.
	for _, r := range spinnerDotFrames {
		require.True(t, registered[r], "spinner frame %q left out of flavorRunes", r)
	}
	for _, row := range brailleHalftone {
		for _, r := range row {
			require.True(t, registered[r], "halftone dot %q left out of flavorRunes", r)
		}
	}
}

// TestScreensWithoutDecorativeGlyphs renders every screen with the glyphs
// hidden, as a console whose fonts lack them receives them, and checks that
// not one flavor rune — prompt, cursor, marks, asking mark, panel rule,
// rating dot, braille — is left to draw as a box while the rest keeps its
// color. The same renders with the glyphs on must keep drawing some, so the
// check cannot pass on screens that never drew any.
func TestScreensWithoutDecorativeGlyphs(t *testing.T) {
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })
	t.Cleanup(func() { SetGlyphs(true) })

	theme := NewTheme(true)
	size := tea.WindowSizeMsg{Width: 120, Height: 31}
	view := func(m tea.Model) string {
		_, _ = m.Update(size)
		return m.View().Content
	}
	screens := map[string]func() string{
		"header": func() string {
			shell := NewShell(&theme, "Search > Results")
			return shell.header(120)
		},
		"search home": func() string {
			continueAll, recent := homeTestEntries()
			return view(newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent}))
		},
		"searching": func() string {
			plan, fire, _, _ := searchingTestEvents()
			m := newSearchingModel("namakura", nil)
			_, _ = m.Update(size)
			applyAll(m, plan, fire)
			width, _ := m.shell.ContentSize()
			return m.View().Content + m.panelRows(width-resultsListWidth(width))
		},
		"try":       func() string { return view(newTryModelWithOptions(tryTestOptions())) },
		"subtitles": func() string { return view(newSubtitlesModel(subtitleTestOptions())) },
		"results": func() string {
			return view(newAnimeResultsModel([]*models.Anime{{Name: "[PT-BR] Namakura Gatana", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime}}))
		},
		"picker with actions and a panel": func() string {
			opts := actionTestOptions()
			opts.Panel = &PickPanel{Label: "Now playing", Gloss: "再生中", Title: "Momotarou", Lines: []string{"Episode 1"}}
			return view(newPickerModel([]PickItem{{Label: "Ep 1"}, {Label: "Ep 2"}}, opts))
		},
		"transfer": func() string {
			return RenderTransfer(&Transfer{Label: "Download", Gloss: "ダウンロード", Title: "Namakura Gatana", Bar: NewTransferBar().ViewAs(0.5), Percent: 0.5}, 100)
		},
		"help": func() string {
			return RenderHelp(&theme, "Results", []HelpRow{{Key: "q", Desc: "quit"}}, nil, 100, 30)
		},
	}
	screenDrewGlyph := make(map[string]bool, len(screens))

	for name, render := range screens {
		SetGlyphs(true)
		withGlyphs := render()
		_, drewSomething := anyFlavorGlyph(ansi.Strip(withGlyphs))
		screenDrewGlyph[name] = drewSomething

		SetGlyphs(false)
		without := render()
		r, seen := anyFlavorGlyph(ansi.Strip(without))
		assert.False(t, seen, "%s left %q to draw as a box:\n%s", name, string(r), ansi.Strip(without))
		assert.NotEqual(t, ansi.Strip(without), without, "%s keeps its colors: only plain mode drops them", name)
	}

	// Where the terminals with fonts drew glyphs, at least one screen drew
	// one: the pin cannot stand on screens that never drew any.
	assert.True(t, func() bool {
		for _, drew := range screenDrewGlyph {
			if drew {
				return true
			}
		}
		return false
	}(), "with the glyphs on some screen draws one, proving the flip does something")
}

// anyFlavorGlyph finds a flavor rune inside s and tells the first one it
// saw, or nothing drawn.
func anyFlavorGlyph(s string) (rune, bool) {
	for _, r := range flavorRunes {
		if strings.ContainsRune(s, r) {
			return r, true
		}
	}
	return 0, false
}

// TestProbePatternMatchesProbeRunes keeps fontconfig's pattern and the
// Windows coverage probe asking for the same three runes.
func TestProbePatternMatchesProbeRunes(t *testing.T) {
	t.Parallel()

	var probed []rune
	for field := range strings.FieldsSeq(strings.TrimPrefix(japaneseFontPattern, ":charset=")) {
		code, err := strconv.ParseUint(field, 16, 32)
		require.NoError(t, err)
		probed = append(probed, rune(code))
	}
	assert.Equal(t, glossProbeRunes, probed, "the fontconfig pattern and glossProbeRunes ask different fonts")
}
