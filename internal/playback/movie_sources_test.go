package playback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The movie-side source scheme mirrors the anime flow: when a movie's stream
// fails, the result screen the user is sent to explains why this copy failed
// and where the same movie is still listed. The playback panel is shared
// between StartFlix/TopCine/Pobreflix — the alternates respect that instead
// of pretending three sources exist where there is one panel.

func TestAlternateMovieSources(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		current string
		want    []string
	}{
		{"a StartFlix listing names Pobreflix", "StartFlix", []string{"Pobreflix"}},
		{"TopCine folds into StartFlixir", "TopCine", []string{"Pobreflix"}},
		{"a Pobreflix listing names StartFlix", "Pobreflix", []string{"StartFlix"}},
		{"no source something: both listings", "", []string{"StartFlix", "Pobreflix"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, alternateMovieSources(tt.current))
		})
	}
}

func TestAlternateMovieSourcesExcludesDisabled(t *testing.T) {
	t.Setenv("GOANIME_DISABLED_SOURCES", "Pobreflix")
	assert.Empty(t, alternateMovieSources("StartFlix"), "Pobreflix is on the kill-switch")
	assert.Empty(t, alternateMovieSources("TopCine"))
	assert.Equal(t, []string{"StartFlix"}, alternateMovieSources("Pobreflix"),
		"the kill-switch does not touch the remaining listing")
}

func TestAlternateMovieSourcesRegistered(t *testing.T) {
	t.Parallel()
	for _, k := range []apisource.SourceKind{apisource.StartFlix, apisource.Pobreflix} {
		s, ok := apisource.Registered(k)
		require.True(t, ok, "%s must be registered for the movie scheme to name it", k)
		assert.Equal(t, k, s.Describe().Kind)
	}
}

func TestMovieFriendlyErr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "friendly wording rides the chain",
			err: fmt.Errorf("failed to get TopCine stream URL: %w",
				api.Friendly(errors.New("dead server"),
					"TopCine only offers this on servers GoAnime can't play yet. Try it on another source.")),
			want: "TopCine only offers this on servers GoAnime can't play yet. Try it on another source.",
		},
		{
			name: "chain without friendly wording is returned whole",
			err: fmt.Errorf("failed to get TopCine stream URL: %w",
				errors.New("some internal detail")),
			want: "failed to get TopCine stream URL: some internal detail",
		},
		{name: "raw error survives as-is", err: errors.New("context deadline exceeded"),
			want: "context deadline exceeded"},
		{name: "nil error is an empty notice", err: nil, want: ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, movieFriendlyErr(tt.err))
		})
	}
}

func TestMovieStreamFailedNotice(t *testing.T) {
	t.Parallel()
	todo := &models.Anime{Name: "Todo Poderoso", Source: "TopCine"}

	t.Run("panel wording plus where else it is listed", func(t *testing.T) {
		err := fmt.Errorf("failed to get TopCine stream URL: %w",
			api.Friendly(errors.New("actelecup.com is dead"),
				"TopCine only offers this on dead servers. Try it later, or on another source."))
		notice := movieStreamFailedNotice(todo, err)
		assert.Contains(t, notice, "TopCine only offers this on dead servers")
		assert.Contains(t, notice, "The same movie is also listed on: Pobreflix.", "where the movie is listed now")
		assert.NotContains(t, notice, "failed to get", "the internal chain stays out of the notice")
		assert.NotContains(t, notice, "StartFlix", "StartFlix is the same panel, not another listing")
	})

	t.Run("raw error with alternates", func(t *testing.T) {
		notice := movieStreamFailedNotice(
			&models.Anime{Name: "Some Movie", Source: "Pobreflix"}, errors.New("context deadline exceeded"))
		assert.Contains(t, notice, "context deadline exceeded")
		assert.Contains(t, notice, "The same movie is also listed on: StartFlix.")
	})

	t.Run("nil anime keeps the wording", func(t *testing.T) {
		assert.Equal(t, "plain text", movieStreamFailedNotice(nil, errors.New("plain text")))
	})

	t.Run("nil error and nil anime", func(t *testing.T) {
		assert.Empty(t, movieStreamFailedNotice(nil, nil))
		assert.Empty(t, movieStreamFailedNotice(todo, nil), "no error, nothing to queue")
	})
}
