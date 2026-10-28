package api

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests, for the Pobreflix gate (pobreflix_gated_test.go and the
// pobreflix package's gated_test.go). Same method as audio_guard_meta_test.go:
// each mutant puts back one way the bug can come back — the code as it was
// before the fix, or a plausible slip in it — builds with go test -overlay, and
// the guard tests must fail against it. A mutant whose snippet is no longer in
// the code fails here too, so a refactor cannot quietly retire a guard.

const (
	pobreflixPkg   = "./internal/scraper/providers/pobreflix"
	pbTypesFile    = "internal/scraper/providers/pobreflix/types.go"
	pbEmbeddedFile = "internal/scraper/providers/pobreflix/embedded.go"
	pbAPIFile      = "internal/api/pobreflix.go"
	pbSitemapFile  = "internal/scraper/providers/pobreflix/sitemap.go"
)

var pobreflixGateMutants = []mutant{
	{
		name: "pre-fix: any listed URL is a file",
		edits: []codeEdit{
			{pbTypesFile, "return isMediaFile(s.URL) || isMediaFile(s.Download)", "return strings.TrimSpace(s.URL) != \"\" || strings.TrimSpace(s.Download) != \"\""},
		},
		kills: map[string][]string{
			pobreflixPkg: {"TestServerDirectVersusGated", "TestParseTitle_GatedMovieIsNotDirect"},
			apiPkg:       {"TestPobreflixGatedMovieCascade", "TestPobreflixGatedMoviePlaysOriginalWhenAsked"},
		},
	},
	{
		name:  "a movie with only player pages still takes the direct path",
		edits: []codeEdit{{pbTypesFile, "return len(DirectServers(t.Servers)) > 0", "return len(t.Servers) > 0"}},
		kills: map[string][]string{
			pobreflixPkg: {"TestParseTitle_GatedMovieIsNotDirect"},
			apiPkg:       {"TestPobreflixGatedMovieCascade"},
		},
	},
	{
		name:  "a player page is handed out as the stream",
		edits: []codeEdit{{pbAPIFile, "rememberPobreflixGate(titleURL, hosts)\n\t\treturn nil, errPobreflixGated", "rememberPobreflixGate(titleURL, hosts)\n\t\treturn servers, nil"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixGatedEpisodeIsNotAFile", "TestPobreflixGateLine"}},
	},
	{
		name:  "the gate is never explained to the user",
		edits: []codeEdit{{"internal/api/startflix.go", "\tif tcLine == \"\" {\n\t\ttcLine = pobreflixGateLine(media.URL)\n\t}\n", ""}},
		kills: map[string][]string{apiPkg: {"TestPobreflixGatedMovieCascade"}},
	},
	{
		name:  "the gate is not recorded when the title loads",
		edits: []codeEdit{{pbAPIFile, "\t\t_, _ = pobreflixFiles(media.URL, title.Servers)\n", ""}},
		kills: map[string][]string{apiPkg: {"TestPobreflixGatedMovieCascade"}},
	},
	{
		name:  "the episode id is a number again",
		edits: []codeEdit{{pbEmbeddedFile, "ID        json.RawMessage `json:\"id\"`", "ID        int             `json:\"id\"`"}},
		kills: map[string][]string{
			pobreflixPkg: {"TestEpisode_StringIDDecodes"},
			apiPkg:       {"TestPobreflixGatedEpisodeIsNotAFile"},
		},
	},
	{
		// The opposite failure: gating so hard that real files stop playing.
		name:  "public mp4s are treated as player pages",
		edits: []codeEdit{{pbTypesFile, "\t\".mp4\": true, \".m4v\": true,", "\t\".m4v\": true,"}},
		kills: map[string][]string{
			pobreflixPkg: {"TestServerDirectVersusGated", "TestParseTitle_EmbeddedMovie"},
			apiPkg:       {"TestPobreflixDirectMovie"},
		},
	},
	{
		name: "pre-fix: the corpus only knows the .html form",
		edits: []codeEdit{{pbSitemapFile,
			"currentRecordRe  = regexp.MustCompile(`^/(filme|serie)/([^/]+?)(?:\\.html)?$`)",
			"currentRecordRe  = regexp.MustCompile(`^/(filme|serie)/([^/]+)\\.html$`)"}},
		kills: map[string][]string{pobreflixPkg: {"TestEntryFromURL_BareSitemapURLs", "TestRank_FindsBareSitemapTitles"}},
	},
	{
		name:  "a bare sitemap URL is indexed as it is, not as the .html page the app opens",
		edits: []codeEdit{{pbSitemapFile, "		u.Path += \".html\"\n", ""}},
		kills: map[string][]string{pobreflixPkg: {"TestEntryFromURL_BareSitemapURLs", "TestRank_FindsBareSitemapTitles"}},
	},
	{
		// The cascade itself: a hop that breaks early stops it there.
		name:    "cascade stops at the broken hop",
		edits:   []codeEdit{{pbAPIFile, "\tif err == nil && title.HasDirect() {\n\t\tif title.Kind == pobreflix.KindMovie {\n\t\t\treturn []models.Episode{{", "\tif err == nil && true {\n\t\tif title.Kind == pobreflix.KindMovie {\n\t\t\treturn []models.Episode{{"}},
		kills:   map[string][]string{apiPkg: {"TestPobreflixGatedMovieCascade"}},
		stopsAt: "2 listing reaches the panel by IMDb id",
	},
}

func TestPobreflixGateKillsMutants(t *testing.T) {
	goBin, root := requireGo(t)

	// Control: the guard tests run and pass on the real code.
	guards := map[string][]string{}
	for _, m := range pobreflixGateMutants {
		for pkg, tests := range m.kills {
			guards[pkg] = append(guards[pkg], tests...)
		}
	}
	for pkg, tests := range guards {
		out, err := runGuardTests(t, goBin, root, "", pkg, tests)
		require.NoError(t, err, "the guard tests must pass on the real code:\n%s", out)
		for _, name := range tests {
			require.Contains(t, out, "--- PASS: "+name+" (", "guard %s did not run and pass in %s", name, pkg)
		}
	}

	for _, m := range pobreflixGateMutants {
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
				if m.stopsAt != "" {
					assert.Contains(t, out, `cascade stopped: stage "`+m.stopsAt+`" failed`)
					broken, err := strconv.Atoi(strings.Fields(m.stopsAt)[0])
					require.NoError(t, err, "a cascade stage name starts with its number")
					for _, match := range cascadeStageRe.FindAllStringSubmatch(out, -1) {
						ran, _ := strconv.Atoi(match[1])
						assert.LessOrEqual(t, ran, broken, "stage %d ran after the broken stage %d", ran, broken)
					}
				}
			}
		})
	}
}

// requireGo returns the go toolchain and the module root, skipping where there
// is no toolchain to build the mutants with.
func requireGo(t *testing.T) (goBin, root string) {
	t.Helper()
	goBin, err := lookGo()
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	return goBin, moduleRoot(t)
}

func lookGo() (string, error) { return exec.LookPath("go") }
