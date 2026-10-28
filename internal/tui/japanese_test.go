package tui

import (
	"encoding/binary"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain pins the Japanese glosses and the flavor glyphs on: renders must
// not depend on the fonts of the machine running the tests. Tests about
// hiding them switch them off and back on.
func TestMain(m *testing.M) {
	SetJapanese(true)
	SetGlyphs(true)
	os.Exit(m.Run())
}

// isJapanese reports whether r is a kana, a kanji or the long vowel mark.
func isJapanese(r rune) bool {
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) || r == 'ー'
}

func japaneseIn(s string) string {
	var out []rune
	for _, r := range s {
		if isJapanese(r) {
			out = append(out, r)
		}
	}
	return string(out)
}

// fakeMachine answers the probes and counts what resolveJapanese asked.
type fakeMachine struct {
	goos      string
	classic   bool
	cjk       bool
	covers    bool // the classic console font's coverage of the glosses
	fonts     fontState
	asked     int
	installed int
}

func (f *fakeMachine) probes() *japaneseProbes {
	return &japaneseProbes{
		goos:              f.goos,
		classicConsole:    func() bool { return f.classic },
		cjkCodePage:       func() bool { return f.cjk },
		consoleFontCovers: func([]rune) bool { return f.covers },
		fonts: func() fontState {
			f.asked++
			return f.fonts
		},
		installFont: func() { f.installed++ },
	}
}

