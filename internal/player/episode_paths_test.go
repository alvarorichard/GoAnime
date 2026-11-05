package player

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	charmlog "charm.land/log/v2"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/v2/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// withHome points the default download directories at a temporary home so
// the anime and movie base directories are distinct and predictable.
func withHome(t *testing.T) (animeBase, movieBase string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	prev := util.GlobalOutputDir
	util.GlobalOutputDir = ""
	t.Cleanup(func() { util.GlobalOutputDir = prev })
	return util.DefaultDownloadDir(), util.DefaultMovieDownloadDir()
}

type pathsFn func(snap mediaSnapshot, animeURL, epStr string, n int) (string, string)

type pathsCase struct {
	name  string
	snap  mediaSnapshot
	url   string
	epStr string
	n     int
	// want builds the expected dir and file from the two base directories.
	want func(anime, movie string) (dir, file string)
}

func pathsCases() []pathsCase {
	season2 := mediaSnapshot{
		AnimeName: "Show", AnimeSeason: 2, MediaType: "tv",
		SeasonMap: []metadata.SeasonMapping{{Season: 2, EpisodeCount: 12}},
	}
	absolute := mediaSnapshot{
		AnimeName: "Show", MediaType: "anime",
		SeasonMap: []metadata.SeasonMapping{
			{Season: 1, StartEp: 1, EndEp: 12, EpisodeCount: 12},
			{Season: 2, StartEp: 13, EndEp: 24, EpisodeCount: 12},
		},
	}
	return []pathsCase{
		{
			name: "an absolute episode maps to its season and local number",
			snap: absolute, url: "u", epStr: "14", n: 14,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 2, nil), util.FormatPlexEpisodePath(a, "Show", 2, 2, nil)
			},
		},
		{
			name: "no name falls back to the url folder under the anime base",
			snap: mediaSnapshot{}, url: "https://animefire.io/animes/foo/1", epStr: "3", n: 3,
			want: func(a, _ string) (string, string) {
				d := filepath.Join(a, DownloadFolderFormatter("https://animefire.io/animes/foo/1"))
				return d, filepath.Join(d, "3.mp4")
			},
		},
		{
			name: "no name but movie or tv uses the movie base",
			snap: mediaSnapshot{IsMovieOrTV: true}, url: "https://x/y", epStr: "1", n: 1,
			want: func(_, m string) (string, string) {
				d := filepath.Join(m, DownloadFolderFormatter("https://x/y"))
				return d, filepath.Join(d, "1.mp4")
			},
		},
		{
			name: "movie uses the flat plex layout under the movie base",
			snap: mediaSnapshot{AnimeName: "Film", IsMovieOrTV: true, MediaType: "movie"}, url: "u", epStr: "1", n: 1,
			want: func(_, m string) (string, string) {
				return util.FormatPlexMovieDir(m, "Film", nil), util.FormatPlexMoviePath(m, "Film", "", nil)
			},
		},
		{
			name: "anime episode uses the season layout under the anime base",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "4", n: 4,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 4, nil)
			},
		},
		{
			name: "tv episode goes under the movie base",
			snap: mediaSnapshot{AnimeName: "Show", IsMovieOrTV: true, MediaType: "tv"}, url: "u", epStr: "2", n: 2,
			want: func(_, m string) (string, string) {
				return util.FormatPlexEpisodeDir(m, "Show", 1, nil), util.FormatPlexEpisodePath(m, "Show", 1, 2, nil)
			},
		},
		{
			name: "the int episode wins over the string",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "9", n: 4,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 4, nil)
			},
		},
		{
			name: "episode one as an int is kept, not re-parsed",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "9", n: 1,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 1, nil)
			},
		},
		{
			name: "no int episode parses the string",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "7", n: 0,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 7, nil)
			},
		},
		{
			name: "an unparsable string falls back to episode one",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "special", n: 0,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 1, nil)
			},
		},
		{
			name: "a zero or negative string falls back to episode one",
			snap: mediaSnapshot{AnimeName: "Show", MediaType: "anime"}, url: "u", epStr: "-3", n: -1,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 1, nil), util.FormatPlexEpisodePath(a, "Show", 1, 1, nil)
			},
		},
		{
			name: "a selected later season keeps its season and local episode",
			snap: season2, url: "u", epStr: "3", n: 3,
			want: func(a, _ string) (string, string) {
				return util.FormatPlexEpisodeDir(a, "Show", 2, nil), util.FormatPlexEpisodePath(a, "Show", 2, 3, nil)
			},
		},
	}
}

