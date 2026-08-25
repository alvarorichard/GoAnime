package scraper

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
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

func TestScraperDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		st   ScraperType
		want string
	}{
		{AniDBType, "AniDB"},
		{AnimefireType, "Animefire.io"},
		{GoyabuType, "Goyabu"},
		{SuperFlixType, "SuperFlix"},
		{ScraperType(999), "Desconhecido"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, scraperDisplayName(tt.st))
	}
}

func TestScraperLanguageTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		st   ScraperType
		want string
	}{
		{AniDBType, "[English]"},
		{AnimefireType, "[PT-BR]"},
		{GoyabuType, "[PT-BR]"},
		{SuperFlixType, "[PT-BR]"},
		{ScraperType(999), "[Unknown]"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, scraperLanguageTag(tt.st))
	}
}
