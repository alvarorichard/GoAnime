package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- resumeClock / resumeTitle ----------------------------------------------------

var clockCases = map[int]string{0: "0s", 1: "1s", 59: "59s", 60: "1m 0s", 61: "1m 1s", 125: "2m 5s", 3600: "60m 0s", 3725: "62m 5s"}

func clockFailures(impl func(int) string) (f []string) {
	for in, want := range clockCases {
		if got := impl(in); got != want {
			f = append(f, fmt.Sprintf("%d: got %q want %q", in, got, want))
		}
	}
	return f
}

func TestResumeClock(t *testing.T) {
	assert.Empty(t, clockFailures(resumeClock))
}

func TestResumeClockMeta(t *testing.T) {
	require.Empty(t, clockFailures(resumeClock))
	mutants := map[string]func(int) string{
		"always seconds": func(s int) string { return fmt.Sprintf("%ds", s) },
		"minutes at 61": func(s int) string {
			if s <= 60 {
				return fmt.Sprintf("%ds", s)
			}
			return resumeClock(s)
		},
		"drops the remainder": func(s int) string {
			if m := s / 60; m > 0 {
				return fmt.Sprintf("%dm 0s", m)
			}
			return fmt.Sprintf("%ds", s)
		},
		"hours": func(s int) string {
			if s >= 3600 {
				return fmt.Sprintf("%dh %dm %ds", s/3600, s%3600/60, s%60)
			}
			return resumeClock(s)
		},
		"swapped": func(s int) string {
			if m := s / 60; m > 0 {
				return fmt.Sprintf("%ds %dm", s%60, m)
			}
			return fmt.Sprintf("%ds", s)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, clockFailures(m), "mutant %q survived", name)
	}
}

func TestResumeTitle(t *testing.T) {
	assert.Equal(t, "Resume episode 4 from 2m 5s?", resumeTitle(4, 125, false))
	assert.Equal(t, "Resume movie from 42s?", resumeTitle(1, 42, true), "a movie has no episode number")
	assert.Equal(t, "Resume episode 12 from 0s?", resumeTitle(12, 0, false))
	assert.NotContains(t, resumeTitle(7, 60, true), "7")
}

func titleFailures(impl func(int, int, bool) string) (f []string) {
	for _, tc := range []struct {
		ep, sec int
		movie   bool
		want    string
	}{
		{4, 125, false, "Resume episode 4 from 2m 5s?"},
		{1, 42, true, "Resume movie from 42s?"},
		{12, 3, false, "Resume episode 12 from 3s?"},
		{9, 600, true, "Resume movie from 10m 0s?"},
	} {
		if got := impl(tc.ep, tc.sec, tc.movie); got != tc.want {
			f = append(f, fmt.Sprintf("%d/%d/%v: %q", tc.ep, tc.sec, tc.movie, got))
		}
	}
	return f
}

func TestResumeTitleMeta(t *testing.T) {
	require.Empty(t, titleFailures(resumeTitle))
	mutants := map[string]func(int, int, bool) string{
		"movie wording always":   func(e, s int, _ bool) string { return resumeTitle(e, s, true) },
		"episode wording always": func(e, s int, _ bool) string { return resumeTitle(e, s, false) },
		"inverted":               func(e, s int, m bool) string { return resumeTitle(e, s, !m) },
		"episode number dropped": func(_, s int, m bool) string { return resumeTitle(0, s, m) },
		"no question mark":       func(e, s int, m bool) string { return strings.TrimSuffix(resumeTitle(e, s, m), "?") },
		"seconds ignored":        func(e, _ int, m bool) string { return resumeTitle(e, 0, m) },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, titleFailures(m), "mutant %q survived", name)
	}
}

// ---- findProgress -----------------------------------------------------------------

type storeMock struct {
	rows  map[string]*tracking.Anime
	errAt map[string]error
	asked []string
}

func (s *storeMock) GetAnime(_ int, key string) (*tracking.Anime, error) {
	s.asked = append(s.asked, key)
	if err := s.errAt[key]; err != nil {
		return nil, err
	}
	return s.rows[key], nil
}

func TestFindProgress(t *testing.T) {
	const url = "https://x/anime"
	own := &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 100}
	legacy := &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 50}
	boom := errors.New("db locked")
	tests := []struct {
		name      string
		store     storeMock
		want      *tracking.Anime
		wantErr   error
		wantAsked []string
	}{
		{"own key", storeMock{rows: map[string]*tracking.Anime{url + ":ep3": own, url: legacy}}, own, nil, []string{url + ":ep3"}},
		{"legacy key after a miss", storeMock{rows: map[string]*tracking.Anime{url: legacy}}, legacy, nil, []string{url + ":ep3", url}},
		{"nothing saved", storeMock{}, nil, nil, []string{url + ":ep3", url}},
		{"error on the own key stops the search", storeMock{errAt: map[string]error{url + ":ep3": boom}, rows: map[string]*tracking.Anime{url: legacy}}, nil, boom, []string{url + ":ep3"}},
		{"error on the legacy key", storeMock{errAt: map[string]error{url: boom}}, nil, boom, []string{url + ":ep3", url}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.store
			got, err := findProgress(&s, 7, url, 3)
			assert.Same(t, tc.want, got)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantAsked, s.asked)
		})
	}
}

