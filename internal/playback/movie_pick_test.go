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
	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err)
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
	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err)
	req, ok := player.TakeEpisodeMenu()
	require.True(t, ok, "the choice is left for the player, which downloads instead of playing")
	assert.Equal(t, 1, req.DownloadCode)
}

func TestPickMovieItemGoesBackToTheResults(t *testing.T) {
	withTerminal(t, true)
	withSelector(t, func([]models.Episode) (string, string, error) { return "", "", player.ErrBackRequested }, nil)
	_, err := pickMovieItem(movieEpisodes())
	require.ErrorIs(t, err, player.ErrBackToAnimeSelection)
}

func TestPickMovieItemPlaysWhenTheScreenCannotOpen(t *testing.T) {
	withTerminal(t, true)
	player.ClearEpisodeMenu()
	t.Cleanup(player.ClearEpisodeMenu)
	withSelector(t, func([]models.Episode) (string, string, error) {
		player.RequestEpisodeMenu(player.EpisodeMenuRequest{DownloadCode: 5}) // half-chosen, then it broke
		return "", "", errors.New("terminal too small")
	}, nil)
	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err, "a broken screen is no reason not to play")
	_, pending := player.TakeEpisodeMenu()
	assert.False(t, pending, "and what it left half-chosen is not taken for a choice")
}

func TestPickMovieItemDoesNotAskWithoutATerminal(t *testing.T) {
	withTerminal(t, false)
	withSelector(t, func([]models.Episode) (string, string, error) {
		t.Error("no screen is opened without a terminal")
		return "", "", nil
	}, nil)
	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err)
}

func TestTheRealFlowOpensTheScreenForOneItemTitles(t *testing.T) {
	env := realFlow()
	require.NotNil(t, env.pickMovie)
	withTerminal(t, false) // so calling the real one opens nothing
	_, err := env.pickMovie(movieEpisodes())
	assert.NoError(t, err)
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

	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err)
	assert.Zero(t, shown, "CONTINUE already chose the item")
	assert.False(t, PlayOneItemDirectly, "and only once")

	_, err = pickMovieItem(movieEpisodes())
	require.NoError(t, err)
	assert.Equal(t, 1, shown, "the next title asks again")
}

func TestPickMovieItemConsumesTheContinueSkipEvenWithoutATerminal(t *testing.T) {
	withTerminal(t, false)
	PlayOneItemDirectly = true
	t.Cleanup(func() { PlayOneItemDirectly = false })
	_, err := pickMovieItem(movieEpisodes())
	require.NoError(t, err)
	assert.False(t, PlayOneItemDirectly, "not left over for a later title")
}

func TestTheRealTerminalCheckAnswersWithoutTouchingTheScreen(t *testing.T) {
	// The default reads os.Stdin's mode: whatever it answers here, asking must
	// not fail or block.
	assert.NotPanics(t, func() { _ = movieStdinIsTerminal() })
}

// ---- a movie listed in several entries ---------------------------------------

func movieParts() []models.Episode {
	return []models.Episode{
		{Number: "1", Num: 1, URL: "https://example.test/film/1", Duration: 60},
		{Number: "2", Num: 2, URL: "https://example.test/film/2", Duration: 70},
		{Number: "Extra", URL: "https://example.test/film/extra", Duration: 80},
	}
}

