package ytdlpdl

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Headers ---------------------------------------------------------------

var ytdlpHeaderCases = []struct {
	name, referer string
	want          []string
}{
	{"none", "", nil},
	{"site root", "https://animefire.io/", []string{"Referer:https://animefire.io/", "Origin:https://animefire.io"}},
	{"no trailing slash", "https://cdn.test", []string{"Referer:https://cdn.test", "Origin:https://cdn.test"}},
	{"path is dropped from the origin", "https://cdn.test/watch/ep1?x=1", []string{"Referer:https://cdn.test/watch/ep1?x=1", "Origin:https://cdn.test"}},
	{"port is kept", "http://127.0.0.1:8080/p/", []string{"Referer:http://127.0.0.1:8080/p/", "Origin:http://127.0.0.1:8080"}},
	{"no scheme is not turned into ://", "cdn.test/", []string{"Referer:cdn.test/", "Origin:cdn.test"}},
	{"scheme-relative keeps the referer as the origin", "//cdn.test/p", []string{"Referer://cdn.test/p", "Origin://cdn.test/p"}},
	{"no host keeps the referer as the origin", "mailto:a@b.test", []string{"Referer:mailto:a@b.test", "Origin:mailto:a@b.test"}},
}

func ytdlpHeaderFailures(impl func(string) []string) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range ytdlpHeaderCases {
		if got := impl(tc.referer); fmt.Sprint(got) != fmt.Sprint(tc.want) {
			f = append(f, fmt.Sprintf("%s: got %v want %v", tc.name, got, tc.want))
		}
	}
	return f
}

func TestYtdlpHeaders(t *testing.T) {
	assert.Empty(t, ytdlpHeaderFailures(Headers))
}

func TestYtdlpHeadersMeta(t *testing.T) {
	require.Empty(t, ytdlpHeaderFailures(Headers))
	mutants := map[string]func(string) []string{
		"always headers": func(r string) []string { return append([]string{"Referer:" + r}, "Origin:"+r) },
		"origin keeps the path": func(r string) []string {
			h := Headers(r)
			if h != nil {
				h[1] = "Origin:" + strings.TrimSuffix(r, "/")
			}
			return h
		},
		"origin and referer swapped": func(r string) []string {
			h := Headers(r)
			if len(h) == 2 {
				h[0], h[1] = h[1], h[0]
			}
			return h
		},
		"referer only": func(r string) []string {
			h := Headers(r)
			if len(h) == 2 {
				return h[:1]
			}
			return h
		},
		"no colon": func(r string) []string {
			h := Headers(r)
			for i := range h {
				h[i] = strings.Replace(h[i], ":", " ", 1)
			}
			return h
		},
		"scheme-less origin gets ://": func(r string) []string {
			h := Headers(r)
			if len(h) == 2 && !strings.Contains(r, "://") {
				h[1] = "Origin:://"
			}
			return h
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, ytdlpHeaderFailures(m), "mutant %q survived", name)
	}
}

// ---- RunWithRetry ----------------------------------------------------------------

type retryRun struct {
	results []error // one per call; the last repeats
	calls   int
	sleeps  []time.Duration
	retries []int
}

func (r *retryRun) fn() error {
	i := min(r.calls, len(r.results)-1)
	r.calls++
	return r.results[i]
}

var (
	errTransient = errors.New("connection reset by peer")
	errTimeout   = errors.New("request timeout")
	errFatal     = errors.New("HTTP Error 404: Not Found")
)

var retryCases = []struct {
	name       string
	results    []error
	max        int
	wantErr    error
	wantCalls  int
	wantSleeps []time.Duration
}{
	{"succeeds at once", []error{nil}, 2, nil, 1, nil},
	{"recovers on the second try", []error{errTransient, nil}, 2, nil, 2, []time.Duration{2 * time.Second}},
	{"recovers on the last try", []error{errTransient, errTimeout, nil}, 2, nil, 3, []time.Duration{2 * time.Second, 4 * time.Second}},
	{"gives up after max retries", []error{errTransient}, 2, errTransient, 3, []time.Duration{2 * time.Second, 4 * time.Second}},
	{"fatal errors are not retried", []error{errFatal}, 2, errFatal, 1, nil},
	{"fatal after a transient one stops there", []error{errTransient, errFatal}, 2, errFatal, 2, []time.Duration{2 * time.Second}},
	{"no retries allowed", []error{errTransient}, 0, errTransient, 1, nil},
	{"returns the last error", []error{errTransient, errTimeout, errTransient}, 2, errTransient, 3, []time.Duration{2 * time.Second, 4 * time.Second}},
}

