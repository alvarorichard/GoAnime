package player

import (
	"sync"
	"testing"
)

// TestRaceOnGlobalMediaVars verifies that the mutex-protected media state is
// free of data races. Before the fix the bare globals caused DATA RACE
// warnings under -race. After the fix this test must PASS cleanly.
//
// Run with: go test -race -run TestRaceOnGlobalMediaVars ./internal/player/
func TestRaceOnGlobalMediaVars(t *testing.T) {
	var wg sync.WaitGroup

	// Writer goroutine — simulates what HandleDownloadAndPlay / download workflow does
	wg.Go(func() {
		for i := range 1000 {
			SetAnimeName("Naruto", i%5+1)
			SetExactMediaType("anime")
			SetMediaType(false)
			setLastAnimeURL("https://example.com/anime/" + string(rune('A'+i%26)))
		}
	})

	// Reader goroutines — simulate what createEpisodePath / batch download goroutines do
	for range 4 {
		wg.Go(func() {
			for range 1000 {
				_ = GetExactMediaType()
				_ = IsCurrentMediaMovie()
				_ = getLastAnimeURL()
				snap := snapshotMedia()
				_ = snap.AnimeName
				_ = snap.AnimeSeason
			}
		})
	}

	wg.Wait()
}
