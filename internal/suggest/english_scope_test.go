package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The ENGLISH tab searches with the "english" selector: HiAnime's anime and
// Pluto TV's movies and series. Its autocomplete therefore suggests every
// kind, as ALL does; suggesting anime only hid the titles the tab can find.
func TestTheEnglishScopeSuggestsEveryKind(t *testing.T) {
	assert.Nil(t, KindsForScope("english"), "nil means every kind")
	assert.Nil(t, KindsForScope(" English "))
	assert.Equal(t, KindsForScope(""), KindsForScope("english"), "the same as ALL")
}

func TestTheOtherScopesSuggestWhatTheyDid(t *testing.T) {
	assert.Equal(t, []Kind{KindAnime}, KindsForScope("hianime"), "--source hianime is anime alone")
	assert.Equal(t, []Kind{KindAnime}, KindsForScope("ptbr"))
	assert.Equal(t, []Kind{KindMovie, KindSeries}, KindsForScope("movietv"))
	assert.Nil(t, KindsForScope("plutotv"), "a single source has no kind of its own here")
}
