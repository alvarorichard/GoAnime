package tui

import (
	"archive/zip"
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// windowsTerminalLaunchCase is one row of the decision table for the
// standalone handoff. Keeping the table in a function lets the meta-tests
// inspect exactly the rows the real test runs.
type windowsTerminalLaunchCase struct {
	name        string
	goos        string
	env         []string
	interactive bool
	classic     bool
	sole        bool
	want        bool
}

func windowsTerminalLaunchCases() []windowsTerminalLaunchCase {
	cases := make([]windowsTerminalLaunchCase, 0, 16+4+len(windowsTerminalOptOuts))
	for _, goos := range []string{"windows", "linux"} {
		for _, interactive := range []bool{false, true} {
			for _, classic := range []bool{false, true} {
				for _, sole := range []bool{false, true} {
					cases = append(cases, windowsTerminalLaunchCase{
						name:        "matrix " + goos + " interactive=" + strconv.FormatBool(interactive) + " classic=" + strconv.FormatBool(classic) + " sole=" + strconv.FormatBool(sole),
						goos:        goos,
						interactive: interactive,
						classic:     classic,
						sole:        sole,
						want:        goos == "windows" && interactive && classic && sole,
					})
				}
			}
		}
	}
	cases = append(cases,
		windowsTerminalLaunchCase{
			name:        "Windows Terminal sets WT_SESSION",
			goos:        "windows",
			env:         []string{"WT_SESSION=8f3a"},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		windowsTerminalLaunchCase{
			name:        "the relaunched child must not loop",
			goos:        "windows",
			env:         []string{"GOANIME_WT_RELAUNCH=1"},
			interactive: true,
			classic:     true,
			sole:        true,
		},
		windowsTerminalLaunchCase{
			name:        "GOANIME_TERMINAL=auto keeps the default",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL=auto"},
			interactive: true,
			classic:     true,
			sole:        true,
			want:        true,
		},
		windowsTerminalLaunchCase{
			name:        "GOANIME_TERMINAL is trimmed and case-folded",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL= Off "},
			interactive: true,
			classic:     true,
			sole:        true,
		},
	)
	for _, value := range windowsTerminalOptOuts {
		cases = append(cases, windowsTerminalLaunchCase{
			name:        "GOANIME_TERMINAL=" + value + " opts out",
			goos:        "windows",
			env:         []string{"GOANIME_TERMINAL=" + value},
			interactive: true,
			classic:     true,
			sole:        true,
		})
	}
	return cases
}

func TestWindowsTerminalLaunch(t *testing.T) {
	t.Parallel()

	for _, tt := range windowsTerminalLaunchCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := windowsTerminalLaunch(tt.goos, tt.env, tt.interactive, tt.classic, tt.sole)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWindowsTerminalLaunchTableCoversEveryBooleanCombination is a test about
// the test: the decision table must exercise all 16 combinations of
// goos x interactive x classic x sole, and with an empty environment exactly
// the standalone Windows conhost combination may say yes.
func TestWindowsTerminalLaunchTableCoversEveryBooleanCombination(t *testing.T) {
	t.Parallel()

	covered := make(map[string]bool, 16)
	var defaults []string
	for _, tt := range windowsTerminalLaunchCases() {
		key := tt.goos + "/" + strconv.FormatBool(tt.interactive) + "/" + strconv.FormatBool(tt.classic) + "/" + strconv.FormatBool(tt.sole)
		covered[key] = true
		if len(tt.env) == 0 && tt.want {
			defaults = append(defaults, key)
		}
	}
	require.Len(t, covered, 16)
	for _, goos := range []string{"windows", "linux"} {
		for _, interactive := range []bool{false, true} {
			for _, classic := range []bool{false, true} {
				for _, sole := range []bool{false, true} {
					key := goos + "/" + strconv.FormatBool(interactive) + "/" + strconv.FormatBool(classic) + "/" + strconv.FormatBool(sole)
					assert.True(t, covered[key], "decision table is missing %s", key)
				}
			}
		}
	}
	assert.Equal(t, []string{"windows/true/true/true"}, defaults,
		"only a standalone Windows conhost launch may hand off by default")
}

func TestWindowsTerminalLaunchTableNamesAreUnique(t *testing.T) {
	t.Parallel()

	names := make(map[string]bool)
	for _, tt := range windowsTerminalLaunchCases() {
		require.NotEmpty(t, tt.name)
		require.False(t, names[tt.name], "duplicate decision table name %q", tt.name)
		names[tt.name] = true
	}
}

func TestWindowsTerminalLaunchTableCoversEveryOptOut(t *testing.T) {
	t.Parallel()

	covered := make(map[string]bool)
	for _, tt := range windowsTerminalLaunchCases() {
		for _, item := range tt.env {
			value, ok := strings.CutPrefix(item, "GOANIME_TERMINAL=")
			if ok {
				covered[strings.ToLower(strings.TrimSpace(value))] = true
			}
		}
	}
	for _, optOut := range windowsTerminalOptOuts {
		assert.True(t, covered[optOut], "decision table is missing GOANIME_TERMINAL=%s", optOut)
	}
}

func TestWindowsTerminalOptOutAliases(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	for _, value := range windowsTerminalOptOuts {
		require.NotEmpty(t, value)
		require.Equal(t, strings.ToLower(value), value, "opt-out aliases are stored lowercase")
		require.False(t, seen[value], "duplicate opt-out alias %q", value)
		seen[value] = true

		assert.True(t, windowsTerminalDisabled([]string{"GOANIME_TERMINAL=" + value}), value)
		assert.True(t, windowsTerminalDisabled([]string{"GOANIME_TERMINAL=" + strings.ToUpper(value)}), value)
	}
	assert.False(t, windowsTerminalDisabled(nil))
	assert.False(t, windowsTerminalDisabled([]string{"GOANIME_TERMINAL="}))
	assert.False(t, windowsTerminalDisabled([]string{"GOANIME_TERMINAL=auto"}))
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
		{Name: "Microsoft.WindowsTerminal_1.0_x86.zip"},
	}

	got, err := selectWindowsTerminalAsset(assets, "amd64")
	require.NoError(t, err)
	assert.Equal(t, "Microsoft.WindowsTerminal_1.0_x64.zip", got.Name)

	got, err = selectWindowsTerminalAsset(assets, "arm64")
	require.NoError(t, err)
	assert.Equal(t, "Microsoft.WindowsTerminal_1.0_arm64.zip", got.Name)

	_, err = selectWindowsTerminalAsset(assets, "386")
	require.Error(t, err, "an asset without a download URL is not usable")

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
	require.NoError(t, os.MkdirAll(dest, 0o750))
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

func TestInstallWindowsTerminalArchive(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	previous := filepath.Join(root, "current", "stale.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(previous), 0o750))
	require.NoError(t, os.WriteFile(previous, []byte("stale"), 0o600))

	reader := testZipReader(t, []testZipEntry{
		{name: "terminal-1.0/wt.exe", body: "wt"},
		{name: "terminal-1.0/WindowsTerminal.exe", body: "terminal"},
	})
	wt, err := installWindowsTerminalArchive(reader, root)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "current", "wt.exe"), wt)
	assert.FileExists(t, wt)
	assert.NoFileExists(t, previous, "the previous private copy is replaced")

	got, err := os.ReadFile(wt)
	require.NoError(t, err)
	assert.Equal(t, "wt", string(got))

	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	assert.ElementsMatch(t, []string{"current"}, names, "staging directories are cleaned up")
}

func TestInstallWindowsTerminalArchivePreservesPreviousCopyOnFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	current := filepath.Join(root, "current")
	require.NoError(t, os.MkdirAll(current, 0o750))
	previous := filepath.Join(current, "wt.exe")
	require.NoError(t, os.WriteFile(previous, []byte("old"), 0o600))

	reader := testZipReader(t, []testZipEntry{
		{name: "terminal-1.0/WindowsTerminal.exe", body: "terminal"},
	})
	_, err := installWindowsTerminalArchive(reader, root)
	require.Error(t, err)

	got, readErr := os.ReadFile(previous)
	require.NoError(t, readErr)
	assert.Equal(t, "old", string(got), "a failed extraction must not destroy the working copy")
}

func TestInstallWindowsTerminalArchiveRejectsTraversal(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := filepath.Join(base, "root")
	reader := testZipReader(t, []testZipEntry{
		{name: "../evil.exe", body: "evil"},
	})

	_, err := installWindowsTerminalArchive(reader, root)
	require.Error(t, err)
	assert.NoFileExists(t, filepath.Join(base, "evil.exe"))
}

func TestInstallWindowsTerminalArchiveRejectsSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	reader := testZipReader(t, []testZipEntry{
		{name: "terminal-1.0/wt.exe", body: "wt"},
		{name: "terminal-1.0/link.exe", body: "target", mode: os.ModeSymlink | 0o777},
	})

	_, err := installWindowsTerminalArchive(reader, root)
	require.Error(t, err)
	assert.NoDirExists(t, filepath.Join(root, "current"))
}

