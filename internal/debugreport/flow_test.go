package debugreport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/diag"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Start / runReport -------------------------------------------------------

func TestStartDoesNothingWithoutDebug(t *testing.T) {
	prevDebug, prevBG := util.IsDebug, startBackground
	t.Cleanup(func() { util.IsDebug, startBackground = prevDebug, prevBG })
	started := 0
	startBackground = func(func()) { started++ }

	util.IsDebug = false
	Start()
	assert.Zero(t, started, "--debug is what opts in; without it nothing starts")

	util.IsDebug = true
	Start()
	assert.Equal(t, 1, started, "with --debug exactly one background job starts")
}

// runReportFailures judges any runReport against the ordering and deadline
// contract, observed through mock build/write/probe functions.
func runReportFailures(impl func(time.Duration, func(context.Context) string, func(string), func(context.Context))) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	var order []string
	var buildCtx, probeCtx context.Context
	var written string
	impl(30*time.Second,
		func(ctx context.Context) string { order = append(order, "build"); buildCtx = ctx; return "REPORT" },
		func(s string) { order = append(order, "write"); written = s },
		func(ctx context.Context) { order = append(order, "probe"); probeCtx = ctx },
	)
	if fmt.Sprint(order) != "[build write probe]" {
		f = append(f, fmt.Sprintf("order %v", order))
	}
	if written != "REPORT" {
		f = append(f, fmt.Sprintf("wrote %q", written))
	}
	if buildCtx == nil || probeCtx == nil {
		return append(f, "context not passed")
	}
	if buildCtx != probeCtx {
		f = append(f, "build and probe do not share one deadline")
	}
	dl, ok := probeCtx.Deadline()
	if !ok || time.Until(dl) > 30*time.Second || time.Until(dl) < 25*time.Second {
		f = append(f, fmt.Sprintf("deadline %v ok=%v", time.Until(dl), ok))
	}
	if probeCtx.Err() == nil {
		f = append(f, "context not cancelled after the run")
	}
	return f
}

func TestRunReport(t *testing.T) {
	assert.Empty(t, runReportFailures(runReport))
}

func TestRunReportMeta(t *testing.T) {
	require.Empty(t, runReportFailures(runReport))
	type fn = func(time.Duration, func(context.Context) string, func(string), func(context.Context))
	mutants := map[string]fn{
		"probe before write": func(d time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), d)
			defer cancel()
			text := b(ctx)
			p(ctx)
			w(text)
		},
		"never probes": func(d time.Duration, b func(context.Context) string, w func(string), _ func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), d)
			defer cancel()
			w(b(ctx))
		},
		"never writes": func(d time.Duration, b func(context.Context) string, _ func(string), p func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), d)
			defer cancel()
			_ = b(ctx)
			p(ctx)
		},
		"no deadline": func(_ time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			w(b(ctx))
			p(ctx)
		},
		"fresh deadline per step": func(d time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			c1, x1 := context.WithTimeout(context.Background(), d)
			defer x1()
			w(b(c1))
			c2, x2 := context.WithTimeout(context.Background(), d)
			defer x2()
			p(c2)
		},
		"context leaked": func(d time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), d)
			_ = cancel // deliberately never called: the mutant under test leaks the context
			w(b(ctx))
			p(ctx)
		},
		"wrong timeout": func(_ time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()
			w(b(ctx))
			p(ctx)
		},
		"writes empty": func(d time.Duration, b func(context.Context) string, w func(string), p func(context.Context)) {
			ctx, cancel := context.WithTimeout(context.Background(), d)
			defer cancel()
			_ = b(ctx)
			w("")
			p(ctx)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, runReportFailures(m), "mutant %q survived", name)
	}
}

// ---- Report ------------------------------------------------------------------

func TestReportCarriesEverySection(t *testing.T) {
	t.Setenv("GOANIME_CFSOLVE_BIN", "/nonexistent/challenge")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	text := Report(ctx)

	for _, want := range []string{"Paths", "Logs", "Sources", "cfsolve", version.Version} {
		assert.Contains(t, text, want)
	}
	if home, err := os.UserHomeDir(); err == nil && len(home) > 3 {
		assert.NotContains(t, text, home, "the home directory is scrubbed (reports are shared)")
	}
}

// ---- pathsSectionFrom --------------------------------------------------------