func findFailures(impl func(progressStore, int, string, int) (*tracking.Anime, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	const url = "u"
	own := &tracking.Anime{PlaybackTime: 1}
	legacy := &tracking.Anime{PlaybackTime: 2}
	boom := errors.New("x")
	check := func(name string, s *storeMock, want *tracking.Anime, wantErr error, wantAsked string) {
		got, err := impl(s, 1, url, 5)
		if got != want || !errors.Is(err, wantErr) || (wantErr == nil && err != nil) || fmt.Sprint(s.asked) != wantAsked {
			f = append(f, fmt.Sprintf("%s: got (%v,%v) asked %v", name, got, err, s.asked))
		}
	}
	check("own", &storeMock{rows: map[string]*tracking.Anime{"u:ep5": own, "u": legacy}}, own, nil, "[u:ep5]")
	check("legacy", &storeMock{rows: map[string]*tracking.Anime{"u": legacy}}, legacy, nil, "[u:ep5 u]")
	check("none", &storeMock{}, nil, nil, "[u:ep5 u]")
	check("error first", &storeMock{errAt: map[string]error{"u:ep5": boom}, rows: map[string]*tracking.Anime{"u": legacy}}, nil, boom, "[u:ep5]")
	check("error second", &storeMock{errAt: map[string]error{"u": boom}}, nil, boom, "[u:ep5 u]")
	return f
}

func TestFindProgressMeta(t *testing.T) {
	require.Empty(t, findFailures(findProgress))
	type fn = func(progressStore, int, string, int) (*tracking.Anime, error)
	mutants := map[string]fn{
		"legacy key first": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			if p, err := s.GetAnime(a, u); p != nil || err != nil {
				return p, err
			}
			return s.GetAnime(a, trackingKey(u, n))
		},
		"never legacy": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			return s.GetAnime(a, trackingKey(u, n))
		},
		"always both": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			p, err := s.GetAnime(a, trackingKey(u, n))
			if err != nil {
				return nil, err
			}
			l, lerr := s.GetAnime(a, u)
			if p != nil {
				return p, nil
			}
			return l, lerr
		},
		"legacy after an error": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			p, err := s.GetAnime(a, trackingKey(u, n))
			if p == nil {
				return s.GetAnime(a, u)
			}
			return p, err
		},
		"error swallowed": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			p, err := findProgress(s, a, u, n)
			if err != nil {
				return nil, nil //nolint:nilerr // deliberate mutant: hides the failure
			}
			return p, nil
		},
		"key without episode number": func(s progressStore, a int, u string, n int) (*tracking.Anime, error) {
			return findProgress(s, a, u, 0)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, findFailures(m), "mutant %q survived", name)
	}
}

// ---- resumePositionFor --------------------------------------------------------------

type resumeAsk struct {
	answer bool
	calls  int
	gotEp  int
	gotSec int
}

func (a *resumeAsk) fn(ep, sec int) bool { a.calls++; a.gotEp, a.gotSec = ep, sec; return a.answer }

type overrideMock struct {
	pos       int
	fromStart bool
	ok        bool
	calls     int
}

func (o *overrideMock) fn() (int, bool, bool) { o.calls++; return o.pos, o.fromStart, o.ok }

var resumeCases = []struct {
	name        string
	progress    *tracking.Anime
	episode     int
	override    overrideMock
	ask         bool
	want        int
	wantAsked   bool
	wantTookOvr bool
}{
	{"no progress", nil, 3, overrideMock{}, true, 0, false, false},
	{"zero seconds saved", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 0}, 3, overrideMock{}, true, 0, false, false},
	{"negative seconds saved", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: -5}, 3, overrideMock{}, true, 0, false, false},
	{"saved for another episode", &tracking.Anime{EpisodeNumber: 2, PlaybackTime: 300}, 3, overrideMock{ok: true, pos: 99}, true, 0, false, false},
	{"user accepts", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 300}, 3, overrideMock{}, true, 300, true, true},
	{"user declines", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 300}, 3, overrideMock{}, false, 0, true, true},
	{"one second is worth resuming", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 1}, 3, overrideMock{}, true, 1, true, true},
	{"override resumes without asking", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 300}, 3, overrideMock{ok: true, pos: 120}, false, 120, false, true},
	{"override from start without asking", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 300}, 3, overrideMock{ok: true, fromStart: true, pos: 120}, true, 0, false, true},
	{"override position 0 is respected", &tracking.Anime{EpisodeNumber: 3, PlaybackTime: 300}, 3, overrideMock{ok: true, pos: 0}, true, 0, false, true},
}

