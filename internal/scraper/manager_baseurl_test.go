package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScraperManager_BaseURLForKnownTypes(t *testing.T) {
	// After SFlix/NineAnime/AnimeDrive removal, getScraperBaseURL returns
	// empty for every type — homepage probes are no longer performed.
	sm := &ScraperManager{}

	assert.Empty(t, sm.getScraperBaseURL(AllAnimeType),
		"AllAnime uses a GraphQL endpoint, not a probable HTML root")
	assert.Empty(t, sm.getScraperBaseURL(GoyabuType),
		"Goyabu serves challenge pages on its homepage")
	assert.Empty(t, sm.getScraperBaseURL(AnimefireType))
	assert.Empty(t, sm.getScraperBaseURL(SuperFlixType))
}
