// Package sourcesguard holds the tests of the tests for the SOURCES screen
// (the ctrl+s checklist that turns search sources on and off). It has no code
// of its own: coverage_test.go proves every function of the feature is
// exercised by a test, and mutation_test.go proves those tests fail when the
// feature breaks.
package sourcesguard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unit is one source file of the feature, the test files that must exercise
// it, and which of its functions count. funcs == nil means every function the
// file declares; otherwise exactly the named ones (a file that predates the
// feature only owes tests for what the feature added or changed).
type unit struct {
	file  string
	tests []string
	funcs []string
}

var featureUnits = []unit{
	{
		file:  "internal/util/source_prefs.go",
		tests: []string{"internal/util/source_prefs_test.go"},
	},
	{
		file:  "internal/api/source/states.go",
		tests: []string{"internal/api/source/states_test.go"},
	},
	{
		file:  "internal/api/source/enablement.go",
		tests: []string{"internal/api/source/states_test.go", "internal/api/source/enablement_test.go"},
		funcs: []string{"IsSearchEnabled"},
	},
	{
		file:  "internal/api/providers/live_search.go",
		tests: []string{"internal/api/providers/source_choice_test.go", "internal/api/providers/live_search_test.go"},
		funcs: []string{"searchOffReason", "planLiveSearch"},
	},
	{
		file:  "internal/api/providers/source_providers.go",
		tests: []string{"internal/api/providers/source_choice_test.go"},
		funcs: []string{"goyabuNote"},
	},
	{
		file:  "internal/tui/sources_screen.go",
		tests: []string{"internal/tui/sources_screen_test.go", "internal/tui/search_home_sources_test.go"},
	},
	{
		file:  "internal/tui/search_home.go",
		tests: []string{"internal/tui/search_home_sources_test.go"},
		funcs: []string{"openSources", "updateSources"},
	},
	{
		file:  "internal/handlers/searchhome.go",
		tests: []string{"internal/handlers/searchhome_sources_test.go"},
		funcs: []string{"sourcesLine", "homeSources", "homeSourceRows", "setHomeSource"},
	},
}

// declared lists the functions and methods src declares, by name.
func declared(t *testing.T, src string) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "src.go", src, 0)
	require.NoError(t, err)
	var names []string
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
			names = append(names, fn.Name.Name)
		}
	}
	return names
}

// referenced lists the names test sources use from outside themselves. The
// parser resolves identifiers declared inside a file (locals, parameters,
// := targets, the file's own helpers) to their declaration; those are not
// uses of the code under test, so a local variable called like a function
// does not count as exercising it. Selectors (x.name) and unresolved names are
// the references that leave the file.
func referenced(t *testing.T, srcs []string) map[string]bool {
	t.Helper()
	used := map[string]bool{}
	for _, src := range srcs {
		f, err := parser.ParseFile(token.NewFileSet(), "test.go", src, 0)
		require.NoError(t, err)
		skip := map[*ast.Ident]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				skip[fn.Name] = true
			}
			return true
		})
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && !skip[id] && id.Obj == nil {
				used[id.Name] = true
			}
			return true
		})
	}
	return used
}

