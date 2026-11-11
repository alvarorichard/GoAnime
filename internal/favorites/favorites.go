// Package favorites keeps the works the user marked as FAVORITOS, like the
// "My List" of the streaming services. They live in one small JSON file next
// to the watch history, so they work on every platform and in builds without
// SQLite (no CGO).
//
// A Store serialises its own calls, but two GoAnime processes on one file do
// not lock each other: the last writer wins, so two favorites toggled in the
// same few milliseconds from two windows can lose one. Within a process use
// Default, the one shared Store.
package favorites

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/tracking"
)

// maxEntries bounds the list; the oldest favorite is dropped past it.
const maxEntries = 500

// schemaVersion is the file layout this build reads and writes. A file from a
// newer build is left alone rather than rewritten without its new fields.
const schemaVersion = 1

// Entry is one favorite work.
type Entry struct {
	// Key identifies the work across sources: its normalized title.
	Key     string    `json:"key"`
	Title   string    `json:"title"`
	Year    string    `json:"year,omitempty"`
	Media   string    `json:"media,omitempty"` // "Anime", "Movie" or "TV"
	Source  string    `json:"source,omitempty"`
	AddedAt time.Time `json:"added_at"`
}

type file struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// Store reads and writes the favorites file.
type Store struct {
	path string
	mu   sync.Mutex
}

// Open returns a store backed by path.
func Open(path string) *Store { return &Store{path: path} }

var (
	defaultOnce  sync.Once
	defaultStore *Store
)

// Default is the store in the user's GoAnime data directory, beside the watch
// history. Without a home directory it has no path and every call fails.
func Default() *Store {
	defaultOnce.Do(func() {
		path := ""
		if db := tracking.DBPath(); db != "" {
			path = filepath.Join(filepath.Dir(db), "favorites.json")
		}
		defaultStore = Open(path)
	})
	return defaultStore
}

var (
	errNoPath = errors.New("favorites: no data directory")
	// ErrNewerFile means favorites.json was written by a newer GoAnime.
	ErrNewerFile = errors.New("favorites: favorites.json was written by a newer GoAnime; update to change it")
)

// reason is err without the file path os errors carry, which is noise (and a
// temp file name) on a screen: "mkdir: permission denied", not
// "mkdir /x/y: permission denied".
func reason(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return fmt.Errorf("%s: %w", pe.Op, pe.Err)
	}
	return err
}

func normalize(key string) string { return strings.ToLower(strings.TrimSpace(key)) }

// load reads the file. A missing file is an empty list. A file that is not
// valid JSON is moved aside (favorites.json.corrupt.<time>) and treated as
// empty, so one bad write never locks the feature; when it cannot be moved
// aside the error says so, so it is never overwritten unseen. Duplicate keys
// (a hand-edited file) collapse to the newest.
func (s *Store) load() ([]Entry, error) {
	if s.path == "" {
		return nil, errNoPath
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("favorites: %w", reason(err))
	}
	f, ok := decode(data)
	if !ok {
		aside := s.path + ".corrupt." + time.Now().Format("20060102-150405")
		if err := os.Rename(s.path, aside); err != nil {
			return nil, fmt.Errorf("favorites: favorites.json is damaged and could not be set aside: %w", reason(err))
		}
		return nil, nil
	}
	if f.Version > schemaVersion {
		return nil, ErrNewerFile
	}
	return dedupe(f.Entries), nil
}

// dedupe keeps one entry per key, the one added last.
func dedupe(entries []Entry) []Entry {
	seen := make(map[string]int, len(entries))
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		e.Key = normalize(e.Key)
		if e.Key == "" {
			continue
		}
		if i, ok := seen[e.Key]; ok {
			if e.AddedAt.After(out[i].AddedAt) {
				out[i] = e
			}
			continue
		}
		seen[e.Key] = len(out)
		out = append(out, e)
	}
	return out
}

// decode parses the file's JSON; ok is false when it is not valid.
func decode(data []byte) (f file, ok bool) {
	if json.Unmarshal(data, &f) != nil {
		return file{}, false
	}
	return f, true
}

// save writes entries atomically: a temp file in the same directory, renamed
// over the real one.
func (s *Store) save(entries []Entry) error {
	if s.path == "" {
		return errNoPath
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("favorites: %w", reason(err))
	}
	data, err := json.MarshalIndent(file{Version: schemaVersion, Entries: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("favorites: %w", reason(err))
	}
	tmp, err := os.CreateTemp(dir, ".favorites-*.tmp")
	if err != nil {
		return fmt.Errorf("favorites: %w", reason(err))
	}
	name := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(name)
		return fmt.Errorf("favorites: %w", errors.Join(werr, cerr))
	}
	if err := os.Rename(name, s.path); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("favorites: %w", reason(err))
	}
	return nil
}

// List returns the favorites, newest first.
func (s *Store) List() ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.load()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].AddedAt.After(entries[j].AddedAt) })
	return entries, nil
}

// Has reports whether the work with key is a favorite. Any failure reads as no.
func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.load()
	if err != nil {
		return false
	}
	key = normalize(key)
	for i := range entries {
		if entries[i].Key == key {
			return true
		}
	}
	return false
}

// Keys returns the key of every favorite, from one read of the file.
func (s *Store) Keys() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.load()
	if err != nil {
		return map[string]bool{}
	}
	keys := make(map[string]bool, len(entries))
	for i := range entries {
		keys[entries[i].Key] = true
	}
	return keys
}

// Toggle adds e when its key is not a favorite and removes it when it is. It
// reports whether the work is a favorite afterwards.
func (s *Store) Toggle(e Entry) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.load()
	if err != nil {
		return false, err
	}
	e.Key = normalize(e.Key)
	if e.Key == "" {
		return false, errors.New("favorites: a work needs a key")
	}
	for i := range entries {
		if entries[i].Key == e.Key {
			entries = append(entries[:i], entries[i+1:]...)
			return false, s.save(entries)
		}
	}
	if e.AddedAt.IsZero() {
		e.AddedAt = time.Now()
	}
	entries = append(entries, e)
	if len(entries) > maxEntries {
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].AddedAt.After(entries[j].AddedAt) })
		entries = entries[:maxEntries]
	}
	return true, s.save(entries)
}

// Remove drops the favorite with key. A key that is not there is not an error.
func (s *Store) Remove(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.load()
	if err != nil {
		return err
	}
	key = normalize(key)
	kept := entries[:0]
	for _, e := range entries {
		if e.Key != key {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(entries) {
		return nil
	}
	return s.save(kept)
}

// Clear removes every favorite.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.load(); err != nil {
		return err
	}
	return s.save(nil)
}
