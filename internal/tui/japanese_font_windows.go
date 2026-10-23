//go:build windows

package tui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/zzl/go-win32api/win32"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// GoAnime Glyphs is Noto Sans Mono CJK JP cut down to the characters GoAnime
// draws (fonts/build.py): 11 KB instead of the 16 MB a Japanese font takes.
// Windows machines without the Japanese supplemental fonts — a Windows
// Sandbox, an en-US install that never opened any Japanese document — fall
// back over nothing and draw every gloss as a box, so GoAnime installs its
// font for the user and Windows Terminal finds it once it is reopened, the
// way fontconfig finds it on Linux.

// glyphFontFile and the font bytes live in japanese.go, shared with Linux;
// glyphFontFamily is the family name fonts/build.py stamps into the font;
// gdi's enumeration hands it back so the probes can tell our font apart.
const glyphFontFamily = "GoAnime Glyphs"

// The font and console calls go through typed Win32 bindings: the syscall
// pointer plumbing stays in that dependency instead of living here.
//
// gdi's DEFAULT_CHARSET (1): enumerate every face the fonts register, and
// GGI_MARK_NONEXISTING_GLYPHS (1): GetGlyphIndices maps a rune the face
// lacks to FFFF instead of some stand-in glyph.
const (
	defaultCharSet           = 1
	ggiMarkNonexistingGlyphs = 1
	// maxProbeRunes bounds the fixed UTF-16 buffers gdi maps. Every caller
	// asks about a handful of glyphs; the bound keeps the count a typed
	// constant instead of an unchecked int-to-int32 conversion.
	maxProbeRunes = 32
)

// consoleFontFace names the face the classic console draws with today.
// An empty name means the question lost an answer — the handle is not the
// console, or the API refused.
func consoleFontFace() string {
	console, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || console == windows.InvalidHandle {
		return ""
	}
	size := reflect.TypeOf(win32.CONSOLE_FONT_INFOEX{}).Size()
	if size > math.MaxUint32 {
		return ""
	}
	info := win32.CONSOLE_FONT_INFOEX{CbSize: uint32(size)}
	if ok, _ := win32.GetCurrentConsoleFontEx(win32.HANDLE(console), 0, &info); ok == 0 {
		return ""
	}
	return windows.UTF16ToString(info.FaceName[:])
}

// screenDC borrows the screen's device context, where the system's fonts
// live; the done func hands it back. A borrowed context of 0 means gdi could
// not answer at all.
func screenDC() (dc win32.HDC, done func()) {
	dc = win32.GetDC(0)
	if dc == 0 {
		return 0, func() {}
	}
	return dc, func() { win32.ReleaseDC(0, dc) }
}

// utf16Face copies a face name into a LOGFONTW's fixed buffer.
func utf16Face(face string) [32]uint16 {
	var name [32]uint16
	encoded := utf16.Encode([]rune(face))
	if len(encoded) > 31 {
		encoded = encoded[:31]
	}
	copy(name[:], encoded)
	return name
}

// gdiFaceGlyphs asks gdi to map each rune of runes with face and says
// whether that face draws it. ok is false when gdi could not answer at all.
// Every rune here is in the BMP, so runes and their utf16 words pair up.
func gdiFaceGlyphs(dc win32.HDC, face string, runes []rune) (map[rune]bool, bool) {
	if dc == 0 || face == "" || len(runes) == 0 || len(runes) > maxProbeRunes {
		return nil, false
	}
	drawn := make(map[rune]bool, len(runes))
	f := win32.LOGFONTW{LfCharSet: defaultCharSet}
	f.LfFaceName = utf16Face(face)
	hfont := win32.CreateFontIndirectW(&f)
	if hfont == 0 {
		return drawn, false
	}
	defer win32.DeleteObject(hfont)
	previous := win32.SelectObject(dc, hfont)
	defer win32.SelectObject(dc, previous)
	chars := utf16.Encode(runes)
	if len(chars) != len(runes) {
		// The probes ask about BMP runes only, where one rune is one
		// UTF-16 word; a surrogate pair would break the index match.
		return drawn, false
	}
	var charsBuf [maxProbeRunes]uint16
	var mapped [maxProbeRunes]uint16
	copy(charsBuf[:], chars)
	r1 := win32.GetGlyphIndicesW(dc,
		&charsBuf[0], maxProbeRunes,
		&mapped[0], ggiMarkNonexistingGlyphs)
	if r1 == 0 {
		return drawn, false
	}
	for i, r := range runes {
		drawn[r] = mapped[i] != 0xFFFF
	}
	return drawn, true
}

// enumFontState guards the callback's collection: gdi runs it here and now,
// once per enumeration.
var (
	enumMu            sync.Mutex
	enumFacesTalked   []string
	enumFontFacesProc = windows.NewCallback(enumFontFace)
)

