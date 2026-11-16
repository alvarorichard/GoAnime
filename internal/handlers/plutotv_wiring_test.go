package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Pluto TV on the SEARCH home screen and in CONTINUE: the tab it opens on, the
// selector a saved row reopens it with, the name recovered from a saved URL.

func TestPlutoTVOpensOnTheMovieAndSeriesTab(t *testing.T) {
	assert.Equal(t, 3, scopeIndexForSource("plutotv"))
	assert.Equal(t, 3, scopeIndexForSource(" PlutoTV "), "the flag is not case-sensitive")
	assert.Equal(t, 3, scopeIndexForSource("movietv"), "the tab it shares")
	assert.Equal(t, 0, scopeIndexForSource("pluto-tv-typo"), "an unknown selector opens on the first tab")
}

func TestAContinueRowOfPlutoTVReopensOnPlutoTV(t *testing.T) {
	assert.Equal(t, "plutotv", continueSourceToken("PlutoTV"))
	assert.Equal(t, "plutotv", continueSourceToken(" plutotv "))
	assert.Empty(t, continueSourceToken("Pluto TV"), "only the source's own spelling, as for the others")
}

func TestASavedPlutoTVURLNamesItsSource(t *testing.T) {
	assert.Equal(t, "PlutoTV", sourceFromURL("https://pluto.tv/on-demand/movies/62b9ea9ec4eb8f00140cb633"))
	assert.Equal(t, "PlutoTV", sourceFromURL("https://www.pluto.tv/on-demand/series/6206bba315c8610013683974/episode/6206bbaa15c8610013683a28"))
	assert.Equal(t, "StartFlix", sourceFromURL("https://www.startflix.test/series/dark:ep1"), "the neighbours are unmoved")
}

// "pluto" is the selector's other spelling everywhere else (--source pluto);
// the home screen opens on the same tab for it, so the search that follows
// keeps the one source that was asked for instead of widening to all.
func TestThePlutoAliasOpensOnTheSameTab(t *testing.T) {
	assert.Equal(t, scopeIndexForSource("plutotv"), scopeIndexForSource("pluto"))
	assert.Equal(t, 3, scopeIndexForSource(" Pluto "))
}
