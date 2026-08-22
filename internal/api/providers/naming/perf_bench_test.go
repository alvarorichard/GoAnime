package naming

import "testing"

func BenchmarkSanitizeFilename(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		SanitizeFilename(`Fate/stay night: Unlimited Blade Works  <Episode 12?>`)
	}
}

func BenchmarkCleanTitle(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		CleanTitle("[PT-BR] [SuperFlix] Vinland Saga [TV]")
	}
}
