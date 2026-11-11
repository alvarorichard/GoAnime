package animesfhd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestDefaultHost20261111(t *testing.T) {
	t.Parallel()
	c := NewAnimesFHDClient()
	if c.baseURL != "https://animesfhd.net" || c.http == nil || c.userAgent == "" {
		t.Fatalf("unexpected client defaults: base=%q", c.baseURL)
	}
}

func TestSearchAnime(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, query, file, title, id, audio string
		count int
	}{
		{"search", " frieren ", "search_2026_11_11.json", "Sousou no Frieren", "2751618645542326", "Dubbed and subtitled", 1},
		{"popular", "", "popular_2026_11_11.json", "", "", "", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := fixture(t, tc.file)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if r.URL.Path != "/api/series" || q.Get("q") != strings.TrimSpace(tc.query) || q.Get("pagination") != "1" || q.Get("page") != "1" || q.Get("per_page") != "50" {
					t.Errorf("unexpected search request: %s", r.URL)
				}
				if tc.query == "" && q.Get("sort") != "popular" {
					t.Errorf("empty query should list popular titles")
				}
				if r.Header.Get("Referer") != DefaultBase+"/" || r.Header.Get("User-Agent") == "" {
					t.Error("missing request headers")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			c := NewClientForTest(srv.Client(), srv.URL)
			got, err := c.SearchAnime(tc.query)
			if err != nil || len(got) != tc.count {
				t.Fatalf("results=%d err=%v", len(got), err)
			}
			if tc.title != "" && (got[0].Name != tc.title || got[0].URL != srv.URL+"/serie/"+tc.id || got[0].Audio != tc.audio || got[0].Source != SourceName) {
				t.Fatalf("result=%+v", got[0])
			}
		})
	}
}

func TestGetAnimeEpisodesFixtures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, file, id, first string
		count int
	}{
		{"series", "series_2026_11_11.json", "1709160706373932", "22882", 6},
		{"movie", "movie_2026_11_11.json", "2110072639862535", "1049086484300956", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := fixture(t, tc.file)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/series/"+tc.id || r.URL.Query().Get("per_page") != "200" {
					t.Errorf("request=%s", r.URL)
				}
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			got, err := NewClientForTest(srv.Client(), srv.URL).GetAnimeEpisodes(srv.URL+"/serie/"+tc.id)
			if err != nil || len(got) != tc.count {
				t.Fatalf("episodes=%d err=%v", len(got), err)
			}
			if got[0].URL != srv.URL+"/episod/"+tc.first || got[0].Number != "1" || got[0].Num != 1 || got[0].Audio != "Legendado" {
				t.Fatalf("first episode=%+v", got[0])
			}
		})
	}
}

func TestEpisodePagination(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, second string
		wantErr bool
	}{
		{"sort and deduplicate", `{"series":{"id":1},"episodes":[{"id":9,"episode_number":2},{"id":8,"episode_number":1.5}],"has_more":false}`, false},
		{"repeated page", `{"series":{"id":1},"episodes":[{"id":9,"episode_number":2}],"has_more":true}`, true},
		{"empty continuation", `{"series":{"id":1},"episodes":[],"has_more":true}`, true},
		{"wrong series", `{"series":{"id":2},"episodes":[{"id":8,"episode_number":1}],"has_more":false}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := `{"series":{"id":1},"episodes":[{"id":9,"episode_number":2}],"has_more":true}`
				if r.URL.Query().Get("page") != "1" {
					body = tc.second
				}
				_, _ = w.Write([]byte(body))
			}))
			defer srv.Close()
			got, err := NewClientForTest(srv.Client(), srv.URL).GetAnimeEpisodes(srv.URL+"/serie/1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("episodes=%v err=%v", got, err)
			}
			if !tc.wantErr && (len(got) != 2 || got[0].Number != "1.5" || got[1].Num != 2) {
				t.Fatalf("episodes=%v", got)
			}
		})
	}
}

func TestVersionsKeepSeasonPartAndAudio(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"series":{"id":1},"versions":[{"series_id":1,"title":"Show","season":1,"part":0,"season_label":"Temporada 1","audio":"legendado"},{"series_id":2,"title":"Show","season":2,"part":1,"season_label":"Temporada 2 - Parte 1","audio":"dublado"},{"series_id":0}],"episodes":[]}`))
	}))
	defer srv.Close()
	got, err := NewClientForTest(srv.Client(), srv.URL).GetVersionsContext(context.Background(), srv.URL+"/serie/1")
	if err != nil || len(got) != 2 || got[1].URL != srv.URL+"/serie/2" || got[1].Label != "Temporada 2 - Parte 1 — Dublado" || got[1].Season != 2 {
		t.Fatalf("versions=%+v err=%v", got, err)
	}
}

func TestGetEpisodeStreamFixtures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, file, id, path, subtitle string
	}{
		{"s6 HLS", "episode_2026_11_11.json", "22882", "/api/v/s6?source=2016384480747057&playlist=.m3u8", ""},
		{"movie HLS with captions", "movie_episode_2026_11_11.json", "1049086484300956", "/api/v/s4?source=2591675809550102&playlist=.m3u8", "/api/v/s4?source=2591675809550102&subtitle=1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := fixture(t, tc.file)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/v/") {
					w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
					_, _ = w.Write([]byte("#EXTM3U\n"))
					return
				}
				if r.URL.Path != "/api/episodes/"+tc.id {
					t.Errorf("unexpected request: %s", r.URL)
				}
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			c := NewClientForTest(srv.Client(), srv.URL)
			got, err := c.GetEpisodeStreamContext(context.Background(), srv.URL+"/episod/"+tc.id)
			if err != nil || got.URL != srv.URL+tc.path || got.Referer != DefaultBase+"/" {
				t.Fatalf("stream=%+v err=%v", got, err)
			}
			if tc.subtitle != "" && (len(got.Subtitles) != 1 || got.Subtitles[0].URL != srv.URL+tc.subtitle) {
				t.Fatalf("subtitles=%+v", got.Subtitles)
			}
			url, err := c.GetEpisodeStreamURL(srv.URL+"/episod/"+tc.id)
			if err != nil || url != got.URL {
				t.Fatalf("URL wrapper=%q err=%v", url, err)
			}
		})
	}
}

