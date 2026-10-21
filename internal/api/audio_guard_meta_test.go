package api

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests. The Todo Poderoso guard (topcine_dub_cascade_test.go,
// panel_audio_test.go, and the startflix package's audio_test.go) is only
// worth having if it fails when the bug comes back — a guard that can never
// fire is worse than none. Each mutant below puts back one way it can come
// back: the code as it was before the fix, or a plausible slip in the fix.
// The meta test builds the package with the mutant swapped in (go test
// -overlay, so no file is touched), runs the guard tests, and requires them
// to fail. A mutant whose snippet is no longer in the code fails here too: a
// refactor cannot quietly retire a guard, it has to carry the mutant along.

// codeEdit swaps one exact snippet of a module file.
type codeEdit struct {
	file     string // module-relative, slash-separated
	from, to string // from must occur exactly once
}

type mutant struct {
	name  string
	edits []codeEdit
	// kills maps a package to the guard tests in it; in every package at
	// least one of them must fail against the mutant.
	kills map[string][]string
	// stopsAt, for a cascade, is the stage the run must stop at: the stages
	// after it must not run.
	stopsAt string
}

const (
	apiPkg       = "./internal/api"
	startflixPkg = "./internal/scraper/providers/startflix"
	sfxAPIFile   = "internal/api/startflix.go"
	sfxTypesFile = "internal/scraper/providers/startflix/types.go"
	sfxClient    = "internal/scraper/providers/startflix/client.go"
)

// The guard as it stands. The pre-fix mutant is the whole report: no swap
// question, the fallback trusted, mpv told Portuguese.
var audioGuardMutants = []mutant{
	{
		name: "pre-fix: original audio passed off as the dub",
		edits: []codeEdit{
			{sfxAPIFile, "\tif have != want {\n\t\tif err := acceptAudioSwap(", "\tif false && have != want {\n\t\tif err := acceptAudioSwap("},
			{sfxAPIFile, "util.SetGlobalAudioLanguage(audioLangForStartFlix(have))", "util.SetGlobalAudioLanguage(audioLangForStartFlix(startflix.AudioDubbed))"},
		},
		kills: map[string][]string{apiPkg: {"TestTopCineTodoPoderosoCascade", "TestTopCineTodoPoderosoOriginalOnlyWhenAsked", "TestPobreflixDeadPanelPlaysThroughVIDSRC"}},
	},
	{
		name:  "swap question skipped",
		edits: []codeEdit{{sfxAPIFile, "\tif have != want {\n\t\tif err := acceptAudioSwap(", "\tif false && have != want {\n\t\tif err := acceptAudioSwap("}},
		kills: map[string][]string{apiPkg: {"TestTopCineTodoPoderosoCascade", "TestPanelMovieDeadDubAsksBeforeLegendado", "TestPanelSeriesFallbackAsksBeforeOriginalAudio"}},
	},
	{
		name:  "mpv told the wanted audio, not the one playing",
		edits: []codeEdit{{sfxAPIFile, "util.SetGlobalAudioLanguage(audioLangForStartFlix(have))", "util.SetGlobalAudioLanguage(audioLangForStartFlix(want))"}},
		kills: map[string][]string{apiPkg: {"TestTopCineTodoPoderosoOriginalOnlyWhenAsked", "TestPanelMovieLegendadoOnlyPlaysUnasked"}},
	},
	{
		name:  "every swap taken for a single-audio listing",
		edits: []codeEdit{{sfxAPIFile, "if len(offered) > 0 && !slices.Contains(offered, want) {", "if true || slices.Contains(offered, want) {"}},
		kills: map[string][]string{apiPkg: {"TestTopCineTodoPoderosoCascade", "TestPanelMovieDeadDubAsksBeforeLegendado"}},
	},
	{
		name:  "an unlabelled fallback stream trusted as the dub",
		edits: []codeEdit{{sfxAPIFile, "\t\t\t\tfs.Audio = startflix.AudioSubtitled\n", "\t\t\t\tfs.Audio = startflix.AudioDubbed\n"}},
		kills: map[string][]string{apiPkg: {"TestPanelSeriesFallbackAsksBeforeOriginalAudio"}},
	},
	{
		name:  "vidsrc stream does not say it is the original audio",
		edits: []codeEdit{{"internal/scraper/providers/startflix/vidsrc.go", "\t\tAudio:   AudioSubtitled,\n\t}, u.Host, nil", "\t}, u.Host, nil"}},
		kills: map[string][]string{startflixPkg: {"TestVIDSRCMovieChainFixture", "TestVIDSRCTVSeasonAndEpisode"}},
	},
	{
		name:  "tallest stream wins whatever its audio",
		edits: []codeEdit{{sfxClient, "if am, bm := audioMatches(a.Audio, want), audioMatches(b.Audio, want); am != bm {", "if am, bm := true, true; am != bm {"}},
		kills: map[string][]string{
			startflixPkg: {"TestResolveStreamForPrefersWantedAudio"},
			apiPkg:       {"TestPanelMovieMixedAudioPlaysTheDub"},
		},
	},
	{
		name:  "player labels ignored",
		edits: []codeEdit{{sfxTypesFile, "\tlabel := strings.ToLower(p.Label)\n\tswitch {\n", "\tlabel := strings.ToLower(p.Label)\n\tswitch {\n\tcase label != \"\":\n\t\treturn \"\"\n"}},
		kills: map[string][]string{
			startflixPkg: {"TestPlayerAudio", "TestListedAudios"},
			apiPkg:       {"TestPanelMovieMixedAudioPlaysTheDub", "TestPanelMovieLegendadoOnlyPlaysUnasked", "TestTopCineTodoPoderosoCascade"},
		},
	},
	{
		name:  "yt-dlp races both audios at once",
		edits: []codeEdit{{sfxClient, "for _, round := range [][]Player{wanted, others} {", "for _, round := range [][]Player{append(wanted, others...)} {"}},
		kills: map[string][]string{startflixPkg: {"TestResolveStreamForYtDlpTriesWantedAudioFirst"}},
	},
	{
		name:  "TopCine's 1take dub left unexplained",
		edits: []codeEdit{{"internal/api/topcine.go", "\tif !isTopCine(media) {\n\t\treturn \"\"\n\t}\n\tlangs, err := topCineLanguages(media)", "\tif media != nil {\n\t\treturn \"\"\n\t}\n\tlangs, err := topCineLanguages(media)"}},
		kills: map[string][]string{apiPkg: {"TestTopCineTodoPoderosoCascade"}},
	},
	{
		name:  "a season's listed audio forgotten",
		edits: []codeEdit{{sfxAPIFile, "\trememberListedAudio(media.URL, season.Key(), audio)\n", ""}},
		kills: map[string][]string{apiPkg: {"TestWantedStreamAudioFollowsTheListedSeason"}},
	},
	{
		// The cascade itself: a hop that breaks early stops it there, so the
		// failure names that hop instead of everything downstream.
		name:    "cascade stops at the broken hop",
		edits:   []codeEdit{{"internal/scraper/providers/topcine/parse.go", "Title{Kind: KindMovie, TMDBID: id, PlayerURL", "Title{Kind: KindMovie, TMDBID: id + 1, PlayerURL"}},
		kills:   map[string][]string{apiPkg: {"TestTopCineTodoPoderosoCascade"}},
		stopsAt: "2 title page names TMDB 310",
	},
}

func TestAudioGuardKillsMutants(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := moduleRoot(t)

	// Control: the guard tests run and pass on the real code. Without it a
	// test that always fails — or one renamed away, which -run would select
	// nothing of — would "kill" every mutant.
	guards := map[string][]string{}
	for _, m := range audioGuardMutants {
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

	for _, m := range audioGuardMutants {
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

// cascadeStageRe reads a cascade stage's number from go test -v output.
var cascadeStageRe = regexp.MustCompile(`=== RUN\s+\S+/(\d+)_`)

// writeMutantOverlay writes the mutated files and the -overlay file that
// swaps them in for the build.
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

// runGuardTests runs the named tests of one package, with the overlay when
// one is given, and returns the verbose output.
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
