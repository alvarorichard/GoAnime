package api

import (
	"encoding/json"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sfEp(t *testing.T, num, title, airDate string) superflix.SuperFlixEpisode {
	t.Helper()
	return superflix.SuperFlixEpisode{
		EpiNum:  json.Number(num),
		Title:   title,
		AirDate: airDate,
	}
}

func TestSeasonDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"regular season", "2", "Season 2"},
		{"season zero is specials", "0", "Specials"},
		{"double digit", "10", "Season 10"},
		{"non-numeric key passes through", "2019", "Season 2019"},
		{"non-numeric alpha", "OVA", "Season OVA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, seasonDisplayName(tt.in))
		})
	}
}

func TestEpisodeCountLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   int
		want string
	}{
		{"singular", 1, "1 episode"},
		{"plural", 12, "12 episodes"},
		{"zero is plural", 0, "0 episodes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, episodeCountLabel(tt.in))
		})
	}
}

func TestAirYear(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"iso date", "2019-04-01", "2019"},
		{"year only", "2021", "2021"},
		{"empty", "", ""},
		{"too short", "20", ""},
		{"non-numeric prefix", "abcd-01-01", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, airYear(tt.in))
		})
	}
}

func TestSeasonYearRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		eps  []superflix.SuperFlixEpisode
		want string
	}{
		{
			"single year",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", "2019-04-01"), sfEp(t, "2", "b", "2019-06-20")},
			"2019",
		},
		{
			"year span",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", "2019-10-01"), sfEp(t, "2", "b", "2020-01-05")},
			"2019-2020",
		},
		{
			"unordered dates still span correctly",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", "2021-01-01"), sfEp(t, "2", "b", "2019-01-01")},
			"2019-2021",
		},
		{
			"missing dates ignored",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", ""), sfEp(t, "2", "b", "2020-01-05")},
			"2020",
		},
		{
			"all dates missing",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", ""), sfEp(t, "2", "b", "")},
			"",
		},
		{"empty slice", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, seasonYearRange(tt.eps))
		})
	}
}

func TestSeasonLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		sn   string
		eps  []superflix.SuperFlixEpisode
		want string
	}{
		{
			"with year",
			"2",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", "2019-04-01")},
			"Season 2  ·  1 episode  ·  2019",
		},
		{
			"without year",
			"1",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", ""), sfEp(t, "2", "b", "")},
			"Season 1  ·  2 episodes",
		},
		{
			"specials",
			"0",
			[]superflix.SuperFlixEpisode{sfEp(t, "1", "a", "2020-05-05")},
			"Specials  ·  1 episode  ·  2020",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, seasonLabel(tt.sn, tt.eps))
		})
	}
}

func TestFormatEpisodeLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ep   superflix.SuperFlixEpisode
		want string
	}{
		{"full", sfEp(t, "1", "Pilot", "2019-04-01"), "E01  Pilot  (2019-04-01)"},
		{"no date", sfEp(t, "2", "Second", ""), "E02  Second"},
		{"no title", sfEp(t, "3", "", "2019-04-15"), "E03  (2019-04-15)"},
		{"double digit not repadded", sfEp(t, "12", "Finale", ""), "E12  Finale"},
		{"whitespace title dropped", sfEp(t, "4", "   ", ""), "E04"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, formatEpisodeLine(tt.ep))
		})
	}
}

func TestSeasonPreviewText(t *testing.T) {
	t.Parallel()

	t.Run("header and all episodes fit", func(t *testing.T) {
		t.Parallel()
		eps := []superflix.SuperFlixEpisode{
			sfEp(t, "1", "Pilot", "2019-04-01"),
			sfEp(t, "2", "Second", "2019-04-08"),
		}
		got := seasonPreviewText("2", eps, 20)
		assert.Contains(t, got, "Season 2 (2019) — 2 episodes")
		assert.Contains(t, got, "E01  Pilot  (2019-04-01)")
		assert.Contains(t, got, "E02  Second  (2019-04-08)")
		assert.NotContains(t, got, "more")
	})

	t.Run("truncates to pane height with tail", func(t *testing.T) {
		t.Parallel()
		var eps []superflix.SuperFlixEpisode
		for _, n := range []string{"1", "2", "3", "4", "5", "6"} {
			eps = append(eps, sfEp(t, n, "Ep "+n, ""))
		}
		got := seasonPreviewText("1", eps, 6) // 6-4 = 2 episode lines
		assert.Contains(t, got, "E01")
		assert.Contains(t, got, "E02")
		assert.NotContains(t, got, "E03")
		assert.Contains(t, got, "... and 4 more")
	})

	t.Run("tiny pane still shows one episode", func(t *testing.T) {
		t.Parallel()
		eps := []superflix.SuperFlixEpisode{
			sfEp(t, "1", "Only", ""),
			sfEp(t, "2", "Cut", ""),
		}
		got := seasonPreviewText("1", eps, 1)
		assert.Contains(t, got, "E01  Only")
		assert.Contains(t, got, "... and 1 more")
	})
}

func TestSeasonDisplayOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"reverses ascending seasons", []string{"1", "2", "3"}, []string{"3", "2", "1"}},
		{"single element", []string{"1"}, []string{"1"}},
		{"empty", []string{}, []string{}},
		{"specials last visually", []string{"0", "1", "2"}, []string{"2", "1", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := seasonDisplayOrder(tt.in)
			assert.Equal(t, tt.want, got)
			// Input must not be mutated: callers still index allEpisodes by it.
			if len(tt.in) > 1 {
				assert.NotEqual(t, tt.in, got)
			}
		})
	}
}

// selectSuperFlixSeason opens a full-screen TUI for multi-season titles, which
// cannot run headlessly; the single-season fast path is pure and pinned here.
func TestSelectSuperFlixSeason_SingleSeasonSkipsPicker(t *testing.T) {
	t.Parallel()
	media := &models.Anime{Name: "One Season Show"}
	eps := map[string][]superflix.SuperFlixEpisode{
		"1": {sfEp(t, "1", "Pilot", "2019-04-01")},
	}
	got, err := selectSuperFlixSeason(media, []string{"1"}, eps)
	require.NoError(t, err)
	assert.Equal(t, "1", got,
		"single-season titles must be selected automatically without opening the picker")
}
