package player

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- routeForPlayback ---------------------------------------------------------------------

var routeCases = []struct {
	name  string
	url   string
	isHLS bool
	want  playbackRoute
}{
	{"hls stream", "https://cdn.test/master.m3u8", true, routeHLS},
	{"hls wins over an extractable page", "https://animefire.io/video/x", true, routeHLS},
	{"hls with an empty URL is still hls", "", true, routeHLS},
	{"animefire video api", "https://animefire.io/video/x", false, routeResolve},
	{"animefire plus", "https://animefire.plus/video/x", false, routeResolve},
	{"blogger embed", "https://www.blogger.com/video.g?token=abc", false, routeResolve},
	{"blogspot embed", "https://x.blogspot.com/video/abc", false, routeResolve},
	{"extraction is case-insensitive", "HTTPS://ANIMEFIRE.IO/VIDEO/X", false, routeResolve},
	{"a direct mp4", "https://cdn.test/ep.mp4", false, routeDirect},
	{"a direct http link", "http://cdn.test/ep.mp4", false, routeDirect},
	{"an episode id", "abc123", false, routeLegacy},
	{"a path", "/files/ep.mp4", false, routeLegacy},
	{"https in capitals is not detected as http", "HTTPS://cdn.test/ep.mp4", false, routeLegacy},
	{"nothing", "", false, routeNone},
}

func routeFailures(impl func(string, bool) playbackRoute) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range routeCases {
		if got := impl(tc.url, tc.isHLS); got != tc.want {
			f = append(f, fmt.Sprintf("%s: got %d want %d", tc.name, got, tc.want))
		}
	}
	return f
}

func TestRouteForPlayback(t *testing.T) {
	assert.Empty(t, routeFailures(routeForPlayback))
}

func TestRouteForPlaybackMeta(t *testing.T) {
	require.Empty(t, routeFailures(routeForPlayback))
	mutants := map[string]func(string, bool) playbackRoute{
		"hls ignored": func(u string, _ bool) playbackRoute { return routeForPlayback(u, false) },
		"extraction first": func(u string, h bool) playbackRoute {
			if u != "" && needsVideoExtraction(u) {
				return routeResolve
			}
			return routeForPlayback(u, h)
		},
		"direct links re-extracted": func(u string, h bool) playbackRoute {
			if r := routeForPlayback(u, h); r == routeDirect {
				return routeResolve
			}
			return routeForPlayback(u, h)
		},
		"ids treated as direct": func(u string, h bool) playbackRoute {
			if r := routeForPlayback(u, h); r == routeLegacy {
				return routeDirect
			}
			return routeForPlayback(u, h)
		},
		"empty URL is legacy": func(u string, h bool) playbackRoute {
			if r := routeForPlayback(u, h); r == routeNone {
				return routeLegacy
			}
			return routeForPlayback(u, h)
		},
		"https only": func(u string, h bool) playbackRoute {
			if r := routeForPlayback(u, h); r == routeDirect && !strings.HasPrefix(u, "https") {
				return routeLegacy
			}
			return routeForPlayback(u, h)
		},
		"blogger not extracted": func(u string, h bool) playbackRoute {
			if strings.Contains(u, "blogger.com") && !h {
				return routeDirect
			}
			return routeForPlayback(u, h)
		},
		"always none": func(string, bool) playbackRoute { return routeNone },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, routeFailures(m), "mutant %q survived", name)
	}
}

// ---- resolvePlaybackURL ---------------------------------------------------------------------

type resolveSeams struct {
	extractOut map[string]string // extractActualVideoURL answers
	extractErr map[string]error
	promptOut  map[string]string // ExtractVideoSourcesWithPrompt answers
	promptErr  map[string]error

	extracted []string
	prompted  []string
}

func epList(urls ...string) []models.Episode {
	var out []models.Episode
	for i, u := range urls {
		out = append(out, models.Episode{Num: i + 1, URL: u})
	}
	return out
}

