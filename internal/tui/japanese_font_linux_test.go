//go:build linux

package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallGlyphFont(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	dir := filepath.Join(data, "fonts", "goanime")

	note := installGlyphFont()
	assert.Contains(t, note, dir)
	assert.Contains(t, note, "reopened")

	font, err := os.ReadFile(filepath.Join(dir, glyphFontFile))
	require.NoError(t, err)
	assert.Equal(t, glyphFont, font)
	license, err := os.ReadFile(filepath.Join(dir, "OFL.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(license), "SIL Open Font License", "the OFL travels with the font")
	assert.NoFileExists(t, filepath.Join(dir, glyphFontFile+".tmp"))

	info, err := os.Stat(filepath.Join(dir, glyphFontFile))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	again := installGlyphFont()
	assert.Equal(t, note, again, "installing over an earlier install works")
}

func TestInstallGlyphFont_Fails(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	// A file where the fonts folder should be.
	require.NoError(t, os.WriteFile(filepath.Join(data, "fonts"), nil, 0o600))

	note := installGlyphFont()
	assert.Contains(t, note, "installing GoAnime Glyphs failed")
}

func TestGlyphFontDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/data")
	dir, err := glyphFontDir()
	require.NoError(t, err)
	assert.Equal(t, "/data/fonts/goanime", dir)

	t.Setenv("XDG_DATA_HOME", "relative/data")
	t.Setenv("HOME", "/home/someone")
	dir, err = glyphFontDir()
	require.NoError(t, err)
	assert.Equal(t, "/home/someone/.local/share/fonts/goanime", dir, "a relative XDG_DATA_HOME is ignored, as the XDG spec asks")
}

func TestGlyphFontState(t *testing.T) {
	saved := terminalStart
	t.Cleanup(func() { terminalStart = saved })
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	ours := filepath.Join(data, "fonts", "goanime", glyphFontFile)

	assert.Equal(t, fontPresent, glyphFontState([]string{"/usr/share/fonts/noto/NotoSansCJK.ttc", ours}),
		"another Japanese font: the terminal already had it")

	installed := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, fontPending, reloadedState(installed, installed.Add(-time.Hour), true),
		"a terminal opened before the install has not loaded GoAnime Glyphs")
	assert.Equal(t, fontPending, reloadedState(installed, installed, true))
	assert.Equal(t, fontPresent, reloadedState(installed, installed.Add(time.Second), true),
		"a terminal opened after the install has")
	assert.Equal(t, fontPresent, reloadedState(installed, time.Time{}, false),
		"an unknown start time keeps the glosses")
}

func TestParseProcStat(t *testing.T) {
	t.Parallel()

	line := "4242 (zsh) S 4100 4242 4242 34816 4300 4194560 2300 9000 1 2 3 4 5 6 20 0 1 0 9433605 12345 678"
	got, ok := parseProcStat([]byte(line))
	require.True(t, ok)
	assert.Equal(t, procStat{ppid: 4100, tty: 34816, start: 9433605}, got)

	tricky := "77 (a) b (c) S 1 77 77 0 -1 4194560 2300 9000 1 2 3 4 5 6 20 0 1 0 500 12345 678"
	got, ok = parseProcStat([]byte(tricky))
	require.True(t, ok, "a command name with spaces and parentheses")
	assert.Equal(t, procStat{ppid: 1, tty: 0, start: 500}, got)

	for _, bad := range []string{"", "4242 zsh S 1", "4242 (zsh) S 1 2 3", "4242 (zsh) S x 4242 4242 34816 4300 4194560 2300 9000 1 2 3 4 5 6 20 0 1 0 9433605"} {
		_, ok := parseProcStat([]byte(bad))
		assert.False(t, ok, "%q", bad)
	}
}

func TestTicksToDuration(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 1500*time.Millisecond, ticksToDuration(150))
	// Three years of uptime: ticks times time.Second would overflow int64.
	threeYears := 3 * 365 * 24 * time.Hour
	ticks := int64(threeYears/time.Second)*clockTicks + 50
	assert.Equal(t, threeYears+500*time.Millisecond, ticksToDuration(ticks))
}

func TestTerminalStarted(t *testing.T) {
	t.Parallel()

	// go test runs without a controlling terminal here or with one under a
	// developer's shell; either way the answer must be sane, never a panic.
	started, ok := terminalStarted()
	if ok {
		assert.True(t, started.Before(time.Now().Add(time.Minute)), "the terminal started in the past")
		assert.True(t, started.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	}

	proc, err := os.OpenRoot("/proc")
	require.NoError(t, err)
	defer func() { _ = proc.Close() }()
	boot, ok := bootTime(proc)
	require.True(t, ok)
	assert.True(t, boot.Before(time.Now()))
	self, ok := readProcStat(proc, os.Getpid())
	require.True(t, ok)
	assert.Equal(t, os.Getppid(), self.ppid)
}

// TestJapaneseFontCascade follows the #215 machine with the real fontconfig:
// a Linux desktop with no Japanese font. Each stage needs the one before, so
// the first broken stage names the hop that failed.
func TestJapaneseFontCascade(t *testing.T) {
	if _, err := exec.LookPath("fc-list"); err != nil {
		t.Skip("fontconfig is not installed")
	}
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	cfg := filepath.Join(t.TempDir(), "fonts.conf")
	require.NoError(t, os.WriteFile(cfg, []byte(`<?xml version="1.0"?>
<fontconfig>
  <dir>`+filepath.Join(data, "fonts")+`</dir>
  <cachedir>`+filepath.Join(t.TempDir(), "cache")+`</cachedir>
</fontconfig>
`), 0o600))
	t.Setenv("FONTCONFIG_FILE", cfg)
	saved := terminalStart
	t.Cleanup(func() { terminalStart = saved })

	desktop := []string{"WAYLAND_DISPLAY=wayland-0", "TERM=xterm-kitty"}
	var note string
	run := func() bool {
		return resolveJapanese(desktop, &japaneseProbes{
			goos:        "linux",
			fonts:       fontconfigState,
			installFont: func() { note = installGlyphFont() },
		})
	}
	font := filepath.Join(data, "fonts", "goanime", glyphFontFile)
	var installed time.Time

	stages := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"no font: glosses hidden and GoAnime Glyphs installed", func(t *testing.T) {
			require.Equal(t, fontMissing, fontconfigState())
			terminalStart = func() (time.Time, bool) { return time.Now().Add(-time.Hour), true }
			require.False(t, run(), "no box is drawn on the first run")
			require.FileExists(t, font)
			require.Contains(t, note, "reopened")
			info, err := os.Stat(font)
			require.NoError(t, err)
			installed = info.ModTime()
		}},
		{"fontconfig finds GoAnime Glyphs", func(t *testing.T) {
			require.NotEqual(t, fontMissing, fontconfigState())
		}},
		{"the same terminal has not loaded it: still hidden, no second install", func(t *testing.T) {
			terminalStart = func() (time.Time, bool) { return installed.Add(-time.Minute), true }
			note = ""
			require.Equal(t, fontPending, fontconfigState())
			require.False(t, run())
			require.Empty(t, note, "nothing installed again")
		}},
		{"a terminal opened after the install shows the glosses", func(t *testing.T) {
			terminalStart = func() (time.Time, bool) { return installed.Add(time.Second), true }
			require.Equal(t, fontPresent, fontconfigState())
			require.True(t, run())
		}},
	}
	for _, stage := range stages {
		if !t.Run(stage.name, stage.run) {
			t.Fatalf("stopped at %q", stage.name)
		}
	}
}
