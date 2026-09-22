package appflow

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #203: the reporter's log ends with the result screen listing 15 Bleach
// matches, then
//
//	ERROR No anime found with the name: bleach
//	INFO  Please enter a new search term.
//	ERROR Failed to search for anime: search cancelled by user
//
// Quitting the picker (q / Ctrl+C) was funnelled into the same branch as a
// genuinely empty search, so a deliberate exit was reported as a broken search
// and answered with another prompt.

func TestSearchAnimeWithRetry_QuitExitsWithoutPrompting(t *testing.T) {
	prompted := false
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			return nil, api.ErrSearchAborted
		},
		promptForName: func(string) (string, error) {
			prompted = true
			return "", errors.New("prompt must not run")
		},
	})

	anime, err := SearchAnimeWithRetry("bleach")
	assert.Nil(t, anime)
	require.ErrorIs(t, err, api.ErrSearchAborted)
	assert.False(t, prompted, "quitting the result screen must not re-prompt for a search term")
}

// The picker's own cancel error must reach the retry loop as an abort, so the
// wiring between tui, api and appflow cannot drift apart silently.
func TestSearchAnimeWithRetry_PickerCancelIsAnAbort(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			// What api.searchAnimeEnhanced returns for tui.ErrSelectionCancelled.
			return nil, errAbortedWithCause()
		},
		promptForName: func(string) (string, error) {
			t.Fatal("prompt must not run after a quit")
			return "", nil
		},
	})

	_, err := SearchAnimeWithRetry("bleach")
	require.ErrorIs(t, err, api.ErrSearchAborted)
	assert.ErrorIs(t, err, tui.ErrSelectionCancelled, "the original cause stays in the chain")
}

func errAbortedWithCause() error {
	return errors.Join(api.ErrSearchAborted, tui.ErrSelectionCancelled)
}

// A back request still re-prompts: it means "different search term", not "quit".
func TestSearchAnimeWithRetry_BackStillPrompts(t *testing.T) {
	want := &models.Anime{Name: "Bleach", Source: "Goyabu"}
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, api.ErrBackToSearch
			}
			assert.Equal(t, "bleach dublado", name)
			return want, nil
		},
		promptForName: func(string) (string, error) { return "bleach dublado", nil },
	})

	got, err := SearchAnimeWithRetry("bleach")
	require.NoError(t, err)
	assert.Same(t, want, got)
	assert.Equal(t, 2, calls)
}

// A transport failure is reported as a failure, not as "nothing matched", but
// it still re-prompts so the user can try again.
func TestSearchAnimeWithRetry_TransportErrorStillPrompts(t *testing.T) {
	want := &models.Anime{Name: "Bleach"}
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("dial tcp: connection refused")
			}
			return want, nil
		},
		promptForName: func(string) (string, error) { return "bleach", nil },
	})

	got, err := SearchAnimeWithRetry("bleach")
	require.NoError(t, err)
	assert.Same(t, want, got)
}