func TestResolveJapanese(t *testing.T) {
	t.Parallel()

	desktop := []string{"DISPLAY=:0", "TERM=xterm-256color"}
	wayland := []string{"WAYLAND_DISPLAY=wayland-0", "TERM=xterm-kitty"}
	with := func(env []string, extra ...string) []string {
		return append(slices.Clone(env), extra...)
	}

	tests := []struct {
		name          string
		env           []string
		machine       fakeMachine
		want          bool
		wantAsked     bool // fontconfig was asked
		wantInstalled bool
	}{
		// The #215 recording: an Arch Linux desktop with no Japanese font.
		{name: "linux desktop without a Japanese font", env: wayland, machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: false, wantAsked: true, wantInstalled: true},
		{name: "linux desktop with a Japanese font", env: desktop, machine: fakeMachine{goos: "linux", fonts: fontPresent}, want: true, wantAsked: true},
		{name: "font installed after the terminal started", env: desktop, machine: fakeMachine{goos: "linux", fonts: fontPending}, want: false, wantAsked: true},
		{name: "no fontconfig to ask", env: desktop, machine: fakeMachine{goos: "linux", fonts: fontUnknown}, want: true, wantAsked: true},
		{name: "freebsd desktop without a Japanese font", env: desktop, machine: fakeMachine{goos: "freebsd", fonts: fontMissing}, want: false, wantAsked: true, wantInstalled: true},

		{name: "GOANIME_JAPANESE=1 wins over a missing font", env: with(desktop, "GOANIME_JAPANESE=1"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "GOANIME_JAPANESE=on", env: with(desktop, "GOANIME_JAPANESE= On "), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "GOANIME_JAPANESE=0 hides and installs nothing", env: with(desktop, "GOANIME_JAPANESE=0"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: false},
		{name: "GOANIME_JAPANESE=off on macOS", env: []string{"GOANIME_JAPANESE=off"}, machine: fakeMachine{goos: "darwin"}, want: false},
		{name: "GOANIME_JAPANESE=auto detects", env: with(desktop, "GOANIME_JAPANESE=auto"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: false, wantAsked: true, wantInstalled: true},

		{name: "linux virtual console", env: []string{"TERM=linux"}, machine: fakeMachine{goos: "linux", fonts: fontPresent}, want: false},
		{name: "linux virtual console over ssh", env: []string{"TERM=linux", "SSH_CONNECTION=10.0.0.2 5000 10.0.0.1 22"}, machine: fakeMachine{goos: "linux"}, want: false},

		{name: "macOS", env: []string{"TERM=xterm-256color"}, machine: fakeMachine{goos: "darwin"}, want: true},
		{name: "android", env: nil, machine: fakeMachine{goos: "android"}, want: true},

		{name: "windows terminal", env: []string{"WT_SESSION=1"}, machine: fakeMachine{goos: "windows"}, want: true, wantAsked: true, wantInstalled: true},
		// A Windows Sandbox or any en-US machine ships no Japanese font at
		// all for its terminals to fall back on: the probe installs GoAnime
		// Glyphs and hides the glosses until the terminal is reopened.
		{name: "windows terminal without a Japanese font", env: []string{"WT_SESSION=1"}, machine: fakeMachine{goos: "windows", fonts: fontMissing}, want: false, wantAsked: true, wantInstalled: true},
		{name: "windows terminal whose only Japanese font got installed now", env: []string{"WT_SESSION=1"}, machine: fakeMachine{goos: "windows", fonts: fontPending}, want: false, wantAsked: true},
		{name: "windows terminal whose probe cannot answer still gets the glyphs", env: []string{"WT_SESSION=1"}, machine: fakeMachine{goos: "windows", fonts: fontUnknown}, want: true, wantAsked: true, wantInstalled: true},
		{name: "classic windows console without coverage asks the fonts", env: nil, machine: fakeMachine{goos: "windows", classic: true, fonts: fontUnknown}, want: false, wantAsked: true, wantInstalled: true},
		// Classic consoles draw glosses only with the font the console itself
		// has selected — no fallback — so a machine font the console ignores
		// is not enough. With no Japanese font anywhere, the glyphs font
		// installs for the console's own font list.
		{name: "classic console with no Japanese font installs the glyphs font", env: nil, machine: fakeMachine{goos: "windows", classic: true, fonts: fontMissing}, want: false, wantAsked: true, wantInstalled: true},
		{name: "classic console over a machine font that cannot help", env: nil, machine: fakeMachine{goos: "windows", classic: true, fonts: fontPending}, want: false, wantAsked: true},
		{name: "classic console whose font covers the glosses", env: nil, machine: fakeMachine{goos: "windows", classic: true, covers: true}, want: true},
		{name: "classic console on a japanese windows", env: nil, machine: fakeMachine{goos: "windows", classic: true, cjk: true}, want: true},

		{name: "over ssh the client draws", env: with(desktop, "SSH_CONNECTION=10.0.0.2 5000 10.0.0.1 22"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "ssh tty", env: with(desktop, "SSH_TTY=/dev/pts/3"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "under WSL windows draws", env: with(desktop, "WSL_DISTRO_NAME=Ubuntu"), machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "no local display", env: []string{"TERM=xterm-256color"}, machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
		{name: "empty display", env: []string{"DISPLAY=", "WAYLAND_DISPLAY= "}, machine: fakeMachine{goos: "linux", fonts: fontMissing}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			machine := tt.machine
			got := resolveJapanese(tt.env, machine.probes())
			assert.Equal(t, tt.want, got, "shown")
			assert.Equal(t, tt.wantAsked, machine.asked > 0, "fontconfig asked")
			assert.LessOrEqual(t, machine.asked, 1, "fontconfig asked once")
			assert.Equal(t, tt.wantInstalled, machine.installed > 0, "font installed")
			assert.LessOrEqual(t, machine.installed, 1, "font installed once")
		})
	}
}

// writeFakeFcList puts an fc-list on PATH that records its arguments and
// prints out, or fails when out is "fail".
func writeFakeFcList(t *testing.T, out string) (argsFile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake fc-list is a shell script")
	}
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "args")
	script := "#!/bin/sh\nfor a in \"$@\"; do printf '%s\\n' \"$a\"; done > '" + argsFile + "'\n"
	if out == "fail" {
		script += "exit 1\n"
	} else {
		script += "printf '%s' '" + out + "'\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fc-list"), []byte(script), 0o700))
	t.Setenv("PATH", dir)
	return argsFile
}

func TestFontconfigState(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	t.Run("a font has the glyphs", func(t *testing.T) {
		argsFile := writeFakeFcList(t, "/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc\n/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc\n")
		assert.Equal(t, fontPresent, fontconfigState())
		args, err := os.ReadFile(argsFile)
		require.NoError(t, err)
		assert.Equal(t, "--format=%{file}\\n\n:charset=30b4 30fc 691c\n", string(args))
	})
	t.Run("no font has them", func(t *testing.T) {
		writeFakeFcList(t, "")
		assert.Equal(t, fontMissing, fontconfigState())
	})
	t.Run("only blank lines", func(t *testing.T) {
		writeFakeFcList(t, "\n \n")
		assert.Equal(t, fontMissing, fontconfigState())
	})
	t.Run("fc-list fails", func(t *testing.T) {
		writeFakeFcList(t, "fail")
		assert.Equal(t, fontUnknown, fontconfigState())
	})
	t.Run("no fc-list", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		assert.Equal(t, fontUnknown, fontconfigState())
	})
}

// TestFontconfigStateRealFontconfig runs the real fc-list against a
// configuration holding only fonts/, so the pattern is checked against
// fontconfig's own parser and the font GoAnime installs.
func TestFontconfigStateRealFontconfig(t *testing.T) {
	if _, err := exec.LookPath("fc-list"); err != nil {
		t.Skip("fontconfig is not installed")
	}
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	fontsDir, err := filepath.Abs("fonts")
	require.NoError(t, err)

	useConfig := func(t *testing.T, dirs ...string) {
		t.Helper()
		cfg := filepath.Join(t.TempDir(), "fonts.conf")
		xml := "<?xml version=\"1.0\"?>\n<fontconfig>\n"
		for _, dir := range dirs {
			xml += "  <dir>" + dir + "</dir>\n"
		}
		xml += "  <cachedir>" + filepath.Join(t.TempDir(), "cache") + "</cachedir>\n</fontconfig>\n"
		require.NoError(t, os.WriteFile(cfg, []byte(xml), 0o600))
		t.Setenv("FONTCONFIG_FILE", cfg)
	}

	t.Run("no fonts at all", func(t *testing.T) {
		useConfig(t)
		assert.Equal(t, fontMissing, fontconfigState())
	})
	t.Run("GoAnime Glyphs answers the probe", func(t *testing.T) {
		useConfig(t, fontsDir)
		assert.Equal(t, fontPresent, fontconfigState(), "a font outside GoAnime's install folder counts as present")
	})
}

