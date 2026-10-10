package handlers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
)

// Opening a season that is announced but has not aired used to end the
// session with the terminal cleared, which looked like a silent crash. The
// user is now sent back to the results with a notice saying why.
func TestEpisodesUnavailableNotice(t *testing.T) {
	t.Parallel()
	frieren := &models.Anime{Name: "[English] Sousou no Frieren 3rd Season", Source: "HiAnime"}

	tests := []struct {
		name       string
		anime      *models.Anime
		err        error
		want       string
		wantFailed bool
	}{
		{
			name:  "unaired season",
			anime: frieren,
			err:   fmt.Errorf("failed to fetch episodes: HiAnime anime 7220: %w", netx.ErrNoEpisodes),
			want:  "[English] Sousou no Frieren 3rd Season has no episodes on HiAnime yet.",
		},
		{
			name:  "empty list without an error",
			anime: frieren,
			want:  "[English] Sousou no Frieren 3rd Season has no episodes on HiAnime yet.",
		},
		{
			name:       "the source failed",
			anime:      frieren,
			err:        errors.New("HiAnime episodes: request failed"),
			want:       "Could not load the episodes of [English] Sousou no Frieren 3rd Season on HiAnime. Try another result.",
			wantFailed: true,
		},
		{
			name:  "unknown source",
			anime: &models.Anime{Name: "Namakura Gatana"},
			err:   netx.ErrNoEpisodes,
			want:  "Namakura Gatana has no episodes yet.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, failed := episodesUnavailableNotice(tt.anime, tt.err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantFailed, failed)
		})
	}
}
