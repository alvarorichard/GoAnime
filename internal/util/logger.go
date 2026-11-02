package util

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/alvarorichard/Goanime/v2/internal/diag"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

var Logger *log.Logger

var consoleLogMu sync.Mutex
var consoleLogSuppressions int

// fileLogger is a separate logger that writes plain text to the log file (no ANSI codes)
var fileLogger *log.Logger

// logFile holds the reference to the open log file so it can be closed on cleanup
var logFile *logSink

// logFileMu guards logFile: the debug report is written from a goroutine that
// can outlive the log, and CloseLogFile runs on the main one.
var logFileMu sync.RWMutex

func currentLogFile() *logSink {
	logFileMu.RLock()
	defer logFileMu.RUnlock()
	return logFile
}

func setLogFile(s *logSink) {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	logFile = s
}

// Retention: a debug session writes a file each run, so without a limit the
// folder only grows. The newest are kept; crash reports are kept apart so a
// busy week of debugging does not push out the one report that matters.
const (
	maxSessionLogs = 30
	maxCrashLogs   = 10
)

var (
	sessionID    = newSessionID()
	sessionStart = time.Now()
	warnCount    atomic.Int32
	errorCount   atomic.Int32
	// sessionReport is the environment block written at the start of the
	// session, kept so a crash report can carry it too.
	sessionReport atomic.Value // string
)

var (
	debugHooksMu   sync.Mutex
	debugHooks     []func()
	debugHooksOnce sync.Once
)

// OnDebugSession registers f to run once, as soon as a debug session's log
// file is open (never without --debug). The program uses it to record the
// machine and the providers: the util package cannot know about either.
func OnDebugSession(f func()) {
	debugHooksMu.Lock()
	defer debugHooksMu.Unlock()
	debugHooks = append(debugHooks, f)
}

func runDebugHooks() {
	debugHooksOnce.Do(func() {
		debugHooksMu.Lock()
		hooks := append([]func(){}, debugHooks...)
		debugHooksMu.Unlock()
		for _, f := range hooks {
			f()
		}
	})
}

func newSessionID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
	}
	return hex.EncodeToString(b[:])
}

// SessionID names this run. It is in the log header, in the summary and in any
// crash report, so a report and a log from the same run can be matched.
func SessionID() string { return sessionID }

// logSink is the one place the log file is written. Everything that reaches it
// — log lines, the environment report, a subprocess's output — has credentials,
// tickets and the home directory removed first, so a log can be attached to a
// bug report as it is.
type logSink struct {
	mu     sync.Mutex
	f      *os.File
	closed bool

	subMu sync.Mutex
	sub   *lineWriter
}

func (s *logSink) Write(p []byte) (int, error) {
	// Text only: no escape sequence reaches the file, whether the logger put it
	// there or it came in with the data (mpv's colours, a scraped title).
	clean := ansi.Strip(diag.Redact(string(p)))
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, os.ErrClosed
	}
	if _, err := s.f.WriteString(clean); err != nil {
		return 0, err
	}
	return len(p), nil // the caller wrote len(p) bytes; redaction may shorten them
}

// lines is the writer for output that arrives in arbitrary chunks (a
// subprocess's pipe): it hands the sink whole lines only.
func (s *logSink) lines() io.Writer {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	if s.sub == nil {
		s.sub = &lineWriter{sink: s}
	}
	return s.sub
}

func (s *logSink) close() { s.finish("") }

// finish empties the partial subprocess line, writes footer (redacted like
// everything else) and closes the file. Later writes answer os.ErrClosed.
func (s *logSink) finish(footer string) {
	s.subMu.Lock()
	sub := s.sub
	s.subMu.Unlock()
	if sub != nil {
		sub.flush()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	if footer != "" {
		_, _ = s.f.WriteString(diag.Redact(footer))
	}
	s.closed = true
	_ = s.f.Sync()
	_ = s.f.Close()
}

// maxPartialLine bounds what lineWriter holds waiting for a newline.
const maxPartialLine = 8 << 10

// lineWriter buffers input up to each newline before it reaches the sink. A
// secret split across two writes matches no rule in either half, so redaction
// has to see the whole line. A line that never ends is flushed at
// maxPartialLine rather than held without limit.
type lineWriter struct {
	mu   sync.Mutex
	sink *logSink
	buf  []byte
}

func (l *lineWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, p...)
	for {
		i := bytes.IndexByte(l.buf, '\n')
		if i < 0 {
			break
		}
		if _, err := l.sink.Write(l.buf[:i+1]); err != nil {
			return 0, err
		}
		l.buf = append(l.buf[:0], l.buf[i+1:]...)
	}
	if len(l.buf) > maxPartialLine {
		if _, err := l.sink.Write(l.buf); err != nil {
			return 0, err
		}
		l.buf = l.buf[:0]
	}
	return len(p), nil
}