var resolveCases2 = []struct {
	name     string
	route    playbackRoute
	url      string
	eps      []models.Episode
	num      int
	seams    resolveSeams
	want     string
	wantBack bool
	wantEx   []string
	wantPr   []string
}{
	{name: "hls as is", route: routeHLS, url: "https://c/m.m3u8", want: "https://c/m.m3u8"},
	{name: "direct as is", route: routeDirect, url: "https://c/e.mp4", want: "https://c/e.mp4"},
	{name: "resolve to the CDN url", route: routeResolve, url: "https://animefire.io/video/x", want: "https://cdn/720.mp4",
		seams:  resolveSeams{extractOut: map[string]string{"https://animefire.io/video/x": "https://cdn/720.mp4"}},
		wantEx: []string{"https://animefire.io/video/x"}},
	{name: "resolve fails", route: routeResolve, url: "https://animefire.io/video/x", want: "",
		seams:  resolveSeams{extractErr: map[string]error{"https://animefire.io/video/x": errors.New("blocked")}},
		wantEx: []string{"https://animefire.io/video/x"}},
	{name: "resolve error discards a partial answer", route: routeResolve, url: "https://animefire.io/video/x", want: "",
		seams:  resolveSeams{extractOut: map[string]string{"https://animefire.io/video/x": "https://cdn/partial.mp4"}, extractErr: map[string]error{"https://animefire.io/video/x": errors.New("cut off")}},
		wantEx: []string{"https://animefire.io/video/x"}},
	{name: "resolve returns nothing", route: routeResolve, url: "https://animefire.io/video/x", want: "",
		wantEx: []string{"https://animefire.io/video/x"}},
	{name: "legacy: the episode page first", route: routeLegacy, url: "id9", eps: epList("page1", "page2"), num: 2, want: "https://cdn/p2.mp4",
		seams:  resolveSeams{promptOut: map[string]string{"page2": "https://cdn/p2.mp4"}},
		wantPr: []string{"page2"}},
	{name: "legacy: falls back to the original URL", route: routeLegacy, url: "id9", eps: epList("page1"), num: 1, want: "https://cdn/orig.mp4",
		seams:  resolveSeams{promptErr: map[string]error{"page1": errors.New("x")}, promptOut: map[string]string{"id9": "https://cdn/orig.mp4"}},
		wantPr: []string{"page1", "id9"}},
	{name: "legacy: an unknown episode goes straight to the URL", route: routeLegacy, url: "id9", eps: epList("page1"), num: 5, want: "https://cdn/orig.mp4",
		seams:  resolveSeams{promptOut: map[string]string{"id9": "https://cdn/orig.mp4"}},
		wantPr: []string{"id9"}},
	{name: "legacy: no episode number goes straight to the URL", route: routeLegacy, url: "id9", eps: epList("page1"), num: 0, want: "https://cdn/orig.mp4",
		seams:  resolveSeams{promptOut: map[string]string{"id9": "https://cdn/orig.mp4"}},
		wantPr: []string{"id9"}},
	{name: "legacy: no episodes goes straight to the URL", route: routeLegacy, url: "id9", num: 3, want: "https://cdn/orig.mp4",
		seams:  resolveSeams{promptOut: map[string]string{"id9": "https://cdn/orig.mp4"}},
		wantPr: []string{"id9"}},
	{name: "legacy: back at the episode page", route: routeLegacy, url: "id9", eps: epList("page1"), num: 1, wantBack: true,
		seams:  resolveSeams{promptErr: map[string]error{"page1": ErrBackRequested}},
		wantPr: []string{"page1"}},
	{name: "legacy: back at the original URL", route: routeLegacy, url: "id9", eps: epList("page1"), num: 1, wantBack: true,
		seams:  resolveSeams{promptOut: map[string]string{}, promptErr: map[string]error{"id9": fmt.Errorf("w: %w", ErrBackRequested)}},
		wantPr: []string{"page1", "id9"}},
	{name: "legacy: an error discards a partial answer", route: routeLegacy, url: "id9", num: 0, want: "",
		seams:  resolveSeams{promptOut: map[string]string{"id9": "https://cdn/partial.mp4"}, promptErr: map[string]error{"id9": errors.New("cut off")}},
		wantPr: []string{"id9"}},
	{name: "legacy: episode number zero never selects an episode", route: routeLegacy, url: "id9",
		eps: []models.Episode{{Num: 0, URL: "page0"}}, num: 0, want: "https://cdn/orig.mp4",
		seams:  resolveSeams{promptOut: map[string]string{"id9": "https://cdn/orig.mp4", "page0": "https://cdn/wrong.mp4"}},
		wantPr: []string{"id9"}},
	{name: "legacy: nothing works", route: routeLegacy, url: "id9", eps: epList("page1"), num: 1, want: "",
		wantPr: []string{"page1", "id9"}},
	{name: "none", route: routeNone, want: ""},
}

