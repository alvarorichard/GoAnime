package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The saved source choices are process-global, so these tests do not run in
// parallel and each one leaves the table empty.
func cleanPrefs(t *testing.T) {
	t.Helper()
	ResetSourcePrefs()
	t.Cleanup(ResetSourcePrefs)
}

func TestSourcePrefSetClearReset(t *testing.T) {
	cleanPrefs(t)

	_, set := SourcePref("Goyabu")
	assert.False(t, set, "nothing chosen yet")

	SetSourcePref("Goyabu", true)
	on, set := SourcePref("goyabu")
	assert.True(t, on)
	assert.True(t, set, "lookups are forgiving about case")

	SetSourcePref("Animefire.io", false)
	on, set = SourcePref("animefire")
	assert.False(t, on)
	assert.True(t, set, "an explicit off is a choice, not an absence")

	ClearSourcePref("GOYABU")
	_, set = SourcePref("Goyabu")
	assert.False(t, set, "cleared goes back to the default")
	_, set = SourcePref("Animefire.io")
	assert.True(t, set, "clearing one leaves the others")

	ResetSourcePrefs()
	_, set = SourcePref("Animefire.io")
	assert.False(t, set)
}

func TestSourcePrefIgnoresEmptyNames(t *testing.T) {
	cleanPrefs(t)
	SetSourcePref("  ", true)
	SetSourcePref("", true)
	_, set := SourcePref("")
	assert.False(t, set)
	ClearSourcePref("") // must not panic
	path := filepath.Join(t.TempDir(), "sources.json")
	require.NoError(t, SaveSourcePrefs(path))
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"":`, "an empty name never reaches the file")
}

func TestSourcePrefsPath(t *testing.T) {
	old := userConfigDir
	t.Cleanup(func() { userConfigDir = old })

	userConfigDir = func() (string, error) { return "cfg", nil }
	got, err := SourcePrefsPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("cfg", "goanime", "sources.json"), got)

	userConfigDir = func() (string, error) { return "", errors.New("no home") }
	_, err = SourcePrefsPath()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no config folder")
}

func TestSaveLoadRoundTrip(t *testing.T) {
	cleanPrefs(t)
	path := filepath.Join(t.TempDir(), "nested", "goanime", "sources.json")

	SetSourcePref("Goyabu", true)
	SetSourcePref("HiAnime", false)
	require.NoError(t, SaveSourcePrefs(path), "the folder is created on demand")

	ResetSourcePrefs()
	require.NoError(t, LoadSourcePrefs(path))
	on, set := SourcePref("Goyabu")
	assert.True(t, on && set)
	on, set = SourcePref("HiAnime")
	assert.False(t, on)
	assert.True(t, set)

	if info, err := os.Stat(path); err == nil && filepath.Separator == '/' {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "the file is the user's alone")
	}
}

func TestLoadSourcePrefsMissingFileIsFine(t *testing.T) {
	cleanPrefs(t)
	SetSourcePref("Goyabu", true)
	require.NoError(t, LoadSourcePrefs(filepath.Join(t.TempDir(), "absent.json")))
	_, set := SourcePref("Goyabu")
	assert.True(t, set, "a missing file changes nothing")
}

func TestLoadSourcePrefsRejectsBadFilesWithoutTouchingChoices(t *testing.T) {
	cases := map[string]string{
		"not json":       "{nope",
		"newer version":  `{"version": 99, "sources": {"goyabu": false}}`,
		"wrong type":     `{"version": 1, "sources": {"goyabu": "yes"}}`,
		"sources a list": `{"version": 1, "sources": ["goyabu"]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			cleanPrefs(t)
			SetSourcePref("HiAnime", true)
			path := filepath.Join(t.TempDir(), "sources.json")
			require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

			err := LoadSourcePrefs(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "source preferences")
			on, set := SourcePref("HiAnime")
			assert.True(t, on && set, "a bad file must not flip any source")
			_, set = SourcePref("Goyabu")
			assert.False(t, set)
		})
	}
}

