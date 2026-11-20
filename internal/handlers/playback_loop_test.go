package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #229: the user searched "karakai-jouzu", changed the title during
// playback to "wind breaker", and backing out of it landed on the results for
// "karakai-jouzu". The loop repeated the search it began with. These tests run
// the loop against stand-ins for the sources and the player.
// Not parallel: util.GlobalSource is process-wide.

type searchCall struct{ name, scope string }

// loopRig is a playbackEnv that records what the loop asks of it. last stands
// for appflow's record of the search that found the title now playing.
type loopRig struct {
	env      playbackEnv
	searches []searchCall
	played   []string // "series:<name>" / "movie:<name>", in order
	discord  []bool
	last     *searchCall
	homeDone bool

	// scripted answers
	searchErr    func(call int) error
	nilAnime     func(call int) bool
	episodes     func(*models.Anime) ([]models.Episode, error)
	onSeries     func(call int, anime *models.Anime) error
	onMovie      func(call int, anime *models.Anime) error
	drainHomeNow func() bool
}

func manyEpisodes(n int) []models.Episode {
	eps := make([]models.Episode, n)
	for i := range eps {
		eps[i] = models.Episode{Num: i + 1}
	}
	return eps
}

func newLoopRig(t *testing.T) *loopRig {
	t.Helper()
	r := &loopRig{
		episodes: func(*models.Anime) ([]models.Episode, error) { return manyEpisodes(12), nil },
	}
	var seriesCalls, movieCalls int
	r.env = playbackEnv{
		search: func(name, scope string) (*models.Anime, error) {
			r.searches = append(r.searches, searchCall{name, scope})
			call := len(r.searches)
			if r.searchErr != nil {
				if err := r.searchErr(call); err != nil {
					return nil, err
				}
			}
			if r.nilAnime != nil && r.nilAnime(call) {
				return nil, nil
			}
			// The search that finds a title is the one to come back to.
			r.last = &searchCall{name, scope}
			return &models.Anime{Name: name, Source: "AnimeFire"}, nil
		},
		fetchDetails: func(*models.Anime) {},
		episodes:     func(a *models.Anime) ([]models.Episode, error) { return r.episodes(a) },
		handleSeries: func(_ context.Context, a *models.Anime, _ []models.Episode, _ int, discordOn bool) error {
			seriesCalls++
			r.played = append(r.played, "series:"+a.Name)
			r.discord = append(r.discord, discordOn)
			if r.onSeries != nil {
				return r.onSeries(seriesCalls, a)
			}
			return nil
		},
		handleMovie: func(_ context.Context, a *models.Anime, _ []models.Episode, discordOn bool) error {
			movieCalls++
			r.played = append(r.played, "movie:"+a.Name)
			r.discord = append(r.discord, discordOn)
			if r.onMovie != nil {
				return r.onMovie(movieCalls, a)
			}
			return nil
		},
		drainHome: func() bool {
			if r.drainHomeNow != nil {
				return r.drainHomeNow()
			}
			return false
		},
		repeatSearch: func(name, scope string) (string, string) {
			if r.last == nil {
				return name, scope
			}
			return r.last.name, r.last.scope
		},
		discordOn: func() bool { return true },
	}
	prev := util.GlobalSource
	util.GlobalSource = ""
	t.Cleanup(func() { util.GlobalSource = prev })
	return r
}

func (r *loopRig) run(animeName string) { runPlaybackLoop(context.Background(), r.env, animeName) }

// changeTitle simulates "c" during playback: another title is searched, found
// and played in place of the first, and backing out of it ends the series.
func (r *loopRig) changeTitle(to searchCall) {
	r.last = &to
}

func TestBackingOutAfterChangingTheTitleShowsTheChangedSearch(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(call int, _ *models.Anime) error {
		if call == 1 {
			r.changeTitle(searchCall{"wind breaker", ""}) // the user pressed c and searched this
			return player.ErrBackToAnimeSelection
		}
		return nil
	}

	r.run("karakai-jouzu")

	require.Len(t, r.searches, 2)
	assert.Equal(t, searchCall{"karakai-jouzu", ""}, r.searches[0])
	assert.Equal(t, searchCall{"wind breaker", ""}, r.searches[1],
		"the results the user comes back to are the ones the title they left was in")
}

