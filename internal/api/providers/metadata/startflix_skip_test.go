package metadata

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

// countingClient counts the requests that reach the mocked APIs.
type countingClient struct {
	next  HTTPClient
	calls atomic.Int32
}

func (c *countingClient) Do(req *http.Request) (*http.Response, error) {
	c.calls.Add(1)
	return c.next.Do(req)
}

// TestEnrichAnime_StartFlixNeverQueriesAniList pins that StartFlix titles are
// kept away from AniList on the playback and download paths too, as they
// already are at selection (TestEnrichAnimeData_StartFlixNeverQueriesAniList).
// StartFlix numbers episodes per season, as TMDB groups them; an AniList
// season map for the same name re-reads those numbers as absolute ones. With
// "Attack on Titan" — 25 episodes then a 12-episode sequel on AniList —
// StartFlix's season 4 episode 20 fell outside every mapped season and was
// filed as S01E20, on top of the real one. The AniList ids it would add to
// the folder may belong to another title altogether.
func TestEnrichAnime_StartFlixNeverQueriesAniList(t *testing.T) {
	t.Parallel()
	media := makeMedia(16498, 16498, "Attack on Titan", "Shingeki no Kyojin", 2013, 25)
	_ = json.Unmarshal([]byte(`{"edges":[{"relationType":"SEQUEL","node":{"id":20958,
		"title":{"romaji":"Shingeki no Kyojin Season 2","english":"Attack on Titan Season 2"},
		"episodes":12,"format":"TV","startDate":{"year":2017,"month":4}}}]}`), &media.Relations)
	mock := newMockClient()
	mock.addAniListResponse(media)
	client := &countingClient{next: mock}

	anime := &models.Anime{
		Name: "Attack on Titan", Source: "StartFlix", MediaType: models.MediaTypeTV,
		CurrentSeason: 4, TMDBID: 1429,
	}
	seasonMap, err := NewEnricherWithClient(client).EnrichAnime(context.Background(), anime)
	if err != nil {
		t.Fatal(err)
	}
	if n := client.calls.Load(); n != 0 {
		t.Errorf("AniList was queried %d times for a StartFlix title", n)
	}
	if len(seasonMap) != 0 {
		t.Errorf("season map = %+v; StartFlix numbers are per season already", seasonMap)
	}
	if anime.AnilistID != 0 || anime.MalID != 0 {
		t.Errorf("AniList ids %d/%d were attached to a StartFlix title", anime.AnilistID, anime.MalID)
	}
	if anime.CurrentSeason != 4 {
		t.Errorf("the selected season changed to %d", anime.CurrentSeason)
	}
}
