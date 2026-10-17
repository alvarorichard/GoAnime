package api

import (
	"context"
	"testing"

	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StartFlix's WordPress search matches description text, so "todo poderoso"
// returns unrelated titles. Relevance sorting must put the exact TopCine
// matches first even though the StartFlix garbage arrived first.
func TestSearchAnimeEnhanced_SortsClosestMatchFirst(t *testing.T) {
	t.Parallel()
	garbage := []*models.Anime{
		{Name: "[Movie] [PT-BR] A Casa Mórbida", Source: "StartFlix", URL: "https://www.startflix.biz/filmes/a-casa-morbida/", Year: "2024"},
		{Name: "[Movie] [PT-BR] Conclave", Source: "StartFlix", URL: "https://www.startflix.biz/filmes/conclave/", Year: "2024"},
		{Name: "[Movie] [PT-BR] Eu Voltarei", Source: "StartFlix", URL: "https://www.startflix.biz/filmes/eu-voltarei/", Year: "2023"},
	}
	exact := &models.Anime{Name: "[Movie] [PT-BR] Todo Poderoso", Source: "TopCine", URL: "https://topcine3.site/filme/todo-poderoso", Year: "2003"}
	sequel := &models.Anime{Name: "[Movie] [PT-BR] A Volta do Todo Poderoso", Source: "TopCine", URL: "https://topcine3.site/filme/a-volta-do-todo-poderoso", Year: "2007"}
	// Worst order first: garbage before the exact matches.
	provider := append(append([]*models.Anime{}, garbage...), sequel, exact)

	var shown []*models.Anime
	search := func(context.Context, string, []apisource.SourceKind) ([]*models.Anime, error) {
		return provider, nil
	}

	selected, err := searchAnimeEnhanced("todo-poderoso", "", search, func(animes []*models.Anime) (*models.Anime, error) {
		shown = animes
		return animes[0], nil
	}, nil)
	require.NoError(t, err)
	require.Len(t, shown, 5)
	assert.Same(t, exact, shown[0], "exact title match opens first, not StartFlix description matches")
	assert.Same(t, sequel, shown[1])
	assert.Same(t, exact, selected)
}