func (l *lineWriter) flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.buf) > 0 {
		_, _ = l.sink.Write(append(l.buf, '\n'))
		l.buf = nil
	}
}

// LogFilePath stores the path to the current debug log file (exported for user visibility)
var LogFilePath string

// PrintSavedLocation prints a colored message showing where a downloaded file
// or directory was saved: the label in the accent color, the path in plain text.
// Falls back to plain text when the console cannot render ANSI safely.
func PrintSavedLocation(label, path string) {
	if !tui.SupportsANSI(os.Stdout) {
		fmt.Printf("%s %s\n", label, path)
		return
	}
	theme := tui.CurrentTheme()
	fmt.Printf("%s %s\n", theme.Primary.Render(label), theme.Text.Render(path))
}

// getColoredPrefix returns a styled logger prefix, or plain text when the
// console cannot render ANSI (classic Windows cmd.exe without VT).
func getColoredPrefix() string {
	return prefixForProfile(tui.ConsoleColorProfile(os.Stderr))
}

// prefixForProfile builds the GoAnime logger prefix for the given profile.
// Exported-to-tests via package scope so regressions cannot reintroduce
// baked-in TrueColor ANSI on ASCII/NoTTY profiles.
func prefixForProfile(p colorprofile.Profile) string {
	if p <= colorprofile.ASCII {
		return "GoAnime"
	}
	theme := tui.CurrentTheme()
	return theme.Header.MarginRight(1).Render("GoAnime")
}

// GetLogDir returns the platform-specific directory for storing log files.
// The paths are chosen to be easily accessible for non-technical users:
//   - Windows: %LOCALAPPDATA%\GoAnime\logs
//   - macOS:   ~/Library/Logs/GoAnime
//   - Linux:   ~/.local/share/goanime/logs
func GetLogDir() string {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "GoAnime", "logs")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", "GoAnime")
	default: // linux and others
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "goanime", "logs")
	}
}

// initFileLogger creates the log file and initializes the file-only logger.
// Each run creates a unique log file (date + time) so logs are never overwritten or mixed.
// Returns the file handle (caller must close) or nil on error.
func initFileLogger() *logSink {
	logDir := GetLogDir()
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log directory %s: %v\n", logDir, err)
		return nil
	}

	// Each session gets a unique file: goanime_2026-02-27_15-44-10.log
	// This ensures multiple runs per day never collide or mix logs
	now := time.Now()
	filename := fmt.Sprintf("goanime_%s.log", now.Format("2006-01-02_15-04-05"))
	logPath := filepath.Join(logDir, filename)

	// In the unlikely event of two runs in the same second, append a counter
	if _, err := os.Stat(logPath); err == nil {
		for i := 2; i <= 100; i++ {
			candidate := filepath.Join(logDir, fmt.Sprintf("goanime_%s_%d.log", now.Format("2006-01-02_15-04-05"), i))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				logPath = candidate
				break
			}
		}
	}

	LogFilePath = logPath

	// Create new file exclusively for this session (O_CREATE|O_WRONLY, no O_APPEND needed since it's a fresh file)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600) // #nosec G304
	if err != nil {
		// Fallback to append mode if O_EXCL fails for any reason
		f, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // #nosec G304
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open log file %s: %v\n", logPath, err)
			return nil
		}
	}

	sink := &logSink{f: f}
	pruneLogs(logDir, "goanime_", maxSessionLogs, logPath)
	pruneLogs(logDir, "crash_", maxCrashLogs, "")

	_, _ = sink.Write([]byte(sessionHeader(now, logPath)))

	// Create a plain-text logger that writes to the file (no ANSI colors)
	fileLogger = log.NewWithOptions(sink, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Prefix:          "GoAnime",
	})
	fileLogger.SetLevel(log.DebugLevel)
	// NoTTY, not ASCII: ASCII drops the colours but keeps bold and faint, which
	// leave "^[[1mDEBUG^[[m" in every line of a file read in an editor or less.
	fileLogger.SetColorProfile(colorprofile.NoTTY)

	return sink
}