func resumeFailures(impl func(*tracking.Anime, int, func() (int, bool, bool), func(int, int) bool) int) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range resumeCases {
		o := tc.override
		a := &resumeAsk{answer: tc.ask}
		got := impl(tc.progress, tc.episode, o.fn, a.fn)
		if got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d want %d", tc.name, got, tc.want))
		}
		if (a.calls > 0) != tc.wantAsked {
			f = append(f, fmt.Sprintf("%s: asked=%v want %v", tc.name, a.calls > 0, tc.wantAsked))
		}
		if (o.calls > 0) != tc.wantTookOvr {
			f = append(f, fmt.Sprintf("%s: override consulted=%v want %v", tc.name, o.calls > 0, tc.wantTookOvr))
		}
		if a.calls > 0 && (a.gotEp != tc.episode || a.gotSec != tc.progress.PlaybackTime) {
			f = append(f, fmt.Sprintf("%s: asked about (%d,%d)", tc.name, a.gotEp, a.gotSec))
		}
	}
	return f
}

func TestResumePositionFor(t *testing.T) {
	assert.Empty(t, resumeFailures(resumePositionFor))
}

func TestResumePositionForMeta(t *testing.T) {
	require.Empty(t, resumeFailures(resumePositionFor))
	type fn = func(*tracking.Anime, int, func() (int, bool, bool), func(int, int) bool) int
	mutants := map[string]fn{
		"always resumes": func(p *tracking.Anime, _ int, _ func() (int, bool, bool), _ func(int, int) bool) int {
			if p == nil {
				return 0
			}
			return p.PlaybackTime
		},
		"never resumes": func(*tracking.Anime, int, func() (int, bool, bool), func(int, int) bool) int { return 0 },
		"ignores the episode check": func(p *tracking.Anime, _ int, o func() (int, bool, bool), a func(int, int) bool) int {
			if p == nil {
				return 0
			}
			return resumePositionFor(p, p.EpisodeNumber, o, a)
		},
		"never asks": func(p *tracking.Anime, e int, o func() (int, bool, bool), _ func(int, int) bool) int {
			return resumePositionFor(p, e, o, func(int, int) bool { return false })
		},
		"asks even with an override": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			if p != nil && p.PlaybackTime > 0 && p.EpisodeNumber == e {
				a(e, p.PlaybackTime)
			}
			return resumePositionFor(p, e, o, a)
		},
		"override ignored": func(p *tracking.Anime, e int, _ func() (int, bool, bool), a func(int, int) bool) int {
			return resumePositionFor(p, e, func() (int, bool, bool) { return 0, false, false }, a)
		},
		"from-start ignored": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			return resumePositionFor(p, e, func() (int, bool, bool) { pos, _, ok := o(); return pos, false, ok }, a)
		},
		"declining still resumes": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			return resumePositionFor(p, e, o, func(x, y int) bool { a(x, y); return true })
		},
		"zero seconds is resumable": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			if p != nil && p.PlaybackTime == 0 && p.EpisodeNumber == e {
				return 1
			}
			return resumePositionFor(p, e, o, a)
		},
		"asks about the wrong time": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			return resumePositionFor(p, e, o, func(x, y int) bool { return a(x, y+1) })
		},
		"override position off by one": func(p *tracking.Anime, e int, o func() (int, bool, bool), a func(int, int) bool) int {
			got := resumePositionFor(p, e, o, a)
			if got > 0 && p != nil && got != p.PlaybackTime {
				return got + 1
			}
			return got
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, resumeFailures(m), "mutant %q survived", name)
	}
}

// ---- trackingRecord -----------------------------------------------------------------

func TestTrackingRecord(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	ep := &models.Episode{URL: "https://x/ep3", Title: models.TitleDetails{English: "The Test"}}
	ctx := mediaContext{animeTitle: "Show", source: "HiAnime", audio: "Sub"}

	got := trackingRecord(20, ep, 3, 1425*time.Second+700*time.Millisecond, ctx, now)

	assert.Equal(t, tracking.Anime{
		AnilistID: 20, AllanimeID: "https://x/ep3", EpisodeNumber: 3, Duration: 1425,
		Title: "The Test", AnimeTitle: "Show", Source: "HiAnime", Audio: "Sub", LastUpdated: now,
	}, got)
	assert.Zero(t, got.PlaybackTime, "a new row starts at zero; the player updates it")
}

