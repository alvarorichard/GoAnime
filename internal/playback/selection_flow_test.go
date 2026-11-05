package playback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pickerMock is a selector that records its calls and serves a canned answer.
type pickerMock struct {
	url, num string
	err      error
	calls    int
	got      [][]models.Episode
}

func (p *pickerMock) fn(eps []models.Episode) (string, string, error) {
	p.calls++
	p.got = append(p.got, eps)
	return p.url, p.num, p.err
}

type selection struct {
	url, num string
	idx      int
	err      error
}

// ---- SelectEpisodeWithFuzzy --------------------------------------------------

var fuzzyCases = []struct {
	name       string
	pick       pickerMock
	want       selection
	wantErrHas string
}{
	{"episode chosen", pickerMock{url: "u7", num: "Episode 7"}, selection{"u7", "Episode 7", 7, nil}, ""},
	{"back is a signal, not an error", pickerMock{err: player.ErrBackRequested}, selection{"", "back", -1, nil}, ""},
	{"wrapped back is still back", pickerMock{err: fmt.Errorf("x: %w", player.ErrBackRequested)}, selection{"", "back", -1, nil}, ""},
	{"selector failure is wrapped", pickerMock{err: errors.New("tty gone")}, selection{"", "", 0, nil}, "episode selection failed"},
	{"number that does not parse", pickerMock{url: "u", num: "Episode ?"}, selection{"", "", 0, nil}, "episode number"},
}

