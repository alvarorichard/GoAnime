package topcine

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func docOf(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestParseSearchResultsSkipsWhatIsNotATitle(t *testing.T) {
	t.Parallel()
	base, _ := url.Parse("https://topcine.test/")
	card := func(href, title, small string) string {
		return `<article class="pagina-busca-card"><a href="` + href + `"><img src="https://image.tmdb.org/t/p/w342/p.jpg"><h2>` +
			title + `</h2><small>` + small + `</small></a></article>`
	}
	doc := docOf(t, card("http://[::1", "Bad URL", "2020")+
		card("/categoria/acao", "A category", "")+
		card("/filme/", "Bare prefix", "")+
		card("/serie/x", "   ", "")+
		card("/serie/x", "  Série   X  ", "Série")+
		card("https://topcine.test/serie/x", "Duplicate", "2020")+
		card("/filme/y", "Y", "1999 • Filme"))

	got := parseSearchResults(doc, base)
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2: %+v", len(got), got)
	}
	if got[0].Title != "Série X" || got[0].URL != "https://topcine.test/serie/x" || got[0].Year != "" || got[0].Kind != KindSeries {
		t.Errorf("first = %+v, want whitespace collapsed and no year", got[0])
	}
	if got[1].Year != "1999" || got[1].Kind != KindMovie {
		t.Errorf("second = %+v", got[1])
	}
}

func TestKindFromPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want MediaKind
		ok   bool
	}{
		{"/filme/zona-zero", KindMovie, true},
		{"/serie/breaking-bad", KindSeries, true},
		{"/filme/", "", false},
		{"/serie/", "", false},
		{"/filmes", "", false},
		{"/categoria/acao", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := kindFromPath(tt.path)
		if got != tt.want || ok != tt.ok {
			t.Errorf("kindFromPath(%q) = (%q, %v), want (%q, %v)", tt.path, got, ok, tt.want, tt.ok)
		}
	}
}

func TestPlayerURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw string
		ok  bool
	}{
		{"https://azullog.top/filme/1", true},
		{" http://azullog.top/filme/1 ", true},
		{"/filme/1", false},
		{"//azullog.top/filme/1", false},
		{"javascript:alert(1)", false},
		{"ftp://azullog.top/filme/1", false},
		{"http://[::1", false},
		{"", false},
	}
	for _, tt := range tests {
		if _, ok := playerURL(tt.raw); ok != tt.ok {
			t.Errorf("playerURL(%q) ok = %v, want %v", tt.raw, ok, tt.ok)
		}
	}
}

func TestEpisodeTitle(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"Piloto":            "Piloto",
		"  O   Tormento  ":  "O Tormento",
		"Episódio 3":        "",
		"episodio":          "",
		"EPISÓDIO 12":       "",
		"Episódio 3: Volta": "Episódio 3: Volta",
		"":                  "",
	}
	for raw, want := range tests {
		if got := episodeTitle(raw); got != want {
			t.Errorf("episodeTitle(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestParseTitleSeriesSkipsBadOptions(t *testing.T) {
	t.Parallel()
	doc := docOf(t, `<select>
		<option data-player="/serie/1/1/1">relative</option>
		<option data-player="https://azullog.top/filme/1">a movie player</option>
		<option data-player="https://azullog.top/serie/0/1/1">no tmdb id</option>
		<option data-player="https://azullog.top/serie/5/1/0">no episode number</option>
		<option data-player="https://azullog.top/serie/5/2/2" data-titulo="Dois">ok</option>
		<option data-player="https://other.host/serie/5/2/1/" data-titulo="Um">ok, rotated host</option>
		<option data-player="https://azullog.top/serie/6/1/1">another title's player is skipped</option>
	</select>`)
	title, ok := parseTitle(doc)
	if !ok {
		t.Fatal("no title parsed")
	}
	if title.Kind != KindSeries || title.TMDBID != 5 {
		t.Fatalf("title = %v/%d, want series 5", title.Kind, title.TMDBID)
	}
	if len(title.Seasons) != 1 || title.Seasons[0].Number != 2 {
		t.Fatalf("seasons = %+v, want season 2 only (season 1's players are bad or another title's)", title.Seasons)
	}
	s2 := title.Seasons[0].Episodes
	if len(s2) != 2 || s2[0].Number != 1 || s2[0].Title != "Um" || s2[1].Number != 2 {
		t.Errorf("season 2 = %+v, want episodes sorted", s2)
	}
}

func TestParseTitleMovieSkipsOtherIframes(t *testing.T) {
	t.Parallel()
	doc := docOf(t, `<body>
		<iframe src="/filme/1"></iframe>
		<iframe src="https://www.youtube.com/embed/abc"></iframe>
		<iframe src="https://azullog.top/filme/0"></iframe>
		<iframe src="https://azullog.top/filme/42/"></iframe>
		<iframe src="https://azullog.top/filme/43"></iframe>
	</body>`)
	title, ok := parseTitle(doc)
	if !ok || title.Kind != KindMovie || title.TMDBID != 42 || title.PlayerURL != "https://azullog.top/filme/42/" {
		t.Errorf("title = %+v (ok %v), want the first movie player, tmdb 42", title, ok)
	}
	if len(title.Seasons) != 0 {
		t.Errorf("a movie has no seasons: %+v", title.Seasons)
	}
}

func TestParseTitleNothing(t *testing.T) {
	t.Parallel()
	if title, ok := parseTitle(docOf(t, `<p>Em breve</p>`)); ok {
		t.Errorf("parsed %+v from a page with no player", title)
	}
}

func TestTitleEpisodeTitle(t *testing.T) {
	t.Parallel()
	title := Title{Seasons: []Season{
		{Number: 1, Episodes: []Episode{{Number: 1, Title: "A"}, {Number: 2}}},
		{Number: 2, Episodes: []Episode{{Number: 1, Title: "B"}}},
	}}
	tests := []struct {
		season, episode int
		want            string
	}{
		{1, 1, "A"},
		{1, 2, ""},
		{2, 1, "B"},
		{1, 3, ""},
		{3, 1, ""},
	}
	for _, tt := range tests {
		if got := title.EpisodeTitle(tt.season, tt.episode); got != tt.want {
			t.Errorf("EpisodeTitle(%d, %d) = %q, want %q", tt.season, tt.episode, got, tt.want)
		}
	}
}
