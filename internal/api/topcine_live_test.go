package api

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveTopCinePlays drives TopCine search → title page → StartFlix panel by
// id → stream for a series and a movie, keyless, against the live sites. Opt
// in with GOANIME_LIVE=1; never runs in CI.
func TestLiveTopCinePlays(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("OMDB_API_KEY", "")
	util.InitLogger()
	prevPick := sfxPickFn
	sfxPickFn = func(string, []string) (int, error) { return 0, nil } // Dublado
	t.Cleanup(func() { sfxPickFn = prevPick })

	tests := []struct {
		query, title string
		kind         topcine.MediaKind
	}{
		{"breaking bad", "Breaking Bad", topcine.KindSeries},
		{"zona zero", "Zona Zero", topcine.KindMovie},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			results, err := topcine.Shared().Search(ctx, tt.query)
			cancel()
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			var anime *models.Anime
			for _, r := range results {
				if r.Kind == tt.kind && r.Title == tt.title {
					anime = r.ToAnimeModel()
					break
				}
			}
			if anime == nil {
				t.Fatalf("%q is not in the results", tt.title)
			}

			var eps []models.Episode
			if tt.kind == topcine.KindSeries {
				eps, err = GetTopCineSeasonEpisodes(anime, 1)
			} else {
				eps, err = GetTopCineEpisodes(anime)
			}
			if err != nil {
				t.Fatalf("listing: %v", err)
			}
			if len(eps) == 0 {
				t.Fatal("nothing listed")
			}
			t.Logf("%s: tmdb %d imdb %q official %q, %d episode(s), first %q at %s",
				anime.Name, anime.TMDBID, anime.IMDBID, anime.OfficialTitle(), len(eps), eps[0].Title.English, eps[0].URL)

			streamURL, err := GetTopCineStreamURL(anime, &eps[0], "best")
			if err != nil {
				t.Fatalf("stream: %v", err)
			}
			if !strings.HasPrefix(streamURL, "http") {
				t.Fatalf("stream URL = %q", streamURL)
			}
			t.Logf("stream: %s", streamURL[:min(len(streamURL), 100)])
		})
	}
}
