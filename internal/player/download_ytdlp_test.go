package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- ytdlpHeaders ---------------------------------------------------------------

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
	assert.Empty(t, ytdlpHeaderFailures(ytdlpHeaders))
}

func TestYtdlpHeadersMeta(t *testing.T) {
	require.Empty(t, ytdlpHeaderFailures(ytdlpHeaders))
	mutants := map[string]func(string) []string{
		"always headers": func(r string) []string { return append([]string{"Referer:" + r}, "Origin:"+r) },
		"origin keeps the path": func(r string) []string {
			h := ytdlpHeaders(r)
			if h != nil {
				h[1] = "Origin:" + strings.TrimSuffix(r, "/")
			}
			return h
		},
		"origin and referer swapped": func(r string) []string {
			h := ytdlpHeaders(r)
			if len(h) == 2 {
				h[0], h[1] = h[1], h[0]
			}
			return h
		},
		"referer only": func(r string) []string {
			h := ytdlpHeaders(r)
			if len(h) == 2 {
				return h[:1]
			}
			return h
		},
		"no colon": func(r string) []string {
			h := ytdlpHeaders(r)
			for i := range h {
				h[i] = strings.Replace(h[i], ":", " ", 1)
			}
			return h
		},
		"scheme-less origin gets ://": func(r string) []string {
			h := ytdlpHeaders(r)
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

// ---- runWithRetry ----------------------------------------------------------------

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
	assert.Empty(t, retryFailures(runWithRetry))
}

func TestRunWithRetryMeta(t *testing.T) {
	require.Empty(t, retryFailures(runWithRetry))
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
			return runWithRetry(run, max+1, sl, on)
		},
		"one retry too few": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			return runWithRetry(run, max-1, sl, on)
		},
		"linear backoff": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			return runWithRetry(run, max, func(d time.Duration) { sl(d / 2) }, on)
		},
		"no sleep": func(run func() error, max int, _ func(time.Duration), on func(int)) error {
			return runWithRetry(run, max, func(time.Duration) {}, on)
		},
		"no reset hook": func(run func() error, max int, sl func(time.Duration), _ func(int)) error {
			return runWithRetry(run, max, sl, func(int) {})
		},
		"swallows the final failure": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			if err := runWithRetry(run, max, sl, on); err != nil && isRetryableError(err) {
				return nil
			}
			return runWithRetry(run, max, sl, on)
		},
		"returns the first error": func(run func() error, max int, sl func(time.Duration), on func(int)) error {
			first := run()
			if first == nil {
				return nil
			}
			_ = runWithRetry(run, max, sl, on)
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

func (s *sinkMock) addProgressReceived(d int64) { s.received = append(s.received, d) }
func (s *sinkMock) setProgressTotal(t int64)    { s.totals = append(s.totals, t) }
func (s *sinkMock) setProgressPeak(p float64)   { s.peaks = append(s.peaks, p) }

func upd(status ytdlp.ProgressStatus, file string, done, total, frag, fragCount int) ytdlp.ProgressUpdate {
	return ytdlp.ProgressUpdate{Status: status, Filename: file, DownloadedBytes: done, TotalBytes: total, FragmentIndex: frag, FragmentCount: fragCount}
}

func TestYtdlpProgressCountsBytesPerFile(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading

	p.handle(upd(dl, "video.mp4", 100, 1000, 0, 0))
	p.handle(upd(dl, "video.mp4", 250, 1000, 0, 0))
	p.handle(upd(dl, "video.mp4", 250, 1000, 0, 0)) // no change: nothing to add
	assert.Equal(t, []int64{100, 150}, s.received, "the delta since the last update")

	p.handle(upd(dl, "audio.m4a", 40, 200, 0, 0)) // a new file starts from zero
	assert.Equal(t, []int64{100, 150, 40}, s.received)
	assert.Equal(t, int64(1200), s.totals[len(s.totals)-1], "video 1000 + audio 200")
}

func TestYtdlpProgressTotalsSumEveryFile(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.handle(upd(dl, "a", 1, 500, 0, 0))
	p.handle(upd(dl, "b", 1, 300, 0, 0))
	p.handle(upd(dl, "a", 2, 600, 0, 0)) // a's size is revised, not added twice
	assert.Equal(t, []int64{500, 800, 900}, s.totals)
}

func TestYtdlpProgressUnnamedFilesShareOneSlot(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.handle(upd(dl, "", 1, 700, 0, 0))
	p.handle(upd(dl, "", 2, 700, 0, 0))
	assert.Equal(t, []int64{700, 700}, s.totals, "an empty file name is one file, not many")
}

func TestYtdlpProgressHLSUsesFragments(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.handle(upd(dl, "f", 10, 0, 1, 4))
	p.handle(upd(dl, "f", 20, 0, 2, 4))
	p.handle(upd(dl, "f", 30, 0, 4, 4)) // 100% is held back to 99%
	assert.Equal(t, []float64{0.25, 0.5, 0.99}, s.peaks)
	assert.Empty(t, s.totals, "no byte total is known for HLS")
}

func TestYtdlpProgressIgnoresFinishedAndPostProcessing(t *testing.T) {
	for _, st := range []ytdlp.ProgressStatus{ytdlp.ProgressStatusFinished, ytdlp.ProgressStatusPostProcessing} {
		s := &sinkMock{}
		newYtdlpProgress(s).handle(upd(st, "f", 999, 999, 1, 2))
		assert.Empty(t, s.received, st)
		assert.Empty(t, s.totals, st)
		assert.Empty(t, s.peaks, st)
	}
}

func TestYtdlpProgressPrefersBytesOverFragments(t *testing.T) {
	s := &sinkMock{}
	newYtdlpProgress(s).handle(upd(ytdlp.ProgressStatusDownloading, "f", 5, 100, 1, 2))
	assert.Equal(t, []int64{100}, s.totals)
	assert.Empty(t, s.peaks)
}

func TestYtdlpProgressIgnoresIncompleteFragmentData(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.handle(upd(dl, "f", 0, 0, 0, 5)) // no index yet
	p.handle(upd(dl, "f", 0, 0, 3, 0)) // no count yet
	assert.Empty(t, s.peaks)
}

func TestYtdlpProgressResetForRetryRestartsTheCount(t *testing.T) {
	s := &sinkMock{}
	p := newYtdlpProgress(s)
	dl := ytdlp.ProgressStatusDownloading
	p.handle(upd(dl, "v", 500, 1000, 0, 0))
	p.resetForRetry()
	p.handle(upd(dl, "v", 120, 1000, 0, 0))
	assert.Equal(t, []int64{500, 120}, s.received, "after a retry the first update counts in full")

	// Without the reset the 120 would be swallowed as going backwards.
	s2 := &sinkMock{}
	p2 := newYtdlpProgress(s2)
	p2.handle(upd(dl, "v", 500, 1000, 0, 0))
	p2.handle(upd(dl, "v", 120, 1000, 0, 0))
	assert.Equal(t, []int64{500}, s2.received)
}

// progressScript is replayed against any progress handler; the fingerprint of
// what reached the sink is compared with the real one.
func progressFingerprint(newHandler func(progressSink) (handle func(ytdlp.ProgressUpdate), reset func())) string {
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
	want := progressFingerprint(func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
		p := newYtdlpProgress(s)
		return p.handle, p.resetForRetry
	})
	mutants := map[string]func(progressSink) (func(ytdlp.ProgressUpdate), func()){
		"no reset": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return p.handle, func() {}
		},
		"total of the latest file only": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				p.handle(u)
				p.totals = map[string]int64{}
			}, p.resetForRetry
		},
		"reports finished updates": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.Status == ytdlp.ProgressStatusFinished {
					u.Status = ytdlp.ProgressStatusDownloading
				}
				p.handle(u)
			}, p.resetForRetry
		},
		"fragment progress reaches 100": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.FragmentCount > 0 && u.FragmentIndex == u.FragmentCount && u.TotalBytes == 0 {
					s.setProgressPeak(1.0)
					return
				}
				p.handle(u)
			}, p.resetForRetry
		},
		"adds the running total, not the delta": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if u.Status == ytdlp.ProgressStatusDownloading && u.DownloadedBytes > 0 && u.TotalBytes > 0 {
					s.addProgressReceived(int64(u.DownloadedBytes))
					u.DownloadedBytes = 0
				}
				p.handle(u)
			}, p.resetForRetry
		},
		"ignores new files": func(s progressSink) (func(ytdlp.ProgressUpdate), func()) {
			p := newYtdlpProgress(s)
			return func(u ytdlp.ProgressUpdate) {
				if p.lastFile != "" {
					u.Filename = p.lastFile
				}
				p.handle(u)
			}, p.resetForRetry
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
		newYtdlpProgress(s).handle(upd(dl, "f", 1, 1, 0, 0))
		assert.Equal(t, []int64{1}, s.totals)
	})
	t.Run("a single fragment is complete but held at 99", func(t *testing.T) {
		s := &sinkMock{}
		newYtdlpProgress(s).handle(upd(dl, "f", 0, 0, 1, 1))
		assert.Equal(t, []float64{0.99}, s.peaks)
	})
	t.Run("the first fragment of many counts", func(t *testing.T) {
		s := &sinkMock{}
		newYtdlpProgress(s).handle(upd(dl, "f", 0, 0, 1, 10))
		assert.Equal(t, []float64{0.1}, s.peaks)
	})
	t.Run("an unnamed file after a retry counts from zero", func(t *testing.T) {
		s := &sinkMock{}
		p := newYtdlpProgress(s)
		p.handle(upd(dl, "", 500, 0, 0, 0))
		p.resetForRetry()
		p.handle(upd(dl, "", 100, 0, 0, 0))
		assert.Equal(t, []int64{500, 100}, s.received, "with no file name only the byte counter can tell the attempt restarted")
	})
	t.Run("a new file starts from zero even when it is smaller", func(t *testing.T) {
		s := &sinkMock{}
		p := newYtdlpProgress(s)
		p.handle(upd(dl, "a", 500, 0, 0, 0))
		p.handle(upd(dl, "b", 1, 0, 0, 0))
		assert.Equal(t, []int64{500, 1}, s.received)
	})
}
