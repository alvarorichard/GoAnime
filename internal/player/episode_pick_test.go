package player

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpisodeDisplayTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ep   models.Episode
		want string
	}{
		{"romaji preferred", models.Episode{Title: models.TitleDetails{Romaji: "Romaji", English: "English"}}, "Romaji"},
		{"english fallback", models.Episode{Title: models.TitleDetails{English: "English"}}, "English"},
		{"empty", models.Episode{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, episodeDisplayTitle(tt.ep))
		})
	}
}

func TestEpisodeLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ep   models.Episode
		want string
	}{
		{"with title", models.Episode{Number: "12", Title: models.TitleDetails{Romaji: "Pilot"}}, "12 — Pilot"},
		{"number only", models.Episode{Number: "Episode 3"}, "Episode 3"},
		{"english title", models.Episode{Number: "1", Title: models.TitleDetails{English: "Start"}}, "1 — Start"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, episodeLabel(tt.ep))
		})
	}
}

func TestEpisodeDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ep   models.Episode
		want string
	}{
		{"aired only", models.Episode{Aired: "2019-04-01"}, "2019-04-01"},
		{"filler and recap", models.Episode{Aired: "2020-01-01", IsFiller: true, IsRecap: true}, "2020-01-01  •  Filler  •  Recap"},
		{"filler only", models.Episode{IsFiller: true}, "Filler"},
		{"empty", models.Episode{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, episodeDetails(tt.ep))
		})
	}
}

func TestEpisodePickItems(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{Number: "1", URL: "u1", Title: models.TitleDetails{English: "Start"}},
		{Number: "2", URL: "u2", IsRecap: true, Aired: "2020-02-02"},
	}
	got := episodePickItems(eps)
	require.Len(t, got, 2)
	assert.Equal(t, "1 — Start", got[0].Label)
	assert.Equal(t, "2", got[1].Label)
	assert.Equal(t, "2020-02-02  •  Recap", got[1].Details)
	assert.Empty(t, got[0].Preview, "no side panel — preview left empty")
}

func TestSelectEpisodeWithPicker(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{Number: "1", URL: "http://ep/1", Title: models.TitleDetails{Romaji: "Start"}},
		{Number: "12", URL: "http://ep/12", Title: models.TitleDetails{English: "Climax"}},
		{Number: "99", URL: "http://ep/99"},
	}

	t.Run("empty episodes", func(t *testing.T) {
		t.Parallel()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, error) {
			t.Fatal("picker must not open for empty list")
			return 0, nil
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no episodes provided")
	})

	t.Run("nil pick", func(t *testing.T) {
		t.Parallel()
		_, _, err := selectEpisodeWithPicker(nil, eps)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("selection returns url and number", func(t *testing.T) {
		t.Parallel()
		url, num, err := selectEpisodeWithPicker(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
			require.Len(t, items, 3)
			assert.Equal(t, "episode", opts.ItemSingular)
			assert.Equal(t, "episodes", opts.ItemPlural)
			assert.Equal(t, "1 — Start", items[0].Label)
			assert.Equal(t, "12 — Climax", items[1].Label)
			return 1, nil
		}, eps)
		require.NoError(t, err)
		assert.Equal(t, "http://ep/12", url)
		assert.Equal(t, "12", num)
	})

	t.Run("back maps to ErrBackRequested", func(t *testing.T) {
		t.Parallel()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickBack
		}, eps)
		assert.ErrorIs(t, err, ErrBackRequested)
	})

	t.Run("cancel maps to ErrBackRequested", func(t *testing.T) {
		t.Parallel()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, tui.ErrPickCancelled
		}, eps)
		assert.ErrorIs(t, err, ErrBackRequested)
	})

	t.Run("other errors wrap", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("terminal failed")
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return -1, boom
		}, eps)
		require.Error(t, err)
		assert.ErrorIs(t, err, boom)
		assert.NotErrorIs(t, err, ErrBackRequested)
	})

	t.Run("invalid index", func(t *testing.T) {
		t.Parallel()
		_, _, err := selectEpisodeWithPicker(func([]tui.PickItem, tui.PickOptions) (int, error) {
			return 99, nil
		}, eps)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid index")
	})
}

func TestSelectEpisodeWithFuzzyFinder_EmptyReturnsError(t *testing.T) {
	t.Parallel()
	_, _, err := SelectEpisodeWithFuzzyFinder(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no episodes provided")
}

func TestEpisodePickItems_OnePieceScale(t *testing.T) {
	t.Parallel()
	const n = 1169
	eps := make([]models.Episode, n)
	for i := range eps {
		eps[i] = models.Episode{
			Number: fmt.Sprintf("%d", i+1),
			URL:    fmt.Sprintf("http://ep/%d", i+1),
			Title:  models.TitleDetails{English: fmt.Sprintf("One Piece - Episódio %d", i+1)},
		}
	}
	items := episodePickItems(eps)
	require.Len(t, items, n)
	assert.Equal(t, "1 — One Piece - Episódio 1", items[0].Label)
	assert.Equal(t, "1169 — One Piece - Episódio 1169", items[n-1].Label)
	assert.Empty(t, items[0].Preview)
	// Filter tokens must include the bare episode number for type-to-jump.
	assert.Contains(t, items[1099].Label, "1100")
}

func TestSelectEpisodeWithPicker_OptionsContract(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{{Number: "7", URL: "u7", Title: models.TitleDetails{Romaji: "Seven"}}}
	url, num, err := selectEpisodeWithPicker(func(items []tui.PickItem, opts tui.PickOptions) (int, error) {
		assert.Equal(t, "Search > Episodes", opts.Breadcrumb)
		assert.Equal(t, "GoAnime - Episodes", opts.WindowTitle)
		assert.Equal(t, "episode", opts.ItemSingular)
		assert.Equal(t, "episodes", opts.ItemPlural)
		assert.Equal(t, "7 — Seven", items[0].Label)
		return 0, nil
	}, eps)
	require.NoError(t, err)
	assert.Equal(t, "u7", url)
	assert.Equal(t, "7", num)
}

func TestEpisodeDetails_WhitespaceAiredIgnored(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", episodeDetails(models.Episode{Aired: "   "}))
	assert.Equal(t, "Recap", episodeDetails(models.Episode{Aired: "\t", IsRecap: true}))
}

func TestEpisodeLabel_JapaneseTitleFallback(t *testing.T) {
	t.Parallel()
	// Japanese alone is not used by episodeDisplayTitle — number only.
	ep := models.Episode{Number: "3", Title: models.TitleDetails{Japanese: "第三話"}}
	assert.Equal(t, "3", episodeLabel(ep))
	assert.Equal(t, "", episodeDisplayTitle(ep))
}
