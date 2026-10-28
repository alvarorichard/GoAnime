package playback

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Every value the playback loop writes to the log can come from a source: an
// anime's name, the text of an error. A line break in one of them forges a log
// line the program never wrote (CodeQL go/log-injection), so those values go
// through util.LogSafe. This scan fails when a log.Printf in this package
// formats a value with %s or %v without it.
var logVerbRe = regexp.MustCompile(`log\.Printf\(.*%[sv]`)

func TestLogPrintfValuesAreSanitized(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	checked := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(file) // #nosec G304 -- a source file of this package
		require.NoError(t, err)
		for i, line := range strings.Split(string(raw), "\n") {
			if !logVerbRe.MatchString(line) {
				continue
			}
			checked++
			require.Contains(t, line, "util.LogSafe(",
				"%s:%d logs a value that may come from a source without util.LogSafe:\n%s", file, i+1, strings.TrimSpace(line))
		}
	}
	require.Positive(t, checked, "the scan found no log.Printf lines to check; the pattern drifted")
}