// identifiers used by the handoff cascade scenarios.
const (
	testWindowsTerminalAlias   = `C:\Users\ana\AppData\Local\Microsoft\WindowsApps\wt.exe`
	testWindowsTerminalPrivate = `C:\Users\ana\AppData\Local\GoAnime\WindowsTerminal\current\wt.exe`
	testWinget                 = `C:\Users\ana\AppData\Local\Microsoft\WindowsApps\winget.exe`
)

// windowsTerminalHandoffScenario describes one run of the cascade. The event
// list is the exact sequence of side effects, so a reordered fallback fails
// even when the final answer stays the same.
type windowsTerminalHandoffScenario struct {
	name        string
	find        string
	findOK      bool
	elevated    bool
	winget      string
	wingetOK    bool
	installed   bool
	installOK   bool
	await       string
	awaitOK     bool
	portable    string
	portableOK  bool
	start       map[string]bool
	want        bool
	wantEvents  []string
	wantNotices []windowsTerminalNotice
}

func windowsTerminalHandoffScenarios() []windowsTerminalHandoffScenario {
	return []windowsTerminalHandoffScenario{
		{
			name:   "an installed terminal is used as is",
			find:   testWindowsTerminalAlias,
			findOK: true,
			start:  map[string]bool{testWindowsTerminalAlias: true},
			want:   true,
			wantEvents: []string{
				"find",
				"start:" + testWindowsTerminalAlias,
			},
		},
		{
			name:       "an elevated launch skips winget and uses the bundle",
			elevated:   true,
			winget:     testWinget,
			wingetOK:   true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalPrivate: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:     "an elevated launch without a bundle leaves conhost",
			elevated: true,
			winget:   testWinget,
			wingetOK: true,
			want:     false,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"portable",
				"notice:install-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing, terminalNoticeInstallFailed},
		},
		{
			name:       "winget installs and the alias appears",
			winget:     testWinget,
			wingetOK:   true,
			installOK:  true,
			await:      testWindowsTerminalAlias,
			awaitOK:    true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalAlias: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"install:" + testWinget,
				"await:" + testWindowsTerminalAlias,
				"start:" + testWindowsTerminalAlias,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:       "winget installs but the alias never appears, so the bundle is used",
			winget:     testWinget,
			wingetOK:   true,
			installOK:  true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalPrivate: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"install:" + testWinget,
				"await",
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:       "winget fails, so the bundle is used",
			winget:     testWinget,
			wingetOK:   true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalPrivate: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"install:" + testWinget,
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:       "winget is missing, so the bundle is used, as in Windows Sandbox",
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalPrivate: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget",
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:       "the package exists but its alias is disabled",
			winget:     testWinget,
			wingetOK:   true,
			installed:  true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start:      map[string]bool{testWindowsTerminalPrivate: true},
			want:       true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"notice:alias-disabled",
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing, terminalNoticeAliasDisabled},
		},
		{
			name:      "a disabled alias and a failed bundle leaves conhost",
			winget:    testWinget,
			wingetOK:  true,
			installed: true,
			want:      false,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"notice:alias-disabled",
				"portable",
				"notice:install-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing, terminalNoticeAliasDisabled, terminalNoticeInstallFailed},
		},
		{
			name: "no winget and a failed bundle leaves conhost",
			want: false,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget",
				"portable",
				"notice:install-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing, terminalNoticeInstallFailed},
		},
		{
			name:       "a stale alias retries with the private copy",
			find:       testWindowsTerminalAlias,
			findOK:     true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start: map[string]bool{
				testWindowsTerminalAlias:   false,
				testWindowsTerminalPrivate: true,
			},
			want: true,
			wantEvents: []string{
				"find",
				"start:" + testWindowsTerminalAlias,
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
		},
		{
			name:       "the private copy failing is not retried against itself",
			find:       testWindowsTerminalPrivate,
			findOK:     true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			want:       false,
			wantEvents: []string{
				"find",
				"start:" + testWindowsTerminalPrivate,
				"portable:" + testWindowsTerminalPrivate,
				"notice:start-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeStartFailed},
		},
		{
			name:       "both binaries fail to start",
			find:       testWindowsTerminalAlias,
			findOK:     true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			want:       false,
			wantEvents: []string{
				"find",
				"start:" + testWindowsTerminalAlias,
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
				"notice:start-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeStartFailed},
		},
		{
			name:   "a failed start without a private copy leaves conhost",
			find:   testWindowsTerminalAlias,
			findOK: true,
			want:   false,
			wantEvents: []string{
				"find",
				"start:" + testWindowsTerminalAlias,
				"portable",
				"notice:start-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeStartFailed},
		},
		{
			name:       "a failed start after a fresh install retries the private copy",
			winget:     testWinget,
			wingetOK:   true,
			installOK:  true,
			await:      testWindowsTerminalAlias,
			awaitOK:    true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			start: map[string]bool{
				testWindowsTerminalAlias:   false,
				testWindowsTerminalPrivate: true,
			},
			want: true,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"install:" + testWinget,
				"await:" + testWindowsTerminalAlias,
				"start:" + testWindowsTerminalAlias,
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing},
		},
		{
			name:       "a failed start after the bundle is used is not retried against itself",
			winget:     testWinget,
			wingetOK:   true,
			installOK:  true,
			portable:   testWindowsTerminalPrivate,
			portableOK: true,
			want:       false,
			wantEvents: []string{
				"find",
				"elevated",
				"notice:missing",
				"winget:" + testWinget,
				"installed:" + testWinget,
				"install:" + testWinget,
				"await",
				"portable:" + testWindowsTerminalPrivate,
				"start:" + testWindowsTerminalPrivate,
				"portable:" + testWindowsTerminalPrivate,
				"notice:start-failed",
			},
			wantNotices: []windowsTerminalNotice{terminalNoticeMissing, terminalNoticeStartFailed},
		},
	}
}

type handoffRecorder struct {
	events  []string
	notices []windowsTerminalNotice
}

func (r *handoffRecorder) step(name string) {
	r.events = append(r.events, name)
}

func (r *handoffRecorder) result(name, value string, ok bool) {
	if ok {
		r.events = append(r.events, name+":"+value)
		return
	}
	r.events = append(r.events, name)
}

func (r *handoffRecorder) notice(notice windowsTerminalNotice) {
	r.notices = append(r.notices, notice)
	r.events = append(r.events, "notice:"+string(notice))
}

func TestWindowsTerminalHandoffCascade(t *testing.T) {
	t.Parallel()

	for _, tt := range windowsTerminalHandoffScenarios() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := &handoffRecorder{}
			got := windowsTerminalHandoff(windowsTerminalHandoffSteps{
				find: func() (string, bool) {
					recorder.step("find")
					return tt.find, tt.findOK
				},
				elevated: func() bool {
					recorder.step("elevated")
					return tt.elevated
				},
				winget: func() (string, bool) {
					recorder.result("winget", tt.winget, tt.wingetOK)
					return tt.winget, tt.wingetOK
				},
				installed: func(string) bool {
					recorder.result("installed", tt.winget, tt.wingetOK)
					return tt.installed
				},
				install: func(string) bool {
					recorder.result("install", tt.winget, tt.wingetOK)
					return tt.installOK
				},
				await: func() (string, bool) {
					recorder.result("await", tt.await, tt.awaitOK)
					return tt.await, tt.awaitOK
				},
				portable: func() (string, bool) {
					recorder.result("portable", tt.portable, tt.portableOK)
					return tt.portable, tt.portableOK
				},
				start: func(wt string) bool {
					recorder.step("start:" + wt)
					return tt.start[wt]
				},
				notice: recorder.notice,
			})

			assert.Equal(t, tt.want, got, "handoff result")
			assert.Truef(t, slices.Equal(tt.wantEvents, recorder.events),
				"events\n got: %q\nwant: %q", recorder.events, tt.wantEvents)
			assert.Equal(t, tt.wantNotices, recorder.notices)
		})
	}
}

