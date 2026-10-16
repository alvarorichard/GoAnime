//go:build ignore

// Command seed fills the throwaway HOME's watch history with CONTINUE and
// RECENT examples so the README demo opens on a lived-in SEARCH home screen
// (see demo.tape). It is not part of GoAnime.
//
// Every row points at the public-domain shorts server.go serves, and every
// title plays the same silent test pattern, so the recording neither shows
// nor streams copyrighted material. Run with HOME already pointed at the
// demo directory:
//
//	HOME=$DEMO_DIR/home go run docs/demo/seed.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alvarorichard/Goanime/internal/tracking"
)

func main() {
	now := time.Now()
	tracker := tracking.NewLocalTracker(tracking.DBPath())
	if tracker == nil {
		fmt.Fprintln(os.Stderr, "seed: watch history unavailable (CGO/sqlite missing?)")
		os.Exit(1)
	}
	defer func() { _ = tracker.Close() }()

	// CONTINUE, most recent first: one episode stopped midway (resumable)
	// and one watched to the end (done). Positions are seconds; anything at
	// or past nine tenths of the duration reads as done.
	continueRows := []tracking.Anime{
		{
			AllanimeID:    "animefire:saru-to-kani-no-gassen:ep1",
			EpisodeNumber: 1,
			PlaybackTime:  872, // 14:32 of 20:00
			Duration:      1200,
			Title:         "Saru to Kani no Gassen",
			AnimeTitle:    "Saru to Kani no Gassen",
			Source:        "Animefire.io",
			Audio:         "Legendado",
			MediaType:     "anime",
			LastUpdated:   now.Add(-2 * time.Hour),
		},
		{
			AllanimeID:    "animefire:momotarou:ep1",
			EpisodeNumber: 1,
			PlaybackTime:  1150, // past the credits: done
			Duration:      1200,
			Title:         "Momotarou",
			AnimeTitle:    "Momotarou",
			Source:        "Animefire.io",
			Audio:         "Legendado",
			MediaType:     "anime",
			LastUpdated:   now.Add(-26 * time.Hour),
		},
	}
	for _, row := range continueRows {
		if err := tracker.UpdateProgress(row); err != nil {
			fmt.Fprintln(os.Stderr, "seed: continue:", err)
			os.Exit(1)
		}
	}

	// RECENT, oldest first: searched_at is stamped at insert time, so the
	// last insert lands on top. Counts match what the stand-in catalog
	// returns for these queries.
	for _, q := range []struct {
		query string
		count int
	}{
		{"saru to kani", 5},
		{"kuro nyago", 5},
		{"momotarou", 5},
	} {
		if err := tracker.RecordRecentSearch(q.query, q.count); err != nil {
			fmt.Fprintln(os.Stderr, "seed: recent:", err)
			os.Exit(1)
		}
		time.Sleep(1100 * time.Millisecond)
	}

	fmt.Println("seed: 2 continue rows, 3 recent searches")
}
