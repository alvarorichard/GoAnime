package api

import (
	"encoding/json"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/tui"
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

func TestSeasonPickItems(t *testing.T) {
	t.Parallel()
	all := map[string][]superflix.SuperFlixEpisode{
		"0": {sfEp(t, "1", "Special", "2020-01-01")},
		"1": {sfEp(t, "1", "Pilot", "2019-04-01"), sfEp(t, "2", "Second", "2019-04-08")},
		"2": {sfEp(t, "1", "Return", "2020-05-01")},
	}
	got := seasonPickItems([]string{"0", "1", "2"}, all)
	require.Len(t, got, 3)
	assert.Equal(t, "Specials", got[0].Label)
	assert.Contains(t, got[0].Details, "1 episode")
	assert.Contains(t, got[0].Details, "2020")
	assert.Equal(t, "Season 1", got[1].Label)
	assert.Contains(t, got[1].Details, "2 episodes")
	assert.Equal(t, "Season 2", got[2].Label)
	assert.Empty(t, got[1].Preview, "no side panel — preview left empty")
	// Ascending order is preserved — no reverse feed needed with Bubble Tea list.
	assert.Equal(t, []string{"Specials", "Season 1", "Season 2"}, []string{got[0].Label, got[1].Label, got[2].Label})
}

func TestCurrentSeasonIndex(t *testing.T) {
	t.Parallel()
	seasons := []string{"0", "1", "2"}
	tests := []struct {
		name  string
		media *models.Anime
		want  int
	}{
		{"nil media", nil, 0},
		{"unset current", &models.Anime{CurrentSeason: 0}, 0},
		{"negative current", &models.Anime{CurrentSeason: -1}, 0},
		{"season 1", &models.Anime{CurrentSeason: 1}, 1},
		{"season 2", &models.Anime{CurrentSeason: 2}, 2},
		{"missing season falls back", &models.Anime{CurrentSeason: 9}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, currentSeasonIndex(tt.media, seasons))
		})
	}
}

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

func TestSelectSuperFlixSeasonWith(t *testing.T) {
	t.Parallel()
	media := &models.Anime{Name: "Multi\x1b[31mSeason", CurrentSeason: 2}
	all := map[string][]superflix.SuperFlixEpisode{
		"1": {sfEp(t, "1", "Pilot", "2019-04-01")},
		"2": {sfEp(t, "1", "Return", "2020-05-01")},
		"3": {sfEp(t, "1", "Finale", "2021-06-01")},
	}
	seasons := []string{"1", "2", "3"}

	t.Run("selects preselected current season", func(t *testing.T) {
		t.Parallel()
		got, err := selectSuperFlixSeasonWith(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
			require.Len(t, items, 3)
			assert.Equal(t, 1, opts.InitialIndex, "current season 2 should preselect index 1")
			assert.Equal(t, "season", opts.ItemSingular)
			assert.Equal(t, "seasons", opts.ItemPlural)
			assert.Contains(t, opts.Breadcrumb, "Seasons")
			assert.NotContains(t, opts.Breadcrumb, "\x1b")
			assert.Equal(t, "Season 1", items[0].Label)
			assert.Equal(t, "Season 2", items[1].Label)
			return opts.InitialIndex, nil
		}, media, seasons, all)
		require.NoError(t, err)
		assert.Equal(t, "2", got)
	})

	t.Run("back propagates unwrapped", func(t *testing.T) {
		t.Parallel()
		_, err := selectSuperFlixSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickBack
		}, media, seasons, all)
		assert.ErrorIs(t, err, tui.ErrPickBack)
	})

	t.Run("cancel propagates unwrapped", func(t *testing.T) {
		t.Parallel()
		_, err := selectSuperFlixSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickCancelled
		}, media, seasons, all)
		assert.ErrorIs(t, err, tui.ErrPickCancelled)
	})

	t.Run("nil pick is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSuperFlixSeasonWith(nil, media, seasons, all)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("invalid index is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSuperFlixSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return 99, nil
		}, media, seasons, all)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid index")
	})

	t.Run("empty name becomes Title breadcrumb", func(t *testing.T) {
		t.Parallel()
		got, err := selectSuperFlixSeasonWith(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
			assert.Contains(t, opts.Breadcrumb, "Search > Title > Seasons")
			return 0, nil
		}, &models.Anime{Name: ""}, seasons, all)
		require.NoError(t, err)
		assert.Equal(t, "1", got)
	})

	t.Run("nil media does not panic", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			got, err := selectSuperFlixSeasonWith(func(_ []tui.PickItem, opts tui.PickOptions) (int, error) {
				assert.Contains(t, opts.Breadcrumb, "Title")
				return 2, nil
			}, nil, seasons, all)
			require.NoError(t, err)
			assert.Equal(t, "3", got)
		})
	})

	t.Run("empty season list is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSuperFlixSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			t.Fatal("picker must not open")
			return 0, nil
		}, media, nil, all)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no seasons")
	})
}
