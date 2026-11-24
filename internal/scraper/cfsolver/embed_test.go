package cfsolver

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestUnpackEmbedded(t *testing.T) {
	fsys := fstest.MapFS{"embedded/challenge-linux-amd64": {Data: []byte("binary-v1")}}
	dir := t.TempDir()

	p, err := unpackEmbedded(fsys, "embedded", "linux", "amd64", dir)
	if err != nil || p != filepath.Join(dir, "challenge") {
		t.Fatalf("got %q, %v", p, err)
	}
	if b, _ := os.ReadFile(p); string(b) != "binary-v1" {
		t.Fatalf("content %q", b)
	}
	if st, _ := os.Stat(p); st.Mode().Perm()&0o100 == 0 {
		t.Fatalf("not executable: %v", st.Mode())
	}

	// A stale file of another size is replaced by the bundled one.
	if err := os.WriteFile(p, []byte("old"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := unpackEmbedded(fsys, "embedded", "linux", "amd64", dir); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); string(b) != "binary-v1" {
		t.Fatalf("not refreshed: %q", b)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".challenge-*")); len(left) != 0 {
		t.Fatalf("temp files left: %v", left)
	}
}

func TestUnpackEmbeddedAbsent(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct{ goos, goarch string }{
		{"linux", "amd64"},   // nothing bundled
		{"windows", "arm64"}, // no asset exists for this platform
	} {
		p, err := unpackEmbedded(fstest.MapFS{}, "embedded", c.goos, c.goarch, dir)
		if p != "" || err != nil {
			t.Fatalf("%s/%s: got %q, %v", c.goos, c.goarch, p, err)
		}
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Fatalf("wrote files for an absent binary: %v", ents)
	}
}

func TestUnpackEmbeddedWindowsName(t *testing.T) {
	fsys := fstest.MapFS{"embedded/challenge-windows-amd64.exe": {Data: []byte("x")}}
	dir := t.TempDir()
	p, err := unpackEmbedded(fsys, "embedded", "windows", "amd64", dir)
	if err != nil || filepath.Base(p) != "challenge.exe" {
		t.Fatalf("got %q, %v", p, err)
	}
}
