package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper"
)

func TestPlutoTVIsASourceOfTheSDK(t *testing.T) {
	assert.Equal(t, "PlutoTV", SourcePlutoTV.String())
	assert.Equal(t, scraper.PlutoTVType, SourcePlutoTV.ToScraperType())
	assert.Contains(t, Sources(), SourcePlutoTV)
	assert.False(t, SourcePlutoTV.SupportsEpisodes(), "its seasons are the CLI's picker, until the SDK has one of its own")
	assert.False(t, SourcePlutoTV.IsCatalog())
}

func TestPlutoTVParsesFromEverySpellingItAnswersTo(t *testing.T) {
	for _, spelling := range []string{"PlutoTV", "plutotv", "pluto"} {
		got, err := ParseSource(spelling)
		require.NoError(t, err, spelling)
		assert.Equal(t, SourcePlutoTV, got, spelling)
	}
	_, err := ParseSource("pluto tv")
	assert.Error(t, err, "spellings are exact")
}

func TestSourceValuesStayAppendOnlyAfterPlutoTV(t *testing.T) {
	assert.Equal(t, int(SourceAnimesFHD)+1, int(SourcePlutoTV), "a new source goes at the end; a stored value keeps its meaning")
	assert.Equal(t, int(SourcePlutoTV)+1, int(SourceMiruro), "Miruro came after Pluto TV")
	assert.Equal(t, int(SourceMiruro)+1, len(Sources()))
}
