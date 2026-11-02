package logview

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeBuffer is a buffer the follower can write while the test reads it.
type safeBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// logsIn points the command at a fresh logs folder and speeds the poll up.
func logsIn(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prevDir, prevPoll := logDirFn, pollEvery
	logDirFn = func() string { return dir }
	pollEvery = 10 * time.Millisecond
	t.Cleanup(func() { logDirFn, pollEvery = prevDir, prevPoll })
	return dir
}

func writeLog(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
	return p
}

func appendTo(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- a file the test made
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
}

func run(t *testing.T, color bool, args ...string) (string, string, int) {
	t.Helper()
	var out, errw bytes.Buffer
	code := Run(context.Background(), args, &out, &errw, color)
	return out.String(), errw.String(), code
}

const session = "===== GoAnime debug session =====\nSession  abc\n\n--- Environment ---\nSystem\n  OS / arch  linux / amd64\n--- Log ---\n" +
	"19:13:38 Searching for anime query=naruto\n19:13:39 Provider check source=Goyabu result=challenge ms=3\n19:13:43 INFO  Loading episode...\n" +
	"\n===== Session abc ended 19:14:10 · ran 32s · 0 warnings · 0 errors =====\n"

func TestShowsTheNewestLogAsItIsWithoutColour(t *testing.T) {
	dir := logsIn(t)
	writeLog(t, dir, "goanime_2026-11-01_10-00-00.log", "an older session\n")
	writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", session)

	out, _, code := run(t, false)

	assert.Equal(t, 0, code)
	assert.Equal(t, session, out, "without colour the file is printed exactly")
}

// "crash_" sorts before "goanime_", so a crash report is only ever picked by
// mistake when it is all there is.
func TestCrashReportsAreNotSessionLogs(t *testing.T) {
	dir := logsIn(t)
	writeLog(t, dir, "crash_2026-11-02_19-13-38_abcd.log", "a crash report\n")

	out, errw, code := run(t, false)

	assert.Equal(t, 1, code)
	assert.Empty(t, out, "a crash report is not what 'logs' shows")
	assert.Contains(t, errw, "goanime --debug")
}

func TestWithColourOnlyEscapeSequencesAreAdded(t *testing.T) {
	dir := logsIn(t)
	writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", session)

	out, _, code := run(t, true)

	assert.Equal(t, 0, code)
	assert.NotEqual(t, session, out, "it is coloured")
	assert.Contains(t, out, "\x1b[")
	assert.Equal(t, session, ansi.Strip(out), "and nothing else changed")
}

func TestLastNLinesAreShownButTheReportStateIsKept(t *testing.T) {
	dir := logsIn(t)
	writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", session)

	out, _, _ := run(t, false, "-n", "2")
	lines := strings.Split(strings.TrimSuffix(session, "\n"), "\n")
	assert.Equal(t, strings.Join(lines[len(lines)-2:], "\n")+"\n", out)

	out, _, _ = run(t, false, "-n", "0")
	assert.Empty(t, out)
}

func TestAFileCanBeNamedByNameByPathAndBeforeOrAfterTheFlags(t *testing.T) {
	dir := logsIn(t)
	older := writeLog(t, dir, "goanime_2026-11-01_10-00-00.log", "the older one\n")
	writeLog(t, dir, "goanime_2026-11-02_10-00-00.log", "the newer one\n")

	byName, _, code := run(t, false, "goanime_2026-11-01_10-00-00.log")
	assert.Equal(t, 0, code)
	assert.Equal(t, "the older one\n", byName)
	byPath, _, _ := run(t, false, older)
	assert.Equal(t, "the older one\n", byPath)
	flagAfter, _, _ := run(t, false, "goanime_2026-11-01_10-00-00.log", "-n", "1")
	assert.Equal(t, "the older one\n", flagAfter)
}

func TestMissingThingsAreReportedNotHidden(t *testing.T) {
	logsIn(t)

	_, errw, code := run(t, false)
	assert.Equal(t, 1, code)
	assert.Contains(t, errw, "goanime --debug")

	_, errw, code = run(t, false, "nope.log")
	assert.Equal(t, 1, code)
	assert.Contains(t, errw, "no such log: nope.log")

	_, errw, code = run(t, false, "-x")
	assert.Equal(t, 2, code, "a usage error")
	assert.Contains(t, errw, "Usage: goanime logs")

	_, _, code = run(t, false, "-n", "-3")
	assert.Equal(t, 2, code)
}

// follow runs the command in the background until the test cancels it. The
// goroutine is always waited for, also when the test fails, so it never reads
// the package variables the cleanup puts back.
func follow(t *testing.T, args ...string) (*safeBuffer, *safeBuffer, func() int) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	out, errw := &safeBuffer{}, &safeBuffer{}
	var code int
	finished := make(chan struct{})
	go func() {
		code = Run(ctx, args, out, errw, false)
		close(finished)
	}()
	t.Cleanup(func() {
		cancel()
		<-finished
	})
	return out, errw, func() int {
		cancel()
		select {
		case <-finished:
			return code
		case <-time.After(3 * time.Second):
			t.Fatal("Run did not stop after the context was cancelled")
			return -1
		}
	}
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	require.Eventually(t, cond, 3*time.Second, 5*time.Millisecond, what)
}

