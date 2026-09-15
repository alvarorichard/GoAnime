package movie

import (
	"strings"
	"testing"
)

// legacyStripTrailingYear is the pre-Go-1.27 tail of CleanMediaName: the year
// removal built on two strings.LastIndex calls. It is the oracle for the
// differential test.
func legacyStripTrailingYear(name string) string {
	if idx := strings.LastIndex(name, "("); idx > 0 {
		if endIdx := strings.LastIndex(name, ")"); endIdx > idx {
			possibleYear := strings.TrimSpace(name[idx+1 : endIdx])
			if len(possibleYear) == 4 {
				isYear := true
				for _, c := range possibleYear {
					if c < '0' || c > '9' {
						isYear = false
						break
					}
				}
				if isYear {
					name = name[:idx]
				}
			}
		}
	}
	return strings.TrimSpace(name)
}

func TestCleanMediaName_YearStripping(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips trailing year", "Spirited Away (2001)", "Spirited Away"},
		{"keeps non-year parenthetical", "Blade Runner (Final Cut)", "Blade Runner (Final Cut)"},
		{"keeps five digit group", "Movie (20015)", "Movie (20015)"},
		{"keeps non numeric four chars", "Movie (abcd)", "Movie (abcd)"},
		{"uses the last parenthesis pair", "Movie (Director) (1999)", "Movie (Director)"},
		{"no parenthesis", "Akira", "Akira"},
		{"paren at index zero is ignored", "(2001)", "(2001)"},
		{"unclosed paren", "Movie (2001", "Movie (2001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanMediaName(tt.input); got != tt.want {
				t.Errorf("CleanMediaName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanMediaNameYearStrippingMatchesLegacyImplementation(t *testing.T) {
	corpus := []string{
		"", "(", ")", "()", "(2001)", "a (2001)", "a (2001) b", "a (20 1)",
		"a ( 2001 )", "a (Director) (1999)", "a (1999) (Director)", "a (abcd)",
		"a (12345)", "a (123)", "Spirited Away (2001)", "Blade Runner (Final Cut)",
		"Movie (2001", "Movie 2001)", "蟲師 (2005)",
	}
	for _, in := range corpus {
		// CleanMediaName applies tag/underscore normalisation before the year
		// logic; feed the oracle the same pre-normalised input by comparing the
		// year step in isolation.
		if got, want := legacyStripTrailingYear(in), CleanMediaName(in); got != want {
			t.Errorf("input %q: CleanMediaName = %q, legacy year-strip = %q", in, want, got)
		}
	}
}
