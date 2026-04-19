package source

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func BenchmarkResolve(b *testing.B) {
	benchCases := []struct {
		name  string
		anime *models.Anime
	}{
		{
			name: "ExplicitSource",
			anime: &models.Anime{
				Name:   "Naruto",
				URL:    "https://animefire.plus/animes/naruto",
				Source: "Goyabu",
			},
		},
		{
			name: "MediaType",
			anime: &models.Anime{
				Name:      "Inception",
				MediaType: models.MediaTypeMovie,
			},
		},
		{
			name: "PTBRURL",
			anime: &models.Anime{
				Name: "[PT-BR] Naruto",
				URL:  "https://goyabu.to/anime/naruto",
			},
		},
		{
			name: "ShortID",
			anime: &models.Anime{
				Name: "Naruto",
				URL:  "hHjXnUTda",
			},
		},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := Resolve(benchCase.anime); err != nil {
					b.Fatalf("Resolve failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkResolveURL(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"AnimeFire", "https://animefire.plus/ep/naruto-1"},
		{"AnimeDrive", "https://animesdrive.blog/ep/naruto"},
		{"Goyabu", "https://goyabu.to/ep/naruto-1"},
		{"AllAnimeShortID", "hHjXnUTda"},
		{"Unknown", "https://example.com/video"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = ResolveURL(benchCase.url)
			}
		})
	}
}

func BenchmarkExtractAllAnimeID(b *testing.B) {
	benchCases := []struct {
		name  string
		value string
	}{
		{"BareID", "hHjXnUTda"},
		{"AnimeURL", "https://allanime.to/anime/hHjXnUTda"},
		{"NestedAnimeURL", "https://allanime.to/anime/hHjXnUTda/episode-1"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = ExtractAllAnimeID(benchCase.value)
			}
		})
	}
}

func BenchmarkIsAllAnimeShortID(b *testing.B) {
	benchCases := []struct {
		name  string
		value string
	}{
		{"Valid", "hHjXnUTda"},
		{"NumericOnly", "8143"},
		{"HTTPURL", "https://example.com/anime"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = IsAllAnimeShortID(benchCase.value)
			}
		})
	}
}
