package playback

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- seriesActionAfterPlayback -------------------------------------------------

var seriesAfterPlaybackCases = []struct {
	name string
	err  error
	want seriesAction
}{
	{"finished normally", nil, seriesAskUser},
	{"user quit", player.ErrUserQuit, seriesQuit},
	{"user quit, wrapped", fmt.Errorf("x: %w", player.ErrUserQuit), seriesQuit},
	{"back from server selection", player.ErrBackToEpisodeSelection, seriesReselectFromServer},
	{"back from server selection, wrapped", fmt.Errorf("x: %w", player.ErrBackToEpisodeSelection), seriesReselectFromServer},
	{"change title", player.ErrChangeAnime, seriesChange},
	{"change title, wrapped", fmt.Errorf("x: %w", player.ErrChangeAnime), seriesChange},
	{"playback error asks", errors.New("mpv crashed"), seriesAskUser},
	{"back to anime selection is not handled here", player.ErrBackToAnimeSelection, seriesAskUser},
}

func seriesAfterPlaybackFailures(impl func(error) seriesAction) []string {
	var f []string
	for _, tc := range seriesAfterPlaybackCases {
		if got := impl(tc.err); got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d want %d", tc.name, got, tc.want))
		}
	}
	return f
}

func TestSeriesActionAfterPlayback(t *testing.T) {
	assert.Empty(t, seriesAfterPlaybackFailures(seriesActionAfterPlayback))
}

func TestSeriesActionAfterPlaybackMeta(t *testing.T) {
	require.Empty(t, seriesAfterPlaybackFailures(seriesActionAfterPlayback))
	mutants := map[string]func(error) seriesAction{
		"always ask": func(error) seriesAction { return seriesAskUser },
		"quit and change swapped": func(e error) seriesAction {
			switch seriesActionAfterPlayback(e) {
			case seriesQuit:
				return seriesChange
			case seriesChange:
				return seriesQuit
			}
			return seriesActionAfterPlayback(e)
		},
		"server back becomes change": func(e error) seriesAction {
			if errors.Is(e, player.ErrBackToEpisodeSelection) {
				return seriesChange
			}
			return seriesActionAfterPlayback(e)
		},
		"equality, not errors.Is": func(e error) seriesAction {
			switch e {
			case player.ErrUserQuit:
				return seriesQuit
			case player.ErrBackToEpisodeSelection:
				return seriesReselectFromServer
			case player.ErrChangeAnime:
				return seriesChange
			}
			return seriesAskUser
		},
		"any error quits": func(e error) seriesAction {
			if e != nil {
				return seriesQuit
			}
			return seriesAskUser
		},
		"server back ignored": func(e error) seriesAction {
			if errors.Is(e, player.ErrBackToEpisodeSelection) {
				return seriesAskUser
			}
			return seriesActionAfterPlayback(e)
		},
		"quit ignored": func(e error) seriesAction {
			if errors.Is(e, player.ErrUserQuit) {
				return seriesAskUser
			}
			return seriesActionAfterPlayback(e)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, seriesAfterPlaybackFailures(m), "mutant %q survived", name)
	}
}

// ---- seriesActionAfterPrompt ---------------------------------------------------

var seriesPromptCases = []struct {
	input string
	want  seriesAction
}{
	{"q", seriesQuit},
	{"quit", seriesQuit},
	{"c", seriesChange},
	{"back", seriesChange},
	{"e", seriesPickEpisode},
	{"n", seriesNavigate},
	{"p", seriesNavigate},
	{"", seriesNavigate},
	{"x", seriesNavigate},
	{"Q", seriesNavigate},
	{"E", seriesNavigate},
}

func seriesPromptFailures(impl func(string) seriesAction) []string {
	var f []string
	for _, tc := range seriesPromptCases {
		if got := impl(tc.input); got != tc.want {
			f = append(f, fmt.Sprintf("%q: got %d want %d", tc.input, got, tc.want))
		}
	}
	return f
}

func TestSeriesActionAfterPrompt(t *testing.T) {
	assert.Empty(t, seriesPromptFailures(seriesActionAfterPrompt))
}

// Every answer the series menu can give is understood as its label promises.
func TestSeriesMenuAnswersMatchTheLabels(t *testing.T) {
	prev := findMenuFunc
	t.Cleanup(func() { findMenuFunc = prev })
	var labels []string
	findMenuFunc = func(items []menuItem) (int, error) {
		for _, it := range items {
			labels = append(labels, it.Label)
		}
		return -1, nil
	}
	GetUserInput()
	require.Equal(t, []string{"Next episode", "Previous episode", "Select episode", "Change anime", "← Back", "Quit"}, labels)

	want := map[string]seriesAction{
		"Next episode": seriesNavigate, "Previous episode": seriesNavigate, "Select episode": seriesPickEpisode,
		"Change anime": seriesChange, "← Back": seriesChange, "Quit": seriesQuit,
	}
	for i, label := range labels {
		findMenuFunc = func([]menuItem) (int, error) { return i, nil }
		assert.Equal(t, want[label], seriesActionAfterPrompt(GetUserInput()), label)
	}
}

