package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

type mockHTTPClient struct {
	responses map[string]*http.Response
}

func newMockClient() *mockHTTPClient {
	return &mockHTTPClient{
		responses: make(map[string]*http.Response),
	}
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Try exact Method:Host+Path match first (for TMDB), then Method:Host
	fullKey := req.Method + ":" + req.URL.Host + req.URL.Path
	if resp, ok := m.responses[fullKey]; ok {
		return resp, nil
	}
	key := req.Method + ":" + req.URL.Host
	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil
}

func (m *mockHTTPClient) addAniListResponse(media aniListMedia) {
	resp := aniListResponse{}
	resp.Data.Media = media
	body, _ := json.Marshal(resp)
	m.responses["POST:graphql.anilist.co"] = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func makeMedia(id, malID int, english, romaji string, year, episodes int) aniListMedia {
	m := aniListMedia{
		ID:       id,
		IDMal:    malID,
		Episodes: episodes,
		Status:   "FINISHED",
	}
	m.Title.English = english
	m.Title.Romaji = romaji
	m.StartDate.Year = year
	return m
}

func TestEnrichFromAniList_BasicMetadata(t *testing.T) {
	mock := newMockClient()
	mock.addAniListResponse(makeMedia(20, 20, "Naruto", "NARUTO", 2002, 220))

	enricher := NewEnricherWithClient(mock)
	meta, err := enricher.EnrichFromAniList(context.Background(), "Naruto [English]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.TitleEnglish != "Naruto" {
		t.Errorf("TitleEnglish = %q, want %q", meta.TitleEnglish, "Naruto")
	}
	if meta.TitleRomaji != "NARUTO" {
		t.Errorf("TitleRomaji = %q, want %q", meta.TitleRomaji, "NARUTO")
	}
	if meta.Year != "2002" {
		t.Errorf("Year = %q, want %q", meta.Year, "2002")
	}
	if meta.TotalEpisodes != 220 {
		t.Errorf("TotalEpisodes = %d, want %d", meta.TotalEpisodes, 220)
	}
	if meta.AniListID != 20 {
		t.Errorf("AniListID = %d, want %d", meta.AniListID, 20)
	}
	if meta.MalID != 20 {
		t.Errorf("MalID = %d, want %d", meta.MalID, 20)
	}
}

func TestEnrichFromAniList_WithSequels(t *testing.T) {
	media := makeMedia(16498, 16498, "Attack on Titan", "Shingeki no Kyojin", 2013, 25)
	relJSON := `{
		"edges": [{
			"relationType": "SEQUEL",
			"node": {
				"id": 20958,
				"title": {"romaji": "Shingeki no Kyojin Season 2", "english": "Attack on Titan Season 2"},
				"episodes": 12,
				"format": "TV",
				"startDate": {"year": 2017, "month": 4}
			}
		}, {
			"relationType": "SIDE_STORY",
			"node": {
				"id": 99634,
				"episodes": 3,
				"format": "OVA"
			}
		}]
	}`
	_ = json.Unmarshal([]byte(relJSON), &media.Relations)

	mock := newMockClient()
	mock.addAniListResponse(media)

	enricher := NewEnricherWithClient(mock)
	meta, err := enricher.EnrichFromAniList(context.Background(), "Attack on Titan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(meta.SeasonMap) != 2 {
		t.Fatalf("expected 2 seasons, got %d", len(meta.SeasonMap))
	}

	s1 := meta.SeasonMap[0]
	if s1.Season != 1 || s1.StartEp != 1 || s1.EndEp != 25 {
		t.Errorf("Season 1 = %+v, want {1, 1, 25}", s1)
	}

	s2 := meta.SeasonMap[1]
	if s2.Season != 2 || s2.StartEp != 26 || s2.EndEp != 37 {
		t.Errorf("Season 2 = %+v, want {2, 26, 37}", s2)
	}
}

func TestAbsoluteToSeason(t *testing.T) {
	meta := &AnimeMetadata{
		SeasonMap: []SeasonMapping{
			{Season: 1, StartEp: 1, EndEp: 25, EpisodeCount: 25},
			{Season: 2, StartEp: 26, EndEp: 37, EpisodeCount: 12},
			{Season: 3, StartEp: 38, EndEp: 59, EpisodeCount: 22},
		},
	}

	tests := []struct {
		absEp, wantS, wantE int
	}{
		{1, 1, 1},
		{25, 1, 25},
		{26, 2, 1},
		{37, 2, 12},
		{38, 3, 1},
		{59, 3, 22},
		{60, 3, 23},
	}
	for _, tt := range tests {
		s, e := meta.AbsoluteToSeason(tt.absEp)
		if s != tt.wantS || e != tt.wantE {
			t.Errorf("AbsoluteToSeason(%d) = (%d, %d), want (%d, %d)", tt.absEp, s, e, tt.wantS, tt.wantE)
		}
	}
}

func TestAbsoluteToSeason_NoMap(t *testing.T) {
	meta := &AnimeMetadata{}
	s, e := meta.AbsoluteToSeason(42)
	if s != 1 || e != 42 {
		t.Errorf("AbsoluteToSeason(42) with no map = (%d, %d), want (1, 42)", s, e)
	}
}

func TestApplyToAnime(t *testing.T) {
	mock := newMockClient()
	mock.addAniListResponse(makeMedia(20, 20, "Naruto", "NARUTO", 2002, 220))

	enricher := NewEnricherWithClient(mock)
	anime := &models.Anime{Name: "Naruto [English]"}

	err := enricher.ApplyToAnime(context.Background(), anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if anime.Details.Title.English != "Naruto" {
		t.Errorf("Title.English = %q, want %q", anime.Details.Title.English, "Naruto")
	}
	if anime.Year != "2002" {
		t.Errorf("Year = %q, want %q", anime.Year, "2002")
	}
	if anime.AnilistID != 20 {
		t.Errorf("AnilistID = %d, want %d", anime.AnilistID, 20)
	}
	if anime.MalID != 20 {
		t.Errorf("MalID = %d, want %d", anime.MalID, 20)
	}
}

func TestApplyToAnime_NilAnime(t *testing.T) {
	enricher := NewEnricher()
	err := enricher.ApplyToAnime(context.Background(), nil)
	if err != nil {
		t.Errorf("ApplyToAnime(nil) should not error, got %v", err)
	}
}

func TestApplyToAnime_DoesNotOverwrite(t *testing.T) {
	mock := newMockClient()
	mock.addAniListResponse(makeMedia(20, 20, "AniList Title", "", 2002, 0))

	enricher := NewEnricherWithClient(mock)
	anime := &models.Anime{
		Name:      "Naruto",
		Year:      "2001",
		AnilistID: 99,
		Details: models.AniListDetails{
			Title: models.Title{English: "Existing Title"},
		},
	}

	_ = enricher.ApplyToAnime(context.Background(), anime)

	if anime.Year != "2001" {
		t.Errorf("Year was overwritten: got %q, want %q", anime.Year, "2001")
	}
	if anime.AnilistID != 99 {
		t.Errorf("AnilistID was overwritten: got %d, want %d", anime.AnilistID, 99)
	}
	if anime.Details.Title.English != "Existing Title" {
		t.Errorf("Title was overwritten: got %q, want %q", anime.Details.Title.English, "Existing Title")
	}
}

func TestCleanSearchName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Naruto [English]", "Naruto"},
		{"Naruto [PT-BR] [AnimeFire]", "Naruto"},
		{"Naruto", "Naruto"},
		{"Attack on Titan [English] [AllAnime]", "Attack on Titan"},
		{"[Only Tags]", ""},
		{"[PT-BR] Black Clover (Dublado)", "Black Clover"},
		{"Black Clover (Legendado)", "Black Clover"},
		{"Naruto (Dub)", "Naruto"},
		{"One Piece (Dual Audio)", "One Piece"},
		{"Bleach (Completo)", "Bleach"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := cleanSearchName(tt.input); got != tt.want {
				t.Errorf("cleanSearchName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnrichAnime_TMDBFallbackForSingleSeason(t *testing.T) {
	// Black Clover: single AniList entry (170 eps, no sequels)
	// but TMDB splits it into 4 seasons.
	mock := newMockClient()
	mock.addAniListResponse(makeMedia(97986, 34572, "Black Clover", "Black Clover", 2017, 170))

	// Mock TMDB find (MAL ID → TMDB TV ID)
	findBody, _ := json.Marshal(map[string]any{
		"tv_results": []map[string]any{{"id": 73223}},
	})
	mock.responses["GET:api.themoviedb.org/3/find/mal-34572"] = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(findBody)),
	}

	// Mock TMDB TV details (4 seasons)
	tvBody, _ := json.Marshal(map[string]any{
		"seasons": []map[string]any{
			{"season_number": 0, "episode_count": 4}, // Specials — skipped
			{"season_number": 1, "episode_count": 51},
			{"season_number": 2, "episode_count": 51},
			{"season_number": 3, "episode_count": 52},
			{"season_number": 4, "episode_count": 16},
		},
	})
	mock.responses["GET:api.themoviedb.org/3/tv/73223"] = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(tvBody)),
	}

	// Set TMDB_API_KEY for the test
	t.Setenv("TMDB_API_KEY", "test-key")

	enricher := NewEnricherWithClient(mock)
	anime := &models.Anime{Name: "Black Clover (Dublado)"}
	seasonMap, err := enricher.EnrichAnime(context.Background(), anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(seasonMap) != 4 {
		t.Fatalf("expected 4 seasons, got %d: %+v", len(seasonMap), seasonMap)
	}

	// Season 1: eps 1-51
	if seasonMap[0].Season != 1 || seasonMap[0].StartEp != 1 || seasonMap[0].EndEp != 51 {
		t.Errorf("season 1: got %+v", seasonMap[0])
	}
	// Season 2: eps 52-102
	if seasonMap[1].Season != 2 || seasonMap[1].StartEp != 52 || seasonMap[1].EndEp != 102 {
		t.Errorf("season 2: got %+v", seasonMap[1])
	}
	// Season 3: eps 103-154
	if seasonMap[2].Season != 3 || seasonMap[2].StartEp != 103 || seasonMap[2].EndEp != 154 {
		t.Errorf("season 3: got %+v", seasonMap[2])
	}
	// Season 4: eps 155-170
	if seasonMap[3].Season != 4 || seasonMap[3].StartEp != 155 || seasonMap[3].EndEp != 170 {
		t.Errorf("season 4: got %+v", seasonMap[3])
	}

	// Verify episode resolution
	meta := &AnimeMetadata{SeasonMap: seasonMap}
	s, e := meta.AbsoluteToSeason(27)
	if s != 1 || e != 27 {
		t.Errorf("ep 27: got S%02dE%02d, want S01E27", s, e)
	}
	s, e = meta.AbsoluteToSeason(52)
	if s != 2 || e != 1 {
		t.Errorf("ep 52: got S%02dE%02d, want S02E01", s, e)
	}
	s, e = meta.AbsoluteToSeason(103)
	if s != 3 || e != 1 {
		t.Errorf("ep 103: got S%02dE%02d, want S03E01", s, e)
	}
	s, e = meta.AbsoluteToSeason(170)
	if s != 4 || e != 16 {
		t.Errorf("ep 170: got S%02dE%02d, want S04E16", s, e)
	}
}

func TestEnrichAnime_SuperFlixFallback(t *testing.T) {
	// Black Clover: AniList has 170 eps, no sequels.
	// No TMDB_API_KEY. SuperFlix provides season data.
	t.Setenv("TMDB_API_KEY", "") // ensure TMDB path is skipped

	mock := newMockClient()
	mock.addAniListResponse(makeMedia(97986, 34572, "Black Clover", "Black Clover", 2017, 170))

	// Mock SuperFlix search page — contains a serie link with TMDB ID
	searchHTML := `<html><body>
		<div class="card">
			<h3>Black Clover</h3>
			<button data-msg="Copiar TMDB" data-copy="73223">TMDB</button>
			<button data-msg="Copiar Link" data-copy="https://superflixapi.rest/serie/73223">Link</button>
		</div>
	</body></html>`
	mock.responses["GET:superflixapi.rest/pesquisar"] = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchHTML)),
	}

	// Mock SuperFlix player page — contains ALL_EPISODES JS variable
	allEpisodes := `var ALL_EPISODES = {"1":[` +
		strings.Repeat(`{"epi_num":"1","title":"ep","air_date":"2017-10-03"},`, 50) +
		`{"epi_num":"51","title":"ep","air_date":"2018-09-25"}` +
		`],"2":[` +
		strings.Repeat(`{"epi_num":"1","title":"ep","air_date":"2018-10-02"},`, 50) +
		`{"epi_num":"51","title":"ep","air_date":"2019-09-24"}` +
		`],"3":[` +
		strings.Repeat(`{"epi_num":"1","title":"ep","air_date":"2019-10-01"},`, 51) +
		`{"epi_num":"52","title":"ep","air_date":"2020-09-29"}` +
		`],"4":[` +
		strings.Repeat(`{"epi_num":"1","title":"ep","air_date":"2020-12-01"},`, 15) +
		`{"epi_num":"16","title":"ep","air_date":"2021-03-30"}` +
		`]};`
	playerHTML := `<html><body><script>` + allEpisodes + `</script></body></html>`
	mock.responses["GET:superflixapi.rest/serie/73223"] = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(playerHTML)),
	}

	enricher := NewEnricherWithClient(mock)
	anime := &models.Anime{Name: "[PT-BR] Black Clover (Dublado)"}
	seasonMap, err := enricher.EnrichAnime(context.Background(), anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(seasonMap) != 4 {
		t.Fatalf("expected 4 seasons from SuperFlix, got %d: %+v", len(seasonMap), seasonMap)
	}

	// Season 1: 51 eps (1-51)
	if seasonMap[0].Season != 1 || seasonMap[0].EpisodeCount != 51 {
		t.Errorf("season 1: got %+v", seasonMap[0])
	}
	// Season 2: 51 eps (52-102)
	if seasonMap[1].Season != 2 || seasonMap[1].StartEp != 52 || seasonMap[1].EpisodeCount != 51 {
		t.Errorf("season 2: got %+v", seasonMap[1])
	}
	// Season 4: 16 eps (155-170)
	if seasonMap[3].Season != 4 || seasonMap[3].EpisodeCount != 16 {
		t.Errorf("season 4: got %+v", seasonMap[3])
	}

	// Episode 55 should map to Season 2 Episode 4
	meta := &AnimeMetadata{SeasonMap: seasonMap}
	s, e := meta.AbsoluteToSeason(55)
	if s != 2 || e != 4 {
		t.Errorf("ep 55: got S%02dE%02d, want S02E04", s, e)
	}
}

