package providers

import (
	"slices"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// streamSources are the sources that return a Stream: their resolver writes no
// playback state, and fetchAndPublish publishes what they return. The list is
// exact on purpose: a source added to the registry joins it here, so it cannot
// ship resolving through FetchStreamURL alone and writing the player's state
// from inside its own code.
var streamSources = []source.SourceKind{
	source.Pomfy, source.StartFlix, source.TopCine, source.Pobreflix,
	source.HiAnime, source.Anroll, source.AnimeDrive,
	source.AnimeFire, source.Goyabu, source.AnimesDigital,
}

func TestEveryRegisteredSourceReturnsAStream(t *testing.T) {
	var fetchers, legacy []source.SourceKind
	for _, kind := range source.Kinds() {
		src, ok := source.Registered(kind)
		require.True(t, ok)
		if _, isFetcher := src.(source.StreamFetcher); isFetcher {
			fetchers = append(fetchers, kind)
		} else {
			legacy = append(legacy, kind)
		}
	}
	slices.Sort(fetchers)
	want := slices.Clone(streamSources)
	slices.Sort(want)

	assert.Empty(t, legacy, "these sources still resolve only through FetchStreamURL; give them FetchStream")
	assert.Equal(t, want, fetchers, "streamSources must list exactly the registered sources")
}