// enumFontFace receives a face per call; 1 keeps enumerating, 0 stops it.
// NewCallback hands each argument over register-sized, so the LOGFONTW
// pointer arrives already typed — real pointers, not raw uintptrs to cast.
func enumFontFace(lfont *win32.LOGFONTW, _, _, _ uintptr) uintptr {
	if lfont == nil {
		return 1
	}
	if name := windows.UTF16ToString(lfont.LfFaceName[:]); name != "" {
		enumFacesTalked = append(enumFacesTalked, name)
	}
	return 1
}

// enumFontFaces lists every face name the system's fonts register, each once,
// in gdi's own order. ok is false when gdi could not enumerate at all.
func enumFontFaces() ([]string, bool) {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumFacesTalked = nil
	dc, done := screenDC()
	if dc == 0 {
		return nil, false
	}
	defer done()
	f := win32.LOGFONTW{LfCharSet: defaultCharSet}
	r1 := win32.EnumFontFamiliesExW(dc, &f, enumFontFacesProc, 0, 0)
	if r1 == 0 {
		return nil, false
	}
	faces := make([]string, 0, len(enumFacesTalked))
	seen := make(map[string]bool, len(enumFacesTalked))
	for _, face := range enumFacesTalked {
		if seen[face] {
			continue
		}
		seen[face] = true
		faces = append(faces, face)
	}
	return faces, true
}

// windowsFontState answers for a ConPTY host: Windows Terminal, VS Code,
// WezTerm and friends fall back over the fonts the machine actually has —
// nothing in Windows ships them by itself, so without a Japanese font
// anywhere they cannot draw a gloss however hard they try. GoAnime Glyphs
// alone only counts once the terminal host has opened after its install, as
// on Linux.
func windowsFontState() fontState {
	dc, done := screenDC()
	if dc == 0 {
		return fontUnknown
	}
	defer done()
	faces, ok := enumFontFaces()
	if !ok {
		return fontUnknown
	}
	ours := false
	for _, face := range faces {
		drawn, askOK := gdiFaceGlyphs(dc, face, glossProbeRunes)
		if !askOK {
			continue
		}
		if !every(drawn) {
			continue
		}
		if strings.EqualFold(face, glyphFontFamily) {
			ours = true
			continue
		}
		return fontPresent
	}
	if !ours {
		return fontMissing
	}
	return glyphFontReloadedState()
}

// every reports whether drawn has a glyph for all of its runes.
func every(drawn map[rune]bool) bool {
	for _, yes := range drawn {
		if !yes {
			return false
		}
	}
	return true
}

// glyphFontReloadedState decides for GoAnime Glyphs alone: present once the
// terminal host opened after the install, pending before. An unknown start
// time counts as present, like every probe that cannot answer.
func glyphFontReloadedState() fontState {
	dir, err := userFontDir()
	if err != nil {
		return fontPresent
	}
	info, err := os.Stat(filepath.Join(dir, glyphFontFile))
	if err != nil {
		return fontPresent
	}
	started, ok := consoleHostStart()
	if !ok || started.After(info.ModTime()) {
		return fontPresent
	}
	return fontPending
}

// consoleHosts are the images standing in for the terminal drawing this
// process: a ConPTY host walks the ancestor chain to the future-known one
// and reads its start time.
var consoleHosts = []string{
	"WindowsTerminal.exe", "OpenConsole.exe", "Code.exe",
	"wezterm-gui.exe", "Alacritty.exe", "mintty.exe",
	"Tabby.exe", "Hyper.exe",
}

// consoleHostStart reports when the terminal drawing this process started:
// the start time of the nearest ancestor whose image is a known host.
func consoleHostStart() (time.Time, bool) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()

	parentOf := make(map[uint32]windows.ProcessEntry32, 64)
	var entry windows.ProcessEntry32
	size := reflect.TypeOf(entry).Size()
	if size > math.MaxUint32 {
		return time.Time{}, false
	}
	entry.Size = uint32(size)
	for err := windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		parentOf[entry.ProcessID] = entry
	}

	// Windows PIDs are 32-bit DWORDs; x/sys hands the current one over natively,
	// so no int→uint32 conversion (and no overflow doubt) happens here.
	pid := windows.GetCurrentProcessId()
	for range 64 {
		item, ok := parentOf[pid]
		if !ok {
			return time.Time{}, false
		}
		if isConsoleHost(windows.UTF16ToString(item.ExeFile[:])) {
			return processStartTime(item.ProcessID)
		}
		pid = item.ParentProcessID
		if pid <= 4 {
			return time.Time{}, false
		}
	}
	return time.Time{}, false
}