// sessionHeader opens the log: who this run is and the promise that the file
// is safe to share.
func sessionHeader(now time.Time, path string) string {
	return fmt.Sprintf(`===== GoAnime debug session =====
Session  %s
Started  %s
Log      %s

This file is safe to attach to a bug report: credentials, one-time tickets and
your home folder are removed before anything is written.
`, sessionID, now.Format("2006-01-02 15:04:05 -07:00"), path)
}

// pruneLogs keeps the newest `keep` files named prefix*.log in dir and removes
// the rest. Names carry their timestamp, so the order of the names is the order
// of the runs. keepFile (the file just created) is never removed.
func pruneLogs(dir, prefix string, keep int, keepFile string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".log") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for len(names) > keep {
		victim := filepath.Join(dir, names[0])
		names = names[1:]
		if victim == keepFile {
			continue
		}
		_ = os.Remove(victim)
	}
}

// WriteSessionReport records the environment report (machine, terminal,
// tools, sources) in the debug log and keeps it for any crash report. It does
// nothing when no log file is open.
func WriteSessionReport(text string) {
	sessionReport.Store(text)
	sink := currentLogFile()
	if sink == nil {
		return
	}
	// The log may close between the check and the write: a closed sink answers
	// with an error, which is dropped.
	_, _ = sink.Write([]byte("\n--- Environment ---\n" + text + "\n--- Log ---\n"))
}

// themedLogStyles gives the console log the look of the screens: the level is
// the site's accent, an inverted badge or the crimson "!" block of a notice,
// instead of the library's own cyan, yellow, pink and violet. Everything is a
// theme style, so a light terminal and NO_COLOR follow the theme as the
// screens do.
func themedLogStyles(theme *tui.Theme) *log.Styles {
	level := func(text string, st lipgloss.Style) lipgloss.Style {
		return st.SetString(text).Padding(0, 1)
	}
	return &log.Styles{
		Timestamp: theme.Muted,
		Caller:    theme.Faint,
		Prefix:    lipgloss.NewStyle(),
		Message:   theme.Text,
		Key:       theme.Muted,
		Value:     theme.Soft,
		Separator: theme.Faint,
		Levels: map[log.Level]lipgloss.Style{
			log.DebugLevel: level("DEBUG", theme.Muted.Bold(true)),
			log.InfoLevel:  level("INFO ", theme.Primary),
			log.WarnLevel:  level("WARN ", theme.Tag.Padding(0, 0)),
			log.ErrorLevel: level("ERROR", theme.Notice.Padding(0, 0)),
			log.FatalLevel: level("FATAL", theme.Notice.Padding(0, 0)),
		},
		Keys:   map[string]lipgloss.Style{},
		Values: map[string]lipgloss.Style{},
	}
}

// newConsoleLogger builds the console logger for w at the given profile.
func newConsoleLogger(w io.Writer, profile colorprofile.Profile, timestamps bool) *log.Logger {
	l := log.NewWithOptions(w, log.Options{
		ReportTimestamp: timestamps,
		TimeFormat:      "15:04:05",
		Prefix:          prefixForProfile(profile),
	})
	theme := tui.CurrentTheme()
	l.SetStyles(themedLogStyles(&theme))
	l.SetLevel(log.InfoLevel)
	l.SetColorProfile(profile)
	return l
}

