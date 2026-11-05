package playback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/appflow"
	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// fakeFlow is a flowEnv made of recorders. It serves scripted answers and
// logs every call, in order, so a scenario can assert the exact sequence.
type fakeFlow struct {
	env flowEnv

	mu     sync.Mutex
	events []string

	streamErrs   []error // one per stream resolution; none left = success
	plays        []error // one per playback (movie or episode); none left = the user quits
	inputs       []string
	moviePlays   []moviePlay
	episodePlays []episodePlay
	updaterArgs  []updaterArg
	mediaTypes   []string
	metas        []*util.MediaMeta
	streamedEps  []string // the Number of the episode each stream resolution got
	streamedAnim []string
	subSeries    []subCall
	subMovies    []subCall
	subErr       error // returned by the delegated HandleSeries / HandleMovie
	inputMovie   []bool
}

type moviePlay struct {
	url, animeURL, number, name string
	selected, mal, anilist      int
	season                      int
	episodes                    []models.Episode
	anime                       *models.Anime
}

type episodePlay struct {
	anime   *models.Anime
	num     int
	url     string
	label   string
	discord bool
	paused  bool
}

type updaterArg struct {
	duration time.Duration
	enabled  bool
	paused   bool
}

type subCall struct {
	anime *models.Anime
	eps   int
	total int
}

func (f *fakeFlow) log(format string, a ...any) {
	f.mu.Lock()
	f.events = append(f.events, fmt.Sprintf(format, a...))
	f.mu.Unlock()
}

func (f *fakeFlow) nextPlay() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.plays) == 0 {
		return player.ErrUserQuit
	}
	err := f.plays[0]
	f.plays = f.plays[1:]
	return err
}

func newFakeFlow() *fakeFlow {
	f := &fakeFlow{}
	f.env = flowEnv{
		preWarmMPV:        func() { f.log("prewarm") },
		setEpisodeListCtx: func(name, _, _ string) { f.log("listctx:%s", name) },
		resetTerminal:     func() { f.log("reset") },
		streamURL: func(_ context.Context, ep *models.Episode, a *models.Anime) (string, error) {
			f.mu.Lock()
			i := len(f.streamedEps)
			f.streamedEps = append(f.streamedEps, ep.Number)
			f.streamedAnim = append(f.streamedAnim, a.Name)
			var err error
			if i < len(f.streamErrs) {
				err = f.streamErrs[i]
			}
			f.events = append(f.events, "stream")
			f.mu.Unlock()
			if err != nil {
				return "", err
			}
			return "stream://" + ep.Number, nil
		},
		reportStreamFailed: func(_ *models.Anime, err error) { f.log("report:%v", err) },
		newUpdater: func(_ *models.Anime, paused *bool, _ *sync.Mutex, d time.Duration, enabled bool) *discord.RichPresenceUpdater {
			f.mu.Lock()
			f.updaterArgs = append(f.updaterArgs, updaterArg{d, enabled, *paused})
			f.mu.Unlock()
			return nil
		},
		setExactMediaType: func(t string) { f.mu.Lock(); f.mediaTypes = append(f.mediaTypes, t); f.mu.Unlock() },
		setMediaMeta:      func(m *util.MediaMeta) { f.mu.Lock(); f.metas = append(f.metas, m); f.mu.Unlock() },
		downloadAndPlay: func(url string, eps []models.Episode, sel int, animeURL, number string, mal, anilist int, _ *discord.RichPresenceUpdater, name string, season int, a *models.Anime) error {
			f.mu.Lock()
			f.moviePlays = append(f.moviePlays, moviePlay{url, animeURL, number, name, sel, mal, anilist, season, eps, a})
			f.mu.Unlock()
			f.log("play")
			return f.nextPlay()
		},
		userInput: func(isMovie ...bool) string {
			f.mu.Lock()
			f.inputMovie = append(f.inputMovie, len(isMovie) > 0 && isMovie[0])
			in := "q"
			if len(f.inputs) > 0 {
				in, f.inputs = f.inputs[0], f.inputs[1:]
			}
			f.events = append(f.events, "input:"+in)
			f.mu.Unlock()
			return in
		},
		playEpisode: func(_ context.Context, a *models.Anime, _ []models.Episode, num int, url, label string, discordOn bool, paused *bool, _ *sync.Mutex) error {
			f.mu.Lock()
			f.episodePlays = append(f.episodePlays, episodePlay{a, num, url, label, discordOn, *paused})
			f.mu.Unlock()
			f.log("episode:%d", num)
			return f.nextPlay()
		},
		handleMovie: func(_ context.Context, a *models.Anime, eps []models.Episode, _ bool) error {
			f.mu.Lock()
			f.subMovies = append(f.subMovies, subCall{a, len(eps), 0})
			f.mu.Unlock()
			f.log("sub-movie")
			return f.subErr
		},
		handleSeries: func(_ context.Context, a *models.Anime, eps []models.Episode, total int, _ bool) error {
			f.mu.Lock()
			f.subSeries = append(f.subSeries, subCall{a, len(eps), total})
			f.mu.Unlock()
			f.log("sub-series")
			return f.subErr
		},
	}
	return f
}

