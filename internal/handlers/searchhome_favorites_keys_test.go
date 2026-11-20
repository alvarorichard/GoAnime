package handlers

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeFavoriteKeysAreHandedBackExactly(t *testing.T) {
	store := useFavorites(t)
	// removal looks the row up by this key, so it must not be cleaned up
	_, err := store.Toggle(favorites.Entry{Key: "a\tb\x1b[0m", Title: "Odd"})
	require.NoError(t, err)

	rows := homeFavorites()

	require.Len(t, rows, 1)
	assert.Equal(t, "a\tb\x1b[0m", rows[0].Key)
	require.NoError(t, forgetFavorite(rows[0].Key))
	assert.Empty(t, homeFavorites())
}