// InitLogger initializes the beautiful charmbracelet logger.
// When debug mode is enabled, logs are written to a file on disk
// so users can easily share them for troubleshooting.
// The console logger always stays at InfoLevel to avoid corrupting
// interactive TUI components (menus, prompts, etc.).
func InitLogger() {
	// Safe profile: enables VT on Windows when possible; falls back to ASCII
	// when classic cmd.exe cannot process ANSI (avoids raw ←[38;2;...m garbage).
	profile := tui.ConsoleColorProfile(os.Stderr)

	// Console logger is always InfoLevel to keep the terminal clean for TUI.
	// Debug-level messages are routed exclusively to the log file.
	Logger = newConsoleLogger(os.Stderr, profile, IsDebug)

	if IsDebug {
		// Initialize file logging — all debug output goes here
		setLogFile(initFileLogger())
		if currentLogFile() != nil {
			RegisterCleanup(CloseLogFile)
			// Reset first so the banner starts on its own fresh line rather
			// than risking a glue-on with whatever column the shell's
			// prompt left the cursor at when this process started.
			tui.ResetTerminal()
			showDebugBanner()
			runDebugHooks()
		} else {
			Logger.Info("Debug mode enabled (file logging unavailable — logs will appear in console)")
			// Fallback: if we can't write to a file, allow debug on console
			Logger.SetLevel(log.DebugLevel)
			Logger.SetReportCaller(true)
		}
	}
	// The fate of the Japanese glosses is not debug gossip. Whatever startup
	// did about the fonts — installed and machine-verified, could not land,
	// or the glosses hidden — the user reads the line in any mode. The file
	// log keeps a copy when there is one.
	if !tui.Plain() {
		note := tui.JapaneseNote()
		if note == "" && !tui.ShowJapanese() {
			note = "Japanese labels hidden: this terminal cannot draw them (GOANIME_JAPANESE=1 shows them anyway)"
		}
		if note != "" {
			if IsDebug && currentLogFile() != nil {
				Debug(note)
			}
			Info(note)
		}
	}
}

// SuppressConsoleLogging temporarily silences the console logger while keeping
// file logging available through the util logging helpers. This prevents async
// logs from corrupting interactive progress bars.
func SuppressConsoleLogging() func() {
	consoleLogMu.Lock()
	if Logger != nil {
		if consoleLogSuppressions == 0 {
			Logger.SetOutput(io.Discard)
		}
		consoleLogSuppressions++
	}
	consoleLogMu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			consoleLogMu.Lock()
			defer consoleLogMu.Unlock()
			if Logger == nil || consoleLogSuppressions == 0 {
				return
			}
			consoleLogSuppressions--
			if consoleLogSuppressions == 0 {
				Logger.SetOutput(os.Stderr)
			}
		})
	}
}

// showDebugBanner prints a styled notice so the user knows where to find
// the debug log and how to follow it in real-time from another terminal.
// The follow command adapts to the user's OS. Plain text when ANSI unsafe.
func showDebugBanner() {
	// Pick the right "follow file" command for each OS
	var followCmd string
	switch runtime.GOOS {
	case "windows":
		followCmd = fmt.Sprintf("Get-Content -Wait -Tail 50 %q", LogFilePath)
	default: // linux, darwin, etc.
		followCmd = fmt.Sprintf("tail -f %q", LogFilePath)
	}

	if !tui.SupportsANSI(os.Stderr) {
		fmt.Fprintf(os.Stderr, " DEBUG  Debug log → %s\n       %s (run in another terminal to follow live)\n",
			LogFilePath, followCmd)
		return
	}

	if panel := tui.DebugPanel(LogFilePath, followCmd, debugPanelWidth()); panel != "" {
		fmt.Fprintln(os.Stderr, panel)
		return
	}

	theme := tui.CurrentTheme()
	banner := theme.Header.Render(" DEBUG ")
	path := theme.Soft.Render(LogFilePath)
	tailCmd := theme.Primary.Render(followCmd)
	hint := theme.Muted.Render("(run in another terminal to follow live)")

	fmt.Fprintf(os.Stderr, "%s Debug log → %s\n       %s %s\n", banner, path, tailCmd, hint)
}

// debugPanelWidth is the room for the panel: the terminal's width, capped so a
// wide window does not stretch it, with a sensible default when unknown.
func debugPanelWidth() int {
	w, _, err := term.GetSize(os.Stderr.Fd())
	if err != nil || w <= 0 {
		w = 80
	}
	return min(w, 120)
}

