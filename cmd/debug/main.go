// Simple test to understand what's happening with AllAnime URLs
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

func main() {
	fmt.Println("🧪 AllAnime URL Debug Test")
	fmt.Println("==========================")

	client := scraper.NewAllAnimeClient()

	// Test search
	fmt.Println("1. Testing search...")
	animes, err := client.SearchAnime("road of naruto", "sub")
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}

	if len(animes) == 0 {
		fmt.Println("No results found")
		return
	}

	anime := animes[0]
	fmt.Printf("Found: %s (ID: %s)\n", anime.Name, anime.URL)

	// Test episodes list
	fmt.Println("\n2. Testing episodes list...")
	episodes, err := client.GetEpisodesList(anime.URL, "sub")
	if err != nil {
		log.Printf("Episodes list failed: %v", err)
		return
	}

	fmt.Printf("Found %d episodes: %v\n", len(episodes), episodes)

	if len(episodes) == 0 {
		fmt.Println("No episodes found")
		return
	}

	// Test getting episode URL
	fmt.Println("\n3. Testing episode URL extraction...")
	episodeURL, metadata, err := client.GetEpisodeURL(anime.URL, episodes[0], "sub", "best")
	if err != nil {
		log.Printf("Episode URL failed: %v", err)
		return
	}

	fmt.Printf("Episode URL: %s\n", episodeURL)
	fmt.Printf("Metadata: %v\n", metadata)
}
