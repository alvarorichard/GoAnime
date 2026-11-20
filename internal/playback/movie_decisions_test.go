package playback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/appflow"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- movieActionAfterPlayback --------------------------------------------------

var afterPlaybackCases = []struct {
	name string
	err  error
	want movieAction
}{
	{"finished normally", nil, movieAskUser},
	{"user quit", player.ErrUserQuit, movieQuit},
	{"user quit, wrapped", fmt.Errorf("mpv: %w", player.ErrUserQuit), movieQuit},
	{"change title from the player", player.ErrChangeAnime, movieChange},
	{"change title, wrapped", fmt.Errorf("menu: %w", player.ErrChangeAnime), movieChange},
	{"a playback error still asks", errors.New("mpv crashed"), movieAskUser},
	{"back to selection is not handled here", player.ErrBackToAnimeSelection, movieAskUser},
	{"back to download options asks", player.ErrBackToDownloadOptions, movieAskUser},
}

func afterPlaybackFailures(impl func(error) movieAction) []string {
	var f []string
	for _, tc := range afterPlaybackCases {
		if got := impl(tc.err); got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d want %d", tc.name, got, tc.want))
		}
	}
	return f
}

func TestMovieActionAfterPlayback(t *testing.T) {
	assert.Empty(t, afterPlaybackFailures(movieActionAfterPlayback))
}

func TestMovieActionAfterPlaybackMeta(t *testing.T) {
	require.Empty(t, afterPlaybackFailures(movieActionAfterPlayback))
	mutants := map[string]func(error) movieAction{
		"always ask":  func(error) movieAction { return movieAskUser },
		"always quit": func(error) movieAction { return movieQuit },
		"quit and change swapped": func(e error) movieAction {
			switch movieActionAfterPlayback(e) {
			case movieQuit:
				return movieChange
			case movieChange:
				return movieQuit
			}
			return movieActionAfterPlayback(e)
		},
		"compares by equality, not errors.Is": func(e error) movieAction {
			switch e {
			case player.ErrUserQuit:
				return movieQuit
			case player.ErrChangeAnime:
				return movieChange
			}
			return movieAskUser
		},
		"any error quits": func(e error) movieAction {
			if e != nil {
				return movieQuit
			}
			return movieAskUser
		},
		"change ignored": func(e error) movieAction {
			if errors.Is(e, player.ErrChangeAnime) {
				return movieAskUser
			}
			return movieActionAfterPlayback(e)
		},
		"back to selection quits": func(e error) movieAction {
			if errors.Is(e, player.ErrBackToAnimeSelection) {
				return movieQuit
			}
			return movieActionAfterPlayback(e)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, afterPlaybackFailures(m), "mutant %q survived", name)
	}
}

// ---- movieActionAfterPrompt ----------------------------------------------------

var afterPromptCases = []struct {
	input string
	want  movieAction
}{
	{"q", movieQuit},
	{"c", movieChange},
	{"back", movieChange},
	{"n", movieReplay}, // "Replay movie" answers n
	{"", movieReplay},
	{"p", movieReplay},
	{"e", movieReplay},
	{"Q", movieReplay},
	{"quit", movieReplay},
	{"C", movieReplay},
}

func afterPromptFailures(impl func(string) movieAction) []string {
	var f []string
	for _, tc := range afterPromptCases {
		if got := impl(tc.input); got != tc.want {
			f = append(f, fmt.Sprintf("%q: got %d want %d", tc.input, got, tc.want))
		}
	}
	return f
}

func TestMovieActionAfterPrompt(t *testing.T) {
	assert.Empty(t, afterPromptFailures(movieActionAfterPrompt))
}

// The menu's own answers must all be understood: the values GetUserInput can
// return for a movie map to the action each label promises.
func TestMovieMenuAnswersMatchTheLabels(t *testing.T) {
	prev := findMenuFunc
	t.Cleanup(func() { findMenuFunc = prev })
	var labels []string
	findMenuFunc = func(items []menuItem) (int, error) {
		for _, it := range items {
			labels = append(labels, it.Label)
		}
		return -1, nil
	}
	GetUserInput(true)
	require.Equal(t, []string{"Replay movie", "Change movie", "← Back", "Quit"}, labels)

	want := map[string]movieAction{"Replay movie": movieReplay, "Change movie": movieChange, "← Back": movieChange, "Quit": movieQuit}
	for i, label := range labels {
		findMenuFunc = func([]menuItem) (int, error) { return i, nil }
		assert.Equal(t, want[label], movieActionAfterPrompt(GetUserInput(true)), label)
	}
}

