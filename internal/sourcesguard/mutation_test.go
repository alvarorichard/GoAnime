package sourcesguard

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mutation test of the SOURCES feature. Each mutant puts back one way the
// feature can break; the test builds the code with the mutant swapped in
// (go test -overlay, so no file is touched), runs the named guard tests and
// requires one of them to fail. A mutant that survives means a guard cannot
// see that breakage; a snippet that is no longer in the code fails the test,
// so a refactor must carry its mutant along instead of retiring the guard.

type codeEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type mutant struct {
	name  string
	edits []codeEdit
	// kills maps a package to the guard tests in it; in every package at least
	// one of them must fail against the mutant.
	kills map[string][]string
	// posixOnly marks a mutant whose guard cannot exist on Windows (file
	// permission bits): it is skipped there, loudly, never silently passed.
	posixOnly bool
	// stopsAt, for a cascade, is the stage the run must stop at: later stages
	// must not run.
	stopsAt string
}

const (
	utilPkg      = "./internal/util"
	sourcePkg    = "./internal/api/source"
	providersPkg = "./internal/api/providers"
	tuiPkg       = "./internal/tui"
	handlersPkg  = "./internal/handlers"

	prefsFile   = "internal/util/source_prefs.go"
	enableFile  = "internal/api/source/enablement.go"
	statesFile  = "internal/api/source/states.go"
	liveFile    = "internal/api/providers/live_search.go"
	screenFile  = "internal/tui/sources_screen.go"
	homeFile    = "internal/tui/search_home.go"
	handlerFile = "internal/handlers/searchhome.go"
)

