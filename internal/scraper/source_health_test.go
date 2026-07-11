package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHealthCheckQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source ScraperType
		want   string
	}{
		{"superflix", SuperFlixType, "dexter"},
		{"allanime default", AllAnimeType, "naruto"},
		{"animefire default", AnimefireType, "naruto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, DefaultHealthCheckQuery(tt.source))
		})
	}
}

func TestHealthTargets_DeterministicOrder(t *testing.T) {
	t.Parallel()
	targets := healthTargets()
	require.NotEmpty(t, targets)
	for i := 1; i < len(targets); i++ {
		assert.LessOrEqual(t, targets[i-1], targets[i], "must be sorted asc")
	}
}

func TestCheckSourceHealth_NilScraperFails(t *testing.T) {
	t.Parallel()
	res := checkSourceHealthWith(context.Background(), AllAnimeType, nil, "naruto")
	assert.Equal(t, SourceHealthFailed, res.Status)
	assert.NotNil(t, res.Diagnostic)
}

func TestCheckAllSourcesHealth_ReturnsOnePerSource(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := CheckAllSourcesHealth(ctx)
	assert.Equal(t, len(healthTargets()), len(results))
}
