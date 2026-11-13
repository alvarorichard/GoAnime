package tui

import (
	"strings"
	"time"
	"unicode"
)

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
// Errors are plain-words phrases in lowercase with no final period; the
// screens show them through friendly.
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

// friendly turns an error phrase ("you already have a list with that name")
// into a sentence for the screen ("You already have a list with that name.").
//
// A problem with the saved file itself ("favorites: open: permission denied")
// becomes a sentence a beginner can act on instead of the developer's text.
func friendly(err error) string {
	text := singleLine(err.Error())
	if text == "" {
		return "Something went wrong."
	}
	if rest, ok := strings.CutPrefix(strings.ToLower(text), "favorites: "); ok {
		return fileProblemSentence(rest)
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	text = string(runes)
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
	}
	return text
}

// fileProblemSentence words a failure to read or write the saved lists.
func fileProblemSentence(reason string) string {
	switch {
	case strings.Contains(reason, "newer goanime"):
		return "Your saved lists were written by a newer GoAnime, so this one leaves them alone. Update GoAnime to change them."
	case strings.Contains(reason, "damaged"):
		return "Your saved lists file is damaged and could not be moved aside, so nothing was changed."
	case strings.Contains(reason, "permission denied"), strings.Contains(reason, "read-only"):
		return "GoAnime could not save your lists: it is not allowed to write to its data folder. Check that the folder is not read-only."
	case strings.Contains(reason, "no space left"):
		return "GoAnime could not save your lists: the disk is full."
	}
	return "GoAnime could not save your lists (" + reason + ")."
}
