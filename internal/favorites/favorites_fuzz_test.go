package favorites

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzStoreFile feeds arbitrary bytes in as favorites.json. Whatever is in the
// file, no call may panic and the store must stay usable: a work can always be
// added, seen and removed afterwards.
func FuzzStoreFile(f *testing.F) {
	for _, seed := range []string{
		``, `{}`, `null`, `[]`, `{"version":1,"entries":null}`, `{"version":1,"entries":[]}`,
		`{"version":1,"entries":[{"key":"fz","title":"x","added_at":"2024-01-01T00:00:00Z"}]}`,
		`{"version":1,"entries":[{"key":"FZ ","title":"x"},{"key":"fz"},{"key":"fz"}]}`,
		`{"entries":[{"key":"","title":""}]}`, `{"entries":[1,"a",null,[]]}`,
		`{"version":1,"entries":[{"key":"a","added_at":"not a date"}]}`,
		`{"version":1,"entries":[{"key":"a","added_at":"9999-12-31T23:59:59Z"}]}`,
		"\x00\x01\x02", "{\"entries\":[{\"key\":\"\xff\xfe\"}]}", `{"version":`, "\xef\xbb\xbf{}",
		`{"version":1e999,"entries":[]}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		path := filepath.Join(t.TempDir(), "favorites.json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Skip()
		}
		s := Open(path)

		if _, err := s.List(); err != nil {
			t.Fatalf("List on a readable file must not fail: %v", err)
		}
		s.Has("fz")
		s.Has("")
		// Put the store in a known state and prove it still works.
		if err := s.Remove("fz"); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		if s.Has("fz") {
			t.Fatal("Remove must drop every copy of the key")
		}
		on, err := s.Toggle(Entry{Key: " FZ ", Title: "fuzz"})
		if err != nil {
			t.Fatalf("Toggle on a fuzzed file: %v", err)
		}
		if !on {
			t.Fatal("after Remove, Toggle must add; it reported removed")
		}
		list, err := s.List()
		if err != nil {
			t.Fatalf("List after Toggle: %v", err)
		}
		// A file full of future-dated entries can push the new one out of a
		// full list; otherwise what Toggle said is what Has must say.
		if len(list) < maxEntries && !s.Has("fz") {
			t.Fatal("Toggle said added but Has says no")
		}
		if len(list) < maxEntries {
			if off, err := s.Toggle(Entry{Key: "fz"}); err != nil || off {
				t.Fatalf("second Toggle: on=%v err=%v", off, err)
			}
		}
		if err := s.Clear(); err != nil {
			t.Fatalf("Clear: %v", err)
		}
		if list, err := s.List(); err != nil || len(list) != 0 {
			t.Fatalf("after Clear: %v %v", list, err)
		}
		// the file on disk is valid for the next run
		if _, err := Open(path).Toggle(Entry{Key: "next"}); err != nil {
			t.Fatalf("next run: %v", err)
		}
	})
}
