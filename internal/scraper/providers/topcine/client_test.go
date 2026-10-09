package topcine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// routeTo sends every request to srv whatever host it names, so tests can use
// the real player host names.
type routeTo struct{ srv *httptest.Server }

func (r routeTo) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(r.srv.URL)
	out := req.Clone(req.Context())
	out.Header.Set("X-Original-Host", req.URL.Host)
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host
	out.Host = target.Host
	return http.DefaultTransport.RoundTrip(out)
}

func newTestClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewClientForTest(&http.Client{Transport: routeTo{srv}}, "https://topcine.test")
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// The catalog host rotates by mirror generation; a rotation should fail here
// loudly rather than as empty searches. Pinned 2026-10-09.
func TestDefaultBasePinned(t *testing.T) {
	t.Parallel()
	if DefaultBase != "https://topcine3.site" {
		t.Errorf("DefaultBase = %q; update this pin together with the host", DefaultBase)
	}
}

func TestNewClientHonoursEnvOverride(t *testing.T) {
	t.Setenv("GOANIME_TOPCINE_URL", " https://topcine4.site/ ")
	if got := NewClient().BaseURL(); got != "https://topcine4.site" {
		t.Errorf("BaseURL = %q, want the trimmed override", got)
	}
}

func TestClientSearch(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotReferer string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery, gotReferer = r.URL.Path, r.URL.Query().Get("q"), r.Header.Get("Referer")
		_, _ = w.Write(fixture(t, "search_2026_10_09.html"))
	}))

	results, err := c.Search(context.Background(), "  breaking-bad_ ")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/buscar" {
		t.Errorf("path = %q, want /buscar (buscar.php only redirects there, over http)", gotPath)
	}
	if gotQuery != "breaking bad" {
		t.Errorf("query = %q, want slug separators turned into spaces", gotQuery)
	}
	if gotReferer != "https://topcine.test/" {
		t.Errorf("referer = %q", gotReferer)
	}

	want := []Media{
		{Title: "Untold - Raygun: A Polêmica do Breaking", URL: "https://topcine.test/filme/untold-raygun-a-polemica-do-breaking", Year: "2026", Kind: KindMovie},
		{Title: "Breaking Bad", URL: "https://topcine.test/serie/breaking-bad", Year: "2008", Kind: KindSeries},
		{Title: "Better Call Saul", URL: "https://topcine.test/serie/better-call-saul", Year: "2015", Kind: KindSeries},
		{Title: "El Camino: A Breaking Bad Film", URL: "https://topcine.test/filme/el-camino-a-breaking-bad-film", Year: "2019", Kind: KindMovie},
	}
	if len(results) != len(want) {
		t.Fatalf("got %d results, want %d: %+v", len(results), len(want), results)
	}
	for i, w := range want {
		got := results[i]
		if got.Title != w.Title || got.URL != w.URL || got.Year != w.Year || got.Kind != w.Kind {
			t.Errorf("result %d = %+v, want %+v", i, got, w)
		}
	}
	if results[1].Poster != "https://image.tmdb.org/t/p/w500/30erzlzIOtOK3k3T3BAl1GiVMP1.jpg" {
		t.Errorf("poster = %q, want the w500 upgrade", results[1].Poster)
	}

	empty, err := c.Search(context.Background(), " - ")
	if err != nil || empty != nil {
		t.Errorf("blank query = (%v, %v), want no request and no results", empty, err)
	}
}

func TestClientSearchHTTPError(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	_, err := c.Search(context.Background(), "x")
	if _, ok := errors.AsType[*netx.SourceDiagnostic](err); !ok {
		t.Fatalf("err = %v, want a classified HTTP status error", err)
	}
}

