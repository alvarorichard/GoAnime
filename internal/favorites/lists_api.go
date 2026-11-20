package favorites

import (
	"errors"
	"time"
)

// FavoritesID is the ID of the automatic Favorites list: the one Toggle, Has,
// Keys, List, Remove and Clear work on.
const FavoritesID = "favorites"

// Names and limits of the lists the user makes.
const (
	maxListNameRunes = 40
	maxLists         = 50
)

// Errors about lists. Their text is a plain-words phrase, lowercase and with
// no final period (the lint rule for errors); the screens capitalise it and
// end it with a period when they show it.
var (
	ErrNameEmpty    = errors.New("give the list a name")
	ErrNameTooLong  = errors.New("a list name can have at most 40 characters")
	ErrNameTaken    = errors.New("you already have a list with that name")
	ErrReservedName = errors.New("that name is already used by the Favorites list")
	ErrTooManyLists = errors.New("you can have at most 50 lists, delete one first")
	ErrNoSuchList   = errors.New("that list no longer exists")
	ErrAutoList     = errors.New("the Favorites list cannot be renamed or deleted")
)

// Summary is a list as the screens show it.
type Summary struct {
	ID        string
	Name      string
	Auto      bool      // the automatic Favorites list
	CreatedAt time.Time // zero for the automatic list
	Count     int
	Titles    []string // the titles of its works, newest first
}
