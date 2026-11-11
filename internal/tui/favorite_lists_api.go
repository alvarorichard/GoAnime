package tui

import "time"

// HomeList is one list of works in FAVORITES: the automatic "Favorites" list
// (what ctrl+f fills) or a list the user made.
type HomeList struct {
	ID      string
	Name    string
	Auto    bool      // the automatic Favorites list: it cannot be renamed or deleted
	Created time.Time // zero for the automatic list
	Count   int
	Titles  []string // the titles of its works, newest first
}

// FavoritesListID is the ID of the automatic Favorites list.
const FavoritesListID = "favorites"

// FavoriteLists is what the screens need from the saved lists. The saved file
// is behind it (see StoreLists); the screens' tests use fakes.
//
// Errors are shown to the user as they are, so their text is plain words.
type FavoriteLists interface {
	// Lists returns the automatic Favorites list first, then the user's lists,
	// oldest first.
	Lists() ([]HomeList, error)
	// Works returns the works of a list, newest first.
	Works(id string) ([]HomeFavorite, error)
	// Create makes a new empty list. The name is trimmed and its spaces
	// collapsed; an empty, too long, taken or reserved name is an error.
	Create(name string) (HomeList, error)
	// Rename renames a user's list (not the automatic one).
	Rename(id, name string) error
	// Delete deletes a user's list and what it holds (not the automatic one).
	Delete(id string) error
	// ToggleWork adds the work to the list or, when it is there, takes it off.
	// It reports whether the work is in the list afterwards.
	ToggleWork(id string, w HomeFavorite) (bool, error)
	// RemoveWork takes the work with key off a list.
	RemoveWork(id, key string) error
	// Clear empties a list (the list itself stays).
	Clear(id string) error
	// Membership is the IDs of the lists that hold the work with key.
	Membership(key string) map[string]bool
}
