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

// Errors about lists. Their text is shown to the user as it is.
var (
	ErrNameEmpty    = errors.New("Give the list a name.")
	ErrNameTooLong  = errors.New("A list name can have at most 40 characters.")
	ErrNameTaken    = errors.New("You already have a list with that name.")
	ErrReservedName = errors.New("That name is already used by the Favorites list.")
	ErrTooManyLists = errors.New("You can have at most 50 lists. Delete one first.")
	ErrNoSuchList   = errors.New("That list no longer exists.")
	ErrAutoList     = errors.New("The Favorites list cannot be renamed or deleted.")
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
