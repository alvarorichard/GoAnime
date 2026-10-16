package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Resolve* functions are pure: every branch is covered without touching
// the process environment (t.Setenv forbids parallel tests, so these stay
// sequential and hermetic).

func TestResolvePlain(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  []string
		want bool
	}{
		{"no_color", []string{"NO_COLOR=1"}, true},
		{"no_color_empty", []string{"NO_COLOR="}, true},
		{"term_dumb", []string{"TERM=dumb"}, true},
		{"term_dumb_case", []string{"TERM=DUMB"}, true},
		{"goanime_plain", []string{"GOANIME_PLAIN=1"}, true},
		{"goanime_plain_true", []string{"GOANIME_PLAIN=true"}, true},
		{"goanime_plain_zero", []string{"GOANIME_PLAIN=0"}, false},
		{"locale_without_utf8", []string{"LANG=pt_BR.ISO-8859-1"}, true},
		{"lc_all_utf8_wins", []string{"LC_ALL=en_US.UTF-8", "LANG=pt_BR.ISO-8859-1"}, false},
		{"bare_c_keeps_rich", []string{"LC_ALL=C"}, false},
		{"bare_posix_keeps_rich", []string{"LANG=POSIX"}, false},
		{"utf8_locale", []string{"LANG=pt_BR.UTF-8"}, false},
		{"c_utf8", []string{"LANG=C.UTF-8"}, false},
		{"unset_locale", []string{"TERM=xterm-256color"}, false},
		{"empty", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ResolvePlain(tc.env))
		})
	}
}

func TestResolveIsDark(t *testing.T) {
	isDark, explicit := ResolveIsDark([]string{"GOANIME_THEME=light"})
	assert.False(t, isDark)
	assert.True(t, explicit)

	isDark, explicit = ResolveIsDark([]string{"GOANIME_THEME=dark"})
	assert.True(t, isDark)
	assert.True(t, explicit)

	isDark, explicit = ResolveIsDark([]string{"GOANIME_THEME=auto"})
	assert.True(t, isDark, "auto falls back to dark until the terminal is asked")
	assert.False(t, explicit)

	isDark, explicit = ResolveIsDark(nil)
	assert.True(t, isDark)
	assert.False(t, explicit)
}

func TestParseBackgroundIsLight(t *testing.T) {
	light, ok := parseBackgroundIsLight("\x1b]11;rgb:ffff/ffff/ffff\x07")
	assert.True(t, ok)
	assert.True(t, light, "white background is light")

	light, ok = parseBackgroundIsLight("\x1b]11;rgb:0000/0000/0000\x1b\\")
	assert.True(t, ok)
	assert.False(t, light, "black background is dark")

	light, ok = parseBackgroundIsLight("\x1b]11;rgb:ff/ff/ff\x07")
	assert.True(t, ok)
	assert.True(t, light, "short hex digits scale up")

	_, ok = parseBackgroundIsLight("no color here")
	assert.False(t, ok)

	_, ok = parseBackgroundIsLight("\x1b]11;rgb:zzzz/0000/0000\x07")
	assert.False(t, ok)
}

func TestStatusMark_PlainUsesASCII(t *testing.T) {
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })

	assert.Equal(t, "✓", statusMark(PickWatched))
	assert.Equal(t, "◐", statusMark(PickInProgress))
	assert.Equal(t, "●", statusMark(PickCurrent))

	SetPlain(true)
	assert.Equal(t, "x", statusMark(PickWatched))
	assert.Equal(t, "~", statusMark(PickInProgress))
	assert.Equal(t, "*", statusMark(PickCurrent))
	assert.Equal(t, " ", statusMark(PickNone))
}

func TestShellHeader_PlainHidesGloss(t *testing.T) {
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })

	theme := NewTheme(true)
	shell := NewShell(&theme, "Search > Results")
	rich := ansi.Strip(shell.header(120))
	assert.Contains(t, rich, "ゴーアニメ", "rich header keeps the brand gloss")

	SetPlain(true)
	plain := ansi.Strip(shell.header(120))
	assert.NotContains(t, plain, "ゴーアニメ", "plain header hides Japanese glosses (tela 9)")
	assert.NotContains(t, plain, "結果")
	assert.Contains(t, plain, "RESULTS")
	assert.Contains(t, plain, ">", "plain breadcrumb uses ASCII separators")
}

func TestFitPanel_PlainIsTextBelow(t *testing.T) {
	theme := NewTheme(true)
	content := &panelContent{
		label: "Details", gloss: "詳細", title: "Momotarou",
		fields: []detailField{{"Source", "AnimeFire"}, {"Year", "1918"}},
	}

	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })
	rich := fitPanel(&theme, content, 40, 30)
	require.NotEmpty(t, rich)
	assert.Contains(t, ansi.Strip(rich), "詳細")

	SetPlain(true)
	plain := fitPanel(&theme, content, 40, 30)
	require.NotEmpty(t, plain)
	stripped := ansi.Strip(plain)
	assert.Contains(t, stripped, "DETAILS")
	assert.Contains(t, stripped, "Momotarou")
	assert.NotContains(t, stripped, "詳細", "plain panel hides the gloss")
	assert.NotContains(t, stripped, "⠁", "plain panel drops the braille halftone")
	assert.NotContains(t, stripped, "█", "plain panel drops the block shadow")
}

func TestRenderBar_PlainUsesASCII(t *testing.T) {
	theme := NewTheme(true)
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	bar := ansi.Strip(renderBar(10, 0.5, &theme.Meter, &theme.MeterEmpty))
	assert.Equal(t, "=====-----", bar)
	assert.NotContains(t, bar, "━")
}

func TestRenderHelp(t *testing.T) {
	theme := NewTheme(true)
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })

	help := RenderHelp(&theme, "Results", []HelpRow{
		{Key: "esc", Desc: "edit the search"},
		{Key: "q", Desc: "quit"},
	}, nil, 100, 30)
	require.NotEmpty(t, help)
	stripped := ansi.Strip(help)
	assert.Contains(t, stripped, "KEYS")
	assert.Contains(t, stripped, "EVERY SCREEN")
	assert.Contains(t, stripped, "THIS SCREEN")
	assert.Contains(t, stripped, "RESULTS")
	assert.Contains(t, stripped, "LISTS")
	assert.Contains(t, stripped, "esc never quits. ctrl+c always does.")

	SetPlain(true)
	plain := RenderHelp(&theme, "Results", nil, nil, 100, 30)
	require.NotEmpty(t, plain)
	assert.NotContains(t, ansi.Strip(plain), "キー", "plain help hides Japanese")
	assert.True(t, strings.Contains(plain, "up/dn"), "plain help uses ASCII move keys")
}

func TestCurrentTheme_PlainReversesSelection(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	theme := CurrentTheme()
	assert.True(t, theme.Selected.GetReverse(), "plain cursor row is reverse video, not color")
}

func TestCurrentTheme_RespectsThemeOverride(t *testing.T) {
	SetThemeIsDark(false)
	t.Cleanup(func() { SetThemeIsDark(true) })
	assert.False(t, IsDark())
	light := CurrentTheme()

	SetThemeIsDark(true)
	assert.True(t, IsDark())
	dark := CurrentTheme()

	assert.NotEqual(t, light.Primary.GetForeground(), dark.Primary.GetForeground())
}