func TestInferSeasonNumberFromTitle(t *testing.T) {
	tests := []struct {
		title string
		want  int
	}{
		{title: "JUJUTSU KAISEN Season 2", want: 2},
		{title: "Jujutsu Kaisen 2 Season", want: 2},
		{title: "Jujutsu Kaisen 2nd Season", want: 2},
		{title: "Black Clover", want: 0},
		{title: "Season 1", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			if got := inferSeasonNumberFromTitle(tc.title); got != tc.want {
				t.Fatalf("inferSeasonNumberFromTitle(%q) = %d, want %d", tc.title, got, tc.want)
			}
		})
	}
}

func TestEnrichAnime_JujutsuKaisenSeason2InfersCurrentSeasonFromMockedAniList(t *testing.T) {
	media := makeMedia(145064, 51009, "JUJUTSU KAISEN Season 2", "Jujutsu Kaisen 2nd Season", 2023, 23)
	relJSON := `{
		"edges": [{
			"relationType": "SEQUEL",
			"node": {
				"id": 999999,
				"title": {"romaji": "Jujutsu Kaisen Culling Game", "english": "JUJUTSU KAISEN Season 3"},
				"episodes": 23,
				"format": "TV",
				"startDate": {"year": 2026, "month": 1}
			}
		}]
	}`
	if err := json.Unmarshal([]byte(relJSON), &media.Relations); err != nil {
		t.Fatalf("failed to build mock relations: %v", err)
	}

	mock := newMockClient()
	mock.addAniListResponse(media)
	enricher := NewEnricherWithClient(mock)
	anime := &models.Anime{
		Name:      "[PT-BR] Jujutsu Kaisen 2 Season (Dublado)",
		URL:       "https://goyabu.io/anime/jujutsu-kaisen-2-season-dublado",
		Source:    "Goyabu",
		AnilistID: 145064,
		MalID:     51009,
	}

	seasonMap, err := enricher.EnrichAnime(context.Background(), anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if anime.CurrentSeason != 2 {
		t.Fatalf("CurrentSeason = %d, want 2", anime.CurrentSeason)
	}
	if anime.OfficialTitle() != "JUJUTSU KAISEN Season 2" {
		t.Fatalf("OfficialTitle = %q, want JUJUTSU KAISEN Season 2", anime.OfficialTitle())
	}
	if len(seasonMap) != 2 {
		t.Fatalf("seasonMap len = %d, want 2: %+v", len(seasonMap), seasonMap)
	}
	if seasonMap[0].Season != 1 || seasonMap[0].StartEp != 1 || seasonMap[0].EndEp != 23 {
		t.Fatalf("seasonMap[0] = %+v, want season 1 range 1-23", seasonMap[0])
	}
	if seasonMap[1].Season != 2 || seasonMap[1].StartEp != 24 || seasonMap[1].EndEp != 46 {
		t.Fatalf("seasonMap[1] = %+v, want season 2 range 24-46", seasonMap[1])
	}
}
