// Package favorites keeps the works the user marked as FAVORITES, like the
// "My List" of the streaming services. They live in one small JSON file next
// to the watch history, so they work on every platform and in builds without
// SQLite (no CGO).
//
// Besides the automatic Favorites list (what ctrl+f fills) the user can make
// any number of named lists of works ("Weekend marathon"). They live in the
// same file, under an optional "lists" field, so a file written before lists
// existed still loads and a change to one kind of list never drops the other.
// The automatic list is reached by List, Has, Keys, Toggle, Remove and Clear;
// the named ones by the *List, *In and *From methods in lists.go, which also
// accept FavoritesID for the automatic list.
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

// maxEntries bounds each list; the oldest work is dropped past it.
const maxEntries = 500

// schemaVersion is the file layout this build reads and writes. A file from a
// newer build is left alone rather than rewritten without its new fields. The
// optional "lists" field is additive, so it did not need a new version.
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

// file is the whole document: the automatic Favorites list in Entries and the
// lists the user made, which a file written before lists existed does not have.
type file struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
	Lists   []list  `json:"lists,omitempty"`
}

// Store reads and writes the favorites file.
type Store struct {
	path string
	mu   sync.Mutex
	// damaged is where a damaged file was set aside, until TakeDamaged says so.
	damaged string
	// now is the clock; nil means time.Now. Tests set it.
	now func() time.Time
}

func (s *Store) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
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
// "mkdir /x/y: permission denied". A rename error carries two paths.
func reason(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return fmt.Errorf("%s: %w", pe.Op, pe.Err)
	}
	var le *os.LinkError
	if errors.As(err, &le) {
		return fmt.Errorf("%s: %w", le.Op, le.Err)
	}
	return err
}

func normalize(key string) string { return strings.ToLower(strings.TrimSpace(key)) }

// load reads the whole document. A missing file is an empty one. A file that
// is not valid JSON is moved aside (favorites.json.corrupt.<time>) and treated
// as empty, so one bad write never locks the feature; when it cannot be moved
// aside the error says so, so it is never overwritten unseen. A file from a
// newer GoAnime is refused with ErrNewerFile. What is read is normalised (see
// normalizeDocument): duplicate keys collapse to the newest, and the lists get
// unique IDs and clean names.
func (s *Store) load() (file, error) {
	if s.path == "" {
		return file{}, errNoPath
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return file{}, nil
	}
	if err != nil {
		return file{}, fmt.Errorf("favorites: %w", reason(err))
	}
	f, ok := decode(data)
	if !ok {
		aside := s.path + ".corrupt." + time.Now().Format("20060102-150405")
		if err := os.Rename(s.path, aside); err != nil {
			return file{}, fmt.Errorf("favorites: favorites.json is damaged and could not be set aside: %w", reason(err))
		}
		s.damaged = aside
		return file{}, nil
	}
	if f.Version > schemaVersion {
		return file{}, ErrNewerFile
	}
	normalizeDocument(&f)
	return f, nil
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

// save writes the whole document atomically: a temp file in the same
// directory, renamed over the real one.
func (s *Store) save(f file) error {
	if s.path == "" {
		return errNoPath
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("favorites: %w", reason(err))
	}
	f.Version = schemaVersion
	lists := make([]list, len(f.Lists))
	copy(lists, f.Lists)
	for i := range lists {
		if lists[i].Entries == nil {
			lists[i].Entries = []Entry{} // "entries": [], never null
		}
	}
	f.Lists = lists
	data, err := json.MarshalIndent(f, "", "  ")
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
func (s *Store) List() ([]Entry, error) { return s.Entries(FavoritesID) }

// Has reports whether the work with key is a favorite. Any failure reads as no.
func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return false
	}
	return hasKey(f.Entries, normalize(key))
}

func hasKey(entries []Entry, key string) bool {
	for i := range entries {
		if entries[i].Key == key {
			return true
		}
	}
	return false
}

// TakeDamaged reports, once, where a damaged favorites file was set aside by
// an earlier call, or "" when none was, so the screen can tell the user.
func (s *Store) TakeDamaged() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	aside := s.damaged
	s.damaged = ""
	return aside
}

// Keys returns the key of every favorite, from one read of the file.
func (s *Store) Keys() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return map[string]bool{}
	}
	keys := make(map[string]bool, len(f.Entries))
	for i := range f.Entries {
		keys[f.Entries[i].Key] = true
	}
	return keys
}

// Toggle adds e when its key is not a favorite and removes it when it is. It
// reports whether the work is a favorite afterwards.
func (s *Store) Toggle(e Entry) (bool, error) { return s.ToggleIn(FavoritesID, e) }

// Remove drops the favorite with key. A key that is not there is not an error.
func (s *Store) Remove(key string) error { return s.RemoveFrom(FavoritesID, key) }

// Clear removes every favorite.
func (s *Store) Clear() error { return s.ClearList(FavoritesID) }

// toggleEntry adds e to entries when its key is not there and takes it off
// when it is; it reports whether e is in the result. A full list drops its
// oldest work.
func toggleEntry(entries []Entry, e Entry, now time.Time) ([]Entry, bool) {
	for i := range entries {
		if entries[i].Key == e.Key {
			return append(entries[:i], entries[i+1:]...), false
		}
	}
	if e.AddedAt.IsZero() {
		e.AddedAt = now
	}
	entries = append(entries, e)
	if len(entries) > maxEntries {
		sortNewestFirst(entries)
		entries = entries[:maxEntries]
	}
	return entries, true
}

func sortNewestFirst(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].AddedAt.After(entries[j].AddedAt) })
}