// TestWindowsTerminalHandoffScenariosAreWellFormed is a test about the
// cascade test: unique names, known steps, notices extracted from the event
// trace, and a last event that matches the expected answer.
func TestWindowsTerminalHandoffScenariosAreWellFormed(t *testing.T) {
	t.Parallel()

	steps := []string{"find", "elevated", "winget", "installed", "install", "await", "portable", "start", "notice"}
	names := make(map[string]bool)
	scenarios := windowsTerminalHandoffScenarios()
	require.NotEmpty(t, scenarios)

	for _, scenario := range scenarios {
		require.NotEmpty(t, scenario.name)
		require.False(t, names[scenario.name], "duplicate cascade scenario %q", scenario.name)
		names[scenario.name] = true
		require.NotEmpty(t, scenario.wantEvents, scenario.name)

		var notices []windowsTerminalNotice
		for _, event := range scenario.wantEvents {
			step, value, hasValue := strings.Cut(event, ":")
			require.True(t, slices.Contains(steps, step), "%s: unknown step %q", scenario.name, step)
			if step == "notice" {
				require.True(t, hasValue, "%s: notice without a code", scenario.name)
				notices = append(notices, windowsTerminalNotice(value))
			}
		}
		require.Equal(t, notices, scenario.wantNotices,
			"%s: wantNotices must match the notice events", scenario.name)

		last := scenario.wantEvents[len(scenario.wantEvents)-1]
		if scenario.want {
			require.True(t, strings.HasPrefix(last, "start:"),
				"%s: a successful handoff must end in a start", scenario.name)
		} else {
			require.True(t, strings.HasPrefix(last, "notice:"),
				"%s: a failed handoff must explain itself", scenario.name)
		}
	}
}