// pathsFailures runs the table against fn and reports every violation.
func pathsFailures(t *testing.T, fn pathsFn) []string {
	t.Helper()
	anime, movie := withHome(t)
	var bad []string
	for _, c := range pathsCases() {
		wantDir, wantFile := c.want(anime, movie)
		dir, file := fn(c.snap, c.url, c.epStr, c.n)
		if dir != wantDir {
			bad = append(bad, fmt.Sprintf("%s: dir %q, want %q", c.name, dir, wantDir))
		}
		if file != wantFile {
			bad = append(bad, fmt.Sprintf("%s: file %q, want %q", c.name, file, wantFile))
		}
	}
	return bad
}

func TestEpisodeDownloadPaths(t *testing.T) {
	if bad := pathsFailures(t, episodeDownloadPaths); len(bad) > 0 {
		t.Fatalf("violations:\n%s", strings.Join(bad, "\n"))
	}
}

func TestEpisodeDownloadPaths_DirsAreDistinctAndFileIsInsideDir(t *testing.T) {
	anime, movie := withHome(t)
	if anime == movie {
		t.Fatalf("anime and movie bases must differ, both %q", anime)
	}
	for _, c := range pathsCases() {
		dir, file := episodeDownloadPaths(c.snap, c.url, c.epStr, c.n)
		if filepath.Dir(file) != dir {
			t.Errorf("%s: file %q is not directly inside dir %q", c.name, file, dir)
		}
	}
}

func TestEpisodeDownloadPaths_OutputDirOverridesBothBases(t *testing.T) {
	withHome(t)
	util.GlobalOutputDir = t.TempDir()
	dir, _ := episodeDownloadPaths(mediaSnapshot{IsMovieOrTV: true}, "u", "1", 1)
	if !strings.HasPrefix(dir, util.GlobalOutputDir) {
		t.Fatalf("dir %q is not under -o %q", dir, util.GlobalOutputDir)
	}
}

