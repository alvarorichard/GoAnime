package source

import (
	"context"
	"strings"
)

// A search made on a language tab (PT-BR, ENGLISH) asks for that language.
// Most sources are of one language and are simply on the tab or not; a source
// whose catalog mixes languages (Pluto TV: per title, and by the country of
// the session) reads the language asked for from the search's context and
// keeps the titles that fit.

type languageKey struct{}

// WithLanguage returns ctx asking for titles in lang, a two-letter code ("PT",
// "EN"). An empty or unknown lang asks for nothing and returns ctx unchanged.
func WithLanguage(ctx context.Context, lang string) context.Context {
	lang = strings.ToUpper(strings.TrimSpace(lang))
	if lang != "PT" && lang != "EN" {
		return ctx
	}
	return context.WithValue(ctx, languageKey{}, lang)
}

// LanguageFrom is the language ctx asks for ("PT", "EN"), or "" for any.
func LanguageFrom(ctx context.Context) string {
	lang, _ := ctx.Value(languageKey{}).(string)
	return lang
}

// ScopeLanguage is the language a search scope asks for: "PT" on the PT-BR
// tab, "EN" on the ENGLISH one, "" everywhere else (ALL, MOVIES/TV, a single
// source named with --source).
func ScopeLanguage(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "ptbr", "pt-br":
		return "PT"
	case "english":
		return "EN"
	default:
		return ""
	}
}
