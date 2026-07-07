package scraper

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
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

func TestAvailableSources_DeterministicOrder(t *testing.T) {
	t.Parallel()
	sm := NewScraperManager()
	sources := sm.AvailableSources()
	require.NotEmpty(t, sources)
	for i := 1; i < len(sources); i++ {
		assert.LessOrEqual(t, sources[i-1], sources[i], "must be sorted asc")
	}
}

func TestCheckSourceHealth_UnregisteredFails(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{scrapers: map[ScraperType]UnifiedScraper{}}
	res := sm.CheckSourceHealth(context.Background(), AllAnimeType, "naruto")
	assert.Equal(t, SourceHealthFailed, res.Status)
	assert.NotNil(t, res.Diagnostic)
}

func TestCheckSourceHealth_CircuitOpenSkips(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{scrapers: map[ScraperType]UnifiedScraper{AllAnimeType: nil}}
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}
	for i := 0; i < defaultSourceFailureThreshold; i++ {
		sm.recordSourceFailure(AllAnimeType, diag)
	}
	res := sm.CheckSourceHealth(context.Background(), AllAnimeType, "naruto")
	assert.Equal(t, SourceHealthSkipped, res.Status)
}

func TestCheckAllSourcesHealth_ReturnsOnePerSource(t *testing.T) {
	t.Parallel()
	sm := NewScraperManager()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := sm.CheckAllSourcesHealth(ctx)
	assert.Equal(t, len(sm.AvailableSources()), len(results))
}