func TestSeriesActionAfterPromptMeta(t *testing.T) {
	require.Empty(t, seriesPromptFailures(seriesActionAfterPrompt))
	mutants := map[string]func(string) seriesAction{
		"always navigate": func(string) seriesAction { return seriesNavigate },
		"quit word dropped": func(s string) seriesAction {
			if s == "quit" {
				return seriesNavigate
			}
			return seriesActionAfterPrompt(s)
		},
		"back navigates": func(s string) seriesAction {
			if s == "back" {
				return seriesNavigate
			}
			return seriesActionAfterPrompt(s)
		},
		"c quits": func(s string) seriesAction {
			if s == "c" {
				return seriesQuit
			}
			return seriesActionAfterPrompt(s)
		},
		"e navigates": func(s string) seriesAction {
			if s == "e" {
				return seriesNavigate
			}
			return seriesActionAfterPrompt(s)
		},
		"case-insensitive": func(s string) seriesAction {
			switch s {
			case "Q":
				return seriesQuit
			case "E":
				return seriesPickEpisode
			}
			return seriesActionAfterPrompt(s)
		},
		"unknown asks again": func(s string) seriesAction {
			if a := seriesActionAfterPrompt(s); a == seriesNavigate && s != "n" && s != "p" {
				return seriesAskUser
			}
			return seriesActionAfterPrompt(s)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, seriesPromptFailures(m), "mutant %q survived", name)
	}
}

// ---- navigationResult ----------------------------------------------------------

func TestNavigationResult(t *testing.T) {
	cur := episodeSel{url: "cur", label: "Episode 2", num: 2}

	got, moved := navigationResult(cur, "next", "Episode 3", 3)
	assert.True(t, moved)
	assert.Equal(t, episodeSel{url: "next", label: "Episode 3", num: 3}, got)

	got, moved = navigationResult(cur, "", "Episode 3", 3)
	assert.False(t, moved, "an empty URL is a failed navigation")
	assert.Equal(t, cur, got, "and keeps the episode that was playing")

	got, moved = navigationResult(cur, "next", "", 0)
	assert.True(t, moved, "only the URL decides: a label-less answer still moves")
	assert.Equal(t, episodeSel{url: "next"}, got)
}

func navigationFailures(impl func(episodeSel, string, string, int) (episodeSel, bool)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	cur := episodeSel{url: "cur", label: "L", num: 2}
	if got, ok := impl(cur, "u", "Episode 3", 3); !ok || got != (episodeSel{"u", "Episode 3", 3}) {
		f = append(f, fmt.Sprintf("move: %v %v", got, ok))
	}
	if got, ok := impl(cur, "", "Episode 3", 3); ok || got != cur {
		f = append(f, fmt.Sprintf("failed navigation: %v %v", got, ok))
	}
	return f
}

func TestNavigationResultMeta(t *testing.T) {
	require.Empty(t, navigationFailures(navigationResult))
	mutants := map[string]func(episodeSel, string, string, int) (episodeSel, bool){
		"always moves": func(c episodeSel, u, l string, n int) (episodeSel, bool) {
			return episodeSel{u, l, n}, true
		},
		"never moves": func(c episodeSel, _, _ string, _ int) (episodeSel, bool) { return c, false },
		"zeroes the selection on failure": func(c episodeSel, u, l string, n int) (episodeSel, bool) {
			if u == "" {
				return episodeSel{}, false
			}
			return episodeSel{u, l, n}, true
		},
		"label and url swapped": func(c episodeSel, u, l string, n int) (episodeSel, bool) {
			if u == "" {
				return c, false
			}
			return episodeSel{l, u, n}, true
		},
		"number dropped": func(c episodeSel, u, l string, _ int) (episodeSel, bool) {
			if u == "" {
				return c, false
			}
			return episodeSel{u, l, 0}, true
		},
		"label decides": func(c episodeSel, u, l string, n int) (episodeSel, bool) {
			if l == "" {
				return c, false
			}
			return episodeSel{u, l, n}, true
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, navigationFailures(m), "mutant %q survived", name)
	}
}

// ---- pickEpisode / selectFirstEpisode / reselectEpisode -------------------------