func TestWindowsTerminalHandoffScenariosCoverEveryNotice(t *testing.T) {
	t.Parallel()

	constants := terminalNoticeConstants(t)
	require.NotEmpty(t, constants)

	covered := make(map[windowsTerminalNotice]string)
	for _, scenario := range windowsTerminalHandoffScenarios() {
		for _, notice := range scenario.wantNotices {
			covered[notice] = scenario.name
		}
	}
	for _, constant := range constants {
		assert.NotEmpty(t, covered[constant.value],
			"terminal notice %s (%q) has no cascade scenario", constant.name, constant.value)
	}
}

func TestWindowsTerminalHandoffScenariosExerciseEveryStep(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	outcomes := make(map[bool]bool)
	for _, scenario := range windowsTerminalHandoffScenarios() {
		outcomes[scenario.want] = true
		for _, event := range scenario.wantEvents {
			step, _, _ := strings.Cut(event, ":")
			seen[step] = true
		}
	}
	for _, step := range []string{"find", "elevated", "winget", "installed", "install", "await", "portable", "start", "notice"} {
		assert.True(t, seen[step], "no cascade scenario exercises %s", step)
	}
	assert.True(t, outcomes[true], "scenarios must cover a successful handoff")
	assert.True(t, outcomes[false], "scenarios must cover a handoff that stays in conhost")
}

