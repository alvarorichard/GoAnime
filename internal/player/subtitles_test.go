package player

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- resolveFFmpeg -------------------------------------------------------------------------

type fakeFileInfo struct {
	os.FileInfo
	dir bool
}

func (f fakeFileInfo) IsDir() bool { return f.dir }

type ffmpegLookup struct {
	lookErr, evalErr, statErr error
	found, resolved           string
	isDir                     bool
}

func (l ffmpegLookup) install(t *testing.T) {
	t.Helper()
	oldLook, oldEval, oldStat := lookPathFn, evalSymlinksFn, statFn
	t.Cleanup(func() { lookPathFn, evalSymlinksFn, statFn = oldLook, oldEval, oldStat })
	lookPathFn = func(name string) (string, error) {
		if name != "ffmpeg" {
			return "", fmt.Errorf("asked for %q", name)
		}
		return l.found, l.lookErr
	}
	evalSymlinksFn = func(string) (string, error) { return l.resolved, l.evalErr }
	statFn = func(string) (os.FileInfo, error) { return fakeFileInfo{dir: l.isDir}, l.statErr }
}

// absBin is an absolute path on every OS (a drive letter on Windows).
func absBin(name string) string {
	root, _ := filepath.Abs(string(filepath.Separator))
	return filepath.Join(root, "usr", "bin", name)
}