func recordFailures(impl func(int, *models.Episode, int, time.Duration, mediaContext, time.Time) tracking.Anime) (f []string) {
	now := time.Unix(1_700_000_000, 0)
	ep := &models.Episode{URL: "u", Title: models.TitleDetails{Romaji: "Tesuto"}}
	got := impl(20, ep, 3, 90*time.Second, mediaContext{"Show", "Src", "Dub"}, now)
	want := tracking.Anime{AnilistID: 20, AllanimeID: "u", EpisodeNumber: 3, Duration: 90, Title: "Tesuto", AnimeTitle: "Show", Source: "Src", Audio: "Dub", LastUpdated: now}
	if got != want {
		f = append(f, fmt.Sprintf("got %+v", got))
	}
	return f
}

func TestTrackingRecordMeta(t *testing.T) {
	require.Empty(t, recordFailures(trackingRecord))
	type fn = func(int, *models.Episode, int, time.Duration, mediaContext, time.Time) tracking.Anime
	with := func(mut func(*tracking.Anime)) fn {
		return func(a int, e *models.Episode, n int, d time.Duration, c mediaContext, now time.Time) tracking.Anime {
			r := trackingRecord(a, e, n, d, c, now)
			mut(&r)
			return r
		}
	}
	mutants := map[string]fn{
		"anilist lost":             with(func(r *tracking.Anime) { r.AnilistID = 0 }),
		"key lost":                 with(func(r *tracking.Anime) { r.AllanimeID = "" }),
		"episode lost":             with(func(r *tracking.Anime) { r.EpisodeNumber = 0 }),
		"duration in ms":           with(func(r *tracking.Anime) { r.Duration *= 1000 }),
		"title lost":               with(func(r *tracking.Anime) { r.Title = "" }),
		"show lost":                with(func(r *tracking.Anime) { r.AnimeTitle = "" }),
		"source lost":              with(func(r *tracking.Anime) { r.Source = "" }),
		"audio lost":               with(func(r *tracking.Anime) { r.Audio = "" }),
		"time lost":                with(func(r *tracking.Anime) { r.LastUpdated = time.Time{} }),
		"source and audio swapped": with(func(r *tracking.Anime) { r.Source, r.Audio = r.Audio, r.Source }),
	}
	for name, m := range mutants {
		assert.NotEmpty(t, recordFailures(m), "mutant %q survived", name)
	}
}

func TestUpdateTrackingWithDurationIsSafeWithoutATracker(t *testing.T) {
	ep := &models.Episode{URL: "u"}
	assert.NotPanics(t, func() { updateTrackingWithDuration(nil, 1, ep, 1, time.Minute) })
	assert.NotPanics(t, func() { updateTrackingWithDuration(nil, 1, ep, 1, 0) })
}

func TestGetEpisodeTitleOrder(t *testing.T) {
	assert.Equal(t, "En", getEpisodeTitle(models.TitleDetails{English: "En", Romaji: "Ro", Japanese: "Ja"}))
	assert.Equal(t, "Ro", getEpisodeTitle(models.TitleDetails{Romaji: "Ro", Japanese: "Ja"}))
	assert.Equal(t, "Ja", getEpisodeTitle(models.TitleDetails{Japanese: "Ja"}))
	assert.Equal(t, "No title", getEpisodeTitle(models.TitleDetails{}))
}

// With a real tracker the row lands in SQLite under the episode's key, and a
// zero duration saves nothing.
func TestUpdateTrackingWithDurationSavesTheRow(t *testing.T) {
	if !tracking.IsCgoEnabled {
		t.Skip("tracking needs CGO/SQLite")
	}
	_ = tracking.CloseGlobalTracker()
	tr := tracking.NewLocalTracker(t.TempDir() + "/progress.db")
	if tr == nil {
		t.Skip("tracking unavailable in this build")
	}
	t.Cleanup(func() { _ = tracking.CloseGlobalTracker() })

	ep := &models.Episode{URL: "https://x/ep2", Title: models.TitleDetails{English: "Two"}}
	SetMediaContext("Show", "HiAnime", "Sub")
	t.Cleanup(func() { SetMediaContext("", "", "") })

	updateTrackingWithDuration(tr, 20, ep, 2, 0)
	got, err := tr.GetAnime(20, ep.URL)
	require.NoError(t, err)
	assert.Nil(t, got, "no duration, no row")

	updateTrackingWithDuration(tr, 20, ep, 2, 1450*time.Second)
	got, err = tr.GetAnime(20, ep.URL)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 2, got.EpisodeNumber)
	assert.Equal(t, 1450, got.Duration)
	assert.Equal(t, "Two", got.Title)
	assert.Equal(t, "Show", got.AnimeTitle)
	assert.Equal(t, "HiAnime", got.Source)
}