// fuzzyFailures judges an implementation of SelectEpisodeWithFuzzy against
// fuzzyCases. The implementation receives the selector it must use.
func fuzzyFailures(impl func([]models.Episode) (string, string, int, error)) []string {
	var f []string
	oldSel, oldExt := selectEpisodeFunc, extractEpisodeNumberFunc
	defer func() { selectEpisodeFunc, extractEpisodeNumberFunc = oldSel, oldExt }()
	extractEpisodeNumberFunc = func(s string) string {
		if s == "Episode 7" {
			return "7"
		}
		return "?"
	}
	for _, tc := range fuzzyCases {
		p := tc.pick
		selectEpisodeFunc = p.fn
		url, num, idx, err := impl(sampleEpisodes(2))
		if tc.wantErrHas != "" {
			if err == nil || !contains(err.Error(), tc.wantErrHas) {
				f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
			}
			if url != "" || num != "" || idx != 0 {
				f = append(f, fmt.Sprintf("%s: values next to an error: (%q,%q,%d)", tc.name, url, num, idx))
			}
			continue
		}
		if err != nil || url != tc.want.url || num != tc.want.num || idx != tc.want.idx {
			f = append(f, fmt.Sprintf("%s: got (%q,%q,%d,%v)", tc.name, url, num, idx, err))
		}
	}
	return f
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestSelectEpisodeWithFuzzy(t *testing.T) {
	withSelector(t, nil, nil)
	assert.Empty(t, fuzzyFailures(SelectEpisodeWithFuzzy))

	t.Run("passes the episode list to the selector", func(t *testing.T) {
		p := &pickerMock{url: "u", num: "Episode 7"}
		withSelectorInline(p.fn, func(string) string { return "7" })
		eps := sampleEpisodes(3)
		_, _, _, err := SelectEpisodeWithFuzzy(eps)
		require.NoError(t, err)
		require.Equal(t, 1, p.calls)
		assert.Equal(t, eps, p.got[0])
	})

	t.Run("failure errors keep the cause", func(t *testing.T) {
		boom := errors.New("tty gone")
		withSelectorInline((&pickerMock{err: boom}).fn, nil)
		_, _, _, err := SelectEpisodeWithFuzzy(nil)
		assert.ErrorIs(t, err, boom)
	})
}

// withSelectorInline swaps the injections inside a test that already holds
// selectorOverrideMu through withSelector.
func withSelectorInline(sel selectEpisodeFuncType, ext extractEpisodeNumberFuncType) {
	if sel != nil {
		selectEpisodeFunc = sel
	}
	if ext != nil {
		extractEpisodeNumberFunc = ext
	}
}

func TestSelectEpisodeWithFuzzyMeta(t *testing.T) {
	withSelector(t, nil, nil)
	require.Empty(t, fuzzyFailures(SelectEpisodeWithFuzzy))

	mutants := map[string]func([]models.Episode) (string, string, int, error){
		"back becomes an error": func(e []models.Episode) (string, string, int, error) {
			u, n, i, err := SelectEpisodeWithFuzzy(e)
			if n == "back" {
				return u, n, i, player.ErrBackRequested
			}
			return u, n, i, err
		},
		"back index zero": func(e []models.Episode) (string, string, int, error) {
			u, n, _, err := SelectEpisodeWithFuzzy(e)
			if n == "back" {
				return u, n, 0, err
			}
			return u, n, 0, err
		},
		"selector error swallowed": func(e []models.Episode) (string, string, int, error) {
			u, n, i, err := SelectEpisodeWithFuzzy(e)
			if err != nil {
				return "", "", 0, nil
			}
			return u, n, i, err
		},
		"parse error swallowed": func(e []models.Episode) (string, string, int, error) {
			u, n, i, err := SelectEpisodeWithFuzzy(e)
			if err != nil && contains(err.Error(), "episode number") {
				return u, n, 0, nil
			}
			return u, n, i, err
		},
		"url dropped": func(e []models.Episode) (string, string, int, error) {
			_, n, i, err := SelectEpisodeWithFuzzy(e)
			return "", n, i, err
		},
		"index off by one": func(e []models.Episode) (string, string, int, error) {
			u, n, i, err := SelectEpisodeWithFuzzy(e)
			if err == nil && n != "back" {
				i++
			}
			return u, n, i, err
		},
		"never back": func(e []models.Episode) (string, string, int, error) {
			u, n, i, err := SelectEpisodeWithFuzzy(e)
			if n == "back" {
				return "", "", 0, errors.New("episode selection failed")
			}
			return u, n, i, err
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, fuzzyFailures(m), "mutant %q survived", name)
	}
}

// ---- FindEpisodeByNumber ------------------------------------------------------

func TestFindEpisodeByNumberHit(t *testing.T) {
	withSelector(t, nil, nil)
	p := &pickerMock{url: "SHOULD-NOT-BE-USED", num: "Episode 1"}
	withSelectorInline(p.fn, nil)

	url, num, idx, err := FindEpisodeByNumber(sampleEpisodes(5), 4)
	require.NoError(t, err)
	assert.Equal(t, "https://example.test/ep4", url)
	assert.Equal(t, "Episode 4", num)
	assert.Equal(t, 4, idx)
	assert.Zero(t, p.calls, "a hit never opens the picker")
}

func TestFindEpisodeByNumberMissFallsBackToThePicker(t *testing.T) {
	withSelector(t, nil, nil)
	p := &pickerMock{url: "picked", num: "Episode 2"}
	withSelectorInline(p.fn, nil)
	eps := sampleEpisodes(3)

	url, num, idx, err := FindEpisodeByNumber(eps, 99)
	require.NoError(t, err)
	assert.Equal(t, "picked", url)
	assert.Equal(t, "Episode 2", num)
	assert.Equal(t, 2, idx)
	require.Equal(t, 1, p.calls, "re-selecting asks exactly once")
	assert.Equal(t, eps, p.got[0], "the same list is offered")
}

func TestFindEpisodeByNumberMissPropagatesThePickersAnswers(t *testing.T) {
	withSelector(t, nil, nil)
	withSelectorInline((&pickerMock{err: player.ErrBackRequested}).fn, nil)
	url, num, idx, err := FindEpisodeByNumber(sampleEpisodes(2), 9)
	require.NoError(t, err)
	assert.Equal(t, selection{"", "back", -1, nil}, selection{url, num, idx, err})

	boom := errors.New("tty gone")
	withSelectorInline((&pickerMock{err: boom}).fn, nil)
	_, _, _, err = FindEpisodeByNumber(sampleEpisodes(2), 9)
	assert.ErrorIs(t, err, boom)
}

func TestFindEpisodeByNumberSkipsUnrelatedLabelsAndTakesTheFirstMatch(t *testing.T) {
	withSelector(t, nil, nil)
	withSelectorInline((&pickerMock{err: errors.New("unused")}).fn, nil)
	eps := []models.Episode{
		{URL: "x", Number: "OVA"}, // the extractor reads an unnumbered label as 1
		{URL: "first", Number: "3"},
		{URL: "second", Number: "3"},
	}
	url, _, idx, err := FindEpisodeByNumber(eps, 3)
	require.NoError(t, err)
	assert.Equal(t, "first", url, "first match wins")
	assert.Equal(t, 3, idx)
}

func findFailures(impl func([]models.Episode, int) (string, string, int, error)) []string {
	var f []string
	oldSel, oldExt := selectEpisodeFunc, extractEpisodeNumberFunc
	defer func() { selectEpisodeFunc, extractEpisodeNumberFunc = oldSel, oldExt }()
	p := &pickerMock{url: "picked", num: "Episode 1"}
	selectEpisodeFunc = p.fn
	extractEpisodeNumberFunc = player.ExtractEpisodeNumber

	eps := sampleEpisodes(5)
	for n := 1; n <= 5; n++ {
		url, num, idx, err := impl(eps, n)
		if err != nil || url != fmt.Sprintf("https://example.test/ep%d", n) || num != fmt.Sprintf("Episode %d", n) || idx != n {
			f = append(f, fmt.Sprintf("episode %d: got (%q,%q,%d,%v)", n, url, num, idx, err))
		}
	}
	if p.calls != 0 {
		f = append(f, "picker opened on hits")
	}
	dup := append(append([]models.Episode(nil), eps...),
		models.Episode{URL: "dup-a", Number: "Episode 9"}, models.Episode{URL: "dup-b", Number: "Episode 9"})
	if url, _, _, err := impl(dup, 9); err != nil || url != "dup-a" {
		f = append(f, fmt.Sprintf("duplicate numbers: got %q,%v want the first", url, err))
	}
	url, _, idx, err := impl(eps, 0)
	if err != nil || url != "picked" || idx != 1 || p.calls != 1 {
		f = append(f, fmt.Sprintf("miss(0): got (%q,%d,%v) calls=%d", url, idx, err, p.calls))
	}
	url, _, _, _ = impl(eps, 6)
	if url != "picked" || p.calls != 2 {
		f = append(f, "miss(6) did not fall back")
	}
	return f
}

func TestFindEpisodeByNumberMeta(t *testing.T) {
	withSelector(t, nil, nil)
	require.Empty(t, findFailures(FindEpisodeByNumber))

	mutants := map[string]func([]models.Episode, int) (string, string, int, error){
		"zero-based index": func(e []models.Episode, n int) (string, string, int, error) {
			u, s, _, err := FindEpisodeByNumber(e, n)
			return u, s, n - 1, err
		},
		"off by one lookup": func(e []models.Episode, n int) (string, string, int, error) {
			return FindEpisodeByNumber(e, n+1)
		},
		"miss returns error instead of picker": func(e []models.Episode, n int) (string, string, int, error) {
			if n < 1 || n > len(e) {
				return "", "", 0, errors.New("not found")
			}
			return FindEpisodeByNumber(e, n)
		},
		"always picker": func(e []models.Episode, _ int) (string, string, int, error) {
			return SelectEpisodeWithFuzzy(e)
		},
		"number dropped": func(e []models.Episode, n int) (string, string, int, error) {
			u, _, i, err := FindEpisodeByNumber(e, n)
			return u, "", i, err
		},
		"last match wins": func(e []models.Episode, n int) (string, string, int, error) {
			rev := make([]models.Episode, len(e))
			for i := range e {
				rev[len(e)-1-i] = e[i]
			}
			return FindEpisodeByNumber(rev, n)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, findFailures(m), "mutant %q survived", name)
	}
}

// ---- SelectInitialEpisode: the CONTINUE preselect --------------------------------

func TestSelectInitialEpisodePreselect(t *testing.T) {
	t.Run("a present episode opens directly and the preselect is consumed", func(t *testing.T) {
		p := &pickerMock{url: "picker", num: "Episode 1"}
		withSelector(t, p.fn, nil)
		PreselectEpisodeNum = 3
		t.Cleanup(func() { PreselectEpisodeNum = 0 })

		url, num, idx, err := SelectInitialEpisode(sampleEpisodes(5))
		require.NoError(t, err)
		assert.Equal(t, selection{"https://example.test/ep3", "Episode 3", 3, nil}, selection{url, num, idx, err})
		assert.Zero(t, p.calls, "no picker")
		assert.Zero(t, PreselectEpisodeNum, "one-shot: the next title asks again")
	})

	t.Run("episode 1 can be preselected", func(t *testing.T) {
		p := &pickerMock{url: "picker", num: "Episode 9"}
		withSelector(t, p.fn, nil)
		PreselectEpisodeNum = 1
		t.Cleanup(func() { PreselectEpisodeNum = 0 })

		url, _, idx, err := SelectInitialEpisode(sampleEpisodes(3))
		require.NoError(t, err)
		assert.Equal(t, "https://example.test/ep1", url)
		assert.Equal(t, 1, idx)
		assert.Zero(t, p.calls)
	})

	t.Run("a missing episode falls back to the picker", func(t *testing.T) {
		p := &pickerMock{url: "picked", num: "Episode 2"}
		withSelector(t, p.fn, nil)
		PreselectEpisodeNum = 40
		t.Cleanup(func() { PreselectEpisodeNum = 0 })

		url, _, idx, err := SelectInitialEpisode(sampleEpisodes(5))
		require.NoError(t, err)
		assert.Equal(t, "picked", url)
		assert.Equal(t, 2, idx)
		assert.Equal(t, 1, p.calls, "exactly one prompt (FindEpisodeByNumber's re-selection, not a second)")
		assert.Zero(t, PreselectEpisodeNum)
	})

	t.Run("zero and negative do not preselect", func(t *testing.T) {
		withSelector(t, nil, nil) // takes the lock once; the loop swaps inline
		for _, v := range []int{0, -3} {
			p := &pickerMock{url: "picked", num: "Episode 2"}
			withSelectorInline(p.fn, nil)
			PreselectEpisodeNum = v
			_, _, _, err := SelectInitialEpisode(sampleEpisodes(5))
			require.NoError(t, err)
			assert.Equal(t, 1, p.calls, "preselect %d asks", v)
			assert.Equal(t, v, PreselectEpisodeNum, "an unused value is left alone")
			PreselectEpisodeNum = 0
		}
	})
}

// ---- reportMovieStreamFailed ----------------------------------------------------

func TestReportMovieStreamFailedHandsTheNoticeToTheResultsScreen(t *testing.T) {
	var got []string
	prev := noticeOnNextResults
	noticeOnNextResults = func(s string) { got = append(got, s) }
	t.Cleanup(func() { noticeOnNextResults = prev })

	anime := &models.Anime{Name: "Movie", Source: "Pobreflix"}
	err := errors.New("stream resolution failed")
	reportMovieStreamFailed(anime, err)

	require.Len(t, got, 1, "one notice per failure")
	assert.Equal(t, movieStreamFailedNotice(anime, err), got[0])
	assert.NotEmpty(t, got[0])

	reportMovieStreamFailed(anime, nil)
	require.Len(t, got, 2)
	assert.Empty(t, got[1], "no error clears the notice instead of showing a stale one")
}

func noticeFailures(report func(*models.Anime, error)) []string {
	var f []string
	var got []string
	prev := noticeOnNextResults
	noticeOnNextResults = func(s string) { got = append(got, s) }
	defer func() { noticeOnNextResults = prev }()

	a := &models.Anime{Name: "M", Source: "Pobreflix"}
	e := errors.New("boom")
	report(a, e)
	report(a, nil)
	report(nil, e)
	if len(got) != 3 {
		return append(f, fmt.Sprintf("%d notices, want 3", len(got)))
	}
	if got[0] != movieStreamFailedNotice(a, e) || got[0] == "" {
		f = append(f, "failure notice wrong")
	}
	if got[1] != "" {
		f = append(f, "nil error should clear")
	}
	if got[2] != movieStreamFailedNotice(nil, e) || got[2] == "" {
		f = append(f, "nil anime should still explain")
	}
	return f
}

func TestReportMovieStreamFailedMeta(t *testing.T) {
	require.Empty(t, noticeFailures(reportMovieStreamFailed))
	mutants := map[string]func(*models.Anime, error){
		"reports nothing": func(*models.Anime, error) {},
		"raw error text":  func(_ *models.Anime, e error) { noticeOnNextResults(fmt.Sprint(e)) },
		"always the same": func(*models.Anime, error) { noticeOnNextResults("failed") },
		"stale on nil error": func(a *models.Anime, e error) {
			if e == nil {
				return
			}
			reportMovieStreamFailed(a, e)
		},
		"drops the sources hint": func(_ *models.Anime, e error) { noticeOnNextResults(movieStreamFailedNotice(nil, e)) },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, noticeFailures(m), "mutant %q survived", name)
	}
}

