package offlinecache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMarkThenIsOffline(t *testing.T) {
	c := New("", time.Hour)
	if c.IsOffline("AnimeFire", "u1") {
		t.Fatal("unrecorded title reported offline")
	}
	if err := c.MarkOffline("AnimeFire", "u1"); err != nil {
		t.Fatal(err)
	}
	if !c.IsOffline("animefire", "u1") {
		t.Fatal("recorded title not offline (source match is case-insensitive)")
	}
	if c.IsOffline("Goyabu", "u1") {
		t.Fatal("verdict leaked to another source")
	}
}

func TestExpiry(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	c := New("", time.Hour)
	c.now = func() time.Time { return now }
	_ = c.MarkOffline("s", "u")
	now = now.Add(time.Hour + time.Second)
	if c.IsOffline("s", "u") {
		t.Fatal("expired entry still reported offline")
	}
}

func TestPersistsAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", fileName)
	if err := New(path, time.Hour).MarkOffline("s", "u"); err != nil {
		t.Fatal(err)
	}
	if !New(path, time.Hour).IsOffline("s", "u") {
		t.Fatal("record not read back from disk")
	}
}

func TestTwoInstancesDoNotEraseEachOther(t *testing.T) {
	path := filepath.Join(t.TempDir(), fileName)
	a, b := New(path, time.Hour), New(path, time.Hour)
	_ = a.IsOffline("s", "x") // a loads an empty file first
	_ = b.MarkOffline("s", "from-b")
	_ = a.MarkOffline("s", "from-a")
	c := New(path, time.Hour)
	if !c.IsOffline("s", "from-a") || !c.IsOffline("s", "from-b") {
		t.Fatal("a save dropped the other instance's entry")
	}
}

func TestCorruptFileIsReplaced(t *testing.T) {
	path := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := New(path, time.Hour)
	if c.IsOffline("s", "u") {
		t.Fatal("corrupt file reported a title offline")
	}
	if err := c.MarkOffline("s", "u"); err != nil {
		t.Fatal(err)
	}
	if !New(path, time.Hour).IsOffline("s", "u") {
		t.Fatal("corrupt file not replaced by a good one")
	}
}

func TestUnwritablePathKeepsMemory(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	c := New(filepath.Join(blocker, "x", fileName), time.Hour) // parent is a file
	if err := c.MarkOffline("s", "u"); err == nil {
		t.Fatal("expected a save error")
	}
	if !c.IsOffline("s", "u") {
		t.Fatal("record lost after a failed save")
	}
}

func TestHidingEnabled(t *testing.T) {
	for v, want := range map[string]bool{"": true, "1": true, "0": false, "off": false, "False": false, "no": false} {
		t.Setenv(EnvHide, v)
		if got := HidingEnabled(); got != want {
			t.Errorf("%s=%q: got %v, want %v", EnvHide, v, got, want)
		}
	}
}