var ffmpegCases = []struct {
	name    string
	l       ffmpegLookup
	want    string
	wantErr error
}{
	{"found", ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg")}, absBin("ffmpeg"), nil},
	{"missing", ffmpegLookup{lookErr: errors.New("not in PATH")}, "", errFFmpegMissing},
	{"symlink cannot be resolved", ffmpegLookup{found: "ffmpeg", evalErr: errors.New("loop")}, "", errFFmpegUnresolvable},
	{"resolves to a relative path", ffmpegLookup{found: "ffmpeg", resolved: "bin/ffmpeg"}, "", errFFmpegUnsafe},
	{"resolves to a bare name", ffmpegLookup{found: "ffmpeg", resolved: "ffmpeg"}, "", errFFmpegUnsafe},
	{"is a directory", ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg"), isDir: true}, "", errFFmpegUnsafe},
	{"cannot be stat'ed", ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg"), statErr: os.ErrNotExist}, "", errFFmpegUnsafe},
}

func ffmpegFailures(impl func() (string, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range ffmpegCases {
		oldLook, oldEval, oldStat := lookPathFn, evalSymlinksFn, statFn
		l := tc.l
		lookPathFn = func(string) (string, error) { return l.found, l.lookErr }
		evalSymlinksFn = func(string) (string, error) { return l.resolved, l.evalErr }
		statFn = func(string) (os.FileInfo, error) { return fakeFileInfo{dir: l.isDir}, l.statErr }
		got, err := impl()
		lookPathFn, evalSymlinksFn, statFn = oldLook, oldEval, oldStat
		if got != tc.want || !errors.Is(err, tc.wantErr) || (tc.wantErr == nil && err != nil) {
			f = append(f, fmt.Sprintf("%s: got (%q,%v)", tc.name, got, err))
		}
	}
	return f
}

func TestResolveFFmpeg(t *testing.T) {
	assert.Empty(t, ffmpegFailures(resolveFFmpeg))
}

func TestResolveFFmpegAsksForFFmpegByName(t *testing.T) {
	ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg")}.install(t)
	got, err := resolveFFmpeg()
	require.NoError(t, err)
	assert.Equal(t, absBin("ffmpeg"), got, "the symlink-resolved path is what gets executed, not the PATH entry")
}

func TestResolveFFmpegMeta(t *testing.T) {
	require.Empty(t, ffmpegFailures(resolveFFmpeg))
	mutants := map[string]func() (string, error){
		"accepts relative paths": func() (string, error) {
			p, err := resolveFFmpeg()
			if errors.Is(err, errFFmpegUnsafe) {
				return "bin/ffmpeg", nil
			}
			return p, err
		},
		"accepts directories": func() (string, error) {
			p, err := resolveFFmpeg()
			if errors.Is(err, errFFmpegUnsafe) {
				return absBin("ffmpeg"), nil
			}
			return p, err
		},
		"returns the PATH entry not the resolved one": func() (string, error) {
			_, err := resolveFFmpeg()
			if err != nil {
				return "", err
			}
			p, _ := lookPathFn("ffmpeg")
			return p, nil
		},
		"missing is unsafe": func() (string, error) {
			p, err := resolveFFmpeg()
			if errors.Is(err, errFFmpegMissing) {
				return "", errFFmpegUnsafe
			}
			return p, err
		},
		"unresolvable is missing": func() (string, error) {
			p, err := resolveFFmpeg()
			if errors.Is(err, errFFmpegUnresolvable) {
				return "", errFFmpegMissing
			}
			return p, err
		},
		"always found": func() (string, error) { return absBin("ffmpeg"), nil },
		"never found":  func() (string, error) { return "", errFFmpegMissing },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, ffmpegFailures(m), "mutant %q survived", name)
	}
}

// ---- subtitleExt / subtitleLang -------------------------------------------------------------

var extCases = map[string]string{
	"https://x/sub.srt": "srt", "https://x/sub.SRT": "srt", "https://x/sub.ass?token=1": "ass",
	"https://x/sub.vtt": "vtt", "https://x/sub": "vtt", "": "vtt", "https://x/a.srt.ass": "srt",
}

func extFailures(impl func(string) string) (f []string) {
	for in, want := range extCases {
		if got := impl(in); got != want {
			f = append(f, fmt.Sprintf("%q: got %q want %q", in, got, want))
		}
	}
	return f
}

func TestSubtitleExt(t *testing.T) {
	assert.Empty(t, extFailures(subtitleExt))
}

func TestSubtitleExtMeta(t *testing.T) {
	require.Empty(t, extFailures(subtitleExt))
	mutants := map[string]func(string) string{
		"always vtt": func(string) string { return "vtt" },
		"ass before srt": func(u string) string {
			if strings.Contains(strings.ToLower(u), ".ass") {
				return "ass"
			}
			return subtitleExt(u)
		},
		"srt only": func(u string) string {
			e := subtitleExt(u)
			if e == "ass" {
				return "vtt"
			}
			return e
		},
		"case-sensitive": func(u string) string {
			switch {
			case strings.Contains(u, ".srt"):
				return "srt"
			case strings.Contains(u, ".ass"):
				return "ass"
			}
			return "vtt"
		},
		"suffix only": func(u string) string {
			l := strings.ToLower(u)
			switch {
			case strings.HasSuffix(l, ".srt"):
				return "srt"
			case strings.HasSuffix(l, ".ass"):
				return "ass"
			}
			return "vtt"
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, extFailures(m), "mutant %q survived", name)
	}
}

func TestSubtitleLang(t *testing.T) {
	assert.Equal(t, "English", subtitleLang("English", "en"), "the label wins")
	assert.Equal(t, "pt-BR", subtitleLang("", "pt-BR"), "the language when there is no label")
	assert.Equal(t, "unknown", subtitleLang("", ""))
	assert.Equal(t, "unknown", subtitleLang("///", "..."), "names that sanitise to nothing are unknown")
	assert.NotContains(t, subtitleLang("a/b\\c", ""), "/")
	assert.NotContains(t, subtitleLang("a/b\\c", ""), "\\")
}

func langFailures(impl func(string, string) string) (f []string) {
	for _, tc := range []struct{ label, lang, want string }{
		{"English", "en", "English"}, {"", "pt-BR", "pt-BR"}, {"", "", "unknown"}, {"///", "...", "unknown"},
	} {
		if got := impl(tc.label, tc.lang); got != tc.want {
			f = append(f, fmt.Sprintf("%q/%q: got %q want %q", tc.label, tc.lang, got, tc.want))
		}
	}
	return f
}

func TestSubtitleLangMeta(t *testing.T) {
	require.Empty(t, langFailures(subtitleLang))
	mutants := map[string]func(string, string) string{
		"language first": func(l, g string) string { return subtitleLang(g, l) },
		"label only": func(l, _ string) string {
			if v := util.SanitizeForFilename(l); v != "" {
				return v
			}
			return "unknown"
		},
		"empty name": func(l, g string) string {
			if v := subtitleLang(l, g); v != "unknown" {
				return v
			}
			return ""
		},
		"unsanitised": func(l, g string) string {
			if l != "" {
				return l
			}
			return subtitleLang(l, g)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, langFailures(m), "mutant %q survived", name)
	}
}

// ---- fetchSubtitleFiles ---------------------------------------------------------------------

func subtitleServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var agents []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		agents = append(agents, r.Header.Get("User-Agent"))
		mu.Unlock()
		switch {
		case strings.HasPrefix(r.URL.Path, "/gone"):
			w.WriteHeader(http.StatusNotFound)
		default:
			_, _ = w.Write([]byte("subtitle body for " + r.URL.Path))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &agents
}

func TestFetchSubtitleFiles(t *testing.T) {
	srv, agents := subtitleServer(t)
	dir := t.TempDir()
	subs := []util.SubtitleInfo{
		{URL: srv.URL + "/en.srt", Label: "English", Language: "en"},
		{URL: "", Label: "Skipped", Language: "xx"},
		{URL: srv.URL + "/gone.vtt", Label: "Gone", Language: "gg"},
		{URL: srv.URL + "/pt.ass", Label: "", Language: "pt-BR"},
		{URL: srv.URL + "/x.vtt", Label: "", Language: ""},
	}

	got := fetchSubtitleFiles(srv.Client(), dir, subs)

	require.Len(t, got, 3, "no URL and HTTP 404 are skipped, the rest arrive")
	assert.Equal(t, subtitleEntry{tmpPath: filepath.Join(dir, ".tmp_sub_English.srt"), label: "English", langCode: "en"}, got[0])
	assert.Equal(t, subtitleEntry{tmpPath: filepath.Join(dir, ".tmp_sub_pt-BR.ass"), label: "", langCode: "pt-BR"}, got[1])
	assert.Equal(t, subtitleEntry{tmpPath: filepath.Join(dir, ".tmp_sub_unknown.vtt"), label: "", langCode: "unknown"}, got[2],
		"with no language either, the track is named unknown and so is its language code")
	b, err := os.ReadFile(got[0].tmpPath)
	require.NoError(t, err)
	assert.Equal(t, "subtitle body for /en.srt", string(b))
	assert.NoFileExists(t, filepath.Join(dir, ".tmp_sub_Gone.vtt"), "nothing is written for a failed download")
	assert.Len(t, *agents, 4, "one request per track that has a URL")
	for _, ua := range *agents {
		assert.Contains(t, ua, "Mozilla/5.0", "a browser User-Agent is sent")
	}
}

func TestFetchSubtitleFilesSurvivesFailures(t *testing.T) {
	dir := t.TempDir()
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close() // connection refused

	got := fetchSubtitleFiles(&http.Client{}, dir, []util.SubtitleInfo{
		{URL: down.URL + "/a.srt", Label: "A"},
		{URL: "http://bad host/\x7f", Label: "B"},
	})
	assert.Empty(t, got)

	assert.Empty(t, fetchSubtitleFiles(&http.Client{}, dir, nil))

	srv, _ := subtitleServer(t)
	got = fetchSubtitleFiles(srv.Client(), filepath.Join(dir, "no-such-dir"), []util.SubtitleInfo{{URL: srv.URL + "/en.srt", Label: "E"}})
	assert.Empty(t, got, "a track that cannot be written is skipped")
}

func fetchFailures(impl func(*http.Client, string, []util.SubtitleInfo) []subtitleEntry, srv *httptest.Server) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	dir, _ := os.MkdirTemp("", "subs")
	defer os.RemoveAll(dir)
	got := impl(srv.Client(), dir, []util.SubtitleInfo{
		{URL: srv.URL + "/en.srt", Label: "English", Language: "en"},
		{URL: "", Label: "None"},
		{URL: srv.URL + "/gone.vtt", Label: "Gone"},
		{URL: srv.URL + "/pt.ass", Language: "pt-BR"},
	})
	if len(got) != 2 {
		return append(f, fmt.Sprintf("%d entries want 2", len(got)))
	}
	if got[0].label != "English" || got[0].langCode != "en" || !strings.HasSuffix(got[0].tmpPath, ".tmp_sub_English.srt") {
		f = append(f, fmt.Sprintf("first: %+v", got[0]))
	}
	if got[1].langCode != "pt-BR" || !strings.HasSuffix(got[1].tmpPath, ".tmp_sub_pt-BR.ass") {
		f = append(f, fmt.Sprintf("second: %+v", got[1]))
	}
	if b, err := os.ReadFile(got[0].tmpPath); err != nil || string(b) != "subtitle body for /en.srt" {
		f = append(f, "body not saved")
	}
	return f
}

func TestFetchSubtitleFilesMeta(t *testing.T) {
	srv, _ := subtitleServer(t)
	require.Empty(t, fetchFailures(fetchSubtitleFiles, srv))
	type fn = func(*http.Client, string, []util.SubtitleInfo) []subtitleEntry
	mutants := map[string]fn{
		"keeps failed downloads": func(c *http.Client, d string, s []util.SubtitleInfo) []subtitleEntry {
			e := fetchSubtitleFiles(c, d, s)
			return append(e, subtitleEntry{tmpPath: "x", label: "Gone"})
		},
		"drops the last": func(c *http.Client, d string, s []util.SubtitleInfo) []subtitleEntry {
			e := fetchSubtitleFiles(c, d, s)
			return e[:max(0, len(e)-1)]
		},
		"language code is the label": func(c *http.Client, d string, s []util.SubtitleInfo) []subtitleEntry {
			e := fetchSubtitleFiles(c, d, s)
			for i := range e {
				e[i].langCode = e[i].label
			}
			return e
		},
		"wrong extension": func(c *http.Client, d string, s []util.SubtitleInfo) []subtitleEntry {
			e := fetchSubtitleFiles(c, d, s)
			for i := range e {
				e[i].tmpPath = strings.TrimSuffix(e[i].tmpPath, filepath.Ext(e[i].tmpPath)) + ".vtt"
			}
			return e
		},
		"empty body saved": func(c *http.Client, d string, s []util.SubtitleInfo) []subtitleEntry {
			e := fetchSubtitleFiles(c, d, s)
			for _, x := range e {
				_ = os.WriteFile(x.tmpPath, nil, 0o600)
			}
			return e
		},
		"nothing fetched": func(*http.Client, string, []util.SubtitleInfo) []subtitleEntry { return nil },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, fetchFailures(m, srv), "mutant %q survived", name)
	}
}

// ---- subtitleMuxArgs ------------------------------------------------------------------------

func muxArgsFailures(impl func(string, []subtitleEntry, string, string) []string) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	entries := []subtitleEntry{
		{tmpPath: "/d/.tmp_sub_English.srt", label: "English", langCode: "en"},
		{tmpPath: "/d/.tmp_sub_pt.ass", label: "Português", langCode: "pt-BR"},
	}
	got := impl("/d/ep.mp4", entries, "mov_text", "/d/ep.mp4.muxing.mp4")
	want := []string{
		"-y", "-fflags", "+genpts", "-i", "/d/ep.mp4",
		"-i", "/d/.tmp_sub_English.srt", "-i", "/d/.tmp_sub_pt.ass",
		"-map", "0:v", "-map", "0:a", "-map", "1", "-map", "2",
		"-c:v", "copy", "-c:a", "copy", "-c:s", "mov_text",
		"-metadata:s:s:0", "language=en", "-metadata:s:s:0", "title=English",
		"-metadata:s:s:1", "language=pt-BR", "-metadata:s:s:1", "title=Português",
		"/d/ep.mp4.muxing.mp4",
	}
	if len(got) != len(want) {
		return append(f, fmt.Sprintf("len %d want %d: %v", len(got), len(want), got))
	}
	for i := range want {
		if filepath.ToSlash(got[i]) != want[i] {
			f = append(f, fmt.Sprintf("arg %d: %q want %q", i, got[i], want[i]))
		}
	}
	one := impl("/d/ep.mp4", entries[:1], "srt", "/d/o.mkv")
	if !slices.Contains(one, "-map") || slices.Contains(one, "-metadata:s:s:1") || filepath.ToSlash(one[len(one)-1]) != "/d/o.mkv" {
		f = append(f, "single track command wrong")
	}
	return f
}

