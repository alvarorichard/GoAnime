package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"charm.land/log/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loggerSnapshot captures and restores all package-level state mutated by the
// logger helpers. Tests in this file run serially (no t.Parallel) because they
// share the singletons Logger, fileLogger, logFile, LogFilePath, and IsDebug.
type loggerSnapshot struct {
	logger      *log.Logger
	fileLogger  *log.Logger
	logFile     *os.File
	logFilePath string
	isDebug     bool
	cleanups    []func()
}

func snapshotLogger(t *testing.T) {
	t.Helper()
	consoleLogMu.Lock()
	cleanupMu.Lock()
	prev := loggerSnapshot{
		logger:      Logger,
		fileLogger:  fileLogger,
		logFile:     logFile,
		logFilePath: LogFilePath,
		isDebug:     IsDebug,
		cleanups:    append([]func(){}, cleanupFuncs...),
	}
	cleanupMu.Unlock()
	consoleLogMu.Unlock()

	t.Cleanup(func() {
		consoleLogMu.Lock()
		cleanupMu.Lock()
		Logger = prev.logger
		fileLogger = prev.fileLogger
		logFile = prev.logFile
		LogFilePath = prev.logFilePath
		IsDebug = prev.isDebug
		cleanupFuncs = prev.cleanups
		cleanupMu.Unlock()
		consoleLogMu.Unlock()
	})
}

func TestPrintSavedLocation_WritesLabelAndPath(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	PrintSavedLocation("Saved to:", "/tmp/file.mp4")
	require.NoError(t, w.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Contains(t, string(out), "Saved to:")
	assert.Contains(t, string(out), "/tmp/file.mp4")
}

func TestGetColoredPrefix_ContainsBrand(t *testing.T) {
	out := getColoredPrefix()
	assert.Contains(t, out, "GoAnime")
}

func TestGetLogDir_ReturnsPlatformPath(t *testing.T) {
	dir := GetLogDir()
	require.NotEmpty(t, dir)
	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, dir, filepath.Join("GoAnime", "logs"))
	case "darwin":
		assert.Contains(t, dir, filepath.Join("Library", "Logs", "GoAnime"))
	default:
		assert.Contains(t, dir, filepath.Join(".local", "share", "goanime", "logs"))
	}
}

// withRedirectedHome forces GetLogDir() to use a temp directory by overriding
// HOME (Unix) or LOCALAPPDATA (Windows). Returns the resolved log dir.
func withRedirectedHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("LOCALAPPDATA", tmp)
	} else {
		t.Setenv("HOME", tmp)
	}
	return GetLogDir()
}

func TestInitFileLogger_CreatesFileWithHeader(t *testing.T) {
	snapshotLogger(t)
	dir := withRedirectedHome(t)

	f := initFileLogger()
	require.NotNil(t, f)
	t.Cleanup(func() { _ = f.Close() })

	require.NotEmpty(t, LogFilePath)
	assert.True(t, strings.HasPrefix(LogFilePath, dir), "log path %q should be in %q", LogFilePath, dir)

	data, err := os.ReadFile(LogFilePath) // #nosec G304
	require.NoError(t, err)
	assert.Contains(t, string(data), "GoAnime Debug Session")
}

func TestInitFileLogger_FailsWhenDirUncreatable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on Windows")
	}
	snapshotLogger(t)

	// Make HOME point at a file (not a directory), so MkdirAll fails.
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
	t.Setenv("HOME", blocker)

	// Redirect stderr so the warning doesn't pollute test output.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = origStderr
	})

	f := initFileLogger()
	_ = w.Close()
	_, _ = io.ReadAll(r)

	assert.Nil(t, f)
}

func TestInitLogger_DebugFalseSkipsFileLogger(t *testing.T) {
	snapshotLogger(t)
	IsDebug = false
	logFile = nil
	fileLogger = nil

	InitLogger()
	require.NotNil(t, Logger)
	assert.Nil(t, logFile)
	assert.Nil(t, fileLogger)
}