func resolveURLFailures(impl func(playbackRoute, string, []models.Episode, int) (string, bool)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range resolveCases2 {
		s := tc.seams
		oldE, oldP := extractActualVideoURLFn, extractWithPromptFn
		extractActualVideoURLFn = func(u string) (string, error) {
			s.extracted = append(s.extracted, u)
			return s.extractOut[u], s.extractErr[u]
		}
		extractWithPromptFn = func(u string) (string, error) {
			s.prompted = append(s.prompted, u)
			return s.promptOut[u], s.promptErr[u]
		}
		got, back := impl(tc.route, tc.url, tc.eps, tc.num)
		extractActualVideoURLFn, extractWithPromptFn = oldE, oldP
		if got != tc.want || back != tc.wantBack {
			f = append(f, fmt.Sprintf("%s: got (%q,%v) want (%q,%v)", tc.name, got, back, tc.want, tc.wantBack))
		}
		if !slices.Equal(s.extracted, tc.wantEx) || !slices.Equal(s.prompted, tc.wantPr) {
			f = append(f, fmt.Sprintf("%s: asked extract=%v prompt=%v", tc.name, s.extracted, s.prompted))
		}
	}
	return f
}

func TestResolvePlaybackURL(t *testing.T) {
	assert.Empty(t, resolveURLFailures(resolvePlaybackURL))
}

func TestResolvePlaybackURLMeta(t *testing.T) {
	require.Empty(t, resolveURLFailures(resolvePlaybackURL))
	type fn = func(playbackRoute, string, []models.Episode, int) (string, bool)
	mutants := map[string]fn{
		"back becomes failure": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			got, back := resolvePlaybackURL(r, u, e, n)
			return got, false && back
		},
		"failure becomes back": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			got, back := resolvePlaybackURL(r, u, e, n)
			return got, back || got == ""
		},
		"direct is re-extracted": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			if r == routeDirect {
				r = routeResolve
			}
			return resolvePlaybackURL(r, u, e, n)
		},
		"hls is re-extracted": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			if r == routeHLS {
				r = routeResolve
			}
			return resolvePlaybackURL(r, u, e, n)
		},
		"episode page skipped": func(r playbackRoute, u string, _ []models.Episode, _ int) (string, bool) {
			return resolvePlaybackURL(r, u, nil, 0)
		},
		"fallback skipped": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			if r != routeLegacy {
				return resolvePlaybackURL(r, u, e, n)
			}
			if len(e) > 0 && n > 0 {
				if ep, ok := findEpisode(e, n); ok {
					got, err := extractWithPromptFn(ep.URL)
					if errors.Is(err, ErrBackRequested) {
						return "", true
					}
					if err == nil {
						return got, false
					}
				}
			}
			return "", false
		},
		"resolve error ignored": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			if r == routeResolve {
				got, _ := extractActualVideoURLFn(u)
				return got, false
			}
			return resolvePlaybackURL(r, u, e, n)
		},
		"empty answer accepted": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			got, back := resolvePlaybackURL(r, u, e, n)
			if got == "" && !back && r == routeLegacy {
				return u, false
			}
			return got, back
		},
		"none plays the URL": func(r playbackRoute, u string, e []models.Episode, n int) (string, bool) {
			if r == routeNone {
				return "x", false
			}
			return resolvePlaybackURL(r, u, e, n)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, resolveURLFailures(m), "mutant %q survived", name)
	}
}

// ---- HandleDownloadAndPlay, end to end over mock steps -----------------------------------------

type flowMock struct {
	options []int // successive answers of the options menu; the last repeats
	asked   int

	downloadErr  []error // successive results of downloading one episode
	downloads    int
	downloadArgs []string

	rangeErr, batchErr, allErr, playErr []error
	rangeStarts                         []int
	batches, alls, plays                int
	playedURLs                          []string

	upscaleErr  error
	upscales    int
	menuReqs    []EpisodeMenuRequest
	extractOut  map[string]string
	promptErr   error
	promptCalls int
}