func TestBackingOutKeepsTheScopeOfTheChangedSearch(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(call int, _ *models.Anime) error {
		if call == 1 {
			r.changeTitle(searchCall{"wind breaker", "hianime"})
			return player.ErrBackToAnimeSelection
		}
		return nil
	}

	r.run("karakai-jouzu")

	require.Len(t, r.searches, 2)
	assert.Equal(t, "hianime", r.searches[1].scope, "a search limited to a source is repeated limited to it")
}

func TestBackingOutWithoutChangingTheTitleRepeatsTheSameSearch(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(call int, _ *models.Anime) error {
		if call == 1 {
			return player.ErrBackToAnimeSelection
		}
		return nil
	}

	r.run("karakai-jouzu")

	require.Len(t, r.searches, 2)
	assert.Equal(t, searchCall{"karakai-jouzu", ""}, r.searches[1])
}

func TestEachBackFollowsTheLatestSearch(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(call int, _ *models.Anime) error {
		switch call {
		case 1:
			r.changeTitle(searchCall{"wind breaker", ""})
			return player.ErrBackToAnimeSelection
		case 2:
			r.changeTitle(searchCall{"naruto", "hianime"})
			return player.ErrBackToAnimeSelection
		}
		return nil
	}

	r.run("karakai-jouzu")

	assert.Equal(t, []searchCall{
		{"karakai-jouzu", ""}, {"wind breaker", ""}, {"naruto", "hianime"},
	}, r.searches)
}

// A title the source lists but cannot open sends the user back to the results
// too; if the search was edited on the way to it, those are the edited ones.
func TestAnUnopenableTitleSendsTheUserBackToTheResultsItWasIn(t *testing.T) {
	r := newLoopRig(t)
	opened := 0
	r.episodes = func(a *models.Anime) ([]models.Episode, error) {
		opened++
		if opened == 1 {
			r.changeTitle(searchCall{"wind breaker", "hianime"}) // the search was edited to find this
			return nil, errors.New("no episodes for this title")
		}
		return manyEpisodes(3), nil
	}

	r.run("wind braker")

	require.Len(t, r.searches, 2)
	assert.Equal(t, searchCall{"wind breaker", "hianime"}, r.searches[1])
}

func TestATitleWithNoEpisodesSendsTheUserBackToTheResults(t *testing.T) {
	r := newLoopRig(t)
	opened := 0
	r.episodes = func(*models.Anime) ([]models.Episode, error) {
		opened++
		if opened == 1 {
			return nil, nil // listed, with nothing in it
		}
		return manyEpisodes(3), nil
	}

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 2)
	assert.Equal(t, []string{"series:karakai-jouzu"}, r.played, "nothing played from the empty title")
}

func TestBackingOutOfTheSeasonPickerOpensAFreshSearchInTheGlobalScope(t *testing.T) {
	r := newLoopRig(t)
	util.GlobalSource = "anroll"
	opened := 0
	r.episodes = func(*models.Anime) ([]models.Episode, error) {
		opened++
		if opened == 1 {
			return nil, api.ErrBackToSearch
		}
		return manyEpisodes(3), nil
	}
	r.last = &searchCall{"stale", "hianime"} // must not be repeated: the user asked for a new search

	r.run("karakai-jouzu")

	require.Len(t, r.searches, 2)
	assert.Equal(t, searchCall{"", "anroll"}, r.searches[1], "an empty name opens the search prompt, over the scope the session began with")
}

func TestASearchTheUserQuitEndsTheLoop(t *testing.T) {
	r := newLoopRig(t)
	r.searchErr = func(int) error { return api.ErrSearchAborted }

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 1)
	assert.Empty(t, r.played)
}

func TestASearchThatFailsEndsTheLoop(t *testing.T) {
	r := newLoopRig(t)
	r.searchErr = func(int) error { return errors.New("every source failed") }

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 1)
	assert.Empty(t, r.played)
}

func TestAContinueRowPendedByTheSearchEndsTheLoop(t *testing.T) {
	r := newLoopRig(t)
	r.nilAnime = func(int) bool { return true }
	r.drainHomeNow = func() bool { r.homeDone = true; return true }

	r.run("")

	assert.True(t, r.homeDone)
	assert.Len(t, r.searches, 1)
	assert.Empty(t, r.played)
}