func TestSubtitleMuxArgs(t *testing.T) {
	assert.Empty(t, muxArgsFailures(subtitleMuxArgs))
}

func TestSubtitleMuxArgsNeverMapsDataStreams(t *testing.T) {
	args := subtitleMuxArgs("v.mp4", []subtitleEntry{{tmpPath: "s.srt"}}, "mov_text", "o.mp4")
	i := slices.Index(args, "0:v")
	require.GreaterOrEqual(t, i, 1)
	assert.Equal(t, "0:a", args[i+2], "only the video and audio streams are taken from the input")
	assert.NotContains(t, args, "0", "no blanket -map 0: timed_id3 data streams crash muxing")
	assert.Equal(t, "copy", args[slices.Index(args, "-c:v")+1], "video is never re-encoded")
	assert.Equal(t, "copy", args[slices.Index(args, "-c:a")+1], "audio is never re-encoded")
}

func TestSubtitleMuxArgsMeta(t *testing.T) {
	require.Empty(t, muxArgsFailures(subtitleMuxArgs))
	type fn = func(string, []subtitleEntry, string, string) []string
	drop := func(token string) fn {
		return func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			if i := slices.Index(a, token); i >= 0 {
				a = slices.Delete(slices.Clone(a), i, min(i+2, len(a)))
			}
			return a
		}
	}
	mutants := map[string]fn{
		"re-encodes video": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			a[slices.Index(a, "-c:v")+1] = "libx264"
			return a
		},
		"re-encodes audio": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			a[slices.Index(a, "-c:a")+1] = "aac"
			return a
		},
		"maps everything": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			a[slices.Index(a, "0:v")] = "0"
			return a
		},
		"wrong codec":  func(v string, e []subtitleEntry, _, o string) []string { return subtitleMuxArgs(v, e, "srt", o) },
		"no overwrite": drop("-y"),
		"no genpts":    drop("-fflags"),
		"output first": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			return append([]string{a[len(a)-1]}, a[:len(a)-1]...)
		},
		"language dropped": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			for i, x := range a {
				if strings.HasPrefix(x, "language=") {
					a[i] = "language="
				}
			}
			return a
		},
		"titles swapped": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			var idx []int
			for i, x := range a {
				if strings.HasPrefix(x, "title=") {
					idx = append(idx, i)
				}
			}
			if len(idx) == 2 {
				a[idx[0]], a[idx[1]] = a[idx[1]], a[idx[0]]
			}
			return a
		},
		"track not mapped": func(v string, e []subtitleEntry, c, o string) []string {
			a := subtitleMuxArgs(v, e, c, o)
			i := slices.Index(a, "2")
			if i > 0 {
				a = slices.Delete(slices.Clone(a), i-1, i+1)
			}
			return a
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, muxArgsFailures(m), "mutant %q survived", name)
	}
}