func okDir(p string) func() (string, error) { return func() (string, error) { return p, nil } }
func badDir() func() (string, error)        { return func() (string, error) { return "", os.ErrNotExist } }
func freeGiB(g float64) func(string) (uint64, bool) {
	return func(string) (uint64, bool) { return uint64(g * (1 << 30)), true }
}
func noFree(string) (uint64, bool) { return 0, false }

func rowsOf(s diag.Section) []string {
	var out []string
	for _, r := range s.Rows {
		out = append(out, r.Key+"="+r.Value)
	}
	return out
}

func pathsFailures(impl func(string, func(string) (uint64, bool), func() (string, error), func() (string, error)) diag.Section) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	check := func(name string, got, want []string) {
		if fmt.Sprint(got) != fmt.Sprint(want) {
			f = append(f, fmt.Sprintf("%s: got %v want %v", name, got, want))
		}
	}
	j := filepath.Join
	all := impl("/logs", freeGiB(2.5), okDir("/cfg"), okDir("/cache"))
	if all.Title != "Paths" {
		f = append(f, "title "+all.Title)
	}
	check("all", rowsOf(all), []string{
		"Logs=/logs", "Free space=2.5 GiB where the logs are written", "Config=" + j("/cfg", "goanime"), "Cache=" + j("/cache", "goanime"),
	})
	check("no free space", rowsOf(impl("/logs", noFree, okDir("/cfg"), okDir("/cache"))),
		[]string{"Logs=/logs", "Config=" + j("/cfg", "goanime"), "Cache=" + j("/cache", "goanime")})
	check("no config", rowsOf(impl("/logs", noFree, badDir(), okDir("/cache"))),
		[]string{"Logs=/logs", "Cache=" + j("/cache", "goanime")})
	check("no cache", rowsOf(impl("/logs", noFree, okDir("/cfg"), badDir())),
		[]string{"Logs=/logs", "Config=" + j("/cfg", "goanime")})
	check("only logs", rowsOf(impl("/l", noFree, badDir(), badDir())), []string{"Logs=/l"})
	check("rounding", rowsOf(impl("/l", freeGiB(0.04), badDir(), badDir())), []string{"Logs=/l", "Free space=0.0 GiB where the logs are written"})
	return f
}

func TestPathsSectionFrom(t *testing.T) {
	assert.Empty(t, pathsFailures(pathsSectionFrom))
}

func TestPathsSectionAskedForTheLogDirectory(t *testing.T) {
	var asked []string
	pathsSectionFrom("/where/logs/go", func(p string) (uint64, bool) { asked = append(asked, p); return 0, false }, badDir(), badDir())
	assert.Equal(t, []string{"/where/logs/go"}, asked, "free space is measured where the logs are written")
}

func TestPathsSectionOnThisMachine(t *testing.T) {
	s := pathsSection()
	assert.Equal(t, "Paths", s.Title)
	require.NotEmpty(t, s.Rows)
	assert.Equal(t, "Logs", s.Rows[0].Key)
	assert.Equal(t, util.GetLogDir(), s.Rows[0].Value)
}

func TestPathsSectionFromMeta(t *testing.T) {
	require.Empty(t, pathsFailures(pathsSectionFrom))
	type fn = func(string, func(string) (uint64, bool), func() (string, error), func() (string, error)) diag.Section
	mutants := map[string]fn{
		"config and cache swapped": func(l string, f func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			return pathsSectionFrom(l, f, k, c)
		},
		"free space ignored": func(l string, _ func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			return pathsSectionFrom(l, noFree, c, k)
		},
		"free space in GB not GiB": func(l string, f func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			return pathsSectionFrom(l, func(p string) (uint64, bool) { v, ok := f(p); return v / 2, ok }, c, k)
		},
		"errors still printed": func(l string, f func(string) (uint64, bool), _, _ func() (string, error)) diag.Section {
			return pathsSectionFrom(l, f, okDir(""), okDir(""))
		},
		"no goanime subfolder": func(l string, f func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			s := pathsSectionFrom(l, f, c, k)
			for i := range s.Rows {
				s.Rows[i].Value = strings.ReplaceAll(s.Rows[i].Value, "goanime", "")
			}
			return s
		},
		"wrong title": func(l string, f func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			s := pathsSectionFrom(l, f, c, k)
			s.Title = "Path"
			return s
		},
		"logs row missing": func(l string, f func(string) (uint64, bool), c, k func() (string, error)) diag.Section {
			s := pathsSectionFrom(l, f, c, k)
			s.Rows = s.Rows[1:]
			return s
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, pathsFailures(m), "mutant %q survived", name)
	}
}