// CloseLogFile flushes and closes the debug log file
func CloseLogFile() {
	logFileMu.Lock()
	sink := logFile
	logFile = nil
	logFileMu.Unlock()
	if sink != nil {
		sink.finish(sessionFooter(time.Now()))
	}
}

// sessionFooter closes the log with how the run went, so the first thing a
// reader of the tail learns is whether anything went wrong.
func sessionFooter(now time.Time) string {
	plural := func(n int32, word string) string {
		if n == 1 {
			return fmt.Sprintf("1 %s", word)
		}
		return fmt.Sprintf("%d %ss", n, word)
	}
	return fmt.Sprintf("\n===== Session %s ended %s · ran %s · %s · %s =====\n",
		sessionID, now.Format("15:04:05"), now.Sub(sessionStart).Round(time.Second),
		plural(warnCount.Load(), "warning"), plural(errorCount.Load(), "error"))
}

// writeToFile writes a message to the file logger if active
func writeToFile(level log.Level, msg string, keyvals ...any) {
	if fileLogger == nil {
		return
	}
	fileLogger.Helper()
	switch level {
	case log.WarnLevel:
		warnCount.Add(1)
	case log.ErrorLevel, log.FatalLevel:
		errorCount.Add(1)
	}
	switch level {
	case log.DebugLevel:
		fileLogger.Debug(msg, keyvals...)
	case log.InfoLevel:
		fileLogger.Info(msg, keyvals...)
	case log.WarnLevel:
		fileLogger.Warn(msg, keyvals...)
	case log.ErrorLevel:
		fileLogger.Error(msg, keyvals...)
	case log.FatalLevel:
		fileLogger.Error(msg, keyvals...) // don't call Fatal on file logger to avoid double-exit
	}
}

// GetLogFileWriter returns the log file as an io.Writer for external use (e.g., capturing subprocess output).
// Returns nil if file logging is not active.
func GetLogFileWriter() io.Writer {
	sink := currentLogFile()
	if sink == nil {
		return nil
	}
	return sink.lines() // whole lines through the redacting sink, never the bare file
}

// Debug logs a debug message (only when debug mode is enabled).
// Debug messages are written exclusively to the log file to avoid
// corrupting interactive TUI elements on the terminal.
func Debug(msg any, keyvals ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if IsDebug {
		formatted := fmt.Sprintf("%v", msg)
		if fileLogger != nil {
			writeToFile(log.DebugLevel, formatted, keyvals...)
		} else if Logger != nil {
			// Fallback: no log file, write to console
			Logger.Debug(formatted, keyvals...)
		}
	}
}

// Info logs an info message
func Info(msg any, keyvals ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Info(formatted, keyvals...)
		writeToFile(log.InfoLevel, formatted, keyvals...)
	}
}

// Warn logs a warning message
func Warn(msg any, keyvals ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Warn(formatted, keyvals...)
		writeToFile(log.WarnLevel, formatted, keyvals...)
	}
}

// Error logs an error message
func Error(msg any, keyvals ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Error(formatted, keyvals...)
		writeToFile(log.ErrorLevel, formatted, keyvals...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg any, keyvals ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		writeToFile(log.FatalLevel, formatted, keyvals...)
		CloseLogFile() // ensure file is flushed before exit
		Logger.Fatal(formatted, keyvals...)
	}
}

// Debugf logs a formatted debug message (only when debug mode is enabled).
// Debug messages are written exclusively to the log file.
func Debugf(format string, args ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if IsDebug {
		formatted := fmt.Sprintf(format, args...)
		if fileLogger != nil {
			writeToFile(log.DebugLevel, formatted)
		} else if Logger != nil {
			// Fallback: no log file, write to console
			Logger.Debug(formatted)
		}
	}
}

// Infof logs a formatted info message
func Infof(format string, args ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Info(formatted)
		writeToFile(log.InfoLevel, formatted)
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Warn(formatted)
		writeToFile(log.WarnLevel, formatted)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...any) {
	if fileLogger != nil {
		fileLogger.Helper() // log lines name whoever logged, not this wrapper
	}
	if Logger != nil {
		Logger.Helper()
	}
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Error(formatted)
		writeToFile(log.ErrorLevel, formatted)
	}
}