var sourcesMutants = []mutant{
	// --- saved choices (util) ---
	{
		name:  "lookups stop being case-forgiving",
		edits: []codeEdit{{prefsFile, "func SourcePref(name string) (on, set bool) {\n\ttoken := canonSourceToken(name)", "func SourcePref(name string) (on, set bool) {\n\ttoken := name"}},
		kills: map[string][]string{utilPkg: {"TestSourcePrefSetClearReset"}},
	},
	{
		name: "a failed save leaves the choice in memory",
		edits: []codeEdit{{prefsFile,
			"\t\tif hadPrev {\n\t\t\tSetSourcePref(name, prevOn)\n\t\t} else {\n\t\t\tClearSourcePref(name)\n\t\t}\n",
			"\t\t_, _ = prevOn, hadPrev\n"}},
		kills: map[string][]string{utilPkg: {"TestPersistSourcePrefRollsBackOnFailure"}},
	},
	{
		name:  "a file from a newer GoAnime is read anyway",
		edits: []codeEdit{{prefsFile, "if doc.Version > sourcePrefsVersion {", "if false && doc.Version > sourcePrefsVersion {"}},
		kills: map[string][]string{utilPkg: {"TestLoadSourcePrefsRejectsBadFilesWithoutTouchingChoices"}},
	},
	{
		name:  "a missing file is an error at first launch",
		edits: []codeEdit{{prefsFile, "if errors.Is(err, fs.ErrNotExist) {", "if false && errors.Is(err, fs.ErrNotExist) {"}},
		kills: map[string][]string{utilPkg: {"TestLoadSourcePrefsMissingFileIsFine"}},
	},
	{
		name:      "the file is world readable",
		edits:     []codeEdit{{prefsFile, "os.Chmod(tmp.Name(), 0o600)", "os.Chmod(tmp.Name(), 0o644)"}},
		kills:     map[string][]string{utilPkg: {"TestSaveLoadRoundTrip"}},
		posixOnly: true,
	},
	{
		name:  "an empty name reaches the file",
		edits: []codeEdit{{prefsFile, "\t\tif token := canonSourceToken(name); token != \"\" {\n\t\t\tnext[token] = on\n\t\t}", "\t\tnext[canonSourceToken(name)] = on"}},
		kills: map[string][]string{utilPkg: {"TestLoadSourcePrefsDropsEmptyNamesAndNormalises"}},
	},
	{
		name:  "the temp file is left behind when the rename fails",
		edits: []codeEdit{{prefsFile, "\tif err := os.Rename(tmp.Name(), path); err != nil {\n\t\t_ = os.Remove(tmp.Name())\n\t\treturn err", "\tif err := os.Rename(tmp.Name(), path); err != nil {\n\t\treturn err"}},
		kills: map[string][]string{utilPkg: {"TestWriteFileAtomic"}},
	},

	// --- who searches (source) ---
	{
		name:  "the saved choice is ignored by the search",
		edits: []codeEdit{{enableFile, "\tif on, set := util.SourcePref(name); set {\n\t\treturn on\n\t}\n", ""}},
		kills: map[string][]string{sourcePkg: {"TestIsSearchEnabledSavedChoice", "TestSourceStatesFollowsRegistryAndPriority"}},
	},
	{
		name:  "the env opt-in no longer outranks a saved off",
		edits: []codeEdit{{enableFile, "\tif util.SourceForceEnabled(name) {\n\t\treturn true\n\t}\n", ""}},
		kills: map[string][]string{sourcePkg: {"TestIsSearchEnabledSavedChoice", "TestIsSearchEnabled"}},
	},
	{
		name:  "a saved off becomes the hard kill-switch (history stops opening)",
		edits: []codeEdit{{enableFile, "\treturn !util.SourceDisabled(string(d.Kind))\n}", "\tif on, set := util.SourcePref(string(d.Kind)); set && !on {\n\t\treturn false\n\t}\n\treturn !util.SourceDisabled(string(d.Kind))\n}"}},
		kills: map[string][]string{
			sourcePkg:   {"TestIsSearchEnabledSavedChoice"},
			handlersPkg: {"TestSourcesChoiceCascade"},
		},
	},
	{
		name:  "the env kill-switch stops outranking a saved on",
		edits: []codeEdit{{enableFile, "\tif !IsEnabled(d) {\n\t\treturn false\n\t}\n\tname := string(d.Kind)", "\tname := string(d.Kind)"}},
		kills: map[string][]string{sourcePkg: {"TestIsSearchEnabledSavedChoice", "TestIsSearchEnabled"}},
	},
	{
		name:  "the screen cannot tell a locked source",
		edits: []codeEdit{{statesFile, "\t\tst.Locked = true\n", "\t\tst.Locked = false\n"}},
		kills: map[string][]string{sourcePkg: {"TestStateOf"}, handlersPkg: {"TestSourcesChoiceCascade"}},
	},
	{
		name:  "the screen hides a source the environment switched off",
		edits: []codeEdit{{statesFile, "srcs := allByPriority()", "srcs := registeredByPriority()"}},
		kills: map[string][]string{sourcePkg: {"TestSourceStatesKeepsLockedSourcesVisible"}, handlersPkg: {"TestSourcesChoiceCascade"}},
	},
	{
		name:  "the screen's state drifts from the search's",
		edits: []codeEdit{{statesFile, "st := State{Kind: d.Kind, On: IsSearchEnabled(d)}", "st := State{Kind: d.Kind, On: !d.DefaultDisabled}"}},
		kills: map[string][]string{sourcePkg: {"TestStateOf", "TestSourceStatesFollowsRegistryAndPriority"}},
	},
	{
		name:  "a source's own caveat is dropped",
		edits: []codeEdit{{statesFile, "\tif note == \"\" {\n\t\treturn reason\n\t}\n\treturn reason + \"; \" + note", "\treturn reason"}},
		kills: map[string][]string{sourcePkg: {"TestJoinNote", "TestStateOf"}},
	},

	// --- the SEARCHING panel (providers) ---
	{
		name:  "the user's own choice is blamed on the default",
		edits: []codeEdit{{liveFile, "\t\treturn \"turned off\"", "\t\treturn \"off by default\""}},
		kills: map[string][]string{providersPkg: {"TestSearchOffReason", "TestPlanLiveSearchHonoursSavedChoices"}},
	},
	{
		name:  "the search plan ignores the saved choice",
		edits: []codeEdit{{liveFile, "\t\tcase !source.IsSearchEnabled(desc):", "\t\tcase desc.DefaultDisabled && !util.SourceForceEnabled(string(kind)):"}},
		kills: map[string][]string{providersPkg: {"TestPlanLiveSearchHonoursSavedChoices"}},
	},

	// --- the screen (tui) ---
	{
		name:  "a locked source can be flipped",
		edits: []codeEdit{{screenFile, "\tif row.Locked {\n\t\tp.status", "\tif false && row.Locked {\n\t\tp.status"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelToggleOnALockedRowSavesNothing"}},
	},
	{
		name:  "a failed save looks like a success",
		edits: []codeEdit{{screenFile, "err := p.set(row.ID, !row.On); err != nil {", "err := p.set(row.ID, !row.On); false && err != nil {"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelToggleSaveFailureIsShownAndKeepsTheRow"}},
	},
	{
		name:  "esc no longer leaves the screen",
		edits: []codeEdit{{screenFile, "\tcase \"esc\", \"ctrl+s\", \"q\":", "\tcase \"ctrl+s\", \"q\":"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelKeys", "TestSearchHomeCtrlSOpensAndEscCloses"}},
	},
	{
		name:  "space stops flipping",
		edits: []codeEdit{{screenFile, "\tcase \"space\", \" \", \"enter\", \"x\":", "\tcase \"enter\", \"x\":"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelKeys"}},
	},
	{
		name:  "the user is not warned when nothing is on",
		edits: []codeEdit{{screenFile, "if len(p.rows) > 0 && p.onCount() == 0 {", "if false {"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelView"}},
	},
	{
		name:  "the checkbox no longer says what is on",
		edits: []codeEdit{{screenFile, "\tbox := \"[ ]\"\n\tif r.On {\n\t\tbox = \"[x]\"\n\t}", "\tbox := \"[ ]\""}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelView", "TestSourcesPanelRow"}},
	},
	{
		name:  "the header says \"1 sources\"",
		edits: []codeEdit{{screenFile, "\tif on == 1 {\n\t\tnoun = \"source\"\n\t}", ""}},
		kills: map[string][]string{tuiPkg: {"TestSourcesSummary"}},
	},
	{
		name:  "the cursor runs off the list",
		edits: []codeEdit{{screenFile, "\tp.cursor = max(min(p.cursor+delta, len(p.rows)-1), 0)", "\tp.cursor += delta"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelMoveStopsAtTheEnds"}},
	},
	{
		name:  "the instruction line overflows a narrow terminal",
		edits: []codeEdit{{screenFile, "theme.Muted.Render(ansi.Truncate(\"Choose where GoAnime searches. Changes are saved as you make them.\", width, \"…\")),", "theme.Muted.Render(\"Choose where GoAnime searches. Changes are saved as you make them.\"),"}},
		kills: map[string][]string{tuiPkg: {"TestSourcesPanelViewFitsNarrowTerminals", "TestSearchHomeSourcesSurviveResize"}},
	},
	{
		name:  "ctrl+s does nothing on the home screen",
		edits: []codeEdit{{homeFile, "\t\tcase \"ctrl+s\":\n\t\t\tm.openSources()", "\t\tcase \"ctrl+alt+s\":\n\t\t\tm.openSources()"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeCtrlSOpensAndEscCloses"}},
	},
	{
		name:  "the keys leak into the search box under the screen",
		edits: []codeEdit{{homeFile, "\t\tif m.sources != nil {\n\t\t\treturn m.updateSources(msg)\n\t\t}\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSourcesScreenOwnsTheKeys", "TestSearchHomeCtrlSOpensAndEscCloses"}},
	},
	{
		name:  "the header keeps the old count after a flip",
		edits: []codeEdit{{homeFile, "\tm.shell.Context = SourcesSummary(m.sources.rows)\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeToggleUpdatesTheHeaderAtOnce", "TestUpdateSourcesRoutesKeysToThePanel"}},
	},
	{
		name:  "an open dropdown stays under the screen",
		edits: []codeEdit{{homeFile, "\tm.dropSuggestions()\n\tm.sources = newSourcesPanel", "\tm.sources = newSourcesPanel"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeOpeningSourcesDropsOpenSuggestions"}},
	},
	{
		name:  "ctrl+c is swallowed by the screen",
		edits: []codeEdit{{homeFile, "\tif msg.String() == \"ctrl+c\" {\n\t\tm.err = ErrSearchHomeCancelled\n\t\treturn m, tea.Quit\n\t}\n\tclosed := m.sources.update(msg)", "\tclosed := m.sources.update(msg)"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSourcesCtrlCStillQuits", "TestUpdateSourcesRoutesKeysToThePanel"}},
	},
	{
		name:  "the footer stops advertising ctrl+s",
		edits: []codeEdit{{homeFile, "\tif m.loadSources != nil {\n\t\thints = append(", "\tif false && m.loadSources != nil {\n\t\thints = append("}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSourcesAreAdvertised"}},
	},

	// --- the host (handlers) ---
	{
		name:  "a failed save is swallowed",
		edits: []codeEdit{{handlerFile, "\treturn persistSourcePref(path, id, on)", "\t_ = persistSourcePref(path, id, on)\n\treturn nil"}},
		kills: map[string][]string{handlersPkg: {"TestSetHomeSource"}},
	},
	{
		name:  "the screen is not handed its sources",
		edits: []codeEdit{{handlerFile, "\t\tSources:       homeSources,\n", ""}},
		kills: map[string][]string{handlersPkg: {"TestOpenSearchHomeHandsTheScreenItsSources"}},
	},
	{
		name:  "the header hides a source that ships off",
		edits: []codeEdit{{handlerFile, "\treturn tui.SourcesSummary(homeSources())", "\treturn tui.SourcesSummary(homeSourceRows(nil))"}},
		kills: map[string][]string{handlersPkg: {"TestSourcesLineCountsWhatSearches"}},
	},

	// --- the cascade stops at the broken hop ---
	{
		name:    "cascade stops at the unsaved choice",
		edits:   []codeEdit{{prefsFile, "\tif err := SaveSourcePrefs(path); err != nil {\n\t\tif hadPrev", "\tif err := error(nil); err != nil {\n\t\tif hadPrev"}},
		kills:   map[string][]string{handlersPkg: {"TestSourcesChoiceCascade"}},
		stopsAt: "3 the choice is on disk",
	},
}

func TestSourcesGuardKillsMutants(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := moduleRoot(t)

	// Control: the guard tests pass on the real code. Without it a test that
	// always fails, or one renamed away (-run would select nothing), would
	// "kill" every mutant.
	guards := map[string]map[string]bool{}
	for _, m := range sourcesMutants {
		for pkg, tests := range m.kills {
			if guards[pkg] == nil {
				guards[pkg] = map[string]bool{}
			}
			for _, name := range tests {
				guards[pkg][name] = true
			}
		}
	}
	for pkg, set := range guards {
		var tests []string
		for name := range set {
			tests = append(tests, name)
		}
		out, err := runGuardTests(t, goBin, root, "", pkg, tests)
		require.NoError(t, err, "the guard tests must pass on the real code:\n%s", out)
		for _, name := range tests {
			require.Contains(t, out, "--- PASS: "+name+" (", "guard %s did not run and pass in %s", name, pkg)
		}
	}

	for _, m := range sourcesMutants {
		t.Run(m.name, func(t *testing.T) {
			if m.posixOnly && runtime.GOOS == "windows" {
				t.Skip("file permission bits do not exist on Windows; this guard is POSIX-only")
			}
			t.Parallel()
			overlay := writeMutantOverlay(t, root, m.edits)
			for pkg, tests := range m.kills {
				out, err := runGuardTests(t, goBin, root, overlay, pkg, tests)
				require.NotContains(t, out, "[build failed]", "the mutant must build, so the tests — not the compiler — catch it:\n%s", out)
				require.Error(t, err, "mutant survived %s: the guard tests passed against it\n%s", pkg, out)
				killed := false
				for _, name := range tests {
					killed = killed || strings.Contains(out, "--- FAIL: "+name+" (")
				}
				assert.True(t, killed, "%s failed, but not in a guard test %v:\n%s", pkg, tests, out)
				if m.stopsAt != "" {
					assert.Contains(t, out, `cascade stopped: stage "`+m.stopsAt+`" failed`)
					broken, err := strconv.Atoi(strings.Fields(m.stopsAt)[0])
					require.NoError(t, err, "a cascade stage name starts with its number")
					for _, match := range cascadeStageRe.FindAllStringSubmatch(out, -1) {
						ran, _ := strconv.Atoi(match[1])
						assert.LessOrEqual(t, ran, broken, "stage %d ran after the broken stage %d", ran, broken)
					}
				}
			}
		})
	}
}

// Every mutant is a real, distinct claim: named, with at least one guard in at
// least one package, and no two swap the same snippet the same way.
// Only the permission mutant may be POSIX-only: the exemption must not grow
// into a way to skip guards.
func TestOnlyPermissionMutantsArePosixOnly(t *testing.T) {
	var exempt []string
	for _, m := range sourcesMutants {
		if m.posixOnly {
			exempt = append(exempt, m.name)
		}
	}
	assert.Equal(t, []string{"the file is world readable"}, exempt)
}

func TestMutantTableIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range sourcesMutants {
		assert.NotEmpty(t, m.name)
		assert.NotEmpty(t, m.edits, m.name)
		assert.NotEmpty(t, m.kills, m.name)
		assert.False(t, seen[m.name], "duplicate mutant %q", m.name)
		seen[m.name] = true
		for _, e := range m.edits {
			assert.NotEqual(t, e.from, e.to, "%s: a mutant that changes nothing kills nothing", m.name)
			assert.NotEmpty(t, e.file, m.name)
		}
		for pkg, tests := range m.kills {
			assert.NotEmpty(t, tests, "%s: package %s names no guard", m.name, pkg)
		}
	}
	assert.GreaterOrEqual(t, len(sourcesMutants), 30, "the table shrank: a guard was retired")
}

// The overlay machinery itself: it swaps exactly one snippet, refuses a snippet
// that drifted, and keeps CRLF checkouts working.
func TestWriteMutantOverlay(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(body), 0o600))
	}

	t.Run("swaps one snippet and maps the original path", func(t *testing.T) {
		write("a.go", "package a\nvar x = 1\n")
		overlay := writeMutantOverlay(t, root, []codeEdit{{"a.go", "var x = 1", "var x = 2"}})
		raw, err := os.ReadFile(overlay)
		require.NoError(t, err)
		var doc struct{ Replace map[string]string }
		require.NoError(t, jsonx.Unmarshal(raw, &doc))
		require.Len(t, doc.Replace, 1)
		mutated, err := os.ReadFile(doc.Replace[filepath.Join(root, "a.go")])
		require.NoError(t, err)
		assert.Equal(t, "package a\nvar x = 2\n", string(mutated))
		orig, err := os.ReadFile(filepath.Join(root, "a.go"))
		require.NoError(t, err)
		assert.Equal(t, "package a\nvar x = 1\n", string(orig), "the source file is never touched")
	})

	t.Run("several edits to one file compose", func(t *testing.T) {
		write("b.go", "one\ntwo\n")
		overlay := writeMutantOverlay(t, root, []codeEdit{{"b.go", "one", "1"}, {"b.go", "two", "2"}})
		raw, err := os.ReadFile(overlay)
		require.NoError(t, err)
		var doc struct{ Replace map[string]string }
		require.NoError(t, jsonx.Unmarshal(raw, &doc))
		mutated, err := os.ReadFile(doc.Replace[filepath.Join(root, "b.go")])
		require.NoError(t, err)
		assert.Equal(t, "1\n2\n", string(mutated))
	})

	t.Run("a CRLF checkout still matches", func(t *testing.T) {
		write("c.go", "a\r\nb\r\n")
		overlay := writeMutantOverlay(t, root, []codeEdit{{"c.go", "a\nb", "a\nB"}})
		raw, err := os.ReadFile(overlay)
		require.NoError(t, err)
		var doc struct{ Replace map[string]string }
		require.NoError(t, jsonx.Unmarshal(raw, &doc))
		mutated, err := os.ReadFile(doc.Replace[filepath.Join(root, "c.go")])
		require.NoError(t, err)
		assert.Equal(t, "a\r\nB\r\n", string(mutated))
	})
}

// cascadeStageRe reads a cascade stage's number from go test -v output.
var cascadeStageRe = regexp.MustCompile(`=== RUN\s+\S+/(\d+)_`)

// writeMutantOverlay writes the mutated files and the -overlay file that
// swaps them in for the build. A snippet that is not in the file exactly once
// fails the calling test.
func writeMutantOverlay(t *testing.T, root string, edits []codeEdit) string {
	t.Helper()
	dir := t.TempDir()
	mutated := map[string]string{}
	for _, e := range edits {
		path := filepath.Join(root, filepath.FromSlash(e.file))
		src, ok := mutated[path]
		if !ok {
			raw, err := os.ReadFile(path) // #nosec G304 -- a module file named in this test
			require.NoError(t, err)
			src = string(raw)
		}
		from, to := e.from, e.to
		if strings.Contains(src, "\r\n") { // a Windows checkout
			from, to = strings.ReplaceAll(from, "\n", "\r\n"), strings.ReplaceAll(to, "\n", "\r\n")
		}
		require.Equal(t, 1, strings.Count(src, from),
			"%s no longer holds the snippet this mutant breaks; move the mutant with the code:\n%s", e.file, e.from)
		mutated[path] = strings.Replace(src, from, to, 1)
	}
	replace := map[string]string{}
	for path, src := range mutated {
		out := filepath.Join(dir, filepath.Base(path))
		require.NoError(t, os.WriteFile(out, []byte(src), 0o600))
		replace[path] = out
	}
	raw, err := jsonx.Marshal(map[string]any{"Replace": replace})
	require.NoError(t, err)
	overlay := filepath.Join(dir, "overlay.json")
	require.NoError(t, os.WriteFile(overlay, raw, 0o600))
	return overlay
}

// runGuardTests runs the named tests of one package, with the overlay when one
// is given, and returns the verbose output.
func runGuardTests(t *testing.T, goBin, root, overlay, pkg string, tests []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Minute)
	defer cancel()
	args := []string{"test", "-count=1", "-v", "-run", "^(" + strings.Join(tests, "|") + ")$"}
	if overlay != "" {
		args = append(args, "-overlay="+overlay)
	}
	args = append(args, pkg)
	cmd := exec.CommandContext(ctx, goBin, args...) // #nosec G204 -- the go toolchain, with this test's own arguments
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}
