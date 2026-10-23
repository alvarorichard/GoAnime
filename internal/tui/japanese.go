package tui

import (
	"context"
	_ "embed"
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
// terminal to use once it is reopened (japanese_font_linux.go); Windows
// installs it under the user's fonts folder (japanese_font_windows.go),
// because a Windows Sandbox or an en-US machine ships no Japanese font for
// Windows Terminal to fall back on.
//
// resolveJapanese is the pure decision: env and the platform probes are
// passed in so unit tests cover every branch. ShowJapanese is the cached
// runtime wrapper every screen asks. GOANIME_JAPANESE=1|0 overrides it.

// GoAnime Glyphs is Noto Sans Mono CJK JP cut down to the characters GoAnime
// draws (fonts/build.py): 11 KB instead of the 16 MB a Japanese font takes.

//go:embed fonts/GoAnimeGlyphs.ttf
var glyphFont []byte

//go:embed fonts/OFL.txt
var glyphFontLicense []byte

const glyphFontFile = "GoAnimeGlyphs.ttf"

// writeGlyphFont writes GoAnime Glyphs and its license into dir, writing
// aside and renaming, so a launcher never reads half a font.
func writeGlyphFont(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	if err := root.WriteFile("OFL.txt", glyphFontLicense, 0o600); err != nil {
		return err
	}
	if err := root.WriteFile(glyphFontFile+".tmp", glyphFont, 0o600); err != nil {
		return err
	}
	return root.Rename(glyphFontFile+".tmp", glyphFontFile)
}

// glossProbeRunes are ゴ (a katakana), the long vowel mark ー and 検 (a
// kanji): one of each kind the glosses draw. A font answering them is a
// Japanese font. fontconfig's pattern asks for exactly these, and the
// Windows gdi probe asks for the same three.
var glossProbeRunes = []rune{'ゴ', 'ー', '検'}

// japaneseFontPattern asks fontconfig for fonts holding the glossProbeRunes
// — the katakana ゴ, the mark ー and the kanji 検: a font with all three is a
// Japanese font.
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
	// consoleFontCovers asks gdi whether the one font the classic console
	// draws with today answers every rune — the question cjkCodePage only
	// guessed at from the code page.
	consoleFontCovers func([]rune) bool
	// fonts asks the machine for a font with the glossProbeRunes' glyphs.
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
		if probes.classicConsole() {
			// The classic console draws with the one font its settings chose
			// and never falls back for a missing glyph. The code page is the
			// age-old guess; the font's own gdi coverage answers it. While
			// today's font cannot, this run keeps the glosses hidden — but
			// GoAnime Glyphs installs unless a Japanese face is already
			// accounted for, so the console's own font list offers it.
			show := probes.cjkCodePage() || probes.consoleFontCovers(glossProbeRunes)
			if !show {
				switch probes.fonts() {
				case fontPresent, fontPending:
					// A Japanese face, or the glyphs font already waiting on
					// disk, is accounted for: Properties can select the
					// font, and re-installing would only repeat the note.
				default:
					// Missing, or the probe could not answer at all: install
					// anyway. The install is idempotent, and a probe that
					// cannot ask is no reason to leave the font out.
					probes.installFont()
				}
			}
			return show
		}
		// Windows Terminal, VS Code and every other ConPTY host fall back
		// over the Japanese fonts the machine has — a Windows Sandbox ships
		// without any, so the probe must in fact find one.
		switch probes.fonts() {
		case fontMissing:
			probes.installFont()
			return false
		case fontPending:
			return false
		case fontPresent:
			// A Japanese face answers; the glosses just draw.
			return true
		default:
			// The probe could not answer at all (fontUnknown). The glosses
			// show on evidence it cannot refute — and the glyphs install
			// on the side, so the next open terminal has the font however
			// the probe behaves.
			probes.installFont()
			return true
		}
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
			goos:              runtime.GOOS,
			classicConsole:    classicConsole,
			cjkCodePage:       cjkCodePage,
			consoleFontCovers: probeConsoleFontCovers,
			fonts:             japaneseFontProbe,
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
