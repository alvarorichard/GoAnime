package netx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Accept-Language has to describe the same browser the User-Agent claims.
//
// Three scrapers used to send Chrome's q-ladder under a Firefox User-Agent — a
// pair no real browser produces, and the exact line copied through countless
// scraping snippets. It is not a theoretical concern in this ecosystem: hosts
// here have been seen matching Accept-Language BY VALUE, answering 403 or 429
// to one specific ladder while serving every other.
//
// So the ladders live here as named constants, one per browser, and these tests
// keep every client referencing one instead of inventing its own.

func TestAcceptLanguage_MatchesTheUserAgentsBrowser(t *testing.T) {
	t.Parallel()
	require.Contains(t, UserAgent, "Firefox",
		"the shared User-Agent changed browser; the q-ladder below has to change with it")

	// Firefox's ladder steps 0.8 / 0.5 / 0.3. Chrome's steps 0.9 / 0.8 / 0.7.
	assert.Equal(t, "pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3", AcceptLanguage)
	assert.NotEqual(t, ChromeAcceptLanguage, AcceptLanguage,
		"the shared UA is Firefox; sending Chrome's ladder with it is the pair that got us blocked")
	assert.Contains(t, EnglishAcceptLanguage, "q=0.5", "the English ladder must also be Firefox's")
}

// A new scraper is written by copying an existing one, which is how the old
// pair spread to three clients. Scanning the tree is the only check that
// catches a literal coming back in a file nobody thought to test.
func TestNoClientHardcodesAnAcceptLanguageOfItsOwn(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	allowed := map[string]bool{
		filepath.Join("internal", "scraper", "netx", "useragent.go"):              true,
		filepath.Join("internal", "scraper", "netx", "useragent_pairing_test.go"): true,
	}

	var offenders []string
	err := filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if allowed[rel] {
			return nil
		}
		body, readErr := os.ReadFile(path) // #nosec G304 -- walking the repo's own source tree in a test
		if readErr != nil {
			return readErr
		}
		// A literal Accept-Language value always contains "q=0." in this tree;
		// referencing the constant does not.
		for _, line := range strings.Split(string(body), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if strings.Contains(trimmed, "Accept-Language") && strings.Contains(trimmed, "q=0.") {
				offenders = append(offenders, rel+": "+trimmed)
			}
		}
		return nil
	})
	require.NoError(t, err)

	assert.Emptyf(t, offenders,
		"Accept-Language must come from a named netx constant so it cannot drift away "+
			"from the User-Agent: %v", offenders)
}

// repoRoot walks up from the test's working directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for range 10 {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate the module root from " + dir)
	return ""
}
