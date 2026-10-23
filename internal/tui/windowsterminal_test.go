package tui

import (
	"archive/zip"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func TestWindowsTerminalArchitecture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		goarch  string
		want    string
		wantErr bool
	}{
		{goarch: "amd64", want: "x64"},
		{goarch: "arm64", want: "arm64"},
		{goarch: "386", want: "x86"},
		{goarch: "riscv64", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.goarch, func(t *testing.T) {
			t.Parallel()
			got, err := windowsTerminalArchitecture(tt.goarch)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSelectWindowsTerminalAsset(t *testing.T) {
	t.Parallel()

	assets := []windowsTerminalAsset{
		{Name: "GroupPolicyTemplates_1.0.zip", URL: "https://github.com/gpo.zip"},
		{Name: "Microsoft.WindowsTerminal_1.0_8wekyb3d8bbwe.msixbundle", URL: "https://github.com/msix"},
		{Name: "Microsoft.WindowsTerminal_1.0_8wekyb3d8bbwe.msixbundle_Windows10_PreinstallKit.zip", URL: "https://github.com/kit.zip"},
		{Name: "Microsoft.WindowsTerminal_1.0_x64.zip", URL: "https://github.com/x64.zip"},
		{Name: "Microsoft.WindowsTerminal_1.0_arm64.zip", URL: "https://github.com/arm64.zip"},
	}

	got, err := selectWindowsTerminalAsset(assets, "amd64")
	require.NoError(t, err)
	assert.Equal(t, "Microsoft.WindowsTerminal_1.0_x64.zip", got.Name)

	got, err = selectWindowsTerminalAsset(assets, "arm64")
	require.NoError(t, err)
	assert.Equal(t, "Microsoft.WindowsTerminal_1.0_arm64.zip", got.Name)

	_, err = selectWindowsTerminalAsset(assets, "ppc64")
	require.Error(t, err)
}

func TestSafeArchivePath(t *testing.T) {
	t.Parallel()

	dest := t.TempDir()
	got, err := safeArchivePath(dest, "terminal-1.0/wt.exe")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dest, "terminal-1.0", "wt.exe"), got)

	got, err = safeArchivePath(dest, "./terminal-1.0/wt.exe")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dest, "terminal-1.0", "wt.exe"), got)

	for _, name := range []string{
		"",
		".",
		"..",
		filepath.Join("..", "evil.exe"),
		filepath.Join(t.TempDir(), "evil.exe"),
	} {
		_, err := safeArchivePath(dest, name)
		assert.Error(t, err, "entry %q must be rejected", name)
	}
}

func TestExtractWindowsTerminalZip(t *testing.T) {
	t.Parallel()

	zipPath := writeTestZip(t, []testZipEntry{
		{name: "terminal-1.0/", body: ""},
		{name: "terminal-1.0/wt.exe", body: "wt"},
		{name: "terminal-1.0/WindowsTerminal.exe", body: "terminal"},
		{name: "terminal-1.0/Images/logo.png", body: "png"},
	})
	dest := t.TempDir()

	wtDir, err := extractWindowsTerminalZip(zipPath, dest)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dest, "terminal-1.0"), wtDir)

	for name, want := range map[string]string{
		"wt.exe":              "wt",
		"WindowsTerminal.exe": "terminal",
		"Images/logo.png":     "png",
	} {
		got, err := os.ReadFile(filepath.Join(wtDir, filepath.FromSlash(name)))
		require.NoError(t, err, name)
		assert.Equal(t, want, string(got), name)
	}
}

func TestExtractWindowsTerminalZipRejectsTraversal(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dest := filepath.Join(base, "extract")
	require.NoError(t, os.MkdirAll(dest, 0o755))
	zipPath := writeTestZip(t, []testZipEntry{
		{name: "../evil.exe", body: "evil"},
	})

	_, err := extractWindowsTerminalZip(zipPath, dest)
	require.Error(t, err)
	assert.NoFileExists(t, filepath.Join(base, "evil.exe"))
}

func TestExtractWindowsTerminalZipRequiresWT(t *testing.T) {
	t.Parallel()

	zipPath := writeTestZip(t, []testZipEntry{
		{name: "terminal-1.0/WindowsTerminal.exe", body: "terminal"},
	})

	_, err := extractWindowsTerminalZip(zipPath, t.TempDir())
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "wt.exe"))
}

type testZipEntry struct {
	name string
	body string
}

func writeTestZip(t *testing.T, entries []testZipEntry) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "terminal.zip")
	file, err := os.Create(path)
	require.NoError(t, err)
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if strings.HasSuffix(entry.name, "/") {
			header.SetMode(os.ModeDir | 0o755)
		} else {
			header.SetMode(0o644)
		}
		out, err := writer.CreateHeader(header)
		require.NoError(t, err)
		if !strings.HasSuffix(entry.name, "/") {
			_, err = out.Write([]byte(entry.body))
			require.NoError(t, err)
		}
	}
	require.NoError(t, writer.Close())
	require.NoError(t, file.Close())
	return path
}