// seqPicker answers one canned reply per call and then repeats the last one.
type seqPicker struct {
	answers []pickerMock
	calls   int
}

func (s *seqPicker) fn(eps []models.Episode) (string, string, error) {
	i := min(s.calls, len(s.answers)-1)
	s.calls++
	a := s.answers[i]
	return a.url, a.num, a.err
}

func okPick(n int) pickerMock {
	return pickerMock{url: fmt.Sprintf("https://example.test/ep%d", n), num: fmt.Sprintf("Episode %d", n)}
}

func TestPickEpisode(t *testing.T) {
	withSelector(t, nil, nil)
	withSelectorInline((&seqPicker{answers: []pickerMock{okPick(4)}}).fn, nil)
	got, err := pickEpisode(sampleEpisodes(5))
	require.NoError(t, err)
	assert.Equal(t, episodeSel{url: "https://example.test/ep4", label: "Episode 4", num: 4}, got)

	boom := errors.New("tty gone")
	withSelectorInline((&seqPicker{answers: []pickerMock{{err: boom}}}).fn, nil)
	got, err = pickEpisode(sampleEpisodes(5))
	assert.ErrorIs(t, err, boom)
	assert.Equal(t, episodeSel{}, got, "no half-filled selection next to an error")
}

type firstEpisodeCase struct {
	name      string
	answers   []pickerMock
	wantCalls int
	wantSel   episodeSel
	wantErr   error
	errHas    string
}

func firstEpisodeCases() []firstEpisodeCase {
	boom := errors.New("tty gone")
	fail := pickerMock{err: boom}
	back := pickerMock{err: player.ErrBackRequested}
	return []firstEpisodeCase{
		{"first try", []pickerMock{okPick(2)}, 1, episodeSel{"https://example.test/ep2", "Episode 2", 2}, nil, ""},
		{"second try", []pickerMock{fail, okPick(3)}, 2, episodeSel{"https://example.test/ep3", "Episode 3", 3}, nil, ""},
		{"third try", []pickerMock{fail, fail, okPick(1)}, 3, episodeSel{"https://example.test/ep1", "Episode 1", 1}, nil, ""},
		{"gives up after three", []pickerMock{fail}, 3, episodeSel{}, boom, "after 3 attempts"},
		{"back at once", []pickerMock{back}, 1, episodeSel{}, player.ErrBackToAnimeSelection, ""},
		{"back after a failure", []pickerMock{fail, back}, 2, episodeSel{}, player.ErrBackToAnimeSelection, ""},
	}
}