func TestMovieActionAfterPromptMeta(t *testing.T) {
	require.Empty(t, afterPromptFailures(movieActionAfterPrompt))
	mutants := map[string]func(string) movieAction{
		"always replay": func(string) movieAction { return movieReplay },
		"back quits": func(s string) movieAction {
			if s == "back" {
				return movieQuit
			}
			return movieActionAfterPrompt(s)
		},
		"back replays": func(s string) movieAction {
			if s == "back" {
				return movieReplay
			}
			return movieActionAfterPrompt(s)
		},
		"c ignored": func(s string) movieAction {
			if s == "c" {
				return movieReplay
			}
			return movieActionAfterPrompt(s)
		},
		"case-insensitive": func(s string) movieAction {
			switch s {
			case "Q":
				return movieQuit
			case "C":
				return movieChange
			}
			return movieActionAfterPrompt(s)
		},
		"unknown quits": func(s string) movieAction {
			if a := movieActionAfterPrompt(s); a != movieReplay {
				return a
			}
			if s == "n" {
				return movieReplay
			}
			return movieQuit
		},
		"n changes": func(s string) movieAction {
			if s == "n" {
				return movieChange
			}
			return movieActionAfterPrompt(s)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, afterPromptFailures(m), "mutant %q survived", name)
	}
}

// ---- isSeriesSwitch ------------------------------------------------------------

var seriesSwitchCases = []struct {
	name  string
	media models.MediaType
	eps   int
	want  bool
}{
	{"anime with episodes", models.MediaTypeAnime, 12, true},
	{"anime with two", models.MediaTypeAnime, 2, true},
	{"anime with one is an OVA", models.MediaTypeAnime, 1, false},
	{"anime with none", models.MediaTypeAnime, 0, false},
	{"tv with episodes", models.MediaTypeTV, 8, true},
	{"tv with one", models.MediaTypeTV, 1, false},
	{"movie with many entries stays a movie", models.MediaTypeMovie, 5, false},
	{"movie with one", models.MediaTypeMovie, 1, false},
	{"unset type with episodes", "", 3, true},
}

func seriesSwitchFailures(impl func(*models.Anime, int) bool) []string {
	var f []string
	for _, tc := range seriesSwitchCases {
		if got := impl(&models.Anime{MediaType: tc.media}, tc.eps); got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %v want %v", tc.name, got, tc.want))
		}
	}
	return f
}

func TestIsSeriesSwitch(t *testing.T) {
	assert.Empty(t, seriesSwitchFailures(isSeriesSwitch))
}

func TestIsSeriesSwitchMeta(t *testing.T) {
	require.Empty(t, seriesSwitchFailures(isSeriesSwitch))
	mutants := map[string]func(*models.Anime, int) bool{
		"count only":      func(_ *models.Anime, n int) bool { return n > 1 },
		"type only":       func(a *models.Anime, _ int) bool { return !a.IsMovie() },
		"threshold at 1":  func(a *models.Anime, n int) bool { return !a.IsMovie() && n >= 1 },
		"threshold at 3":  func(a *models.Anime, n int) bool { return !a.IsMovie() && n > 2 },
		"tv is a movie":   func(a *models.Anime, n int) bool { return !a.IsMovieOrTV() && n > 1 },
		"movie is series": func(a *models.Anime, n int) bool { return a.IsMovie() && n > 1 },
		"always":          func(*models.Anime, int) bool { return true },
		"never":           func(*models.Anime, int) bool { return false },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, seriesSwitchFailures(m), "mutant %q survived", name)
	}
}

// ---- resolveAnimeChange --------------------------------------------------------

type changerMock struct {
	anime *models.Anime
	eps   []models.Episode
	err   error
	calls int
}

func (c *changerMock) fn() (*models.Anime, []models.Episode, error) {
	c.calls++
	return c.anime, c.eps, c.err
}

func withChanger(t *testing.T, c *changerMock) {
	t.Helper()
	prev := changeAnimeFn
	changeAnimeFn = c.fn
	t.Cleanup(func() { changeAnimeFn = prev })
}

func episodesN(n int) []models.Episode {
	out := make([]models.Episode, n)
	for i := range out {
		out[i] = models.Episode{Num: i + 1, Number: fmt.Sprint(i + 1), URL: fmt.Sprintf("u%d", i+1)}
	}
	return out
}