func (f *fakeFlow) trace() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.events, ",")
}

func movieAnime() *models.Anime {
	return &models.Anime{
		Name: "Film", URL: "https://example.test/film", MediaType: models.MediaTypeMovie,
		MalID: 11, AnilistID: 22, TMDBID: 33, IMDBID: "tt44", Year: "1999", CurrentSeason: 7,
	}
}

func movieEpisodes() []models.Episode {
	return []models.Episode{{Number: "Film", Num: 1, URL: "https://example.test/film/ep", Duration: 90}}
}

func seriesAnime() *models.Anime {
	return &models.Anime{Name: "Show", URL: "https://example.test/show", MediaType: models.MediaTypeAnime, Source: "src", Audio: "sub"}
}

// ---- runMovie ----

type movieScenario struct {
	name  string
	setup func(f *fakeFlow, c *changerMock)
	// the loop under test returns want (nil allowed) and logs exactly trace
	want  error
	trace string
	check func(f *fakeFlow) []string
}

func movieScenarios() []movieScenario {
	return []movieScenario{
		{
			name:  "the user quits from inside the player",
			setup: func(f *fakeFlow, _ *changerMock) {},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play",
			check: func(f *fakeFlow) []string {
				var bad []string
				if len(f.inputMovie) != 0 {
					bad = append(bad, "the post-playback menu must not open after a quit")
				}
				if len(f.moviePlays) != 1 {
					return append(bad, fmt.Sprintf("%d playbacks, want 1", len(f.moviePlays)))
				}
				p := f.moviePlays[0]
				if p.url != "stream://Film" || p.selected != 1 || p.animeURL != "https://example.test/film" ||
					p.number != "Film" || p.name != "Film" || p.season != 7 {
					bad = append(bad, fmt.Sprintf("playback got %+v", p))
				}
				if p.mal != 11 || p.anilist != 22 {
					bad = append(bad, fmt.Sprintf("ids routed as mal=%d anilist=%d, want 11/22", p.mal, p.anilist))
				}
				if len(p.episodes) != 1 || p.anime == nil || p.anime.Name != "Film" {
					bad = append(bad, "the episode list or anime handed to the player is wrong")
				}
				if len(f.mediaTypes) != 1 || f.mediaTypes[0] != "movie" {
					bad = append(bad, fmt.Sprintf("media type %v, want [movie]", f.mediaTypes))
				}
				if len(f.metas) != 1 || *f.metas[0] != (util.MediaMeta{OfficialTitle: "Film", Year: "1999", TMDBID: 33, IMDBID: "tt44", AnilistID: 22, MalID: 11}) {
					bad = append(bad, fmt.Sprintf("media meta %+v", f.metas))
				}
				if len(f.updaterArgs) != 1 || f.updaterArgs[0].duration != 90*time.Second || f.updaterArgs[0].enabled || f.updaterArgs[0].paused {
					bad = append(bad, fmt.Sprintf("updater args %+v, want 90s, discord off and not paused", f.updaterArgs))
				}
				return bad
			},
		},
		{
			name:  "a stream failure reports and returns to anime selection",
			setup: func(f *fakeFlow, _ *changerMock) { f.streamErrs = []error{errors.New("no stream")} },
			want:  player.ErrBackToAnimeSelection,
			trace: "prewarm,listctx:Film,stream,report:no stream",
			check: func(f *fakeFlow) []string {
				if len(f.moviePlays) != 0 || len(f.updaterArgs) != 0 || len(f.mediaTypes) != 0 {
					return []string{"nothing may start after a failed stream"}
				}
				return nil
			},
		},
		{
			name:  "q after the movie ends quits",
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"q"} },
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,input:q",
			check: func(f *fakeFlow) []string {
				if len(f.inputMovie) != 1 || !f.inputMovie[0] {
					return []string{"the post-playback menu must be the movie menu"}
				}
				return nil
			},
		},
		{
			name: "a playback error still shows the menu",
			setup: func(f *fakeFlow, _ *changerMock) {
				f.plays = []error{errors.New("mpv crashed")}
				f.inputs = []string{"q"}
			},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,input:q",
		},
		{
			name:  "replay plays the same movie again",
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil, nil}; f.inputs = []string{"n", "q"} },
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,input:n,stream,play,input:q",
		},
		{
			name: "change with a pending CONTINUE row goes back to the session loop",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.err = appflow.ErrHomeContinue
			},
			want:  player.ErrBackToAnimeSelection,
			trace: "prewarm,listctx:Film,stream,play",
		},
		{
			name: "a failed change keeps the current movie",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime, player.ErrUserQuit}
				c.err = errors.New("no network")
			},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,stream,play",
		},
		{
			name: "change to another movie plays that one",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{nil, player.ErrUserQuit}
				f.inputs = []string{"c"}
				c.anime = &models.Anime{Name: "Other", MediaType: models.MediaTypeMovie}
				c.eps = []models.Episode{{Number: "Other", Num: 1}}
			},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,input:c,stream,play",
			check: func(f *fakeFlow) []string {
				if len(f.streamedAnim) != 2 || f.streamedAnim[1] != "Other" || f.streamedEps[1] != "Other" {
					return []string{fmt.Sprintf("second stream was for %v/%v, want Other", f.streamedAnim, f.streamedEps)}
				}
				return nil
			},
		},
		{
			name: "change to a series hands over to the series loop",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.anime = seriesAnime()
				c.eps = sampleEpisodes(12)
			},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,sub-series",
			check: func(f *fakeFlow) []string {
				if len(f.subSeries) != 1 || f.subSeries[0].anime.Name != "Show" || f.subSeries[0].eps != 12 || f.subSeries[0].total != 12 {
					return []string{fmt.Sprintf("series call %+v", f.subSeries)}
				}
				return nil
			},
		},
		{
			name: "the series loop going back to selection propagates",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				f.subErr = player.ErrBackToAnimeSelection
				c.anime = seriesAnime()
				c.eps = sampleEpisodes(12)
			},
			want:  player.ErrBackToAnimeSelection,
			trace: "prewarm,listctx:Film,stream,play,sub-series",
		},
		{
			name: "any other series loop error is not the movie loop's to return",
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				f.subErr = errors.New("series blew up")
				c.anime = seriesAnime()
				c.eps = sampleEpisodes(12)
			},
			want:  nil,
			trace: "prewarm,listctx:Film,stream,play,sub-series",
		},
	}
}

