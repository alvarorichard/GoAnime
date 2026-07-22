package naming

import "testing"

func BenchmarkSanitizeFilename(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SanitizeFilename(`Fate/stay night: Unlimited Blade Works  <Episode 12?>`)
	}
}

func BenchmarkCleanTitle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		CleanTitle("[PT-BR] [SuperFlix] Vinland Saga [TV]")
	}
}
