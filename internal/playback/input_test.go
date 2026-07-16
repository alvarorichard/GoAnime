package playback

import (
	"errors"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swapFindMenu replaces the menu finder seam for the duration of the test.
// Tests using it mutate a package global, so they must not run in parallel.
func swapFindMenu(t *testing.T, fn func(items []menuItem, itemFunc func(i int) string, opts ...fuzzyfinder.Option) (int, error)) {
	t.Helper()
	orig := findMenuFunc
	findMenuFunc = fn
	t.Cleanup(func() { findMenuFunc = orig })
}

func TestGetUserInput_MenuErrorDoesNotAutoAdvance(t *testing.T) {
	// Regression: on menu failure (non-TTY, Esc, broken terminal) GetUserInput
	// returned "n", which made HandleSeries/HandleMovie auto-play the next
	// episode forever with zero user input. Failure must quit.
	swapFindMenu(t, func(_ []menuItem, _ func(i int) string, _ ...fuzzyfinder.Option) (int, error) {
		return 0, errors.New("failed to open TTY")
	})
	assert.Equal(t, "q", GetUserInput(), "menu failure must map to quit, never to next-episode")
	assert.Equal(t, "q", GetUserInput(true), "movie menu failure must map to quit, never to replay")
}

func TestGetUserInput_SeriesMenuMapping(t *testing.T) {
	wantLabels := []string{"Next episode", "Previous episode", "Select episode", "Change anime", "← Back", "Quit"}
	wantValues := []string{"n", "p", "e", "c", "back", "q"}

	for i, wantValue := range wantValues {
		var gotLabels []string
		swapFindMenu(t, func(items []menuItem, itemFunc func(i int) string, _ ...fuzzyfinder.Option) (int, error) {
			gotLabels = nil
			for j := range items {
				gotLabels = append(gotLabels, itemFunc(j))
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
		swapFindMenu(t, func(items []menuItem, itemFunc func(i int) string, _ ...fuzzyfinder.Option) (int, error) {
			gotLabels = nil
			for j := range items {
				gotLabels = append(gotLabels, itemFunc(j))
			}
			return i, nil
		})
		got := GetUserInput(true)
		require.Equal(t, wantLabels, gotLabels)
		assert.Equal(t, wantValue, got, "label %q must map to %q", wantLabels[i], wantValue)
	}
}