func nextErr(seq []error, i int) error {
	if len(seq) == 0 {
		return nil
	}
	return seq[min(i, len(seq)-1)]
}

func (f *flowMock) install(t *testing.T) {
	t.Helper()
	old := struct {
		menu, dl, rng, batch, all, up, play, req, ex, pr any
	}{downloadMenuOptionFn, downloadAndPlayEpisodeFn, downloadRangeFromFn, handleBatchDownloadFn, handleDownloadAllFn,
		upscaleMenuFn, playVideoFn, requestEpisodeMenuFn, extractActualVideoURLFn, extractWithPromptFn}
	oldTTY, oldConfirm, oldReq := stdinIsTerminal, confirmFn, currentMenuRequest
	oldSeason, oldName := gMedia.animeSeason, gMedia.animeName
	t.Cleanup(func() {
		downloadMenuOptionFn, downloadAndPlayEpisodeFn = old.menu.(func(int, int, string) int), old.dl.(func(string, []models.Episode, int, string, string, int, int, *discord.RichPresenceUpdater) error)
		downloadRangeFromFn, handleBatchDownloadFn = old.rng.(func(int, []models.Episode, *models.Anime) error), old.batch.(func([]models.Episode, *models.Anime) error)
		handleDownloadAllFn, upscaleMenuFn = old.all.(func([]models.Episode, *models.Anime) error), old.up.(func() error)
		playVideoFn, requestEpisodeMenuFn = old.play.(func(string, []models.Episode, int, int, int, *discord.RichPresenceUpdater) error), old.req.(func(EpisodeMenuRequest))
		extractActualVideoURLFn, extractWithPromptFn = old.ex.(func(string) (string, error)), old.pr.(func(string) (string, error))
		stdinIsTerminal, confirmFn, currentMenuRequest = oldTTY, oldConfirm, oldReq
		SetAnimeName(oldName, oldSeason)
	})

	downloadMenuOptionFn = func(int, int, string) int {
		i := min(f.asked, len(f.options)-1)
		f.asked++
		return f.options[i]
	}
	downloadAndPlayEpisodeFn = func(url string, _ []models.Episode, n int, animeURL, numStr string, mal, anilist int, _ *discord.RichPresenceUpdater) error {
		f.downloadArgs = append(f.downloadArgs, fmt.Sprintf("%s|%d|%s|%s|%d|%d", url, n, animeURL, numStr, mal, anilist))
		err := nextErr(f.downloadErr, f.downloads)
		f.downloads++
		return err
	}
	downloadRangeFromFn = func(start int, _ []models.Episode, _ *models.Anime) error {
		f.rangeStarts = append(f.rangeStarts, start)
		return nextErr(f.rangeErr, len(f.rangeStarts)-1)
	}
	handleBatchDownloadFn = func([]models.Episode, *models.Anime) error {
		f.batches++
		return nextErr(f.batchErr, f.batches-1)
	}
	handleDownloadAllFn = func([]models.Episode, *models.Anime) error {
		f.alls++
		return nextErr(f.allErr, f.alls-1)
	}
	upscaleMenuFn = func() error { f.upscales++; return f.upscaleErr }
	playVideoFn = func(url string, _ []models.Episode, _, _, _ int, _ *discord.RichPresenceUpdater) error {
		f.playedURLs = append(f.playedURLs, url)
		err := nextErr(f.playErr, f.plays)
		f.plays++
		return err
	}
	requestEpisodeMenuFn = func(r EpisodeMenuRequest) { f.menuReqs = append(f.menuReqs, r) }
	extractActualVideoURLFn = func(u string) (string, error) {
		if v, ok := f.extractOut[u]; ok {
			return v, nil
		}
		return "", errors.New("unresolved")
	}
	extractWithPromptFn = func(string) (string, error) { f.promptCalls++; return "", f.promptErr }
	stdinIsTerminal = func() bool { return true }
	confirmFn = func(string, string, string, string) (bool, error) { return true, nil } // never the real prompt
	currentMenuRequest = EpisodeMenuRequest{}
}

func runFlow(f *flowMock, videoURL string, eps []models.Episode) error {
	return HandleDownloadAndPlay(videoURL, eps, 3, "https://x/anime", "Episode 3", 21, 20, nil, "Show", 2, &models.Anime{Source: "HiAnime", Audio: "Sub"})
}

