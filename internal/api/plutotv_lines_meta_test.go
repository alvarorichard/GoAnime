package api

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// No line of the Pluto TV feature may go untested. The function-level guard
// (TestPlutoTVFunctionsAreFullyCovered) holds the feature's own files; this one
// holds everything else too: every statement of a file that belongs to the
// feature, and every statement that names Pluto TV in a file it only touches
// (a case in a switch of sources, a route, a tab), wherever it is in the
// repository. A line added for Pluto TV without a test that runs it fails here.

// plutoLineRe finds a line written for this feature in a shared file.
var plutoLineRe = regexp.MustCompile(`(?i)pluto`)

// isPlutoFile reports a file that holds nothing but the feature.
func isPlutoFile(slashPath string) bool {
	return strings.Contains(strings.ToLower(slashPath), "plutotv")
}

// plutoLines maps each non-test Go file under root (module-relative, slashes)
// to the lines of it that belong to the feature. whole is set for a file of
// the feature's own, in which every line counts.
func plutoLines(t *testing.T, root string) (lines map[string]map[int]bool, whole map[string]bool) {
	t.Helper()
	lines, whole = map[string]map[int]bool{}, map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "vendor" || name == "testdata" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if isPlutoFile(rel) {
			whole[rel] = true
			return nil
		}
		f, err := os.Open(path) // #nosec G304 -- a source file of this repository
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for n := 1; sc.Scan(); n++ {
			if plutoLineRe.MatchString(sc.Text()) {
				if lines[rel] == nil {
					lines[rel] = map[int]bool{}
				}
				lines[rel][n] = true
			}
		}
		return sc.Err()
	})
	require.NoError(t, err)
	return lines, whole
}

// coverBlock is one statement block of a cover profile.
type coverBlock struct {
	file       string // module-relative, slashes
	start, end int    // lines, inclusive
	count      int
}

var coverRowRe = regexp.MustCompile(`^(.+):(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)$`)

// parseCoverProfile reads a profile written by go test -coverprofile. A block
// listed more than once (as merged profiles list it) counts as run if any
// listing ran it.
func parseCoverProfile(profile, module string) []coverBlock {
	type key struct {
		file       string
		start, end int
	}
	counts := map[key]int{}
	var order []key
	for _, row := range strings.Split(strings.ReplaceAll(profile, "\r\n", "\n"), "\n") {
		m := coverRowRe.FindStringSubmatch(strings.TrimSpace(row))
		if m == nil {
			continue
		}
		start, _ := strconv.Atoi(m[2])
		end, _ := strconv.Atoi(m[3])
		count, _ := strconv.Atoi(m[4])
		k := key{strings.TrimPrefix(m[1], module+"/"), start, end}
		if _, seen := counts[k]; !seen {
			order = append(order, k)
		}
		counts[k] = max(counts[k], count)
	}
	blocks := make([]coverBlock, 0, len(order))
	for _, k := range order {
		blocks = append(blocks, coverBlock{k.file, k.start, k.end, counts[k]})
	}
	return blocks
}

