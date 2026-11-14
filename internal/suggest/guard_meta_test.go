package suggest

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

// This file tests the tests. Each mutant is an exact snippet swap, injected
// with `go test -overlay` in a subprocess (no file on disk is touched), that
// re-creates one way the autocomplete could break: a cache that never holds,
// a ranking that stopped ranking, a stale answer shown under new text. Every
// mutant must still build, and the guard tests named for it must fail.
// A snippet that drifted out of the code fails here too, so the mutants move
// with the code instead of silently testing nothing.

type codeEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type mutant struct {
	name  string
	edits []codeEdit
	kills map[string][]string // package -> guard tests, one of which must fail
	// notRun names cascade stages (a subtest name, spaces as underscores)
	// that must never start: the cascade stops at the broken hop.
	notRun []string
}

const (
	suggestPkg  = "./internal/suggest"
	tuiPkg      = "./internal/tui"
	handlersPkg = "./internal/handlers"
	engineFile  = "internal/suggest/suggest.go"
	cacheFile   = "internal/suggest/cache.go"
	provFile    = "internal/suggest/providers.go"
	homeFile    = "internal/tui/search_home.go"
	handlerFile = "internal/handlers/searchhome.go"
)

var tuiGuards = []string{
	"TestSearchHomeSuggestCascade", "TestSearchHomeSuggestStaleAnswersAreDropped",
	"TestSearchHomeSuggestFailureShowsNothing", "TestSearchHomeSuggestRowCleaning",
	"TestSearchHomeSuggestKeys", "TestSearchHomeSuggestFollowsScope", "TestSearchHomeGhostText",
	"TestSearchHomePrefilledQueryDoesNotAutoAsk", "TestSearchHomeSuggestHintsAndPlain",
}

