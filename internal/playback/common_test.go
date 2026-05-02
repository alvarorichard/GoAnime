package playback

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/term"
)

// TestSelectEpisodeWithFuzzy_EmptyList verifies that passing an empty episode
// list returns an error instead of calling log.Fatal (which was the old behavior).
func TestSelectEpisodeWithFuzzy_EmptyList(t *testing.T) {
	_, _, _, err := SelectEpisodeWithFuzzy([]models.Episode{})

	assert.Error(t, err, "expected error for empty episode list")
	t.Logf("Got expected error: %v", err)
}

// TestFindEpisodeByNumber_NotFound verifies that searching for a non-existent
// episode number returns an error instead of fataling.
func TestFindEpisodeByNumber_NotFound(t *testing.T) {
	// This test falls back to SelectEpisodeWithFuzzy which opens an interactive
	// fuzzy finder. Keep it opt-in so local and CI runs stay hermetic by default.
	if os.Getenv("GOANIME_INTERACTIVE_TESTS") != "1" || !term.IsTerminal(int(os.Stdout.Fd())) {
		t.Skip("Skipping interactive fuzzy-finder test without GOANIME_INTERACTIVE_TESTS=1 and a real TTY")
	}

	episodes := []models.Episode{
		{URL: "https://example.com/ep1", Number: "1"},
		{URL: "https://example.com/ep2", Number: "2"},
	}

	// Episode 999 doesn't exist — FindEpisodeByNumber falls back to
	// SelectEpisodeWithFuzzy which will fail on non-interactive env.
	// The important thing is it returns an error, not os.Exit.
	_, _, _, err := FindEpisodeByNumber(episodes, 999)

	assert.Error(t, err, "expected error for non-existent episode number")
	t.Logf("Got expected error: %v", err)
}
