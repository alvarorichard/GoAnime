package api

import (
	"errors"
	"fmt"
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
		{"movie", "[Movie] Inception", 3},
		{"tv", "[TV] Show", 3},
		{"movies tv", "[Movies/TV] Show", 3},
		{"unknown default", "Plain title", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, languagePriority(tt.in))
		})
	}
}

// TestFriendlyMessageSticksToTheChain: the typed rephrasing the api layers
// use keeps the user-facing wording reachable through the whole error chain
// (playback shows FriendlyMessage's result on the result screen), errors.Is
// still sees the root cause, and plain errors answer false instead of being
// mistaken for friendly wording.
func TestFriendlyMessageSticksToTheChain(t *testing.T) {
	t.Parallel()
	root := errors.New("dead server")
	wrapped := fmt.Errorf("failed to get TopCine stream URL: %w",
		Friendly(root, "TopCine only offers this on dead servers. Try it later."))

	msg, ok := FriendlyMessage(wrapped)
	assert.True(t, ok)
	assert.Equal(t, "TopCine only offers this on dead servers. Try it later.", msg)

	assert.ErrorIs(t, wrapped, root, "the cause stays reachable")

	msg, ok = FriendlyMessage(&friendlyError{msg: "direct"})
	assert.True(t, ok)
	assert.Equal(t, "direct", msg)

	_, ok = FriendlyMessage(errors.New("raw failure"))
	assert.False(t, ok, "plain errors carry no friendly wording")

	_, ok = FriendlyMessage(nil)
	assert.False(t, ok)
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
