package offlinecache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMarkAndExpire(t *testing.T) {
	c := New("", time.Hour)
	now := time.Unix(1_000_000, 0)
	c.now = func() time.Time { return now }

	if c.IsOffline("animefire", "u1") {
		t.Fatal("unknown title must not be offline")
	}
	c.MarkOffline("animefire", "u1")
	if !c.IsOffline("animefire", "u1") {
		t.Fatal("marked title should be offline")
	}
	if c.IsOffline("goyabu", "u1") {
		t.Fatal("verdict must be per source")
	}

	now = now.Add(time.Hour + time.Second)
	if c.IsOffline("animefire", "u1") {
		t.Fatal("verdict should expire after the TTL")
	}
}

func TestPersistsAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", fileName)
	a := New(path, time.Hour)
	a.MarkOffline("animefire", "u1")
	if err := a.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	b := New(path, time.Hour)
	if !b.IsOffline("animefire", "u1") {
		t.Fatal("a new instance should read the saved verdict")
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("cache file should be 0600, got %v (%v)", info, err)
	}
}

func TestSaveDropsExpiredEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), fileName)
	a := New(path, time.Hour)
	now := time.Unix(1_000_000, 0)
	a.now = func() time.Time { return now }
	a.MarkOffline("animefire", "old")
	now = now.Add(2 * time.Hour)
	a.MarkOffline("animefire", "new")
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}

	b := New(path, time.Hour)
	b.now = func() time.Time { return now }
	if b.IsOffline("animefire", "old") {
		t.Fatal("expired entry should not survive a save")
	}
	if !b.IsOffline("animefire", "new") {
		t.Fatal("fresh entry should survive")
	}
}

func TestCorruptFileIsEmptyCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := New(path, time.Hour)
	if c.IsOffline("animefire", "u1") {
		t.Fatal("corrupt cache must read as empty")
	}
	c.MarkOffline("animefire", "u1")
	if err := c.Save(); err != nil {
		t.Fatalf("save over corrupt file: %v", err)
	}
}

func TestZeroTTLDisables(t *testing.T) {
	c := New("", 0)
	c.MarkOffline("animefire", "u1")
	if c.IsOffline("animefire", "u1") || c.Enabled() {
		t.Fatal("ttl <= 0 must disable the cache")
	}
}
