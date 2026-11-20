package suggest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return b
}

// serve answers every request with status and body, recording the last URL.
func serve(t *testing.T, status int, body []byte, seen *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = r.URL.RequestURI()
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestKitsuRecordedReply(t *testing.T) {
	t.Parallel()
	var seen string
	srv := serve(t, 200, fixture(t, "kitsu_narut.json"), &seen)
	k := NewKitsu(srv.Client())
	k.base = srv.URL

	got, err := k.Suggest(t.Context(), "narut", 5)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Contains(t, titles(got), "Naruto")
	for _, s := range got {
		assert.Equal(t, KindAnime, s.Kind)
		assert.Equal(t, "kitsu", s.Source)
		assert.NotEmpty(t, s.Title)
	}
	assert.Contains(t, seen, "filter%5Btext%5D=narut")
	assert.Contains(t, seen, "page%5Blimit%5D=5")
	assert.Equal(t, "kitsu", k.Name())
	assert.Equal(t, []Kind{KindAnime}, k.Kinds())
}

func TestKitsuTitleFallbacks(t *testing.T) {
	t.Parallel()
	body := `{"data":[
	 {"attributes":{"canonicalTitle":"  ","titles":{"en":"English One"},"startDate":"2001-04-05"}},
	 {"attributes":{"canonicalTitle":"","titles":{"en_jp":"Romaji Two"},"startDate":"bad"}},
	 {"attributes":{"canonicalTitle":"","titles":{},"startDate":null}},
	 {"attributes":{"canonicalTitle":"Canon","titles":{"en":"ignored"}}}]}`
	srv := serve(t, 200, []byte(body), nil)
	k := NewKitsu(srv.Client())
	k.base = srv.URL

	got, err := k.Suggest(t.Context(), "xx", 8)
	require.NoError(t, err)
	assert.Equal(t, []Suggestion{
		{Title: "English One", Year: 2001, Kind: KindAnime, Source: "kitsu"},
		{Title: "Romaji Two", Kind: KindAnime, Source: "kitsu"},
		{Title: "Canon", Kind: KindAnime, Source: "kitsu"},
	}, got, "the title falls back to en, then en_jp; a row with none is dropped")
}

func TestIMDbRecordedReply(t *testing.T) {
	t.Parallel()
	var seen string
	srv := serve(t, 200, fixture(t, "imdb_breaking_ba.json"), &seen)
	i := NewIMDb(srv.Client())
	i.base = srv.URL

	got, err := i.Suggest(t.Context(), "Breaking Ba", 8)
	require.NoError(t, err)
	assert.Equal(t, "/suggestion/b/breaking%20ba.json", seen, "sharded by first char, folded, escaped")
	assert.Equal(t, Suggestion{Title: "Breaking Bad", Year: 2008, Kind: KindSeries, Source: "imdb"}, got[0])
	assert.Contains(t, titles(got), "El Camino")
	assert.NotContains(t, titles(got), "Breaking Bad: Chicks 'n' Guns", "videos are not titles to watch")
	for _, s := range got {
		assert.Contains(t, []Kind{KindMovie, KindSeries}, s.Kind)
	}
	assert.Equal(t, "imdb", i.Name())
	assert.Equal(t, []Kind{KindMovie, KindSeries}, i.Kinds())
}

func TestIMDbFiltersAndLimit(t *testing.T) {
	t.Parallel()
	body := `{"d":[
	 {"id":"nm1","l":"A Person","qid":"movie","y":1},
	 {"id":"tt1","l":"","qid":"movie"},
	 {"id":"tt2","l":"Telefilm","qid":"tvMovie","y":2001},
	 {"id":"tt3","l":"Mini","qid":"tvMiniSeries"},
	 {"id":"tt4","l":"Episode","qid":"tvEpisode"},
	 {"id":"tt5","l":"Third","qid":"movie"}]}`
	srv := serve(t, 200, []byte(body), nil)
	i := NewIMDb(srv.Client())
	i.base = srv.URL

	got, err := i.Suggest(t.Context(), "tele", 2)
	require.NoError(t, err)
	assert.Equal(t, []Suggestion{
		{Title: "Telefilm", Year: 2001, Kind: KindMovie, Source: "imdb"},
		{Title: "Mini", Kind: KindSeries, Source: "imdb"},
	}, got, "people, empty titles and episodes are dropped; limit stops the loop")
}

func TestIMDbUnfoldableQuery(t *testing.T) {
	t.Parallel()
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	t.Cleanup(srv.Close)
	i := NewIMDb(srv.Client())
	i.base = srv.URL
	got, err := i.Suggest(t.Context(), "::", 5)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.False(t, called, "punctuation-only text never leaves the machine")
}

func TestProviderFailureModes(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, base string, hc Doer) (kitsuErr, imdbErr error) {
		k, i := NewKitsu(hc), NewIMDb(hc)
		k.base, i.base = base, base
		_, kitsuErr = k.Suggest(t.Context(), "naruto", 5)
		_, imdbErr = i.Suggest(t.Context(), "naruto", 5)
		return kitsuErr, imdbErr
	}

	t.Run("bad status", func(t *testing.T) {
		srv := serve(t, 503, []byte(`{}`), nil)
		k, i := run(t, srv.URL, srv.Client())
		require.ErrorContains(t, k, "unexpected status 503")
		require.ErrorContains(t, i, "unexpected status 503")
	})
	t.Run("bad json", func(t *testing.T) {
		srv := serve(t, 200, []byte(`<html>`), nil)
		k, i := run(t, srv.URL, srv.Client())
		require.ErrorContains(t, k, "decode")
		require.ErrorContains(t, i, "decode")
	})
	t.Run("oversized body", func(t *testing.T) {
		huge := `{"data":[],"d":[],"pad":"` + strings.Repeat("x", maxBody) + `"}`
		srv := serve(t, 200, []byte(huge), nil)
		k, i := run(t, srv.URL, srv.Client())
		require.ErrorContains(t, k, "exceeds size limit")
		require.ErrorContains(t, i, "exceeds size limit")
	})
	t.Run("transport error", func(t *testing.T) {
		k, i := run(t, "http://example.invalid", failingDoer{})
		require.ErrorContains(t, k, "request")
		require.ErrorContains(t, i, "request")
	})
	t.Run("unbuildable request", func(t *testing.T) {
		k, i := run(t, "http://bad\x7fhost", failingDoer{})
		require.ErrorContains(t, k, "build request")
		require.ErrorContains(t, i, "build request")
	})
	t.Run("cancelled context", func(t *testing.T) {
		srv := serve(t, 200, []byte(`{}`), nil)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		k := NewKitsu(srv.Client())
		k.base = srv.URL
		_, err := k.Suggest(ctx, "naruto", 5)
		require.ErrorIs(t, err, context.Canceled)
	})
}

