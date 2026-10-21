//go:build unix

package pobreflix

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCorpusCacheTightensExistingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "index.json")
	require.NoError(t, os.WriteFile(path, make([]byte, 1024), 0o600))
	// Reproduce the permissions of an older cache independently of the umask.
	require.NoError(t, os.Chmod(path, 0o644))
	before, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o644), before.Mode().Perm())

	e, ok := entryFromURL("https://x.test/filmes/online/a-film-dublado-1/")
	require.True(t, ok)
	cc := corpus{cachePath: path, entries: []entry{e}, fetchedAt: time.Now()}
	cc.saveCache()

	after, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), after.Mode().Perm(), "saving must tighten permissions of an existing cache")
	loaded := corpus{cachePath: path}
	loaded.loadCache("https://x.test")
	require.Equal(t, []entry{e}, loaded.entries)
}
