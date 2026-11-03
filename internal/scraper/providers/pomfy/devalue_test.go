package pomfy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// showPayloadJSON is the pomfy.my/serie/__data.json shape (record 2026-11-02,
// Breaking Bad) with the unrelated fields stripped.
func showPayloadJSON() []byte {
	// indexes:  0 {"tvShow":1}
	//           1 tvShow key map
	//           2 false            (adult)
	//           3 "/poster.jpg"    (poster_path)
	//           4 [5]              (genres)
	//           5 {"id":6,"name":7} (genre)
	//           6 10759
	//           7 "Action"
	//           8 "2008-01-20"    (first_air_date)
	//           9 "Breaking Bad"  (name)
	//          10 "A chemistry teacher drug dealer drama." (overview)
	//          11 {"season_number":12,"episode_count":13,"name":14}
	//          12 1
	//          13 7
	//          14 "Temporada 1"
	//          15 {"season_number":16,"episode_count":17,"name":18}
	//          16 2
	//          17 13
	//          18 "Temporada 2"
	//          19 [11,15]        (seasons)
	node := []any{
		map[string]any{"tvShow": 1},
		map[string]any{
			"adult": 2, "poster_path": 3, "genres": 4, "first_air_date": 8,
			"name": 9, "overview": 10, "seasons": 19,
		},
		false,
		"/poster.jpg",
		[]any{5},
		map[string]any{"id": 6, "name": 7},
		10759,
		"Action",
		"2008-01-20",
		"Breaking Bad",
		"A chemistry teacher drug dealer drama.",
		map[string]any{"season_number": 12, "episode_count": 13, "name": 14},
		1,
		7,
		"Temporada 1",
		map[string]any{"season_number": 16, "episode_count": 17, "name": 18},
		2,
		13,
		"Temporada 2",
		[]any{11, 15},
	}
	raw, _ := json.Marshal(map[string]any{
		"nodes": []map[string]any{
			{"data": []any{[]any{}, nil}},
			{"data": node},
		},
	})
	return raw
}

func TestParseShowInfoSeasons(t *testing.T) {
	info, err := parseShowInfo(showPayloadJSON())
	if err != nil {
		t.Fatalf("parseShowInfo: %v", err)
	}
	if info.Name != "Breaking Bad" || info.Year != "2008" {
		t.Errorf("name/year = %q/%q", info.Name, info.Year)
	}
	if info.PosterPath != "/poster.jpg" {
		t.Errorf("poster = %q", info.PosterPath)
	}
	if len(info.Seasons) != 2 {
		t.Fatalf("seasons = %+v, want 2", info.Seasons)
	}
	if info.Seasons[0].Number != 1 || info.Seasons[0].EpisodeCount != 7 || info.Seasons[0].Name != "Temporada 1" {
		t.Errorf("season 1 = %+v", info.Seasons[0])
	}
	if info.Seasons[1].Number != 2 || info.Seasons[1].EpisodeCount != 13 {
		t.Errorf("season 2 = %+v", info.Seasons[1])
	}
}

func TestParseShowInfoGarbagePayload(t *testing.T) {
	for _, body := range []string{`{}`, `{"nodes":[]}`, `not json`} {
		if _, err := parseShowInfo([]byte(body)); err == nil {
			t.Errorf("garbage payload %q must fail", body)
		}
	}
}

// TestSeasonListingAvailability joins the episode numbers with the available
// ones and names from the tmdbEps map, exactly what /api/tmdb-season says.
func TestSeasonListingAvailability(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/tmdb-season") {
			if got := r.URL.Query().Get("id"); got != "1396" {
				t.Errorf("id = %q, want 1396", got)
			}
			if got := r.URL.Query().Get("s"); got != "2" {
				t.Errorf("s = %q, want 2", got)
			}
			fmt.Fprint(w, `{"epNums":[1,2,3],"sbEps":[1,3],"tmdbEps":{"1":{"name":"Seven Thirty-Seven"},"3":{"name":"Bit by a Dead Bee"}}}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)

	season, err := c.Season(t.Context(), 1396, 2)
	if err != nil {
		t.Fatalf("Season: %v", err)
	}
	if len(season.EpisodeNums) != 3 {
		t.Errorf("episodeNums = %v", season.EpisodeNums)
	}
	if len(season.AvailableNums) != 2 || season.AvailableNums[0] != 1 || season.AvailableNums[1] != 3 {
		t.Errorf("available = %v, want [1 3]", season.AvailableNums)
	}
	if len(season.Episodes) != 2 || season.Episodes[0].Title != "Seven Thirty-Seven" {
		t.Fatalf("episodes = %+v", season.Episodes)
	}
}

// TestDevalueGraphResolve checks the reference walk directly on the payload
// shape the show pages carry (objects pointing into the array by index).
func TestDevalueGraphResolve(t *testing.T) {
	payload := showPayloadJSON()

	var parsed struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatal(err)
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(parsed.Nodes[1].Data, &arr); err != nil {
		t.Fatal(err)
	}
	g := newDevalueGraph(arr)
	head, ok := g.object(0)
	if !ok || head["tvShow"] != 1 {
		t.Fatalf("head = %+v", head)
	}
	show, ok := g.object(1)
	if !ok {
		t.Fatal("tvShow object not resolvable at index 1")
	}
	name, _ := g.stringRef(show["name"])
	if name != "Breaking Bad" {
		t.Errorf("name = %q", name)
	}
	seasons := g.array(show["seasons"])
	if len(seasons) != 2 {
		t.Fatalf("seasons = %v", seasons)
	}
	season, ok := g.object(seasons[0])
	if !ok || g.integerRef(season["episode_count"]) != 7 {
		t.Errorf("season 1 = %+v", season)
	}
}
