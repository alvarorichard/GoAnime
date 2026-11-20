package tui

import (
	"errors"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
)

// listStore is the part of the favorites store the lists use. *favorites.Store
// satisfies it; tests give the adapter a fake.
type listStore interface {
	Summaries() ([]favorites.Summary, error)
	CreateList(name string) (favorites.Summary, error)
	RenameList(id, name string) error
	DeleteList(id string) error
	Entries(id string) ([]favorites.Entry, error)
	ToggleIn(id string, e favorites.Entry) (bool, error)
	RemoveFrom(id, key string) error
	ClearList(id string) error
	Membership(key string) map[string]bool
}

// errListsUnavailable is what every method answers when there is no store.
var errListsUnavailable = errors.New("your lists are not available")

// StoreLists adapts the saved favorites file to the FavoriteLists the screens
// use. A nil store gives lists that all fail politely instead of panicking.
func StoreLists(s listStore) FavoriteLists {
	if isNilStore(s) {
		return noLists{}
	}
	return &storeLists{store: s, now: time.Now}
}

// isNilStore reports a nil interface or a nil *favorites.Store inside one.
func isNilStore(s listStore) bool {
	if s == nil {
		return true
	}
	fs, ok := s.(*favorites.Store)
	return ok && fs == nil
}

// storeLists is FavoriteLists over a listStore.
type storeLists struct {
	store listStore
	now   func() time.Time
}

func (l *storeLists) Lists() ([]HomeList, error) {
	sums, err := l.store.Summaries()
	if err != nil {
		return nil, err
	}
	out := make([]HomeList, 0, len(sums))
	for _, s := range sums {
		out = append(out, homeListOf(s))
	}
	// The automatic list leads; the user's lists keep the store's order.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Auto && !out[j].Auto })
	return out, nil
}

func (l *storeLists) Works(id string) ([]HomeFavorite, error) {
	entries, err := l.store.Entries(id)
	if err != nil {
		return nil, err
	}
	entries = slices.Clone(entries)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].AddedAt.After(entries[j].AddedAt) })
	out := make([]HomeFavorite, 0, len(entries))
	for _, e := range entries {
		out = append(out, HomeFavorite{Key: e.Key, Title: e.Title, Year: e.Year, Media: e.Media, Source: e.Source})
	}
	return out, nil
}

func (l *storeLists) Create(name string) (HomeList, error) {
	s, err := l.store.CreateList(name)
	if err != nil {
		return HomeList{}, err
	}
	return homeListOf(s), nil
}

func (l *storeLists) Rename(id, name string) error { return l.store.RenameList(id, name) }

func (l *storeLists) Delete(id string) error { return l.store.DeleteList(id) }

func (l *storeLists) ToggleWork(id string, w HomeFavorite) (bool, error) {
	key := w.Key
	if strings.TrimSpace(key) == "" {
		key = normalizeTitleKey(w.Title)
	}
	return l.store.ToggleIn(id, favorites.Entry{
		Key: key, Title: w.Title, Year: w.Year, Media: w.Media, Source: w.Source,
		AddedAt: l.now(),
	})
}

func (l *storeLists) RemoveWork(id, key string) error { return l.store.RemoveFrom(id, key) }

func (l *storeLists) Clear(id string) error { return l.store.ClearList(id) }

func (l *storeLists) Membership(key string) map[string]bool {
	in := l.store.Membership(key)
	out := make(map[string]bool, len(in))
	for id, ok := range in {
		if ok {
			out[id] = true
		}
	}
	return out
}

// homeListOf converts a store summary to the screens' list.
func homeListOf(s favorites.Summary) HomeList {
	return HomeList{
		ID: s.ID, Name: s.Name, Auto: s.Auto, Created: s.CreatedAt,
		Count: s.Count, Titles: slices.Clone(s.Titles),
	}
}

// noLists is the lists when there is no saved file to back them.
type noLists struct{}

func (noLists) Lists() ([]HomeList, error)                    { return nil, errListsUnavailable }
func (noLists) Works(string) ([]HomeFavorite, error)          { return nil, errListsUnavailable }
func (noLists) Create(string) (HomeList, error)               { return HomeList{}, errListsUnavailable }
func (noLists) Rename(string, string) error                   { return errListsUnavailable }
func (noLists) Delete(string) error                           { return errListsUnavailable }
func (noLists) ToggleWork(string, HomeFavorite) (bool, error) { return false, errListsUnavailable }
func (noLists) RemoveWork(string, string) error               { return errListsUnavailable }
func (noLists) Clear(string) error                            { return errListsUnavailable }
func (noLists) Membership(string) map[string]bool             { return map[string]bool{} }
