package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readLog(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(LogFilePath) // #nosec G304 -- a path the test just created
	require.NoError(t, err)
	return string(b)
}

// A log is attached to bug reports as it is, so what reaches the file has no
// credentials in it, however the line was produced.
func TestTheLogFileNeverReceivesASecret(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	sink := initFileLogger()
	require.NotNil(t, sink)
	logFile = sink
	t.Cleanup(CloseLogFile)

	Debug("resolving", "url", "https://cdn.example/m/f.mp4?X-Amz-Signature=SUPERSECRETSIG&X-Amz-Expires=300")
	Debugf("ticket %s", "v1.AAAAAAAAAA.BBBBBBBBBB.CCCCCCCCCC")
	Warn("proxy http://alice:hunter2@10.0.0.1:8080 refused")
	n, err := GetLogFileWriter().Write([]byte("subprocess: Authorization: Bearer abcdefghijklmnop\n"))
	require.NoError(t, err)
	assert.Equal(t, len("subprocess: Authorization: Bearer abcdefghijklmnop\n"), n, "the caller is told its whole write was taken")

	CloseLogFile()
	got := readLog(t)
	for _, secret := range []string{"SUPERSECRETSIG", "AAAAAAAAAA.BBBBBBBBBB", "hunter2", "abcdefghijklmnop"} {
		assert.NotContains(t, got, secret)
	}
	assert.Contains(t, got, "X-Amz-Expires=300", "what is not a secret stays, so the log is still useful")
	assert.Contains(t, got, "[redacted]")
}

func TestTheLogNamesItsSessionAndSummarisesHowItWent(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	warnCount.Store(0)
	errorCount.Store(0)
	Logger = newConsoleLogger(io.Discard, colorprofile.NoTTY, false) // Warn and Error log only through a logger
	logFile = initFileLogger()
	require.NotNil(t, logFile)

	Warn("first")
	Warn("second")
	Error("boom")
	CloseLogFile()

	got := readLog(t)
	assert.Contains(t, got, "Session  "+SessionID())
	assert.Contains(t, got, "safe to attach to a bug report")
	assert.Contains(t, got, "2 warnings")
	assert.Contains(t, got, "1 error =====", "singular for one")
	assert.NotContains(t, got, "1 errors")
	assert.True(t, strings.HasSuffix(strings.TrimSpace(got), "====="), "the summary is the last line")
}

func TestSessionFooterPluralisation(t *testing.T) {
	warnCount.Store(1)
	errorCount.Store(0)
	t.Cleanup(func() { warnCount.Store(0) })
	f := sessionFooter(sessionStart.Add(95 * time.Second))
	assert.Contains(t, f, "ran 1m35s")
	assert.Contains(t, f, "1 warning ")
	assert.Contains(t, f, "0 errors")
}

func TestPruneLogsKeepsTheNewestAndOnlyOurOwnFiles(t *testing.T) {
	dir := t.TempDir()
	for i := range 35 {
		name := fmt.Sprintf("goanime_2026-10-%02d_10-00-00.log", i+1)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600))
	}
	stranger := filepath.Join(dir, "notes.log")
	other := filepath.Join(dir, "goanime_backup.txt")
	require.NoError(t, os.WriteFile(stranger, []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(other, []byte("x"), 0o600))

	pruneLogs(dir, "goanime_", 30, "")

	entries, _ := os.ReadDir(dir)
	var logs []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "goanime_2026") {
			logs = append(logs, e.Name())
		}
	}
	assert.Len(t, logs, 30)
	assert.NotContains(t, logs, "goanime_2026-10-01_10-00-00.log", "the oldest went")
	assert.Contains(t, logs, "goanime_2026-10-35_10-00-00.log", "the newest stayed")
	assert.FileExists(t, stranger, "a file that is not one of ours is never touched")
	assert.FileExists(t, other)
}

func TestPruneNeverRemovesTheFileJustCreated(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "goanime_2026-01-01_00-00-00.log") // sorts first, i.e. looks oldest
	require.NoError(t, os.WriteFile(current, []byte("x"), 0o600))
	for i := range 3 {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("goanime_2026-02-0%d_00-00-00.log", i+1)), []byte("x"), 0o600))
	}

	pruneLogs(dir, "goanime_", 2, current)

	assert.FileExists(t, current)
}

func TestCrashReportIsWrittenWithoutDebugAndIsSafeToShare(t *testing.T) {
	snapshotLogger(t)
	dir := withRedirectedHome(t)
	IsDebug = false
	sessionReport.Store("Application\n  Version  2.0.0\n")
	t.Cleanup(func() { sessionReport.Store("") })
	home, _ := os.UserHomeDir()
	stack := []byte("goroutine 1 [running]:\nmain.main()\n\t" + filepath.Join(home, "src", "main.go") + ":42 +0x1d\n")

	path, err := WriteCrashReport("index out of range with token=LEAKYTOKEN", stack)

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(path, dir), "it lives in the log folder")
	assert.Contains(t, filepath.Base(path), "crash_")
	assert.Contains(t, filepath.Base(path), SessionID())
	b, err := os.ReadFile(path) // #nosec G304 -- a file the test just made
	require.NoError(t, err)
	got := string(b)
	assert.Contains(t, got, "index out of range")
	assert.Contains(t, got, "goroutine 1 [running]")
	assert.Contains(t, got, "Version  2.0.0", "the environment travels with the crash")
	assert.NotContains(t, got, "LEAKYTOKEN")
	if len(home) > 1 {
		assert.NotContains(t, got, home, "the account name stays out")
	}
}