func TestHandleDownloadAndPlayWatchesByDefault(t *testing.T) {
	f := &flowMock{options: []int{downloadWatch}}
	f.install(t)
	require.NoError(t, runFlow(f, "https://cdn.test/master.m3u8", epList("a", "b", "c")))
	assert.Equal(t, []string{"https://cdn.test/master.m3u8"}, f.playedURLs)
	assert.Equal(t, 1, f.asked)
	assert.Zero(t, f.downloads+f.batches+f.alls+f.upscales)
}

func TestHandleDownloadAndPlayRecordsTheShowContext(t *testing.T) {
	f := &flowMock{options: []int{downloadWatch}}
	f.install(t)
	prev := util.GlobalDownloadRequest
	util.GlobalDownloadRequest = nil
	t.Cleanup(func() { util.GlobalDownloadRequest = prev; setLastAnimeURL("") })
	SetMediaContext("", "", "")
	t.Cleanup(func() { SetMediaContext("", "", "") })

	require.NoError(t, runFlow(f, "https://cdn.test/e.mp4", epList("a")))

	assert.Equal(t, "https://x/anime", getLastAnimeURL(), "remembered for episode switching when Discord is off")
	assert.Equal(t, 2, GetAnimeSeason())
	ctx := snapshotMediaContext()
	assert.Equal(t, "Show", ctx.animeTitle)
	assert.Equal(t, "HiAnime", ctx.source)
	assert.Equal(t, "Sub", ctx.audio)
}

func TestHandleDownloadAndPlaySeasonRules(t *testing.T) {
	prev := util.GlobalDownloadRequest
	t.Cleanup(func() { util.GlobalDownloadRequest = prev })

	for _, tc := range []struct {
		name   string
		season int
		req    *util.DownloadRequest
		want   int
	}{
		{"given season", 3, nil, 3},
		{"zero becomes one", 0, nil, 1},
		{"negative becomes one", -2, nil, 1},
		{"a download request's season wins", 2, &util.DownloadRequest{SeasonNum: 4}, 4},
		{"a request without a season does not", 2, &util.DownloadRequest{}, 2},
		{"a request for season 1 wins too", 3, &util.DownloadRequest{SeasonNum: 1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &flowMock{options: []int{downloadWatch}}
			f.install(t)
			util.GlobalDownloadRequest = tc.req
			require.NoError(t, HandleDownloadAndPlay("https://cdn.test/e.mp4", epList("a"), 1, "u", "1", 0, 0, nil, "Show", tc.season, nil))
			assert.Equal(t, tc.want, GetAnimeSeason())
		})
	}
}

func TestHandleDownloadAndPlayWithoutAnAnimeLeavesSourceAndAudioBlank(t *testing.T) {
	f := &flowMock{options: []int{downloadWatch}}
	f.install(t)
	t.Cleanup(func() { SetMediaContext("", "", "") })
	require.NoError(t, HandleDownloadAndPlay("https://cdn.test/e.mp4", epList("a"), 1, "u", "1", 0, 0, nil, "Show", 1, nil))
	ctx := snapshotMediaContext()
	assert.Equal(t, "Show", ctx.animeTitle)
	assert.Empty(t, ctx.source)
	assert.Empty(t, ctx.audio)
}