func TestRunMoviePlaysTheEntryThatWasPicked(t *testing.T) {
	for idx, want := range []struct {
		number string
		num    int
		dur    int
	}{{"1", 1, 60}, {"2", 2, 70}, {"Extra", 3, 80}} {
		f := newFakeFlow()
		f.pickIdx = idx
		a := movieAnime()
		require.NoError(t, runMovie(f.env, context.Background(), a, movieParts(), false))

		require.Len(t, f.moviePlays, 1)
		p := f.moviePlays[0]
		assert.Equal(t, []string{want.number}, f.streamedEps, "entry %d: the stream resolved is the picked entry's", idx)
		assert.Equal(t, want.number, p.number, "entry %d", idx)
		assert.Equal(t, want.num, p.selected, "entry %d: the number the player and the downloader are told", idx)
		assert.Len(t, p.episodes, 3, "the whole list still goes along")
		require.Len(t, p.anime.Episodes, 1)
		assert.Equal(t, want.number, p.anime.Episodes[0].Number, "entry %d: the title's current entry", idx)
		require.Len(t, f.updaterArgs, 1)
		assert.Equal(t, want.dur, int(f.updaterArgs[0].duration.Seconds()), "entry %d: its own duration", idx)
	}
}

func TestRunMovieFallsBackToTheFirstEntryForAnAnswerOutOfRange(t *testing.T) {
	for _, idx := range []int{-1, 3, 99} {
		f := newFakeFlow()
		f.pickIdx = idx
		require.NoError(t, runMovie(f.env, context.Background(), movieAnime(), movieParts(), false))
		assert.Equal(t, []string{"1"}, f.streamedEps, "index %d", idx)
		assert.Equal(t, 1, f.moviePlays[0].selected)
	}
}

func TestRunMovieReplaysTheSameEntry(t *testing.T) {
	f := newFakeFlow()
	f.pickIdx = 1
	f.plays = []error{nil, player.ErrUserQuit}
	f.inputs = []string{"n"}
	require.NoError(t, runMovie(f.env, context.Background(), movieAnime(), movieParts(), false))
	assert.Equal(t, []string{"2", "2"}, f.streamedEps, "a replay is of what was playing")
}

func TestPickMovieItemAnswersTheEntryPicked(t *testing.T) {
	withTerminal(t, true)
	parts := movieParts()
	// One selector for the whole test: withSelector holds its lock until the
	// test ends, so it is installed once and told what to answer.
	var answer models.Episode
	withSelector(t, func([]models.Episode) (string, string, error) { return answer.URL, answer.Number, nil },
		func(string) string { return "1" })
	for want, part := range parts {
		answer = part
		got, err := pickMovieItem(parts)
		require.NoError(t, err)
		assert.Equal(t, want, got, part.Number)
	}
}

func TestPickedEntry(t *testing.T) {
	eps := []models.Episode{
		{Number: "1", URL: "u"}, {Number: "2", URL: "u"}, // one page, two numbers
		{Number: "3", URL: "v"}, {Number: "3", URL: "w"}, // one number, two pages
	}
	assert.Equal(t, 1, pickedEntry(eps, "u", "2"), "both name it")
	assert.Equal(t, 3, pickedEntry(eps, "w", "3"))
	assert.Equal(t, 2, pickedEntry(eps, "v", "?"), "the URL alone")
	assert.Equal(t, 0, pickedEntry(eps, "u", "?"), "the first with that URL")
	assert.Equal(t, 2, pickedEntry(eps, "", "3"), "the number alone, the first with it")
	assert.Equal(t, 0, pickedEntry(eps, "nowhere", "9"), "nothing names it: the first")
	assert.Equal(t, 0, pickedEntry(eps, "", ""))
	assert.Equal(t, 0, pickedEntry(nil, "u", "1"))
}

func TestMoviePartNumber(t *testing.T) {
	parts := movieParts()
	assert.Equal(t, 1, moviePartNumber(parts, 0), "a movie's first entry is 1, as it always was")
	assert.Equal(t, 2, moviePartNumber(parts, 1), "its own number")
	assert.Equal(t, 3, moviePartNumber(parts, 2), "its place, when it has no number")
	assert.Equal(t, 1, moviePartNumber(parts, -1))
	assert.Equal(t, 1, moviePartNumber(parts, 3))
	assert.Equal(t, 1, moviePartNumber(nil, 0))
	assert.Equal(t, 1, moviePartNumber([]models.Episode{{Num: 7}}, 0), "a single entry is 1 whatever it calls itself")
}
