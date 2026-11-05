package player

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ytdlpArgs is the command line yt-dlp would be started with for spec. The
// executable is fixed so nothing is resolved, installed or run.
func ytdlpArgs(spec ytdlpSpec, extra ...string) []string {
	return newYtdlpCommand(spec).SetExecutable("yt-dlp").BuildCommand(context.Background(), extra...).Args
}

// flagValue is the argument following flag, "" when the flag is absent.
func flagValue(args []string, flag string) string {
	if i := slices.Index(args, flag); i >= 0 && i+1 < len(args) {
		return args[i+1]
	}
	return ""
}

func flagValues(args []string, flag string) []string {
	var out []string
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			out = append(out, args[i+1])
		}
	}
	return out
}

// ---- newYtdlpCommand -------------------------------------------------------------------

func commandFailures(impl func(ytdlpSpec) []string) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	full := impl(ytdlpSpec{URL: "https://cdn/x.m3u8", Output: "/o/ep.mp4", Impersonate: true, Headers: []string{"Referer:https://a/", "Origin:https://a"}})
	plain := impl(ytdlpSpec{URL: "https://cdn/x.m3u8", Output: "/o/ep.mp4"})

	for flag, want := range map[string]string{
		"--output": "/o/ep.mp4", "--format": "bestvideo+bestaudio/best", "--concurrent-fragments": "24",
		"--buffer-size": "32M", "--fragment-retries": "5", "--retries": "5", "--socket-timeout": "30",
	} {
		for name, args := range map[string][]string{"full": full, "plain": plain} {
			if got := flagValue(args, flag); got != want {
				f = append(f, fmt.Sprintf("%s: %s=%q want %q", name, flag, got, want))
			}
		}
	}
	if flagValue(full, "--impersonate") != "chrome" {
		f = append(f, "impersonation missing when asked for")
	}
	if slices.Contains(plain, "--impersonate") {
		f = append(f, "impersonation set when not asked for")
	}
	if got := flagValues(full, "--add-headers"); fmt.Sprint(got) != "[Referer:https://a/ Origin:https://a]" {
		f = append(f, fmt.Sprintf("headers %v", got))
	}
	if slices.Contains(plain, "--add-headers") {
		f = append(f, "headers set when none were given")
	}
	if full[len(full)-1] != "--hls-use-mpegts" || full[len(full)-2] != "https://cdn/x.m3u8" {
		f = append(f, fmt.Sprintf("URL and raw flag not last: %v", full[len(full)-2:]))
	}
	return f
}

func realCommandArgs(spec ytdlpSpec) []string { return ytdlpArgs(spec, spec.URL, "--hls-use-mpegts") }

func TestNewYtdlpCommand(t *testing.T) {
	assert.Empty(t, commandFailures(realCommandArgs))
}

func TestNewYtdlpCommandMeta(t *testing.T) {
	require.Empty(t, commandFailures(realCommandArgs))
	with := func(edit func(*ytdlpSpec)) func(ytdlpSpec) []string {
		return func(s ytdlpSpec) []string { edit(&s); return realCommandArgs(s) }
	}
	strip := func(flag string) func(ytdlpSpec) []string {
		return func(s ytdlpSpec) []string {
			a := realCommandArgs(s)
			for {
				i := slices.Index(a, flag)
				if i < 0 {
					return a
				}
				a = slices.Delete(slices.Clone(a), i, min(i+2, len(a)))
			}
		}
	}
	mutants := map[string]func(ytdlpSpec) []string{
		"always impersonates": with(func(s *ytdlpSpec) { s.Impersonate = true }),
		"never impersonates":  with(func(s *ytdlpSpec) { s.Impersonate = false }),
		"headers dropped":     with(func(s *ytdlpSpec) { s.Headers = nil }),
		"first header only":   with(func(s *ytdlpSpec) { s.Headers = s.Headers[:min(1, len(s.Headers))] }),
		"output elsewhere":    with(func(s *ytdlpSpec) { s.Output += ".part" }),
		"no output":           strip("--output"),
		"no format":           strip("--format"),
		"no fragments":        strip("--concurrent-fragments"),
		"no retries":          strip("--retries"),
		"no fragment retries": strip("--fragment-retries"),
		"no socket timeout":   strip("--socket-timeout"),
		"no buffer":           strip("--buffer-size"),
		"mpegts dropped": func(s ytdlpSpec) []string {
			return ytdlpArgs(s, s.URL)
		},
		"url dropped": func(s ytdlpSpec) []string {
			return ytdlpArgs(s, "--hls-use-mpegts")
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, commandFailures(m), "mutant %q survived", name)
	}
}

