package tui

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live search must rank the exact title first even when StartFlix's loose
// WordPress search answers first with unrelated description matches.
func TestSearchingSortsRelevanceBeforeLanguage(t *testing.T) {
	t.Parallel()
	m := newSearchingModel("todo poderoso", nil)
	m.applyEvent(models.SearchEvent{Plan: true, Searched: []string{"StartFlix", "TopCine"}, GiveUp: time.Second})
	m.applyEvent(models.SearchEvent{Name: "StartFlix", State: models.SearchDone,
		Results: []*models.Anime{
			{Name: "[Movie] [PT-BR] A Casa Mórbida", Source: "StartFlix", Year: "2024"},
			{Name: "[Movie] [PT-BR] Conclave", Source: "StartFlix", Year: "2024"},
		}, Count: 2})
	m.applyEvent(models.SearchEvent{Name: "TopCine", State: models.SearchDone,
		Results: []*models.Anime{
			{Name: "[Movie] [PT-BR] Todo Poderoso", Source: "TopCine", Year: "2003"},
		}, Count: 1})
	require.Len(t, m.groups, 3)
	assert.Equal(t, "Todo Poderoso", m.groups[0].title)
}