func TestHandleDownloadAndPlayOptions(t *testing.T) {
	boom := errors.New("disk full")
	tests := []struct {
		name  string
		f     flowMock
		eps   []models.Episode
		start int // currentMenuRequest.RangeStart
		check func(*testing.T, *flowMock, error)
	}{
		{"back returns to server selection", flowMock{options: []int{downloadBack}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				assert.ErrorIs(t, err, ErrBackToEpisodeSelection)
				assert.Zero(t, f.plays)
			}},
		{"download this episode passes everything on", flowMock{options: []int{downloadThis}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, []string{"https://cdn.test/e.mp4|3|https://x/anime|Episode 3|21|20"}, f.downloadArgs)
				assert.Zero(t, f.plays, "downloading does not also play")
			}},
		{"a download failure is returned", flowMock{options: []int{downloadThis}, downloadErr: []error{boom}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				assert.ErrorIs(t, err, boom)
				assert.Empty(t, f.menuReqs)
			}},
		{"back from the download reopens the options", flowMock{options: []int{downloadThis, downloadWatch}, downloadErr: []error{ErrBackToDownloadOptions}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, []EpisodeMenuRequest{{ShowMenu: true}}, f.menuReqs)
				assert.Equal(t, 2, f.asked)
				assert.Equal(t, 1, f.plays, "then the user watches")
			}},
		{"a range preset to start at episode 1", flowMock{options: []int{downloadRange}}, epList("a"), 1,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, []int{1}, f.rangeStarts, "1 is a valid start, not 'unset'")
				assert.Zero(t, f.batches)
			}},
		{"a range with a preset start", flowMock{options: []int{downloadRange}}, epList("a"), 7,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, []int{7}, f.rangeStarts)
				assert.Zero(t, f.batches)
			}},
		{"a range without a start asks for it", flowMock{options: []int{downloadRange}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, f.batches)
				assert.Empty(t, f.rangeStarts)
			}},
		{"a range failure is returned", flowMock{options: []int{downloadRange}, rangeErr: []error{boom}}, epList("a"), 4,
			func(t *testing.T, f *flowMock, err error) { assert.ErrorIs(t, err, boom) }},
		{"a batch failure is returned", flowMock{options: []int{downloadRange}, batchErr: []error{boom}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) { assert.ErrorIs(t, err, boom) }},
		{"upscale returns to the menu", flowMock{options: []int{downloadUpscale, downloadWatch}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, f.upscales)
				assert.Equal(t, 2, f.asked)
				assert.Equal(t, 1, f.plays)
			}},
		{"an upscale failure does not end the session", flowMock{options: []int{downloadUpscale, downloadWatch}, upscaleErr: boom}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, f.plays)
			}},
		{"download all, confirmed", flowMock{options: []int{downloadAll}}, epList("a", "b"), 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, f.alls)
			}},
		{"download all with nothing to download asks again", flowMock{options: []int{downloadAll, downloadWatch}}, nil, 0,
			func(t *testing.T, f *flowMock, err error) {
				require.NoError(t, err)
				assert.Zero(t, f.alls)
				assert.Equal(t, 2, f.asked)
			}},
		{"quitting during download all is not an error", flowMock{options: []int{downloadAll}, allErr: []error{ErrUserQuit}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) { assert.NoError(t, err) }},
		{"a download-all failure is returned", flowMock{options: []int{downloadAll}, allErr: []error{boom}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) { assert.ErrorIs(t, err, boom) }},
		{"a playback failure is returned", flowMock{options: []int{downloadWatch}, playErr: []error{boom}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) { assert.ErrorIs(t, err, boom) }},
		{"back from playback reopens the options", flowMock{options: []int{downloadWatch, downloadBack}, playErr: []error{ErrBackToDownloadOptions}}, epList("a"), 0,
			func(t *testing.T, f *flowMock, err error) {
				assert.ErrorIs(t, err, ErrBackToEpisodeSelection)
				assert.Equal(t, []EpisodeMenuRequest{{ShowMenu: true}}, f.menuReqs)
			}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f
			f.install(t)
			currentMenuRequest = EpisodeMenuRequest{RangeStart: tc.start}
			err := runFlow(&f, "https://cdn.test/e.mp4", tc.eps)
			tc.check(t, &f, err)
		})
	}
}

func TestHandleDownloadAndPlayDeclinedDownloadAllAsksAgain(t *testing.T) {
	f := &flowMock{options: []int{downloadAll, downloadWatch}}
	f.install(t)
	confirmFn = func(string, string, string, string) (bool, error) { return false, nil }
	require.NoError(t, runFlow(f, "https://cdn.test/e.mp4", epList("a", "b")))
	assert.Zero(t, f.alls, "nothing is downloaded without the yes")
	assert.Equal(t, 2, f.asked)
	assert.Equal(t, 1, f.plays)
}