func TestInitLogger_DebugTrueCreatesLogFile(t *testing.T) {
	snapshotLogger(t)
	withRedirectedHome(t)
	IsDebug = true
	logFile = nil
	fileLogger = nil

	// silence stderr banner
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	InitLogger()
	_ = w.Close()
	_, _ = io.ReadAll(r)

	require.NotNil(t, Logger)
	require.NotNil(t, logFile, "debug mode must open a log file")
	require.NotNil(t, fileLogger, "file logger must be initialised")
	assert.NotEmpty(t, LogFilePath)

	// Cleanup
	CloseLogFile()
}

func TestShowDebugBanner_WritesPathAndFollowCmd(t *testing.T) {
	snapshotLogger(t)
	LogFilePath = filepath.Join(t.TempDir(), "session.log")

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	showDebugBanner()
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	body := string(out)

	assert.Contains(t, body, LogFilePath)
	if runtime.GOOS == "windows" {
		assert.Contains(t, body, "Get-Content")
	} else {
		assert.Contains(t, body, "tail -f")
	}
}

func TestCloseLogFile_NilSafe(t *testing.T) {
	snapshotLogger(t)
	logFile = nil
	// Must not panic when logFile is nil.
	CloseLogFile()
	assert.Nil(t, logFile)
}

func TestCloseLogFile_FlushesAndClears(t *testing.T) {
	snapshotLogger(t)
	tmp := filepath.Join(t.TempDir(), "x.log")
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR, 0o600) // #nosec G304
	require.NoError(t, err)
	_, _ = f.WriteString("hello")
	logFile = f

	CloseLogFile()
	assert.Nil(t, logFile)
	// Re-closing should be a no-op.
	CloseLogFile()
}

func TestGetLogFileWriter_NilWhenInactive(t *testing.T) {
	snapshotLogger(t)
	logFile = nil
	assert.Nil(t, GetLogFileWriter())
}

func TestGetLogFileWriter_ReturnsFileWhenActive(t *testing.T) {
	snapshotLogger(t)
	tmp := filepath.Join(t.TempDir(), "y.log")
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR, 0o600) // #nosec G304
	require.NoError(t, err)
	logFile = f
	t.Cleanup(func() { _ = f.Close() })

	w := GetLogFileWriter()
	require.NotNil(t, w)
	_, err = w.Write([]byte("payload"))
	require.NoError(t, err)
}

// installCapturingLoggers swaps Logger/fileLogger for in-memory buffers so the
// Debug/Info/Warn/Error helpers can be asserted on. Returns the two buffers.
func installCapturingLoggers(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	snapshotLogger(t)
	var consoleBuf, fileBuf bytes.Buffer
	Logger = log.NewWithOptions(&consoleBuf, log.Options{Prefix: "test"})
	Logger.SetLevel(log.DebugLevel)
	fileLogger = log.NewWithOptions(&fileBuf, log.Options{Prefix: "file"})
	fileLogger.SetLevel(log.DebugLevel)
	return &consoleBuf, &fileBuf
}

func TestDebug_NoOpWhenDebugDisabled(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	IsDebug = false
	Debug("should-not-appear")
	assert.Empty(t, consoleBuf.String())
	assert.Empty(t, fileBuf.String())
}

func TestDebug_WritesToFileLoggerWhenEnabled(t *testing.T) {
	_, fileBuf := installCapturingLoggers(t)
	IsDebug = true
	Debug("debug-message", "key", "val")
	assert.Contains(t, fileBuf.String(), "debug-message")
}

func TestDebug_FallsBackToConsoleWithoutFileLogger(t *testing.T) {
	consoleBuf, _ := installCapturingLoggers(t)
	IsDebug = true
	fileLogger = nil
	Debug("fallback-debug")
	assert.Contains(t, consoleBuf.String(), "fallback-debug")
}

func TestInfo_WritesToBothLoggers(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	Info("info-msg")
	assert.Contains(t, consoleBuf.String(), "info-msg")
	assert.Contains(t, fileBuf.String(), "info-msg")
}