func retryFailures(impl func(func() error, int, func(time.Duration), func(int)) error) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range retryCases {
		r := &retryRun{results: tc.results}
		err := impl(r.fn, tc.max,
			func(d time.Duration) { r.sleeps = append(r.sleeps, d) },
			func(n int) { r.retries = append(r.retries, n) })
		if !errors.Is(err, tc.wantErr) || (tc.wantErr == nil && err != nil) {
			f = append(f, fmt.Sprintf("%s: error %v", tc.name, err))
		}
		if r.calls != tc.wantCalls {
			f = append(f, fmt.Sprintf("%s: %d calls want %d", tc.name, r.calls, tc.wantCalls))
		}
		if fmt.Sprint(r.sleeps) != fmt.Sprint(tc.wantSleeps) {
			f = append(f, fmt.Sprintf("%s: sleeps %v want %v", tc.name, r.sleeps, tc.wantSleeps))
		}
		wantRetries := make([]int, 0, len(tc.wantSleeps))
		for i := range tc.wantSleeps {
			wantRetries = append(wantRetries, i+1)
		}
		if fmt.Sprint(r.retries) != fmt.Sprint(wantRetries) {
			f = append(f, fmt.Sprintf("%s: onRetry %v want %v", tc.name, r.retries, wantRetries))
		}
	}
	return f
}

func TestRunWithRetry(t *testing.T) {
	assert.Empty(t, retryFailures(RunWithRetry))
}

func TestRunWithRetryMeta(t *testing.T) {
	require.Empty(t, retryFailures(RunWithRetry))
	type fn = func(func() error, int, func(time.Duration), func(int)) error
	mutants := map[string]fn{
		"never retries": func(run func() error, _ int, _ func(time.Duration), _ func(int)) error { return run() },
		"retries fatal errors": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			var err error
			for a := 0; a <= max; a++ {
				if a > 0 {
					sl(time.Duration(a*2) * time.Second)
					on(a)
				}
				if err = run(); err == nil {
					return nil
				}
			}
			return err
		},
		"one retry too many": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			return RunWithRetry(run, max+1, sl, on)
		},
		"one retry too few": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			return RunWithRetry(run, max-1, sl, on)
		},
		"linear backoff": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			return RunWithRetry(run, max, func(d time.Duration) { sl(d / 2) }, on)
		},
		"no sleep": func(run func() error, max int, _ func(time.Duration), on func(int)) error {
			return RunWithRetry(run, max, func(time.Duration) {}, on)
		},
		"no reset hook": func(run func() error, max int, sl func(time.Duration), _ func(int)) error {
			return RunWithRetry(run, max, sl, func(int) {})
		},
		"swallows the final failure": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			if err := RunWithRetry(run, max, sl, on); err != nil && IsRetryable(err) {
				return nil
			}
			return RunWithRetry(run, max, sl, on)
		},
		"returns the first error": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			first := run()
			if first == nil {
				return nil
			}
			_ = RunWithRetry(run, max, sl, on)
			return first
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, retryFailures(m), "mutant %q survived", name)
	}
}

// ---- ytdlpProgress ----------------------------------------------------------------

type sinkMock struct {
	received []int64
	totals   []int64
	peaks    []float64
}

func (s *sinkMock) AddProgressReceived(d int64) { s.received = append(s.received, d) }
func (s *sinkMock) SetProgressTotal(t int64)    { s.totals = append(s.totals, t) }
func (s *sinkMock) SetProgressPeak(p float64)   { s.peaks = append(s.peaks, p) }

func upd(status ytdlp.ProgressStatus, file string, done, total, frag, fragCount int) ytdlp.ProgressUpdate {
	return ytdlp.ProgressUpdate{Status: status, Filename: file, DownloadedBytes: done, TotalBytes: total, FragmentIndex: frag, FragmentCount: fragCount}
}

