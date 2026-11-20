package diag

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func section(t *testing.T, r Report, title string) Section {
	t.Helper()
	for _, s := range r.Sections {
		if s.Title == title {
			return s
		}
	}
	t.Fatalf("no %q section in %v", title, r.Sections)
	return Section{}
}

func value(s Section, key string) string {
	for _, r := range s.Rows {
		if r.Key == key {
			return r.Value
		}
	}
	return ""
}

func TestCollectCoversTheMachineTheTerminalAndTheEnvironment(t *testing.T) {
	env := map[string]string{
		"TERM": "xterm-256color", "COLORTERM": "truecolor", "TERM_PROGRAM": "WezTerm", "TERM_PROGRAM_VERSION": "2026",
		"LANG": "pt_BR.UTF-8", "SHELL": "/bin/zsh", "HTTPS_PROXY": "http://alice:s3cret@10.0.0.1:8080",
		"GOANIME_PLAIN": "1", "GOANIME_COOKIE_TOKEN": "https://x/?token=abc",
	}
	r := Collect(t.Context(), Options{
		Version: "2.0.0",
		Getenv:  func(k string) string { return env[k] },
		Environ: func() []string {
			var out []string
			for k, v := range env {
				out = append(out, k+"="+v)
			}
			return out
		},
		Now: func() time.Time { return time.Date(2026, 11, 2, 15, 49, 12, 0, time.FixedZone("-03", -3*3600)) },
	})

	app := section(t, r, "Application")
	assert.Equal(t, "2.0.0", value(app, "Version"))
	assert.Contains(t, value(app, "Go"), "go1")

	sys := section(t, r, "System")
	assert.Equal(t, runtime.GOOS+" / "+runtime.GOARCH, value(sys, "OS / arch"))
	assert.Equal(t, "-03 (UTC-03:00)", value(sys, "Time zone"))
	assert.Equal(t, "pt_BR.UTF-8", value(sys, "Locale"))

	tm := section(t, r, "Terminal")
	assert.Equal(t, "WezTerm 2026", value(tm, "Program"))
	assert.Equal(t, "/bin/zsh", value(tm, "Shell"))

	en := section(t, r, "Environment")
	assert.Equal(t, "1", value(en, "GOANIME_PLAIN"))
	assert.Equal(t, "https://x/?token=[redacted]", value(en, "GOANIME_COOKIE_TOKEN"))
	assert.Equal(t, "http://[redacted]@10.0.0.1:8080", value(en, "HTTPS_PROXY"), "a proxy is shown, its password is not")
}

func TestReportTextNeverCarriesSecretsOrTheHomeDirectory(t *testing.T) {
	home, _ := os.UserHomeDir()
	r := Report{Sections: []Section{{Title: "Paths", Rows: []Row{
		{"Logs", filepath.Join(home, ".local", "share", "goanime", "logs")},
		{"Proxy", "http://bob:hunter2@proxy:3128"},
	}}}}

	text := r.Text()

	assert.NotContains(t, text, "hunter2")
	if len(home) > 1 {
		assert.NotContains(t, text, home)
		assert.Contains(t, text, "~")
	}
	assert.Contains(t, text, "Paths\n  Logs ")
}

func writeScript(t *testing.T, name, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stand-ins do not run on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o700)) // #nosec G306 -- a test executable
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return p
}

func TestToolsReportTheFirstVersionLineAndWhereTheyLive(t *testing.T) {
	path := writeScript(t, "faketool", `echo "faketool 9.9.9 (build x)"; echo "second line is dropped"`)

	r := Collect(t.Context(), Options{Tools: []Tool{{Name: "faketool", VersionArgs: []string{"--version"}}, {Name: "no-such-tool-xyz", VersionArgs: []string{"-v"}}}})

	tools := section(t, r, "Tools")
	assert.Equal(t, "faketool 9.9.9 (build x)  ["+path+"]", value(tools, "faketool"))
	assert.Equal(t, "not found", value(tools, "no-such-tool-xyz"))
}

// A tool that hangs must cost the timeout, not the program.
func TestAHungToolCostsOnlyTheTimeout(t *testing.T) {
	writeScript(t, "hangtool", `sleep 30`)

	start := time.Now()
	r := Collect(t.Context(), Options{Timeout: 300 * time.Millisecond, Tools: []Tool{{Name: "hangtool", VersionArgs: []string{"--version"}}}})

	assert.Less(t, time.Since(start), 5*time.Second)
	assert.Contains(t, value(section(t, r, "Tools"), "hangtool"), "timed out")
}

func TestToolsRunConcurrently(t *testing.T) {
	writeScript(t, "slowtool", `sleep 1; echo ok`)

	start := time.Now()
	Collect(context.Background(), Options{Tools: []Tool{
		{Name: "slowtool", VersionArgs: []string{"x"}}, {Name: "slowtool", VersionArgs: []string{"y"}}, {Name: "slowtool", VersionArgs: []string{"z"}},
	}})

	assert.Less(t, time.Since(start), 2500*time.Millisecond, "three one-second tools together take about one second")
}

func TestExtraSectionsAreAppendedAsGiven(t *testing.T) {
	r := Collect(t.Context(), Options{Extra: []Section{{Title: "Sources", Rows: []Row{{"AnimeFire", "enabled"}}}}})
	assert.Equal(t, "enabled", value(section(t, r, "Sources"), "AnimeFire"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(r.Text()), "AnimeFire  enabled"))
}