func TestStreamFiltersAndRedirects(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, sources, contentType, target string
		wantErr bool
	}{
		{"native MP4 after embed and failed player", `{"native_video":true,"kind":"embed","url":"/forbidden"},{"native_video":false,"kind":"file","url":"/forbidden"},{"native_video":true,"kind":"file","url":"/missing"},{"native_video":true,"kind":"file","url":"/api/v/s5?source=4","audio":"dublado"}`, "video/mp4", "/media/video.mp4", false},
		{"redirect to HLS", `{"native_video":true,"kind":"file","url":"/api/v/s5?source=4"}`, "application/vnd.apple.mpegurl", "/media/master.m3u8", false},
		{"failed HLS falls through", `{"native_video":true,"kind":"file","url":"/missing.m3u8"},{"native_video":true,"kind":"file","url":"/media/master.m3u8"}`, "application/vnd.apple.mpegurl", "/media/master.m3u8", false},
		{"embed only", `{"native_video":true,"kind":"embed","url":"/forbidden"}`, "", "", true},
		{"unsafe URL", `{"native_video":true,"kind":"file","url":"javascript:alert(1)"}`, "", "", true},
		{"HTML player", `{"native_video":true,"kind":"file","url":"/api/v/s5?source=4"}`, "text/html", "/media/page", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Referer") != DefaultBase+"/" || r.Header.Get("User-Agent") == "" {
					t.Error("headers missing from API/redirect/media request")
				}
				switch r.URL.Path {
				case "/api/episodes/1":
					_, _ = w.Write([]byte(`{"episode":{"id":1},"sources":[`+tc.sources+`]}`))
				case "/api/v/s5":
					http.Redirect(w, r, tc.target, http.StatusTemporaryRedirect)
				case tc.target:
					w.Header().Set("Content-Type", tc.contentType)
					_, _ = w.Write([]byte("media"))
				case "/missing", "/missing.m3u8":
					http.Error(w, "unavailable", http.StatusNotFound)
				default:
					t.Errorf("must not request %s", r.URL)
					http.NotFound(w, r)
				}
			}))
			defer srv.Close()
			got, err := NewClientForTest(srv.Client(), srv.URL).GetEpisodeStreamContext(context.Background(), srv.URL+"/episod/1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("stream=%+v err=%v", got, err)
			}
			if !tc.wantErr && got.URL != srv.URL+tc.target {
				t.Fatalf("stream=%+v", got)
			}
			if tc.name == "native MP4 after embed and failed player" && got.AudioLang != "por" {
				t.Fatalf("dubbed audio=%q", got.AudioLang)
			}
		})
	}
}

func TestAPIErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status int
		kind netx.DiagnosticKind
	}{
		{"HTTP status", `{}`, 503, netx.DiagnosticSourceUnavailable},
		{"challenge", `<html>Just a moment</html>`, 200, netx.DiagnosticBlockedChallenge},
		{"malformed JSON", `{`, 200, netx.DiagnosticParserBroken},
		{"missing list", `{}`, 200, netx.DiagnosticParserBroken},
		{"null list", `{"series":null}`, 200, netx.DiagnosticParserBroken},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			_, err := NewClientForTest(srv.Client(), srv.URL).SearchAnime("test")
			if err == nil {
				t.Fatal("expected a typed API error")
			}
			if diagnostic := netx.DiagnoseError(SourceName, "search", err); diagnostic.Kind != tc.kind {
				t.Fatalf("diagnostic=%s want=%s err=%v", diagnostic.Kind, tc.kind, err)
			}
		})
	}
}

func TestInvalidPageURLsAndCancellation(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("invalid or cancelled input must not reach HTTP")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	for _, input := range []string{"https://other.example/serie/1", srv.URL+"/serie/abc", srv.URL+"/serie/-1", srv.URL+"/episod/1", srv.URL+"/serie/9223372036854775808"} {
		if _, err := c.GetAnimeEpisodes(input); err == nil {
			t.Errorf("accepted invalid title URL %q", input)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.SearchAnimeContext(ctx, "test"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled search: %v", err)
	}
	if _, err := c.GetAnimeEpisodesContext(ctx, srv.URL+"/serie/1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled episodes: %v", err)
	}
	if _, err := c.GetEpisodeStreamContext(ctx, srv.URL+"/episod/1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled stream: %v", err)
	}
}

func TestSearchMovieAndQueryEscaping(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "Fate & ação +" {
			t.Errorf("query=%s", r.URL)
		}
		_, _ = w.Write([]byte(`{"series":[{"id":2110072639862535,"title":"Fate","category":"filme","year":2017}],"has_more":false}`))
	}))
	defer srv.Close()
	got, err := NewClientForTest(srv.Client(), srv.URL).SearchAnime("Fate & ação +")
	if err != nil || len(got) != 1 || got[0].MediaType != models.MediaTypeMovie || got[0].Year != "2017" {
		t.Fatalf("results=%+v err=%v", got, err)
	}
}