// ---- runYtdlpDownload / ytdlpOutcome ------------------------------------------------------

type engineMock struct {
	installErr error
	results    []error // one per Run; the last repeats
	updates    [][]ytdlp.ProgressUpdate

	installs int
	runs     int
	specs    []ytdlpSpec
	gotCB    []bool // whether a progress callback was given
}

func (e *engineMock) Install(context.Context) error { e.installs++; return e.installErr }

func (e *engineMock) Run(_ context.Context, spec ytdlpSpec, cb func(ytdlp.ProgressUpdate)) error {
	i := e.runs
	e.runs++
	e.specs = append(e.specs, spec)
	e.gotCB = append(e.gotCB, cb != nil)
	if cb != nil && i < len(e.updates) {
		for _, u := range e.updates[i] {
			cb(u)
		}
	}
	return e.results[min(i, len(e.results)-1)]
}

func TestRunYtdlpDownloadRetriesAndResetsProgress(t *testing.T) {
	dl := ytdlp.ProgressStatusDownloading
	eng := &engineMock{
		results: []error{errTransient, nil},
		updates: [][]ytdlp.ProgressUpdate{
			{upd(dl, "v.mp4", 400, 1000, 0, 0)},
			{upd(dl, "v.mp4", 150, 1000, 0, 0)}, // the retry starts over from fewer bytes
		},
	}
	var sleeps []time.Duration
	m := &model{}

	err := runYtdlpDownload(context.Background(), eng, ytdlpSpec{URL: "u", Output: "o"}, m, func(d time.Duration) { sleeps = append(sleeps, d) })

	require.NoError(t, err)
	assert.Equal(t, 2, eng.runs)
	assert.Equal(t, []time.Duration{2 * time.Second}, sleeps)
	assert.Equal(t, int64(550), m.received, "400 from the first attempt plus all 150 of the retry")
	assert.Equal(t, int64(1000), m.totalBytes)
	assert.Equal(t, ytdlpSpec{URL: "u", Output: "o"}, eng.specs[1], "the retry runs the same download")
}

func TestRunYtdlpDownloadWithoutAModelGivesNoProgressCallback(t *testing.T) {
	eng := &engineMock{results: []error{nil}}
	require.NoError(t, runYtdlpDownload(context.Background(), eng, ytdlpSpec{}, nil, func(time.Duration) {}))
	assert.Equal(t, []bool{false}, eng.gotCB)

	eng = &engineMock{results: []error{nil}}
	require.NoError(t, runYtdlpDownload(context.Background(), eng, ytdlpSpec{}, &model{}, func(time.Duration) {}))
	assert.Equal(t, []bool{true}, eng.gotCB)
}

func TestRunYtdlpDownloadStopsOnFatalErrors(t *testing.T) {
	eng := &engineMock{results: []error{errFatal}}
	err := runYtdlpDownload(context.Background(), eng, ytdlpSpec{}, &model{}, func(time.Duration) { t.Fatal("slept before a fatal error") })
	assert.ErrorIs(t, err, errFatal)
	assert.Equal(t, 1, eng.runs)
}

func TestRunYtdlpDownloadGivesUpAfterTheRetryBudget(t *testing.T) {
	eng := &engineMock{results: []error{errTransient}}
	err := runYtdlpDownload(context.Background(), eng, ytdlpSpec{}, &model{}, func(time.Duration) {})
	assert.ErrorIs(t, err, errTransient)
	assert.Equal(t, ytdlpMaxRetries+1, eng.runs)
}

func TestYtdlpOutcome(t *testing.T) {
	called := 0
	direct := func() error { called++; return errors.New("direct result") }

	assert.NoError(t, ytdlpOutcome(nil, direct))
	assert.Zero(t, called)

	err := ytdlpOutcome(errors.New("HTTP 404"), direct)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "go-ytdlp download failed")
	assert.Zero(t, called, "an ordinary failure does not fall back")

	err = ytdlpOutcome(errors.New("The extension .aspx is unusual and will be skipped"), direct)
	assert.EqualError(t, err, "direct result", "the unsafe-extension rejection falls back to a direct download and returns its result")
	assert.Equal(t, 1, called)
}

