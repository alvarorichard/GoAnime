package cfsolver

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// A release build carries the cfsolve binary for its own platform inside the
// GoAnime executable, so the tool needs no download and works offline. The
// release workflow copies the pinned build (cfsolve.lock) into embedded/ under
// its release asset name before "go build"; a development build finds the
// folder empty and keeps looking for an installed binary or downloading one.
// The binary is unpacked on first use into GoAnime's config folder.

//go:embed all:embedded
var embeddedFS embed.FS

// EnsureEmbedded unpacks the bundled cfsolve into GoAnime's config folder and
// returns its path. It reports false when this build carries none or the
// folder cannot be written; callers then fall back to the other ways of
// finding the tool.
func EnsureEmbedded() (string, bool) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}
	path, err := unpackEmbedded(embeddedFS, "embedded", runtime.GOOS, runtime.GOARCH, filepath.Join(cfg, "goanime", "cfsolve"))
	if err != nil {
		util.Debug("cfsolve: bundled binary not unpacked", "err", err)
		return "", false
	}
	return path, path != ""
}

// unpackEmbedded writes the platform's binary from fsys into dir, unless a
// file of the same size is already there (it is only ever written whole, by a
// rename). It returns "" with no error when fsys holds no binary for the
// platform.
func unpackEmbedded(fsys fs.FS, root, goos, goarch, dir string) (string, error) {
	asset, ok := assetName(goos, goarch)
	if !ok {
		return "", nil
	}
	data, err := fs.ReadFile(fsys, root+"/"+asset)
	if err != nil {
		return "", nil // not bundled in this build
	}
	if len(data) == 0 {
		return "", nil
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, localBinaryName(goos))
	if st, err := os.Stat(dst); err == nil && st.Mode().IsRegular() && st.Size() == int64(len(data)) {
		return dst, nil
	}
	tmp, err := os.CreateTemp(dir, ".challenge-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("writing the bundled cfsolve: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpName, 0o700); err != nil { // #nosec G302 -- an executable the user's own account runs
		return "", err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return "", err
	}
	return dst, nil
}
