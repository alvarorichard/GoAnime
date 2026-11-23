package startflix

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// Tests of the tests. The year guard (years_test.go) is only worth having if
// it fails when the bug comes back: iCarly 2007 and 2021 reaching the results
// screen with no year, and folding into one row. Each mutant below puts back
// one way it can come back. The meta test builds the package with the mutant
// swapped in (go test -overlay, so no file is touched), runs the guard tests,
// and requires one of them to fail. A mutant whose snippet is no longer in the
// code fails here too: a refactor cannot quietly retire a guard.

type yearEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type yearMutant struct {
	name  string
	edits []yearEdit
	kills []string // at least one must fail against the mutant
}

const (
	sfxYearsFile  = "internal/scraper/providers/startflix/years.go"
	sfxClientFile = "internal/scraper/providers/startflix/client.go"
)

var yearMutants = []yearMutant{
	{
		name:  "pre-fix: search never reads the title pages",
		edits: []yearEdit{{sfxClientFile, "\tc.fillAmbiguousYears(ctx, results)\n", ""}},
		kills: []string{"TestSearchFillsTheYearsOfSameTitleResults"},
	},
	{
		name:  "the year is read from any date on the page",
		edits: []yearEdit{{sfxYearsFile, `doc.Find(".extra .date")`, `doc.Find(".date")`}},
		kills: []string{"TestParseTitleYear"},
	},
	{
		name:  "every result fetches its page, twin or not",
		edits: []yearEdit{{sfxYearsFile, "[strings.ToLower(strings.TrimSpace(m.Title))] > 1", "[strings.ToLower(strings.TrimSpace(m.Title))] > 0"}},
		kills: []string{"TestAmbiguousYearless", "TestSearchLeavesUniqueTitlesAlone"},
	},
	{
		name:  "a result that has a year is fetched anyway",
		edits: []yearEdit{{sfxYearsFile, `if m.Year == "" && count[`, `if count[`}},
		kills: []string{"TestAmbiguousYearless"},
	},
	{
		name:  "titles compared with their case and spacing",
		edits: []yearEdit{{sfxYearsFile, "\t\tcount[strings.ToLower(strings.TrimSpace(m.Title))]++", "\t\tcount[m.Title]++"}},
		kills: []string{"TestAmbiguousYearless", "TestSearchFillsTheYearsOfSameTitleResults"},
	},
	{
		name:  "a broad query fetches every twin's page",
		edits: []yearEdit{{sfxYearsFile, "\tif len(idx) > maxYearLookups {", "\tif false {"}},
		kills: []string{"TestSearchCapsTheYearLookups"},
	},
	{
		name:  "a page's year is never cached",
		edits: []yearEdit{{sfxYearsFile, "\tc.years.Store(pageURL, year)\n", ""}},
		kills: []string{"TestSearchYearsAreCachedPerPage"},
	},
	{
		name:  "a year the page cannot give wipes the result",
		edits: []yearEdit{{sfxYearsFile, "results[i].Year = c.titleYear(ctx, results[i].URL)", "results[i].Year = c.titleYear(ctx, results[i].URL)\n\t\t\tresults[i].Title = \"\""}},
		kills: []string{"TestSearchFillsTheYearsOfSameTitleResults", "TestSearchSurvivesATitlePageThatFails"},
	},
}

func TestYearGuardKillsMutants(t *testing.T) {
	if testing.Short() {
		t.Skip("mutation tests are skipped with -short; run go test ./internal/scraper/providers/startflix -run KillsMutants")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := yearModuleRoot(t)

	// Control: the guards run and pass on the real code. Without it a test
	// that always fails — or one renamed away, which -run would select
	// nothing of — would "kill" every mutant.
	seen := map[string]bool{}
	var guards []string
	for _, m := range yearMutants {
		for _, name := range m.kills {
			if !seen[name] {
				seen[name] = true
				guards = append(guards, name)
			}
		}
	}
	out, err := runYearGuards(t, goBin, root, "", guards)
	if err != nil {
		t.Fatalf("the guard tests must pass on the real code:\n%s", out)
	}
	for _, name := range guards {
		if !strings.Contains(out, "--- PASS: "+name+" (") {
			t.Fatalf("guard %s did not run and pass", name)
		}
	}

	for _, m := range yearMutants {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			overlay := writeYearOverlay(t, root, m.edits)
			out, err := runYearGuards(t, goBin, root, overlay, m.kills)
			if strings.Contains(out, "[build failed]") {
				t.Fatalf("the mutant must build, so the tests — not the compiler — catch it:\n%s", out)
			}
			if err == nil {
				t.Fatalf("mutant survived: the guard tests passed against it\n%s", out)
			}
			for _, name := range m.kills {
				if strings.Contains(out, "--- FAIL: "+name+" (") {
					return
				}
			}
			t.Errorf("the run failed, but not in a guard test %v:\n%s", m.kills, out)
		})
	}
}

func writeYearOverlay(t *testing.T, root string, edits []yearEdit) string {
	t.Helper()
	dir := t.TempDir()
	mutated := map[string]string{}
	for _, e := range edits {
		path := filepath.Join(root, filepath.FromSlash(e.file))
		src, ok := mutated[path]
		if !ok {
			raw, err := os.ReadFile(path) // #nosec G304 -- a module file named in this test
			if err != nil {
				t.Fatal(err)
			}
			src = string(raw)
		}
		from, to := e.from, e.to
		if strings.Contains(src, "\r\n") { // a Windows checkout
			from, to = strings.ReplaceAll(from, "\n", "\r\n"), strings.ReplaceAll(to, "\n", "\r\n")
		}
		if n := strings.Count(src, from); n != 1 {
			t.Fatalf("%s holds the snippet %d times, want 1; move the mutant with the code:\n%s", e.file, n, e.from)
		}
		mutated[path] = strings.Replace(src, from, to, 1)
	}
	replace := map[string]string{}
	for path, src := range mutated {
		out := filepath.Join(dir, filepath.Base(path))
		if err := os.WriteFile(out, []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
		replace[path] = out
	}
	raw, err := jsonx.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return overlay
}

func runYearGuards(t *testing.T, goBin, root, overlay string, tests []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	args := []string{"test", "-count=1", "-v", "-run", "^(" + strings.Join(tests, "|") + ")$"}
	if overlay != "" {
		args = append(args, "-overlay="+overlay)
	}
	args = append(args, "./internal/scraper/providers/startflix")
	cmd := exec.CommandContext(ctx, goBin, args...) // #nosec G204 -- the go toolchain, with this test's own arguments
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func yearModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test's directory")
		}
		dir = parent
	}
}