// ---- muxSubtitles ---------------------------------------------------------------------------

// ffmpegMock stands in for the encoder: a run succeeds by creating its output
// file, as ffmpeg does, and fails when it is told to.
type ffmpegMock struct {
	failCodecs map[string]bool // fail runs using these subtitle codecs
	runs       [][]string
}

func (f *ffmpegMock) run(args []string) error {
	f.runs = append(f.runs, args)
	codec := args[slices.Index(args, "-c:s")+1]
	out := args[len(args)-1]
	if f.failCodecs[codec] {
		_ = os.WriteFile(out, []byte("partial"), 0o600) // ffmpeg leaves a partial file when it fails
		return errors.New("exit status 1")
	}
	return os.WriteFile(out, []byte("muxed with "+codec), 0o600)
}

func muxFixture(t *testing.T) (video string, entries []subtitleEntry) {
	t.Helper()
	dir := t.TempDir()
	video = filepath.Join(dir, "ep.mp4")
	require.NoError(t, os.WriteFile(video, []byte("original"), 0o600))
	return video, []subtitleEntry{{tmpPath: filepath.Join(dir, "s.srt"), label: "English", langCode: "en"}}
}

func TestMuxSubtitles(t *testing.T) {
	t.Run("mp4 first, replacing the video", func(t *testing.T) {
		video, entries := muxFixture(t)
		f := &ffmpegMock{}
		assert.Equal(t, muxedMP4, muxSubtitles(f.run, video, entries))
		assert.Len(t, f.runs, 1)
		b, _ := os.ReadFile(video)
		assert.Equal(t, "muxed with mov_text", string(b))
		assert.NoFileExists(t, video+".muxing.mp4")
	})
	t.Run("mkv when mp4 fails, leaving the video alone", func(t *testing.T) {
		video, entries := muxFixture(t)
		f := &ffmpegMock{failCodecs: map[string]bool{"mov_text": true}}
		assert.Equal(t, muxedMKV, muxSubtitles(f.run, video, entries))
		assert.Len(t, f.runs, 2)
		mkv := strings.TrimSuffix(video, ".mp4") + ".mkv"
		b, err := os.ReadFile(mkv)
		require.NoError(t, err)
		assert.Equal(t, "muxed with srt", string(b))
		orig, _ := os.ReadFile(video)
		assert.Equal(t, "original", string(orig), "the MKV is a new file; the original is untouched")
		assert.NoFileExists(t, video+".muxing.mp4", "the failed attempt's partial output is removed")
		assert.NoFileExists(t, mkv+".tmp.mkv")
	})
	t.Run("both fail", func(t *testing.T) {
		video, entries := muxFixture(t)
		f := &ffmpegMock{failCodecs: map[string]bool{"mov_text": true, "srt": true}}
		assert.Equal(t, muxFailed, muxSubtitles(f.run, video, entries))
		assert.Len(t, f.runs, 2)
		orig, _ := os.ReadFile(video)
		assert.Equal(t, "original", string(orig))
		assert.NoFileExists(t, video+".muxing.mp4")
		assert.NoFileExists(t, strings.TrimSuffix(video, ".mp4")+".mkv.tmp.mkv")
		assert.NoFileExists(t, strings.TrimSuffix(video, ".mp4")+".mkv")
	})
	t.Run("the attempts write to temporary files", func(t *testing.T) {
		video, entries := muxFixture(t)
		f := &ffmpegMock{failCodecs: map[string]bool{"mov_text": true}}
		muxSubtitles(f.run, video, entries)
		assert.Equal(t, video+".muxing.mp4", f.runs[0][len(f.runs[0])-1])
		assert.Equal(t, strings.TrimSuffix(video, ".mp4")+".mkv.tmp.mkv", f.runs[1][len(f.runs[1])-1])
	})
}

