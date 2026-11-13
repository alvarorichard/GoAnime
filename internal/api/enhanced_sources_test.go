package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
)

func TestRegistryKindsForSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src  string
		want []apisource.SourceKind
	}{
		{"", nil},
		{"unknown", nil},
		{"animesfhd", []apisource.SourceKind{apisource.AnimesFHD}},
		{"anroll", []apisource.SourceKind{apisource.Anroll}},
		{"movietv", []apisource.SourceKind{apisource.StartFlix, apisource.TopCine, apisource.Pobreflix, apisource.Pomfy}},
	}
	for _, tc := range tests {
		t.Run(tc.src, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, registryKindsForSource(tc.src))
		})
	}
	for _, alias := range []string{"ptbr", "pt-br"} {
		assert.Contains(t, registryKindsForSource(alias), apisource.AnimesFHD, alias)
	}
}
