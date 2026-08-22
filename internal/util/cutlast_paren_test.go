package util

import "testing"

// TestStrip9AnimeParenMeta pins the exact behaviour of the trailing-parenthesis
// stripper after it was rewritten on top of strings.CutLast (Go 1.27). Every
// branch of the old LastIndex arithmetic has a case here so a regression in the
// cut boundaries cannot slip through.
func TestStrip9AnimeParenMeta(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips SUB DUB episode meta", "Naruto (HD SUB DUB Ep 293/293)", "Naruto"},
		{"strips multilanguage meta", "One Piece (Multilanguage SUB Ep 100)", "One Piece"},
		{"strips HD meta", "Bleach (HD)", "Bleach"},
		{"keeps legitimate subtitle", "Naruto (Shippuuden)", "Naruto (Shippuuden)"},
		// PRE-EXISTING BEHAVIOUR, NOT A REGRESSION: "Dublado" contains "DUB", so
		// the metadata heuristic fires and the marker is stripped, contradicting
		// the function's own doc comment. The differential test below proves the
		// CutLast rewrite did not change this; fixing the heuristic is a separate
		// decision.
		{"dublado marker is stripped by the DUB heuristic", "Black Clover (Dublado)", "Black Clover"},
		{"no parenthesis at all", "Naruto", "Naruto"},
		{"empty string", "", ""},
		{"separator at index zero is never stripped", " (HD SUB)", " (HD SUB)"},
		{"closing paren not at end", "Naruto (HD SUB) extra", "Naruto (HD SUB) extra"},
		{"open paren without close", "Naruto (HD SUB", "Naruto (HD SUB"},
		{"uses the LAST separator", "Naruto (Shippuuden) (HD SUB Ep 5)", "Naruto (Shippuuden)"},
		{"paren glued to title is not a separator", "Naruto(HD SUB)", "Naruto(HD SUB)"},
		{"trailing spaces trimmed after strip", "Naruto  (HD SUB)", "Naruto"},
		{"unicode title preserved", "Mushishi 蟲師 (HD SUB Ep 26)", "Mushishi 蟲師"},
		{"empty parens are not metadata", "Naruto ()", "Naruto ()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strip9AnimeParenMeta(tt.input); got != tt.want {
				t.Errorf("strip9AnimeParenMeta(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStrip9AnimeParenMetaMatchesLegacyImplementation is a differential test:
// it runs the CutLast version against the original LastIndex version over a
// corpus of adversarial inputs and requires byte-identical results.
func TestStrip9AnimeParenMetaMatchesLegacyImplementation(t *testing.T) {
	corpus := []string{
		"", " ", "(", ")", " (", " ()", " ()", "()", "a (", "a )",
		"Naruto", "Naruto (HD)", "Naruto (HD SUB Ep 1)", " (HD SUB)",
		"Naruto (Shippuuden)", "Naruto (Shippuuden) (SUB)", "Naruto (SUB) (Shippuuden)",
		"Naruto (SUB) trailing", "Naruto (SUB", "Naruto SUB)", "Naruto ( )",
		"Naruto ((SUB))", "Naruto (sub)", "Naruto (MULTI)", "Naruto (EP 5)",
		"Naruto (Ep5)", "蟲師 (HD SUB)", "Naruto  (HD)", "Naruto (HD) ",
	}

	for _, in := range corpus {
		want := legacyStrip9AnimeParenMeta(in)
		if got := strip9AnimeParenMeta(in); got != want {
			t.Errorf("input %q: CutLast version = %q, legacy = %q", in, got, want)
		}
	}
}
