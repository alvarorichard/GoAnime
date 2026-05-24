package appflow

import (
	"errors"
	"runtime"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfWindowsNoTTY skips tests whose retry loop reaches tui.RunClean(huh.Input.Run),
// which blocks indefinitely on Windows without a real console instead of returning an error.
func skipIfWindowsNoTTY(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("huh.NewInput blocks on Windows without a TTY; cannot drive headlessly")
	}
}

// withSearchFn swaps searchEnhancedFn for the duration of the test and restores it.
// Do NOT use with t.Parallel() — shares a package-level var.
func withSearchFn(t *testing.T, fn func(string, string) (*models.Anime, error)) {
	t.Helper()
	prev := searchEnhancedFn
	searchEnhancedFn = fn
	t.Cleanup(func() { searchEnhancedFn = prev })
}

// withSearchWithRetryFn swaps searchWithRetryFn for the duration of the test.
// Do NOT use with t.Parallel() — shares a package-level var.
func withSearchWithRetryFn(t *testing.T, fn func(string, string) (*models.Anime, error)) {
	t.Helper()
	prev := searchWithRetryFn
	searchWithRetryFn = fn
	t.Cleanup(func() { searchWithRetryFn = prev })
}

func TestSearchAnime_Success(t *testing.T) {
	want := &models.Anime{Name: "Naruto", Source: "AllAnime"}
	withSearchFn(t, func(name, source string) (*models.Anime, error) {
		return want, nil
	})

	got, err := SearchAnime("Naruto")
	require.NoError(t, err)
	assert.Equal(t, want.Name, got.Name)
}

func TestSearchAnime_PropagatesError(t *testing.T) {
	withSearchFn(t, func(name, source string) (*models.Anime, error) {
		return nil, errors.New("search failed")
	})

	_, err := SearchAnime("Naruto")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search for anime")
}

func TestSearchAnimeEnhanced_Success(t *testing.T) {
	want := &models.Anime{Name: "One Piece", Source: "AnimeFire"}
	withSearchFn(t, func(name, source string) (*models.Anime, error) {
		assert.Equal(t, "", source, "SearchAnimeEnhanced should pass empty source")
		return want, nil
	})

	got, err := SearchAnimeEnhanced("One Piece")
	require.NoError(t, err)
	assert.Equal(t, "One Piece", got.Name)
}

func TestSearchAnimeEnhanced_PropagatesError(t *testing.T) {
	withSearchFn(t, func(name, source string) (*models.Anime, error) {
		return nil, errors.New("network timeout")
	})

	_, err := SearchAnimeEnhanced("One Piece")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search for anime")
}

func TestGetAnimeEpisodes_SFlixSourceBypassesSpinner(t *testing.T) {
	// SFlix source calls GetAnimeEpisodesEnhanced directly (no spinner).
	// The call will fail (no TMDB ID) but that exercises the non-spinner branch.
	anime := &models.Anime{
		Source: "SFlix",
		URL:    "",
		Name:   "TestMovie",
	}
	_, err := GetAnimeEpisodes(anime)
	require.Error(t, err)
}

func TestGetAnimeEpisodes_MovieTypeBypassesSpinner(t *testing.T) {
	anime := &models.Anime{
		Source:    "SFlix",
		MediaType: models.MediaTypeMovie,
		URL:       "",
		Name:      "TestMovie",
	}
	_, err := GetAnimeEpisodes(anime)
	require.Error(t, err)
}

func TestGetAnimeEpisodes_TVTypeBypassesSpinner(t *testing.T) {
	anime := &models.Anime{
		Source:    "SFlix",
		MediaType: models.MediaTypeTV,
		URL:       "",
		Name:      "TestTV",
	}
	_, err := GetAnimeEpisodes(anime)
	require.Error(t, err)
}

// --- SearchAnimeWithRetry ---

func TestSearchAnimeWithRetry_FirstAttemptSucceeds(t *testing.T) {
	// When the first search attempt returns a valid anime, the function returns
	// immediately without showing any retry prompt.
	want := &models.Anime{Name: "Naruto", Source: "AllAnime"}
	withSearchWithRetryFn(t, func(name, source string) (*models.Anime, error) {
		return want, nil
	})

	got, err := SearchAnimeWithRetry("Naruto")
	require.NoError(t, err)
	assert.Equal(t, want.Name, got.Name)
}

func TestSearchAnimeWithRetry_FirstAttemptReturnsNilAnime(t *testing.T) {
	skipIfWindowsNoTTY(t)
	// searchWithRetryFn returning nil anime (no error) should not be treated as
	// success — the function continues the retry loop.  Without a TTY the prompt
	// call via tui.RunClean returns an error, which triggers "search cancelled".
	withSearchWithRetryFn(t, func(name, source string) (*models.Anime, error) {
		return nil, nil // nil anime with no error → loop continues
	})

	// The retry loop hits tui.RunClean(prompt.Run) which fails without a TTY.
	_, err := SearchAnimeWithRetry("missing")
	require.Error(t, err)
}

func TestSearchAnimeWithRetry_FirstAttemptBackToSearch(t *testing.T) {
	skipIfWindowsNoTTY(t)
	// api.ErrBackToSearch triggers the "Going back to new search..." branch,
	// then falls through to the TUI prompt which fails without a terminal.
	withSearchWithRetryFn(t, func(name, source string) (*models.Anime, error) {
		return nil, api.ErrBackToSearch
	})

	_, err := SearchAnimeWithRetry("anything")
	require.Error(t, err)
	// Error comes from the TUI prompt being unavailable (no TTY).
	assert.Contains(t, err.Error(), "search cancelled")
}

func TestSearchAnimeWithRetry_FirstAttemptOtherError(t *testing.T) {
	skipIfWindowsNoTTY(t)
	// A non-ErrBackToSearch error triggers the "No anime found" branch,
	// then the TUI prompt which fails without a terminal.
	withSearchWithRetryFn(t, func(name, source string) (*models.Anime, error) {
		return nil, errors.New("network timeout")
	})

	_, err := SearchAnimeWithRetry("anything")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "search cancelled")
}
