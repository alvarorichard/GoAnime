package api

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveStartFlixOfficialNames drives StartFlix search → panel → official
// record against the live sites, with no API key of any kind, and checks the
// Plex/Jellyfin path a download lands on. Opt in with GOANIME_LIVE=1; never
// runs in CI.
func TestLiveStartFlixOfficialNames(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("OMDB_API_KEY", "")
	util.InitLogger()

	tests := []struct {
		query, wantPath string
	}{
		{"Coração Selvagem", "movies/Heart of the Beast (2026) {tmdb-1263337} {imdb-tt7526136}/Heart of the Beast (2026).mp4"},
		{"Velozes e Furiosos 5", "movies/Fast Five (2011) {tmdb-51497} {imdb-tt1596343}/Fast Five (2011).mp4"},
	}
	c := startflix.Shared()
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		results, err := c.Search(ctx, tt.query)
		cancel()
		if err != nil {
			t.Fatalf("%q: search: %v", tt.query, err)
		}
		var media *startflix.Media
		for i := range results {
			if results[i].Kind == startflix.KindMovie {
				media = &results[i]
				break
			}
		}
		if media == nil {
			t.Fatalf("%q: no movie in the results", tt.query)
		}
		anime := media.ToAnimeModel()
		if _, err := GetStartFlixEpisodes(anime); err != nil {
			t.Fatalf("%q: listing: %v", tt.query, err)
		}
		meta := &util.MediaMeta{OfficialTitle: anime.OfficialTitle(), Year: anime.Year, TMDBID: anime.TMDBID, IMDBID: anime.IMDBID}
		got := util.FormatPlexMoviePath("movies", anime.Name, "", meta)
		t.Logf("%q -> %s", anime.Name, got)
		if got != tt.wantPath {
			t.Errorf("%q lands at %q, want %q", anime.Name, got, tt.wantPath)
		}
	}
}