func runMovieScenario(t *testing.T, s movieScenario, tamper func(*fakeFlow)) []string {
	t.Helper()
	f := newFakeFlow()
	c := &changerMock{}
	prev := changeAnimeFn
	changeAnimeFn = c.fn
	defer func() { changeAnimeFn = prev }()
	s.setup(f, c)
	if tamper != nil {
		tamper(f)
	}
	err := runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), false)
	var bad []string
	if (s.want == nil) != (err == nil) || (err != nil && !errors.Is(err, s.want)) {
		bad = append(bad, fmt.Sprintf("returned %v, want %v", err, s.want))
	}
	if got := f.trace(); got != s.trace {
		bad = append(bad, fmt.Sprintf("trace %q, want %q", got, s.trace))
	}
	if s.check != nil {
		bad = append(bad, s.check(f)...)
	}
	return bad
}

func TestRunMovie(t *testing.T) {
	for _, s := range movieScenarios() {
		t.Run(s.name, func(t *testing.T) {
			if bad := runMovieScenario(t, s, nil); len(bad) > 0 {
				t.Fatal(strings.Join(bad, "\n"))
			}
		})
	}
}

// The movie handed to the stream resolver is the first episode, and the
// anime's own episode list holds exactly that one while it plays.
func TestRunMovie_PlaysWithOnlyTheMovieInTheAnimeEpisodeList(t *testing.T) {
	f := newFakeFlow()
	a := movieAnime()
	a.Episodes = sampleEpisodes(5)
	_ = runMovie(f.env, context.Background(), a, movieEpisodes(), false)
	if len(f.moviePlays) != 1 || len(f.moviePlays[0].anime.Episodes) != 1 || f.moviePlays[0].anime.Episodes[0].Number != "Film" {
		t.Fatalf("anime episodes while playing: %+v", f.moviePlays)
	}
}

