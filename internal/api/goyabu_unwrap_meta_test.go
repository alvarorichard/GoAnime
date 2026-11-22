package api

import "testing"

// Tests of the tests for Goyabu's player URL unwrapping (unwrapPlayerURL).
//
// Kimetsu no Yaiba Movie (2026-11-22) fell back to Goyabu, whose player data
// named api.anivideo.net/videohls.php?d=<HLS>&amp;… — HTML-escaped, and a
// wrapper that answers 404 for every file. Each mutant puts part of that back;
// the guard tests it names must fail on it.

const (
	goyabuPkg  = "./internal/scraper/providers/goyabu"
	goyabuFile = "internal/scraper/providers/goyabu/client.go"
)

var goyabuUnwrapMutants = []mutant{
	{name: "the pre-fix fallback: the wrapper is handed over", edits: []codeEdit{{goyabuFile,
		"return netx.ValidateStreamURL(unwrapPlayerURL(bloggerURL), \"Goyabu\")", "return netx.ValidateStreamURL(bloggerURL, \"Goyabu\")"}},
		kills: map[string][]string{goyabuPkg: {"TestGoyabuHandsOverTheHLSBehindTheAnivideoWrapper"}}},
	{name: "a script match is handed over escaped", edits: []codeEdit{{goyabuFile,
		"return netx.ValidateStreamURL(unwrapPlayerURL(matches[1]), \"Goyabu\")", "return netx.ValidateStreamURL(matches[1], \"Goyabu\")"}},
		kills: map[string][]string{goyabuPkg: {"TestGoyabuUnescapesAndUnwrapsWhatTheScriptsName"}}},
	{name: "the address is not unescaped", edits: []codeEdit{{goyabuFile,
		"s := htmlent.UnescapeString(strings.TrimSpace(raw))", "s := strings.TrimSpace(raw + htmlent.UnescapeString(\"\"))"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL", "TestGoyabuHandsOverTheHLSBehindTheAnivideoWrapper"}}},
	{name: "the address keeps its blanks", edits: []codeEdit{{goyabuFile,
		"s := htmlent.UnescapeString(strings.TrimSpace(raw))", "s := htmlent.UnescapeString(raw)"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
	{name: "any host's videohls.php is unwrapped", edits: []codeEdit{{goyabuFile,
		"if strings.EqualFold(u.Hostname(), \"api.anivideo.net\") && strings.Contains(u.Path, \"videohls.php\") {", "if strings.Contains(u.Path, \"videohls.php\") {"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
	{name: "any anivideo endpoint is unwrapped", edits: []codeEdit{{goyabuFile,
		"if strings.EqualFold(u.Hostname(), \"api.anivideo.net\") && strings.Contains(u.Path, \"videohls.php\") {", "if strings.EqualFold(u.Hostname(), \"api.anivideo.net\") {"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
	{name: "the wrapper's host is matched case-sensitively", edits: []codeEdit{{goyabuFile,
		"if strings.EqualFold(u.Hostname(), \"api.anivideo.net\") && strings.Contains(u.Path, \"videohls.php\") {", "if u.Hostname() == \"api.anivideo.net\" && strings.Contains(u.Path, \"videohls.php\") {"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
	{name: "an empty d is handed over", edits: []codeEdit{{goyabuFile,
		"\t\tif d := strings.TrimSpace(u.Query().Get(\"d\")); d != \"\" {\n\t\t\treturn d", "\t\tif d := strings.TrimSpace(u.Query().Get(\"d\")); true {\n\t\t\treturn d"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
	{name: "what is no URL is dropped", edits: []codeEdit{{goyabuFile,
		"\tu, err := url.Parse(s)\n\tif err != nil {\n\t\treturn s\n\t}\n\tif strings.EqualFold(u.Hostname(), \"api.anivideo.net\")", "\tu, err := url.Parse(s)\n\tif err != nil {\n\t\treturn \"\"\n\t}\n\tif strings.EqualFold(u.Hostname(), \"api.anivideo.net\")"}},
		kills: map[string][]string{goyabuPkg: {"TestUnwrapPlayerURL"}}},
}

func TestGoyabuUnwrapGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, goyabuUnwrapMutants)
}
