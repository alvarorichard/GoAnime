package cfsolver

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// A release build carries the cfsolve binary for its own platform inside the
// GoAnime executable, so the tool needs no download and works offline. The
// release and CI workflows build cfsolve's latest commit, take that commit's
// profile and copy the binary into embedded/ under its release asset name
// before "go build"; a development build finds the folder empty and keeps
// looking for an installed binary or downloading one.
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

// unpackEmbedded writes the platform's binary from fsys into dir, unless the
// file already there holds exactly those bytes (it is only ever written whole,
// by a rename). A file of the same size is not enough: two cfsolve builds can
// have the same size, and a stale one (an older GoAnime's bundle, a download,
// a local build) is refused by this GoAnime's profile. It returns "" with no
// error when fsys holds no binary for the platform.
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
	if sameFileContent(dst, data) {
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

// compareChunk is how much of the file sameFileContent reads at a time.
const compareChunk = 256 << 10

// sameFileContent reports whether path is a regular file holding exactly data.
// It runs on every start, so it reads the file in chunks (no second copy of a
// 40 MB binary in memory) and stops at the first difference; a file of another
// size is not read at all. On an SSD the whole comparison of a 20 MB binary
// takes a few milliseconds.
func sameFileContent(path string, data []byte) bool {
	f, err := os.Open(path) // #nosec G304 -- GoAnime's own config folder
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() != int64(len(data)) {
		return false
	}
	buf := make([]byte, min(compareChunk, len(data)))
	for off := 0; off < len(data); {
		n, err := io.ReadFull(f, buf[:min(len(buf), len(data)-off)])
		if err != nil || !bytes.Equal(buf[:n], data[off:off+n]) {
			return false
		}
		off += n
	}
	return true
}
