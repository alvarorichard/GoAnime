// Package debugreport records, in the debug log, what a maintainer needs to
// diagnose a bug: the machine, the terminal, the tools and how each provider
// answers. It lives outside cmd/goanime so that main.go stays self-contained
// and "go run cmd/goanime/main.go" keeps working.
package debugreport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/diag"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/cfsolver"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/version"
)

// debugTools are the programs GoAnime plays, downloads and converts with.
var debugTools = []diag.Tool{
	{Name: "mpv", VersionArgs: []string{"--version"}},
	{Name: "yt-dlp", VersionArgs: []string{"--version"}},
	{Name: "ffmpeg", VersionArgs: []string{"-version"}},
	{Name: "ffprobe", VersionArgs: []string{"-version"}},
}

// Start writes the environment report into the debug log and
// then checks each provider, all in the background: --debug must not make the
// program slower to open. Nothing here runs without --debug.
func Start() {
	if !util.IsDebug {
		return
	}
	startBackground(func() {
		runReport(45*time.Second, Report, util.WriteSessionReport, func(ctx context.Context) {
			probeProviders(ctx, registeredSources())
		})
	})
}

// startBackground runs f off the caller's goroutine; tests replace it to run
// f in place or to prove nothing was started.
var startBackground = func(f func()) { go f() }

// runReport builds the report and writes it, then runs the provider probe,
// all under one deadline.
func runReport(timeout time.Duration, build func(context.Context) string, write func(string), probe func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	write(build(ctx))
	probe(ctx)
}

// Report builds the full environment report as text: the machine, the
// terminal, the tools, the paths, each source's state and cfsolve. A crash
// report asks for it too (util.SetReportCollector), so the person who hit the
// crash does not have to have used --debug.
func Report(ctx context.Context) string {
	return diag.Collect(ctx, diag.Options{
		Version: version.Version,
		Tools:   debugTools,
		Extra:   []diag.Section{pathsSection(), sourcesSection(registeredSources()), cfsolveSection()},
	}).Text()
}

func registeredSources() []source.Descriptor {
	var out []source.Descriptor
	for _, kind := range source.Kinds() {
		if s, ok := source.Registered(kind); ok {
			out = append(out, s.Describe())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

func pathsSection() diag.Section {
	return pathsSectionFrom(util.GetLogDir(), diag.FreeSpace, os.UserConfigDir, os.UserCacheDir)
}

// pathsSectionFrom builds the Paths section from its sources, so the layout
// (which rows, in which order, when one is missing) is testable on any machine.
func pathsSectionFrom(logDir string, freeSpace func(string) (uint64, bool), configDir, cacheDir func() (string, error)) diag.Section {
	s := diag.Section{Title: "Paths"}
	add := func(k, v string) { s.Rows = append(s.Rows, diag.Row{Key: k, Value: v}) }
	add("Logs", logDir)
	if free, ok := freeSpace(logDir); ok {
		add("Free space", fmt.Sprintf("%.1f GiB where the logs are written", float64(free)/(1<<30)))
	}
	if cfg, err := configDir(); err == nil {
		add("Config", filepath.Join(cfg, "goanime"))
	}
	if cache, err := cacheDir(); err == nil {
		add("Cache", filepath.Join(cache, "goanime"))
	}
	return s
}

func sourcesSection(descs []source.Descriptor) diag.Section {
	s := diag.Section{Title: "Sources"}
	for _, d := range descs {
		state := "enabled"
		switch {
		case !source.IsEnabled(d):
			state = "disabled (GOANIME_DISABLED_SOURCES)"
		case !source.IsSearchEnabled(d):
			state = "off by default (GOANIME_ENABLED_SOURCES turns it on)"
		}
		s.Rows = append(s.Rows, diag.Row{Key: string(d.Kind), Value: state})
	}
	return s
}

// cfsolveSection says whether the tool that clears Cloudflare gates is there,
// since "Goyabu is missing" is usually this.
func cfsolveSection() diag.Section {
	s := diag.Section{Title: "cfsolve"}
	add := func(k, v string) { s.Rows = append(s.Rows, diag.Row{Key: k, Value: v}) }
	if sol, ok := cfsolver.Discover(); ok {
		add("Binary", sol.Bin)
	} else {
		add("Binary", "not installed")
	}
	if cfsolver.AutoDownloadEnabled() {
		add("Auto-download", "on")
	} else {
		add("Auto-download", "off (GOANIME_CFSOLVE_AUTODOWNLOAD=0)")
	}
	return s
}

// probeProviders asks each enabled provider's homepage how it answers and
// records the verdict. A provider that moved domain shows where it ended.
func probeProviders(ctx context.Context, descs []source.Descriptor) {
	client := &http.Client{Transport: netx.SafeScraperTransport(10 * time.Second), Timeout: 12 * time.Second}
	probeProvidersWith(ctx, client, descs, util.Debug)
}

// providerProbes picks the sources worth asking: those with a homepage to
// probe that take part in search (a source that is off is not worth a request).
func providerProbes(descs []source.Descriptor) []diag.Probe {
	var probes []diag.Probe
	for _, d := range descs {
		if d.ProbeURL != "" && source.IsSearchEnabled(d) {
			probes = append(probes, diag.Probe{Name: string(d.Kind), URL: d.ProbeURL})
		}
	}
	return probes
}

// probeFields turns one probe verdict into the key/value pairs of its log line.
func probeFields(r diag.Result) []any {
	kv := []any{"source", r.Name, "result", r.Outcome, "ms", r.Elapsed.Milliseconds()}
	if r.Status != 0 {
		kv = append(kv, "status", r.Status)
	}
	if r.Detail != "" {
		kv = append(kv, "detail", strings.ToLower(r.Detail))
	}
	if r.FinalURL != "" {
		kv = append(kv, "moved_to", r.FinalURL)
	}
	return kv
}

// probeProvidersWith is probeProviders over an injected client and logger.
func probeProvidersWith(ctx context.Context, client *http.Client, descs []source.Descriptor, logf func(any, ...any)) {
	probes := providerProbes(descs)
	if len(probes) == 0 {
		return
	}
	logf(fmt.Sprintf("Checking %d providers", len(probes)))
	diag.ProbeAll(ctx, client, probes, 10*time.Second, func(r diag.Result) {
		logf("Provider check", probeFields(r)...)
	})
}
