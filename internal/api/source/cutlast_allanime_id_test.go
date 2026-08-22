package source

import (
	"strings"
	"testing"
)

// legacyExtractAllAnimeID is the pre-Go-1.27 implementation, kept as the oracle
// for the differential test below.
func legacyExtractAllAnimeID(s string) string {
	if IsAllAnimeShortID(s) {
		return s
	}
	if idx := strings.LastIndex(s, "/"); idx >= 0 && idx < len(s)-1 {
		candidate := s[idx+1:]
		if IsAllAnimeShortID(candidate) {
			return candidate
		}
	}
	return s
}

func TestExtractAllAnimeID_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bare short id", "ReooPAxPMsHM4KPMY", "ReooPAxPMsHM4KPMY"},
		{"id at end of url", "https://allanime.to/anime/ReooPAxPMsHM4KPMY", "ReooPAxPMsHM4KPMY"},
		{"trailing slash keeps input", "https://allanime.to/anime/ReooPAxPMsHM4KPMY/", "https://allanime.to/anime/ReooPAxPMsHM4KPMY/"},
		{"no slash and not an id", "naruto", "naruto"},
		{"empty string", "", ""},
		{"only a slash", "/", "/"},
		// IsAllAnimeShortID accepts any alphanumeric-with-letters slug, so a
		// human-readable last segment is also extracted. Pre-existing behaviour,
		// preserved by the CutLast rewrite (see the differential test below).
		{"human readable last segment is still extracted", "https://allanime.to/anime/naruto", "naruto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractAllAnimeID(tt.input); got != tt.want {
				t.Errorf("ExtractAllAnimeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractAllAnimeIDMatchesLegacyImplementation(t *testing.T) {
	corpus := []string{
		"", "/", "//", "a", "a/", "/a", "ReooPAxPMsHM4KPMY", "/ReooPAxPMsHM4KPMY",
		"ReooPAxPMsHM4KPMY/", "https://allanime.to/anime/ReooPAxPMsHM4KPMY",
		"https://allanime.to/anime/ReooPAxPMsHM4KPMY/", "https://allanime.to/anime/naruto",
		"12345", "naruto-shippuuden", "x/y/z",
	}
	for _, in := range corpus {
		if got, want := ExtractAllAnimeID(in), legacyExtractAllAnimeID(in); got != want {
			t.Errorf("input %q: CutLast version = %q, legacy = %q", in, got, want)
		}
	}
}
