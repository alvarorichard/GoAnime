package debugreport

import (
	"testing"

	_ "github.com/alvarorichard/Goanime/v2/internal/api/providers" // registers the sources, as the program does at start-up
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rowFor(t *testing.T, kind string) string {
	t.Helper()
	section := sourcesSection(registeredSources())
	for _, r := range section.Rows {
		if r.Key == kind {
			return r.Value
		}
	}
	t.Fatalf("no row for %s in %v", kind, section.Rows)
	return ""
}

// "Goyabu is missing from my search" is answered by this section: it says
// whether the source was switched off, and by what.
func TestSourcesSectionExplainsWhyASourceIsMissing(t *testing.T) {
	descs := registeredSources()
	require.NotEmpty(t, descs, "the providers register themselves at start-up")

	t.Setenv("GOANIME_DISABLED_SOURCES", "Goyabu")
	assert.Contains(t, rowFor(t, "Goyabu"), "disabled")
	assert.Contains(t, rowFor(t, "Goyabu"), "GOANIME_DISABLED_SOURCES")
	assert.Equal(t, "enabled", rowFor(t, "AnimeFire"))
}

func TestSourcesSectionNamesTheOptInForSourcesOffByDefault(t *testing.T) {
	var off string
	for _, d := range registeredSources() {
		if d.DefaultDisabled {
			off = string(d.Kind)
		}
	}
	if off == "" {
		t.Skip("every registered source is on by default")
	}
	assert.Contains(t, rowFor(t, off), "GOANIME_ENABLED_SOURCES")
}

func TestRegisteredSourcesAreSortedForAStableReport(t *testing.T) {
	descs := registeredSources()
	for i := 1; i < len(descs); i++ {
		assert.Less(t, string(descs[i-1].Kind), string(descs[i].Kind))
	}
	assert.Equal(t, len(source.Kinds()), len(descs))
}

func TestCfsolveSectionAlwaysSaysWhetherTheToolIsThere(t *testing.T) {
	t.Setenv("GOANIME_CFSOLVE_BIN", "/nonexistent/challenge")
	t.Setenv("GOANIME_CFSOLVE_AUTODOWNLOAD", "0")

	s := cfsolveSection()

	got := map[string]string{}
	for _, r := range s.Rows {
		got[r.Key] = r.Value
	}
	assert.Equal(t, "not installed", got["Binary"])
	assert.Contains(t, got["Auto-download"], "off")
}
