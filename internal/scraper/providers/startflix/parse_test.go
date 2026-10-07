package startflix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Fixtures are the live pages captured on 2026-10-07, with scripts, styles and
// SEO filler removed. The date in each name is when the markup was current, so
// a later site change shows up as a failing fixture rather than a silent drift.

func loadDoc(t *testing.T, name string) *goquery.Document {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return docFrom(t, string(raw))
}

func docFrom(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return doc
}

func TestParseSearchResults(t *testing.T) {
	t.Parallel()
	got := parseSearchResults(loadDoc(t, "search_2026_10_07.html"))

	want := []Media{
		{Title: "Boruto: Naruto Next Generations", URL: "https://www.startflix.biz/series/boruto-naruto-next-generations/", Kind: KindSeries},
		{Title: "Naruto Shippuden", URL: "https://www.startflix.biz/series/naruto-shippuden/", Kind: KindSeries},
		{Title: "Naruto", URL: "https://www.startflix.biz/series/naruto/", Kind: KindSeries},
		{Title: "Oppenheimer", URL: "https://www.startflix.biz/filmes/oppenheimer/", Kind: KindMovie, Year: "2023"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		g := got[i]
		if g.Title != w.Title || g.URL != w.URL || g.Kind != w.Kind || g.Year != w.Year {
			t.Errorf("result %d = %+v, want %+v", i, g, w)
		}
		if !strings.Contains(g.Poster, "/t/p/w500/") {
			t.Errorf("result %d poster %q was not upgraded from the 92px thumbnail", i, g.Poster)
		}
	}
}

func TestParseSearchResultsSkipsNonTitlesAndDuplicates(t *testing.T) {
	t.Parallel()
	html := `<div class="result-item"><article><div class="details"><div class="title">
		<a href="https://www.startflix.biz/episodios/x-1x1/">Episode</a></div></div></article></div>
	<div class="result-item"><article><div class="details"><div class="title">
		<a href="https://www.startflix.biz/filmes/a/">A</a></div></div></article></div>
	<div class="result-item"><article><div class="details"><div class="title">
		<a href="https://www.startflix.biz/filmes/a/">A again</a></div></div></article></div>`
	got := parseSearchResults(docFrom(t, html))
	if len(got) != 1 || got[0].Title != "A" {
		t.Fatalf("got %+v, want only the first /filmes/a/", got)
	}
}

func TestKindFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url  string
		kind MediaKind
		ok   bool
	}{
		{"https://www.startflix.biz/filmes/oppenheimer/", KindMovie, true},
		{"https://www.startflix.biz/series/naruto/", KindSeries, true},
		{"https://www.startflix.biz/episodios/naruto-1x1/", "", false},
		{"https://www.startflix.biz/generos/acao/", "", false},
		{"::not a url", "", false},
	}
	for _, tt := range tests {
		kind, ok := kindFromURL(tt.url)
		if kind != tt.kind || ok != tt.ok {
			t.Errorf("kindFromURL(%q) = (%q, %v), want (%q, %v)", tt.url, kind, ok, tt.kind, tt.ok)
		}
	}
}

func TestPanelFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want Panel
		ok   bool
	}{
		{"series", "https://www.painel-aso.sbs/embed/194583", Panel{URL: "https://www.painel-aso.sbs/embed/194583", Kind: KindSeries, TMDBID: 194583}, true},
		{"series trailing slash", "https://www.painel-aso.sbs/embed/194583/", Panel{URL: "https://www.painel-aso.sbs/embed/194583", Kind: KindSeries, TMDBID: 194583}, true},
		{"movie", "https://www.painel-aso.sbs/filme/tt7526136", Panel{URL: "https://www.painel-aso.sbs/filme/tt7526136", Kind: KindMovie, IMDBID: "tt7526136"}, true},
		{"youtube trailer", "https://www.youtube.com/embed/F0jWGqOyQ8g?autoplay=0", Panel{}, false},
		{"movie without imdb id", "https://www.painel-aso.sbs/filme/7526136", Panel{}, false},
		{"relative", "/embed/194583", Panel{}, false},
		{"javascript", "javascript:alert(1)", Panel{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := panelFromURL(tt.url)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("panelFromURL(%q) = (%+v, %v), want (%+v, %v)", tt.url, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestParsePanel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		doc  func(t *testing.T) *goquery.Document
		want string
		ok   bool
	}{
		{"series page", func(t *testing.T) *goquery.Document { return loadDoc(t, "title_series_2026_10_07.html") }, "https://www.painel-aso.sbs/embed/194583", true},
		{"movie page", func(t *testing.T) *goquery.Document { return loadDoc(t, "title_movie_2026_10_07.html") }, "https://www.painel-aso.sbs/filme/tt7526136", true},
		{"trailer before panel", func(t *testing.T) *goquery.Document {
			return docFrom(t, `<iframe src="https://www.youtube.com/embed/abc"></iframe><iframe src="https://p.example/filme/tt1"></iframe>`)
		}, "https://p.example/filme/tt1", true},
		{"no panel", func(t *testing.T) *goquery.Document {
			return docFrom(t, `<iframe src="https://www.youtube.com/embed/abc"></iframe>`)
		}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parsePanel(tt.doc(t))
			if ok != tt.ok || got.URL != tt.want {
				t.Fatalf("parsePanel = (%q, %v), want (%q, %v)", got.URL, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestParseSeasons(t *testing.T) {
	t.Parallel()
	seasons := parseSeasons(loadDoc(t, "panel_series_2026_10_07.html"))

	type counts struct{ number, dub, sub int }
	want := []counts{{1, 6, 6}, {2, 8, 3}, {3, 0, 7}}
	if len(seasons) != len(want) {
		t.Fatalf("got %d seasons, want %d", len(seasons), len(want))
	}
	for i, w := range want {
		s := seasons[i]
		if s.Number != w.number || len(s.Dubbed) != w.dub || len(s.Subtitled) != w.sub {
			t.Errorf("season %d = {n=%d dub=%d sub=%d}, want %+v", i, s.Number, len(s.Dubbed), len(s.Subtitled), w)
		}
		for _, list := range [][]Episode{s.Dubbed, s.Subtitled} {
			for j, ep := range list {
				if ep.Number != j+1 {
					t.Errorf("season %d: episode at %d numbered %d", s.Number, j, ep.Number)
				}
			}
		}
	}

	s1 := seasons[0]
	if got := s1.Dubbed[0]; got.ID != "228789" || got.Audio != AudioDubbed || got.Title != "" {
		t.Errorf("S1 dub E1 = %+v, want id 228789, dubbed, generic title dropped", got)
	}
	if got := s1.Subtitled[0]; got.ID != "211262" || got.Audio != AudioSubtitled || got.Title != "Velhos Conhecimentos" {
		t.Errorf("S1 sub E1 = %+v, want id 211262 titled Velhos Conhecimentos", got)
	}
	if got := seasons[2].Audios(); len(got) != 1 || got[0] != AudioSubtitled {
		t.Errorf("S3 audios = %v, want subtitled only", got)
	}
	if got := s1.Audios(); len(got) != 2 || got[0] != AudioDubbed {
		t.Errorf("S1 audios = %v, want dub first", got)
	}
}

func TestParseSeasonsDropsEmptyAndUnknown(t *testing.T) {
	t.Parallel()
	html := `<ul class="header-navigation">
		<li data-season-id="1" data-season-number="1">1</li>
		<li data-season-id="2" data-season-number="2">2</li>
		<li data-season-id="x" data-season-number="?">bad</li>
	</ul>
	<div class="cards">
		<div class="card"><h2 class="card-title">Dublado</h2><ul>
			<li data-season-id="1" data-episode-id="11"><a>2 - Episódio</a></li>
			<li data-season-id="1" data-episode-id="10"><a>1 - Episódio</a></li>
			<li data-season-id="9" data-episode-id="99"><a>1 - Orphan</a></li>
		</ul></div>
		<div class="card"><h2 class="card-title">Trailer</h2><ul>
			<li data-season-id="2" data-episode-id="20"><a>1</a></li>
		</ul></div>
	</div>`
	seasons := parseSeasons(docFrom(t, html))
	if len(seasons) != 1 || seasons[0].Number != 1 {
		t.Fatalf("got %+v, want only season 1 (season 2 has no recognised card)", seasons)
	}
	if ids := []string{seasons[0].Dubbed[0].ID, seasons[0].Dubbed[1].ID}; ids[0] != "10" || ids[1] != "11" {
		t.Errorf("episodes not sorted by number: %v", ids)
	}
}

func TestEpisodeNumberAndTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		label    string
		position int
		number   int
		title    string
	}{
		{"1 - Episódio", 9, 1, ""},
		{"12 - Velhos Conhecimentos", 9, 12, "Velhos Conhecimentos"},
		{"Episódio 7", 9, 7, ""},
		{"Episodio 3", 9, 3, ""},
		{"Especial de Natal", 4, 4, "Especial de Natal"},
		{"", 2, 2, ""},
	}
	for _, tt := range tests {
		if got := episodeNumber(tt.label, tt.position); got != tt.number {
			t.Errorf("episodeNumber(%q) = %d, want %d", tt.label, got, tt.number)
		}
		if got := episodeTitle(tt.label); got != tt.title {
			t.Errorf("episodeTitle(%q) = %q, want %q", tt.label, got, tt.title)
		}
	}
}

func TestParsePlayers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fixture string
		hosts   []string
		direct  []bool
	}{
		{
			"episode_2026_10_07.html",
			[]string{"playembedapi.site", "embedplayapiupn.upns.xyz", "apiblogger.click", "vidsrcme.su"},
			[]bool{false, false, true, false},
		},
		{
			"panel_movie_2026_10_07.html",
			[]string{"playembedapi.site", "embedplaybyse.top", "vidsrcme.su"},
			[]bool{false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			t.Parallel()
			players := parsePlayers(loadDoc(t, tt.fixture))
			if len(players) != len(tt.hosts) {
				t.Fatalf("got %d players, want %d", len(players), len(tt.hosts))
			}
			for i, p := range players {
				if !strings.Contains(p.URL, tt.hosts[i]) {
					t.Errorf("player %d URL %q, want host %s", i, p.URL, tt.hosts[i])
				}
				if p.IsDirectFile() != tt.direct[i] {
					t.Errorf("player %d (%s) direct = %v, want %v", i, p.Type, p.IsDirectFile(), tt.direct[i])
				}
				if p.Label == "" || p.ID == "" {
					t.Errorf("player %d missing label or id: %+v", i, p)
				}
			}
		})
	}
}

func TestIsPanelNotFound(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "panel_notfound_2026_10_07.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !isPanelNotFound(raw) {
		t.Error("the panel's not-found body was not recognised")
	}
	page, err := os.ReadFile(filepath.Join("testdata", "panel_series_2026_10_07.html"))
	if err != nil {
		t.Fatal(err)
	}
	if isPanelNotFound(page) {
		t.Error("a real panel page was taken for not-found")
	}
}

func TestPanelEpisodeURL(t *testing.T) {
	t.Parallel()
	p := Panel{URL: "https://www.painel-aso.sbs/embed/194583"}
	if got := p.EpisodeURL("299174"); got != "https://www.painel-aso.sbs/episodio/299174" {
		t.Errorf("EpisodeURL = %q", got)
	}
}

func TestMediaToAnimeModel(t *testing.T) {
	t.Parallel()
	movie := Media{Title: "Oppenheimer", URL: "https://www.startflix.biz/filmes/oppenheimer/", Kind: KindMovie, Year: "2023"}.ToAnimeModel()
	if movie.MediaType != "movie" || movie.Source != SourceName || movie.URL == "" || movie.Year != "2023" {
		t.Errorf("movie model = %+v", movie)
	}
	series := Media{Title: "Naruto", URL: "https://www.startflix.biz/series/naruto/", Kind: KindSeries}.ToAnimeModel()
	if series.MediaType != "tv" {
		t.Errorf("series media type = %q, want tv", series.MediaType)
	}
}
