package api

import "testing"

// Tests of the tests. The yearless-copy guard (ambiguous_year_test.go) is only
// worth having if it fails when the bug comes back: PlutoTV's iCarly, with no
// year, grouped with the 2021 revival instead of the 2007 original. Each mutant
// below puts back one way it can come back, or one plausible slip in the fix;
// the harness (stream_publish_meta_test.go) builds the package with it swapped
// in, runs the guard tests and requires one of them to fail. A mutant whose
// snippet is no longer in the code fails too: a refactor cannot quietly retire
// a guard, it has to carry the mutant along.

const (
	ambYearFile = "internal/api/ambiguous_year.go"
	enhancedGo  = "internal/api/enhanced.go"
)

var ambiguousYearMutants = []mutant{
	{
		name:  "pre-fix: the search never resolves a yearless copy",
		edits: []codeEdit{{enhancedGo, "\t\tresolveAmbiguousYears(validAnimes)\n", ""}},
		kills: map[string][]string{apiPkg: {"TestSearchPipelineResolvesYearlessCopies"}},
	},
	{
		name:  "IMDb asked although the years agree",
		edits: []codeEdit{{ambYearFile, "len(g.yearless) > 0 && len(g.years) >= 2", "len(g.yearless) > 0 && len(g.years) >= 1"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoLookupUnlessYearsDisagree"}},
	},
	{
		name:  "IMDb asked although no copy lacks a year",
		edits: []codeEdit{{ambYearFile, "len(g.yearless) > 0 && len(g.years) >= 2", "len(g.years) >= 2"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoLookupUnlessYearsDisagree"}},
	},
	{
		name:  "anime looked up like a movie or series",
		edits: []codeEdit{{ambYearFile, "if r == nil || (r.MediaType != models.MediaTypeMovie && r.MediaType != models.MediaTypeTV) {", "if r == nil {"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoLookupUnlessYearsDisagree"}},
	},
	{
		name:  "a movie and a series share one group",
		edits: []codeEdit{{ambYearFile, "key := string(r.MediaType) + \"|\" + yearTitleKey(r.Name)", "key := yearTitleKey(r.Name)"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoLookupUnlessYearsDisagree"}},
	},
	{
		name:  "the revival's year, not the original's",
		edits: []codeEdit{{ambYearFile, "if best == 0 || t.Year < best {", "if best == 0 || t.Year > best {"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_PlutoICarlyIsTheOriginal", "TestOriginalYearPicksTheEarliestExactMatch", "TestSearchPipelineResolvesYearlessCopies"}},
	},
	{
		name:  "a game or a TV movie counts as the series",
		edits: []codeEdit{{ambYearFile, "if t.Year <= 0 || !kinds[t.Kind] || yearTitleKey(t.Title) != want {", "if t.Year <= 0 || (!kinds[t.Kind] && false) || yearTitleKey(t.Title) != want {"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoExactTitleLeavesTheCopyAlone", "TestResolveAmbiguousYears_KindDecidesWhichTitleCounts", "TestOriginalYearPicksTheEarliestExactMatch"}},
	},
	{
		name:  "a spin-off with a longer name counts as the title",
		edits: []codeEdit{{ambYearFile, "if t.Year <= 0 || !kinds[t.Kind] || yearTitleKey(t.Title) != want {", "if t.Year <= 0 || !kinds[t.Kind] || (yearTitleKey(t.Title) != want && false) {"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_NoExactTitleLeavesTheCopyAlone"}},
	},
	{
		name:  "a title with no year gives year zero",
		edits: []codeEdit{{ambYearFile, "if t.Year <= 0 || !kinds[t.Kind]", "if !kinds[t.Kind]"}},
		kills: map[string][]string{apiPkg: {"TestOriginalYearPicksTheEarliestExactMatch"}},
	},
	{
		name:  "IMDb asked again on every search",
		edits: []codeEdit{{ambYearFile, "\timdbTitlesCache[key] = titles\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_AskedOncePerTitleAcrossSearches"}},
	},
	{
		name:  "a broad query asks IMDb about every title",
		edits: []codeEdit{{ambYearFile, "\tif len(todo) > maxAmbiguousLookups {", "\tif false {"}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_CapsTheLookups"}},
	},
	{
		name:  "IMDb asked with the scraper's tags",
		edits: []codeEdit{{ambYearFile, "\tname = movie.CleanMediaName(name)\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolveAmbiguousYears_AsksIMDbWithTheCleanTitle", "TestYearTitleKey"}},
	},
	{
		name:  "titles compared with their case",
		edits: []codeEdit{{ambYearFile, "for _, r := range strings.ToLower(yearLookupName(name)) {", "for _, r := range yearLookupName(name) {"}},
		kills: map[string][]string{apiPkg: {"TestYearTitleKey", "TestOriginalYearPicksTheEarliestExactMatch"}},
	},
}

func TestAmbiguousYearGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, ambiguousYearMutants)
}
