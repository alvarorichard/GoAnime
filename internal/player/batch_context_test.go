package player

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// These tests swap package seams and process globals; none run in parallel.

// fakeResolver stands in for a source's stream resolution: like StartFlix,
// it hands each episode's referer and subtitles over in the globals.
func fakeResolver(t *testing.T, streams map[int]struct {
	url, referer, sub string
}) {
	t.Helper()
	prev := resolveEpisodeStreamFn
	prevRef := util.GetGlobalReferer()
	resolveEpisodeStreamFn = func(ep models.Episode, _ *models.Anime) (string, error) {
		s, ok := streams[ep.Num]
		if !ok {
			return "", errors.New("no stream")
		}
		util.SetGlobalReferer(s.referer)
		if s.sub != "" {
			util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: s.sub, Language: "pt-br", Label: "Português"}})
		}
		return s.url, nil
	}
	t.Cleanup(func() {
		resolveEpisodeStreamFn = prev
		util.SetGlobalReferer(prevRef)
		util.ClearGlobalSubtitles()
	})
}

// writeVideo leaves a file big enough to pass validateDownloadedVideo,
// sparse so it costs nothing.
func writeVideo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Truncate(minDownloadedVideoSize)
}

// TestBatchEpisodesKeepTheirOwnRefererAndSubtitles pins the per-episode
// context: a batch resolves every episode before downloading any, so the
// referer and subtitle globals end up holding the last episode's. Each
// episode must still be fetched with its own referer and muxed with its own
// subtitles — also with downloads running in parallel, as the batch runs
// them. They used to all get the last episode's.
func TestBatchEpisodesKeepTheirOwnRefererAndSubtitles(t *testing.T) {
	restore := installDownloadRangeTestState(t.TempDir())
	defer restore()
	home := t.TempDir()
	streams := map[int]struct{ url, referer, sub string }{
		1: {"https://byse.cdn.test/ep1/master.m3u8", "https://embedplaybyse.top/", "https://subs.test/ep1.vtt"},
		2: {"https://painel.test/ep2/master.m3u8", "https://painel.test/", "https://subs.test/ep2.vtt"},
		3: {"https://abyss.local.test/ep3/master.m3u8", "", ""}, // Abyss: no referer, no subtitles
	}
	fakeResolver(t, streams)

	var mu sync.Mutex
	gotReferer := map[string]string{}
	gotSubs := map[string]string{}
	calls := useHLSChainMocks(t,
		func(u, p string, _ *model) error {
			mu.Lock()
			gotReferer[u] = downloadReferer(u)
			mu.Unlock()
			return writeVideo(p)
		}, ok, ok)
	prevEmbed := embedSubtitlesFn
	embedSubtitlesFn = func(path string, subs []util.SubtitleInfo, _ func(string, ...any)) {
		var urls []string
		for _, s := range subs {
			urls = append(urls, s.URL)
		}
		mu.Lock()
		gotSubs[filepath.Base(path)] = strings.Join(urls, ",")
		mu.Unlock()
	}
	t.Cleanup(func() { embedSubtitlesFn = prevEmbed })

	// First pass, as the batch does it: every episode resolved before any download.
	var eps []*batchEpisode
	for n := 1; n <= 3; n++ {
		be, err := resolveBatchEpisode(models.Episode{Num: n}, &models.Anime{Source: "StartFlix"})
		if err != nil {
			t.Fatal(err)
		}
		be.num, be.path = n, filepath.Join(home, "Show", "ep"+string(rune('0'+n))+".mp4")
		eps = append(eps, &be)
	}
	// Second pass, in parallel.
	var wg sync.WaitGroup
	errs := make([]error, len(eps))
	for i, be := range eps {
		wg.Go(func() { errs[i] = downloadBatchEpisode(be, &models.Anime{Source: "StartFlix"}, nil, nil) })
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("episode %d: %v", i+1, err)
		}
	}
	if calls.native != 3 {
		t.Errorf("native downloads = %d, want 3", calls.native)
	}

	for n, s := range streams {
		// An episode whose resolver gave no referer (Abyss, served locally)
		// falls back to the global one.
		want := s.referer
		if want == "" {
			want = util.GetGlobalReferer()
		}
		if got := gotReferer[s.url]; got != want {
			t.Errorf("episode %d sent Referer %q, want its own %q", n, got, want)
		}
		if got := gotSubs["ep"+string(rune('0'+n))+".mp4"]; got != s.sub {
			t.Errorf("episode %d embedded subtitles %q, want its own %q", n, got, s.sub)
		}
	}
}

// TestFailedBatchDownloadRemovesItsPartialFile: a download that fails after
// writing something must not leave it behind, or every later run would skip
// the episode as "already exists".
func TestFailedBatchDownloadRemovesItsPartialFile(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "Show", "ep1.mp4")
	partial := func(_, p string, _ *model) error {
		_ = os.MkdirAll(filepath.Dir(p), 0o700)
		_ = os.WriteFile(p, []byte("half a segment"), 0o600)
		return errors.New("connection reset")
	}
	useHLSChainMocks(t, partial, partial, partial)
	be := &batchEpisode{num: 1, path: path, url: "https://cdn.test/ep1/master.m3u8"}
	if err := downloadBatchEpisode(be, &models.Anime{}, nil, nil); err == nil {
		t.Fatal("a failed download reported success")
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("the partial file was left on disk")
	}

	// A download that "succeeds" with too little data is removed too.
	useHLSChainMocks(t, func(_, p string, _ *model) error {
		_ = os.MkdirAll(filepath.Dir(p), 0o700)
		return os.WriteFile(p, []byte("#EXTM3U"), 0o600)
	}, ok, ok)
	if err := downloadBatchEpisode(be, &models.Anime{}, nil, nil); err == nil {
		t.Fatal("a 7-byte download passed validation")
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("the undersized file was left on disk")
	}
}