// unrunPlutoLines lists, as "file:line", what belongs to the feature and sits
// in a block no test ran: every line of an unrun block of a whole file (named
// by the block's first line), and each feature line of a shared file.
func unrunPlutoLines(blocks []coverBlock, lines map[string]map[int]bool, whole map[string]bool) []string {
	seen := map[string]bool{}
	for _, b := range blocks {
		if b.count > 0 {
			continue
		}
		if whole[b.file] {
			seen[fmt.Sprintf("%s:%d", b.file, b.start)] = true
			continue
		}
		for n := range lines[b.file] {
			if b.start <= n && n <= b.end {
				seen[fmt.Sprintf("%s:%d", b.file, n)] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// modulePath reads the module line of go.mod.
func modulePath(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "go.mod")) // #nosec G304
	require.NoError(t, err)
	for _, line := range strings.Split(string(raw), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	t.Fatal("go.mod names no module")
	return ""
}

func TestPlutoTVHasNoLineThatNoTestRuns(t *testing.T) {
	skipMutantsWhenShort(t)
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH")
	}
	root := moduleRoot(t)
	module := modulePath(t, root)
	lines, whole := plutoLines(t, root)
	require.NotEmpty(t, whole, "the feature's own files were not found")
	require.NotEmpty(t, lines, "the files the feature touches were not found")

	packages := map[string]bool{}
	for file := range lines {
		packages[filepath.ToSlash(filepath.Dir(file))] = true
	}
	for file := range whole {
		packages[filepath.ToSlash(filepath.Dir(file))] = true
	}
	dirs := make([]string, 0, len(packages))
	for dir := range packages {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	var blocks []coverBlock
	for _, dir := range dirs {
		profile := filepath.Join(t.TempDir(), "cover.out")
		// -short: the unit tests, without the network and without the mutation
		// suites (this test among them, which would otherwise run itself).
		run := exec.Command(goBin, "test", "-short", "-count=1", "-coverprofile="+profile, "./"+dir) // #nosec G204 -- the go toolchain on this repository's own packages
		run.Dir = root
		out, err := run.CombinedOutput()
		require.NoError(t, err, "the tests of %s must pass to be measured:\n%s", dir, out)
		raw, err := os.ReadFile(profile) // #nosec G304
		require.NoError(t, err)
		blocks = append(blocks, parseCoverProfile(string(raw), module)...)
	}

	measured := map[string]bool{}
	for _, b := range blocks {
		measured[b.file] = true
	}
	for file := range whole {
		if strings.HasSuffix(file, "/doc.go") {
			continue // package documentation: no statement to run
		}
		assert.True(t, measured[file], "%s is the feature's own file and no test of its package measures it", file)
	}
	assert.Empty(t, unrunPlutoLines(blocks, lines, whole), "lines of the Pluto TV feature that no unit test runs")
}

// ---- the check itself ------------------------------------------------------

func TestParseCoverProfileReadsBlocksAndMergesRepeats(t *testing.T) {
	profile := "mode: set\r\n" +
		"example.org/m/v2/internal/a/x.go:10.2,12.3 2 1\n" +
		"example.org/m/v2/internal/a/x.go:14.5,14.20 1 0\n" +
		"example.org/m/v2/internal/a/x.go:14.5,14.20 1 3\n" +
		"example.org/m/v2/internal/a/y.go:1.1,9.2 4 0\n" +
		"not a row\n\n"
	assert.Equal(t, []coverBlock{
		{"internal/a/x.go", 10, 12, 1},
		{"internal/a/x.go", 14, 14, 3},
		{"internal/a/y.go", 1, 9, 0},
	}, parseCoverProfile(profile, "example.org/m/v2"), "module-relative files; a block run in any listing is run")
	assert.Empty(t, parseCoverProfile("mode: set\n", "example.org/m/v2"))
}

func TestTheLineCheckSeesALineNoTestRuns(t *testing.T) {
	lines := map[string]map[int]bool{"internal/a/shared.go": {5: true, 20: true, 31: true}}
	whole := map[string]bool{"internal/a/plutotv.go": true}
	blocks := []coverBlock{
		{"internal/a/shared.go", 4, 6, 1},    // the Pluto line 5 runs
		{"internal/a/shared.go", 18, 22, 0},  // the Pluto line 20 does not
		{"internal/a/shared.go", 40, 50, 0},  // unrun, but not this feature's
		{"internal/a/shared.go", 31, 31, 2},  // runs
		{"internal/a/plutotv.go", 7, 9, 0},   // any unrun block of the feature's own file
		{"internal/a/plutotv.go", 12, 15, 4}, // runs
		{"internal/a/other.go", 1, 99, 0},    // another feature entirely
	}
	assert.Equal(t, []string{"internal/a/plutotv.go:7", "internal/a/shared.go:20"}, unrunPlutoLines(blocks, lines, whole))

	blocks[1].count, blocks[4].count = 1, 1
	assert.Empty(t, unrunPlutoLines(blocks, lines, whole), "and nothing once they run")
}

func TestTheLineCheckFindsTheFeatureInTheRepository(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write("go.mod", "module example.org/m/v2\n\ngo 1.24\n")
	write("internal/a/shared.go", "package a\n\nfunc f(s string) int {\n\tif s == \"PlutoTV\" {\n\t\treturn 1\n\t}\n\treturn 0\n}\n")
	write("internal/a/shared_test.go", "package a\n\n// pluto in a test is not the feature's code\n")
	write("internal/a/plutotv.go", "package a\n\nfunc g() {}\n")
	write("internal/providers/plutotv/client.go", "package plutotv\n")
	write("internal/b/none.go", "package b\n")
	write("internal/a/testdata/pluto.go", "package x // pluto\n")
	write("vendor/x/pluto.go", "package x // pluto\n")

	lines, whole := plutoLines(t, root)
	assert.Equal(t, map[string]map[int]bool{"internal/a/shared.go": {4: true}}, lines, "the one line that names it, in the one shared file")
	assert.Equal(t, map[string]bool{"internal/a/plutotv.go": true, "internal/providers/plutotv/client.go": true}, whole, "its own files, by name or by directory")
	assert.Equal(t, "example.org/m/v2", modulePath(t, root))
}