func TestFollowingShowsWhatIsAppendedAndStopsWhenCancelled(t *testing.T) {
	dir := logsIn(t)
	p := writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", "19:00:00 first\n")

	out, _, stop := follow(t, "-f")
	eventually(t, "the history", func() bool { return strings.Contains(out.String(), "first") })

	appendTo(t, p, "19:00:01 second\n19:00:02 third\n")
	eventually(t, "the appended lines", func() bool { return strings.Contains(out.String(), "third") })

	assert.Equal(t, "19:00:00 first\n19:00:01 second\n19:00:02 third\n", out.String(), "each line once")
	assert.Equal(t, 0, stop())
}

// A line the program is still writing is shown once it is whole, never in two
// pieces.
func TestFollowingWaitsForTheEndOfALine(t *testing.T) {
	dir := logsIn(t)
	p := writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", "19:00:00 first\n")
	out, _, stop := follow(t, "-f")
	eventually(t, "the history", func() bool { return strings.Contains(out.String(), "first") })

	appendTo(t, p, "19:00:01 half of a li")
	time.Sleep(80 * time.Millisecond)
	assert.NotContains(t, out.String(), "half", "an unfinished line is held back")

	appendTo(t, p, "ne\n")
	eventually(t, "the whole line", func() bool { return strings.Contains(out.String(), "19:00:01 half of a line\n") })
	assert.Equal(t, 0, stop())
}

// The usual use: "goanime logs -f" is open, then goanime --debug is started.
func TestFollowingMovesToTheNewSession(t *testing.T) {
	dir := logsIn(t)
	writeLog(t, dir, "goanime_2026-11-02_10-00-00.log", "19:00:00 the previous session\n")
	out, _, stop := follow(t, "-f")
	eventually(t, "the previous session", func() bool { return strings.Contains(out.String(), "previous session") })

	next := writeLog(t, dir, "goanime_2026-11-02_11-00-00.log", "11:00:00 the new session starts\n")
	eventually(t, "the new session", func() bool { return strings.Contains(out.String(), "the new session starts") })
	appendTo(t, next, "11:00:01 and goes on\n")
	eventually(t, "its later lines", func() bool { return strings.Contains(out.String(), "and goes on") })

	got := out.String()
	assert.Contains(t, got, "new session: goanime_2026-11-02_11-00-00.log")
	assert.Less(t, strings.Index(got, "previous session"), strings.Index(got, "new session:"))
	assert.Equal(t, 0, stop())
}

func TestFollowingBeforeTheFirstSessionWaitsForIt(t *testing.T) {
	dir := logsIn(t)
	out, errw, stop := follow(t, "-f")
	eventually(t, "the waiting notice", func() bool { return strings.Contains(errw.String(), "Waiting") })

	writeLog(t, dir, "goanime_2026-11-02_10-00-00.log", "10:00:00 finally\n")
	eventually(t, "the first session", func() bool { return strings.Contains(out.String(), "finally") })
	assert.Equal(t, 0, stop())
}

func TestANamedFileIsFollowedAndNotLeftForANewerOne(t *testing.T) {
	dir := logsIn(t)
	named := writeLog(t, dir, "goanime_2026-11-01_10-00-00.log", "19:00:00 mine\n")
	out, _, stop := follow(t, "-f", "goanime_2026-11-01_10-00-00.log")
	eventually(t, "the history", func() bool { return strings.Contains(out.String(), "mine") })

	writeLog(t, dir, "goanime_2026-11-02_10-00-00.log", "19:00:00 someone else's\n")
	appendTo(t, named, "19:00:01 still mine\n")
	eventually(t, "my later line", func() bool { return strings.Contains(out.String(), "still mine") })

	assert.NotContains(t, out.String(), "someone else's", "an explicit file is not abandoned")
	assert.Equal(t, 0, stop())
}

func TestATruncatedLogIsReadFromTheStartAgain(t *testing.T) {
	dir := logsIn(t)
	p := writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", "19:00:00 a long first line of the old content\n")
	out, _, stop := follow(t, "-f")
	eventually(t, "the history", func() bool { return strings.Contains(out.String(), "old content") })

	require.NoError(t, os.WriteFile(p, []byte("19:05:00 new\n"), 0o600))
	eventually(t, "the new content", func() bool { return strings.Contains(out.String(), "19:05:00 new") })
	assert.Contains(t, out.String(), "log was reset")
	assert.Equal(t, 0, stop())
}

// A large file is read from its end, and following it then continues exactly
// where that ended: no line twice, none lost.
func TestALargeLogIsReadFromItsTailAndFollowedWithoutRepeats(t *testing.T) {
	dir := logsIn(t)
	var b strings.Builder
	for i := range 200000 { // well over maxHistory
		fmt.Fprintf(&b, "19:00:00 filler line number %07d with some padding to make it long\n", i)
	}
	p := writeLog(t, dir, "goanime_2026-11-02_19-13-38.log", b.String())
	require.Greater(t, int64(b.Len()), int64(maxHistory))

	out, _, stop := follow(t, "-f", "-n", "2")
	eventually(t, "the tail", func() bool { return strings.Contains(out.String(), "0199999") })
	appendTo(t, p, "19:00:01 after the tail\n")
	eventually(t, "the appended line", func() bool { return strings.Contains(out.String(), "after the tail") })

	got := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	assert.Len(t, got, 3, "the last two lines, then the new one: %v", got)
	assert.Contains(t, got[0], "0199998")
	assert.Contains(t, got[2], "after the tail")
	assert.Equal(t, 0, stop())
}
