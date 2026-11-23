package tui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests. The works guard (anime_results_works_test.go) is only
// worth having if it fails when the bug comes back: iCarly 2007 and 2021
// folded into one results row. Each mutant below puts back one way it can
// come back. The meta test builds the package with the mutant swapped in (go
// test -overlay, so no file is touched), runs the guard tests, and requires
// one of them to fail. A mutant whose snippet is no longer in the code fails
// here too: a refactor cannot quietly retire a guard, it has to carry the
// mutant along.

type worksEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type worksMutant struct {
	name  string
	edits []worksEdit
	// kills are the guard tests; at least one must fail against the mutant.
	kills []string
}

const resultsFile = "internal/tui/anime_results.go"

var worksMutants = []worksMutant{
	{
		name:  "pre-fix: grouped by title alone",
		edits: []worksEdit{{resultsFile, "\tif !isFilmOrShow(media) {\n\t\treturn true\n\t}\n\tif g.filmKind", "\treturn true\n\tif g.filmKind"}},
		kills: []string{"TestGroupSameTitleRemakesStayApart", "TestGroupSameWorkRules", "TestResultsScreenPicksTheRemake"},
	},
	{
		name:  "different release years merge",
		edits: []worksEdit{{resultsFile, "return g.filmYear == \"\" || year == \"\" || g.filmYear == year", "return true"}},
		kills: []string{"TestGroupSameTitleRemakesStayApart", "TestGroupSameWorkRules", "TestResultsScreenPicksTheRemake"},
	},
	{
		name:  "a movie merges into its series",
		edits: []worksEdit{{resultsFile, "if g.filmKind != \"\" && g.filmKind != media {", "if false {"}},
		kills: []string{"TestGroupSameWorkRules"},
	},
	{
		name:  "different titles merge",
		edits: []worksEdit{{resultsFile, "\tif g.key != key {\n\t\treturn false\n\t}", "\tif false {\n\t\treturn false\n\t}"}},
		kills: []string{"TestGroupSameWorkRules", "TestGroupSameTitleRemakesStayApart"},
	},
	{
		name:  "an unknown year splits the group",
		edits: []worksEdit{{resultsFile, "return g.filmYear == \"\" || year == \"\" || g.filmYear == year", "return g.filmYear == year"}},
		kills: []string{"TestGroupSameWorkRules", "TestGroupYearlessCopyJoinsFirstMatch"},
	},
	{
		name:  "anime split by season year",
		edits: []worksEdit{{resultsFile, "return media == \"Movie\" || media == \"TV\"", "return media == \"Movie\" || media == \"TV\" || media == \"Anime\""}},
		kills: []string{"TestGroupAnimeYearsNeverSplit"},
	},
	{
		name:  "the group never remembers its year",
		edits: []worksEdit{{resultsFile, "\t\tif g.filmYear == \"\" {\n\t\t\tg.filmYear = year\n\t\t}\n", ""}},
		kills: []string{"TestGroupSameTitleRemakesStayApart", "TestResultsScreenPicksTheRemake", "TestGroupSameWorkRules"},
	},
	{
		name:  "the group never remembers its kind",
		edits: []worksEdit{{resultsFile, "\t\tif g.filmKind == \"\" {\n\t\t\tg.filmKind = media\n\t\t}\n", ""}},
		kills: []string{"TestGroupSameWorkRules"},
	},
	{
		name:  "a yearless copy joins the last match, not the first",
		edits: []worksEdit{{resultsFile, "\t\t\t\tpos = i\n\t\t\t\tbreak\n", "\t\t\t\tpos = i\n"}},
		kills: []string{"TestGroupYearlessCopyJoinsFirstMatch"},
	},
}

func TestWorksGuardKillsMutants(t *testing.T) {
	if testing.Short() {
		t.Skip("mutation tests are skipped with -short; run go test ./internal/tui -run KillsMutants")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := worksModuleRoot(t)

	// Control: the guards run and pass on the real code. Without it a test
	// that always fails — or one renamed away, which -run would select
	// nothing of — would "kill" every mutant.
	seen := map[string]bool{}
	var guards []string
	for _, m := range worksMutants {
		for _, name := range m.kills {
			if !seen[name] {
				seen[name] = true
				guards = append(guards, name)
			}
		}
	}
	out, err := runWorksGuards(t, goBin, root, "", guards)
	require.NoError(t, err, "the guard tests must pass on the real code:\n%s", out)
	for _, name := range guards {
		require.Contains(t, out, "--- PASS: "+name+" (", "guard %s did not run and pass", name)
	}

	for _, m := range worksMutants {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			overlay := writeWorksOverlay(t, root, m.edits)
			out, err := runWorksGuards(t, goBin, root, overlay, m.kills)
			require.NotContains(t, out, "[build failed]", "the mutant must build, so the tests — not the compiler — catch it:\n%s", out)
			require.Error(t, err, "mutant survived: the guard tests passed against it\n%s", out)
			killed := false
			for _, name := range m.kills {
				killed = killed || strings.Contains(out, "--- FAIL: "+name+" (")
			}
			assert.True(t, killed, "the run failed, but not in a guard test %v:\n%s", m.kills, out)
		})
	}
}

func writeWorksOverlay(t *testing.T, root string, edits []worksEdit) string {
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

func runWorksGuards(t *testing.T, goBin, root, overlay string, tests []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	args := []string{"test", "-count=1", "-v", "-run", "^(" + strings.Join(tests, "|") + ")$"}
	if overlay != "" {
		args = append(args, "-overlay="+overlay)
	}
	args = append(args, "./internal/tui")
	cmd := exec.CommandContext(ctx, goBin, args...) // #nosec G204 -- the go toolchain, with this test's own arguments
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func worksModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "no go.mod above the test's directory")
		dir = parent
	}
}
