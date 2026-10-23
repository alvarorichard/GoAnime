package movie

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichMedia_AnimeIsSkipped(t *testing.T) {
	t.Parallel()
	m := &models.Media{MediaType: models.MediaTypeAnime, Name: "Foo"}
	err := EnrichMedia(m)
	require.NoError(t, err)
	assert.Empty(t, m.IMDBID)
	assert.Zero(t, m.Rating)
}

func TestEnrichWithOMDb_AnimeIsSkipped(t *testing.T) {
	t.Parallel()
	m := &models.Media{MediaType: models.MediaTypeAnime, Name: "Foo"}
	err := EnrichWithOMDb(m)
	require.NoError(t, err)
}

func TestFormatMediaInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		media *models.Media
		want  string
	}{
		{
			name:  "empty media",
			media: &models.Media{},
			want:  "",
		},
		{
			name:  "year only",
			media: &models.Media{Year: "2010"},
			want:  "2010",
		},
		{
			name:  "rating only",
			media: &models.Media{Rating: 8.7},
			want:  "8.7/10",
		},
		{
			name:  "runtime hours+mins",
			media: &models.Media{Runtime: 148},
			want:  "2h 28m",
		},
		{
			name:  "runtime mins only",
			media: &models.Media{Runtime: 45},
			want:  "45m",
		},
		{
			name:  "genres truncated to 3",
			media: &models.Media{Genres: []string{"A", "B", "C", "D", "E"}},
			want:  "A, B, C",
		},
		{
			name:  "genres exact 2",
			media: &models.Media{Genres: []string{"A", "B"}},
			want:  "A, B",
		},
		{
			name: "all fields combined",
			media: &models.Media{
				Year:    "2010",
				Rating:  8.7,
				Runtime: 148,
				Genres:  []string{"Action", "Sci-Fi"},
			},
			want: "2010 | 8.7/10 | 2h 28m | Action, Sci-Fi",
		},
		{
			name:  "zero rating omitted",
			media: &models.Media{Year: "2020", Rating: 0.0},
			want:  "2020",
		},
		{
			name:  "zero runtime omitted",
			media: &models.Media{Year: "2020", Runtime: 0},
			want:  "2020",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatMediaInfo(tt.media)
			assert.Equal(t, tt.want, got)
		})
	}
}
