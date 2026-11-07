package player

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/ytdlpdl"
	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// engineMock is a yt-dlp engine that records its runs and replays canned results.
type engineMock struct {
	installErr error
	results    []error // one per Run; the last repeats
	updates    [][]ytdlp.ProgressUpdate

	installs int
	runs     int
	specs    []ytdlpdl.Spec
	gotCB    []bool // whether a progress callback was given
}

func (e *engineMock) Install(context.Context) error { e.installs++; return e.installErr }

func (e *engineMock) Run(_ context.Context, spec ytdlpdl.Spec, cb func(ytdlp.ProgressUpdate)) error {
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

var (
	errTransient = errors.New("connection reset by peer")
	errTimeout   = errors.New("i/o timeout")
	errFatal     = errors.New("HTTP Error 404: Not Found")
)

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
	newYtdlpEngine = func() ytdlpdl.Engine { return eng }
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
	assert.Equal(t, ytdlpdl.Spec{
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
	assert.IsType(t, ytdlpdl.GoEngine{}, newYtdlpEngine())
}
