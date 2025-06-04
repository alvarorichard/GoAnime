package util

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

var Logger *log.Logger

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

// InitLogger initializes the beautiful charmbracelet logger
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
		Logger.Debug("Debug logging enabled with charmbracelet/log")
	} else {
		Logger.SetLevel(log.InfoLevel)
		Logger.SetColorProfile(termenv.TrueColor)
	}
}

// Debug logs a debug message (only when debug mode is enabled)
func Debug(msg interface{}, keyvals ...interface{}) {
	if IsDebug && Logger != nil {
		Logger.Debug(fmt.Sprintf("%v", msg), keyvals...)
	}
}

// Info logs an info message
func Info(msg interface{}, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Info(fmt.Sprintf("%v", msg), keyvals...)
	}
}

// Warn logs a warning message
func Warn(msg interface{}, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Warn(fmt.Sprintf("%v", msg), keyvals...)
	}
}

// Error logs an error message
func Error(msg interface{}, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Error(fmt.Sprintf("%v", msg), keyvals...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg interface{}, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Fatal(fmt.Sprintf("%v", msg), keyvals...)
	}
}

// Debugf logs a formatted debug message (only when debug mode is enabled)
func Debugf(format string, args ...interface{}) {
	if IsDebug && Logger != nil {
		Logger.Debug(fmt.Sprintf(format, args...))
	}
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Info(fmt.Sprintf(format, args...))
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warn(fmt.Sprintf(format, args...))
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Error(fmt.Sprintf(format, args...))
	}
}