func TestClientTitleSeries(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write(fixture(t, "title_series_2026_10_09.html"))
	}))

	page := "https://topcine.test/serie/breaking-bad"
	title, err := c.Title(context.Background(), page)
	if err != nil {
		t.Fatal(err)
	}
	if title.Kind != KindSeries || title.TMDBID != 1396 {
		t.Fatalf("title = %v/%d, want series 1396", title.Kind, title.TMDBID)
	}
	counts := []int{7, 13, 13, 13, 16}
	if len(title.Seasons) != len(counts) {
		t.Fatalf("got %d seasons, want %d", len(title.Seasons), len(counts))
	}
	for i, s := range title.Seasons {
		if s.Number != i+1 || len(s.Episodes) != counts[i] {
			t.Errorf("season %d: number %d with %d episodes, want %d", i, s.Number, len(s.Episodes), counts[i])
		}
		for j, e := range s.Episodes {
			if e.Number != j+1 {
				t.Errorf("season %d episode %d numbered %d; episodes must sort numerically", s.Number, j+1, e.Number)
				break
			}
		}
	}
	first := title.Seasons[0].Episodes[0]
	if first.Title != "Piloto" || first.PlayerURL != "https://azullog.top/serie/1396/1/1" {
		t.Errorf("S1E1 = %+v", first)
	}
	if got := title.EpisodeTitle(1, 1); got != "Piloto" {
		t.Errorf("EpisodeTitle(1,1) = %q", got)
	}
	if got := title.EpisodeTitle(9, 1); got != "" {
		t.Errorf("EpisodeTitle of a missing season = %q, want empty", got)
	}

	if _, err := c.Title(context.Background(), page); err != nil {
		t.Fatal(err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("title page fetched %d times, want 1 (cached)", n)
	}
}

func TestClientTitleMovie(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture(t, "title_movie_2026_10_09.html"))
	}))
	title, err := c.Title(context.Background(), "https://topcine.test/filme/zona-zero")
	if err != nil {
		t.Fatal(err)
	}
	if title.Kind != KindMovie || title.TMDBID != 1375646 || title.PlayerURL != "https://azullog.top/filme/1375646" {
		t.Errorf("title = %+v", title)
	}
}

func TestClientTitleWithoutPlayer(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// A trailer iframe and the site's own relative links are not players.
		_, _ = w.Write([]byte(`<html><body>
			<iframe src="https://www.youtube.com/embed/HhesaQXLuRY"></iframe>
			<a href="/filme/1917">1917</a>
			<option data-player="/serie/1/1/1">x</option>
		</body></html>`))
	}))
	_, err := c.Title(context.Background(), "https://topcine.test/filme/x")
	if !errors.Is(err, ErrNoPlayer) {
		t.Errorf("err = %v, want ErrNoPlayer", err)
	}
}

func TestClientLanguages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		body      string
		want      Languages
		wantError bool
	}{
		{
			name: "both tracks",
			body: `{"sucesso":true,"tem_dublado":true,"tem_legendado":true,"url_dublado":"https://1take.top/e/tvtmdb1396t1e1dub","url_legendado":"https://1take.top/e/tvtmdb1396t1e1leg"}`,
			want: Languages{Dubbed: true, Subtitled: true, Hosts: []string{"1take.top"}},
		},
		{
			name: "subtitled only",
			body: `{"sucesso":true,"tem_dublado":false,"tem_legendado":true,"url_dublado":null,"url_legendado":"https://1take.top/e/tmdb1375646leg"}`,
			want: Languages{Subtitled: true, Hosts: []string{"1take.top"}},
		},
		{name: "refused", body: `{"sucesso":false}`, wantError: true},
		{name: "not json", body: `<html>`, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotAction, gotAjax string
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAction, gotAjax = r.URL.Query().Get("acao"), r.Header.Get("X-Requested-With")
				_, _ = w.Write([]byte(tt.body))
			}))
			got, err := c.Languages(context.Background(), "https://azullog.top/serie/1396/1/1")
			if gotAction != "carregar_idiomas" || gotAjax != "XMLHttpRequest" {
				t.Errorf("request acao=%q X-Requested-With=%q", gotAction, gotAjax)
			}
			if tt.wantError {
				if err == nil {
					t.Errorf("got %+v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Dubbed != tt.want.Dubbed || got.Subtitled != tt.want.Subtitled || !slices.Equal(got.Hosts, tt.want.Hosts) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestClientLanguagesInvalidURL(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(http.DefaultClient, "https://topcine.test")
	if _, err := c.Languages(context.Background(), "/serie/1/1/1"); err == nil {
		t.Error("relative player URL accepted")
	}
}

func TestMediaToAnimeModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind MediaKind
		want models.MediaType
	}{
		{KindMovie, models.MediaTypeMovie},
		{KindSeries, models.MediaTypeTV},
	}
	for _, tt := range tests {
		m := Media{Title: "T", URL: "https://topcine.test/x", Poster: "p", Year: "2020", Kind: tt.kind}
		a := m.ToAnimeModel()
		if a.Source != SourceName || a.MediaType != tt.want || a.Name != "T" || a.URL != m.URL || a.ImageURL != "p" || a.Year != "2020" {
			t.Errorf("%s -> %+v", tt.kind, a)
		}
	}
}