// ---- handleUserNavigationEnhanced -------------------------------------------------

func TestHandleUserNavigationEnhancedMatchesTheBaseNavigation(t *testing.T) {
	withSelector(t, nil, nil)
	withSelectorInline((&pickerMock{url: "picked", num: "Episode 2"}).fn, nil)
	eps := sampleEpisodes(4)
	for _, in := range []string{"n", "p", "e", "", "x"} {
		for _, cur := range []int{1, 2, 4} {
			u1, n1, i1 := handleUserNavigation(in, eps, cur, 4)
			u2, n2, i2 := handleUserNavigationEnhanced(in, eps, cur, 4, &models.Anime{Name: "ignored"})
			assert.Equal(t, selection{u1, n1, i1, nil}, selection{u2, n2, i2, nil}, "input %q at %d", in, cur)
			u3, n3, i3 := handleUserNavigationEnhanced(in, eps, cur, 4, nil)
			assert.Equal(t, selection{u1, n1, i1, nil}, selection{u3, n3, i3, nil}, "nil anime, input %q at %d", in, cur)
		}
	}
}

func TestHandleUserNavigationFailureKeepsTheCurrentEpisode(t *testing.T) {
	withSelector(t, nil, nil)
	withSelectorInline((&pickerMock{err: errors.New("tty gone")}).fn, nil)
	url, num, idx := handleUserNavigation("e", sampleEpisodes(3), 2, 3)
	assert.Empty(t, url)
	assert.Empty(t, num)
	assert.Equal(t, 2, idx, "on error the player stays where it was")
}

// Next and previous step by exactly one episode, clamped to [1, total].
func TestHandleUserNavigationStepsByOne(t *testing.T) {
	withSelector(t, nil, nil)
	eps := sampleEpisodes(5)
	for _, tc := range []struct {
		in       string
		cur      int
		wantIdx  int
		wantName string
	}{
		{"n", 1, 2, "Episode 2"}, {"n", 3, 4, "Episode 4"}, {"n", 5, 5, "Episode 5"},
		{"p", 5, 4, "Episode 4"}, {"p", 3, 2, "Episode 2"}, {"p", 2, 1, "Episode 1"}, {"p", 1, 1, "Episode 1"},
		{"p", 0, 1, "Episode 1"}, {"n", 9, 5, "Episode 5"},
	} {
		_, num, idx := handleUserNavigation(tc.in, eps, tc.cur, 5)
		assert.Equal(t, tc.wantIdx, idx, "%s from %d", tc.in, tc.cur)
		assert.Equal(t, tc.wantName, num, "%s from %d", tc.in, tc.cur)
	}
}
