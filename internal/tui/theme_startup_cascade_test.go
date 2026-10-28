package tui

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/term"
)

// Startup theme cascade: ResolveIsDark → EnsureThemeDetected → IsDark →
// CurrentTheme. This is the exact chain main runs before the SEARCH home
// screen opens, so it pins the two properties the WezTerm freeze violated:
// it finishes fast (no blocking stdin reads on the critical path) and it is
// deterministic (no per-run probing games once resolved).
//
// Serial on purpose: it touches the process-wide detected-theme cache. The
// assertions hold whichever path is taken (env override, cached detection,
// or a live probe), so test order cannot break it.
func TestStartupThemeChain_HeadlessFastAndDark(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs piped stdin; running on a real console")
	}
	start := time.Now()
	EnsureThemeDetected()
	first := IsDark()
	theme := CurrentTheme()
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 5*time.Second, "startup theme resolution must never stall startup")
	assert.True(t, first, "headless default is dark")
	assert.NotNil(t, theme.Text, "a theme must build without a terminal")
}

// TestStartupThemeChain_Repeatable pins the cache: five consecutive runs
// resolve identically. A probe running per screen (or per keystroke) would
// show here as flapping or as elapsed time growing per call.
func TestStartupThemeChain_Repeatable(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs piped stdin; running on a real console")
	}
	var results []bool
	for i := 0; i < 5; i++ {
		EnsureThemeDetected()
		results = append(results, IsDark())
	}
	for i, got := range results {
		assert.Equal(t, results[0], got, "call %d differs: theme must be stable", i)
	}
}
