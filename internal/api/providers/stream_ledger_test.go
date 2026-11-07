package providers

import (
	"slices"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The migration to source.Stream, one source at a time. A source that returns
// a Stream no longer writes the player's state from inside its resolver; one
// that does not still goes through FetchStreamURL. The ledger is exact on
// purpose: migrating a source, or registering a new one, means moving it here,
// so the list cannot drift from the code.
var (
	streamMigrated = []source.SourceKind{
		source.Pomfy, source.StartFlix, source.TopCine, source.Pobreflix,
		source.HiAnime, source.Anroll, source.AnimeDrive,
	}
	streamNotYetMigrated = []source.SourceKind{
		source.AnimeFire, source.Goyabu, source.AnimesDigital,
	}
)

func TestStreamMigrationLedgerMatchesTheRegistry(t *testing.T) {
	registered := source.Kinds()
	var fetchers, legacy []source.SourceKind
	for _, kind := range registered {
		src, ok := source.Registered(kind)
		require.True(t, ok)
		if _, isFetcher := src.(source.StreamFetcher); isFetcher {
			fetchers = append(fetchers, kind)
		} else {
			legacy = append(legacy, kind)
		}
	}
	sortKinds := func(k []source.SourceKind) []source.SourceKind {
		k = slices.Clone(k)
		slices.Sort(k)
		return k
	}
	assert.Equal(t, sortKinds(streamMigrated), sortKinds(fetchers),
		"a source that returns a Stream must be in streamMigrated, and only those")
	assert.Equal(t, sortKinds(streamNotYetMigrated), sortKinds(legacy),
		"a source that does not must be in streamNotYetMigrated; a new source starts there or is migrated")
}
