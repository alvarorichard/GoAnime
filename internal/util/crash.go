package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/diag"
	"github.com/alvarorichard/Goanime/v2/internal/version"
)

// WriteCrashReport saves what a maintainer needs to fix a panic — what
// happened, where, and on which machine — in a file of its own and returns its
// path. It is written whether or not --debug was on: the user who hit a crash
// did not know to ask for it. Credentials, tickets and the home folder are
// removed, so the file can be attached to an issue as it is.
func WriteCrashReport(recovered any, stack []byte) (string, error) {
	dir := GetLogDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	now := time.Now()
	path := filepath.Join(dir, fmt.Sprintf("crash_%s_%s.log", now.Format("2006-01-02_15-04-05"), sessionID))

	env, _ := sessionReport.Load().(string)
	if env == "" {
		// No debug session collected one: gather the cheap, local parts now.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		env = diag.Collect(ctx, diag.Options{Version: version.Version, Timeout: time.Second}).Text()
	}

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