func TestYtdlpProgressCountsBytesPerFile(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading

	p.Handle(upd(dl, "video.mp4", 100, 1000, 0, 0))
	p.Handle(upd(dl, "video.mp4", 250, 1000, 0, 0))
	p.Handle(upd(dl, "video.mp4", 250, 1000, 0, 0)) // no change: nothing to add
	assert.Equal(t, []int64{100, 150}, s.received, "the delta since the last update")

	p.Handle(upd(dl, "audio.m4a", 40, 200, 0, 0)) // a new file starts from zero
	assert.Equal(t, []int64{100, 150, 40}, s.received)
	assert.Equal(t, int64(1200), s.totals[len(s.totals)-1], "video 1000 + audio 200")
}

func TestYtdlpProgressTotalsSumEveryFile(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.Handle(upd(dl, "a", 1, 500, 0, 0))
	p.Handle(upd(dl, "b", 1, 300, 0, 0))
	p.Handle(upd(dl, "a", 2, 600, 0, 0)) // a's size is revised, not added twice
	assert.Equal(t, []int64{500, 800, 900}, s.totals)
}

func TestYtdlpProgressUnnamedFilesShareOneSlot(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.Handle(upd(dl, "", 1, 700, 0, 0))
	p.Handle(upd(dl, "", 2, 700, 0, 0))
	assert.Equal(t, []int64{700, 700}, s.totals, "an empty file name is one file, not many")
}

func TestYtdlpProgressHLSUsesFragments(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.Handle(upd(dl, "f", 10, 0, 1, 4))
	p.Handle(upd(dl, "f", 20, 0, 2, 4))
	p.Handle(upd(dl, "f", 30, 0, 4, 4)) // 100% is held back to 99%
	assert.Equal(t, []float64{0.25, 0.5, 0.99}, s.peaks)
	assert.Empty(t, s.totals, "no byte total is known for HLS")
}

func TestYtdlpProgressIgnoresFinishedAndPostProcessing(t *testing.T) {
	for _, st := range []ytdlp.ProgressStatus{ytdlp.ProgressStatusFinished, ytdlp.ProgressStatusPostProcessing} {
		s := &sinkMock{}
		NewProgress(s).Handle(upd(st, "f", 999, 999, 1, 2))
		assert.Empty(t, s.received, st)
		assert.Empty(t, s.totals, st)
		assert.Empty(t, s.peaks, st)
	}
}

func TestYtdlpProgressPrefersBytesOverFragments(t *testing.T) {
	s := &sinkMock{}
	NewProgress(s).Handle(upd(ytdlp.ProgressStatusDownloading, "f", 5, 100, 1, 2))
	assert.Equal(t, []int64{100}, s.totals)
	assert.Empty(t, s.peaks)
}

func TestYtdlpProgressIgnoresIncompleteFragmentData(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.Handle(upd(dl, "f", 0, 0, 0, 5)) // no index yet
	p.Handle(upd(dl, "f", 0, 0, 3, 0)) // no count yet
	assert.Empty(t, s.peaks)
}

func TestYtdlpProgressResetForRetryRestartsTheCount(t *testing.T) {
	s := &sinkMock{}
	p := NewProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.Handle(upd(dl, "v", 500, 1000, 0, 0))
	p.ResetForRetry()
	p.Handle(upd(dl, "v", 120, 1000, 0, 0))
	assert.Equal(t, []int64{500, 120}, s.received, "after a retry the first update counts in full")

	// Without the reset the 120 would be swallowed as going backwards.
	s2 := &sinkMock{}
	p2 := NewProgress(s2)
	p2.Handle(upd(dl, "v", 500, 1000, 0, 0))
	p2.Handle(upd(dl, "v", 120, 1000, 0, 0))
	assert.Equal(t, []int64{500}, s2.received)
}

