package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"normal", "Naruto", "Naruto"},
		{"english tag", "Naruto [English]", "Naruto"},
		{"ptbr tag", "Naruto [PT-BR]", "Naruto"},
		{"portugues tag", "Naruto [Português]", "Naruto"},
		{"dub parens", "Naruto (Dublado)", "Naruto"},
		{"sub parens", "Naruto (Legendado)", "Naruto"},
		{"forbidden chars", `a/b\c:d*e?f"g<h>i|j`, "a_b_c_d_e_f_g_h_i_j"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sanitizeFilename(tt.in))
		})
	}
}

func TestLanguagePriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"ptbr tag", "[PT-BR] Naruto", 0},
		{"portuguese tag", "[Portuguese] Naruto", 0},
		{"portugues accented", "[Português] Naruto", 0},
		{"movie + ptbr", "[Movie] [PT-BR] Inception", 0},
		{"multilanguage", "[Multilanguage] Naruto", 1},
		{"english", "[English] Naruto", 2},
		{"italian tag", "[Italian] Naruto", 3},
		{"movie", "[Movie] Inception", 4},
		{"tv", "[TV] Show", 4},
		{"movies tv", "[Movies/TV] Show", 4},
		{"unknown default", "Plain title", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, languagePriority(tt.in))
		})
	}
}

// isStdoutTerminal caches via sync.Once. Test runs in CI where stdout is not
// a TTY → should return false. Locally a developer may see true; just verify
// it doesn't panic and result is stable across calls.
func TestIsStdoutTerminal_Stable(t *testing.T) {
	t.Parallel()
	first := isStdoutTerminal()
	second := isStdoutTerminal()
	assert.Equal(t, first, second)
}
