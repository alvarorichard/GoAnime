// Example: Basic anime search using GoAnime library
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
	// Create a new client
	client := goanime.NewClient()

	// Search for anime across all sources
	fmt.Println("Searching for 'One Piece'...")
	results, err := client.SearchAnime("One Piece", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d results:\n\n", len(results))
	for i, anime := range results {
		fmt.Printf("%d. %s\n", i+1, anime.Name)
		fmt.Printf("   Source: %s\n", anime.Source)
		fmt.Printf("   URL: %s\n", anime.URL)
		if anime.ImageURL != "" {
			fmt.Printf("   Image: %s\n", anime.ImageURL)
		}
		if anime.Details != nil && anime.Details.Description != "" {
			desc := anime.Details.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("   Description: %s\n", desc)
		}
		fmt.Println()
	}
}