func TestRunMovie_DiscordFlagReachesTheUpdater(t *testing.T) {
	f := newFakeFlow()
	_ = runMovie(f.env, context.Background(), movieAnime(), movieEpisodes(), true)
	if len(f.updaterArgs) != 1 || !f.updaterArgs[0].enabled {
		t.Fatalf("updater args %+v", f.updaterArgs)
	}
}

// Metadata that arrives before the stream is resolved replaces the episode
// list; metadata that is still on its way never holds playback up.
func TestRunMovie_KeepsArrivedMetadataOnly(t *testing.T) {
	t.Run("arrived", func(t *testing.T) {
		done := make(chan struct{})
		useMovieData(t, func(_ int, a *models.Anime) error {
			a.Episodes = sampleEpisodes(3)
			close(done)
			return nil
		})
		f := newFakeFlow()
		inner := f.env.streamURL
		f.env.streamURL = func(ctx context.Context, ep *models.Episode, a *models.Anime) (string, error) {
			<-done
			time.Sleep(50 * time.Millisecond) // let the lookup's result reach the channel
			return inner(ctx, ep, a)
		}
		a := movieAnime()
		a.MediaType = models.MediaTypeAnime // Jikan only serves anime
		_ = runMovie(f.env, context.Background(), a, movieEpisodes(), false)
		if got := len(a.Episodes); got != 3 {
			t.Fatalf("episodes after arrival = %d, want 3", got)
		}
	})
	t.Run("late", func(t *testing.T) {
		release := make(chan struct{})
		t.Cleanup(func() { close(release) })
		useMovieData(t, func(_ int, a *models.Anime) error {
			<-release
			a.Episodes = sampleEpisodes(3)
			return nil
		})
		f := newFakeFlow()
		a := movieAnime()
		a.MediaType = models.MediaTypeAnime
		_ = runMovie(f.env, context.Background(), a, movieEpisodes(), false)
		if got := len(a.Episodes); got != 1 {
			t.Fatalf("episodes with late metadata = %d, want the movie alone", got)
		}
	})
}

func TestMeta_RunMovie_CatchesTamperedCollaborators(t *testing.T) {
	// Each tamper makes one collaborator misbehave the way a broken loop would
	// look from outside; the scenarios must notice every one.
	tampers := map[string]func(f *fakeFlow){
		"swaps the mal and anilist ids": func(f *fakeFlow) {
			inner := f.env.downloadAndPlay
			f.env.downloadAndPlay = func(u string, e []models.Episode, s int, au, n string, mal, anilist int, up *discord.RichPresenceUpdater, nm string, se int, a *models.Anime) error {
				return inner(u, e, s, au, n, anilist, mal, up, nm, se, a)
			}
		},
		"passes the wrong episode index": func(f *fakeFlow) {
			inner := f.env.downloadAndPlay
			f.env.downloadAndPlay = func(u string, e []models.Episode, s int, au, n string, mal, anilist int, up *discord.RichPresenceUpdater, nm string, se int, a *models.Anime) error {
				return inner(u, e, s+1, au, n, mal, anilist, up, nm, se, a)
			}
		},
		"passes the wrong season": func(f *fakeFlow) {
			inner := f.env.downloadAndPlay
			f.env.downloadAndPlay = func(u string, e []models.Episode, s int, au, n string, mal, anilist int, up *discord.RichPresenceUpdater, nm string, se int, a *models.Anime) error {
				return inner(u, e, s, au, n, mal, anilist, up, nm, 0, a)
			}
		},
		"plays a different url": func(f *fakeFlow) {
			inner := f.env.downloadAndPlay
			f.env.downloadAndPlay = func(u string, e []models.Episode, s int, au, n string, mal, anilist int, up *discord.RichPresenceUpdater, nm string, se int, a *models.Anime) error {
				return inner(u+"x", e, s, au, n, mal, anilist, up, nm, se, a)
			}
		},
		"the user is asked twice": func(f *fakeFlow) {
			inner := f.env.userInput
			f.env.userInput = func(m ...bool) string { inner(m...); return inner(m...) }
		},
		"the menu is not the movie menu": func(f *fakeFlow) {
			inner := f.env.userInput
			f.env.userInput = func(...bool) string { return inner() }
		},
		"the stream is resolved twice": func(f *fakeFlow) {
			inner := f.env.streamURL
			f.env.streamURL = func(c context.Context, e *models.Episode, a *models.Anime) (string, error) {
				_, _ = inner(c, e, a)
				return inner(c, e, a)
			}
		},
		"the failure report is dropped": func(f *fakeFlow) { f.env.reportStreamFailed = func(*models.Anime, error) {} },
		"mpv is never warmed":           func(f *fakeFlow) { f.env.preWarmMPV = func() {} },
		"the media type is wrong": func(f *fakeFlow) {
			inner := f.env.setExactMediaType
			f.env.setExactMediaType = func(string) { inner("tv") }
		},
		"the meta loses the ids": func(f *fakeFlow) {
			inner := f.env.setMediaMeta
			f.env.setMediaMeta = func(m *util.MediaMeta) { c := *m; c.MalID, c.AnilistID = 0, 0; inner(&c) }
		},
		"the sub-series result is swallowed": func(f *fakeFlow) {
			inner := f.env.handleSeries
			f.env.handleSeries = func(c context.Context, a *models.Anime, e []models.Episode, n int, d bool) error {
				_ = inner(c, a, e, n, d)
				return nil
			}
		},
	}
	for name, tamper := range tampers {
		t.Run(name, func(t *testing.T) {
			for _, s := range movieScenarios() {
				if len(runMovieScenario(t, s, tamper)) > 0 {
					return
				}
			}
			t.Fatalf("tamper %q went unnoticed by every scenario", name)
		})
	}
}

