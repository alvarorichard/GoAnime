//go:build windows

package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortableWindowsTerminalPath(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	_, ok := portableWindowsTerminalPath()
	assert.False(t, ok, "no copy exists before the first download")

	wt := filepath.Join(local, "GoAnime", "WindowsTerminal", "current", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(wt), 0o755))
	require.NoError(t, os.WriteFile(wt, []byte("wt"), 0o600))

	got, ok := portableWindowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, wt, got)

	got, ok = windowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, wt, got, "the private copy is used when the app alias is missing")
}

func TestInstalledAliasWinsOverPortableCopy(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	alias := filepath.Join(local, "Microsoft", "WindowsApps", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o755))
	require.NoError(t, os.WriteFile(alias, []byte("wt"), 0o600))

	portable := filepath.Join(local, "GoAnime", "WindowsTerminal", "current", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(portable), 0o755))
	require.NoError(t, os.WriteFile(portable, []byte("wt"), 0o600))

	got, ok := windowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, alias, got, "a real installation takes precedence over the fallback copy")
}

func TestPortableWindowsTerminalRootRequiresLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")

	_, err := portableWindowsTerminalRoot()
	require.Error(t, err)
}
