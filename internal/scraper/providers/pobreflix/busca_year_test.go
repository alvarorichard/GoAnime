package pobreflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// /api/busca names each record's release year. It used to be read and thrown
// away, so Pobreflix's iCarly (2007) reached the results screen with no year
// and could not be told from the 2021 revival the other catalogs list.

func TestBuscaSearchCarriesTheYear(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[
		 {"title":"iCarly","year":2007,"type":"series","url":"/series/online/icarly-dublado-5371"},
		 {"title":"iCarly","year":2021,"type":"series","url":"/series/online/icarly-legendado-9001"},
		 {"title":"Sem Ano","year":0,"type":"movie","url":"/filmes/online/sem-ano-dublado-77"},
		 {"title":"Ano Absurdo","year":99999,"type":"movie","url":"/filmes/online/ano-absurdo-dublado-78"}]`))
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(http.DefaultClient, srv.URL, t.TempDir()+"/corpus.json")
	got, err := c.buscaSearch(context.Background(), "icarly")
	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Equal(t, "2007", got[0].Year)
	assert.Equal(t, "2021", got[1].Year)
	assert.Empty(t, got[2].Year, "year 0 means unknown")
	assert.Empty(t, got[3].Year, "a year that cannot be one is not trusted")
}

func TestToAnimeModelCarriesTheYear(t *testing.T) {
	t.Parallel()
	a := Media{Title: "iCarly", URL: "u", Kind: KindSeries, Audio: AudioDubbed, Year: "2007"}.ToAnimeModel()
	assert.Equal(t, "2007", a.Year)
	assert.Equal(t, models.MediaTypeTV, a.MediaType)
	assert.Empty(t, Media{Title: "x", Kind: KindMovie}.ToAnimeModel().Year)
}
