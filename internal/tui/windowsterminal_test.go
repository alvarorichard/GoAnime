package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWindowsTerminalLaunch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		goos        string
		env         []string
		interactive bool
		classic     bool
		sole        bool
		want        bool
	}{
		{
			name:        "standalone double-click in conhost",
			goos:        "windows",
			interactive: true,
			classic:     true,
			sole:        true,
			want:        true,
		},
		{
			name:        "not windows",
			goos:        "linux",
			interactive: true,
			classic:     true,
			sole:        true,
		},
		{
			name:    "no terminal, such as a pipe or CI",
			goos:    "windows",
			classic: true,
			sole:    true,
		},
		{
			name:        "already under a ConPTY host",
			goos:        "windows",
			interactive: true,
			sole:        true,
		},
		{
			name:        "a shell keeps the console alive",
			goos:        "windows",
			interactive: true,
			classic:     true,
		},
		{
			name:        "Windows Terminal sets WT_SESSION",
			goos:        "windows",
			env:         []string{"WT_SESSION=8f3a"},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		{
			name:        "the relaunched child must not loop",
			goos:        "windows",
			env:         []string{"GOANIME_WT_RELAUNCH=1"},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		{
			name:        "GOANIME_TERMINAL=conhost opts out",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL=conhost"},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		{
			name:        "opt-out is case and space insensitive",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL= Off "},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		{
			name:        "GOANIME_TERMINAL=auto keeps the default",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL=auto"},
			interactive: true,
			classic:     true,
			sole:        true,
			want:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := windowsTerminalLaunch(tt.goos, tt.env, tt.interactive, tt.classic, tt.sole)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWindowsTerminalArgs(t *testing.T) {
	t.Parallel()

	got := windowsTerminalArgs(
		`C:\Program Files\GoAnime\goanime.exe`,
		`C:\Users\ana\go anime; backup`,
		[]string{"--debug", "a;b", ""},
	)
	want := []string{
		"-w", "-1", "new-tab",
		"-d", `C:\Users\ana\go anime\; backup`,
		"--", `C:\Program Files\GoAnime\goanime.exe`,
		"--debug", `a\;b`, "",
	}
	assert.True(t, slices.Equal(want, got), "got %q, want %q", got, want)
}

func TestWindowsTerminalArgsWithoutWorkingDirectory(t *testing.T) {
	t.Parallel()

	got := windowsTerminalArgs(`C:\GoAnime\goanime.exe`, "", nil)
	want := []string{"-w", "-1", "new-tab", "--", `C:\GoAnime\goanime.exe`}
	assert.True(t, slices.Equal(want, got), "got %q, want %q", got, want)
}

func TestEscapeWindowsTerminalArg(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `a\;b`, escapeWindowsTerminalArg("a;b"))
	assert.Equal(t, "plain", escapeWindowsTerminalArg("plain"))
	assert.Equal(t, `a\;b\;c`, escapeWindowsTerminalArg("a;b;c"))
	assert.Equal(t, "", escapeWindowsTerminalArg(""))
}

// TestMainRelaunchesBeforeTouchingTheConsole pins the startup order: the
// Windows Terminal handoff must run before main enables VT or UTF-8 on the
// classic console, or the conhost parent has already started painting a
// degraded screen when the child takes over.
func TestMainRelaunchesBeforeTouchingTheConsole(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../cmd/goanime/main.go", nil, parser.SkipObjectResolution)
	require.NoError(t, err, "cmd/goanime/main.go must parse")

	var relaunchPos, vtPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch {
		case x.Name == "tui" && sel.Sel.Name == "RelaunchInWindowsTerminal":
			relaunchPos = n.Pos()
		case x.Name == "tui" && sel.Sel.Name == "EnableVirtualTerminal":
			vtPos = n.Pos()
		}
		return true
	})
	require.NotZero(t, relaunchPos, "startup must call tui.RelaunchInWindowsTerminal from main.go")
	require.NotZero(t, vtPos, "main.go still enables VT on the classic console")
	assert.True(t, relaunchPos < vtPos,
		"the Windows Terminal handoff runs before the parent console is touched")
}
