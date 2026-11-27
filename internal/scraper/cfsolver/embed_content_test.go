package cfsolver

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"
)

// A bundled binary is kept only when the file already unpacked holds exactly
// its bytes. Size alone is not identity: two cfsolve builds can have the same
// size, and the stale one would be refused by this GoAnime's profile.

// contentCase is one file on disk (or none) compared with the bundled bytes.
type contentCase struct {
	name string
	file []byte // nil: no file at the path
	dir  bool   // the path is a directory
	data []byte
	want bool
}

// multiChunk is larger than compareChunk and not a multiple of it, so the
// comparison has full chunks and a short last one.
var multiChunk = func() []byte {
	b := make([]byte, 2*compareChunk+123)
	for i := range b {
		b[i] = byte(i*7 + i/251)
	}
	return b
}()

// oneByteTail ends in a last chunk of a single byte, the one a loop that stops
// one byte early never compares.
var oneByteTail = multiChunk[:compareChunk+1]

func flipped(b []byte, at int) []byte {
	out := bytes.Clone(b)
	out[at] ^= 0xff
	return out
}

func contentCases() []contentCase {
	small := []byte("binary-v1")
	return []contentCase{
		{name: "identical", file: small, data: small, want: true},
		{name: "no file", data: small, want: false},
		{name: "a directory", dir: true, data: small, want: false},
		{name: "same size, last byte differs", file: []byte("binary-v2"), data: small, want: false},
		{name: "same size, first byte differs", file: []byte("Xinary-v1"), data: small, want: false},
		{name: "shorter file", file: []byte("binary-v"), data: small, want: false},
		{name: "longer file", file: []byte("binary-v1!"), data: small, want: false},
		{name: "empty file", file: []byte{}, data: small, want: false},
		{name: "identical, several chunks", file: multiChunk, data: multiChunk, want: true},
		{name: "same size, differs inside the second chunk", file: flipped(multiChunk, compareChunk+7), data: multiChunk, want: false},
		{name: "same size, differs at the second chunk's first byte", file: flipped(multiChunk, compareChunk), data: multiChunk, want: false},
		{name: "same size, differs at the last chunk's first byte", file: flipped(multiChunk, 2*compareChunk), data: multiChunk, want: false},
		{name: "same size, differs at the very end", file: flipped(multiChunk, len(multiChunk)-1), data: multiChunk, want: false},
		{name: "identical, one-byte last chunk", file: oneByteTail, data: oneByteTail, want: true},
		{name: "same size, the one-byte last chunk differs", file: flipped(oneByteTail, len(oneByteTail)-1), data: oneByteTail, want: false},
	}
}