func TestMeta_EpisodeDownloadPaths_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]pathsFn{
		"always the anime base": func(s mediaSnapshot, u, e string, n int) (string, string) {
			s.IsMovieOrTV = false
			return episodeDownloadPaths(s, u, e, n)
		},
		"ignores the series number string": func(s mediaSnapshot, u, e string, n int) (string, string) {
			return episodeDownloadPaths(s, u, "", n)
		},
		"never uses the int episode": func(s mediaSnapshot, u, e string, _ int) (string, string) {
			return episodeDownloadPaths(s, u, e, 0)
		},
		"treats every named media as a movie": func(s mediaSnapshot, u, e string, n int) (string, string) {
			s.MediaType = "movie"
			return episodeDownloadPaths(s, u, e, n)
		},
		"drops the season map": func(s mediaSnapshot, u, e string, n int) (string, string) {
			s.SeasonMap = nil
			return episodeDownloadPaths(s, u, e, n)
		},
		"file and dir swapped": func(s mediaSnapshot, u, e string, n int) (string, string) {
			d, f := episodeDownloadPaths(s, u, e, n)
			return f, d
		},
		"drops the anime name": func(s mediaSnapshot, u, e string, n int) (string, string) {
			s.AnimeName = ""
			return episodeDownloadPaths(s, u, e, n)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(pathsFailures(t, fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- downloadKindFor ----

var kindCases = []struct {
	url  string
	want downloadKind
}{
	{"http://127.0.0.1:5000/blogger_proxy?x=1", downloadKindBlogger},
	{"http://127.0.0.1:5000/other", downloadKindDirect},
	{"http://example.com/blogger_proxy", downloadKindDirect},
	{"https://www.blogger.com/video.g?token=1", downloadKindStream},
	{"https://cdn.example.com/a/master.m3u8", downloadKindStream},
	{"https://cdn.example.com/hls/seg", downloadKindStream},
	{"https://x.wixmp.com/v.mp4", downloadKindStream},
	{"https://x.sharepoint.com/v.aspx", downloadKindStream},
	{"https://cdn.example.com/v.mp4", downloadKindDirect},
}

func kindFailures(fn func(string) downloadKind) []string {
	var bad []string
	for _, c := range kindCases {
		if got := fn(c.url); got != c.want {
			bad = append(bad, fmt.Sprintf("%s: got %d want %d", c.url, got, c.want))
		}
	}
	return bad
}

func TestDownloadKindFor(t *testing.T) {
	if bad := kindFailures(downloadKindFor); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_DownloadKindFor_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]func(string) downloadKind{
		"never blogger": func(u string) downloadKind {
			k := downloadKindFor(u)
			if k == downloadKindBlogger {
				return downloadKindStream
			}
			return k
		},
		"never stream": func(u string) downloadKind {
			k := downloadKindFor(u)
			if k == downloadKindStream {
				return downloadKindDirect
			}
			return k
		},
		"always direct": func(string) downloadKind { return downloadKindDirect },
		"blogger needs only 127.0.0.1": func(u string) downloadKind {
			if strings.Contains(u, "127.0.0.1") {
				return downloadKindBlogger
			}
			return downloadKindFor(u)
		},
		"no sharepoint": func(u string) downloadKind {
			if strings.Contains(u, "sharepoint.com") {
				return downloadKindDirect
			}
			return downloadKindFor(u)
		},
		"no wixmp": func(u string) downloadKind {
			if strings.Contains(u, "wixmp.com") {
				return downloadKindDirect
			}
			return downloadKindFor(u)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(kindFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- downloadStreamWithFallbacks ----

type engineLog struct{ calls []string }

type engineFn func(url, path string, m *model) error

// fakeEngines records the order engines are called in. A failing engine also
// leaves progress behind, so the test can tell whether it was reset.
func fakeEngines(log *engineLog, results map[string]error) streamEngines {
	mk := func(name string) engineFn {
		return func(_, _ string, m *model) error {
			log.calls = append(log.calls, name)
			err := results[name]
			if err != nil {
				m.mu.Lock()
				m.received, m.totalBytes, m.peakPct = 55, 99, 0.5
				m.mu.Unlock()
			}
			return err
		}
	}
	return streamEngines{nativeHLS: mk("hls"), ytdlp: mk("ytdlp"), directHTTP: mk("http")}
}

type chainCase struct {
	name    string
	url     string
	results map[string]error
	calls   string
	wantErr error // nil = success; errAny = any error
}

var errBoom = errors.New("boom")

func chainCases() []chainCase {
	return []chainCase{
		{name: "plain url goes straight to yt-dlp", url: "https://x/v.mp4", calls: "ytdlp"},
		{name: "plain url yt-dlp error is final", url: "https://x/v.mp4", results: map[string]error{"ytdlp": errBoom}, calls: "ytdlp", wantErr: errBoom},
		{name: "hls success stops at native", url: "https://x/a.m3u8", calls: "hls"},
		{name: "unsafe extension uses native first", url: "https://x/a.aspx", calls: "hls"},
		{name: "separate audio goes to yt-dlp only", url: "https://x/a.m3u8", results: map[string]error{"hls": hls.ErrSeparateAudioTracks}, calls: "hls,ytdlp"},
		{name: "separate audio yt-dlp error is final", url: "https://x/a.m3u8", results: map[string]error{"hls": hls.ErrSeparateAudioTracks, "ytdlp": errBoom}, calls: "hls,ytdlp", wantErr: errBoom},
		{name: "wrapped separate audio is recognised", url: "https://x/a.m3u8", results: map[string]error{"hls": fmt.Errorf("wrap: %w", hls.ErrSeparateAudioTracks), "ytdlp": errBoom}, calls: "hls,ytdlp", wantErr: errBoom},
		{name: "native failure then yt-dlp rescues", url: "https://x/a.m3u8", results: map[string]error{"hls": errBoom}, calls: "hls,ytdlp"},
		{name: "native and yt-dlp fail then http rescues", url: "https://x/a.m3u8", results: map[string]error{"hls": errBoom, "ytdlp": errBoom}, calls: "hls,ytdlp,http"},
		{name: "everything fails returns the http error", url: "https://x/a.m3u8", results: map[string]error{"hls": errBoom, "ytdlp": errBoom, "http": errors.New("last")}, calls: "hls,ytdlp,http", wantErr: errors.New("last")},
	}
}

type chainFn func(url, path string, m *model, e streamEngines) error

func chainFailures(fn chainFn) []string {
	var bad []string
	for _, c := range chainCases() {
		var log engineLog
		m := newDownloadModel("t", "d", 0)
		err := fn(c.url, "p", m, fakeEngines(&log, c.results))
		if got := strings.Join(log.calls, ","); got != c.calls {
			bad = append(bad, fmt.Sprintf("%s: calls %q, want %q", c.name, got, c.calls))
		}
		switch {
		case c.wantErr == nil && err != nil:
			bad = append(bad, fmt.Sprintf("%s: unexpected error %v", c.name, err))
		case c.wantErr != nil && (err == nil || err.Error() != c.wantErr.Error()):
			bad = append(bad, fmt.Sprintf("%s: error %v, want %v", c.name, err, c.wantErr))
		}
	}
	return bad
}

func TestDownloadStreamWithFallbacks(t *testing.T) {
	setTestLogger(t)
	if bad := chainFailures(downloadStreamWithFallbacks); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

// A failing engine's progress must be cleared before the next engine runs,
// otherwise the bar would start from the wrong place.
func TestDownloadStreamWithFallbacks_ResetsProgressBetweenEngines(t *testing.T) {
	setTestLogger(t)
	seen := map[string][3]float64{}
	m := newDownloadModel("t", "d", 0)
	spy := func(name string, err error) engineFn {
		return func(_, _ string, m *model) error {
			m.mu.Lock()
			seen[name] = [3]float64{float64(m.received), float64(m.totalBytes), m.peakPct}
			m.received, m.totalBytes, m.peakPct = 55, 99, 0.5
			m.mu.Unlock()
			return err
		}
	}
	e := streamEngines{
		nativeHLS:  spy("hls", errBoom),
		ytdlp:      spy("ytdlp", errBoom),
		directHTTP: spy("http", nil),
	}
	if err := downloadStreamWithFallbacks("https://x/a.m3u8", "p", m, e); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ytdlp"} {
		if seen[name] != [3]float64{} {
			t.Errorf("%s started with leftover progress %v", name, seen[name])
		}
	}
}

func TestMeta_DownloadStreamWithFallbacks_CatchesBrokenImplementations(t *testing.T) {
	setTestLogger(t)
	mutants := map[string]chainFn{
		"always yt-dlp": func(u, p string, m *model, e streamEngines) error { return e.ytdlp(u, p, m) },
		"never http fallback": func(u, p string, m *model, e streamEngines) error {
			if err := downloadStreamWithFallbacks(u, p, m, streamEngines{nativeHLS: e.nativeHLS, ytdlp: e.ytdlp, directHTTP: func(string, string, *model) error { return errBoom }}); err != nil {
				return errBoom
			}
			return nil
		},
		"swallows the native error": func(u, p string, m *model, e streamEngines) error {
			_ = e.nativeHLS(u, p, m)
			return nil
		},
		"ignores the unsafe extension": func(u, p string, m *model, e streamEngines) error {
			if hasUnsafeExtension(u) && !LooksLikeHLS(u) {
				return e.ytdlp(u, p, m)
			}
			return downloadStreamWithFallbacks(u, p, m, e)
		},
		"separate audio tries http first": func(u, p string, m *model, e streamEngines) error {
			err := e.nativeHLS(u, p, m)
			if err == nil {
				return nil
			}
			return e.directHTTP(u, p, m)
		},
		"returns the yt-dlp error instead of the http one": func(u, p string, m *model, e streamEngines) error {
			err := downloadStreamWithFallbacks(u, p, m, e)
			if err != nil {
				return errors.New("ytdlp")
			}
			return nil
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(chainFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- size checks ----

func TestEpisodeTooSmall(t *testing.T) {
	for size, want := range map[int64]bool{0: true, 1: true, 1023: true, 1024: false, 1025: false, 1 << 30: false} {
		if got := episodeTooSmall(size); got != want {
			t.Errorf("episodeTooSmall(%d) = %v, want %v", size, got, want)
		}
	}
}

type sizeFn func(path string) error

func sizeFailures(t *testing.T, fn sizeFn) []string {
	t.Helper()
	dir := t.TempDir()
	mk := func(name string, size int64) string {
		p := filepath.Join(dir, name)
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(size); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		return p
	}
	var bad []string
	if err := fn(filepath.Join(dir, "missing.mp4")); err == nil || !strings.Contains(err.Error(), "not created") {
		bad = append(bad, fmt.Sprintf("missing file: %v", err))
	}
	small := mk("small.mp4", minBloggerEpisodeSize-1)
	if err := fn(small); err == nil || !strings.Contains(err.Error(), "incomplete") {
		bad = append(bad, fmt.Sprintf("one byte short: %v", err))
	}
	if _, err := os.Stat(small); err == nil {
		bad = append(bad, "truncated file was not removed")
	}
	exact := mk("exact.mp4", minBloggerEpisodeSize)
	if err := fn(exact); err != nil {
		bad = append(bad, fmt.Sprintf("exact minimum rejected: %v", err))
	}
	if _, err := os.Stat(exact); err != nil {
		bad = append(bad, "valid file was removed")
	}
	big := mk("big.mp4", minBloggerEpisodeSize*3)
	if err := fn(big); err != nil {
		bad = append(bad, fmt.Sprintf("large file rejected: %v", err))
	}
	return bad
}

func TestCheckBloggerEpisodeSize(t *testing.T) {
	if bad := sizeFailures(t, checkBloggerEpisodeSize); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_CheckBloggerEpisodeSize_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]sizeFn{
		"accepts missing files": func(p string) error {
			if err := checkBloggerEpisodeSize(p); err != nil && strings.Contains(err.Error(), "not created") {
				return nil
			}
			return checkBloggerEpisodeSize(p)
		},
		"off by one (<=)": func(p string) error {
			st, err := os.Stat(p)
			if err == nil && st.Size() == minBloggerEpisodeSize {
				return errors.New("incomplete")
			}
			return checkBloggerEpisodeSize(p)
		},
		"keeps the truncated file": func(p string) error {
			st, err := os.Stat(p)
			if err == nil && st.Size() < minBloggerEpisodeSize {
				return errors.New("download incomplete")
			}
			return checkBloggerEpisodeSize(p)
		},
		"accepts anything":   func(string) error { return nil },
		"rejects everything": func(string) error { return errors.New("incomplete") },
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(sizeFailures(t, fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// setTestLogger gives util.Logger a discarding logger: the program sets it at
// start and the code under test logs through it.
func setTestLogger(t *testing.T) {
	t.Helper()
	prev := util.Logger
	util.Logger = charmlog.New(io.Discard)
	t.Cleanup(func() { util.Logger = prev })
}