func firstEpisodeFailures(impl func([]models.Episode) (episodeSel, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	old := selectEpisodeFunc
	defer func() { selectEpisodeFunc = old }()
	for _, tc := range firstEpisodeCases() {
		sp := &seqPicker{answers: tc.answers}
		selectEpisodeFunc = sp.fn
		got, err := impl(sampleEpisodes(5))
		if got != tc.wantSel {
			f = append(f, fmt.Sprintf("%s: selection %v", tc.name, got))
		}
		if sp.calls != tc.wantCalls {
			f = append(f, fmt.Sprintf("%s: %d calls want %d", tc.name, sp.calls, tc.wantCalls))
		}
		switch {
		case tc.wantErr == nil && err != nil:
			f = append(f, fmt.Sprintf("%s: unexpected error %v", tc.name, err))
		case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
			f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
		case tc.errHas != "" && (err == nil || !strings.Contains(err.Error(), tc.errHas)):
			f = append(f, fmt.Sprintf("%s: message %v", tc.name, err))
		}
	}
	return f
}

func TestSelectFirstEpisode(t *testing.T) {
	withSelector(t, nil, nil)
	assert.Empty(t, firstEpisodeFailures(selectFirstEpisode))
}

func TestSelectFirstEpisodeMeta(t *testing.T) {
	withSelector(t, nil, nil)
	require.Empty(t, firstEpisodeFailures(selectFirstEpisode))
	mutants := map[string]func([]models.Episode) (episodeSel, error){
		"no retry": pickEpisode,
		"retries forever on back": func(e []models.Episode) (episodeSel, error) {
			for range maxEpisodeRetries {
				if sel, err := pickEpisode(e); err == nil {
					return sel, nil
				}
			}
			return episodeSel{}, player.ErrBackToAnimeSelection
		},
		"back is an error": func(e []models.Episode) (episodeSel, error) {
			sel, err := selectFirstEpisode(e)
			if errors.Is(err, player.ErrBackToAnimeSelection) {
				return sel, errors.New("episode selection failed")
			}
			return sel, err
		},
		"swallows the final failure": func(e []models.Episode) (episodeSel, error) {
			sel, err := selectFirstEpisode(e)
			if err != nil && !errors.Is(err, player.ErrBackToAnimeSelection) {
				return sel, nil
			}
			return sel, err
		},
		"cause lost": func(e []models.Episode) (episodeSel, error) {
			sel, err := selectFirstEpisode(e)
			if err != nil && !errors.Is(err, player.ErrBackToAnimeSelection) {
				return sel, errors.New("episode selection failed after 3 attempts")
			}
			return sel, err
		},
		"two attempts only": func(e []models.Episode) (episodeSel, error) {
			var last error
			for range 2 {
				sel, err := pickEpisode(e)
				if err == nil {
					return sel, nil
				}
				if errors.Is(err, player.ErrBackRequested) {
					return episodeSel{}, player.ErrBackToAnimeSelection
				}
				last = err
			}
			return episodeSel{}, fmt.Errorf("episode selection failed after 3 attempts: %w", last)
		},
		"returns the failed selection": func(e []models.Episode) (episodeSel, error) {
			sel, err := selectFirstEpisode(e)
			if err != nil {
				sel = episodeSel{url: "stale"}
			}
			return sel, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, firstEpisodeFailures(m), "mutant %q survived", name)
	}
}

func reselectFailures(impl func([]models.Episode, episodeSel, bool) (episodeSel, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	old := selectEpisodeFunc
	defer func() { selectEpisodeFunc = old }()
	cur := episodeSel{url: "cur", label: "Episode 2", num: 2}
	boom := errors.New("tty gone")
	cases := []struct {
		name       string
		answer     pickerMock
		backLeaves bool
		wantSel    episodeSel
		wantErr    error
	}{
		{"new pick", okPick(5), false, episodeSel{"https://example.test/ep5", "Episode 5", 5}, nil},
		{"new pick, back leaves", okPick(5), true, episodeSel{"https://example.test/ep5", "Episode 5", 5}, nil},
		{"back keeps the episode", pickerMock{err: player.ErrBackRequested}, false, cur, nil},
		{"back leaves for anime selection", pickerMock{err: player.ErrBackRequested}, true, cur, player.ErrBackToAnimeSelection},
		{"failure keeps the episode", pickerMock{err: boom}, false, cur, nil},
		{"failure keeps it even when back leaves", pickerMock{err: boom}, true, cur, nil},
	}
	for _, tc := range cases {
		selectEpisodeFunc = (&seqPicker{answers: []pickerMock{tc.answer}}).fn
		got, err := impl(sampleEpisodes(5), cur, tc.backLeaves)
		if got != tc.wantSel {
			f = append(f, fmt.Sprintf("%s: selection %v", tc.name, got))
		}
		if !errors.Is(err, tc.wantErr) || (tc.wantErr == nil && err != nil) {
			f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
		}
	}
	return f
}

func TestReselectEpisode(t *testing.T) {
	withSelector(t, nil, nil)
	assert.Empty(t, reselectFailures(reselectEpisode))
}

func TestReselectEpisodeMeta(t *testing.T) {
	withSelector(t, nil, nil)
	require.Empty(t, reselectFailures(reselectEpisode))
	type fn = func([]models.Episode, episodeSel, bool) (episodeSel, error)
	mutants := map[string]fn{
		"back always leaves": func(e []models.Episode, c episodeSel, _ bool) (episodeSel, error) {
			return reselectEpisode(e, c, true)
		},
		"back never leaves": func(e []models.Episode, c episodeSel, _ bool) (episodeSel, error) {
			return reselectEpisode(e, c, false)
		},
		"failure leaves": func(e []models.Episode, c episodeSel, b bool) (episodeSel, error) {
			next, err := pickEpisode(e)
			if err != nil {
				return c, player.ErrBackToAnimeSelection
			}
			return next, nil
		},
		"failure commits zeros": func(e []models.Episode, c episodeSel, b bool) (episodeSel, error) {
			next, err := pickEpisode(e)
			if err != nil {
				return episodeSel{}, nil //nolint:nilerr // deliberate mutant: commits zeros on failure
			}
			return next, nil
		},
		"success ignored": func(e []models.Episode, c episodeSel, b bool) (episodeSel, error) {
			reselectEpisode(e, c, b)
			return c, nil
		},
		"back returns the picker's zeros": func(e []models.Episode, c episodeSel, b bool) (episodeSel, error) {
			sel, err := reselectEpisode(e, c, b)
			if err != nil {
				sel = episodeSel{}
			}
			return sel, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, reselectFailures(m), "mutant %q survived", name)
	}
}
