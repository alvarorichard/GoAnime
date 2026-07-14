package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SuperFlix must be recognized by SOURCE, not by media type: its streams are
// multi-audio HLS with an external subtitle track whether the entry is a movie,
// a series, an anime or a dorama. Gating mpv's language preferences on
// IsMovieOrTV silently dropped the chosen audio track and the subtitles for
// every SuperFlix anime.
func TestIsSuperFlixSource(t *testing.T) {
	prev := GlobalAnimeSource
	t.Cleanup(func() { GlobalAnimeSource = prev })

	tests := []struct {
		source string
		want   bool
	}{
		{"SuperFlix", true},
		{"AnimeFire", false},
		{"AllAnime", false},
		{"Goyabu", false},
		{"9Anime", false},
		{"", false},
	}
	for _, tt := range tests {
		SetGlobalAnimeSource(tt.source)
		assert.Equal(t, tt.want, IsSuperFlixSource(), "source=%q", tt.source)
	}
}
