package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/diag"
	"github.com/alvarorichard/Goanime/v2/internal/version"
)

// WriteCrashReport saves what a maintainer needs to fix a panic — what
// happened, where, and on which machine — in a file of its own and returns its
// path. It is written whether or not --debug was on: the user who hit a crash
// did not know to ask for it. Credentials, tickets and the home folder are
// removed, so the file can be attached to an issue as it is.
func WriteCrashReport(recovered any, stack []byte) (path string, err error) {
	// This runs inside the program's own panic handler: a second panic here
	// would lose the very report it exists to write.
	defer func() {
		if r := recover(); r != nil {
			path, err = "", fmt.Errorf("writing the crash report failed: %v", r)
		}
	}()
	dir := GetLogDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	now := time.Now()
	path = filepath.Join(dir, fmt.Sprintf("crash_%s_%s.log", now.Format("2006-01-02_15-04-05"), sessionID))
	env := crashEnvironment()

	var b strings.Builder
	fmt.Fprintf(&b, "===== GoAnime crash report =====\nSession  %s\nWhen     %s\nRan for  %s\n\n",
		sessionID, now.Format("2006-01-02 15:04:05 -07:00"), now.Sub(sessionStart).Round(time.Second))
	fmt.Fprintf(&b, "Panic\n  %s\n\nStack (%s)\n%s\n", LogSafe(recovered), runtime.Version(), stack)
	fmt.Fprintf(&b, "\n--- Environment ---\n%s", env)
	fmt.Fprintf(&b, "\nWarnings %d, errors %d before the crash.\n", warnCount.Load(), errorCount.Load())
	if LogFilePath != "" {
		fmt.Fprintf(&b, "Debug log for this run: %s\n", LogFilePath)
	}

	if err := os.WriteFile(path, []byte(diag.Redact(b.String())), 0o600); err != nil {
		return "", err
	}
	pruneLogs(dir, "crash_", maxCrashLogs, path)
	return path, nil
}

var reportCollector atomic.Pointer[func(context.Context) string]

// SetReportCollector registers the function that builds the full environment
// report (machine, terminal, tools, sources). A crash report uses it when no
// debug session already collected one, so even a user who never heard of
// --debug sends the maintainer the versions of mpv and ffmpeg. The util
// package cannot build it itself: it does not know the sources.
func SetReportCollector(f func(context.Context) string) {
	if f == nil {
		reportCollector.Store(nil)
		return
	}
	reportCollector.Store(&f)
}

// crashEnvironment returns the environment for a crash report: the debug
// session's, else the registered collector's, else the machine and terminal
// alone. Each step is protected, so a failing collector costs a section.
func crashEnvironment() string {
	if env, _ := sessionReport.Load().(string); env != "" {
		return env
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if f := reportCollector.Load(); f != nil {
		if env := guardedCollect(func() string { return (*f)(ctx) }); env != "" {
			return env
		}
	}
	return guardedCollect(func() string {
		return diag.Collect(ctx, diag.Options{Version: version.Version, Timeout: time.Second}).Text()
	})
}

func guardedCollect(collect func() string) (env string) {
	defer func() {
		if r := recover(); r != nil {
			env = ""
		}
	}()
	return collect()
}
