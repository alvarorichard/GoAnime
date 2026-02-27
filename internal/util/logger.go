package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

var Logger *log.Logger

// fileLogger is a separate logger that writes plain text to the log file (no ANSI codes)
var fileLogger *log.Logger

// logFile holds the reference to the open log file so it can be closed on cleanup
var logFile *os.File

// LogFilePath stores the path to the current debug log file (exported for user visibility)
var LogFilePath string

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
// Returns the file handle (caller must close) or nil on error.
func initFileLogger() *os.File {
	logDir := GetLogDir()
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log directory %s: %v\n", logDir, err)
		return nil
	}

	// Log file named with date for easy identification
	filename := fmt.Sprintf("goanime_%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, filename)
	LogFilePath = logPath

	// Open in append mode so multiple runs on the same day accumulate
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // #nosec G304
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open log file %s: %v\n", logPath, err)
		return nil
	}

	// Write a session header to the file
	header := fmt.Sprintf("\n===== GoAnime Debug Session — %s =====\n", time.Now().Format("2006-01-02 15:04:05"))
	_, _ = f.WriteString(header)

	// Create a plain-text logger that writes to the file (no ANSI colors)
	fileLogger = log.NewWithOptions(f, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Prefix:          "GoAnime",
	})
	fileLogger.SetLevel(log.DebugLevel)
	fileLogger.SetColorProfile(termenv.Ascii) // no colors in the file

	return f
}

// InitLogger initializes the beautiful charmbracelet logger.
// When debug mode is enabled, logs are also written to a file on disk
// so users can easily share them for troubleshooting.
func InitLogger() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    IsDebug,
		ReportTimestamp: IsDebug,
		TimeFormat:      "15:04:05",
		Prefix:          getColoredPrefix(),
	})

	// Set the appropriate log level based on debug mode
	if IsDebug {
		Logger.SetLevel(log.DebugLevel)
		Logger.SetColorProfile(termenv.TrueColor)

		// Initialize file logging
		logFile = initFileLogger()
		if logFile != nil {
			// Register cleanup to close the file on exit
			RegisterCleanup(CloseLogFile)
			Logger.Debug("Debug logging enabled — logs are being saved to file", "path", LogFilePath)
		} else {
			Logger.Debug("Debug logging enabled (file logging unavailable)")
		}
	} else {
		Logger.SetLevel(log.InfoLevel)
		Logger.SetColorProfile(termenv.TrueColor)
	}
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

// GetLogFileWriter returns the log file as an io.Writer for external use (e.g., capturing subprocess output).
// Returns nil if file logging is not active.
func GetLogFileWriter() io.Writer {
	if logFile == nil {
		return nil
	}
	return logFile
}

// Debug logs a debug message (only when debug mode is enabled)
func Debug(msg any, keyvals ...any) {
	if IsDebug && Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Debug(formatted, keyvals...)
		writeToFile(log.DebugLevel, formatted, keyvals...)
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

// Error logs an error message
func Error(msg any, keyvals ...any) {
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		Logger.Error(formatted, keyvals...)
		writeToFile(log.ErrorLevel, formatted, keyvals...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg any, keyvals ...any) {
	if Logger != nil {
		formatted := fmt.Sprintf("%v", msg)
		writeToFile(log.FatalLevel, formatted, keyvals...)
		CloseLogFile() // ensure file is flushed before exit
		Logger.Fatal(formatted, keyvals...)
	}
}

// Debugf logs a formatted debug message (only when debug mode is enabled)
func Debugf(format string, args ...any) {
	if IsDebug && Logger != nil {
		formatted := fmt.Sprintf(format, args...)
		Logger.Debug(formatted)
		writeToFile(log.DebugLevel, formatted)
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