func outcomeFailures(impl func(error, func() error) error) (f []string) {
	direct := errors.New("direct")
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"success", nil, ""},
		{"ordinary", errors.New("boom"), "go-ytdlp download failed: boom"},
		{"unsafe extension", errors.New("unsafe extension"), "direct"},
		{"unusual extension", errors.New("unusual extension"), "direct"},
	}
	for _, tc := range cases {
		got := impl(tc.err, func() error { return direct })
		s := ""
		if got != nil {
			s = got.Error()
		}
		if s != tc.want {
			f = append(f, fmt.Sprintf("%s: %q want %q", tc.name, s, tc.want))
		}
	}
	return f
}

func TestYtdlpOutcomeMeta(t *testing.T) {
	require.Empty(t, outcomeFailures(ytdlpOutcome))
	mutants := map[string]func(error, func() error) error{
		"always falls back": func(e error, d func() error) error {
			if e == nil {
				return nil
			}
			return d()
		},
		"never falls back": func(e error, _ func() error) error { return ytdlpOutcome(e, func() error { return e }) },
		"failure swallowed": func(e error, d func() error) error {
			if e != nil && !isUnsafeExtensionError(e) {
				return nil
			}
			return ytdlpOutcome(e, d)
		},
		"success becomes error": func(e error, d func() error) error {
			if e == nil {
				return errors.New("go-ytdlp download failed: <nil>")
			}
			return ytdlpOutcome(e, d)
		},
		"cause not wrapped": func(e error, d func() error) error {
			if e != nil && !isUnsafeExtensionError(e) {
				return errors.New("go-ytdlp download failed: " + strings.ToUpper(e.Error()))
			}
			return ytdlpOutcome(e, d)
		},
		"fallback result dropped": func(e error, d func() error) error {
			if e != nil && isUnsafeExtensionError(e) {
				_ = d()
				return nil
			}
			return ytdlpOutcome(e, d)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, outcomeFailures(m), "mutant %q survived", name)
	}
}

// ---- downloadWithYtDlp, end to end over the mock engine ------------------------------------

type ytdlpHarness struct {
	eng      *engineMock
	direct   [][2]string
	directEr error
	sleeps   []time.Duration
	impers   bool
}

func installYtdlpHarness(t *testing.T, eng *engineMock) *ytdlpHarness {
	t.Helper()
	h := &ytdlpHarness{eng: eng}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // sanitizeOutputPath keeps outputs under the home directory

	oldEng, oldImp, oldSleep, oldDirect := newYtdlpEngine, ytdlpCanImpersonate, ytdlpSleep, downloadDirectHTTPFn
	t.Cleanup(func() {
		newYtdlpEngine, ytdlpCanImpersonate, ytdlpSleep, downloadDirectHTTPFn = oldEng, oldImp, oldSleep, oldDirect
	})
	newYtdlpEngine = func() ytdlpEngine { return eng }
	ytdlpCanImpersonate = func() bool { return h.impers }
	ytdlpSleep = func(d time.Duration) { h.sleeps = append(h.sleeps, d) }
	downloadDirectHTTPFn = func(u, p string, _ *model) error {
		h.direct = append(h.direct, [2]string{u, p})
		return h.directEr
	}
	return h
}

func homeOut(t *testing.T, rel string) string {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	return filepath.Join(home, rel)
}

func TestDownloadWithYtDlpRunsTheEngineWithTheRightSpec(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	eng := &engineMock{results: []error{nil}}
	h := installYtdlpHarness(t, eng)
	h.impers = true
	downloadReferers.Store("https://cdn.test/master.m3u8", "https://site.test/watch/")
	t.Cleanup(func() { downloadReferers.Delete("https://cdn.test/master.m3u8") })
	out := homeOut(t, "dl/ep1.mp4")

	require.NoError(t, downloadWithYtDlp("https://cdn.test/master.m3u8", out, &model{}))

	assert.Equal(t, 1, eng.installs)
	require.Len(t, eng.specs, 1)
	assert.Equal(t, ytdlpSpec{
		URL: "https://cdn.test/master.m3u8", Output: out, Impersonate: true,
		Headers: []string{"Referer:https://site.test/watch/", "Origin:https://site.test"},
	}, eng.specs[0])
	assert.DirExists(t, filepath.Dir(out), "the output directory is created before the run")
	assert.Empty(t, h.direct)
}