// ---- providerProbes ----------------------------------------------------------

func probeNames(ps []diag.Probe) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Name+"@"+p.URL)
	}
	return out
}

func providerProbeFailures(impl func([]source.Descriptor) []diag.Probe) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	descs := []source.Descriptor{
		{Kind: "ProbeA", ProbeURL: "https://a.test"},
		{Kind: "NoURL"},
		{Kind: "OptIn", ProbeURL: "https://optin.test", DefaultDisabled: true},
		{Kind: "ProbeB", ProbeURL: "https://b.test"},
	}
	got := probeNames(impl(descs))
	if fmt.Sprint(got) != "[ProbeA@https://a.test ProbeB@https://b.test]" {
		f = append(f, fmt.Sprintf("default: %v", got))
	}
	if len(impl(nil)) != 0 {
		f = append(f, "nil descriptors yield probes")
	}
	return f
}

func TestProviderProbes(t *testing.T) {
	assert.Empty(t, providerProbeFailures(providerProbes))

	t.Run("a kill-switched source is not probed", func(t *testing.T) {
		t.Setenv("GOANIME_DISABLED_SOURCES", "ProbeA")
		got := probeNames(providerProbes([]source.Descriptor{
			{Kind: "ProbeA", ProbeURL: "https://a.test"}, {Kind: "ProbeB", ProbeURL: "https://b.test"},
		}))
		assert.Equal(t, []string{"ProbeB@https://b.test"}, got)
	})

	t.Run("an opted-in source is probed", func(t *testing.T) {
		t.Setenv("GOANIME_ENABLED_SOURCES", "OptIn")
		got := probeNames(providerProbes([]source.Descriptor{{Kind: "OptIn", ProbeURL: "https://optin.test", DefaultDisabled: true}}))
		assert.Equal(t, []string{"OptIn@https://optin.test"}, got)
	})
}

func TestProviderProbesMeta(t *testing.T) {
	require.Empty(t, providerProbeFailures(providerProbes))
	mutants := map[string]func([]source.Descriptor) []diag.Probe{
		"probes sources without a URL": func(ds []source.Descriptor) []diag.Probe {
			var out []diag.Probe
			for _, d := range ds {
				if source.IsSearchEnabled(d) {
					out = append(out, diag.Probe{Name: string(d.Kind), URL: d.ProbeURL})
				}
			}
			return out
		},
		"probes off-by-default sources": func(ds []source.Descriptor) []diag.Probe {
			var out []diag.Probe
			for _, d := range ds {
				if d.ProbeURL != "" {
					out = append(out, diag.Probe{Name: string(d.Kind), URL: d.ProbeURL})
				}
			}
			return out
		},
		"drops the last": func(ds []source.Descriptor) []diag.Probe {
			p := providerProbes(ds)
			if len(p) > 0 {
				p = p[:len(p)-1]
			}
			return p
		},
		"name and URL swapped": func(ds []source.Descriptor) []diag.Probe {
			p := providerProbes(ds)
			for i := range p {
				p[i].Name, p[i].URL = p[i].URL, p[i].Name
			}
			return p
		},
		"nothing": func([]source.Descriptor) []diag.Probe { return nil },
		"reversed": func(ds []source.Descriptor) []diag.Probe {
			p := providerProbes(ds)
			for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
				p[i], p[j] = p[j], p[i]
			}
			return p
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, providerProbeFailures(m), "mutant %q survived", name)
	}
}

// ---- probeFields -------------------------------------------------------------

func probeFieldFailures(impl func(diag.Result) []any) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprintf("panic: %v", r))
		}
	}()
	check := func(name string, got []any, want string) {
		if fmt.Sprint(got) != want {
			f = append(f, fmt.Sprintf("%s: got %v want %s", name, got, want))
		}
	}
	check("minimal", impl(diag.Result{Name: "S", Outcome: "ok", Elapsed: 120 * time.Millisecond}), "[source S result ok ms 120]")
	check("status", impl(diag.Result{Name: "S", Outcome: "blocked", Status: 403}), "[source S result blocked ms 0 status 403]")
	check("detail lowercased", impl(diag.Result{Name: "S", Outcome: "tls", Detail: "Bad Certificate"}), "[source S result tls ms 0 detail bad certificate]")
	check("moved", impl(diag.Result{Name: "S", Outcome: "ok", FinalURL: "https://new.test"}), "[source S result ok ms 0 moved_to https://new.test]")
	check("everything", impl(diag.Result{Name: "S", Outcome: "challenge", Status: 503, Detail: "CF", FinalURL: "https://x", Elapsed: time.Second}),
		"[source S result challenge ms 1000 status 503 detail cf moved_to https://x]")
	return f
}

