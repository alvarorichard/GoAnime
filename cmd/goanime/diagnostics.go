package main

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

// startDebugDiagnostics writes the environment report into the debug log and
// then checks each provider, all in the background: --debug must not make the
// program slower to open. Nothing here runs without --debug.
func startDebugDiagnostics() {
	if !util.IsDebug {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		descs := registeredSources()
		report := diag.Collect(ctx, diag.Options{
			Version: version.Version,
			Tools:   debugTools,
			Extra:   []diag.Section{pathsSection(), sourcesSection(descs), cfsolveSection()},
		})
		util.WriteSessionReport(report.Text())
		probeProviders(ctx, descs)
	}()
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
	s := diag.Section{Title: "Paths"}
	add := func(k, v string) { s.Rows = append(s.Rows, diag.Row{Key: k, Value: v}) }
	logDir := util.GetLogDir()
	add("Logs", logDir)
	if free, ok := diag.FreeSpace(logDir); ok {
		add("Free space", fmt.Sprintf("%.1f GiB where the logs are written", float64(free)/(1<<30)))
	}
	if cfg, err := os.UserConfigDir(); err == nil {
		add("Config", filepath.Join(cfg, "goanime"))
	}
	if cache, err := os.UserCacheDir(); err == nil {
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
	var probes []diag.Probe
	for _, d := range descs {
		if d.ProbeURL != "" && source.IsSearchEnabled(d) {
			probes = append(probes, diag.Probe{Name: string(d.Kind), URL: d.ProbeURL})
		}
	}
	if len(probes) == 0 {
		return
	}
	client := &http.Client{Transport: netx.SafeScraperTransport(10 * time.Second), Timeout: 12 * time.Second}
	util.Debug(fmt.Sprintf("Checking %d providers", len(probes)))
	diag.ProbeAll(ctx, client, probes, 10*time.Second, func(r diag.Result) {
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
		util.Debug("Provider check", kv...)
	})
}
