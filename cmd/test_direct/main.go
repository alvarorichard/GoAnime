package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

func main() {
	// Test AllAnime URL extraction step by step
	client := scraper.NewAllAnimeClient()

	animeID := "2oXgpDPd3xKWdgnoz"
	episodeNo := "1"
	mode := "sub"
	quality := "best"

	fmt.Printf("Testing AllAnime GetEpisodeURL with:\n")
	fmt.Printf("  Anime ID: %s\n", animeID)
	fmt.Printf("  Episode: %s\n", episodeNo)
	fmt.Printf("  Mode: %s\n", mode)
	fmt.Printf("  Quality: %s\n", quality)
	fmt.Printf("\n")

	// Call GetEpisodeURL and see what happens
	streamURL, metadata, err := client.GetEpisodeURL(animeID, episodeNo, mode, quality)
	if err != nil {
		log.Printf("ERROR: %v", err)
	} else {
		fmt.Printf("SUCCESS!\n")
		fmt.Printf("Stream URL: %s\n", streamURL)
		fmt.Printf("Metadata: %v\n", metadata)
	}
}
