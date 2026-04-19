package util

import (
	"bytes"
	"io"
	"testing"

	"charm.land/log/v2"
)

func TestSuppressConsoleLoggingKeepsFileLoggingAndRestoresNestedState(t *testing.T) {
	previousLogger := Logger
	previousFileLogger := fileLogger
	previousSuppressions := consoleLogSuppressions
	t.Cleanup(func() {
		Logger = previousLogger
		fileLogger = previousFileLogger
		consoleLogSuppressions = previousSuppressions
	})

	var fileBuf bytes.Buffer
	Logger = log.NewWithOptions(io.Discard, log.Options{Prefix: "test"})
	fileLogger = log.NewWithOptions(&fileBuf, log.Options{Prefix: "file-test"})
	consoleLogSuppressions = 0

	restoreOuter := SuppressConsoleLogging()
	restoreInner := SuppressConsoleLogging()
	if consoleLogSuppressions != 2 {
		t.Fatalf("consoleLogSuppressions = %d, want 2", consoleLogSuppressions)
	}

	Warn("AnimeFire source returned 404, retrying fallback source", "episode", 20)
	if !bytes.Contains(fileBuf.Bytes(), []byte("AnimeFire source returned 404")) {
		t.Fatalf("suppressed console logging must still write diagnostics to file logger; got %q", fileBuf.String())
	}

	restoreOuter()
	if consoleLogSuppressions != 1 {
		t.Fatalf("consoleLogSuppressions after outer restore = %d, want 1", consoleLogSuppressions)
	}

	restoreOuter()
	if consoleLogSuppressions != 1 {
		t.Fatalf("restore must be idempotent; consoleLogSuppressions = %d, want 1", consoleLogSuppressions)
	}

	restoreInner()
	if consoleLogSuppressions != 0 {
		t.Fatalf("consoleLogSuppressions after inner restore = %d, want 0", consoleLogSuppressions)
	}
}