// progressScript is replayed against any progress handler; the fingerprint of
// what reached the sink is compared with the real one.
func progressFingerprint(newHandler func(Sink) (Handle func(ytdlp.ProgressUpdate), reset func())) string {
	s := &sinkMock{}
	handle, reset := newHandler(s)
	dl := ytdlp.ProgressStatusDownloading
	for _, u := range []ytdlp.ProgressUpdate{
		upd(dl, "v", 100, 1000, 0, 0), upd(dl, "v", 300, 1000, 0, 0), upd(dl, "a", 50, 200, 0, 0),
		upd(ytdlp.ProgressStatusFinished, "a", 200, 200, 0, 0), upd(dl, "h", 1, 0, 1, 4), upd(dl, "h", 2, 0, 4, 4),
		upd(dl, "", 7, 70, 0, 0), upd(ytdlp.ProgressStatusPostProcessing, "m", 1, 1, 0, 0),
	} {
		handle(u)
	}
	reset()
	handle(upd(dl, "h", 1, 0, 1, 4)) // same file as before, fewer bytes: only a reset lets it count
	return fmt.Sprint(s.received, s.totals, s.peaks)
}

func TestYtdlpProgressMeta(t *testing.T) {
	want := progressFingerprint(func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
		p := NewProgress(s)
		return p.Handle, p.ResetForRetry
	})
	mutants := map[string]func(Sink) (func(ytdlp.ProgressUpdate), func()){
		"no reset": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return p.Handle, func() {}
		},
		"total of the latest file only": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				p.Handle(u)
				p.totals = map[string]int64{}
			}, p.ResetForRetry
		},
		"reports finished updates": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.Status == ytdlp.ProgressStatusFinished {
					u.Status = ytdlp.ProgressStatusDownloading
				}
				p.Handle(u)
			}, p.ResetForRetry
		},
		"fragment progress reaches 100": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.FragmentCount > 0 && u.FragmentIndex == u.FragmentCount && u.TotalBytes == 0 {
					s.SetProgressPeak(1.0)
					return
				}
				p.Handle(u)
			}, p.ResetForRetry
		},
		"adds the running total, not the delta": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.Status == ytdlp.ProgressStatusDownloading && u.DownloadedBytes > 0 && u.TotalBytes > 0 {
					s.AddProgressReceived(int64(u.DownloadedBytes))
					u.DownloadedBytes = 0
				}
				p.Handle(u)
			}, p.ResetForRetry
		},
		"ignores new files": func(s Sink) (func(ytdlp.ProgressUpdate), func()) {
			p := NewProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if p.lastFile != "" {
					u.Filename = p.lastFile
				}
				p.Handle(u)
			}, p.ResetForRetry
		},
	}
	for name, m := range mutants {
		assert.NotEqual(t, want, progressFingerprint(m), "mutant %q survived", name)
	}
}

func TestYtdlpProgressBoundaries(t *testing.T) {
	dl := ytdlp.ProgressStatusDownloading

	t.Run("a one-byte total counts", func(t *testing.T) {
		s := &sinkMock{}
		NewProgress(s).Handle(upd(dl, "f", 1, 1, 0, 0))
		assert.Equal(t, []int64{1}, s.totals)
	})
	t.Run("a single fragment is complete but held at 99", func(t *testing.T) {
		s := &sinkMock{}
		NewProgress(s).Handle(upd(dl, "f", 0, 0, 1, 1))
		assert.Equal(t, []float64{0.99}, s.peaks)
	})
	t.Run("the first fragment of many counts", func(t *testing.T) {
		s := &sinkMock{}
		NewProgress(s).Handle(upd(dl, "f", 0, 0, 1, 10))
		assert.Equal(t, []float64{0.1}, s.peaks)
	})
	t.Run("an unnamed file after a retry counts from zero", func(t *testing.T) {
		s := &sinkMock{}
		p := NewProgress(s)
		p.Handle(upd(dl, "", 500, 0, 0, 0))
		p.ResetForRetry()
		p.Handle(upd(dl, "", 100, 0, 0, 0))
		assert.Equal(t, []int64{500, 100}, s.received, "with no file name only the byte counter can tell the attempt restarted")
	})
	t.Run("a new file starts from zero even when it is smaller", func(t *testing.T) {
		s := &sinkMock{}
		p := NewProgress(s)
		p.Handle(upd(dl, "a", 500, 0, 0, 0))
		p.Handle(upd(dl, "b", 1, 0, 0, 0))
		assert.Equal(t, []int64{500, 1}, s.received)
	})
}

