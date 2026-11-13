package favorites

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// list is a list of works the user made, as it is saved.
type list struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Entries   []Entry   `json:"entries"`
}

// untitledList names a list whose saved name is empty (a hand-edited file).
const untitledList = "Untitled list"

// normalizeDocument makes a document read from disk safe to work on. The file
// is untrusted (hand-edited, or from a bug), so it is repaired rather than
// refused, and the same file always repairs the same way:
//
//   - the automatic list and every list's works collapse duplicate keys to the
//     newest, and works without a key are dropped;
//   - a list whose ID is empty, is FavoritesID, or repeats the ID of an earlier
//     list keeps its works and gets the next free ID of the form "r1", "r2"...
//     (the first list with a given valid ID keeps it), so no list is lost and
//     no list can shadow the automatic one;
//   - a list name goes through cleanName, is cut to maxListNameRunes runes, and
//     an empty one becomes "Untitled list". Duplicate or reserved names are
//     left alone: they are harmless to read, and renaming them is up to the
//     user;
//   - more than maxLists lists are all kept (nothing is lost on a save);
//     CreateList refuses to add one more.
func normalizeDocument(f *file) {
	f.Entries = dedupe(f.Entries)
	if len(f.Lists) == 0 {
		f.Lists = nil
		return
	}
	taken := make(map[string]bool, len(f.Lists)+1)
	taken[FavoritesID] = true
	keeps := make([]bool, len(f.Lists))
	for i := range f.Lists {
		id := f.Lists[i].ID
		if id != "" && !taken[id] {
			taken[id] = true
			keeps[i] = true
		}
	}
	next := 0
	for i := range f.Lists {
		l := &f.Lists[i]
		if !keeps[i] {
			for {
				next++
				id := "r" + strconv.Itoa(next)
				if !taken[id] {
					taken[id] = true
					l.ID = id
					break
				}
			}
		}
		l.Name = cleanName(l.Name)
		if r := []rune(l.Name); len(r) > maxListNameRunes {
			l.Name = strings.TrimSpace(string(r[:maxListNameRunes]))
		}
		if l.Name == "" {
			l.Name = untitledList
		}
		l.Entries = dedupe(l.Entries)
	}
}

// cleanName collapses the white space of a name to single spaces, trims it,
// and drops control characters and bytes that are not valid UTF-8.
func cleanName(name string) string {
	fields := strings.Fields(strings.ToValidUTF8(name, ""))
	out := fields[:0]
	for _, w := range fields {
		w = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, w)
		if w != "" {
			out = append(out, w)
		}
	}
	return strings.Join(out, " ")
}

// foldName is name for comparing: lower case, without accents and without
// invisible format characters (a zero-width space makes "Favor\u200bites" look
// like the reserved name).
func foldName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, norm.NFD.String(strings.ToLower(name)))
}

// isReserved reports whether name would pass for the automatic list: it reads
// "favorites" ignoring case, accents and spaces.
func isReserved(name string) bool {
	return strings.ReplaceAll(foldName(name), " ", "") == "favorites"
}

// checkName cleans name and says whether the user may give it to the list
// with ID except ("" for a new list).
func checkName(f *file, name, except string) (string, error) {
	name = cleanName(name)
	switch {
	case name == "":
		return "", ErrNameEmpty
	case utf8.RuneCountInString(name) > maxListNameRunes:
		return "", ErrNameTooLong
	case isReserved(name):
		return "", ErrReservedName
	}
	folded := foldName(name)
	for i := range f.Lists {
		if f.Lists[i].ID != except && foldName(f.Lists[i].Name) == folded {
			return "", ErrNameTaken
		}
	}
	return name, nil
}

// find returns the index of the list with id, or -1.
func (f *file) find(id string) int {
	for i := range f.Lists {
		if f.Lists[i].ID == id {
			return i
		}
	}
	return -1
}

// works returns the works of a list: the automatic one for FavoritesID.
func (f *file) works(id string) (*[]Entry, error) {
	if id == FavoritesID {
		return &f.Entries, nil
	}
	i := f.find(id)
	if i < 0 {
		return nil, ErrNoSuchList
	}
	return &f.Lists[i].Entries, nil
}

func titles(entries []Entry) []string {
	sorted := append([]Entry(nil), entries...)
	sortNewestFirst(sorted)
	out := make([]string, len(sorted))
	for i := range sorted {
		out[i] = sorted[i].Title
	}
	return out
}

