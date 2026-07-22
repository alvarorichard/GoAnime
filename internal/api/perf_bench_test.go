package api

import "testing"

// Benchmark inputs mirror real titles produced by the PT-BR and English
// scrapers so the numbers reflect the actual search hot path.
var benchTitles = []string{
	"🔥[AnimeFire] Naruto Shippuuden Dublado - Todos os Episódios",
	"[PT-BR] Jujutsu Kaisen 2ª Temporada (Dublado) 8.39 A16",
	"[English] Black Clover (170 episodes) N/A",
	"[Movies/TV] One Piece Film: Red – Todos os Episódios",
	"Shingeki no Kyojin: The Final Season Parte 2 Legendado",
	"black-clover-dublado",
}

func BenchmarkCleanTitle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, title := range benchTitles {
			CleanTitle(title)
		}
	}
}

func BenchmarkGenerateSearchVariations(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		generateSearchVariations("Naruto Shippuuden Clássico III")
	}
}
