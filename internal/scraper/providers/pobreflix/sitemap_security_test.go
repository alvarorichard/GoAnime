package pobreflix

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCorpusCacheCreatesPrivateFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "index.json")
	e, ok := entryFromURL("https://x.test/filmes/online/a-film-dublado-1/")
	require.True(t, ok)
	cc := corpus{cachePath: path, entries: []entry{e}, fetchedAt: time.Now()}
	cc.saveCache()
	info, err := os.Stat(path)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Zero(t, info.Mode().Perm()&0o077, "the cache must be private to its owner")
	}
	loaded := corpus{cachePath: path}
	loaded.loadCache("https://x.test")
	require.Equal(t, []entry{e}, loaded.entries)
}

func TestCorpusCacheRejectsEscapingSymlink(t *testing.T) {
	t.Parallel()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside.json")
	outsideRoot, err := os.OpenRoot(outsideDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = outsideRoot.Close() })
	raw := []byte(`{"entries":["https://x.test/filmes/online/a-film-dublado-1/"]}`)
	require.NoError(t, os.WriteFile(outside, raw, 0o600))
	path := filepath.Join(t.TempDir(), "index.json")
	if err := os.Symlink(outside, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cc := corpus{cachePath: path}
	cc.loadCache("https://x.test")
	require.Empty(t, cc.entries, "a cache cannot read through a link outside its directory")
	e, ok := entryFromURL("https://x.test/filmes/online/another-film-dublado-2/")
	require.True(t, ok)
	cc.entries = []entry{e}
	cc.saveCache()
	got, err := outsideRoot.ReadFile("outside.json")
	require.NoError(t, err)
	require.Equal(t, raw, got, "a cache cannot overwrite a link target outside its directory")
}