func TestWindowsTerminalNoticeMessagesAreDistinct(t *testing.T) {
	t.Parallel()

	messages := make(map[string]string)
	for _, constant := range terminalNoticeConstants(t) {
		message := windowsTerminalNoticeMessage(constant.value)
		assert.Contains(t, message, "GoAnime:")
		other, duplicate := messages[message]
		assert.False(t, duplicate, "%s and %s share a message", constant.name, other)
		messages[message] = constant.name
	}
}

// TestInstallerLaunchesAsOriginalUser is a meta-test over the Inno Setup
// script. The installer elevates, so a postinstall launch without
// runasoriginaluser would leave GoAnime under the administrator token and
// install the private terminal into the wrong profile.
func TestInstallerLaunchesAsOriginalUser(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../build/install.iss")
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, "PrivilegesRequired=admin")
	assert.Contains(t, text, "ArchitecturesInstallIn64BitMode=x64compatible")

	var runLine string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Filename:") && strings.Contains(trimmed, "{#MyAppExeName}") {
			runLine = trimmed
			break
		}
	}
	require.NotEmpty(t, runLine, "install.iss must launch the app after installation")
	for _, flag := range []string{"nowait", "postinstall", "skipifsilent", "runasoriginaluser"} {
		assert.Contains(t, runLine, flag)
	}
	assert.NotContains(t, runLine, "shellexec",
		"runasoriginaluser cannot be combined with shellexec")
}