// ---- runSeries ----

type seriesScenario struct {
	name   string
	setup  func(f *fakeFlow, c *changerMock) // also sets the picker via the picks queue
	picks  []pickerMock                      // answers of the episode picker, in order
	total  int
	want   error
	trace  string
	checks func(f *fakeFlow) []string
}

func seriesScenarios() []seriesScenario {
	back := pickerMock{err: player.ErrBackRequested}
	return []seriesScenario{
		{
			name:  "the first pick plays and a quit ends the session",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) {},
			want:  nil,
			trace: "reset,listctx:Show,episode:2",
			checks: func(f *fakeFlow) []string {
				if len(f.episodePlays) != 1 {
					return []string{"want one playback"}
				}
				p := f.episodePlays[0]
				var bad []string
				if p.num != 2 || p.url != "https://example.test/ep2" || p.label != "Episode 2" || p.anime.Name != "Show" || !p.discord || p.paused {
					bad = append(bad, fmt.Sprintf("playback got %+v", p))
				}
				if len(f.inputMovie) != 0 {
					bad = append(bad, "no menu after a quit")
				}
				return bad
			},
		},
		{
			name:  "backing out of the first pick leaves for anime selection",
			picks: []pickerMock{back}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) {},
			want:  player.ErrBackToAnimeSelection,
			trace: "reset,listctx:Show",
		},
		{
			name:  "q in the menu quits",
			picks: []pickerMock{okPick(1)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"q"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:1,input:q",
			checks: func(f *fakeFlow) []string {
				if len(f.inputMovie) != 1 || f.inputMovie[0] {
					return []string{"the post-playback menu must be the series menu"}
				}
				return nil
			},
		},
		{
			name:  "a playback error still shows the menu",
			picks: []pickerMock{okPick(1)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{errors.New("boom")}; f.inputs = []string{"q"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:1,input:q",
		},
		{
			name:  "n plays the next episode",
			picks: []pickerMock{okPick(3)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"n"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:3,input:n,episode:4",
		},
		{
			name:  "p plays the previous episode",
			picks: []pickerMock{okPick(3)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"p"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:3,input:p,episode:2",
		},
		{
			name:  "next stops at the last episode",
			picks: []pickerMock{okPick(4)}, total: 4,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"n"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:4,input:n,episode:4",
		},
		{
			name:  "previous stops at the first episode",
			picks: []pickerMock{okPick(1)}, total: 4,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"p"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:1,input:p,episode:1",
		},
		{
			name:  "e picks another episode",
			picks: []pickerMock{okPick(2), okPick(9)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"e"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:2,input:e,episode:9",
		},
		{
			name:  "backing out of the e pick keeps the current episode",
			picks: []pickerMock{okPick(2), back}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"e"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:2,input:e,episode:2",
		},
		{
			name:  "a failing e pick keeps the current episode",
			picks: []pickerMock{okPick(2), {err: errors.New("tty gone")}}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{nil}; f.inputs = []string{"e"} },
			want:  nil,
			trace: "reset,listctx:Show,episode:2,input:e,episode:2",
		},
		{
			name:  "back from server selection picks again",
			picks: []pickerMock{okPick(2), okPick(5)}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{player.ErrBackToEpisodeSelection} },
			want:  nil,
			trace: "reset,listctx:Show,episode:2,episode:5",
		},
		{
			name:  "back from server selection then backing out of the picker leaves",
			picks: []pickerMock{okPick(2), back}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) { f.plays = []error{player.ErrBackToEpisodeSelection} },
			want:  player.ErrBackToAnimeSelection,
			trace: "reset,listctx:Show,episode:2",
		},
		{
			name:  "change with a pending CONTINUE row goes back to the session loop",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.err = appflow.ErrHomeContinue
			},
			want:  player.ErrBackToAnimeSelection,
			trace: "reset,listctx:Show,episode:2",
		},
		{
			name:  "a failed change keeps playing the same series",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime, player.ErrUserQuit}
				c.err = errors.New("no network")
			},
			want:  nil,
			trace: "reset,listctx:Show,episode:2,episode:2",
		},
		{
			name:  "change to a movie hands over to the movie loop",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.anime = movieAnime()
				c.eps = movieEpisodes()
			},
			want:  nil,
			trace: "reset,listctx:Show,episode:2,sub-movie",
			checks: func(f *fakeFlow) []string {
				if len(f.subMovies) != 1 || f.subMovies[0].anime.Name != "Film" || f.subMovies[0].eps != 1 {
					return []string{fmt.Sprintf("movie call %+v", f.subMovies)}
				}
				return nil
			},
		},
		{
			name:  "the movie loop going back to selection propagates",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				f.subErr = player.ErrBackToAnimeSelection
				c.anime = movieAnime()
				c.eps = movieEpisodes()
			},
			want:  player.ErrBackToAnimeSelection,
			trace: "reset,listctx:Show,episode:2,sub-movie",
		},
		{
			name:  "any other movie loop error is not the series loop's to return",
			picks: []pickerMock{okPick(2)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				f.subErr = errors.New("movie blew up")
				c.anime = movieAnime()
				c.eps = movieEpisodes()
			},
			want:  nil,
			trace: "reset,listctx:Show,episode:2,sub-movie",
		},
		{
			name:  "change to another series picks its episode and plays it",
			picks: []pickerMock{okPick(2), okPick(6)}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.anime = &models.Anime{Name: "Other", MediaType: models.MediaTypeAnime}
				c.eps = sampleEpisodes(20)
			},
			want:  nil,
			trace: "reset,listctx:Show,episode:2,episode:6",
			checks: func(f *fakeFlow) []string {
				if len(f.episodePlays) != 2 || f.episodePlays[1].anime.Name != "Other" {
					return []string{"the second playback must be of the new series"}
				}
				return nil
			},
		},
		{
			name:  "change to another series with a failed pick bounces to selection",
			picks: []pickerMock{okPick(2), {err: errors.New("tty gone")}}, total: 12,
			setup: func(f *fakeFlow, c *changerMock) {
				f.plays = []error{player.ErrChangeAnime}
				c.anime = &models.Anime{Name: "Other", MediaType: models.MediaTypeAnime}
				c.eps = sampleEpisodes(20)
			},
			want:  player.ErrBackToAnimeSelection,
			trace: "reset,listctx:Show,episode:2",
		},
		{
			name:  "navigation that finds nothing keeps the episode",
			picks: []pickerMock{okPick(2), {err: errors.New("tty gone")}}, total: 12,
			setup: func(f *fakeFlow, _ *changerMock) {
				// the next episode is missing from the list, the fallback picker fails
				f.plays = []error{nil}
				f.inputs = []string{"n"}
			},
			want:  nil,
			trace: "reset,listctx:Show,episode:2,input:n,episode:2",
		},
	}
}