// Summaries returns every list: the automatic Favorites first, then the
// user's lists, oldest first (lists made at the same instant keep the order of
// the file).
func (s *Store) Summaries() ([]Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(f.Lists)+1)
	out = append(out, Summary{
		ID: FavoritesID, Name: "Favorites", Auto: true,
		Count: len(f.Entries), Titles: titles(f.Entries),
	})
	for i := range f.Lists {
		l := &f.Lists[i]
		out = append(out, Summary{
			ID: l.ID, Name: l.Name, CreatedAt: l.CreatedAt,
			Count: len(l.Entries), Titles: titles(l.Entries),
		})
	}
	user := out[1:]
	sort.SliceStable(user, func(i, j int) bool { return user[i].CreatedAt.Before(user[j].CreatedAt) })
	return out, nil
}

// CreateList makes a new empty list. The name's white space is collapsed and
// control characters are dropped; an empty, too long (over 40 characters),
// reserved ("Favorites") or taken name (ignoring case and accents) is an
// error, and so is a 51st list.
func (s *Store) CreateList(name string) (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return Summary{}, err
	}
	name, err = checkName(&f, name, "")
	if err != nil {
		return Summary{}, err
	}
	if len(f.Lists) >= maxLists {
		return Summary{}, ErrTooManyLists
	}
	now := s.clock().UTC()
	l := list{ID: newListID(&f, now), Name: name, CreatedAt: now}
	f.Lists = append(f.Lists, l)
	if err := s.save(f); err != nil {
		return Summary{}, err
	}
	return Summary{ID: l.ID, Name: l.Name, CreatedAt: l.CreatedAt, Titles: []string{}}, nil
}

// newListID returns an ID no list has: "l" and the base 36 of the time in
// nanoseconds, bumped until it is free. It is URL-safe and never FavoritesID.
func newListID(f *file, now time.Time) string {
	n := now.UnixNano()
	if n < 0 {
		n = -n
	}
	for {
		id := "l" + strconv.FormatInt(n, 36)
		if f.find(id) < 0 {
			return id
		}
		n++
	}
}

// RenameList renames a list of the user's. Its own current name is allowed.
func (s *Store) RenameList(id, name string) error {
	if id == FavoritesID {
		return ErrAutoList
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return err
	}
	i := f.find(id)
	if i < 0 {
		return ErrNoSuchList
	}
	name, err = checkName(&f, name, id)
	if err != nil {
		return err
	}
	f.Lists[i].Name = name
	return s.save(f)
}

// DeleteList deletes a list of the user's and the works in it.
func (s *Store) DeleteList(id string) error {
	if id == FavoritesID {
		return ErrAutoList
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return err
	}
	i := f.find(id)
	if i < 0 {
		return ErrNoSuchList
	}
	f.Lists = append(f.Lists[:i], f.Lists[i+1:]...)
	return s.save(f)
}

// Entries returns the works of a list, newest first. FavoritesID is the
// automatic list.
func (s *Store) Entries(id string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return nil, err
	}
	entries, err := f.works(id)
	if err != nil {
		return nil, err
	}
	sortNewestFirst(*entries)
	return *entries, nil
}

// ToggleIn adds e to the list when its key is not there and takes it off when
// it is. It reports whether the work is in the list afterwards. A list holds
// at most 500 works; the oldest goes first.
func (s *Store) ToggleIn(id string, e Entry) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return false, err
	}
	entries, err := f.works(id)
	if err != nil {
		return false, err
	}
	e.Key = normalize(e.Key)
	if e.Key == "" {
		return false, errors.New("favorites: a work needs a key")
	}
	var on bool
	*entries, on = toggleEntry(*entries, e, s.clock())
	return on, s.save(f)
}

// RemoveFrom takes the work with key off a list. A key that is not there is
// not an error.
func (s *Store) RemoveFrom(id, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return err
	}
	entries, err := f.works(id)
	if err != nil {
		return err
	}
	key = normalize(key)
	kept := (*entries)[:0]
	for _, e := range *entries {
		if e.Key != key {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(*entries) {
		return nil
	}
	*entries = kept
	return s.save(f)
}

// ClearList empties a list; the list itself stays.
func (s *Store) ClearList(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.load()
	if err != nil {
		return err
	}
	entries, err := f.works(id)
	if err != nil {
		return err
	}
	*entries = nil
	return s.save(f)
}

// Membership returns the IDs of the lists that hold the work with key,
// FavoritesID included. Any failure reads as no lists.
func (s *Store) Membership(key string) map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	in := map[string]bool{}
	f, err := s.load()
	if err != nil {
		return in
	}
	key = normalize(key)
	if hasKey(f.Entries, key) {
		in[FavoritesID] = true
	}
	for i := range f.Lists {
		if hasKey(f.Lists[i].Entries, key) {
			in[f.Lists[i].ID] = true
		}
	}
	return in
}