func TestAContinueRowPendedWhileBackingOutEndsTheLoop(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(call int, _ *models.Anime) error {
		if call == 1 {
			return player.ErrBackToAnimeSelection
		}
		return errors.New("the loop went on after the pended row") // ends a loop that did not stop
	}
	r.drainHomeNow = func() bool { return true }

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 1, "the pended row runs instead of another search")
	assert.Len(t, r.played, 1)
}

func TestTheLoopEndsWhenPlaybackEndsNormally(t *testing.T) {
	r := newLoopRig(t)

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 1)
	assert.Equal(t, []string{"series:karakai-jouzu"}, r.played)
}

func TestTheLoopEndsOnAPlaybackErrorThatIsNotABack(t *testing.T) {
	r := newLoopRig(t)
	r.onSeries = func(int, *models.Anime) error { return errors.New("mpv would not start") }

	r.run("karakai-jouzu")

	assert.Len(t, r.searches, 1, "an error is not a request for another search")
}

func TestASeriesAndAMovieGoToTheirOwnHandlers(t *testing.T) {
	t.Run("several episodes is a series", func(t *testing.T) {
		r := newLoopRig(t)
		r.episodes = func(*models.Anime) ([]models.Episode, error) { return manyEpisodes(2), nil }
		r.run("show")
		assert.Equal(t, []string{"series:show"}, r.played)
	})
	t.Run("one episode is a movie", func(t *testing.T) {
		r := newLoopRig(t)
		r.episodes = func(*models.Anime) ([]models.Episode, error) { return manyEpisodes(1), nil }
		r.run("film")
		assert.Equal(t, []string{"movie:film"}, r.played)
	})
	t.Run("a title marked as a movie is one whatever its episode count", func(t *testing.T) {
		r := newLoopRig(t)
		r.episodes = func(*models.Anime) ([]models.Episode, error) { return manyEpisodes(5), nil }
		inner := r.env.search
		r.env.search = func(name, scope string) (*models.Anime, error) {
			a, err := inner(name, scope)
			a.MediaType = models.MediaTypeMovie
			return a, err
		}
		r.run("film")
		assert.Equal(t, []string{"movie:film"}, r.played)
	})
}

func TestThePlayersDiscordSettingReachesTheHandlers(t *testing.T) {
	r := newLoopRig(t)
	r.env.discordOn = func() bool { return true }
	r.run("show")
	assert.Equal(t, []bool{true}, r.discord)

	r = newLoopRig(t)
	r.env.discordOn = func() bool { return false }
	r.run("show")
	assert.Equal(t, []bool{false}, r.discord)

	// a movie gets the same setting
	r = newLoopRig(t)
	r.episodes = func(*models.Anime) ([]models.Episode, error) { return manyEpisodes(1), nil }
	r.env.discordOn = func() bool { return true }
	r.run("film")
	assert.Equal(t, []string{"movie:film"}, r.played)
	assert.Equal(t, []bool{true}, r.discord)
}

func TestTheLoopStartsInTheGlobalScope(t *testing.T) {
	r := newLoopRig(t)
	util.GlobalSource = "hianime"

	r.run("naruto")

	assert.Equal(t, []searchCall{{"naruto", "hianime"}}, r.searches, "--source is honoured by the first search")
}

func TestAFreshSearchAfterAChangedScopeGoesBackToTheGlobalScope(t *testing.T) {
	r := newLoopRig(t)
	util.GlobalSource = ""
	opened := 0
	r.onSeries = func(call int, _ *models.Anime) error {
		if call == 1 {
			r.changeTitle(searchCall{"wind breaker", "hianime"}) // a search limited to one source
			return player.ErrBackToAnimeSelection
		}
		return nil
	}
	r.episodes = func(*models.Anime) ([]models.Episode, error) {
		opened++
		if opened == 2 { // the second title: the season picker is backed out of
			return nil, api.ErrBackToSearch
		}
		return manyEpisodes(3), nil
	}

	r.run("karakai-jouzu")

	require.Len(t, r.searches, 3)
	assert.Equal(t, searchCall{"wind breaker", "hianime"}, r.searches[1])
	assert.Equal(t, searchCall{"", ""}, r.searches[2], "the new search is over every source again, not stuck on the last one's scope")
}