func runSeriesScenario(t *testing.T, s seriesScenario, tamper func(*fakeFlow)) []string {
	t.Helper()
	f := newFakeFlow()
	c := &changerMock{}
	prev := changeAnimeFn
	changeAnimeFn = c.fn
	defer func() { changeAnimeFn = prev }()
	s.setup(f, c)
	if tamper != nil {
		tamper(f)
	}
	seq := &seqPicker{answers: s.picks}
	withSelectorInline(seq.fn, nil)

	episodes := sampleEpisodes(s.total)
	if s.name == "navigation that finds nothing keeps the episode" {
		// episode 3 is absent so "next" from 2 has nowhere to go
		episodes = []models.Episode{episodes[0], episodes[1]}
	}
	err := runSeries(f.env, context.Background(), seriesAnime(), episodes, s.total, true)
	var bad []string
	if (s.want == nil) != (err == nil) || (err != nil && !errors.Is(err, s.want)) {
		bad = append(bad, fmt.Sprintf("returned %v, want %v", err, s.want))
	}
	if got := f.trace(); got != s.trace {
		bad = append(bad, fmt.Sprintf("trace %q, want %q", got, s.trace))
	}
	if s.checks != nil {
		bad = append(bad, s.checks(f)...)
	}
	return bad
}

func TestRunSeries(t *testing.T) {
	withSelector(t, nil, nil)
	for _, s := range seriesScenarios() {
		t.Run(s.name, func(t *testing.T) {
			if bad := runSeriesScenario(t, s, nil); len(bad) > 0 {
				t.Fatal(strings.Join(bad, "\n"))
			}
		})
	}
}

