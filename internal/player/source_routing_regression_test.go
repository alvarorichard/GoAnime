package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// Regression tests (added 2026-05-01)
//
// Context: triage on a movie playback failure was misled by a debug log that
// announced "FlixHQ source detected" for a non-FlixHQ movie. The helper that
// routes movie/TV stream extraction was named isFlixHQSourcePlayer but actually
// returned true for any MediaTypeMovie/MediaTypeTV regardless of source. After
// the rename the routing semantics are unchanged, but the helper name and log
// labels reflect that the branch is shared.

func TestIsMovieOrTVSourcePlayer_FlixHQ_2026_05_01(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "SFlix", MediaType: models.MediaTypeMovie}
	assert.True(t, isMovieOrTVSourcePlayer(anime))
}

func TestIsMovieOrTVSourcePlayer_AnimeFireAnime_2026_05_01(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "Animefire.io", MediaType: models.MediaTypeAnime}
	assert.False(t, isMovieOrTVSourcePlayer(anime),
		"anime content must not be misrouted through the movie/TV branch")
}

func TestIsMovieOrTVSourcePlayer_NilAnime_2026_05_01(t *testing.T) {
	t.Parallel()
	assert.False(t, isMovieOrTVSourcePlayer(nil))
}
