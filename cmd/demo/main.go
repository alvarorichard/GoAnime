// Package main demonstrates the enhanced GoAnime scraping functionality
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

func main() {
	fmt.Println("🌟 GoAnime Enhanced Scraper Demo 🌟")
	fmt.Println("=====================================")

	// Create scraper manager
	manager := scraper.NewScraperManager()

	// Test search functionality
	fmt.Println("\n🔍 Testing Multi-Source Search...")
	query := "naruto"

	results, err := manager.SearchAnime(query, nil)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d results for '%s':\n", len(results), query)
	for i, anime := range results {
		if i >= 5 { // Limit to first 5 results
			break
		}
		fmt.Printf("  %d. %s\n", i+1, anime.Name)
		fmt.Printf("     URL: %s\n", anime.URL)
		if anime.ImageURL != "" {
			fmt.Printf("     Image: %s\n", anime.ImageURL)
		}
		fmt.Println()
	}

	// Test AllAnime specific search
	fmt.Println("\n🎯 Testing AllAnime Specific Search...")
	allAnimeType := scraper.AllAnimeType
	allAnimeResults, err := manager.SearchAnime("one piece", &allAnimeType)
	if err != nil {
		log.Printf("AllAnime search failed: %v", err)
	} else {
		fmt.Printf("AllAnime found %d results:\n", len(allAnimeResults))
		for i, anime := range allAnimeResults {
			if i >= 3 { // Limit to first 3 results
				break
			}
			fmt.Printf("  %d. %s\n", i+1, anime.Name)
		}
	}

	// Test AnimeFire specific search
	fmt.Println("\n🔥 Testing AnimeFire Specific Search...")
	animefireType := scraper.AnimefireType
	animefireResults, err := manager.SearchAnime("attack on titan", &animefireType)
	if err != nil {
		log.Printf("AnimeFire search failed: %v", err)
	} else {
		fmt.Printf("AnimeFire found %d results:\n", len(animefireResults))
		for i, anime := range animefireResults {
			if i >= 3 { // Limit to first 3 results
				break
			}
			fmt.Printf("  %d. %s\n", i+1, anime.Name)
		}
	}

	// Test episode listing (using first result if available)
	if len(results) > 0 {
		fmt.Println("\n📺 Testing Episode Listing...")
		anime := results[0]

		// Get scraper for this anime
		var scraperType scraper.ScraperType
		if anime.URL != "" {
			scraperType = scraper.AllAnimeType // Default assumption
		}

		scraperInstance, err := manager.GetScraper(scraperType)
		if err != nil {
			log.Printf("Failed to get scraper: %v", err)
		} else {
			episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
			if err != nil {
				log.Printf("Failed to get episodes: %v", err)
			} else {
				fmt.Printf("Found %d episodes for '%s':\n", len(episodes), anime.Name)
				for i, episode := range episodes {
					if i >= 5 { // Limit to first 5 episodes
						break
					}
					fmt.Printf("  Episode %s: %s\n", episode.Number, episode.Title.Romaji)
				}
			}
		}
	}

	fmt.Println("\n✅ Demo completed! The enhanced scraper is ready to use.")
	fmt.Println("\nUsage examples:")
	fmt.Println("  goanime -d --source allanime \"your anime\" 1")
	fmt.Println("  goanime -d --source animefire --quality 720p \"your anime\" 5")
	fmt.Println("  goanime -d -r --quality best \"your anime\" 1-12")
}
