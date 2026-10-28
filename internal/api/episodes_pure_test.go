package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEpisodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"single digit", "Episódio 1", 1, false},
		{"two digit", "Episódio 42", 42, false},
		{"no match returns 1", "Movie", 1, false},
		{"upper case", "EPISÓDIO 12", 12, false},
		{"ascii o", "Episodio 5", 5, false},
		{"empty returns 1", "", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseEpisodeNumber(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortEpisodesByNum(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{Num: 3},
		{Num: 1},
		{Num: 2},
	}
	sortEpisodesByNum(eps)
	assert.Equal(t, 1, eps[0].Num)
	assert.Equal(t, 2, eps[1].Num)
	assert.Equal(t, 3, eps[2].Num)
}

func TestParseEpisodes(t *testing.T) {
	t.Parallel()

	t.Run("parses anchor list", func(t *testing.T) {
		t.Parallel()
		html := `<html><body>
<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/1">Episódio 1</a>
<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/2">Episódio 2</a>
</body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)
		eps := parseEpisodes(doc)
		require.Len(t, eps, 2)
		assert.Equal(t, 1, eps[0].Num)
		assert.Equal(t, "/ep/1", eps[0].URL)
	})

	t.Run("empty doc", func(t *testing.T) {
		t.Parallel()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<html></html>`))
		require.NoError(t, err)
		assert.Empty(t, parseEpisodes(doc))
	})
}

// GetAnimeEpisodes wires SafeGet (SSRF-guarded). Loopback rejected at the
// dial layer, so success-path is integration-only. Exercise error path here.
func TestGetAnimeEpisodes_ErrorOnInvalidURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"malformed", "://bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			eps, err := GetAnimeEpisodes(tt.url)
			assert.Error(t, err)
			assert.Nil(t, eps)
		})
	}
}

// httptest binds 127.0.0.1; SafeGet rejects it. Confirm the error path
// surfaces a clean failure and no partial data is returned.
func TestGetAnimeEpisodes_LoopbackRejected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html></html>`))
	}))
	t.Cleanup(srv.Close)
	eps, err := GetAnimeEpisodes(srv.URL)
	assert.Error(t, err)
	assert.Nil(t, eps)
}
