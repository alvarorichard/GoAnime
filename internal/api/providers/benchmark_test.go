package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
)

func BenchmarkForKindCached(b *testing.B) {
	ResetForTesting()
	if _, err := ForKind(source.Goyabu); err != nil {
		b.Fatalf("warm cache failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if _, err := ForKind(source.Goyabu); err != nil {
			b.Fatalf("ForKind failed: %v", err)
		}
	}
}

func BenchmarkForKindCold(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		ResetForTesting()
		if _, err := ForKind(source.Goyabu); err != nil {
			b.Fatalf("ForKind failed: %v", err)
		}
	}
}

func BenchmarkForKindCachedParallel(b *testing.B) {
	ResetForTesting()
	if _, err := ForKind(source.Goyabu); err != nil {
		b.Fatalf("warm cache failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := ForKind(source.Goyabu); err != nil {
				b.Fatalf("ForKind failed: %v", err)
			}
		}
	})
}

func BenchmarkForAnimeCached(b *testing.B) {
	ResetForTesting()
	anime := &models.Anime{
		Name: "[PT-BR] Naruto",
		URL:  "https://goyabu.to/anime/naruto",
	}

	if _, _, err := ForAnime(anime); err != nil {
		b.Fatalf("warm cache failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if _, _, err := ForAnime(anime); err != nil {
			b.Fatalf("ForAnime failed: %v", err)
		}
	}
}

func BenchmarkHasProvider(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = HasProvider(source.Goyabu)
	}
}