func muxFailures(impl func(func([]string) error, string, []subtitleEntry) muxResult) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	cases := []struct {
		name string
		fail map[string]bool
		want muxResult
		mp4  string
	}{
		{"mp4", nil, muxedMP4, "muxed with mov_text"},
		{"mkv", map[string]bool{"mov_text": true}, muxedMKV, "original"},
		{"none", map[string]bool{"mov_text": true, "srt": true}, muxFailed, "original"},
	}
	for _, tc := range cases {
		dir, _ := os.MkdirTemp("", "mux")
		video := filepath.Join(dir, "ep.mp4")
		_ = os.WriteFile(video, []byte("original"), 0o600)
		fm := &ffmpegMock{failCodecs: tc.fail}
		got := impl(fm.run, video, []subtitleEntry{{tmpPath: filepath.Join(dir, "s.srt"), label: "E", langCode: "en"}})
		b, _ := os.ReadFile(video)
		_, mkvErr := os.Stat(strings.TrimSuffix(video, ".mp4") + ".mkv")
		leftovers, _ := filepath.Glob(filepath.Join(dir, "*.muxing.mp4"))
		tmpMKV, _ := filepath.Glob(filepath.Join(dir, "*.tmp.mkv"))
		_ = os.RemoveAll(dir)
		if got != tc.want {
			f = append(f, fmt.Sprintf("%s: result %d want %d", tc.name, got, tc.want))
		}
		if string(b) != tc.mp4 {
			f = append(f, fmt.Sprintf("%s: video holds %q", tc.name, b))
		}
		if (mkvErr == nil) != (tc.want == muxedMKV) {
			f = append(f, tc.name+": .mkv presence wrong")
		}
		if len(leftovers)+len(tmpMKV) != 0 {
			f = append(f, tc.name+": temporary files left behind")
		}
	}
	return f
}