func TestCrashReportsAreKeptApartFromSessionLogs(t *testing.T) {
	dir := t.TempDir()
	for i := range 12 {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("crash_2026-10-%02d_00-00-00_abcd.log", i+1)), []byte("x"), 0o600))
	}
	for i := range 3 {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("goanime_2026-10-%02d_00-00-00.log", i+1)), []byte("x"), 0o600))
	}

	pruneLogs(dir, "goanime_", 30, "")
	pruneLogs(dir, "crash_", 10, "")

	crashes, _ := filepath.Glob(filepath.Join(dir, "crash_*.log"))
	sessions, _ := filepath.Glob(filepath.Join(dir, "goanime_*.log"))
	assert.Len(t, crashes, 10)
	assert.Len(t, sessions, 3, "session logs are not counted against the crash reports")
}

// A line names the code that logged it. Before, every line said logger.go,
// the wrapper, which told a reader nothing.
func TestLogLinesNameTheCallerNotTheWrapper(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	Logger = newConsoleLogger(io.Discard, colorprofile.NoTTY, false)
	logFile = initFileLogger()
	require.NotNil(t, logFile)

	Debug("marker-debug ")
	Debugf("marker-debugf %d", 1)
	Warn("marker-warn")
	Error("marker-error")
	CloseLogFile()

	got := readLog(t)
	for _, msg := range []string{"marker-debug ", "marker-debugf 1", "marker-warn", "marker-error"} {
		var line string
		for _, l := range strings.Split(got, "\n") {
			if strings.Contains(l, msg) {
				line = l
			}
		}
		require.NotEmpty(t, line, msg)
		assert.Contains(t, line, "logger_pro_test.go", "%q should be attributed to the test that logged it", msg)
		assert.NotContains(t, line, "util/logger.go", "%q must not be attributed to the wrapper", msg)
	}
}

// debugreport's goroutine writes the environment report some seconds after
// start, and a short run can close the log in between: a read of the global
// followed by a use of it would then dereference nil (and race).
func TestWriteSessionReportSurvivesTheLogClosingUnderIt(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	setLogFile(initFileLogger())
	require.NotNil(t, currentLogFile())

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 200 {
			WriteSessionReport("Application\n  Version  2.0.0\n")
		}
	}()
	for range 200 {
		CloseLogFile()
		setLogFile(initFileLogger())
	}
	<-done
	CloseLogFile()
}

// A subprocess's output reaches the log in chunks of any size. A secret split
// across two writes matches no rule in either half, so the sink must hold a
// partial line until it is complete.
func TestSubprocessOutputSplitAcrossWritesIsStillRedacted(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	setLogFile(initFileLogger())
	w := GetLogFileWriter()
	require.NotNil(t, w)

	for _, chunk := range []string{"mpv: GET https://h/p?tok", "en=SPLITSECRET1234&q=naruto\nAuthori", "zation: Bearer SPLITBEARER", "TOKEN99\nv1.AAAAAAAA", ".BBBBBBBB.CCCCCCCC done\n"} {
		_, err := w.Write([]byte(chunk))
		require.NoError(t, err)
	}
	CloseLogFile()

	got := readLog(t)
	for _, secret := range []string{"SPLITSECRET1234", "SPLITBEARER", "TOKEN99", "AAAAAAAA.BBBBBBBB"} {
		assert.NotContains(t, got, secret)
	}
	assert.Contains(t, got, "q=naruto", "what is not a secret survives")
}

func TestSubprocessOutputWithoutATrailingNewlineIsNotLost(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	setLogFile(initFileLogger())
	_, err := GetLogFileWriter().Write([]byte("last words without a newline"))
	require.NoError(t, err)
	CloseLogFile()

	assert.Contains(t, readLog(t), "last words without a newline")
}

// The crash handler must not itself panic: a collector that blows up costs the
// environment section, not the report.
func TestCrashReportSurvivesAFailingCollectorAndUsesAGoodOne(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = false
	sessionReport.Store("")
	t.Cleanup(func() { SetReportCollector(nil) })

	SetReportCollector(func(context.Context) string { panic("collector exploded") })
	path, err := WriteCrashReport("boom", []byte("stack"))
	require.NoError(t, err, "a panicking collector must not lose the crash report")
	b, _ := os.ReadFile(path) // #nosec G304 -- a file the test just made
	assert.Contains(t, string(b), "boom")

	SetReportCollector(func(context.Context) string { return "Tools\n  mpv  mpv v0.41.0\n" })
	path, err = WriteCrashReport("boom", []byte("stack"))
	require.NoError(t, err)
	b, _ = os.ReadFile(path) // #nosec G304 -- a file the test just made
	assert.Contains(t, string(b), "mpv v0.41.0", "the full report, with the tools, is what a maintainer needs")
}