// ytdlpArgs is the command line yt-dlp would be started with for spec. The
// executable is fixed so nothing is resolved, installed or run.
func ytdlpArgs(spec Spec, extra ...string) []string {
	return NewCommand(spec).SetExecutable("yt-dlp").BuildCommand(context.Background(), extra...).Args
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

// ---- NewCommand -------------------------------------------------------------------

func commandFailures(impl func(Spec) []string) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	full := impl(Spec{URL: "https://cdn/x.m3u8", Output: "/o/ep.mp4", Impersonate: true, Headers: []string{"Referer:https://a/", "Origin:https://a"}})
	plain := impl(Spec{URL: "https://cdn/x.m3u8", Output: "/o/ep.mp4"})

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

func realCommandArgs(spec Spec) []string { return ytdlpArgs(spec, spec.URL, "--hls-use-mpegts") }

func TestNewYtdlpCommand(t *testing.T) {
	assert.Empty(t, commandFailures(realCommandArgs))
}

func TestNewYtdlpCommandMeta(t *testing.T) {
	require.Empty(t, commandFailures(realCommandArgs))
	with := func(edit func(*Spec)) func(Spec) []string {
		return func(s Spec) []string { edit(&s); return realCommandArgs(s) }
	}
	strip := func(flag string) func(Spec) []string {
		return func(s Spec) []string {
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
	mutants := map[string]func(Spec) []string{
		"always impersonates": with(func(s *Spec) { s.Impersonate = true }),
		"never impersonates":  with(func(s *Spec) { s.Impersonate = false }),
		"headers dropped":     with(func(s *Spec) { s.Headers = nil }),
		"first header only":   with(func(s *Spec) { s.Headers = s.Headers[:min(1, len(s.Headers))] }),
		"output elsewhere":    with(func(s *Spec) { s.Output += ".part" }),
		"no output":           strip("--output"),
		"no format":           strip("--format"),
		"no fragments":        strip("--concurrent-fragments"),
		"no retries":          strip("--retries"),
		"no fragment retries": strip("--fragment-retries"),
		"no socket timeout":   strip("--socket-timeout"),
		"no buffer":           strip("--buffer-size"),
		"mpegts dropped": func(s Spec) []string {
			return ytdlpArgs(s, s.URL)
		},
		"url dropped": func(s Spec) []string {
			return ytdlpArgs(s, "--hls-use-mpegts")
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, commandFailures(m), "mutant %q survived", name)
	}
}

// ---- Run / Outcome ------------------------------------------------------

type engineMock struct {
	installErr error
	results    []error // one per Run; the last repeats
	updates    [][]ytdlp.ProgressUpdate

	installs int
	runs     int
	specs    []Spec
	gotCB    []bool // whether a progress callback was given
}

func (e *engineMock) Install(context.Context) error { e.installs++; return e.installErr }

func (e *engineMock) Run(_ context.Context, spec Spec, cb func(ytdlp.ProgressUpdate)) error {
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
	s := &sinkMock{}

	err := Run(context.Background(), eng, Spec{URL: "u", Output: "o"}, s, func(d time.Duration) { sleeps = append(sleeps, d) }, nil)

	require.NoError(t, err)
	assert.Equal(t, 2, eng.runs)
	assert.Equal(t, []time.Duration{2 * time.Second}, sleeps)
	var received int64
	for _, d := range s.received {
		received += d
	}
	assert.Equal(t, int64(550), received, "400 from the first attempt plus all 150 of the retry")
	assert.Equal(t, int64(1000), s.totals[len(s.totals)-1])
	assert.Equal(t, Spec{URL: "u", Output: "o"}, eng.specs[1], "the retry runs the same download")
}

func TestRunYtdlpDownloadWithoutASinkGivesNoProgressCallback(t *testing.T) {
	eng := &engineMock{results: []error{nil}}
	require.NoError(t, Run(context.Background(), eng, Spec{}, nil, func(time.Duration) {}, nil))
	assert.Equal(t, []bool{false}, eng.gotCB)

	eng = &engineMock{results: []error{nil}}
	require.NoError(t, Run(context.Background(), eng, Spec{}, &sinkMock{}, func(time.Duration) {}, nil))
	assert.Equal(t, []bool{true}, eng.gotCB)
}

func TestRunYtdlpDownloadStopsOnFatalErrors(t *testing.T) {
	eng := &engineMock{results: []error{errFatal}}
	err := Run(context.Background(), eng, Spec{}, &sinkMock{}, func(time.Duration) { t.Fatal("slept before a fatal error") }, nil)
	assert.ErrorIs(t, err, errFatal)
	assert.Equal(t, 1, eng.runs)
}

func TestRunYtdlpDownloadGivesUpAfterTheRetryBudget(t *testing.T) {
	eng := &engineMock{results: []error{errTransient}}
	err := Run(context.Background(), eng, Spec{}, &sinkMock{}, func(time.Duration) {}, nil)
	assert.ErrorIs(t, err, errTransient)
	assert.Equal(t, MaxRetries+1, eng.runs)
}

func TestYtdlpOutcome(t *testing.T) {
	called := 0
	direct := func() error { called++; return errors.New("direct result") }

	assert.NoError(t, Outcome(nil, direct))
	assert.Zero(t, called)

	err := Outcome(errors.New("HTTP 404"), direct)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "go-ytdlp download failed")
	assert.Zero(t, called, "an ordinary failure does not fall back")

	err = Outcome(errors.New("The extension .aspx is unusual and will be skipped"), direct)
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
	require.Empty(t, outcomeFailures(Outcome))
	mutants := map[string]func(error, func() error) error{
		"always falls back": func(e error, d func() error) error {
			if e == nil {
				return nil
			}
			return d()
		},
		"never falls back": func(e error, _ func() error) error { return Outcome(e, func() error { return e }) },
		"failure swallowed": func(e error, d func() error) error {
			if e != nil && !IsUnsafeExtensionError(e) {
				return nil
			}
			return Outcome(e, d)
		},
		"success becomes error": func(e error, d func() error) error {
			if e == nil {
				return errors.New("go-ytdlp download failed: <nil>")
			}
			return Outcome(e, d)
		},
		"cause not wrapped": func(e error, d func() error) error {
			if e != nil && !IsUnsafeExtensionError(e) {
				return errors.New("go-ytdlp download failed: " + strings.ToUpper(e.Error()))
			}
			return Outcome(e, d)
		},
		"fallback result dropped": func(e error, d func() error) error {
			if e != nil && IsUnsafeExtensionError(e) {
				_ = d()
				return nil
			}
			return Outcome(e, d)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, outcomeFailures(m), "mutant %q survived", name)
	}
}

// =====================================================================
// Tests: IsRetryable
// =====================================================================

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"timeout", errors.New("i/o timeout"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"network unreachable", errors.New("network is unreachable"), true},
		{"temporary failure", errors.New("temporary failure in name resolution"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"permission denied", errors.New("permission denied"), false},
		{"not found", errors.New("404 not found"), false},
		{"ffmpeg exit code", errors.New("ffmpeg exited with code 183"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsRetryable(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// =====================================================================
// Tests: IsUnsafeExtensionError
// =====================================================================

func TestIsUnsafeExtensionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"unsafe extension msg", errors.New("unsafe file extension detected"), true},
		{"unusual extension msg", errors.New("the file has an unusual extension"), true},
		{"skip msg", errors.New("is unusual and will be skipped"), true},
		{"yt-dlp CVE message", errors.New("URL has an unsafe extension"), true},
		{"generic error", errors.New("download failed"), false},
		{"ffmpeg extension error", errors.New("not in allowed_segment_extensions"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsUnsafeExtensionError(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Cover the bare error-text branches of IsUnsafeExtensionError & IsRetryable
// that the existing tests do not hit.
func TestIsUnsafeExtensionError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, IsUnsafeExtensionError(nil))
	assert.True(t, IsUnsafeExtensionError(errors.New("file has unsafe extension")))
	assert.True(t, IsUnsafeExtensionError(errors.New("file is unusual and will be skipped")))
	assert.False(t, IsUnsafeExtensionError(errors.New("other failure")))
}

func TestIsRetryableError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, IsRetryable(nil))
	assert.True(t, IsRetryable(errors.New("connection reset")))
	assert.True(t, IsRetryable(errors.New("network down")))
	assert.True(t, IsRetryable(errors.New("connection refused")))
	assert.True(t, IsRetryable(errors.New("temporary error")))
	assert.False(t, IsRetryable(errors.New("auth denied")))
}