func TestMuxSubtitlesMeta(t *testing.T) {
	require.Empty(t, muxFailures(muxSubtitles))
	type fn = func(func([]string) error, string, []subtitleEntry) muxResult
	mutants := map[string]fn{
		"no mkv fallback": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			res := muxSubtitles(func(a []string) error {
				if a[slices.Index(a, "-c:s")+1] == "srt" {
					return errors.New("blocked")
				}
				return r(a)
			}, v, e)
			return res
		},
		"mp4 never attempted": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			return muxSubtitles(func(a []string) error {
				if a[slices.Index(a, "-c:s")+1] == "mov_text" {
					return errors.New("blocked")
				}
				return r(a)
			}, v, e)
		},
		"always reports success": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			muxSubtitles(r, v, e)
			return muxedMP4
		},
		"failure reported as mkv": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			if got := muxSubtitles(r, v, e); got == muxFailed {
				return muxedMKV
			}
			return muxSubtitles(r, v, e)
		},
		"partial output kept": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			got := muxSubtitles(r, v, e)
			_ = os.WriteFile(v+".muxing.mp4", []byte("partial"), 0o600)
			return got
		},
		"mkv reported as mp4": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			got := muxSubtitles(r, v, e)
			if got == muxedMKV {
				return muxedMP4
			}
			return got
		},
		"video overwritten on failure": func(r func([]string) error, v string, e []subtitleEntry) muxResult {
			got := muxSubtitles(r, v, e)
			if got == muxFailed {
				_ = os.WriteFile(v, []byte("corrupt"), 0o600)
			}
			return got
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, muxFailures(m), "mutant %q survived", name)
	}
}

