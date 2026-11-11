package tui

// Shared plumbing of the regression net (regression_home_test.go and
// regression_results_test.go): pinning the process-wide display settings,
// the golden files, and driving a model with named keys.
//
// Golden files live in testdata/regression/<name>.golden. Each holds sections
// "=== <name> ===" whose lines are the screen with its ANSI stripped, every
// line cut of trailing blanks and ended with "|" so the end of a line is
// visible. They are written ONLY on purpose:
//
//	REGRESSION_UPDATE=1 go test -run Regression ./internal/tui
//
// and the diff of the golden files is then the review of the change.

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regSize is a terminal size the screens are frozen at.
type regSize struct{ w, h int }

func (s regSize) String() string { return fmt.Sprintf("%dx%d", s.w, s.h) }

var regSizes = []regSize{{140, 40}, {100, 30}, {80, 24}, {60, 20}}

// regPin fixes the process-wide display settings the renders depend on: plain
// or rich, flavor glyphs on, Japanese glosses on, dark theme. It restores what
// was there when the test ends. Tests that pin are not parallel: the settings
// are shared by the whole package.
func regPin(t *testing.T, plain bool) {
	t.Helper()
	plainMu.RLock()
	prevPlain := plainOverride
	plainMu.RUnlock()
	themeMu.RLock()
	prevTheme := themeOverride
	themeMu.RUnlock()
	glyphMu.RLock()
	prevGlyph := glyphOverride
	glyphMu.RUnlock()
	japaneseMu.RLock()
	prevJapanese := japaneseOverride
	japaneseMu.RUnlock()
	t.Cleanup(func() {
		plainMu.Lock()
		plainOverride = prevPlain
		plainMu.Unlock()
		themeMu.Lock()
		themeOverride = prevTheme
		themeMu.Unlock()
		glyphMu.Lock()
		glyphOverride = prevGlyph
		glyphMu.Unlock()
		japaneseMu.Lock()
		japaneseOverride = prevJapanese
		japaneseMu.Unlock()
	})
	SetPlain(plain)
	SetThemeIsDark(true)
	SetGlyphs(true)
	SetJapanese(true)
}

func regMode(plain bool) string {
	if plain {
		return "plain"
	}
	return "rich"
}

// regFrame is a rendered screen as the golden files hold it. It also checks
// what every screen promises: no line wider than the terminal, no more lines
// than rows.
func regFrame(t *testing.T, content string, size regSize) string {
	t.Helper()
	lines := strings.Split(ansi.Strip(content), "\n")
	assert.LessOrEqual(t, len(lines), size.h, "the screen fits the height")
	out := make([]string, len(lines))
	for i, line := range lines {
		assert.LessOrEqual(t, ansi.StringWidth(line), size.w, "line %d fits the width: %q", i+1, line)
		out[i] = strings.TrimRight(line, " ") + "|"
	}
	return strings.Join(out, "\n")
}

// ---- golden files ---------------------------------------------------------------

var regGoldens = struct {
	sync.Mutex
	files map[string]*regGoldenFile
}{files: map[string]*regGoldenFile{}}

type regGoldenFile struct {
	order    []string
	sections map[string]string
}

func regGoldenPath(name string) string {
	return filepath.Join("testdata", "regression", name+".golden")
}

func regLoadGolden(t *testing.T, name string) *regGoldenFile {
	t.Helper()
	g := &regGoldenFile{sections: map[string]string{}}
	if os.Getenv("REGRESSION_UPDATE") == "1" {
		return g // an update starts every file from nothing: no stale section survives
	}
	data, err := os.ReadFile(regGoldenPath(name))
	require.NoError(t, err, "golden file %s is missing; REGRESSION_UPDATE=1 writes it", name)
	var current string
	var body []string
	flush := func() {
		if current != "" {
			g.order = append(g.order, current)
			g.sections[current] = strings.Join(body, "\n")
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		if strings.HasPrefix(line, "=== ") && strings.HasSuffix(line, " ===") {
			flush()
			current, body = strings.TrimSuffix(strings.TrimPrefix(line, "=== "), " ==="), nil
			continue
		}
		body = append(body, line)
	}
	flush()
	return g
}

// regGolden compares got with section of the golden file name.
func regGolden(t *testing.T, name, section, got string) {
	t.Helper()
	regGoldens.Lock()
	defer regGoldens.Unlock()
	g, ok := regGoldens.files[name]
	if !ok {
		g = regLoadGolden(t, name)
		regGoldens.files[name] = g
	}
	if os.Getenv("REGRESSION_UPDATE") == "1" {
		if _, seen := g.sections[section]; !seen {
			g.order = append(g.order, section)
		}
		g.sections[section] = got
		var sb strings.Builder
		for _, s := range g.order {
			sb.WriteString("=== " + s + " ===\n" + g.sections[s] + "\n")
		}
		require.NoError(t, os.MkdirAll(filepath.Dir(regGoldenPath(name)), 0o755))
		require.NoError(t, os.WriteFile(regGoldenPath(name), []byte(sb.String()), 0o644))
		return
	}
	want, found := g.sections[section]
	require.True(t, found, "golden %s has no section %q; REGRESSION_UPDATE=1 writes it", name, section)
	if want != got {
		assert.Equal(t, want, got, "golden %s, section %q", name, section)
	}
}

// ---- driving a model -------------------------------------------------------------

// regKey turns a name into the key press. "type:abc" is not a key: see regPress.
func regKey(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "pgup":
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyPressMsg{Code: tea.KeyPgDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "space":
		return tea.KeyPressMsg{Code: ' ', Text: " "}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	case "ctrl+f":
		return tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	case "ctrl+l":
		return tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl}
	}
	r := []rune(name)
	if len(r) != 1 {
		panic("regKey: unknown key " + name)
	}
	return tea.KeyPressMsg{Code: r[0], Text: name}
}

// regIsQuit reports whether cmd is tea.Quit.
func regIsQuit(cmd tea.Cmd) bool {
	return cmd != nil && reflect.ValueOf(cmd).Pointer() == reflect.ValueOf(tea.Quit).Pointer()
}

// regPress sends the keys one by one ("type:abc" types a, b and c) and says
// whether the last one asked to quit.
func regPress(m tea.Model, keys ...string) (quit bool) {
	for _, k := range keys {
		if text, ok := strings.CutPrefix(k, "type:"); ok {
			for _, r := range text {
				_, _ = m.Update(regKey(string(r)))
			}
			quit = false
			continue
		}
		_, cmd := m.Update(regKey(k))
		quit = regIsQuit(cmd)
	}
	return quit
}

// regHintLine is a footer as one line: key:description, a "!" after the
// essential ones, " | " between them.
func regHintLine(hs []Hint) string {
	parts := make([]string, len(hs))
	for i, h := range hs {
		parts[i] = h.Key + ":" + h.Desc
		if h.Essential {
			parts[i] += "!"
		}
	}
	return strings.Join(parts, " | ")
}
