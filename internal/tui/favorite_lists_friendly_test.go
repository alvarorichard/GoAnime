package tui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyTurnsAPhraseIntoASentence(t *testing.T) {
	assert.Equal(t, "You already have a list with that name.", friendly(errors.New("you already have a list with that name")))
	assert.Equal(t, "Done!", friendly(errors.New("done!")))
	assert.Equal(t, "Something went wrong.", friendly(errors.New("  ")))
}

func TestFriendlyWordsFileProblemsForABeginner(t *testing.T) {
	for _, tc := range []struct {
		err  string
		want string
	}{
		{"favorites: open: permission denied", "not allowed to write to its data folder"},
		{"favorites: mkdir: permission denied", "not allowed to write to its data folder"},
		{"favorites: rename: read-only file system", "Check that the folder is not read-only"},
		{"favorites: write: no space left on device", "the disk is full"},
		{"favorites: favorites.json is damaged and could not be set aside: rename: permission denied", "damaged and could not be moved aside"},
		{"favorites: favorites.json was written by a newer GoAnime; update to change it", "Update GoAnime to change them"},
		{"favorites: something odd", "could not save your lists (something odd)"},
	} {
		got := friendly(errors.New(tc.err))
		assert.Contains(t, got, tc.want, tc.err)
		assert.NotContains(t, got, "favorites:", "no developer prefix on the screen: %s", tc.err)
		assert.NotContains(t, got, "open:", tc.err)
	}
}

func TestFriendlyLeavesTheListMessagesAlone(t *testing.T) {
	assert.Equal(t, "That name is already used by the Favorites list.", friendly(errors.New("that name is already used by the Favorites list")))
	assert.Equal(t, "The Favorites list cannot be renamed or deleted.", friendly(errors.New("the Favorites list cannot be renamed or deleted")))
}