// TestEpisodeAlreadyDownloaded: only a finished file counts; what an earlier
// failed run left behind is removed so the episode is fetched again.
func TestEpisodeAlreadyDownloaded(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "stale.mp4")
	_ = os.WriteFile(stale, []byte("partial"), 0o600)
	if episodeAlreadyDownloaded(stale) {
		t.Error("a partial file counted as downloaded")
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("the partial file was not removed")
	}
	done := filepath.Join(dir, "done.mp4")
	if err := writeVideo(done); err != nil {
		t.Fatal(err)
	}
	if !episodeAlreadyDownloaded(done) {
		t.Error("a finished file did not count as downloaded")
	}
	if episodeAlreadyDownloaded(filepath.Join(dir, "missing.mp4")) {
		t.Error("a missing file counted as downloaded")
	}
}

// TestHandleDownloadAll_CoversEveryListedEpisode: "download all" used to be
// HandleBatchDownloadRange(eps, anime, 1, len(eps)) — for a list [1 2 4 5]
// that looked at 1..4, never tried episode 5 and reported the gap at 3 as a
// failure. Every listed episode is attempted now, and gaps are not failures.
func TestHandleDownloadAll_CoversEveryListedEpisode(t *testing.T) {
	restore := installDownloadRangeTestState(t.TempDir())
	defer restore()
	SetAnimeName("Show", 1)
	SetExactMediaType("tv")
	prev := resolveEpisodeStreamFn
	resolveEpisodeStreamFn = func(models.Episode, *models.Anime) (string, error) {
		return "", errors.New("offline") // stop each episode at resolution: no UI, no network
	}
	t.Cleanup(func() { resolveEpisodeStreamFn = prev })

	var eps []models.Episode
	for _, n := range []int{1, 2, 4, 5} {
		eps = append(eps, models.Episode{Num: n})
	}
	err := HandleDownloadAll(eps, &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/"})
	var be batchDownloadError
	if !errors.As(err, &be) {
		t.Fatalf("err = %v, want the batch's failures", err)
	}
	var attempted []int
	for _, f := range be.Failures {
		attempted = append(attempted, f.Episode)
	}
	sort.Ints(attempted)
	if want := []int{1, 2, 4, 5}; !equalInts(attempted, want) {
		t.Errorf("episodes attempted = %v, want every listed one %v", attempted, want)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEpisodeNumberSpan(t *testing.T) {
	first, last, ok := episodeNumberSpan([]models.Episode{{Num: 4}, {Num: 2}, {Num: 0}, {Num: 9}})
	if !ok || first != 2 || last != 9 {
		t.Errorf("span = %d-%d (%v), want 2-9", first, last, ok)
	}
	if _, _, ok := episodeNumberSpan([]models.Episode{{Num: 0}}); ok {
		t.Error("a list with no numbered episode has a span")
	}
}

// TestKeepChosenSubtitles: the batch's one subtitle choice is applied to each
// episode's own tracks, by label and language.
func TestKeepChosenSubtitles(t *testing.T) {
	pt := util.SubtitleInfo{URL: "https://s/ep2-pt.vtt", Language: "pt-br", Label: "Português"}
	en := util.SubtitleInfo{URL: "https://s/ep2-en.vtt", Language: "en", Label: "English"}
	chosen := []util.SubtitleInfo{{URL: "https://s/ep1-pt.vtt", Language: "pt-br", Label: "Português"}}
	got := keepChosenSubtitles([]util.SubtitleInfo{pt, en}, chosen)
	if len(got) != 1 || got[0].URL != pt.URL {
		t.Errorf("kept %+v, want episode 2's own Portuguese track", got)
	}
	if got := keepChosenSubtitles([]util.SubtitleInfo{en}, nil); len(got) != 0 {
		t.Errorf("kept %+v with nothing chosen", got)
	}
}

// TestBatchReportsEachEpisodeWhilePreparing: the first pass resolves every
// episode before the progress bar appears, seconds each; it used to print
// nothing, so a StartFlix season sat silent for a minute and a half.
func TestBatchReportsEachEpisodeWhilePreparing(t *testing.T) {
	restore := installDownloadRangeTestState(t.TempDir())
	defer restore()
	SetAnimeName("Show", 1)
	SetExactMediaType("tv")
	prevResolve, prevStatus := resolveEpisodeStreamFn, batchStatusf
	var status []string
	batchStatusf = func(format string, a ...any) { status = append(status, strings.TrimSpace(fmt.Sprintf(format, a...))) }
	resolveEpisodeStreamFn = func(models.Episode, *models.Anime) (string, error) { return "", errors.New("offline") }
	t.Cleanup(func() { resolveEpisodeStreamFn, batchStatusf = prevResolve, prevStatus })

	eps := []models.Episode{{Num: 1}, {Num: 2}, {Num: 3}}
	_ = HandleDownloadAll(eps, &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/"})
	want := []string{"Preparing episode 1 (1 of 3)...", "Preparing episode 2 (2 of 3)...", "Preparing episode 3 (3 of 3)...", ""}
	if strings.Join(status, "|") != strings.Join(want, "|") {
		t.Errorf("status = %q, want %q", status, want)
	}
}