func TestMeta_RunSeries_CatchesTamperedCollaborators(t *testing.T) {
	withSelector(t, nil, nil)
	tampers := map[string]func(f *fakeFlow){
		"the terminal is not reset": func(f *fakeFlow) { f.env.resetTerminal = func() {} },
		"the list context is not set": func(f *fakeFlow) {
			f.env.setEpisodeListCtx = func(string, string, string) {}
		},
		"the wrong episode number is played": func(f *fakeFlow) {
			inner := f.env.playEpisode
			f.env.playEpisode = func(c context.Context, a *models.Anime, e []models.Episode, n int, u, l string, d bool, p *bool, m *sync.Mutex) error {
				return inner(c, a, e, n+1, u, l, d, p, m)
			}
		},
		"the discord flag is dropped": func(f *fakeFlow) {
			inner := f.env.playEpisode
			f.env.playEpisode = func(c context.Context, a *models.Anime, e []models.Episode, n int, u, l string, d bool, p *bool, m *sync.Mutex) error {
				return inner(c, a, e, n, u, l, false, p, m)
			}
		},
		"the label is dropped": func(f *fakeFlow) {
			inner := f.env.playEpisode
			f.env.playEpisode = func(c context.Context, a *models.Anime, e []models.Episode, n int, u, _ string, d bool, p *bool, m *sync.Mutex) error {
				return inner(c, a, e, n, u, "", d, p, m)
			}
		},
		"the menu is asked twice": func(f *fakeFlow) {
			inner := f.env.userInput
			f.env.userInput = func(m ...bool) string { inner(m...); return inner(m...) }
		},
		"the menu is the movie menu": func(f *fakeFlow) {
			inner := f.env.userInput
			f.env.userInput = func(...bool) string { return inner(true) }
		},
		"the movie loop's result is swallowed": func(f *fakeFlow) {
			inner := f.env.handleMovie
			f.env.handleMovie = func(c context.Context, a *models.Anime, e []models.Episode, d bool) error {
				_ = inner(c, a, e, d)
				return nil
			}
		},
		"the movie loop is not called": func(f *fakeFlow) {
			f.env.handleMovie = func(context.Context, *models.Anime, []models.Episode, bool) error { return nil }
		},
	}
	for name, tamper := range tampers {
		t.Run(name, func(t *testing.T) {
			for _, s := range seriesScenarios() {
				if len(runSeriesScenario(t, s, tamper)) > 0 {
					return
				}
			}
			t.Fatalf("tamper %q went unnoticed by every scenario", name)
		})
	}
}

func TestRealFlow_WiresEveryCollaborator(t *testing.T) {
	e := realFlow()
	if e.preWarmMPV == nil || e.setEpisodeListCtx == nil || e.resetTerminal == nil || e.streamURL == nil ||
		e.reportStreamFailed == nil || e.newUpdater == nil || e.setExactMediaType == nil || e.setMediaMeta == nil ||
		e.downloadAndPlay == nil || e.userInput == nil || e.playEpisode == nil || e.handleMovie == nil || e.handleSeries == nil {
		t.Fatalf("a collaborator is nil: %+v", e)
	}
}