// TestGlyphFontCoversEveryGloss fails when a Japanese string is added
// without its characters in GoAnime Glyphs: on a machine that relies on
// that font they would be boxes. fonts/build.py has the steps.
func TestGlyphFontCoversEveryGloss(t *testing.T) {
	t.Parallel()

	used := map[rune][]string{}
	for _, dir := range []string{"../../internal", "../../pkg", "../../cmd"} {
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
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					return true
				}
				for _, r := range japaneseIn(value) {
					used[r] = append(used[r], path)
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
	require.NotEmpty(t, used, "the scan found the glosses")
	require.Contains(t, used, 'ゴ', "the scan reads the brand gloss")

	glyphs, err := os.ReadFile("fonts/glyphs.txt")
	require.NoError(t, err)
	font, err := os.ReadFile("fonts/GoAnimeGlyphs.ttf")
	require.NoError(t, err)
	cmap := readCmap(t, font)

	for r, files := range used {
		assert.Contains(t, string(glyphs), string(r), "%q (in %s) is missing from fonts/glyphs.txt; add it and run fonts/build.py", r, files[0])
		assert.Contains(t, cmap, r, "%q (in %s) is missing from GoAnimeGlyphs.ttf; run fonts/build.py", r, files[0])
	}
	for field := range strings.FieldsSeq(strings.TrimPrefix(japaneseFontPattern, ":charset=")) {
		code, err := strconv.ParseUint(field, 16, 32)
		require.NoError(t, err)
		assert.Contains(t, cmap, rune(code), "the fontconfig probe asks for %q, so GoAnime Glyphs must answer it", rune(code))
	}
}

// readCmap returns the characters of an OpenType font's Windows Unicode BMP
// cmap (platform 3, encoding 1, format 4).
func readCmap(t *testing.T, font []byte) map[rune]bool {
	t.Helper()
	u16 := func(b []byte, at int) int { return int(binary.BigEndian.Uint16(b[at:])) }
	u32 := func(b []byte, at int) int { return int(binary.BigEndian.Uint32(b[at:])) }

	var cmap []byte
	for i, n := 0, u16(font, 4); i < n; i++ {
		record := 12 + 16*i
		if string(font[record:record+4]) == "cmap" {
			offset, length := u32(font, record+8), u32(font, record+12)
			cmap = font[offset : offset+length]
		}
	}
	require.NotNil(t, cmap, "the font has a cmap table")

	for i, n := 0, u16(cmap, 2); i < n; i++ {
		record := 4 + 8*i
		if u16(cmap, record) != 3 || u16(cmap, record+2) != 1 {
			continue
		}
		sub := cmap[u32(cmap, record+4):]
		require.Equal(t, 4, u16(sub, 0), "format 4 subtable")
		segments := u16(sub, 6) / 2
		ends, starts := 14, 16+2*segments
		deltas, ranges := starts+2*segments, starts+4*segments
		chars := map[rune]bool{}
		for s := range segments {
			end, start := u16(sub, ends+2*s), u16(sub, starts+2*s)
			delta, rangeOffset := u16(sub, deltas+2*s), u16(sub, ranges+2*s)
			for c := start; c <= end && c != 0xFFFF; c++ {
				glyph := (c + delta) & 0xFFFF
				if rangeOffset != 0 {
					glyph = u16(sub, ranges+2*s+rangeOffset+2*(c-start))
					if glyph != 0 {
						glyph = (glyph + delta) & 0xFFFF
					}
				}
				if glyph != 0 {
					chars[rune(c)] = true
				}
			}
		}
		return chars
	}
	t.Fatal("the font has no Windows Unicode BMP cmap")
	return nil
}

// TestScreensWithoutJapaneseFont renders every screen with Japanese glosses
// as it does on a terminal with no font for them, and checks that not one
// kana or kanji is left to draw as a box while the rest stays in color. The
// same screens with the glosses on must show some: otherwise the check
// would pass on screens that never had any.
func TestScreensWithoutJapaneseFont(t *testing.T) {
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })
	t.Cleanup(func() { SetJapanese(true) })

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

	for name, render := range screens {
		SetJapanese(true)
		withGlosses := render()
		assert.NotEmpty(t, japaneseIn(ansi.Strip(withGlosses)), "%s shows Japanese when the terminal can draw it", name)

		SetJapanese(false)
		without := render()
		assert.Empty(t, japaneseIn(ansi.Strip(without)), "%s left Japanese to draw as boxes:\n%s", name, ansi.Strip(without))
		assert.NotEqual(t, ansi.Strip(without), without, "%s keeps its colors: only plain mode drops them", name)
	}
}

func TestShowJapanese_PlainWins(t *testing.T) {
	t.Cleanup(func() {
		SetPlain(false)
		SetJapanese(true)
	})
	SetJapanese(true)
	SetPlain(true)
	assert.False(t, ShowJapanese(), "plain mode hides the glosses even where a font has them")
	SetPlain(false)
	assert.True(t, ShowJapanese())
}
