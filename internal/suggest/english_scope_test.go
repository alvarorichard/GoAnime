package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The ENGLISH tab searches with the "english" selector (HiAnime and Pluto TV).
// It is an anime tab like PT-BR, so its autocomplete suggests anime, as it did
// when the tab searched with "hianime".
func TestTheEnglishScopeSuggestsAnime(t *testing.T) {
	assert.Equal(t, []Kind{KindAnime}, KindsForScope("english"))
	assert.Equal(t, []Kind{KindAnime}, KindsForScope(" English "))
	assert.Equal(t, KindsForScope("hianime"), KindsForScope("english"))
	assert.Equal(t, []Kind{KindMovie, KindSeries}, KindsForScope("movietv"))
	assert.Nil(t, KindsForScope(""), "ALL suggests every kind")
}