var suggestMutants = []mutant{
	// --- the engine ---
	{
		name:  "cache never stores an answer",
		edits: []codeEdit{{engineFile, "\te.cache.put(key, items)\n", ""}},
		kills: map[string][]string{suggestPkg: {"TestSuggestCachesSuccessOnly", "TestSuggestCacheExpires"}},
	},
	{
		name:  "cache never expires",
		edits: []codeEdit{{cacheFile, "if !c.now().Before(entry.expires) {", "if false && !c.now().Before(entry.expires) {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestCacheExpires", "TestCacheLRUAndCopies"}},
	},
	{
		name:  "cache never evicts",
		edits: []codeEdit{{cacheFile, "if c.order.Len() > c.max {", "if false && c.order.Len() > c.max {"}},
		kills: map[string][]string{suggestPkg: {"TestCacheLRUAndCopies"}},
	},
	{
		name:  "cache hands out its own slice",
		edits: []codeEdit{{cacheFile, "return slices.Clone(entry.items), true", "return entry.items, true"}},
		kills: map[string][]string{suggestPkg: {"TestCacheLRUAndCopies"}},
	},
	{
		name:  "ranking stopped ranking",
		edits: []codeEdit{{engineFile, "if ti != tj {\n\t\t\treturn ti < tj\n\t\t}", "if ti != tj {\n\t\t\treturn false\n\t\t}"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestRankTiers", "TestSuggestMergesRanksAndDedupes"}},
	},
	{
		name:  "longest completion first",
		edits: []codeEdit{{engineFile, "if ti == 1 {", "if false {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestShortestPrefixFirst"}},
	},
	{
		name:  "duplicates kept",
		edits: []codeEdit{{engineFile, "if key == \"\" || seen[key] {", "if key == \"\" {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestMergesRanksAndDedupes"}},
	},
	{
		name:  "scope kind filter dropped",
		edits: []codeEdit{{engineFile, "if len(kinds) > 0 && !servesAny([]Kind{s.Kind}, kinds) {", "if false {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestLimitAndKindFilter"}},
	},
	{
		name:  "limit not applied",
		edits: []codeEdit{{engineFile, "if len(items) > limit {", "if false {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestLimitAndKindFilter"}},
	},
	{
		name:  "one provider down breaks the box",
		edits: []codeEdit{{engineFile, "if len(lists) == 0 {", "if len(errs) > 0 {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestPartialAndTotalFailure"}},
	},
	{
		name:  "every provider down looks like no results",
		edits: []codeEdit{{engineFile, "if len(lists) == 0 {", "if false {"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestPartialAndTotalFailure"}},
	},
	{
		name:  "cancellation reported as a provider failure",
		edits: []codeEdit{{engineFile, "\tif err := ctx.Err(); err != nil {\n\t\treturn nil, err\n\t}\n\tvar errs []error", "\tvar errs []error"}},
		kills: map[string][]string{suggestPkg: {"TestSuggestTimeoutAndCancel"}},
	},
	{
		name:  "one-letter queries sent upstream",
		edits: []codeEdit{{engineFile, "if len([]rune(q)) < MinQuery {", "if len([]rune(q)) < 0 {"}},
		kills: map[string][]string{suggestPkg: {"TestNormalize", "TestSuggestShortQueryAndNoProviders"}},
	},
	{
		name:  "query length uncapped",
		edits: []codeEdit{{engineFile, "len(r) > MaxQuery", "len(r) > MaxQuery*100"}},
		kills: map[string][]string{suggestPkg: {"TestNormalize"}},
	},
	// --- the providers ---
	{
		name:  "IMDb videos taken for movies",
		edits: []codeEdit{{provFile, "case \"movie\", \"tvMovie\":", "case \"movie\", \"tvMovie\", \"video\":"}},
		kills: map[string][]string{suggestPkg: {"TestIMDbRecordedReply"}},
	},
	{
		name:  "IMDb franchises and people taken for titles",
		edits: []codeEdit{{provFile, "|| !strings.HasPrefix(d.ID, \"tt\")", ""}},
		kills: map[string][]string{suggestPkg: {"TestIMDbFiltersAndLimit"}},
	},
	{
		name:  "IMDb query not folded",
		edits: []codeEdit{{provFile, "folded := fold(query)", "folded := query"}},
		kills: map[string][]string{suggestPkg: {"TestIMDbRecordedReply", "TestIMDbUnfoldableQuery"}},
	},
	{
		name:  "bad status read as an answer",
		edits: []codeEdit{{provFile, "if resp.StatusCode != http.StatusOK {", "if false {"}},
		kills: map[string][]string{suggestPkg: {"TestProviderFailureModes"}},
	},
	{
		name:  "response size uncapped",
		edits: []codeEdit{{provFile, "jsonx.Decode(resp.Body, maxBody, out)", "jsonx.Decode(resp.Body, 0, out)"}},
		kills: map[string][]string{suggestPkg: {"TestProviderFailureModes"}},
	},
	{
		name:  "API calls dressed as a browser",
		edits: []codeEdit{{provFile, "req.Header.Set(\"User-Agent\", netx.APIUserAgent)", "req.Header.Set(\"User-Agent\", \"Mozilla/5.0\")"}},
		kills: map[string][]string{suggestPkg: {"TestRequestHeaders"}},
	},
	{
		name:  "Kitsu title fallback lost",
		edits: []codeEdit{{provFile, "title = strings.TrimSpace(a.Titles[\"en\"])", "title = \"\""}},
		kills: map[string][]string{suggestPkg: {"TestKitsuTitleFallbacks"}},
	},
	// --- the search box ---
	{
		name:  "a stale answer shown under new text",
		edits: []codeEdit{{homeFile, "\tif msg.seq != m.seq {\n\t\treturn\n\t}\n\tm.cancel = nil", "\tm.cancel = nil"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestStaleAnswersAreDropped"}},
	},
	{
		name:  "a stale debounce tick starts a lookup",
		edits: []codeEdit{{homeFile, "if seq != m.seq || m.suggest == nil {", "if m.suggest == nil {"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestStaleAnswersAreDropped"}},
	},
	{
		name:  "old rows stay under new text",
		edits: []codeEdit{{homeFile, "\tm.asked = key\n\tm.dropSuggestions()\n", "\tm.asked = key\n"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestStaleAnswersAreDropped", "TestSearchHomeSuggestKeys"}},
	},
	{
		name:  "typing on never cancels the lookup in flight",
		edits: []codeEdit{{homeFile, "\tif m.cancel != nil {\n\t\tm.cancel()\n\t\tm.cancel = nil\n\t}\n", "\tm.cancel = nil\n"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestStaleAnswersAreDropped"}},
	},
	{
		name:  "a failed lookup shows its stale rows",
		edits: []codeEdit{{homeFile, "\tif msg.err != nil {\n\t\treturn\n\t}\n\tquery :=", "\tquery :="}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestFailureShowsNothing"}},
	},
	{
		name:  "dropdown uncapped",
		edits: []codeEdit{{homeFile, "if len(m.suggestions) == maxHomeSuggestions {", "if false {"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestRowCleaning"}},
	},
	{
		name:  "the typed text offered back as a suggestion",
		edits: []codeEdit{{homeFile, "|| strings.EqualFold(title, query) {", "|| (strings.EqualFold(title, query) && false) {"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestRowCleaning"}},
	},
	{
		name:  "multi-line titles break the dropdown",
		edits: []codeEdit{{homeFile, "title := singleLine(item.Title)\n\t\t// A row", "title := item.Title\n\t\t// A row"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestRowCleaning"}},
	},
	{
		name:  "enter ignores the highlighted suggestion",
		edits: []codeEdit{{homeFile, "if m.sugCursor >= 0 {\n\t\t\tquery =", "if false {\n\t\t\tquery ="}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestCascade"}},
	},
	{
		name:  "esc skips closing the dropdown",
		edits: []codeEdit{{homeFile, "\t\t\tif m.suggestionsOpen() {\n\t\t\t\tm.dropSuggestions()", "\t\t\tif false {\n\t\t\t\tm.dropSuggestions()"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestKeys"}},
	},
	{
		name:  "completing re-asks for the completed title",
		edits: []codeEdit{{homeFile, "\tm.asked = m.suggestKey()\n\tm.dropSuggestions()\n}", "\tm.dropSuggestions()\n}"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestKeys"}},
	},
	{
		name:  "right completes mid-text",
		edits: []codeEdit{{homeFile, "&& m.input.Position() >= len([]rune(m.input.Value())) {", "{"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestKeys"}},
	},
	{
		name:  "down never leaves the dropdown",
		edits: []codeEdit{{homeFile, "\tcase next >= len(m.suggestions):\n\t\tm.sugCursor = -1\n\t\treturn false", "\tcase next >= len(m.suggestions):\n\t\treturn true"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestKeys"}},
	},
	{
		name:  "switching scope keeps the old scope's rows",
		edits: []codeEdit{{homeFile, "m.scopeIdx = (m.scopeIdx + 1) % len(m.scopes)\n\t\t\tcmd := m.scopeChanged()\n\t\t\treturn m, cmd", "m.scopeIdx = (m.scopeIdx + 1) % len(m.scopes)\n\t\t\treturn m, nil"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestFollowsScope"}},
	},
	{
		name:  "a prefilled box asks at once",
		edits: []codeEdit{{homeFile, "\tm.asked = m.suggestKey()\n\tm.notice = singleLine(opts.Notice)", "\tm.notice = singleLine(opts.Notice)"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomePrefilledQueryDoesNotAutoAsk"}},
	},
	{
		name:  "the footer forgets the completion key",
		edits: []codeEdit{{homeFile, "if _, ok := m.completionPick(); ok {\n\t\tright :=", "if false {\n\t\tright :="}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeSuggestHintsAndPlain"}},
	},
	{
		name:  "the inline ghost text is off",
		edits: []codeEdit{{homeFile, "input.ShowSuggestions = true", "input.ShowSuggestions = false"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeGhostText"}},
	},
	{
		name:  "the ghost is not refreshed after a key",
		edits: []codeEdit{{homeFile, "\tm.syncGhost()\n\treturn model, cmd", "\treturn model, cmd"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeGhostText"}},
	},
	{
		name:  "the ghost shows a row that does not continue the text",
		edits: []codeEdit{{homeFile, "if strings.HasPrefix(strings.ToLower(s.Title), typed) {", "if strings.HasPrefix(strings.ToLower(s.Title), typed) || true {"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeGhostText"}},
	},
	{
		name:  "the ghost ignores the highlighted row",
		edits: []codeEdit{{homeFile, "if m.sugCursor >= 0 {\n\t\treturn m.sugCursor, true", "if false {\n\t\treturn m.sugCursor, true"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeGhostText"}},
	},
	// --- the wiring ---
	{
		name:  "the home screen opened without autocomplete",
		edits: []codeEdit{{handlerFile, "\t\tSuggest:       homeSuggest,\n", ""}},
		kills: map[string][]string{handlersPkg: {"TestSearchHomeWiresAutocomplete"}},
	},
	{
		name:  "the scope never narrows the kinds",
		edits: []codeEdit{{handlerFile, "suggest.KindsForScope(scope)))", "nil))"}},
		kills: map[string][]string{handlersPkg: {"TestHomeSuggestPassesTheScope"}},
	},
	{
		name:   "cascade stops at the broken hop",
		edits:  []codeEdit{{homeFile, "Detail: singleLine(item.Detail)})", "Detail: \"\"})"}},
		kills:  map[string][]string{tuiPkg: {"TestSearchHomeSuggestCascade"}},
		notRun: []string{"4_down_highlights,_enter_searches_the_pick"},
	},
}

func TestSuggestGuardKillsMutants(t *testing.T) {
	if testing.Short() {
		t.Skip("the mutants rebuild three packages; skipped with -short")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := moduleRoot(t)

	// Control: every guard passes on the real code, and really runs.
	guards := map[string][]string{suggestPkg: nil, tuiPkg: tuiGuards, handlersPkg: nil}
	for _, m := range suggestMutants {
		for pkg, tests := range m.kills {
			if pkg != tuiPkg {
				guards[pkg] = appendMissing(guards[pkg], tests...)
			}
		}
	}
	for pkg, tests := range guards {
		out, err := runGuardTests(t, goBin, root, "", pkg, tests)
		require.NoError(t, err, "the guard tests must pass on the real code:\n%s", out)
		for _, name := range tests {
			require.Contains(t, out, "--- PASS: "+name+" (", "guard %s did not run and pass in %s", name, pkg)
		}
	}

	for _, m := range suggestMutants {
		t.Run(m.name, func(t *testing.T) {
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
				for _, stage := range m.notRun {
					assert.NotContains(t, out, "=== RUN   TestSearchHomeSuggestCascade/"+stage, "stage %s ran after the cascade broke", stage)
				}
			}
		})
	}
}

func appendMissing(have []string, add ...string) []string {
	for _, a := range add {
		found := false
		for _, h := range have {
			found = found || h == a
		}
		if !found {
			have = append(have, a)
		}
	}
	return have
}

// writeMutantOverlay applies edits in memory and writes an overlay file that
// swaps the mutated sources in for one go build.
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

// runGuardTests runs the named top-level tests of pkg, optionally under an
// overlay, and returns the verbose output.
func runGuardTests(t *testing.T, goBin, root, overlay, pkg string, tests []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
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

func moduleRoot(t *testing.T) string {
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
