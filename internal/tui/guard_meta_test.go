package tui

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the tests. The guards in glyphs_test.go keep the console
// typography honest only while their own mechanics stay faithful: the rune
// classes needsRegistration sorts, the way scanUsedRunes reads sources, and
// the pins TestMain sets for every render check. When a meta test here
// fails, a guard's verdict — silence or complaint — cannot be trusted until
// the mechanic is fixed.

// TestNeedsRegistrationClasses checks the sorting itself: every class the
// console draws without a probe stays out of the registry's demands, and
// everything else asks for a registered face and its ASCII fallback. The
// flavor marks sit deliberately on both sides — ● ○ ━ • draw in every
// listed class, ❯ ▶ ✓ ◐ ⁖ in none — the registry exists exactly for the
// second half, and the coverage probe asks the classic console about the
// first half anyway, since it never falls back per glyph.
func TestNeedsRegistrationClasses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		r    rune
		want bool
		why  string
	}{
		{'a', false, "ascii is safe everywhere"},
		{'~', false, "ascii punctuation too"},
		{'\x7f', false, "the last byte below 128"},
		{'¡', false, "latin-1"},
		{'ÿ', false, "latin-1's last rune"},
		{'…', false, "cp1252's ellipsis"},
		{'—', false, "the em dash windows consoles draw"},
		{'•', false, "cp1252's bullet"},
		{'ゴ', false, "a kana — japanese_test's own scan asks these"},
		{'検', false, "a kanji"},
		{'ー', false, "the long vowel mark"},
		{'←', false, "the basic arrows open"},
		{'↟', false, "the basic arrows close"},
		{'↠', true, "past the basics a face may lack the arrow"},
		{'─', false, "box drawing opens"},
		{'▓', false, "the block shades close"},
		{'◐', true, "outside WGL4 — flavorRunes registers it"},
		{'❯', true, "the prompt mark needs the probe"},
		{'⁖', true, "the asking mark needs the probe"},
		{'☎', true, "a symbol no listed class draws"},
		{'🚀', true, "an emoji"},
		{'\ufffd', false, "data-only: the replacement"},
	}
	for _, it := range cases {
		require.Equalf(t, it.want, needsRegistration(it.r), "%q (%U) — %s", it.r, it.r, it.why)
	}
	for _, known := range cp1252Upper {
		require.Falsef(t, needsRegistration(known), "%q must stay in cp1252's safe list", known)
	}
	for r := range wgl4Shapes {
		require.Falsef(t, needsRegistration(r), "%q must stay in the WGL4 safe list", r)
	}
	for r := range dataOnlyRunes {
		require.Falsef(t, needsRegistration(r), "%q must stay data-only", r)
	}
}

// TestFlavorRunesStayDistinct keeps the registered set free of duplicates:
// a repeat rune would read as two flavor marks while answering the probe
// once, and the scan would never notice on its own.
func TestFlavorRunesStayDistinct(t *testing.T) {
	t.Parallel()
	seen := make(map[rune]int)
	for _, r := range flavorRunes {
		seen[r]++
	}
	require.Len(t, seen, len(flavorRunes), "one mention per rune")
}

// scratchSource writes a .go file for scanUsedRunes to read; scratch trees
// live in a temp dir, so they never reach the real scan's roots.
func scratchSource(t *testing.T, path, src string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(src), 0o600))
}

// TestGlyphScanWalksSources feeds scanUsedRunes a synthetic tree: it must
// read string literals only, attach findings to the files that hold them,
// and skip test files, testdata trees and comments — the mechanics every
// guard verdict rides on.
func TestGlyphScanWalksSources(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// The prompt glyph drawn from a real literal.
	scratchSource(t, filepath.Join(root, "gloss", "tui.go"), `package gloss

// The prompt draws ❯ there — a comment, and comments never draw.
func Prompt() string {
	return "❯ "
}
`)
	// An unregistered symbol in a literal is what the guards must catch.
	scratchSource(t, filepath.Join(root, "utility", "dial.go"), `package utility

// ☎ in a comment reads as nothing.
func Dial() string {
	return "☎ reach"
}
`)
	// A test file never runs on a console, and testdata never draws one.
	scratchSource(t, filepath.Join(root, "utility", "dial_test.go"), `package utility

func TheTestString() string { return "☎ far from a console" }
`)
	scratchSource(t, filepath.Join(root, "fixture", "testdata", "buried.go"), `package buried

const Where = "☎ nowhere"
`)

	used := scanUsedRunes(t, root)
	require.Len(t, used, 2, "exactly the two literal-drawn runes showed up")
	require.Equal(t, []rune{'☎', '❯'}, runesIn(used), "no other rune reached the findings")

	promptFiles := used['❯']
	require.Len(t, promptFiles, 1, "one literal, one file")
	require.True(t, strings.HasSuffix(promptFiles[0], filepath.Join("gloss", "tui.go")),
		"the finding names the file that drew it")

	dialFiles := used['☎']
	require.Len(t, dialFiles, 1, "the comment, the test and the testdata stay out")
	require.True(t, strings.HasSuffix(dialFiles[0], filepath.Join("utility", "dial.go")),
		"the finding names the file that drew it")
}

