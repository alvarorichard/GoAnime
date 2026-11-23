package modguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// github.com/enetx/utls v1.8.1 is a September 2025 tag, older than the real
// v1.0.x line (November 2026) but sorting above it. It is a broken mirror of
// refraction-networking/utls that imports that module's internal/boring, so
// anything that picks "the latest utls" (go get -u, go mod tidy, an IDE's
// dependency updater) leaves a tree where `go run ./cmd/goanime` stops with
// "use of internal package .../utls/internal/boring not allowed" (2026-11-23).
// The enetx account has since gone from GitHub, so the proxy's cache is all
// that serves these modules; the exclude keeps upgrades off the dead tag.

const (
	badUTLS = "github.com/enetx/utls v1.8.1"
)

// problems reports what is wrong with a go.mod and go.sum, one line each.
func problems(goMod, goSum string) []string {
	var out []string
	if !hasLine(goMod, "exclude "+badUTLS) {
		out = append(out, "go.mod must carry `exclude "+badUTLS+"`")
	}
	for _, line := range strings.Split(goMod, "\n") {
		f := strings.Fields(strings.TrimSpace(line))
		if len(f) >= 2 && f[0] == "github.com/enetx/utls" && f[1] == "v1.8.1" {
			out = append(out, "go.mod requires "+badUTLS)
		}
		if len(f) >= 3 && f[0] == "require" && f[1] == "github.com/enetx/utls" && f[2] == "v1.8.1" {
			out = append(out, "go.mod requires "+badUTLS)
		}
	}
	for _, line := range strings.Split(goSum, "\n") {
		if strings.HasPrefix(line, badUTLS+" ") || strings.HasPrefix(line, badUTLS+"/go.mod ") {
			out = append(out, "go.sum pins "+badUTLS)
		}
	}
	return out
}

// hasLine reports whether text has a line that is exactly want, ignoring the
// space around it. A comment that mentions the directive does not count.
func hasLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", name)) // #nosec G304 -- the module's own file
	require.NoError(t, err)
	return strings.ReplaceAll(string(raw), "\r\n", "\n")
}

// TestGoModStaysOffTheBrokenUTLS is the guard: the repository's own files.
func TestGoModStaysOffTheBrokenUTLS(t *testing.T) {
	t.Parallel()
	assert.Empty(t, problems(readRepoFile(t, "go.mod"), readRepoFile(t, "go.sum")))
}

const goodMod = `module example.test/m

go 1.27.4

require (
	github.com/enetx/utls v1.0.10 // indirect
)

exclude github.com/enetx/utls v1.8.1
`

// TestModguardCatchesEveryWayItCanComeBack is the guard's own check: each
// mutant of a good go.mod/go.sum is one way the bad release returns, and must
// be reported; the good pair must not be.
func TestModguardCatchesEveryWayItCanComeBack(t *testing.T) {
	t.Parallel()
	goodSum := "github.com/enetx/utls v1.0.10 h1:abc=\ngithub.com/enetx/utls v1.0.10/go.mod h1:def=\n"
	assert.Empty(t, problems(goodMod, goodSum), "the good pair passes")

	for name, tc := range map[string]struct{ mod, sum string }{
		"the exclude is deleted":            {strings.Replace(goodMod, "exclude github.com/enetx/utls v1.8.1\n", "", 1), goodSum},
		"the exclude names another version": {strings.Replace(goodMod, "exclude github.com/enetx/utls v1.8.1", "exclude github.com/enetx/utls v1.8.0", 1), goodSum},
		"the exclude is only in a comment":  {strings.Replace(goodMod, "exclude github.com/enetx/utls v1.8.1", "// exclude github.com/enetx/utls v1.8.1", 1), goodSum},
		"the exclude is for another module": {strings.Replace(goodMod, "exclude github.com/enetx/utls v1.8.1", "exclude github.com/enetx/surf v1.8.1", 1), goodSum},
		"a block require is bumped":         {strings.Replace(goodMod, "github.com/enetx/utls v1.0.10", "github.com/enetx/utls v1.8.1", 1), goodSum},
		"a single-line require is bumped":   {strings.Replace(goodMod, "require (\n\tgithub.com/enetx/utls v1.0.10 // indirect\n)", "require github.com/enetx/utls v1.8.1", 1), goodSum},
		"go.sum gains the release":          {goodMod, goodSum + "github.com/enetx/utls v1.8.1 h1:xyz=\n"},
		"go.sum gains the go.mod hash":      {goodMod, goodSum + "github.com/enetx/utls v1.8.1/go.mod h1:xyz=\n"},
	} {
		assert.NotEmpty(t, problems(tc.mod, tc.sum), name)
	}
}

func TestHasLine(t *testing.T) {
	t.Parallel()
	assert.True(t, hasLine("a\n\texclude x v1\nb", "exclude x v1"))
	assert.False(t, hasLine("// exclude x v1", "exclude x v1"))
	assert.False(t, hasLine("exclude x v10", "exclude x v1"))
	assert.False(t, hasLine("", "exclude x v1"))
}
