package playback

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
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
	// fuzzy finder (tcell-based TUI). On CI there is no TTY, so the fuzzy finder
	// either panics (Windows) or hangs indefinitely waiting for terminal input.
	if os.Getenv("CI") != "" || !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("Skipping interactive fuzzy-finder test: requires a real TTY")
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

// TestFindEpisodeByNumber_Found_FirstMatch verifies the happy-path direct hit
// (no fuzzy-finder fallback). Returns the exact URL/Number/Num triple.
func TestFindEpisodeByNumber_Found_FirstMatch(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{URL: "https://example.com/ep1", Number: "1"},
		{URL: "https://example.com/ep2", Number: "2"},
		{URL: "https://example.com/ep3", Number: "3"},
	}

	url, numStr, num, err := FindEpisodeByNumber(episodes, 2)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/ep2", url)
	assert.Equal(t, "2", numStr)
	assert.Equal(t, 2, num)
}

// TestFindEpisodeByNumber_Found_PrefixedNumber verifies that prefixed episode
// numbers ("Episode 5", "Ep5", etc.) are matched via player.ExtractEpisodeNumber.
func TestFindEpisodeByNumber_Found_PrefixedNumber(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{
		{URL: "u1", Number: "Episode 1"},
		{URL: "u5", Number: "Episode 5"},
	}

	url, _, num, err := FindEpisodeByNumber(episodes, 5)
	assert.NoError(t, err)
	assert.Equal(t, "u5", url)
	assert.Equal(t, 5, num)
}
