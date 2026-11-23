package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTagResults_HiAnimeEnglish(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "Naruto", URL: "id1"}}
	tagResults(res, source.HiAnime)
	assert.Equal(t, "[English] Naruto", res[0].Name)
	assert.Equal(t, "HiAnime", res[0].Source)
}

func TestTagResults_AnimeFirePTBRAndSource(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "Naruto Dublado", URL: "https://animefire.plus/dublado/x"}}
	tagResults(res, source.AnimeFire)
	assert.Contains(t, res[0].Name, "[PT-BR]")
	assert.Contains(t, res[0].Name, "(Dublado)")
	assert.Equal(t, "Animefire.io", res[0].Source, "AnimeFire stamps the canonical Animefire.io source")
}

func TestTagResults_StartFlixMediaTypeTags(t *testing.T) {
	t.Parallel()
	movie := []*models.Anime{{Name: "Inception", MediaType: models.MediaTypeMovie}}
	tagResults(movie, source.StartFlix)
	assert.Equal(t, "[Movie] [PT-BR] Inception", movie[0].Name)

	tv := []*models.Anime{{Name: "Breaking Bad", MediaType: models.MediaTypeTV}}
	tagResults(tv, source.StartFlix)
	assert.Equal(t, "[TV] [PT-BR] Breaking Bad", tv[0].Name)
}

func TestTagResults_TopCineMediaTypeTags(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{
		{Name: "Zona Zero", MediaType: models.MediaTypeMovie},
		{Name: "Breaking Bad", MediaType: models.MediaTypeTV},
	}
	tagResults(res, source.TopCine)
	assert.Equal(t, "[Movie] [PT-BR] Zona Zero", res[0].Name)
	assert.Equal(t, "[TV] [PT-BR] Breaking Bad", res[1].Name)
	assert.Equal(t, "TopCine", res[0].Source)
}

func TestTagResults_AnimesDigitalAudioVariants(t *testing.T) {
	t.Parallel()
	paired := []*models.Anime{{Name: "One Piece", URL: "https://animesdigital.org/anime/a/one-piece#goanime-dub=b3BhcXVl"}}
	tagResults(paired, source.AnimesDigital)
	assert.Equal(t, "[PT-BR] One Piece", paired[0].Name, "a merged title defaults to subtitles and must not look dubbed")
	assert.Equal(t, "AnimesDigital", paired[0].Source)

	dubbedOnly := []*models.Anime{{Name: "One Piece Dublado", URL: "https://animesdigital.org/anime/a/one-piece#goanime-dublado-only=1"}}
	tagResults(dubbedOnly, source.AnimesDigital)
	assert.Equal(t, "[PT-BR] One Piece (Dublado)", dubbedOnly[0].Name)
}

func TestTagResults_DoesNotDoubleTag(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{{Name: "[English] Naruto"}}
	tagResults(res, source.HiAnime)
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

// TopCine lists iCarly (2007), iCarly (2021) and the 2021 reunion movie under
// near-identical names; without the year the user could not tell the working
// 2021 series from the movie that has no playable record.
func TestTagResults_TopCineKeepsSameTitleEntriesApart(t *testing.T) {
	t.Parallel()
	res := []*models.Anime{
		{Name: "iCarly: A Reunião", Year: "2021", MediaType: models.MediaTypeMovie, URL: "/filme/icarly-a-reuniao"},
		{Name: "iCarly", Year: "2007", MediaType: models.MediaTypeTV, URL: "/serie/icarly-2"},
		{Name: "iCarly", Year: "2021", MediaType: models.MediaTypeTV, URL: "/serie/icarly"},
		{Name: "Sem Ano", MediaType: models.MediaTypeTV, URL: "/serie/sem-ano"},
	}
	tagResults(res, source.TopCine)
	assert.Equal(t, "[Movie] [PT-BR] iCarly: A Reunião (2021)", res[0].Name)
	assert.Equal(t, "[TV] [PT-BR] iCarly (2007)", res[1].Name)
	assert.Equal(t, "[TV] [PT-BR] iCarly (2021)", res[2].Name)
	assert.NotEqual(t, res[1].Name, res[2].Name)
	assert.Equal(t, "[TV] [PT-BR] Sem Ano", res[3].Name)

	// Tagging twice must not stack the year.
	tagResults(res, source.TopCine)
	assert.Equal(t, "[TV] [PT-BR] iCarly (2021)", res[2].Name)
}