// TestGlyphScanKeepsDataOnlyQuiet feeds the scan the byte-shaped BOMs real
// HTTP bodies carry: parsed data, never drawn — exemptions the scan must
// keep quiet about.
func TestGlyphScanKeepsDataOnlyQuiet(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	scratchSource(t, filepath.Join(root, "body", "bom.go"), "package body\n\n"+
		"// The UTF-8 and UTF-16 BOMs arrive as body bytes a stray parse strips.\n"+
		"const (\n"+
		"\tUtf8Bom  = \"\\xEF\\xBB\\xBF\"\n"+
		"\tUtf16Bom = \"\\xff\\xfe\"\n"+
		")\n")

	used := scanUsedRunes(t, root)
	assert.Empty(t, used, "the BOM strings stay out of the findings")
}

// TestTestMainPinsTheRenderPins guards the pins every render check leans
// on: the Japanese glosses and the flavor glyphs are forced on for the
// whole package, so screens compare by wording and not by the fonts of the
// machine running the tests. A sequential test runs exclusively — parallel
// ones pause until every sequential test of the package has finished — so
// the reads see a settled state.
func TestTestMainPinsTheRenderPins(t *testing.T) {
	assert.True(t, glyphPrioritized(), "TestMain sets the glyph override")
	assert.True(t, ShowJapanese(), "TestMain sets the japanese override")

	for _, r := range flavorRunes {
		require.Truef(t, GlyphCovered(r), "every flavor rune draws pinned-on: %q (%U)", r, r)
	}
}

// TestPackageSourcesStayGofmtClean is the CI gate landing locally: every
// source of this package formats clean through format.Source, the library
// gofmt itself uses. The compare normalizes the CRLF the Windows checkout
// adds, so it only complains about what the Linux CI's `gofmt -l` would.
func TestPackageSourcesStayGofmtClean(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "testdata" {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		require.NoErrorf(t, readErr, "%s: read", path)
		lf := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
		formatted, fmtErr := format.Source(lf)
		require.NoErrorf(t, fmtErr, "%s: parses as source", path)
		require.Equalf(t, string(lf), string(formatted), "%s: not gofmt-clean", path)
		return nil
	})
	require.NoError(t, err)
}

// runesIn lists a used map's keys sorted, so assertions compare a stable
// order instead of Go's deliberately random map walk.
func runesIn(used map[rune][]string) []rune {
	list := make([]rune, 0, len(used))
	for r := range used {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
	return list
}

// TestEveryRealLaunchAsksAboutTheFonts pins the install guarantee's entry
// point: cmd/goanime/main.go calls tui.EnsureJapaneseDetected once the flags
// are parsed — after help/version already returned, before the search home
// or a named search can paint glosses, and out of any handler's hands. The
// decision layer may not depend on a particular handler remembering
// InitLogger; remove the startup call and a font-less machine pins its
// install on whichever path happens to open first.
func TestEveryRealLaunchAsksAboutTheFonts(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../cmd/goanime/main.go", nil, parser.SkipObjectResolution)
	require.NoError(t, err, "cmd/goanime/main.go must parse")

	var parsePos, ensurePos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch {
		case x.Name == "util" && sel.Sel.Name == "FlagParser":
			parsePos = n.Pos()
		case x.Name == "tui" && sel.Sel.Name == "EnsureJapaneseDetected":
			ensurePos = n.Pos()
		}
		return true
	})
	require.NotZero(t, parsePos, "the flag parse call must exist in main.go")
	require.NotZero(t, ensurePos, "startup must call tui.EnsureJapaneseDetected from main.go")
	assert.True(t, ensurePos > parsePos,
		"the font decision runs after flag parsing, so help and version stay silent about it")
}
