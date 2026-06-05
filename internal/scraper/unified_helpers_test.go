package scraper

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortPTBRFirst_StableOrder(t *testing.T) {
	t.Parallel()
	in := []*models.Anime{
		{Name: "[English] A"},
		{Name: "[PT-BR] B"},
		{Name: "[English] C"},
		{Name: "[PT-BR] D"},
	}
	sortPTBRFirst(in)
	assert.Equal(t, "[PT-BR] B", in[0].Name)
	assert.Equal(t, "[PT-BR] D", in[1].Name)
	assert.Equal(t, "[English] A", in[2].Name)
	assert.Equal(t, "[English] C", in[3].Name)
}

func TestCleanPTBRTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"strip dublado", "Naruto Dublado", "Naruto"},
		{"strip legendado parens", "Naruto (Legendado)", "Naruto"},
		{"age rating", "Naruto A16", "Naruto"},
		{"numeric rating", "Naruto 8.39", "Naruto"},
		{"na rating", "Naruto N/A", "Naruto"},
		{"type suffix", "Naruto (TV)", "Naruto"},
		{"movie suffix", "Inception (Movie)", "Inception"},
		{"compound", "Naruto (TV) 8.39 A16 Dublado", "Naruto"},
		{"keep useful", "Boku no Hero", "Boku no Hero"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cleanPTBRTitle(tt.in))
		})
	}
}

func TestNeedsMediaTypeDisambig(t *testing.T) {
	t.Parallel()
	in := []*models.Anime{
		{Name: "Dexter", MediaType: models.MediaTypeMovie},
		{Name: "Dexter", MediaType: models.MediaTypeTV},
		{Name: "Inception", MediaType: models.MediaTypeMovie},
	}
	got := needsMediaTypeDisambig(in)
	assert.True(t, got["dexter"])
	assert.False(t, got["inception"])
}

func TestScraperManager_GetScraperDisplayName(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{}
	tests := []struct {
		st   ScraperType
		want string
	}{
		{AllAnimeType, "AllAnime"},
		{AnimefireType, "Animefire.io"},
		{GoyabuType, "Goyabu"},
		{ScraperType(999), "Desconhecido"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, sm.getScraperDisplayName(tt.st))
	}
}

func TestScraperManager_GetLanguageTag(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{}
	tests := []struct {
		st   ScraperType
		want string
	}{
		{AllAnimeType, "[English]"},
		{AnimefireType, "[PT-BR]"},
		{GoyabuType, "[PT-BR]"},
		{ScraperType(999), "[Unknown]"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, sm.getLanguageTag(tt.st))
	}
}

func TestNewScraperManager_Singleton(t *testing.T) {
	t.Parallel()
	a := NewScraperManager()
	b := NewScraperManager()
	assert.Same(t, a, b)
	assert.NotEmpty(t, a.scrapers)
}

func TestPreWarmScraperManager_NoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { PreWarmScraperManager() })
}

func TestLogSearchSummary_DebugEnabled(t *testing.T) {
	// Exercise the full body (count map + util.Debug call) by enabling debug mode.
	// Cannot be parallel — modifies a package-level var.
	prev := util.IsDebug
	util.IsDebug = true
	t.Cleanup(func() { util.IsDebug = prev })

	sm := &ScraperManager{}
	results := []*models.Anime{
		{Source: "AllAnime"},
		{Source: "AllAnime"},
		{Source: "Animefire.io"},
		{Source: "Goyabu"},
	}
	assert.NotPanics(t, func() { sm.logSearchSummary(results) })
}

func TestLogSearchSummary_DebugDisabled(t *testing.T) {
	// Confirm early return when IsDebug is false (the count map is never built).
	prev := util.IsDebug
	util.IsDebug = false
	t.Cleanup(func() { util.IsDebug = prev })

	sm := &ScraperManager{}
	assert.NotPanics(t, func() { sm.logSearchSummary([]*models.Anime{{Source: "AllAnime"}}) })
}

func TestGetScraper_NotFound(t *testing.T) {
	t.Parallel()
	sm := NewScraperManagerForTest() // empty manager, no scrapers registered
	_, err := sm.GetScraper(AllAnimeType)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetScraper_Found(t *testing.T) {
	t.Parallel()
	sm := NewScraperManagerForTest()
	mock := &MockScraper{}
	sm.RegisterScraperForTest(AllAnimeType, mock)
	got, err := sm.GetScraper(AllAnimeType)
	require.NoError(t, err)
	assert.Equal(t, mock, got)
}