func TestHandleDownloadAndPlayPlaybackRouting(t *testing.T) {
	t.Run("an intermediate page is resolved before playing", func(t *testing.T) {
		f := &flowMock{options: []int{downloadWatch}, extractOut: map[string]string{"https://animefire.io/video/x": "https://cdn/720.mp4"}}
		f.install(t)
		require.NoError(t, runFlow(f, "https://animefire.io/video/x", epList("a")))
		assert.Equal(t, []string{"https://cdn/720.mp4"}, f.playedURLs)
	})
	t.Run("an unresolvable page is an error, not a silent exit", func(t *testing.T) {
		f := &flowMock{options: []int{downloadWatch}}
		f.install(t)
		err := runFlow(f, "https://animefire.io/video/x", epList("a"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no valid video URL found")
		assert.Zero(t, f.plays)
	})
	t.Run("an empty URL is an error", func(t *testing.T) {
		f := &flowMock{options: []int{downloadWatch}}
		f.install(t)
		err := runFlow(f, "", epList("a"))
		require.Error(t, err)
		assert.Zero(t, f.plays)
	})
	t.Run("backing out of a quality prompt returns to the options", func(t *testing.T) {
		f := &flowMock{options: []int{downloadWatch, downloadBack}, promptErr: ErrBackRequested}
		f.install(t)
		err := runFlow(f, "episode-id-9", epList("a", "b", "c"))
		assert.ErrorIs(t, err, ErrBackToEpisodeSelection)
		assert.Equal(t, 2, f.asked, "the menu was shown again")
		assert.Zero(t, f.plays)
	})
}

// flowOutcome is the fingerprint of one scripted session, compared between the
// real HandleDownloadAndPlay and its broken variants.
func flowOutcome(t *testing.T, impl func(string, []models.Episode, int, string, string, int, int, *discord.RichPresenceUpdater, string, int, *models.Anime) error) string {
	t.Helper()
	var out []string
	for _, opts := range [][]int{
		{downloadWatch}, {downloadBack}, {downloadThis}, {downloadRange}, {downloadAll},
		{downloadUpscale, downloadWatch}, {downloadAll, downloadWatch},
	} {
		f := &flowMock{options: opts}
		f.install(t)
		currentMenuRequest = EpisodeMenuRequest{}
		if len(opts) > 1 && opts[0] == downloadAll { // this script declines the confirmation
			confirmFn = func(string, string, string, string) (bool, error) { return false, nil }
		}
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v", r)
				}
			}()
			err = impl("https://cdn.test/e.mp4", epList("a", "b"), 1, "u", "1", 0, 0, nil, "Show", 1, nil)
		}()
		out = append(out, fmt.Sprintf("%v=>err=%v plays=%d dl=%d batch=%d all=%d up=%d asked=%d", opts, err, f.plays, f.downloads, f.batches, f.alls, f.upscales, f.asked))
	}
	return strings.Join(out, "\n")
}

func TestHandleDownloadAndPlayMeta(t *testing.T) {
	type fn = func(string, []models.Episode, int, string, string, int, int, *discord.RichPresenceUpdater, string, int, *models.Anime) error
	want := flowOutcome(t, HandleDownloadAndPlay)
	require.Contains(t, want, "plays=1", "the baseline plays")

	mutants := map[string]fn{
		"back plays instead": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			err := HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
			if errors.Is(err, ErrBackToEpisodeSelection) {
				return nil
			}
			return err
		},
		"errors swallowed": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			_ = HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
			return nil
		},
		"never plays": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			old := playVideoFn
			playVideoFn = func(string, []models.Episode, int, int, int, *discord.RichPresenceUpdater) error { return nil }
			defer func() { playVideoFn = old }()
			return HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
		},
		"download also plays": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			err := HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
			if err == nil {
				_ = playVideoFn(u, e, n, m, l, up)
			}
			return err
		},
		"download all unconfirmed": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			old := confirmFn
			confirmFn = func(string, string, string, string) (bool, error) { return true, nil }
			defer func() { confirmFn = old }()
			return HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
		},
		"upscale ends the flow": func(u string, e []models.Episode, n int, a, s string, m, l int, up *discord.RichPresenceUpdater, nm string, se int, an *models.Anime) error {
			old := upscaleMenuFn
			upscaleMenuFn = func() error { return ErrUserQuit }
			defer func() { upscaleMenuFn = old }()
			return HandleDownloadAndPlay(u, e, n, a, s, m, l, up, nm, se, an)
		},
	}
	for name, m := range mutants {
		assert.NotEqual(t, want, flowOutcome(t, m), "mutant %q survived", name)
	}
}
