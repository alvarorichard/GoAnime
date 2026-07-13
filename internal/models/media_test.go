package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMedia_IsAnime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mt   MediaType
		want bool
	}{
		{"anime", MediaTypeAnime, true},
		{"movie", MediaTypeMovie, false},
		{"tv", MediaTypeTV, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{MediaType: tt.mt}
			assert.Equal(t, tt.want, m.IsAnime())
		})
	}
}

func TestMedia_IsMovie(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mt   MediaType
		want bool
	}{
		{"movie", MediaTypeMovie, true},
		{"tv", MediaTypeTV, false},
		{"anime", MediaTypeAnime, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{MediaType: tt.mt}
			assert.Equal(t, tt.want, m.IsMovie())
		})
	}
}

func TestMedia_IsTV(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mt   MediaType
		want bool
	}{
		{"tv", MediaTypeTV, true},
		{"movie", MediaTypeMovie, false},
		{"anime", MediaTypeAnime, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{MediaType: tt.mt}
			assert.Equal(t, tt.want, m.IsTV())
		})
	}
}

func TestMedia_IsMovieOrTV(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mt   MediaType
		want bool
	}{
		{"movie", MediaTypeMovie, true},
		{"tv", MediaTypeTV, true},
		{"anime", MediaTypeAnime, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{MediaType: tt.mt}
			assert.Equal(t, tt.want, m.IsMovieOrTV())
		})
	}
}

func TestMedia_GetDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    *Media
		want string
	}{
		{"name and year", &Media{Name: "Naruto", Year: "2002"}, "Naruto (2002)"},
		{"name only", &Media{Name: "Naruto"}, "Naruto"},
		{"empty year", &Media{Name: "Naruto", Year: ""}, "Naruto"},
		{"empty all", &Media{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.m.GetDisplayName())
		})
	}
}

func TestMedia_OfficialTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    *Media
		want string
	}{
		{
			"tmdb title wins",
			&Media{Name: "scraped", TMDBDetails: &TMDBDetails{Title: "TMDB Title", Name: "tmdb name"}},
			"TMDB Title",
		},
		{
			"tmdb name when no title",
			&Media{Name: "scraped", TMDBDetails: &TMDBDetails{Name: "TMDB Name"}},
			"TMDB Name",
		},
		{
			"anilist english",
			&Media{Name: "scraped", Details: AniListDetails{Title: Title{English: "Eng", Romaji: "Rom"}}},
			"Eng",
		},
		{
			"anilist romaji",
			&Media{Name: "scraped", Details: AniListDetails{Title: Title{Romaji: "Rom"}}},
			"Rom",
		},
		{
			"fallback name",
			&Media{Name: "scraped"},
			"scraped",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.m.OfficialTitle())
		})
	}
}

func TestMedia_GetRatingDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    float64
		want string
	}{
		{"positive", 8.5, "★ 8.5"},
		{"zero", 0.0, ""},
		{"negative", -1.0, ""},
		{"one digit", 7.0, "★ 7.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{Rating: tt.r}
			assert.Equal(t, tt.want, m.GetRatingDisplay())
		})
	}
}

func TestMedia_GetGenresDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		g    []string
		want string
	}{
		{"empty", nil, ""},
		{"one", []string{"Action"}, "Action"},
		{"three", []string{"Action", "Drama", "SciFi"}, "Action, Drama, SciFi"},
		{"four caps at three", []string{"A", "B", "C", "D"}, "A, B, C"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{Genres: tt.g}
			assert.Equal(t, tt.want, m.GetGenresDisplay())
		})
	}
}

func TestMedia_HasInteractiveEpisodeFlow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		mt     MediaType
		want   bool
	}{
		{"sflix by source", "SFlix", MediaTypeAnime, true},
		{"superflix by source even when tagged anime", "SuperFlix", MediaTypeAnime, true},
		{"superflix with empty media type", "SuperFlix", "", true},
		{"movie by media type", "AllAnime", MediaTypeMovie, true},
		{"tv by media type", "AllAnime", MediaTypeTV, true},
		{"regular anime source", "AllAnime", MediaTypeAnime, false},
		{"animefire anime", "Animefire.io", MediaTypeAnime, false},
		{"empty everything", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{Source: tt.source, MediaType: tt.mt}
			assert.Equal(t, tt.want, m.HasInteractiveEpisodeFlow())
		})
	}
}

func TestMedia_GetRuntimeDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    int
		want string
	}{
		{"zero", 0, ""},
		{"negative", -5, ""},
		{"minutes only", 45, "45m"},
		{"hours only", 120, "2h 0m"},
		{"hours and minutes", 95, "1h 35m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Media{Runtime: tt.r}
			assert.Equal(t, tt.want, m.GetRuntimeDisplay())
		})
	}
}
