package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestIsAllAnimeSourceAPI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{
			name:  "explicit Source field",
			anime: &models.Anime{Source: "AllAnime", URL: "irrelevant"},
			want:  true,
		},
		{
			name:  "URL contains allanime",
			anime: &models.Anime{URL: "https://allanime.to/anime/abc"},
			want:  true,
		},
		{
			name:  "short alphanumeric ID",
			anime: &models.Anime{URL: "Bnp4XYZ"},
			want:  true,
		},
		{
			name:  "long URL not matching anything",
			anime: &models.Anime{URL: "https://goyabu.com/anime/some-very-long-slug-here"},
			want:  false,
		},
		{
			name:  "empty URL and Source",
			anime: &models.Anime{},
			want:  false,
		},
		{
			name:  "short URL with http prefix is rejected",
			anime: &models.Anime{URL: "http://x.io"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isAllAnimeSourceAPI(tt.anime)
			assert.Equal(t, tt.want, got)
		})
	}
}
