package player

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
)

// The next episode is read ahead only when its address is a page the generic
// extractor can read. A Pluto TV address is an identifier: reading it as a page
// fetched the site's app shell for seconds, found no video in it, and did so
// while the episode before it was buffering.
func TestTheNextEpisodeIsReadAheadOnlyWhenItsAddressIsAPage(t *testing.T) {
	assert.True(t, preloadable("https://animefire.io/animes/naruto-shippuden/2"))
	assert.True(t, preloadable("http://example.org/a-long-enough-page-address/episode-2"))

	assert.False(t, preloadable(plutotv.EpisodeURL("622102dae9e2e9001399c4a9", "62210348e9e2e9001399d90a")), "an episode of Pluto TV")
	assert.False(t, preloadable(plutotv.MovieURL("62b9ea9ec4eb8f00140cb633")), "a movie of Pluto TV")
	assert.False(t, preloadable("https://www.pluto.tv/on-demand/series/622102dae9e2e9001399c4a9/episode/62210348e9e2e9001399d90a"))
	assert.False(t, preloadable("HTTPS://PLUTO.TV/on-demand/movies/62b9ea9ec4eb8f00140cb633"), "however it is cased")

	assert.False(t, preloadable("12345"), "a bare id is not an address at all")
	assert.False(t, preloadable(""))
	assert.True(t, preloadable("https://notpluto.tv.example/on-demand/series/622102dae9e2e9001399c4a9"), "only Pluto TV's own host is left out")
}