var resolveCases = []struct {
	name        string
	mock        changerMock
	want        changeOutcome
	wantAnime   bool
	wantEpCount int
}{
	{"continue row pends", changerMock{err: appflow.ErrHomeContinue}, changeBack, false, 0},
	{"continue row, wrapped", changerMock{err: fmt.Errorf("home: %w", appflow.ErrHomeContinue)}, changeBack, false, 0},
	{"search failed", changerMock{err: errors.New("no network")}, changeStay, false, 0},
	{"search cancelled", changerMock{err: player.ErrBackToAnimeSelection}, changeStay, false, 0},
	{"series", changerMock{anime: &models.Anime{Name: "S", MediaType: models.MediaTypeAnime}, eps: episodesN(12)}, changeToSeries, true, 12},
	{"movie", changerMock{anime: &models.Anime{Name: "M", MediaType: models.MediaTypeMovie}, eps: episodesN(1)}, changeToMovie, true, 1},
	{"ova", changerMock{anime: &models.Anime{Name: "O", MediaType: models.MediaTypeAnime}, eps: episodesN(1)}, changeToMovie, true, 1},
	{"movie listing several entries", changerMock{anime: &models.Anime{Name: "M", MediaType: models.MediaTypeMovie}, eps: episodesN(4)}, changeToMovie, true, 4},
	{"tv show", changerMock{anime: &models.Anime{Name: "T", MediaType: models.MediaTypeTV}, eps: episodesN(8)}, changeToSeries, true, 8},
}

func resolveFailures(impl func() animeChange) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	prev := changeAnimeFn
	defer func() { changeAnimeFn = prev }()
	for _, tc := range resolveCases {
		m := tc.mock
		changeAnimeFn = m.fn
		got := impl()
		if got.outcome != tc.want {
			f = append(f, fmt.Sprintf("%s: outcome %d want %d", tc.name, got.outcome, tc.want))
		}
		if (got.anime != nil) != tc.wantAnime || len(got.episodes) != tc.wantEpCount {
			f = append(f, fmt.Sprintf("%s: payload anime=%v eps=%d", tc.name, got.anime != nil, len(got.episodes)))
		}
		if tc.wantAnime && got.anime != tc.mock.anime {
			f = append(f, tc.name+": a different anime came back")
		}
		if m.calls != 1 {
			f = append(f, fmt.Sprintf("%s: asked %d times", tc.name, m.calls))
		}
	}
	return f
}

func TestResolveAnimeChange(t *testing.T) {
	assert.Empty(t, resolveFailures(resolveAnimeChange))
}

func TestResolveAnimeChangeMeta(t *testing.T) {
	require.Empty(t, resolveFailures(resolveAnimeChange))
	mutants := map[string]func() animeChange{
		"errors all stay": func() animeChange {
			a := resolveAnimeChange()
			if a.outcome == changeBack {
				a.outcome = changeStay
			}
			return a
		},
		"errors all go back": func() animeChange {
			a := resolveAnimeChange()
			if a.outcome == changeStay {
				a.outcome = changeBack
			}
			return a
		},
		"never series": func() animeChange {
			a := resolveAnimeChange()
			if a.outcome == changeToSeries {
				a.outcome = changeToMovie
			}
			return a
		},
		"always series": func() animeChange {
			a := resolveAnimeChange()
			if a.outcome == changeToMovie {
				a.outcome = changeToSeries
			}
			return a
		},
		"payload dropped": func() animeChange {
			a := resolveAnimeChange()
			a.anime, a.episodes = nil, nil
			return a
		},
		"episodes dropped": func() animeChange {
			a := resolveAnimeChange()
			a.episodes = nil
			return a
		},
		"asks twice": func() animeChange {
			resolveAnimeChange()
			return resolveAnimeChange()
		},
		"payload on error": func() animeChange {
			a := resolveAnimeChange()
			if a.outcome == changeBack || a.outcome == changeStay {
				a.anime = &models.Anime{}
			}
			return a
		},
		"swapped anime": func() animeChange {
			a := resolveAnimeChange()
			if a.anime != nil {
				a.anime = &models.Anime{Name: "other"}
			}
			return a
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, resolveFailures(m), "mutant %q survived", name)
	}
}

func TestResolveAnimeChangeDoesNotTouchStateOnFailure(t *testing.T) {
	c := &changerMock{err: errors.New("boom")}
	withChanger(t, c)
	got := resolveAnimeChange()
	assert.Equal(t, animeChange{outcome: changeStay}, got, "a failed change carries nothing, so the caller keeps its title")
}
