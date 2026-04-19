package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/charmbracelet/colorprofile"
)

var Logger *log.Logger

var consoleLogMu sync.Mutex
var consoleLogSuppressions int

// fileLogger is a separate logger that writes plain text to the log file (no ANSI codes)
var fileLogger *log.Logger

// logFile holds the reference to the open log file so it can be closed on cleanup
var logFile *os.File

// LogFilePath stores the path to the current debug log file (exported for user visibility)
var LogFilePath string

// PrintSavedLocation prints a colored message showing where a downloaded file
// or directory was saved. The label is shown in purple and the path in light purple.
func PrintSavedLocation(label, path string) {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6366F1")).
		Bold(true)
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA"))
	fmt.Printf("%s %s\n", labelStyle.Render(label), pathStyle.Render(path))
}

// getColoredPrefix returns a styled prefix with colors
func getColoredPrefix() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#6366F1")).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)
	return style.Render("GoAnime")
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
func initFileLogger() *os.File {
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

	// Write session header
	header := fmt.Sprintf("===== GoAnime Debug Session — %s =====\n\n", now.Format("2006-01-02 15:04:05"))
	_, _ = f.WriteString(header)

	// Create a plain-text logger that writes to the file (no ANSI colors)
	fileLogger = log.NewWithOptions(f, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Prefix:          "GoAnime",
	})
	fileLogger.SetLevel(log.DebugLevel)
	fileLogger.SetColorProfile(colorprofile.ASCII) // no colors in the file

	return f
}

// InitLogger initializes the beautiful charmbracelet logger.
// When debug mode is enabled, logs are written to a file on disk
// so users can easily share them for troubleshooting.
// The console logger always stays at InfoLevel to avoid corrupting
// interactive TUI components (menus, prompts, etc.).
func InitLogger() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: IsDebug,
		TimeFormat:      "15:04:05",
		Prefix:          getColoredPrefix(),
	})

	// Console logger is always InfoLevel to keep the terminal clean for TUI.
	// Debug-level messages are routed exclusively to the log file.
	Logger.SetLevel(log.InfoLevel)
	Logger.SetColorProfile(colorprofile.TrueColor)

	if IsDebug {
		// Initialize file logging — all debug output goes here
		logFile = initFileLogger()
		if logFile != nil {
			RegisterCleanup(CloseLogFile)
			showDebugBanner()
			tui.ResetTerminal()
		} else {
			Logger.Info("Debug mode enabled (file logging unavailable — logs will appear in console)")
			// Fallback: if we can't write to a file, allow debug on console
			Logger.SetLevel(log.DebugLevel)
			Logger.SetReportCaller(true)
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
// The follow command adapts to the user's OS.
func showDebugBanner() {
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#6366F1")).
		Bold(true).
		Padding(0, 1).
		Render(" DEBUG ")

	path := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA")).
		Italic(true).
		Render(LogFilePath)

	// Pick the right "follow file" command for each OS
	var followCmd string
	switch runtime.GOOS {
	case "windows":
		followCmd = fmt.Sprintf("Get-Content -Wait -Tail 50 \"%s\"", LogFilePath)
	default: // linux, darwin, etc.
		followCmd = fmt.Sprintf("tail -f \"%s\"", LogFilePath)
	}

	tailCmd := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FCD34D")).
		Bold(true).
		Render(followCmd)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Render("(run in another terminal to follow live)")

	fmt.Fprintf(os.Stderr, "%s Debug log → %s\n       %s %s\n", banner, path, tailCmd, hint)
}

// CloseLogFile flushes and closes the debug log file
func CloseLogFile() {
	if logFile != nil {
		_ = logFile.Sync()
		_ = logFile.Close()
		logFile = nil
	}
}

// writeToFile writes a message to the file logger if active
func writeToFile(level log.Level, msg string, keyvals ...any) {
	if fileLogger == nil {
		return
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

// Debug logs a debug message (only when debug mode is enabled).
// Debug messages are written exclusively to the log file to avoid
// corrupting interactive TUI elements on the terminal.
func Debug(msg any, keyvals ...any) {
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
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Info(formatted, keyvals...)
		writeToFile(log.InfoLevel, formatted, keyvals...)
	}
}

// Warn logs a warning message
func Warn(msg any, keyvals ...any) {
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Warn(formatted, keyvals...)
		writeToFile(log.WarnLevel, formatted, keyvals...)
	}
}

// Debugf logs a formatted debug message (only when debug mode is enabled).
// Debug messages are written exclusively to the log file.
func Debugf(format string, args ...any) {
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
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Info(formatted)
		writeToFile(log.InfoLevel, formatted)
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...any) {
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Warn(formatted)
		writeToFile(log.WarnLevel, formatted)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...any) {
	if Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Error(formatted)
		writeToFile(log.ErrorLevel, formatted)
	}
}
