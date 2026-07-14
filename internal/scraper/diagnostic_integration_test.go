package scraper

import (
	"context"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckSourceHealthFailsOnParserBreakButSkipsOffline pins the health
// classification: an offline/blocked origin (HTTP 521) is SKIPPED so CI doesn't
// fail on upstream outages, while a source that responds but no longer parses
// (parser break) is FAILED so a real regression is caught.
func TestCheckSourceHealthFailsOnParserBreakButSkipsOffline(t *testing.T) {
	t.Parallel()

	offlineMock := &MockScraper{
		scraperType: AllAnimeType,
		searchFunc: func(_ string) ([]*models.Anime, error) {
			return nil, netx.NewHTTPStatusError("AllAnime", "search", 521)
		},
	}
	parserMock := &MockScraper{
		scraperType: AnimefireType,
		searchFunc: func(_ string) ([]*models.Anime, error) {
			return nil, fmt.Errorf("no video URL found in AJAX response")
		},
	}

	offline := checkSourceHealthWith(context.Background(), AllAnimeType, offlineMock, "naruto")
	assert.Equal(t, SourceHealthSkipped, offline.Status)
	require.NotNil(t, offline.Diagnostic)
	assert.Equal(t, netx.DiagnosticSourceUnavailable, offline.Diagnostic.Kind)

	parserBroken := checkSourceHealthWith(context.Background(), AnimefireType, parserMock, "naruto")
	assert.Equal(t, SourceHealthFailed, parserBroken.Status)
	require.NotNil(t, parserBroken.Diagnostic)
	assert.Equal(t, netx.DiagnosticParserBroken, parserBroken.Diagnostic.Kind)
}
