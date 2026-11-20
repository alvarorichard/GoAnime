package appflow

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The home screen restores its own place from the decision it made, so the
// loop between the results and the home passes the query only, and routes the
// next search by Query and Scope, never by where the screen was.
func TestBackToTheHomeKeepsTheQueryAndIgnoresWhereTheScreenWas(t *testing.T) {
	cleanLastSearch(t)
	var opened []string
	prev := OpenSearchHomeFn
	t.Cleanup(func() { OpenSearchHomeFn = prev })
	OpenSearchHomeFn = func(query string) (tui.SearchHomeResult, error) {
		opened = append(opened, query)
		return tui.SearchHomeResult{
			Action: tui.HomeSearch, Query: "naruto", Scope: "ptbr",
			Restore: tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k", Query: "never searched"},
		}, nil
	}
	var searched [][2]string
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, scope string) (*models.Anime, error) {
			searched = append(searched, [2]string{name, scope})
			if len(searched) == 1 {
				return nil, api.ErrBackToSearch
			}
			return &models.Anime{Name: "Naruto"}, nil
		},
	})

	anime, err := SearchAnimeWithRetryIn("one-piece", "")

	require.NoError(t, err)
	assert.Equal(t, "Naruto", anime.Name)
	assert.Equal(t, []string{"one-piece"}, opened, "the home gets the query that was searched, as before")
	assert.Equal(t, [][2]string{{"one-piece", ""}, {"naruto", "ptbr"}}, searched)
}
