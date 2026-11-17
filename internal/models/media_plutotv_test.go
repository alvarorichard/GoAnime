package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlutoTVHasAnInteractiveEpisodeFlowBySource(t *testing.T) {
	assert.True(t, (&Media{Source: "PlutoTV"}).HasInteractiveEpisodeFlow(), "named by source, whatever the media type")
	assert.True(t, (&Media{Source: "PlutoTV", MediaType: MediaTypeTV}).HasInteractiveEpisodeFlow())
	assert.False(t, (&Media{Source: "Animefire.io", MediaType: MediaTypeAnime}).HasInteractiveEpisodeFlow(), "an anime source does not")
}