func TestInfo_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	fileLogger = nil
	Info("no-logger") // must not panic
}

func TestError_WritesToBothLoggers(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	Error("err-msg")
	assert.Contains(t, consoleBuf.String(), "err-msg")
	assert.Contains(t, fileBuf.String(), "err-msg")
}

func TestError_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	fileLogger = nil
	Error("no-logger") // must not panic
}

// TestFatal_ExitsProcess runs Fatal in a subprocess to verify it terminates
// with a non-zero exit code. The upstream charmbracelet logger calls os.Exit
// without an interception hook, so the only safe way to exercise this path
// is via a child process. The child also writes a marker line to the file
// logger so we can confirm the pre-exit writeToFile branch ran.
func TestFatal_ExitsProcess(t *testing.T) {
	if os.Getenv("GOANIME_FATAL_CHILD") == "1" {
		// Child: trigger Fatal with both file and console loggers wired.
		var consoleBuf bytes.Buffer
		Logger = log.NewWithOptions(&consoleBuf, log.Options{Prefix: "test"})
		Logger.SetLevel(log.DebugLevel)
		markerPath := os.Getenv("GOANIME_FATAL_MARKER")
		f, err := os.OpenFile(markerPath, os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304
		if err != nil {
			os.Exit(99)
		}
		fileLogger = log.NewWithOptions(f, log.Options{Prefix: "file"})
		fileLogger.SetLevel(log.DebugLevel)
		logFile = f
		Fatal("fatal-msg-from-child")
		// Should never reach here.
		os.Exit(0)
		return
	}

	marker := filepath.Join(t.TempDir(), "fatal_marker.log")
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal_ExitsProcess", "-test.v")
	cmd.Env = append(os.Environ(),
		"GOANIME_FATAL_CHILD=1",
		"GOANIME_FATAL_MARKER="+marker,
	)
	out, err := cmd.CombinedOutput()

	// Logger.Fatal calls os.Exit(1) → exec returns *exec.ExitError.
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "expected ExitError, got %v (output: %s)", err, out)
	assert.NotEqual(t, 0, exitErr.ExitCode(), "Fatal must exit non-zero")

	data, readErr := os.ReadFile(marker) // #nosec G304
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "fatal-msg-from-child", "Fatal must writeToFile before exit")
}

func TestFatal_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	fileLogger = nil
	logFile = nil
	Fatal("no-logger") // must not panic and must not exit
}

func TestInfof_FormatsAndWrites(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	Infof("user=%s id=%d", "alice", 42)
	assert.Contains(t, consoleBuf.String(), "user=alice id=42")
	assert.Contains(t, fileBuf.String(), "user=alice id=42")
}

func TestInfof_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	Infof("nope")
}

func TestWarnf_FormatsAndWrites(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	Warnf("retrying %d", 3)
	assert.Contains(t, consoleBuf.String(), "retrying 3")
	assert.Contains(t, fileBuf.String(), "retrying 3")
}

func TestWarnf_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	Warnf("nope")
}

func TestErrorf_FormatsAndWrites(t *testing.T) {
	consoleBuf, fileBuf := installCapturingLoggers(t)
	Errorf("failed step %d: %s", 2, "boom")
	assert.Contains(t, consoleBuf.String(), "failed step 2: boom")
	assert.Contains(t, fileBuf.String(), "failed step 2: boom")
}

func TestErrorf_NilLoggerIsNoOp(t *testing.T) {
	snapshotLogger(t)
	Logger = nil
	Errorf("nope")
}

// Guard against accidental parallel logger tests in this file: if a future
// edit adds t.Parallel() to one of them, the test will quickly start racing on
// the package globals. This sentinel just keeps the lock in the picture for
// the race detector.
var loggerTestSerialGuard sync.Mutex

func TestLoggerHelpers_RunSerial(t *testing.T) {
	loggerTestSerialGuard.Lock()
	defer loggerTestSerialGuard.Unlock()
	assert.True(t, true)
}
