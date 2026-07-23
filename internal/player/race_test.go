package player

import (
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/util"
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

func TestMediaStateDoesNotLeakMutableAliases(t *testing.T) {
	seasonMap := []metadata.SeasonMapping{{Season: 1, StartEp: 1, EndEp: 12}}
	meta := &util.MediaMeta{OfficialTitle: "original", Year: "2024"}
	SetSeasonMap(seasonMap)
	SetMediaMeta(meta)

	var wg sync.WaitGroup
	wg.Go(func() {
		for range 1000 {
			seasonMap[0].Season++
			meta.OfficialTitle = "caller mutation"
		}
	})
	for range 4 {
		wg.Go(func() {
			for range 1000 {
				snap := snapshotMedia()
				if len(snap.SeasonMap) > 0 {
					snap.SeasonMap[0].Season++
				}
				if snap.Meta != nil {
					snap.Meta.OfficialTitle = "reader mutation"
				}
				got := GetMediaMeta()
				if got != nil {
					got.Year = "mutated copy"
				}
			}
		})
	}
	wg.Wait()

	snap := snapshotMedia()
	if snap.SeasonMap[0].Season != 1 {
		t.Fatalf("season map escaped synchronization: got season %d", snap.SeasonMap[0].Season)
	}
	if snap.Meta.OfficialTitle != "original" || snap.Meta.Year != "2024" {
		t.Fatalf("media metadata escaped synchronization: %+v", snap.Meta)
	}
}
