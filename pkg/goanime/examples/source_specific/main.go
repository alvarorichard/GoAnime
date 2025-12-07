// Example: Search in specific source
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
	client := goanime.NewClient()

	// List available sources
	fmt.Println("Available sources:")
	sources := client.GetAvailableSources()
	for i, source := range sources {
		fmt.Printf("%d. %s\n", i+1, source.String())
	}
	fmt.Println()

	// Search only in AllAnime
	source := types.SourceAllAnime
	fmt.Printf("Searching for 'Attack on Titan' in %s...\n\n", source.String())

	results, err := client.SearchAnime("Attack on Titan", &source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d results from %s:\n\n", len(results), source.String())
	for i, anime := range results {
		fmt.Printf("%d. %s\n", i+1, anime.Name)
		fmt.Printf("   URL: %s\n", anime.URL)

		if anime.Details != nil {
			if len(anime.Details.Genres) > 0 {
				fmt.Printf("   Genres: %v\n", anime.Details.Genres)
			}
			if anime.Details.AverageScore > 0 {
				fmt.Printf("   Score: %d/100\n", anime.Details.AverageScore)
			}
			if anime.Details.Episodes > 0 {
				fmt.Printf("   Episodes: %d\n", anime.Details.Episodes)
			}
		}
		fmt.Println()
	}

	// Now search in AnimeFire
	source = types.SourceAnimeFire
	fmt.Printf("\nSearching for 'Attack on Titan' in %s...\n\n", source.String())

	results, err = client.SearchAnime("Attack on Titan", &source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d results from %s:\n\n", len(results), source.String())
	for i, anime := range results {
		fmt.Printf("%d. %s\n", i+1, anime.Name)
		fmt.Printf("   URL: %s\n\n", anime.URL)
	}
}
