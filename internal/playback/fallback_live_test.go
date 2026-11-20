//go:build live

package playback

import (
	"context"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// Against the real sources: the same episode of a title that failed on one
// source, looked for on the others. Nothing is asserted about what the sources
// answer today (they come and go); the log says what happened, and the test
// fails only if the cascade panics or never returns.
//
//	go test -tags live -run Live -v ./internal/playback
func TestLiveFallbackAcrossRealSources(t *testing.T) {
	for _, tc := range []struct{ title, failedOn string }{
		{"Wind Breaker", "Animefire.io"},
		{"Karakai Jouzu no Takagi-san 2", "AnimesDigital"},
		{"Naruto", "Animefire.io"},
	} {
		t.Run(tc.title+" failed on "+tc.failedOn, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			anime := &models.Anime{Name: tc.title, Source: tc.failedOn, MediaType: models.MediaTypeAnime}

			start := time.Now()
			url, from, tried := videoFromAnotherSource(ctx, anime, "1", 1)

			t.Logf("tried %v in %s", tried, time.Since(start).Round(time.Millisecond))
			if url == "" {
				t.Logf("no other source gave episode 1")
				return
			}
			t.Logf("episode 1 plays from %s: %s", from, url)
		})
	}
}