type terminalNoticeConstant struct {
	name  string
	value windowsTerminalNotice
}

// terminalNoticeConstants parses the production file so a new terminalNotice*
// constant cannot be added without a cascade scenario covering it.
func terminalNoticeConstants(t *testing.T) []terminalNoticeConstant {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "windowsterminal.go", nil, parser.SkipObjectResolution)
	require.NoError(t, err)

	var constants []terminalNoticeConstant
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != len(value.Values) {
				continue
			}
			typeIdent, ok := value.Type.(*ast.Ident)
			if !ok || typeIdent.Name != "windowsTerminalNotice" {
				continue
			}
			for i, name := range value.Names {
				if !strings.HasPrefix(name.Name, "terminalNotice") {
					continue
				}
				literal, ok := value.Values[i].(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				text, err := strconv.Unquote(literal.Value)
				require.NoError(t, err)
				constants = append(constants, terminalNoticeConstant{
					name:  name.Name,
					value: windowsTerminalNotice(text),
				})
			}
		}
	}
	return constants
}

type testZipEntry struct {
	name string
	body string
	mode os.FileMode
}

func testZipReader(t *testing.T, entries []testZipEntry) *zip.Reader {
	t.Helper()

	data := buildTestZip(t, entries)
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	return reader
}

func buildTestZip(t *testing.T, entries []testZipEntry) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		switch {
		case entry.mode != 0:
			header.SetMode(entry.mode)
		case strings.HasSuffix(entry.name, "/"):
			header.SetMode(os.ModeDir | 0o750)
		default:
			header.SetMode(0o600)
		}
		out, err := writer.CreateHeader(header)
		require.NoError(t, err)
		if !strings.HasSuffix(entry.name, "/") {
			_, err = out.Write([]byte(entry.body))
			require.NoError(t, err)
		}
	}
	require.NoError(t, writer.Close())
	return buffer.Bytes()
}

func writeTestZip(t *testing.T, entries []testZipEntry) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "terminal.zip")
	require.NoError(t, os.WriteFile(path, buildTestZip(t, entries), 0o600))
	return path
}