type failingDoer struct{}

func (failingDoer) Do(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }

func TestRequestHeaders(t *testing.T) {
	t.Parallel()
	var accept, agent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept, agent = r.Header.Get("Accept"), r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	k := NewKitsu(srv.Client())
	k.base = srv.URL
	_, err := k.Suggest(t.Context(), "naruto", 5)
	require.NoError(t, err)
	assert.Equal(t, "application/vnd.api+json", accept)
	assert.True(t, strings.HasPrefix(agent, "GoAnime/"), "an honest API client UA, not a browser one: %q", agent)
}

func TestYearOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 2002, yearOf("2002-10-03"))
	assert.Equal(t, 1999, yearOf("1999"))
	assert.Zero(t, yearOf(""))
	assert.Zero(t, yearOf("199"))
	assert.Zero(t, yearOf("19x9-01-01"))
}

func TestDefaultEngine(t *testing.T) {
	t.Parallel()
	e := Default()
	require.NotNil(t, e)
	assert.Same(t, e, Default(), "one shared engine, one shared cache")
	require.Len(t, e.providers, 2)
	assert.Equal(t, "kitsu", e.providers[0].Name())
	assert.Equal(t, "imdb", e.providers[1].Name())
	assert.NotNil(t, NewHTTPClient().Transport)
}
