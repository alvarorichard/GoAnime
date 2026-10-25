package handlers

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func TestScopeIndexForSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src  string
		want int
	}{
		{"", 0},
		{"unknown", 0},
		{"hianime", 2},
		{"ptbr", 1},
		{"pt-br", 1},
		{"animefire", 1},
		{"goyabu", 1},
		{"animesdigital", 1},
		{"movietv", 3},
		{"startflix", 3},
		{"topcine", 3},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, scopeIndexForSource(tt.src))
		})
	}
}

func TestContinueSourceToken(t *testing.T) {
	t.Parallel()
	tests := []struct{ src, want string }{
		{"Animefire.io", "animefire"},
		{"HiAnime", "hianime"},
		{"Goyabu", "goyabu"},
		{"AnimesDigital", "animesdigital"},
		{"StartFlix", "startflix"},
		{"TopCine", "topcine"},
		{"AnimeDrive", "animedrive"},
		{"", ""},
		{"Weird Site", ""},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, continueSourceToken(tt.src))
		})
	}
}

func TestPickContinueVersion(t *testing.T) {
	t.Parallel()
	fire := &models.Anime{Name: "X", Source: "Animefire.io"}
	hi := &models.Anime{Name: "X", Source: "HiAnime"}

	pick := pickContinueVersion("Animefire.io", "")
	assert.Same(t, fire, pick([]*models.Anime{hi, fire}), "tracked source wins over order")

	pick = pickContinueVersion("HiAnime", "")
	assert.Same(t, hi, pick([]*models.Anime{fire, hi}))

	pick = pickContinueVersion("", "")
	assert.Same(t, fire, pick([]*models.Anime{fire, hi}), "unknown source keeps closest match")

	pick = pickContinueVersion("HiAnime", "")
	assert.Nil(t, pick(nil))
	assert.Nil(t, pick([]*models.Anime{nil}), "nil entries are skipped")
}

func TestHandleSearchHome_NoTTYErrors(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs a non-interactive stdin")
	}
	_, err := HandleSearchHome()
	require.Error(t, err)
}

func TestHandleContinue_NilEntryReturns(t *testing.T) {
	assert.NotPanics(t, func() { HandleContinue(SearchHomeDecision{}) })
}
