package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTagResults_AniDBEnglish(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "Naruto", URL: "id1"}}
	tagResults(res, source.AniDB)
	assert.Equal(t, "[English] Naruto", res[0].Name)
	assert.Equal(t, "AniDB", res[0].Source)
}

func TestTagResults_AnimeFirePTBRAndSource(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "Naruto Dublado", URL: "https://animefire.plus/dublado/x"}}
	tagResults(res, source.AnimeFire)
	assert.Contains(t, res[0].Name, "[PT-BR]")
	assert.Contains(t, res[0].Name, "(Dublado)")
	assert.Equal(t, "Animefire.io", res[0].Source, "AnimeFire stamps the canonical Animefire.io source")
}

func TestTagResults_SuperFlixMediaTypeTags(t *testing.T) {
	t.Parallel()
	movie := []*models.Anime{{Name: "Inception", MediaType: models.MediaTypeMovie}}
	tagResults(movie, source.SuperFlix)
	assert.Equal(t, "[Movie] [PT-BR] Inception", movie[0].Name)

	tv := []*models.Anime{{Name: "Breaking Bad", MediaType: models.MediaTypeTV}}
	tagResults(tv, source.SuperFlix)
	assert.Equal(t, "[TV] [PT-BR] Breaking Bad", tv[0].Name)
}

func TestTagResults_DoesNotDoubleTag(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "[English] Naruto"}}
	tagResults(res, source.AniDB)
	assert.Equal(t, "[English] Naruto", res[0].Name, "an already-tagged name must not be tagged twice")
}

func TestCleanPTBRTitle(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"Naruto (Dublado)", "Naruto"},
		{"One Piece  A14", "One Piece"},
		{"Bleach 8.39", "Bleach"},
		{"Death Note (TV)", "Death Note"},
		{"  Spaced   Out  ", "Spaced Out"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, cleanPTBRTitle(tt.in))
	}
}