func TestLoadSourcePrefsUnreadableFile(t *testing.T) {
	cleanPrefs(t)
	dir := t.TempDir() // a directory where a file is expected
	err := LoadSourcePrefs(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestLoadSourcePrefsDropsEmptyNamesAndNormalises(t *testing.T) {
	cleanPrefs(t)
	path := filepath.Join(t.TempDir(), "sources.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"version":1,"sources":{"Animefire.io":false," ":true}}`), 0o600))
	require.NoError(t, LoadSourcePrefs(path))
	on, set := SourcePref("animefire")
	assert.False(t, on)
	assert.True(t, set)
	_, set = SourcePref(" ")
	assert.False(t, set)

	// An empty name can't be looked up, so the proof it was dropped is the
	// next save: it must not write an empty key.
	out := filepath.Join(t.TempDir(), "again.json")
	require.NoError(t, SaveSourcePrefs(out))
	raw, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"":`)
	assert.Contains(t, string(raw), `"animefire":false`)
}

func TestLoadSavedSourcePrefsUsesTheConfigFolder(t *testing.T) {
	cleanPrefs(t)
	old := userConfigDir
	t.Cleanup(func() { userConfigDir = old })

	cfg := t.TempDir()
	userConfigDir = func() (string, error) { return cfg, nil }
	require.NoError(t, os.MkdirAll(filepath.Join(cfg, "goanime"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(cfg, "goanime", "sources.json"), []byte(`{"version":1,"sources":{"goyabu":true}}`), 0o600))

	require.NoError(t, LoadSavedSourcePrefs())
	on, set := SourcePref("goyabu")
	assert.True(t, on && set)

	userConfigDir = func() (string, error) { return "", errors.New("no home") }
	require.Error(t, LoadSavedSourcePrefs())
}

func TestPersistSourcePrefSavesAndKeepsIt(t *testing.T) {
	cleanPrefs(t)
	path := filepath.Join(t.TempDir(), "sources.json")

	require.NoError(t, PersistSourcePref(path, "Goyabu", true))
	ResetSourcePrefs()
	require.NoError(t, LoadSourcePrefs(path))
	on, set := SourcePref("Goyabu")
	assert.True(t, on && set, "what Persist returned ok for is on disk")
}

// A save that fails must leave memory as it was, or the screen would show a
// choice the next start forgets.
func TestPersistSourcePrefRollsBackOnFailure(t *testing.T) {
	old := writeSourcePrefs
	t.Cleanup(func() { writeSourcePrefs = old })
	writeSourcePrefs = func(string, []byte) error { return errors.New("disk full") }
	path := filepath.Join(t.TempDir(), "sources.json")

	t.Run("no previous choice: back to the default", func(t *testing.T) {
		cleanPrefs(t)
		err := PersistSourcePref(path, "Goyabu", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "disk full")
		_, set := SourcePref("Goyabu")
		assert.False(t, set)
	})

	t.Run("previous choice restored", func(t *testing.T) {
		cleanPrefs(t)
		SetSourcePref("Goyabu", false)
		require.Error(t, PersistSourcePref(path, "Goyabu", true))
		on, set := SourcePref("Goyabu")
		assert.False(t, on)
		assert.True(t, set, "the old explicit off survives a failed flip")
	})
}

func TestSaveSourcePrefsReportsWriteFailure(t *testing.T) {
	cleanPrefs(t)
	old := writeSourcePrefs
	t.Cleanup(func() { writeSourcePrefs = old })
	var gotPath string
	var gotData []byte
	writeSourcePrefs = func(p string, d []byte) error { gotPath, gotData = p, d; return errors.New("denied") }
	SetSourcePref("Goyabu", true)

	err := SaveSourcePrefs("/x/sources.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/x/sources.json")
	assert.Equal(t, "/x/sources.json", gotPath)
	assert.Contains(t, string(gotData), `"goyabu"`)
	assert.Contains(t, string(gotData), `"version"`)
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "f.json")
	require.NoError(t, writeFileAtomic(path, []byte("one")))
	require.NoError(t, writeFileAtomic(path, []byte("two")), "overwrites in place")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "two", string(raw))

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1, "no temp file is left behind")

	t.Run("folder cannot be created", func(t *testing.T) {
		file := filepath.Join(dir, "plain")
		require.NoError(t, os.WriteFile(file, nil, 0o600))
		require.Error(t, writeFileAtomic(filepath.Join(file, "sub", "f.json"), []byte("x")))
	})

	t.Run("target is a directory: rename fails and the temp is cleaned", func(t *testing.T) {
		target := filepath.Join(dir, "isdir")
		require.NoError(t, os.MkdirAll(filepath.Join(target, "child"), 0o750))
		require.Error(t, writeFileAtomic(target, []byte("x")))
		left, err := filepath.Glob(filepath.Join(dir, ".sources-*.tmp"))
		require.NoError(t, err)
		assert.Empty(t, left)
	})
}

func TestSourcePrefsAreRaceFree(t *testing.T) {
	cleanPrefs(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := "s" + strings.Repeat("x", i%3)
			SetSourcePref(name, i%2 == 0)
			SourcePref(name)
			_ = SaveSourcePrefs(filepath.Join(t.TempDir(), "s.json"))
			ClearSourcePref(name)
		}()
	}
	wg.Wait()
}
