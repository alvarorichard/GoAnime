package modguard

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The github.com/enetx account (surf, g, http, quic-go, utls) left GitHub
// between 2026-11-16 and 2026-11-19, and the name is free for anyone to
// register: a version published under it would not be the author's. GoAnime
// builds against its own copies of those five modules instead,
// github.com/alvarorichard/enetx-g, -utls, -http, -quic-go and -surf, at the
// versions it already used, so nothing may name the old path again. A
// dependency that pulls it back in (a `go get -u` that reaches a newer surf,
// say) would put an unowned name in the build; this fails when it does.
//
// Before the copies, the module path also had a stale September 2025 tag of
// enetx/utls, v1.8.1, that sorted above the real v1.0.x line and did not build;
// it cannot come back either, because the path cannot.

const gone = "github.com/enetx/"

// modProblems reports what is wrong with a go.mod and go.sum, one line each.
// A comment may mention the old path; a requirement, replace or exclude may not.
func modProblems(goMod, goSum string) []string {
	var out []string
	for _, line := range strings.Split(goMod, "\n") {
		if code := strings.TrimSpace(strings.SplitN(line, "//", 2)[0]); strings.Contains(code, gone) {
			out = append(out, "go.mod names "+strings.TrimSpace(code))
		}
	}
	for _, line := range strings.Split(goSum, "\n") {
		if strings.HasPrefix(line, gone) {
			out = append(out, "go.sum pins "+strings.Fields(line)[0])
		}
	}
	return out
}

// sourceProblems reports every import of the old path in the given files.
func sourceProblems(files map[string]string) []string {
	var out []string
	for name, src := range files {
		f, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range f.Imports {
			if path, err := strconv.Unquote(imp.Path.Value); err == nil && strings.HasPrefix(path, gone) {
				out = append(out, name+" imports "+path)
			}
		}
	}
	return out
}

func repoRoot() string { return filepath.Join("..", "..") }

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(), name)) // #nosec G304 -- the module's own file
	require.NoError(t, err)
	return strings.ReplaceAll(string(raw), "\r\n", "\n")
}

// TestNothingNamesTheGoneAccount is the guard: the repository's own files.
func TestNothingNamesTheGoneAccount(t *testing.T) {
	t.Parallel()
	assert.Empty(t, modProblems(readRepoFile(t, "go.mod"), readRepoFile(t, "go.sum")))

	files := map[string]string{}
	require.NoError(t, filepath.WalkDir(repoRoot(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "modguard_test.go") {
			raw, err := os.ReadFile(path) // #nosec G304 -- a source file of this module
			if err != nil {
				return err
			}
			files[path] = string(raw)
		}
		return nil
	}))
	require.NotEmpty(t, files, "the walk must find the module's sources")
	assert.Empty(t, sourceProblems(files))
}

const goodMod = `module example.test/m

go 1.27.4

require (
	github.com/alvarorichard/enetx-surf v1.0.256
	github.com/alvarorichard/enetx-utls v1.0.10 // indirect
)

// formerly github.com/enetx/surf, which left GitHub
`

const goodSum = "github.com/alvarorichard/enetx-surf v1.0.256 h1:abc=\ngithub.com/alvarorichard/enetx-surf v1.0.256/go.mod h1:def=\n"

// TestModguardCatchesEveryWayItCanComeBack is the guard's own check: each
// mutant is one way the old path returns, and must be reported; the good files,
// with the copies and a comment that remembers the old name, must not be.
func TestModguardCatchesEveryWayItCanComeBack(t *testing.T) {
	t.Parallel()
	assert.Empty(t, modProblems(goodMod, goodSum), "the copies and a comment pass")
	assert.Empty(t, sourceProblems(map[string]string{"a.go": "package a\n\nimport \"github.com/alvarorichard/enetx-surf\"\n"}))

	for name, tc := range map[string]struct{ mod, sum string }{
		"a block require is back":       {strings.Replace(goodMod, "github.com/alvarorichard/enetx-surf v1.0.256", "github.com/enetx/surf v1.0.256", 1), goodSum},
		"an indirect require is back":   {strings.Replace(goodMod, "github.com/alvarorichard/enetx-utls v1.0.10", "github.com/enetx/utls v1.0.10", 1), goodSum},
		"a single-line require is back": {strings.Replace(goodMod, "require (\n\tgithub.com/alvarorichard/enetx-surf v1.0.256\n\tgithub.com/alvarorichard/enetx-utls v1.0.10 // indirect\n)", "require github.com/enetx/g v1.1.1", 1), goodSum},
		"a replace to it":               {goodMod + "\nreplace github.com/alvarorichard/enetx-surf => github.com/enetx/surf v1.0.256\n", goodSum},
		"an exclude of it":              {goodMod + "\nexclude github.com/enetx/utls v1.8.1\n", goodSum},
		"go.sum gains a module":         {goodMod, goodSum + "github.com/enetx/surf v1.0.257 h1:xyz=\n"},
		"go.sum gains a go.mod hash":    {goodMod, goodSum + "github.com/enetx/g v1.1.1/go.mod h1:xyz=\n"},
	} {
		assert.NotEmpty(t, modProblems(tc.mod, tc.sum), name)
	}

	for name, src := range map[string]string{
		"a plain import": "package a\n\nimport \"github.com/enetx/surf\"\n",
		"an aliased one": "package a\n\nimport s \"github.com/enetx/surf\"\n",
		"a blank import": "package a\n\nimport _ \"github.com/enetx/utls\"\n",
		"a block import": "package a\n\nimport (\n\t\"fmt\"\n\t\"github.com/enetx/g\"\n)\n",
		"a sub-package":  "package a\n\nimport \"github.com/enetx/http/httptrace\"\n",
	} {
		assert.NotEmpty(t, sourceProblems(map[string]string{"a.go": src}), name)
	}
	assert.Empty(t, sourceProblems(map[string]string{"a.go": "package a\n\n// see github.com/enetx/surf\nvar x = \"github.com/enetx/surf\"\n"}),
		"only an import counts, not a comment or a string")
}
