package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rowsFor builds picker rows the way a source would, one per season key.
func rowsFor(seasons []string) func() []tui.PickItem {
	return func() []tui.PickItem {
		items := make([]tui.PickItem, len(seasons))
		for i, sn := range seasons {
			items[i] = tui.PickItem{Label: seasonDisplayName(sn)}
		}
		return items
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

func TestSelectSeason_SingleSeasonSkipsPicker(t *testing.T) {
	t.Parallel()
	media := &models.Anime{Name: "One Season Show"}
	got, err := selectSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
		t.Fatal("picker must not open for a single season")
		return 0, nil
	}, media, []string{"1"}, rowsFor([]string{"1"}))
	require.NoError(t, err)
	assert.Equal(t, "1", got,
		"single-season titles must be selected automatically without opening the picker")
}

func TestSelectSeasonWith(t *testing.T) {
	t.Parallel()
	media := &models.Anime{Name: "Multi\x1b[31mSeason", CurrentSeason: 2}
	seasons := []string{"1", "2", "3"}
	all := rowsFor(seasons)

	t.Run("selects preselected current season", func(t *testing.T) {
		t.Parallel()
		got, err := selectSeasonWith(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
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
		_, err := selectSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickBack
		}, media, seasons, all)
		assert.ErrorIs(t, err, tui.ErrPickBack)
	})

	t.Run("cancel propagates unwrapped", func(t *testing.T) {
		t.Parallel()
		_, err := selectSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickCancelled
		}, media, seasons, all)
		assert.ErrorIs(t, err, tui.ErrPickCancelled)
	})

	t.Run("nil pick is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSeasonWith(nil, media, seasons, all)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("invalid index is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return 99, nil
		}, media, seasons, all)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid index")
	})

	t.Run("empty name becomes Title breadcrumb", func(t *testing.T) {
		t.Parallel()
		got, err := selectSeasonWith(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
			assert.Contains(t, opts.Breadcrumb, "Search > Title > Seasons")
			return 0, nil
		}, &models.Anime{Name: ""}, seasons, all)
		require.NoError(t, err)
		assert.Equal(t, "1", got)
	})

	t.Run("nil media does not panic", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			got, err := selectSeasonWith(func(_ []tui.PickItem, opts tui.PickOptions) (int, error) {
				assert.Contains(t, opts.Breadcrumb, "Title")
				return 2, nil
			}, nil, seasons, all)
			require.NoError(t, err)
			assert.Equal(t, "3", got)
		})
	})

	t.Run("empty season list is controlled", func(t *testing.T) {
		t.Parallel()
		_, err := selectSeasonWith(func([]tui.PickItem, tui.PickOptions) (int, error) {
			t.Fatal("picker must not open")
			return 0, nil
		}, media, nil, rowsFor(nil))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no seasons")
	})
}
