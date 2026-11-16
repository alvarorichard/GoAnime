package playback

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
)

// A title of one item (a movie, or a series season of one episode) gets the
// episodes screen before it plays, as every series does, so the download and
// option keys are there for it too. It used to play at once, from every source.

// ---- runMovie --------------------------------------------------------------

func TestRunMovieOpensTheEpisodesScreenBeforeAnythingIsResolved(t *testing.T) {
	f := newFakeFlow()
	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	require.NoError(t, err)

	require.Len(t, f.picks, 1, "the screen opens once")
	assert.Equal(t, 2, f.picks[0], "after the mpv prewarm and the list context, before the stream: prewarm,listctx | stream")
	assert.Equal(t, "prewarm,listctx:Film,stream,play", f.trace())
	require.Len(t, f.pickedEps, 1)
	assert.Equal(t, movieEpisodes(), f.pickedEps[0], "the screen lists the title's own item")
}

func TestRunMovieGoesBackToTheResultsWhenTheScreenIsLeft(t *testing.T) {
	f := newFakeFlow()
	f.pickErr = player.ErrBackToAnimeSelection
	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	require.ErrorIs(t, err, player.ErrBackToAnimeSelection)
	assert.Equal(t, "prewarm,listctx:Film", f.trace(), "nothing is resolved or played")
	assert.Empty(t, f.moviePlays)
	assert.Empty(t, f.streamedEps)
}

func TestRunMovieDoesNotAskAgainForAReplay(t *testing.T) {
	f := newFakeFlow()
	f.plays = []error{nil, player.ErrUserQuit}
	f.inputs = []string{"n"} // "Replay movie"
	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	require.NoError(t, err)
	assert.Len(t, f.moviePlays, 2, "played twice")
	assert.Len(t, f.picks, 1, "but the screen opened only before the first")
}

func TestRunMovieAsksAgainForAnotherMovie(t *testing.T) {
	f := newFakeFlow()
	c := &changerMock{anime: &models.Anime{Name: "Other", MediaType: models.MediaTypeMovie}, eps: []models.Episode{{Number: "Other", Num: 1}}}
	withChanger(t, c)
	f.plays = []error{nil, player.ErrUserQuit}
	f.inputs = []string{"c"}

	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	require.NoError(t, err)
	require.Len(t, f.picks, 2, "the new title gets its own screen")
	assert.Equal(t, c.eps, f.pickedEps[1], "listing the new title's item")
	assert.Equal(t, []string{"Film", "Other"}, f.streamedEps)
}

func TestRunMovieKeepsTheCurrentTitleWithoutAskingWhenAChangeFails(t *testing.T) {
	f := newFakeFlow()
	withChanger(t, &changerMock{err: errors.New("no network")})
	f.plays = []error{player.ErrChangeAnime, player.ErrUserQuit}
	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	require.NoError(t, err)
	assert.Len(t, f.moviePlays, 2)
	assert.Len(t, f.picks, 1, "staying on the same title is not a new one")
}

// ---- pickMovieItem ---------------------------------------------------------

func withTerminal(t *testing.T, interactive bool) {
	t.Helper()
	prev := movieStdinIsTerminal
	movieStdinIsTerminal = func() bool { return interactive }
	t.Cleanup(func() { movieStdinIsTerminal = prev })
}

func TestPickMovieItemShowsTheScreenAndLetsEnterPlay(t *testing.T) {
	withTerminal(t, true)
	shown := 0
	withSelector(t, func(eps []models.Episode) (string, string, error) {
		shown++
		assert.Len(t, eps, 1)
		return eps[0].URL, "1", nil
	}, nil)
	require.NoError(t, pickMovieItem(movieEpisodes()))
	assert.Equal(t, 1, shown)
}

func TestPickMovieItemCarriesTheOptionChosenOnTheScreen(t *testing.T) {
	withTerminal(t, true)
	player.ClearEpisodeMenu()
	t.Cleanup(player.ClearEpisodeMenu)
	withSelector(t, func(eps []models.Episode) (string, string, error) {
		// what the d key does on the real screen
		player.RequestEpisodeMenu(player.EpisodeMenuRequest{DownloadCode: 1})
		return eps[0].URL, "1", nil
	}, nil)
	require.NoError(t, pickMovieItem(movieEpisodes()))
	req, ok := player.TakeEpisodeMenu()
	require.True(t, ok, "the choice is left for the player, which downloads instead of playing")
	assert.Equal(t, 1, req.DownloadCode)
}

func TestPickMovieItemGoesBackToTheResults(t *testing.T) {
	withTerminal(t, true)
	withSelector(t, func([]models.Episode) (string, string, error) { return "", "", player.ErrBackRequested }, nil)
	require.ErrorIs(t, pickMovieItem(movieEpisodes()), player.ErrBackToAnimeSelection)
}

func TestPickMovieItemPlaysWhenTheScreenCannotOpen(t *testing.T) {
	withTerminal(t, true)
	player.ClearEpisodeMenu()
	t.Cleanup(player.ClearEpisodeMenu)
	withSelector(t, func([]models.Episode) (string, string, error) {
		player.RequestEpisodeMenu(player.EpisodeMenuRequest{DownloadCode: 5}) // half-chosen, then it broke
		return "", "", errors.New("terminal too small")
	}, nil)
	require.NoError(t, pickMovieItem(movieEpisodes()), "a broken screen is no reason not to play")
	_, pending := player.TakeEpisodeMenu()
	assert.False(t, pending, "and what it left half-chosen is not taken for a choice")
}

func TestPickMovieItemDoesNotAskWithoutATerminal(t *testing.T) {
	withTerminal(t, false)
	withSelector(t, func([]models.Episode) (string, string, error) {
		t.Error("no screen is opened without a terminal")
		return "", "", nil
	}, nil)
	require.NoError(t, pickMovieItem(movieEpisodes()))
}

func TestTheRealFlowOpensTheScreenForOneItemTitles(t *testing.T) {
	env := realFlow()
	require.NotNil(t, env.pickMovie)
	withTerminal(t, false) // so calling the real one opens nothing
	assert.NoError(t, env.pickMovie(movieEpisodes()))
}

func TestPickMovieItemSkipsTheScreenOnceForAContinueReopen(t *testing.T) {
	withTerminal(t, true)
	shown := 0
	withSelector(t, func(eps []models.Episode) (string, string, error) {
		shown++
		return eps[0].URL, "1", nil
	}, nil)
	PlayOneItemDirectly = true
	t.Cleanup(func() { PlayOneItemDirectly = false })

	require.NoError(t, pickMovieItem(movieEpisodes()))
	assert.Zero(t, shown, "CONTINUE already chose the item")
	assert.False(t, PlayOneItemDirectly, "and only once")

	require.NoError(t, pickMovieItem(movieEpisodes()))
	assert.Equal(t, 1, shown, "the next title asks again")
}

func TestPickMovieItemConsumesTheContinueSkipEvenWithoutATerminal(t *testing.T) {
	withTerminal(t, false)
	PlayOneItemDirectly = true
	t.Cleanup(func() { PlayOneItemDirectly = false })
	require.NoError(t, pickMovieItem(movieEpisodes()))
	assert.False(t, PlayOneItemDirectly, "not left over for a later title")
}