// placeContent puts the case's file (or directory) at a fresh path.
func placeContent(t *testing.T, c contentCase) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "challenge")
	switch {
	case c.dir:
		if err := os.Mkdir(path, 0o750); err != nil {
			t.Fatal(err)
		}
	case c.file != nil:
		if err := os.WriteFile(path, c.file, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

// sameContentFailures runs every case through same and lists the ones it gets
// wrong. The real sameFileContent must give none.
func sameContentFailures(t *testing.T, same func(path string, data []byte) bool) []string {
	t.Helper()
	var bad []string
	for _, c := range contentCases() {
		if got := same(placeContent(t, c), c.data); got != c.want {
			bad = append(bad, fmt.Sprintf("%s: got %v, want %v", c.name, got, c.want))
		}
	}
	return bad
}

func TestSameFileContent(t *testing.T) {
	for _, c := range contentCases() {
		t.Run(c.name, func(t *testing.T) {
			if got := sameFileContent(placeContent(t, c), c.data); got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

// The comparison runs on every start against a binary of up to 40 MB, so it
// must read the file a chunk at a time rather than load a second copy.
func TestSameFileContentReadsInChunks(t *testing.T) {
	data := make([]byte, 32*compareChunk) // 8 MiB
	path := placeContent(t, contentCase{file: data})
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	same := sameFileContent(path, data)
	runtime.ReadMemStats(&after)
	if !same {
		t.Fatal("an identical file was reported different")
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 4*compareChunk {
		t.Fatalf("allocated %d bytes to compare %d; want at most %d", grew, len(data), 4*compareChunk)
	}
}

// sizeOnlySame is the check unpackEmbedded made before: a regular file of the
// same size counted as the same binary.
func sizeOnlySame(path string, data []byte) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular() && st.Size() == int64(len(data))
}

func TestSameFileContentMetatest(t *testing.T) {
	if bad := sameContentFailures(t, sameFileContent); len(bad) != 0 {
		t.Fatalf("the real comparison breaks the contract: %v", bad)
	}
	readAll := func(path string) ([]byte, bool) {
		b, err := os.ReadFile(path) // #nosec G304 -- a file this test wrote
		return b, err == nil
	}
	mutants := map[string]func(string, []byte) bool{
		"size only (the old check)": sizeOnlySame,
		"always the same":           func(string, []byte) bool { return true },
		"never the same":            func(string, []byte) bool { return false },
		"first chunk only": func(path string, data []byte) bool {
			b, ok := readAll(path)
			n := min(compareChunk, len(data))
			return ok && len(b) == len(data) && bytes.Equal(b[:n], data[:n])
		},
		"size ignored, prefix compared": func(path string, data []byte) bool {
			f, err := os.Open(path) // #nosec G304 -- a file this test wrote
			if err != nil {
				return false
			}
			defer func() { _ = f.Close() }()
			b := make([]byte, len(data))
			_, err = io.ReadFull(f, b)
			return err == nil && bytes.Equal(b, data)
		},
		"last short chunk skipped": func(path string, data []byte) bool {
			b, ok := readAll(path)
			full := len(data) / compareChunk * compareChunk
			if len(data) < compareChunk {
				full = len(data) - 1
			}
			return ok && len(b) == len(data) && bytes.Equal(b[:full], data[:full])
		},
		"stops one byte early": func(path string, data []byte) bool {
			b, ok := readAll(path)
			if !ok || len(b) != len(data) {
				return false
			}
			for off := 0; off < len(data)-1; off += compareChunk {
				end := min(off+compareChunk, len(data))
				if !bytes.Equal(b[off:end], data[off:end]) {
					return false
				}
			}
			return true
		},
		"chunks compared at the wrong offset": func(path string, data []byte) bool {
			b, ok := readAll(path)
			if !ok || len(b) != len(data) {
				return false
			}
			for off := 0; off < len(data); off += compareChunk {
				end := min(off+compareChunk, len(data))
				if !bytes.Equal(b[off+1:end], data[off+1:end]) {
					return false
				}
			}
			return true
		},
	}
	for name, m := range mutants {
		if bad := sameContentFailures(t, m); len(bad) == 0 {
			t.Errorf("mutant %q passes the contract", name)
		}
	}
}

// unpackCase is a file already at the destination when unpackEmbedded runs.
type unpackCase struct {
	name        string
	existing    []byte // nil: nothing there yet
	wantReplace bool   // the file must be written anew (a new file, not the old one)
}

func unpackCases(bundled []byte) []unpackCase {
	return []unpackCase{
		{name: "nothing there", wantReplace: true},
		{name: "the same binary", existing: bundled, wantReplace: false},
		{name: "another build of the same size", existing: flipped(bundled, len(bundled)/2), wantReplace: true},
		{name: "another build of another size", existing: bundled[:len(bundled)-1], wantReplace: true},
	}
}

type unpackFunc func(fsys fs.FS, root, goos, goarch, dir string) (string, error)

// oldStamp marks a file that was already there: a rewritten file is a new one
// with a fresh modification time, so a file still carrying it was kept. (Inode
// numbers would not do: a filesystem may hand a freed one to the next file.)
var oldStamp = time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)

// unpackFailures runs every case through unpack and lists what it gets wrong:
// the path, the bytes left on disk, and whether the file was rewritten.
func unpackFailures(t *testing.T, unpack unpackFunc) []string {
	t.Helper()
	bundled := multiChunk
	fsys := fstest.MapFS{"embedded/challenge-linux-amd64": {Data: bundled}}
	var bad []string
	for _, c := range unpackCases(bundled) {
		dir := t.TempDir()
		dst := filepath.Join(dir, "challenge")
		if c.existing != nil {
			if err := os.WriteFile(dst, c.existing, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(dst, oldStamp, oldStamp); err != nil {
				t.Fatal(err)
			}
		}
		p, err := unpack(fsys, "embedded", "linux", "amd64", dir)
		if err != nil || p != dst {
			bad = append(bad, fmt.Sprintf("%s: got %q, %v", c.name, p, err))
			continue
		}
		if b, _ := os.ReadFile(dst); !bytes.Equal(b, bundled) { // #nosec G304 -- a file this test wrote
			bad = append(bad, c.name+": the file does not hold the bundled binary")
		}
		after, err := os.Stat(dst)
		if err != nil {
			bad = append(bad, c.name+": "+err.Error())
			continue
		}
		if replaced := !after.ModTime().Equal(oldStamp); replaced != c.wantReplace {
			bad = append(bad, fmt.Sprintf("%s: replaced=%v, want %v", c.name, replaced, c.wantReplace))
		}
	}
	return bad
}

func TestUnpackEmbeddedReplacesOnlyAnotherBinary(t *testing.T) {
	if bad := unpackFailures(t, unpackEmbedded); len(bad) != 0 {
		t.Fatal(bad)
	}
}

func TestUnpackEmbeddedMetatest(t *testing.T) {
	existing := func(dir string) (os.FileInfo, bool) {
		st, err := os.Stat(filepath.Join(dir, "challenge"))
		return st, err == nil
	}
	mutants := map[string]unpackFunc{
		// What 2.0.4 shipped with at first: a same-size stale build is kept.
		"keeps a file of the same size": func(fsys fs.FS, root, goos, goarch, dir string) (string, error) {
			data, _ := fs.ReadFile(fsys, root+"/challenge-"+goos+"-"+goarch)
			if st, ok := existing(dir); ok && st.Size() == int64(len(data)) {
				return filepath.Join(dir, "challenge"), nil
			}
			return unpackEmbedded(fsys, root, goos, goarch, dir)
		},
		"keeps any file": func(fsys fs.FS, root, goos, goarch, dir string) (string, error) {
			if _, ok := existing(dir); ok {
				return filepath.Join(dir, "challenge"), nil
			}
			return unpackEmbedded(fsys, root, goos, goarch, dir)
		},
		"rewrites every time": func(fsys fs.FS, root, goos, goarch, dir string) (string, error) {
			_ = os.Remove(filepath.Join(dir, "challenge"))
			return unpackEmbedded(fsys, root, goos, goarch, dir)
		},
	}
	for name, m := range mutants {
		if bad := unpackFailures(t, m); len(bad) == 0 {
			t.Errorf("mutant %q passes the contract", name)
		}
	}
}
