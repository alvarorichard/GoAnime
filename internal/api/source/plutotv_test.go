package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper"
)

func TestPlutoTVKindMapsToItsScraper(t *testing.T) {
	st, ok := ScraperTypeFor(PlutoTV)
	require.True(t, ok)
	assert.Equal(t, scraper.PlutoTVType, st)
	kind, ok := KindFor(scraper.PlutoTVType)
	require.True(t, ok)
	assert.Equal(t, PlutoTV, kind)
	assert.Equal(t, "PlutoTV", string(PlutoTV))
}

func TestEveryScraperTypeHasOneKindAndBack(t *testing.T) {
	seen := map[scraper.ScraperType]SourceKind{}
	for kind, st := range scraperTypeMap {
		if other, dup := seen[st]; dup {
			t.Errorf("%s and %s share scraper type %v", kind, other, st)
		}
		seen[st] = kind
		back, ok := KindFor(st)
		assert.True(t, ok, kind)
		assert.Equal(t, kind, back)
	}
	assert.Contains(t, seen, scraper.PlutoTVType)
}
