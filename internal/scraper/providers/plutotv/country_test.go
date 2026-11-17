package plutotv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The catalog is the country's the service places the session in. A VPN moves
// it, and some countries list nothing on demand: a search that comes back
// empty for that reason looks exactly like a title that does not exist. So the
// client knows the session's country, and says it when a search finds nothing.

func captureEmptyCatalog(t *testing.T) *[][2]string {
	t.Helper()
	var said [][2]string
	prev := emptyCatalogFn
	emptyCatalogFn = func(country, query string) { said = append(said, [2]string{country, query}) }
	t.Cleanup(func() { emptyCatalogFn = prev })
	return &said
}

func TestCountryIsTheSessionsCountry(t *testing.T) {
	f := newFakePluto(t)
	f.country = "za"
	f.set("/v1/search", searchBody())
	c := f.client()
	assert.Empty(t, c.Country(), "no request yet, so no session: nothing is known")
	assert.Zero(t, f.bootCount(), "asking makes no request")

	_, err := c.Search(t.Context(), "naruto")
	require.NoError(t, err)
	assert.Equal(t, "ZA", c.Country(), "upper case, however the service wrote it")
	assert.Equal(t, 1, f.bootCount())
}

func TestCountryIsEmptyWhenTheServiceDoesNotSayOrSaysSomethingElse(t *testing.T) {
	for name, country := range map[string]string{"no session object": "-", "three letters": "BRA", "digits": "12", "a phrase": "South Africa", "blank": "  "} {
		f := newFakePluto(t)
		f.country = country
		f.set("/v1/search", searchBody(hit("movie", fakeMovieID, "Heat", "heat", "EN")))
		c := f.client()
		titles, err := c.Search(t.Context(), "heat")
		require.NoError(t, err, name)
		assert.Len(t, titles, 1, "%s: the session still works", name)
		assert.Empty(t, c.Country(), name)
	}
}

func TestAnEmptySearchSaysWhoseCatalogItWas(t *testing.T) {
	said := captureEmptyCatalog(t)
	f := newFakePluto(t)
	f.country = "ZA"
	f.set("/v1/search", searchBody())

	titles, err := f.client().Search(t.Context(), "  naruto ")
	require.NoError(t, err, "nothing found is not an error")
	assert.Empty(t, titles)
	assert.Equal(t, [][2]string{{"ZA", "naruto"}}, *said, "the country and the query, once")
}

func TestAnEmptySearchInAnUnknownCountrySaysSo(t *testing.T) {
	said := captureEmptyCatalog(t)
	f := newFakePluto(t)
	f.country = "-"
	f.set("/v1/search", `{"data": []}`)
	_, err := f.client().Search(t.Context(), "x")
	require.NoError(t, err)
	assert.Equal(t, [][2]string{{"unknown", "x"}}, *said)
}

func TestASearchThatFindsSomethingSaysNothing(t *testing.T) {
	said := captureEmptyCatalog(t)
	f := newFakePluto(t)
	// A channel is an answer, even though it is not a title: the catalog is
	// not empty, only this query's titles are.
	f.set("/v1/search", searchBody(hit("channel", fakeMovieID, "Pluto TV Anime", "anime", "PT")))
	titles, err := f.client().Search(t.Context(), "anime")
	require.NoError(t, err)
	assert.Empty(t, titles)
	assert.Empty(t, *said)

	_, err = f.client().Search(t.Context(), "   ")
	require.NoError(t, err)
	assert.Empty(t, *said, "a blank query asks nothing")
}

func TestTheRealEmptyCatalogNoteIsSafeToCall(t *testing.T) {
	assert.NotPanics(t, func() { emptyCatalogFn("ZA", "naruto") })
	assert.Equal(t, "unknown", countryOrUnknown(""))
	assert.Equal(t, "BR", countryOrUnknown("BR"))
}
