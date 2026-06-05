package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpisodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *models.Episode
		want string
	}{
		{"nil episode", nil, ""},
		{"empty episode", &models.Episode{}, ""},
		{"Number string set", &models.Episode{Number: "12"}, "12"},
		{"Num int set", &models.Episode{Num: 7}, "7"},
		{"Number wins over Num", &models.Episode{Number: "abc", Num: 5}, "abc"},
		{"Num zero falls through", &models.Episode{Num: 0}, ""},
		{"Num negative ignored", &models.Episode{Num: -1}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EpisodeNumber(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllAnimeProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.AllAnime)
	require.NoError(t, err)
	assert.Equal(t, source.AllAnime, p.Kind())
	assert.False(t, p.HasSeasons())
}

func TestAnimeFireProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.AnimeFire)
	require.NoError(t, err)
	assert.Equal(t, source.AnimeFire, p.Kind())
	assert.False(t, p.HasSeasons())
}

func TestGoyabuProvider_KindAndHasSeasons(t *testing.T) {
	t.Parallel()
	p, err := ForKind(source.Goyabu)
	require.NoError(t, err)
	assert.Equal(t, source.Goyabu, p.Kind())
	assert.False(t, p.HasSeasons())
}