// ---- downloadSubtitleFiles, end to end ---------------------------------------------------------

func TestDownloadSubtitleFilesEndToEnd(t *testing.T) {
	srv, _ := subtitleServer(t)
	dir := t.TempDir()
	video := filepath.Join(dir, "ep.mp4")
	require.NoError(t, os.WriteFile(video, []byte("original"), 0o600))
	ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg")}.install(t)

	oldClient, oldRun := subtitleHTTPClientFn, runFFmpegFn
	t.Cleanup(func() { subtitleHTTPClientFn, runFFmpegFn = oldClient, oldRun })
	subtitleHTTPClientFn = func() *http.Client { return srv.Client() }
	f := &ffmpegMock{}
	var ranWith string
	runFFmpegFn = func(path string, args []string) error { ranWith = path; return f.run(args) }

	var out []string
	printFn := func(format string, a ...any) { out = append(out, strings.TrimSpace(fmt.Sprintf(format, a...))) }
	downloadSubtitleFiles(video, []util.SubtitleInfo{{URL: srv.URL + "/en.srt", Label: "English", Language: "en"}}, printFn)

	assert.Equal(t, absBin("ffmpeg"), ranWith, "the validated path is what runs")
	assert.Equal(t, []string{"Embedding 1 subtitle(s) into video...", "Subtitles embedded successfully!"}, out)
	b, _ := os.ReadFile(video)
	assert.Equal(t, "muxed with mov_text", string(b))
	left, _ := filepath.Glob(filepath.Join(dir, ".tmp_sub_*"))
	assert.Empty(t, left, "downloaded subtitle files are cleaned up")
}