func isConsoleHost(name string) bool {
	for _, host := range consoleHosts {
		if strings.EqualFold(host, name) {
			return true
		}
	}
	return false
}

// processStartTime reads when a process started.
func processStartTime(pid uint32) (time.Time, bool) {
	const processQueryLimitedInformation = 0x0800
	handle, err := windows.OpenProcess(processQueryLimitedInformation, false, pid)
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return time.Time{}, false
	}
	return time.Unix(0, creation.Nanoseconds()), true
}

// userFontDir is the per-user folder Windows reads fonts from — one of the
// user's own, so GoAnime never touches a system directory.
func userFontDir() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" || !filepath.IsAbs(local) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		local = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(local, "Microsoft", "Windows", "Fonts"), nil
}

// installGlyphFont installs GoAnime Glyphs for the user: the license, then
// the font, then gdi loading it for this session and the per-user Fonts
// registry key the next login finds it from, and a fontchange broadcast so
// running apps refresh their lists. The terminal that is open now loaded its
// fonts before, so the glosses appear once it is reopened.
func installGlyphFont() string {
	dir, err := userFontDir()
	if err != nil {
		return "No Japanese font, and no fonts folder to install GoAnime Glyphs into: " + err.Error()
	}
	return installGlyphFontInto(dir, true)
}

// installGlyphFontInto writes and loads GoAnime Glyphs from dir, and owns up
// to whatever happens in its note. register says whether the per-user Fonts
// registry key and the fontchange broadcast go with it: the real install
// registers, while a test installing into a scratch folder keeps register
// false so the machine's own font registry is never written. After a load
// that answers, the note is only allowed to say "installed" once gdi itself
// confirms the face maps every Japanese probe glyph — an install this
// machine cannot see drawing is recorded as a failure, not a hope.
func installGlyphFontInto(dir string, register bool) string {
	if err := writeGlyphFont(dir); err != nil {
		return "No Japanese font, and installing GoAnime Glyphs failed: " + err.Error()
	}
	path := filepath.Join(dir, glyphFontFile)
	fontPath, err := windows.UTF16PtrFromString(path)
	if err != nil || win32.AddFontResourceW(fontPath) == 0 {
		return "No Japanese font, and GoAnime Glyphs in " + dir + " did not load this session; reopen the terminal and GoAnime again"
	}
	if register {
		fonts, err := registry.OpenKey(registry.CURRENT_USER,
			`Software\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.SET_VALUE)
		if err == nil {
			defer func() { _ = fonts.Close() }()
			_ = fonts.SetStringValue(glyphFontFamily+" (TrueType)", path)
		}
		_, _ = win32.PostMessageW(win32.HWND(0xFFFF), win32.WM_FONTCHANGE, 0, 0)
	}
	note := fmt.Sprintf("No Japanese font: installed GoAnime Glyphs in %s. %s", dir, terminalAdvice())
	drawn, ok := gdiFaceGlyphsOnScreen(glyphFontFamily, glossProbeRunes)
	switch {
	case ok && every(drawn):
		note = fmt.Sprintf("No Japanese font: installed GoAnime Glyphs in %s, and this session's gdi maps every Japanese glyph it carries. %s", dir, terminalAdvice())
	case ok:
		// The resource loaded, yet the face stays silent: that is an
		// install that did not take, and the note says so instead of
		// promising glyphs that will not draw.
		note = fmt.Sprintf("No Japanese font, and GoAnime Glyphs in %s did not answer this session; reopen the terminal and GoAnime again", dir)
	}
	return note
}

// gdiFaceGlyphsOnScreen borrows the screen's device context and asks face
// whether it draws each of runes. ok is false when gdi could not answer.
func gdiFaceGlyphsOnScreen(face string, runes []rune) (map[rune]bool, bool) {
	dc, done := screenDC()
	if dc == 0 {
		return nil, false
	}
	defer done()
	return gdiFaceGlyphs(dc, face, runes)
}

// terminalAdvice names what this terminal must do before the glosses draw,
// because the two Windows hosts listen differently: a ConPTY host built its
// font collection at startup, so reopening it is enough, while the classic
// console only ever draws with the one face its Properties selected — and a
// console opened after the install lists GoAnime Glyphs there.
func terminalAdvice() string {
	if classicConsole() {
		return "Pick GoAnime Glyphs in this console's Properties — a console opened after the install lists it"
	}
	return "The labels appear once the terminal is reopened"
}

// japaneseFontProbe answers on Windows with gdi: which installed faces have
// the glyphs — not with fontconfig, fc-list which nothing here runs.
func japaneseFontProbe() fontState { return windowsFontState() }

// glyphFontState answers for files an fc-list found. fontconfig never runs on
// Windows, so this never answers beyond its placeholder.
func glyphFontState([]string) fontState { return fontPresent }
