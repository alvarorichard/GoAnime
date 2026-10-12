package animefire

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The API numbers episodes within their season: One Piece lists 1180 episodes
// over 23 seasons and 983 of those numbers repeat (2026-10-12). Everything
// after the picker finds an episode by its number and takes the first match —
// next episode, resume, ranged downloads — so asking for season 2 episode 1
// played season 1 episode 1.
//
// data.seasons gives each season's first_episode_number. For all 24 titles
// sampled, first_episode_number + number - 1 numbered the list 1..N with no
// repeat, no gap and in list order.

func TestEpisodesAPI_NumbersRunAcrossSeasons(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/anime/op": writeJSON(map[string]any{"data": map[string]any{
			"seasons": []map[string]any{
				{"title": nil, "number": 1, "first_episode_number": 1},
				{"title": nil, "number": 2, "first_episode_number": 62},
			},
			"episodes": []map[string]any{
				{"id": "a", "season": 1, "number": 1, "title": "Eu Sou Luffy!"},
				{"id": "b", "season": 1, "number": 61, "title": "Fim da saga"},
				{"id": "c", "season": 2, "number": 1, "title": "Surge Laboon"},
				{"id": "d", "season": 2, "number": 2, "title": "Seguinte"},
			},
		}}),
	})

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/op")

	require.NoError(t, err)
	require.Len(t, eps, 4)
	var numbers []string
	var nums []int
	for _, e := range eps {
		numbers = append(numbers, e.Number)
		nums = append(nums, e.Num)
	}
	assert.Equal(t, []string{"1", "61", "62", "63"}, numbers)
	assert.Equal(t, []int{1, 61, 62, 63}, nums)
	assert.Equal(t, "2", eps[2].SeasonID, "the season is still recorded")
	assert.Equal(t, "Surge Laboon", eps[2].Title.English)
}

// Without season entries to place them, numbers that repeat are replaced by
// the position in the list, which the API gives in order.
func TestEpisodesAPI_RepeatedNumbersWithoutSeasonEntriesFollowTheList(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/anime/x": writeJSON(map[string]any{"data": map[string]any{
			"episodes": []map[string]any{
				{"id": "a", "season": 1, "number": 1},
				{"id": "b", "season": 1, "number": 2},
				{"id": "c", "season": 2, "number": 1},
			},
		}}),
	})

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/x")

	require.NoError(t, err)
	require.Len(t, eps, 3)
	assert.Equal(t, []int{1, 2, 3}, []int{eps[0].Num, eps[1].Num, eps[2].Num})
	assert.Equal(t, "3", eps[2].Number)
}

// A movie has no season and keeps its number.
func TestEpisodesAPI_MovieKeepsItsNumber(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/anime/m": writeJSON(map[string]any{"data": map[string]any{
			"format":   "movie",
			"seasons":  []any{},
			"episodes": []map[string]any{{"id": "a", "season": nil, "number": 1}},
		}}),
	})

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/m")

	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "1", eps[0].Number)
	assert.Equal(t, 1, eps[0].Num)
}
