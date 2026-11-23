package pobreflix

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

// Tests of the tests. The guards of the 2026-11-23 rebuild (mirror_query_test,
// sitemap_posttype_test, busca_year_test) are only worth having if they fail
// when the bug comes back: a first search that came back empty, a dead sitemap
// fallback, a year thrown away. Each mutant below puts back one way it can come
// back. The meta test builds the package with the mutant swapped in (go test
// -overlay, so no file is touched), runs the guard tests, and requires one of
// them to fail. A mutant whose snippet is no longer in the code fails here too:
// a refactor cannot quietly retire a guard, it has to carry the mutant along.

type pfEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type pfMutant struct {
	name  string
	edits []pfEdit
	kills []string // at least one must fail against the mutant
}

const (
	pfClient = "internal/scraper/providers/pobreflix/client.go"
	pfV3     = "internal/scraper/providers/pobreflix/v3.go"
	pfTypes  = "internal/scraper/providers/pobreflix/types.go"
)

var pfMutants = []pfMutant{
	{
		name:  "pre-fix: the client starts on the mirror that drops the query",
		edits: []pfEdit{{pfClient, `DefaultBase = "https://pobreflixtvfree.cc"`, `DefaultBase = "https://pobreflixtvfree.cv"`}},
		kills: []string{"TestDefaultBaseIsNotTheRedirectingMirror"},
	},
	{
		name:  "pre-fix: a redirect that dropped the query is not asked again",
		edits: []pfEdit{{pfClient, "rehostedURL(req.URL, resp.Request.URL); ok && !opt.rehosted {", "rehostedURL(req.URL, resp.Request.URL); ok && false {"}},
		kills: []string{"TestClientAsksTheNewMirrorWhenARedirectDropsTheQuery", "TestBuscaSearchOnTheOldMirrorFindsTheTitle"},
	},
	{
		name:  "a re-ask can be re-asked again",
		edits: []pfEdit{{pfClient, "rehostedURL(req.URL, resp.Request.URL); ok && !opt.rehosted {", "rehostedURL(req.URL, resp.Request.URL); ok {"}},
		kills: []string{"TestClientRehostsAtMostOnce"},
	},
	{
		name:  "the re-ask goes to whatever host redirected",
		edits: []pfEdit{{pfClient, "origin, ok := mirrorOf(requested.Scheme+\"://\"+requested.Host, final)", "origin, ok := final.Scheme+\"://\"+final.Host, true"}},
		kills: []string{"TestRehostedURL"},
	},
	{
		name:  "a redirect that kept the query is asked again",
		edits: []pfEdit{{pfClient, "requested.RawQuery == \"\" || final.RawQuery != \"\" {", "requested.RawQuery == \"\" {"}},
		kills: []string{"TestClientDoesNotRehostWhenTheRedirectKeepsTheQuery", "TestRehostedURL"},
	},
	{
		name:  "a request that had no query is asked again",
		edits: []pfEdit{{pfClient, "requested.RawQuery == \"\" || final.RawQuery != \"\" {", "final.RawQuery != \"\" {"}},
		kills: []string{"TestClientDoesNotRehostARequestWithoutAQuery", "TestRehostedURL"},
	},
	{
		name:  "a redirect loop is retried",
		edits: []pfEdit{{pfClient, "return nil, ctx.Err() == nil && !isRedirectLoop(err), err", "return nil, ctx.Err() == nil, err"}},
		kills: []string{"TestClientDoesNotRetryARedirectLoop"},
	},
	{
		name:  "a redirect loop is not recognised by its message",
		edits: []pfEdit{{pfClient, "\t\tstrings.Contains(err.Error(), http.ErrUseLastResponse.Error()) ||\n", ""}},
		kills: []string{"TestIsRedirectLoop"},
	},
	{
		name:  "pre-fix: the post-type sitemaps are not content",
		edits: []pfEdit{{pfClient, "/sitemap-(?:posttype-)?(movies|filmes|filme|series|serie)", "/sitemap-(movies|filmes|series)"}},
		kills: []string{"TestSitemapShardsReadsThePostTypeLayout", "TestSitemapSearchFindsTitlesInThePostTypeLayout", "TestCurrentSitemapReMatches"},
	},
	{
		name:  "every sitemap is content",
		edits: []pfEdit{{pfClient, "/sitemap-(?:posttype-)?(movies|filmes|filme|series|serie)(?:[-.](\\d+))?\\.xml$", "/sitemap-(?:posttype-)?([a-z_-]+?)(?:[-.](\\d+))?\\.xml$"}},
		kills: []string{"TestSitemapShardsReadsThePostTypeLayout", "TestCurrentSitemapReMatches"},
	},
	{
		name:  "the search throws the year away",
		edits: []pfEdit{{pfV3, "\t\t\tYear:    year,\n", "\t\t\tYear:    year[:0],\n"}},
		kills: []string{"TestBuscaSearchCarriesTheYear"},
	},
	{
		name:  "any number is a year",
		edits: []pfEdit{{pfV3, "if f.Year >= 1888 && f.Year <= 2200 {", "if f.Year != 0 {"}},
		kills: []string{"TestBuscaSearchCarriesTheYear"},
	},
	{
		name:  "the model drops the year",
		edits: []pfEdit{{pfTypes, "\t\tYear:   m.Year,\n", ""}},
		kills: []string{"TestToAnimeModelCarriesTheYear"},
	},
}

func TestPobreflixRebuildGuardKillsMutants(t *testing.T) {
	if testing.Short() {
		t.Skip("mutation tests are skipped with -short; run go test ./internal/scraper/providers/pobreflix -run KillsMutants")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := pfModuleRoot(t)

	// Control: the guards run and pass on the real code. Without it a test
	// that always fails — or one renamed away, which -run would select
	// nothing of — would "kill" every mutant.
	seen := map[string]bool{}
	var guards []string
	for _, m := range pfMutants {
		for _, name := range m.kills {
			if !seen[name] {
				seen[name] = true
				guards = append(guards, name)
			}
		}
	}
	out, err := runPfGuards(t, goBin, root, "", guards)
	if err != nil {
		t.Fatalf("the guard tests must pass on the real code:\n%s", out)
	}
	for _, name := range guards {
		if !strings.Contains(out, "--- PASS: "+name+" (") {
			t.Fatalf("guard %s did not run and pass", name)
		}
	}

	for _, m := range pfMutants {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			overlay := writePfOverlay(t, root, m.edits)
			out, err := runPfGuards(t, goBin, root, overlay, m.kills)
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

func writePfOverlay(t *testing.T, root string, edits []pfEdit) string {
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

func runPfGuards(t *testing.T, goBin, root, overlay string, tests []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	args := []string{"test", "-count=1", "-v", "-run", "^(" + strings.Join(tests, "|") + ")$"}
	if overlay != "" {
		args = append(args, "-overlay="+overlay)
	}
	args = append(args, "./internal/scraper/providers/pobreflix")
	cmd := exec.CommandContext(ctx, goBin, args...) // #nosec G204 -- the go toolchain, with this test's own arguments
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func pfModuleRoot(t *testing.T) string {
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
