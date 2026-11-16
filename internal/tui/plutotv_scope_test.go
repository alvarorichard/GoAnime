package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A search the user asked to try again opens on the tab that holds the source
// it was made on; for Pluto TV that is the movie and series tab.
func TestPlutoTVScopeIsTheMovieTab(t *testing.T) {
	assert.Equal(t, 3, indexForScope("plutotv"))
	assert.Equal(t, 3, indexForScope("PlutoTV"))
	assert.Equal(t, indexForScope("movietv"), indexForScope("plutotv"))
	assert.NotEqual(t, indexForScope("ptbr"), indexForScope("plutotv"))
}

// The ENGLISH tab searches HiAnime and Pluto TV, so it has a selector of its
// own: "hianime" stays the name of one source, and a search made on the tab
// reopens on it.
func TestTheEnglishTabSearchesWithItsOwnSelector(t *testing.T) {
	assert.Equal(t, "ENGLISH", searchScopes[2].Name)
	assert.Equal(t, "english", searchScopes[2].Value)
	assert.Equal(t, "english", TabEnglish)
	assert.Equal(t, 2, indexForScope("english"), "a retry of a search made on the tab opens on the tab")
	assert.Equal(t, 2, indexForScope(" English "))
	assert.Equal(t, 2, indexForScope("hianime"), "--source hianime still opens there")
}

func TestThePlutoAliasRetriesOnTheSameTab(t *testing.T) {
	assert.Equal(t, indexForScope("plutotv"), indexForScope("pluto"))
	assert.Equal(t, 3, indexForScope(" PLUTO "))
}
