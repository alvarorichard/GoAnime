// Package favorites keeps the works the user marked as FAVORITOS, like the
// "My List" of the streaming services. They live in one small JSON file next
// to the watch history, so they work on every platform and in builds without
// SQLite (no CGO).
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

var errNoPath = errors.New("favorites: no data directory")

func normalize(key string) string { return strings.ToLower(strings.TrimSpace(key)) }

// load reads the file. A missing file is an empty list. A file that is not
// valid JSON is moved aside (favorites.json.corrupt) and treated as empty, so
// one bad write never locks the feature.
func (s *Store) load() ([]Entry, error) {
	if s.path == "" {
		return nil, errNoPath
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("favorites: %w", err)
	}
	entries, ok := decode(data)
	if !ok {
		_ = os.Rename(s.path, s.path+".corrupt")
		return nil, nil
	}
	return entries, nil
}

// decode parses the file's JSON; ok is false when it is not valid.
func decode(data []byte) (entries []Entry, ok bool) {
	var f file
	if json.Unmarshal(data, &f) != nil {
		return nil, false
	}
	return f.Entries, true
}

// save writes entries atomically: a temp file in the same directory, renamed
// over the real one.
func (s *Store) save(entries []Entry) error {
	if s.path == "" {
		return errNoPath
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("favorites: %w", err)
	}
	data, err := json.MarshalIndent(file{Version: 1, Entries: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("favorites: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".favorites-*.tmp")
	if err != nil {
		return fmt.Errorf("favorites: %w", err)
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
		return fmt.Errorf("favorites: %w", err)
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
