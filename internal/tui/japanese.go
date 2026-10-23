package tui

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Japanese glosses (ゴーアニメ, 検索, ソース…) are drawn by the terminal with
// its own fonts. On a machine without a Japanese font every gloss becomes a
// row of boxes, as in the Arch Linux recording of #215, so GoAnime hides
// them there, as plain mode does, instead of printing boxes. On Linux it
// also installs GoAnime Glyphs, a font with just those characters, for the
// terminal to use once it is reopened (japanese_font_linux.go).
//
// resolveJapanese is the pure decision: env and the platform probes are
// passed in so unit tests cover every branch. ShowJapanese is the cached
// runtime wrapper every screen asks. GOANIME_JAPANESE=1|0 overrides it.

// japaneseFontPattern asks fontconfig for fonts holding ゴ (U+30B4), the
// long vowel mark ー (U+30FC) and 検 (U+691C): a katakana, the mark and a
// kanji from the glosses. A font with all three is a Japanese font.
const japaneseFontPattern = ":charset=30b4 30fc 691c"

// fontProbeTimeout bounds fc-list at startup. A probe that does not answer
// in time keeps the glosses.
const fontProbeTimeout = time.Second

// fontState is what fontconfig knows about a font for the glosses.
type fontState int

const (
	fontUnknown fontState = iota // there is no fontconfig to ask
	fontMissing                  // no installed font has the glyphs
	fontPending                  // only GoAnime Glyphs has them, installed after the terminal started
	fontPresent
)

// japaneseProbes are the questions about the machine behind the decision.
type japaneseProbes struct {
	goos string
	// classicConsole reports whether the process writes to a classic Windows
	// console window (conhost). It draws with one font and, unlike Windows
	// Terminal, never falls back to another for a missing glyph.
	classicConsole func() bool
	// cjkCodePage reports whether Windows runs a Chinese, Japanese or Korean
	// code page, whose console fonts carry kana and kanji.
	cjkCodePage func() bool
	// fonts asks fontconfig for a font with the glosses' glyphs.
	fonts func() fontState
	// installFont installs GoAnime Glyphs when no font has the glyphs.
	installFont func()
}

// resolveJapanese reports whether the terminal can draw the Japanese glosses.
// It hides them only on evidence that it cannot: when the fonts doing the
// drawing cannot be asked, the glosses stay.
func resolveJapanese(env []string, probes *japaneseProbes) bool {
	get := func(key string) string {
		prefix := key + "="
		for _, item := range env {
			if strings.HasPrefix(item, prefix) {
				return strings.TrimSpace(item[len(prefix):])
			}
		}
		return ""
	}
	switch strings.ToLower(get("GOANIME_JAPANESE")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	// The Linux virtual console font holds 512 glyphs at most: no kana.
	if strings.EqualFold(get("TERM"), "linux") {
		return false
	}
	switch probes.goos {
	case "darwin", "ios", "android":
		// macOS ships Hiragino and Android its CJK fonts, and their
		// terminals fall back to them.
		return true
	case "windows":
		// Windows Terminal, VS Code and every other ConPTY host fall back
		// to the Japanese fonts Windows ships. The classic console does not.
		return !probes.classicConsole() || probes.cjkCodePage()
	}
	// The terminal draws with the fonts of the machine it runs on. Under
	// WSL, over SSH or without a local display, that machine is not this
	// one and its fonts cannot be asked.
	for _, key := range []string{"WSL_DISTRO_NAME", "SSH_CONNECTION", "SSH_TTY"} {
		if get(key) != "" {
			return true
		}
	}
	if get("DISPLAY") == "" && get("WAYLAND_DISPLAY") == "" {
		return true
	}
	switch probes.fonts() {
	case fontMissing:
		probes.installFont()
		return false
	case fontPending:
		return false
	default:
		return true
	}
}

// fontconfigState asks fontconfig for a font with the glosses' glyphs. Linux
// and BSD terminals find the font for a glyph their own font lacks through
// fontconfig, so no match means boxes.
func fontconfigState() fontState {
	ctx, cancel := context.WithTimeout(context.Background(), fontProbeTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "fc-list", "--format=%{file}\\n", japaneseFontPattern).Output()
	if err != nil {
		return fontUnknown
	}
	var files []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if file := strings.TrimSpace(line); file != "" {
			files = append(files, file)
		}
	}
	if len(files) == 0 {
		return fontMissing
	}
	return glyphFontState(files)
}

var (
	japaneseOverride *bool
	japaneseMu       sync.RWMutex
	japaneseOnce     sync.Once
	japaneseDetected bool
)

// SetJapanese forces the Japanese glosses on (true) or off (false) for the
// process, whatever the fonts. Tests pin it so renders do not depend on the
// fonts of the machine running them.
func SetJapanese(show bool) {
	japaneseMu.Lock()
	defer japaneseMu.Unlock()
	japaneseOverride = &show
}

// ShowJapanese reports whether screens draw their Japanese glosses: never in
// plain mode, and otherwise only where the terminal has a font for them.
func ShowJapanese() bool {
	if Plain() {
		return false
	}
	japaneseMu.RLock()
	override := japaneseOverride
	japaneseMu.RUnlock()
	if override != nil {
		return *override
	}
	EnsureJapaneseDetected()
	return japaneseDetected
}

// EnsureJapaneseDetected looks for a Japanese font once. Startup calls it
// before the first RunClean, which rewrites TERM while a screen runs.
// Callers do not need it: ShowJapanese detects lazily. Plain mode never
// shows the glosses, so it neither looks nor installs.
func EnsureJapaneseDetected() {
	if Plain() {
		return
	}
	japaneseOnce.Do(func() {
		japaneseDetected = resolveJapanese(os.Environ(), &japaneseProbes{
			goos:           runtime.GOOS,
			classicConsole: classicConsole,
			cjkCodePage:    cjkCodePage,
			fonts:          fontconfigState,
			installFont: func() {
				// Only for someone at the terminal: not for a pipe or a script.
				if HasTerminal() && isStdoutTerminal() {
					japaneseNote = installGlyphFont()
				}
			},
		})
	})
}

// japaneseNote says what the detection did about a missing font, for the
// debug log. It is written once, inside japaneseOnce.
var japaneseNote string

// JapaneseNote reports what startup did about a missing Japanese font, such
// as installing GoAnime Glyphs, or "" when it did nothing.
func JapaneseNote() string {
	EnsureJapaneseDetected()
	return japaneseNote
}