// uncovered returns the functions of src (all of them, or only the named
// ones) that no test source references, sorted. A named function src does not
// declare is reported too: a rename must carry its guard along.
func uncovered(t *testing.T, src string, testSrcs, only []string) []string {
	t.Helper()
	have := declared(t, src)
	want := have
	var missing []string
	if only != nil {
		want = only
		set := map[string]bool{}
		for _, n := range have {
			set[n] = true
		}
		for _, n := range only {
			if !set[n] {
				missing = append(missing, n+" (declared nowhere in the file)")
			}
		}
	}
	used := referenced(t, testSrcs)
	var out []string
	for _, n := range want {
		if !used[n] && !contains(missing, n+" (declared nowhere in the file)") {
			out = append(out, n)
		}
	}
	out = append(out, missing...)
	sort.Strings(out)
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func readModuleFile(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(moduleRoot(t), filepath.FromSlash(rel))) // #nosec G304 -- a module file named in this test
	require.NoError(t, err, "the guard names %s; move the guard with the file", rel)
	return string(raw)
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

// TestEveryFeatureFunctionHasATest is the rule: no function of the SOURCES
// feature ships without a test that calls it.
func TestEveryFeatureFunctionHasATest(t *testing.T) {
	for _, u := range featureUnits {
		t.Run(u.file, func(t *testing.T) {
			var tests []string
			for _, tf := range u.tests {
				tests = append(tests, readModuleFile(t, tf))
			}
			assert.Empty(t, uncovered(t, readModuleFile(t, u.file), tests, u.funcs),
				"%s has functions no test in %v calls", u.file, u.tests)
		})
	}
}

// TestFeatureUnitsNameRealFiles keeps the table honest: a unit whose file or
// test file was renamed or deleted fails here instead of silently guarding
// nothing.
func TestFeatureUnitsNameRealFiles(t *testing.T) {
	require.NotEmpty(t, featureUnits)
	for _, u := range featureUnits {
		assert.FileExists(t, filepath.Join(moduleRoot(t), filepath.FromSlash(u.file)))
		require.NotEmpty(t, u.tests, "%s lists no test file", u.file)
		for _, tf := range u.tests {
			assert.FileExists(t, filepath.Join(moduleRoot(t), filepath.FromSlash(tf)))
		}
	}
}

// The checker itself, on synthetic sources: it must report an untested
// function, accept a tested one, and not be fooled by a local variable.
func TestUncoveredChecker(t *testing.T) {
	const src = `package p
func Alpha() {}
func beta() {}
type T struct{}
func (T) gamma() {}
func declaredOnlyInterface()
`
	t.Run("everything called is covered", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) { Alpha(); beta(); T{}.gamma() }`
		assert.Empty(t, uncovered(t, src, []string{test}, nil))
	})

	t.Run("an untouched function is reported", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) { Alpha(); T{}.gamma() }`
		assert.Equal(t, []string{"beta"}, uncovered(t, src, []string{test}, nil))
	})

	t.Run("a function value counts as used", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) { run(Alpha, beta, T.gamma) }`
		assert.Empty(t, uncovered(t, src, []string{test}, nil))
	})

	t.Run("a local variable of the same name does not count", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) {
	beta := 1
	var gamma = 2
	for Alpha := range xs { _ = Alpha }
	_, _ = beta, gamma
}`
		got := uncovered(t, src, []string{test}, nil)
		assert.Equal(t, []string{"Alpha", "beta", "gamma"}, got)
	})

	t.Run("a test named like the function does not count", func(t *testing.T) {
		test := `package p
func beta(t *testing.T) {}`
		assert.Contains(t, uncovered(t, src, []string{test}, nil), "beta")
	})

	t.Run("a parameter of the same name does not count", func(t *testing.T) {
		test := `package p
func helper(beta int) int { return beta }`
		assert.Contains(t, uncovered(t, src, []string{test}, nil), "beta")
	})

	t.Run("only the named functions are owed a test", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) { Alpha() }`
		assert.Empty(t, uncovered(t, src, []string{test}, []string{"Alpha"}))
		assert.Equal(t, []string{"beta"}, uncovered(t, src, []string{test}, []string{"Alpha", "beta"}))
	})

	t.Run("a renamed function is reported, not forgotten", func(t *testing.T) {
		test := `package p
func TestX(t *testing.T) { Alpha() }`
		got := uncovered(t, src, []string{test}, []string{"Alpha", "Renamed"})
		assert.Equal(t, []string{"Renamed (declared nowhere in the file)"}, got)
	})

	t.Run("no test sources means nothing is covered", func(t *testing.T) {
		assert.Equal(t, []string{"Alpha", "beta", "gamma"}, uncovered(t, src, nil, nil))
	})

	t.Run("declared skips bodiless declarations", func(t *testing.T) {
		assert.Equal(t, []string{"Alpha", "beta", "gamma"}, declared(t, src))
	})
}

func TestContainsAndModuleRoot(t *testing.T) {
	assert.True(t, contains([]string{"a", "b"}, "b"))
	assert.False(t, contains([]string{"a"}, "z"))
	assert.False(t, contains(nil, "z"))
	assert.FileExists(t, filepath.Join(moduleRoot(t), "go.mod"))
}
