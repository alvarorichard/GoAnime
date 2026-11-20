package logview

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func newColorizer() *Colorizer {
	theme := tui.NewTheme(true)
	return NewColorizer(&theme)
}

// A real session, from the header to the footer: every kind of line the file
// holds.
var sampleSession = []string{
	"===== GoAnime debug session =====",
	"Session  3f549c79",
	"Started  2026-11-02 18:58:09 -03:00",
	"Log      ~/.local/share/goanime/logs/goanime_2026-11-02_18-58-09.log",
	"",
	"This file is safe to attach to a bug report: credentials, one-time tickets and",
	"19:13:38 [PERF] starting Goanime v2.0.0",
	"19:13:38 Searching for anime/media query=the-walking-dead kinds=\"[StartFlix TopCine Pobreflix]\"",
	"",
	"--- Environment ---",
	"Application",
	"  Version     2.0.0",
	"  Executable  ~/.cache/go-build/b4/main",
	"",
	"System",
	"  OS / arch     linux / amd64",
	"  Distribution  Fedora Linux 44 (KDE Plasma Desktop Edition)",
	"--- Log ---",
	"19:13:39 Provider check source=Goyabu result=challenge ms=32 status=403 detail=\"cloudflare challenge (challenge)\"",
	"19:13:39 Provider check source=HiAnime result=ok ms=53 status=200",
	"19:13:43 INFO  Loading episode...",
	"19:13:50 WARN  4 source(s) did not answer; results may be incomplete:",
	"19:13:50 WARN    Goyabu blocked the request (captcha/challenge)",
	"19:13:51 ERROR something broke source=Goyabu err=\"cfsolve: exit status 1\"",
	"19:13:52 Starting mpv with arguments: [--no-terminal --input-ipc-server=/var/tmp/x --alang=por,pob http://127.0.0.1:39987/abyss/a.mp4?token=[redacted]]",
	"19:13:53 something without any fields at all",
	"a line that is not in any known shape",
	"===== Session 49a013c9 ended 19:14:10 · ran 44s · 0 warnings · 0 errors =====",
}

// The promise of the viewer: colour is added, nothing is changed.
func TestColouringNeverChangesAnyCharacter(t *testing.T) {
	t.Parallel()
	c := newColorizer()
	for _, line := range sampleSession {
		got := c.Line(line)
		assert.Equal(t, line, ansi.Strip(got), "the text of %q must survive", line)
	}
}

func TestColouringNeverChangesTextForGeneratedLines(t *testing.T) {
	t.Parallel()
	fragments := []string{"19:13:38", "INFO ", "WARN ", "ERROR", "FATAL", " ", "  ", "key=value", "result=challenge", `detail="a b"`, `x="un"closed`,
		"[PERF]", "[TRACE]", "=", "==", "a=b=c", "http://h/p?q=1&r=2", "(x=1)", "Application", "--- Log ---", "naruto", "İ", "日本語", "\t"}
	rng := rand.New(rand.NewPCG(3, 5))
	c := newColorizer()
	for range 20000 {
		var b strings.Builder
		for range 1 + rng.IntN(9) {
			b.WriteString(fragments[rng.IntN(len(fragments))])
		}
		line := b.String()
		if strings.ContainsAny(line, "\r\n") {
			continue
		}
		if got := ansi.Strip(c.Line(line)); got != line {
			t.Fatalf("colouring changed %q into %q", line, got)
		}
	}
}

func TestEachKindOfLineIsStyledAsItsOwnKind(t *testing.T) {
	t.Parallel()
	c := newColorizer()
	plain := func(s string) bool { return s == ansi.Strip(s) }

	for _, line := range sampleSession {
		if line == "" {
			assert.Equal(t, "", c.Line(line))
			continue
		}
		assert.False(t, plain(c.Line(line)), "%q should carry some colour", line)
	}

	info := c.Line("19:13:43 INFO  Loading episode...")
	warn := c.Line("19:13:50 WARN  careful")
	errl := c.Line("19:13:51 ERROR boom")
	assert.NotEqual(t, info, warn)
	assert.NotEqual(t, warn, errl)
	assert.Contains(t, errl, "48;2;163;0;31", "ERROR is the crimson block of a notice, as on the console")

	ok := c.Line("19:13:39 Provider check source=HiAnime result=ok ms=5")
	bad := c.Line("19:13:39 Provider check source=Goyabu result=challenge ms=5")
	assert.Contains(t, bad, "\x1b[", "a problem is coloured")
	okValue := render(&c.theme.Soft, "ok")
	badValue := render(&c.theme.Primary, "challenge")
	assert.Contains(t, ok, okValue, "a healthy provider stays quiet")
	assert.Contains(t, bad, badValue, "a problem is the accent")
}

// A "=" that does not start a field (inside a URL, say) is text, and a field
// with a quoted value keeps the value whole.
func TestFieldsAreRecognisedOnlyWhereAKeyStartsAWord(t *testing.T) {
	t.Parallel()
	c := newColorizer()
	line := `19:13:39 GET http://h/p?token=abc&q=1 source=Goyabu detail="two words"`

	got := c.Line(line)

	assert.Equal(t, line, ansi.Strip(got))
	assert.Contains(t, got, render(&c.theme.Muted, "source"), "a real field")
	assert.Contains(t, got, render(&c.theme.Soft, `"two words"`), "a quoted value is one value")
	assert.NotContains(t, got, render(&c.theme.Muted, "token"), "the = inside the URL is not a field")
}

func TestTheEnvironmentReportIsReadAsRowsAndTheLogAfterItAsLines(t *testing.T) {
	t.Parallel()
	c := newColorizer()

	title := c.Line("Application") // before the report: a plain line
	_ = c.Line("--- Environment ---")
	section := c.Line("System")
	row := c.Line("  Distribution  Fedora Linux 44")
	_ = c.Line("--- Log ---")
	after := c.Line("19:13:39 Checking 9 providers")

	assert.Equal(t, render(&c.theme.Soft, "Application"), title, "outside the report a bare word is just text")
	assert.Equal(t, render(&c.theme.Primary, "System"), section, "inside it, a section title")
	assert.Contains(t, row, render(&c.theme.Muted, "Distribution"))
	assert.Contains(t, row, render(&c.theme.Text, "Fedora Linux 44"))
	assert.Contains(t, after, render(&c.theme.Muted, "19:13:39"), "the log after the report is read as log lines again")
}

func TestResetForgetsBeingInsideTheReport(t *testing.T) {
	t.Parallel()
	c := newColorizer()
	_ = c.Line("--- Environment ---")
	c.Reset()
	assert.Equal(t, render(&c.theme.Soft, "System"), c.Line("System"))
}