func TestDownloadSubtitleFilesMessages(t *testing.T) {
	srv, _ := subtitleServer(t)
	subs := []util.SubtitleInfo{{URL: srv.URL + "/en.srt", Label: "English", Language: "en"}}

	run := func(t *testing.T, l ffmpegLookup, failCodecs map[string]bool, subs []util.SubtitleInfo) (out []string, f *ffmpegMock) {
		t.Helper()
		dir := t.TempDir()
		video := filepath.Join(dir, "ep.mp4")
		require.NoError(t, os.WriteFile(video, []byte("original"), 0o600))
		l.install(t)
		oldClient, oldRun := subtitleHTTPClientFn, runFFmpegFn
		t.Cleanup(func() { subtitleHTTPClientFn, runFFmpegFn = oldClient, oldRun })
		subtitleHTTPClientFn = func() *http.Client { return srv.Client() }
		f = &ffmpegMock{failCodecs: failCodecs}
		runFFmpegFn = func(_ string, args []string) error { return f.run(args) }
		downloadSubtitleFiles(video, subs, func(format string, a ...any) {
			out = append(out, strings.TrimSpace(fmt.Sprintf(format, a...)))
		})
		return out, f
	}
	good := ffmpegLookup{found: "ffmpeg", resolved: absBin("ffmpeg")}

	t.Run("no subtitles does nothing at all", func(t *testing.T) {
		out, f := run(t, good, nil, nil)
		assert.Empty(t, out)
		assert.Empty(t, f.runs)
	})
	t.Run("ffmpeg missing warns", func(t *testing.T) {
		out, f := run(t, ffmpegLookup{lookErr: errors.New("none")}, nil, subs)
		assert.Equal(t, []string{"Warning: ffmpeg not found — cannot embed subtitles"}, out)
		assert.Empty(t, f.runs)
	})
	t.Run("ffmpeg unresolvable warns", func(t *testing.T) {
		out, _ := run(t, ffmpegLookup{found: "ffmpeg", evalErr: errors.New("loop")}, nil, subs)
		assert.Equal(t, []string{"Warning: failed to resolve ffmpeg path"}, out)
	})
	t.Run("an unsafe ffmpeg is refused silently", func(t *testing.T) {
		out, f := run(t, ffmpegLookup{found: "ffmpeg", resolved: "relative/ffmpeg"}, nil, subs)
		assert.Empty(t, out)
		assert.Empty(t, f.runs, "never executed")
	})
	t.Run("every track failing to download stops before ffmpeg", func(t *testing.T) {
		out, f := run(t, good, nil, []util.SubtitleInfo{{URL: srv.URL + "/gone.srt", Label: "Gone"}})
		assert.Empty(t, out)
		assert.Empty(t, f.runs)
	})
	t.Run("the mkv fallback says so", func(t *testing.T) {
		out, _ := run(t, good, map[string]bool{"mov_text": true}, subs)
		assert.Equal(t, []string{
			"Embedding 1 subtitle(s) into video...",
			"Note: saved as .mkv for better subtitle compatibility",
			"Subtitles embedded successfully!",
		}, out)
	})
	t.Run("both muxes failing warns", func(t *testing.T) {
		out, _ := run(t, good, map[string]bool{"mov_text": true, "srt": true}, subs)
		assert.Equal(t, []string{
			"Embedding 1 subtitle(s) into video...",
			"Warning: Could not embed subtitles — both MP4 and MKV muxing failed",
		}, out)
	})
}
