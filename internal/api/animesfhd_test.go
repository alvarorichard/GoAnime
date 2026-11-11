package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesfhd"
)

func TestAnimesFHDVersionChoice(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, audio, wantPath string
		pick                  int
		cancel                bool
	}{
		{"subtitled first season", "", "/api/series/1", 0, false},
		{"dubbed second season", "", "/api/series/2", 1, false},
		{"pinned dubbed", "Dublado", "/api/series/2", 0, false},
		{"cancel must not play another version", "", "", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var paths []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				id := strings.TrimPrefix(r.URL.Path, "/api/series/")
				_, _ = w.Write([]byte(`{"series":{"id":` + id + `},"versions":[{"series_id":1,"season":1,"audio":"legendado"},{"series_id":2,"season":2,"audio":"dublado"}],"episodes":[{"id":` + id + `,"episode_number":1}],"has_more":false}`))
			}))
			defer srv.Close()
			client := animesfhd.NewClientForTest(srv.Client(), srv.URL)
			pick := func(labels []string) (int, error) {
				if tc.audio != "" {
					t.Error("single pinned version must not prompt")
				}
				if len(labels) != 2 || labels[0] != "Temporada 1 — Legendado" || labels[1] != "Temporada 2 — Dublado" {
					t.Errorf("labels=%v", labels)
				}
				if tc.cancel {
					return 0, context.Canceled
				}
				return tc.pick, nil
			}
			anime := &models.Anime{URL: srv.URL + "/serie/1", Source: "AnimesFHD"}
			got, err := getAnimesFHDEpisodes(context.Background(), anime, client, pick, tc.audio)
			if tc.cancel {
				if !errors.Is(err, context.Canceled) || len(paths) != 1 || anime.URL != srv.URL+"/serie/1" {
					t.Fatalf("cancelled selection: paths=%v anime=%+v err=%v", paths, anime, err)
				}
				return
			}
			if err != nil || len(got) != 1 || len(paths) != 2 || paths[1] != tc.wantPath {
				t.Fatalf("episodes=%v paths=%v err=%v", got, paths, err)
			}
			if strings.HasSuffix(tc.wantPath, "/2") && (anime.CurrentSeason != 2 || got[0].Audio != "Dublado" || anime.URL != srv.URL+"/serie/2") {
				t.Fatalf("selected identity/audio lost: anime=%+v episode=%+v", anime, got[0])
			}
		})
	}
}
