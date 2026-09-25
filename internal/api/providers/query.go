package providers

import "strings"

// What the user typed has to survive the trip to a site's search box.
//
// It did not. util.TreatingAnimeName lowercases the query and joins its words
// with dashes — "o todo poderoso" becomes "o-todo-poderoso" — and it runs on
// BOTH input paths, the command-line argument and the interactive prompt. That
// slug form is a leftover from when a source was addressed by URL slug; every
// source today takes a plain query string, and a site's search engine treats
// the dash as a literal character.
//
// Measured 2026-09-24 through the real fan-out:
//
//	"o-todo-poderoso"   0 results
//	"o todo poderoso"   32 results
//
// SuperFlix had already found this and undone the dashes inside its own client
// (search.go, "CLI args arrive hyphenated like the-boys"). The other three
// sources never did, so every multi-word search in the app was quietly asking
// them for a string no catalogue contains — and the fan-out then reported the
// empty result as if the title did not exist.
//
// Fixing it here rather than at TreatingAnimeName is deliberate: the slug form
// still feeds download folder names and URL building, and this is the one place
// every source's query passes through.

// normalizeSearchQuery turns a slugged query back into words.
//
// Underscores go too — the other separator a pasted slug arrives with. Runs of
// separators collapse, so "the---boys" does not become a query with three
// spaces in it.
func normalizeSearchQuery(q string) string {
	q = strings.ReplaceAll(q, "-", " ")
	q = strings.ReplaceAll(q, "_", " ")
	return strings.Join(strings.Fields(q), " ")
}
