package anroll

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
)

func TestIsAniDrivePlayer(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"https://anidrive.click/token/hv0688292sn4his":  true,
		"https://ANIDRIVE.click/token/abc":              true,
		"https://evil.example/?x=anidrive.click/token/": false,
		"https://evil.example/anidrive.click/token/abc": false,
		"https://anidrive.click.evil.example/token/abc": false,
		"https://static.anidrive.click/token/abc":       false,
		"https://anidrive.click/embed/abc":              false,
		"":                                              false,
	}
	for raw, want := range tests {
		if got := isAniDrivePlayer(raw); got != want {
			t.Errorf("isAniDrivePlayer(%q) = %v, want %v", raw, got, want)
		}
	}
}

func TestSearchAnimeKeepsSeriesAndMovieWithSameTitle(t *testing.T) {
	t.Parallel()
	const page = `<html><body>
<article class="anime-card"><a href="https://anroll.io/anime/akira/"><img src="a.jpg" alt="Akira"><span class="card-type">TV</span><span class="card-sub">Legendado</span></a></article>
<article class="anime-card"><a href="https://anroll.io/anime/akira-filme/"><img src="b.jpg" alt="Akira"><span class="card-type">Filme</span><span class="card-sub">Legendado</span></a></article>
</body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(page)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a constant page
	}))
	defer srv.Close()
	results, err := NewClientForTest(srv.Client(), srv.URL).SearchAnime("akira")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want the series and the movie", len(results))
	}
	if results[0].MediaType != models.MediaTypeAnime || results[1].MediaType != models.MediaTypeMovie {
		t.Fatalf("media types = %v, %v", results[0].MediaType, results[1].MediaType)
	}
}
