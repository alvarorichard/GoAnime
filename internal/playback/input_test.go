package playback

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swapFindMenu replaces the menu finder seam for the duration of the test.
// Tests using it mutate a package global, so they must not run in parallel.
func swapFindMenu(t *testing.T, fn func(items []menuItem) (int, error)) {
	t.Helper()
	orig := findMenuFunc
	findMenuFunc = fn
	t.Cleanup(func() { findMenuFunc = orig })
}

func TestGetUserInput_MenuErrorDoesNotAutoAdvance(t *testing.T) {
	// Regression: on menu failure (non-TTY, broken terminal) GetUserInput
	// returned "n", which made HandleSeries/HandleMovie auto-play the next
	// episode forever with zero user input. Failure must quit.
	swapFindMenu(t, func(_ []menuItem) (int, error) {
		return 0, errors.New("failed to open TTY")
	})
	assert.Equal(t, "q", GetUserInput(), "menu failure must map to quit, never to next-episode")
	assert.Equal(t, "q", GetUserInput(true), "movie menu failure must map to quit, never to replay")
}

func TestGetUserInput_PickBackMapsToBack(t *testing.T) {
	swapFindMenu(t, func(_ []menuItem) (int, error) {
		return -1, tui.ErrPickBack
	})
	assert.Equal(t, "back", GetUserInput())
	assert.Equal(t, "back", GetUserInput(true))
}

func TestGetUserInput_PickCancelMapsToQuit(t *testing.T) {
	swapFindMenu(t, func(_ []menuItem) (int, error) {
		return -1, tui.ErrPickCancelled
	})
	assert.Equal(t, "q", GetUserInput())
}

func TestGetUserInput_SeriesMenuMapping(t *testing.T) {
	wantLabels := []string{"Next episode", "Previous episode", "Select episode", "Change anime", "← Back", "Quit"}
	wantValues := []string{"n", "p", "e", "c", "back", "q"}

	for i, wantValue := range wantValues {
		var gotLabels []string
		swapFindMenu(t, func(items []menuItem) (int, error) {
			gotLabels = nil
			for _, it := range items {
				gotLabels = append(gotLabels, it.Label)
			}
			return i, nil
		})
		got := GetUserInput()
		require.Equal(t, wantLabels, gotLabels)
		assert.Equal(t, wantValue, got, "label %q must map to %q", wantLabels[i], wantValue)
	}
}

func TestGetUserInput_MovieMenuMapping(t *testing.T) {
	wantLabels := []string{"Replay movie", "Change movie", "← Back", "Quit"}
	wantValues := []string{"n", "c", "back", "q"}

	for i, wantValue := range wantValues {
		var gotLabels []string
		swapFindMenu(t, func(items []menuItem) (int, error) {
			gotLabels = nil
			for _, it := range items {
				gotLabels = append(gotLabels, it.Label)
			}
			return i, nil
		})
		got := GetUserInput(true)
		require.Equal(t, wantLabels, gotLabels)
		assert.Equal(t, wantValue, got, "label %q must map to %q", wantLabels[i], wantValue)
	}
}

func TestGetUserInput_InvalidIndexQuits(t *testing.T) {
	swapFindMenu(t, func(_ []menuItem) (int, error) {
		return 99, nil
	})
	assert.Equal(t, "q", GetUserInput())
}
