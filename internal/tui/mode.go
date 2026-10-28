package tui

import (
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

// Display mode and theme resolution for GoAnime's TUI.
//
// The design (telas 8 e 9) requires two behaviors:
//
//  1. Light terminals: GoAnime asks the terminal for its background color once,
//     before the first RunClean, and keeps the answer for the whole process.
//     GOANIME_THEME=light|dark wins when the terminal does not answer.
//  2. Plain mode: NO_COLOR, TERM=dumb, a locale without UTF-8, or --plain /
//     GOANIME_PLAIN turn the TUI into plain text: the cursor is ">" with
//     reverse video instead of color, Japanese glosses are hidden (they render
//     as boxes without a CJK font and confuse screen readers), the braille
//     halftone is dropped, and status marks become x ~ *.
//
// The pure Resolve* functions take an explicit env so unit tests cover every
// branch without touching the process environment. Plain() and IsDark() are
// the cached runtime wrappers used by every screen.

// plainOverride forces plain mode regardless of the environment (tests and
// the --plain flag).
var (
	plainOverride *bool
	plainMu       sync.RWMutex
)

// SetPlain forces plain mode on (true) or off (false) for the process.
// The --plain flag calls this during startup.
func SetPlain(plain bool) {
	plainMu.Lock()
	defer plainMu.Unlock()
	plainOverride = &plain
}

// ResolvePlain reports whether env asks for the plain interface: NO_COLOR,
// TERM=dumb, a non-UTF-8 locale, or GOANIME_PLAIN=1/true.
func ResolvePlain(env []string) bool {
	get := func(key string) (string, bool) {
		prefix := key + "="
		for _, item := range env {
			if strings.HasPrefix(item, prefix) {
				return strings.TrimSpace(item[len(prefix):]), true
			}
		}
		return "", false
	}
	if _, ok := get("NO_COLOR"); ok {
		return true
	}
	if termVal, ok := get("TERM"); ok && strings.EqualFold(strings.TrimSpace(termVal), "dumb") {
		return true
	}
	if value, ok := get("GOANIME_PLAIN"); ok && parseBool(value) {
		return true
	}
	return resolveNonUTF8Locale(env)
}

// resolveNonUTF8Locale reports whether a locale is set without UTF-8.
// LC_ALL wins over LC_CTYPE, which wins over LANG. An unset locale keeps the
// rich interface: most systems render UTF-8 fine even with LANG=C.
func resolveNonUTF8Locale(env []string) bool {
	lookup := func(key string) (string, bool) {
		prefix := key + "="
		for _, item := range env {
			if strings.HasPrefix(item, prefix) {
				return item[len(prefix):], true
			}
		}
		return "", false
	}
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value, ok := lookup(key)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || strings.EqualFold(value, "C") || strings.EqualFold(value, "POSIX") {
			// Bare "C"/"POSIX" with no other locale set says nothing about
			// UTF-8 support; keep looking at the next variable.
			continue
		}
		return !strings.Contains(strings.ToLower(value), "utf-8") &&
			!strings.Contains(strings.ToLower(value), "utf8")
	}
	return false
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// Plain reports whether the TUI must render as plain text.
func Plain() bool {
	plainMu.RLock()
	override := plainOverride
	plainMu.RUnlock()
	if override != nil {
		return *override
	}
	for _, arg := range os.Args[1:] {
		if arg == "--plain" || arg == "-plain" {
			return true
		}
	}
	return ResolvePlain(os.Environ())
}

// themeOverride forces the dark (true) or light (false) theme, set by
// GOANIME_THEME=dark|light. Empty/nil means "ask the terminal".
var (
	themeOverride *bool
	themeMu       sync.RWMutex

	detectedIsDark     = true // the safe default: almost every terminal is dark
	detectedTheme      bool
	detectedThemeMutex sync.Mutex
)

// SetThemeIsDark forces the dark (true) or light (false) theme for tests.
func SetThemeIsDark(isDark bool) {
	themeMu.Lock()
	defer themeMu.Unlock()
	themeOverride = &isDark
}

// ResolveIsDark maps GOANIME_THEME to a theme. It returns explicit=false when
// the terminal should be asked (empty or "auto").
func ResolveIsDark(env []string) (isDark, explicit bool) {
	for _, item := range env {
		if !strings.HasPrefix(item, "GOANIME_THEME=") {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(item, "GOANIME_THEME="))) {
		case "light":
			return false, true
		case "dark":
			return true, true
		default:
			return true, false
		}
	}
	return true, false
}

// IsDark reports whether the TUI must use the dark theme.
func IsDark() bool {
	themeMu.RLock()
	override := themeOverride
	themeMu.RUnlock()
	if override != nil {
		return *override
	}
	if isDark, explicit := ResolveIsDark(os.Environ()); explicit {
		return isDark
	}
	detectedThemeMutex.Lock()
	defer detectedThemeMutex.Unlock()
	if detectedTheme {
		return detectedIsDark
	}
	detectedIsDark = queryTerminalIsDark()
	detectedTheme = true
	return detectedIsDark
}

// CurrentTheme builds the theme for this process: dark or light following
// IsDark(), with reverse-video selection when plain (no color to mark the
// cursor row with).
func CurrentTheme() Theme {
	t := NewTheme(IsDark())
	if Plain() {
		applyPlainSelection(&t)
	}
	return t
}

// applyPlainSelection turns the cursor row into reverse video instead of
// color: with NO_COLOR there is no crimson to fill it with (tela 9).
func applyPlainSelection(t *Theme) {
	reverse := lipgloss.NewStyle().Reverse(true)
	t.Selected = reverse.Bold(true)
	t.SelectedMuted = reverse
	t.SelectedFaint = reverse
	t.FilterMatch = lipgloss.NewStyle().Underline(true)
}

// EnsureThemeDetected asks the terminal for its background once, before the
// first RunClean suppresses capability probes. Callers do not need it:
// IsDark detects lazily. It exists so startup can pay the query cost up
// front, once, instead of on the first screen.
func EnsureThemeDetected() {
	themeMu.RLock()
	override := themeOverride
	themeMu.RUnlock()
	if override != nil {
		return
	}
	if isDark, explicit := ResolveIsDark(os.Environ()); explicit {
		detectedThemeMutex.Lock()
		detectedIsDark, detectedTheme = isDark, true
		detectedThemeMutex.Unlock()
		return
	}
	detectedThemeMutex.Lock()
	defer detectedThemeMutex.Unlock()
	if !detectedTheme {
		detectedIsDark = queryTerminalIsDark()
		detectedTheme = true
	}
}

// queryTerminalIsDark asks the terminal for its background color with OSC 11
// and reports whether it is dark. Anything inconclusive (no TTY, no answer,
// unparsable answer) keeps the dark default.
func queryTerminalIsDark() bool {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) || !isStdoutTerminal() {
		return true
	}
	light, ok := queryBackgroundIsLight(osc11Timeout)
	if !ok {
		return true
	}
	return !light
}

func isStdoutTerminal() bool {
	if f := os.Stdout; f != nil {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