func TestDownloadWithYtDlpFailureModes(t *testing.T) {
	t.Run("an unsafe URL never reaches the engine", func(t *testing.T) {
		eng := &engineMock{results: []error{nil}}
		installYtdlpHarness(t, eng)
		err := downloadWithYtDlp("--exec=evil", homeOut(t, "x.mp4"), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid download URL")
		assert.Zero(t, eng.installs+eng.runs)
	})
	t.Run("an output path outside the home directory never reaches the engine", func(t *testing.T) {
		eng := &engineMock{results: []error{nil}}
		installYtdlpHarness(t, eng)
		err := downloadWithYtDlp("https://cdn.test/a.m3u8", filepath.Join(os.TempDir(), "elsewhere", "x.mp4"), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output path")
		assert.Zero(t, eng.installs+eng.runs)
	})
	t.Run("an output directory that cannot be created", func(t *testing.T) {
		eng := &engineMock{results: []error{nil}}
		installYtdlpHarness(t, eng)
		blocker := homeOut(t, "afile")
		require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
		err := downloadWithYtDlp("https://cdn.test/a.m3u8", filepath.Join(blocker, "sub", "x.mp4"), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output directory")
		assert.Zero(t, eng.installs+eng.runs)
	})
	t.Run("install failure stops before any run", func(t *testing.T) {
		boom := errors.New("offline")
		eng := &engineMock{installErr: boom, results: []error{nil}}
		installYtdlpHarness(t, eng)
		err := downloadWithYtDlp("https://cdn.test/a.m3u8", homeOut(t, "x.mp4"), nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, boom)
		assert.Contains(t, err.Error(), "failed to install yt-dlp")
		assert.Zero(t, eng.runs)
	})
	t.Run("a run failure is wrapped", func(t *testing.T) {
		eng := &engineMock{results: []error{errFatal}}
		installYtdlpHarness(t, eng)
		err := downloadWithYtDlp("https://cdn.test/a.m3u8", homeOut(t, "x.mp4"), nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, errFatal)
		assert.Contains(t, err.Error(), "go-ytdlp download failed")
	})
	t.Run("transient failures are retried with a pause", func(t *testing.T) {
		eng := &engineMock{results: []error{errTransient, errTimeout, nil}}
		h := installYtdlpHarness(t, eng)
		require.NoError(t, downloadWithYtDlp("https://cdn.test/a.m3u8", homeOut(t, "x.mp4"), nil))
		assert.Equal(t, 3, eng.runs)
		assert.Equal(t, []time.Duration{2 * time.Second, 4 * time.Second}, h.sleeps)
	})
	t.Run("an unsafe extension falls back to a direct download", func(t *testing.T) {
		eng := &engineMock{results: []error{errors.New("ERROR: The extension is unusual and will be skipped for security reasons")}}
		h := installYtdlpHarness(t, eng)
		out := homeOut(t, "x.mp4")
		require.NoError(t, downloadWithYtDlp("https://cdn.test/a.aspx", out, nil))
		assert.Equal(t, [][2]string{{"https://cdn.test/a.aspx", out}}, h.direct)
	})
	t.Run("the fallback's own failure is the result", func(t *testing.T) {
		eng := &engineMock{results: []error{errors.New("unsafe extension")}}
		h := installYtdlpHarness(t, eng)
		h.directEr = errors.New("HTTP 403")
		err := downloadWithYtDlp("https://cdn.test/a.aspx", homeOut(t, "x.mp4"), nil)
		assert.EqualError(t, err, "HTTP 403")
	})
}

func TestDownloadWithYtDlpPassesHeadersOnlyWhenAReferrerIsKnown(t *testing.T) {
	eng := &engineMock{results: []error{nil}}
	installYtdlpHarness(t, eng)
	require.NoError(t, downloadWithYtDlp("https://nobody-set-a-referer.test/a.m3u8", homeOut(t, "x.mp4"), nil))
	assert.Empty(t, eng.specs[0].Headers)
	assert.False(t, eng.specs[0].Impersonate)
}

func TestGoYtdlpEngineIsTheDefault(t *testing.T) {
	assert.IsType(t, goYtdlpEngine{}, newYtdlpEngine())
}
