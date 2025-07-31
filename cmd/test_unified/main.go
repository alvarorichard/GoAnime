package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

func main() {
	// Test the unified scraper adapter directly
	fmt.Printf("Testing AllAnimeAdapter through unified interface:\n")

	scraperManager := scraper.NewScraperManager()
	allAnimeScraper, err := scraperManager.GetScraper(scraper.AllAnimeType)
	if err != nil {
		log.Printf("ERROR getting scraper: %v", err)
		return
	}

	animeID := "2oXgpDPd3xKWdgnoz"
	episodeNo := "1"
	quality := "best"

	fmt.Printf("  Anime ID: %s\n", animeID)
	fmt.Printf("  Episode: %s\n", episodeNo)
	fmt.Printf("  Quality: %s\n", quality)
	fmt.Printf("\n")

	// Call through the unified interface like the enhanced API does
	streamURL, metadata, err := allAnimeScraper.GetStreamURL(animeID, episodeNo, quality)
	if err != nil {
		log.Printf("ERROR: %v", err)
	} else {
		fmt.Printf("SUCCESS!\n")
		fmt.Printf("Stream URL: %s\n", streamURL)
		fmt.Printf("Metadata: %v\n", metadata)
	}
}
