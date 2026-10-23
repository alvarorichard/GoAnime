package playback

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This cascade walks every failure shape the movie flow can hand back — the
// api layer's friendly rephrasings and the raw errors past them — and pins
// what lands on the result screen for each: the plain wording first, the
// stray alternates attached, and never the internal chain or a drawn emoji.
// Every shape here is built exactly the way the api layers return it: an
// error chain whose friendly wording sits typed inside (api.Friendly), the
// same shape GetTopCineStreamURL, GetStartFlixStreamURL and friends hand to
// reportMovieStreamFailed.

func TestMovieFailureCascade(t *testing.T) {
	t.Parallel()
	todo := &models.Anime{Name: "Todo Poderoso", Source: "TopCine"}

	shapes := []struct {
		name     string
		err      error
		wantLead string // the wording that must open the notice
	}{
		{
			name: "no player for the title",
			err: fmt.Errorf("failed to get TopCine stream URL: %w",
				api.Friendly(errors.New("topcine.ErrNoPlayer"),
					"TopCine has no player for this title yet. Try searching it on another source.")),
			wantLead: "TopCine has no player for this title yet.",
		},
		{
			name: "the panel carries it only on a browser-checked server",
			err: fmt.Errorf("failed to get TopCine stream URL: %w",
				api.Friendly(errors.New("1take.top floods the check"),
					"TopCine plays this title only on 1take.top, 2take.top, which asks for a browser check GoAnime can't pass. Try searching it on another source.")),
			wantLead: "TopCine plays this title only on 1take.top, 2take.top",
		},
		{
			name: "the other audio is refused before it plays",
			err: fmt.Errorf("audio unavailable: %w",
				api.Friendly(errors.New("swap refused"),
					"No server GoAnime can play has this title Dublado right now, so the Legendado was not played in its place.")),
			wantLead: "No server GoAnime can play has this title Dublado",
		},
		{
			name:     "a raw error past all the rephrasing survives whole",
			err:      errors.New("context deadline exceeded"),
			wantLead: "context deadline exceeded",
		},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			notice := movieStreamFailedNotice(todo, shape.err)

			require.True(t, strings.HasPrefix(notice, shape.wantLead),
				"the wording leads the notice, got: %q", notice)
			assert.NotContains(t, notice, "failed to get", "the internal chain stays out")
			assert.NotContains(t, notice, "audio unavailable", "no jargon rides the chain either")
			assert.NotContains(t, notice, "swap refused", "the internal causes stay out")
			for _, drawn := range []string{"⚠", "🚀", "★", "💬", "🎙", "↩"} {
				assert.NotContainsf(t, notice, drawn,
					"the notice is plain text — %q never lands on a console", drawn)
			}
			assert.True(t, utf8.ValidString(notice), "the notice is valid utf-8 end to end")
			assert.Contains(t, notice, "The same movie is also listed on: Pobreflix.",
				"TopCine fades into the one sibling that lists it")
			assert.NotContains(t, notice, "StartFlix", "the panel is the same player, not another listing")
		})
	}

	t.Run("a Pobreflix failure names StartFlix alone", func(t *testing.T) {
		t.Parallel()
		notice := movieStreamFailedNotice(
			&models.Anime{Name: "A Primeira Vista", Source: "Pobreflix"},
			fmt.Errorf("failed to get Pobreflix stream URL: %w",
				api.Friendly(errors.New("actelecup.com is dead"),
					"Pobreflix has this movie only on servers that no longer answer (actelecup.com). Try again later, or search it on another source.")))

		require.True(t, strings.HasPrefix(notice,
			"Pobreflix has this movie only on servers that no longer answer"),
			"the wording leads, got: %q", notice)
		assert.Contains(t, notice, "The same movie is also listed on: StartFlix.",
			"StartFlix still lists the movie")
		assert.NotContains(t, notice, "TopCine", "TopCine is never offered as a rescue")
		assert.NotContains(t, notice, "failed to get", "the chain stays out of the screen")
	})

	t.Run("nothing arrives with nothing to say", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, movieStreamFailedNotice(todo, nil), "no error, no notice")
		assert.Equal(t, "only the wording", movieStreamFailedNotice(nil, errors.New("only the wording")),
			"a nil anime passes the error through whole")
	})
}
