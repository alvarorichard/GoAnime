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
