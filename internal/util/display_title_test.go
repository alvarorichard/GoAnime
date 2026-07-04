package util

// Tests for the display-title sanitizer split out of SanitizeForFilename
// (2026-07-01). Window titles (mpv, Discord presence) must keep the title's
// own punctuation: the old code ran the FILENAME sanitizer on them, degrading
// "Need for Speed: O Filme" to "Need for Speed O Filme".

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeForDisplayTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"keeps colon, strips tags", "[Movie] [PT-BR] Need for Speed: O Filme", "Need for Speed: O Filme"},
		{"strips language tag", "[PT-BR] Os Simpsons", "Os Simpsons"},
		{"keeps question mark", "[TV] Quem Matou?", "Quem Matou?"},
		{"strips trailing score", "Jujutsu Kaisen 7.27", "Jujutsu Kaisen"},
		{"strips age classification", "Naruto A14", "Naruto"},
		{"collapses inner spaces left by tag removal", "[HD] Bleach [SUB] Thousand-Year", "Bleach Thousand-Year"},
		{"plain title untouched", "Attack on Titan", "Attack on Titan"},
		{"empty stays empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, SanitizeForDisplayTitle(tt.input))
		})
	}
}

func TestStripSourceMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bracket tags removed", "[Movie] [PT-BR] Munique", "Munique"},
		{"punctuation preserved", "Kingsglaive: Final Fantasy XV", "Kingsglaive: Final Fantasy XV"},
		{"9anime paren meta removed", "Boruto (HD SUB DUB Ep 293/293)", "Boruto"},
		{"trailing rating removed", "One Piece 8.62", "One Piece"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, stripSourceMetadata(tt.input))
		})
	}
}

func TestCollapseSpaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"double spaces collapse", "a  b   c", "a b c"},
		{"trims ends", "  abc  ", "abc"},
		{"single spaces untouched", "a b c", "a b c"},
		{"empty stays empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, collapseSpaces(tt.input))
		})
	}
}
