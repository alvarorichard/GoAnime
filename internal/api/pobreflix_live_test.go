package api

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLivePobreflixPlays drives Pobreflix search → title page → StartFlix
// panel by id → stream for a series and a movie, keyless, against the live
// sites. The catalog comes from Pobreflix's sitemap corpus; the stream from
// the panel's player hosts. Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLivePobreflixPlays(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("OMDB_API_KEY", "")
	util.InitLogger()
	prevPick := sfxPickFn
	sfxPickFn = func(string, []string) (int, error) { return 0, nil } // Dublado
	t.Cleanup(func() { sfxPickFn = prevPick })

	// Titles verified present in the site's own sitemap corpus (probe,
	// 2026-10-16): a series and a movie, exactly as the catalog names them.
	tests := []struct {
		query, title string
		kind         pobreflix.MediaKind
	}{
		{"ben hur serie", "Ben Hur A Serie", pobreflix.KindSeries},
		{"interestelar", "Interestelar", pobreflix.KindMovie},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			results, err := pobreflix.Shared().Search(ctx, tt.query)
			cancel()
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			var anime *models.Anime
			for _, r := range results {
				t.Logf("result: %s (%s) %s", r.Title, r.Kind, r.URL)
				if r.Kind == tt.kind && r.Title == tt.title {
					anime = r.ToAnimeModel()
					break
				}
			}
			if anime == nil {
				t.Fatalf("%q is not in the results", tt.title)
			}

			var eps []models.Episode
			if tt.kind == pobreflix.KindSeries {
				eps, err = GetPobreflixSeasonEpisodes(anime, 1)
			} else {
				eps, err = GetPobreflixEpisodes(anime)
			}
			if err != nil {
				t.Fatalf("listing: %v", err)
			}
			if len(eps) == 0 {
				t.Fatal("nothing listed")
			}
			t.Logf("%s: tmdb %d imdb %q official %q, %d episode(s), first %q at %s",
				anime.Name, anime.TMDBID, anime.IMDBID, anime.OfficialTitle(), len(eps), eps[0].Title.English, eps[0].URL)

			streamURL, err := GetPobreflixStreamURL(anime, &eps[0], "best")
			if err != nil {
				t.Fatalf("stream: %v", err)
			}
			if !strings.HasPrefix(streamURL, "http") {
				t.Fatalf("stream URL = %q", streamURL)
			}
			lim := len(streamURL)
			if lim > 100 {
				lim = 100
			}
			t.Logf("stream: %s", streamURL[:lim])
		})
	}
}