func TestProbeFields(t *testing.T) {
	assert.Empty(t, probeFieldFailures(probeFields))
}

func TestProbeFieldsMeta(t *testing.T) {
	require.Empty(t, probeFieldFailures(probeFields))
	mutants := map[string]func(diag.Result) []any{
		"status always": func(r diag.Result) []any { return append(probeFields(r)[:6:6], "status", r.Status) },
		"detail not lowercased": func(r diag.Result) []any {
			kv := probeFields(r)
			for i, v := range kv {
				if s, ok := v.(string); ok && strings.EqualFold(s, r.Detail) && r.Detail != "" {
					kv[i] = r.Detail
				}
			}
			return kv
		},
		"seconds not ms": func(r diag.Result) []any {
			kv := probeFields(r)
			kv[5] = int64(r.Elapsed.Seconds())
			return kv
		},
		"moved dropped": func(r diag.Result) []any {
			r.FinalURL = ""
			return probeFields(r)
		},
		"outcome and name swapped": func(r diag.Result) []any {
			r.Name, r.Outcome = r.Outcome, r.Name
			return probeFields(r)
		},
		"status dropped": func(r diag.Result) []any {
			r.Status = 0
			return probeFields(r)
		},
		"detail dropped": func(r diag.Result) []any {
			r.Detail = ""
			return probeFields(r)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, probeFieldFailures(m), "mutant %q survived", name)
	}
}

// ---- probeProvidersWith (mock host + mock logger) --------------------------------

type logLine struct {
	msg string
	kv  map[string]any
}

type logRecorder struct {
	mu    sync.Mutex
	lines []logLine
}

func (l *logRecorder) log(msg any, kv ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	m := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[fmt.Sprint(kv[i])] = kv[i+1]
	}
	l.lines = append(l.lines, logLine{fmt.Sprint(msg), m})
}

func TestProbeProvidersWithRecordsOneVerdictPerProvider(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"), "probes never send credentials")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(up.Close)
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	t.Cleanup(down.Close)

	rec := &logRecorder{}
	probeProvidersWith(context.Background(), up.Client(), []source.Descriptor{
		{Kind: "Up", ProbeURL: up.URL}, {Kind: "Down", ProbeURL: down.URL}, {Kind: "Silent"},
	}, rec.log)

	require.Len(t, rec.lines, 3, "one header line plus a verdict for each probed provider")
	assert.Equal(t, "Checking 2 providers", rec.lines[0].msg, "the unprobeable source is not counted")
	verdicts := map[string]logLine{}
	for _, l := range rec.lines[1:] {
		assert.Equal(t, "Provider check", l.msg)
		verdicts[fmt.Sprint(l.kv["source"])] = l
	}
	require.Contains(t, verdicts, "Up")
	require.Contains(t, verdicts, "Down")
	assert.Equal(t, "ok", fmt.Sprint(verdicts["Up"].kv["result"]))
	assert.NotEqual(t, "ok", fmt.Sprint(verdicts["Down"].kv["result"]), "a 503 is not ok")
	assert.Equal(t, 503, verdicts["Down"].kv["status"])
}

func TestProbeProvidersWithIsSilentWithNothingToProbe(t *testing.T) {
	rec := &logRecorder{}
	probeProvidersWith(context.Background(), http.DefaultClient, nil, rec.log)
	probeProvidersWith(context.Background(), http.DefaultClient, []source.Descriptor{{Kind: "NoURL"}}, rec.log)
	assert.Empty(t, rec.lines, "no probes, no log lines, no requests")
}

func TestProbeProvidersWithStopsWhenTheContextIsCancelled(t *testing.T) {
	hit := make(chan struct{}, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit <- struct{}{} }))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := &logRecorder{}
	done := make(chan struct{})
	go func() {
		probeProvidersWith(ctx, srv.Client(), []source.Descriptor{{Kind: "P", ProbeURL: srv.URL}}, rec.log)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("probing ignored the cancelled context")
	}
	assert.Empty(t, hit, "no request leaves after cancellation")
}
